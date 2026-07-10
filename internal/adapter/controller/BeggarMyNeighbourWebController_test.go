//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustBeggarMyNeighbourOutputJSON(msg string) string {
	out := &controller.BeggarMyNeighbourWebOutput{
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, _ := json.Marshal(out)
	return string(b)
}

func TestBeggarMyNeighbourWebController(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"gameEndFlag":false,"winnerIdx":-1}`

	wiMock := new(usecase.MockBeggarMyNeighbourInteractor)
	wiMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	wiMock.On("Step").Return(mockOutput)
	wiMock.On("AutoPlay").Return(mockOutput)
	wiMock.On("ActionLog").Return(`{"log":[]}`)
	wiMock.On("GetConfig").Return(domain.DefaultBeggarMyNeighbourConfig())

	factory := func() uc.BeggarMyNeighbourInteractorIF { return wiMock }
	bwc := controller.NewBeggarMyNeighbourWebController(factory)
	defer bwc.Stop()

	t.Run("reset", func(t *testing.T) {
		var input controller.BeggarMyNeighbourWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		rec := execRequest(t, bwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
	})

	t.Run("reset full word", func(t *testing.T) {
		var input controller.BeggarMyNeighbourWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","maxRounds":1000,"sessionId":"s1"}`), &input)
		rec := execRequest(t, bwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("step", func(t *testing.T) {
		var input controller.BeggarMyNeighbourWebInput
		_ = json.Unmarshal([]byte(`{"command":"s","sessionId":"s1"}`), &input)
		rec := execRequest(t, bwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("step full word", func(t *testing.T) {
		var input controller.BeggarMyNeighbourWebInput
		_ = json.Unmarshal([]byte(`{"command":"step","sessionId":"s1"}`), &input)
		rec := execRequest(t, bwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("autoplay short", func(t *testing.T) {
		var input controller.BeggarMyNeighbourWebInput
		_ = json.Unmarshal([]byte(`{"command":"a","sessionId":"s1"}`), &input)
		rec := execRequest(t, bwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("autoplay full word", func(t *testing.T) {
		var input controller.BeggarMyNeighbourWebInput
		_ = json.Unmarshal([]byte(`{"command":"autoplay","sessionId":"s1"}`), &input)
		rec := execRequest(t, bwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.BeggarMyNeighbourWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		rec := execRequest(t, bwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("quit", func(t *testing.T) {
		var input controller.BeggarMyNeighbourWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		rec := execRequest(t, bwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mustBeggarMyNeighbourOutputJSON("bye."))
	})

	t.Run("unknown command returns 400", func(t *testing.T) {
		var input controller.BeggarMyNeighbourWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s1"}`), &input)
		rec := execRequest(t, bwc.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestBeggarMyNeighbourWebInput_ToConfig(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		input := controller.BeggarMyNeighbourWebInput{}
		cfg := input.ToConfig()
		assert.Equal(t, domain.BeggarMyNeighbourDefaultMaxRounds, cfg.MaxRounds)
	})
	t.Run("valid value", func(t *testing.T) {
		v := 1000
		input := controller.BeggarMyNeighbourWebInput{MaxRounds: &v}
		assert.Equal(t, 1000, input.ToConfig().MaxRounds)
	})
	t.Run("out of range returns default", func(t *testing.T) {
		v := 0
		input := controller.BeggarMyNeighbourWebInput{MaxRounds: &v}
		assert.Equal(t, domain.BeggarMyNeighbourDefaultMaxRounds, input.ToConfig().MaxRounds)
	})
}
