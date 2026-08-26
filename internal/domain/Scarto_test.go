//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// --- helpers (scarto-prefixed) ---

func scartoTrumpCard(v int) *domain.Card {
	return domain.NewCard(domain.ScartoTrumpDesign, v, false)
}

func scartoExcuseCard() *domain.Card {
	return domain.NewCard(domain.ScartoExcuseDesign, domain.ScartoExcuseValue, false)
}

func scartoSuitCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func scartoNewReset() *domain.Scarto {
	g := domain.NewDefaultScarto()
	g.Reset()
	return g
}

func scartoSetHand(g *domain.Scarto, idx int, cards ...*domain.Card) {
	p := g.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func scartoTrickCards(cards ...*domain.TrickCard) []*domain.TrickCard {
	return cards
}

// --- Deck ---

func TestScartoDeckIs78(t *testing.T) {
	deck := domain.BuildScartoDeckPublic()
	require.Len(t, deck, domain.ScartoDeckSize)
	suits, trumps, excuse := 0, 0, 0
	total := 0
	for _, c := range deck {
		total += domain.ScartoCardHalfPointsPublic(c)
		switch c.GetDesign() {
		case domain.ScartoExcuseDesign:
			excuse++
			assert.Equal(t, domain.ScartoExcuseValue, c.GetValue())
		case domain.ScartoTrumpDesign:
			trumps++
			assert.GreaterOrEqual(t, c.GetValue(), 1)
			assert.LessOrEqual(t, c.GetValue(), domain.ScartoMaxTrump)
		default:
			suits++
		}
	}
	assert.Equal(t, 56, suits)
	assert.Equal(t, 21, trumps)
	assert.Equal(t, 1, excuse)
	// Total = 91 points = 182 half-points.
	assert.Equal(t, 182, total)
}

func TestScartoDealDistribution(t *testing.T) {
	g := scartoNewReset()
	assert.Equal(t, domain.ScartoPhaseScarto, g.GetPhase())
	assert.Equal(t, 0, g.GetDealerIdx())
	// Dealer holds 25 + surplus (3) = 28; the other two hold 25.
	assert.Equal(t, domain.ScartoHandSize+domain.ScartoSurplus, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.ScartoHandSize, g.GetPlayer(1).GetCardsSize())
	assert.Equal(t, domain.ScartoHandSize, g.GetPlayer(2).GetCardsSize())
	// All 78 cards are distributed among the three hands.
	total := g.GetPlayer(0).GetCardsSize() + g.GetPlayer(1).GetCardsSize() + g.GetPlayer(2).GetCardsSize()
	assert.Equal(t, domain.ScartoDeckSize, total)
}

// --- Card classification / points ---

func TestScartoClassification(t *testing.T) {
	assert.True(t, domain.ScartoIsTrumpPublic(scartoTrumpCard(5)))
	assert.False(t, domain.ScartoIsTrumpPublic(scartoSuitCard(1, 5)))
	assert.True(t, domain.ScartoIsExcusePublic(scartoExcuseCard()))
	assert.False(t, domain.ScartoIsExcusePublic(scartoTrumpCard(1)))

	assert.True(t, domain.ScartoIsBoutPublic(scartoTrumpCard(1)))
	assert.True(t, domain.ScartoIsBoutPublic(scartoTrumpCard(21)))
	assert.True(t, domain.ScartoIsBoutPublic(scartoExcuseCard()))
	assert.False(t, domain.ScartoIsBoutPublic(scartoTrumpCard(10)))
	assert.False(t, domain.ScartoIsBoutPublic(scartoSuitCard(1, 14)))
}

func TestScartoHalfPoints(t *testing.T) {
	cases := []struct {
		card *domain.Card
		want int
	}{
		{scartoSuitCard(1, 14), 9}, // Roi
		{scartoSuitCard(1, 13), 7}, // Dame
		{scartoSuitCard(1, 12), 5}, // Cavalier
		{scartoSuitCard(1, 11), 3}, // Valet
		{scartoSuitCard(1, 5), 1},  // pip
		{scartoTrumpCard(1), 9},    // petit bout
		{scartoTrumpCard(21), 9},   // 21 bout
		{scartoExcuseCard(), 9},    // excuse bout
		{scartoTrumpCard(10), 1},   // plain trump
		{nil, 0},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, domain.ScartoCardHalfPointsPublic(c.card))
	}
}

