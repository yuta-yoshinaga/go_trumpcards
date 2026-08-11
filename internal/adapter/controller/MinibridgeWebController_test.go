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

func intPtrMb(v int) *int { return &v }

func mustMinibridgeOutputJSON(msg string) string {
	out := &controller.MinibridgeWebOutput{
		Players:       []*controller.MinibridgeWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		DummyHand:     []*controller.WebOutputCard{},
		TeamScores:    []int{},
		DeclarerIdx:   -1,
		DummyIdx:      -1,
		WinnerTeam:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustMinibridgeOutputJSON: %v", err))
	}
	return string(b)
}

func TestMinibridgeWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	miMock := new(usecase.MockMinibridgeInteractor)
	miMock.On("ResetWithConfig", domain.DefaultMinibridgeConfig()).Return(mockOutput)
	miMock.On("ResetWithConfig", domain.MinibridgeConfig{Rounds: 8}).Return(mockOutput)
	miMock.On("Contract", 3, 0).Return(mockOutput)
	miMock.On("Contract", 2, 4).Return(mockOutput)
	miMock.On("NextRound").Return(mockOutput)
	miMock.On("GiveUp").Return(mockOutput)
	miMock.On("Hint").Return(mockOutput)
	miMock.On("ActionLog").Return(mockOutput)
	miMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewMinibridgeWebController(func() uc.MinibridgeInteractorIF { return miMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.MinibridgeWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustMinibridgeOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with rounds", `{"command":"reset","sessionId":"s1","config":{"rounds":8}}`},
		// **ノートランプは suit:0。** 省略とは違う。
		{"contract no-trump", `{"command":"c","sessionId":"s1","level":3,"suit":0}`},
		{"contract a suit", `{"command":"contract","sessionId":"s1","level":2,"suit":4}`},
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

	t.Run("contract missing level", func(t *testing.T) {
		exec(t, `{"command":"c","sessionId":"s1","suit":3}`).CodeIs(http.StatusBadRequest)
	})

	// **スート無しの契約も通さない。** 0 埋めするとノートランプを選んでしまう。
	t.Run("contract missing suit", func(t *testing.T) {
		exec(t, `{"command":"c","sessionId":"s1","level":3}`).CodeIs(http.StatusBadRequest)
	})

	// **競りは無いので bid コマンドも無い。**
	t.Run("no bid command", func(t *testing.T) {
		exec(t, `{"command":"bid","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestMinibridgeWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultMinibridgeConfig().Rounds
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrMb(domain.MinibridgeRoundsMin - 1), def},
		{"above the maximum", intPtrMb(domain.MinibridgeRoundsMax + 1), def},
		{"the minimum is kept", intPtrMb(domain.MinibridgeRoundsMin), domain.MinibridgeRoundsMin},
		{"the maximum is kept", intPtrMb(domain.MinibridgeRoundsMax), domain.MinibridgeRoundsMax},
		// **4 の倍数でないと親が一巡しない。** エラーにせず丸める。
		{"rounded down to a multiple of four", intPtrMb(6), 4},
		{"rounded down again", intPtrMb(11), 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := (&controller.MinibridgeWebConfig{Rounds: tc.in}).ToConfig()
			if cfg.Rounds != tc.want {
				t.Fatalf("Rounds = %d, want %d", cfg.Rounds, tc.want)
			}
			// **丸めた結果はドメインに必ず受理される。**
			if err := cfg.Validate(); err != nil {
				t.Fatalf("clamped config must always validate: %v", err)
			}
		})
	}

	var input controller.MinibridgeWebInput
	if got := input.ToConfig(); got != domain.DefaultMinibridgeConfig() {
		t.Fatalf("ToConfig() = %+v, want the default", got)
	}
}
