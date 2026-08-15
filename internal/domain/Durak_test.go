//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// ---- ヘルパー ----

func newTestDurak() *domain.Durak {
	players := []*domain.DurakPlayer{
		domain.NewDurakPlayer(true),
		domain.NewDurakPlayer(false),
		domain.NewDurakPlayer(false),
		domain.NewDurakPlayer(false),
	}
	tc := domain.NewTrumpCardsShortDeck()
	return domain.NewDurak(tc, players)
}

func newTestDurak2P() *domain.Durak {
	players := []*domain.DurakPlayer{
		domain.NewDurakPlayer(true),
		domain.NewDurakPlayer(false),
	}
	tc := domain.NewTrumpCardsShortDeck()
	d := domain.NewDurak(tc, players)
	cfg := domain.DefaultDurakConfig()
	cfg.PlayerCount = 2
	d.SetConfig(cfg)
	return d
}

// setupDurakForHumanAttack configures a Durak game where human is attacker.
func setupDurakForHumanAttack() *domain.Durak {
	d := newTestDurak()
	d.Reset()
	// Ensure human is attacker
	d.SetAttackerIdx(0)
	d.SetDefenderIdx(1)
	d.SetCurrentTurn(0)
	d.SetPhase(domain.DurakPhaseAttack)
	return d
}

// setupDurakForHumanDefend configures a Durak game where human is defender.
func setupDurakForHumanDefend() *domain.Durak {
	d := newTestDurak()
	d.Reset()
	// Set human as defender with a known attack card on table
	d.SetAttackerIdx(1)
	d.SetDefenderIdx(0)
	d.SetCurrentTurn(0)
	d.SetPhase(domain.DurakPhaseDefend)

	// Put an attack card on table
	attackCard := domain.NewCard(domain.CardDesignClover, 7, false)
	d.SetTablePairs([]*domain.DurakTablePair{
		{Attack: attackCard},
	})
	return d
}

// ---- DurakConfig テスト ----

func TestDurakConfig_Default(t *testing.T) {
	cfg := domain.DefaultDurakConfig()
	assert.Equal(t, 4, cfg.PlayerCount)
	assert.Equal(t, domain.DurakDifficultyNormal, cfg.CpuDifficulty)
	assert.False(t, cfg.TransferEnabled)
}

func TestDurakConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*domain.DurakConfig)
		wantErr bool
	}{
		{"valid default", func(_ *domain.DurakConfig) {}, false},
		{"valid 2 players", func(c *domain.DurakConfig) { c.PlayerCount = 2 }, false},
		{"valid 6 players", func(c *domain.DurakConfig) { c.PlayerCount = 6 }, false},
		{"invalid 1 player", func(c *domain.DurakConfig) { c.PlayerCount = 1 }, true},
		{"invalid 7 players", func(c *domain.DurakConfig) { c.PlayerCount = 7 }, true},
		{"invalid difficulty", func(c *domain.DurakConfig) { c.CpuDifficulty = 99 }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := domain.DefaultDurakConfig()
			tt.modify(&cfg)
			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDurakConfig_JSON(t *testing.T) {
	cfg := domain.DefaultDurakConfig()
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	var restored domain.DurakConfig
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, cfg, restored)
}

// ---- DurakPlayer テスト ----

func TestDurakPlayer_New(t *testing.T) {
	p := domain.NewDurakPlayer(true)
	assert.True(t, p.GetIsHuman())
	assert.False(t, p.GetIsFinished())
	assert.Equal(t, 0, p.GetCardsSize())
}

func TestDurakPlayer_JSON(t *testing.T) {
	p := domain.NewDurakPlayer(true)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var restored domain.DurakPlayer
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.True(t, restored.GetIsHuman())
	assert.Equal(t, 1, restored.GetCardsSize())
}

func TestDurakPlayer_SortCards(t *testing.T) {
	p := domain.NewDurakPlayer(false)
	trumpSuit := domain.CardDesignHeart

	// Add cards: trump A, non-trump K, non-trump 6, trump 7
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))  // trump A
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false)) // non-trump K
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))  // non-trump 6
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 7, false))  // trump 7

	p.SortCards(trumpSuit)

	// Non-trumps first (spade 6, spade K), then trumps (heart 7, heart A)
	assert.Equal(t, domain.CardDesignSpade, p.GetCard(0).GetDesign())
	assert.Equal(t, 6, p.GetCard(0).GetValue())
	assert.Equal(t, domain.CardDesignSpade, p.GetCard(1).GetDesign())
	assert.Equal(t, 13, p.GetCard(1).GetValue())
	assert.Equal(t, domain.CardDesignHeart, p.GetCard(2).GetDesign())
	assert.Equal(t, 7, p.GetCard(2).GetValue())
	assert.Equal(t, domain.CardDesignHeart, p.GetCard(3).GetDesign())
	assert.Equal(t, 1, p.GetCard(3).GetValue()) // Ace (rank 14)
}

