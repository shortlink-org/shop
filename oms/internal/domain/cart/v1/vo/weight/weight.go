package weight

import (
	"errors"
	"fmt"
)

// Weight validation errors
var (
	ErrWeightZero     = errors.New("weight must be greater than zero")
	ErrWeightNegative = errors.New("weight cannot be negative")
)

// Constants for weight validation bounds
const (
	// MinWeightGrams is the minimum valid weight in grams
	MinWeightGrams int = 1
	// MaxWeightGrams is the maximum valid weight in grams (1 ton = 1,000,000 grams)
	MaxWeightGrams int = 1_000_000
)

// Weight represents a weight value object.
// Weight is stored in grams as an integer for precision and performance.
// A value object is immutable and defined by its attributes.
// Two weights are considered equal if they have the same value in grams.
type Weight struct {
	// grams is the weight value in grams
	grams int
}

// NewWeight creates a new Weight value object with validation.
//
// Args:
//   - grams: Weight in grams (must be positive)
//
// Returns:
//   - Weight: The validated weight value object
//   - error: Error if weight is invalid
//
// Example:
//
//	weight, err := vo.NewWeight(1000) // 1 kilogram
//	if err != nil {
//	    return err
//	}
func NewWeight(grams int) (Weight, error) {
	if grams < MinWeightGrams {
		if grams < 0 {
			return Weight{}, ErrWeightNegative
		}
		return Weight{}, ErrWeightZero
	}
	if grams > MaxWeightGrams {
		return Weight{}, fmt.Errorf("weight %d grams exceeds maximum allowed %d grams", grams, MaxWeightGrams)
	}

	return Weight{
		grams: grams,
	}, nil
}

// MustNewWeight creates a new Weight value object or panics if invalid.
// Use this only when you are certain the weight is valid.
func MustNewWeight(grams int) Weight {
	w, err := NewWeight(grams)
	if err != nil {
		panic(fmt.Sprintf("invalid weight: %v", err))
	}
	return w
}

// Grams returns the weight in grams.
func (w Weight) Grams() int {
	return w.grams
}

// Kilograms returns the weight in kilograms as a float64.
func (w Weight) Kilograms() float64 {
	return float64(w.grams) / 1000.0
}

// Add adds another weight to this weight and returns a new Weight.
// Returns an error if the result exceeds MaxWeightGrams.
func (w Weight) Add(other Weight) (Weight, error) {
	totalGrams := w.grams + other.grams
	if totalGrams > MaxWeightGrams {
		return Weight{}, fmt.Errorf("combined weight %d grams exceeds maximum allowed %d grams", totalGrams, MaxWeightGrams)
	}
	return Weight{grams: totalGrams}, nil
}

// Multiply multiplies this weight by a factor and returns a new Weight.
// Factor must be positive.
// Returns an error if the result exceeds MaxWeightGrams or factor is invalid.
func (w Weight) Multiply(factor int) (Weight, error) {
	if factor <= 0 {
		return Weight{}, errors.New("multiplication factor must be positive")
	}
	resultGrams := w.grams * factor
	if resultGrams > MaxWeightGrams {
		return Weight{}, fmt.Errorf("multiplied weight %d grams exceeds maximum allowed %d grams", resultGrams, MaxWeightGrams)
	}
	return Weight{grams: resultGrams}, nil
}

// IsGreaterThan checks if this weight is greater than another weight.
func (w Weight) IsGreaterThan(other Weight) bool {
	return w.grams > other.grams
}

// IsLessThan checks if this weight is less than another weight.
func (w Weight) IsLessThan(other Weight) bool {
	return w.grams < other.grams
}

// IsZero checks if the weight is zero (should not happen with valid Weight).
func (w Weight) IsZero() bool {
	return w.grams == 0
}

// String returns the string representation of the weight.
func (w Weight) String() string {
	if w.grams >= 1000 {
		return fmt.Sprintf("%.2f kg", w.Kilograms())
	}
	return fmt.Sprintf("%d g", w.grams)
}

