package presenter_test

import (
	"encoding/json"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

	"github.com/stretchr/testify/assert"
)

func TestSevensWebPresenter_Method(t *testing.T) {
	tswp := presenter.NewSevensWebPresenter()

	makeSPlayers := func() []*domain.SevensPlayer {
		return []*domain.SevensPlayer{
			domain.NewSevensPlayer(true),
			domain.NewSevensPlayer(false),
			domain.NewSevensPlayer(false),
			domain.NewSevensPlayer(false),
		}
	}

	setupSWebTest := func() (*domain.Sevens, []*domain.SevensPlayer) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		return s, players
	}

	t.Run("success Output initial state", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))

		result := tswp.Output(s, nil)
		assert.NotEmpty(t, result)

		var resObj controller.SevensWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 4, len(resObj.Players))
		assert.False(t, resObj.GameEndFlag)
		assert.Equal(t, 0, resObj.CurrentTurn)
		assert.Equal(t, "", resObj.Message)
		// Table initially all 7s
		assert.Equal(t, 7, resObj.TableMinVals[domain.CardDesignSpade])
		assert.Equal(t, 7, resObj.TableMaxVals[domain.CardDesignSpade])
		// Config defaults
		assert.False(t, resObj.Config.TunnelEnabled)
		assert.Equal(t, 0, resObj.Config.JokerCount)
		assert.False(t, resObj.Config.CpuStrategy)
		assert.False(t, resObj.Config.NoJokerFinish)
		// TablePlaced: bit 7 set for each suit = 128
		assert.Equal(t, 128, resObj.TablePlaced[domain.CardDesignSpade])
		assert.Equal(t, 128, resObj.TablePlaced[domain.CardDesignHeart])
	})

	t.Run("success Output shows human cards", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))

		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		humanPlayer := resObj.Players[0]
		assert.True(t, humanPlayer.IsHuman)
		assert.Equal(t, 2, humanPlayer.CardCount)
		assert.Len(t, humanPlayer.Cards, 2)
		assert.Equal(t, "SPADE", humanPlayer.Cards[0].Design)
		assert.Equal(t, 6, humanPlayer.Cards[0].Value)
	})

	t.Run("success Output CPU cards hidden", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 9, false))

		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		cpu1 := resObj.Players[1]
		assert.False(t, cpu1.IsHuman)
		assert.Equal(t, 2, cpu1.CardCount)
		assert.Len(t, cpu1.Cards, 0) // no cards shown for CPU
	})

	t.Run("success Output shows pass info", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // not playable
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 8, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		s.PlayerPlay(-1) // pass

		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		humanPlayer := resObj.Players[0]
		assert.Equal(t, 1, humanPlayer.PassesUsed)
		assert.Equal(t, domain.SevensMaxPasses, humanPlayer.MaxPasses)
		assert.NotNil(t, resObj.HumanAction)
		assert.Nil(t, resObj.HumanAction.PlayedCard) // pass → null
	})

	t.Run("success Output table state after play", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		s.PlayerPlay(0) // play 6♠ → minVal[Spade] = 6

		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Equal(t, 6, resObj.TableMinVals[domain.CardDesignSpade])
		assert.Equal(t, 7, resObj.TableMaxVals[domain.CardDesignSpade])
		assert.NotNil(t, resObj.HumanAction)
		assert.NotNil(t, resObj.HumanAction.PlayedCard)
		assert.Equal(t, 6, resObj.HumanAction.PlayedCard.Value)
	})

	t.Run("success Output game ended message", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		s.PlayerPlay(0) // human plays last card → game ends

		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.GameEndFlag)
		assert.Contains(t, resObj.Message, "ゲーム終了")
	})

	t.Run("success Output CPU actions after human play", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false)) // not playable
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		s.PlayerPlay(0) // human plays 8♠
		s.CpuPlay()     // CPU 1 passes

		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.CpuActions, 1)
		assert.Nil(t, resObj.CpuActions[0].PlayedCard) // pass
	})

	t.Run("success Output player rank", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].SetIsFinished(true)
		players[1].SetRank(1)
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		s.PlayerPlay(0) // human → rank 4

		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		humanPlayer := resObj.Players[0]
		assert.True(t, humanPlayer.IsFinished)
		assert.Equal(t, 4, humanPlayer.Rank)
	})

	t.Run("success Output joker card shown as JOKER", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.Len(t, resObj.Players[0].Cards, 2)
		assert.Equal(t, "JOKER", resObj.Players[0].Cards[0].Design)
		assert.Equal(t, 0, resObj.Players[0].Cards[0].Value)
	})

	t.Run("success Output config with features enabled", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		cfg := domain.SevensConfig{TunnelEnabled: true, JokerCount: 2, CpuStrategy: true, MaxPasses: 3, NoJokerFinish: true}
		s := domain.NewSevens(tc, players, cfg)

		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.True(t, resObj.Config.TunnelEnabled)
		assert.Equal(t, 2, resObj.Config.JokerCount)
		assert.True(t, resObj.Config.CpuStrategy)
		assert.Equal(t, 3, resObj.Config.MaxPasses)
		assert.True(t, resObj.Config.NoJokerFinish)
	})

	t.Run("success Output config NoJokerFinish true", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		cfg := domain.SevensConfig{NoJokerFinish: true, JokerCount: 1, MaxPasses: domain.SevensMaxPasses}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.True(t, resObj.Config.NoJokerFinish)
		assert.Equal(t, 1, resObj.Config.JokerCount)
	})

	t.Run("success Output tablePlaced updated after play", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		s.PlayerPlay(0) // play 6♠

		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		// bit 7 + bit 6 = 128 + 64 = 192
		assert.Equal(t, 192, resObj.TablePlaced[domain.CardDesignSpade])
	})

	t.Run("success Output joker action with target", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignJoker, 0, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 9, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[2].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		players[3].AddCard(domain.NewCard(domain.CardDesignHeart, 2, false))
		s.PlayerPlayJoker(0, domain.CardDesignSpade, 6) // joker → SPADE 6

		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		_ = json.Unmarshal([]byte(result), &resObj)
		assert.NotNil(t, resObj.HumanAction)
		assert.NotNil(t, resObj.HumanAction.PlayedCard)
		assert.Equal(t, "JOKER", resObj.HumanAction.PlayedCard.Design)
		assert.Equal(t, domain.CardDesignSpade, resObj.HumanAction.TargetSuit)
		assert.Equal(t, 6, resObj.HumanAction.TargetValue)
	})

	t.Run("success Output shows error message", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tswp.Output(s, domain.ErrInvalidPlay)
		var resObj controller.SevensWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Contains(t, resObj.Message, domain.ErrInvalidPlay.Error())
	})

	t.Run("success Output buildResultMessage with rank 0 skipped", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		s := domain.NewSevens(tc, players, domain.DefaultSevensConfig())
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 8, false))
		players[1].SetIsFinished(true)
		players[1].SetRank(0) // rank < 1 → skip in buildResultMessage
		players[2].SetIsFinished(true)
		players[2].SetRank(2)
		players[3].SetIsFinished(true)
		players[3].SetRank(3)
		s.PlayerPlay(0) // human plays last card → game ends

		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Contains(t, resObj.Message, "ゲーム終了")
		assert.NotContains(t, resObj.Message, "CPU 1") // skipped due to rank 0
	})

	t.Run("success Output humanAction nil", func(t *testing.T) {
		s, _ := setupSWebTest()
		// humanAction is nil by default (no play yet)
		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Nil(t, resObj.HumanAction)
	})

	t.Run("success Output getCardObj nil card in cpuActions", func(t *testing.T) {
		s, _ := setupSWebTest()
		// CPU pass action → PlayedCard is nil → getCardObj(nil) → nil
		s.SetCpuActions([]*domain.SevensCpuAction{
			{PlayerIdx: 1, PlayedCard: nil},
		})
		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Len(t, resObj.CpuActions, 1)
		assert.Nil(t, resObj.CpuActions[0].PlayedCard)
	})

	setupSMock := func() *interfaces.MockSevensGame {
		m := new(interfaces.MockSevensGame)
		m.On("GetCurrentTurn").Return(0)
		m.On("GetTableMinVals").Return([5]int{7, 7, 7, 7, 7})
		m.On("GetTableMaxVals").Return([5]int{7, 7, 7, 7, 7})
		m.On("GetTablePlaced").Return([5]uint16{128, 128, 128, 128, 128})
		m.On("GetConfig").Return(domain.SevensConfig{})
		m.On("GetCpuActions").Return([]*domain.SevensCpuAction(nil))
		m.On("GetHumanAction").Return((*domain.SevensCpuAction)(nil))
		return m
	}

	t.Run("success Output skips nil player in player loop", func(t *testing.T) {
		m := setupSMock()
		m.On("GetGameEndFlag").Return(false)
		m.On("GetPlayerCnt").Return(1)
		m.On("GetPlayer", 0).Return((*domain.SevensPlayer)(nil))

		result := tswp.Output(m, nil)
		var resObj controller.SevensWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Len(t, resObj.Players, 0) // nil player skipped
	})

	t.Run("success buildResultMessage skips nil player", func(t *testing.T) {
		m := setupSMock()
		m.On("GetGameEndFlag").Return(true)
		m.On("GetPlayerCnt").Return(1)
		m.On("GetPlayer", 0).Return((*domain.SevensPlayer)(nil))

		result := tswp.Output(m, nil)
		var resObj controller.SevensWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Len(t, resObj.Players, 0)           // nil player skipped in Output loop
		assert.Equal(t, "ゲーム終了！ ", resObj.Message) // buildResultMessage called
	})

	t.Run("success Output config includes MaxPasses", func(t *testing.T) {
		tc := domain.NewTrumpCards(0)
		players := makeSPlayers()
		cfg := domain.SevensConfig{MaxPasses: 0}
		s := domain.NewSevens(tc, players, cfg)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 6, false))

		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Equal(t, 0, resObj.Config.MaxPasses)
	})

	t.Run("success Output ForcedPass in CPU action", func(t *testing.T) {
		s, _ := setupSWebTest()
		s.SetCpuActions([]*domain.SevensCpuAction{
			{
				PlayerIdx:  1,
				PlayedCard: nil,
				ForcedPass: true,
			},
		})
		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.Len(t, resObj.CpuActions, 1)
		assert.True(t, resObj.CpuActions[0].ForcedPass)
	})

	t.Run("success Output ForcedPass in human action", func(t *testing.T) {
		s, _ := setupSWebTest()
		s.SetHumanAction(&domain.SevensCpuAction{
			PlayerIdx:  0,
			PlayedCard: nil,
			ForcedPass: true,
		})
		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.HumanAction)
		assert.True(t, resObj.HumanAction.ForcedPass)
	})

	t.Run("success Output non-forced pass has ForcedPass false", func(t *testing.T) {
		s, _ := setupSWebTest()
		s.SetHumanAction(&domain.SevensCpuAction{
			PlayerIdx:  0,
			PlayedCard: nil,
			ForcedPass: false,
		})
		result := tswp.Output(s, nil)
		var resObj controller.SevensWebOutput
		err := json.Unmarshal([]byte(result), &resObj)
		assert.NoError(t, err)
		assert.NotNil(t, resObj.HumanAction)
		assert.False(t, resObj.HumanAction.ForcedPass)
	})
}
