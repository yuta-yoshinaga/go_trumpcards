//go:build test

package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// rbState builds a RussianBank in an arbitrary state via JSON restore.
func rbState(t *testing.T, js string) *domain.RussianBank {
	t.Helper()
	g := domain.NewDefaultRussianBank()
	if err := json.Unmarshal([]byte(js), g); err != nil {
		t.Fatalf("restore: %v", err)
	}
	return g
}

const rbTwoEmptyPlayers = `"pl":[{"n":"You","c":false,"s":0,"r":[],"h":[],"w":[]},` +
	`{"n":"CPU","c":true,"s":1,"r":[],"h":[],"w":[]}],"cf":{"cd":1}`

func TestRussianBankCuiPresenter_Output(t *testing.T) {
	p := new(presenter.RussianBankCuiPresenter)

	t.Run("human turn prompt", func(t *testing.T) {
		g := rbState(t, "{"+rbTwoEmptyPlayers+`,"ph":1,"cu":0}`)
		out := p.Output(g, nil)
		assert.Contains(t, out, "Russian Bank")
		assert.Contains(t, out, "ファウンデーション")
		assert.Contains(t, out, "あなたの手番")
	})

	t.Run("cpu turn prompt", func(t *testing.T) {
		g := rbState(t, "{"+rbTwoEmptyPlayers+`,"ph":1,"cu":1}`)
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("stop available", func(t *testing.T) {
		// Opponent (seat 1) has an Ace on its reserve playable to an empty foundation.
		js := `{"pl":[{"n":"You","c":false,"s":0,"r":[],"h":[],"w":[]},` +
			`{"n":"CPU","c":true,"s":1,"r":[{"d":0,"v":1,"w":true}],"h":[],"w":[]}],` +
			`"cf":{"cd":1},"ph":1,"cu":0}`
		g := rbState(t, js)
		assert.True(t, g.CanCallStop())
		assert.NotEmpty(t, p.Output(g, nil))
	})

	t.Run("game end banners", func(t *testing.T) {
		win := rbState(t, "{"+rbTwoEmptyPlayers+`,"ph":2,"cu":0,"wn":0}`)
		assert.Contains(t, p.Output(win, nil), "勝ち")
		draw := rbState(t, "{"+rbTwoEmptyPlayers+`,"ph":2,"cu":0,"wn":-1}`)
		assert.Contains(t, p.Output(draw, nil), "引き分け")
	})

	t.Run("error block", func(t *testing.T) {
		g := rbState(t, "{"+rbTwoEmptyPlayers+`,"ph":1,"cu":0}`)
		assert.NotEmpty(t, p.Output(g, assert.AnError))
	})

	t.Run("hint output", func(t *testing.T) {
		g := domain.NewDefaultRussianBank()
		g.Reset()
		assert.NotEmpty(t, p.HintOutput(g))
		// A game not on the human's turn yields the no-hint message.
		noHint := rbState(t, "{"+rbTwoEmptyPlayers+`,"ph":1,"cu":1}`)
		assert.NotEmpty(t, p.HintOutput(noHint))
	})

	t.Run("action log", func(t *testing.T) {
		g := domain.NewDefaultRussianBank()
		g.Reset()
		assert.NotEmpty(t, p.ActionLogOutput(g))
	})
}
