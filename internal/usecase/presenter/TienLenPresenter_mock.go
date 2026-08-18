//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTienLenPresenter Tien Len プレゼンターモック
type MockTienLenPresenter struct {
	MockGamePresenter[interfaces.TienLenGame]
}

// HintOutput モック
func (_m *MockTienLenPresenter) HintOutput(tg interfaces.TienLenGame) string {
	ret := _m.Called(tg)
	return ret.Get(0).(string)
}
