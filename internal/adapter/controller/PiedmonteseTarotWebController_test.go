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

func mustPiedmonteseTarotOutputJSON(msg string) string {
	out := &controller.PiedmonteseTarotWebOutput{
		Players:         []*controller.PiedmonteseTarotWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		PlayableIndices: []int{},
		PlayerScores:    []int{},
		DealScores:      []int{},
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustPiedmonteseTarotOutputJSON: %v", err))
	}
	return string(b)
}

func TestPiedmonteseTarotWebController_Method(t *testing.T) {
	mockOutput := `{"phase":1}`

	diMock := new(usecase.MockPiedmonteseTarotInteractor)
	diMock.On("ResetWithConfig", domain.DefaultPiedmonteseTarotConfig()).Return(mockOutput)
	diMock.On("Discard", []int{0, 1}).Return(mockOutput)
	diMock.On("Play", 3).Return(mockOutput)
	diMock.On("NextTrick").Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewPiedmonteseTarotWebController(
		func() uc.PiedmonteseTarotInteractorIF { return diMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.PiedmonteseTarotWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustPiedmonteseTarotOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("scarto", func(t *testing.T) {
		run(t, `{"command":"scarto","cardIndices":[0,1],"sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("play", func(t *testing.T) {
		run(t, `{"command":"play","cardIndex":3,"sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("next", func(t *testing.T) {
		run(t, `{"command":"next","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("nextround", func(t *testing.T) {
		run(t, `{"command":"nextround","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint", func(t *testing.T) {
		run(t, `{"command":"hint","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("log", func(t *testing.T) {
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
}

// **席数は丸めない。** 配り方があるのは 3 人と 4 人だけで、5 を 4 に丸めると
// 要求した卓とは違う卓が「成功」として返る。
func TestPiedmonteseTarotWebConfig_ToConfig(t *testing.T) {
	seats := func(n int) *int { return &n }

	for _, tt := range []struct {
		name string
		in   controller.PiedmonteseTarotWebConfig
		want int
	}{
		{"three seats", controller.PiedmonteseTarotWebConfig{Seats: seats(3)}, 3},
		{"four seats", controller.PiedmonteseTarotWebConfig{Seats: seats(4)}, 4},
		{"five falls back to the default", controller.PiedmonteseTarotWebConfig{Seats: seats(5)}, 4},
		{"zero falls back to the default", controller.PiedmonteseTarotWebConfig{Seats: seats(0)}, 4},
		{"absent keeps the default", controller.PiedmonteseTarotWebConfig{}, 4},
	} {
		t.Run(tt.name, func(t *testing.T) {
			in := tt.in
			cfg := in.ToConfig()
			assert.Equal(t, tt.want, cfg.Seats)
			assert.NoError(t, cfg.Validate(), "組み立てた設定がドメインの検査を通らない")
		})
	}

	// 難易度とディール数は範囲で丸める (既存の設定と同じ扱い)。
	d, deals := 9, 0
	wc := controller.PiedmonteseTarotWebConfig{CpuDifficulty: &d, TargetDeals: &deals}
	cfg := wc.ToConfig()
	assert.NoError(t, cfg.Validate())
}
