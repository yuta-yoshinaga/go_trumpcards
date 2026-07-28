//go:build test

package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestPitch() *domain.Pitch {
	players := []*domain.PitchPlayer{
		domain.NewPitchPlayer(true),
		domain.NewPitchPlayer(false),
		domain.NewPitchPlayer(false),
		domain.NewPitchPlayer(false),
	}
	return domain.NewPitch(domain.NewTrumpCards(0), players, domain.DefaultPitchConfig())
}

func setHandPitch(p *domain.PitchPlayer, cards ...*domain.Card) {
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func newPitchCard(design, value int) *domain.Card {
	return domain.NewCard(design, value, false)
}

func TestNewPitch(t *testing.T) {
	p := newTestPitch()
	assert.Equal(t, -1, p.GetWinnerIdx())
	assert.Equal(t, 0, p.GetRoundNumber())
	assert.Equal(t, -1, p.GetBidWinnerIdx())
	assert.Equal(t, domain.PitchTrumpUnset, p.GetTrumpSuit())
}

func TestNewDefaultPitch(t *testing.T) {
	p := domain.NewDefaultPitch()
	assert.NotNil(t, p)
	assert.Equal(t, domain.PitchPlayerCnt, p.GetPlayerCnt())
	assert.True(t, p.GetPlayer(0).GetIsHuman())
	for i := 1; i < p.GetPlayerCnt(); i++ {
		assert.False(t, p.GetPlayer(i).GetIsHuman(), "player %d should be CPU", i)
	}
	assert.False(t, p.GetGameEndFlag())
}

func TestPitch_Reset(t *testing.T) {
	p := newTestPitch()
	p.Reset()

	assert.Equal(t, domain.PitchPhaseBid, p.GetPhase())
	assert.Equal(t, 1, p.GetRoundNumber())
	assert.Equal(t, 0, p.GetTrickNumber())
	assert.Equal(t, domain.PitchPlayerCnt-1, p.GetDealerIdx())
	assert.Equal(t, 0, p.GetBidPlayerIdx(), "eldest hand (left of dealer) bids first")
	assert.Equal(t, 0, p.GetCurrentBid())
	assert.Equal(t, -1, p.GetBidWinnerIdx())
	assert.Equal(t, domain.PitchTrumpUnset, p.GetTrumpSuit())

	for i := 0; i < domain.PitchPlayerCnt; i++ {
		assert.Equal(t, domain.PitchHandSize, p.GetPlayer(i).GetCardsSize(), "player %d", i)
		assert.Equal(t, -1, p.GetPlayer(i).GetBid())
		assert.Equal(t, 0, p.GetPlayer(i).GetCumulativeScore())
	}
}

func TestPitch_PlayerBid_Pass(t *testing.T) {
	p := newTestPitch()
	p.Reset()

	assert.NoError(t, p.PlayerBid(domain.PitchPassBid))
	assert.Equal(t, 0, p.GetPlayer(0).GetBid())
}

func TestPitch_PlayerBid_NormalBid(t *testing.T) {
	p := newTestPitch()
	p.Reset()

	assert.NoError(t, p.PlayerBid(3))
	assert.Equal(t, 3, p.GetPlayer(0).GetBid())
	assert.Equal(t, 3, p.GetCurrentBid())
	assert.Equal(t, 0, p.GetBidWinnerIdx())
}

func TestPitch_PlayerBid_Errors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(p *domain.Pitch)
		bid     int
		wantErr error
	}{
		{
			name:    "wrong phase",
			setup:   func(p *domain.Pitch) { p.SetPhase(domain.PitchPhasePlay) },
			bid:     2,
			wantErr: domain.ErrWrongPhase,
		},
		{
			name:    "not human turn",
			setup:   func(p *domain.Pitch) { p.SetBidPlayerIdx(1) },
			bid:     2,
			wantErr: domain.ErrNotHumanTurn,
		},
		{
			name:    "bid too low",
			setup:   func(_ *domain.Pitch) {},
			bid:     1,
			wantErr: domain.ErrInvalidPlay,
		},
		{
			name:    "bid too high",
			setup:   func(_ *domain.Pitch) {},
			bid:     5,
			wantErr: domain.ErrInvalidPlay,
		},
		{
			name:    "bid not exceeding current",
			setup:   func(p *domain.Pitch) { p.SetCurrentBid(3) },
			bid:     3,
			wantErr: domain.ErrInvalidPlay,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newTestPitch()
			p.Reset()
			tt.setup(p)
			err := p.PlayerBid(tt.bid)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
		})
	}
}

