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

func TestLobaWebInput_ToConfigWithNoConfigDoesNotPanic(t *testing.T) {
	// `config` is optional on the wire, so a plain reset arrives with a nil
	// *LobaWebConfig -- the crash E2E found in Bura.
	var input controller.LobaWebInput
	assert.NotPanics(t, func() {
		assert.Equal(t, domain.DefaultLobaConfig(), input.ToConfig())
	})
}

func TestLobaWebInput_ToConfigClampsAnOutOfRangeDifficulty(t *testing.T) {
	bad := 99
	cfg := controller.LobaWebInput{
		Config: &controller.LobaWebConfig{CpuDifficulty: &bad},
	}.ToConfig()
	assert.NoError(t, cfg.Validate())
}

func TestNewLobaDefaultOutput(t *testing.T) {
	// An error response still has to render: the page maps over melds and
	// prints the knock-out threshold.
	out := controller.NewLobaDefaultOutputForTest("boom")
	assert.Equal(t, domain.LobaKnockOut, out.KnockOut)
	assert.Equal(t, -1, out.RoundWinner)
	assert.Equal(t, -1, out.WinnerIdx)
	assert.Equal(t, "boom", out.Message)
	assert.NotNil(t, out.Players)
	assert.NotNil(t, out.Melds)
}

func TestLobaWebController_Dispatch(t *testing.T) {
	const mockOutput = `{"phase":0}`

	liMock := new(ucmock.MockLobaInteractor)
	liMock.On("ResetWithConfig", domain.DefaultLobaConfig()).Return(mockOutput)
	liMock.On("DrawStock").Return(mockOutput)
	liMock.On("DrawDiscard").Return(mockOutput)
	liMock.On("Meld", []int{0, 2, 4}).Return(mockOutput)
	liMock.On("LayOff", 1, 0).Return(mockOutput)
	liMock.On("Discard", 3).Return(mockOutput)
	liMock.On("NextRound").Return(mockOutput)
	liMock.On("Hint").Return(mockOutput)
	liMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewLobaWebController(func() uc.LobaInteractorIF { return liMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.LobaWebInput
		if err := json.Unmarshal([]byte(body), &input); err != nil {
			t.Fatalf("unmarshal %s: %v", body, err)
		}
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("reset without a config block", func(t *testing.T) {
		exec(t, `{"command":"r","sessionId":"b1"}`).CodeIs(http.StatusOK)
	})

	t.Run("the two draws are distinct commands", func(t *testing.T) {
		exec(t, `{"command":"ds","sessionId":"b1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"dd","sessionId":"b1"}`).CodeIs(http.StatusOK)
		liMock.AssertCalled(t, "DrawStock")
		liMock.AssertCalled(t, "DrawDiscard")
	})

	t.Run("meld, layoff, discard and next", func(t *testing.T) {
		exec(t, `{"command":"m","cardIndices":[0,2,4],"sessionId":"b1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"o","cardIndex":1,"meldIndex":0,"sessionId":"b1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"d","cardIndex":3,"sessionId":"b1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"n","sessionId":"b1"}`).CodeIs(http.StatusOK)
		liMock.AssertCalled(t, "Meld", []int{0, 2, 4})
		// カード添字とメルド添字は別物。取り違えると別のメルドに付けてしまう。
		liMock.AssertCalled(t, "LayOff", 1, 0)
		liMock.AssertCalled(t, "Discard", 3)
	})

	t.Run("missing parameters are errors", func(t *testing.T) {
		for body, want := range map[string]string{
			`{"command":"m","sessionId":"b1"}`:               "cardIndices is required",
			`{"command":"d","sessionId":"b1"}`:               "cardIndex is required",
			`{"command":"o","cardIndex":1,"sessionId":"b1"}`: "cardIndex and meldIndex are required",
			`{"command":"o","meldIndex":0,"sessionId":"b1"}`: "cardIndex and meldIndex are required",
		} {
			rec := exec(t, body)
			rec.CodeIs(http.StatusBadRequest)
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf("expected %q, got %s", want, rec.Body.String())
			}
		}
	})

	t.Run("hint, log and quit", func(t *testing.T) {
		exec(t, `{"command":"h","sessionId":"b1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"l","sessionId":"b1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"q","sessionId":"b1"}`).CodeIs(http.StatusOK)
	})
}

func newLobaCui() (*controller.LobaCuiController, *ucmock.MockLobaInteractor) {
	li := new(ucmock.MockLobaInteractor)
	return controller.NewLobaCuiController(li), li
}

func TestLobaCuiController_Commands(t *testing.T) {
	c, li := newLobaCui()
	cfg := domain.DefaultLobaConfig()
	li.On("GetConfig").Return(cfg)
	li.On("ResetWithConfig", cfg).Return("reset")
	li.On("DrawStock").Return("stock")
	li.On("DrawDiscard").Return("discard pile")
	li.On("Meld", []int{0, 2, 5}).Return("melded")
	li.On("LayOff", 1, 0).Return("laid off")
	li.On("Discard", 3).Return("discarded")
	li.On("NextRound").Return("next")
	li.On("Hint").Return("hint")
	li.On("ActionLog").Return("log")

	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "stock", c.Exec("ds"))
	assert.Equal(t, "discard pile", c.Exec("dd"))
	// 複数枚を 1 コマンドで指定する必要があるのでカンマ区切り。
	assert.Equal(t, "melded", c.Exec("m 0,2,5"))
	assert.Equal(t, "laid off", c.Exec("o 1 0"))
	assert.Equal(t, "discarded", c.Exec("d 3"))
	assert.Equal(t, "next", c.Exec("n"))
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "log", c.Exec("log"))
	assert.Contains(t, c.Exec("q"), "bye")
}

func TestLobaCuiController_RejectsBadInput(t *testing.T) {
	c, li := newLobaCui()
	for _, input := range []string{"m", "m 0,x,2", "o", "o 1", "o x 1", "o 1 y", "d", "d z"} {
		assert.NotEmpty(t, c.Exec(input))
	}
	li.AssertNotCalled(t, "Meld", mock.Anything)
	li.AssertNotCalled(t, "LayOff", mock.Anything, mock.Anything)
	li.AssertNotCalled(t, "Discard", mock.Anything)

	assert.NotEmpty(t, c.Exec("mel"), "a near miss should suggest")
	assert.NotEmpty(t, c.Exec("zzzz"))
}
