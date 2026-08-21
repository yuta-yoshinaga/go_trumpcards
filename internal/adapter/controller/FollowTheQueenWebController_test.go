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

func newFollowTheQueenTestController(mi *mockUsecase.MockFollowTheQueenInteractor) *controller.FollowTheQueenWebController {
	return controller.NewFollowTheQueenWebController(func() usecase.FollowTheQueenInteractorIF {
		return mi
	})
}

func TestFollowTheQueenWebController_Reset(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	cfg := domain.DefaultFollowTheQueenConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1,"message":""}`)

	recorded := execRequest(t, c.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "test-session",
	})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_Reset_WithConfig(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	ante := 5
	bringIn := 10
	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.Ante = ante
	cfg.BringIn = bringIn
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1,"message":""}`)

	recorded := execRequest(t, c.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "test-session",
		"ante":      ante,
		"bringIn":   bringIn,
	})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_Fold(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	mi.On("Action", domain.FollowTheQueenActionFold, 0, 0).Return(`{"phase":1}`)

	recorded := execRequest(t, c.Exec, map[string]interface{}{
		"command":   "fold",
		"sessionId": "test-session",
	})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_Check(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("Action", domain.FollowTheQueenActionCheck, 0, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "check", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_Call(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("Action", domain.FollowTheQueenActionCall, 0, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "call", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_Bet(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("Action", domain.FollowTheQueenActionBet, 50, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "bet", "sessionId": "s1", "amount": 50})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_Raise(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("Action", domain.FollowTheQueenActionRaise, 30, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "raise", "sessionId": "s1", "amount": 30})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_AllIn(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("Action", domain.FollowTheQueenActionAllIn, 0, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "allin", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_Quit(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "quit", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_QuitShort(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "q", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_Unknown(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	cfg := domain.DefaultFollowTheQueenConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)
	execRequest(t, c.Exec, map[string]interface{}{"command": "reset", "sessionId": "s1"})

	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "xyz", "sessionId": "s1"})
	recorded.CodeIs(400)
}

func TestFollowTheQueenWebController_BadRequest_EmptyBody(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	recorded := execRequest(t, c.Exec, nil)
	recorded.CodeIs(400)
}

func TestFollowTheQueenWebController_BadRequest_NoCommand(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	recorded := execRequest(t, c.Exec, map[string]interface{}{"sessionId": "s1"})
	recorded.CodeIs(400)
}

func TestFollowTheQueenWebController_BadRequest_NoSession(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "reset"})
	recorded.CodeIs(400)
}

func TestFollowTheQueenWebController_ShortCommands(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	cfg := domain.DefaultFollowTheQueenConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)
	mi.On("Action", domain.FollowTheQueenActionFold, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.FollowTheQueenActionCheck, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.FollowTheQueenActionCall, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.FollowTheQueenActionBet, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.FollowTheQueenActionRaise, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.FollowTheQueenActionAllIn, 0, 0).Return(`{"phase":1}`)

	commands := []string{"r", "f", "ck", "c", "b", "ra", "a"}
	for _, cmd := range commands {
		recorded := execRequest(t, c.Exec, map[string]interface{}{"command": cmd, "sessionId": "s-short"})
		recorded.CodeIs(200)
	}
}

func TestFollowTheQueenWebController_LongSessionId(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	recorded := execRequest(t, c.Exec,
		map[string]interface{}{
			"command":   "reset",
			"sessionId": strings.Repeat("a", controller.SessionMaxIDLen+1),
		})
	recorded.CodeIs(400)
}

func TestFollowTheQueenWebController_Reset_WithTableSize_Valid(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.TableSize = 4
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, c.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "s1",
		"tableSize": 4,
	})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_Reset_WithTableSize_Invalid(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	recorded := execRequest(t, c.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "s1",
		"tableSize": 8,
	})
	recorded.CodeIs(400)
}

func TestFollowTheQueenWebController_Reset_WithTableSize_Invalid_1(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	recorded := execRequest(t, c.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "s1",
		"tableSize": 1,
	})
	recorded.CodeIs(400)
}