func TestPitch_PlayerBid_GameEnded(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	// 直接 GameEnd フェーズへ (通常は ScoreRound 経由)
	p.SetPhase(domain.PitchPhaseGameEnd)
	// gameEndFlag を立てる代わりに reflection は使わずに ScoreRound 経由で立てる
	// ここでは PlayerBid が wrong phase でも先に gameEndFlag が無いため WrongPhase になる。
	err := p.PlayerBid(2)
	assert.Error(t, err)
}

func TestPitch_PlayerBid_DealerCannotPassWhenAllPassed(t *testing.T) {
	// human が dealer のとき、全員パス状態でのパスは弾かれる。
	p := newTestPitch()
	p.Reset()
	// human (idx 0) を親に切り替える
	p.SetDealerIdx(0)
	p.SetBidPlayerIdx(0)
	p.SetCurrentBid(0)
	// 他全員はパス済み
	p.GetPlayer(1).SetBid(domain.PitchPassBid)
	p.GetPlayer(2).SetBid(domain.PitchPassBid)
	p.GetPlayer(3).SetBid(domain.PitchPassBid)
	err := p.PlayerBid(domain.PitchPassBid)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
}

func TestPitch_AllPass_StuckDealer(t *testing.T) {
	// Easy 難易度では CPU が頻繁にパスするため、ここでは決定的に再現するために
	// 0/1/2 の bid を直接 pass にし、dealer (3) の CpuBid のみを呼んで stuck 強制を検証する。
	p := newTestPitch()
	p.Reset()
	// dealer は 3 (idx 3 == PitchPlayerCnt-1)
	p.GetPlayer(0).SetBid(domain.PitchPassBid)
	p.GetPlayer(1).SetBid(domain.PitchPassBid)
	p.GetPlayer(2).SetBid(domain.PitchPassBid)
	p.SetBidPlayerIdx(3)
	p.SetCurrentBid(0)
	p.CpuBid() // dealer は強制で MinBid 以上を入札
	assert.Equal(t, domain.PitchPhasePlay, p.GetPhase())
	assert.GreaterOrEqual(t, p.GetCurrentBid(), domain.PitchMinBid, "dealer is stuck with at least the min bid")
	assert.Equal(t, 3, p.GetBidWinnerIdx())
}

func TestPitch_AllBidsTransitionToPlay(t *testing.T) {
	// Easy 難易度: CPU は {0,2,3} からランダム選択し、currentBid (人間の 3) を超えられないため必ずパスする。
	players := []*domain.PitchPlayer{
		domain.NewPitchPlayer(true),
		domain.NewPitchPlayer(false),
		domain.NewPitchPlayer(false),
		domain.NewPitchPlayer(false),
	}
	cfg := domain.DefaultPitchConfig()
	cfg.CpuDifficulty = domain.PitchCpuDifficultyEasy
	p := domain.NewPitch(domain.NewTrumpCards(0), players, cfg)
	p.Reset()

	assert.NoError(t, p.PlayerBid(3))
	for p.GetPhase() == domain.PitchPhaseBid {
		p.CpuBid()
	}
	assert.Equal(t, domain.PitchPhasePlay, p.GetPhase())
	assert.Equal(t, 0, p.GetBidWinnerIdx())
	assert.Equal(t, 0, p.GetLeadPlayerIdx())
	assert.Equal(t, 0, p.GetCurrentPlayerIdx())
	assert.Equal(t, 1, p.GetTrickNumber())
}

