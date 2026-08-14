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
func TestColourWhistWebController_DefaultOutputHasArrays(t *testing.T) {
	out := &controller.ColourWhistWebOutput{
		Players:      make([]*controller.ColourWhistWebOutputPlayer, 0),
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

func TestColourWhistWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"validPlays":[],"dealerIdx":0,"contract":0,` +
		`"declarerIdx":-1,"partnerIdx":-1,"trumpSuit":-1,"troelForced":false,"currentTurn":0,` +
		`"isHumanTurn":true,"currentTrick":[],"lastTrick":[],"lastTrickWinner":-1,"trickCount":0,` +
		`"declarerTricks":0,"roundNumber":1,"gameEndFlag":false,"winnerIdx":-1,"message":""}`

	m := new(usecase.MockColourWhistInteractor)
	m.On("Reset").Return(mockOutput)
	m.On("PlayCard", 3).Return(mockOutput)
	m.On("Bid", domain.ColourWhistContractNone).Return(mockOutput)
	m.On("Bid", domain.ColourWhistContractAlleen).Return(mockOutput)
	m.On("Call", domain.CardDesignHeart).Return(mockOutput)
	m.On("NextRound").Return(mockOutput)
	m.On("GiveUp").Return(mockOutput)
	m.On("Hint").Return(mockOutput)
	m.On("ActionLog").Return(mockOutput)

	factory := func() uc.ColourWhistInteractorIF { return m }
	ctrl := controller.NewColourWhistWebController(factory)
	defer ctrl.Stop()

	cases := []struct{ name, body string }{
		{"reset", `{"command":"reset","sessionId":"s1"}`},
		{"play", `{"command":"play","cardIndex":3,"sessionId":"s2"}`},
		{"bid alleen", `{"command":"bid","contract":2,"sessionId":"s3"}`},
		{"bid pass", `{"command":"bid","contract":0,"sessionId":"s4"}`},
		{"call", `{"command":"call","suit":3,"sessionId":"s5"}`},
		{"next", `{"command":"next","sessionId":"s6"}`},
		{"giveup", `{"command":"giveup","sessionId":"s7"}`},
		{"hint", `{"command":"hint","sessionId":"s8"}`},
		{"log", `{"command":"log","sessionId":"s9"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.ColourWhistWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.ColourWhistWebInput
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
func TestColourWhistWebController_PassIsAValue(t *testing.T) {
	m := new(usecase.MockColourWhistInteractor)
	m.On("Bid", domain.ColourWhistContractNone).Return(`{}`)

	factory := func() uc.ColourWhistInteractorIF { return m }
	ctrl := controller.NewColourWhistWebController(factory)
	defer ctrl.Stop()

	var pass controller.ColourWhistWebInput
	if err := json.Unmarshal([]byte(`{"command":"bid","contract":0,"sessionId":"p1"}`), &pass); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &pass).CodeIs(http.StatusOK)
	m.AssertCalled(t, "Bid", domain.ColourWhistContractNone)

	var missing controller.ColourWhistWebInput
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

// **切り札の省略も 400。**
func TestColourWhistWebController_CallWithoutSuitIsRejected(t *testing.T) {
	m := new(usecase.MockColourWhistInteractor)
	m.On("Call", domain.CardDesignSpade).Return(`{}`)

	factory := func() uc.ColourWhistInteractorIF { return m }
	ctrl := controller.NewColourWhistWebController(factory)
	defer ctrl.Stop()

	var missing controller.ColourWhistWebInput
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
