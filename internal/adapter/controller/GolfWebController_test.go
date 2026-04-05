//go:build test

package controller_test

import (
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func golfIntPtr(v int) *int { return &v }

func setupGolfWebTest(t *testing.T) (*usecase.MockGolfInteractor, *controller.GolfWebController, string) {
	t.Helper()
	mockOutput := `{"layout":[],"stockCount":0,"waste":[],"phase":0,"moveCount":0,"message":""}`
	giMock := new(usecase.MockGolfInteractor)
	factory := func() uc.GolfInteractorIF { return giMock }
	ctrl := controller.NewGolfWebController(factory)
	t.Cleanup(func() { ctrl.Stop() })

	return giMock, ctrl, mockOutput
}

func golfPost(t *testing.T, handler http.HandlerFunc, input *controller.GolfWebInput) *recorded {
	t.Helper()
	return execRequest(t, handler, input)
}

func TestGolfWebController_Commands(t *testing.T) {
	giMock, ctrl, mockOutput := setupGolfWebTest(t)

	giMock.On("Reset").Return(mockOutput)
	giMock.On("Draw").Return(mockOutput)
	giMock.On("GiveUp").Return(mockOutput)
	giMock.On("Hint").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)
	giMock.On("Undo").Return(mockOutput)

	tests := []struct {
		name    string
		command string
	}{
		{"reset", "reset"},
		{"draw", "draw"},
		{"giveup", "giveup"},
		{"hint", "hint"},
		{"log", "log"},
		{"undo", "undo"},
		{"short-r", "r"},
		{"short-d", "d"},
		{"short-g", "g"},
		{"short-u", "u"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &controller.GolfWebInput{
				BaseWebInput: controller.BaseWebInput{Command: tt.command, SessionID: "s1"},
			}
			rec := golfPost(t, ctrl.Exec, input)
			rec.CodeIs(http.StatusOK)
		})
	}
}

func TestGolfWebController_Remove_ShortAlias(t *testing.T) {
	giMock, ctrl, mockOutput := setupGolfWebTest(t)
	giMock.On("Reset").Return(mockOutput)
	giMock.On("Remove", 2).Return(mockOutput)

	golfPost(t, ctrl.Exec, &controller.GolfWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s5"},
	})

	input := &controller.GolfWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "rm", SessionID: "s5"},
		Col:          golfIntPtr(2),
	}
	rec := golfPost(t, ctrl.Exec, input)
	rec.CodeIs(http.StatusOK)
}

func TestGolfWebController_Remove(t *testing.T) {
	giMock, ctrl, mockOutput := setupGolfWebTest(t)
	giMock.On("Reset").Return(mockOutput)
	giMock.On("Remove", 3).Return(mockOutput)

	// First reset to create session
	golfPost(t, ctrl.Exec, &controller.GolfWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s2"},
	})

	// Remove with col
	input := &controller.GolfWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "remove", SessionID: "s2"},
		Col:          golfIntPtr(3),
	}
	rec := golfPost(t, ctrl.Exec, input)
	rec.CodeIs(http.StatusOK)
}

func TestGolfWebController_Remove_MissingParams(t *testing.T) {
	giMock, ctrl, mockOutput := setupGolfWebTest(t)
	giMock.On("Reset").Return(mockOutput)

	// First reset to create session
	golfPost(t, ctrl.Exec, &controller.GolfWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s3"},
	})

	// Remove without col
	input := &controller.GolfWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "remove", SessionID: "s3"},
	}
	rec := golfPost(t, ctrl.Exec, input)
	rec.CodeIs(http.StatusBadRequest)
}

func TestGolfWebController_UndoN(t *testing.T) {
	giMock, ctrl, mockOutput := setupGolfWebTest(t)
	giMock.On("UndoN", 3).Return(mockOutput)

	t.Run("undo_n with valid n", func(t *testing.T) {
		n := 3
		rec := golfPost(t, ctrl.Exec, &controller.GolfWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1", N: &n},
		})
		rec.CodeIs(http.StatusOK)
	})

	t.Run("undo_n with missing n", func(t *testing.T) {
		rec := golfPost(t, ctrl.Exec, &controller.GolfWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1"},
		})
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestGolfWebController_UnknownCommand(t *testing.T) {
	_, ctrl, _ := setupGolfWebTest(t)

	input := &controller.GolfWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "unknown", SessionID: "s4"},
	}
	rec := golfPost(t, ctrl.Exec, input)
	rec.CodeIs(http.StatusBadRequest)
}
