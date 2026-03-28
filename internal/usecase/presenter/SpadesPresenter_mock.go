//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSpadesPresenter スペードプレゼンターモック
type MockSpadesPresenter struct {
	MockGamePresenter[interfaces.SpadesGame]
}

// HintOutput モック
func (_m *MockSpadesPresenter) HintOutput(s interfaces.SpadesGame) string {
	ret := _m.Called(s)
	return ret.Get(0).(string)
}
