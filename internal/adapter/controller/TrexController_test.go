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

func TestTrexWebInput_ToConfigWithNoConfigDoesNotPanic(t *testing.T) {
	// `config` is optional on the wire, so a plain reset arrives with a nil
	// *TrexWebConfig -- the crash E2E found in Bura.
	var input controller.TrexWebInput
	assert.NotPanics(t, func() {
		assert.Equal(t, domain.DefaultTrexConfig(), input.ToConfig())
	})
}

func TestTrexWebInput_ToConfigClampsAnOutOfRangeDifficulty(t *testing.T) {
	bad := 99
	cfg := controller.TrexWebInput{
		Config: &controller.TrexWebConfig{CpuDifficulty: &bad},
	}.ToConfig()
	assert.NoError(t, cfg.Validate())
}

func TestNewTrexDefaultOutput(t *testing.T) {
	// An error response still has to render: the page maps over runs, the
	// contract list and the index arrays without guarding for absence.
	out := controller.NewTrexDefaultOutputForTest("boom")
	assert.Equal(t, int(domain.TrexContractNone), out.Contract, "no contract is chosen on an error")
	assert.Equal(t, domain.TrexTotalDeals, out.TotalDeals)
	assert.Equal(t, -1, out.WinnerIdx)
	assert.Equal(t, "boom", out.Message)
	assert.NotNil(t, out.Players)
	assert.NotNil(t, out.AvailableContracts)
	assert.NotNil(t, out.Trick)
	assert.NotNil(t, out.Runs)
	assert.NotNil(t, out.FinishOrder)
	assert.NotNil(t, out.ValidIndices)
}

func TestTrexWebController_Dispatch(t *testing.T) {
	const mockOutput = `{"phase":0}`

	tiMock := new(ucmock.MockTrexInteractor)
	tiMock.On("ResetWithConfig", domain.DefaultTrexConfig()).Return(mockOutput)
	tiMock.On("Choose", 2).Return(mockOutput)
	tiMock.On("Choose", 0).Return(mockOutput)
	tiMock.On("Play", 3).Return(mockOutput)
	tiMock.On("Pass").Return(mockOutput)
	tiMock.On("NextDeal").Return(mockOutput)
	tiMock.On("Hint").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewTrexWebController(func() uc.TrexInteractorIF { return tiMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.TrexWebInput
		if err := json.Unmarshal([]byte(body), &input); err != nil {
			t.Fatalf("unmarshal %s: %v", body, err)
		}
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("reset without a config block", func(t *testing.T) {
		exec(t, `{"command":"r","sessionId":"t1"}`).CodeIs(http.StatusOK)
	})

	t.Run("choose, play, pass and next", func(t *testing.T) {
		exec(t, `{"command":"c","contract":2,"sessionId":"t1"}`).CodeIs(http.StatusOK)
		// Contract 0 is the king of hearts, a real value -- not an omission.
		exec(t, `{"command":"c","contract":0,"sessionId":"t1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"p","cardIndex":3,"sessionId":"t1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"s","sessionId":"t1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"n","sessionId":"t1"}`).CodeIs(http.StatusOK)
		tiMock.AssertCalled(t, "Choose", 2)
		tiMock.AssertCalled(t, "Choose", 0)
		tiMock.AssertCalled(t, "Play", 3)
		tiMock.AssertCalled(t, "Pass")
		tiMock.AssertCalled(t, "NextDeal")
	})

	t.Run("missing parameters are errors", func(t *testing.T) {
		for body, want := range map[string]string{
			`{"command":"c","sessionId":"t1"}`: "contract is required",
			`{"command":"p","sessionId":"t1"}`: "cardIndex is required",
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

func newTrexCui() (*controller.TrexCuiController, *ucmock.MockTrexInteractor) {
	ti := new(ucmock.MockTrexInteractor)
	return controller.NewTrexCuiController(ti), ti
}

func TestTrexCuiController_Commands(t *testing.T) {
	c, ti := newTrexCui()
	cfg := domain.DefaultTrexConfig()
	ti.On("GetConfig").Return(cfg)
	ti.On("ResetWithConfig", cfg).Return("reset")
	ti.On("Choose", 4).Return("chose")
	ti.On("Play", 1).Return("played")
	ti.On("Pass").Return("passed")
	ti.On("NextDeal").Return("next")
	ti.On("Hint").Return("hint")
	ti.On("ActionLog").Return("log")

	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "chose", c.Exec("c 4"))
	assert.Equal(t, "played", c.Exec("p 1"))
	assert.Equal(t, "passed", c.Exec("s"))
	assert.Equal(t, "next", c.Exec("n"))
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "log", c.Exec("log"))
	assert.Contains(t, c.Exec("q"), "bye")
}

func TestTrexCuiController_RejectsBadInput(t *testing.T) {
	c, ti := newTrexCui()
	for _, input := range []string{"c", "c x", "p", "p y"} {
		assert.NotEmpty(t, c.Exec(input))
	}
	ti.AssertNotCalled(t, "Choose", mock.Anything)
	ti.AssertNotCalled(t, "Play", mock.Anything)

	assert.NotEmpty(t, c.Exec("chose"), "a near miss should suggest")
	assert.NotEmpty(t, c.Exec("zzzz"))
}
