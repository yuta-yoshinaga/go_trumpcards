//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTehonbikiInteractor 手本引きインタラクターモック
type MockTehonbikiInteractor struct {
	mock.Mock
}

func (m *MockTehonbikiInteractor) Reset() string { return m.Called().String(0) }

func (m *MockTehonbikiInteractor) ResetWithConfig(cfg domain.TehonbikiConfig) string {
	return m.Called(cfg).String(0)
}

func (m *MockTehonbikiInteractor) PlaceBet(idx, bet int) string {
	return m.Called(idx, bet).String(0)
}

func (m *MockTehonbikiInteractor) NextRound() string { return m.Called().String(0) }

func (m *MockTehonbikiInteractor) GetConfig() domain.TehonbikiConfig {
	return m.Called().Get(0).(domain.TehonbikiConfig)
}

func (m *MockTehonbikiInteractor) Hint() string { return m.Called().String(0) }

func (m *MockTehonbikiInteractor) ActionLog() string { return m.Called().String(0) }

// Snapshot モック
func (m *MockTehonbikiInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}
