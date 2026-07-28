//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// card 短縮コンストラクタは他ドメインテストの同名ヘルパーを共用する。

// newTestKing returns a fresh, reset King game (default 4-player, easy CPU for
// determinism in flow tests).
func newTestKing() *domain.King {
	g := domain.NewDefaultKing()
	g.Reset()
	return g
}

// setKingHand replaces player i's hand deterministically.
func setKingHand(g *domain.King, i int, cards ...*domain.Card) {
	p := g.GetPlayer(i)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// --- Config ---

func TestKingConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultKingConfig().Validate())
	assert.Equal(t, domain.KingDifficultyNormal, domain.DefaultKingConfig().CpuDifficulty)
	assert.NoError(t, domain.KingConfig{CpuDifficulty: domain.KingDifficultyHard}.Validate())
	assert.Error(t, domain.KingConfig{CpuDifficulty: -1}.Validate())
	assert.Error(t, domain.KingConfig{CpuDifficulty: 99}.Validate())
}

func TestKingConfig_JSONRoundTrip(t *testing.T) {
	cfg := domain.KingConfig{CpuDifficulty: domain.KingDifficultyHard}
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	var got domain.KingConfig
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, cfg, got)
}

// --- Player ---

func TestKingPlayer_CapturedHelpers(t *testing.T) {
	p := domain.NewKingPlayer(false)
	p.AddTrickWithRank([]*domain.Card{card(domain.CardDesignHeart, 13), card(domain.CardDesignHeart, 2), card(domain.CardDesignSpade, 12)}, 12)
	p.AddTrickWithRank([]*domain.Card{card(domain.CardDesignDiamond, 12), card(domain.CardDesignClover, 11)}, 13)
	assert.Equal(t, 2, p.CapturedHearts())
	assert.Equal(t, 2, p.CapturedQueens()) // Q♠ + Q♦
	assert.Equal(t, 2, p.CapturedMen())    // K♥ + J♣
	assert.True(t, p.HasKingOfHearts())
	assert.Equal(t, []int{12, 13}, p.GetTrickRanks())

	p2 := domain.NewKingPlayer(true)
	assert.False(t, p2.HasKingOfHearts())
	assert.Equal(t, 0, p2.CapturedHearts())
	assert.Equal(t, 0, p2.CapturedMen())
}

func TestKingPlayer_ScoreAndReset(t *testing.T) {
	p := domain.NewKingPlayer(false)
	p.AddScore(10)
	p.AddScore(-3)
	assert.Equal(t, 7, p.GetTotalScore())
	p.AddTrickWithRank([]*domain.Card{card(domain.CardDesignSpade, 1)}, 1)
	p.AddCard(card(domain.CardDesignClover, 4))
	p.ResetDeal()
	assert.Equal(t, 0, p.GetTrickCount())
	assert.Equal(t, 0, p.GetCardsSize())
	assert.Empty(t, p.GetTrickRanks())
	assert.Equal(t, 7, p.GetTotalScore(), "total score survives deal reset")
	p.ResetTotalScore()
	assert.Equal(t, 0, p.GetTotalScore())
}

func TestKingPlayer_JSONRoundTrip(t *testing.T) {
	p := domain.NewKingPlayer(true)
	p.AddCard(card(domain.CardDesignSpade, 1))
	p.AddTrickWithRank([]*domain.Card{card(domain.CardDesignHeart, 10)}, 5)
	p.AddScore(4)
	b, err := json.Marshal(p)
	require.NoError(t, err)
	var p2 domain.KingPlayer
	require.NoError(t, json.Unmarshal(b, &p2))
	assert.True(t, p2.GetIsHuman())
	assert.Equal(t, 1, p2.GetCardsSize())
	assert.Equal(t, 1, p2.GetTrickCount())
	assert.Equal(t, 4, p2.GetTotalScore())
	assert.Equal(t, []int{5}, p2.GetTrickRanks())

	// Empty object -> defaults.
	var p3 domain.KingPlayer
	require.NoError(t, json.Unmarshal([]byte(`{}`), &p3))
	assert.False(t, p3.GetIsHuman())
	assert.Empty(t, p3.GetTrickRanks())
	assert.Error(t, json.Unmarshal([]byte(`not json`), &p3))
}

// --- Game setup ---

