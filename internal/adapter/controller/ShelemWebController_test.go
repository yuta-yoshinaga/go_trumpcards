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

func intPtrSh(v int) *int { return &v }

func mustShelemOutputJSON(msg string) string {
	out := &controller.ShelemWebOutput{
		Players:       []*controller.ShelemWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		Scores:        []int{},
		RoundPoints:   []int{},
		TeamTricks:    []int{},
		DeclarerIdx:   -1,
		MinBid:        domain.ShelemMinBid,
		DiscardCount:  domain.ShelemWidowSize,
		WinnerTeam:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustShelemOutputJSON: %v", err))
	}
	return string(b)
}

func TestShelemWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	siMock := new(usecase.MockShelemInteractor)
	siMock.On("ResetWithConfig", domain.DefaultShelemConfig()).Return(mockOutput)
	siMock.On("ResetWithConfig", domain.ShelemConfig{Target: 700}).Return(mockOutput)
	siMock.On("Bid", 80).Return(mockOutput)
	siMock.On("BidShelem").Return(mockOutput)
	siMock.On("Pass").Return(mockOutput)
	siMock.On("Discard", []int{0, 1, 2, 3}, 3).Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("GiveUp").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)
	siMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewShelemWebController(func() uc.ShelemInteractorIF { return siMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.ShelemWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustShelemOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with target", `{"command":"reset","sessionId":"s1","config":{"target":700}}`},
		{"bid b", `{"command":"b","sessionId":"s1","bid":80}`},
		{"shelem", `{"command":"shelem","sessionId":"s1"}`},
		{"pass", `{"command":"pass","sessionId":"s1"}`},
		{"discard d", `{"command":"d","sessionId":"s1","discards":[0,1,2,3],"suit":3}`},
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

	// **入札額は既定値で埋めない。** 埋めると出していない額で落札する。
	t.Run("bid missing bid", func(t *testing.T) {
		exec(t, `{"command":"b","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	// **捨て札とスートは両方が要る。** 片方でも欠けたら通さない。
	t.Run("discard missing suit", func(t *testing.T) {
		exec(t, `{"command":"d","sessionId":"s1","discards":[0,1,2,3]}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("discard missing cards", func(t *testing.T) {
		exec(t, `{"command":"d","sessionId":"s1","suit":3}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestShelemWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultShelemConfig().Target
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrSh(10), def},
		{"above the maximum", intPtrSh(9999), def},
		{"in range is kept", intPtrSh(700), 700},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := (&controller.ShelemWebConfig{Target: tc.in}).ToConfig().Target; got != tc.want {
				t.Fatalf("Target = %d, want %d", got, tc.want)
			}
		})
	}
}
