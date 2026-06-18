//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKalookiGame カルーキゲームのモック
type MockKalookiGame struct {
	mock.Mock
}

func (m *MockKalookiGame) Reset()                       { m.Called() }
func (m *MockKalookiGame) NextRound()                   { m.Called() }
func (m *MockKalookiGame) PlayerDrawFromStock() error   { return m.Called().Error(0) }
func (m *MockKalookiGame) PlayerDrawFromDiscard() error { return m.Called().Error(0) }
func (m *MockKalookiGame) PlayerMeld(groups [][]int) error {
	return m.Called(groups).Error(0)
}
func (m *MockKalookiGame) PlayerLayoff(t, mi, ci int) error {
	return m.Called(t, mi, ci).Error(0)
}
func (m *MockKalookiGame) PlayerDiscard(i int) error { return m.Called(i).Error(0) }
func (m *MockKalookiGame) CpuPlay()                  { m.Called() }
func (m *MockKalookiGame) GetConfig() domain.KalookiConfig {
	return m.Called().Get(0).(domain.KalookiConfig)
}
func (m *MockKalookiGame) SetConfig(cfg domain.KalookiConfig) { m.Called(cfg) }
func (m *MockKalookiGame) GetGameEndFlag() bool               { return m.Called().Bool(0) }
func (m *MockKalookiGame) GetPhase() domain.KalookiPhase {
	return m.Called().Get(0).(domain.KalookiPhase)
}
func (m *MockKalookiGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockKalookiGame) GetOpeningThreshold() int { return m.Called().Int(0) }
func (m *MockKalookiGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockKalookiGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockKalookiGame) GetDiscardPile() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}
func (m *MockKalookiGame) GetDrawPileCount() int { return m.Called().Int(0) }
func (m *MockKalookiGame) GetWinnerIdx() int     { return m.Called().Int(0) }
func (m *MockKalookiGame) GetPlayerCnt() int     { return m.Called().Int(0) }
func (m *MockKalookiGame) GetPlayer(i int) *domain.KalookiPlayer {
	return m.Called(i).Get(0).(*domain.KalookiPlayer)
}
func (m *MockKalookiGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockKalookiGame) GetRoundWinnerIdx() int { return m.Called().Int(0) }
