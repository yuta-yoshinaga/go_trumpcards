package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestCard_Accesser(t *testing.T) {
	t.Run("success Accesser", func(t *testing.T) {
		e := domain.NewCard(domain.CardDesignSpade, domain.CardValueMax, true)
		assert.Equal(t, domain.CardDesignSpade, e.GetDesign())
		assert.Equal(t, domain.CardValueMax, e.GetValue())
		assert.Equal(t, true, e.GetDraw())
		e.SetDraw(false)
		assert.Equal(t, false, e.GetDraw())
	})
}
