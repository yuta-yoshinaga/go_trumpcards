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

func TestSjavsWebInput_ToConfigWithNoConfigDoesNotPanic(t *testing.T) {
	// `config` is optional on the wire, so a plain reset arrives with a nil
	// *SjavsWebConfig -- the crash E2E found in Bura.
	var input controller.SjavsWebInput
	assert.NotPanics(t, func() {
		assert.Equal(t, domain.DefaultSjavsConfig(), input.ToConfig())
	})
}

func TestSjavsWebInput_ToConfigClampsAnOutOfRangeDifficulty(t *testing.T) {
	bad := 99
	cfg := controller.SjavsWebInput{
		Config: &controller.SjavsWebConfig{CpuDifficulty: &bad},
	}.ToConfig()
	assert.NoError(t, cfg.Validate())
}

func TestNewSjavsDefaultOutput(t *testing.T) {
	// An error response still has to render: the page maps over the index
	// arrays and reads the rubber counters without guarding for absence.
	out := controller.NewSjavsDefaultOutputForTest("boom")
	assert.Equal(t, -1, out.TrumpSuit, "trump is undecided until bidding ends")
	assert.Equal(t, -1, out.BidderIdx)
	assert.Equal(t, -1, out.WinnerTeam)
	assert.Equal(t, domain.SjavsMinBid, out.MinBid)
	// 残りを 0 で初期化すると、エラー画面がラバー決着済みに見える。
	assert.Equal(t, []int{domain.SjavsRubber, domain.SjavsRubber}, out.Remaining)
	assert.Equal(t, "boom", out.Message)
	assert.NotNil(t, out.Players)
	assert.NotNil(t, out.Trick)
	assert.NotNil(t, out.ValidIndices)
}

func TestSjavsWebController_Dispatch(t *testing.T) {
	const mockOutput = `{"phase":0}`

	siMock := new(ucmock.MockSjavsInteractor)
	siMock.On("ResetWithConfig", domain.DefaultSjavsConfig()).Return(mockOutput)
	siMock.On("Bid", 6).Return(mockOutput)
	siMock.On("Bid", 0).Return(mockOutput)
	siMock.On("Play", 2).Return(mockOutput)
	siMock.On("NextHand").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewSjavsWebController(func() uc.SjavsInteractorIF { return siMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.SjavsWebInput
		if err := json.Unmarshal([]byte(body), &input); err != nil {
			t.Fatalf("unmarshal %s: %v", body, err)
		}
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("reset without a config block", func(t *testing.T) {
		exec(t, `{"command":"r","sessionId":"j1"}`).CodeIs(http.StatusOK)
	})

	t.Run("bid, pass, play and next", func(t *testing.T) {
		exec(t, `{"command":"b","bidLength":6,"sessionId":"j1"}`).CodeIs(http.StatusOK)
		// 0 はパス。省略と同じ扱いにすると、パスが「必須項目の欠落」になる。
		exec(t, `{"command":"b","bidLength":0,"sessionId":"j1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"p","cardIndex":2,"sessionId":"j1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"n","sessionId":"j1"}`).CodeIs(http.StatusOK)
		siMock.AssertCalled(t, "Bid", 6)
		siMock.AssertCalled(t, "Bid", 0)
		siMock.AssertCalled(t, "Play", 2)
		siMock.AssertCalled(t, "NextHand")
	})

	t.Run("missing parameters are errors", func(t *testing.T) {
		for body, want := range map[string]string{
			`{"command":"b","sessionId":"j1"}`: "bidLength is required",
			`{"command":"p","sessionId":"j1"}`: "cardIndex is required",
		} {
			rec := exec(t, body)
			rec.CodeIs(http.StatusBadRequest)
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf("expected %q, got %s", want, rec.Body.String())
			}
		}
	})

	t.Run("hint, log and quit", func(t *testing.T) {
		exec(t, `{"command":"h","sessionId":"j1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"l","sessionId":"j1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"q","sessionId":"j1"}`).CodeIs(http.StatusOK)
	})
}

func newSjavsCui() (*controller.SjavsCuiController, *ucmock.MockSjavsInteractor) {
	si := new(ucmock.MockSjavsInteractor)
	return controller.NewSjavsCuiController(si), si
}

func TestSjavsCuiController_Commands(t *testing.T) {
	c, si := newSjavsCui()
	cfg := domain.DefaultSjavsConfig()
	si.On("GetConfig").Return(cfg)
	si.On("ResetWithConfig", cfg).Return("reset")
	si.On("Bid", 6).Return("bid")
	si.On("Bid", 0).Return("pass")
	si.On("Play", 1).Return("played")
	si.On("NextHand").Return("next")
	si.On("Hint").Return("hint")
	si.On("ActionLog").Return("log")

	assert.Equal(t, "reset", c.Exec("r"))
	assert.Equal(t, "bid", c.Exec("b 6"))
	assert.Equal(t, "pass", c.Exec("b 0"))
	assert.Equal(t, "played", c.Exec("p 1"))
	assert.Equal(t, "next", c.Exec("n"))
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "log", c.Exec("log"))
	assert.Contains(t, c.Exec("q"), "bye")
}

func TestSjavsCuiController_RejectsBadInput(t *testing.T) {
	c, si := newSjavsCui()
	for _, input := range []string{"b", "b x", "p", "p y"} {
		assert.NotEmpty(t, c.Exec(input))
	}
	si.AssertNotCalled(t, "Bid", mock.Anything)
	si.AssertNotCalled(t, "Play", mock.Anything)

	assert.NotEmpty(t, c.Exec("bd"), "a near miss should suggest")
	assert.NotEmpty(t, c.Exec("zzzz"))
}
