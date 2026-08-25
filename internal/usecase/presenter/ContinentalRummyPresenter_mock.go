//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockContinentalRummyPresenter はコンチネンタル・ラミーのプレゼンターモック。
type MockContinentalRummyPresenter struct {
	MockGamePresenter[interfaces.ContinentalRummyGame]
}

// HintOutput モック
func (_m *MockContinentalRummyPresenter) HintOutput(g interfaces.ContinentalRummyGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
