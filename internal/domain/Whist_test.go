//go:build test

package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func newTestWhist() *domain.Whist {
	players := []*domain.WhistPlayer{
		domain.NewWhistPlayer(true, 0),
		domain.NewWhistPlayer(false, 1),
		domain.NewWhistPlayer(false, 0),
		domain.NewWhistPlayer(false, 1),
	}
	return domain.NewWhist(domain.NewTrumpCards(0), players, domain.DefaultWhistConfig())
}

func setupWhistPlayPhase(w *domain.Whist, currentIdx, leadIdx, trickNum int) {
	w.SetPhase(domain.WhistPhasePlay)
	w.SetCurrentPlayerIdx(currentIdx)
	w.SetLeadPlayerIdx(leadIdx)
	w.SetTrickNumber(trickNum)
}

func TestNewWhist(t *testing.T) {
	w := newTestWhist()
	assert.Equal(t, -1, w.GetWinnerTeam())
	assert.Equal(t, 0, w.GetRoundNumber())
}

func TestWhist_Reset(t *testing.T) {
	w := newTestWhist()
	w.Reset()

	assert.Equal(t, domain.WhistPhasePlay, w.GetPhase())
	assert.Equal(t, 1, w.GetRoundNumber())
	assert.Equal(t, 1, w.GetTrickNumber())
	assert.False(t, w.GetGameEndFlag())
	assert.Equal(t, -1, w.GetWinnerTeam())
	assert.Equal(t, 0, w.GetTeamScore(0))
	assert.Equal(t, 0, w.GetTeamScore(1))

	// 全プレイヤーに13枚ずつ配られている
	for i := 0; i < 4; i++ {
		assert.Equal(t, 13, w.GetPlayer(i).GetCardsSize())
		assert.Equal(t, 0, w.GetPlayer(i).GetCumulativeScore())
	}

	// トランプスートが設定されている
	trumpSuit := w.GetTrumpSuit()
	assert.True(t, trumpSuit >= domain.CardDesignSpade && trumpSuit <= domain.CardDesignDiamond)
}

func TestWhist_Reset_ClearsAllState(t *testing.T) {
	w := newTestWhist()
	w.Reset()

	w.SetPhase(domain.WhistPhaseGameEnd)
	w.SetTeamScore(0, 10)
	w.SetTeamScore(1, 5)

	w.Reset()

	assert.Equal(t, domain.WhistPhasePlay, w.GetPhase())
	assert.Equal(t, 0, w.GetTeamScore(0))
	assert.Equal(t, 0, w.GetTeamScore(1))
}

func TestWhist_PlayerPlay(t *testing.T) {
	t.Run("valid play", func(t *testing.T) {
		w := newTestWhist()
		w.Reset()
		setupWhistPlayPhase(w, 0, 0, 1)

		err := w.PlayerPlay(0)
		assert.NoError(t, err)
		assert.Equal(t, 12, w.GetPlayer(0).GetCardsSize())
	})

	t.Run("game ended", func(t *testing.T) {
		w := newTestWhist()
		w.Reset()
		setupWhistPlayPhase(w, 0, 0, 1)
		w.SetGameEndFlag(true)

		err := w.PlayerPlay(0)
		assert.ErrorIs(t, err, domain.ErrGameEnded)
	})

	t.Run("wrong phase", func(t *testing.T) {
		w := newTestWhist()
		w.Reset()
		w.SetPhase(domain.WhistPhaseTrickEnd)

		err := w.PlayerPlay(0)
		assert.ErrorIs(t, err, domain.ErrWrongPhase)
	})

	t.Run("not human turn", func(t *testing.T) {
		w := newTestWhist()
		w.Reset()
		setupWhistPlayPhase(w, 1, 0, 1) // CPU at index 1

		err := w.PlayerPlay(0)
		assert.ErrorIs(t, err, domain.ErrNotHumanTurn)
	})

	t.Run("invalid card index", func(t *testing.T) {
		w := newTestWhist()
		w.Reset()
		setupWhistPlayPhase(w, 0, 0, 1)

		err := w.PlayerPlay(99)
		assert.Error(t, err)
	})

	t.Run("must follow suit", func(t *testing.T) {
		w := newTestWhist()
		w.Reset()
		setupWhistPlayPhase(w, 0, 1, 1)

		// リードカードを設定
		leadCard := domain.NewCard(domain.CardDesignHeart, 10, false)
		w.SetCurrentTrick([]*domain.TrickCard{
			{PlayerIdx: 1, Card: leadCard},
		})

		// プレイヤーの手札を設定（ハートを持っている）
		p := w.GetPlayer(0)
		p.Reset()
		p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		p.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))

		// スペード（index 1）を出そうとする → フォロースート違反
		err := w.PlayerPlay(1)
		assert.Error(t, err)

		// ハート（index 0）を出す → OK
		err = w.PlayerPlay(0)
		assert.NoError(t, err)
	})
}

