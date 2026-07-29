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

func mustPontoonOutputJSON(msg string) string {
	out := &controller.PontoonWebOutput{
		Seats:         []*controller.PontoonWebOutputSeat{},
		NextBanker:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustPontoonOutputJSON: %v", err))
	}
	return string(b)
}

func TestPontoonWebController_Method(t *testing.T) {
	mockOutput := `{"seats":[],"bankerIdx":0,"isHumanBanker":false,"chips":1000,"activeSeat":0,"activeHand":0,"nextBanker":-1,"lastResult":"","phase":1,"canStick":false,"canTwist":false,"canBuy":false,"canSplit":false,"message":""}`

	piMock := new(usecase.MockPontoonInteractor)
	piMock.On("Reset").Return(mockOutput)
	piMock.On("Bet", 100).Return(mockOutput)
	piMock.On("Deal").Return(mockOutput)
	piMock.On("Stick").Return(mockOutput)
	piMock.On("Twist").Return(mockOutput)
	piMock.On("Buy", 50).Return(mockOutput)
	piMock.On("Split").Return(mockOutput)
	piMock.On("BankerTwist").Return(mockOutput)
	piMock.On("BankerStay").Return(mockOutput)
	piMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewPontoonWebController(func() uc.PontoonInteractorIF { return piMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.PontoonWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustPontoonOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"bet", `{"command":"b","sessionId":"s1","amount":100}`},
		{"deal as banker", `{"command":"deal","sessionId":"s1"}`},
		{"stick", `{"command":"s","sessionId":"s1"}`},
		{"twist", `{"command":"t","sessionId":"s1"}`},
		{"buy", `{"command":"buy","sessionId":"s1","amount":50}`},
		{"split", `{"command":"sp","sessionId":"s1"}`},
		{"banker twist", `{"command":"bt","sessionId":"s1"}`},
		{"banker stay", `{"command":"bs","sessionId":"s1"}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := exec(t, tc.body)
			r.CodeIs(http.StatusOK)
			r.BodyIs(mockOutput)
		})
	}

	for _, tc := range []struct{ name, body string }{
		{"bet missing amount", `{"command":"b","sessionId":"s1"}`},
		{"buy missing amount", `{"command":"buy","sessionId":"s1"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec(t, tc.body).CodeIs(http.StatusBadRequest)
		})
	}
}
