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

func lbTestGame(t *testing.T) *domain.Loba {
	t.Helper()
	l := domain.NewDefaultLoba()
	l.Reset()
	return l
}

func lbDecode(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

// lbStub wires a MockLobaGame with every accessor the presenters touch.
func lbStub(phase domain.LobaPhase, gameEnd bool, winner int, melds []*domain.LobaMeld) *interfaces.MockLobaGame {
	g := new(interfaces.MockLobaGame)
	g.On("GetPhase").Return(phase)
	g.On("GetGameEndFlag").Return(gameEnd)
	g.On("GetWinnerIdx").Return(winner)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetStockCount").Return(70)
	g.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 9, true))
	g.On("GetMelds").Return(melds)
	g.On("HasMelded", mock.Anything).Return(false)
	g.On("GetScore", mock.Anything).Return(12)
	g.On("IsEliminated", mock.Anything).Return(false)
	g.On("GetRoundNumber").Return(1)
	g.On("GetRoundWinner").Return(-1)
	g.On("IsRoundClean").Return(false)
	g.On("GetConfig").Return(domain.DefaultLobaConfig())
	players := make([]*domain.LobaPlayer, 0, domain.LobaPlayerCnt)
	players = append(players, domain.NewLobaPlayer(true))
	for range domain.LobaPlayerCnt - 1 {
		players = append(players, domain.NewLobaPlayer(false))
	}
	g.On("GetPlayers").Return(players)
	g.On("GetPlayer", mock.Anything).Return(domain.NewLobaPlayer(false))
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	g.On("LobaCpuDecide", mock.Anything).Return(domain.LobaCpuAction{DiscardIdx: 2})
	return g
}

func TestLobaWebPresenter_HidesTheCpuHandButNeverTheScore(t *testing.T) {
	// 101 で脱落するので、誰があと何点なのかが最大の判断材料。
	out := lbDecode(t, new(LobaWebPresenter).Output(lbTestGame(t), nil))
	players, ok := out["players"].([]any)
	require.True(t, ok)
	require.Len(t, players, domain.LobaPlayerCnt)

	human, _ := players[0].(map[string]any)
	assert.False(t, human["hidden"].(bool))
	assert.NotEmpty(t, human["cards"])

	cpu, _ := players[1].(map[string]any)
	assert.True(t, cpu["hidden"].(bool))
	assert.Empty(t, cpu["cards"], "the opponent's hand must not reach the browser")
	assert.Positive(t, cpu["cardCount"], "but its size is public")
	assert.NotNil(t, cpu["score"], "and so is its score")
	assert.Equal(t, float64(domain.LobaKnockOut), out["knockOut"])
}

func TestLobaWebPresenter_ShipsTheMeldsWithTheirKind(t *testing.T) {
	// ピエルナとエスカレラで付けられる札が違うので、種別を送らないと
	// クライアントが判定を再実装することになる。
	melds := []*domain.LobaMeld{
		{Owner: 1, Kind: domain.LobaMeldPierna, Cards: []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 7, true),
			domain.NewCard(domain.CardDesignHeart, 7, true),
			domain.NewCard(domain.CardDesignClover, 7, true),
		}},
	}
	g := lbStub(domain.LobaPhaseAct, false, -1, melds)
	out := lbDecode(t, new(LobaWebPresenter).Output(g, nil))
	got, ok := out["melds"].([]any)
	require.True(t, ok)
	require.Len(t, got, 1)
	first, _ := got[0].(map[string]any)
	assert.Equal(t, float64(1), first["owner"])
	assert.Equal(t, float64(domain.LobaMeldPierna), first["kind"])
	assert.Len(t, first["cards"], 3)
}

func TestLobaWebPresenter_CarriesTheHintOnAPlainStateResponse(t *testing.T) {
	g := lbStub(domain.LobaPhaseAct, false, -1, nil)
	assert.NotNil(t, lbDecode(t, new(LobaWebPresenter).Output(g, nil))["hint"],
		"the hint toggle reads the ordinary response")
}

