//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCrazyEightsGame モック
type MockCrazyEightsGame struct {
	mock.Mock
}

func (m *MockCrazyEightsGame) Reset()                          { m.Called() }
func (m *MockCrazyEightsGame) NextRound()                      { m.Called() }
func (m *MockCrazyEightsGame) PlayerPlay(cardIndex int) error  { return m.Called(cardIndex).Error(0) }
func (m *MockCrazyEightsGame) PlayerChooseSuit(suit int) error { return m.Called(suit).Error(0) }
func (m *MockCrazyEightsGame) PlayerDraw() error               { return m.Called().Error(0) }
func (m *MockCrazyEightsGame) CpuPlay()                        { m.Called() }
func (m *MockCrazyEightsGame) CpuChooseSuit()                  { m.Called() }
func (m *MockCrazyEightsGame) ScoreRound()                     { m.Called() }
func (m *MockCrazyEightsGame) GetConfig() domain.CrazyEightsConfig {
	return m.Called().Get(0).(domain.CrazyEightsConfig)
}
func (m *MockCrazyEightsGame) SetConfig(cfg domain.CrazyEightsConfig) { m.Called(cfg) }
func (m *MockCrazyEightsGame) GetGameEndFlag() bool                   { return m.Called().Bool(0) }
func (m *MockCrazyEightsGame) GetPhase() domain.CrazyEightsPhase {
	return m.Called().Get(0).(domain.CrazyEightsPhase)
}
func (m *MockCrazyEightsGame) IsHumanTurn() bool           { return m.Called().Bool(0) }
func (m *MockCrazyEightsGame) GetRoundNumber() int         { return m.Called().Int(0) }
func (m *MockCrazyEightsGame) GetCurrentPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockCrazyEightsGame) GetDiscardTop() *domain.Card { return m.Called().Get(0).(*domain.Card) }
func (m *MockCrazyEightsGame) GetDrawPileCount() int       { return m.Called().Int(0) }
func (m *MockCrazyEightsGame) GetChosenSuit() int          { return m.Called().Int(0) }
func (m *MockCrazyEightsGame) GetWinnerIdx() int           { return m.Called().Int(0) }
func (m *MockCrazyEightsGame) GetPlayerCnt() int           { return m.Called().Int(0) }
func (m *MockCrazyEightsGame) GetPlayer(i int) *domain.CrazyEightsPlayer {
	return m.Called(i).Get(0).(*domain.CrazyEightsPlayer)
}

// GetHint はサーバー計算の推奨手を返すモック。
func (m *MockCrazyEightsGame) GetHint() *domain.CrazyEightsHint {
	out, _ := m.Called().Get(0).(*domain.CrazyEightsHint)
	return out
}

func (m *MockCrazyEightsGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
