//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockTexasHoldemBonusInteractor テキサスホールデムボーナスポーカーインタラクターモック
type MockTexasHoldemBonusInteractor struct {
	mock.Mock
}

func (m *MockTexasHoldemBonusInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTexasHoldemBonusInteractor) Bet(ante, bonus int) string {
	args := m.Called(ante, bonus)
	return args.String(0)
}

func (m *MockTexasHoldemBonusInteractor) Play() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTexasHoldemBonusInteractor) Fold() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTexasHoldemBonusInteractor) Check() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTexasHoldemBonusInteractor) Raise() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTexasHoldemBonusInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockTexasHoldemBonusInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
