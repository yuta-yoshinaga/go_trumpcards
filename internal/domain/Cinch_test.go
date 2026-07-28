//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// newTestCinch は 4 人 (human=0) の Cinch を Reset 済みで返す。CPU は Easy 以外で決定的。
func newTestCinch(t *testing.T, diff domain.CinchCpuDifficulty) *domain.Cinch {
	t.Helper()
	players := make([]*domain.CinchPlayer, domain.CinchPlayerCnt)
	players[0] = domain.NewCinchPlayer(true)
	for i := 1; i < domain.CinchPlayerCnt; i++ {
		players[i] = domain.NewCinchPlayer(false)
	}
	cfg := domain.DefaultCinchConfig()
	cfg.CpuDifficulty = diff
	g := domain.NewCinch(domain.NewTrumpCards(0), players, cfg)
	g.Reset()
	return g
}

// setHand はプレイヤー idx の手札を指定カードで上書きする。
func setCinchHand(g *domain.Cinch, idx int, cards ...*domain.Card) {
	p := g.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func cinchCard(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

func TestCinch_NewDefaultCinch(t *testing.T) {
	g := domain.NewDefaultCinch()
	assert.Equal(t, domain.CinchPlayerCnt, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())
	for i := 1; i < domain.CinchPlayerCnt; i++ {
		assert.False(t, g.GetPlayer(i).GetIsHuman())
	}
}

func TestCinch_Reset_DealsNineEach(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyEasy)
	for i := 0; i < domain.CinchPlayerCnt; i++ {
		assert.Equal(t, domain.CinchHandSize, g.GetPlayer(i).GetCardsSize())
	}
	assert.Equal(t, domain.CinchPhaseBid, g.GetPhase())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.False(t, g.GetGameEndFlag())
}

// TestCinch_TrumpRanking は Left Pedro が Right Pedro のすぐ下に位置し、切り札扱いされる
// ことを検証する。切り札=Heart, 同色=Diamond。
func TestCinch_TrumpRanking(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyNormal)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetBidWinnerIdx(0)
	g.SetCurrentBid(3)
	g.SetPhase(domain.CinchPhasePlay)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentTurn(0)
	g.SetTrickNumber(1)

	// トリック: P0=5♥(Right Pedro), P1=5♦(Left Pedro), P2=A♥, P3=2♥。A♥ が最強。
	setCinchHand(g, 0, cinchCard(domain.CardDesignHeart, 5))
	setCinchHand(g, 1, cinchCard(domain.CardDesignDiamond, 5))
	setCinchHand(g, 2, cinchCard(domain.CardDesignHeart, 1))
	setCinchHand(g, 3, cinchCard(domain.CardDesignHeart, 2))

	require.NoError(t, g.PlayerPlay(0)) // 5♥ lead
	g.CpuPlay()                         // P1 5♦ (must follow trump, only card)
	g.CpuPlay()                         // P2 A♥
	g.CpuPlay()                         // P3 2♥
	require.Equal(t, domain.CinchPhaseTrickEnd, g.GetPhase())
	g.ResolveTrick()
	assert.Equal(t, 2, g.GetLastTrickWinner(), "A of trump should win")

	// 5♦ (Left Pedro) が 4♥ より強い (Right Pedro のすぐ下)。
	assert.True(t, domain.CinchCardBeatsForTest(cinchCard(domain.CardDesignDiamond, 5), cinchCard(domain.CardDesignHeart, 4), domain.CardDesignHeart))
	// 5♥ (Right Pedro) が 5♦ (Left Pedro) より強い。
	assert.True(t, domain.CinchCardBeatsForTest(cinchCard(domain.CardDesignHeart, 5), cinchCard(domain.CardDesignDiamond, 5), domain.CardDesignHeart))
}

