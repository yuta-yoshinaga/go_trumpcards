//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCribbageSquaresPresenter はクリベッジ・スクエアズのプレゼンターモック。
type MockCribbageSquaresPresenter struct {
	MockGamePresenter[interfaces.CribbageSquaresGame]
}

// HintOutput はヒント出力のモック。
func (_m *MockCribbageSquaresPresenter) HintOutput(p interfaces.CribbageSquaresGame) string {
	ret := _m.Called(p)
	return ret.String(0)
}
