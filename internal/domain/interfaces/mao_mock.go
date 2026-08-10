//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMaoGame モック
type MockMaoGame struct {
	mock.Mock
}

func (m *MockMaoGame) Reset()                          { m.Called() }
func (m *MockMaoGame) NextRound()                      { m.Called() }
func (m *MockMaoGame) PlayerPlay(cardIndex int) error  { return m.Called(cardIndex).Error(0) }
func (m *MockMaoGame) PlayerChooseSuit(suit int) error { return m.Called(suit).Error(0) }
func (m *MockMaoGame) PlayerDraw() error               { return m.Called().Error(0) }
func (m *MockMaoGame) PlayerDeclare() error            { return m.Called().Error(0) }
func (m *MockMaoGame) PlayerSkipDeclare() error        { return m.Called().Error(0) }
func (m *MockMaoGame) PlayerDeclareWord(word string) error {
	return m.Called(word).Error(0)
}
func (m *MockMaoGame) CpuPlay()       { m.Called() }
func (m *MockMaoGame) CpuChooseSuit() { m.Called() }
func (m *MockMaoGame) CpuDeclare()    { m.Called() }
func (m *MockMaoGame) ScoreRound()    { m.Called() }
func (m *MockMaoGame) GetConfig() domain.MaoConfig {
	return m.Called().Get(0).(domain.MaoConfig)
}
func (m *MockMaoGame) SetConfig(cfg domain.MaoConfig) { m.Called(cfg) }
func (m *MockMaoGame) GetGameEndFlag() bool           { return m.Called().Bool(0) }
func (m *MockMaoGame) GetPhase() domain.MaoPhase {
	return m.Called().Get(0).(domain.MaoPhase)
}
func (m *MockMaoGame) IsHumanTurn() bool           { return m.Called().Bool(0) }
func (m *MockMaoGame) GetRoundNumber() int         { return m.Called().Int(0) }
func (m *MockMaoGame) GetCurrentPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockMaoGame) GetDiscardTop() *domain.Card { return m.Called().Get(0).(*domain.Card) }
func (m *MockMaoGame) GetDrawPileCount() int       { return m.Called().Int(0) }
func (m *MockMaoGame) GetChosenSuit() int          { return m.Called().Int(0) }
func (m *MockMaoGame) GetPenaltyDrawCount() int    { return m.Called().Int(0) }
func (m *MockMaoGame) GetDirection() int           { return m.Called().Int(0) }
func (m *MockMaoGame) GetWinnerIdx() int           { return m.Called().Int(0) }
func (m *MockMaoGame) GetPlayerCnt() int           { return m.Called().Int(0) }
func (m *MockMaoGame) GetPlayer(i int) *domain.MaoPlayer {
	return m.Called(i).Get(0).(*domain.MaoPlayer)
}
func (m *MockMaoGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockMaoGame) GetAwaitingWord() bool      { return m.Called().Bool(0) }
func (m *MockMaoGame) GetPlayerCorrectCount() int { return m.Called().Int(0) }
func (m *MockMaoGame) GetHintUnlocked() bool      { return m.Called().Bool(0) }
func (m *MockMaoGame) GetRuleHintKey() string     { return m.Called().String(0) }
func (m *MockMaoGame) GetRulePenaltyFlag() bool   { return m.Called().Bool(0) }
