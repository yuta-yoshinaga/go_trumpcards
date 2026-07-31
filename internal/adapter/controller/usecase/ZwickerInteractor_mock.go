//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockZwickerInteractor ツヴィッカー インタラクターモック
type MockZwickerInteractor struct {
	mock.Mock
}

func (_m *MockZwickerInteractor) Reset() string { return _m.Called().Get(0).(string) }

func (_m *MockZwickerInteractor) ResetWithConfig(cfg domain.ZwickerConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

func (_m *MockZwickerInteractor) Take(handIdx, playedValue int, tableIdxs, buildIdxs []int) string {
	return _m.Called(handIdx, playedValue, tableIdxs, buildIdxs).Get(0).(string)
}

func (_m *MockZwickerInteractor) Build(handIdx int, tableIdxs []int, declaredValue int) string {
	return _m.Called(handIdx, tableIdxs, declaredValue).Get(0).(string)
}

func (_m *MockZwickerInteractor) Trail(handIdx int) string {
	return _m.Called(handIdx).Get(0).(string)
}

func (_m *MockZwickerInteractor) NextRound() string { return _m.Called().Get(0).(string) }

func (_m *MockZwickerInteractor) GetConfig() domain.ZwickerConfig {
	return _m.Called().Get(0).(domain.ZwickerConfig)
}

func (_m *MockZwickerInteractor) Hint() string { return _m.Called().Get(0).(string) }

func (_m *MockZwickerInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

// Snapshot モック
func (_m *MockZwickerInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
