# Diveplanner
Diveplanner is a pure-Go library for planning SCUBA dives.

## WARNING
This library was developed solely as a learning exercise. It is not intended to, and indeed SHOULD NOT, be used in the planning of any real SCUBA dives. The code is provided as-is and has not been rigorously tested in any way. Additionally, the Bühlmann algorithm that the library implements is generally regarded as being too liberal for real-life dive planning and should not be used without significant modification to make it more conservative.

## Usage
The library can be used as follows:

```
import (
    "github.com/m5lapp/diveplanner"
    "github.com/m5lapp/diveplanner/buhlmann"
    "github.com/m5lapp/diveplanner/gasmix"
)

func main() {
	gm, _ := gasmix.NewNitroxMix(0.32)
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
		GasMix:          gm,
		MaxPPO2:         1.4,
		Stops: []*DivePlanStop{
			{30.0, 20, false, ""},
			{18.0, 15, false, ""},
			{12.0, 23, false, ""},
			{5.0, 3, false, "Safety stop"},
		},
	}

	fmt.Printf("Dive possible: %v\n", plan.DiveIsPossible())
	fmt.Printf("Min gas: %v\n", plan.MinGas())
	fmt.Printf("Runtime: %v\n", plan.Runtime())
	fmt.Printf("%v\n", plan)
	for _, s := range plan.ChartProfile(10) {
		fmt.Println(s.Time, s.Depth, s.NDL)
	}
}
```

## Buhlmann Decompression Algorithm
The diveplanner/buhlmann module implements the [Bühlmann ZH-L16 algorithm](https://en.wikipedia.org/wiki/B%C3%BChlmann_decompression_algorithm) for tracking inert gas loading in a diver's tissues. This can be used stand-alone from the rest of the library.

Three sets of coefficient values are available:

 * ZHL16A - The original coefficient set developed by Dr. Albert Bühlmann
 * ZHL16B - A slightly more conservative version of the initial algorithm
 * ZHL16C - Suitable for use in dive computers

The Bühlmann library can be used as follows:

```
import (
    "github.com/m5lapp/diveplanner/buhlmann"
    "github.com/m5lapp/diveplanner/gasmix"
)

// Create a breathing gas mixture to use in the algorithm, e.g. 32% Nitrox:
gm, err := gasmix.NewNitroxMix(0.32)

// Initialise a new Bühlmann model with the gas mix and one of the available
// sets of coefficients, ZHL-16B in this case.
bmann := buhlmann.New(gm, buhlmann.ZHL16B)

// Model the descent from the surface to 30 metres at a rate of eighteen
// metres/min.
bmann.transitionCalc(30.0, 18.0)

// Model staying at the current depth (30 metres) for 25 minutes.
bmann.stopCalc(25.0)

// Model the ascent from the current depth (30 metres) to 18 metres at a rate of
// nine metres/min.
bmann.transitionCalc(18.0, 9.0)

// Get the No Decompression Limit at the current point in the dive.
bmann.GetNDL()

// Get the length of each mandatory decompression stop which will start at three
// times the number of decompression stops metres and go down in multiples of
// three metres to the last stop at three metres at an ascent rate of 6
// metres/min.
bmann.DecompStopLengths(6.0)
```