//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHokm(t *testing.T) *Hokm {
	t.Helper()
	h := NewDefaultHokm()
	h.Reset()
	return h
}

// hokmHandOf は指定プレイヤーの手札を固定の並びに差し替える。
func hokmHandOf(h *Hokm, idx int, cards ...*Card) {
	p := h.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// settleTrump は切り札の宣言を済ませる。
func settleTrump(t *testing.T, h *Hokm) {
	t.Helper()
	if h.IsHumanTrumpTurn() {
		require.NoError(t, h.PlayerDeclareTrump(CardDesignSpade))
		return
	}
	h.CpuDeclareTrump()
	require.Equal(t, HokmPhasePlay, h.GetPhase())
}

// --- 配り ---

// **親だけに 5 枚。** 宣言に使えるのはこの 5 枚ぶんの情報だけ。
func TestHokm_OnlyTheHakemSeesFiveCardsFirst(t *testing.T) {
	h := newTestHokm(t)

	assert.Equal(t, HokmPhaseTrump, h.GetPhase())
	assert.Equal(t, HokmPeekSize, h.GetPlayer(h.GetHakemIdx()).GetCardsSize())
	for i := range HokmPlayerCnt {
		if i == h.GetHakemIdx() {
			continue
		}
		assert.Equal(t, 0, h.GetPlayer(i).GetCardsSize(), "player %d must not see cards yet", i)
	}
}

// 宣言すると残りが配られ、全員 13 枚になる。
func TestHokm_DealsThirteenEachAfterDeclaring(t *testing.T) {
	h := newTestHokm(t)
	settleTrump(t, h)

	total := 0
	for i := range HokmPlayerCnt {
		assert.Equal(t, HokmHandSize, h.GetPlayer(i).GetCardsSize(), "player %d", i)
		total += h.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 52, total)
	assert.Equal(t, HokmPhasePlay, h.GetPhase())
	assert.GreaterOrEqual(t, h.GetTrumpSuit(), CardDesignSpade)
}

// --- 切り札の宣言 ---

func TestHokm_OnlyTheHakemDeclares(t *testing.T) {
	h := newTestHokm(t)
	h.SetHakemIdxForTest(2)
	assert.False(t, h.IsHumanTrumpTurn())
	assert.Error(t, h.PlayerDeclareTrump(CardDesignSpade))

	h.SetHakemIdxForTest(0)
	assert.True(t, h.IsHumanTrumpTurn())
	assert.NoError(t, h.PlayerDeclareTrump(CardDesignSpade))
}

func TestHokm_DeclareRejectsInvalidSuit(t *testing.T) {
	h := newTestHokm(t)
	h.SetHakemIdxForTest(0)
	assert.Error(t, h.PlayerDeclareTrump(0))
	assert.Error(t, h.PlayerDeclareTrump(99))
	assert.Equal(t, HokmPhaseTrump, h.GetPhase())
}

func TestHokm_DeclareGuards(t *testing.T) {
	h := newTestHokm(t)
	h.SetHakemIdxForTest(0)
	h.SetPhaseForTest(HokmPhasePlay)
	assert.Error(t, h.PlayerDeclareTrump(CardDesignSpade))

	h.SetPhaseForTest(HokmPhaseTrump)
	h.GiveUp()
	assert.Error(t, h.PlayerDeclareTrump(CardDesignSpade))
}

func TestHokm_CpuDeclareIgnoresHumanHakem(t *testing.T) {
	h := newTestHokm(t)
	h.SetHakemIdxForTest(0)
	h.CpuDeclareTrump()
	assert.Equal(t, HokmPhaseTrump, h.GetPhase())
	assert.Equal(t, 0, h.GetTrumpSuit())
}

// --- 7 トリック早取り ---

// **7 トリック取った時点でハンドは終わる。** 残りの札は打たれない。
func TestHokm_HandEndsAtSevenTricks(t *testing.T) {
	h := newTestHokm(t)
	h.SetHakemIdxForTest(0)
	require.NoError(t, h.PlayerDeclareTrump(CardDesignSpade))
	// チーム0 に 6 トリック、チーム1 に 3 トリック持たせる。
	h.GiveTricksForTest(0, 6)
	h.GiveTricksForTest(1, 3)
	h.SetTrickNumberForTest(9)
	require.Equal(t, HokmPhasePlay, h.GetPhase())

	// 7 つ目を取らせる。
	h.SetCurrentPlayerIdxForTest(0)
	h.SetLeadPlayerIdxForTest(0)
	hokmHandOf(h, 0, NewCard(CardDesignSpade, 1, false))
	hokmHandOf(h, 1, NewCard(CardDesignSpade, 2, false))
	hokmHandOf(h, 2, NewCard(CardDesignSpade, 3, false))
	hokmHandOf(h, 3, NewCard(CardDesignSpade, 4, false))
	for i := range HokmPlayerCnt {
		require.NoError(t, h.PlayForTest(i, 0))
	}

	assert.Equal(t, HokmTricksToWin, h.TeamTricks(0))
	assert.Equal(t, HokmPhaseHandEnd, h.GetPhase(), "7 取った時点で終わる")
	assert.Equal(t, 10, h.GetTrickNumber(), "13 トリックは消化しない")
	assert.Equal(t, 0, h.GetLastHandWinner())
}

// **13 トリックまで行くことはない。** 7+7 = 14 > 13 なので必ず先に決まる。
func TestHokm_AHandAlwaysEndsBeforeThirteenTricks(t *testing.T) {
	for range 20 {
		h := newTestHokm(t)
		settleTrump(t, h)
		for h.GetPhase() == HokmPhasePlay {
			if h.IsHumanTurn() {
				valid := h.GetValidPlayIndices(0)
				require.NotEmpty(t, valid)
				require.NoError(t, h.PlayerPlay(valid[0]))
				continue
			}
			h.CpuPlay()
		}
		assert.LessOrEqual(t, h.GetTrickNumber(), HokmHandSize)
		assert.GreaterOrEqual(t, h.TeamTricks(h.GetLastHandWinner()), HokmTricksToWin)
		// **早取りなので、手札が残ったまま終わることがある。**
		assert.LessOrEqual(t, h.GetTrickNumber(), 13)
	}
}

// --- Kot と勝ち点 ---

// **相手が 1 トリックも取れていなければ Kot で 2 点。** 通常は 1 点。
func TestHokm_KotDoublesTheHandPoint(t *testing.T) {
	t.Run("kot", func(t *testing.T) {
		h := newTestHokm(t)
		h.GiveTricksForTest(0, HokmTricksToWin)
		h.FinishHandForTest(0)

		assert.Equal(t, HokmKotPoints, h.GetScore(0))
		assert.True(t, h.GetLastHandKot())
	})

	t.Run("not a kot", func(t *testing.T) {
		h := newTestHokm(t)
		h.GiveTricksForTest(0, HokmTricksToWin)
		h.GiveTricksForTest(1, 1)
		h.FinishHandForTest(0)

		assert.Equal(t, HokmHandPoints, h.GetScore(0))
		assert.False(t, h.GetLastHandKot())
	})
}

// **親は勝っているあいだ交代しない。** 負けたときだけ左隣へ移る。
func TestHokm_HakemStaysWhileWinning(t *testing.T) {
	t.Run("the hakem's team wins", func(t *testing.T) {
		h := newTestHokm(t)
		h.SetHakemIdxForTest(1)
		h.GiveTricksForTest(1, HokmTricksToWin)
		h.GiveTricksForTest(0, 1)
		h.FinishHandForTest(HokmTeamOf(1))

		assert.Equal(t, 1, h.GetHakemIdx(), "勝った親は続投")
	})

	t.Run("the hakem's team loses", func(t *testing.T) {
		h := newTestHokm(t)
		h.SetHakemIdxForTest(1)
		h.GiveTricksForTest(0, HokmTricksToWin)
		h.GiveTricksForTest(1, 1)
		h.FinishHandForTest(HokmTeamOf(0))

		assert.Equal(t, 2, h.GetHakemIdx(), "負けた親は左隣へ移る")
	})
}

// 向かい合う席が味方。
func TestHokm_TeamAssignment(t *testing.T) {
	assert.Equal(t, HokmTeamOf(0), HokmTeamOf(2))
	assert.Equal(t, HokmTeamOf(1), HokmTeamOf(3))
	assert.NotEqual(t, HokmTeamOf(0), HokmTeamOf(1))
}

// チームのトリック数は 2 席ぶんの合計。
func TestHokm_TeamTricks(t *testing.T) {
	h := newTestHokm(t)
	h.GiveTricksForTest(0, 2)
	h.GiveTricksForTest(2, 3)
	h.GiveTricksForTest(1, 1)

	assert.Equal(t, 5, h.TeamTricks(0))
	assert.Equal(t, 1, h.TeamTricks(1))
	assert.Equal(t, 0, h.TeamTricks(-1))
	assert.Equal(t, 0, h.TeamTricks(HokmTeamCnt))
}

// 規定ハンドに届かなければ次のハンドへ、届けば終局。
func TestHokm_NextHandAndGameEnd(t *testing.T) {
	h := newTestHokm(t)
	h.SetConfig(HokmConfig{Target: HokmTargetMax})
	h.GiveTricksForTest(0, HokmTricksToWin)
	h.GiveTricksForTest(1, 1)
	h.FinishHandForTest(0)
	require.Equal(t, HokmPhaseHandEnd, h.GetPhase())

	h.NextHand()
	assert.Equal(t, 2, h.GetHandNumber())
	assert.Equal(t, HokmPhaseTrump, h.GetPhase())
	assert.Equal(t, 0, h.GetTrumpSuit(), "切り札はハンドごとに宣言し直す")
	assert.Equal(t, HokmPeekSize, h.GetPlayer(h.GetHakemIdx()).GetCardsSize())
	for i := range HokmPlayerCnt {
		assert.Equal(t, 0, h.GetPlayer(i).GetTrickCount(), "獲得トリックも白紙")
	}
}

func TestHokm_GameEndsAtTarget(t *testing.T) {
	h := newTestHokm(t)
	h.SetConfig(HokmConfig{Target: HokmTargetMin})
	h.GiveTricksForTest(0, HokmTricksToWin)
	h.GiveTricksForTest(1, 1)
	h.FinishHandForTest(0)

	assert.True(t, h.GetGameEndFlag())
	assert.Equal(t, HokmPhaseGameEnd, h.GetPhase())
	assert.Equal(t, 0, h.GetWinnerTeam())

	before := h.GetHandNumber()
	h.NextHand()
	assert.Equal(t, before, h.GetHandNumber(), "終局後は進まない")
}

func TestHokm_TieHasNoWinner(t *testing.T) {
	h := newTestHokm(t)
	h.SetScoreForTestUse(0, 5)
	h.SetScoreForTestUse(1, 5)
	h.FinishGameForTest()
	assert.Equal(t, -1, h.GetWinnerTeam())
}

// --- プレイ ---

func TestHokm_MustFollowSuit(t *testing.T) {
	h := newTestHokm(t)
	h.SetTrumpSuitForTest(CardDesignHeart)
	h.SetPhaseForTest(HokmPhasePlay)
	h.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}})
	h.SetCurrentPlayerIdxForTest(0)
	hokmHandOf(h, 0, NewCard(CardDesignHeart, 1, false), NewCard(CardDesignSpade, 7, false))

	require.Error(t, h.PlayerPlay(0))
	require.NoError(t, h.PlayerPlay(1))
}

