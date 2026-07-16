//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockChinchonGame モック
type MockChinchonGame struct {
	mock.Mock
}

func (m *MockChinchonGame) Reset()                       { m.Called() }
func (m *MockChinchonGame) NextRound()                   { m.Called() }
func (m *MockChinchonGame) PlayerDrawFromStock() error   { return m.Called().Error(0) }
func (m *MockChinchonGame) PlayerDrawFromDiscard() error { return m.Called().Error(0) }
func (m *MockChinchonGame) PlayerDiscard(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}
func (m *MockChinchonGame) PlayerKnock(cardIndex int) error { return m.Called(cardIndex).Error(0) }
func (m *MockChinchonGame) PlayerLayoff(cardIndices []int) error {
	return m.Called(cardIndices).Error(0)
}
func (m *MockChinchonGame) CpuPlay() { m.Called() }
func (m *MockChinchonGame) GetConfig() domain.ChinchonConfig {
	return m.Called().Get(0).(domain.ChinchonConfig)
}
func (m *MockChinchonGame) SetConfig(cfg domain.ChinchonConfig) { m.Called(cfg) }
func (m *MockChinchonGame) GetGameEndFlag() bool                { return m.Called().Bool(0) }
func (m *MockChinchonGame) GetPhase() domain.ChinchonPhase {
	return m.Called().Get(0).(domain.ChinchonPhase)
}
func (m *MockChinchonGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockChinchonGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockChinchonGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockChinchonGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockChinchonGame) GetDrawPileCount() int { return m.Called().Int(0) }
func (m *MockChinchonGame) GetWinnerIdx() int     { return m.Called().Int(0) }
func (m *MockChinchonGame) GetPlayerCnt() int     { return m.Called().Int(0) }
func (m *MockChinchonGame) GetPlayer(i int) *domain.ChinchonPlayer {
	return m.Called(i).Get(0).(*domain.ChinchonPlayer)
}
func (m *MockChinchonGame) GetPlayerDeadwoodValue(i int) int { return m.Called(i).Int(0) }
func (m *MockChinchonGame) GetKnockThreshold() int           { return m.Called().Int(0) }
func (m *MockChinchonGame) GetKnockerIdx() int               { return m.Called().Int(0) }
func (m *MockChinchonGame) GetKnockerMelds() [][]*domain.Card {
	return m.Called().Get(0).([][]*domain.Card)
}
func (m *MockChinchonGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