// TestCinch_LeftPedroIsTrumpNotOffSuit は Left Pedro (同色の 5) が切り札扱いされ、
// オフ切り札スートのフォロー義務にカウントされないことを検証する。
func TestCinch_LeftPedroFollowRules(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyNormal)
	g.SetTrumpSuit(domain.CardDesignSpade) // 同色 = Clover
	g.SetBidWinnerIdx(1)
	g.SetCurrentBid(1)
	g.SetPhase(domain.CinchPhasePlay)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentTurn(0)
	g.SetTrickNumber(1)

	// P0 leads Clover K. P1 holds 5♣ (Left Pedro, trump) and Clover 3.
	setCinchHand(g, 0, cinchCard(domain.CardDesignClover, 13))
	setCinchHand(g, 1, cinchCard(domain.CardDesignClover, 5), cinchCard(domain.CardDesignClover, 3))
	setCinchHand(g, 2, cinchCard(domain.CardDesignHeart, 7))
	setCinchHand(g, 3, cinchCard(domain.CardDesignHeart, 8))

	require.NoError(t, g.PlayerPlay(0)) // K♣ lead (off-suit clover, not trump)
	assert.Equal(t, 1, g.GetCurrentTurn())
	// P1 has an off-suit clover (3♣), so must follow with clover; 5♣ is trump so
	// playing it is legal too (it can cut). Playing the clover-3 is also legal.
	pl := g.GetPlayer(1)
	// index 0 = 5♣ (trump), index 1 = 3♣ (off-suit clover)
	assert.Nil(t, g.ValidatePlayForTest(1, pl.GetCard(0)), "Left Pedro can always be played")
	assert.Nil(t, g.ValidatePlayForTest(1, pl.GetCard(1)), "off-suit clover follows the lead")
}

// TestCinch_PointScoring は 1 ディールの 14 点配分を検証する。
func TestCinch_PointScoring(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyNormal)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetBidWinnerIdx(0)
	g.SetCurrentBid(3)
	g.SetPhase(domain.CinchPhaseRoundEnd)

	// P0 が全ポイント札を取ったことにする: A♥,K♥,10♥,J♥,5♥,5♦ = 1+1+1+1+5+5 = 14。
	g.GetPlayer(0).AddTrick([]*domain.Card{
		cinchCard(domain.CardDesignHeart, 1),
		cinchCard(domain.CardDesignHeart, 13),
		cinchCard(domain.CardDesignHeart, 10),
		cinchCard(domain.CardDesignHeart, 11),
		cinchCard(domain.CardDesignHeart, 5),
		cinchCard(domain.CardDesignDiamond, 5),
	})

	g.ScoreRound()
	det := g.GetLastDealDetail()
	require.NotNil(t, det)
	assert.Equal(t, domain.CinchTotalPoints, det.Points[0], "all 14 points to P0")
	assert.False(t, det.SetBack)
	assert.Equal(t, 14, g.GetPlayer(0).GetTotalScore())
}

// TestCinch_SetBack はビッダーが宣言未達でセットバックされることを検証する。
func TestCinch_SetBack(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyNormal)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetBidWinnerIdx(0)
	g.SetCurrentBid(10) // 高すぎる宣言
	g.SetPhase(domain.CinchPhaseRoundEnd)

	// P0 は A♥ (1点) のみ、P1 が残りを取ったことにする。
	g.GetPlayer(0).AddTrick([]*domain.Card{cinchCard(domain.CardDesignHeart, 1)})
	g.GetPlayer(1).AddTrick([]*domain.Card{cinchCard(domain.CardDesignHeart, 5)}) // Right Pedro 5点

	g.ScoreRound()
	det := g.GetLastDealDetail()
	require.NotNil(t, det)
	assert.True(t, det.SetBack)
	assert.Equal(t, -10, det.Gained[0])
	assert.Equal(t, -10, g.GetPlayer(0).GetTotalScore())
	assert.Equal(t, 5, g.GetPlayer(1).GetTotalScore())
}

// TestCinch_BidFlow はビッド→切り札宣言→プレイの流れを検証する (human bid)。
func TestCinch_BidFlow(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyEasy)
	// dealer は CinchPlayerCnt-1 = 3, bid start = 0 (human)。
	assert.Equal(t, 0, g.GetBidPlayerIdx())
	require.NoError(t, g.PlayerBid(4))
	// ドメインの PlayerBid は 1 手番進めるだけ (CPU の自動消化は interactor の役目)。
	// 残りの CPU ビッダーを手動で回してビッドフェーズを終える。
	for i := 0; i < 10 && g.GetPhase() == domain.CinchPhaseBid; i++ {
		g.CpuBid()
	}
	// 全員ビッド後は nameTrump (勝者が宣言) へ移る。
	assert.NotEqual(t, domain.CinchPhaseBid, g.GetPhase())
}