func TestDurakPlayer_SortCardsByValue(t *testing.T) {
	p := domain.NewDurakPlayer(false)
	trumpSuit := domain.CardDesignHeart

	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 6, false))
	p.AddCard(domain.NewCard(domain.CardDesignClover, 6, false))

	p.SortCardsByValue(trumpSuit)

	// 6s first (non-trump clover, then trump heart), then Ace
	assert.Equal(t, 6, p.GetCard(0).GetValue())
	assert.Equal(t, domain.CardDesignClover, p.GetCard(0).GetDesign())
	assert.Equal(t, 6, p.GetCard(1).GetValue())
	assert.Equal(t, domain.CardDesignHeart, p.GetCard(1).GetDesign())
	assert.Equal(t, 1, p.GetCard(2).GetValue()) // Ace
}

// ---- Durak ゲームテスト ----

func TestDurak_Reset(t *testing.T) {
	d := newTestDurak()
	d.Reset()

	// 各プレイヤー6枚ずつ
	for i := 0; i < d.GetPlayerCnt(); i++ {
		assert.Equal(t, domain.DurakHandSize, d.GetPlayer(i).GetCardsSize())
	}
	// 山札: 36 - 4*6 = 12
	assert.Equal(t, 12, d.GetStockCount())
	assert.Equal(t, domain.DurakPhaseAttack, d.GetPhase())
	assert.False(t, d.GetGameEndFlag())
	assert.Equal(t, -1, d.GetLoserIdx())
	assert.NotNil(t, d.GetTrumpCard())
}

func TestDurak_Reset_2Players(t *testing.T) {
	d := newTestDurak2P()
	d.Reset()
	assert.Equal(t, 2, d.GetPlayerCnt())
	for i := 0; i < d.GetPlayerCnt(); i++ {
		assert.Equal(t, domain.DurakHandSize, d.GetPlayer(i).GetCardsSize())
	}
	// 山札: 36 - 2*6 = 24
	assert.Equal(t, 24, d.GetStockCount())
}

func TestDurak_PlayerAttack_Success(t *testing.T) {
	d := setupDurakForHumanAttack()
	err := d.PlayerAttack(0)
	assert.NoError(t, err)
	assert.Equal(t, domain.DurakPhaseDefend, d.GetPhase())
	assert.Equal(t, 1, len(d.GetTablePairs()))
}

func TestDurak_PlayerAttack_GameEnded(t *testing.T) {
	d := setupDurakForHumanAttack()
	d.SetGameEndFlag(true)
	err := d.PlayerAttack(0)
	assert.ErrorIs(t, err, domain.ErrGameEnded)
}

func TestDurak_PlayerAttack_NotHumanTurn(t *testing.T) {
	d := setupDurakForHumanAttack()
	d.SetCurrentTurn(1) // CPU turn
	err := d.PlayerAttack(0)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestDurak_PlayerAttack_WrongPhase(t *testing.T) {
	d := setupDurakForHumanAttack()
	d.SetPhase(domain.DurakPhaseDefend)
	err := d.PlayerAttack(0)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestDurak_PlayerAttack_InvalidCardIndex(t *testing.T) {
	d := setupDurakForHumanAttack()
	err := d.PlayerAttack(-1)
	assert.ErrorIs(t, err, domain.ErrInvalidCard)
	err = d.PlayerAttack(99)
	assert.ErrorIs(t, err, domain.ErrInvalidCard)
}

func TestDurak_PlayerAttack_InvalidRank(t *testing.T) {
	d := newTestDurak()
	d.Reset()
	d.SetAttackerIdx(0)
	d.SetDefenderIdx(1)
	d.SetCurrentTurn(0)
	d.SetPhase(domain.DurakPhaseAttack)
	d.SetTrumpSuit(domain.CardDesignSpade)

	// Clear and set specific cards
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
	d.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 9, false))

	d.GetPlayer(1).Reset()
	d.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
	d.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	// First attack succeeds
	err := d.PlayerAttack(0) // play 7♣
	assert.NoError(t, err)

	// Now try to add 9 (not matching rank 7 on table)
	d.SetPhase(domain.DurakPhaseAttack)
	d.SetCurrentTurn(0)
	err = d.PlayerAttack(0) // remaining card is 9♣, rank 9 not on table
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
}

