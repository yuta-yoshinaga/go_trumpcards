//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockFollowTheQueenPresenter フォロー・ザ・クイーンプレゼンターモック
type MockFollowTheQueenPresenter struct {
	MockGamePresenter[interfaces.FollowTheQueenGame]
}

// HintOutput モック
func (_m *MockFollowTheQueenPresenter) HintOutput(g interfaces.FollowTheQueenGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
