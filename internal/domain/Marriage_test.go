//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func marriageCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func marriageJoker(value int) *domain.Card {
	return domain.NewCard(domain.CardDesignJoker, value, false)
}

func newTestMarriage(n int) *domain.Marriage {
	players := make([]*domain.MarriagePlayer, n)
	players[0] = domain.NewMarriagePlayer(true)
	for i := 1; i < n; i++ {
		players[i] = domain.NewMarriagePlayer(false)
	}
	cfg := domain.DefaultMarriageConfig()
	cfg.PlayerCount = n
	return domain.NewMarriage(domain.NewTrumpCardsWithDecks(3, 6), players, cfg)
}

func setMarriageHand(p *domain.MarriagePlayer, cards []*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// validMarriageHand は 3 つのピュアシーケンスと 4 つのセットを含む有効宣言 21 枚。
func validMarriageHand() []*domain.Card {
	return []*domain.Card{
		// pure run ♠3-4-5
		marriageCard(domain.CardDesignSpade, 3),
		marriageCard(domain.CardDesignSpade, 4),
		marriageCard(domain.CardDesignSpade, 5),
		// pure run ♥6-7-8
		marriageCard(domain.CardDesignHeart, 6),
		marriageCard(domain.CardDesignHeart, 7),
		marriageCard(domain.CardDesignHeart, 8),
		// pure run ♣10-J-Q
		marriageCard(domain.CardDesignClover, 10),
		marriageCard(domain.CardDesignClover, 11),
		marriageCard(domain.CardDesignClover, 12),
		// set of 9s
		marriageCard(domain.CardDesignClover, 9),
		marriageCard(domain.CardDesignDiamond, 9),
		marriageCard(domain.CardDesignSpade, 9),
		// run ♦10-11-12-13
		// set of 10s
		marriageCard(domain.CardDesignSpade, 10),
		marriageCard(domain.CardDesignHeart, 10),
		marriageCard(domain.CardDesignDiamond, 10),
		// set of 13s
		marriageCard(domain.CardDesignSpade, 13),
		marriageCard(domain.CardDesignHeart, 13),
		marriageCard(domain.CardDesignDiamond, 13),
		// set of 2s
		marriageCard(domain.CardDesignSpade, 2),
		marriageCard(domain.CardDesignHeart, 2),
		marriageCard(domain.CardDesignDiamond, 2),
	}
}

func TestNewMarriage(t *testing.T) {
	g := newTestMarriage(4)
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 0, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetDeclarerIdx())
}

func TestMarriage_Reset(t *testing.T) {
	g := newTestMarriage(4)
	g.Reset()

	assert.Equal(t, domain.MarriagePhaseDraw, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, -1, g.GetWinnerIdx())
	assert.Equal(t, 4, g.GetPlayerCnt())
	// 各プレイヤーに 21 枚配られている。
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, domain.MarriageHandSize, g.GetPlayer(i).GetCardsSize())
	}
	assert.NotNil(t, g.GetWildJoker())
	assert.True(t, g.GetWildRank() >= 0 && g.GetWildRank() <= domain.CardValueMax)
	assert.NotNil(t, g.GetDiscardTop())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx()) // 席 0 のディーラーの左隣
}

func TestMarriage_DefaultConstructor(t *testing.T) {
	g := domain.NewDefaultMarriage()
	g.Reset()
	assert.Equal(t, domain.MarriageDefaultPlayerCount, g.GetPlayerCnt())
}

func TestMarriageConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultMarriageConfig().Validate())

	bad := domain.DefaultMarriageConfig()
	bad.PlayerCount = 1
	assert.Error(t, bad.Validate())

	bad2 := domain.DefaultMarriageConfig()
	bad2.PlayerCount = 6
	assert.Error(t, bad2.Validate())

	bad3 := domain.DefaultMarriageConfig()
	bad3.CpuDifficulty = 9
	assert.Error(t, bad3.Validate())

	bad4 := domain.DefaultMarriageConfig()
	bad4.TargetRounds = 0
	assert.Error(t, bad4.Validate())
}

