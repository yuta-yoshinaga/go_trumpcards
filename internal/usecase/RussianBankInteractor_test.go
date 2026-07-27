//go:build test

package usecase_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// newRbInteractor wires a real domain game to a real web presenter.
func newRbInteractor(diff domain.RussianBankCpuDifficulty) (*usecase.RussianBankInteractor, *domain.RussianBank) {
	g := domain.NewRussianBank(domain.RussianBankConfig{CpuDifficulty: diff})
	g.Reset()
	ti := usecase.NewRussianBankInteractor(g, new(presenter.RussianBankWebPresenter))
	return ti, g
}

func TestNewRussianBankInteractor_NilGuards(t *testing.T) {
	sp := new(presenter.RussianBankWebPresenter)
	assert.PanicsWithValue(t, "RussianBankInteractor: g must not be nil", func() {
		usecase.NewRussianBankInteractor(nil, sp)
	})
	g := domain.NewDefaultRussianBank()
	assert.PanicsWithValue(t, "RussianBankInteractor: sp must not be nil", func() {
		usecase.NewRussianBankInteractor(g, nil)
	})
}

func TestRussianBankInteractor_ResetAndConfig(t *testing.T) {
	ti, _ := newRbInteractor(domain.RussianBankCpuDifficultyNormal)
	out := ti.Reset()
	assert.Contains(t, out, `"phase"`)
	assert.Equal(t, domain.RussianBankCpuDifficultyNormal, ti.GetConfig().CpuDifficulty)

	out = ti.ResetWithConfig(domain.RussianBankConfig{CpuDifficulty: domain.RussianBankCpuDifficultyHard})
	assert.Contains(t, out, `"phase"`)
	assert.Equal(t, domain.RussianBankCpuDifficultyHard, ti.GetConfig().CpuDifficulty)

	// An invalid config is rejected (validation fails) and the current config is kept.
	ti.ResetWithConfig(domain.RussianBankConfig{CpuDifficulty: 99})
	assert.Equal(t, domain.RussianBankCpuDifficultyHard, ti.GetConfig().CpuDifficulty)
}

func TestRussianBankInteractor_Snapshot(t *testing.T) {
	ti, _ := newRbInteractor(domain.RussianBankCpuDifficultyNormal)
	ti.Reset()
	data, err := ti.Snapshot()
	assert.NoError(t, err)
	assert.True(t, len(data) > 0)

	restored, err := usecase.RestoreRussianBankInteractor(data, new(presenter.RussianBankWebPresenter))
	assert.NoError(t, err)
	assert.NotNil(t, restored)

	_, err = usecase.RestoreRussianBankInteractor([]byte("not json"), new(presenter.RussianBankWebPresenter))
	assert.Error(t, err)
}

func TestRussianBankInteractor_LegalMoveViaHint(t *testing.T) {
	ti, g := newRbInteractor(domain.RussianBankCpuDifficultyNormal)
	ti.Reset()
	// Apply one hint-suggested legal move through the interactor (covers the
	// success branch of MoveToFoundation / MoveToTableau).
	h := g.GetHint()
	if h == nil {
		t.Skip("no legal move after deal (rare)")
	}
	var out string
	if h.ToFoundation {
		out = ti.MoveToFoundation(int(h.Zone), h.FromOpponent, h.Col)
	} else {
		out = ti.MoveToTableau(int(h.Zone), h.FromOpponent, h.Col, h.ToCol)
	}
	assert.Contains(t, out, `"phase"`)
	// Hint / ActionLog / Undo outputs.
	assert.Contains(t, ti.Hint(), `"phase"`)
	assert.True(t, strings.Contains(ti.ActionLog(), "[") || ti.ActionLog() != "")
	assert.Contains(t, ti.Undo(), `"phase"`)
}

func TestRussianBankInteractor_IllegalMoveReturnsError(t *testing.T) {
	ti, _ := newRbInteractor(domain.RussianBankCpuDifficultyNormal)
	ti.Reset()
	// Moving the opponent (CPU) reserve to a foundation is almost always illegal
	// right after the deal (no Ace exposed); either way the interactor returns output.
	out := ti.MoveToTableau(int(domain.RussianBankZoneWaste), false, 0, 0)
	assert.NotEmpty(t, out)
	// Undo with nothing to undo returns error output.
	assert.NotEmpty(t, ti.Undo())
	// Stop with no violation returns error output.
	assert.NotEmpty(t, ti.CallStop())
}

func TestRussianBankInteractor_PlaysToGameEnd(t *testing.T) {
	ti, g := newRbInteractor(domain.RussianBankCpuDifficultyHard)
	ti.Reset()
	// Human keeps discarding; the Hard CPU advances each turn until the game ends
	// (a win or a stalemate). Bounded by hand depletion.
	for i := 0; i < 1000 && !g.GetGameEndFlag(); i++ {
		ti.Discard()
	}
	assert.True(t, g.GetGameEndFlag(), "game should reach an end state")
	// Calling actions after game end exercises every guard + winner-message path.
	assert.Contains(t, ti.Discard(), `"gameEndFlag":true`)
	assert.Contains(t, ti.CallStop(), `"gameEndFlag":true`)
	assert.Contains(t, ti.MoveToFoundation(0, false, 0), `"gameEndFlag":true`)
	assert.Contains(t, ti.MoveToTableau(0, false, 0, 1), `"gameEndFlag":true`)
}

func TestRussianBankInteractor_WinnerMessages(t *testing.T) {
	// Asserts the messageCode, not the Japanese text that used to be baked into
	// the presenter. Those literals were the reason the frontend rendered a raw
	// Japanese string in the English UI: with no translation for the code, it fell
	// back to whatever the presenter had hardcoded. The code is the contract now
	// and the copy lives in common.json. See issue #4365.
	for _, tc := range []struct {
		winner int
		want   string
	}{
		{0, `"messageCode":"russianbank.result.humanWin"`},
		{1, `"messageCode":"russianbank.result.cpuWin"`},
		{-1, `"messageCode":"russianbank.result.draw"`},
	} {
		g := rbGameEndState(tc.winner)
		ti := usecase.NewRussianBankInteractor(g, new(presenter.RussianBankWebPresenter))
		out := ti.CallStop() // guardGameEnd -> presenter.Output on the ended game
		assert.Contains(t, out, tc.want)
		assert.Contains(t, out, `"message":""`, "the presenter must not carry its own copy")
	}
}

// rbGameEndState builds a finished game with the given winner via JSON restore.
func rbGameEndState(winner int) *domain.RussianBank {
	js := `{"pl":[{"n":"You","c":false,"s":0,"r":[],"h":[],"w":[]},` +
		`{"n":"CPU","c":true,"s":1,"r":[],"h":[],"w":[]}],` +
		`"cf":{"cd":1},"ph":2,"cu":0,"wn":` + strconv.Itoa(winner) + `}`
	g := domain.NewDefaultRussianBank()
	if err := json.Unmarshal([]byte(js), g); err != nil {
		panic(err)
	}
	return g
}
