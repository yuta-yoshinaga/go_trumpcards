//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBalootInteractor バルートインタラクターモック
type MockBalootInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockBalootInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockBalootInteractor) Reset() string { return _m.Called().Get(0).(string) }

// ResetWithConfig モック
func (_m *MockBalootInteractor) ResetWithConfig(cfg domain.BalootConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// DeclareSun モック
func (_m *MockBalootInteractor) DeclareSun() string { return _m.Called().Get(0).(string) }

// DeclareHokom モック
func (_m *MockBalootInteractor) DeclareHokom(suit int) string {
	return _m.Called(suit).Get(0).(string)
}

// PassDeclaration モック
func (_m *MockBalootInteractor) PassDeclaration() string { return _m.Called().Get(0).(string) }

// Play モック
func (_m *MockBalootInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextRound モック
func (_m *MockBalootInteractor) NextRound() string { return _m.Called().Get(0).(string) }

// GiveUp モック
func (_m *MockBalootInteractor) GiveUp() string { return _m.Called().Get(0).(string) }

// GetConfig モック
func (_m *MockBalootInteractor) GetConfig() domain.BalootConfig {
	return _m.Called().Get(0).(domain.BalootConfig)
}

// Hint モック
func (_m *MockBalootInteractor) Hint() string { return _m.Called().Get(0).(string) }

// ActionLog モック
func (_m *MockBalootInteractor) ActionLog() string { return _m.Called().Get(0).(string) }
