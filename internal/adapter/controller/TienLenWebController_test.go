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

func mustTienLenOutputJSON(msg string) string {
	out := &controller.TienLenWebOutput{
		Players:           []*controller.TienLenWebOutputPlayer{},
		TableCards:        []*controller.WebOutputCard{},
		CpuActions:        []*controller.TienLenWebOutputAction{},
		LastPlayPlayerIdx: -1,
		WebOutputBase:     controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTienLenOutputJSON: %v", err))
	}
	return string(b)
}

func TestTienLenWebController_Method(t *testing.T) {
	mockOutput := `{"players":[]}`

	tiMock := new(usecase.MockTienLenInteractor)
	tiMock.On("Reset").Return(mockOutput)
	tiMock.On("Play", []int{0, 1}).Return(mockOutput)
	tiMock.On("Play", []int{}).Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)
	tiMock.On("Hint").Return(mockOutput)

	factory := func() uc.TienLenInteractorIF { return tiMock }
	ctrl := controller.NewTienLenWebController(factory)
	defer ctrl.Stop()

	t.Run("reset", func(t *testing.T) {
		for _, cmd := range []string{"r", "reset"} {
			var input controller.TienLenWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	// #5624: interactor に Hint() が生えたので Web も dispatch する。構造ガード
	// (TestWebControllersDispatchHintWhenTheInteractorHasIt) はソースを grep する
	// だけなので、**実際にコマンドが解決して結果が返ること**はここで見る。
	t.Run("hint", func(t *testing.T) {
		for _, cmd := range []string{"h", "hint"} {
			var input controller.TienLenWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
		tiMock.AssertCalled(t, "Hint")
	})

	t.Run("play with indices", func(t *testing.T) {
		var input controller.TienLenWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","indices":[0,1],"sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play with nil indices is a pass", func(t *testing.T) {
		var input controller.TienLenWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		for _, cmd := range []string{"log", "l"} {
			var input controller.TienLenWebInput
			_ = json.Unmarshal([]byte(fmt.Sprintf(`{"command":"%s","sessionId":"s1"}`, cmd)), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		}
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.TienLenWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustTienLenOutputJSON("Unsupported command."))
	})
}

func TestTienLenWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	expected := domain.TienLenConfig{CpuDifficulty: domain.TienLenDifficultyHard}
	tiMock := new(usecase.MockTienLenInteractor)
	tiMock.On("ResetWithConfig", expected).Return(mockOutput)

	factory := func() uc.TienLenInteractorIF { return tiMock }
	ctrl := controller.NewTienLenWebController(factory)
	defer ctrl.Stop()

	diff := 2
	input := controller.TienLenWebInput{
		BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "cfg-1"},
		Config:       &controller.TienLenWebConfig{CpuDifficulty: diff},
	}
	recorded := execRequest(t, ctrl.Exec, &input)
	recorded.CodeIs(http.StatusOK)
	tiMock.AssertCalled(t, "ResetWithConfig", expected)
}
