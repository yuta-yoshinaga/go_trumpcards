//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockPontoonInteractor ポンツーン インタラクターモック
type MockPontoonInteractor struct {
	mock.Mock
}

func (_m *MockPontoonInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPontoonInteractor) Bet(amount int) string {
	ret := _m.Called(amount)
	return ret.Get(0).(string)
}

func (_m *MockPontoonInteractor) Deal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPontoonInteractor) Stick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPontoonInteractor) Twist() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPontoonInteractor) Buy(extra int) string {
	ret := _m.Called(extra)
	return ret.Get(0).(string)
}

func (_m *MockPontoonInteractor) Split() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPontoonInteractor) BankerTwist() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPontoonInteractor) BankerStay() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPontoonInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockPontoonInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
