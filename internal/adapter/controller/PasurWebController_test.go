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

func intPtrPs(v int) *int { return &v }

func mustPasurOutputJSON(msg string) string {
	out := &controller.PasurWebOutput{
		Players:        []*controller.PasurWebOutputPlayer{},
		Table:          []*controller.WebOutputCard{},
		CaptureOptions: [][][]int{},
		Winners:        []int{},
		LastCaptureIdx: -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustPasurOutputJSON: %v", err))
	}
	return string(b)
}

func TestPasurWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"table":[],"message":""}`

	piMock := new(usecase.MockPasurInteractor)
	piMock.On("ResetWithConfig", domain.DefaultPasurConfig()).Return(mockOutput)
	piMock.On("ResetWithConfig", domain.PasurConfig{PlayerCnt: 2}).Return(mockOutput)
	piMock.On("GiveUp").Return(mockOutput)
	piMock.On("Hint").Return(mockOutput)
	piMock.On("ActionLog").Return(mockOutput)
	piMock.On("Play", 4, []int{0, 1}).Return(mockOutput)
	piMock.On("Play", 4, []int(nil)).Return(mockOutput)

	ctrl := controller.NewPasurWebController(func() uc.PasurInteractorIF { return piMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.PasurWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustPasurOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with players", `{"command":"reset","sessionId":"s1","config":{"playerCnt":2}}`},
		{"play with a capture", `{"command":"p","sessionId":"s1","cardIndex":4,"table":[0,1]}`},
		// **table の省略はトレール。** エラーではない。
		{"play as a trail", `{"command":"play","sessionId":"s1","cardIndex":4}`},
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

func TestPasurWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultPasurConfig().PlayerCnt
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrPs(domain.PasurPlayerCntMin - 1), def},
		{"above the maximum", intPtrPs(domain.PasurPlayerCntMax + 1), def},
		{"the minimum is kept", intPtrPs(domain.PasurPlayerCntMin), domain.PasurPlayerCntMin},
		{"the maximum is kept", intPtrPs(domain.PasurPlayerCntMax), domain.PasurPlayerCntMax},
		{"three is kept", intPtrPs(3), 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := (&controller.PasurWebConfig{PlayerCnt: tc.in}).ToConfig()
			if cfg.PlayerCnt != tc.want {
				t.Fatalf("PlayerCnt = %d, want %d", cfg.PlayerCnt, tc.want)
			}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("clamped config must always validate: %v", err)
			}
		})
	}

	var input controller.PasurWebInput
	if got := input.ToConfig(); got != domain.DefaultPasurConfig() {
		t.Fatalf("ToConfig() = %+v, want the default", got)
	}
}
