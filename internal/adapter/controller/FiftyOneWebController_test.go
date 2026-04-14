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

func TestFiftyOneWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"tableCards":[],"phase":0,"currentTurn":0,"gameEndFlag":false,"winnerIdx":-1,"turnNumber":0,"stopCallerIdx":-1,"lastAction":"","lastHandIdx":-1,"lastTableIdx":-1,"message":"","config":{"cpuDifficulty":1}}`
	fiMock := new(mockusecase.MockFiftyOneInteractor)
	fiMock.On("Reset", mock.Anything).Return(mockOutput)
	fiMock.On("GetConfig").Return(domain.DefaultFiftyOneConfig())
	fiMock.On("ExchangeOne", 0, 1).Return(mockOutput)
	fiMock.On("ExchangeAll").Return(mockOutput)
	fiMock.On("Stop").Return(mockOutput)
	fiMock.On("ActionLog").Return(`{"entries":[]}`)

	factory := func() uc.FiftyOneInteractorIF { return fiMock }
	ctrl := controller.NewFiftyOneWebController(factory)
	defer ctrl.Stop()

	t.Run("success reset", func(t *testing.T) {
		var input controller.FiftyOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"test-fo"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mockOutput)
	})

	t.Run("success play", func(t *testing.T) {
		var input controller.FiftyOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","handIdx":0,"tableIdx":1,"sessionId":"test-fo"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play without params returns error", func(t *testing.T) {
		var input controller.FiftyOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","sessionId":"test-fo"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("success exchangeall", func(t *testing.T) {
		var input controller.FiftyOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"exchangeall","sessionId":"test-fo"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success stop", func(t *testing.T) {
		var input controller.FiftyOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"stop","sessionId":"test-fo"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("action log", func(t *testing.T) {
		var input controller.FiftyOneWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"test-fo"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(`{"entries":[]}`)
	})
}
