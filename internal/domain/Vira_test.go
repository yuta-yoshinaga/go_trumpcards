//go:build test

package domain

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestVira は配り札を固定した Vira を返す。
//
// **Reset 直後の手札に依存する主張は必ずシードを固定する。**固定しないと
// 4/52 のような低確率で落ちるテストになり、develop を赤くする (#4467)。
func newTestVira(t *testing.T) *Vira {
	t.Helper()
	g := NewDefaultVira()
	g.SetRand(rand.New(rand.NewSource(42)))
	g.Reset()
	return g
}

func TestVira_ResetDealsThirteenEach(t *testing.T) {
	g := newTestVira(t)
	assert.Equal(t, ViraPhaseBid, g.GetPhase())
	for i := 0; i < ViraPlayerCnt; i++ {
		assert.Equal(t, ViraHandSize, g.GetPlayer(i).GetCardsSize(), "席 %d", i)
	}
}

func TestVira_ResetTakesAnteFromEveryone(t *testing.T) {
	g := newTestVira(t)
	// アンティは全員から等しく取る。ポットはその合計。
	assert.Equal(t, ViraAnte*ViraPlayerCnt, g.GetPot())
	for i, s := range g.GetPlayerScores() {
		assert.Equal(t, -ViraAnte, s, "席 %d", i)
	}
}

// **同値も下位も通らない。**通すと席順だけで宣言者が決まり、階梯が意味を失う。
//
// 以前ここにあった 2 本は、名前もコメントも「上回らない入札を弾く」と言いながら
// 実際には下位入札を一度も試していなかった —— 別インスタンスで上位入札が通る
// ことを確かめていただけで、`applyBid` の拒否分岐は踏んでいない。
func TestVira_BidMustOutrankTheStanding(t *testing.T) {
	cases := map[string]struct {
		standing, attempt ViraBid
		wantErr           bool
	}{
		"lower is refused":       {ViraBidSolo, ViraBidGask, true},
		"equal is refused":       {ViraBidSolo, ViraBidSolo, true},
		"higher is accepted":     {ViraBidSolo, ViraBidMisere, false},
		"pass is always allowed": {ViraBidVira, ViraBidPass, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			g := newTestVira(t)
			g.SetCurrentPlayerIdx(0)
			require.NoError(t, g.PlayerBid(tc.standing))

			// 席 1 に手番を渡してから試す。同じ席では「入札済み」で弾かれ、
			// 階梯ではなく重複が理由になってしまう。
			g.SetCurrentPlayerIdx(1)
			err := g.applyBid(1, tc.attempt)
			if tc.wantErr {
				require.Error(t, err, "%v must not stand against %v", tc.attempt, tc.standing)
				assert.Equal(t, ViraBidPass, g.GetBids()[1], "a refused bid must not be recorded")
				assert.False(t, g.GetBidDone()[1], "a refused bid must not close the seat")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.attempt, g.GetBids()[1])
			assert.True(t, g.GetBidDone()[1])
		})
	}
}

func TestVira_BidGuards(t *testing.T) {
	t.Run("out of range", func(t *testing.T) {
		g := newTestVira(t)
		assert.Error(t, g.applyBid(0, ViraBidPass-1))
		assert.Error(t, g.applyBid(0, ViraBidVira+1))
	})
	t.Run("a seat cannot bid twice", func(t *testing.T) {
		g := newTestVira(t)
		g.SetCurrentPlayerIdx(0)
		require.NoError(t, g.PlayerBid(ViraBidGask))
		assert.Error(t, g.applyBid(0, ViraBidVira), "the seat has already bid")
	})
}

