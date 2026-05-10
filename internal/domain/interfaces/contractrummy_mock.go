//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockContractRummyGame コントラクトラミーゲームのモック
type MockContractRummyGame struct {
	mock.Mock
}

func (m *MockContractRummyGame) Reset()                       { m.Called() }
func (m *MockContractRummyGame) NextRound()                   { m.Called() }
func (m *MockContractRummyGame) PlayerDrawFromStock() error   { return m.Called().Error(0) }
func (m *MockContractRummyGame) PlayerDrawFromDiscard() error { return m.Called().Error(0) }
func (m *MockContractRummyGame) PlayerMeldContract(idx [][]int) error {
	return m.Called(idx).Error(0)
}
func (m *MockContractRummyGame) PlayerMeldExtra(idx []int) error { return m.Called(idx).Error(0) }
func (m *MockContractRummyGame) PlayerLayoff(t, mi, ci int) error {
	return m.Called(t, mi, ci).Error(0)
}
func (m *MockContractRummyGame) PlayerDiscard(i int) error { return m.Called(i).Error(0) }
func (m *MockContractRummyGame) CpuPlay()                  { m.Called() }
func (m *MockContractRummyGame) GetConfig() domain.ContractRummyConfig {
	return m.Called().Get(0).(domain.ContractRummyConfig)
}
func (m *MockContractRummyGame) SetConfig(cfg domain.ContractRummyConfig) { m.Called(cfg) }
func (m *MockContractRummyGame) GetGameEndFlag() bool                     { return m.Called().Bool(0) }
func (m *MockContractRummyGame) GetPhase() domain.ContractRummyPhase {
	return m.Called().Get(0).(domain.ContractRummyPhase)
}
func (m *MockContractRummyGame) IsHumanTurn() bool   { return m.Called().Bool(0) }
func (m *MockContractRummyGame) GetRoundNumber() int { return m.Called().Int(0) }
func (m *MockContractRummyGame) GetCurrentContract() domain.Contract {
	return m.Called().Get(0).(domain.Contract)
}
func (m *MockContractRummyGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockContractRummyGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockContractRummyGame) GetDiscardPile() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockContractRummyGame) GetDrawPileCount() int { return m.Called().Int(0) }
func (m *MockContractRummyGame) GetWinnerIdx() int     { return m.Called().Int(0) }
func (m *MockContractRummyGame) GetPlayerCnt() int     { return m.Called().Int(0) }
func (m *MockContractRummyGame) GetPlayer(i int) *domain.ContractRummyPlayer {
	return m.Called(i).Get(0).(*domain.ContractRummyPlayer)
}
func (m *MockContractRummyGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockContractRummyGame) GetRoundWinnerIdx() int { return m.Called().Int(0) }
