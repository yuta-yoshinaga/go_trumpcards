//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockNapoleonsSquarePresenter ナポレオンズ・スクエア プレゼンターモック
type MockNapoleonsSquarePresenter struct {
	MockGamePresenter[interfaces.NapoleonsSquareGame]
}

// HintOutput モック
func (_m *MockNapoleonsSquarePresenter) HintOutput(ns interfaces.NapoleonsSquareGame) string {
	ret := _m.Called(ns)
	return ret.Get(0).(string)
}
