//go:build test

package usecase_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

const mightyMockOutput = `{"phase":0}`

func newMightyMocks() (*interfaces.MockMightyGame, *presenter.MockMightyPresenter) {
	return new(interfaces.MockMightyGame), new(presenter.MockMightyPresenter)
}

func TestNewMightyInteractor_NilGuards(t *testing.T) {
	mp := new(presenter.MockMightyPresenter)

	t.Run("panics when m is nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "MightyInteractor: m must not be nil", func() {
			usecase.NewMightyInteractor(nil, mp)
		})
	})

	t.Run("panics when mp is nil", func(t *testing.T) {
		g := new(interfaces.MockMightyGame)
		assert.PanicsWithValue(t, "MightyInteractor: mp must not be nil", func() {
			usecase.NewMightyInteractor(g, nil)
		})
	})
}

func TestMightyInteractor_Reset(t *testing.T) {
	t.Run("reset stays in bid phase", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("Reset").Return()
		g.On("GetGameEndFlag").Return(false)
		g.On("GetPhase").Return(domain.MightyPhaseBid)
		g.On("IsHumanBidTurn").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		assert.Equal(t, mightyMockOutput, mi.Reset())
		g.AssertCalled(t, "Reset")
	})

	t.Run("reset already in play phase triggers CPU turns", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("Reset").Return()
		g.On("GetGameEndFlag").Return(false)
		g.On("GetPhase").Return(domain.MightyPhasePlay)
		g.On("IsHumanDeclareTurn").Return(true)
		g.On("IsHumanExchangeTurn").Return(true)
		g.On("IsHumanTurn").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		assert.Equal(t, mightyMockOutput, mi.Reset())
	})
}

func TestMightyInteractor_ResetWithConfig(t *testing.T) {
	t.Run("sets config then resets", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		cfg := domain.MightyConfig{
			CpuDifficulty: domain.MightyCpuDifficultyHard,
			MinBid:        15,
			NoTrumpExtra:  3,
			PointLimit:    150,
		}
		g.On("SetConfig", cfg).Return()
		g.On("Reset").Return()
		g.On("GetGameEndFlag").Return(false)
		g.On("GetPhase").Return(domain.MightyPhaseBid)
		g.On("IsHumanBidTurn").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		assert.Equal(t, mightyMockOutput, mi.ResetWithConfig(cfg))
		g.AssertCalled(t, "SetConfig", cfg)
	})

	t.Run("invalid config short-circuits before Reset", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		invalid := domain.MightyConfig{
			CpuDifficulty: domain.MightyCpuDifficulty(99),
			MinBid:        13,
			NoTrumpExtra:  2,
			PointLimit:    100,
		}

		mi := usecase.NewMightyInteractor(g, mp)
		mi.ResetWithConfig(invalid)
		mp.AssertCalled(t, "Output", g, mock.MatchedBy(func(err error) bool { return err != nil }))
		g.AssertNotCalled(t, "Reset")
	})
}

func TestMightyInteractor_Bid(t *testing.T) {
	t.Run("game ended short-circuits", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		assert.Equal(t, mightyMockOutput, mi.Bid(14, false))
		g.AssertNotCalled(t, "PlayerBid", mock.Anything, mock.Anything)
	})

	t.Run("invalid bid returns error output", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(false)
		bidErr := errors.New("bad bid")
		g.On("PlayerBid", 5, false).Return(bidErr)

		mi := usecase.NewMightyInteractor(g, mp)
		mi.Bid(5, false)
		mp.AssertCalled(t, "Output", g, bidErr)
	})

	t.Run("valid no-trump bid loops CPU bids and stays in bid phase", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerBid", 15, true).Return(nil)
		g.On("GetPhase").Return(domain.MightyPhaseBid)
		g.On("IsHumanBidTurn").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		assert.Equal(t, mightyMockOutput, mi.Bid(15, true))
		g.AssertCalled(t, "PlayerBid", 15, true)
	})
}

