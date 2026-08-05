//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockConquianGame モック
type MockConquianGame struct {
	mock.Mock
}

func (m *MockConquianGame) Reset()                       { m.Called() }
func (m *MockConquianGame) NextRound()                   { m.Called() }
func (m *MockConquianGame) PlayerDrawFromStock() error   { return m.Called().Error(0) }
func (m *MockConquianGame) PlayerDrawFromDiscard() error { return m.Called().Error(0) }
func (m *MockConquianGame) PlayerMeldWithTargets(meldGroups [][]int, extendTargets []int) error {
	return m.Called(meldGroups, extendTargets).Error(0)
}

func (m *MockConquianGame) GetExtendableMeldIndices(playerIdx int, card *domain.Card) []int {
	out, _ := m.Called(playerIdx, card).Get(0).([]int)
	return out
}

func (m *MockConquianGame) PlayerMeld(meldGroups [][]int) error {
	return m.Called(meldGroups).Error(0)
}
func (m *MockConquianGame) PlayerDiscard(cardIndex int) error { return m.Called(cardIndex).Error(0) }
func (m *MockConquianGame) CpuPlay()                          { m.Called() }
func (m *MockConquianGame) GetConfig() domain.ConquianConfig {
	return m.Called().Get(0).(domain.ConquianConfig)
}
func (m *MockConquianGame) SetConfig(cfg domain.ConquianConfig) { m.Called(cfg) }
func (m *MockConquianGame) GetGameEndFlag() bool                { return m.Called().Bool(0) }
func (m *MockConquianGame) GetPhase() domain.ConquianPhase {
	return m.Called().Get(0).(domain.ConquianPhase)
}
func (m *MockConquianGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockConquianGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockConquianGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockConquianGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockConquianGame) GetDrawPileCount() int  { return m.Called().Int(0) }
func (m *MockConquianGame) GetWinnerIdx() int      { return m.Called().Int(0) }
func (m *MockConquianGame) GetRoundWinnerIdx() int { return m.Called().Int(0) }
func (m *MockConquianGame) GetPlayerCnt() int      { return m.Called().Int(0) }
func (m *MockConquianGame) GetPlayer(i int) *domain.ConquianPlayer {
	return m.Called(i).Get(0).(*domain.ConquianPlayer)
}
func (m *MockConquianGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockConquianGame) GetTookDiscard() bool { return m.Called().Bool(0) }
