//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newBauernschnapsen() *domain.Bauernschnapsen { return domain.NewDefaultBauernschnapsen() }

func TestBauernschnapsen_DeckAndPoints(t *testing.T) {
	g := newBauernschnapsen()
	deck := g.GetConfigDeckHelper()
	// **20 枚 1 組。** クローン元のガイゲルは 7 を含む 24 枚を 2 組重ねた 48 枚。
	assert.Equal(t, 20, deck.GetTotalCount())

	// Card points.
	assert.Equal(t, 11, domain.BauernschnapsenCardPoints(domain.NewCard(domain.CardDesignSpade, 1, false)))
	assert.Equal(t, 10, domain.BauernschnapsenCardPoints(domain.NewCard(domain.CardDesignSpade, 10, false)))
	assert.Equal(t, 4, domain.BauernschnapsenCardPoints(domain.NewCard(domain.CardDesignSpade, 13, false)))
	assert.Equal(t, 3, domain.BauernschnapsenCardPoints(domain.NewCard(domain.CardDesignSpade, 12, false)))
	assert.Equal(t, 2, domain.BauernschnapsenCardPoints(domain.NewCard(domain.CardDesignSpade, 11, false)))
	assert.Equal(t, 0, domain.BauernschnapsenCardPoints(nil))

	// Total = 120 (シュナプセン / 66 と同じ)。**7 は配らないので 0 点札は無い。**
	total := (11 + 10 + 4 + 3 + 2) * 4
	assert.Equal(t, 120, total)
	assert.Equal(t, 120, domain.BauernschnapsenRoundCardPointsTotal)
}

// TestBauernschnapsen_DeckPointsAddUp は、**実際に配られる札の合計**が
// 定数と一致することを見る。
//
// 個別のケースを並べるだけでは、デッキに入っているのに点数表に無い額面
// (クローン元の 7 のような) を見落とす。デッキ側から数える。
func TestBauernschnapsen_DeckPointsAddUp(t *testing.T) {
	g := newBauernschnapsen()
	deck := g.GetConfigDeckHelper()
	seen, total := 0, 0
	for {
		c := deck.DrawCard()
		if c == nil {
			break
		}
		seen++
		total += domain.BauernschnapsenCardPoints(c)
	}
	assert.Equal(t, 20, seen, "20 枚配り切る")
	assert.Equal(t, domain.BauernschnapsenRoundCardPointsTotal, total,
		"デッキ全体の点が定数と食い違う —— 点のある札が盤から消えているか、無い札を数えている")
}

func TestBauernschnapsen_RankOrder(t *testing.T) {
	// A>10>K>Q>J>7
	a := domain.BauernschnapsenRankOrder(domain.NewCard(domain.CardDesignSpade, 1, false))
	ten := domain.BauernschnapsenRankOrder(domain.NewCard(domain.CardDesignSpade, 10, false))
	k := domain.BauernschnapsenRankOrder(domain.NewCard(domain.CardDesignSpade, 13, false))
	q := domain.BauernschnapsenRankOrder(domain.NewCard(domain.CardDesignSpade, 12, false))
	j := domain.BauernschnapsenRankOrder(domain.NewCard(domain.CardDesignSpade, 11, false))
	seven := domain.BauernschnapsenRankOrder(domain.NewCard(domain.CardDesignSpade, 7, false))
	assert.True(t, a > ten && ten > k && k > q && q > j && j > seven)
	assert.Equal(t, 0, domain.BauernschnapsenRankOrder(nil))
}

func TestBauernschnapsen_ResetDeal(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	// **配り終えたら契約フェーズ。** 切り札は表向きの札ではなく宣言で決まる。
	assert.Equal(t, domain.BauernschnapsenPhaseContract, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())

	// 4 players * 5 cards = 20 = デッキ全部。
	totalHand := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		assert.Equal(t, 5, g.GetPlayer(i).GetCardsSize(), "seat %d", i)
		totalHand += g.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 20, totalHand)

	// **山札も切り札表示カードも残らない。** クローン元のガイゲルは 48 枚から
	// 20 枚配って 27 枚の山札と 1 枚の切り札表示カードが残っていた。
	// 配った枚数がデッキ全部と一致する = 山に 1 枚も残らない、と読む。
	// 手で書いた 20 ではなく**デッキから数える**ので、デッキを変えれば追随する。
	assert.Equal(t, g.GetConfigDeckHelper().GetTotalCount(), totalHand, "山札は残らない")
	assert.Equal(t, domain.BauernschnapsenNoTrump, g.GetTrumpSuit(), "切り札は宣言で決まる")

	// 山札が無いので最初から追従必須。
	assert.True(t, g.IsEndgame(), "1 トリック目から追従必須")
}

