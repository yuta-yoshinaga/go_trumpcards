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

func mustCasinoWarOutputJSON(msg string) string {
	out := &controller.CasinoWarWebOutput{
		BurnCards:     make([]*controller.WebOutputCard, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustCasinoWarOutputJSON: %v", err))
	}
	return string(b)
}

func TestCasinoWarWebController_Method(t *testing.T) {
	mockOutput := `{"burnCards":[],"phase":0,"chips":0,"ante":0,"warBet":0,"result":0,"totalPayout":0,"message":""}`
	expectedBody := mockOutput

	cwMock := new(usecase.MockCasinoWarInteractor)
	cwMock.On("Reset").Return(mockOutput)
	cwMock.On("Bet", 100).Return(mockOutput)
	cwMock.On("Surrender").Return(mockOutput)
	cwMock.On("War").Return(mockOutput)
	cwMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.CasinoWarInteractorIF { return cwMock }
	ctrl := controller.NewCasinoWarWebController(factory)
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
		{"surrender", `{"command":"surrender","sessionId":"s7"}`},
		{"war", `{"command":"war","sessionId":"s8"}`},
		{"log", `{"command":"log","sessionId":"s9"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.CasinoWarWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			if strings.HasPrefix(tc.name, "quit") {
				recorded.BodyIs(mustCasinoWarOutputJSON("bye."))
			} else {
				recorded.BodyIs(expectedBody)
			}
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.CasinoWarWebInput
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
