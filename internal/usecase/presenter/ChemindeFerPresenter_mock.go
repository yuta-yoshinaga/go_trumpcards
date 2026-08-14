//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockChemindeFerPresenter シュマン・ド・フェールプレゼンターモック
type MockChemindeFerPresenter struct {
	MockGamePresenter[interfaces.ChemindeFerGame]
}

// HintOutput モック
func (_m *MockChemindeFerPresenter) HintOutput(s interfaces.ChemindeFerGame) string {
	return _m.Called(s).Get(0).(string)
}
