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

func mustBlackJackSwitchOutputJSON(msg string) string {
	out := &controller.BlackJackSwitchWebOutput{
		Hands:         make([]*controller.BlackJackSwitchWebOutputHand, 0),
		DealerCards:   make([]*controller.WebOutputCard, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBlackJackSwitchOutputJSON: %v", err))
	}
	return string(b)
}

func TestBlackJackSwitchWebController_Routes(t *testing.T) {
	mockOutput := `{"hands":[],"dealerCards":[],"dealerScore":0,"phase":0,"currentHandIdx":0,"chips":0,"switched":false,"dealerPushed22":false,"overallResult":0,"totalPayout":0,"message":""}`

	mk := func() *usecase.MockBlackJackSwitchInteractor {
		m := new(usecase.MockBlackJackSwitchInteractor)
		m.On("Reset").Return(mockOutput)
		m.On("Bet", 100).Return(mockOutput)
		m.On("Switch").Return(mockOutput)
		m.On("Keep").Return(mockOutput)
		m.On("Hit").Return(mockOutput)
		m.On("Stand").Return(mockOutput)
		m.On("DoubleDown").Return(mockOutput)
		m.On("ActionLog").Return(mockOutput)
		return m
	}
	mockInteractor := mk()

	factory := func() uc.BlackJackSwitchInteractorIF { return mockInteractor }
	ctrl := controller.NewBlackJackSwitchWebController(factory)
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
		{"switch", `{"command":"switch","sessionId":"s7"}`},
		{"switch sw", `{"command":"sw","sessionId":"s8"}`},
		{"keep", `{"command":"keep","sessionId":"s9"}`},
		{"keep k", `{"command":"k","sessionId":"s10"}`},
		{"hit", `{"command":"hit","sessionId":"s11"}`},
		{"stand", `{"command":"stand","sessionId":"s12"}`},
		{"doubledown", `{"command":"doubledown","sessionId":"s13"}`},
		{"doubledown dd", `{"command":"dd","sessionId":"s14"}`},
		{"log", `{"command":"log","sessionId":"s15"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.BlackJackSwitchWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			if strings.HasPrefix(tc.name, "quit") {
				recorded.BodyIs(mustBlackJackSwitchOutputJSON("bye."))
			} else {
				recorded.BodyIs(mockOutput)
			}
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.BlackJackSwitchWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s99"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("invalid json", func(t *testing.T) {
		recorded := execRequest(t, ctrl.Exec, strings.NewReader("{invalid"))
		recorded.CodeIs(http.StatusBadRequest)
	})
}
