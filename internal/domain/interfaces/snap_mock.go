//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSnapGame スナップゲームモック
type MockSnapGame struct {
	mock.Mock
}

func (m *MockSnapGame) Reset()  { m.Called() }
func (m *MockSnapGame) GiveUp() { m.Called() }

func (m *MockSnapGame) PlayerStep() error { return m.Called().Error(0) }
func (m *MockSnapGame) PlayerSnap() error { return m.Called().Error(0) }

func (m *MockSnapGame) Tick() domain.SnapPendingKind {
	return m.Called().Get(0).(domain.SnapPendingKind)
}

func (m *MockSnapGame) GetConfig() domain.SnapConfig {
	return m.Called().Get(0).(domain.SnapConfig)
}

func (m *MockSnapGame) SetConfig(cfg domain.SnapConfig) { m.Called(cfg) }

func (m *MockSnapGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockSnapGame) GetPhase() domain.SnapPhase {
	return m.Called().Get(0).(domain.SnapPhase)
}

func (m *MockSnapGame) IsSnapAvailable() bool  { return m.Called().Bool(0) }
func (m *MockSnapGame) GetCenterPileSize() int { return m.Called().Int(0) }
func (m *MockSnapGame) GetCurrentTurnIdx() int { return m.Called().Int(0) }
func (m *MockSnapGame) GetPlayerCnt() int      { return m.Called().Int(0) }
func (m *MockSnapGame) GetWinnerIdx() int      { return m.Called().Int(0) }

func (m *MockSnapGame) GetPending() domain.SnapPending {
	return m.Called().Get(0).(domain.SnapPending)
}

func (m *MockSnapGame) GetLastEvent() domain.SnapLastEvent {
	return m.Called().Get(0).(domain.SnapLastEvent)
}

func (m *MockSnapGame) GetCenterPile() []*domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

func (m *MockSnapGame) GetTopCard() *domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

func (m *MockSnapGame) GetPlayer(i int) *domain.SnapPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.SnapPlayer)
	}
	return nil
}

func (m *MockSnapGame) GetHint() *domain.SnapHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.SnapHint)
	}
	return nil
}

func (m *MockSnapGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
