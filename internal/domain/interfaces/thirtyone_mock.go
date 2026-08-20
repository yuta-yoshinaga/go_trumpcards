//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockThirtyOneGame モック
type MockThirtyOneGame struct {
	mock.Mock
}

func (m *MockThirtyOneGame) Reset()                       { m.Called() }
func (m *MockThirtyOneGame) NextRound()                   { m.Called() }
func (m *MockThirtyOneGame) PlayerDrawFromStock() error   { return m.Called().Error(0) }
func (m *MockThirtyOneGame) PlayerDrawFromDiscard() error { return m.Called().Error(0) }
func (m *MockThirtyOneGame) PlayerDiscard(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}
func (m *MockThirtyOneGame) PlayerKnock() error { return m.Called().Error(0) }
func (m *MockThirtyOneGame) CpuPlay()           { m.Called() }
func (m *MockThirtyOneGame) GetConfig() domain.ThirtyOneConfig {
	return m.Called().Get(0).(domain.ThirtyOneConfig)
}
func (m *MockThirtyOneGame) SetConfig(cfg domain.ThirtyOneConfig) { m.Called(cfg) }

// GetCpuKnockThreshold モック
func (m *MockThirtyOneGame) GetCpuKnockThreshold() int { return m.Called().Int(0) }
func (m *MockThirtyOneGame) GetGameEndFlag() bool      { return m.Called().Bool(0) }
func (m *MockThirtyOneGame) GetPhase() domain.ThirtyOnePhase {
	return m.Called().Get(0).(domain.ThirtyOnePhase)
}
func (m *MockThirtyOneGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockThirtyOneGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockThirtyOneGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockThirtyOneGame) GetDiscardTop() *domain.Card {
	return m.Called().Get(0).(*domain.Card)
}
func (m *MockThirtyOneGame) GetDrawPileCount() int { return m.Called().Int(0) }
func (m *MockThirtyOneGame) GetWinnerIdx() int     { return m.Called().Int(0) }
func (m *MockThirtyOneGame) GetPlayerCnt() int     { return m.Called().Int(0) }
func (m *MockThirtyOneGame) GetPlayer(i int) *domain.ThirtyOnePlayer {
	return m.Called(i).Get(0).(*domain.ThirtyOnePlayer)
}
func (m *MockThirtyOneGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockThirtyOneGame) GetKnockerIdx() int { return m.Called().Int(0) }

func (m *MockThirtyOneGame) GetHint() *domain.ThirtyOneHint {
	h, _ := m.Called().Get(0).(*domain.ThirtyOneHint)
	return h
}

func (m *MockThirtyOneGame) GetThirtyOneIdx() int   { return m.Called().Int(0) }
func (m *MockThirtyOneGame) GetRoundWinnerIdx() int { return m.Called().Int(0) }
func (m *MockThirtyOneGame) GetRoundLosers() []int {
	return m.Called().Get(0).([]int)
}
