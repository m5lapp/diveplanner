package buhlmann

// Sources of information used for the Bühlmann ZHL-16 algorithm:
//   http://www.lizardland.co.uk/DIYDeco.html
//   https://github.com/eianlei/pydplan/blob/master/pydplan_buhlmann.py
//   https://github.com/AquaBSD/libbuhlmann/tree/master/src
//   https://scholars.unh.edu/cgi/viewcontent.cgi?article=1511&context=thesis
//   http://www.diveresearch.org/download/Publicaties/Haldane%20en%20bellen%202006.pdf
//   https://wrobell.dcmod.org/decotengu/model.html
//   https://www.medmastery.com/guide/blood-gas-analysis-clinical-guide/partial-pressure-and-alveolar-air-equation-made-simple

import (
	"math"

	"github.com/m5lapp/diveplanner/gasmix"
	"github.com/m5lapp/diveplanner/helpers"
)

const (
	// Atmospheric pressure in bar at sea-level.
	// TODO: Make this a function that accounts for altitude.
	// ISSUE: 1.01325 bar is a more accurate figure.
	atmPressure = 1.0
	// Number of compartments in each ZH-L model.
	compartCount = 16
	// Partial pressure of water vapour in the lungs in bar. This is constant
	// regardless of pressure. Value is equivalent to 47 mmHg.
	pH2O = 0.06266
)

type compartCoefs struct {
	n    int
	n2Ht float64
	n2A  float64
	n2B  float64
	heHt float64
	heA  float64
	heB  float64
}

// Custom type to represent a set of compartment coefficients.
type compartCoefSet int

const (
	ZHL16A compartCoefSet = iota
	ZHL16B
	ZHL16C
)

func (ccs compartCoefSet) String() string {
	return [...]string{"ZH-L16A", "ZH-L16B", "ZH-L16C"}[ccs]
}

