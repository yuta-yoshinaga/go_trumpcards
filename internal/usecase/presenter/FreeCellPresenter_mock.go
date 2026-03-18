package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockFreeCellPresenter フリーセルプレゼンターモック
type MockFreeCellPresenter struct {
	MockGamePresenter[interfaces.FreeCellGame]
}

// HintOutput モック
func (_m *MockFreeCellPresenter) HintOutput(f interfaces.FreeCellGame) string {
	ret := _m.Called(f)
	return ret.Get(0).(string)
}
