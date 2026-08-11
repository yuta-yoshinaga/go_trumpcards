//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMendikot(t *testing.T) *Mendikot {
	t.Helper()
	m := NewDefaultMendikot()
	m.Reset()
	return m
}

// mendikotHandOf は指定プレイヤーの手札を固定の並びに差し替える。
func mendikotHandOf(m *Mendikot, idx int, cards ...*Card) {
	p := m.GetPlayer(idx)
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// --- 配り ---

func TestMendikot_DealsThirteenEachWithNoTrump(t *testing.T) {
	m := newTestMendikot(t)

	total := 0
	for i := range MendikotPlayerCnt {
		assert.Equal(t, MendikotHandSize, m.GetPlayer(i).GetCardsSize(), "player %d", i)
		total += m.GetPlayer(i).GetCardsSize()
	}
	assert.Equal(t, 52, total)
	// **切り札は配られた時点では存在しない。**
	assert.Equal(t, 0, m.GetTrumpSuit())
	assert.Equal(t, -1, m.GetTrumpChooserIdx())
	assert.Equal(t, MendikotPhasePlay, m.GetPhase())
}

// デッキには 10 がちょうど 4 枚ある。
func TestMendikot_FourTensInTheDeck(t *testing.T) {
	m := newTestMendikot(t)
	tens := 0
	for i := range MendikotPlayerCnt {
		p := m.GetPlayer(i)
		for j := range p.GetCardsSize() {
			if p.GetCard(j).GetValue() == 10 {
				tens++
			}
		}
	}
	assert.Equal(t, MendikotTensInDeck, tens)
}

// --- 切り札の決まり方 ---

// **最初にフォローできなかった人が出した札のスートが切り札になる。**
func TestMendikot_TrumpIsSetByTheFirstPlayerWhoCannotFollow(t *testing.T) {
	m := newTestMendikot(t)
	m.SetLeadPlayerIdxForTest(0)
	m.SetCurrentPlayerIdxForTest(0)

	mendikotHandOf(m, 0, NewCard(CardDesignSpade, 5, false))
	mendikotHandOf(m, 1, NewCard(CardDesignSpade, 7, false))
	// プレイヤー2 はスペードを持っていない。
	mendikotHandOf(m, 2, NewCard(CardDesignHeart, 3, false))
	mendikotHandOf(m, 3, NewCard(CardDesignSpade, 9, false))

	require.NoError(t, m.PlayForTest(0, 0))
	require.NoError(t, m.PlayForTest(1, 0))
	assert.Equal(t, 0, m.GetTrumpSuit(), "まだ誰もフォローに失敗していない")

	require.NoError(t, m.PlayForTest(2, 0))
	assert.Equal(t, CardDesignHeart, m.GetTrumpSuit(), "出した札のスートが切り札")
	assert.Equal(t, 2, m.GetTrumpChooserIdx())

	require.NoError(t, m.PlayForTest(3, 0))
	// **切り札はそのトリックから効く。** ♥3 が ♠9 に勝つ。
	assert.Equal(t, 1, m.GetPlayer(2).GetTrickCount())
}

// **リードした本人は切り札を決められない。** 決められてしまうと、切りたい
// スートを持っていないふりをして 1 枚目に出すだけで切り札を選べることになり、
// 「フォローできなかった人が決める」という規則そのものが消える。
//
// いまこれを防いでいるのは resolveTrick が currentTrick を空にしてから次の
// リードに手番を渡すことだけで、**needsTrump 側には何の防御も無い**
// （レビュー指摘 PR #5307）。トリックの切り替えを触ると静かに壊れる。
func TestMendikot_TheLeadPlayerCannotSetTrump(t *testing.T) {
	m := newTestMendikot(t)
	m.SetLeadPlayerIdxForTest(0)
	m.SetCurrentPlayerIdxForTest(0)

	for i := range MendikotPlayerCnt {
		mendikotHandOf(m, i, NewCard(CardDesignSpade, 5+i, false), NewCard(CardDesignHeart, 2+i, false))
	}

	// リードが ♥ を出しても切り札にはならない（フォロー失敗ではない）。
	require.NoError(t, m.PlayForTest(0, 1))
	assert.Equal(t, 0, m.GetTrumpSuit(), "リードは切り札を決められない")
	assert.Equal(t, -1, m.GetTrumpChooserIdx())

	// 負のコントロール: 2 人目が ♥ を持っていないと、そこで決まる。
	m2 := newTestMendikot(t)
	m2.SetLeadPlayerIdxForTest(0)
	m2.SetCurrentPlayerIdxForTest(0)
	mendikotHandOf(m2, 0, NewCard(CardDesignHeart, 4, false))
	mendikotHandOf(m2, 1, NewCard(CardDesignClover, 9, false))
	mendikotHandOf(m2, 2, NewCard(CardDesignHeart, 6, false))
	mendikotHandOf(m2, 3, NewCard(CardDesignHeart, 7, false))
	require.NoError(t, m2.PlayForTest(0, 0))
	assert.Equal(t, 0, m2.GetTrumpSuit())
	require.NoError(t, m2.PlayForTest(1, 0))
	assert.Equal(t, CardDesignClover, m2.GetTrumpSuit(), "2 人目のフォロー失敗で決まる")
	assert.Equal(t, 1, m2.GetTrumpChooserIdx())
}

// **切り札が決まったら、以降は変わらない。**
func TestMendikot_TrumpIsFixedOnceChosen(t *testing.T) {
	m := newTestMendikot(t)
	m.SetTrumpForTest(CardDesignHeart, 2)
	m.SetLeadPlayerIdxForTest(0)
	m.SetCurrentPlayerIdxForTest(0)

	mendikotHandOf(m, 0, NewCard(CardDesignSpade, 5, false))
	mendikotHandOf(m, 1, NewCard(CardDesignSpade, 7, false))
	mendikotHandOf(m, 2, NewCard(CardDesignClover, 3, false))
	mendikotHandOf(m, 3, NewCard(CardDesignSpade, 9, false))

	for i := range MendikotPlayerCnt {
		require.NoError(t, m.PlayForTest(i, 0))
	}
	assert.Equal(t, CardDesignHeart, m.GetTrumpSuit(), "♣ を出しても切り札は変わらない")
	assert.Equal(t, 2, m.GetTrumpChooserIdx())
}

// **切り札が決まるまではリードのスートだけが勝ちうる。**
func TestMendikot_NoTrumpMeansLeadSuitWins(t *testing.T) {
	m := newTestMendikot(t)
	m.SetLeadPlayerIdxForTest(0)
	m.SetCurrentPlayerIdxForTest(0)

	// 全員がリードのスートを持っているので、切り札は決まらない。
	mendikotHandOf(m, 0, NewCard(CardDesignSpade, 5, false))
	mendikotHandOf(m, 1, NewCard(CardDesignSpade, 1, false))
	mendikotHandOf(m, 2, NewCard(CardDesignSpade, 3, false))
	mendikotHandOf(m, 3, NewCard(CardDesignSpade, 9, false))

	for i := range MendikotPlayerCnt {
		require.NoError(t, m.PlayForTest(i, 0))
	}
	assert.Equal(t, 0, m.GetTrumpSuit())
	assert.Equal(t, 1, m.GetPlayer(1).GetTrickCount(), "A が勝つ")
}

// --- 勝敗判定 ---

// **issue の「10 を 3 枚かつ 7 トリック以上」では決着しないハンドが出る。**
// 3 枚あればトリック数に関係なく勝ち、2 枚ずつのときだけトリックで決まる。
func TestMendikot_ResultTable(t *testing.T) {
	for _, tc := range []struct {
		name                           string
		tens0, tens1, tricks0, tricks1 int
		wantTeam, wantPoints           int
		wantKind                       string
	}{
		{"three tens wins even with a minority of tricks", 3, 1, 5, 8, 0, MendikotWinPoints, "tens"},
		{"three tens for the other side", 1, 3, 8, 5, 1, MendikotWinPoints, "tens"},
		{"all four tens is a mendikot", 4, 0, 7, 6, 0, MendikotMendikotPoints, "mendikot"},
		{"all four tens for the other side", 0, 4, 6, 7, 1, MendikotMendikotPoints, "mendikot"},
		{"two each is settled by tricks", 2, 2, 7, 6, 0, MendikotWinPoints, "tricks"},
		{"two each the other way", 2, 2, 6, 7, 1, MendikotWinPoints, "tricks"},
		{"every trick is a whitewash", 4, 0, 13, 0, 0, MendikotWhitewashPoints, "whitewash"},
		{"a whitewash outranks a mendikot", 0, 4, 0, 13, 1, MendikotWhitewashPoints, "whitewash"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := MendikotResultFor(tc.tens0, tc.tens1, tc.tricks0, tc.tricks1)
			assert.Equal(t, tc.wantTeam, got.WinnerTeam)
			assert.Equal(t, tc.wantPoints, got.Points)
			assert.Equal(t, tc.wantKind, got.Kind)
		})
	}
}