func TestVira_AllPassCarriesThePotForward(t *testing.T) {
	// **全員 CPU にして手を空にする。**CPU の入札は手札の強さで決まるので、
	// 実在の手を配ったままでは「全パス」を確実に作れない。空の手なら
	// 絵札 0 枚・最長スート 0 で Misère を狙うため、Easy に落として
	// パスさせる。Easy でも階梯の検証は変わらない。
	players := make([]*ViraPlayer, ViraPlayerCnt)
	for i := range players {
		players[i] = NewViraPlayer(false)
	}
	g := NewVira(NewTrumpCards(0), players,
		ViraConfig{CpuDifficulty: ViraCpuDifficultyNormal, TargetRounds: 6})
	g.SetRand(rand.New(rand.NewSource(42)))
	g.Reset()

	for i := 0; i < ViraPlayerCnt; i++ {
		g.SetCurrentPlayerIdx(i)
		require.NoError(t, g.ForcePassForTest(i))
	}
	assert.Equal(t, ViraPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, -1, g.GetDeclarerIdx(), "全パスなら宣言者はいない")
	// **ポットは戻さない。**次局へ持ち越すのがポット式の要。
	assert.Equal(t, ViraAnte*ViraPlayerCnt, g.GetPot())
}

func TestVira_BidTargetsMatchTheLadder(t *testing.T) {
	// 目標トリック数は階梯の順に上がる。Misère だけが 0。
	cases := []struct {
		bid    ViraBid
		target int
		value  int
	}{
		{ViraBidGask, 7, 2},
		{ViraBidSolo, 8, 4},
		{ViraBidMisere, 0, 6},
		{ViraBidVira, 10, 8},
	}
	for _, c := range cases {
		g := newTestVira(t)
		g.SetCurrentPlayerIdx(0)
		require.NoError(t, g.PlayerBid(c.bid))
		assert.Equal(t, c.bid, g.GetBids()[0])
	}
}

func TestVira_TrumpBeatsTheLedSuit(t *testing.T) {
	g := newTestVira(t)
	// 入札を通してプレイフェーズへ。
	g.SetCurrentPlayerIdx(0)
	require.NoError(t, g.PlayerBid(ViraBidSolo))
	// 席 1・2 は CPU。PlayerBid は人間の手番でしか通らないので CpuBid で進める。
	for i := 1; i < ViraPlayerCnt; i++ {
		g.SetCurrentPlayerIdx(i)
		g.CpuBid()
	}
	require.Equal(t, ViraPhasePlay, g.GetPhase())
	assert.Equal(t, 0, g.GetDeclarerIdx())
	assert.NotEqual(t, 0, g.GetTrumpSuit(), "Solo は切り札を持つ")
}

func TestVira_MisereHasNoTrump(t *testing.T) {
	g := newTestVira(t)
	g.SetCurrentPlayerIdx(0)
	require.NoError(t, g.PlayerBid(ViraBidMisere))
	// 席 1・2 は CPU。PlayerBid は人間の手番でしか通らないので CpuBid で進める。
	for i := 1; i < ViraPlayerCnt; i++ {
		g.SetCurrentPlayerIdx(i)
		g.CpuBid()
	}
	// **Misère に切り札は無い。**持たせると 0 トリックが不可能に近くなる。
	assert.Equal(t, 0, g.GetTrumpSuit())
}

func TestVira_ValidPlayIndicesEnforceFollow(t *testing.T) {
	g := newTestVira(t)
	g.SetCurrentPlayerIdx(0)
	require.NoError(t, g.PlayerBid(ViraBidSolo))
	// 席 1・2 は CPU。PlayerBid は人間の手番でしか通らないので CpuBid で進める。
	for i := 1; i < ViraPlayerCnt; i++ {
		g.SetCurrentPlayerIdx(i)
		g.CpuBid()
	}
	lead := g.GetLeadPlayerIdx()
	// リード前は全部出せる。
	assert.Len(t, g.GetValidPlayIndices(lead), ViraHandSize)
}

func TestVira_ConfigRejectsRoundsNotDivisibleByPlayers(t *testing.T) {
	// **局数はプレイヤー数の倍数。**そうでないと各プレイヤーがディーラーを
	// 務めた回数が揃わず、配り順の有利不利が精算に残る。
	assert.Error(t, ViraConfig{CpuDifficulty: ViraCpuDifficultyNormal, TargetRounds: 5}.Validate())
	assert.NoError(t, ViraConfig{CpuDifficulty: ViraCpuDifficultyNormal, TargetRounds: 6}.Validate())
}

