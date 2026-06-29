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

func mustAgnesOutputJSON(msg string) string {
	out := &controller.AgnesWebOutput{
		Tableau:       [][]*controller.AgnesWebOutputTableauCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustAgnesOutputJSON: %v", err))
	}
	return string(b)
}

func agnesIntPtr(v int) *int { return &v }

func TestAgnesWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"foundation":[],"baseRank":0,"phase":0,"moveCount":0,"canUndo":false,"message":""}`
	expectedBody := mockOutput

	ciMock := new(usecase.MockAgnesInteractor)
	ciMock.On("Reset").Return(mockOutput)
	ciMock.On("DealStock").Return(mockOutput)
	ciMock.On("GiveUp").Return(mockOutput)
	ciMock.On("Hint").Return(mockOutput)
	ciMock.On("ActionLog").Return(mockOutput)
	ciMock.On("MoveTableauToTableau", 0, 2, 3).Return(mockOutput)
	ciMock.On("MoveTableauToFoundation", 1).Return(mockOutput)

	factory := func() uc.AgnesInteractorIF { return ciMock }
	ctrl := controller.NewAgnesWebController(factory)
	defer ctrl.Stop()

	t.Run("reset", func(t *testing.T) {
		var input controller.AgnesWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("deal", func(t *testing.T) {
		var input controller.AgnesWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup", func(t *testing.T) {
		var input controller.AgnesWebInput
		_ = json.Unmarshal([]byte(`{"command":"giveup","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.AgnesWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.AgnesWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move tableau to tableau", func(t *testing.T) {
		input := controller.AgnesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.AgnesWebZone{Zone: "tableau", Col: agnesIntPtr(0), CardIndex: agnesIntPtr(2)},
			To:           &controller.AgnesWebZone{Zone: "tableau", Col: agnesIntPtr(3)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move tableau to foundation", func(t *testing.T) {
		input := controller.AgnesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
			From:         &controller.AgnesWebZone{Zone: "tableau", Col: agnesIntPtr(1)},
			To:           &controller.AgnesWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.AgnesWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustAgnesOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		var input controller.AgnesWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustAgnesOutputJSON("param error."))
	})
}

func TestAgnesWebController_MoveErrors(t *testing.T) {
	ciMock := new(usecase.MockAgnesInteractor)
	factory := func() uc.AgnesInteractorIF { return ciMock }
	ctrl := controller.NewAgnesWebController(factory)
	defer ctrl.Stop()

	t.Run("move without from/to", func(t *testing.T) {
		var input controller.AgnesWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau tableau missing params", func(t *testing.T) {
		input := controller.AgnesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.AgnesWebZone{Zone: "tableau"},
			To:           &controller.AgnesWebZone{Zone: "tableau"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau foundation no col", func(t *testing.T) {
		input := controller.AgnesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.AgnesWebZone{Zone: "tableau"},
			To:           &controller.AgnesWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zones", func(t *testing.T) {
		input := controller.AgnesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.AgnesWebZone{Zone: "foundation"},
			To:           &controller.AgnesWebZone{Zone: "tableau", Col: agnesIntPtr(0)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestAgnesWebController_Undo(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"foundation":[],"baseRank":0,"phase":0,"moveCount":0,"canUndo":false,"message":""}`

	ciMock := new(usecase.MockAgnesInteractor)
	ciMock.On("Undo").Return(mockOutput)

	factory := func() uc.AgnesInteractorIF { return ciMock }
	ctrl := controller.NewAgnesWebController(factory)
	defer ctrl.Stop()

	var input controller.AgnesWebInput
	_ = json.Unmarshal([]byte(`{"command":"u","sessionId":"s1"}`), &input)
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	recorded.BodyIs(mockOutput)
}

func TestAgnesWebController_UndoN(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"foundation":[],"baseRank":0,"phase":0,"moveCount":0,"canUndo":false,"message":""}`

	ciMock := new(usecase.MockAgnesInteractor)
	ciMock.On("UndoN", 3).Return(mockOutput)

	factory := func() uc.AgnesInteractorIF { return ciMock }
	ctrl := controller.NewAgnesWebController(factory)
	defer ctrl.Stop()

	t.Run("undo_n valid", func(t *testing.T) {
		n := 3
		input := controller.AgnesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1", N: &n},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("undo_n missing n", func(t *testing.T) {
		input := controller.AgnesWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestAgnesWebController_Stop(t *testing.T) {
	ciMock := new(usecase.MockAgnesInteractor)
	factory := func() uc.AgnesInteractorIF { return ciMock }
	c := controller.NewAgnesWebController(factory)
	c.Stop()
	c.Stop()
}
