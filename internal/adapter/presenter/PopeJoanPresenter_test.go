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

func pjTestGame(t *testing.T) *domain.PopeJoan {
	t.Helper()
	p := domain.NewDefaultPopeJoan()
	p.Reset()
	return p
}

func pjDecode(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

func pjpCard(design, value int) *domain.Card { return domain.NewCard(design, value, true) }

// pjDressedBoard returns a board with the dealer's fixed stake on it.
func pjDressedBoard() domain.PopeJoanBoard {
	var b domain.PopeJoanBoard
	b.Dress()
	return b
}

// pjStub wires a MockPopeJoanGame with every accessor the presenters touch.
func pjStub(phase domain.PopeJoanPhase, gameEnd bool, winner int, awards []*domain.PopeJoanAward) *interfaces.MockPopeJoanGame {
	g := new(interfaces.MockPopeJoanGame)
	g.On("GetPhase").Return(phase)
	g.On("PopeJoanValidPlays", mock.Anything).Return([]int{}).Maybe()
	g.On("GetGameEndFlag").Return(gameEnd)
	g.On("GetWinnerIdx").Return(winner)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetBoard").Return(pjDressedBoard())
	g.On("GetTrumpSuit").Return(domain.CardDesignSpade)
	g.On("GetTurnUp").Return(pjpCard(domain.CardDesignSpade, 5))
	g.On("GetAwards").Return(awards)
	g.On("GetPlayedPile").Return([]*domain.Card{})
	g.On("GetRunSuit").Return(-1)
	g.On("GetRunRank").Return(0)
	g.On("GetDealNumber").Return(0)
	g.On("GetDealWinner").Return(-1)
	g.On("GetConfig").Return(domain.DefaultPopeJoanConfig())
	players := make([]*domain.PopeJoanPlayer, 0, domain.PopeJoanPlayerCnt)
	players = append(players, domain.NewPopeJoanPlayer(true))
	for range domain.PopeJoanPlayerCnt - 1 {
		players = append(players, domain.NewPopeJoanPlayer(false))
	}
	g.On("GetPlayers").Return(players)
	g.On("GetPlayer", mock.Anything).Return(domain.NewPopeJoanPlayer(false))
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	g.On("PopeJoanCpuDecide", mock.Anything).Return(-1)
	return g
}

func TestPopeJoanWebPresenter_HidesTheCpuHandButNeverTheCount(t *testing.T) {
	// 出し切られると残り札 1 枚につき 1 チップ払うので、枚数は負債そのもの。
	out := pjDecode(t, new(PopeJoanWebPresenter).Output(pjTestGame(t), nil))
	players, ok := out["players"].([]any)
	require.True(t, ok)
	require.Len(t, players, domain.PopeJoanPlayerCnt)

	human, _ := players[0].(map[string]any)
	assert.False(t, human["hidden"].(bool))
	assert.NotEmpty(t, human["cards"])

	cpu, _ := players[1].(map[string]any)
	assert.True(t, cpu["hidden"].(bool))
	assert.Empty(t, cpu["cards"], "the opponent's hand must not reach the browser")
	assert.Positive(t, cpu["cardCount"], "but its size is public")
	assert.NotNil(t, cpu["chips"])
	// **Pope 保持者は支払いを免除される**ので、伏せると精算が理解できない。
	assert.NotNil(t, cpu["holdsPope"])
}

// TestPopeJoanWebPresenter_ShipsAllEightCompartments is what the board is for:
// unclaimed chips carry over, so what sits on each compartment drives the deal.
func TestPopeJoanWebPresenter_ShipsAllEightCompartments(t *testing.T) {
	g := pjStub(domain.PopeJoanPhasePlay, false, -1, nil)
	out := pjDecode(t, new(PopeJoanWebPresenter).Output(g, nil))
	comps, ok := out["compartments"].([]any)
	require.True(t, ok)
	require.Len(t, comps, int(domain.PopeJoanCompartmentCount))

	// **ディーラーの内訳は固定。**Pope 6 / Matrimony 2 / Intrigue 2 / 残り各 1。
	want := []struct {
		name  string
		chips float64
	}{
		{"ace", 1}, {"king", 1}, {"queen", 1}, {"jack", 1}, {"game", 1},
		{"pope", 6}, {"matrimony", 2}, {"intrigue", 2},
	}
	for i, w := range want {
		comp, _ := comps[i].(map[string]any)
		assert.Equal(t, w.name, comp["name"], "compartment %d", i)
		assert.Equal(t, w.chips, comp["chips"], "compartment %s", w.name)
	}
}

// TestPopeJoanWebPresenter_MarksATurnUpAward covers the rule the issue omits:
// a turn-up of Pope/A/K/Q/J hands the dealer that compartment outright.
func TestPopeJoanWebPresenter_MarksATurnUpAward(t *testing.T) {
	awards := []*domain.PopeJoanAward{
		{Compartment: domain.PopeJoanPope, Player: 0, Chips: 6, ByTurnUp: true},
	}
	g := pjStub(domain.PopeJoanPhasePlay, false, -1, awards)
	out := pjDecode(t, new(PopeJoanWebPresenter).Output(g, nil))
	got, ok := out["awards"].([]any)
	require.True(t, ok)
	require.Len(t, got, 1)
	first, _ := got[0].(map[string]any)
	assert.Equal(t, "pope", first["compartment"])
	assert.Equal(t, float64(6), first["chips"])
	assert.Equal(t, true, first["byTurnUp"])
}

// TestPopeJoanWebPresenter_KeepsAStoppedRunDistinct は -1 を潰さないことを
// 確かめる。0 に丸めると ♠ の並びの途中に見えてしまう。
func TestPopeJoanWebPresenter_KeepsAStoppedRunDistinct(t *testing.T) {
	stopped := pjStub(domain.PopeJoanPhasePlay, false, -1, nil)
	assert.Equal(t, float64(-1), pjDecode(t, new(PopeJoanWebPresenter).Output(stopped, nil))["runSuit"])

	live := pjStub(domain.PopeJoanPhasePlay, false, -1, nil)
	live.On("GetRunSuit").Unset()
	live.On("GetRunSuit").Return(domain.CardDesignHeart)
	live.On("GetRunRank").Unset()
	live.On("GetRunRank").Return(9)
	out := pjDecode(t, new(PopeJoanWebPresenter).Output(live, nil))
	assert.Equal(t, float64(domain.CardDesignHeart), out["runSuit"])
	assert.Equal(t, float64(9), out["runRank"])
}

func TestPopeJoanWebPresenter_CarriesTheHintOnAPlainStateResponse(t *testing.T) {
	g := pjStub(domain.PopeJoanPhasePlay, false, -1, nil)
	g.On("PopeJoanCpuDecide", 0).Unset()
	g.On("PopeJoanCpuDecide", 0).Return(1)
	assert.NotNil(t, pjDecode(t, new(PopeJoanWebPresenter).Output(g, nil))["hint"],
		"the hint toggle reads the ordinary response")
}

func TestPopeJoanWebPresenter_HintReasonsCoverEveryBranch(t *testing.T) {
	for name, tc := range map[string]struct {
		gameEnd bool
		phase   domain.PopeJoanPhase
		current int
		runSuit int
		decide  int
		want    string
	}{
		"game over":     {true, domain.PopeJoanPhaseGameEnd, 0, -1, 0, "popejoan.hint.game_end"},
		"deal over":     {false, domain.PopeJoanPhaseDealEnd, 0, -1, 0, "popejoan.hint.deal_end"},
		"not your turn": {false, domain.PopeJoanPhasePlay, 1, -1, 0, "popejoan.hint.not_your_turn"},
		"nothing":       {false, domain.PopeJoanPhasePlay, 0, -1, -1, "popejoan.hint.none"},
		// 並びが止まっていれば「最も低い札」、途中なら「同スートの次」。
		"lead":   {false, domain.PopeJoanPhasePlay, 0, -1, 2, "popejoan.hint.lead"},
		"follow": {false, domain.PopeJoanPhasePlay, 0, domain.CardDesignSpade, 2, "popejoan.hint.follow"},
	} {
		t.Run(name, func(t *testing.T) {
			g := new(interfaces.MockPopeJoanGame)
			g.On("GetGameEndFlag").Return(tc.gameEnd)
			g.On("GetPhase").Return(tc.phase)
			g.On("GetCurrentPlayerIdx").Return(tc.current)
			g.On("GetRunSuit").Return(tc.runSuit)
			g.On("PopeJoanCpuDecide", 0).Return(tc.decide)

			assert.Equal(t, tc.want, popeJoanHint(g).Reason)
		})
	}
}

func TestPopeJoanWebPresenter_MessageCodes(t *testing.T) {
	win := pjStub(domain.PopeJoanPhaseGameEnd, true, 0, nil)
	assert.Equal(t, "popejoan.win", pjDecode(t, new(PopeJoanWebPresenter).Output(win, nil))["messageCode"])

	lose := pjStub(domain.PopeJoanPhaseGameEnd, true, 2, nil)
	assert.Equal(t, "popejoan.lose", pjDecode(t, new(PopeJoanWebPresenter).Output(lose, nil))["messageCode"])

	running := pjStub(domain.PopeJoanPhasePlay, false, -1, nil)
	assert.Empty(t, pjDecode(t, new(PopeJoanWebPresenter).Output(running, nil))["message"])
	assert.Equal(t, "boom", pjDecode(t, new(PopeJoanWebPresenter).Output(running, errors.New("boom")))["message"])
}

func TestPopeJoanWebPresenter_HintOutputAndActionLog(t *testing.T) {
	p := pjTestGame(t)
	assert.NotNil(t, pjDecode(t, new(PopeJoanWebPresenter).HintOutput(p))["hint"])
	assert.NotEmpty(t, new(PopeJoanWebPresenter).ActionLogOutput(p))
}

func TestPopeJoanCuiPresenter_ShowsTheRulesAndTheBoard(t *testing.T) {
	out := new(PopeJoanCuiPresenter).Output(pjTestGame(t), nil)
	assert.Contains(t, out, i18n.T("popejoan.ruleLine"))
	assert.Contains(t, out, "[0]", "your hand is indexed")
	for _, name := range []string{"ace", "king", "queen", "jack", "game", "pope", "matrimony", "intrigue"} {
		assert.Contains(t, out, i18n.T("popejoan.compartment."+name), "compartment %s missing", name)
	}
}

// TestPopeJoanCuiPresenter_TellsATurnUpAwardApart は、めくり札での即取りが
// 通常の獲得と区別して読めることを確かめる。
func TestPopeJoanCuiPresenter_TellsATurnUpAwardApart(t *testing.T) {
	turnUp := pjStub(domain.PopeJoanPhasePlay, false, -1, []*domain.PopeJoanAward{
		{Compartment: domain.PopeJoanPope, Player: 1, Chips: 6, ByTurnUp: true},
	})
	// **名前は ANSI で装飾される**ので、名前より後ろの部分で照合する。
	tail := strings.SplitN(i18n.Tf("popejoan.awardTurnUpLine",
		"name", "@@@", "compartment", i18n.T("popejoan.compartment.pope"), "chips", "6"), "@@@", 2)[1]
	assert.Contains(t, new(PopeJoanCuiPresenter).Output(turnUp, nil), tail)

	plain := pjStub(domain.PopeJoanPhasePlay, false, -1, []*domain.PopeJoanAward{
		{Compartment: domain.PopeJoanPope, Player: 1, Chips: 6, ByTurnUp: false},
	})
	assert.NotContains(t, new(PopeJoanCuiPresenter).Output(plain, nil), tail)
}

// TestPopeJoanCuiPresenter_TellsAStoppedRunApart は、止まっているかどうかで
// 出せる札がまるで違うことを案内に反映していることを確かめる。
func TestPopeJoanCuiPresenter_TellsAStoppedRunApart(t *testing.T) {
	stopped := pjStub(domain.PopeJoanPhasePlay, false, -1, nil)
	out := new(PopeJoanCuiPresenter).promptBlock(stopped)
	assert.Contains(t, out, i18n.T("popejoan.promptLead"))
	assert.NotContains(t, out, i18n.T("popejoan.promptFollow"))

	live := pjStub(domain.PopeJoanPhasePlay, false, -1, nil)
	live.On("GetRunSuit").Unset()
	live.On("GetRunSuit").Return(domain.CardDesignHeart)
	liveOut := new(PopeJoanCuiPresenter).promptBlock(live)
	assert.Contains(t, liveOut, i18n.T("popejoan.promptFollow"))
	assert.NotContains(t, liveOut, i18n.T("popejoan.promptLead"))
}

func TestPopeJoanCuiPresenter_PromptsAtTheEndOfADeal(t *testing.T) {
	dealEnd := pjStub(domain.PopeJoanPhaseDealEnd, false, -1, nil)
	dealEnd.On("GetDealWinner").Unset()
	dealEnd.On("GetDealWinner").Return(2)
	out := new(PopeJoanCuiPresenter).promptBlock(dealEnd)
	assert.Contains(t, out, i18n.T("popejoan.promptNext"))
}

func TestPopeJoanCuiPresenter_BannersAndErrors(t *testing.T) {
	p := new(PopeJoanCuiPresenter)
	win := pjStub(domain.PopeJoanPhaseGameEnd, true, 0, nil)
	lose := pjStub(domain.PopeJoanPhaseGameEnd, true, 2, nil)
	assert.NotEqual(t, p.Output(win, nil), p.Output(lose, nil), "the two endings must not read the same")

	running := pjStub(domain.PopeJoanPhasePlay, false, -1, nil)
	assert.Contains(t, p.Output(running, errors.New("boom")), "boom")
}

func TestPopeJoanCuiPresenter_HintRendersEveryShape(t *testing.T) {
	p := new(PopeJoanCuiPresenter)

	play := new(interfaces.MockPopeJoanGame)
	play.On("GetGameEndFlag").Return(false)
	play.On("GetPhase").Return(domain.PopeJoanPhasePlay)
	play.On("GetCurrentPlayerIdx").Return(0)
	play.On("GetRunSuit").Return(-1)
	play.On("PopeJoanCpuDecide", 0).Return(3)
	assert.Contains(t, p.HintOutput(play), "3")

	none := new(interfaces.MockPopeJoanGame)
	none.On("GetGameEndFlag").Return(true)
	assert.NotEmpty(t, p.HintOutput(none))
}

func TestPopeJoanCuiPresenter_HintReasonKeysAreAllMapped(t *testing.T) {
	for _, reason := range []string{
		"popejoan.hint.game_end", "popejoan.hint.deal_end", "popejoan.hint.not_your_turn",
		"popejoan.hint.lead", "popejoan.hint.follow", "popejoan.hint.none",
	} {
		assert.NotEmpty(t, popeJoanHintReasonKeys[reason], "unmapped reason %s", reason)
	}
}

func TestPopeJoanCuiPresenter_ActionLog(t *testing.T) {
	assert.NotEmpty(t, new(PopeJoanCuiPresenter).ActionLogOutput(pjTestGame(t)))
}

// **出せる札を応答に載せる。**フロントは押して弾かれるまで違反を示せない (#4934)。
func TestPopeJoanWebPresenter_ValidPlays(t *testing.T) {
	decode := func(raw string) controller.PopeJoanWebOutput {
		var parsed controller.PopeJoanWebOutput
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		return parsed
	}

	g := domain.NewDefaultPopeJoan()
	g.Reset()
	raw := new(PopeJoanWebPresenter).Output(g, nil)
	got := decode(raw).ValidPlays
	// **null ではなく空配列。**フロントが `?? []` を忘れても壊れない。
	if got == nil {
		t.Fatal("validPlays must be [] rather than null")
	}
	// 応答に必ず現れる。
	if !strings.Contains(raw, `"validPlays"`) {
		t.Fatalf("validPlays missing from the response: %s", raw)
	}
}
