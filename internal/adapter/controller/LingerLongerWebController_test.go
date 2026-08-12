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

func intPtrLl(v int) *int { return &v }

func mustLingerLongerOutputJSON(msg string) string {
	out := &controller.LingerLongerWebOutput{
		Players:       []*controller.LingerLongerWebOutputPlayer{},
		ValidPlays:    []int{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		LastDrawIdx:   -1,
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustLingerLongerOutputJSON: %v", err))
	}
	return string(b)
}

func TestLingerLongerWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"validPlays":[],"currentTrick":[],"message":""}`

	liMock := new(usecase.MockLingerLongerInteractor)
	liMock.On("ResetWithConfig", domain.DefaultLingerLongerConfig()).Return(mockOutput)
	liMock.On("ResetWithConfig", domain.LingerLongerConfig{PlayerCnt: 6}).Return(mockOutput)
	liMock.On("Play", 4).Return(mockOutput)
	liMock.On("GiveUp").Return(mockOutput)
	liMock.On("Hint").Return(mockOutput)
	liMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewLingerLongerWebController(func() uc.LingerLongerInteractorIF { return liMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.LingerLongerWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustLingerLongerOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with players", `{"command":"reset","sessionId":"s1","config":{"playerCnt":6}}`},
		{"play p", `{"command":"p","sessionId":"s1","cardIndex":4}`},
		{"play long", `{"command":"play","sessionId":"s1","cardIndex":4}`},
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

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestLingerLongerWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultLingerLongerConfig().PlayerCnt
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrLl(domain.LingerLongerPlayerCntMin - 1), def},
		{"above the maximum", intPtrLl(domain.LingerLongerPlayerCntMax + 1), def},
		{"the minimum is kept", intPtrLl(domain.LingerLongerPlayerCntMin), domain.LingerLongerPlayerCntMin},
		{"the maximum is kept", intPtrLl(domain.LingerLongerPlayerCntMax), domain.LingerLongerPlayerCntMax},
		{"five is kept", intPtrLl(5), 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := (&controller.LingerLongerWebConfig{PlayerCnt: tc.in}).ToConfig()
			if cfg.PlayerCnt != tc.want {
				t.Fatalf("PlayerCnt = %d, want %d", cfg.PlayerCnt, tc.want)
			}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("clamped config must always validate: %v", err)
			}
		})
	}

	var input controller.LingerLongerWebInput
	if got := input.ToConfig(); got != domain.DefaultLingerLongerConfig() {
		t.Fatalf("ToConfig() = %+v, want the default", got)
	}
}
