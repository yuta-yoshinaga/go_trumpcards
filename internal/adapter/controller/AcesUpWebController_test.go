//go:build test

package controller_test

import (
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func acesupIntPtr(v int) *int { return &v }

func setupAcesUpWebTest(t *testing.T) (*usecase.MockAcesUpInteractor, *controller.AcesUpWebController, string) {
	t.Helper()
	mockOutput := `{"columns":[],"stockCount":0,"discardCount":0,"phase":0,"moveCount":0,"message":""}`
	giMock := new(usecase.MockAcesUpInteractor)
	factory := func() uc.AcesUpInteractorIF { return giMock }
	ctrl := controller.NewAcesUpWebController(factory)
	t.Cleanup(func() { ctrl.Stop() })

	return giMock, ctrl, mockOutput
}

func acesupPost(t *testing.T, handler http.HandlerFunc, input *controller.AcesUpWebInput) *recorded {
	t.Helper()
	return execRequest(t, handler, input)
}

func TestAcesUpWebController_Commands(t *testing.T) {
	giMock, ctrl, mockOutput := setupAcesUpWebTest(t)

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
			input := &controller.AcesUpWebInput{
				BaseWebInput: controller.BaseWebInput{Command: tt.command, SessionID: "s1"},
			}
			rec := acesupPost(t, ctrl.Exec, input)
			rec.CodeIs(http.StatusOK)
		})
	}
}

func TestAcesUpWebController_Remove(t *testing.T) {
	giMock, ctrl, mockOutput := setupAcesUpWebTest(t)
	giMock.On("Reset").Return(mockOutput)
	giMock.On("Remove", 3).Return(mockOutput)

	acesupPost(t, ctrl.Exec, &controller.AcesUpWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s2"},
	})

	rec := acesupPost(t, ctrl.Exec, &controller.AcesUpWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "remove", SessionID: "s2"},
		Col:          acesupIntPtr(3),
	})
	rec.CodeIs(http.StatusOK)
}

func TestAcesUpWebController_Remove_ShortAlias(t *testing.T) {
	giMock, ctrl, mockOutput := setupAcesUpWebTest(t)
	giMock.On("Reset").Return(mockOutput)
	giMock.On("Remove", 2).Return(mockOutput)

	acesupPost(t, ctrl.Exec, &controller.AcesUpWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s5"},
	})

	rec := acesupPost(t, ctrl.Exec, &controller.AcesUpWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "rm", SessionID: "s5"},
		Col:          acesupIntPtr(2),
	})
	rec.CodeIs(http.StatusOK)
}

func TestAcesUpWebController_Remove_MissingParams(t *testing.T) {
	giMock, ctrl, mockOutput := setupAcesUpWebTest(t)
	giMock.On("Reset").Return(mockOutput)

	acesupPost(t, ctrl.Exec, &controller.AcesUpWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s3"},
	})

	rec := acesupPost(t, ctrl.Exec, &controller.AcesUpWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "remove", SessionID: "s3"},
	})
	rec.CodeIs(http.StatusBadRequest)
}

func TestAcesUpWebController_Move(t *testing.T) {
	giMock, ctrl, mockOutput := setupAcesUpWebTest(t)
	giMock.On("Reset").Return(mockOutput)
	giMock.On("Move", 1).Return(mockOutput)

	acesupPost(t, ctrl.Exec, &controller.AcesUpWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s6"},
	})

	t.Run("move with col", func(t *testing.T) {
		rec := acesupPost(t, ctrl.Exec, &controller.AcesUpWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s6"},
			Col:          acesupIntPtr(1),
		})
		rec.CodeIs(http.StatusOK)
	})

	t.Run("move short alias", func(t *testing.T) {
		rec := acesupPost(t, ctrl.Exec, &controller.AcesUpWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "mv", SessionID: "s6"},
			Col:          acesupIntPtr(1),
		})
		rec.CodeIs(http.StatusOK)
	})

	t.Run("move missing col", func(t *testing.T) {
		rec := acesupPost(t, ctrl.Exec, &controller.AcesUpWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s6"},
		})
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestAcesUpWebController_UndoN(t *testing.T) {
	giMock, ctrl, mockOutput := setupAcesUpWebTest(t)
	giMock.On("UndoN", 3).Return(mockOutput)

	t.Run("undo_n with valid n", func(t *testing.T) {
		n := 3
		rec := acesupPost(t, ctrl.Exec, &controller.AcesUpWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1", N: &n},
		})
		rec.CodeIs(http.StatusOK)
	})

	t.Run("undo_n with missing n", func(t *testing.T) {
		rec := acesupPost(t, ctrl.Exec, &controller.AcesUpWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1"},
		})
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestAcesUpWebController_UnknownCommand(t *testing.T) {
	_, ctrl, _ := setupAcesUpWebTest(t)

	rec := acesupPost(t, ctrl.Exec, &controller.AcesUpWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "unknown", SessionID: "s4"},
	})
	rec.CodeIs(http.StatusBadRequest)
}
