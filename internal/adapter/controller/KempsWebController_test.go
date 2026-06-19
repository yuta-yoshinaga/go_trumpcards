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

func mustKempsOutputJSON(msg string) string {
	out := &controller.KempsWebOutput{
		WinnerTeam:      -1,
		FourHolderIdx:   -1,
		RoundWinnerTeam: -1,
		CpuDifficulty:   int(domain.KempsCpuNormal),
		TargetScore:     domain.KempsTargetScore,
		TeamScores:      make([]int, 0),
		Field:           make([]*controller.WebOutputCard, 0),
		Players:         make([]*controller.KempsWebPlayer, 0),
		WebOutputBase:   controller.WebOutputBase{Message: msg},
	}
	b, _ := json.Marshal(out)
	return string(b)
}

func TestKempsWebController(t *testing.T) {
	mockOutput := `{"phase":0,"winnerTeam":-1,"players":[]}`

	kiMock := new(usecase.MockKempsInteractor)
	kiMock.On("Reset").Return(mockOutput)
	kiMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	kiMock.On("Swap", mock.Anything, mock.Anything).Return(mockOutput)
	kiMock.On("Pass").Return(mockOutput)
	kiMock.On("SetSignal", mock.Anything).Return(mockOutput)
	kiMock.On("DeclareKemps").Return(mockOutput)
	kiMock.On("DeclareCounterKemps", mock.Anything).Return(mockOutput)
	kiMock.On("NextRound").Return(mockOutput)
	kiMock.On("ActionLog").Return(`{"log":[]}`)
	kiMock.On("GetConfig").Return(domain.DefaultKempsConfig())

	factory := func() uc.KempsInteractorIF { return kiMock }
	kwc := controller.NewKempsWebController(factory)
	defer kwc.Stop()

	cases := []struct {
		name string
		body string
		want int
	}{
		{"reset", `{"command":"r","sessionId":"s1"}`, http.StatusOK},
		{"reset word", `{"command":"reset","sessionId":"s1"}`, http.StatusOK},
		{"reset with config", `{"command":"reset","config":{"cpuDifficulty":2,"targetScore":7},"sessionId":"s1"}`, http.StatusOK},
		{"swap", `{"command":"s","handIndex":1,"fieldIndex":2,"sessionId":"s1"}`, http.StatusOK},
		{"swap word", `{"command":"swap","sessionId":"s1"}`, http.StatusOK},
		{"pass", `{"command":"p","sessionId":"s1"}`, http.StatusOK},
		{"signal", `{"command":"sig","signalType":1,"sessionId":"s1"}`, http.StatusOK},
		{"kemps", `{"command":"k","sessionId":"s1"}`, http.StatusOK},
		{"counter", `{"command":"c","targetSeat":1,"sessionId":"s1"}`, http.StatusOK},
		{"next", `{"command":"n","sessionId":"s1"}`, http.StatusOK},
		{"log", `{"command":"log","sessionId":"s1"}`, http.StatusOK},
		{"log alias", `{"command":"l","sessionId":"s1"}`, http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.KempsWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			rec := execRequest(t, kwc.Exec, &input)
			rec.CodeIs(tc.want)
			rec.ContentTypeIsJson()
		})
	}

	t.Run("quit", func(t *testing.T) {
		var input controller.KempsWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		rec := execRequest(t, kwc.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mustKempsOutputJSON("bye."))
	})

	t.Run("unknown command returns 400", func(t *testing.T) {
		var input controller.KempsWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s1"}`), &input)
		rec := execRequest(t, kwc.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
	})

	kiMock.AssertCalled(t, "Swap", 1, 2)
	kiMock.AssertCalled(t, "DeclareCounterKemps", 1)
}
