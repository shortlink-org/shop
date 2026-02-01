package policy_evaluator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/open-policy-agent/opa/rego"

	logger "github.com/shortlink-org/go-sdk/logger"
	"github.com/shortlink-org/shop/pricer/internal/domain"
)

const (
	// Cache configuration for OPA evaluation results
	cacheNumCounters = 10_000    // track 10k evaluations
	cacheMaxCost     = 1_000_000 // ~1MB (results are small float64 values)
	cacheBufferItems = 64
	cacheTTL         = 30 * time.Minute // pricing rules don't change frequently
)

// PolicyEvaluator interface as defined
type PolicyEvaluator interface {
	Evaluate(ctx context.Context, cart *domain.Cart, params map[string]interface{}) (float64, error)
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

func NewOPAEvaluator(log logger.Logger, policyPath string, query string) (*OPAEvaluator, error) {
	// Log the policy path and query
	log.Info("Initializing OPA evaluator",
		slog.String("policy_path", policyPath),
		slog.String("query", query),
	)

	// Check if the policy directory exists
	if _, err := os.Stat(policyPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("policy directory does not exist: %s", policyPath)
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
func (e *OPAEvaluator) Evaluate(ctx context.Context, cart *domain.Cart, params map[string]interface{}) (float64, error) {
	// Generate cache key from cart and params
	cacheKey := e.generateCacheKey(cart, params)

	// Check L1 cache first
	if cachedResult, found := e.cache.Get(cacheKey); found {
		return cachedResult, nil
	}

	// Cache miss - evaluate the policy
	input := transformCartToInput(cart, params)

	rs, err := e.preparedQuery.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return 0.0, fmt.Errorf("OPA evaluation error: %w", err)
	}

	if len(rs) == 0 {
		return 0.0, nil // No result from policy
	}

	// Assuming the policy returns a single value
	expr := rs[0].Expressions[0].Value
	result, err := parseOPAResult(expr)
	if err != nil {
		return 0.0, err
	}

	// Store in L1 cache (cost=1 since float64 is small)
	e.cache.SetWithTTL(cacheKey, result, 1, cacheTTL)

	return result, nil
}

// generateCacheKey creates a deterministic hash key from cart and params.
func (e *OPAEvaluator) generateCacheKey(cart *domain.Cart, params map[string]interface{}) string {
	h := sha256.New()

	// Include policy path in the key (different policies = different results)
	h.Write([]byte(e.policyPath))
	h.Write([]byte(e.query))

	// Hash cart items in a deterministic order
	for _, item := range cart.Items {
		h.Write([]byte(item.GoodID.String()))
		h.Write([]byte(fmt.Sprintf("%d", item.Quantity)))
		h.Write([]byte(item.Price.String()))
		h.Write([]byte(item.Brand))
	}

	// Hash params in sorted order for determinism
	if params != nil {
		keys := make([]string, 0, len(params))
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			h.Write([]byte(k))
			h.Write([]byte(fmt.Sprintf("%v", params[k])))
		}
	}

	return hex.EncodeToString(h.Sum(nil))
}

// transformCartToInput converts the domain.Cart to the input format expected by OPA
func transformCartToInput(cart *domain.Cart, params map[string]interface{}) map[string]interface{} {
	var items []map[string]interface{}
	for _, item := range cart.Items {
		items = append(items, map[string]interface{}{
			"goodId":   item.GoodID.String(), // Convert UUID to string
			"quantity": item.Quantity,
			"price":    item.Price.InexactFloat64(), // Convert decimal to float64
			"brand":    item.Brand,
		})
	}

	return map[string]interface{}{
		"items":  items,
		"params": params, // Include additional parameters if needed
	}
}

// parseOPAResult handles different types that OPA might return and converts them to float64
func parseOPAResult(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case string:
		// Attempt to parse string to float64
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0.0, fmt.Errorf("invalid string format for result: %v", err)
		}
		return parsed, nil
	case json.Number:
		parsed, err := v.Float64()
		if err != nil {
			return 0.0, fmt.Errorf("invalid json.Number format for result: %v", err)
		}
		return parsed, nil
	default:
		return 0.0, fmt.Errorf("unexpected type for result: %T", v)
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
			return nil, fmt.Errorf("failed to list .rego files in %s: %v", dir, err)
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
