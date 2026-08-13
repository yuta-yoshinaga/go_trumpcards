//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockIronCrossInteractor アイアンクロスインタラクターモック
type MockIronCrossInteractor struct {
	mock.Mock
}

func (m *MockIronCrossInteractor) Reset() string { return m.Called().String(0) }

func (m *MockIronCrossInteractor) ResetWithConfig(cfg domain.IronCrossConfig) string {
	return m.Called(cfg).String(0)
}

func (m *MockIronCrossInteractor) Action(action, amount int) string {
	return m.Called(action, amount).String(0)
}

func (m *MockIronCrossInteractor) ChooseLine(line int) string {
	return m.Called(line).String(0)
}

func (m *MockIronCrossInteractor) NextHand() string { return m.Called().String(0) }

func (m *MockIronCrossInteractor) GetConfig() domain.IronCrossConfig {
	return m.Called().Get(0).(domain.IronCrossConfig)
}

func (m *MockIronCrossInteractor) Hint() string { return m.Called().String(0) }

func (m *MockIronCrossInteractor) ActionLog() string { return m.Called().String(0) }

// Snapshot モック
func (m *MockIronCrossInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}
