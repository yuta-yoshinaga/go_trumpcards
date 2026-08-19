//go:build test

package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestReversis(t *testing.T) *Reversis {
	t.Helper()
	r := NewDefaultReversis()
	r.Reset()
	return r
}

// totalChips 全員の持ちチップ + プール。**このゲームで最も重要な不変量。**
func totalChips(r *Reversis) int {
	total := r.GetPool()
	for i := range ReversisPlayerCnt {
		total += r.GetPlayer(i).GetChips()
	}
	return total
}

// --- デッキ ---

// **10 を 4 枚抜いた 48 枚。** ピノクルの 48 枚とは構成が違う。
func TestReversis_DeckIs48WithoutTens(t *testing.T) {
	r := newTestReversis(t)

	bySuit := map[int][]int{}
	total := 0
	for i := range ReversisPlayerCnt {
		p := r.GetPlayer(i)
		for j := range p.GetCardsSize() {
			c := p.GetCard(j)
			bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], c.GetValue())
			total++
		}
	}

	assert.Equal(t, 48, total, "48枚がちょうど配りきられる")
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		require.Len(t, bySuit[suit], 12, "スート %d は 12 枚", suit)
		assert.NotContains(t, bySuit[suit], 10, "スート %d に 10 が無い", suit)
	}
}

func TestReversis_ResetDealsTwelveEach(t *testing.T) {
	r := newTestReversis(t)

	for i := range ReversisPlayerCnt {
		assert.Equal(t, ReversisHandSize, r.GetPlayer(i).GetCardsSize(), "player %d", i)
		// アンティを払った直後なので初期値より少ない。
		assert.Equal(t, ReversisStartingChips-ReversisAnte, r.GetPlayer(i).GetChips(), "player %d", i)
	}
	assert.Equal(t, ReversisAnte*ReversisPlayerCnt, r.GetPool(), "全員分のアンティがプールに乗る")
	assert.Equal(t, ReversisPhasePlay, r.GetPhase())
	assert.Equal(t, 1, r.GetRoundNumber())
}

// --- 失点の配分 ---

// **この配分は issue に書かれていない。** 歴史的な A=4/K=3/Q=2/J=1 を採用している。
func TestReversisCardPenalty(t *testing.T) {
	for _, tc := range []struct {
		value int
		want  int
		name  string
	}{
		{1, 4, "A"}, {13, 3, "K"}, {12, 2, "Q"}, {11, 1, "J"},
		{9, 0, "9"}, {2, 0, "2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ReversisCardPenalty(NewCard(CardDesignSpade, tc.value, false)))
		})
	}
	assert.Equal(t, 0, ReversisCardPenalty(nil))
}

// 1 スート 10 点、4 スートで 40 点ちょうど。
func TestReversis_TotalCardPenaltyIsForty(t *testing.T) {
	total := 0
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		for _, v := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13} {
			total += ReversisCardPenalty(NewCard(suit, v, false))
		}
	}
	assert.Equal(t, 40, total, "A4+K3+Q2+J1 = 10 点 × 4 スート")
}

// --- 印付きの 2 枚 ---

func TestReversis_MarkedCardIdentification(t *testing.T) {
	assert.True(t, ReversisIsQuinola(NewCard(CardDesignHeart, 11, false)), "♥J がキノラ")
	assert.False(t, ReversisIsQuinola(NewCard(CardDesignSpade, 11, false)), "♠J はキノラでない")
	assert.False(t, ReversisIsQuinola(NewCard(CardDesignHeart, 12, false)))
	assert.False(t, ReversisIsQuinola(nil))

	assert.True(t, ReversisIsDiamondAce(NewCard(CardDesignDiamond, 1, false)))
	assert.False(t, ReversisIsDiamondAce(NewCard(CardDesignSpade, 1, false)), "♠A は無印")
	assert.False(t, ReversisIsDiamondAce(NewCard(CardDesignDiamond, 13, false)))
	assert.False(t, ReversisIsDiamondAce(nil))
}

