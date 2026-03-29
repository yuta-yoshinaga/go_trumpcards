//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBridgePlayer(t *testing.T) {
	p := NewBridgePlayer(true, 0)
	assert.True(t, p.GetIsHuman())
	assert.Equal(t, 0, p.GetTeam())

	p2 := NewBridgePlayer(false, 1)
	assert.False(t, p2.GetIsHuman())
	assert.Equal(t, 1, p2.GetTeam())
}

func TestBridgePlayerResetRound(t *testing.T) {
	p := NewBridgePlayer(true, 0)
	p.AddCard(NewCard(CardDesignSpade, 1, true))
	p.AddTrick([]*Card{NewCard(CardDesignHeart, 2, true)})
	p.SetIsFinished(true)

	p.ResetRound()

	assert.Equal(t, 0, p.GetCardsSize())
	assert.Equal(t, 0, p.GetTrickCount())
	assert.False(t, p.GetIsFinished())
}

func TestBridgePlayerJSON(t *testing.T) {
	p := NewBridgePlayer(true, 1)
	p.AddCard(NewCard(CardDesignSpade, 1, true))
	p.AddTrick([]*Card{NewCard(CardDesignHeart, 5, true)})

	data, err := json.Marshal(p)
	require.NoError(t, err)

	p2 := &BridgePlayer{}
	err = json.Unmarshal(data, p2)
	require.NoError(t, err)

	assert.True(t, p2.GetIsHuman())
	assert.Equal(t, 1, p2.GetTeam())
	assert.Equal(t, 1, p2.GetCardsSize())
	assert.Equal(t, 1, p2.GetTrickCount())
}

func TestBridgePlayerJSONNilFields(t *testing.T) {
	data := []byte(`{"tm":0}`)
	p := &BridgePlayer{}
	err := json.Unmarshal(data, p)
	require.NoError(t, err)
	assert.False(t, p.GetIsHuman())
	assert.Equal(t, 0, p.GetTeam())
}
