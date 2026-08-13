//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCincinnatiInteractor シンシナティインタラクターモック
type MockCincinnatiInteractor struct {
	mock.Mock
}

func (m *MockCincinnatiInteractor) Reset() string { return m.Called().String(0) }

func (m *MockCincinnatiInteractor) ResetWithConfig(cfg domain.CincinnatiConfig) string {
	return m.Called(cfg).String(0)
}

func (m *MockCincinnatiInteractor) Action(action, amount int) string {
	return m.Called(action, amount).String(0)
}

func (m *MockCincinnatiInteractor) NextHand() string { return m.Called().String(0) }

func (m *MockCincinnatiInteractor) GetConfig() domain.CincinnatiConfig {
	return m.Called().Get(0).(domain.CincinnatiConfig)
}

func (m *MockCincinnatiInteractor) Hint() string { return m.Called().String(0) }

func (m *MockCincinnatiInteractor) ActionLog() string { return m.Called().String(0) }

// Snapshot モック
func (m *MockCincinnatiInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}
