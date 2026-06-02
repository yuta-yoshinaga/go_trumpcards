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

func mustOsmosisOutputJSON(msg string) string {
	out := &controller.OsmosisWebOutput{
		Reserve:       [][]*controller.WebOutputCard{},
		Waste:         []*controller.WebOutputCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustOsmosisOutputJSON: %v", err))
	}
	return string(b)
}

func osmosisIntPtr(v int) *int { return &v }

func TestOsmosisWebController_Method(t *testing.T) {
	mockOutput := `{"reserve":[],"stockCount":0,"waste":[],"foundation":[],"baseRank":0,"phase":0,"moveCount":0,"canUndo":false,"message":""}`
	expectedBody := mockOutput

	oiMock := new(usecase.MockOsmosisInteractor)
	oiMock.On("Reset").Return(mockOutput)
	oiMock.On("Draw").Return(mockOutput)
	oiMock.On("GiveUp").Return(mockOutput)
	oiMock.On("Hint").Return(mockOutput)
	oiMock.On("AutoComplete").Return(mockOutput)
	oiMock.On("ActionLog").Return(mockOutput)
	oiMock.On("MoveWasteToFoundation", 1).Return(mockOutput)
	oiMock.On("MoveReserveToFoundation", 2, 3).Return(mockOutput)

	factory := func() uc.OsmosisInteractorIF { return oiMock }
	ctrl := controller.NewOsmosisWebController(factory)
	defer ctrl.Stop()

	t.Run("quit", func(t *testing.T) {
		var input controller.OsmosisWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustOsmosisOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.OsmosisWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("draw", func(t *testing.T) {
		var input controller.OsmosisWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup", func(t *testing.T) {
		var input controller.OsmosisWebInput
		_ = json.Unmarshal([]byte(`{"command":"giveup","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.OsmosisWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("autocomplete", func(t *testing.T) {
		var input controller.OsmosisWebInput
		_ = json.Unmarshal([]byte(`{"command":"ac","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.OsmosisWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move waste to foundation", func(t *testing.T) {
		input := controller.OsmosisWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
			From:         &controller.OsmosisWebZone{Zone: "waste"},
			To:           &controller.OsmosisWebZone{Zone: "foundation", Col: osmosisIntPtr(1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move reserve to foundation", func(t *testing.T) {
		input := controller.OsmosisWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.OsmosisWebZone{Zone: "reserve", Col: osmosisIntPtr(2)},
			To:           &controller.OsmosisWebZone{Zone: "foundation", Col: osmosisIntPtr(3)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.OsmosisWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustOsmosisOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		var input controller.OsmosisWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustOsmosisOutputJSON("param error."))
	})

	t.Run("session too long", func(t *testing.T) {
		input := controller.OsmosisWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustOsmosisOutputJSON("param error."))
	})
}

func TestOsmosisWebController_MoveErrors(t *testing.T) {
	oiMock := new(usecase.MockOsmosisInteractor)
	factory := func() uc.OsmosisInteractorIF { return oiMock }
	ctrl := controller.NewOsmosisWebController(factory)
	defer ctrl.Stop()

	t.Run("move without from/to", func(t *testing.T) {
		var input controller.OsmosisWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move waste foundation no col", func(t *testing.T) {
		input := controller.OsmosisWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.OsmosisWebZone{Zone: "waste"},
			To:           &controller.OsmosisWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move reserve foundation missing cols", func(t *testing.T) {
		input := controller.OsmosisWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.OsmosisWebZone{Zone: "reserve"},
			To:           &controller.OsmosisWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zones", func(t *testing.T) {
		input := controller.OsmosisWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.OsmosisWebZone{Zone: "foundation"},
			To:           &controller.OsmosisWebZone{Zone: "waste"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestOsmosisWebController_Undo(t *testing.T) {
	mockOutput := `{"reserve":[],"stockCount":0,"waste":[],"foundation":[],"baseRank":0,"phase":0,"moveCount":0,"canUndo":false,"message":""}`

	oiMock := new(usecase.MockOsmosisInteractor)
	oiMock.On("Undo").Return(mockOutput)

	factory := func() uc.OsmosisInteractorIF { return oiMock }
	ctrl := controller.NewOsmosisWebController(factory)
	defer ctrl.Stop()

	var input controller.OsmosisWebInput
	_ = json.Unmarshal([]byte(`{"command":"u","sessionId":"s1"}`), &input)
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	recorded.BodyIs(mockOutput)
}

func TestOsmosisWebController_UndoN(t *testing.T) {
	mockOutput := `{"reserve":[],"stockCount":0,"waste":[],"foundation":[],"baseRank":0,"phase":0,"moveCount":0,"canUndo":false,"message":""}`

	oiMock := new(usecase.MockOsmosisInteractor)
	oiMock.On("UndoN", 3).Return(mockOutput)

	factory := func() uc.OsmosisInteractorIF { return oiMock }
	ctrl := controller.NewOsmosisWebController(factory)
	defer ctrl.Stop()

	t.Run("undo_n valid", func(t *testing.T) {
		n := 3
		input := controller.OsmosisWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1", N: &n},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("undo_n missing n", func(t *testing.T) {
		input := controller.OsmosisWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestOsmosisWebController_Stop(t *testing.T) {
	oiMock := new(usecase.MockOsmosisInteractor)
	factory := func() uc.OsmosisInteractorIF { return oiMock }
	c := controller.NewOsmosisWebController(factory)
	c.Stop()
	c.Stop()
}
