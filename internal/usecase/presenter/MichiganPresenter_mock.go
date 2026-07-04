//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMichiganPresenter はミシガン (Michigan) プレゼンターモック。
type MockMichiganPresenter struct {
	MockGamePresenter[interfaces.MichiganGame]
}

// HintOutput モック
func (_m *MockMichiganPresenter) HintOutput(g interfaces.MichiganGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
