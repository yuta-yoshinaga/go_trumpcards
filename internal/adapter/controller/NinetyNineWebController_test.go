//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustNinetyNineOutputJSON(msg string) string {
	out := &controller.NinetyNineWebOutput{
		Players:       []*controller.NinetyNineWebOutputPlayer{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		WinnerIdx:     -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustNinetyNineOutputJSON: %v", err))
	}
	return string(b)
}

func TestNinetyNineWebController_Method(t *testing.T) {
	mockOutput := `{"ok":true}`

	oiMock := new(usecase.MockNinetyNineInteractor)
	oiMock.On("ResetWithConfig", domain.DefaultNinetyNineConfig()).Return(mockOutput)
	oiMock.On("Bid", []int{0, 1, 2}).Return(mockOutput)
	oiMock.On("Play", 3).Return(mockOutput)
	oiMock.On("NextTrick").Return(mockOutput)
	oiMock.On("NextRound").Return(mockOutput)
	oiMock.On("Hint").Return(mockOutput)
	oiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.NinetyNineInteractorIF { return oiMock }
	ctrl := controller.NewNinetyNineWebController(factory)
	defer ctrl.Stop()

	t.Run("success Exec q", func(t *testing.T) {
		var input controller.NinetyNineWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.ContentTypeIsJson()
		recorded.BodyIs(mustNinetyNineOutputJSON("bye."))
	})

	t.Run("success Exec r", func(t *testing.T) {
		var input controller.NinetyNineWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec b", func(t *testing.T) {
		var input controller.NinetyNineWebInput
		_ = json.Unmarshal([]byte(`{"command":"b","buryIndices":[0,1,2],"sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("bid missing param", func(t *testing.T) {
		var input controller.NinetyNineWebInput
		_ = json.Unmarshal([]byte(`{"command":"b","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("success Exec p", func(t *testing.T) {
		var input controller.NinetyNineWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","cardIndex":3,"sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("play missing param", func(t *testing.T) {
		var input controller.NinetyNineWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})

	t.Run("success Exec n", func(t *testing.T) {
		var input controller.NinetyNineWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec nr", func(t *testing.T) {
		var input controller.NinetyNineWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec h", func(t *testing.T) {
		var input controller.NinetyNineWebInput
		_ = json.Unmarshal([]byte(`{"command":"h","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("success Exec log", func(t *testing.T) {
		var input controller.NinetyNineWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusOK)
		recorded.BodyIs(mockOutput)
	})

	t.Run("unknown command", func(t *testing.T) {
		var input controller.NinetyNineWebInput
		_ = json.Unmarshal([]byte(`{"command":"xxx","sessionId":"s1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
	})
}

func TestNinetyNineWebConfig_ToConfig(t *testing.T) {
	t.Run("all nil uses defaults", func(t *testing.T) {
		cfg := (&controller.NinetyNineWebConfig{}).ToConfig()
		expected := domain.DefaultNinetyNineConfig()
		if cfg != expected {
			t.Errorf("expected %+v, got %+v", expected, cfg)
		}
	})

	t.Run("values are bounded", func(t *testing.T) {
		v99 := 99
		cfg := (&controller.NinetyNineWebConfig{CpuDifficulty: &v99}).ToConfig()
		if int(cfg.CpuDifficulty) > 2 {
			t.Errorf("expected bounded difficulty, got %d", cfg.CpuDifficulty)
		}
		vBig := 99999
		cfg2 := (&controller.NinetyNineWebConfig{TargetScore: &vBig}).ToConfig()
		if cfg2.TargetScore > 1000 {
			t.Errorf("expected bounded target, got %d", cfg2.TargetScore)
		}
	})

	t.Run("input ToConfig with nil config", func(t *testing.T) {
		in := controller.NinetyNineWebInput{}
		if in.ToConfig() != domain.DefaultNinetyNineConfig() {
			t.Errorf("expected default config")
		}
	})
}
