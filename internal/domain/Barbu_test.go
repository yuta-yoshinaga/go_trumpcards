//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// card 短縮コンストラクタは PaiGow_test.go の同名ヘルパーを共用する。

// --- Config ---

func TestBarbuConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultBarbuConfig().Validate())
	assert.NoError(t, domain.BarbuConfig{CpuDifficulty: domain.BarbuDifficultyHard}.Validate())
	assert.Error(t, domain.BarbuConfig{CpuDifficulty: -1}.Validate())
	assert.Error(t, domain.BarbuConfig{CpuDifficulty: 99}.Validate())
}

func TestBarbuConfig_JSONRoundTrip(t *testing.T) {
	cfg := domain.BarbuConfig{CpuDifficulty: domain.BarbuDifficultyHard}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	var got domain.BarbuConfig
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, cfg, got)
}

// --- Player ---

func TestBarbuPlayer_CapturedHelpers(t *testing.T) {
	p := domain.NewBarbuPlayer(false)
	p.AddTrick([]*domain.Card{card(domain.CardDesignHeart, 13), card(domain.CardDesignHeart, 2), card(domain.CardDesignSpade, 12)})
	p.AddTrick([]*domain.Card{card(domain.CardDesignDiamond, 12), card(domain.CardDesignClover, 5)})
	assert.Equal(t, 2, p.CapturedHearts())
	assert.Equal(t, 2, p.CapturedQueens()) // Q♠ + Q♦
	assert.True(t, p.HasKingOfHearts())

	p2 := domain.NewBarbuPlayer(true)
	assert.False(t, p2.HasKingOfHearts())
	assert.Equal(t, 0, p2.CapturedHearts())
}

func TestBarbuPlayer_ScoreAndReset(t *testing.T) {
	p := domain.NewBarbuPlayer(false)
	p.AddScore(10)
	p.AddScore(-3)
	assert.Equal(t, 7, p.GetTotalScore())
	p.SetDominoRank(2)
	p.AddTrick([]*domain.Card{card(domain.CardDesignSpade, 1)})
	p.AddCard(card(domain.CardDesignClover, 4))
	p.ResetDeal()
	assert.Equal(t, 0, p.GetDominoRank())
	assert.Equal(t, 0, p.GetTrickCount())
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Equal(t, 7, p.GetTotalScore()) // 累計は維持
	p.ResetTotalScore()
	assert.Equal(t, 0, p.GetTotalScore())
}

func TestBarbuPlayer_JSONRoundTrip(t *testing.T) {
	p := domain.NewBarbuPlayer(true)
	p.AddCard(card(domain.CardDesignHeart, 7))
	p.AddTrick([]*domain.Card{card(domain.CardDesignSpade, 1)})
	p.SetDominoRank(3)
	p.AddScore(15)
	data, err := json.Marshal(p)
	require.NoError(t, err)
	var got domain.BarbuPlayer
	require.NoError(t, json.Unmarshal(data, &got))
	assert.True(t, got.GetIsHuman())
	assert.Equal(t, 1, got.GetCardsSize())
	assert.Equal(t, 3, got.GetDominoRank())
	assert.Equal(t, 15, got.GetTotalScore())
	assert.Equal(t, 1, got.GetTrickCount())
}

// --- Contract scoring (deterministic) ---

func newScoringGame(t *testing.T, contract, trumpSuit int) *domain.Barbu {
	t.Helper()
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(contract, trumpSuit)
	return b
}

func TestBarbuScore_NoTricks(t *testing.T) {
	b := newScoringGame(t, domain.BarbuContractNoTricks, -1)
	b.BarbuTestAddTrick(0, []*domain.Card{card(domain.CardDesignSpade, 5)})
	b.BarbuTestAddTrick(0, []*domain.Card{card(domain.CardDesignSpade, 6)})
	b.BarbuTestAddTrick(2, []*domain.Card{card(domain.CardDesignClover, 9)})
	d := b.BarbuTestScoreDeal()
	assert.Equal(t, -2*domain.BarbuNoTrickPenalty, d.Gained[0])
	assert.Equal(t, 0, d.Gained[1])
	assert.Equal(t, -1*domain.BarbuNoTrickPenalty, d.Gained[2])
}

