//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func TestPenguinWebPresenterOutputPlaying(t *testing.T) {
	p := new(PenguinWebPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhasePlaying)

	result := p.Output(g, nil)

	var out controller.PenguinWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, int(domain.PenguinPhasePlaying), out.Phase)
	assert.Equal(t, domain.PenguinTableauCnt, len(out.Tableau))
	assert.Equal(t, domain.PenguinCellCnt, len(out.FreeCells))
	assert.Equal(t, domain.PenguinFoundationCnt, len(out.Foundation))
	assert.Equal(t, 0, out.MoveCount)
	assert.Equal(t, "penguin.playing", out.MessageCode)
}

func TestPenguinWebPresenterOutputGameClear(t *testing.T) {
	p := new(PenguinWebPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhaseGameClear)

	result := p.Output(g, nil)

	var out controller.PenguinWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, "penguin.gameClear", out.MessageCode)
	assert.NotNil(t, out.MessageParams)
}

func TestPenguinWebPresenterOutputGameOver(t *testing.T) {
	p := new(PenguinWebPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhaseGameOver)

	result := p.Output(g, nil)

	var out controller.PenguinWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, "penguin.gameOver", out.MessageCode)
}

func TestPenguinWebPresenterOutputStalemate(t *testing.T) {
	p := new(PenguinWebPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhasePlaying)
	g.SetIsStalemate(true)

	result := p.Output(g, nil)

	var out controller.PenguinWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.True(t, out.IsStalemate)
	assert.Equal(t, -1, out.UndoToEscape)
	assert.Equal(t, "penguin.stalemate", out.MessageCode)
	assert.Empty(t, out.MessageParams)
}

func TestPenguinWebPresenterOutputStalemateWithEscape(t *testing.T) {
	p := new(PenguinWebPresenter)
	pg := new(interfaces.MockPenguinGame)
	pg.On("GetPhase").Return(domain.PenguinPhasePlaying).Maybe()
	pg.On("GetMoveCount").Return(7).Maybe()
	pg.On("CanUndo").Return(true).Maybe()
	pg.On("IsStalemate").Return(true).Maybe()
	pg.On("UndoToEscape").Return(4).Maybe()
	var tableau [domain.PenguinTableauCnt][]*domain.Card
	pg.On("GetTableau").Return(tableau).Maybe()
	var freeCells [domain.PenguinCellCnt]*domain.Card
	pg.On("GetFreeCells").Return(freeCells).Maybe()
	var foundation [domain.PenguinFoundationCnt][]*domain.Card
	pg.On("GetFoundation").Return(foundation).Maybe()
	pg.On("GetBaseRank").Return(1).Maybe()
	pg.On("GetActionLog").Return(nil).Maybe()
	pg.On("GetGameEndFlag").Return(false).Maybe()

	result := p.Output(pg, nil)

	var out controller.PenguinWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.True(t, out.IsStalemate)
	assert.Equal(t, 4, out.UndoToEscape)
	assert.Equal(t, "penguin.stalemateWithEscape", out.MessageCode)
	assert.Equal(t, "4", out.MessageParams["count"])
}

func TestPenguinWebPresenterOutputError(t *testing.T) {
	p := new(PenguinWebPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()

	result := p.Output(g, errors.New("test error"))

	var out controller.PenguinWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Contains(t, out.Message, "test error")
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestPenguinWebPresenterOutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := domain.NewPenguin(domain.NewTrumpCards(0))
		g.Reset()
		g.SetPhase(domain.PenguinPhasePlaying)

		// **配りに依存させない。**SetTableau で場を丸ごと置き換えるので、
		// 動かせる札が 1 枚だけ残り、ヒントが必ず出る。
		baseRank := g.GetBaseRank()
		var tableau [domain.PenguinTableauCnt][]*domain.Card
		tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, baseRank, false)}
		g.SetTableau(tableau)

		var out controller.PenguinWebOutput
		assert.NoError(t, json.Unmarshal([]byte(new(PenguinWebPresenter).Output(g, nil)), &out))
		assert.NotNil(t, out.Hint, "Output must carry the hint -- the frontend reads state.hint")
	})

	t.Run("not while cleared", func(t *testing.T) {
		g := domain.NewPenguin(domain.NewTrumpCards(0))
		g.Reset()
		g.SetPhase(domain.PenguinPhaseGameClear)

		var out controller.PenguinWebOutput
		assert.NoError(t, json.Unmarshal([]byte(new(PenguinWebPresenter).Output(g, nil)), &out))
		assert.Nil(t, out.Hint)
	})
}

func TestPenguinWebPresenterHintOutputWithHint(t *testing.T) {
	p := new(PenguinWebPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhasePlaying)

	// Place a card in tableau that can move to foundation (baseRank of the game)
	// After Reset, freeCells[0..2] hold base-rank cards. We set tableau[0] to the same base rank
	// to give GetHint something to work with.
	baseRank := g.GetBaseRank()
	var tableau [domain.PenguinTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, baseRank, false)}
	g.SetTableau(tableau)

	result := p.HintOutput(g)

	var out controller.PenguinWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.NotNil(t, out.Hint)
	assert.Equal(t, "penguin.hintAvailable", out.MessageCode)
}

func TestPenguinWebPresenterHintOutputNoHint(t *testing.T) {
	p := new(PenguinWebPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhaseGameOver)

	result := p.HintOutput(g)

	var out controller.PenguinWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, "penguin.noHint", out.MessageCode)
}

func TestPenguinWebPresenterActionLogPlaying(t *testing.T) {
	p := new(PenguinWebPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()
	g.SetPhase(domain.PenguinPhasePlaying)

	result := p.ActionLogOutput(g)

	var out controller.ActionLogWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Empty(t, out.Entries)
}

func TestPenguinWebPresenterActionLogGameOver(t *testing.T) {
	p := new(PenguinWebPresenter)
	g := domain.NewPenguin(domain.NewTrumpCards(0))
	g.Reset()

	// Move a card to foundation to generate an action log entry
	var tableau [domain.PenguinTableauCnt][]*domain.Card
	baseRank := g.GetBaseRank()
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, baseRank, false)}
	g.SetTableau(tableau)
	g.SetPhase(domain.PenguinPhasePlaying)
	_ = g.MoveTableauToFoundation(0)
	g.SetPhase(domain.PenguinPhaseGameOver)

	result := p.ActionLogOutput(g)

	var out controller.ActionLogWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.NotEmpty(t, out.Entries)
}
