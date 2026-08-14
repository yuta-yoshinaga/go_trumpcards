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

func intPtrGs(v int) *int { return &v }

func mustGoofspielOutputJSON(msg string) string {
	out := &controller.GoofspielWebOutput{
		Players:       []*controller.GoofspielWebOutputPlayer{},
		ValidPlays:    []int{},
		CarriedPrizes: []*controller.WebOutputCard{},
		LastWinnerIdx: -1,
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustGoofspielOutputJSON: %v", err))
	}
	return string(b)
}

func TestGoofspielWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"validPlays":[],"carriedPrizes":[],"message":""}`

	giMock := new(usecase.MockGoofspielInteractor)
	giMock.On("ResetWithConfig", domain.DefaultGoofspielConfig()).Return(mockOutput)
	giMock.On("ResetWithConfig", domain.GoofspielConfig{PlayerCnt: 3, TieRule: domain.GoofspielTieCarryOver}).Return(mockOutput)
	giMock.On("Bid", 4).Return(mockOutput)
	giMock.On("NextRound").Return(mockOutput)
	giMock.On("GiveUp").Return(mockOutput)
	giMock.On("Hint").Return(mockOutput)
	giMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewGoofspielWebController(func() uc.GoofspielInteractorIF { return giMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.GoofspielWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustGoofspielOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with config", `{"command":"reset","sessionId":"s1","config":{"playerCnt":3,"tieRule":1}}`},
		{"bid b", `{"command":"b","sessionId":"s1","cardIndex":4}`},
		{"bid long", `{"command":"bid","sessionId":"s1","cardIndex":4}`},
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
	t.Run("bid missing cardIndex", func(t *testing.T) {
		exec(t, `{"command":"b","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestGoofspielWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultGoofspielConfig()
	for _, tc := range []struct {
		name    string
		players *int
		tie     *int
		want    domain.GoofspielConfig
	}{
		{"nil uses the default", nil, nil, def},
		{"below the minimum", intPtrGs(domain.GoofspielPlayerCntMin - 1), nil, def},
		{"above the maximum", intPtrGs(domain.GoofspielPlayerCntMax + 1), nil, def},
		{"the minimum is kept", intPtrGs(domain.GoofspielPlayerCntMin), nil,
			domain.GoofspielConfig{PlayerCnt: domain.GoofspielPlayerCntMin, TieRule: def.TieRule}},
		{"the maximum is kept", intPtrGs(domain.GoofspielPlayerCntMax), nil,
			domain.GoofspielConfig{PlayerCnt: domain.GoofspielPlayerCntMax, TieRule: def.TieRule}},
		{"carry-over is kept", nil, intPtrGs(int(domain.GoofspielTieCarryOver)),
			domain.GoofspielConfig{PlayerCnt: def.PlayerCnt, TieRule: domain.GoofspielTieCarryOver}},
		{"tie rule out of range falls back", nil, intPtrGs(9), def},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := (&controller.GoofspielWebConfig{PlayerCnt: tc.players, TieRule: tc.tie}).ToConfig()
			if cfg != tc.want {
				t.Fatalf("ToConfig() = %+v, want %+v", cfg, tc.want)
			}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("clamped config must always validate: %v", err)
			}
		})
	}

	var input controller.GoofspielWebInput
	if got := input.ToConfig(); got != domain.DefaultGoofspielConfig() {
		t.Fatalf("ToConfig() = %+v, want the default", got)
	}
}
