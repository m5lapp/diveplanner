package helpers

import "testing"

func TestDepth(t *testing.T) {
	tests := []struct {
		name     string
		pressure float64
		want     float64
	}{
		{name: "Perfect vaccuum", pressure: 0.0, want: -10.0},
		{name: "Surface", pressure: 1.0, want: 0.0},
		{name: "Safety stop", pressure: 1.5, want: 5.0},
		{name: "Open water", pressure: 2.8, want: 18.0},
		{name: "Advanced", pressure: 3.75, want: 27.5},
		{name: "Deep", pressure: 10.9, want: 99.0},
		{name: "World record", pressure: 34.235, want: 332.35},
		{name: "Negative", pressure: -2.2, want: 12.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Depth(tt.pressure)

			if p != tt.want {
				t.Errorf("want %f; got %f", tt.want, p)
			}
		})
	}
}

func TestPressure(t *testing.T) {
	tests := []struct {
		name  string
		depth float64
		want  float64
	}{
		{name: "Surface", depth: 0.0, want: 1.0},
		{name: "Safety stop", depth: 5.0, want: 1.5},
		{name: "Open water", depth: 18.0, want: 2.8},
		{name: "Advanced", depth: 27.5, want: 3.75},
		{name: "Deep", depth: 99.0, want: 10.9},
		{name: "World record", depth: 332.35, want: 34.235},
		{name: "Negative", depth: -12.0, want: 2.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Pressure(tt.depth)

			if p != tt.want {
				t.Errorf("want %f; got %f", tt.want, p)
			}
		})
	}
}
