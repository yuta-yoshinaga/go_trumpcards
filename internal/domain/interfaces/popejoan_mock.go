//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPopeJoanGame ポープ・ジョーン ゲームモック
type MockPopeJoanGame struct {
	mock.Mock
}

func (_m *MockPopeJoanGame) Reset() { _m.Called() }

func (_m *MockPopeJoanGame) Play(player, handIdx int) error {
	return _m.Called(player, handIdx).Error(0)
}

func (_m *MockPopeJoanGame) NextDeal() error { return _m.Called().Error(0) }

func (_m *MockPopeJoanGame) PopeJoanCpuDecide(idx int) int { return _m.Called(idx).Int(0) }

func (_m *MockPopeJoanGame) GetConfig() domain.PopeJoanConfig {
	return _m.Called().Get(0).(domain.PopeJoanConfig)
}

func (_m *MockPopeJoanGame) SetConfig(cfg domain.PopeJoanConfig) { _m.Called(cfg) }

func (_m *MockPopeJoanGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

func (_m *MockPopeJoanGame) GetPhase() domain.PopeJoanPhase {
	return _m.Called().Get(0).(domain.PopeJoanPhase)
}

func (_m *MockPopeJoanGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

func (_m *MockPopeJoanGame) GetBoard() domain.PopeJoanBoard {
	return _m.Called().Get(0).(domain.PopeJoanBoard)
}

func (_m *MockPopeJoanGame) GetTrumpSuit() int { return _m.Called().Int(0) }

func (_m *MockPopeJoanGame) GetTurnUp() *domain.Card {
	if v, ok := _m.Called().Get(0).(*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockPopeJoanGame) GetAwards() []*domain.PopeJoanAward {
	if v, ok := _m.Called().Get(0).([]*domain.PopeJoanAward); ok {
		return v
	}
	return nil
}

func (_m *MockPopeJoanGame) GetPlayedPile() []*domain.Card {
	if v, ok := _m.Called().Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockPopeJoanGame) GetRunSuit() int { return _m.Called().Int(0) }

func (_m *MockPopeJoanGame) GetRunRank() int { return _m.Called().Int(0) }

func (_m *MockPopeJoanGame) GetDealNumber() int { return _m.Called().Int(0) }

func (_m *MockPopeJoanGame) GetDealWinner() int { return _m.Called().Int(0) }

func (_m *MockPopeJoanGame) GetWinnerIdx() int { return _m.Called().Int(0) }

func (_m *MockPopeJoanGame) GetPlayers() []*domain.PopeJoanPlayer {
	if v, ok := _m.Called().Get(0).([]*domain.PopeJoanPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockPopeJoanGame) GetPlayer(idx int) *domain.PopeJoanPlayer {
	if v, ok := _m.Called(idx).Get(0).(*domain.PopeJoanPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockPopeJoanGame) GetActionLog() []*domain.ActionLogEntry {
	if v, ok := _m.Called().Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
