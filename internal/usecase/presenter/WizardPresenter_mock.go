//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockWizardPresenter ウィザードプレゼンターモック
type MockWizardPresenter struct {
	MockGamePresenter[interfaces.WizardGame]
}

// HintOutput モック
func (_m *MockWizardPresenter) HintOutput(o interfaces.WizardGame) string {
	ret := _m.Called(o)
	return ret.Get(0).(string)
}
