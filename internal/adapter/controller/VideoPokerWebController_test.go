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

func mustVideoPokerOutputJSON(msg string) string {
	out := &controller.VideoPokerWebOutput{
		Hand:          make([]*controller.WebOutputCard, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustVideoPokerOutputJSON: %v", err))
	}
	return string(b)
}

func TestVideoPokerWebController_Method(t *testing.T) {
	mockOutput := mustVideoPokerOutputJSON("")
	expectedBody := mockOutput

	viMock := new(usecase.MockVideoPokerInteractor)
	viMock.On("Reset").Return(mockOutput)
	viMock.On("Bet", 3).Return(mockOutput)
	viMock.On("Hold", []int{0, 2, 4}).Return(mockOutput)
	viMock.On("Hold", []int{}).Return(mockOutput)
	viMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.VideoPokerInteractorIF { return viMock }
	ctrl := controller.NewVideoPokerWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.VideoPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustVideoPokerOutputJSON("bye."))
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.VideoPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.VideoPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet", func(t *testing.T) {
		input := controller.VideoPokerWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "s1"},
			Amount:       3,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("bet long", func(t *testing.T) {
		input := controller.VideoPokerWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bet", SessionID: "s1"},
			Amount:       3,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hold", func(t *testing.T) {
		input := controller.VideoPokerWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "h", SessionID: "s1"},
			Indices:      []int{0, 2, 4},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hold long", func(t *testing.T) {
		input := controller.VideoPokerWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "hold", SessionID: "s1"},
			Indices:      []int{0, 2, 4},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hold nil indices", func(t *testing.T) {
		var input controller.VideoPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("action log", func(t *testing.T) {
		var input controller.VideoPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("action log l", func(t *testing.T) {
		var input controller.VideoPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.VideoPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustVideoPokerOutputJSON("Unsupported command."))
	})

	t.Run("param error empty", func(t *testing.T) {
		recorded := execRequest(t, ctrl.Exec, strings.NewReader(""))
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustVideoPokerOutputJSON("param error."))
	})

	t.Run("param error no command", func(t *testing.T) {
		var input controller.VideoPokerWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustVideoPokerOutputJSON("param error."))
	})

	t.Run("stop twice", func(t *testing.T) {
		ctrl2 := controller.NewVideoPokerWebController(factory)
		ctrl2.Stop()
		ctrl2.Stop()
	})
}