// **印付きの札は追加失点とプールへの支払いの両方を科す。**
func TestReversis_MarkedCardChargesPenaltyAndChips(t *testing.T) {
	r := newTestReversis(t)
	poolBefore := r.GetPool()
	chipsBefore := r.GetPlayer(0).GetChips()

	r.trickNumber = 2
	r.leadPlayerIdx = 0
	r.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignHeart, 1, false)},  // A♥ = 4点、これが勝つ
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 11, false)}, // キノラ
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 2, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 3, false)},
	}
	r.resolveTrick()

	p := r.GetPlayer(0)
	assert.True(t, p.GetTookQuinola(), "取った人に印が付く")
	// A♥(4) + ♥J(1) + キノラの追加(5) = 10
	assert.Equal(t, 10, p.GetRoundPenalty())
	assert.Equal(t, chipsBefore-ReversisMarkedStake, p.GetChips(), "プールへ 5 払う")
	assert.Equal(t, poolBefore+ReversisMarkedStake, r.GetPool())
	assert.False(t, r.GetPlayer(1).GetTookQuinola(), "出した人には付かない")
}

func TestReversis_DiamondAceCharges(t *testing.T) {
	r := newTestReversis(t)
	r.trickNumber = 2
	r.leadPlayerIdx = 1
	r.currentTrick = []*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignDiamond, 1, false)}, // ♦A、これが勝つ
		{PlayerIdx: 2, Card: NewCard(CardDesignDiamond, 2, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignDiamond, 3, false)},
		{PlayerIdx: 0, Card: NewCard(CardDesignDiamond, 4, false)},
	}
	r.resolveTrick()

	p := r.GetPlayer(1)
	assert.True(t, p.GetTookDiamondAce())
	assert.Equal(t, 4+ReversisMarkedPenalty, p.GetRoundPenalty(), "A の 4 点 + 印の 5 点")
}

// 印の無いトリックは追加負担なし。負のコントロール。
func TestReversis_UnmarkedTrickChargesNoStake(t *testing.T) {
	r := newTestReversis(t)
	poolBefore := r.GetPool()
	chipsBefore := r.GetPlayer(0).GetChips()

	r.trickNumber = 2
	r.leadPlayerIdx = 0
	r.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 9, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 3, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 4, false)},
	}
	r.resolveTrick()

	assert.Equal(t, 0, r.GetPlayer(0).GetRoundPenalty(), "絵札が無ければ 0 点")
	assert.Equal(t, chipsBefore, r.GetPlayer(0).GetChips(), "チップは動かない")
	assert.Equal(t, poolBefore, r.GetPool())
	assert.False(t, r.GetPlayer(0).GetTookQuinola())
	assert.False(t, r.GetPlayer(0).GetTookDiamondAce())
}

// --- 切り札なし・フォロー義務 ---

func TestReversis_NoTrump(t *testing.T) {
	r := newTestReversis(t)
	r.leadPlayerIdx = 0
	r.currentTrick = []*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 2, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 1, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignDiamond, 1, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignClover, 1, false)},
	}
	assert.Equal(t, 0, r.trickWinner(), "リードのスートの 2 が他スートの A 3 枚に勝つ")
}

func TestReversisRank_AceIsHighest(t *testing.T) {
	assert.Greater(t, reversisRank(NewCard(CardDesignSpade, 1, false)), reversisRank(NewCard(CardDesignSpade, 13, false)))
	assert.Greater(t, reversisRank(NewCard(CardDesignSpade, 13, false)), reversisRank(NewCard(CardDesignSpade, 2, false)))
	assert.Equal(t, 0, reversisRank(nil))
}

func TestReversis_MustFollowSuit(t *testing.T) {
	r := newTestReversis(t)
	p := r.GetPlayer(1)
	p.Reset()
	p.AddCard(NewCard(CardDesignSpade, 8, false))
	p.AddCard(NewCard(CardDesignHeart, 9, false))
	r.currentTrick = []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 13, false)}}

	assert.Equal(t, []int{0}, r.GetValidPlayIndices(1))
}

func TestReversis_VoidPlaysAnything(t *testing.T) {
	r := newTestReversis(t)
	p := r.GetPlayer(1)
	p.Reset()
	p.AddCard(NewCard(CardDesignHeart, 9, false))
	p.AddCard(NewCard(CardDesignClover, 8, false))
	r.currentTrick = []*TrickCard{{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 13, false)}}

	assert.Equal(t, []int{0, 1}, r.GetValidPlayIndices(1))
}

