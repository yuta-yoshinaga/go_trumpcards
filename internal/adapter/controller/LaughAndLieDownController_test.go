//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	ucmock "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func TestLaughAndLieDownWebInput_ToConfigWithNoConfigDoesNotPanic(t *testing.T) {
	// `config` is optional on the wire, so a plain reset arrives with a nil
	// *LaughAndLieDownWebConfig -- the crash E2E found in Bura.
	var input controller.LaughAndLieDownWebInput
	assert.NotPanics(t, func() {
		assert.Equal(t, domain.DefaultLaughAndLieDownConfig(), input.ToConfig())
	})
}

func TestLaughAndLieDownWebInput_ToConfigClampsAnOutOfRangeDifficulty(t *testing.T) {
	bad := 99
	cfg := controller.LaughAndLieDownWebInput{
		Config: &controller.LaughAndLieDownWebConfig{CpuDifficulty: &bad},
	}.ToConfig()
	assert.NoError(t, cfg.Validate())
}

func TestLaughAndLieDownWebInput_TakeCountDefaultsToOne(t *testing.T) {
	// 1 枚取りが普通なので、毎リクエストに takeCount を要求しない。省略が
	// 3 枚取りに化けると、指定していない札まで取られる。
	var input controller.LaughAndLieDownWebInput
	assert.Equal(t, 1, input.TakeCountOrOne())

	three := 3
	assert.Equal(t, 3, controller.LaughAndLieDownWebInput{TakeCount: &three}.TakeCountOrOne())

	// 不正値はここで潰さず、ドメインの検証へ素通しする (規則の持ち主は 1 箇所)。
	bad := 2
	assert.Equal(t, 2, controller.LaughAndLieDownWebInput{TakeCount: &bad}.TakeCountOrOne())
}

func TestNewLaughAndLieDownDefaultOutput(t *testing.T) {
	// An error response still has to render: the page maps over layout and the
	// index arrays without guarding for absence, and prints the pot.
	out := controller.NewLaughAndLieDownDefaultOutputForTest("boom")
	assert.Equal(t, domain.LaughAndLieDownPot, out.Pot)
	assert.Equal(t, -1, out.LastInIdx)
	assert.Equal(t, "boom", out.Message)
	assert.NotNil(t, out.Players)
	assert.NotNil(t, out.Layout)
	assert.NotNil(t, out.ValidIndices)
	assert.NotNil(t, out.ThreeTakeIndices)
}

func TestLaughAndLieDownWebController_Dispatch(t *testing.T) {
	const mockOutput = `{"phase":0}`

	liMock := new(ucmock.MockLaughAndLieDownInteractor)
	liMock.On("ResetWithConfig", domain.DefaultLaughAndLieDownConfig()).Return(mockOutput)
	liMock.On("Play", 2, 1).Return(mockOutput)
	liMock.On("Play", 4, 3).Return(mockOutput)
	liMock.On("Hint").Return(mockOutput)
	liMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewLaughAndLieDownWebController(func() uc.LaughAndLieDownInteractorIF { return liMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.LaughAndLieDownWebInput
		if err := json.Unmarshal([]byte(body), &input); err != nil {
			t.Fatalf("unmarshal %s: %v", body, err)
		}
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("reset without a config block", func(t *testing.T) {
		exec(t, `{"command":"r","sessionId":"l1"}`).CodeIs(http.StatusOK)
	})

	t.Run("play defaults to one and honours three", func(t *testing.T) {
		exec(t, `{"command":"p","cardIndex":2,"sessionId":"l1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"p","cardIndex":4,"takeCount":3,"sessionId":"l1"}`).CodeIs(http.StatusOK)
		liMock.AssertCalled(t, "Play", 2, 1)
		liMock.AssertCalled(t, "Play", 4, 3)
	})

	t.Run("a missing card index is a parameter error", func(t *testing.T) {
		rec := exec(t, `{"command":"p","sessionId":"l1"}`)
		rec.CodeIs(http.StatusBadRequest)
		if !strings.Contains(rec.Body.String(), "cardIndex is required") {
			t.Errorf("expected a cardIndex error, got %s", rec.Body.String())
		}
	})

	t.Run("hint, log and quit", func(t *testing.T) {
		exec(t, `{"command":"h","sessionId":"l1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"l","sessionId":"l1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"q","sessionId":"l1"}`).CodeIs(http.StatusOK)
	})
}

func newLaughAndLieDownCui() (*controller.LaughAndLieDownCuiController, *ucmock.MockLaughAndLieDownInteractor) {
	li := new(ucmock.MockLaughAndLieDownInteractor)
	return controller.NewLaughAndLieDownCuiController(li), li
}

func TestLaughAndLieDownCuiController_Commands(t *testing.T) {
	c, li := newLaughAndLieDownCui()
	cfg := domain.DefaultLaughAndLieDownConfig()
	li.On("GetConfig").Return(cfg)
	li.On("ResetWithConfig", cfg).Return("reset")
	li.On("Play", 1, 1).Return("took one")
	li.On("Play", 1, 3).Return("took three")
	li.On("Hint").Return("hint")
	li.On("ActionLog").Return("log")

	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "took one", c.Exec("p 1"), "the take count is optional")
	assert.Equal(t, "took three", c.Exec("p 1 3"))
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "log", c.Exec("log"))
	assert.Contains(t, c.Exec("q"), "bye")
}

func TestLaughAndLieDownCuiController_RejectsBadInput(t *testing.T) {
	c, li := newLaughAndLieDownCui()
	for _, input := range []string{"p", "p x", "p -1", "p 1 x"} {
		assert.NotEmpty(t, c.Exec(input))
	}
	li.AssertNotCalled(t, "Play", mock.Anything, mock.Anything)

	assert.NotEmpty(t, c.Exec("pla"), "a near miss should suggest")
	assert.NotEmpty(t, c.Exec("zzzz"))
}
