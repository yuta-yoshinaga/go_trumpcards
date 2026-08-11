//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTarabishInteractor タラビッシュインタラクターモック
type MockTarabishInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockTarabishInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockTarabishInteractor) Reset() string { return _m.Called().Get(0).(string) }

// ResetWithConfig モック
func (_m *MockTarabishInteractor) ResetWithConfig(cfg domain.TarabishConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// TakeTrump モック
func (_m *MockTarabishInteractor) TakeTrump() string { return _m.Called().Get(0).(string) }

// PassTrump モック
func (_m *MockTarabishInteractor) PassTrump() string { return _m.Called().Get(0).(string) }

// Play モック
func (_m *MockTarabishInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextRound モック
func (_m *MockTarabishInteractor) NextRound() string { return _m.Called().Get(0).(string) }

// GiveUp モック
func (_m *MockTarabishInteractor) GiveUp() string { return _m.Called().Get(0).(string) }

// GetConfig モック
func (_m *MockTarabishInteractor) GetConfig() domain.TarabishConfig {
	return _m.Called().Get(0).(domain.TarabishConfig)
}

// Hint モック
func (_m *MockTarabishInteractor) Hint() string { return _m.Called().Get(0).(string) }

// ActionLog モック
func (_m *MockTarabishInteractor) ActionLog() string { return _m.Called().Get(0).(string) }
