package buhlmann

// A lot of the test values for the algorithmic portions of these tests were
// generated in this Spreadsheet:
// https://docs.google.com/spreadsheets/d/1ZXxxTV2FoBjKvZPALfITcl3Y0LoJ_hwVL6Dud_yBwrY/edit#gid=1156961245

import (
	"testing"

	"github.com/m5lapp/diveplanner/gasmix"
	"github.com/m5lapp/diveplanner/helpers"
)

// Common breathing gas mixtures for use in tests.
var (
	air        *gasmix.GasMix = gasmix.NewAirMix()
	ean32      *gasmix.GasMix
	trimix2135 *gasmix.GasMix
)

type testNewWant struct {
	ccs      string
	c1n2b    float64 // Compartment 1 N2B value.
	c4heht   float64
	c8n2a    float64
	c13n2a   float64
	currTime int
	gmStr    string
}

func TestNew(t *testing.T) {
	ean32, _ = gasmix.NewNitroxMix(0.32)
	trimix2135, _ = gasmix.NewTrimixMix(0.21, 0.35)

	tests := []struct {
		name  string
		model *zhlModel
		want  testNewWant
	}{
		{
			name:  "ZHL16A Air",
			model: New(air, ZHL16A),
			want: testNewWant{
				ccs:    "ZH-L16A",
				c1n2b:  0.5050,
				c4heht: 7.0,
				c8n2a:  0.4701,
				c13n2a: 0.2971,
				gmStr:  "Air",
			},
		}, {
			name:  "ZHL16B EAN32",
			model: New(ean32, ZHL16B),
			want: testNewWant{
				ccs:    "ZH-L16B",
				c1n2b:  0.5240,
				c4heht: 6.99,
				c8n2a:  0.4443,
				c13n2a: 0.2828,
				gmStr:  "Nitrox",
			},
		}, {
			name:  "ZHL16C Trimix21/35",
			model: New(trimix2135, ZHL16C),
			want: testNewWant{
				ccs:    "ZH-L16C",
				c1n2b:  0.5240,
				c4heht: 6.99,
				c8n2a:  0.4500,
				c13n2a: 0.2850,
				gmStr:  "Trimix",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.model.ccs.String() != tt.want.ccs {
				t.Errorf("ccs want: %s; got %s", tt.want.ccs, tt.model.ccs)
			}

			if tt.model.gasMix.MixType().String() != tt.want.gmStr {
				t.Errorf("gasMix want: %s; got %s",
					tt.want.gmStr, tt.model.gasMix.MixType())
			}

			if tt.model.currP != atmPressure {
				t.Errorf("currPressure want: %f; got %f",
					atmPressure, tt.model.currP)
			}

			if tt.model.currT != 0.0 {
				t.Errorf("currTimewant: %f; got %f", 0.0, tt.model.currT)
			}

			for i, c := range tt.model.compartments {
				if c.pHe != 0.0 || c.pN2 != 0.745 {
					t.Errorf("compartment %d invalid; want: %f, %f, %f, got %f, %f",
						i, 0.0, 0.745, 0.745, c.pHe, c.pN2)
				}
			}
		})
	}
}

func TestSchreinerEquation(t *testing.T) {
	tests := []struct {
		name string
		t    float64
		rate float64
		lp   float64
		fig  float64
		po   float64
		ht   float64
		want float64
	}{
		{
			name: "Surface to 30m @20m/min",
			t:    1.5,
			rate: 20.0,
			lp:   1.0,
			fig:  0.68,
			po:   0.74065446,
			ht:   5.0,
			want: 0.919397,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := schreinerEquation(tt.t, tt.rate, tt.lp, tt.fig, tt.po, tt.ht)
			if val != tt.want {
				t.Errorf("want: %f; got: %f", tt.want, val)
			}
		})
	}
}