// **引き分けは起こりえない。** 10 の分け方すべてで必ず勝者が出る。
func TestMendikot_EveryDistributionDecides(t *testing.T) {
	for tens0 := 0; tens0 <= MendikotTensInDeck; tens0++ {
		tens1 := MendikotTensInDeck - tens0
		for tricks0 := 0; tricks0 <= MendikotTricksPerRound; tricks0++ {
			tricks1 := MendikotTricksPerRound - tricks0
			got := MendikotResultFor(tens0, tens1, tricks0, tricks1)
			assert.Contains(t, []int{0, 1}, got.WinnerTeam,
				"tens %d-%d tricks %d-%d", tens0, tens1, tricks0, tricks1)
			assert.Positive(t, got.Points)
		}
	}
}

// 10 を取ると枚数が数えられる。
func TestMendikot_TensAreCounted(t *testing.T) {
	m := newTestMendikot(t)
	m.SetTrumpForTest(CardDesignHeart, 2)
	m.SetLeadPlayerIdxForTest(0)
	m.SetCurrentPlayerIdxForTest(0)

	mendikotHandOf(m, 0, NewCard(CardDesignSpade, 1, false))
	mendikotHandOf(m, 1, NewCard(CardDesignSpade, 10, false))
	mendikotHandOf(m, 2, NewCard(CardDesignSpade, 10, false))
	mendikotHandOf(m, 3, NewCard(CardDesignSpade, 3, false))

	for i := range MendikotPlayerCnt {
		require.NoError(t, m.PlayForTest(i, 0))
	}
	assert.Equal(t, 2, m.GetPlayer(0).GetTens(), "A で取った側に 10 が 2 枚入る")
	assert.Equal(t, 2, m.TeamTens(0))
	assert.Equal(t, 0, m.TeamTens(1))
}

