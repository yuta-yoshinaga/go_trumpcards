//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func dsTestGame(t *testing.T) *domain.Desmoche {
	t.Helper()
	d := domain.NewDefaultDesmoche()
	d.Reset()
	return d
}

func dsDecode(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

// dsStub wires a MockDesmocheGame with every accessor the presenters touch.
func dsStub(phase domain.DesmochePhase, gameEnd bool, winner int, melds []*domain.DesmocheMeld) *interfaces.MockDesmocheGame {
	g := new(interfaces.MockDesmocheGame)
	g.On("GetPhase").Return(phase)
	g.On("GetGameEndFlag").Return(gameEnd)
	g.On("GetWinnerIdx").Return(winner)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetStockCount").Return(15)
	g.On("GetDiscardTop").Return(domain.NewCard(domain.CardDesignHeart, 9, true))
	g.On("GetMelds").Return(melds)
	g.On("MeldedCount", mock.Anything).Return(3)
	g.On("GetPot").Return(40)
	g.On("GetScore", mock.Anything).Return(-10)
	g.On("GetRoundNumber").Return(1)
	g.On("GetRoundWinner").Return(-1)
	g.On("IsRoundExhausted").Return(false)
	g.On("GetConfig").Return(domain.DefaultDesmocheConfig())
	players := make([]*domain.DesmochePlayer, 0, domain.DesmochePlayerCnt)
	players = append(players, domain.NewDesmochePlayer(true))
	for range domain.DesmochePlayerCnt - 1 {
		players = append(players, domain.NewDesmochePlayer(false))
	}
	g.On("GetPlayers").Return(players)
	g.On("GetPlayer", mock.Anything).Return(domain.NewDesmochePlayer(false))
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	g.On("DesmocheCpuDecide", mock.Anything).Return(domain.DesmocheCpuAction{DiscardIdx: 2})
	return g
}

func TestDesmocheWebPresenter_HidesTheCpuHandButNeverTheProgress(t *testing.T) {
	// メルドは表向きに並ぶ規則なので、10 枚まであと何枚かは隠せない。
	out := dsDecode(t, new(DesmocheWebPresenter).Output(dsTestGame(t), nil))
	players, ok := out["players"].([]any)
	require.True(t, ok)
	require.Len(t, players, domain.DesmochePlayerCnt)

	human, _ := players[0].(map[string]any)
	assert.False(t, human["hidden"].(bool))
	assert.NotEmpty(t, human["cards"])

	cpu, _ := players[1].(map[string]any)
	assert.True(t, cpu["hidden"].(bool))
	assert.Empty(t, cpu["cards"], "the opponent's hand must not reach the browser")
	assert.Positive(t, cpu["cardCount"], "but its size is public")
	assert.NotNil(t, cpu["meldedCount"], "and so is how far down it is")
	assert.Equal(t, float64(domain.DesmocheGoOutSize), out["goOutSize"])
}

func TestDesmocheWebPresenter_ShipsThePotSoTheCarryOverIsVisible(t *testing.T) {
	// **勝者なしのラウンドではポットが持ち越される。**クライアントに掛け金を
	// 積算させると必ずずれるので、額そのものを送る。
	g := dsStub(domain.DesmochePhaseAct, false, -1, nil)
	out := dsDecode(t, new(DesmocheWebPresenter).Output(g, nil))
	assert.Equal(t, float64(40), out["pot"])
	assert.Equal(t, false, out["roundExhausted"])
}

func TestDesmocheWebPresenter_ShipsTheMeldsWithTheirKind(t *testing.T) {
	// セットとランで付けられる札が違うので、種別を送らないとクライアントが
	// 判定を再実装することになる。
	melds := []*domain.DesmocheMeld{
		{Owner: 1, Kind: domain.DesmocheMeldSet, Cards: []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 7, true),
			domain.NewCard(domain.CardDesignHeart, 7, true),
			domain.NewCard(domain.CardDesignClover, 7, true),
		}},
	}
	g := dsStub(domain.DesmochePhaseAct, false, -1, melds)
	out := dsDecode(t, new(DesmocheWebPresenter).Output(g, nil))
	got, ok := out["melds"].([]any)
	require.True(t, ok)
	require.Len(t, got, 1)
	first, _ := got[0].(map[string]any)
	assert.Equal(t, float64(1), first["owner"])
	assert.Equal(t, float64(domain.DesmocheMeldSet), first["kind"])
	assert.Len(t, first["cards"], 3)
}

func TestDesmocheWebPresenter_CarriesTheHintOnAPlainStateResponse(t *testing.T) {
	g := dsStub(domain.DesmochePhaseAct, false, -1, nil)
	assert.NotNil(t, dsDecode(t, new(DesmocheWebPresenter).Output(g, nil))["hint"],
		"the hint toggle reads the ordinary response")
}