func TestLobaWebPresenter_HintReasonsCoverEveryBranch(t *testing.T) {
	for name, tc := range map[string]struct {
		gameEnd bool
		current int
		phase   domain.LobaPhase
		action  domain.LobaCpuAction
		want    string
	}{
		"game over":     {true, 0, domain.LobaPhaseAct, domain.LobaCpuAction{}, "loba.hint.game_end"},
		"not your turn": {false, 1, domain.LobaPhaseAct, domain.LobaCpuAction{}, "loba.hint.not_your_turn"},
		"draw":          {false, 0, domain.LobaPhaseDraw, domain.LobaCpuAction{}, "loba.hint.draw"},
		"meld":          {false, 0, domain.LobaPhaseAct, domain.LobaCpuAction{MeldIdxs: []int{0, 1, 2}, DiscardIdx: -1}, "loba.hint.meld"},
		"discard":       {false, 0, domain.LobaPhaseAct, domain.LobaCpuAction{DiscardIdx: 3}, "loba.hint.discard"},
		"nothing":       {false, 0, domain.LobaPhaseAct, domain.LobaCpuAction{DiscardIdx: -1}, "loba.hint.none"},
	} {
		t.Run(name, func(t *testing.T) {
			g := new(interfaces.MockLobaGame)
			g.On("GetGameEndFlag").Return(tc.gameEnd)
			g.On("GetCurrentPlayerIdx").Return(tc.current)
			g.On("GetPhase").Return(tc.phase)
			g.On("LobaCpuDecide", 0).Return(tc.action)

			assert.Equal(t, tc.want, lobaHint(g).Reason)
		})
	}
}

func TestLobaWebPresenter_MessageCodes(t *testing.T) {
	win := lbStub(domain.LobaPhaseGameEnd, true, 0, nil)
	assert.Equal(t, "loba.win", lbDecode(t, new(LobaWebPresenter).Output(win, nil))["messageCode"])

	lose := lbStub(domain.LobaPhaseGameEnd, true, 2, nil)
	assert.Equal(t, "loba.lose", lbDecode(t, new(LobaWebPresenter).Output(lose, nil))["messageCode"])

	running := lbStub(domain.LobaPhaseAct, false, -1, nil)
	assert.Empty(t, lbDecode(t, new(LobaWebPresenter).Output(running, nil))["message"])
	assert.Equal(t, "boom", lbDecode(t, new(LobaWebPresenter).Output(running, errors.New("boom")))["message"])
}

func TestLobaWebPresenter_HintOutputAndActionLog(t *testing.T) {
	l := lbTestGame(t)
	assert.NotNil(t, lbDecode(t, new(LobaWebPresenter).HintOutput(l))["hint"])
	assert.NotEmpty(t, new(LobaWebPresenter).ActionLogOutput(l))
}

func TestLobaCuiPresenter_ShowsTheRulesAndTheDiscard(t *testing.T) {
	out := new(LobaCuiPresenter).Output(lbTestGame(t), nil)
	assert.Contains(t, out, i18n.T("loba.ruleLine"))
	assert.Contains(t, out, "[0]", "your hand is indexed")
}

func TestLobaCuiPresenter_NamesTheMeldKind(t *testing.T) {
	// 種別が出ていないと、付けられる札が判断できない。
	melds := []*domain.LobaMeld{
		{Owner: 1, Kind: domain.LobaMeldEscalera, Cards: []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, true),
			domain.NewCard(domain.CardDesignSpade, 6, true),
			domain.NewCard(domain.CardDesignSpade, 7, true),
		}},
	}
	g := lbStub(domain.LobaPhaseAct, false, -1, melds)
	out := new(LobaCuiPresenter).Output(g, nil)
	// 「ピエルナ」は常時表示のルール行にも出るので、出力全体に対する
	// NotContains は原理的に通らない。メルド行そのものを組み立てて照合する。
	assert.Contains(t, out, i18n.Tf("loba.meldLine",
		"idx", "0", "kind", i18n.T("loba.escalera"), "owner", "1",
		"cards", "SPADE 5 SPADE 6 SPADE 7"))
}

func TestLobaCuiPresenter_PromptsPerPhase(t *testing.T) {
	p := new(LobaCuiPresenter)

	draw := lbStub(domain.LobaPhaseDraw, false, -1, nil)
	drawOut := p.Output(draw, nil)
	assert.Contains(t, drawOut, i18n.T("loba.promptDraw"))
	assert.NotContains(t, drawOut, i18n.T("loba.promptAct"))

	act := lbStub(domain.LobaPhaseAct, false, -1, nil)
	assert.Contains(t, p.Output(act, nil), i18n.T("loba.promptAct"))
}

