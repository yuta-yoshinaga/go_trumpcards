//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func zwTestGame(t *testing.T) *domain.Zwicker {
	t.Helper()
	z := domain.NewDefaultZwicker()
	z.Reset()
	return z
}

func zwDecode(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

func zwpCard(design, value int) *domain.Card { return domain.NewCard(design, value, true) }

// zwStub wires a MockZwickerGame with every accessor the presenters touch.
func zwStub(phase domain.ZwickerPhase, gameEnd bool, winner int, builds []*domain.ZwickerBuild) *interfaces.MockZwickerGame {
	g := new(interfaces.MockZwickerGame)
	g.On("GetPhase").Return(phase)
	g.On("GetGameEndFlag").Return(gameEnd)
	g.On("GetWinnerTeam").Return(winner)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetStockCount").Return(20)
	g.On("GetTableCards").Return([]*domain.Card{zwpCard(domain.CardDesignHeart, 9)})
	g.On("GetBuilds").Return(builds)
	g.On("GetTeamScore", mock.Anything).Return(12)
	g.On("GetLastRoundScore").Return((*domain.ZwickerRoundScore)(nil))
	g.On("GetConfig").Return(domain.DefaultZwickerConfig())
	players := make([]*domain.ZwickerPlayer, 0, domain.ZwickerPlayerCnt)
	players = append(players, domain.NewZwickerPlayer(true))
	for range domain.ZwickerPlayerCnt - 1 {
		players = append(players, domain.NewZwickerPlayer(false))
	}
	g.On("GetPlayers").Return(players)
	g.On("GetPlayer", mock.Anything).Return(domain.NewZwickerPlayer(false))
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	g.On("ZwickerCpuDecide", mock.Anything).Return(domain.ZwickerCpuAction{Type: "trail", HandIdx: 2})
	return g
}

func TestZwickerWebPresenter_HidesTheCpuHandButNeverTheProgress(t *testing.T) {
	// 枚数最多の 3 点と Zwick 1 点はどちらも公開情報から直に読める。
	out := zwDecode(t, new(ZwickerWebPresenter).Output(zwTestGame(t), nil))
	players, ok := out["players"].([]any)
	require.True(t, ok)
	require.Len(t, players, domain.ZwickerPlayerCnt)

	human, _ := players[0].(map[string]any)
	assert.False(t, human["hidden"].(bool))
	assert.NotEmpty(t, human["cards"])

	cpu, _ := players[1].(map[string]any)
	assert.True(t, cpu["hidden"].(bool))
	assert.Empty(t, cpu["cards"], "the opponent's hand must not reach the browser")
	assert.Positive(t, cpu["cardCount"], "but its size is public")
	assert.NotNil(t, cpu["capturedCount"], "and so is how much it has taken")
	assert.NotNil(t, cpu["zwicks"])
	// 向かい合わせが味方。
	assert.Equal(t, float64(0), human["team"])
	assert.Equal(t, float64(1), cpu["team"])
}

// TestZwickerWebPresenter_ShipsTheMatchingValuesWithEveryCard is the one the
// page cannot work without: A/J/Q/K each carry two values and the jokers carry
// 15/20/25, so the client must not re-derive the table.
func TestZwickerWebPresenter_ShipsTheMatchingValuesWithEveryCard(t *testing.T) {
	g := zwStub(domain.ZwickerPhasePlay, false, -1, nil)
	g.On("GetTableCards").Unset()
	g.On("GetTableCards").Return([]*domain.Card{
		zwpCard(domain.CardDesignSpade, 1),
		zwpCard(domain.CardDesignHeart, 13),
		zwpCard(domain.CardDesignClover, 7),
		domain.NewCard(domain.CardDesignJoker, 3, true),
	})
	out := zwDecode(t, new(ZwickerWebPresenter).Output(g, nil))
	table, ok := out["tableCards"].([]any)
	require.True(t, ok)
	require.Len(t, table, 4)

	want := [][]float64{{1, 11}, {4, 14}, {7}, {domain.ZwickerJokerLarge}}
	for i, w := range want {
		card, _ := table[i].(map[string]any)
		vals, ok := card["values"].([]any)
		require.True(t, ok, "card %d has no values", i)
		require.Len(t, vals, len(w), "card %d", i)
		for j := range w {
			assert.Equal(t, w[j], vals[j], "card %d value %d", i, j)
		}
	}
}

func TestZwickerWebPresenter_ShipsTheBuildsWithTheirDeclaredValue(t *testing.T) {
	builds := []*domain.ZwickerBuild{
		{Owner: 1, Value: 9, Cards: []*domain.Card{
			zwpCard(domain.CardDesignSpade, 5), zwpCard(domain.CardDesignHeart, 4),
		}},
	}
	g := zwStub(domain.ZwickerPhasePlay, false, -1, builds)
	out := zwDecode(t, new(ZwickerWebPresenter).Output(g, nil))
	got, ok := out["builds"].([]any)
	require.True(t, ok)
	require.Len(t, got, 1)
	first, _ := got[0].(map[string]any)
	assert.Equal(t, float64(1), first["owner"])
	// ビルドは宣言値ちょうどでしか取れないので、値を送らないと押せるかが判らない。
	assert.Equal(t, float64(9), first["value"])
	assert.Len(t, first["cards"], 2)
}

func TestZwickerWebPresenter_ShipsTheTargetScore(t *testing.T) {
	out := zwDecode(t, new(ZwickerWebPresenter).Output(zwTestGame(t), nil))
	assert.Equal(t, float64(domain.DefaultZwickerConfig().TargetScore), out["targetScore"])
}

func TestZwickerWebPresenter_ShipsTheRoundBreakdown(t *testing.T) {
	g := zwStub(domain.ZwickerPhaseRoundEnd, false, -1, nil)
	g.On("GetLastRoundScore").Unset()
	g.On("GetLastRoundScore").Return(&domain.ZwickerRoundScore{
		CardPoints: [2]int{17, 10}, Cards: [2]int{30, 25},
		MajorityTeam: 0, Zwicks: [2]int{1, 0}, Total: [2]int{21, 10},
	})
	out := zwDecode(t, new(ZwickerWebPresenter).Output(g, nil))
	last, ok := out["lastRound"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(0), last["majorityTeam"])
	totals, _ := last["total"].([]any)
	assert.Equal(t, float64(21), totals[0])
}

func TestZwickerWebPresenter_CarriesTheHintOnAPlainStateResponse(t *testing.T) {
	g := zwStub(domain.ZwickerPhasePlay, false, -1, nil)
	assert.NotNil(t, zwDecode(t, new(ZwickerWebPresenter).Output(g, nil))["hint"],
		"the hint toggle reads the ordinary response")
}

func TestZwickerWebPresenter_HintReasonsCoverEveryBranch(t *testing.T) {
	for name, tc := range map[string]struct {
		gameEnd bool
		phase   domain.ZwickerPhase
		current int
		action  domain.ZwickerCpuAction
		want    string
	}{
		"game over":     {true, domain.ZwickerPhaseGameEnd, 0, domain.ZwickerCpuAction{}, "zwicker.hint.game_end"},
		"deal over":     {false, domain.ZwickerPhaseRoundEnd, 0, domain.ZwickerCpuAction{}, "zwicker.hint.round_end"},
		"not your turn": {false, domain.ZwickerPhasePlay, 1, domain.ZwickerCpuAction{}, "zwicker.hint.not_your_turn"},
		"take":          {false, domain.ZwickerPhasePlay, 0, domain.ZwickerCpuAction{Type: "take", HandIdx: 1, Value: 7, TableIdxs: []int{0}}, "zwicker.hint.take"},
		"trail":         {false, domain.ZwickerPhasePlay, 0, domain.ZwickerCpuAction{Type: "trail", HandIdx: 2}, "zwicker.hint.trail"},
		"nothing":       {false, domain.ZwickerPhasePlay, 0, domain.ZwickerCpuAction{Type: "trail", HandIdx: -1}, "zwicker.hint.none"},
	} {
		t.Run(name, func(t *testing.T) {
			g := new(interfaces.MockZwickerGame)
			g.On("GetGameEndFlag").Return(tc.gameEnd)
			g.On("GetPhase").Return(tc.phase)
			g.On("GetCurrentPlayerIdx").Return(tc.current)
			g.On("ZwickerCpuDecide", 0).Return(tc.action)

			assert.Equal(t, tc.want, zwickerHint(g).Reason)
		})
	}
}

func TestZwickerWebPresenter_MessageCodes(t *testing.T) {
	win := zwStub(domain.ZwickerPhaseGameEnd, true, domain.ZwickerTeamOf(0), nil)
	assert.Equal(t, "zwicker.win", zwDecode(t, new(ZwickerWebPresenter).Output(win, nil))["messageCode"])

	lose := zwStub(domain.ZwickerPhaseGameEnd, true, 1-domain.ZwickerTeamOf(0), nil)
	assert.Equal(t, "zwicker.lose", zwDecode(t, new(ZwickerWebPresenter).Output(lose, nil))["messageCode"])

	running := zwStub(domain.ZwickerPhasePlay, false, -1, nil)
	assert.Empty(t, zwDecode(t, new(ZwickerWebPresenter).Output(running, nil))["message"])
	assert.Equal(t, "boom", zwDecode(t, new(ZwickerWebPresenter).Output(running, errors.New("boom")))["message"])
}

func TestZwickerWebPresenter_HintOutputAndActionLog(t *testing.T) {
	z := zwTestGame(t)
	assert.NotNil(t, zwDecode(t, new(ZwickerWebPresenter).HintOutput(z))["hint"])
	assert.NotEmpty(t, new(ZwickerWebPresenter).ActionLogOutput(z))
}

func TestZwickerCuiPresenter_ShowsTheRulesAndTheTable(t *testing.T) {
	out := new(ZwickerCuiPresenter).Output(zwTestGame(t), nil)
	assert.Contains(t, out, i18n.T("zwicker.ruleLine"))
	assert.Contains(t, out, "[0]", "your hand is indexed")
}

// TestZwickerCuiPresenter_PrintsTheMatchingValues は、値を書かないと遊べない
// ことを押さえる。ランクからは何と取れるか分からない。
func TestZwickerCuiPresenter_PrintsTheMatchingValues(t *testing.T) {
	g := zwStub(domain.ZwickerPhasePlay, false, -1, nil)
	g.On("GetTableCards").Unset()
	g.On("GetTableCards").Return([]*domain.Card{
		zwpCard(domain.CardDesignHeart, 13),
		domain.NewCard(domain.CardDesignJoker, 2, true),
	})
	out := new(ZwickerCuiPresenter).Output(g, nil)
	assert.Contains(t, out, "(4/14)", "a king must show both of its values")
	assert.Contains(t, out, "(20)", "the middle joker is fixed at twenty")
}

func TestZwickerCuiPresenter_NamesTheBuildValue(t *testing.T) {
	builds := []*domain.ZwickerBuild{
		{Owner: 1, Value: 9, Cards: []*domain.Card{
			zwpCard(domain.CardDesignSpade, 5), zwpCard(domain.CardDesignHeart, 4),
		}},
	}
	g := zwStub(domain.ZwickerPhasePlay, false, -1, builds)
	out := new(ZwickerCuiPresenter).Output(g, nil)
	// **赤いスートは ANSI で装飾される**ので、札を含めた連結文字列にはならない。
	// 札より前の部分だけで照合する。
	head := strings.SplitN(i18n.Tf("zwicker.buildLine",
		"idx", "0", "value", "9", "owner", "1", "cards", "@@@"), "@@@", 2)[0]
	assert.Contains(t, out, head)
	// 札そのものは値つきで出ていること。
	assert.Contains(t, out, "SPADE 5(5)")
}

func TestZwickerCuiPresenter_PromptsPerPhase(t *testing.T) {
	p := new(ZwickerCuiPresenter)
	play := zwStub(domain.ZwickerPhasePlay, false, -1, nil)
	assert.Contains(t, p.Output(play, nil), i18n.T("zwicker.promptPlay"))

	over := zwStub(domain.ZwickerPhaseRoundEnd, false, -1, nil)
	overOut := p.Output(over, nil)
	assert.Contains(t, overOut, i18n.T("zwicker.promptNext"))
	assert.NotContains(t, overOut, i18n.T("zwicker.promptPlay"))
}

// TestZwickerCuiPresenter_SaysWhenTheMajorityWasTied は、枚数が並んで 3 点が
// 宙に浮いたことを表示から読めるようにする。黙っていると合計が合わないように
// 見える。
func TestZwickerCuiPresenter_SaysWhenTheMajorityWasTied(t *testing.T) {
	tied := zwStub(domain.ZwickerPhaseRoundEnd, false, -1, nil)
	tied.On("GetLastRoundScore").Unset()
	tied.On("GetLastRoundScore").Return(&domain.ZwickerRoundScore{MajorityTeam: -1, Total: [2]int{5, 5}})
	assert.Contains(t, new(ZwickerCuiPresenter).promptBlock(tied), i18n.T("zwicker.majorityTied"))

	clear := zwStub(domain.ZwickerPhaseRoundEnd, false, -1, nil)
	clear.On("GetLastRoundScore").Unset()
	clear.On("GetLastRoundScore").Return(&domain.ZwickerRoundScore{MajorityTeam: 0, Total: [2]int{8, 5}})
	assert.NotContains(t, new(ZwickerCuiPresenter).promptBlock(clear), i18n.T("zwicker.majorityTied"))
}

func TestZwickerCuiPresenter_BannersAndErrors(t *testing.T) {
	p := new(ZwickerCuiPresenter)
	win := zwStub(domain.ZwickerPhaseGameEnd, true, domain.ZwickerTeamOf(0), nil)
	lose := zwStub(domain.ZwickerPhaseGameEnd, true, 1-domain.ZwickerTeamOf(0), nil)
	assert.NotEqual(t, p.Output(win, nil), p.Output(lose, nil), "the two endings must not read the same")

	running := zwStub(domain.ZwickerPhasePlay, false, -1, nil)
	assert.Contains(t, p.Output(running, errors.New("boom")), "boom")
}

func TestZwickerCuiPresenter_HintRendersEveryShape(t *testing.T) {
	p := new(ZwickerCuiPresenter)

	take := new(interfaces.MockZwickerGame)
	take.On("GetGameEndFlag").Return(false)
	take.On("GetPhase").Return(domain.ZwickerPhasePlay)
	take.On("GetCurrentPlayerIdx").Return(0)
	take.On("ZwickerCpuDecide", 0).Return(domain.ZwickerCpuAction{
		Type: "take", HandIdx: 1, Value: 7, TableIdxs: []int{0, 2},
	})
	takeOut := p.HintOutput(take)
	assert.Contains(t, takeOut, "0,2")
	assert.Contains(t, takeOut, "7")

	trail := new(interfaces.MockZwickerGame)
	trail.On("GetGameEndFlag").Return(false)
	trail.On("GetPhase").Return(domain.ZwickerPhasePlay)
	trail.On("GetCurrentPlayerIdx").Return(0)
	trail.On("ZwickerCpuDecide", 0).Return(domain.ZwickerCpuAction{Type: "trail", HandIdx: 3})
	assert.Contains(t, p.HintOutput(trail), "3")

	over := new(interfaces.MockZwickerGame)
	over.On("GetGameEndFlag").Return(true)
	assert.NotEmpty(t, p.HintOutput(over))
}

func TestZwickerCuiPresenter_HintReasonKeysAreAllMapped(t *testing.T) {
	for _, reason := range []string{
		"zwicker.hint.game_end", "zwicker.hint.round_end", "zwicker.hint.not_your_turn",
		"zwicker.hint.take", "zwicker.hint.trail", "zwicker.hint.none",
	} {
		assert.NotEmpty(t, zwickerHintReasonKeys[reason], "unmapped reason %s", reason)
	}
}

func TestZwickerCuiPresenter_ActionLog(t *testing.T) {
	assert.NotEmpty(t, new(ZwickerCuiPresenter).ActionLogOutput(zwTestGame(t)))
}

// #5721: ビルドの Owner は作った席を記録するだけで、宣言値ちょうどの札を出せる
// 誰でも取れる (domain の TestZwickerBuildIsNotOwnerOnly)。「所有者」という語だけを
// 出すと「持ち主しか取れない」と誤解され、育てている間に取られて手番を失う。
func TestZwickerCuiPresenter_SaysBuildsAreNotOwnerOnly(t *testing.T) {
	p := new(ZwickerCuiPresenter)
	g := domain.NewDefaultZwicker()
	g.Reset()

	out := p.Output(g, nil)

	rule := i18n.T("zwicker.ruleLine")
	assert.Contains(t, out, rule)
	assert.Contains(t, rule, "誰でも", "the rule line must say builds are not owner-only")
}