func TestWhist_CpuPlay(t *testing.T) {
	w := newTestWhist()
	w.Reset()
	setupWhistPlayPhase(w, 1, 1, 1) // CPU at index 1

	initialCards := w.GetPlayer(1).GetCardsSize()
	w.CpuPlay()

	assert.Equal(t, initialCards-1, w.GetPlayer(1).GetCardsSize())
}

func TestWhist_CpuPlay_SkipsWhenHuman(t *testing.T) {
	w := newTestWhist()
	w.Reset()
	setupWhistPlayPhase(w, 0, 0, 1)

	initialCards := w.GetPlayer(0).GetCardsSize()
	w.CpuPlay()

	assert.Equal(t, initialCards, w.GetPlayer(0).GetCardsSize())
}

func TestWhist_ResolveTrick(t *testing.T) {
	w := newTestWhist()
	w.Reset()
	w.SetTrumpSuit(domain.CardDesignSpade)
	w.SetTrickNumber(1)
	w.SetPhase(domain.WhistPhaseTrickEnd)

	trick := []*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 8, false)},
	}
	w.SetCurrentTrick(trick)

	w.ResolveTrick()

	// Player 2 (K of hearts) wins
	assert.Equal(t, 2, w.GetLeadPlayerIdx())
	assert.Equal(t, 1, w.GetPlayer(2).GetTrickCount())
}

func TestWhist_ResolveTrick_TrumpWins(t *testing.T) {
	w := newTestWhist()
	w.Reset()
	w.SetTrumpSuit(domain.CardDesignSpade)
	w.SetTrickNumber(1)
	w.SetPhase(domain.WhistPhaseTrickEnd)

	trick := []*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignSpade, 2, false)}, // trump
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 8, false)},
	}
	w.SetCurrentTrick(trick)

	w.ResolveTrick()

	// Player 1 (2 of spades/trump) wins over K of hearts
	assert.Equal(t, 1, w.GetLeadPlayerIdx())
	assert.Equal(t, 1, w.GetPlayer(1).GetTrickCount())
}

func TestWhist_ResolveTrick_LastTrick(t *testing.T) {
	w := newTestWhist()
	w.Reset()
	w.SetTrumpSuit(domain.CardDesignSpade)
	w.SetTrickNumber(13) // last trick
	w.SetPhase(domain.WhistPhaseTrickEnd)

	trick := []*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
		{PlayerIdx: 3, Card: domain.NewCard(domain.CardDesignHeart, 8, false)},
	}
	w.SetCurrentTrick(trick)

	w.ResolveTrick()

	assert.Equal(t, domain.WhistPhaseRoundEnd, w.GetPhase())
}

func TestWhist_NextTrick(t *testing.T) {
	w := newTestWhist()
	w.Reset()
	w.SetPhase(domain.WhistPhaseTrickEnd)
	w.SetLeadPlayerIdx(2)
	w.SetTrickNumber(1)

	w.NextTrick()

	assert.Equal(t, domain.WhistPhasePlay, w.GetPhase())
	assert.Equal(t, 2, w.GetCurrentPlayerIdx())
	assert.Equal(t, 2, w.GetTrickNumber())
	assert.Nil(t, w.GetCurrentTrick())
}

func TestWhist_ScoreRound(t *testing.T) {
	w := newTestWhist()
	w.Reset()
	w.SetPhase(domain.WhistPhaseRoundEnd)

	// チーム0 (players 0,2) が8トリック、チーム1 (players 1,3) が5トリック
	w.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 1, false)})
	w.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 2, false)})
	w.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 3, false)})
	w.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 4, false)})
	w.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 5, false)})
	w.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 6, false)})
	w.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 7, false)})
	w.GetPlayer(2).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, 8, false)})
	w.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
	w.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 2, false)})
	w.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
	w.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 4, false)})
	w.GetPlayer(3).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, false)})

	w.ScoreRound()

	// Team 0: 8 - 6 = 2 points
	assert.Equal(t, 2, w.GetTeamScore(0))
	// Team 1: 5 - 6 = 0 points (below threshold)
	assert.Equal(t, 0, w.GetTeamScore(1))
}