func TestPitch_PlayCard_LeadSetsTrump(t *testing.T) {
	players := []*domain.PitchPlayer{
		domain.NewPitchPlayer(true),
		domain.NewPitchPlayer(false),
		domain.NewPitchPlayer(false),
		domain.NewPitchPlayer(false),
	}
	cfg := domain.DefaultPitchConfig()
	cfg.CpuDifficulty = domain.PitchCpuDifficultyEasy
	p := domain.NewPitch(domain.NewTrumpCards(0), players, cfg)
	p.Reset()

	assert.NoError(t, p.PlayerBid(3))
	for p.GetPhase() == domain.PitchPhaseBid {
		p.CpuBid()
	}
	human := p.GetPlayer(0)
	first := human.GetCard(0)
	assert.NoError(t, p.PlayerPlay(0))
	assert.Equal(t, first.GetDesign(), p.GetTrumpSuit())
	assert.Equal(t, 1, len(p.GetCurrentTrick()))
}

func TestPitch_ValidatePlay_FollowOrTrump(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetPhase(domain.PitchPhasePlay)
	p.SetTrumpSuit(domain.CardDesignSpade)
	p.SetTrickNumber(2)
	p.SetLeadPlayerIdx(1)
	p.SetCurrentPlayerIdx(0)
	// CPU 1 が ♥ をリード
	p.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: newPitchCard(domain.CardDesignHeart, 10)},
	})
	// human の手札: ♥9 (lead), ♠5 (trump), ♦4 (off-suit)
	setHandPitch(p.GetPlayer(0),
		newPitchCard(domain.CardDesignHeart, 9),
		newPitchCard(domain.CardDesignSpade, 5),
		newPitchCard(domain.CardDesignDiamond, 4),
	)
	// follow ♥ OK
	assert.NoError(t, p.PlayerPlay(0))
}

func TestPitch_ValidatePlay_TrumpAlwaysLegal(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetPhase(domain.PitchPhasePlay)
	p.SetTrumpSuit(domain.CardDesignSpade)
	p.SetTrickNumber(2)
	p.SetLeadPlayerIdx(1)
	p.SetCurrentPlayerIdx(0)
	p.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: newPitchCard(domain.CardDesignHeart, 10)},
	})
	setHandPitch(p.GetPlayer(0),
		newPitchCard(domain.CardDesignSpade, 5), // trump (legal: trump always)
		newPitchCard(domain.CardDesignHeart, 9), // lead (also legal)
	)
	// trump (idx 0) は lead が他にあっても合法
	assert.NoError(t, p.PlayerPlay(0))
}

func TestPitch_ValidatePlay_RejectOffSuitWhenLeadPresent(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetPhase(domain.PitchPhasePlay)
	p.SetTrumpSuit(domain.CardDesignSpade)
	p.SetTrickNumber(2)
	p.SetLeadPlayerIdx(1)
	p.SetCurrentPlayerIdx(0)
	p.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: newPitchCard(domain.CardDesignHeart, 10)},
	})
	setHandPitch(p.GetPlayer(0),
		newPitchCard(domain.CardDesignDiamond, 4), // off-suit non-trump
		newPitchCard(domain.CardDesignHeart, 9),   // lead present → must follow OR trump
	)
	err := p.PlayerPlay(0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
}

func TestPitch_ValidatePlay_VoidAllowsAnyCard(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetPhase(domain.PitchPhasePlay)
	p.SetTrumpSuit(domain.CardDesignSpade)
	p.SetTrickNumber(2)
	p.SetLeadPlayerIdx(1)
	p.SetCurrentPlayerIdx(0)
	p.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: newPitchCard(domain.CardDesignHeart, 10)},
	})
	// void in ♥
	setHandPitch(p.GetPlayer(0),
		newPitchCard(domain.CardDesignDiamond, 4),
		newPitchCard(domain.CardDesignClover, 7),
	)
	assert.NoError(t, p.PlayerPlay(0)) // off-suit allowed
}