func TestMarriage_CardPoints(t *testing.T) {
	assert.Equal(t, 10, domain.MarriageCardPoints(marriageCard(domain.CardDesignSpade, 1), 0))  // Ace
	assert.Equal(t, 10, domain.MarriageCardPoints(marriageCard(domain.CardDesignSpade, 13), 0)) // King
	assert.Equal(t, 10, domain.MarriageCardPoints(marriageCard(domain.CardDesignSpade, 10), 0))
	assert.Equal(t, 5, domain.MarriageCardPoints(marriageCard(domain.CardDesignSpade, 5), 0))
	assert.Equal(t, 0, domain.MarriageCardPoints(marriageJoker(1), 0))                        // printed joker
	assert.Equal(t, 0, domain.MarriageCardPoints(marriageCard(domain.CardDesignSpade, 7), 7)) // wild rank
	assert.Equal(t, 0, domain.MarriageCardPoints(marriageCard(domain.CardDesignSpade, 8), 7)) // poplu
	assert.Equal(t, 0, domain.MarriageCardPoints(marriageCard(domain.CardDesignSpade, 6), 7)) // jhiplu
}

func TestMarriage_WildRoles(t *testing.T) {
	tiplu := marriageCard(domain.CardDesignSpade, 13)
	assert.True(t, domain.MarriageIsTiplu(marriageCard(domain.CardDesignSpade, 13), tiplu))
	assert.True(t, domain.MarriageIsPoplu(marriageCard(domain.CardDesignSpade, 1), tiplu))
	assert.True(t, domain.MarriageIsJhiplu(marriageCard(domain.CardDesignSpade, 12), tiplu))
	assert.True(t, domain.MarriageIsAlter(marriageCard(domain.CardDesignHeart, 13), tiplu))
	assert.False(t, domain.MarriageIsPoplu(marriageCard(domain.CardDesignHeart, 1), tiplu))
	assert.False(t, domain.MarriageIsJhiplu(marriageCard(domain.CardDesignHeart, 12), tiplu))
}

func TestMarriage_MaalClassificationAndPoints(t *testing.T) {
	t.Run("K wraps to A and A wraps to K", func(t *testing.T) {
		tiplu := marriageCard(domain.CardDesignSpade, 13)
		assert.Equal(t, domain.MarriageMaalPoplu, domain.MarriageMaalOf(marriageCard(domain.CardDesignSpade, 1), tiplu))
		assert.Equal(t, domain.MarriageMaalJhiplu, domain.MarriageMaalOf(marriageCard(domain.CardDesignSpade, 12), tiplu))
	})
	t.Run("all kinds", func(t *testing.T) {
		tiplu := marriageCard(domain.CardDesignHeart, 7)
		cases := []struct {
			card   *domain.Card
			kind   domain.MarriageMaalKind
			points int
		}{
			{marriageCard(domain.CardDesignHeart, 7), domain.MarriageMaalTiplu, 3},
			{marriageCard(domain.CardDesignHeart, 8), domain.MarriageMaalPoplu, 2},
			{marriageCard(domain.CardDesignHeart, 6), domain.MarriageMaalJhiplu, 2},
			{marriageCard(domain.CardDesignSpade, 7), domain.MarriageMaalAlter, 1},
			{marriageJoker(1), domain.MarriageMaalJoker, 1},
			{marriageCard(domain.CardDesignSpade, 8), domain.MarriageMaalNone, 0},
		}
		for _, tc := range cases {
			assert.Equal(t, tc.kind, domain.MarriageMaalOf(tc.card, tiplu))
			assert.Equal(t, tc.points, domain.MarriageMaalPoints(tc.card, tiplu))
		}
	})
}

