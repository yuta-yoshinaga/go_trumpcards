//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustSlapjackOutputJSON(msg string) string {
	out := &controller.SlapjackWebOutput{
		WinnerIdx:     -1,
		CpuDifficulty: int(domain.SlapjackCpuNormal),
		Players:       make([]*controller.SlapjackWebPlayer, 0),
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, _ := json.Marshal(out)
	return string(b)
}

func TestSlapjackWebController(t *testing.T) {
	mockOutput := `{"phase":0,"winnerIdx":-1,"players":[]}`

	siMock := new(usecase.MockSlapjackInteractor)
	siMock.On("Reset").Return(mockOutput)
	siMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	siMock.On("Step").Return(mockOutput)
	siMock.On("Slap", mock.Anything).Return(mockOutput)
	siMock.On("Tick").Return(mockOutput)
	siMock.On("ActionLog").Return(`{"log":[]}`)
	siMock.On("GetConfig").Return(domain.DefaultSlapjackConfig())

	factory := func() uc.SlapjackInteractorIF { return siMock }
	swc := controller.NewSlapjackWebController(factory)
	defer swc.Stop()

	cases := []struct {
		name string
		body string
		want int
	}{
		{"reset", `{"command":"r","sessionId":"s1"}`, http.StatusOK},
		{"reset full word", `{"command":"reset","sessionId":"s1"}`, http.StatusOK},
		{"reset with config", `{"command":"reset","config":{"cpuDifficulty":2},"sessionId":"s1"}`, http.StatusOK},
		{"step", `{"command":"s","sessionId":"s1"}`, http.StatusOK},
		{"step full word", `{"command":"step","sessionId":"s1"}`, http.StatusOK},
		{"slap default playerIdx", `{"command":"j","sessionId":"s1"}`, http.StatusOK},
		{"slap full word", `{"command":"slap","playerIdx":0,"sessionId":"s1"}`, http.StatusOK},
		{"tick", `{"command":"tick","sessionId":"s1"}`, http.StatusOK},
		{"log", `{"command":"log","sessionId":"s1"}`, http.StatusOK},
		{"log alias", `{"command":"l","sessionId":"s1"}`, http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.SlapjackWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			rec := execRequest(t, swc.Exec, &input)
			rec.CodeIs(tc.want)
			rec.ContentTypeIsJson()
		})
	}

	t.Run("quit", func(t *testing.T) {
		var input controller.SlapjackWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		rec := execRequest(t, swc.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mustSlapjackOutputJSON("bye."))
	})

	t.Run("unknown command returns 400", func(t *testing.T) {
		var input controller.SlapjackWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s1"}`), &input)
		rec := execRequest(t, swc.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
	})
}
