//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDurakInteractor ドゥラークインタラクターモック
type MockDurakInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockDurakInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Attack モック
func (_m *MockDurakInteractor) Attack(cardIdx int) string {
	ret := _m.Called(cardIdx)
	return ret.Get(0).(string)
}

// Defend モック
func (_m *MockDurakInteractor) Defend(attackIdx, handIdx int) string {
	ret := _m.Called(attackIdx, handIdx)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockDurakInteractor) Pass() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// TakeCards モック
func (_m *MockDurakInteractor) TakeCards() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Transfer モック
func (_m *MockDurakInteractor) Transfer(handIdx int) string {
	ret := _m.Called(handIdx)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockDurakInteractor) GetConfig() domain.DurakConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.DurakConfig)
}

// ResetWithConfig モック
func (_m *MockDurakInteractor) ResetWithConfig(config domain.DurakConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// Sort モック
func (_m *MockDurakInteractor) Sort(mode domain.DurakSortMode) string {
	ret := _m.Called(mode)
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockDurakInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Hint モック
func (_m *MockDurakInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockDurakInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