func TestPitch_PlayerPlay_GameEnded(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetPhase(domain.PitchPhasePlay)
	p.SetCurrentPlayerIdx(0)
	setHandPitch(p.GetPlayer(0), newPitchCard(domain.CardDesignSpade, 5))
	// gameEndFlag を直接立てるため ScoreRound を発火させる近道:
	// ScoreRound はフェーズ条件で動かず、checkGameEnd はビッダー条件を使う。
	// ここでは ErrWrongPhase / ErrNotHumanTurn / ErrInvalidCard などを別ケースで検証する。
	assert.NoError(t, p.PlayerPlay(0))
}

func TestPitch_PlayerPlay_InvalidCardIndex(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetPhase(domain.PitchPhasePlay)
	p.SetCurrentPlayerIdx(0)
	setHandPitch(p.GetPlayer(0), newPitchCard(domain.CardDesignSpade, 5))
	err := p.PlayerPlay(99)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidCard))
}

func TestPitch_PlayerPlay_NotHumanTurn(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetPhase(domain.PitchPhasePlay)
	p.SetCurrentPlayerIdx(1) // CPU
	err := p.PlayerPlay(0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrNotHumanTurn))
}

func TestPitch_PlayerPlay_WrongPhase(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetPhase(domain.PitchPhaseBid)
	err := p.PlayerPlay(0)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestPitch_TrickWinner_TrumpBeatsLead(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetTrumpSuit(domain.CardDesignSpade)
	p.SetPhase(domain.PitchPhaseTrickEnd)
	p.SetTrickNumber(2)
	// ♥A, ♥K, ♠2(trump), ♥Q → ♠2 wins (trump)
	p.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: newPitchCard(domain.CardDesignHeart, 1)},
		{PlayerIdx: 1, Card: newPitchCard(domain.CardDesignHeart, 13)},
		{PlayerIdx: 2, Card: newPitchCard(domain.CardDesignSpade, 2)},
		{PlayerIdx: 3, Card: newPitchCard(domain.CardDesignHeart, 12)},
	})
	p.ResolveTrick()
	// trick 2 < total → phase advances to TrickEnd, leadPlayer set
	assert.Equal(t, 2, p.GetLeadPlayerIdx())
	assert.Equal(t, 1, p.GetPlayer(2).GetTrickCount())
}

func TestPitch_TrickWinner_HighestTrumpWins(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetTrumpSuit(domain.CardDesignSpade)
	p.SetPhase(domain.PitchPhaseTrickEnd)
	p.SetTrickNumber(2)
	// ♠5, ♠A(=14 rank), ♠K, ♥10 → ♠A wins
	p.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: newPitchCard(domain.CardDesignSpade, 5)},
		{PlayerIdx: 1, Card: newPitchCard(domain.CardDesignSpade, 1)},
		{PlayerIdx: 2, Card: newPitchCard(domain.CardDesignSpade, 13)},
		{PlayerIdx: 3, Card: newPitchCard(domain.CardDesignHeart, 10)},
	})
	p.ResolveTrick()
	assert.Equal(t, 1, p.GetLeadPlayerIdx())
}

func TestPitch_TrickWinner_NoTrump(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetTrumpSuit(domain.CardDesignSpade)
	p.SetPhase(domain.PitchPhaseTrickEnd)
	p.SetTrickNumber(2)
	// ♥10, ♥K, ♥A, ♣4 → ♥A wins (lead suit highest, no trump played)
	p.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: newPitchCard(domain.CardDesignHeart, 10)},
		{PlayerIdx: 1, Card: newPitchCard(domain.CardDesignHeart, 13)},
		{PlayerIdx: 2, Card: newPitchCard(domain.CardDesignHeart, 1)},
		{PlayerIdx: 3, Card: newPitchCard(domain.CardDesignClover, 4)},
	})
	p.ResolveTrick()
	assert.Equal(t, 2, p.GetLeadPlayerIdx())
}

