//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPolignacInteractor ポリニャックインタラクターモック
type MockPolignacInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockPolignacInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockPolignacInteractor) Reset() string { return _m.Called().Get(0).(string) }

// ResetWithConfig モック
func (_m *MockPolignacInteractor) ResetWithConfig(cfg domain.PolignacConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// DeclareCapot モック
func (_m *MockPolignacInteractor) DeclareCapot() string { return _m.Called().Get(0).(string) }

// Pass モック
func (_m *MockPolignacInteractor) Pass() string { return _m.Called().Get(0).(string) }

// Play モック
func (_m *MockPolignacInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextRound モック
func (_m *MockPolignacInteractor) NextRound() string { return _m.Called().Get(0).(string) }

// GiveUp モック
func (_m *MockPolignacInteractor) GiveUp() string { return _m.Called().Get(0).(string) }

// GetConfig モック
func (_m *MockPolignacInteractor) GetConfig() domain.PolignacConfig {
	return _m.Called().Get(0).(domain.PolignacConfig)
}

// Hint モック
func (_m *MockPolignacInteractor) Hint() string { return _m.Called().Get(0).(string) }

// ActionLog モック
func (_m *MockPolignacInteractor) ActionLog() string { return _m.Called().Get(0).(string) }
