//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestEgyptianRatscrewCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)

	p := new(presenter.EgyptianRatscrewCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		result := p.Output(g, nil)
		assert.Contains(t, result, "Egyptian Ratscrew")
		assert.Contains(t, result, "CPU:")
		assert.Contains(t, result, "あなた:")
		assert.Contains(t, result, "[場札]")
	})

	t.Run("human turn prompt", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		result := p.Output(g, nil)
		assert.Contains(t, result, "あなたの番です")
	})

	t.Run("error", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		result := p.Output(g, errors.New("oops"))
		assert.Contains(t, result, "oops")
	})

	t.Run("win message", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(0)
		raw["ph"], _ = json.Marshal(domain.EgyptianRatscrewPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		assert.Contains(t, p.Output(g, nil), "あなたの勝ちです")
	})

	t.Run("lose message", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(1)
		raw["ph"], _ = json.Marshal(domain.EgyptianRatscrewPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)

		assert.Contains(t, p.Output(g, nil), "CPUの勝ちです")
	})

	t.Run("cpu turn prompt", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ct"], _ = json.Marshal(1)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)
		assert.Contains(t, p.Output(g, nil), "CPU の番です")
	})

	t.Run("slappable on top prompts slap", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		// 場に同ランクのペアを直接投入
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		// centerPile に [{value:5}, {value:5}] を入れる
		card := domain.NewCard(domain.CardDesignSpade, 5, false)
		cardData, _ := json.Marshal([]*domain.Card{card, card})
		raw["cp"] = cardData
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)
		result := p.Output(g, nil)
		assert.Contains(t, result, "ペア/サンドイッチ")
		assert.Contains(t, result, "j (slap)")
	})

	t.Run("chance battle indicator", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["cr"], _ = json.Marshal(2)
		raw["ci"], _ = json.Marshal(0)
		raw["ct"], _ = json.Marshal(0) // human must answer the chance
		// A King sits on top, so the trigger line renders.
		raw["cp"], _ = json.Marshal([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 13, true)})
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)
		out := p.Output(g, nil)
		assert.Contains(t, out, "チャンスバトル中")
		assert.Contains(t, out, i18n.Tf("egyptianratscrew.chanceResponder",
			"name", i18n.T("egyptianratscrew.responderHuman")))
		triggerPrefix := strings.SplitN(i18n.T("egyptianratscrew.chanceTrigger"), "{{", 2)[0]
		assert.Contains(t, out, triggerPrefix)
	})

	t.Run("last event correct slap (pair) by human", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		// last event: SlapCorrect, player 0, reason Pair
		evt := domain.EgyptianRatscrewLastEvent{
			Kind:       domain.EgyptianRatscrewEventSlapCorrect,
			PlayerIdx:  0,
			SlapReason: domain.EgyptianRatscrewSlapReasonPair,
		}
		evtData, _ := json.Marshal(evt)
		raw["le"] = evtData
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)
		result := p.Output(g, nil)
		assert.Contains(t, result, "ペアスラップ")
		assert.Contains(t, result, "あなた")
	})

	t.Run("last event correct slap (sandwich) by cpu", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		evt := domain.EgyptianRatscrewLastEvent{
			Kind:       domain.EgyptianRatscrewEventSlapCorrect,
			PlayerIdx:  1,
			SlapReason: domain.EgyptianRatscrewSlapReasonSandwich,
		}
		evtData, _ := json.Marshal(evt)
		raw["le"] = evtData
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)
		result := p.Output(g, nil)
		assert.Contains(t, result, "サンドイッチスラップ")
		assert.Contains(t, result, "CPU")
	})

	t.Run("last event wrong slap human", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		evt := domain.EgyptianRatscrewLastEvent{
			Kind:      domain.EgyptianRatscrewEventSlapWrong,
			PlayerIdx: 0,
		}
		evtData, _ := json.Marshal(evt)
		raw["le"] = evtData
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)
		assert.Contains(t, p.Output(g, nil), "誤スラップ")
	})

	t.Run("last event wrong slap cpu", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		evt := domain.EgyptianRatscrewLastEvent{
			Kind:      domain.EgyptianRatscrewEventSlapWrong,
			PlayerIdx: 1,
		}
		evtData, _ := json.Marshal(evt)
		raw["le"] = evtData
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)
		assert.Contains(t, p.Output(g, nil), "CPU が誤スラップ")
	})

	t.Run("last event chance win human", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		evt := domain.EgyptianRatscrewLastEvent{
			Kind:      domain.EgyptianRatscrewEventChanceWin,
			PlayerIdx: 0,
		}
		evtData, _ := json.Marshal(evt)
		raw["le"] = evtData
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)
		assert.Contains(t, p.Output(g, nil), "チャンスバトル勝利")
	})

	t.Run("last event chance win cpu", func(t *testing.T) {
		g := setupEgyptianRatscrewTest()
		data, _ := json.Marshal(g)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		evt := domain.EgyptianRatscrewLastEvent{
			Kind:      domain.EgyptianRatscrewEventChanceWin,
			PlayerIdx: 1,
		}
		evtData, _ := json.Marshal(evt)
		raw["le"] = evtData
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, g)
		assert.Contains(t, p.Output(g, nil), "CPU に場札を奪われた")
	})
}

func TestEgyptianRatscrewCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.EgyptianRatscrewCuiPresenter)
	g := setupEgyptianRatscrewTest()
	assert.NotEmpty(t, p.ActionLogOutput(g))
}
