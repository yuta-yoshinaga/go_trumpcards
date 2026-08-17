//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func sjTestGame(t *testing.T) *domain.Sjavs {
	t.Helper()
	s := domain.NewDefaultSjavs()
	s.Reset()
	return s
}

func sjDecode(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

// sjStub wires a MockSjavsGame with every accessor the presenters touch, so a
// test can pin an exact state rather than shuffling until one appears.
func sjStub(phase domain.SjavsPhase, trump int, gameEnd bool, winner int) *interfaces.MockSjavsGame {
	g := new(interfaces.MockSjavsGame)
	g.On("GetPhase").Return(phase)
	g.On("GetGameEndFlag").Return(gameEnd)
	g.On("GetWinnerTeam").Return(winner)
	g.On("IsDoubleVictory").Return(false)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetDealerIdx").Return(0)
	g.On("GetTrumpSuit").Return(trump)
	g.On("GetBidderIdx").Return(0)
	g.On("GetBidLength").Return(6)
	g.On("GetBids").Return([]int{6, 0, 0, 0})
	g.On("LongestTrumpLength", mock.Anything).Return(6)
	g.On("GetTrick").Return([]domain.SjavsTrickCard{})
	g.On("GetTrickNumber").Return(0)
	g.On("GetValidPlayIndices", mock.Anything).Return([]int{0})
	g.On("GetTeamPoints", mock.Anything).Return(60)
	g.On("GetRemaining", mock.Anything).Return(domain.SjavsRubber)
	g.On("GetCrosses", mock.Anything).Return(0)
	g.On("GetCarryOver").Return(0)
	g.On("GetHandResult").Return((*domain.SjavsHandResult)(nil))
	g.On("GetConfig").Return(domain.DefaultSjavsConfig())
	players := make([]*domain.SjavsPlayer, 0, domain.SjavsPlayerCnt)
	players = append(players, domain.NewSjavsPlayer(true))
	for range domain.SjavsPlayerCnt - 1 {
		players = append(players, domain.NewSjavsPlayer(false))
	}
	g.On("GetPlayers").Return(players)
	g.On("GetPlayer", mock.Anything).Return(domain.NewSjavsPlayer(false))
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	g.On("SjavsCpuDecide", mock.Anything).Return(domain.SjavsCpuAction{BidLength: 6, HandIdx: 0})
	return g
}

func TestSjavsWebPresenter_HidesTheCpuHandButNeverTheBids(t *testing.T) {
	// 申告は卓上で聞こえる宣言。味方が何枚持っているかはパートナーシップの読み。
	out := sjDecode(t, new(SjavsWebPresenter).Output(sjTestGame(t), nil))
	players, ok := out["players"].([]any)
	require.True(t, ok)
	require.Len(t, players, domain.SjavsPlayerCnt)

	human, _ := players[0].(map[string]any)
	assert.False(t, human["hidden"].(bool))
	assert.NotEmpty(t, human["cards"])

	cpu, _ := players[1].(map[string]any)
	assert.True(t, cpu["hidden"].(bool))
	assert.Empty(t, cpu["cards"], "the opponent's hand must not reach the browser")
	assert.Positive(t, cpu["cardCount"], "but its size is public")
	assert.NotNil(t, cpu["bid"], "and so is its bid")
	assert.Equal(t, float64(1), cpu["team"], "seat 1 is the other team")
}

func TestSjavsWebPresenter_ShipsTheTrumpCountSoTheClientNeverRecountsIt(t *testing.T) {
	// 常時切札 6 枚を含むので、クライアントが切札スートだけ数えると必ず足りない。
	red := sjStub(domain.SjavsPhasePlay, domain.CardDesignHeart, false, -1)
	assert.Equal(t, float64(13), sjDecode(t, new(SjavsWebPresenter).Output(red, nil))["trumpCount"])

	black := sjStub(domain.SjavsPhasePlay, domain.CardDesignSpade, false, -1)
	assert.Equal(t, float64(12), sjDecode(t, new(SjavsWebPresenter).Output(black, nil))["trumpCount"])

	// 未確定のうちは 0 のまま。適当な数字を出すと切札を推測させてしまう。
	undecided := sjStub(domain.SjavsPhaseBid, -1, false, -1)
	out := sjDecode(t, new(SjavsWebPresenter).Output(undecided, nil))
	assert.Equal(t, float64(-1), out["trumpSuit"])
	assert.Equal(t, float64(0), out["trumpCount"])
}

func TestSjavsWebPresenter_OffersLegalPlaysOnlyDuringThePlayPhase(t *testing.T) {
	// ビッド中に出せる札を送ると、フロントがカードを押せる状態に見える。
	bid := sjStub(domain.SjavsPhaseBid, -1, false, -1)
	assert.Empty(t, sjDecode(t, new(SjavsWebPresenter).Output(bid, nil))["validIndices"])

	play := sjStub(domain.SjavsPhasePlay, domain.CardDesignHeart, false, -1)
	assert.Equal(t, []any{float64(0)}, sjDecode(t, new(SjavsWebPresenter).Output(play, nil))["validIndices"])
}

func TestSjavsWebPresenter_CarriesTheHintOnAPlainStateResponse(t *testing.T) {
	g := sjStub(domain.SjavsPhasePlay, domain.CardDesignHeart, false, -1)
	assert.NotNil(t, sjDecode(t, new(SjavsWebPresenter).Output(g, nil))["hint"],
		"the hint toggle reads the ordinary response")
}

func TestSjavsWebPresenter_HintReasonsCoverEveryBranch(t *testing.T) {
	for name, tc := range map[string]struct {
		gameEnd bool
		current int
		phase   domain.SjavsPhase
		action  domain.SjavsCpuAction
		want    string
	}{
		"game over":     {true, 0, domain.SjavsPhasePlay, domain.SjavsCpuAction{}, "sjavs.hint.game_end"},
		"not your turn": {false, 1, domain.SjavsPhasePlay, domain.SjavsCpuAction{}, "sjavs.hint.not_your_turn"},
		"bid":           {false, 0, domain.SjavsPhaseBid, domain.SjavsCpuAction{BidLength: 6}, "sjavs.hint.bid"},
		"pass":          {false, 0, domain.SjavsPhaseBid, domain.SjavsCpuAction{BidLength: 0}, "sjavs.hint.pass"},
		"play":          {false, 0, domain.SjavsPhasePlay, domain.SjavsCpuAction{HandIdx: 2}, "sjavs.hint.play"},
		"no card":       {false, 0, domain.SjavsPhasePlay, domain.SjavsCpuAction{HandIdx: -1}, "sjavs.hint.none"},
	} {
		t.Run(name, func(t *testing.T) {
			g := new(interfaces.MockSjavsGame)
			g.On("GetGameEndFlag").Return(tc.gameEnd)
			g.On("GetCurrentPlayerIdx").Return(tc.current)
			g.On("GetPhase").Return(tc.phase)
			g.On("SjavsCpuDecide", 0).Return(tc.action)

			assert.Equal(t, tc.want, sjavsHint(g).Reason)
		})
	}
}

func TestSjavsWebPresenter_MessageCodesTellTheDoubleVictoryApart(t *testing.T) {
	win := sjStub(domain.SjavsPhaseGameEnd, domain.CardDesignHeart, true, 0)
	assert.Equal(t, "sjavs.win", sjDecode(t, new(SjavsWebPresenter).Output(win, nil))["messageCode"])

	lose := sjStub(domain.SjavsPhaseGameEnd, domain.CardDesignHeart, true, 1)
	assert.Equal(t, "sjavs.lose", sjDecode(t, new(SjavsWebPresenter).Output(lose, nil))["messageCode"])

	// ダブル勝ちは別のコード。同じ「勝ち」に潰すと、相手が 24 のままだった
	// ことが伝わらない。
	dbl := new(interfaces.MockSjavsGame)
	for _, m := range []string{"GetGameEndFlag"} {
		dbl.On(m).Return(true)
	}
	dbl.On("GetWinnerTeam").Return(0)
	dbl.On("IsDoubleVictory").Return(true)
	assert.Equal(t, "sjavs.winDouble", presenterMessageCode(t, dbl))

	dblLose := new(interfaces.MockSjavsGame)
	dblLose.On("GetGameEndFlag").Return(true)
	dblLose.On("GetWinnerTeam").Return(1)
	dblLose.On("IsDoubleVictory").Return(true)
	assert.Equal(t, "sjavs.loseDouble", presenterMessageCode(t, dblLose))

	running := sjStub(domain.SjavsPhasePlay, domain.CardDesignHeart, false, -1)
	assert.Empty(t, sjDecode(t, new(SjavsWebPresenter).Output(running, nil))["message"])
	assert.Equal(t, "boom", sjDecode(t, new(SjavsWebPresenter).Output(running, errors.New("boom")))["message"])
}

// presenterMessageCode calls buildMessage directly so the double-victory branch
// can be pinned without stubbing the whole state.
func presenterMessageCode(t *testing.T, g interfaces.SjavsGame) string {
	t.Helper()
	_, code, _ := new(SjavsWebPresenter).buildMessage(g, nil)
	return code
}

func TestSjavsWebPresenter_HintOutputAndActionLog(t *testing.T) {
	s := sjTestGame(t)
	assert.NotNil(t, sjDecode(t, new(SjavsWebPresenter).HintOutput(s))["hint"])
	assert.NotEmpty(t, new(SjavsWebPresenter).ActionLogOutput(s))
}

func TestSjavsCuiPresenter_ShowsThePermanentTrumpsEveryFrame(t *testing.T) {
	// 切札スートの札しか切札でないと思い込むと、♣Q が飛んでくる理由が分からない。
	out := new(SjavsCuiPresenter).Output(sjTestGame(t), nil)
	assert.Contains(t, out, i18n.T("sjavs.ruleLine"))
	assert.Contains(t, out, i18n.T("sjavs.trumpUndecided"), "trump is undecided while bidding")
	assert.Contains(t, out, "[0]", "your hand is indexed")
}

func TestSjavsCuiPresenter_PromptsPerPhase(t *testing.T) {
	p := new(SjavsCuiPresenter)

	// ビッド中に「札を出せ」と促すと、押せないカードを押しに行かせる。
	bid := sjStub(domain.SjavsPhaseBid, -1, false, -1)
	bidOut := p.Output(bid, nil)
	assert.Contains(t, bidOut, i18n.Tf("sjavs.promptBid", "min", "5", "longest", "6"))
	assert.NotContains(t, bidOut, i18n.T("sjavs.promptPlay"))
	assert.NotContains(t, bidOut, i18n.Tf("sjavs.trumpCountLine", "n", "13"), "trump is undecided")

	play := sjStub(domain.SjavsPhasePlay, domain.CardDesignHeart, false, -1)
	playOut := p.Output(play, nil)
	assert.Contains(t, playOut, i18n.T("sjavs.promptPlay"))
	assert.NotContains(t, playOut, i18n.Tf("sjavs.promptBid", "min", "5", "longest", "6"))
	assert.Contains(t, playOut, i18n.Tf("sjavs.trumpCountLine", "n", "13"))
}

func TestSjavsCuiPresenter_ShowsTheHandSettlementAndTheTie(t *testing.T) {
	scored := sjStub(domain.SjavsPhaseHandEnd, domain.CardDesignHeart, false, -1)
	assert.NotEmpty(t, new(SjavsCuiPresenter).Output(scored, nil))

	// 60-60 は「加点なし」と明示する。何も出ないと入力を取りこぼしたように見える。
	tie := new(interfaces.MockSjavsGame)
	tie.On("GetGameEndFlag").Return(false)
	tie.On("GetPhase").Return(domain.SjavsPhaseHandEnd)
	tie.On("GetHandResult").Return(&domain.SjavsHandResult{ScoringTeam: -1})
	tie.On("GetCurrentPlayerIdx").Return(0)
	tie.On("LongestTrumpLength", mock.Anything).Return(6)
	assert.Contains(t, new(SjavsCuiPresenter).promptBlock(tie), i18n.T("sjavs.handTie"))
}

func TestSjavsCuiPresenter_BannersAndErrors(t *testing.T) {
	p := new(SjavsCuiPresenter)
	win := sjStub(domain.SjavsPhaseGameEnd, domain.CardDesignHeart, true, 0)
	lose := sjStub(domain.SjavsPhaseGameEnd, domain.CardDesignHeart, true, 1)
	assert.NotEqual(t, p.Output(win, nil), p.Output(lose, nil), "the two endings must not read the same")

	running := sjStub(domain.SjavsPhasePlay, domain.CardDesignHeart, false, -1)
	assert.Contains(t, p.Output(running, errors.New("boom")), "boom")
}

func TestSjavsCuiPresenter_HintRendersEveryShape(t *testing.T) {
	p := new(SjavsCuiPresenter)

	bid := new(interfaces.MockSjavsGame)
	bid.On("GetGameEndFlag").Return(false)
	bid.On("GetCurrentPlayerIdx").Return(0)
	bid.On("GetPhase").Return(domain.SjavsPhaseBid)
	bid.On("SjavsCpuDecide", 0).Return(domain.SjavsCpuAction{BidLength: 6})
	assert.Contains(t, p.HintOutput(bid), "6")

	play := new(interfaces.MockSjavsGame)
	play.On("GetGameEndFlag").Return(false)
	play.On("GetCurrentPlayerIdx").Return(0)
	play.On("GetPhase").Return(domain.SjavsPhasePlay)
	play.On("SjavsCpuDecide", 0).Return(domain.SjavsCpuAction{HandIdx: 2})
	assert.Contains(t, p.HintOutput(play), "2")

	over := new(interfaces.MockSjavsGame)
	over.On("GetGameEndFlag").Return(true)
	assert.NotEmpty(t, p.HintOutput(over))
}

func TestSjavsCuiPresenter_HintReasonKeysAreAllMapped(t *testing.T) {
	// 未マッピングの reason は生キーを表示してしまう。
	for _, reason := range []string{
		"sjavs.hint.game_end", "sjavs.hint.not_your_turn",
		"sjavs.hint.bid", "sjavs.hint.pass", "sjavs.hint.play", "sjavs.hint.none",
	} {
		assert.NotEmpty(t, sjavsHintReasonKeys[reason], "unmapped reason %s", reason)
	}
}

func TestSjavsCuiPresenter_ActionLog(t *testing.T) {
	assert.NotEmpty(t, new(SjavsCuiPresenter).ActionLogOutput(sjTestGame(t)))
}

// #5575: **常時切札の 6 枚 (♣Q ♠Q ♣J ♠J ♥J ♦J) はスートを見ても分からない。**
// 規則文は出ているのに、手札のどれがそれかは暗記に頼らせていた。
func TestSjavsWebPresenter_ShipsWhichCardsAreTrumps(t *testing.T) {
	human := domain.NewSjavsPlayer(true)
	// ♣Q は常時切札。♥7 は切札が ♥ のときだけ切札。♠8 はどちらでもない。
	human.AddCard(domain.NewCard(domain.CardDesignClover, 12, true))
	human.AddCard(domain.NewCard(domain.CardDesignHeart, 7, true))
	human.AddCard(domain.NewCard(domain.CardDesignSpade, 8, true))

	g := sjStub(domain.SjavsPhasePlay, domain.CardDesignHeart, false, -1)
	g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPlayer")
	g.On("GetPlayer", mock.Anything).Return(human)

	out := sjDecode(t, new(SjavsWebPresenter).Output(g, nil))
	assert.Equal(t, []any{float64(0), float64(1)}, out["trumpIndices"])

	// **スートが変わると答えも変わる。**常時切札だけは残ること。
	black := sjStub(domain.SjavsPhasePlay, domain.CardDesignSpade, false, -1)
	black.ExpectedCalls = filterCalls(black.ExpectedCalls, "GetPlayer")
	black.On("GetPlayer", mock.Anything).Return(human)
	blackOut := sjDecode(t, new(SjavsWebPresenter).Output(black, nil))
	assert.Equal(t, []any{float64(0), float64(2)}, blackOut["trumpIndices"])

	// 切札未確定のうちは空。埋めると、まだ決まっていない切札を推測させる。
	undecided := sjStub(domain.SjavsPhaseBid, -1, false, -1)
	undecided.ExpectedCalls = filterCalls(undecided.ExpectedCalls, "GetPlayer")
	undecided.On("GetPlayer", mock.Anything).Return(human)
	assert.Empty(t, sjDecode(t, new(SjavsWebPresenter).Output(undecided, nil))["trumpIndices"])
}

// CUI も同じ判定を通ること。片方だけ印を付けると、同じ手札が画面ごとに違って見える。
func TestSjavsCuiPresenter_MarksTheTrumpsInHand(t *testing.T) {
	i18n.SetLang("ja")
	human := domain.NewSjavsPlayer(true)
	human.AddCard(domain.NewCard(domain.CardDesignClover, 12, true)) // 常時切札
	human.AddCard(domain.NewCard(domain.CardDesignSpade, 8, true))   // 平札

	g := sjStub(domain.SjavsPhasePlay, domain.CardDesignHeart, false, -1)
	g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPlayers")
	g.On("GetPlayers").Return([]*domain.SjavsPlayer{
		human, domain.NewSjavsPlayer(false), domain.NewSjavsPlayer(false), domain.NewSjavsPlayer(false),
	})

	out := new(SjavsCuiPresenter).Output(g, nil)
	mark := i18n.T("sjavs.trumpMark")
	assert.Contains(t, out, cuiCardStr(human.GetCard(0))+color.Yellow(mark))
	// **平札には付けないこと。**全部に付ける実装でも「含む」検査だけなら通る。
	assert.NotContains(t, out, cuiCardStr(human.GetCard(1))+color.Yellow(mark))
}

// ビッド前は印を出さない。切札が決まっていないので、どれが切札かは決まっていない。
func TestSjavsCuiPresenter_MarksNothingBeforeTheTrumpIsNamed(t *testing.T) {
	i18n.SetLang("ja")
	human := domain.NewSjavsPlayer(true)
	human.AddCard(domain.NewCard(domain.CardDesignClover, 12, true))

	g := sjStub(domain.SjavsPhaseBid, -1, false, -1)
	g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetPlayers")
	g.On("GetPlayers").Return([]*domain.SjavsPlayer{
		human, domain.NewSjavsPlayer(false), domain.NewSjavsPlayer(false), domain.NewSjavsPlayer(false),
	})

	assert.NotContains(t, new(SjavsCuiPresenter).Output(g, nil), color.Yellow(i18n.T("sjavs.trumpMark")))
}
