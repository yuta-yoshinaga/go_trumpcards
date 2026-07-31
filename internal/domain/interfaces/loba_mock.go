//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockLobaGame ロバ ゲームモック
type MockLobaGame struct {
	mock.Mock
}

func (_m *MockLobaGame) Reset() { _m.Called() }

func (_m *MockLobaGame) DrawFromStock(player int) error { return _m.Called(player).Error(0) }

func (_m *MockLobaGame) DrawFromDiscard(player int) error { return _m.Called(player).Error(0) }

func (_m *MockLobaGame) Meld(player int, handIdxs []int) error {
	return _m.Called(player, handIdxs).Error(0)
}

func (_m *MockLobaGame) LayOff(player, handIdx, meldIdx int) error {
	return _m.Called(player, handIdx, meldIdx).Error(0)
}

func (_m *MockLobaGame) Discard(player, handIdx int) error {
	return _m.Called(player, handIdx).Error(0)
}

func (_m *MockLobaGame) NextRound() error { return _m.Called().Error(0) }

func (_m *MockLobaGame) LobaCpuDecide(idx int) domain.LobaCpuAction {
	return _m.Called(idx).Get(0).(domain.LobaCpuAction)
}

func (_m *MockLobaGame) GetConfig() domain.LobaConfig {
	return _m.Called().Get(0).(domain.LobaConfig)
}

func (_m *MockLobaGame) SetConfig(cfg domain.LobaConfig) { _m.Called(cfg) }

func (_m *MockLobaGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

func (_m *MockLobaGame) GetPhase() domain.LobaPhase {
	return _m.Called().Get(0).(domain.LobaPhase)
}

func (_m *MockLobaGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

func (_m *MockLobaGame) GetStockCount() int { return _m.Called().Int(0) }

func (_m *MockLobaGame) GetDiscardTop() *domain.Card {
	if v, ok := _m.Called().Get(0).(*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockLobaGame) GetMelds() []*domain.LobaMeld {
	if v, ok := _m.Called().Get(0).([]*domain.LobaMeld); ok {
		return v
	}
	return nil
}

func (_m *MockLobaGame) HasMelded(idx int) bool { return _m.Called(idx).Bool(0) }

func (_m *MockLobaGame) GetScore(idx int) int { return _m.Called(idx).Int(0) }

func (_m *MockLobaGame) IsEliminated(idx int) bool { return _m.Called(idx).Bool(0) }

func (_m *MockLobaGame) GetRoundNumber() int { return _m.Called().Int(0) }

func (_m *MockLobaGame) GetRoundWinner() int { return _m.Called().Int(0) }

func (_m *MockLobaGame) IsRoundClean() bool { return _m.Called().Bool(0) }

func (_m *MockLobaGame) GetWinnerIdx() int { return _m.Called().Int(0) }

func (_m *MockLobaGame) GetPlayers() []*domain.LobaPlayer {
	if v, ok := _m.Called().Get(0).([]*domain.LobaPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockLobaGame) GetPlayer(idx int) *domain.LobaPlayer {
	if v, ok := _m.Called(idx).Get(0).(*domain.LobaPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockLobaGame) GetActionLog() []*domain.ActionLogEntry {
	if v, ok := _m.Called().Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