func TestVira_ConfigRejectsOutOfRangeDifficulty(t *testing.T) {
	assert.Error(t, ViraConfig{CpuDifficulty: 99, TargetRounds: 6}.Validate())
}

// The whole game state lives in unexported fields, so a session that round-trips
// through KV on the Worker only survives if the custom codec carries it. Plain
// json.Marshal on the struct would emit "{}" and silently reset every hand.
func TestVira_JSONRoundTripPreservesState(t *testing.T) {
	src := NewDefaultVira()
	src.SetRand(rand.New(rand.NewSource(11)))
	src.Reset()

	data, err := src.MarshalJSON()
	assert.NoError(t, err)

	var got Vira
	assert.NoError(t, got.UnmarshalJSON(data))
	assert.Equal(t, src.GetPhase(), got.GetPhase())
	assert.Equal(t, src.GetDealerIdx(), got.GetDealerIdx())
	assert.Equal(t, src.GetPlayerScores(), got.GetPlayerScores())
	assert.Equal(t, ViraPlayerCnt, got.GetPlayerCnt())
	for i := 0; i < ViraPlayerCnt; i++ {
		assert.Equal(t, src.GetPlayer(i).GetCardsSize(), got.GetPlayer(i).GetCardsSize())
	}
}

// The pot carries forward between rounds — including through an all-pass round —
// so dropping it from the wire format would reset the accumulated stake on every
// Worker request, which is the one number a Vira player is tracking.
func TestVira_JSONRoundTripPreservesThePot(t *testing.T) {
	src := NewDefaultVira()
	src.SetRand(rand.New(rand.NewSource(11)))
	src.Reset()
	assert.Positive(t, src.GetPot(), "the ante must have seeded the pot")

	data, err := src.MarshalJSON()
	assert.NoError(t, err)
	var got Vira
	assert.NoError(t, got.UnmarshalJSON(data))
	assert.Equal(t, src.GetPot(), got.GetPot())
}

func TestVira_UnmarshalRejectsBadState(t *testing.T) {
	cases := map[string]string{
		"not json":              `{`,
		"no players":            `{"pl":[],"rn":1,"tn":1}`,
		"negative pot":          `{"pl":[{},{},{}],"rn":1,"tn":1,"pot":-1,"cfg":{"cd":1,"tr":6}}`,
		"trick number too high": `{"pl":[{},{},{}],"rn":1,"tn":99,"cfg":{"cd":1,"tr":6}}`,
		"bid out of range":      `{"pl":[{},{},{}],"rn":1,"tn":1,"bd":[9,0,0],"cfg":{"cd":1,"tr":6}}`,
		"nil trick card":        `{"pl":[{},{},{}],"rn":1,"tn":1,"ct":[null],"cfg":{"cd":1,"tr":6}}`,
		"rounds not a multiple": `{"pl":[{},{},{}],"rn":1,"tn":1,"cfg":{"cd":1,"tr":7}}`,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			var got Vira
			assert.Error(t, got.UnmarshalJSON([]byte(payload)))
		})
	}
}

// setViraHand replaces a player's hand with exactly the given cards.
func setViraHand(t *testing.T, g *Vira, idx int, cards ...*Card) {
	t.Helper()
	p := g.GetPlayer(idx)
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	for _, c := range cards {
		p.AddCard(c)
	}
}

// settleAllBids drives the bidding to a close from wherever it stands.
func settleAllBids(t *testing.T, g *Vira) {
	t.Helper()
	for range ViraPlayerCnt * 2 {
		if g.GetPhase() != ViraPhaseBid {
			return
		}
		if g.IsHumanBidTurn() {
			require.NoError(t, g.PlayerBid(ViraBidPass))
			continue
		}
		g.CpuBid()
	}
}

