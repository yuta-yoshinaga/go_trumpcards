package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func TestCassinoWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`
	ciMock := new(usecase.MockCassinoInteractor)
	ciMock.On("Reset").Return(mockOutput)
	ciMock.On("NextRound").Return(mockOutput)
	ciMock.On("Take", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
	ciMock.On("Build", mock.Anything, mock.Anything, mock.Anything).Return(mockOutput)
	ciMock.On("Trail", mock.Anything).Return(mockOutput)
	ciMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)

	factory := func() uc.CassinoInteractorIF { return ciMock }
	ctrl := controller.NewCassinoWebController(factory)
	defer ctrl.Stop()

	var jsonInput controller.CassinoWebInput

	t.Run("reset r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
		rec.BodyIs(mockOutput)
	})

	t.Run("reset with config", func(t *testing.T) {
		input := controller.CassinoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
			Config:       &controller.CassinoWebConfig{TargetScore: 21, MultiBuildEnabled: true, SweepBonusEnabled: true, CpuDifficulty: 2},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		ciMock.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("take", func(t *testing.T) {
		input := controller.CassinoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "t", SessionID: "s1"},
			HandIndex:    0,
			TableIndices: []int{0, 1},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		ciMock.AssertCalled(t, "Take", 0, []int{0, 1}, []int(nil))
	})

	t.Run("build", func(t *testing.T) {
		input := controller.CassinoWebInput{
			BaseWebInput:  controller.BaseWebInput{Command: "b", SessionID: "s1"},
			HandIndex:     0,
			TableIndices:  []int{1},
			DeclaredValue: 8,
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		ciMock.AssertCalled(t, "Build", 0, []int{1}, 8)
	})

	t.Run("trail", func(t *testing.T) {
		input := controller.CassinoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "tr", SessionID: "s1"},
			HandIndex:    1,
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		ciMock.AssertCalled(t, "Trail", 1)
	})

	t.Run("next", func(t *testing.T) {
		input := controller.CassinoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "n", SessionID: "s1"},
		}
		rec := execRequest(t, ctrl.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		ciMock.AssertCalled(t, "NextRound")
	})

	t.Run("unsupported command returns 400", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &jsonInput)
		rec := execRequest(t, ctrl.Exec, &jsonInput)
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestCassinoWebConfig_ToConfig(t *testing.T) {
	wc := controller.CassinoWebConfig{
		TargetScore:       15,
		MultiBuildEnabled: false,
		SweepBonusEnabled: true,
		CpuDifficulty:     2,
	}
	c := wc.ToConfig()
	if c.TargetScore != 15 || c.MultiBuildEnabled || !c.SweepBonusEnabled || c.CpuDifficulty != 2 {
		t.Fatalf("ToConfig mapping incorrect: %+v", c)
	}
}
