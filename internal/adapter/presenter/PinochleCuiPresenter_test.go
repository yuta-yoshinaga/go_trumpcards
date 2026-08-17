//go:build test

package presenter_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupPinochleCuiMock() *interfaces.MockPinochleGame {
	m := new(interfaces.MockPinochleGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.PinochlePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetHighestBid").Return(20)
	m.On("GetHighestBidder").Return(0)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultPinochleConfig())
	m.On("GetPlayerMelds").Return([domain.PinochlePlayerCnt][]*domain.PinochleMeld{})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	m.On("IsHumanTurn").Return(true)
	m.On("GetValidPlayIndices", 0).Return([]int{0, 1})
	return m
}

func setupPinochleCuiMockWithPlayers() (*interfaces.MockPinochleGame, []*domain.PinochlePlayer) {
	m := setupPinochleCuiMock()
	players := []*domain.PinochlePlayer{
		domain.NewPinochlePlayer(true, 0),
		domain.NewPinochlePlayer(false, 1),
		domain.NewPinochlePlayer(false, 0),
		domain.NewPinochlePlayer(false, 1),
	}
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestPinochleCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.PinochleCuiPresenter)

	t.Run("shows game header and round/trick info", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "Pinochle (ピノクル)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "トリック: 1")
	})

	t.Run("shows trump suit when set", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: SPADE")
	})

	t.Run("shows undecided trump when trumpSuit is 0", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: 未決定")
	})

	t.Run("shows highest bid and bidder", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "最高ビッド: 20")
		assert.Contains(t, result, "あなた")
	})

	t.Run("shows team scores", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "チーム0: 0点  チーム1: 0点")
	})

	t.Run("shows dealer", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "ディーラー: あなた")
	})

	t.Run("shows human player cards", func(t *testing.T) {
		m, players := setupPinochleCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "[0]SPADE 1")
	})

	t.Run("shows legal-play legend on human play turn", func(t *testing.T) {
		m, players := setupPinochleCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 10, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "合法手: [0] [1]")
	})

	t.Run("hides legal-play legend outside the play phase", func(t *testing.T) {
		m, players := setupPinochleCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PinochlePhaseMeld)
		result := p.Output(m, nil)
		assert.NotContains(t, result, "合法手:")
	})

	t.Run("shows current trick on table", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
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
		m, _ := setupPinochleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerMelds")
		m.On("GetPhase").Return(domain.PinochlePhaseMeld)
		melds := [domain.PinochlePlayerCnt][]*domain.PinochleMeld{}
		melds[0] = []*domain.PinochleMeld{{Type: domain.PinochleMeldPinochle, Points: 40}}
		m.On("GetPlayerMelds").Return(melds)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ピノクル")
		assert.Contains(t, result, "40点")
	})

	t.Run("shows error message", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		result := p.Output(m, errors.New("invalid bid"))
		assert.Contains(t, result, "invalid bid")
	})

	t.Run("shows game end message", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！")
		assert.Contains(t, result, "チーム0の勝ち！")
	})

	t.Run("shows bid phase message", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PinochlePhaseBid)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ビッド番")
	})

	t.Run("shows trump phase message", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PinochlePhaseTrump)
		result := p.Output(m, nil)
		assert.Contains(t, result, "トランプスートを選んで")
	})

	t.Run("shows meld phase message", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PinochlePhaseMeld)
		result := p.Output(m, nil)
		assert.Contains(t, result, "メルドフェーズ")
	})

	t.Run("shows play phase message", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "あなたの番")
	})

	t.Run("shows trick end message", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PinochlePhaseTrickEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
	})

	t.Run("shows round end message", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.PinochlePhaseRoundEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
	})

	t.Run("bidder index -1 hides bidder name", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetHighestBidder")
		m.On("GetHighestBidder").Return(-1)
		result := p.Output(m, nil)
		assert.Contains(t, result, "最高ビッド:")
	})

	t.Run("melds shown in round end phase too", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPlayerMelds")
		m.On("GetPhase").Return(domain.PinochlePhaseRoundEnd)
		melds := [domain.PinochlePlayerCnt][]*domain.PinochleMeld{}
		melds[1] = []*domain.PinochleMeld{{Type: domain.PinochleMeldCommonMarriage, Points: 20}}
		m.On("GetPlayerMelds").Return(melds)
		result := p.Output(m, nil)
		assert.Contains(t, result, "コモンマリッジ")
	})
}

func TestPinochleCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.PinochleCuiPresenter)

	t.Run("nil hint returns no hint message", func(t *testing.T) {
		m := new(interfaces.MockPinochleGame)
		m.On("GetHint").Return((*domain.PinochleHint)(nil))
		result := p.HintOutput(m)
		assert.Equal(t, "ヒントなし", result)
	})

	t.Run("bid amount hint", func(t *testing.T) {
		m := new(interfaces.MockPinochleGame)
		bid := 25
		m.On("GetHint").Return(&domain.PinochleHint{BidAmount: &bid, Reason: "hint_bid"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "ビッド 25")
	})

	t.Run("pass hint", func(t *testing.T) {
		m := new(interfaces.MockPinochleGame)
		pass := true
		m.On("GetHint").Return(&domain.PinochleHint{Pass: &pass, Reason: "hint_pass"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "パス推奨")
	})

	t.Run("suit hint", func(t *testing.T) {
		m := new(interfaces.MockPinochleGame)
		suit := 2
		m.On("GetHint").Return(&domain.PinochleHint{Suit: &suit, Reason: "hint_trump"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "スート")
	})

	t.Run("card index hint", func(t *testing.T) {
		m := new(interfaces.MockPinochleGame)
		idx := 3
		m.On("GetHint").Return(&domain.PinochleHint{CardIndex: &idx, Reason: "hint_play"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "カード 3")
	})
}

func TestPinochleCuiPresenter_ActionLogOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.PinochleCuiPresenter)

	t.Run("nil entries returns no log message", func(t *testing.T) {
		m := new(interfaces.MockPinochleGame)
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜はありません")
	})

	t.Run("with entries returns log", func(t *testing.T) {
		m := new(interfaces.MockPinochleGame)
		entries := []*domain.ActionLogEntry{
			{TurnNumber: 1, PlayerIdx: 0, ActionType: "bid", Detail: "bid 25"},
		}
		m.On("GetGameEndFlag").Return(true)
		m.On("GetActionLog").Return(entries)
		result := p.ActionLogOutput(m)
		assert.Contains(t, result, "棋譜")
		assert.Contains(t, result, "bid")
	})
}

// TestPinochleCuiPresenter_English verifies issue #1699 Phase 2: every
// previously-hardcoded Japanese string in PinochleCuiPresenter now follows
// the active locale. The default ja path is exercised by the assertions
// above; this suite re-runs the hint API and the player-line builder
// under LANG=en and checks the English keys win out.
func TestPinochleCuiPresenter_English(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	p := new(presenter.PinochleCuiPresenter)

	t.Run("hint output uses English labels", func(t *testing.T) {
		m := new(interfaces.MockPinochleGame)
		m.On("GetHint").Return((*domain.PinochleHint)(nil))
		assert.Equal(t, "No hint", p.HintOutput(m))
	})

	t.Run("hint suit uses English label", func(t *testing.T) {
		m := new(interfaces.MockPinochleGame)
		suit := 2
		m.On("GetHint").Return(&domain.PinochleHint{Suit: &suit, Reason: "hint_trump"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "Hint:")
		assert.Contains(t, result, "suit")
	})

	t.Run("hint card index uses English label", func(t *testing.T) {
		m := new(interfaces.MockPinochleGame)
		idx := 3
		m.On("GetHint").Return(&domain.PinochleHint{CardIndex: &idx, Reason: "hint_play"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "card 3")
	})

	t.Run("output uses English game-end banner", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "Game over!")
		assert.Contains(t, result, "Team 0 wins")
		assert.NotContains(t, result, "チーム") // no Japanese leakage
	})

	t.Run("output uses English headers, dealer, scores, prompt", func(t *testing.T) {
		m, _ := setupPinochleCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.Contains(t, result, "Round: 1")
		assert.Contains(t, result, "Trick: 1")
		assert.Contains(t, result, "Dealer: You")
		assert.Contains(t, result, "Team 0: 0pt")
		assert.Contains(t, result, "to play")
	})
}

// #5519: ビッドを決める段階で、15種類のメルドが何点なのかを見る場所が
// どちらのUIにも無かった。
func TestPinochleCuiPresenter_MeldTable(t *testing.T) {
	p := new(presenter.PinochleCuiPresenter)

	outputInPhase := func(phase domain.PinochlePhase) string {
		m, _ := setupPinochleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(phase)
		return p.Output(m, nil)
	}

	// 出力から早見表の部分だけを取り出し、"名前 点数" の並びに割る。
	entriesOf := func(out string) []string {
		header := i18n.T("pinochle.meldTableHeader")
		idx := strings.Index(out, header)
		if idx < 0 {
			return nil
		}
		// ヘッダの後ろから、"名前 点数" の形で読める行だけを拾う。
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
		entries := entriesOf(outputInPhase(domain.PinochlePhaseBid))
		table := domain.PinochleMeldTable()
		require.Len(t, entries, len(table))

		for i, e := range table {
			// **点数は domain の表から読む。**ここで数字を書くと、
			// 表示と加点が食い違ってもテストは通ってしまう。
			assert.True(t, strings.HasSuffix(entries[i], " "+strconv.Itoa(e.Points)),
				"entry %d = %q, want it to end with %d points", i, entries[i], e.Points)
			name := strings.TrimSuffix(entries[i], " "+strconv.Itoa(e.Points))
			// 名前が訳されていること。未訳だとキーか "meld#3" がそのまま出る。
			assert.NotContains(t, name, "meld#")
			assert.NotContains(t, name, "pinochle.")
			assert.NotEmpty(t, name)
		}
		// 並びは安い順。1行目がディックス、最後がダブルラン。
		assert.Equal(t, "ディクス 10", entries[0])
		assert.Equal(t, "ダブルラン 1500", entries[len(entries)-1])
	})

	t.Run("is available while choosing melds too", func(t *testing.T) {
		assert.Contains(t, outputInPhase(domain.PinochlePhaseMeld), i18n.T("pinochle.meldTableHeader"))
	})

	// **プレイ中は出さない。**毎トリック15行流れると盤面が読めなくなる。
	t.Run("is not printed once the cards are being played", func(t *testing.T) {
		assert.NotContains(t, outputInPhase(domain.PinochlePhasePlay), i18n.T("pinochle.meldTableHeader"))
	})
}
