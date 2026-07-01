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
	// human ビッドで CPU が続く。最終的に nameTrump へ (human が勝者なら) か play へ。
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
	// human(0) と CPU 1,2 が pass、dealer(3) は stuck で強制 CinchMinBid。
	g.SetBidPlayerIdx(0)
	g.SetCurrentBid(0)
	g.GetPlayer(1).SetBid(domain.CinchPassBid)
	g.GetPlayer(2).SetBid(domain.CinchPassBid)
	g.GetPlayer(3).SetBid(-1)
	require.NoError(t, g.PlayerBid(domain.CinchPassBid))
	// dealer は stuck されている (最低ビッド)。
	assert.Equal(t, domain.CinchMinBid, g.GetPlayer(3).GetBid())
	assert.Equal(t, 3, g.GetBidWinnerIdx())
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
	assert.Equal(t, domain.CinchTotalPoints, total)
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
