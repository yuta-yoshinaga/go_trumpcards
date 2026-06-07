//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSchnapsenPresenter シュナプセンプレゼンターモック
type MockSchnapsenPresenter struct {
	MockGamePresenter[interfaces.SchnapsenGame]
}

// HintOutput モック
func (_m *MockSchnapsenPresenter) HintOutput(s interfaces.SchnapsenGame) string {
	ret := _m.Called(s)
	return ret.Get(0).(string)
}
