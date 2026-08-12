//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// **既定の出力も配列で返る。** null を返すと TS 側の非 optional な配列が破れる。
func TestChemindeFerWebController_DefaultOutputHasArrays(t *testing.T) {
	out := &controller.ChemindeFerWebOutput{
		Players:    make([]*controller.ChemindeFerWebOutputPlayer, 0),
		BankerHand: make([]*controller.WebOutputCard, 0),
		PunterHand: make([]*controller.WebOutputCard, 0),
	}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"players", "bankerHand", "punterHand"} {
		if got := string(raw[key]); got != "[]" {
			t.Errorf("%s = %s, want []", key, got)
		}
	}
}

func TestChemindeFerWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"bankerIdx":0,"betTurn":-1,"stake":0,` +
		`"bankerHand":[],"punterHand":[],"roundNumber":1,"gameEndFlag":false,"message":""}`

	m := new(usecase.MockChemindeFerInteractor)
	m.On("Reset").Return(mockOutput)
	m.On("SetStake", 200).Return(mockOutput)
	m.On("PlaceBet", 0, 50).Return(mockOutput)
	m.On("PlaceBet", 0, 0).Return(mockOutput)
	m.On("PunterDraw").Return(mockOutput)
	m.On("PunterStand").Return(mockOutput)
	m.On("BankerDraw").Return(mockOutput)
	m.On("BankerStand").Return(mockOutput)
	m.On("PassBank").Return(mockOutput)
	m.On("NextRound").Return(mockOutput)
	m.On("GiveUp").Return(mockOutput)
	m.On("Hint").Return(mockOutput)
	m.On("ActionLog").Return(mockOutput)

	factory := func() uc.ChemindeFerInteractorIF { return m }
	ctrl := controller.NewChemindeFerWebController(factory)
	defer ctrl.Stop()

	cases := []struct{ name, body string }{
		{"reset", `{"command":"reset","sessionId":"s1"}`},
		{"stake", `{"command":"stake","stake":200,"sessionId":"s2"}`},
		{"bet", `{"command":"bet","amount":50,"sessionId":"s3"}`},
		{"bet pass", `{"command":"bet","amount":0,"sessionId":"s4"}`},
		{"punter draw", `{"command":"pd","sessionId":"s5"}`},
		{"punter stand", `{"command":"ps","sessionId":"s6"}`},
		{"banker draw", `{"command":"bd","sessionId":"s7"}`},
		{"banker stand", `{"command":"bs","sessionId":"s8"}`},
		{"pass bank", `{"command":"pb","sessionId":"s9"}`},
		{"next", `{"command":"next","sessionId":"s10"}`},
		{"giveup", `{"command":"giveup","sessionId":"s11"}`},
		{"hint", `{"command":"hint","sessionId":"s12"}`},
		{"log", `{"command":"log","sessionId":"s13"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.ChemindeFerWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.ChemindeFerWebInput
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

// **賭け 0 は「降りる」で、送らなかったのとは違う。**
//
// omitempty な int にすると両方 0 に潰れて、降りたつもりが「金額が無い」になる。
func TestChemindeFerWebController_ZeroBetIsAValue(t *testing.T) {
	m := new(usecase.MockChemindeFerInteractor)
	m.On("PlaceBet", 0, 0).Return(`{}`)

	factory := func() uc.ChemindeFerInteractorIF { return m }
	ctrl := controller.NewChemindeFerWebController(factory)
	defer ctrl.Stop()

	var pass controller.ChemindeFerWebInput
	if err := json.Unmarshal([]byte(`{"command":"bet","amount":0,"sessionId":"b1"}`), &pass); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &pass).CodeIs(http.StatusOK)
	m.AssertCalled(t, "PlaceBet", 0, 0)

	var missing controller.ChemindeFerWebInput
	if err := json.Unmarshal([]byte(`{"command":"bet","sessionId":"b2"}`), &missing); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	recorded := execRequest(t, ctrl.Exec, &missing)
	recorded.CodeIs(http.StatusBadRequest)
	if !strings.Contains(recorded.Body.String(), "amount is required") {
		t.Errorf("expected 'amount is required', got: %s", recorded.Body.String())
	}
	m.AssertNumberOfCalls(t, "PlaceBet", 1)
}

// **バンク額の省略も 400。**
func TestChemindeFerWebController_StakeWithoutAmountIsRejected(t *testing.T) {
	m := new(usecase.MockChemindeFerInteractor)
	m.On("SetStake", 100).Return(`{}`)

	factory := func() uc.ChemindeFerInteractorIF { return m }
	ctrl := controller.NewChemindeFerWebController(factory)
	defer ctrl.Stop()

	var missing controller.ChemindeFerWebInput
	if err := json.Unmarshal([]byte(`{"command":"stake","sessionId":"k1"}`), &missing); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	recorded := execRequest(t, ctrl.Exec, &missing)
	recorded.CodeIs(http.StatusBadRequest)
	if !strings.Contains(recorded.Body.String(), "stake is required") {
		t.Errorf("expected 'stake is required', got: %s", recorded.Body.String())
	}
	m.AssertNotCalled(t, "SetStake")
}

// **親の手と子の手は別コマンド。** 取り違えると相手側の判断を勝手に確定させてしまう。
func TestChemindeFerWebController_DrawCommandsAreSideSpecific(t *testing.T) {
	m := new(usecase.MockChemindeFerInteractor)
	m.On("PunterDraw").Return(`{}`)
	m.On("PunterStand").Return(`{}`)
	m.On("BankerDraw").Return(`{}`)
	m.On("BankerStand").Return(`{}`)

	factory := func() uc.ChemindeFerInteractorIF { return m }
	ctrl := controller.NewChemindeFerWebController(factory)
	defer ctrl.Stop()

	for _, tc := range []struct{ command, called, notCalled string }{
		{"pd", "PunterDraw", "BankerDraw"},
		{"ps", "PunterStand", "BankerStand"},
		{"bd", "BankerDraw", "PunterDraw"},
		{"bs", "BankerStand", "PunterStand"},
	} {
		t.Run(tc.command, func(t *testing.T) {
			fresh := new(usecase.MockChemindeFerInteractor)
			fresh.On("PunterDraw").Return(`{}`)
			fresh.On("PunterStand").Return(`{}`)
			fresh.On("BankerDraw").Return(`{}`)
			fresh.On("BankerStand").Return(`{}`)
			c := controller.NewChemindeFerWebController(func() uc.ChemindeFerInteractorIF { return fresh })
			defer c.Stop()

			var input controller.ChemindeFerWebInput
			body := `{"command":"` + tc.command + `","sessionId":"d-` + tc.command + `"}`
			if err := json.Unmarshal([]byte(body), &input); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			execRequest(t, c.Exec, &input).CodeIs(http.StatusOK)
			fresh.AssertCalled(t, tc.called)
			fresh.AssertNotCalled(t, tc.notCalled)
		})
	}
}
