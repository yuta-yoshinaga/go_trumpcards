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

func mustWarOutputJSON(msg string) string {
	out := &controller.WarWebOutput{
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, _ := json.Marshal(out)
	return string(b)
}

func TestWarWebController(t *testing.T) {
	mockOutput := `{"players":[],"phase":0,"gameEndFlag":false,"winnerIdx":-1}`

	wiMock := new(usecase.MockWarInteractor)
	wiMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	wiMock.On("Step").Return(mockOutput)
	wiMock.On("AutoPlay").Return(mockOutput)
	wiMock.On("ActionLog").Return(`{"log":[]}`)
	wiMock.On("GetConfig").Return(domain.DefaultWarConfig())

	factory := func() uc.WarInteractorIF { return wiMock }
	wwc := controller.NewWarWebController(factory)
	defer wwc.Stop()

	t.Run("reset", func(t *testing.T) {
		var input controller.WarWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		rec := execRequest(t, wwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
	})

	t.Run("reset full word", func(t *testing.T) {
		var input controller.WarWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","maxRounds":300,"sessionId":"s1"}`), &input)
		rec := execRequest(t, wwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("step", func(t *testing.T) {
		var input controller.WarWebInput
		_ = json.Unmarshal([]byte(`{"command":"s","sessionId":"s1"}`), &input)
		rec := execRequest(t, wwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("step full word", func(t *testing.T) {
		var input controller.WarWebInput
		_ = json.Unmarshal([]byte(`{"command":"step","sessionId":"s1"}`), &input)
		rec := execRequest(t, wwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("autoplay short", func(t *testing.T) {
		var input controller.WarWebInput
		_ = json.Unmarshal([]byte(`{"command":"a","sessionId":"s1"}`), &input)
		rec := execRequest(t, wwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("autoplay full word", func(t *testing.T) {
		var input controller.WarWebInput
		_ = json.Unmarshal([]byte(`{"command":"autoplay","sessionId":"s1"}`), &input)
		rec := execRequest(t, wwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.WarWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		rec := execRequest(t, wwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("quit", func(t *testing.T) {
		var input controller.WarWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		rec := execRequest(t, wwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mustWarOutputJSON("bye."))
	})

	t.Run("unknown command returns 400", func(t *testing.T) {
		var input controller.WarWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s1"}`), &input)
		rec := execRequest(t, wwc.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestWarWebInput_ToConfig(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		input := controller.WarWebInput{}
		cfg := input.ToConfig()
		assert.Equal(t, domain.WarDefaultMaxRounds, cfg.MaxRounds)
	})
	t.Run("valid value", func(t *testing.T) {
		v := 300
		input := controller.WarWebInput{MaxRounds: &v}
		assert.Equal(t, 300, input.ToConfig().MaxRounds)
	})
	t.Run("out of range returns default", func(t *testing.T) {
		v := 0
		input := controller.WarWebInput{MaxRounds: &v}
		assert.Equal(t, domain.WarDefaultMaxRounds, input.ToConfig().MaxRounds)
	})
}
