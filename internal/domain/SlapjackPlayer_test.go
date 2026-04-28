//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlapjackPlayer_NewAndAccessors(t *testing.T) {
	p := NewSlapjackPlayer(true)
	assert.True(t, p.GetIsHuman())
	assert.False(t, p.HasStock())
	assert.Equal(t, 0, p.GetStockSize())
}

func TestSlapjackPlayer_AddDrawAndOrder(t *testing.T) {
	p := NewSlapjackPlayer(false)
	p.AddToStockBottom(card(2), card(3), card(4))
	assert.Equal(t, 3, p.GetStockSize())
	assert.True(t, p.HasStock())

	c := p.DrawTop()
	assert.Equal(t, 2, c.GetValue())
	assert.Equal(t, 2, p.GetStockSize())

	c = p.DrawTop()
	assert.Equal(t, 3, c.GetValue())

	// AddToStockTop は先頭挿入
	p.AddToStockTop(card(10))
	assert.Equal(t, 10, p.DrawTop().GetValue())
	assert.Equal(t, 4, p.DrawTop().GetValue())

	// 空ストックからの DrawTop は nil
	assert.Nil(t, p.DrawTop())
}

func TestSlapjackPlayer_ResetStock(t *testing.T) {
	p := NewSlapjackPlayer(false)
	p.AddToStockBottom(card(5), card(6))
	p.ResetStock()
	assert.False(t, p.HasStock())
}

func TestSlapjackPlayer_JSON(t *testing.T) {
	p := NewSlapjackPlayer(true)
	p.AddToStockBottom(card(7), card(8))

	data, err := json.Marshal(p)
	assert.NoError(t, err)

	var decoded SlapjackPlayer
	assert.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, p.GetIsHuman(), decoded.GetIsHuman())
	assert.Equal(t, p.GetStockSize(), decoded.GetStockSize())
	assert.Equal(t, p.DrawTop().GetValue(), decoded.DrawTop().GetValue())
}

func TestSlapjackPlayer_JSON_NilStock(t *testing.T) {
	// GamePlayer == nil でもデコードできることを確認 (defensive default branch)
	raw := []byte(`{"gp":null,"st":null}`)
	var p SlapjackPlayer
	assert.NoError(t, json.Unmarshal(raw, &p))
	assert.NotNil(t, p.GamePlayer)
	assert.NotNil(t, p)
	assert.Equal(t, 0, p.GetStockSize())
}
