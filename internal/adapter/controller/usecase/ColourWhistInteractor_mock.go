//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockColourWhistInteractor カラーホイストインタラクターモック
type MockColourWhistInteractor struct {
	mock.Mock
}

func (m *MockColourWhistInteractor) Reset() string { return m.Called().String(0) }

func (m *MockColourWhistInteractor) ResetWithConfig(cfg domain.ColourWhistConfig) string {
	return m.Called(cfg).String(0)
}

func (m *MockColourWhistInteractor) Bid(contract int) string { return m.Called(contract).String(0) }

func (m *MockColourWhistInteractor) Call(trumpSuit int) string {
	return m.Called(trumpSuit).String(0)
}

func (m *MockColourWhistInteractor) PlayCard(cardIndex int) string {
	return m.Called(cardIndex).String(0)
}

func (m *MockColourWhistInteractor) NextRound() string { return m.Called().String(0) }

func (m *MockColourWhistInteractor) GiveUp() string { return m.Called().String(0) }

func (m *MockColourWhistInteractor) GetConfig() domain.ColourWhistConfig {
	return m.Called().Get(0).(domain.ColourWhistConfig)
}

func (m *MockColourWhistInteractor) Hint() string { return m.Called().String(0) }

func (m *MockColourWhistInteractor) ActionLog() string { return m.Called().String(0) }

// Snapshot モック
func (m *MockColourWhistInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}
