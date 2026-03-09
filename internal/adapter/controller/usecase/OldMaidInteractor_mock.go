package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockOldMaidInteractor ババ抜きインタラクターモック
type MockOldMaidInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockOldMaidInteractor) Reset(config domain.OldMaidConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockOldMaidInteractor) GetConfig() domain.OldMaidConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.OldMaidConfig)
}

// Draw モック
func (_m *MockOldMaidInteractor) Draw(cardIdx int) string {
	ret := _m.Called(cardIdx)
	return ret.Get(0).(string)
}

// Shuffle モック
func (_m *MockOldMaidInteractor) Shuffle() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Reorder モック
func (_m *MockOldMaidInteractor) Reorder(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// ResetProfile モック
func (_m *MockOldMaidInteractor) ResetProfile() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
