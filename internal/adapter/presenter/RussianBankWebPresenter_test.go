//go:build test

package presenter_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

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
