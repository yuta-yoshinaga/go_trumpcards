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

func intPtrBal(v int) *int { return &v }

func mustBalootOutputJSON(msg string) string {
	out := &controller.BalootWebOutput{
		Players:       []*controller.BalootWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		ValidPlays:    []int{},
		Scores:        []int{},
		RoundPoints:   []int{},
		DeclarerIdx:   -1,
		WinnerTeam:    -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustBalootOutputJSON: %v", err))
	}
	return string(b)
}

func TestBalootWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTrick":[],"validPlays":[],"message":""}`

	biMock := new(usecase.MockBalootInteractor)
	biMock.On("ResetWithConfig", domain.DefaultBalootConfig()).Return(mockOutput)
	biMock.On("ResetWithConfig", domain.BalootConfig{Target: 200}).Return(mockOutput)
	biMock.On("DeclareSun").Return(mockOutput)
	biMock.On("DeclareHokom", domain.CardDesignHeart).Return(mockOutput)
	biMock.On("PassDeclaration").Return(mockOutput)
	biMock.On("NextRound").Return(mockOutput)
	biMock.On("GiveUp").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)
	biMock.On("Play", 4).Return(mockOutput)

	ctrl := controller.NewBalootWebController(func() uc.BalootInteractorIF { return biMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.BalootWebInput
		_ = json.Unmarshal([]byte(body), &input)
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("quit q", func(t *testing.T) {
		r := exec(t, `{"command":"q","sessionId":"s1"}`)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustBalootOutputJSON("bye."))
	})

	for _, tc := range []struct{ name, body string }{
		{"reset r", `{"command":"r","sessionId":"s1"}`},
		{"reset with target", `{"command":"reset","sessionId":"s1","config":{"target":200}}`},
		{"sun", `{"command":"sun","sessionId":"s1"}`},
		{"hokom", `{"command":"hokom","sessionId":"s1","suit":3}`},
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

	// **スート無しの hokom は通さない。** 既定値で埋めると、プレイヤーが
	// 選んでいないスートが切り札になる。
	t.Run("hokom missing suit", func(t *testing.T) {
		exec(t, `{"command":"hokom","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})

	t.Run("unknown command", func(t *testing.T) {
		exec(t, `{"command":"nope","sessionId":"s1"}`).CodeIs(http.StatusBadRequest)
	})
}

func TestBalootWebConfig_ToConfigClamps(t *testing.T) {
	def := domain.DefaultBalootConfig().Target
	for _, tc := range []struct {
		name string
		in   *int
		want int
	}{
		{"nil uses the default", nil, def},
		{"below the minimum", intPtrBal(10), def},
		{"above the maximum", intPtrBal(9999), def},
		{"in range is kept", intPtrBal(200), 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := (&controller.BalootWebConfig{Target: tc.in}).ToConfig().Target; got != tc.want {
				t.Fatalf("Target = %d, want %d", got, tc.want)
			}
		})
	}
}
