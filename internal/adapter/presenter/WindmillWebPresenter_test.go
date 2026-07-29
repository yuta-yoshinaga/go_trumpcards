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

func setupWindmillWebMockDefaults(g *interfaces.MockWindmillGame) {
	g.On("GetPhase").Return(domain.WindmillPhasePlaying).Maybe()
	g.On("GetMoveCount").Return(0).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("IsStalemate").Return(false).Maybe()
	g.On("UndoToEscape").Return(0).Maybe()
	g.On("GetStockCount").Return(95).Maybe()
	g.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 9, true)}).Maybe()
	g.On("IsTransferBlocked").Return(false).Maybe()
	g.On("GetCenter").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, true)}).Maybe()

	var sails [domain.WindmillSailCnt]*domain.Card
	for i := range domain.WindmillSailCnt {
		sails[i] = domain.NewCard(domain.CardDesignClover, i+2, true)
	}
	g.On("GetSails").Return(sails).Maybe()

	var corners [domain.WindmillCornerCnt][]*domain.Card
	g.On("GetCorners").Return(corners).Maybe()
}

func parseWindmillOutput(t *testing.T, jsonStr string) *controller.WindmillWebOutput {
	t.Helper()
	var out controller.WindmillWebOutput
	assert.NoError(t, json.Unmarshal([]byte(jsonStr), &out))
	return &out
}

func TestWindmillWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		setupWindmillWebMockDefaults(g)

		result := parseWindmillOutput(t, new(WindmillWebPresenter).Output(g, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 95, result.StockCount)
		assert.Len(t, result.Sails, domain.WindmillSailCnt)
		assert.Len(t, result.Corners, domain.WindmillCornerCnt)
		assert.Len(t, result.Center, 1)
		assert.Len(t, result.Waste, 1)
		assert.False(t, result.TransferBlocked)
		assert.Equal(t, "windmill.playing", result.MessageCode)
	})

	// A sail that can no longer be refilled stays as a hole, so the slot has to
	// survive the wire as null rather than being compacted away.
	t.Run("an unrefillable sail is sent as null", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		setupWindmillWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetSails")
		var sails [domain.WindmillSailCnt]*domain.Card
		sails[0] = domain.NewCard(domain.CardDesignSpade, 5, true)
		g.On("GetSails").Return(sails)

		result := parseWindmillOutput(t, new(WindmillWebPresenter).Output(g, nil))
		assert.Len(t, result.Sails, domain.WindmillSailCnt)
		assert.NotNil(t, result.Sails[0])
		assert.Nil(t, result.Sails[1])
	})

	// The client must not have to remember that the previous move was a pull-back:
	// the restriction is the domain's judgement and gets its own field and message.
	t.Run("the transfer block is surfaced", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		setupWindmillWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsTransferBlocked")
		g.On("IsTransferBlocked").Return(true)

		result := parseWindmillOutput(t, new(WindmillWebPresenter).Output(g, nil))
		assert.True(t, result.TransferBlocked)
		assert.Equal(t, "windmill.transferBlocked", result.MessageCode)
	})

	t.Run("stalemate outranks the transfer block", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		setupWindmillWebMockDefaults(g)
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsTransferBlocked")
		g.ExpectedCalls = filterCalls(g.ExpectedCalls, "IsStalemate")
		g.On("IsTransferBlocked").Return(true)
		g.On("IsStalemate").Return(true)

		result := parseWindmillOutput(t, new(WindmillWebPresenter).Output(g, nil))
		assert.Equal(t, "windmill.stalemate", result.MessageCode)
	})

	t.Run("error message", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		setupWindmillWebMockDefaults(g)

		result := parseWindmillOutput(t, new(WindmillWebPresenter).Output(g, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	for _, tc := range []struct {
		name string
		val  domain.WindmillPhase
		code string
	}{
		{"game clear", domain.WindmillPhaseGameClear, "windmill.gameClear"},
		{"game over", domain.WindmillPhaseGameOver, "windmill.gameOver"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := new(interfaces.MockWindmillGame)
			setupWindmillWebMockDefaults(g)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPhase")
			g.On("GetPhase").Return(tc.val)

			result := parseWindmillOutput(t, new(WindmillWebPresenter).Output(g, nil))
			assert.Equal(t, tc.code, result.MessageCode)
		})
	}
}

func TestWindmillWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		setupWindmillWebMockDefaults(g)
		g.On("GetHint").Return(&domain.WindmillHint{
			FromZone: "corner", FromIdx: 2, ToZone: "center", ToIdx: -1,
		})

		result := parseWindmillOutput(t, new(WindmillWebPresenter).HintOutput(g))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "corner", result.Hint.FromZone)
		assert.Equal(t, 2, result.Hint.FromIdx)
		assert.Equal(t, "center", result.Hint.ToZone)
		assert.Equal(t, "windmill.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		setupWindmillWebMockDefaults(g)
		g.On("GetHint").Return((*domain.WindmillHint)(nil))

		result := parseWindmillOutput(t, new(WindmillWebPresenter).HintOutput(g))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "windmill.noHint", result.MessageCode)
	})
}

func TestWindmillWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		g.On("GetPhase").Return(domain.WindmillPhasePlaying)
		g.On("GetGameEndFlag").Return(false)

		assert.Contains(t, new(WindmillWebPresenter).ActionLogOutput(g), "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		g := new(interfaces.MockWindmillGame)
		g.On("GetPhase").Return(domain.WindmillPhaseGameOver)
		g.On("GetGameEndFlag").Return(true)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "move", Detail: "test"},
		})

		assert.Contains(t, new(WindmillWebPresenter).ActionLogOutput(g), "move")
	})
}
