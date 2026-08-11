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

func intPtrMdk(v int) *int { return &v }

func mustMendikotOutputJSON(msg string) string {
	out := &controller.MendikotWebOutput{
		Players:         []*controller.MendikotWebOutputPlayer{},
		CurrentTrick:    []*controller.WebOutputTrickCard{},
		ValidPlays:      []int{},
		Scores:          []int{},
		TeamTens:        []int{},
		TeamTricks:      []int{},
		TensInDeck:      domain.MendikotTensInDeck,
		TrumpChooserIdx: -1,
		LastHandWinner:  -1,
		WinnerTeam:      -1,
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMendikotOutputJSON: %v", err))
	}
	return string(b)
}

func TestMendikotWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	miMock := new(usecase.MockMendikotInteractor)
	miMock.On("ResetWithConfig", domain.DefaultMendikotConfig()).Return(mockOutput)
	miMock.On("ResetWithConfig", domain.MendikotConfig{Target: 7}).Return(mockOutput)
	miMock.On("NextHand").Return(mockOutput)
	miMock.On("GiveUp").Return(mockOutput)
	miMock.On("Hint").Return(mockOutput)
	miMock.On("ActionLog").Return(mockOutput)
	miMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewMendikotWebController(func() uc.MendikotInteractorIF { return miMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.MendikotWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustMendikotOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with target", `{"command":"reset","sessionId":"s1","config":{"target":7}}`},
		{"play p", `{"command":"p","sessionId":"s1","cardIndex":4}`},
		{"play long", `{"command":"play","sessionId":"s1","cardIndex":4}`},
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

	// **切り札を選ぶコマンドは無い。** 受け付けたらフォロー不能で決まる規則と二重になる。
	t.Run("no trump command", func(t *testing.T) {
		exec(t, `{"command":"t","sessionId":"s1","suit":3}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestMendikotWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultMendikotConfig().Target
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrMdk(0), def},
		{"above the maximum", intPtrMdk(999), def},
		{"in range is kept", intPtrMdk(7), 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := (&controller.MendikotWebConfig{Target: tc.in}).ToConfig().Target; got != tc.want {
				t.Fatalf("Target = %d, want %d", got, tc.want)
			}
		})
	}
}
