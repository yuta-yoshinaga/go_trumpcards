package presenter_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func makeSevensPlayersForPresenter() []*domain.SevensPlayer {
	return []*domain.SevensPlayer{
		domain.NewSevensPlayer(true),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
		domain.NewSevensPlayer(false),
	}
}

// setupSevensCuiTest creates a Sevens game with standard setup (player[0] SPADE 6, players[1-3] HEART 2).
func setupSevensCuiTest() (*domain.Sevens, []*domain.SevensPlayer) {
	tc := domain.NewTrumpCards(0)
	players := makeSevensPlayersForPresenter()
	s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
	players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
	players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
	return s, players
}

func TestSevensCuiPresenter_Method(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	tsp := new(presenter.SevensCuiPresenter)

	t.Run("success Output initial state", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 9, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "Sevens (7並べ)")
		assert.Contains(t, result, "あなた: 2枚")
		assert.Contains(t, result, "[0]SPADE 6")
		assert.Contains(t, result, "CPU 1: 1枚")
		assert.Contains(t, result, "ボード:")
		// Only the seed 7 is on the board initially (positions 1..13 shown).
		assert.Contains(t, result, "SPADE: _ _ _ _ _ _ 7 _ _ _ _ _ _")
		assert.Contains(t, result, "手番: あなた")
	})

	t.Run("success Output shows pass count", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // not playable
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		_ = s.PlayerPlay(-1) // human passes

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "パス: 1/5")
		assert.Contains(t, result, "パスしました")
	})

	t.Run("success Output shows board state after play", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = s.PlayerPlay(0) // play 6♠ → 6 and 7 now placed on spades

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "SPADE: _ _ _ _ _ 6 7 _ _ _ _ _ _")
	})

	t.Run("success Output game ended", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		_ = s.PlayerPlay(0)

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "ゲーム終了")
	})

	t.Run("success Output shows CPU actions", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // not playable → pass
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()         // CPU 1 passes

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "[CPUの行動]")
		assert.Contains(t, result, "パスしました")
	})

	t.Run("success Output finished player shows rank", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "上がり/失格 (ランク: 1位)")
	})

	t.Run("success Output joker card in hand", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "[0]JOKER")
		assert.Contains(t, result, "[1]SPADE 6")
	})

	t.Run("success Output rule header with tunnel", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		cfg := domain.SevensConfig{TunnelEnabled: true, JokerCount: 0, CpuStrategy: domain.SevensCpuSimple}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "ルール:")
		assert.Contains(t, result, "[トンネル]")
	})

	t.Run("success Output rule header with tunnel skip width", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		cfg := domain.SevensConfig{TunnelSkipWidth: 3}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "ルール:")
		assert.Contains(t, result, "[トンネルスキップ3]")
	})

	t.Run("success Output rule header with joker", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		cfg := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: domain.SevensCpuSimple}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "ルール:")
		assert.Contains(t, result, "[ジョーカー×2]")
		assert.Contains(t, result, "j [カードインデックス]")
	})

	t.Run("success Output rule header with CPU strategy", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		cfg := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: domain.SevensCpuStrategic}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "ルール:")
		assert.Contains(t, result, "[CPU戦略]")
	})

	t.Run("success Output rule header with CPU harassment", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		cfg := domain.SevensConfig{CpuStrategy: domain.SevensCpuHarassment}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "ルール:")
		assert.Contains(t, result, "[嫌がらせ特化]")
		assert.NotContains(t, result, "[CPU戦略]")
	})

	t.Run("success Output no rule header with default config", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.NotContains(t, result, "ルール:")
	})

	t.Run("success Output joker play action with target", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = s.PlayerPlayJoker(0, domain.CardDesignSpade, 6) // joker → SPADE 6

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "JOKER")
		assert.Contains(t, result, "SPADE 6")
		assert.Contains(t, result, "を出しました")
	})

	t.Run("success Output shows error message", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, domain.ErrInvalidPlay)
		assert.Contains(t, result, domain.ErrInvalidPlay.Error())
	})

	t.Run("success Output getCardStr nil and unknown design", func(t *testing.T) {
		s, players := setupSevensCuiTest()
		players[0].AddCard(nil)
		players[0].AddCard(domain.NewCard(99, 1, false))
		result := tsp.Output(s, nil)
		assert.Contains(t, result, "??")
		assert.Contains(t, result, "UNKNOWN")
	})

	t.Run("success Output getCardStr all designs", func(t *testing.T) {
		s, players := setupSevensCuiTest()
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignDiamond, 6, false))
		result := tsp.Output(s, nil)
		assert.Contains(t, result, "CLOVER 6")
		assert.Contains(t, result, "HEART 8")
		assert.Contains(t, result, "DIAMOND 6")
	})

	t.Run("success Output getSuitName all suits via joker play", func(t *testing.T) {
		suitTests := []struct {
			suit     int
			expected string
		}{
			{domain.CardDesignSpade, "SPADE"},
			{domain.CardDesignClover, "CLOVER"},
			{domain.CardDesignHeart, "HEART"},
			{domain.CardDesignDiamond, "DIAMOND"},
		}
		for _, st := range suitTests {
			s, _ := setupSevensCuiTest()
			s.SetHumanAction(&domain.SevensCpuAction{
				PlayerIdx:   0,
				PlayedCard:  domain.NewCard(domain.CardDesignJoker, 0, false),
				TargetSuit:  st.suit,
				TargetValue: 6,
			})
			result := tsp.Output(s, nil)
			assert.Contains(t, result, st.expected)
			assert.Contains(t, result, "を出しました")
		}
	})

	t.Run("success Output getSuitName default case", func(t *testing.T) {
		s, _ := setupSevensCuiTest()
		s.SetHumanAction(&domain.SevensCpuAction{
			PlayerIdx:   0,
			PlayedCard:  domain.NewCard(domain.CardDesignJoker, 0, false),
			TargetSuit:  999,
			TargetValue: 6,
		})
		result := tsp.Output(s, nil)
		assert.Contains(t, result, "UNKNOWN")
	})

	t.Run("success Output getPlayerName nil player", func(t *testing.T) {
		s, _ := setupSevensCuiTest()
		// Set human action with out-of-bounds player idx
		s.SetHumanAction(&domain.SevensCpuAction{
			PlayerIdx:  99,
			PlayedCard: nil,
		})
		result := tsp.Output(s, nil)
		assert.Contains(t, result, "UNKNOWN")
	})

	t.Run("success Output human action pass", func(t *testing.T) {
		s, players := setupSevensCuiTest()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		s.SetHumanAction(&domain.SevensCpuAction{
			PlayerIdx:  0,
			PlayedCard: nil,
		})
		result := tsp.Output(s, nil)
		assert.Contains(t, result, "パスしました")
	})

	t.Run("success Output human action non-joker play", func(t *testing.T) {
		s, players := setupSevensCuiTest()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		s.SetHumanAction(&domain.SevensCpuAction{
			PlayerIdx:  0,
			PlayedCard: domain.NewCard(domain.CardDesignSpade, 8, false),
		})
		result := tsp.Output(s, nil)
		assert.Contains(t, result, "SPADE 8")
		assert.Contains(t, result, "を出しました")
	})

	t.Run("success Output CPU action with joker and target", func(t *testing.T) {
		s, _ := setupSevensCuiTest()
		s.SetCpuActions([]*domain.SevensCpuAction{
			{
				PlayerIdx:   1,
				PlayedCard:  domain.NewCard(domain.CardDesignJoker, 0, false),
				TargetSuit:  domain.CardDesignSpade,
				TargetValue: 8,
			},
		})
		result := tsp.Output(s, nil)
		assert.Contains(t, result, "[CPUの行動]")
		assert.Contains(t, result, "JOKER")
		assert.Contains(t, result, "SPADE 8")
		assert.Contains(t, result, "を出しました")
	})

	t.Run("success Output CPU action pass", func(t *testing.T) {
		s, _ := setupSevensCuiTest()
		s.SetCpuActions([]*domain.SevensCpuAction{
			{
				PlayerIdx:  1,
				PlayedCard: nil,
			},
		})
		result := tsp.Output(s, nil)
		assert.Contains(t, result, "[CPUの行動]")
		assert.Contains(t, result, "パスしました")
	})

	t.Run("success Output CPU action non-joker play", func(t *testing.T) {
		s, _ := setupSevensCuiTest()
		s.SetCpuActions([]*domain.SevensCpuAction{
			{
				PlayerIdx:  1,
				PlayedCard: domain.NewCard(domain.CardDesignHeart, 8, false),
			},
		})
		result := tsp.Output(s, nil)
		assert.Contains(t, result, "[CPUの行動]")
		assert.Contains(t, result, "HEART 8")
		assert.Contains(t, result, "を出しました")
	})

	t.Run("success Output unlimited pass display", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		cfg := domain.SevensConfig{MaxPasses: 0}
		s := domain.NewSevens(tc, players, cfg)
		for _, p := range players {
			p.SetMaxPasses(0)
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "パス: 0/∞")
	})

	t.Run("success Output custom pass count display", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		cfg := domain.SevensConfig{MaxPasses: 3}
		s := domain.NewSevens(tc, players, cfg)
		for _, p := range players {
			p.SetMaxPasses(3)
		}
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "パス: 0/3")
	})

	t.Run("success Output rule header with unlimited passes", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		cfg := domain.SevensConfig{MaxPasses: 0}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "ルール:")
		assert.Contains(t, result, "[パス無制限]")
	})

	t.Run("success Output rule header with custom passes", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		cfg := domain.SevensConfig{MaxPasses: 3}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "ルール:")
		assert.Contains(t, result, "[パス3回]")
	})

	t.Run("success Output forced pass annotation on human action", func(t *testing.T) {
		s, _ := setupSevensCuiTest()
		s.SetHumanAction(&domain.SevensCpuAction{
			PlayerIdx:  0,
			PlayedCard: nil,
			ForcedPass: true,
		})
		result := tsp.Output(s, nil)
		assert.Contains(t, result, "パスしました (出せるカードなし)")
	})

	t.Run("success Output forced pass annotation on CPU action", func(t *testing.T) {
		s, _ := setupSevensCuiTest()
		s.SetCpuActions([]*domain.SevensCpuAction{
			{
				PlayerIdx:  1,
				PlayedCard: nil,
				ForcedPass: true,
			},
		})
		result := tsp.Output(s, nil)
		assert.Contains(t, result, "[CPUの行動]")
		assert.Contains(t, result, "パスしました (出せるカードなし)")
	})

	t.Run("success Output rule header with NoJokerFinish", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		cfg := domain.SevensConfig{NoJokerFinish: true, JokerCount: 1, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "ルール:")
		assert.Contains(t, result, "[ジョーカー上がり禁止]")
	})

	t.Run("success Output rule header with JokerReclaim", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		cfg := domain.SevensConfig{JokerReclaimEnabled: true, JokerCount: 1, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "ルール:")
		assert.Contains(t, result, "[ジョーカー回収]")
	})

	t.Run("success Output rule header without JokerReclaim when disabled", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.NotContains(t, result, "[ジョーカー回収]")
	})

	t.Run("success Output rule header with EndStop", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		cfg := domain.SevensConfig{EndStopEnabled: true, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "ルール:")
		assert.Contains(t, result, "[片側ストップ]")
	})

	t.Run("success Output rule header without EndStop when disabled", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.NotContains(t, result, "[片側ストップ]")
	})

	t.Run("success Output rule header with JokerConsecutiveBanned", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		cfg := domain.SevensConfig{JokerConsecutiveBanned: true, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "ルール:")
		assert.Contains(t, result, "[ジョーカー連続禁止]")
	})

	t.Run("success Output rule header without JokerConsecutiveBanned when disabled", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.NotContains(t, result, "[ジョーカー連続禁止]")
	})

	t.Run("success Output non-forced pass does not show forced annotation", func(t *testing.T) {
		s, _ := setupSevensCuiTest()
		s.SetHumanAction(&domain.SevensCpuAction{
			PlayerIdx:  0,
			PlayedCard: nil,
			ForcedPass: false,
		})
		result := tsp.Output(s, nil)
		assert.Contains(t, result, "パスしました")
		assert.NotContains(t, result, "(出せるカードなし)")
	})
}

func TestSevensCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.SevensCuiPresenter)

	t.Run("with entries", func(t *testing.T) {
		mockGame := new(interfaces.MockSevensGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "played 7 of hearts"},
		}
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(entries)
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		mockGame.On("GetPlayer", mock.Anything).Return(domain.NewSevensPlayer(true)).Maybe()

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "play")
		assert.Contains(t, result, "played 7 of hearts")
		mockGame.AssertExpectations(t)
	})

	t.Run("nil_entries", func(t *testing.T) {
		mockGame := new(interfaces.MockSevensGame)
		mockGame.On("GetGameEndFlag").Return(true)
		mockGame.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
		mockGame.On("GetPlayer", mock.Anything).Return(domain.NewSevensPlayer(true)).Maybe()

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})

	t.Run("game_not_ended", func(t *testing.T) {
		mockGame := new(interfaces.MockSevensGame)
		mockGame.On("GetGameEndFlag").Return(false)

		result := p.ActionLogOutput(mockGame)

		assert.Contains(t, result, "棋譜はありません")
		mockGame.AssertExpectations(t)
	})
}

// TestSevensCuiPresenter_English verifies issue #1699 Phase 2: every
// previously-hardcoded Japanese string in SevensCuiPresenter now follows
// the active locale. The default ja path is already exercised by the
// table-driven tests above; this suite re-runs the same setup under
// LANG=en and checks that the English keys win out.
func TestSevensCuiPresenter_English(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	tsp := new(presenter.SevensCuiPresenter)

	t.Run("initial state uses English labels", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 9, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "Sevens")
		// cui_common labels (Phase 1) flow through the helpers.
		assert.Contains(t, result, "You: 1 cards")
		assert.Contains(t, result, "CPU 1: 1 cards")
		assert.Contains(t, result, "Turn: You")
		// Sevens-scoped labels migrated in this PR.
		assert.Contains(t, result, "Board:")
		assert.Contains(t, result, "p [index] to play a card")
	})

	t.Run("CPU pass action renders in English", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // not playable → forced pass
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		_ = s.PlayerPlay(0)
		s.CpuPlay()

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "[CPU actions]")
		// The forced-pass branch ("no playable card") is what the fixture
		// triggers — pin the assertion to that key so a future regression
		// that produces an unrelated 'passed' substring fails CI.
		assert.Contains(t, result, "passed (no playable card)")
	})

	t.Run("game-end summary uses rank-line keys", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		_ = s.PlayerPlay(0)

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "Game over!")
		assert.Contains(t, result, "rank 1")
	})

	t.Run("non-default rules render English badges", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		cfg := domain.DefaultSevensConfig()
		cfg.JokerCount = 2
		cfg.NoJokerFinish = true
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "Rules:")
		assert.Contains(t, result, "[joker x2]")
		assert.Contains(t, result, "[no joker finish]")
	})
}

