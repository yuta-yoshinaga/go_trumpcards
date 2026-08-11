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

func intPtrBh(v int) *int { return &v }

func mustBhabhiOutputJSON(msg string) string {
	out := &controller.BhabhiWebOutput{
		Players:         []*controller.BhabhiWebOutputPlayer{},
		Pile:            []*controller.WebOutputTrickCard{},
		ValidPlays:      []int{},
		LastPickupIdx:   -1,
		BhabhiIdx:       -1,
		StalemateTricks: domain.BhabhiStalemateTricks,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBhabhiOutputJSON: %v", err))
	}
	return string(b)
}

func TestBhabhiWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"pile":[],"validPlays":[],"message":""}`

	biMock := new(usecase.MockBhabhiInteractor)
	biMock.On("ResetWithConfig", domain.DefaultBhabhiConfig()).Return(mockOutput)
	biMock.On("ResetWithConfig", domain.BhabhiConfig{PlayerCnt: 6}).Return(mockOutput)
	biMock.On("GiveUp").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)
	biMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewBhabhiWebController(func() uc.BhabhiInteractorIF { return biMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.BhabhiWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustBhabhiOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with player count", `{"command":"reset","sessionId":"s1","config":{"playerCnt":6}}`},
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

	// **ハンドの区切りが無いので next は無い。**
	t.Run("no next command", func(t *testing.T) {
		exec(t, `{"command":"n","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

// **人数は 3〜7 人に丸める。** 範囲外は既定に落とす。
func TestBhabhiWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultBhabhiConfig().PlayerCnt
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrBh(domain.BhabhiMinPlayers - 1), def},
		{"above the maximum", intPtrBh(domain.BhabhiMaxPlayers + 1), def},
		{"the minimum is kept", intPtrBh(domain.BhabhiMinPlayers), domain.BhabhiMinPlayers},
		{"the maximum is kept", intPtrBh(domain.BhabhiMaxPlayers), domain.BhabhiMaxPlayers},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := (&controller.BhabhiWebConfig{PlayerCnt: tc.in}).ToConfig().PlayerCnt; got != tc.want {
				t.Fatalf("PlayerCnt = %d, want %d", got, tc.want)
			}
		})
	}
}
