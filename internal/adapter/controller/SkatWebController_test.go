//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	ctrlusecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustSkatOutputJSON(msg string) string {
	out := &controller.SkatWebOutput{
		Players:           []*controller.SkatWebOutputPlayer{},
		CurrentTrick:      []*controller.WebOutputTrickCard{},
		WinnerSide:        domain.SkatWinnerUndecided,
		DeclarerIdx:       -1,
		ActiveBidActorIdx: -1,
		WebOutputBase:     controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustSkatOutputJSON: %v", err))
	}
	return string(b)
}

func TestSkatWebController_Method(t *testing.T) {
	mockOutput := `{"phase":0}`

	siMock := new(ctrlusecase.MockSkatInteractor)
	siMock.On("ResetWithConfig", domain.DefaultSkatConfig()).Return(mockOutput)
	siMock.On("Bid", true).Return(mockOutput)
	siMock.On("PickSkat", true).Return(mockOutput)
	siMock.On("Discard", 0, 1).Return(mockOutput)
	siMock.On("DeclareGame", domain.SkatGameSuit, 1).Return(mockOutput)
	siMock.On("Play", 2).Return(mockOutput)
	siMock.On("NextTrick").Return(mockOutput)
	siMock.On("NextRound").Return(mockOutput)
	siMock.On("Hint").Return(mockOutput)
	siMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.SkatInteractorIF { return siMock }
	ctrl := controller.NewSkatWebController(factory)
	defer ctrl.Stop()

	t.Run("quit", func(t *testing.T) {
		var input controller.SkatWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"s1"}`), &input)
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mustSkatOutputJSON("bye."))
	})

	t.Run("reset", func(t *testing.T) {
		var input controller.SkatWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"s1"}`), &input)
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mockOutput)
	})

	t.Run("bid missing param", func(t *testing.T) {
		var input controller.SkatWebInput
		_ = json.Unmarshal([]byte(`{"command":"bid","sessionId":"s1"}`), &input)
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusBadRequest)
	})

	t.Run("bid accept", func(t *testing.T) {
		yes := true
		input := controller.SkatWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "b", SessionID: "s1"},
			Accept:       &yes,
		}
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mockOutput)
	})

	t.Run("pickskat", func(t *testing.T) {
		yes := true
		input := controller.SkatWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "ps", SessionID: "s1"},
			Pickup:       &yes,
		}
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mockOutput)
	})

	t.Run("pickskat missing", func(t *testing.T) {
		var input controller.SkatWebInput
		_ = json.Unmarshal([]byte(`{"command":"pickskat","sessionId":"s1"}`), &input)
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusBadRequest)
	})

	t.Run("discard", func(t *testing.T) {
		a, b := 0, 1
		input := controller.SkatWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "s1"},
			DiscardA:     &a,
			DiscardB:     &b,
		}
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mockOutput)
	})

	t.Run("discard missing", func(t *testing.T) {
		a := 0
		input := controller.SkatWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "d", SessionID: "s1"},
			DiscardA:     &a,
		}
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusBadRequest)
	})

	t.Run("game declare", func(t *testing.T) {
		gt := int(domain.SkatGameSuit)
		ts := 1
		input := controller.SkatWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "g", SessionID: "s1"},
			GameType:     &gt,
			TrumpSuit:    &ts,
		}
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mockOutput)
	})

	t.Run("game declare missing", func(t *testing.T) {
		var input controller.SkatWebInput
		_ = json.Unmarshal([]byte(`{"command":"g","sessionId":"s1"}`), &input)
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusBadRequest)
	})

	t.Run("play", func(t *testing.T) {
		i := 2
		input := controller.SkatWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "p", SessionID: "s1"},
			CardIndex:    &i,
		}
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mockOutput)
	})

	t.Run("play missing", func(t *testing.T) {
		var input controller.SkatWebInput
		_ = json.Unmarshal([]byte(`{"command":"play","sessionId":"s1"}`), &input)
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusBadRequest)
	})

	t.Run("next", func(t *testing.T) {
		var input controller.SkatWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"s1"}`), &input)
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mockOutput)
	})

	t.Run("nextround", func(t *testing.T) {
		var input controller.SkatWebInput
		_ = json.Unmarshal([]byte(`{"command":"nr","sessionId":"s1"}`), &input)
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusOK)
		r.BodyIs(mockOutput)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.SkatWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"s1"}`), &input)
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusOK)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.SkatWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"s1"}`), &input)
		r := execRequest(t, ctrl.Exec, &input)
		r.CodeIs(http.StatusOK)
	})
}

func TestSkatWebInputToConfig(t *testing.T) {
	t.Run("default config when input has no config", func(t *testing.T) {
		in := controller.SkatWebInput{}
		got := in.ToConfig()
		if got != domain.DefaultSkatConfig() {
			t.Fatalf("expected default config, got %+v", got)
		}
	})

	t.Run("clamps target score", func(t *testing.T) {
		ts := -1
		cd := 9
		in := controller.SkatWebInput{
			Config: &controller.SkatWebConfig{
				CpuDifficulty: &cd,
				TargetScore:   &ts,
			},
		}
		got := in.ToConfig()
		if got.TargetScore < 1 {
			t.Fatalf("target score not clamped: %d", got.TargetScore)
		}
		if got.CpuDifficulty != domain.DefaultSkatConfig().CpuDifficulty {
			t.Fatalf("difficulty out of range should fall back to default, got %v", got.CpuDifficulty)
		}
	})
}
