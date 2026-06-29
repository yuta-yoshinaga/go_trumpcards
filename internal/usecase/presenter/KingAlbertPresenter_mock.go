//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockKingAlbertPresenter King Albert プレゼンターモック
type MockKingAlbertPresenter struct {
	MockGamePresenter[interfaces.KingAlbertGame]
}

// HintOutput モック
func (_m *MockKingAlbertPresenter) HintOutput(bc interfaces.KingAlbertGame) string {
	ret := _m.Called(bc)
	return ret.Get(0).(string)
}
