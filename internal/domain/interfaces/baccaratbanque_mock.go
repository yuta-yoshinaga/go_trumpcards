//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBaccaratBanqueGame はバカラ・バンクのゲームモック。
type MockBaccaratBanqueGame struct {
	mock.Mock
}

// GetActionLog モック
func (_m *MockBaccaratBanqueGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v, _ := ret.Get(0).([]*domain.ActionLogEntry)
	return v
}

// Reset モック
func (_m *MockBaccaratBanqueGame) Reset() { _m.Called() }

// NextCoup モック
func (_m *MockBaccaratBanqueGame) NextCoup() { _m.Called() }

// BankerDraw モック
func (_m *MockBaccaratBanqueGame) BankerDraw(draw bool) error {
	ret := _m.Called(draw)
	err, _ := ret.Get(0).(error)
	return err
}

// Retire モック
func (_m *MockBaccaratBanqueGame) Retire() error {
	ret := _m.Called()
	err, _ := ret.Get(0).(error)
	return err
}

// GetConfig モック
func (_m *MockBaccaratBanqueGame) GetConfig() domain.BaccaratBanqueConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BaccaratBanqueConfig)
}

// SetConfig モック
func (_m *MockBaccaratBanqueGame) SetConfig(cfg domain.BaccaratBanqueConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockBaccaratBanqueGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockBaccaratBanqueGame) GetPhase() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// IsHumanTurn モック
func (_m *MockBaccaratBanqueGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetCoupNumber モック
func (_m *MockBaccaratBanqueGame) GetCoupNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetBankHeld モック
func (_m *MockBaccaratBanqueGame) GetBankHeld() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetShoeRemaining モック
func (_m *MockBaccaratBanqueGame) GetShoeRemaining() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// IsRetired モック
func (_m *MockBaccaratBanqueGame) IsRetired() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPlayerCnt モック
func (_m *MockBaccaratBanqueGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockBaccaratBanqueGame) GetPlayer(i int) *domain.BaccaratBanquePlayer {
	ret := _m.Called(i)
	v, _ := ret.Get(0).(*domain.BaccaratBanquePlayer)
	return v
}

// GetLastResult モック
func (_m *MockBaccaratBanqueGame) GetLastResult() *domain.BaccaratBanqueCoupResult {
	ret := _m.Called()
	v, _ := ret.Get(0).(*domain.BaccaratBanqueCoupResult)
	return v
}

// GetWinnerIdx モック
func (_m *MockBaccaratBanqueGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetHint モック
func (_m *MockBaccaratBanqueGame) GetHint() *domain.BaccaratBanqueHint {
	ret := _m.Called()
	v, _ := ret.Get(0).(*domain.BaccaratBanqueHint)
	return v
}
