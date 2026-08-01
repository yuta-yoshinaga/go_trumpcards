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

func TestFreeCellWebPresenterOutputPlaying(t *testing.T) {
	p := new(FreeCellWebPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	result := p.Output(f, nil)

	var out controller.FreeCellWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, int(domain.FreeCellPhasePlaying), out.Phase)
	assert.Equal(t, domain.FreeCellTableauCnt, len(out.Tableau))
	assert.Equal(t, domain.FreeCellCellCnt, len(out.FreeCells))
	assert.Equal(t, domain.FreeCellFoundationCnt, len(out.Foundation))
	assert.Equal(t, 0, out.MoveCount)
	assert.Equal(t, "freecell.playing", out.MessageCode)
}

func TestFreeCellWebPresenterOutputGameClear(t *testing.T) {
	p := new(FreeCellWebPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhaseGameClear)

	result := p.Output(f, nil)

	var out controller.FreeCellWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, "freecell.gameClear", out.MessageCode)
	assert.NotNil(t, out.MessageParams)
}

func TestFreeCellWebPresenterOutputGameOver(t *testing.T) {
	p := new(FreeCellWebPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhaseGameOver)

	result := p.Output(f, nil)

	var out controller.FreeCellWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, "freecell.gameOver", out.MessageCode)
}

func TestFreeCellWebPresenterOutputStalemate(t *testing.T) {
	p := new(FreeCellWebPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)
	f.SetIsStalemate(true)

	result := p.Output(f, nil)

	var out controller.FreeCellWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.True(t, out.IsStalemate)
	assert.Equal(t, -1, out.UndoToEscape)
	assert.Equal(t, "freecell.stalemate", out.MessageCode)
	assert.Empty(t, out.MessageParams)
}

func TestFreeCellWebPresenterOutputStalemateWithEscape(t *testing.T) {
	p := new(FreeCellWebPresenter)
	fg := new(interfaces.MockFreeCellGame)
	fg.On("GetPhase").Return(domain.FreeCellPhasePlaying).Maybe()
	fg.On("GetMoveCount").Return(7).Maybe()
	fg.On("CanUndo").Return(true).Maybe()
	fg.On("IsStalemate").Return(true).Maybe()
	fg.On("UndoToEscape").Return(4).Maybe()
	var tableau [domain.FreeCellTableauCnt][]*domain.Card
	fg.On("GetTableau").Return(tableau).Maybe()
	var freeCells [domain.FreeCellCellCnt]*domain.Card
	fg.On("GetFreeCells").Return(freeCells).Maybe()
	var foundation [domain.FreeCellFoundationCnt][]*domain.Card
	fg.On("GetFoundation").Return(foundation).Maybe()

	result := p.Output(fg, nil)

	var out controller.FreeCellWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.True(t, out.IsStalemate)
	assert.Equal(t, 4, out.UndoToEscape)
	assert.Equal(t, "freecell.stalemateWithEscape", out.MessageCode)
	assert.Equal(t, "4", out.MessageParams["count"])
}

func TestFreeCellWebPresenterOutputError(t *testing.T) {
	p := new(FreeCellWebPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()

	result := p.Output(f, errors.New("test error"))

	var out controller.FreeCellWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Contains(t, out.Message, "test error")
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestFreeCellWebPresenterOutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		f := domain.NewFreeCell(domain.NewTrumpCards(0))
		f.Reset()
		f.SetPhase(domain.FreeCellPhasePlaying)

		// **配りに依存させない。**SetTableau で場を丸ごと置き換えるので、
		// エースが 1 枚だけ = 台札へ動かせる = ヒントが必ず出る。
		var tableau [domain.FreeCellTableauCnt][]*domain.Card
		tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		f.SetTableau(tableau)

		var out controller.FreeCellWebOutput
		assert.NoError(t, json.Unmarshal([]byte(new(FreeCellWebPresenter).Output(f, nil)), &out))
		assert.NotNil(t, out.Hint, "Output must carry the hint -- the frontend reads state.hint")
	})

	t.Run("not while cleared", func(t *testing.T) {
		f := domain.NewFreeCell(domain.NewTrumpCards(0))
		f.Reset()
		f.SetPhase(domain.FreeCellPhaseGameClear)

		var out controller.FreeCellWebOutput
		assert.NoError(t, json.Unmarshal([]byte(new(FreeCellWebPresenter).Output(f, nil)), &out))
		assert.Nil(t, out.Hint)
	})
}

func TestFreeCellWebPresenterHintOutputWithHint(t *testing.T) {
	p := new(FreeCellWebPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	// Place an Ace on a tableau column so hint suggests moving it to foundation
	var tableau [domain.FreeCellTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	f.SetTableau(tableau)

	result := p.HintOutput(f)

	var out controller.FreeCellWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.NotNil(t, out.Hint)
	assert.Equal(t, "freecell.hintAvailable", out.MessageCode)
}

func TestFreeCellWebPresenterHintOutputNoHint(t *testing.T) {
	p := new(FreeCellWebPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhaseGameOver)

	result := p.HintOutput(f)

	var out controller.FreeCellWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, "freecell.noHint", out.MessageCode)
}

func TestFreeCellWebPresenterActionLogPlaying(t *testing.T) {
	p := new(FreeCellWebPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	result := p.ActionLogOutput(f)

	var out controller.ActionLogWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Empty(t, out.Entries)
}

func TestFreeCellWebPresenterActionLogGameOver(t *testing.T) {
	p := new(FreeCellWebPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()

	// Make a move to generate action log
	var tableau [domain.FreeCellTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	f.SetTableau(tableau)
	f.SetPhase(domain.FreeCellPhasePlaying)
	_ = f.MoveTableauToFoundation(0)
	f.SetPhase(domain.FreeCellPhaseGameOver)

	result := p.ActionLogOutput(f)

	var out controller.ActionLogWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.NotEmpty(t, out.Entries)
}