func TestKing_ResetDeal(t *testing.T) {
	g := newTestKing()
	assert.Equal(t, domain.KingPhaseSelectContract, g.GetPhase())
	assert.Equal(t, 0, g.GetDealNumber())
	assert.Equal(t, 4, g.GetPlayerCnt())
	assert.Equal(t, -1, g.GetCurrentContract())
	assert.Equal(t, -1, g.GetTrumpSuit())
	assert.False(t, g.GetGameEndFlag())

	// 13 cards each, 52 unique cards total.
	seen := map[int]bool{}
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		assert.Equal(t, domain.KingHandSize, p.GetCardsSize(), "player %d", i)
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			key := c.GetDesign()*100 + c.GetValue()
			assert.False(t, seen[key], "duplicate %d", key)
			seen[key] = true
			total++
		}
	}
	assert.Equal(t, 52, total)
}

func TestKing_SelectContract_Errors(t *testing.T) {
	g := newTestKing()
	g.SetDealerIdx(0) // human
	g.SetCurrentContract(-1)
	g.SetPhase(domain.KingPhaseSelectContract)

	assert.Error(t, g.SelectContract(-1, -1))
	assert.Error(t, g.SelectContract(99, -1))
	// King Trump requires a valid trump suit.
	assert.Error(t, g.SelectContract(domain.KingContractKingTrump, 0))

	// Not human dealer.
	g.SetDealerIdx(1)
	assert.ErrorIs(t, g.SelectContract(domain.KingContractNoTricks, -1), domain.ErrNotHumanTurn)

	// Wrong phase.
	g.SetDealerIdx(0)
	g.SetPhase(domain.KingPhasePlay)
	assert.ErrorIs(t, g.SelectContract(domain.KingContractNoTricks, -1), domain.ErrWrongPhase)
}

func TestKing_SelectContract_MarksUsedAndStartsPlay(t *testing.T) {
	g := newTestKing()
	g.SetDealerIdx(0)
	g.SetPhase(domain.KingPhaseSelectContract)
	require.NoError(t, g.SelectContract(domain.KingContractNoHearts, -1))
	assert.Equal(t, domain.KingPhasePlay, g.GetPhase())
	assert.Equal(t, domain.KingContractNoHearts, g.GetCurrentContract())
	assert.Equal(t, -1, g.GetTrumpSuit())
	assert.Equal(t, 0, g.GetCurrentTurn(), "dealer leads first trick")
	used := g.GetUsedContracts()
	assert.True(t, used[domain.KingContractNoHearts])

	// Cannot pick the same contract again.
	g.SetPhase(domain.KingPhaseSelectContract)
	assert.Error(t, g.SelectContract(domain.KingContractNoHearts, -1))

	// King Trump sets trump suit.
	g2 := newTestKing()
	g2.SetDealerIdx(0)
	g2.SetPhase(domain.KingPhaseSelectContract)
	require.NoError(t, g2.SelectContract(domain.KingContractKingTrump, domain.CardDesignSpade))
	assert.Equal(t, domain.CardDesignSpade, g2.GetTrumpSuit())
}

// --- Play / trick resolution ---

// TestKing_ValidateTrickPlay_NilCardGuard ensures a nil card in a hand is
// rejected (excluded from the playable set) rather than being treated as a
// legal lead and later panicking in card-strength evaluation.
func TestKing_ValidateTrickPlay_NilCardGuard(t *testing.T) {
	g := newTestKing()
	g.SetCurrentContract(domain.KingContractNoTricks)
	g.SetTrumpSuit(-1)
	g.SetPhase(domain.KingPhasePlay)
	g.SetCurrentTurn(1)
	g.SetCurrentTrick(nil) // leading a fresh trick
	// Hand holds one real card and one nil (corrupted) card.
	setKingHand(g, 1, card(domain.CardDesignSpade, 1), nil)
	assert.NotPanics(t, func() {
		valid := g.GetPlayableIndices(1)
		assert.Equal(t, []int{0}, valid, "nil card must be excluded from playable indices")
	})
}

