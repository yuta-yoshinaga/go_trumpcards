//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBigBenPresenter ビッグ・ベン プレゼンターモック
type MockBigBenPresenter struct {
	MockGamePresenter[interfaces.BigBenGame]
}

// HintOutput モック
func (_m *MockBigBenPresenter) HintOutput(gc interfaces.BigBenGame) string {
	ret := _m.Called(gc)
	return ret.Get(0).(string)
}
