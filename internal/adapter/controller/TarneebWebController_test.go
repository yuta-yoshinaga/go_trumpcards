//go:build test

package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func mustTarneebOutputJSON(msg string) string {
	out := &controller.TarneebWebOutput{
		Players:       []*controller.TarneebWebOutputPlayer{},
		TeamScores:    []int{},
		CurrentTrick:  []*controller.WebOutputTrickCard{},
		WinnerTeam:    -1,
		BidWinnerIdx:  -1,
		WebOutputBase: controller.WebOutputBase{Message: msg},
	}
	b, err := json.Marshal(out)
	if err != nil {
		panic(fmt.Sprintf("mustTarneebOutputJSON: %v", err))
	}
	return string(b)
}

func TestTarneebWebController_Method(t *testing.T) {
	mockOutput := `{"players":[],"phase":0}`

	tiMock := new(usecase.MockTarneebInteractor)
	tiMock.On("ResetWithConfig", domain.DefaultTarneebConfig()).Return(mockOutput)
	tiMock.On("Bid", 8).Return(mockOutput)
	tiMock.On("DeclareTrump", domain.CardDesignSpade).Return(mockOutput)
	tiMock.On("Play", 3).Return(mockOutput)
	tiMock.On("NextTrick").Return(mockOutput)
	tiMock.On("NextRound").Return(mockOutput)
	tiMock.On("Hint").Return(mockOutput)
	tiMock.On("ActionLog").Return(mockOutput)

	factory := func() uc.TarneebInteractorIF { return tiMock }
	ctrl := controller.NewTarneebWebController(factory)
	defer ctrl.Stop()

	t.Run("q quits", func(t *testing.T) {
		var input controller.TarneebWebInput
		_ = json.Unmarshal([]byte(`{"command":"q","sessionId":"tn1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mustTarneebOutputJSON("bye."))
	})

	t.Run("reset r", func(t *testing.T) {
		var input controller.TarneebWebInput
		_ = json.Unmarshal([]byte(`{"command":"r","sessionId":"tn1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("bid valid", func(t *testing.T) {
		bid := 8
		input := controller.TarneebWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "bid", SessionID: "tn1"},
			Bid:          &bid,
		}
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("trump valid", func(t *testing.T) {
		suit := domain.CardDesignSpade
		input := controller.TarneebWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "trump", SessionID: "tn1"},
			TrumpSuit:    &suit,
		}
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("play valid", func(t *testing.T) {
		idx := 3
		input := controller.TarneebWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "play", SessionID: "tn1"},
			CardIndex:    &idx,
		}
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("next n", func(t *testing.T) {
		var input controller.TarneebWebInput
		_ = json.Unmarshal([]byte(`{"command":"n","sessionId":"tn1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("nextround", func(t *testing.T) {
		var input controller.TarneebWebInput
		_ = json.Unmarshal([]byte(`{"command":"nextround","sessionId":"tn1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("log", func(t *testing.T) {
		var input controller.TarneebWebInput
		_ = json.Unmarshal([]byte(`{"command":"log","sessionId":"tn1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("hint", func(t *testing.T) {
		var input controller.TarneebWebInput
		_ = json.Unmarshal([]byte(`{"command":"hint","sessionId":"tn1"}`), &input)
		execRequest(t, ctrl.Exec, &input).BodyIs(mockOutput)
	})

	t.Run("unsupported command", func(t *testing.T) {
		var input controller.TarneebWebInput
		_ = json.Unmarshal([]byte(`{"command":"other","sessionId":"tn1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustTarneebOutputJSON("Unsupported command."))
	})

	t.Run("bid missing", func(t *testing.T) {
		var input controller.TarneebWebInput
		_ = json.Unmarshal([]byte(`{"command":"bid","sessionId":"tn1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustTarneebOutputJSON("param error: bid is required."))
	})

	t.Run("trump missing", func(t *testing.T) {
		var input controller.TarneebWebInput
		_ = json.Unmarshal([]byte(`{"command":"trump","sessionId":"tn1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustTarneebOutputJSON("param error: trumpSuit is required."))
	})

	t.Run("play missing cardIndex", func(t *testing.T) {
		var input controller.TarneebWebInput
		_ = json.Unmarshal([]byte(`{"command":"p","sessionId":"tn1"}`), &input)
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustTarneebOutputJSON("param error: cardIndex is required."))
	})

	t.Run("session id too long", func(t *testing.T) {
		input := controller.TarneebWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: strings.Repeat("a", controller.SessionMaxIDLen+1)},
		}
		recorded := execRequest(t, ctrl.Exec, &input)
		recorded.CodeIs(http.StatusBadRequest)
		recorded.BodyIs(mustTarneebOutputJSON("param error."))
	})
}

func TestTarneebWebController_CustomConfig(t *testing.T) {
	mockOutput := `{"players":[]}`

	t.Run("custom config", func(t *testing.T) {
		diff := 2
		limit := 41
		minBid := 9
		expected := domain.TarneebConfig{
			CpuDifficulty: domain.TarneebCpuDifficultyHard,
			PointLimit:    41,
			MinBid:        9,
		}
		tiMock := new(usecase.MockTarneebInteractor)
		tiMock.On("ResetWithConfig", expected).Return(mockOutput)

		factory := func() uc.TarneebInteractorIF { return tiMock }
		ctrl := controller.NewTarneebWebController(factory)
		defer ctrl.Stop()

		input := controller.TarneebWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "tn-cfg-1"},
			Config: &controller.TarneebWebConfig{
				CpuDifficulty: &diff,
				PointLimit:    &limit,
				MinBid:        &minBid,
			},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		tiMock.AssertCalled(t, "ResetWithConfig", expected)
	})

	t.Run("nil config uses default", func(t *testing.T) {
		tiMock := new(usecase.MockTarneebInteractor)
		tiMock.On("ResetWithConfig", domain.DefaultTarneebConfig()).Return(mockOutput)

		factory := func() uc.TarneebInteractorIF { return tiMock }
		ctrl := controller.NewTarneebWebController(factory)
		defer ctrl.Stop()

		input := controller.TarneebWebInput{
			BaseWebInput: controller.BaseWebInput{Command: "reset", SessionID: "tn-cfg-2"},
		}
		execRequest(t, ctrl.Exec, &input).CodeIs(http.StatusOK)
		tiMock.AssertCalled(t, "ResetWithConfig", domain.DefaultTarneebConfig())
	})
}