// playOneViraTrick drives a full three-card trick from the current lead.
func playOneViraTrick(t *testing.T, g *Vira) {
	t.Helper()
	for range ViraPlayerCnt {
		if g.IsHumanTurn() {
			idx := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
			require.NotEmpty(t, idx)
			require.NoError(t, g.PlayerPlay(idx[0]))
			continue
		}
		g.CpuPlay()
	}
}

// The bid ladder and the value table are two separate numbers per contract, and
// the CUI reads both. Pinning them stops a rename from quietly changing payouts.
func TestVira_BidLadderAndValues(t *testing.T) {
	cases := []struct {
		bid           ViraBid
		target, value int
	}{
		{ViraBidPass, 0, 0},
		{ViraBidGask, 7, 2},
		{ViraBidSolo, 8, 4},
		{ViraBidMisere, 0, 6}, // Misère targets zero tricks but is worth more than Gask
		{ViraBidVira, 10, 8},
	}
	for _, tc := range cases {
		t.Run(ViraBidNames[tc.bid], func(t *testing.T) {
			assert.Equal(t, tc.target, ViraBidTarget(tc.bid))
			assert.Equal(t, tc.value, viraBidValue(tc.bid))
		})
	}
}

func TestVira_TrickWinner(t *testing.T) {
	cases := []struct {
		name  string
		trump int
		trick []*TrickCard
		want  int
	}{
		{"highest of the led suit wins", 0, []*TrickCard{
			{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 5, false)},
			{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 13, false)},
			{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 9, false)},
		}, 1},
		// Ace is high, so a 1 must beat a king rather than lose to it.
		{"the ace outranks the king", 0, []*TrickCard{
			{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 13, false)},
			{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 1, false)},
			{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 12, false)},
		}, 1},
		{"the lowest trump beats the highest plain card", CardDesignClover, []*TrickCard{
			{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 1, false)},
			{PlayerIdx: 1, Card: NewCard(CardDesignClover, 2, false)},
			{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 13, false)},
		}, 1},
		{"an off-suit discard never wins", CardDesignClover, []*TrickCard{
			{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 3, false)},
			{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 1, false)},
			{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 2, false)},
		}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newTestVira(t)
			g.trumpSuit = tc.trump
			g.currentTrick = tc.trick
			assert.Equal(t, tc.want, g.trickWinner())
		})
	}
}

func TestVira_PlayerPlayRejectsRenege(t *testing.T) {
	g := newTestVira(t)
	g.SetPhase(ViraPhasePlay)
	g.SetCurrentPlayerIdx(0)
	setViraHand(t, g, 0, NewCard(CardDesignSpade, 5, false), NewCard(CardDesignHeart, 9, false))
	g.currentTrick = []*TrickCard{{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 8, false)}}

	assert.Error(t, g.PlayerPlay(1), "holding the led suit, an off-suit card must be refused")
	assert.NoError(t, g.PlayerPlay(0))
}

func TestVira_PlayerPlayGuards(t *testing.T) {
	t.Run("wrong phase", func(t *testing.T) {
		g := newTestVira(t)
		assert.Error(t, g.PlayerPlay(0), "the bid phase is not a play phase")
	})
	t.Run("index out of range", func(t *testing.T) {
		g := newTestVira(t)
		g.SetPhase(ViraPhasePlay)
		g.SetCurrentPlayerIdx(0)
		assert.Error(t, g.PlayerPlay(-1))
		assert.Error(t, g.PlayerPlay(ViraHandSize))
	})
	t.Run("not the human turn", func(t *testing.T) {
		g := newTestVira(t)
		g.SetPhase(ViraPhasePlay)
		g.SetCurrentPlayerIdx(1)
		assert.Error(t, g.PlayerPlay(0))
	})
	t.Run("game already ended", func(t *testing.T) {
		g := newTestVira(t)
		g.SetPhase(ViraPhasePlay)
		g.gameEndFlag = true
		assert.Error(t, g.PlayerPlay(0))
	})
}

