//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKoiKoiInteractor はこいこい (Koi-Koi) のインタラクターモック。
type MockKoiKoiInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockKoiKoiInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockKoiKoiInteractor) ResetWithConfig(cfg domain.KoiKoiConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockKoiKoiInteractor) Play(handIdx, fieldIdx int) string {
	ret := _m.Called(handIdx, fieldIdx)
	return ret.Get(0).(string)
}

// Decide モック
func (_m *MockKoiKoiInteractor) Decide(koikoi bool) string {
	ret := _m.Called(koikoi)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockKoiKoiInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockKoiKoiInteractor) GetConfig() domain.KoiKoiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.KoiKoiConfig)
}

// Hint モック
func (_m *MockKoiKoiInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockKoiKoiInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockKoiKoiInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
