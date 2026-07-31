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

func TestBostonWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	siMock := new(usecase.MockBostonInteractor)
	siMock.On("ResetWithConfig", domain.DefaultBostonConfig()).Return(mockOutput)
	siMock.On("Bid", domain.BostonBidSeven, domain.CardDesignHeart).Return(mockOutput)
	siMock.On("Bid", domain.BostonBidLittleMisere, 0).Return(mockOutput)
	siMock.On("PassBid").Return(mockOutput)
	siMock.On("CallPartner", 2).Return(mockOutput)
	siMock.On("CallPartner", -1).Return(mockOutput)
	siMock.On("PlayCard", 3).Return(mockOutput)
	siMock.On("NextHand").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.BostonInteractorIF { return siMock }
	ctrl := controller.NewBostonWebController(factory)
	defer ctrl.Stop()

	t.Run("commands that need no parameter", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset", "ps", "pass", "n", "next", "log", "l"} {
			var input controller.BostonWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("commands with a parameter", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"b","sessionId":"s1","level":4,"suit":3}`,
			`{"command":"bid","sessionId":"s1","level":3}`,
			`{"command":"cp","sessionId":"s1","partner":2}`,
			// **-1 は単独で戦う有効な選択。**弾いてはいけない。
			`{"command":"callpartner","sessionId":"s1","partner":-1}`,
			`{"command":"p","sessionId":"s1","cardIndex":3}`,
			`{"command":"play","sessionId":"s1","cardIndex":3}`,
		} {
			var input controller.BostonWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("a missing parameter is a bad request", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"b","sessionId":"s1"}`,
			`{"command":"cp","sessionId":"s1"}`,
			`{"command":"p","sessionId":"s1"}`,
		} {
			var input controller.BostonWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusBadRequest)
		}
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.BostonWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestBostonWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	run := func(t *testing.T, session string, cfg *controller.BostonWebConfig, expected domain.BostonConfig) {
		t.Helper()
		siMock := new(usecase.MockBostonInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.BostonInteractorIF { return siMock }
		ctrl := controller.NewBostonWebController(factory)
		defer ctrl.Stop()

		input := controller.BostonWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: session},
			Config:       cfg,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	}

	t.Run("custom values are passed", func(t *testing.T) {
		hands := 5
		expected := domain.DefaultBostonConfig()
		expected.TargetHands = 5
		run(t, "cfg-1", &controller.BostonWebConfig{TargetHands: &hands}, expected)
	})

	t.Run("out-of-range values fall back to defaults", func(t *testing.T) {
		hands := 999
		diff := 9
		run(t, "cfg-2", &controller.BostonWebConfig{TargetHands: &hands, CpuDifficulty: &diff}, domain.DefaultBostonConfig())
	})

	// **config はワイヤ上で任意。**省略時に落ちるとフロントの reset が死ぬ。
	t.Run("nil config uses defaults", func(t *testing.T) {
		run(t, "cfg-3", nil, domain.DefaultBostonConfig())
	})
}
