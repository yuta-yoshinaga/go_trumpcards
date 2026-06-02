//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockOsmosisPresenter オズモシスプレゼンターモック
type MockOsmosisPresenter struct {
	MockGamePresenter[interfaces.OsmosisGame]
}

// HintOutput モック
func (_m *MockOsmosisPresenter) HintOutput(o interfaces.OsmosisGame) string {
	ret := _m.Called(o)
	return ret.Get(0).(string)
}
