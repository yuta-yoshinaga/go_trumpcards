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

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func pcTestGame(t *testing.T) *domain.Poch {
	t.Helper()
	p := domain.NewDefaultPoch()
	p.Reset()
	return p
}

func pcDecode(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

func pcpCard(design, value int) *domain.Card { return domain.NewCard(design, value, true) }

// pcBoardWith returns a board with every pool holding n chips.
func pcBoardWith(n int) domain.PochBoard {
	var b domain.PochBoard
	b.Ante(n)
	return b
}

// pcStub wires a MockPochGame with every accessor the presenters touch.
func pcStub(phase domain.PochPhase, gameEnd bool, winner int, awards []*domain.PochStakingAward) *interfaces.MockPochGame {
	g := new(interfaces.MockPochGame)
	g.On("GetPhase").Return(phase)
	g.On("PochValidPlays", mock.Anything).Return([]int{}).Maybe()
	g.On("GetGameEndFlag").Return(gameEnd)
	g.On("GetWinnerIdx").Return(winner)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetBoard").Return(pcBoardWith(4))
	g.On("GetPaySuit").Return(domain.CardDesignSpade)
	g.On("GetTurnUp").Return(pcpCard(domain.CardDesignSpade, 9))
	g.On("GetStakingAwards").Return(awards)
	g.On("GetBetTarget").Return(1)
	g.On("GetPochenWinner").Return(-1)
	g.On("GetPochenPot").Return(0)
	g.On("GetPlayedPile").Return([]*domain.Card{})
	g.On("GetStopsSuit").Return(-1)
	g.On("GetStopsRank").Return(0)
	g.On("GetDealNumber").Return(0)
	g.On("GetDealWinner").Return(-1)
	g.On("GetConfig").Return(domain.DefaultPochConfig())
	players := make([]*domain.PochPlayer, 0, domain.PochPlayerCnt)
	players = append(players, domain.NewPochPlayer(true))
	for range domain.PochPlayerCnt - 1 {
		players = append(players, domain.NewPochPlayer(false))
	}
	g.On("GetPlayers").Return(players)
	g.On("GetPlayer", mock.Anything).Return(domain.NewPochPlayer(false))
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	g.On("PochCpuDecide", mock.Anything).Return(domain.PochCpuAction{Type: "fold", HandIdx: -1})
	return g
}

func TestPochWebPresenter_HidesTheCpuHandButNeverTheCount(t *testing.T) {
	// 出し切った人は他家から**残り札 1 枚につき 1 チップ**受け取るので、
	// 枚数はそのまま負債額。隠しようがない。
	out := pcDecode(t, new(PochWebPresenter).Output(pcTestGame(t), nil))
	players, ok := out["players"].([]any)
	require.True(t, ok)
	require.Len(t, players, domain.PochPlayerCnt)

	human, _ := players[0].(map[string]any)
	assert.False(t, human["hidden"].(bool))
	assert.NotEmpty(t, human["cards"])

	cpu, _ := players[1].(map[string]any)
	assert.True(t, cpu["hidden"].(bool))
	assert.Empty(t, cpu["cards"], "the opponent's hand must not reach the browser")
	assert.Positive(t, cpu["cardCount"], "but its size is public")
	assert.NotNil(t, cpu["chips"])
}

// TestPochWebPresenter_ShipsAllNinePools is what the board is for: the chips
// that go unclaimed carry over, so "what is sitting on each pool" is the whole
// motivation for the next deal.
func TestPochWebPresenter_ShipsAllNinePools(t *testing.T) {
	g := pcStub(domain.PochPhasePochen, false, -1, nil)
	out := pcDecode(t, new(PochWebPresenter).Output(g, nil))
	pools, ok := out["pools"].([]any)
	require.True(t, ok)
	require.Len(t, pools, int(domain.PochPoolCount))

	want := []string{"ace", "king", "queen", "jack", "ten", "marriage", "sequence", "pocher", "centre"}
	for i, w := range want {
		pool, _ := pools[i].(map[string]any)
		assert.Equal(t, w, pool["name"], "pool %d", i)
		assert.Equal(t, float64(4), pool["chips"], "pool %d", i)
	}
}

// TestPochWebPresenter_ShipsTheStakingResults matters because stage one
// resolves itself the instant the cards are dealt -- without the record the
// player cannot tell what happened.
func TestPochWebPresenter_ShipsTheStakingResults(t *testing.T) {
	awards := []*domain.PochStakingAward{
		{Pool: domain.PochPoolMarriage, Player: 2, Chips: 12},
	}
	g := pcStub(domain.PochPhasePochen, false, -1, awards)
	out := pcDecode(t, new(PochWebPresenter).Output(g, nil))
	got, ok := out["stakingAwards"].([]any)
	require.True(t, ok)
	require.Len(t, got, 1)
	first, _ := got[0].(map[string]any)
	assert.Equal(t, "marriage", first["pool"])
	assert.Equal(t, float64(2), first["player"])
	assert.Equal(t, float64(12), first["chips"])
	assert.Equal(t, float64(domain.CardDesignSpade), out["paySuit"])
	assert.NotNil(t, out["turnUp"])
}

// TestPochWebPresenter_KeepsAStoppedRunDistinct は -1 を潰さないことを確かめる。
// 0 に丸めると ♠ の並びの途中に見えてしまう。
func TestPochWebPresenter_KeepsAStoppedRunDistinct(t *testing.T) {
	stopped := pcStub(domain.PochPhaseStops, false, -1, nil)
	assert.Equal(t, float64(-1), pcDecode(t, new(PochWebPresenter).Output(stopped, nil))["stopsSuit"])

	live := pcStub(domain.PochPhaseStops, false, -1, nil)
	live.On("GetStopsSuit").Unset()
	live.On("GetStopsSuit").Return(domain.CardDesignHeart)
	live.On("GetStopsRank").Unset()
	live.On("GetStopsRank").Return(9)
	out := pcDecode(t, new(PochWebPresenter).Output(live, nil))
	assert.Equal(t, float64(domain.CardDesignHeart), out["stopsSuit"])
	assert.Equal(t, float64(9), out["stopsRank"])
}

func TestPochWebPresenter_CarriesTheHintOnAPlainStateResponse(t *testing.T) {
	g := pcStub(domain.PochPhasePochen, false, -1, nil)
	assert.NotNil(t, pcDecode(t, new(PochWebPresenter).Output(g, nil))["hint"],
		"the hint toggle reads the ordinary response")
}

func TestPochWebPresenter_HintReasonsCoverEveryBranch(t *testing.T) {
	for name, tc := range map[string]struct {
		gameEnd bool
		phase   domain.PochPhase
		current int
		action  domain.PochCpuAction
		want    string
	}{
		"game over":     {true, domain.PochPhaseGameEnd, 0, domain.PochCpuAction{}, "poch.hint.game_end"},
		"deal over":     {false, domain.PochPhaseDealEnd, 0, domain.PochCpuAction{}, "poch.hint.deal_end"},
		"not your turn": {false, domain.PochPhasePochen, 1, domain.PochCpuAction{}, "poch.hint.not_your_turn"},
		"bet":           {false, domain.PochPhasePochen, 0, domain.PochCpuAction{Type: "bet", HandIdx: -1}, "poch.hint.bet"},
		"fold":          {false, domain.PochPhasePochen, 0, domain.PochCpuAction{Type: "fold", HandIdx: -1}, "poch.hint.fold"},
		"play":          {false, domain.PochPhaseStops, 0, domain.PochCpuAction{Type: "play", HandIdx: 2}, "poch.hint.play"},
		"nothing":       {false, domain.PochPhaseStops, 0, domain.PochCpuAction{Type: "play", HandIdx: -1}, "poch.hint.none"},
	} {
		t.Run(name, func(t *testing.T) {
			g := new(interfaces.MockPochGame)
			g.On("GetGameEndFlag").Return(tc.gameEnd)
			g.On("GetPhase").Return(tc.phase)
			g.On("GetCurrentPlayerIdx").Return(tc.current)
			g.On("PochCpuDecide", 0).Return(tc.action)

			assert.Equal(t, tc.want, pochHint(g).Reason)
		})
	}
}

func TestPochWebPresenter_MessageCodes(t *testing.T) {
	win := pcStub(domain.PochPhaseGameEnd, true, 0, nil)
	assert.Equal(t, "poch.win", pcDecode(t, new(PochWebPresenter).Output(win, nil))["messageCode"])

	lose := pcStub(domain.PochPhaseGameEnd, true, 2, nil)
	assert.Equal(t, "poch.lose", pcDecode(t, new(PochWebPresenter).Output(lose, nil))["messageCode"])

	running := pcStub(domain.PochPhasePochen, false, -1, nil)
	assert.Empty(t, pcDecode(t, new(PochWebPresenter).Output(running, nil))["message"])
	assert.Equal(t, "boom", pcDecode(t, new(PochWebPresenter).Output(running, errors.New("boom")))["message"])
}

func TestPochWebPresenter_HintOutputAndActionLog(t *testing.T) {
	p := pcTestGame(t)
	assert.NotNil(t, pcDecode(t, new(PochWebPresenter).HintOutput(p))["hint"])
	assert.NotEmpty(t, new(PochWebPresenter).ActionLogOutput(p))
}

func TestPochCuiPresenter_ShowsTheRulesAndTheBoard(t *testing.T) {
	out := new(PochCuiPresenter).Output(pcTestGame(t), nil)
	assert.Contains(t, out, i18n.T("poch.ruleLine"))
	assert.Contains(t, out, "[0]", "your hand is indexed")
	// 9 区画すべてが出ていること。
	for _, name := range []string{"ace", "king", "queen", "jack", "ten", "marriage", "sequence", "pocher", "centre"} {
		assert.Contains(t, out, i18n.T("poch.pool."+name), "pool %s missing", name)
	}
}

// TestPochCuiPresenter_ReportsTheStakingResults は、自動で解決する第 1 段階の
// 結果が読めることを確かめる。
func TestPochCuiPresenter_ReportsTheStakingResults(t *testing.T) {
	awards := []*domain.PochStakingAward{
		{Pool: domain.PochPoolSequence, Player: 1, Chips: 20},
	}
	g := pcStub(domain.PochPhasePochen, false, -1, awards)
	out := new(PochCuiPresenter).Output(g, nil)
	// **名前は ANSI で装飾される**ので、名前より後ろの部分で照合する。
	tail := strings.SplitN(i18n.Tf("poch.awardLine",
		"name", "@@@", "pool", i18n.T("poch.pool.sequence"), "chips", "20"), "@@@", 2)[1]
	assert.Contains(t, out, tail)
}

func TestPochCuiPresenter_PromptsPerPhase(t *testing.T) {
	p := new(PochCuiPresenter)

	pochen := pcStub(domain.PochPhasePochen, false, -1, nil)
	assert.Contains(t, p.Output(pochen, nil), i18n.Tf("poch.promptPochen", "target", "1"))

	dealEnd := pcStub(domain.PochPhaseDealEnd, false, -1, nil)
	assert.Contains(t, p.Output(dealEnd, nil), i18n.T("poch.promptNext"))

	staking := pcStub(domain.PochPhaseStaking, false, -1, nil)
	assert.Contains(t, p.Output(staking, nil), i18n.T("poch.promptStaking"))
}

// TestPochCuiPresenter_TellsAStoppedRunApart は、並びが止まっているかどうかで
// 案内が変わることを確かめる。出せる札がまるで違う。
func TestPochCuiPresenter_TellsAStoppedRunApart(t *testing.T) {
	stopped := pcStub(domain.PochPhaseStops, false, -1, nil)
	out := new(PochCuiPresenter).promptBlock(stopped)
	assert.Contains(t, out, i18n.T("poch.promptFreeLead"))
	assert.NotContains(t, out, i18n.T("poch.promptFollow"))

	live := pcStub(domain.PochPhaseStops, false, -1, nil)
	live.On("GetStopsSuit").Unset()
	live.On("GetStopsSuit").Return(domain.CardDesignHeart)
	liveOut := new(PochCuiPresenter).promptBlock(live)
	assert.Contains(t, liveOut, i18n.T("poch.promptFollow"))
	assert.NotContains(t, liveOut, i18n.T("poch.promptFreeLead"))
}

func TestPochCuiPresenter_BannersAndErrors(t *testing.T) {
	p := new(PochCuiPresenter)
	win := pcStub(domain.PochPhaseGameEnd, true, 0, nil)
	lose := pcStub(domain.PochPhaseGameEnd, true, 2, nil)
	assert.NotEqual(t, p.Output(win, nil), p.Output(lose, nil), "the two endings must not read the same")

	running := pcStub(domain.PochPhasePochen, false, -1, nil)
	assert.Contains(t, p.Output(running, errors.New("boom")), "boom")
}

func TestPochCuiPresenter_HintRendersEveryShape(t *testing.T) {
	p := new(PochCuiPresenter)

	for _, tc := range []struct {
		phase  domain.PochPhase
		action domain.PochCpuAction
		want   string
	}{
		{domain.PochPhasePochen, domain.PochCpuAction{Type: "bet", HandIdx: -1}, ""},
		{domain.PochPhasePochen, domain.PochCpuAction{Type: "fold", HandIdx: -1}, ""},
		{domain.PochPhaseStops, domain.PochCpuAction{Type: "play", HandIdx: 3}, "3"},
	} {
		g := new(interfaces.MockPochGame)
		g.On("GetGameEndFlag").Return(false)
		g.On("GetPhase").Return(tc.phase)
		g.On("GetCurrentPlayerIdx").Return(0)
		g.On("PochCpuDecide", 0).Return(tc.action)
		out := p.HintOutput(g)
		assert.NotEmpty(t, out)
		if tc.want != "" {
			assert.Contains(t, out, tc.want)
		}
	}

	over := new(interfaces.MockPochGame)
	over.On("GetGameEndFlag").Return(true)
	assert.NotEmpty(t, p.HintOutput(over))
}

func TestPochCuiPresenter_HintReasonKeysAreAllMapped(t *testing.T) {
	for _, reason := range []string{
		"poch.hint.game_end", "poch.hint.deal_end", "poch.hint.not_your_turn",
		"poch.hint.bet", "poch.hint.fold", "poch.hint.play", "poch.hint.none",
	} {
		assert.NotEmpty(t, pochHintReasonKeys[reason], "unmapped reason %s", reason)
	}
}

func TestPochCuiPresenter_ActionLog(t *testing.T) {
	assert.NotEmpty(t, new(PochCuiPresenter).ActionLogOutput(pcTestGame(t)))
}

// **出せる札を応答に載せる。**フロントは押して弾かれるまで違反を示せない (#4933)。
func TestPochWebPresenter_ValidPlays(t *testing.T) {
	decode := func(raw string) controller.PochWebOutput {
		var parsed controller.PochWebOutput
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		return parsed
	}

	g := domain.NewDefaultPoch()
	g.Reset()
	raw := new(PochWebPresenter).Output(g, nil)
	got := decode(raw).ValidPlays
	// **null ではなく空配列。**フロントが `?? []` を忘れても壊れない。
	if got == nil {
		t.Fatal("validPlays must be [] rather than null")
	}
	// 応答に必ず現れる。
	if !strings.Contains(raw, `"validPlays"`) {
		t.Fatalf("validPlays missing from the response: %s", raw)
	}

	// **人間の手番でなければ空。**ドメインは nil を返すので、presenter 側で
	// 空スライスに寄せている。ここを通さないと null が漏れる。
	g.SetCurrentPlayerForTest(1)
	off := decode(new(PochWebPresenter).Output(g, nil)).ValidPlays
	if off == nil {
		t.Fatal("off-turn validPlays must be [] rather than null")
	}
	if len(off) != 0 {
		t.Fatalf("off-turn validPlays must be empty, got %v", off)
	}
}
