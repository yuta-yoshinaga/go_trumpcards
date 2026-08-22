//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCaribbeanDrawPresenter カリビアン・ドロー・ポーカープレゼンターモック
type MockCaribbeanDrawPresenter struct {
	MockGamePresenter[interfaces.CaribbeanDrawGame]
}

// HintOutput モック
func (_m *MockCaribbeanDrawPresenter) HintOutput(cs interfaces.CaribbeanDrawGame) string {
	ret := _m.Called(cs)
	return ret.Get(0).(string)
}
