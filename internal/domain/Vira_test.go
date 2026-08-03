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

func TestVira_BidMustOutrankTheStanding(t *testing.T) {
	g := newTestVira(t)
	g.SetCurrentPlayerIdx(0)
	require.NoError(t, g.PlayerBid(ViraBidSolo))

	// 同値も下位も通らない。通すと席順だけで宣言者が決まる。
	g.SetCurrentPlayerIdx(0)
	g2 := newTestVira(t)
	g2.SetCurrentPlayerIdx(0)
	require.NoError(t, g2.PlayerBid(ViraBidVira))
}

func TestVira_BidLadderRejectsEqualOrLower(t *testing.T) {
	g := NewDefaultVira()
	g.SetRand(rand.New(rand.NewSource(7)))
	g.Reset()

	// 席 0 が Solo。席 1 が Gask (下位) と Solo (同値) を出しても通らない。
	g.SetCurrentPlayerIdx(0)
	require.NoError(t, g.PlayerBid(ViraBidSolo))
	assert.Equal(t, ViraBidSolo, g.GetBids()[0])
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
