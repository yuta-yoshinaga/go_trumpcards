//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMonteBankInteractor モンテバンクインタラクターモック
type MockMonteBankInteractor struct {
	mock.Mock
}

func (m *MockMonteBankInteractor) Reset() string { return m.Called().String(0) }

func (m *MockMonteBankInteractor) ResetWithConfig(cfg domain.MonteBankConfig) string {
	return m.Called(cfg).String(0)
}

func (m *MockMonteBankInteractor) PlaceBet(idx, bet int) string {
	return m.Called(idx, bet).String(0)
}

func (m *MockMonteBankInteractor) NextRound() string { return m.Called().String(0) }

func (m *MockMonteBankInteractor) GetConfig() domain.MonteBankConfig {
	return m.Called().Get(0).(domain.MonteBankConfig)
}

func (m *MockMonteBankInteractor) Hint() string { return m.Called().String(0) }

func (m *MockMonteBankInteractor) ActionLog() string { return m.Called().String(0) }

// Snapshot モック
func (m *MockMonteBankInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}
