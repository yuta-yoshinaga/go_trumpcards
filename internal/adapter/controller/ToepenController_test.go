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

func TestToepenWebInput_ToConfigWithNoConfigDoesNotPanic(t *testing.T) {
	// `config` is optional on the wire, so a plain reset arrives with a nil
	// *ToepenWebConfig -- the crash E2E found in Bura.
	var input controller.ToepenWebInput
	assert.NotPanics(t, func() {
		assert.Equal(t, domain.DefaultToepenConfig(), input.ToConfig())
	})
}

func TestToepenWebInput_ToConfigClampsOutOfRangeValues(t *testing.T) {
	badDifficulty, badSeats := 99, 999
	cfg := controller.ToepenWebInput{
		Config: &controller.ToepenWebConfig{CpuDifficulty: &badDifficulty, PlayerCnt: &badSeats},
	}.ToConfig()
	assert.NoError(t, cfg.Validate())
	assert.LessOrEqual(t, cfg.PlayerCnt, domain.ToepenMaxPlayers)
	assert.GreaterOrEqual(t, cfg.PlayerCnt, domain.ToepenMinPlayers)
}

func TestNewToepenDefaultOutput(t *testing.T) {
	// An error response still has to render: the page reads stake, maxLives and
	// validPlayIndices without guarding for absence.
	out := controller.NewToepenDefaultOutputForTest("boom")
	assert.Equal(t, 1, out.Stake, "a stake of zero would make a hand cost nothing")
	assert.Equal(t, domain.ToepenMaxLives, out.MaxLives)
	assert.Equal(t, -1, out.WinnerIdx)
	assert.Equal(t, -1, out.LeadSuit)
	assert.Equal(t, -1, out.PendingRespondent)
	assert.Equal(t, "boom", out.Message)
	assert.NotNil(t, out.Players)
	assert.NotNil(t, out.ValidPlayIndices)
}

func TestToepenWebController_Dispatch(t *testing.T) {
	const mockOutput = `{"phase":0}`

	tiMock := new(ucmock.MockToepenInteractor)
	tiMock.On("ResetWithConfig", domain.DefaultToepenConfig()).Return(mockOutput)
	tiMock.On("Play", 2).Return(mockOutput)
	tiMock.On("Toep").Return(mockOutput)
	tiMock.On("Respond", true).Return(mockOutput)
	tiMock.On("Respond", false).Return(mockOutput)
	tiMock.On("NextHand").Return(mockOutput)
	tiMock.On("Hint").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewToepenWebController(func() uc.ToepenInteractorIF { return tiMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.ToepenWebInput
		if err := json.Unmarshal([]byte(body), &input); err != nil {
			t.Fatalf("unmarshal %s: %v", body, err)
		}
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("reset without a config block", func(t *testing.T) {
		exec(t, `{"command":"r","sessionId":"t1"}`).CodeIs(http.StatusOK)
	})

	t.Run("play, toep and next", func(t *testing.T) {
		exec(t, `{"command":"p","cardIndex":2,"sessionId":"t1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"t","sessionId":"t1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"n","sessionId":"t1"}`).CodeIs(http.StatusOK)
		tiMock.AssertCalled(t, "Play", 2)
		tiMock.AssertCalled(t, "Toep")
	})

	t.Run("answer carries the decision both ways", func(t *testing.T) {
		// `stay` is a separate parameter from cardIndex precisely so a fold can
		// never be read as a play.
		exec(t, `{"command":"a","stay":true,"sessionId":"t1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"a","stay":false,"sessionId":"t1"}`).CodeIs(http.StatusOK)
		tiMock.AssertCalled(t, "Respond", true)
		tiMock.AssertCalled(t, "Respond", false)
	})

	t.Run("missing parameters are errors", func(t *testing.T) {
		for body, want := range map[string]string{
			`{"command":"p","sessionId":"t1"}`: "cardIndex is required",
			`{"command":"a","sessionId":"t1"}`: "stay is required",
		} {
			rec := exec(t, body)
			rec.CodeIs(http.StatusBadRequest)
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf("expected %q, got %s", want, rec.Body.String())
			}
		}
	})

	t.Run("hint, log and quit", func(t *testing.T) {
		exec(t, `{"command":"h","sessionId":"t1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"l","sessionId":"t1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"q","sessionId":"t1"}`).CodeIs(http.StatusOK)
	})
}

func newToepenCui() (*controller.ToepenCuiController, *ucmock.MockToepenInteractor) {
	ti := new(ucmock.MockToepenInteractor)
	return controller.NewToepenCuiController(ti), ti
}

func TestToepenCuiController_Commands(t *testing.T) {
	c, ti := newToepenCui()
	cfg := domain.DefaultToepenConfig()
	ti.On("GetConfig").Return(cfg)
	ti.On("ResetWithConfig", cfg).Return("reset")
	ti.On("Play", 1).Return("played")
	ti.On("Toep").Return("toeped")
	ti.On("Respond", true).Return("stayed")
	ti.On("Respond", false).Return("folded")
	ti.On("NextHand").Return("next")
	ti.On("Hint").Return("hint")
	ti.On("ActionLog").Return("log")

	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "played", c.Exec("p 1"))
	assert.Equal(t, "toeped", c.Exec("t"))
	// stay and fold are distinct words so a mistyped boolean cannot invert them.
	assert.Equal(t, "stayed", c.Exec("s"))
	assert.Equal(t, "folded", c.Exec("f"))
	assert.Equal(t, "next", c.Exec("n"))
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "log", c.Exec("log"))
	assert.Contains(t, c.Exec("q"), "bye")
}

func TestToepenCuiController_RejectsBadInput(t *testing.T) {
	c, ti := newToepenCui()
	for _, input := range []string{"p", "p x"} {
		assert.NotEmpty(t, c.Exec(input))
	}
	ti.AssertNotCalled(t, "Play", mock.Anything)
	assert.NotEmpty(t, c.Exec("tope"), "a near miss should suggest")
	assert.NotEmpty(t, c.Exec("zzzz"))
}
