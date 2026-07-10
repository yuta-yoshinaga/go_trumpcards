//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPanGame パングインゲゲームのモック
type MockPanGame struct {
	mock.Mock
}

func (m *MockPanGame) Reset()                       { m.Called() }
func (m *MockPanGame) NextRound()                   { m.Called() }
func (m *MockPanGame) PlayerDrawFromStock() error   { return m.Called().Error(0) }
func (m *MockPanGame) PlayerDrawFromDiscard() error { return m.Called().Error(0) }
func (m *MockPanGame) PlayerMeld(cardIndices []int) error {
	return m.Called(cardIndices).Error(0)
}
func (m *MockPanGame) PlayerLayoff(meldOwner, meldIdx, cardIndex int) error {
	return m.Called(meldOwner, meldIdx, cardIndex).Error(0)
}
func (m *MockPanGame) PlayerDiscard(cardIndex int) error { return m.Called(cardIndex).Error(0) }
func (m *MockPanGame) CpuPlay()                          { m.Called() }
func (m *MockPanGame) GetConfig() domain.PanConfig {
	return m.Called().Get(0).(domain.PanConfig)
}
func (m *MockPanGame) SetConfig(cfg domain.PanConfig) { m.Called(cfg) }
func (m *MockPanGame) GetGameEndFlag() bool           { return m.Called().Bool(0) }
func (m *MockPanGame) GetPhase() domain.PanPhase {
	return m.Called().Get(0).(domain.PanPhase)
}
func (m *MockPanGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockPanGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockPanGame) GetTargetRounds() int     { return m.Called().Int(0) }
func (m *MockPanGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockPanGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockPanGame) GetDiscardPile() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockPanGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockPanGame) GetDrawPileCount() int { return m.Called().Int(0) }
func (m *MockPanGame) GetWinnerIdx() int     { return m.Called().Int(0) }
func (m *MockPanGame) GetPlayerCnt() int     { return m.Called().Int(0) }
func (m *MockPanGame) GetPlayer(i int) *domain.PanPlayer {
	return m.Called(i).Get(0).(*domain.PanPlayer)
}
func (m *MockPanGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockPanGame) GetPanDeclarerIdx() int      { return m.Called().Int(0) }
func (m *MockPanGame) PlayerHandPoints(i int) int  { return m.Called(i).Int(0) }
func (m *MockPanGame) PlayerMeldedCount(i int) int { return m.Called(i).Int(0) }
