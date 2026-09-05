//go:build test

package domain_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestBatak() *domain.Batak {
	players := []*domain.BatakPlayer{
		domain.NewBatakPlayer(true),
		domain.NewBatakPlayer(false),
		domain.NewBatakPlayer(false),
		domain.NewBatakPlayer(false),
	}
	return domain.NewBatak(domain.NewTrumpCards(0), players, domain.DefaultBatakConfig())
}

func setupBatakBid(cb *domain.Batak, idx int) {
	cb.SetPhase(domain.BatakPhaseBid)
	cb.SetBidPlayerIdx(idx)
}

func setupBatakPlay(cb *domain.Batak, current, lead, trickNum int) {
	cb.SetPhase(domain.BatakPhasePlay)
	cb.SetCurrentPlayerIdx(current)
	cb.SetLeadPlayerIdx(lead)
	cb.SetTrickNumber(trickNum)
}

func TestNewBatak(t *testing.T) {
	cb := newTestBatak()
	assert.Equal(t, -1, cb.GetWinnerIdx())
	assert.Equal(t, 0, cb.GetRoundNumber())
}

func TestNewDefaultBatak(t *testing.T) {
	cb := domain.NewDefaultBatak()
	require.NotNil(t, cb)
	assert.Equal(t, domain.BatakPlayerCnt, cb.GetPlayerCnt())
	assert.True(t, cb.GetPlayer(0).GetIsHuman())
	for i := 1; i < cb.GetPlayerCnt(); i++ {
		assert.False(t, cb.GetPlayer(i).GetIsHuman())
	}
	assert.False(t, cb.GetGameEndFlag())
	assert.Equal(t, domain.DefaultBatakConfig(), cb.GetConfig())
}

func TestBatak_Reset(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()

	assert.Equal(t, domain.BatakPhaseBid, cb.GetPhase())
	assert.Equal(t, 1, cb.GetRoundNumber())
	assert.Equal(t, 0, cb.GetTrickNumber())
	assert.False(t, cb.GetSpadesBroken())
	assert.False(t, cb.GetGameEndFlag())
	assert.Equal(t, -1, cb.GetWinnerIdx())
	assert.Equal(t, 0, cb.GetBidStartIdx())
	assert.Equal(t, 0, cb.GetBidPlayerIdx())
	assert.Equal(t, -1, cb.GetDeclarerIdx())
	assert.Equal(t, 0, cb.GetHighBid())
	assert.Equal(t, 0, cb.GetLeadPlayerIdx())

	for i := 0; i < 4; i++ {
		assert.Equal(t, 13, cb.GetPlayer(i).GetCardsSize())
		assert.Equal(t, -1, cb.GetPlayer(i).GetBid())
		assert.Equal(t, 0, cb.GetPlayer(i).GetCumulativeScore())
	}
}

func TestBatak_Reset_ClearsAccumulated(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()

	cb.GetPlayer(0).SetCumulativeScore(123)
	cb.SetPhase(domain.BatakPhaseGameEnd)

	cb.Reset()
	assert.Equal(t, 0, cb.GetPlayer(0).GetCumulativeScore())
	assert.Equal(t, domain.BatakPhaseBid, cb.GetPhase())
}

func TestBatak_PlayerBid_Valid(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	setupBatakBid(cb, 0)
	err := cb.PlayerBid(5)
	require.NoError(t, err)
	assert.Equal(t, 5, cb.GetPlayer(0).GetBid())
	assert.Equal(t, 5, cb.GetHighBid())
}

func TestBatak_PlayerBid_Pass(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	setupBatakBid(cb, 0)
	err := cb.PlayerBid(domain.BatakPassBid)
	require.NoError(t, err)
	assert.Equal(t, domain.BatakPassBid, cb.GetPlayer(0).GetBid())
	assert.Equal(t, 0, cb.GetHighBid())
}

func TestBatak_PlayerBid_MaxBid(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	setupBatakBid(cb, 0)
	err := cb.PlayerBid(domain.BatakMaxBid)
	require.NoError(t, err)
	assert.Equal(t, domain.BatakMaxBid, cb.GetPlayer(0).GetBid())
	assert.Equal(t, domain.BatakMaxBid, cb.GetHighBid())
}

