//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// **既定の出力も配列で返る。**
func TestRikkenWebController_DefaultOutputHasArrays(t *testing.T) {
	out := &controller.RikkenWebOutput{
		Players:      make([]*controller.RikkenWebOutputPlayer, 0),
		ValidPlays:   make([]int, 0),
		CurrentTrick: make([]*controller.WebOutputTrickCard, 0),
		LastTrick:    make([]*controller.WebOutputTrickCard, 0),
	}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"players", "validPlays", "currentTrick", "lastTrick"} {
		if got := string(raw[key]); got != "[]" {
			t.Errorf("%s = %s, want []", key, got)
		}
	}
}

func TestRikkenWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"validPlays":[],"dealerIdx":0,"contract":0,` +
		`"declarerIdx":-1,"partnerIdx":-1,"trumpSuit":-1,"currentTurn":0,"isHumanTurn":true,` +
		`"currentTrick":[],"lastTrick":[],"lastTrickWinner":-1,"trickCount":0,"declarerTricks":0,` +
		`"roundNumber":1,"gameEndFlag":false,"winnerIdx":-1,"message":""}`

	m := new(usecase.MockRikkenInteractor)
	m.On("Reset").Return(mockOutput)
	m.On("PlayCard", 4).Return(mockOutput)
	m.On("Bid", domain.RikkenContractNone).Return(mockOutput)
	m.On("Bid", domain.RikkenContractSolo).Return(mockOutput)
	m.On("Call", domain.CardDesignHeart).Return(mockOutput)
	m.On("NextRound").Return(mockOutput)
	m.On("GiveUp").Return(mockOutput)
	m.On("Hint").Return(mockOutput)
	m.On("ActionLog").Return(mockOutput)

	factory := func() uc.RikkenInteractorIF { return m }
	ctrl := controller.NewRikkenWebController(factory)
	defer ctrl.Stop()

	cases := []struct{ name, body string }{
		{"reset", `{"command":"reset","sessionId":"s1"}`},
		{"play", `{"command":"play","cardIndex":4,"sessionId":"s2"}`},
		{"bid solo", `{"command":"bid","contract":3,"sessionId":"s3"}`},
		{"bid pass", `{"command":"bid","contract":0,"sessionId":"s4"}`},
		{"call", `{"command":"call","suit":3,"sessionId":"s5"}`},
		{"next", `{"command":"next","sessionId":"s6"}`},
		{"giveup", `{"command":"giveup","sessionId":"s7"}`},
		{"hint", `{"command":"hint","sessionId":"s8"}`},
		{"log", `{"command":"log","sessionId":"s9"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.RikkenWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.RikkenWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s99"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(recorded.Body.String(), "Unsupported command") {
			t.Errorf("expected Unsupported command, got: %s", recorded.Body.String())
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		execRequest(t, ctrl.Exec, strings.NewReader("{invalid")).CodeIs(http.StatusBadRequest)
	})
}

// **パス (契約 0) が「送らなかった」に化けない。**
//
// contract をポインタで受けているのはこのためです。値型だと 0 が既定値と
// 区別できず、パスするたびに 400 になるか、逆に省略がパスとして通ってしまいます。
func TestRikkenWebController_PassIsAValue(t *testing.T) {
	m := new(usecase.MockRikkenInteractor)
	m.On("Bid", domain.RikkenContractNone).Return(`{}`)

	factory := func() uc.RikkenInteractorIF { return m }
	ctrl := controller.NewRikkenWebController(factory)
	defer ctrl.Stop()

	var pass controller.RikkenWebInput
	if err := json.Unmarshal([]byte(`{"command":"bid","contract":0,"sessionId":"p1"}`), &pass); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &pass).CodeIs(http.StatusOK)
	m.AssertCalled(t, "Bid", domain.RikkenContractNone)

	// 省略は 400。
	var missing controller.RikkenWebInput
	if err := json.Unmarshal([]byte(`{"command":"bid","sessionId":"p2"}`), &missing); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	recorded := execRequest(t, ctrl.Exec, &missing)
	recorded.CodeIs(http.StatusBadRequest)
	if !strings.Contains(recorded.Body.String(), "contract is required") {
		t.Errorf("expected 'contract is required', got: %s", recorded.Body.String())
	}
	m.AssertNumberOfCalls(t, "Bid", 1)
}

// **切り札の省略も 400。** スートは 1..4 なので 0 は無効値です。
func TestRikkenWebController_CallWithoutSuitIsRejected(t *testing.T) {
	m := new(usecase.MockRikkenInteractor)
	m.On("Call", domain.CardDesignSpade).Return(`{}`)

	factory := func() uc.RikkenInteractorIF { return m }
	ctrl := controller.NewRikkenWebController(factory)
	defer ctrl.Stop()

	var missing controller.RikkenWebInput
	if err := json.Unmarshal([]byte(`{"command":"call","sessionId":"c1"}`), &missing); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	recorded := execRequest(t, ctrl.Exec, &missing)
	recorded.CodeIs(http.StatusBadRequest)
	if !strings.Contains(recorded.Body.String(), "suit is required") {
		t.Errorf("expected 'suit is required', got: %s", recorded.Body.String())
	}
	m.AssertNotCalled(t, "Call", 0)
}
