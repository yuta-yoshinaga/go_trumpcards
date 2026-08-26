//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newTestQuadrille returns a fresh, reset Quadrille game with the default 3-player setup.
func newTestQuadrille() *domain.Quadrille {
	g := domain.NewDefaultQuadrille()
	g.Reset()
	return g
}

// setQuadrilleHand replaces player i's hand with the supplied cards deterministically.
func setQuadrilleHand(g *domain.Quadrille, i int, cards ...*domain.Card) {
	p := g.GetPlayer(i)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// quadrilleCard is a shorthand constructor for a face-up card.
func quadrilleCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// quadrilleGiveTricks awards player i exactly n tricks by adding dummy trick groups.
func quadrilleGiveTricks(g *domain.Quadrille, i, n int) {
	p := g.GetPlayer(i)
	for k := 0; k < n; k++ {
		p.AddTrick([]*domain.Card{quadrilleCard(domain.CardDesignSpade, 2)})
	}
}

func TestQuadrille_ResetDeal(t *testing.T) {
	g := newTestQuadrille()
	assert.Equal(t, domain.QuadrillePhaseBid, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, domain.QuadrillePlayerCnt, g.GetPlayerCnt())
	assert.Equal(t, -1, g.GetQuadrilleIdx())
	assert.Equal(t, -1, g.GetTrumpSuit())
	assert.Equal(t, domain.QuadrilleBidNone, g.GetWinningBid())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerPlayer())

	totalHand := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		totalHand += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, domain.QuadrilleHandSize*domain.QuadrillePlayerCnt, totalHand)

	assert.Equal(t, (g.GetDealerIdx()+1)%domain.QuadrillePlayerCnt, g.GetForehandIdx())
	assert.Equal(t, g.GetForehandIdx(), g.GetCurrentBidderIdx())
}

func TestQuadrille_DeckIsUnique40(t *testing.T) {
	g := newTestQuadrille()
	seen := map[int]bool{}
	count := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			key := c.GetDesign()*100 + c.GetValue()
			assert.False(t, seen[key], "duplicate card %d", key)
			seen[key] = true
			count++
		}
	}
	// **40 枚を 4 人で配り切る。** クローン元のオンブルは 3 人 × 9 枚で
	// 13 枚を配り残していたが、こちらは山に 1 枚も残らない。
	// 手で書いた枚数ではなく**デッキと席数から数える**ので、どちらかを
	// 変えれば追随する。
	assert.Equal(t, domain.QuadrilleHandSize*domain.QuadrillePlayerCnt, count)
	assert.Equal(t, domain.QuadrilleDeckSize, count, "山に 1 枚も残らないこと")
	valid := map[int]bool{1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 11: true, 12: true, 13: true}
	for k := range seen {
		assert.True(t, valid[k%100], "unexpected rank %d", k%100)
	}
}

func TestQuadrille_Bidding_WrongPhaseAndSuitRequired(t *testing.T) {
	g := newTestQuadrille()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = domain.QuadrilleCpuDifficultyEasy // CPUs always pass
	g.SetConfig(cfg)
	g.SetPhase(domain.QuadrillePhaseBid)

	// Drive CPU bids until it is the human's (seat 0) turn.
	guard := 0
	for g.GetPhase() == domain.QuadrillePhaseBid && g.GetCurrentBidderIdx() != 0 && guard < 10 {
		guard++
		g.CpuBid()
	}
	if g.GetPhase() == domain.QuadrillePhaseBid && g.GetCurrentBidderIdx() == 0 {
		// Entrar without a valid trump suit is rejected.
		assert.Error(t, g.PlayerBid(domain.QuadrilleBidEntrar, -1))
		// Entrar with a valid trump suit succeeds.
		require.NoError(t, g.PlayerBid(domain.QuadrilleBidEntrar, domain.CardDesignHeart))
	}

	// Wrong phase -> error.
	g.SetPhase(domain.QuadrillePhasePlay)
	assert.ErrorIs(t, g.PlayerBid(domain.QuadrilleBidEntrar, domain.CardDesignHeart), domain.ErrWrongPhase)
}

func TestQuadrille_Bidding_EveryonePasses_DealerForced(t *testing.T) {
	g := newTestQuadrille()
	cfg := g.GetConfig()
	cfg.CpuDifficulty = domain.QuadrilleCpuDifficultyEasy // CPUs always pass
	g.SetConfig(cfg)
	g.SetPhase(domain.QuadrillePhaseBid)

	guard := 0
	for g.GetPhase() == domain.QuadrillePhaseBid && guard < 50 {
		guard++
		if g.GetPlayer(g.GetCurrentBidderIdx()).GetIsHuman() {
			require.NoError(t, g.PlayerBid(domain.QuadrilleBidNone, -1)) // human passes
		} else {
			g.CpuBid()
		}
	}
	// Auction resolved -> dealer forced to be Quadrille with a chosen trump.
	assert.Equal(t, g.GetDealerIdx(), g.GetQuadrilleIdx())
	assert.GreaterOrEqual(t, int(g.GetWinningBid()), int(domain.QuadrilleBidEntrar))
	assert.True(t, g.GetTrumpSuit() >= domain.CardDesignSpade && g.GetTrumpSuit() <= domain.CardDesignDiamond)
	// **落札の次はプレイではなく王呼び。** オンブルはここで直接プレイに
	// 入っていたが、こちらは味方を呼ぶまで盤面が進まない。落札者が王を
	// 4 枚とも持っていれば呼ぶ相手がいないので Roi seul でプレイへ飛ぶ。
	if g.IsRoiSeul() {
		assert.Equal(t, domain.QuadrillePhasePlay, g.GetPhase())
	} else {
		assert.Equal(t, domain.QuadrillePhaseKingCall, g.GetPhase())
		assert.NotEmpty(t, g.GetCallableKingSuits(), "呼べる王が提示されること")
	}
}

func TestQuadrille_BidOrdering(t *testing.T) {
	assert.Greater(t, int(domain.QuadrilleBidSolo), int(domain.QuadrilleBidEntrar))
	assert.Greater(t, int(domain.QuadrilleBidEntrar), int(domain.QuadrilleBidNone))
}

