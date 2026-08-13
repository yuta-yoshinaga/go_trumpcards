//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBanLuckInteractor バンラックインタラクターモック
type MockBanLuckInteractor struct {
	mock.Mock
}

func (m *MockBanLuckInteractor) Reset() string { return m.Called().String(0) }

func (m *MockBanLuckInteractor) ResetWithConfig(cfg domain.BanLuckConfig) string {
	return m.Called(cfg).String(0)
}

func (m *MockBanLuckInteractor) PlaceBet(bet int) string { return m.Called(bet).String(0) }

func (m *MockBanLuckInteractor) Hit() string { return m.Called().String(0) }

func (m *MockBanLuckInteractor) Stand() string { return m.Called().String(0) }

func (m *MockBanLuckInteractor) NextRound() string { return m.Called().String(0) }

func (m *MockBanLuckInteractor) GetConfig() domain.BanLuckConfig {
	return m.Called().Get(0).(domain.BanLuckConfig)
}

func (m *MockBanLuckInteractor) Hint() string { return m.Called().String(0) }

func (m *MockBanLuckInteractor) ActionLog() string { return m.Called().String(0) }

func (m *MockBanLuckInteractor) Snapshot() ([]byte, error) {
	args := m.Called()
	if b, ok := args.Get(0).([]byte); ok {
		return b, args.Error(1)
	}
	return nil, args.Error(1)
}
