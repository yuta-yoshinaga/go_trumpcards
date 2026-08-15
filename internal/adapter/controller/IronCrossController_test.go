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
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// **既定の出力も配列で返る。**
func TestIronCrossWebController_DefaultOutputHasArrays(t *testing.T) {
	out := &controller.IronCrossWebOutput{
		Seats:             make([]*controller.IronCrossWebOutputSeat, 0),
		Cross:             make([]*controller.WebOutputCard, 0),
		VerticalIndexes:   domain.IronCrossLineIndexes(domain.IronCrossLineVertical),
		HorizontalIndexes: domain.IronCrossLineIndexes(domain.IronCrossLineHorizontal),
	}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"seats", "cross", "verticalIndexes", "horizontalIndexes"} {
		if got := string(raw[key]); got == "null" {
			t.Errorf("%s = null, want an array", key)
		}
	}
}

func newIcMock() *usecase.MockIronCrossInteractor {
	m := new(usecase.MockIronCrossInteractor)
	out := `{"phase":0,"seats":[],"cross":[],"message":""}`
	m.On("Reset").Return(out)
	m.On("Action", domain.IronCrossActionFold, 0).Return(out)
	m.On("Action", domain.IronCrossActionCheck, 0).Return(out)
	m.On("Action", domain.IronCrossActionCall, 0).Return(out)
	m.On("Action", domain.IronCrossActionBet, 20).Return(out)
	m.On("Action", domain.IronCrossActionRaise, 20).Return(out)
	m.On("ChooseLine", int(domain.IronCrossLineVertical)).Return(out)
	m.On("ChooseLine", int(domain.IronCrossLineHorizontal)).Return(out)
	m.On("NextHand").Return(out)
	m.On("Hint").Return(out)
	m.On("ActionLog").Return(out)
	return m
}

// **コマンド名がドメインのアクション値に正しく対応する。**
func TestIronCrossWebController_MapsCommandsToActions(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		wantAction int
		wantAmount int
	}{
		{"fold", `{"command":"fold","sessionId":"s1"}`, domain.IronCrossActionFold, 0},
		{"fold alias", `{"command":"f","sessionId":"s2"}`, domain.IronCrossActionFold, 0},
		{"check", `{"command":"check","sessionId":"s3"}`, domain.IronCrossActionCheck, 0},
		{"check alias", `{"command":"k","sessionId":"s4"}`, domain.IronCrossActionCheck, 0},
		{"call", `{"command":"call","sessionId":"s5"}`, domain.IronCrossActionCall, 0},
		{"call alias", `{"command":"c","sessionId":"s6"}`, domain.IronCrossActionCall, 0},
		{"bet", `{"command":"bet","amount":20,"sessionId":"s7"}`, domain.IronCrossActionBet, 20},
		{"raise", `{"command":"raise","amount":20,"sessionId":"s8"}`, domain.IronCrossActionRaise, 20},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newIcMock()
			ctrl := controller.NewIronCrossWebController(func() uc.IronCrossInteractorIF { return m })
			defer ctrl.Stop()

			var input controller.IronCrossWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
			m.AssertCalled(t, "Action", tc.wantAction, tc.wantAmount)
		})
	}
}

// **縦横のコマンドがそのままの列で届く。**
//
// ここが入れ替わると、選んだのと反対の 3 枚で役が組まれて負ける ── しかも
// 画面には「選べた」としか出ないので、誰も気づけない形で静かに壊れる。
func TestIronCrossWebController_MapsLineCommands(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		wantLine   domain.IronCrossLine
	}{
		{"vertical", `{"command":"vertical","sessionId":"l1"}`, domain.IronCrossLineVertical},
		{"vertical alias", `{"command":"v","sessionId":"l2"}`, domain.IronCrossLineVertical},
		{"horizontal", `{"command":"horizontal","sessionId":"l3"}`, domain.IronCrossLineHorizontal},
		{"horizontal alias", `{"command":"h","sessionId":"l4"}`, domain.IronCrossLineHorizontal},
		{"explicit vertical", `{"command":"line","line":1,"sessionId":"l5"}`, domain.IronCrossLineVertical},
		{"explicit horizontal", `{"command":"line","line":2,"sessionId":"l6"}`, domain.IronCrossLineHorizontal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newIcMock()
			ctrl := controller.NewIronCrossWebController(func() uc.IronCrossInteractorIF { return m })
			defer ctrl.Stop()

			var input controller.IronCrossWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
			m.AssertCalled(t, "ChooseLine", int(tc.wantLine))
		})
	}
}

