//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockSetteEMezzoInteractor セッテ・エ・メッツォ インタラクターモック
type MockSetteEMezzoInteractor struct {
	mock.Mock
}

func (_m *MockSetteEMezzoInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSetteEMezzoInteractor) Bet(amount int) string {
	ret := _m.Called(amount)
	return ret.Get(0).(string)
}

func (_m *MockSetteEMezzoInteractor) Deal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSetteEMezzoInteractor) Hit() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSetteEMezzoInteractor) Stand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSetteEMezzoInteractor) Matta(halves int) string {
	ret := _m.Called(halves)
	return ret.Get(0).(string)
}

func (_m *MockSetteEMezzoInteractor) BankerHit() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSetteEMezzoInteractor) BankerStand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSetteEMezzoInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSetteEMezzoInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
