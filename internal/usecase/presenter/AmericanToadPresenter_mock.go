//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockAmericanToadPresenter アメリカン・トード プレゼンターモック
type MockAmericanToadPresenter struct {
	MockGamePresenter[interfaces.AmericanToadGame]
}

// HintOutput モック
func (_m *MockAmericanToadPresenter) HintOutput(at interfaces.AmericanToadGame) string {
	ret := _m.Called(at)
	return ret.Get(0).(string)
}
