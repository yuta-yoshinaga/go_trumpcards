//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRistikontraGame は Pişti のゲームモック。
type MockRistikontraGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockRistikontraGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockRistikontraGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockRistikontraGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockRistikontraGame) CpuPlay() {
	_m.Called()
}

// SetConfig モック
func (_m *MockRistikontraGame) SetConfig(config domain.RistikontraConfig) {
	_m.Called(config)
}

// GetConfig モック
func (_m *MockRistikontraGame) GetConfig() domain.RistikontraConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.RistikontraConfig)
}

// GetGameEndFlag モック
func (_m *MockRistikontraGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockRistikontraGame) GetPhase() domain.RistikontraPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.RistikontraPhase)
}

// IsHumanTurn モック
func (_m *MockRistikontraGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetCurrentTurn モック
func (_m *MockRistikontraGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPile モック
func (_m *MockRistikontraGame) GetPile() []*domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

// GetPileTop モック
func (_m *MockRistikontraGame) GetPileTop() *domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

// GetLastCaptureIdx モック
func (_m *MockRistikontraGame) GetLastCaptureIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockRistikontraGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockRistikontraGame) GetPlayer(i int) *domain.RistikontraPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.RistikontraPlayer)
	}
	return nil
}

// GetRemainingDeck モック
func (_m *MockRistikontraGame) GetRemainingDeck() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinners モック
func (_m *MockRistikontraGame) GetWinners() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetFinalScores モック
func (_m *MockRistikontraGame) GetFinalScores() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetActionLog モック
func (_m *MockRistikontraGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

// GetProvisionalScores モック
func (m *MockRistikontraGame) GetProvisionalScores() []int {
	ret := m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetProvisionalLeader モック
func (m *MockRistikontraGame) GetProvisionalLeader() int { return m.Called().Int(0) }
