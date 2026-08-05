//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHandAndFootGame モック
type MockHandAndFootGame struct {
	mock.Mock
}

func (m *MockHandAndFootGame) Reset()     { m.Called() }
func (m *MockHandAndFootGame) NextRound() { m.Called() }
func (m *MockHandAndFootGame) PlayerDrawFromStock() error {
	return m.Called().Error(0)
}
func (m *MockHandAndFootGame) PlayerDrawFromDiscard(naturalPairIndices []int) error {
	return m.Called(naturalPairIndices).Error(0)
}
func (m *MockHandAndFootGame) PlayerMeld(meldGroups [][]int) error {
	return m.Called(meldGroups).Error(0)
}
func (m *MockHandAndFootGame) SuggestMelds(playerIdx int) [][]*domain.Card {
	ret := m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([][]*domain.Card)
	}
	return nil
}
func (m *MockHandAndFootGame) PlayerSkipMeld() error { return m.Called().Error(0) }
func (m *MockHandAndFootGame) PlayerDiscard(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}
func (m *MockHandAndFootGame) PlayerGoOut() error { return m.Called().Error(0) }
func (m *MockHandAndFootGame) CpuPlay()           { m.Called() }
func (m *MockHandAndFootGame) GetConfig() domain.HandAndFootConfig {
	return m.Called().Get(0).(domain.HandAndFootConfig)
}
func (m *MockHandAndFootGame) SetConfig(cfg domain.HandAndFootConfig) { m.Called(cfg) }
func (m *MockHandAndFootGame) GetGameEndFlag() bool                   { return m.Called().Bool(0) }
func (m *MockHandAndFootGame) GetPhase() domain.HandAndFootPhase {
	return m.Called().Get(0).(domain.HandAndFootPhase)
}
func (m *MockHandAndFootGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockHandAndFootGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockHandAndFootGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockHandAndFootGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockHandAndFootGame) GetDrawPileCount() int    { return m.Called().Int(0) }
func (m *MockHandAndFootGame) GetDiscardPileCount() int { return m.Called().Int(0) }
func (m *MockHandAndFootGame) GetIsFrozen() bool        { return m.Called().Bool(0) }
func (m *MockHandAndFootGame) GetWinnerTeam() int       { return m.Called().Int(0) }
func (m *MockHandAndFootGame) GetWinnerIdx() int        { return m.Called().Int(0) }
func (m *MockHandAndFootGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockHandAndFootGame) GetPlayer(i int) *domain.HandAndFootPlayer {
	return m.Called(i).Get(0).(*domain.HandAndFootPlayer)
}
func (m *MockHandAndFootGame) GetGoOutStatus(playerIdx int) domain.HandAndFootGoOutStatus {
	st, _ := m.Called(playerIdx).Get(0).(domain.HandAndFootGoOutStatus)
	return st
}

func (m *MockHandAndFootGame) GetTeamMelds(team int) []*domain.CanastaMeld {
	return m.Called(team).Get(0).([]*domain.CanastaMeld)
}
func (m *MockHandAndFootGame) GetTeamRed3s(team int) []*domain.Card {
	return m.Called(team).Get(0).([]*domain.Card)
}
func (m *MockHandAndFootGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockHandAndFootGame) GetDrewFromDiscard() bool { return m.Called().Bool(0) }
