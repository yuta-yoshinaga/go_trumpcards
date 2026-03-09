package presenter_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

	"github.com/stretchr/testify/assert"
)

func TestBlackJackWebPresenters_Method(t *testing.T) {
	tbp := presenter.NewBlackJackWebPresenter()

	t.Run("success Output bet phase (no cards)", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.BJPhaseBet, result.Phase)
		assert.Equal(t, domain.BJDefaultChips, result.Player.Chips)
		assert.Equal(t, domain.BJDefaultChips, result.Dealer.Chips)
		assert.Equal(t, 0, len(result.Dealer.Cards))
		assert.Equal(t, 1, len(result.Hands))
		assert.Equal(t, 0, result.Hands[0].Score)
	})
	t.Run("success Output action phase", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 11, false))
		bj.SetPhase(domain.BJPhaseAction)
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, 0, result.Dealer.Score) // Not game end, so score hidden
		assert.Equal(t, 1, len(result.Dealer.Cards))
		assert.Equal(t, 900, result.Player.Chips)
		assert.Equal(t, 1000, result.Dealer.Chips)
		assert.Equal(t, domain.BJPhaseAction, result.Phase)
		assert.Equal(t, 1, len(result.Hands))
		assert.Equal(t, 100, result.Hands[0].Bet)
	})
	t.Run("success Output end phase lose", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 2, false))
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 2, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 11, false))
		bj.SetPhase(domain.BJPhaseAction)
		_ = bj.PlayerStand()
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, 22, result.Dealer.Score)
		assert.Equal(t, 3, len(result.Dealer.Cards))
		assert.Equal(t, "It is your loss.", result.Message)
		assert.Equal(t, "blackjack.result.lose", result.MessageCode)
		assert.Equal(t, domain.BJPhaseEnd, result.Phase)
	})
	t.Run("success Output end phase draw", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		bj.SetPhase(domain.BJPhaseAction)
		_ = bj.PlayerStand()
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, "It is a draw.", result.Message)
		assert.Equal(t, "blackjack.result.draw", result.MessageCode)
	})
	t.Run("success Output end phase win", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		// Dealer score 19 (>= 17)
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		bj.SetPhase(domain.BJPhaseAction)
		_ = bj.PlayerStand()
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, "You are the winner.", result.Message)
		assert.Equal(t, "blackjack.result.win", result.MessageCode)
	})
	t.Run("success Output hands fields", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.SetDoubled(true)
		hand.SetBusted(true)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		hand.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		bj.SetPhase(domain.BJPhaseAction)
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result.Hands))
		assert.True(t, result.Hands[0].Doubled)
		assert.True(t, result.Hands[0].Busted)
		assert.Equal(t, 100, result.Hands[0].Bet)
		assert.Equal(t, 3, len(result.Hands[0].Cards))
	})
	t.Run("success Output insurance fields", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false))
		bj.SetPhase(domain.BJPhaseInsurance)
		_ = bj.PlayerInsurance() // cost = 50
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, 50, result.InsuranceBet)
	})
	t.Run("success Output canSplit and isBlackJack", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(900)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		bj.SetPhase(domain.BJPhaseAction)
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result.Hands[0].CanSplit)
		assert.False(t, result.Hands[0].IsBlackJack)
	})
	t.Run("success Output error message via lastErr", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		testErr := errors.New("Invalid bet amount.")
		output := tbp.Output(bj, testErr)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.BJPhaseBet, result.Phase)
		assert.Equal(t, "Invalid bet amount.", result.Message)
	})
	t.Run("success Output nil error has no error message in bet phase", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, "", result.Message)
	})
}

