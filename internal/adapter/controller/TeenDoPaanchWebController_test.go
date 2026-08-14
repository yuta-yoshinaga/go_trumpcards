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

func intPtrTdp(v int) *int { return &v }

func mustTeenDoPaanchOutputJSON(msg string) string {
	out := &controller.TeenDoPaanchWebOutput{
		Players:       []*controller.TeenDoPaanchWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTeenDoPaanchOutputJSON: %v", err))
	}
	return string(b)
}

func TestTeenDoPaanchWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	tiMock := new(usecase.MockTeenDoPaanchInteractor)
	tiMock.On("ResetWithConfig", domain.DefaultTeenDoPaanchConfig()).Return(mockOutput)
	tiMock.On("ResetWithConfig", domain.TeenDoPaanchConfig{Rounds: 6}).Return(mockOutput)
	tiMock.On("DeclareTrump", 3).Return(mockOutput)
	tiMock.On("NextRound").Return(mockOutput)
	tiMock.On("GiveUp").Return(mockOutput)
	tiMock.On("Hint").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)
	tiMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewTeenDoPaanchWebController(func() uc.TeenDoPaanchInteractorIF { return tiMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.TeenDoPaanchWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustTeenDoPaanchOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with rounds", `{"command":"reset","sessionId":"s1","config":{"rounds":6}}`},
		{"trump t", `{"command":"t","sessionId":"s1","suit":3}`},
		{"trump long", `{"command":"trump","sessionId":"s1","suit":3}`},
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

	// **スート無しの宣言は通さない。** 既定値で埋めると選んでいないスートが切り札になる。
	t.Run("trump missing suit", func(t *testing.T) {
		exec(t, `{"command":"t","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	// **ノルマを宣言するコマンドは無い。**
	t.Run("no bid command", func(t *testing.T) {
		exec(t, `{"command":"bid","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestTeenDoPaanchWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultTeenDoPaanchConfig().Rounds
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrTdp(domain.TeenDoPaanchRoundsMin - 1), def},
		{"above the maximum", intPtrTdp(domain.TeenDoPaanchRoundsMax + 1), def},
		{"the minimum is kept", intPtrTdp(domain.TeenDoPaanchRoundsMin), domain.TeenDoPaanchRoundsMin},
		{"the maximum is kept", intPtrTdp(domain.TeenDoPaanchRoundsMax), domain.TeenDoPaanchRoundsMax},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := (&controller.TeenDoPaanchWebConfig{Rounds: tc.in}).ToConfig().Rounds; got != tc.want {
				t.Fatalf("Rounds = %d, want %d", got, tc.want)
			}
		})
	}
}
