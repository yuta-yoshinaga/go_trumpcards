//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSevenTwentySevenPresenter はセブン・トゥエンティセブン (SevenTwentySeven) プレゼンターモック。
type MockSevenTwentySevenPresenter struct {
	MockGamePresenter[interfaces.SevenTwentySevenGame]
}

// HintOutput モック
func (_m *MockSevenTwentySevenPresenter) HintOutput(g interfaces.SevenTwentySevenGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
