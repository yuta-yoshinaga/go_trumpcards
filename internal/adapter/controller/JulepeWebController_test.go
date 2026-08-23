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

func intPtrJulepe(v int) *int { return &v }

func mustJulepeOutputJSON(msg string) string {
	out := &controller.JulepeWebOutput{
		Players:       []*controller.JulepeWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustJulepeOutputJSON: %v", err))
	}
	return string(b)
}

func TestJulepeWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	riMock := new(usecase.MockJulepeInteractor)
	riMock.On("ResetWithConfig", domain.DefaultJulepeConfig()).Return(mockOutput)
	riMock.On("ResetWithConfig", domain.JulepeConfig{PlayerCnt: 5, Rounds: 6}).Return(mockOutput)
	riMock.On("Play").Return(mockOutput)
	riMock.On("Pass").Return(mockOutput)
	riMock.On("NextRound").Return(mockOutput)
	riMock.On("GiveUp").Return(mockOutput)
	riMock.On("Hint").Return(mockOutput)
	riMock.On("ActionLog").Return(mockOutput)
	riMock.On("PlayCard", 4).Return(mockOutput)

	ctrl := controller.NewJulepeWebController(func() uc.JulepeInteractorIF { return riMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.JulepeWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustJulepeOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with five players", `{"command":"reset","sessionId":"s1","config":{"playerCnt":5,"rounds":6}}`},
		{"play in", `{"command":"in","sessionId":"s1"}`},
		{"play long", `{"command":"play","sessionId":"s1"}`},
		{"pass out", `{"command":"out","sessionId":"s1"}`},
		{"pass long", `{"command":"pass","sessionId":"s1"}`},
		{"card c", `{"command":"c","sessionId":"s1","cardIndex":4}`},
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
	t.Run("card missing cardIndex", func(t *testing.T) {
		exec(t, `{"command":"c","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

// **人数は 3〜5 に丸める。** 範囲外をドメインへ通すと席数と手札が食い違う。
func TestJulepeWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultJulepeConfig()
	for _, tc := range []struct {
		name              string
		players, rounds   *int
		wantPlayers, want int
	}{
		{"nil uses the defaults", nil, nil, def.PlayerCnt, def.Rounds},
		{"two players is clamped", intPtrJulepe(2), nil, def.PlayerCnt, def.Rounds},
		{"six players is clamped", intPtrJulepe(6), nil, def.PlayerCnt, def.Rounds},
		{"three players is kept", intPtrJulepe(3), nil, 3, def.Rounds},
		{"five players is kept", intPtrJulepe(5), intPtrJulepe(6), 5, 6},
		{"rounds out of range is clamped", intPtrJulepe(4), intPtrJulepe(99), 4, def.Rounds},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := (&controller.JulepeWebConfig{PlayerCnt: tc.players, Rounds: tc.rounds}).ToConfig()
			if cfg.PlayerCnt != tc.wantPlayers {
				t.Fatalf("PlayerCnt = %d, want %d", cfg.PlayerCnt, tc.wantPlayers)
			}
			if cfg.Rounds != tc.want {
				t.Fatalf("Rounds = %d, want %d", cfg.Rounds, tc.want)
			}
		})
	}
}
