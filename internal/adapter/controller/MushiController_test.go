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

func TestMushiWebInput_ToConfigWithNoConfigDoesNotPanic(t *testing.T) {
	// `config` is optional on the wire, so an ordinary reset arrives with a nil
	// *MushiWebConfig. Calling the method on that nil pointer is a 500 on the
	// very request that starts a game -- the bug E2E found in Bura.
	var input controller.MushiWebInput
	assert.NotPanics(t, func() {
		assert.Equal(t, domain.DefaultMushiConfig(), input.ToConfig())
	})
}

func TestMushiWebInput_ToConfigClampsOutOfRangeValues(t *testing.T) {
	badDifficulty, badRounds := 99, 999
	cfg := controller.MushiWebInput{
		Config: &controller.MushiWebConfig{CpuDifficulty: &badDifficulty, TargetRounds: &badRounds},
	}.ToConfig()
	assert.NoError(t, cfg.Validate(), "out-of-range values must be clamped, not passed through")
	assert.LessOrEqual(t, cfg.TargetRounds, domain.MushiMaxRounds)
}

func TestNewMushiDefaultOutput(t *testing.T) {
	// An error response still has to be renderable: the page reads
	// targetRounds and selectableIndices without guarding for absence.
	out := controller.NewMushiDefaultOutputForTest("boom")
	assert.Equal(t, domain.MushiMaxRounds, out.TargetRounds)
	assert.Equal(t, -1, out.WinnerIdx)
	assert.Equal(t, "boom", out.Message)
	assert.NotNil(t, out.Players)
	assert.NotNil(t, out.Field)
	assert.NotNil(t, out.SelectableIndices)
}

func TestMushiWebController_Dispatch(t *testing.T) {
	const mockOutput = `{"phase":0}`

	miMock := new(ucmock.MockMushiInteractor)
	miMock.On("ResetWithConfig", domain.DefaultMushiConfig()).Return(mockOutput)
	miMock.On("Play", 2).Return(mockOutput)
	miMock.On("Select", 1).Return(mockOutput)
	miMock.On("NextRound").Return(mockOutput)
	miMock.On("Hint").Return(mockOutput)
	miMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewMushiWebController(func() uc.MushiInteractorIF { return miMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.MushiWebInput
		if err := json.Unmarshal([]byte(body), &input); err != nil {
			t.Fatalf("unmarshal %s: %v", body, err)
		}
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("reset without a config block", func(t *testing.T) {
		exec(t, `{"command":"r","sessionId":"m1"}`).CodeIs(http.StatusOK)
	})

	t.Run("play and select carry their own index", func(t *testing.T) {
		exec(t, `{"command":"p","cardIndex":2,"sessionId":"m1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"s","fieldIndex":1,"sessionId":"m1"}`).CodeIs(http.StatusOK)
		// A field index must not be read as a card index, or the wrong card is played.
		miMock.AssertCalled(t, "Play", 2)
		miMock.AssertCalled(t, "Select", 1)
	})

	t.Run("missing indices are parameter errors", func(t *testing.T) {
		for body, want := range map[string]string{
			`{"command":"p","sessionId":"m1"}`: "cardIndex is required",
			`{"command":"s","sessionId":"m1"}`: "fieldIndex is required",
		} {
			rec := exec(t, body)
			rec.CodeIs(http.StatusBadRequest)
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf("expected %q, got %s", want, rec.Body.String())
			}
		}
	})

	t.Run("next, hint, log and quit", func(t *testing.T) {
		exec(t, `{"command":"n","sessionId":"m1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"h","sessionId":"m1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"l","sessionId":"m1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"q","sessionId":"m1"}`).CodeIs(http.StatusOK)
	})
}

func newMushiCui() (*controller.MushiCuiController, *ucmock.MockMushiInteractor) {
	mi := new(ucmock.MockMushiInteractor)
	return controller.NewMushiCuiController(mi), mi
}

func TestMushiCuiController_Commands(t *testing.T) {
	c, mi := newMushiCui()
	cfg := domain.DefaultMushiConfig()
	mi.On("GetConfig").Return(cfg)
	mi.On("ResetWithConfig", cfg).Return("reset")
	mi.On("Play", 1).Return("played")
	mi.On("Select", 2).Return("selected")
	mi.On("NextRound").Return("next")
	mi.On("Hint").Return("hint")
	mi.On("ActionLog").Return("log")

	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "played", c.Exec("p 1"))
	assert.Equal(t, "selected", c.Exec("s 2"))
	assert.Equal(t, "next", c.Exec("n"))
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "log", c.Exec("log"))
	assert.Contains(t, c.Exec("q"), "bye")
}

func TestMushiCuiController_RejectsBadInput(t *testing.T) {
	c, mi := newMushiCui()
	for _, input := range []string{"p", "p x", "s", "s y"} {
		assert.NotEmpty(t, c.Exec(input))
	}
	mi.AssertNotCalled(t, "Play", mock.Anything)
	mi.AssertNotCalled(t, "Select", mock.Anything)

	assert.NotEmpty(t, c.Exec("nexr"), "a near miss should suggest")
	assert.NotEmpty(t, c.Exec("zzzz"))
}
