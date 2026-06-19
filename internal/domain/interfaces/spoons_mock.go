//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSpoonsGame はスプーンのゲームモック。
type MockSpoonsGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockSpoonsGame) Reset() { _m.Called() }

// ResetWithConfig モック
func (_m *MockSpoonsGame) ResetWithConfig(cfg domain.SpoonsConfig) { _m.Called(cfg) }

// NextRound モック
func (_m *MockSpoonsGame) NextRound() { _m.Called() }

// PlayerPass モック
func (_m *MockSpoonsGame) PlayerPass(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// PlayerGrabSpoon モック
func (_m *MockSpoonsGame) PlayerGrabSpoon() error {
	ret := _m.Called()
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockSpoonsGame) CpuPlay() { _m.Called() }

// GetConfig モック
func (_m *MockSpoonsGame) GetConfig() domain.SpoonsConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SpoonsConfig)
}

// SetConfig モック
func (_m *MockSpoonsGame) SetConfig(cfg domain.SpoonsConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockSpoonsGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockSpoonsGame) GetPhase() domain.SpoonsPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SpoonsPhase)
}

// IsHumanTurn モック
func (_m *MockSpoonsGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPlayerCnt モック
func (_m *MockSpoonsGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockSpoonsGame) GetPlayer(i int) *domain.SpoonsPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.SpoonsPlayer)
	}
	return nil
}

// GetWinnerIdx モック
func (_m *MockSpoonsGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetSpoonsRemaining モック
func (_m *MockSpoonsGame) GetSpoonsRemaining() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockSpoonsGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetFeederIdx モック
func (_m *MockSpoonsGame) GetFeederIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDrawPileSize モック
func (_m *MockSpoonsGame) GetDrawPileSize() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPassedCard モック
func (_m *MockSpoonsGame) GetPassedCard() *domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

// IsGrabWindowOpen モック
func (_m *MockSpoonsGame) IsGrabWindowOpen() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetFirstGrabberIdx モック
func (_m *MockSpoonsGame) GetFirstGrabberIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundLoserIdx モック
func (_m *MockSpoonsGame) GetRoundLoserIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundNumber モック
func (_m *MockSpoonsGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetActionLog モック
func (_m *MockSpoonsGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