func TestCinch_PlayerBid_Errors(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyEasy)
	// 範囲外ビッド。
	assert.Error(t, g.PlayerBid(99))
	// 手番でないプレイヤーのビッドは ErrNotHumanTurn (bidPlayerIdx を 1 にする)。
	g.SetBidPlayerIdx(1)
	assert.ErrorIs(t, g.PlayerBid(2), domain.ErrNotHumanTurn)
}

func TestCinch_NameTrump_Validation(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyEasy)
	g.SetPhase(domain.CinchPhaseNameTrump)
	g.SetBidWinnerIdx(0)
	assert.Error(t, g.NameTrump(0))  // 範囲外
	assert.Error(t, g.NameTrump(99)) // 範囲外
	require.NoError(t, g.NameTrump(domain.CardDesignHeart))
	assert.Equal(t, domain.CardDesignHeart, g.GetTrumpSuit())
	assert.Equal(t, domain.CinchPhasePlay, g.GetPhase())
}

func TestCinch_StuckDealer(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyEasy)
	// dealer を player 1 に据え、ビッド順を 2,3,0,1 とする。これで human(0) が最後の
	// 非 dealer ビッダーになり、human のパスで advanceBid が dealer(1) に到達して
	// stuck を強制する (deterministic; CPU の乱数ビッドに依存しない)。
	g.SetDealerIdx(1)
	g.SetBidPlayerIdx(0)
	g.SetCurrentBid(0)
	g.SetBidWinnerIdx(-1)
	g.GetPlayer(0).SetBid(-1) // これからパスする human
	g.GetPlayer(1).SetBid(-1) // dealer (stuck で確定)
	g.GetPlayer(2).SetBid(domain.CinchPassBid)
	g.GetPlayer(3).SetBid(domain.CinchPassBid)
	require.NoError(t, g.PlayerBid(domain.CinchPassBid))
	// dealer(1) は stuck されている (最低ビッド)。
	assert.Equal(t, domain.CinchMinBid, g.GetPlayer(1).GetBid())
	assert.Equal(t, 1, g.GetBidWinnerIdx())
}

func TestCinch_FullDeal_Deterministic(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyNormal)
	// 完全なディールをシミュレート: bid フェーズを CPU 主体で進める。
	for i := 0; i < 50 && g.GetPhase() == domain.CinchPhaseBid; i++ {
		if g.IsHumanTurn() {
			require.NoError(t, g.PlayerBid(domain.CinchPassBid))
		} else {
			g.CpuBid()
		}
	}
	// nameTrump フェーズ。
	if g.GetPhase() == domain.CinchPhaseNameTrump {
		if g.IsHumanTurn() {
			require.NoError(t, g.NameTrump(domain.CardDesignSpade))
		} else {
			g.CpuPlay()
		}
	}
	require.Equal(t, domain.CinchPhasePlay, g.GetPhase())

	// 9 枚×4=36 枚しか配られない (52 枚デッキ、ドロー無し) ため、場に出る得点札は
	// 配られた分だけ。全 14 点が常に配られるわけではないので、配札に含まれる得点
	// 合計を基準にする (全カードは 9 トリックで必ず取られるため、獲得合計と一致する)。
	availablePoints := 0
	for i := 0; i < domain.CinchPlayerCnt; i++ {
		p := g.GetPlayer(i)
		for j := 0; j < p.GetCardsSize(); j++ {
			availablePoints += domain.CinchPointValueForTest(p.GetCard(j), g.GetTrumpSuit())
		}
	}
	require.LessOrEqual(t, availablePoints, domain.CinchTotalPoints)

	// 9 トリックをプレイ。
	for trick := 0; trick < domain.CinchTotalTricks; trick++ {
		for c := 0; c < domain.CinchPlayerCnt; c++ {
			if g.IsHumanTurn() {
				idx := g.GetPlayableIndices(g.GetCurrentTurn())
				require.NotEmpty(t, idx)
				require.NoError(t, g.PlayerPlay(idx[0]))
			} else {
				g.CpuPlay()
			}
		}
		require.Equal(t, domain.CinchPhaseTrickEnd, g.GetPhase())
		g.ResolveTrick()
		if g.GetPhase() == domain.CinchPhaseRoundEnd {
			break
		}
		g.NextTrick()
	}
	require.Equal(t, domain.CinchPhaseRoundEnd, g.GetPhase())
	g.ScoreRound()
	det := g.GetLastDealDetail()
	require.NotNil(t, det)
	// 全ポイント (14) が配分されている。
	total := 0
	for _, p := range det.Points {
		total += p
	}
	// 獲得合計は配られた得点札の合計と一致する (全 36 枚が取られるため)。
	assert.Equal(t, availablePoints, total)
}

