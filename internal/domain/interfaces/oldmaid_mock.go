package interfaces

import (
	"github.com/stretchr/testify/mock"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockOldMaidGame ババ抜きゲームモック
type MockOldMaidGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockOldMaidGame) Reset() {
	_m.Called()
}

// GetGameEndFlag モック
func (_m *MockOldMaidGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockOldMaidGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerDraw モック
func (_m *MockOldMaidGame) PlayerDraw(cardIdx int) error {
	ret := _m.Called(cardIdx)
	return ret.Error(0)
}

// CpuDraw モック
func (_m *MockOldMaidGame) CpuDraw() error {
	ret := _m.Called()
	return ret.Error(0)
}

// GetPlayerCnt モック
func (_m *MockOldMaidGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockOldMaidGame) GetPlayer(i int) *domain.OldMaidPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.OldMaidPlayer); ok {
		return val
	}
	return nil
}

// GetHasDrawn モック
func (_m *MockOldMaidGame) GetHasDrawn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetLastDrawPlayerIdx モック
func (_m *MockOldMaidGame) GetLastDrawPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastDrawFromIdx モック
func (_m *MockOldMaidGame) GetLastDrawFromIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastDrawCard モック
func (_m *MockOldMaidGame) GetLastDrawCard() *domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.Card); ok {
		return val
	}
	return nil
}

// GetLastDiscardedPairs モック
func (_m *MockOldMaidGame) GetLastDiscardedPairs() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastDiscardedCards モック
func (_m *MockOldMaidGame) GetLastDiscardedCards() []*domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.Card); ok {
		return val
	}
	return nil
}

// GetCpuActions モック
func (_m *MockOldMaidGame) GetCpuActions() []*domain.OldMaidCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.OldMaidCpuAction); ok {
		return val
	}
	return nil
}

// GetHumanAction モック
func (_m *MockOldMaidGame) GetHumanAction() *domain.OldMaidCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.OldMaidCpuAction); ok {
		return val
	}
	return nil
}

// GetLoserIdx モック
func (_m *MockOldMaidGame) GetLoserIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentTurn モック
func (_m *MockOldMaidGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetNextDrawTargetIdx モック
func (_m *MockOldMaidGame) GetNextDrawTargetIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}
