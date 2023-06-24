package entities_test

import (
	"fmt"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"
)

func TestTrumpCards_Method(t *testing.T) {
	t.Run("success Shuffle", func(t *testing.T) {
		tc := entities.NewTrumpCards(2)
		tc.Shuffle()
	})
	t.Run("success DrawCard", func(t *testing.T) {
		tc := entities.NewTrumpCards(2)
		card := tc.DrawCard()
		fmt.Println(card)
	})
}
