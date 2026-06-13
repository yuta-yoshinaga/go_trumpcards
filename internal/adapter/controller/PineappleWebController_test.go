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

func newPineappleTestController(mi *mockUsecase.MockPineappleInteractor) *controller.PineappleWebController {
	pwc := controller.NewPineappleWebController(func() usecase.PineappleInteractorIF {
		return mi
	})
	return pwc
}

func TestPineappleWebController_Reset(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()

	cfg := domain.DefaultPineappleConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1,"message":""}`)

	recorded := execRequest(t, pwc.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "test-session",
	})
	recorded.CodeIs(200)
}

func TestPineappleWebController_Reset_WithConfig(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()

	sb := 10
	bb := 20
	cfg := domain.DefaultPineappleConfig()
	cfg.SmallBlind = sb
	cfg.BigBlind = bb
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1,"message":""}`)

	recorded := execRequest(t, pwc.Exec, map[string]interface{}{
		"command":    "reset",
		"sessionId":  "test-session",
		"smallBlind": sb,
		"bigBlind":   bb,
	})
	recorded.CodeIs(200)
}

func TestPineappleWebController_Fold(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()

	mi.On("Action", domain.PineappleActionFold, 0, 0).Return(`{"phase":1}`)

	recorded := execRequest(t, pwc.Exec, map[string]interface{}{
		"command":   "fold",
		"sessionId": "test-session",
	})
	recorded.CodeIs(200)
}

func TestPineappleWebController_Check(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("Action", domain.PineappleActionCheck, 0, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "check", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestPineappleWebController_Call(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("Action", domain.PineappleActionCall, 0, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "call", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestPineappleWebController_Bet(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("Action", domain.PineappleActionBet, 50, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "bet", "sessionId": "s1", "amount": 50})
	recorded.CodeIs(200)
}

func TestPineappleWebController_Raise(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("Action", domain.PineappleActionRaise, 30, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "raise", "sessionId": "s1", "amount": 30})
	recorded.CodeIs(200)
}

func TestPineappleWebController_AllIn(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("Action", domain.PineappleActionAllIn, 0, 0).Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "allin", "sessionId": "s1"})
	recorded.CodeIs(200)
}

// --- discard (Pineapple-specific) ---

func TestPineappleWebController_Discard_ValidCardIdx(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()

	cardIdx := 2
	mi.On("Discard", cardIdx).Return(`{"phase":2}`)

	recorded := execRequest(t, pwc.Exec, map[string]interface{}{
		"command":   "discard",
		"sessionId": "s1",
		"cardIdx":   cardIdx,
	})
	recorded.CodeIs(200)
}

func TestPineappleWebController_Discard_MultipleCardIdxs(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()

	// Irish Poker: cardIdxs takes precedence and routes to DiscardMany.
	mi.On("DiscardMany", []int{1, 3}).Return(`{"phase":2}`)

	recorded := execRequest(t, pwc.Exec, map[string]interface{}{
		"command":   "discard",
		"sessionId": "s1",
		"cardIdxs":  []int{1, 3},
	})
	recorded.CodeIs(200)
	mi.AssertCalled(t, "DiscardMany", []int{1, 3})
}

func TestPineappleWebController_Discard_ShortCommand(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()

	cardIdx := 0
	mi.On("Discard", cardIdx).Return(`{"phase":2}`)

	recorded := execRequest(t, pwc.Exec, map[string]interface{}{
		"command":   "d",
		"sessionId": "s1",
		"cardIdx":   cardIdx,
	})
	recorded.CodeIs(200)
}

func TestPineappleWebController_Discard_NilCardIdx(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()

	// cardIdx is omitted → nil in parsed input → should return 400
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{
		"command":   "discard",
		"sessionId": "s1",
	})
	recorded.CodeIs(400)
}

func TestPineappleWebController_Quit(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "quit", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestPineappleWebController_QuitShort(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "q", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestPineappleWebController_Unknown(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	// Must have a session first
	cfg := domain.DefaultPineappleConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)
	execRequest(t, pwc.Exec,
		map[string]interface{}{"command": "reset", "sessionId": "s1"})

	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "xyz", "sessionId": "s1"})
	recorded.CodeIs(400)
}

func TestPineappleWebController_BadRequest_EmptyBody(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	recorded := execRequest(t, pwc.Exec, nil)
	recorded.CodeIs(400)
}

func TestPineappleWebController_BadRequest_NoCommand(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"sessionId": "s1"})
	recorded.CodeIs(400)
}

func TestPineappleWebController_BadRequest_NoSession(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "reset"})
	recorded.CodeIs(400)
}

