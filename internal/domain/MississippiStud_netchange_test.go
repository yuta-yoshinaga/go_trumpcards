package domain

import "testing"

func TestMississippiStud_GetNetChange(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		ante    int
		streets [MississippiStudStreetCnt]int
		payout  int
		want    int
	}{
		{name: "win", ante: 100, streets: [MississippiStudStreetCnt]int{1, 1, 1}, payout: 800, want: 400},
		{name: "fold loss", ante: 100, payout: 0, want: -100},
		{name: "push", ante: 100, streets: [MississippiStudStreetCnt]int{1, 1, 1}, payout: 400, want: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := &MississippiStud{anteAmount: tc.ante, streetMultipliers: tc.streets, totalPayout: tc.payout}
			if got := m.GetNetChange(); got != tc.want {
				t.Errorf("GetNetChange() = %d, want %d", got, tc.want)
			}
		})
	}
}
