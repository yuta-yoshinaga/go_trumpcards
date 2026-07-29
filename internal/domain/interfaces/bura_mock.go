//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBuraGame ブラ ゲームモック
type MockBuraGame struct {
	mock.Mock
}

func (_m *MockBuraGame) Reset() { _m.Called() }

func (_m *MockBuraGame) PlayCards(idx int, indices []int) error {
	ret := _m.Called(idx, indices)
	return ret.Error(0)
}

func (_m *MockBuraGame) Claim(idx int) error {
	ret := _m.Called(idx)
	return ret.Error(0)
}

func (_m *MockBuraGame) DeclareCombination(idx int) error {
	ret := _m.Called(idx)
	return ret.Error(0)
}

func (_m *MockBuraGame) BuraCpuDecide(idx int) domain.BuraCpuAction {
	ret := _m.Called(idx)
	return ret.Get(0).(domain.BuraCpuAction)
}

func (_m *MockBuraGame) GetConfig() domain.BuraConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BuraConfig)
}

func (_m *MockBuraGame) SetConfig(cfg domain.BuraConfig) { _m.Called(cfg) }

func (_m *MockBuraGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockBuraGame) GetPhase() domain.BuraPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.BuraPhase)
}

func (_m *MockBuraGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBuraGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBuraGame) GetCurrentLead() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockBuraGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBuraGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBuraGame) GetTrumpCard() *domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockBuraGame) GetStock() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

func (_m *MockBuraGame) GetPlayers() []*domain.BuraPlayer {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.BuraPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockBuraGame) GetPlayer(i int) *domain.BuraPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.BuraPlayer); ok {
		return v
	}
	return nil
}

func (_m *MockBuraGame) GetPlayerPoints(i int) int {
	ret := _m.Called(i)
	return ret.Int(0)
}

func (_m *MockBuraGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBuraGame) IsDraw() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockBuraGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
