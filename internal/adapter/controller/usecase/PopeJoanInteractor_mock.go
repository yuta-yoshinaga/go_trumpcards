//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPopeJoanInteractor ポープ・ジョーン インタラクターモック
type MockPopeJoanInteractor struct {
	mock.Mock
}

func (_m *MockPopeJoanInteractor) Reset() string { return _m.Called().Get(0).(string) }

func (_m *MockPopeJoanInteractor) ResetWithConfig(cfg domain.PopeJoanConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

func (_m *MockPopeJoanInteractor) Play(handIdx int) string {
	return _m.Called(handIdx).Get(0).(string)
}

func (_m *MockPopeJoanInteractor) NextDeal() string { return _m.Called().Get(0).(string) }

func (_m *MockPopeJoanInteractor) GetConfig() domain.PopeJoanConfig {
	return _m.Called().Get(0).(domain.PopeJoanConfig)
}

func (_m *MockPopeJoanInteractor) Hint() string { return _m.Called().Get(0).(string) }

func (_m *MockPopeJoanInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

// Snapshot モック
func (_m *MockPopeJoanInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
