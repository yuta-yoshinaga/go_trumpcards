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

func newKlaberjassCtrlMock(mockOutput string) *usecase.MockKlaberjassInteractor {
	m := new(usecase.MockKlaberjassInteractor)
	m.On("ResetWithConfig", domain.DefaultKlaberjassConfig()).Return(mockOutput)
	m.On("AcceptTrump").Return(mockOutput)
	m.On("CallTrump", 3).Return(mockOutput)
	m.On("Pass").Return(mockOutput)
	m.On("Schmeiss").Return(mockOutput)
	m.On("AnswerSchmeiss", true).Return(mockOutput)
	m.On("PlayCard", 1).Return(mockOutput)
	m.On("NextDeal").Return(mockOutput)
	m.On("ActionLog").Return(mockOutput)
	return m
}

func TestKlaberjassWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`
	siMock := newKlaberjassCtrlMock(mockOutput)

	factory := func() uc.KlaberjassInteractorIF { return siMock }
	ctrl := controller.NewKlaberjassWebController(factory)
	defer ctrl.Stop()

	t.Run("commands that need no parameter", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset", "a", "accept", "ps", "pass", "sm", "schmeiss", "n", "next", "log", "l"} {
			var input controller.KlaberjassWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("commands with a parameter", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"c","sessionId":"s1","suit":3}`,
			`{"command":"call","sessionId":"s1","suit":3}`,
			`{"command":"p","sessionId":"s1","cardIndex":1}`,
			`{"command":"play","sessionId":"s1","cardIndex":1}`,
			`{"command":"as","sessionId":"s1","accept":true}`,
			`{"command":"answerschmeiss","sessionId":"s1","accept":true}`,
		} {
			var input controller.KlaberjassWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	// **省略された必須パラメータは 400 で返す。**nil 参照で落とさない。
	t.Run("a missing parameter is a bad request", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"c","sessionId":"s1"}`,
			`{"command":"p","sessionId":"s1"}`,
			`{"command":"as","sessionId":"s1"}`,
		} {
			var input controller.KlaberjassWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusBadRequest)
		}
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.KlaberjassWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestKlaberjassWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	run := func(t *testing.T, session string, cfg *controller.KlaberjassWebConfig, expected domain.KlaberjassConfig) {
		t.Helper()
		siMock := new(usecase.MockKlaberjassInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.KlaberjassInteractorIF { return siMock }
		ctrl := controller.NewKlaberjassWebController(factory)
		defer ctrl.Stop()

		input := controller.KlaberjassWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: session},
			Config:       cfg,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	}

	t.Run("custom values are passed", func(t *testing.T) {
		target := 300
		schmeiss := false
		expected := domain.DefaultKlaberjassConfig()
		expected.TargetScore = 300
		expected.AllowSchmeiss = false
		run(t, "cfg-1", &controller.KlaberjassWebConfig{TargetScore: &target, AllowSchmeiss: &schmeiss}, expected)
	})

	t.Run("out-of-range values fall back to defaults", func(t *testing.T) {
		target := 5
		diff := 9
		run(t, "cfg-2", &controller.KlaberjassWebConfig{TargetScore: &target, CpuDifficulty: &diff}, domain.DefaultKlaberjassConfig())
	})

	// **config はワイヤ上で任意。**省略時に落ちるとフロントの reset が死ぬ。
	t.Run("nil config uses defaults", func(t *testing.T) {
		run(t, "cfg-3", nil, domain.DefaultKlaberjassConfig())
	})
}
