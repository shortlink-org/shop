package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shortlink-org/shop/pricer/internal/domain"
	"github.com/shortlink-org/shop/pricer/internal/usecases/cart/command/calculate_total"
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
		"totalTax":      total.TotalTax.StringFixed(2),
		"totalDiscount": total.TotalDiscount.StringFixed(2),
		"finalPrice":    total.FinalPrice.StringFixed(2),
		"policies":      total.Policies,
	}

	// Save the result
	filename := fmt.Sprintf("cart_result_%s.json", cart.CustomerID.String())
	if err := saveResultToFile(result, h.OutputDir, filename); err != nil {
		return fmt.Errorf("failed to save result for cart %s: %w", cartFile, err)
	}

	fmt.Printf("Final result saved to %s\n", filepath.Join(h.OutputDir, filename))

	return nil
}

// loadCart reads and unmarshals the cart JSON file
func loadCart(filePath string) (domain.Cart, error) {
	var cart domain.Cart

	file, err := os.ReadFile(filePath)
	if err != nil {
		return cart, fmt.Errorf("read cart file: %w", err)
	}

	if err := json.Unmarshal(file, &cart); err != nil {
		return cart, fmt.Errorf("unmarshal cart: %w", err)
	}

	return cart, nil
}

// saveResultToFile marshals the result to JSON and writes it to a file
func saveResultToFile(result map[string]any, outDir, filename string) error {
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	outputFile := filepath.Join(outDir, filename)

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}

	if err := os.WriteFile(outputFile, data, 0o644); err != nil {
		return fmt.Errorf("write result file: %w", err)
	}

	return nil
}