var compartCoefSets = [][compartCount]compartCoefs{
	{
		{n: 1, n2Ht: 4.0, n2A: 1.2599, n2B: 0.5050, heHt: 1.5, heA: 1.7435, heB: 0.1911},
		{n: 2, n2Ht: 8.0, n2A: 1.0000, n2B: 0.6514, heHt: 3.0, heA: 1.3838, heB: 0.4295},
		{n: 3, n2Ht: 12.5, n2A: 0.8618, n2B: 0.7222, heHt: 4.7, heA: 1.1925, heB: 0.5446},
		{n: 4, n2Ht: 18.5, n2A: 0.7562, n2B: 0.7725, heHt: 7.0, heA: 1.0465, heB: 0.6265},
		{n: 5, n2Ht: 27.0, n2A: 0.6667, n2B: 0.8125, heHt: 10.2, heA: 0.9226, heB: 0.6917},
		{n: 6, n2Ht: 38.3, n2A: 0.5933, n2B: 0.8434, heHt: 14.5, heA: 0.8211, heB: 0.7420},
		{n: 7, n2Ht: 54.3, n2A: 0.5282, n2B: 0.8693, heHt: 20.5, heA: 0.7309, heB: 0.7841},
		{n: 8, n2Ht: 77.0, n2A: 0.4701, n2B: 0.8910, heHt: 29.1, heA: 0.6506, heB: 0.8195},
		{n: 9, n2Ht: 109.0, n2A: 0.4187, n2B: 0.9092, heHt: 41.1, heA: 0.5794, heB: 0.8491},
		{n: 10, n2Ht: 146.0, n2A: 0.3798, n2B: 0.9222, heHt: 55.1, heA: 0.5256, heB: 0.8703},
		{n: 11, n2Ht: 187.0, n2A: 0.3497, n2B: 0.9319, heHt: 70.6, heA: 0.4840, heB: 0.8860},
		{n: 12, n2Ht: 239.0, n2A: 0.3223, n2B: 0.9403, heHt: 90.2, heA: 0.4460, heB: 0.8997},
		{n: 13, n2Ht: 305.0, n2A: 0.2971, n2B: 0.9477, heHt: 115.1, heA: 0.4112, heB: 0.9118},
		{n: 14, n2Ht: 390.0, n2A: 0.2737, n2B: 0.9544, heHt: 147.2, heA: 0.3788, heB: 0.9226},
		{n: 15, n2Ht: 498.0, n2A: 0.2523, n2B: 0.9602, heHt: 187.9, heA: 0.3492, heB: 0.9321},
		{n: 16, n2Ht: 635.0, n2A: 0.2327, n2B: 0.9653, heHt: 239.6, heA: 0.3220, heB: 0.9404},
	}, {
		{n: 1, n2Ht: 4.0, n2A: 1.2599, n2B: 0.5240, heHt: 1.51, heA: 1.6189, heB: 0.4245},
		{n: 2, n2Ht: 8.0, n2A: 1.0000, n2B: 0.6514, heHt: 3.02, heA: 1.3830, heB: 0.5747},
		{n: 3, n2Ht: 12.5, n2A: 0.8618, n2B: 0.7222, heHt: 4.72, heA: 1.1919, heB: 0.6527},
		{n: 4, n2Ht: 18.5, n2A: 0.7562, n2B: 0.7825, heHt: 6.99, heA: 1.0458, heB: 0.7223},
		{n: 5, n2Ht: 27.0, n2A: 0.6667, n2B: 0.8126, heHt: 10.21, heA: 0.9220, heB: 0.7582},
		{n: 6, n2Ht: 38.3, n2A: 0.5505, n2B: 0.8434, heHt: 14.48, heA: 0.8205, heB: 0.7957},
		{n: 7, n2Ht: 54.3, n2A: 0.4858, n2B: 0.8693, heHt: 20.53, heA: 0.7305, heB: 0.8279},
		{n: 8, n2Ht: 77.0, n2A: 0.4443, n2B: 0.8910, heHt: 29.11, heA: 0.6502, heB: 0.8553},
		{n: 9, n2Ht: 109.0, n2A: 0.4187, n2B: 0.9092, heHt: 41.20, heA: 0.5950, heB: 0.8757},
		{n: 10, n2Ht: 146.0, n2A: 0.3798, n2B: 0.9222, heHt: 55.19, heA: 0.5545, heB: 0.8903},
		{n: 11, n2Ht: 187.0, n2A: 0.3497, n2B: 0.9319, heHt: 70.69, heA: 0.5333, heB: 0.8997},
		{n: 12, n2Ht: 239.0, n2A: 0.3223, n2B: 0.9403, heHt: 90.34, heA: 0.5189, heB: 0.9073},
		{n: 13, n2Ht: 305.0, n2A: 0.2828, n2B: 0.9477, heHt: 115.29, heA: 0.5181, heB: 0.9122},
		{n: 14, n2Ht: 390.0, n2A: 0.2737, n2B: 0.9544, heHt: 147.42, heA: 0.5176, heB: 0.9171},
		{n: 15, n2Ht: 498.0, n2A: 0.2523, n2B: 0.9602, heHt: 188.24, heA: 0.5172, heB: 0.9217},
		{n: 16, n2Ht: 635.0, n2A: 0.2327, n2B: 0.9653, heHt: 240.03, heA: 0.5119, heB: 0.9267},
	}, {
		{n: 1, n2Ht: 4.0, n2A: 1.2599, n2B: 0.5240, heHt: 1.51, heA: 1.6189, heB: 0.4245},
		{n: 2, n2Ht: 8.0, n2A: 1.0000, n2B: 0.6514, heHt: 3.02, heA: 1.3830, heB: 0.5747},
		{n: 3, n2Ht: 12.5, n2A: 0.8618, n2B: 0.7222, heHt: 4.72, heA: 1.1919, heB: 0.6527},
		{n: 4, n2Ht: 18.5, n2A: 0.7562, n2B: 0.7825, heHt: 6.99, heA: 1.0458, heB: 0.7223},
		{n: 5, n2Ht: 27.0, n2A: 0.6667, n2B: 0.8126, heHt: 10.21, heA: 0.9220, heB: 0.7582},
		{n: 6, n2Ht: 38.3, n2A: 0.5600, n2B: 0.8434, heHt: 14.48, heA: 0.8205, heB: 0.7957},
		{n: 7, n2Ht: 54.3, n2A: 0.4947, n2B: 0.8693, heHt: 20.53, heA: 0.7305, heB: 0.8279},
		{n: 8, n2Ht: 77.0, n2A: 0.4500, n2B: 0.8910, heHt: 29.11, heA: 0.6502, heB: 0.8553},
		{n: 9, n2Ht: 109.0, n2A: 0.4187, n2B: 0.9092, heHt: 41.20, heA: 0.5950, heB: 0.8757},
		{n: 10, n2Ht: 146.0, n2A: 0.3798, n2B: 0.9222, heHt: 55.19, heA: 0.5545, heB: 0.8903},
		{n: 11, n2Ht: 187.0, n2A: 0.3497, n2B: 0.9319, heHt: 70.69, heA: 0.5333, heB: 0.8997},
		{n: 12, n2Ht: 239.0, n2A: 0.3223, n2B: 0.9403, heHt: 90.34, heA: 0.5189, heB: 0.9073},
		{n: 13, n2Ht: 305.0, n2A: 0.2850, n2B: 0.9477, heHt: 115.29, heA: 0.5181, heB: 0.9122},
		{n: 14, n2Ht: 390.0, n2A: 0.2737, n2B: 0.9544, heHt: 147.42, heA: 0.5176, heB: 0.9171},
		{n: 15, n2Ht: 498.0, n2A: 0.2523, n2B: 0.9602, heHt: 188.24, heA: 0.5172, heB: 0.9217},
		{n: 16, n2Ht: 635.0, n2A: 0.2327, n2B: 0.9653, heHt: 240.03, heA: 0.5119, heB: 0.9267},
	},
}