func TestDurak_PlayerDefend_Success(t *testing.T) {
	d := setupDurakForHumanDefend()

	// Give human a card that can beat 7♣
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
	d.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
	d.SetTrumpSuit(domain.CardDesignSpade)

	// Give attacker matching cards so attack can continue (prevents bout ending)
	d.GetPlayer(1).Reset()
	d.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 7, false))  // rank 7 matches table
	d.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 9, false))  // rank 9 will match defense
	d.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 10, false)) // padding

	err := d.PlayerDefend(0, 0) // defend first attack with 9♣
	assert.NoError(t, err)
	// After defense, if attacker can continue, phase goes to attack
	assert.Equal(t, domain.DurakPhaseAttack, d.GetPhase())
}

func TestDurak_PlayerDefend_GameEnded(t *testing.T) {
	d := setupDurakForHumanDefend()
	d.SetGameEndFlag(true)
	err := d.PlayerDefend(0, 0)
	assert.ErrorIs(t, err, domain.ErrGameEnded)
}

func TestDurak_PlayerDefend_NotHumanTurn(t *testing.T) {
	d := setupDurakForHumanDefend()
	d.SetCurrentTurn(1) // CPU turn
	err := d.PlayerDefend(0, 0)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestDurak_PlayerDefend_WrongPhase(t *testing.T) {
	d := setupDurakForHumanDefend()
	d.SetPhase(domain.DurakPhaseAttack)
	err := d.PlayerDefend(0, 0)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestDurak_PlayerDefend_InvalidAttackIndex(t *testing.T) {
	d := setupDurakForHumanDefend()
	err := d.PlayerDefend(-1, 0)
	assert.ErrorIs(t, err, domain.ErrInvalidCard)
	err = d.PlayerDefend(5, 0)
	assert.ErrorIs(t, err, domain.ErrInvalidCard)
}

func TestDurak_PlayerDefend_AlreadyDefended(t *testing.T) {
	d := setupDurakForHumanDefend()
	d.GetTablePairs()[0].Defense = domain.NewCard(domain.CardDesignClover, 9, false)
	err := d.PlayerDefend(0, 0)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
}

func TestDurak_PlayerDefend_InvalidHandIndex(t *testing.T) {
	d := setupDurakForHumanDefend()
	err := d.PlayerDefend(0, -1)
	assert.ErrorIs(t, err, domain.ErrInvalidCard)
	err = d.PlayerDefend(0, 99)
	assert.ErrorIs(t, err, domain.ErrInvalidCard)
}

func TestDurak_PlayerDefend_CannotBeat(t *testing.T) {
	d := setupDurakForHumanDefend()

	// Attack is 7♣, give human a weaker card
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 6, false))
	d.SetTrumpSuit(domain.CardDesignSpade)

	err := d.PlayerDefend(0, 0)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
}

func TestDurak_PlayerDefend_TrumpBeatsNonTrump(t *testing.T) {
	d := setupDurakForHumanDefend()

	// Attack is 7♣ (non-trump), defend with 6♠ (trump)
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	d.SetTrumpSuit(domain.CardDesignSpade)

	err := d.PlayerDefend(0, 0)
	assert.NoError(t, err)
}

func TestDurak_PlayerPass_Success(t *testing.T) {
	d := setupDurakForHumanAttack()
	// Need at least one card on table to pass
	d.SetTablePairs([]*domain.DurakTablePair{
		{Attack: domain.NewCard(domain.CardDesignClover, 7, false),
			Defense: domain.NewCard(domain.CardDesignClover, 9, false)},
	})
	err := d.PlayerPass()
	assert.NoError(t, err)
}

func TestDurak_PlayerPass_CannotPassFirstAttack(t *testing.T) {
	d := setupDurakForHumanAttack()
	d.SetTablePairs(nil)
	err := d.PlayerPass()
	assert.ErrorIs(t, err, domain.ErrCannotPass)
}

func TestDurak_PlayerPass_GameEnded(t *testing.T) {
	d := setupDurakForHumanAttack()
	d.SetGameEndFlag(true)
	err := d.PlayerPass()
	assert.ErrorIs(t, err, domain.ErrGameEnded)
}