func TestBauernschnapsen_TrickWinner_DoubleDeckTie(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignHeart) // trump = heart, lead = spade
	// Two identical SPADE Aces: the earlier (player 0) wins the tie.
	a1 := domain.NewCard(domain.CardDesignSpade, 1, false)
	a2 := domain.NewCard(domain.CardDesignSpade, 1, false)
	low := domain.NewCard(domain.CardDesignSpade, 7, false)
	low2 := domain.NewCard(domain.CardDesignSpade, 7, false)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: a1},
		{PlayerIdx: 1, Card: low},
		{PlayerIdx: 2, Card: a2},
		{PlayerIdx: 3, Card: low2},
	})
	g.SetTrickNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.BauernschnapsenPhaseTrickEnd)
	g.ResolveTrick()
	assert.Equal(t, 0, g.GetLeadPlayerIdx(), "earlier identical card should win")
}

func TestBauernschnapsen_TrumpBeatsNonTrump(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignDiamond)
	// Lead spade Ace, player 2 plays trump 7 -> trump wins.
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 10, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignDiamond, 7, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignSpade, 13, false)},
	})
	g.SetTrickNumber(1)
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.BauernschnapsenPhaseTrickEnd)
	g.ResolveTrick()
	assert.Equal(t, 2, g.GetLeadPlayerIdx())
}

func TestBauernschnapsen_Marriage(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignHeart)
	// Give player 0 a non-trump marriage (spade K+Q) on lead.
	p0 := g.GetPlayer(0)
	p0.Reset()
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false)) // Q idx0
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false)) // K idx1
	p0.AddCard(domain.NewCard(domain.CardDesignClover, 7, false))
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	g.SetPhase(domain.BauernschnapsenPhasePlay)

	idxs := g.GetMarriageIndices(0)
	require.NotEmpty(t, idxs)

	err := g.PlayerDeclareMarriage(0) // declare via Q
	require.NoError(t, err)
	// Non-trump marriage = 20 to team 0.
	assert.Equal(t, 20, g.GetRoundMarriagePoints(0))
	// The Q was led.
	assert.Len(t, g.GetCurrentTrick(), 1)
	// Re-declaring the same suit is now blocked.
	assert.Empty(t, g.GetMarriageIndices(0))
}

func TestBauernschnapsen_RoyalMarriage(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignSpade)
	p0 := g.GetPlayer(0)
	p0.Reset()
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false)) // trump Q
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false)) // trump K
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	g.SetPhase(domain.BauernschnapsenPhasePlay)
	require.NoError(t, g.PlayerDeclareMarriage(0))
	assert.Equal(t, 40, g.GetRoundMarriagePoints(0))
}

func TestBauernschnapsen_MarriageRejectsNonStarter(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignHeart)
	p0 := g.GetPlayer(0)
	p0.Reset()
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false)) // not a K/Q
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	g.SetPhase(domain.BauernschnapsenPhasePlay)
	assert.Error(t, g.PlayerDeclareMarriage(0))
	assert.Error(t, g.PlayerDeclareMarriage(99)) // out of range
}

