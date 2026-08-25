//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBoliviaGame モック
type MockBoliviaGame struct {
	mock.Mock
}

func (m *MockBoliviaGame) Reset()     { m.Called() }
func (m *MockBoliviaGame) NextRound() { m.Called() }
func (m *MockBoliviaGame) PlayerDrawFromStock() error {
	return m.Called().Error(0)
}
func (m *MockBoliviaGame) PlayerDrawFromDiscard(naturalPairIndices []int) error {
	return m.Called(naturalPairIndices).Error(0)
}
func (m *MockBoliviaGame) PlayerMeld(meldGroups [][]int) error {
	return m.Called(meldGroups).Error(0)
}
func (m *MockBoliviaGame) PlayerSkipMeld() error { return m.Called().Error(0) }
func (m *MockBoliviaGame) PlayerDiscard(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}
func (m *MockBoliviaGame) PlayerGoOut() error { return m.Called().Error(0) }
func (m *MockBoliviaGame) CpuPlay()           { m.Called() }
func (m *MockBoliviaGame) GetConfig() domain.BoliviaConfig {
	return m.Called().Get(0).(domain.BoliviaConfig)
}
func (m *MockBoliviaGame) SetConfig(cfg domain.BoliviaConfig) { m.Called(cfg) }
func (m *MockBoliviaGame) GetGameEndFlag() bool               { return m.Called().Bool(0) }
func (m *MockBoliviaGame) GetPhase() domain.BoliviaPhase {
	return m.Called().Get(0).(domain.BoliviaPhase)
}
func (m *MockBoliviaGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockBoliviaGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockBoliviaGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockBoliviaGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockBoliviaGame) GetDrawPileCount() int    { return m.Called().Int(0) }
func (m *MockBoliviaGame) GetDiscardPileCount() int { return m.Called().Int(0) }
func (m *MockBoliviaGame) GetIsFrozen() bool        { return m.Called().Bool(0) }
func (m *MockBoliviaGame) GetWinnerIdx() int        { return m.Called().Int(0) }
func (m *MockBoliviaGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockBoliviaGame) GetPlayer(i int) *domain.BoliviaPlayer {
	return m.Called(i).Get(0).(*domain.BoliviaPlayer)
}
func (m *MockBoliviaGame) GetTeamCount() int { return m.Called().Int(0) }
func (m *MockBoliviaGame) GetTeamScore(team int) int {
	return m.Called(team).Int(0)
}

// GetMinimumMeldValue モック
func (m *MockBoliviaGame) GetMinimumMeldValue(playerIdx int) int {
	return m.Called(playerIdx).Int(0)
}
func (m *MockBoliviaGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockBoliviaGame) GetDrewFromDiscard() bool { return m.Called().Bool(0) }