func TestQuadrille_MatadorRanking(t *testing.T) {
	resolve := func(trump int, trick []*domain.TrickCard) int {
		g := newTestQuadrille()
		g.SetQuadrilleIdx(0)
		g.SetTrumpSuit(trump)
		g.SetTrickNumber(1) // not the last trick, avoid triggering round-end scoring
		g.SetPhase(domain.QuadrillePhaseTrickEnd)
		g.SetCurrentTrick(trick)
		g.ResolveTrick()
		return g.GetLeadPlayerIdx()
	}

	// Spadille (♠A) beats Manille (7 of trump) and Basto (♣A) even when spades isn't trump.
	assert.Equal(t, 1, resolve(domain.CardDesignHeart, []*domain.TrickCard{
		{PlayerIdx: 0, Card: quadrilleCard(domain.CardDesignHeart, 7)},   // Manille
		{PlayerIdx: 1, Card: quadrilleCard(domain.CardDesignSpade, 1)},   // Spadille (highest)
		{PlayerIdx: 2, Card: quadrilleCard(domain.CardDesignClover, 1)},  // Basto
		{PlayerIdx: 3, Card: quadrilleCard(domain.CardDesignDiamond, 2)}, // 平札の最低位 (勝者を動かさない)
	}))

	// Manille (7 of trump) beats Basto, Punto and the trump King.
	assert.Equal(t, 2, resolve(domain.CardDesignHeart, []*domain.TrickCard{
		{PlayerIdx: 0, Card: quadrilleCard(domain.CardDesignHeart, 1)},   // Punto
		{PlayerIdx: 1, Card: quadrilleCard(domain.CardDesignHeart, 13)},  // trump K
		{PlayerIdx: 2, Card: quadrilleCard(domain.CardDesignHeart, 7)},   // Manille (highest)
		{PlayerIdx: 3, Card: quadrilleCard(domain.CardDesignDiamond, 2)}, // 平札の最低位 (勝者を動かさない)
	}))

	// Basto (♣A) beats Punto (red-trump Ace) and lower trumps.
	assert.Equal(t, 0, resolve(domain.CardDesignHeart, []*domain.TrickCard{
		{PlayerIdx: 0, Card: quadrilleCard(domain.CardDesignClover, 1)},  // Basto (highest here)
		{PlayerIdx: 1, Card: quadrilleCard(domain.CardDesignHeart, 1)},   // Punto
		{PlayerIdx: 2, Card: quadrilleCard(domain.CardDesignHeart, 13)},  // trump K
		{PlayerIdx: 3, Card: quadrilleCard(domain.CardDesignDiamond, 2)}, // 平札の最低位 (勝者を動かさない)
	}))

	// Any trump beats any plain card (low trump 2 beats plain King).
	assert.Equal(t, 1, resolve(domain.CardDesignHeart, []*domain.TrickCard{
		{PlayerIdx: 0, Card: quadrilleCard(domain.CardDesignSpade, 13)},  // plain K (lead)
		{PlayerIdx: 1, Card: quadrilleCard(domain.CardDesignHeart, 2)},   // trump 2
		{PlayerIdx: 2, Card: quadrilleCard(domain.CardDesignSpade, 12)},  // plain Q
		{PlayerIdx: 3, Card: quadrilleCard(domain.CardDesignDiamond, 2)}, // 平札の最低位 (勝者を動かさない)
	}))

	// No Punto for a BLACK trump: ♠A is Spadille, trump-suit K ranks below it.
	assert.Equal(t, 0, resolve(domain.CardDesignSpade, []*domain.TrickCard{
		{PlayerIdx: 0, Card: quadrilleCard(domain.CardDesignSpade, 1)},   // Spadille
		{PlayerIdx: 1, Card: quadrilleCard(domain.CardDesignSpade, 13)},  // trump K
		{PlayerIdx: 2, Card: quadrilleCard(domain.CardDesignSpade, 2)},   // trump 2
		{PlayerIdx: 3, Card: quadrilleCard(domain.CardDesignDiamond, 2)}, // 平札の最低位 (勝者を動かさない)
	}))
}

func TestQuadrille_PlainSuit_AceLowAndOffSuit(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(0)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetTrickNumber(1)
	g.SetPhase(domain.QuadrillePhaseTrickEnd)
	// Plain red suit (diamonds): K>Q>J>A>2>3>4>5>6>7 — the Ace is the
	// 4th-highest (outranking 2..7), NOT low as in the black plain suits.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: quadrilleCard(domain.CardDesignDiamond, 2)}, // 2
		{PlayerIdx: 1, Card: quadrilleCard(domain.CardDesignDiamond, 1)}, // A beats 2 and 7
		{PlayerIdx: 2, Card: quadrilleCard(domain.CardDesignDiamond, 7)}, // 7 (lowest red plain)
		{PlayerIdx: 3, Card: quadrilleCard(domain.CardDesignClover, 7)},  // 別スートの平札 (勝てない)
	})
	g.ResolveTrick()
	assert.Equal(t, 1, g.GetLeadPlayerIdx(), "red plain Ace outranks 2 and 7")

	// **次のトリックへ進んでから 2 つめを組む。** ResolveTrick が冪等になった
	// ので、精算済みのまま札を差し替えても二度目は何もしない (#6230)。
	g.NextTrick()

	// Off-suit plain cannot win; a higher follower of the led suit does.
	// Red ranking: diamond 3 outranks diamond 7.
	g.SetTrickNumber(1)
	g.SetPhase(domain.QuadrillePhaseTrickEnd)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: quadrilleCard(domain.CardDesignDiamond, 7)}, // lead 7 (lowest red)
		{PlayerIdx: 1, Card: quadrilleCard(domain.CardDesignClover, 13)}, // off-suit K (cannot win, ♣K is plain)
		{PlayerIdx: 2, Card: quadrilleCard(domain.CardDesignDiamond, 3)}, // follows, higher (red 3 > red 7)
		{PlayerIdx: 3, Card: quadrilleCard(domain.CardDesignClover, 7)},  // 別スートの平札 (勝てない)
	})
	g.ResolveTrick()
	assert.Equal(t, 2, g.GetLeadPlayerIdx(), "off-suit plain cannot win")
}

func TestQuadrille_MustFollow_TrumpGroupIsASuit(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(0)
	g.SetTrumpSuit(domain.CardDesignDiamond)
	g.SetPhase(domain.QuadrillePhasePlay)
	g.SetCurrentPlayerIdx(1)
	// Trump (diamond) is led -> ♠A counts as trump and must be followed.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: quadrilleCard(domain.CardDesignDiamond, 5)},
	})
	setQuadrilleHand(g, 1,
		quadrilleCard(domain.CardDesignSpade, 1), // ♠A is trump -> must follow
		quadrilleCard(domain.CardDesignHeart, 3)) // plain heart (illegal while holding trump)
	valid := g.GetPlayableIndices(1)
	assert.Equal(t, []int{0}, valid, "must follow trump group with ♠A")

	// Void in trump -> any card playable.
	setQuadrilleHand(g, 1,
		quadrilleCard(domain.CardDesignClover, 13),
		quadrilleCard(domain.CardDesignHeart, 2))
	assert.Len(t, g.GetPlayableIndices(1), 2, "void in trump: all cards playable")
}

