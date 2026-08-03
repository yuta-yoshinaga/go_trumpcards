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

func TestSixBidSoloWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	siMock := new(usecase.MockSixBidSoloInteractor)
	siMock.On("ResetWithConfig", domain.DefaultSixBidSoloConfig()).Return(mockOutput)
	siMock.On("Bid", 3).Return(mockOutput)
	siMock.On("PassBid").Return(mockOutput)
	siMock.On("Declare", 1, 0, 0).Return(mockOutput)
	siMock.On("Declare", 1, 3, 1).Return(mockOutput)
	siMock.On("PlayCard", 4).Return(mockOutput)
	siMock.On("NextHand").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.SixBidSoloInteractorIF { return siMock }
	ctrl := controller.NewSixBidSoloWebController(factory)
	defer ctrl.Stop()

	t.Run("commands that need no parameter", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset", "ps", "pass", "n", "next", "log", "l"} {
			var input controller.SixBidSoloWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("commands with a parameter", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"b","sessionId":"s1","bid":3}`,
			`{"command":"bid","sessionId":"s1","bid":3}`,
			`{"command":"d","sessionId":"s1","suit":1}`,
			`{"command":"declare","sessionId":"s1","suit":1,"calledSuit":3,"calledValue":1}`,
			`{"command":"p","sessionId":"s1","cardIndex":4}`,
			`{"command":"play","sessionId":"s1","cardIndex":4}`,
		} {
			var input controller.SixBidSoloWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	// **指名札はスートとランクの両方が要る。**片方だけでは札が決まらない。
	t.Run("a missing parameter is a bad request", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"b","sessionId":"s1"}`,
			`{"command":"d","sessionId":"s1"}`,
			`{"command":"d","sessionId":"s1","suit":1,"calledSuit":3}`,
			`{"command":"d","sessionId":"s1","suit":1,"calledValue":1}`,
			`{"command":"p","sessionId":"s1"}`,
		} {
			var input controller.SixBidSoloWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusBadRequest)
		}
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.SixBidSoloWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestSixBidSoloWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	run := func(t *testing.T, session string, cfg *controller.SixBidSoloWebConfig, expected domain.SixBidSoloConfig) {
		t.Helper()
		siMock := new(usecase.MockSixBidSoloInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.SixBidSoloInteractorIF { return siMock }
		ctrl := controller.NewSixBidSoloWebController(factory)
		defer ctrl.Stop()

		input := controller.SixBidSoloWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: session},
			Config:       cfg,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	}

	t.Run("out-of-range values fall back to defaults", func(t *testing.T) {
		diff, hands := 9, 99
		run(t, "cfg-1", &controller.SixBidSoloWebConfig{CpuDifficulty: &diff, TargetHands: &hands},
			domain.DefaultSixBidSoloConfig())
	})

	// **config はワイヤ上で任意。**省略時に落ちるとフロントの reset が死ぬ。
	t.Run("nil config uses defaults", func(t *testing.T) {
		run(t, "cfg-2", nil, domain.DefaultSixBidSoloConfig())
	})

	t.Run("a valid hand count comes through", func(t *testing.T) {
		hands := 9
		want := domain.DefaultSixBidSoloConfig()
		want.TargetHands = hands
		run(t, "cfg-3", &controller.SixBidSoloWebConfig{TargetHands: &hands}, want)
	})
}