func TestMightyInteractor_DeclareTrumpAndFriend(t *testing.T) {
	t.Run("game ended short-circuits", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		mi.DeclareTrumpAndFriend(domain.CardDesignHeart, domain.CardDesignSpade, 1)
		g.AssertNotCalled(t, "PlayerDeclareTrumpAndFriend", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("invalid declaration returns error output", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(false)
		declErr := errors.New("bad declare")
		g.On("PlayerDeclareTrumpAndFriend", domain.CardDesignHeart, domain.CardDesignSpade, 1).Return(declErr)

		mi := usecase.NewMightyInteractor(g, mp)
		mi.DeclareTrumpAndFriend(domain.CardDesignHeart, domain.CardDesignSpade, 1)
		mp.AssertCalled(t, "Output", g, declErr)
	})

	t.Run("successful declaration leaves human in kitty exchange", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerDeclareTrumpAndFriend", domain.CardDesignHeart, domain.CardDesignSpade, 1).Return(nil)
		g.On("GetPhase").Return(domain.MightyPhaseKittyExchange)
		g.On("IsHumanExchangeTurn").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		assert.Equal(t, mightyMockOutput, mi.DeclareTrumpAndFriend(domain.CardDesignHeart, domain.CardDesignSpade, 1))
	})
}

func TestMightyInteractor_ExchangeKitty(t *testing.T) {
	indices := []int{0, 1, 2}

	t.Run("game ended short-circuits", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		mi.ExchangeKitty(indices)
		g.AssertNotCalled(t, "PlayerExchangeKitty", mock.Anything)
	})

	t.Run("error from PlayerExchangeKitty surfaces", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(false)
		exErr := errors.New("bad kitty")
		g.On("PlayerExchangeKitty", indices).Return(exErr)

		mi := usecase.NewMightyInteractor(g, mp)
		mi.ExchangeKitty(indices)
		mp.AssertCalled(t, "Output", g, exErr)
	})

	t.Run("successful exchange transitions to play and runs CPUs", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(false)
		g.On("PlayerExchangeKitty", indices).Return(nil)
		g.On("GetPhase").Return(domain.MightyPhasePlay)
		g.On("IsHumanTurn").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		assert.Equal(t, mightyMockOutput, mi.ExchangeKitty(indices))
	})
}

func TestMightyInteractor_Play(t *testing.T) {
	t.Run("game ended short-circuits", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		mi.Play(0)
		g.AssertNotCalled(t, "PlayerPlay", mock.Anything)
	})

	t.Run("invalid card returns error output", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(false)
		g.On("IsHumanTurn").Return(true)
		g.On("GetPhase").Return(domain.MightyPhasePlay)
		playErr := errors.New("bad play")
		g.On("PlayerPlay", 2).Return(playErr)

		mi := usecase.NewMightyInteractor(g, mp)
		mi.Play(2)
		mp.AssertCalled(t, "Output", g, playErr)
	})

	t.Run("valid play advances and runs CPUs", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(false)
		g.On("GetPhase").Return(domain.MightyPhasePlay)
		g.On("PlayerPlay", 3).Return(nil)
		g.On("IsHumanTurn").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		assert.Equal(t, mightyMockOutput, mi.Play(3))
	})
}

func TestMightyInteractor_PlayJokerLead(t *testing.T) {
	t.Run("game ended short-circuits", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		mi.PlayJokerLead(0, domain.CardDesignHeart)
		g.AssertNotCalled(t, "PlayerPlayJokerLead", mock.Anything, mock.Anything)
	})

	t.Run("error surfaces via presenter", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(false)
		g.On("IsHumanTurn").Return(true)
		g.On("GetPhase").Return(domain.MightyPhasePlay)
		leadErr := errors.New("bad joker lead")
		g.On("PlayerPlayJokerLead", 0, domain.CardDesignHeart).Return(leadErr)

		mi := usecase.NewMightyInteractor(g, mp)
		mi.PlayJokerLead(0, domain.CardDesignHeart)
		mp.AssertCalled(t, "Output", g, leadErr)
	})

	t.Run("successful joker lead runs CPU turns", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(false)
		g.On("GetPhase").Return(domain.MightyPhasePlay)
		g.On("PlayerPlayJokerLead", 0, domain.CardDesignHeart).Return(nil)
		g.On("IsHumanTurn").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		assert.Equal(t, mightyMockOutput, mi.PlayJokerLead(0, domain.CardDesignHeart))
	})
}