// Represents the pressure of Helium and Nitrogen in a tissue compartment.
type compartModel struct {
	pHe float64 // Pressure of Helium.
	pN2 float64 // Pressure of Nitrogen.
}

type ZhlModel struct {
	ccs          compartCoefSet
	coefs        *[compartCount]compartCoefs
	compartments *[compartCount]compartModel
	currP        float64
	currT        float64
	gasMix       *gasmix.GasMix
}

// Constructor that creates, initialises and returns a new Bühlmann ZHL-16
// model. The initial value of pN takes into account the Partial Pressure of
// water vapour in the lungs which offsets some of the volume of Nitrogen in the
// air.
func New(gm *gasmix.GasMix, ccs compartCoefSet) *ZhlModel {
	// Create the compartment model and initialise the values for each one.
	var c [compartCount]compartModel
	for i := 0; i < compartCount; i++ {
		c[i] = compartModel{
			pHe: 0.0,
			pN2: 0.79 * (1.0 - pH2O),
		}
	}

	return &ZhlModel{
		ccs:          ccs,
		coefs:        &compartCoefSets[ccs],
		compartments: &c,
		currP:        atmPressure,
		currT:        0.0,
		gasMix:       gm,
	}
}

// copyModel() returns a deep copy of the Bühlmann model that can be used for
// extrapolation calculations from the current state without modifying the main
// model instance.
func (m *ZhlModel) copyModel() *ZhlModel {
	// Create a deep copy of the existing model's compartments.
	var compartCopy [compartCount]compartModel
	for i := 0; i < compartCount; i++ {
		compartCopy[i] = compartModel{
			pHe: m.compartments[i].pHe,
			pN2: m.compartments[i].pN2,
		}
	}

	return &ZhlModel{
		ccs:          m.ccs,
		coefs:        m.coefs,
		compartments: &compartCopy,
		currP:        m.currP,
		currT:        m.currT,
		gasMix:       m.gasMix,
	}
}

