//go:build test

package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewHeartsPlayer(t *testing.T) {
	t.Run("人間プレイヤーの初期状態", func(t *testing.T) {
		p := domain.NewHeartsPlayer(true)
		assert.True(t, p.GetIsHuman())
		assert.False(t, p.GetIsFinished())
		assert.Equal(t, 0, p.GetCardsSize())
		assert.Equal(t, 0, p.GetRoundScore())
		assert.Equal(t, 0, p.GetCumulativeScore())
		assert.Nil(t, p.GetTricksTaken())
		assert.Equal(t, 0, p.GetTrickCount())
	})

	t.Run("CPUプレイヤーの初期状態", func(t *testing.T) {
		p := domain.NewHeartsPlayer(false)
		assert.False(t, p.GetIsHuman())
		assert.False(t, p.GetIsFinished())
		assert.Equal(t, 0, p.GetRoundScore())
		assert.Equal(t, 0, p.GetCumulativeScore())
	})
}

func TestHeartsPlayer_RoundScore(t *testing.T) {
	t.Run("ラウンドスコアの設定と取得", func(t *testing.T) {
		p := domain.NewHeartsPlayer(false)
		p.SetRoundScore(13)
		assert.Equal(t, 13, p.GetRoundScore())
	})
}

func TestHeartsPlayer_CumulativeScore(t *testing.T) {
	t.Run("累積スコアの設定と取得", func(t *testing.T) {
		p := domain.NewHeartsPlayer(false)
		p.SetCumulativeScore(26)
		assert.Equal(t, 26, p.GetCumulativeScore())
	})
}

func TestHeartsPlayer_AddTrick(t *testing.T) {
	t.Run("トリックを追加する", func(t *testing.T) {
		p := domain.NewHeartsPlayer(false)
		trick1 := []*domain.Card{
			domain.NewCard(domain.CardDesignHeart, 5, false),
			domain.NewCard(domain.CardDesignSpade, 3, false),
		}
		trick2 := []*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 10, false),
			domain.NewCard(domain.CardDesignClover, 7, false),
		}

		p.AddTrick(trick1)
		assert.Equal(t, 1, p.GetTrickCount())
		assert.Len(t, p.GetTricksTaken(), 1)

		p.AddTrick(trick2)
		assert.Equal(t, 2, p.GetTrickCount())
		assert.Len(t, p.GetTricksTaken(), 2)

		// トリックの中身を確認
		assert.Equal(t, trick1, p.GetTricksTaken()[0])
		assert.Equal(t, trick2, p.GetTricksTaken()[1])
	})
}

func TestHeartsPlayer_CommitRoundScore(t *testing.T) {
	t.Run("ラウンドスコアを累積スコアに加算する", func(t *testing.T) {
		p := domain.NewHeartsPlayer(false)

		p.SetRoundScore(10)
		p.CommitRoundScore()
		assert.Equal(t, 10, p.GetCumulativeScore())

		// 再度コミットすると加算される
		p.SetRoundScore(5)
		p.CommitRoundScore()
		assert.Equal(t, 15, p.GetCumulativeScore())
	})
}

func TestHeartsPlayer_ResetRound(t *testing.T) {
	t.Run("ラウンドリセットで全状態が初期化される", func(t *testing.T) {
		p := domain.NewHeartsPlayer(true)

		// 状態を設定
		p.SetRoundScore(13)
		p.SetCumulativeScore(50)
		p.SetIsFinished(true)
		p.AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
		p.AddTrick([]*domain.Card{
			domain.NewCard(domain.CardDesignDiamond, 3, false),
		})

		p.ResetRound()

		assert.Equal(t, 0, p.GetRoundScore())
		// 累積スコアはリセットされない
		assert.Equal(t, 50, p.GetCumulativeScore())
		assert.Equal(t, 0, p.GetCardsSize())
		assert.Nil(t, p.GetTricksTaken())
		assert.Equal(t, 0, p.GetTrickCount())
		assert.False(t, p.GetIsFinished())
	})
}
