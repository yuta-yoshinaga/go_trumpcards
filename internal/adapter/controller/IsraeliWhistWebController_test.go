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

func intPtrIw(v int) *int { return &v }

func mustIsraeliWhistOutputJSON(msg string) string {
	out := &controller.IsraeliWhistWebOutput{
		Players:       []*controller.IsraeliWhistWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		DeclarerIdx:   -1,
		RestrictedBid: -1,
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustIsraeliWhistOutputJSON: %v", err))
	}
	return string(b)
}

func TestIsraeliWhistWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	wiMock := new(usecase.MockIsraeliWhistInteractor)
	wiMock.On("ResetWithConfig", domain.DefaultIsraeliWhistConfig()).Return(mockOutput)
	wiMock.On("ResetWithConfig", domain.IsraeliWhistConfig{Rounds: 6}).Return(mockOutput)
	wiMock.On("AuctionBid", 7, 3).Return(mockOutput)
	wiMock.On("AuctionPass").Return(mockOutput)
	wiMock.On("Bid", 4).Return(mockOutput)
	wiMock.On("Bid", 0).Return(mockOutput)
	wiMock.On("NextRound").Return(mockOutput)
	wiMock.On("GiveUp").Return(mockOutput)
	wiMock.On("Hint").Return(mockOutput)
	wiMock.On("ActionLog").Return(mockOutput)
	wiMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewIsraeliWhistWebController(func() uc.IsraeliWhistInteractorIF { return wiMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.IsraeliWhistWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustIsraeliWhistOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with rounds", `{"command":"reset","sessionId":"s1","config":{"rounds":6}}`},
		{"auction a", `{"command":"a","sessionId":"s1","bid":7,"suit":3}`},
		{"auction long", `{"command":"auction","sessionId":"s1","bid":7,"suit":3}`},
		{"pass", `{"command":"pass","sessionId":"s1"}`},
		{"bid b", `{"command":"b","sessionId":"s1","bid":4}`},
		// **0 は「取らない」という宣言。** 省略と区別できること。
		{"bid zero", `{"command":"bid","sessionId":"s1","bid":0}`},
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

	// **入札は数とスートの両方が要る。** 片方でも欠けたら通さない。
	t.Run("auction missing suit", func(t *testing.T) {
		exec(t, `{"command":"a","sessionId":"s1","bid":7}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("auction missing bid", func(t *testing.T) {
		exec(t, `{"command":"a","sessionId":"s1","suit":3}`).CodeIs(http.StatusBadRequest)
	})

	// **省略された bid を 0 と読まない。** 読むと勝手に 0 を宣言したことになる。
	t.Run("bid missing bid", func(t *testing.T) {
		exec(t, `{"command":"b","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestIsraeliWhistWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultIsraeliWhistConfig().Rounds
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrIw(0), def},
		{"above the maximum", intPtrIw(999), def},
		{"in range is kept", intPtrIw(6), 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := (&controller.IsraeliWhistWebConfig{Rounds: tc.in}).ToConfig().Rounds; got != tc.want {
				t.Fatalf("Rounds = %d, want %d", got, tc.want)
			}
		})
	}
}
