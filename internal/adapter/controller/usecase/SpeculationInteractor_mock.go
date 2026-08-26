//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MockSpeculationInteractor スペキュレーションインタラクターモック
type MockSpeculationInteractor struct {
	mock.Mock
}

// **インタフェースを満たすことをここで固定する。** これが無いと、モックが
// 実装からずれても「誰もインタフェースに代入していない」あいだは黙って
// コンパイルが通り、最初に共有モックを使うテストを書いた人がパッケージごと
// 壊れているのを見つけることになる (実際そうなっていた: PlaceBet /
// ResetWithConfig が残り、Flip / Accept / Decline / Bid が無かった)。
var _ usecase.SpeculationInteractorIF = (*MockSpeculationInteractor)(nil)

func (m *MockSpeculationInteractor) Reset() string { return m.Called().String(0) }

func (m *MockSpeculationInteractor) Flip() string { return m.Called().String(0) }

func (m *MockSpeculationInteractor) Accept() string { return m.Called().String(0) }

func (m *MockSpeculationInteractor) Decline() string { return m.Called().String(0) }

func (m *MockSpeculationInteractor) Bid(amount int) string {
	return m.Called(amount).String(0)
}

func (m *MockSpeculationInteractor) NextRound() string { return m.Called().String(0) }

func (m *MockSpeculationInteractor) Hint() string { return m.Called().String(0) }

func (m *MockSpeculationInteractor) ActionLog() string { return m.Called().String(0) }

func (m *MockSpeculationInteractor) Snapshot() ([]byte, error) {
	args := m.Called()
	if v, ok := args.Get(0).([]byte); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}
