package controller_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func TestDaifugoWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableCards":[],"lastPlayPlayerIdx":-1,"gameEndFlag":false,"revolutionActive":false,"elevenBackActive":false,"suitLocked":false,"lockedSuit":"","tableIsSequence":false,"config":{"jokerCount":0,"eightCutEnabled":false,"suitLockMode":0,"elevenBackEnabled":false,"sequenceEnabled":false,"cardExchangeEnabled":false,"blindExchangeEnabled":false,"fiveSkipEnabled":false,"fiveSkipCount":0,"sevenPassEnabled":false,"tenDiscardEnabled":false,"spadeThreeEnabled":false,"capitalFallEnabled":false,"nineReverseEnabled":false,"coupDetatEnabled":false,"numberLockEnabled":false,"sandstormEnabled":false,"emperorEnabled":false,"sequenceRevolutionEnabled":false,"sequenceLockEnabled":false,"illegalFinishEnabled":false,"queenBomberEnabled":false,"cpuDifficulty":0},"exchangeActions":[],"cpuActions":[],"humanAction":null,"message":""}`
	dgiMock := new(usecase.MockDaifugoInteractor)
	dgiMock.On("Reset").Return(mockOutput).Times(2)
	dgiMock.On("Play", []int{}).Return(mockOutput)
	dgiMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	dgiMock.On("Sort", mock.Anything).Return(mockOutput)

	factory := func() uc.DaifugoInteractorIF { return dgiMock }
	tdwc := controller.NewDaifugoWebController(factory)
	defer tdwc.Stop()

	var jsonInput controller.DaifugoWebInput
	// For "q"/"quit": responseStr = {"message":"bye."} → other fields get zero values
	qBody := `{"players":[],"currentTurn":0,"tableCards":[],"lastPlayPlayerIdx":0,"gameEndFlag":false,"revolutionActive":false,"elevenBackActive":false,"suitLocked":false,"lockedSuit":"","tableIsSequence":false,"config":{"jokerCount":0,"eightCutEnabled":false,"suitLockMode":0,"elevenBackEnabled":false,"sequenceEnabled":false,"cardExchangeEnabled":false,"blindExchangeEnabled":false,"fiveSkipEnabled":false,"fiveSkipCount":0,"sevenPassEnabled":false,"tenDiscardEnabled":false,"spadeThreeEnabled":false,"capitalFallEnabled":false,"nineReverseEnabled":false,"coupDetatEnabled":false,"numberLockEnabled":false,"sandstormEnabled":false,"emperorEnabled":false,"sequenceRevolutionEnabled":false,"sequenceLockEnabled":false,"illegalFinishEnabled":false,"queenBomberEnabled":false,"cpuDifficulty":0},"exchangeActions":[],"cpuActions":[],"humanAction":null,"pendingAction":"none","pendingActionTarget":-1,"reverseDirection":false,"numberLocked":false,"sequenceLocked":false,"sortMode":0,"playableCardIndices":null,"message":"bye."}`

	t.Run("success Exec q", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "q", "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})

	t.Run("success Exec quit", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "quit", "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})

	t.Run("success Exec r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "r", "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec p (pass)", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "p", "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec reset with config calls ResetWithConfig", func(t *testing.T) {
		input := controller.DaifugoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "test-session-1"},
			Config: &controller.DaifugoWebConfig{
				JokerCount:      2,
				FiveSkipEnabled: true,
			},
		}
		recorded := execRequest(t, tdwc.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
		dgiMock.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("success Exec reset with sandstorm and emperor config calls ResetWithConfig", func(t *testing.T) {
		input := controller.DaifugoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "test-session-1"},
			Config: &controller.DaifugoWebConfig{
				SandstormEnabled: true,
				EmperorEnabled:   true,
			},
		}
		recorded := execRequest(t, tdwc.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
		dgiMock.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("success Exec reset with queenBomberEnabled config calls ResetWithConfig", func(t *testing.T) {
		input := controller.DaifugoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "test-session-1"},
			Config: &controller.DaifugoWebConfig{
				QueenBomberEnabled: true,
			},
		}
		recorded := execRequest(t, tdwc.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
		dgiMock.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("success Exec sort default mode", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "sort", "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec reset with sequenceRevolution and illegalFinish config calls ResetWithConfig", func(t *testing.T) {
		input := controller.DaifugoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "test-session-1"},
			Config: &controller.DaifugoWebConfig{
				SequenceRevolutionEnabled: true,
				IllegalFinishEnabled:      true,
			},
		}
		recorded := execRequest(t, tdwc.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
		dgiMock.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("success Exec reset with cpuDifficulty config calls ResetWithConfig", func(t *testing.T) {
		input := controller.DaifugoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "test-session-1"},
			Config: &controller.DaifugoWebConfig{
				CpuDifficulty: 2,
			},
		}
		recorded := execRequest(t, tdwc.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
		dgiMock.AssertCalled(t, "ResetWithConfig", mock.Anything)
	})

	t.Run("success Exec sort with sortMode", func(t *testing.T) {
		input := controller.DaifugoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "sort", SessionID: "test-session-1"},
			SortMode:     func() *int { v := 2; return &v }(),
		}
		recorded := execRequest(t, tdwc.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})

	t.Run("failed Exec other", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "other", "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableCards":[],"lastPlayPlayerIdx":0,"gameEndFlag":false,"revolutionActive":false,"elevenBackActive":false,"suitLocked":false,"lockedSuit":"","tableIsSequence":false,"config":{"jokerCount":0,"eightCutEnabled":false,"suitLockMode":0,"elevenBackEnabled":false,"sequenceEnabled":false,"cardExchangeEnabled":false,"blindExchangeEnabled":false,"fiveSkipEnabled":false,"fiveSkipCount":0,"sevenPassEnabled":false,"tenDiscardEnabled":false,"spadeThreeEnabled":false,"capitalFallEnabled":false,"nineReverseEnabled":false,"coupDetatEnabled":false,"numberLockEnabled":false,"sandstormEnabled":false,"emperorEnabled":false,"sequenceRevolutionEnabled":false,"sequenceLockEnabled":false,"illegalFinishEnabled":false,"queenBomberEnabled":false,"cpuDifficulty":0},"exchangeActions":[],"cpuActions":[],"humanAction":null,"pendingAction":"none","pendingActionTarget":-1,"reverseDirection":false,"numberLocked":false,"sequenceLocked":false,"sortMode":0,"playableCardIndices":null,"message":"Unsupported command."}`)
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "", "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableCards":[],"lastPlayPlayerIdx":0,"gameEndFlag":false,"revolutionActive":false,"elevenBackActive":false,"suitLocked":false,"lockedSuit":"","tableIsSequence":false,"config":{"jokerCount":0,"eightCutEnabled":false,"suitLockMode":0,"elevenBackEnabled":false,"sequenceEnabled":false,"cardExchangeEnabled":false,"blindExchangeEnabled":false,"fiveSkipEnabled":false,"fiveSkipCount":0,"sevenPassEnabled":false,"tenDiscardEnabled":false,"spadeThreeEnabled":false,"capitalFallEnabled":false,"nineReverseEnabled":false,"coupDetatEnabled":false,"numberLockEnabled":false,"sandstormEnabled":false,"emperorEnabled":false,"sequenceRevolutionEnabled":false,"sequenceLockEnabled":false,"illegalFinishEnabled":false,"queenBomberEnabled":false,"cpuDifficulty":0},"exchangeActions":[],"cpuActions":[],"humanAction":null,"pendingAction":"none","pendingActionTarget":-1,"reverseDirection":false,"numberLocked":false,"sequenceLocked":false,"sortMode":0,"playableCardIndices":null,"message":"param error."}`)
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": ""}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableCards":[],"lastPlayPlayerIdx":0,"gameEndFlag":false,"revolutionActive":false,"elevenBackActive":false,"suitLocked":false,"lockedSuit":"","tableIsSequence":false,"config":{"jokerCount":0,"eightCutEnabled":false,"suitLockMode":0,"elevenBackEnabled":false,"sequenceEnabled":false,"cardExchangeEnabled":false,"blindExchangeEnabled":false,"fiveSkipEnabled":false,"fiveSkipCount":0,"sevenPassEnabled":false,"tenDiscardEnabled":false,"spadeThreeEnabled":false,"capitalFallEnabled":false,"nineReverseEnabled":false,"coupDetatEnabled":false,"numberLockEnabled":false,"sandstormEnabled":false,"emperorEnabled":false,"sequenceRevolutionEnabled":false,"sequenceLockEnabled":false,"illegalFinishEnabled":false,"queenBomberEnabled":false,"cpuDifficulty":0},"exchangeActions":[],"cpuActions":[],"humanAction":null,"pendingAction":"none","pendingActionTarget":-1,"reverseDirection":false,"numberLocked":false,"sequenceLocked":false,"sortMode":0,"playableCardIndices":null,"message":"param error."}`)
	})
	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.DaifugoWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, tdwc.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableCards":[],"lastPlayPlayerIdx":0,"gameEndFlag":false,"revolutionActive":false,"elevenBackActive":false,"suitLocked":false,"lockedSuit":"","tableIsSequence":false,"config":{"jokerCount":0,"eightCutEnabled":false,"suitLockMode":0,"elevenBackEnabled":false,"sequenceEnabled":false,"cardExchangeEnabled":false,"blindExchangeEnabled":false,"fiveSkipEnabled":false,"fiveSkipCount":0,"sevenPassEnabled":false,"tenDiscardEnabled":false,"spadeThreeEnabled":false,"capitalFallEnabled":false,"nineReverseEnabled":false,"coupDetatEnabled":false,"numberLockEnabled":false,"sandstormEnabled":false,"emperorEnabled":false,"sequenceRevolutionEnabled":false,"sequenceLockEnabled":false,"illegalFinishEnabled":false,"queenBomberEnabled":false,"cpuDifficulty":0},"exchangeActions":[],"cpuActions":[],"humanAction":null,"pendingAction":"none","pendingActionTarget":-1,"reverseDirection":false,"numberLocked":false,"sequenceLocked":false,"sortMode":0,"playableCardIndices":null,"message":"param error."}`)
	})

}

func TestDaifugoWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableCards":[],"lastPlayPlayerIdx":-1,"gameEndFlag":false,"revolutionActive":false,"elevenBackActive":false,"suitLocked":false,"lockedSuit":"","tableIsSequence":false,"config":{"jokerCount":0,"eightCutEnabled":false,"suitLockMode":0,"elevenBackEnabled":false,"sequenceEnabled":false,"cardExchangeEnabled":false,"blindExchangeEnabled":false,"fiveSkipEnabled":false,"fiveSkipCount":0,"sevenPassEnabled":false,"tenDiscardEnabled":false,"spadeThreeEnabled":false,"capitalFallEnabled":false,"nineReverseEnabled":false,"coupDetatEnabled":false,"numberLockEnabled":false,"sandstormEnabled":false,"emperorEnabled":false,"sequenceRevolutionEnabled":false,"sequenceLockEnabled":false,"illegalFinishEnabled":false,"queenBomberEnabled":false,"cpuDifficulty":0},"exchangeActions":[],"cpuActions":[],"humanAction":null,"message":""}`
	mockA := new(usecase.MockDaifugoInteractor)
	mockA.On("Reset").Return(mockOutput)
	mockB := new(usecase.MockDaifugoInteractor)
	mockB.On("Reset").Return(mockOutput)

	callCount := 0
	isoController := controller.NewDaifugoWebController(func() uc.DaifugoInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.DaifugoWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "Reset")
		mockB.AssertNotCalled(t, "Reset")
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.DaifugoWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-B"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "Reset")
	})

	t.Run("session-A second call reuses mockA", func(t *testing.T) {
		var input controller.DaifugoWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

func TestDaifugoWebController_Log(t *testing.T) {
	mockLogOutput := `{"entries":[]}`
	dgiMock := new(usecase.MockDaifugoInteractor)
	dgiMock.On("ActionLog").Return(mockLogOutput)

	factory := func() uc.DaifugoInteractorIF { return dgiMock }
	ctrl := controller.NewDaifugoWebController(factory)
	defer ctrl.Stop()

	t.Run("log command", func(t *testing.T) {
		var input controller.DaifugoWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"dg-log-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockLogOutput)
	})

	t.Run("l shorthand", func(t *testing.T) {
		var input controller.DaifugoWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"dg-log-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockLogOutput)
	})
}

func TestDaifugoWebController_Stop(t *testing.T) {
	dgiMock := new(usecase.MockDaifugoInteractor)
	factory := func() uc.DaifugoInteractorIF { return dgiMock }
	c := controller.NewDaifugoWebController(factory)
	c.Stop()
	c.Stop()
}
