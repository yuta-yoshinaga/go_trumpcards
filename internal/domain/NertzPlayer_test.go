//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newNertzCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func TestNewNertzPlayer(t *testing.T) {
	p := domain.NewNertzPlayer("Alice", true, 2)
	assert.Equal(t, "Alice", p.GetName())
	assert.True(t, p.GetIsCpu())
	assert.Equal(t, 2, p.GetDeckIdx())
	assert.Equal(t, 0, p.GetScore())
	assert.Equal(t, 0, p.NertzSize())
	assert.Equal(t, 0, p.WasteSize())
	assert.Equal(t, 0, p.StockSize())
	for i := 0; i < domain.NertzTableauCnt; i++ {
		assert.Equal(t, 0, p.TableauSize(i))
	}
}

func TestNertzPlayer_NertzPile(t *testing.T) {
	p := domain.NewNertzPlayer("p", false, 0)
	c1 := newNertzCard(domain.CardDesignSpade, 1)
	c2 := newNertzCard(domain.CardDesignHeart, 5)
	p.PushNertz(c1)
	p.PushNertz(c2)
	assert.Equal(t, 2, p.NertzSize())
	assert.Equal(t, c2, p.NertzTop()) // last pushed = top

	popped := p.PopNertz()
	assert.Equal(t, c2, popped)
	assert.Equal(t, 1, p.NertzSize())
	assert.Equal(t, c1, p.NertzTop())

	p.PopNertz()
	assert.Equal(t, 0, p.NertzSize())
	assert.Nil(t, p.NertzTop())
	assert.Nil(t, p.PopNertz())
}

func TestNertzPlayer_Tableau(t *testing.T) {
	p := domain.NewNertzPlayer("p", false, 0)
	c1 := newNertzCard(domain.CardDesignClover, 5)
	c2 := newNertzCard(domain.CardDesignHeart, 4)

	p.PushTableau(1, &domain.NertzTableauCard{Card: c1, FaceUp: true})
	p.PushTableau(1, &domain.NertzTableauCard{Card: c2, FaceUp: true})
	assert.Equal(t, 2, p.TableauSize(1))
	assert.Equal(t, c2, p.TableauTop(1))

	tail := p.TakeTableauTail(1, 1) // take from index 1 to end
	require.Len(t, tail, 1)
	assert.Equal(t, c2, tail[0].Card)
	assert.Equal(t, 1, p.TableauSize(1))

	// out-of-range column returns no-op
	assert.Equal(t, 0, p.TableauSize(99))
	assert.Nil(t, p.TableauTop(99))
	assert.Nil(t, p.TakeTableauTail(99, 0))
	p.PushTableau(99, &domain.NertzTableauCard{Card: c1, FaceUp: true}) // ignored
}

func TestNertzPlayer_Waste(t *testing.T) {
	p := domain.NewNertzPlayer("p", false, 0)
	c1 := newNertzCard(domain.CardDesignSpade, 7)
	c2 := newNertzCard(domain.CardDesignDiamond, 8)
	p.PushWaste(c1)
	p.PushWaste(c2)
	assert.Equal(t, 2, p.WasteSize())
	assert.Equal(t, c2, p.WasteTop())

	assert.Equal(t, c2, p.PopWaste())
	assert.Equal(t, 1, p.WasteSize())
	assert.Equal(t, c1, p.PopWaste())
	assert.Equal(t, 0, p.WasteSize())
	assert.Nil(t, p.WasteTop())
	assert.Nil(t, p.PopWaste())
}

func TestNertzPlayer_Stock(t *testing.T) {
	p := domain.NewNertzPlayer("p", false, 0)
	c1 := newNertzCard(domain.CardDesignSpade, 1)
	c2 := newNertzCard(domain.CardDesignHeart, 2)
	p.PushStock(c1)
	p.PushStock(c2)
	assert.Equal(t, 2, p.StockSize())

	got := p.PopStock()
	assert.Equal(t, c2, got)
	assert.Equal(t, 1, p.StockSize())
	p.PopStock()
	assert.Nil(t, p.PopStock())
}

func TestNertzPlayer_RecycleWasteToStock(t *testing.T) {
	p := domain.NewNertzPlayer("p", false, 0)
	c1 := newNertzCard(domain.CardDesignSpade, 1)
	c2 := newNertzCard(domain.CardDesignHeart, 2)
	c3 := newNertzCard(domain.CardDesignClover, 3)
	p.PushWaste(c1)
	p.PushWaste(c2)
	p.PushWaste(c3)
	p.RecycleWasteToStock()
	assert.Equal(t, 0, p.WasteSize())
	assert.Equal(t, 3, p.StockSize())
	// Order: waste (c1, c2, c3) reversed → stock so that PopStock returns c1 first
	assert.Equal(t, c1, p.PopStock())
	assert.Equal(t, c2, p.PopStock())
	assert.Equal(t, c3, p.PopStock())
}

func TestNertzPlayer_Score(t *testing.T) {
	p := domain.NewNertzPlayer("p", false, 0)
	p.AddScore(5)
	p.AddScore(-2)
	assert.Equal(t, 3, p.GetScore())
	p.SetScore(100)
	assert.Equal(t, 100, p.GetScore())
}

func TestNertzPlayer_ResetRoundPiles(t *testing.T) {
	p := domain.NewNertzPlayer("p", false, 0)
	p.PushNertz(newNertzCard(domain.CardDesignSpade, 1))
	p.PushTableau(0, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignClover, 5), FaceUp: true})
	p.PushWaste(newNertzCard(domain.CardDesignDiamond, 3))
	p.PushStock(newNertzCard(domain.CardDesignHeart, 9))
	p.SetScore(42)

	p.ResetRoundPiles()

	assert.Equal(t, 0, p.NertzSize())
	assert.Equal(t, 0, p.WasteSize())
	assert.Equal(t, 0, p.StockSize())
	for i := 0; i < domain.NertzTableauCnt; i++ {
		assert.Equal(t, 0, p.TableauSize(i))
	}
	assert.Equal(t, 42, p.GetScore(), "ResetRoundPiles must preserve cumulative score")
}

func TestNertzPlayer_JSON(t *testing.T) {
	p := domain.NewNertzPlayer("Bob", true, 1)
	p.PushNertz(newNertzCard(domain.CardDesignSpade, 4))
	p.PushTableau(2, &domain.NertzTableauCard{Card: newNertzCard(domain.CardDesignHeart, 9), FaceUp: true})
	p.PushWaste(newNertzCard(domain.CardDesignDiamond, 7))
	p.PushStock(newNertzCard(domain.CardDesignClover, 2))
	p.AddScore(33)

	data, err := json.Marshal(p)
	require.NoError(t, err)

	restored := &domain.NertzPlayer{}
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, p.GetName(), restored.GetName())
	assert.Equal(t, p.GetIsCpu(), restored.GetIsCpu())
	assert.Equal(t, p.GetDeckIdx(), restored.GetDeckIdx())
	assert.Equal(t, p.GetScore(), restored.GetScore())
	assert.Equal(t, p.NertzSize(), restored.NertzSize())
	assert.Equal(t, p.TableauSize(2), restored.TableauSize(2))
	assert.Equal(t, p.WasteSize(), restored.WasteSize())
	assert.Equal(t, p.StockSize(), restored.StockSize())
}
