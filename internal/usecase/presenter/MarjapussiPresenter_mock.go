//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMarjapussiPresenter マルヤプッシ (Marjapussi) のプレゼンターモック
type MockMarjapussiPresenter struct {
	MockGamePresenter[interfaces.MarjapussiGame]
}

// HintOutput モック
func (_m *MockMarjapussiPresenter) HintOutput(g interfaces.MarjapussiGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
