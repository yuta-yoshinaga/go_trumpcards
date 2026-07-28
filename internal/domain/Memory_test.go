package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestMemory() *Memory {
	config := DefaultMemoryConfig()
	players := []*MemoryPlayer{
		NewMemoryPlayer(true),
		NewMemoryPlayer(false),
		NewMemoryPlayer(false),
		NewMemoryPlayer(false),
	}
	tc := NewTrumpCards(0)
	return NewMemory(tc, players, config)
}

func setupTestBoard(m *Memory) {
	// Set up a deterministic board: pairs at known positions
	// positions 0,1 = rank 1, positions 2,3 = rank 2, etc.
	for i := 0; i < MemoryBoardSize; i++ {
		rank := (i / 2) + 1
		if rank > 13 {
			rank = rank - 13
		}
		design := CardDesignSpade
		if i%2 == 1 {
			design = CardDesignHeart
		}
		m.board[i] = &MemoryBoardCard{
			Card:   NewCard(design, rank, false),
			FaceUp: false,
			Taken:  false,
		}
	}
}

func TestNewMemory(t *testing.T) {
	m := newTestMemory()
	assert.NotNil(t, m)
	assert.Equal(t, 4, m.GetPlayerCnt())
	assert.Equal(t, -1, m.GetWinnerIdx())
}

func TestNewDefaultMemory(t *testing.T) {
	m := NewDefaultMemory()
	assert.NotNil(t, m)
	assert.Equal(t, MemoryPlayerCnt, m.GetPlayerCnt())
	assert.True(t, m.GetPlayer(0).GetIsHuman())
	for i := 1; i < m.GetPlayerCnt(); i++ {
		assert.False(t, m.GetPlayer(i).GetIsHuman(), "player %d should be CPU", i)
	}
	assert.Equal(t, -1, m.GetWinnerIdx())
	assert.False(t, m.GetGameEndFlag())
}

func TestMemoryReset(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	assert.Equal(t, MemoryPhaseFlip1, m.GetPhase())
	assert.Equal(t, 0, m.GetCurrentPlayerIdx())
	assert.Equal(t, -1, m.GetFirstFlipPos())
	assert.Equal(t, -1, m.GetSecondFlipPos())
	assert.False(t, m.GetLastMatchResult())
	assert.False(t, m.GetGameEndFlag())
	assert.Equal(t, -1, m.GetWinnerIdx())
	assert.Equal(t, 0, m.GetTurnNumber())
	assert.Nil(t, m.GetActionLog())

	// Board should be set up
	for i := 0; i < MemoryBoardSize; i++ {
		bc := m.GetBoardCard(i)
		assert.NotNil(t, bc)
		assert.False(t, bc.FaceUp)
		assert.False(t, bc.Taken)
		assert.NotNil(t, bc.Card)
	}

	// All players reset
	for i := 0; i < m.GetPlayerCnt(); i++ {
		p := m.GetPlayer(i)
		assert.Equal(t, 0, p.GetPairCount())
	}
}

func TestMemoryPlayerFlip(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m)

	t.Run("first flip", func(t *testing.T) {
		err := m.PlayerFlip(0)
		assert.NoError(t, err)
		assert.True(t, m.GetBoardCard(0).FaceUp)
		assert.Equal(t, MemoryPhaseFlip2, m.GetPhase())
		assert.Equal(t, 0, m.GetFirstFlipPos())
	})

	t.Run("second flip", func(t *testing.T) {
		err := m.PlayerFlip(1)
		assert.NoError(t, err)
		assert.True(t, m.GetBoardCard(1).FaceUp)
		assert.Equal(t, MemoryPhaseResult, m.GetPhase())
		assert.Equal(t, 1, m.GetSecondFlipPos())
	})
}

