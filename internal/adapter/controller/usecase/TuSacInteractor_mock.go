//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTuSacInteractor 四色牌インタラクターモック
type MockTuSacInteractor struct {
	mock.Mock
}

func (m *MockTuSacInteractor) Reset() string { return m.Called().String(0) }

func (m *MockTuSacInteractor) ResetWithConfig(cfg domain.TuSacConfig) string {
	return m.Called(cfg).String(0)
}

func (m *MockTuSacInteractor) Draw(fromDiscard bool) string {
	return m.Called(fromDiscard).String(0)
}

func (m *MockTuSacInteractor) Meld(indexes []int) string { return m.Called(indexes).String(0) }

func (m *MockTuSacInteractor) Discard(index int) string { return m.Called(index).String(0) }

func (m *MockTuSacInteractor) NextRound() string { return m.Called().String(0) }

func (m *MockTuSacInteractor) GetConfig() domain.TuSacConfig {
	return m.Called().Get(0).(domain.TuSacConfig)
}

func (m *MockTuSacInteractor) Hint() string { return m.Called().String(0) }

func (m *MockTuSacInteractor) ActionLog() string { return m.Called().String(0) }

// Snapshot モック
func (m *MockTuSacInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}
