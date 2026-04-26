//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewNertzFoundation(t *testing.T) {
	f := domain.NewNertzFoundation()
	assert.True(t, f.IsEmpty())
	assert.False(t, f.IsComplete())
	assert.Equal(t, 0, f.Size())
	assert.Nil(t, f.Top())
	assert.Equal(t, -1, f.Suit())
}

func TestNertzFoundation_CanAccept_Empty(t *testing.T) {
	f := domain.NewNertzFoundation()
	tests := []struct {
		name string
		card *domain.Card
		want bool
	}{
		{"ace of spades accepted", newNertzCard(domain.CardDesignSpade, 1), true},
		{"ace of hearts accepted", newNertzCard(domain.CardDesignHeart, 1), true},
		{"two of spades rejected", newNertzCard(domain.CardDesignSpade, 2), false},
		{"king of spades rejected", newNertzCard(domain.CardDesignSpade, 13), false},
		{"joker rejected", newNertzCard(domain.CardDesignJoker, 1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, f.CanAccept(tt.card))
		})
	}
}

func TestNertzFoundation_PushAceStartsSuit(t *testing.T) {
	f := domain.NewNertzFoundation()
	ace := newNertzCard(domain.CardDesignSpade, 1)
	require.NoError(t, f.Push(ace, 0))
	assert.False(t, f.IsEmpty())
	assert.Equal(t, 1, f.Size())
	assert.Equal(t, ace, f.Top())
	assert.Equal(t, domain.CardDesignSpade, f.Suit())
	assert.Equal(t, 0, f.ContributorAt(0))
}

func TestNertzFoundation_CanAccept_Suited(t *testing.T) {
	f := domain.NewNertzFoundation()
	require.NoError(t, f.Push(newNertzCard(domain.CardDesignClover, 1), 0))
	require.NoError(t, f.Push(newNertzCard(domain.CardDesignClover, 2), 1))

	tests := []struct {
		name string
		card *domain.Card
		want bool
	}{
		{"next rank same suit", newNertzCard(domain.CardDesignClover, 3), true},
		{"skip rank rejected", newNertzCard(domain.CardDesignClover, 4), false},
		{"wrong suit rejected", newNertzCard(domain.CardDesignSpade, 3), false},
		{"another ace rejected", newNertzCard(domain.CardDesignClover, 1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, f.CanAccept(tt.card))
		})
	}
}

func TestNertzFoundation_PushRejectsInvalid(t *testing.T) {
	f := domain.NewNertzFoundation()
	err := f.Push(newNertzCard(domain.CardDesignSpade, 5), 0)
	assert.Error(t, err)
	assert.True(t, f.IsEmpty())
}

func TestNertzFoundation_IsComplete(t *testing.T) {
	f := domain.NewNertzFoundation()
	for v := 1; v <= 13; v++ {
		require.NoError(t, f.Push(newNertzCard(domain.CardDesignDiamond, v), v%4))
	}
	assert.True(t, f.IsComplete())
	// further push rejected
	err := f.Push(newNertzCard(domain.CardDesignDiamond, 1), 0)
	assert.Error(t, err)
}

func TestNertzFoundation_CountByContributor(t *testing.T) {
	f := domain.NewNertzFoundation()
	require.NoError(t, f.Push(newNertzCard(domain.CardDesignHeart, 1), 0))
	require.NoError(t, f.Push(newNertzCard(domain.CardDesignHeart, 2), 1))
	require.NoError(t, f.Push(newNertzCard(domain.CardDesignHeart, 3), 0))
	require.NoError(t, f.Push(newNertzCard(domain.CardDesignHeart, 4), 0))

	assert.Equal(t, 3, f.CountByContributor(0))
	assert.Equal(t, 1, f.CountByContributor(1))
	assert.Equal(t, 0, f.CountByContributor(2))
}

func TestNertzFoundation_JSON(t *testing.T) {
	f := domain.NewNertzFoundation()
	require.NoError(t, f.Push(newNertzCard(domain.CardDesignClover, 1), 2))
	require.NoError(t, f.Push(newNertzCard(domain.CardDesignClover, 2), 0))

	data, err := json.Marshal(f)
	require.NoError(t, err)

	restored := &domain.NertzFoundation{}
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, f.Size(), restored.Size())
	assert.Equal(t, f.Suit(), restored.Suit())
	assert.Equal(t, f.ContributorAt(0), restored.ContributorAt(0))
	assert.Equal(t, f.ContributorAt(1), restored.ContributorAt(1))
	assert.Equal(t, f.Top().GetValue(), restored.Top().GetValue())
	assert.Equal(t, f.Top().GetDesign(), restored.Top().GetDesign())
}