func TestQuadrille_NextTrick(t *testing.T) {
	g := newTestQuadrille()
	g.SetPhase(domain.QuadrillePhaseTrickEnd)
	g.SetLeadPlayerIdx(2)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 0, Card: quadrilleCard(domain.CardDesignSpade, 1)}})
	g.NextTrick()
	assert.Equal(t, domain.QuadrillePhasePlay, g.GetPhase())
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	assert.Equal(t, 2, g.GetTrickNumber())
	assert.Empty(t, g.GetCurrentTrick())

	// Wrong phase is a no-op.
	g.SetPhase(domain.QuadrillePhasePlay)
	g.NextTrick()
	assert.Equal(t, domain.QuadrillePhasePlay, g.GetPhase())
}

func TestQuadrille_Outcomes(t *testing.T) {
	// **勝敗は席ではなく「側」で決まる。** 落札者 + 呼ばれた王の持ち主 が
	// 一方、残り 2 席がもう一方。クローン元のオンブルは 3 人卓で味方が
	// いないので「落札者 1 人 対 他席の最大」で足りたが、その形のままだと
	// 味方の取ったトリックが相手側に計上されて勝敗が反転する。
	//
	// 10 トリックを 4 席に配る。partner は席 2 (落札者 0 の味方)。
	cases := []struct {
		name    string
		tricks  [domain.QuadrillePlayerCnt]int
		outcome domain.QuadrilleOutcome
		scores  [domain.QuadrillePlayerCnt]int
	}{
		// 側の合計 6 対 4 で勝ち。**落札者ひとりでは 3 しか取っておらず、
		// 相手の最大 (席 1 の 3) を上回らない** —— 席で見る実装はここを落とす。
		{
			"sacar via the partner", [domain.QuadrillePlayerCnt]int{3, 3, 3, 1},
			domain.QuadrilleOutcomeSacar,
			[domain.QuadrillePlayerCnt]int{2, -1, 2, -1},
		},
		// 5 対 5 の引き分けは Puesta (軽い罰)。
		{
			"puesta on a split", [domain.QuadrillePlayerCnt]int{3, 3, 2, 2},
			domain.QuadrilleOutcomePuesta,
			[domain.QuadrillePlayerCnt]int{-2, 1, -2, 1},
		},
		// 相手側が上回れば Codille (倍の罰)。
		{
			"codille", [domain.QuadrillePlayerCnt]int{2, 4, 1, 3},
			domain.QuadrilleOutcomeCodille,
			[domain.QuadrillePlayerCnt]int{-4, 2, -4, 2},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := newTestQuadrille()
			g.SetQuadrilleIdx(0)
			g.SetPartnerForTest(2, true)
			g.SetWinningBid(domain.QuadrilleBidEntrar)
			g.SetPhase(domain.QuadrillePhaseRoundEnd)
			for i := 0; i < domain.QuadrillePlayerCnt; i++ {
				quadrilleGiveTricks(g, i, c.tricks[i])
			}
			g.ScoreRound()
			assert.Equal(t, c.outcome, g.GetOutcome())
			assert.Equal(t, c.scores, g.GetPlayerScores())

			// ScoreRound is idempotent (scored flag).
			g.ScoreRound()
			assert.Equal(t, c.scores, g.GetPlayerScores())
		})
	}
}

// TestQuadrille_RoiSeulScoresAlone は単独プレイの配分を見る。
//
// 味方がいないので落札者 1 席 対 3 席。味方の席をそのまま使うと、
// 単独のはずが 2 席で分け合う形になる。
func TestQuadrille_RoiSeulScoresAlone(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(0)
	g.SetPartnerForTest(2, true) // 前のディールの残骸を模す
	g.SetRoiSeulForTest(true)
	g.SetWinningBid(domain.QuadrilleBidEntrar)
	g.SetPhase(domain.QuadrillePhaseRoundEnd)
	// 落札者が 6、相手 3 席で 4。単独でも 6 > 4 で勝ち。
	quadrilleGiveTricks(g, 0, 6)
	quadrilleGiveTricks(g, 1, 2)
	quadrilleGiveTricks(g, 2, 1)
	quadrilleGiveTricks(g, 3, 1)
	g.ScoreRound()

	assert.Equal(t, domain.QuadrilleOutcomeSacar, g.GetOutcome())
	// **席 2 は相手側。** 単独なのに味方として扱うと +2 が入ってしまう。
	assert.Equal(t, [domain.QuadrillePlayerCnt]int{2, -1, -1, -1}, g.GetPlayerScores())
}

func TestQuadrille_ScoreRound_WrongPhaseNoop(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(0)
	g.SetPhase(domain.QuadrillePhasePlay)
	g.ScoreRound()
	assert.Equal(t, [domain.QuadrillePlayerCnt]int{0, 0, 0}, g.GetPlayerScores())
}

func TestQuadrille_GameEnd_HumanWins(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(0)
	g.SetWinningBid(domain.QuadrilleBidSolo)
	g.SetRoundNumber(domain.QuadrilleWinRounds) // final deal
	g.SetPhase(domain.QuadrillePhaseRoundEnd)
	quadrilleGiveTricks(g, 0, 9) // human (Quadrille) sweeps -> Sacar
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerPlayer())
	assert.Equal(t, domain.QuadrillePhaseGameEnd, g.GetPhase())
	assert.Equal(t, domain.QuadrilleResultWin, g.GetResult())
}

func TestQuadrille_GameEnd_HumanLoses(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(1) // a CPU is Quadrille
	g.SetWinningBid(domain.QuadrilleBidSolo)
	g.SetRoundNumber(domain.QuadrilleWinRounds)
	g.SetPhase(domain.QuadrillePhaseRoundEnd)
	g.SetPlayerScores([domain.QuadrillePlayerCnt]int{0, 5, 0})
	quadrilleGiveTricks(g, 1, 9) // CPU1 sweeps -> Sacar for CPU1
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 1, g.GetWinnerPlayer())
	assert.Equal(t, domain.QuadrilleResultLose, g.GetResult())
}

func TestQuadrille_NextRound(t *testing.T) {
	g := newTestQuadrille()
	g.SetPhase(domain.QuadrillePhaseRoundEnd)
	prevDealer := g.GetDealerIdx()
	prevRound := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, prevRound+1, g.GetRoundNumber())
	assert.Equal(t, (prevDealer+1)%domain.QuadrillePlayerCnt, g.GetDealerIdx())
	assert.Equal(t, domain.QuadrillePhaseBid, g.GetPhase())

	// Wrong phase -> no-op.
	g.SetPhase(domain.QuadrillePhasePlay)
	r := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, r, g.GetRoundNumber())
}

