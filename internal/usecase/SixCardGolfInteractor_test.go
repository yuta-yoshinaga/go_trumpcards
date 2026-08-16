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

// runCpuTurns はこのファイルで唯一テストの無い関数だった (develop 上で 0%)。
// #5418 で反復上限を入れたので、その分岐も含めて一度通しておく。
func TestSixCardGolfInteractor_ResetRunsCpuTurns(t *testing.T) {
	game := domain.NewDefaultSixCardGolf()
	pMock := new(presenter.MockSixCardGolfPresenter)
	pMock.On("Output", game, nil).Return("board")

	ci := usecase.NewSixCardGolfInteractor(game, pMock)
	assert.Equal(t, "board", ci.Reset())

	// **配った直後は人間の手番。** ここで CPU が動き続けていたら、上限に当たるまで
	// 回ってから返ってくる = セットアップが壊れている。
	assert.True(t, game.IsHumanTurn(), "Reset 後に人間の手番になっていない")
	assert.False(t, game.GetGameEndFlag())
	pMock.AssertExpectations(t)
}
