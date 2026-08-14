//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustTrogguOutputJSON(msg string) string {
	out := &controller.TrogguWebOutput{
		Players:         []*controller.TrogguWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		LastTrickCards:  []*controller.WebOutputCard{},
		PlayableIndices: []int{},
		DeclarerIdx:     -1,
		LastTrickWinner: -1,
		WinnerPlayer:    -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTrogguOutputJSON: %v", err))
	}
	return string(b)
}

func TestTrogguWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	tiMock := new(usecase.MockTrogguInteractor)
	tiMock.On("ResetWithConfig", domain.DefaultTrogguConfig()).Return(mockOutput)
	tiMock.On("Bid", mock.Anything).Return(mockOutput)
	tiMock.On("Pass").Return(mockOutput)
	tiMock.On("Play", 3).Return(mockOutput)
	tiMock.On("NextTrick").Return(mockOutput)
	tiMock.On("NextRound").Return(mockOutput)
	tiMock.On("Hint").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewTrogguWebController(func() uc.TrogguInteractorIF { return tiMock })
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.TrogguWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustTrogguOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})

	// **4 つの契約がすべて Web API から打てる。**
	for _, tt := range []struct {
		arg  string
		want domain.TrogguBid
	}{
		{"trois", domain.TrogguBidTrois},
		{"solo", domain.TrogguBidSolo},
		{"piccolo", domain.TrogguBidPiccolo},
		{"misere", domain.TrogguBidMisere},
	} {
		t.Run("bid "+tt.arg, func(t *testing.T) {
			run(t, `{"command":"b","bid":"`+tt.arg+`","sessionId":"s1"}`, mockOutput, http.StatusOK)
			tiMock.AssertCalled(t, "Bid", tt.want)
		})
	}

	t.Run("bid missing bid", func(t *testing.T) {
		run(t, `{"command":"b","sessionId":"s1"}`,
			mustTrogguOutputJSON("param error: bid is required."), http.StatusBadRequest)
	})
	t.Run("pass", func(t *testing.T) {
		run(t, `{"command":"pass","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.TrogguWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`,
			mustTrogguOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
	})
	t.Run("next", func(t *testing.T) {
		run(t, `{"command":"n","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("nextround", func(t *testing.T) {
		run(t, `{"command":"nr","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("hint and log", func(t *testing.T) {
		run(t, `{"command":"h","sessionId":"s1"}`, mockOutput, http.StatusOK)
		run(t, `{"command":"log","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("unsupported command", func(t *testing.T) {
		run(t, `{"command":"other","sessionId":"s1"}`,
			mustTrogguOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
}

// **知らない入札は名指しで断る。** パスに化けさせてドメインへ運ぶと、送った文字列を
// 含まない汎用エラーになる。
func TestTrogguWebController_RejectsUnknownBid(t *testing.T) {
	tiMock := new(usecase.MockTrogguInteractor)
	tiMock.On("Bid", mock.Anything).Return(`{"phase":0}`)

	ctrl := controller.NewTrogguWebController(func() uc.TrogguInteractorIF { return tiMock })
	defer ctrl.Stop()

	for _, bid := range []string{"pass", "nonsense", ""} {
		t.Run("bid="+bid, func(t *testing.T) {
			var input controller.TrogguWebInput
			require.NoError(t, json.Unmarshal(
				[]byte(`{"command":"b","bid":"`+bid+`","sessionId":"s1"}`), &input))
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusBadRequest)
			recorded.BodyIs(mustTrogguOutputJSON("param error: bid must be trois, solo, piccolo or misere."))
		})
	}
	tiMock.AssertNotCalled(t, "Bid", mock.Anything)
}

// **範囲外の設定は既定に落ちる。** どちらにせよ不正な設定はドメインへ届かない。
func TestTrogguWebConfig_ToConfigNeverYieldsAnInvalidConfig(t *testing.T) {
	tests := []struct {
		name string
		body string
		want domain.TrogguConfig
	}{
		{"未指定なら既定", `{}`, domain.DefaultTrogguConfig()},
		{"ディールが多すぎる", `{"targetDeals":99}`, domain.DefaultTrogguConfig()},
		{"ディールが 0", `{"targetDeals":0}`, domain.DefaultTrogguConfig()},
		{"難易度が範囲外", `{"cpuDifficulty":9}`, domain.DefaultTrogguConfig()},
		{"範囲内はそのまま通る", `{"cpuDifficulty":2,"targetDeals":12}`, domain.TrogguConfig{
			CpuDifficulty: domain.TrogguCpuDifficultyHard,
			TargetDeals:   domain.TrogguMaxDeals,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg controller.TrogguWebConfig
			require.NoError(t, json.Unmarshal([]byte(tt.body), &cfg))
			got := cfg.ToConfig()
			assert.Equal(t, tt.want, got)
			// **丸めた結果を検証まで通す。** 範囲チェックだけでは、範囲内に見えて
			// 不正な組み合わせが素通りする。
			assert.NoError(t, got.Validate())
		})
	}
}