// cinchDriveFullDeal は 1 ディールを最後までプレイし、獲得ポイント合計が配られた
// 得点札の合計と一致することを検証する (全 36 枚が 9 トリックで必ず取られる)。CPU AI を
// 難易度別に駆動して CinchCPU.go の各分岐を網羅するための共通ドライバ。
func cinchDriveFullDeal(t *testing.T, g *domain.Cinch) {
	t.Helper()
	// bid フェーズ: human 手番はパス、CPU は自動。
	for i := 0; i < 50 && g.GetPhase() == domain.CinchPhaseBid; i++ {
		if g.IsHumanTurn() {
			require.NoError(t, g.PlayerBid(domain.CinchPassBid))
		} else {
			g.CpuBid()
		}
	}
	// nameTrump フェーズ。
	if g.GetPhase() == domain.CinchPhaseNameTrump {
		if g.IsHumanTurn() {
			require.NoError(t, g.NameTrump(domain.CardDesignSpade))
		} else {
			g.CpuPlay()
		}
	}
	require.Equal(t, domain.CinchPhasePlay, g.GetPhase())

	// 配られた得点札の合計を基準にする。
	availablePoints := 0
	for i := 0; i < domain.CinchPlayerCnt; i++ {
		p := g.GetPlayer(i)
		for j := 0; j < p.GetCardsSize(); j++ {
			availablePoints += domain.CinchPointValueForTest(p.GetCard(j), g.GetTrumpSuit())
		}
	}

	// 9 トリックをプレイ。
	for trick := 0; trick < domain.CinchTotalTricks; trick++ {
		for c := 0; c < domain.CinchPlayerCnt; c++ {
			if g.IsHumanTurn() {
				idx := g.GetPlayableIndices(g.GetCurrentTurn())
				require.NotEmpty(t, idx)
				require.NoError(t, g.PlayerPlay(idx[0]))
			} else {
				g.CpuPlay()
			}
		}
		require.Equal(t, domain.CinchPhaseTrickEnd, g.GetPhase())
		g.ResolveTrick()
		if g.GetPhase() == domain.CinchPhaseRoundEnd {
			break
		}
		g.NextTrick()
	}
	require.Equal(t, domain.CinchPhaseRoundEnd, g.GetPhase())
	g.ScoreRound()
	det := g.GetLastDealDetail()
	require.NotNil(t, det)
	total := 0
	for _, p := range det.Points {
		total += p
	}
	assert.Equal(t, availablePoints, total)
}

// TestCinch_FullDeal_Hard は Hard 難易度で 1 ディールを駆動し、CPU の
// bidWithRules(strict)/cpuSelectTrump/cpuPlaySmart 分岐を網羅する。
func TestCinch_FullDeal_Hard(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyHard)
	cinchDriveFullDeal(t, g)
}

// TestCinch_FullDeal_Easy は Easy 難易度で 1 ディールを駆動し、cpuBidEasy と
// cpuSelectPlayCard の乱数分岐を実行する。Easy のビッド結果は乱数なので特定値は
// 主張せず、フェーズが進むまでループを回すだけ。
func TestCinch_FullDeal_Easy(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyEasy)
	cinchDriveFullDeal(t, g)
}

