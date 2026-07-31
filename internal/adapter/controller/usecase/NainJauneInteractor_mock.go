//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockNainJauneInteractor ル・ナン・ジョーヌ インタラクターモック
type MockNainJauneInteractor struct {
	mock.Mock
}

func (_m *MockNainJauneInteractor) Reset() string { return _m.Called().Get(0).(string) }

func (_m *MockNainJauneInteractor) ResetWithConfig(cfg domain.NainJauneConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

func (_m *MockNainJauneInteractor) Play(handIdx int) string {
	return _m.Called(handIdx).Get(0).(string)
}

func (_m *MockNainJauneInteractor) NextDeal() string { return _m.Called().Get(0).(string) }

func (_m *MockNainJauneInteractor) GetConfig() domain.NainJauneConfig {
	return _m.Called().Get(0).(domain.NainJauneConfig)
}

func (_m *MockNainJauneInteractor) Hint() string { return _m.Called().Get(0).(string) }

func (_m *MockNainJauneInteractor) ActionLog() string { return _m.Called().Get(0).(string) }

// Snapshot モック
func (_m *MockNainJauneInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
