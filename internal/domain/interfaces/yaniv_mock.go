//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockYanivGame モック
type MockYanivGame struct {
	mock.Mock
}

func (m *MockYanivGame) Reset()     { m.Called() }
func (m *MockYanivGame) NextRound() { m.Called() }
func (m *MockYanivGame) PlayerDiscard(cardIndices []int) error {
	return m.Called(cardIndices).Error(0)
}
func (m *MockYanivGame) PlayerDeclareYaniv() error  { return m.Called().Error(0) }
func (m *MockYanivGame) PlayerDrawFromStock() error { return m.Called().Error(0) }
func (m *MockYanivGame) PlayerDrawFromPickup(end int) error {
	return m.Called(end).Error(0)
}
func (m *MockYanivGame) CpuPlay() { m.Called() }
func (m *MockYanivGame) GetConfig() domain.YanivConfig {
	return m.Called().Get(0).(domain.YanivConfig)
}
func (m *MockYanivGame) SetConfig(cfg domain.YanivConfig) { m.Called(cfg) }
func (m *MockYanivGame) GetGameEndFlag() bool             { return m.Called().Bool(0) }
func (m *MockYanivGame) GetPhase() domain.YanivPhase {
	return m.Called().Get(0).(domain.YanivPhase)
}
func (m *MockYanivGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockYanivGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockYanivGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockYanivGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockYanivGame) GetPickupCards() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockYanivGame) GetDrawPileCount() int { return m.Called().Int(0) }
func (m *MockYanivGame) GetWinnerIdx() int     { return m.Called().Int(0) }
func (m *MockYanivGame) GetPlayerCnt() int     { return m.Called().Int(0) }
func (m *MockYanivGame) GetPlayer(i int) *domain.YanivPlayer {
	return m.Called(i).Get(0).(*domain.YanivPlayer)
}
func (m *MockYanivGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockYanivGame) GetCallerIdx() int     { return m.Called().Int(0) }
func (m *MockYanivGame) GetAsafWinnerIdx() int { return m.Called().Int(0) }
func (m *MockYanivGame) GetIsAsaf() bool       { return m.Called().Bool(0) }
func (m *MockYanivGame) GetRoundScores() []int {
	return m.Called().Get(0).([]int)
}
