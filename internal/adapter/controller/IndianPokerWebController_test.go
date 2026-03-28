//go:build test

package controller_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newIndianPokerTestController(mi *mockUsecase.MockIndianPokerInteractor) *controller.IndianPokerWebController {
	iwc := controller.NewIndianPokerWebController(func() usecase.IndianPokerInteractorIF {
		return mi
	})
	return iwc
}

func TestIndianPokerWebController_Reset(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()

	cfg := domain.DefaultIndianPokerConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1,"message":""}`)

	recorded := execRequest(t, iwc.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "test-session",
	})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_Reset_WithConfig(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()

	ante := 50
	bl := 1
	cpuMeta := true
	cfg := domain.DefaultIndianPokerConfig()
	cfg.Ante = ante
	cfg.BettingLimit = domain.BettingLimitPotLimit
	cfg.CpuMetaAI = cpuMeta
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1,"message":""}`)

	recorded := execRequest(t, iwc.Exec, map[string]interface{}{
		"command":      "reset",
		"sessionId":    "test-session",
		"ante":         ante,
		"bettingLimit": bl,
		"cpuMetaAI":    cpuMeta,
	})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_Fold(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()

	mi.On("Action", domain.IndianPokerActionFold, 0, 0).Return(`{"phase":1}`)

	recorded := execRequest(t, iwc.Exec, map[string]interface{}{
		"command":   "fold",
		"sessionId": "test-session",
	})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_Check(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()
	mi.On("Action", domain.IndianPokerActionCheck, 0, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, iwc.Exec, map[string]interface{}{"command": "check", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_Call(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()
	mi.On("Action", domain.IndianPokerActionCall, 0, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, iwc.Exec, map[string]interface{}{"command": "call", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_Bet(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()
	mi.On("Action", domain.IndianPokerActionBet, 50, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, iwc.Exec, map[string]interface{}{"command": "bet", "sessionId": "s1", "amount": 50})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_Raise(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()
	mi.On("Action", domain.IndianPokerActionRaise, 30, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, iwc.Exec, map[string]interface{}{"command": "raise", "sessionId": "s1", "amount": 30})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_AllIn(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()
	mi.On("Action", domain.IndianPokerActionAllIn, 0, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, iwc.Exec, map[string]interface{}{"command": "allin", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_Log(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()

	mockLogOutput := `{"entries":[]}`
	mi.On("ActionLog").Return(mockLogOutput)

	t.Run("log command", func(t *testing.T) {
		recorded := execRequest(t, iwc.Exec, map[string]interface{}{
			"command":   "log",
			"sessionId": "ip-log-1",
		})
		recorded.CodeIs(200)
		recorded.ContentTypeIsJson()
		mi.AssertCalled(t, "ActionLog")
	})

	t.Run("l shorthand", func(t *testing.T) {
		recorded := execRequest(t, iwc.Exec, map[string]interface{}{
			"command":   "l",
			"sessionId": "ip-log-1",
		})
		recorded.CodeIs(200)
		recorded.ContentTypeIsJson()
		mi.AssertCalled(t, "ActionLog")
	})
}

func TestIndianPokerWebController_Quit(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()
	recorded := execRequest(t, iwc.Exec, map[string]interface{}{"command": "quit", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_QuitShort(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()
	recorded := execRequest(t, iwc.Exec, map[string]interface{}{"command": "q", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_Unknown(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()
	// Must have a session first
	cfg := domain.DefaultIndianPokerConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)
	execRequest(t, iwc.Exec,
		map[string]interface{}{"command": "reset", "sessionId": "s1"})

	recorded := execRequest(t, iwc.Exec, map[string]interface{}{"command": "xyz", "sessionId": "s1"})
	recorded.CodeIs(400)
}

func TestIndianPokerWebController_BadRequest_EmptyBody(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()
	recorded := execRequest(t, iwc.Exec, nil)
	recorded.CodeIs(400)
}

func TestIndianPokerWebController_BadRequest_NoCommand(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()
	recorded := execRequest(t, iwc.Exec, map[string]interface{}{"sessionId": "s1"})
	recorded.CodeIs(400)
}

func TestIndianPokerWebController_BadRequest_NoSession(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()
	recorded := execRequest(t, iwc.Exec, map[string]interface{}{"command": "reset"})
	recorded.CodeIs(400)
}

func TestIndianPokerWebController_ShortCommands(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()

	cfg := domain.DefaultIndianPokerConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)
	mi.On("Action", domain.IndianPokerActionFold, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.IndianPokerActionCheck, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.IndianPokerActionCall, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.IndianPokerActionBet, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.IndianPokerActionRaise, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.IndianPokerActionAllIn, 0, 0).Return(`{"phase":1}`)

	commands := []string{"r", "f", "ck", "c", "b", "ra", "a"}
	for _, cmd := range commands {
		recorded := execRequest(t, iwc.Exec, map[string]interface{}{"command": cmd, "sessionId": "s-short"})
		recorded.CodeIs(200)
	}
}

func TestIndianPokerWebController_LongSessionId(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()
	recorded := execRequest(t, iwc.Exec,
		map[string]interface{}{
			"command":   "reset",
			"sessionId": strings.Repeat("a", controller.SessionMaxIDLen+1),
		})
	recorded.CodeIs(400)
}

func TestIndianPokerWebController_Reset_AnteOnly(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()

	ante := 50
	cfg := domain.DefaultIndianPokerConfig()
	cfg.Ante = ante
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, iwc.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "s-ante",
		"ante":      ante,
	})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_Reset_WithBettingLimit_Valid(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()

	cfg := domain.DefaultIndianPokerConfig()
	cfg.BettingLimit = domain.BettingLimitPotLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, iwc.Exec, map[string]interface{}{
		"command":      "reset",
		"sessionId":    "s1",
		"bettingLimit": 1,
	})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_Reset_WithBettingLimit_AboveMax(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()

	cfg := domain.DefaultIndianPokerConfig()
	cfg.BettingLimit = domain.BettingLimitNoLimit // clamped from 5 to 2
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, iwc.Exec, map[string]interface{}{
		"command":      "reset",
		"sessionId":    "s1",
		"bettingLimit": 5,
	})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_Reset_WithBettingLimit_BelowMin(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()

	cfg := domain.DefaultIndianPokerConfig()
	cfg.BettingLimit = domain.BettingLimitFixed // clamped from -1 to 0
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, iwc.Exec, map[string]interface{}{
		"command":      "reset",
		"sessionId":    "s1",
		"bettingLimit": -1,
	})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_Reset_InvalidAnte(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()

	// ante=0 should be ignored (below threshold), use default
	cfg := domain.DefaultIndianPokerConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, iwc.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "s-inv",
		"ante":      0,
	})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_Reset_CpuMetaAI(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()

	cfg := domain.DefaultIndianPokerConfig()
	cfg.CpuMetaAI = false
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	metaAI := false
	recorded := execRequest(t, iwc.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "s-meta",
		"cpuMetaAI": metaAI,
	})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_Stop(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	c := controller.NewIndianPokerWebController(func() usecase.IndianPokerInteractorIF {
		return mi
	})
	c.Stop()
	c.Stop()
}

func TestIndianPokerWebController_BetWithHumanPlayMs(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()
	mi.On("Action", domain.IndianPokerActionBet, 100, 500).Return(`{"phase":1}`)
	recorded := execRequest(t, iwc.Exec, map[string]interface{}{"command": "bet", "sessionId": "s1", "amount": 100, "humanPlayMs": 500})
	recorded.CodeIs(200)
}

func TestIndianPokerWebController_FoldWithHumanPlayMs(t *testing.T) {
	mi := new(mockUsecase.MockIndianPokerInteractor)
	iwc := newIndianPokerTestController(mi)
	defer iwc.Stop()
	mi.On("Action", domain.IndianPokerActionFold, 0, 300).Return(`{"phase":1}`)
	recorded := execRequest(t, iwc.Exec, map[string]interface{}{"command": "fold", "sessionId": "s1", "humanPlayMs": 300})
	recorded.CodeIs(200)
}
