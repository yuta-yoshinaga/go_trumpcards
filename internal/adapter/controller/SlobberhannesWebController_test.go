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

func intPtrSlob(v int) *int { return &v }

func mustSlobberhannesOutputJSON(msg string) string {
	out := &controller.SlobberhannesWebOutput{
		Players:       []*controller.SlobberhannesWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSlobberhannesOutputJSON: %v", err))
	}
	return string(b)
}

func TestSlobberhannesWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	siMock := new(usecase.MockSlobberhannesInteractor)
	siMock.On("ResetWithConfig", domain.DefaultSlobberhannesConfig()).Return(mockOutput)
	siMock.On("ResetWithConfig", domain.SlobberhannesConfig{Rounds: 6}).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("GiveUp").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)
	siMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewSlobberhannesWebController(func() uc.SlobberhannesInteractorIF { return siMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.SlobberhannesWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustSlobberhannesOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with rounds", `{"command":"reset","sessionId":"s1","config":{"rounds":6}}`},
		{"next n", `{"command":"n","sessionId":"s1"}`},
		{"giveup g", `{"command":"g","sessionId":"s1"}`},
		{"hint h", `{"command":"h","sessionId":"s1"}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
		{"play p", `{"command":"p","sessionId":"s1","cardIndex":4}`},
		{"play long", `{"command":"play","sessionId":"s1","cardIndex":4}`},
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

// 範囲外のラウンド数は既定値に丸められる（ドメインまで壊れた値を通さない）。
func TestSlobberhannesWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultSlobberhannesConfig().Rounds
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrSlob(0), def},
		{"above the maximum", intPtrSlob(domain.SlobberhannesRoundsMax + 1), def},
		{"in range is kept", intPtrSlob(6), 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := (&controller.SlobberhannesWebConfig{Rounds: tc.in}).ToConfig()
			if cfg.Rounds != tc.want {
				t.Fatalf("Rounds = %d, want %d", cfg.Rounds, tc.want)
			}
		})
	}
}
