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

func mustOichoKabuOutputJSON(msg string) string {
	out := &controller.OichoKabuWebOutput{
		PlayerHand:    make([]*controller.WebOutputCard, 0),
		BankerHand:    make([]*controller.WebOutputCard, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustOichoKabuOutputJSON: %v", err))
	}
	return string(b)
}

func TestOichoKabuWebController_Method(t *testing.T) {
	mockOutput := mustOichoKabuOutputJSON("")

	okMock := new(usecase.MockOichoKabuInteractor)
	okMock.On("Reset").Return(mockOutput)
	okMock.On("Bet", 100).Return(mockOutput)
	okMock.On("Draw").Return(mockOutput)
	okMock.On("Stand").Return(mockOutput)
	okMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.OichoKabuInteractorIF { return okMock }
	ctrl := controller.NewOichoKabuWebController(factory)
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
		{"draw", `{"command":"draw","sessionId":"s7"}`},
		{"draw d", `{"command":"d","sessionId":"s8"}`},
		{"stand", `{"command":"stand","sessionId":"s9"}`},
		{"stand s", `{"command":"s","sessionId":"s10"}`},
		{"log", `{"command":"log","sessionId":"s11"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.OichoKabuWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			if strings.HasPrefix(tc.name, "quit") {
				recorded.BodyIs(mustOichoKabuOutputJSON("bye."))
			} else {
				recorded.BodyIs(mockOutput)
			}
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.OichoKabuWebInput
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