func TestReversis_GetValidPlayIndicesOutOfRange(t *testing.T) {
	r := newTestReversis(t)
	assert.Nil(t, r.GetValidPlayIndices(-1))
	assert.Nil(t, r.GetValidPlayIndices(ReversisPlayerCnt))
}

// --- プールと配当 ---

// **失点が最も少ないプレイヤーがプールを総取りする。**
func TestReversis_FewestPenaltyTakesThePool(t *testing.T) {
	r := newTestReversis(t)
	r.config.Rounds = 2
	pool := r.GetPool()
	chipsBefore := r.GetPlayer(2).GetChips()

	r.GetPlayer(0).SetRoundPenalty(12)
	r.GetPlayer(1).SetRoundPenalty(20)
	r.GetPlayer(2).SetRoundPenalty(3)
	r.GetPlayer(3).SetRoundPenalty(5)

	r.FinishRoundForTest()

	assert.Equal(t, chipsBefore+pool, r.GetPlayer(2).GetChips(), "最少失点がプールを取る")
	assert.Equal(t, 0, r.GetPool(), "プールは空になる")
	assert.Equal(t, ReversisPhaseRoundEnd, r.GetPhase())
}

// **同点ならプールは次ラウンドへ持ち越す。** チップが宙に消えてはいけない。
func TestReversis_TiedRoundCarriesThePoolOver(t *testing.T) {
	r := newTestReversis(t)
	r.config.Rounds = 3
	pool := r.GetPool()
	for i := range ReversisPlayerCnt {
		r.GetPlayer(i).SetRoundPenalty(10)
	}
	before := totalChips(r)

	r.FinishRoundForTest()

	assert.Equal(t, pool, r.GetPool(), "プールはそのまま残る")
	assert.Equal(t, before, totalChips(r), "チップの総量は変わらない")

	r.NextRound()
	assert.Equal(t, pool+ReversisAnte*ReversisPlayerCnt, r.GetPool(),
		"次ラウンドのアンティが持ち越し分に積み上がる")
}

// **チップは生まれも消えもしない。** プール制のゲームで最も重要な不変量。
func TestReversis_ChipsAreConserved(t *testing.T) {
	for range 30 {
		r := NewDefaultReversis()
		r.Reset()
		want := ReversisStartingChips * ReversisPlayerCnt
		require.Equal(t, want, totalChips(r), "配り直後")

		guard := 0
		for !r.GetGameEndFlag() && guard < 1000 {
			guard++
			if r.GetPhase() == ReversisPhaseRoundEnd {
				r.NextRound()
				require.Equal(t, want, totalChips(r), "ラウンド開始時")
				continue
			}
			if r.IsHumanTurn() {
				valid := r.GetValidPlayIndices(0)
				require.NotEmpty(t, valid)
				require.NoError(t, r.PlayerPlay(valid[0]))
			} else {
				r.CpuPlay()
			}
			require.Equal(t, want, totalChips(r), "1 手ごと")
		}
		require.True(t, r.GetGameEndFlag())
		require.Equal(t, want, totalChips(r), "終局時")
		// **終局時にプールが残っていてはいけない。** 総量が保たれていても、
		// 盤上に取り残されたチップは GetChips() に入らず勝敗に反映されない。
		require.Equal(t, 0, r.GetPool(), "終局時にプールは空")
	}
}

// **最終ラウンドが同点でも、プールは必ず誰かの手に渡る。**
// 持ち越し先が無いので、ここを持ち越し扱いにするとチップが盤上に取り残される
// （総量は保たれるので、保存則の assert だけでは捕まらない）。
func TestReversis_TieOnTheFinalRoundSplitsThePool(t *testing.T) {
	r := newTestReversis(t)
	r.config.Rounds = 1
	r.SetPoolForTest(20)
	before := totalChips(r)
	chipsBefore := make([]int, ReversisPlayerCnt)
	for i := range ReversisPlayerCnt {
		r.GetPlayer(i).SetRoundPenalty(10) // 全員同点
		chipsBefore[i] = r.GetPlayer(i).GetChips()
	}

	r.FinishRoundForTest()

	assert.True(t, r.GetGameEndFlag(), "最終ラウンドなので終局する")
	assert.Equal(t, 0, r.GetPool(), "プールは残らない")
	assert.Equal(t, before, totalChips(r), "チップの総量は変わらない")
	for i := range ReversisPlayerCnt {
		assert.Equal(t, chipsBefore[i]+5, r.GetPlayer(i).GetChips(), "20 を 4 人で分けて 5 ずつ (player %d)", i)
	}
}