func TestTransitionStopCalc(t *testing.T) {
	ean32, _ = gasmix.NewNitroxMix(0.32)
	trimix2135, _ = gasmix.NewTrimixMix(0.21, 0.35)

	tests := []struct {
		name  string
		m     *zhlModel
		dRate float64
		aRate float64
		stops [5]float64
		want1 [compartCount]compartModel
		want2 [compartCount]compartModel
		want3 [compartCount]compartModel
		want4 [compartCount]compartModel
		want5 [compartCount]compartModel
	}{
		{
			name:  "ZHL16B EAN32",
			m:     New(ean32, ZHL16B),
			dRate: 20.0,
			aRate: 9.0,
			stops: [5]float64{30.0, 20.0, 5.0, 3.0, 0.0},
			want1: [compartCount]compartModel{
				{pHe: 0.0, pN2: 0.9604734065},
				{pHe: 0.0, pN2: 0.854935828},
				{pHe: 0.0, pN2: 0.8148063852},
				{pHe: 0.0, pN2: 0.7911298294},
				{pHe: 0.0, pN2: 0.7753825758},
				{pHe: 0.0, pN2: 0.7651780404},
				{pHe: 0.0, pN2: 0.7579497155},
				{pHe: 0.0, pN2: 0.7528268538},
				{pHe: 0.0, pN2: 0.7492183981},
				{pHe: 0.0, pN2: 0.7470135324},
				{pHe: 0.0, pN2: 0.7455876214},
				{pHe: 0.0, pN2: 0.744481901},
				{pHe: 0.0, pN2: 0.7436208648},
				{pHe: 0.0, pN2: 0.7429409417},
				{pHe: 0.0, pN2: 0.742411625},
				{pHe: 0.0, pN2: 0.7419991062},
			},
			want2: [compartCount]compartModel{
				{pHe: 0.0, pN2: 2.623737519},
				{pHe: 0.0, pN2: 2.355223562},
				{pHe: 0.0, pN2: 2.062967351},
				{pHe: 0.0, pN2: 1.785803563},
				{pHe: 0.0, pN2: 1.53916835},
				{pHe: 0.0, pN2: 1.34589331},
				{pHe: 0.0, pN2: 1.190434688},
				{pHe: 0.0, pN2: 1.069921384},
				{pHe: 0.0, pN2: 0.9794951773},
				{pHe: 0.0, pN2: 0.9218731511},
				{pHe: 0.0, pN2: 0.8836191826},
				{pHe: 0.0, pN2: 0.853408643},
				{pHe: 0.0, pN2: 0.82954755},
				{pHe: 0.0, pN2: 0.8104951739},
				{pHe: 0.0, pN2: 0.7955332978},
				{pHe: 0.0, pN2: 0.7837935633},
			},
			want3: [compartCount]compartModel{
				{pHe: 0.0, pN2: 2.293538484},
				{pHe: 0.0, pN2: 2.235032027},
				{pHe: 0.0, pN2: 2.026223547},
				{pHe: 0.0, pN2: 1.788457129},
				{pHe: 0.0, pN2: 1.558310923},
				{pHe: 0.0, pN2: 1.369151536},
				{pHe: 0.0, pN2: 1.212449621},
				{pHe: 0.0, pN2: 1.088539823},
				{pHe: 0.0, pN2: 0.9942973446},
				{pHe: 0.0, pN2: 0.9337120068},
				{pHe: 0.0, pN2: 0.8932717494},
				{pHe: 0.0, pN2: 0.8612144689},
				{pHe: 0.0, pN2: 0.8358212842},
				{pHe: 0.0, pN2: 0.8154997178},
				{pHe: 0.0, pN2: 0.7995129364},
				{pHe: 0.0, pN2: 0.7869518044},
			},
			want4: [compartCount]compartModel{
				{pHe: 0.0, pN2: 1.759977057},
				{pHe: 0.0, pN2: 1.947164849},
				{pHe: 0.0, pN2: 1.865485073},
				{pHe: 0.0, pN2: 1.702228415},
				{pHe: 0.0, pN2: 1.515250081},
				{pHe: 0.0, pN2: 1.348448594},
				{pHe: 0.0, pN2: 1.203618141},
				{pHe: 0.0, pN2: 1.085578342},
				{pHe: 0.0, pN2: 0.9939778755},
				{pHe: 0.0, pN2: 0.9343297093},
				{pHe: 0.0, pN2: 0.8942019767},
				{pHe: 0.0, pN2: 0.8622208907},
				{pHe: 0.0, pN2: 0.8367832026},
				{pHe: 0.0, pN2: 0.8163606101},
				{pHe: 0.0, pN2: 0.8002541338},
				{pHe: 0.0, pN2: 0.7875744184},
			},
			want5: [compartCount]compartModel{
				{pHe: 0.0, pN2: 1.672296857},
				{pHe: 0.0, pN2: 1.893536995},
				{pHe: 0.0, pN2: 1.833359353},
				{pHe: 0.0, pN2: 1.68378254},
				{pHe: 0.0, pN2: 1.505220325},
				{pHe: 0.0, pN2: 1.343033018},
				{pHe: 0.0, pN2: 1.200816721},
				{pHe: 0.0, pN2: 1.084189877},
				{pHe: 0.0, pN2: 0.9933195009},
				{pHe: 0.0, pN2: 0.9339951474},
				{pHe: 0.0, pN2: 0.8940232746},
				{pHe: 0.0, pN2: 0.8621325456},
				{pHe: 0.0, pN2: 0.8367460716},
				{pHe: 0.0, pN2: 0.8163517305},
				{pHe: 0.0, pN2: 0.8002596335},
				{pHe: 0.0, pN2: 0.7875864217},
			},
		}, {
			name:  "ZHL16C Trimix2135",
			m:     New(trimix2135, ZHL16C),
			dRate: 12.0,
			aRate: 6.0,
			stops: [5]float64{28.0, 26.0, 5.0, 3.0, 0.0},
			want1: [compartCount]compartModel{
				{pHe: 0.594200479, pN2: 0.8500272426},
				{pHe: 0.357260812, pN2: 0.7969987678},
				{pHe: 0.2454398644, pN2: 0.7770555484},
				{pHe: 0.172881504, pN2: 0.7653529845},
				{pHe: 0.1217401447, pN2: 0.7575972655},
				{pHe: 0.08741980569, pN2: 0.7525835467},
				{pHe: 0.06246123299, pN2: 0.7490379952},
				{pHe: 0.0444573127, pN2: 0.7465281843},
				{pHe: 0.03161528562, pN2: 0.7447618205},
				{pHe: 0.02369476862, pN2: 0.7436831392},
				{pHe: 0.01854668408, pN2: 0.7429857936},
				{pHe: 0.01454138666, pN2: 0.7424451735},
				{pHe: 0.01141210176, pN2: 0.7420242687},
				{pHe: 0.00893575019, pN2: 0.7416919493},
				{pHe: 0.007004678942, pN2: 0.7414332713},
				{pHe: 0.005497387525, pN2: 0.7412316916},
			},
			want2: [compartCount]compartModel{
				{pHe: 1.308064319, pN2: 1.635652611},
				{pHe: 1.305634308, pN2: 1.555354406},
				{pHe: 1.284725518, pN2: 1.439283136},
				{pHe: 1.2218989, pN2: 1.312567586},
				{pHe: 1.105008579, pN2: 1.189482624},
				{pHe: 0.9564539984, pN2: 1.087326816},
				{pHe: 0.7902892357, pN2: 1.001929166},
				{pHe: 0.6276999946, pN2: 0.9339000175},
				{pHe: 0.4838650609, pN2: 0.8818646374},
				{pHe: 0.3815029106, pN2: 0.8482803635},
				{pHe: 0.3087385943, pN2: 0.8258060744},
				{pHe: 0.2484767375, pN2: 0.8079583739},
				{pHe: 0.19905427, pN2: 0.7938008072},
				{pHe: 0.1584286766, pN2: 0.7824581337},
				{pHe: 0.1257898281, pN2: 0.7735270561},
				{pHe: 0.09971522764, pN2: 0.7665048642},
			},
			want3: [compartCount]compartModel{
				{pHe: 0.8818120156, pN2: 1.367334721},
				{pHe: 1.0374402, pN2: 1.429626768},
				{pHe: 1.105340623, pN2: 1.378241814},
				{pHe: 1.113823743, pN2: 1.28765016},
				{pHe: 1.055314405, pN2: 1.183920105},
				{pHe: 0.9458607373, pN2: 1.090359796},
				{pHe: 0.8032302862, pN2: 1.008250697},
				{pHe: 0.6514057834, pN2: 0.9407386974},
				{pHe: 0.5099329176, pN2: 0.887993734},
				{pHe: 0.4059839995, pN2: 0.8534853713},
				{pHe: 0.3306625986, pN2: 0.8301997955},
				{pHe: 0.2674643901, pN2: 0.8116016497},
				{pHe: 0.2151157403, pN2: 0.7967836928},
				{pHe: 0.1717527341, pN2: 0.784871218},
				{pHe: 0.1367060556, pN2: 0.7754664038},
				{pHe: 0.1085784335, pN2: 0.7680564037},
			},
			want4: [compartCount]compartModel{
				{pHe: 0.5986280434, pN2: 1.0694068},
				{pHe: 0.7714839022, pN2: 1.247152651},
				{pHe: 0.8907368422, pN2: 1.263942596},
				{pHe: 0.9566652184, pN2: 1.217990193},
				{pHe: 0.953554269, pN2: 1.143040713},
				{pHe: 0.8866273265, pN2: 1.066160046},
				{pHe: 0.7743165857, pN2: 0.9941305625},
				{pHe: 0.6411791186, pN2: 0.932524005},
				{pHe: 0.5095950801, pN2: 0.8831644351},
				{pHe: 0.4095739007, pN2: 0.8503592444},
				{pHe: 0.3356602977, pN2: 0.8280127717},
				{pHe: 0.2728255865, pN2: 0.8100495088},
				{pHe: 0.2202628848, pN2: 0.7956669639},
				{pHe: 0.1763933293, pN2: 0.7840605774},
				{pHe: 0.1407309063, pN2: 0.7748703866},
				{pHe: 0.1119812456, pN2: 0.7676129915},
			},
			want5: [compartCount]compartModel{
				{pHe: 0.5386705562, pN2: 0.9955025044},
				{pHe: 0.7090384785, pN2: 1.19657782},
				{pHe: 0.8357987022, pN2: 1.230418803},
				{pHe: 0.913653473, pN2: 1.196590753},
				{pHe: 0.9239177414, pN2: 1.129896423},
				{pHe: 0.8681834207, pN2: 1.058017147},
				{pHe: 0.7643523167, pN2: 0.9891373186},
				{pHe: 0.6367408295, pN2: 0.9294581259},
				{pHe: 0.5082831889, pN2: 0.8812573392},
				{pHe: 0.4096347296, pN2: 0.8490641266},
				{pHe: 0.3363096181, pN2: 0.8270701402},
				{pHe: 0.2737347642, pN2: 0.8093551124},
				{pHe: 0.2212385896, pN2: 0.7951499187},
				{pHe: 0.177328413, pN2: 0.7836733266},
				{pHe: 0.1415728293, pN2: 0.7745777295},
				{pHe: 0.1127108548, pN2: 0.7673900486},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.transitionCalc(tt.stops[0], tt.dRate)
			for i, c := range tt.m.compartments {
				if !helpers.EqualFloat64(c.pHe, tt.want1[i].pHe) {
					t.Errorf("s1c%dpHe: want: %f; got: %f", i+1, tt.want1[i].pHe, c.pHe)
				}
				if !helpers.EqualFloat64(c.pN2, tt.want1[i].pN2) {
					t.Errorf("s1c%dpN2: want: %f; got: %f", i+1, tt.want1[i].pN2, c.pN2)
				}
			}

			tt.m.stopCalc(tt.stops[1])
			for i, c := range tt.m.compartments {
				if !helpers.EqualFloat64(c.pHe, tt.want2[i].pHe) {
					t.Errorf("s2c%dpHe: want: %f; got: %f", i+1, tt.want2[i].pHe, c.pHe)
				}
				if !helpers.EqualFloat64(c.pN2, tt.want2[i].pN2) {
					t.Errorf("s2c%dpN2: want: %f; got: %f", i+1, tt.want2[i].pN2, c.pN2)
				}
			}

			tt.m.transitionCalc(tt.stops[2], tt.aRate)
			for i, c := range tt.m.compartments {
				if !helpers.EqualFloat64(c.pHe, tt.want3[i].pHe) {
					t.Errorf("s3c%dpHe: want: %f; got: %f", i+1, tt.want3[i].pHe, c.pHe)
				}
				if !helpers.EqualFloat64(c.pN2, tt.want3[i].pN2) {
					t.Errorf("s3c%dpN2: want: %f; got: %f", i+1, tt.want3[i].pN2, c.pN2)
				}
			}

			tt.m.stopCalc(tt.stops[3])
			for i, c := range tt.m.compartments {
				if !helpers.EqualFloat64(c.pHe, tt.want4[i].pHe) {
					t.Errorf("s4c%dpHe: want: %f; got: %f", i+1, tt.want4[i].pHe, c.pHe)
				}
				if !helpers.EqualFloat64(c.pN2, tt.want4[i].pN2) {
					t.Errorf("s4c%dpN2: want: %f; got: %f", i+1, tt.want4[i].pN2, c.pN2)
				}
			}

			tt.m.transitionCalc(tt.stops[4], tt.aRate)
			for i, c := range tt.m.compartments {
				if !helpers.EqualFloat64(c.pHe, tt.want5[i].pHe) {
					t.Errorf("s5c%dpHe: want: %f; got: %f", i+1, tt.want5[i].pHe, c.pHe)
				}
				if !helpers.EqualFloat64(c.pN2, tt.want5[i].pN2) {
					t.Errorf("s5c%dpN2: want: %f; got: %f", i+1, tt.want5[i].pN2, c.pN2)
				}
			}
		})
	}
}

