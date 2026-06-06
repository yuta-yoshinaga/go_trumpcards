package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func TestTichuWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":"play","currentTurn":0,"scores":[0,0],"gameEndFlag":false,"message":""}`
	tgiMock := new(usecase.MockTichuInteractor)
	tgiMock.On("Reset").Return(mockOutput)
	tgiMock.On("Declare", mock.Anything).Return(mockOutput)
	tgiMock.On("Play", mock.Anything).Return(mockOutput)
	tgiMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	tgiMock.On("ActionLog").Return(`{"entries":[]}`)

	factory := func() uc.TichuInteractorIF { return tgiMock }
	twc := controller.NewTichuWebController(factory)
	defer twc.Stop()

	var jsonInput controller.TichuWebInput

	cases := []struct {
		name string
		body string
	}{
		{"reset", `{"command":"reset","sessionId":"s1"}`},
		{"reset with config", `{"command":"r","config":{"cpuDifficulty":1},"sessionId":"s1"}`},
		{"declare", `{"command":"declare","declType":1,"sessionId":"s1"}`},
		{"declare default", `{"command":"declare","sessionId":"s1"}`},
		{"play", `{"command":"p","indices":[0,1],"sessionId":"s1"}`},
		{"play pass", `{"command":"play","sessionId":"s1"}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_ = json.Unmarshal([]byte(c.body), &jsonInput)
			recorded := execRequest(t, twc.Exec, &jsonInput)
			recorded.CodeIs(http.StatusOK)
		})
	}
}

func TestTichuWebConfig_ToConfig(t *testing.T) {
	c := controller.TichuWebConfig{CpuDifficulty: 2}
	cfg := c.ToConfig()
	if int(cfg.CpuDifficulty) != 2 {
		t.Errorf("ToConfig difficulty = %d, want 2", cfg.CpuDifficulty)
	}
}
