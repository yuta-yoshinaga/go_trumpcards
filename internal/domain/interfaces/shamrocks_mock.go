//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockShamrocksGame is a testify mock for ShamrocksGame.
type MockShamrocksGame struct{ mock.Mock }

func (m *MockShamrocksGame) GetActionLog() []*domain.ActionLogEntry {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).([]*domain.ActionLogEntry)
}
func (m *MockShamrocksGame) GetGameEndFlag() bool               { return m.Called().Bool(0) }
func (m *MockShamrocksGame) Reset()                             { m.Called() }
func (m *MockShamrocksGame) MoveFanToFan(from, to int) error    { return m.Called(from, to).Error(0) }
func (m *MockShamrocksGame) MoveFanToFoundation(from int) error { return m.Called(from).Error(0) }
func (m *MockShamrocksGame) GiveUp()                            { m.Called() }
func (m *MockShamrocksGame) AutoComplete() error                { return m.Called().Error(0) }
func (m *MockShamrocksGame) Undo() error                        { return m.Called().Error(0) }
func (m *MockShamrocksGame) CanUndo() bool                      { return m.Called().Bool(0) }
func (m *MockShamrocksGame) UndoN(n int) error                  { return m.Called(n).Error(0) }
func (m *MockShamrocksGame) GetHint() *domain.ShamrocksHint {
	ret := m.Called().Get(0)
	if ret == nil {
		return nil
	}
	return ret.(*domain.ShamrocksHint)
}
func (m *MockShamrocksGame) GetPhase() domain.ShamrocksPhase {
	return m.Called().Get(0).(domain.ShamrocksPhase)
}
func (m *MockShamrocksGame) GetMoveCount() int     { return m.Called().Int(0) }
func (m *MockShamrocksGame) HasAnyLegalMove() bool { return m.Called().Bool(0) }
func (m *MockShamrocksGame) GetFans() [][]*domain.Card {
	ret := m.Called().Get(0)
	if ret == nil {
		return nil
	}
	return ret.([][]*domain.Card)
}
func (m *MockShamrocksGame) GetFoundation() [domain.ShamrocksFoundationCnt][]*domain.Card {
	return m.Called().Get(0).([domain.ShamrocksFoundationCnt][]*domain.Card)
}
func (m *MockShamrocksGame) GetMovableFans() []int {
	ret := m.Called().Get(0)
	if ret == nil {
		return nil
	}
	return ret.([]int)
}
func (m *MockShamrocksGame) LegalTargets(fan int) ([]int, []int) {
	ret := m.Called(fan)
	return ret.Get(0).([]int), ret.Get(1).([]int)
}
