//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCucumberPresenter キューカンバープレゼンターモック
type MockCucumberPresenter struct {
	MockGamePresenter[interfaces.CucumberGame]
}

// HintOutput モック
func (_m *MockCucumberPresenter) HintOutput(s interfaces.CucumberGame) string {
	return _m.Called(s).Get(0).(string)
}