// 割り切れない端数も 1 枚残さず配る。
func TestReversis_TieSplitHandsOutTheRemainder(t *testing.T) {
	r := newTestReversis(t)
	r.config.Rounds = 1
	r.SetPoolForTest(22) // 4 人で割ると 5 余り 2
	before := totalChips(r)
	for i := range ReversisPlayerCnt {
		r.GetPlayer(i).SetRoundPenalty(10)
	}

	r.FinishRoundForTest()

	assert.Equal(t, 0, r.GetPool(), "端数も残さない")
	assert.Equal(t, before, totalChips(r))
}

// 同点が一部だけなら、その人たちだけで分ける。
func TestReversis_TieSplitOnlyAmongTheTiedLeaders(t *testing.T) {
	r := newTestReversis(t)
	r.config.Rounds = 1
	r.SetPoolForTest(20)
	r.GetPlayer(0).SetRoundPenalty(3)
	r.GetPlayer(1).SetRoundPenalty(3)
	r.GetPlayer(2).SetRoundPenalty(9)
	r.GetPlayer(3).SetRoundPenalty(12)
	before := []int{
		r.GetPlayer(0).GetChips(), r.GetPlayer(1).GetChips(),
		r.GetPlayer(2).GetChips(), r.GetPlayer(3).GetChips(),
	}

	r.FinishRoundForTest()

	assert.Equal(t, 0, r.GetPool())
	assert.Equal(t, before[0]+10, r.GetPlayer(0).GetChips(), "同点の 2 人で 20 を折半")
	assert.Equal(t, before[1]+10, r.GetPlayer(1).GetChips())
	assert.Equal(t, before[2], r.GetPlayer(2).GetChips(), "失点が多い人は貰えない")
	assert.Equal(t, before[3], r.GetPlayer(3).GetChips())
}

// 1 ラウンドで動く失点は 40（カード）＋ 5＋5（印 2 枚）= 50 ちょうど。
func TestReversis_ExactlyFiftyPenaltyPerRound(t *testing.T) {
	for range 30 {
		r := NewDefaultReversis()
		r.config.Rounds = 1
		r.Reset()
		guard := 0
		for !r.GetGameEndFlag() && guard < 200 {
			guard++
			if r.IsHumanTurn() {
				valid := r.GetValidPlayIndices(0)
				require.NoError(t, r.PlayerPlay(valid[0]))
				continue
			}
			r.CpuPlay()
		}
		total := 0
		for i := range ReversisPlayerCnt {
			total += r.GetPlayer(i).GetRoundPenalty()
		}
		require.Equal(t, 50, total, "カード40 + キノラ5 + ♦A 5")
	}
}

// --- ゲーム終了 ---

// **チップが最も多いプレイヤーが勝つ。**
func TestReversis_MostChipsWins(t *testing.T) {
	r := newTestReversis(t)
	r.GetPlayer(0).SetChips(30)
	r.GetPlayer(1).SetChips(80)
	r.GetPlayer(2).SetChips(45)
	r.GetPlayer(3).SetChips(45)

	r.FinishGameForTest()

	assert.True(t, r.GetGameEndFlag())
	assert.Equal(t, ReversisPhaseGameEnd, r.GetPhase())
	assert.Equal(t, 1, r.GetWinnerIdx())
}

func TestReversis_TieHasNoWinner(t *testing.T) {
	r := newTestReversis(t)
	for i := range ReversisPlayerCnt {
		r.GetPlayer(i).SetChips(50)
	}
	r.FinishGameForTest()
	assert.Equal(t, -1, r.GetWinnerIdx())
}

