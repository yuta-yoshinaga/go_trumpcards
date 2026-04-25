//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockSpiteAndMaliceInteractor Spite & Malice インタラクターモック
type MockSpiteAndMaliceInteractor struct {
	mock.Mock
}

func (_m *MockSpiteAndMaliceInteractor) Reset() string {
	return _m.Called().Get(0).(string)
}

func (_m *MockSpiteAndMaliceInteractor) PlayFromHand(handIdx, foundationIdx int) string {
	return _m.Called(handIdx, foundationIdx).Get(0).(string)
}

func (_m *MockSpiteAndMaliceInteractor) PlayFromGoal(foundationIdx int) string {
	return _m.Called(foundationIdx).Get(0).(string)
}

func (_m *MockSpiteAndMaliceInteractor) PlayFromSide(sideIdx, foundationIdx int) string {
	return _m.Called(sideIdx, foundationIdx).Get(0).(string)
}

func (_m *MockSpiteAndMaliceInteractor) Discard(handIdx, sideIdx int) string {
	return _m.Called(handIdx, sideIdx).Get(0).(string)
}

func (_m *MockSpiteAndMaliceInteractor) CpuStep() string {
	return _m.Called().Get(0).(string)
}

func (_m *MockSpiteAndMaliceInteractor) Hint() string {
	return _m.Called().Get(0).(string)
}

func (_m *MockSpiteAndMaliceInteractor) ActionLog() string {
	return _m.Called().Get(0).(string)
}

func (_m *MockSpiteAndMaliceInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
