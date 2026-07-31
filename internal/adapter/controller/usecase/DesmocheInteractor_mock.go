//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDesmocheInteractor デスモチェ インタラクターモック
type MockDesmocheInteractor struct {
	mock.Mock
}

func (_m *MockDesmocheInteractor) Reset() string { return _m.Called().Get(0).(string) }

func (_m *MockDesmocheInteractor) ResetWithConfig(cfg domain.DesmocheConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

func (_m *MockDesmocheInteractor) DrawStock() string { return _m.Called().Get(0).(string) }

func (_m *MockDesmocheInteractor) DrawDiscard() string { return _m.Called().Get(0).(string) }

func (_m *MockDesmocheInteractor) Meld(handIdxs []int) string {
	return _m.Called(handIdxs).Get(0).(string)
}

func (_m *MockDesmocheInteractor) LayOff(handIdx, meldIdx int) string {
	return _m.Called(handIdx, meldIdx).Get(0).(string)
}

func (_m *MockDesmocheInteractor) Desmoche(fromMeldIdx, cardIdx, toMeldIdx int) string {
	return _m.Called(fromMeldIdx, cardIdx, toMeldIdx).Get(0).(string)
}

func (_m *MockDesmocheInteractor) Discard(handIdx int) string {
	return _m.Called(handIdx).Get(0).(string)
}

func (_m *MockDesmocheInteractor) NextRound() string { return _m.Called().Get(0).(string) }

func (_m *MockDesmocheInteractor) GetConfig() domain.DesmocheConfig {
	return _m.Called().Get(0).(domain.DesmocheConfig)
}

func (_m *MockDesmocheInteractor) Hint() string { return _m.Called().Get(0).(string) }

func (_m *MockDesmocheInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

// Snapshot モック
func (_m *MockDesmocheInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
