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

func TestStalactitesWebPresenterOutputPlaying(t *testing.T) {
	p := new(StalactitesWebPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)

	result := p.Output(f, nil)

	var out controller.StalactitesWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, int(domain.StalactitesPhasePlaying), out.Phase)
	assert.Equal(t, domain.StalactitesTableauCnt, len(out.Tableau))
	assert.Equal(t, domain.StalactitesCellCnt, len(out.Cells))
	assert.Equal(t, domain.StalactitesFoundationCnt, len(out.Foundation))
	assert.Equal(t, 0, out.MoveCount)
	assert.Equal(t, "stalactites.playing", out.MessageCode)
}

func TestStalactitesWebPresenterOutputGameClear(t *testing.T) {
	p := new(StalactitesWebPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhaseGameClear)

	result := p.Output(f, nil)

	var out controller.StalactitesWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, "stalactites.gameClear", out.MessageCode)
	assert.NotNil(t, out.MessageParams)
}

func TestStalactitesWebPresenterOutputGameOver(t *testing.T) {
	p := new(StalactitesWebPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhaseGameOver)

	result := p.Output(f, nil)

	var out controller.StalactitesWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, "stalactites.gameOver", out.MessageCode)
}

func TestStalactitesWebPresenterOutputStalemate(t *testing.T) {
	p := new(StalactitesWebPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)
	f.SetIsStalemate(true)

	result := p.Output(f, nil)

	var out controller.StalactitesWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.True(t, out.IsStalemate)
	assert.Equal(t, -1, out.UndoToEscape)
	assert.Equal(t, "stalactites.stalemate", out.MessageCode)
	assert.Empty(t, out.MessageParams)
}

func TestStalactitesWebPresenterOutputStalemateWithEscape(t *testing.T) {
	p := new(StalactitesWebPresenter)
	fg := new(interfaces.MockStalactitesGame)
	fg.On("GetPhase").Return(domain.StalactitesPhasePlaying).Maybe()
	fg.On("GetMoveCount").Return(7).Maybe()
	fg.On("CanUndo").Return(true).Maybe()
	fg.On("IsStalemate").Return(true).Maybe()
	fg.On("UndoToEscape").Return(4).Maybe()
	var tableau [domain.StalactitesTableauCnt][]*domain.Card
	fg.On("GetTableau").Return(tableau).Maybe()
	var cells [domain.StalactitesCellCnt]*domain.Card
	fg.On("GetCells").Return(cells).Maybe()
	var foundation [domain.StalactitesFoundationCnt][]*domain.Card
	fg.On("GetFoundation").Return(foundation).Maybe()
	fg.On("GetMaxMovableCards").Return(1).Maybe()
	fg.On("GetBaseRank").Return(1).Maybe()
	fg.On("GetMaxMovableCardsToEmptyColumn").Return(0).Maybe()

	result := p.Output(fg, nil)

	var out controller.StalactitesWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.True(t, out.IsStalemate)
	assert.Equal(t, 4, out.UndoToEscape)
	assert.Equal(t, "stalactites.stalemateWithEscape", out.MessageCode)
	assert.Equal(t, "4", out.MessageParams["count"])
}

func TestStalactitesWebPresenterOutputError(t *testing.T) {
	p := new(StalactitesWebPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()

	result := p.Output(f, errors.New("test error"))

	var out controller.StalactitesWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Contains(t, out.Message, "test error")
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestStalactitesWebPresenterOutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		f := domain.NewStalactites(domain.NewTrumpCards(0))
		f.Reset()
		f.SetPhase(domain.StalactitesPhasePlaying)

		// **配りに依存させない。**SetTableau で場を丸ごと置き換えるので、
		// エースが 1 枚だけ = 台札へ動かせる = ヒントが必ず出る。
		var tableau [domain.StalactitesTableauCnt][]*domain.Card
		tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, f.GetBaseRank(), false)}
		f.SetTableau(tableau)

		var out controller.StalactitesWebOutput
		assert.NoError(t, json.Unmarshal([]byte(new(StalactitesWebPresenter).Output(f, nil)), &out))
		assert.NotNil(t, out.Hint, "Output must carry the hint -- the frontend reads state.hint")
	})

	t.Run("not while cleared", func(t *testing.T) {
		f := domain.NewStalactites(domain.NewTrumpCards(0))
		f.Reset()
		f.SetPhase(domain.StalactitesPhaseGameClear)

		var out controller.StalactitesWebOutput
		assert.NoError(t, json.Unmarshal([]byte(new(StalactitesWebPresenter).Output(f, nil)), &out))
		assert.Nil(t, out.Hint)
	})
}

func TestStalactitesWebPresenterHintOutputWithHint(t *testing.T) {
	p := new(StalactitesWebPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)

	// Place an Ace on a tableau column so hint suggests moving it to foundation
	var tableau [domain.StalactitesTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, f.GetBaseRank(), false)}
	f.SetTableau(tableau)

	result := p.HintOutput(f)

	var out controller.StalactitesWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.NotNil(t, out.Hint)
	assert.Equal(t, "stalactites.hintAvailable", out.MessageCode)
}

func TestStalactitesWebPresenterHintOutputNoHint(t *testing.T) {
	p := new(StalactitesWebPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhaseGameOver)

	result := p.HintOutput(f)

	var out controller.StalactitesWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Equal(t, "stalactites.noHint", out.MessageCode)
}

func TestStalactitesWebPresenterActionLogPlaying(t *testing.T) {
	p := new(StalactitesWebPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)

	result := p.ActionLogOutput(f)

	var out controller.ActionLogWebOutput
	err := json.Unmarshal([]byte(result), &out)
	assert.NoError(t, err)
	assert.Empty(t, out.Entries)
}

func TestStalactitesWebPresenterActionLogGameOver(t *testing.T) {
	p := new(StalactitesWebPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()

	// Make a move to generate action log
	var tableau [domain.StalactitesTableauCnt][]*domain.Card
	tableau[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, f.GetBaseRank(), false)}
	f.SetTableau(tableau)
	f.SetPhase(domain.StalactitesPhasePlaying)
	_ = f.MoveTableauToFoundation(0)
	f.SetPhase(domain.StalactitesPhaseGameOver)

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
func TestStalactitesWebPresenterCarriesBothMoveLimits(t *testing.T) {
	p := new(StalactitesWebPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)

	var out controller.StalactitesWebOutput
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
func TestStalactitesWebPresenterEmptyColumnLimitIsLower(t *testing.T) {
	p := new(StalactitesWebPresenter)
	f := domain.NewStalactites(domain.NewTrumpCards(0))
	f.Reset()
	f.SetPhase(domain.StalactitesPhasePlaying)
	// 1 列空ける。
	tableau := f.GetTableau()
	tableau[0] = nil
	f.SetTableau(tableau)

	var out controller.StalactitesWebOutput
	err := json.Unmarshal([]byte(p.Output(f, nil)), &out)
	assert.NoError(t, err)

	assert.Less(t, out.MaxMovableCardsToEmptyColumn, out.MaxMovableCards,
		"空き列自身を経由地に使えないぶん上限は下がる")
}