// A CPU with an empty hand must not be asked to play: RemoveCard returns nil and
// passing that on would nil-deref the whole request (#4606).
func TestVira_CpuPlayWithAnEmptyHandIsANoOp(t *testing.T) {
	g := newTestVira(t)
	g.SetPhase(ViraPhasePlay)
	g.SetCurrentPlayerIdx(1)
	setViraHand(t, g, 1)
	assert.NotPanics(t, func() { g.CpuPlay() })
	assert.Empty(t, g.GetCurrentTrick())
}

func TestVira_CpuPlayIgnoresAHumanSeatAndAFinishedGame(t *testing.T) {
	g := newTestVira(t)
	g.SetPhase(ViraPhasePlay)
	g.SetCurrentPlayerIdx(0)
	g.CpuPlay()
	assert.Empty(t, g.GetCurrentTrick())

	g.SetCurrentPlayerIdx(1)
	g.gameEndFlag = true
	g.CpuPlay()
	assert.Empty(t, g.GetCurrentTrick())
}

func TestVira_TrickFlowAwardsAndLeadsFromTheWinner(t *testing.T) {
	g := newTestVira(t)
	settleAllBids(t, g)
	require.Equal(t, ViraPhasePlay, g.GetPhase())

	before := g.GetTrickNumber()
	playOneViraTrick(t, g)
	require.Equal(t, ViraPhaseTrickEnd, g.GetPhase())
	g.ResolveTrick()

	won := 0
	for i := range ViraPlayerCnt {
		won += g.GetRoundTricks()[i]
	}
	assert.Equal(t, 1, won)
	assert.Equal(t, g.GetLeadPlayerIdx(), g.GetCurrentPlayerIdx(), "the winner leads next")

	g.NextTrick()
	assert.Equal(t, ViraPhasePlay, g.GetPhase())
	assert.Equal(t, before+1, g.GetTrickNumber())
	assert.Empty(t, g.GetCurrentTrick())
}

func TestVira_ResolveAndNextTrickAreNoOpsOutsideTrickEnd(t *testing.T) {
	g := newTestVira(t)
	g.SetPhase(ViraPhasePlay)
	g.ResolveTrick()
	g.NextTrick()
	assert.Equal(t, ViraPhasePlay, g.GetPhase())
	assert.Equal(t, 1, g.GetTrickNumber())
}

// Settlement is where the pot does its work, and made/failed move it in opposite
// directions. The declarer sweeps it on success and feeds it on failure.
func TestVira_SettlementMovesThePot(t *testing.T) {
	t.Run("a made contract sweeps the pot and collects from each defender", func(t *testing.T) {
		g := newTestVira(t)
		g.declarerIdx = 0
		g.contract = ViraBidGask
		g.roundTricks = [ViraPlayerCnt]int{7, 3, 3}
		g.pot = 30
		g.playerScores = [ViraPlayerCnt]int{100, 100, 100}
		g.settleRound()

		value := viraBidValue(ViraBidGask)
		assert.True(t, g.GetLastRoundMade())
		assert.Zero(t, g.GetPot(), "a made contract empties the pot")
		assert.Equal(t, 100+30+2*value, g.GetPlayerScores()[0])
		assert.Equal(t, 100-value, g.GetPlayerScores()[1])
		assert.Equal(t, 30+2*value, g.GetLastRoundDelta()[0])
	})

	t.Run("a failed contract feeds the pot and pays each defender", func(t *testing.T) {
		g := newTestVira(t)
		g.declarerIdx = 0
		g.contract = ViraBidVira
		g.roundTricks = [ViraPlayerCnt]int{4, 5, 4}
		g.pot = 30
		g.playerScores = [ViraPlayerCnt]int{100, 100, 100}
		g.settleRound()

		value := viraBidValue(ViraBidVira)
		assert.False(t, g.GetLastRoundMade())
		assert.Equal(t, 30+value, g.GetPot(), "the failed contract's value is added to the pot")
		assert.Equal(t, 100-3*value, g.GetPlayerScores()[0])
		assert.Equal(t, 100+value, g.GetPlayerScores()[1])
	})

	// Misère inverts the test: it is made on zero tricks and lost on the first.
	t.Run("misere is made on zero tricks", func(t *testing.T) {
		g := newTestVira(t)
		g.declarerIdx = 1
		g.contract = ViraBidMisere
		g.roundTricks = [ViraPlayerCnt]int{7, 0, 6}
		g.settleRound()
		assert.True(t, g.GetLastRoundMade())
	})

	t.Run("misere fails on a single trick", func(t *testing.T) {
		g := newTestVira(t)
		g.declarerIdx = 1
		g.contract = ViraBidMisere
		g.roundTricks = [ViraPlayerCnt]int{7, 1, 5}
		g.settleRound()
		assert.False(t, g.GetLastRoundMade())
	})

	// An all-pass round has no declarer, so nothing changes hands — but the ante
	// already in the pot must stay there for the next round to compete over.
	t.Run("an all-pass round leaves the pot alone", func(t *testing.T) {
		g := newTestVira(t)
		g.declarerIdx = -1
		g.pot = 45
		g.playerScores = [ViraPlayerCnt]int{100, 100, 100}
		g.settleRound()
		assert.Equal(t, 45, g.GetPot(), "the stake carries forward")
		assert.Equal(t, [ViraPlayerCnt]int{100, 100, 100}, g.GetPlayerScores())
	})
}

