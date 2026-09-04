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

// napBidHolderPrefix returns the localized high-bidder line's text up to the
// name placeholder, so assertions match regardless of the substituted name.
func napBidHolderPrefix() string {
	return strings.Split(i18n.T("nap.promptBidHolder"), "{{")[0]
}

func makeNapPlayers() []*domain.NapPlayer {
	return []*domain.NapPlayer{
		domain.NewNapPlayer(true),
		domain.NewNapPlayer(false),
		domain.NewNapPlayer(false),
		domain.NewNapPlayer(false),
	}
}

func setupNapCuiMock() *interfaces.MockNapGame {
	m := new(interfaces.MockNapGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.NapPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetDeclarerIdx").Return(0)
	m.On("GetDeclarerProgress").Return((*domain.NapDeclarerProgress)(nil)).Maybe()
	m.On("GetContract").Return(domain.NapBidThree)
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetPlayerScores").Return([domain.NapPlayerCnt]int{0, 0, 0, 0})
	m.On("GetRoundTricks").Return([domain.NapPlayerCnt]int{0, 0, 0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupNapCuiMockWithPlayers() (*interfaces.MockNapGame, []*domain.NapPlayer) {
	m := setupNapCuiMock()
	players := makeNapPlayers()
	m.On("GetPlayerCnt").Return(4)
	for i := 0; i < 4; i++ {
		m.On("GetPlayer", i).Return(players[i])
	}
	return m, players
}

func TestNapCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.NapCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupNapCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Nap")
		assert.NotEmpty(t, result)
	})

	t.Run("bid phase prompt", func(t *testing.T) {
		m, _ := setupNapCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDeclarerIdx")
		m.On("GetPhase").Return(domain.NapPhaseBid)
		m.On("GetDeclarerIdx").Return(-1)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
		// No one has bid yet → the high-bidder line is omitted.
		assert.NotContains(t, result, napBidHolderPrefix())
	})

	t.Run("bid phase names the current high bidder", func(t *testing.T) {
		m, _ := setupNapCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NapPhaseBid)
		// GetDeclarerIdx defaults to 0 (a bidder exists) → holder line is shown.
		result := p.Output(m, nil)
		assert.Contains(t, result, napBidHolderPrefix())
	})

	t.Run("no trump during bid", func(t *testing.T) {
		m, _ := setupNapCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupNapCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NapPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupNapCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.NapPhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupNapCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerPlayer")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerPlayer").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupNapCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestNapCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.NapCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupNapCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.NapHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupNapCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.NapHint{CardIndices: []int{0}, Reason: "lead_high"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupNapCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.NapHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestNapCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.NapCuiPresenter)
	m := new(interfaces.MockNapGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewNapPlayer(true)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}

// **CUI は宣言者が何トリック取ったかを一切知らせていなかった (#4763)。**
// CLI プレイヤーは自分でトリック数を数えるしかなかった。
func TestNapCuiPresenter_DeclarerProgress(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.NapCuiPresenter)

	withProgress := func(pr *domain.NapDeclarerProgress, phase domain.NapPhase) *interfaces.MockNapGame {
		m, _ := setupNapCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetDeclarerProgress")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetDeclarerProgress").Return(pr)
		m.On("GetPhase").Return(phase)
		return m
	}

	t.Run("shows tricks won, needed and remaining", func(t *testing.T) {
		out := p.Output(withProgress(&domain.NapDeclarerProgress{
			Won: 1, Needed: 3, Remaining: 3,
		}, domain.NapPhasePlay), nil)
		assert.Contains(t, out, "1/3")
		assert.Contains(t, out, "残り3")
	})

	// **もう届かないなら押す意味が変わる。**同じ文言だと区別が付かない。
	t.Run("calls out a contract that can no longer be made", func(t *testing.T) {
		out := p.Output(withProgress(&domain.NapDeclarerProgress{
			Won: 0, Needed: 5, Remaining: 4, Unreachable: true,
		}, domain.NapPhasePlay), nil)
		assert.Contains(t, out, "達成不可能")
	})

	t.Run("does not call a reachable contract unreachable", func(t *testing.T) {
		out := p.Output(withProgress(&domain.NapDeclarerProgress{
			Won: 1, Needed: 3, Remaining: 3,
		}, domain.NapPhasePlay), nil)
		assert.NotContains(t, out, "達成不可能")
	})

	t.Run("shows the progress at trick end too", func(t *testing.T) {
		out := p.Output(withProgress(&domain.NapDeclarerProgress{
			Won: 2, Needed: 3, Remaining: 2,
		}, domain.NapPhaseTrickEnd), nil)
		assert.Contains(t, out, "2/3")
	})

	t.Run("shows nothing when there is no declarer yet", func(t *testing.T) {
		// 既存の「宣言者: あなた — スリー」行とは別物なので、進捗行だけを狙う。
		assert.NotContains(t, p.Output(withProgress(nil, domain.NapPhasePlay), nil), "トリック (残り")
	})
}

