//go:build test
// +build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func TestBassetWebController_Constructs(t *testing.T) {
	assert.NotNil(t, controller.NewBassetWebController)
}

func TestBassetWebController_Commands(t *testing.T) {
	mockOutput := `{"phase":1,"chips":1000,"bet":null,"hit":false,"turnsPlayed":0,"turnsTotal":24,"remaining":52,"remainingByRank":[],"totalPayout":0,"gameEndFlag":false,"message":"ok"}`
	m := new(usecase.MockBassetInteractor)
	m.On("Reset").Return(mockOutput)
	m.On("NextRound").Return(mockOutput)
	m.On("PlaceBet", 7, 100).Return(mockOutput)
	m.On("DealTurn").Return(mockOutput)
	m.On("TakeWinnings").Return(mockOutput)
	m.On("PressParoli").Return(mockOutput)
	m.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewBassetWebController(func() uc.BassetInteractorIF { return m })
	defer ctrl.Stop()
	cases := []struct {
		name string
		body string
	}{
		{"reset", `{"command":"reset","sessionId":"b1"}`},
		{"bet", `{"command":"b","rank":7,"amount":100,"sessionId":"b2"}`},
		{"bet alias", `{"command":"bet","rank":7,"amount":100,"sessionId":"b3"}`},
		{"deal", `{"command":"d","sessionId":"b4"}`},
		{"deal alias", `{"command":"deal","sessionId":"b5"}`},
		{"take", `{"command":"take","sessionId":"b6"}`},
		{"take alias", `{"command":"takeWinnings","sessionId":"b7"}`},
		{"paroli", `{"command":"paroli","sessionId":"b8"}`},
		{"paroli alias", `{"command":"pressParoli","sessionId":"b9"}`},
		{"next", `{"command":"n","sessionId":"b10"}`},
		{"next alias", `{"command":"next","sessionId":"b11"}`},
		{"log", `{"command":"log","sessionId":"b12"}`},
		{"quit", `{"command":"q","sessionId":"b13"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.BassetWebInput
			assert.NoError(t, json.Unmarshal([]byte(tc.body), &input))
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.ContentTypeIsJson()
			if tc.name == "quit" {
				if !strings.Contains(recorded.Body.String(), `"message":"bye."`) {
					t.Errorf("quit response = %s", recorded.Body.String())
				}
			} else {
				recorded.BodyIs(mockOutput)
			}
		})
	}
}

func TestBassetWebController_InvalidRequests(t *testing.T) {
	m := new(usecase.MockBassetInteractor)
	ctrl := controller.NewBassetWebController(func() uc.BassetInteractorIF { return m })
	defer ctrl.Stop()
	for _, body := range []string{
		`{"command":"unknown","sessionId":"bad1"}`,
		`{"command":"reset"}`,
		`{"sessionId":"bad3"}`,
		"{invalid",
	} {
		recorded := execRequest(t, ctrl.Exec, strings.NewReader(body))
		recorded.CodeIs(http.StatusBadRequest)
	}
}
