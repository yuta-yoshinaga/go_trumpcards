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

func mustKlondikeOutputJSON(msg string) string {
	out := &controller.KlondikeWebOutput{
		Tableau:       [][]*controller.KlondikeWebOutputTableauCard{},
		Waste:         []*controller.WebOutputCard{},
		Foundation:    [][]*controller.WebOutputCard{},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustKlondikeOutputJSON: %v", err))
	}
	return string(b)
}

func intPtr(v int) *int { return &v }

func TestKlondikeWebController_Method(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"waste":[],"foundation":[],"phase":0,"moveCount":0,"message":""}`
	expectedBody := mockOutput

	kiMock := new(usecase.MockKlondikeInteractor)
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

	factory := func() uc.KlondikeInteractorIF { return kiMock }
	ctrl := controller.NewKlondikeWebController(factory)
	defer ctrl.Stop()

	t.Run("quit q", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustKlondikeOutputJSON("bye."))
	})

	t.Run("quit", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"quit","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mustKlondikeOutputJSON("bye."))
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("draw d", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("draw", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"draw","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup g", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"g","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("giveup", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"giveup","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hint h", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("autocomplete ac", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"ac","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("autocomplete", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"autocomplete","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("l shorthand", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	// Move commands
	t.Run("move waste to tableau", func(t *testing.T) {
		input := controller.KlondikeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.KlondikeWebZone{Zone: "waste"},
			To:           &controller.KlondikeWebZone{Zone: "tableau", Col: intPtr(3)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move waste to foundation", func(t *testing.T) {
		input := controller.KlondikeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "move", SessionID: "s1"},
			From:         &controller.KlondikeWebZone{Zone: "waste"},
			To:           &controller.KlondikeWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move tableau to tableau", func(t *testing.T) {
		input := controller.KlondikeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.KlondikeWebZone{Zone: "tableau", Col: intPtr(0), CardIndex: intPtr(2)},
			To:           &controller.KlondikeWebZone{Zone: "tableau", Col: intPtr(4)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	t.Run("move tableau to foundation", func(t *testing.T) {
		input := controller.KlondikeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.KlondikeWebZone{Zone: "tableau", Col: intPtr(1)},
			To:           &controller.KlondikeWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(expectedBody)
	})

	// Error cases
	t.Run("unsupported command", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustKlondikeOutputJSON("Unsupported command."))
	})

	t.Run("empty command", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustKlondikeOutputJSON("param error."))
	})

	t.Run("empty session", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":""}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustKlondikeOutputJSON("param error."))
	})

	t.Run("session too long", func(t *testing.T) {
		input := controller.KlondikeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustKlondikeOutputJSON("param error."))
	})
}

func TestKlondikeWebController_MoveErrors(t *testing.T) {
	kiMock := new(usecase.MockKlondikeInteractor)
	factory := func() uc.KlondikeInteractorIF { return kiMock }
	ctrl := controller.NewKlondikeWebController(factory)
	defer ctrl.Stop()

	t.Run("move without from/to", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"m","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move waste to tableau no col", func(t *testing.T) {
		input := controller.KlondikeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.KlondikeWebZone{Zone: "waste"},
			To:           &controller.KlondikeWebZone{Zone: "tableau"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau to tableau missing params", func(t *testing.T) {
		input := controller.KlondikeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.KlondikeWebZone{Zone: "tableau"},
			To:           &controller.KlondikeWebZone{Zone: "tableau"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move tableau to foundation no col", func(t *testing.T) {
		input := controller.KlondikeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.KlondikeWebZone{Zone: "tableau"},
			To:           &controller.KlondikeWebZone{Zone: "foundation"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("move invalid zones", func(t *testing.T) {
		input := controller.KlondikeWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "m", SessionID: "s1"},
			From:         &controller.KlondikeWebZone{Zone: "foundation"},
			To:           &controller.KlondikeWebZone{Zone: "waste"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestKlondikeWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"waste":[],"foundation":[],"phase":0,"moveCount":0,"drawCount":3,"message":""}`

	kiMock := new(usecase.MockKlondikeInteractor)
	kiMock.On("ResetWithConfig", domain.KlondikeConfig{DrawCount: 3}).Return(mockOutput)

	factory := func() uc.KlondikeInteractorIF { return kiMock }
	ctrl := controller.NewKlondikeWebController(factory)
	defer ctrl.Stop()

	input := controller.KlondikeWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
		Config:       &controller.KlondikeWebConfig{DrawCount: intPtr(3)},
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	recorded.BodyIs(mockOutput)
}

func TestKlondikeWebController_ResetWithScoringMode(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"waste":[],"foundation":[],"phase":0,"moveCount":0,"message":""}`

	kiMock := new(usecase.MockKlondikeInteractor)
	kiMock.On("ResetWithConfig", domain.KlondikeConfig{ScoringMode: domain.KlondikeScoringVegas}).Return(mockOutput)

	factory := func() uc.KlondikeInteractorIF { return kiMock }
	ctrl := controller.NewKlondikeWebController(factory)
	defer ctrl.Stop()

	scoringMode := 1
	input := controller.KlondikeWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
		Config:       &controller.KlondikeWebConfig{ScoringMode: &scoringMode},
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	recorded.BodyIs(mockOutput)
}

func TestKlondikeWebController_Undo(t *testing.T) {
	mockOutput := `{"tableau":[],"stockCount":0,"waste":[],"foundation":[],"phase":0,"moveCount":0,"message":""}`

	kiMock := new(usecase.MockKlondikeInteractor)
	kiMock.On("Undo").Return(mockOutput)

	factory := func() uc.KlondikeInteractorIF { return kiMock }
	ctrl := controller.NewKlondikeWebController(factory)
	defer ctrl.Stop()

	t.Run("undo u", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"u","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("undo", func(t *testing.T) {
		var input controller.KlondikeWebInput
		_ = json.Unmarshal([]byte(`{"command":"undo","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
}

func TestKlondikeWebController_Stop(t *testing.T) {
	kiMock := new(usecase.MockKlondikeInteractor)
	factory := func() uc.KlondikeInteractorIF { return kiMock }
	c := controller.NewKlondikeWebController(factory)
	c.Stop()
	c.Stop()
}