func TestReversis_RoundEndVsGameEnd(t *testing.T) {
	t.Run("more rounds remain", func(t *testing.T) {
		r := newTestReversis(t)
		r.config.Rounds = 3
		r.roundNumber = 1
		r.FinishRoundForTest()
		assert.Equal(t, ReversisPhaseRoundEnd, r.GetPhase())
		assert.False(t, r.GetGameEndFlag())
	})
	t.Run("final round", func(t *testing.T) {
		r := newTestReversis(t)
		r.config.Rounds = 3
		r.roundNumber = 3
		r.FinishRoundForTest()
		assert.Equal(t, ReversisPhaseGameEnd, r.GetPhase())
		assert.True(t, r.GetGameEndFlag())
	})
}

func TestReversis_NextRoundRedealsAndKeepsChips(t *testing.T) {
	r := newTestReversis(t)
	r.config.Rounds = 3
	r.GetPlayer(0).SetChips(70)
	r.GetPlayer(0).SetTookQuinola(true)
	r.SetPhaseForTest(ReversisPhaseRoundEnd)
	r.SetPoolForTest(0)
	dealer := r.GetDealerIdx()

	r.NextRound()

	assert.Equal(t, 2, r.GetRoundNumber())
	assert.Equal(t, ReversisPhasePlay, r.GetPhase())
	assert.Equal(t, (dealer+1)%ReversisPlayerCnt, r.GetDealerIdx(), "ディーラーが回る")
	assert.Equal(t, 70-ReversisAnte, r.GetPlayer(0).GetChips(), "チップは持ち越し、アンティを払う")
	assert.False(t, r.GetPlayer(0).GetTookQuinola(), "印はラウンドごとに消える")
	for i := range ReversisPlayerCnt {
		assert.Equal(t, ReversisHandSize, r.GetPlayer(i).GetCardsSize())
	}
}

func TestReversis_NextRoundIgnoredOutsideRoundEnd(t *testing.T) {
	r := newTestReversis(t)
	r.NextRound()
	assert.Equal(t, 1, r.GetRoundNumber())

	r.gameEndFlag = true
	r.SetPhaseForTest(ReversisPhaseRoundEnd)
	r.NextRound()
	assert.Equal(t, 1, r.GetRoundNumber())
}

// --- プレイ ---

func TestReversis_PlayerPlayRejections(t *testing.T) {
	t.Run("not your turn", func(t *testing.T) {
		r := newTestReversis(t)
		r.SetCurrentPlayerIdxForTest(2)
		assert.Error(t, r.PlayerPlay(0))
	})
	t.Run("game over", func(t *testing.T) {
		r := newTestReversis(t)
		r.gameEndFlag = true
		assert.Error(t, r.PlayerPlay(0))
	})
	t.Run("round ended", func(t *testing.T) {
		r := newTestReversis(t)
		r.SetPhaseForTest(ReversisPhaseRoundEnd)
		r.SetCurrentPlayerIdxForTest(0)
		assert.Error(t, r.PlayerPlay(0))
	})
	t.Run("index out of range", func(t *testing.T) {
		r := newTestReversis(t)
		r.SetCurrentPlayerIdxForTest(0)
		assert.Error(t, r.PlayerPlay(99))
		assert.Error(t, r.PlayerPlay(-1))
	})
	t.Run("must follow suit", func(t *testing.T) {
		r := newTestReversis(t)
		p := r.GetPlayer(0)
		p.Reset()
		p.AddCard(NewCard(CardDesignSpade, 8, false))
		p.AddCard(NewCard(CardDesignHeart, 9, false))
		r.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}}
		r.SetCurrentPlayerIdxForTest(0)
		assert.Error(t, r.PlayerPlay(1))
		assert.NoError(t, r.PlayerPlay(0))
	})
}

func TestReversis_CpuPlayIsANoOpOnHumanTurn(t *testing.T) {
	r := newTestReversis(t)
	r.SetCurrentPlayerIdxForTest(0)
	r.CpuPlay()
	assert.Equal(t, ReversisHandSize, r.GetPlayer(0).GetCardsSize())
}