// **チップの授受が CUI に一切出ていなかった。**Web は nap-round-payout で出している
// のに、CUI はスコアの推移を見比べて何枚動いたのか推測するしかなかった (#6447)。
func TestNapCuiPresenter_ShowsTheChipSettlement(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.NapCuiPresenter)

	roundEnd := func(contract domain.NapBid, declarerTricks int) *interfaces.MockNapGame {
		m, _ := setupNapCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetContract")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundTricks")
		m.On("GetPhase").Return(domain.NapPhaseRoundEnd)
		m.On("GetContract").Return(contract)
		m.On("GetRoundTricks").Return([domain.NapPlayerCnt]int{declarerTricks, 0, 0, 0})
		return m
	}

	// **Nap 契約だけ非対称** (達成 +10 / 失敗は相手が各 +5)。数字は
	// domain.NapBidPayout から引くので、表と実際の増減がずれない。
	t.Run("a made Nap pays the declarer ten", func(t *testing.T) {
		makeValue, _ := domain.NapBidPayout(domain.NapBidNap)
		out := p.Output(roundEnd(domain.NapBidNap, 5), nil)
		assert.Contains(t, out, i18n.Tf("nap.payoutMade",
			"name", color.Bold(i18n.T("cuiPlayerYou")), "chips", strconv.Itoa(makeValue)))
		assert.Equal(t, 10, makeValue, "Nap は達成 +10 の非対称契約")
		assert.NotContains(t, out, "{{")
	})

	t.Run("a failed Nap pays each opponent five", func(t *testing.T) {
		_, failValue := domain.NapBidPayout(domain.NapBidNap)
		out := p.Output(roundEnd(domain.NapBidNap, 4), nil)
		assert.Contains(t, out, i18n.Tf("nap.payoutFailed", "chips", strconv.Itoa(failValue)))
		assert.Equal(t, 5, failValue, "Nap の失敗は相手が各 +5")
		assert.NotContains(t, out, "{{")
	})

	// ほかの契約は達成・失敗とも契約数と同じ数が動く。
	t.Run("other contracts move their own trick count", func(t *testing.T) {
		makeValue, failValue := domain.NapBidPayout(domain.NapBidThree)
		assert.Equal(t, failValue, makeValue, "Nap 以外は対称")
		assert.Contains(t, p.Output(roundEnd(domain.NapBidThree, 3), nil),
			i18n.Tf("nap.payoutMade", "name", color.Bold(i18n.T("cuiPlayerYou")), "chips", strconv.Itoa(makeValue)))
	})

	// パスの局は何も賭けていないので、行ごと出さない。
	//
	// **`i18n.T` を否定に使わない。**プレースホルダを持つキーに `T` を当てると
	// 生のテンプレート (`{{chips}} チップ...`) が返り、置換済みの出力には決して
	// 含まれないので、その否定は何があっても通る。ガードを外すと実際には
	// 「0 チップ獲得（契約達成）」が出るので、**その解決済みの文字列**を見る。
	t.Run("says nothing when everyone passed", func(t *testing.T) {
		out := p.Output(roundEnd(domain.NapBidPass, 0), nil)
		assert.NotContains(t, out, i18n.Tf("nap.payoutMade",
			"name", color.Bold(i18n.T("cuiPlayerYou")), "chips", "0"))
		assert.NotContains(t, out, i18n.Tf("nap.payoutFailed", "chips", "0"))
		// 精算行の骨格そのものが無いことも見る (数字に依らない)。
		_, tail, ok := strings.Cut(i18n.Tf("nap.payoutMade", "name", "\x00", "chips", "\x00"), "\x00")
		require.True(t, ok)
		suffix := tail[strings.LastIndex(tail, "\x00")+1:]
		require.NotEmpty(t, suffix)
		assert.NotContains(t, out, suffix)
	})

	// ラウンド終了以外では出さない。
	t.Run("says nothing mid-round", func(t *testing.T) {
		m, _ := setupNapCuiMockWithPlayers()
		out := p.Output(m, nil) // 既定は Play フェーズ
		assert.NotContains(t, out, i18n.Tf("nap.payoutFailed", "chips", "3"))
	})
}
