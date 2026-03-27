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

func newOmahaTestController(mi *mockUsecase.MockOmahaInteractor) *controller.OmahaWebController {
	owc := controller.NewOmahaWebController(func() usecase.OmahaInteractorIF {
		return mi
	})
	return owc
}

func TestOmahaWebController_Reset(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	cfg := domain.DefaultOmahaConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1,"message":""}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "test-session",
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Reset_WithConfig(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	sb := 10
	bb := 20
	cfg := domain.DefaultOmahaConfig()
	cfg.SmallBlind = sb
	cfg.BigBlind = bb
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1,"message":""}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":    "reset",
		"sessionId":  "test-session",
		"smallBlind": sb,
		"bigBlind":   bb,
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Fold(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	mi.On("Action", domain.OmahaActionFold, 0, 0).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":   "fold",
		"sessionId": "test-session",
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Check(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("Action", domain.OmahaActionCheck, 0, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "check", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Call(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("Action", domain.OmahaActionCall, 0, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "call", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Bet(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("Action", domain.OmahaActionBet, 50, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "bet", "sessionId": "s1", "amount": 50})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Raise(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("Action", domain.OmahaActionRaise, 30, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "raise", "sessionId": "s1", "amount": 30})
	recorded.CodeIs(200)
}

func TestOmahaWebController_AllIn(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("Action", domain.OmahaActionAllIn, 0, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "allin", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Quit(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "quit", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_QuitShort(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "q", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Unknown(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	// Must have a session first
	cfg := domain.DefaultOmahaConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)
	execRequest(t, owc.Exec,
		map[string]interface{}{"command": "reset", "sessionId": "s1"})

	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "xyz", "sessionId": "s1"})
	recorded.CodeIs(400)
}

func TestOmahaWebController_BadRequest_EmptyBody(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	recorded := execRequest(t, owc.Exec, nil)
	recorded.CodeIs(400)
}

func TestOmahaWebController_BadRequest_NoCommand(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"sessionId": "s1"})
	recorded.CodeIs(400)
}

func TestOmahaWebController_BadRequest_NoSession(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "reset"})
	recorded.CodeIs(400)
}

func TestOmahaWebController_ShortCommands(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	cfg := domain.DefaultOmahaConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)
	mi.On("Action", domain.OmahaActionFold, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.OmahaActionCheck, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.OmahaActionCall, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.OmahaActionBet, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.OmahaActionRaise, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.OmahaActionAllIn, 0, 0).Return(`{"phase":1}`)

	commands := []string{"r", "f", "ck", "c", "b", "ra", "a"}
	for _, cmd := range commands {
		recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": cmd, "sessionId": "s-short"})
		recorded.CodeIs(200)
	}
}

func TestOmahaWebController_LongSessionId(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	recorded := execRequest(t, owc.Exec,
		map[string]interface{}{
			"command":   "reset",
			"sessionId": strings.Repeat("a", controller.SessionMaxIDLen+1),
		})
	recorded.CodeIs(400)
}