func TestReversis_CpuAlwaysPlaysLegally(t *testing.T) {
	for range 50 {
		r := NewDefaultReversis()
		r.Reset()
		guard := 0
		for !r.GetGameEndFlag() && guard < 1000 {
			guard++
			if r.GetPhase() == ReversisPhaseRoundEnd {
				r.NextRound()
				continue
			}
			if r.IsHumanTurn() {
				valid := r.GetValidPlayIndices(0)
				require.NotEmpty(t, valid)
				require.NoError(t, r.PlayerPlay(valid[0]))
				continue
			}
			idx := r.GetCurrentPlayerIdx()
			before := r.GetPlayer(idx).GetCardsSize()
			r.CpuPlay()
			require.Equal(t, before-1, r.GetPlayer(idx).GetCardsSize())
		}
		require.True(t, r.GetGameEndFlag())
	}
}

func TestReversis_GiveUp(t *testing.T) {
	r := newTestReversis(t)
	r.GiveUp()
	assert.True(t, r.GetGameEndFlag())
	assert.Equal(t, ReversisPhaseGameEnd, r.GetPhase())
	assert.Equal(t, -1, r.GetWinnerIdx())

	r.GiveUp()
	assert.True(t, r.GetGameEndFlag())
}

func TestReversis_IsHumanTurn(t *testing.T) {
	r := newTestReversis(t)
	r.SetCurrentPlayerIdxForTest(0)
	assert.True(t, r.IsHumanTurn())
	r.SetCurrentPlayerIdxForTest(2)
	assert.False(t, r.IsHumanTurn())
	r.SetCurrentPlayerIdxForTest(0)
	r.SetPhaseForTest(ReversisPhaseRoundEnd)
	assert.False(t, r.IsHumanTurn())
	r.SetPhaseForTest(ReversisPhasePlay)
	r.gameEndFlag = true
	assert.False(t, r.IsHumanTurn())
}

func TestReversis_GetPlayerOutOfRange(t *testing.T) {
	r := newTestReversis(t)
	assert.Nil(t, r.GetPlayer(-1))
	assert.Nil(t, r.GetPlayer(ReversisPlayerCnt))
}

func TestReversis_Config(t *testing.T) {
	r := newTestReversis(t)
	assert.Equal(t, ReversisRoundsDefault, r.GetConfig().Rounds)

	r.SetConfig(ReversisConfig{Rounds: 6})
	assert.Equal(t, 6, r.GetConfig().Rounds)

	assert.NoError(t, ReversisConfig{Rounds: ReversisRoundsMin}.Validate())
	assert.NoError(t, ReversisConfig{Rounds: ReversisRoundsMax}.Validate())
	assert.Error(t, ReversisConfig{Rounds: 0}.Validate())
	assert.Error(t, ReversisConfig{Rounds: ReversisRoundsMax + 1}.Validate())
}

// --- ヒント ---

func TestReversis_GetHint_NilWhenNotHumanTurn(t *testing.T) {
	r := newTestReversis(t)
	r.SetCurrentPlayerIdxForTest(2)
	assert.Nil(t, r.GetHint())

	r.SetCurrentPlayerIdxForTest(0)
	r.gameEndFlag = true
	assert.Nil(t, r.GetHint())
}

func TestReversis_GetHint_SuggestsALegalCard(t *testing.T) {
	r := newTestReversis(t)
	r.SetCurrentPlayerIdxForTest(0)

	h := r.GetHint()
	if assert.NotNil(t, h) && assert.NotNil(t, h.CardIndex) {
		assert.Contains(t, r.GetValidPlayIndices(0), *h.CardIndex)
		assert.NotEmpty(t, h.Reason)
	}
}

