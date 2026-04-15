//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustLetItRideOutputJSON(msg string) string {
	out := &controller.LetItRideWebOutput{
		PlayerHand:     make([]*controller.WebOutputCard, 0),
		CommunityCards: make([]*controller.WebOutputCard, 0),
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustLetItRideOutputJSON: %v", err))
	}
	return string(b)
}

func TestLetItRideWebController_Method(t *testing.T) {
	mockOutput := `{"playerHand":[],"communityCards":[],"phase":0,"chips":0,"betAmount":0,"bet1Active":false,"bet2Active":false,"bet3Active":false,"result":0,"handRank":0,"bet1Payout":0,"bet2Payout":0,"bet3Payout":0,"totalPayout":0,"message":""}`
	expectedBody := mockOutput

	liMock := new(usecase.MockLetItRideInteractor)
	liMock.On("Reset").Return(mockOutput)
	liMock.On("Bet", 100).Return(mockOutput)
	liMock.On("Pull").Return(mockOutput)
	liMock.On("LetItRide").Return(mockOutput)
	liMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.LetItRideInteractorIF { return liMock }
	ctrl := controller.NewLetItRideWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.LetItRideWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustLetItRideOutputJSON("bye."))
	})

	t.Run("quit", func(t *testing.T) {
		var input controller.LetItRideWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s2"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.LetItRideWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s3"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.LetItRideWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s4"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet", func(t *testing.T) {
		var input controller.LetItRideWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"sessionId":"s5"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet b", func(t *testing.T) {
		var input controller.LetItRideWebInput
		_ = json.Unmarshal([]byte(`{"command":"b","amount":100,"sessionId":"s6"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("pull", func(t *testing.T) {
		var input controller.LetItRideWebInput
		_ = json.Unmarshal([]byte(`{"command":"pull","sessionId":"s7"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("pull p", func(t *testing.T) {
		var input controller.LetItRideWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s8"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("letitride", func(t *testing.T) {
		var input controller.LetItRideWebInput
		_ = json.Unmarshal([]byte(`{"command":"letitride","sessionId":"s9"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("letitride l", func(t *testing.T) {
		var input controller.LetItRideWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"s10"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.LetItRideWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s11"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.LetItRideWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s12"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		body := strings.TrimSpace(recorded.Body.String())
		if !strings.Contains(body, "Unsupported command") {
			t.Errorf("expected Unsupported command in body, got: %s", body)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		recorded := execRequest(t, ctrl.Exec, strings.NewReader("{invalid"))
		recorded.CodeIs(http.StatusBadRequest)
	})
}
