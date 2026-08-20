package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupJassCuiMock() *interfaces.MockJassGame {
	m := new(interfaces.MockJassGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.JassPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetBidPlayerIdx").Return(0)
	m.On("GetDealerIdx").Return(0)
	m.On("GetTrumpSuit").Return(1)
	m.On("GetMakerTeam").Return(0)
	m.On("GetMakerPlayerIdx").Return(0)
	m.On("GetTeamScore", 0).Return(0)
	m.On("GetTeamScore", 1).Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetLeadPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultJassConfig())
	m.On("GetRoundWeisPoints", 0).Return(0)
	m.On("GetRoundWeisPoints", 1).Return(0)
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func makeJassPlayers() []*domain.JassPlayer {
	return []*domain.JassPlayer{
		domain.NewJassPlayer(true, 0),
		domain.NewJassPlayer(false, 1),
		domain.NewJassPlayer(false, 0),
		domain.NewJassPlayer(false, 1),
	}
}

func setupJassCuiMockWithPlayers() (*interfaces.MockJassGame, []*domain.JassPlayer) {
	m := setupJassCuiMock()
	players := makeJassPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

// **Weis の内訳が無いとラウンド得点が説明できない。**Web は専用パネルを出すのに、
// CUI は Weis で加点があったことすら伝えていなかった (#4918)。
func TestJassCuiPresenter_ShowsTheRoundWeisPoints(t *testing.T) {
	p := new(presenter.JassCuiPresenter)

	weisMock := func(w0, w1 int, enable bool) *interfaces.MockJassGame {
		m, _ := setupJassCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundWeisPoints")
		m.On("GetRoundWeisPoints", 0).Return(w0)
		m.On("GetRoundWeisPoints", 1).Return(w1)
		if !enable {
			cfg := domain.DefaultJassConfig()
			cfg.EnableWeis = false
			m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetConfig")
			m.On("GetConfig").Return(cfg)
		}
		return m
	}

	t.Run("shows both teams so the zero side is explained too", func(t *testing.T) {
		out := p.Output(weisMock(150, 0, true), nil)
		assert.Contains(t, out, "当ラウンドの Weis:")
		assert.Contains(t, out, "チーム0 150点")
		assert.Contains(t, out, "チーム1 0点")
		// **総取りの規則を書かないと、0 点の側が理由不明になる。**
		assert.Contains(t, out, "総取り")
	})

	// 誰も Weis を宣言していないラウンドでは行そのものを出さない。
	t.Run("no line when neither team declared anything", func(t *testing.T) {
		assert.NotContains(t, p.Output(weisMock(0, 0, true), nil), "当ラウンドの Weis")
	})

	// enableWeis が無効なら、点があっても出さない (受け入れ条件2)。
	t.Run("no line when weis is disabled", func(t *testing.T) {
		assert.NotContains(t, p.Output(weisMock(150, 0, false), nil), "当ラウンドの Weis")
	})
}

func TestJassCuiPresenter_Output(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.JassCuiPresenter)

	t.Run("initial state", func(t *testing.T) {
		m, players := setupJassCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 1, false))
		players[0].AddCard(domain.NewCard(domain.CardDesignHeart, 11, false))

		result := p.Output(m, nil)
		assert.Contains(t, result, "Jass")
		assert.Contains(t, result, "ラウンド: 1")
		assert.Contains(t, result, "切り札: SPADE")
		assert.Contains(t, result, "[0]SPADE 1")
	})

	t.Run("trump undecided", func(t *testing.T) {
		m, _ := setupJassCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: 未決定")
	})

	t.Run("phase: bid trump", func(t *testing.T) {
		m, _ := setupJassCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.JassPhaseBidTrump)

		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札ビッド")
	})

	t.Run("phase: bid partner", func(t *testing.T) {
		m, _ := setupJassCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.JassPhaseBidPartner)

		result := p.Output(m, nil)
		assert.Contains(t, result, "パートナー")
	})

	t.Run("phase: trick end", func(t *testing.T) {
		m, _ := setupJassCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.JassPhaseTrickEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
	})

	t.Run("phase: round end", func(t *testing.T) {
		m, _ := setupJassCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.JassPhaseRoundEnd)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
	})

	t.Run("game end", func(t *testing.T) {
		m, _ := setupJassCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)

		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了")
	})
}

