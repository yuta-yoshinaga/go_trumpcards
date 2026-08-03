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

func TestZwickerWebInput_ToConfigWithNoConfigDoesNotPanic(t *testing.T) {
	// `config` is optional on the wire, so a plain reset arrives with a nil
	// *ZwickerWebConfig -- the crash E2E found in Bura.
	var input controller.ZwickerWebInput
	assert.NotPanics(t, func() {
		assert.Equal(t, domain.DefaultZwickerConfig(), input.ToConfig())
	})
}

func TestZwickerWebInput_ToConfigClampsOutOfRangeValues(t *testing.T) {
	bad, huge := 99, 99999
	cfg := controller.ZwickerWebInput{
		Config: &controller.ZwickerWebConfig{CpuDifficulty: &bad, TargetScore: &huge},
	}.ToConfig()
	assert.NoError(t, cfg.Validate())
}

func TestNewZwickerDefaultOutput(t *testing.T) {
	// An error response still has to render: the page maps over the table,
	// the builds and the two team scores.
	out := controller.NewZwickerDefaultOutputForTest("boom")
	assert.Equal(t, domain.DefaultZwickerConfig().TargetScore, out.TargetScore)
	assert.Equal(t, -1, out.WinnerTeam)
	assert.Equal(t, "boom", out.Message)
	assert.NotNil(t, out.Players)
	assert.NotNil(t, out.TableCards)
	assert.NotNil(t, out.Builds)
}

func TestZwickerWebController_Dispatch(t *testing.T) {
	const mockOutput = `{"phase":0}`

	ziMock := new(ucmock.MockZwickerInteractor)
	ziMock.On("ResetWithConfig", domain.DefaultZwickerConfig()).Return(mockOutput)
	ziMock.On("Take", 0, 7, []int{1, 2}, []int{0}).Return(mockOutput)
	ziMock.On("Build", 1, []int{2}, 9).Return(mockOutput)
	ziMock.On("Trail", 3).Return(mockOutput)
	ziMock.On("NextRound").Return(mockOutput)
	ziMock.On("Hint").Return(mockOutput)
	ziMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewZwickerWebController(func() uc.ZwickerInteractorIF { return ziMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.ZwickerWebInput
		if err := json.Unmarshal([]byte(body), &input); err != nil {
			t.Fatalf("unmarshal %s: %v", body, err)
		}
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("reset without a config block", func(t *testing.T) {
		exec(t, `{"command":"r","sessionId":"z1"}`).CodeIs(http.StatusOK)
	})

	t.Run("take carries the played value separately from the card", func(t *testing.T) {
		exec(t, `{"command":"t","cardIndex":0,"playedValue":7,"tableIndices":[1,2],"buildIndices":[0],"sessionId":"z1"}`).
			CodeIs(http.StatusOK)
		// **A と絵札は 2 択を持つ**ので、札だけでは捕獲が決まらない。
		ziMock.AssertCalled(t, "Take", 0, 7, []int{1, 2}, []int{0})
	})

	t.Run("build, trail and next", func(t *testing.T) {
		exec(t, `{"command":"b","cardIndex":1,"tableIndices":[2],"declaredValue":9,"sessionId":"z1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"tr","cardIndex":3,"sessionId":"z1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"n","sessionId":"z1"}`).CodeIs(http.StatusOK)
		ziMock.AssertCalled(t, "Build", 1, []int{2}, 9)
		ziMock.AssertCalled(t, "Trail", 3)
	})

	t.Run("missing parameters are errors", func(t *testing.T) {
		for body, want := range map[string]string{
			`{"command":"t","cardIndex":0,"sessionId":"z1"}`:     "cardIndex and playedValue are required",
			`{"command":"t","playedValue":7,"sessionId":"z1"}`:   "cardIndex and playedValue are required",
			`{"command":"b","cardIndex":0,"sessionId":"z1"}`:     "cardIndex and declaredValue are required",
			`{"command":"b","declaredValue":9,"sessionId":"z1"}`: "cardIndex and declaredValue are required",
			`{"command":"tr","sessionId":"z1"}`:                  "cardIndex is required",
		} {
			rec := exec(t, body)
			rec.CodeIs(http.StatusBadRequest)
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf("expected %q, got %s", want, rec.Body.String())
			}
		}
	})

	// trail のエイリアスに "l" を使うと、棋譜の "l" を奪って log が死ぬ。
	t.Run("l still reaches the action log", func(t *testing.T) {
		exec(t, `{"command":"l","sessionId":"z1"}`).CodeIs(http.StatusOK)
		ziMock.AssertCalled(t, "ActionLog")
	})

	t.Run("hint and quit", func(t *testing.T) {
		exec(t, `{"command":"h","sessionId":"z1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"q","sessionId":"z1"}`).CodeIs(http.StatusOK)
	})
}

func newZwickerCui() (*controller.ZwickerCuiController, *ucmock.MockZwickerInteractor) {
	zi := new(ucmock.MockZwickerInteractor)
	return controller.NewZwickerCuiController(zi), zi
}

func TestZwickerCuiController_Commands(t *testing.T) {
	c, zi := newZwickerCui()
	cfg := domain.DefaultZwickerConfig()
	zi.On("GetConfig").Return(cfg)
	zi.On("ResetWithConfig", cfg).Return("reset")
	zi.On("Take", 0, 7, []int{1, 2}, []int(nil)).Return("took")
	zi.On("Take", 0, 9, []int(nil), []int{1}).Return("took a build")
	zi.On("Take", 0, 9, []int{0}, []int{1}).Return("took both")
	zi.On("Build", 0, []int{1, 2}, 9).Return("built")
	zi.On("Trail", 3).Return("trailed")
	zi.On("NextRound").Return("next")
	zi.On("Hint").Return("hint")
	zi.On("ActionLog").Return("log")

	assert.Equal(t, "reset", c.Exec("r"))
	// 値を明示させるのは、A と絵札が 2 つの値を持つため。
	assert.Equal(t, "took", c.Exec("t 0 7 t:1,2"))
	assert.Equal(t, "took", c.Exec("t 0 7 1,2"), "the t: prefix is optional")
	assert.Equal(t, "took a build", c.Exec("t 0 9 b:1"))
	assert.Equal(t, "took both", c.Exec("t 0 9 t:0 b:1"))
	assert.Equal(t, "built", c.Exec("b 0 1,2 9"))
	assert.Equal(t, "trailed", c.Exec("tr 3"))
	assert.Equal(t, "next", c.Exec("n"))
	assert.Equal(t, "hint", c.Exec("h"))
	assert.Equal(t, "log", c.Exec("log"))
	assert.Contains(t, c.Exec("q"), "bye")
}

func TestZwickerCuiController_RejectsBadInput(t *testing.T) {
	c, zi := newZwickerCui()
	for _, input := range []string{
		"t", "t 0", "t x 7 t:1", "t 0 y t:1", "t 0 0 t:1", "t 0 7", "t 0 7 x:1", "t 0 7 t:",
		"b", "b 0", "b 0 1,2", "b x 1 9", "b 0 y 9", "b 0 1 z", "b 0 1 0",
		"tr", "tr z",
	} {
		assert.NotEmpty(t, c.Exec(input), "input %q", input)
	}
	zi.AssertNotCalled(t, "Take", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	zi.AssertNotCalled(t, "Build", mock.Anything, mock.Anything, mock.Anything)
	zi.AssertNotCalled(t, "Trail", mock.Anything)

	assert.NotEmpty(t, c.Exec("tak"), "a near miss should suggest")
	assert.NotEmpty(t, c.Exec("zzzz"))
}
