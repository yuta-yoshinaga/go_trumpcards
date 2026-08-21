//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockFortressPresenter Fortress プレゼンターモック
type MockFortressPresenter struct {
	MockGamePresenter[interfaces.FortressGame]
}

// HintOutput モック
func (_m *MockFortressPresenter) HintOutput(bc interfaces.FortressGame) string {
	ret := _m.Called(bc)
	return ret.Get(0).(string)
}
