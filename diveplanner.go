package diveplanner

import (
	"fmt"
	"math"
	"time"

	"github.com/m5lapp/diveplanner/gasmix"
	"github.com/m5lapp/diveplanner/helpers"
)

const (
	otuRepetitiveDiveLimit float64 = 300.0
	otuSingleDiveLimit     float64 = 850.0
	safetyStopDepth        float64 = 5.0
)

type DivePlanStop struct {
	Stop     int     `bson:"stop" json:"stop"`
	Depth    float64 `bson:"depth" json:"depth"`
	Duration float64 `bson:"duration" json:"duration"`
	Comment  string  `bson:"comment" json:"comment"`
}

// GasRequirement() calculates the amount of breathing gas that a diver with a
// given Surface Air Consumption (SAC) rate in litres/minute requires for a
// given stop.
func (s *DivePlanStop) GasRequirement(sacRate float64) float64 {
	return helpers.Pressure(s.Depth) * sacRate * float64(s.Duration)
}

type DivePlan struct {
	Created         time.Time       `bson:"created" json:"created"`
	Updated         time.Time       `bson:"updated" json:"updated"`
	Name            string          `bson:"name" json:"name"`
	Notes           string          `bson:"notes" json:"notes"`
	IsSoloDive      bool            `bson:"is_solo_dive" json:"is_solo_dive"`
	DescentRate     float64         `bson:"descent_rate" json:"descent_rate"`
	AscentRate      float64         `bson:"asent_rate" json:"asent_rate"`
	SACRate         float64         `bson:"sac_rate" json:"sac_rate"`
	TankCount       int             `bson:"tank_count" json:"tank_count"`
	TankCapacity    float64         `bson:"tank_capacity" json:"tank_capacity"`
	WorkingPressure int             `bson:"working_pressure" json:"working_pressure"`
	NixtroxMix      gasmix.GasMix   `bson:"nitrox_mix" json:"nitrox_mix"`
	MaxPPO2         float64         `bson:"max_ppo2" json:"max_ppo2"`
	Stops           []*DivePlanStop `bson:"stops" json:"stops"`
}

// transitionDuration() calculates the amount of time in minutes required to
// transition from depth d1 in metres to depth d2 in metres at the configured
// ascent or descent rate.
func (dp *DivePlan) transitionDuration(d1, d2 float64) (float64, error) {
	depthDelta := d2 - d1
	if depthDelta == 0.0 {
		return 0.0, fmt.Errorf("calcTransition: no delta between the depths "+
			"provided: d1: %f, d2: %f.", d1, d2)
	} else if depthDelta > 0.0 {
		// The depth delta is positive which means we are descending.
		return depthDelta / dp.DescentRate, nil
	} else {
		// The depth delta is negative which means we are ascending.
		return depthDelta / dp.AscentRate, nil
	}
}

// MaxDepth() simply returns the depth at the deepest point of the dive plan.
func (dp *DivePlan) MaxDepth() float64 {
	maxDepth := 0.0
	for _, s := range dp.Stops {
		if s.Depth > maxDepth {
			maxDepth = s.Depth
		}
	}
	return maxDepth
}

// Runtime() simply sums the duration of each stop in the plan and returns that
// value in minutes.
func (dp *DivePlan) Runtime() float64 {
	runtime := 0.0
	for _, s := range dp.Stops {
		runtime += s.Duration
	}
	return runtime
}

// Pulmonary Oxygen Toxicity calculates the number of Oxygen Tolerence Units
// (OTU) for the dive. One OTU is equivalent to breathing 100% Oxygen at 1 bar
// for 1 minute. The single dive limit is 850 OTU on day 1 and 300 OTU for
// repetitive dives on day 2+.
func (dp *DivePlan) POT() (otu float64) {
	otu = 0.0
	for _, s := range dp.Stops {
		ppo2 := dp.NixtroxMix.PPO2(s.Depth)
		otu += ppo2 * float64(s.Duration)
	}
	return otu
}

// MinGas() returns the amount of gas required to get two divers (or one if
// diving solo) to the surface in an emergency from the deepest part of the dive
// with a safety stop.
func (dp *DivePlan) MinGas() float64 {
	maxDepth := dp.MaxDepth()
	maxPressure := helpers.Pressure(maxDepth)
	avgPressure := helpers.Pressure(maxDepth / 2.0)
	stopPressure := helpers.Pressure(safetyStopDepth)
	ascentTime := maxDepth / dp.AscentRate
	buddyMultiplier := 2.0
	if dp.IsSoloDive {
		buddyMultiplier = 1.0
	}
	// Account for elevated breathing rate in an emergency.
	elevatedSACRate := dp.SACRate * buddyMultiplier * 1.5

	// Allow one minute to sort yourself out at the maximum depth.
	preperationGas := 1.0 * maxPressure * elevatedSACRate

	// Gas required for the ascent to reach the surface with an elevated SAC
	// rate breathing at the average depth ambient pressure.
	ascentGas := ascentTime * avgPressure * elevatedSACRate

	// Include three minutes at the safety stop depth.
	stopGas := 3.0 * stopPressure * elevatedSACRate

	return preperationGas + ascentGas + stopGas
}

