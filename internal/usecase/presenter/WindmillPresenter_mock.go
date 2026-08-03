//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockWindmillPresenter ウィンドミル プレゼンターモック
type MockWindmillPresenter struct {
	MockGamePresenter[interfaces.WindmillGame]
}

// HintOutput モック
func (_m *MockWindmillPresenter) HintOutput(w interfaces.WindmillGame) string {
	ret := _m.Called(w)
	return ret.Get(0).(string)
}