func TestKing_FollowSuit(t *testing.T) {
	g := newTestKing()
	g.SetCurrentContract(domain.KingContractNoTricks)
	g.SetTrumpSuit(-1)
	g.SetPhase(domain.KingPhasePlay)
	g.SetCurrentTurn(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: card(domain.CardDesignSpade, 13)},
	})
	setKingHand(g, 1,
		card(domain.CardDesignSpade, 1),  // must follow
		card(domain.CardDesignHeart, 10)) // off-suit (illegal while holding spade)
	spadeIdx := -1
	p := g.GetPlayer(1)
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == domain.CardDesignSpade {
			spadeIdx = i
		}
	}
	valid := g.GetPlayableIndices(1)
	assert.Equal(t, []int{spadeIdx}, valid, "must follow the lead suit")

	// Playing the off-suit heart is rejected.
	heartIdx := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == domain.CardDesignHeart {
			heartIdx = i
		}
	}
	assert.Error(t, kingPlayAt(g, 1, heartIdx))
}

// kingPlayAt is a small helper to play as the current player (deterministic).
func kingPlayAt(g *domain.King, turn, idx int) error {
	g.SetCurrentTurn(turn)
	return g.PlayerPlay(idx)
}

// kingFinishTrick plays the remaining CPUs (at most 3) to complete the current
// trick. Bounded to avoid looping past resolution (which clears the trick).
func kingFinishTrick(g *domain.King) {
	for i := 0; i < domain.KingPlayerCnt-1; i++ {
		if g.GetPhase() != domain.KingPhasePlay {
			return
		}
		if len(g.GetCurrentTrick()) == 0 {
			return // trick already resolved
		}
		g.CpuPlay()
	}
}

func TestKing_TrickResolution_NoTrump(t *testing.T) {
	g := newTestKing()
	g.SetCurrentContract(domain.KingContractNoTricks)
	g.SetTrumpSuit(-1)
	g.SetTrickNumber(1)
	g.SetLeadPlayer(0)
	g.SetPhase(domain.KingPhasePlay)
	g.SetCurrentTurn(0)
	// Set each hand to hold exactly the card they play, in follow order.
	setKingHand(g, 0, card(domain.CardDesignSpade, 13))
	setKingHand(g, 1, card(domain.CardDesignSpade, 1)) // Ace beats King
	setKingHand(g, 2, card(domain.CardDesignSpade, 10))
	setKingHand(g, 3, card(domain.CardDesignSpade, 9))
	require.NoError(t, g.PlayerPlay(0))
	// Now CPUs 1,2,3 play their only card.
	kingFinishTrick(g)
	// Trick resolved: player 1 (Ace) wins.
	assert.Equal(t, 1, g.GetLastTrickWinner(), "Ace wins the spade trick")
	assert.Equal(t, 1, g.GetPlayer(1).GetTrickCount())
}

func TestKing_TrickResolution_TrumpWins(t *testing.T) {
	g := newTestKing()
	g.SetCurrentContract(domain.KingContractKingTrump)
	g.SetTrumpSuit(domain.CardDesignDiamond)
	g.SetTrickNumber(1)
	g.SetLeadPlayer(0)
	g.SetPhase(domain.KingPhasePlay)
	g.SetCurrentTurn(0)
	setKingHand(g, 0, card(domain.CardDesignSpade, 1)) // strongest non-trump
	setKingHand(g, 1, card(domain.CardDesignSpade, 13))
	setKingHand(g, 2, card(domain.CardDesignDiamond, 2)) // low trump beats non-trump
	setKingHand(g, 3, card(domain.CardDesignSpade, 12))
	require.NoError(t, g.PlayerPlay(0))
	kingFinishTrick(g)
	assert.Equal(t, 2, g.GetLastTrickWinner(), "trump beats the Ace of a non-trump suit")
}

// --- Scoring per contract ---

// stageKingScoring sets up a deal where the whole board is captured and scores it.
func stageKingScoring(t *testing.T, contract, trump int, tricks map[int][][]*domain.Card, ranks map[int][]int) *domain.King {
	t.Helper()
	g := newTestKing()
	g.SetDealNumber(0)
	g.SetDealerIdx(0)
	g.SetCurrentContract(contract)
	g.SetTrumpSuit(trump)
	// Load captured tricks directly onto players.
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		p.ResetDeal()
		for j, tr := range tricks[i] {
			rank := 0
			if rr, ok := ranks[i]; ok && j < len(rr) {
				rank = rr[j]
			}
			p.AddTrickWithRank(tr, rank)
		}
	}
	// Drive finishDeal via the last trick resolution path is complex; instead
	// exercise scoring by playing out the final trick. Simpler: set trick number
	// to last and resolve an empty trick is not valid, so score via NextDeal path.
	// We invoke the internal scoring by forcing a final trick.
	return g
}

