//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBanLuckPresenter バンラックプレゼンターモック
type MockBanLuckPresenter struct {
	MockGamePresenter[interfaces.BanLuckGame]
}

// HintOutput モック
func (_m *MockBanLuckPresenter) HintOutput(s interfaces.BanLuckGame) string {
	return _m.Called(s).Get(0).(string)
}
