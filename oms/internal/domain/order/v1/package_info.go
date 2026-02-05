package v1

// PackageInfo contains package physical characteristics for order delivery.
type PackageInfo struct {
	weightKg float64
}

// NewPackageInfo creates a new PackageInfo value object.
func NewPackageInfo(weightKg float64) PackageInfo {
	return PackageInfo{weightKg: weightKg}
}

// GetWeightKg returns the weight in kilograms.
func (p PackageInfo) GetWeightKg() float64 {
	return p.weightKg
}

// IsValid checks if the package info is valid (weight > 0).
func (p PackageInfo) IsValid() bool {
	return p.weightKg > 0
}