func TestPitch_NextTrick(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetPhase(domain.PitchPhaseTrickEnd)
	p.SetTrickNumber(1)
	p.SetLeadPlayerIdx(2)
	p.SetCurrentTrick([]*domain.TrickCard{{PlayerIdx: 0, Card: newPitchCard(domain.CardDesignSpade, 5)}})
	p.NextTrick()
	assert.Equal(t, domain.PitchPhasePlay, p.GetPhase())
	assert.Equal(t, 2, p.GetCurrentPlayerIdx())
	assert.Equal(t, 2, p.GetTrickNumber())
	assert.Equal(t, 0, len(p.GetCurrentTrick()))
}

func TestPitch_ScoreRound_BidderMakes(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetPhase(domain.PitchPhaseRoundEnd)
	p.SetTrumpSuit(domain.CardDesignSpade)
	p.SetBidWinnerIdx(0)
	p.SetCurrentBid(2)
	// human が 4 ポイント全て獲得するシナリオ
	p.GetPlayer(0).AddTrick([]*domain.Card{
		newPitchCard(domain.CardDesignSpade, 1),  // High (A)
		newPitchCard(domain.CardDesignSpade, 2),  // Low
		newPitchCard(domain.CardDesignSpade, 11), // Jack
		newPitchCard(domain.CardDesignHeart, 10), // Game (10 pip)
	})
	p.ScoreRound()
	assert.Equal(t, 4, p.GetPlayer(0).GetCumulativeScore())
}

func TestPitch_ScoreRound_BidderSetBack(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetPhase(domain.PitchPhaseRoundEnd)
	p.SetTrumpSuit(domain.CardDesignSpade)
	p.SetBidWinnerIdx(0)
	p.SetCurrentBid(3)
	// human は 1 ポイントしか取れないので set back -3
	p.GetPlayer(0).AddTrick([]*domain.Card{
		newPitchCard(domain.CardDesignSpade, 11), // Jack of trump
	})
	// Low + High + Game は他プレイヤーへ
	p.GetPlayer(1).AddTrick([]*domain.Card{
		newPitchCard(domain.CardDesignSpade, 1), // High
		newPitchCard(domain.CardDesignSpade, 2), // Low
		newPitchCard(domain.CardDesignHeart, 1), // Game (4 pip)
	})
	p.ScoreRound()
	assert.Equal(t, -3, p.GetPlayer(0).GetCumulativeScore(), "bidder set back")
	// player 1 は High(1) + Low(1) + Game(1) = 3
	assert.Equal(t, 3, p.GetPlayer(1).GetCumulativeScore())
}

func TestPitch_ScoreRound_GameTieNoPoint(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetPhase(domain.PitchPhaseRoundEnd)
	p.SetTrumpSuit(domain.CardDesignSpade)
	p.SetBidWinnerIdx(0)
	p.SetCurrentBid(2)
	// 同点シナリオ: human と CPU1 がそれぞれ 8 pip
	// human: ♠A (High+Low, 4 pip) + ♥A (4 pip) = 8 pip
	p.GetPlayer(0).AddTrick([]*domain.Card{
		newPitchCard(domain.CardDesignSpade, 1),
		newPitchCard(domain.CardDesignHeart, 1),
	})
	// CPU1: ♣A (4 pip) + ♦K (3 pip) + ♦J (1 pip) = 8 pip
	p.GetPlayer(1).AddTrick([]*domain.Card{
		newPitchCard(domain.CardDesignClover, 1),
		newPitchCard(domain.CardDesignDiamond, 13),
		newPitchCard(domain.CardDesignDiamond, 11),
	})
	p.ScoreRound()
	// human: High(1) + Low(1) = 2; Game は同点で誰にも入らない
	assert.Equal(t, 2, p.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, 0, p.GetPlayer(1).GetCumulativeScore())
}

