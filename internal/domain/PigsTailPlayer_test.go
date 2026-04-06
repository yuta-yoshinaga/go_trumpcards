//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPigsTailPlayer(t *testing.T) {
	t.Run("human player", func(t *testing.T) {
		p := NewPigsTailPlayer(true)
		assert.True(t, p.GetIsHuman())
		assert.False(t, p.GetIsFinished())
		assert.Equal(t, 0, p.GetCardsSize())
	})
	t.Run("cpu player", func(t *testing.T) {
		p := NewPigsTailPlayer(false)
		assert.False(t, p.GetIsHuman())
	})
}

func TestPigsTailPlayer_JSONRoundTrip(t *testing.T) {
	p := NewPigsTailPlayer(true)
	p.AddCard(NewCard(CardDesignSpade, 1, false))
	p.AddCard(NewCard(CardDesignHeart, 5, false))

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored PigsTailPlayer
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)
	assert.True(t, restored.GetIsHuman())
	assert.Equal(t, 2, restored.GetCardsSize())
}

func TestPigsTailPlayer_UnmarshalJSON_NilGamePlayer(t *testing.T) {
	data := []byte(`{"gp":null}`)
	var p PigsTailPlayer
	err := json.Unmarshal(data, &p)
	require.NoError(t, err)
	assert.False(t, p.GetIsHuman())
	assert.Equal(t, 0, p.GetCardsSize())
}

func TestPigsTailPlayer_UnmarshalJSON_InvalidJSON(t *testing.T) {
	var p PigsTailPlayer
	err := json.Unmarshal([]byte(`{invalid`), &p)
	assert.Error(t, err)
}
