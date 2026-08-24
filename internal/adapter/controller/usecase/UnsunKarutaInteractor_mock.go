//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockUnsunKarutaInteractor はうんすんカルタのインタラクターモック。
type MockUnsunKarutaInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockUnsunKarutaInteractor) Reset() string {
	return _m.Called().Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockUnsunKarutaInteractor) ResetWithConfig(cfg domain.UnsunKarutaConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// Play モック
func (_m *MockUnsunKarutaInteractor) Play(cardIndex int, declare bool) string {
	return _m.Called(cardIndex, declare).Get(0).(string)
}

// NextTrick モック
func (_m *MockUnsunKarutaInteractor) NextTrick() string {
	return _m.Called().Get(0).(string)
}

// NextRound モック
func (_m *MockUnsunKarutaInteractor) NextRound() string {
	return _m.Called().Get(0).(string)
}

// GetConfig モック
func (_m *MockUnsunKarutaInteractor) GetConfig() domain.UnsunKarutaConfig {
	return _m.Called().Get(0).(domain.UnsunKarutaConfig)
}

// Hint モック
func (_m *MockUnsunKarutaInteractor) Hint() string {
	return _m.Called().Get(0).(string)
}

// ActionLog モック
func (_m *MockUnsunKarutaInteractor) ActionLog() string {
	return _m.Called().Get(0).(string)
}

// Snapshot モック
func (_m *MockUnsunKarutaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