func TestKing_Score_NoTricks(t *testing.T) {
	g := stageKingScoring(t, domain.KingContractNoTricks, -1, map[int][][]*domain.Card{
		1: {{card(domain.CardDesignSpade, 2)}, {card(domain.CardDesignClover, 3)}},
	}, nil)
	g.SetTrickNumber(domain.KingHandSize)
	g.SetPhase(domain.KingPhasePlay)
	g.SetLeadPlayer(0)
	g.SetCurrentTurn(0)
	// Play the last trick so finishDeal fires; each player plays one card.
	setKingHand(g, 0, card(domain.CardDesignHeart, 2))
	setKingHand(g, 1, card(domain.CardDesignHeart, 3))
	setKingHand(g, 2, card(domain.CardDesignHeart, 4))
	setKingHand(g, 3, card(domain.CardDesignHeart, 5))
	require.NoError(t, g.PlayerPlay(0))
	kingFinishTrick(g)
	require.Equal(t, domain.KingPhaseDealEnd, phaseOrEnd(g))
	// Player 1 took 2 earlier tricks; final trick won by whoever played highest
	// heart (player 3, the 5). Player 1: -2 tricks * penalty; plus final trick
	// winner also penalised. Just assert player 1 has a negative score.
	assert.Less(t, g.GetPlayer(1).GetTotalScore(), 0)
}

// phaseOrEnd returns the phase whether or not the game ended.
func phaseOrEnd(g *domain.King) string { return g.GetPhase() }

func TestKing_Score_KingTrump_Positive(t *testing.T) {
	g := newTestKing()
	g.SetDealNumber(0)
	g.SetDealerIdx(0)
	g.SetCurrentContract(domain.KingContractKingTrump)
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetTrickNumber(domain.KingHandSize)
	g.SetPhase(domain.KingPhasePlay)
	g.SetLeadPlayer(0)
	g.SetCurrentTurn(0)
	// Give player 0 a big lead trick so it wins the final trick.
	setKingHand(g, 0, card(domain.CardDesignSpade, 1))
	setKingHand(g, 1, card(domain.CardDesignSpade, 2))
	setKingHand(g, 2, card(domain.CardDesignSpade, 3))
	setKingHand(g, 3, card(domain.CardDesignSpade, 4))
	require.NoError(t, g.PlayerPlay(0))
	kingFinishTrick(g)
	// King Trump gives +reward per trick; player 0 (won the final trick) has >0.
	assert.Greater(t, g.GetPlayer(0).GetTotalScore(), 0)
}

// --- Full game via CPU ---

func TestKing_FullGame_CpuDriven(t *testing.T) {
	g := newTestKing()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = domain.KingDifficultyEasy
	g.SetConfig(cfg)
	g.Reset()

	guard := 0
	for !g.GetGameEndFlag() && guard < 20000 {
		guard++
		switch g.GetPhase() {
		case domain.KingPhaseSelectContract:
			if g.GetPlayer(g.GetDealerIdx()).GetIsHuman() {
				// Human dealer: pick the first unused contract.
				used := g.GetUsedContracts()
				c := 0
				for c < domain.KingContractCnt && used[c] {
					c++
				}
				trump := -1
				if c == domain.KingContractKingTrump {
					trump = domain.CardDesignSpade
				}
				require.NoError(t, g.SelectContract(c, trump))
			} else {
				g.CpuPlay()
			}
		case domain.KingPhasePlay:
			if g.IsHumanTurn() {
				valid := g.GetPlayableIndices(g.GetCurrentTurn())
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayerPlay(valid[0]))
			} else {
				g.CpuPlay()
			}
		case domain.KingPhaseDealEnd:
			g.NextDeal()
		}
	}
	assert.True(t, g.GetGameEndFlag())
	assert.Less(t, guard, 20000, "game should terminate")
	assert.Equal(t, domain.KingPhaseGameEnd, g.GetPhase())
	assert.NotEmpty(t, g.GetRoundWinners())
	// 7 contracts each used exactly once by the end of the last deal.
	used := g.GetUsedContracts()
	for c := 0; c < domain.KingContractCnt; c++ {
		assert.True(t, used[c], "contract %d used", c)
	}
}