func TestAscentCeilingNDL(t *testing.T) {
	ean32, _ = gasmix.NewNitroxMix(0.32)
	trimix2135, _ = gasmix.NewTrimixMix(0.21, 0.35)

	tests := []struct {
		name    string
		m       *zhlModel
		dRate   float64
		stops   [2]float64
		wantAc  float64
		wantNdl int
	}{
		{
			name:    "EAN32: 20min @ 30m",
			m:       New(ean32, ZHL16B),
			dRate:   20,
			stops:   [2]float64{30.0, 20.0},
			wantAc:  -1.172073717,
			wantNdl: 6,
		},
		{
			name:    "EAN32: 30min @ 30m",
			m:       New(ean32, ZHL16B),
			dRate:   20,
			stops:   [2]float64{30.0, 30.0},
			wantAc:  0.5636003878,
			wantNdl: 0,
		},
		{
			name:    "EAN32: 1min @ 10m",
			m:       New(ean32, ZHL16B),
			dRate:   20,
			stops:   [2]float64{10.0, 1.0},
			wantAc:  -5.090898233,
			wantNdl: 60,
		},
		{
			name:    "EAN32: 25min @ 24m",
			m:       New(ean32, ZHL16B),
			dRate:   20,
			stops:   [2]float64{24.0, 25.0},
			wantAc:  -2.510879382,
			wantNdl: 24,
		},
		{
			name:    "Trimix2135: 10min @ 26m",
			m:       New(trimix2135, ZHL16C),
			dRate:   9,
			stops:   [2]float64{26.0, 10.0},
			wantAc:  -0.8575469199,
			wantNdl: 2,
		},
		{
			name:    "Trimix2135: 20min @ 18m",
			m:       New(trimix2135, ZHL16C),
			dRate:   9,
			stops:   [2]float64{18.0, 20.0},
			wantAc:  -1.597315895,
			wantNdl: 14,
		},
		{
			name:    "Trimix2135: 45min @ 12m",
			m:       New(trimix2135, ZHL16C),
			dRate:   9,
			stops:   [2]float64{12.0, 45.0},
			wantAc:  -1.933904326,
			wantNdl: 60,
		},
		{
			name:    "Trimix2135: 27min @ 24m",
			m:       New(trimix2135, ZHL16C),
			dRate:   9,
			stops:   [2]float64{24.0, 27.0},
			wantAc:  2.166049527,
			wantNdl: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.transitionCalc(tt.stops[0], tt.dRate)
			tt.m.stopCalc(tt.stops[1])

			ac := tt.m.ascentCeiling()
			if !helpers.EqualFloat64(ac, tt.wantAc) {
				t.Errorf("Ascent ceil want: %f; got: %f", tt.wantAc, ac)
			}

			ndl := tt.m.getNDL()
			if ndl != tt.wantNdl {
				t.Errorf("NDL want: %d; got: %d", tt.wantNdl, ndl)
			}
		})
	}
}

func equalIntSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestDecompStopLengths(t *testing.T) {
	ean32, _ = gasmix.NewNitroxMix(0.32)
	trimix2135, _ = gasmix.NewTrimixMix(0.21, 0.35)

	tests := []struct {
		name  string
		m     *zhlModel
		dRate float64
		aRate float64
		stops [2]float64
		want  []int
	}{
		{
			name:  "EAN32: 20min @ 30m",
			m:     New(ean32, ZHL16B),
			dRate: 20.0,
			aRate: 9.0,
			stops: [2]float64{30.0, 20.0},
			want:  []int{}, // No decompression obligations for this dive.
		},
		{
			name:  "EAN32: 60min @ 30m",
			m:     New(ean32, ZHL16B),
			dRate: 20.0,
			aRate: 9.0,
			stops: [2]float64{30.0, 60.0},
			want:  []int{1, 15},
		},
		{
			name:  "Trimix2135: 22min @ 45m",
			m:     New(trimix2135, ZHL16B),
			dRate: 20.0,
			aRate: 9.0,
			stops: [2]float64{45.0, 22.0},
			want:  []int{1, 4, 10, 22},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.m.transitionCalc(tt.stops[0], tt.dRate)
			tt.m.stopCalc(tt.stops[1])
			modelBkup := tt.m.copyModel()

			dsl := tt.m.decompStopLengths(tt.aRate)
			if !equalIntSlice(dsl, tt.want) {
				t.Errorf("want: %v; got: %v", tt.want, dsl)
			}

			// Check that the main model has not been modified.
			for i, c1 := range tt.m.compartments {
				c2 := modelBkup.compartments[i]
				if c1.pHe != c2.pHe || c1.pN2 != c2.pN2 {
					t.Errorf("Model compartment %d changed: want: %v; got %v",
						i+1, c2, c1)
				}
			}
		})
	}
}
