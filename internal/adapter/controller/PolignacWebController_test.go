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

func intPtrPol(v int) *int { return &v }

func mustPolignacOutputJSON(msg string) string {
	out := &controller.PolignacWebOutput{
		Players:       []*controller.PolignacWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		CapotIdx:      -1,
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustPolignacOutputJSON: %v", err))
	}
	return string(b)
}

func TestPolignacWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	piMock := new(usecase.MockPolignacInteractor)
	piMock.On("ResetWithConfig", domain.DefaultPolignacConfig()).Return(mockOutput)
	piMock.On("ResetWithConfig", domain.PolignacConfig{Rounds: 6}).Return(mockOutput)
	piMock.On("DeclareCapot").Return(mockOutput)
	piMock.On("Pass").Return(mockOutput)
	piMock.On("NextRound").Return(mockOutput)
	piMock.On("GiveUp").Return(mockOutput)
	piMock.On("Hint").Return(mockOutput)
	piMock.On("ActionLog").Return(mockOutput)
	piMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewPolignacWebController(func() uc.PolignacInteractorIF { return piMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.PolignacWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustPolignacOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with rounds", `{"command":"reset","sessionId":"s1","config":{"rounds":6}}`},
		{"capot c", `{"command":"c","sessionId":"s1"}`},
		{"capot long", `{"command":"capot","sessionId":"s1"}`},
		{"pass", `{"command":"pass","sessionId":"s1"}`},
		{"next n", `{"command":"n","sessionId":"s1"}`},
		{"giveup g", `{"command":"g","sessionId":"s1"}`},
		{"hint h", `{"command":"h","sessionId":"s1"}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
		{"play p", `{"command":"p","sessionId":"s1","cardIndex":4}`},
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

func TestPolignacWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultPolignacConfig().Rounds
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrPol(0), def},
		{"above the maximum", intPtrPol(domain.PolignacRoundsMax + 1), def},
		{"in range is kept", intPtrPol(6), 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := (&controller.PolignacWebConfig{Rounds: tc.in}).ToConfig().Rounds; got != tc.want {
				t.Fatalf("Rounds = %d, want %d", got, tc.want)
			}
		})
	}
}