// TestBauernschnapsen_FollowIsRequiredFromTheFirstTrick は、**最初から**
// 追従必須であることを見る。
//
// クローン元のガイゲルは山札がある間は自由出しで、尽きてから追従必須の
// 第 2 フェーズに入る二相構造だった。20 枚を配り切るこのゲームに山札は無く、
// 「自由に出せる前半」が存在しない。
func TestBauernschnapsen_FollowIsRequiredFromTheFirstTrick(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignHeart)
	p0 := g.GetPlayer(0)
	p0.Reset()
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))   // リードスート
	p0.AddCard(domain.NewCard(domain.CardDesignClover, 11, false)) // 別スート
	p0.AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))  // 切り札
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 13, false)},
	})
	g.SetPhase(domain.BauernschnapsenPhasePlay)

	assert.True(t, g.IsEndgame())
	// リードスートを持っているので、それしか出せない。
	valid := g.GetValidPlayIndices(0)
	assert.Equal(t, 1, len(valid), "リードスートを持つなら追従のみ: %v", valid)
	assert.Equal(t, domain.CardDesignSpade, p0.GetCard(valid[0]).GetDesign())

	// リードスートを持たなければ切り札のみ。
	p0.Reset()
	p0.AddCard(domain.NewCard(domain.CardDesignClover, 11, false))
	p0.AddCard(domain.NewCard(domain.CardDesignHeart, 11, false)) // 切り札
	valid = g.GetValidPlayIndices(0)
	assert.Equal(t, 1, len(valid), "リードスート無し・切り札ありなら切り札のみ")
	assert.Equal(t, domain.CardDesignHeart, p0.GetCard(valid[0]).GetDesign())

	// どちらも持たなければ自由。
	p0.Reset()
	p0.AddCard(domain.NewCard(domain.CardDesignClover, 11, false))
	p0.AddCard(domain.NewCard(domain.CardDesignDiamond, 12, false))
	assert.Equal(t, 2, len(g.GetValidPlayIndices(0)), "どちらも無ければ自由")
}

func TestBauernschnapsen_PhaseSwitch_Phase2MustFollow(t *testing.T) {
	// Drive real play until the stock + trump card are exhausted (endgame),
	// then verify must-follow restricts a player who can follow the led suit.
	g := newBauernschnapsen()
	g.Reset()
	guard := 0
	for !g.IsEndgame() && !g.GetGameEndFlag() && guard < 2000 {
		guard++
		switch g.GetPhase() {
		case domain.BauernschnapsenPhasePlay:
			if g.IsHumanTurn() {
				idxs := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, idxs)
				require.NoError(t, g.PlayerPlay(idxs[0]))
			} else {
				g.CpuPlay()
			}
		case domain.BauernschnapsenPhaseTrickEnd:
			g.ResolveTrick()
			g.NextTrick()
		default:
			guard = 2000
		}
	}
	require.True(t, g.IsEndgame(), "should reach endgame (stock drained)")

	// In endgame, construct a lead and assert a player holding the led suit
	// is restricted to following it.
	g.SetTrumpSuit(domain.CardDesignHeart)
	p := g.GetPlayer(g.GetCurrentPlayerIdx())
	p.Reset()
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))   // led suit
	p.AddCard(domain.NewCard(domain.CardDesignClover, 10, false)) // off suit
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: (g.GetCurrentPlayerIdx() + 3) % 4, Card: domain.NewCard(domain.CardDesignSpade, 7, false)},
	})
	g.SetPhase(domain.BauernschnapsenPhasePlay)
	valid := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
	assert.Equal(t, []int{0}, valid, "phase 2 forces following the led suit")
}

func TestBauernschnapsen_FullRoundFlow(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	guard := 0
	for !g.GetGameEndFlag() && guard < 2000 {
		guard++
		switch g.GetPhase() {
		case domain.BauernschnapsenPhaseContract:
			// **契約フェーズを回す。** クローン元 (ガイゲル) には無いので、
			// 元のループはここで止まって無限に回っていた。
			if g.IsHumanContractTurn() {
				require.NoError(t, g.DeclareContract(0,
					domain.BauernschnapsenContractRufer, domain.CardDesignSpade))
			} else {
				g.CpuDeclareContract()
			}
		case domain.BauernschnapsenPhasePlay:
			if g.IsHumanTurn() {
				idxs := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
				require.NotEmpty(t, idxs)
				require.NoError(t, g.PlayerPlay(idxs[0]))
			} else {
				g.CpuPlay()
			}
		case domain.BauernschnapsenPhaseTrickEnd:
			g.ResolveTrick()
			g.NextTrick()
		case domain.BauernschnapsenPhaseRoundEnd:
			g.ScoreRound()
			if !g.GetGameEndFlag() {
				g.NextRound()
			}
		case domain.BauernschnapsenPhaseGameEnd:
		}
	}
	assert.True(t, guard < 2000, "flow should terminate")
	// **契約が確定していること。** 切り札が未定のままトリックに入ると
	// 比較が壊れるので、進んだこと自体では足りない。
	assert.NotEqual(t, domain.BauernschnapsenContractNone, g.GetContract(),
		"a contract must have been settled")
	assert.True(t, g.GetDeclarerIdx() >= 0 && g.GetDeclarerIdx() < g.GetPlayerCnt(),
		"declarer %d is not a seat", g.GetDeclarerIdx())
}

