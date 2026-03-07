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

func TestOldMaidWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"nextDrawTargetIdx":1,"gameEndFlag":false,"loserIdx":-1,"lastDrawPlayerIdx":-1,"lastDrawFromIdx":-1,"lastDiscardedPairs":0,"hasDrawn":false,"message":""}`
	omiMock := new(usecase.MockOldMaidInteractor)
	omiMock.On("Reset", mock.Anything).Return(mockOutput).Times(4)
	omiMock.On("Draw", -1).Return(mockOutput)
	omiMock.On("Draw", 2).Return(mockOutput)
	omiMock.On("Shuffle").Return(mockOutput)
	omiMock.On("Reorder", mock.Anything).Return(mockOutput)

	factory := func() uc.OldMaidInteractorIF { return omiMock }
	towc := controller.NewOldMaidWebController(factory)
	defer towc.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/oldmaid/exec", towc.Exec),
	)
	api.SetApp(router)

	var jsonInput controller.OldMaidWebInput
	// When "q" / "quit": newDefaultOutput is used (cpuHighlightedCardIdx=-1, removedCard=null, mode=0)
	qBody := `{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"lastDiscardedCards":null,"hasDrawn":false,"cpuActions":[],"humanAction":null,"drawHistory":[],"cpuHighlightedCardIdx":-1,"removedCard":null,"mode":0,"message":"bye."}`

	t.Run("success Exec q", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "q", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})
	t.Run("success Exec quit", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "quit", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})
	t.Run("success Exec r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "r", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec reset", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec d", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "d", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec draw", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "draw", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec draw with drawIdx", func(t *testing.T) {
		drawIdx := 2
		input := controller.OldMaidWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "test-session-1"},
			DrawIdx:      &drawIdx,
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec reset with mode and cpuPlacementStrategy", func(t *testing.T) {
		input := controller.OldMaidWebInput{
			BaseWebInput:         controller.BaseWebInput{Command: "reset", SessionID: "test-session-1"},
			Mode:                 1,
			CpuPlacementStrategy: true,
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("success Exec reset with cpuMemoryAI", func(t *testing.T) {
		input := controller.OldMaidWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "test-session-1"},
			CpuMemoryAI:  true,
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
	})
	t.Run("success Exec s (shuffle)", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "s", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec shuffle", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "shuffle", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec reorder with indices", func(t *testing.T) {
		input := controller.OldMaidWebInput{
			BaseWebInput:   controller.BaseWebInput{Command: "reorder", SessionID: "test-session-1"},
			ReorderIndices: []int{2, 0, 1},
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("failed Exec reorder without indices", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reorder", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"lastDiscardedCards":null,"hasDrawn":false,"cpuActions":[],"humanAction":null,"drawHistory":[],"cpuHighlightedCardIdx":-1,"removedCard":null,"mode":0,"message":"param error: reorderIndices is required."}`)
	})
	t.Run("failed Exec reset with invalid mode negative", func(t *testing.T) {
		input := controller.OldMaidWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "test-session-1"},
			Mode:         -1,
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"lastDiscardedCards":null,"hasDrawn":false,"cpuActions":[],"humanAction":null,"drawHistory":[],"cpuHighlightedCardIdx":-1,"removedCard":null,"mode":0,"message":"param error: mode must be between 0 and 1."}`)
	})
	t.Run("failed Exec reset with invalid mode too large", func(t *testing.T) {
		input := controller.OldMaidWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "test-session-1"},
			Mode:         99,
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"lastDiscardedCards":null,"hasDrawn":false,"cpuActions":[],"humanAction":null,"drawHistory":[],"cpuHighlightedCardIdx":-1,"removedCard":null,"mode":0,"message":"param error: mode must be between 0 and 1."}`)
	})
	t.Run("failed Exec other", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "other", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"lastDiscardedCards":null,"hasDrawn":false,"cpuActions":[],"humanAction":null,"drawHistory":[],"cpuHighlightedCardIdx":-1,"removedCard":null,"mode":0,"message":"Unsupported command."}`)
	})
	t.Run("failed Exec command empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "", "sessionId": "test-session-1"}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"lastDiscardedCards":null,"hasDrawn":false,"cpuActions":[],"humanAction":null,"drawHistory":[],"cpuHighlightedCardIdx":-1,"removedCard":null,"mode":0,"message":"param error."}`)
	})
	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": ""}`), &jsonInput)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &jsonInput)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"lastDiscardedCards":null,"hasDrawn":false,"cpuActions":[],"humanAction":null,"drawHistory":[],"cpuHighlightedCardIdx":-1,"removedCard":null,"mode":0,"message":"param error."}`)
	})
	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.OldMaidWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"nextDrawTargetIdx":0,"gameEndFlag":false,"loserIdx":0,"lastDrawPlayerIdx":0,"lastDrawFromIdx":0,"lastDrawCard":null,"lastDiscardedPairs":0,"lastDiscardedCards":null,"hasDrawn":false,"cpuActions":[],"humanAction":null,"drawHistory":[],"cpuHighlightedCardIdx":-1,"removedCard":null,"mode":0,"message":"param error."}`)
	})
}

func TestOldMaidWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"nextDrawTargetIdx":1,"gameEndFlag":false,"loserIdx":-1,"lastDrawPlayerIdx":-1,"lastDrawFromIdx":-1,"lastDiscardedPairs":0,"hasDrawn":false,"message":""}`
	mockA := new(usecase.MockOldMaidInteractor)
	mockA.On("Reset", mock.Anything).Return(mockOutput)
	mockB := new(usecase.MockOldMaidInteractor)
	mockB.On("Reset", mock.Anything).Return(mockOutput)

	callCount := 0
	isoController := controller.NewOldMaidWebController(func() uc.OldMaidInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/oldmaid/exec", isoController.Exec),
	)
	api.SetApp(router)

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.OldMaidWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "Reset", mock.Anything)
		mockB.AssertNotCalled(t, "Reset", mock.Anything)
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.OldMaidWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-B"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "Reset", mock.Anything)
	})

	t.Run("session-A second call reuses mockA", func(t *testing.T) {
		var input controller.OldMaidWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		req := test.MakeSimpleRequest("POST", "http://1.2.3.4/oldmaid/exec", &input)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		recorded := test.RunRequest(t, api.MakeHandler(), req)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

func TestOldMaidWebController_Stop(t *testing.T) {
	omiMock := new(usecase.MockOldMaidInteractor)
	factory := func() uc.OldMaidInteractorIF { return omiMock }
	c := controller.NewOldMaidWebController(factory)
	c.Stop()
	c.Stop()
}
