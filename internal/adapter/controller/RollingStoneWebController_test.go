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

func intPtrRs(v int) *int { return &v }

func mustRollingStoneOutputJSON(msg string) string {
	out := &controller.RollingStoneWebOutput{
		Players:       []*controller.RollingStoneWebOutputPlayer{},
		ValidPlays:    []int{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		LastPickupIdx: -1,
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustRollingStoneOutputJSON: %v", err))
	}
	return string(b)
}

func TestRollingStoneWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"validPlays":[],"currentTrick":[],"message":""}`

	riMock := new(usecase.MockRollingStoneInteractor)
	riMock.On("ResetWithConfig", domain.DefaultRollingStoneConfig()).Return(mockOutput)
	riMock.On("ResetWithConfig", domain.RollingStoneConfig{PlayerCnt: 6}).Return(mockOutput)
	riMock.On("Play", 4).Return(mockOutput)
	riMock.On("PickUp").Return(mockOutput)
	riMock.On("GiveUp").Return(mockOutput)
	riMock.On("Hint").Return(mockOutput)
	riMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewRollingStoneWebController(func() uc.RollingStoneInteractorIF { return riMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.RollingStoneWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustRollingStoneOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with players", `{"command":"reset","sessionId":"s1","config":{"playerCnt":6}}`},
		{"play p", `{"command":"p","sessionId":"s1","cardIndex":4}`},
		// **引き取りは別のコマンド。** カード指定が無いのは省略ではない。
		{"pickup u", `{"command":"u","sessionId":"s1"}`},
		{"pickup long", `{"command":"pickup","sessionId":"s1"}`},
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

func TestRollingStoneWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultRollingStoneConfig().PlayerCnt
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrRs(domain.RollingStonePlayerCntMin - 1), def},
		{"above the maximum", intPtrRs(domain.RollingStonePlayerCntMax + 1), def},
		{"the minimum is kept", intPtrRs(domain.RollingStonePlayerCntMin), domain.RollingStonePlayerCntMin},
		{"the maximum is kept", intPtrRs(domain.RollingStonePlayerCntMax), domain.RollingStonePlayerCntMax},
		{"five is kept", intPtrRs(5), 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := (&controller.RollingStoneWebConfig{PlayerCnt: tc.in}).ToConfig()
			if cfg.PlayerCnt != tc.want {
				t.Fatalf("PlayerCnt = %d, want %d", cfg.PlayerCnt, tc.want)
			}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("clamped config must always validate: %v", err)
			}
		})
	}

	var input controller.RollingStoneWebInput
	if got := input.ToConfig(); got != domain.DefaultRollingStoneConfig() {
		t.Fatalf("ToConfig() = %+v, want the default", got)
	}
}