func TestWhist_ScoreRound_GameEnd(t *testing.T) {
	w := newTestWhist()
	w.Reset()
	w.SetPhase(domain.WhistPhaseRoundEnd)
	w.SetTeamScore(0, 4) // 1 more point to win

	// Give team 0 enough tricks to win (7 = 1 point over book)
	for i := 0; i < 7; i++ {
		w.GetPlayer(0).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignHeart, i+1, false)})
	}
	for i := 0; i < 6; i++ {
		w.GetPlayer(1).AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, i+1, false)})
	}

	w.ScoreRound()

	assert.True(t, w.GetGameEndFlag())
	assert.Equal(t, domain.WhistPhaseGameEnd, w.GetPhase())
	assert.Equal(t, 0, w.GetWinnerTeam())
	assert.Equal(t, 5, w.GetTeamScore(0))
}

func TestWhist_NextRound(t *testing.T) {
	w := newTestWhist()
	w.Reset()
	w.SetPhase(domain.WhistPhaseRoundEnd)
	w.SetRoundNumber(1)
	initialDealer := w.GetDealerIdx()

	w.NextRound()

	assert.Equal(t, domain.WhistPhasePlay, w.GetPhase())
	assert.Equal(t, 2, w.GetRoundNumber())
	assert.Equal(t, 1, w.GetTrickNumber())
	assert.Equal(t, (initialDealer+1)%4, w.GetDealerIdx())

	for i := 0; i < 4; i++ {
		assert.Equal(t, 13, w.GetPlayer(i).GetCardsSize())
	}
}

func TestWhist_NextRound_WrongPhase(t *testing.T) {
	w := newTestWhist()
	w.Reset()
	w.SetPhase(domain.WhistPhasePlay)

	w.NextRound()

	assert.Equal(t, domain.WhistPhasePlay, w.GetPhase())
}

func TestWhist_IsHumanTurn(t *testing.T) {
	w := newTestWhist()
	w.Reset()

	w.SetCurrentPlayerIdx(0)
	assert.True(t, w.IsHumanTurn())

	w.SetCurrentPlayerIdx(1)
	assert.False(t, w.IsHumanTurn())

	w.SetCurrentPlayerIdx(-1)
	assert.False(t, w.IsHumanTurn())
}

func TestWhist_GetPlayer_OutOfBounds(t *testing.T) {
	w := newTestWhist()
	assert.Nil(t, w.GetPlayer(-1))
	assert.Nil(t, w.GetPlayer(99))
}

func TestWhist_GetTeamScore_OutOfBounds(t *testing.T) {
	w := newTestWhist()
	assert.Equal(t, 0, w.GetTeamScore(-1))
	assert.Equal(t, 0, w.GetTeamScore(99))
}

func TestWhist_GetHint(t *testing.T) {
	w := newTestWhist()
	w.Reset()
	setupWhistPlayPhase(w, 0, 0, 1)

	hint := w.GetHint()
	assert.NotNil(t, hint)
	assert.NotNil(t, hint.CardIndex)
	assert.NotEmpty(t, hint.Reason)
}

func TestWhist_GetHint_NilWhenNotHumanTurn(t *testing.T) {
	w := newTestWhist()
	w.Reset()
	setupWhistPlayPhase(w, 1, 1, 1)

	hint := w.GetHint()
	assert.Nil(t, hint)
}

func TestWhist_GetConfig_SetConfig(t *testing.T) {
	w := newTestWhist()
	cfg := domain.WhistConfig{CpuDifficulty: domain.WhistCpuDifficultyHard, PointLimit: 10}
	w.SetConfig(cfg)
	assert.Equal(t, cfg, w.GetConfig())
}

func TestWhist_MarshalUnmarshalJSON(t *testing.T) {
	w := newTestWhist()
	w.Reset()

	data, err := json.Marshal(w)
	require.NoError(t, err)

	var restored domain.Whist
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, w.GetPhase(), restored.GetPhase())
	assert.Equal(t, w.GetRoundNumber(), restored.GetRoundNumber())
	assert.Equal(t, w.GetTrickNumber(), restored.GetTrickNumber())
	assert.Equal(t, w.GetTrumpSuit(), restored.GetTrumpSuit())
	assert.Equal(t, w.GetDealerIdx(), restored.GetDealerIdx())
	assert.Equal(t, w.GetTeamScore(0), restored.GetTeamScore(0))
	assert.Equal(t, w.GetTeamScore(1), restored.GetTeamScore(1))
}

