package entities_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/assert"
)

func TestCard_Accesser(t *testing.T) {
	t.Run("success Accesser", func(t *testing.T) {
		e := entities.NewCard(entities.CardDesignSpade, entities.CardValueMax, true)
		assert.Equal(t, entities.CardDesignSpade, e.GetDesign())
		assert.Equal(t, entities.CardValueMax, e.GetValue())
		assert.Equal(t, true, e.GetDraw())
		e.SetDraw(false)
		assert.Equal(t, false, e.GetDraw())
	})
}
