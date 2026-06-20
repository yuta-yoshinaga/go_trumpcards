//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockOpenFaceChinesePresenter オープンフェイス・チャイニーズポーカー (OFC) のプレゼンターモック
type MockOpenFaceChinesePresenter struct {
	MockGamePresenter[interfaces.OpenFaceChineseGame]
}

// HintOutput モック
func (_m *MockOpenFaceChinesePresenter) HintOutput(g interfaces.OpenFaceChineseGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
