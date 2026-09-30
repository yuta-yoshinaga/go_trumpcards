//go:build test && (!js || !wasm || extra6)

package domain

import "testing"

func TestJassFamilyCardPoints(t *testing.T) {
	cases := []struct{ value, plain, trump int }{{1, 11, 11}, {2, 0, 0}, {3, 0, 0}, {4, 0, 0}, {5, 0, 0}, {6, 0, 0}, {7, 0, 0}, {8, 0, 0}, {9, 0, 14}, {10, 10, 10}, {11, 2, 20}, {12, 3, 3}, {13, 4, 4}}
	for _, tc := range cases {
		for _, s := range []struct {
			name string
			suit int
			want int
		}{{"plain", CardDesignSpade, tc.plain}, {"trump", CardDesignHeart, tc.trump}} {
			t.Run(s.name+string(rune('A'+tc.value)), func(t *testing.T) {
				if got := JassFamilyCardPoints(NewCard(s.suit, tc.value, false), CardDesignHeart); got != s.want {
					t.Errorf("value %d: got %d want %d", tc.value, got, s.want)
				}
			})
		}
	}
	if got := JassFamilyCardPoints(nil, CardDesignHeart); got != 0 {
		t.Errorf("nil: got %d want 0", got)
	}
}
