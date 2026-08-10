//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newTestLoo は 4 人 (human=0) の Loo を Reset 済みで返す。
func newTestLoo(t *testing.T, diff domain.LooCpuDifficulty) *domain.Loo {
	t.Helper()
	players := make([]*domain.LooPlayer, domain.LooPlayerCnt)
	players[0] = domain.NewLooPlayer(true)
	for i := 1; i < domain.LooPlayerCnt; i++ {
		players[i] = domain.NewLooPlayer(false)
	}
	cfg := domain.DefaultLooConfig()
	cfg.CpuDifficulty = diff
	g := domain.NewLoo(domain.NewTrumpCards(0), players, cfg)
	g.Reset()
	return g
}

// setLooHand はプレイヤー idx の手札を指定カードで上書きする。
func setLooHand(g *domain.Loo, idx int, cards ...*domain.Card) {
	p := g.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func looCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

// looDriveFullDeal は 1 ディールを最後まで駆動する。CPU 手番は CpuPlay/CpuDecide、
// 人間手番は自動的に decide=play / 先頭の合法札で進める。ScoreRound まで到達させる。
func looDriveFullDeal(t *testing.T, g *domain.Loo) {
	t.Helper()
	for step := 0; step < 2000; step++ {
		switch g.GetPhase() {
		case domain.LooPhaseDecide:
			if g.IsHumanTurn() {
				require.NoError(t, g.PlayerDecide(true))
			} else {
				g.CpuDecide()
			}
		case domain.LooPhasePlay:
			if g.IsHumanTurn() {
				idx := g.GetPlayableIndices(g.GetCurrentTurn())
				require.NotEmpty(t, idx)
				require.NoError(t, g.PlayerPlay(idx[0]))
			} else {
				g.CpuPlay()
			}
		case domain.LooPhaseTrickEnd:
			g.ResolveTrick()
			if g.GetPhase() == domain.LooPhaseTrickEnd {
				g.NextTrick()
			}
		case domain.LooPhaseRoundEnd:
			g.ScoreRound()
			return
		}
	}
	t.Fatal("looDriveFullDeal did not reach RoundEnd")
}

func TestLoo_NewDefaultLoo(t *testing.T) {
	g := domain.NewDefaultLoo()
	assert.Equal(t, domain.LooPlayerCnt, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	for i := 1; i < domain.LooPlayerCnt; i++ {
		assert.False(t, g.GetPlayer(i).GetIsHuman())
	}
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
}

func TestLoo_Reset_DealsAndAntes(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyEasy)
	for i := 0; i < domain.LooPlayerCnt; i++ {
		assert.Equal(t, domain.LooHandSize, g.GetPlayer(i).GetCardsSize())
	}
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
	// アンティ: 各プレイヤーが Ante を支払い、ポットに入る。
	ante := g.GetConfig().Ante
	assert.Equal(t, domain.LooPlayerCnt*ante, g.GetPot())
	assert.Equal(t, g.GetPot(), g.GetPotStart())
	for i := 0; i < domain.LooPlayerCnt; i++ {
		assert.Equal(t, -ante, g.GetPlayer(i).GetChips())
	}
	// 切り札は turn-up のスート。
	assert.NotNil(t, g.GetTurnUp())
	assert.Equal(t, g.GetTurnUp().GetDesign(), g.GetTrumpSuit())
}

func TestLoo_DecidePhase_HumanNotTurn(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyEasy)
	// decide 手番を CPU (1) に設定して人間の decide を拒否。
	g.SetDecidePlayerIdx(1)
	err := g.PlayerDecide(true)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestLoo_Decide_WrongPhase(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyEasy)
	g.SetPhase(domain.LooPhaseRoundEnd)
	err := g.PlayerDecide(true)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

// TestLoo_AllPass_PotCarries は全員パスするとポットが繰り越されることを検証する。
func TestLoo_AllPass_PotCarries(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyNormal)
	for i := 0; i < domain.LooPlayerCnt; i++ {
		g.GetPlayer(i).SetPlaying(false)
	}
	g.SetPhase(domain.LooPhaseRoundEnd)
	before := g.GetPot()
	g.ScoreRound()
	assert.Equal(t, before, g.GetPot(), "全員パスならポットは変わらない")
	require.NotNil(t, g.GetLastDealDetail())
	assert.Empty(t, g.GetLastDealDetail().Looed)
}

