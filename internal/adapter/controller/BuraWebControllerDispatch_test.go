//go:build test

package controller_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	uc "github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func TestBuraWebController_Dispatch(t *testing.T) {
	const mockOutput = `{"phase":0}`

	biMock := new(usecase.MockBuraInteractor)
	biMock.On("ResetWithConfig", domain.DefaultBuraConfig()).Return(mockOutput)
	biMock.On("Play", []int{0, 1}).Return(mockOutput)
	biMock.On("Claim").Return(mockOutput)
	biMock.On("Declare").Return(mockOutput)
	biMock.On("Hint").Return(mockOutput)
	biMock.On("ActionLog").Return(mockOutput)

	ctrl := controller.NewBuraWebController(func() uc.BuraInteractorIF { return biMock })
	defer ctrl.Stop()

	exec := func(t *testing.T, body string) *recorded {
		t.Helper()
		var input controller.BuraWebInput
		if err := json.Unmarshal([]byte(body), &input); err != nil {
			t.Fatalf("unmarshal %s: %v", body, err)
		}
		return execRequest(t, ctrl.Exec, &input)
	}

	t.Run("reset without a config block does not crash", func(t *testing.T) {
		// `config` is optional on the wire; this is the request that panicked
		// on a nil *BuraWebConfig before configOrDefault was used.
		rec := exec(t, `{"command":"r","sessionId":"b1"}`)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
	})

	t.Run("play forwards every index", func(t *testing.T) {
		rec := exec(t, `{"command":"p","cardIndices":[0,1],"sessionId":"b1"}`)
		rec.CodeIs(http.StatusOK)
		rec.BodyIs(mockOutput)
		biMock.AssertCalled(t, "Play", []int{0, 1})
	})

	t.Run("play without indices is a parameter error", func(t *testing.T) {
		// A lead needs at least one card, and an empty list would otherwise
		// reach the domain as a zero-card play.
		rec := exec(t, `{"command":"p","sessionId":"b1"}`)
		rec.CodeIs(http.StatusBadRequest)
		if !strings.Contains(rec.Body.String(), "cardIndices is required") {
			t.Errorf("expected a parameter error, got %s", rec.Body.String())
		}
	})

	t.Run("claim", func(t *testing.T) {
		exec(t, `{"command":"c","sessionId":"b1"}`).CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "Claim")
	})

	t.Run("declare", func(t *testing.T) {
		exec(t, `{"command":"d","sessionId":"b1"}`).CodeIs(http.StatusOK)
		biMock.AssertCalled(t, "Declare")
	})

	t.Run("hint and log", func(t *testing.T) {
		exec(t, `{"command":"h","sessionId":"b1"}`).CodeIs(http.StatusOK)
		exec(t, `{"command":"l","sessionId":"b1"}`).CodeIs(http.StatusOK)
	})

	t.Run("quit", func(t *testing.T) {
		exec(t, `{"command":"q","sessionId":"b1"}`).CodeIs(http.StatusOK)
	})
}
