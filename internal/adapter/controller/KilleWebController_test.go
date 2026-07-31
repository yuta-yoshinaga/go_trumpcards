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

func TestKilleWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	siMock := new(usecase.MockKilleInteractor)
	siMock.On("ResetWithConfig", domain.DefaultKilleConfig()).Return(mockOutput)
	siMock.On("Exchange").Return(mockOutput)
	siMock.On("Satisfied").Return(mockOutput)
	siMock.On("Reenter").Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.KilleInteractorIF { return siMock }
	ctrl := controller.NewKilleWebController(factory)
	defer ctrl.Stop()

	t.Run("every command and alias", func(t *testing.T) {
		for _, cmd := range []string{
			"r", "reset", "e", "exchange", "s", "satisfied",
			"re", "reenter", "nr", "nextround", "log", "l",
		} {
			var input controller.KilleWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.KilleWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestKilleWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	run := func(t *testing.T, session string, cfg *controller.KilleWebConfig, expected domain.KilleConfig) {
		t.Helper()
		siMock := new(usecase.MockKilleInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.KilleInteractorIF { return siMock }
		ctrl := controller.NewKilleWebController(factory)
		defer ctrl.Stop()

		input := controller.KilleWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: session},
			Config:       cfg,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	}

	t.Run("custom stake is passed", func(t *testing.T) {
		stake := 20
		expected := domain.DefaultKilleConfig()
		expected.Stake = 20
		run(t, "cfg-1", &controller.KilleWebConfig{Stake: &stake}, expected)
	})

	t.Run("out-of-range values fall back to defaults", func(t *testing.T) {
		diff := 9
		stake := 0
		run(t, "cfg-2", &controller.KilleWebConfig{CpuDifficulty: &diff, Stake: &stake}, domain.DefaultKilleConfig())
	})

	// **config はワイヤ上で任意。**省略時に落ちるとフロントの reset が死ぬ。
	t.Run("nil config uses defaults", func(t *testing.T) {
		run(t, "cfg-3", nil, domain.DefaultKilleConfig())
	})
}
