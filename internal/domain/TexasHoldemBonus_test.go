package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func makeHandTHB(specs ...cd) []*domain.Card {
	cards := make([]*domain.Card, len(specs))
	for i, s := range specs {
		cards[i] = domain.NewCard(s.d, s.v, false)
	}
	return cards
}

// playToShowdown drives the game from the bet phase through to the river so
// the showdown payout logic can be exercised with deterministically-set hands.
func playToShowdown(t *testing.T, g *domain.TexasHoldemBonus, ante, bonus int) {
	t.Helper()
	require.NoError(t, g.Bet(ante, bonus))
	require.NoError(t, g.Play())
	require.NoError(t, g.Check())
	require.NoError(t, g.Check())
}

func TestNewDefaultTexasHoldemBonus(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	assert.Equal(t, domain.TexasHoldemBonusPhaseBet, g.GetPhase())
	assert.Equal(t, domain.TexasHoldemBonusDefaultChips, g.GetChips())
	assert.False(t, g.GetGameEndFlag())
	assert.Nil(t, g.GetPlayerHand())
	assert.Nil(t, g.GetDealerHand())
	assert.Nil(t, g.GetCommunity())
}

func TestTexasHoldemBonus_Reset(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Fold())
	assert.Equal(t, domain.TexasHoldemBonusPhaseEnd, g.GetPhase())

	g.Reset()
	assert.Equal(t, domain.TexasHoldemBonusPhaseBet, g.GetPhase())
	assert.False(t, g.GetGameEndFlag())
	assert.Nil(t, g.GetPlayerHand())
	assert.Nil(t, g.GetDealerHand())
	assert.Nil(t, g.GetCommunity())
	assert.Equal(t, 0, g.GetAnteBet())
	assert.Equal(t, 0, g.GetBonusBet())
	assert.Equal(t, 0, g.GetFlopBet())
	assert.Equal(t, 0, g.GetTurnBet())
	assert.Equal(t, 0, g.GetRiverBet())
}

func TestTexasHoldemBonus_PlayerHandRank_PopulatedAtFlopAndTurn(t *testing.T) {
	// Drive a hand to FLOP with a deterministic pair (player Sd2 / Hd2 + flop with a 2),
	// then verify GetPlayerHandRank() reflects the pair *before* showdown so the
	// frontend hint at FLOP/TURN can read a meaningful rank.
	g := domain.NewDefaultTexasHoldemBonus()
	require.NoError(t, g.Bet(100, 0))
	g.SetPlayerHand(makeHandTHB(
		cd{domain.CardDesignSpade, 2},
		cd{domain.CardDesignHeart, 2},
	))
	g.SetCommunity(makeHandTHB(
		cd{domain.CardDesignClover, 7},
		cd{domain.CardDesignDiamond, 9},
		cd{domain.CardDesignSpade, 13},
	))
	g.SetPhase(domain.TexasHoldemBonusPhasePreFlop)
	// Drive Play() so the production code path runs (dealFlop overwrites the
	// community, so we re-set the override then trigger updatePlayerCurrentRank
	// via a Check transitioning to TURN with our deterministic community).
	g.SetCommunity(makeHandTHB(
		cd{domain.CardDesignClover, 7},
		cd{domain.CardDesignDiamond, 9},
		cd{domain.CardDesignSpade, 13},
	))
	g.SetPhase(domain.TexasHoldemBonusPhaseFlop)
	require.NoError(t, g.Check()) // deals turn card and updates current rank
	require.GreaterOrEqual(t, g.GetPlayerHandRank(), domain.PokerHandOnePair, "expected pair or better mid-hand")
}

func TestTexasHoldemBonus_Reset_RefillChips(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	g.SetChips(5)
	g.Reset()
	assert.Equal(t, domain.TexasHoldemBonusDefaultChips, g.GetChips())
}

