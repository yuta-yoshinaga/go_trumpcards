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

func mustTappTarockOutputJSON(msg string) string {
	out := &controller.TappTarockWebOutput{
		Players:            []*controller.TappTarockWebOutputPlayer{},
		CurrentTrick:       []*controller.WebOutputTrickCard{},
		LastTrickCards:     []*controller.WebOutputCard{},
		PlayableIndices:    []int{},
		DiscardableIndices: []int{},
		DeclarerIdx:        -1,
		LastTrickWinner:    -1,
		WinnerPlayer:       -1,
		WebOutputBase:      controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTappTarockOutputJSON: %v", err))
	}
	return string(b)
}

func TestTappTarockWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	ziMock := new(usecase.MockTappTarockInteractor)
	ziMock.On("ResetWithConfig", domain.DefaultTappTarockConfig()).Return(mockOutput)
	ziMock.On("Bid", domain.TappTarockBidDreier).Return(mockOutput)
	ziMock.On("Bid", domain.TappTarockBidSolo).Return(mockOutput)
	ziMock.On("Pass").Return(mockOutput)
	ziMock.On("Discard", []int{0, 1, 2, 3, 4, 5}).Return(mockOutput)
	ziMock.On("Play", 3).Return(mockOutput)
	ziMock.On("NextTrick").Return(mockOutput)
	ziMock.On("NextRound").Return(mockOutput)
	ziMock.On("Hint").Return(mockOutput)
	ziMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.TappTarockInteractorIF { return ziMock }
	ctrl := controller.NewTappTarockWebController(factory)
	defer ctrl.Stop()

	run := func(t *testing.T, body, want string, code int) {
		t.Helper()
		var input controller.TappTarockWebInput
		_ = json.Unmarshal([]byte(body), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(code)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(want)
	}

	t.Run("q", func(t *testing.T) {
		run(t, `{"command":"q","sessionId":"s1"}`, mustTappTarockOutputJSON("bye."), http.StatusOK)
	})
	t.Run("reset", func(t *testing.T) {
		run(t, `{"command":"reset","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("bid dreier", func(t *testing.T) {
		run(t, `{"command":"b","bid":"dreier","sessionId":"s1"}`, mockOutput, http.StatusOK)
		ziMock.AssertCalled(t, "Bid", domain.TappTarockBidDreier)
	})
	t.Run("bid solo", func(t *testing.T) {
		run(t, `{"command":"b","bid":"solo","sessionId":"s1"}`, mockOutput, http.StatusOK)
		ziMock.AssertCalled(t, "Bid", domain.TappTarockBidSolo)
	})
	t.Run("bid missing bid", func(t *testing.T) {
		run(t, `{"command":"b","sessionId":"s1"}`,
			mustTappTarockOutputJSON("param error: bid is required."), http.StatusBadRequest)
	})
	t.Run("pass", func(t *testing.T) {
		run(t, `{"command":"pass","sessionId":"s1"}`, mockOutput, http.StatusOK)
	})
	t.Run("discard", func(t *testing.T) {
		input := controller.TappTarockWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "s1"},
			CardIndices:  []int{0, 1, 2, 3, 4, 5},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("discard missing indices", func(t *testing.T) {
		run(t, `{"command":"d","sessionId":"s1"}`,
			mustTappTarockOutputJSON("param error: cardIndices is required."), http.StatusBadRequest)
	})
	t.Run("play", func(t *testing.T) {
		input := controller.TappTarockWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    func() *int { v := 3; return &v }(),
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
	t.Run("play missing cardIndex", func(t *testing.T) {
		run(t, `{"command":"p","sessionId":"s1"}`,
			mustTappTarockOutputJSON("param error: cardIndex is required."), http.StatusBadRequest)
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
			mustTappTarockOutputJSON("Unsupported command."), http.StatusBadRequest)
	})
}

// **trischaken は Web API でも宣言できない。** 全員パスの結果としてしか成立しない
// 契約なので、受け付けると「誰も落札しなかった」という前提が崩れる。
func TestTappTarockWebController_RejectsTrischakenBid(t *testing.T) {
	ziMock := new(usecase.MockTappTarockInteractor)
	ziMock.On("Bid", mock.Anything).Return(`{"phase":0}`)

	ctrl := controller.NewTappTarockWebController(func() uc.TappTarockInteractorIF { return ziMock })
	defer ctrl.Stop()

	for _, bid := range []string{"trischaken", "pass", "nonsense"} {
		t.Run(bid, func(t *testing.T) {
			var input controller.TappTarockWebInput
			require.NoError(t, json.Unmarshal(
				[]byte(`{"command":"b","bid":"`+bid+`","sessionId":"s1"}`), &input))
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusBadRequest)
			// **送った文字列に触れる返事を返す。** ドメインの汎用エラーだと
			// 何が悪かったのか分からない。
			recorded.BodyIs(mustTappTarockOutputJSON("param error: bid must be dreier or solo."))
		})
	}
	ziMock.AssertNotCalled(t, "Bid", mock.Anything)
}

// **範囲外の設定は既定に落ちる。** 丸めではなく既定への差し戻しが webutil の
// 契約で、どちらにせよ不正な設定はドメインへ届かない。
func TestTappTarockWebConfig_ToConfigNeverYieldsAnInvalidConfig(t *testing.T) {
	tests := []struct {
		name string
		body string
		want domain.TappTarockConfig
	}{
		{"未指定なら既定", `{}`, domain.DefaultTappTarockConfig()},
		{"ディールが多すぎる", `{"targetDeals":99}`, domain.DefaultTappTarockConfig()},
		{"ディールが 0", `{"targetDeals":0}`, domain.DefaultTappTarockConfig()},
		{"難易度が範囲外", `{"cpuDifficulty":9}`, domain.DefaultTappTarockConfig()},
		{"範囲内はそのまま通る", `{"cpuDifficulty":2,"targetDeals":12}`, domain.TappTarockConfig{
			CpuDifficulty: domain.TappTarockCpuDifficultyHard,
			TargetDeals:   domain.TappTarockMaxDeals,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg controller.TappTarockWebConfig
			require.NoError(t, json.Unmarshal([]byte(tt.body), &cfg))
			got := cfg.ToConfig()
			assert.Equal(t, tt.want, got)
			// **結果はドメインの検証を通る。** 範囲チェックの結果を検証まで通さないと、
			// 「範囲内に見えるが不正」な設定が素通りする。
			assert.NoError(t, got.Validate())
		})
	}
}
