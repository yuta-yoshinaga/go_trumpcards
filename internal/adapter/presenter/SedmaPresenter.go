//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// sedmaTrickPoints returns the points in a Sedma trick; each ace and ten is worth 10 points.
func sedmaTrickPoints(trick []*domain.TrickCard) int {
	points := 0
	for _, tc := range trick {
		if tc == nil || tc.Card == nil {
			continue
		}
		if value := tc.Card.GetValue(); value == 1 || value == 10 {
			points += 10
		}
	}
	return points
}
