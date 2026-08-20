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

func intPtrCu(v int) *int { return &v }

func mustCucumberOutputJSON(msg string) string {
	out := &controller.CucumberWebOutput{
		Players:            []*controller.CucumberWebOutputPlayer{},
		ValidPlays:         []int{},
		CurrentTrick:       []*controller.WebOutputTrickCard{},
		LastTrickWinnerIdx: -1,
		// エラー応答でも総トリック数は規則どおりの固定値。
		TotalTricks:   domain.CucumberHandSize,
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustCucumberOutputJSON: %v", err))
	}
	return string(b)
}

func TestCucumberWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"validPlays":[],"currentTrick":[],"message":""}`

	ciMock := new(usecase.MockCucumberInteractor)
	ciMock.On("ResetWithConfig", domain.DefaultCucumberConfig()).Return(mockOutput)
	ciMock.On("ResetWithConfig", domain.CucumberConfig{PlayerCnt: 6, TargetScore: 50}).Return(mockOutput)
	ciMock.On("Play", 4).Return(mockOutput)
	ciMock.On("NextRound").Return(mockOutput)
	ciMock.On("GiveUp").Return(mockOutput)
	ciMock.On("Hint").Return(mockOutput)
	ciMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewCucumberWebController(func() uc.CucumberInteractorIF { return ciMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.CucumberWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustCucumberOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with config", `{"command":"reset","sessionId":"s1","config":{"playerCnt":6,"targetScore":50}}`},
		{"play p", `{"command":"p","sessionId":"s1","cardIndex":4}`},
		{"play long", `{"command":"play","sessionId":"s1","cardIndex":4}`},
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
	t.Run("play missing cardIndex", func(t *testing.T) {
		exec(t, `{"command":"p","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestCucumberWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultCucumberConfig()
	for _, tc := range []struct {
		name    string
		players *int
		target  *int
		want    domain.CucumberConfig
	}{
		{"nil uses the default", nil, nil, def},
		{"below the minimum", intPtrCu(domain.CucumberPlayerCntMin - 1), nil, def},
		{"above the maximum", intPtrCu(domain.CucumberPlayerCntMax + 1), nil, def},
		{"the minimum is kept", intPtrCu(domain.CucumberPlayerCntMin), nil,
			domain.CucumberConfig{PlayerCnt: domain.CucumberPlayerCntMin, TargetScore: def.TargetScore}},
		{"the maximum is kept", intPtrCu(domain.CucumberPlayerCntMax), nil,
			domain.CucumberConfig{PlayerCnt: domain.CucumberPlayerCntMax, TargetScore: def.TargetScore}},
		{"target score is kept", nil, intPtrCu(50),
			domain.CucumberConfig{PlayerCnt: def.PlayerCnt, TargetScore: 50}},
		{"target score out of range falls back", nil, intPtrCu(9), def},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := (&controller.CucumberWebConfig{PlayerCnt: tc.players, TargetScore: tc.target}).ToConfig()
			if cfg != tc.want {
				t.Fatalf("ToConfig() = %+v, want %+v", cfg, tc.want)
			}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("clamped config must always validate: %v", err)
			}
		})
	}

	var input controller.CucumberWebInput
	if got := input.ToConfig(); got != domain.DefaultCucumberConfig() {
		t.Fatalf("ToConfig() = %+v, want the default", got)
	}
}