func TestBauernschnapsen_Getters(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	g.SetRoundNumber(3)
	assert.Equal(t, 3, g.GetRoundNumber())
	g.SetTrickNumber(2)
	assert.Equal(t, 2, g.GetTrickNumber())
	g.SetCurrentPlayerIdx(1)
	assert.Equal(t, 1, g.GetCurrentPlayerIdx())
	g.SetLeadPlayerIdx(2)
	assert.Equal(t, 2, g.GetLeadPlayerIdx())
	g.SetDealerIdx(3)
	assert.Equal(t, 3, g.GetDealerIdx())
	g.SetTeamScore(0, 55)
	assert.Equal(t, 55, g.GetTeamScore(0))
	assert.Equal(t, 0, g.GetTeamScore(99))
	g.AddRoundPointsForTest(1, 12)
	assert.Equal(t, 12, g.GetRoundPoints(1))
	assert.Equal(t, 0, g.GetRoundPoints(99))
	assert.Equal(t, 0, g.GetRoundMarriagePoints(99))
	assert.Nil(t, g.GetPlayer(99))
	assert.Equal(t, 4, g.GetPlayerCnt())
	g.SetPhase(domain.BauernschnapsenPhaseTrickEnd)
	assert.Equal(t, domain.BauernschnapsenPhaseTrickEnd, g.GetPhase())
	assert.NotNil(t, g.GetConfig())
	g.SetConfig(domain.DefaultBauernschnapsenConfig())
	assert.Equal(t, 2, g.CardPointsPublic(domain.NewCard(domain.CardDesignSpade, 11, false)))
	assert.True(t, g.CardRankPublic(domain.NewCard(domain.CardDesignSpade, 1, false)) > 0)
}

func TestBauernschnapsen_GameEnd(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	g.SetTeamScore(0, 200)
	g.SetPhase(domain.BauernschnapsenPhaseRoundEnd)
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, 0, g.GetWinnerTeam())
	assert.Equal(t, domain.BauernschnapsenPhaseGameEnd, g.GetPhase())
}

func TestBauernschnapsen_GetHint(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	g.SetPhase(domain.BauernschnapsenPhasePlay)
	// Lead hint when it is human's turn.
	if g.GetPlayer(0).GetCardsSize() > 0 {
		h := g.GetHint()
		assert.NotNil(t, h)
	}
	// Not human turn -> nil.
	g.SetCurrentPlayerIdx(1)
	assert.Nil(t, g.GetHint())
	// Wrong phase -> nil.
	g.SetCurrentPlayerIdx(0)
	g.SetPhase(domain.BauernschnapsenPhaseTrickEnd)
	assert.Nil(t, g.GetHint())
}

func TestBauernschnapsen_GetHint_Marriage(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	g.SetTrumpSuit(domain.CardDesignHeart)
	p0 := g.GetPlayer(0)
	p0.Reset()
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
	p0.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
	g.SetCurrentPlayerIdx(0)
	g.SetCurrentTrick(nil)
	g.SetPhase(domain.BauernschnapsenPhasePlay)
	h := g.GetHint()
	require.NotNil(t, h)
	assert.True(t, h.IsMarriage)
	assert.Equal(t, "marriage", h.Reason)
}

func TestBauernschnapsen_JSON_RoundTrip(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var g2 domain.Bauernschnapsen
	require.NoError(t, json.Unmarshal(data, &g2))
	assert.Equal(t, g.GetPhase(), g2.GetPhase())
	assert.Equal(t, g.GetTrumpSuit(), g2.GetTrumpSuit())
	assert.Equal(t, g.GetPlayerCnt(), g2.GetPlayerCnt())
}

