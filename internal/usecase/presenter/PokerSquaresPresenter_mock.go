//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPokerSquaresPresenter はポーカー・スクエアズのプレゼンターモック。
type MockPokerSquaresPresenter struct {
	MockGamePresenter[interfaces.PokerSquaresGame]
}
