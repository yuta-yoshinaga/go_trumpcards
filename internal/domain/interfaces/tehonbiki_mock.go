//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTehonbikiGame 手本引きゲームモック
type MockTehonbikiGame struct {
	mock.Mock
}

func (m *MockTehonbikiGame) Reset() { m.Called() }

func (m *MockTehonbikiGame) PlaceBet(idx, bet int) error { return m.Called(idx, bet).Error(0) }

func (m *MockTehonbikiGame) NextRound() error { return m.Called().Error(0) }

func (m *MockTehonbikiGame) GetConfig() domain.TehonbikiConfig {
	return m.Called().Get(0).(domain.TehonbikiConfig)
}

func (m *MockTehonbikiGame) SetConfig(cfg domain.TehonbikiConfig) { m.Called(cfg) }

func (m *MockTehonbikiGame) GetPhase() domain.TehonbikiPhase {
	return m.Called().Get(0).(domain.TehonbikiPhase)
}

func (m *MockTehonbikiGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockTehonbikiGame) GetLayout() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}

func (m *MockTehonbikiGame) GetGate() *domain.Card {
	if c := m.Called().Get(0); c != nil {
		return c.(*domain.Card)
	}
	return nil
}

func (m *MockTehonbikiGame) GetPick() int { return m.Called().Int(0) }

func (m *MockTehonbikiGame) GetBet() int { return m.Called().Int(0) }

func (m *MockTehonbikiGame) GetResult() domain.TehonbikiResult {
	return m.Called().Get(0).(domain.TehonbikiResult)
}

func (m *MockTehonbikiGame) GetPayout() int { return m.Called().Int(0) }

func (m *MockTehonbikiGame) SuitCountInLayout(design int) int { return m.Called(design).Int(0) }

func (m *MockTehonbikiGame) RemainingOfSuit(design int) int { return m.Called(design).Int(0) }

func (m *MockTehonbikiGame) GetChips() int { return m.Called().Int(0) }

func (m *MockTehonbikiGame) GetRoundNumber() int { return m.Called().Int(0) }

func (m *MockTehonbikiGame) GetRemainingCards() int { return m.Called().Int(0) }

func (m *MockTehonbikiGame) GetHint() *domain.TehonbikiHint {
	if h := m.Called().Get(0); h != nil {
		return h.(*domain.TehonbikiHint)
	}
	return nil
}

func (m *MockTehonbikiGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
