//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSomersetPresenter Somerset プレゼンターモック
type MockSomersetPresenter struct {
	MockGamePresenter[interfaces.SomersetGame]
}

// HintOutput モック
func (_m *MockSomersetPresenter) HintOutput(bc interfaces.SomersetGame) string {
	ret := _m.Called(bc)
	return ret.Get(0).(string)
}