func TestMarriage_HasPureSequences(t *testing.T) {
	seq := func(s, start int) []*domain.Card {
		return []*domain.Card{marriageCard(s, start), marriageCard(s, start+1), marriageCard(s, start+2)}
	}
	cases := []struct {
		name  string
		cards []*domain.Card
		n     int
		want  bool
	}{
		{"zero", []*domain.Card{marriageCard(domain.CardDesignSpade, 2)}, 1, false},
		{"one", seq(domain.CardDesignSpade, 2), 1, true},
		{"two do not meet the maal gate", append(seq(domain.CardDesignSpade, 2), seq(domain.CardDesignHeart, 5)...), domain.MarriageMaalGateSequences, false},
		{"three meet the maal gate", append(append(seq(domain.CardDesignSpade, 2), seq(domain.CardDesignHeart, 5)...), seq(domain.CardDesignDiamond, 8)...), domain.MarriageMaalGateSequences, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, domain.MarriageHasPureSequences(tc.cards, 0, tc.n))
		})
	}
}

func TestMarriage_HasPureSequences_ManyCandidates(t *testing.T) {
	cards := make([]*domain.Card, 0, 21)
	for value := 1; value <= 13; value++ {
		cards = append(cards, marriageCard(domain.CardDesignSpade, value))
	}
	for value := 1; value <= 8; value++ {
		cards = append(cards, marriageCard(domain.CardDesignHeart, value))
	}

	// This guards the early-exit cap from dropping a valid answer in a
	// hand with the maximum number of pure-sequence candidates.
	assert.True(t, domain.MarriageHasPureSequences(cards, 0, 3))
}

func TestMarriage_HasPureSequences_ManyCandidatesWithoutThreeDisjointRuns(t *testing.T) {
	cards := make([]*domain.Card, 0, 21)
	for value := 1; value <= 5; value++ {
		cards = append(cards, marriageCard(domain.CardDesignSpade, value))
	}
	for value := 1; value <= 5; value++ {
		cards = append(cards, marriageCard(domain.CardDesignHeart, value))
	}
	for _, value := range []int{1, 3, 5, 7, 9, 11, 13} {
		cards = append(cards, marriageCard(domain.CardDesignDiamond, value))
	}
	for _, value := range []int{2, 4, 6, 8} {
		cards = append(cards, marriageCard(domain.CardDesignClover, value))
	}

	// This negative control guards the early-exit cap from accepting three
	// runs when only two disjoint pure runs exist.
	assert.False(t, domain.MarriageHasPureSequences(cards, 0, 3))
	assert.True(t, domain.MarriageHasPureSequences(cards, 0, 2))
}

func TestMarriage_PlayerMaalValueRequiresThreePureSequences(t *testing.T) {
	g := newTestMarriage(1)
	g.Reset()
	g.SetWildRank(0)
	g.SetWildJoker(marriageCard(domain.CardDesignSpade, 13))
	tiplu := g.GetWildJoker()
	base := []*domain.Card{
		marriageCard(domain.CardDesignSpade, 2), marriageCard(domain.CardDesignSpade, 3), marriageCard(domain.CardDesignSpade, 4),
		marriageCard(domain.CardDesignHeart, 5), marriageCard(domain.CardDesignHeart, 6), marriageCard(domain.CardDesignHeart, 7),
	}
	maal := []*domain.Card{tiplu, marriageJoker(1)}
	setMarriageHand(g.GetPlayer(0), append(append([]*domain.Card{}, base...), maal...))
	assert.Equal(t, 0, g.PlayerMaalValue(0), "two pure sequences do not unlock maal")
	setMarriageHand(g.GetPlayer(0), append(append(append([]*domain.Card{}, base...), seqCards(domain.CardDesignDiamond, 9)...), maal...))
	assert.Equal(t, domain.MarriageMaalTotal(marriageCollectForTest(g.GetPlayer(0)), tiplu), g.PlayerMaalValue(0))
}

