//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMarriageGame マリッジゲームのモック
type MockMarriageGame struct {
	mock.Mock
}

func (m *MockMarriageGame) Reset()                       { m.Called() }
func (m *MockMarriageGame) NextRound()                   { m.Called() }
func (m *MockMarriageGame) PlayerDrawFromStock() error   { return m.Called().Error(0) }
func (m *MockMarriageGame) PlayerDrawFromDiscard() error { return m.Called().Error(0) }
func (m *MockMarriageGame) PlayerDiscard(i int) error    { return m.Called(i).Error(0) }
func (m *MockMarriageGame) PlayerDeclare(i int) error    { return m.Called(i).Error(0) }
func (m *MockMarriageGame) CpuPlay()                     { m.Called() }
func (m *MockMarriageGame) GetConfig() domain.MarriageConfig {
	return m.Called().Get(0).(domain.MarriageConfig)
}
func (m *MockMarriageGame) SetConfig(cfg domain.MarriageConfig) { m.Called(cfg) }
func (m *MockMarriageGame) GetGameEndFlag() bool                { return m.Called().Bool(0) }
func (m *MockMarriageGame) GetPhase() domain.MarriagePhase {
	return m.Called().Get(0).(domain.MarriagePhase)
}
func (m *MockMarriageGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockMarriageGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockMarriageGame) GetTargetRounds() int     { return m.Called().Int(0) }
func (m *MockMarriageGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockMarriageGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockMarriageGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockMarriageGame) GetDrawPileCount() int { return m.Called().Int(0) }
func (m *MockMarriageGame) GetWildJoker() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockMarriageGame) GetWildRank() int  { return m.Called().Int(0) }
func (m *MockMarriageGame) GetWinnerIdx() int { return m.Called().Int(0) }
func (m *MockMarriageGame) GetPlayerCnt() int { return m.Called().Int(0) }
func (m *MockMarriageGame) GetPlayer(i int) *domain.MarriagePlayer {
	return m.Called(i).Get(0).(*domain.MarriagePlayer)
}
func (m *MockMarriageGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockMarriageGame) GetDeclarerIdx() int           { return m.Called().Int(0) }
func (m *MockMarriageGame) GetDeclarationValid() bool     { return m.Called().Bool(0) }
func (m *MockMarriageGame) PlayerDeadwoodValue(i int) int { return m.Called(i).Int(0) }
func (m *MockMarriageGame) PlayerMaalValue(i int) int     { return m.Called(i).Int(0) }
func (m *MockMarriageGame) PlayerHasPureSequence(i int) bool {
	return m.Called(i).Bool(0)
}
