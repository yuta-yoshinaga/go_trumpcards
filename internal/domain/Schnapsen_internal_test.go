package domain

import "testing"

func TestSchnapsenDetermineWinner(t *testing.T) {
	cases := []struct {
		name   string
		p0, p1 int
		lead   int
		want   int
	}{
		{name: "player zero reaches threshold", p0: 70, p1: 30, lead: 1, want: 0},
		{name: "player one reaches threshold", p0: 30, p1: 70, lead: 0, want: 1},
		{name: "fallback to player one", p0: 40, p1: 30, lead: 1, want: 1},
		{name: "fallback to player zero", p0: 40, p1: 30, lead: 0, want: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Schnapsen{playerPoints: []int{tc.p0, tc.p1}, leadPlayerIdx: tc.lead}
			if got := s.determineWinner(); got != tc.want {
				t.Errorf("determineWinner() = %d, want %d", got, tc.want)
			}
		})
	}
}