// pulmonaryPPHe() calculates the partial pressure of Helium in the lungs
// (alveoli) where the water vapour content reduces the PPHe from what it would
// otherwise be under the given pressure.
func (m *ZhlModel) pulmonaryPPHe(ambPressure float64) float64 {
	return (ambPressure - pH2O) * m.gasMix.PPHe(ambPressure)
}

// pulmonaryPPN2() calculates the partial pressure of Nitrogen in the lungs
// (alveoli) where the water vapour content reduces the PPN2 from what it would
// otherwise be under the given pressure.
func (m *ZhlModel) pulmonaryPPN2(ambPressure float64) float64 {
	return (ambPressure - pH2O) * m.gasMix.PPHe(ambPressure)
}

// The Schreiner Equation calculates the gas loading for a descent or ascent.
// pamb is the ambient pressure at the start of the calculation.
// t is the time that the transition will take in minutes.
// prate is the pressure change in bar per minute.
// fig is the fraction of inert gas (Nitrogen or Helium).
// pi is the initial pressure of the inert gas in the compartment.
// ht is the inert gas half-time for the curent compartment.
func schreinerEquation(pamb, t, prate, fig, pi, ht float64) float64 {
	// palv is the partial pressure of the inert gas being inspired inside the
	// lungs.
	palv := (pamb - pH2O) * fig
	// k is the inert gas' half-time constant.
	k := math.Log(2.0) / ht
	// r is the rate of change in the inspired inert gas' pressure in bar/min.
	r := prate * fig

	return palv + r*(t-(1.0/k)) - (palv-pi-(r/k))*math.Pow(math.E, (-k*t))
}

// TransitionCalc() recalculates the model's compartment inert gas pressures
// following a descent or ascent to the given depth at the given rate in m/min.
func (m *ZhlModel) TransitionCalc(depth, rate float64) {
	// Ambient pressure at the end of the transition.
	nextP := helpers.Pressure(depth)
	// Pressure change in bar per minute at the given rate of metres per minute.
	pRate := rate / 10.0
	if nextP < m.currP && rate >= 0.0 {
		// We are ascending, so pressure change rate should be negative.
		pRate *= -1.0
	}
	// Time taken to do the transition at the specified rate.
	time := (nextP - m.currP) / pRate

	// Calculate the new compartment pressures for He and N2 for each
	// compartment.
	// TODO: Can these be parallelised?
	for i, c := range m.compartments {
		m.compartments[i].pHe = schreinerEquation(m.currP, time, pRate, m.gasMix.FHe, c.pHe, m.coefs[i].heHt)
		m.compartments[i].pN2 = schreinerEquation(m.currP, time, pRate, m.gasMix.FN2, c.pN2, m.coefs[i].n2Ht)
	}

	// Update the time and ambient pressure at the end of the transition.
	m.currP = nextP
	m.currT += math.Abs(time)
}

// Like transitionCalc(), StopCalc() also recalculates the model's compartment
// inert gas pressures but when staying at the current depth for a given time in
// minutes.
func (m *ZhlModel) StopCalc(time float64) {
	// Calculate the new compartment pressures for He and N2 for each
	// compartment. Note that prate is set to zero as we are staying at one
	// level.
	for i, c := range m.compartments {
		m.compartments[i].pHe = schreinerEquation(m.currP, time, 0.0, m.gasMix.FHe, c.pHe, m.coefs[i].heHt)
		m.compartments[i].pN2 = schreinerEquation(m.currP, time, 0.0, m.gasMix.FN2, c.pN2, m.coefs[i].n2Ht)
	}

	// Update the time at the end of the transition. The ambient pressure
	// remains the same and does not need to be updated.
	m.currT += math.Abs(time)
}