func TestDurak_PlayerPass_WrongPhase(t *testing.T) {
	d := setupDurakForHumanAttack()
	d.SetPhase(domain.DurakPhaseDefend)
	err := d.PlayerPass()
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestDurak_PlayerTakeCards_Success(t *testing.T) {
	d := newTestDurak()
	d.Reset()
	d.SetAttackerIdx(1)
	d.SetDefenderIdx(0)
	d.SetCurrentTurn(0)
	d.SetPhase(domain.DurakPhaseDefend)

	attackCard := domain.NewCard(domain.CardDesignClover, 7, false)
	d.SetTablePairs([]*domain.DurakTablePair{
		{Attack: attackCard},
	})

	handBefore := d.GetPlayer(0).GetCardsSize()
	err := d.PlayerTakeCards()
	assert.NoError(t, err)
	// Defender should have picked up the attack card
	assert.Greater(t, d.GetPlayer(0).GetCardsSize(), handBefore)
}

func TestDurak_PlayerTakeCards_GameEnded(t *testing.T) {
	d := setupDurakForHumanDefend()
	d.SetGameEndFlag(true)
	err := d.PlayerTakeCards()
	assert.ErrorIs(t, err, domain.ErrGameEnded)
}

func TestDurak_PlayerTakeCards_WrongPhase(t *testing.T) {
	d := setupDurakForHumanDefend()
	d.SetPhase(domain.DurakPhaseAttack)
	err := d.PlayerTakeCards()
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestDurak_CpuPlay_Attack(t *testing.T) {
	d := newTestDurak()
	d.Reset()
	// Make CPU player 1 the attacker
	d.SetAttackerIdx(1)
	d.SetDefenderIdx(2)
	d.SetCurrentTurn(1)
	d.SetPhase(domain.DurakPhaseAttack)

	d.CpuPlay()
	// CPU should have played a card
	assert.Equal(t, 1, len(d.GetTablePairs()))
	assert.Equal(t, domain.DurakPhaseDefend, d.GetPhase())
}

func TestDurak_CpuPlay_Defend(t *testing.T) {
	d := newTestDurak()
	d.Reset()
	d.SetAttackerIdx(0)
	d.SetDefenderIdx(1)
	d.SetCurrentTurn(1)
	d.SetPhase(domain.DurakPhaseDefend)
	d.SetTrumpSuit(domain.CardDesignSpade)

	// Put a weak attack card on table
	d.SetTablePairs([]*domain.DurakTablePair{
		{Attack: domain.NewCard(domain.CardDesignClover, 6, false)},
	})

	d.CpuPlay()
	// CPU should have either defended or taken cards
	assert.NotEqual(t, domain.DurakPhaseDefend, d.GetPhase())
}

func TestDurak_CpuPlay_SkipsHuman(t *testing.T) {
	d := newTestDurak()
	d.Reset()
	d.SetCurrentTurn(0) // human
	phase := d.GetPhase()
	d.CpuPlay()
	// Should not change anything
	assert.Equal(t, phase, d.GetPhase())
}

func TestDurak_CpuPlay_SkipsGameEnd(t *testing.T) {
	d := newTestDurak()
	d.Reset()
	d.SetGameEndFlag(true)
	d.SetCurrentTurn(1)
	d.CpuPlay()
	// nothing should happen
}

func TestDurak_SortHumanHand(t *testing.T) {
	d := newTestDurak()
	d.Reset()
	err := d.SortHumanHand(domain.DurakSortByValue)
	assert.NoError(t, err)
	assert.Equal(t, domain.DurakSortByValue, d.GetSortMode())

	err = d.SortHumanHand(domain.DurakSortBySuit)
	assert.NoError(t, err)
	assert.Equal(t, domain.DurakSortBySuit, d.GetSortMode())
}

func TestDurak_FullGame(t *testing.T) {
	// Run multiple full games to cover various branches
	for range 50 {
		d := newTestDurak()
		d.Reset()

		turns := 0
		for !d.GetGameEndFlag() && turns < 500 {
			if d.IsHumanTurn() {
				switch d.GetPhase() {
				case domain.DurakPhaseAttack:
					if len(d.GetTablePairs()) > 0 {
						// Try to pass
						_ = d.PlayerPass()
					} else {
						// Play first card
						_ = d.PlayerAttack(0)
					}
				case domain.DurakPhaseDefend:
					// Try to take cards
					_ = d.PlayerTakeCards()
				}
			} else {
				d.CpuPlay()
			}
			turns++
		}
		// Game should eventually end
		assert.True(t, d.GetGameEndFlag() || turns >= 500)
	}
}

func TestDurak_FullGame_CpuOnly(t *testing.T) {
	// All CPU players
	players := []*domain.DurakPlayer{
		domain.NewDurakPlayer(false),
		domain.NewDurakPlayer(false),
		domain.NewDurakPlayer(false),
		domain.NewDurakPlayer(false),
	}
	tc := domain.NewTrumpCardsShortDeck()
	d := domain.NewDurak(tc, players)

	for range 20 {
		d.Reset()
		turns := 0
		for !d.GetGameEndFlag() && turns < 500 {
			d.CpuPlay()
			turns++
		}
		assert.True(t, d.GetGameEndFlag())
	}
}

func TestDurak_FullGame_Difficulties(t *testing.T) {
	difficulties := []domain.DurakCpuDifficulty{
		domain.DurakDifficultyEasy,
		domain.DurakDifficultyNormal,
		domain.DurakDifficultyHard,
	}
	for _, diff := range difficulties {
		// Run all-CPU games until one terminates. A specific shuffled deal can
		// occasionally fail to reach an end state within the per-deal turn cap
		// (which made this flake on CI); retrying fresh deals keeps the test
		// deterministic while still catching a systematic non-termination bug
		// (all attempts would fail).
		ended := false
		for attempt := 0; attempt < 20 && !ended; attempt++ {
			players := []*domain.DurakPlayer{
				domain.NewDurakPlayer(false),
				domain.NewDurakPlayer(false),
				domain.NewDurakPlayer(false),
				domain.NewDurakPlayer(false),
			}
			tc := domain.NewTrumpCardsShortDeck()
			d := domain.NewDurak(tc, players)
			cfg := domain.DefaultDurakConfig()
			cfg.CpuDifficulty = diff
			d.SetConfig(cfg)

			d.Reset()
			turns := 0
			for !d.GetGameEndFlag() && turns < 2000 {
				d.CpuPlay()
				turns++
			}
			ended = d.GetGameEndFlag()
		}
		assert.True(t, ended, "game should end for difficulty %d within 20 fresh deals", diff)
	}
}

func TestDurak_JSON(t *testing.T) {
	d := newTestDurak()
	d.Reset()
	data, err := json.Marshal(d)
	require.NoError(t, err)

	var restored domain.Durak
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, d.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, d.GetPhase(), restored.GetPhase())
	assert.Equal(t, d.GetTrumpSuit(), restored.GetTrumpSuit())
	assert.Equal(t, d.GetStockCount(), restored.GetStockCount())
	assert.Equal(t, d.GetAttackerIdx(), restored.GetAttackerIdx())
	assert.Equal(t, d.GetDefenderIdx(), restored.GetDefenderIdx())
}

func TestDurak_JSON_WithTablePairs(t *testing.T) {
	d := newTestDurak()
	d.Reset()
	// Play a turn to get table pairs
	d.SetAttackerIdx(1)
	d.SetDefenderIdx(2)
	d.SetCurrentTurn(1)
	d.SetPhase(domain.DurakPhaseAttack)
	d.CpuPlay() // CPU attacks

	data, err := json.Marshal(d)
	require.NoError(t, err)

	var restored domain.Durak
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, len(d.GetTablePairs()), len(restored.GetTablePairs()))
}