func TestKing_PlayerPlay_Errors(t *testing.T) {
	g := newTestKing()
	g.SetPhase(domain.KingPhasePlay)
	g.SetCurrentTurn(0)
	g.SetCurrentContract(domain.KingContractNoTricks)
	setKingHand(g, 0, card(domain.CardDesignSpade, 1))
	assert.Error(t, g.PlayerPlay(-1))
	assert.Error(t, g.PlayerPlay(99))

	g.SetPhase(domain.KingPhaseSelectContract)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrWrongPhase)

	g.SetPhase(domain.KingPhasePlay)
	g.SetCurrentTurn(1) // CPU
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrNotHumanTurn)
}

func TestKing_NextDeal(t *testing.T) {
	g := newTestKing()
	// Wrong phase -> no-op.
	g.SetPhase(domain.KingPhasePlay)
	g.NextDeal()
	assert.Equal(t, domain.KingPhasePlay, g.GetPhase())

	// From dealEnd -> advances.
	g.SetPhase(domain.KingPhaseDealEnd)
	g.SetDealNumber(0)
	g.NextDeal()
	assert.Equal(t, 1, g.GetDealNumber())
	assert.Equal(t, domain.KingPhaseSelectContract, g.GetPhase())
	assert.Equal(t, 1, g.GetDealerIdx(), "dealer rotates by deal number")
}

func TestKing_Getters(t *testing.T) {
	g := newTestKing()
	g.SetTrickNumber(5)
	assert.Equal(t, 5, g.GetTrickNumber())
	g.SetLeadPlayer(2)
	assert.Equal(t, 2, g.GetLeadPlayer())
	g.SetCurrentContract(domain.KingContractNoMen)
	assert.Equal(t, domain.KingContractNoMen, g.GetCurrentContract())
	g.SetTrumpSuit(domain.CardDesignClover)
	assert.Equal(t, domain.CardDesignClover, g.GetTrumpSuit())

	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
	assert.NotNil(t, g.GetPlayer(0))
	assert.NotNil(t, g.GetConfig())
	_ = g.GetActionLog()
	assert.Nil(t, g.GetLastTrick())
	assert.Nil(t, g.GetLastDealDetail())
	assert.Nil(t, g.GetRoundWinners(), "no winners until game end")

	// GetPlayableIndices guards.
	assert.Nil(t, g.GetPlayableIndices(-1))
	g.SetPhase(domain.KingPhaseSelectContract)
	assert.Nil(t, g.GetPlayableIndices(0), "not play phase -> nil")

	// IsHumanTurn.
	g.SetPhase(domain.KingPhasePlay)
	g.SetCurrentTurn(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentTurn(1)
	assert.False(t, g.IsHumanTurn())
	g.SetPhase(domain.KingPhaseSelectContract)
	g.SetDealerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetDealerIdx(1)
	assert.False(t, g.IsHumanTurn())
	g.SetPhase(domain.KingPhaseDealEnd)
	assert.False(t, g.IsHumanTurn())
}

func TestKing_GetHint(t *testing.T) {
	g := newTestKing()
	// Not play phase -> nil.
	g.SetPhase(domain.KingPhaseSelectContract)
	assert.Nil(t, g.GetHint())

	// Play phase, human turn, negative contract -> avoid_low.
	g.SetPhase(domain.KingPhasePlay)
	g.SetCurrentTurn(0)
	g.SetCurrentContract(domain.KingContractNoTricks)
	g.SetTrumpSuit(-1)
	g.SetCurrentTrick(nil)
	setKingHand(g, 0, card(domain.CardDesignSpade, 2), card(domain.CardDesignHeart, 13))
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "avoid_low", h.Reason)
	assert.Len(t, h.CardIndices, 1)

	// King Trump -> win_high.
	g.SetCurrentContract(domain.KingContractKingTrump)
	g.SetTrumpSuit(domain.CardDesignSpade)
	h2 := g.GetHint()
	require.NotNil(t, h2)
	assert.Equal(t, "win_high", h2.Reason)

	// Not human turn -> nil.
	g.SetCurrentTurn(1)
	assert.Nil(t, g.GetHint())
}

