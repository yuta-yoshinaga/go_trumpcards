//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKingoInteractor キンゴインタラクターモック
type MockKingoInteractor struct {
	mock.Mock
}

func (m *MockKingoInteractor) Reset() string { return m.Called().String(0) }

func (m *MockKingoInteractor) ResetWithConfig(cfg domain.KingoConfig) string {
	return m.Called(cfg).String(0)
}

func (m *MockKingoInteractor) Bet(amount int) string { return m.Called(amount).String(0) }

func (m *MockKingoInteractor) Deal() string { return m.Called().String(0) }

func (m *MockKingoInteractor) NextRound() string { return m.Called().String(0) }

func (m *MockKingoInteractor) GetConfig() domain.KingoConfig {
	return m.Called().Get(0).(domain.KingoConfig)
}

func (m *MockKingoInteractor) Hint() string { return m.Called().String(0) }

func (m *MockKingoInteractor) ActionLog() string { return m.Called().String(0) }

// Snapshot モック
func (m *MockKingoInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}
