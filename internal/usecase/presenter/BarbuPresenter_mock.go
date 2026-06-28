//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBarbuPresenter はバルブプレゼンターモック。
type MockBarbuPresenter struct {
	MockGamePresenter[interfaces.BarbuGame]
}

// HintOutput モック
func (_m *MockBarbuPresenter) HintOutput(g interfaces.BarbuGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
