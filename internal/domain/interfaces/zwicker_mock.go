//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockZwickerGame ツヴィッカー ゲームモック
type MockZwickerGame struct {
	mock.Mock
}

func (_m *MockZwickerGame) Reset() { _m.Called() }

func (_m *MockZwickerGame) Take(player, handIdx, playedValue int, tableIdxs, buildIdxs []int) error {
	return _m.Called(player, handIdx, playedValue, tableIdxs, buildIdxs).Error(0)
}

func (_m *MockZwickerGame) Build(player, handIdx int, tableIdxs []int, declaredValue int) error {
	return _m.Called(player, handIdx, tableIdxs, declaredValue).Error(0)
}

func (_m *MockZwickerGame) Trail(player, handIdx int) error {
	return _m.Called(player, handIdx).Error(0)
}

func (_m *MockZwickerGame) NextRound() error { return _m.Called().Error(0) }

func (_m *MockZwickerGame) ZwickerCpuDecide(idx int) domain.ZwickerCpuAction {
	return _m.Called(idx).Get(0).(domain.ZwickerCpuAction)
}

func (_m *MockZwickerGame) GetConfig() domain.ZwickerConfig {
	return _m.Called().Get(0).(domain.ZwickerConfig)
}

func (_m *MockZwickerGame) SetConfig(cfg domain.ZwickerConfig) { _m.Called(cfg) }

func (_m *MockZwickerGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

func (_m *MockZwickerGame) GetPhase() domain.ZwickerPhase {
	return _m.Called().Get(0).(domain.ZwickerPhase)
}

func (_m *MockZwickerGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

func (_m *MockZwickerGame) GetTableCards() []*domain.Card {
	if v, ok := _m.Called().Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockZwickerGame) GetBuilds() []*domain.ZwickerBuild {
	if v, ok := _m.Called().Get(0).([]*domain.ZwickerBuild); ok {
		return v
	}
	return nil
}

func (_m *MockZwickerGame) GetStockCount() int { return _m.Called().Int(0) }

func (_m *MockZwickerGame) GetTeamScore(team int) int { return _m.Called(team).Int(0) }

func (_m *MockZwickerGame) GetLastRoundScore() *domain.ZwickerRoundScore {
	if v, ok := _m.Called().Get(0).(*domain.ZwickerRoundScore); ok {
		return v
	}
	return nil
}

func (_m *MockZwickerGame) GetWinnerTeam() int { return _m.Called().Int(0) }

func (_m *MockZwickerGame) GetPlayers() []*domain.ZwickerPlayer {
	if v, ok := _m.Called().Get(0).([]*domain.ZwickerPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockZwickerGame) GetPlayer(idx int) *domain.ZwickerPlayer {
	if v, ok := _m.Called(idx).Get(0).(*domain.ZwickerPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockZwickerGame) GetActionLog() []*domain.ActionLogEntry {
	if v, ok := _m.Called().Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
