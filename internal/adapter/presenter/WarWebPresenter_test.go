//go:build test

package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func setupWarTest() *domain.War {
	tc := domain.NewTrumpCards(0)
	players := []*domain.WarPlayer{
		domain.NewWarPlayer(true),
		domain.NewWarPlayer(false),
	}
	w := domain.NewWar(tc, players, domain.DefaultWarConfig())
	w.Reset()
	return w
}

func TestWarWebPresenter_Output(t *testing.T) {
	p := new(presenter.WarWebPresenter)

	t.Run("initial state", func(t *testing.T) {
		w := setupWarTest()
		result := p.Output(w, nil)
		assert.NotEmpty(t, result)

		var resObj controller.WarWebOutput
		assert.NoError(t, json.Unmarshal([]byte(result), &resObj))
		assert.Len(t, resObj.Players, 2)
		assert.Equal(t, 26, resObj.Players[0].DrawPileSize)
		assert.Equal(t, 26, resObj.Players[1].DrawPileSize)
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, -1, resObj.WinnerIdx)
		assert.Empty(t, resObj.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		w := setupWarTest()
		result := p.Output(w, errors.New("bad"))
		var resObj controller.WarWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "error", resObj.MessageCode)
		assert.Equal(t, "bad", resObj.Message)
	})

	t.Run("reveal populated after step", func(t *testing.T) {
		w := setupWarTest()
		_ = w.Step()
		result := p.Output(w, nil)
		var resObj controller.WarWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.PlayerRevealed)
		assert.NotNil(t, resObj.CpuRevealed)
		assert.Equal(t, 2, resObj.WarPotSize)
	})

	// **各ラウンドの決着も伝える。** 以前は最終勝敗のときしかコードが出ず、
	// 盤面の変化はリング色と不透明度だけだった。読み上げ利用者はどのラウンドで
	// 誰が勝ったのかも、いつ戦争になったのかも知りようがなかった (#5530)。
	t.Run("毎ラウンドの決着にコードが出る", func(t *testing.T) {
		override := func(phase domain.WarPhase, lastWinner int) string {
			w := setupWarTest()
			data, _ := json.Marshal(w)
			var raw map[string]json.RawMessage
			_ = json.Unmarshal(data, &raw)
			raw["ph"], _ = json.Marshal(phase)
			raw["lw"], _ = json.Marshal(lastWinner)
			newData, _ := json.Marshal(raw)
			require.NoError(t, json.Unmarshal(newData, w))
			require.Equal(t, phase, w.GetPhase(), "改竄が効いていない")

			var resObj controller.WarWebOutput
			require.NoError(t, json.Unmarshal([]byte(p.Output(w, nil)), &resObj))
			return resObj.MessageCode
		}

		assert.Equal(t, "war.round.humanWin", override(domain.WarPhaseResolved, 0))
		assert.Equal(t, "war.round.cpuWin", override(domain.WarPhaseResolved, 1))
		assert.Equal(t, "war.round.warBury", override(domain.WarPhaseWarBury, -1))

		// **負のコントロール: めくる前は何も言わない。** ここでコードが出ると
		// 1 手ごとに同じ文が読み上げられる。
		assert.Empty(t, override(domain.WarPhaseReveal, -1))
	})

	t.Run("game end message via JSON override", func(t *testing.T) {
		w := setupWarTest()
		data, _ := json.Marshal(w)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(0)
		raw["ph"], _ = json.Marshal(domain.WarPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, w)

		result := p.Output(w, nil)
		var resObj controller.WarWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "war.result.humanWin", resObj.MessageCode)
		assert.True(t, resObj.GameEndFlag)
	})

	t.Run("game end CPU win message", func(t *testing.T) {
		w := setupWarTest()
		data, _ := json.Marshal(w)
		var raw map[string]json.RawMessage
		_ = json.Unmarshal(data, &raw)
		raw["ge"], _ = json.Marshal(true)
		raw["wi"], _ = json.Marshal(1)
		raw["ph"], _ = json.Marshal(domain.WarPhaseGameEnd)
		newData, _ := json.Marshal(raw)
		_ = json.Unmarshal(newData, w)

		result := p.Output(w, nil)
		var resObj controller.WarWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, "war.result.cpuWin", resObj.MessageCode)
	})

}

func TestWarWebPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.WarWebPresenter)
	w := setupWarTest()
	assert.NotEmpty(t, p.ActionLogOutput(w))
}