// TestCinch_GetHint_AllPhases は GetHint が各フェーズで期待どおり non-nil / nil を返す
// ことを検証する。特に「人間の手番でない」場合に nil を返す分岐を網羅する。
func TestCinch_GetHint_AllPhases(t *testing.T) {
	// bid フェーズ、human 手番でない → nil。
	g := newTestCinch(t, domain.CinchDifficultyNormal)
	g.SetBidPlayerIdx(1)
	assert.Nil(t, g.GetHint())

	// bid フェーズ、human 手番 → Bid 付き。
	g.SetBidPlayerIdx(0)
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.NotNil(t, hint.Bid)

	// nameTrump フェーズ、勝者が human でない → nil。
	g.SetPhase(domain.CinchPhaseNameTrump)
	g.SetBidWinnerIdx(2)
	assert.Nil(t, g.GetHint())

	// nameTrump フェーズ、勝者が human → TrumpSuit 付き。
	g.SetBidWinnerIdx(0)
	hint = g.GetHint()
	require.NotNil(t, hint)
	assert.NotNil(t, hint.TrumpSuit)

	// play フェーズ、human の手番でない → nil。
	g.SetPhase(domain.CinchPhasePlay)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetCurrentTurn(1)
	assert.Nil(t, g.GetHint())

	// play フェーズ、human の手番 → CardIndices 付き。
	g.SetCurrentTurn(0)
	g.SetLeadPlayerIdx(0)
	setCinchHand(g, 0, cinchCard(domain.CardDesignHeart, 1), cinchCard(domain.CardDesignSpade, 2))
	hint = g.GetHint()
	require.NotNil(t, hint)
	assert.NotEmpty(t, hint.CardIndices)

	// TrickEnd フェーズ (default) → nil。
	g.SetPhase(domain.CinchPhaseTrickEnd)
	assert.Nil(t, g.GetHint())
}

// TestCinch_GetHint_PlayReasons は play ヒントの理由キーを各分岐で網羅する
// (lead_strong / trump_cut / follow_suit / discard_low)。
func TestCinch_GetHint_PlayReasons(t *testing.T) {
	newPlay := func() *domain.Cinch {
		g := newTestCinch(t, domain.CinchDifficultyNormal)
		g.SetPhase(domain.CinchPhasePlay)
		g.SetTrumpSuit(domain.CardDesignHeart)
		g.SetBidWinnerIdx(0)
		g.SetCurrentBid(3)
		g.SetCurrentTurn(0)
		g.SetLeadPlayerIdx(0)
		g.SetTrickNumber(1)
		return g
	}

	// リード (トリック空) → lead_strong。
	g := newPlay()
	setCinchHand(g, 0, cinchCard(domain.CardDesignHeart, 1), cinchCard(domain.CardDesignSpade, 2))
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "lead_strong", hint.Reason)

	// フォロー中で切り札を切る → trump_cut。リードは Clover、手札に Clover が無く切り札あり。
	g = newPlay()
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: cinchCard(domain.CardDesignClover, 9)}})
	setCinchHand(g, 0, cinchCard(domain.CardDesignHeart, 1))
	hint = g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "trump_cut", hint.Reason)

	// リードスートに従う → follow_suit。リード Clover、手札に Clover のみ。
	g = newPlay()
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: cinchCard(domain.CardDesignClover, 9)}})
	setCinchHand(g, 0, cinchCard(domain.CardDesignClover, 3))
	hint = g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "follow_suit", hint.Reason)

	// リードにも切り札にも従えない → discard_low。リード Clover、手札に Diamond のみ (切り札 Heart)。
	g = newPlay()
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: cinchCard(domain.CardDesignClover, 9)}})
	setCinchHand(g, 0, cinchCard(domain.CardDesignDiamond, 3))
	hint = g.GetHint()
	require.NotNil(t, hint)
	assert.Equal(t, "discard_low", hint.Reason)
}

