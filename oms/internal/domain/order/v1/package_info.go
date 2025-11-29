package v1

// PackageInfo contains package physical characteristics for order delivery.
type PackageInfo struct {
	// weightKg is the weight of the package in kilograms
	weightKg float64
	// dimensions is the package dimensions in "LxWxH" format (e.g., "30x20x15")
	dimensions string
}

// NewPackageInfo creates a new PackageInfo value object.
func NewPackageInfo(weightKg float64, dimensions string) PackageInfo {
	return PackageInfo{
		weightKg:   weightKg,
		dimensions: dimensions,
	}
}

// GetWeightKg returns the weight in kilograms.
func (p PackageInfo) GetWeightKg() float64 {
	return p.weightKg
}

// GetDimensions returns the dimensions string.
func (p PackageInfo) GetDimensions() string {
	return p.dimensions
}

// IsValid checks if the package info is valid (weight > 0).
func (p PackageInfo) IsValid() bool {
	return p.weightKg > 0
}