func TestHokm_MayDiscardWhenVoid(t *testing.T) {
	h := newTestHokm(t)
	h.SetTrumpSuitForTest(CardDesignHeart)
	h.SetPhaseForTest(HokmPhasePlay)
	h.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}})
	h.SetCurrentPlayerIdxForTest(0)
	hokmHandOf(h, 0, NewCard(CardDesignHeart, 1, false), NewCard(CardDesignClover, 7, false))

	assert.Equal(t, []int{0, 1}, h.GetValidPlayIndices(0))
}

func TestHokm_TrumpBeatsTheLeadSuit(t *testing.T) {
	h := newTestHokm(t)
	h.SetTrumpSuitForTest(CardDesignHeart)
	h.SetPhaseForTest(HokmPhasePlay)
	h.SetCurrentPlayerIdxForTest(0)
	h.SetLeadPlayerIdxForTest(0)

	hokmHandOf(h, 0, NewCard(CardDesignSpade, 1, false))
	hokmHandOf(h, 1, NewCard(CardDesignSpade, 13, false))
	hokmHandOf(h, 2, NewCard(CardDesignHeart, 2, false))
	hokmHandOf(h, 3, NewCard(CardDesignSpade, 12, false))

	for i := range HokmPlayerCnt {
		require.NoError(t, h.PlayForTest(i, 0))
	}
	assert.Equal(t, 1, h.GetPlayer(2).GetTrickCount(), "切り札の 2 が A に勝つ")
}

