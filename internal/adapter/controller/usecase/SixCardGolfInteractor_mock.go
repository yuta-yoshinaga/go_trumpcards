//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSixCardGolfInteractor モック
type MockSixCardGolfInteractor struct {
	mock.Mock
}

func (_m *MockSixCardGolfInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockSixCardGolfInteractor) ResetWithConfig(cfg domain.SixCardGolfConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockSixCardGolfInteractor) FlipInitial(pos int) string {
	return _m.Called(pos).String(0)
}

func (_m *MockSixCardGolfInteractor) DrawStock() string {
	return _m.Called().String(0)
}

func (_m *MockSixCardGolfInteractor) DrawDiscard() string {
	return _m.Called().String(0)
}

func (_m *MockSixCardGolfInteractor) SwapCard(pos int) string {
	return _m.Called(pos).String(0)
}

func (_m *MockSixCardGolfInteractor) DiscardDrawn() string {
	return _m.Called().String(0)
}

func (_m *MockSixCardGolfInteractor) FlipCard(pos int) string {
	return _m.Called(pos).String(0)
}

func (_m *MockSixCardGolfInteractor) SkipFlip() string {
	return _m.Called().String(0)
}

func (_m *MockSixCardGolfInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockSixCardGolfInteractor) GetConfig() domain.SixCardGolfConfig {
	return _m.Called().Get(0).(domain.SixCardGolfConfig)
}

func (_m *MockSixCardGolfInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Hint モック
func (_m *MockSixCardGolfInteractor) Hint() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockSixCardGolfInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
