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

func intPtrSb(v int) *int { return &v }

func mustStealingBundlesOutputJSON(msg string) string {
	out := &controller.StealingBundlesWebOutput{
		Players:        []*controller.StealingBundlesWebOutputPlayer{},
		TableCards:     []*controller.WebOutputCard{},
		TableMatches:   map[string][]int{},
		StealTargets:   map[string][]int{},
		LastCaptureIdx: -1,
		WinnerIdx:      -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustStealingBundlesOutputJSON: %v", err))
	}
	return string(b)
}

func TestStealingBundlesWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"tableCards":[],"message":""}`

	siMock := new(usecase.MockStealingBundlesInteractor)
	siMock.On("ResetWithConfig", domain.DefaultStealingBundlesConfig()).Return(mockOutput)
	siMock.On("ResetWithConfig", domain.StealingBundlesConfig{PlayerCnt: 2}).Return(mockOutput)
	siMock.On("Take", 3).Return(mockOutput)
	siMock.On("Steal", 1, 2).Return(mockOutput)
	siMock.On("Trail", 0).Return(mockOutput)
	siMock.On("GiveUp").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewStealingBundlesWebController(func() uc.StealingBundlesInteractorIF { return siMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.StealingBundlesWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustStealingBundlesOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with players", `{"command":"reset","sessionId":"s1","config":{"playerCnt":2}}`},
		{"take t", `{"command":"t","sessionId":"s1","cardIndex":3}`},
		{"take long", `{"command":"take","sessionId":"s1","cardIndex":3}`},
		{"steal s", `{"command":"s","sessionId":"s1","cardIndex":1,"victimIdx":2}`},
		{"steal long", `{"command":"steal","sessionId":"s1","cardIndex":1,"victimIdx":2}`},
		{"trail d", `{"command":"d","sessionId":"s1","cardIndex":0}`},
		{"trail long", `{"command":"trail","sessionId":"s1","cardIndex":0}`},
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
	for _, tc := range []struct{ name, body string }{
		{"take missing cardIndex", `{"command":"t","sessionId":"s1"}`},
		{"trail missing cardIndex", `{"command":"d","sessionId":"s1"}`},
		// **略奪は相手の指名が要る。** 片方だけでは動かしません。
		{"steal missing victimIdx", `{"command":"s","sessionId":"s1","cardIndex":1}`},
		{"steal missing cardIndex", `{"command":"s","sessionId":"s1","victimIdx":2}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exec(t, tc.body).CodeIs(http.StatusBadRequest)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestStealingBundlesWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultStealingBundlesConfig().PlayerCnt
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrSb(domain.StealingBundlesPlayerCntMin - 1), def},
		{"above the maximum", intPtrSb(domain.StealingBundlesPlayerCntMax + 1), def},
		{"the minimum is kept", intPtrSb(domain.StealingBundlesPlayerCntMin), domain.StealingBundlesPlayerCntMin},
		{"the maximum is kept", intPtrSb(domain.StealingBundlesPlayerCntMax), domain.StealingBundlesPlayerCntMax},
		{"three is kept", intPtrSb(3), 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := (&controller.StealingBundlesWebConfig{PlayerCnt: tc.in}).ToConfig()
			if cfg.PlayerCnt != tc.want {
				t.Fatalf("PlayerCnt = %d, want %d", cfg.PlayerCnt, tc.want)
			}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("clamped config must always validate: %v", err)
			}
		})
	}

	var input controller.StealingBundlesWebInput
	if got := input.ToConfig(); got != domain.DefaultStealingBundlesConfig() {
		t.Fatalf("ToConfig() = %+v, want the default", got)
	}
}
