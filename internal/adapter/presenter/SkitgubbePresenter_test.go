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

func sgTestGame(t *testing.T) *domain.Skitgubbe {
	t.Helper()
	s := domain.NewDefaultSkitgubbe()
	s.Reset()
	return s
}

func sgDecode(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

// sgStub wires a MockSkitgubbeGame with every accessor the presenters touch, so
// a test can pin an exact phase/pile rather than shuffling until one appears.
func sgStub(phase domain.SkitgubbePhase, pile []*domain.Card, valid []int, gameEnd bool, loser int) *interfaces.MockSkitgubbeGame {
	g := new(interfaces.MockSkitgubbeGame)
	g.On("GetPhase").Return(phase)
	g.On("GetGameEndFlag").Return(gameEnd)
	g.On("GetLoserIdx").Return(loser)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetStockCount").Return(0)
	g.On("GetTrumpSuit").Return(domain.CardDesignHeart)
	g.On("GetDuel").Return([]*domain.Card{})
	g.On("GetDuelLeader").Return(0)
	g.On("GetPile").Return(pile)
	g.On("GetValidPlayIndices", mock.Anything).Return(valid)
	g.On("GetCollectedCount", mock.Anything).Return(0)
	g.On("IsFinished", mock.Anything).Return(false)
	g.On("GetConfig").Return(domain.DefaultSkitgubbeConfig())
	g.On("GetPlayers").Return([]*domain.SkitgubbePlayer{
		domain.NewSkitgubbePlayer(true), domain.NewSkitgubbePlayer(false), domain.NewSkitgubbePlayer(false),
	})
	g.On("GetPlayer", mock.Anything).Return(domain.NewSkitgubbePlayer(false))
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	g.On("SkitgubbeCpuDecide", mock.Anything).Return(domain.SkitgubbeCpuAction{HandIdx: 0})
	return g
}

func TestSkitgubbeWebPresenter_HidesTheCpuHandButNotItsCount(t *testing.T) {
	// 手札は伏せる。枚数と「集めた枚数」は卓上で数えられるので公開する。
	out := sgDecode(t, new(SkitgubbeWebPresenter).Output(sgTestGame(t), nil))
	players, ok := out["players"].([]any)
	require.True(t, ok)
	require.Len(t, players, domain.SkitgubbePlayerCnt)

	human, _ := players[0].(map[string]any)
	assert.False(t, human["hidden"].(bool))
	assert.NotEmpty(t, human["cards"])

	cpu, _ := players[1].(map[string]any)
	assert.True(t, cpu["hidden"].(bool))
	assert.Empty(t, cpu["cards"], "the opponent's hand must not reach the browser")
	assert.Positive(t, cpu["cardCount"], "but its size is public")
}

func TestSkitgubbeWebPresenter_ShipsTheLegalMovesSoTheClientNeverRecomputesThem(t *testing.T) {
	// 「直前の札を上回る」判定はサーバーが一度だけ行う。
	g := sgStub(domain.SkitgubbePhaseShed, []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)}, []int{1, 3}, false, -1)
	out := sgDecode(t, new(SkitgubbeWebPresenter).Output(g, nil))

	assert.Equal(t, []any{float64(1), float64(3)}, out["validIndices"])
	assert.False(t, out["canPickUp"].(bool), "you may not duck while you can beat the pile")
}

func TestSkitgubbeWebPresenter_OffersPickUpOnlyWhenNothingBeatsThePile(t *testing.T) {
	g := sgStub(domain.SkitgubbePhaseShed, []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)}, nil, false, -1)
	assert.True(t, sgDecode(t, new(SkitgubbeWebPresenter).Output(g, nil))["canPickUp"].(bool))

	// 第1フェーズには引き取りが存在しない。
	collect := sgStub(domain.SkitgubbePhaseCollect, nil, nil, false, -1)
	assert.False(t, sgDecode(t, new(SkitgubbeWebPresenter).Output(collect, nil))["canPickUp"].(bool))

	// 場が空なら引き取るものがない。
	empty := sgStub(domain.SkitgubbePhaseShed, nil, nil, false, -1)
	assert.False(t, sgDecode(t, new(SkitgubbeWebPresenter).Output(empty, nil))["canPickUp"].(bool))
}

