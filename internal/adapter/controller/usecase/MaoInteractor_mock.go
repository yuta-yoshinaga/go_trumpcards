//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMaoInteractor モック
type MockMaoInteractor struct {
	mock.Mock
}

func (_m *MockMaoInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockMaoInteractor) ResetWithConfig(cfg domain.MaoConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockMaoInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockMaoInteractor) ChooseSuit(suit int) string {
	return _m.Called(suit).String(0)
}

func (_m *MockMaoInteractor) Draw() string {
	return _m.Called().String(0)
}

func (_m *MockMaoInteractor) Declare() string {
	return _m.Called().String(0)
}

func (_m *MockMaoInteractor) SkipDeclare() string {
	return _m.Called().String(0)
}

func (_m *MockMaoInteractor) DeclareWord(word string) string {
	return _m.Called(word).String(0)
}

func (_m *MockMaoInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockMaoInteractor) GetConfig() domain.MaoConfig {
	return _m.Called().Get(0).(domain.MaoConfig)
}

func (_m *MockMaoInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockMaoInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

// IsHumanChooseSuitTurn モック
func (_m *MockMaoInteractor) IsHumanChooseSuitTurn() bool {
	return _m.Called().Bool(0)
}

// IsHumanDeclareTurn モック
func (_m *MockMaoInteractor) IsHumanDeclareTurn() bool {
	return _m.Called().Bool(0)
}

// IsHumanAwaitingWord モック
func (_m *MockMaoInteractor) IsHumanAwaitingWord() bool {
	return _m.Called().Bool(0)
}
