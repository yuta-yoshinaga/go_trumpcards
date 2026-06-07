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

func TestBourreWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":"decide","gameEndFlag":false,"message":""}`
	bgiMock := new(usecase.MockBourreInteractor)
	bgiMock.On("Reset").Return(mockOutput)
	bgiMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	bgiMock.On("Decide", mock.Anything).Return(mockOutput)
	bgiMock.On("Draw", mock.Anything).Return(mockOutput)
	bgiMock.On("Play", mock.Anything).Return(mockOutput)
	bgiMock.On("NextHand").Return(mockOutput)
	bgiMock.On("ActionLog").Return(`{"entries":[]}`)

	factory := func() uc.BourreInteractorIF { return bgiMock }
	bwc := controller.NewBourreWebController(factory)
	defer bwc.Stop()

	var jsonInput controller.BourreWebInput

	cases := []struct {
		name string
		body string
	}{
		{"reset", `{"command":"reset","sessionId":"s1"}`},
		{"reset with config", `{"command":"r","config":{"cpuDifficulty":1},"sessionId":"s1"}`},
		{"decide play", `{"command":"decide","decide":true,"sessionId":"s1"}`},
		{"decide default", `{"command":"decide","sessionId":"s1"}`},
		{"draw", `{"command":"draw","indices":[0,2],"sessionId":"s1"}`},
		{"draw empty", `{"command":"draw","sessionId":"s1"}`},
		{"play", `{"command":"p","cardIndex":1,"sessionId":"s1"}`},
		{"next", `{"command":"next","sessionId":"s1"}`},
		{"log", `{"command":"log","sessionId":"s1"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_ = json.Unmarshal([]byte(c.body), &jsonInput)
			recorded := execRequest(t, bwc.Exec, &jsonInput)
			recorded.CodeIs(http.StatusOK)
		})
	}
}

func TestBourreWebConfig_ToConfig(t *testing.T) {
	c := controller.BourreWebConfig{CpuDifficulty: 2}
	cfg := c.ToConfig()
	if int(cfg.CpuDifficulty) != 2 {
		t.Errorf("ToConfig difficulty = %d, want 2", cfg.CpuDifficulty)
	}
}
