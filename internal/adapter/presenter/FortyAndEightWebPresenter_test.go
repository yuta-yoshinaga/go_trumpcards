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

func setupFortyAndEightWebMockDefaults(fg *interfaces.MockFortyAndEightGame) {
	fg.On("GetPhase").Return(domain.FortyAndEightPhasePlaying).Maybe()
	fg.On("GetMoveCount").Return(0).Maybe()
	fg.On("GetStockCount").Return(64).Maybe()
	fg.On("GetWaste").Return(([]*domain.Card)(nil)).Maybe()
	fg.On("CanUndo").Return(false).Maybe()
	fg.On("IsStalemate").Return(false).Maybe()
	fg.On("UndoToEscape").Return(0).Maybe()
	fg.On("GetRedealUsed").Return(false).Maybe()
	fg.On("CanRedeal").Return(false).Maybe()

	var tableau [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard
	for i := range domain.FortyAndEightTableauCnt {
		tableau[i] = make([]*domain.FortyAndEightTableauCard, 5)
		for j := range 5 {
			tableau[i][j] = &domain.FortyAndEightTableauCard{
				Card:   domain.NewCard(domain.CardDesignSpade, j+1, false),
				FaceUp: true,
			}
		}
	}
	fg.On("GetTableau").Return(tableau).Maybe()

	var foundation [domain.FortyAndEightFoundationCnt][]*domain.Card
	fg.On("GetFoundation").Return(foundation).Maybe()
}

func parseFortyAndEightOutput(t *testing.T, jsonStr string) *controller.FortyAndEightWebOutput {
	t.Helper()
	var out controller.FortyAndEightWebOutput
	err := json.Unmarshal([]byte(jsonStr), &out)
	assert.NoError(t, err)
	return &out
}

func TestFortyAndEightWebPresenter_Output(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightWebMockDefaults(fg)
		p := new(FortyAndEightWebPresenter)

		result := parseFortyAndEightOutput(t, p.Output(fg, nil))
		assert.Equal(t, 0, result.Phase)
		assert.Equal(t, 0, result.MoveCount)
		assert.Equal(t, 64, result.StockCount)
		assert.False(t, result.RedealUsed)
		assert.False(t, result.CanRedeal)
		assert.Empty(t, result.Waste)
		assert.Len(t, result.Tableau, domain.FortyAndEightTableauCnt)
		assert.Len(t, result.Foundation, domain.FortyAndEightFoundationCnt)
		assert.Equal(t, "fortyandeight.playing", result.MessageCode)
	})

	t.Run("waste with cards", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightWebMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetWaste")
		fg.On("GetWaste").Return([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})

		p := new(FortyAndEightWebPresenter)
		result := parseFortyAndEightOutput(t, p.Output(fg, nil))
		assert.Len(t, result.Waste, 1)
		assert.Equal(t, "HEART", result.Waste[0].Design)
		assert.Equal(t, 5, result.Waste[0].Value)
	})

	t.Run("redeal available", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightWebMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "CanRedeal")
		fg.On("CanRedeal").Return(true)

		p := new(FortyAndEightWebPresenter)
		result := parseFortyAndEightOutput(t, p.Output(fg, nil))
		assert.True(t, result.CanRedeal)
	})

	t.Run("redeal used", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightWebMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetRedealUsed")
		fg.On("GetRedealUsed").Return(true)

		p := new(FortyAndEightWebPresenter)
		result := parseFortyAndEightOutput(t, p.Output(fg, nil))
		assert.True(t, result.RedealUsed)
	})

	t.Run("all face up", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightWebMockDefaults(fg)
		p := new(FortyAndEightWebPresenter)

		result := parseFortyAndEightOutput(t, p.Output(fg, nil))
		for _, col := range result.Tableau {
			for _, tc := range col {
				assert.True(t, tc.FaceUp)
				assert.NotNil(t, tc.Card)
			}
		}
	})

	t.Run("error message", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightWebMockDefaults(fg)
		p := new(FortyAndEightWebPresenter)

		result := parseFortyAndEightOutput(t, p.Output(fg, errors.New("test error")))
		assert.Equal(t, "test error", result.Message)
	})

	t.Run("game clear", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightWebMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetPhase")
		fg.On("GetPhase").Return(domain.FortyAndEightPhaseGameClear)

		p := new(FortyAndEightWebPresenter)
		result := parseFortyAndEightOutput(t, p.Output(fg, nil))
		assert.Equal(t, "fortyandeight.gameClear", result.MessageCode)
	})

	t.Run("game over", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightWebMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "GetPhase")
		fg.On("GetPhase").Return(domain.FortyAndEightPhaseGameOver)

		p := new(FortyAndEightWebPresenter)
		result := parseFortyAndEightOutput(t, p.Output(fg, nil))
		assert.Equal(t, "fortyandeight.gameOver", result.MessageCode)
	})

	t.Run("stalemate", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightWebMockDefaults(fg)
		fg.ExpectedCalls = filterCalls(fg.ExpectedCalls, "IsStalemate")
		fg.On("IsStalemate").Return(true)

		p := new(FortyAndEightWebPresenter)
		result := parseFortyAndEightOutput(t, p.Output(fg, nil))
		assert.Equal(t, "fortyandeight.stalemate", result.MessageCode)
	})
}

func TestFortyAndEightWebPresenter_HintOutput(t *testing.T) {
	t.Run("with hint", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightWebMockDefaults(fg)
		fg.On("GetHint").Return(&domain.FortyAndEightHint{
			FromZone:  "tableau",
			FromCol:   0,
			CardIndex: 3,
			ToZone:    "foundation",
			ToCol:     0,
		})

		p := new(FortyAndEightWebPresenter)
		result := parseFortyAndEightOutput(t, p.HintOutput(fg))
		assert.NotNil(t, result.Hint)
		assert.Equal(t, "tableau", result.Hint.FromZone)
		assert.Equal(t, "fortyandeight.hintAvailable", result.MessageCode)
	})

	t.Run("no hint", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		setupFortyAndEightWebMockDefaults(fg)
		fg.On("GetHint").Return((*domain.FortyAndEightHint)(nil))

		p := new(FortyAndEightWebPresenter)
		result := parseFortyAndEightOutput(t, p.HintOutput(fg))
		assert.Nil(t, result.Hint)
		assert.Equal(t, "fortyandeight.noHint", result.MessageCode)
	})
}

func TestFortyAndEightWebPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing phase returns empty", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		fg.On("GetPhase").Return(domain.FortyAndEightPhasePlaying)

		fg.On("GetGameEndFlag").Return(false)
		p := new(FortyAndEightWebPresenter)
		result := p.ActionLogOutput(fg)
		assert.Contains(t, result, "[]")
	})

	t.Run("game over returns log", func(t *testing.T) {
		fg := new(interfaces.MockFortyAndEightGame)
		fg.On("GetPhase").Return(domain.FortyAndEightPhaseGameOver)
		fg.On("GetGameEndFlag").Return(true)
		fg.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{TurnNumber: 1, ActionType: "draw", Detail: "test"},
		})

		p := new(FortyAndEightWebPresenter)
		result := p.ActionLogOutput(fg)
		assert.Contains(t, result, "draw")
	})
}