func TestBatak_Auction_InvalidBids(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	setupBatakBid(cb, 0)

	// BatakMinBid (5) 未満の非ゼロ宣言はエラー
	assert.Error(t, cb.PlayerBid(4))
	assert.Error(t, cb.PlayerBid(1))
	assert.Error(t, cb.PlayerBid(-1))

	// BatakMaxBid (13) 超過はエラー
	assert.Error(t, cb.PlayerBid(14))

	// 現在の highBid 以下の宣言はエラー、highBid+1 は通る
	cb.SetHighBid(6)
	assert.Error(t, cb.PlayerBid(6), "PlayerBid(highBid) should error")
	assert.NoError(t, cb.PlayerBid(7), "PlayerBid(highBid+1) should succeed")

	// 13 宣言後の MinLegalBid は 0 (パスのみ可能)
	cb.SetHighBid(13)
	assert.Equal(t, domain.BatakPassBid, cb.MinLegalBid())
	cb.SetBidPlayerIdx(0)
	assert.Error(t, cb.PlayerBid(13))
	assert.NoError(t, cb.PlayerBid(domain.BatakPassBid))
}

func TestBatak_MinLegalBid(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	assert.Equal(t, domain.BatakMinBid, cb.MinLegalBid())

	cb.SetHighBid(5)
	assert.Equal(t, 6, cb.MinLegalBid())

	cb.SetHighBid(12)
	assert.Equal(t, 13, cb.MinLegalBid())

	cb.SetHighBid(13)
	assert.Equal(t, domain.BatakPassBid, cb.MinLegalBid())
}

func TestBatak_PlayerBid_WrongPhase(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	cb.SetPhase(domain.BatakPhasePlay)
	err := cb.PlayerBid(5)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestBatak_PlayerBid_NotHumanTurn(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	cb.SetBidPlayerIdx(1) // CPU
	err := cb.PlayerBid(5)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestBatak_CpuBid_AdvancesIndex(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	cb.SetBidPlayerIdx(1)
	cb.CpuBid()
	assert.GreaterOrEqual(t, cb.GetPlayer(1).GetBid(), domain.BatakPassBid)
	assert.Equal(t, 2, cb.GetBidPlayerIdx())
}

func TestBatak_CpuBid_SkipsHuman(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	cb.SetBidPlayerIdx(0) // Human
	cb.CpuBid()           // 何もしない
	assert.Equal(t, -1, cb.GetPlayer(0).GetBid())
	assert.Equal(t, 0, cb.GetBidPlayerIdx())
}

func TestBatak_Auction_Determinism(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()

	// 席 0 (人間): 5 をビッド
	require.NoError(t, cb.PlayerBid(5))
	assert.Equal(t, 5, cb.GetPlayer(0).GetBid())
	assert.Equal(t, 5, cb.GetHighBid())
	assert.Equal(t, 1, cb.GetBidPlayerIdx())

	// 席 1 (CPU): 手札に Spade 10, 11, 12, 13, Heart 13, Clover 13 を与えて 6 以上のビッドを出させる
	cb.GetPlayer(1).Reset()
	cb.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 10, false))
	cb.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
	cb.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 12, false))
	cb.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
	cb.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))
	cb.GetPlayer(1).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
	cb.CpuBid()
	assert.Equal(t, 6, cb.GetPlayer(1).GetBid())
	assert.Equal(t, 6, cb.GetHighBid())
	assert.Equal(t, 2, cb.GetBidPlayerIdx())

	// 席 2 (CPU): 弱い手札を与えてパス (0) させる
	cb.GetPlayer(2).Reset()
	for i := 2; i <= 9; i++ {
		cb.GetPlayer(2).AddCard(domain.NewCard(domain.CardDesignHeart, i, false))
	}
	cb.CpuBid()
	assert.Equal(t, domain.BatakPassBid, cb.GetPlayer(2).GetBid())
	assert.Equal(t, 6, cb.GetHighBid())
	assert.Equal(t, 3, cb.GetBidPlayerIdx())

	// 席 3 (CPU): 手札にスペード多数 + A/K 多数を与えて 7 以上のビッドを出させる
	cb.GetPlayer(3).Reset()
	for i := 7; i <= 13; i++ {
		cb.GetPlayer(3).AddCard(domain.NewCard(domain.CardDesignSpade, i, false))
	}
	cb.GetPlayer(3).AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))
	cb.GetPlayer(3).AddCard(domain.NewCard(domain.CardDesignDiamond, 13, false))
	cb.GetPlayer(3).AddCard(domain.NewCard(domain.CardDesignClover, 13, false))
	cb.CpuBid()
	assert.Equal(t, 8, cb.GetPlayer(3).GetBid())
	assert.Equal(t, 8, cb.GetHighBid())

	// 4 席全員が発言し、競り終了してプレイフェーズへ
	assert.Equal(t, domain.BatakPhasePlay, cb.GetPhase())
	assert.Equal(t, 3, cb.GetDeclarerIdx(), "席 3 が最高ビッド 8 で親になること")
	assert.Equal(t, 3, cb.GetLeadPlayerIdx(), "親が第 1 トリックをリードすること")
	assert.Equal(t, 3, cb.GetCurrentPlayerIdx(), "親が最初の手番であること")
}

