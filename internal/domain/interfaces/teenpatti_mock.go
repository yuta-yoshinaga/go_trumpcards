//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTeenPattiGame ティーン・パティのゲームモック
type MockTeenPattiGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockTeenPattiGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockTeenPattiGame) NextRound() {
	_m.Called()
}

// PlayerSee モック
func (_m *MockTeenPattiGame) PlayerSee() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerBet モック
func (_m *MockTeenPattiGame) PlayerBet() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerRaise モック
func (_m *MockTeenPattiGame) PlayerRaise(newStake int) error {
	ret := _m.Called(newStake)
	return ret.Error(0)
}

// PlayerFold モック
func (_m *MockTeenPattiGame) PlayerFold() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerShow モック
func (_m *MockTeenPattiGame) PlayerShow() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerRequestSideShow モック
func (_m *MockTeenPattiGame) PlayerRequestSideShow() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerRespondSideShow モック
func (_m *MockTeenPattiGame) PlayerRespondSideShow(accept bool) error {
	ret := _m.Called(accept)
	return ret.Error(0)
}

// CpuAct モック
func (_m *MockTeenPattiGame) CpuAct() {
	_m.Called()
}

// GetConfig モック
func (_m *MockTeenPattiGame) GetConfig() domain.TeenPattiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TeenPattiConfig)
}

// SetConfig モック
func (_m *MockTeenPattiGame) SetConfig(cfg domain.TeenPattiConfig) {
	_m.Called(cfg)
}

// GetPhase モック
func (_m *MockTeenPattiGame) GetPhase() domain.TeenPattiPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.TeenPattiPhase)
}

// GetRoundNumber モック
func (_m *MockTeenPattiGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockTeenPattiGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockTeenPattiGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPot モック
func (_m *MockTeenPattiGame) GetPot() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRaiseRange はレイズ可能な最小/最大額を返すモック。
func (_m *MockTeenPattiGame) GetRaiseRange(playerIdx int) (int, int, bool) {
	ret := _m.Called(playerIdx)
	return ret.Int(0), ret.Int(1), ret.Bool(2)
}

// GetStake モック
func (_m *MockTeenPattiGame) GetStake() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundWinnerIdx モック
func (_m *MockTeenPattiGame) GetRoundWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// IsShowdown モック
func (_m *MockTeenPattiGame) IsShowdown() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// CanRequestSideShow モック
func (_m *MockTeenPattiGame) CanRequestSideShow() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetSideShowRequester モック
func (_m *MockTeenPattiGame) GetSideShowRequester() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetSideShowTarget モック
func (_m *MockTeenPattiGame) GetSideShowTarget() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLastSideShow モック
func (_m *MockTeenPattiGame) GetLastSideShow() (int, int, int, bool) {
	ret := _m.Called()
	return ret.Int(0), ret.Int(1), ret.Int(2), ret.Bool(3)
}

// GetGameEndFlag モック
func (_m *MockTeenPattiGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetMatchWinnerIdx モック
func (_m *MockTeenPattiGame) GetMatchWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockTeenPattiGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockTeenPattiGame) GetPlayer(i int) *domain.TeenPattiPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.TeenPattiPlayer)
	}
	return nil
}

// IsHumanTurn モック
func (_m *MockTeenPattiGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// CanShow モック
func (_m *MockTeenPattiGame) CanShow() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetHint モック
func (_m *MockTeenPattiGame) GetHint() *domain.TeenPattiHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.TeenPattiHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockTeenPattiGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
