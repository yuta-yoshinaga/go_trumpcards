package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// cardToOutput カードを共通WebOutputCardに変換 (nil → nil)
func cardToOutput(card *domain.Card) *controller.WebOutputCard {
	if card == nil {
		return nil
	}
	return &controller.WebOutputCard{
		Design: cardDesignToString(card.GetDesign()),
		Value:  card.GetValue(),
	}
}

// cardsToOutput カードスライスを共通WebOutputCardスライスに変換 (nil → nil)
func cardsToOutput(cards []*domain.Card) []*controller.WebOutputCard {
	if cards == nil {
		return nil
	}
	result := make([]*controller.WebOutputCard, len(cards))
	for i, c := range cards {
		result[i] = cardToOutput(c)
	}
	return result
}
