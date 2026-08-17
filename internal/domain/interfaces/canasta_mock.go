//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCanastaGame モック
type MockCanastaGame struct {
	mock.Mock
}

func (m *MockCanastaGame) Reset()     { m.Called() }
func (m *MockCanastaGame) NextRound() { m.Called() }
func (m *MockCanastaGame) PlayerDrawFromStock() error {
	return m.Called().Error(0)
}
func (m *MockCanastaGame) PlayerDrawFromDiscard(naturalPairIndices []int) error {
	return m.Called(naturalPairIndices).Error(0)
}
func (m *MockCanastaGame) PlayerMeld(meldGroups [][]int) error {
	return m.Called(meldGroups).Error(0)
}
func (m *MockCanastaGame) PlayerSkipMeld() error { return m.Called().Error(0) }
func (m *MockCanastaGame) PlayerDiscard(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}
func (m *MockCanastaGame) PlayerGoOut() error { return m.Called().Error(0) }
func (m *MockCanastaGame) CpuPlay()           { m.Called() }
func (m *MockCanastaGame) GetConfig() domain.CanastaConfig {
	return m.Called().Get(0).(domain.CanastaConfig)
}
func (m *MockCanastaGame) SetConfig(cfg domain.CanastaConfig) { m.Called(cfg) }
func (m *MockCanastaGame) GetGameEndFlag() bool               { return m.Called().Bool(0) }
func (m *MockCanastaGame) GetPhase() domain.CanastaPhase {
	return m.Called().Get(0).(domain.CanastaPhase)
}
func (m *MockCanastaGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockCanastaGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockCanastaGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockCanastaGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockCanastaGame) GetDiscardPile() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockCanastaGame) GetDrawPileCount() int    { return m.Called().Int(0) }
func (m *MockCanastaGame) GetDiscardPileCount() int { return m.Called().Int(0) }
func (m *MockCanastaGame) GetPozzettoCount() int    { return m.Called().Int(0) }
func (m *MockCanastaGame) GetIsFrozen() bool        { return m.Called().Bool(0) }

// GetDrawFromDiscardBlocker モック
func (m *MockCanastaGame) GetDrawFromDiscardBlocker() string { return m.Called().String(0) }
func (m *MockCanastaGame) GetWinnerIdx() int                 { return m.Called().Int(0) }
func (m *MockCanastaGame) GetPlayerCnt() int                 { return m.Called().Int(0) }
func (m *MockCanastaGame) GetPlayer(i int) *domain.CanastaPlayer {
	return m.Called(i).Get(0).(*domain.CanastaPlayer)
}
func (m *MockCanastaGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockCanastaGame) GetDrewFromDiscard() bool { return m.Called().Bool(0) }
func (m *MockCanastaGame) GetHint() *domain.CanastaHint {
	ret := m.Called().Get(0)
	if ret == nil {
		return nil
	}
	return ret.(*domain.CanastaHint)
}