func TestMemoryPlayerFlipErrors(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m)

	t.Run("invalid position negative", func(t *testing.T) {
		err := m.PlayerFlip(-1)
		assert.Error(t, err)
	})
	t.Run("invalid position too large", func(t *testing.T) {
		err := m.PlayerFlip(52)
		assert.Error(t, err)
	})
	t.Run("card already taken", func(t *testing.T) {
		m.board[5].Taken = true
		err := m.PlayerFlip(5)
		assert.Error(t, err)
	})
	t.Run("card already face up", func(t *testing.T) {
		m.board[10].FaceUp = true
		err := m.PlayerFlip(10)
		assert.Error(t, err)
	})
	t.Run("not human turn", func(t *testing.T) {
		m.SetCurrentPlayerIdx(1) // CPU
		err := m.PlayerFlip(0)
		assert.Error(t, err)
	})
	t.Run("game already over", func(t *testing.T) {
		m.gameEndFlag = true
		m.SetCurrentPlayerIdx(0)
		err := m.PlayerFlip(0)
		assert.Error(t, err)
	})
}

func TestMemoryResolveFlipMatch(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m) // 0,1 are both rank 1

	_ = m.PlayerFlip(0)
	_ = m.PlayerFlip(1)
	assert.Equal(t, MemoryPhaseResult, m.GetPhase())

	m.ResolveFlip()
	assert.True(t, m.GetLastMatchResult())
	assert.True(t, m.GetBoardCard(0).Taken)
	assert.True(t, m.GetBoardCard(1).Taken)
	assert.Equal(t, 1, m.GetPlayer(0).GetPairCount())
	assert.Equal(t, MemoryPhaseFlip1, m.GetPhase())
	// Same player continues
	assert.Equal(t, 0, m.GetCurrentPlayerIdx())
}

func TestMemoryResolveFlipMiss(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m)
	// 0=rank1, 2=rank2 → mismatch

	_ = m.PlayerFlip(0)
	_ = m.PlayerFlip(2)
	m.ResolveFlip()

	assert.False(t, m.GetLastMatchResult())
	assert.False(t, m.GetBoardCard(0).FaceUp) // flipped back
	assert.False(t, m.GetBoardCard(2).FaceUp)
	assert.False(t, m.GetBoardCard(0).Taken)
	assert.Equal(t, 0, m.GetPlayer(0).GetPairCount())
	assert.Equal(t, 1, m.GetCurrentPlayerIdx()) // next player
	assert.Equal(t, MemoryPhaseFlip1, m.GetPhase())
}

func TestMemoryResolveFlipNotInResultPhase(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m)
	// Not in result phase → no-op
	m.SetPhase(MemoryPhaseFlip1)
	m.ResolveFlip()
	assert.Equal(t, MemoryPhaseFlip1, m.GetPhase())
}

func TestMemoryGameEnd(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m)

	// Mark all but last 2 as taken
	for i := 2; i < MemoryBoardSize; i++ {
		m.board[i].Taken = true
	}

	// Flip the last pair (positions 0, 1 both rank 1)
	_ = m.PlayerFlip(0)
	_ = m.PlayerFlip(1)
	m.ResolveFlip()

	assert.True(t, m.GetGameEndFlag())
	assert.Equal(t, MemoryPhaseGameEnd, m.GetPhase())
	assert.Equal(t, 0, m.GetWinnerIdx()) // human has 1 pair, others 0
}

func TestMemoryIsHumanTurn(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	assert.True(t, m.IsHumanTurn())
	m.SetCurrentPlayerIdx(1)
	assert.False(t, m.IsHumanTurn())
}

func TestMemoryGetBoardCardOutOfRange(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	assert.Nil(t, m.GetBoardCard(-1))
	assert.Nil(t, m.GetBoardCard(52))
}

func TestMemoryGetPlayerOutOfRange(t *testing.T) {
	m := newTestMemory()
	assert.Nil(t, m.GetPlayer(-1))
	assert.Nil(t, m.GetPlayer(4))
}

func TestMemorySetConfig(t *testing.T) {
	m := newTestMemory()
	cfg := MemoryConfig{CpuDifficulty: MemoryCpuDifficultyHard}
	m.SetConfig(cfg)
	assert.Equal(t, cfg, m.GetConfig())
}

func TestMemorySetBoard(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	board := make([]*MemoryBoardCard, MemoryBoardSize)
	for i := 0; i < MemoryBoardSize; i++ {
		board[i] = &MemoryBoardCard{
			Card:   NewCard(CardDesignSpade, 1, false),
			FaceUp: false,
			Taken:  false,
		}
	}
	m.SetBoard(board)
	assert.Equal(t, board, m.GetBoard())
}

