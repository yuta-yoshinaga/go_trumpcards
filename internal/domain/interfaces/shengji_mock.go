//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockShengJiGame は ShengJiGame のモック。
type MockShengJiGame struct {
	mock.Mock
}

func (m *MockShengJiGame) Reset()          { m.Called() }
func (m *MockShengJiGame) NextHand() error { return m.Called().Error(0) }
func (m *MockShengJiGame) Declare(seat, suit int) error {
	return m.Called(seat, suit).Error(0)
}
func (m *MockShengJiGame) BuryKitty(seat int, idxs []int) error {
	return m.Called(seat, idxs).Error(0)
}
func (m *MockShengJiGame) Play(seat int, idxs []int) error {
	return m.Called(seat, idxs).Error(0)
}
func (m *MockShengJiGame) CpuPlay() { m.Called() }
func (m *MockShengJiGame) ShengJiDeclareStrength(seat, suit int) int {
	return m.Called(seat, suit).Int(0)
}
func (m *MockShengJiGame) GetConfig() domain.ShengJiConfig {
	return m.Called().Get(0).(domain.ShengJiConfig)
}
func (m *MockShengJiGame) SetConfig(cfg domain.ShengJiConfig) { m.Called(cfg) }
func (m *MockShengJiGame) GetGameEndFlag() bool               { return m.Called().Bool(0) }
func (m *MockShengJiGame) GetPhase() domain.ShengJiPhase {
	return m.Called().Get(0).(domain.ShengJiPhase)
}
func (m *MockShengJiGame) IsHumanTurn() bool         { return m.Called().Bool(0) }
func (m *MockShengJiGame) GetCurrentPlayerIdx() int  { return m.Called().Int(0) }
func (m *MockShengJiGame) GetLevel() int             { return m.Called().Int(0) }
func (m *MockShengJiGame) GetTeamLevel(team int) int { return m.Called(team).Int(0) }
func (m *MockShengJiGame) GetDeclarerTeam() int      { return m.Called().Int(0) }
func (m *MockShengJiGame) GetTrumpSuit() int         { return m.Called().Int(0) }
func (m *MockShengJiGame) GetDeclaration() *domain.ShengJiDeclaration {
	v := m.Called().Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.ShengJiDeclaration)
}
func (m *MockShengJiGame) GetKittySize() int { return m.Called().Int(0) }
func (m *MockShengJiGame) GetKitty() []*domain.Card {
	v := m.Called().Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}
func (m *MockShengJiGame) GetTrick() [][]*domain.Card {
	v := m.Called().Get(0)
	if v == nil {
		return nil
	}
	return v.([][]*domain.Card)
}
func (m *MockShengJiGame) GetTrickLeader() int { return m.Called().Int(0) }
func (m *MockShengJiGame) GetLeadCombo() *domain.ShengJiCombo {
	v := m.Called().Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.ShengJiCombo)
}
func (m *MockShengJiGame) GetTeamPoints(team int) int { return m.Called(team).Int(0) }
func (m *MockShengJiGame) GetTrickCount() int         { return m.Called().Int(0) }
func (m *MockShengJiGame) GetLastTrickWinner() int    { return m.Called().Int(0) }
func (m *MockShengJiGame) GetLastResult() *domain.ShengJiHandResult {
	v := m.Called().Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.ShengJiHandResult)
}
func (m *MockShengJiGame) GetHandNumber() int { return m.Called().Int(0) }
func (m *MockShengJiGame) GetWinnerTeam() int { return m.Called().Int(0) }
func (m *MockShengJiGame) GetPlayers() []*domain.ShengJiPlayer {
	return m.Called().Get(0).([]*domain.ShengJiPlayer)
}
func (m *MockShengJiGame) GetPlayer(idx int) *domain.ShengJiPlayer {
	v := m.Called(idx).Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.ShengJiPlayer)
}
func (m *MockShengJiGame) GetActionLog() []*domain.ActionLogEntry {
	v := m.Called().Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}
