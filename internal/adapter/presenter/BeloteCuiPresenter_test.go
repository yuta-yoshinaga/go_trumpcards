package presenter_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupBeloteCuiMock() *interfaces.MockBeloteGame {
	m := new(interfaces.MockBeloteGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.BelotePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetFaceUpCard").Return((*domain.Card)(nil))
	m.On("GetMakerTeam").Return(0)
	m.On("GetMakerPlayerIdx").Return(0)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetRoundPoints", 0).Return(0)
	m.On("GetRoundPoints", 1).Return(0)
	m.On("GetRoundBeloteBonus", 0).Return(0)
	m.On("GetRoundBeloteBonus", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultBeloteConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeBelotePlayers() []*domain.BelotePlayer {
	return []*domain.BelotePlayer{
		domain.NewBelotePlayer(true, 0),
		domain.NewBelotePlayer(false, 1),
		domain.NewBelotePlayer(false, 0),
		domain.NewBelotePlayer(false, 1),
	}
}

func setupBeloteCuiMockWithPlayers() (*interfaces.MockBeloteGame, []*domain.BelotePlayer) {
	m := setupBeloteCuiMock()
	players := makeBelotePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

// **20 点規模のボーナスに気づけない。**Web は専用バッジと読み上げまで用意して
// いるのに、CUI は累計点しか出していなかった (#4913)。
func TestBeloteCuiPresenter_NamesTheBeloteRebeloteBonus(t *testing.T) {
	p := new(presenter.BeloteCuiPresenter)

	build := func(t0, t1 int) *interfaces.MockBeloteGame {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundBeloteBonus")
		m.On("GetRoundBeloteBonus", 0).Return(t0)
		m.On("GetRoundBeloteBonus", 1).Return(t1)
		return m
	}

	out := p.Output(build(domain.BeloteRebeloteBonus, 0), nil)
	assert.Contains(t, out, "ベロート・ルベロート成立")
	assert.Contains(t, out, "チーム0")
	assert.Contains(t, out, "+"+strconv.Itoa(domain.BeloteRebeloteBonus)+"点")
	// 成立していないチームの行は出さない。
	assert.NotContains(t, out, "チーム1 に")

	// どちらも 0 なら行そのものを出さない。
	assert.NotContains(t, p.Output(build(0, 0), nil), "ベロート・ルベロート成立")
}

func TestBeloteCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BeloteCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupBeloteCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Belote (ベロート)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "トリック: 1")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "切り札: SPADE (メイカー: チーム0)")
		assert.Contains(t, result, "[0]SPADE 1")
	})

	t.Run("trump undecided", func(t *testing.T) {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: 未決定")
	})

	t.Run("face-up card", func(t *testing.T) {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetFaceUpCard")
		m.On("GetFaceUpCard").Return(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "表向きカード: HEART 11")
	})

	t.Run("phase: bid pickup", func(t *testing.T) {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BelotePhaseBidPickUp)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ピックアップフェーズ")
	})

	t.Run("phase: call trump", func(t *testing.T) {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BelotePhaseBidCallTrump)

		result := p.Output(m, nil)
		assert.Contains(t, result, "コールトランプフェーズ")
	})

	t.Run("phase: trick end", func(t *testing.T) {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BelotePhaseTrickEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
	})

	t.Run("phase: round end", func(t *testing.T) {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.BelotePhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})
}

func TestBeloteCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.BeloteCuiPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m := setupBeloteCuiMock()
		m.On("GetHint").Return((*domain.BeloteHint)(nil))

		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("order up hint", func(t *testing.T) {
		m := setupBeloteCuiMock()
		ok := true
		m.On("GetHint").Return(&domain.BeloteHint{OrderUp: &ok, Reason: "strategic_pickup"})

		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("pass hint", func(t *testing.T) {
		m := setupBeloteCuiMock()
		ok := false
		m.On("GetHint").Return(&domain.BeloteHint{OrderUp: &ok, Reason: "pass_recommended"})

		result := p.HintOutput(m)
		assert.Contains(t, result, "パス")
	})

	t.Run("call suit hint", func(t *testing.T) {
		m := setupBeloteCuiMock()
		suit := 2
		m.On("GetHint").Return(&domain.BeloteHint{Suit: &suit, Reason: "strategic_call"})

		result := p.HintOutput(m)
		assert.Contains(t, result, "コール")
	})

	t.Run("card hint", func(t *testing.T) {
		m, players := setupBeloteCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		idx := 0
		m.On("GetHint").Return(&domain.BeloteHint{CardIndex: &idx, Reason: "trump_cut"})

		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestBeloteCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupBeloteCuiMock()
	p := new(presenter.BeloteCuiPresenter)
	result := p.ActionLogOutput(m)
	// even empty log should not panic; result should be a string
	assert.NotNil(t, result)
}

// #5592: 最終トリックが特別だということを、CUI は点数計算を見るまで教えて
// いなかった。Web はバッジを点滅させている。
func TestBeloteCuiPresenter_AnnouncesTheDixDeDerOnTheLastTrick(t *testing.T) {
	i18n.SetLang("ja")
	build := func(trick, dixDeDer int) string {
		m, _ := setupBeloteCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrickNumber")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetConfig")
		cfg := domain.DefaultBeloteConfig()
		cfg.DixDeDer = dixDeDer
		m.On("GetTrickNumber").Return(trick)
		m.On("GetConfig").Return(cfg)
		return new(presenter.BeloteCuiPresenter).Output(m, nil)
	}

	notice := func(points int) string {
		return i18n.Tf("belote.dixDeDerNotice", "points", strconv.Itoa(points))
	}

	// 8 トリック目 = 最終。**プレイ前に**出る。
	assert.Contains(t, build(domain.BeloteHandSize, 10), notice(10))
	// **点数は設定から。**訳文に 10 と書くと、設定を変えたとき案内だけが嘘になる。
	assert.Contains(t, build(domain.BeloteHandSize, 25), notice(25))

	// 最終トリック以外では出さない。
	for _, trick := range []int{1, domain.BeloteHandSize - 1} {
		assert.NotContains(t, build(trick, 10), notice(10), "trick %d", trick)
	}

	// 0 に設定されていれば出さない (受け入れ条件2) ── ボーナスが無いので。
	assert.NotContains(t, build(domain.BeloteHandSize, 0),
		strings.SplitN(i18n.T("belote.dixDeDerNotice"), "{{", 2)[0])
}

// TestBeloteCuiPresenter_JapanesePromptsAreTranslated は ja ロケールで
// 操作案内が英語のまま出ないことを見る。
//
// **キーの有無でなく、描かれた行を見る。** ja のロケールに値があっても
// 英語のままなら、日本語でプレイしている人には英語が出る (#6388)。
// 札の表記 (SPADE 1 など) は cuiSuitName の全ゲーム共通の規約なので対象外。
func TestBeloteCuiPresenter_JapanesePromptsAreTranslated(t *testing.T) {
	old := i18n.Lang()
	i18n.SetLang("ja")
	defer i18n.SetLang(old)

	p := new(presenter.BeloteCuiPresenter)
	for _, tt := range []struct {
		phase domain.BelotePhase
		want  string
	}{
		{domain.BelotePhaseBidPickUp, "表向きの札のスートを切り札にする"},
		{domain.BelotePhaseBidCallTrump, "切り札を宣言"},
		{domain.BelotePhasePlay, "カードを出す"},
		{domain.BelotePhaseTrickEnd, "次のトリックへ"},
		{domain.BelotePhaseRoundEnd, "次のラウンドへ"},
	} {
		g := domain.NewDefaultBelote()
		g.Reset()
		g.SetPhase(tt.phase)
		out := p.Output(g, nil)
		assert.Contains(t, out, tt.want, "phase %d の案内が日本語になっていない", tt.phase)
		// コマンド構文は残す。訳しただけで打ち方が消えては案内にならない。
		assert.Contains(t, out, "・・・", "phase %d でコマンド構文と説明の区切りが無い", tt.phase)
	}
}
