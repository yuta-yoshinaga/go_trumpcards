//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBasraPresenter はバスラプレゼンターモック。
type MockBasraPresenter struct {
	MockGamePresenter[interfaces.BasraGame]
}

// HintOutput モック
func (_m *MockBasraPresenter) HintOutput(g interfaces.BasraGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
