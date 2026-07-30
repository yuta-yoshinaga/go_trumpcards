//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSkitgubbeGame シートグッベ ゲームモック
type MockSkitgubbeGame struct {
	mock.Mock
}

func (_m *MockSkitgubbeGame) Reset() { _m.Called() }

func (_m *MockSkitgubbeGame) PlayCard(player, handIdx int) error {
	return _m.Called(player, handIdx).Error(0)
}

func (_m *MockSkitgubbeGame) PickUp(player int) error { return _m.Called(player).Error(0) }

func (_m *MockSkitgubbeGame) SkitgubbeCpuDecide(idx int) domain.SkitgubbeCpuAction {
	return _m.Called(idx).Get(0).(domain.SkitgubbeCpuAction)
}

func (_m *MockSkitgubbeGame) GetConfig() domain.SkitgubbeConfig {
	return _m.Called().Get(0).(domain.SkitgubbeConfig)
}

func (_m *MockSkitgubbeGame) SetConfig(cfg domain.SkitgubbeConfig) { _m.Called(cfg) }

func (_m *MockSkitgubbeGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

func (_m *MockSkitgubbeGame) GetPhase() domain.SkitgubbePhase {
	return _m.Called().Get(0).(domain.SkitgubbePhase)
}

func (_m *MockSkitgubbeGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

func (_m *MockSkitgubbeGame) GetStockCount() int { return _m.Called().Int(0) }

func (_m *MockSkitgubbeGame) GetTrumpSuit() int { return _m.Called().Int(0) }

func (_m *MockSkitgubbeGame) GetDuel() []*domain.Card {
	if v, ok := _m.Called().Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockSkitgubbeGame) GetDuelLeader() int { return _m.Called().Int(0) }

func (_m *MockSkitgubbeGame) GetPile() []*domain.Card {
	if v, ok := _m.Called().Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockSkitgubbeGame) GetCollectedCount(idx int) int { return _m.Called(idx).Int(0) }

func (_m *MockSkitgubbeGame) GetValidPlayIndices(player int) []int {
	if v, ok := _m.Called(player).Get(0).([]int); ok {
		return v
	}
	return nil
}

func (_m *MockSkitgubbeGame) IsFinished(idx int) bool { return _m.Called(idx).Bool(0) }

func (_m *MockSkitgubbeGame) GetPlayers() []*domain.SkitgubbePlayer {
	if v, ok := _m.Called().Get(0).([]*domain.SkitgubbePlayer); ok {
		return v
	}
	return nil
}

func (_m *MockSkitgubbeGame) GetPlayer(idx int) *domain.SkitgubbePlayer {
	if v, ok := _m.Called(idx).Get(0).(*domain.SkitgubbePlayer); ok {
		return v
	}
	return nil
}

func (_m *MockSkitgubbeGame) GetLoserIdx() int { return _m.Called().Int(0) }

func (_m *MockSkitgubbeGame) GetActionLog() []*domain.ActionLogEntry {
	if v, ok := _m.Called().Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
