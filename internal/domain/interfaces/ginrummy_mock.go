//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGinRummyGame モック
type MockGinRummyGame struct {
	mock.Mock
}

func (m *MockGinRummyGame) Reset()                            { m.Called() }
func (m *MockGinRummyGame) NextRound()                        { m.Called() }
func (m *MockGinRummyGame) PlayerDrawFromStock() error        { return m.Called().Error(0) }
func (m *MockGinRummyGame) PlayerDrawFromDiscard() error      { return m.Called().Error(0) }
func (m *MockGinRummyGame) PlayerDiscard(cardIndex int) error { return m.Called(cardIndex).Error(0) }
func (m *MockGinRummyGame) PlayerKnock(cardIndex int) error   { return m.Called(cardIndex).Error(0) }
func (m *MockGinRummyGame) PlayerLayoff(cardIndices []int) error {
	return m.Called(cardIndices).Error(0)
}
func (m *MockGinRummyGame) CpuPlay()    { m.Called() }
func (m *MockGinRummyGame) ScoreRound() { m.Called() }
func (m *MockGinRummyGame) GetConfig() domain.GinRummyConfig {
	return m.Called().Get(0).(domain.GinRummyConfig)
}
func (m *MockGinRummyGame) SetConfig(cfg domain.GinRummyConfig) { m.Called(cfg) }
func (m *MockGinRummyGame) GetGameEndFlag() bool                { return m.Called().Bool(0) }
func (m *MockGinRummyGame) GetPhase() domain.GinRummyPhase {
	return m.Called().Get(0).(domain.GinRummyPhase)
}
func (m *MockGinRummyGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockGinRummyGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockGinRummyGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockGinRummyGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockGinRummyGame) GetDrawPileCount() int { return m.Called().Int(0) }
func (m *MockGinRummyGame) GetWinnerIdx() int     { return m.Called().Int(0) }
func (m *MockGinRummyGame) GetPlayerCnt() int     { return m.Called().Int(0) }
func (m *MockGinRummyGame) GetPlayer(i int) *domain.GinRummyPlayer {
	return m.Called(i).Get(0).(*domain.GinRummyPlayer)
}
func (m *MockGinRummyGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockGinRummyGame) GetKnockerIdx() int { return m.Called().Int(0) }
func (m *MockGinRummyGame) LayoffTargets(card *domain.Card) []int {
	out, _ := m.Called(card).Get(0).([]int)
	return out
}

func (m *MockGinRummyGame) GetKnockerMelds() [][]*domain.Card {
	return m.Called().Get(0).([][]*domain.Card)
}
func (m *MockGinRummyGame) GetKnockerDeadwood() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockGinRummyGame) GetIsGin() bool { return m.Called().Bool(0) }
