//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPageOneGame モック
type MockPageOneGame struct {
	mock.Mock
}

func (m *MockPageOneGame) Reset()                         { m.Called() }
func (m *MockPageOneGame) NextRound()                     { m.Called() }
func (m *MockPageOneGame) PlayerPlay(cardIndex int) error { return m.Called(cardIndex).Error(0) }
func (m *MockPageOneGame) PlayerDraw() error              { return m.Called().Error(0) }
func (m *MockPageOneGame) PlayerDeclare() error           { return m.Called().Error(0) }
func (m *MockPageOneGame) PlayerSkipDeclare() error       { return m.Called().Error(0) }
func (m *MockPageOneGame) CpuPlay()                       { m.Called() }
func (m *MockPageOneGame) CpuDeclare()                    { m.Called() }
func (m *MockPageOneGame) ScoreRound()                    { m.Called() }
func (m *MockPageOneGame) GetConfig() domain.PageOneConfig {
	return m.Called().Get(0).(domain.PageOneConfig)
}
func (m *MockPageOneGame) SetConfig(cfg domain.PageOneConfig) { m.Called(cfg) }
func (m *MockPageOneGame) GetGameEndFlag() bool               { return m.Called().Bool(0) }
func (m *MockPageOneGame) GetPhase() domain.PageOnePhase {
	return m.Called().Get(0).(domain.PageOnePhase)
}
func (m *MockPageOneGame) IsHumanTurn() bool           { return m.Called().Bool(0) }
func (m *MockPageOneGame) GetRoundNumber() int         { return m.Called().Int(0) }
func (m *MockPageOneGame) GetCurrentPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockPageOneGame) GetDiscardTop() *domain.Card { return m.Called().Get(0).(*domain.Card) }
func (m *MockPageOneGame) GetDrawPileCount() int       { return m.Called().Int(0) }
func (m *MockPageOneGame) GetWinnerIdx() int           { return m.Called().Int(0) }
func (m *MockPageOneGame) GetPlayerCnt() int           { return m.Called().Int(0) }
func (m *MockPageOneGame) GetPlayer(i int) *domain.PageOnePlayer {
	return m.Called(i).Get(0).(*domain.PageOnePlayer)
}
func (m *MockPageOneGame) IsValidPlay(card *domain.Card) bool {
	return m.Called(card).Bool(0)
}
func (m *MockPageOneGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
func (m *MockPageOneGame) GetRecentPenalties() []domain.PageOnePenalty {
	if v := m.Called().Get(0); v != nil {
		return v.([]domain.PageOnePenalty)
	}
	return nil
}
