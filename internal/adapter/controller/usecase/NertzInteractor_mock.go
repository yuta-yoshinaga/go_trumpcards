//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockNertzInteractor Nertz インタラクターモック
type MockNertzInteractor struct {
	mock.Mock
}

func (_m *MockNertzInteractor) Reset() string {
	return _m.Called().Get(0).(string)
}

func (_m *MockNertzInteractor) ResetWithConfig(cfg domain.NertzConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

func (_m *MockNertzInteractor) NextRound() string {
	return _m.Called().Get(0).(string)
}

func (_m *MockNertzInteractor) Draw(playerIdx int) string {
	return _m.Called(playerIdx).Get(0).(string)
}

func (_m *MockNertzInteractor) MoveNertzToFoundation(playerIdx, foundationIdx int) string {
	return _m.Called(playerIdx, foundationIdx).Get(0).(string)
}

func (_m *MockNertzInteractor) MoveNertzToTableau(playerIdx, toCol int) string {
	return _m.Called(playerIdx, toCol).Get(0).(string)
}

func (_m *MockNertzInteractor) MoveWasteToFoundation(playerIdx, foundationIdx int) string {
	return _m.Called(playerIdx, foundationIdx).Get(0).(string)
}

func (_m *MockNertzInteractor) MoveWasteToTableau(playerIdx, toCol int) string {
	return _m.Called(playerIdx, toCol).Get(0).(string)
}

func (_m *MockNertzInteractor) MoveTableauToFoundation(playerIdx, fromCol, foundationIdx int) string {
	return _m.Called(playerIdx, fromCol, foundationIdx).Get(0).(string)
}

func (_m *MockNertzInteractor) MoveTableauToTableau(playerIdx, fromCol, fromIdx, toCol int) string {
	return _m.Called(playerIdx, fromCol, fromIdx, toCol).Get(0).(string)
}

func (_m *MockNertzInteractor) Tick() string {
	return _m.Called().Get(0).(string)
}

func (_m *MockNertzInteractor) Hint() string {
	return _m.Called().Get(0).(string)
}

func (_m *MockNertzInteractor) Undo() string {
	return _m.Called().Get(0).(string)
}

func (_m *MockNertzInteractor) ActionLog() string {
	return _m.Called().Get(0).(string)
}

func (_m *MockNertzInteractor) GetConfig() domain.NertzConfig {
	return _m.Called().Get(0).(domain.NertzConfig)
}

func (_m *MockNertzInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
