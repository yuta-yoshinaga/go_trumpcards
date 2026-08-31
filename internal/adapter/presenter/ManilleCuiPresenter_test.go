//go:build test

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

func makeManillePlayers() []*domain.ManillePlayer {
	return []*domain.ManillePlayer{
		domain.NewManillePlayer(true),
		domain.NewManillePlayer(false),
		domain.NewManillePlayer(false),
		domain.NewManillePlayer(false),
	}
}

func setupManilleCuiMock() *interfaces.MockManilleGame {
	m := new(interfaces.MockManilleGame)
	m.On("GetRoundNumber").Return(1)
	m.On("GetTrickNumber").Return(1)
	m.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	m.On("GetCurrentTrick").Return(([]*domain.TrickCard)(nil))
	m.On("GetGameEndFlag").Return(false)
	m.On("GetPhase").Return(domain.ManillePhasePlay)
	m.On("GetCurrentPlayerIdx").Return(0)
	m.On("GetWinnerTeam").Return(-1)
	m.On("GetRoundCardPoints").Return([domain.ManilleTeamCnt]int{0, 0})
	m.On("GetTeamScores").Return([domain.ManilleTeamCnt]int{0, 0})
	m.On("GetActionLog").Return(([]*domain.ActionLogEntry)(nil))
	return m
}

func setupManilleCuiMockWithPlayers() (*interfaces.MockManilleGame, []*domain.ManillePlayer) {
	m := setupManilleCuiMock()
	players := makeManillePlayers()
	m.On("GetPlayerCnt").Return(4)
	m.On("GetLeadPlayerIdx").Return(-1).Maybe()
	m.On("GetPlayer", 0).Return(players[0])
	m.On("GetPlayer", 1).Return(players[1])
	m.On("GetPlayer", 2).Return(players[2])
	m.On("GetPlayer", 3).Return(players[3])
	return m, players
}

func TestManilleCuiPresenter_Output(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ManilleCuiPresenter)

	t.Run("play phase shows current player", func(t *testing.T) {
		m, players := setupManilleCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		result := p.Output(m, nil)
		assert.Contains(t, result, "Manille")
		// The play prompt includes the inverted rank-order reminder.
		assert.Contains(t, result, i18n.T("manille.rankHelp"))
	})

	t.Run("trick end prompt", func(t *testing.T) {
		m, _ := setupManilleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ManillePhaseTrickEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("round end prompt", func(t *testing.T) {
		m, _ := setupManilleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ManillePhaseRoundEnd)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("game end banner", func(t *testing.T) {
		m, _ := setupManilleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetGameEndFlag")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetWinnerTeam")
		m.On("GetGameEndFlag").Return(true)
		m.On("GetWinnerTeam").Return(0)
		result := p.Output(m, nil)
		assert.NotEmpty(t, result)
	})

	t.Run("error block", func(t *testing.T) {
		m, _ := setupManilleCuiMockWithPlayers()
		result := p.Output(m, errors.New("boom"))
		assert.Contains(t, result, "boom")
	})
}

func TestManilleCuiPresenter_HintOutput(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ManilleCuiPresenter)

	t.Run("no hint", func(t *testing.T) {
		m, _ := setupManilleCuiMockWithPlayers()
		m.On("GetHint").Return((*domain.ManilleHint)(nil))
		result := p.HintOutput(m)
		assert.NotEmpty(t, result)
	})

	t.Run("play hint with card index", func(t *testing.T) {
		m, players := setupManilleCuiMockWithPlayers()
		players[0].AddCard(domain.NewCard(domain.CardDesignSpade, 13, false))
		m.On("GetHint").Return(&domain.ManilleHint{CardIndices: []int{0}, Reason: "lead_low"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})

	t.Run("hint no card indices", func(t *testing.T) {
		m, _ := setupManilleCuiMockWithPlayers()
		m.On("GetHint").Return(&domain.ManilleHint{CardIndices: nil, Reason: "follow_win"})
		result := p.HintOutput(m)
		assert.Contains(t, result, "HINT")
	})
}

func TestManilleCuiPresenter_ActionLogOutput(t *testing.T) {
	p := new(presenter.ManilleCuiPresenter)
	m := new(interfaces.MockManilleGame)
	m.On("GetGameEndFlag").Return(true)
	m.On("GetActionLog").Return([]*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "play", Detail: "You plays ♠K"},
	})
	// 棋譜の座席名は同じ画面の他の行と同じ解決を通る (#5977)。
	m.On("GetPlayer", mock.Anything).Return(domain.NewManillePlayer(true)).Maybe()
	result := p.ActionLogOutput(m)
	assert.Contains(t, result, "play")
}

