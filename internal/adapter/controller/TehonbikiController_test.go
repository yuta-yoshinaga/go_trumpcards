//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

type tehonbikiControllerInteractor struct{ mock.Mock }

func (m *tehonbikiControllerInteractor) Reset() string { return m.Called().String(0) }
func (m *tehonbikiControllerInteractor) ResetWithConfig(domain.TehonbikiConfig) string {
	return m.Called().String(0)
}
func (m *tehonbikiControllerInteractor) PlaceBet(numbers []int, kind domain.TehonbikiBetType, bet int) string {
	return m.Called(numbers, kind, bet).String(0)
}
func (m *tehonbikiControllerInteractor) NextRound() string { return m.Called().String(0) }
func (m *tehonbikiControllerInteractor) GetConfig() domain.TehonbikiConfig {
	return m.Called().Get(0).(domain.TehonbikiConfig)
}
func (m *tehonbikiControllerInteractor) Hint() string      { return m.Called().String(0) }
func (m *tehonbikiControllerInteractor) ActionLog() string { return m.Called().String(0) }
func (m *tehonbikiControllerInteractor) Snapshot() ([]byte, error) {
	r := m.Called()
	return r.Get(0).([]byte), r.Error(1)
}

func TestTehonbikiWebControllerDispatchesAllWagerData(t *testing.T) {
	m := new(tehonbikiControllerInteractor)
	out := `{"phase":0,"numbers":[],"betType":"","bet":0}`
	m.On("Reset").Return(out)
	m.On("PlaceBet", []int{2, 5}, domain.TehonbikiBetDouble, 50).Return(out)
	m.On("NextRound").Return(out)
	m.On("Hint").Return(out)
	m.On("ActionLog").Return(out)
	ctrl := controller.NewTehonbikiWebController(func() usecase.TehonbikiInteractorIF { return m })
	defer ctrl.Stop()

	var input controller.TehonbikiWebInput
	requireJSON := func(body string) {
		t.Helper()
		if err := json.Unmarshal([]byte(body), &input); err != nil {
			t.Fatal(err)
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
	}
	requireJSON(`{"command":"reset","sessionId":"s1"}`)
	requireJSON(`{"command":"bet","numbers":[2,5],"betType":"double","bet":50,"sessionId":"s2"}`)
	requireJSON(`{"command":"next","sessionId":"s3"}`)
	requireJSON(`{"command":"hint","sessionId":"s4"}`)
	requireJSON(`{"command":"log","sessionId":"s5"}`)
	m.AssertCalled(t, "PlaceBet", []int{2, 5}, domain.TehonbikiBetDouble, 50)
}

func TestTehonbikiWebControllerRejectsUnknownCommand(t *testing.T) {
	m := new(tehonbikiControllerInteractor)
	ctrl := controller.NewTehonbikiWebController(func() usecase.TehonbikiInteractorIF { return m })
	defer ctrl.Stop()
	var input controller.TehonbikiWebInput
	if err := json.Unmarshal([]byte(`{"command":"xyz","sessionId":"unknown"}`), &input); err != nil {
		t.Fatal(err)
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusBadRequest)
	if !strings.Contains(recorded.Body.String(), "Unsupported command") {
		t.Fatalf("body=%s", recorded.Body.String())
	}
}

func TestTehonbikiCuiControllerParsesBetTypeNumbersAndChips(t *testing.T) {
	originalLang := i18n.Lang()
	t.Cleanup(func() { i18n.SetLang(originalLang) })
	i18n.SetLang("en")
	m := new(tehonbikiControllerInteractor)
	m.On("Reset").Return("reset")
	m.On("PlaceBet", []int{1, 4, 6}, domain.TehonbikiBetTriple, 50).Return("bet")
	c := controller.NewTehonbikiCuiController(m)
	if got := c.Exec("bet triple 1 4 6 50"); got != "bet" {
		t.Fatalf("got %q", got)
	}
	m.AssertCalled(t, "PlaceBet", []int{1, 4, 6}, domain.TehonbikiBetTriple, 50)
	if got := c.Exec("bet triple 1 x 6 50"); func() bool {
		body, isErr := i18n.StripErrorPrefix(got)
		return !isErr || body != "Invalid index. Please enter a number."
	}() {
		t.Fatalf("got %q", got)
	}
	if got := c.Exec("bet single 1"); !strings.Contains(got, "Bet is required.") {
		t.Fatalf("got %q", got)
	}
}