func TestBatak_Auction_AllPass(t *testing.T) {
	for startIdx := 0; startIdx < domain.BatakPlayerCnt; startIdx++ {
		t.Run(fmt.Sprintf("bidStartIdx=%d", startIdx), func(t *testing.T) {
			cb := newTestBatak()
			cb.Reset()
			cb.SetBidStartIdx(startIdx)
			cb.SetBidPlayerIdx(startIdx)

			// 全員パスさせる (手動で全員の CPU 手札を弱くするか、Human は PlayerBid(0))
			for i := 0; i < domain.BatakPlayerCnt; i++ {
				curr := (startIdx + i) % domain.BatakPlayerCnt
				if cb.GetPlayer(curr).GetIsHuman() {
					require.NoError(t, cb.PlayerBid(domain.BatakPassBid))
				} else {
					cb.GetPlayer(curr).Reset()
					for v := 2; v <= 9; v++ {
						cb.GetPlayer(curr).AddCard(domain.NewCard(domain.CardDesignHeart, v, false))
					}
					cb.CpuBid()
				}
			}

			// 最後に発言した席 (bidStartIdx + 3) % 4 が最低額 BatakMinBid で親になる
			forcedIdx := (startIdx + 3) % domain.BatakPlayerCnt
			assert.Equal(t, domain.BatakPhasePlay, cb.GetPhase())
			assert.Equal(t, forcedIdx, cb.GetDeclarerIdx())
			assert.Equal(t, domain.BatakMinBid, cb.GetPlayer(forcedIdx).GetBid())
			assert.Equal(t, domain.BatakMinBid, cb.GetHighBid())
			assert.Equal(t, forcedIdx, cb.GetLeadPlayerIdx())
			assert.Equal(t, forcedIdx, cb.GetCurrentPlayerIdx())
		})
	}
}

func TestBatak_DeclarerLeadsFirstTrick(t *testing.T) {
	for declarerSeat := 0; declarerSeat < domain.BatakPlayerCnt; declarerSeat++ {
		t.Run(fmt.Sprintf("declarerSeat=%d", declarerSeat), func(t *testing.T) {
			cb := newTestBatak()
			cb.Reset()

			for i := 0; i < domain.BatakPlayerCnt; i++ {
				if cb.IsHumanBidTurn() {
					if declarerSeat == 0 {
						require.NoError(t, cb.PlayerBid(5))
					} else {
						require.NoError(t, cb.PlayerBid(domain.BatakPassBid))
					}
				} else {
					curr := cb.GetBidPlayerIdx()
					cb.GetPlayer(curr).Reset()
					if curr == declarerSeat {
						// 強い手札を与えて 5 以上のビッドを出させる
						for v := 10; v <= 13; v++ {
							cb.GetPlayer(curr).AddCard(domain.NewCard(domain.CardDesignSpade, v, false))
						}
						cb.GetPlayer(curr).AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
						cb.GetPlayer(curr).AddCard(domain.NewCard(domain.CardDesignHeart, 1, false))
					}
					cb.CpuBid()
				}
			}

			require.Equal(t, domain.BatakPhasePlay, cb.GetPhase())
			require.Equal(t, declarerSeat, cb.GetDeclarerIdx(), "declarerSeat should be %d", declarerSeat)
			assert.Equal(t, declarerSeat, cb.GetLeadPlayerIdx(), "declarer must lead first trick")
			assert.Equal(t, declarerSeat, cb.GetCurrentPlayerIdx(), "declarer must be current player")
		})
	}
}

