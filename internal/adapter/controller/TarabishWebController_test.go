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

func intPtrTar(v int) *int { return &v }

func mustTarabishOutputJSON(msg string) string {
	out := &controller.TarabishWebOutput{
		Players:       []*controller.TarabishWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		Scores:        []int{},
		RoundPoints:   []int{},
		TrumpTakerIdx: -1,
		WinnerTeam:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTarabishOutputJSON: %v", err))
	}
	return string(b)
}

func TestTarabishWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	tiMock := new(usecase.MockTarabishInteractor)
	tiMock.On("ResetWithConfig", domain.DefaultTarabishConfig()).Return(mockOutput)
	tiMock.On("ResetWithConfig", domain.TarabishConfig{Target: 300}).Return(mockOutput)
	tiMock.On("TakeTrump").Return(mockOutput)
	tiMock.On("PassTrump").Return(mockOutput)
	tiMock.On("NextRound").Return(mockOutput)
	tiMock.On("GiveUp").Return(mockOutput)
	tiMock.On("Hint").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)
	tiMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewTarabishWebController(func() uc.TarabishInteractorIF { return tiMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.TarabishWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustTarabishOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with target", `{"command":"reset","sessionId":"s1","config":{"target":300}}`},
		{"take t", `{"command":"t","sessionId":"s1"}`},
		{"take long", `{"command":"take","sessionId":"s1"}`},
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

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestTarabishWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultTarabishConfig().Target
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrTar(10), def},
		{"above the maximum", intPtrTar(9999), def},
		{"in range is kept", intPtrTar(300), 300},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := (&controller.TarabishWebConfig{Target: tc.in}).ToConfig().Target; got != tc.want {
				t.Fatalf("Target = %d, want %d", got, tc.want)
			}
		})
	}
}
