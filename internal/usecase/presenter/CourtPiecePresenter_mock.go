//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCourtPiecePresenter Court Piece プレゼンターモック
type MockCourtPiecePresenter struct {
	MockGamePresenter[interfaces.CourtPieceGame]
}

// HintOutput モック
func (_m *MockCourtPiecePresenter) HintOutput(t interfaces.CourtPieceGame) string {
	ret := _m.Called(t)
	return ret.Get(0).(string)
}