// GasAvailable returns the total amount of gas available to the diver with the
// equipment configuration specified.
func (dp *DivePlan) GasAvailable() float64 {
	return float64(dp.TankCount) * dp.TankCapacity * float64(dp.WorkingPressure)
}

// WorkingGas() is the gas available across all tanks once the minimum gas has
// been accounted for.
func (dp *DivePlan) WorkingGas() float64 {
	return dp.GasAvailable() - (dp.MinGas() * float64(dp.TankCount))
}

// baseGasRequired() calculates the amount of gas required for the dive as
// planned; the descent, the ascent and each stop. It does not include any
// contingency and so should not be used without applying the rule of thirds.
func (dp *DivePlan) baseGasRequired() float64 {
	gasRequired := 0.0
	for _, s := range dp.Stops {
		gasRequired += s.GasRequirement(dp.SACRate)
	}
	return gasRequired
}

// GasRequired() applies the rule of thirds to calculate the amount of gas
// required for the dive as configured; one-third out, one-third back and
// one-third in reserve.
func (dp *DivePlan) GasRequired() float64 {
	return dp.baseGasRequired() * 1.5
}

// GasSpare() calculates how much gas will be remaining across all tanks at the
// end of the planned dive.
func (dp *DivePlan) GasSpare() float64 {
	return dp.WorkingGas() - dp.GasRequired()
}

// IsSawToothProfile() indicates if the dive plan has a saw-tooth profile, that
// is, there are some stops in the dive plan that are deeper than the ones
// preceeding it.
func (dp *DivePlan) IsSawToothProfile() bool {
	prevDepth := 0.0
	for i, s := range dp.Stops {
		// Check if i != 0 so that the first descent is not included.
		if s.Depth > prevDepth && i != 0 {
			return true
		}
		prevDepth = s.Depth
	}
	return false
}

// DiveIsPossible() returns a boolean value that indicates whether or not the
// dive plan, is possible as it is currently configured, taking various factors
// into account.
func (dp *DivePlan) DiveIsPossible() bool {
	isSawTooth := dp.IsSawToothProfile()
	sufficientGas := dp.GasSpare() >= 0.0
	withinMOD := dp.MaxDepth() <= dp.NixtroxMix.MOD(dp.MaxPPO2)
	return !isSawTooth && sufficientGas && withinMOD
}

type ProfileSample struct {
	Time  int
	Depth float64
}

func (dp *DivePlan) ChartProfile(resolution int) []ProfileSample {
	var profile []ProfileSample
	currDepth := 0.0
	currTime := 0
	profile = append(profile, ProfileSample{currTime, currDepth})
	for _, s := range dp.Stops {
		currTime, currDepth = dp.walkTransition(currDepth, s.Depth, currTime, resolution, &profile)
		samples := (float64(s.Duration) * 60.0) / float64(resolution)
		for i := 0; i < int(math.Floor(samples)); i++ {
			// Reasign currDepth to the Stop depth to account for any
			// floating-point errors.
			currDepth = s.Depth
			currTime += resolution
			profile = append(profile, ProfileSample{currTime, currDepth})
		}
	}
	// Final transition back to the surface.
	currTime, currDepth = dp.walkTransition(currDepth, 0.0, currTime, resolution, &profile)
	return profile
}

func (dp *DivePlan) walkTransition(currDepth, targetDepth float64,
	currTime, res int, profile *[]ProfileSample) (int, float64) {
	// The distance in metres between the current depth and the next one. A
	// positive value means descending, negative is ascending.
	depthDelta := targetDepth - currDepth
	// The amount of time in seconds it will take to transfer to the next stop.
	transitionTime := math.Abs(depthDelta / dp.DescentRate * 60.0)
	// The number of samples to get to the next stop.
	samples := transitionTime / float64(res)
	// The difference in depth between each consecutive sample.
	sampleDelta := depthDelta / samples

	for i := 0; i < int(math.Floor(samples)); i++ {
		currDepth += sampleDelta
		currTime += res
		*profile = append(*profile, ProfileSample{currTime, currDepth})
	}
	return currTime, currDepth
}

func main() {
	var plan DivePlan = DivePlan{
		Created:         time.Now().UTC(),
		Updated:         time.Now().UTC(),
		Name:            "Sail Rock",
		Notes:           "Good, conservative dive plan.",
		IsSoloDive:      false,
		DescentRate:     9.0,
		AscentRate:      9.0,
		SACRate:         15.0,
		TankCount:       2,
		TankCapacity:    11.0,
		WorkingPressure: 200,
		NixtroxMix:      gasmix.GasMix{FO2: 0.32},
		MaxPPO2:         1.4,
		Stops: []*DivePlanStop{
			//{1, 12.5, 3, "Descent, depth is average"},
			{2, 25.0, 13, ""},
			{3, 18.0, 15, ""},
			{4, 12.0, 23, ""},
			{5, 5.0, 3, ""},
			//{6, 12.5, 3, "Ascent, depth is average"},
		},
	}

	fmt.Printf("Dive possible: %v\n", plan.DiveIsPossible())
	fmt.Printf("Min gas: %v\n", plan.MinGas())
	fmt.Printf("%v\n", plan)
	for _, s := range plan.ChartProfile(10) {
		fmt.Println(s.Time, s.Depth)
	}
}
