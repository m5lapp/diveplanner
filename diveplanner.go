package diveplanner

import (
	"fmt"
	"math"
	"time"

	"github.com/m5lapp/diveplanner/buhlmann"
	"github.com/m5lapp/diveplanner/gasmix"
	"github.com/m5lapp/diveplanner/helpers"
)

const (
	// Common/useful dive factor multipliers.
	DiveFactorEasy          float64 = 1.5
	DiveFactorModerate      float64 = 1.8
	DiveFactorTough         float64 = 2.0
	DiveFactorStressful     float64 = 2.5
	DiveFactorSeriousStress float64 = 3.0

	otuRepetitiveDiveLimit float64 = 300.0
	otuSingleDiveLimit     float64 = 850.0

	safetyStopDepth float64 = 5.0
)

type DivePlanStop struct {
	Depth        float64 `bson:"depth" json:"depth"`
	Duration     float64 `bson:"duration" json:"duration"`
	IsTransition bool    `bson:"is_transition" json:"is_transition"`
	Comment      string  `bson:"comment" json:"comment"`
}

// GasRequirement() calculates the amount of breathing gas that a diver with a
// given Surface Air Consumption (SAC) rate in litres/minute requires for a
// given stop.
func (s *DivePlanStop) GasRequirement(sacRate, diveFactor float64) float64 {
	p := helpers.Pressure(s.Depth)
	return p * sacRate * diveFactor * float64(s.Duration)
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
	DiveFactor      float64         `bson:"dive_factor" json:"dive_factor"`
	GasMix          *gasmix.GasMix  `bson:"nitrox_mix" json:"nitrox_mix"`
	MaxPPO2         float64         `bson:"max_ppo2" json:"max_ppo2"`
	Stops           []*DivePlanStop `bson:"stops" json:"stops"`
}

// transitionDuration() calculates the amount of time in minutes required to
// transition from one depth in metres to another at the configured ascent or
// descent rate, rounded up to the nearest minute for conservatism. If the
// difference between the two depths is less than half a metre, then that is not
// considered a transition and zero is returned.
func (dp *DivePlan) transitionDuration(fromD, toD float64) float64 {
	var time float64 = 0.0
	depthDelta := toD - fromD

	// If the depthDelta is less than half a metre either way, then time remains
	// set to zero as it's not worth spending a minute to get there.
	if depthDelta >= 0.5 {
		// The depth delta is positive (or zero) which means we are descending.
		time = math.Abs(depthDelta / dp.DescentRate)
	} else if depthDelta <= -0.5 {
		// The depth delta is negative which means we are ascending.
		time = math.Abs(depthDelta / dp.AscentRate)
	}

	// For conservatism, round the time up to the nearest minute.
	return math.Ceil(time)
}

// transitionStop() returns a DivePlanStop that represents a transition from one
// depth to another; it's Depth value is the average of the two.
func (dp *DivePlan) transitionStop(fromD, toD float64) *DivePlanStop {
	dir := "Descent"
	if toD < fromD {
		dir = "Ascent"
	}

	return &DivePlanStop{
		Duration:     dp.transitionDuration(fromD, toD),
		Depth:        math.Abs((fromD + toD) / 2.0),
		IsTransition: true,
		Comment:      fmt.Sprintf("%s from %.1fm to %.1fm", dir, fromD, toD),
	}
}

// DiveProfile() returns a slice of DivePlanStops with all the stops in the dive
// plan and the transition to each stop from the previous one. Finally it will
// include the transition from the last stop back to the surface.
func (dp *DivePlan) DiveProfile() []*DivePlanStop {
	var currDepth float64
	var profile []*DivePlanStop

	for _, s := range dp.Stops {
		// Check that the stop is a valid stop, otherwise, don't include it.
		if s.Depth > 0.0 && s.Duration > 0.0 {
			t := dp.transitionStop(currDepth, s.Depth)
			profile = append(profile, t, s)
			currDepth = s.Depth
		}
	}

	// Include the final transition back to the surface and update dp.Stops.
	if len(profile) > 0 {
		t := dp.transitionStop(currDepth, 0.0)
		profile = append(profile, t)
	}

	return profile
}

