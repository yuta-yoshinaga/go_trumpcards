package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBaccarat(t *testing.T) {
	tc := NewTrumpCards(0)
	b := NewBaccarat(tc)
	assert.NotNil(t, b)
	assert.Equal(t, BaccaratPhaseBet, b.GetPhase())
	assert.False(t, b.GetGameEndFlag())
}

func TestNewDefaultBaccarat(t *testing.T) {
	b := NewDefaultBaccarat()
	assert.NotNil(t, b)
	assert.Equal(t, BaccaratDefaultChips, b.GetChips())
	assert.Equal(t, BaccaratPhaseBet, b.GetPhase())
}

func TestBaccarat_Reset(t *testing.T) {
	b := NewDefaultBaccarat()

	t.Run("reset clears state", func(t *testing.T) {
		b.SetPhase(BaccaratPhaseEnd)
		b.SetGameEndFlag(true)
		b.SetPlayerHand([]*Card{NewCard(CardDesignSpade, 1, false)})
		b.SetBankerHand([]*Card{NewCard(CardDesignHeart, 2, false)})
		b.SetBetAmount(100)
		b.SetBetType(BaccaratBetBanker)
		b.SetResult(GameResultWin)
		b.Reset()
		assert.Equal(t, BaccaratPhaseBet, b.GetPhase())
		assert.False(t, b.GetGameEndFlag())
		assert.Nil(t, b.GetPlayerHand())
		assert.Nil(t, b.GetBankerHand())
		assert.Equal(t, 0, b.GetBetAmount())
		assert.Equal(t, 0, b.GetBetType())
		assert.Equal(t, GameResult(0), b.GetResult())
		assert.Equal(t, 0, b.GetPayout())
		assert.Nil(t, b.GetActionLog())
	})

	t.Run("reset refills chips when below minimum", func(t *testing.T) {
		b.SetChips(1) // below BaccaratMinBet
		b.Reset()
		assert.Equal(t, BaccaratDefaultChips, b.GetChips())
	})

	t.Run("reset keeps chips when above minimum", func(t *testing.T) {
		b.SetChips(500)
		b.Reset()
		assert.Equal(t, 500, b.GetChips())
	})
}

func TestBaccarat_Bet_Errors(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		b := NewDefaultBaccarat()
		b.SetPhase(BaccaratPhaseEnd)
		err := b.Bet(100, BaccaratBetPlayer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bet phase")
	})

	t.Run("invalid bet type", func(t *testing.T) {
		b := NewDefaultBaccarat()
		err := b.Bet(100, 3)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid bet type")
	})

	t.Run("invalid bet type negative", func(t *testing.T) {
		b := NewDefaultBaccarat()
		err := b.Bet(100, -1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid bet type")
	})

	t.Run("bet too small", func(t *testing.T) {
		b := NewDefaultBaccarat()
		err := b.Bet(5, BaccaratBetPlayer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid bet amount")
	})

	t.Run("bet not multiple of min", func(t *testing.T) {
		b := NewDefaultBaccarat()
		err := b.Bet(15, BaccaratBetPlayer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid bet amount")
	})

	t.Run("bet too large", func(t *testing.T) {
		b := NewDefaultBaccarat()
		b.SetChips(20000)
		err := b.Bet(BaccaratMaxBet+10, BaccaratBetPlayer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid bet amount")
	})

	t.Run("insufficient chips", func(t *testing.T) {
		b := NewDefaultBaccarat()
		b.SetChips(50)
		err := b.Bet(100, BaccaratBetPlayer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Insufficient chips")
	})
}

func TestBaccarat_Bet_Success(t *testing.T) {
	b := NewDefaultBaccarat()
	err := b.Bet(100, BaccaratBetPlayer)
	assert.NoError(t, err)
	assert.Equal(t, BaccaratPhaseEnd, b.GetPhase())
	assert.True(t, b.GetGameEndFlag())
	assert.Equal(t, 100, b.GetBetAmount())
	assert.Equal(t, BaccaratBetPlayer, b.GetBetType())
	assert.NotNil(t, b.GetPlayerHand())
	assert.NotNil(t, b.GetBankerHand())
	assert.GreaterOrEqual(t, len(b.GetPlayerHand()), 2)
	assert.GreaterOrEqual(t, len(b.GetBankerHand()), 2)
	assert.LessOrEqual(t, len(b.GetPlayerHand()), 3)
	assert.LessOrEqual(t, len(b.GetBankerHand()), 3)
	assert.NotNil(t, b.GetActionLog())
}

func TestBaccarat_CardPointValue(t *testing.T) {
	b := NewDefaultBaccarat()
	tests := []struct {
		name     string
		card     *Card
		expected int
	}{
		{"ace", NewCard(CardDesignSpade, 1, false), 1},
		{"two", NewCard(CardDesignSpade, 2, false), 2},
		{"nine", NewCard(CardDesignSpade, 9, false), 9},
		{"ten", NewCard(CardDesignSpade, 10, false), 0},
		{"jack", NewCard(CardDesignSpade, 11, false), 0},
		{"queen", NewCard(CardDesignSpade, 12, false), 0},
		{"king", NewCard(CardDesignSpade, 13, false), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, b.cardPointValue(tt.card))
		})
	}
}

func TestBaccarat_CalculateHandValue(t *testing.T) {
	b := NewDefaultBaccarat()

	t.Run("simple sum", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignHeart, 4, false),
		}
		assert.Equal(t, 7, b.CalculateHandValue(cards))
	})

	t.Run("mod 10", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 8, false),
			NewCard(CardDesignHeart, 7, false),
		}
		assert.Equal(t, 5, b.CalculateHandValue(cards))
	})

	t.Run("face cards are zero", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 10, false),
			NewCard(CardDesignHeart, 13, false),
		}
		assert.Equal(t, 0, b.CalculateHandValue(cards))
	})

	t.Run("three cards", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 4, false),
			NewCard(CardDesignHeart, 3, false),
			NewCard(CardDesignDiamond, 6, false),
		}
		assert.Equal(t, 3, b.CalculateHandValue(cards))
	})

	t.Run("empty hand", func(t *testing.T) {
		assert.Equal(t, 0, b.CalculateHandValue(nil))
	})
}

