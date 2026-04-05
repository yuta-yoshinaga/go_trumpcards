package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"

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
	t.Run("success Exec action", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"action","sessionId":"test-session-1"}`), &jsonInput)
		recorded := execRequest(t, towc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})
	t.Run("success Exec a", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command":"a","sessionId":"test-session-1"}`), &jsonInput)
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
