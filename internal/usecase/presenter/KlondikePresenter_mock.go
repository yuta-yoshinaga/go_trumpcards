package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockKlondikePresenter クロンダイクプレゼンターモック
type MockKlondikePresenter struct {
	MockGamePresenter[interfaces.KlondikeGame]
}

// HintOutput モック
func (_m *MockKlondikePresenter) HintOutput(k interfaces.KlondikeGame) string {
	ret := _m.Called(k)
	return ret.Get(0).(string)
}
