//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMrsMopPresenter ミセス・モップソリティアプレゼンターモック
type MockMrsMopPresenter struct {
	MockGamePresenter[interfaces.MrsMopGame]
}

// HintOutput モック
func (_m *MockMrsMopPresenter) HintOutput(s interfaces.MrsMopGame) string {
	ret := _m.Called(s)
	return ret.Get(0).(string)
}