// MaxDepth() returns the depth at the deepest point of the dive plan or zero if
// there are no stops in the plan.
func (dp *DivePlan) MaxDepth() float64 {
	maxDepth := 0.0
	for _, s := range dp.Stops {
		if s.Depth > maxDepth {
			maxDepth = s.Depth
		}
	}
	return maxDepth
}

// Runtime() sums the duration of each stage in the plan including the ascents
// and descents and returns that value in minutes.
func (dp *DivePlan) Runtime() float64 {
	var runtime float64

	for _, s := range dp.DiveProfile() {
		runtime += s.Duration
	}

	return runtime
}

// DSRTable() returns a slice of arrays of three values representing the Depth,
// Stop and Run (DSR) values for the dive. This is a list of each stop in the
// dive, its duration and the runtime at the end of the stop.
func (dp *DivePlan) DSRTable() *[][3]float64 {
	var dsrTable [][3]float64
	var run float64

	for _, s := range dp.DiveProfile() {
		run += s.Duration
		if !s.IsTransition {
			dsrTable = append(dsrTable, [3]float64{s.Depth, s.Duration, run})
		}
	}

	return &dsrTable
}

// Pulmonary Oxygen Toxicity calculates the number of Oxygen Tolerence Units
// (OTUs) for the dive. One OTU is equivalent to breathing 100% Oxygen at 1 bar
// for 1 minute. The single dive limit is 850 OTU on day 1 and 300 OTU for
// repetitive dives on day 2+.
func (dp *DivePlan) POT() float64 {
	var otu float64

	// Sum the OTUs for each stage in the profile.
	for _, s := range dp.DiveProfile() {
		otu += dp.GasMix.PPO2(s.Depth) * s.Duration
	}

	return otu
}

// MinGas() returns the amount of gas required to get two divers (or one if
// diving solo) to the surface in an emergency from the deepest part of the dive
// with a safety stop. For solo dives, the minimum gas is still doubled as it is
// required to be available from two independent gas sources.
func (dp *DivePlan) MinGas() float64 {
	const buddyMultiplier float64 = 2.0
	maxDepth := dp.MaxDepth()
	maxPressure := helpers.Pressure(maxDepth)
	avgPressure := helpers.Pressure(maxDepth / 2.0)
	stopPressure := helpers.Pressure(safetyStopDepth)
	ascentTime := dp.transitionDuration(maxDepth, 0.0)

	// Account for elevated breathing rate in an emergency with a budy.
	elevatedSACRate := dp.SACRate * dp.DiveFactor * buddyMultiplier * 1.5

	// Allow one minute to sort yourself out at the maximum depth.
	preperationGas := 1.0 * maxPressure * elevatedSACRate

	// Gas required for the ascent to reach the surface with an elevated SAC
	// rate breathing at the average depth ambient pressure.
	ascentGas := ascentTime * avgPressure * elevatedSACRate

	// Include three minutes at the safety stop depth.
	stopGas := 3.0 * stopPressure * elevatedSACRate

	return preperationGas + ascentGas + stopGas
}

