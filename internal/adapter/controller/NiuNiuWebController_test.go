//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustNiuNiuOutputJSON(msg string) string {
	out := &controller.NiuNiuWebOutput{
		Seats:         []*controller.NiuNiuWebOutputSeat{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustNiuNiuOutputJSON: %v", err))
	}
	return string(b)
}

func TestNiuNiuWebController_Method(t *testing.T) {
	mockOutput := `{"seats":[],"bankerIdx":3,"chips":1000,"lastResult":"","phase":1,"message":""}`

	niMock := new(usecase.MockNiuNiuInteractor)
	niMock.On("Reset").Return(mockOutput)
	niMock.On("Bet", 100).Return(mockOutput)
	niMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewNiuNiuWebController(func() uc.NiuNiuInteractorIF { return niMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.NiuNiuWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustNiuNiuOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"bet", `{"command":"b","sessionId":"s1","amount":100}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := exec(t, tc.body)
			r.CodeIs(http.StatusOK)
			r.BodyIs(mockOutput)
		})
	}

	t.Run("bet missing amount", func(t *testing.T) {
		exec(t, `{"command":"b","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}
