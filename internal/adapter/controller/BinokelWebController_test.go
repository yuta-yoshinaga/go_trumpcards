//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func TestBinokelWebController_Commands(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	piMock := new(usecase.MockBinokelInteractor)
	piMock.On("ResetWithConfig", domain.DefaultBinokelConfig()).Return(mockOutput)
	piMock.On("Bid", 150).Return(mockOutput)
	piMock.On("Pass").Return(mockOutput)
	piMock.On("DiscardToDabb", []int{0, 1, 2}).Return(mockOutput)
	piMock.On("CallTrump", 3).Return(mockOutput)
	piMock.On("ConfirmMelds").Return(mockOutput)
	piMock.On("Play", 2).Return(mockOutput)
	piMock.On("NextTrick").Return(mockOutput)
	piMock.On("NextRound").Return(mockOutput)
	piMock.On("Hint").Return(mockOutput)
	piMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.BinokelInteractorIF { return piMock }
	ctrl := controller.NewBinokelWebController(factory)
	defer ctrl.Stop()

	tests := []struct {
		name     string
		body     string
		wantCode int
	}{
		{"quit", `{"command":"q","sessionId":"s1"}`, http.StatusOK},
		{"reset", `{"command":"r","sessionId":"s1"}`, http.StatusOK},
		{"bid", `{"command":"b","bidAmount":150,"sessionId":"s1"}`, http.StatusOK},
		{"bid missing amount", `{"command":"b","sessionId":"s1"}`, http.StatusBadRequest},
		{"pass", `{"command":"pa","sessionId":"s1"}`, http.StatusOK},
		{"discard", `{"command":"d","discardIndices":[0,1,2],"sessionId":"s1"}`, http.StatusOK},
		{"discard cardIndices alias", `{"command":"discard","cardIndices":[0,1,2],"sessionId":"s1"}`, http.StatusOK},
		{"discard missing indices", `{"command":"d","sessionId":"s1"}`, http.StatusBadRequest},
		{"discard wrong count", `{"command":"d","discardIndices":[0,1],"sessionId":"s1"}`, http.StatusBadRequest},
		{"trump", `{"command":"t","suit":3,"sessionId":"s1"}`, http.StatusOK},
		{"trump missing suit", `{"command":"t","sessionId":"s1"}`, http.StatusBadRequest},
		{"meld", `{"command":"m","sessionId":"s1"}`, http.StatusOK},
		{"play", `{"command":"p","cardIndex":2,"sessionId":"s1"}`, http.StatusOK},
		{"play missing cardIndex", `{"command":"p","sessionId":"s1"}`, http.StatusBadRequest},
		{"next", `{"command":"n","sessionId":"s1"}`, http.StatusOK},
		{"nextround", `{"command":"nr","sessionId":"s1"}`, http.StatusOK},
		{"hint", `{"command":"h","sessionId":"s1"}`, http.StatusOK},
		{"log", `{"command":"l","sessionId":"s1"}`, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input controller.BinokelWebInput
			_ = json.Unmarshal([]byte(tt.body), &input)
			recorded := execRequest(t, ctrl.Exec, &input)
			recorded.CodeIs(tt.wantCode)
		})
	}
}

func TestBinokelWebConfig_ToConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		input := controller.BinokelWebInput{}
		cfg := input.ToConfig()
		expected := domain.DefaultBinokelConfig()
		if cfg.CpuDifficulty != expected.CpuDifficulty {
			t.Errorf("expected cpuDifficulty %d, got %d", expected.CpuDifficulty, cfg.CpuDifficulty)
		}
		if cfg.PointLimit != expected.PointLimit {
			t.Errorf("expected pointLimit %d, got %d", expected.PointLimit, cfg.PointLimit)
		}
	})

	t.Run("custom config", func(t *testing.T) {
		diff := 2
		limit := 3000
		input := controller.BinokelWebInput{
			Config: &controller.BinokelWebConfig{
				CpuDifficulty: &diff,
				PointLimit:    &limit,
			},
		}
		cfg := input.ToConfig()
		if int(cfg.CpuDifficulty) != 2 {
			t.Errorf("expected cpuDifficulty 2, got %d", cfg.CpuDifficulty)
		}
		if cfg.PointLimit != 3000 {
			t.Errorf("expected pointLimit 3000, got %d", cfg.PointLimit)
		}
	})
}
