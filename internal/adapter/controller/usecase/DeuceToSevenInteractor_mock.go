//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDeuceToSevenInteractor is the testify/mock implementation of
// DeuceToSevenInteractorIF.
type MockDeuceToSevenInteractor struct {
	mock.Mock
}

// Reset mock.
func (m *MockDeuceToSevenInteractor) Reset() string {
	ret := m.Called()
	return ret.Get(0).(string)
}

// GetConfig mock.
func (m *MockDeuceToSevenInteractor) GetConfig() domain.DeuceToSevenConfig {
	ret := m.Called()
	return ret.Get(0).(domain.DeuceToSevenConfig)
}

// ResetWithConfig mock.
func (m *MockDeuceToSevenInteractor) ResetWithConfig(cfg domain.DeuceToSevenConfig, profileData []byte) string {
	ret := m.Called(cfg, profileData)
	return ret.Get(0).(string)
}

// Action mock.
func (m *MockDeuceToSevenInteractor) Action(action, amount, humanPlayMs int) string {
	ret := m.Called(action, amount, humanPlayMs)
	return ret.Get(0).(string)
}

// Exchange mock.
func (m *MockDeuceToSevenInteractor) Exchange(indices []int) string {
	ret := m.Called(indices)
	return ret.Get(0).(string)
}

// Stand mock.
func (m *MockDeuceToSevenInteractor) Stand() string {
	ret := m.Called()
	return ret.Get(0).(string)
}

func (m *MockDeuceToSevenInteractor) Hint() string {
	ret := m.Called()
	return ret.Get(0).(string)
}

// ActionLog mock.
func (m *MockDeuceToSevenInteractor) ActionLog() string {
	ret := m.Called()
	return ret.Get(0).(string)
}

// Snapshot mock.
func (m *MockDeuceToSevenInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
