//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSalicLawPresenter サリカ法典 プレゼンターモック
type MockSalicLawPresenter struct {
	MockGamePresenter[interfaces.SalicLawGame]
}

// HintOutput モック
func (_m *MockSalicLawPresenter) HintOutput(c interfaces.SalicLawGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
