//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPolignacGame ポリニャックゲームモック
type MockPolignacGame struct {
	mock.Mock
}

func (m *MockPolignacGame) Reset()                 { m.Called() }
func (m *MockPolignacGame) DeclareCapot() error    { return m.Called().Error(0) }
func (m *MockPolignacGame) PassDeclaration() error { return m.Called().Error(0) }
func (m *MockPolignacGame) CpuPlay()               { m.Called() }
func (m *MockPolignacGame) NextRound()             { m.Called() }
func (m *MockPolignacGame) GiveUp()                { m.Called() }

func (m *MockPolignacGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockPolignacGame) GetConfig() domain.PolignacConfig {
	return m.Called().Get(0).(domain.PolignacConfig)
}

func (m *MockPolignacGame) SetConfig(cfg domain.PolignacConfig) { m.Called(cfg) }

func (m *MockPolignacGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockPolignacGame) GetPhase() domain.PolignacPhase {
	return m.Called().Get(0).(domain.PolignacPhase)
}

func (m *MockPolignacGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockPolignacGame) IsDeclarePhase() bool     { return m.Called().Bool(0) }
func (m *MockPolignacGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockPolignacGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockPolignacGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockPolignacGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockPolignacGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockPolignacGame) GetCapotIdx() int         { return m.Called().Int(0) }
func (m *MockPolignacGame) GetCapotTricks() int      { return m.Called().Int(0) }
func (m *MockPolignacGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockPolignacGame) GetWinnerIdx() int        { return m.Called().Int(0) }

func (m *MockPolignacGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockPolignacGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockPolignacGame) GetPlayer(i int) *domain.PolignacPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.PolignacPlayer)
}

func (m *MockPolignacGame) GetHint() *domain.PolignacHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.PolignacHint)
}

func (m *MockPolignacGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}