// GasAvailable() returns the total amount of gas available to the diver with
// the equipment configuration specified.
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
// contingency and so should not be used without using additonal gas planning.
func (dp *DivePlan) baseGasRequired() float64 {
	var gasRequired float64

	// Calculate the gas required for each stage in the profle with the given
	// SAC rate and dive factor.
	for _, s := range dp.DiveProfile() {
		gasRequired += s.GasRequirement(dp.SACRate, dp.DiveFactor)
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
	var prevDepth float64

	for i, s := range dp.Stops {
		// Check if i != 0 so that the first descent is not included.
		if s.Depth > prevDepth && i != 0 {
			return true
		}
		prevDepth = s.Depth
	}

	return false
}

// WithinNDLs() returns true if the dive stays with No-Decompression Limits.
// That is, no mandatory decompression stops are required.
func (dp *DivePlan) WithinNDLs() bool {
	var bmann *buhlmann.ZhlModel = buhlmann.New(dp.GasMix, buhlmann.ZHL16C)
	var prevDepth float64

	for _, s := range dp.Stops {
		if !s.IsTransition {
			rate := dp.DescentRate
			if helpers.DescOrAsc(prevDepth, s.Depth) == -1.0 {
				rate = dp.AscentRate
			}

			// Simulate the transition to the stop depth and check our NDLs.
			bmann.TransitionCalc(s.Depth, rate)
			if bmann.GetNDL() <= 0 {
				return false
			}

			// Simulate the stop, then check our NDLs at the end of it.
			bmann.StopCalc(s.Duration)
			if bmann.GetNDL() <= 0 {
				return false
			}

			prevDepth = s.Depth
		}
	}

	return true
}

// DiveIsPossible() returns a boolean value that indicates whether or not the
// dive plan, is possible as it is currently configured, taking various factors
// into account.
func (dp *DivePlan) DiveIsPossible() bool {
	isSawTooth := dp.IsSawToothProfile()
	sufficientGas := dp.GasSpare() >= 0.0
	withinMOD := dp.MaxDepth() <= dp.GasMix.MOD(dp.MaxPPO2)
	withinNDLs := dp.WithinNDLs()
	return !isSawTooth && sufficientGas && withinMOD && withinNDLs
}

type ProfileSample struct {
	Time  int
	Depth float64
	NDL   int
}

// ChartProfile() returns a slice of ProfileSamples that contains the time in
// seconds, depth and NDLs at each step of the dive in increments of the
// resolution parameter provided, in seconds.
func (dp *DivePlan) ChartProfile(resolution int) []ProfileSample {
	var profile []ProfileSample
	var bmann *buhlmann.ZhlModel = buhlmann.New(dp.GasMix, buhlmann.ZHL16B)
	var currDepth float64
	var currTime int
	profile = append(profile, ProfileSample{currTime, currDepth, bmann.GetNDL()})

	for _, s := range dp.Stops {
		currTime, currDepth = dp.walkTransition(currDepth, s.Depth, currTime, resolution, bmann, &profile)
		samples := (float64(s.Duration) * 60.0) / float64(resolution)
		for i := 0; i < int(math.Floor(samples)); i++ {
			// Reasign currDepth to the Stop depth to account for any
			// floating-point errors.
			currDepth = s.Depth
			currTime += resolution
			bmann.StopCalc(float64(resolution) / 60.0)
			ndl := bmann.GetNDL()
			profile = append(profile, ProfileSample{currTime, currDepth, ndl})
		}
	}

	// Final transition back to the surface.
	currTime, currDepth = dp.walkTransition(currDepth, 0.0, currTime, resolution, bmann, &profile)
	return profile
}

// walkTransition() calculates the depth and NDLs at each step of res seconds
// through a transition from one depth to another and appends them to the
// profile slice provided. At the end, it returns the final depth and time so
// that the calling function knows where the dive profile is up to.
func (dp *DivePlan) walkTransition(currDepth, targetDepth float64,
	currTime, res int, bmann *buhlmann.ZhlModel,
	profile *[]ProfileSample) (int, float64) {
	// The distance in metres between the current depth and the next one. A
	// positive value means descending, negative is ascending.
	depthDelta := targetDepth - currDepth
	// The amount of time in minutes it will take to transfer to the next stop.
	tranDuration := dp.transitionDuration(currDepth, targetDepth)
	if tranDuration == 0.0 {
		// The delta between the two depths is not large enough to be considered
		// a proper transition.
		return currTime, currDepth
	}

	// The number of samples to get to the next stop.
	samples := (tranDuration * 60.0) / float64(res)
	// The difference in depth between each consecutive sample.
	sampleDelta := depthDelta / samples
	// The actual rate of descent/ascent which will be slower than those defined
	// in the dive plan given that the transition times are rounded up for
	// conservatism.
	rate := math.Abs(depthDelta / tranDuration)

	for i := 0; i < int(math.Floor(samples)); i++ {
		currDepth += sampleDelta
		currTime += res
		bmann.TransitionCalc(currDepth, rate)
		ndl := bmann.GetNDL()
		*profile = append(*profile, ProfileSample{currTime, currDepth, ndl})
	}

	return currTime, currDepth
}
