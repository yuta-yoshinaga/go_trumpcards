//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMushiInteractor 虫 インタラクターモック
type MockMushiInteractor struct {
	mock.Mock
}

func (_m *MockMushiInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMushiInteractor) ResetWithConfig(cfg domain.MushiConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

func (_m *MockMushiInteractor) Play(handIdx int) string {
	ret := _m.Called(handIdx)
	return ret.Get(0).(string)
}

func (_m *MockMushiInteractor) Select(fieldIdx int) string {
	ret := _m.Called(fieldIdx)
	return ret.Get(0).(string)
}

func (_m *MockMushiInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMushiInteractor) GetConfig() domain.MushiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.MushiConfig)
}

func (_m *MockMushiInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMushiInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockMushiInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
