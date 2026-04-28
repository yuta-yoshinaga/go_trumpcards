//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockNertzGame Nertz ゲームモック
type MockNertzGame struct {
	mock.Mock
}

func (_m *MockNertzGame) Reset() {
	_m.Called()
}

func (_m *MockNertzGame) ResetWithConfig(cfg domain.NertzConfig) {
	_m.Called(cfg)
}

func (_m *MockNertzGame) NextRound() {
	_m.Called()
}

func (_m *MockNertzGame) DrawStock(playerIdx int) error {
	return _m.Called(playerIdx).Error(0)
}

func (_m *MockNertzGame) MoveNertzToFoundation(playerIdx, foundationIdx int) error {
	return _m.Called(playerIdx, foundationIdx).Error(0)
}

func (_m *MockNertzGame) MoveNertzToTableau(playerIdx, toCol int) error {
	return _m.Called(playerIdx, toCol).Error(0)
}

func (_m *MockNertzGame) MoveWasteToFoundation(playerIdx, foundationIdx int) error {
	return _m.Called(playerIdx, foundationIdx).Error(0)
}

func (_m *MockNertzGame) MoveWasteToTableau(playerIdx, toCol int) error {
	return _m.Called(playerIdx, toCol).Error(0)
}

func (_m *MockNertzGame) MoveTableauToFoundation(playerIdx, fromCol, foundationIdx int) error {
	return _m.Called(playerIdx, fromCol, foundationIdx).Error(0)
}

func (_m *MockNertzGame) MoveTableauToTableau(playerIdx, fromCol, fromIdx, toCol int) error {
	return _m.Called(playerIdx, fromCol, fromIdx, toCol).Error(0)
}

func (_m *MockNertzGame) Tick() []*domain.NertzAction {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.NertzAction)
}

func (_m *MockNertzGame) GetHint() *domain.NertzHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.NertzHint)
}

func (_m *MockNertzGame) CanUndo() bool {
	return _m.Called().Bool(0)
}

func (_m *MockNertzGame) Undo() error {
	return _m.Called().Error(0)
}

func (_m *MockNertzGame) GetPhase() domain.NertzPhase {
	return _m.Called().Get(0).(domain.NertzPhase)
}

func (_m *MockNertzGame) GetRoundNo() int {
	return _m.Called().Int(0)
}

func (_m *MockNertzGame) GetWinnerIdx() int {
	return _m.Called().Int(0)
}

func (_m *MockNertzGame) GetMatchWinner() int {
	return _m.Called().Int(0)
}

func (_m *MockNertzGame) GetConfig() domain.NertzConfig {
	return _m.Called().Get(0).(domain.NertzConfig)
}

func (_m *MockNertzGame) GetPlayers() []*domain.NertzPlayer {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.NertzPlayer)
}

func (_m *MockNertzGame) GetFoundations() []*domain.NertzFoundation {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.NertzFoundation)
}

func (_m *MockNertzGame) GetMoveCount() int {
	return _m.Called().Int(0)
}

func (_m *MockNertzGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockNertzGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
