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

func mustCanfieldOutputJSON(msg string) string {
	out := &controller.CanfieldWebOutput{
		Tableau:       [][]*controller.CanfieldWebOutputTableauCard{},
		Reserve:       []*controller.WebOutputCard{},
		Waste:         []*controller.WebOutputCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustCanfieldOutputJSON: %v", err))
	}
	return string(b)
}

func canfieldIntPtr(v int) *int { return &v }

func TestCanfieldWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"reserve":[],"stockCount":0,"waste":[],"foundation":[],"baseRank":0,"phase":0,"moveCount":0,"canUndo":false,"message":""}`
	expectedBody := mockOutput

	ciMock := new(usecase.MockCanfieldInteractor)
	ciMock.On("Reset").Return(mockOutput)
	ciMock.On("Draw").Return(mockOutput)
	ciMock.On("GiveUp").Return(mockOutput)
	ciMock.On("Hint").Return(mockOutput)
	ciMock.On("AutoComplete").Return(mockOutput)
	ciMock.On("ActionLog").Return(mockOutput)
	ciMock.On("MoveWasteToTableau", 3).Return(mockOutput)
	ciMock.On("MoveWasteToFoundation").Return(mockOutput)
	ciMock.On("MoveReserveToTableau", 2).Return(mockOutput)
	ciMock.On("MoveReserveToFoundation").Return(mockOutput)
	ciMock.On("MoveTableauToTableau", 0, 2, 3).Return(mockOutput)
	ciMock.On("MoveTableauToFoundation", 1).Return(mockOutput)

	factory := func() uc.CanfieldInteractorIF { return ciMock }
	ctrl := controller.NewCanfieldWebController(factory)
	defer ctrl.Stop()

	t.Run("quit", func(t *testing.T) {
		var input controller.CanfieldWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustCanfieldOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.CanfieldWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("draw", func(t *testing.T) {
		var input controller.CanfieldWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup", func(t *testing.T) {
		var input controller.CanfieldWebInput
		_ = json.Unmarshal([]byte(`{"command":"giveup","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.CanfieldWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("autocomplete", func(t *testing.T) {
		var input controller.CanfieldWebInput
		_ = json.Unmarshal([]byte(`{"command":"ac","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.CanfieldWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move waste to tableau", func(t *testing.T) {
		input := controller.CanfieldWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.CanfieldWebZone{Zone: "waste"},
			To:           &controller.CanfieldWebZone{Zone: "tableau", Col: canfieldIntPtr(3)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move waste to foundation", func(t *testing.T) {
		input := controller.CanfieldWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
			From:         &controller.CanfieldWebZone{Zone: "waste"},
			To:           &controller.CanfieldWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move reserve to tableau", func(t *testing.T) {
		input := controller.CanfieldWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.CanfieldWebZone{Zone: "reserve"},
			To:           &controller.CanfieldWebZone{Zone: "tableau", Col: canfieldIntPtr(2)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move reserve to foundation", func(t *testing.T) {
		input := controller.CanfieldWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.CanfieldWebZone{Zone: "reserve"},
			To:           &controller.CanfieldWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move tableau to tableau", func(t *testing.T) {
		input := controller.CanfieldWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.CanfieldWebZone{Zone: "tableau", Col: canfieldIntPtr(0), CardIndex: canfieldIntPtr(2)},
			To:           &controller.CanfieldWebZone{Zone: "tableau", Col: canfieldIntPtr(3)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move tableau to foundation", func(t *testing.T) {
		input := controller.CanfieldWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.CanfieldWebZone{Zone: "tableau", Col: canfieldIntPtr(1)},
			To:           &controller.CanfieldWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.CanfieldWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustCanfieldOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		var input controller.CanfieldWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustCanfieldOutputJSON("param error."))
	})

	t.Run("empty session", func(t *testing.T) {
		var input controller.CanfieldWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustCanfieldOutputJSON("param error."))
	})

	t.Run("session too long", func(t *testing.T) {
		input := controller.CanfieldWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustCanfieldOutputJSON("param error."))
	})
}

func TestCanfieldWebController_MoveErrors(t *testing.T) {
	ciMock := new(usecase.MockCanfieldInteractor)
	factory := func() uc.CanfieldInteractorIF { return ciMock }
	ctrl := controller.NewCanfieldWebController(factory)
	defer ctrl.Stop()

	t.Run("move without from/to", func(t *testing.T) {
		var input controller.CanfieldWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move waste tableau no col", func(t *testing.T) {
		input := controller.CanfieldWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.CanfieldWebZone{Zone: "waste"},
			To:           &controller.CanfieldWebZone{Zone: "tableau"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move reserve tableau no col", func(t *testing.T) {
		input := controller.CanfieldWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.CanfieldWebZone{Zone: "reserve"},
			To:           &controller.CanfieldWebZone{Zone: "tableau"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau tableau missing params", func(t *testing.T) {
		input := controller.CanfieldWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.CanfieldWebZone{Zone: "tableau"},
			To:           &controller.CanfieldWebZone{Zone: "tableau"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau foundation no col", func(t *testing.T) {
		input := controller.CanfieldWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.CanfieldWebZone{Zone: "tableau"},
			To:           &controller.CanfieldWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zones", func(t *testing.T) {
		input := controller.CanfieldWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.CanfieldWebZone{Zone: "foundation"},
			To:           &controller.CanfieldWebZone{Zone: "waste"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestCanfieldWebController_Undo(t *testing.T) {
	mockOutput := `{"tableau":[],"reserve":[],"stockCount":0,"waste":[],"foundation":[],"baseRank":0,"phase":0,"moveCount":0,"canUndo":false,"message":""}`

	ciMock := new(usecase.MockCanfieldInteractor)
	ciMock.On("Undo").Return(mockOutput)

	factory := func() uc.CanfieldInteractorIF { return ciMock }
	ctrl := controller.NewCanfieldWebController(factory)
	defer ctrl.Stop()

	var input controller.CanfieldWebInput
	_ = json.Unmarshal([]byte(`{"command":"u","sessionId":"s1"}`), &input)
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	recorded.BodyIs(mockOutput)
}

func TestCanfieldWebController_UndoN(t *testing.T) {
	mockOutput := `{"tableau":[],"reserve":[],"stockCount":0,"waste":[],"foundation":[],"baseRank":0,"phase":0,"moveCount":0,"canUndo":false,"message":""}`

	ciMock := new(usecase.MockCanfieldInteractor)
	ciMock.On("UndoN", 3).Return(mockOutput)

	factory := func() uc.CanfieldInteractorIF { return ciMock }
	ctrl := controller.NewCanfieldWebController(factory)
	defer ctrl.Stop()

	t.Run("undo_n valid", func(t *testing.T) {
		n := 3
		input := controller.CanfieldWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1", N: &n},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("undo_n missing n", func(t *testing.T) {
		input := controller.CanfieldWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestCanfieldWebController_Stop(t *testing.T) {
	ciMock := new(usecase.MockCanfieldInteractor)
	factory := func() uc.CanfieldInteractorIF { return ciMock }
	c := controller.NewCanfieldWebController(factory)
	c.Stop()
	c.Stop()
}
