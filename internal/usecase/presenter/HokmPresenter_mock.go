//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockHokmPresenter ホクムプレゼンターモック
type MockHokmPresenter struct {
	MockGamePresenter[interfaces.HokmGame]
}

// HintOutput モック
func (_m *MockHokmPresenter) HintOutput(h interfaces.HokmGame) string {
	return _m.Called(h).Get(0).(string)
}
