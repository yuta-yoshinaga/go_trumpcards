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

func mustTrenteEtQuaranteOutputJSON(msg string) string {
	out := &controller.TrenteEtQuaranteWebOutput{
		NoirRow:       []*controller.WebOutputCard{},
		RougeRow:      []*controller.WebOutputCard{},
		WinningRow:    domain.TrenteEtQuaranteRowNone,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTrenteEtQuaranteOutputJSON: %v", err))
	}
	return string(b)
}

func TestTrenteEtQuaranteWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	biMock := new(usecase.MockTrenteEtQuaranteInteractor)
	biMock.On("ResetWithConfig", domain.DefaultTrenteEtQuaranteConfig()).Return(mockOutput)
	biMock.On("Bet", domain.TrenteEtQuaranteBetRouge, 100).Return(mockOutput)
	biMock.On("NextRound").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.TrenteEtQuaranteInteractorIF { return biMock }
	ctrl := controller.NewTrenteEtQuaranteWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.TrenteEtQuaranteWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustTrenteEtQuaranteOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("bet", func(t *testing.T) {
		betType := int(domain.TrenteEtQuaranteBetRouge)
		amount := 100
		input := controller.TrenteEtQuaranteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bet", SessionID: "s1"},
			Bet:          &betType,
			Stake:        &amount,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
		biMock.AssertCalled(t, "Bet", domain.TrenteEtQuaranteBetRouge, 100)
	})
	t.Run("bet missing betType", func(t *testing.T) {
		run(t, `{"command":"bet","sessionId":"s1"}`, mustTrenteEtQuaranteOutputJSON("param error: bet is required."), http.StatusBadRequest)
	})
	t.Run("bet missing amount", func(t *testing.T) {
		run(t, `{"command":"bet","sessionId":"s1","bet":1}`, mustTrenteEtQuaranteOutputJSON("param error: stake is required."), http.StatusBadRequest)
	})
	t.Run("nextround", func(t *testing.T) {
		run(t, `{"command":"nr","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("next alias", func(t *testing.T) {
		run(t, `{"command":"n","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint", func(t *testing.T) {
		run(t, `{"command":"h","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("log", func(t *testing.T) {
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("unsupported", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`, mustTrenteEtQuaranteOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
	t.Run("empty", func(t *testing.T) {
		run(t, `{"command":"","sessionId":"s1"}`, mustTrenteEtQuaranteOutputJSON("param error."), http.StatusBadRequest)
	})
}

func TestTrenteEtQuaranteWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"phase":0}`

	t.Run("custom config passed through", func(t *testing.T) {
		bet := 2
		expected := domain.TrenteEtQuaranteConfig{DefaultBet: domain.TrenteEtQuaranteBetCouleur}
		biMock := new(usecase.MockTrenteEtQuaranteInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTrenteEtQuaranteWebController(func() uc.TrenteEtQuaranteInteractorIF { return biMock })
		defer ctrl.Stop()

		input := controller.TrenteEtQuaranteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c1"},
			Config:       &controller.TrenteEtQuaranteWebConfig{DefaultBet: &bet},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("out of range bet falls back to default", func(t *testing.T) {
		bet := 9
		expected := domain.DefaultTrenteEtQuaranteConfig()
		biMock := new(usecase.MockTrenteEtQuaranteInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTrenteEtQuaranteWebController(func() uc.TrenteEtQuaranteInteractorIF { return biMock })
		defer ctrl.Stop()

		input := controller.TrenteEtQuaranteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c2"},
			Config:       &controller.TrenteEtQuaranteWebConfig{DefaultBet: &bet},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		expected := domain.DefaultTrenteEtQuaranteConfig()
		biMock := new(usecase.MockTrenteEtQuaranteInteractor)
		biMock.On("ResetWithConfig", expected).Return(mockOutput)
		ctrl := controller.NewTrenteEtQuaranteWebController(func() uc.TrenteEtQuaranteInteractorIF { return biMock })
		defer ctrl.Stop()

		input := controller.TrenteEtQuaranteWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "c3"},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "ResetWithConfig", expected)
	})
}

func TestTrenteEtQuaranteWebController_Stop(t *testing.T) {
	biMock := new(usecase.MockTrenteEtQuaranteInteractor)
	c := controller.NewTrenteEtQuaranteWebController(func() uc.TrenteEtQuaranteInteractorIF { return biMock })
	c.Stop()
	c.Stop()
}
