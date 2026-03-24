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

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
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

func setupPyramidWebTest(t *testing.T) (*usecase.MockPyramidInteractor, *rest.Api, string) {
	t.Helper()
	mockOutput := `{"pyramid":[],"stockCount":0,"waste":[],"phase":0,"moveCount":0,"message":""}`
	piMock := new(usecase.MockPyramidInteractor)
	factory := func() uc.PyramidInteractorIF { return piMock }
	ctrl := controller.NewPyramidWebController(factory)
	t.Cleanup(func() { ctrl.Stop() })

	api := rest.NewApi()
	router, _ := rest.MakeRouter(
		rest.Post("/pyramid/exec", ctrl.Exec),
	)
	api.SetApp(router)
	return piMock, api, mockOutput
}

func pyramidPost(t *testing.T, api *rest.Api, body string) *test.Recorded {
	t.Helper()
	var input controller.PyramidWebInput
	_ = json.Unmarshal([]byte(body), &input)
	req := test.MakeSimpleRequest("POST", "http://1.2.3.4/pyramid/exec", &input)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	return test.RunRequest(t, api.MakeHandler(), req)
}

func pyramidPostInput(t *testing.T, api *rest.Api, input controller.PyramidWebInput) *test.Recorded {
	t.Helper()
	req := test.MakeSimpleRequest("POST", "http://1.2.3.4/pyramid/exec", &input)
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	return test.RunRequest(t, api.MakeHandler(), req)
}

func TestPyramidWebController_Commands(t *testing.T) {
	piMock, api, mockOutput := setupPyramidWebTest(t)

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
			recorded := pyramidPost(t, api, tt.command)
			recorded.CodeIs(http.StatusOK)
		})
	}
}

func TestPyramidWebController_Quit(t *testing.T) {
	_, api, _ := setupPyramidWebTest(t)
	recorded := pyramidPost(t, api, `{"command":"q","sessionId":"s1"}`)
	recorded.CodeIs(http.StatusOK)
	recorded.BodyIs(mustPyramidOutputJSON("bye."))
}

func TestPyramidWebController_RemoveKing(t *testing.T) {
	piMock, api, mockOutput := setupPyramidWebTest(t)
	piMock.On("RemoveKing", 6, 0).Return(mockOutput)

	recorded := pyramidPostInput(t, api, controller.PyramidWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "rm", SessionID: "s1"},
		Card1:        &controller.PyramidWebCard{Zone: "pyramid", Row: pyramidIntPtr(6), Col: pyramidIntPtr(0)},
	})
	recorded.CodeIs(http.StatusOK)
}

func TestPyramidWebController_RemovePair(t *testing.T) {
	piMock, api, mockOutput := setupPyramidWebTest(t)
	piMock.On("RemovePair", 6, 0, 6, 1).Return(mockOutput)

	recorded := pyramidPostInput(t, api, controller.PyramidWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "rm", SessionID: "s1"},
		Card1:        &controller.PyramidWebCard{Zone: "pyramid", Row: pyramidIntPtr(6), Col: pyramidIntPtr(0)},
		Card2:        &controller.PyramidWebCard{Zone: "pyramid", Row: pyramidIntPtr(6), Col: pyramidIntPtr(1)},
	})
	recorded.CodeIs(http.StatusOK)
}

func TestPyramidWebController_RemoveWithWaste(t *testing.T) {
	piMock, api, mockOutput := setupPyramidWebTest(t)
	piMock.On("RemoveWithWaste", 6, 0).Return(mockOutput)

	recorded := pyramidPostInput(t, api, controller.PyramidWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "rm", SessionID: "s1"},
		Card1:        &controller.PyramidWebCard{Zone: "waste"},
		Card2:        &controller.PyramidWebCard{Zone: "pyramid", Row: pyramidIntPtr(6), Col: pyramidIntPtr(0)},
	})
	recorded.CodeIs(http.StatusOK)
}

func TestPyramidWebController_RemoveWasteKing(t *testing.T) {
	piMock, api, mockOutput := setupPyramidWebTest(t)
	piMock.On("RemoveWasteKing").Return(mockOutput)

	recorded := pyramidPostInput(t, api, controller.PyramidWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "rm", SessionID: "s1"},
		Card1:        &controller.PyramidWebCard{Zone: "waste"},
	})
	recorded.CodeIs(http.StatusOK)
}

func TestPyramidWebController_RemoveNoCard1(t *testing.T) {
	_, api, _ := setupPyramidWebTest(t)
	recorded := pyramidPostInput(t, api, controller.PyramidWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "rm", SessionID: "s1"},
	})
	recorded.CodeIs(http.StatusBadRequest)
}

func TestPyramidWebController_UnknownCommand(t *testing.T) {
	_, api, _ := setupPyramidWebTest(t)
	recorded := pyramidPost(t, api, `{"command":"xyz","sessionId":"s1"}`)
	recorded.CodeIs(http.StatusBadRequest)
}