func TestMarriage_ScoreRoundDeductsMaal(t *testing.T) {
	g := newTestMarriage(1)
	g.Reset()
	g.SetWildRank(0)
	g.SetWildJoker(marriageCard(domain.CardDesignSpade, 13))
	tiplu := g.GetWildJoker()
	hand := append(append(seqCards(domain.CardDesignSpade, 2), seqCards(domain.CardDesignHeart, 5)...), seqCards(domain.CardDesignDiamond, 8)...)
	hand = append(hand, tiplu, marriageJoker(1))
	setMarriageHand(g.GetPlayer(0), hand)
	base := domain.MarriageDeadwoodScore(hand, 0)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MarriagePhaseDraw)
	g.SetDrawPile(nil)
	g.SetDiscardPile(nil)
	assert.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, base-g.PlayerMaalValue(0), g.GetPlayer(0).GetRoundScore())
}

func seqCards(s, start int) []*domain.Card {
	return []*domain.Card{marriageCard(s, start), marriageCard(s, start+1), marriageCard(s, start+2)}
}

func marriageCollectForTest(p *domain.MarriagePlayer) []*domain.Card {
	cards := make([]*domain.Card, p.GetCardsSize())
	for i := range cards {
		cards[i] = p.GetCard(i)
	}
	return cards
}

func TestMarriage_HasPureSequence(t *testing.T) {
	pure := []*domain.Card{
		marriageCard(domain.CardDesignSpade, 3),
		marriageCard(domain.CardDesignSpade, 4),
		marriageCard(domain.CardDesignSpade, 5),
	}
	assert.True(t, domain.MarriageHasPureSequence(pure, 0))

	noPure := []*domain.Card{
		marriageCard(domain.CardDesignSpade, 3),
		marriageCard(domain.CardDesignHeart, 7),
		marriageCard(domain.CardDesignDiamond, 11),
	}
	assert.False(t, domain.MarriageHasPureSequence(noPure, 0))

	// wild-rank card cannot make a pure sequence.
	withWild := []*domain.Card{
		marriageCard(domain.CardDesignSpade, 3),
		marriageCard(domain.CardDesignSpade, 4),
		marriageCard(domain.CardDesignSpade, 5), // 5 is wild
	}
	assert.False(t, domain.MarriageHasPureSequence(withWild, 5))
}

func TestMarriage_ValidateDeclaration_Valid(t *testing.T) {
	assert.True(t, domain.MarriageValidateDeclaration(validMarriageHand(), 0))
}

func TestMarriage_ValidateDeclaration_WrongLength(t *testing.T) {
	short := validMarriageHand()[:12]
	assert.False(t, domain.MarriageValidateDeclaration(short, 0))
}

func TestMarriage_ValidateDeclaration_OnlyOneSequence(t *testing.T) {
	// 1 run + 3 sets: only a single sequence → invalid (needs 2 sequences).
	hand := []*domain.Card{
		marriageCard(domain.CardDesignSpade, 3),
		marriageCard(domain.CardDesignSpade, 4),
		marriageCard(domain.CardDesignSpade, 5),
		marriageCard(domain.CardDesignSpade, 7),
		marriageCard(domain.CardDesignHeart, 7),
		marriageCard(domain.CardDesignDiamond, 7),
		marriageCard(domain.CardDesignSpade, 9),
		marriageCard(domain.CardDesignHeart, 9),
		marriageCard(domain.CardDesignDiamond, 9),
		marriageCard(domain.CardDesignSpade, 11),
		marriageCard(domain.CardDesignHeart, 11),
		marriageCard(domain.CardDesignDiamond, 11),
		marriageCard(domain.CardDesignClover, 11),
	}
	assert.False(t, domain.MarriageValidateDeclaration(hand, 0))
}