func TestHokm_PlayRejectsInvalidIndex(t *testing.T) {
	h := newTestHokm(t)
	settleTrump(t, h)
	h.SetCurrentPlayerIdxForTest(0)

	assert.Error(t, h.PlayerPlay(-1))
	assert.Error(t, h.PlayerPlay(99))
}

func TestHokm_PlayGuards(t *testing.T) {
	h := newTestHokm(t)
	assert.Error(t, h.PlayerPlay(0), "宣言フェーズでは出せない")

	settleTrump(t, h)
	h.SetCurrentPlayerIdxForTest(1)
	assert.Error(t, h.PlayerPlay(0))

	h.SetCurrentPlayerIdxForTest(0)
	h.GiveUp()
	assert.Error(t, h.PlayerPlay(0))
}

func TestHokm_ValidIndicesOutOfRange(t *testing.T) {
	h := newTestHokm(t)
	assert.Nil(t, h.GetValidPlayIndices(-1))
	assert.Nil(t, h.GetValidPlayIndices(HokmPlayerCnt))
}

func TestHokm_CpuPlayIgnoresHumanTurn(t *testing.T) {
	h := newTestHokm(t)
	settleTrump(t, h)
	h.SetCurrentPlayerIdxForTest(0)
	size := h.GetPlayer(0).GetCardsSize()

	h.CpuPlay()
	assert.Equal(t, size, h.GetPlayer(0).GetCardsSize())
}

