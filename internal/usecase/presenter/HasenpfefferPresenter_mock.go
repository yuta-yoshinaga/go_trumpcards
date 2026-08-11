//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockHasenpfefferPresenter ハーゼンプフェファープレゼンターモック
type MockHasenpfefferPresenter struct {
	MockGamePresenter[interfaces.HasenpfefferGame]
}

// HintOutput モック
func (_m *MockHasenpfefferPresenter) HintOutput(h interfaces.HasenpfefferGame) string {
	return _m.Called(h).Get(0).(string)
}