// **列の未送信を「0 番の列」にしない。** 0 は「まだ選んでいない」を意味する
// 実在の値なので、省略と同一視すると選択が黙って消える。
func TestIronCrossWebController_LineIsRequired(t *testing.T) {
	m := newIcMock()
	ctrl := controller.NewIronCrossWebController(func() uc.IronCrossInteractorIF { return m })
	defer ctrl.Stop()

	var input controller.IronCrossWebInput
	if err := json.Unmarshal([]byte(`{"command":"line","sessionId":"l9"}`), &input); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusBadRequest)
	assert.Contains(t, recorded.Body.String(), "line is required")
	m.AssertNotCalled(t, "ChooseLine", int(domain.IronCrossLineNone))
}

func TestIronCrossWebController_Method(t *testing.T) {
	m := newIcMock()
	ctrl := controller.NewIronCrossWebController(func() uc.IronCrossInteractorIF { return m })
	defer ctrl.Stop()

	for _, tc := range []struct{ name, body string }{
		{"reset", `{"command":"reset","sessionId":"m1"}`},
		{"next", `{"command":"next","sessionId":"m2"}`},
		{"hint", `{"command":"hint","sessionId":"m3"}`},
		{"log", `{"command":"log","sessionId":"m4"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.IronCrossWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.IronCrossWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"m9"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(recorded.Body.String(), "Unsupported command") {
			t.Errorf("expected Unsupported command, got: %s", recorded.Body.String())
		}
	})
}

// **額が要る手では未送信を弾く。** 額の要らない手では送られても無視する。
func TestIronCrossWebController_AmountIsRequiredOnlyWhereItMatters(t *testing.T) {
	m := newIcMock()
	ctrl := controller.NewIronCrossWebController(func() uc.IronCrossInteractorIF { return m })
	defer ctrl.Stop()

	for _, cmd := range []string{"bet", "raise"} {
		var input controller.IronCrossWebInput
		if err := json.Unmarshal([]byte(`{"command":"`+cmd+`","sessionId":"a1"}`), &input); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(recorded.Body.String(), "amount is required") {
			t.Errorf("%s: expected 'amount is required', got: %s", cmd, recorded.Body.String())
		}
	}

	for _, cmd := range []string{"fold", "check", "call"} {
		var input controller.IronCrossWebInput
		if err := json.Unmarshal([]byte(`{"command":"`+cmd+`","sessionId":"a2"}`), &input); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
	}
}

// --- CUI ---

func TestIronCrossCuiController(t *testing.T) {
	m := newIcMock()
	c := controller.NewIronCrossCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.NotEmpty(t, c.Exec("r"))

	assert.NotEmpty(t, c.Exec("fold"))
	m.AssertCalled(t, "Action", domain.IronCrossActionFold, 0)
	assert.NotEmpty(t, c.Exec("k"))
	m.AssertCalled(t, "Action", domain.IronCrossActionCheck, 0)
	assert.NotEmpty(t, c.Exec("call"))
	m.AssertCalled(t, "Action", domain.IronCrossActionCall, 0)

	assert.NotEmpty(t, c.Exec("bet 20"))
	m.AssertCalled(t, "Action", domain.IronCrossActionBet, 20)
	assert.NotEmpty(t, c.Exec("raise 20"))
	m.AssertCalled(t, "Action", domain.IronCrossActionRaise, 20)

	assert.NotEmpty(t, c.Exec("vertical"))
	m.AssertCalled(t, "ChooseLine", int(domain.IronCrossLineVertical))
	assert.NotEmpty(t, c.Exec("h"))
	m.AssertCalled(t, "ChooseLine", int(domain.IronCrossLineHorizontal))

	assert.Contains(t, c.Exec("bet"), msgAmountRequired())
	assert.Contains(t, c.Exec("bet xyz"), msgInvalidAmountPrefix())

	for _, cmd := range []string{"next", "hint", "log"} {
		assert.NotEmpty(t, c.Exec(cmd), "command %s produced nothing", cmd)
	}
	assert.NotEmpty(t, c.Exec("zzz"))
}
