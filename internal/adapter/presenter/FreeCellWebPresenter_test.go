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
	fg.On("GetMaxMovableCards").Return(1).Maybe()
	fg.On("GetMaxMovableCardsToEmptyColumn").Return(0).Maybe()

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

// #5975: 空き列へ動かすときの上限はドメインが別に持っており (その空き列自身を
// 経由地に使えないぶん低い)、レスポンスには**どちらの上限も入っていなかった**。
// ページは一般式で計算し直すしかなく、空き列宛ての手をサーバーが弾くまで
// 気づけなかった。
func TestFreeCellWebPresenterCarriesBothMoveLimits(t *testing.T) {
	p := new(FreeCellWebPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)

	var out controller.FreeCellWebOutput
	err := json.Unmarshal([]byte(p.Output(f, nil)), &out)
	assert.NoError(t, err)

	// **ドメインの値をそのまま運ぶ。**presenter で数え直すと、空き列の扱いが
	// 食い違ったときに画面とサーバーで別の答えが出る。
	assert.Equal(t, f.GetMaxMovableCards(), out.MaxMovableCards)
	assert.Equal(t, f.GetMaxMovableCardsToEmptyColumn(), out.MaxMovableCardsToEmptyColumn)
	assert.Positive(t, out.MaxMovableCards, "配り直後は必ず1枚以上動かせる")
}

// 空き列がある局面では、空き列宛ての上限が一般の上限より**低い**。
// 同じ値が返るだけなら、フロントが低い方を使う意味が無い。
func TestFreeCellWebPresenterEmptyColumnLimitIsLower(t *testing.T) {
	p := new(FreeCellWebPresenter)
	f := domain.NewFreeCell(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.FreeCellPhasePlaying)
	// 1 列空ける。
	tableau := f.GetTableau()
	tableau[0] = nil
	f.SetTableau(tableau)

	var out controller.FreeCellWebOutput
	err := json.Unmarshal([]byte(p.Output(f, nil)), &out)
	assert.NoError(t, err)

	assert.Less(t, out.MaxMovableCardsToEmptyColumn, out.MaxMovableCards,
		"空き列自身を経由地に使えないぶん上限は下がる")
}
