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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustMrsMopOutputJSON(msg string) string {
	out := &controller.MrsMopWebOutput{
		Tableau:       [][]*controller.MrsMopWebOutputTableauCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMrsMopOutputJSON: %v", err))
	}
	return string(b)
}

func TestMrsMopWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"completedSuits":0,"phase":0,"moveCount":0,"canUndo":false,"isStalemate":false,"score":0,"difficulty":0,"message":""}`
	expectedBody := mockOutput

	siMock := new(usecase.MockMrsMopInteractor)
	siMock.On("Reset").Return(mockOutput)
	siMock.On("Deal").Return(mockOutput)
	siMock.On("GiveUp").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("AutoComplete").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)
	siMock.On("MoveTableauToTableau", 0, 2, 4).Return(mockOutput)
	siMock.On("Undo").Return(mockOutput)

	factory := func() uc.MrsMopInteractorIF { return siMock }
	ctrl := controller.NewMrsMopWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustMrsMopOutputJSON("bye."))
	})

	t.Run("quit", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustMrsMopOutputJSON("bye."))
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	// **配るコマンドは存在しない。**山札が無いのに 200 を返すと、盤が変わらない
	// 無言の no-op になる。
	t.Run("deal is rejected", func(t *testing.T) {
		for _, cmd := range []string{"d", "deal"} {
			var input controller.MrsMopWebInput
			_ = json.Unmarshal([]byte(`{"command":"`+cmd+`","sessionId":"s1"}`), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusBadRequest)
		}
	})

	t.Run("giveup g", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"g","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"giveup","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hint h", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("autocomplete ac", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"ac","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("autocomplete", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"autocomplete","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("l shorthand", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo u", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"u","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("undo", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"undo","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	// Move command
	t.Run("move tableau to tableau", func(t *testing.T) {
		input := controller.MrsMopWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.MrsMopWebZone{Zone: "tableau", Col: intPtr(0), CardIndex: intPtr(2)},
			To:           &controller.MrsMopWebZone{Zone: "tableau", Col: intPtr(4)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	// Error cases
	t.Run("unsupported command", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustMrsMopOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustMrsMopOutputJSON("param error."))
	})

	t.Run("empty session", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustMrsMopOutputJSON("param error."))
	})

	t.Run("session too long", func(t *testing.T) {
		input := controller.MrsMopWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustMrsMopOutputJSON("param error."))
	})
}

func TestMrsMopWebController_MoveErrors(t *testing.T) {
	siMock := new(usecase.MockMrsMopInteractor)
	factory := func() uc.MrsMopInteractorIF { return siMock }
	ctrl := controller.NewMrsMopWebController(factory)
	defer ctrl.Stop()

	t.Run("move without from/to", func(t *testing.T) {
		var input controller.MrsMopWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau to tableau missing params", func(t *testing.T) {
		input := controller.MrsMopWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.MrsMopWebZone{Zone: "tableau"},
			To:           &controller.MrsMopWebZone{Zone: "tableau"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zones", func(t *testing.T) {
		input := controller.MrsMopWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.MrsMopWebZone{Zone: "waste"},
			To:           &controller.MrsMopWebZone{Zone: "tableau"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestMrsMopWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"completedSuits":0,"phase":0,"moveCount":0,"canUndo":false,"isStalemate":false,"score":0,"difficulty":2,"message":""}`

	siMock := new(usecase.MockMrsMopInteractor)
	siMock.On("ResetWithConfig", domain.MrsMopConfig{Difficulty: domain.MrsMopDifficulty2Suit}).Return(mockOutput)

	factory := func() uc.MrsMopInteractorIF { return siMock }
	ctrl := controller.NewMrsMopWebController(factory)
	defer ctrl.Stop()

	difficulty := 2
	input := controller.MrsMopWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
		Config:       &controller.MrsMopWebConfig{Difficulty: &difficulty},
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	recorded.BodyIs(mockOutput)
}

func TestMrsMopWebController_UndoN(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"completedSuits":0,"phase":0,"moveCount":0,"canUndo":false,"isStalemate":false,"score":0,"difficulty":0,"message":""}`

	siMock := new(usecase.MockMrsMopInteractor)
	siMock.On("UndoN", 3).Return(mockOutput)

	factory := func() uc.MrsMopInteractorIF { return siMock }
	ctrl := controller.NewMrsMopWebController(factory)
	defer ctrl.Stop()

	t.Run("undo_n with valid n", func(t *testing.T) {
		n := 3
		input := controller.MrsMopWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1", N: &n},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("undo_n with missing n", func(t *testing.T) {
		input := controller.MrsMopWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestMrsMopWebController_Stop(t *testing.T) {
	siMock := new(usecase.MockMrsMopInteractor)
	factory := func() uc.MrsMopInteractorIF { return siMock }
	c := controller.NewMrsMopWebController(factory)
	c.Stop()
	c.Stop()
}