func TestLobaCuiPresenter_TellsACleanGoOutApart(t *testing.T) {
	// -10 が付くかどうかは表示から読めなければならない。
	// 名前は ANSI で装飾されるので、名前を含めた連結文字列にはならない。
	// 文言の名前より後ろの部分で照合する。
	plainTail := strings.SplitN(i18n.Tf("loba.roundWinner", "name", "X"), "X", 2)[1]
	cleanTail := strings.SplitN(i18n.Tf("loba.roundWinnerClean", "name", "X"), "X", 2)[1]
	require.NotEqual(t, plainTail, cleanTail, "the two endings must read differently")

	plain := new(interfaces.MockLobaGame)
	lbWireRoundEnd(plain, false)
	plainOut := new(LobaCuiPresenter).promptBlock(plain)
	assert.Contains(t, plainOut, plainTail)
	assert.NotContains(t, plainOut, cleanTail)

	clean := new(interfaces.MockLobaGame)
	lbWireRoundEnd(clean, true)
	assert.Contains(t, new(LobaCuiPresenter).promptBlock(clean), cleanTail)
}

// lbWireRoundEnd stubs just enough for promptBlock at the end of a round.
func lbWireRoundEnd(g *interfaces.MockLobaGame, clean bool) {
	g.On("GetGameEndFlag").Return(false)
	g.On("GetPhase").Return(domain.LobaPhaseRoundEnd)
	g.On("GetRoundWinner").Return(1)
	g.On("IsRoundClean").Return(clean)
	g.On("GetPlayer", 1).Return(domain.NewLobaPlayer(false))
}

func TestLobaCuiPresenter_BannersAndErrors(t *testing.T) {
	p := new(LobaCuiPresenter)
	win := lbStub(domain.LobaPhaseGameEnd, true, 0, nil)
	lose := lbStub(domain.LobaPhaseGameEnd, true, 2, nil)
	assert.NotEqual(t, p.Output(win, nil), p.Output(lose, nil), "the two endings must not read the same")

	running := lbStub(domain.LobaPhaseAct, false, -1, nil)
	assert.Contains(t, p.Output(running, errors.New("boom")), "boom")
}

func TestLobaCuiPresenter_HintRendersEveryShape(t *testing.T) {
	p := new(LobaCuiPresenter)

	draw := new(interfaces.MockLobaGame)
	draw.On("GetGameEndFlag").Return(false)
	draw.On("GetCurrentPlayerIdx").Return(0)
	draw.On("GetPhase").Return(domain.LobaPhaseDraw)
	assert.NotEmpty(t, p.HintOutput(draw))

	meld := new(interfaces.MockLobaGame)
	meld.On("GetGameEndFlag").Return(false)
	meld.On("GetCurrentPlayerIdx").Return(0)
	meld.On("GetPhase").Return(domain.LobaPhaseAct)
	meld.On("LobaCpuDecide", 0).Return(domain.LobaCpuAction{MeldIdxs: []int{0, 1, 2}, DiscardIdx: -1})
	assert.Contains(t, p.HintOutput(meld), "0,1,2")

	discard := new(interfaces.MockLobaGame)
	discard.On("GetGameEndFlag").Return(false)
	discard.On("GetCurrentPlayerIdx").Return(0)
	discard.On("GetPhase").Return(domain.LobaPhaseAct)
	discard.On("LobaCpuDecide", 0).Return(domain.LobaCpuAction{DiscardIdx: 3})
	assert.Contains(t, p.HintOutput(discard), "3")

	over := new(interfaces.MockLobaGame)
	over.On("GetGameEndFlag").Return(true)
	assert.NotEmpty(t, p.HintOutput(over))
}

func TestLobaCuiPresenter_HintReasonKeysAreAllMapped(t *testing.T) {
	for _, reason := range []string{
		"loba.hint.game_end", "loba.hint.not_your_turn", "loba.hint.draw",
		"loba.hint.meld", "loba.hint.discard", "loba.hint.none",
	} {
		assert.NotEmpty(t, lobaHintReasonKeys[reason], "unmapped reason %s", reason)
	}
}

func TestLobaCuiPresenter_ActionLog(t *testing.T) {
	assert.NotEmpty(t, new(LobaCuiPresenter).ActionLogOutput(lbTestGame(t)))
}