func TestDurak_GetActionLog(t *testing.T) {
	d := newTestDurak()
	d.Reset()
	// Action log starts empty (nil slice)
	assert.Empty(t, d.GetActionLog())

	// After a CPU attack, action log should have entries
	d.SetAttackerIdx(1)
	d.SetDefenderIdx(2)
	d.SetCurrentTurn(1)
	d.SetPhase(domain.DurakPhaseAttack)
	d.CpuPlay()
	assert.Greater(t, len(d.GetActionLog()), 0)
}

func TestDurak_HasPendingAction(t *testing.T) {
	d := newTestDurak()
	assert.False(t, d.HasPendingAction())
}

func TestDurak_GetBoutNumber(t *testing.T) {
	d := newTestDurak()
	d.Reset()
	assert.Equal(t, 0, d.GetBoutNumber())
}

func TestDurak_PlayerAttack_DefenderNoCards(t *testing.T) {
	d := newTestDurak()
	d.Reset()
	d.SetAttackerIdx(0)
	d.SetDefenderIdx(1)
	d.SetCurrentTurn(0)
	d.SetPhase(domain.DurakPhaseAttack)
	d.SetTrumpSuit(domain.CardDesignSpade)

	// Set specific cards
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
	d.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 8, false))

	// Defender has only 1 card
	d.GetPlayer(1).Reset()
	d.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 9, false))

	// Put first card on table (simulating first attack already done)
	d.SetTablePairs([]*domain.DurakTablePair{
		{Attack: domain.NewCard(domain.CardDesignHeart, 7, false)},
	})

	// Try to add another attack card when defender has 1 card and 1 undefended
	err := d.PlayerAttack(0) // 7♣ matches rank 7 on table
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
}

