//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockKempsPresenter はケムプスのプレゼンターモック。
type MockKempsPresenter = MockGamePresenter[interfaces.KempsGame]
