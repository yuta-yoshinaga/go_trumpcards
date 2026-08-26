//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustQuodlibetOutputJSON(msg string) string {
	out := &controller.QuodlibetWebOutput{
		Players:                []*controller.QuodlibetWebOutputPlayer{},
		AvailableContracts:     []int{},
		AvailableContractNames: []string{},
		CurrentTrick:           []*controller.WebOutputTrickCard{},
		LastTrick:              []*controller.WebOutputTrickCard{},
		PlayableIndices:        []int{},
		TablePlaced:            [][]int{},
		Stack:                  []*controller.WebOutputCard{},
		DealHistory:            []*controller.QuodlibetWebOutputDeal{},
		Winners:                []int{},
		CurrentContract:        -1,
		LastTrickWinner:        -1,
		HintContract:           -1,
		TotalDeals:             domain.QuodlibetTotalDeals,
		TrickCount:             domain.QuodlibetHandSize,
		WebOutputBase:          controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustQuodlibetOutputJSON: %v", err))
	}
	return string(b)
}

func TestQuodlibetWebController_Method(t *testing.T) {
	mockOutput := `{"phase":"play"}`

	diMock := new(usecase.MockQuodlibetInteractor)
	diMock.On("ResetWithConfig", domain.DefaultQuodlibetConfig()).Return(mockOutput)
	diMock.On("SelectContract", 2).Return(mockOutput)
	diMock.On("SelectContract", 0).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("Play", -1).Return(mockOutput)
	diMock.On("NextDeal").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewQuodlibetWebController(func() uc.QuodlibetInteractorIF { return diMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.QuodlibetWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustQuodlibetOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("contract", func(t *testing.T) {
		run(t, `{"command":"contract","contract":2,"sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "SelectContract", 2)
	})
	// **0 は正当な種目番号 (プラス)。** 未指定と同じ扱いにすると選べない。
	t.Run("contract 0 is a real choice", func(t *testing.T) {
		run(t, `{"command":"contract","contract":0,"sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "SelectContract", 0)
	})
	t.Run("contract without a number is refused", func(t *testing.T) {
		var input controller.QuodlibetWebInput
		_ = json.Unmarshal([]byte(`{"command":"contract","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		run(t, `{"command":"play","cardIndex":3,"sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Play", 3)
	})
	// **札の指定が無い要求は断る。** 0 として通すと、意図しない札が出る。
	t.Run("play without a card index is refused", func(t *testing.T) {
		var input controller.QuodlibetWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
	// **パスは -1 のプレイ。** 別命令にせず同じ経路に載せることで、
	// 「パスできないのにパスした」の判定がドメイン 1 箇所で済む。
	t.Run("pass goes through play as -1", func(t *testing.T) {
		run(t, `{"command":"pass","sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Play", -1)
	})
	t.Run("nextdeal", func(t *testing.T) {
		run(t, `{"command":"nextdeal","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint and log", func(t *testing.T) {
		run(t, `{"command":"hint","sessionId":"s1"}`, mockOutput, http.StatusOK)
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
}

// 設定は範囲で丸め、組み立てた設定はドメインの検査を通る。
func TestQuodlibetWebConfig_ToConfig(t *testing.T) {
	i := func(n int) *int { return &n }
	b := func(v bool) *bool { return &v }

	empty := controller.QuodlibetWebConfig{}
	assert.Equal(t, domain.DefaultQuodlibetConfig(), empty.ToConfig())

	wc := controller.QuodlibetWebConfig{CpuDifficulty: i(9)}
	cfg := wc.ToConfig()
	assert.NoError(t, cfg.Validate(), "範囲外の難易度が丸められていない")

	wc = controller.QuodlibetWebConfig{CpuDifficulty: i(0), AutoSelectContract: b(true)}
	cfg = wc.ToConfig()
	assert.Equal(t, domain.QuodlibetCpuDifficultyEasy, cfg.CpuDifficulty)
	assert.True(t, cfg.AutoSelectContract)
	assert.NoError(t, cfg.Validate())
}
