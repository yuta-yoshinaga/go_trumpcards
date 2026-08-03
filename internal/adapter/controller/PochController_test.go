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

func TestPochWebInput_ToConfigWithNoConfigDoesNotPanic(t *testing.T) {
	// `config` is optional on the wire, so a plain reset arrives with a nil
	// *PochWebConfig -- the crash E2E found in Bura.
	var input controller.PochWebInput
	assert.NotPanics(t, func() {
		assert.Equal(t, domain.DefaultPochConfig(), input.ToConfig())
	})
}

func TestPochWebInput_ToConfigClampsOutOfRangeValues(t *testing.T) {
	bad, huge := 99, 99999
	cfg := controller.PochWebInput{
		Config: &controller.PochWebConfig{CpuDifficulty: &bad, TargetDeals: &huge},
	}.ToConfig()
	assert.NoError(t, cfg.Validate())
}

func TestNewPochDefaultOutput(t *testing.T) {
	// An error response still has to render: the page maps over the nine pools
	// and the staking results.
	out := controller.NewPochDefaultOutputForTest("boom")
	assert.Equal(t, domain.DefaultPochConfig().TargetDeals, out.TargetDeals)
	assert.Equal(t, -1, out.PochenWinner)
	assert.Equal(t, -1, out.DealWinner)
	assert.Equal(t, -1, out.WinnerIdx)
	// **-1 は「好きな札で始められる」という意味を持つ。**0 だと ♠ の並びの
	// 途中に見えてしまう。
	assert.Equal(t, -1, out.StopsSuit)
	assert.Equal(t, "boom", out.Message)
	assert.NotNil(t, out.Players)
	assert.NotNil(t, out.Pools)
	assert.NotNil(t, out.StakingAwards)
	assert.NotNil(t, out.PlayedPile)
}

func TestPochWebController_Dispatch(t *testing.T) {
	const mockOutput = `{"phase":0}`

	piMock := new(ucmock.MockPochInteractor)
	piMock.On("ResetWithConfig", domain.DefaultPochConfig()).Return(mockOutput)
	piMock.On("Bet").Return(mockOutput)
	piMock.On("Fold").Return(mockOutput)
	piMock.On("Play", 2).Return(mockOutput)
	piMock.On("NextDeal").Return(mockOutput)
	piMock.On("Hint").Return(mockOutput)
	piMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewPochWebController(func() uc.PochInteractorIF { return piMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.PochWebInput
		if err := json.Unmarshal([]byte(body), &input); err != nil {
			t.Fatalf("unmarshal %s: %v", body, err)
		}
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("reset without a config block", func(t *testing.T) {
		exec(t, `{"command":"r","sessionId":"p1"}`).CodeIs(http.StatusOK)
	})

	t.Run("bet, fold, play and next", func(t *testing.T) {
		exec(t, `{"command":"b","sessionId":"p1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"f","sessionId":"p1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"p","cardIndex":2,"sessionId":"p1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"n","sessionId":"p1"}`).CodeIs(http.StatusOK)
		piMock.AssertCalled(t, "Bet")
		piMock.AssertCalled(t, "Fold")
		piMock.AssertCalled(t, "Play", 2)
		piMock.AssertCalled(t, "NextDeal")
	})

	t.Run("play needs a card index", func(t *testing.T) {
		rec := exec(t, `{"command":"p","sessionId":"p1"}`)
		rec.CodeIs(http.StatusBadRequest)
		if !strings.Contains(rec.Body.String(), "cardIndex is required") {
			t.Errorf("got %s", rec.Body.String())
		}
	})

	t.Run("hint, log and quit", func(t *testing.T) {
		exec(t, `{"command":"h","sessionId":"p1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"l","sessionId":"p1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"q","sessionId":"p1"}`).CodeIs(http.StatusOK)
		piMock.AssertCalled(t, "ActionLog")
	})
}

func newPochCui() (*controller.PochCuiController, *ucmock.MockPochInteractor) {
	pi := new(ucmock.MockPochInteractor)
	return controller.NewPochCuiController(pi), pi
}

func TestPochCuiController_Commands(t *testing.T) {
	c, pi := newPochCui()
	cfg := domain.DefaultPochConfig()
	pi.On("GetConfig").Return(cfg)
	pi.On("ResetWithConfig", cfg).Return("reset")
	pi.On("Bet").Return("bet")
	pi.On("Fold").Return("fold")
	pi.On("Play", 2).Return("played")
	pi.On("NextDeal").Return("next")
	pi.On("Hint").Return("hint")
	pi.On("ActionLog").Return("log")

	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "bet", c.Exec("b"))
	assert.Equal(t, "fold", c.Exec("f"))
	assert.Equal(t, "played", c.Exec("p 2"))
	assert.Equal(t, "next", c.Exec("n"))
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "log", c.Exec("log"))
	assert.Contains(t, c.Exec("q"), "bye")
}

func TestPochCuiController_RejectsBadInput(t *testing.T) {
	c, pi := newPochCui()
	for _, input := range []string{"p", "p z", "p -1"} {
		assert.NotEmpty(t, c.Exec(input), "input %q", input)
	}
	pi.AssertNotCalled(t, "Play", mock.Anything)

	assert.NotEmpty(t, c.Exec("bt"), "a near miss should suggest")
	assert.NotEmpty(t, c.Exec("zzzz"))
}
