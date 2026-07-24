//go:build test

package controller_test

import (
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func setupClockSolitaireWebTest(t *testing.T) (*usecase.MockClockSolitaireInteractor, *controller.ClockSolitaireWebController, string) {
	t.Helper()
	mockOutput := `{"piles":[],"faceUpCount":[],"phase":0,"stepCount":0,"message":""}`
	ciMock := new(usecase.MockClockSolitaireInteractor)
	factory := func() uc.ClockSolitaireInteractorIF { return ciMock }
	ctrl := controller.NewClockSolitaireWebController(factory)
	t.Cleanup(func() { ctrl.Stop() })

	return ciMock, ctrl, mockOutput
}

func clockSolitairePost(t *testing.T, handler http.HandlerFunc, input *controller.ClockSolitaireWebInput) *recorded {
	t.Helper()
	return execRequest(t, handler, input)
}

func TestClockSolitaireWebController_Commands(t *testing.T) {
	ciMock, ctrl, mockOutput := setupClockSolitaireWebTest(t)

	ciMock.On("Reset").Return(mockOutput)
	ciMock.On("Step").Return(mockOutput)
	ciMock.On("AutoPlay").Return(mockOutput)
	ciMock.On("Undo").Return(mockOutput)
	ciMock.On("ActionLog").Return(mockOutput)

	tests := []struct {
		name    string
		command string
	}{
		{"reset", "reset"},
		{"step", "step"},
		{"autoplay", "autoplay"},
		{"undo", "undo"},
		{"log", "log"},
		{"short-r", "r"},
		{"short-s", "s"},
		{"short-a", "a"},
		{"short-u", "u"},
		{"short-l", "l"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &controller.ClockSolitaireWebInput{
				BaseWebInput: controller.BaseWebInput{Command: tt.command, SessionID: "s1"},
			}
			rec := clockSolitairePost(t, ctrl.Exec, input)
			rec.CodeIs(http.StatusOK)
		})
	}
}

func TestClockSolitaireWebController_UnknownCommand(t *testing.T) {
	_, ctrl, _ := setupClockSolitaireWebTest(t)

	input := &controller.ClockSolitaireWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "unknown", SessionID: "s1"},
	}
	rec := clockSolitairePost(t, ctrl.Exec, input)
	rec.CodeIs(http.StatusBadRequest)
}
