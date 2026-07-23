//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockClockSolitaireGame クロックソリティアゲームモック
type MockClockSolitaireGame struct {
	mock.Mock
}

func (_m *MockClockSolitaireGame) Reset() {
	_m.Called()
}

func (_m *MockClockSolitaireGame) Step() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockClockSolitaireGame) AutoPlay() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockClockSolitaireGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockClockSolitaireGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockClockSolitaireGame) GetPhase() domain.ClockSolitairePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.ClockSolitairePhase)
}

func (_m *MockClockSolitaireGame) GetPiles() [domain.ClockSolitairePileCount][]*domain.ClockSolitaireCard {
	ret := _m.Called()
	return ret.Get(0).([domain.ClockSolitairePileCount][]*domain.ClockSolitaireCard)
}

func (_m *MockClockSolitaireGame) GetFaceUpCount() [domain.ClockSolitairePileCount]int {
	ret := _m.Called()
	return ret.Get(0).([domain.ClockSolitairePileCount]int)
}

func (_m *MockClockSolitaireGame) GetStepCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

func (_m *MockClockSolitaireGame) GetCurrentCard() *domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.Card)
}

func (_m *MockClockSolitaireGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockClockSolitaireGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
