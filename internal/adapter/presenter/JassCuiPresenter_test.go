package presenter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
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
