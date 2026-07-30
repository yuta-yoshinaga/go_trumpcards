//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockToepenInteractor トゥーペン インタラクターモック
type MockToepenInteractor struct {
	mock.Mock
}

func (_m *MockToepenInteractor) Reset() string { return _m.Called().Get(0).(string) }

func (_m *MockToepenInteractor) ResetWithConfig(cfg domain.ToepenConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

func (_m *MockToepenInteractor) Play(handIdx int) string {
	return _m.Called(handIdx).Get(0).(string)
}

func (_m *MockToepenInteractor) Toep() string { return _m.Called().Get(0).(string) }

func (_m *MockToepenInteractor) Respond(stay bool) string {
	return _m.Called(stay).Get(0).(string)
}

func (_m *MockToepenInteractor) Redeal() string { return _m.Called().Get(0).(string) }

func (_m *MockToepenInteractor) NextHand() string { return _m.Called().Get(0).(string) }

func (_m *MockToepenInteractor) GetConfig() domain.ToepenConfig {
	return _m.Called().Get(0).(domain.ToepenConfig)
}

func (_m *MockToepenInteractor) Hint() string { return _m.Called().Get(0).(string) }

func (_m *MockToepenInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

// Snapshot モック
func (_m *MockToepenInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
