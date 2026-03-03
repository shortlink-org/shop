package policy_evaluator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/open-policy-agent/opa/rego" //nolint:staticcheck // SA1019: legacy OPA API for v0.x compatibility
	logger "github.com/shortlink-org/go-sdk/logger"

	"github.com/shortlink-org/shop/pricer/internal/domain"
)

// Infrastructure errors for policy evaluation (OPA). Use errors.Is when handling.
var (
	ErrPolicyDirNotExist       = errors.New("policy directory does not exist")
	ErrOPAResultInvalidStr     = errors.New("invalid string format for OPA result")
	ErrOPAResultInvalidNum     = errors.New("invalid json.Number format for OPA result")
	ErrOPAResultUnexpectedType = errors.New("unexpected type for OPA result")
	ErrListRegoFiles           = errors.New("failed to list .rego files")
)

const (
	// Cache configuration for OPA evaluation results
	cacheNumCounters = 10_000    // track 10k evaluations
	cacheMaxCost     = 1_000_000 // ~1MB (results are small float64 values)
	cacheBufferItems = 64
	cacheTTL         = 30 * time.Minute // pricing rules don't change frequently
)

// PolicyEvaluator interface as defined (used by DI and callers).
//
//nolint:iface // interface is implemented by OPAEvaluator and used by DI
type PolicyEvaluator interface {
	Evaluate(ctx context.Context, cart *domain.Cart, params map[string]any) (float64, error)
	Close()
}

// OPAEvaluator implements the PolicyEvaluator interface using OPA's rego package
// with L1 Ristretto cache for evaluation results.
type OPAEvaluator struct {
	preparedQuery rego.PreparedEvalQuery
	query         string
	policyPath    string
	cache         *ristretto.Cache[string, float64]
}

func NewOPAEvaluator(log logger.Logger, policyPath, query string) (*OPAEvaluator, error) {
	// Log the policy path and query
	log.Info("Initializing OPA evaluator",
		slog.String("policy_path", policyPath),
		slog.String("query", query),
	)

	// Check if the policy directory exists
	_, err := os.Stat(policyPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("%s: %w", policyPath, ErrPolicyDirNotExist)
	}

	// Prepare the query
	r := rego.New(
		rego.Query(query),
		rego.Load([]string{policyPath}, nil),
	)

	preparedQuery, err := r.PrepareForEval(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to prepare OPA query: %w", err)
	}

	// Initialize L1 cache
	cache, err := ristretto.NewCache(&ristretto.Config[string, float64]{
		NumCounters: cacheNumCounters,
		MaxCost:     cacheMaxCost,
		BufferItems: cacheBufferItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create evaluation cache: %w", err)
	}

	return &OPAEvaluator{
		preparedQuery: preparedQuery,
		query:         query,
		policyPath:    policyPath,
		cache:         cache,
	}, nil
}

// Close closes the evaluator and releases resources.
func (e *OPAEvaluator) Close() {
	if e.cache != nil {
		e.cache.Close()
	}
}

// Evaluate executes the OPA policy against the provided cart and parameters.
// Uses L1 cache to avoid re-evaluating identical inputs.
func (e *OPAEvaluator) Evaluate(ctx context.Context, cart *domain.Cart, params map[string]any) (float64, error) {
	// Generate cache key from cart and params
	cacheKey := e.generateCacheKey(cart, params)

	// Check L1 cache first
	if cachedResult, found := e.cache.Get(cacheKey); found {
		return cachedResult, nil
	}

	// Cache miss - evaluate the policy
	input := transformCartToInput(cart, params)

	resultSet, err := e.preparedQuery.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return 0.0, fmt.Errorf("OPA evaluation error: %w", err)
	}

	if len(resultSet) == 0 {
		return 0.0, nil // No result from policy
	}

	// Assuming the policy returns a single value
	expr := resultSet[0].Expressions[0].Value

	result, err := parseOPAResult(expr)
	if err != nil {
		return 0.0, err
	}

	// Store in L1 cache (cost=1 since float64 is small)
	e.cache.SetWithTTL(cacheKey, result, 1, cacheTTL)

	return result, nil
}

// generateCacheKey creates a deterministic hash key from cart and params.
func (e *OPAEvaluator) generateCacheKey(cart *domain.Cart, params map[string]any) string {
	hasher := sha256.New()

	// Include policy path in the key (different policies = different results)
	_, _ = hasher.Write([]byte(e.policyPath))
	_, _ = hasher.Write([]byte(e.query))

	// Hash cart items in a deterministic order
	for _, item := range cart.Items {
		_, _ = hasher.Write([]byte(item.GoodID.String()))
		_, _ = fmt.Fprintf(hasher, "%d", item.Quantity) //nolint:errcheck // hash write best-effort
		_, _ = hasher.Write([]byte(item.Price.String()))
	}

	// Hash params in sorted order for determinism
	if params != nil {
		keys := make([]string, 0, len(params))
		for k := range params {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, k := range keys {
			_, _ = hasher.Write([]byte(k))
			_, _ = fmt.Fprintf(hasher, "%v", params[k]) //nolint:errcheck // hash write best-effort
		}
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// transformCartToInput converts the domain.Cart to the input format expected by OPA
func transformCartToInput(cart *domain.Cart, params map[string]any) map[string]any {
	items := make([]map[string]any, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, map[string]any{
			"productId": item.GoodID.String(),
			"quantity":  item.Quantity,
			"price":     item.Price.InexactFloat64(),
		})
	}

	return map[string]any{
		"items":  items,
		"params": params, // Include additional parameters if needed
	}
}

const float64Bits = 64

// parseOPAResult handles different types that OPA might return and converts them to float64
func parseOPAResult(value any) (float64, error) {
	switch val := value.(type) {
	case float64:
		return val, nil
	case string:
		parsed, err := strconv.ParseFloat(val, float64Bits)
		if err != nil {
			return 0.0, fmt.Errorf("%w: %w", ErrOPAResultInvalidStr, err)
		}

		return parsed, nil
	case json.Number:
		parsed, err := val.Float64()
		if err != nil {
			return 0.0, fmt.Errorf("%w: %w", ErrOPAResultInvalidNum, err)
		}

		return parsed, nil
	default:
		return 0.0, fmt.Errorf("%T: %w", val, ErrOPAResultUnexpectedType)
	}
}

// GetPolicyNames retrieves the names of all .rego files in the specified directories.
func GetPolicyNames(dirs ...string) ([]string, error) {
	var policyNames []string

	for _, dir := range dirs {
		// Use filepath.Glob to find all .rego files in the directory
		pattern := filepath.Join(dir, "*.rego")

		files, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %w", dir, ErrListRegoFiles, err)
		}

		for _, file := range files {
			// Extract the base name without the directory and extension
			base := filepath.Base(file)
			name := base[:len(base)-len(filepath.Ext(base))]
			policyNames = append(policyNames, name)
		}
	}

	return policyNames, nil
}
