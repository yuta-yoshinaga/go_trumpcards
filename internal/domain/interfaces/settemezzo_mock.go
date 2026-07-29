//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSetteEMezzoGame セッテ・エ・メッツォ ゲームモック
type MockSetteEMezzoGame struct {
	mock.Mock
}

func (_m *MockSetteEMezzoGame) Reset() { _m.Called() }

func (_m *MockSetteEMezzoGame) PlaceBet(bet int) error {
	ret := _m.Called(bet)
	return ret.Error(0)
}

func (_m *MockSetteEMezzoGame) StartAsBanker() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSetteEMezzoGame) Hit() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSetteEMezzoGame) Stand() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSetteEMezzoGame) SetMattaValue(halves int) error {
	ret := _m.Called(halves)
	return ret.Error(0)
}

func (_m *MockSetteEMezzoGame) BankerHit() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSetteEMezzoGame) BankerStand() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSetteEMezzoGame) GetPhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSetteEMezzoGame) GetChips() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSetteEMezzoGame) GetSeats() []*domain.SetteEMezzoSeat {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.SetteEMezzoSeat)
}

func (_m *MockSetteEMezzoGame) GetBankerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSetteEMezzoGame) IsHumanBanker() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSetteEMezzoGame) GetBankerHand() *domain.SetteEMezzoHand {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.SetteEMezzoHand)
}

func (_m *MockSetteEMezzoGame) GetActiveSeat() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSetteEMezzoGame) GetNextBanker() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSetteEMezzoGame) GetLastResult() string {
	ret := _m.Called()
	return ret.String(0)
}

func (_m *MockSetteEMezzoGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSetteEMezzoGame) GetHandHalves(h *domain.SetteEMezzoHand) int {
	ret := _m.Called(h)
	return ret.Int(0)
}

func (_m *MockSetteEMezzoGame) FormatHalves(halves int) string {
	ret := _m.Called(halves)
	return ret.String(0)
}

func (_m *MockSetteEMezzoGame) CanHit() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSetteEMezzoGame) CanStand() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSetteEMezzoGame) CanSetMatta() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSetteEMezzoGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}
