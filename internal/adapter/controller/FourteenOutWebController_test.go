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

func mustFourteenOutOutputJSON(msg string) string {
	out := &controller.FourteenOutWebOutput{
		Columns:       [][]*controller.FourteenOutWebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustFourteenOutOutputJSON: %v", err))
	}
	return string(b)
}

func fourteenOutIntPtr(v int) *int { return &v }

func setupFourteenOutWebTest(t *testing.T) (*usecase.MockFourteenOutInteractor, *controller.FourteenOutWebController, string) {
	t.Helper()
	mockOutput := `{"columns":[],"phase":0,"removedCount":0,"removablePairs":0,"canUndo":false,"isStalemate":false,"message":""}`
	miMock := new(usecase.MockFourteenOutInteractor)
	factory := func() uc.FourteenOutInteractorIF { return miMock }
	ctrl := controller.NewFourteenOutWebController(factory)
	t.Cleanup(func() { ctrl.Stop() })
	return miMock, ctrl, mockOutput
}

func fourteenOutPostInput(t *testing.T, handler http.HandlerFunc, input controller.FourteenOutWebInput) *recorded {
	t.Helper()
	return execRequest(t, handler, &input)
}

func fourteenOutPost(t *testing.T, handler http.HandlerFunc, body string) *recorded {
	t.Helper()
	var input controller.FourteenOutWebInput
	_ = json.Unmarshal([]byte(body), &input)
	return execRequest(t, handler, &input)
}

func TestFourteenOutWebController_Commands(t *testing.T) {
	miMock, ctrl, mockOutput := setupFourteenOutWebTest(t)
	miMock.On("Reset").Return(mockOutput)
	miMock.On("Undo").Return(mockOutput)
	miMock.On("GiveUp").Return(mockOutput)
	miMock.On("Hint").Return(mockOutput)
	miMock.On("ActionLog").Return(mockOutput)

	tests := []struct {
		name    string
		command string
	}{
		{"reset", `{"command":"reset","sessionId":"s1"}`},
		{"undo", `{"command":"undo","sessionId":"s1"}`},
		{"giveup", `{"command":"giveup","sessionId":"s1"}`},
		{"hint", `{"command":"hint","sessionId":"s1"}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := fourteenOutPost(t, ctrl.Exec, tt.command)
			rec.CodeIs(http.StatusOK)
		})
	}
}

func TestFourteenOutWebController_Remove(t *testing.T) {
	miMock, ctrl, mockOutput := setupFourteenOutWebTest(t)
	miMock.On("Remove", 0, 3).Return(mockOutput)

	rec := fourteenOutPostInput(t, ctrl.Exec, controller.FourteenOutWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "remove", SessionID: "s1"},
		FromCol:      fourteenOutIntPtr(0),
		ToCol:        fourteenOutIntPtr(3),
	})
	rec.CodeIs(http.StatusOK)
	// **列番号がそのままドメインに届くこと。**動かせるのは末尾だけなので、
	// クローン元のような (行,列) x2 は要らない。
	miMock.AssertCalled(t, "Remove", 0, 3)
}

func TestFourteenOutWebController_RemoveMissingColumns(t *testing.T) {
	_, ctrl, _ := setupFourteenOutWebTest(t)
	rec := fourteenOutPostInput(t, ctrl.Exec, controller.FourteenOutWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "remove", SessionID: "s1"},
		FromCol:      fourteenOutIntPtr(0),
	})
	rec.CodeIs(http.StatusBadRequest)
}

// **補充コマンドは存在しない。**山札が無いのに 200 を返すと、盤が変わらない
// 無言の no-op になる。
func TestFourteenOutWebController_DealIsNotACommand(t *testing.T) {
	_, ctrl, _ := setupFourteenOutWebTest(t)
	fourteenOutPost(t, ctrl.Exec, `{"command":"deal","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	fourteenOutPost(t, ctrl.Exec, `{"command":"d","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
}

func TestFourteenOutWebController_Quit(t *testing.T) {
	_, ctrl, _ := setupFourteenOutWebTest(t)
	rec := fourteenOutPost(t, ctrl.Exec, `{"command":"q","sessionId":"s1"}`)
	rec.CodeIs(http.StatusOK)
	rec.BodyIs(mustFourteenOutOutputJSON("bye."))
}

func TestFourteenOutWebController_UnknownCommand(t *testing.T) {
	_, ctrl, _ := setupFourteenOutWebTest(t)
	rec := fourteenOutPost(t, ctrl.Exec, `{"command":"xyz","sessionId":"s1"}`)
	rec.CodeIs(http.StatusBadRequest)
}
