package domain_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestPlayer_Method(t *testing.T) {
	tp := new(domain.Player)
	t.Run("success AddCard", func(t *testing.T) {
		tp.AddCard(domain.NewCard(domain.CardDesignJoker, domain.CardValueJoker, false))
		assert.Equal(t, 1, tp.GetCardsSize())
	})
	t.Run("success GetCard", func(t *testing.T) {
		card := tp.GetCard(0)
		assert.Equal(t, domain.CardDesignJoker, card.GetDesign())
		assert.Equal(t, domain.CardValueJoker, card.GetValue())
		assert.Equal(t, false, card.GetDraw())
	})
	t.Run("failed GetCard idx -1", func(t *testing.T) {
		card := tp.GetCard(-1)
		assert.Empty(t, card)
	})
	t.Run("failed GetCard idx 1", func(t *testing.T) {
		card := tp.GetCard(1)
		assert.Empty(t, card)
	})
	t.Run("success Reset", func(t *testing.T) {
		tp.Reset()
		assert.Equal(t, 0, tp.GetCardsSize())
	})
	t.Run("success ShuffleCards does not change size", func(t *testing.T) {
		p := new(domain.Player)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		p.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		p.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		p.ShuffleCards()
		assert.Equal(t, 3, p.GetCardsSize())
	})
	t.Run("success ShuffleCards randomizes order", func(t *testing.T) {
		p := new(domain.Player)
		for i := 2; i <= 10; i++ {
			p.AddCard(domain.NewCard(domain.CardDesignSpade, i, false))
		}
		original := make([]int, p.GetCardsSize())
		for i := range original {
			original[i] = p.GetCard(i).GetValue()
		}
		// 多数回シャッフルして順番が変わることを確認
		changed := false
		for attempt := 0; attempt < 100; attempt++ {
			p2 := new(domain.Player)
			for _, v := range original {
				p2.AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
			}
			p2.ShuffleCards()
			for i := range original {
				if p2.GetCard(i).GetValue() != original[i] {
					changed = true
					break
				}
			}
			if changed {
				break
			}
		}
		assert.True(t, changed, "ShuffleCards should change card order")
	})
	t.Run("success ShuffleCards on empty hand", func(t *testing.T) {
		p := new(domain.Player)
		p.ShuffleCards()
		assert.Equal(t, 0, p.GetCardsSize())
	})
	t.Run("success ShuffleCards on single card hand", func(t *testing.T) {
		p := new(domain.Player)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		p.ShuffleCards()
		assert.Equal(t, 1, p.GetCardsSize())
		assert.Equal(t, 2, p.GetCard(0).GetValue())
	})
}

func TestPlayer_ReorderCards(t *testing.T) {
	t.Run("success valid permutation", func(t *testing.T) {
		p := new(domain.Player)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		p.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		p.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		err := p.ReorderCards([]int{2, 0, 1})
		assert.NoError(t, err)
		assert.Equal(t, 7, p.GetCard(0).GetValue())
		assert.Equal(t, 2, p.GetCard(1).GetValue())
		assert.Equal(t, 5, p.GetCard(2).GetValue())
	})
	t.Run("success identity permutation", func(t *testing.T) {
		p := new(domain.Player)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		p.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		err := p.ReorderCards([]int{0, 1})
		assert.NoError(t, err)
		assert.Equal(t, 2, p.GetCard(0).GetValue())
		assert.Equal(t, 5, p.GetCard(1).GetValue())
	})
	t.Run("success empty hand", func(t *testing.T) {
		p := new(domain.Player)
		err := p.ReorderCards([]int{})
		assert.NoError(t, err)
	})
	t.Run("error length mismatch", func(t *testing.T) {
		p := new(domain.Player)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		p.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		err := p.ReorderCards([]int{0})
		assert.ErrorIs(t, err, domain.ErrInvalidIndices)
	})
	t.Run("error duplicate indices", func(t *testing.T) {
		p := new(domain.Player)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		p.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		err := p.ReorderCards([]int{0, 0})
		assert.ErrorIs(t, err, domain.ErrInvalidIndices)
	})
	t.Run("error out of range positive", func(t *testing.T) {
		p := new(domain.Player)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		p.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		err := p.ReorderCards([]int{0, 5})
		assert.ErrorIs(t, err, domain.ErrInvalidIndices)
	})
	t.Run("error out of range negative", func(t *testing.T) {
		p := new(domain.Player)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		p.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		err := p.ReorderCards([]int{-1, 0})
		assert.ErrorIs(t, err, domain.ErrInvalidIndices)
	})
}

