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

func sampleAcesUpColumns() [domain.AcesUpColCnt][]*domain.Card {
	var cols [domain.AcesUpColCnt][]*domain.Card
	cols[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)}
	cols[1] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)}
	cols[2] = nil
	cols[3] = []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 2, false),
		domain.NewCard(domain.CardDesignDiamond, 6, false),
	}
	return cols
}

func setupAcesUpWebMockDefaults(g *interfaces.MockAcesUpGame) {
	g.On("GetPhase").Return(domain.AcesUpPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("GetStockCount").Return(44).Maybe()
	g.On("GetDiscardCount").Return(4).Maybe()
	g.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignClover, 7, false)).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetColumns").Return(sampleAcesUpColumns()).Maybe()
	for c := range domain.AcesUpColCnt {
		g.On("CanRemove", c).Return(c == 0).Maybe()
		g.On("CanMove", c).Return(true).Maybe()
	}
}

func parseAcesUpOutput(t *testing.T, jsonStr string) *controller.AcesUpWebOutput {
	t.Helper()
	var out controller.AcesUpWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupAcesUpOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupAcesUpOutputMock(g *interfaces.MockAcesUpGame) {
	setupAcesUpWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestAcesUpWebPresenterOutput_Playing(t *testing.T) {
	g := new(interfaces.MockAcesUpGame)
	setupAcesUpOutputMock(g)
	p := &AcesUpWebPresenter{}

	out := parseAcesUpOutput(t, p.Output(g, nil))
	assert.Equal(t, 0, out.Phase)
	assert.Equal(t, 44, out.StockCount)
	assert.Equal(t, 4, out.DiscardCount)
	// The discard pile top card is surfaced for the on-board discard display.
	assert.NotNil(t, out.DiscardTop)
	assert.Equal(t, "CLOVER", out.DiscardTop.Design)
	assert.Equal(t, 7, out.DiscardTop.Value)
	assert.Equal(t, "acesup.playing", out.MessageCode)
	assert.Len(t, out.Columns, domain.AcesUpColCnt)
	// col0 top is removable
	assert.True(t, out.Columns[0][0].Top)
	assert.True(t, out.Columns[0][0].Removable)
	// empty column
	assert.Len(t, out.Columns[2], 0)
	// col3 has two cards, only the last is top
	assert.False(t, out.Columns[3][0].Top)
	assert.True(t, out.Columns[3][1].Top)
}

func TestAcesUpWebPresenterOutput_EmptyDiscard(t *testing.T) {
	g := new(interfaces.MockAcesUpGame)
	setupAcesUpOutputMock(g)
	// Override the discard top to nil (no cards removed yet).
	g.ExpectedCalls = nil
	g.On("GetPhase").Return(domain.AcesUpPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("GetStockCount").Return(48).Maybe()
	g.On("GetDiscardCount").Return(0).Maybe()
	g.On("GetDiscardTop").Return((*domain.Card)(nil)).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetColumns").Return(sampleAcesUpColumns()).Maybe()
	for c := range domain.AcesUpColCnt {
		g.On("CanRemove", c).Return(false).Maybe()
		g.On("CanMove", c).Return(false).Maybe()
	}
	g.On("GetHint").Return(nil).Maybe()
	p := &AcesUpWebPresenter{}

	out := parseAcesUpOutput(t, p.Output(g, nil))
	// With an empty discard pile, the omitempty field is absent.
	assert.Nil(t, out.DiscardTop)
	assert.Equal(t, 0, out.DiscardCount)
}

func TestAcesUpWebPresenterOutput_Error(t *testing.T) {
	g := new(interfaces.MockAcesUpGame)
	setupAcesUpOutputMock(g)
	p := &AcesUpWebPresenter{}

	out := parseAcesUpOutput(t, p.Output(g, errors.New("test error")))
	assert.Equal(t, "test error", out.Message)
}

func TestAcesUpWebPresenterOutput_Stalemate(t *testing.T) {
	g := new(interfaces.MockAcesUpGame)
	g.On("GetPhase").Return(domain.AcesUpPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(5).Maybe()
	g.On("GetStockCount").Return(0).Maybe()
	g.On("GetDiscardCount").Return(10).Maybe()
	g.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 3, false)).Maybe()
	g.On("CanUndo").Return(true).Maybe()
	g.On("IsStalemate").Return(true).Maybe()
	g.On("UndoToEscape").Return(-1).Maybe()
	var cols [domain.AcesUpColCnt][]*domain.Card
	g.On("GetColumns").Return(cols).Maybe()
	p := &AcesUpWebPresenter{}

	out := parseAcesUpOutput(t, p.Output(g, nil))
	assert.Equal(t, "acesup.stalemate", out.MessageCode)
}

func TestAcesUpWebPresenterOutput_GameClear(t *testing.T) {
	g := new(interfaces.MockAcesUpGame)
	g.On("GetPhase").Return(domain.AcesUpPhaseGameClear).Maybe()
	g.On("GetMoveCount").Return(20).Maybe()
	g.On("GetStockCount").Return(0).Maybe()
	g.On("GetDiscardCount").Return(48).Maybe()
	g.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignClover, 5, false)).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	var cols [domain.AcesUpColCnt][]*domain.Card
	g.On("GetColumns").Return(cols).Maybe()
	p := &AcesUpWebPresenter{}

	out := parseAcesUpOutput(t, p.Output(g, nil))
	assert.Equal(t, "acesup.gameClear", out.MessageCode)
	assert.Contains(t, out.Message, "20")
}

