package entities_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/entities"

	"github.com/stretchr/testify/assert"
)

func TestPoker_Method(t *testing.T) {
	tc := entities.NewTrumpCards(0)
	player := entities.NewPokerPlayer()
	dealer := entities.NewPokerPlayer()
	tp := entities.NewPoker(tc, player, dealer)

	t.Run("success Reset", func(t *testing.T) {
		tp.Reset()
		assert.Equal(t, entities.PokerPhaseDeal, tp.GetPhase())
		assert.Equal(t, 5, tp.GetPlayer().GetCardsSize())
		assert.Equal(t, 5, tp.GetDealer().GetCardsSize())
	})

	t.Run("success GetPlayer", func(t *testing.T) {
		assert.NotEmpty(t, tp.GetPlayer())
	})

	t.Run("success GetDealer", func(t *testing.T) {
		assert.NotEmpty(t, tp.GetDealer())
	})

	t.Run("success PlayerStand moves to End phase", func(t *testing.T) {
		tp.Reset()
		tp.PlayerStand()
		assert.Equal(t, entities.PokerPhaseEnd, tp.GetPhase())
	})

	t.Run("success PlayerExchange moves to End phase", func(t *testing.T) {
		tp.Reset()
		tp.PlayerExchange([]int{0, 1})
		assert.Equal(t, entities.PokerPhaseEnd, tp.GetPhase())
	})

	t.Run("success PlayerExchange ignored when not in Deal phase", func(t *testing.T) {
		tp.Reset()
		tp.PlayerStand()
		assert.Equal(t, entities.PokerPhaseEnd, tp.GetPhase())
		tp.PlayerExchange([]int{0})
		assert.Equal(t, entities.PokerPhaseEnd, tp.GetPhase())
	})

	t.Run("success PlayerStand ignored when not in Deal phase", func(t *testing.T) {
		tp.Reset()
		tp.PlayerStand()
		assert.Equal(t, entities.PokerPhaseEnd, tp.GetPhase())
		tp.PlayerStand()
		assert.Equal(t, entities.PokerPhaseEnd, tp.GetPhase())
	})

	t.Run("success GameJudgment player win higher rank", func(t *testing.T) {
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// player: Royal Flush
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 10, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 12, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 13, false))
		// dealer: High Card
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 7, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 11, false))
		player.EvalHand()
		dealer.EvalHand()
		assert.Equal(t, 1, tp.GameJudgment())
	})

	t.Run("success GameJudgment player lose lower rank", func(t *testing.T) {
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// player: High Card
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		player.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignHeart, 7, false))
		player.AddCard(entities.NewCard(entities.CardDesignDiamond, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		// dealer: One Pair
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignDiamond, 7, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignHeart, 11, false))
		player.EvalHand()
		dealer.EvalHand()
		assert.Equal(t, -1, tp.GameJudgment())
	})

	t.Run("success GameJudgment draw same rank same high cards", func(t *testing.T) {
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// both: High Card with same values
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 7, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
		player.EvalHand()
		dealer.EvalHand()
		assert.Equal(t, 0, tp.GameJudgment())
	})

	t.Run("success GameJudgment player win same rank higher cards", func(t *testing.T) {
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// both: High Card, player has higher top card
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 13, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 7, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
		player.EvalHand()
		dealer.EvalHand()
		assert.Equal(t, 1, tp.GameJudgment())
	})

	t.Run("success GameJudgment player lose same rank lower cards", func(t *testing.T) {
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// both: High Card, player has lower top card
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 2, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 2, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 7, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 13, false))
		player.EvalHand()
		dealer.EvalHand()
		assert.Equal(t, -1, tp.GameJudgment())
	})

	t.Run("success GameJudgment ace treated as high card", func(t *testing.T) {
		tp.Reset()
		player.Reset()
		dealer.Reset()
		// player has ace (high), dealer has king
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 1, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 5, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 7, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 9, false))
		player.AddCard(entities.NewCard(entities.CardDesignSpade, 11, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 13, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 5, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 7, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 9, false))
		dealer.AddCard(entities.NewCard(entities.CardDesignClover, 11, false))
		player.EvalHand()
		dealer.EvalHand()
		assert.Equal(t, 1, tp.GameJudgment())
	})
}