func TestBaccarat_ShouldPlayerDraw(t *testing.T) {
	b := NewDefaultBaccarat()

	for total := 0; total <= 5; total++ {
		assert.True(t, b.shouldPlayerDraw(total), "should draw for total %d", total)
	}
	for total := 6; total <= 9; total++ {
		assert.False(t, b.shouldPlayerDraw(total), "should not draw for total %d", total)
	}
}

func TestBaccarat_ShouldBankerDraw(t *testing.T) {
	b := NewDefaultBaccarat()

	t.Run("player did not draw", func(t *testing.T) {
		for total := 0; total <= 5; total++ {
			assert.True(t, b.shouldBankerDraw(total, 0, false), "banker total %d", total)
		}
		for total := 6; total <= 9; total++ {
			assert.False(t, b.shouldBankerDraw(total, 0, false), "banker total %d", total)
		}
	})

	t.Run("banker 0-2 always draw", func(t *testing.T) {
		for total := 0; total <= 2; total++ {
			for pv := 0; pv <= 9; pv++ {
				assert.True(t, b.shouldBankerDraw(total, pv, true))
			}
		}
	})

	t.Run("banker 3", func(t *testing.T) {
		for pv := 0; pv <= 9; pv++ {
			if pv == 8 {
				assert.False(t, b.shouldBankerDraw(3, pv, true))
			} else {
				assert.True(t, b.shouldBankerDraw(3, pv, true))
			}
		}
	})

	t.Run("banker 4", func(t *testing.T) {
		for pv := 0; pv <= 9; pv++ {
			if pv >= 2 && pv <= 7 {
				assert.True(t, b.shouldBankerDraw(4, pv, true))
			} else {
				assert.False(t, b.shouldBankerDraw(4, pv, true))
			}
		}
	})

	t.Run("banker 5", func(t *testing.T) {
		for pv := 0; pv <= 9; pv++ {
			if pv >= 4 && pv <= 7 {
				assert.True(t, b.shouldBankerDraw(5, pv, true))
			} else {
				assert.False(t, b.shouldBankerDraw(5, pv, true))
			}
		}
	})

	t.Run("banker 6", func(t *testing.T) {
		for pv := 0; pv <= 9; pv++ {
			if pv == 6 || pv == 7 {
				assert.True(t, b.shouldBankerDraw(6, pv, true))
			} else {
				assert.False(t, b.shouldBankerDraw(6, pv, true))
			}
		}
	})

	t.Run("banker 7+ never draws", func(t *testing.T) {
		for total := 7; total <= 9; total++ {
			for pv := 0; pv <= 9; pv++ {
				assert.False(t, b.shouldBankerDraw(total, pv, true))
			}
		}
	})
}

