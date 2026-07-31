//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKlaberjassInteractor モック
type MockKlaberjassInteractor struct {
	mock.Mock
}

func (_m *MockKlaberjassInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockKlaberjassInteractor) ResetWithConfig(cfg domain.KlaberjassConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockKlaberjassInteractor) AcceptTrump() string {
	return _m.Called().String(0)
}

func (_m *MockKlaberjassInteractor) CallTrump(suit int) string {
	return _m.Called(suit).String(0)
}

func (_m *MockKlaberjassInteractor) Pass() string {
	return _m.Called().String(0)
}

func (_m *MockKlaberjassInteractor) Schmeiss() string {
	return _m.Called().String(0)
}

func (_m *MockKlaberjassInteractor) AnswerSchmeiss(accept bool) string {
	return _m.Called(accept).String(0)
}

func (_m *MockKlaberjassInteractor) PlayCard(idx int) string {
	return _m.Called(idx).String(0)
}

func (_m *MockKlaberjassInteractor) NextDeal() string {
	return _m.Called().String(0)
}

func (_m *MockKlaberjassInteractor) GetConfig() domain.KlaberjassConfig {
	return _m.Called().Get(0).(domain.KlaberjassConfig)
}

func (_m *MockKlaberjassInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockKlaberjassInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
