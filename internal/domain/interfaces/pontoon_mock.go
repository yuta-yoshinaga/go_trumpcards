//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPontoonGame ポンツーン ゲームモック
type MockPontoonGame struct {
	mock.Mock
}

func (_m *MockPontoonGame) Reset() {
	_m.Called()
}

func (_m *MockPontoonGame) PlaceBet(bet int) error {
	ret := _m.Called(bet)
	return ret.Error(0)
}

func (_m *MockPontoonGame) StartAsBanker() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockPontoonGame) Stick() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockPontoonGame) Twist() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockPontoonGame) Buy(extra int) error {
	ret := _m.Called(extra)
	return ret.Error(0)
}

func (_m *MockPontoonGame) Split() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockPontoonGame) BankerTwist() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockPontoonGame) BankerStay() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockPontoonGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockPontoonGame) GetChips() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockPontoonGame) GetSeats() []*domain.PontoonSeat {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.PontoonSeat)
}

func (_m *MockPontoonGame) GetBankerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockPontoonGame) IsHumanBanker() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockPontoonGame) GetBankerHand() *domain.PontoonHand {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.PontoonHand)
}

func (_m *MockPontoonGame) GetActiveSeat() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockPontoonGame) GetActiveHand() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockPontoonGame) GetNextBanker() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockPontoonGame) GetLastResult() string {
	ret := _m.Called()
	return ret.String(0)
}

func (_m *MockPontoonGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockPontoonGame) GetHandTotal(cards []*domain.Card) int {
	ret := _m.Called(cards)
	return ret.Int(0)
}

func (_m *MockPontoonGame) GetHandRank(cards []*domain.Card) domain.PontoonRank {
	ret := _m.Called(cards)
	return ret.Get(0).(domain.PontoonRank)
}

func (_m *MockPontoonGame) CanStick() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockPontoonGame) CanTwist() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockPontoonGame) CanBuy() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockPontoonGame) CanSplit() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockPontoonGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}