func TestBlackJackWebPresenter_ConfigFields(t *testing.T) {
	tbp := presenter.NewBlackJackWebPresenter()

	t.Run("success Output includes dealerHitsSoft17 true", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		_ = bj.SetConfig(domain.BlackJackConfig{DealerHitsSoft17: true, CpuPlayerCount: 0, CountingEnabled: false})
		bj.Reset()
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result.DealerHitsSoft17)
		assert.False(t, result.CountingEnabled)
		assert.Equal(t, 0, result.CpuPlayerCount)
	})
	t.Run("success Output includes countingEnabled and counts", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		_ = bj.SetConfig(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: true})
		bj.Reset()
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.True(t, result.CountingEnabled)
		assert.Equal(t, 0, result.RunningCount)
		assert.Equal(t, 0.0, result.TrueCount)
		assert.Equal(t, domain.BJCountingHiLo, result.CountingSystem)
	})
	t.Run("success Output includes countingSystem KO", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		_ = bj.SetConfig(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 0, CountingEnabled: true, CountingSystem: domain.BJCountingKO})
		bj.Reset()
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.BJCountingKO, result.CountingSystem)
	})
	t.Run("success Output includes cpuPlayerCount", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		_ = bj.SetConfig(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 2, CountingEnabled: false})
		bj.Reset()
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, 2, result.CpuPlayerCount)
	})
	t.Run("success Output CPU players serialization in action phase", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		_ = bj.SetConfig(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 1, CountingEnabled: false})
		bj.Reset()
		// Bet and move to action phase
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		bj.SetPhase(domain.BJPhaseAction)
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result.CpuPlayers))
		assert.True(t, result.CpuPlayers[0].Chips > 0)
	})
	t.Run("success Output CPU players serialization in end phase", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		_ = bj.SetConfig(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 1, CountingEnabled: false})
		bj.Reset()
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		bj.SetPhase(domain.BJPhaseAction)
		_ = bj.PlayerStand()
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result.CpuPlayers))
		// In end phase, CPU hands should have cards
		assert.NotNil(t, result.CpuPlayers[0].Hands)
	})
	t.Run("success Output no CPU players when cpuPlayerCount is 0", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Nil(t, result.CpuPlayers)
		assert.Equal(t, 0, result.CpuPlayerCount)
	})
	t.Run("success Output default config values", func(t *testing.T) {
		bj := domain.NewDefaultBlackJack()
		bj.Reset()
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.False(t, result.DealerHitsSoft17)
		assert.False(t, result.CountingEnabled)
		assert.Equal(t, 0, result.CpuPlayerCount)
		assert.Equal(t, 0, result.RunningCount)
		assert.Equal(t, 0.0, result.TrueCount)
		assert.True(t, result.DoubleAfterSplit, "default DAS should be true")
	})
	t.Run("success Output includes doubleAfterSplit false", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		_ = bj.SetConfig(domain.BlackJackConfig{DoubleAfterSplit: false})
		bj.Reset()
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.False(t, result.DoubleAfterSplit)
	})
	t.Run("success Output with no side bet results", func(t *testing.T) {
		tc := domain.NewTrumpCardsWithDecks(1, 0)
		for i := 0; i < 10; i++ {
			tc.Shuffle()
		}
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		bj.Reset()
		// Manually set up hand with a pair
		hand := bj.GetPlayerHands()[0]
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.SetBet(100)
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
		// Use PlayerBet to trigger side bet evaluation - but we already manually set cards,
		// so instead test the getter directly
		bj.SetPhase(domain.BJPhaseAction)
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, 0, result.PerfectPairsBet)
		assert.Equal(t, 0, result.TwentyOnePlus3Bet)
		assert.Nil(t, result.SideBetResults)
	})
	t.Run("success Output CPU hand cards in action phase", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		player := domain.NewBlackJackPlayer()
		dealer := domain.NewBlackJackPlayer()
		player.SetChips(1000)
		dealer.SetChips(1000)
		bj := domain.NewBlackJack(tc, player, dealer)
		_ = bj.SetConfig(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 1, CountingEnabled: false})
		bj.Reset()
		// Add cards to CPU hand so the card loop body is exercised
		cpuPlayers := bj.GetCpuPlayers()
		cpuHand := cpuPlayers[0].GetHands()[0]
		cpuHand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		cpuHand.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		cpuHand.SetBet(50)
		cpuHand.SetStood(true)
		// Player hand
		hand := bj.GetPlayerHands()[0]
		hand.SetBet(100)
		hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
		hand.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
		dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
		bj.SetPhase(domain.BJPhaseAction)
		output := tbp.Output(bj, nil)
		var result controller.BlackJackWebOutput
		err := json.Unmarshal([]byte(output), &result)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(result.CpuPlayers))
		assert.Equal(t, 2, len(result.CpuPlayers[0].Hands[0].Cards), "CPU hand should have 2 cards in action phase")
		assert.Equal(t, "SPADE", result.CpuPlayers[0].Hands[0].Cards[0].Design)
		assert.Equal(t, 10, result.CpuPlayers[0].Hands[0].Cards[0].Value)
	})
}

func TestBlackJackWebPresenter_DeckPenetration(t *testing.T) {
	tbp := presenter.NewBlackJackWebPresenter()
	bj := domain.NewDefaultBlackJack()
	bj.Reset()
	output := tbp.Output(bj, nil)
	var result controller.BlackJackWebOutput
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Equal(t, 75, result.DeckPenetration)
}

func TestBlackJackWebPresenter_DeckPenetration50(t *testing.T) {
	tbp := presenter.NewBlackJackWebPresenter()
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(1000)
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	_ = bj.SetConfig(domain.BlackJackConfig{DeckPenetration: 50, DoubleAfterSplit: true})
	bj.Reset()
	output := tbp.Output(bj, nil)
	var result controller.BlackJackWebOutput
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Equal(t, 50, result.DeckPenetration)
}

