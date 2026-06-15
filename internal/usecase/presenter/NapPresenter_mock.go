//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockNapPresenter ナップのプレゼンターモック
type MockNapPresenter struct {
	MockGamePresenter[interfaces.NapGame]
}

// HintOutput モック
func (_m *MockNapPresenter) HintOutput(g interfaces.NapGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
