//go:build test

package presenter_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestRussianBankWebPresenter_Output(t *testing.T) {
	p := new(presenter.RussianBankWebPresenter)

	t.Run("playing state serialises the board and players", func(t *testing.T) {
		g := domain.NewDefaultRussianBank()
		g.Reset()
		out := p.Output(g, nil)
		for _, frag := range []string{`"phase"`, `"players"`, `"tableau"`, `"foundations"`, `"reserveCount"`, `"canCallStop"`, `"messageCode":"russianbank.playing"`} {
			assert.Contains(t, out, frag)
		}
	})

	t.Run("error is surfaced in the message", func(t *testing.T) {
		g := domain.NewDefaultRussianBank()
		g.Reset()
		out := p.Output(g, assert.AnError)
		assert.Contains(t, out, assert.AnError.Error())
	})

	t.Run("winner messages", func(t *testing.T) {
		for _, tc := range []struct{ winner, code string }{
			{"0", "russianbank.result.humanWin"},
			{"1", "russianbank.result.cpuWin"},
			{"-1", "russianbank.result.draw"},
		} {
			js := "{" + rbTwoEmptyPlayers + `,"ph":2,"cu":0,"wn":` + tc.winner + `}`
			g := rbState(t, js)
			out := p.Output(g, nil)
			assert.Contains(t, out, tc.code)
			assert.Contains(t, out, `"gameEndFlag":true`)
		}
	})

	t.Run("hint output includes the hint object", func(t *testing.T) {
		g := domain.NewDefaultRussianBank()
		g.Reset()
		out := p.HintOutput(g)
		// After a deal there is at least one legal move, so a hint is present.
		assert.True(t, strings.Contains(out, `"hint"`) || strings.Contains(out, `"zone"`))
	})

	t.Run("action log is valid JSON", func(t *testing.T) {
		g := domain.NewDefaultRussianBank()
		g.Reset()
		out := p.ActionLogOutput(g)
		assert.True(t, json.Valid([]byte(out)))
	})
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
//
// Output 側にゲートは置きません。RussianBank.GetHint() が「人間の手番で、かつ
// 行動を選べる状態か」を自分で確かめて nil を返します。
func TestRussianBankWebPresenterOutputCarriesTheHint(t *testing.T) {
	g := domain.NewDefaultRussianBank()
	g.Reset()
	if g.GetHint() == nil {
		t.Fatal("fixture must actually produce a hint")
	}

	result := new(presenter.RussianBankWebPresenter).Output(g, nil)
	assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
}

// **なぜボタンが増えたのかを言う (#4817)。**CUI は同じ状態を黄色で明示している。
func TestRussianBankWebPresenter_StopAvailableMessage(t *testing.T) {
	p := new(presenter.RussianBankWebPresenter)

	// CPU (席 1) のリザーブトップを A にすると、ファウンデーションへ必ず置ける
	// = 強制手の取りこぼし。JSON 経由で盤面を作る。
	g := domain.NewDefaultRussianBank()
	g.Reset()
	data, err := json.Marshal(g)
	assert.NoError(t, err)
	var raw map[string]json.RawMessage
	assert.NoError(t, json.Unmarshal(data, &raw))
	var players []map[string]json.RawMessage
	assert.NoError(t, json.Unmarshal(raw["pl"], &players))
	players[1]["r"], err = json.Marshal([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
	assert.NoError(t, err)
	raw["pl"], err = json.Marshal(players)
	assert.NoError(t, err)
	newData, err := json.Marshal(raw)
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(newData, g))

	assert.True(t, g.CanCallStop(), "この盤面では stop を呼べる")
	var out controller.RussianBankWebOutput
	assert.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &out))
	assert.Equal(t, "russianbank.stopAvailable", out.MessageCode)

	// 呼べない盤面では従来どおり。
	players[1]["r"], err = json.Marshal([]*domain.Card{})
	assert.NoError(t, err)
	raw["pl"], err = json.Marshal(players)
	assert.NoError(t, err)
	newData, err = json.Marshal(raw)
	assert.NoError(t, err)
	assert.NoError(t, json.Unmarshal(newData, g))
	// 配り直後は CPU の捨て札が空なので、リザーブを空にすれば強制手は必ず消える
	// (シャッフル依存ではない)。
	assert.False(t, g.CanCallStop())
	var out2 controller.RussianBankWebOutput
	assert.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &out2))
	assert.Equal(t, "russianbank.playing", out2.MessageCode)
}
