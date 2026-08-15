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

// **既定の出力も配列で返り、賭ける前の pick は -1。**
func TestMonteBankWebController_DefaultOutput(t *testing.T) {
	out := &controller.MonteBankWebOutput{
		Layout: make([]*controller.MonteBankWebOutputCard, 0),
		Pick:   -1,
	}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := string(raw["layout"]); got != "[]" {
		t.Errorf("layout = %s, want []", got)
	}
	if got := string(raw["pick"]); got != "-1" {
		t.Errorf("pick = %s, want -1", got)
	}
	// **賭ける前のゲートは省略される。** null の札をページに渡さない。
	if _, ok := raw["gate"]; ok {
		t.Errorf("gate should be omitted before the bet, got %s", raw["gate"])
	}
}

func newMBMock() *usecase.MockMonteBankInteractor {
	m := new(usecase.MockMonteBankInteractor)
	out := `{"phase":0,"layout":[],"pick":-1,"message":""}`
	m.On("Reset").Return(out)
	m.On("PlaceBet", 0, 50).Return(out)
	m.On("PlaceBet", 3, 50).Return(out)
	m.On("NextRound").Return(out)
	m.On("Hint").Return(out)
	m.On("ActionLog").Return(out)
	return m
}

func TestMonteBankWebController_Method(t *testing.T) {
	m := newMBMock()
	ctrl := controller.NewMonteBankWebController(func() uc.MonteBankInteractorIF { return m })
	defer ctrl.Stop()

	for _, tc := range []struct{ name, body string }{
		{"reset", `{"command":"reset","sessionId":"s1"}`},
		{"bet", `{"command":"bet","idx":0,"bet":50,"sessionId":"s2"}`},
		{"bet alias", `{"command":"b","idx":3,"bet":50,"sessionId":"s3"}`},
		{"next", `{"command":"next","sessionId":"s4"}`},
		{"hint", `{"command":"hint","sessionId":"s5"}`},
		{"log", `{"command":"log","sessionId":"s6"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.MonteBankWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.MonteBankWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s99"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(recorded.Body.String(), "Unsupported command") {
			t.Errorf("expected Unsupported command, got: %s", recorded.Body.String())
		}
	})
}

// **場札 0 に賭けるのは普通の操作。**
//
// 0 を「省略」と同一視すると、いちばん左の札を選んだリクエストが全部 400 になる。
func TestMonteBankWebController_ZeroIndexIsAValue(t *testing.T) {
	m := newMBMock()
	ctrl := controller.NewMonteBankWebController(func() uc.MonteBankInteractorIF { return m })
	defer ctrl.Stop()

	var zero controller.MonteBankWebInput
	if err := json.Unmarshal([]byte(`{"command":"bet","idx":0,"bet":50,"sessionId":"z1"}`), &zero); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	execRequest(t, ctrl.Exec, &zero).CodeIs(http.StatusOK)
	m.AssertCalled(t, "PlaceBet", 0, 50)

	for _, tc := range []struct{ name, body, want string }{
		{"idx missing", `{"command":"bet","bet":50,"sessionId":"z2"}`, "idx is required"},
		{"bet missing", `{"command":"bet","idx":0,"sessionId":"z3"}`, "bet is required"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.MonteBankWebInput
			if err := json.Unmarshal([]byte(tc.body), &input); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusBadRequest)
			if !strings.Contains(recorded.Body.String(), tc.want) {
				t.Errorf("expected %q, got: %s", tc.want, recorded.Body.String())
			}
		})
	}
	m.AssertNumberOfCalls(t, "PlaceBet", 1)
}

// --- CUI ---

func TestMonteBankCuiController(t *testing.T) {
	m := newMBMock()
	c := controller.NewMonteBankCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.NotEmpty(t, c.Exec("r"))

	// **画面は 1 始まり、ドメインは 0 始まり。** ここで 1 度だけ引く。
	assert.NotEmpty(t, c.Exec("bet 1 50"))
	m.AssertCalled(t, "PlaceBet", 0, 50)
	assert.NotEmpty(t, c.Exec("bet 4 50"))
	m.AssertCalled(t, "PlaceBet", 3, 50)

	// 0 は画面上の番号として無効 (1〜4)。
	assert.True(t, msgRejected(c.Exec("bet 0 50")))
	assert.True(t, msgRejected(c.Exec("bet")))
	assert.True(t, msgRejected(c.Exec("bet 1")))
	assert.True(t, msgRejected(c.Exec("bet xyz 50")))
	assert.True(t, msgRejected(c.Exec("bet 1 xyz")))

	for _, cmd := range []string{"next", "hint", "log"} {
		assert.NotEmpty(t, c.Exec(cmd), "command %s produced nothing", cmd)
	}
	assert.NotEmpty(t, c.Exec("zzz"))
}
