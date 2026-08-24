//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGermanSoloPresenter ジャーマン・ソロ (GermanSolo) のプレゼンターモック
type MockGermanSoloPresenter struct {
	MockGamePresenter[interfaces.GermanSoloGame]
}

// HintOutput モック
func (_m *MockGermanSoloPresenter) HintOutput(g interfaces.GermanSoloGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