func TestMemoryCpuFlip(t *testing.T) {
	t.Run("CPU flips when game not ended", func(t *testing.T) {
		m := newTestMemory()
		m.Reset()
		setupTestBoard(m)
		m.SetCurrentPlayerIdx(1) // CPU player
		m.CpuFlip()
		// CPU should have flipped 2 cards, now in Result phase
		assert.Equal(t, MemoryPhaseResult, m.GetPhase())
	})

	t.Run("CPU does nothing when game over", func(t *testing.T) {
		m := newTestMemory()
		m.Reset()
		setupTestBoard(m)
		m.gameEndFlag = true
		m.SetCurrentPlayerIdx(1)
		m.CpuFlip()
		assert.Equal(t, MemoryPhaseFlip1, m.GetPhase())
	})

	t.Run("CPU does nothing when human turn", func(t *testing.T) {
		m := newTestMemory()
		m.Reset()
		setupTestBoard(m)
		m.SetCurrentPlayerIdx(0) // human
		m.CpuFlip()
		assert.Equal(t, MemoryPhaseFlip1, m.GetPhase())
	})

	t.Run("CPU uses known pair", func(t *testing.T) {
		m := newTestMemory()
		m.Reset()
		setupTestBoard(m)
		m.SetCurrentPlayerIdx(1)

		// Give CPU knowledge of a pair
		cpu := m.GetPlayer(1)
		rank := m.board[0].Card.GetValue()
		cpu.RecordRevealedCard(0, rank, 1.0, 0)
		cpu.RecordRevealedCard(1, rank, 1.0, 0)

		m.CpuFlip()
		assert.Equal(t, MemoryPhaseResult, m.GetPhase())
	})
}

func TestMemoryCpuFlipKnownMatchOnFirstFlip(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m)
	m.SetCurrentPlayerIdx(1)
	m.SetConfig(MemoryConfig{CpuDifficulty: MemoryCpuDifficultyHard})

	// CPU knows position 1 has same rank as position 0
	cpu := m.GetPlayer(1)
	rank := m.board[0].Card.GetValue()
	// Only tell CPU about position 1 (not position 0)
	cpu.RecordRevealedCard(1, rank, 1.0, 0)

	// CPU won't find a full pair from memory, will pick random first card.
	// If it happens to pick pos 0, it'll find known match at pos 1.
	// Otherwise it picks a random second card.
	// Either way, should end up in Result phase.
	m.CpuFlip()
	assert.Equal(t, MemoryPhaseResult, m.GetPhase())
}

func TestMemoryDetermineWinnerTieBreak(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m)
	// Give player 0 and player 1 equal pairs, player 0 should win (first to reach max)
	m.GetPlayer(0).SetPairCount(5)
	m.GetPlayer(1).SetPairCount(5)
	m.GetPlayer(2).SetPairCount(3)
	m.GetPlayer(3).SetPairCount(3)

	// Mark all as taken to trigger game end
	for i := 0; i < MemoryBoardSize; i++ {
		m.board[i].Taken = true
	}
	// Undo last two to allow a valid flip
	m.board[0].Taken = false
	m.board[1].Taken = false

	_ = m.PlayerFlip(0)
	_ = m.PlayerFlip(1)
	m.ResolveFlip()

	// Player 0 has 6 pairs now (5+1), wins
	assert.True(t, m.GetGameEndFlag())
	assert.Equal(t, 0, m.GetWinnerIdx())
}

func TestMemoryRetentionChanceByDifficulty(t *testing.T) {
	m := newTestMemory()
	m.SetConfig(MemoryConfig{CpuDifficulty: MemoryCpuDifficultyEasy})
	assert.InDelta(t, memoryRetentionEasy, m.retentionChance(), 0.001)
	m.SetConfig(MemoryConfig{CpuDifficulty: MemoryCpuDifficultyNormal})
	assert.InDelta(t, memoryRetentionNormal, m.retentionChance(), 0.001)
	m.SetConfig(MemoryConfig{CpuDifficulty: MemoryCpuDifficultyHard})
	assert.InDelta(t, memoryRetentionHard, m.retentionChance(), 0.001)
}