func TestBarbuScore_NoHearts(t *testing.T) {
	b := newScoringGame(t, domain.BarbuContractNoHearts, -1)
	b.BarbuTestAddTrick(1, []*domain.Card{card(domain.CardDesignHeart, 2), card(domain.CardDesignHeart, 9), card(domain.CardDesignSpade, 3)})
	d := b.BarbuTestScoreDeal()
	assert.Equal(t, -2*domain.BarbuHeartPenalty, d.Gained[1])
	assert.Equal(t, 0, d.Gained[0])
}

func TestBarbuScore_NoQueens(t *testing.T) {
	b := newScoringGame(t, domain.BarbuContractNoQueens, -1)
	b.BarbuTestAddTrick(3, []*domain.Card{card(domain.CardDesignHeart, 12), card(domain.CardDesignSpade, 12)})
	d := b.BarbuTestScoreDeal()
	assert.Equal(t, -2*domain.BarbuQueenPenalty, d.Gained[3])
}

func TestBarbuScore_KingHeart(t *testing.T) {
	b := newScoringGame(t, domain.BarbuContractKingHeart, -1)
	b.BarbuTestAddTrick(2, []*domain.Card{card(domain.CardDesignHeart, 13)})
	d := b.BarbuTestScoreDeal()
	assert.Equal(t, -domain.BarbuKingHeartPenalty, d.Gained[2])
	assert.Equal(t, 0, d.Gained[0])
}

func TestBarbuScore_NoLastTrick(t *testing.T) {
	b := newScoringGame(t, domain.BarbuContractNoLastTrick, -1)
	b.BarbuTestSetLastTrickWinner(1)
	d := b.BarbuTestScoreDeal()
	assert.Equal(t, -domain.BarbuLastTrickPenalty, d.Gained[1])
	assert.Equal(t, 0, d.Gained[0])
}

func TestBarbuScore_Trumps(t *testing.T) {
	b := newScoringGame(t, domain.BarbuContractTrumps, domain.CardDesignSpade)
	b.BarbuTestAddTrick(0, []*domain.Card{card(domain.CardDesignSpade, 5)})
	b.BarbuTestAddTrick(0, []*domain.Card{card(domain.CardDesignSpade, 6)})
	d := b.BarbuTestScoreDeal()
	assert.Equal(t, 2*domain.BarbuTrumpReward, d.Gained[0])
}

func TestBarbuScore_Dominoes(t *testing.T) {
	b := newScoringGame(t, domain.BarbuContractDominoes, -1)
	b.GetPlayer(0).SetDominoRank(1)
	b.GetPlayer(1).SetDominoRank(2)
	b.GetPlayer(2).SetDominoRank(3)
	b.GetPlayer(3).SetDominoRank(4)
	d := b.BarbuTestScoreDeal()
	assert.Equal(t, domain.BarbuDominoScores[0], d.Gained[0])
	assert.Equal(t, domain.BarbuDominoScores[1], d.Gained[1])
	assert.Equal(t, domain.BarbuDominoScores[2], d.Gained[2])
	assert.Equal(t, domain.BarbuDominoScores[3], d.Gained[3])
}

func TestBarbuScore_DominoesUnfinishedIsLast(t *testing.T) {
	b := newScoringGame(t, domain.BarbuContractDominoes, -1)
	b.GetPlayer(0).SetDominoRank(1)
	// player 1-3 未上がり (rank 0) → 最下位扱い
	d := b.BarbuTestScoreDeal()
	assert.Equal(t, domain.BarbuDominoScores[0], d.Gained[0])
	assert.Equal(t, domain.BarbuDominoScores[domain.BarbuPlayerCnt-1], d.Gained[1])
}

// --- SelectContract validation ---

