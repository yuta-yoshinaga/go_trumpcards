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

func intPtrHb(v int) *int { return &v }

func mustHoneymoonBridgeOutputJSON(msg string) string {
	out := &controller.HoneymoonBridgeWebOutput{
		Players:       []*controller.HoneymoonBridgeWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		DeclarerIdx:   -1,
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustHoneymoonBridgeOutputJSON: %v", err))
	}
	return string(b)
}

func TestHoneymoonBridgeWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	hiMock := new(usecase.MockHoneymoonBridgeInteractor)
	hiMock.On("ResetWithConfig", domain.DefaultHoneymoonBridgeConfig()).Return(mockOutput)
	hiMock.On("ResetWithConfig", domain.HoneymoonBridgeConfig{Target: 200}).Return(mockOutput)
	hiMock.On("Bid", 3, 0).Return(mockOutput)
	hiMock.On("Bid", 2, 4).Return(mockOutput)
	hiMock.On("Pass").Return(mockOutput)
	hiMock.On("NextRound").Return(mockOutput)
	hiMock.On("GiveUp").Return(mockOutput)
	hiMock.On("Hint").Return(mockOutput)
	hiMock.On("ActionLog").Return(mockOutput)
	hiMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewHoneymoonBridgeWebController(func() uc.HoneymoonBridgeInteractorIF { return hiMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.HoneymoonBridgeWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustHoneymoonBridgeOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with target", `{"command":"reset","sessionId":"s1","config":{"target":200}}`},
		// **ノートランプは suit:0。** 省略とは違う。
		{"bid no-trump", `{"command":"b","sessionId":"s1","level":3,"suit":0}`},
		{"bid a suit", `{"command":"bid","sessionId":"s1","level":2,"suit":4}`},
		{"pass", `{"command":"pass","sessionId":"s1"}`},
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

	// **レベル無しの宣言は通さない。**
	t.Run("bid missing level", func(t *testing.T) {
		exec(t, `{"command":"b","sessionId":"s1","suit":3}`).CodeIs(http.StatusBadRequest)
	})

	// **スート無しの宣言も通さない。** 0 埋めするとノートランプを宣言してしまう。
	t.Run("bid missing suit", func(t *testing.T) {
		exec(t, `{"command":"b","sessionId":"s1","level":3}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestHoneymoonBridgeWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultHoneymoonBridgeConfig().Target
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrHb(domain.HoneymoonBridgeTargetMin - 1), def},
		{"above the maximum", intPtrHb(domain.HoneymoonBridgeTargetMax + 1), def},
		{"the minimum is kept", intPtrHb(domain.HoneymoonBridgeTargetMin), domain.HoneymoonBridgeTargetMin},
		{"the maximum is kept", intPtrHb(domain.HoneymoonBridgeTargetMax), domain.HoneymoonBridgeTargetMax},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := (&controller.HoneymoonBridgeWebConfig{Target: tc.in}).ToConfig()
			if cfg.Target != tc.want {
				t.Fatalf("Target = %d, want %d", cfg.Target, tc.want)
			}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("clamped config must always validate: %v", err)
			}
		})
	}

	// 設定を省略した入力は既定値になる。
	var input controller.HoneymoonBridgeWebInput
	if got := input.ToConfig(); got != domain.DefaultHoneymoonBridgeConfig() {
		t.Fatalf("ToConfig() = %+v, want the default", got)
	}
}