func TestWhist_CpuDifficulties(t *testing.T) {
	difficulties := []domain.WhistCpuDifficulty{
		domain.WhistCpuDifficultyEasy,
		domain.WhistCpuDifficultyNormal,
		domain.WhistCpuDifficultyHard,
	}
	for _, diff := range difficulties {
		t.Run(difficultyName(diff), func(t *testing.T) {
			w := newTestWhist()
			cfg := domain.DefaultWhistConfig()
			cfg.CpuDifficulty = diff
			w.SetConfig(cfg)
			w.Reset()
			setupWhistPlayPhase(w, 1, 1, 1)

			initialCards := w.GetPlayer(1).GetCardsSize()
			w.CpuPlay()
			assert.Equal(t, initialCards-1, w.GetPlayer(1).GetCardsSize())
		})
	}
}

func difficultyName(d domain.WhistCpuDifficulty) string {
	switch d {
	case domain.WhistCpuDifficultyEasy:
		return "Easy"
	case domain.WhistCpuDifficultyNormal:
		return "Normal"
	case domain.WhistCpuDifficultyHard:
		return "Hard"
	default:
		return "Unknown"
	}
}

func TestWhist_FullRound(t *testing.T) {
	// 1ラウンド通して全トリックを完了できることの確認
	w := newTestWhist()
	w.Reset()

	for trick := 0; trick < 13; trick++ {
		for player := 0; player < 4; player++ {
			if w.GetPhase() != domain.WhistPhasePlay {
				break
			}
			if w.GetPlayer(w.GetCurrentPlayerIdx()).GetIsHuman() {
				err := w.PlayerPlay(0)
				if err != nil {
					// フォロースート違反の場合、有効なカードを探す
					p := w.GetPlayer(0)
					for ci := 0; ci < p.GetCardsSize(); ci++ {
						err = w.PlayerPlay(ci)
						if err == nil {
							break
						}
					}
				}
			} else {
				w.CpuPlay()
			}
		}
		if w.GetPhase() == domain.WhistPhaseTrickEnd {
			w.ResolveTrick()
			if w.GetPhase() == domain.WhistPhaseRoundEnd {
				break
			}
			w.NextTrick()
		}
	}

	assert.Equal(t, domain.WhistPhaseRoundEnd, w.GetPhase())

	// 全13トリックが分配されている
	totalTricks := 0
	for i := 0; i < 4; i++ {
		totalTricks += w.GetPlayer(i).GetTrickCount()
	}
	assert.Equal(t, 13, totalTricks)
}

func TestWhist_GetActionLog(t *testing.T) {
	w := newTestWhist()
	w.Reset()

	log := w.GetActionLog()
	assert.NotEmpty(t, log)
}

func TestWhistConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := domain.DefaultWhistConfig()
		assert.NoError(t, cfg.Validate())
	})

	t.Run("invalid difficulty", func(t *testing.T) {
		cfg := domain.WhistConfig{CpuDifficulty: 99, PointLimit: 5}
		assert.Error(t, cfg.Validate())
	})

	t.Run("invalid point limit", func(t *testing.T) {
		cfg := domain.WhistConfig{CpuDifficulty: domain.WhistCpuDifficultyNormal, PointLimit: 0}
		assert.Error(t, cfg.Validate())
	})
}

func TestWhistPlayer_Team(t *testing.T) {
	p := domain.NewWhistPlayer(true, 1)
	assert.Equal(t, 1, p.GetTeam())
	assert.True(t, p.GetIsHuman())
}

func TestWhistPlayer_ResetRound(t *testing.T) {
	p := domain.NewWhistPlayer(false, 0)
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})
	p.SetRoundScore(10)

	p.ResetRound()

	assert.Equal(t, 0, p.GetCardsSize())
	assert.Equal(t, 0, p.GetTrickCount())
	assert.Equal(t, 0, p.GetRoundScore())
}

func TestWhistPlayer_MarshalUnmarshalJSON(t *testing.T) {
	p := domain.NewWhistPlayer(true, 1)
	p.AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
	p.AddTrick([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 1, false)})

	data, err := json.Marshal(p)
	require.NoError(t, err)

	var restored domain.WhistPlayer
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, 1, restored.GetTeam())
	assert.Equal(t, 1, restored.GetCardsSize())
	assert.Equal(t, 1, restored.GetTrickCount())
}

func TestWhist_GetValidPlayIndices(t *testing.T) {
	w := newTestWhist()
	w.Reset()
	setupWhistPlayPhase(w, 0, 0, 1)

	indices := w.GetValidPlayIndices(0)
	assert.Equal(t, 13, len(indices)) // All cards playable on lead
}

