//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockAlaskaPresenter アラスカプレゼンターモック
type MockAlaskaPresenter struct {
	mock.Mock
}

func (_m *MockAlaskaPresenter) Output(r interfaces.AlaskaGame, lastErr error) string {
	ret := _m.Called(r, lastErr)
	return ret.String(0)
}

func (_m *MockAlaskaPresenter) HintOutput(r interfaces.AlaskaGame) string {
	ret := _m.Called(r)
	return ret.String(0)
}

func (_m *MockAlaskaPresenter) ActionLogOutput(r interfaces.AlaskaGame) string {
	ret := _m.Called(r)
	return ret.String(0)
}
