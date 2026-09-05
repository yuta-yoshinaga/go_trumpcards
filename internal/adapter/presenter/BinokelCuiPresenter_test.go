//go:build test

package presenter_test

import (
	"errors"
	"strconv"
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

func setupBinokelCuiMock() *interfaces.MockBinokelGame {
	m := new(interfaces.MockBinokelGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BinokelPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetHighestBid").Return(150)
	m.On("GetHighestBidder").Return(0)
	m.On("GetScores").Return([domain.BinokelPlayerCnt]int{0, 0, 0})
	m.On("GetScore", 0).Return(0)
	m.On("GetScore", 1).Return(0)
	m.On("GetScore", 2).Return(0)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetDabb").Return([]*domain.Card(nil))
	m.On("GetDabbDiscarded").Return([]*domain.Card(nil))
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultBinokelConfig())
	m.On("GetPlayerMelds").Return([domain.BinokelPlayerCnt][]*domain.BinokelMeld{})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("IsHumanTurn").Return(true)
	m.On("GetValidPlayIndices", 0).Return([]int{0, 1})
	return m
}

func setupBinokelCuiMockWithPlayers() (*interfaces.MockBinokelGame, []*domain.BinokelPlayer) {
	m := setupBinokelCuiMock()
	players := []*domain.BinokelPlayer{
		domain.NewBinokelPlayer(true),
		domain.NewBinokelPlayer(false),
		domain.NewBinokelPlayer(false),
	}
	m.On("GetPlayerCnt").Return(3)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	return m, players
}

func TestBinokelCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BinokelCuiPresenter)

	t.Run("shows game header and round/trick info", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "Binokel (ビノクル)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "トリック: 1")
	})

	t.Run("shows trump suit when set", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: SPADE (落札者: あなた)")
	})

	t.Run("shows undecided trump when trumpSuit is 0", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: 未決定")
	})

	t.Run("shows highest bid and bidder", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "最高ビッド: 150 (あなた)")
	})

	t.Run("shows player scores", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "累計スコア: あなた=0  CPU 1=0  CPU 2=0")
	})

	t.Run("shows dealer", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "ディーラー: あなた")
	})

	t.Run("shows human player cards", func(t *testing.T) {
		m, players := setupBinokelCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "[0]SPADE 1")
	})

	t.Run("shows legal-play legend on human play turn", func(t *testing.T) {
		m, players := setupBinokelCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "合法手: [0] [1]")
	})

	t.Run("hides legal-play legend outside the play phase", func(t *testing.T) {
		m, players := setupBinokelCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BinokelPhaseMeld)
		result := p.Output(m, nil)
		assert.NotContains(t, result, "合法手:")
	})

	t.Run("shows current trick on table", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentTrick")
		trick := []*domain.TrickCard{
			{PlayerIdx: 0, Card: domain.NewCard(domain.CardDesignSpade, 1, false)},
			{PlayerIdx: 1, Card: domain.NewCard(domain.CardDesignHeart, 10, false)},
		}
		m.On("GetCurrentTrick").Return(trick)
		result := p.Output(m, nil)
		assert.Contains(t, result, "テーブル:")
		assert.Contains(t, result, "SPADE 1")
	})

	t.Run("shows melds in meld phase", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerMelds")
		m.On("GetPhase").Return(domain.BinokelPhaseMeld)
		melds := [domain.BinokelPlayerCnt][]*domain.BinokelMeld{}
		melds[0] = []*domain.BinokelMeld{{Type: domain.BinokelMeldBinokel, Points: 40}}
		m.On("GetPlayerMelds").Return(melds)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ビノクル")
		assert.Contains(t, result, "40点")
	})

	t.Run("shows dabb cards and prompt in dabb phase", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDabb")
		m.On("GetPhase").Return(domain.BinokelPhaseDabb)
		dabb := []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 1, false),
			domain.NewCard(domain.CardDesignHeart, 10, false),
			domain.NewCard(domain.CardDesignDiamond, 11, false),
		}
		m.On("GetDabb").Return(dabb)
		result := p.Output(m, nil)
		assert.Contains(t, result, "Dabb: SPADE 1 HEART 10 DIAMOND 11")
		assert.Contains(t, result, "discard")
	})

	t.Run("shows error message", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		result := p.Output(m, errors.New("invalid bid"))
		assert.Contains(t, result, "invalid bid")
	})

	t.Run("shows game end message", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！ あなたの勝ち！")
	})

	t.Run("shows bid phase message", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BinokelPhaseBid)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッド番")
	})

	t.Run("shows trump phase message", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BinokelPhaseTrump)
		result := p.Output(m, nil)
		assert.Contains(t, result, "トランプスートを選んで")
	})

	t.Run("shows meld phase message", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BinokelPhaseMeld)
		result := p.Output(m, nil)
		assert.Contains(t, result, "メルドフェーズ")
	})

	t.Run("shows play phase message", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "あなたの番")
	})

	t.Run("shows trick end message", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BinokelPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
	})

	t.Run("shows round end message", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BinokelPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
	})

	t.Run("bidder index -1 hides bidder name", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHighestBidder")
		m.On("GetHighestBidder").Return(-1)
		result := p.Output(m, nil)
		assert.Contains(t, result, "最高ビッド:")
	})

	t.Run("melds shown in round end phase too", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerMelds")
		m.On("GetPhase").Return(domain.BinokelPhaseRoundEnd)
		melds := [domain.BinokelPlayerCnt][]*domain.BinokelMeld{}
		melds[1] = []*domain.BinokelMeld{{Type: domain.BinokelMeldCommonMarriage, Points: 20}}
		m.On("GetPlayerMelds").Return(melds)
		result := p.Output(m, nil)
		assert.Contains(t, result, "コモンマリッジ")
	})
}

