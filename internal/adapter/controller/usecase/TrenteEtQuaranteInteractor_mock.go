//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTrenteEtQuaranteInteractor はトラント・エ・カラント (Trente et Quarante) の
// インタラクターモック。
type MockTrenteEtQuaranteInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockTrenteEtQuaranteInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockTrenteEtQuaranteInteractor) ResetWithConfig(cfg domain.TrenteEtQuaranteConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bet モック
func (_m *MockTrenteEtQuaranteInteractor) Bet(bet domain.TrenteEtQuaranteBet, stake int) string {
	ret := _m.Called(bet, stake)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockTrenteEtQuaranteInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockTrenteEtQuaranteInteractor) GetConfig() domain.TrenteEtQuaranteConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TrenteEtQuaranteConfig)
}

// Hint モック
func (_m *MockTrenteEtQuaranteInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockTrenteEtQuaranteInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockTrenteEtQuaranteInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