func TestBauernschnapsen_JSON_Invalid(t *testing.T) {
	cases := []string{
		`{"ph":99}`,                      // bad phase
		`{"ph":0,"ts":9}`,                // bad trump suit
		`{"ph":0,"pl":[null,null,null]}`, // wrong player count
		`not json`,                       // malformed
		`{"ph":0,"pl":[null,null,null,null],"cp":100}`, // currentPlayerIdx out of range
		`{"ph":0,"pl":[null,null,null,null],"cp":-1}`,  // currentPlayerIdx negative
		`{"ph":0,"pl":[null,null,null,null],"di":99}`,  // dealerIdx out of range
		`{"ph":0,"pl":[null,null,null,null],"li":99}`,  // leadPlayerIdx out of range
		`{"ph":0,"pl":[null,null,null,null],"lw":99}`,  // lastTrickWinner out of range
		`{"ph":0,"pl":[null,null,null,null],"li":-2}`,  // leadPlayerIdx below -1 sentinel
		`{"ph":0,"pl":[null,null,null,null],"wt":99}`,  // winnerTeam out of range
	}
	for _, c := range cases {
		var g domain.Bauernschnapsen
		assert.Error(t, json.Unmarshal([]byte(c), &g), c)
	}

	// Out-of-range player team is rejected.
	var p domain.BauernschnapsenPlayer
	assert.Error(t, json.Unmarshal([]byte(`{"tm":99}`), &p))

	// Valid trumpSuit=0 (undecided) with 4 players is accepted.
	valid := domain.NewDefaultBauernschnapsen()
	b, _ := json.Marshal(valid)
	var g2 domain.Bauernschnapsen
	assert.NoError(t, json.Unmarshal(b, &g2))
}

func TestBauernschnapsen_PlayerJSON(t *testing.T) {
	p := domain.NewBauernschnapsenPlayer(true, 1)
	p.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
	b, err := json.Marshal(p)
	require.NoError(t, err)
	var p2 domain.BauernschnapsenPlayer
	require.NoError(t, json.Unmarshal(b, &p2))
	assert.Equal(t, 1, p2.GetTeam())
	p2.ResetRound()
	assert.Equal(t, 0, p2.GetCardsSize())
}

func TestBauernschnapsenConfig_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultBauernschnapsenConfig().Validate())
	assert.Error(t, domain.BauernschnapsenConfig{CpuDifficulty: 99, TargetScore: 101}.Validate())
	assert.Error(t, domain.BauernschnapsenConfig{CpuDifficulty: 1, TargetScore: 0}.Validate())
}

// TestBauernschnapsen_ContractSurvivesTheWire は、契約が JSON を往復しても
// 残ることを見る。
//
// **切り札は宣言で決まる。** 契約を落とすと、復元した盤は切り札も遊び方も
// 分からなくなり、Worker がリクエストごとに別のゲームを再開することになる。
func TestBauernschnapsen_ContractSurvivesTheWire(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	// **先頭はディーラーの左隣。** 席 0 がディーラーなので席 0 からではない。
	require.NoError(t, g.DeclareContract(g.GetCurrentPlayerIdx(),
		domain.BauernschnapsenContractBettel, domain.BauernschnapsenNoTrump))
	for g.GetPhase() == domain.BauernschnapsenPhaseContract {
		if g.IsHumanContractTurn() {
			require.NoError(t, g.DeclareContract(0,
				domain.BauernschnapsenContractNone, domain.BauernschnapsenNoTrump))
			continue
		}
		g.CpuDeclareContract()
	}
	require.Equal(t, domain.BauernschnapsenContractBettel, g.GetContract(),
		"Bettel は一番高い宣言なので採用される")

	blob, err := json.Marshal(g)
	require.NoError(t, err)
	var got domain.Bauernschnapsen
	require.NoError(t, json.Unmarshal(blob, &got))

	assert.Equal(t, g.GetContract(), got.GetContract(), "契約が往復で消えた")
	assert.Equal(t, g.GetDeclarerIdx(), got.GetDeclarerIdx(), "declarer が往復で消えた")
	assert.Equal(t, g.GetTrumpSuit(), got.GetTrumpSuit(), "切り札が往復で消えた")
	assert.Equal(t, domain.BauernschnapsenNoTrump, got.GetTrumpSuit(),
		"Bettel は切り札なし")
}

// TestBauernschnapsen_HighestContractWins は、一番高い宣言が採用されることを見る。
func TestBauernschnapsen_HighestContractWins(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	// 席 1 から順に宣言する (ディーラーは席 0 なので左隣の席 1 が先頭)。
	lead := g.GetCurrentPlayerIdx()
	require.NoError(t, g.DeclareContract(lead,
		domain.BauernschnapsenContractRufer, domain.CardDesignSpade))
	require.NoError(t, g.DeclareContract((lead+1)%4,
		domain.BauernschnapsenContractFarbenzwang, domain.CardDesignHeart))
	require.NoError(t, g.DeclareContract((lead+2)%4,
		domain.BauernschnapsenContractNone, domain.BauernschnapsenNoTrump))
	require.NoError(t, g.DeclareContract((lead+3)%4,
		domain.BauernschnapsenContractRufer, domain.CardDesignClover))

	assert.Equal(t, domain.BauernschnapsenContractFarbenzwang, g.GetContract())
	assert.Equal(t, (lead+1)%4, g.GetDeclarerIdx())
	assert.Equal(t, domain.CardDesignHeart, g.GetTrumpSuit(),
		"採用された宣言のスートが切り札になる")
	assert.Equal(t, domain.BauernschnapsenPhasePlay, g.GetPhase(),
		"全員宣言したらプレイへ")
}

