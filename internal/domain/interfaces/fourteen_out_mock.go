//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFourteenOutGame はフォーティーンアウト・ソリティアのモック。
type MockFourteenOutGame struct {
	mock.Mock
}

// CountRemovablePairs は取り除ける組の数を返す。
func (m *MockFourteenOutGame) CountRemovablePairs() int {
	args := m.Called()
	return args.Int(0)
}

// Reset モック
func (_m *MockFourteenOutGame) Reset() { _m.Called() }

// Remove モック
func (_m *MockFourteenOutGame) Remove(c1, c2 int) error {
	ret := _m.Called(c1, c2)
	return ret.Error(0)
}

// Undo モック
func (_m *MockFourteenOutGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

// CanUndo モック
func (_m *MockFourteenOutGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GiveUp モック
func (_m *MockFourteenOutGame) GiveUp() { _m.Called() }

// Hint モック
func (_m *MockFourteenOutGame) Hint() *domain.FourteenOutHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.FourteenOutHint)
}

// GetPhase モック
func (_m *MockFourteenOutGame) GetPhase() domain.FourteenOutPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.FourteenOutPhase)
}

// GetColumns mocks the GetColumns call.
func (_m *MockFourteenOutGame) GetColumns() [][]*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([][]*domain.Card); ok {
		return v
	}
	return nil
}

// GetRemovedCount モック
func (_m *MockFourteenOutGame) GetRemovedCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// IsComplete モック
func (_m *MockFourteenOutGame) IsComplete() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsStalemate モック
func (_m *MockFourteenOutGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetActionLog モック
func (_m *MockFourteenOutGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag モック
func (_m *MockFourteenOutGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
