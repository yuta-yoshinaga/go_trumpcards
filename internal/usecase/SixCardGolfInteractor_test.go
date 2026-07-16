//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

func TestSixCardGolfInteractor_Hint(t *testing.T) {
	game := domain.NewDefaultSixCardGolf()
	pMock := new(presenter.MockSixCardGolfPresenter)
	pMock.On("HintOutput", game).Return("hint")

	ci := usecase.NewSixCardGolfInteractor(game, pMock)
	assert.Equal(t, "hint", ci.Hint())
	pMock.AssertExpectations(t)
}
