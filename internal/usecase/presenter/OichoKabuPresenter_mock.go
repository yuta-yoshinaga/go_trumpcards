//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockOichoKabuPresenter おいちょかぶプレゼンターモック
type MockOichoKabuPresenter = MockGamePresenter[interfaces.OichoKabuGame]
