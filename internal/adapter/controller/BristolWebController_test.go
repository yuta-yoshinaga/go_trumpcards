//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustBristolOutputJSON(msg string) string {
	out := &controller.BristolWebOutput{
		Tableau:       [][]*controller.WebOutputCard{},
		Fan:           [][]*controller.WebOutputCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBristolOutputJSON: %v", err))
	}
	return string(b)
}

func bristolIntPtr(v int) *int { return &v }

func TestBristolWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"fan":[],"stockCount":0,"foundation":[],"phase":0,"moveCount":0,"canUndo":false,"message":""}`
	expectedBody := mockOutput

	biMock := new(usecase.MockBristolInteractor)
	biMock.On("Reset").Return(mockOutput)
	biMock.On("Draw").Return(mockOutput)
	biMock.On("GiveUp").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("AutoComplete").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)
	biMock.On("MoveTableauToTableau", 0, 1).Return(mockOutput)
	biMock.On("MoveTableauToFoundation", 2).Return(mockOutput)
	biMock.On("MoveFanToTableau", 1, 3).Return(mockOutput)
	biMock.On("MoveFanToFoundation", 0).Return(mockOutput)

	factory := func() uc.BristolInteractorIF { return biMock }
	ctrl := controller.NewBristolWebController(factory)
	defer ctrl.Stop()

	t.Run("quit", func(t *testing.T) {
		var input controller.BristolWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustBristolOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.BristolWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("draw", func(t *testing.T) {
		var input controller.BristolWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup", func(t *testing.T) {
		var input controller.BristolWebInput
		_ = json.Unmarshal([]byte(`{"command":"giveup","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.BristolWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("autocomplete", func(t *testing.T) {
		var input controller.BristolWebInput
		_ = json.Unmarshal([]byte(`{"command":"ac","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.BristolWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move tableau to tableau", func(t *testing.T) {
		input := controller.BristolWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
			From:         &controller.BristolWebZone{Zone: "tableau", Col: bristolIntPtr(0)},
			To:           &controller.BristolWebZone{Zone: "tableau", Col: bristolIntPtr(1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move tableau to foundation", func(t *testing.T) {
		input := controller.BristolWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.BristolWebZone{Zone: "tableau", Col: bristolIntPtr(2)},
			To:           &controller.BristolWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move fan to tableau", func(t *testing.T) {
		input := controller.BristolWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.BristolWebZone{Zone: "fan", Col: bristolIntPtr(1)},
			To:           &controller.BristolWebZone{Zone: "tableau", Col: bristolIntPtr(3)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move fan to foundation", func(t *testing.T) {
		input := controller.BristolWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.BristolWebZone{Zone: "fan", Col: bristolIntPtr(0)},
			To:           &controller.BristolWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.BristolWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustBristolOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		var input controller.BristolWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustBristolOutputJSON("param error."))
	})

	t.Run("session too long", func(t *testing.T) {
		input := controller.BristolWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustBristolOutputJSON("param error."))
	})
}

func TestBristolWebController_MoveErrors(t *testing.T) {
	biMock := new(usecase.MockBristolInteractor)
	factory := func() uc.BristolInteractorIF { return biMock }
	ctrl := controller.NewBristolWebController(factory)
	defer ctrl.Stop()

	t.Run("move without from/to", func(t *testing.T) {
		var input controller.BristolWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau to tableau missing cols", func(t *testing.T) {
		input := controller.BristolWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.BristolWebZone{Zone: "tableau"},
			To:           &controller.BristolWebZone{Zone: "tableau"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau to foundation missing col", func(t *testing.T) {
		input := controller.BristolWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.BristolWebZone{Zone: "tableau"},
			To:           &controller.BristolWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move fan to tableau missing cols", func(t *testing.T) {
		input := controller.BristolWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.BristolWebZone{Zone: "fan"},
			To:           &controller.BristolWebZone{Zone: "tableau"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zones", func(t *testing.T) {
		input := controller.BristolWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.BristolWebZone{Zone: "foundation"},
			To:           &controller.BristolWebZone{Zone: "tableau"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestBristolWebController_Undo(t *testing.T) {
	mockOutput := `{"tableau":[],"fan":[],"stockCount":0,"foundation":[],"phase":0,"moveCount":0,"canUndo":false,"message":""}`

	biMock := new(usecase.MockBristolInteractor)
	biMock.On("Undo").Return(mockOutput)

	factory := func() uc.BristolInteractorIF { return biMock }
	ctrl := controller.NewBristolWebController(factory)
	defer ctrl.Stop()

	var input controller.BristolWebInput
	_ = json.Unmarshal([]byte(`{"command":"u","sessionId":"s1"}`), &input)
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	recorded.BodyIs(mockOutput)
}

func TestBristolWebController_UndoN(t *testing.T) {
	mockOutput := `{"tableau":[],"fan":[],"stockCount":0,"foundation":[],"phase":0,"moveCount":0,"canUndo":false,"message":""}`

	biMock := new(usecase.MockBristolInteractor)
	biMock.On("UndoN", 3).Return(mockOutput)

	factory := func() uc.BristolInteractorIF { return biMock }
	ctrl := controller.NewBristolWebController(factory)
	defer ctrl.Stop()

	t.Run("undo_n valid", func(t *testing.T) {
		n := 3
		input := controller.BristolWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1", N: &n},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("undo_n missing n", func(t *testing.T) {
		input := controller.BristolWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestBristolWebController_Stop(t *testing.T) {
	biMock := new(usecase.MockBristolInteractor)
	factory := func() uc.BristolInteractorIF { return biMock }
	c := controller.NewBristolWebController(factory)
	c.Stop()
	c.Stop()
}
