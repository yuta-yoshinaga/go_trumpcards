//go:build test

package controller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	mockUsecases "github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestDesmocheCuiController_SetDifficulty(t *testing.T) {
	for _, tc := range []struct {
		command string
		want    domain.DesmocheCpuDifficulty
	}{
		{"sd 0", domain.DesmocheCpuDifficultyEasy},
		{"sd 1", domain.DesmocheCpuDifficultyNormal},
		{"setdifficulty 2", domain.DesmocheCpuDifficultyHard},
	} {
		t.Run(tc.command, func(t *testing.T) {
			m := new(mockUsecases.MockDesmocheInteractor)
			m.On("GetConfig").Return(domain.DefaultDesmocheConfig())
			m.On("ResetWithConfig", mock.MatchedBy(func(cfg domain.DesmocheConfig) bool {
				return cfg.CpuDifficulty == tc.want
			})).Return("リセットしました")
			assert.Equal(t, "リセットしました", controller.NewDesmocheCuiController(m).Exec(tc.command))
			m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.DesmocheConfig) bool {
				return cfg.CpuDifficulty == tc.want
			}))
		})
	}

	m := new(mockUsecases.MockDesmocheInteractor)
	c := controller.NewDesmocheCuiController(m)
	for _, command := range []string{"sd 3", "sd -1", "sd abc"} {
		assert.Contains(t, c.Exec(command), "無効なCPU難易度です")
	}
}