func TestScartoDiscardable(t *testing.T) {
	assert.True(t, domain.ScartoDiscardablePublic(scartoSuitCard(1, 5)))   // pip
	assert.False(t, domain.ScartoDiscardablePublic(scartoSuitCard(1, 14))) // king
	assert.False(t, domain.ScartoDiscardablePublic(scartoSuitCard(1, 11))) // court
	assert.False(t, domain.ScartoDiscardablePublic(scartoTrumpCard(5)))    // trump
	assert.False(t, domain.ScartoDiscardablePublic(scartoExcuseCard()))    // excuse
	assert.False(t, domain.ScartoDiscardablePublic(nil))
}

// --- Settlement helper (pure) ---

func TestScartoSettleDealZeroSum(t *testing.T) {
	half := [domain.ScartoPlayerCnt]int{60, 60, 62}
	scores := domain.ScartoSettleDeal(half)
	// score_i = 3*half_i - total(182)
	assert.Equal(t, -2, scores[0])
	assert.Equal(t, -2, scores[1])
	assert.Equal(t, 4, scores[2])
	assert.Equal(t, 0, scores[0]+scores[1]+scores[2])
}

func TestScartoSettleDealAllEqual(t *testing.T) {
	// Everyone at the average -> zero deltas.
	half := [domain.ScartoPlayerCnt]int{40, 40, 40}
	scores := domain.ScartoSettleDeal(half)
	assert.Equal(t, [domain.ScartoPlayerCnt]int{0, 0, 0}, scores)
}

func TestScartoSettleDealFullDeckZeroSum(t *testing.T) {
	// A real distribution of the full 182 half-points is always zero-sum.
	half := [domain.ScartoPlayerCnt]int{100, 50, 32}
	scores := domain.ScartoSettleDeal(half)
	assert.Equal(t, 0, scores[0]+scores[1]+scores[2])
}

// --- Scarto (dealer discard) ---

func TestScartoDiscardValidation(t *testing.T) {
	g := scartoNewReset()
	g.SetDealerIdx(0)
	g.SetPhase(domain.ScartoPhaseScarto)
	scartoSetHand(g, 0,
		scartoSuitCard(1, 14), // 0 king (illegal)
		scartoExcuseCard(),    // 1 excuse (illegal)
		scartoSuitCard(1, 11), // 2 court (illegal)
		scartoTrumpCard(1),    // 3 bout (illegal)
		scartoSuitCard(1, 2),  // 4 pip
		scartoSuitCard(1, 3),  // 5 pip
		scartoSuitCard(1, 4),  // 6 pip
		scartoSuitCard(1, 5),  // 7 pip
	)
	// Wrong count.
	require.Error(t, g.PlayerScarto([]int{4, 5}))
	// King in discard.
	require.Error(t, g.PlayerScarto([]int{0, 4, 5}))
	// Excuse in discard.
	require.Error(t, g.PlayerScarto([]int{1, 4, 5}))
	// Court in discard.
	require.Error(t, g.PlayerScarto([]int{2, 4, 5}))
	// Bout (trump 1) in discard.
	require.Error(t, g.PlayerScarto([]int{3, 4, 5}))
	// Duplicate index.
	require.Error(t, g.PlayerScarto([]int{4, 4, 5}))
	// Legal discard of 3 pips.
	require.NoError(t, g.PlayerScarto([]int{4, 5, 6}))
	assert.Equal(t, domain.ScartoPhasePlay, g.GetPhase())
	assert.Equal(t, domain.ScartoSurplus, g.GetScartoCount())
	// The dealer keeps 25 cards (started at 8 here, minus 3).
	assert.Equal(t, 5, g.GetPlayer(0).GetCardsSize())
}

func TestScartoDiscardCountsToDealerPoints(t *testing.T) {
	g := scartoNewReset()
	g.SetDealerIdx(0)
	g.SetPhase(domain.ScartoPhaseScarto)
	scartoSetHand(g, 0,
		scartoSuitCard(1, 2),
		scartoSuitCard(1, 3),
		scartoSuitCard(1, 4),
		scartoSuitCard(1, 5),
	)
	require.NoError(t, g.PlayerScarto([]int{0, 1, 2}))
	// Three 0.5-point pips = 3 half-points credited to the dealer.
	assert.Equal(t, 3, g.GetCardPoints(0))
}