func TestSkitgubbeWebPresenter_CarriesTheHintOnAPlainStateResponse(t *testing.T) {
	// HintOutput にしか載せないと、フロントは通常の state を読むので何も出ない。
	out := sgDecode(t, new(SkitgubbeWebPresenter).Output(sgTestGame(t), nil))
	assert.NotNil(t, out["hint"], "the hint toggle reads the ordinary response")
}

func TestSkitgubbeWebPresenter_HintTellsYouToPickUpWhenYouMust(t *testing.T) {
	g := new(interfaces.MockSkitgubbeGame)
	g.On("GetGameEndFlag").Return(false)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("SkitgubbeCpuDecide", 0).Return(domain.SkitgubbeCpuAction{PickUp: true, HandIdx: -1})
	g.On("GetPhase").Return(domain.SkitgubbePhaseShed)

	hint := skitgubbeHint(g)
	assert.True(t, hint.PickUp)
	assert.Nil(t, hint.CardIndex)
	assert.Equal(t, "skitgubbe.hint.pickup", hint.Reason)
}

func TestSkitgubbeWebPresenter_HintReasonsCoverEveryBranch(t *testing.T) {
	for name, tc := range map[string]struct {
		gameEnd bool
		current int
		action  domain.SkitgubbeCpuAction
		phase   domain.SkitgubbePhase
		want    string
	}{
		"game over":     {true, 0, domain.SkitgubbeCpuAction{}, domain.SkitgubbePhaseShed, "skitgubbe.hint.game_end"},
		"not your turn": {false, 1, domain.SkitgubbeCpuAction{}, domain.SkitgubbePhaseShed, "skitgubbe.hint.not_your_turn"},
		"no card":       {false, 0, domain.SkitgubbeCpuAction{HandIdx: -1}, domain.SkitgubbePhaseShed, "skitgubbe.hint.none"},
		"duel":          {false, 0, domain.SkitgubbeCpuAction{HandIdx: 1}, domain.SkitgubbePhaseCollect, "skitgubbe.hint.duel"},
		"beat":          {false, 0, domain.SkitgubbeCpuAction{HandIdx: 1}, domain.SkitgubbePhaseShed, "skitgubbe.hint.beat"},
	} {
		t.Run(name, func(t *testing.T) {
			g := new(interfaces.MockSkitgubbeGame)
			g.On("GetGameEndFlag").Return(tc.gameEnd)
			g.On("GetCurrentPlayerIdx").Return(tc.current)
			g.On("SkitgubbeCpuDecide", 0).Return(tc.action)
			g.On("GetPhase").Return(tc.phase)

			assert.Equal(t, tc.want, skitgubbeHint(g).Reason)
		})
	}
}

func TestSkitgubbeWebPresenter_MessageCodes(t *testing.T) {
	lose := sgStub(domain.SkitgubbePhaseShed, nil, nil, true, 0)
	assert.Equal(t, "skitgubbe.lose", sgDecode(t, new(SkitgubbeWebPresenter).Output(lose, nil))["messageCode"])

	win := sgStub(domain.SkitgubbePhaseShed, nil, nil, true, 2)
	assert.Equal(t, "skitgubbe.win", sgDecode(t, new(SkitgubbeWebPresenter).Output(win, nil))["messageCode"])

	// 進行中はメッセージなし。エラーはそのまま載せる。
	running := sgStub(domain.SkitgubbePhaseShed, nil, nil, false, -1)
	assert.Empty(t, sgDecode(t, new(SkitgubbeWebPresenter).Output(running, nil))["message"])
	assert.Equal(t, "boom", sgDecode(t, new(SkitgubbeWebPresenter).Output(running, errors.New("boom")))["message"])
}

func TestSkitgubbeWebPresenter_HintOutputAndActionLog(t *testing.T) {
	s := sgTestGame(t)
	assert.NotNil(t, sgDecode(t, new(SkitgubbeWebPresenter).HintOutput(s))["hint"])
	assert.NotEmpty(t, new(SkitgubbeWebPresenter).ActionLogOutput(s))
}

func TestSkitgubbeCuiPresenter_ShowsYourHandAndNotTheOpponents(t *testing.T) {
	s := sgTestGame(t)
	out := new(SkitgubbeCuiPresenter).Output(s, nil)

	// 自分の手札は添字つきで並ぶ。相手の手札は枚数だけ。
	assert.Contains(t, out, "[0]")
	assert.Equal(t, 1, strings.Count(out, "[0]"), "only the human hand is indexed")
}

