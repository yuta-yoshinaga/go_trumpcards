//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockNainJauneGame ル・ナン・ジョーヌ ゲームモック
type MockNainJauneGame struct {
	mock.Mock
}

func (_m *MockNainJauneGame) Reset() { _m.Called() }

func (_m *MockNainJauneGame) Play(player, handIdx int) error {
	return _m.Called(player, handIdx).Error(0)
}

func (_m *MockNainJauneGame) NextDeal() error { return _m.Called().Error(0) }

func (_m *MockNainJauneGame) NainJauneCpuDecide(idx int) int { return _m.Called(idx).Int(0) }

func (_m *MockNainJauneGame) GetConfig() domain.NainJauneConfig {
	return _m.Called().Get(0).(domain.NainJauneConfig)
}

func (_m *MockNainJauneGame) SetConfig(cfg domain.NainJauneConfig) { _m.Called(cfg) }

func (_m *MockNainJauneGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

func (_m *MockNainJauneGame) GetPhase() domain.NainJaunePhase {
	return _m.Called().Get(0).(domain.NainJaunePhase)
}

func (_m *MockNainJauneGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

func (_m *MockNainJauneGame) GetBoard() domain.NainJauneBoard {
	return _m.Called().Get(0).(domain.NainJauneBoard)
}

func (_m *MockNainJauneGame) GetTalonCount() int { return _m.Called().Int(0) }

func (_m *MockNainJauneGame) GetAwards() []*domain.NainJauneAward {
	if v, ok := _m.Called().Get(0).([]*domain.NainJauneAward); ok {
		return v
	}
	return nil
}

func (_m *MockNainJauneGame) GetPlayedPile() []*domain.Card {
	if v, ok := _m.Called().Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockNainJauneGame) GetRunRank() int { return _m.Called().Int(0) }

func (_m *MockNainJauneGame) GetDealNumber() int { return _m.Called().Int(0) }

func (_m *MockNainJauneGame) GetDealWinner() int { return _m.Called().Int(0) }

func (_m *MockNainJauneGame) GetWinnerIdx() int { return _m.Called().Int(0) }

func (_m *MockNainJauneGame) GetPlayers() []*domain.NainJaunePlayer {
	if v, ok := _m.Called().Get(0).([]*domain.NainJaunePlayer); ok {
		return v
	}
	return nil
}

func (_m *MockNainJauneGame) GetPlayer(idx int) *domain.NainJaunePlayer {
	if v, ok := _m.Called(idx).Get(0).(*domain.NainJaunePlayer); ok {
		return v
	}
	return nil
}

func (_m *MockNainJauneGame) GetActionLog() []*domain.ActionLogEntry {
	if v, ok := _m.Called().Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
