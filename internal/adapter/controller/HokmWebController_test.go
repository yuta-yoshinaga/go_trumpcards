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

func intPtrHk(v int) *int { return &v }

func mustHokmOutputJSON(msg string) string {
	out := &controller.HokmWebOutput{
		Players:        []*controller.HokmWebOutputPlayer{},
		CurrentTrick:   []*controller.WebOutputTrickCard{},
		ValidPlays:     []int{},
		Scores:         []int{},
		TeamTricks:     []int{},
		TricksToWin:    domain.HokmTricksToWin,
		LastHandWinner: -1,
		WinnerTeam:     -1,
		WebOutputBase:  controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustHokmOutputJSON: %v", err))
	}
	return string(b)
}

func TestHokmWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	hiMock := new(usecase.MockHokmInteractor)
	hiMock.On("ResetWithConfig", domain.DefaultHokmConfig()).Return(mockOutput)
	hiMock.On("ResetWithConfig", domain.HokmConfig{Target: 9}).Return(mockOutput)
	hiMock.On("DeclareTrump", 3).Return(mockOutput)
	hiMock.On("NextHand").Return(mockOutput)
	hiMock.On("GiveUp").Return(mockOutput)
	hiMock.On("Hint").Return(mockOutput)
	hiMock.On("ActionLog").Return(mockOutput)
	hiMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewHokmWebController(func() uc.HokmInteractorIF { return hiMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.HokmWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustHokmOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with target", `{"command":"reset","sessionId":"s1","config":{"target":9}}`},
		{"trump t", `{"command":"t","sessionId":"s1","suit":3}`},
		{"trump long", `{"command":"trump","sessionId":"s1","suit":3}`},
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

	// **スート無しの宣言は通さない。** 既定値で埋めると選んでいないスートが切り札になる。
	t.Run("trump missing suit", func(t *testing.T) {
		exec(t, `{"command":"t","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestHokmWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultHokmConfig().Target
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrHk(0), def},
		{"above the maximum", intPtrHk(999), def},
		{"in range is kept", intPtrHk(9), 9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := (&controller.HokmWebConfig{Target: tc.in}).ToConfig().Target; got != tc.want {
				t.Fatalf("Target = %d, want %d", got, tc.want)
			}
		})
	}
}