// ハンド精算で勝ち点が入る。
func TestMendikot_FinishHandScores(t *testing.T) {
	m := newTestMendikot(t)
	m.GetPlayer(0).SetTens(3)
	m.GetPlayer(1).SetTens(1)
	m.GiveTricksForTest(0, 5)
	m.GiveTricksForTest(1, 8)
	m.FinishHandForTest()

	assert.Equal(t, MendikotWinPoints, m.GetScore(0), "10 が 3 枚あればトリックが少なくても勝ち")
	assert.Equal(t, 0, m.GetScore(1))
	assert.Equal(t, 0, m.GetLastHandWinner())
	assert.Equal(t, "tens", m.GetLastHandKind())
}

// **親は負けたチームへ移る。** 勝ったチームは親を守る。
func TestMendikot_DealerMovesOnLoss(t *testing.T) {
	t.Run("the dealer's team wins", func(t *testing.T) {
		m := newTestMendikot(t)
		m.SetDealerIdxForTest(0)
		m.GetPlayer(0).SetTens(3)
		m.GetPlayer(1).SetTens(1)
		m.FinishHandForTest()
		assert.Equal(t, 1, m.GetDealerIdx(), "勝ったので次へ渡す")
	})

	t.Run("the dealer's team loses", func(t *testing.T) {
		m := newTestMendikot(t)
		m.SetDealerIdxForTest(1)
		m.GetPlayer(0).SetTens(3)
		m.GetPlayer(1).SetTens(1)
		m.FinishHandForTest()
		assert.Equal(t, 1, m.GetDealerIdx(), "負けた側は親のまま")
	})
}

// 規定点に届かなければ次のハンドへ、届けば終局。
func TestMendikot_NextHandAndGameEnd(t *testing.T) {
	m := newTestMendikot(t)
	m.SetConfig(MendikotConfig{Target: MendikotTargetMax})
	m.GetPlayer(0).SetTens(4)
	m.FinishHandForTest()
	require.Equal(t, MendikotPhaseHandEnd, m.GetPhase())

	m.NextHand()
	assert.Equal(t, 2, m.GetHandNumber())
	assert.Equal(t, MendikotPhasePlay, m.GetPhase())
	assert.Equal(t, 0, m.GetTrumpSuit(), "切り札はハンドごとに決め直す")
	assert.Equal(t, -1, m.GetTrumpChooserIdx())
	for i := range MendikotPlayerCnt {
		assert.Equal(t, 0, m.GetPlayer(i).GetTens(), "10 の枚数も白紙")
	}
}

