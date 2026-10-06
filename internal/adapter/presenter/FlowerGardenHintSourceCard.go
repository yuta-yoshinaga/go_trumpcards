//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// flowerGardenHintSourceCard returns the card referenced by a Flower Garden hint,
// or nil when the hint's source zone or indices are invalid.
func flowerGardenHintSourceCard(tableau [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard, reserve []*domain.Card, hint *domain.FlowerGardenHint) *domain.Card {
	if hint.FromZone == "reserve" {
		if hint.FromCol >= 0 && hint.FromCol < len(reserve) {
			return reserve[hint.FromCol]
		}
		return nil
	}
	if hint.FromZone == "tableau" && hint.FromCol >= 0 && hint.FromCol < len(tableau) {
		col := tableau[hint.FromCol]
		if hint.CardIndex >= 0 && hint.CardIndex < len(col) {
			return col[hint.CardIndex].Card
		}
	}
	return nil
}