func TestVira_MadeLabel(t *testing.T) {
	assert.Equal(t, "成功", viraMadeLabel(true))
	assert.Equal(t, "失敗", viraMadeLabel(false))
}

func TestVira_NextRoundRedealsAndRotatesTheDealer(t *testing.T) {
	g := newTestVira(t)
	g.SetPhase(ViraPhaseRoundEnd)
	dealer, round := g.GetDealerIdx(), g.GetRoundNumber()

	g.NextRound()
	assert.Equal(t, round+1, g.GetRoundNumber())
	assert.Equal(t, (dealer+1)%ViraPlayerCnt, g.GetDealerIdx())
	assert.Equal(t, ViraPhaseBid, g.GetPhase())
	assert.Equal(t, -1, g.GetDeclarerIdx())
	for i := range ViraPlayerCnt {
		assert.Equal(t, ViraHandSize, g.GetPlayer(i).GetCardsSize())
		assert.Zero(t, g.GetRoundTricks()[i])
	}
}

func TestVira_NextRoundEndsTheMatchOnTheLastRound(t *testing.T) {
	g := newTestVira(t)
	g.SetPhase(ViraPhaseRoundEnd)
	g.roundNumber = g.GetConfig().TargetRounds
	g.NextRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, ViraPhaseGameEnd, g.GetPhase())
}

func TestVira_NextRoundAndScoreRoundAreNoOpsOutsideRoundEnd(t *testing.T) {
	g := newTestVira(t)
	g.SetPhase(ViraPhasePlay)
	g.roundNumber = g.GetConfig().TargetRounds
	g.NextRound()
	g.ScoreRound()
	assert.False(t, g.GetGameEndFlag())
	assert.Equal(t, g.GetConfig().TargetRounds, g.GetRoundNumber(), "neither call advanced the round")
}

func TestVira_ScoreRoundEndsTheMatchOnTheLastRound(t *testing.T) {
	g := newTestVira(t)
	g.SetPhase(ViraPhaseRoundEnd)

	g.roundNumber = g.GetConfig().TargetRounds - 1
	g.ScoreRound()
	assert.False(t, g.GetGameEndFlag())

	g.roundNumber = g.GetConfig().TargetRounds
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
}