func TestMarriage_ValidateDeclaration_NoPureSequence(t *testing.T) {
	// 2 impure sequences (each uses a joker), no pure → invalid.
	hand := []*domain.Card{
		marriageCard(domain.CardDesignSpade, 3),
		marriageCard(domain.CardDesignSpade, 4),
		marriageJoker(1), // fills ♠5
		marriageCard(domain.CardDesignHeart, 7),
		marriageCard(domain.CardDesignHeart, 8),
		marriageJoker(2), // fills ♥9
		marriageCard(domain.CardDesignSpade, 10),
		marriageCard(domain.CardDesignHeart, 10),
		marriageCard(domain.CardDesignDiamond, 10),
		marriageCard(domain.CardDesignSpade, 13),
		marriageCard(domain.CardDesignHeart, 13),
		marriageCard(domain.CardDesignDiamond, 13),
		marriageCard(domain.CardDesignClover, 13),
	}
	assert.False(t, domain.MarriageValidateDeclaration(hand, 0))
}

func TestMarriage_ValidateDeclaration_ImpurePlusPure(t *testing.T) {
	// 1 pure run + 1 impure run + sets is still invalid: Marriage requires 3 pure sequences.
	hand := []*domain.Card{
		marriageCard(domain.CardDesignSpade, 3),
		marriageCard(domain.CardDesignSpade, 4),
		marriageCard(domain.CardDesignSpade, 5), // pure run
		marriageCard(domain.CardDesignHeart, 7),
		marriageCard(domain.CardDesignHeart, 8),
		marriageJoker(1), // impure run fills ♥9
		marriageCard(domain.CardDesignSpade, 10),
		marriageCard(domain.CardDesignHeart, 10),
		marriageCard(domain.CardDesignDiamond, 10),
		marriageCard(domain.CardDesignSpade, 13),
		marriageCard(domain.CardDesignHeart, 13),
		marriageCard(domain.CardDesignDiamond, 13),
		marriageCard(domain.CardDesignClover, 13),
	}
	assert.False(t, domain.MarriageValidateDeclaration(hand, 0))
}

func TestMarriage_DeadwoodScore(t *testing.T) {
	// No pure sequence → full cap 80.
	noPure := []*domain.Card{
		marriageCard(domain.CardDesignSpade, 3),
		marriageCard(domain.CardDesignHeart, 7),
		marriageCard(domain.CardDesignDiamond, 11),
	}
	assert.Equal(t, domain.MarriageDeadwoodCap, domain.MarriageDeadwoodScore(noPure, 0))

	// Pure run + small deadwood.
	smallDW := []*domain.Card{
		marriageCard(domain.CardDesignSpade, 3),
		marriageCard(domain.CardDesignSpade, 4),
		marriageCard(domain.CardDesignSpade, 5),
		marriageCard(domain.CardDesignHeart, 2),
	}
	assert.Equal(t, 2, domain.MarriageDeadwoodScore(smallDW, 0))

	// Pure run + 100 points of un-meldable high cards → capped at 80.
	capped := []*domain.Card{
		marriageCard(domain.CardDesignSpade, 1),
		marriageCard(domain.CardDesignSpade, 2),
		marriageCard(domain.CardDesignSpade, 3), // pure run
		marriageCard(domain.CardDesignSpade, 10),
		marriageCard(domain.CardDesignSpade, 12),
		marriageCard(domain.CardDesignHeart, 11),
		marriageCard(domain.CardDesignHeart, 13),
		marriageCard(domain.CardDesignClover, 10),
		marriageCard(domain.CardDesignClover, 12),
		marriageCard(domain.CardDesignDiamond, 11),
		marriageCard(domain.CardDesignDiamond, 13),
		marriageCard(domain.CardDesignHeart, 1),
		marriageCard(domain.CardDesignDiamond, 1),
	}
	assert.Equal(t, domain.MarriageDeadwoodCap, domain.MarriageDeadwoodScore(capped, 0))
}

func TestMarriage_DrawAndDiscard(t *testing.T) {
	g := newTestMarriage(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MarriagePhaseDraw)

	before := g.GetPlayer(0).GetCardsSize()
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, before+1, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.MarriagePhaseDiscard, g.GetPhase())

	require.NoError(t, g.PlayerDiscard(0))
	assert.Equal(t, before, g.GetPlayer(0).GetCardsSize())
	assert.Equal(t, domain.MarriagePhaseDraw, g.GetPhase())
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
}