func TestQuadrille_PlayerPlay_Errors(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(0)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.QuadrillePhasePlay)
	g.SetCurrentPlayerIdx(0)
	setQuadrilleHand(g, 0, quadrilleCard(domain.CardDesignSpade, 13))

	assert.Error(t, g.PlayerPlay(-1))
	assert.Error(t, g.PlayerPlay(99))

	// Wrong phase.
	g.SetPhase(domain.QuadrillePhaseBid)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrWrongPhase)

	// Not human turn.
	g.SetPhase(domain.QuadrillePhasePlay)
	g.SetCurrentPlayerIdx(1)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrNotHumanTurn)
}

func TestQuadrille_PlayerPlay_FollowSuitViolation(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(0)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.QuadrillePhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: quadrilleCard(domain.CardDesignDiamond, 13)}, // plain diamond lead
	})
	setQuadrilleHand(g, 0,
		quadrilleCard(domain.CardDesignDiamond, 5), // legal (follows diamond)
		quadrilleCard(domain.CardDesignClover, 3))  // illegal (must follow diamond)
	cloverIdx := -1
	p := g.GetPlayer(0)
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == domain.CardDesignClover {
			cloverIdx = i
		}
	}
	assert.ErrorIs(t, g.PlayerPlay(cloverIdx), domain.ErrInvalidPlay)
}

func TestQuadrille_PlayerPlay_SuccessAndTrickComplete(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(0)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.QuadrillePhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: quadrilleCard(domain.CardDesignSpade, 5)},
		{PlayerIdx: 2, Card: quadrilleCard(domain.CardDesignSpade, 6)},
		{PlayerIdx: 3, Card: quadrilleCard(domain.CardDesignSpade, 4)},
	})
	setQuadrilleHand(g, 0, quadrilleCard(domain.CardDesignSpade, 13))
	require.NoError(t, g.PlayerPlay(0))
	// **4 人卓なので 4 枚目でトリックが閉じる。** 3 枚で閉じていたのは
	// クローン元の 3 人卓の前提。
	assert.Equal(t, domain.QuadrillePhaseTrickEnd, g.GetPhase(), "fourth card completes the trick")
}

func TestQuadrille_GetHint_AllPhases(t *testing.T) {
	// Bid phase hint.
	g := newTestQuadrille()
	g.SetPhase(domain.QuadrillePhaseBid)
	guard := 0
	for g.GetPhase() == domain.QuadrillePhaseBid && g.GetCurrentBidderIdx() != 0 && guard < 10 {
		guard++
		g.CpuBid()
	}
	if g.GetPhase() == domain.QuadrillePhaseBid && g.GetCurrentBidderIdx() == 0 {
		h := g.GetHint()
		require.NotNil(t, h)
		assert.Contains(t, []string{"bid_entrar", "bid_solo", "bid_pass"}, h.Reason)
	}

	// Play phase lead hint (human is Quadrille -> lead_high).
	g2 := newTestQuadrille()
	g2.SetPhase(domain.QuadrillePhasePlay)
	g2.SetQuadrilleIdx(0)
	g2.SetTrumpSuit(domain.CardDesignHeart)
	g2.SetCurrentPlayerIdx(0)
	g2.SetCurrentTrick(nil)
	setQuadrilleHand(g2, 0, quadrilleCard(domain.CardDesignHeart, 13), quadrilleCard(domain.CardDesignSpade, 4))
	ph := g2.GetHint()
	require.NotNil(t, ph)
	assert.Equal(t, "lead_high", ph.Reason)
	assert.Len(t, ph.CardIndices, 1)

	// Play phase lead hint (human is coalition -> lead_low).
	g2.SetQuadrilleIdx(1)
	g2.SetCurrentPlayerIdx(0)
	g2.SetCurrentTrick(nil)
	lh := g2.GetHint()
	require.NotNil(t, lh)
	assert.Equal(t, "lead_low", lh.Reason)

	// Play phase, not human's turn -> nil.
	g2.SetCurrentPlayerIdx(1)
	assert.Nil(t, g2.GetHint())

	// Unhandled phase -> nil.
	g2.SetPhase(domain.QuadrillePhaseTrickEnd)
	assert.Nil(t, g2.GetHint())
}

func TestQuadrille_GetHint_PlayReasons(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(2) // player 0 is coalition
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.QuadrillePhasePlay)
	g.SetCurrentPlayerIdx(0)
	// Opponent (Quadrille seat 2) leads a plain diamond Queen; player 0 can win or duck.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 2, Card: quadrilleCard(domain.CardDesignDiamond, 12)},
	})
	setQuadrilleHand(g, 0,
		quadrilleCard(domain.CardDesignDiamond, 13), // K beats Q -> follow_win
		quadrilleCard(domain.CardDesignDiamond, 4))  // 4 loses -> follow_duck
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Contains(t, []string{"follow_win", "follow_duck"}, h.Reason)

	// discard_low: void in lead suit.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 2, Card: quadrilleCard(domain.CardDesignDiamond, 12)},
	})
	setQuadrilleHand(g, 0, quadrilleCard(domain.CardDesignClover, 13)) // off-suit only
	h2 := g.GetHint()
	require.NotNil(t, h2)
	assert.Equal(t, "discard_low", h2.Reason)

	// give_partner: partner (same side) is winning.
	g.SetQuadrilleIdx(1) // players 0 and 2 are coalition (same side)
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 2, Card: quadrilleCard(domain.CardDesignDiamond, 13)}, // partner winning with K
	})
	setQuadrilleHand(g, 0,
		quadrilleCard(domain.CardDesignDiamond, 11), // J loses to K
		quadrilleCard(domain.CardDesignDiamond, 7))  // 7 loses to K
	h3 := g.GetHint()
	require.NotNil(t, h3)
	assert.Equal(t, "give_partner", h3.Reason)
}

func TestQuadrille_CpuFullRound(t *testing.T) {
	g := newTestQuadrille()
	guard := 0
	for !g.GetGameEndFlag() && guard < 8000 {
		guard++
		switch g.GetPhase() {
		case domain.QuadrillePhaseBid:
			if g.GetPlayer(g.GetCurrentBidderIdx()).GetIsHuman() {
				require.NoError(t, g.PlayerBid(domain.QuadrilleBidNone, -1)) // human passes
			} else {
				g.CpuBid()
			}
		case domain.QuadrillePhaseKingCall:
			// **王を呼ぶまで盤面は進まない。** ここが無いとループは
			// 王呼びフェーズで空回りし、guard を撃ち尽くす。
			if g.IsHumanKingCallTurn() {
				callable := g.GetCallableKingSuits()
				require.NotEmpty(t, callable, "呼べる王が提示されること")
				require.NoError(t, g.DeclareKing(g.GetQuadrilleIdx(), callable[0]))
			} else {
				g.CpuDeclareKing()
			}
		case domain.QuadrillePhasePlay:
			if g.IsHumanTurn() {
				valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayerPlay(valid[0]))
			} else {
				g.CpuPlay()
			}
		case domain.QuadrillePhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == domain.QuadrillePhaseTrickEnd {
				g.NextTrick()
			}
		case domain.QuadrillePhaseRoundEnd:
			g.ScoreRound()
			if !g.GetGameEndFlag() {
				g.NextRound()
			}
		case domain.QuadrillePhaseGameEnd:
			guard = 8000
		}
	}
	assert.Less(t, guard, 8000, "game flow should progress")
}

