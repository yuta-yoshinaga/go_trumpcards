//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHorseGame は H.O.R.S.E. のゲームモック。
type MockHorseGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockHorseGame) Reset() { _m.Called() }

// NextHand モック
func (_m *MockHorseGame) NextHand() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerAction モック
func (_m *MockHorseGame) PlayerAction(action, amount, humanPlayMs int) error {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Error(0)
}

// GetConfig モック
func (_m *MockHorseGame) GetConfig() domain.HorseConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.HorseConfig)
}

// SetConfig モック
func (_m *MockHorseGame) SetConfig(cfg domain.HorseConfig) { _m.Called(cfg) }

// GetPhase モック
func (_m *MockHorseGame) GetPhase() domain.HorsePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.HorsePhase)
}

// GetGameEndFlag モック
func (_m *MockHorseGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetDiscipline モック
func (_m *MockHorseGame) GetDiscipline() domain.HorseDiscipline {
	ret := _m.Called()
	return ret.Get(0).(domain.HorseDiscipline)
}

// GetDisciplineLetter モック
func (_m *MockHorseGame) GetDisciplineLetter() string {
	ret := _m.Called()
	return ret.String(0)
}

// GetHandInDiscipline モック
func (_m *MockHorseGame) GetHandInDiscipline() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHandNumber モック
func (_m *MockHorseGame) GetHandNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetSeatChips モック
func (_m *MockHorseGame) GetSeatChips(i int) int {
	ret := _m.Called(i)
	return ret.Int(0)
}

// GetSeatName モック
func (_m *MockHorseGame) GetSeatName(i int) string {
	ret := _m.Called(i)
	return ret.String(0)
}

// GetSeatIsHuman モック
func (_m *MockHorseGame) GetSeatIsHuman(i int) bool {
	ret := _m.Called(i)
	return ret.Bool(0)
}

// GetSeatCount モック
func (_m *MockHorseGame) GetSeatCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetHumanSeat モック
func (_m *MockHorseGame) GetHumanSeat() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCurrentTurn モック
func (_m *MockHorseGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// IsHumanTurn モック
func (_m *MockHorseGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetPot モック
func (_m *MockHorseGame) GetPot() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTablePhase モック
func (_m *MockHorseGame) GetTablePhase() int {
	ret := _m.Called()
	return ret.Int(0)
}

// WinnerSeat モック
func (_m *MockHorseGame) WinnerSeat() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetActionLog モック
func (_m *MockHorseGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
