//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockQuinzeGame カーンズ ゲームモック
type MockQuinzeGame struct {
	mock.Mock
}

func (_m *MockQuinzeGame) Reset() { _m.Called() }

func (_m *MockQuinzeGame) PlaceBet(bet int) error {
	ret := _m.Called(bet)
	return ret.Error(0)
}

func (_m *MockQuinzeGame) StartAsBanker() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockQuinzeGame) Hit() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockQuinzeGame) Stand() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockQuinzeGame) BankerHit() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockQuinzeGame) BankerStand() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockQuinzeGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockQuinzeGame) GetChips() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockQuinzeGame) GetSeats() []*domain.QuinzeSeat {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.QuinzeSeat)
}

func (_m *MockQuinzeGame) GetBankerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockQuinzeGame) IsHumanBanker() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockQuinzeGame) GetBankerHand() *domain.QuinzeHand {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.QuinzeHand)
}

func (_m *MockQuinzeGame) GetActiveSeat() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockQuinzeGame) GetNextBanker() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockQuinzeGame) GetLastResult() string {
	ret := _m.Called()
	return ret.String(0)
}

func (_m *MockQuinzeGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockQuinzeGame) GetHandPoints(h *domain.QuinzeHand) int {
	ret := _m.Called(h)
	return ret.Int(0)
}

func (_m *MockQuinzeGame) FormatPoints(points int) string {
	ret := _m.Called(points)
	return ret.String(0)
}

func (_m *MockQuinzeGame) CanHit() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockQuinzeGame) CanStand() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockQuinzeGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}
