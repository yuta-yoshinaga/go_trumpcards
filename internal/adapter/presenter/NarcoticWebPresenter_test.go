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

func sampleNarcoticColumns() [domain.NarcoticColCnt][]*domain.Card {
	var cols [domain.NarcoticColCnt][]*domain.Card
	cols[0] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)}
	cols[1] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, 9, false)}
	cols[2] = nil
	cols[3] = []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 2, false),
		domain.NewCard(domain.CardDesignDiamond, 6, false),
	}
	return cols
}

func setupNarcoticWebMockDefaults(g *interfaces.MockNarcoticGame) {
	g.On("GetPhase").Return(domain.NarcoticPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("GetStockCount").Return(44).Maybe()
	g.On("GetDiscardCount").Return(4).Maybe()
	g.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignClover, 7, false)).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetColumns").Return(sampleNarcoticColumns()).Maybe()
	// **Removable は盤面全体の性質。**4枚揃ったときだけ真で、そのときは全列が真。
	g.On("CanRemoveSet").Return(true).Maybe()
	g.On("GetRedealCount").Return(0).Maybe()
	for c := range domain.NarcoticColCnt {
		g.On("CanMove", c).Return(true).Maybe()
	}
}

func parseNarcoticOutput(t *testing.T, jsonStr string) *controller.NarcoticWebOutput {
	t.Helper()
	var out controller.NarcoticWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupNarcoticOutputMock は Output 用の既定。**Output() も受動ヒントを埋める**ように
// なった (#4483) ので GetHint を呼べるようにする。共有ヘルパーに置くと、先に
// 登録されたこの期待が HintOutput テストの「ヒントあり」を食う。
func setupNarcoticOutputMock(g *interfaces.MockNarcoticGame) {
	setupNarcoticWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestNarcoticWebPresenterOutput_Playing(t *testing.T) {
	g := new(interfaces.MockNarcoticGame)
	setupNarcoticOutputMock(g)
	p := &NarcoticWebPresenter{}

	out := parseNarcoticOutput(t, p.Output(g, nil))
	assert.Equal(t, 0, out.Phase)
	assert.Equal(t, 44, out.StockCount)
	assert.Equal(t, 4, out.DiscardCount)
	// The discard pile top card is surfaced for the on-board discard display.
	assert.NotNil(t, out.DiscardTop)
	assert.Equal(t, "CLOVER", out.DiscardTop.Design)
	assert.Equal(t, 7, out.DiscardTop.Value)
	assert.Equal(t, "narcotic.playing", out.MessageCode)
	assert.Len(t, out.Columns, domain.NarcoticColCnt)
	// col0 top is removable
	assert.True(t, out.Columns[0][0].Top)
	assert.True(t, out.Columns[0][0].Removable)
	// empty column
	assert.Len(t, out.Columns[2], 0)
	// col3 has two cards, only the last is top
	assert.False(t, out.Columns[3][0].Top)
	assert.True(t, out.Columns[3][1].Top)
}

func TestNarcoticWebPresenterOutput_EmptyDiscard(t *testing.T) {
	g := new(interfaces.MockNarcoticGame)
	setupNarcoticOutputMock(g)
	// Override the discard top to nil (no cards removed yet).
	g.ExpectedCalls = nil
	g.On("GetPhase").Return(domain.NarcoticPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("GetStockCount").Return(48).Maybe()
	g.On("GetDiscardCount").Return(0).Maybe()
	g.On("GetDiscardTop").Return((*domain.Card)(nil)).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetColumns").Return(sampleNarcoticColumns()).Maybe()
	g.On("CanRemoveSet").Return(false).Maybe()
	g.On("GetRedealCount").Return(0).Maybe()
	for c := range domain.NarcoticColCnt {
		g.On("CanMove", c).Return(false).Maybe()
	}
	g.On("GetHint").Return(nil).Maybe()
	p := &NarcoticWebPresenter{}

	out := parseNarcoticOutput(t, p.Output(g, nil))
	// With an empty discard pile, the omitempty field is absent.
	assert.Nil(t, out.DiscardTop)
	assert.Equal(t, 0, out.DiscardCount)
}

func TestNarcoticWebPresenterOutput_Error(t *testing.T) {
	g := new(interfaces.MockNarcoticGame)
	setupNarcoticOutputMock(g)
	p := &NarcoticWebPresenter{}

	out := parseNarcoticOutput(t, p.Output(g, errors.New("test error")))
	assert.Equal(t, "test error", out.Message)
}

func TestNarcoticWebPresenterOutput_Stalemate(t *testing.T) {
	g := new(interfaces.MockNarcoticGame)
	g.On("GetPhase").Return(domain.NarcoticPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(5).Maybe()
	g.On("GetStockCount").Return(0).Maybe()
	g.On("GetDiscardCount").Return(10).Maybe()
	g.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 3, false)).Maybe()
	g.On("CanUndo").Return(true).Maybe()
	g.On("IsStalemate").Return(true).Maybe()
	g.On("UndoToEscape").Return(-1).Maybe()
	var cols [domain.NarcoticColCnt][]*domain.Card
	g.On("GetColumns").Return(cols).Maybe()
	p := &NarcoticWebPresenter{}

	out := parseNarcoticOutput(t, p.Output(g, nil))
	assert.Equal(t, "narcotic.stalemate", out.MessageCode)
}

func TestNarcoticWebPresenterOutput_GameClear(t *testing.T) {
	g := new(interfaces.MockNarcoticGame)
	g.On("GetPhase").Return(domain.NarcoticPhaseGameClear).Maybe()
	g.On("GetMoveCount").Return(20).Maybe()
	g.On("GetStockCount").Return(0).Maybe()
	g.On("GetDiscardCount").Return(48).Maybe()
	g.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignClover, 5, false)).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	var cols [domain.NarcoticColCnt][]*domain.Card
	g.On("GetColumns").Return(cols).Maybe()
	p := &NarcoticWebPresenter{}

	out := parseNarcoticOutput(t, p.Output(g, nil))
	assert.Equal(t, "narcotic.gameClear", out.MessageCode)
	assert.Contains(t, out.Message, "20")
}