func TestPlayer_RemoveCard(t *testing.T) {
	t.Run("valid first index", func(t *testing.T) {
		p := new(domain.Player)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		p.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		p.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		card := p.RemoveCard(0)
		assert.NotNil(t, card)
		assert.Equal(t, 2, card.GetValue())
		assert.Equal(t, 2, p.GetCardsSize())
		assert.Equal(t, 5, p.GetCard(0).GetValue())
	})

	t.Run("valid middle index", func(t *testing.T) {
		p := new(domain.Player)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		p.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		p.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
		card := p.RemoveCard(1)
		assert.NotNil(t, card)
		assert.Equal(t, 5, card.GetValue())
		assert.Equal(t, 2, p.GetCardsSize())
	})

	t.Run("valid last index", func(t *testing.T) {
		p := new(domain.Player)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		p.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
		card := p.RemoveCard(1)
		assert.NotNil(t, card)
		assert.Equal(t, 5, card.GetValue())
		assert.Equal(t, 1, p.GetCardsSize())
	})

	t.Run("negative index returns nil", func(t *testing.T) {
		p := new(domain.Player)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		card := p.RemoveCard(-1)
		assert.Nil(t, card)
		assert.Equal(t, 1, p.GetCardsSize())
	})

	t.Run("out-of-bounds index returns nil", func(t *testing.T) {
		p := new(domain.Player)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		card := p.RemoveCard(5)
		assert.Nil(t, card)
		assert.Equal(t, 1, p.GetCardsSize())
	})

	t.Run("empty hand returns nil", func(t *testing.T) {
		p := new(domain.Player)
		card := p.RemoveCard(0)
		assert.Nil(t, card)
		assert.Equal(t, 0, p.GetCardsSize())
	})
}

