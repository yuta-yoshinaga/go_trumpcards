//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCaribbeanStudPresenter カリビアンスタッドポーカープレゼンターモック
type MockCaribbeanStudPresenter struct {
	MockGamePresenter[interfaces.CaribbeanStudGame]
}

// HintOutput モック
func (_m *MockCaribbeanStudPresenter) HintOutput(cs interfaces.CaribbeanStudGame) string {
	ret := _m.Called(cs)
	return ret.Get(0).(string)
}
