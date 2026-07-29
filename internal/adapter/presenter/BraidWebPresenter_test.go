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

func setupBraidWebMockDefaults(g *interfaces.MockBraidGame) {
	g.On("GetPhase").Return(domain.BraidPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(71).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()
	g.On("GetBraid").Return([]*domain.Card{domain.NewCard(domain.CardDesignClover, 3, true)}).Maybe()
	g.On("GetBaseRank").Return(5).Maybe()
	g.On("GetDirection").Return(domain.BraidDirectionAscending).Maybe()
	g.On("IsAwaitingDirection").Return(false).Maybe()
	g.On("GetPassesUsed").Return(0).Maybe()
	g.On("CanRedeal").Return(false).Maybe()

	// 枠 1 だけ空にして、空き枠が詰められずに null で残ることを見る。
	var fields [domain.BraidFieldCnt]*domain.Card
	for i := range domain.BraidFieldCnt {
		if i == 1 {
			continue
		}
		fields[i] = domain.NewCard(domain.CardDesignSpade, i+2, true)
	}
	g.On("GetFields").Return(fields).Maybe()

	var helpers [domain.BraidHelperCnt]*domain.Card
	for i := range domain.BraidHelperCnt {
		helpers[i] = domain.NewCard(domain.CardDesignDiamond, i+2, true)
	}
	g.On("GetHelpers").Return(helpers).Maybe()

	var foundation [domain.BraidFoundationCnt][]*domain.Card
	g.On("GetFoundation").Return(foundation).Maybe()
}

func parseBraidOutput(t *testing.T, jsonStr string) *controller.BraidWebOutput {
	t.Helper()
	var out controller.BraidWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

func TestBraidWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidWebMockDefaults(g)

		result := parseBraidOutput(t, new(BraidWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 71, result.StockCount)
		assert.Len(t, result.Braid, 1)
		assert.Len(t, result.Fields, domain.BraidFieldCnt)
		assert.Len(t, result.Helpers, domain.BraidHelperCnt)
		assert.Len(t, result.Foundation, domain.BraidFoundationCnt)
		assert.Equal(t, 5, result.BaseRank)
		assert.Equal(t, int(domain.BraidDirectionAscending), result.Direction)
		assert.False(t, result.AwaitingDirection)
		assert.Equal(t, domain.BraidMaxPasses-1, result.RedealsLeft)
		assert.Equal(t, "braid.playing", result.MessageCode)
	})

	// 空き枠を詰めるとインデックスがずれ、ヒントの枠番号が画面と食い違う。
	t.Run("an empty slot stays null instead of being squeezed out", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidWebMockDefaults(g)

		result := parseBraidOutput(t, new(BraidWebPresenter).Output(g, nil))
		assert.Len(t, result.Fields, domain.BraidFieldCnt)
		assert.Nil(t, result.Fields[1])
		assert.NotNil(t, result.Fields[0])
		assert.NotNil(t, result.Fields[2])
	})

	// The client must not have to infer "direction 0 means unchosen".
	t.Run("awaiting the direction has its own message", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsAwaitingDirection")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetDirection")
		g.On("IsAwaitingDirection").Return(true)
		g.On("GetDirection").Return(domain.BraidDirectionUnset)

		result := parseBraidOutput(t, new(BraidWebPresenter).Output(g, nil))
		assert.True(t, result.AwaitingDirection)
		assert.Equal(t, 0, result.Direction)
		assert.Equal(t, "braid.chooseDirection", result.MessageCode)
	})

	t.Run("awaiting the direction outranks stalemate", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsAwaitingDirection")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsAwaitingDirection").Return(true)
		g.On("IsStalemate").Return(true)

		result := parseBraidOutput(t, new(BraidWebPresenter).Output(g, nil))
		assert.Equal(t, "braid.chooseDirection", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsStalemate").Return(true)

		result := parseBraidOutput(t, new(BraidWebPresenter).Output(g, nil))
		assert.Equal(t, "braid.stalemate", result.MessageCode)
	})

	t.Run("redeals left counts down with the passes used", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPassesUsed")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "CanRedeal")
		g.On("GetPassesUsed").Return(1)
		g.On("CanRedeal").Return(true)

		result := parseBraidOutput(t, new(BraidWebPresenter).Output(g, nil))
		assert.Equal(t, 1, result.RedealsLeft)
		assert.True(t, result.CanRedeal)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidWebMockDefaults(g)

		result := parseBraidOutput(t, new(BraidWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name string
		val  domain.BraidPhase
		code string
	}{
		{"game clear", domain.BraidPhaseGameClear, "braid.gameClear"},
		{"game over", domain.BraidPhaseGameOver, "braid.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockBraidGame)
			setupBraidWebMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseBraidOutput(t, new(BraidWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}
}

func TestBraidWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidWebMockDefaults(g)
		g.On("GetHint").Return(&domain.BraidHint{
			FromZone: "field", FromIdx: 2, ToZone: "foundation", ToIdx: 3,
		})

		result := parseBraidOutput(t, new(BraidWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "field", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.FromIdx)
		assert.Equal(t, 3, result.Hint.ToIdx)
		assert.Equal(t, "braid.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		setupBraidWebMockDefaults(g)
		g.On("GetHint").Return((*domain.BraidHint)(nil))

		result := parseBraidOutput(t, new(BraidWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "braid.noHint", result.MessageCode)
	})
}

func TestBraidWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		g.On("GetPhase").Return(domain.BraidPhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(BraidWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockBraidGame)
		g.On("GetPhase").Return(domain.BraidPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(BraidWebPresenter).ActionLogOutput(g), "move")
	})
}
