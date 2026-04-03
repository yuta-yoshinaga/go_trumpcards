//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func TestGoFishWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"currentTurn":0,"gameEndFlag":false,"winnerIdx":-1,"turnNumber":1,"deckRemaining":32,"message":"","config":{"cpuDifficulty":1}}`
	giMock := new(mockusecase.MockGoFishInteractor)
	giMock.On("Reset", mock.Anything).Return(mockOutput)
	giMock.On("GetConfig").Return(domain.DefaultGoFishConfig())
	giMock.On("Ask", 1, 3).Return(mockOutput)
	giMock.On("ActionLog").Return(`{"entries":[]}`)

	factory := func() uc.GoFishInteractorIF { return giMock }
	ctrl := controller.NewGoFishWebController(factory)
	defer ctrl.Stop()

	t.Run("success reset", func(t *testing.T) {
		var input controller.GoFishWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"test-gf"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})

	t.Run("success ask", func(t *testing.T) {
		var input controller.GoFishWebInput
		_ = json.Unmarshal([]byte(`{"command":"ask","targetIdx":1,"rank":3,"sessionId":"test-gf"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})

	t.Run("ask without params returns error", func(t *testing.T) {
		var input controller.GoFishWebInput
		_ = json.Unmarshal([]byte(`{"command":"ask","sessionId":"test-gf"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("action log", func(t *testing.T) {
		var input controller.GoFishWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-gf"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(`{"entries":[]}`)
	})
}
