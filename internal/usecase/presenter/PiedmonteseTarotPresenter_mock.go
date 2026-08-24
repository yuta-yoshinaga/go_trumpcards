//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPiedmonteseTarotPresenter はピエモンテ・タロッコのプレゼンターモック。
type MockPiedmonteseTarotPresenter struct {
	MockGamePresenter[interfaces.PiedmonteseTarotGame]
}

// HintOutput モック
func (_m *MockPiedmonteseTarotPresenter) HintOutput(g interfaces.PiedmonteseTarotGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
