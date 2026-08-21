//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockRankAndFilePresenter ランク・アンド・ファイルプレゼンターモック
type MockRankAndFilePresenter struct {
	MockGamePresenter[interfaces.RankAndFileGame]
}

// HintOutput モック
func (_m *MockRankAndFilePresenter) HintOutput(ft interfaces.RankAndFileGame) string {
	ret := _m.Called(ft)
	return ret.Get(0).(string)
}
