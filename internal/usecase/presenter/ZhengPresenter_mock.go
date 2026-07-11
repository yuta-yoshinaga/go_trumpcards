//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockZhengPresenter 争上游プレゼンターモック
type MockZhengPresenter = MockGamePresenter[interfaces.ZhengGame]
