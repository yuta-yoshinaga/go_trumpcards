//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBadugiInteractor is the testify/mock implementation of BadugiInteractorIF.
type MockBadugiInteractor struct {
	mock.Mock
}

// Reset mock.
func (m *MockBadugiInteractor) Reset() string {
	ret := m.Called()
	return ret.Get(0).(string)
}

// GetConfig mock.
func (m *MockBadugiInteractor) GetConfig() domain.BadugiConfig {
	ret := m.Called()
	return ret.Get(0).(domain.BadugiConfig)
}

// ResetWithConfig mock.
func (m *MockBadugiInteractor) ResetWithConfig(cfg domain.BadugiConfig, profileData []byte) string {
	ret := m.Called(cfg, profileData)
	return ret.Get(0).(string)
}

// Action mock.
func (m *MockBadugiInteractor) Action(action, amount, humanPlayMs int) string {
	ret := m.Called(action, amount, humanPlayMs)
	return ret.Get(0).(string)
}

// Exchange mock.
func (m *MockBadugiInteractor) Exchange(indices []int, humanPlayMs int) string {
	ret := m.Called(indices, humanPlayMs)
	return ret.Get(0).(string)
}

// Stand mock.
func (m *MockBadugiInteractor) Stand(humanPlayMs int) string {
	ret := m.Called(humanPlayMs)
	return ret.Get(0).(string)
}

// ActionLog mock.
func (m *MockBadugiInteractor) Hint() string {
	ret := m.Called()
	return ret.Get(0).(string)
}

func (m *MockBadugiInteractor) ActionLog() string {
	ret := m.Called()
	return ret.Get(0).(string)
}

// Snapshot mock.
func (m *MockBadugiInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
