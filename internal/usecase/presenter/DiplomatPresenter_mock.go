//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockDiplomatPresenter ディプロマット プレゼンターモック
type MockDiplomatPresenter struct {
	MockGamePresenter[interfaces.DiplomatGame]
}

// HintOutput モック
func (_m *MockDiplomatPresenter) HintOutput(c interfaces.DiplomatGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