func TestDurak_DefendAllCards_BoutEnds(t *testing.T) {
	d := newTestDurak()
	d.Reset()
	d.SetAttackerIdx(1)
	d.SetDefenderIdx(0)
	d.SetCurrentTurn(0)
	d.SetPhase(domain.DurakPhaseDefend)
	d.SetTrumpSuit(domain.CardDesignSpade)

	// Attacker has no more matching cards
	d.GetPlayer(1).Reset()
	d.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignDiamond, 13, false))

	// Put an attack card on table
	d.SetTablePairs([]*domain.DurakTablePair{
		{Attack: domain.NewCard(domain.CardDesignClover, 7, false)},
	})

	// Defender has matching card
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignClover, 9, false))
	d.GetPlayer(0).AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))

	err := d.PlayerDefend(0, 0) // defend 7♣ with 9♣
	assert.NoError(t, err)
	// Since attacker has no matching cards (rank 7 or 9 not in hand), bout should end
}

func TestDurak_GetCpuActions(t *testing.T) {
	d := newTestDurak()
	d.Reset()
	assert.Nil(t, d.GetCpuActions())

	d.SetAttackerIdx(1)
	d.SetDefenderIdx(2)
	d.SetCurrentTurn(1)
	d.SetPhase(domain.DurakPhaseAttack)
	d.CpuPlay()
	assert.NotNil(t, d.GetCpuActions())
}

func TestDurak_GetHumanAction(t *testing.T) {
	d := setupDurakForHumanAttack()
	assert.Nil(t, d.GetHumanAction())

	_ = d.PlayerAttack(0)
	assert.NotNil(t, d.GetHumanAction())
}

func TestDurak_PlayerPass_WithUndefended(t *testing.T) {
	d := setupDurakForHumanAttack()
	// Table has undefended card
	d.SetTablePairs([]*domain.DurakTablePair{
		{Attack: domain.NewCard(domain.CardDesignClover, 7, false)},
	})
	err := d.PlayerPass()
	assert.NoError(t, err)
	// Phase should switch to defend since there are undefended cards
	assert.Equal(t, domain.DurakPhaseDefend, d.GetPhase())
}