func TestMightyInteractor_NextTrick(t *testing.T) {
	t.Run("game ended short-circuits", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		mi.NextTrick()
		g.AssertNotCalled(t, "NextTrick")
	})

	t.Run("advances to next trick and runs CPUs", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("GetGameEndFlag").Return(false)
		g.On("NextTrick").Return()
		g.On("GetPhase").Return(domain.MightyPhasePlay)
		g.On("IsHumanTurn").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		assert.Equal(t, mightyMockOutput, mi.NextTrick())
		g.AssertCalled(t, "NextTrick")
	})
}

func TestMightyInteractor_NextRound(t *testing.T) {
	t.Run("scores, advances round, then runs CPUs", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("ScoreRound").Return()
		g.On("GetGameEndFlag").Return(false)
		g.On("NextRound").Return()
		g.On("GetPhase").Return(domain.MightyPhaseBid)
		g.On("IsHumanBidTurn").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		assert.Equal(t, mightyMockOutput, mi.NextRound())
		g.AssertCalled(t, "ScoreRound")
		g.AssertCalled(t, "NextRound")
	})

	t.Run("game end after scoring stops the chain", func(t *testing.T) {
		g, mp := newMightyMocks()
		mp.On("Output", mock.Anything, mock.Anything).Return(mightyMockOutput)
		g.On("ScoreRound").Return()
		g.On("GetGameEndFlag").Return(true)

		mi := usecase.NewMightyInteractor(g, mp)
		assert.Equal(t, mightyMockOutput, mi.NextRound())
		g.AssertCalled(t, "ScoreRound")
		g.AssertNotCalled(t, "NextRound")
	})
}

func TestMightyInteractor_GetConfig(t *testing.T) {
	g, mp := newMightyMocks()
	cfg := domain.MightyConfig{CpuDifficulty: domain.MightyCpuDifficultyNormal, MinBid: 13, NoTrumpExtra: 2, PointLimit: 100}
	g.On("GetConfig").Return(cfg)

	mi := usecase.NewMightyInteractor(g, mp)
	assert.Equal(t, cfg, mi.GetConfig())
}

func TestMightyInteractor_Hint(t *testing.T) {
	g, mp := newMightyMocks()
	mp.On("HintOutput", g).Return(`{"hint":"play 0"}`)
	mi := usecase.NewMightyInteractor(g, mp)
	assert.Equal(t, `{"hint":"play 0"}`, mi.Hint())
}

func TestMightyInteractor_ActionLog(t *testing.T) {
	g, mp := newMightyMocks()
	mp.On("ActionLogOutput", g).Return(`[]`)
	mi := usecase.NewMightyInteractor(g, mp)
	assert.Equal(t, `[]`, mi.ActionLog())
}

func TestRestoreMightyInteractor(t *testing.T) {
	t.Run("invalid JSON returns error", func(t *testing.T) {
		mp := new(presenter.MockMightyPresenter)
		_, err := usecase.RestoreMightyInteractor([]byte("not-json"), mp)
		assert.Error(t, err)
	})

	t.Run("valid snapshot round-trips", func(t *testing.T) {
		// Capture a snapshot from a real default game and verify Restore works.
		realGame := domain.NewDefaultMighty()
		realGame.Reset()
		data, err := realGame.MarshalJSON()
		assert.NoError(t, err)

		mp := new(presenter.MockMightyPresenter)
		mi, err := usecase.RestoreMightyInteractor(data, mp)
		assert.NoError(t, err)
		assert.NotNil(t, mi)
	})
}

func TestMightyInteractor_Snapshot(t *testing.T) {
	// Snapshot is provided by the embedded GameBase; verifying it round-trips
	// through MightyInteractor by using a real Mighty as the underlying game.
	realGame := domain.NewDefaultMighty()
	realGame.Reset()
	mp := new(presenter.MockMightyPresenter)
	mi := usecase.NewMightyInteractor(realGame, mp)

	data, err := mi.Snapshot()
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}