func TestScartoCpuScarto(t *testing.T) {
	g := scartoNewReset()
	g.SetDealerIdx(1) // CPU dealer
	g.SetPhase(domain.ScartoPhaseScarto)
	scartoSetHand(g, 1,
		scartoSuitCard(1, 14), // king (kept)
		scartoTrumpCard(1),    // bout (kept)
		scartoSuitCard(2, 2),
		scartoSuitCard(2, 3),
		scartoSuitCard(2, 4),
	)
	g.CpuScarto()
	assert.Equal(t, domain.ScartoPhasePlay, g.GetPhase())
	assert.Equal(t, domain.ScartoSurplus, g.GetScartoCount())
	// King and bout must be retained.
	kept := g.GetPlayer(1)
	sawKing, sawBout := false, false
	for i := 0; i < kept.GetCardsSize(); i++ {
		c := kept.GetCard(i)
		if !domain.ScartoIsTrumpPublic(c) && c.GetValue() == domain.ScartoKingValue {
			sawKing = true
		}
		if domain.ScartoIsBoutPublic(c) {
			sawBout = true
		}
	}
	assert.True(t, sawKing, "king should be kept")
	assert.True(t, sawBout, "bout should be kept")
}

// --- Follow / trump priority / overtrump ---

func TestScartoFollowSuit(t *testing.T) {
	g := scartoNewReset()
	g.SetPhase(domain.ScartoPhasePlay)
	scartoSetHand(g, 1,
		scartoSuitCard(domain.CardDesignHeart, 5),
		scartoSuitCard(domain.CardDesignHeart, 9),
		scartoSuitCard(domain.CardDesignSpade, 3),
		scartoTrumpCard(4),
	)
	g.SetCurrentTrick(scartoTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: scartoSuitCard(domain.CardDesignHeart, 7)},
	))
	g.SetCurrentPlayerIdx(1)
	valid := g.GetPlayableIndices(1)
	assert.ElementsMatch(t, []int{0, 1}, valid)
}

func TestScartoVoidMustTrump(t *testing.T) {
	g := scartoNewReset()
	g.SetPhase(domain.ScartoPhasePlay)
	scartoSetHand(g, 1,
		scartoSuitCard(domain.CardDesignSpade, 3),
		scartoTrumpCard(4),
		scartoTrumpCard(9),
	)
	g.SetCurrentTrick(scartoTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: scartoSuitCard(domain.CardDesignHeart, 7)},
	))
	g.SetCurrentPlayerIdx(1)
	valid := g.GetPlayableIndices(1)
	assert.ElementsMatch(t, []int{1, 2}, valid)
}

func TestScartoOvertrumpObligation(t *testing.T) {
	g := scartoNewReset()
	g.SetPhase(domain.ScartoPhasePlay)
	scartoSetHand(g, 2,
		scartoTrumpCard(5),
		scartoTrumpCard(15),
	)
	g.SetCurrentTrick(scartoTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: scartoSuitCard(domain.CardDesignHeart, 7)},
		&domain.TrickCard{PlayerIdx: 1, Card: scartoTrumpCard(10)},
	))
	g.SetCurrentPlayerIdx(2)
	valid := g.GetPlayableIndices(2)
	assert.ElementsMatch(t, []int{1}, valid)
}

func TestScartoCannotOvertrumpPlaysAnyTrump(t *testing.T) {
	g := scartoNewReset()
	g.SetPhase(domain.ScartoPhasePlay)
	scartoSetHand(g, 2,
		scartoTrumpCard(5),
		scartoTrumpCard(9),
	)
	g.SetCurrentTrick(scartoTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: scartoSuitCard(domain.CardDesignHeart, 7)},
		&domain.TrickCard{PlayerIdx: 1, Card: scartoTrumpCard(18)},
	))
	g.SetCurrentPlayerIdx(2)
	valid := g.GetPlayableIndices(2)
	assert.ElementsMatch(t, []int{0, 1}, valid)
}

func TestScartoExcuseAlwaysPlayable(t *testing.T) {
	g := scartoNewReset()
	g.SetPhase(domain.ScartoPhasePlay)
	scartoSetHand(g, 1,
		scartoSuitCard(domain.CardDesignHeart, 5),
		scartoExcuseCard(),
		scartoSuitCard(domain.CardDesignSpade, 3),
	)
	g.SetCurrentTrick(scartoTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: scartoSuitCard(domain.CardDesignHeart, 7)},
	))
	g.SetCurrentPlayerIdx(1)
	valid := g.GetPlayableIndices(1)
	assert.ElementsMatch(t, []int{0, 1}, valid) // heart + excuse (not the spade)
}