// **味方が勝っていれば安い札を落とす。** そうでなければ一番安く取りに行く。
func TestHokm_CpuSavesCardsWhenThePartnerIsWinning(t *testing.T) {
	build := func(winner int) *Hokm {
		h := newTestHokm(t)
		h.SetTrumpSuitForTest(CardDesignHeart)
		h.SetPhaseForTest(HokmPhasePlay)
		// **勝てる札を用意する。** A を置くと K で取れず、どちらの分岐でも
		// 「一番安い札を落とす」に落ちてしまい、テストが何も区別しない。
		h.SetCurrentTrickForTest([]*TrickCard{
			{PlayerIdx: winner, Card: NewCard(CardDesignSpade, 12, false)},
			{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 3, false)},
		})
		h.SetCurrentPlayerIdxForTest(2)
		hokmHandOf(h, 2,
			NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignSpade, 2, false))
		return h
	}

	// 味方 (0) が勝っている → 安い札。
	assert.True(t, build(0).partnerIsWinning(2))
	assert.Equal(t, 1, build(0).CpuChoiceForTest(2))

	// 相手 (3) が勝っている → K で取りに行く。
	h := build(3)
	assert.False(t, h.partnerIsWinning(2))
	assert.Equal(t, 0, h.CpuChoiceForTest(2))

	// **取れないなら安い札を落とす。** A が出ていれば K でも勝てない。
	cannot := build(3)
	cannot.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 3, false)},
	})
	assert.Equal(t, 1, cannot.CpuChoiceForTest(2))
}

// --- ヒント ---

func TestHokm_HintDuringTrumpDeclaration(t *testing.T) {
	h := newTestHokm(t)
	h.SetHakemIdxForTest(0)

	hint := h.GetHint()
	require.NotNil(t, hint)
	assert.Nil(t, hint.CardIndex)
	assert.Equal(t, "hokmDeclareTrump", hint.Reason)
	assert.GreaterOrEqual(t, hint.Suit, CardDesignSpade)
}