// TestLoo_Walkover_SinglePlayerTakesPot は 1 人だけ参加した場合の総取りを検証する。
func TestLoo_Walkover_SinglePlayerTakesPot(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyNormal)
	pot := g.GetPot()
	// player 0 のみ参加。
	g.GetPlayer(0).SetPlaying(true)
	for i := 1; i < domain.LooPlayerCnt; i++ {
		g.GetPlayer(i).SetPlaying(false)
	}
	chipsBefore := g.GetPlayer(0).GetChips()
	g.SetPhase(domain.LooPhaseRoundEnd)
	g.ScoreRound()
	assert.Equal(t, chipsBefore+pot, g.GetPlayer(0).GetChips())
	assert.Equal(t, 0, g.GetPot())
}

// TestLoo_Walkover_NoFabricatedTricks は decide フローを実際に駆動して walkover に
// 到達したとき、(1) RoundEnd で即座に精算され (lastDealDetail が埋まる)、(2) 勝者の
// 表示トリック数が 0 (捏造された 5 ではない) ことを検証する。CPU の decide は乱数を
// 含むため、retry ループで walkover ケースを引く。
func TestLoo_Walkover_NoFabricatedTricks(t *testing.T) {
	for attempt := 0; attempt < 1000; attempt++ {
		g := newTestLoo(t, domain.LooCpuDifficultyNormal)
		// decide フェーズを最後まで駆動: 人間は play、CPU は自動判断。
		for g.GetPhase() == domain.LooPhaseDecide {
			if g.IsHumanTurn() {
				require.NoError(t, g.PlayerDecide(true))
			} else {
				g.CpuDecide()
			}
		}
		playing, winner := 0, -1
		for i := 0; i < domain.LooPlayerCnt; i++ {
			if g.GetPlayer(i).GetPlaying() {
				playing++
				winner = i
			}
		}
		if playing != 1 {
			continue // walkover ではない
		}
		// walkover: resolveDecide が enterRoundEnd 経由で即精算しているはず。
		require.Equal(t, domain.LooPhaseRoundEnd, g.GetPhase())
		det := g.GetLastDealDetail()
		require.NotNil(t, det, "walkover は RoundEnd 移行時に即精算されるべき")
		assert.Equal(t, 0, det.Tricks[winner], "walkover の勝者はトリック 0 (捏造 5 ではない)")
		assert.Equal(t, 0, g.GetPot(), "walkover は勝者がポットを総取り")
		return
	}
	t.Skip("1000 回の試行で walkover シナリオに到達しなかった")
}

// TestLoo_Settle_LooedPaysPenalty は参加して 0 トリックのプレイヤーが looed になることを検証する。
func TestLoo_Settle_LooedPaysPenalty(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyNormal)
	potStart := g.GetPotStart()
	// 2 人参加 (0, 1)。0 が全トリック、1 が 0 トリック。
	g.GetPlayer(0).SetPlaying(true)
	g.GetPlayer(1).SetPlaying(true)
	g.GetPlayer(2).SetPlaying(false)
	g.GetPlayer(3).SetPlaying(false)
	g.SetRoundTricks([domain.LooPlayerCnt]int{domain.LooTrickCount, 0, 0, 0})

	c0 := g.GetPlayer(0).GetChips()
	c1 := g.GetPlayer(1).GetChips()
	g.SetPhase(domain.LooPhaseRoundEnd)
	g.ScoreRound()

	share := potStart / domain.LooTrickCount
	// player 0 は 5 トリック分獲得。
	assert.Equal(t, c0+share*domain.LooTrickCount, g.GetPlayer(0).GetChips())
	// player 1 は looed でペナルティ (potStart) を支払う。
	assert.Equal(t, c1-potStart, g.GetPlayer(1).GetChips())
	det := g.GetLastDealDetail()
	require.NotNil(t, det)
	assert.Contains(t, det.Looed, 1)
	assert.NotContains(t, det.Looed, 0)
}