func TestBatak_ScoreRound_Asymmetry(t *testing.T) {
	t.Run("親が宣言達成の卓", func(t *testing.T) {
		cb := newTestBatak()
		cb.Reset()
		cb.SetPhase(domain.BatakPhaseRoundEnd)
		cb.SetDeclarerIdx(0) // 席 0 が親

		cb.GetPlayer(0).SetBid(6) // 親は 6 宣言
		cb.GetPlayer(1).SetBid(0) // 子はパス
		cb.GetPlayer(2).SetBid(0) // 子はパス
		cb.GetPlayer(3).SetBid(0) // 子はパス

		// 席 0: 6 トリック獲得 (達成) -> +6
		for k := 0; k < 6; k++ {
			cb.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
		}
		// 席 1: 3 トリック獲得 -> +3
		for k := 0; k < 3; k++ {
			cb.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
		}
		// 席 2: 4 トリック獲得 -> +4
		for k := 0; k < 4; k++ {
			cb.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
		}
		// 席 3: 0 トリック獲得 -> +0

		cb.ScoreRound()

		// 同一卓内で親 (+bid) と子 (+tricks) を同時に検証
		assert.Equal(t, 6, cb.GetPlayer(0).GetRoundScore(), "親は達成で +bid")
		assert.Equal(t, 3, cb.GetPlayer(1).GetRoundScore(), "子はトリック数がそのまま加点")
		assert.Equal(t, 4, cb.GetPlayer(2).GetRoundScore(), "子はトリック数がそのまま加点")
		assert.Equal(t, 0, cb.GetPlayer(3).GetRoundScore(), "子はトリック数がそのまま加点")
	})

	t.Run("親が余剰達成 (オーバートリック) でもボーナスなし", func(t *testing.T) {
		cb := newTestBatak()
		cb.Reset()
		cb.SetPhase(domain.BatakPhaseRoundEnd)
		cb.SetDeclarerIdx(1) // 席 1 が親

		cb.GetPlayer(0).SetBid(0)
		cb.GetPlayer(1).SetBid(5) // 親は 5 宣言
		cb.GetPlayer(2).SetBid(0)
		cb.GetPlayer(3).SetBid(0)

		// 席 1: 7 トリック獲得 (5 宣言に対して 2 オーバートリック) -> +5
		for k := 0; k < 7; k++ {
			cb.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
		}
		for k := 0; k < 2; k++ {
			cb.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
			cb.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
			cb.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
		}

		cb.ScoreRound()

		assert.Equal(t, 2, cb.GetPlayer(0).GetRoundScore())
		assert.Equal(t, 5, cb.GetPlayer(1).GetRoundScore(), "親はオーバートリックでも +bid のみ")
		assert.Equal(t, 2, cb.GetPlayer(2).GetRoundScore())
		assert.Equal(t, 2, cb.GetPlayer(3).GetRoundScore())
	})

	t.Run("親が宣言未達の卓", func(t *testing.T) {
		cb := newTestBatak()
		cb.Reset()
		cb.SetPhase(domain.BatakPhaseRoundEnd)
		cb.SetDeclarerIdx(2) // 席 2 が親

		cb.GetPlayer(0).SetBid(0)
		cb.GetPlayer(1).SetBid(0)
		cb.GetPlayer(2).SetBid(7) // 親は 7 宣言
		cb.GetPlayer(3).SetBid(0)

		// 席 2: 5 トリックしか取れず (未達) -> -7
		for k := 0; k < 5; k++ {
			cb.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
		}
		// 席 0: 4 トリック -> +4
		for k := 0; k < 4; k++ {
			cb.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
		}
		// 席 1: 3 トリック -> +3
		for k := 0; k < 3; k++ {
			cb.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
		}
		// 席 3: 1 トリック -> +1
		cb.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})

		cb.ScoreRound()

		// 同一卓内で親 (-bid) と子 (+tricks) を同時に検証
		assert.Equal(t, 4, cb.GetPlayer(0).GetRoundScore(), "子はトリック数がそのまま加点")
		assert.Equal(t, 3, cb.GetPlayer(1).GetRoundScore(), "子はトリック数がそのまま加点")
		assert.Equal(t, -7, cb.GetPlayer(2).GetRoundScore(), "親は未達で -bid")
		assert.Equal(t, 1, cb.GetPlayer(3).GetRoundScore(), "子はトリック数がそのまま加点")
	})
}