func TestBlackJackWebPresenter_MultiHandCount(t *testing.T) {
	tbp := presenter.NewBlackJackWebPresenter()
	bj := domain.NewDefaultBlackJack()
	bj.Reset()
	bj.GetPlayer().SetChips(2000)
	err := bj.PlayerBet(100, 0, 0, 2)
	assert.NoError(t, err)
	output := tbp.Output(bj, nil)
	var result controller.BlackJackWebOutput
	err = json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Equal(t, 2, result.MultiHandCount)
}

func TestBlackJackWebPresenter_CpuInsuranceBet(t *testing.T) {
	tbp := presenter.NewBlackJackWebPresenter()
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(1000)
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	_ = bj.SetConfig(domain.BlackJackConfig{DealerHitsSoft17: false, CpuPlayerCount: 1, CountingEnabled: false})
	bj.Reset()
	cpuPlayers := bj.GetCpuPlayers()
	cpuHand := cpuPlayers[0].GetHands()[0]
	cpuHand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	cpuHand.AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
	cpuHand.SetBet(100)
	cpuPlayers[0].SetInsuranceBet(50)
	hand := bj.GetPlayerHands()[0]
	hand.SetBet(100)
	hand.AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	hand.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 1, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))
	bj.SetPhase(domain.BJPhaseAction)
	output := tbp.Output(bj, nil)
	var result controller.BlackJackWebOutput
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.CpuPlayers))
	assert.Equal(t, 50, result.CpuPlayers[0].InsuranceBet)
}

func TestBlackJackWebPresenter_SurrenderRule(t *testing.T) {
	tbp := presenter.NewBlackJackWebPresenter()
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(1000)
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	_ = bj.SetConfig(domain.BlackJackConfig{SurrenderRule: domain.BJSurrenderEarly, DoubleAfterSplit: true})
	bj.Reset()
	output := tbp.Output(bj, nil)
	var result controller.BlackJackWebOutput
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Equal(t, domain.BJSurrenderEarly, result.SurrenderRule)
}

func TestBlackJackWebPresenter_CanSurrenderHand(t *testing.T) {
	tbp := presenter.NewBlackJackWebPresenter()
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(1000)
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	_ = bj.SetConfig(domain.BlackJackConfig{SurrenderRule: domain.BJSurrenderLate, DoubleAfterSplit: true})
	bj.Reset()
	hand := bj.GetPlayerHands()[0]
	hand.SetBet(100)
	hand.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	hand.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))
	bj.SetPhase(domain.BJPhaseAction)
	output := tbp.Output(bj, nil)
	var result controller.BlackJackWebOutput
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	// CanSurrender should reflect bj.CanSurrenderHand(0) which checks game-level rules
	assert.Equal(t, bj.CanSurrenderHand(0), result.Hands[0].CanSurrender)
}

func TestBlackJackWebPresenter_EarlySurrenderPhase(t *testing.T) {
	tbp := presenter.NewBlackJackWebPresenter()
	tc := domain.NewTrumpCards(0)
	player := domain.NewBlackJackPlayer()
	dealer := domain.NewBlackJackPlayer()
	player.SetChips(1000)
	dealer.SetChips(1000)
	bj := domain.NewBlackJack(tc, player, dealer)
	_ = bj.SetConfig(domain.BlackJackConfig{SurrenderRule: domain.BJSurrenderEarly, DoubleAfterSplit: true})
	bj.Reset()
	hand := bj.GetPlayerHands()[0]
	hand.SetBet(100)
	hand.AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
	hand.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))
	dealer.AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))
	bj.SetPhase(domain.BJPhaseEarlySurrender)
	output := tbp.Output(bj, nil)
	var result controller.BlackJackWebOutput
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err)
	assert.Equal(t, domain.BJPhaseEarlySurrender, result.Phase)
}

func TestBlackJackWebPresenter_ActionLogOutput(t *testing.T) {
	p := presenter.NewBlackJackWebPresenter()

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockBlackJackGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 0, PlayerIdx: 0, ActionType: "hit", Detail: "drew a card", Cards: []*domain.Card{domain.NewCard(domain.CardDesignSpade, 10, true)}},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, `"actionType":"hit"`)
		assert.Contains(t, result, `"detail":"drew a card"`)
		mockGame.AssertExpectations(t)
	})

	t.Run("nil entries", func(t *testing.T) {
		mockGame := new(interfaces.MockBlackJackGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, `"entries":[]`)
		mockGame.AssertExpectations(t)
	})
}
