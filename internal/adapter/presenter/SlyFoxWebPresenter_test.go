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

func setupSlyFoxWebMockDefaults(g *interfaces.MockSlyFoxGame) {
	g.On("GetPhase").Return(domain.SlyFoxPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(84).Maybe()
	g.On("DealtThisCycle").Return(domain.SlyFoxDealCycle).Maybe()
	g.On("ReserveIsLocked").Return(false).Maybe()

	var tableau [domain.SlyFoxTableauCnt][]*domain.Card
	for i := range domain.SlyFoxTableauCnt {
		tableau[i] = []*domain.Card{domain.NewCard(domain.CardDesignSpade, i+2, true)}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.SlyFoxFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()

	// The first half build up from the Ace, the second half down from the King.
	for i := range domain.SlyFoxFoundationCnt {
		g.On("IsAscendingFoundation", i).Return(i < domain.SlyFoxAscendingCnt).Maybe()
	}
}

func parseSlyFoxOutput(t *testing.T, jsonStr string) *controller.SlyFoxWebOutput {
	t.Helper()
	var out controller.SlyFoxWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupSlyFoxOutputMock は Output 用の既定を組む。
//
// **Output() も受動ヒントを埋めるようになった** (#4483) ので、GetHint を
// 呼べるようにしておく必要がある。共有ヘルパー側に置くと、先に登録された
// この期待が HintOutput テストの「ヒントあり」を食ってしまう。
func setupSlyFoxOutputMock(g *interfaces.MockSlyFoxGame) {
	setupSlyFoxWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestSlyFoxWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		setupSlyFoxOutputMock(g)

		result := parseSlyFoxOutput(t, new(SlyFoxWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 84, result.StockCount)
		assert.Len(t, result.Tableau, domain.SlyFoxTableauCnt)
		assert.Len(t, result.Foundation, domain.SlyFoxFoundationCnt)
		// The build direction has to be on the wire: the page cannot label the
		// piles from the index without silently duplicating the rule.
		assert.Equal(t,
			[]bool{true, true, true, true, false, false, false, false},
			result.FoundationAscending)
		// **周の進みを API に載せる。**載せないと、ページは「まだ開いていない
		// リザーブ」を押せる状態で描く。
		assert.Equal(t, domain.SlyFoxDealCycle, result.DealtThisCycle)
		assert.Equal(t, domain.SlyFoxDealCycle, result.DealCycle)
		assert.False(t, result.ReserveLocked)
		assert.Equal(t, "slyfox.playing", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		setupSlyFoxOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)

		result := parseSlyFoxOutput(t, new(SlyFoxWebPresenter).Output(g, nil))
		assert.Equal(t, "slyfox.stalemate", result.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		setupSlyFoxOutputMock(g)

		result := parseSlyFoxOutput(t, new(SlyFoxWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name string
		val  domain.SlyFoxPhase
		code string
	}{
		{"game clear", domain.SlyFoxPhaseGameClear, "slyfox.gameClear"},
		{"game over", domain.SlyFoxPhaseGameOver, "slyfox.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockSlyFoxGame)
			setupSlyFoxOutputMock(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseSlyFoxOutput(t, new(SlyFoxWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない。ここが埋まっていないと
// フロントの `state.hint` は常に undefined で、それを読む分岐が全部死ぬ (#4483)。
func TestSlyFoxWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.SlyFoxHint{FromZone: "tableau", FromIdx: 2, ToZone: "foundation", ToIdx: 1}

	t.Run("while the game is playable", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		setupSlyFoxWebMockDefaults(g)
		g.On("GetHint").Return(hint).Maybe()

		result := parseSlyFoxOutput(t, new(SlyFoxWebPresenter).Output(g, nil))
		if result.Hint == nil {
			t.Fatal("Output must carry the hint -- the frontend reads state.hint")
		}
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.FromIdx)
	})
}

func TestSlyFoxWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		setupSlyFoxWebMockDefaults(g)
		g.On("GetHint").Return(&domain.SlyFoxHint{
			FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: 3,
		})

		result := parseSlyFoxOutput(t, new(SlyFoxWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "stock", result.Hint.FromZone)
		assert.Equal(t, "tableau", result.Hint.ToZone)
		assert.Equal(t, 3, result.Hint.ToIdx)
		assert.Equal(t, "slyfox.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		setupSlyFoxWebMockDefaults(g)
		g.On("GetHint").Return((*domain.SlyFoxHint)(nil))

		result := parseSlyFoxOutput(t, new(SlyFoxWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "slyfox.noHint", result.MessageCode)
	})
}

func TestSlyFoxWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		g.On("GetPhase").Return(domain.SlyFoxPhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(SlyFoxWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockSlyFoxGame)
		g.On("GetPhase").Return(domain.SlyFoxPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(SlyFoxWebPresenter).ActionLogOutput(g), "move")
	})
}
