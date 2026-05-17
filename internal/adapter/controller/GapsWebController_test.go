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

func mustGapsOutputJSON(msg string) string {
	out := &controller.GapsWebOutput{
		Grid:          [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustGapsOutputJSON: %v", err))
	}
	return string(b)
}

func gapsIntPtr(v int) *int { return &v }

func setupGapsWebTest(t *testing.T) (*usecase.MockGapsInteractor, *controller.GapsWebController, string) {
	t.Helper()
	mockOutput := `{"grid":[],"redealsUsed":0,"redealsRemaining":3,"phase":0,"moveCount":0,"canUndo":false,"isStalemate":false,"undoToEscape":0,"message":""}`
	giMock := new(usecase.MockGapsInteractor)
	factory := func() uc.GapsInteractorIF { return giMock }
	ctrl := controller.NewGapsWebController(factory)
	t.Cleanup(func() { ctrl.Stop() })
	return giMock, ctrl, mockOutput
}

func gapsPostInput(t *testing.T, handler http.HandlerFunc, input controller.GapsWebInput) *recorded {
	t.Helper()
	return execRequest(t, handler, &input)
}

func TestGapsWebController_Commands(t *testing.T) {
	giMock, ctrl, mockOutput := setupGapsWebTest(t)

	giMock.On("Reset").Return(mockOutput)
	giMock.On("Redeal").Return(mockOutput)
	giMock.On("GiveUp").Return(mockOutput)
	giMock.On("Hint").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)
	giMock.On("Undo").Return(mockOutput)

	for _, tt := range []struct {
		name string
		cmd  string
	}{
		{"reset", "reset"}, {"redeal", "redeal"}, {"giveup", "giveup"},
		{"hint", "hint"}, {"log", "log"}, {"undo", "undo"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rec := gapsPostInput(t, ctrl.Exec, controller.GapsWebInput{
				BaseWebInput: controller.BaseWebInput{Command: tt.cmd, SessionID: "s1"},
			})
			rec.CodeIs(http.StatusOK)
		})
	}
}

func TestGapsWebController_Move(t *testing.T) {
	giMock, ctrl, mockOutput := setupGapsWebTest(t)
	giMock.On("Move", 0, 1, 2, 3).Return(mockOutput)
	rec := gapsPostInput(t, ctrl.Exec, controller.GapsWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
		From:         &controller.GapsWebZone{Zone: "grid", Row: gapsIntPtr(0), Col: gapsIntPtr(1)},
		To:           &controller.GapsWebZone{Zone: "grid", Row: gapsIntPtr(2), Col: gapsIntPtr(3)},
	})
	rec.CodeIs(http.StatusOK)
}

func TestGapsWebController_Move_MissingFromTo(t *testing.T) {
	_, ctrl, _ := setupGapsWebTest(t)
	rec := gapsPostInput(t, ctrl.Exec, controller.GapsWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
	})
	rec.CodeIs(http.StatusBadRequest)
}

func TestGapsWebController_Move_MissingRowCol(t *testing.T) {
	_, ctrl, _ := setupGapsWebTest(t)
	rec := gapsPostInput(t, ctrl.Exec, controller.GapsWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
		From:         &controller.GapsWebZone{Zone: "grid"},
		To:           &controller.GapsWebZone{Zone: "grid", Row: gapsIntPtr(0), Col: gapsIntPtr(0)},
	})
	rec.CodeIs(http.StatusBadRequest)
}

func TestGapsWebController_UndoN(t *testing.T) {
	giMock, ctrl, mockOutput := setupGapsWebTest(t)
	giMock.On("UndoN", 2).Return(mockOutput)
	n := 2
	rec := gapsPostInput(t, ctrl.Exec, controller.GapsWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1", N: &n},
	})
	rec.CodeIs(http.StatusOK)
}

func TestGapsWebController_UndoN_MissingN(t *testing.T) {
	_, ctrl, _ := setupGapsWebTest(t)
	rec := gapsPostInput(t, ctrl.Exec, controller.GapsWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1"},
	})
	rec.CodeIs(http.StatusBadRequest)
}

func TestGapsWebController_Quit(t *testing.T) {
	_, ctrl, _ := setupGapsWebTest(t)
	rec := gapsPostInput(t, ctrl.Exec, controller.GapsWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "q", SessionID: "s1"},
	})
	rec.CodeIs(http.StatusOK)
	rec.BodyIs(mustGapsOutputJSON("bye."))
}
