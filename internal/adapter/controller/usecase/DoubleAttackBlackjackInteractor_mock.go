//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDoubleAttackBlackjackInteractor 追加ベット・ブラックジャックインタラクターモック
type MockDoubleAttackBlackjackInteractor struct {
	mock.Mock
}

func (m *MockDoubleAttackBlackjackInteractor) Reset() string { return m.Called().String(0) }

func (m *MockDoubleAttackBlackjackInteractor) ResetWithConfig(cfg domain.DoubleAttackBlackjackConfig) string {
	return m.Called(cfg).String(0)
}

func (m *MockDoubleAttackBlackjackInteractor) PlaceBet(ante, bustIt int) string {
	return m.Called(ante, bustIt).String(0)
}

func (m *MockDoubleAttackBlackjackInteractor) Attack(amount int) string {
	return m.Called(amount).String(0)
}

func (m *MockDoubleAttackBlackjackInteractor) Hit() string { return m.Called().String(0) }

func (m *MockDoubleAttackBlackjackInteractor) Stand() string { return m.Called().String(0) }

func (m *MockDoubleAttackBlackjackInteractor) Double() string { return m.Called().String(0) }

func (m *MockDoubleAttackBlackjackInteractor) Split() string { return m.Called().String(0) }

func (m *MockDoubleAttackBlackjackInteractor) NextRound() string { return m.Called().String(0) }

func (m *MockDoubleAttackBlackjackInteractor) GetConfig() domain.DoubleAttackBlackjackConfig {
	return m.Called().Get(0).(domain.DoubleAttackBlackjackConfig)
}

func (m *MockDoubleAttackBlackjackInteractor) Hint() string { return m.Called().String(0) }

func (m *MockDoubleAttackBlackjackInteractor) ActionLog() string { return m.Called().String(0) }

// Snapshot モック
func (m *MockDoubleAttackBlackjackInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}