func TestHokm_HintDuringPlay(t *testing.T) {
	h := newTestHokm(t)
	settleTrump(t, h)
	h.SetCurrentPlayerIdxForTest(0)

	hint := h.GetHint()
	require.NotNil(t, hint)
	require.NotNil(t, hint.CardIndex)
	assert.Contains(t, h.GetValidPlayIndices(0), *hint.CardIndex)
	assert.Contains(t, []string{"hokmWinTrick", "hokmSaveCards"}, hint.Reason)
}

func TestHokm_HintNilWhenNotHumanTurn(t *testing.T) {
	h := newTestHokm(t)
	settleTrump(t, h)
	h.SetCurrentPlayerIdxForTest(2)
	assert.Nil(t, h.GetHint())

	h.GiveUp()
	assert.Nil(t, h.GetHint())
}

// --- その他 ---

func TestHokm_GiveUp(t *testing.T) {
	h := newTestHokm(t)
	h.GiveUp()
	assert.True(t, h.GetGameEndFlag())
	assert.Equal(t, 1, h.GetWinnerTeam())

	h.SetWinnerTeamForTest(0)
	h.GiveUp()
	assert.Equal(t, 0, h.GetWinnerTeam(), "二度目は何も起きない")
}

func TestHokm_AccessorsOutOfRange(t *testing.T) {
	h := newTestHokm(t)
	assert.Nil(t, h.GetPlayer(-1))
	assert.Nil(t, h.GetPlayer(HokmPlayerCnt))
	assert.Equal(t, HokmPlayerCnt, h.GetPlayerCnt())
	assert.Equal(t, 0, h.GetScore(-1))
	assert.Equal(t, 0, h.GetScore(HokmTeamCnt))
	assert.Equal(t, 1, h.GetHandNumber())
	assert.Empty(t, h.GetCurrentTrick())
	assert.Equal(t, -1, h.GetLastHandWinner())

	h.SetScoreForTestUse(0, 4)
	assert.Equal(t, 4, h.GetScore(0))
	h.SetScoreForTestUse(-1, 9)
	assert.Equal(t, 4, h.GetScore(0))
}

func TestHokmConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultHokmConfig().Validate())
	assert.NoError(t, HokmConfig{Target: HokmTargetMin}.Validate())
	assert.NoError(t, HokmConfig{Target: HokmTargetMax}.Validate())
	assert.Error(t, HokmConfig{Target: HokmTargetMin - 1}.Validate())
	assert.Error(t, HokmConfig{Target: HokmTargetMax + 1}.Validate())
}

// --- JSON 往復 ---

func TestHokm_JSONRoundTrip(t *testing.T) {
	h := newTestHokm(t)
	h.SetHakemIdxForTest(0)
	require.NoError(t, h.PlayerDeclareTrump(CardDesignDiamond))
	h.SetScoreForTestUse(0, 3)
	h.GiveTricksForTest(0, 2)

	data, err := json.Marshal(h)
	require.NoError(t, err)

	var got Hokm
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, h.GetPhase(), got.GetPhase())
	assert.Equal(t, CardDesignDiamond, got.GetTrumpSuit())
	assert.Equal(t, 0, got.GetHakemIdx())
	assert.Equal(t, 3, got.GetScore(0))
	// **獲得トリック数が往復しないと 7 先取も Kot も判定できない。**
	assert.Equal(t, 2, got.GetPlayer(0).GetTrickCount())
	assert.Equal(t, h.GetConfig().Target, got.GetConfig().Target)
	for i := range HokmPlayerCnt {
		assert.Equal(t, h.GetPlayer(i).GetCardsSize(), got.GetPlayer(i).GetCardsSize(), "player %d", i)
	}
}

