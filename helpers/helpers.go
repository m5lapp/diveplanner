package helpers

import "math"

// EqualFloat64() compares two float64 values to see if they are as close enough
// together within a defined threshold to be considered equal.
func EqualFloat64(a, b float64) bool {
	const float64EqualityThreshold float64 = 1e-9
	return math.Abs(a-b) <= float64EqualityThreshold
}

// Depth() calculates the depth in metres for a given pressure in bar.
func Depth(pressure float64) float64 {
	return (pressure - 1.0) * 10.0
}

// Pressure() calculates the pressure in bar for a given depth in metres.
func Pressure(depth float64) float64 {
	return depth/10.0 + 1.0
}

// Pressure() calculates the pressure in bar for a given depth in metres.
func PressureChangePerMin(rate float64) float64 {
	return rate / 10.0
}

// DescOrAsc() indicates whether a diver is descending (positive pressure delta,
// 1.0 is returned), ascending (negative pressure delta, -1.0 is returned) or
// staying at the same level (0 is returned) when they move from one depth to
// another.
func DescOrAsc(fromD, toD float64) float64 {
	if EqualFloat64(fromD, toD) {
		return 0.0
	} else if fromD < toD {
		return 1.0
	} else {
		return -1.0
	}
}

func MetresToFeet(depth float64) float64 {
	return depth * 3.281
}

func FeetToMetres(depth float64) float64 {
	return depth / 3.281
}

func LitresToCubicFeet(volume float64) float64 {
	return volume * 0.03531
}

func CubicFeetToLitres(volume float64) float64 {
	return volume / 0.03531
}

func BarToPSI(pressure float64) float64 {
	return pressure * 14.5038
}

func PSIToBar(pressure float64) float64 {
	return pressure / 14.5038
}