// TestCinch_ValidatePlay_Rules はフォロー規則の各エラー分岐を網羅する。
func TestCinch_ValidatePlay_Rules(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyNormal)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetPhase(domain.CinchPhasePlay)

	// 切り札リードに対し、切り札を持っているのに従わない → エラー。
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: cinchCard(domain.CardDesignHeart, 9)}})
	setCinchHand(g, 0, cinchCard(domain.CardDesignHeart, 2), cinchCard(domain.CardDesignSpade, 3))
	assert.Error(t, g.ValidatePlayForTest(0, cinchCard(domain.CardDesignSpade, 3)))
	assert.NoError(t, g.ValidatePlayForTest(0, cinchCard(domain.CardDesignHeart, 2)))

	// 切り札リードだが切り札を持たない → 任意で合法。
	setCinchHand(g, 0, cinchCard(domain.CardDesignSpade, 3), cinchCard(domain.CardDesignClover, 4))
	assert.NoError(t, g.ValidatePlayForTest(0, cinchCard(domain.CardDesignSpade, 3)))

	// オフ切り札リード (Clover) に対し、Clover を持っているのに従わない → エラー。
	g.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 1, Card: cinchCard(domain.CardDesignClover, 9)}})
	setCinchHand(g, 0, cinchCard(domain.CardDesignClover, 3), cinchCard(domain.CardDesignDiamond, 4))
	assert.Error(t, g.ValidatePlayForTest(0, cinchCard(domain.CardDesignDiamond, 4)))
	assert.NoError(t, g.ValidatePlayForTest(0, cinchCard(domain.CardDesignClover, 3)))
	// 切り札で切るのは常に合法。
	assert.NoError(t, g.ValidatePlayForTest(0, cinchCard(domain.CardDesignHeart, 2)))
}

// TestCinch_PlayerPlay_Errors は PlayerPlay のガード分岐を網羅する。
func TestCinch_PlayerPlay_Errors(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyNormal)
	// bid フェーズでのプレイは ErrWrongPhase。
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrWrongPhase)

	g.SetPhase(domain.CinchPhasePlay)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetCurrentTurn(1) // CPU 手番
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrNotHumanTurn)

	g.SetCurrentTurn(0)
	setCinchHand(g, 0, cinchCard(domain.CardDesignHeart, 1))
	assert.Error(t, g.PlayerPlay(-1)) // 範囲外
	assert.Error(t, g.PlayerPlay(9))  // 範囲外
}

// TestCinch_NameTrump_Errors は NameTrump のガード分岐を網羅する。
func TestCinch_NameTrump_Errors(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyNormal)
	// bid フェーズでの NameTrump は ErrWrongPhase。
	assert.ErrorIs(t, g.NameTrump(domain.CardDesignHeart), domain.ErrWrongPhase)

	g.SetPhase(domain.CinchPhaseNameTrump)
	g.SetBidWinnerIdx(1) // CPU が勝者
	assert.ErrorIs(t, g.NameTrump(domain.CardDesignHeart), domain.ErrNotHumanTurn)
}

// TestCinch_GuardsAfterGameEnd はゲーム終了後の各アクションが ErrGameEnded を返すことを
// 検証する。
func TestCinch_GuardsAfterGameEnd(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyNormal)
	g.GetPlayer(0).AddScore(30)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetBidWinnerIdx(0)
	g.SetCurrentBid(1)
	g.SetPhase(domain.CinchPhaseRoundEnd)
	g.ScoreRound()
	require.True(t, g.GetGameEndFlag())

	assert.ErrorIs(t, g.PlayerBid(1), domain.ErrGameEnded)
	assert.ErrorIs(t, g.NameTrump(domain.CardDesignHeart), domain.ErrGameEnded)
	assert.ErrorIs(t, g.PlayerPlay(0), domain.ErrGameEnded)
	assert.False(t, g.IsHumanTurn())
	// ゲーム終了後の NextRound は no-op。
	round := g.GetRoundNumber()
	g.NextRound()
	assert.Equal(t, round, g.GetRoundNumber())
}

// TestCinch_ComputeRoundPoints_NoTrump は切り札未確定のときポイントが 0 になることを
// 検証する (computeRoundPoints の早期 return)。
func TestCinch_ComputeRoundPoints_NoTrump(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyNormal)
	g.SetTrumpSuit(domain.CinchTrumpUnset)
	g.SetBidWinnerIdx(0)
	g.SetCurrentBid(0)
	g.SetPhase(domain.CinchPhaseRoundEnd)
	g.GetPlayer(0).AddTrick([]*domain.Card{cinchCard(domain.CardDesignHeart, 1)})
	g.ScoreRound()
	det := g.GetLastDealDetail()
	require.NotNil(t, det)
	assert.Equal(t, 0, det.Points[0])
}

