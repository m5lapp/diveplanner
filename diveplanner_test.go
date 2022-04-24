package diveplanner

import "testing"

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

func TestCalcTransition(t *testing.T) {
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
			tt.dp.calcTransitions()
			for i, s := range tt.dp.Stops {
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
