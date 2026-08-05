//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSpiteAndMaliceGame Spite & Malice ゲームモック
type MockSpiteAndMaliceGame struct {
	mock.Mock
}

func (_m *MockSpiteAndMaliceGame) Reset() {
	_m.Called()
}

func (_m *MockSpiteAndMaliceGame) PlayFromHand(handIdx, foundationIdx int) error {
	return _m.Called(handIdx, foundationIdx).Error(0)
}

func (_m *MockSpiteAndMaliceGame) PlayFromGoal(foundationIdx int) error {
	return _m.Called(foundationIdx).Error(0)
}

func (_m *MockSpiteAndMaliceGame) PlayFromSide(sideIdx, foundationIdx int) error {
	return _m.Called(sideIdx, foundationIdx).Error(0)
}

func (_m *MockSpiteAndMaliceGame) Discard(handIdx, sideIdx int) error {
	return _m.Called(handIdx, sideIdx).Error(0)
}

func (_m *MockSpiteAndMaliceGame) CpuStep() error {
	return _m.Called().Error(0)
}

func (_m *MockSpiteAndMaliceGame) AutoComplete() error {
	return _m.Called().Error(0)
}

func (_m *MockSpiteAndMaliceGame) CanAutoComplete() bool {
	return _m.Called().Bool(0)
}

func (_m *MockSpiteAndMaliceGame) IsCpuTurn() bool {
	return _m.Called().Bool(0)
}

func (_m *MockSpiteAndMaliceGame) GetHint() *domain.SpiteAndMaliceHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.SpiteAndMaliceHint)
}

func (_m *MockSpiteAndMaliceGame) GetPhase() domain.SpiteAndMalicePhase {
	return _m.Called().Get(0).(domain.SpiteAndMalicePhase)
}

func (_m *MockSpiteAndMaliceGame) GetCurrent() int {
	return _m.Called().Int(0)
}

func (_m *MockSpiteAndMaliceGame) GetMoveCount() int {
	return _m.Called().Int(0)
}

func (_m *MockSpiteAndMaliceGame) GetWinner() int {
	return _m.Called().Int(0)
}

func (_m *MockSpiteAndMaliceGame) GetStockSize() int {
	return _m.Called().Int(0)
}

func (_m *MockSpiteAndMaliceGame) GetCompletedSize() int {
	return _m.Called().Int(0)
}

func (_m *MockSpiteAndMaliceGame) GetFoundations() [domain.SpiteAndMaliceFoundationCnt][]*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return [domain.SpiteAndMaliceFoundationCnt][]*domain.Card{}
	}
	return v.([domain.SpiteAndMaliceFoundationCnt][]*domain.Card)
}

func (_m *MockSpiteAndMaliceGame) GetFoundationTopValue(foundationIdx int) int {
	return _m.Called(foundationIdx).Int(0)
}

func (_m *MockSpiteAndMaliceGame) IsGoalTopPlayable(playerIdx int) bool {
	return _m.Called(playerIdx).Bool(0)
}

func (_m *MockSpiteAndMaliceGame) GetPlayer(idx int) *domain.SpiteAndMalicePlayer {
	ret := _m.Called(idx)
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.SpiteAndMalicePlayer)
}

func (_m *MockSpiteAndMaliceGame) GetConfig() domain.SpiteAndMaliceConfig {
	return _m.Called().Get(0).(domain.SpiteAndMaliceConfig)
}

func (_m *MockSpiteAndMaliceGame) SetConfig(cfg domain.SpiteAndMaliceConfig) {
	_m.Called(cfg)
}

func (_m *MockSpiteAndMaliceGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockSpiteAndMaliceGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
