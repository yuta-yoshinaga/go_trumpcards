package controller_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/stretchr/testify/mock"
)

func TestSevensWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":0,"jokerCount":0,"cpuStrategy":0,"maxPasses":5,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
	expectedBody := mockOutput
	sgiMock := new(usecase.MockSevensInteractor)
	sgiMock.On("Reset").Return(mockOutput).Times(2)
	sgiMock.On("Play", -1).Return(mockOutput) // pass
	sgiMock.On("Play", 0).Return(mockOutput)
	sgiMock.On("PlayJoker", 0, 1, 6).Return(mockOutput)

	factory := func() uc.SevensInteractorIF { return sgiMock }
	tswc := controller.NewSevensWebController(factory)
	defer tswc.Stop()

	var jsonInput controller.SevensWebInput
	// For "q"/"quit": responseStr = {"message":"bye."} → other fields get zero values
	qBody := `{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":0,"jokerCount":0,"cpuStrategy":0,"maxPasses":0,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"bye."}`

	t.Run("success Exec q", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "q", "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})

	t.Run("success Exec quit", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "quit", "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})

	t.Run("success Exec r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "r", "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec reset", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec p pass (no index, defaults to 0)", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "p", "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec p with index", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "p", "index": 0, "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("success Exec j joker command", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "j", "index": 0, "jokerTargetSuit": 1, "jokerTargetValue": 6, "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(expectedBody)
	})

	t.Run("failed Exec other", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "other", "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":0,"jokerCount":0,"cpuStrategy":0,"maxPasses":0,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"Unsupported command."}`)
	})

	t.Run("failed Exec command empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "", "sessionId": "test-session-1"}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":0,"jokerCount":0,"cpuStrategy":0,"maxPasses":0,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"param error."}`)
	})

	t.Run("failed Exec sessionId empty", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": ""}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":0,"jokerCount":0,"cpuStrategy":0,"maxPasses":0,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"param error."}`)
	})
	t.Run("failed Exec sessionId too long", func(t *testing.T) {
		input := controller.SevensWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, tswc.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`{"players":[],"currentTurn":0,"tableMinVals":[0,0,0,0,0],"tableMaxVals":[0,0,0,0,0],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":0,"jokerCount":0,"cpuStrategy":0,"maxPasses":0,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":"param error."}`)
	})

}

func TestSevensWebController_ResetWithConfig(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":true,"tunnelSkipWidth":0,"jokerCount":2,"cpuStrategy":1,"maxPasses":5,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`

	t.Run("reset with all config fields calls ResetWithConfig", func(t *testing.T) {
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()

		var jsonInput controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "tunnelEnabled": true, "jokerCount": 2, "cpuStrategy": 1, "sessionId": "test-cfg-1"}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
		sgiMock.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{TunnelEnabled: true, JokerCount: 2, CpuStrategy: domain.SevensCpuStrategic, MaxPasses: 5})
		sgiMock.AssertNotCalled(t, "Reset")
	})

	t.Run("reset without config fields calls Reset", func(t *testing.T) {
		defaultOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":0,"jokerCount":0,"cpuStrategy":0,"maxPasses":5,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("Reset").Return(defaultOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()

		var jsonInput controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "test-cfg-2"}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(defaultOutput)
		sgiMock.AssertCalled(t, "Reset")
		sgiMock.AssertNotCalled(t, "ResetWithConfig")
	})

	t.Run("reset with partial config calls ResetWithConfig with defaults", func(t *testing.T) {
		partialOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":true,"tunnelSkipWidth":0,"jokerCount":0,"cpuStrategy":0,"maxPasses":5,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("ResetWithConfig", mock.Anything).Return(partialOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()

		var jsonInput controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "tunnelEnabled": true, "sessionId": "test-cfg-3"}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(partialOutput)
		sgiMock.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{TunnelEnabled: true, MaxPasses: 5})
	})

	t.Run("reset with maxPasses field calls ResetWithConfig", func(t *testing.T) {
		passesOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":0,"jokerCount":0,"cpuStrategy":0,"maxPasses":3,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("ResetWithConfig", mock.Anything).Return(passesOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()

		var jsonInput controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "maxPasses": 3, "sessionId": "test-cfg-4"}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(passesOutput)
		sgiMock.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{MaxPasses: 3})
		sgiMock.AssertNotCalled(t, "Reset")
	})

	t.Run("reset with negative jokerCount passes through to domain SetConfig", func(t *testing.T) {
		passesOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":0,"jokerCount":0,"cpuStrategy":0,"maxPasses":5,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("ResetWithConfig", mock.Anything).Return(passesOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()

		negOne := -1
		input := controller.SevensWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "test-cfg-neg"},
			JokerCount:   &negOne,
		}
		recorded := execRequest(t, tswc.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		sgiMock.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{JokerCount: -1, MaxPasses: 5})
	})

	t.Run("reset with only maxPasses field calls ResetWithConfig with default maxPasses", func(t *testing.T) {
		passesOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":0,"jokerCount":0,"cpuStrategy":0,"maxPasses":0,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("ResetWithConfig", mock.Anything).Return(passesOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()

		var jsonInput controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "maxPasses": 0, "sessionId": "test-cfg-5"}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(passesOutput)
		sgiMock.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{MaxPasses: 0})
		sgiMock.AssertNotCalled(t, "Reset")
	})

	t.Run("reset with noJokerFinish field calls ResetWithConfig", func(t *testing.T) {
		njfOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":0,"jokerCount":0,"cpuStrategy":0,"maxPasses":5,"noJokerFinish":true,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("ResetWithConfig", mock.Anything).Return(njfOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()

		var jsonInput controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "noJokerFinish": true, "sessionId": "test-cfg-njf"}`), &jsonInput)
		recorded := execRequest(t, tswc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(njfOutput)
		sgiMock.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{MaxPasses: 5, NoJokerFinish: true})
		sgiMock.AssertNotCalled(t, "Reset")
	})

	t.Run("reset with jokerReclaim field calls ResetWithConfig", func(t *testing.T) {
		jrOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":0,"jokerCount":1,"cpuStrategy":0,"maxPasses":5,"noJokerFinish":false,"jokerReclaimEnabled":true,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("ResetWithConfig", mock.Anything).Return(jrOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()

		jokerCount := 1
		jokerReclaim := true
		input := controller.SevensWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "test-cfg-jr"},
			JokerCount:   &jokerCount,
			JokerReclaim: &jokerReclaim,
		}
		recorded := execRequest(t, tswc.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(jrOutput)
		sgiMock.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{JokerCount: 1, JokerReclaimEnabled: true, MaxPasses: 5})
		sgiMock.AssertNotCalled(t, "Reset")
	})

	t.Run("reset with endStop field calls ResetWithConfig", func(t *testing.T) {
		esOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":0,"jokerCount":0,"cpuStrategy":0,"maxPasses":5,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":true,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("ResetWithConfig", mock.Anything).Return(esOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()

		endStop := true
		input := controller.SevensWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "test-cfg-es"},
			EndStop:      &endStop,
		}
		recorded := execRequest(t, tswc.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(esOutput)
		sgiMock.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{EndStopEnabled: true, MaxPasses: 5})
		sgiMock.AssertNotCalled(t, "Reset")
	})

	t.Run("reset with tunnelSkipWidth field calls ResetWithConfig", func(t *testing.T) {
		tswOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":3,"jokerCount":0,"cpuStrategy":0,"maxPasses":5,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("ResetWithConfig", mock.Anything).Return(tswOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()

		skipWidth := 3
		input := controller.SevensWebInput{
			BaseWebInput:    controller.BaseWebInput{Command: "reset", SessionID: "test-cfg-tsw"},
			TunnelSkipWidth: &skipWidth,
		}
		recorded := execRequest(t, tswc.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(tswOutput)
		sgiMock.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{TunnelSkipWidth: 3, MaxPasses: 5})
		sgiMock.AssertNotCalled(t, "Reset")
	})

	t.Run("reset with jokerConsecutiveBanned field calls ResetWithConfig", func(t *testing.T) {
		jcbOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":0,"jokerCount":0,"cpuStrategy":0,"maxPasses":5,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":true},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
		sgiMock := new(usecase.MockSevensInteractor)
		sgiMock.On("ResetWithConfig", mock.Anything).Return(jcbOutput)
		factory := func() uc.SevensInteractorIF { return sgiMock }
		tswc := controller.NewSevensWebController(factory)
		defer tswc.Stop()

		jcb := true
		input := controller.SevensWebInput{
			BaseWebInput:           controller.BaseWebInput{Command: "reset", SessionID: "test-cfg-jcb"},
			JokerConsecutiveBanned: &jcb,
		}
		recorded := execRequest(t, tswc.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(jcbOutput)
		sgiMock.AssertCalled(t, "ResetWithConfig", domain.SevensConfig{JokerConsecutiveBanned: true, MaxPasses: 5})
		sgiMock.AssertNotCalled(t, "Reset")
	})
}

func TestSevensWebController_SessionIsolation(t *testing.T) {
	mockOutput := `{"players":[],"currentTurn":0,"tableMinVals":[0,7,7,7,7],"tableMaxVals":[0,7,7,7,7],"tablePlaced":[0,0,0,0,0],"config":{"tunnelEnabled":false,"tunnelSkipWidth":0,"jokerCount":0,"cpuStrategy":0,"maxPasses":5,"noJokerFinish":false,"jokerReclaimEnabled":false,"endStopEnabled":false,"jokerConsecutiveBanned":false},"gameEndFlag":false,"cpuActions":[],"humanAction":null,"message":""}`
	mockA := new(usecase.MockSevensInteractor)
	mockA.On("Reset").Return(mockOutput)
	mockB := new(usecase.MockSevensInteractor)
	mockB.On("Reset").Return(mockOutput)

	callCount := 0
	isoController := controller.NewSevensWebController(func() uc.SevensInteractorIF {
		callCount++
		if callCount == 1 {
			return mockA
		}
		return mockB
	})
	defer isoController.Stop()

	t.Run("session-A reset calls mockA", func(t *testing.T) {
		var input controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockA.AssertCalled(t, "Reset")
		mockB.AssertNotCalled(t, "Reset")
	})

	t.Run("session-B reset calls mockB", func(t *testing.T) {
		var input controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-B"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		mockB.AssertCalled(t, "Reset")
	})

	t.Run("session-A second call reuses mockA", func(t *testing.T) {
		var input controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "session-A"}`), &input)
		recorded := execRequest(t, isoController.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		if callCount != 2 {
			t.Errorf("expected factory to be called 2 times, got %d", callCount)
		}
	})
}

func TestSevensWebController_Log(t *testing.T) {
	mockLogOutput := `{"entries":[]}`
	sgiMock := new(usecase.MockSevensInteractor)
	sgiMock.On("ActionLog").Return(mockLogOutput)

	factory := func() uc.SevensInteractorIF { return sgiMock }
	ctrl := controller.NewSevensWebController(factory)
	defer ctrl.Stop()

	t.Run("log command", func(t *testing.T) {
		var input controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"sv-log-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockLogOutput)
	})

	t.Run("l shorthand", func(t *testing.T) {
		var input controller.SevensWebInput
		_ = json.Unmarshal([]byte(`{"command":"l","sessionId":"sv-log-1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockLogOutput)
	})
}

func TestSevensWebController_Stop(t *testing.T) {
	sgiMock := new(usecase.MockSevensInteractor)
	factory := func() uc.SevensInteractorIF { return sgiMock }
	c := controller.NewSevensWebController(factory)
	c.Stop()
	c.Stop()
}