func TestBarbuSelectContract_Valid(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetHand(0, []*domain.Card{card(domain.CardDesignSpade, 2)})
	require.NoError(t, b.SelectContract(domain.BarbuContractNoTricks, -1))
	assert.Equal(t, domain.BarbuPhasePlay, b.GetPhase())
	assert.Equal(t, domain.BarbuContractNoTricks, b.GetCurrentContract())
	assert.Equal(t, 0, b.GetCurrentTurn())
	assert.Equal(t, 1, b.GetTrickNumber())
}

func TestBarbuSelectContract_TrumpsRequiresSuit(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	assert.Error(t, b.SelectContract(domain.BarbuContractTrumps, -1))
	require.NoError(t, b.SelectContract(domain.BarbuContractTrumps, domain.CardDesignHeart))
	assert.Equal(t, domain.CardDesignHeart, b.GetTrumpSuit())
}

func TestBarbuSelectContract_Errors(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	assert.Error(t, b.SelectContract(99, -1)) // unknown
	assert.Error(t, b.SelectContract(-1, -1)) // unknown
	b.BarbuTestSetUsedContract(0, domain.BarbuContractNoHearts, true)
	assert.Error(t, b.SelectContract(domain.BarbuContractNoHearts, -1)) // already used

	// wrong phase
	b2 := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b2.BarbuTestSetPhase(domain.BarbuPhasePlay)
	assert.Error(t, b2.SelectContract(domain.BarbuContractNoTricks, -1))

	// not human dealer
	b3 := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b3.BarbuTestSetDealer(1)
	assert.ErrorIs(t, b3.SelectContract(domain.BarbuContractNoTricks, -1), domain.ErrNotHumanTurn)

	// game ended
	b4 := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b4.BarbuTestSetGameEnd(true)
	assert.ErrorIs(t, b4.SelectContract(domain.BarbuContractNoTricks, -1), domain.ErrGameEnded)
}

// --- Trick play & winner ---

func TestBarbuTrickPlay_FollowSuit(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(domain.BarbuContractNoTricks, -1)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetTrickNumber(1)
	b.BarbuTestSetCurrentPlayer(0)
	b.BarbuTestSetLeadPlayer(0)
	b.BarbuTestSetHand(0, []*domain.Card{card(domain.CardDesignSpade, 5), card(domain.CardDesignHeart, 9)})
	// human leads spade 5 via the public API
	require.NoError(t, b.PlayerPlay(0, nil))
	assert.Len(t, b.GetCurrentTrick(), 1)
	assert.Equal(t, 1, b.GetCurrentTurn()) // advanced

	// Player 1 has a spade and a heart, must follow spade.
	b.BarbuTestSetHand(1, []*domain.Card{card(domain.CardDesignHeart, 3), card(domain.CardDesignSpade, 8)})
	// playing the heart (idx 0) is illegal
	assert.Error(t, b.BarbuTestApplyTrickPlay(1, 0))
	// playing the spade (idx 1) is legal
	require.NoError(t, b.BarbuTestApplyTrickPlay(1, 1))
}

func TestBarbuTrickWinner_NoTrump(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(domain.BarbuContractNoTricks, -1)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetTrickNumber(1)
	b.BarbuTestSetCurrentPlayer(0)
	b.BarbuTestSetLeadPlayer(0)
	// craft 4 hands of 1 card each; lead is spade, highest spade wins
	b.BarbuTestSetHand(0, []*domain.Card{card(domain.CardDesignSpade, 5)})
	b.BarbuTestSetHand(1, []*domain.Card{card(domain.CardDesignSpade, 11)})
	b.BarbuTestSetHand(2, []*domain.Card{card(domain.CardDesignHeart, 13)}) // off-suit, can't win
	b.BarbuTestSetHand(3, []*domain.Card{card(domain.CardDesignSpade, 9)})
	for i := 0; i < 4; i++ {
		require.NoError(t, b.BarbuTestApplyTrickPlay(i, 0))
	}
	// player 1 (spade 11) wins, gets the trick
	assert.Equal(t, 1, b.GetLastTrickWinner())
	assert.Equal(t, 1, b.GetPlayer(1).GetTrickCount())
}