func TestMendikot_GameEndsAtTarget(t *testing.T) {
	m := newTestMendikot(t)
	m.SetConfig(MendikotConfig{Target: MendikotTargetMin})
	m.GetPlayer(0).SetTens(3)
	m.GetPlayer(1).SetTens(1)
	m.FinishHandForTest()

	assert.True(t, m.GetGameEndFlag())
	assert.Equal(t, MendikotPhaseGameEnd, m.GetPhase())
	assert.Equal(t, 0, m.GetWinnerTeam())

	before := m.GetHandNumber()
	m.NextHand()
	assert.Equal(t, before, m.GetHandNumber())
}

func TestMendikot_TieHasNoWinner(t *testing.T) {
	m := newTestMendikot(t)
	m.SetScoreForTestUse(0, 3)
	m.SetScoreForTestUse(1, 3)
	m.FinishGameForTest()
	assert.Equal(t, -1, m.GetWinnerTeam())
}

// 1 ハンド通すと 10 は 4 枚、トリックは 13 個がちょうど分配される。
func TestMendikot_AHandDistributesEverything(t *testing.T) {
	m := newTestMendikot(t)
	for m.GetPhase() == MendikotPhasePlay {
		if m.IsHumanTurn() {
			valid := m.GetValidPlayIndices(0)
			require.NotEmpty(t, valid)
			require.NoError(t, m.PlayerPlay(valid[0]))
			continue
		}
		m.CpuPlay()
	}

	assert.Equal(t, MendikotTricksPerRound, m.GetTrickNumber())
	assert.Equal(t, MendikotTensInDeck, m.TeamTens(0)+m.TeamTens(1))
	assert.Equal(t, MendikotTricksPerRound, m.TeamTricks(0)+m.TeamTricks(1))
	assert.Contains(t, []int{0, 1}, m.GetLastHandWinner())
}

// --- プレイ ---

func TestMendikot_MustFollowSuit(t *testing.T) {
	m := newTestMendikot(t)
	m.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)}})
	m.SetCurrentPlayerIdxForTest(0)
	mendikotHandOf(m, 0, NewCard(CardDesignHeart, 1, false), NewCard(CardDesignSpade, 7, false))

	require.Error(t, m.PlayerPlay(0))
	require.NoError(t, m.PlayerPlay(1))
}

func TestMendikot_PlayRejectsInvalidIndex(t *testing.T) {
	m := newTestMendikot(t)
	m.SetCurrentPlayerIdxForTest(0)
	assert.Error(t, m.PlayerPlay(-1))
	assert.Error(t, m.PlayerPlay(99))
}

func TestMendikot_PlayGuards(t *testing.T) {
	m := newTestMendikot(t)
	m.SetCurrentPlayerIdxForTest(1)
	assert.Error(t, m.PlayerPlay(0))

	m.SetCurrentPlayerIdxForTest(0)
	m.SetPhaseForTest(MendikotPhaseHandEnd)
	assert.Error(t, m.PlayerPlay(0))

	m.SetPhaseForTest(MendikotPhasePlay)
	m.GiveUp()
	assert.Error(t, m.PlayerPlay(0))
}

func TestMendikot_ValidIndicesOutOfRange(t *testing.T) {
	m := newTestMendikot(t)
	assert.Nil(t, m.GetValidPlayIndices(-1))
	assert.Nil(t, m.GetValidPlayIndices(MendikotPlayerCnt))
}

func TestMendikot_CpuPlayIgnoresHumanTurn(t *testing.T) {
	m := newTestMendikot(t)
	m.SetCurrentPlayerIdxForTest(0)
	size := m.GetPlayer(0).GetCardsSize()
	m.CpuPlay()
	assert.Equal(t, size, m.GetPlayer(0).GetCardsSize())
}

// **10 が出ているトリックは取りに行く。** 勝敗そのものなので。
func TestMendikot_CpuChasesTricksHoldingATen(t *testing.T) {
	m := newTestMendikot(t)
	m.SetTrumpForTest(CardDesignHeart, 1)
	m.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 10, false)},
	})
	m.SetCurrentPlayerIdxForTest(2)
	mendikotHandOf(m, 2,
		NewCard(CardDesignSpade, 3, false),  // 勝てない
		NewCard(CardDesignSpade, 11, false)) // 勝てる

	assert.True(t, m.trickHasTen())
	assert.Equal(t, 1, m.CpuChoiceForTest(2))
}