func TestScartoLeadAllValid(t *testing.T) {
	g := scartoNewReset()
	g.SetPhase(domain.ScartoPhasePlay)
	scartoSetHand(g, 0,
		scartoSuitCard(domain.CardDesignHeart, 5),
		scartoTrumpCard(4),
		scartoExcuseCard(),
	)
	g.SetCurrentTrick(nil)
	g.SetCurrentPlayerIdx(0)
	valid := g.GetPlayableIndices(0)
	assert.ElementsMatch(t, []int{0, 1, 2}, valid)
}

// --- Trick winner ---

func TestScartoTrickWinnerHighestTrump(t *testing.T) {
	g := scartoNewReset()
	g.SetCurrentTrick(scartoTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: scartoSuitCard(domain.CardDesignHeart, 14)},
		&domain.TrickCard{PlayerIdx: 1, Card: scartoTrumpCard(3)},
		&domain.TrickCard{PlayerIdx: 2, Card: scartoTrumpCard(9)},
	))
	assert.Equal(t, 2, g.TrickWinnerPublic())
}

func TestScartoTrickWinnerLedSuit(t *testing.T) {
	g := scartoNewReset()
	g.SetCurrentTrick(scartoTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: scartoSuitCard(domain.CardDesignHeart, 8)},
		&domain.TrickCard{PlayerIdx: 1, Card: scartoSuitCard(domain.CardDesignHeart, 14)},
		&domain.TrickCard{PlayerIdx: 2, Card: scartoSuitCard(domain.CardDesignSpade, 14)},
	))
	assert.Equal(t, 1, g.TrickWinnerPublic()) // heart 14
}

func TestScartoExcuseNeverWinsAndLedSuit(t *testing.T) {
	g := scartoNewReset()
	g.SetCurrentTrick(scartoTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: scartoExcuseCard()},
		&domain.TrickCard{PlayerIdx: 1, Card: scartoSuitCard(domain.CardDesignSpade, 5)},
		&domain.TrickCard{PlayerIdx: 2, Card: scartoSuitCard(domain.CardDesignSpade, 9)},
	))
	assert.Equal(t, domain.CardDesignSpade, g.LedSuitPublic())
	assert.Equal(t, 2, g.TrickWinnerPublic()) // spade 9, excuse never wins
}

// --- ResolveTrick / excuse ownership ---

func TestScartoResolveTrickExcuseKeptByOwner(t *testing.T) {
	g := scartoNewReset()
	g.SetTrickNumber(1)
	g.SetLeadPlayerIdx(0)
	g.SetPhase(domain.ScartoPhaseTrickEnd)
	g.SetCurrentTrick(scartoTrickCards(
		&domain.TrickCard{PlayerIdx: 0, Card: scartoSuitCard(domain.CardDesignSpade, 5)},
		&domain.TrickCard{PlayerIdx: 1, Card: scartoTrumpCard(4)},
		&domain.TrickCard{PlayerIdx: 2, Card: scartoExcuseCard()},
	))
	g.ResolveTrick()
	// Winner (player 1) keeps the two non-excuse cards; excuse owner (player 2) keeps the excuse.
	assert.Equal(t, 1, g.GetPlayer(1).GetTrickCount())
	assert.Equal(t, 1, g.GetPlayer(2).GetTrickCount())
	assert.Equal(t, 9, g.GetCardPoints(2)) // the excuse bout is 9 half-points
}

// --- Round-end scoring zero-sum ---

func TestScartoEnterRoundEndZeroSum(t *testing.T) {
	g := scartoNewReset()
	g.SetDealerIdx(0)
	g.SetTrickNumber(domain.ScartoTrickCount)
	g.SetLeadPlayerIdx(0)
	g.SetPhase(domain.ScartoPhaseTrickEnd)
	g.SetCurrentTrick(scartoTrickCards(
		&domain.TrickCard{PlayerIdx: 1, Card: scartoSuitCard(domain.CardDesignSpade, 14)},
		&domain.TrickCard{PlayerIdx: 2, Card: scartoSuitCard(domain.CardDesignSpade, 3)},
		&domain.TrickCard{PlayerIdx: 0, Card: scartoTrumpCard(5)}, // player 0 wins with a trump
	))
	g.ResolveTrick() // last trick -> RoundEnd + enterRoundEnd
	assert.Equal(t, domain.ScartoPhaseRoundEnd, g.GetPhase())
	scores := g.GetPlayerScores()
	assert.Equal(t, 0, scores[0]+scores[1]+scores[2])
	deal := g.GetDealScores()
	assert.Equal(t, 0, deal[0]+deal[1]+deal[2])
	assert.Contains(t, []domain.ScartoOutcome{
		domain.ScartoOutcomeWin, domain.ScartoOutcomeLoss, domain.ScartoOutcomeNone,
	}, g.GetOutcome())
}

