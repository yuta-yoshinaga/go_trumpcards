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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustDragonTigerOutputJSON(msg string) string {
	out := &controller.DragonTigerWebOutput{
		History:       make([]int, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustDragonTigerOutputJSON: %v", err))
	}
	return string(b)
}

func TestDragonTigerWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0,"chips":0,"betAmount":0,"betType":0,"result":0,"payout":0,"history":[],"message":""}`
	expectedBody := mockOutput

	dtMock := new(usecase.MockDragonTigerInteractor)
	dtMock.On("Reset").Return(mockOutput)
	dtMock.On("Bet", 100, domain.DragonTigerBetDragon).Return(mockOutput)
	dtMock.On("ClearHistory").Return(mockOutput)
	dtMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.DragonTigerInteractorIF { return dtMock }
	ctrl := controller.NewDragonTigerWebController(factory)
	defer ctrl.Stop()

	cases := []struct {
		name string
		body string
	}{
		{"quit q", `{"command":"q","sessionId":"s1"}`},
		{"quit long", `{"command":"quit","sessionId":"s2"}`},
		{"reset", `{"command":"reset","sessionId":"s3"}`},
		{"reset r", `{"command":"r","sessionId":"s4"}`},
		{"bet", `{"command":"bet","amount":100,"betType":0,"sessionId":"s5"}`},
		{"bet b", `{"command":"b","amount":100,"betType":0,"sessionId":"s6"}`},
		{"clear", `{"command":"clear","sessionId":"s7"}`},
		{"log", `{"command":"log","sessionId":"s8"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.DragonTigerWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			if strings.HasPrefix(tc.name, "quit") {
				recorded.BodyIs(mustDragonTigerOutputJSON("bye."))
			} else {
				recorded.BodyIs(expectedBody)
			}
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.DragonTigerWebInput
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
