//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockClockSolitaireInteractor クロックソリティアインタラクターモック
type MockClockSolitaireInteractor struct {
	mock.Mock
}

func (_m *MockClockSolitaireInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockClockSolitaireInteractor) Step() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockClockSolitaireInteractor) AutoPlay() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockClockSolitaireInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
