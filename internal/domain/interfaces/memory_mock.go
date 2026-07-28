//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMemoryGame 神経衰弱ゲームモック
type MockMemoryGame struct {
	mock.Mock
}

func (_m *MockMemoryGame) Reset() {
	_m.Called()
}

func (_m *MockMemoryGame) PlayerFlip(pos int) error {
	ret := _m.Called(pos)
	return ret.Error(0)
}

func (_m *MockMemoryGame) CpuFlip() {
	_m.Called()
}

func (_m *MockMemoryGame) ResolveFlip() {
	_m.Called()
}

func (_m *MockMemoryGame) GetConfig() domain.MemoryConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.MemoryConfig)
}

func (_m *MockMemoryGame) SetConfig(cfg domain.MemoryConfig) {
	_m.Called(cfg)
}

func (_m *MockMemoryGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockMemoryGame) GetPhase() domain.MemoryPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.MemoryPhase)
}

func (_m *MockMemoryGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockMemoryGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMemoryGame) GetFirstFlipPos() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMemoryGame) GetSecondFlipPos() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMemoryGame) GetLastMatchResult() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockMemoryGame) GetTurnNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMemoryGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMemoryGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMemoryGame) GetPlayer(i int) *domain.MemoryPlayer {
	ret := _m.Called(i)
	return ret.Get(0).(*domain.MemoryPlayer)
}

func (_m *MockMemoryGame) GetBoard() []*domain.MemoryBoardCard {
	ret := _m.Called()
	return ret.Get(0).([]*domain.MemoryBoardCard)
}

func (_m *MockMemoryGame) GetBoardCard(pos int) *domain.MemoryBoardCard {
	ret := _m.Called(pos)
	if v := ret.Get(0); v == nil {
		return nil
	}
	return ret.Get(0).(*domain.MemoryBoardCard)
}

func (_m *MockMemoryGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v == nil {
		return nil
	}
	return ret.Get(0).([]*domain.ActionLogEntry)
}
