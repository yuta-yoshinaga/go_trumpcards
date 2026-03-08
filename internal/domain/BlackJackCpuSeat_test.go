package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func TestNewBlackJackCpuSeat(t *testing.T) {
	cpu := domain.NewBlackJackCpuSeat()
	assert.NotNil(t, cpu)
	assert.NotNil(t, cpu.GetPlayer())
	assert.Equal(t, domain.BJDefaultChips, cpu.GetPlayer().GetChips())
	assert.NotNil(t, cpu.GetHands())
	assert.Equal(t, 1, len(cpu.GetHands()))
	assert.Equal(t, 0, cpu.GetHands()[0].GetCardsSize())
}

func TestBlackJackCpuSeat_GetPlayer(t *testing.T) {
	cpu := domain.NewBlackJackCpuSeat()
	player := cpu.GetPlayer()
	assert.NotNil(t, player)
	assert.Equal(t, domain.BJDefaultChips, player.GetChips())

	// Modify chips and verify
	player.SetChips(500)
	assert.Equal(t, 500, cpu.GetPlayer().GetChips())
}

func TestBlackJackCpuSeat_GetHands(t *testing.T) {
	cpu := domain.NewBlackJackCpuSeat()
	hands := cpu.GetHands()
	assert.Equal(t, 1, len(hands))

	// Add a card to the hand
	hands[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	assert.Equal(t, 1, cpu.GetHands()[0].GetCardsSize())
}

func TestBlackJackCpuSeat_SetHands(t *testing.T) {
	cpu := domain.NewBlackJackCpuSeat()

	// Create new hands (simulating a split)
	hand1 := domain.NewBlackJackHand()
	hand1.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
	hand1.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	hand1.SetBet(50)

	hand2 := domain.NewBlackJackHand()
	hand2.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))
	hand2.AddCard(domain.NewCard(domain.CardDesignDiamond, 5, false))
	hand2.SetBet(50)

	cpu.SetHands([]*domain.BlackJackHand{hand1, hand2})

	assert.Equal(t, 2, len(cpu.GetHands()))
	assert.Equal(t, 11, cpu.GetHands()[0].GetScore())
	assert.Equal(t, 13, cpu.GetHands()[1].GetScore())
}

func TestBlackJackCpuSeat_InsuranceBet(t *testing.T) {
	cpu := domain.NewBlackJackCpuSeat()

	// Default is 0
	assert.Equal(t, 0, cpu.GetInsuranceBet())

	// Set and get
	cpu.SetInsuranceBet(25)
	assert.Equal(t, 25, cpu.GetInsuranceBet())

	// Update
	cpu.SetInsuranceBet(50)
	assert.Equal(t, 50, cpu.GetInsuranceBet())
}

func TestBlackJackCpuSeat_Reset(t *testing.T) {
	cpu := domain.NewBlackJackCpuSeat()

	// Add cards and modify state
	cpu.GetHands()[0].AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	cpu.GetHands()[0].AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))
	cpu.GetHands()[0].SetBet(100)
	cpu.GetHands()[0].SetStood(true)
	cpu.GetPlayer().SetChips(500)
	cpu.SetInsuranceBet(25)

	// Add a second hand (split scenario)
	hand2 := domain.NewBlackJackHand()
	hand2.AddCard(domain.NewCard(domain.CardDesignClover, 5, false))
	cpu.SetHands(append(cpu.GetHands(), hand2))
	assert.Equal(t, 2, len(cpu.GetHands()))

	cpu.Reset()

	// After reset: player cards cleared, hands reset to 1 empty hand
	assert.Equal(t, 0, cpu.GetPlayer().GetCardsSize())
	assert.Equal(t, 1, len(cpu.GetHands()))
	assert.Equal(t, 0, cpu.GetHands()[0].GetCardsSize())
	assert.Equal(t, 0, cpu.GetHands()[0].GetBet())
	assert.False(t, cpu.GetHands()[0].IsStood())

	// Chips should be preserved
	assert.Equal(t, 500, cpu.GetPlayer().GetChips())

	// Insurance bet should be cleared
	assert.Equal(t, 0, cpu.GetInsuranceBet())
}
