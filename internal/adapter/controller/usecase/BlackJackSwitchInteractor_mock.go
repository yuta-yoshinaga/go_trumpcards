//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockBlackJackSwitchInteractor ブラックジャック・スイッチインタラクターモック
type MockBlackJackSwitchInteractor struct {
	mock.Mock
}

// Reset モック
func (m *MockBlackJackSwitchInteractor) Reset() string {
	return m.Called().String(0)
}

// Bet モック
func (m *MockBlackJackSwitchInteractor) Bet(amount int) string {
	return m.Called(amount).String(0)
}

// Switch モック
func (m *MockBlackJackSwitchInteractor) Switch() string { return m.Called().String(0) }

// Keep モック
func (m *MockBlackJackSwitchInteractor) Keep() string { return m.Called().String(0) }

// Hit モック
func (m *MockBlackJackSwitchInteractor) Hit() string { return m.Called().String(0) }

// Stand モック
func (m *MockBlackJackSwitchInteractor) Stand() string { return m.Called().String(0) }

// DoubleDown モック
func (m *MockBlackJackSwitchInteractor) DoubleDown() string { return m.Called().String(0) }

// ActionLog モック
func (m *MockBlackJackSwitchInteractor) ActionLog() string {
	return m.Called().String(0)
}

// Snapshot モック
func (m *MockBlackJackSwitchInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
