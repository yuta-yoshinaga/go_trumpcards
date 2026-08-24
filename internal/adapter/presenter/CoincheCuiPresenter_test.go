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

func setupCoincheCuiMock() *interfaces.MockCoincheGame {
	m := new(interfaces.MockCoincheGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.CoinchePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetContractPoints").Return(0)
	m.On("GetMultiplier").Return(1)
	m.On("GetDouble").Return(domain.CoincheDoubleNone)
	m.On("GetBiddablePoints").Return(domain.CoincheContractPoints)
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
	m.On("GetConfig").Return(domain.DefaultCoincheConfig())
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeCoinchePlayers() []*domain.CoinchePlayer {
	return []*domain.CoinchePlayer{
		domain.NewCoinchePlayer(true, 0),
		domain.NewCoinchePlayer(false, 1),
		domain.NewCoinchePlayer(false, 0),
		domain.NewCoinchePlayer(false, 1),
	}
}

func setupCoincheCuiMockWithPlayers() (*interfaces.MockCoincheGame, []*domain.CoinchePlayer) {
	m := setupCoincheCuiMock()
	players := makeCoinchePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

// **20 点規模のボーナスに気づけない。**Web は専用バッジと読み上げまで用意して
// いるのに、CUI は累計点しか出していなかった (#4913)。
func TestCoincheCuiPresenter_NamesTheBeloteRebeloteBonus(t *testing.T) {
	p := new(presenter.CoincheCuiPresenter)

	build := func(t0, t1 int) *interfaces.MockCoincheGame {
		m, _ := setupCoincheCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundBeloteBonus")
		m.On("GetRoundBeloteBonus", 0).Return(t0)
		m.On("GetRoundBeloteBonus", 1).Return(t1)
		return m
	}

	out := p.Output(build(domain.CoincheRebeloteBonus, 0), nil)
	assert.Contains(t, out, "ベロート・ルベロート成立")
	assert.Contains(t, out, "チーム0")
	assert.Contains(t, out, "+"+strconv.Itoa(domain.CoincheRebeloteBonus)+"点")
	// 成立していないチームの行は出さない。
	assert.NotContains(t, out, "チーム1 に")

	// どちらも 0 なら行そのものを出さない。
	assert.NotContains(t, p.Output(build(0, 0), nil), "ベロート・ルベロート成立")
}

func TestCoincheCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CoincheCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupCoincheCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Coinche (コワンシュ)")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "トリック: 1")
		assert.Contains(t, result, "あなた")
		assert.Contains(t, result, "切り札: SPADE (宣言: チーム0)")
		assert.Contains(t, result, "[0]SPADE 1")
	})

	t.Run("trump undecided", func(t *testing.T) {
		m, _ := setupCoincheCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: 未決定")
	})

	// **契約と倍率は精算そのもの。** 出さないと、同じカード点でも勝ち負けが
	// 変わる理由が盤面から読めない。
	t.Run("contract line carries the points and the multiplier", func(t *testing.T) {
		m, _ := setupCoincheCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetContractPoints")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMultiplier")
		m.On("GetContractPoints").Return(120)
		m.On("GetMultiplier").Return(2)

		result := p.Output(m, nil)
		assert.Contains(t, result, "120")
		assert.Contains(t, result, "x2")
	})

	t.Run("no contract line before the auction settles", func(t *testing.T) {
		m, _ := setupCoincheCuiMockWithPlayers()
		result := p.Output(m, nil)
		assert.NotContains(t, result, "契約:")
	})

	t.Run("phase: bid", func(t *testing.T) {
		m, _ := setupCoincheCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CoinchePhaseBid)

		result := p.Output(m, nil)
		assert.Contains(t, result, "競り")
		// **打てる点だけを案内する。** 全部並べると、打てば必ず拒否される
		// 値を勧めることになる。
		assert.Contains(t, result, "80")
	})

	t.Run("phase: double", func(t *testing.T) {
		m, _ := setupCoincheCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CoinchePhaseDouble)

		result := p.Output(m, nil)
		assert.Contains(t, result, "コワンシュ")
	})

	t.Run("phase: trick end", func(t *testing.T) {
		m, _ := setupCoincheCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CoinchePhaseTrickEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
	})

	t.Run("phase: round end", func(t *testing.T) {
		m, _ := setupCoincheCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.CoinchePhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupCoincheCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})
}

func TestCoincheCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.CoincheCuiPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m := setupCoincheCuiMock()
		m.On("GetHint").Return((*domain.CoincheHint)(nil))

		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません")
	})

	t.Run("bid hint names both the points and the suit", func(t *testing.T) {
		m := setupCoincheCuiMock()
		points, suit := 110, domain.CardDesignSpade
		m.On("GetHint").Return(&domain.CoincheHint{Bid: &points, Suit: &suit, Reason: "strategic_bid"})

		result := p.HintOutput(m)
		assert.Contains(t, result, "110")
		// スートを落とすと、点だけ言って何で取るのか言わない助言になる。
		assert.Contains(t, result, "SPADE")
	})

	t.Run("pass hint", func(t *testing.T) {
		m := setupCoincheCuiMock()
		m.On("GetHint").Return(&domain.CoincheHint{Reason: "pass_recommended"})

		result := p.HintOutput(m)
		assert.Contains(t, result, "パス")
	})

	t.Run("card hint", func(t *testing.T) {
		m, players := setupCoincheCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		idx := 0
		m.On("GetHint").Return(&domain.CoincheHint{CardIndex: &idx, Reason: "trump_cut"})

		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestCoincheCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupCoincheCuiMock()
	p := new(presenter.CoincheCuiPresenter)
	result := p.ActionLogOutput(m)
	// even empty log should not panic; result should be a string
	assert.NotNil(t, result)
}

// #5592: 最終トリックが特別だということを、CUI は点数計算を見るまで教えて
// いなかった。Web はバッジを点滅させている。
func TestCoincheCuiPresenter_AnnouncesTheDixDeDerOnTheLastTrick(t *testing.T) {
	i18n.SetLang("ja")
	build := func(trick, dixDeDer int) string {
		m, _ := setupCoincheCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrickNumber")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetConfig")
		cfg := domain.DefaultCoincheConfig()
		cfg.DixDeDer = dixDeDer
		m.On("GetTrickNumber").Return(trick)
		m.On("GetConfig").Return(cfg)
		return new(presenter.CoincheCuiPresenter).Output(m, nil)
	}

	notice := func(points int) string {
		return i18n.Tf("coinche.dixDeDerNotice", "points", strconv.Itoa(points))
	}

	// 8 トリック目 = 最終。**プレイ前に**出る。
	assert.Contains(t, build(domain.CoincheHandSize, 10), notice(10))
	// **点数は設定から。**訳文に 10 と書くと、設定を変えたとき案内だけが嘘になる。
	assert.Contains(t, build(domain.CoincheHandSize, 25), notice(25))

	// 最終トリック以外では出さない。
	for _, trick := range []int{1, domain.CoincheHandSize - 1} {
		assert.NotContains(t, build(trick, 10), notice(10), "trick %d", trick)
	}

	// 0 に設定されていれば出さない (受け入れ条件2) ── ボーナスが無いので。
	assert.NotContains(t, build(domain.CoincheHandSize, 0),
		strings.SplitN(i18n.T("coinche.dixDeDerNotice"), "{{", 2)[0])
}
