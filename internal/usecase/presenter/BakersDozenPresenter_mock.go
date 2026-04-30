//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBakersDozenPresenter ベーカーズダズンプレゼンターモック
type MockBakersDozenPresenter struct {
	MockGamePresenter[interfaces.BakersDozenGame]
}

// HintOutput モック
func (_m *MockBakersDozenPresenter) HintOutput(bd interfaces.BakersDozenGame) string {
	ret := _m.Called(bd)
	return ret.Get(0).(string)
}