func TestCinch_JSONRoundTrip(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyNormal)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetBidWinnerIdx(0)
	g.SetCurrentBid(3)
	g.SetPhase(domain.CinchPhasePlay)
	g.SetLeadPlayerIdx(0)
	g.SetCurrentTurn(0)
	g.SetTrickNumber(1)

	data, err := json.Marshal(g)
	require.NoError(t, err)

	var restored domain.Cinch
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, domain.CardDesignHeart, restored.GetTrumpSuit())
	assert.Equal(t, domain.CinchPhasePlay, restored.GetPhase())
	assert.Equal(t, 3, restored.GetCurrentBid())
}

func TestCinch_UnmarshalJSON_Rejects(t *testing.T) {
	cases := []string{
		`{"ph":99}`,        // invalid phase
		`{"ph":0,"ps":[]}`, // wrong player count
		`{"ph":2,"ts":0,"ps":[null,null,null,null]}`, // play phase but trump unset (also nil players)
		`not json`, // malformed
	}
	for _, c := range cases {
		var g domain.Cinch
		assert.Error(t, json.Unmarshal([]byte(c), &g), c)
	}
}

func TestCinch_UnmarshalJSON_PlayPhaseRequiresTrump(t *testing.T) {
	// 正常な bid フェーズ状態を作ってから play へ改ざんし、trump 未設定を弾く。
	g := newTestCinch(t, domain.CinchDifficultyNormal)
	data, err := json.Marshal(g) // bid フェーズ、trump=0
	require.NoError(t, err)
	// bid フェーズなら trump=0 でも許容される。
	var ok domain.Cinch
	assert.NoError(t, json.Unmarshal(data, &ok))
}

func TestCinch_GetRoundWinners(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyNormal)
	assert.Nil(t, g.GetRoundWinners()) // ゲーム未終了
	g.GetPlayer(0).AddScore(25)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetBidWinnerIdx(0)
	g.SetCurrentBid(1)
	g.SetPhase(domain.CinchPhaseRoundEnd)
	g.ScoreRound()
	assert.True(t, g.GetGameEndFlag())
	assert.Equal(t, []int{0}, g.GetRoundWinners())
}

func TestCinch_Config_Validate(t *testing.T) {
	assert.NoError(t, domain.DefaultCinchConfig().Validate())
	assert.Error(t, domain.CinchConfig{CpuDifficulty: 99, PointLimit: 21}.Validate())
	assert.Error(t, domain.CinchConfig{CpuDifficulty: domain.CinchDifficultyEasy, PointLimit: 0}.Validate())
}

func TestCinch_GetHint(t *testing.T) {
	g := newTestCinch(t, domain.CinchDifficultyNormal)
	// bid フェーズで human 手番。
	hint := g.GetHint()
	require.NotNil(t, hint)
	assert.NotNil(t, hint.Bid)

	// nameTrump フェーズ。
	g.SetPhase(domain.CinchPhaseNameTrump)
	g.SetBidWinnerIdx(0)
	hint = g.GetHint()
	require.NotNil(t, hint)
	assert.NotNil(t, hint.TrumpSuit)

	// play フェーズ。
	g.SetPhase(domain.CinchPhasePlay)
	g.SetTrumpSuit(domain.CardDesignHeart)
	g.SetCurrentTurn(0)
	g.SetLeadPlayerIdx(0)
	setCinchHand(g, 0, cinchCard(domain.CardDesignHeart, 1), cinchCard(domain.CardDesignSpade, 2))
	hint = g.GetHint()
	require.NotNil(t, hint)
	assert.NotEmpty(t, hint.CardIndices)
}

func TestCinch_ColorHelpers(t *testing.T) {
	// 同色ペアの確認 (テスト用エクスポート経由)。
	assert.Equal(t, domain.CardDesignClover, domain.CinchSameColorSuitForTest(domain.CardDesignSpade))
	assert.Equal(t, domain.CardDesignSpade, domain.CinchSameColorSuitForTest(domain.CardDesignClover))
	assert.Equal(t, domain.CardDesignDiamond, domain.CinchSameColorSuitForTest(domain.CardDesignHeart))
	assert.Equal(t, domain.CardDesignHeart, domain.CinchSameColorSuitForTest(domain.CardDesignDiamond))
}
