package controller_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/stretchr/testify/mock"
)

func TestBlackJackWebController_Method(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":1000},"player":{"score":0,"cards":null,"chips":1000},"message":"bye."}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("Reset").Return(mockOutput).Times(2)
	bjiMock.On("Hit").Return(mockOutput)
	bjiMock.On("Stand").Return(mockOutput)
	bjiMock.On("Bet", 100, 0, 0, 0).Return(mockOutput)
	bjiMock.On("DoubleDown").Return(mockOutput)
	bjiMock.On("Split").Return(mockOutput)
	bjiMock.On("Insurance").Return(mockOutput)
	bjiMock.On("DeclineInsurance").Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/blackjack/exec", tbc.Exec),
	)
	api.SetApp(router)
	var jsonCase1 controller.BlackJackWebInput
	t.Run("success Exec q", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "q", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("success Exec quit", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "quit", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("success Exec r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "r", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("success Exec reset", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("success Exec h", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "h", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("success Exec hit", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "hit", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("success Exec s", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "s", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("success Exec stand", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "stand", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("success Exec bet", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "b", "amount": 100, "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("success Exec doubledown", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "d", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("success Exec split", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "sp", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("success Exec insurance", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "i", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("success Exec declineinsurance", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "di", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("failed Exec other", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "other", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
	})
	t.Run("failed Exec command empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "", "sessionId": "test-session-1"}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
	})
	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": ""}`), &jsonCase1)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &jsonCase1)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
	})
	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.BlackJackWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
	})
}

func TestBlackJackWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":1000},"player":{"score":0,"cards":null,"chips":1000},"message":"reset."}`
	mockA := new(usecase.MockBlackJackInteractor)
	mockA.On("Reset").Return(mockOutput)
	mockB := new(usecase.MockBlackJackInteractor)
	mockB.On("Reset").Return(mockOutput)

	callCount := 0
	isoController := controller.NewBlackJackWebController(func() uc.BlackJackInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/blackjack/exec", isoController.Exec),
	)
	api.SetApp(router)

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "Reset")
		mockB.AssertNotCalled(t, "Reset")
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-B"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "Reset")
	})

	t.Run("session-A second call reuses mockA not mockB", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		// factory should have been called only twice total (once per session)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

func TestBlackJackWebController_NewCommands(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":0},"player":{"score":0,"cards":null,"chips":0},"message":"ok","deckCount":1}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("Surrender").Return(mockOutput)
	bjiMock.On("ToggleHint").Return(mockOutput)
	bjiMock.On("SetDeckCount", 6).Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	t.Run("sur", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"sur","sessionId":"bj-new-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
	})
	t.Run("surrender", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"surrender","sessionId":"bj-new-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
	})
	t.Run("togglehint", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"togglehint","sessionId":"bj-new-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
	})
	t.Run("sd with amount", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"sd","amount":6,"sessionId":"bj-new-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
	})
	t.Run("setdeckcount with amount", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"setdeckcount","amount":6,"sessionId":"bj-new-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
	})
}

func TestBlackJackWebController_ToggleSoft17(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":0},"player":{"score":0,"cards":null,"chips":0},"message":"","dealerHitsSoft17":true}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("ToggleSoft17").Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	var input controller.BlackJackWebInput
	_ = json.Unmarshal([]byte(`{"command":"togglesoft17","sessionId":"bj-soft17-1"}`), &input)
	req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorded := test.RunRequest(t, api.MakeHandler(), req)
	recorded.CodeIs(http.StatusOK)
	recorded.ContentTypeIsJson()
}

func TestBlackJackWebController_ToggleCounting(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":0},"player":{"score":0,"cards":null,"chips":0},"message":"","countingEnabled":true}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("ToggleCounting").Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	var input controller.BlackJackWebInput
	_ = json.Unmarshal([]byte(`{"command":"togglecounting","sessionId":"bj-counting-1"}`), &input)
	req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorded := test.RunRequest(t, api.MakeHandler(), req)
	recorded.CodeIs(http.StatusOK)
	recorded.ContentTypeIsJson()
}