func TestBarbuTrickWinner_Trump(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(domain.BarbuContractTrumps, domain.CardDesignDiamond)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetTrickNumber(1)
	b.BarbuTestSetLeadPlayer(0)
	b.BarbuTestSetHand(0, []*domain.Card{card(domain.CardDesignSpade, 13)})  // lead high spade
	b.BarbuTestSetHand(1, []*domain.Card{card(domain.CardDesignDiamond, 2)}) // low trump beats
	b.BarbuTestSetHand(2, []*domain.Card{card(domain.CardDesignSpade, 1)})
	b.BarbuTestSetHand(3, []*domain.Card{card(domain.CardDesignClover, 10)})
	for i := 0; i < 4; i++ {
		require.NoError(t, b.BarbuTestApplyTrickPlay(i, 0))
	}
	assert.Equal(t, 1, b.GetLastTrickWinner()) // trump wins
}

// TestBarbuTrickWinner_AceHigh は A がトリックで最強 (K より強い) ことを確認する。
func TestBarbuTrickWinner_AceHigh(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(domain.BarbuContractNoTricks, -1)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetTrickNumber(1)
	b.BarbuTestSetLeadPlayer(0)
	// 同スートで A(1) は K(13) に勝つ。
	b.BarbuTestSetHand(0, []*domain.Card{card(domain.CardDesignSpade, 13)}) // K leads
	b.BarbuTestSetHand(1, []*domain.Card{card(domain.CardDesignSpade, 1)})  // A
	b.BarbuTestSetHand(2, []*domain.Card{card(domain.CardDesignSpade, 12)}) // Q
	b.BarbuTestSetHand(3, []*domain.Card{card(domain.CardDesignSpade, 2)})  // 2
	for i := 0; i < 4; i++ {
		require.NoError(t, b.BarbuTestApplyTrickPlay(i, 0))
	}
	assert.Equal(t, 1, b.GetLastTrickWinner()) // A (player 1) wins
}

// TestBarbuTrickWinner_AceTrumpHigh は切り札の A が切り札の K に勝つことを確認する。
func TestBarbuTrickWinner_AceTrumpHigh(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(domain.BarbuContractTrumps, domain.CardDesignSpade)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetTrickNumber(1)
	b.BarbuTestSetLeadPlayer(0)
	b.BarbuTestSetHand(0, []*domain.Card{card(domain.CardDesignSpade, 13)}) // K♠ (trump) leads
	b.BarbuTestSetHand(1, []*domain.Card{card(domain.CardDesignSpade, 1)})  // A♠ (trump)
	b.BarbuTestSetHand(2, []*domain.Card{card(domain.CardDesignHeart, 5)})
	b.BarbuTestSetHand(3, []*domain.Card{card(domain.CardDesignSpade, 7)})
	for i := 0; i < 4; i++ {
		require.NoError(t, b.BarbuTestApplyTrickPlay(i, 0))
	}
	assert.Equal(t, 1, b.GetLastTrickWinner()) // A♠ (player 1) wins
}

// TestBarbuDominoes_AceStaysLow は Dominoes では A が低位 (2 の隣) として
// 扱われることを確認する (トリックの A-high 化が漏れていないこと)。
func TestBarbuDominoes_AceStaysLow(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(domain.BarbuContractDominoes, -1)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetCurrentPlayer(0)
	var table [5]uint16
	for v := 2; v <= 7; v++ { // spades 2..7 placed
		table[domain.CardDesignSpade] |= uint16(1) << uint(v)
	}
	b.BarbuTestSetTablePlaced(table)
	b.BarbuTestSetHand(0, []*domain.Card{card(domain.CardDesignSpade, 1)}) // A♠ playable next to 2♠
	assert.Equal(t, []int{0}, b.GetDominoPlayableIndices(0))
}

// --- Dominoes mechanic ---

