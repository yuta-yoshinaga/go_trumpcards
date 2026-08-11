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

func intPtrSm(v int) *int { return &v }

func mustSergeantMajorOutputJSON(msg string) string {
	out := &controller.SergeantMajorWebOutput{
		Players:       []*controller.SergeantMajorWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		DiscardCount:  domain.SergeantMajorKittySize,
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSergeantMajorOutputJSON: %v", err))
	}
	return string(b)
}

func TestSergeantMajorWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	siMock := new(usecase.MockSergeantMajorInteractor)
	siMock.On("ResetWithConfig", domain.DefaultSergeantMajorConfig()).Return(mockOutput)
	siMock.On("ResetWithConfig", domain.SergeantMajorConfig{Rounds: 6}).Return(mockOutput)
	siMock.On("DeclareTrump", 3).Return(mockOutput)
	siMock.On("Discard", []int{0, 1, 2, 3}).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("GiveUp").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)
	siMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewSergeantMajorWebController(func() uc.SergeantMajorInteractorIF { return siMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.SergeantMajorWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustSergeantMajorOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with rounds", `{"command":"reset","sessionId":"s1","config":{"rounds":6}}`},
		{"trump t", `{"command":"t","sessionId":"s1","suit":3}`},
		{"discard d", `{"command":"d","sessionId":"s1","discards":[0,1,2,3]}`},
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

	// **スート無しの宣言は通さない。**
	t.Run("trump missing suit", func(t *testing.T) {
		exec(t, `{"command":"t","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	// **捨て札の指定無しは通さない。**
	t.Run("discard missing discards", func(t *testing.T) {
		exec(t, `{"command":"d","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	// **ノルマを宣言するコマンドは無い。**
	t.Run("no bid command", func(t *testing.T) {
		exec(t, `{"command":"bid","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestSergeantMajorWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultSergeantMajorConfig().Rounds
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrSm(domain.SergeantMajorRoundsMin - 1), def},
		{"above the maximum", intPtrSm(domain.SergeantMajorRoundsMax + 1), def},
		{"the minimum is kept", intPtrSm(domain.SergeantMajorRoundsMin), domain.SergeantMajorRoundsMin},
		{"the maximum is kept", intPtrSm(domain.SergeantMajorRoundsMax), domain.SergeantMajorRoundsMax},
		// **3 の倍数でないと親の役が一巡しない。** エラーにせず丸める。
		{"4 snaps down to 3", intPtrSm(4), 3},
		{"5 snaps down to 3", intPtrSm(5), 3},
		{"7 snaps down to 6", intPtrSm(7), 6},
		{"29 snaps down to 27", intPtrSm(29), 27},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := (&controller.SergeantMajorWebConfig{Rounds: tc.in}).ToConfig()
			if cfg.Rounds != tc.want {
				t.Fatalf("Rounds = %d, want %d", cfg.Rounds, tc.want)
			}
			// **丸めた結果は必ずドメインが受け取れる。**
			if err := cfg.Validate(); err != nil {
				t.Fatalf("ToConfig produced a config the domain rejects: %v", err)
			}
		})
	}
}