func TestBatak_ScoreRound_WrongPhase_NoOp(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	cb.SetPhase(domain.BatakPhasePlay)
	cb.GetPlayer(0).SetBid(5)
	cb.ScoreRound()
	assert.Equal(t, 0, cb.GetPlayer(0).GetCumulativeScore())
}

func TestBatak_ScoreRound_TriggersGameEndAtMaxRounds(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	cfg := cb.GetConfig()
	cfg.MaxRounds = 2
	cb.SetConfig(cfg)

	// 1st round
	cb.SetPhase(domain.BatakPhaseRoundEnd)
	cb.SetDeclarerIdx(0)
	cb.GetPlayer(0).SetBid(5)
	for k := 0; k < 5; k++ {
		cb.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	}
	for i := 1; i < domain.BatakPlayerCnt; i++ {
		cb.GetPlayer(i).SetBid(0)
		for k := 0; k < 2; k++ {
			cb.GetPlayer(i).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
		}
	}
	cb.ScoreRound()
	assert.False(t, cb.GetGameEndFlag(), "should not end after round 1 of 2")

	cb.NextRound()
	require.Equal(t, 2, cb.GetRoundNumber())

	// 2nd / final round
	cb.SetPhase(domain.BatakPhaseRoundEnd)
	cb.SetDeclarerIdx(2)
	cb.GetPlayer(2).SetBid(6)
	for k := 0; k < 6; k++ {
		cb.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	}
	for i := 0; i < domain.BatakPlayerCnt; i++ {
		if i != 2 {
			cb.GetPlayer(i).SetBid(0)
			for k := 0; k < 2; k++ {
				cb.GetPlayer(i).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
			}
		}
	}

	cb.ScoreRound()

	assert.True(t, cb.GetGameEndFlag())
	assert.Equal(t, domain.BatakPhaseGameEnd, cb.GetPhase())
	assert.Equal(t, 2, cb.GetWinnerIdx(), "player 2 should win with highest cumulative score (2 + 6 = 8)")
}

func TestBatak_NextRound_OnlyFromRoundEnd(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	cb.SetPhase(domain.BatakPhasePlay)
	before := cb.GetRoundNumber()
	cb.NextRound()
	assert.Equal(t, before, cb.GetRoundNumber())
}

func TestBatak_NextRound_AdvancesAndDealsAgain(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	cb.SetPhase(domain.BatakPhaseRoundEnd)
	cb.NextRound()
	assert.Equal(t, 2, cb.GetRoundNumber())
	assert.Equal(t, domain.BatakPhaseBid, cb.GetPhase())
	for i := 0; i < domain.BatakPlayerCnt; i++ {
		assert.Equal(t, 13, cb.GetPlayer(i).GetCardsSize())
		assert.Equal(t, -1, cb.GetPlayer(i).GetBid())
	}
}

func TestBatak_NextTrick_OnlyFromTrickEnd(t *testing.T) {
	cb := newTestBatak()
	cb.SetPhase(domain.BatakPhasePlay)
	cb.SetTrickNumber(3)
	cb.NextTrick()
	assert.Equal(t, 3, cb.GetTrickNumber())
}