// **取れないなら 10 を捨てない。** 相手に渡してしまう。
func TestMendikot_CpuKeepsItsTensWhenItCannotWin(t *testing.T) {
	m := newTestMendikot(t)
	m.SetTrumpForTest(CardDesignHeart, 1)
	m.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 1, false)},
	})
	m.SetCurrentPlayerIdxForTest(2)
	mendikotHandOf(m, 2,
		NewCard(CardDesignSpade, 10, false), // 10 は温存する
		NewCard(CardDesignSpade, 3, false))

	assert.Equal(t, 1, m.CpuChoiceForTest(2))
}

// **味方が勝っているなら 10 を乗せてよい。**
func TestMendikot_CpuFeedsTensToAWinningPartner(t *testing.T) {
	m := newTestMendikot(t)
	m.SetTrumpForTest(CardDesignHeart, 1)
	m.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 2, false)},
	})
	m.SetCurrentPlayerIdxForTest(2)
	mendikotHandOf(m, 2,
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignSpade, 10, false))

	assert.True(t, m.partnerIsWinning(2))
	assert.Equal(t, 1, m.CpuChoiceForTest(2))
}

// --- ヒント ---

func TestMendikot_HintDuringPlay(t *testing.T) {
	m := newTestMendikot(t)
	m.SetCurrentPlayerIdxForTest(0)

	h := m.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	assert.Contains(t, m.GetValidPlayIndices(0), *h.CardIndex)
	assert.Contains(t, []string{"mendikotChaseTen", "mendikotFeedPartner", "mendikotDuck"}, h.Reason)
}

// 10 が出ているトリックでは理由が変わる。
func TestMendikot_HintChasesATen(t *testing.T) {
	m := newTestMendikot(t)
	m.SetTrumpForTest(CardDesignHeart, 1)
	m.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 10, false)}})
	m.SetCurrentPlayerIdxForTest(0)
	mendikotHandOf(m, 0, NewCard(CardDesignSpade, 11, false), NewCard(CardDesignSpade, 3, false))

	h := m.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "mendikotChaseTen", h.Reason)
}

func TestMendikot_HintNilWhenNotHumanTurn(t *testing.T) {
	m := newTestMendikot(t)
	m.SetCurrentPlayerIdxForTest(2)
	assert.Nil(t, m.GetHint())

	m.GiveUp()
	assert.Nil(t, m.GetHint())
}

// --- その他 ---

func TestMendikot_GiveUp(t *testing.T) {
	m := newTestMendikot(t)
	m.GiveUp()
	assert.True(t, m.GetGameEndFlag())
	assert.Equal(t, 1, m.GetWinnerTeam())

	m.SetWinnerTeamForTest(0)
	m.GiveUp()
	assert.Equal(t, 0, m.GetWinnerTeam(), "二度目は何も起きない")
}

func TestMendikot_TeamAssignment(t *testing.T) {
	assert.Equal(t, MendikotTeamOf(0), MendikotTeamOf(2))
	assert.Equal(t, MendikotTeamOf(1), MendikotTeamOf(3))
	assert.NotEqual(t, MendikotTeamOf(0), MendikotTeamOf(1))
}

func TestMendikot_AccessorsOutOfRange(t *testing.T) {
	m := newTestMendikot(t)
	assert.Nil(t, m.GetPlayer(-1))
	assert.Nil(t, m.GetPlayer(MendikotPlayerCnt))
	assert.Equal(t, MendikotPlayerCnt, m.GetPlayerCnt())
	assert.Equal(t, 0, m.GetScore(-1))
	assert.Equal(t, 0, m.GetScore(MendikotTeamCnt))
	assert.Equal(t, 0, m.TeamTens(-1))
	assert.Equal(t, 0, m.TeamTens(MendikotTeamCnt))
	assert.Equal(t, 0, m.TeamTricks(-1))
	assert.Equal(t, 1, m.GetHandNumber())
	assert.Empty(t, m.GetCurrentTrick())
	assert.Empty(t, m.GetLastHandKind())
}

func TestMendikotConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultMendikotConfig().Validate())
	assert.NoError(t, MendikotConfig{Target: MendikotTargetMin}.Validate())
	assert.NoError(t, MendikotConfig{Target: MendikotTargetMax}.Validate())
	assert.Error(t, MendikotConfig{Target: MendikotTargetMin - 1}.Validate())
	assert.Error(t, MendikotConfig{Target: MendikotTargetMax + 1}.Validate())
}

// --- JSON 往復 ---