// TestBauernschnapsen_ContractDecidesTheRound は、契約ごとに成否の条件が
// 違うことを見る。
//
// **クローン元のガイゲルはカード点をそのまま積むだけ**だった。こちらは
// 宣言した契約を達成できたかで点が動き、Bettel と Farbenzwang は
// カード点ではなく**トリック数**で決まる。
func TestBauernschnapsen_ContractDecidesTheRound(t *testing.T) {
	cases := []struct {
		name           string
		contract       domain.BauernschnapsenContract
		declarerTricks int
		declarerSeat   int
		otherTricks    int
		declarerPoints int
		wantMade       bool
	}{
		// Bettel: 1 つでも取ったら失敗。
		{"bettel with no tricks succeeds", domain.BauernschnapsenContractBettel, 0, 0, 5, 0, true},
		{"bettel with one trick fails", domain.BauernschnapsenContractBettel, 1, 1, 4, 11, false},
		// **点とトリック数を食い違わせる。** 0 点のトリックを 1 つ取った局面。
		// 点で判定する実装 (クローン元の形) はここを「達成」と読んでしまう。
		{"bettel with a worthless trick still fails", domain.BauernschnapsenContractBettel, 1, 1, 4, 0, false},
		// **チーム合計では判定しない。** パートナー (席 2) が 2 つ取っても、
		// 宣言者自身が 0 なら達成 —— チームで見る実装はここを落とす。
		{"bettel ignores the partner's tricks", domain.BauernschnapsenContractBettel, 2, 0, 3, 30, true},
		// Farbenzwang: 相手に 1 つでも渡したら失敗。
		{"farbenzwang sweeping succeeds", domain.BauernschnapsenContractFarbenzwang, 5, 5, 0, 120, true},
		{"farbenzwang conceding one fails", domain.BauernschnapsenContractFarbenzwang, 4, 4, 1, 100, false},
		// **自分のトリック数では判定しない。** 宣言側が 0 トリックでも
		// 相手が 0 なら (全部流れた局面) 達成扱いになってはいけない ——
		// 逆に自分のトリックを見る実装はここを取り違える。
		{"farbenzwang judged on the opponent, not on itself", domain.BauernschnapsenContractFarbenzwang, 0, 0, 5, 0, false},
		// Rufer: カード点の過半 (120 の半分超 = 61 以上)。
		{"rufer with 61 succeeds", domain.BauernschnapsenContractRufer, 3, 3, 2, 61, true},
		{"rufer with exactly half fails", domain.BauernschnapsenContractRufer, 2, 2, 3, 60, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := newBauernschnapsen()
			g.Reset()
			g.SetContractForTest(c.contract, 0) // 席 0 = チーム 0 が宣言
			g.SetRoundResultForTest(0, c.declarerPoints, c.declarerTricks)
			g.SetRoundResultForTest(1, 0, c.otherTricks)
			g.SetSeatTricksForTest(0, c.declarerSeat)

			assert.Equal(t, c.wantMade, g.ContractMadeForTest(0),
				"contract %d: declarer team %d tricks (seat %d) / %d pts, other %d tricks",
				c.contract, c.declarerTricks, c.declarerSeat, c.declarerPoints, c.otherTricks)
		})
	}
}

