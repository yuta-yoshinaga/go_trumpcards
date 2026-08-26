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

func mustUnsunKarutaOutputJSON(msg string) string {
	out := &controller.UnsunKarutaWebOutput{
		Players:         []*controller.UnsunKarutaWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		PlayableIndices: []int{},
		TeamTricks:      []int{},
		TeamScores:      []int{},
		LastTrickWinner: -1,
		WinnerTeam:      -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustUnsunKarutaOutputJSON: %v", err))
	}
	return string(b)
}

func TestUnsunKarutaWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	diMock := new(usecase.MockUnsunKarutaInteractor)
	diMock.On("ResetWithConfig", domain.DefaultUnsunKarutaConfig()).Return(mockOutput)
	diMock.On("Play", 3, false).Return(mockOutput)
	diMock.On("Play", 2, true).Return(mockOutput)
	diMock.On("NextTrick").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewUnsunKarutaWebController(func() uc.UnsunKarutaInteractorIF { return diMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.UnsunKarutaWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustUnsunKarutaOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	// **宣言は札と一緒に届く。** 省略時は宣言なし。
	t.Run("play without declaring", func(t *testing.T) {
		run(t, `{"command":"play","cardIndex":3,"sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Play", 3, false)
	})
	t.Run("play with a declaration", func(t *testing.T) {
		run(t, `{"command":"play","cardIndex":2,"declare":true,"sessionId":"s1"}`, mockOutput, http.StatusOK)
		diMock.AssertCalled(t, "Play", 2, true)
	})
	// **札の指定が無い要求は断る。** 0 として通すと、意図しない札が出る。
	t.Run("play without a card index is refused", func(t *testing.T) {
		var input controller.UnsunKarutaWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
	t.Run("next and nextround", func(t *testing.T) {
		run(t, `{"command":"next","sessionId":"s1"}`, mockOutput, http.StatusOK)
		run(t, `{"command":"nextround","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint and log", func(t *testing.T) {
		run(t, `{"command":"hint","sessionId":"s1"}`, mockOutput, http.StatusOK)
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
}

// 設定は範囲で丸め、組み立てた設定はドメインの検査を通る。
func TestUnsunKarutaWebConfig_ToConfig(t *testing.T) {
	i := func(n int) *int { return &n }

	for _, tt := range []struct {
		name      string
		in        controller.UnsunKarutaWebConfig
		wantDeals int
	}{
		{"defaults", controller.UnsunKarutaWebConfig{}, domain.UnsunKarutaDefaultDeals},
		{"in range", controller.UnsunKarutaWebConfig{TargetDeals: i(8)}, 8},
		{"too many falls back", controller.UnsunKarutaWebConfig{TargetDeals: i(99)}, domain.UnsunKarutaDefaultDeals},
		{"zero falls back", controller.UnsunKarutaWebConfig{TargetDeals: i(0)}, domain.UnsunKarutaDefaultDeals},
	} {
		t.Run(tt.name, func(t *testing.T) {
			in := tt.in
			cfg := in.ToConfig()
			assert.Equal(t, tt.wantDeals, cfg.TargetDeals)
			assert.NoError(t, cfg.Validate())
		})
	}

	wc := controller.UnsunKarutaWebConfig{CpuDifficulty: i(9)}
	cfg := wc.ToConfig()
	assert.NoError(t, cfg.Validate(), "範囲外の難易度が丸められていない")
}
