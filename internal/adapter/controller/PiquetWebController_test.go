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

func TestPiquetWebController_Dispatch(t *testing.T) {
	mockOutput := `{"phase":0}`

	piMock := new(usecase.MockPiquetInteractor)
	piMock.On("Reset").Return(mockOutput)
	piMock.On("ResetWithConfig", domain.DefaultPiquetConfig()).Return(mockOutput)
	piMock.On("ExchangeElder", []int{0, 1, 2}).Return(mockOutput)
	piMock.On("ExchangeYounger", []int{0}).Return(mockOutput)
	piMock.On("ResolveDeclaration").Return(mockOutput)
	piMock.On("Play", 5).Return(mockOutput)
	piMock.On("NextDeal").Return(mockOutput)
	piMock.On("Hint").Return(mockOutput)
	piMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.PiquetInteractorIF { return piMock }
	ctrl := controller.NewPiquetWebController(factory)
	defer ctrl.Stop()

	cases := []struct {
		name  string
		body  string
		want  int
		check string
	}{
		{"quit q", `{"command":"q","sessionId":"s1"}`, http.StatusOK, "bye."},
		{"reset r", `{"command":"r","sessionId":"s1"}`, http.StatusOK, "phase"},
		{"elder e", `{"command":"e","discardIndices":[0,1,2],"sessionId":"s1"}`, http.StatusOK, "phase"},
		{"younger y", `{"command":"y","discardIndices":[0],"sessionId":"s1"}`, http.StatusOK, "phase"},
		{"declare d", `{"command":"d","sessionId":"s1"}`, http.StatusOK, "phase"},
		{"play p", `{"command":"p","cardIndex":5,"sessionId":"s1"}`, http.StatusOK, "phase"},
		{"nextdeal nd", `{"command":"nd","sessionId":"s1"}`, http.StatusOK, "phase"},
		{"hint h", `{"command":"h","sessionId":"s1"}`, http.StatusOK, "phase"},
		{"log l", `{"command":"l","sessionId":"s1"}`, http.StatusOK, "phase"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var input controller.PiquetWebInput
			_ = json.Unmarshal([]byte(tt.body), &input)
			rec := execRequest(t, ctrl.Exec, &input)
			rec.CodeIs(tt.want)
		})
	}
}

func TestPiquetWebController_PlayMissingCardIndex(t *testing.T) {
	piMock := new(usecase.MockPiquetInteractor)
	factory := func() uc.PiquetInteractorIF { return piMock }
	ctrl := controller.NewPiquetWebController(factory)
	defer ctrl.Stop()

	var input controller.PiquetWebInput
	_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1"}`), &input)
	rec := execRequest(t, ctrl.Exec, &input)
	rec.CodeIs(http.StatusBadRequest)
}

func TestPiquetWebController_ElderMissingIndices(t *testing.T) {
	piMock := new(usecase.MockPiquetInteractor)
	factory := func() uc.PiquetInteractorIF { return piMock }
	ctrl := controller.NewPiquetWebController(factory)
	defer ctrl.Stop()

	var input controller.PiquetWebInput
	_ = json.Unmarshal([]byte(`{"command":"e","sessionId":"s1"}`), &input)
	rec := execRequest(t, ctrl.Exec, &input)
	rec.CodeIs(http.StatusBadRequest)
}

func TestPiquetWebConfig_ToConfig_DefaultsAndBounds(t *testing.T) {
	t.Run("nil input returns default", func(t *testing.T) {
		in := controller.PiquetWebInput{}
		cfg := in.ToConfig()
		if cfg.DealsPerPartie != 6 {
			t.Errorf("default DealsPerPartie = %d, want 6", cfg.DealsPerPartie)
		}
	})
	t.Run("out-of-range falls back to default", func(t *testing.T) {
		d := 999
		n := 0
		webCfg := &controller.PiquetWebConfig{CpuDifficulty: &d, DealsPerPartie: &n}
		in := controller.PiquetWebInput{Config: webCfg}
		cfg := in.ToConfig()
		// out-of-range difficulty -> default (Normal)
		if cfg.CpuDifficulty != domain.PiquetCpuDifficultyNormal {
			t.Errorf("out-of-range difficulty did not fall back, got %d", cfg.CpuDifficulty)
		}
		// out-of-range deals -> default (6)
		if cfg.DealsPerPartie != 6 {
			t.Errorf("out-of-range deals did not fall back, got %d", cfg.DealsPerPartie)
		}
	})
	t.Run("in-range values preserved", func(t *testing.T) {
		d := int(domain.PiquetCpuDifficultyHard)
		n := 3
		webCfg := &controller.PiquetWebConfig{CpuDifficulty: &d, DealsPerPartie: &n}
		in := controller.PiquetWebInput{Config: webCfg}
		cfg := in.ToConfig()
		if cfg.CpuDifficulty != domain.PiquetCpuDifficultyHard {
			t.Errorf("in-range difficulty lost, got %d", cfg.CpuDifficulty)
		}
		if cfg.DealsPerPartie != 3 {
			t.Errorf("in-range deals lost, got %d", cfg.DealsPerPartie)
		}
	})
}