func TestBatak_NextTrick_Advances(t *testing.T) {
	cb := newTestBatak()
	cb.SetPhase(domain.BatakPhaseTrickEnd)
	cb.SetLeadPlayerIdx(2)
	cb.SetTrickNumber(3)
	cb.NextTrick()
	assert.Equal(t, 4, cb.GetTrickNumber())
	assert.Equal(t, 2, cb.GetCurrentPlayerIdx())
	assert.Equal(t, domain.BatakPhasePlay, cb.GetPhase())
}

func TestBatak_PlayerPlay_WrongPhase(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	cb.SetPhase(domain.BatakPhaseBid)
	err := cb.PlayerPlay(0)
	assert.ErrorIs(t, err, domain.ErrWrongPhase)
}

func TestBatak_PlayerPlay_NotHumanTurn(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	setupBatakPlay(cb, 1, 1, 1)
	err := cb.PlayerPlay(0)
	assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
}

func TestBatak_PlayerPlay_InvalidIndex(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	setupBatakPlay(cb, 0, 0, 1)
	err := cb.PlayerPlay(99)
	assert.Error(t, err)
}

func TestBatak_PlayerPlay_GameEnded(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()

	cfg := cb.GetConfig()
	cfg.MaxRounds = 1
	cb.SetConfig(cfg)

	cb.SetPhase(domain.BatakPhaseRoundEnd)
	for i := 0; i < domain.BatakPlayerCnt; i++ {
		cb.GetPlayer(i).SetBid(1)
		cb.GetPlayer(i).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignClover, 2, false)})
	}
	cb.ScoreRound()
	require.True(t, cb.GetGameEndFlag())

	err := cb.PlayerPlay(0)
	assert.ErrorIs(t, err, domain.ErrGameEnded)
}

func TestBatak_ResolveTrick(t *testing.T) {
	cb := newTestBatak()
	cb.SetPhase(domain.BatakPhaseTrickEnd)
	cb.SetTrickNumber(1)
	trick := []*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 11, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 4, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
	}
	cb.SetCurrentTrick(trick)
	cb.ResolveTrick()
	assert.Equal(t, 1, cb.GetPlayer(1).GetTrickCount())
	assert.Equal(t, 1, cb.GetLeadPlayerIdx())
}

func TestBatak_ResolveTrick_TrumpWins(t *testing.T) {
	cb := newTestBatak()
	cb.SetPhase(domain.BatakPhaseTrickEnd)
	cb.SetTrickNumber(1)
	trick := []*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 2, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 4, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
	}
	cb.SetCurrentTrick(trick)
	cb.ResolveTrick()
	assert.Equal(t, 1, cb.GetPlayer(1).GetTrickCount())
}

func TestBatak_ResolveTrick_AdvancesToRoundEnd(t *testing.T) {
	cb := newTestBatak()
	cb.SetPhase(domain.BatakPhaseTrickEnd)
	cb.SetTrickNumber(domain.BatakHandSize)
	cb.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 11, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 4, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 9, false)},
	})
	cb.ResolveTrick()
	assert.Equal(t, domain.BatakPhaseRoundEnd, cb.GetPhase())
}

func TestBatak_IsHumanTurn(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()

	cb.SetCurrentPlayerIdx(-1)
	assert.False(t, cb.IsHumanTurn())

	cb.SetCurrentPlayerIdx(0)
	assert.True(t, cb.IsHumanTurn())

	cb.SetCurrentPlayerIdx(1)
	assert.False(t, cb.IsHumanTurn())
}

func TestBatak_IsHumanBidTurn(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	cb.SetBidPlayerIdx(0)
	assert.True(t, cb.IsHumanBidTurn())
	cb.SetBidPlayerIdx(2)
	assert.False(t, cb.IsHumanBidTurn())
	cb.SetBidPlayerIdx(99)
	assert.False(t, cb.IsHumanBidTurn())

	// Phase が Bid でないときは bidPlayerIdx が 0 でも false
	cb.SetBidPlayerIdx(0)
	cb.SetPhase(domain.BatakPhasePlay)
	assert.False(t, cb.IsHumanBidTurn())
}