// --- Full game drive (smoke) ---

func TestScartoFullGameDrive(t *testing.T) {
	g := scartoNewReset()
	guard := 0
	for !g.GetGameEndFlag() && guard < 20000 {
		guard++
		switch g.GetPhase() {
		case domain.ScartoPhaseScarto:
			if g.IsHumanScartoTurn() {
				_ = g.PlayerScarto(scartoFirstLegalDiscards(g))
			} else {
				g.CpuScarto()
			}
		case domain.ScartoPhasePlay:
			if g.IsHumanTurn() {
				idx := g.GetPlayableIndices(0)
				require.NotEmpty(t, idx)
				_ = g.PlayerPlay(idx[0])
				if g.GetPhase() == domain.ScartoPhaseTrickEnd {
					g.ResolveTrick()
				}
			} else {
				g.CpuPlay()
				if g.GetPhase() == domain.ScartoPhaseTrickEnd {
					g.ResolveTrick()
				}
			}
		case domain.ScartoPhaseTrickEnd:
			g.NextTrick()
		case domain.ScartoPhaseRoundEnd:
			g.ScoreRound()
			g.NextRound()
		case domain.ScartoPhaseGameEnd:
		}
	}
	assert.True(t, g.GetGameEndFlag())
	assert.GreaterOrEqual(t, g.GetWinnerPlayer(), -1)
	assert.Less(t, g.GetWinnerPlayer(), domain.ScartoPlayerCnt)
	// Cumulative scores remain zero-sum across the whole match.
	scores := g.GetPlayerScores()
	assert.Equal(t, 0, scores[0]+scores[1]+scores[2])
}

// scartoFirstLegalDiscards returns 3 indices of legal scarto cards from the
// dealer's hand (avoiding kings, courts, bouts, and trumps when possible).
func scartoFirstLegalDiscards(g *domain.Scarto) []int {
	p := g.GetPlayer(g.GetDealerIdx())
	legal := make([]int, 0)
	fallback := make([]int, 0)
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if domain.ScartoDiscardablePublic(c) {
			legal = append(legal, i)
		} else if domain.ScartoIsTrumpPublic(c) && !domain.ScartoIsBoutPublic(c) {
			fallback = append(fallback, i)
		}
	}
	legal = append(legal, fallback...)
	if len(legal) > domain.ScartoSurplus {
		legal = legal[:domain.ScartoSurplus]
	}
	return legal
}

// --- JSON round-trip & validation ---

func TestScartoJSONRoundTrip(t *testing.T) {
	g := scartoNewReset()
	data, err := json.Marshal(g)
	require.NoError(t, err)
	var restored domain.Scarto
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetDealerIdx(), restored.GetDealerIdx())
	assert.Equal(t, g.GetPlayer(0).GetCardsSize(), restored.GetPlayer(0).GetCardsSize())
}

func TestScartoUnmarshalErrors(t *testing.T) {
	cases := []string{
		`{"ps":[]}`,               // wrong player count
		`{"ps":[null,null,null]}`, // nil players
		`{"ph":99,"ps":` + scartoThreePlayers() + `}`, // bad phase
	}
	for _, c := range cases {
		var g domain.Scarto
		assert.Error(t, json.Unmarshal([]byte(c), &g), "input: %s", c)
	}
}

func TestScartoUnmarshalBadCard(t *testing.T) {
	in := `{"ps":` + scartoThreePlayers() + `,"ph":0,"ct":[{"pi":0,"c":{"design":5,"value":99}}]}`
	var g domain.Scarto
	assert.Error(t, json.Unmarshal([]byte(in), &g))
}

func TestScartoUnmarshalBadIndex(t *testing.T) {
	in := `{"ps":` + scartoThreePlayers() + `,"ph":0,"ci":9}`
	var g domain.Scarto
	assert.Error(t, json.Unmarshal([]byte(in), &g))
}