func TestBinokelCuiPresenter_BidDistinction(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BinokelCuiPresenter)

	t.Run("distinguishes unbid, passed, and declared bids on the same table (Japanese)", func(t *testing.T) {
		m, players := setupBinokelCuiMockWithPlayers()
		// Player 0 (Human): has not bid yet
		players[0].SetBid(0)
		players[0].SetHasPassed(false)
		// Player 1 (CPU 1): has passed
		players[1].SetBid(0)
		players[1].SetHasPassed(true)
		// Player 2 (CPU 2): declared 150
		players[2].SetBid(150)
		players[2].SetHasPassed(false)

		result := p.Output(m, nil)

		assert.Contains(t, result, "あなた: スコア:0点 ビッド:未ビッド メルド:0点")
		assert.Contains(t, result, "CPU 1: スコア:0点 ビッド:パス メルド:0点")
		assert.Contains(t, result, "CPU 2: スコア:0点 ビッド:150 メルド:0点")
	})

	t.Run("shows highest bid as none when nobody has bid (Japanese)", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHighestBid")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHighestBidder")
		m.On("GetHighestBid").Return(0)
		m.On("GetHighestBidder").Return(-1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "最高ビッド: なし")
		assert.NotContains(t, result, "最高ビッド: 0")
	})

	t.Run("meld table uses localized family names (Japanese)", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BinokelPhaseBid)

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札ファミリー 150")
		assert.Contains(t, result, "ダブル切り札ファミリー 1500")
		assert.NotContains(t, result, "ラン 150")
		assert.NotContains(t, result, "ダブルラン 1500")
	})
}

func TestBinokelCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BinokelCuiPresenter)

	t.Run("nil hint returns no hint message", func(t *testing.T) {
		m := new(interfaces.MockBinokelGame)
		m.On("GetHint").Return((*domain.BinokelHint)(nil))
		result := p.HintOutput(m)
		assert.Equal(t, "ヒントなし", result)
	})

	t.Run("bid amount hint", func(t *testing.T) {
		m := new(interfaces.MockBinokelGame)
		bid := 150
		m.On("GetHint").Return(&domain.BinokelHint{BidAmount: &bid, Reason: "hint_bid"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ビッド 150")
	})

	t.Run("pass hint", func(t *testing.T) {
		m := new(interfaces.MockBinokelGame)
		pass := true
		m.On("GetHint").Return(&domain.BinokelHint{Pass: &pass, Reason: "hint_pass"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "パス推奨")
	})

	t.Run("suit hint", func(t *testing.T) {
		m := new(interfaces.MockBinokelGame)
		suit := 2
		m.On("GetHint").Return(&domain.BinokelHint{Suit: &suit, Reason: "hint_trump"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "スート")
	})

	t.Run("card index hint", func(t *testing.T) {
		m := new(interfaces.MockBinokelGame)
		idx := 3
		m.On("GetHint").Return(&domain.BinokelHint{CardIndex: &idx, Reason: "hint_play"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "カード 3")
	})
}

func TestBinokelCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BinokelCuiPresenter)

	t.Run("nil entries returns no log message", func(t *testing.T) {
		m := new(interfaces.MockBinokelGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		m.On("GetPlayer", mock.Anything).Return(domain.NewBinokelPlayer(true)).Maybe()
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("with entries returns log", func(t *testing.T) {
		m := new(interfaces.MockBinokelGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "bid", Detail: "bid 150"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		m.On("GetPlayer", mock.Anything).Return(domain.NewBinokelPlayer(true)).Maybe()
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "bid")
	})
}

func TestBinokelCuiPresenter_English(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	p := new(presenter.BinokelCuiPresenter)

	t.Run("hint output uses English labels", func(t *testing.T) {
		m := new(interfaces.MockBinokelGame)
		m.On("GetHint").Return((*domain.BinokelHint)(nil))
		assert.Equal(t, "No hint", p.HintOutput(m))
	})

	t.Run("hint suit uses English label", func(t *testing.T) {
		m := new(interfaces.MockBinokelGame)
		suit := 2
		m.On("GetHint").Return(&domain.BinokelHint{Suit: &suit, Reason: "hint_trump"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "Hint:")
		assert.Contains(t, result, "suit")
	})

	t.Run("hint card index uses English label", func(t *testing.T) {
		m := new(interfaces.MockBinokelGame)
		idx := 3
		m.On("GetHint").Return(&domain.BinokelHint{CardIndex: &idx, Reason: "hint_play"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "card 3")
	})

	t.Run("output uses English game-end banner", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "Game over! You wins!")
		assert.NotContains(t, result, "チーム")
	})

	t.Run("output uses English headers, dealer, scores, prompt", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "Round: 1")
		assert.Contains(t, result, "Trick: 1")
		assert.Contains(t, result, "Dealer: You")
		assert.Contains(t, result, "Scores: You=0  CPU 1=0  CPU 2=0")
		assert.Contains(t, result, "to play")
	})

	t.Run("distinguishes unbid, passed, and declared bids on the same table (English)", func(t *testing.T) {
		m, players := setupBinokelCuiMockWithPlayers()
		// Player 0 (Human): has not bid yet
		players[0].SetBid(0)
		players[0].SetHasPassed(false)
		// Player 1 (CPU 1): has passed
		players[1].SetBid(0)
		players[1].SetHasPassed(true)
		// Player 2 (CPU 2): declared 150
		players[2].SetBid(150)
		players[2].SetHasPassed(false)

		result := p.Output(m, nil)

		assert.Contains(t, result, "You: score:0pt bid:no bid meld:0pt")
		assert.Contains(t, result, "CPU 1: score:0pt bid:pass meld:0pt")
		assert.Contains(t, result, "CPU 2: score:0pt bid:150 meld:0pt")
	})

	t.Run("shows highest bid as none when nobody has bid (English)", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHighestBid")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHighestBidder")
		m.On("GetHighestBid").Return(0)
		m.On("GetHighestBidder").Return(-1)

		result := p.Output(m, nil)
		assert.Contains(t, result, "Highest bid: none")
		assert.NotContains(t, result, "Highest bid: 0")
	})

	t.Run("meld table uses localized family names (English)", func(t *testing.T) {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BinokelPhaseBid)

		result := p.Output(m, nil)
		assert.Contains(t, result, "Trump family 150")
		assert.Contains(t, result, "Double trump family 1500")
		assert.NotContains(t, result, "Run 150")
		assert.NotContains(t, result, "Double run 1500")
	})
}

func TestBinokelCuiPresenter_MeldTable(t *testing.T) {
	p := new(presenter.BinokelCuiPresenter)

	outputInPhase := func(phase domain.BinokelPhase) string {
		m, _ := setupBinokelCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(phase)
		return p.Output(m, nil)
	}

	entriesOf := func(out string) []string {
		header := i18n.T("binokel.meldTableHeader")
		idx := strings.Index(out, header)
		if idx < 0 {
			return nil
		}
		var entries []string
		for _, line := range strings.Split(out[idx+len(header):], "\n") {
			lineEntries := strings.Split(strings.TrimSpace(line), " / ")
			for _, e := range lineEntries {
				e = strings.TrimSpace(e)
				fields := strings.Fields(e)
				if len(fields) < 2 {
					return entries
				}
				if _, err := strconv.Atoi(fields[len(fields)-1]); err != nil {
					return entries
				}
				entries = append(entries, e)
			}
		}
		return entries
	}

	t.Run("lists every meld with the points the domain scores it at", func(t *testing.T) {
		entries := entriesOf(outputInPhase(domain.BinokelPhaseBid))
		table := domain.BinokelMeldTable()
		require.Len(t, entries, len(table))

		for i, e := range table {
			assert.True(t, strings.HasSuffix(entries[i], " "+strconv.Itoa(e.Points)),
				"entry %d = %q, want it to end with %d points", i, entries[i], e.Points)
			name := strings.TrimSuffix(entries[i], " "+strconv.Itoa(e.Points))
			assert.NotContains(t, name, "meld#")
			assert.NotContains(t, name, "binokel.")
			assert.NotEmpty(t, name)
		}
		assert.Equal(t, "ディクス 10", entries[0])
		assert.Equal(t, "ダブル切り札ファミリー 1500", entries[len(entries)-1])
	})

	t.Run("is available while choosing melds too", func(t *testing.T) {
		assert.Contains(t, outputInPhase(domain.BinokelPhaseMeld), i18n.T("binokel.meldTableHeader"))
	})

	t.Run("is not printed once the cards are being played", func(t *testing.T) {
		assert.NotContains(t, outputInPhase(domain.BinokelPhasePlay), i18n.T("binokel.meldTableHeader"))
	})
}