func TestMendikot_JSONRoundTrip(t *testing.T) {
	m := newTestMendikot(t)
	m.SetTrumpForTest(CardDesignDiamond, 2)
	m.GetPlayer(0).SetTens(2)
	m.GiveTricksForTest(0, 3)
	m.SetScoreForTestUse(0, 2)

	data, err := json.Marshal(m)
	require.NoError(t, err)

	var got Mendikot
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, CardDesignDiamond, got.GetTrumpSuit())
	assert.Equal(t, 2, got.GetTrumpChooserIdx())
	// **10 の枚数は勝敗そのもの。** 往復しないと勝ったハンドが勝ちでなくなる。
	assert.Equal(t, 2, got.GetPlayer(0).GetTens())
	assert.Equal(t, 3, got.TeamTricks(0))
	assert.Equal(t, 2, got.GetScore(0))
	assert.Equal(t, m.GetConfig().Target, got.GetConfig().Target)
}

func TestMendikot_UnmarshalRejectsInvalid(t *testing.T) {
	valid := func() mendikotJSON {
		return mendikotJSON{
			Config:          DefaultMendikotConfig(),
			Phase:           MendikotPhasePlay,
			HandNumber:      1,
			TrumpSuit:       CardDesignSpade,
			TrumpChooserIdx: 2,
			WinnerTeam:      -1,
			LastHandWinner:  -1,
		}
	}
	cases := map[string]func(*mendikotJSON){
		"bad config":      func(j *mendikotJSON) { j.Config.Target = 0 },
		"bad phase":       func(j *mendikotJSON) { j.Phase = MendikotPhase(99) },
		"bad trick":       func(j *mendikotJSON) { j.TrickNumber = MendikotTricksPerRound + 1 },
		"bad hand":        func(j *mendikotJSON) { j.HandNumber = 0 },
		"bad current":     func(j *mendikotJSON) { j.CurrentPlayerIdx = MendikotPlayerCnt },
		"bad lead":        func(j *mendikotJSON) { j.LeadPlayerIdx = -1 },
		"bad dealer":      func(j *mendikotJSON) { j.DealerIdx = MendikotPlayerCnt },
		"bad winner":      func(j *mendikotJSON) { j.WinnerTeam = MendikotTeamCnt },
		"bad last winner": func(j *mendikotJSON) { j.LastHandWinner = MendikotTeamCnt },
		// **切り札と決定者は対で整合していなければならない。** 両方向を踏む。
		"bogus trump": func(j *mendikotJSON) { j.TrumpSuit = 99 },
		"trump without a chooser": func(j *mendikotJSON) {
			j.TrumpSuit, j.TrumpChooserIdx = CardDesignHeart, -1
		},
		"chooser without a trump": func(j *mendikotJSON) {
			j.TrumpSuit, j.TrumpChooserIdx = 0, 2
		},
		"long trick": func(j *mendikotJSON) {
			j.CurrentTrick = make([]*TrickCard, MendikotPlayerCnt+1)
		},
		"long log": func(j *mendikotJSON) {
			j.ActionLog = make([]*ActionLogEntry, mendikotMaxSliceLen+1)
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			j := valid()
			mutate(&j)
			data, err := json.Marshal(j)
			require.NoError(t, err)
			var got Mendikot
			assert.Error(t, json.Unmarshal(data, &got))
		})
	}

	var got Mendikot
	assert.Error(t, got.UnmarshalJSON([]byte("{")))

	// 正のコントロール: 正しいスナップショットは通る。
	data, err := json.Marshal(valid())
	require.NoError(t, err)
	assert.NoError(t, json.Unmarshal(data, &got))

	// **切り札未決定（0 + 決定者なし）も正当な状態。** 一律に弾いていないこと。
	j := valid()
	j.TrumpSuit, j.TrumpChooserIdx = 0, -1
	pre, err := json.Marshal(j)
	require.NoError(t, err)
	var okPre Mendikot
	assert.NoError(t, json.Unmarshal(pre, &okPre))
}

func TestMendikot_ActionLog(t *testing.T) {
	m := newTestMendikot(t)
	require.NotEmpty(t, m.actionLog, "配りが記録される")
	m.SetCurrentPlayerIdxForTest(0)
	require.NoError(t, m.PlayerPlay(m.GetValidPlayIndices(0)[0]))

	kinds := map[string]bool{}
	for _, e := range m.actionLog {
		kinds[e.ActionType] = true
	}
	assert.True(t, kinds["play"])
}
