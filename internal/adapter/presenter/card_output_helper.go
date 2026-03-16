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

// cardHolder はインデックスベースでカードを取得できるオブジェクトの共通インターフェース
type cardHolder interface {
	GetCardsSize() int
	GetCard(i int) *domain.Card
}

// playerCardsToOutput cardHolder のカードを WebOutputCard スライスに変換する。
// shouldShow が false の場合は空スライスを返す。
func playerCardsToOutput(holder cardHolder, shouldShow bool) []*controller.WebOutputCard {
	if !shouldShow {
		return make([]*controller.WebOutputCard, 0)
	}
	cards := make([]*controller.WebOutputCard, 0, holder.GetCardsSize())
	for i := 0; i < holder.GetCardsSize(); i++ {
		cards = append(cards, cardToOutput(holder.GetCard(i)))
	}
	return cards
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

// cardsToOutputOrEmpty カードスライスを共通WebOutputCardスライスに変換 (nil → 空スライス)
func cardsToOutputOrEmpty(cards []*domain.Card) []*controller.WebOutputCard {
	if cards == nil {
		return make([]*controller.WebOutputCard, 0)
	}
	return cardsToOutput(cards)
}
