//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPageOnePresenter ページワンプレゼンターモック
type MockPageOnePresenter struct {
	MockGamePresenter[interfaces.PageOneGame]
}

// HintOutput モック
func (_m *MockPageOnePresenter) HintOutput(g interfaces.PageOneGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
