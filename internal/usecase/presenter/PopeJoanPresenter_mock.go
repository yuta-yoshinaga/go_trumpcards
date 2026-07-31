//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPopeJoanPresenter ポープ・ジョーン プレゼンターモック
type MockPopeJoanPresenter struct {
	MockGamePresenter[interfaces.PopeJoanGame]
}

// HintOutput モック
func (_m *MockPopeJoanPresenter) HintOutput(c interfaces.PopeJoanGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
