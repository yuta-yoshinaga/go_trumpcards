//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCuckooGame モック
type MockCuckooGame struct {
	mock.Mock
}

func (m *MockCuckooGame) Reset()                  { m.Called() }
func (m *MockCuckooGame) NextRound()              { m.Called() }
func (m *MockCuckooGame) PlayerKeep() error       { return m.Called().Error(0) }
func (m *MockCuckooGame) PlayerSwap() error       { return m.Called().Error(0) }
func (m *MockCuckooGame) PlayerRefuse() error     { return m.Called().Error(0) }
func (m *MockCuckooGame) PlayerAcceptSwap() error { return m.Called().Error(0) }
func (m *MockCuckooGame) CpuPlay()                { m.Called() }
func (m *MockCuckooGame) GetConfig() domain.CuckooConfig {
	return m.Called().Get(0).(domain.CuckooConfig)
}
func (m *MockCuckooGame) SetConfig(cfg domain.CuckooConfig) { m.Called(cfg) }
func (m *MockCuckooGame) GetGameEndFlag() bool              { return m.Called().Bool(0) }
func (m *MockCuckooGame) GetPhase() domain.CuckooPhase {
	return m.Called().Get(0).(domain.CuckooPhase)
}
func (m *MockCuckooGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockCuckooGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockCuckooGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockCuckooGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockCuckooGame) GetStockCount() int       { return m.Called().Int(0) }
func (m *MockCuckooGame) GetWinnerIdx() int        { return m.Called().Int(0) }
func (m *MockCuckooGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockCuckooGame) GetPlayer(i int) *domain.CuckooPlayer {
	return m.Called(i).Get(0).(*domain.CuckooPlayer)
}
func (m *MockCuckooGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockCuckooGame) GetPendingSwapFrom() int   { return m.Called().Int(0) }
func (m *MockCuckooGame) GetPendingSwapTo() int     { return m.Called().Int(0) }
func (m *MockCuckooGame) IsKingRevealed(i int) bool { return m.Called(i).Bool(0) }
func (m *MockCuckooGame) GetRoundLowest() int       { return m.Called().Int(0) }
func (m *MockCuckooGame) GetRoundLosers() []int {
	return m.Called().Get(0).([]int)
}
