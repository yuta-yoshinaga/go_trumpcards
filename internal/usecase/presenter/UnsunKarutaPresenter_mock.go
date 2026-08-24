//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockUnsunKarutaPresenter はうんすんカルタのプレゼンターモック。
type MockUnsunKarutaPresenter struct {
	MockGamePresenter[interfaces.UnsunKarutaGame]
}

// HintOutput モック
func (_m *MockUnsunKarutaPresenter) HintOutput(g interfaces.UnsunKarutaGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