func TestBatak_GetPlayer_OutOfRange(t *testing.T) {
	cb := newTestBatak()
	assert.Nil(t, cb.GetPlayer(-1))
	assert.Nil(t, cb.GetPlayer(99))
}

func TestBatak_GetHint_BidPhase(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	hint := cb.GetHint()
	require.NotNil(t, hint)
	assert.NotNil(t, hint.Bid)
	assert.Nil(t, hint.CardIndex)
	assert.NotEmpty(t, hint.Reason)
	if *hint.Bid == domain.BatakPassBid {
		assert.Equal(t, "pass_weak_hand", hint.Reason)
	} else {
		assert.GreaterOrEqual(t, *hint.Bid, domain.BatakMinBid)
		assert.Equal(t, "strategic_bid", hint.Reason)
	}

	// 人間手番でないときは nil
	cb.SetBidPlayerIdx(1)
	assert.Nil(t, cb.GetHint())
}

func TestBatak_GetHint_PlayPhase(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	cb.SetPhase(domain.BatakPhasePlay)
	cb.SetCurrentPlayerIdx(0)
	cb.SetCurrentTrick(nil)
	cb.SetSpadesBroken(true)
	hint := cb.GetHint()
	require.NotNil(t, hint)
	assert.NotNil(t, hint.CardIndex)
}

func TestBatak_GetHint_NotPlayerTurn(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	cb.SetPhase(domain.BatakPhasePlay)
	cb.SetCurrentPlayerIdx(1)
	assert.Nil(t, cb.GetHint())
}

func TestBatak_GetValidPlayIndices_NotEmpty(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	indices := cb.GetValidPlayIndices(0)
	assert.NotEmpty(t, indices)
}

func TestBatak_GetActionLog(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	require.NoError(t, cb.PlayerBid(5))
	log := cb.GetActionLog()
	assert.NotEmpty(t, log)
}

func TestBatak_JSONRoundTrip(t *testing.T) {
	cb := newTestBatak()
	cb.Reset()
	require.NoError(t, cb.PlayerBid(5))
	cb.SetDeclarerIdx(2)
	cb.SetHighBid(7)
	cb.SetBidStartIdx(3)

	data, err := json.Marshal(cb)
	require.NoError(t, err)

	cb2 := newTestBatak()
	require.NoError(t, json.Unmarshal(data, cb2))
	assert.Equal(t, cb.GetPhase(), cb2.GetPhase())
	assert.Equal(t, cb.GetRoundNumber(), cb2.GetRoundNumber())
	assert.Equal(t, 5, cb2.GetPlayer(0).GetBid())
	assert.Equal(t, 2, cb2.GetDeclarerIdx(), "declarerIdx should roundtrip")
	assert.Equal(t, 7, cb2.GetHighBid(), "highBid should roundtrip")
	assert.Equal(t, 3, cb2.GetBidStartIdx(), "bidStartIdx should roundtrip")
}

func TestBatak_UnmarshalJSON_Invalid(t *testing.T) {
	cb := newTestBatak()
	err := json.Unmarshal([]byte("not json"), cb)
	assert.Error(t, err)
}

func TestBatak_UnmarshalJSON_Empty(t *testing.T) {
	cb := newTestBatak()
	require.NoError(t, json.Unmarshal([]byte("{}"), cb))
	assert.Empty(t, cb.GetCurrentTrick())
	assert.Empty(t, cb.GetActionLog())
}

