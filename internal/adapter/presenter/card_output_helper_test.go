package presenter

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestCardToOutput(t *testing.T) {
	t.Run("SPADE", func(t *testing.T) {
		card := cardToOutput(domain.NewCard(domain.CardDesignSpade, 1, false))
		assert.Equal(t, "SPADE", card.Design)
		assert.Equal(t, 1, card.Value)
	})
	t.Run("CLOVER", func(t *testing.T) {
		card := cardToOutput(domain.NewCard(domain.CardDesignClover, 5, false))
		assert.Equal(t, "CLOVER", card.Design)
		assert.Equal(t, 5, card.Value)
	})
	t.Run("HEART", func(t *testing.T) {
		card := cardToOutput(domain.NewCard(domain.CardDesignHeart, 10, false))
		assert.Equal(t, "HEART", card.Design)
		assert.Equal(t, 10, card.Value)
	})
	t.Run("DIAMOND", func(t *testing.T) {
		card := cardToOutput(domain.NewCard(domain.CardDesignDiamond, 13, false))
		assert.Equal(t, "DIAMOND", card.Design)
		assert.Equal(t, 13, card.Value)
	})
	t.Run("JOKER", func(t *testing.T) {
		card := cardToOutput(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		assert.Equal(t, "JOKER", card.Design)
		assert.Equal(t, 0, card.Value)
	})
	t.Run("nil input", func(t *testing.T) {
		assert.Nil(t, cardToOutput(nil))
	})
}

// mockCardHolder is a test implementation of cardHolder
type mockCardHolder struct {
	cards []*domain.Card
}

func (m *mockCardHolder) GetCardsSize() int {
	return len(m.cards)
}

func (m *mockCardHolder) GetCard(i int) *domain.Card {
	return m.cards[i]
}

func TestPlayerCardsToOutput(t *testing.T) {
	t.Run("shouldShow true with cards", func(t *testing.T) {
		holder := &mockCardHolder{
			cards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
				domain.NewCard(domain.CardDesignHeart, 13, false),
			},
		}
		result := playerCardsToOutput(holder, true)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "SPADE", result[0].Design)
		assert.Equal(t, 1, result[0].Value)
		assert.Equal(t, "HEART", result[1].Design)
		assert.Equal(t, 13, result[1].Value)
	})
	t.Run("shouldShow false", func(t *testing.T) {
		holder := &mockCardHolder{
			cards: []*domain.Card{
				domain.NewCard(domain.CardDesignSpade, 1, false),
			},
		}
		result := playerCardsToOutput(holder, false)
		assert.NotNil(t, result)
		assert.Equal(t, 0, len(result))
	})
	t.Run("shouldShow true with zero cards", func(t *testing.T) {
		holder := &mockCardHolder{cards: []*domain.Card{}}
		result := playerCardsToOutput(holder, true)
		assert.NotNil(t, result)
		assert.Equal(t, 0, len(result))
	})
}

func TestCardsToOutput(t *testing.T) {
	t.Run("normal slice", func(t *testing.T) {
		cards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 13, false),
		}
		result := cardsToOutput(cards)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "SPADE", result[0].Design)
		assert.Equal(t, 1, result[0].Value)
		assert.Equal(t, "HEART", result[1].Design)
		assert.Equal(t, 13, result[1].Value)
	})
	t.Run("nil slice", func(t *testing.T) {
		assert.Nil(t, cardsToOutput(nil))
	})
	t.Run("empty slice", func(t *testing.T) {
		result := cardsToOutput([]*domain.Card{})
		assert.NotNil(t, result)
		assert.Equal(t, 0, len(result))
	})
}

func TestBuildTrickCards(t *testing.T) {
	type fakeTrick struct {
		idx  int
		card *domain.Card
	}
	mapper := func(tc fakeTrick) struct {
		PlayerIdx int
		Design    string
	} {
		return struct {
			PlayerIdx int
			Design    string
		}{PlayerIdx: tc.idx, Design: cardDesignToString(tc.card.GetDesign())}
	}

	t.Run("empty slice", func(t *testing.T) {
		result := buildTrickCards([]fakeTrick{}, mapper)
		assert.NotNil(t, result)
		assert.Equal(t, 0, len(result))
	})
	t.Run("multiple elements", func(t *testing.T) {
		tricks := []fakeTrick{
			{idx: 0, card: domain.NewCard(domain.CardDesignSpade, 1, false)},
			{idx: 2, card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		}
		result := buildTrickCards(tricks, mapper)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, 0, result[0].PlayerIdx)
		assert.Equal(t, "SPADE", result[0].Design)
		assert.Equal(t, 2, result[1].PlayerIdx)
		assert.Equal(t, "HEART", result[1].Design)
	})
}

func TestCardsToOutputOrEmpty(t *testing.T) {
	t.Run("nil slice returns empty", func(t *testing.T) {
		result := cardsToOutputOrEmpty(nil)
		assert.NotNil(t, result)
		assert.Equal(t, 0, len(result))
	})
	t.Run("non-nil slice", func(t *testing.T) {
		cards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
		}
		result := cardsToOutputOrEmpty(cards)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, "SPADE", result[0].Design)
	})
}
