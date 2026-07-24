//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustZhengOutputJSON(msg string) string {
	out := &controller.ZhengWebOutput{
		Players:           []*controller.ZhengWebOutputPlayer{},
		TableCards:        []*controller.WebOutputCard{},
		CpuActions:        []*controller.ZhengWebOutputAction{},
		LastPlayPlayerIdx: -1,
		WebOutputBase:     controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustZhengOutputJSON: %v", err))
	}
	return string(b)
}

func TestZhengWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`

	ziMock := new(usecase.MockZhengInteractor)
	ziMock.On("Reset").Return(mockOutput)
	ziMock.On("Play", []int{0, 1}).Return(mockOutput)
	ziMock.On("Play", []int{}).Return(mockOutput)
	ziMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.ZhengInteractorIF { return ziMock }
	ctrl := controller.NewZhengWebController(factory)
	defer ctrl.Stop()

	t.Run("reset", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset"} {
			var input controller.ZhengWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("play with indices", func(t *testing.T) {
		var input controller.ZhengWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","indices":[0,1],"sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play with nil indices is a pass", func(t *testing.T) {
		var input controller.ZhengWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		for _, cmd := range []string{"log", "l"} {
			var input controller.ZhengWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.ZhengWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustZhengOutputJSON("Unsupported command."))
	})
}

func TestZhengWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	expected := domain.ZhengConfig{CpuDifficulty: domain.ZhengDifficultyHard}
	ziMock := new(usecase.MockZhengInteractor)
	ziMock.On("ResetWithConfig", expected).Return(mockOutput)

	factory := func() uc.ZhengInteractorIF { return ziMock }
	ctrl := controller.NewZhengWebController(factory)
	defer ctrl.Stop()

	diff := 2
	input := controller.ZhengWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-1"},
		Config:       &controller.ZhengWebConfig{CpuDifficulty: diff},
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	ziMock.AssertCalled(t, "ResetWithConfig", expected)
}
