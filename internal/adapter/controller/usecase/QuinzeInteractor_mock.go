//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockQuinzeInteractor カーンズ インタラクターモック
type MockQuinzeInteractor struct {
	mock.Mock
}

func (_m *MockQuinzeInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockQuinzeInteractor) Bet(amount int) string {
	ret := _m.Called(amount)
	return ret.Get(0).(string)
}

func (_m *MockQuinzeInteractor) Deal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockQuinzeInteractor) Hit() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockQuinzeInteractor) Stand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockQuinzeInteractor) BankerHit() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockQuinzeInteractor) BankerStand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockQuinzeInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockQuinzeInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
