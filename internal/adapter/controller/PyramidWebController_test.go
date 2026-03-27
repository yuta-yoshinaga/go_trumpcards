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

func mustPyramidOutputJSON(msg string) string {
	out := &controller.PyramidWebOutput{
		Pyramid:       [][]*controller.PyramidWebOutputCard{},
		Waste:         []*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustPyramidOutputJSON: %v", err))
	}
	return string(b)
}

func pyramidIntPtr(v int) *int { return &v }

func setupPyramidWebTest(t *testing.T) (*usecase.MockPyramidInteractor, *controller.PyramidWebController, string) {
	t.Helper()
	mockOutput := `{"pyramid":[],"stockCount":0,"waste":[],"phase":0,"moveCount":0,"message":""}`
	piMock := new(usecase.MockPyramidInteractor)
	factory := func() uc.PyramidInteractorIF { return piMock }
	ctrl := controller.NewPyramidWebController(factory)
	t.Cleanup(func() { ctrl.Stop() })

	return piMock, ctrl, mockOutput
}

func pyramidPost(t *testing.T, handler http.HandlerFunc, body string) *recorded {
	t.Helper()
	var input controller.PyramidWebInput
	_ = json.Unmarshal([]byte(body), &input)
	return execRequest(t, handler, &input)
}

func pyramidPostInput(t *testing.T, handler http.HandlerFunc, input controller.PyramidWebInput) *recorded {
	t.Helper()
	return execRequest(t, handler, &input)
}

func TestPyramidWebController_Commands(t *testing.T) {
	piMock, ctrl, mockOutput := setupPyramidWebTest(t)

	piMock.On("Reset").Return(mockOutput)
	piMock.On("Draw").Return(mockOutput)
	piMock.On("GiveUp").Return(mockOutput)
	piMock.On("Hint").Return(mockOutput)
	piMock.On("ActionLog").Return(mockOutput)
	piMock.On("Undo").Return(mockOutput)

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
			rec := pyramidPost(t, ctrl.Exec, tt.command)
			rec.CodeIs(http.StatusOK)
		})
	}
}

func TestPyramidWebController_Quit(t *testing.T) {
	_, ctrl, _ := setupPyramidWebTest(t)
	rec := pyramidPost(t, ctrl.Exec, `{"command":"q","sessionId":"s1"}`)
	rec.CodeIs(http.StatusOK)
	rec.BodyIs(mustPyramidOutputJSON("bye."))
}

func TestPyramidWebController_RemoveKing(t *testing.T) {
	piMock, ctrl, mockOutput := setupPyramidWebTest(t)
	piMock.On("RemoveKing", 6, 0).Return(mockOutput)

	rec := pyramidPostInput(t, ctrl.Exec, controller.PyramidWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "rm", SessionID: "s1"},
		Card1:        &controller.PyramidWebCard{Zone: "pyramid", Row: pyramidIntPtr(6), Col: pyramidIntPtr(0)},
	})
	rec.CodeIs(http.StatusOK)
}

func TestPyramidWebController_RemovePair(t *testing.T) {
	piMock, ctrl, mockOutput := setupPyramidWebTest(t)
	piMock.On("RemovePair", 6, 0, 6, 1).Return(mockOutput)

	rec := pyramidPostInput(t, ctrl.Exec, controller.PyramidWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "rm", SessionID: "s1"},
		Card1:        &controller.PyramidWebCard{Zone: "pyramid", Row: pyramidIntPtr(6), Col: pyramidIntPtr(0)},
		Card2:        &controller.PyramidWebCard{Zone: "pyramid", Row: pyramidIntPtr(6), Col: pyramidIntPtr(1)},
	})
	rec.CodeIs(http.StatusOK)
}

func TestPyramidWebController_RemoveWithWaste(t *testing.T) {
	piMock, ctrl, mockOutput := setupPyramidWebTest(t)
	piMock.On("RemoveWithWaste", 6, 0).Return(mockOutput)

	rec := pyramidPostInput(t, ctrl.Exec, controller.PyramidWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "rm", SessionID: "s1"},
		Card1:        &controller.PyramidWebCard{Zone: "waste"},
		Card2:        &controller.PyramidWebCard{Zone: "pyramid", Row: pyramidIntPtr(6), Col: pyramidIntPtr(0)},
	})
	rec.CodeIs(http.StatusOK)
}

func TestPyramidWebController_RemoveWasteKing(t *testing.T) {
	piMock, ctrl, mockOutput := setupPyramidWebTest(t)
	piMock.On("RemoveWasteKing").Return(mockOutput)

	rec := pyramidPostInput(t, ctrl.Exec, controller.PyramidWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "rm", SessionID: "s1"},
		Card1:        &controller.PyramidWebCard{Zone: "waste"},
	})
	rec.CodeIs(http.StatusOK)
}

func TestPyramidWebController_RemoveNoCard1(t *testing.T) {
	_, ctrl, _ := setupPyramidWebTest(t)
	rec := pyramidPostInput(t, ctrl.Exec, controller.PyramidWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "rm", SessionID: "s1"},
	})
	rec.CodeIs(http.StatusBadRequest)
}

func TestPyramidWebController_UnknownCommand(t *testing.T) {
	_, ctrl, _ := setupPyramidWebTest(t)
	rec := pyramidPost(t, ctrl.Exec, `{"command":"xyz","sessionId":"s1"}`)
	rec.CodeIs(http.StatusBadRequest)
}