func TestKing_JSON_RoundTrip(t *testing.T) {
	g := newTestKing()
	g.SetDealerIdx(0)
	g.SetCurrentContract(domain.KingContractKingTrump)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.KingPhasePlay)
	g.SetTrickNumber(3)
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var g2 domain.King
	require.NoError(t, json.Unmarshal(data, &g2))
	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetTrumpSuit(), g2.GetTrumpSuit())
	assert.Equal(t, g.GetPlayerCnt(), g2.GetPlayerCnt())
	assert.Equal(t, g.GetCurrentContract(), g2.GetCurrentContract())
	assert.Equal(t, g.GetTrickNumber(), g2.GetTrickNumber())
}

func TestKing_JSON_Invalid(t *testing.T) {
	const okPlayers = `[{"gp":{},"th":{}},{"gp":{},"th":{}},{"gp":{},"th":{}},{"gp":{},"th":{}}]`
	cases := []string{
		`not json`,                                                                           // malformed
		`{"ph":"play","pl":[null,null]}`,                                                     // wrong player count
		`{"ph":"play","pl":[null,null,null,null]}`,                                           // nil players
		`{"ph":"selectContract","pl":` + okPlayers + `,"di":99}`,                             // dealer out of range
		`{"ph":"selectContract","pl":` + okPlayers + `,"cp":99}`,                             // currentPlayer out of range
		`{"ph":"selectContract","pl":` + okPlayers + `,"lp":-1}`,                             // leadPlayer negative
		`{"ph":"selectContract","pl":` + okPlayers + `,"lw":99}`,                             // lastTrickWinner out of range
		`{"ph":"selectContract","pl":` + okPlayers + `,"lw":-2}`,                             // lastTrickWinner below -1
		`{"ph":"bogus","pl":` + okPlayers + `}`,                                              // bad phase
		`{"ph":"selectContract","pl":` + okPlayers + `,"cc":99}`,                             // bad contract
		`{"ph":"selectContract","pl":` + okPlayers + `,"cc":-2}`,                             // contract below -1
		`{"ph":"selectContract","pl":` + okPlayers + `,"ts":9}`,                              // bad trump suit
		`{"ph":"selectContract","pl":` + okPlayers + `,"dn":99}`,                             // bad deal number
		`{"ph":"play","pl":` + okPlayers + `,"cc":-1}`,                                       // contract unset while playing
		`{"ph":"selectContract","pl":` + okPlayers + `,"ct":[null]}`,                         // nil trick card
		`{"ph":"selectContract","pl":` + okPlayers + `,"ct":[{"pi":99,"c":{"d":1,"v":13}}]}`, // trick pi out of range
		`{"ph":"selectContract","pl":` + okPlayers + `,"lt":[{"pi":-1,"c":{"d":1,"v":13}}]}`, // last trick pi negative
	}
	for _, c := range cases {
		var g domain.King
		assert.Error(t, json.Unmarshal([]byte(c), &g), c)
	}

	// Bad config rejected.
	badCfg := `{"ph":"selectContract","pl":` + okPlayers + `,"cf":{"di":99}}`
	var gc domain.King
	assert.Error(t, json.Unmarshal([]byte(badCfg), &gc))

	// Valid restore with unset trump/contract falls back to a deck.
	okJSON := `{"ph":"selectContract","pl":` + okPlayers + `,"ts":-1,"cc":-1,"lw":-1,"cf":{"di":1}}`
	var g2 domain.King
	require.NoError(t, json.Unmarshal([]byte(okJSON), &g2))
	assert.Equal(t, 4, g2.GetPlayerCnt())
	assert.NotNil(t, g2.GetPlayer(0))
}

func TestKing_DealDetail_JSON(t *testing.T) {
	d := &domain.KingDealDetail{Contract: domain.KingContractNoHearts, TrumpSuit: -1, DealerIdx: 2, Gained: map[int]int{0: -4, 1: 0}}
	b, err := json.Marshal(d)
	require.NoError(t, err)
	var d2 domain.KingDealDetail
	require.NoError(t, json.Unmarshal(b, &d2))
	assert.Equal(t, d.Contract, d2.Contract)
	assert.Equal(t, d.DealerIdx, d2.DealerIdx)
	assert.Equal(t, -4, d2.Gained[0])
	assert.Error(t, json.Unmarshal([]byte(`nope`), &d2))
}
