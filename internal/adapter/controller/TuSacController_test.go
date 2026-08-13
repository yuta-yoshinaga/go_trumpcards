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
func TestTuSacWebController_DefaultOutputHasArrays(t *testing.T) {
	out := &controller.TuSacWebOutput{Seats: make([]*controller.TuSacWebOutputSeat, 0)}
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

func newTuSacMock() *usecase.MockTuSacInteractor {
	m := new(usecase.MockTuSacInteractor)
	out := `{"phase":0,"seats":[],"message":""}`
	m.On("Reset").Return(out)
	m.On("Draw", false).Return(out)
	m.On("Draw", true).Return(out)
	m.On("Meld", []int{0, 3, 6}).Return(out)
	m.On("Discard", 0).Return(out)
	m.On("Discard", 4).Return(out)
	m.On("NextRound").Return(out)
	m.On("Hint").Return(out)
	m.On("ActionLog").Return(out)
	return m
}

// **山と捨て札は別のコマンド。** 引き先を本文にすると、送り忘れが
// 「山から」に化けて、狙って拾った札が黙って流れる。
func TestTuSacWebController_DrawSourceIsInTheCommand(t *testing.T) {
	for _, tc := range []struct {
		name, body  string
		wantDiscard bool
	}{
		{"draw", `{"command":"draw","sessionId":"s1"}`, false},
		{"draw alias", `{"command":"d","sessionId":"s2"}`, false},
		{"take", `{"command":"take","sessionId":"s3"}`, true},
		{"take alias", `{"command":"t","sessionId":"s4"}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newTuSacMock()
			ctrl := controller.NewTuSacWebController(func() uc.TuSacInteractorIF { return m })
			defer ctrl.Stop()

			var input controller.TuSacWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
			m.AssertCalled(t, "Draw", tc.wantDiscard)
			m.AssertNotCalled(t, "Draw", !tc.wantDiscard)
		})
	}
}

func TestTuSacWebController_MeldAndDiscard(t *testing.T) {
	m := newTuSacMock()
	ctrl := controller.NewTuSacWebController(func() uc.TuSacInteractorIF { return m })
	defer ctrl.Stop()

	var meld controller.TuSacWebInput
	_ = json.Unmarshal([]byte(`{"command":"meld","indexes":[0,3,6],"sessionId":"m1"}`), &meld)
	execRequest(t, ctrl.Exec, &meld).CodeIs(http.StatusOK)
	m.AssertCalled(t, "Meld", []int{0, 3, 6})

	var discard controller.TuSacWebInput
	_ = json.Unmarshal([]byte(`{"command":"discard","index":4,"sessionId":"m2"}`), &discard)
	execRequest(t, ctrl.Exec, &discard).CodeIs(http.StatusOK)
	m.AssertCalled(t, "Discard", 4)
}

// **位置 0 は「省略」ではない。** 先頭の札を指す実在の値。
func TestTuSacWebController_IndexZeroIsNotOmitted(t *testing.T) {
	m := newTuSacMock()
	ctrl := controller.NewTuSacWebController(func() uc.TuSacInteractorIF { return m })
	defer ctrl.Stop()

	var input controller.TuSacWebInput
	if err := json.Unmarshal([]byte(`{"command":"discard","index":0,"sessionId":"z1"}`), &input); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
	m.AssertCalled(t, "Discard", 0)

	// 未送信は弾く。
	var missing controller.TuSacWebInput
	_ = json.Unmarshal([]byte(`{"command":"discard","sessionId":"z2"}`), &missing)
	recorded := execRequest(t, ctrl.Exec, &missing)
	recorded.CodeIs(http.StatusBadRequest)
	assert.Contains(t, recorded.Body.String(), "index is required")
}

func TestTuSacWebController_MeldNeedsIndexes(t *testing.T) {
	m := newTuSacMock()
	ctrl := controller.NewTuSacWebController(func() uc.TuSacInteractorIF { return m })
	defer ctrl.Stop()

	var input controller.TuSacWebInput
	_ = json.Unmarshal([]byte(`{"command":"meld","sessionId":"n1"}`), &input)
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusBadRequest)
	assert.Contains(t, recorded.Body.String(), "indexes are required")
}

func TestTuSacWebController_Method(t *testing.T) {
	m := newTuSacMock()
	ctrl := controller.NewTuSacWebController(func() uc.TuSacInteractorIF { return m })
	defer ctrl.Stop()

	for _, tc := range []struct{ name, body string }{
		{"reset", `{"command":"reset","sessionId":"m1"}`},
		{"next", `{"command":"next","sessionId":"m2"}`},
		{"hint", `{"command":"hint","sessionId":"m3"}`},
		{"log", `{"command":"log","sessionId":"m4"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.TuSacWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.TuSacWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"m9"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(recorded.Body.String(), "Unsupported command") {
			t.Errorf("expected Unsupported command, got: %s", recorded.Body.String())
		}
	})
}

// **札の総数と得点表はサーバが載せる。** 画面に書き写させない。
func TestTuSacWebController_DefaultOutputCarriesTheRules(t *testing.T) {
	m := newTuSacMock()
	ctrl := controller.NewTuSacWebController(func() uc.TuSacInteractorIF { return m })
	defer ctrl.Stop()

	var input controller.TuSacWebInput
	_ = json.Unmarshal([]byte(`{"command":"meld","sessionId":"r1"}`), &input)
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusBadRequest)

	var got controller.TuSacWebOutput
	if err := json.Unmarshal(recorded.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	assert.Equal(t, domain.TuSacHandSize, got.HandSize)
	assert.Equal(t, domain.TuSacDeckSize, got.DeckSize)
	assert.Equal(t, 112, got.DeckSize, "四色牌は 112 枚")
	assert.Len(t, got.MeldPointsByKind, int(domain.TuSacMeldKindMax)+1)
	assert.Greater(t, got.MeldPointsByKind[domain.TuSacMeldSoldierSet],
		got.MeldPointsByKind[domain.TuSacMeldSameColorSet], "卒 5 枚の得点が上でない")
	assert.Equal(t, -1, got.WentOutSeat, "上がりが立っている")
}

// --- CUI ---

func TestTuSacCuiController(t *testing.T) {
	m := newTuSacMock()
	c := controller.NewTuSacCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.NotEmpty(t, c.Exec("r"))

	assert.NotEmpty(t, c.Exec("draw"))
	m.AssertCalled(t, "Draw", false)
	assert.NotEmpty(t, c.Exec("take"))
	m.AssertCalled(t, "Draw", true)

	// **画面は 1 始まり、内部は 0 始まり。** ここがずれると別の札が動く。
	assert.NotEmpty(t, c.Exec("meld 1 4 7"))
	m.AssertCalled(t, "Meld", []int{0, 3, 6})
	assert.NotEmpty(t, c.Exec("discard 5"))
	m.AssertCalled(t, "Discard", 4)
	assert.NotEmpty(t, c.Exec("x 1"))
	m.AssertCalled(t, "Discard", 0)

	assert.Contains(t, c.Exec("meld"), "required")
	assert.Contains(t, c.Exec("meld abc"), "Invalid")
	assert.Contains(t, c.Exec("meld 0"), "Invalid", "0 番を受け付けている")
	assert.Contains(t, c.Exec("discard"), "required")
	assert.Contains(t, c.Exec("discard xyz"), "Invalid")

	for _, cmd := range []string{"next", "hint", "log"} {
		assert.NotEmpty(t, c.Exec(cmd), "command %s produced nothing", cmd)
	}
	assert.NotEmpty(t, c.Exec("zzz"))
}
