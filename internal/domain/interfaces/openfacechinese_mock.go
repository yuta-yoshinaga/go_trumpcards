//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockOpenFaceChineseGame オープンフェイス・チャイニーズポーカー (OFC) のゲームモック
type MockOpenFaceChineseGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockOpenFaceChineseGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockOpenFaceChineseGame) NextRound() {
	_m.Called()
}

// PlayerPlace モック
func (_m *MockOpenFaceChineseGame) PlayerPlace(row int) error {
	ret := _m.Called(row)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockOpenFaceChineseGame) CpuPlay() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockOpenFaceChineseGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockOpenFaceChineseGame) GetConfig() domain.OpenFaceChineseConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.OpenFaceChineseConfig)
}

// SetConfig モック
func (_m *MockOpenFaceChineseGame) SetConfig(cfg domain.OpenFaceChineseConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockOpenFaceChineseGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockOpenFaceChineseGame) GetPhase() domain.OpenFaceChinesePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.OpenFaceChinesePhase)
}

// IsHumanTurn モック
func (_m *MockOpenFaceChineseGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockOpenFaceChineseGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockOpenFaceChineseGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockOpenFaceChineseGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinnerIdx モック
func (_m *MockOpenFaceChineseGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockOpenFaceChineseGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockOpenFaceChineseGame) GetPlayer(i int) *domain.OpenFaceChinesePlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.OpenFaceChinesePlayer)
	}
	return nil
}

// GetCurrentCard モック
func (_m *MockOpenFaceChineseGame) GetCurrentCard() *domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

// GetHint モック
func (_m *MockOpenFaceChineseGame) GetHint() *domain.OpenFaceChineseHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.OpenFaceChineseHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockOpenFaceChineseGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
