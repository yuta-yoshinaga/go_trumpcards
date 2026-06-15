//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPreferencePresenter プレフェランスのプレゼンターモック
type MockPreferencePresenter struct {
	MockGamePresenter[interfaces.PreferenceGame]
}

// HintOutput モック
func (_m *MockPreferencePresenter) HintOutput(g interfaces.PreferenceGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