func TestHokm_UnmarshalRejectsInvalid(t *testing.T) {
	valid := func() hokmJSON {
		return hokmJSON{
			Config:         DefaultHokmConfig(),
			Phase:          HokmPhasePlay,
			HandNumber:     1,
			TrumpSuit:      CardDesignSpade,
			WinnerTeam:     -1,
			LastHandWinner: -1,
		}
	}
	cases := map[string]func(*hokmJSON){
		"bad config": func(j *hokmJSON) { j.Config.Target = 0 },
		"bad phase":  func(j *hokmJSON) { j.Phase = HokmPhase(99) },
		"bad trick":  func(j *hokmJSON) { j.TrickNumber = HokmHandSize + 1 },
		"bad hand":   func(j *hokmJSON) { j.HandNumber = 0 },
		"bad current": func(j *hokmJSON) {
			j.CurrentPlayerIdx = HokmPlayerCnt
		},
		"bad lead":        func(j *hokmJSON) { j.LeadPlayerIdx = -1 },
		"bad hakem":       func(j *hokmJSON) { j.HakemIdx = HokmPlayerCnt },
		"bad winner":      func(j *hokmJSON) { j.WinnerTeam = HokmTeamCnt },
		"bad last winner": func(j *hokmJSON) { j.LastHandWinner = HokmTeamCnt },
		// **切り札はフェーズと整合していなければならない。** 両方向を踏む。
		"trump before it was declared": func(j *hokmJSON) {
			j.Phase, j.TrumpSuit = HokmPhaseTrump, CardDesignHeart
		},
		"no trump after it was declared": func(j *hokmJSON) {
			j.Phase, j.TrumpSuit = HokmPhasePlay, 0
		},
		"bogus trump": func(j *hokmJSON) { j.TrumpSuit = 99 },
		"long trick": func(j *hokmJSON) {
			j.CurrentTrick = make([]*TrickCard, HokmPlayerCnt+1)
		},
		"long log": func(j *hokmJSON) {
			j.ActionLog = make([]*ActionLogEntry, hokmMaxSliceLen+1)
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			j := valid()
			mutate(&j)
			data, err := json.Marshal(j)
			require.NoError(t, err)
			var got Hokm
			assert.Error(t, json.Unmarshal(data, &got))
		})
	}

	var got Hokm
	assert.Error(t, got.UnmarshalJSON([]byte("{")))

	// 正のコントロール: 正しいスナップショットは通る。
	data, err := json.Marshal(valid())
	require.NoError(t, err)
	assert.NoError(t, json.Unmarshal(data, &got))

	// 宣言前で切り札 0 も通る（ガードが 0 を一律に弾いていないこと）。
	j := valid()
	j.Phase, j.TrumpSuit = HokmPhaseTrump, 0
	pre, err := json.Marshal(j)
	require.NoError(t, err)
	var okPre Hokm
	assert.NoError(t, json.Unmarshal(pre, &okPre))
}

func TestHokm_ActionLog(t *testing.T) {
	h := newTestHokm(t)
	require.NotEmpty(t, h.actionLog, "配りが記録される")
	h.SetHakemIdxForTest(0)
	require.NoError(t, h.PlayerDeclareTrump(CardDesignSpade))

	kinds := map[string]bool{}
	for _, e := range h.actionLog {
		kinds[e.ActionType] = true
	}
	assert.True(t, kinds["trump"])
}

// **親は負けたときだけ交代する** (#5753)。次に切り札を選ぶ席が変わるかどうかを
// 表示できるよう、その事実を残す。
func TestHokmRecordsWhetherTheHakemChanged(t *testing.T) {
	t.Run("the hakem's team loses", func(t *testing.T) {
		h := newTestHokm(t)
		before := h.GetHakemIdx()
		h.FinishHandForTest(1 - HokmTeamOf(before))

		if !h.GetLastHandHakemChanged() {
			t.Error("losing the hand must move the hakem")
		}
		if h.GetHakemIdx() == before {
			t.Errorf("the hakem stayed at %d", before)
		}
	})

	t.Run("the hakem's team wins", func(t *testing.T) {
		h := newTestHokm(t)
		before := h.GetHakemIdx()
		h.FinishHandForTest(HokmTeamOf(before))

		if h.GetLastHandHakemChanged() {
			t.Error("winning the hand must keep the hakem")
		}
		if h.GetHakemIdx() != before {
			t.Errorf("the hakem moved to %d", h.GetHakemIdx())
		}
	})
}
