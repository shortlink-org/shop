package weight

import (
	"testing"
)

func TestNewWeight(t *testing.T) {
	tests := []struct {
		name    string
		grams   int
		wantErr bool
		errType error
	}{
		{
			name:    "valid weight - 1 gram",
			grams:   1,
			wantErr: false,
		},
		{
			name:    "valid weight - 1000 grams (1 kg)",
			grams:   1000,
			wantErr: false,
		},
		{
			name:    "valid weight - maximum",
			grams:   MaxWeightGrams,
			wantErr: false,
		},
		{
			name:    "zero weight",
			grams:   0,
			wantErr: true,
			errType: ErrWeightZero,
		},
		{
			name:    "negative weight",
			grams:   -1,
			wantErr: true,
			errType: ErrWeightNegative,
		},
		{
			name:    "weight exceeds maximum",
			grams:   MaxWeightGrams + 1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weight, err := NewWeight(tt.grams)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewWeight() expected error but got none")
					return
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("NewWeight() error = %v, want %v", err, tt.errType)
				}
				if !weight.IsZero() {
					t.Errorf("NewWeight() should return zero weight on error, got %d", weight.Grams())
				}
			} else {
				if err != nil {
					t.Errorf("NewWeight() unexpected error: %v", err)
					return
				}
				if weight.Grams() != tt.grams {
					t.Errorf("NewWeight() grams = %d, want %d", weight.Grams(), tt.grams)
				}
			}
		})
	}
}

func TestWeight_Grams(t *testing.T) {
	want := 1500
	weight := MustNewWeight(want)
	got := weight.Grams()
	if got != want {
		t.Errorf("Weight.Grams() = %d, want %d", got, want)
	}
}

func TestWeight_Kilograms(t *testing.T) {
	tests := []struct {
		name     string
		grams    int
		expected float64
	}{
		{
			name:     "1000 grams = 1 kg",
			grams:    1000,
			expected: 1.0,
		},
		{
			name:     "1500 grams = 1.5 kg",
			grams:    1500,
			expected: 1.5,
		},
		{
			name:     "500 grams = 0.5 kg",
			grams:    500,
			expected: 0.5,
		},
		{
			name:     "2500 grams = 2.5 kg",
			grams:    2500,
			expected: 2.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weight := MustNewWeight(tt.grams)
			got := weight.Kilograms()
			if got != tt.expected {
				t.Errorf("Weight.Kilograms() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWeight_Add(t *testing.T) {
	tests := []struct {
		name    string
		w1      Weight
		w2      Weight
		want    int
		wantErr bool
	}{
		{
			name:    "add two valid weights",
			w1:      MustNewWeight(500),
			w2:      MustNewWeight(300),
			want:    800,
			wantErr: false,
		},
		{
			name:    "add weights at boundary",
			w1:      MustNewWeight(MaxWeightGrams - 100),
			w2:      MustNewWeight(100),
			want:    MaxWeightGrams,
			wantErr: false,
		},
		{
			name:    "add weights exceeds maximum",
			w1:      MustNewWeight(MaxWeightGrams - 100),
			w2:      MustNewWeight(200),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.w1.Add(tt.w2)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Weight.Add() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Weight.Add() unexpected error: %v", err)
					return
				}
				if got.Grams() != tt.want {
					t.Errorf("Weight.Add() = %d, want %d", got.Grams(), tt.want)
				}
			}
		})
	}
}

func TestWeight_Multiply(t *testing.T) {
	tests := []struct {
		name    string
		weight  Weight
		factor  int
		want    int
		wantErr bool
	}{
		{
			name:    "multiply by 2",
			weight:  MustNewWeight(500),
			factor:  2,
			want:    1000,
			wantErr: false,
		},
		{
			name:    "multiply by 1",
			weight:  MustNewWeight(500),
			factor:  1,
			want:    500,
			wantErr: false,
		},
		{
			name:    "multiply by zero",
			weight:  MustNewWeight(500),
			factor:  0,
			wantErr: true,
		},
		{
			name:    "multiply by negative",
			weight:  MustNewWeight(500),
			factor:  -1,
			wantErr: true,
		},
		{
			name:    "multiply exceeds maximum",
			weight:  MustNewWeight(MaxWeightGrams / 2),
			factor:  3,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.weight.Multiply(tt.factor)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Weight.Multiply() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Weight.Multiply() unexpected error: %v", err)
					return
				}
				if got.Grams() != tt.want {
					t.Errorf("Weight.Multiply() = %d, want %d", got.Grams(), tt.want)
				}
			}
		})
	}
}

func TestWeight_IsGreaterThan(t *testing.T) {
	w1 := MustNewWeight(1000)
	w2 := MustNewWeight(500)

	if !w1.IsGreaterThan(w2) {
		t.Errorf("Weight.IsGreaterThan() = false, want true")
	}
	if w2.IsGreaterThan(w1) {
		t.Errorf("Weight.IsGreaterThan() = true, want false")
	}
}

func TestWeight_IsLessThan(t *testing.T) {
	w1 := MustNewWeight(500)
	w2 := MustNewWeight(1000)

	if !w1.IsLessThan(w2) {
		t.Errorf("Weight.IsLessThan() = false, want true")
	}
	if w2.IsLessThan(w1) {
		t.Errorf("Weight.IsLessThan() = true, want false")
	}
}

func TestWeight_String(t *testing.T) {
	tests := []struct {
		name     string
		grams    int
		expected string
	}{
		{
			name:     "less than 1 kg",
			grams:    500,
			expected: "500 g",
		},
		{
			name:     "exactly 1 kg",
			grams:    1000,
			expected: "1.00 kg",
		},
		{
			name:     "more than 1 kg",
			grams:    1500,
			expected: "1.50 kg",
		},
		{
			name:     "2.5 kg",
			grams:    2500,
			expected: "2.50 kg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weight := MustNewWeight(tt.grams)
			got := weight.String()
			if got != tt.expected {
				t.Errorf("Weight.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWeight_Equality(t *testing.T) {
	w1 := MustNewWeight(1000)
	w2 := MustNewWeight(1000)
	w3 := MustNewWeight(500)

	// Two weights with same grams should be equal
	if w1 != w2 {
		t.Errorf("Weights with same grams should be equal")
	}

	// Two weights with different grams should not be equal
	if w1 == w3 {
		t.Errorf("Weights with different grams should not be equal")
	}
}

func TestMustNewWeight(t *testing.T) {
	t.Run("valid weight", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustNewWeight() panicked on valid weight: %v", r)
			}
		}()
		weight := MustNewWeight(1000)
		if weight.Grams() != 1000 {
			t.Errorf("MustNewWeight() = %d, want 1000", weight.Grams())
		}
	})

	t.Run("invalid weight panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustNewWeight() should panic on invalid weight")
			}
		}()
		_ = MustNewWeight(-1)
	})
}

func TestWeight_IsZero(t *testing.T) {
	weight := MustNewWeight(1000)
	if weight.IsZero() {
		t.Errorf("Weight.IsZero() = true for non-zero weight")
	}
}

