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

func njTestGame(t *testing.T) *domain.NainJaune {
	t.Helper()
	n := domain.NewDefaultNainJaune()
	n.Reset()
	return n
}

func njDecode(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

func njpCard(design, value int) *domain.Card { return domain.NewCard(design, value, true) }

// njAntedBoard returns a board with one player's graduated ante on it.
func njAntedBoard() domain.NainJauneBoard {
	var b domain.NainJauneBoard
	b.Ante(1)
	return b
}

// njStub wires a MockNainJauneGame with every accessor the presenters touch.
func njStub(phase domain.NainJaunePhase, gameEnd bool, winner int, awards []*domain.NainJauneAward) *interfaces.MockNainJauneGame {
	g := new(interfaces.MockNainJauneGame)
	g.On("GetPhase").Return(phase)
	g.On("GetGameEndFlag").Return(gameEnd)
	g.On("GetWinnerIdx").Return(winner)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetBoard").Return(njAntedBoard())
	g.On("GetTalonCount").Return(4)
	g.On("GetAwards").Return(awards)
	g.On("GetPlayedPile").Return([]*domain.Card{})
	g.On("GetRunRank").Return(0)
	g.On("GetDealNumber").Return(0)
	g.On("GetDealWinner").Return(-1)
	g.On("GetConfig").Return(domain.DefaultNainJauneConfig())
	players := make([]*domain.NainJaunePlayer, 0, domain.NainJaunePlayerCnt)
	players = append(players, domain.NewNainJaunePlayer(true))
	for range domain.NainJaunePlayerCnt - 1 {
		players = append(players, domain.NewNainJaunePlayer(false))
	}
	g.On("GetPlayers").Return(players)
	g.On("GetPlayer", mock.Anything).Return(domain.NewNainJaunePlayer(false))
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	g.On("NainJauneCpuDecide", mock.Anything).Return(-1)
	return g
}

func TestNainJauneWebPresenter_HidesTheCpuHandButNeverThePoints(t *testing.T) {
	// **支払いは枚数ではなく点数。**枚数だけでは相手の負債額が読めない。
	out := njDecode(t, new(NainJauneWebPresenter).Output(njTestGame(t), nil))
	players, ok := out["players"].([]any)
	require.True(t, ok)
	require.Len(t, players, domain.NainJaunePlayerCnt)

	human, _ := players[0].(map[string]any)
	assert.False(t, human["hidden"].(bool))
	assert.NotEmpty(t, human["cards"])

	cpu, _ := players[1].(map[string]any)
	assert.True(t, cpu["hidden"].(bool))
	assert.Empty(t, cpu["cards"], "the opponent's hand must not reach the browser")
	assert.Positive(t, cpu["cardCount"], "but its size is public")
	assert.Positive(t, cpu["points"], "and so is what it is worth")
}

// TestNainJauneWebPresenter_ShipsAllFiveBoxesWithTheirCard is what the board is
// for. The card matters as much as the chips: **only the exact suit claims a
// box**, and the issue named two of them with the wrong suit.
func TestNainJauneWebPresenter_ShipsAllFiveBoxesWithTheirCard(t *testing.T) {
	g := njStub(domain.NainJaunePhasePlay, false, -1, nil)
	out := njDecode(t, new(NainJauneWebPresenter).Output(g, nil))
	boxes, ok := out["boxes"].([]any)
	require.True(t, ok)
	require.Len(t, boxes, int(domain.NainJauneBoxCount))

	want := []struct {
		name   string
		chips  float64
		design string
		value  float64
	}{
		{"ten", 1, "DIAMOND", 10},
		{"jack", 2, "CLOVER", 11},
		{"queen", 3, "SPADE", 12},
		{"king", 4, "HEART", 13},
		{"dwarf", 5, "DIAMOND", 7},
	}
	for i, w := range want {
		box, _ := boxes[i].(map[string]any)
		assert.Equal(t, w.name, box["name"], "box %d", i)
		// **アンティは均等ではない。**♦10 に 1 … ♦7 に 5。
		assert.Equal(t, w.chips, box["chips"], "box %s", w.name)
		card, _ := box["card"].(map[string]any)
		require.NotNil(t, card, "box %s has no card", w.name)
		assert.Equal(t, w.design, card["design"], "box %s suit", w.name)
		assert.Equal(t, w.value, card["value"], "box %s rank", w.name)
	}
}

func TestNainJauneWebPresenter_ShipsTheAwardsAndTalon(t *testing.T) {
	awards := []*domain.NainJauneAward{
		{Box: domain.NainJauneBoxDwarf, Player: 2, Chips: 20},
	}
	g := njStub(domain.NainJaunePhasePlay, false, -1, awards)
	out := njDecode(t, new(NainJauneWebPresenter).Output(g, nil))
	got, ok := out["awards"].([]any)
	require.True(t, ok)
	require.Len(t, got, 1)
	first, _ := got[0].(map[string]any)
	assert.Equal(t, "dwarf", first["box"])
	assert.Equal(t, float64(20), first["chips"])
	assert.Equal(t, float64(4), out["talonCount"])
}

// TestNainJauneWebPresenter_KeepsAStoppedRunDistinct は 0 の意味を守る。
func TestNainJauneWebPresenter_KeepsAStoppedRunDistinct(t *testing.T) {
	stopped := njStub(domain.NainJaunePhasePlay, false, -1, nil)
	assert.Equal(t, float64(0), njDecode(t, new(NainJauneWebPresenter).Output(stopped, nil))["runRank"])

	live := njStub(domain.NainJaunePhasePlay, false, -1, nil)
	live.On("GetRunRank").Unset()
	live.On("GetRunRank").Return(9)
	assert.Equal(t, float64(9), njDecode(t, new(NainJauneWebPresenter).Output(live, nil))["runRank"])
}

func TestNainJauneWebPresenter_CarriesTheHintOnAPlainStateResponse(t *testing.T) {
	g := njStub(domain.NainJaunePhasePlay, false, -1, nil)
	g.On("NainJauneCpuDecide", 0).Unset()
	g.On("NainJauneCpuDecide", 0).Return(1)
	g.On("GetPlayer", 0).Unset()
	g.On("GetPlayer", 0).Return(domain.NewNainJaunePlayer(true))
	assert.NotNil(t, njDecode(t, new(NainJauneWebPresenter).Output(g, nil))["hint"],
		"the hint toggle reads the ordinary response")
}

func TestNainJauneWebPresenter_HintReasonsCoverEveryBranch(t *testing.T) {
	// 区画つきの札を持つ席を組み立てる。♦7 は 5 枚積まれている。
	withDwarf := domain.NewNainJaunePlayer(true)
	withDwarf.AddCard(njpCard(domain.CardDesignDiamond, 7))

	for name, tc := range map[string]struct {
		gameEnd bool
		phase   domain.NainJaunePhase
		current int
		runRank int
		decide  int
		player  *domain.NainJaunePlayer
		want    string
	}{
		"game over":     {true, domain.NainJaunePhaseGameEnd, 0, 0, 0, nil, "nainjaune.hint.game_end"},
		"deal over":     {false, domain.NainJaunePhaseDealEnd, 0, 0, 0, nil, "nainjaune.hint.deal_end"},
		"not your turn": {false, domain.NainJaunePhasePlay, 1, 0, 0, nil, "nainjaune.hint.not_your_turn"},
		"nothing":       {false, domain.NainJaunePhasePlay, 0, 0, -1, nil, "nainjaune.hint.none"},
		"lead":          {false, domain.NainJaunePhasePlay, 0, 0, 0, domain.NewNainJaunePlayer(true), "nainjaune.hint.lead"},
		"follow":        {false, domain.NainJaunePhasePlay, 0, 5, 0, domain.NewNainJaunePlayer(true), "nainjaune.hint.follow"},
		"claims a box":  {false, domain.NainJaunePhasePlay, 0, 6, 0, withDwarf, "nainjaune.hint.box"},
	} {
		t.Run(name, func(t *testing.T) {
			g := new(interfaces.MockNainJauneGame)
			g.On("GetGameEndFlag").Return(tc.gameEnd)
			g.On("GetPhase").Return(tc.phase)
			g.On("GetCurrentPlayerIdx").Return(tc.current)
			g.On("GetRunRank").Return(tc.runRank)
			g.On("NainJauneCpuDecide", 0).Return(tc.decide)
			g.On("GetBoard").Return(njAntedBoard())
			g.On("GetPlayer", 0).Return(tc.player)

			assert.Equal(t, tc.want, nainJauneHint(g).Reason)
		})
	}
}

func TestNainJauneWebPresenter_MessageCodes(t *testing.T) {
	win := njStub(domain.NainJaunePhaseGameEnd, true, 0, nil)
	assert.Equal(t, "nainjaune.win", njDecode(t, new(NainJauneWebPresenter).Output(win, nil))["messageCode"])

	lose := njStub(domain.NainJaunePhaseGameEnd, true, 2, nil)
	assert.Equal(t, "nainjaune.lose", njDecode(t, new(NainJauneWebPresenter).Output(lose, nil))["messageCode"])

	running := njStub(domain.NainJaunePhasePlay, false, -1, nil)
	assert.Empty(t, njDecode(t, new(NainJauneWebPresenter).Output(running, nil))["message"])
	assert.Equal(t, "boom", njDecode(t, new(NainJauneWebPresenter).Output(running, errors.New("boom")))["message"])
}

func TestNainJauneWebPresenter_HintOutputAndActionLog(t *testing.T) {
	n := njTestGame(t)
	assert.NotNil(t, njDecode(t, new(NainJauneWebPresenter).HintOutput(n))["hint"])
	assert.NotEmpty(t, new(NainJauneWebPresenter).ActionLogOutput(n))
}

func TestNainJauneCuiPresenter_ShowsTheRulesAndTheBoard(t *testing.T) {
	out := new(NainJauneCuiPresenter).Output(njTestGame(t), nil)
	assert.Contains(t, out, i18n.T("nainjaune.ruleLine"))
	assert.Contains(t, out, "[0]", "your hand is indexed")
	for _, name := range []string{"ten", "jack", "queen", "king", "dwarf"} {
		assert.Contains(t, out, i18n.T("nainjaune.box."+name), "box %s missing", name)
	}
}

func TestNainJauneCuiPresenter_ReportsAnAward(t *testing.T) {
	g := njStub(domain.NainJaunePhasePlay, false, -1, []*domain.NainJauneAward{
		{Box: domain.NainJauneBoxDwarf, Player: 1, Chips: 20},
	})
	// **名前は ANSI で装飾される**ので、名前より後ろの部分で照合する。
	tail := strings.SplitN(i18n.Tf("nainjaune.awardLine",
		"name", "@@@", "box", i18n.T("nainjaune.box.dwarf"), "chips", "20"), "@@@", 2)[1]
	assert.Contains(t, new(NainJauneCuiPresenter).Output(g, nil), tail)
}

// TestNainJauneCuiPresenter_TellsAStoppedRunApart は、止まっているかどうかで
// 案内が変わることを確かめる。続きは**ランクだけ**で決まる。
func TestNainJauneCuiPresenter_TellsAStoppedRunApart(t *testing.T) {
	stopped := njStub(domain.NainJaunePhasePlay, false, -1, nil)
	out := new(NainJauneCuiPresenter).promptBlock(stopped)
	assert.Contains(t, out, i18n.T("nainjaune.promptLead"))

	live := njStub(domain.NainJaunePhasePlay, false, -1, nil)
	live.On("GetRunRank").Unset()
	live.On("GetRunRank").Return(5)
	liveOut := new(NainJauneCuiPresenter).promptBlock(live)
	// 次に出せるのは 6。スートは問わない。
	assert.Contains(t, liveOut, i18n.Tf("nainjaune.promptFollow", "rank", "6"))
	assert.NotContains(t, liveOut, i18n.T("nainjaune.promptLead"))
}

func TestNainJauneCuiPresenter_PromptsAtTheEndOfADeal(t *testing.T) {
	dealEnd := njStub(domain.NainJaunePhaseDealEnd, false, -1, nil)
	dealEnd.On("GetDealWinner").Unset()
	dealEnd.On("GetDealWinner").Return(2)
	assert.Contains(t, new(NainJauneCuiPresenter).promptBlock(dealEnd), i18n.T("nainjaune.promptNext"))
}

func TestNainJauneCuiPresenter_BannersAndErrors(t *testing.T) {
	p := new(NainJauneCuiPresenter)
	win := njStub(domain.NainJaunePhaseGameEnd, true, 0, nil)
	lose := njStub(domain.NainJaunePhaseGameEnd, true, 2, nil)
	assert.NotEqual(t, p.Output(win, nil), p.Output(lose, nil), "the two endings must not read the same")

	running := njStub(domain.NainJaunePhasePlay, false, -1, nil)
	assert.Contains(t, p.Output(running, errors.New("boom")), "boom")
}

func TestNainJauneCuiPresenter_HintRendersEveryShape(t *testing.T) {
	p := new(NainJauneCuiPresenter)

	play := new(interfaces.MockNainJauneGame)
	play.On("GetGameEndFlag").Return(false)
	play.On("GetPhase").Return(domain.NainJaunePhasePlay)
	play.On("GetCurrentPlayerIdx").Return(0)
	play.On("GetRunRank").Return(0)
	play.On("NainJauneCpuDecide", 0).Return(3)
	play.On("GetBoard").Return(njAntedBoard())
	play.On("GetPlayer", 0).Return(domain.NewNainJaunePlayer(true))
	assert.Contains(t, p.HintOutput(play), "3")

	over := new(interfaces.MockNainJauneGame)
	over.On("GetGameEndFlag").Return(true)
	assert.NotEmpty(t, p.HintOutput(over))
}

func TestNainJauneCuiPresenter_HintReasonKeysAreAllMapped(t *testing.T) {
	for _, reason := range []string{
		"nainjaune.hint.game_end", "nainjaune.hint.deal_end", "nainjaune.hint.not_your_turn",
		"nainjaune.hint.lead", "nainjaune.hint.follow", "nainjaune.hint.box", "nainjaune.hint.none",
	} {
		assert.NotEmpty(t, nainJauneHintReasonKeys[reason], "unmapped reason %s", reason)
	}
}

func TestNainJauneCuiPresenter_ActionLog(t *testing.T) {
	assert.NotEmpty(t, new(NainJauneCuiPresenter).ActionLogOutput(njTestGame(t)))
}
