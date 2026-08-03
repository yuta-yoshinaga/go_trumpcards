//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockKarnoffelPresenter カルニッフェル (Karnöffel) プレゼンターモック
type MockKarnoffelPresenter = MockGamePresenter[interfaces.KarnoffelGame]
