//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecase "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

func newBadugiTestController(mi *mockUsecase.MockBadugiInteractor) *controller.BadugiWebController {
	return controller.NewBadugiWebController(func() usecase.BadugiInteractorIF { return mi })
}

func TestBadugiWebController_Reset_Default(t *testing.T) {
	mi := new(mockUsecase.MockBadugiInteractor)
	bwc := newBadugiTestController(mi)
	defer bwc.Stop()

	cfg := domain.DefaultBadugiConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	rec := execRequest(t, bwc.Exec, map[string]any{
		"command":   "reset",
		"sessionId": "s1",
	})
	rec.CodeIs(200)
}

func TestBadugiWebController_Reset_WithConfigOverrides(t *testing.T) {
	mi := new(mockUsecase.MockBadugiInteractor)
	bwc := newBadugiTestController(mi)
	defer bwc.Stop()

	cfg := domain.DefaultBadugiConfig()
	cfg.CpuCount = 2
	cfg.BettingLimit = domain.BettingLimitNoLimit
	cfg.CpuMetaAI = true
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	rec := execRequest(t, bwc.Exec, map[string]any{
		"command":      "reset",
		"sessionId":    "s1",
		"cpuCount":     2,
		"bettingLimit": int(domain.BettingLimitNoLimit),
		"cpuMetaAI":    true,
	})
	rec.CodeIs(200)
}

func TestBadugiWebController_Reset_OutOfRangeClamped(t *testing.T) {
	mi := new(mockUsecase.MockBadugiInteractor)
	bwc := newBadugiTestController(mi)
	defer bwc.Stop()

	// Out-of-range inputs fall back to the defaults.
	cfg := domain.DefaultBadugiConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	rec := execRequest(t, bwc.Exec, map[string]any{
		"command":      "reset",
		"sessionId":    "s1",
		"cpuCount":     99,
		"bettingLimit": 99,
	})
	rec.CodeIs(200)
}

func TestBadugiWebController_Betting(t *testing.T) {
	tests := []struct {
		command string
		action  int
		amount  int
	}{
		{"fold", domain.BadugiActionFold, 0},
		{"check", domain.BadugiActionCheck, 0},
		{"call", domain.BadugiActionCall, 0},
		{"bet", domain.BadugiActionBet, 20},
		{"raise", domain.BadugiActionRaise, 40},
		{"allin", domain.BadugiActionAllIn, 0},
	}
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			mi := new(mockUsecase.MockBadugiInteractor)
			bwc := newBadugiTestController(mi)
			defer bwc.Stop()

			mi.On("Action", tt.action, tt.amount, 0).Return(`{}`)
			rec := execRequest(t, bwc.Exec, map[string]any{
				"command":   tt.command,
				"sessionId": "s1",
				"amount":    tt.amount,
			})
			rec.CodeIs(200)
		})
	}
}

func TestBadugiWebController_Exchange(t *testing.T) {
	mi := new(mockUsecase.MockBadugiInteractor)
	bwc := newBadugiTestController(mi)
	defer bwc.Stop()

	mi.On("Exchange", []int{0, 2}, 4200).Return(`{}`)
	rec := execRequest(t, bwc.Exec, map[string]any{
		"command":     "exchange",
		"sessionId":   "s1",
		"indices":     []int{0, 2},
		"humanPlayMs": 4200,
	})
	rec.CodeIs(200)
}

func TestBadugiWebController_Exchange_NoIndices(t *testing.T) {
	mi := new(mockUsecase.MockBadugiInteractor)
	bwc := newBadugiTestController(mi)
	defer bwc.Stop()

	mi.On("Exchange", []int{}, 0).Return(`{}`)
	rec := execRequest(t, bwc.Exec, map[string]any{
		"command":   "exchange",
		"sessionId": "s1",
	})
	rec.CodeIs(200)
}

func TestBadugiWebController_Stand(t *testing.T) {
	mi := new(mockUsecase.MockBadugiInteractor)
	bwc := newBadugiTestController(mi)
	defer bwc.Stop()

	mi.On("Stand", 4200).Return(`{}`)
	rec := execRequest(t, bwc.Exec, map[string]any{
		"command":     "stand",
		"sessionId":   "s1",
		"humanPlayMs": 4200,
	})
	rec.CodeIs(200)
}

func TestBadugiWebController_ActionLog(t *testing.T) {
	mi := new(mockUsecase.MockBadugiInteractor)
	bwc := newBadugiTestController(mi)
	defer bwc.Stop()

	mi.On("ActionLog").Return(`{"entries":[]}`)
	rec := execRequest(t, bwc.Exec, map[string]any{
		"command":   "log",
		"sessionId": "s1",
	})
	rec.CodeIs(200)
}
