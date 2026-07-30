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

func TestChineseTenWebInput_ToConfigWithNoConfigDoesNotPanic(t *testing.T) {
	// `config` is optional on the wire, so a plain reset arrives with a nil
	// *ChineseTenWebConfig -- the crash E2E found in Bura.
	var input controller.ChineseTenWebInput
	assert.NotPanics(t, func() {
		assert.Equal(t, domain.DefaultChineseTenConfig(), input.ToConfig())
	})
}

func TestChineseTenWebInput_ToConfigClampsAnOutOfRangeDifficulty(t *testing.T) {
	bad := 99
	cfg := controller.ChineseTenWebInput{
		Config: &controller.ChineseTenWebConfig{CpuDifficulty: &bad},
	}.ToConfig()
	assert.NoError(t, cfg.Validate())
}

func TestNewChineseTenDefaultOutput(t *testing.T) {
	// An error response still has to render: the page reads tieScore and
	// selectableIndices without guarding for absence.
	out := controller.NewChineseTenDefaultOutputForTest("boom")
	assert.Equal(t, domain.ChineseTenTieScore, out.TieScore)
	assert.Equal(t, -1, out.WinnerIdx)
	assert.Equal(t, "boom", out.Message)
	assert.NotNil(t, out.Players)
	assert.NotNil(t, out.Layout)
	assert.NotNil(t, out.SelectableIndices)
}

func TestChineseTenWebController_Dispatch(t *testing.T) {
	const mockOutput = `{"phase":0}`

	ciMock := new(ucmock.MockChineseTenInteractor)
	ciMock.On("ResetWithConfig", domain.DefaultChineseTenConfig()).Return(mockOutput)
	ciMock.On("Play", 2).Return(mockOutput)
	ciMock.On("Select", 1).Return(mockOutput)
	ciMock.On("Hint").Return(mockOutput)
	ciMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewChineseTenWebController(func() uc.ChineseTenInteractorIF { return ciMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.ChineseTenWebInput
		if err := json.Unmarshal([]byte(body), &input); err != nil {
			t.Fatalf("unmarshal %s: %v", body, err)
		}
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("reset without a config block", func(t *testing.T) {
		exec(t, `{"command":"r","sessionId":"c1"}`).CodeIs(http.StatusOK)
	})

	t.Run("play and select carry their own index", func(t *testing.T) {
		exec(t, `{"command":"p","cardIndex":2,"sessionId":"c1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"s","layoutIndex":1,"sessionId":"c1"}`).CodeIs(http.StatusOK)
		// A layout index must not be read as a hand index, or the wrong card
		// is played.
		ciMock.AssertCalled(t, "Play", 2)
		ciMock.AssertCalled(t, "Select", 1)
	})

	t.Run("missing indices are parameter errors", func(t *testing.T) {
		for body, want := range map[string]string{
			`{"command":"p","sessionId":"c1"}`: "cardIndex is required",
			`{"command":"s","sessionId":"c1"}`: "layoutIndex is required",
		} {
			rec := exec(t, body)
			rec.CodeIs(http.StatusBadRequest)
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf("expected %q, got %s", want, rec.Body.String())
			}
		}
	})

	t.Run("hint, log and quit", func(t *testing.T) {
		exec(t, `{"command":"h","sessionId":"c1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"l","sessionId":"c1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"q","sessionId":"c1"}`).CodeIs(http.StatusOK)
	})
}

func newChineseTenCui() (*controller.ChineseTenCuiController, *ucmock.MockChineseTenInteractor) {
	ci := new(ucmock.MockChineseTenInteractor)
	return controller.NewChineseTenCuiController(ci), ci
}

func TestChineseTenCuiController_Commands(t *testing.T) {
	c, ci := newChineseTenCui()
	cfg := domain.DefaultChineseTenConfig()
	ci.On("GetConfig").Return(cfg)
	ci.On("ResetWithConfig", cfg).Return("reset")
	ci.On("Play", 1).Return("played")
	ci.On("Select", 2).Return("selected")
	ci.On("Hint").Return("hint")
	ci.On("ActionLog").Return("log")

	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "played", c.Exec("p 1"))
	assert.Equal(t, "selected", c.Exec("s 2"))
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "log", c.Exec("log"))
	assert.Contains(t, c.Exec("q"), "bye")
}

func TestChineseTenCuiController_RejectsBadInput(t *testing.T) {
	c, ci := newChineseTenCui()
	for _, input := range []string{"p", "p x", "s", "s y"} {
		assert.NotEmpty(t, c.Exec(input))
	}
	ci.AssertNotCalled(t, "Play", mock.Anything)
	ci.AssertNotCalled(t, "Select", mock.Anything)

	assert.NotEmpty(t, c.Exec("slect"), "a near miss should suggest")
	assert.NotEmpty(t, c.Exec("zzzz"))
}
