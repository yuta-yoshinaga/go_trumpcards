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

func mustPokerSquaresOutputJSON(msg string) string {
	out := &controller.PokerSquaresWebOutput{
		Board:         [][]*controller.PokerSquaresWebOutputCard{},
		RowScores:     []int{},
		ColScores:     []int{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustPokerSquaresOutputJSON: %v", err))
	}
	return string(b)
}

func pokerSquaresIntPtr(v int) *int { return &v }

func setupPokerSquaresWebTest(t *testing.T) (*usecase.MockPokerSquaresInteractor, *controller.PokerSquaresWebController, string) {
	t.Helper()
	mockOutput := `{"board":[],"placedCount":0,"phase":0,"canUndo":false,"rowScores":[],"colScores":[],"totalScore":0,"message":""}`
	piMock := new(usecase.MockPokerSquaresInteractor)
	factory := func() uc.PokerSquaresInteractorIF { return piMock }
	ctrl := controller.NewPokerSquaresWebController(factory)
	t.Cleanup(func() { ctrl.Stop() })
	return piMock, ctrl, mockOutput
}

func pokerSquaresPostInput(t *testing.T, handler http.HandlerFunc, input controller.PokerSquaresWebInput) *recorded {
	t.Helper()
	return execRequest(t, handler, &input)
}

func pokerSquaresPost(t *testing.T, handler http.HandlerFunc, body string) *recorded {
	t.Helper()
	var input controller.PokerSquaresWebInput
	_ = json.Unmarshal([]byte(body), &input)
	return execRequest(t, handler, &input)
}

func TestPokerSquaresWebController_Commands(t *testing.T) {
	piMock, ctrl, mockOutput := setupPokerSquaresWebTest(t)
	piMock.On("Reset").Return(mockOutput)
	piMock.On("Undo").Return(mockOutput)
	piMock.On("GiveUp").Return(mockOutput)
	piMock.On("ActionLog").Return(mockOutput)
	// **hint は Web からも呼べる (#4790)。**シナジー考慮ヒントは CUI しか
	// 受け取れていなかった。
	piMock.On("Hint").Return(mockOutput)

	tests := []struct {
		name    string
		command string
	}{
		{"reset", `{"command":"reset","sessionId":"s1"}`},
		{"undo", `{"command":"undo","sessionId":"s1"}`},
		{"giveup", `{"command":"giveup","sessionId":"s1"}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
		{"hint", `{"command":"hint","sessionId":"s1"}`},
		{"hint short", `{"command":"h","sessionId":"s1"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := pokerSquaresPost(t, ctrl.Exec, tt.command)
			rec.CodeIs(http.StatusOK)
		})
	}
}

func TestPokerSquaresWebController_Place(t *testing.T) {
	piMock, ctrl, mockOutput := setupPokerSquaresWebTest(t)
	piMock.On("Place", 1, 2).Return(mockOutput)

	rec := pokerSquaresPostInput(t, ctrl.Exec, controller.PokerSquaresWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "place", SessionID: "s1"},
		Row:          pokerSquaresIntPtr(1),
		Col:          pokerSquaresIntPtr(2),
	})
	rec.CodeIs(http.StatusOK)
}

func TestPokerSquaresWebController_PlaceMissingCoords(t *testing.T) {
	_, ctrl, _ := setupPokerSquaresWebTest(t)
	rec := pokerSquaresPostInput(t, ctrl.Exec, controller.PokerSquaresWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "place", SessionID: "s1"},
	})
	rec.CodeIs(http.StatusBadRequest)
}

func TestPokerSquaresWebController_Quit(t *testing.T) {
	_, ctrl, _ := setupPokerSquaresWebTest(t)
	rec := pokerSquaresPost(t, ctrl.Exec, `{"command":"q","sessionId":"s1"}`)
	rec.CodeIs(http.StatusOK)
	rec.BodyIs(mustPokerSquaresOutputJSON("bye."))
}

func TestPokerSquaresWebController_UnknownCommand(t *testing.T) {
	_, ctrl, _ := setupPokerSquaresWebTest(t)
	rec := pokerSquaresPost(t, ctrl.Exec, `{"command":"xyz","sessionId":"s1"}`)
	rec.CodeIs(http.StatusBadRequest)
}
