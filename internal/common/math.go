package common

// PctChange returns the percentage change from previous to current as an integer.
// Returns 100 if previous is zero and current is positive, 0 if both are zero.
func PctChange(current, previous float64) int {
	if previous == 0 {
		if current > 0 {
			return 100
		}
		return 0
	}
	return int(((current - previous) / previous) * 100)
}

// Fmin returns the smaller of two float64 values.
func Fmin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Fmax returns the larger of two float64 values.
func Fmax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
