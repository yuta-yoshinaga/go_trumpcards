//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockNiuNiuInteractor 闘牛 インタラクターモック
type MockNiuNiuInteractor struct {
	mock.Mock
}

func (_m *MockNiuNiuInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNiuNiuInteractor) Bet(amount int) string {
	ret := _m.Called(amount)
	return ret.Get(0).(string)
}

func (_m *MockNiuNiuInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockNiuNiuInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