// ascentCeiling() calculates the minimum (shallowest) depth in metres to which
// the diver can ascend safely based on their current compartment loading. If
// the ascent ceiling is greater than zero metres, then the dive is a
// decompression dive. The return value is an absolute pressure in bar.
func (m *ZhlModel) ascentCeiling() float64 {
	ascentCeil := -(math.MaxFloat64)

	for i, c := range m.compartments {
		var a, b float64
		if m.gasMix.MixType() == gasmix.Heliox {
			a, b = m.coefs[i].heA, m.coefs[i].heB
		} else {
			// For any Nitrogen-based mixes, use the Nitrogen a and b values.
			// For Trimix, this is more conservative than interpolating the a
			// and b values based on the pressure of each inert gas in the
			// compartment.
			a, b = m.coefs[i].n2A, m.coefs[i].n2B
		}

		ceil := ((c.pHe + c.pN2) - a) * b
		ascentCeil = math.Max(ascentCeil, ceil)
	}
	return helpers.Depth(ascentCeil)
}

// firstDecompStop() returns the depth in meters rounded up to the
// nearest multiple of three where the first decompression stop should take
// place. A zero or negative value means that the diver is within
// no-decompression limits and can ascend to the surface directly.
func (m *ZhlModel) firstDecompStop() float64 {
	return math.Ceil(m.ascentCeiling()/3.0) * 3.0
}

// Get the No Decompression Limits (NDLs) by copying the model, then simulating
// staying at the current pressure in one minute intervals until a positive
// ascent ceiling is found. The number of iterations is then the NDL value. Up
// to 60 iterations will be performed, if 60 is returned then it is assumed to
// be read as 60+ minutes.
func (m *ZhlModel) GetNDL() int {
	maxNDL := 60

	if m.currT == 0.0 {
		return maxNDL
	}

	// Make a copy of the model's compartments data structure so that we do not
	// overwrite it with data from the NDL calculations.
	ndlModel := m.copyModel()
	for i := 0; i <= maxNDL; i++ {
		ndlModel.StopCalc(1.0)
		ac := ndlModel.ascentCeiling()
		if ac > 0.0 {
			return i
		}
	}

	return maxNDL
}

// decompStopLengths() calculates the length of each decompression stop for the
// model if the dive stopped wherever the model is currently up to. It first
// calculates the depth of the first stop, then calculates the number of minutes
// that the diver must stay there until their ascent ceiling is less than or
// equal to the depth that is 3 metres shallower than that one. This process is
// repeated up to and including the last stop at 3 metres. If there are no
// decompression stops required, then an empty slice is returned.
func (m *ZhlModel) decompStopLengths(aRate float64) []int {
	var stops []int

	firstStop := m.firstDecompStop()
	lastStop := 3.0
	model := m.copyModel()

	// If the firstStop value calculated is shallower than the lastStop constant
	// value then the whole loop is skipped as there are no decompression
	// requirements and an empty slice will be returned.
	for currStop := firstStop; currStop >= lastStop; currStop -= 3.0 {
		model.TransitionCalc(currStop, aRate)
		// TODO: Allow different deco gases to be used.
		nextStop := currStop - 3.0
		ac := model.ascentCeiling()

		// Check for the case where during the ascent to the current
		// decompression stop depth, the diver has off-gased sufficiently such
		// that the stop is no longer required, that is, their new ascent
		// ceiling is shallower than the depth of the stop they are about to
		// start. For instance, if their ascent ceiling at depth is 3.005 and
		// after their ascent to the first stop at 6m, their new ascent ceiling
		// is 2.951, then the 6m stop can be skipped.
		if ac < nextStop {
			continue
		}

		stopLength := 0
		for ac >= nextStop {
			model.StopCalc(1.0)
			ac = model.ascentCeiling()
			stopLength += 1
		}

		stops = append(stops, stopLength)
	}

	return stops
}
