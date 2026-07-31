//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPochGame ポッホ ゲームモック
type MockPochGame struct {
	mock.Mock
}

func (_m *MockPochGame) Reset() { _m.Called() }

func (_m *MockPochGame) Bet(player int) error { return _m.Called(player).Error(0) }

func (_m *MockPochGame) Fold(player int) error { return _m.Called(player).Error(0) }

func (_m *MockPochGame) Play(player, handIdx int) error {
	return _m.Called(player, handIdx).Error(0)
}

func (_m *MockPochGame) NextDeal() error { return _m.Called().Error(0) }

func (_m *MockPochGame) PochCpuDecide(idx int) domain.PochCpuAction {
	return _m.Called(idx).Get(0).(domain.PochCpuAction)
}

func (_m *MockPochGame) GetConfig() domain.PochConfig {
	return _m.Called().Get(0).(domain.PochConfig)
}

func (_m *MockPochGame) SetConfig(cfg domain.PochConfig) { _m.Called(cfg) }

func (_m *MockPochGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

func (_m *MockPochGame) GetPhase() domain.PochPhase {
	return _m.Called().Get(0).(domain.PochPhase)
}

func (_m *MockPochGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

func (_m *MockPochGame) GetBoard() domain.PochBoard {
	return _m.Called().Get(0).(domain.PochBoard)
}

func (_m *MockPochGame) GetPaySuit() int { return _m.Called().Int(0) }

func (_m *MockPochGame) GetTurnUp() *domain.Card {
	if v, ok := _m.Called().Get(0).(*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockPochGame) GetStakingAwards() []*domain.PochStakingAward {
	if v, ok := _m.Called().Get(0).([]*domain.PochStakingAward); ok {
		return v
	}
	return nil
}

func (_m *MockPochGame) GetBetTarget() int { return _m.Called().Int(0) }

func (_m *MockPochGame) GetPochenWinner() int { return _m.Called().Int(0) }

func (_m *MockPochGame) GetPochenPot() int { return _m.Called().Int(0) }

func (_m *MockPochGame) GetPlayedPile() []*domain.Card {
	if v, ok := _m.Called().Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockPochGame) GetStopsSuit() int { return _m.Called().Int(0) }

func (_m *MockPochGame) GetStopsRank() int { return _m.Called().Int(0) }

func (_m *MockPochGame) GetDealNumber() int { return _m.Called().Int(0) }

func (_m *MockPochGame) GetDealWinner() int { return _m.Called().Int(0) }

func (_m *MockPochGame) GetWinnerIdx() int { return _m.Called().Int(0) }

func (_m *MockPochGame) GetPlayers() []*domain.PochPlayer {
	if v, ok := _m.Called().Get(0).([]*domain.PochPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockPochGame) GetPlayer(idx int) *domain.PochPlayer {
	if v, ok := _m.Called(idx).Get(0).(*domain.PochPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockPochGame) GetActionLog() []*domain.ActionLogEntry {
	if v, ok := _m.Called().Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
