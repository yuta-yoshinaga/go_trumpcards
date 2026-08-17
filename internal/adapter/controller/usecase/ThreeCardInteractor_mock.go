//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockThreeCardInteractor スリーカードポーカーインタラクターモック
type MockThreeCardInteractor struct {
	mock.Mock
}

func (m *MockThreeCardInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockThreeCardInteractor) Bet(ante, pairPlus int) string {
	args := m.Called(ante, pairPlus)
	return args.String(0)
}

// Rebet モック
func (m *MockThreeCardInteractor) Rebet() string { return m.Called().String(0) }

func (m *MockThreeCardInteractor) Play() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockThreeCardInteractor) Fold() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockThreeCardInteractor) Hint() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockThreeCardInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockThreeCardInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