func scartoThreePlayers() string {
	one := `{"gp":{"isHuman":false,"cards":[]},"th":{}}`
	return "[" + one + "," + one + "," + one + "]"
}

// --- Config ---

func TestScartoConfigValidate(t *testing.T) {
	cfg := domain.DefaultScartoConfig()
	assert.NoError(t, cfg.Validate())
	assert.Equal(t, domain.ScartoDefaultDeals, cfg.TargetDeals)
	bad := cfg
	bad.TargetDeals = 0
	assert.Error(t, bad.Validate())
	bad2 := cfg
	bad2.CpuDifficulty = domain.ScartoCpuDifficulty(99)
	assert.Error(t, bad2.Validate())
}

// --- Accessors / misc ---

func TestScartoAccessors(t *testing.T) {
	g := scartoNewReset()
	g.SetRoundNumber(2)
	assert.Equal(t, 2, g.GetRoundNumber())
	g.SetCurrentPlayerIdx(1)
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	g.SetTrickNumber(4)
	assert.Equal(t, 4, g.GetTrickNumber())
	g.SetLeadPlayerIdx(2)
	assert.Equal(t, 2, g.GetLeadPlayerIdx())
	g.SetScarto([]*domain.Card{scartoSuitCard(1, 2)})
	assert.Equal(t, 1, g.GetScartoCount())
	assert.Nil(t, g.GetPlayer(99))
	assert.Nil(t, g.GetPlayableIndices(99))
	assert.Equal(t, 0, g.GetCardPoints(99))
	assert.NotNil(t, g.GetActionLog())
}

func TestScartoNextRoundGuardAndRotation(t *testing.T) {
	g := scartoNewReset()
	// NextRound is a no-op unless in RoundEnd.
	g.SetPhase(domain.ScartoPhasePlay)
	before := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, before, g.GetRoundNumber())
}

// **ピップが 3 枚に満たないときはブーでない切り札も捨てられる。** ドメインの
// 検証は前からそれを許していたのに、提示側 (CUI の一覧・Web の選択可能
// インデックス) は切り札を常に除外していたので、その手を引いた親は**画面から
// 枚数を揃えられなかった** (#6236)。
func TestScarto_DiscardableIndicesAllowTrumpWhenPipsRunShort(t *testing.T) {
	g := scartoNewReset()
	g.SetPhase(domain.ScartoPhaseScarto)
	g.SetDealerIdx(0)
	scartoSetHand(g, 0,
		scartoSuitCard(domain.CardDesignSpade, 5), // ピップ (1 枚だけ = 3 に満たない)
		scartoTrumpCard(10),                       // 切り札 (非ブー)
		scartoTrumpCard(21),                       // ブー
		scartoExcuseCard(),                        // エクスキューズ
		scartoSuitCard(domain.CardDesignSpade, domain.ScartoCourtMin), // コート
	)

	idxs := g.GetDiscardableIndices()
	assert.Contains(t, idxs, 0, "ピップは常に捨てられる")
	assert.Contains(t, idxs, 1, "ピップが足りないので非ブー切り札も捨てられる")
	assert.NotContains(t, idxs, 2, "ブーは捨てられない")
	assert.NotContains(t, idxs, 3, "エクスキューズは捨てられない")
	assert.NotContains(t, idxs, 4, "コートは捨てられない")

	// 提示とドメインの検証が一致していること。
	for _, i := range idxs {
		assert.NoError(t, g.PlayerScartoValidateForTest([]int{i}),
			"捨てられると提示した札 %d を検証が拒否している", i)
	}

	// **負のコントロール: ピップが 3 枚あれば切り札は捨てられない。**
	scartoSetHand(g, 0,
		scartoSuitCard(domain.CardDesignSpade, 5),
		scartoSuitCard(domain.CardDesignSpade, 6),
		scartoSuitCard(domain.CardDesignSpade, 7),
		scartoTrumpCard(10),
	)
	idxs = g.GetDiscardableIndices()
	assert.Equal(t, []int{0, 1, 2}, idxs, "ピップが足りるなら切り札は出さない")
	assert.Error(t, g.PlayerScartoValidateForTest([]int{3}), "検証も切り札を拒否する")
}

// スカルトフェーズでなければ何も捨てられない。
func TestScarto_DiscardableIndicesEmptyOutsideScarto(t *testing.T) {
	g := scartoNewReset()
	g.SetPhase(domain.ScartoPhasePlay)
	assert.Empty(t, g.GetDiscardableIndices())
}