func TestTexasHoldemBonus_Bet_WrongPhase(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	g.SetPhase(domain.TexasHoldemBonusPhasePreFlop)
	err := g.Bet(100, 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestTexasHoldemBonus_Bet_InvalidAnteAmount(t *testing.T) {
	tests := []struct {
		name string
		ante int
	}{
		{"TooLow", 5},
		{"NotMultiple", 15},
		{"TooHigh", 20000},
		{"Zero", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := domain.NewDefaultTexasHoldemBonus()
			err := g.Bet(tt.ante, 0)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestTexasHoldemBonus_Bet_InvalidBonusAmount(t *testing.T) {
	tests := []struct {
		name  string
		bonus int
	}{
		{"Negative", -10},
		{"TooLow", 5},
		{"NotMultiple", 15},
		{"TooHigh", 20000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := domain.NewDefaultTexasHoldemBonus()
			err := g.Bet(100, tt.bonus)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestTexasHoldemBonus_Bet_InsufficientChips(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	g.SetChips(50)
	err := g.Bet(100, 0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestTexasHoldemBonus_Bet_Success(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	err := g.Bet(100, 50)
	assert.NoError(t, err)
	assert.Equal(t, domain.TexasHoldemBonusPhasePreFlop, g.GetPhase())
	assert.Equal(t, 100, g.GetAnteBet())
	assert.Equal(t, 50, g.GetBonusBet())
	assert.Len(t, g.GetPlayerHand(), 2)
	assert.Len(t, g.GetDealerHand(), 2)
	assert.Equal(t, domain.TexasHoldemBonusDefaultChips-150, g.GetChips())
}

func TestTexasHoldemBonus_Play_WrongPhase(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	err := g.Play()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestTexasHoldemBonus_Play_InsufficientChips(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	require.NoError(t, g.Bet(100, 0))
	g.SetChips(0)
	err := g.Play()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestTexasHoldemBonus_Play_DealsFlop(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Play())
	assert.Equal(t, 200, g.GetFlopBet())
	assert.Equal(t, domain.TexasHoldemBonusPhaseFlop, g.GetPhase())
	assert.Len(t, g.GetCommunity(), 3)
}

func TestTexasHoldemBonus_Fold_WrongPhase(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	err := g.Fold()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestTexasHoldemBonus_Fold_Success(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	require.NoError(t, g.Bet(100, 0))
	chipsBefore := g.GetChips()
	require.NoError(t, g.Fold())
	assert.Equal(t, domain.TexasHoldemBonusPhaseEnd, g.GetPhase())
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, domain.GameResultLose, g.GetResult())
	assert.Equal(t, chipsBefore, g.GetChips())
}

func TestTexasHoldemBonus_Fold_BonusStillEvaluated(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	require.NoError(t, g.Bet(100, 10))

	// Pocket aces => 30:1 bonus payout
	g.SetPlayerHand(makeHandTHB(
		cd{domain.CardDesignSpade, 1},
		cd{domain.CardDesignClover, 1},
	))
	chipsBefore := g.GetChips()
	require.NoError(t, g.Fold())
	expected := 10 + 10*domain.TexasHoldemBonusBonusPayAA
	assert.Equal(t, expected, g.GetBonusPayout())
	assert.Equal(t, chipsBefore+expected, g.GetChips())
}

func TestTexasHoldemBonus_Check_WrongPhase(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	err := g.Check()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestTexasHoldemBonus_Raise_WrongPhase(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	err := g.Raise()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestTexasHoldemBonus_Check_FlopAdvancesToTurn(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Play())
	require.NoError(t, g.Check())
	assert.Equal(t, domain.TexasHoldemBonusPhaseTurn, g.GetPhase())
	assert.Len(t, g.GetCommunity(), 4)
	assert.Equal(t, 0, g.GetTurnBet())
}

func TestTexasHoldemBonus_Check_TurnTriggersShowdown(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Play())
	require.NoError(t, g.Check())
	require.NoError(t, g.Check())
	assert.Equal(t, domain.TexasHoldemBonusPhaseEnd, g.GetPhase())
	assert.Len(t, g.GetCommunity(), 5)
	assert.True(t, g.GetGameEndFlag())
}

func TestTexasHoldemBonus_Raise_FlopPlacesTurnBet(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Play())
	chipsBefore := g.GetChips()
	require.NoError(t, g.Raise())
	assert.Equal(t, 100, g.GetTurnBet())
	assert.Equal(t, chipsBefore-100, g.GetChips())
	assert.Equal(t, domain.TexasHoldemBonusPhaseTurn, g.GetPhase())
	assert.Len(t, g.GetCommunity(), 4)
}

func TestTexasHoldemBonus_Raise_TurnPlacesRiverAndShowdown(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Play())
	require.NoError(t, g.Raise())
	require.NoError(t, g.Raise())
	assert.Equal(t, 100, g.GetRiverBet())
	assert.Equal(t, domain.TexasHoldemBonusPhaseEnd, g.GetPhase())
	assert.True(t, g.GetGameEndFlag())
}

func TestTexasHoldemBonus_Raise_InsufficientChipsFlop(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Play())
	g.SetChips(0)
	err := g.Raise()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestTexasHoldemBonus_Raise_InsufficientChipsTurn(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Play())
	require.NoError(t, g.Check())
	g.SetChips(0)
	err := g.Raise()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestTexasHoldemBonus_Showdown_PlayerWins(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	g.SetChips(100000)
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Play())
	// Player: AA pair, Dealer: KK pair, board does not improve either to trips/full.
	g.SetPlayerHand(makeHandTHB(
		cd{domain.CardDesignSpade, 1},
		cd{domain.CardDesignClover, 1},
	))
	g.SetDealerHand(makeHandTHB(
		cd{domain.CardDesignHeart, 13},
		cd{domain.CardDesignDiamond, 13},
	))
	g.SetCommunity(makeHandTHB(
		cd{domain.CardDesignSpade, 2},
		cd{domain.CardDesignClover, 4},
		cd{domain.CardDesignHeart, 7},
		cd{domain.CardDesignSpade, 9},
		cd{domain.CardDesignDiamond, 12},
	))
	g.ForceResolve()

	assert.Equal(t, domain.GameResultWin, g.GetResult())
	// Pair (One Pair) carries no ante bonus -> ante 100*2 = 200, flop bet 200*2 = 400.
	assert.Equal(t, 200, g.GetAntePayout())
	assert.Equal(t, 400, g.GetPlayPayout())
}

func TestTexasHoldemBonus_Showdown_DealerWins(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	g.SetChips(100000)
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Play())
	// Player: 2,3 unsuited; Dealer: AA -> Dealer wins.
	g.SetPlayerHand(makeHandTHB(
		cd{domain.CardDesignSpade, 2},
		cd{domain.CardDesignClover, 3},
	))
	g.SetDealerHand(makeHandTHB(
		cd{domain.CardDesignHeart, 1},
		cd{domain.CardDesignDiamond, 1},
	))
	g.SetCommunity(makeHandTHB(
		cd{domain.CardDesignSpade, 5},
		cd{domain.CardDesignClover, 7},
		cd{domain.CardDesignHeart, 9},
		cd{domain.CardDesignSpade, 11},
		cd{domain.CardDesignClover, 13},
	))
	g.ForceResolve()

	assert.Equal(t, domain.GameResultLose, g.GetResult())
	assert.Equal(t, 0, g.GetAntePayout())
	assert.Equal(t, 0, g.GetPlayPayout())
}

func TestTexasHoldemBonus_Showdown_Push(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	g.SetChips(100000)
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Play())
	// Place turn + river bets directly via SetPhase to avoid the implicit
	// resolve() inside the second Raise(); we want to override the hands and
	// drive resolve() ourselves with deterministic cards.
	g.SetTurnBet(100)
	g.SetRiverBet(100)
	g.SetPhase(domain.TexasHoldemBonusPhaseTurn)
	// Override game state to a deterministic push: both players play the
	// board straight 5-6-7-8-9 (player and dealer hole cards are unused).
	g.SetPlayerHand(makeHandTHB(
		cd{domain.CardDesignSpade, 2},
		cd{domain.CardDesignSpade, 3},
	))
	g.SetDealerHand(makeHandTHB(
		cd{domain.CardDesignClover, 2},
		cd{domain.CardDesignClover, 3},
	))
	g.SetCommunity(makeHandTHB(
		cd{domain.CardDesignHeart, 5},
		cd{domain.CardDesignDiamond, 6},
		cd{domain.CardDesignSpade, 7},
		cd{domain.CardDesignDiamond, 8},
		cd{domain.CardDesignClover, 9},
	))
	g.ForceResolve()

	assert.Equal(t, domain.GameResultDraw, g.GetResult())
	// On a push: ante and play bets are returned (1× their wager).
	// Player straight => +1× ante bonus on top of the returned ante.
	expectedAnte := 100 + 100*domain.TexasHoldemBonusAntePayStraight
	assert.Equal(t, expectedAnte, g.GetAntePayout())
	expectedPlay := 200 + 100 + 100 // flop + turn + river bets
	assert.Equal(t, expectedPlay, g.GetPlayPayout())
}

func TestTexasHoldemBonus_AnteBonusPayouts(t *testing.T) {
	weakDealer := makeHandTHB(
		cd{domain.CardDesignDiamond, 3},
		cd{domain.CardDesignHeart, 4},
	)
	tests := []struct {
		name       string
		player     []*domain.Card
		community  []*domain.Card
		multiplier int
	}{
		{
			name: "RoyalFlush",
			player: makeHandTHB(
				cd{domain.CardDesignSpade, 1},
				cd{domain.CardDesignSpade, 13},
			),
			community: makeHandTHB(
				cd{domain.CardDesignSpade, 12},
				cd{domain.CardDesignSpade, 11},
				cd{domain.CardDesignSpade, 10},
				cd{domain.CardDesignClover, 2},
				cd{domain.CardDesignHeart, 4},
			),
			multiplier: domain.TexasHoldemBonusAntePayRoyalFlush,
		},
		{
			name: "StraightFlush",
			player: makeHandTHB(
				cd{domain.CardDesignClover, 5},
				cd{domain.CardDesignClover, 6},
			),
			community: makeHandTHB(
				cd{domain.CardDesignClover, 7},
				cd{domain.CardDesignClover, 8},
				cd{domain.CardDesignClover, 9},
				cd{domain.CardDesignHeart, 2},
				cd{domain.CardDesignDiamond, 4},
			),
			multiplier: domain.TexasHoldemBonusAntePayStraightFlush,
		},
		{
			name: "FourOfAKind",
			player: makeHandTHB(
				cd{domain.CardDesignSpade, 7},
				cd{domain.CardDesignClover, 7},
			),
			community: makeHandTHB(
				cd{domain.CardDesignHeart, 7},
				cd{domain.CardDesignDiamond, 7},
				cd{domain.CardDesignSpade, 4},
				cd{domain.CardDesignClover, 9},
				cd{domain.CardDesignHeart, 2},
			),
			multiplier: domain.TexasHoldemBonusAntePayFourOfAKind,
		},
		{
			name: "FullHouse",
			player: makeHandTHB(
				cd{domain.CardDesignSpade, 7},
				cd{domain.CardDesignClover, 7},
			),
			community: makeHandTHB(
				cd{domain.CardDesignHeart, 7},
				cd{domain.CardDesignDiamond, 5},
				cd{domain.CardDesignSpade, 5},
				cd{domain.CardDesignClover, 2},
				cd{domain.CardDesignHeart, 9},
			),
			multiplier: domain.TexasHoldemBonusAntePayFullHouse,
		},
		{
			name: "Flush",
			player: makeHandTHB(
				cd{domain.CardDesignSpade, 2},
				cd{domain.CardDesignSpade, 5},
			),
			community: makeHandTHB(
				cd{domain.CardDesignSpade, 7},
				cd{domain.CardDesignSpade, 9},
				cd{domain.CardDesignSpade, 11},
				cd{domain.CardDesignClover, 4},
				cd{domain.CardDesignHeart, 13},
			),
			multiplier: domain.TexasHoldemBonusAntePayFlush,
		},
		{
			name: "Straight",
			player: makeHandTHB(
				cd{domain.CardDesignSpade, 5},
				cd{domain.CardDesignClover, 6},
			),
			community: makeHandTHB(
				cd{domain.CardDesignHeart, 7},
				cd{domain.CardDesignDiamond, 8},
				cd{domain.CardDesignClover, 9},
				cd{domain.CardDesignSpade, 2},
				cd{domain.CardDesignHeart, 13},
			),
			multiplier: domain.TexasHoldemBonusAntePayStraight,
		},
		{
			name: "ThreeOfAKindNoBonus",
			player: makeHandTHB(
				cd{domain.CardDesignSpade, 7},
				cd{domain.CardDesignClover, 7},
			),
			community: makeHandTHB(
				cd{domain.CardDesignHeart, 7},
				cd{domain.CardDesignDiamond, 4},
				cd{domain.CardDesignSpade, 2},
				cd{domain.CardDesignClover, 9},
				cd{domain.CardDesignHeart, 12},
			),
			multiplier: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := domain.NewDefaultTexasHoldemBonus()
			g.SetChips(100000)
			require.NoError(t, g.Bet(100, 0))
			require.NoError(t, g.Play())
			g.SetPlayerHand(tt.player)
			g.SetDealerHand(weakDealer)
			g.SetCommunity(tt.community)
			g.ForceResolve()

			assert.Equal(t, domain.GameResultWin, g.GetResult())
			expectedAnte := 200 + 100*tt.multiplier // ante 1:1 win + ante bonus
			assert.Equal(t, expectedAnte, g.GetAntePayout())
		})
	}
}

// TestTexasHoldemBonus_AutoResolveFromCheck verifies that walking the public
// state machine via Bet→Play→Check→Check ends in PhaseEnd with a defined
// result. Card values are not asserted because turn/river are random.
func TestTexasHoldemBonus_AutoResolveFromCheck(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	g.SetChips(100000)
	playToShowdown(t, g, 100, 0)
	assert.Equal(t, domain.TexasHoldemBonusPhaseEnd, g.GetPhase())
	assert.True(t, g.GetGameEndFlag())
	// Result is one of Win/Lose/Draw.
	r := g.GetResult()
	assert.True(t, r == domain.GameResultWin || r == domain.GameResultLose || r == domain.GameResultDraw)
}

func TestTexasHoldemBonus_BonusSideBetMultipliers(t *testing.T) {
	tests := []struct {
		name       string
		holeCards  []*domain.Card
		multiplier int
	}{
		{
			name: "PocketAces",
			holeCards: makeHandTHB(
				cd{domain.CardDesignSpade, 1},
				cd{domain.CardDesignClover, 1},
			),
			multiplier: domain.TexasHoldemBonusBonusPayAA,
		},
		{
			name: "AKSuited",
			holeCards: makeHandTHB(
				cd{domain.CardDesignSpade, 1},
				cd{domain.CardDesignSpade, 13},
			),
			multiplier: domain.TexasHoldemBonusBonusPayAKSuited,
		},
		{
			name: "AQSuited",
			holeCards: makeHandTHB(
				cd{domain.CardDesignHeart, 1},
				cd{domain.CardDesignHeart, 12},
			),
			multiplier: domain.TexasHoldemBonusBonusPayAQAJSuited,
		},
		{
			name: "AJSuited",
			holeCards: makeHandTHB(
				cd{domain.CardDesignDiamond, 1},
				cd{domain.CardDesignDiamond, 11},
			),
			multiplier: domain.TexasHoldemBonusBonusPayAQAJSuited,
		},
		{
			name: "AKOff",
			holeCards: makeHandTHB(
				cd{domain.CardDesignSpade, 1},
				cd{domain.CardDesignClover, 13},
			),
			multiplier: domain.TexasHoldemBonusBonusPayAKOff,
		},
		{
			name: "PocketKings",
			holeCards: makeHandTHB(
				cd{domain.CardDesignSpade, 13},
				cd{domain.CardDesignClover, 13},
			),
			multiplier: domain.TexasHoldemBonusBonusPayKKQQJJ,
		},
		{
			name: "PocketJacks",
			holeCards: makeHandTHB(
				cd{domain.CardDesignSpade, 11},
				cd{domain.CardDesignClover, 11},
			),
			multiplier: domain.TexasHoldemBonusBonusPayKKQQJJ,
		},
		{
			name: "AQOff",
			holeCards: makeHandTHB(
				cd{domain.CardDesignSpade, 1},
				cd{domain.CardDesignClover, 12},
			),
			multiplier: domain.TexasHoldemBonusBonusPayAQAJOff,
		},
		{
			name: "AJOff",
			holeCards: makeHandTHB(
				cd{domain.CardDesignSpade, 1},
				cd{domain.CardDesignClover, 11},
			),
			multiplier: domain.TexasHoldemBonusBonusPayAQAJOff,
		},
		{
			name: "PocketTens",
			holeCards: makeHandTHB(
				cd{domain.CardDesignSpade, 10},
				cd{domain.CardDesignClover, 10},
			),
			multiplier: domain.TexasHoldemBonusBonusPayMediumPair,
		},
		{
			name: "PocketTwos",
			holeCards: makeHandTHB(
				cd{domain.CardDesignSpade, 2},
				cd{domain.CardDesignClover, 2},
			),
			multiplier: domain.TexasHoldemBonusBonusPayMediumPair,
		},
		{
			name: "ATSuitedNoBonus",
			holeCards: makeHandTHB(
				cd{domain.CardDesignHeart, 1},
				cd{domain.CardDesignHeart, 10},
			),
			multiplier: 0,
		},
		{
			name: "K9OffNoBonus",
			holeCards: makeHandTHB(
				cd{domain.CardDesignSpade, 13},
				cd{domain.CardDesignClover, 9},
			),
			multiplier: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := domain.NewDefaultTexasHoldemBonus()
			g.SetChips(100000)
			require.NoError(t, g.Bet(100, 10))
			g.SetPlayerHand(tt.holeCards)
			require.NoError(t, g.Fold())
			if tt.multiplier == 0 {
				assert.Equal(t, 0, g.GetBonusPayout())
				return
			}
			assert.Equal(t, 10+10*tt.multiplier, g.GetBonusPayout())
		})
	}
}

func TestTexasHoldemBonus_GetActionLog(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	require.NoError(t, g.Bet(100, 0))
	require.NoError(t, g.Fold())
	log := g.GetActionLog()
	assert.NotEmpty(t, log)
}

func TestTexasHoldemBonus_GetTotalPlayBet(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	g.SetFlopBet(200)
	g.SetTurnBet(100)
	g.SetRiverBet(100)
	assert.Equal(t, 400, g.GetTotalPlayBet())
}

func TestTexasHoldemBonus_GetTotalPayout(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	require.NoError(t, g.Bet(100, 10))
	// Force a non-paying hole-card combo so the bonus does not win on fold.
	g.SetPlayerHand(makeHandTHB(
		cd{domain.CardDesignSpade, 7},
		cd{domain.CardDesignClover, 2},
	))
	require.NoError(t, g.Fold())
	assert.Equal(t, 0, g.GetTotalPayout())
}

func TestTexasHoldemBonus_JSONRoundTrip(t *testing.T) {
	g := domain.NewDefaultTexasHoldemBonus()
	require.NoError(t, g.Bet(100, 10))
	require.NoError(t, g.Play())

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored domain.TexasHoldemBonus
	require.NoError(t, json.Unmarshal(data, &restored))

	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetAnteBet(), restored.GetAnteBet())
	assert.Equal(t, g.GetBonusBet(), restored.GetBonusBet())
	assert.Equal(t, g.GetFlopBet(), restored.GetFlopBet())
	assert.Equal(t, g.GetChips(), restored.GetChips())
	assert.Equal(t, len(g.GetPlayerHand()), len(restored.GetPlayerHand()))
	assert.Equal(t, len(g.GetDealerHand()), len(restored.GetDealerHand()))
	assert.Equal(t, len(g.GetCommunity()), len(restored.GetCommunity()))
}

func TestTexasHoldemBonus_JSONUnmarshal_TooLarge(t *testing.T) {
	// Build a JSON payload with an oversized PlayerHand array.
	bigCards := make([]map[string]any, 1001)
	for i := range bigCards {
		bigCards[i] = map[string]any{"d": 1, "v": 2, "f": false}
	}
	payload := map[string]any{"ph": bigCards}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored domain.TexasHoldemBonus
	err = json.Unmarshal(data, &restored)
	assert.Error(t, err)
}

// **画面に出す額と実際に引かれる額は同じ関数から来なければならない。**
// 別々に計算すると、表示だけ正しくて請求がずれる余地が残る (#4698)。
func TestTexasHoldemBonus_GetNextBetCost(t *testing.T) {
	newBetGame := func(t *testing.T) *domain.TexasHoldemBonus {
		t.Helper()
		g := domain.NewDefaultTexasHoldemBonus()
		g.Reset()
		require.NoError(t, g.Bet(100, 0))
		return g
	}

	t.Run("no cost before the ante is placed", func(t *testing.T) {
		g := domain.NewDefaultTexasHoldemBonus()
		g.Reset()
		assert.Equal(t, 0, g.GetNextBetCost())
	})

	t.Run("pre-flop Play costs twice the ante", func(t *testing.T) {
		assert.Equal(t, 200, newBetGame(t).GetNextBetCost())
	})

	// Play が実際に引いた額 = 直前に表示していた額。
	t.Run("Play charges exactly what it advertised", func(t *testing.T) {
		g := newBetGame(t)
		cost := g.GetNextBetCost()
		before := g.GetChips()
		require.NoError(t, g.Play())
		assert.Equal(t, before-cost, g.GetChips())
		assert.Equal(t, cost, g.GetFlopBet())
	})

	t.Run("flop and turn Raise cost one ante", func(t *testing.T) {
		g := newBetGame(t)
		require.NoError(t, g.Play())
		assert.Equal(t, 100, g.GetNextBetCost())

		cost := g.GetNextBetCost()
		before := g.GetChips()
		require.NoError(t, g.Raise())
		assert.Equal(t, before-cost, g.GetChips())
		assert.Equal(t, 100, g.GetNextBetCost(), "ターンも1×アンテ")
	})

	t.Run("no cost once the round is resolved", func(t *testing.T) {
		g := newBetGame(t)
		require.NoError(t, g.Play())
		require.NoError(t, g.Raise())
		require.NoError(t, g.Raise())
		assert.Equal(t, 0, g.GetNextBetCost())
	})
}