func TestVira_MatchWinner(t *testing.T) {
	cases := map[string]struct {
		scores [ViraPlayerCnt]int
		want   int
	}{
		"clear leader":        {[ViraPlayerCnt]int{140, 90, 70}, 0},
		"leader in last seat": {[ViraPlayerCnt]int{70, 90, 140}, 2},
		// A tie must leave no winner: breaking it by seat order would hand the
		// match to the lowest seat every time.
		"two-way tie at the top": {[ViraPlayerCnt]int{140, 140, 20}, -1},
		"three-way tie":          {[ViraPlayerCnt]int{100, 100, 100}, -1},
		"tie below the leader":   {[ViraPlayerCnt]int{140, 80, 80}, 0},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			g := newTestVira(t)
			g.playerScores = tc.scores
			g.finishMatch()
			assert.Equal(t, tc.want, g.GetWinnerPlayer())
			assert.True(t, g.GetGameEndFlag())
		})
	}
}

// Each of the six reasons is reachable, and Misère's two are the opposite
// advice for the two sides of the same contract.
func TestVira_HintReasons(t *testing.T) {
	cases := []struct {
		name       string
		contract   ViraBid
		declarer   int
		trickCards int
		want       string
	}{
		{"declarer leads high", ViraBidGask, 0, 0, "lead_high"},
		{"defender leads low", ViraBidGask, 1, 0, "lead_low"},
		{"declarer follows to win", ViraBidGask, 0, 1, "follow_win"},
		{"defender blocks", ViraBidGask, 1, 1, "follow_block"},
		{"misere declarer ducks", ViraBidMisere, 0, 0, "misere_duck"},
		{"misere defender forces", ViraBidMisere, 1, 0, "misere_force"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newTestVira(t)
			g.contract = tc.contract
			g.declarerIdx = tc.declarer
			g.currentTrick = nil
			for i := range tc.trickCards {
				g.currentTrick = append(g.currentTrick,
					&TrickCard{PlayerIdx: i + 1, Card: NewCard(CardDesignSpade, 3, false)})
			}
			assert.Equal(t, tc.want, g.playHintReason(0))
		})
	}
}

func TestVira_GetHintOnlyOnTheHumanPlayTurn(t *testing.T) {
	g := newTestVira(t)
	g.SetPhase(ViraPhaseBid)
	assert.Nil(t, g.GetHint(), "no play hint during bidding")

	g.SetPhase(ViraPhasePlay)
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint(), "no hint on a CPU turn")

	g.SetCurrentPlayerIdx(0)
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Len(t, hint.CardIndices, 1)
	assert.NotEmpty(t, hint.Reason)

	setViraHand(t, g, 0)
	assert.Nil(t, g.GetHint(), "an empty hand has nothing to suggest")
}

func TestVira_IsHumanBidTurn(t *testing.T) {
	g := newTestVira(t)
	g.SetPhase(ViraPhaseBid)
	g.SetCurrentPlayerIdx(0)
	assert.True(t, g.IsHumanBidTurn())

	require.NoError(t, g.PlayerBid(ViraBidPass))
	g.SetCurrentPlayerIdx(0)
	assert.False(t, g.IsHumanBidTurn(), "a seat that has bid does not bid again")
	assert.True(t, g.GetBidDone()[0])

	g.SetPhase(ViraPhasePlay)
	assert.False(t, g.IsHumanBidTurn())
}

func TestVira_ConfigAndAccessors(t *testing.T) {
	g := newTestVira(t)
	cfg := ViraConfig{CpuDifficulty: ViraCpuDifficultyHard, TargetRounds: 9}
	g.SetConfig(cfg)
	assert.Equal(t, cfg, g.GetConfig())
	assert.Len(t, g.GetPlayers(), ViraPlayerCnt)
	assert.Equal(t, ViraBidPass, g.GetContract())
	assert.NotNil(t, g.GetActionLog())
}

func TestVira_ValidPlayIndicesFallBackToTheWholeHandWhenVoid(t *testing.T) {
	g := newTestVira(t)
	setViraHand(t, g, 0, NewCard(CardDesignHeart, 4, false), NewCard(CardDesignClover, 9, false))
	g.currentTrick = []*TrickCard{{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 8, false)}}
	assert.Equal(t, []int{0, 1}, g.GetPlayableIndices(0))
}
