//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBurracoGame モック
type MockBurracoGame struct {
	mock.Mock
}

func (m *MockBurracoGame) Reset()     { m.Called() }
func (m *MockBurracoGame) NextRound() { m.Called() }
func (m *MockBurracoGame) PlayerDrawFromStock() error {
	return m.Called().Error(0)
}
func (m *MockBurracoGame) PlayerDrawFromDiscard(naturalPairIndices []int) error {
	return m.Called(naturalPairIndices).Error(0)
}
func (m *MockBurracoGame) PlayerMeld(meldGroups [][]int) error {
	return m.Called(meldGroups).Error(0)
}
func (m *MockBurracoGame) PlayerSkipMeld() error { return m.Called().Error(0) }
func (m *MockBurracoGame) PlayerDiscard(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}
func (m *MockBurracoGame) PlayerGoOut() error { return m.Called().Error(0) }
func (m *MockBurracoGame) CpuPlay()           { m.Called() }
func (m *MockBurracoGame) GetConfig() domain.BurracoConfig {
	return m.Called().Get(0).(domain.BurracoConfig)
}
func (m *MockBurracoGame) SetConfig(cfg domain.BurracoConfig) { m.Called(cfg) }
func (m *MockBurracoGame) GetGameEndFlag() bool               { return m.Called().Bool(0) }
func (m *MockBurracoGame) GetPhase() domain.BurracoPhase {
	return m.Called().Get(0).(domain.BurracoPhase)
}
func (m *MockBurracoGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockBurracoGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockBurracoGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockBurracoGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockBurracoGame) GetDrawPileCount() int    { return m.Called().Int(0) }
func (m *MockBurracoGame) GetDiscardPileCount() int { return m.Called().Int(0) }
func (m *MockBurracoGame) GetPozzettoCount() int    { return m.Called().Int(0) }
func (m *MockBurracoGame) GetIsFrozen() bool        { return m.Called().Bool(0) }
func (m *MockBurracoGame) GetWinnerIdx() int        { return m.Called().Int(0) }
func (m *MockBurracoGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockBurracoGame) GetPlayer(i int) *domain.BurracoPlayer {
	return m.Called(i).Get(0).(*domain.BurracoPlayer)
}
func (m *MockBurracoGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockBurracoGame) GetDrewFromDiscard() bool { return m.Called().Bool(0) }
