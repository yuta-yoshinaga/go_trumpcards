//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockThreeThirteenGame スリー・サーティーンゲームのモック
type MockThreeThirteenGame struct {
	mock.Mock
}

func (m *MockThreeThirteenGame) Reset()                       { m.Called() }
func (m *MockThreeThirteenGame) NextRound()                   { m.Called() }
func (m *MockThreeThirteenGame) PlayerDrawFromStock() error   { return m.Called().Error(0) }
func (m *MockThreeThirteenGame) PlayerDrawFromDiscard() error { return m.Called().Error(0) }
func (m *MockThreeThirteenGame) PlayerDiscard(i int) error    { return m.Called(i).Error(0) }
func (m *MockThreeThirteenGame) PlayerKnock(i int) error      { return m.Called(i).Error(0) }
func (m *MockThreeThirteenGame) CpuPlay()                     { m.Called() }
func (m *MockThreeThirteenGame) GetConfig() domain.ThreeThirteenConfig {
	return m.Called().Get(0).(domain.ThreeThirteenConfig)
}
func (m *MockThreeThirteenGame) SetConfig(cfg domain.ThreeThirteenConfig) { m.Called(cfg) }
func (m *MockThreeThirteenGame) GetGameEndFlag() bool                     { return m.Called().Bool(0) }
func (m *MockThreeThirteenGame) GetPhase() domain.ThreeThirteenPhase {
	return m.Called().Get(0).(domain.ThreeThirteenPhase)
}
func (m *MockThreeThirteenGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockThreeThirteenGame) GetRound() int            { return m.Called().Int(0) }
func (m *MockThreeThirteenGame) WildRank() int            { return m.Called().Int(0) }
func (m *MockThreeThirteenGame) GetDealCount() int        { return m.Called().Int(0) }
func (m *MockThreeThirteenGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockThreeThirteenGame) GetKnockerIdx() int       { return m.Called().Int(0) }
func (m *MockThreeThirteenGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockThreeThirteenGame) GetDiscardPile() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockThreeThirteenGame) GetDrawPileCount() int { return m.Called().Int(0) }
func (m *MockThreeThirteenGame) GetWinnerIdx() int     { return m.Called().Int(0) }
func (m *MockThreeThirteenGame) GetPlayerCnt() int     { return m.Called().Int(0) }
func (m *MockThreeThirteenGame) GetPlayer(i int) *domain.ThreeThirteenPlayer {
	return m.Called(i).Get(0).(*domain.ThreeThirteenPlayer)
}
func (m *MockThreeThirteenGame) GetPlayerDeadwoodValue(i int) int { return m.Called(i).Int(0) }

func (m *MockThreeThirteenGame) GetDeadwoodAfterDiscard(playerIdx, cardIndex int) int {
	return m.Called(playerIdx, cardIndex).Int(0)
}
func (m *MockThreeThirteenGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
