package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPyramidPresenter ピラミッドプレゼンターモック
type MockPyramidPresenter struct {
	MockGamePresenter[interfaces.PyramidGame]
}

// HintOutput モック
func (_m *MockPyramidPresenter) HintOutput(p interfaces.PyramidGame) string {
	ret := _m.Called(p)
	return ret.Get(0).(string)
}