func TestDesmocheWebPresenter_HintReasonsCoverEveryBranch(t *testing.T) {
	for name, tc := range map[string]struct {
		gameEnd bool
		current int
		phase   domain.DesmochePhase
		action  domain.DesmocheCpuAction
		want    string
	}{
		"game over":     {true, 0, domain.DesmochePhaseAct, domain.DesmocheCpuAction{}, "desmoche.hint.game_end"},
		"not your turn": {false, 1, domain.DesmochePhaseAct, domain.DesmocheCpuAction{}, "desmoche.hint.not_your_turn"},
		"draw":          {false, 0, domain.DesmochePhaseDraw, domain.DesmocheCpuAction{}, "desmoche.hint.draw"},
		"meld":          {false, 0, domain.DesmochePhaseAct, domain.DesmocheCpuAction{MeldIdxs: []int{0, 1, 2}, DiscardIdx: -1}, "desmoche.hint.meld"},
		"discard":       {false, 0, domain.DesmochePhaseAct, domain.DesmocheCpuAction{DiscardIdx: 3}, "desmoche.hint.discard"},
		"nothing":       {false, 0, domain.DesmochePhaseAct, domain.DesmocheCpuAction{DiscardIdx: -1}, "desmoche.hint.none"},
	} {
		t.Run(name, func(t *testing.T) {
			g := new(interfaces.MockDesmocheGame)
			g.On("GetGameEndFlag").Return(tc.gameEnd)
			g.On("GetCurrentPlayerIdx").Return(tc.current)
			g.On("GetPhase").Return(tc.phase)
			g.On("DesmocheCpuDecide", 0).Return(tc.action)

			assert.Equal(t, tc.want, desmocheHint(g).Reason)
		})
	}
}

func TestDesmocheWebPresenter_MessageCodes(t *testing.T) {
	win := dsStub(domain.DesmochePhaseGameEnd, true, 0, nil)
	assert.Equal(t, "desmoche.win", dsDecode(t, new(DesmocheWebPresenter).Output(win, nil))["messageCode"])

	lose := dsStub(domain.DesmochePhaseGameEnd, true, 2, nil)
	assert.Equal(t, "desmoche.lose", dsDecode(t, new(DesmocheWebPresenter).Output(lose, nil))["messageCode"])

	running := dsStub(domain.DesmochePhaseAct, false, -1, nil)
	assert.Empty(t, dsDecode(t, new(DesmocheWebPresenter).Output(running, nil))["message"])
	assert.Equal(t, "boom", dsDecode(t, new(DesmocheWebPresenter).Output(running, errors.New("boom")))["message"])
}

func TestDesmocheWebPresenter_HintOutputAndActionLog(t *testing.T) {
	d := dsTestGame(t)
	assert.NotNil(t, dsDecode(t, new(DesmocheWebPresenter).HintOutput(d))["hint"])
	assert.NotEmpty(t, new(DesmocheWebPresenter).ActionLogOutput(d))
}

func TestDesmocheCuiPresenter_ShowsTheRulesAndTheDiscard(t *testing.T) {
	out := new(DesmocheCuiPresenter).Output(dsTestGame(t), nil)
	assert.Contains(t, out, i18n.T("desmoche.ruleLine"))
	assert.Contains(t, out, "[0]", "your hand is indexed")
}

func TestDesmocheCuiPresenter_NamesTheMeldKind(t *testing.T) {
	// 種別が出ていないと、付けられる札が判断できない。
	melds := []*domain.DesmocheMeld{
		{Owner: 1, Kind: domain.DesmocheMeldRun, Cards: []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, true),
			domain.NewCard(domain.CardDesignSpade, 6, true),
			domain.NewCard(domain.CardDesignSpade, 7, true),
		}},
	}
	g := dsStub(domain.DesmochePhaseAct, false, -1, melds)
	out := new(DesmocheCuiPresenter).Output(g, nil)
	assert.Contains(t, out, i18n.Tf("desmoche.meldLine",
		"idx", "0", "kind", i18n.T("desmoche.run"), "owner", "1",
		"cards", "SPADE 5 SPADE 6 SPADE 7"))
}

func TestDesmocheCuiPresenter_PromptsPerPhase(t *testing.T) {
	p := new(DesmocheCuiPresenter)

	draw := dsStub(domain.DesmochePhaseDraw, false, -1, nil)
	drawOut := p.Output(draw, nil)
	assert.Contains(t, drawOut, i18n.T("desmoche.promptDraw"))
	assert.NotContains(t, drawOut, i18n.T("desmoche.promptAct"))

	act := dsStub(domain.DesmochePhaseAct, false, -1, nil)
	assert.Contains(t, p.Output(act, nil), i18n.T("desmoche.promptAct"))
}

