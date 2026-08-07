//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockLetItRidePresenter レット・イット・ライドプレゼンターモック
type MockLetItRidePresenter struct {
	MockGamePresenter[interfaces.LetItRideGame]
}

// PullConfirmOutput モック
func (_m *MockLetItRidePresenter) PullConfirmOutput(g interfaces.LetItRideGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
