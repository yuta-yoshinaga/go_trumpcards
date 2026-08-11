//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPasurGame パスールゲームモック
type MockPasurGame struct {
	mock.Mock
}

func (m *MockPasurGame) Reset()   { m.Called() }
func (m *MockPasurGame) CpuPlay() { m.Called() }
func (m *MockPasurGame) GiveUp()  { m.Called() }

func (m *MockPasurGame) PlayerPlay(cardIndex int, tableIndices []int) error {
	return m.Called(cardIndex, tableIndices).Error(0)
}

func (m *MockPasurGame) GetConfig() domain.PasurConfig {
	return m.Called().Get(0).(domain.PasurConfig)
}

func (m *MockPasurGame) SetConfig(cfg domain.PasurConfig) { m.Called(cfg) }

func (m *MockPasurGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockPasurGame) GetPhase() domain.PasurPhase {
	return m.Called().Get(0).(domain.PasurPhase)
}

func (m *MockPasurGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockPasurGame) GetDeckRemaining() int    { return m.Called().Int(0) }
func (m *MockPasurGame) GetPacksDealt() int       { return m.Called().Int(0) }
func (m *MockPasurGame) GetLastCaptureIdx() int   { return m.Called().Int(0) }
func (m *MockPasurGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockPasurGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockPasurGame) GetScore(i int) int       { return m.Called(i).Int(0) }

func (m *MockPasurGame) GetTableCards() []*domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

func (m *MockPasurGame) GetCaptureOptions(playerIdx, cardIndex int) [][]int {
	args := m.Called(playerIdx, cardIndex)
	if v := args.Get(0); v != nil {
		return v.([][]int)
	}
	return nil
}

func (m *MockPasurGame) GetWinners() []int {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockPasurGame) GetPlayer(i int) *domain.PasurPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.PasurPlayer)
	}
	return nil
}

func (m *MockPasurGame) GetHint() *domain.PasurHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.PasurHint)
	}
	return nil
}

func (m *MockPasurGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
