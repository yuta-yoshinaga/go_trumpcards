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

func intPtrEst(v int) *int { return &v }

func mustEstimationOutputJSON(msg string) string {
	out := &controller.EstimationWebOutput{
		Players:       []*controller.EstimationWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		RestrictedBid: -1,
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustEstimationOutputJSON: %v", err))
	}
	return string(b)
}

func TestEstimationWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	eiMock := new(usecase.MockEstimationInteractor)
	eiMock.On("ResetWithConfig", domain.DefaultEstimationConfig()).Return(mockOutput)
	eiMock.On("ResetWithConfig", domain.EstimationConfig{Rounds: 9}).Return(mockOutput)
	eiMock.On("SelectTrump", 3).Return(mockOutput)
	eiMock.On("Bid", 5).Return(mockOutput)
	eiMock.On("Bid", 0).Return(mockOutput)
	eiMock.On("NextRound").Return(mockOutput)
	eiMock.On("GiveUp").Return(mockOutput)
	eiMock.On("Hint").Return(mockOutput)
	eiMock.On("ActionLog").Return(mockOutput)
	eiMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewEstimationWebController(func() uc.EstimationInteractorIF { return eiMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.EstimationWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustEstimationOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with rounds", `{"command":"reset","sessionId":"s1","config":{"rounds":9}}`},
		{"trump t", `{"command":"t","sessionId":"s1","suit":3}`},
		{"trump long", `{"command":"trump","sessionId":"s1","suit":3}`},
		{"bid b", `{"command":"b","sessionId":"s1","bid":5}`},
		// **0 は Dash Call という別の宣言。** 省略と区別できること。
		{"dash call", `{"command":"bid","sessionId":"s1","bid":0}`},
		{"play p", `{"command":"p","sessionId":"s1","cardIndex":4}`},
		{"next n", `{"command":"n","sessionId":"s1"}`},
		{"giveup g", `{"command":"g","sessionId":"s1"}`},
		{"hint h", `{"command":"h","sessionId":"s1"}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := exec(t, tc.body)
			r.CodeIs(http.StatusOK)
			r.BodyIs(mockOutput)
		})
	}

	// クライアントとサーバでキー名が食い違うとここだけが気付ける (#5289)。
	t.Run("play missing cardIndex", func(t *testing.T) {
		exec(t, `{"command":"p","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("trump missing suit", func(t *testing.T) {
		exec(t, `{"command":"t","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	// **省略された bid を 0 と読まない。** 読むと勝手に Dash Call になる。
	t.Run("bid missing bid", func(t *testing.T) {
		exec(t, `{"command":"b","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestEstimationWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultEstimationConfig().Rounds
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrEst(0), def},
		{"above the maximum", intPtrEst(999), def},
		{"in range is kept", intPtrEst(9), 9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := (&controller.EstimationWebConfig{Rounds: tc.in}).ToConfig().Rounds; got != tc.want {
				t.Fatalf("Rounds = %d, want %d", got, tc.want)
			}
		})
	}
}
