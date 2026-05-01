//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEgyptianRatscrewPlayer_NewAndStock(t *testing.T) {
	p := NewEgyptianRatscrewPlayer(true)
	assert.NotNil(t, p)
	assert.True(t, p.GetIsHuman())
	assert.False(t, p.HasStock())
	assert.Equal(t, 0, p.GetStockSize())
	assert.Nil(t, p.DrawTop())
}

func TestEgyptianRatscrewPlayer_AddBottomAndDraw(t *testing.T) {
	p := NewEgyptianRatscrewPlayer(false)
	p.AddToStockBottom(card(2), card(3), card(4))
	assert.Equal(t, 3, p.GetStockSize())
	assert.True(t, p.HasStock())

	c := p.DrawTop()
	assert.Equal(t, 2, c.GetValue())
	assert.Equal(t, 2, p.GetStockSize())
}

func TestEgyptianRatscrewPlayer_AddTop(t *testing.T) {
	p := NewEgyptianRatscrewPlayer(false)
	p.AddToStockBottom(card(2), card(3))
	p.AddToStockTop(card(9))
	assert.Equal(t, 3, p.GetStockSize())
	assert.Equal(t, 9, p.DrawTop().GetValue())
}

func TestEgyptianRatscrewPlayer_ResetStock(t *testing.T) {
	p := NewEgyptianRatscrewPlayer(false)
	p.AddToStockBottom(card(5))
	p.ResetStock()
	assert.False(t, p.HasStock())
}

func TestEgyptianRatscrewPlayer_JSONRoundtrip(t *testing.T) {
	p := NewEgyptianRatscrewPlayer(true)
	p.AddToStockBottom(card(7), card(8))

	data, err := json.Marshal(p)
	assert.NoError(t, err)
	var d EgyptianRatscrewPlayer
	assert.NoError(t, json.Unmarshal(data, &d))
	assert.Equal(t, p.GetIsHuman(), d.GetIsHuman())
	assert.Equal(t, 2, d.GetStockSize())
	assert.Equal(t, 7, d.DrawTop().GetValue())
}

func TestEgyptianRatscrewPlayer_UnmarshalInvalidJSON(t *testing.T) {
	var p EgyptianRatscrewPlayer
	assert.Error(t, p.UnmarshalJSON([]byte("not json")))
}

func TestEgyptianRatscrewPlayer_UnmarshalNilStock(t *testing.T) {
	var p EgyptianRatscrewPlayer
	// stock 欠如 → 空スライスにフォールバック
	assert.NoError(t, p.UnmarshalJSON([]byte(`{"gp":null}`)))
	assert.NotNil(t, p.GamePlayer)
	assert.Equal(t, 0, p.GetStockSize())
}
