//go:build test

package presenter_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

func makeBigTwoPlayers() []*domain.BigTwoPlayer {
	return []*domain.BigTwoPlayer{
		domain.NewBigTwoPlayer(true),
		domain.NewBigTwoPlayer(false),
		domain.NewBigTwoPlayer(false),
		domain.NewBigTwoPlayer(false),
	}
}

func setupBigTwoCuiMock() (*interfaces.MockBigTwoGame, []*domain.BigTwoPlayer) {
	m := new(interfaces.MockBigTwoGame)
	players := makeBigTwoPlayers()
	m.On("GetGameEndFlag").Return(false)
	m.On("GetTableCards").Return(([]*domain.Card)(nil))
	m.On("GetCpuActions").Return(([]*domain.BigTwoAction)(nil))
	m.On("IsHumanTurn").Return(true)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

// **場の役名を出す。**Web は `bt-table-playtype` バッジで常時出しているのに、
// CUI は生のカード列だけで、何を出せば通るのかを読み取らせていた (#4859)。
func TestBigTwoCuiPresenter_ShowsTablePlayType(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.BigTwoCuiPresenter)

	build := func(cards []*domain.Card, pt domain.BigTwoPlayType) *interfaces.MockBigTwoGame {
		m := new(interfaces.MockBigTwoGame)
		players := makeBigTwoPlayers()
		// **defaults より先に登録する。**testify は最初に一致した期待値を使う。
		m.On("GetTableCards").Return(cards)
		m.On("GetTablePlayType").Return(pt)
		m.On("GetGameEndFlag").Return(false)
		m.On("GetCpuActions").Return(([]*domain.BigTwoAction)(nil))
		m.On("IsHumanTurn").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		m.On("GetPlayerCnt").Return(4)
		for i := 0; i < 4; i++ {
			m.On("GetPlayer", i).Return(players[i])
		}
		return m
	}

	t.Run("names the play on the table", func(t *testing.T) {
		cards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 3, false),
			domain.NewCard(domain.CardDesignHeart, 4, false),
			domain.NewCard(domain.CardDesignClover, 5, false),
			domain.NewCard(domain.CardDesignDiamond, 6, false),
			domain.NewCard(domain.CardDesignSpade, 7, false),
		}
		out := p.Output(build(cards, domain.BigTwoPlayStraight), nil)
		assert.Contains(t, out, "[ストレート]")
		assert.NotContains(t, out, "[フラッシュ]")
	})

	t.Run("a pair is named a pair", func(t *testing.T) {
		cards := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 9, false),
			domain.NewCard(domain.CardDesignHeart, 9, false),
		}
		assert.Contains(t, p.Output(build(cards, domain.BigTwoPlayPair), nil), "[ペア]")
	})

	t.Run("an empty table keeps the free-lead line alone", func(t *testing.T) {
		out := p.Output(build(nil, domain.BigTwoPlayInvalid), nil)
		assert.Contains(t, out, "場: なし (自由に出せます)")
		// 役名は付かない。Invalid を "" に落としているのがここ。
		assert.NotContains(t, out, "[シングル]")
		assert.NotContains(t, out, "[ペア]")
	})
}

func TestBigTwoCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BigTwoCuiPresenter)

	t.Run("initial empty table shows title and human turn", func(t *testing.T) {
		m, players := setupBigTwoCuiMock()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 3, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Big Two")
		assert.Contains(t, result, "自由に出せます")
		assert.Contains(t, result, "あなたのターン")
	})

	t.Run("cpu action line uses localized CPU name, not hardcoded", func(t *testing.T) {
		m, _ := setupBigTwoCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCpuActions")
		m.On("GetCpuActions").Return([]*domain.BigTwoAction{
			{PlayerIdx: 1, PlayedCards: nil},
			{PlayerIdx: 2, PlayedCards: []*domain.Card{domain.NewCard(domain.CardDesignHeart, 4, false)}},
		})
		result := p.Output(m, nil)
		// cuiPlayerName renders "CPU 1"/"CPU 2" via the shared cuiPlayerCpu key.
		assert.Contains(t, result, "CPU 1")
		assert.Contains(t, result, "CPU 2")
	})

	t.Run("game ended rankings use localized player names", func(t *testing.T) {
		m, players := setupBigTwoCuiMock()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.On("GetGameEndFlag").Return(true)
		players[0].SetRank(1)
		players[1].SetRank(2)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
		// Human -> "あなた", CPU -> "CPU 1" (both via cuiPlayerName).
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "CPU 1")
		assert.Contains(t, result, "1位")
	})

	t.Run("error is shown", func(t *testing.T) {
		m, _ := setupBigTwoCuiMock()
		assert.Contains(t, p.Output(m, errors.New("invalid play")), "invalid play")
	})
}

// 8 種類の役名が全部揃っていること。#5024 で switch にしたとき、patch coverage が
// 落ちた分をここで埋める。名前の抜けは Web のバッジとの食い違いになる。
func TestBigTwoCuiPresenter_NamesEveryPlayType(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.BigTwoCuiPresenter)

	cases := []struct {
		playType domain.BigTwoPlayType
		want     string
	}{
		{domain.BigTwoPlaySingle, "[シングル]"},
		{domain.BigTwoPlayPair, "[ペア]"},
		{domain.BigTwoPlayTriple, "[トリプル]"},
		{domain.BigTwoPlayStraight, "[ストレート]"},
		{domain.BigTwoPlayFlush, "[フラッシュ]"},
		{domain.BigTwoPlayFullHouse, "[フルハウス]"},
		{domain.BigTwoPlayFourOfAKind, "[フォーカード]"},
		{domain.BigTwoPlayStraightFlush, "[ストレートフラッシュ]"},
	}
	for _, tc := range cases {
		m := new(interfaces.MockBigTwoGame)
		players := makeBigTwoPlayers()
		m.On("GetTableCards").Return([]*domain.Card{domain.NewCard(domain.CardDesignSpade, 3, false)})
		m.On("GetTablePlayType").Return(tc.playType)
		m.On("GetGameEndFlag").Return(false)
		m.On("GetCpuActions").Return(([]*domain.BigTwoAction)(nil))
		m.On("IsHumanTurn").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		m.On("GetPlayerCnt").Return(4)
		for i := 0; i < 4; i++ {
			m.On("GetPlayer", i).Return(players[i])
		}
		assert.Contains(t, p.Output(m, nil), tc.want)
	}
}

// **エラー行が赤くないと通常の状態行と見分けが付かない (#4821)。**共通ヘルパーの
// cuiErrorBlock を使っているかを、色を出したまま確認する。
func TestBigTwoCuiPresenter_ErrorIsRed(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(orig)

	m, _ := setupBigTwoCuiMock()
	out := new(presenter.BigTwoCuiPresenter).Output(m, errors.New("invalid play"))
	assert.Contains(t, out, color.Red("invalid play"))
}