func TestMemoryDecayRateByDifficulty(t *testing.T) {
	m := newTestMemory()
	m.SetConfig(MemoryConfig{CpuDifficulty: MemoryCpuDifficultyEasy})
	assert.InDelta(t, memoryDecayEasy, m.decayRate(), 0.001)
	m.SetConfig(MemoryConfig{CpuDifficulty: MemoryCpuDifficultyNormal})
	assert.InDelta(t, memoryDecayNormal, m.decayRate(), 0.001)
	m.SetConfig(MemoryConfig{CpuDifficulty: MemoryCpuDifficultyHard})
	assert.InDelta(t, memoryDecayHard, m.decayRate(), 0.001)
}

func TestMemoryActionLog(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m)

	_ = m.PlayerFlip(0)
	_ = m.PlayerFlip(1)
	m.ResolveFlip()

	log := m.GetActionLog()
	assert.Len(t, log, 1)
	assert.Equal(t, "match", log[0].ActionType)
}

func TestMemoryAdvancePlayerWraps(t *testing.T) {
	m := newTestMemory()
	m.Reset()
	setupTestBoard(m)
	m.SetCurrentPlayerIdx(3)
	// Flip mismatched cards
	m.SetPhase(MemoryPhaseFlip1)
	_ = m.flip(0)
	_ = m.flip(2)
	m.ResolveFlip()
	assert.Equal(t, 0, m.GetCurrentPlayerIdx()) // wraps around
}

// --- ADR-0035: 可変ペア数 ---

func TestDefaultMemoryConfig_PairCountIsFullDeck(t *testing.T) {
	cfg := DefaultMemoryConfig()
	assert.Equal(t, MemoryMaxPairCount, cfg.PairCount, "既定は従来どおりフルデッキ相当でなければならない")
	assert.Equal(t, MemoryBoardSize, cfg.PairCount*2)
}

func TestMemory_ResetHonorsPairCount(t *testing.T) {
	for _, pairs := range []int{MemoryMinPairCount, 20, MemoryMaxPairCount} {
		m := newTestMemory()
		cfg := DefaultMemoryConfig()
		cfg.PairCount = pairs
		m.SetConfig(cfg)
		m.Reset()
		assert.Len(t, m.GetBoard(), pairs*2, "ペア数 %d のとき盤面は %d 枚", pairs, pairs*2)
	}
}

// 一致判定はランク基準なので、配られたカードは各ランクが偶数枚でなければ
// 相方のいないカードが残り、allTaken() に到達できずゲームが終われなくなる。
// 52枚では各ランク4枚なので自然に成立していたが、枚数を減らすと崩れうる。
func TestMemory_ResetDealsOnlyMatchableCards(t *testing.T) {
	for _, pairs := range []int{MemoryMinPairCount, 13, 20, 25, MemoryMaxPairCount} {
		m := newTestMemory()
		cfg := DefaultMemoryConfig()
		cfg.PairCount = pairs
		m.SetConfig(cfg)
		m.Reset()

		byRank := map[int]int{}
		for _, bc := range m.GetBoard() {
			if assert.NotNil(t, bc, "空きマスがあってはならない") && assert.NotNil(t, bc.Card) {
				byRank[bc.Card.GetValue()]++
			}
		}
		for rank, n := range byRank {
			assert.Zero(t, n%2, "ペア数 %d: ランク %d が %d 枚で奇数 — 相方がおらずゲームが終わらない", pairs, rank, n)
		}
	}
}

func TestMemoryConfig_ClampPairCount(t *testing.T) {
	cases := []struct{ in, want int }{
		{0, MemoryMaxPairCount}, // 未設定は既定へ
		{MemoryMinPairCount - 1, MemoryMinPairCount},
		{MemoryMaxPairCount + 1, MemoryMaxPairCount},
		{20, 20},
	}
	for _, c := range cases {
		cfg := MemoryConfig{PairCount: c.in}
		assert.Equal(t, c.want, cfg.NormalizedPairCount(), "入力 %d", c.in)
	}
}