func TestMarriage_DrawFromDiscard(t *testing.T) {
	g := newTestMarriage(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MarriagePhaseDraw)
	g.SetDiscardPile([]*domain.Card{marriageCard(domain.CardDesignSpade, 7)})

	require.NoError(t, g.PlayerDrawFromDiscard())
	assert.Equal(t, domain.MarriagePhaseDiscard, g.GetPhase())
}

func TestMarriage_DrawGuards(t *testing.T) {
	g := newTestMarriage(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)

	g.SetPhase(domain.MarriagePhaseDiscard)
	assert.Error(t, g.PlayerDrawFromStock()) // wrong phase

	g.SetPhase(domain.MarriagePhaseDraw)
	g.SetCurrentPlayerIdx(1) // CPU turn
	assert.Error(t, g.PlayerDrawFromStock())

	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MarriagePhaseDraw)
	g.SetDiscardPile(nil)
	assert.Error(t, g.PlayerDrawFromDiscard()) // empty discard
}

func TestMarriage_DiscardOutOfRange(t *testing.T) {
	g := newTestMarriage(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MarriagePhaseDiscard)
	assert.Error(t, g.PlayerDiscard(-1))
	assert.Error(t, g.PlayerDiscard(999))
}

func TestMarriage_Declare_Valid(t *testing.T) {
	g := newTestMarriage(2)
	g.Reset()
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MarriagePhaseDiscard)
	g.SetWildRank(0)

	hand := append(validMarriageHand(), marriageCard(domain.CardDesignClover, 2)) // 22nd = finish
	setMarriageHand(g.GetPlayer(0), hand)
	// give opponent a no-pure hand.
	setMarriageHand(g.GetPlayer(1), []*domain.Card{
		marriageCard(domain.CardDesignSpade, 13),
		marriageCard(domain.CardDesignHeart, 11),
	})
	require.NoError(t, g.PlayerDeclare(21))
	assert.True(t, g.GetDeclarationValid())
	assert.Equal(t, 0, g.GetDeclarerIdx())
	assert.Equal(t, domain.MarriagePhaseRoundEnd, g.GetPhase())
	assert.Equal(t, -g.PlayerMaalValue(0), g.GetPlayer(0).GetRoundScore()) // winner scores minus maal
	assert.Less(t, g.GetPlayer(0).GetRoundScore(), 0)
	assert.Equal(t, domain.MarriageDeadwoodCap, g.GetPlayer(1).GetRoundScore()) // no pure → 80
}

func TestMarriage_Declare_Invalid(t *testing.T) {
	g := newTestMarriage(2)
	g.Reset()
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MarriagePhaseDiscard)
	g.SetWildRank(0)

	// Intentionally short invalid hand with non-melding cards and a finish card.
	junk := []*domain.Card{
		marriageCard(domain.CardDesignSpade, 2),
		marriageCard(domain.CardDesignHeart, 5),
		marriageCard(domain.CardDesignDiamond, 8),
		marriageCard(domain.CardDesignClover, 11),
		marriageCard(domain.CardDesignSpade, 13),
		marriageCard(domain.CardDesignHeart, 3),
		marriageCard(domain.CardDesignDiamond, 6),
		marriageCard(domain.CardDesignClover, 9),
		marriageCard(domain.CardDesignSpade, 12),
		marriageCard(domain.CardDesignHeart, 4),
		marriageCard(domain.CardDesignDiamond, 7),
		marriageCard(domain.CardDesignClover, 10),
		marriageCard(domain.CardDesignSpade, 1),
		marriageCard(domain.CardDesignHeart, 13), // finish
	}
	setMarriageHand(g.GetPlayer(0), junk)
	require.NoError(t, g.PlayerDeclare(13))
	assert.False(t, g.GetDeclarationValid())
	assert.Equal(t, domain.MarriageDeadwoodCap-g.PlayerMaalValue(0), g.GetPlayer(0).GetRoundScore()) // invalid → cap minus maal
}

