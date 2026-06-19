//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockOpenFaceChineseInteractor オープンフェイス・チャイニーズポーカー (OFC) のインタラクターモック
type MockOpenFaceChineseInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockOpenFaceChineseInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockOpenFaceChineseInteractor) ResetWithConfig(cfg domain.OpenFaceChineseConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Place モック
func (_m *MockOpenFaceChineseInteractor) Place(row int) string {
	ret := _m.Called(row)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockOpenFaceChineseInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockOpenFaceChineseInteractor) GetConfig() domain.OpenFaceChineseConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.OpenFaceChineseConfig)
}

// Hint モック
func (_m *MockOpenFaceChineseInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockOpenFaceChineseInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockOpenFaceChineseInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
