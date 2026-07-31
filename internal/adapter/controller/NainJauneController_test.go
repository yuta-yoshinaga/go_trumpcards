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

func TestNainJauneWebInput_ToConfigWithNoConfigDoesNotPanic(t *testing.T) {
	// `config` is optional on the wire, so a plain reset arrives with a nil
	// *NainJauneWebConfig -- the crash E2E found in Bura.
	var input controller.NainJauneWebInput
	assert.NotPanics(t, func() {
		assert.Equal(t, domain.DefaultNainJauneConfig(), input.ToConfig())
	})
}

func TestNainJauneWebInput_ToConfigClampsOutOfRangeValues(t *testing.T) {
	bad, huge := 99, 99999
	cfg := controller.NainJauneWebInput{
		Config: &controller.NainJauneWebConfig{CpuDifficulty: &bad, TargetDeals: &huge},
	}.ToConfig()
	assert.NoError(t, cfg.Validate())
}

func TestNewNainJauneDefaultOutput(t *testing.T) {
	// An error response still has to render: the page maps over five boxes
	// and the award list.
	out := controller.NewNainJauneDefaultOutputForTest("boom")
	assert.Equal(t, domain.DefaultNainJauneConfig().TargetDeals, out.TargetDeals)
	assert.Equal(t, -1, out.DealWinner)
	assert.Equal(t, -1, out.WinnerIdx)
	// **0 は「好きな札で始められる」という意味を持つ。**この game は
	// スートを持たないので、並びの状態はランクだけで表す。
	assert.Equal(t, 0, out.RunRank)
	assert.Equal(t, "boom", out.Message)
	assert.NotNil(t, out.Players)
	assert.NotNil(t, out.Boxes)
	assert.NotNil(t, out.Awards)
	assert.NotNil(t, out.PlayedPile)
}

func TestNainJauneWebController_Dispatch(t *testing.T) {
	const mockOutput = `{"phase":0}`

	piMock := new(ucmock.MockNainJauneInteractor)
	piMock.On("ResetWithConfig", domain.DefaultNainJauneConfig()).Return(mockOutput)
	piMock.On("Play", 2).Return(mockOutput)
	piMock.On("NextDeal").Return(mockOutput)
	piMock.On("Hint").Return(mockOutput)
	piMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewNainJauneWebController(func() uc.NainJauneInteractorIF { return piMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.NainJauneWebInput
		if err := json.Unmarshal([]byte(body), &input); err != nil {
			t.Fatalf("unmarshal %s: %v", body, err)
		}
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("reset without a config block", func(t *testing.T) {
		exec(t, `{"command":"r","sessionId":"j1"}`).CodeIs(http.StatusOK)
	})

	t.Run("play and next", func(t *testing.T) {
		exec(t, `{"command":"p","cardIndex":2,"sessionId":"j1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"n","sessionId":"j1"}`).CodeIs(http.StatusOK)
		piMock.AssertCalled(t, "Play", 2)
		piMock.AssertCalled(t, "NextDeal")
	})

	t.Run("play needs a card index", func(t *testing.T) {
		rec := exec(t, `{"command":"p","sessionId":"j1"}`)
		rec.CodeIs(http.StatusBadRequest)
		if !strings.Contains(rec.Body.String(), "cardIndex is required") {
			t.Errorf("got %s", rec.Body.String())
		}
	})

	t.Run("hint, log and quit", func(t *testing.T) {
		exec(t, `{"command":"h","sessionId":"j1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"l","sessionId":"j1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"q","sessionId":"j1"}`).CodeIs(http.StatusOK)
		piMock.AssertCalled(t, "ActionLog")
	})
}

func newNainJauneCui() (*controller.NainJauneCuiController, *ucmock.MockNainJauneInteractor) {
	pi := new(ucmock.MockNainJauneInteractor)
	return controller.NewNainJauneCuiController(pi), pi
}

func TestNainJauneCuiController_Commands(t *testing.T) {
	c, pi := newNainJauneCui()
	cfg := domain.DefaultNainJauneConfig()
	pi.On("GetConfig").Return(cfg)
	pi.On("ResetWithConfig", cfg).Return("reset")
	pi.On("Play", 2).Return("played")
	pi.On("NextDeal").Return("next")
	pi.On("Hint").Return("hint")
	pi.On("ActionLog").Return("log")

	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "played", c.Exec("p 2"))
	assert.Equal(t, "next", c.Exec("n"))
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "log", c.Exec("log"))
	assert.Contains(t, c.Exec("q"), "bye")
}

func TestNainJauneCuiController_RejectsBadInput(t *testing.T) {
	c, pi := newNainJauneCui()
	for _, input := range []string{"p", "p z", "p -1"} {
		assert.NotEmpty(t, c.Exec(input), "input %q", input)
	}
	pi.AssertNotCalled(t, "Play", mock.Anything)

	assert.NotEmpty(t, c.Exec("pl"), "a near miss should suggest")
	assert.NotEmpty(t, c.Exec("zzzz"))
}
