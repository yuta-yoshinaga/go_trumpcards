//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPrsiGame モック
type MockPrsiGame struct {
	mock.Mock
}

func (m *MockPrsiGame) Reset()                         { m.Called() }
func (m *MockPrsiGame) PlayerPlay(cardIndex int) error { return m.Called(cardIndex).Error(0) }
func (m *MockPrsiGame) PlayerDraw() error              { return m.Called().Error(0) }
func (m *MockPrsiGame) CpuPlay()                       { m.Called() }
func (m *MockPrsiGame) GetConfig() domain.PrsiConfig {
	return m.Called().Get(0).(domain.PrsiConfig)
}
func (m *MockPrsiGame) SetConfig(cfg domain.PrsiConfig) { m.Called(cfg) }
func (m *MockPrsiGame) GetGameEndFlag() bool            { return m.Called().Bool(0) }
func (m *MockPrsiGame) GetPhase() domain.PrsiPhase {
	return m.Called().Get(0).(domain.PrsiPhase)
}
func (m *MockPrsiGame) IsHumanTurn() bool           { return m.Called().Bool(0) }
func (m *MockPrsiGame) GetCurrentPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockPrsiGame) GetDiscardTop() *domain.Card { return m.Called().Get(0).(*domain.Card) }
func (m *MockPrsiGame) GetDrawPileCount() int       { return m.Called().Int(0) }
func (m *MockPrsiGame) GetPenaltyDrawCount() int    { return m.Called().Int(0) }
func (m *MockPrsiGame) GetPendingSkips() int        { return m.Called().Int(0) }
func (m *MockPrsiGame) GetWinnerIdx() int           { return m.Called().Int(0) }
func (m *MockPrsiGame) GetPlayerCnt() int           { return m.Called().Int(0) }
func (m *MockPrsiGame) GetPlayer(i int) *domain.PrsiPlayer {
	return m.Called(i).Get(0).(*domain.PrsiPlayer)
}
func (m *MockPrsiGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
