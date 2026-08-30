//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func TestDurakWebController_Method(t *testing.T) {
	mockOutput := `{"test":"ok"}`
	diMock := new(usecase.MockDurakInteractor)
	diMock.On("Reset").Return(mockOutput)
	diMock.On("Attack", mock.Anything).Return(mockOutput)
	diMock.On("Defend", mock.Anything, mock.Anything).Return(mockOutput)
	diMock.On("Pass").Return(mockOutput)
	diMock.On("TakeCards").Return(mockOutput)
	diMock.On("Transfer", mock.Anything).Return(mockOutput)
	diMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	diMock.On("Sort", mock.Anything).Return(mockOutput)
	diMock.On("Hint").Return(mockOutput)
	diMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.DurakInteractorIF { return diMock }
	tdwc := controller.NewDurakWebController(factory)
	defer tdwc.Stop()

	var jsonInput controller.DurakWebInput

	t.Run("success Exec reset", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	// The Web CLI sends "hint"/"h" to this endpoint. Before #5791 the default branch
	// used dispatchLog, which does not match either, so the request 400'd with
	// "Unsupported command." even though the interactor had Hint() all along.
	for _, cmd := range []string{"h", "hint"} {
		t.Run("hint via "+cmd, func(t *testing.T) {
			_ = json.Unmarshal([]byte(`{"command":"`+cmd+`","sessionId":"s1"}`), &jsonInput)
			recorded := execRequest(t, tdwc.Exec, &jsonInput)
			recorded.CodeIs(http.StatusOK)
			recorded.BodyIs(mockOutput)
		})
	}

	t.Run("success Exec attack", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"a","sessionId":"s1","cardIdx":0}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec defend", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"s1","attackIdx":0,"cardIdx":1}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec pass", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec take", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"t","sessionId":"s1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec transfer", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"transfer","sessionId":"s1","cardIdx":2}`), &jsonInput)
		defer diMock.AssertCalled(t, "Transfer", 2)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec sort", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"sort","sessionId":"s1","sortMode":1}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec reset with config", func(t *testing.T) {
		input := controller.DurakWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "s1"},
			Config:       &controller.DurakWebConfig{PlayerCount: 4, CpuDifficulty: 1},
		}
		recorded := execRequest(t, tdwc.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})
}
