//go:build test

package domain

import "testing"

func TestAceTenCardPoints(t *testing.T) {
	cases := []struct{ value, want int }{{1, 11}, {2, 0}, {3, 0}, {4, 0}, {5, 0}, {6, 0}, {7, 0}, {8, 0}, {9, 0}, {10, 10}, {11, 2}, {12, 3}, {13, 4}}
	for _, tc := range cases {
		t.Run(string(rune('A'+tc.value)), func(t *testing.T) {
			if got := AceTenCardPoints(NewCard(CardDesignSpade, tc.value, false)); got != tc.want {
				t.Errorf("value %d: got %d want %d", tc.value, got, tc.want)
			}
		})
	}
	if got := AceTenCardPoints(nil); got != 0 {
		t.Errorf("nil: got %d want 0", got)
	}
}
