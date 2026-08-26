//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDilotiGame はディロティのゲームモック。
type MockDilotiGame struct {
	mock.Mock
}

// GetActionLog モック
func (_m *MockDilotiGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v, _ := ret.Get(0).([]*domain.ActionLogEntry)
	return v
}

// Reset モック
func (_m *MockDilotiGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockDilotiGame) NextRound() { _m.Called() }

// PlayerPlay モック
func (_m *MockDilotiGame) PlayerPlay(handIdx int, action string, tableIdxs, declIdxs []int, declValue int) error {
	ret := _m.Called(handIdx, action, tableIdxs, declIdxs, declValue)
	err, _ := ret.Get(0).(error)
	return err
}

// CpuPlay モック
func (_m *MockDilotiGame) CpuPlay() { _m.Called() }

// GetConfig モック
func (_m *MockDilotiGame) GetConfig() domain.DilotiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.DilotiConfig)
}

// SetConfig モック
func (_m *MockDilotiGame) SetConfig(cfg domain.DilotiConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockDilotiGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockDilotiGame) GetPhase() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// IsHumanTurn モック
func (_m *MockDilotiGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetTable モック
func (_m *MockDilotiGame) GetTable() []*domain.Card {
	ret := _m.Called()
	v, _ := ret.Get(0).([]*domain.Card)
	return v
}

// GetDeclarations モック
func (_m *MockDilotiGame) GetDeclarations() []*domain.DilotiDeclaration {
	ret := _m.Called()
	v, _ := ret.Get(0).([]*domain.DilotiDeclaration)
	return v
}

// GetRoundNumber モック
func (_m *MockDilotiGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockDilotiGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockDilotiGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLastCapturer モック
func (_m *MockDilotiGame) GetLastCapturer() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDeckRemaining モック
func (_m *MockDilotiGame) GetDeckRemaining() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTakeOptions モック
func (_m *MockDilotiGame) GetTakeOptions(seat, handIdx int) []domain.DilotiTake {
	ret := _m.Called(seat, handIdx)
	v, _ := ret.Get(0).([]domain.DilotiTake)
	return v
}

// GetDeclareOptions モック
func (_m *MockDilotiGame) GetDeclareOptions(seat, handIdx int) []domain.DilotiDeclCandidate {
	ret := _m.Called(seat, handIdx)
	v, _ := ret.Get(0).([]domain.DilotiDeclCandidate)
	return v
}

// CanTrail モック
func (_m *MockDilotiGame) CanTrail(seat, handIdx int) bool {
	ret := _m.Called(seat, handIdx)
	return ret.Get(0).(bool)
}

// GetPlayerCnt モック
func (_m *MockDilotiGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockDilotiGame) GetPlayer(i int) *domain.DilotiPlayer {
	ret := _m.Called(i)
	v, _ := ret.Get(0).(*domain.DilotiPlayer)
	return v
}

// GetLastResult モック
func (_m *MockDilotiGame) GetLastResult() *domain.DilotiRoundResult {
	ret := _m.Called()
	v, _ := ret.Get(0).(*domain.DilotiRoundResult)
	return v
}

// GetWinnerIdx モック
func (_m *MockDilotiGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetHint モック
func (_m *MockDilotiGame) GetHint() *domain.DilotiHint {
	ret := _m.Called()
	v, _ := ret.Get(0).(*domain.DilotiHint)
	return v
}
