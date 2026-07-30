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

func TestSkitgubbeWebInput_ToConfigWithNoConfigDoesNotPanic(t *testing.T) {
	// `config` is optional on the wire, so a plain reset arrives with a nil
	// *SkitgubbeWebConfig -- the crash E2E found in Bura.
	var input controller.SkitgubbeWebInput
	assert.NotPanics(t, func() {
		assert.Equal(t, domain.DefaultSkitgubbeConfig(), input.ToConfig())
	})
}

func TestSkitgubbeWebInput_ToConfigClampsAnOutOfRangeDifficulty(t *testing.T) {
	bad := 99
	cfg := controller.SkitgubbeWebInput{
		Config: &controller.SkitgubbeWebConfig{CpuDifficulty: &bad},
	}.ToConfig()
	assert.NoError(t, cfg.Validate())
}

func TestNewSkitgubbeDefaultOutput(t *testing.T) {
	// An error response still has to render: the page maps over duel, pile and
	// validIndices without guarding for absence, and reads trumpSuit directly.
	out := controller.NewSkitgubbeDefaultOutputForTest("boom")
	assert.Equal(t, -1, out.TrumpSuit, "trump is undecided until the stock runs out")
	assert.Equal(t, -1, out.LoserIdx)
	assert.Equal(t, "boom", out.Message)
	assert.NotNil(t, out.Players)
	assert.NotNil(t, out.Duel)
	assert.NotNil(t, out.Pile)
	assert.NotNil(t, out.ValidIndices)
}

func TestSkitgubbeWebController_Dispatch(t *testing.T) {
	const mockOutput = `{"phase":0}`

	siMock := new(ucmock.MockSkitgubbeInteractor)
	siMock.On("ResetWithConfig", domain.DefaultSkitgubbeConfig()).Return(mockOutput)
	siMock.On("Play", 2).Return(mockOutput)
	siMock.On("PickUp").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewSkitgubbeWebController(func() uc.SkitgubbeInteractorIF { return siMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.SkitgubbeWebInput
		if err := json.Unmarshal([]byte(body), &input); err != nil {
			t.Fatalf("unmarshal %s: %v", body, err)
		}
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("reset without a config block", func(t *testing.T) {
		exec(t, `{"command":"r","sessionId":"s1"}`).CodeIs(http.StatusOK)
	})

	t.Run("play carries a card index, pickup carries none", func(t *testing.T) {
		exec(t, `{"command":"p","cardIndex":2,"sessionId":"s1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"u","sessionId":"s1"}`).CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "Play", 2)
		siMock.AssertCalled(t, "PickUp")
	})

	t.Run("a missing card index is a parameter error", func(t *testing.T) {
		rec := exec(t, `{"command":"p","sessionId":"s1"}`)
		rec.CodeIs(http.StatusBadRequest)
		if !strings.Contains(rec.Body.String(), "cardIndex is required") {
			t.Errorf("expected a cardIndex error, got %s", rec.Body.String())
		}
	})

	t.Run("hint, log and quit", func(t *testing.T) {
		exec(t, `{"command":"h","sessionId":"s1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"l","sessionId":"s1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"q","sessionId":"s1"}`).CodeIs(http.StatusOK)
	})
}

func newSkitgubbeCui() (*controller.SkitgubbeCuiController, *ucmock.MockSkitgubbeInteractor) {
	si := new(ucmock.MockSkitgubbeInteractor)
	return controller.NewSkitgubbeCuiController(si), si
}

func TestSkitgubbeCuiController_Commands(t *testing.T) {
	c, si := newSkitgubbeCui()
	cfg := domain.DefaultSkitgubbeConfig()
	si.On("GetConfig").Return(cfg)
	si.On("ResetWithConfig", cfg).Return("reset")
	si.On("Play", 1).Return("played")
	si.On("PickUp").Return("picked up")
	si.On("Hint").Return("hint")
	si.On("ActionLog").Return("log")

	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "played", c.Exec("p 1"))
	assert.Equal(t, "picked up", c.Exec("u"))
	assert.Equal(t, "picked up", c.Exec("pickup"))
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "log", c.Exec("log"))
	assert.Contains(t, c.Exec("q"), "bye")
}

func TestSkitgubbeCuiController_RejectsBadInput(t *testing.T) {
	c, si := newSkitgubbeCui()
	for _, input := range []string{"p", "p x"} {
		assert.NotEmpty(t, c.Exec(input))
	}
	si.AssertNotCalled(t, "Play", mock.Anything)

	assert.NotEmpty(t, c.Exec("pickp"), "a near miss should suggest")
	assert.NotEmpty(t, c.Exec("zzzz"))
}
