//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPishtiGame は Pişti のゲームモック。
type MockPishtiGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockPishtiGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockPishtiGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockPishtiGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockPishtiGame) CpuPlay() {
	_m.Called()
}

// SetConfig モック
func (_m *MockPishtiGame) SetConfig(config domain.PishtiConfig) {
	_m.Called(config)
}

// GetConfig モック
func (_m *MockPishtiGame) GetConfig() domain.PishtiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.PishtiConfig)
}

// GetGameEndFlag モック
func (_m *MockPishtiGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockPishtiGame) GetPhase() domain.PishtiPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.PishtiPhase)
}

// IsHumanTurn モック
func (_m *MockPishtiGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetCurrentTurn モック
func (_m *MockPishtiGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPile モック
func (_m *MockPishtiGame) GetPile() []*domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

// GetPileTop モック
func (_m *MockPishtiGame) GetPileTop() *domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

// GetLastCaptureIdx モック
func (_m *MockPishtiGame) GetLastCaptureIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockPishtiGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockPishtiGame) GetPlayer(i int) *domain.PishtiPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.PishtiPlayer)
	}
	return nil
}

// GetRemainingDeck モック
func (_m *MockPishtiGame) GetRemainingDeck() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinners モック
func (_m *MockPishtiGame) GetWinners() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetFinalScores モック
func (_m *MockPishtiGame) GetFinalScores() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetActionLog モック
func (_m *MockPishtiGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

// GetProvisionalScores モック
func (m *MockPishtiGame) GetProvisionalScores() []int {
	ret := m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetProvisionalLeader モック
func (m *MockPishtiGame) GetProvisionalLeader() int { return m.Called().Int(0) }