// #5479: CUI の手札は出せる札とそうでない札を区別していなかった。番号を入れて
// 弾かれることでしか学べず、Web は SevensHumanArea.tsx が色付きで示している。
func TestSevensCuiPresenter_PlayableMarks(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	tsp := new(presenter.SevensCuiPresenter)

	newGame := func(hand ...*domain.Card) *domain.Sevens {
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(domain.NewTrumpCards(0), players, domain.DefaultSevensConfig())
		for _, c := range hand {
			players[0].AddCard(c)
		}
		for i := 1; i < 4; i++ {
			players[i].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		}
		return s
	}
	card := func(design, value int) *domain.Card { return domain.NewCard(design, value, false) }

	t.Run("stars the playable card and leaves the others bare", func(t *testing.T) {
		s := newGame(card(domain.CardDesignSpade, 6), card(domain.CardDesignSpade, 11))
		out := tsp.Output(s, nil)
		assert.Contains(t, out, "[0]SPADE 6"+presenter.CuiLegalMark)
		assert.NotContains(t, out, "[1]SPADE 11"+presenter.CuiLegalMark)
	})

	// **1枚も出せないときに無印にすると「判定していない」と区別が付かない。**
	// 7並べではこれが普通に起きる局面なので、明示的に言う。
	t.Run("says so when nothing is playable, rather than showing a bare hand", func(t *testing.T) {
		s := newGame(card(domain.CardDesignSpade, 11), card(domain.CardDesignSpade, 12))
		out := tsp.Output(s, nil)
		assert.Contains(t, out, i18n.T("sevens.noPlayable"))
		assert.NotContains(t, out, presenter.CuiLegalMark)
	})

	// 出せる札があるときに「出せる札がありません」を出さない (負のコントロール)。
	t.Run("does not claim a dead hand when a card is playable", func(t *testing.T) {
		s := newGame(card(domain.CardDesignSpade, 6))
		assert.NotContains(t, tsp.Output(s, nil), i18n.T("sevens.noPlayable"))
	})
}
