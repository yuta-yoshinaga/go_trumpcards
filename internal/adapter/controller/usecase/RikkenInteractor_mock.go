//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRikkenInteractor リッケンインタラクターモック
type MockRikkenInteractor struct {
	mock.Mock
}

func (m *MockRikkenInteractor) Reset() string { return m.Called().String(0) }

func (m *MockRikkenInteractor) ResetWithConfig(cfg domain.RikkenConfig) string {
	return m.Called(cfg).String(0)
}

func (m *MockRikkenInteractor) Bid(contract int) string { return m.Called(contract).String(0) }

func (m *MockRikkenInteractor) Call(trumpSuit int) string { return m.Called(trumpSuit).String(0) }

func (m *MockRikkenInteractor) PlayCard(cardIndex int) string {
	return m.Called(cardIndex).String(0)
}

func (m *MockRikkenInteractor) NextRound() string { return m.Called().String(0) }

func (m *MockRikkenInteractor) GiveUp() string { return m.Called().String(0) }

func (m *MockRikkenInteractor) GetConfig() domain.RikkenConfig {
	return m.Called().Get(0).(domain.RikkenConfig)
}

func (m *MockRikkenInteractor) Hint() string { return m.Called().String(0) }

func (m *MockRikkenInteractor) ActionLog() string { return m.Called().String(0) }

// Snapshot モック
func (m *MockRikkenInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}
