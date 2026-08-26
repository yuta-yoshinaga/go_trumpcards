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

func mustWhiteheadOutputJSON(msg string) string {
	out := &controller.WhiteheadWebOutput{
		Tableau:       [][]*controller.WhiteheadWebOutputTableauCard{},
		Waste:         []*controller.WebOutputCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustWhiteheadOutputJSON: %v", err))
	}
	return string(b)
}

func intPtrWH(v int) *int { return &v }

func TestWhiteheadWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"waste":[],"foundation":[],"phase":0,"moveCount":0,"message":""}`
	expectedBody := mockOutput

	kiMock := new(usecase.MockWhiteheadInteractor)
	kiMock.On("Reset").Return(mockOutput)
	kiMock.On("Draw").Return(mockOutput)
	kiMock.On("GiveUp").Return(mockOutput)
	kiMock.On("Hint").Return(mockOutput)
	kiMock.On("AutoComplete").Return(mockOutput)
	kiMock.On("ActionLog").Return(mockOutput)
	kiMock.On("MoveWasteToTableau", 3).Return(mockOutput)
	kiMock.On("MoveWasteToFoundation").Return(mockOutput)
	kiMock.On("MoveTableauToTableau", 0, 2, 4).Return(mockOutput)
	kiMock.On("MoveTableauToFoundation", 1).Return(mockOutput)

	factory := func() uc.WhiteheadInteractorIF { return kiMock }
	ctrl := controller.NewWhiteheadWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustWhiteheadOutputJSON("bye."))
	})

	t.Run("quit", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustWhiteheadOutputJSON("bye."))
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("draw d", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("draw", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"draw","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup g", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"g","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"giveup","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hint h", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("autocomplete ac", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"ac","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("autocomplete", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"autocomplete","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("l shorthand", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	// Move commands
	t.Run("move waste to tableau", func(t *testing.T) {
		input := controller.WhiteheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.WhiteheadWebZone{Zone: "waste"},
			To:           &controller.WhiteheadWebZone{Zone: "tableau", Col: intPtrWH(3)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move waste to foundation", func(t *testing.T) {
		input := controller.WhiteheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
			From:         &controller.WhiteheadWebZone{Zone: "waste"},
			To:           &controller.WhiteheadWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move tableau to tableau", func(t *testing.T) {
		input := controller.WhiteheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.WhiteheadWebZone{Zone: "tableau", Col: intPtrWH(0), CardIndex: intPtrWH(2)},
			To:           &controller.WhiteheadWebZone{Zone: "tableau", Col: intPtrWH(4)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move tableau to foundation", func(t *testing.T) {
		input := controller.WhiteheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.WhiteheadWebZone{Zone: "tableau", Col: intPtrWH(1)},
			To:           &controller.WhiteheadWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	// Error cases
	t.Run("unsupported command", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustWhiteheadOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustWhiteheadOutputJSON("param error."))
	})

	t.Run("empty session", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustWhiteheadOutputJSON("param error."))
	})

	t.Run("session too long", func(t *testing.T) {
		input := controller.WhiteheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustWhiteheadOutputJSON("param error."))
	})
}

func TestWhiteheadWebController_MoveErrors(t *testing.T) {
	kiMock := new(usecase.MockWhiteheadInteractor)
	factory := func() uc.WhiteheadInteractorIF { return kiMock }
	ctrl := controller.NewWhiteheadWebController(factory)
	defer ctrl.Stop()

	t.Run("move without from/to", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move waste to tableau no col", func(t *testing.T) {
		input := controller.WhiteheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.WhiteheadWebZone{Zone: "waste"},
			To:           &controller.WhiteheadWebZone{Zone: "tableau"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau to tableau missing params", func(t *testing.T) {
		input := controller.WhiteheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.WhiteheadWebZone{Zone: "tableau"},
			To:           &controller.WhiteheadWebZone{Zone: "tableau"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau to foundation no col", func(t *testing.T) {
		input := controller.WhiteheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.WhiteheadWebZone{Zone: "tableau"},
			To:           &controller.WhiteheadWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zones", func(t *testing.T) {
		input := controller.WhiteheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.WhiteheadWebZone{Zone: "foundation"},
			To:           &controller.WhiteheadWebZone{Zone: "waste"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestWhiteheadWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"waste":[],"foundation":[],"phase":0,"moveCount":0,"drawCount":3,"message":""}`

	kiMock := new(usecase.MockWhiteheadInteractor)
	kiMock.On("ResetWithConfig", domain.WhiteheadConfig{DrawCount: 3}).Return(mockOutput)

	factory := func() uc.WhiteheadInteractorIF { return kiMock }
	ctrl := controller.NewWhiteheadWebController(factory)
	defer ctrl.Stop()

	input := controller.WhiteheadWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
		Config:       &controller.WhiteheadWebConfig{DrawCount: intPtrWH(3)},
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	recorded.BodyIs(mockOutput)
}

func TestWhiteheadWebController_ResetWithScoringMode(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"waste":[],"foundation":[],"phase":0,"moveCount":0,"message":""}`

	kiMock := new(usecase.MockWhiteheadInteractor)
	kiMock.On("ResetWithConfig", domain.WhiteheadConfig{ScoringMode: domain.WhiteheadScoringVegas}).Return(mockOutput)

	factory := func() uc.WhiteheadInteractorIF { return kiMock }
	ctrl := controller.NewWhiteheadWebController(factory)
	defer ctrl.Stop()

	scoringMode := 1
	input := controller.WhiteheadWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
		Config:       &controller.WhiteheadWebConfig{ScoringMode: &scoringMode},
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	recorded.BodyIs(mockOutput)
}

func TestWhiteheadWebController_Undo(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"waste":[],"foundation":[],"phase":0,"moveCount":0,"message":""}`

	kiMock := new(usecase.MockWhiteheadInteractor)
	kiMock.On("Undo").Return(mockOutput)

	factory := func() uc.WhiteheadInteractorIF { return kiMock }
	ctrl := controller.NewWhiteheadWebController(factory)
	defer ctrl.Stop()

	t.Run("undo u", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"u","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("undo", func(t *testing.T) {
		var input controller.WhiteheadWebInput
		_ = json.Unmarshal([]byte(`{"command":"undo","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
}

func TestWhiteheadWebController_UndoN(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"waste":[],"foundation":[],"phase":0,"moveCount":0,"message":""}`

	kiMock := new(usecase.MockWhiteheadInteractor)
	kiMock.On("UndoN", 3).Return(mockOutput)

	factory := func() uc.WhiteheadInteractorIF { return kiMock }
	ctrl := controller.NewWhiteheadWebController(factory)
	defer ctrl.Stop()

	t.Run("undo_n with valid n", func(t *testing.T) {
		n := 3
		input := controller.WhiteheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1", N: &n},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("undo_n with missing n", func(t *testing.T) {
		input := controller.WhiteheadWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "undo_n", SessionID: "s1"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestWhiteheadWebController_Stop(t *testing.T) {
	kiMock := new(usecase.MockWhiteheadInteractor)
	factory := func() uc.WhiteheadInteractorIF { return kiMock }
	c := controller.NewWhiteheadWebController(factory)
	c.Stop()
	c.Stop()
}