func TestBlackJackWebController_ToggleDAS(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":0},"player":{"score":0,"cards":null,"chips":0},"message":"","doubleAfterSplit":false}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("ToggleDAS").Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	var input controller.BlackJackWebInput
	_ = json.Unmarshal([]byte(`{"command":"toggledas","sessionId":"bj-das-1"}`), &input)
	req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorded2 := test.RunRequest(t, api.MakeHandler(), req)
	recorded2.CodeIs(http.StatusOK)
	recorded2.ContentTypeIsJson()
}

func TestBlackJackWebController_ResetWithDASParam(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":1000},"player":{"score":0,"cards":null,"chips":1000},"message":"","doubleAfterSplit":false}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("ResetWithConfig", false, 0, false, false, 0, 0, 0).Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	var input controller.BlackJackWebInput
	_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"bj-das-config-1","doubleAfterSplit":false}`), &input)
	req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorded2 := test.RunRequest(t, api.MakeHandler(), req)
	recorded2.CodeIs(http.StatusOK)
	recorded2.ContentTypeIsJson()
	bjiMock.AssertCalled(t, "ResetWithConfig", false, 0, false, false, 0, 0, 0)
}

func TestBlackJackWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":1000},"player":{"score":0,"cards":null,"chips":1000},"message":"","dealerHitsSoft17":true,"cpuPlayerCount":2,"countingEnabled":true,"doubleAfterSplit":true,"countingSystem":1}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("ResetWithConfig", true, 2, true, true, 1, 0, 0).Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	var input controller.BlackJackWebInput
	_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"bj-config-1","dealerHitsSoft17":true,"cpuPlayerCount":2,"countingEnabled":true,"doubleAfterSplit":true,"countingSystem":1}`), &input)
	req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorded := test.RunRequest(t, api.MakeHandler(), req)
	recorded.CodeIs(http.StatusOK)
	recorded.ContentTypeIsJson()
	bjiMock.AssertCalled(t, "ResetWithConfig", true, 2, true, true, 1, 0, 0)
}

func TestBlackJackWebController_ResetWithoutConfig(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":1000},"player":{"score":0,"cards":null,"chips":1000},"message":""}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("Reset").Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	var input controller.BlackJackWebInput
	_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"bj-noconfig-1"}`), &input)
	req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorded := test.RunRequest(t, api.MakeHandler(), req)
	recorded.CodeIs(http.StatusOK)
	recorded.ContentTypeIsJson()
	bjiMock.AssertCalled(t, "Reset")
	bjiMock.AssertNotCalled(t, "ResetWithConfig", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestBlackJackWebController_SetCountingSystem(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":1000},"player":{"score":0,"cards":null,"chips":1000},"message":"","countingSystem":2}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("SetCountingSystem", 2).Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	t.Run("setcountingsystem with amount", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"setcountingsystem","amount":2,"sessionId":"bj-scs-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})

	t.Run("scs shorthand", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"scs","amount":2,"sessionId":"bj-scs-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
	})
}

func TestBlackJackWebController_BetWithSideBets(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":1000},"player":{"score":0,"cards":null,"chips":1000},"message":"","perfectPairsBet":10,"twentyOnePlus3Bet":20}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("Bet", 100, 10, 20, 0).Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	t.Run("bet with side bets", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"perfectPairsBet":10,"twentyOnePlus3Bet":20,"sessionId":"bj-side-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		bjiMock.AssertCalled(t, "Bet", 100, 10, 20, 0)
	})
}

func TestBlackJackWebController_BetWithoutSideBets(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":1000},"player":{"score":0,"cards":null,"chips":1000},"message":""}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("Bet", 100, 0, 0, 0).Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	t.Run("bet without side bets (nil defaults to 0)", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"sessionId":"bj-side-2"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		bjiMock.AssertCalled(t, "Bet", 100, 0, 0, 0)
	})
}

func TestBlackJackWebController_BetWithHandCount(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":1000},"player":{"score":0,"cards":null,"chips":1000},"message":"","multiHandCount":2}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("Bet", 100, 0, 0, 2).Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	t.Run("bet with handCount", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"handCount":2,"sessionId":"bj-hc-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		bjiMock.AssertCalled(t, "Bet", 100, 0, 0, 2)
	})
}

func TestBlackJackWebController_SetPenetration(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":1000},"player":{"score":0,"cards":null,"chips":1000},"message":"","deckPenetration":50}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("SetDeckPenetration", 50).Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	t.Run("setpenetration with amount", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"setpenetration","amount":50,"sessionId":"bj-pen-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})

	t.Run("pen shorthand", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"pen","amount":50,"sessionId":"bj-pen-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
	})
}

func TestBlackJackWebController_SetCpuPlayerCount(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":1000},"player":{"score":0,"cards":null,"chips":1000},"message":"","cpuPlayerCount":2}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("SetCpuPlayerCount", 2).Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	t.Run("setcpucount with amount", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"setcpucount","amount":2,"sessionId":"bj-scc-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})

	t.Run("scc shorthand", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"scc","amount":2,"sessionId":"bj-scc-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
	})
}

func TestBlackJackWebController_ResetWithPenetration(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":1000},"player":{"score":0,"cards":null,"chips":1000},"message":"","deckPenetration":50}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("ResetWithConfig", false, 0, false, true, 0, 50, 0).Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	var input controller.BlackJackWebInput
	_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"bj-pen-config-1","deckPenetration":50}`), &input)
	req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorded := test.RunRequest(t, api.MakeHandler(), req)
	recorded.CodeIs(http.StatusOK)
	recorded.ContentTypeIsJson()
	bjiMock.AssertCalled(t, "ResetWithConfig", false, 0, false, true, 0, 50, 0)
}

