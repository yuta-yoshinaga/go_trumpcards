//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockNiuNiuGame 闘牛 ゲームモック
type MockNiuNiuGame struct {
	mock.Mock
}

func (_m *MockNiuNiuGame) Reset() { _m.Called() }

func (_m *MockNiuNiuGame) PlaceBet(bet int) error {
	ret := _m.Called(bet)
	return ret.Error(0)
}

func (_m *MockNiuNiuGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockNiuNiuGame) GetChips() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockNiuNiuGame) GetSeats() []*domain.NiuNiuSeat {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.NiuNiuSeat)
}

func (_m *MockNiuNiuGame) GetBankerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockNiuNiuGame) GetBankerHand() *domain.NiuNiuHand {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.NiuNiuHand)
}

func (_m *MockNiuNiuGame) GetLastResult() string {
	ret := _m.Called()
	return ret.String(0)
}

func (_m *MockNiuNiuGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockNiuNiuGame) GetRankLabel(rank domain.NiuNiuRank) string {
	ret := _m.Called(rank)
	return ret.String(0)
}

func (_m *MockNiuNiuGame) GetMultiplier(rank domain.NiuNiuRank) int {
	ret := _m.Called(rank)
	return ret.Int(0)
}

func (_m *MockNiuNiuGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}
