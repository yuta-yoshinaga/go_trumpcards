//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockToepenGame トゥーペン ゲームモック
type MockToepenGame struct {
	mock.Mock
}

func (_m *MockToepenGame) Reset() { _m.Called() }

func (_m *MockToepenGame) PlayCard(player, handIdx int) error {
	return _m.Called(player, handIdx).Error(0)
}

func (_m *MockToepenGame) Toep(player int) error { return _m.Called(player).Error(0) }

func (_m *MockToepenGame) Respond(player int, stay bool) error {
	return _m.Called(player, stay).Error(0)
}

func (_m *MockToepenGame) Redeal(player int) error { return _m.Called(player).Error(0) }

func (_m *MockToepenGame) CanRedeal(player int) bool { return _m.Called(player).Bool(0) }

func (_m *MockToepenGame) NextHand() error { return _m.Called().Error(0) }

func (_m *MockToepenGame) ToepenCpuDecide(idx int) domain.ToepenCpuAction {
	return _m.Called(idx).Get(0).(domain.ToepenCpuAction)
}

func (_m *MockToepenGame) GetConfig() domain.ToepenConfig {
	return _m.Called().Get(0).(domain.ToepenConfig)
}

func (_m *MockToepenGame) SetConfig(cfg domain.ToepenConfig) { _m.Called(cfg) }

func (_m *MockToepenGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

func (_m *MockToepenGame) GetPhase() domain.ToepenPhase {
	return _m.Called().Get(0).(domain.ToepenPhase)
}

func (_m *MockToepenGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

func (_m *MockToepenGame) GetLeadPlayerIdx() int { return _m.Called().Int(0) }

func (_m *MockToepenGame) GetDealerIdx() int { return _m.Called().Int(0) }

func (_m *MockToepenGame) GetCurrentTrick() []*domain.TrickCard {
	if v, ok := _m.Called().Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

func (_m *MockToepenGame) GetLeadSuit() int { return _m.Called().Int(0) }

func (_m *MockToepenGame) GetTrickNumber() int { return _m.Called().Int(0) }

func (_m *MockToepenGame) GetStake() int { return _m.Called().Int(0) }

func (_m *MockToepenGame) GetKnockerIdx() int { return _m.Called().Int(0) }

func (_m *MockToepenGame) GetPendingRespondent() int { return _m.Called().Int(0) }

func (_m *MockToepenGame) GetLastTrickWinner() int { return _m.Called().Int(0) }

func (_m *MockToepenGame) GetHandNumber() int { return _m.Called().Int(0) }

func (_m *MockToepenGame) GetPlayers() []*domain.ToepenPlayer {
	if v, ok := _m.Called().Get(0).([]*domain.ToepenPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockToepenGame) GetPlayer(idx int) *domain.ToepenPlayer {
	if v, ok := _m.Called(idx).Get(0).(*domain.ToepenPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockToepenGame) GetLives(idx int) int { return _m.Called(idx).Int(0) }

func (_m *MockToepenGame) IsFolded(idx int) bool { return _m.Called(idx).Bool(0) }

func (_m *MockToepenGame) IsEliminated(idx int) bool { return _m.Called(idx).Bool(0) }

func (_m *MockToepenGame) GetValidPlayIndices(player int) []int {
	if v, ok := _m.Called(player).Get(0).([]int); ok {
		return v
	}
	return nil
}

func (_m *MockToepenGame) GetWinnerIdx() int { return _m.Called().Int(0) }

func (_m *MockToepenGame) GetActionLog() []*domain.ActionLogEntry {
	if v, ok := _m.Called().Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
