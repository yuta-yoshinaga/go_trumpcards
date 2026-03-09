package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDoubtInteractor ダウトインタラクターモック
type MockDoubtInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockDoubtInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockDoubtInteractor) ResetWithConfig(cfg domain.DoubtConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockDoubtInteractor) Play(cardIndices []int, claimedValue int) string {
	ret := _m.Called(cardIndices, claimedValue)
	return ret.Get(0).(string)
}

// ResolveDoubt モック
func (_m *MockDoubtInteractor) ResolveDoubt(doubterIndices []int) string {
	ret := _m.Called(doubterIndices)
	return ret.Get(0).(string)
}

// SkipDoubt モック
func (_m *MockDoubtInteractor) SkipDoubt() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetCpuDoubters モック
func (_m *MockDoubtInteractor) GetCpuDoubters() []int {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]int); ok {
		return val
	}
	return nil
}

// GetConfig モック
func (_m *MockDoubtInteractor) GetConfig() domain.DoubtConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.DoubtConfig)
}

// ResetProfile モック
func (_m *MockDoubtInteractor) ResetProfile() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockDoubtInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