// TestLoo_PotConservation はポットの保存則 (out == in) をおおまかに検証する。
func TestLoo_PotConservation(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyNormal)
	// 全プレイヤーのチップ合計 + ポット は常に 0 (ゼロサム) であるべき。
	total := g.GetPot()
	for i := 0; i < domain.LooPlayerCnt; i++ {
		total += g.GetPlayer(i).GetChips()
	}
	assert.Equal(t, 0, total, "チップ + ポット はゼロサム")

	looDriveFullDeal(t, g)

	total2 := g.GetPot()
	for i := 0; i < domain.LooPlayerCnt; i++ {
		total2 += g.GetPlayer(i).GetChips()
	}
	assert.Equal(t, 0, total2, "精算後もチップ + ポット はゼロサム")
}

// TestLoo_MustFollowAndHead はマストフォロー・マストヘッドの検証を確認する。
func TestLoo_MustFollowAndHead(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyNormal)
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetPhase(domain.LooPhasePlay)
	g.GetPlayer(0).SetPlaying(true)
	g.GetPlayer(1).SetPlaying(true)

	// リード: player 1 が ハート 7 を出した状態を作る。
	g.SetCurrentTurn(0)
	g.SetLeadPlayerIdx(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: looCard(domain.CardDesignHeart, 7)},
	})
	// player 0 は ハート A と ハート 3 を持つ。マストヘッドで A を出す義務がある。
	setLooHand(g, 0, looCard(domain.CardDesignHeart, 1), looCard(domain.CardDesignHeart, 3))
	// ハート 3 (index 1) はヘッドできないので拒否される。
	err := g.PlayerPlay(1)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	// ハート A (index 0) は勝てるので合法。
	require.NoError(t, g.PlayerPlay(0))
}

// TestLoo_MustTrumpWhenVoid はリードスートがないとき切り札を出す義務を検証する。
func TestLoo_MustTrumpWhenVoid(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyNormal)
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetPhase(domain.LooPhasePlay)
	g.GetPlayer(0).SetPlaying(true)
	g.GetPlayer(1).SetPlaying(true)
	g.SetCurrentTurn(0)
	g.SetLeadPlayerIdx(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: looCard(domain.CardDesignHeart, 9)},
	})
	// player 0 は ハートなし。スペード (切り札) K とクラブ 4 を持つ。切り札を出す義務。
	setLooHand(g, 0, looCard(domain.CardDesignSpade, 13), looCard(domain.CardDesignClover, 4))
	// クラブ 4 (index 1) は切り札を持つのに切り札でないため拒否。
	err := g.PlayerPlay(1)
	assert.ErrorIs(t, err, domain.ErrInvalidPlay)
	// スペード K (index 0) は合法。
	require.NoError(t, g.PlayerPlay(0))
}

// TestLoo_DiscardWhenNoLeadNoTrump はリードも切り札もなければ任意札を捨てられることを検証する。
func TestLoo_DiscardWhenNoLeadNoTrump(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyNormal)
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetPhase(domain.LooPhasePlay)
	g.GetPlayer(0).SetPlaying(true)
	g.GetPlayer(1).SetPlaying(true)
	g.SetCurrentTurn(0)
	g.SetLeadPlayerIdx(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: looCard(domain.CardDesignHeart, 9)},
	})
	setLooHand(g, 0, looCard(domain.CardDesignClover, 4), looCard(domain.CardDesignDiamond, 2))
	require.NoError(t, g.PlayerPlay(0))
}

