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

func newDeuceToSevenTestController(mi *mockUsecase.MockDeuceToSevenInteractor) *controller.DeuceToSevenWebController {
	return controller.NewDeuceToSevenWebController(func() usecase.DeuceToSevenInteractorIF { return mi })
}

func TestDeuceToSevenWebController_Reset_Default(t *testing.T) {
	mi := new(mockUsecase.MockDeuceToSevenInteractor)
	dwc := newDeuceToSevenTestController(mi)
	defer dwc.Stop()

	cfg := domain.DefaultDeuceToSevenConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	rec := execRequest(t, dwc.Exec, map[string]any{
		"command":   "reset",
		"sessionId": "s1",
	})
	rec.CodeIs(200)
}

func TestDeuceToSevenWebController_Reset_WithConfigOverrides(t *testing.T) {
	mi := new(mockUsecase.MockDeuceToSevenInteractor)
	dwc := newDeuceToSevenTestController(mi)
	defer dwc.Stop()

	cfg := domain.DefaultDeuceToSevenConfig()
	cfg.CpuCount = 2
	cfg.BettingLimit = domain.BettingLimitNoLimit
	cfg.CpuMetaAI = true
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	rec := execRequest(t, dwc.Exec, map[string]any{
		"command":      "reset",
		"sessionId":    "s1",
		"cpuCount":     2,
		"bettingLimit": int(domain.BettingLimitNoLimit),
		"cpuMetaAI":    true,
	})
	rec.CodeIs(200)
}

func TestDeuceToSevenWebController_Reset_OutOfRangeClamped(t *testing.T) {
	mi := new(mockUsecase.MockDeuceToSevenInteractor)
	dwc := newDeuceToSevenTestController(mi)
	defer dwc.Stop()

	cfg := domain.DefaultDeuceToSevenConfig()
	mi.On("ResetWithConfig", cfg, mock.Anything).Return(`{"phase":1}`)

	rec := execRequest(t, dwc.Exec, map[string]any{
		"command":      "reset",
		"sessionId":    "s1",
		"cpuCount":     99,
		"bettingLimit": 99,
	})
	rec.CodeIs(200)
}

func TestDeuceToSevenWebController_Betting(t *testing.T) {
	tests := []struct {
		command string
		action  int
		amount  int
	}{
		{"fold", domain.DeuceToSevenActionFold, 0},
		{"check", domain.DeuceToSevenActionCheck, 0},
		{"call", domain.DeuceToSevenActionCall, 0},
		{"bet", domain.DeuceToSevenActionBet, 20},
		{"raise", domain.DeuceToSevenActionRaise, 40},
		{"allin", domain.DeuceToSevenActionAllIn, 0},
	}
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			mi := new(mockUsecase.MockDeuceToSevenInteractor)
			dwc := newDeuceToSevenTestController(mi)
			defer dwc.Stop()

			mi.On("Action", tt.action, tt.amount, 0).Return(`{}`)
			rec := execRequest(t, dwc.Exec, map[string]any{
				"command":   tt.command,
				"sessionId": "s1",
				"amount":    tt.amount,
			})
			rec.CodeIs(200)
		})
	}
}

func TestDeuceToSevenWebController_Exchange(t *testing.T) {
	mi := new(mockUsecase.MockDeuceToSevenInteractor)
	dwc := newDeuceToSevenTestController(mi)
	defer dwc.Stop()

	mi.On("Exchange", []int{0, 2}).Return(`{}`)
	rec := execRequest(t, dwc.Exec, map[string]any{
		"command":   "exchange",
		"sessionId": "s1",
		"indices":   []int{0, 2},
	})
	rec.CodeIs(200)
}

func TestDeuceToSevenWebController_Exchange_NoIndices(t *testing.T) {
	mi := new(mockUsecase.MockDeuceToSevenInteractor)
	dwc := newDeuceToSevenTestController(mi)
	defer dwc.Stop()

	mi.On("Exchange", []int{}).Return(`{}`)
	rec := execRequest(t, dwc.Exec, map[string]any{
		"command":   "exchange",
		"sessionId": "s1",
	})
	rec.CodeIs(200)
}

func TestDeuceToSevenWebController_Stand(t *testing.T) {
	mi := new(mockUsecase.MockDeuceToSevenInteractor)
	dwc := newDeuceToSevenTestController(mi)
	defer dwc.Stop()

	mi.On("Stand").Return(`{}`)
	rec := execRequest(t, dwc.Exec, map[string]any{
		"command":   "stand",
		"sessionId": "s1",
	})
	rec.CodeIs(200)
}

func TestDeuceToSevenWebController_ActionLog(t *testing.T) {
	mi := new(mockUsecase.MockDeuceToSevenInteractor)
	dwc := newDeuceToSevenTestController(mi)
	defer dwc.Stop()

	mi.On("ActionLog").Return(`{"entries":[]}`)
	rec := execRequest(t, dwc.Exec, map[string]any{
		"command":   "log",
		"sessionId": "s1",
	})
	rec.CodeIs(200)
}
