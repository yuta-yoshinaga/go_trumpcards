//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockLobaInteractor ロバ インタラクターモック
type MockLobaInteractor struct {
	mock.Mock
}

func (_m *MockLobaInteractor) Reset() string { return _m.Called().Get(0).(string) }

func (_m *MockLobaInteractor) ResetWithConfig(cfg domain.LobaConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

func (_m *MockLobaInteractor) DrawStock() string { return _m.Called().Get(0).(string) }

func (_m *MockLobaInteractor) DrawDiscard() string { return _m.Called().Get(0).(string) }

func (_m *MockLobaInteractor) Meld(handIdxs []int) string {
	return _m.Called(handIdxs).Get(0).(string)
}

func (_m *MockLobaInteractor) LayOff(handIdx, meldIdx int) string {
	return _m.Called(handIdx, meldIdx).Get(0).(string)
}

func (_m *MockLobaInteractor) Discard(handIdx int) string {
	return _m.Called(handIdx).Get(0).(string)
}

func (_m *MockLobaInteractor) NextRound() string { return _m.Called().Get(0).(string) }

func (_m *MockLobaInteractor) GetConfig() domain.LobaConfig {
	return _m.Called().Get(0).(domain.LobaConfig)
}

func (_m *MockLobaInteractor) Hint() string { return _m.Called().Get(0).(string) }

func (_m *MockLobaInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

// Snapshot モック
func (_m *MockLobaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