func TestDurak_NotAttackerTurn(t *testing.T) {
	d := newTestDurak()
	d.Reset()
	d.SetAttackerIdx(1) // CPU is attacker
	d.SetCurrentTurn(0) // but human's turn
	d.SetPhase(domain.DurakPhaseAttack)
	err := d.PlayerAttack(0)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

// **他のトリック系はサーバー計算の理由付きヒントを持つのに、Durak は CUI に
// hint コマンドすら無かった (#4740)。**推奨手は CPU の選択ロジックをそのまま使う。
//
// 配りに賭けず、既存の setupDurakForHuman* で局面を作って全分岐を踏む。
func TestDurak_GetHint(t *testing.T) {
	t.Run("initial attack recommends a playable card", func(t *testing.T) {
		d := setupDurakForHumanAttack()
		hint := d.GetHint()
		if hint == nil || hint.CardIndex == nil {
			t.Fatal("初回攻撃では推奨カードが出る")
		}
		if hint.Reason != "attack_weakest" {
			t.Errorf("Reason = %q, want attack_weakest", hint.Reason)
		}
		// **勧めた手が本番の入口を通ること。**別ロジックで選ぶと出せない札を勧める。
		if err := d.PlayerAttack(*hint.CardIndex); err != nil {
			t.Errorf("推奨札 (index %d) が実際には出せなかった: %v", *hint.CardIndex, err)
		}
	})

	// 追撃: テーブルに札があるとき。手札に同じ数字があれば追撃、無ければパス。
	t.Run("additional attack or pass depending on the hand", func(t *testing.T) {
		d := setupDurakForHumanAttack()
		human := d.GetPlayer(0)
		for human.GetCardsSize() > 0 {
			human.RemoveCard(0)
		}
		// テーブルに 7 が出ている状態で、手札にも 7 を持たせる → 追撃できる。
		//
		// **切り札を渡してはいけない。**Normal 難易度の追撃は
		// cpuFindMatchingCard(..., includeTrumps=false) で切り札を除外するため、
		// 渡した 7 がたまたま切り札スートだと追撃されず pass になる。切り札は
		// 配りごとにランダムなので、固定スートだと実測 36/200 (18%) で落ちた。
		follow := domain.CardDesignHeart
		if d.GetTrumpSuit() == follow {
			follow = domain.CardDesignSpade
		}
		d.SetTablePairs([]*domain.DurakTablePair{
			{Attack: domain.NewCard(domain.CardDesignClover, 7, false)},
		})
		human.AddCard(domain.NewCard(follow, 7, false))

		hint := d.GetHint()
		if hint == nil || hint.CardIndex == nil || hint.Reason != "attack_additional" {
			t.Fatalf("同じ数字を持っていれば追撃を勧める: %+v", hint)
		}

		// 手札を無関係な札だけにすると、追撃できずパスになる。
		for human.GetCardsSize() > 0 {
			human.RemoveCard(0)
		}
		// 追撃できない札。こちらも切り札だと「切り札は除外」以前に数字が
		// 合わないので pass になるが、意図を明示するため非切り札にする。
		noMatch := domain.CardDesignSpade
		if d.GetTrumpSuit() == noMatch {
			noMatch = domain.CardDesignHeart
		}
		human.AddCard(domain.NewCard(noMatch, 3, false))
		hint = d.GetHint()
		if hint == nil || hint.CardIndex != nil || hint.Reason != "pass_no_addition" {
			t.Fatalf("追撃できなければパスを勧める: %+v", hint)
		}
	})

	t.Run("defend recommends a card that beats the attack", func(t *testing.T) {
		d := setupDurakForHumanDefend()
		human := d.GetPlayer(0)
		for human.GetCardsSize() > 0 {
			human.RemoveCard(0)
		}
		// 場は CLOVER 7。同スートの上位で返せる。
		human.AddCard(domain.NewCard(domain.CardDesignClover, 10, false))

		hint := d.GetHint()
		if hint == nil || hint.CardIndex == nil || hint.AttackIdx == nil {
			t.Fatalf("防御できる札があれば札と対象を示す: %+v", hint)
		}
		if hint.Reason != "defend_beat" {
			t.Errorf("Reason = %q, want defend_beat", hint.Reason)
		}
		if err := d.PlayerDefend(*hint.AttackIdx, *hint.CardIndex); err != nil {
			t.Errorf("推奨した防御が実際には通らなかった: %v", err)
		}
	})

	// **「引き取る」も助言のうち。**返せる札が無い局面で黙ると、プレイヤーは
	// 手が無いのか判断が付かない。
	t.Run("advises taking the cards when nothing beats the attack", func(t *testing.T) {
		d := setupDurakForHumanDefend()
		human := d.GetPlayer(0)
		for human.GetCardsSize() > 0 {
			human.RemoveCard(0)
		}
		// 場は CLOVER 7。同スートの下位・切り札でない札しか持たない。
		human.AddCard(domain.NewCard(domain.CardDesignClover, 3, false))

		hint := d.GetHint()
		if hint == nil || !hint.TakeCards || hint.Reason != "take_cannot_beat" {
			t.Fatalf("返せなければ引き取りを勧める: %+v", hint)
		}
	})

	t.Run("no hint on a CPU turn", func(t *testing.T) {
		d := setupDurakForHumanAttack()
		d.SetCurrentTurn(1)
		if d.GetHint() != nil {
			t.Error("CPU の手番ではヒントを出さない")
		}
	})

	t.Run("no hint once the game has ended", func(t *testing.T) {
		d := setupDurakForHumanAttack()
		d.SetPhase(domain.DurakPhaseGameEnd)
		if d.GetHint() != nil {
			t.Error("ゲーム終了フェーズではヒントを出さない")
		}
	})
}

// **上限に当たった局も必ず終わり、必ず敗者が決まる。** 引き分け (loserIdx = -1) は
// 「全員上がり」の意味なので、循環の打ち切りをそこへ流すと別の結末と区別できなくなる。
//
// 20 万局に 14 局という頻度なので、1 局ずつ回して踏むのを待つのは現実的でない。
// ここでは上限そのものの規則を確かめ、頻度は #5414 の実測に任せる。
func TestDurak_BoutLimitEndsTheGame(t *testing.T) {
	players := []*domain.DurakPlayer{
		domain.NewDurakPlayer(false), domain.NewDurakPlayer(false),
		domain.NewDurakPlayer(false), domain.NewDurakPlayer(false),
	}
	d := domain.NewDurak(domain.NewTrumpCardsShortDeck(), players)

	// 健全な局は上限に触れない (実測の最大は 78 バウト)。
	t.Run("a normal game finishes well inside the limit", func(t *testing.T) {
		for range 200 {
			d.Reset()
			turns := 0
			for !d.GetGameEndFlag() && turns < 500 {
				d.CpuPlay()
				turns++
			}
			assert.True(t, d.GetGameEndFlag())
			assert.Less(t, d.GetBoutNumber(), domain.DurakMaxBouts)
		}
	})

	// 上限は健全な局の 2 倍以上離れていること。近すぎると普通の局を打ち切る。
	t.Run("the limit is far above a normal game", func(t *testing.T) {
		assert.Greater(t, domain.DurakMaxBouts, 150)
	})

	// 終わった局は必ず敗者を持つ (全員上がりを除く)。
	t.Run("every finished game names a loser", func(t *testing.T) {
		for range 500 {
			d.Reset()
			turns := 0
			for !d.GetGameEndFlag() && turns < 500 {
				d.CpuPlay()
				turns++
			}
			if !d.GetGameEndFlag() {
				continue
			}
			active := 0
			for i := 0; i < d.GetPlayerCnt(); i++ {
				if !d.GetPlayer(i).GetIsFinished() {
					active++
				}
			}
			if active == 0 {
				assert.Equal(t, -1, d.GetLoserIdx(), "全員上がりのときだけ -1")
				continue
			}
			assert.GreaterOrEqual(t, d.GetLoserIdx(), 0, "敗者が決まっていない")
			assert.False(t, d.GetPlayer(d.GetLoserIdx()).GetIsFinished(), "上がった人が敗者になっている")
		}
	})
}

// **打ち切り経路を直接踏む。** 20 万局に 14 局という頻度なので、局を回して
// 待っていては CI で一度も通らない。JSON 往復でバウト数を上限直前に置き、
// 次の bout 終了で必ず打ち切りに入る局面を作る。
func TestDurak_BoutLimitCutoffPath(t *testing.T) {
	newGame := func() *domain.Durak {
		players := []*domain.DurakPlayer{
			domain.NewDurakPlayer(false), domain.NewDurakPlayer(false),
			domain.NewDurakPlayer(false), domain.NewDurakPlayer(false),
		}
		d := domain.NewDurak(domain.NewTrumpCardsShortDeck(), players)
		d.Reset()
		return d
	}

	// バウト数だけ上限直前へ動かす。他の状態は配ったままなので、続きを回せば
	// 数バウトで上限に当たる。
	atLimit := func(d *domain.Durak) *domain.Durak {
		raw, err := json.Marshal(d)
		assert.NoError(t, err)
		var m map[string]any
		assert.NoError(t, json.Unmarshal(raw, &m))
		m["bn"] = domain.DurakMaxBouts - 1
		patched, err := json.Marshal(m)
		assert.NoError(t, err)
		out := newGame()
		assert.NoError(t, json.Unmarshal(patched, out))
		return out
	}

	t.Run("the game ends at the limit", func(t *testing.T) {
		d := atLimit(newGame())
		turns := 0
		for !d.GetGameEndFlag() && turns < 500 {
			d.CpuPlay()
			turns++
		}
		assert.True(t, d.GetGameEndFlag(), "上限を越えても終わっていない")
		assert.GreaterOrEqual(t, d.GetBoutNumber(), domain.DurakMaxBouts)

		// **打ち切りで終わったことを確かめる。** 自然終了なら上がっていない
		// プレイヤーは 1 人だけ。2 人以上残っているなら、終わらせたのは上限。
		active := 0
		for i := 0; i < d.GetPlayerCnt(); i++ {
			if !d.GetPlayer(i).GetIsFinished() {
				active++
			}
		}
		assert.Greater(t, active, 1,
			"自然終了してしまっている -- この局面では打ち切り経路を踏めていない")
	})

	t.Run("the cutoff names a loser who still holds cards", func(t *testing.T) {
		d := atLimit(newGame())
		turns := 0
		for !d.GetGameEndFlag() && turns < 500 {
			d.CpuPlay()
			turns++
		}
		loser := d.GetLoserIdx()
		if loser < 0 {
			// 全員上がりで終わった場合だけ -1 が許される
			for i := 0; i < d.GetPlayerCnt(); i++ {
				assert.True(t, d.GetPlayer(i).GetIsFinished())
			}
			return
		}
		assert.False(t, d.GetPlayer(loser).GetIsFinished(), "上がった人が敗者になっている")

		// 敗者は手札が最多であること。少ない人を負けにすると規則が逆になる。
		for i := 0; i < d.GetPlayerCnt(); i++ {
			if d.GetPlayer(i).GetIsFinished() {
				continue
			}
			assert.LessOrEqual(t, d.GetPlayer(i).GetCardsSize(), d.GetPlayer(loser).GetCardsSize())
		}
	})
}
