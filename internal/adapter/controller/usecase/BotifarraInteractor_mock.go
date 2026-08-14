//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBotifarraInteractor ボティファラインタラクターモック
type MockBotifarraInteractor struct {
	mock.Mock
}

func (m *MockBotifarraInteractor) Reset() string { return m.Called().String(0) }

func (m *MockBotifarraInteractor) ResetWithConfig(cfg domain.BotifarraConfig) string {
	return m.Called(cfg).String(0)
}

func (m *MockBotifarraInteractor) Declare(suit int) string { return m.Called(suit).String(0) }

func (m *MockBotifarraInteractor) Delegate() string { return m.Called().String(0) }

func (m *MockBotifarraInteractor) Double() string { return m.Called().String(0) }

func (m *MockBotifarraInteractor) PassDouble() string { return m.Called().String(0) }

func (m *MockBotifarraInteractor) PlayCard(cardIndex int) string {
	return m.Called(cardIndex).String(0)
}

func (m *MockBotifarraInteractor) NextRound() string { return m.Called().String(0) }

func (m *MockBotifarraInteractor) GiveUp() string { return m.Called().String(0) }

func (m *MockBotifarraInteractor) GetConfig() domain.BotifarraConfig {
	return m.Called().Get(0).(domain.BotifarraConfig)
}

func (m *MockBotifarraInteractor) Hint() string { return m.Called().String(0) }

func (m *MockBotifarraInteractor) ActionLog() string { return m.Called().String(0) }

// Snapshot モック
func (m *MockBotifarraInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}