// TestDesmocheCuiPresenter_SaysWhenNobodyWon は、**勝者なしという結末**が読める
// ことを確かめる。何も出ないとポットが増えた理由が分からない。
func TestDesmocheCuiPresenter_SaysWhenNobodyWon(t *testing.T) {
	exhausted := new(interfaces.MockDesmocheGame)
	exhausted.On("GetGameEndFlag").Return(false)
	exhausted.On("GetPhase").Return(domain.DesmochePhaseRoundEnd)
	exhausted.On("GetRoundWinner").Return(-1)
	exhausted.On("GetPot").Return(80)
	out := new(DesmocheCuiPresenter).promptBlock(exhausted)
	assert.Contains(t, out, i18n.Tf("desmoche.roundNoWinner", "pot", "80"))

	won := new(interfaces.MockDesmocheGame)
	won.On("GetGameEndFlag").Return(false)
	won.On("GetPhase").Return(domain.DesmochePhaseRoundEnd)
	won.On("GetRoundWinner").Return(1)
	won.On("GetPlayer", 1).Return(domain.NewDesmochePlayer(false))
	assert.NotContains(t, new(DesmocheCuiPresenter).promptBlock(won),
		i18n.Tf("desmoche.roundNoWinner", "pot", "80"))
}

func TestDesmocheCuiPresenter_BannersAndErrors(t *testing.T) {
	p := new(DesmocheCuiPresenter)
	win := dsStub(domain.DesmochePhaseGameEnd, true, 0, nil)
	lose := dsStub(domain.DesmochePhaseGameEnd, true, 2, nil)
	assert.NotEqual(t, p.Output(win, nil), p.Output(lose, nil), "the two endings must not read the same")

	running := dsStub(domain.DesmochePhaseAct, false, -1, nil)
	assert.Contains(t, p.Output(running, errors.New("boom")), "boom")
}

func TestDesmocheCuiPresenter_HintRendersEveryShape(t *testing.T) {
	p := new(DesmocheCuiPresenter)

	draw := new(interfaces.MockDesmocheGame)
	draw.On("GetGameEndFlag").Return(false)
	draw.On("GetCurrentPlayerIdx").Return(0)
	draw.On("GetPhase").Return(domain.DesmochePhaseDraw)
	assert.NotEmpty(t, p.HintOutput(draw))

	meld := new(interfaces.MockDesmocheGame)
	meld.On("GetGameEndFlag").Return(false)
	meld.On("GetCurrentPlayerIdx").Return(0)
	meld.On("GetPhase").Return(domain.DesmochePhaseAct)
	meld.On("DesmocheCpuDecide", 0).Return(domain.DesmocheCpuAction{MeldIdxs: []int{0, 1, 2}, DiscardIdx: -1})
	assert.Contains(t, p.HintOutput(meld), "0,1,2")

	discard := new(interfaces.MockDesmocheGame)
	discard.On("GetGameEndFlag").Return(false)
	discard.On("GetCurrentPlayerIdx").Return(0)
	discard.On("GetPhase").Return(domain.DesmochePhaseAct)
	discard.On("DesmocheCpuDecide", 0).Return(domain.DesmocheCpuAction{DiscardIdx: 3})
	assert.Contains(t, p.HintOutput(discard), "3")

	over := new(interfaces.MockDesmocheGame)
	over.On("GetGameEndFlag").Return(true)
	assert.NotEmpty(t, p.HintOutput(over))
}

func TestDesmocheCuiPresenter_HintReasonKeysAreAllMapped(t *testing.T) {
	for _, reason := range []string{
		"desmoche.hint.game_end", "desmoche.hint.not_your_turn", "desmoche.hint.draw",
		"desmoche.hint.meld", "desmoche.hint.discard", "desmoche.hint.none",
	} {
		assert.NotEmpty(t, desmocheHintReasonKeys[reason], "unmapped reason %s", reason)
	}
}

func TestDesmocheCuiPresenter_ActionLog(t *testing.T) {
	assert.NotEmpty(t, new(DesmocheCuiPresenter).ActionLogOutput(dsTestGame(t)))
}

// **10 枚上がりが勝利条件なのに、他家のメルドへ付けても自分の枚数は増えない。**
// Web は foreignMeldWarning で警告しているが、CUI には対応する文言が無く、
// `o <i> <m>` を打つ人はそれに気づけなかった (#5720)。
func TestDesmocheCuiPresenter_WarnsThatForeignLayoffsDoNotCount(t *testing.T) {
	melds := []*domain.DesmocheMeld{
		{Owner: 1, Kind: domain.DesmocheMeldRun, Cards: []*domain.Card{
			domain.NewCard(domain.CardDesignSpade, 5, true),
			domain.NewCard(domain.CardDesignSpade, 6, true),
			domain.NewCard(domain.CardDesignSpade, 7, true),
		}},
	}
	g := dsStub(domain.DesmochePhaseAct, false, -1, melds)
	out := new(DesmocheCuiPresenter).Output(g, nil)

	assert.Contains(t, out, i18n.T("desmoche.promptActLayoffNote"))
	// コマンド一覧そのものは残っていること。注記で置き換えては打ち方が消える。
	assert.Contains(t, out, i18n.T("desmoche.promptAct"))

	// 英語ロケールでも出て、日本語が漏れないこと。
	i18n.SetLang("en")
	defer i18n.SetLang("ja")
	en := new(DesmocheCuiPresenter).Output(g, nil)
	assert.Contains(t, en, "does not count toward your 10")
	assert.NotContains(t, en, "他家のメルド")
}
