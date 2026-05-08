//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestSlapjackCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := new(presenter.SlapjackCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		g := setupSlapjackTest()
		result := p.Output(g, nil)
		assert.Contains(t, result, "Slapjack")
		assert.Contains(t, result, "CPU:")
		assert.Contains(t, result, "あなた:")
		assert.Contains(t, result, "[場札]")
	})

	t.Run("human turn prompt", func(t *testing.T) {
		g := setupSlapjackTest()
		result := p.Output(g, nil)
		assert.Contains(t, result, "あなたの番です")
	})

	t.Run("error", func(t *testing.T) {
		g := setupSlapjackTest()
		result := p.Output(g, errors.New("oops"))
		assert.Contains(t, result, "oops")
	})

	t.Run("win message", func(t *testing.T) {
		g := setupSlapjackTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(0)
		raw["ph"], _ = json.Marshal(domain.SlapjackPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		assert.Contains(t, p.Output(g, nil), "あなたの勝ちです")
	})

	t.Run("lose message", func(t *testing.T) {
		g := setupSlapjackTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(1)
		raw["ph"], _ = json.Marshal(domain.SlapjackPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		assert.Contains(t, p.Output(g, nil), "CPUの勝ちです")
	})

	t.Run("cpu turn prompt", func(t *testing.T) {
		g := setupSlapjackTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ct"], _ = json.Marshal(1)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)
		assert.Contains(t, p.Output(g, nil), "CPU の番です")
	})

	t.Run("jack on top prompts slap", func(t *testing.T) {
		g := setupSlapjackTest()
		// Force a J on top via a step using a hand-crafted setup.
		g.GetPlayer(0).ResetStock()
		g.GetPlayer(0).AddToStockTop(domain.NewCard(domain.CardDesignSpade, domain.SlapjackJackValue, false))
		assert.NoError(t, g.Step())
		assert.Contains(t, p.Output(g, nil), "j (slap)")
	})

	t.Run("last event correct slap by human", func(t *testing.T) {
		g := setupSlapjackTest()
		g.GetPlayer(0).ResetStock()
		g.GetPlayer(0).AddToStockTop(domain.NewCard(domain.CardDesignSpade, domain.SlapjackJackValue, false))
		assert.NoError(t, g.Step())
		assert.NoError(t, g.Slap(0))
		assert.Contains(t, p.Output(g, nil), "あなたが正しくスラップ")
	})

	t.Run("last event correct slap by cpu", func(t *testing.T) {
		g := setupSlapjackTest()
		g.GetPlayer(0).ResetStock()
		g.GetPlayer(0).AddToStockTop(domain.NewCard(domain.CardDesignSpade, domain.SlapjackJackValue, false))
		assert.NoError(t, g.Step())
		assert.NoError(t, g.Slap(1))
		assert.Contains(t, p.Output(g, nil), "CPU が先にスラップ")
	})

	t.Run("last event wrong slap human", func(t *testing.T) {
		g := setupSlapjackTest()
		// Force a non-Jack on top of player 0's stock so Step always flips a
		// non-Jack — without this the flipped card is whatever Reset() shuffled
		// to the top, so ~1/13 of the time it's a Jack and the subsequent
		// Slap(0) is registered as a CORRECT slap, breaking the assertion.
		g.GetPlayer(0).ResetStock()
		g.GetPlayer(0).AddToStockTop(domain.NewCard(domain.CardDesignSpade, 2, false))
		assert.NoError(t, g.Step())
		// Force the last event to wrong-slap by human via Slap
		_ = g.Slap(0)
		assert.Contains(t, p.Output(g, nil), "誤スラップ")
	})

	t.Run("last event wrong slap cpu", func(t *testing.T) {
		g := setupSlapjackTest()
		// Same determinism fix as the human variant above.
		g.GetPlayer(0).ResetStock()
		g.GetPlayer(0).AddToStockTop(domain.NewCard(domain.CardDesignSpade, 2, false))
		assert.NoError(t, g.Step())
		_ = g.Slap(1)
		assert.Contains(t, p.Output(g, nil), "CPU が誤スラップ")
	})
}

func TestSlapjackCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.SlapjackCuiPresenter)
	g := setupSlapjackTest()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
