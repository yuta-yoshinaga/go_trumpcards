//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockDuchessPresenter ダッチェス プレゼンターモック
type MockDuchessPresenter struct {
	MockGamePresenter[interfaces.DuchessGame]
}

// HintOutput モック
func (_m *MockDuchessPresenter) HintOutput(d interfaces.DuchessGame) string {
	ret := _m.Called(d)
	return ret.Get(0).(string)
}
