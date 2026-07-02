//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTablanetInteractor はタブラネット (Tablanet) のインタラクターモック。
type MockTablanetInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockTablanetInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockTablanetInteractor) ResetWithConfig(cfg domain.TablanetConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockTablanetInteractor) Play(handIdx int, tableIdxs []int) string {
	ret := _m.Called(handIdx, tableIdxs)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockTablanetInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockTablanetInteractor) GetConfig() domain.TablanetConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TablanetConfig)
}

// Hint モック
func (_m *MockTablanetInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockTablanetInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockTablanetInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
