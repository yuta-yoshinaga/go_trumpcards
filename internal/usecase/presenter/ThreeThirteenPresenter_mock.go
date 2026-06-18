//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockThreeThirteenPresenter スリー・サーティーンプレゼンターモック
type MockThreeThirteenPresenter = MockGamePresenter[interfaces.ThreeThirteenGame]
