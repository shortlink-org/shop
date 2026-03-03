package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/shortlink-org/shop/pricer/internal/domain"
	"github.com/shortlink-org/shop/pricer/internal/usecases/cart/command/calculate_total"
)

const (
	decimalPlaces  = 2
	dirMode        = 0o750
	outputFileMode = 0o600
)

// CLIHandler handles command-line interactions
type CLIHandler struct {
	calculateTotalHandler *calculate_total.Handler
	OutputDir             string
}

// NewCLIHandler creates a new CLIHandler
func NewCLIHandler(calculateTotalHandler *calculate_total.Handler, outputDir string) *CLIHandler {
	return &CLIHandler{
		calculateTotalHandler: calculateTotalHandler,
		OutputDir:             outputDir,
	}
}

// Run processes a single cart file with provided parameters
func (h *CLIHandler) Run(cartFile string, discountParams, taxParams map[string]any) error {
	// Load the cart
	cart, err := loadCart(cartFile)
	if err != nil {
		return fmt.Errorf("failed to load cart %s: %w", cartFile, err)
	}

	// Calculate totals
	cmd := calculate_total.NewCommand(&cart, discountParams, taxParams)

	total, err := h.calculateTotalHandler.Handle(context.Background(), cmd)
	if err != nil {
		return fmt.Errorf("failed to calculate total for cart %s: %w", cartFile, err)
	}

	// Prepare the result map
	result := map[string]any{
		"customerId":    cart.CustomerID.String(),
		"totalTax":      total.TotalTax.StringFixed(decimalPlaces),
		"totalDiscount": total.TotalDiscount.StringFixed(decimalPlaces),
		"finalPrice":    total.FinalPrice.StringFixed(decimalPlaces),
		"policies":      total.Policies,
	}

	// Save the result
	filename := fmt.Sprintf("cart_result_%s.json", cart.CustomerID.String())

	err = saveResultToFile(result, h.OutputDir, filename)
	if err != nil {
		return fmt.Errorf("failed to save result for cart %s: %w", cartFile, err)
	}

	slog.Info("Final result saved", slog.String("path", filepath.Join(h.OutputDir, filename)))

	return nil
}

// loadCart reads and unmarshals the cart JSON file.
// FilePath must be a path under current dir or otherwise validated by the caller to avoid path traversal.
func loadCart(filePath string) (domain.Cart, error) {
	var cart domain.Cart

	file, err := os.ReadFile(filePath) //nolint:gosec // G304: filePath is validated by caller (CLI args)
	if err != nil {
		return cart, fmt.Errorf("read cart file: %w", err)
	}

	err = json.Unmarshal(file, &cart)
	if err != nil {
		return cart, fmt.Errorf("unmarshal cart: %w", err)
	}

	return cart, nil
}

// saveResultToFile marshals the result to JSON and writes it to a file
func saveResultToFile(result map[string]any, outDir, filename string) error {
	err := os.MkdirAll(outDir, dirMode)
	if err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	outputFile := filepath.Join(outDir, filename)

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}

	err = os.WriteFile(outputFile, data, outputFileMode)
	if err != nil {
		return fmt.Errorf("write result file: %w", err)
	}

	return nil
}
