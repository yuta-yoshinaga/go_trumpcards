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

func intPtrHpf(v int) *int { return &v }

func mustHasenpfefferOutputJSON(msg string) string {
	out := &controller.HasenpfefferWebOutput{
		Players:       []*controller.HasenpfefferWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		Scores:        []int{},
		TeamTricks:    []int{},
		DeclarerIdx:   -1,
		WinnerTeam:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustHasenpfefferOutputJSON: %v", err))
	}
	return string(b)
}

func TestHasenpfefferWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	hiMock := new(usecase.MockHasenpfefferInteractor)
	hiMock.On("ResetWithConfig", domain.DefaultHasenpfefferConfig()).Return(mockOutput)
	hiMock.On("ResetWithConfig", domain.HasenpfefferConfig{Target: 15}).Return(mockOutput)
	hiMock.On("Bid", 4).Return(mockOutput)
	hiMock.On("Bid", 0).Return(mockOutput)
	hiMock.On("Discard", 2, 3).Return(mockOutput)
	hiMock.On("NextHand").Return(mockOutput)
	hiMock.On("GiveUp").Return(mockOutput)
	hiMock.On("Hint").Return(mockOutput)
	hiMock.On("ActionLog").Return(mockOutput)
	hiMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewHasenpfefferWebController(func() uc.HasenpfefferInteractorIF { return hiMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.HasenpfefferWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustHasenpfefferOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with target", `{"command":"reset","sessionId":"s1","config":{"target":15}}`},
		{"bid b", `{"command":"b","sessionId":"s1","bid":4}`},
		{"pass via bid 0", `{"command":"bid","sessionId":"s1","bid":0}`},
		{"discard d", `{"command":"d","sessionId":"s1","cardIndex":2,"suit":3}`},
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

	// **未指定を 0 で埋めない。** 埋めると降りるつもりの無い人が降ろされる。
	t.Run("bid missing bid", func(t *testing.T) {
		exec(t, `{"command":"b","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	// **捨て札は 2 引数とも要る。** 埋めると選んでいないスートが切り札になる。
	t.Run("discard missing cardIndex", func(t *testing.T) {
		exec(t, `{"command":"d","sessionId":"s1","suit":3}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("discard missing suit", func(t *testing.T) {
		exec(t, `{"command":"d","sessionId":"s1","cardIndex":2}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestHasenpfefferWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultHasenpfefferConfig().Target
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrHpf(domain.HasenpfefferTargetMin - 1), def},
		{"above the maximum", intPtrHpf(domain.HasenpfefferTargetMax + 1), def},
		{"the minimum is kept", intPtrHpf(domain.HasenpfefferTargetMin), domain.HasenpfefferTargetMin},
		{"the maximum is kept", intPtrHpf(domain.HasenpfefferTargetMax), domain.HasenpfefferTargetMax},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := (&controller.HasenpfefferWebConfig{Target: tc.in}).ToConfig().Target; got != tc.want {
				t.Fatalf("Target = %d, want %d", got, tc.want)
			}
		})
	}
}