func TestMarriage_DeclareGuards(t *testing.T) {
	g := newTestMarriage(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MarriagePhaseDraw)
	assert.Error(t, g.PlayerDeclare(0)) // wrong phase

	g.SetPhase(domain.MarriagePhaseDiscard)
	assert.Error(t, g.PlayerDeclare(999)) // out of range
}

func TestMarriage_Recycle(t *testing.T) {
	g := newTestMarriage(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MarriagePhaseDraw)
	g.SetDrawPile(nil)
	g.SetDiscardPile([]*domain.Card{
		marriageCard(domain.CardDesignSpade, 2),
		marriageCard(domain.CardDesignHeart, 3),
		marriageCard(domain.CardDesignDiamond, 4),
	})
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, domain.MarriagePhaseDiscard, g.GetPhase())
}

func TestMarriage_StockOut(t *testing.T) {
	g := newTestMarriage(2)
	g.Reset()
	g.SetRoundNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MarriagePhaseDraw)
	g.SetDrawPile(nil)
	g.SetDiscardPile([]*domain.Card{marriageCard(domain.CardDesignSpade, 2)})
	require.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, domain.MarriagePhaseRoundEnd, g.GetPhase())
	assert.Equal(t, -1, g.GetDeclarerIdx())
}

func TestMarriage_NextRound(t *testing.T) {
	g := newTestMarriage(2)
	g.Reset()
	g.SetRoundNumber(1)
	g.SetPhase(domain.MarriagePhaseRoundEnd)
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
	assert.Equal(t, domain.MarriagePhaseDraw, g.GetPhase())

	// wrong phase → no-op
	g.SetPhase(domain.MarriagePhaseDraw)
	g.NextRound()
	assert.Equal(t, 2, g.GetRoundNumber())
}

func TestMarriage_NextRound_GameEnd(t *testing.T) {
	g := newTestMarriage(2)
	cfg := domain.DefaultMarriageConfig()
	cfg.PlayerCount = 2
	cfg.TargetRounds = 2
	g.SetConfig(cfg)
	g.Reset()
	g.SetRoundNumber(2)
	g.SetPhase(domain.MarriagePhaseRoundEnd)
	g.NextRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.MarriagePhaseGameEnd, g.GetPhase())
	assert.True(t, g.GetWinnerIdx() >= 0)
}

func TestMarriage_CpuBoundedLoop(t *testing.T) {
	g := newTestMarriage(2)
	g.Reset()
	for iter := 0; iter < 400 && !g.GetGameEndFlag(); iter++ {
		phase := g.GetPhase()
		if phase == domain.MarriagePhaseRoundEnd {
			g.NextRound()
			continue
		}
		if g.IsHumanTurn() {
			if g.GetPhase() == domain.MarriagePhaseDraw {
				_ = g.PlayerDrawFromStock()
			} else {
				_ = g.PlayerDiscard(0)
			}
			continue
		}
		g.CpuPlay()
	}
	// 一貫した状態であることのみ確認（ゲーム終了は保証しない）。
	p := g.GetPhase()
	assert.True(t, p >= domain.MarriagePhaseDraw && p <= domain.MarriagePhaseGameEnd)
}

func TestMarriage_CpuDifficulties(t *testing.T) {
	for _, d := range []domain.MarriageCpuDifficulty{
		domain.MarriageCpuDifficultyEasy,
		domain.MarriageCpuDifficultyNormal,
		domain.MarriageCpuDifficultyHard,
	} {
		g := newTestMarriage(2)
		cfg := domain.DefaultMarriageConfig()
		cfg.PlayerCount = 2
		cfg.CpuDifficulty = d
		g.SetConfig(cfg)
		g.Reset()
		g.SetCurrentPlayerIdx(1) // CPU
		g.SetPhase(domain.MarriagePhaseDraw)
		g.CpuPlay()
		g.CpuPlay()
		assert.NotNil(t, g.GetPlayer(1))
	}
}

