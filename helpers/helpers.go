package helpers

import "math"

const float64EqualityThreshold = 1e-9

// EqualFloat64() compares two float64 values to see if they are as close enough
// together within a defined threshold to be considered equal.
func EqualFloat64(a, b float64) bool {
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
