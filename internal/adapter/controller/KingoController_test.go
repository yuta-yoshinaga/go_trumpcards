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
func TestKingoWebController_DefaultOutputHasArrays(t *testing.T) {
	out := &controller.KingoWebOutput{Seats: make([]*controller.KingoWebOutputSeat, 0)}
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

func newKingoMock() *usecase.MockKingoInteractor {
	m := new(usecase.MockKingoInteractor)
	out := `{"phase":0,"seats":[],"message":""}`
	m.On("Reset").Return(out)
	m.On("Bet", 20).Return(out)
	m.On("Bet", 50).Return(out)
	m.On("Deal").Return(out)
	m.On("NextRound").Return(out)
	m.On("Hint").Return(out)
	m.On("ActionLog").Return(out)
	return m
}

// **張りと配るは別のコマンド。** 親と子で求められる手が違う。
func TestKingoWebController_MapsCommands(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		assertFn   func(*testing.T, *usecase.MockKingoInteractor)
	}{
		{
			"bet", `{"command":"bet","amount":20,"sessionId":"s1"}`,
			func(t *testing.T, m *usecase.MockKingoInteractor) { m.AssertCalled(t, "Bet", 20) },
		},
		{
			"bet alias", `{"command":"b","amount":50,"sessionId":"s2"}`,
			func(t *testing.T, m *usecase.MockKingoInteractor) { m.AssertCalled(t, "Bet", 50) },
		},
		{
			"deal", `{"command":"deal","sessionId":"s3"}`,
			func(t *testing.T, m *usecase.MockKingoInteractor) { m.AssertCalled(t, "Deal") },
		},
		{
			"deal alias", `{"command":"d","sessionId":"s4"}`,
			func(t *testing.T, m *usecase.MockKingoInteractor) { m.AssertCalled(t, "Deal") },
		},
		{
			"next", `{"command":"next","sessionId":"s5"}`,
			func(t *testing.T, m *usecase.MockKingoInteractor) { m.AssertCalled(t, "NextRound") },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newKingoMock()
			ctrl := controller.NewKingoWebController(func() uc.KingoInteractorIF { return m })
			defer ctrl.Stop()

			var input controller.KingoWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
			tc.assertFn(t, m)
		})
	}
}

// **配るコマンドが張りに化けない。** 逆向きも見る。
func TestKingoWebController_DealDoesNotBet(t *testing.T) {
	m := newKingoMock()
	ctrl := controller.NewKingoWebController(func() uc.KingoInteractorIF { return m })
	defer ctrl.Stop()

	var input controller.KingoWebInput
	_ = json.Unmarshal([]byte(`{"command":"deal","sessionId":"d9"}`), &input)
	execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
	m.AssertNotCalled(t, "Bet", 20)
	m.AssertNotCalled(t, "Bet", 50)
}

// **額は必須。** 0 を「省略」と同一視しない。
func TestKingoWebController_AmountIsRequiredForBet(t *testing.T) {
	m := newKingoMock()
	ctrl := controller.NewKingoWebController(func() uc.KingoInteractorIF { return m })
	defer ctrl.Stop()

	var input controller.KingoWebInput
	if err := json.Unmarshal([]byte(`{"command":"bet","sessionId":"a1"}`), &input); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusBadRequest)
	assert.Contains(t, recorded.Body.String(), "amount is required")

	// 配るのに額は要らない。
	var deal controller.KingoWebInput
	_ = json.Unmarshal([]byte(`{"command":"deal","sessionId":"a2"}`), &deal)
	execRequest(t, ctrl.Exec, &deal).CodeIs(http.StatusOK)
}

func TestKingoWebController_Method(t *testing.T) {
	m := newKingoMock()
	ctrl := controller.NewKingoWebController(func() uc.KingoInteractorIF { return m })
	defer ctrl.Stop()

	for _, tc := range []struct{ name, body string }{
		{"reset", `{"command":"reset","sessionId":"m1"}`},
		{"hint", `{"command":"hint","sessionId":"m2"}`},
		{"log", `{"command":"log","sessionId":"m3"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.KingoWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		var input controller.KingoWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"m9"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		if !strings.Contains(recorded.Body.String(), "Unsupported command") {
			t.Errorf("expected Unsupported command, got: %s", recorded.Body.String())
		}
	})
}

// **配当と枚数はサーバが載せる。** 画面に書き写させない。
func TestKingoWebController_DefaultOutputCarriesTheRules(t *testing.T) {
	m := newKingoMock()
	ctrl := controller.NewKingoWebController(func() uc.KingoInteractorIF { return m })
	defer ctrl.Stop()

	var input controller.KingoWebInput
	if err := json.Unmarshal([]byte(`{"command":"bet","sessionId":"r1"}`), &input); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusBadRequest)

	var got controller.KingoWebOutput
	if err := json.Unmarshal(recorded.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	assert.Equal(t, domain.KingoHandSize, got.HandSize)
	assert.Equal(t, domain.KingoPayout(domain.KingoRankArashi), got.PayoutArashi)
	assert.Equal(t, domain.KingoPayout(domain.KingoRankPair), got.PayoutPair)
	assert.Greater(t, got.PayoutArashi, got.PayoutPair, "嵐の配当が上でない")
}

// --- CUI ---

func TestKingoCuiController(t *testing.T) {
	m := newKingoMock()
	c := controller.NewKingoCuiController(m)

	assert.Equal(t, "bye.", c.Exec("q"))
	assert.NotEmpty(t, c.Exec("r"))

	assert.NotEmpty(t, c.Exec("bet 20"))
	m.AssertCalled(t, "Bet", 20)
	assert.NotEmpty(t, c.Exec("b 50"))
	m.AssertCalled(t, "Bet", 50)

	assert.NotEmpty(t, c.Exec("deal"))
	m.AssertCalled(t, "Deal")
	assert.NotEmpty(t, c.Exec("d"))

	assert.Contains(t, c.Exec("bet"), "required")
	assert.Contains(t, c.Exec("bet xyz"), "Invalid")

	for _, cmd := range []string{"next", "hint", "log"} {
		assert.NotEmpty(t, c.Exec(cmd), "command %s produced nothing", cmd)
	}
	assert.NotEmpty(t, c.Exec("zzz"))
}
