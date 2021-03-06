package diveplanner

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		dp   *DivePlan
		want []error
	}{
		{
			name: "invalid dive",
			dp: &DivePlan{
				DescentRate:     50.0,
				AscentRate:      -8.0,
				SACRate:         0.5,
				TankCount:       7,
				TankCapacity:    25.0,
				WorkingPressure: 350.0,
				DiveFactor:      0.7,
				MaxPPO2:         2.0,
				Stops: []*DivePlanStop{
					{22.0, 26, false, ""},
					{5.0, 3, false, ""},
				},
			},
			want: []error{
				errors.New("name cannot be empty"),
				errors.New("Descent Rate value (50) must be between 1 and 30 inclusive"),
				errors.New("Ascent Rate value (-8) must be between 1 and 18 inclusive"),
				errors.New("SAC Rate value (0.5) must be between 1 and 100 inclusive"),
				errors.New("Tank Count value (7) must be between 1 and 6 inclusive"),
				errors.New("Tank Capacity value (25) must be between 3 and 20 inclusive"),
				errors.New("Tank Working Pressure value (350) must be between 150 and 300 inclusive"),
				errors.New("Dive Factor value (0.7) must be between 1 and 6 inclusive"),
				errors.New("Max PPO2 value (2) must be between 0.21 and 1.6 inclusive"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.dp.Validate()
			if len(errs) != len(tt.want) {
				t.Errorf("validate incorrect num of errors, want: %d; got: %d - %v", len(tt.want), len(errs), errs)
			}

			for i, e := range errs {
				w := tt.want[i].Error()
				if e.Error() != w {
					t.Errorf("errors do not match, want:\n%v;\ngot:\n%v", e.Error(), w)
					break
				}
			}
		})
	}
}

func TestTransitionDuration(t *testing.T) {
	tests := []struct {
		name  string
		dp    *DivePlan
		fromD float64
		toD   float64
		want  float64
	}{
		{
			name:  "Desc 0m to 27m",
			dp:    &DivePlan{DescentRate: 20},
			fromD: 0.0,
			toD:   27.0,
			want:  2.0,
		},
		{
			name:  "Desc 12m to 18m",
			dp:    &DivePlan{DescentRate: 9},
			fromD: 12.0,
			toD:   18.0,
			want:  1.0,
		},
		{
			name:  "Asc 16m to 5m",
			dp:    &DivePlan{AscentRate: 10},
			fromD: 16.0,
			toD:   5.0,
			want:  2.0,
		},
		{
			name:  "Tiny delta",
			dp:    &DivePlan{DescentRate: 10},
			fromD: 10.00000000000000000000,
			toD:   10.00000000000000100001,
			want:  0.0,
		},
		{
			name:  "No depth delta",
			dp:    &DivePlan{DescentRate: 10},
			fromD: 5.0,
			toD:   5.0,
			want:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			time := tt.dp.transitionDuration(tt.fromD, tt.toD)
			if time != tt.want {
				t.Errorf("want: %f; got: %f", tt.want, time)
			}
		})
	}
}

func TestDiveProfile(t *testing.T) {
	tests := []struct {
		name string
		dp   *DivePlan
		want []*DivePlanStop
	}{
		{
			name: "No stops",
			dp: &DivePlan{
				DescentRate: 20,
				AscentRate:  9,
				Stops:       []*DivePlanStop{},
			},
			want: []*DivePlanStop{},
		},
		{
			name: "22m bounce dive",
			dp: &DivePlan{
				DescentRate: 20,
				AscentRate:  9,
				Stops: []*DivePlanStop{
					{22.0, 26, false, ""},
					{5.0, 3, false, ""},
				},
			},
			want: []*DivePlanStop{
				{11.0, 2, true, "Descent from 0.0m to 22.0m"},
				{22.0, 26, false, ""},
				{13.5, 2, true, "Ascent from 22.0m to 5.0m"},
				{5.0, 3, false, ""},
				{2.5, 1, true, "Ascent from 5.0m to 0.0m"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, s := range tt.dp.DiveProfile() {
				if s.Depth != tt.want[i].Depth ||
					s.Duration != tt.want[i].Duration ||
					s.IsTransition != tt.want[i].IsTransition ||
					s.Comment != tt.want[i].Comment {
					t.Errorf("want: %v; got: %v", tt.want, s)
				}
			}
		})
	}
}

func TestRuntime(t *testing.T) {
	tests := []struct {
		name string
		dp   *DivePlan
		want float64
	}{
		{
			name: "No stops",
			dp: &DivePlan{
				DescentRate: 20,
				AscentRate:  9,
				Stops:       []*DivePlanStop{},
			},
			want: 0.0,
		}, {
			name: "61min dive",
			dp: &DivePlan{
				DescentRate: 20,
				AscentRate:  9,
				Stops: []*DivePlanStop{
					{25.0, 13, false, ""},
					{18.0, 15, false, ""},
					{12.0, 23, false, ""},
					{5.0, 3, false, ""},
				},
			},
			want: 2 + 13 + 1 + 15 + 1 + 23 + 1 + 3 + 1,
		}, {
			name: "Bounce dive",
			dp: &DivePlan{
				DescentRate: 18,
				AscentRate:  6,
				Stops: []*DivePlanStop{
					{40.0, 1, false, ""},
				},
			},
			want: 3 + 1 + 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := tt.dp.Runtime()
			if rt != tt.want {
				t.Errorf("want: %f; got: %f", tt.want, rt)
			}
		})
	}
}

func TestDSRTable(t *testing.T) {
	tests := []struct {
		name string
		dp   *DivePlan
		want [][3]float64
	}{
		{
			name: "No stops",
			dp: &DivePlan{
				DescentRate: 20,
				AscentRate:  9,
				Stops:       []*DivePlanStop{},
			},
			want: [][3]float64{},
		}, {
			name: "61min dive",
			dp: &DivePlan{
				DescentRate: 20,
				AscentRate:  9,
				Stops: []*DivePlanStop{
					{25.0, 13, false, ""},
					{18.0, 15, false, ""},
					{12.0, 23, false, ""},
					{5.0, 3, false, ""},
				},
			},
			want: [][3]float64{
				{25.0, 13.0, 15.0},
				{18.0, 15.0, 31.0},
				{12.0, 23.0, 55.0},
				{5.0, 3.0, 59.0}},
		}, {
			name: "Bounce dive",
			dp: &DivePlan{
				DescentRate: 18,
				AscentRate:  6,
				Stops: []*DivePlanStop{
					{40.0, 1, false, ""},
					{5.0, 3, false, ""},
				},
			},
			want: [][3]float64{{40.0, 1.0, 4.0}, {5.0, 3.0, 13.0}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, s := range *tt.dp.DSRTable() {
				if s[0] != tt.want[i][0] ||
					s[1] != tt.want[i][1] ||
					s[2] != tt.want[i][2] {
					t.Errorf("want: %v; got: %v", tt.want, s)
				}
			}
		})
	}
}
