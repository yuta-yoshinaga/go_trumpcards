//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRummy500Game モック
type MockRummy500Game struct {
	mock.Mock
}

func (m *MockRummy500Game) Reset()                     { m.Called() }
func (m *MockRummy500Game) NextRound()                 { m.Called() }
func (m *MockRummy500Game) PlayerDrawFromStock() error { return m.Called().Error(0) }
func (m *MockRummy500Game) PlayerDrawFromDiscard(idx int) error {
	return m.Called(idx).Error(0)
}
func (m *MockRummy500Game) PlayerMeld(cardIndices []int) error {
	return m.Called(cardIndices).Error(0)
}
func (m *MockRummy500Game) PlayerLayoff(meldOwner, meldIdx, cardIndex int) error {
	return m.Called(meldOwner, meldIdx, cardIndex).Error(0)
}
func (m *MockRummy500Game) PlayerDiscard(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}
func (m *MockRummy500Game) CpuPlay()    { m.Called() }
func (m *MockRummy500Game) ScoreRound() { m.Called() }
func (m *MockRummy500Game) GetConfig() domain.Rummy500Config {
	return m.Called().Get(0).(domain.Rummy500Config)
}
func (m *MockRummy500Game) SetConfig(cfg domain.Rummy500Config) { m.Called(cfg) }
func (m *MockRummy500Game) GetGameEndFlag() bool                { return m.Called().Bool(0) }
func (m *MockRummy500Game) GetPhase() domain.Rummy500Phase {
	return m.Called().Get(0).(domain.Rummy500Phase)
}
func (m *MockRummy500Game) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockRummy500Game) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockRummy500Game) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockRummy500Game) GetDiscardPile() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockRummy500Game) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockRummy500Game) GetDrawPileCount() int { return m.Called().Int(0) }
func (m *MockRummy500Game) GetWinnerIdx() int     { return m.Called().Int(0) }
func (m *MockRummy500Game) GetPlayerCnt() int     { return m.Called().Int(0) }
func (m *MockRummy500Game) GetPlayer(i int) *domain.Rummy500Player {
	return m.Called(i).Get(0).(*domain.Rummy500Player)
}
func (m *MockRummy500Game) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockRummy500Game) GetRoundEnderIdx() int { return m.Called().Int(0) }
