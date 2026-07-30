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

func lldTestGame(t *testing.T) *domain.LaughAndLieDown {
	t.Helper()
	l := domain.NewDefaultLaughAndLieDown()
	l.Reset()
	return l
}

func lldDecode(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

// lldStub wires a MockLaughAndLieDownGame with every accessor the presenters
// touch, so a test can pin an exact state rather than shuffling until one
// appears.
func lldStub(valid []int, canThree bool, gameEnd bool, score int) *interfaces.MockLaughAndLieDownGame {
	g := new(interfaces.MockLaughAndLieDownGame)
	g.On("GetPhase").Return(domain.LaughAndLieDownPhasePlay)
	g.On("GetGameEndFlag").Return(gameEnd)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetLayout").Return([]*domain.Card{})
	g.On("GetValidPlayIndices", mock.Anything).Return(valid)
	g.On("CanTakeThree", mock.Anything, mock.Anything).Return(canThree)
	g.On("GetWonCount", mock.Anything).Return(8)
	g.On("IsLaidDown", mock.Anything).Return(false)
	g.On("GetScore", mock.Anything).Return(score)
	g.On("GetDealerIdx").Return(0)
	g.On("GetLastInIdx").Return(-1)
	g.On("GetConfig").Return(domain.DefaultLaughAndLieDownConfig())
	players := make([]*domain.LaughAndLieDownPlayer, 0, domain.LaughAndLieDownPlayerCnt)
	players = append(players, domain.NewLaughAndLieDownPlayer(true))
	for range domain.LaughAndLieDownPlayerCnt - 1 {
		players = append(players, domain.NewLaughAndLieDownPlayer(false))
	}
	g.On("GetPlayers").Return(players)
	g.On("GetPlayer", mock.Anything).Return(domain.NewLaughAndLieDownPlayer(false))
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	g.On("LaughAndLieDownCpuDecide", mock.Anything).Return(domain.LaughAndLieDownCpuAction{HandIdx: 0, TakeCount: 1})
	return g
}

func TestLaughAndLieDownWebPresenter_HidesTheCpuHandButNeverItsWonCount(t *testing.T) {
	// 取り札の枚数は 8 との差がそのまま収支なので、隠すと精算が読めなくなる。
	out := lldDecode(t, new(LaughAndLieDownWebPresenter).Output(lldTestGame(t), nil))
	players, ok := out["players"].([]any)
	require.True(t, ok)
	require.Len(t, players, domain.LaughAndLieDownPlayerCnt)

	human, _ := players[0].(map[string]any)
	assert.False(t, human["hidden"].(bool))
	assert.NotEmpty(t, human["cards"])

	cpu, _ := players[1].(map[string]any)
	assert.True(t, cpu["hidden"].(bool))
	assert.Empty(t, cpu["cards"], "the opponent's hand must not reach the browser")
	assert.Positive(t, cpu["cardCount"], "but its size is public")
	assert.NotNil(t, cpu["wonCount"], "and so is what it has captured")
}

func TestLaughAndLieDownWebPresenter_SendsTheWholeFaceUpTable(t *testing.T) {
	// 場は伏せた山ではない。全部送らないと、どのランクが何枚残っているかが
	// 分からず 3 枚取りの判断ができない。
	out := lldDecode(t, new(LaughAndLieDownWebPresenter).Output(lldTestGame(t), nil))
	layout, ok := out["layout"].([]any)
	require.True(t, ok)
	assert.Len(t, layout, domain.LaughAndLieDownLayoutSize)
	assert.Equal(t, float64(domain.LaughAndLieDownPot), out["pot"])
}

func TestLaughAndLieDownWebPresenter_ShipsTheThreeTakeOptionSeparately(t *testing.T) {
	// 「1 枚か 3 枚」の選択肢をクライアントが場札を数え直して再現しないで済むよう、
	// 3 枚取りできる添字だけを別に送る。
	g := lldStub([]int{1, 3}, true, false, 0)
	out := lldDecode(t, new(LaughAndLieDownWebPresenter).Output(g, nil))
	assert.Equal(t, []any{float64(1), float64(3)}, out["validIndices"])
	assert.Equal(t, []any{float64(1), float64(3)}, out["threeTakeIndices"])

	// 3 枚取りできないときは部分集合が空になる。ValidIndices と同一ではない。
	no := lldStub([]int{1, 3}, false, false, 0)
	assert.Empty(t, lldDecode(t, new(LaughAndLieDownWebPresenter).Output(no, nil))["threeTakeIndices"])
}

func TestLaughAndLieDownWebPresenter_CarriesTheHintOnAPlainStateResponse(t *testing.T) {
	// 人間の手番のときだけ載る。素の Reset は親 (席 0 = 人間) の左隣から始まる
	// ので、そのままでは手番が回っていない。
	g := lldStub([]int{0}, false, false, 0)
	assert.NotNil(t, lldDecode(t, new(LaughAndLieDownWebPresenter).Output(g, nil))["hint"],
		"the hint toggle reads the ordinary response")

	fresh := lldTestGame(t)
	require.NotEqual(t, 0, fresh.GetCurrentPlayerIdx(), "play starts to the dealer's left")
	assert.Nil(t, lldDecode(t, new(LaughAndLieDownWebPresenter).Output(fresh, nil))["hint"],
		"and nothing is suggested while it is somebody else's turn")
}

func TestLaughAndLieDownWebPresenter_HintReasonsCoverEveryBranch(t *testing.T) {
	for name, tc := range map[string]struct {
		gameEnd bool
		current int
		action  domain.LaughAndLieDownCpuAction
		want    string
		hasCard bool
	}{
		"game over":     {true, 0, domain.LaughAndLieDownCpuAction{}, "laughandliedown.hint.game_end", false},
		"not your turn": {false, 1, domain.LaughAndLieDownCpuAction{}, "laughandliedown.hint.not_your_turn", false},
		"must lie down": {false, 0, domain.LaughAndLieDownCpuAction{HandIdx: -1}, "laughandliedown.hint.must_lie_down", false},
		"take one":      {false, 0, domain.LaughAndLieDownCpuAction{HandIdx: 1, TakeCount: 1}, "laughandliedown.hint.take_one", true},
		"take three":    {false, 0, domain.LaughAndLieDownCpuAction{HandIdx: 1, TakeCount: 3}, "laughandliedown.hint.take_three", true},
	} {
		t.Run(name, func(t *testing.T) {
			g := new(interfaces.MockLaughAndLieDownGame)
			g.On("GetGameEndFlag").Return(tc.gameEnd)
			g.On("GetCurrentPlayerIdx").Return(tc.current)
			g.On("LaughAndLieDownCpuDecide", 0).Return(tc.action)

			hint := laughAndLieDownHint(g)
			assert.Equal(t, tc.want, hint.Reason)
			assert.Equal(t, tc.hasCard, hint.CardIndex != nil)
		})
	}
}

func TestLaughAndLieDownWebPresenter_MessageCodesFollowTheNetResult(t *testing.T) {
	// 勝敗は「勝ち抜け」ではなく収支で決まる。
	for score, want := range map[int]string{3: "laughandliedown.win", 0: "laughandliedown.even", -2: "laughandliedown.lose"} {
		g := lldStub(nil, false, true, score)
		assert.Equal(t, want, lldDecode(t, new(LaughAndLieDownWebPresenter).Output(g, nil))["messageCode"], "score %d", score)
	}

	running := lldStub(nil, false, false, 0)
	assert.Empty(t, lldDecode(t, new(LaughAndLieDownWebPresenter).Output(running, nil))["message"])
	assert.Equal(t, "boom", lldDecode(t, new(LaughAndLieDownWebPresenter).Output(running, errors.New("boom")))["message"])
}

func TestLaughAndLieDownWebPresenter_HintOutputAndActionLog(t *testing.T) {
	l := lldTestGame(t)
	assert.NotNil(t, lldDecode(t, new(LaughAndLieDownWebPresenter).HintOutput(l))["hint"])
	assert.NotEmpty(t, new(LaughAndLieDownWebPresenter).ActionLogOutput(l))
}

func TestLaughAndLieDownCuiPresenter_ShowsTheWholeTableAndOnlyYourHand(t *testing.T) {
	l := lldTestGame(t)
	out := new(LaughAndLieDownCuiPresenter).Output(l, nil)

	assert.Contains(t, out, "[0]", "your hand is indexed")
	assert.Contains(t, out, i18n.T("laughandliedown.ruleLine"))
	assert.Contains(t, out, i18n.T("laughandliedown.promptPlay"))
}

func TestLaughAndLieDownCuiPresenter_MarksWhoHasLaidDown(t *testing.T) {
	// 降りた人は場に手札を撒いた人なので、誰が降りたかは進行の読みに直結する。
	g := lldStub(nil, false, false, 0)
	assert.NotContains(t, new(LaughAndLieDownCuiPresenter).Output(g, nil), i18n.T("laughandliedown.laidDownMark"))

	down := new(interfaces.MockLaughAndLieDownGame)
	down.On("GetPhase").Return(domain.LaughAndLieDownPhasePlay)
	down.On("GetGameEndFlag").Return(false)
	down.On("GetCurrentPlayerIdx").Return(0)
	down.On("GetLayout").Return([]*domain.Card{})
	down.On("GetValidPlayIndices", mock.Anything).Return([]int(nil))
	down.On("GetWonCount", mock.Anything).Return(0)
	down.On("IsLaidDown", mock.Anything).Return(true)
	down.On("GetDealerIdx").Return(0)
	down.On("GetPlayers").Return([]*domain.LaughAndLieDownPlayer{domain.NewLaughAndLieDownPlayer(false)})
	assert.Contains(t, new(LaughAndLieDownCuiPresenter).Output(down, nil), i18n.T("laughandliedown.laidDownMark"))
}

func TestLaughAndLieDownCuiPresenter_PrintsTheSettlementAtTheEnd(t *testing.T) {
	l := lldTestGame(t)
	for range 300 {
		if l.GetGameEndFlag() {
			break
		}
		idx := l.GetCurrentPlayerIdx()
		action := l.LaughAndLieDownCpuDecide(idx)
		if action.HandIdx < 0 {
			break
		}
		require.NoError(t, l.PlayCard(idx, action.HandIdx, action.TakeCount))
	}
	require.True(t, l.GetGameEndFlag())

	out := new(LaughAndLieDownCuiPresenter).Output(l, nil)
	// 取り札の枚数と収支を全員ぶん出す。数字が出ていないと、なぜその収支に
	// なったのかが追えない。
	assert.Contains(t, out, i18n.Tf("laughandliedown.settleLine",
		"name", "あなた", "won", "0", "score", "0")[:6])
	assert.NotContains(t, out, i18n.T("laughandliedown.promptPlay"))
}

func TestLaughAndLieDownCuiPresenter_ErrorsAndHints(t *testing.T) {
	p := new(LaughAndLieDownCuiPresenter)
	running := lldStub(nil, false, false, 0)
	assert.Contains(t, p.Output(running, errors.New("boom")), "boom")

	play := new(interfaces.MockLaughAndLieDownGame)
	play.On("GetGameEndFlag").Return(false)
	play.On("GetCurrentPlayerIdx").Return(0)
	play.On("LaughAndLieDownCpuDecide", 0).Return(domain.LaughAndLieDownCpuAction{HandIdx: 2, TakeCount: 3})
	assert.Contains(t, p.HintOutput(play), "2")

	over := new(interfaces.MockLaughAndLieDownGame)
	over.On("GetGameEndFlag").Return(true)
	assert.NotEmpty(t, p.HintOutput(over))
}

func TestLaughAndLieDownCuiPresenter_HintReasonKeysAreAllMapped(t *testing.T) {
	// 未マッピングの reason は生キーを表示してしまう。
	for _, reason := range []string{
		"laughandliedown.hint.game_end", "laughandliedown.hint.not_your_turn",
		"laughandliedown.hint.must_lie_down", "laughandliedown.hint.take_one",
		"laughandliedown.hint.take_three",
	} {
		assert.NotEmpty(t, laughAndLieDownHintReasonKeys[reason], "unmapped reason %s", reason)
	}
}

func TestLaughAndLieDownCuiPresenter_ActionLog(t *testing.T) {
	assert.NotEmpty(t, new(LaughAndLieDownCuiPresenter).ActionLogOutput(lldTestGame(t)))
}
