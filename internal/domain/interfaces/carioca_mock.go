//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCariocaGame カリオカゲームのモック
type MockCariocaGame struct {
	mock.Mock
}

func (m *MockCariocaGame) Reset()                       { m.Called() }
func (m *MockCariocaGame) NextRound()                   { m.Called() }
func (m *MockCariocaGame) PlayerDrawFromStock() error   { return m.Called().Error(0) }
func (m *MockCariocaGame) PlayerDrawFromDiscard() error { return m.Called().Error(0) }
func (m *MockCariocaGame) PlayerMeldContract(idx [][]int) error {
	return m.Called(idx).Error(0)
}
func (m *MockCariocaGame) PlayerMeldExtra(idx []int) error { return m.Called(idx).Error(0) }
func (m *MockCariocaGame) PlayerLayoff(t, mi, ci int) error {
	return m.Called(t, mi, ci).Error(0)
}
func (m *MockCariocaGame) PlayerDiscard(i int) error { return m.Called(i).Error(0) }
func (m *MockCariocaGame) CpuPlay()                  { m.Called() }
func (m *MockCariocaGame) GetConfig() domain.CariocaConfig {
	return m.Called().Get(0).(domain.CariocaConfig)
}
func (m *MockCariocaGame) SetConfig(cfg domain.CariocaConfig) { m.Called(cfg) }
func (m *MockCariocaGame) GetGameEndFlag() bool               { return m.Called().Bool(0) }
func (m *MockCariocaGame) GetPhase() domain.CariocaPhase {
	return m.Called().Get(0).(domain.CariocaPhase)
}
func (m *MockCariocaGame) IsHumanTurn() bool   { return m.Called().Bool(0) }
func (m *MockCariocaGame) GetRoundNumber() int { return m.Called().Int(0) }
func (m *MockCariocaGame) GetCurrentContract() domain.Contract {
	return m.Called().Get(0).(domain.Contract)
}
func (m *MockCariocaGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockCariocaGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockCariocaGame) GetDiscardPile() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockCariocaGame) GetDrawPileCount() int { return m.Called().Int(0) }
func (m *MockCariocaGame) GetWinnerIdx() int     { return m.Called().Int(0) }
func (m *MockCariocaGame) GetPlayerCnt() int     { return m.Called().Int(0) }
func (m *MockCariocaGame) GetPlayer(i int) *domain.CariocaPlayer {
	return m.Called(i).Get(0).(*domain.CariocaPlayer)
}
func (m *MockCariocaGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockCariocaGame) GetRoundWinnerIdx() int { return m.Called().Int(0) }
