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

func mustEgyptianRatscrewOutputJSON(msg string) string {
	out := &controller.EgyptianRatscrewWebOutput{
		WinnerIdx:     -1,
		ChanceFromIdx: -1,
		CpuDifficulty: int(domain.EgyptianRatscrewCpuNormal),
		Players:       make([]*controller.EgyptianRatscrewWebPlayer, 0),
		// 回数は盤面が無くても規則なので、既定の応答にも乗る (#5580)。
		FaceChances: &controller.EgyptianRatscrewWebFaceChances{
			Jack:  domain.FaceCardChances(domain.EgyptianRatscrewJackValue),
			Queen: domain.FaceCardChances(domain.EgyptianRatscrewQueenValue),
			King:  domain.FaceCardChances(domain.EgyptianRatscrewKingValue),
			Ace:   domain.FaceCardChances(domain.EgyptianRatscrewAceValue),
		},
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, _ := json.Marshal(out)
	return string(b)
}

func TestEgyptianRatscrewWebController(t *testing.T) {
	mockOutput := `{"phase":0,"winnerIdx":-1,"players":[]}`

	eiMock := new(usecase.MockEgyptianRatscrewInteractor)
	eiMock.On("Reset").Return(mockOutput)
	eiMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	eiMock.On("Step").Return(mockOutput)
	eiMock.On("Slap", mock.Anything).Return(mockOutput)
	eiMock.On("Tick").Return(mockOutput)
	eiMock.On("ActionLog").Return(`{"log":[]}`)
	eiMock.On("GetConfig").Return(domain.DefaultEgyptianRatscrewConfig())

	factory := func() uc.EgyptianRatscrewInteractorIF { return eiMock }
	ewc := controller.NewEgyptianRatscrewWebController(factory)
	defer ewc.Stop()

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
		{"slap alias", `{"command":"j","sessionId":"s1"}`, http.StatusOK},
		{"slap full word", `{"command":"slap","sessionId":"s1"}`, http.StatusOK},
		{"slap ignores client playerIdx", `{"command":"slap","playerIdx":1,"sessionId":"s1"}`, http.StatusOK},
		{"tick", `{"command":"tick","sessionId":"s1"}`, http.StatusOK},
		{"log", `{"command":"log","sessionId":"s1"}`, http.StatusOK},
		{"log alias", `{"command":"l","sessionId":"s1"}`, http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var input controller.EgyptianRatscrewWebInput
			_ = json.Unmarshal([]byte(tc.body), &input)
			rec := execRequest(t, ewc.Exec, &input)
			rec.CodeIs(tc.want)
			rec.ContentTypeIsJson()
		})
	}

	t.Run("quit", func(t *testing.T) {
		var input controller.EgyptianRatscrewWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		rec := execRequest(t, ewc.Exec, &input)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mustEgyptianRatscrewOutputJSON("bye."))
	})

	t.Run("unknown command returns 400", func(t *testing.T) {
		var input controller.EgyptianRatscrewWebInput
		_ = json.Unmarshal([]byte(`{"command":"xyz","sessionId":"s1"}`), &input)
		rec := execRequest(t, ewc.Exec, &input)
		rec.CodeIs(http.StatusBadRequest)
	})

	// Slap must always be invoked with playerIdx=0 (human) regardless of client input,
	// to prevent a malicious client from triggering a CPU slap server-side.
	eiMock.AssertCalled(t, "Slap", 0)
	eiMock.AssertNotCalled(t, "Slap", 1)
}
