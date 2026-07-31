//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDesmocheGame デスモチェ ゲームモック
type MockDesmocheGame struct {
	mock.Mock
}

func (_m *MockDesmocheGame) Reset() { _m.Called() }

func (_m *MockDesmocheGame) DrawFromStock(player int) error { return _m.Called(player).Error(0) }

func (_m *MockDesmocheGame) DrawFromDiscard(player int) error { return _m.Called(player).Error(0) }

func (_m *MockDesmocheGame) Meld(player int, handIdxs []int) error {
	return _m.Called(player, handIdxs).Error(0)
}

func (_m *MockDesmocheGame) LayOff(player, handIdx, meldIdx int) error {
	return _m.Called(player, handIdx, meldIdx).Error(0)
}

func (_m *MockDesmocheGame) Desmoche(player, fromMeldIdx, cardIdx, toMeldIdx int) error {
	return _m.Called(player, fromMeldIdx, cardIdx, toMeldIdx).Error(0)
}

func (_m *MockDesmocheGame) Discard(player, handIdx int) error {
	return _m.Called(player, handIdx).Error(0)
}

func (_m *MockDesmocheGame) NextRound() error { return _m.Called().Error(0) }

func (_m *MockDesmocheGame) DesmocheCpuDecide(idx int) domain.DesmocheCpuAction {
	return _m.Called(idx).Get(0).(domain.DesmocheCpuAction)
}

func (_m *MockDesmocheGame) GetConfig() domain.DesmocheConfig {
	return _m.Called().Get(0).(domain.DesmocheConfig)
}

func (_m *MockDesmocheGame) SetConfig(cfg domain.DesmocheConfig) { _m.Called(cfg) }

func (_m *MockDesmocheGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

func (_m *MockDesmocheGame) GetPhase() domain.DesmochePhase {
	return _m.Called().Get(0).(domain.DesmochePhase)
}

func (_m *MockDesmocheGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

func (_m *MockDesmocheGame) GetStockCount() int { return _m.Called().Int(0) }

func (_m *MockDesmocheGame) GetDiscardTop() *domain.Card {
	if v, ok := _m.Called().Get(0).(*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockDesmocheGame) GetMelds() []*domain.DesmocheMeld {
	if v, ok := _m.Called().Get(0).([]*domain.DesmocheMeld); ok {
		return v
	}
	return nil
}

func (_m *MockDesmocheGame) MeldedCount(player int) int { return _m.Called(player).Int(0) }

func (_m *MockDesmocheGame) GetPot() int { return _m.Called().Int(0) }

func (_m *MockDesmocheGame) GetScore(idx int) int { return _m.Called(idx).Int(0) }

func (_m *MockDesmocheGame) GetRoundNumber() int { return _m.Called().Int(0) }

func (_m *MockDesmocheGame) GetRoundWinner() int { return _m.Called().Int(0) }

func (_m *MockDesmocheGame) IsRoundExhausted() bool { return _m.Called().Bool(0) }

func (_m *MockDesmocheGame) GetWinnerIdx() int { return _m.Called().Int(0) }

func (_m *MockDesmocheGame) GetPlayers() []*domain.DesmochePlayer {
	if v, ok := _m.Called().Get(0).([]*domain.DesmochePlayer); ok {
		return v
	}
	return nil
}

func (_m *MockDesmocheGame) GetPlayer(idx int) *domain.DesmochePlayer {
	if v, ok := _m.Called(idx).Get(0).(*domain.DesmochePlayer); ok {
		return v
	}
	return nil
}

func (_m *MockDesmocheGame) GetActionLog() []*domain.ActionLogEntry {
	if v, ok := _m.Called().Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
