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

func mustFaroOutputJSON(msg string) string {
	out := &controller.FaroWebOutput{
		Bets:      make([]*controller.FaroWebBet, 0),
		CallCards: make([]*controller.WebOutputCard, 0),
		CallOrder: make([]int, 0),
		// 既定値は `newFaroDefaultOutput` と揃える ── `null` を返すと、
		// クライアントが添字で引くケースキーパーが落ちる (#6471)。
		RemainingByRank: make([]int, 0),
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustFaroOutputJSON: %v", err))
	}
	return string(b)
}

func TestFaroWebController_Method(t *testing.T) {
	mockOutput := `{"phase":1,"chips":1000,"bets":[],"split":false,"turnsPlayed":0,"turnsTotal":24,"remaining":51,"callCards":[],"callOrder":[],"callWon":false,"totalPayout":0,"gameEndFlag":false,"message":""}`

	faMock := new(usecase.MockFaroInteractor)
	faMock.On("Reset").Return(mockOutput)
	faMock.On("NextRound").Return(mockOutput)
	faMock.On("PlaceBet", 7, 100, false).Return(mockOutput)
	faMock.On("PlaceBet", 7, 100, true).Return(mockOutput)
	faMock.On("ClearBet", 3).Return(mockOutput)
	faMock.On("ClearAll").Return(mockOutput)
	faMock.On("DealTurn").Return(mockOutput)
	faMock.On("Call", []int{3, 9, 12}).Return(mockOutput)
	faMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.FaroInteractorIF { return faMock }
	ctrl := controller.NewFaroWebController(factory)
	defer ctrl.Stop()

	cases := []struct {
		name string
		body string
	}{
		{"quit q", `{"command":"q","sessionId":"s1"}`},
		{"reset", `{"command":"reset","sessionId":"s2"}`},
		{"bet", `{"command":"b","rank":7,"amount":100,"sessionId":"s3"}`},
		{"bet copper", `{"command":"bet","rank":7,"amount":100,"copper":true,"sessionId":"s4"}`},
		{"clearBet", `{"command":"cb","rank":3,"sessionId":"s5"}`},
		{"clearAll", `{"command":"ca","sessionId":"s6"}`},
		{"deal", `{"command":"d","sessionId":"s7"}`},
		{"call", `{"command":"call","order":[3,9,12],"sessionId":"s8"}`},
		{"next", `{"command":"n","sessionId":"s9"}`},
		{"log", `{"command":"log","sessionId":"s10"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.FaroWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			if strings.HasPrefix(tc.name, "quit") {
				recorded.BodyIs(mustFaroOutputJSON("bye."))
			} else {
				recorded.BodyIs(mockOutput)
			}
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.FaroWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s99"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		body := strings.TrimSpace(recorded.Body.String())
		if !strings.Contains(body, "Unsupported command") {
			t.Errorf("expected Unsupported command, got: %s", body)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		recorded := execRequest(t, ctrl.Exec, strings.NewReader("{invalid"))
		recorded.CodeIs(http.StatusBadRequest)
	})
}