func TestBarbuDominoes_PlayableAndPlace(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(domain.BarbuContractDominoes, -1)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetCurrentPlayer(0)
	// empty board: only 7s playable
	b.BarbuTestSetHand(0, []*domain.Card{card(domain.CardDesignSpade, 6), card(domain.CardDesignSpade, 7)})
	idxs := b.GetDominoPlayableIndices(0)
	require.Equal(t, []int{1}, idxs) // only the 7
	// pass not allowed when a play exists
	assert.Error(t, b.PlayerPlay(-1, nil))
	// place the 7
	require.NoError(t, b.PlayerPlay(1, nil))
	// now spade 6 becomes playable for player 0 next time
	tbl := b.GetTablePlaced()
	assert.NotZero(t, tbl[domain.CardDesignSpade]&(1<<7))
}

func TestBarbuDominoes_PassWhenNoMove(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(domain.BarbuContractDominoes, -1)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetCurrentPlayer(0)
	b.BarbuTestSetHand(0, []*domain.Card{card(domain.CardDesignSpade, 2)}) // not playable on empty board
	assert.Empty(t, b.GetDominoPlayableIndices(0))
	require.NoError(t, b.PlayerPlay(-1, nil)) // pass OK
	assert.NotEqual(t, 0, b.GetCurrentTurn()) // advanced to another player
}

// --- CPU ---

func TestBarbuCpu_SelectContractValid(t *testing.T) {
	for diff := domain.BarbuDifficultyEasy; diff <= domain.BarbuDifficultyHard; diff++ {
		b := domain.BarbuTestNew(domain.BarbuConfig{CpuDifficulty: diff})
		b.BarbuTestSetDealer(1)
		b.BarbuTestSetHand(1, []*domain.Card{
			card(domain.CardDesignSpade, 7), card(domain.CardDesignSpade, 13),
			card(domain.CardDesignHeart, 2), card(domain.CardDesignDiamond, 5),
		})
		// mark 6 of 7 contracts used → only Dominoes left
		for c := 0; c < domain.BarbuContractCnt; c++ {
			if c != domain.BarbuContractDominoes {
				b.BarbuTestSetUsedContract(1, c, true)
			}
		}
		b.CpuPlay() // should select Dominoes and move to play phase
		assert.Equal(t, domain.BarbuPhasePlay, b.GetPhase())
		assert.Equal(t, domain.BarbuContractDominoes, b.GetCurrentContract())
	}
}

func TestBarbuCpu_TrickPlayProgresses(t *testing.T) {
	b := domain.BarbuTestNew(domain.BarbuConfig{CpuDifficulty: domain.BarbuDifficultyHard})
	b.BarbuTestSetContract(domain.BarbuContractTrumps, domain.CardDesignSpade)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetTrickNumber(1)
	b.BarbuTestSetCurrentPlayer(1)
	b.BarbuTestSetLeadPlayer(1)
	b.BarbuTestSetHand(1, []*domain.Card{card(domain.CardDesignSpade, 3), card(domain.CardDesignHeart, 9)})
	before := b.GetPlayer(1).GetCardsSize()
	b.CpuPlay()
	assert.Equal(t, before-1, b.GetPlayer(1).GetCardsSize())
}

// --- IsHumanTurn ---

func TestBarbuIsHumanTurn(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	assert.True(t, b.IsHumanTurn()) // dealer 0 selects

	b.BarbuTestSetDealer(2)
	assert.False(t, b.IsHumanTurn()) // CPU dealer

	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetCurrentPlayer(0)
	assert.True(t, b.IsHumanTurn())
	b.BarbuTestSetCurrentPlayer(3)
	assert.False(t, b.IsHumanTurn())

	b.BarbuTestSetPhase(domain.BarbuPhaseDealEnd)
	assert.False(t, b.IsHumanTurn())

	b.BarbuTestSetGameEnd(true)
	assert.False(t, b.IsHumanTurn())
}

// --- Reset / full game integration ---

func TestBarbuReset_DealsHands(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	assert.Equal(t, domain.BarbuPhaseSelectContract, b.GetPhase())
	assert.Equal(t, 0, b.GetDealNumber())
	assert.Equal(t, 0, b.GetDealerIdx())
	total := 0
	for i := 0; i < b.GetPlayerCnt(); i++ {
		total += b.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 52, total)
	assert.False(t, b.GetGameEndFlag())
}

