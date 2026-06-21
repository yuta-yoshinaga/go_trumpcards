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

func mustSpoonsOutputJSON(msg string) string {
	out := &controller.SpoonsWebOutput{
		WinnerIdx:       -1,
		FirstGrabberIdx: -1,
		RoundLoserIdx:   -1,
		CpuDifficulty:   int(domain.SpoonsCpuNormal),
		Players:         make([]*controller.SpoonsWebPlayer, 0),
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, _ := json.Marshal(out)
	return string(b)
}

func TestSpoonsWebController(t *testing.T) {
	mockOutput := `{"phase":0,"winnerIdx":-1,"players":[]}`

	siMock := new(usecase.MockSpoonsInteractor)
	siMock.On("Reset").Return(mockOutput)
	siMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	siMock.On("Pass", mock.Anything).Return(mockOutput)
	siMock.On("Grab").Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("ActionLog").Return(`{"log":[]}`)
	siMock.On("GetConfig").Return(domain.DefaultSpoonsConfig())

	factory := func() uc.SpoonsInteractorIF { return siMock }
	swc := controller.NewSpoonsWebController(factory)
	defer swc.Stop()

	cases := []struct {
		name string
		body string
		want int
	}{
		{"reset", `{"command":"r","sessionId":"s1"}`, http.StatusOK},
		{"reset word", `{"command":"reset","sessionId":"s1"}`, http.StatusOK},
		{"reset with config", `{"command":"reset","config":{"cpuDifficulty":2},"sessionId":"s1"}`, http.StatusOK},
		{"pass", `{"command":"p","cardIndex":2,"sessionId":"s1"}`, http.StatusOK},
		{"pass word", `{"command":"pass","sessionId":"s1"}`, http.StatusOK},
		{"grab", `{"command":"g","sessionId":"s1"}`, http.StatusOK},
		{"grab word", `{"command":"grab","sessionId":"s1"}`, http.StatusOK},
		{"next", `{"command":"n","sessionId":"s1"}`, http.StatusOK},
		{"next word", `{"command":"next","sessionId":"s1"}`, http.StatusOK},
		{"log", `{"command":"log","sessionId":"s1"}`, http.StatusOK},
		{"log alias", `{"command":"l","sessionId":"s1"}`, http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.SpoonsWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			rec := execRequest(t, swc.Exec, &input)
			rec.CodeIs(tc.want)
			rec.ContentTypeIsJson()
		})
	}

	t.Run("quit", func(t *testing.T) {
		var input controller.SpoonsWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		rec := execRequest(t, swc.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mustSpoonsOutputJSON("bye."))
	})

	t.Run("unknown command returns 400", func(t *testing.T) {
		var input controller.SpoonsWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s1"}`), &input)
		rec := execRequest(t, swc.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
	})

	siMock.AssertCalled(t, "Pass", 2)
}
