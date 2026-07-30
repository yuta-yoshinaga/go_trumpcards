//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockLaughAndLieDownGame ラフ・アンド・ライダウン ゲームモック
type MockLaughAndLieDownGame struct {
	mock.Mock
}

func (_m *MockLaughAndLieDownGame) Reset() { _m.Called() }

func (_m *MockLaughAndLieDownGame) PlayCard(player, handIdx, takeCount int) error {
	return _m.Called(player, handIdx, takeCount).Error(0)
}

func (_m *MockLaughAndLieDownGame) CanTakeThree(player, handIdx int) bool {
	return _m.Called(player, handIdx).Bool(0)
}

func (_m *MockLaughAndLieDownGame) LaughAndLieDownCpuDecide(idx int) domain.LaughAndLieDownCpuAction {
	return _m.Called(idx).Get(0).(domain.LaughAndLieDownCpuAction)
}

func (_m *MockLaughAndLieDownGame) GetConfig() domain.LaughAndLieDownConfig {
	return _m.Called().Get(0).(domain.LaughAndLieDownConfig)
}

func (_m *MockLaughAndLieDownGame) SetConfig(cfg domain.LaughAndLieDownConfig) { _m.Called(cfg) }

func (_m *MockLaughAndLieDownGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

func (_m *MockLaughAndLieDownGame) GetPhase() domain.LaughAndLieDownPhase {
	return _m.Called().Get(0).(domain.LaughAndLieDownPhase)
}

func (_m *MockLaughAndLieDownGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

func (_m *MockLaughAndLieDownGame) GetLayout() []*domain.Card {
	if v, ok := _m.Called().Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockLaughAndLieDownGame) GetValidPlayIndices(player int) []int {
	if v, ok := _m.Called(player).Get(0).([]int); ok {
		return v
	}
	return nil
}

func (_m *MockLaughAndLieDownGame) GetWonCount(idx int) int { return _m.Called(idx).Int(0) }

func (_m *MockLaughAndLieDownGame) IsLaidDown(idx int) bool { return _m.Called(idx).Bool(0) }

func (_m *MockLaughAndLieDownGame) GetScore(idx int) int { return _m.Called(idx).Int(0) }

func (_m *MockLaughAndLieDownGame) GetDealerIdx() int { return _m.Called().Int(0) }

func (_m *MockLaughAndLieDownGame) GetLastInIdx() int { return _m.Called().Int(0) }

func (_m *MockLaughAndLieDownGame) GetPlayers() []*domain.LaughAndLieDownPlayer {
	if v, ok := _m.Called().Get(0).([]*domain.LaughAndLieDownPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockLaughAndLieDownGame) GetPlayer(idx int) *domain.LaughAndLieDownPlayer {
	if v, ok := _m.Called(idx).Get(0).(*domain.LaughAndLieDownPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockLaughAndLieDownGame) GetActionLog() []*domain.ActionLogEntry {
	if v, ok := _m.Called().Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