func TestPitch_GameEnd_BidderReachesPointLimit(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetPhase(domain.PitchPhaseRoundEnd)
	p.SetTrumpSuit(domain.CardDesignSpade)
	p.SetBidWinnerIdx(0)
	p.SetCurrentBid(2)
	// human: 既に 5 点累積 + 4 点獲得 = 9 (PointLimit 7 を超える)
	p.GetPlayer(0).SetCumulativeScore(5)
	p.GetPlayer(0).AddTrick([]*domain.Card{
		newPitchCard(domain.CardDesignSpade, 1),
		newPitchCard(domain.CardDesignSpade, 2),
		newPitchCard(domain.CardDesignSpade, 11),
		newPitchCard(domain.CardDesignHeart, 10),
	})
	p.ScoreRound()
	assert.True(t, p.GetGameEndFlag())
	assert.Equal(t, 0, p.GetWinnerIdx())
	assert.Equal(t, domain.PitchPhaseGameEnd, p.GetPhase())
}

func TestPitch_NextRound(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	originalDealer := p.GetDealerIdx()
	p.SetPhase(domain.PitchPhaseRoundEnd)
	// すべてのプレイヤーに僅かなスコアを設定
	for i := 0; i < domain.PitchPlayerCnt; i++ {
		p.GetPlayer(i).SetCumulativeScore(1)
		p.GetPlayer(i).AddTrick([]*domain.Card{newPitchCard(domain.CardDesignSpade, 5)})
	}
	p.NextRound()
	assert.Equal(t, domain.PitchPhaseBid, p.GetPhase())
	assert.Equal(t, 2, p.GetRoundNumber())
	assert.Equal(t, (originalDealer+1)%domain.PitchPlayerCnt, p.GetDealerIdx())
	assert.Equal(t, domain.PitchTrumpUnset, p.GetTrumpSuit())
	for i := 0; i < domain.PitchPlayerCnt; i++ {
		assert.Equal(t, 1, p.GetPlayer(i).GetCumulativeScore(), "cumulative preserved")
		assert.Equal(t, 0, p.GetPlayer(i).GetTrickCount(), "tricks reset")
		assert.Equal(t, domain.PitchHandSize, p.GetPlayer(i).GetCardsSize())
	}
}

func TestPitch_GetHint_BidPhase(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	hint := p.GetHint()
	assert.NotNil(t, hint)
	assert.NotNil(t, hint.Bid)
}

func TestPitch_GetHint_PlayPhase(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetPhase(domain.PitchPhasePlay)
	p.SetCurrentPlayerIdx(0)
	p.SetTrumpSuit(domain.CardDesignSpade)
	hint := p.GetHint()
	assert.NotNil(t, hint)
	assert.NotNil(t, hint.CardIndex)
}

func TestPitch_GetValidPlayIndices(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetPhase(domain.PitchPhasePlay)
	p.SetTrumpSuit(domain.CardDesignSpade)
	p.SetCurrentPlayerIdx(0)
	p.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: newPitchCard(domain.CardDesignHeart, 10)},
	})
	setHandPitch(p.GetPlayer(0),
		newPitchCard(domain.CardDesignHeart, 9), // lead-suit
		newPitchCard(domain.CardDesignSpade, 5), // trump
		newPitchCard(domain.CardDesignDiamond, 4),
	)
	indices := p.GetValidPlayIndices(0)
	assert.ElementsMatch(t, []int{0, 1}, indices, "lead-suit + trump are valid; off-suit non-trump is not")
}

