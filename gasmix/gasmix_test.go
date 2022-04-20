package gasmix

import "testing"

// TODO: TestNewMix()

func TestMixType(t *testing.T) {
	tests := []struct {
		name string
		fhe  float64
		fn2  float64
		fo2  float64
		want MixType
		str  string
	}{
		{name: "Air", fhe: 0.0, fn2: 0.79, fo2: 0.21, want: Air, str: "Air"},
		{name: "Nitrox32", fhe: 0.0, fn2: 0.68, fo2: 0.32, want: Nitrox, str: "Nitrox"},
		{name: "Nitrox50", fhe: 0.0, fn2: 0.5, fo2: 0.5, want: Nitrox, str: "Nitrox"},
		{name: "Nitrox100", fhe: 0.0, fn2: 0.0, fo2: 1.0, want: Nitrox, str: "Nitrox"},
		{name: "Trimix3040", fhe: 0.4, fn2: 0.3, fo2: 0.3, want: Trimix, str: "Trimix"},
		{name: "Trimix2150", fhe: 0.5, fn2: 0.29, fo2: 0.21, want: Trimix, str: "Trimix"},
		{name: "Trimix5030", fhe: 0.5, fn2: 0.3, fo2: 0.5, want: Trimix, str: "Trimix"},
		{name: "Heliox2179", fhe: 0.79, fn2: 0.0, fo2: 0.21, want: Heliox, str: "Heliox"},
		{name: "Heliox3070", fhe: 70.0, fn2: 0.0, fo2: 0.30, want: Heliox, str: "Heliox"},
		{name: "Heliox5050", fhe: 50.0, fn2: 0.0, fo2: 0.50, want: Heliox, str: "Heliox"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := GasMix{FHe: tt.fhe, FN2: tt.fn2, FO2: tt.fo2}
			mt := gm.MixType()

			if mt != tt.want {
				t.Errorf("want %v; got %v", tt.want, mt)
			}

			if mt.String() != tt.str {
				t.Errorf("want string %s; got %s", tt.str, mt.String())
			}
		})
	}
}

func TestEAD(t *testing.T) {
	tests := []struct {
		name string
		fo2  float64
		ppo2 float64
		want float64
	}{
		{name: "21% @ 1.4", fo2: 0.21, ppo2: 1.4, want: 57.0},
		{name: "21% @ 1.6", fo2: 0.21, ppo2: 1.6, want: 66.0},
		{name: "30% @ 1.4", fo2: 0.30, ppo2: 1.4, want: 37.0},
		{name: "30% @ 1.6", fo2: 0.30, ppo2: 1.6, want: 43.0},
		{name: "32% @ 1.4", fo2: 0.32, ppo2: 1.4, want: 34.0},
		{name: "32% @ 1.6", fo2: 0.32, ppo2: 1.6, want: 40.0},
		{name: "40% @ 1.4", fo2: 0.40, ppo2: 1.4, want: 25.0},
		{name: "40% @ 1.6", fo2: 0.40, ppo2: 1.6, want: 30.0},
		{name: "100% @ 1.4", fo2: 1.00, ppo2: 1.4, want: 4.0},
		{name: "100% @ 1.6", fo2: 1.00, ppo2: 1.6, want: 6.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm, err := NewNitroxMix(tt.fo2)
			mod := gm.MOD(tt.ppo2)

			if err != nil {
				t.Errorf("want %f; got error %v", tt.want, err)
			}

			if gm.MOD(tt.ppo2) != tt.want {
				t.Errorf("want %f; got %f", tt.want, mod)
			}
		})
	}
}

func TestMOD(t *testing.T) {
	tests := []struct {
		name string
		fo2  float64
		ppo2 float64
		want float64
	}{
		{name: "21% @ 1.2", fo2: 0.21, ppo2: 1.2, want: 47.0},
		{name: "21% @ 1.6", fo2: 0.21, ppo2: 1.6, want: 66.0},
		{name: "30% @ 1.4", fo2: 0.30, ppo2: 1.4, want: 37.0},
		{name: "30% @ 1.6", fo2: 0.30, ppo2: 1.6, want: 43.0},
		{name: "32% @ 1.4", fo2: 0.32, ppo2: 1.4, want: 34.0},
		{name: "32% @ 1.6", fo2: 0.32, ppo2: 1.6, want: 40.0},
		{name: "40% @ 1.3", fo2: 0.40, ppo2: 1.3, want: 23.0},
		{name: "40% @ 1.4", fo2: 0.40, ppo2: 1.4, want: 25.0},
		{name: "40% @ 1.6", fo2: 0.40, ppo2: 1.6, want: 30.0},
		{name: "100% @ 1.4", fo2: 1.00, ppo2: 1.4, want: 4.0},
		{name: "100% @ 1.6", fo2: 1.00, ppo2: 1.6, want: 6.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm, err := NewNitroxMix(tt.fo2)
			mod := gm.MOD(tt.ppo2)

			if err != nil {
				t.Errorf("want %f; got error %v", tt.want, err)
			}

			if gm.MOD(tt.ppo2) != tt.want {
				t.Errorf("want %f; got %f", tt.want, mod)
			}
		})
	}
}
