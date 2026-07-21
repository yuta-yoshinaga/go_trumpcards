package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPigsTailWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"circleCount":52,"centerTop":null,"centerCount":0,"currentTurn":0,"gameEndFlag":false,"loserIdx":-1,"lastDrawCard":null,"lastPenalty":false,"cpuActions":[],"humanAction":null,"message":""}`
	ptiMock := new(usecase.MockPigsTailInteractor)
	ptiMock.On("Reset", mock.Anything).Return(mockOutput)
	ptiMock.On("Action", 0).Return(mockOutput)
	ptiMock.On("ActionLog").Return(`[]`)

	factory := func() uc.PigsTailInteractorIF { return ptiMock }
	towc := controller.NewPigsTailWebController(factory)
	defer towc.Stop()

	qBody := `{"players":[],"circleCount":0,"centerTop":null,"centerCount":0,"currentTurn":0,"gameEndFlag":false,"loserIdx":-1,"lastDrawCard":null,"lastPenalty":false,"cpuActions":[],"humanAction":null,"message":"bye."}`

	var jsonInput controller.PigsTailWebInput

	t.Run("success Exec q", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"test-session-1"}`), &jsonInput)
		recorded := execRequest(t, towc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(qBody)
	})
	t.Run("success Exec reset", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"test-session-1"}`), &jsonInput)
		recorded := execRequest(t, towc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec r", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"test-session-1"}`), &jsonInput)
		recorded := execRequest(t, towc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec draw", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"draw","sessionId":"test-session-1"}`), &jsonInput)
		recorded := execRequest(t, towc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec d", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"d","sessionId":"test-session-1"}`), &jsonInput)
		recorded := execRequest(t, towc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec log", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-session-1"}`), &jsonInput)
		recorded := execRequest(t, towc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(`[]`)
	})
}

func TestPigsTailWebController_ResetAppliesPlayerCount(t *testing.T) {
	mockOutput := `{}`
	tests := []struct {
		name      string
		body      string
		wantCount int
	}{
		{"explicit valid count", `{"command":"reset","playerCount":3,"sessionId":"s"}`, 3},
		{"max valid count", `{"command":"reset","playerCount":6,"sessionId":"s"}`, domain.PigsTailMaxPlayers},
		{"omitted uses default", `{"command":"reset","sessionId":"s"}`, domain.PigsTailPlayerCnt},
		{"above max falls back to default", `{"command":"reset","playerCount":99,"sessionId":"s"}`, domain.PigsTailPlayerCnt},
		{"below min falls back to default", `{"command":"reset","playerCount":1,"sessionId":"s"}`, domain.PigsTailPlayerCnt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptiMock := new(usecase.MockPigsTailInteractor)
			var gotCfg domain.PigsTailConfig
			ptiMock.On("Reset", mock.MatchedBy(func(c domain.PigsTailConfig) bool {
				gotCfg = c
				return true
			})).Return(mockOutput)

			factory := func() uc.PigsTailInteractorIF { return ptiMock }
			towc := controller.NewPigsTailWebController(factory)
			defer towc.Stop()

			var in controller.PigsTailWebInput
			_ = json.Unmarshal([]byte(tt.body), &in)
			recorded := execRequest(t, towc.Exec, &in)
			recorded.CodeIs(http.StatusOK)
			assert.Equal(t, tt.wantCount, gotCfg.PlayerCount)
		})
	}
}
