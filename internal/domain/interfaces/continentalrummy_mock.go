//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockContinentalRummyGame はコンチネンタル・ラミーのゲームモック。
type MockContinentalRummyGame struct {
	mock.Mock
}

// **モックが口を満たしていることをここで固定する。** 生成し直したときに
// メソッドを取りこぼしても、使う側のテストが書かれるまで気付けない。
var _ ContinentalRummyGame = (*MockContinentalRummyGame)(nil)

// Reset モック
func (_m *MockContinentalRummyGame) Reset() {
	_m.Called()
}

// DrawStock モック
func (_m *MockContinentalRummyGame) DrawStock() error {
	ret := _m.Called()
	err, _ := ret.Get(0).(error)
	return err
}

// DrawDiscard モック
func (_m *MockContinentalRummyGame) DrawDiscard() error {
	ret := _m.Called()
	err, _ := ret.Get(0).(error)
	return err
}

// Discard モック
func (_m *MockContinentalRummyGame) Discard(i int) error {
	ret := _m.Called(i)
	err, _ := ret.Get(0).(error)
	return err
}

// GoOut モック
func (_m *MockContinentalRummyGame) GoOut(i int) error {
	ret := _m.Called(i)
	err, _ := ret.Get(0).(error)
	return err
}

// NextRound モック
func (_m *MockContinentalRummyGame) NextRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockContinentalRummyGame) GetConfig() domain.ContinentalRummyConfig {
	ret := _m.Called()
	r, _ := ret.Get(0).(domain.ContinentalRummyConfig)
	return r
}

// SetConfig モック
func (_m *MockContinentalRummyGame) SetConfig(cfg domain.ContinentalRummyConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockContinentalRummyGame) GetGameEndFlag() bool {
	ret := _m.Called()
	r, _ := ret.Get(0).(bool)
	return r
}

// GetPhase モック
func (_m *MockContinentalRummyGame) GetPhase() string {
	ret := _m.Called()
	r, _ := ret.Get(0).(string)
	return r
}

// IsHumanTurn モック
func (_m *MockContinentalRummyGame) IsHumanTurn() bool {
	ret := _m.Called()
	r, _ := ret.Get(0).(bool)
	return r
}

// GetCurrentPlayerIdx モック
func (_m *MockContinentalRummyGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	r, _ := ret.Get(0).(int)
	return r
}

// GetDealerIdx モック
func (_m *MockContinentalRummyGame) GetDealerIdx() int {
	ret := _m.Called()
	r, _ := ret.Get(0).(int)
	return r
}

// GetRoundNumber モック
func (_m *MockContinentalRummyGame) GetRoundNumber() int {
	ret := _m.Called()
	r, _ := ret.Get(0).(int)
	return r
}

// GetStockCount モック
func (_m *MockContinentalRummyGame) GetStockCount() int {
	ret := _m.Called()
	r, _ := ret.Get(0).(int)
	return r
}

// GetDiscardTop モック
func (_m *MockContinentalRummyGame) GetDiscardTop() *domain.Card {
	ret := _m.Called()
	r, _ := ret.Get(0).(*domain.Card)
	return r
}

// GetPlayerCnt モック
func (_m *MockContinentalRummyGame) GetPlayerCnt() int {
	ret := _m.Called()
	r, _ := ret.Get(0).(int)
	return r
}

// GetPlayer モック
func (_m *MockContinentalRummyGame) GetPlayer(i int) *domain.ContinentalRummyPlayer {
	ret := _m.Called(i)
	r, _ := ret.Get(0).(*domain.ContinentalRummyPlayer)
	return r
}

// GetLastResult モック
func (_m *MockContinentalRummyGame) GetLastResult() *domain.ContinentalRummyRoundResult {
	ret := _m.Called()
	r, _ := ret.Get(0).(*domain.ContinentalRummyRoundResult)
	return r
}

// GetWinnerIdx モック
func (_m *MockContinentalRummyGame) GetWinnerIdx() int {
	ret := _m.Called()
	r, _ := ret.Get(0).(int)
	return r
}

// CanGoOut モック
func (_m *MockContinentalRummyGame) CanGoOut() (int, bool) {
	ret := _m.Called()
	r0, _ := ret.Get(0).(int)
	r1, _ := ret.Get(1).(bool)
	return r0, r1
}

// GetHint モック
func (_m *MockContinentalRummyGame) GetHint() *domain.ContinentalRummyHint {
	ret := _m.Called()
	r, _ := ret.Get(0).(*domain.ContinentalRummyHint)
	return r
}

// GetActionLog モック
func (_m *MockContinentalRummyGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	r, _ := ret.Get(0).([]*domain.ActionLogEntry)
	return r
}