// humanFirstValidPlay drives one human action (used for the integration game).
func humanFirstValidPlay(t *testing.T, b *domain.Barbu) {
	t.Helper()
	switch b.GetPhase() {
	case domain.BarbuPhaseSelectContract:
		used := b.GetUsedContracts(b.GetDealerIdx())
		for c := 0; c < domain.BarbuContractCnt; c++ {
			if !used[c] {
				trump := -1
				if c == domain.BarbuContractTrumps {
					trump = domain.CardDesignSpade
				}
				require.NoError(t, b.SelectContract(c, trump))
				return
			}
		}
		t.Fatal("no unused contract for human dealer")
	case domain.BarbuPhasePlay:
		if b.GetCurrentContract() == domain.BarbuContractDominoes {
			idxs := b.GetDominoPlayableIndices(b.GetCurrentTurn())
			if len(idxs) == 0 {
				require.NoError(t, b.PlayerPlay(-1, nil))
				return
			}
			require.NoError(t, b.PlayerPlay(idxs[0], nil))
			return
		}
		n := b.GetPlayer(b.GetCurrentTurn()).GetCardsSize()
		for i := 0; i < n; i++ {
			if err := b.PlayerPlay(i, nil); err == nil {
				return
			}
		}
		t.Fatal("no valid trick card for human")
	}
}

func TestBarbuFullGame_ReachesGameEnd(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	guard := 0
	for !b.GetGameEndFlag() {
		guard++
		require.Less(t, guard, 20000, "play loop did not terminate")
		switch b.GetPhase() {
		case domain.BarbuPhaseDealEnd:
			b.NextDeal()
		case domain.BarbuPhaseSelectContract, domain.BarbuPhasePlay:
			if b.IsHumanTurn() {
				humanFirstValidPlay(t, b)
			} else {
				b.CpuPlay()
			}
		default:
			t.Fatalf("unexpected phase %q", b.GetPhase())
		}
	}
	assert.Equal(t, domain.BarbuPhaseGameEnd, b.GetPhase())
	assert.Equal(t, domain.BarbuTotalDeals-1, b.GetDealNumber())
	// every dealer used all 7 contracts exactly once
	for d := 0; d < domain.BarbuPlayerCnt; d++ {
		used := b.GetUsedContracts(d)
		for c := 0; c < domain.BarbuContractCnt; c++ {
			assert.True(t, used[c], "dealer %d should have used contract %d", d, c)
		}
	}
	winners := b.GetRoundWinners()
	assert.NotEmpty(t, winners)
}

// --- Serialization ---

func TestBarbu_JSONRoundTrip(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	// advance a few CPU/contract steps to populate state
	_ = b.SelectContract(domain.BarbuContractNoHearts, -1)

	data, err := json.Marshal(b)
	require.NoError(t, err)
	var got domain.Barbu
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, b.GetPhase(), got.GetPhase())
	assert.Equal(t, b.GetCurrentContract(), got.GetCurrentContract())
	assert.Equal(t, b.GetDealerIdx(), got.GetDealerIdx())
	assert.Equal(t, b.GetPlayerCnt(), got.GetPlayerCnt())
	assert.Equal(t, b.GetCurrentTurn(), got.GetCurrentTurn())
}

func TestBarbu_UnmarshalRejectsMissingDeck(t *testing.T) {
	var got domain.Barbu
	err := json.Unmarshal([]byte(`{"ph":"play"}`), &got)
	assert.Error(t, err)
}

func TestBarbuDealDetail_JSONRoundTrip(t *testing.T) {
	d := &domain.BarbuDealDetail{
		Contract:  domain.BarbuContractTrumps,
		TrumpSuit: domain.CardDesignSpade,
		DealerIdx: 2,
		Gained:    map[int]int{0: 5, 1: 0, 2: 10, 3: -5},
	}
	data, err := json.Marshal(d)
	require.NoError(t, err)
	var got domain.BarbuDealDetail
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, d.Contract, got.Contract)
	assert.Equal(t, d.Gained, got.Gained)
}

// --- Deal history retention ---

