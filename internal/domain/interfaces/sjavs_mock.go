//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSjavsGame シャウス ゲームモック
type MockSjavsGame struct {
	mock.Mock
}

func (_m *MockSjavsGame) Reset() { _m.Called() }

func (_m *MockSjavsGame) Bid(player, length int) error {
	return _m.Called(player, length).Error(0)
}

func (_m *MockSjavsGame) PlayCard(player, handIdx int) error {
	return _m.Called(player, handIdx).Error(0)
}

func (_m *MockSjavsGame) NextHand() error { return _m.Called().Error(0) }

func (_m *MockSjavsGame) SjavsCpuDecide(idx int) domain.SjavsCpuAction {
	return _m.Called(idx).Get(0).(domain.SjavsCpuAction)
}

func (_m *MockSjavsGame) GetConfig() domain.SjavsConfig {
	return _m.Called().Get(0).(domain.SjavsConfig)
}

func (_m *MockSjavsGame) SetConfig(cfg domain.SjavsConfig) { _m.Called(cfg) }

func (_m *MockSjavsGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

func (_m *MockSjavsGame) GetPhase() domain.SjavsPhase {
	return _m.Called().Get(0).(domain.SjavsPhase)
}

func (_m *MockSjavsGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

func (_m *MockSjavsGame) GetDealerIdx() int { return _m.Called().Int(0) }

func (_m *MockSjavsGame) GetTrumpSuit() int { return _m.Called().Int(0) }

func (_m *MockSjavsGame) GetBidderIdx() int { return _m.Called().Int(0) }

func (_m *MockSjavsGame) GetBidLength() int { return _m.Called().Int(0) }

func (_m *MockSjavsGame) GetBids() []int {
	if v, ok := _m.Called().Get(0).([]int); ok {
		return v
	}
	return nil
}

func (_m *MockSjavsGame) LongestTrumpLength(player int) int { return _m.Called(player).Int(0) }

func (_m *MockSjavsGame) GetTrick() []domain.SjavsTrickCard {
	if v, ok := _m.Called().Get(0).([]domain.SjavsTrickCard); ok {
		return v
	}
	return nil
}

func (_m *MockSjavsGame) GetTrickNumber() int { return _m.Called().Int(0) }

func (_m *MockSjavsGame) GetValidPlayIndices(player int) []int {
	if v, ok := _m.Called(player).Get(0).([]int); ok {
		return v
	}
	return nil
}

func (_m *MockSjavsGame) GetTeamPoints(team int) int { return _m.Called(team).Int(0) }

func (_m *MockSjavsGame) GetRemaining(team int) int { return _m.Called(team).Int(0) }

func (_m *MockSjavsGame) GetCrosses(team int) int { return _m.Called(team).Int(0) }

func (_m *MockSjavsGame) GetCarryOver() int { return _m.Called().Int(0) }

func (_m *MockSjavsGame) GetHandResult() *domain.SjavsHandResult {
	if v, ok := _m.Called().Get(0).(*domain.SjavsHandResult); ok {
		return v
	}
	return nil
}

func (_m *MockSjavsGame) GetWinnerTeam() int { return _m.Called().Int(0) }

func (_m *MockSjavsGame) IsDoubleVictory() bool { return _m.Called().Bool(0) }

func (_m *MockSjavsGame) GetPlayers() []*domain.SjavsPlayer {
	if v, ok := _m.Called().Get(0).([]*domain.SjavsPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockSjavsGame) GetPlayer(idx int) *domain.SjavsPlayer {
	if v, ok := _m.Called(idx).Get(0).(*domain.SjavsPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockSjavsGame) GetActionLog() []*domain.ActionLogEntry {
	if v, ok := _m.Called().Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
