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

func mustRedDogOutputJSON(msg string) string {
	out := &controller.RedDogWebOutput{
		InitialCards:  make([]*controller.WebOutputCard, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustRedDogOutputJSON: %v", err))
	}
	return string(b)
}

func TestRedDogWebController_Method(t *testing.T) {
	mockOutput := `{"initialCards":[],"phase":0,"chips":0,"ante":0,"raise":0,"spread":0,"result":0,"totalPayout":0,"message":""}`
	expectedBody := mockOutput

	rdMock := new(usecase.MockRedDogInteractor)
	rdMock.On("Reset").Return(mockOutput)
	rdMock.On("Bet", 100).Return(mockOutput)
	rdMock.On("Raise", 50).Return(mockOutput)
	rdMock.On("Stay").Return(mockOutput)
	rdMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.RedDogInteractorIF { return rdMock }
	ctrl := controller.NewRedDogWebController(factory)
	defer ctrl.Stop()

	cases := []struct {
		name string
		body string
	}{
		{"quit q", `{"command":"q","sessionId":"s1"}`},
		{"quit long", `{"command":"quit","sessionId":"s2"}`},
		{"reset", `{"command":"reset","sessionId":"s3"}`},
		{"reset r", `{"command":"r","sessionId":"s4"}`},
		{"bet", `{"command":"bet","amount":100,"sessionId":"s5"}`},
		{"bet b", `{"command":"b","amount":100,"sessionId":"s6"}`},
		{"raise", `{"command":"raise","amount":50,"sessionId":"s7"}`},
		{"stay", `{"command":"stay","sessionId":"s8"}`},
		{"stay s", `{"command":"s","sessionId":"s9"}`},
		{"log", `{"command":"log","sessionId":"s10"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.RedDogWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			if strings.HasPrefix(tc.name, "quit") {
				recorded.BodyIs(mustRedDogOutputJSON("bye."))
			} else {
				recorded.BodyIs(expectedBody)
			}
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.RedDogWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s99"}`), &input)
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
