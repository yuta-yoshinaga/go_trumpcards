//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockThreeCardRummyInteractor スリーカード・ラミーインタラクターモック
type MockThreeCardRummyInteractor struct {
	mock.Mock
}

func (m *MockThreeCardRummyInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockThreeCardRummyInteractor) Bet(ante, lowBonus int) string {
	args := m.Called(ante, lowBonus)
	return args.String(0)
}

// Rebet モック
func (m *MockThreeCardRummyInteractor) Rebet() string { return m.Called().String(0) }

func (m *MockThreeCardRummyInteractor) Play() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockThreeCardRummyInteractor) Fold() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockThreeCardRummyInteractor) Hint() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockThreeCardRummyInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockThreeCardRummyInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
