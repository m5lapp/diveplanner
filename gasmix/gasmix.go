package gasmix

import (
	"fmt"
	"math"

	"github.com/m5lapp/diveplanner/helpers"
)

// GasMix respresents a breathing gas mixture with a given fraction of Helium
// (FHe), Nitrogen (FN2) Oxygen (FO2). The fraction of Nitrogen and/or Helium
// can be zero depending on the type of gas mixture (Air, Nitrox, pure O2 etc.).
type GasMix struct {
	FHe float64
	FN2 float64
	FO2 float64
}

// Custom type to represent the type of the gas mix,
type MixType int

const (
	Unknown MixType = iota
	Air
	Heliox
	Nitrox
	Trimix
)

func (mt MixType) String() string {
	switch mt {
	case Air:
		return "Air"
	case Heliox:
		return "Heliox"
	case Nitrox:
		return "Nitrox"
	case Trimix:
		return "Trimix"
	}
	return "Unknown Gas Mix Type"
}

// NewAirMix() is a convenience constructor for a gas mix of pure Air.
func NewAirMix() *GasMix {
	return &GasMix{FN2: 0.79, FO2: 0.21}
}

// NewNitroxMix() is a constructor for a Nitrox gas mix with a given Fraction of
// Oxygen. The Fraction of Nitrogen can then be calculated from this.
func NewNitroxMix(fo2 float64) (*GasMix, error) {
	if fo2 < 0.21 || fo2 > 1.0 {
		e := fmt.Errorf("gasmix: Invalid FO2 value (%f), should be between 0.21 and 1.0 inclusive", fo2)
		return nil, e
	}

	gm := GasMix{
		FN2: 1.0 - fo2,
		FO2: fo2,
	}
	return &gm, nil
}

// NewTrimixMix() is a constructor for a Trimix gas mix with a given Fraction of
// Oxygen and a given Fraction of Helium. The Fraction of Nitrogen can then be
// calculated from this.
func NewTrimixMix(fo2, fhe float64) (*GasMix, error) {
	if fo2 < 0.21 || fo2 > 0.98 {
		e := fmt.Errorf("gasmix: Invalid FO2 value (%f), should be between 0.21 and 0.98 inclusive", fo2)
		return nil, e
	}

	if fhe < 0.01 || fhe > 0.78 {
		e := fmt.Errorf("gasmix: Invalid FHe value (%f), should be between 0.21 and 0.78 inclusive", fhe)
		return nil, e
	}

	if fo2+fhe > 1.0 {
		e := fmt.Errorf("gasmix: Invalid FO2 (%f) and FHe (%f) values, total (%f) should not exceed 1.0", fo2, fhe, fo2+fhe)
		return nil, e
	}

	gm := GasMix{
		FHe: fhe,
		FN2: 1.0 - (fhe + fo2),
		FO2: fo2,
	}
	return &gm, nil
}

// NewHelioxMix() is a constructor for a Heliox gas mix with a given Fraction of
// Oxygen. The Fraction of Helium can then be calculated from this.
func NewHelioxMix(fo2 float64) (*GasMix, error) {
	if fo2 < 0.21 || fo2 >= 0.99 {
		e := fmt.Errorf("gasmix: Invalid FO2 value (%f), should be between 0.21 and 0.99 inclusive", fo2)
		return nil, e
	}

	gm := GasMix{
		FHe: 1.0 - fo2,
		FO2: fo2,
	}
	return &gm, nil
}

// NewNitroxBestMix() returns the Nitrox mix the maximises the Oxygen content
// without exceeding the maximum PPO2 specified at the deepest part of the dive.
// The result is floored to the nearest two decimal places for convenience and
// clarity.
func NewNitroxBestMix(depth, maxPPO2 float64) (*GasMix, error) {
	bestMix := maxPPO2 / helpers.Pressure(depth)
	bestMix = math.Floor(bestMix*100.0) / 100.0
	return NewNitroxMix(bestMix)
}

// MixType() returns the appropriate MixType constant for the gas mix,
func (gm *GasMix) MixType() MixType {
	if gm.FO2 == 0.21 && gm.FN2 == 0.79 && gm.FHe == 0.0 {
		return Air
	} else if gm.FHe > 0.0 {
		// The mix contains Helium so is either Heliox or Trimix.
		if gm.FN2 == 0.0 {
			return Heliox
		} else if gm.FN2 > 0.0 {
			return Trimix
		}
	} else if gm.FHe == 0.0 {
		// The mix does not contain Helium and has more than 0.21 Oxygen.
		return Nitrox
	}

	// Could not determine the gas mix type.
	return Unknown
}

// EAD() calculates the Nixtrox mix's Equivalent Air Depth in metres for a given
// depth in metres.
func (gm *GasMix) EAD(depth float64) float64 {
	// Use math.Abs() to handle the case where depth is represented as a
	// negative number. The result of the calculation is the same.
	d := math.Abs(depth)
	// Calculate the fraction of Nitrogen.
	fn2 := 1.0 - gm.FO2

	return (d+10.0)*fn2/0.79 - 10.0
}

// MOD() calculates the gas mix's Maximum Operating Depth in metres for a given
// maximum Partial Pressure of Oxygen in bar.
func (gm *GasMix) MOD(maxPPO2 float64) float64 {
	mod := 10.0 * (maxPPO2/gm.FO2 - 1.0)
	// Round the result for clarity.
	return math.Round(mod)
}

// PPHe() returns the Partial Pressure of Helium for the gas mix at the given
// depth in metres.
func (gm *GasMix) PPHe(depth float64) float64 {
	// Use math.Abs() to handle the case where depth is represented as a
	// negative number. The result of the calculation is the same.
	d := math.Abs(depth)
	return helpers.Pressure(d) * gm.FO2
}

// PPN2() returns the Partial Pressure of Nitrogen for the Gas mix at the given
// depth in metres.
func (gm *GasMix) PPN2(depth float64) float64 {
	// Use math.Abs() to handle the case where depth is represented as a
	// negative number. The result of the calculation is the same.
	d := math.Abs(depth)
	return helpers.Pressure(d) * gm.FN2
}

// PPO2() returns the Partial Pressure of Oxygen for the gas mix at the given
// depth in metres.
func (gm *GasMix) PPO2(depth float64) float64 {
	// Use math.Abs() to handle the case where depth is represented as a
	// negative number. The result of the calculation is the same.
	d := math.Abs(depth)
	return helpers.Pressure(d) * gm.FO2
}