func TestBatak_Auction_ReachabilityProbe(t *testing.T) {
	humanDeclarerCount := 0
	trials := 200

	for trial := 0; trial < trials; trial++ {
		cb := newTestBatak()
		cb.Reset()

		humanTurnVisited := false
		for cb.GetPhase() == domain.BatakPhaseBid {
			if cb.IsHumanBidTurn() {
				assert.False(t, humanTurnVisited, "人間手番は1回だけ訪れるべき")
				humanTurnVisited = true

				hint := cb.GetHint()
				require.NotNil(t, hint, "人間手番ではヒントが取得できること")
				require.NotNil(t, hint.Bid, "ビッド推奨値が存在すること")

				err := cb.PlayerBid(*hint.Bid)
				require.NoError(t, err, "ヒント推奨のビッドは有効であること: %d", *hint.Bid)
			} else {
				cb.CpuBid()
			}
		}

		assert.True(t, humanTurnVisited, "席0の人間のビッド手番が必ず1回訪れること")
		assert.Equal(t, domain.BatakPhasePlay, cb.GetPhase())

		if cb.GetDeclarerIdx() == 0 {
			humanDeclarerCount++
		}
	}

	assert.Greater(t, humanDeclarerCount, 0, "人間が親になる配りが少なくとも1回は存在すること")
	assert.Less(t, humanDeclarerCount, trials, "人間が親にならない配りが少なくとも1回は存在すること")
}

// TestBatak_Auction_Statistics は 2000 配りで親の席分布、強制親割合、人間親割合を
// 統計的に検証する恒久的テスト。人間はヒント推奨どおりに発言する。
func TestBatak_Auction_Statistics(t *testing.T) {
	cb := newTestBatak()
	trials := 2000

	declarerCounts := [domain.BatakPlayerCnt]int{}
	forcedCount := 0

	for i := 0; i < trials; i++ {
		cb.Reset()
		for cb.GetPhase() == domain.BatakPhaseBid {
			if cb.IsHumanBidTurn() {
				hint := cb.GetHint()
				require.NotNil(t, hint)
				require.NotNil(t, hint.Bid)
				require.NoError(t, cb.PlayerBid(*hint.Bid))
			} else {
				cb.CpuBid()
			}
		}
		require.Equal(t, domain.BatakPhasePlay, cb.GetPhase())
		declarerCounts[cb.GetDeclarerIdx()]++

		// ActionLog から自発的なビッド (bids) があったかを判定
		anyBid := false
		for _, log := range cb.GetActionLog() {
			if log.ActionType == "bid" && strings.Contains(log.Detail, "bids") {
				anyBid = true
				break
			}
		}
		if !anyBid {
			forcedCount++
		}
	}

	d0Pct := float64(declarerCounts[0]) / float64(trials) * 100
	d1Pct := float64(declarerCounts[1]) / float64(trials) * 100
	d2Pct := float64(declarerCounts[2]) / float64(trials) * 100
	d3Pct := float64(declarerCounts[3]) / float64(trials) * 100
	forcedPct := float64(forcedCount) / float64(trials) * 100

	t.Logf("Auction Statistics: seat0=%.1f%% seat1=%.1f%% seat2=%.1f%% seat3=%.1f%%, forced=%.1f%%",
		d0Pct, d1Pct, d2Pct, d3Pct, forcedPct)

	// 3. 親の席の分布 (どの席も 10% 〜 45%)
	assert.GreaterOrEqual(t, d0Pct, 10.0, "席 0 の親割合は 10%% 以上であること")
	assert.LessOrEqual(t, d0Pct, 45.0, "席 0 の親割合は 45%% 以下であること")
	assert.GreaterOrEqual(t, d1Pct, 10.0, "席 1 の親割合は 10%% 以上であること")
	assert.LessOrEqual(t, d1Pct, 45.0, "席 1 の親割合は 45%% 以下であること")
	assert.GreaterOrEqual(t, d2Pct, 10.0, "席 2 の親割合は 10%% 以上であること")
	assert.LessOrEqual(t, d2Pct, 45.0, "席 2 の親割合は 45%% 以下であること")
	assert.GreaterOrEqual(t, d3Pct, 10.0, "席 3 の親割合は 10%% 以上であること")
	assert.LessOrEqual(t, d3Pct, 45.0, "席 3 の親割合は 45%% 以下であること")

	// 4. 全員パスで強制的に親が決まった割合: 25% 以下
	assert.LessOrEqual(t, forcedPct, 25.0, "全員パスでの強制親割合は 25%% 以下であること")

	// 5. 人間 (席 0) が親になる割合: 10% 以上
	assert.GreaterOrEqual(t, d0Pct, 10.0, "人間 (席 0) が親になる割合は 10%% 以上であること")
}
