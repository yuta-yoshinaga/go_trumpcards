//go:build test

package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BigTwo has no generated presenter mock, so stub the two methods the way
// interactor_snapshot_test.go does.
type stubBigTwoPresenter struct{ calls int }

func (s *stubBigTwoPresenter) Output(_ interfaces.BigTwoGame, _ error) string {
	s.calls++
	return `{}`
}
func (s *stubBigTwoPresenter) ActionLogOutput(_ interfaces.BigTwoGame) string { return `{}` }

// runCpuTurns was the only function in this interactor with no test at all
// (0% on develop). #5418 put an iteration cap in it, so drive it once.
func TestBigTwoInteractor_ResetRunsCpuTurns(t *testing.T) {
	game := domain.NewDefaultBigTwo()
	p := &stubBigTwoPresenter{}

	bi := usecase.NewBigTwoInteractor(game, p)
	assert.Equal(t, `{}`, bi.Reset())
	assert.Equal(t, 1, p.calls)

	// **配った直後は必ず人間の手番か終局。** どちらでもないなら CPU ループが
	// 上限まで空回りして返ってきたということで、それがこの上限の直す対象。
	assert.True(t, game.IsHumanTurn() || game.GetGameEndFlag(),
		"Reset 後に人間の手番でも終局でもない -- CPU ループが進まないまま抜けている")
}