func TestWhist_FullRound_Hard(t *testing.T) {
	// Hard difficulty でフルラウンドを複数回実行してCPU AIの全分岐をカバー
	for iter := 0; iter < 50; iter++ {
		w := newTestWhist()
		cfg := domain.DefaultWhistConfig()
		cfg.CpuDifficulty = domain.WhistCpuDifficultyHard
		w.SetConfig(cfg)
		w.Reset()

		for trick := 0; trick < 13; trick++ {
			for player := 0; player < 4; player++ {
				if w.GetPhase() != domain.WhistPhasePlay {
					break
				}
				if w.GetPlayer(w.GetCurrentPlayerIdx()).GetIsHuman() {
					for ci := 0; ci < w.GetPlayer(0).GetCardsSize(); ci++ {
						if w.PlayerPlay(ci) == nil {
							break
						}
					}
				} else {
					w.CpuPlay()
				}
			}
			if w.GetPhase() == domain.WhistPhaseTrickEnd {
				w.ResolveTrick()
				if w.GetPhase() == domain.WhistPhaseRoundEnd {
					break
				}
				w.NextTrick()
			}
		}

		assert.Equal(t, domain.WhistPhaseRoundEnd, w.GetPhase())
	}
}

func TestWhist_CpuPlayHard_FollowWithPartnerWinning(t *testing.T) {
	w := newTestWhist()
	cfg := domain.DefaultWhistConfig()
	cfg.CpuDifficulty = domain.WhistCpuDifficultyHard
	w.SetConfig(cfg)
	w.Reset()
	w.SetTrumpSuit(domain.CardDesignSpade)
	setupWhistPlayPhase(w, 2, 1, 2)

	// Player 1 leads heart 5, player 2 (team 0, partner of 0) has hearts
	// Player 2's partner (player 0) is not in the trick yet
	w.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
	})

	// Give player 2 some hearts and other suits
	p2 := w.GetPlayer(2)
	p2.Reset()
	p2.AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))
	p2.AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
	p2.AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))

	w.CpuPlay()
	assert.Equal(t, 2, p2.GetCardsSize())
}

func TestWhist_CpuPlayHard_VoidTrumpCut(t *testing.T) {
	w := newTestWhist()
	cfg := domain.DefaultWhistConfig()
	cfg.CpuDifficulty = domain.WhistCpuDifficultyHard
	w.SetConfig(cfg)
	w.Reset()
	w.SetTrumpSuit(domain.CardDesignSpade)
	setupWhistPlayPhase(w, 1, 0, 2)

	// Player 0 leads heart, player 1 has no hearts but has trump
	w.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
	})

	p1 := w.GetPlayer(1)
	p1.Reset()
	p1.AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
	p1.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))

	w.CpuPlay()
	assert.Equal(t, 1, p1.GetCardsSize())
}

func TestWhist_CpuPlayHard_PartnerWinningDiscard(t *testing.T) {
	w := newTestWhist()
	cfg := domain.DefaultWhistConfig()
	cfg.CpuDifficulty = domain.WhistCpuDifficultyHard
	w.SetConfig(cfg)
	w.Reset()
	w.SetTrumpSuit(domain.CardDesignSpade)
	setupWhistPlayPhase(w, 3, 1, 2)

	// Player 1 (team 1) leads heart 5, player 2 (team 0) plays heart K (winning)
	// Player 3 (team 1) should see partner is NOT winning (team 0 is winning)
	w.SetCurrentTrick([]*domain.TrickCard{
		{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 5, false)},
		{PlayerIdx: 2, Card: domain.NewCard(domain.CardDesignHeart, 13, false)},
	})

	p3 := w.GetPlayer(3)
	p3.Reset()
	p3.AddCard(domain.NewCard(domain.CardDesignDiamond, 3, false))
	p3.AddCard(domain.NewCard(domain.CardDesignClover, 8, false))

	w.CpuPlay()
	assert.Equal(t, 1, p3.GetCardsSize())
}

func TestWhist_CpuPlayHard_LeadLow(t *testing.T) {
	// Hard CPU lead when partner doesn't need tricks — leads strategically
	for iter := 0; iter < 20; iter++ {
		w := newTestWhist()
		cfg := domain.DefaultWhistConfig()
		cfg.CpuDifficulty = domain.WhistCpuDifficultyHard
		w.SetConfig(cfg)
		w.Reset()
		setupWhistPlayPhase(w, 1, 1, 5) // CPU leads mid-game

		initialCards := w.GetPlayer(1).GetCardsSize()
		w.CpuPlay()
		assert.Equal(t, initialCards-1, w.GetPlayer(1).GetCardsSize())
	}
}
