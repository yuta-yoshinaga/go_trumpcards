//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTongitsPresenter Tongitsプレゼンターモック
type MockTongitsPresenter struct {
	MockGamePresenter[interfaces.TongitsGame]
}

// HintOutput モック
func (m *MockTongitsPresenter) HintOutput(g interfaces.TongitsGame) string {
	return m.Called(g).String(0)
}
