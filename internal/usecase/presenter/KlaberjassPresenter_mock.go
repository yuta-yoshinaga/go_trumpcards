//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockKlaberjassPresenter クラバーヤス (Klaberjass) プレゼンターモック
type MockKlaberjassPresenter = MockGamePresenter[interfaces.KlaberjassGame]