func TestBarbuDealHistory_AppendsAndResets(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	// A fresh game has an empty (non-nil) history.
	assert.Empty(t, b.GetDealHistory())

	// Finishing a deal records exactly one entry mirroring the last deal detail.
	b.BarbuTestSetContract(domain.BarbuContractNoTricks, -1)
	b.BarbuTestAddTrick(0, []*domain.Card{card(domain.CardDesignSpade, 5)})
	b.BarbuTestFinishDeal()
	hist := b.GetDealHistory()
	require.Len(t, hist, 1)
	assert.Same(t, b.GetLastDealDetail(), hist[0])
	assert.Equal(t, domain.BarbuContractNoTricks, hist[0].Contract)

	// Reset clears the retained history.
	b.Reset()
	assert.Empty(t, b.GetDealHistory())
}

func TestBarbuDealHistory_JSONRoundTrip(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	b.BarbuTestSetContract(domain.BarbuContractTrumps, domain.CardDesignSpade)
	b.BarbuTestAddTrick(0, []*domain.Card{card(domain.CardDesignSpade, 5)})
	b.BarbuTestFinishDeal()

	data, err := json.Marshal(b)
	require.NoError(t, err)
	var got domain.Barbu
	require.NoError(t, json.Unmarshal(data, &got))
	require.Len(t, got.GetDealHistory(), 1)
	assert.Equal(t, domain.BarbuContractTrumps, got.GetDealHistory()[0].Contract)
}

func TestBarbuDealHistory_UnmarshalRejectsOversizedHistory(t *testing.T) {
	// Build a valid state then swap in an oversized deal-history array.
	b := domain.NewDefaultBarbu()
	b.Reset()
	data, err := json.Marshal(b)
	require.NoError(t, err)
	var raw map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &raw))
	big := make([]map[string]int, 1001)
	rawBig, err := json.Marshal(big)
	require.NoError(t, err)
	raw["dh"] = rawBig
	tampered, err := json.Marshal(raw)
	require.NoError(t, err)
	var got domain.Barbu
	assert.Error(t, json.Unmarshal(tampered, &got))
}

// --- NextDeal guard ---

func TestBarbuNextDeal_RotatesDealer(t *testing.T) {
	b := domain.NewDefaultBarbu()
	b.Reset()
	b.BarbuTestSetPhase(domain.BarbuPhaseDealEnd)
	b.NextDeal()
	assert.Equal(t, 1, b.GetDealNumber())
	assert.Equal(t, 1, b.GetDealerIdx())
	assert.Equal(t, domain.BarbuPhaseSelectContract, b.GetPhase())

	// NextDeal is a no-op outside DealEnd
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.NextDeal()
	assert.Equal(t, 1, b.GetDealNumber())
}

// **フォロー義務は validateTrickPlay が持つ (#4804)。**GetPlayableIndices が
// それと同じ判定を返すこと。別のスキャンだと「出せる」と見えた札が弾かれる。
func TestBarbu_GetPlayableIndices(t *testing.T) {
	b := domain.BarbuTestNew(domain.DefaultBarbuConfig())
	b.BarbuTestSetContract(domain.BarbuContractNoTricks, -1)
	b.BarbuTestSetPhase(domain.BarbuPhasePlay)
	b.BarbuTestSetCurrentPlayer(0)
	b.BarbuTestSetHand(0, []*domain.Card{
		domain.NewCard(domain.CardDesignSpade, 7, false),
		domain.NewCard(domain.CardDesignHeart, 2, false),
		domain.NewCard(domain.CardDesignSpade, 3, false),
	})

	// リードは任意。
	assert.Equal(t, []int{0, 1, 2}, b.GetPlayableIndices(0))

	// ♠ リードに追従できるなら ♠ だけ。
	b.BarbuTestSetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 13, false)},
	})
	assert.Equal(t, []int{0, 2}, b.GetPlayableIndices(0))

	// 追従できないスートのリードなら全部。
	b.BarbuTestSetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignClover, 9, false)},
	})
	assert.Equal(t, []int{0, 1, 2}, b.GetPlayableIndices(0))

	// 範囲外は nil。
	assert.Nil(t, b.GetPlayableIndices(99))
}
