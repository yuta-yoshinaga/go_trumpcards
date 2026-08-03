//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMissMilliganPresenter ミス・ミリガン プレゼンターモック
type MockMissMilliganPresenter struct {
	MockGamePresenter[interfaces.MissMilliganGame]
}

// HintOutput モック
func (_m *MockMissMilliganPresenter) HintOutput(mm interfaces.MissMilliganGame) string {
	ret := _m.Called(mm)
	return ret.Get(0).(string)
}