func TestPineappleWebController_ShortCommands(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()

	cfg := domain.DefaultPineappleConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)
	mi.On("Action", domain.PineappleActionFold, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.PineappleActionCheck, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.PineappleActionCall, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.PineappleActionBet, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.PineappleActionRaise, 0, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.PineappleActionAllIn, 0, 0).Return(`{"phase":1}`)

	commands := []string{"r", "f", "ck", "c", "b", "ra", "a"}
	for _, cmd := range commands {
		recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": cmd, "sessionId": "s-short"})
		recorded.CodeIs(200)
	}
}

func TestPineappleWebController_LongSessionId(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	recorded := execRequest(t, pwc.Exec,
		map[string]interface{}{
			"command":   "reset",
			"sessionId": strings.Repeat("a", controller.SessionMaxIDLen+1),
		})
	recorded.CodeIs(400)
}

func TestPineappleWebController_Stop(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	c := controller.NewPineappleWebController(func() usecase.PineappleInteractorIF {
		return mi
	})
	c.Stop()
	c.Stop()
}

// --- rebuy / addon commands ---

func TestPineappleWebController_Rebuy(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("Rebuy").Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "rebuy", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestPineappleWebController_RebuyShort(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("Rebuy").Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "rb", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestPineappleWebController_SkipRebuy(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("SkipRebuy").Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "skiprebuy", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestPineappleWebController_SkipRebuyShort(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("SkipRebuy").Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "sr", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestPineappleWebController_Addon(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("Addon").Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "addon", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestPineappleWebController_AddonShort(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("Addon").Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "ad", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestPineappleWebController_SkipAddon(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("SkipAddon").Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "skipaddon", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestPineappleWebController_SkipAddonShort(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("SkipAddon").Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "sa", "sessionId": "s1"})
	recorded.CodeIs(200)
}

// --- muck / show commands ---

func TestPineappleWebController_Muck(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("Muck").Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "muck", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestPineappleWebController_MuckShort(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("Muck").Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "m", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestPineappleWebController_ShowHand(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("ShowHand").Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "show", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestPineappleWebController_ShowHandShort(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()
	mi.On("ShowHand").Return(`{"phase":1}`)
	recorded := execRequest(t, pwc.Exec, map[string]interface{}{"command": "sh", "sessionId": "s1"})
	recorded.CodeIs(200)
}

func TestPineappleWebController_Log(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()

	mockLogOutput := `{"entries":[]}`
	mi.On("ActionLog").Return(mockLogOutput)

	t.Run("log command", func(t *testing.T) {
		recorded := execRequest(t, pwc.Exec, map[string]interface{}{
			"command":   "log",
			"sessionId": "pi-log-1",
		})
		recorded.CodeIs(200)
		recorded.ContentTypeIsJson()
		mi.AssertCalled(t, "ActionLog")
	})

	t.Run("l shorthand", func(t *testing.T) {
		recorded := execRequest(t, pwc.Exec, map[string]interface{}{
			"command":   "l",
			"sessionId": "pi-log-1",
		})
		recorded.CodeIs(200)
		recorded.ContentTypeIsJson()
		mi.AssertCalled(t, "ActionLog")
	})
}

// --- reset with blind/tournament config ---

func TestPineappleWebController_Reset_SmallBlindGeBigBlind(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()

	recorded := execRequest(t, pwc.Exec, map[string]interface{}{
		"command":    "reset",
		"sessionId":  "s-ge",
		"smallBlind": 20,
		"bigBlind":   10,
	})
	recorded.CodeIs(400)
}

func TestPineappleWebController_Reset_TournamentMode(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()

	tm := true
	blh := 5
	blm := 200
	cfg := domain.DefaultPineappleConfig()
	cfg.TournamentMode = true
	cfg.BlindLevelHands = 5
	cfg.BlindMultiplier = 200
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	recorded := execRequest(t, pwc.Exec, map[string]interface{}{
		"command":         "reset",
		"sessionId":       "s-tm",
		"tournamentMode":  tm,
		"blindLevelHands": blh,
		"blindMultiplier": blm,
	})
	recorded.CodeIs(200)
}

func TestPineappleWebController_Reset_WithTableSize_Invalid(t *testing.T) {
	mi := new(mockUsecase.MockPineappleInteractor)
	pwc := newPineappleTestController(mi)
	defer pwc.Stop()

	recorded := execRequest(t, pwc.Exec, map[string]interface{}{
		"command":   "reset",
		"sessionId": "s1",
		"tableSize": 5,
	})
	recorded.CodeIs(400)
}
