//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKaiserInteractor モック
type MockKaiserInteractor struct {
	mock.Mock
}

func (_m *MockKaiserInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockKaiserInteractor) ResetWithConfig(cfg domain.KaiserConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockKaiserInteractor) Bid(value int, contract domain.KaiserContract) string {
	return _m.Called(value, contract).String(0)
}

func (_m *MockKaiserInteractor) PassBid() string {
	return _m.Called().String(0)
}

func (_m *MockKaiserInteractor) SetTrump(suit int) string {
	return _m.Called(suit).String(0)
}

func (_m *MockKaiserInteractor) Discard(idxs []int) string {
	return _m.Called(idxs).String(0)
}

func (_m *MockKaiserInteractor) PlayCard(idx int) string {
	return _m.Called(idx).String(0)
}

func (_m *MockKaiserInteractor) NextHand() string {
	return _m.Called().String(0)
}

func (_m *MockKaiserInteractor) GetConfig() domain.KaiserConfig {
	return _m.Called().Get(0).(domain.KaiserConfig)
}

func (_m *MockKaiserInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Hint モック
func (_m *MockKaiserInteractor) Hint() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockKaiserInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