func TestPlayer_RemoveCards(t *testing.T) {
	makePlayer := func() *domain.Player {
		p := new(domain.Player)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false)) // 0
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false)) // 1
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false)) // 2
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 4, false)) // 3
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // 4
		return p
	}

	t.Run("single index removal", func(t *testing.T) {
		p := makePlayer()
		removed := p.RemoveCards([]int{2})
		assert.Len(t, removed, 1)
		assert.Equal(t, 3, removed[0].GetValue())
		assert.Equal(t, 4, p.GetCardsSize())
	})

	t.Run("multiple indices removal", func(t *testing.T) {
		p := makePlayer()
		removed := p.RemoveCards([]int{0, 2, 4})
		assert.Len(t, removed, 3)
		assert.Equal(t, 1, removed[0].GetValue())
		assert.Equal(t, 3, removed[1].GetValue())
		assert.Equal(t, 5, removed[2].GetValue())
		assert.Equal(t, 2, p.GetCardsSize())
	})

	t.Run("deduplicates indices", func(t *testing.T) {
		p := makePlayer()
		removed := p.RemoveCards([]int{1, 1, 3})
		assert.Len(t, removed, 2)
		assert.Equal(t, 2, removed[0].GetValue())
		assert.Equal(t, 4, removed[1].GetValue())
		assert.Equal(t, 3, p.GetCardsSize())
	})

	t.Run("unordered indices handled correctly", func(t *testing.T) {
		p := makePlayer()
		removed := p.RemoveCards([]int{2, 0})
		assert.Len(t, removed, 2)
		assert.Equal(t, 1, removed[0].GetValue())
		assert.Equal(t, 3, removed[1].GetValue())
		assert.Equal(t, 3, p.GetCardsSize())
		assert.Equal(t, 2, p.GetCard(0).GetValue())
	})

	t.Run("out-of-bounds index ignored", func(t *testing.T) {
		p := makePlayer()
		removed := p.RemoveCards([]int{0, 99})
		assert.Len(t, removed, 1)
		assert.Equal(t, 1, removed[0].GetValue())
		assert.Equal(t, 4, p.GetCardsSize())
	})

	t.Run("negative index ignored", func(t *testing.T) {
		p := makePlayer()
		removed := p.RemoveCards([]int{-1, 0})
		assert.Len(t, removed, 1)
		assert.Equal(t, 1, removed[0].GetValue())
		assert.Equal(t, 4, p.GetCardsSize())
	})

	t.Run("empty indices returns empty slice", func(t *testing.T) {
		p := makePlayer()
		removed := p.RemoveCards([]int{})
		assert.Empty(t, removed)
		assert.Equal(t, 5, p.GetCardsSize())
	})

	t.Run("remove all cards", func(t *testing.T) {
		p := makePlayer()
		removed := p.RemoveCards([]int{0, 1, 2, 3, 4})
		assert.Len(t, removed, 5)
		assert.Equal(t, 0, p.GetCardsSize())
	})
}

func TestPlayer_PrependCard(t *testing.T) {
	p := new(domain.Player)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
	p.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))
	front := domain.NewCard(domain.CardDesignHeart, 1, false)
	p.PrependCard(front)
	assert.Equal(t, 3, p.GetCardsSize())
	assert.Equal(t, domain.CardDesignHeart, p.GetCard(0).GetDesign())
	assert.Equal(t, 1, p.GetCard(0).GetValue())
	assert.Equal(t, domain.CardDesignSpade, p.GetCard(1).GetDesign())
	assert.Equal(t, domain.CardDesignClover, p.GetCard(2).GetDesign())
}

func TestPlayer_InsertCard(t *testing.T) {
	makePlayer := func() *domain.Player {
		p := new(domain.Player)
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))
		return p
	}

	t.Run("insert at beginning (pos=0)", func(t *testing.T) {
		p := makePlayer()
		p.InsertCard(domain.NewCard(domain.CardDesignHeart, 10, false), 0)
		assert.Equal(t, 4, p.GetCardsSize())
		assert.Equal(t, 10, p.GetCard(0).GetValue())
		assert.Equal(t, 1, p.GetCard(1).GetValue())
	})

	t.Run("insert at negative pos → prepend", func(t *testing.T) {
		p := makePlayer()
		p.InsertCard(domain.NewCard(domain.CardDesignHeart, 10, false), -1)
		assert.Equal(t, 4, p.GetCardsSize())
		assert.Equal(t, 10, p.GetCard(0).GetValue())
	})

	t.Run("insert at end (pos >= size)", func(t *testing.T) {
		p := makePlayer()
		p.InsertCard(domain.NewCard(domain.CardDesignHeart, 10, false), 5)
		assert.Equal(t, 4, p.GetCardsSize())
		assert.Equal(t, 10, p.GetCard(3).GetValue())
	})

	t.Run("insert in middle", func(t *testing.T) {
		p := makePlayer()
		p.InsertCard(domain.NewCard(domain.CardDesignHeart, 10, false), 1)
		assert.Equal(t, 4, p.GetCardsSize())
		assert.Equal(t, 1, p.GetCard(0).GetValue())
		assert.Equal(t, 10, p.GetCard(1).GetValue())
		assert.Equal(t, 2, p.GetCard(2).GetValue())
		assert.Equal(t, 3, p.GetCard(3).GetValue())
	})
}
