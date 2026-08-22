//go:build test && (!js || !wasm || extra4)

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPutPresenter プットプレゼンターモック
type MockPutPresenter struct {
	MockGamePresenter[interfaces.PutGame]
}

// HintOutput モック
func (_m *MockPutPresenter) HintOutput(t interfaces.PutGame) string {
	ret := _m.Called(t)
	return ret.Get(0).(string)
}
