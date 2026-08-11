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

func setupCrazyQuiltWebMockDefaults(g *interfaces.MockCrazyQuiltGame) {
	g.On("GetPhase").Return(domain.CrazyQuiltPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(32).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()

	// キルトは 64 マス。マス 5 だけ空けて、空きマスの描画も踏む。
	var quilt [domain.CrazyQuiltCells]*domain.Card
	for i := range domain.CrazyQuiltCells {
		quilt[i] = domain.NewCard(domain.CardDesignSpade, (i%13)+1, true)
	}
	quilt[5] = nil
	g.On("GetQuilt").Return(quilt).Maybe()
	for i := range domain.CrazyQuiltCells {
		g.On("IsAvailable", i).Return(i < 8).Maybe()
	}
	g.On("GetRedealsLeft").Return(domain.CrazyQuiltRedealCnt).Maybe()
	for i := range domain.CrazyQuiltFoundationCnt {
		g.On("IsAscendingFoundation", i).Return(i < domain.CrazyQuiltAscendingCnt).Maybe()
	}

	var foundation [domain.CrazyQuiltFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func parseCrazyQuiltOutput(t *testing.T, jsonStr string) *controller.CrazyQuiltWebOutput {
	t.Helper()
	var out controller.CrazyQuiltWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

// setupCrazyQuiltOutputMock は Output 用の既定を組む。
//
// **Output() も受動ヒントを埋めるようになった** (#4483) ので、GetHint を
// 呼べるようにしておく必要がある。共有ヘルパー側に置くと、先に登録された
// この期待が HintOutput テストの「ヒントあり」を食ってしまう。
func setupCrazyQuiltOutputMock(g *interfaces.MockCrazyQuiltGame) {
	setupCrazyQuiltWebMockDefaults(g)
	g.On("GetHint").Return(nil).Maybe()
}

func TestCrazyQuiltWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		setupCrazyQuiltOutputMock(g)

		result := parseCrazyQuiltOutput(t, new(CrazyQuiltWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 32, result.StockCount)
		assert.Len(t, result.Quilt, domain.CrazyQuiltCells)
		// 可動判定はサーバが送る。フロントで短辺の向きを再実装させない。
		assert.Len(t, result.Available, domain.CrazyQuiltCells)
		assert.True(t, result.Available[0])
		assert.False(t, result.Available[63])
		assert.Equal(t, domain.CrazyQuiltRedealCnt, result.RedealsLeft)
		assert.Equal(t,
			[]bool{true, true, true, true, false, false, false, false},
			result.FoundationAscending)
		assert.Len(t, result.Foundation, domain.CrazyQuiltFoundationCnt)
		assert.Len(t, result.Waste, 1)
		assert.Equal(t, "crazyquilt.playing", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		setupCrazyQuiltOutputMock(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)

		result := parseCrazyQuiltOutput(t, new(CrazyQuiltWebPresenter).Output(g, nil))
		assert.Equal(t, "crazyquilt.stalemate", result.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		setupCrazyQuiltOutputMock(g)

		result := parseCrazyQuiltOutput(t, new(CrazyQuiltWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name string
		val  domain.CrazyQuiltPhase
		code string
	}{
		{"game clear", domain.CrazyQuiltPhaseGameClear, "crazyquilt.gameClear"},
		{"game over", domain.CrazyQuiltPhaseGameOver, "crazyquilt.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockCrazyQuiltGame)
			setupCrazyQuiltOutputMock(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseCrazyQuiltOutput(t, new(CrazyQuiltWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}
}

// **受動ヒントは Output() に載る。**HintOutput() は `command: "hint"` 専用の
// レスポンスで、ページの state にはマージされない。ここが埋まっていないと
// フロントの `state.hint` は常に undefined で、それを読む分岐が全部死ぬ (#4483)。
func TestCrazyQuiltWebPresenter_OutputCarriesTheHint(t *testing.T) {
	hint := &domain.CrazyQuiltHint{FromZone: "tableau", FromIdx: 2, ToZone: "foundation", ToIdx: 1}

	t.Run("while the game is playable", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		setupCrazyQuiltWebMockDefaults(g)
		g.On("GetHint").Return(hint).Maybe()

		result := parseCrazyQuiltOutput(t, new(CrazyQuiltWebPresenter).Output(g, nil))
		if result.Hint == nil {
			t.Fatal("Output must carry the hint -- the frontend reads state.hint")
		}
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.FromIdx)
	})
}

func TestCrazyQuiltWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		setupCrazyQuiltWebMockDefaults(g)
		g.On("GetHint").Return(&domain.CrazyQuiltHint{
			FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: 3,
		})

		result := parseCrazyQuiltOutput(t, new(CrazyQuiltWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "stock", result.Hint.FromZone)
		assert.Equal(t, "tableau", result.Hint.ToZone)
		assert.Equal(t, 3, result.Hint.ToIdx)
		assert.Equal(t, "crazyquilt.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		setupCrazyQuiltWebMockDefaults(g)
		g.On("GetHint").Return((*domain.CrazyQuiltHint)(nil))

		result := parseCrazyQuiltOutput(t, new(CrazyQuiltWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "crazyquilt.noHint", result.MessageCode)
	})
}

func TestCrazyQuiltWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		g.On("GetPhase").Return(domain.CrazyQuiltPhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(CrazyQuiltWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockCrazyQuiltGame)
		g.On("GetPhase").Return(domain.CrazyQuiltPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(CrazyQuiltWebPresenter).ActionLogOutput(g), "move")
	})
}
