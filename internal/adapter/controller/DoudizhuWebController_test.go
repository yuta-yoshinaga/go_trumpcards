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

func TestDoudizhuWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":"play","currentTurn":0,"tableCards":[],"tableCombo":"","kittyCards":[],"landlordIdx":-1,"baseBid":0,"highestBid":0,"bombCount":0,"scores":[0,0,0],"gameEndFlag":false,"config":{"cpuDifficulty":0},"cpuActions":[],"humanAction":null,"message":""}`
	dgiMock := new(usecase.MockDoudizhuInteractor)
	dgiMock.On("Reset").Return(mockOutput)
	dgiMock.On("Bid", mock.Anything).Return(mockOutput)
	dgiMock.On("Play", mock.Anything).Return(mockOutput)
	dgiMock.On("ResetWithConfig", mock.Anything).Return(mockOutput)
	dgiMock.On("ActionLog").Return(`{"entries":[]}`)

	factory := func() uc.DoudizhuInteractorIF { return dgiMock }
	tdwc := controller.NewDoudizhuWebController(factory)
	defer tdwc.Stop()

	var jsonInput controller.DoudizhuWebInput

	t.Run("success Exec reset", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "reset", "sessionId": "s1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec reset with config", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "r", "config": {"cpuDifficulty": 1}, "sessionId": "s1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec bid", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "bid", "bidValue": 2, "sessionId": "s1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec bid without value defaults to pass", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "bid", "sessionId": "s1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec play", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "p", "indices": [0, 1], "sessionId": "s1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec play pass (nil indices)", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "play", "sessionId": "s1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec log", func(t *testing.T) {
		_ = json.Unmarshal([]byte(`{"command": "log", "sessionId": "s1"}`), &jsonInput)
		recorded := execRequest(t, tdwc.Exec, &jsonInput)
		recorded.CodeIs(http.StatusOK)
	})
}

func TestDoudizhuWebConfig_ToConfig(t *testing.T) {
	c := controller.DoudizhuWebConfig{CpuDifficulty: 2}
	cfg := c.ToConfig()
	if int(cfg.CpuDifficulty) != 2 {
		t.Fatalf("CpuDifficulty = %d, want 2", int(cfg.CpuDifficulty))
	}
}