func TestNarcoticWebPresenterOutput_GameOver(t *testing.T) {
	g := new(interfaces.MockNarcoticGame)
	g.On("GetPhase").Return(domain.NarcoticPhaseGameOver).Maybe()
	g.On("GetMoveCount").Return(5).Maybe()
	g.On("GetStockCount").Return(0).Maybe()
	g.On("GetDiscardCount").Return(2).Maybe()
	g.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignSpade, 4, false)).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	var cols [domain.NarcoticColCnt][]*domain.Card
	g.On("GetColumns").Return(cols).Maybe()
	p := &NarcoticWebPresenter{}

	out := parseNarcoticOutput(t, p.Output(g, nil))
	assert.Equal(t, "narcotic.gameOver", out.MessageCode)
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない (#4483)。
func TestNarcoticWebPresenterOutputCarriesTheHint(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		aug := new(interfaces.MockNarcoticGame)
		setupNarcoticWebMockDefaults(aug)
		aug.On("GetHint").Return(&domain.NarcoticHint{Type: "remove", Col: 1}).Maybe()

		result := new(NarcoticWebPresenter).Output(aug, nil)
		assert.Contains(t, result, `"hint"`, "Output must carry the hint -- the frontend reads state.hint")
	})

	// 手詰まりのヒントは出さない。逃げ道の提示は stalemate 用のメッセージが持つ。
	t.Run("not while stalemate", func(t *testing.T) {
		aug := new(interfaces.MockNarcoticGame)
		setupNarcoticWebMockDefaults(aug)
		aug.ExpectedCalls = filterCalls(aug.ExpectedCalls, "IsStalemate")
		aug.On("IsStalemate").Return(true)
		aug.On("GetHint").Return(&domain.NarcoticHint{Type: "remove", Col: 1}).Maybe()

		result := new(NarcoticWebPresenter).Output(aug, nil)
		assert.NotContains(t, result, `"hint"`)
	})
}

func TestNarcoticWebPresenterHintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockNarcoticGame)
		g.On("GetHint").Return(&domain.NarcoticHint{Type: "move", Col: 2})
		g.On("GetPhase").Return(domain.NarcoticPhasePlaying)
		g.On("GetMoveCount").Return(0)
		g.On("GetStockCount").Return(0)
		g.On("GetDiscardCount").Return(4)
		g.On("CanUndo").Return(false)
		g.On("IsStalemate").Return(false)
		g.On("UndoToEscape").Return(0)

		p := &NarcoticWebPresenter{}
		out := parseNarcoticOutput(t, p.HintOutput(g))
		assert.NotNil(t, out.Hint)
		assert.Equal(t, "move", out.Hint.Type)
		assert.Equal(t, "narcotic.hintAvailable", out.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockNarcoticGame)
		g.On("GetHint").Return((*domain.NarcoticHint)(nil))
		g.On("GetPhase").Return(domain.NarcoticPhasePlaying)
		g.On("GetMoveCount").Return(0)
		g.On("GetStockCount").Return(0)
		g.On("GetDiscardCount").Return(4)
		g.On("CanUndo").Return(false)
		g.On("IsStalemate").Return(false)
		g.On("UndoToEscape").Return(0)

		p := &NarcoticWebPresenter{}
		out := parseNarcoticOutput(t, p.HintOutput(g))
		assert.Nil(t, out.Hint)
		assert.Equal(t, "narcotic.noHint", out.MessageCode)
	})
}

func TestNarcoticWebPresenterActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockNarcoticGame)
		g.On("GetPhase").Return(domain.NarcoticPhasePlaying)
		g.On("GetGameEndFlag").Return(false)
		p := &NarcoticWebPresenter{}
		assert.Contains(t, p.ActionLogOutput(g), "entries")
	})

	t.Run("game over", func(t *testing.T) {
		g := new(interfaces.MockNarcoticGame)
		g.On("GetPhase").Return(domain.NarcoticPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
		p := &NarcoticWebPresenter{}
		assert.Contains(t, p.ActionLogOutput(g), "entries")
	})
}
