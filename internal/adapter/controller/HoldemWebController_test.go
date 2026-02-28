package controller

import (
	"strings"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	mockUsecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newHoldemTestHandler(mi *mockUsecase.MockHoldemInteractor) *rest.Api {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	hwc := NewHoldemWebController(func() usecase.HoldemInteractorIF {
		return mi
	})
	router, _ := rest.MakeRouter(
		rest.Post("/holdem/exec", hwc.Exec),
	)
	api.SetApp(router)
	return api
}

func TestHoldemWebController_Reset(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)

	cfg := domain.DefaultHoldemConfig()
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1,"message":""}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{
				"command":   "reset",
				"sessionId": "test-session",
			}))
	recorded.CodeIs(200)
}

func TestHoldemWebController_Reset_WithConfig(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)

	sb := 10
	bb := 20
	cfg := domain.DefaultHoldemConfig()
	cfg.SmallBlind = sb
	cfg.BigBlind = bb
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1,"message":""}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{
				"command":    "reset",
				"sessionId":  "test-session",
				"smallBlind": sb,
				"bigBlind":   bb,
			}))
	recorded.CodeIs(200)
}

func TestHoldemWebController_Fold(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)

	mi.On("Action", domain.HoldemActionFold, 0).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{
				"command":   "fold",
				"sessionId": "test-session",
			}))
	recorded.CodeIs(200)
}

func TestHoldemWebController_Check(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)
	mi.On("Action", domain.HoldemActionCheck, 0).Return(`{"phase":1}`)
	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{"command": "check", "sessionId": "s1"}))
	recorded.CodeIs(200)
}

func TestHoldemWebController_Call(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)
	mi.On("Action", domain.HoldemActionCall, 0).Return(`{"phase":1}`)
	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{"command": "call", "sessionId": "s1"}))
	recorded.CodeIs(200)
}

func TestHoldemWebController_Bet(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)
	mi.On("Action", domain.HoldemActionBet, 50).Return(`{"phase":1}`)
	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{"command": "bet", "sessionId": "s1", "amount": 50}))
	recorded.CodeIs(200)
}

func TestHoldemWebController_Raise(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)
	mi.On("Action", domain.HoldemActionRaise, 30).Return(`{"phase":1}`)
	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{"command": "raise", "sessionId": "s1", "amount": 30}))
	recorded.CodeIs(200)
}

func TestHoldemWebController_AllIn(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)
	mi.On("Action", domain.HoldemActionAllIn, 0).Return(`{"phase":1}`)
	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{"command": "allin", "sessionId": "s1"}))
	recorded.CodeIs(200)
}

func TestHoldemWebController_Quit(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)
	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{"command": "quit", "sessionId": "s1"}))
	recorded.CodeIs(200)
}

func TestHoldemWebController_QuitShort(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)
	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{"command": "q", "sessionId": "s1"}))
	recorded.CodeIs(200)
}

func TestHoldemWebController_Unknown(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)
	// Must have a session first
	cfg := domain.DefaultHoldemConfig()
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)
	test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{"command": "reset", "sessionId": "s1"}))

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{"command": "xyz", "sessionId": "s1"}))
	recorded.CodeIs(200)
}

func TestHoldemWebController_BadRequest_EmptyBody(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)
	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec", nil))
	recorded.CodeIs(400)
}

func TestHoldemWebController_BadRequest_NoCommand(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)
	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{"sessionId": "s1"}))
	recorded.CodeIs(400)
}

func TestHoldemWebController_BadRequest_NoSession(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)
	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{"command": "reset"}))
	recorded.CodeIs(400)
}

func TestHoldemWebController_ShortCommands(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)

	cfg := domain.DefaultHoldemConfig()
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)
	mi.On("Action", domain.HoldemActionFold, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.HoldemActionCheck, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.HoldemActionCall, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.HoldemActionBet, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.HoldemActionRaise, 0).Return(`{"phase":1}`)
	mi.On("Action", domain.HoldemActionAllIn, 0).Return(`{"phase":1}`)

	commands := []string{"r", "f", "ck", "c", "b", "ra", "a"}
	for _, cmd := range commands {
		recorded := test.RunRequest(t, api.MakeHandler(),
			test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
				map[string]interface{}{"command": cmd, "sessionId": "s-short"}))
		recorded.CodeIs(200)
	}
}

func TestHoldemWebController_ErrorResponse(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)
	mi.On("Action", domain.HoldemActionFold, 0).Return("invalid json")
	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{"command": "fold", "sessionId": "s-err"}))
	recorded.CodeIs(400)
}

func TestHoldemWebController_LongSessionId(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)
	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{
				"command":   "reset",
				"sessionId": strings.Repeat("a", SessionMaxIDLen+1),
			}))
	recorded.CodeIs(400)
}

func TestHoldemWebController_Reset_SmallBlindOnly(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)

	sb := 3
	cfg := domain.DefaultHoldemConfig()
	cfg.SmallBlind = sb
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{
				"command":    "reset",
				"sessionId":  "s-sb",
				"smallBlind": sb,
			}))
	recorded.CodeIs(200)
}

func TestHoldemWebController_Reset_BigBlindOnly(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)

	bb := 30
	cfg := domain.DefaultHoldemConfig()
	cfg.BigBlind = bb
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{
				"command":   "reset",
				"sessionId": "s-bb",
				"bigBlind":  bb,
			}))
	recorded.CodeIs(200)
}

func TestHoldemWebController_Reset_SmallBlindGeBigBlind(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)

	// smallBlind >= bigBlind は不正
	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{
				"command":    "reset",
				"sessionId":  "s-ge",
				"smallBlind": 20,
				"bigBlind":   10,
			}))
	recorded.CodeIs(400)

	// smallBlind == bigBlind も不正
	recorded2 := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{
				"command":    "reset",
				"sessionId":  "s-eq",
				"smallBlind": 10,
				"bigBlind":   10,
			}))
	recorded2.CodeIs(400)
}

func TestHoldemWebController_Reset_InvalidBlinds(t *testing.T) {
	mi := new(mockUsecase.MockHoldemInteractor)
	api := newHoldemTestHandler(mi)

	cfg := domain.DefaultHoldemConfig() // defaults unchanged when values < 1
	mi.On("ResetWithConfig", cfg).Return(`{"phase":1}`)

	recorded := test.RunRequest(t, api.MakeHandler(),
		test.MakeSimpleRequest("POST", "http://localhost/holdem/exec",
			map[string]interface{}{
				"command":    "reset",
				"sessionId":  "s-inv",
				"smallBlind": 0,
				"bigBlind":   0,
			}))
	recorded.CodeIs(200)
}
