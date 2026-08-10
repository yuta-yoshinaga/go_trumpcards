//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustCribbageSquaresOutputJSON(msg string) string {
	out := &controller.CribbageSquaresWebOutput{
		Board:         [][]*controller.CribbageSquaresWebOutputCard{},
		RowScores:     []int{},
		ColScores:     []int{},
		RowDetails:    []*controller.CribbageSquaresWebOutputScore{},
		ColDetails:    []*controller.CribbageSquaresWebOutputScore{},
		WinScore:      domain.CribbageSquaresWinScore,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustCribbageSquaresOutputJSON: %v", err))
	}
	return string(b)
}

func cribbageSquaresIntPtr(v int) *int { return &v }

func setupCribbageSquaresWebTest(t *testing.T) (*usecase.MockCribbageSquaresInteractor, *controller.CribbageSquaresWebController, string) {
	t.Helper()
	mockOutput := `{"board":[],"placedCount":0,"phase":0,"canUndo":false,"rowScores":[],"colScores":[],"totalScore":0,"message":""}`
	piMock := new(usecase.MockCribbageSquaresInteractor)
	factory := func() uc.CribbageSquaresInteractorIF { return piMock }
	ctrl := controller.NewCribbageSquaresWebController(factory)
	t.Cleanup(func() { ctrl.Stop() })
	return piMock, ctrl, mockOutput
}

func cribbageSquaresPostInput(t *testing.T, handler http.HandlerFunc, input controller.CribbageSquaresWebInput) *recorded {
	t.Helper()
	return execRequest(t, handler, &input)
}

func cribbageSquaresPost(t *testing.T, handler http.HandlerFunc, body string) *recorded {
	t.Helper()
	var input controller.CribbageSquaresWebInput
	_ = json.Unmarshal([]byte(body), &input)
	return execRequest(t, handler, &input)
}

func TestCribbageSquaresWebController_Commands(t *testing.T) {
	piMock, ctrl, mockOutput := setupCribbageSquaresWebTest(t)
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
			rec := cribbageSquaresPost(t, ctrl.Exec, tt.command)
			rec.CodeIs(http.StatusOK)
		})
	}
}

func TestCribbageSquaresWebController_Place(t *testing.T) {
	piMock, ctrl, mockOutput := setupCribbageSquaresWebTest(t)
	piMock.On("Place", 1, 2).Return(mockOutput)

	rec := cribbageSquaresPostInput(t, ctrl.Exec, controller.CribbageSquaresWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "place", SessionID: "s1"},
		Row:          cribbageSquaresIntPtr(1),
		Col:          cribbageSquaresIntPtr(2),
	})
	rec.CodeIs(http.StatusOK)
}

func TestCribbageSquaresWebController_PlaceMissingCoords(t *testing.T) {
	_, ctrl, _ := setupCribbageSquaresWebTest(t)
	rec := cribbageSquaresPostInput(t, ctrl.Exec, controller.CribbageSquaresWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "place", SessionID: "s1"},
	})
	rec.CodeIs(http.StatusBadRequest)
}

func TestCribbageSquaresWebController_Quit(t *testing.T) {
	_, ctrl, _ := setupCribbageSquaresWebTest(t)
	rec := cribbageSquaresPost(t, ctrl.Exec, `{"command":"q","sessionId":"s1"}`)
	rec.CodeIs(http.StatusOK)
	rec.BodyIs(mustCribbageSquaresOutputJSON("bye."))
}

func TestCribbageSquaresWebController_UnknownCommand(t *testing.T) {
	_, ctrl, _ := setupCribbageSquaresWebTest(t)
	rec := cribbageSquaresPost(t, ctrl.Exec, `{"command":"xyz","sessionId":"s1"}`)
	rec.CodeIs(http.StatusBadRequest)
}
