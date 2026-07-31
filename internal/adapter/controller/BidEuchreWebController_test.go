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

func TestBidEuchreWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	siMock := new(usecase.MockBidEuchreInteractor)
	siMock.On("ResetWithConfig", domain.DefaultBidEuchreConfig()).Return(mockOutput)
	siMock.On("Bid", 4).Return(mockOutput)
	siMock.On("PassBid").Return(mockOutput)
	siMock.On("ChooseTrump", 5).Return(mockOutput)
	siMock.On("PlayCard", 4).Return(mockOutput)
	siMock.On("NextHand").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.BidEuchreInteractorIF { return siMock }
	ctrl := controller.NewBidEuchreWebController(factory)
	defer ctrl.Stop()

	t.Run("commands that need no parameter", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset", "ps", "pass", "n", "next", "log", "l"} {
			var input controller.BidEuchreWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("commands with a parameter", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"b","sessionId":"s1","value":4}`,
			`{"command":"bid","sessionId":"s1","value":4}`,
			`{"command":"t","sessionId":"s1","trump":5}`,
			`{"command":"trump","sessionId":"s1","trump":5}`,
			`{"command":"p","sessionId":"s1","cardIndex":4}`,
			`{"command":"play","sessionId":"s1","cardIndex":4}`,
		} {
			var input controller.BidEuchreWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("a missing parameter is a bad request", func(t *testing.T) {
		for _, body := range []string{
			`{"command":"b","sessionId":"s1"}`,
			`{"command":"t","sessionId":"s1"}`,
			`{"command":"p","sessionId":"s1"}`,
		} {
			var input controller.BidEuchreWebInput
			_ = json.Unmarshal([]byte(body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusBadRequest)
		}
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.BidEuchreWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestBidEuchreWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	run := func(t *testing.T, session string, cfg *controller.BidEuchreWebConfig, expected domain.BidEuchreConfig) {
		t.Helper()
		siMock := new(usecase.MockBidEuchreInteractor)
		siMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.BidEuchreInteractorIF { return siMock }
		ctrl := controller.NewBidEuchreWebController(factory)
		defer ctrl.Stop()

		input := controller.BidEuchreWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: session},
			Config:       cfg,
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "ResetWithConfig", expected)
	}

	t.Run("out-of-range values fall back to defaults", func(t *testing.T) {
		diff := 9
		run(t, "cfg-1", &controller.BidEuchreWebConfig{CpuDifficulty: &diff}, domain.DefaultBidEuchreConfig())
	})

	// **config はワイヤ上で任意。**省略時に落ちるとフロントの reset が死ぬ。
	t.Run("nil config uses defaults", func(t *testing.T) {
		run(t, "cfg-2", nil, domain.DefaultBidEuchreConfig())
	})

	// **ノートランプ禁止は設定で切れる。**
	t.Run("allowNoTrump comes through", func(t *testing.T) {
		off := false
		want := domain.DefaultBidEuchreConfig()
		want.AllowNoTrump = false
		run(t, "cfg-3", &controller.BidEuchreWebConfig{AllowNoTrump: &off}, want)
	})
}
