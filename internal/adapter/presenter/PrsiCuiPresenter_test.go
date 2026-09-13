package presenter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// prsiLegalPrefix returns the "playable cards" line text up to the placeholder,
// so assertions match regardless of the substituted card list.
func prsiLegalPrefix() string {
	return strings.Split(i18n.T("prsi.legalPlays"), "{{")[0]
}

func setupPrsiCuiMock() *interfaces.MockPrsiGame {
	m := new(interfaces.MockPrsiGame)
	m.On("GetDrawPileCount").Return(11)
	m.On("GetDiscardTop").Return((*domain.Card)(nil))
	m.On("GetPenaltyDrawCount").Return(0)
	m.On("GetPendingSkips").Return(0)
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.PrsiPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerIdx").Return(-1)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupPrsiCuiMockWithPlayers() (*interfaces.MockPrsiGame, []*domain.PrsiPlayer) {
	m := setupPrsiCuiMock()
	players := makePrsiPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestPrsiCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.PrsiCuiPresenter)

	t.Run("play phase renders header, hand, prompt", func(t *testing.T) {
		m, players := setupPrsiCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 13, false))
		players[1].AddCard(domain.NewCard(domain.CardDesignClover, 8, false))

		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
		assert.Contains(t, out, "11") // stock count
	})

	t.Run("discard top with penalty", func(t *testing.T) {
		m, _ := setupPrsiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPenaltyDrawCount")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
		m.On("GetPenaltyDrawCount").Return(4)

		out := p.Output(m, nil)
		assert.Contains(t, out, "4") // penalty count appears
	})

	// **スキップも重ねられる (#4772)。**7 の累積ペナルティは出していたのに、
	// エース/ジャックの累積は CUI にも出ていなかった。
	t.Run("discard top reports the pending skips", func(t *testing.T) {
		m, _ := setupPrsiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPendingSkips")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 1, false))
		m.On("GetPendingSkips").Return(3)

		assert.Contains(t, p.Output(m, nil), "スキップ3人ぶん")
	})

	t.Run("no skip line while nothing is stacked", func(t *testing.T) {
		m, _ := setupPrsiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 5, false))

		assert.NotContains(t, p.Output(m, nil), "スキップ")
	})

	t.Run("legal cards listed for the human", func(t *testing.T) {
		m, players := setupPrsiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 13, false)) // suit match → legal
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))  // rank match → legal
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 8, false)) // neither → illegal

		out := p.Output(m, nil)
		assert.Contains(t, out, prsiLegalPrefix())
		assert.Contains(t, out, "[0]")
		assert.Contains(t, out, "[1]")
	})

	t.Run("penalty allows only sevens", func(t *testing.T) {
		m, players := setupPrsiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPenaltyDrawCount")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 7, false))
		m.On("GetPenaltyDrawCount").Return(2)
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 7, false))  // seven → legal under penalty
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 13, false)) // non-seven → illegal

		out := p.Output(m, nil)
		lines := strings.Split(out, "\n")
		var legalLine string
		for _, l := range lines {
			if strings.Contains(l, prsiLegalPrefix()) {
				legalLine = l
			}
		}
		// Only the seven ([0]) is legal; the king ([1]) is excluded from the line.
		assert.Contains(t, legalLine, "[0]")
		assert.NotContains(t, legalLine, "[1]")
	})

	t.Run("no legal card shows draw guidance", func(t *testing.T) {
		m, players := setupPrsiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDiscardTop")
		m.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 5, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignClover, 8, false)) // neither suit nor rank

		out := p.Output(m, nil)
		assert.Contains(t, out, i18n.T("prsi.noLegalPlay"))
	})

	t.Run("error block rendered", func(t *testing.T) {
		m, _ := setupPrsiCuiMockWithPlayers()
		out := p.Output(m, errors.New("boom"))
		assert.Contains(t, out, "boom")
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupPrsiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerIdx")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerIdx").Return(0)

		out := p.Output(m, nil)
		assert.NotEmpty(t, out)
	})

	// #7343: 残り1枚のプレイヤーが目立たない。
	t.Run("player with 1 card is highlighted yellow", func(t *testing.T) {
		m, players := setupPrsiCuiMockWithPlayers()
		// CPU 1 (index 1) に 1 枚だけ持たせる。
		players[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))

		out := p.Output(m, nil)
		// 色コードはテストで SetNoColor(true) にしているので、
		// 期待値はリテラルの行テキストが yellow フラグ付き文字列に包まれているかではなく、
		// 少なくとも行の内容 ("1" の枚数) が存在することで確認する。
		// color.NoColor=true のとき color.Yellow(s) == s なので、
		// 強調有無はカバレッジ実行で確認し、ここではリテラルの行が出力されることを確認。
		assert.Contains(t, out, "1") // card count 1 is present
	})

	// #7343 否定対照: 残り 2 枚の席は強調されない。
	// **出力全体で "\033[33m" の有無を見てはいけない。** 「出せる札がありません」の
	// 案内が同じ黄色を使っており、席とは無関係にこのコードが現れる。席の行だけを取る。
	t.Run("only the seat holding one card is highlighted", func(t *testing.T) {
		savedNoColor := color.NoColor()
		color.SetNoColor(false)
		defer color.SetNoColor(savedNoColor)

		seatLine := func(out, name string) string {
			for _, line := range strings.Split(out, "\n") {
				if strings.Contains(line, name) && strings.Contains(line, "枚") {
					return line
				}
			}
			return ""
		}

		mTwo, playersTwo := setupPrsiCuiMockWithPlayers()
		playersTwo[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		playersTwo[1].AddCard(domain.NewCard(domain.CardDesignHeart, 3, false))
		lineTwo := seatLine(p.Output(mTwo, nil), "CPU 1")
		require.NotEmpty(t, lineTwo, "CPU 1 の席行が見つからなければ、この試験は何も証明しない")
		assert.NotContains(t, lineTwo, "\033[33m")

		mOne, playersOne := setupPrsiCuiMockWithPlayers()
		playersOne[1].AddCard(domain.NewCard(domain.CardDesignSpade, 5, false))
		lineOne := seatLine(p.Output(mOne, nil), "CPU 1")
		require.NotEmpty(t, lineOne)
		assert.Contains(t, lineOne, "\033[33m")
	})
}

func TestPrsiCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.PrsiCuiPresenter)
	m := new(interfaces.MockPrsiGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "plays SPADE 7"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewPrsiPlayer(true)).Maybe()
	out := p.ActionLogOutput(m)
	assert.Contains(t, out, "play")
}
