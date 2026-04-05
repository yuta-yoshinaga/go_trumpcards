//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func triPeaksIntPtr(v int) *int { return &v }

func setupTriPeaksWebTest(t *testing.T) (*usecase.MockTriPeaksInteractor, *controller.TriPeaksWebController, string) {
	t.Helper()
	mockOutput := `{"layout":[],"stockCount":0,"waste":[],"phase":0,"moveCount":0,"message":""}`
	tiMock := new(usecase.MockTriPeaksInteractor)
	factory := func() uc.TriPeaksInteractorIF { return tiMock }
	ctrl := controller.NewTriPeaksWebController(factory)
	t.Cleanup(func() { ctrl.Stop() })

	return tiMock, ctrl, mockOutput
}

func triPeaksPost(t *testing.T, handler http.HandlerFunc, body string) *recorded {
	t.Helper()
	var input controller.TriPeaksWebInput
	_ = json.Unmarshal([]byte(body), &input)
	return execRequest(t, handler, &input)
}

func TestTriPeaksWebController_Commands(t *testing.T) {
	tiMock, ctrl, mockOutput := setupTriPeaksWebTest(t)

	tiMock.On("Reset").Return(mockOutput)
	tiMock.On("Draw").Return(mockOutput)
	tiMock.On("GiveUp").Return(mockOutput)
	tiMock.On("Hint").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)
	tiMock.On("Undo").Return(mockOutput)

	tests := []struct {
		name    string
		command string
	}{
		{"reset", `{"command":"reset","sessionId":"s1"}`},
		{"draw", `{"command":"draw","sessionId":"s1"}`},
		{"giveup", `{"command":"giveup","sessionId":"s1"}`},
		{"hint", `{"command":"hint","sessionId":"s1"}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
		{"undo", `{"command":"undo","sessionId":"s1"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := triPeaksPost(t, ctrl.Exec, tt.command)
			rec.CodeIs(http.StatusOK)
		})
	}
}

func TestTriPeaksWebController_Remove(t *testing.T) {
	tiMock, ctrl, mockOutput := setupTriPeaksWebTest(t)
	tiMock.On("Reset").Return(mockOutput)
	tiMock.On("Remove", 3, 0).Return(mockOutput)

	// First reset to create session
	triPeaksPost(t, ctrl.Exec, `{"command":"reset","sessionId":"s2"}`)

	// Remove with row and col
	input := controller.TriPeaksWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "remove", SessionID: "s2"},
		Row:          triPeaksIntPtr(3),
		Col:          triPeaksIntPtr(0),
	}
	rec := execRequest(t, ctrl.Exec, &input)
	rec.CodeIs(http.StatusOK)
}

func TestTriPeaksWebController_Remove_MissingParams(t *testing.T) {
	tiMock, ctrl, mockOutput := setupTriPeaksWebTest(t)
	tiMock.On("Reset").Return(mockOutput)

	// First reset to create session
	triPeaksPost(t, ctrl.Exec, `{"command":"reset","sessionId":"s3"}`)

	// Remove without row/col
	input := controller.TriPeaksWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "remove", SessionID: "s3"},
	}
	rec := execRequest(t, ctrl.Exec, &input)
	rec.CodeIs(http.StatusBadRequest)
}

func TestTriPeaksWebController_UndoN(t *testing.T) {
	tiMock, ctrl, mockOutput := setupTriPeaksWebTest(t)
	tiMock.On("UndoN", 3).Return(mockOutput)

	t.Run("undo_n with valid n", func(t *testing.T) {
		n := 3
		input := controller.TriPeaksWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1", N: &n},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("undo_n with missing n", func(t *testing.T) {
		input := controller.TriPeaksWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1"},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestTriPeaksWebController_UnknownCommand(t *testing.T) {
	_, ctrl, _ := setupTriPeaksWebTest(t)

	rec := triPeaksPost(t, ctrl.Exec, `{"command":"unknown","sessionId":"s4"}`)
	rec.CodeIs(http.StatusBadRequest)
}