// 4 つの理由キーがそれぞれ出る条件を全部踏む。
func TestReversis_GetHint_Reasons(t *testing.T) {
	t.Run("lead", func(t *testing.T) {
		r := newTestReversis(t)
		r.SetCurrentPlayerIdxForTest(0)
		r.currentTrick = nil
		assert.Equal(t, "reversisLeadSafe", r.GetHint().Reason)
	})
	t.Run("a marked card is on the table", func(t *testing.T) {
		r := newTestReversis(t)
		r.SetCurrentPlayerIdxForTest(0)
		r.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 11, false)}}
		assert.Equal(t, "reversisAvoidMarked", r.GetHint().Reason)
	})
	t.Run("plain penalty cards on the table", func(t *testing.T) {
		r := newTestReversis(t)
		r.SetCurrentPlayerIdxForTest(0)
		r.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}}
		assert.Equal(t, "reversisAvoidPoints", r.GetHint().Reason)
	})
	t.Run("nothing worth points on the table", func(t *testing.T) {
		r := newTestReversis(t)
		r.SetCurrentPlayerIdxForTest(0)
		r.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 9, false)}}
		assert.Equal(t, "reversisDumpHigh", r.GetHint().Reason)
	})
}

// --- JSON 往復 ---

func TestReversis_JSONRoundTrip(t *testing.T) {
	r := newTestReversis(t)
	r.GetPlayer(0).SetChips(72)
	r.GetPlayer(0).SetRoundPenalty(9)
	r.GetPlayer(0).SetTookQuinola(true)
	r.GetPlayer(1).SetTookDiamondAce(true)
	r.SetPoolForTest(35)
	r.roundNumber = 2
	r.trickNumber = 5
	r.currentTrick = []*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignHeart, 11, false)}}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	restored := NewDefaultReversis()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, 72, restored.GetPlayer(0).GetChips(), "チップが往復する")
	assert.Equal(t, 9, restored.GetPlayer(0).GetRoundPenalty())
	assert.True(t, restored.GetPlayer(0).GetTookQuinola(), "印が往復する")
	assert.True(t, restored.GetPlayer(1).GetTookDiamondAce())
	assert.Equal(t, 35, restored.GetPool(), "プールが往復する")
	assert.Equal(t, 2, restored.GetRoundNumber())
	assert.Equal(t, 5, restored.GetTrickNumber())
	assert.Equal(t, r.GetConfig().Rounds, restored.GetConfig().Rounds)
	require.Len(t, restored.GetCurrentTrick(), 1)
}

func TestReversis_UnmarshalRejectsGarbage(t *testing.T) {
	assert.Error(t, json.Unmarshal([]byte("not json"), NewDefaultReversis()))
}

// 負のプールは壊れた KV。読み込みを拒否する。
func TestReversis_UnmarshalRejectsNegativePool(t *testing.T) {
	r := newTestReversis(t)
	data, err := json.Marshal(r)
	require.NoError(t, err)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))
	raw["po"] = -1
	broken, err := json.Marshal(raw)
	require.NoError(t, err)

	assert.Error(t, json.Unmarshal(broken, NewDefaultReversis()))
}

func TestReversis_ActionLog(t *testing.T) {
	r := newTestReversis(t)
	assert.NotEmpty(t, r.GetActionLog())
}

// **手札に出す点数と精算の点数が同じであること** (#5747)。TS 側も同じ
// 黄金ベクタを読むので、片方だけ変えれば必ずどちらかが落ちる。
func TestReversisCardPenalty_GoldenVectors(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(
		"..", "..", "frontend", "src", "utils", "__fixtures__", "reversisPoints.golden.json"))
	if err != nil {
		t.Fatalf("read the golden vectors: %v", err)
	}
	var golden struct {
		Cases []struct {
			Name   string `json:"name"`
			Design string `json:"design"`
			Value  int    `json:"value"`
			Points int    `json:"points"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("parse the golden vectors: %v", err)
	}
	if len(golden.Cases) == 0 {
		t.Fatal("no vectors to check")
	}
	designs := map[string]int{
		"SPADE": CardDesignSpade, "CLOVER": CardDesignClover,
		"HEART": CardDesignHeart, "DIAMOND": CardDesignDiamond,
	}
	for _, c := range golden.Cases {
		design, ok := designs[c.Design]
		if !ok {
			t.Fatalf("%s: unknown design %q", c.Name, c.Design)
		}
		// 印付きの 2 枚は基礎点 + ReversisMarkedPenalty で請求される。
		if got := ReversisTotalCardPenalty(NewCard(design, c.Value, true)); got != c.Points {
			t.Errorf("%s: got %d", c.Name, got)
		}
	}
}
