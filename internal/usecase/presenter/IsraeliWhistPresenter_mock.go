//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockIsraeliWhistPresenter イスラエリホイストプレゼンターモック
type MockIsraeliWhistPresenter struct {
	MockGamePresenter[interfaces.IsraeliWhistGame]
}

// HintOutput モック
func (_m *MockIsraeliWhistPresenter) HintOutput(w interfaces.IsraeliWhistGame) string {
	return _m.Called(w).Get(0).(string)
}
