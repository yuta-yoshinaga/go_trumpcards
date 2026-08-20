//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSambaGame モック
type MockSambaGame struct {
	mock.Mock
}

func (m *MockSambaGame) Reset()     { m.Called() }
func (m *MockSambaGame) NextRound() { m.Called() }
func (m *MockSambaGame) PlayerDrawFromStock() error {
	return m.Called().Error(0)
}
func (m *MockSambaGame) PlayerDrawFromDiscard(naturalPairIndices []int) error {
	return m.Called(naturalPairIndices).Error(0)
}
func (m *MockSambaGame) PlayerMeld(meldGroups [][]int) error {
	return m.Called(meldGroups).Error(0)
}
func (m *MockSambaGame) PlayerSkipMeld() error { return m.Called().Error(0) }
func (m *MockSambaGame) PlayerDiscard(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}
func (m *MockSambaGame) PlayerGoOut() error { return m.Called().Error(0) }
func (m *MockSambaGame) CpuPlay()           { m.Called() }
func (m *MockSambaGame) GetConfig() domain.SambaConfig {
	return m.Called().Get(0).(domain.SambaConfig)
}
func (m *MockSambaGame) SetConfig(cfg domain.SambaConfig) { m.Called(cfg) }
func (m *MockSambaGame) GetGameEndFlag() bool             { return m.Called().Bool(0) }
func (m *MockSambaGame) GetPhase() domain.SambaPhase {
	return m.Called().Get(0).(domain.SambaPhase)
}
func (m *MockSambaGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockSambaGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockSambaGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockSambaGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockSambaGame) GetDrawPileCount() int    { return m.Called().Int(0) }
func (m *MockSambaGame) GetDiscardPileCount() int { return m.Called().Int(0) }
func (m *MockSambaGame) GetIsFrozen() bool        { return m.Called().Bool(0) }
func (m *MockSambaGame) GetWinnerIdx() int        { return m.Called().Int(0) }
func (m *MockSambaGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockSambaGame) GetPlayer(i int) *domain.SambaPlayer {
	return m.Called(i).Get(0).(*domain.SambaPlayer)
}
func (m *MockSambaGame) GetTeamCount() int { return m.Called().Int(0) }
func (m *MockSambaGame) GetTeamScore(team int) int {
	return m.Called(team).Int(0)
}

// GetMinimumMeldValue モック
func (m *MockSambaGame) GetMinimumMeldValue(playerIdx int) int {
	return m.Called(playerIdx).Int(0)
}
func (m *MockSambaGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockSambaGame) GetDrewFromDiscard() bool { return m.Called().Bool(0) }
