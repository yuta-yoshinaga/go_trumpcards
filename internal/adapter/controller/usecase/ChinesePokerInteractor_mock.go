//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockChinesePokerInteractor チャイニーズポーカーインタラクターモック
type MockChinesePokerInteractor struct {
	mock.Mock
}

// Reset ゲーム初期化
func (m *MockChinesePokerInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

// Bet ベット
func (m *MockChinesePokerInteractor) Bet(amount int) string {
	args := m.Called(amount)
	return args.String(0)
}

// SetHands ハンド設定
func (m *MockChinesePokerInteractor) SetHands(frontIndices []int, middleIndices []int) string {
	args := m.Called(frontIndices, middleIndices)
	return args.String(0)
}

// ActionLog 棋譜出力
func (m *MockChinesePokerInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockChinesePokerInteractor) Hint() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot スナップショット
func (m *MockChinesePokerInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
