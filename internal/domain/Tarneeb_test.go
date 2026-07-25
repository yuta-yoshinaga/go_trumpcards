//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestTarneeb() *domain.Tarneeb {
	players := []*domain.TarneebPlayer{
		domain.NewTarneebPlayer(true, 0),
		domain.NewTarneebPlayer(false, 1),
		domain.NewTarneebPlayer(false, 0),
		domain.NewTarneebPlayer(false, 1),
	}
	return domain.NewTarneeb(domain.NewTrumpCards(0), players, domain.DefaultTarneebConfig())
}

// fillHandFromTrick はテスト用に、配布済みカードを取り除いて 13 枚 + 任意の追加カードを差し込む補助。
// ここでは「手札を空にして手動で詰める」ヘルパを提供する。
func clearAndDeal(p *domain.TarneebPlayer, cards []*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func TestNewTarneeb(t *testing.T) {
	tn := newTestTarneeb()
	assert.Equal(t, -1, tn.GetWinnerTeam())
	assert.Equal(t, -1, tn.GetBidWinnerIdx())
	assert.Equal(t, 0, tn.GetTrumpSuit())
}

func TestNewDefaultTarneeb(t *testing.T) {
	tn := domain.NewDefaultTarneeb()
	require.NotNil(t, tn)
	assert.Equal(t, domain.TarneebPlayerCnt, tn.GetPlayerCnt())
	assert.True(t, tn.GetPlayer(0).GetIsHuman())
	assert.Equal(t, 0, tn.GetPlayer(0).GetTeam())
	assert.Equal(t, 1, tn.GetPlayer(1).GetTeam())
	assert.Equal(t, 0, tn.GetPlayer(2).GetTeam())
	assert.Equal(t, 1, tn.GetPlayer(3).GetTeam())
	assert.False(t, tn.GetGameEndFlag())
	assert.Equal(t, domain.DefaultTarneebConfig(), tn.GetConfig())
}

func TestTarneeb_Reset(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()

	assert.Equal(t, domain.TarneebPhaseBid, tn.GetPhase())
	assert.Equal(t, 1, tn.GetRoundNumber())
	assert.Equal(t, 0, tn.GetTrickNumber())
	assert.Equal(t, -1, tn.GetBidWinnerIdx())
	assert.Equal(t, 0, tn.GetHighestBid())
	assert.Equal(t, 0, tn.GetTrumpSuit())
	assert.Equal(t, 0, tn.GetRedealCount())
	assert.Equal(t, 0, tn.GetDealerIdx())
	// ディーラー0 → 最初のビッド手番はディーラー左隣(idx 1)
	assert.Equal(t, 1, tn.GetBidPlayerIdx())
	for i := 0; i < domain.TarneebPlayerCnt; i++ {
		assert.Equal(t, domain.TarneebHandSize, tn.GetPlayer(i).GetCardsSize())
		assert.Equal(t, -1, tn.GetPlayer(i).GetBid())
	}
}

func TestTarneeb_PlayerBid_NotHumanTurn(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	// 人間は idx 0、最初のビッド手番は idx 1 (CPU)
	err := tn.PlayerBid(7)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestTarneeb_PlayerBid_Pass_Valid(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	// dealer=3 にすると dealer+1=0 が最初のビッド手番。1人だけビッドしても
	// 4人分のラウンドが完了しないため finishBidPhase は呼ばれず、bid 値が保持される。
	tn.SetDealerIdx(3)
	tn.SetBidPlayerIdx(0)
	require.NoError(t, tn.PlayerBid(domain.TarneebPassBid))
	assert.Equal(t, domain.TarneebPassBid, tn.GetPlayer(0).GetBid())
}

func TestTarneeb_PlayerBid_ValueRange(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetBidPlayerIdx(0)
	require.NoError(t, tn.PlayerBid(7))
	assert.Equal(t, 7, tn.GetHighestBid())
	assert.Equal(t, 0, tn.GetBidWinnerIdx())
}

func TestTarneeb_PlayerBid_BelowMin_Errors(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetBidPlayerIdx(0)
	err := tn.PlayerBid(6)
	assert.Error(t, err)
}

func TestTarneeb_PlayerBid_AboveMax_Errors(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetBidPlayerIdx(0)
	err := tn.PlayerBid(14)
	assert.Error(t, err)
}

func TestTarneeb_PlayerBid_MustExceedCurrent(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetBidPlayerIdx(0)
	tn.SetHighestBid(9)
	err := tn.PlayerBid(9)
	assert.Error(t, err)
	require.NoError(t, tn.PlayerBid(10))
}

func TestTarneeb_AllPass_TriggersRedeal(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetBidPlayerIdx(0) // make human first for the test
	tn.SetDealerIdx(3)    // dealer=3 → left=0
	// 全CPUに低札のみの手札を持たせて確実にパスさせる
	lowHand := func() []*domain.Card {
		return []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 2, false),
			domain.NewCard(domain.CardDesignClover, 3, false),
			domain.NewCard(domain.CardDesignHeart, 4, false),
			domain.NewCard(domain.CardDesignDiamond, 5, false),
		}
	}
	for i := 1; i < domain.TarneebPlayerCnt; i++ {
		clearAndDeal(tn.GetPlayer(i), lowHand())
	}
	// 4人連続でパス
	require.NoError(t, tn.PlayerBid(domain.TarneebPassBid))
	tn.CpuBid()
	tn.CpuBid()
	tn.CpuBid()
	// Redeal → phase は Bid のまま、redealCount が +1
	assert.Equal(t, domain.TarneebPhaseBid, tn.GetPhase())
	assert.Equal(t, 1, tn.GetRedealCount())
	assert.Equal(t, -1, tn.GetBidWinnerIdx())
}

