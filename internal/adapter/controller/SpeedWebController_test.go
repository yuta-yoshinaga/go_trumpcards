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

func mustSpeedOutputJSON(msg string) string {
	out := &controller.SpeedWebOutput{
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, _ := json.Marshal(out)
	return string(b)
}

func TestSpeedWebController(t *testing.T) {
	mockOutput := `{"players":[],"centerPiles":[],"phase":0,"gameEndFlag":false,"winnerIdx":-1,"config":{"cpuDifficulty":1},"message":""}`

	siMock := new(usecase.MockSpeedInteractor)
	siMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	siMock.On("Play", 0, 1).Return(mockOutput)
	siMock.On("Flip").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("ActionLog").Return(`{"log":[]}`)
	siMock.On("GetConfig").Return(domain.DefaultSpeedConfig())

	factory := func() uc.SpeedInteractorIF { return siMock }
	swc := controller.NewSpeedWebController(factory)
	defer swc.Stop()

	t.Run("reset command", func(t *testing.T) {
		var input controller.SpeedWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		rec := execRequest(t, swc.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.ContentTypeIsJson()
	})

	t.Run("reset full word", func(t *testing.T) {
		var input controller.SpeedWebInput
		_ = json.Unmarshal([]byte(`{"command":"reset","sessionId":"s1"}`), &input)
		rec := execRequest(t, swc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("play command", func(t *testing.T) {
		var input controller.SpeedWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","cardIndex":0,"pileIndex":1,"sessionId":"s1"}`), &input)
		rec := execRequest(t, swc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("play missing cardIndex returns 400", func(t *testing.T) {
		var input controller.SpeedWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","pileIndex":1,"sessionId":"s1"}`), &input)
		rec := execRequest(t, swc.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
	})

	t.Run("play missing pileIndex returns 400", func(t *testing.T) {
		var input controller.SpeedWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","cardIndex":0,"sessionId":"s1"}`), &input)
		rec := execRequest(t, swc.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
	})

	t.Run("flip command", func(t *testing.T) {
		var input controller.SpeedWebInput
		_ = json.Unmarshal([]byte(`{"command":"f","sessionId":"s1"}`), &input)
		rec := execRequest(t, swc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("flip full word", func(t *testing.T) {
		var input controller.SpeedWebInput
		_ = json.Unmarshal([]byte(`{"command":"flip","sessionId":"s1"}`), &input)
		rec := execRequest(t, swc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("hint command", func(t *testing.T) {
		var input controller.SpeedWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		rec := execRequest(t, swc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("log command", func(t *testing.T) {
		var input controller.SpeedWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		rec := execRequest(t, swc.Exec, &input)
		rec.CodeIs(http.StatusOK)
	})

	t.Run("quit command", func(t *testing.T) {
		var input controller.SpeedWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		rec := execRequest(t, swc.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mustSpeedOutputJSON("bye."))
	})

	t.Run("unknown command returns 400", func(t *testing.T) {
		var input controller.SpeedWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s1"}`), &input)
		rec := execRequest(t, swc.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
	})
}

func TestSpeedWebInput_ToConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		input := controller.SpeedWebInput{}
		cfg := input.ToConfig()
		assert.Equal(t, domain.SpeedCpuDifficultyNormal, cfg.CpuDifficulty)
	})

	t.Run("with difficulty", func(t *testing.T) {
		d := 2
		input := controller.SpeedWebInput{CpuDifficulty: &d}
		cfg := input.ToConfig()
		assert.Equal(t, domain.SpeedCpuDifficultyHard, cfg.CpuDifficulty)
	})

	t.Run("out of range returns default", func(t *testing.T) {
		d := 5
		input := controller.SpeedWebInput{CpuDifficulty: &d}
		cfg := input.ToConfig()
		assert.Equal(t, domain.SpeedCpuDifficultyNormal, cfg.CpuDifficulty)
	})

	t.Run("negative returns default", func(t *testing.T) {
		d := -1
		input := controller.SpeedWebInput{CpuDifficulty: &d}
		cfg := input.ToConfig()
		assert.Equal(t, domain.SpeedCpuDifficultyNormal, cfg.CpuDifficulty)
	})
}
