//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMacauGame モック
type MockMacauGame struct {
	mock.Mock
}

func (m *MockMacauGame) Reset()                          { m.Called() }
func (m *MockMacauGame) NextRound()                      { m.Called() }
func (m *MockMacauGame) PlayerPlay(cardIndex int) error  { return m.Called(cardIndex).Error(0) }
func (m *MockMacauGame) PlayerChooseSuit(suit int) error { return m.Called(suit).Error(0) }
func (m *MockMacauGame) PlayerDraw() error               { return m.Called().Error(0) }
func (m *MockMacauGame) PlayerDeclare() error            { return m.Called().Error(0) }
func (m *MockMacauGame) PlayerSkipDeclare() error        { return m.Called().Error(0) }
func (m *MockMacauGame) CpuPlay()                        { m.Called() }
func (m *MockMacauGame) CpuChooseSuit()                  { m.Called() }
func (m *MockMacauGame) CpuDeclare()                     { m.Called() }
func (m *MockMacauGame) ScoreRound()                     { m.Called() }
func (m *MockMacauGame) GetConfig() domain.MacauConfig {
	return m.Called().Get(0).(domain.MacauConfig)
}
func (m *MockMacauGame) SetConfig(cfg domain.MacauConfig) { m.Called(cfg) }
func (m *MockMacauGame) GetGameEndFlag() bool             { return m.Called().Bool(0) }
func (m *MockMacauGame) GetPhase() domain.MacauPhase {
	return m.Called().Get(0).(domain.MacauPhase)
}
func (m *MockMacauGame) IsHumanTurn() bool           { return m.Called().Bool(0) }
func (m *MockMacauGame) GetRoundNumber() int         { return m.Called().Int(0) }
func (m *MockMacauGame) GetCurrentPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockMacauGame) GetDiscardTop() *domain.Card { return m.Called().Get(0).(*domain.Card) }
func (m *MockMacauGame) GetDrawPileCount() int       { return m.Called().Int(0) }
func (m *MockMacauGame) GetChosenSuit() int          { return m.Called().Int(0) }
func (m *MockMacauGame) GetPenaltyDrawCount() int    { return m.Called().Int(0) }
func (m *MockMacauGame) GetDirection() int           { return m.Called().Int(0) }
func (m *MockMacauGame) GetWinnerIdx() int           { return m.Called().Int(0) }
func (m *MockMacauGame) GetPlayerCnt() int           { return m.Called().Int(0) }
func (m *MockMacauGame) GetPlayer(i int) *domain.MacauPlayer {
	return m.Called(i).Get(0).(*domain.MacauPlayer)
}
func (m *MockMacauGame) IsValidPlay(card *domain.Card) bool {
	return m.Called(card).Bool(0)
}
func (m *MockMacauGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