func TestBaccarat_CalculatePayout(t *testing.T) {
	b := NewDefaultBaccarat()

	t.Run("player bet - player wins", func(t *testing.T) {
		b.SetBetType(BaccaratBetPlayer)
		b.SetBetAmount(100)
		b.SetResult(GameResultWin)
		assert.Equal(t, 200, b.calculatePayout())
	})

	t.Run("player bet - tie", func(t *testing.T) {
		b.SetBetType(BaccaratBetPlayer)
		b.SetBetAmount(100)
		b.SetResult(GameResultDraw)
		assert.Equal(t, 100, b.calculatePayout())
	})

	t.Run("player bet - banker wins", func(t *testing.T) {
		b.SetBetType(BaccaratBetPlayer)
		b.SetBetAmount(100)
		b.SetResult(GameResultLose)
		assert.Equal(t, 0, b.calculatePayout())
	})

	t.Run("banker bet - banker wins", func(t *testing.T) {
		b.SetBetType(BaccaratBetBanker)
		b.SetBetAmount(100)
		b.SetResult(GameResultLose) // GameResultLose = banker wins
		// 100 + (100 - 5) = 195
		assert.Equal(t, 195, b.calculatePayout())
	})

	t.Run("banker bet - tie", func(t *testing.T) {
		b.SetBetType(BaccaratBetBanker)
		b.SetBetAmount(100)
		b.SetResult(GameResultDraw)
		assert.Equal(t, 100, b.calculatePayout())
	})

	t.Run("banker bet - player wins", func(t *testing.T) {
		b.SetBetType(BaccaratBetBanker)
		b.SetBetAmount(100)
		b.SetResult(GameResultWin) // player wins = banker loses
		assert.Equal(t, 0, b.calculatePayout())
	})

	t.Run("tie bet - tie", func(t *testing.T) {
		b.SetBetType(BaccaratBetTie)
		b.SetBetAmount(100)
		b.SetResult(GameResultDraw)
		assert.Equal(t, 900, b.calculatePayout()) // 100 + 100*8
	})

	t.Run("tie bet - player wins", func(t *testing.T) {
		b.SetBetType(BaccaratBetTie)
		b.SetBetAmount(100)
		b.SetResult(GameResultWin)
		assert.Equal(t, 0, b.calculatePayout())
	})

	t.Run("tie bet - banker wins", func(t *testing.T) {
		b.SetBetType(BaccaratBetTie)
		b.SetBetAmount(100)
		b.SetResult(GameResultLose)
		assert.Equal(t, 0, b.calculatePayout())
	})

	t.Run("invalid bet type", func(t *testing.T) {
		b.SetBetType(99)
		b.SetBetAmount(100)
		b.SetResult(GameResultWin)
		assert.Equal(t, 0, b.calculatePayout())
	})
}

func TestBaccarat_Natural(t *testing.T) {
	// Natural should not draw any third card
	b := NewDefaultBaccarat()
	// Manually set up a natural scenario
	b.SetPlayerHand([]*Card{
		NewCard(CardDesignSpade, 9, false),
		NewCard(CardDesignHeart, 9, false), // total = 8
	})
	b.SetBankerHand([]*Card{
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignDiamond, 4, false), // total = 7
	})
	b.drawPhase()
	assert.Equal(t, 2, len(b.GetPlayerHand()), "natural should not draw")
	assert.Equal(t, 2, len(b.GetBankerHand()), "natural should not draw")
}

func TestBaccarat_BetTypeName(t *testing.T) {
	assert.Equal(t, "player", betTypeName(BaccaratBetPlayer))
	assert.Equal(t, "banker", betTypeName(BaccaratBetBanker))
	assert.Equal(t, "tie", betTypeName(BaccaratBetTie))
	assert.Equal(t, "unknown", betTypeName(99))
}

func TestBaccarat_Getters(t *testing.T) {
	b := NewDefaultBaccarat()
	b.SetPlayerHand([]*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 5, false),
	})
	b.SetBankerHand([]*Card{
		NewCard(CardDesignClover, 7, false),
		NewCard(CardDesignDiamond, 2, false),
	})
	assert.Equal(t, 8, b.GetPlayerHandValue())
	assert.Equal(t, 9, b.GetBankerHandValue())
}

func TestBaccarat_FullGame_PlayerWins(t *testing.T) {
	b := NewDefaultBaccarat()
	initialChips := b.GetChips()
	err := b.Bet(100, BaccaratBetPlayer)
	assert.NoError(t, err)
	assert.True(t, b.GetGameEndFlag())

	// Verify result is one of the valid results
	result := b.GetResult()
	assert.Contains(t, []GameResult{GameResultWin, GameResultDraw, GameResultLose}, result)

	// Verify chip change is consistent with result
	expectedChips := initialChips - 100 + b.GetPayout()
	assert.Equal(t, expectedChips, b.GetChips())
}

func TestBaccarat_FullGame_AllBetTypes(t *testing.T) {
	for _, bt := range []int{BaccaratBetPlayer, BaccaratBetBanker, BaccaratBetTie} {
		t.Run(betTypeName(bt), func(t *testing.T) {
			b := NewDefaultBaccarat()
			err := b.Bet(100, bt)
			assert.NoError(t, err)
			assert.True(t, b.GetGameEndFlag())
			assert.NotNil(t, b.GetActionLog())
		})
	}
}

func TestBaccarat_ResetAndReplay(t *testing.T) {
	b := NewDefaultBaccarat()
	err := b.Bet(100, BaccaratBetPlayer)
	assert.NoError(t, err)
	b.Reset()
	assert.Equal(t, BaccaratPhaseBet, b.GetPhase())
	err = b.Bet(100, BaccaratBetBanker)
	assert.NoError(t, err)
	assert.True(t, b.GetGameEndFlag())
}
