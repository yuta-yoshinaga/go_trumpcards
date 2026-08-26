//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDehlaPakadGame はデーラ・パカドのゲームモック。
type MockDehlaPakadGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockDehlaPakadGame) Reset() {
	_m.Called()
}

// NextHand モック
func (_m *MockDehlaPakadGame) NextHand() {
	_m.Called()
}

// SelectTrump モック
func (_m *MockDehlaPakadGame) SelectTrump(suit int) error {
	ret := _m.Called(suit)
	if err, ok := ret.Get(0).(error); ok {
		return err
	}
	return nil
}

// CpuSelectTrump モック
func (_m *MockDehlaPakadGame) CpuSelectTrump() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockDehlaPakadGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	if err, ok := ret.Get(0).(error); ok {
		return err
	}
	return nil
}

// CpuPlay モック
func (_m *MockDehlaPakadGame) CpuPlay() {
	_m.Called()
}

// GetConfig モック
func (_m *MockDehlaPakadGame) GetConfig() domain.DehlaPakadConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.DehlaPakadConfig)
}

// SetConfig モック
func (_m *MockDehlaPakadGame) SetConfig(cfg domain.DehlaPakadConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockDehlaPakadGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockDehlaPakadGame) GetPhase() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// IsHumanTurn モック
func (_m *MockDehlaPakadGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetHandNumber モック
func (_m *MockDehlaPakadGame) GetHandNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockDehlaPakadGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrumpChooserIdx モック
func (_m *MockDehlaPakadGame) GetTrumpChooserIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrumpSuit モック
func (_m *MockDehlaPakadGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockDehlaPakadGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockDehlaPakadGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTurn モック
func (_m *MockDehlaPakadGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLeadPlayerIdx モック
func (_m *MockDehlaPakadGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockDehlaPakadGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetLastTrick モック
func (_m *MockDehlaPakadGame) GetLastTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetLastTrickWinner モック
func (_m *MockDehlaPakadGame) GetLastTrickWinner() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPrevTrickWinner モック
func (_m *MockDehlaPakadGame) GetPrevTrickWinner() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCentrePile モック
func (_m *MockDehlaPakadGame) GetCentrePile() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetCentrePileTens モック
func (_m *MockDehlaPakadGame) GetCentrePileTens() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayableIndices モック
func (_m *MockDehlaPakadGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetTeamTens モック
func (_m *MockDehlaPakadGame) GetTeamTens() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetTeamKots モック
func (_m *MockDehlaPakadGame) GetTeamKots() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetStreakTeam モック
func (_m *MockDehlaPakadGame) GetStreakTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetStreakCount モック
func (_m *MockDehlaPakadGame) GetStreakCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockDehlaPakadGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockDehlaPakadGame) GetPlayer(i int) *domain.DehlaPakadPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.DehlaPakadPlayer); ok {
		return v
	}
	return nil
}

// GetLastResult モック
func (_m *MockDehlaPakadGame) GetLastResult() *domain.DehlaPakadHandResult {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.DehlaPakadHandResult); ok {
		return v
	}
	return nil
}

// GetHandHistory モック
func (_m *MockDehlaPakadGame) GetHandHistory() []*domain.DehlaPakadHandResult {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.DehlaPakadHandResult); ok {
		return v
	}
	return nil
}

// GetWinnerTeam モック
func (_m *MockDehlaPakadGame) GetWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetHint モック
func (_m *MockDehlaPakadGame) GetHint() *domain.DehlaPakadHint {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.DehlaPakadHint); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockDehlaPakadGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
