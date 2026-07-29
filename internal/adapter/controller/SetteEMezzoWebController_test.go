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

func mustSetteEMezzoOutputJSON(msg string) string {
	out := &controller.SetteEMezzoWebOutput{
		Seats:         []*controller.SetteEMezzoWebOutputSeat{},
		NextBanker:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSetteEMezzoOutputJSON: %v", err))
	}
	return string(b)
}

func TestSetteEMezzoWebController_Method(t *testing.T) {
	mockOutput := `{"seats":[],"bankerIdx":0,"isHumanBanker":false,"chips":1000,"activeSeat":0,"nextBanker":-1,"lastResult":"","phase":1,"targetHalves":15,"canHit":false,"canStand":false,"canSetMatta":false,"message":""}`

	siMock := new(usecase.MockSetteEMezzoInteractor)
	siMock.On("Reset").Return(mockOutput)
	siMock.On("Bet", 100).Return(mockOutput)
	siMock.On("Deal").Return(mockOutput)
	siMock.On("Hit").Return(mockOutput)
	siMock.On("Stand").Return(mockOutput)
	siMock.On("Matta", 6).Return(mockOutput)
	siMock.On("BankerHit").Return(mockOutput)
	siMock.On("BankerStand").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewSetteEMezzoWebController(func() uc.SetteEMezzoInteractorIF { return siMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.SetteEMezzoWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustSetteEMezzoOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"bet", `{"command":"b","sessionId":"s1","amount":100}`},
		{"deal as banker", `{"command":"deal","sessionId":"s1"}`},
		{"hit", `{"command":"h","sessionId":"s1"}`},
		{"stand", `{"command":"s","sessionId":"s1"}`},
		// The matta's value travels in HALVES, so 6 means three points.
		{"matta", `{"command":"matta","sessionId":"s1","amount":6}`},
		{"banker hit", `{"command":"bh","sessionId":"s1"}`},
		{"banker stand", `{"command":"bs","sessionId":"s1"}`},
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
		{"matta missing amount", `{"command":"matta","sessionId":"s1"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec(t, tc.body).CodeIs(http.StatusBadRequest)
		})
	}
}
