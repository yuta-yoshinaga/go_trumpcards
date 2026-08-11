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

func intPtrRev(v int) *int { return &v }

func mustReversisOutputJSON(msg string) string {
	out := &controller.ReversisWebOutput{
		Players:       []*controller.ReversisWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustReversisOutputJSON: %v", err))
	}
	return string(b)
}

func TestReversisWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	riMock := new(usecase.MockReversisInteractor)
	riMock.On("ResetWithConfig", domain.DefaultReversisConfig()).Return(mockOutput)
	riMock.On("ResetWithConfig", domain.ReversisConfig{Rounds: 6}).Return(mockOutput)
	riMock.On("NextRound").Return(mockOutput)
	riMock.On("GiveUp").Return(mockOutput)
	riMock.On("Hint").Return(mockOutput)
	riMock.On("ActionLog").Return(mockOutput)
	riMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewReversisWebController(func() uc.ReversisInteractorIF { return riMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.ReversisWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustReversisOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with rounds", `{"command":"reset","sessionId":"s1","config":{"rounds":6}}`},
		{"next n", `{"command":"n","sessionId":"s1"}`},
		{"giveup g", `{"command":"g","sessionId":"s1"}`},
		{"hint h", `{"command":"h","sessionId":"s1"}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
		{"play p", `{"command":"p","sessionId":"s1","cardIndex":4}`},
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

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestReversisWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultReversisConfig().Rounds
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrRev(0), def},
		{"above the maximum", intPtrRev(domain.ReversisRoundsMax + 1), def},
		{"in range is kept", intPtrRev(6), 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := (&controller.ReversisWebConfig{Rounds: tc.in}).ToConfig().Rounds; got != tc.want {
				t.Fatalf("Rounds = %d, want %d", got, tc.want)
			}
		})
	}
}
