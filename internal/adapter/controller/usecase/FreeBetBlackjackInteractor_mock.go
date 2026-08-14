//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFreeBetBlackjackInteractor フリーベット・ブラックジャックインタラクターモック
type MockFreeBetBlackjackInteractor struct {
	mock.Mock
}

func (m *MockFreeBetBlackjackInteractor) Reset() string { return m.Called().String(0) }

func (m *MockFreeBetBlackjackInteractor) ResetWithConfig(cfg domain.FreeBetBlackjackConfig) string {
	return m.Called(cfg).String(0)
}

func (m *MockFreeBetBlackjackInteractor) PlaceBet(ante int) string {
	return m.Called(ante).String(0)
}

func (m *MockFreeBetBlackjackInteractor) Hit() string { return m.Called().String(0) }

func (m *MockFreeBetBlackjackInteractor) Stand() string { return m.Called().String(0) }

func (m *MockFreeBetBlackjackInteractor) FreeDouble() string { return m.Called().String(0) }

func (m *MockFreeBetBlackjackInteractor) FreeSplit() string { return m.Called().String(0) }

func (m *MockFreeBetBlackjackInteractor) NextRound() string { return m.Called().String(0) }

func (m *MockFreeBetBlackjackInteractor) GetConfig() domain.FreeBetBlackjackConfig {
	return m.Called().Get(0).(domain.FreeBetBlackjackConfig)
}

func (m *MockFreeBetBlackjackInteractor) Hint() string { return m.Called().String(0) }

func (m *MockFreeBetBlackjackInteractor) ActionLog() string { return m.Called().String(0) }

// Snapshot モック
func (m *MockFreeBetBlackjackInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}