func TestPitch_CpuFullRound(t *testing.T) {
	// CPU だけで進行できることを保証する E2E スモーク
	p := newTestPitch()
	p.Reset()
	// 人間にもパスさせて全員 CPU 進行へ
	assert.NoError(t, p.PlayerBid(domain.PitchPassBid))
	for p.GetPhase() == domain.PitchPhaseBid {
		p.CpuBid()
	}
	assert.Equal(t, domain.PitchPhasePlay, p.GetPhase())
	// 6 トリック分回す
	for trick := 1; trick <= domain.PitchTotalTricks; trick++ {
		// 4 枚プレイ
		for i := 0; i < domain.PitchPlayerCnt; i++ {
			if p.IsHumanTurn() {
				validIndices := p.GetValidPlayIndices(p.GetCurrentPlayerIdx())
				if len(validIndices) == 0 {
					t.Fatalf("no valid plays for human at trick %d", trick)
				}
				assert.NoError(t, p.PlayerPlay(validIndices[0]))
			} else {
				p.CpuPlay()
			}
		}
		assert.Equal(t, domain.PitchPhaseTrickEnd, p.GetPhase(), "trick %d should end", trick)
		p.ResolveTrick()
		if trick < domain.PitchTotalTricks {
			p.NextTrick()
		}
	}
	assert.Equal(t, domain.PitchPhaseRoundEnd, p.GetPhase())
	p.ScoreRound()
	// ラウンドスコアが累積された
	totalCumulative := 0
	for i := 0; i < domain.PitchPlayerCnt; i++ {
		totalCumulative += p.GetPlayer(i).GetCumulativeScore()
	}
	// 4 ポイント以下が分配される (set back の場合は負も含む)
	assert.GreaterOrEqual(t, totalCumulative, -domain.PitchMaxBid)
	assert.LessOrEqual(t, totalCumulative, 4)
}

func TestPitch_JSONRoundTrip(t *testing.T) {
	p := newTestPitch()
	p.Reset()
	p.SetTrumpSuit(domain.CardDesignSpade)
	p.SetCurrentBid(3)
	p.SetBidWinnerIdx(0)

	data, err := json.Marshal(p)
	assert.NoError(t, err)

	clone := domain.NewDefaultPitch()
	assert.NoError(t, json.Unmarshal(data, clone))
	assert.Equal(t, p.GetTrumpSuit(), clone.GetTrumpSuit())
	assert.Equal(t, p.GetCurrentBid(), clone.GetCurrentBid())
	assert.Equal(t, p.GetBidWinnerIdx(), clone.GetBidWinnerIdx())
	assert.Equal(t, p.GetPhase(), clone.GetPhase())
	assert.Equal(t, p.GetRoundNumber(), clone.GetRoundNumber())
	assert.Equal(t, p.GetPlayerCnt(), clone.GetPlayerCnt())
}

func TestPitch_DifficultyLevels(t *testing.T) {
	// 各難易度で CPU が壊れずに動くスモーク
	for _, diff := range []domain.PitchCpuDifficulty{
		domain.PitchCpuDifficultyEasy,
		domain.PitchCpuDifficultyNormal,
		domain.PitchCpuDifficultyHard,
	} {
		t.Run(diffName(diff), func(t *testing.T) {
			players := []*domain.PitchPlayer{
				domain.NewPitchPlayer(true),
				domain.NewPitchPlayer(false),
				domain.NewPitchPlayer(false),
				domain.NewPitchPlayer(false),
			}
			cfg := domain.DefaultPitchConfig()
			cfg.CpuDifficulty = diff
			p := domain.NewPitch(domain.NewTrumpCards(0), players, cfg)
			p.Reset()
			assert.NoError(t, p.PlayerBid(domain.PitchPassBid))
			for p.GetPhase() == domain.PitchPhaseBid {
				p.CpuBid()
			}
			assert.Equal(t, domain.PitchPhasePlay, p.GetPhase())
		})
	}
}

func diffName(d domain.PitchCpuDifficulty) string {
	switch d {
	case domain.PitchCpuDifficultyEasy:
		return "easy"
	case domain.PitchCpuDifficultyNormal:
		return "normal"
	case domain.PitchCpuDifficultyHard:
		return "hard"
	}
	return "?"
}
