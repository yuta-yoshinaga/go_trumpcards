//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCribbageGame クリベッジゲームモック
type MockCribbageGame struct {
	mock.Mock
}

func (m *MockCribbageGame) Reset()                            { m.Called() }
func (m *MockCribbageGame) NextRound()                        { m.Called() }
func (m *MockCribbageGame) PlayerDiscard(indices []int) error { return m.Called(indices).Error(0) }
func (m *MockCribbageGame) PlayerCut() error                  { return m.Called().Error(0) }
func (m *MockCribbageGame) PlayerPeg(cardIndex int) error     { return m.Called(cardIndex).Error(0) }
func (m *MockCribbageGame) PlayerGo() error                   { return m.Called().Error(0) }
func (m *MockCribbageGame) ShowNext() error                   { return m.Called().Error(0) }
func (m *MockCribbageGame) CpuPlay()                          { m.Called() }
func (m *MockCribbageGame) GetConfig() domain.CribbageConfig {
	return m.Called().Get(0).(domain.CribbageConfig)
}
func (m *MockCribbageGame) SetConfig(cfg domain.CribbageConfig) { m.Called(cfg) }
func (m *MockCribbageGame) GetGameEndFlag() bool                { return m.Called().Bool(0) }
func (m *MockCribbageGame) GetHint() *domain.CribbageHint {
	ret := m.Called().Get(0)
	if ret == nil {
		return nil
	}
	return ret.(*domain.CribbageHint)
}
func (m *MockCribbageGame) GetPhase() domain.CribbagePhase {
	return m.Called().Get(0).(domain.CribbagePhase)
}
func (m *MockCribbageGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockCribbageGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockCribbageGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockCribbageGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockCribbageGame) GetWinnerIdx() int        { return m.Called().Int(0) }
func (m *MockCribbageGame) GetPlayer(i int) *domain.CribbagePlayer {
	ret := m.Called(i)
	if p, ok := ret.Get(0).(*domain.CribbagePlayer); ok {
		return p
	}
	return nil
}
func (m *MockCribbageGame) GetActionLog() []*domain.ActionLogEntry {
	ret := m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
func (m *MockCribbageGame) GetCrib() []*domain.Card {
	ret := m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}
func (m *MockCribbageGame) GetStarter() *domain.Card {
	ret := m.Called()
	if v, ok := ret.Get(0).(*domain.Card); ok {
		return v
	}
	return nil
}
func (m *MockCribbageGame) GetPegCount() int { return m.Called().Int(0) }
func (m *MockCribbageGame) GetPegPlayedCards() []*domain.Card {
	ret := m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}
func (m *MockCribbageGame) GetShowPhaseStep() int { return m.Called().Int(0) }
func (m *MockCribbageGame) GetHandScoreDetails() [3]*domain.CribbageScoreDetail {
	return m.Called().Get(0).([3]*domain.CribbageScoreDetail)
}
func (m *MockCribbageGame) GetOriginalHand(playerIdx int) []*domain.Card {
	ret := m.Called(playerIdx)
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}
func (m *MockCribbageGame) GetPlayerPeggedCards(playerIdx int) []*domain.Card {
	ret := m.Called(playerIdx)
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}
