package controller

import (
	"strings"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/stretchr/testify/mock"

	mockUsecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newPokerTestHandler(mi *mockUsecase.MockPokerInteractor) (*rest.Api, *PokerWebController) {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	pwc := NewPokerWebController(func() usecase.PokerInteractorIF {
		return mi
	})
	router, _ := rest.MakeRouter(
		rest.Post("/poker/exec", pwc.Exec),
	)
	api.SetApp(router)
	return api, pwc
}

// --- reset command ---

func TestPokerWebController_Reset_Default(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	cfg := domain.DefaultPokerConfig()
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1,"message":""}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "reset",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Reset_ShortCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	cfg := domain.DefaultPokerConfig()
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1,"message":""}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "r",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Reset_WithCpuCount_Valid(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	cfg := domain.DefaultPokerConfig()
	cfg.CpuCount = 2
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "reset",
				"sessionId": "s1",
				"cpuCount":  2,
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Reset_WithCpuCount_BelowMin(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	cfg := domain.DefaultPokerConfig()
	cfg.CpuCount = 1 // clamped from 0 to 1
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "reset",
				"sessionId": "s1",
				"cpuCount":  0,
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Reset_WithCpuCount_AboveMax(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	cfg := domain.DefaultPokerConfig()
	cfg.CpuCount = 3 // clamped from 5 to 3
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "reset",
				"sessionId": "s1",
				"cpuCount":  5,
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Reset_WithJokerCount_Valid(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	cfg := domain.DefaultPokerConfig()
	cfg.JokerCount = 1
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":    "reset",
				"sessionId":  "s1",
				"jokerCount": 1,
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Reset_WithJokerCount_BelowMin(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	cfg := domain.DefaultPokerConfig()
	cfg.JokerCount = 0 // clamped from -1 to 0
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":    "reset",
				"sessionId":  "s1",
				"jokerCount": -1,
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Reset_WithJokerCount_AboveMax(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	cfg := domain.DefaultPokerConfig()
	cfg.JokerCount = 2 // clamped from 5 to 2
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":    "reset",
				"sessionId":  "s1",
				"jokerCount": 5,
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Reset_WithBothCpuAndJoker(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	cfg := domain.DefaultPokerConfig()
	cfg.CpuCount = 2
	cfg.JokerCount = 2
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":    "reset",
				"sessionId":  "s1",
				"cpuCount":   2,
				"jokerCount": 2,
			}))
	recorded.CodeIs(200)
}

// --- exchange command ---

func TestPokerWebController_Exchange_WithIndices(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Exchange", []int{0, 2}).Return(`{"phase":2}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "exchange",
				"sessionId": "s1",
				"indices":   []int{0, 2},
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Exchange_ShortCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Exchange", []int{1}).Return(`{"phase":2}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "e",
				"sessionId": "s1",
				"indices":   []int{1},
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Exchange_NilIndices(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Exchange", []int{}).Return(`{"phase":2}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "exchange",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

// --- stand command ---

func TestPokerWebController_Stand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Stand").Return(`{"phase":3}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "stand",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Stand_ShortCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Stand").Return(`{"phase":3}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "s",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

// --- fold command ---

func TestPokerWebController_Fold(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Action", domain.PokerActionFold, 0).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "fold",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Fold_ShortCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Action", domain.PokerActionFold, 0).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "f",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

// --- check command ---

func TestPokerWebController_Check(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Action", domain.PokerActionCheck, 0).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "check",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Check_ShortCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Action", domain.PokerActionCheck, 0).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "ck",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

// --- call command ---

func TestPokerWebController_Call(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Action", domain.PokerActionCall, 0).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "call",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Call_ShortCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Action", domain.PokerActionCall, 0).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "c",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

// --- bet command ---

func TestPokerWebController_Bet_ValidAmount(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Action", domain.PokerActionBet, 20).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "bet",
				"sessionId": "s1",
				"amount":    20,
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Bet_ShortCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Action", domain.PokerActionBet, 10).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "b",
				"sessionId": "s1",
				"amount":    10,
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Bet_AmountZero(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	// amount=0 passes through to domain which validates
	mi.On("Action", domain.PokerActionBet, 0).Return(`{"phase":1,"message":"error"}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "bet",
				"sessionId": "s1",
				"amount":    0,
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Bet_AmountMissing(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	// No amount field => defaults to 0, passes through to domain
	mi.On("Action", domain.PokerActionBet, 0).Return(`{"phase":1,"message":"error"}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "b",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

// --- raise command ---

func TestPokerWebController_Raise_ValidAmount(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Action", domain.PokerActionRaise, 30).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "raise",
				"sessionId": "s1",
				"amount":    30,
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Raise_ShortCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Action", domain.PokerActionRaise, 10).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "ra",
				"sessionId": "s1",
				"amount":    10,
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Raise_AmountZero(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	// amount=0 passes through to domain which validates
	mi.On("Action", domain.PokerActionRaise, 0).Return(`{"phase":1,"message":"error"}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "ra",
				"sessionId": "s1",
				"amount":    0,
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Raise_AmountMissing(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	// No amount => defaults to 0, passes through to domain
	mi.On("Action", domain.PokerActionRaise, 0).Return(`{"phase":1,"message":"error"}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "raise",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

// --- allin command ---

func TestPokerWebController_AllIn(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Action", domain.PokerActionAllIn, 0).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "allin",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_AllIn_ShortCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Action", domain.PokerActionAllIn, 0).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "a",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

// --- odds command ---

func TestPokerWebController_Odds(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Odds", []int{0, 2}).Return(`{"phase":2,"odds":[]}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "odds",
				"sessionId": "s1",
				"indices":   []int{0, 2},
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Odds_ShortCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Odds", []int{1}).Return(`{"phase":2,"odds":[]}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "o",
				"sessionId": "s1",
				"indices":   []int{1},
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Odds_NilIndices(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	mi.On("Odds", []int{}).Return(`{"phase":2}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "odds",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

// --- quit command ---

func TestPokerWebController_Quit(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "quit",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Quit_ShortCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "q",
				"sessionId": "s1",
			}))
	recorded.CodeIs(200)
}

// --- unknown command ---

func TestPokerWebController_UnknownCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	// Must create session first
	cfg := domain.DefaultPokerConfig()
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)
	test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{"command": "reset", "sessionId": "s1"}))

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "xyz",
				"sessionId": "s1",
			}))
	recorded.CodeIs(400)
}

// --- error cases ---

func TestPokerWebController_EmptyBody(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec", nil))
	recorded.CodeIs(400)
}

func TestPokerWebController_NoCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"sessionId": "s1",
			}))
	recorded.CodeIs(400)
}

func TestPokerWebController_EmptyCommand(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "",
				"sessionId": "s1",
			}))
	recorded.CodeIs(400)
}

func TestPokerWebController_NoSessionId(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command": "reset",
			}))
	recorded.CodeIs(400)
}

func TestPokerWebController_EmptySessionId(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "reset",
				"sessionId": "",
			}))
	recorded.CodeIs(400)
}

func TestPokerWebController_LongSessionId(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":   "reset",
				"sessionId": strings.Repeat("a", SessionMaxIDLen+1),
			}))
	recorded.CodeIs(400)
}

// --- Stop method ---

func TestPokerWebController_Stop(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	c := NewPokerWebController(func() usecase.PokerInteractorIF {
		return mi
	})
	c.Stop()
	c.Stop() // double stop should be safe
}

// --- session isolation ---

func TestPokerWebController_SessionIsolation(t *testing.T) {
	mockA := new(mockUsecase.MockPokerInteractor)
	mockB := new(mockUsecase.MockPokerInteractor)

	cfgA := domain.DefaultPokerConfig()
	cfgB := domain.DefaultPokerConfig()
	mockA.On("ResetWithConfig", cfgA).Return(`{"phase":1}`)
	mockB.On("ResetWithConfig", cfgB).Return(`{"phase":1}`)
	mockA.On("Stand").Return(`{"phase":3}`)

	callCount := 0
	pwc := NewPokerWebController(func() usecase.PokerInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer pwc.Stop()

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, _ := rest.MakeRouter(
		rest.Post("/poker/exec", pwc.Exec),
	)
	api.SetApp(router)

	// session-A creates mockA
	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{"command": "reset", "sessionId": "session-A"}))
	recorded.CodeIs(200)
	mockA.AssertCalled(t, "ResetWithConfig", cfgA)

	// session-B creates mockB
	recorded = test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{"command": "reset", "sessionId": "session-B"}))
	recorded.CodeIs(200)
	mockB.AssertCalled(t, "ResetWithConfig", cfgB)

	// session-A reuses mockA (no new factory call)
	recorded = test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{"command": "s", "sessionId": "session-A"}))
	recorded.CodeIs(200)
	mockA.AssertCalled(t, "Stand")
	if callCount != 2 {
		t.Errorf("expected factory to be called 2 times, got %d", callCount)
	}
}

// --- all short commands exercise ---

func TestPokerWebController_AllShortCommands(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	cfg := domain.DefaultPokerConfig()
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)
	mi.On("Exchange", mock.Anything).Return(`{"phase":2}`)
	mi.On("Stand").Return(`{"phase":3}`)
	mi.On("Action", domain.PokerActionFold, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.PokerActionCheck, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.PokerActionCall, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.PokerActionAllIn, 0).Return(`{"phase":1}`)
	mi.On("Odds", mock.Anything).Return(`{"phase":2}`)

	commands := []string{"r", "e", "s", "f", "ck", "c", "a", "o"}
	for _, cmd := range commands {
		recorded := test.RunRequest(t, api.MakeHandler(),
			test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
				map[string]interface{}{"command": cmd, "sessionId": "s-short"}))
		recorded.CodeIs(200)
	}
}

// --- betting limit ---

func TestPokerWebController_Reset_WithBettingLimit_Valid(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	cfg := domain.DefaultPokerConfig()
	cfg.BettingLimit = domain.BettingLimitPotLimit
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":      "reset",
				"sessionId":    "s1",
				"bettingLimit": 1,
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Reset_WithBettingLimit_AboveMax(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	cfg := domain.DefaultPokerConfig()
	cfg.BettingLimit = domain.BettingLimitNoLimit // clamped from 5 to 2
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":      "reset",
				"sessionId":    "s1",
				"bettingLimit": 5,
			}))
	recorded.CodeIs(200)
}

func TestPokerWebController_Reset_WithBettingLimit_BelowMin(t *testing.T) {
	mi := new(mockUsecase.MockPokerInteractor)
	api, pwc := newPokerTestHandler(mi)
	defer pwc.Stop()

	cfg := domain.DefaultPokerConfig()
	cfg.BettingLimit = domain.BettingLimitFixed // clamped from -1 to 0
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/poker/exec",
			map[string]interface{}{
				"command":      "reset",
				"sessionId":    "s1",
				"bettingLimit": -1,
			}))
	recorded.CodeIs(200)
}
