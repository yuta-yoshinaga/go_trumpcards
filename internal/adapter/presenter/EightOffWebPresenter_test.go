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

func TestEightOffWebPresenterOutputPlaying(t *testing.T) {
	p := new(EightOffWebPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhasePlaying)

	result := p.Output(e, nil)

	var out controller.EightOffWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, int(domain.EightOffPhasePlaying), out.Phase)
	assert.Equal(t, domain.EightOffTableauCnt, len(out.Tableau))
	assert.Equal(t, domain.EightOffCellCnt, len(out.FreeCells))
	assert.Equal(t, domain.EightOffFoundationCnt, len(out.Foundation))
	assert.Equal(t, 0, out.MoveCount)
	assert.Equal(t, "eightoff.playing", out.MessageCode)
}

func TestEightOffWebPresenterOutputGameClear(t *testing.T) {
	p := new(EightOffWebPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhaseGameClear)

	result := p.Output(e, nil)

	var out controller.EightOffWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, "eightoff.gameClear", out.MessageCode)
	assert.NotNil(t, out.MessageParams)
}

func TestEightOffWebPresenterOutputGameOver(t *testing.T) {
	p := new(EightOffWebPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhaseGameOver)

	result := p.Output(e, nil)

	var out controller.EightOffWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, "eightoff.gameOver", out.MessageCode)
}

func TestEightOffWebPresenterOutputStalemate(t *testing.T) {
	p := new(EightOffWebPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhasePlaying)
	e.SetIsStalemate(true)

	result := p.Output(e, nil)

	var out controller.EightOffWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.True(t, out.IsStalemate)
	assert.Equal(t, -1, out.UndoToEscape)
	assert.Equal(t, "eightoff.stalemate", out.MessageCode)
	assert.Empty(t, out.MessageParams)
}

func TestEightOffWebPresenterOutputStalemateWithEscape(t *testing.T) {
	p := new(EightOffWebPresenter)
	eg := new(interfaces.MockEightOffGame)
	eg.On("GetPhase").Return(domain.EightOffPhasePlaying).Maybe()
	eg.On("GetMoveCount").Return(7).Maybe()
	eg.On("CanUndo").Return(true).Maybe()
	eg.On("IsStalemate").Return(true).Maybe()
	eg.On("UndoToEscape").Return(4).Maybe()
	var tableau [domain.EightOffTableauCnt][]*domain.Card
	eg.On("GetTableau").Return(tableau).Maybe()
	var freeCells [domain.EightOffCellCnt]*domain.Card
	eg.On("GetFreeCells").Return(freeCells).Maybe()
	var foundation [domain.EightOffFoundationCnt][]*domain.Card
	eg.On("GetFoundation").Return(foundation).Maybe()

	result := p.Output(eg, nil)

	var out controller.EightOffWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.True(t, out.IsStalemate)
	assert.Equal(t, 4, out.UndoToEscape)
	assert.Equal(t, "eightoff.stalemateWithEscape", out.MessageCode)
	assert.Equal(t, "4", out.MessageParams["count"])
}

func TestEightOffWebPresenterOutputError(t *testing.T) {
	p := new(EightOffWebPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()

	result := p.Output(e, errors.New("test error"))

	var out controller.EightOffWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Contains(t, out.Message, "test error")
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestEightOffWebPresenterOutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		e := domain.NewEightOff(domain.NewTrumpCards(0))
		e.Reset()
		e.SetPhase(domain.EightOffPhasePlaying)

		// **配りに依存させない。**SetTableau で場を丸ごと置き換えるので、
		// 動かせる札が 1 枚だけ残り、ヒントが必ず出る。
		var tableau [domain.EightOffTableauCnt][]*domain.Card
		tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		e.SetTableau(tableau)

		var out controller.EightOffWebOutput
		assert.NoError(t, json.Unmarshal([]byte(new(EightOffWebPresenter).Output(e, nil)), &out))
		assert.NotNil(t, out.Hint, "Output must carry the hint -- the frontend reads state.hint")
	})

	t.Run("not while cleared", func(t *testing.T) {
		e := domain.NewEightOff(domain.NewTrumpCards(0))
		e.Reset()
		e.SetPhase(domain.EightOffPhaseGameClear)

		var out controller.EightOffWebOutput
		assert.NoError(t, json.Unmarshal([]byte(new(EightOffWebPresenter).Output(e, nil)), &out))
		assert.Nil(t, out.Hint)
	})
}

func TestEightOffWebPresenterHintOutputWithHint(t *testing.T) {
	p := new(EightOffWebPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhasePlaying)

	var tableau [domain.EightOffTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	e.SetTableau(tableau)

	result := p.HintOutput(e)

	var out controller.EightOffWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.NotNil(t, out.Hint)
	assert.Equal(t, "eightoff.hintAvailable", out.MessageCode)
}

func TestEightOffWebPresenterHintOutputNoHint(t *testing.T) {
	p := new(EightOffWebPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhaseGameOver)

	result := p.HintOutput(e)

	var out controller.EightOffWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, "eightoff.noHint", out.MessageCode)
}

func TestEightOffWebPresenterActionLogPlaying(t *testing.T) {
	p := new(EightOffWebPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()
	e.SetPhase(domain.EightOffPhasePlaying)

	result := p.ActionLogOutput(e)

	var out controller.ActionLogWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Empty(t, out.Entries)
}

func TestEightOffWebPresenterActionLogGameOver(t *testing.T) {
	p := new(EightOffWebPresenter)
	e := domain.NewEightOff(domain.NewTrumpCards(0))
	e.Reset()

	var tableau [domain.EightOffTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	e.SetTableau(tableau)
	e.SetPhase(domain.EightOffPhasePlaying)
	_ = e.MoveTableauToFoundation(0)
	e.SetPhase(domain.EightOffPhaseGameOver)

	result := p.ActionLogOutput(e)

	var out controller.ActionLogWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.NotEmpty(t, out.Entries)
}