func TestAcesUpWebPresenterOutput_GameOver(t *testing.T) {
	g := new(interfaces.MockAcesUpGame)
	g.On("GetPhase").Return(domain.AcesUpPhaseGameOver).Maybe()
	g.On("GetMoveCount").Return(5).Maybe()
	g.On("GetStockCount").Return(0).Maybe()
	g.On("GetDiscardCount").Return(2).Maybe()
	g.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignSpade, 4, false)).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	var cols [domain.AcesUpColCnt][]*domain.Card
	g.On("GetColumns").Return(cols).Maybe()
	p := &AcesUpWebPresenter{}

	out := parseAcesUpOutput(t, p.Output(g, nil))
	assert.Equal(t, "acesup.gameOver", out.MessageCode)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestAcesUpWebPresenterOutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		aug := new(interfaces.MockAcesUpGame)
		setupAcesUpWebMockDefaults(aug)
		aug.On("GetHint").Return(&domain.AcesUpHint{Type: "remove", Col: 1}).Maybe()

		result := new(AcesUpWebPresenter).Output(aug, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		aug := new(interfaces.MockAcesUpGame)
		setupAcesUpWebMockDefaults(aug)
		aug.ExpectedCalls = filterCalls(aug.ExpectedCalls, "IsStalemate")
		aug.On("IsStalemate").Return(true)
		aug.On("GetHint").Return(&domain.AcesUpHint{Type: "remove", Col: 1}).Maybe()

		result := new(AcesUpWebPresenter).Output(aug, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestAcesUpWebPresenterHintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockAcesUpGame)
		g.On("GetHint").Return(&domain.AcesUpHint{Type: "move", Col: 2})
		g.On("GetPhase").Return(domain.AcesUpPhasePlaying)
		g.On("GetMoveCount").Return(0)
		g.On("GetStockCount").Return(0)
		g.On("GetDiscardCount").Return(4)
		g.On("CanUndo").Return(false)
		g.On("IsStalemate").Return(false)
		g.On("UndoToEscape").Return(0)

		p := &AcesUpWebPresenter{}
		out := parseAcesUpOutput(t, p.HintOutput(g))
		assert.NotNil(t, out.Hint)
		assert.Equal(t, "move", out.Hint.Type)
		assert.Equal(t, "acesup.hintAvailable", out.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockAcesUpGame)
		g.On("GetHint").Return((*domain.AcesUpHint)(nil))
		g.On("GetPhase").Return(domain.AcesUpPhasePlaying)
		g.On("GetMoveCount").Return(0)
		g.On("GetStockCount").Return(0)
		g.On("GetDiscardCount").Return(4)
		g.On("CanUndo").Return(false)
		g.On("IsStalemate").Return(false)
		g.On("UndoToEscape").Return(0)

		p := &AcesUpWebPresenter{}
		out := parseAcesUpOutput(t, p.HintOutput(g))
		assert.Nil(t, out.Hint)
		assert.Equal(t, "acesup.noHint", out.MessageCode)
	})
}

func TestAcesUpWebPresenterActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockAcesUpGame)
		g.On("GetPhase").Return(domain.AcesUpPhasePlaying)
		g.On("GetGameEndFlag").Return(false)
		p := &AcesUpWebPresenter{}
		assert.Contains(t, p.ActionLogOutput(g), "entries")
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockAcesUpGame)
		g.On("GetPhase").Return(domain.AcesUpPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
		p := &AcesUpWebPresenter{}
		assert.Contains(t, p.ActionLogOutput(g), "entries")
	})
}
