//go:build test

package presenter_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func makeMarjapussiCuiPlayers() []*domain.MarjapussiPlayer {
	return []*domain.MarjapussiPlayer{
		domain.NewMarjapussiPlayer(true),
		domain.NewMarjapussiPlayer(false),
		domain.NewMarjapussiPlayer(false),
		domain.NewMarjapussiPlayer(false),
	}
}

func setupMarjapussiCuiMock() *interfaces.MockMarjapussiGame {
	m := new(interfaces.MockMarjapussiGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.MarjapussiPhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetConfig").Return(domain.DefaultMarjapussiConfig())
	m.On("GetWinnerPlayer").Return(-1)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetRoundCardPoints").Return([domain.MarjapussiTeamCnt]int{0, 0})
	m.On("GetRoundMarriage").Return([domain.MarjapussiTeamCnt]int{0, 0})
	m.On("GetTeamScores").Return([domain.MarjapussiTeamCnt]int{0, 0})
	m.On("GetPlayerScores").Return([domain.MarjapussiPlayerCnt]int{0, 0, 0, 0})
	m.On("GetMarriageOptions", mock.Anything).Return(([]domain.MarjapussiMarriageOption)(nil))
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupMarjapussiCuiMockWithPlayers() (*interfaces.MockMarjapussiGame, []*domain.MarjapussiPlayer) {
	m := setupMarjapussiCuiMock()
	players := makeMarjapussiCuiPlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestMarjapussiCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	origLang := i18n.Lang()
	defer i18n.SetLang(origLang)
	i18n.SetLang("ja")
	p := new(presenter.MarjapussiCuiPresenter)

	t.Run("play phase shows current player and header info", func(t *testing.T) {
		m, players := setupMarjapussiCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Marjapussi (マルヤプッシ)")
		assert.Contains(t, result, "Marjapussi")
		assert.Contains(t, result, "マリッジ: 同スートの K と Q を持ちリードすると切り札になり、現在の切り札なら40点、違えば20点（新しい切り札になる）")
		assert.Contains(t, result, "切り札: ♠")
		assert.Contains(t, result, "目標: 500点  チーム0: 0点  チーム1: 0点")
		assert.Contains(t, result, "直近のマリッジ: なし")
		assert.Contains(t, result, "チーム0")
		assert.Contains(t, result, "チーム1")
	})

	t.Run("play phase output in english without partner mistranslation", func(t *testing.T) {
		i18n.SetLang("en")
		defer i18n.SetLang("ja")

		m, players := setupMarjapussiCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Marjapussi")
		assert.Contains(t, result, "Marriage: hold both the K and Q of the same suit and lead one of them to set trump (40 pts if current trump, 20 pts if new trump)")
		assert.NotContains(t, result, "partner")
	})

	t.Run("undetermined trump shows none literal", func(t *testing.T) {
		m, _ := setupMarjapussiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTrumpSuit")
		m.On("GetTrumpSuit").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "切り札: 未決定")
	})

	t.Run("recent marriage displayed in header", func(t *testing.T) {
		m, _ := setupMarjapussiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetActionLog")
		m.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{ActionType: "marriage", Detail: "Player declares a Spades marriage (+40, trump=Spades)"},
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "直近のマリッジ: Player declares a Spades marriage (+40, trump=Spades)")
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupMarjapussiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MarjapussiPhaseTrickEnd)
		result := p.Output(m, nil)
		assert.Contains(t, result, "トリック終了")
		assert.Contains(t, result, "n / next")
	})

	t.Run("round end prompt with pussi win", func(t *testing.T) {
		m, _ := setupMarjapussiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.MarjapussiPhaseRoundEnd)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetActionLog")
		m.On("GetActionLog").Return([]*domain.ActionLogEntry{
			{
				ActionType: "pussi_win",
				PlayerIdx:  0,
				Cards: []*domain.Card{
					domain.NewCard(domain.CardDesignSpade, 1, false),  // 11
					domain.NewCard(domain.CardDesignHeart, 10, false), // 10
				},
			},
		})
		result := p.Output(m, nil)
		assert.Contains(t, result, "ラウンド終了")
		assert.Contains(t, result, "ベリー袋 (pussi): チーム0が獲得 (+21点)")
		assert.Contains(t, result, "nr / nextround")
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupMarjapussiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.Contains(t, result, "ゲーム終了！チーム0がマッチ勝利！")
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupMarjapussiCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestMarjapussiCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.MarjapussiCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupMarjapussiCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.MarjapussiHint)(nil))
		result := p.HintOutput(m)
		assert.Contains(t, result, "ヒントはありません。")
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupMarjapussiCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.MarjapussiHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "低い札でリードして温存する")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupMarjapussiCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.MarjapussiHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "得点があるので勝ちに行く")
	})
}

func TestMarjapussiCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.MarjapussiCuiPresenter)
	m := new(interfaces.MockMarjapussiGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	m.On("GetPlayer", mock.Anything).Return(domain.NewMarjapussiPlayer(true)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}

func TestMarjapussiCuiPresenter_MarriageOptions(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.MarjapussiCuiPresenter)

	withOptions := func(idx int, opts ...domain.MarjapussiMarriageOption) *interfaces.MockMarjapussiGame {
		m, _ := setupMarjapussiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetCurrentPlayerIdx")
		m.On("GetCurrentPlayerIdx").Return(idx)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetMarriageOptions")
		m.On("GetMarriageOptions", mock.Anything).Return(opts)
		return m
	}
	spadeAndClub := []domain.MarjapussiMarriageOption{
		{Suit: domain.CardDesignSpade, Points: 40},
		{Suit: domain.CardDesignClover, Points: 20},
	}

	t.Run("names every declarable suit on the human turn", func(t *testing.T) {
		result := p.Output(withOptions(0, spadeAndClub...), nil)
		assert.Contains(t, result, "いま宣言できるマリッジ: SPADE K-Q (+40), CLOVER K-Q (+20)")
	})

	t.Run("adds nothing when no suit is paired", func(t *testing.T) {
		result := p.Output(withOptions(0), nil)
		assert.NotContains(t, result, "いま宣言できるマリッジ")
	})

	t.Run("never leaks a CPU hand", func(t *testing.T) {
		result := p.Output(withOptions(1, spadeAndClub...), nil)
		assert.NotContains(t, result, "いま宣言できるマリッジ")
	})
}

func TestMarjapussiCuiPresenter_HighlightsThePlayerNearTheTarget(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(false)
	defer color.SetNoColor(orig)
	assert.NotEqual(t, "x", color.Yellow("x"), "colour must be enabled for this test to measure anything")

	p := new(presenter.MarjapussiCuiPresenter)
	target := domain.DefaultMarjapussiConfig().TargetPoints

	withScores := func(scores [domain.MarjapussiTeamCnt]int) string {
		m, _ := setupMarjapussiCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetTeamScores")
		m.On("GetTeamScores").Return(scores)
		return p.Output(m, nil)
	}

	humanLine := func(score int) string {
		return i18n.Tf("marjapussi.playerLine",
			"name", color.Bold(i18n.T("cuiPlayerYou")), "team", "0",
			"cards", "0", "score", strconv.Itoa(score), "tricks", "0")
	}

	justOver := int(float64(target)*0.8) + 1
	justUnder := int(float64(target) * 0.8)

	t.Run("marks the seat past the threshold", func(t *testing.T) {
		out := withScores([domain.MarjapussiTeamCnt]int{justOver, 0})
		assert.Contains(t, out, color.Yellow(humanLine(justOver)))
	})

	t.Run("leaves a seat at the threshold alone", func(t *testing.T) {
		out := withScores([domain.MarjapussiTeamCnt]int{justUnder, 0})
		assert.NotContains(t, out, color.Yellow(humanLine(justUnder)))
		assert.Contains(t, out, humanLine(justUnder))
	})
}