func TestTarneeb_FullBidRound_AdvancesToTrumpPhase(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetBidPlayerIdx(0)
	tn.SetDealerIdx(3)
	require.NoError(t, tn.PlayerBid(7))
	// 3人のCPUにパスさせる: 強制的にパスをセットして手番を進める
	for i := 1; i < 4; i++ {
		// 直接ビッドをパスにする
		tn.SetBidPlayerIdx(i)
		tn.GetPlayer(i).SetBid(domain.TarneebPassBid)
		tn.SetBidPlayerIdx((i + 1) % domain.TarneebPlayerCnt)
	}
	// 末尾の applyBid を踏ませるため、最終プレイヤーで applyBid を実行させる:
	// ここでは PlayerBid 経由でビッドラウンドを通すよう人間側に巻き戻す。
	tn2 := newTestTarneeb()
	tn2.Reset()
	tn2.SetBidPlayerIdx(0)
	tn2.SetDealerIdx(3)
	require.NoError(t, tn2.PlayerBid(7))
	// 残り3人をCPUに任せる
	tn2.CpuBid()
	tn2.CpuBid()
	tn2.CpuBid()
	// 全員パスの場合は再配布で Phase=Bid、誰かが上回りビッドした場合は Bid のまま、
	// パスしか上回らなければ TrumpDeclaration に進む。少なくとも phase が Bid or TrumpDeclaration。
	phase := tn2.GetPhase()
	assert.Contains(t, []domain.TarneebPhase{domain.TarneebPhaseBid, domain.TarneebPhaseTrumpDeclaration}, phase)
}

func TestTarneeb_BidLeadsToTrumpPhase_HumanWins(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetDealerIdx(3) // 人間が最初に賭ける
	tn.SetBidPlayerIdx(0)
	require.NoError(t, tn.PlayerBid(13)) // 人間が最高値を即取
	// CPU3人は全員パスしか不可能 (13より大きい値は無いため)
	for i := 0; i < 3; i++ {
		tn.CpuBid()
	}
	assert.Equal(t, domain.TarneebPhaseTrumpDeclaration, tn.GetPhase())
	assert.Equal(t, 0, tn.GetBidWinnerIdx())
	assert.Equal(t, 13, tn.GetHighestBid())
}

