//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockIndianRummyGame インドラミーゲームのモック
type MockIndianRummyGame struct {
	mock.Mock
}

func (m *MockIndianRummyGame) Reset()                       { m.Called() }
func (m *MockIndianRummyGame) NextRound()                   { m.Called() }
func (m *MockIndianRummyGame) PlayerDrawFromStock() error   { return m.Called().Error(0) }
func (m *MockIndianRummyGame) PlayerDrawFromDiscard() error { return m.Called().Error(0) }
func (m *MockIndianRummyGame) PlayerDiscard(i int) error    { return m.Called(i).Error(0) }
func (m *MockIndianRummyGame) PlayerDeclare(i int) error    { return m.Called(i).Error(0) }
func (m *MockIndianRummyGame) CpuPlay()                     { m.Called() }
func (m *MockIndianRummyGame) GetConfig() domain.IndianRummyConfig {
	return m.Called().Get(0).(domain.IndianRummyConfig)
}
func (m *MockIndianRummyGame) SetConfig(cfg domain.IndianRummyConfig) { m.Called(cfg) }
func (m *MockIndianRummyGame) GetGameEndFlag() bool                   { return m.Called().Bool(0) }
func (m *MockIndianRummyGame) GetPhase() domain.IndianRummyPhase {
	return m.Called().Get(0).(domain.IndianRummyPhase)
}
func (m *MockIndianRummyGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockIndianRummyGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockIndianRummyGame) GetTargetRounds() int     { return m.Called().Int(0) }
func (m *MockIndianRummyGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockIndianRummyGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockIndianRummyGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockIndianRummyGame) GetDrawPileCount() int { return m.Called().Int(0) }
func (m *MockIndianRummyGame) GetWildJoker() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockIndianRummyGame) GetWildRank() int  { return m.Called().Int(0) }
func (m *MockIndianRummyGame) GetWinnerIdx() int { return m.Called().Int(0) }
func (m *MockIndianRummyGame) GetPlayerCnt() int { return m.Called().Int(0) }
func (m *MockIndianRummyGame) GetPlayer(i int) *domain.IndianRummyPlayer {
	return m.Called(i).Get(0).(*domain.IndianRummyPlayer)
}
func (m *MockIndianRummyGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockIndianRummyGame) GetDeclarerIdx() int           { return m.Called().Int(0) }
func (m *MockIndianRummyGame) GetDeclarationValid() bool     { return m.Called().Bool(0) }
func (m *MockIndianRummyGame) PlayerDeadwoodValue(i int) int { return m.Called(i).Int(0) }
func (m *MockIndianRummyGame) PlayerHasPureSequence(i int) bool {
	return m.Called(i).Bool(0)
}