// TestBauernschnapsen_ContractRejectsANonSuit は、切り札を要る契約が
// **本物のスートしか受け取らない**ことを見る。
//
// 0 はスート番号ではない (SPADE=1 .. DIAMOND=4)。受け取ってしまうと切り札が
// 0 の盤面ができ、画面には "UNKNOWN" が出て、トリックの比較でどの札とも
// 一致しない切り札ができあがる。
func TestBauernschnapsen_ContractRejectsANonSuit(t *testing.T) {
	for _, c := range []domain.BauernschnapsenContract{
		domain.BauernschnapsenContractRufer,
		domain.BauernschnapsenContractFarbenzwang,
	} {
		for _, suit := range []int{0, -1, domain.CardDesignMax + 1} {
			g := newBauernschnapsen()
			g.Reset()
			lead := g.GetCurrentPlayerIdx()
			err := g.DeclareContract(lead, c, suit)

			assert.Error(t, err, "契約 %d / スート %d は弾くこと", c, suit)
			assert.Equal(t, domain.BauernschnapsenContractNone, g.GetContract(),
				"弾いた宣言は採用しないこと")
			assert.Equal(t, lead, g.GetCurrentPlayerIdx(),
				"弾いた宣言で手番を進めないこと")
		}
	}

	// **負のコントロール。** 本物のスートは通る —— でなければ上の Error は
	// 「何を渡しても弾く」だけを見ていることになる。
	for suit := domain.CardDesignSpade; suit <= domain.CardDesignMax; suit++ {
		g := newBauernschnapsen()
		g.Reset()
		assert.NoError(t, g.DeclareContract(g.GetCurrentPlayerIdx(),
			domain.BauernschnapsenContractRufer, suit), "スート %d", suit)
	}

	// ベテルは切り札を取らないので、スートを渡さなくても通る。
	g := newBauernschnapsen()
	g.Reset()
	assert.NoError(t, g.DeclareContract(g.GetCurrentPlayerIdx(),
		domain.BauernschnapsenContractBettel, domain.BauernschnapsenNoTrump))
}

// TestBauernschnapsen_BettelDeclarerDucksAWinnableTrick は、ベテルの宣言者が
// **勝てるのにわざと負ける**ことを、選べる盤面で見る。
//
// クローン元のガイゲルは常にトリックを取りに行くので、そのままだと CPU は
// 自分で宣言したベテルを自分で落とす。手札に「勝つ札」と「負ける札」を
// 両方置き、どちらを選ぶかで判定する。**配りに依存しない。**
func TestBauernschnapsen_BettelDeclarerDucksAWinnableTrick(t *testing.T) {
	build := func(c domain.BauernschnapsenContract) *domain.Bauernschnapsen {
		g := newBauernschnapsen()
		g.Reset()
		g.SetContractForTest(c, 0)
		// ベテルは切り札なし。通常契約はリードスート外に切り札を置いて
		// 「切り札で勝つ」経路を避け、リードスートの高低だけで比べる。
		if c == domain.BauernschnapsenContractBettel {
			g.SetTrumpSuit(domain.BauernschnapsenNoTrump)
		} else {
			g.SetTrumpSuit(domain.CardDesignDiamond)
		}
		p0 := g.GetPlayer(0)
		p0.Reset()
		p0.AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))  // A: 場に勝つ
		p0.AddCard(domain.NewCard(domain.CardDesignSpade, 11, false)) // J: 場に負ける
		g.SetCurrentPlayerIdx(0)
		g.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 10, false)}, // 10 (10点)
		})
		g.SetPhase(domain.BauernschnapsenPhasePlay)
		return g
	}

	g := build(domain.BauernschnapsenContractBettel)
	// 両方合法であること。でなければ「選んだ」ことにならない。
	require.ElementsMatch(t, []int{0, 1}, g.GetValidPlayIndices(0),
		"A と J のどちらも出せる盤面であること")

	h := g.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	assert.Equal(t, 1, *h.CardIndex, "ベテルの宣言者は負ける J を選ぶこと")
	assert.Equal(t, "duck", h.Reason)

	// **負のコントロール。** 同じ盤面でも通常契約なら勝ちに行く ——
	// でなければ「いつも弱い札を出す」だけを見ていることになる。
	g2 := build(domain.BauernschnapsenContractRufer)
	h2 := g2.GetHint()
	require.NotNil(t, h2)
	require.NotNil(t, h2.CardIndex)
	assert.Equal(t, 0, *h2.CardIndex, "通常契約なら勝つ A を選ぶこと")
	assert.NotEqual(t, "duck", h2.Reason)
}

