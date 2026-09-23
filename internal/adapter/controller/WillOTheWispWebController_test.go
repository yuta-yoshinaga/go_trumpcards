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

func mustWillOTheWispOutputJSON(msg string) string {
	out := &controller.WillOTheWispWebOutput{
		Tableau:       [][]*controller.WillOTheWispWebOutputTableauCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustWillOTheWispOutputJSON: %v", err))
	}
	return string(b)
}

func TestWillOTheWispWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"completedSuits":0,"score":0,"phase":0,"moveCount":0,"canUndo":false,"isStalemate":false,"message":""}`

	siMock := new(usecase.MockWillOTheWispInteractor)
	siMock.On("Reset").Return(mockOutput)
	siMock.On("Deal").Return(mockOutput)
	siMock.On("GiveUp").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("AutoComplete").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)
	siMock.On("MoveTableauToTableau", 0, 2, 4).Return(mockOutput)
	siMock.On("Undo").Return(mockOutput)

	factory := func() uc.WillOTheWispInteractorIF { return siMock }
	ctrl := controller.NewWillOTheWispWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.WillOTheWispWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustWillOTheWispOutputJSON("bye."))
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.WillOTheWispWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("deal d", func(t *testing.T) {
		var input controller.WillOTheWispWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("giveup g", func(t *testing.T) {
		var input controller.WillOTheWispWebInput
		_ = json.Unmarshal([]byte(`{"command":"g","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("hint h", func(t *testing.T) {
		var input controller.WillOTheWispWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("autocomplete ac", func(t *testing.T) {
		var input controller.WillOTheWispWebInput
		_ = json.Unmarshal([]byte(`{"command":"ac","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.WillOTheWispWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("undo u", func(t *testing.T) {
		var input controller.WillOTheWispWebInput
		_ = json.Unmarshal([]byte(`{"command":"u","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("move tableau to tableau", func(t *testing.T) {
		input := controller.WillOTheWispWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.WillOTheWispWebZone{Zone: "tableau", Col: intPtr(0), CardIndex: intPtr(2)},
			To:           &controller.WillOTheWispWebZone{Zone: "tableau", Col: intPtr(4)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.WillOTheWispWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustWillOTheWispOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		var input controller.WillOTheWispWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustWillOTheWispOutputJSON("param error."))
	})

	t.Run("session too long", func(t *testing.T) {
		input := controller.WillOTheWispWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustWillOTheWispOutputJSON("param error."))
	})
}

func TestWillOTheWispWebController_MoveErrors(t *testing.T) {
	siMock := new(usecase.MockWillOTheWispInteractor)
	factory := func() uc.WillOTheWispInteractorIF { return siMock }
	ctrl := controller.NewWillOTheWispWebController(factory)
	defer ctrl.Stop()

	t.Run("move without from/to", func(t *testing.T) {
		var input controller.WillOTheWispWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau to tableau missing params", func(t *testing.T) {
		input := controller.WillOTheWispWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.WillOTheWispWebZone{Zone: "tableau"},
			To:           &controller.WillOTheWispWebZone{Zone: "tableau"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zones", func(t *testing.T) {
		input := controller.WillOTheWispWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.WillOTheWispWebZone{Zone: "waste"},
			To:           &controller.WillOTheWispWebZone{Zone: "tableau"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestWillOTheWispWebController_UndoN(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"completedSuits":0,"score":0,"phase":0,"moveCount":0,"canUndo":false,"isStalemate":false,"message":""}`
	siMock := new(usecase.MockWillOTheWispInteractor)
	siMock.On("UndoN", 3).Return(mockOutput)

	factory := func() uc.WillOTheWispInteractorIF { return siMock }
	ctrl := controller.NewWillOTheWispWebController(factory)
	defer ctrl.Stop()

	t.Run("undo_n with valid n", func(t *testing.T) {
		n := 3
		input := controller.WillOTheWispWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1", N: &n},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("undo_n with missing n", func(t *testing.T) {
		input := controller.WillOTheWispWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestWillOTheWispWebController_Stop(t *testing.T) {
	siMock := new(usecase.MockWillOTheWispInteractor)
	factory := func() uc.WillOTheWispInteractorIF { return siMock }
	c := controller.NewWillOTheWispWebController(factory)
	c.Stop()
	c.Stop()
}