// TestLoo_TrickWinner_TrumpBeatsLead は切り札がリードを負かすことを検証する。
func TestLoo_TrickWinner_TrumpBeatsLead(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyNormal)
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetPhase(domain.LooPhaseTrickEnd)
	g.GetPlayer(0).SetPlaying(true)
	g.GetPlayer(1).SetPlaying(true)
	g.SetTrickNumber(1)
	g.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: looCard(domain.CardDesignHeart, 1)}, // ハート A (リード)
		{PlayerIdx: 1, Card: looCard(domain.CardDesignSpade, 2)}, // スペード 2 (切り札)
	})
	g.ResolveTrick()
	// 切り札 (player 1) が勝つ。
	tricks := g.GetRoundTricks()
	assert.Equal(t, 1, tricks[1])
	assert.Equal(t, 0, tricks[0])
	assert.Equal(t, 1, g.GetLastTrickWinner())
}

func TestLoo_GetPlayableIndices_WrongPhase(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyEasy)
	g.SetPhase(domain.LooPhaseDecide)
	assert.Nil(t, g.GetPlayableIndices(0))
	assert.Nil(t, g.GetPlayableIndices(-1))
}

// TestLoo_FullDeal_Easy/Normal/Hard は各難易度で 1 ディールを完走させ、CPU AI を網羅する。
func TestLoo_FullDeal_Easy(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyEasy)
	looDriveFullDeal(t, g)
	assert.Equal(t, domain.LooPhaseRoundEnd, g.GetPhase())
	assert.NotNil(t, g.GetLastDealDetail())
}

func TestLoo_FullDeal_Normal(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyNormal)
	looDriveFullDeal(t, g)
	assert.Equal(t, domain.LooPhaseRoundEnd, g.GetPhase())
}

func TestLoo_FullDeal_Hard(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyHard)
	looDriveFullDeal(t, g)
	assert.Equal(t, domain.LooPhaseRoundEnd, g.GetPhase())
}

// TestLoo_MultipleDeals は複数ディールを連続で駆動し、NextRound とチップ累積を検証する。
func TestLoo_MultipleDeals(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyNormal)
	for deal := 0; deal < 5; deal++ {
		looDriveFullDeal(t, g)
		round := g.GetRoundNumber()
		g.NextRound()
		assert.Equal(t, round+1, g.GetRoundNumber())
		assert.Equal(t, domain.LooPhaseDecide, g.GetPhase())
	}
}

func TestLoo_NextRound_WrongPhase(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyEasy)
	g.SetPhase(domain.LooPhasePlay)
	round := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, round, g.GetRoundNumber())
}

// TestLoo_GetHint_Decide は decide フェーズのヒントを検証する。
func TestLoo_GetHint_Decide(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyNormal)
	g.SetDecidePlayerIdx(0) // 人間
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.NotNil(t, hint.Decision)
	// 別プレイヤー手番なら nil。
	g.SetDecidePlayerIdx(1)
	assert.Nil(t, g.GetHint())
}

// TestLoo_GetHint_Play は play フェーズのヒントを検証する。
func TestLoo_GetHint_Play(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyNormal)
	g.SetTrumpSuit(domain.CardDesignSpade)
	g.SetPhase(domain.LooPhasePlay)
	g.SetCurrentTurn(0)
	g.GetPlayer(0).SetPlaying(true)
	setLooHand(g, 0, looCard(domain.CardDesignHeart, 1), looCard(domain.CardDesignClover, 3))
	// リード状況 (空トリック)。
	g.SetCurrentTrick(nil)
	g.SetLeadPlayerIdx(0)
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.NotEmpty(t, hint.CardIndices)
	// 手番でなければ nil。
	g.SetCurrentTurn(1)
	assert.Nil(t, g.GetHint())
}

func TestLoo_GetHint_WrongPhase(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyEasy)
	g.SetPhase(domain.LooPhaseTrickEnd)
	assert.Nil(t, g.GetHint())
}

// TestLoo_JSONRoundTrip は JSON シリアライズ/デシリアライズの往復を検証する。
func TestLoo_JSONRoundTrip(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyNormal)
	looDriveFullDeal(t, g)
	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored domain.Loo
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, g.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, g.GetPhase(), restored.GetPhase())
	assert.Equal(t, g.GetPot(), restored.GetPot())
}

