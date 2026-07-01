//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTysiacPresenter サウザンド (Tysiąc) のプレゼンターモック
type MockTysiacPresenter struct {
	MockGamePresenter[interfaces.TysiacGame]
}

// HintOutput モック
func (_m *MockTysiacPresenter) HintOutput(g interfaces.TysiacGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