// #5646: トリックを取ったのが誰かは、次にリードするのが誰かでもある。Web は
// manille-trick-winner で名前とチームを出し、自チームなら色まで変えているのに、
// CUI は「次のトリックへ」としか言わず勝者に触れていなかった。姉妹の Sueca は
// 同じ場面で GetLeadPlayerIdx から勝者名を組み立てている。
func TestManilleCuiPresenter_TrickEndNamesTheWinner(t *testing.T) {
	p := new(presenter.ManilleCuiPresenter)

	trickEndMock := func(winner int) *interfaces.MockManilleGame {
		m, _ := setupManilleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.On("GetPhase").Return(domain.ManillePhaseTrickEnd)
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetLeadPlayerIdx")
		m.On("GetLeadPlayerIdx").Return(winner)
		return m
	}

	t.Run("names the winner and their team", func(t *testing.T) {
		out := p.Output(trickEndMock(0), nil)

		assert.Contains(t, out, i18n.Tf("manille.trickWinner",
			"name", color.Bold(i18n.T("cuiPlayerYou")), "team", i18n.T("manille.teamA")))
	})

	// 席1 はチームB。席の偶奇でチームが決まるので、片方だけ見ていると気づけない。
	t.Run("reads the team from the seat", func(t *testing.T) {
		out := p.Output(trickEndMock(1), nil)

		// **名前も席のもの。**チーム名だけ見ていると、常に席0を名乗る実装を
		// 見逃す (実際に一度見逃した)。
		assert.Contains(t, out, i18n.Tf("manille.trickWinner",
			"name", color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "1")),
			"team", i18n.T("manille.teamB")))
		// **チーム名そのものを探さない。**同じ画面の途中経過パネルが
		// 「チームA: 0点, チームB: 0点」を出しているので、素の語では当たってしまう
		// (#6442)。見るのは獲得の一文そのもの。
		assert.NotContains(t, out, i18n.Tf("manille.trickWinner",
			"name", color.Bold(i18n.T("cuiPlayerYou")), "team", i18n.T("manille.teamA")))
	})

	// まだ誰も取っていない (リード未確定) 局面では出さない。
	t.Run("says nothing when no one has led yet", func(t *testing.T) {
		out := p.Output(trickEndMock(-1), nil)

		for _, team := range []string{"manille.teamA", "manille.teamB"} {
			assert.NotContains(t, out, i18n.Tf("manille.trickWinner",
				"name", color.Bold(i18n.T("cuiPlayerYou")), "team", i18n.T(team)))
			assert.NotContains(t, out, i18n.Tf("manille.trickWinner",
				"name", color.Bold(i18n.Tf("cuiPlayerCpu", "idx", "1")), "team", i18n.T(team)))
		}
		// 獲得の一文の**末尾**そのものが無いことも見る (名前にもチームにも依らない)。
		_, rest, ok := strings.Cut(i18n.Tf("manille.trickWinner", "name", "\x00", "team", "\x00"), "\x00")
		require.True(t, ok)
		tail := rest[strings.LastIndex(rest, "\x00")+1:]
		require.NotEmpty(t, tail)
		assert.NotContains(t, out, tail)
	})
}

// **途中経過が追えなかった。**`GetRoundCardPoints()` を読むのは RoundEnd の分岐だけで、
// 進行中は累計点しか出ていなかった (#6442)。姉妹ゲームはどれも進行中に出している。
func TestManilleCuiPresenter_ShowsTheRoundProgress(t *testing.T) {
	orig := color.NoColor()
	color.SetNoColor(true)
	defer color.SetNoColor(orig)
	p := new(presenter.ManilleCuiPresenter)

	inPhase := func(phase domain.ManillePhase) *interfaces.MockManilleGame {
		m, _ := setupManilleCuiMockWithPlayers()
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetPhase")
		m.ExpectedCalls = removeMockCall(m.ExpectedCalls, "GetRoundCardPoints")
		m.On("GetPhase").Return(phase)
		m.On("GetRoundCardPoints").Return([domain.ManilleTeamCnt]int{18, 12})
		return m
	}

	want := i18n.Tf("manille.roundProgress", "ptsA", "18", "ptsB", "12")

	for _, phase := range []domain.ManillePhase{domain.ManillePhasePlay, domain.ManillePhaseTrickEnd} {
		t.Run("shows the running points mid-round", func(t *testing.T) {
			out := p.Output(inPhase(phase), nil)
			assert.Contains(t, out, want)
			assert.NotContains(t, out, "{{")
		})
	}

	// ラウンドが終われば `promptRoundEnd` が結果として引き継ぐ。二重に出さない。
	t.Run("stays quiet once the round has ended", func(t *testing.T) {
		out := p.Output(inPhase(domain.ManillePhaseRoundEnd), nil)
		assert.NotContains(t, out, want)
		assert.Contains(t, out, i18n.Tf("manille.promptRoundEnd", "ptsA", "18", "ptsB", "12"))
	})
}
