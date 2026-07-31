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

func TestKaiserWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	siMock := new(usecase.MockKaiserInteractor)
	siMock.On("ResetWithConfig", domain.DefaultKaiserConfig()).Return(mockOutput)
	siMock.On("Bid", 8, domain.KaiserContractTrump).Return(mockOutput)
	siMock.On("Bid", 8, domain.KaiserContractLowNoTrump).Return(mockOutput)
	siMock.On("PassBid").Return(mockOutput)
	siMock.On("SetTrump", 3).Return(mockOutput)
	siMock.On("Discard", []int{0, 1}).Return(mockOutput)
	siMock.On("PlayCard", 2).Return(mockOutput)
	siMock.On("NextHand").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.KaiserInteractorIF { return siMock }
	ctrl := controller.NewKaiserWebController(factory)
	defer ctrl.Stop()

	t.Run("commands that need no parameter", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset", "ps", "pass", "n", "next", "log", "l"} {
			var input controller.KaiserWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("commands with a parameter", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"b","sessionId":"s1","bid":8}`,
			`{"command":"bid","sessionId":"s1","bid":8,"contract":2}`,
			`{"command":"t","sessionId":"s1","suit":3}`,
			`{"command":"trump","sessionId":"s1","suit":3}`,
			`{"command":"d","sessionId":"s1","indices":[0,1]}`,
			`{"command":"discard","sessionId":"s1","indices":[0,1]}`,
			`{"command":"p","sessionId":"s1","cardIndex":2}`,
			`{"command":"play","sessionId":"s1","cardIndex":2}`,
		} {
			var input controller.KaiserWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	// **省略された必須パラメータは 400 で返す。**nil 参照で落とさない。
	t.Run("a missing parameter is a bad request", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"b","sessionId":"s1"}`,
			`{"command":"t","sessionId":"s1"}`,
			`{"command":"d","sessionId":"s1"}`,
			`{"command":"p","sessionId":"s1"}`,
		} {
			var input controller.KaiserWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusBadRequest)
		}
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.KaiserWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestKaiserWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	run := func(t *testing.T, session string, cfg *controller.KaiserWebConfig, expected domain.KaiserConfig) {
		t.Helper()
		siMock := new(usecase.MockKaiserInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.KaiserInteractorIF { return siMock }
		ctrl := controller.NewKaiserWebController(factory)
		defer ctrl.Stop()

		input := controller.KaiserWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: session},
			Config:       cfg,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	}

	t.Run("custom values are passed", func(t *testing.T) {
		allow := false
		expected := domain.DefaultKaiserConfig()
		expected.AllowNoTrump = false
		run(t, "cfg-1", &controller.KaiserWebConfig{AllowNoTrump: &allow}, expected)
	})

	t.Run("out-of-range values fall back to defaults", func(t *testing.T) {
		diff := 9
		run(t, "cfg-2", &controller.KaiserWebConfig{CpuDifficulty: &diff}, domain.DefaultKaiserConfig())
	})

	// **config はワイヤ上で任意。**省略時に落ちるとフロントの reset が死ぬ。
	t.Run("nil config uses defaults", func(t *testing.T) {
		run(t, "cfg-3", nil, domain.DefaultKaiserConfig())
	})
}