func TestFollowTheQueenWebController_Reset_WithBettingLimit_Valid(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.BettingLimit = domain.BettingLimitPotLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, c.Exec, map[string]interface{}{
		"command":      "reset",
		"sessionId":    "s1",
		"bettingLimit": 1,
	})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_Reset_WithBettingLimit_Clamped(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.BettingLimit = domain.BettingLimitNoLimit // clamped from 5 to 2
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, c.Exec, map[string]interface{}{
		"command":      "reset",
		"sessionId":    "s1",
		"bettingLimit": 5,
	})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_Reset_WithBettingLimit_BelowMin(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.BettingLimit = domain.BettingLimitFixed // clamped from -1 to 0
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, c.Exec, map[string]interface{}{
		"command":      "reset",
		"sessionId":    "s1",
		"bettingLimit": -1,
	})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_Reset_TournamentMode(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.TournamentMode = true
	cfg.AnteLevelHands = 5
	cfg.AnteMultiplier = 200
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, c.Exec, map[string]interface{}{
		"command":        "reset",
		"sessionId":      "s-tm",
		"tournamentMode": true,
		"anteLevelHands": 5,
		"anteMultiplier": 200,
	})
	recorded.CodeIs(200)
}

// --- rebuy / addon commands ---

func TestFollowTheQueenWebController_Rebuy(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("Rebuy").Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "rebuy", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_RebuyShort(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("Rebuy").Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "rb", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_SkipRebuy(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("SkipRebuy").Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "skiprebuy", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_SkipRebuyShort(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("SkipRebuy").Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "sr", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_Addon(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("Addon").Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "addon", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_AddonShort(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("Addon").Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "ad", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_SkipAddon(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("SkipAddon").Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "skipaddon", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_SkipAddonShort(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("SkipAddon").Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "sa", "sessionId": "s1"})
	recorded.CodeIs(200)
}

// --- muck / show commands ---

func TestFollowTheQueenWebController_Muck(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("Muck").Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "muck", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_MuckShort(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("Muck").Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "m", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_ShowHand(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("ShowHand").Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "show", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_ShowHandShort(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()
	mi.On("ShowHand").Return(`{"phase":1}`)
	recorded := execRequest(t, c.Exec, map[string]interface{}{"command": "sh", "sessionId": "s1"})
	recorded.CodeIs(200)
}

// --- rebuy/addon config ---

func TestFollowTheQueenWebController_Reset_WithRebuyConfig(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.RebuyEnabled = true
	cfg.RebuyMaxCount = 5
	cfg.RebuyChips = 2000
	cfg.RebuyPeriodHands = 30
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, c.Exec, map[string]interface{}{
		"command":          "reset",
		"sessionId":        "s-rebuy",
		"rebuyEnabled":     true,
		"rebuyMaxCount":    5,
		"rebuyChips":       2000,
		"rebuyPeriodHands": 30,
	})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_Reset_WithAddonConfig(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	cfg := domain.DefaultFollowTheQueenConfig()
	cfg.AddonEnabled = true
	cfg.AddonChips = 3000
	cfg.AddonAfterHand = 25
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, c.Exec, map[string]interface{}{
		"command":        "reset",
		"sessionId":      "s-addon",
		"addonEnabled":   true,
		"addonChips":     3000,
		"addonAfterHand": 25,
	})
	recorded.CodeIs(200)
}

func TestFollowTheQueenWebController_Stop(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := controller.NewFollowTheQueenWebController(func() usecase.FollowTheQueenInteractorIF {
		return mi
	})
	c.Stop()
	c.Stop()
}

func TestFollowTheQueenWebController_Log(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	mockLogOutput := `{"entries":[]}`
	mi.On("ActionLog").Return(mockLogOutput)

	t.Run("log command", func(t *testing.T) {
		recorded := execRequest(t, c.Exec, map[string]interface{}{
			"command":   "log",
			"sessionId": "scs-log-1",
		})
		recorded.CodeIs(200)
		recorded.ContentTypeIsJson()
		mi.AssertCalled(t, "ActionLog")
	})

	t.Run("l shorthand", func(t *testing.T) {
		recorded := execRequest(t, c.Exec, map[string]interface{}{
			"command":   "l",
			"sessionId": "scs-log-1",
		})
		recorded.CodeIs(200)
		recorded.ContentTypeIsJson()
		mi.AssertCalled(t, "ActionLog")
	})
}

func TestFollowTheQueenWebController_Reset_InvalidValues(t *testing.T) {
	mi := new(mockUsecase.MockFollowTheQueenInteractor)
	c := newFollowTheQueenTestController(mi)
	defer c.Stop()

	// Values below threshold should be ignored (use defaults)
	cfg := domain.DefaultFollowTheQueenConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, c.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "s-inv",
		"ante":      0,
		"bringIn":   0,
		"smallBet":  0,
		"bigBet":    0,
	})
	recorded.CodeIs(200)
}