func TestBlackJackWebController_Stop(t *testing.T) {
	bjiMock := new(usecase.MockBlackJackInteractor)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	c := controller.NewBlackJackWebController(factory)
	// Stop should be idempotent and not panic when called multiple times.
	c.Stop()
	c.Stop()
}

func TestBlackJackWebController_EarlySurrender(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":0},"player":{"score":0,"cards":null,"chips":0},"message":"ok"}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("EarlySurrender").Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	t.Run("es", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"es","sessionId":"bj-es-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
	})
	t.Run("earlysurrender", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"earlysurrender","sessionId":"bj-es-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
	})
}

func TestBlackJackWebController_DeclineEarlySurrender(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":0},"player":{"score":0,"cards":null,"chips":0},"message":"ok"}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("DeclineEarlySurrender").Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	t.Run("des", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"des","sessionId":"bj-des-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
	})
	t.Run("declineearlysurrender", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"declineearlysurrender","sessionId":"bj-des-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
	})
}

func TestBlackJackWebController_SetSurrenderRule(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":0},"player":{"score":0,"cards":null,"chips":0},"message":"ok","surrenderRule":1}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("SetSurrenderRule", 1).Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	t.Run("ssr with amount", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"ssr","amount":1,"sessionId":"bj-ssr-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("setsurrenderrule with amount", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"setsurrenderrule","amount":1,"sessionId":"bj-ssr-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
}

func TestBlackJackWebController_ResetWithSurrenderRule(t *testing.T) {
	mockOutput := `{"dealer":{"score":0,"cards":null,"chips":1000},"player":{"score":0,"cards":null,"chips":1000},"message":"","surrenderRule":1}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("ResetWithConfig", false, 0, false, true, 0, 0, 1).Return(mockOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	var input controller.BlackJackWebInput
	_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"bj-sr-config-1","surrenderRule":1}`), &input)
	req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	recorded := test.RunRequest(t, api.MakeHandler(), req)
	recorded.CodeIs(http.StatusOK)
	recorded.ContentTypeIsJson()
	bjiMock.AssertCalled(t, "ResetWithConfig", false, 0, false, true, 0, 0, 1)
}

func TestBlackJackWebController_Log(t *testing.T) {
	mockLogOutput := `{"entries":[]}`
	bjiMock := new(usecase.MockBlackJackInteractor)
	bjiMock.On("ActionLog").Return(mockLogOutput)
	factory := func() uc.BlackJackInteractorIF { return bjiMock }
	tbc := controller.NewBlackJackWebController(factory)
	defer tbc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(rest.Post("/blackjack/exec", tbc.Exec))
	api.SetApp(router)

	t.Run("log command", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"bj-log-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockLogOutput)
	})

	t.Run("l shorthand", func(t *testing.T) {
		var input controller.BlackJackWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"bj-log-1"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/blackjack/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockLogOutput)
	})
}