func TestMarriage_PlayerHelpers(t *testing.T) {
	g := newTestMarriage(2)
	g.Reset()
	g.SetWildRank(0)
	setMarriageHand(g.GetPlayer(0), []*domain.Card{
		marriageCard(domain.CardDesignSpade, 3),
		marriageCard(domain.CardDesignSpade, 4),
		marriageCard(domain.CardDesignSpade, 5),
		marriageCard(domain.CardDesignHeart, 9),
	})
	assert.True(t, g.PlayerHasPureSequence(0))
	assert.Equal(t, 9, g.PlayerDeadwoodValue(0))
	assert.Nil(t, g.GetPlayer(99))
	assert.False(t, g.PlayerHasPureSequence(99))
	assert.Equal(t, 0, g.PlayerDeadwoodValue(99))
}

func TestMarriage_JSONRoundTrip(t *testing.T) {
	g := newTestMarriage(3)
	g.Reset()
	data, err := g.MarshalJSON()
	require.NoError(t, err)

	var restored domain.Marriage
	require.NoError(t, restored.UnmarshalJSON(data))
	assert.Equal(t, g.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetWildRank(), restored.GetWildRank())
}

func TestMarriage_UnmarshalJSON_BadJSON(t *testing.T) {
	var g domain.Marriage
	assert.Error(t, g.UnmarshalJSON([]byte("not json")))
}

func TestMarriage_UnmarshalJSON_Validation(t *testing.T) {
	base := newTestMarriage(2)
	base.Reset()
	data, err := base.MarshalJSON()
	require.NoError(t, err)

	tamper := func(mut func(m map[string]json.RawMessage)) []byte {
		var m map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(data, &m))
		mut(m)
		out, err := json.Marshal(m)
		require.NoError(t, err)
		return out
	}

	cases := map[string]func(m map[string]json.RawMessage){
		"currentPlayerIdx": func(m map[string]json.RawMessage) { m["ci"] = json.RawMessage("99") },
		"dealerIdx":        func(m map[string]json.RawMessage) { m["di"] = json.RawMessage("99") },
		"phase":            func(m map[string]json.RawMessage) { m["ps"] = json.RawMessage("99") },
		"roundNumber":      func(m map[string]json.RawMessage) { m["rn"] = json.RawMessage("-1") },
		"winnerIdx":        func(m map[string]json.RawMessage) { m["wi"] = json.RawMessage("99") },
		"declarerIdx":      func(m map[string]json.RawMessage) { m["de"] = json.RawMessage("99") },
		"wildRank":         func(m map[string]json.RawMessage) { m["wr"] = json.RawMessage("99") },
		"playerCount":      func(m map[string]json.RawMessage) { m["pl"] = json.RawMessage("[]") },
		"nilPlayer":        func(m map[string]json.RawMessage) { m["pl"] = json.RawMessage("[null,null]") },
	}
	for name, mut := range cases {
		t.Run(name, func(t *testing.T) {
			var g domain.Marriage
			assert.Error(t, g.UnmarshalJSON(tamper(mut)))
		})
	}
}

func TestMarriage_UnmarshalJSON_FiltersNilCards(t *testing.T) {
	base := newTestMarriage(2)
	base.Reset()
	data, err := base.MarshalJSON()
	require.NoError(t, err)

	var m map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &m))
	m["dp"] = json.RawMessage(`[null,{"d":1,"v":5,"w":false}]`)
	out, err := json.Marshal(m)
	require.NoError(t, err)

	var g domain.Marriage
	require.NoError(t, g.UnmarshalJSON(out))
	// nil は除去され、有効カード 1 枚だけがトップに残る。
	assert.NotNil(t, g.GetDiscardTop())
}

func TestMarriage_ActionLogAccumulates(t *testing.T) {
	g := newTestMarriage(2)
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.MarriagePhaseDraw)
	require.NoError(t, g.PlayerDrawFromStock())
	assert.NotEmpty(t, g.GetActionLog())
}
