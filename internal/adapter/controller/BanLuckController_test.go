//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// **既定の出力も配列で返る。**
func TestBanLuckWebController_DefaultOutputHasArrays(t *testing.T) {
	out := &controller.BanLuckWebOutput{Seats: make([]*controller.BanLuckWebOutputSeat, 0)}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := string(raw["seats"]); got != "[]" {
		t.Errorf("seats = %s, want []", got)
	}
}

// **席の札も配列で返る。** 配る前の席が null になるとページが落ちる。
func TestBanLuckWebController_SeatCardsAreNeverNull(t *testing.T) {
	b, err := json.Marshal(&controller.BanLuckWebOutputSeat{
		Cards: make([]*controller.WebOutputCard, 0),
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := string(raw["cards"]); got != "[]" {
		t.Errorf("cards = %s, want []", got)
	}
}

func newBLMock() *usecase.MockBanLuckInteractor {
	m := new(usecase.MockBanLuckInteractor)
	out := `{"phase":0,"seats":[],"message":""}`
	m.On("Reset").Return(out)
	m.On("PlaceBet", 50).Return(out)
	m.On("PlaceBet", 0).Return(out)
	m.On("Hit").Return(out)
	m.On("Stand").Return(out)
	m.On("NextRound").Return(out)
	m.On("Hint").Return(out)
	m.On("ActionLog").Return(out)
	return m
}

func TestBanLuckWebController_Method(t *testing.T) {
	m := newBLMock()
	ctrl := controller.NewBanLuckWebController(func() uc.BanLuckInteractorIF { return m })
	defer ctrl.Stop()

	for _, tc := range []struct{ name, body string }{
		{"reset", `{"command":"reset","sessionId":"s1"}`},
		{"bet", `{"command":"bet","bet":50,"sessionId":"s2"}`},
		{"bet alias", `{"command":"b","bet":50,"sessionId":"s3"}`},
		{"bet zero", `{"command":"bet","bet":0,"sessionId":"s4"}`},
		{"hit", `{"command":"hit","sessionId":"s5"}`},
		{"stand", `{"command":"stand","sessionId":"s6"}`},
		{"next", `{"command":"next","sessionId":"s7"}`},
		{"hint", `{"command":"hint","sessionId":"s8"}`},
		{"log", `{"command":"log","sessionId":"s9"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.BanLuckWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.BanLuckWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s99"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(recorded.Body.String(), "Unsupported command") {
			t.Errorf("expected Unsupported command, got: %s", recorded.Body.String())
		}
	})
}

// **賭け金 0 は「親なので賭けない」で、送らなかったのとは違う。**
//
// 0 を「省略」と同じに扱うと、親のラウンドの普通のリクエストが全部 400 になる。
func TestBanLuckWebController_ZeroBetIsAValue(t *testing.T) {
	m := newBLMock()
	ctrl := controller.NewBanLuckWebController(func() uc.BanLuckInteractorIF { return m })
	defer ctrl.Stop()

	var zero controller.BanLuckWebInput
	if err := json.Unmarshal([]byte(`{"command":"bet","bet":0,"sessionId":"z1"}`), &zero); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &zero).CodeIs(http.StatusOK)
	m.AssertCalled(t, "PlaceBet", 0)

	var missing controller.BanLuckWebInput
	if err := json.Unmarshal([]byte(`{"command":"bet","sessionId":"z2"}`), &missing); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	recorded := execRequest(t, ctrl.Exec, &missing)
	recorded.CodeIs(http.StatusBadRequest)
	if !strings.Contains(recorded.Body.String(), "bet is required") {
		t.Errorf("expected 'bet is required', got: %s", recorded.Body.String())
	}
	m.AssertNumberOfCalls(t, "PlaceBet", 1)
}

// **親の義務はコントローラで判定しない。** 可否はドメインだけが知っている。
func TestBanLuckWebController_MustHitIsNotGatedHere(t *testing.T) {
	m := new(usecase.MockBanLuckInteractor)
	m.On("Stand").Return(`{"message":"the banker must hit below the minimum"}`)
	ctrl := controller.NewBanLuckWebController(func() uc.BanLuckInteractorIF { return m })
	defer ctrl.Stop()

	var input controller.BanLuckWebInput
	if err := json.Unmarshal([]byte(`{"command":"stand","sessionId":"g1"}`), &input); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
	m.AssertCalled(t, "Stand")
}

// --- CUI ---

func TestBanLuckCuiController(t *testing.T) {
	m := newBLMock()
	c := controller.NewBanLuckCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.NotEmpty(t, c.Exec("r"))

	assert.NotEmpty(t, c.Exec("bet 50"))
	m.AssertCalled(t, "PlaceBet", 50)
	// **親のラウンドの 0 が通ること。**
	assert.NotEmpty(t, c.Exec("bet 0"))
	m.AssertCalled(t, "PlaceBet", 0)
	assert.Contains(t, c.Exec("bet"), "required")
	assert.Contains(t, c.Exec("bet xyz"), "Invalid")

	for _, cmd := range []string{"hit", "h", "stand", "s", "next", "hint", "log"} {
		assert.NotEmpty(t, c.Exec(cmd), "command %s produced nothing", cmd)
	}
	m.AssertNumberOfCalls(t, "Hit", 2)
	m.AssertNumberOfCalls(t, "Stand", 2)
	assert.NotEmpty(t, c.Exec("zzz"))
}
