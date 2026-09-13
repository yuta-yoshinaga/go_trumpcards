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

func TestZwickerCuiController_SetDifficulty(t *testing.T) {
	for _, tc := range []struct {
		command string
		want    domain.ZwickerCpuDifficulty
	}{
		{"sd 0", domain.ZwickerCpuDifficultyEasy},
		{"sd 1", domain.ZwickerCpuDifficultyNormal},
		{"setdifficulty 2", domain.ZwickerCpuDifficultyHard},
	} {
		t.Run(tc.command, func(t *testing.T) {
			m := new(mockUsecases.MockZwickerInteractor)
			m.On("GetConfig").Return(domain.DefaultZwickerConfig())
			m.On("ResetWithConfig", mock.MatchedBy(func(cfg domain.ZwickerConfig) bool {
				return cfg.CpuDifficulty == tc.want
			})).Return("リセットしました")
			assert.Equal(t, "リセットしました", controller.NewZwickerCuiController(m).Exec(tc.command))
			m.AssertCalled(t, "ResetWithConfig", mock.MatchedBy(func(cfg domain.ZwickerConfig) bool {
				return cfg.CpuDifficulty == tc.want
			}))
		})
	}

	m := new(mockUsecases.MockZwickerInteractor)
	c := controller.NewZwickerCuiController(m)
	for _, command := range []string{"sd 3", "sd -1", "sd abc"} {
		assert.Contains(t, c.Exec(command), "無効なCPU難易度です")
	}
}