// TestBauernschnapsen_BettelRoundCompletes はベテルのラウンドが最後まで
// 回ることを見る。宣言者がリードを持つので、進行が止まらないこと。
func TestBauernschnapsen_BettelRoundCompletes(t *testing.T) {
	g := newBauernschnapsen()
	g.Reset()
	lead := g.GetCurrentPlayerIdx()
	bettelSeat := (lead + 2) % 4
	for i := range 4 {
		seat := (lead + i) % 4
		c := domain.BauernschnapsenContractNone
		if seat == bettelSeat {
			c = domain.BauernschnapsenContractBettel
		}
		require.NoError(t, g.DeclareContract(seat, c, domain.BauernschnapsenNoTrump))
	}
	require.Equal(t, domain.BauernschnapsenContractBettel, g.GetContract())
	declarer := g.GetDeclarerIdx()
	require.Equal(t, bettelSeat, declarer)
	// **既定のリード席と宣言者を食い違わせておく。** 先頭の席に宣言させると
	// 既定の leadPlayerIdx (ディーラーの左隣 = 先頭の席) と一致してしまい、
	// 「ベテルは宣言者からリードする」を消しても気づけない。
	require.NotEqual(t, lead, declarer, "既定のリード席と宣言者がずれていること")
	// 追従必須のこのゲームでリードを他家に握られると、宣言者は出す札を選べない。
	assert.Equal(t, declarer, g.GetLeadPlayerIdx(),
		"ベテルは宣言者からリードすること")

	tricks := 0
	for step := 0; step < 200 && g.GetPhase() != domain.BauernschnapsenPhaseRoundEnd; step++ {
		switch g.GetPhase() {
		case domain.BauernschnapsenPhasePlay:
			if g.GetPlayer(g.GetCurrentPlayerIdx()).GetIsHuman() {
				h := g.GetHint()
				require.NotNil(t, h)
				require.NotNil(t, h.CardIndex)
				require.NoError(t, g.PlayerPlay(*h.CardIndex))
			} else {
				g.CpuPlay()
			}
		case domain.BauernschnapsenPhaseTrickEnd:
			g.ResolveTrick()
			g.NextTrick()
			tricks++
		default:
			t.Fatalf("進めないフェーズ %d", g.GetPhase())
		}
	}

	require.Equal(t, domain.BauernschnapsenPhaseRoundEnd, g.GetPhase())
	// **どれだけ働いたかを主張する。** 20 枚 / 4 席 = 5 トリック。
	assert.Equal(t, 5, tricks, "5 トリック回ること")
	// 成否はチーム合計ではなく宣言者ひとりの席で決まる。
	assert.Equal(t, g.GetSeatTricks(declarer) == 0,
		g.ContractMadeForTest(g.GetPlayer(declarer).GetTeam()),
		"ベテルの成否は宣言者ひとりのトリック数で決まること")
}

// TestBauernschnapsen_NoMarriageUnderBettel は、ベテルではマリッジを
// 宣言できないことを、読み口と書き口の両方で見る。
//
// ベテルは切り札を持たず 1 トリックも取らない契約なので、キング+クイーンの
// 宣言は意味を持たない。クローン元のガイゲルには契約が無く、常に宣言できた。
func TestBauernschnapsen_NoMarriageUnderBettel(t *testing.T) {
	build := func(c domain.BauernschnapsenContract) *domain.Bauernschnapsen {
		g := newBauernschnapsen()
		g.Reset()
		g.SetContractForTest(c, 0)
		g.SetTrumpSuit(domain.CardDesignDiamond)
		p0 := g.GetPlayer(0)
		p0.Reset()
		p0.AddCard(domain.NewCard(domain.CardDesignSpade, 13, false)) // K
		p0.AddCard(domain.NewCard(domain.CardDesignSpade, 12, false)) // Q
		g.SetCurrentPlayerIdx(0)
		g.SetCurrentTrick(nil)
		g.SetPhase(domain.BauernschnapsenPhasePlay)
		return g
	}

	// **負のコントロールが先。** 通常契約では宣言できる盤面であること ——
	// でなければ下の「できない」は盤面の作り方の失敗と区別がつかない。
	g := build(domain.BauernschnapsenContractRufer)
	require.NotEmpty(t, g.GetMarriageIndices(0), "通常契約では宣言できる盤面であること")
	require.NoError(t, g.PlayerDeclareMarriage(0))

	b := build(domain.BauernschnapsenContractBettel)
	assert.Empty(t, b.GetMarriageIndices(0), "ベテルでは宣言候補を出さないこと")
	assert.Error(t, b.PlayerDeclareMarriage(0), "ベテルでは宣言を受け付けないこと")
}
