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

func setupDuchessWebMockDefaults(g *interfaces.MockDuchessGame) {
	g.On("GetPhase").Return(domain.DuchessPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(35).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()
	g.On("GetBaseRank").Return(5).Maybe()
	g.On("IsAwaitingBaseRank").Return(false).Maybe()

	var reserve [domain.DuchessReserveCnt][]*domain.Card
	for i := range domain.DuchessReserveCnt {
		reserve[i] = []*domain.Card{domain.NewCard(domain.CardDesignClover, i+2, true)}
	}
	g.On("GetReserve").Return(reserve).Maybe()

	var tableau [domain.DuchessTableauCnt][]*domain.DuchessTableauCard
	for i := range domain.DuchessTableauCnt {
		tableau[i] = []*domain.DuchessTableauCard{
			{Card: domain.NewCard(domain.CardDesignSpade, i+2, false), FaceUp: true},
		}
	}
	g.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.DuchessFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func parseDuchessOutput(t *testing.T, jsonStr string) *controller.DuchessWebOutput {
	t.Helper()
	var out controller.DuchessWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupDuchessOutputMock は Output 用の既定を組む。
//
// **Output() も受動ヒントを埋めるようになった** (#4483) ので、GetHint を
// 呼べるようにしておく必要がある。共有ヘルパー側に置くと、先に登録された
// この期待が HintOutput テストの「ヒントあり」を食ってしまう。
func setupDuchessOutputMock(g *interfaces.MockDuchessGame) {
	setupDuchessWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestDuchessWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessOutputMock(g)

		result := parseDuchessOutput(t, new(DuchessWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 35, result.StockCount)
		assert.Len(t, result.Reserve, domain.DuchessReserveCnt)
		assert.Len(t, result.Tableau, domain.DuchessTableauCnt)
		assert.Len(t, result.Foundation, domain.DuchessFoundationCnt)
		assert.Len(t, result.Waste, 1)
		assert.Equal(t, 5, result.BaseRank)
		assert.False(t, result.AwaitingBaseRank)
		assert.Equal(t, "duchess.playing", result.MessageCode)
	})

	// The client must not have to infer "rank 0 means unchosen": awaitingBaseRank
	// is the domain's own judgement and is surfaced as its own field and message.
	t.Run("awaiting the base rank has its own message", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsAwaitingBaseRank")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetBaseRank")
		g.On("IsAwaitingBaseRank").Return(true)
		g.On("GetBaseRank").Return(0)

		result := parseDuchessOutput(t, new(DuchessWebPresenter).Output(g, nil))
		assert.True(t, result.AwaitingBaseRank)
		assert.Equal(t, 0, result.BaseRank)
		assert.Equal(t, "duchess.chooseBase", result.MessageCode)
	})

	// Choosing the base rank is always possible, so it must win over stalemate --
	// otherwise the opening board would look like a dead end.
	t.Run("awaiting the base rank outranks stalemate", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsAwaitingBaseRank")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsAwaitingBaseRank").Return(true)
		g.On("IsStalemate").Return(true)

		result := parseDuchessOutput(t, new(DuchessWebPresenter).Output(g, nil))
		assert.Equal(t, "duchess.chooseBase", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)

		result := parseDuchessOutput(t, new(DuchessWebPresenter).Output(g, nil))
		assert.Equal(t, "duchess.stalemate", result.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessOutputMock(g)

		result := parseDuchessOutput(t, new(DuchessWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name string
		val  domain.DuchessPhase
		code string
	}{
		{"game clear", domain.DuchessPhaseGameClear, "duchess.gameClear"},
		{"game over", domain.DuchessPhaseGameOver, "duchess.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockDuchessGame)
			setupDuchessOutputMock(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseDuchessOutput(t, new(DuchessWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない。ここが埋まっていないと
// フロントの `state.hint` は常に undefined で、それを読む分岐が全部死ぬ (#4483)。
func TestDuchessWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.DuchessHint{FromZone: "tableau", FromIdx: 2, ToZone: "foundation", ToIdx: 1}

	t.Run("while the game is playable", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessWebMockDefaults(g)
		g.On("GetHint").Return(hint).Maybe()

		result := parseDuchessOutput(t, new(DuchessWebPresenter).Output(g, nil))
		if result.Hint == nil {
			t.Fatal("Output must carry the hint -- the frontend reads state.hint")
		}
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.FromIdx)
	})
}

func TestDuchessWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessWebMockDefaults(g)
		g.On("GetHint").Return(&domain.DuchessHint{
			FromZone: "reserve", FromIdx: 2, CardIndex: -1, ToZone: "foundation", ToIdx: 1,
		})

		result := parseDuchessOutput(t, new(DuchessWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "reserve", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.FromIdx)
		assert.Equal(t, 1, result.Hint.ToIdx)
		assert.Equal(t, "duchess.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		setupDuchessWebMockDefaults(g)
		g.On("GetHint").Return((*domain.DuchessHint)(nil))

		result := parseDuchessOutput(t, new(DuchessWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "duchess.noHint", result.MessageCode)
	})
}

func TestDuchessWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		g.On("GetPhase").Return(domain.DuchessPhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(DuchessWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockDuchessGame)
		g.On("GetPhase").Return(domain.DuchessPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(DuchessWebPresenter).ActionLogOutput(g), "move")
	})
}