func TestQuadrille_Getters(t *testing.T) {
	g := newTestQuadrille()
	g.SetRoundNumber(4)
	assert.Equal(t, 4, g.GetRoundNumber())
	g.SetTrickNumber(3)
	assert.Equal(t, 3, g.GetTrickNumber())
	g.SetCurrentPlayerIdx(2)
	assert.Equal(t, 2, g.GetCurrentPlayerIdx())
	g.SetLeadPlayerIdx(1)
	assert.Equal(t, 1, g.GetLeadPlayerIdx())
	g.SetWinningBid(domain.QuadrilleBidSolo)
	assert.Equal(t, domain.QuadrilleBidSolo, g.GetWinningBid())
	g.SetQuadrilleIdx(2)
	assert.Equal(t, 2, g.GetQuadrilleIdx())
	g.SetTrumpSuit(domain.CardDesignClover)
	assert.Equal(t, domain.CardDesignClover, g.GetTrumpSuit())
	g.SetPlayerScores([domain.QuadrillePlayerCnt]int{10, 20, 30})
	assert.Equal(t, [domain.QuadrillePlayerCnt]int{10, 20, 30}, g.GetPlayerScores())

	assert.GreaterOrEqual(t, g.GetForehandIdx(), 0)
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
	assert.NotNil(t, g.GetPlayer(0))
	assert.NotNil(t, g.GetConfig())
	assert.Equal(t, domain.QuadrilleOutcomeNone, g.GetOutcome())
	assert.Equal(t, domain.QuadrilleResultNone, g.GetResult())
	_ = g.GetActionLog()

	assert.Nil(t, g.GetPlayableIndices(-1))
	g.SetPhase(domain.QuadrillePhaseBid)
	assert.Nil(t, g.GetPlayableIndices(0), "not play phase -> nil")

	g.SetPhase(domain.QuadrillePhasePlay)
	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanTurn())
	g.SetCurrentPlayerIdx(1)
	assert.False(t, g.IsHumanTurn())
	g.SetPhase(domain.QuadrillePhaseBid)
	assert.Equal(t, g.GetPlayer(g.GetCurrentBidderIdx()).GetIsHuman(), g.IsHumanBidTurn())
	g.SetPhase(domain.QuadrillePhasePlay)
	assert.False(t, g.IsHumanBidTurn())
}

func TestQuadrille_JSON_RoundTrip(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(0)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetWinningBid(domain.QuadrilleBidSolo)
	g.SetPhase(domain.QuadrillePhasePlay)
	g.SetLeadPlayerIdx(0)
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var g2 domain.Quadrille
	require.NoError(t, json.Unmarshal(data, &g2))
	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetPlayerCnt(), g2.GetPlayerCnt())
	assert.Equal(t, g.GetQuadrilleIdx(), g2.GetQuadrilleIdx())
	assert.Equal(t, g.GetTrumpSuit(), g2.GetTrumpSuit())
	assert.Equal(t, g.GetWinningBid(), g2.GetWinningBid())
}