func TestSkitgubbeCuiPresenter_ShowsTheTrumpOnlyOnceItIsDecided(t *testing.T) {
	// 切札は山札から最後に引かれた札で決まるので、序盤は必ず -1 で描画される。
	// そこで実スートを出すと、まだ決まっていない切札を教えることになる。
	assert.Equal(t, i18n.T("skitgubbe.trumpUndecided"), skitgubbeTrumpStr(-1))
	assert.Equal(t, cuiSuitName(domain.CardDesignHeart), skitgubbeTrumpStr(domain.CardDesignHeart))
	assert.Contains(t, new(SkitgubbeCuiPresenter).Output(sgTestGame(t), nil), i18n.T("skitgubbe.trumpUndecided"))
}

func TestSkitgubbeCuiPresenter_PromptsPickUpOnlyWhenItIsTheOnlyMove(t *testing.T) {
	pile := []*domain.Card{domain.NewCard(domain.CardDesignSpade, 5, true)}

	stuck := sgStub(domain.SkitgubbePhaseShed, pile, nil, false, -1)
	assert.Contains(t, new(SkitgubbeCuiPresenter).Output(stuck, nil), i18n.T("skitgubbe.promptPickUp"))

	// 出せる札があるうちは引き取れない ("It is never lawful to duck")。
	canPlay := sgStub(domain.SkitgubbePhaseShed, pile, []int{0}, false, -1)
	out := new(SkitgubbeCuiPresenter).Output(canPlay, nil)
	assert.NotContains(t, out, i18n.T("skitgubbe.promptPickUp"))
	assert.Contains(t, out, i18n.T("skitgubbe.promptPlay"))
}

func TestSkitgubbeCuiPresenter_BannersAndErrors(t *testing.T) {
	lose := sgStub(domain.SkitgubbePhaseShed, nil, nil, true, 0)
	loseOut := new(SkitgubbeCuiPresenter).Output(lose, nil)
	win := sgStub(domain.SkitgubbePhaseShed, nil, nil, true, 2)
	winOut := new(SkitgubbeCuiPresenter).Output(win, nil)
	assert.NotEqual(t, loseOut, winOut, "the two endings must not read the same")

	running := sgStub(domain.SkitgubbePhaseCollect, nil, nil, false, -1)
	assert.Contains(t, new(SkitgubbeCuiPresenter).Output(running, errors.New("boom")), "boom")
}

func TestSkitgubbeCuiPresenter_HintRendersEveryShape(t *testing.T) {
	p := new(SkitgubbeCuiPresenter)

	pick := new(interfaces.MockSkitgubbeGame)
	pick.On("GetGameEndFlag").Return(false)
	pick.On("GetCurrentPlayerIdx").Return(0)
	pick.On("GetPhase").Return(domain.SkitgubbePhaseShed)
	pick.On("SkitgubbeCpuDecide", 0).Return(domain.SkitgubbeCpuAction{PickUp: true, HandIdx: -1})
	assert.NotEmpty(t, p.HintOutput(pick))

	play := new(interfaces.MockSkitgubbeGame)
	play.On("GetGameEndFlag").Return(false)
	play.On("GetCurrentPlayerIdx").Return(0)
	play.On("GetPhase").Return(domain.SkitgubbePhaseCollect)
	play.On("SkitgubbeCpuDecide", 0).Return(domain.SkitgubbeCpuAction{HandIdx: 2})
	assert.Contains(t, p.HintOutput(play), "2")

	over := new(interfaces.MockSkitgubbeGame)
	over.On("GetGameEndFlag").Return(true)
	assert.NotEmpty(t, p.HintOutput(over))
}

func TestSkitgubbeCuiPresenter_HintReasonKeysAreAllMapped(t *testing.T) {
	// 未マッピングの reason は生キーを表示してしまう。Web が送る識別子と
	// CUI の表引きは同じ集合でなければならない。
	for _, reason := range []string{
		"skitgubbe.hint.game_end", "skitgubbe.hint.not_your_turn",
		"skitgubbe.hint.duel", "skitgubbe.hint.beat",
		"skitgubbe.hint.pickup", "skitgubbe.hint.none",
	} {
		assert.NotEmpty(t, skitgubbeHintReasonKeys[reason], "unmapped reason %s", reason)
	}
}

func TestSkitgubbeCuiPresenter_ActionLog(t *testing.T) {
	assert.NotEmpty(t, new(SkitgubbeCuiPresenter).ActionLogOutput(sgTestGame(t)))
}
