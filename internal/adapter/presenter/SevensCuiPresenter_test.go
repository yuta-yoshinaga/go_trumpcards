package presenter_test

import (
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

	"github.com/stretchr/testify/assert"
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
	tsp := presenter.NewSevensCuiPresenter()

	t.Run("success Output initial state", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 9, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "Sevens (7並べ)")
		assert.Contains(t, result, "[You]: 2枚")
		assert.Contains(t, result, "[0]SPADE 6")
		assert.Contains(t, result, "CPU 1: 1枚")
		assert.Contains(t, result, "ボード:")
		assert.Contains(t, result, "SPADE: 7〜7")
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
		_ = s.PlayerPlay(0) // play 6♠ → minVal[Spade] = 6

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "SPADE: 6〜7")
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
		cfg := domain.SevensConfig{TunnelEnabled: true, JokerCount: 0, CpuStrategy: false}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "ルール:")
		assert.Contains(t, result, "[トンネル]")
	})

	t.Run("success Output rule header with joker", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSevensPlayersForPresenter()
		cfg := domain.SevensConfig{TunnelEnabled: false, JokerCount: 2, CpuStrategy: false}
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
		cfg := domain.SevensConfig{TunnelEnabled: false, JokerCount: 0, CpuStrategy: true}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tsp.Output(s, nil)
		assert.Contains(t, result, "ルール:")
		assert.Contains(t, result, "[CPU戦略]")
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
