//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSeahavenTowersPresenter シーヘイブンタワーズプレゼンターモック
type MockSeahavenTowersPresenter struct {
	MockGamePresenter[interfaces.SeahavenTowersGame]
}

// HintOutput モック
func (_m *MockSeahavenTowersPresenter) HintOutput(s interfaces.SeahavenTowersGame) string {
	ret := _m.Called(s)
	return ret.Get(0).(string)
}
