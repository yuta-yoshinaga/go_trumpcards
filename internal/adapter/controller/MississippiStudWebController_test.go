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

func mustMississippiStudOutputJSON(msg string) string {
	out := &controller.MississippiStudWebOutput{
		PlayerHand:        make([]*controller.WebOutputCard, 0),
		CommunityCards:    make([]*controller.WebOutputCard, 0),
		CommunityRevealed: make([]bool, 0),
		StreetMultipliers: make([]int, 0),
		StreetPayouts:     make([]int, 0),
		WebOutputBase:     controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMississippiStudOutputJSON: %v", err))
	}
	return string(b)
}

func TestMississippiStudWebController_Method(t *testing.T) {
	mockOutput := mustMississippiStudOutputJSON("")
	expectedBody := mockOutput

	mi := new(usecase.MockMississippiStudInteractor)
	mi.On("Reset").Return(mockOutput)
	mi.On("Bet", 100).Return(mockOutput)
	mi.On("Play", 1).Return(mockOutput)
	mi.On("Play", 2).Return(mockOutput)
	mi.On("Play", 3).Return(mockOutput)
	mi.On("Fold").Return(mockOutput)
	mi.On("ActionLog").Return(mockOutput)

	factory := func() uc.MississippiStudInteractorIF { return mi }
	ctrl := controller.NewMississippiStudWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.MississippiStudWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustMississippiStudOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.MississippiStudWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s2"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("reset short", func(t *testing.T) {
		var input controller.MississippiStudWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s3"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet", func(t *testing.T) {
		var input controller.MississippiStudWebInput
		_ = json.Unmarshal([]byte(`{"command":"bet","amount":100,"sessionId":"s4"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet short", func(t *testing.T) {
		var input controller.MississippiStudWebInput
		_ = json.Unmarshal([]byte(`{"command":"b","amount":100,"sessionId":"s5"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("play", func(t *testing.T) {
		var input controller.MississippiStudWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","multiplier":2,"sessionId":"s6"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("play short", func(t *testing.T) {
		var input controller.MississippiStudWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","multiplier":3,"sessionId":"s7"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("fold", func(t *testing.T) {
		var input controller.MississippiStudWebInput
		_ = json.Unmarshal([]byte(`{"command":"fold","sessionId":"s8"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("fold short", func(t *testing.T) {
		var input controller.MississippiStudWebInput
		_ = json.Unmarshal([]byte(`{"command":"f","sessionId":"s9"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.MississippiStudWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s10"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.MississippiStudWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s11"}`), &input)
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