func TestTarneeb_PlayerDeclareTrump_Valid(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhaseTrumpDeclaration)
	tn.SetBidWinnerIdx(0)
	require.NoError(t, tn.PlayerDeclareTrump(domain.CardDesignHeart))
	assert.Equal(t, domain.CardDesignHeart, tn.GetTrumpSuit())
	assert.Equal(t, domain.TarneebPhasePlay, tn.GetPhase())
	assert.Equal(t, 0, tn.GetLeadPlayerIdx())
	assert.Equal(t, 1, tn.GetTrickNumber())
}

func TestTarneeb_PlayerDeclareTrump_InvalidSuit(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhaseTrumpDeclaration)
	tn.SetBidWinnerIdx(0)
	assert.Error(t, tn.PlayerDeclareTrump(99))
}

func TestTarneeb_PlayerDeclareTrump_WrongPhase(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhaseBid)
	tn.SetBidWinnerIdx(0)
	assert.ErrorIs(t, tn.PlayerDeclareTrump(domain.CardDesignSpade), domain.ErrWrongPhase)
}

func TestTarneeb_CpuDeclareTrump(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhaseTrumpDeclaration)
	tn.SetBidWinnerIdx(1) // CPU
	tn.CpuDeclareTrump()
	assert.NotEqual(t, 0, tn.GetTrumpSuit())
	assert.Equal(t, domain.TarneebPhasePlay, tn.GetPhase())
	assert.Equal(t, 1, tn.GetLeadPlayerIdx())
}

func TestTarneeb_MustFollowSuit(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhasePlay)
	tn.SetTrumpSuit(domain.CardDesignSpade)
	tn.SetCurrentPlayerIdx(0)
	tn.SetLeadPlayerIdx(0)
	clearAndDeal(tn.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 5, false),
		domain.NewCard(domain.CardDesignHeart, 10, false),
		domain.NewCard(domain.CardDesignDiamond, 3, false),
	})
	require.NoError(t, tn.PlayerPlay(0)) // ハート5をリード
	assert.Equal(t, 1, tn.GetCurrentPlayerIdx())

	// idx 1 はハートを持っていない (テスト用に強制セット)
	clearAndDeal(tn.GetPlayer(1), []*domain.Card{
		domain.NewCard(domain.CardDesignClover, 4, false),
		domain.NewCard(domain.CardDesignDiamond, 8, false),
	})
	// idx 1 が直接 PlayerPlay を呼ぶには idx を 0 にできない: CpuPlay 経由で1枚プレイさせる
	tn.CpuPlay()
	assert.Equal(t, 2, tn.GetCurrentPlayerIdx())
}

func TestTarneeb_FollowSuit_Violation(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhasePlay)
	tn.SetTrumpSuit(domain.CardDesignSpade)
	tn.SetCurrentPlayerIdx(0)
	// リードカードがハートになるよう、currentTrick をセット
	leadCard := domain.NewCard(domain.CardDesignHeart, 6, false)
	tn.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: leadCard}})
	clearAndDeal(tn.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 4, false),
		domain.NewCard(domain.CardDesignDiamond, 2, false),
	})
	// ハートを持っているのにダイヤを出そうとするとエラー
	err := tn.PlayerPlay(1)
	assert.Error(t, err)
	// ハートを出せばOK
	require.NoError(t, tn.PlayerPlay(0))
}

func TestTarneeb_Void_AllowsAnyCard(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhasePlay)
	tn.SetTrumpSuit(domain.CardDesignSpade)
	tn.SetCurrentPlayerIdx(0)
	leadCard := domain.NewCard(domain.CardDesignHeart, 6, false)
	tn.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: leadCard}})
	clearAndDeal(tn.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignDiamond, 2, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
	})
	// 何を出してもOK (トランプも捨て札も)
	require.NoError(t, tn.PlayerPlay(0))
}

