//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockNainJaunePresenter ル・ナン・ジョーヌ プレゼンターモック
type MockNainJaunePresenter struct {
	MockGamePresenter[interfaces.NainJauneGame]
}

// HintOutput モック
func (_m *MockNainJaunePresenter) HintOutput(c interfaces.NainJauneGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
