package presenter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func TestShitheadActionToOutput_PickupCards(t *testing.T) {
	facedown := &domain.ShitheadCpuAction{
		Source:      domain.ShitheadSourceFaceDown,
		PlayedCards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, true)},
		Pickup:      true,
	}
	assert.Len(t, shitheadActionToOutput(facedown).PlayedCards, 1)
	assert.Equal(t, 3, shitheadActionToOutput(facedown).PlayedCards[0].Value)

	hand := &domain.ShitheadCpuAction{
		Source:      domain.ShitheadSourceHand,
		PlayedCards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, true)},
		Pickup:      true,
	}
	assert.NotNil(t, shitheadActionToOutput(hand).PlayedCards)
	assert.Empty(t, shitheadActionToOutput(hand).PlayedCards)
}

func TestFormatShitheadAction_FacedownPickup(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	s := domain.NewDefaultShithead()
	s.Reset()
	var game interfaces.ShitheadGame = s
	card := domain.NewCard(domain.CardDesignSpade, 3, true)

	got := formatShitheadAction(game, &domain.ShitheadCpuAction{
		PlayerIdx:   0,
		Source:      domain.ShitheadSourceFaceDown,
		PlayedCards: []*domain.Card{card},
		Pickup:      true,
	})
	assert.Contains(t, got, "あなた が ♠3 をめくりましたが、出せないため場札を引き取りました")

	normal := formatShitheadAction(game, &domain.ShitheadCpuAction{
		PlayerIdx: 0,
		Pickup:    true,
	})
	assert.Contains(t, normal, "あなた が場札を引き取りました")
	assert.NotContains(t, normal, "めくりました")
}