func TestQuadrille_JSON_Invalid(t *testing.T) {
	const okPlayers = `[{"gp":{},"th":{}},{"gp":{},"th":{}},{"gp":{},"th":{}},{"gp":{},"th":{}}]`
	cases := []string{
		`not json`,
		`{"ph":0,"ps":[null,null,null],"ts":-1}`,                                      // wrong player count (3 != 4)
		`{"ph":0,"ps":[null,null,null,null],"ts":-1}`,                                 // nil players
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"ci":100}`,                            // ci out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"ci":-1}`,                             // ci negative
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"di":99}`,                             // di out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"fh":99}`,                             // fh out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"cbi":99}`,                            // cbi out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"li":99}`,                             // li out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"li":-2}`,                             // li below -1
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"om":99}`,                             // quadrilleIdx out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"lt":99}`,                             // lastTrickWinner out of range
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"wp":99}`,                             // winnerPlayer out of range
		`{"ph":99,"ps":` + okPlayers + `,"ts":-1}`,                                    // bad phase
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"wb":9}`,                              // bad winning bid
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"bd":[9,0,0]}`,                        // bad bid element
		`{"ph":0,"ps":` + okPlayers + `,"ts":9}`,                                      // bad trump suit
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"bt":[9,0,0]}`,                        // bad bidTrump element
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"oc":9}`,                              // bad outcome
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"rs":9}`,                              // bad result
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"ct":[null]}`,                         // nil trick card
		`{"ph":0,"ps":` + okPlayers + `,"ts":-1,"ct":[{"pi":99,"c":{"d":1,"v":13}}]}`, // trick card idx out of range
		`{"ph":1,"ps":` + okPlayers + `,"om":-1,"li":0,"ts":3}`,                       // play phase requires quadrille set
		`{"ph":1,"ps":` + okPlayers + `,"om":0,"li":-1,"ts":3}`,                       // play phase requires lead set
		`{"ph":1,"ps":` + okPlayers + `,"om":0,"li":0,"ts":-1}`,                       // play phase requires trump set
	}
	for _, c := range cases {
		var g domain.Quadrille
		assert.Error(t, json.Unmarshal([]byte(c), &g), c)
	}

	// Config validation failure (bad CPU difficulty) is rejected.
	badCfg := `{"ph":0,"ps":` + okPlayers + `,"ts":-1,"cf":{"cd":99,"tr":5}}`
	var gc domain.Quadrille
	assert.Error(t, json.Unmarshal([]byte(badCfg), &gc))

	// Valid restore.
	okJSON := `{"ph":0,"ps":` + okPlayers + `,"wb":0,"cf":{"cd":1,"tr":5},"lt":-1,"wp":-1,"li":-1,"om":-1,"ts":-1}`
	var g2 domain.Quadrille
	assert.NoError(t, json.Unmarshal([]byte(okJSON), &g2))
	assert.Equal(t, domain.QuadrillePlayerCnt, g2.GetPlayerCnt())
	assert.NotNil(t, g2.GetPlayer(0))
}

func TestQuadrillePlayer_JSON_And_ResetRound(t *testing.T) {
	p := domain.NewQuadrillePlayer(true)
	p.AddCard(quadrilleCard(domain.CardDesignSpade, 1))
	p.AddTrick([]*domain.Card{quadrilleCard(domain.CardDesignHeart, 13)})
	assert.Equal(t, 1, p.GetTrickCount())

	b, err := json.Marshal(p)
	require.NoError(t, err)
	var p2 domain.QuadrillePlayer
	require.NoError(t, json.Unmarshal(b, &p2))
	assert.True(t, p2.GetIsHuman())
	assert.Equal(t, 1, p2.GetCardsSize())
	assert.Equal(t, 1, p2.GetTrickCount())

	p2.ResetRound()
	assert.Equal(t, 0, p2.GetCardsSize())
	assert.Equal(t, 0, p2.GetTrickCount())
	assert.False(t, p2.GetIsFinished())

	assert.Error(t, json.Unmarshal([]byte(`not json`), &p2))
	var p3 domain.QuadrillePlayer
	require.NoError(t, json.Unmarshal([]byte(`{}`), &p3))
	assert.False(t, p3.GetIsHuman())
}

func TestQuadrilleConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultQuadrilleConfig().Validate())
	assert.Equal(t, domain.QuadrilleWinRounds, domain.DefaultQuadrilleConfig().TargetRounds)
	assert.Equal(t, domain.QuadrilleCpuDifficultyNormal, domain.DefaultQuadrilleConfig().CpuDifficulty)

	assert.Error(t, domain.QuadrilleConfig{CpuDifficulty: 99, TargetRounds: 5}.Validate())
	assert.Error(t, domain.QuadrilleConfig{CpuDifficulty: domain.QuadrilleCpuDifficultyEasy, TargetRounds: 0}.Validate())
}

// **マタドールの判定は序列そのものから引く。**別に条件を書くと、序列を
// 変えたときに表示だけ古いままになる (#4919)。
func TestQuadrille_MatadorRank(t *testing.T) {
	spade, clover, heart := domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart
	card := func(d, v int) *domain.Card { return domain.NewCard(d, v, false) }

	// 赤 (♥) が切り札のとき。
	assert.Equal(t, 1, domain.QuadrilleMatadorRank(card(spade, 1), heart), "spadille")
	assert.Equal(t, 2, domain.QuadrilleMatadorRank(card(heart, 7), heart), "manille")
	assert.Equal(t, 3, domain.QuadrilleMatadorRank(card(clover, 1), heart), "basto")
	// 切り札の A は Punto であってマタドールではない。
	assert.Equal(t, 0, domain.QuadrilleMatadorRank(card(heart, 1), heart))
	assert.Equal(t, 0, domain.QuadrilleMatadorRank(card(heart, 13), heart))
	// 他スートの 7 はただの平札。
	assert.Equal(t, 0, domain.QuadrilleMatadorRank(card(spade, 7), heart))

	// **マニーユは切り札スート次第。**♠ が切り札なら ♠7 がマニーユ。
	assert.Equal(t, 2, domain.QuadrilleMatadorRank(card(spade, 7), spade))
	assert.Equal(t, 0, domain.QuadrilleMatadorRank(card(heart, 7), spade))
	// ♠A は切り札が何であれスパディーユ。
	assert.Equal(t, 1, domain.QuadrilleMatadorRank(card(spade, 1), spade))

	// **切り札未確定なら 0。**マニーユだけ決まらないと不揃いな案内になる。
	for _, trump := range []int{-1, 0, 5} {
		assert.Equal(t, 0, domain.QuadrilleMatadorRank(card(spade, 1), trump), "trump %d", trump)
		assert.Equal(t, 0, domain.QuadrilleMatadorRank(card(clover, 1), trump), "trump %d", trump)
	}

	assert.Equal(t, 0, domain.QuadrilleMatadorRank(nil, heart))
}

// TestQuadrille_KingCall_RejectsAKingYouHold は、**自分が持っている王は
// 呼べない**ことを見る。
//
// 呼べてしまうと味方が増えず、単独プレイが「4 人卓の 1 対 3」ではなく
// 黙って成立する ── 盤面はそれを何も知らせない。
func TestQuadrille_KingCall_RejectsAKingYouHold(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(0)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.QuadrillePhaseKingCall)
	setQuadrilleHand(g, 0,
		quadrilleCard(domain.CardDesignSpade, 13), // ♠K を持っている
		quadrilleCard(domain.CardDesignHeart, 5))
	setQuadrilleHand(g, 2, quadrilleCard(domain.CardDesignHeart, 13)) // ♥K は席 2

	assert.Error(t, g.DeclareKing(0, domain.CardDesignSpade), "持っている王は呼べない")
	assert.Equal(t, domain.QuadrillePhaseKingCall, g.GetPhase(), "弾いたらフェーズは進まない")
	assert.NotContains(t, g.GetCallableKingSuits(), domain.CardDesignSpade,
		"持っている王は選択肢にも出さない")

	// **負のコントロール。** 持っていない王は呼べる —— でなければ上の Error は
	// 「何を渡しても弾く」だけを見ていることになる。
	require.NoError(t, g.DeclareKing(0, domain.CardDesignHeart))
	assert.Equal(t, domain.QuadrillePhasePlay, g.GetPhase(), "呼んだらプレイへ進む")
	assert.Equal(t, domain.CardDesignHeart, g.GetCalledKingSuit())
}

// TestQuadrille_PartnerStaysHiddenUntilTheKingIsPlayed は、**誰が味方かは
// 呼ばれた王が場に出るまで伏せられる**ことを見る。
//
// 呼び声は卓で聞こえるので王のスート自体は公開情報だが、持ち主が最初から
// 分かってしまうとこのゲームの緊張感は無くなる。
func TestQuadrille_PartnerStaysHiddenUntilTheKingIsPlayed(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(0)
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetPhase(domain.QuadrillePhaseKingCall)
	setQuadrilleHand(g, 0, quadrilleCard(domain.CardDesignSpade, 5))
	setQuadrilleHand(g, 1, quadrilleCard(domain.CardDesignClover, 5))
	setQuadrilleHand(g, 2, quadrilleCard(domain.CardDesignHeart, 13)) // ♥K の持ち主
	setQuadrilleHand(g, 3, quadrilleCard(domain.CardDesignDiamond, 5))

	require.NoError(t, g.DeclareKing(0, domain.CardDesignHeart))
	// **呼んだ直後は伏せられている。**
	assert.Equal(t, -1, g.GetPartnerIdx(), "王が出る前に相方が漏れている")
	assert.Equal(t, domain.CardDesignHeart, g.GetCalledKingSuit(), "王のスートは公開情報")

	// 席 2 が ♥K を出すと公開される。手札はその 1 枚だけなので必ず出る。
	g.SetPhase(domain.QuadrillePhasePlay)
	g.SetCurrentPlayerIdx(2)
	g.SetCurrentTrick(nil)
	g.CpuPlay()
	require.Len(t, g.GetCurrentTrick(), 1, "席 2 が 1 枚出したこと")
	assert.Equal(t, 2, g.GetPartnerIdx(), "王が出たら相方が分かること")
}

// TestQuadrille_RoiSeulWhenHoldingEveryKing は、王を 4 枚とも持っていたら
// 王呼びを飛ばして単独プレイになることを見る。
func TestQuadrille_RoiSeulWhenHoldingEveryKing(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(0)
	g.SetTrumpSuit(domain.CardDesignSpade)
	setQuadrilleHand(g, 0,
		quadrilleCard(domain.CardDesignSpade, 13),
		quadrilleCard(domain.CardDesignClover, 13),
		quadrilleCard(domain.CardDesignHeart, 13),
		quadrilleCard(domain.CardDesignDiamond, 13))

	assert.Empty(t, g.GetCallableKingSuitsForTest(0), "呼べる王が無いこと")

	// **負のコントロール。** 1 枚手放せば呼べる王が現れる。
	setQuadrilleHand(g, 0,
		quadrilleCard(domain.CardDesignSpade, 13),
		quadrilleCard(domain.CardDesignClover, 13),
		quadrilleCard(domain.CardDesignHeart, 13))
	assert.Equal(t, []int{domain.CardDesignDiamond}, g.GetCallableKingSuitsForTest(0))
}

// TestQuadrille_KingCallSurvivesTheWire は、**味方が確定した後の**盤面を
// 往復させる。
//
// 王呼びが盤面から落ちると、復元後の partnerIdx はゼロ値の **0** になり、
// 席 0 が誰の味方でもないのに落札者側として集計される。呼ばれた王も
// 単独プレイの区別も消えるので、勝敗が静かに変わる。
//
// **ゼロ値と食い違う値で往復させる。** partnerIdx の「未確定」は -1 なので、
// -1 のまま往復させてもフィールドを丸ごと消した実装が同じ答えを返す。
func TestQuadrille_KingCallSurvivesTheWire(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(1)
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetPhase(domain.QuadrillePhasePlay)
	g.SetCalledKingSuitForTest(domain.CardDesignHeart)
	g.SetPartnerForTest(3, true) // 席 3、公開済み — どちらもゼロ値ではない

	require.NotEqual(t, 0, g.GetPartnerIdx(), "ゼロ値と食い違う席であること")
	require.NotEqual(t, 0, g.GetCalledKingSuit(), "ゼロ値と食い違うスートであること")

	blob, err := json.Marshal(g)
	require.NoError(t, err)
	var got domain.Quadrille
	require.NoError(t, json.Unmarshal(blob, &got))

	assert.Equal(t, domain.CardDesignHeart, got.GetCalledKingSuit(), "呼ばれた王が往復で消えた")
	assert.Equal(t, 3, got.GetPartnerIdx(), "味方が往復で消えた")
	assert.False(t, got.IsRoiSeul())

	// **伏せた状態も往復すること。** 公開フラグが落ちると、復元しただけで
	// 相方が画面に漏れる (あるいは公開済みの相方が消える)。
	h := newTestQuadrille()
	h.SetQuadrilleIdx(1)
	h.SetCalledKingSuitForTest(domain.CardDesignClover)
	h.SetPartnerForTest(3, false)
	blob2, err := json.Marshal(h)
	require.NoError(t, err)
	var got2 domain.Quadrille
	require.NoError(t, json.Unmarshal(blob2, &got2))
	assert.Equal(t, -1, got2.GetPartnerIdx(), "伏せたままのはずの相方が漏れている")

	// 単独プレイも往復すること。
	s := newTestQuadrille()
	s.SetQuadrilleIdx(2)
	s.SetRoiSeulForTest(true)
	blob3, err := json.Marshal(s)
	require.NoError(t, err)
	var got3 domain.Quadrille
	require.NoError(t, json.Unmarshal(blob3, &got3))
	assert.True(t, got3.IsRoiSeul(), "単独プレイが往復で消えた")
}

// TestQuadrille_RestoreNormalisesAnAbsentKingCall は、**王呼びのフィールドが
// 無い JSON** から復元しても、席 0 が勝手に味方にならないことを見る。
//
// Go は無いフィールドに 0 を入れるので、素直に信じると partnerIdx=0 に
// なり、**席 0 が誰の味方でもないのに落札者側として集計される**。
// calledKingSuit=0 もどの札とも一致しないので、味方が永久に公開されない。
func TestQuadrille_RestoreNormalisesAnAbsentKingCall(t *testing.T) {
	const okPlayers = `[{"gp":{},"th":{}},{"gp":{},"th":{}},{"gp":{},"th":{}},{"gp":{},"th":{}}]`
	// 王呼びのフィールド (ck / pi / pr / rq) を一切持たない盤面。
	blob := `{"ph":0,"ps":` + okPlayers + `,"wb":0,"cf":{"cd":1,"tr":5},"lt":-1,"wp":-1,"li":-1,"om":-1,"ts":-1}`

	var g domain.Quadrille
	require.NoError(t, json.Unmarshal([]byte(blob), &g))

	assert.Equal(t, -1, g.GetCalledKingSuit(), "呼ばれていない王が 0 として残っている")
	assert.Equal(t, -1, g.GetPartnerIdx(), "席 0 が勝手に味方になっている")
	assert.False(t, g.IsRoiSeul())

	// **負のコントロール。** ちゃんと書いてあれば素通しする。
	withKing := `{"ph":2,"ps":` + okPlayers + `,"wb":1,"cf":{"cd":1,"tr":5},"lt":-1,"wp":-1,"li":0,"om":1,"ts":1,` +
		`"ck":3,"pi":3,"pr":true}`
	var h domain.Quadrille
	require.NoError(t, json.Unmarshal([]byte(withKing), &h))
	assert.Equal(t, domain.CardDesignHeart, h.GetCalledKingSuit())
	assert.Equal(t, 3, h.GetPartnerIdx())

	// 単独プレイなら味方は消える (前のディールの残骸を引きずらない)。
	roi := `{"ph":2,"ps":` + okPlayers + `,"wb":1,"cf":{"cd":1,"tr":5},"lt":-1,"wp":-1,"li":0,"om":1,"ts":1,` +
		`"ck":3,"pi":3,"pr":true,"rq":true}`
	var s domain.Quadrille
	require.NoError(t, json.Unmarshal([]byte(roi), &s))
	assert.True(t, s.IsRoiSeul())
	assert.Equal(t, -1, s.GetPartnerIdx(), "単独プレイに味方がいる")
}

// **40 枚を 4 人で配り切ったなら、10 トリック打って手札が尽きる。**
// クローン元のオンブルは 3 人 × 9 枚でストックを 13 枚残す設計で、そちらの
// TrickCount = 9 を引き継いだまま HandSize だけ 10 に直されていたため、
// 毎ディール全員の手札が 1 枚残ったままラウンドが締まっていた (#6230)。
func TestQuadrille_PlaysOutEveryDealtCard(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(0)
	g.SetWinningBid(domain.QuadrilleBidSolo)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetRoiSeulForTest(true)
	g.SetPhase(domain.QuadrillePhasePlay)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentPlayerIdx(0)
	g.SetTrickNumber(1)

	dealt := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		dealt += g.GetPlayer(i).GetCardsSize()
	}
	require.Equal(t, domain.QuadrilleHandSize*domain.QuadrillePlayerCnt, dealt)

	for step := 0; step < 400 && g.GetPhase() != domain.QuadrillePhaseRoundEnd; step++ {
		switch g.GetPhase() {
		case domain.QuadrillePhasePlay:
			if g.IsHumanTurn() {
				valid := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, valid)
				require.NoError(t, g.PlayerPlay(valid[0]))
				continue
			}
			g.CpuPlay()
		case domain.QuadrillePhaseTrickEnd:
			g.ResolveTrick()
			g.NextTrick()
		}
	}
	require.Equal(t, domain.QuadrillePhaseRoundEnd, g.GetPhase(), "ラウンドが締まること")

	left := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		left += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 0, left, "配った札は全部出る (手札が残らない)")

	quad, coalition := g.GetSideTrickCounts()
	assert.Equal(t, domain.QuadrilleHandSize, quad+coalition, "トリック数は配り札の枚数と一致する")
}

// **呼ばれた王を持つ席は落札者の味方。** sameSide が「落札者かどうか」だけを
// 見ていたため、味方の席では partnerWinning が常に false になり、CPU は自分の
// 味方が勝っているトリックを上から取りに行っていた (#6230)。
func TestQuadrille_SameSideCountsTheCalledKingPartner(t *testing.T) {
	g := newTestQuadrille()
	g.SetQuadrilleIdx(0)
	g.SetRoiSeulForTest(false)
	g.SetPartnerForTest(2, true)

	assert.True(t, g.SameSideForTest(0, 2), "落札者と相方は同じ側")
	assert.True(t, g.SameSideForTest(2, 0), "対称であること")
	assert.True(t, g.SameSideForTest(1, 3), "連合の 2 席は同じ側")
	assert.False(t, g.SameSideForTest(0, 1), "落札者と連合は別の側")
	assert.False(t, g.SameSideForTest(2, 3), "相方と連合は別の側")

	// 負のコントロール: 単独プレイなら相方はいない。
	g.SetRoiSeulForTest(true)
	assert.False(t, g.SameSideForTest(0, 2), "roi seul では相方の席も敵")
}

// **同じトリックを二度精算しない。** ResolveTrick はトリック終了フェーズが
// 続く間なら何度でも呼べてしまうので、二度目は何もしないこと (#6230)。
func TestQuadrille_ResolveTrickIsIdempotent(t *testing.T) {
	g := newTestQuadrille()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetQuadrilleIdx(0)
	g.SetTrickNumber(1)
	g.SetPhase(domain.QuadrillePhaseTrickEnd)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: quadrilleCard(domain.CardDesignHeart, 13)},
		{PlayerIdx: 1, Card: quadrilleCard(domain.CardDesignHeart, 7)},
		{PlayerIdx: 2, Card: quadrilleCard(domain.CardDesignHeart, 6)},
		{PlayerIdx: 3, Card: quadrilleCard(domain.CardDesignDiamond, 6)},
	})
	g.ResolveTrick()
	g.ResolveTrick()
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += g.GetPlayer(i).GetTrickCount()
	}
	assert.Equal(t, 1, total, "二度呼んでもトリックは 1 つ")
}

// **引き分け (Puesta) が起こりうること。** 側の取り分は必ず合計 TrickCount に
// なるので、トリック数が奇数だと同数に並ぶことが構造的に不可能になる。
// 9 トリックで出荷されていた間、Puesta は 300 ディール測って 0 回だった ——
// ドキュメントに載っている 3 つの結果のうち 1 つが起き得なかった。
// 5 対 5 を組んで Puesta を確かめる既存のテストは、**ゲームが決して作れない
// 配分**を手で組んでいたので、それに気づけなかった (#6230)。
func TestQuadrille_PuestaNeedsAnEvenTrickCount(t *testing.T) {
	assert.Equal(t, 0, domain.QuadrilleTrickCount%2,
		"総トリック数が奇数だと側が同数に並べず、Puesta が一度も起きない")
}

// **精算済みかどうかは保存して読み戻しても残る。** Worker は毎リクエストで
// 盤を復元するので、このフラグを JSON に載せ忘れると、復元のたびに
// 「まだ精算していない」状態に戻り、冪等性がリクエストごとに消える ——
// つまり NextTrick を挟まずに 2 リクエスト来たら勝者に札束が二度積まれる。
// フィールドを見比べるのではなく、**復元した盤で実際に呼んで**確かめる (#6230)。
func TestQuadrille_ResolveTrickStaysIdempotentAcrossSaveRestore(t *testing.T) {
	g := newTestQuadrille()
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetQuadrilleIdx(0)
	g.SetTrickNumber(1)
	g.SetPhase(domain.QuadrillePhaseTrickEnd)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: quadrilleCard(domain.CardDesignHeart, 13)},
		{PlayerIdx: 1, Card: quadrilleCard(domain.CardDesignHeart, 7)},
		{PlayerIdx: 2, Card: quadrilleCard(domain.CardDesignHeart, 6)},
		{PlayerIdx: 3, Card: quadrilleCard(domain.CardDesignDiamond, 6)},
	})
	g.ResolveTrick()

	data, err := json.Marshal(g)
	require.NoError(t, err)
	var restored domain.Quadrille
	require.NoError(t, json.Unmarshal(data, &restored))

	restored.ResolveTrick()
	total := 0
	for i := 0; i < restored.GetPlayerCnt(); i++ {
		total += restored.GetPlayer(i).GetTrickCount()
	}
	assert.Equal(t, 1, total, "復元後に呼び直してもトリックは 1 つ")
}
