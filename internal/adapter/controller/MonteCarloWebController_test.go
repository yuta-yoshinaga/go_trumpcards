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

func mustMonteCarloOutputJSON(msg string) string {
	out := &controller.MonteCarloWebOutput{
		Board:         [][]*controller.MonteCarloWebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMonteCarloOutputJSON: %v", err))
	}
	return string(b)
}

func monteCarloIntPtr(v int) *int { return &v }

func setupMonteCarloWebTest(t *testing.T) (*usecase.MockMonteCarloInteractor, *controller.MonteCarloWebController, string) {
	t.Helper()
	mockOutput := `{"board":[],"phase":0,"stockCount":0,"removedCount":0,"dealCount":0,"canUndo":false,"isStalemate":false,"message":""}`
	miMock := new(usecase.MockMonteCarloInteractor)
	factory := func() uc.MonteCarloInteractorIF { return miMock }
	ctrl := controller.NewMonteCarloWebController(factory)
	t.Cleanup(func() { ctrl.Stop() })
	return miMock, ctrl, mockOutput
}

func monteCarloPostInput(t *testing.T, handler http.HandlerFunc, input controller.MonteCarloWebInput) *recorded {
	t.Helper()
	return execRequest(t, handler, &input)
}

func monteCarloPost(t *testing.T, handler http.HandlerFunc, body string) *recorded {
	t.Helper()
	var input controller.MonteCarloWebInput
	_ = json.Unmarshal([]byte(body), &input)
	return execRequest(t, handler, &input)
}

func TestMonteCarloWebController_Commands(t *testing.T) {
	miMock, ctrl, mockOutput := setupMonteCarloWebTest(t)
	miMock.On("Reset").Return(mockOutput)
	miMock.On("Deal").Return(mockOutput)
	miMock.On("Undo").Return(mockOutput)
	miMock.On("GiveUp").Return(mockOutput)
	miMock.On("Hint").Return(mockOutput)
	miMock.On("ActionLog").Return(mockOutput)

	tests := []struct {
		name    string
		command string
	}{
		{"reset", `{"command":"reset","sessionId":"s1"}`},
		{"deal", `{"command":"deal","sessionId":"s1"}`},
		{"undo", `{"command":"undo","sessionId":"s1"}`},
		{"giveup", `{"command":"giveup","sessionId":"s1"}`},
		{"hint", `{"command":"hint","sessionId":"s1"}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := monteCarloPost(t, ctrl.Exec, tt.command)
			rec.CodeIs(http.StatusOK)
		})
	}
}

func TestMonteCarloWebController_Remove(t *testing.T) {
	miMock, ctrl, mockOutput := setupMonteCarloWebTest(t)
	miMock.On("Remove", 0, 0, 0, 1).Return(mockOutput)

	rec := monteCarloPostInput(t, ctrl.Exec, controller.MonteCarloWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "remove", SessionID: "s1"},
		FromR:        monteCarloIntPtr(0),
		FromC:        monteCarloIntPtr(0),
		ToR:          monteCarloIntPtr(0),
		ToC:          monteCarloIntPtr(1),
	})
	rec.CodeIs(http.StatusOK)
}

func TestMonteCarloWebController_RemoveMissingCoords(t *testing.T) {
	_, ctrl, _ := setupMonteCarloWebTest(t)
	rec := monteCarloPostInput(t, ctrl.Exec, controller.MonteCarloWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "remove", SessionID: "s1"},
		FromR:        monteCarloIntPtr(0),
	})
	rec.CodeIs(http.StatusBadRequest)
}

func TestMonteCarloWebController_Quit(t *testing.T) {
	_, ctrl, _ := setupMonteCarloWebTest(t)
	rec := monteCarloPost(t, ctrl.Exec, `{"command":"q","sessionId":"s1"}`)
	rec.CodeIs(http.StatusOK)
	rec.BodyIs(mustMonteCarloOutputJSON("bye."))
}

func TestMonteCarloWebController_UnknownCommand(t *testing.T) {
	_, ctrl, _ := setupMonteCarloWebTest(t)
	rec := monteCarloPost(t, ctrl.Exec, `{"command":"xyz","sessionId":"s1"}`)
	rec.CodeIs(http.StatusBadRequest)
}
