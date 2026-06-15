//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockFortyFivesPresenter オークション・フォーティファイブズのプレゼンターモック
type MockFortyFivesPresenter struct {
	MockGamePresenter[interfaces.FortyFivesGame]
}

// HintOutput モック
func (_m *MockFortyFivesPresenter) HintOutput(g interfaces.FortyFivesGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