func TestOmahaWebController_Reset_SmallBlindOnly(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	sb := 3
	cfg := domain.DefaultOmahaConfig()
	cfg.SmallBlind = sb
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":    "reset",
		"sessionId":  "s-sb",
		"smallBlind": sb,
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Reset_BigBlindOnly(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	bb := 30
	cfg := domain.DefaultOmahaConfig()
	cfg.SmallBlind = bb / 2 // 自動調整: 15
	cfg.BigBlind = bb
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "s-bb",
		"bigBlind":  bb,
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Reset_SmallBlindOnly_AutoAdjust(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	// smallBlind=20のみ指定: bigBlindがデフォルト(10)より大きいので自動調整 bb=40
	sb := 20
	cfg := domain.DefaultOmahaConfig()
	cfg.SmallBlind = sb
	cfg.BigBlind = sb * 2 // 自動調整: 40
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":    "reset",
		"sessionId":  "s-sb-auto",
		"smallBlind": sb,
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Reset_BigBlindOnly_AutoAdjust(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	// bigBlind=4のみ指定: bb>1なので自動調整 sb=2
	bb := 4
	cfg := domain.DefaultOmahaConfig()
	cfg.SmallBlind = bb / 2 // 自動調整: 2
	cfg.BigBlind = bb
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "s-bb-auto",
		"bigBlind":  bb,
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Reset_SmallBlindGeBigBlind(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	// smallBlind >= bigBlind は不正
	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":    "reset",
		"sessionId":  "s-ge",
		"smallBlind": 20,
		"bigBlind":   10,
	})
	recorded.CodeIs(400)

	// smallBlind == bigBlind も不正
	recorded2 := execRequest(t, owc.Exec, map[string]interface{}{
		"command":    "reset",
		"sessionId":  "s-eq",
		"smallBlind": 10,
		"bigBlind":   10,
	})
	recorded2.CodeIs(400)
}

func TestOmahaWebController_Reset_InvalidBlinds(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	cfg := domain.DefaultOmahaConfig() // defaults unchanged when values < 1
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":    "reset",
		"sessionId":  "s-inv",
		"smallBlind": 0,
		"bigBlind":   0,
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Reset_TournamentMode(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	tm := true
	blh := 5
	blm := 200
	cfg := domain.DefaultOmahaConfig()
	cfg.TournamentMode = true
	cfg.BlindLevelHands = 5
	cfg.BlindMultiplier = 200
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":         "reset",
		"sessionId":       "s-tm",
		"tournamentMode":  tm,
		"blindLevelHands": blh,
		"blindMultiplier": blm,
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Reset_TournamentMode_InvalidValues(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	// blindLevelHands=0 and blindMultiplier=100 should be ignored (below threshold)
	cfg := domain.DefaultOmahaConfig()
	cfg.TournamentMode = true
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":         "reset",
		"sessionId":       "s-tm-inv",
		"tournamentMode":  true,
		"blindLevelHands": 0,
		"blindMultiplier": 100,
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Reset_TournamentMode_Nil(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	// No tournament params → default config
	cfg := domain.DefaultOmahaConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "s-tm-nil",
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Stop(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	c := controller.NewOmahaWebController(func() usecase.OmahaInteractorIF {
		return mi
	})
	c.Stop()
	c.Stop()
}

// --- betting limit ---

func TestOmahaWebController_Reset_WithBettingLimit_Valid(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	cfg := domain.DefaultOmahaConfig()
	cfg.BettingLimit = domain.BettingLimitPotLimit
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":      "reset",
		"sessionId":    "s1",
		"bettingLimit": 1,
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Reset_WithBettingLimit_AboveMax(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	cfg := domain.DefaultOmahaConfig()
	cfg.BettingLimit = domain.BettingLimitNoLimit // clamped from 5 to 2
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":      "reset",
		"sessionId":    "s1",
		"bettingLimit": 5,
	})
	recorded.CodeIs(200)
}

// --- table size ---

func TestOmahaWebController_Reset_WithTableSize_Valid(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	cfg := domain.DefaultOmahaConfig()
	cfg.TableSize = domain.HoldemTableSize6
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "s1",
		"tableSize": 6,
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Reset_WithTableSize_9max(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	cfg := domain.DefaultOmahaConfig()
	cfg.TableSize = domain.HoldemTableSize9
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "s1",
		"tableSize": 9,
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Reset_WithTableSize_Invalid(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "s1",
		"tableSize": 5,
	})
	recorded.CodeIs(400)
}

func TestOmahaWebController_Reset_WithTableSize_Invalid_3(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "s1",
		"tableSize": 3,
	})
	recorded.CodeIs(400)
}

// --- rebuy / addon commands ---

func TestOmahaWebController_Rebuy(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("Rebuy").Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "rebuy", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_RebuyShort(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("Rebuy").Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "rb", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_SkipRebuy(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("SkipRebuy").Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "skiprebuy", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_SkipRebuyShort(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("SkipRebuy").Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "sr", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Addon(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("Addon").Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "addon", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_AddonShort(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("Addon").Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "ad", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_SkipAddon(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("SkipAddon").Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "skipaddon", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_SkipAddonShort(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("SkipAddon").Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "sa", "sessionId": "s1"})
	recorded.CodeIs(200)
}

// --- reset with rebuy/addon config ---

func TestOmahaWebController_Reset_WithRebuyConfig(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	cfg := domain.DefaultOmahaConfig()
	cfg.RebuyEnabled = true
	cfg.RebuyMaxCount = 5
	cfg.RebuyChips = 2000
	cfg.RebuyPeriodHands = 30
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":          "reset",
		"sessionId":        "s-rebuy",
		"rebuyEnabled":     true,
		"rebuyMaxCount":    5,
		"rebuyChips":       2000,
		"rebuyPeriodHands": 30,
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Reset_WithAddonConfig(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	cfg := domain.DefaultOmahaConfig()
	cfg.AddonEnabled = true
	cfg.AddonChips = 3000
	cfg.AddonAfterHand = 25
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":        "reset",
		"sessionId":      "s-addon",
		"addonEnabled":   true,
		"addonChips":     3000,
		"addonAfterHand": 25,
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Reset_WithRebuyAddonConfig_InvalidValues(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	// Values below threshold should be ignored (use defaults)
	cfg := domain.DefaultOmahaConfig()
	cfg.RebuyEnabled = true
	cfg.AddonEnabled = true
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":          "reset",
		"sessionId":        "s-inv-ra",
		"rebuyEnabled":     true,
		"rebuyMaxCount":    0,
		"rebuyChips":       0,
		"rebuyPeriodHands": 0,
		"addonEnabled":     true,
		"addonChips":       0,
		"addonAfterHand":   0,
	})
	recorded.CodeIs(200)
}

// --- muck / show commands ---

func TestOmahaWebController_Muck(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("Muck").Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "muck", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_MuckShort(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("Muck").Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "m", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_ShowHand(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("ShowHand").Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "show", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_ShowHandShort(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()
	mi.On("ShowHand").Return(`{"phase":1}`)
	recorded := execRequest(t, owc.Exec, map[string]interface{}{"command": "sh", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Reset_WithBettingLimit_BelowMin(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	cfg := domain.DefaultOmahaConfig()
	cfg.BettingLimit = domain.BettingLimitFixed // clamped from -1 to 0
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, owc.Exec, map[string]interface{}{
		"command":      "reset",
		"sessionId":    "s1",
		"bettingLimit": -1,
	})
	recorded.CodeIs(200)
}

func TestOmahaWebController_Log(t *testing.T) {
	mi := new(mockUsecase.MockOmahaInteractor)
	owc := newOmahaTestController(mi)
	defer owc.Stop()

	mockLogOutput := `{"entries":[]}`
	mi.On("ActionLog").Return(mockLogOutput)

	t.Run("log command", func(t *testing.T) {
		recorded := execRequest(t, owc.Exec, map[string]interface{}{
			"command":   "log",
			"sessionId": "oh-log-1",
		})
		recorded.CodeIs(200)
		recorded.ContentTypeIsJson()
		mi.AssertCalled(t, "ActionLog")
	})

	t.Run("l shorthand", func(t *testing.T) {
		recorded := execRequest(t, owc.Exec, map[string]interface{}{
			"command":   "l",
			"sessionId": "oh-log-1",
		})
		recorded.CodeIs(200)
		recorded.ContentTypeIsJson()
		mi.AssertCalled(t, "ActionLog")
	})
}