func TestJassCuiPresenter_HintOutput(t *testing.T) {
	origNoColor := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(origNoColor)
	p := new(presenter.JassCuiPresenter)

	t.Run("nil hint", func(t *testing.T) {
		m := setupJassCuiMock()
		m.On("GetHint").Return((*domain.JassHint)(nil))
		assert.Contains(t, p.HintOutput(m), "ヒントはありません")
	})

	t.Run("schieben hint", func(t *testing.T) {
		m := setupJassCuiMock()
		sb := true
		m.On("GetHint").Return(&domain.JassHint{Schieben: &sb, Reason: "schieben_recommended"})
		assert.Contains(t, p.HintOutput(m), "シーベン")
	})

	t.Run("trump hint", func(t *testing.T) {
		m := setupJassCuiMock()
		suit := 2
		m.On("GetHint").Return(&domain.JassHint{Suit: &suit, Reason: "strategic_trump"})
		assert.Contains(t, p.HintOutput(m), "切り札")
	})

	t.Run("card hint", func(t *testing.T) {
		m, players := setupJassCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 11, false))
		idx := 0
		m.On("GetHint").Return(&domain.JassHint{CardIndex: &idx, Reason: "trump_cut"})
		assert.Contains(t, p.HintOutput(m), "HINT")
	})
}

func TestJassCuiPresenter_ActionLogOutput(t *testing.T) {
	m := setupJassCuiMock()
	p := new(presenter.JassCuiPresenter)
	assert.NotNil(t, p.ActionLogOutput(m))
}

// #5685: Web は ja-previous-trick で直前トリックの4枚と勝者を常に振り返れるのに、
// CUI は現在のトリックしか出さず、**生の棋譜を自分でスクロールして探すしかなかった。**
func TestJassCuiPresenter_ShowsThePreviousTrick(t *testing.T) {
	p := new(presenter.JassCuiPresenter)
	card := func(d, v int) *domain.Card { return domain.NewCard(d, v, false) }

	// 直前トリック: 席0 ♠A → 席1 ♠9 → 席2 ♠7 → 席3 ♠8、勝者は席0。
	trickLog := []*domain.ActionLogEntry{
		{PlayerIdx: 0, ActionType: "play", Cards: []*domain.Card{card(domain.CardDesignSpade, 1)}},
		{PlayerIdx: 1, ActionType: "play", Cards: []*domain.Card{card(domain.CardDesignSpade, 9)}},
		{PlayerIdx: 2, ActionType: "play", Cards: []*domain.Card{card(domain.CardDesignSpade, 7)}},
		{PlayerIdx: 3, ActionType: "play", Cards: []*domain.Card{card(domain.CardDesignSpade, 8)}},
		{PlayerIdx: 0, ActionType: "trick_win"},
	}

	build := func(trickNo int, log []*domain.ActionLogEntry) *interfaces.MockJassGame {
		m, _ := setupJassCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrickNumber")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetActionLog")
		m.On("GetTrickNumber").Return(trickNo)
		m.On("GetActionLog").Return(log)
		return m
	}

	t.Run("lists the four cards and the winner", func(t *testing.T) {
		out := p.Output(build(2, trickLog), nil)

		assert.Contains(t, out, i18n.T("jass.previousTrick"))
		assert.Contains(t, out, "SPADE 1")
		assert.Contains(t, out, "SPADE 9")
		assert.Contains(t, out, i18n.Tf("jass.previousTrickWinner",
			"name", color.Bold(i18n.T("cuiPlayerYou"))))
	})

	// **直近のトリックを取る。**棋譜はゲーム全体で累積するので、先頭から4枚を
	// 取ると 1 トリック目をいつまでも表示し続ける。
	t.Run("takes the most recent trick, not the first", func(t *testing.T) {
		twoTricks := append(append([]*domain.ActionLogEntry(nil), trickLog...),
			&domain.ActionLogEntry{PlayerIdx: 0, ActionType: "play", Cards: []*domain.Card{card(domain.CardDesignHeart, 13)}},
			&domain.ActionLogEntry{PlayerIdx: 1, ActionType: "play", Cards: []*domain.Card{card(domain.CardDesignHeart, 12)}},
			&domain.ActionLogEntry{PlayerIdx: 2, ActionType: "play", Cards: []*domain.Card{card(domain.CardDesignHeart, 11)}},
			&domain.ActionLogEntry{PlayerIdx: 3, ActionType: "play", Cards: []*domain.Card{card(domain.CardDesignHeart, 10)}},
			&domain.ActionLogEntry{PlayerIdx: 1, ActionType: "trick_win"},
		)

		out := p.Output(build(3, twoTricks), nil)

		assert.Contains(t, out, color.Red("HEART 13"), "2 トリック目の札が出る")
		assert.NotContains(t, out, "SPADE 1", "1 トリック目の札は出ない")
		assert.Contains(t, out, i18n.Tf("jass.previousTrickWinner",
			"name", color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "1"))))
	})

	// **ラウンド最初のトリック中はまだ確定済みトリックが無い** (受け入れ条件2)。
	t.Run("says nothing on the first trick of a round", func(t *testing.T) {
		out := p.Output(build(1, trickLog), nil)

		assert.NotContains(t, out, i18n.T("jass.previousTrick"))
	})

	t.Run("says nothing when the log has no resolved trick", func(t *testing.T) {
		out := p.Output(build(2, nil), nil)

		assert.NotContains(t, out, i18n.T("jass.previousTrick"))
	})
}
