//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGoofspielInteractor ゴフスピールインタラクターモック
type MockGoofspielInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockGoofspielInteractor) Snapshot() ([]byte, error) {
	args := _m.Called()
	if v := args.Get(0); v != nil {
		return v.([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

// Reset モック
func (_m *MockGoofspielInteractor) Reset() string { return _m.Called().String(0) }

// ResetWithConfig モック
func (_m *MockGoofspielInteractor) ResetWithConfig(cfg domain.GoofspielConfig) string {
	return _m.Called(cfg).String(0)
}

// Bid モック
func (_m *MockGoofspielInteractor) Bid(cardIndex int) string { return _m.Called(cardIndex).String(0) }

// NextRound モック
func (_m *MockGoofspielInteractor) NextRound() string { return _m.Called().String(0) }

// GiveUp モック
func (_m *MockGoofspielInteractor) GiveUp() string { return _m.Called().String(0) }

// GetConfig モック
func (_m *MockGoofspielInteractor) GetConfig() domain.GoofspielConfig {
	return _m.Called().Get(0).(domain.GoofspielConfig)
}

// Hint モック
func (_m *MockGoofspielInteractor) Hint() string { return _m.Called().String(0) }

// ActionLog モック
func (_m *MockGoofspielInteractor) ActionLog() string { return _m.Called().String(0) }
