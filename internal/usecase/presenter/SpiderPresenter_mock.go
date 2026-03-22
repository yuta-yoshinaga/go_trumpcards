package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSpiderPresenter スパイダーソリティアプレゼンターモック
type MockSpiderPresenter struct {
	MockGamePresenter[interfaces.SpiderGame]
}

// HintOutput モック
func (_m *MockSpiderPresenter) HintOutput(s interfaces.SpiderGame) string {
	ret := _m.Called(s)
	return ret.Get(0).(string)
}
