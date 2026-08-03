//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTrexGame トリックス ゲームモック
type MockTrexGame struct {
	mock.Mock
}

func (_m *MockTrexGame) Reset() { _m.Called() }

func (_m *MockTrexGame) ChooseContract(player int, contract domain.TrexContract) error {
	return _m.Called(player, contract).Error(0)
}

func (_m *MockTrexGame) PlayCard(player, handIdx int) error {
	return _m.Called(player, handIdx).Error(0)
}

func (_m *MockTrexGame) Pass(player int) error { return _m.Called(player).Error(0) }

func (_m *MockTrexGame) NextDeal() error { return _m.Called().Error(0) }

func (_m *MockTrexGame) TrexCpuDecide(idx int) domain.TrexCpuAction {
	return _m.Called(idx).Get(0).(domain.TrexCpuAction)
}

func (_m *MockTrexGame) GetConfig() domain.TrexConfig {
	return _m.Called().Get(0).(domain.TrexConfig)
}

func (_m *MockTrexGame) SetConfig(cfg domain.TrexConfig) { _m.Called(cfg) }

func (_m *MockTrexGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

func (_m *MockTrexGame) GetPhase() domain.TrexPhase {
	return _m.Called().Get(0).(domain.TrexPhase)
}

func (_m *MockTrexGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

func (_m *MockTrexGame) GetKingIdx() int { return _m.Called().Int(0) }

func (_m *MockTrexGame) GetContract() domain.TrexContract {
	return _m.Called().Get(0).(domain.TrexContract)
}

func (_m *MockTrexGame) AvailableContracts() []domain.TrexContract {
	if v, ok := _m.Called().Get(0).([]domain.TrexContract); ok {
		return v
	}
	return nil
}

func (_m *MockTrexGame) IsContractUsed(king int, contract domain.TrexContract) bool {
	return _m.Called(king, contract).Bool(0)
}

func (_m *MockTrexGame) IsTrix() bool { return _m.Called().Bool(0) }

func (_m *MockTrexGame) GetDealNumber() int { return _m.Called().Int(0) }

func (_m *MockTrexGame) GetTrick() []domain.TrexTrickCard {
	if v, ok := _m.Called().Get(0).([]domain.TrexTrickCard); ok {
		return v
	}
	return nil
}

func (_m *MockTrexGame) GetTrickNumber() int { return _m.Called().Int(0) }

func (_m *MockTrexGame) GetTricksWon(idx int) int { return _m.Called(idx).Int(0) }

func (_m *MockTrexGame) GetValidPlayIndices(player int) []int {
	if v, ok := _m.Called(player).Get(0).([]int); ok {
		return v
	}
	return nil
}

func (_m *MockTrexGame) GetSuitRun(suit int) (bool, int, int) {
	ret := _m.Called(suit)
	return ret.Bool(0), ret.Int(1), ret.Int(2)
}

func (_m *MockTrexGame) GetFinishOrder() []int {
	if v, ok := _m.Called().Get(0).([]int); ok {
		return v
	}
	return nil
}

func (_m *MockTrexGame) GetScore(idx int) int { return _m.Called(idx).Int(0) }

func (_m *MockTrexGame) GetDealScore(idx int) int { return _m.Called(idx).Int(0) }

func (_m *MockTrexGame) GetWinnerIdx() int { return _m.Called().Int(0) }

func (_m *MockTrexGame) GetPlayers() []*domain.TrexPlayer {
	if v, ok := _m.Called().Get(0).([]*domain.TrexPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockTrexGame) GetPlayer(idx int) *domain.TrexPlayer {
	if v, ok := _m.Called(idx).Get(0).(*domain.TrexPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockTrexGame) GetActionLog() []*domain.ActionLogEntry {
	if v, ok := _m.Called().Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
