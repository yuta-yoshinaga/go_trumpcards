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

func TestDesmocheWebInput_ToConfigWithNoConfigDoesNotPanic(t *testing.T) {
	// `config` is optional on the wire, so a plain reset arrives with a nil
	// *DesmocheWebConfig -- the crash E2E found in Bura.
	var input controller.DesmocheWebInput
	assert.NotPanics(t, func() {
		assert.Equal(t, domain.DefaultDesmocheConfig(), input.ToConfig())
	})
}

func TestDesmocheWebInput_ToConfigClampsAnOutOfRangeDifficulty(t *testing.T) {
	bad := 99
	cfg := controller.DesmocheWebInput{
		Config: &controller.DesmocheWebConfig{CpuDifficulty: &bad},
	}.ToConfig()
	assert.NoError(t, cfg.Validate())
}

func TestNewDesmocheDefaultOutput(t *testing.T) {
	// An error response still has to render: the page maps over melds and
	// prints how many of the ten cards each seat is down.
	out := controller.NewDesmocheDefaultOutputForTest("boom")
	assert.Equal(t, domain.DesmocheGoOutSize, out.GoOutSize)
	assert.Equal(t, -1, out.RoundWinner)
	assert.Equal(t, -1, out.WinnerIdx)
	assert.Equal(t, "boom", out.Message)
	assert.NotNil(t, out.Players)
	assert.NotNil(t, out.Melds)
}

func TestDesmocheWebController_Dispatch(t *testing.T) {
	const mockOutput = `{"phase":0}`

	diMock := new(ucmock.MockDesmocheInteractor)
	diMock.On("ResetWithConfig", domain.DefaultDesmocheConfig()).Return(mockOutput)
	diMock.On("DrawStock").Return(mockOutput)
	diMock.On("DrawDiscard").Return(mockOutput)
	diMock.On("Meld", []int{0, 2, 4}).Return(mockOutput)
	diMock.On("LayOff", 1, 0).Return(mockOutput)
	diMock.On("Desmoche", 0, 2, 1).Return(mockOutput)
	diMock.On("Discard", 3).Return(mockOutput)
	diMock.On("NextRound").Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewDesmocheWebController(func() uc.DesmocheInteractorIF { return diMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.DesmocheWebInput
		if err := json.Unmarshal([]byte(body), &input); err != nil {
			t.Fatalf("unmarshal %s: %v", body, err)
		}
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("reset without a config block", func(t *testing.T) {
		exec(t, `{"command":"r","sessionId":"d1"}`).CodeIs(http.StatusOK)
	})

	t.Run("the two draws are distinct commands", func(t *testing.T) {
		exec(t, `{"command":"ds","sessionId":"d1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"dd","sessionId":"d1"}`).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "DrawStock")
		diMock.AssertCalled(t, "DrawDiscard")
	})

	t.Run("meld, layoff, desmoche, discard and next", func(t *testing.T) {
		exec(t, `{"command":"m","cardIndices":[0,2,4],"sessionId":"d1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"o","cardIndex":1,"meldIndex":0,"sessionId":"d1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"x","fromMeldIndex":0,"cardIndex":2,"toMeldIndex":1,"sessionId":"d1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"d","cardIndex":3,"sessionId":"d1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"n","sessionId":"d1"}`).CodeIs(http.StatusOK)
		diMock.AssertCalled(t, "Meld", []int{0, 2, 4})
		// カード添字とメルド添字は別物。取り違えると別のメルドに付けてしまう。
		diMock.AssertCalled(t, "LayOff", 1, 0)
		// desmoche は from / card / to の順で渡さないと別のメルドを崩す。
		diMock.AssertCalled(t, "Desmoche", 0, 2, 1)
		diMock.AssertCalled(t, "Discard", 3)
	})

	t.Run("missing parameters are errors", func(t *testing.T) {
		for body, want := range map[string]string{
			`{"command":"m","sessionId":"d1"}`:                                   "cardIndices is required",
			`{"command":"d","sessionId":"d1"}`:                                   "cardIndex is required",
			`{"command":"o","cardIndex":1,"sessionId":"d1"}`:                     "cardIndex and meldIndex are required",
			`{"command":"o","meldIndex":0,"sessionId":"d1"}`:                     "cardIndex and meldIndex are required",
			`{"command":"x","cardIndex":1,"toMeldIndex":1,"sessionId":"d1"}`:     "fromMeldIndex, cardIndex and toMeldIndex are required",
			`{"command":"x","fromMeldIndex":0,"toMeldIndex":1,"sessionId":"d1"}`: "fromMeldIndex, cardIndex and toMeldIndex are required",
			`{"command":"x","fromMeldIndex":0,"cardIndex":1,"sessionId":"d1"}`:   "fromMeldIndex, cardIndex and toMeldIndex are required",
		} {
			rec := exec(t, body)
			rec.CodeIs(http.StatusBadRequest)
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf("expected %q, got %s", want, rec.Body.String())
			}
		}
	})

	t.Run("hint, log and quit", func(t *testing.T) {
		exec(t, `{"command":"h","sessionId":"d1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"l","sessionId":"d1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"q","sessionId":"d1"}`).CodeIs(http.StatusOK)
	})
}