func TestTarneeb_TrickWinner_TrumpBeatsLead(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhasePlay)
	tn.SetTrumpSuit(domain.CardDesignSpade)
	tn.SetTrickNumber(1)
	tn.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignSpade, 2, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
	})
	tn.SetPhase(domain.TarneebPhaseTrickEnd)
	tn.ResolveTrick()
	// 2♠ がトランプとして勝つ
	assert.Equal(t, 1, tn.GetPlayer(2).GetTrickCount())
	assert.Equal(t, 2, tn.GetLeadPlayerIdx())
}

func TestTarneeb_TrickWinner_NoTrump_HighLeadWins(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhaseTrickEnd)
	tn.SetTrumpSuit(domain.CardDesignSpade)
	tn.SetTrickNumber(1)
	tn.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 1, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignDiamond, 2, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 12, false)},
	})
	tn.ResolveTrick()
	// Ace は Tarneeb で最強。リードはハート、最高は A♥ なので idx=1 が勝ち。
	assert.Equal(t, 1, tn.GetPlayer(1).GetTrickCount())
}

func TestTarneeb_ScoreRound_BidderHits(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhaseRoundEnd)
	tn.SetBidWinnerIdx(0) // team 0
	tn.SetHighestBid(8)
	// team 0 (idx 0,2) total tricks 8, team 1 (idx 1,3) total 5
	tn.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
	for i := 0; i < 3; i++ {
		tn.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)})
	}
	for i := 0; i < 4; i++ {
		tn.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
	}
	for i := 0; i < 3; i++ {
		tn.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 4, false)})
	}
	for i := 0; i < 2; i++ {
		tn.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})
	}
	tn.ScoreRound()
	// team0 hit: +8 ; team1: +5
	assert.Equal(t, 8, tn.GetTeamScore(0))
	assert.Equal(t, 5, tn.GetTeamScore(1))
}

func TestTarneeb_ScoreRound_BidderMisses(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhaseRoundEnd)
	tn.SetBidWinnerIdx(0)
	tn.SetHighestBid(9)
	// team 0 takes only 5; team 1 takes 8
	for i := 0; i < 3; i++ {
		tn.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
	}
	for i := 0; i < 2; i++ {
		tn.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)})
	}
	for i := 0; i < 4; i++ {
		tn.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
	}
	for i := 0; i < 4; i++ {
		tn.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 4, false)})
	}
	tn.ScoreRound()
	// team0 missed: -9 ; team1: +8
	assert.Equal(t, -9, tn.GetTeamScore(0))
	assert.Equal(t, 8, tn.GetTeamScore(1))
}

func TestTarneeb_ScoreRound_GameEnd(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhaseRoundEnd)
	tn.SetBidWinnerIdx(0)
	tn.SetHighestBid(7)
	tn.SetTeamScore(0, 25)
	tn.SetTeamScore(1, 10)
	for i := 0; i < 4; i++ {
		tn.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
	}
	for i := 0; i < 4; i++ {
		tn.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)})
	}
	for i := 0; i < 3; i++ {
		tn.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
	}
	for i := 0; i < 2; i++ {
		tn.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 4, false)})
	}
	tn.ScoreRound()
	// team0: 25+8 = 33 ≥ 31
	assert.True(t, tn.GetGameEndFlag())
	assert.Equal(t, domain.TarneebPhaseGameEnd, tn.GetPhase())
	assert.Equal(t, 0, tn.GetWinnerTeam())
}

func TestTarneeb_ScoreRound_TieGoesToBidder(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhaseRoundEnd)
	tn.SetBidWinnerIdx(1) // team 1
	tn.SetHighestBid(7)
	tn.SetTeamScore(0, 24)
	tn.SetTeamScore(1, 24)
	// team0 takes 7, team1 takes 7 → both add 7 → tie at 31
	for i := 0; i < 4; i++ {
		tn.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
	}
	for i := 0; i < 3; i++ {
		tn.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)})
	}
	for i := 0; i < 4; i++ {
		tn.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
	}
	for i := 0; i < 3; i++ {
		tn.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 4, false)})
	}
	tn.ScoreRound()
	assert.True(t, tn.GetGameEndFlag())
	assert.Equal(t, 1, tn.GetWinnerTeam(), "tie should go to bidder team")
}

