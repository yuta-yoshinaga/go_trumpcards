//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockLiteraturePresenter リテラチャー (Literature) プレゼンターモック
type MockLiteraturePresenter = MockGamePresenter[interfaces.LiteratureGame]