func newDesmocheCui() (*controller.DesmocheCuiController, *ucmock.MockDesmocheInteractor) {
	di := new(ucmock.MockDesmocheInteractor)
	return controller.NewDesmocheCuiController(di), di
}

func TestDesmocheCuiController_Commands(t *testing.T) {
	c, di := newDesmocheCui()
	cfg := domain.DefaultDesmocheConfig()
	di.On("GetConfig").Return(cfg)
	di.On("ResetWithConfig", cfg).Return("reset")
	di.On("DrawStock").Return("stock")
	di.On("DrawDiscard").Return("discard pile")
	di.On("Meld", []int{0, 2, 5}).Return("melded")
	di.On("LayOff", 1, 0).Return("laid off")
	di.On("Desmoche", 0, 2, 1).Return("rearranged")
	di.On("Discard", 3).Return("discarded")
	di.On("NextRound").Return("next")
	di.On("Hint").Return("hint")
	di.On("ActionLog").Return("log")

	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "stock", c.Exec("ds"))
	assert.Equal(t, "discard pile", c.Exec("dd"))
	// 複数枚を 1 コマンドで指定する必要があるのでカンマ区切り。
	assert.Equal(t, "melded", c.Exec("m 0,2,5"))
	assert.Equal(t, "laid off", c.Exec("o 1 0"))
	assert.Equal(t, "rearranged", c.Exec("x 0 2 1"))
	assert.Equal(t, "discarded", c.Exec("d 3"))
	assert.Equal(t, "next", c.Exec("n"))
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "log", c.Exec("log"))
	assert.Contains(t, c.Exec("q"), "bye")
}

func TestDesmocheCuiController_RejectsBadInput(t *testing.T) {
	c, di := newDesmocheCui()
	for _, input := range []string{
		"m", "m 0,x,2", "o", "o 1", "o x 1", "o 1 y",
		"x", "x 0", "x 0 1", "x y 1 2", "x 0 y 2", "x 0 1 y",
		"d", "d z",
	} {
		assert.NotEmpty(t, c.Exec(input))
	}
	di.AssertNotCalled(t, "Meld", mock.Anything)
	di.AssertNotCalled(t, "LayOff", mock.Anything, mock.Anything)
	di.AssertNotCalled(t, "Desmoche", mock.Anything, mock.Anything, mock.Anything)
	di.AssertNotCalled(t, "Discard", mock.Anything)

	assert.NotEmpty(t, c.Exec("mel"), "a near miss should suggest")
	assert.NotEmpty(t, c.Exec("zzzz"))
}
