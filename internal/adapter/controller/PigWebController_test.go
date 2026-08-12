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

func intPtrPig(v int) *int { return &v }

func mustPigOutputJSON(msg string) string {
	out := &controller.PigWebOutput{
		Players:       []*controller.PigWebOutputPlayer{},
		ValidPlays:    []int{},
		SignallerIdx:  -1,
		RoundLoserIdx: -1,
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustPigOutputJSON: %v", err))
	}
	return string(b)
}

func TestPigWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"validPlays":[],"message":""}`

	piMock := new(usecase.MockPigInteractor)
	piMock.On("ResetWithConfig", domain.DefaultPigConfig()).Return(mockOutput)
	piMock.On("ResetWithConfig", domain.PigConfig{PlayerCnt: 6, CpuDifficulty: domain.PigCpuHard}).Return(mockOutput)
	piMock.On("Pass", 3).Return(mockOutput)
	piMock.On("Signal").Return(mockOutput)
	piMock.On("NextRound").Return(mockOutput)
	piMock.On("GiveUp").Return(mockOutput)
	piMock.On("Hint").Return(mockOutput)
	piMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewPigWebController(func() uc.PigInteractorIF { return piMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.PigWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustPigOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with config", `{"command":"reset","sessionId":"s1","config":{"playerCnt":6,"cpuDifficulty":2}}`},
		{"pass p", `{"command":"p","sessionId":"s1","cardIndex":3}`},
		{"pass long", `{"command":"pass","sessionId":"s1","cardIndex":3}`},
		// **合図と次ラウンドは引数を取らない別のコマンド。**
		{"signal s", `{"command":"s","sessionId":"s1"}`},
		{"signal long", `{"command":"signal","sessionId":"s1"}`},
		{"next n", `{"command":"n","sessionId":"s1"}`},
		{"next long", `{"command":"next","sessionId":"s1"}`},
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
	t.Run("pass missing cardIndex", func(t *testing.T) {
		exec(t, `{"command":"p","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestPigWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultPigConfig()
	for _, tc := range []struct {
		name    string
		players *int
		diff    *int
		want    domain.PigConfig
	}{
		{"nil uses the default", nil, nil, def},
		{"below the minimum", intPtrPig(domain.PigPlayerCntMin - 1), nil, def},
		{"above the maximum", intPtrPig(domain.PigPlayerCntMax + 1), nil, def},
		{"the minimum is kept", intPtrPig(domain.PigPlayerCntMin), nil,
			domain.PigConfig{PlayerCnt: domain.PigPlayerCntMin, CpuDifficulty: def.CpuDifficulty}},
		{"the maximum is kept", intPtrPig(domain.PigPlayerCntMax), nil,
			domain.PigConfig{PlayerCnt: domain.PigPlayerCntMax, CpuDifficulty: def.CpuDifficulty}},
		{"difficulty is kept", nil, intPtrPig(int(domain.PigCpuHard)),
			domain.PigConfig{PlayerCnt: def.PlayerCnt, CpuDifficulty: domain.PigCpuHard}},
		{"difficulty out of range falls back", nil, intPtrPig(9), def},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := (&controller.PigWebConfig{PlayerCnt: tc.players, CpuDifficulty: tc.diff}).ToConfig()
			if cfg != tc.want {
				t.Fatalf("ToConfig() = %+v, want %+v", cfg, tc.want)
			}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("clamped config must always validate: %v", err)
			}
		})
	}

	var input controller.PigWebInput
	if got := input.ToConfig(); got != domain.DefaultPigConfig() {
		t.Fatalf("ToConfig() = %+v, want the default", got)
	}
}
