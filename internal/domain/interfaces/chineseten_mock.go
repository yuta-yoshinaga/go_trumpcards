//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockChineseTenGame 撿紅點 ゲームモック
type MockChineseTenGame struct {
	mock.Mock
}

func (_m *MockChineseTenGame) Reset() { _m.Called() }

func (_m *MockChineseTenGame) PlayCard(player, handIdx int) error {
	return _m.Called(player, handIdx).Error(0)
}

func (_m *MockChineseTenGame) SelectCapture(player, layoutIdx int) error {
	return _m.Called(player, layoutIdx).Error(0)
}

func (_m *MockChineseTenGame) ChineseTenCpuDecide(idx int) domain.ChineseTenCpuAction {
	return _m.Called(idx).Get(0).(domain.ChineseTenCpuAction)
}

func (_m *MockChineseTenGame) GetConfig() domain.ChineseTenConfig {
	return _m.Called().Get(0).(domain.ChineseTenConfig)
}

func (_m *MockChineseTenGame) SetConfig(cfg domain.ChineseTenConfig) { _m.Called(cfg) }

func (_m *MockChineseTenGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

func (_m *MockChineseTenGame) GetPhase() domain.ChineseTenPhase {
	return _m.Called().Get(0).(domain.ChineseTenPhase)
}

func (_m *MockChineseTenGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

func (_m *MockChineseTenGame) GetLayout() []*domain.Card {
	if v, ok := _m.Called().Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockChineseTenGame) GetStockCount() int { return _m.Called().Int(0) }

func (_m *MockChineseTenGame) GetCaptured(idx int) []*domain.Card {
	if v, ok := _m.Called(idx).Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockChineseTenGame) GetPlayers() []*domain.ChineseTenPlayer {
	if v, ok := _m.Called().Get(0).([]*domain.ChineseTenPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockChineseTenGame) GetPlayer(idx int) *domain.ChineseTenPlayer {
	if v, ok := _m.Called(idx).Get(0).(*domain.ChineseTenPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockChineseTenGame) GetScore(idx int) int { return _m.Called(idx).Int(0) }

func (_m *MockChineseTenGame) GetPendingCard() *domain.Card {
	if v, ok := _m.Called().Get(0).(*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockChineseTenGame) GetSelectableIndices() []int {
	if v, ok := _m.Called().Get(0).([]int); ok {
		return v
	}
	return nil
}

func (_m *MockChineseTenGame) GetWinnerIdx() int { return _m.Called().Int(0) }

func (_m *MockChineseTenGame) GetActionLog() []*domain.ActionLogEntry {
	if v, ok := _m.Called().Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