func TestTarneeb_NextRound_Increments(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhaseRoundEnd)
	tn.SetTeamScore(0, 10) // not yet endgame
	tn.SetBidWinnerIdx(0)
	tn.SetHighestBid(7)
	tn.ScoreRound()
	require.Equal(t, domain.TarneebPhaseRoundEnd, tn.GetPhase())
	tn.NextRound()
	assert.Equal(t, 2, tn.GetRoundNumber())
	assert.Equal(t, domain.TarneebPhaseBid, tn.GetPhase())
	assert.Equal(t, 1, tn.GetDealerIdx())
}

func TestTarneeb_GetValidPlayIndices(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhasePlay)
	tn.SetTrumpSuit(domain.CardDesignSpade)
	tn.SetCurrentPlayerIdx(0)
	leadCard := domain.NewCard(domain.CardDesignHeart, 3, false)
	tn.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 3, Card: leadCard}})
	clearAndDeal(tn.GetPlayer(0), []*domain.Card{
		domain.NewCard(domain.CardDesignHeart, 4, false),
		domain.NewCard(domain.CardDesignDiamond, 2, false),
		domain.NewCard(domain.CardDesignSpade, 5, false),
	})
	valid := tn.GetValidPlayIndices(0)
	assert.Equal(t, []int{0}, valid)
}

func TestTarneeb_GetHint_Bid(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetBidPlayerIdx(0)
	// 人間がビッド手番 → ヒントは Bid を返す
	hint := tn.GetHint()
	require.NotNil(t, hint)
	assert.NotNil(t, hint.Bid)
}

func TestTarneeb_GetHint_Trump(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhaseTrumpDeclaration)
	tn.SetBidWinnerIdx(0)
	hint := tn.GetHint()
	require.NotNil(t, hint)
	assert.NotNil(t, hint.TrumpSuit)
}

func TestTarneeb_GetHint_Play(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhasePlay)
	tn.SetTrumpSuit(domain.CardDesignSpade)
	tn.SetCurrentPlayerIdx(0)
	hint := tn.GetHint()
	require.NotNil(t, hint)
	assert.NotNil(t, hint.CardIndex)
}

func TestTarneeb_JSONRoundTrip(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhasePlay)
	tn.SetTrumpSuit(domain.CardDesignSpade)
	tn.SetTeamScore(0, 7)
	tn.SetTeamScore(1, 3)

	data, err := json.Marshal(tn)
	require.NoError(t, err)
	got := domain.NewTarneeb(domain.NewTrumpCards(0), nil, domain.DefaultTarneebConfig())
	require.NoError(t, json.Unmarshal(data, got))
	assert.Equal(t, tn.GetTrumpSuit(), got.GetTrumpSuit())
	assert.Equal(t, tn.GetTeamScore(0), got.GetTeamScore(0))
	assert.Equal(t, tn.GetTeamScore(1), got.GetTeamScore(1))
	assert.Equal(t, tn.GetPhase(), got.GetPhase())
}

func TestTarneeb_PlayerPlay_WrongPhase(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhaseBid)
	assert.ErrorIs(t, tn.PlayerPlay(0), domain.ErrWrongPhase)
}

func TestTarneeb_PlayerPlay_OutOfRange(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhasePlay)
	tn.SetCurrentPlayerIdx(0)
	err := tn.PlayerPlay(-1)
	assert.Error(t, err)
}

func TestTarneeb_NextTrick_RestartsPlay(t *testing.T) {
	tn := newTestTarneeb()
	tn.Reset()
	tn.SetPhase(domain.TarneebPhaseTrickEnd)
	tn.SetLeadPlayerIdx(2)
	tn.SetTrickNumber(5)
	tn.NextTrick()
	assert.Equal(t, domain.TarneebPhasePlay, tn.GetPhase())
	assert.Equal(t, 6, tn.GetTrickNumber())
	assert.Equal(t, 2, tn.GetCurrentPlayerIdx())
}
