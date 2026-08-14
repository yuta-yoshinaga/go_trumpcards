//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMonteBankPresenter モンテバンクプレゼンターモック
type MockMonteBankPresenter struct {
	MockGamePresenter[interfaces.MonteBankGame]
}

// HintOutput モック
func (_m *MockMonteBankPresenter) HintOutput(s interfaces.MonteBankGame) string {
	return _m.Called(s).Get(0).(string)
}