func TestLoo_UnmarshalJSON_Errors(t *testing.T) {
	// 不正 JSON。
	var g domain.Loo
	assert.Error(t, g.UnmarshalJSON([]byte("not json")))

	// プレイヤー数不正。
	base := domain.NewDefaultLoo()
	base.Reset()
	good, err := json.Marshal(base)
	require.NoError(t, err)

	// 有効なものは復元できる。
	var g2 domain.Loo
	require.NoError(t, json.Unmarshal(good, &g2))

	// フェーズ不正。
	assert.Error(t, g.UnmarshalJSON([]byte(`{"ps":[{},{},{},{}],"ph":99,"rn":1,"tn":1,"di":3,"dp":0}`)))
	// ポット負。
	assert.Error(t, g.UnmarshalJSON([]byte(`{"ps":[{},{},{},{}],"ph":0,"rn":1,"tn":1,"di":3,"dp":0,"po":-1}`)))
	// プレイヤー数不足。
	assert.Error(t, g.UnmarshalJSON([]byte(`{"ps":[{},{}],"ph":0,"rn":1,"tn":1,"di":3,"dp":0}`)))
}

func TestLoo_Config_Validate(t *testing.T) {
	cfg := domain.DefaultLooConfig()
	assert.NoError(t, cfg.Validate())
	bad := domain.LooConfig{CpuDifficulty: 99, Ante: 3}
	assert.Error(t, bad.Validate())
	bad2 := domain.LooConfig{CpuDifficulty: domain.LooCpuDifficultyNormal, Ante: 0}
	assert.Error(t, bad2.Validate())
}

func TestLoo_Config_JSONRoundTrip(t *testing.T) {
	cfg := domain.DefaultLooConfig()
	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	var restored domain.LooConfig
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, cfg, restored)
}

func TestLoo_Player_ChipsAndReset(t *testing.T) {
	p := domain.NewLooPlayer(true)
	p.AddChips(10)
	assert.Equal(t, 10, p.GetChips())
	p.SetPlaying(true)
	assert.True(t, p.GetPlaying())
	p.AddCard(looCard(domain.CardDesignSpade, 1))
	p.ResetDeal()
	assert.Equal(t, 0, p.GetCardsSize())
	assert.False(t, p.GetPlaying())
	assert.Equal(t, 10, p.GetChips(), "ResetDeal はチップを維持する")
	p.ResetChips()
	assert.Equal(t, 0, p.GetChips())
}

func TestLoo_Setters(t *testing.T) {
	g := newTestLoo(t, domain.LooCpuDifficultyEasy)
	g.SetRoundNumber(7)
	assert.Equal(t, 7, g.GetRoundNumber())
	g.SetTrickNumber(3)
	assert.Equal(t, 3, g.GetTrickNumber())
	g.SetDealerIdx(2)
	assert.Equal(t, 2, g.GetDealerIdx())
	g.SetPot(42)
	assert.Equal(t, 42, g.GetPot())
	assert.NotNil(t, g.GetActionLog())
	assert.False(t, g.GetGameEndFlag())
	cfg := domain.DefaultLooConfig()
	cfg.Ante = 5
	g.SetConfig(cfg)
	assert.Equal(t, 5, g.GetConfig().Ante)
}

// **端数はポットに残る。**表示と精算で割り方が違うと、案内より少ない額しか入らない。
func TestLoo_PerTrickShareAndMaxWin(t *testing.T) {
	// 5 で割り切れる。
	assert.Equal(t, 8, domain.LooPerTrickShare(40))
	assert.Equal(t, 40, domain.LooMaxWin(40))

	// 割り切れない。37 → 1 トリック 7、全部取っても 35。
	assert.Equal(t, 7, domain.LooPerTrickShare(37))
	assert.Equal(t, 35, domain.LooMaxWin(37))
	assert.Less(t, domain.LooMaxWin(37), 37, "the remainder stays in the pot")

	// 0 と負。
	assert.Equal(t, 0, domain.LooPerTrickShare(0))
	assert.Equal(t, 0, domain.LooMaxWin(0))
	assert.Equal(t, 0, domain.LooPerTrickShare(-5))
	assert.Equal(t, 0, domain.LooMaxWin(-5))
}
