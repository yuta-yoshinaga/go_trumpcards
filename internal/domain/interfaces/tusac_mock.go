//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTuSacGame 四色牌ゲームモック
type MockTuSacGame struct {
	mock.Mock
}

func (m *MockTuSacGame) Reset() { m.Called() }

func (m *MockTuSacGame) Draw(fromDiscard bool) error { return m.Called(fromDiscard).Error(0) }

func (m *MockTuSacGame) Meld(indexes []int) error { return m.Called(indexes).Error(0) }

func (m *MockTuSacGame) Discard(index int) error { return m.Called(index).Error(0) }

func (m *MockTuSacGame) NextRound() error { return m.Called().Error(0) }

func (m *MockTuSacGame) GetConfig() domain.TuSacConfig {
	return m.Called().Get(0).(domain.TuSacConfig)
}

func (m *MockTuSacGame) SetConfig(cfg domain.TuSacConfig) { m.Called(cfg) }

func (m *MockTuSacGame) GetPhase() domain.TuSacPhase {
	return m.Called().Get(0).(domain.TuSacPhase)
}

func (m *MockTuSacGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockTuSacGame) GetPlayers() []*domain.TuSacPlayer {
	return m.Called().Get(0).([]*domain.TuSacPlayer)
}

func (m *MockTuSacGame) GetTurnSeat() int { return m.Called().Int(0) }

func (m *MockTuSacGame) GetRoundNumber() int { return m.Called().Int(0) }

func (m *MockTuSacGame) GetStockCount() int { return m.Called().Int(0) }

func (m *MockTuSacGame) GetDiscardTop() *domain.Card {
	if c := m.Called().Get(0); c != nil {
		return c.(*domain.Card)
	}
	return nil
}

func (m *MockTuSacGame) GetDiscardCount() int { return m.Called().Int(0) }

func (m *MockTuSacGame) GetWentOutSeat() int { return m.Called().Int(0) }

func (m *MockTuSacGame) GetResults() []domain.TuSacResult {
	return m.Called().Get(0).([]domain.TuSacResult)
}

func (m *MockTuSacGame) HumanSeat() int { return m.Called().Int(0) }

func (m *MockTuSacGame) IsHumanTurn() bool { return m.Called().Bool(0) }

func (m *MockTuSacGame) WinnerSeat() int { return m.Called().Int(0) }

func (m *MockTuSacGame) GetHint() *domain.TuSacHint {
	if h := m.Called().Get(0); h != nil {
		return h.(*domain.TuSacHint)
	}
	return nil
}

func (m *MockTuSacGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
