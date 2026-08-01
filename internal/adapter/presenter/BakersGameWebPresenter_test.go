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

func TestBakersGameWebPresenterOutputPlaying(t *testing.T) {
	p := new(BakersGameWebPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	result := p.Output(f, nil)

	var out controller.FreeCellWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, int(domain.FreeCellPhasePlaying), out.Phase)
	assert.Equal(t, domain.FreeCellTableauCnt, len(out.Tableau))
	assert.Equal(t, domain.FreeCellCellCnt, len(out.FreeCells))
	assert.Equal(t, domain.FreeCellFoundationCnt, len(out.Foundation))
	assert.Equal(t, "bakersgame.playing", out.MessageCode)
}

func TestBakersGameWebPresenterOutputGameClear(t *testing.T) {
	p := new(BakersGameWebPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhaseGameClear)

	result := p.Output(f, nil)

	var out controller.FreeCellWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, "bakersgame.gameClear", out.MessageCode)
	assert.NotNil(t, out.MessageParams)
}

func TestBakersGameWebPresenterOutputGameOver(t *testing.T) {
	p := new(BakersGameWebPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhaseGameOver)

	result := p.Output(f, nil)

	var out controller.FreeCellWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, "bakersgame.gameOver", out.MessageCode)
}

func TestBakersGameWebPresenterOutputStalemate(t *testing.T) {
	p := new(BakersGameWebPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)
	f.SetIsStalemate(true)

	result := p.Output(f, nil)

	var out controller.FreeCellWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.True(t, out.IsStalemate)
	assert.Equal(t, "bakersgame.stalemate", out.MessageCode)
}

func TestBakersGameWebPresenterOutputStalemateWithEscape(t *testing.T) {
	p := new(BakersGameWebPresenter)
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
	assert.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, 4, out.UndoToEscape)
	assert.Equal(t, "bakersgame.stalemateWithEscape", out.MessageCode)
	assert.Equal(t, "4", out.MessageParams["count"])
}

func TestBakersGameWebPresenterOutputError(t *testing.T) {
	p := new(BakersGameWebPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()

	result := p.Output(f, errors.New("test error"))

	var out controller.FreeCellWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Contains(t, out.Message, "test error")
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
// **フェーズ定数は FreeCell のもの。**Baker's Game は FreeCell の派生で、
// 自分の名前のフェーズ定数を持たない。
func TestBakersGameWebPresenterOutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		f := domain.NewDefaultBakersGame()
		f.Reset()
		f.SetPhase(domain.FreeCellPhasePlaying)

		// **配りに依存させない。**SetTableau で場を丸ごと置き換える。
		var tableau [domain.FreeCellTableauCnt][]*domain.Card
		tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
		f.SetTableau(tableau)

		var out controller.FreeCellWebOutput
		assert.NoError(t, json.Unmarshal([]byte(new(BakersGameWebPresenter).Output(f, nil)), &out))
		assert.NotNil(t, out.Hint, "Output must carry the hint -- the frontend reads state.hint")
	})

	t.Run("not while cleared", func(t *testing.T) {
		f := domain.NewDefaultBakersGame()
		f.Reset()
		f.SetPhase(domain.FreeCellPhaseGameClear)

		var out controller.FreeCellWebOutput
		assert.NoError(t, json.Unmarshal([]byte(new(BakersGameWebPresenter).Output(f, nil)), &out))
		assert.Nil(t, out.Hint)
	})
}

func TestBakersGameWebPresenterHintOutputWithHint(t *testing.T) {
	p := new(BakersGameWebPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	var tableau [domain.FreeCellTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	f.SetTableau(tableau)

	result := p.HintOutput(f)

	var out controller.FreeCellWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.NotNil(t, out.Hint)
	assert.Equal(t, "bakersgame.hintAvailable", out.MessageCode)
}

func TestBakersGameWebPresenterHintOutputNoHint(t *testing.T) {
	p := new(BakersGameWebPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhaseGameOver)

	result := p.HintOutput(f)

	var out controller.FreeCellWebOutput
	assert.NoError(t, json.Unmarshal([]byte(result), &out))
	assert.Equal(t, "bakersgame.noHint", out.MessageCode)
}

func TestBakersGameWebPresenterActionLog(t *testing.T) {
	p := new(BakersGameWebPresenter)
	f := domain.NewDefaultBakersGame()
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	var playing controller.ActionLogWebOutput
	assert.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(f)), &playing))
	assert.Empty(t, playing.Entries)

	var tableau [domain.FreeCellTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)}
	f.SetTableau(tableau)
	_ = f.MoveTableauToFoundation(0)
	f.SetPhase(domain.FreeCellPhaseGameOver)

	var over controller.ActionLogWebOutput
	assert.NoError(t, json.Unmarshal([]byte(p.ActionLogOutput(f)), &over))
	assert.NotEmpty(t, over.Entries)
}
