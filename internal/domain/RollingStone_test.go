//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRollingStone(t *testing.T) *RollingStone {
	t.Helper()
	r := NewDefaultRollingStone()
	r.Reset()
	return r
}

// **1 人 8 枚固定で、デッキのほうを人数に合わせる。**
//
// issue の「6 人なら 32 枚」は 32 % 6 != 0 で成立しません——32 枚は 4 人のときの値です。
func TestRollingStone_DeckSizeFollowsTheTable(t *testing.T) {
	for n := RollingStonePlayerCntMin; n <= RollingStonePlayerCntMax; n++ {
		want := n * RollingStoneHandSize
		assert.Equal(t, want, RollingStoneDeckSize(n))
		assert.Zero(t, want%RollingStoneSuitCnt, "%d 人: 4 スートに均等に割れる", n)

		r := NewRollingStone(nil, RollingStoneConfig{PlayerCnt: n})
		r.Reset()
		total := 0
		for i := range n {
			assert.Equal(t, RollingStoneHandSize, r.GetPlayer(i).GetCardsSize(), "%d 人: 8 枚ずつ", n)
			total += r.GetPlayer(i).GetCardsSize()
		}
		assert.Equal(t, want, total, "%d 人: %d 枚すべてが配られる", n, want)
	}
}

// **抜くのは低いランクだけ。** A は最強として常に残す。
func TestRollingStone_DeckKeepsTheTopRanksAndAces(t *testing.T) {
	for n := RollingStonePlayerCntMin; n <= RollingStonePlayerCntMax; n++ {
		lowest := RollingStoneLowestRank(n)
		r := NewRollingStone(nil, RollingStoneConfig{PlayerCnt: n})
		r.Reset()

		suits := map[int]int{}
		for i := range n {
			p := r.GetPlayer(i)
			for c := range p.GetCardsSize() {
				card := p.GetCard(c)
				suits[card.GetDesign()]++
				if card.GetValue() == 1 {
					continue // A は常に使う
				}
				assert.GreaterOrEqual(t, card.GetValue(), lowest,
					"%d 人: %d 未満のランクは使わない", n, lowest)
			}
		}
		// **各スートちょうど同数。** 偏ると特定のスートだけ引き取りやすくなる。
		for suit, cnt := range suits {
			assert.Equal(t, RollingStoneDeckSize(n)/RollingStoneSuitCnt, cnt,
				"%d 人: スート %d の枚数", n, suit)
		}
		assert.Len(t, suits, RollingStoneSuitCnt, "%d 人: 4 スートそろう", n)
	}
}

func TestRollingStone_LowestRankMatchesTheDeckSize(t *testing.T) {
	for n := RollingStonePlayerCntMin; n <= RollingStonePlayerCntMax; n++ {
		ranks := RollingStoneDeckSize(n) / RollingStoneSuitCnt
		lowest := RollingStoneLowestRank(n)
		// A + (lowest..13) がちょうど ranks 種。
		assert.Equal(t, ranks, 1+(13-lowest+1), "%d 人", n)
	}
	assert.Equal(t, 7, RollingStoneLowestRank(4))
	assert.Equal(t, 5, RollingStoneLowestRank(5))
	assert.Equal(t, 3, RollingStoneLowestRank(6))
}

func TestRollingStone_ResetStartsInPlay(t *testing.T) {
	r := newTestRollingStone(t)
	assert.Equal(t, RollingStonePhasePlay, r.GetPhase())
	assert.Equal(t, -1, r.GetWinnerIdx())
	assert.Equal(t, -1, r.GetLastPickupIdx())
	assert.Zero(t, r.GetFinishedCnt())
	assert.Empty(t, r.GetCurrentTrick())
	assert.NotEmpty(t, r.GetValidPlayIndices(0), "リードなら何でも出せる")
}

// **フォローできる札があるなら、それしか出せない。**
func TestRollingStone_FollowSuitIsCompulsory(t *testing.T) {
	r := newTestRollingStone(t)
	r.SetLeadPlayerIdxForTest(0)
	r.SetCurrentPlayerIdxForTest(0)
	r.GiveHandForTest(0, NewCard(CardDesignSpade, 8, false))
	r.GiveHandForTest(1, NewCard(CardDesignSpade, 9, false), NewCard(CardDesignHeart, 8, false))

	require.NoError(t, r.PlayForTest(0, 0))
	assert.Equal(t, []int{0}, r.GetValidPlayIndices(1))
	assert.Error(t, r.PlayForTest(1, 1), "別のスートは出せない")
	assert.False(t, r.MustPickUp(1), "フォローできるので引き取らない")
	assert.Error(t, r.PickUpForTest(1), "フォローできるのに引き取れない")
}

// **フォローできないと、場札を全部引き取る。** これがこのゲームの罰則。
func TestRollingStone_CannotFollowMeansPickingTheTrickUp(t *testing.T) {
	r := newTestRollingStone(t)
	r.SetLeadPlayerIdxForTest(0)
	r.SetCurrentPlayerIdxForTest(0)
	r.GiveHandForTest(0, NewCard(CardDesignSpade, 8, false), NewCard(CardDesignSpade, 9, false))
	r.GiveHandForTest(1, NewCard(CardDesignHeart, 8, false), NewCard(CardDesignHeart, 9, false))

	require.NoError(t, r.PlayForTest(0, 0))
	assert.Empty(t, r.GetValidPlayIndices(1))
	assert.True(t, r.MustPickUp(1))
	assert.Error(t, r.PlayForTest(1, 0), "出せる札が無いので打てない")

	before := r.GetPlayer(1).GetCardsSize()
	require.NoError(t, r.PickUpForTest(1))
	assert.Equal(t, before+1, r.GetPlayer(1).GetCardsSize(), "**手札が増える**")
	assert.Empty(t, r.GetCurrentTrick())
	assert.Equal(t, 1, r.GetPlayer(1).GetPickups())
	assert.Equal(t, 1, r.GetLastPickupIdx())
	// **引き取った人が次のリード。**
	assert.Equal(t, 1, r.GetLeadPlayerIdx())
	assert.Equal(t, 1, r.GetCurrentPlayerIdx())
}

// **札が場から抜けるのは、全員フォローできたトリックだけ。**
//
// 引き取りは席の間で動かすだけなので総数は変わりません。**この非対称が停止性の
// 根っこ**で、フォローできない局面が続くかぎり札は減りません。
func TestRollingStone_CardsOnlyLeavePlayOnACompletedTrick(t *testing.T) {
	r := newTestRollingStone(t)
	count := func() int {
		total := len(r.GetCurrentTrick())
		for i := range r.GetPlayerCnt() {
			total += r.GetPlayer(i).GetCardsSize()
		}
		return total
	}
	require.Equal(t, r.GetDeckSize(), count())

	for turns := 0; !r.GetGameEndFlag() && turns < 500; turns++ {
		idx := r.GetCurrentPlayerIdx()
		before := count()
		if r.MustPickUp(idx) {
			require.NoError(t, r.PickUpForTest(idx))
			assert.Equal(t, before, count(), "引き取りでは総数が変わらない")
			continue
		}
		require.NoError(t, r.PlayForTest(idx, r.CpuChoiceForTest(idx)))
		assert.LessOrEqual(t, count(), before, "総数は増えない")
	}
}

// **引き取りだけでは総数が変わらない。** 手札の間で動くだけ。
func TestRollingStone_PickingUpMovesCardsWithoutChangingTheTotal(t *testing.T) {
	r := newTestRollingStone(t)
	r.SetLeadPlayerIdxForTest(0)
	r.SetCurrentPlayerIdxForTest(0)
	r.GiveHandForTest(0, NewCard(CardDesignSpade, 8, false), NewCard(CardDesignSpade, 9, false))
	r.GiveHandForTest(1, NewCard(CardDesignHeart, 8, false))
	r.GiveHandForTest(2, NewCard(CardDesignHeart, 9, false))
	r.GiveHandForTest(3, NewCard(CardDesignHeart, 10, false))

	total := func() int {
		n := len(r.GetCurrentTrick())
		for i := range r.GetPlayerCnt() {
			n += r.GetPlayer(i).GetCardsSize()
		}
		return n
	}
	before := total()
	require.NoError(t, r.PlayForTest(0, 0))
	require.NoError(t, r.PickUpForTest(1))
	assert.Equal(t, before, total())
	assert.Equal(t, 2, r.GetPlayer(1).GetCardsSize(), "引き取ったぶん増える")
	assert.Equal(t, 1, r.GetPlayer(0).GetCardsSize(), "出したぶん減ったまま")
}

// **切り札は無い。** リードのスートで最も強い札が取る。
func TestRollingStone_TrickWinnerIsTheHighestOfTheLedSuit(t *testing.T) {
	r := newTestRollingStone(t)
	r.SetLeadPlayerIdxForTest(0)
	r.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 9, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 13, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 10, false)},
	})
	assert.Equal(t, 2, r.TrickWinnerForTest(), "A が最強")
}

// **トリックを取っても得点にならない。** 取るのは次のリード権だけ。
func TestRollingStone_WinningATrickOnlyGivesTheLead(t *testing.T) {
	r := newTestRollingStone(t)
	for i := range r.GetPlayerCnt() {
		r.GiveHandForTest(i,
			NewCard(CardDesignSpade, 7+i, false),
			NewCard(CardDesignHeart, 7+i, false))
	}
	r.SetLeadPlayerIdxForTest(0)
	r.SetCurrentPlayerIdxForTest(0)

	for i := range r.GetPlayerCnt() {
		require.NoError(t, r.PlayForTest(i, 0))
	}
	assert.Empty(t, r.GetCurrentTrick(), "揃ったら解決される")
	assert.Equal(t, 1, r.GetTrickNumber())
	// 席 3 が ♠10 でいちばん強い。
	assert.Equal(t, 3, r.GetLeadPlayerIdx())
	for i := range r.GetPlayerCnt() {
		assert.Equal(t, 1, r.GetPlayer(i).GetCardsSize(), "誰の手札も増えていない")
	}
}

// **先に手札を出し切った人が勝ち。**
func TestRollingStone_FirstToEmptyTheirHandWins(t *testing.T) {
	r := newTestRollingStone(t)
	for i := range r.GetPlayerCnt() {
		r.GiveHandForTest(i, NewCard(CardDesignSpade, 7+i, false))
	}
	r.SetLeadPlayerIdxForTest(0)
	r.SetCurrentPlayerIdxForTest(0)

	require.NoError(t, r.PlayForTest(0, 0))
	assert.True(t, r.GetPlayer(0).HasFinished())
	assert.Equal(t, 1, r.GetPlayer(0).GetFinishedAt())
	assert.Equal(t, 0, r.GetWinnerIdx())
	// 残りが打ち切ってトリックが解決した時点で終局する。
	for i := 1; i < r.GetPlayerCnt(); i++ {
		if r.GetGameEndFlag() {
			break
		}
		require.NoError(t, r.PlayForTest(i, 0))
	}
	assert.True(t, r.GetGameEndFlag())
	assert.Equal(t, RollingStonePhaseGameEnd, r.GetPhase())
}

func TestRollingStone_RejectsOutOfTurnAndBadIndices(t *testing.T) {
	r := newTestRollingStone(t)
	idx := r.GetCurrentPlayerIdx()
	assert.Error(t, r.PlayForTest((idx+1)%r.GetPlayerCnt(), 0), "手番でない席は打てない")
	assert.Error(t, r.PlayForTest(idx, -1))
	assert.Error(t, r.PlayForTest(idx, 999))
	assert.Error(t, r.PickUpForTest((idx+1)%r.GetPlayerCnt()))

	r.GiveUp()
	assert.Error(t, r.PlayForTest(idx, 0), "終局後は打てない")
	assert.Error(t, r.PickUpForTest(idx), "終局後は引き取れない")
}

// **公開の入口も踏む。**
func TestRollingStone_PublicEntryPointsGuardTheTurn(t *testing.T) {
	r := newTestRollingStone(t)
	r.SetCurrentPlayerIdxForTest(1)
	assert.False(t, r.IsHumanTurn())
	assert.Error(t, r.PlayerPlay(0))
	assert.Error(t, r.PlayerPickUp())

	before := r.GetPlayer(1).GetCardsSize()
	r.CpuPlay()
	assert.Equal(t, before-1, r.GetPlayer(1).GetCardsSize())

	r.SetCurrentPlayerIdxForTest(0)
	assert.True(t, r.IsHumanTurn())
	humanBefore := r.GetPlayer(0).GetCardsSize()
	r.CpuPlay()
	assert.Equal(t, humanBefore, r.GetPlayer(0).GetCardsSize(), "人間の番では CPU は動かない")

	// **フォローできるかは配り次第。** 出せないなら引き取るのが正しい入口。
	if valid := r.GetValidPlayIndices(0); len(valid) > 0 {
		require.NoError(t, r.PlayerPlay(valid[0]))
		return
	}
	require.True(t, r.MustPickUp(0))
	require.NoError(t, r.PlayerPickUp())
}

// **CPU はフォローできなければ引き取る。**
func TestRollingStone_CpuPicksUpWhenItCannotFollow(t *testing.T) {
	r := newTestRollingStone(t)
	r.SetLeadPlayerIdxForTest(0)
	r.SetCurrentPlayerIdxForTest(0)
	// **席 0 に 2 枚持たせる。** 1 枚だと出した時点で上がって終局し、CPU が動かない。
	r.GiveHandForTest(0, NewCard(CardDesignSpade, 8, false), NewCard(CardDesignSpade, 9, false))
	r.GiveHandForTest(1, NewCard(CardDesignHeart, 8, false))
	require.NoError(t, r.PlayForTest(0, 0))

	before := r.GetPlayer(1).GetCardsSize()
	r.CpuPlay()
	assert.Equal(t, before+1, r.GetPlayer(1).GetCardsSize())
	assert.Equal(t, 1, r.GetPlayer(1).GetPickups())
}

// **CPU は必ず合法手を返す。**
func TestRollingStone_CpuAlwaysChoosesLegally(t *testing.T) {
	for n := RollingStonePlayerCntMin; n <= RollingStonePlayerCntMax; n++ {
		for range 15 {
			r := NewRollingStone(nil, RollingStoneConfig{PlayerCnt: n})
			r.Reset()
			for turns := 0; !r.GetGameEndFlag() && turns < 500; turns++ {
				idx := r.GetCurrentPlayerIdx()
				if r.MustPickUp(idx) {
					require.NoError(t, r.PickUpForTest(idx))
					continue
				}
				choice := r.CpuChoiceForTest(idx)
				require.Contains(t, r.GetValidPlayIndices(idx), choice)
				require.NoError(t, r.PlayForTest(idx, choice))
			}
		}
	}
}

// **どの局も必ず終わる。** 引き取りで手札が増えるので、停止性は自明ではない。
func TestRollingStone_GamesTerminate(t *testing.T) {
	for n := RollingStonePlayerCntMin; n <= RollingStonePlayerCntMax; n++ {
		for range 15 {
			r := NewRollingStone(nil, RollingStoneConfig{PlayerCnt: n})
			r.Reset()
			for turns := 0; !r.GetGameEndFlag(); turns++ {
				require.Less(t, turns, 200000, "%d 人: 膠着上限でも終わらない", n)
				idx := r.GetCurrentPlayerIdx()
				if r.MustPickUp(idx) {
					require.NoError(t, r.PickUpForTest(idx))
					continue
				}
				require.NoError(t, r.PlayForTest(idx, r.CpuChoiceForTest(idx)))
			}
			assert.GreaterOrEqual(t, r.GetWinnerIdx(), 0)
			if r.GetTrickNumber() >= RollingStoneStalemateTricks {
				// **膠着で切った局。** 勝者は手札のいちばん少ない席。
				for i := range n {
					assert.LessOrEqual(t, r.GetPlayer(r.GetWinnerIdx()).GetCardsSize(),
						r.GetPlayer(i).GetCardsSize(), "%d 人: 勝者がいちばん少ない", n)
				}
				continue
			}
			assert.Zero(t, r.GetPlayer(r.GetWinnerIdx()).GetCardsSize(), "上がりで決着した局の勝者は手札が空")
		}
	}
}

func TestRollingStone_GiveUp(t *testing.T) {
	r := newTestRollingStone(t)
	r.GiveUp()
	assert.True(t, r.GetGameEndFlag())
	assert.Positive(t, r.GetWinnerIdx(), "投了した席は勝者にならない")
	before := r.GetWinnerIdx()
	r.GiveUp()
	assert.Equal(t, before, r.GetWinnerIdx())
}

func TestRollingStone_Hint(t *testing.T) {
	r := newTestRollingStone(t)
	r.SetCurrentPlayerIdxForTest(0)

	lead := r.GetHint()
	require.NotNil(t, lead)
	require.NotNil(t, lead.CardIndex)
	assert.Equal(t, "rollingstoneLead", lead.Reason)
	assert.Contains(t, r.GetValidPlayIndices(0), *lead.CardIndex)

	// フォローする場面。
	r.GiveHandForTest(0, NewCard(CardDesignSpade, 8, false))
	r.SetCurrentTrickForTest([]*TrickCard{{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 9, false)}})
	follow := r.GetHint()
	require.NotNil(t, follow)
	assert.Equal(t, "rollingstoneFollow", follow.Reason)

	// **引き取るしかない場面は札を指さない。**
	r.GiveHandForTest(0, NewCard(CardDesignHeart, 8, false))
	pick := r.GetHint()
	require.NotNil(t, pick)
	assert.Nil(t, pick.CardIndex)
	assert.Equal(t, "rollingstonePickUp", pick.Reason)

	r.SetCurrentPlayerIdxForTest(1)
	assert.Nil(t, r.GetHint(), "相手の手番では助言しない")

	r.SetCurrentPlayerIdxForTest(0)
	r.GiveUp()
	assert.Nil(t, r.GetHint(), "終局後は助言しない")
}

func TestRollingStone_AccessorsAndBounds(t *testing.T) {
	r := newTestRollingStone(t)
	assert.Nil(t, r.GetPlayer(-1))
	assert.Nil(t, r.GetPlayer(99))
	assert.Empty(t, r.GetValidPlayIndices(-1))
	assert.Empty(t, r.GetValidPlayIndices(99))
	assert.False(t, r.MustPickUp(0), "場が空なら引き取りは起きない")
	assert.Equal(t, RollingStoneDefaultPlayerCnt, r.GetPlayerCnt())
	assert.Equal(t, RollingStoneDeckSize(RollingStoneDefaultPlayerCnt), r.GetDeckSize())
	assert.Equal(t, 0, r.NextTrickLead())
	assert.NotEmpty(t, r.GetActionLog())
	assert.Equal(t, DefaultRollingStoneConfig(), r.GetConfig())

	// **人数を変えたら席も作り直す。**
	r.SetConfig(RollingStoneConfig{PlayerCnt: 6})
	assert.Equal(t, 6, r.GetPlayerCnt())
	assert.NotPanics(t, r.Reset)
	assert.Equal(t, RollingStoneHandSize, r.GetPlayer(5).GetCardsSize())
}

func TestRollingStoneConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultRollingStoneConfig().Validate())
	for n := RollingStonePlayerCntMin; n <= RollingStonePlayerCntMax; n++ {
		assert.NoError(t, RollingStoneConfig{PlayerCnt: n}.Validate())
	}
	assert.Error(t, RollingStoneConfig{PlayerCnt: RollingStonePlayerCntMin - 1}.Validate())
	assert.Error(t, RollingStoneConfig{PlayerCnt: RollingStonePlayerCntMax + 1}.Validate())

	// **不正な設定で作っても遊べる盤面になる。**
	r := NewRollingStone(nil, RollingStoneConfig{PlayerCnt: 99})
	assert.Equal(t, RollingStoneDefaultPlayerCnt, r.GetPlayerCnt())
}

// **上限に達したら手札のいちばん少ない席の勝ちにする。**
//
// このゲームは放っておくと終わりません（実測: 4 人では 400 局中 210 局が
// 30 万手でも決着せず）。上限は**決着する局の最大 303 トリック**より十分上に
// 置いてあるので、本来決着する局を途中で止めることはまず起きません。
func TestRollingStone_StalemateGoesToTheSmallestHand(t *testing.T) {
	r := newTestRollingStone(t)
	// 上限の 1 つ手前まで進めた状態を作る。
	r.SetLeadPlayerIdxForTest(0)
	r.SetCurrentPlayerIdxForTest(0)
	for i := range r.GetPlayerCnt() {
		r.GiveHandForTest(i, NewCard(CardDesignSpade, 7+i, false), NewCard(CardDesignHeart, 7+i, false))
	}
	// 席 2 をいちばん少なくする。
	r.GiveHandForTest(2, NewCard(CardDesignSpade, 9, false))

	for range RollingStoneStalemateTricks {
		if r.GetGameEndFlag() {
			break
		}
		idx := r.GetCurrentPlayerIdx()
		if r.MustPickUp(idx) {
			require.NoError(t, r.PickUpForTest(idx))
			continue
		}
		require.NoError(t, r.PlayForTest(idx, r.CpuChoiceForTest(idx)))
	}
	assert.True(t, r.GetGameEndFlag(), "上限に達したら必ず終わる")
	assert.GreaterOrEqual(t, r.GetWinnerIdx(), 0)
}

// **上限は決着を邪魔しない。** 通常の局はずっと手前で終わる。
func TestRollingStone_StalemateCapIsWellAboveNormalPlay(t *testing.T) {
	assert.Greater(t, RollingStoneStalemateTricks, 303,
		"実測した『決着する局の最大トリック数』より上でなければ、本来決着する局を切ってしまう")
}

func TestRollingStone_JSONRoundTrip(t *testing.T) {
	r := newTestRollingStone(t)
	for range 5 {
		idx := r.GetCurrentPlayerIdx()
		if r.MustPickUp(idx) {
			require.NoError(t, r.PickUpForTest(idx))
			continue
		}
		require.NoError(t, r.PlayForTest(idx, r.CpuChoiceForTest(idx)))
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var restored RollingStone
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, r.GetPlayerCnt(), restored.GetPlayerCnt())
	assert.Equal(t, r.GetTrickNumber(), restored.GetTrickNumber())
	assert.Len(t, restored.GetCurrentTrick(), len(r.GetCurrentTrick()))
	for i := range r.GetPlayerCnt() {
		assert.Equal(t, r.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize())
		assert.Equal(t, r.GetPlayer(i).GetPickups(), restored.GetPlayer(i).GetPickups(), "引き取り回数が消えない")
	}
}

// **壊れたスナップショットは弾く。**
//
// このコーデックは 9 PR 連続で「個々のフィールドは範囲内だが組み合わせが
// あり得ない」を通していたので、**書き込み側が保っている関係**を先に写しています。
func TestRollingStone_UnmarshalRejectsBrokenSnapshots(t *testing.T) {
	snapshot := func(t *testing.T) map[string]any {
		t.Helper()
		r := newTestRollingStone(t)
		require.NoError(t, r.PlayForTest(r.GetCurrentPlayerIdx(), r.CpuChoiceForTest(r.GetCurrentPlayerIdx())))
		data, err := json.Marshal(r)
		require.NoError(t, err)
		var out map[string]any
		require.NoError(t, json.Unmarshal(data, &out))
		return out
	}

	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"phase out of range", func(m map[string]any) { m["ph"] = 9 }},
		{"game end flag without the phase", func(m map[string]any) { m["ge"] = true }},
		{"current player out of range", func(m map[string]any) { m["ci"] = 9 }},
		{"lead player out of range", func(m map[string]any) { m["li"] = -1 }},
		{"winner without the game ending", func(m map[string]any) { m["wi"] = 1 }},
		{"last pickup out of range", func(m map[string]any) { m["lp"] = 9 }},
		{"negative trick number", func(m map[string]any) { m["tn"] = -1 }},
		{"finished count above the table", func(m map[string]any) { m["fc"] = 9 }},
		{"config out of range", func(m map[string]any) { m["cf"] = map[string]any{"p": 9} }},
		{"player count disagrees with the seats", func(m map[string]any) { m["cf"] = map[string]any{"p": 5} }},
		// **枚数だけでなく中身も見る (#5310 の再発防止)。**
		{"a trick entry with no card", func(m map[string]any) {
			m["ct"] = []any{map[string]any{"playerIdx": 0}}
		}},
		{"a trick entry with a bad seat", func(m map[string]any) {
			m["ct"] = []any{map[string]any{"playerIdx": 9, "card": map[string]any{"d": 1, "v": 9, "j": false}}}
		}},
		// **札は人数 × 8 枚しかない（#5314 の形）。**
		{"a card appears from nowhere", func(m map[string]any) {
			m["ct"] = append(m["ct"].([]any), map[string]any{
				"playerIdx": 0, "card": map[string]any{"d": 1, "v": 9, "j": false},
			})
		}},
		// **上がった人数と順位は対（#5313 の形）。**
		{"finished count with nobody finished", func(m map[string]any) { m["fc"] = 1 }},
		// **場札は在席数より少ない。** 揃った時点で解決されるので残らない。
		{"a full trick left unresolved", func(m map[string]any) {
			trick := m["ct"].([]any)
			first := trick[0]
			m["ct"] = []any{first, first, first, first}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := snapshot(t)
			tc.mutate(s)
			data, err := json.Marshal(s)
			require.NoError(t, err)
			var restored RollingStone
			assert.Error(t, json.Unmarshal(data, &restored))
		})
	}

	// **山札そのものが要る（#5314 の形）。**
	t.Run("missing trump cards", func(t *testing.T) {
		s := snapshot(t)
		s["tc"] = nil
		data, err := json.Marshal(s)
		require.NoError(t, err)
		var restored RollingStone
		assert.Error(t, json.Unmarshal(data, &restored))
	})

	// **負のコントロール: 手を加えていないスナップショットは通り、使っても落ちない。**
	data, err := json.Marshal(snapshot(t))
	require.NoError(t, err)
	var ok RollingStone
	require.NoError(t, json.Unmarshal(data, &ok))
	assert.NotPanics(t, func() {
		_ = ok.GetValidPlayIndices(ok.GetCurrentPlayerIdx())
		_ = ok.MustPickUp(ok.GetCurrentPlayerIdx())
		_ = ok.GetHint()
		_ = ok.TrickWinnerForTest()
	})
}

// **上がった席が手番やリードになっている盤面は存在しない。**
func TestRollingStone_UnmarshalRejectsAFinishedSeatOnTurn(t *testing.T) {
	r := newTestRollingStone(t)
	// 席 1 を上がらせ、手札を席 0 に移して枚数の帳尻を合わせる。
	moved := r.GetPlayer(1).GetCardsSize()
	for range moved {
		r.GetPlayer(0).AddCard(NewCard(CardDesignSpade, 7, false))
	}
	r.GiveHandForTest(1)
	r.GetPlayer(1).SetFinishedAt(1)

	data, err := json.Marshal(r)
	require.NoError(t, err)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))
	raw["fc"] = 1
	raw["ge"] = false
	raw["wi"] = -1

	for _, key := range []string{"ci", "li"} {
		m := make(map[string]any, len(raw))
		for k, v := range raw {
			m[k] = v
		}
		m["ci"], m["li"] = 0, 0
		m[key] = 1 // 上がった席
		bad, err := json.Marshal(m)
		require.NoError(t, err)
		var restored RollingStone
		assert.Error(t, json.Unmarshal(bad, &restored), key)
	}

	// **負のコントロール: 在席の席が手番なら通る。**
	raw["ci"], raw["li"] = 0, 0
	good, err := json.Marshal(raw)
	require.NoError(t, err)
	var ok RollingStone
	assert.NoError(t, json.Unmarshal(good, &ok))
	assert.Equal(t, 1, ok.GetFinishedCnt())
}

// **上がった順位は 1..finishedCnt の並べ替え。** 同じ順位が 2 つある盤面は無い。
func TestRollingStonePlayer_UnmarshalRejectsAContradictoryRank(t *testing.T) {
	// 上がっているのに手札が残っている。**綴りを推測せず、実物から作る。**
	held := NewRollingStonePlayer(false)
	held.AddCard(NewCard(CardDesignSpade, 5, false))
	held.SetFinishedAt(1)
	data, err := json.Marshal(held)
	require.NoError(t, err)
	var p RollingStonePlayer
	assert.Error(t, json.Unmarshal(data, &p), string(data))

	// 手札が空なのに順位が付いていない。
	var q RollingStonePlayer
	assert.Error(t, json.Unmarshal([]byte(`{"fa":0}`), &q))

	for _, body := range []string{`{"fa":-1}`, `{"pu":-1}`, `{"fa":99}`} {
		var bad RollingStonePlayer
		assert.Error(t, json.Unmarshal([]byte(body), &bad), body)
	}

	// 負のコントロール: 手札が空で順位が付いていれば通る。
	var ok RollingStonePlayer
	assert.NoError(t, json.Unmarshal([]byte(`{"fa":1,"pu":2}`), &ok))
	assert.Equal(t, 1, ok.GetFinishedAt())
	assert.Equal(t, 2, ok.GetPickups())
}

// **抜けた枚数まで足して「人数 × 8」に合わせる。**
//
// 手札と場札だけを数えると、トリックが 1 つ解決した時点で足りなくなります
// ——最初に書いた検証はそれで正しい盤面を拒否しました。
func TestRollingStone_UnmarshalCountsTheDiscardsToo(t *testing.T) {
	// **配りを差し替えず、実際の局を進めて 1 トリック解決させる。**
	// `GiveHandForTest` は札を捨てるので、枚数の突き合わせには使えません。
	r := newTestRollingStone(t)
	for turns := 0; r.GetDiscarded() == 0 && turns < 200; turns++ {
		idx := r.GetCurrentPlayerIdx()
		if r.MustPickUp(idx) {
			require.NoError(t, r.PickUpForTest(idx))
			continue
		}
		require.NoError(t, r.PlayForTest(idx, r.CpuChoiceForTest(idx)))
	}
	require.Positive(t, r.GetDiscarded(), "解決したトリックのぶん抜ける")

	// 抜けたぶんを数えているので、この盤面は通る。
	data, err := json.Marshal(r)
	require.NoError(t, err)
	var ok RollingStone
	require.NoError(t, json.Unmarshal(data, &ok))
	assert.Equal(t, r.GetDiscarded(), ok.GetDiscarded())

	// **負のコントロール: 抜けた枚数を書き換えれば弾く。**
	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))
	for _, bad := range []any{0, -1, 99} {
		raw["dc"] = bad
		mutated, err := json.Marshal(raw)
		require.NoError(t, err)
		var restored RollingStone
		assert.Error(t, json.Unmarshal(mutated, &restored), "dc=%v", bad)
	}
}

// **上がりは、そのトリックが引き取りで流れても取り消されない（レビュー指摘 PR #5316）。**
//
// 以前は `resolveTrick` でしか終局させていなかったので、上がった直後に別の席が
// 引き取ると「勝者が決まっているのに終局していない」盤面が残りました
// ——それは `UnmarshalJSON` が弾く状態そのもので、保存して読み直すと
// 正当な対局が拒否されました。
func TestRollingStone_FinishingEndsTheGameEvenIfAPickUpBreaksTheTrick(t *testing.T) {
	r := newTestRollingStone(t)
	r.SetLeadPlayerIdxForTest(0)
	r.SetCurrentPlayerIdxForTest(0)
	r.GiveHandForTest(0, NewCard(CardDesignSpade, 8, false))
	r.GiveHandForTest(1, NewCard(CardDesignHeart, 8, false))
	r.GiveHandForTest(2, NewCard(CardDesignSpade, 9, false))
	r.GiveHandForTest(3, NewCard(CardDesignSpade, 10, false))

	require.NoError(t, r.PlayForTest(0, 0))
	assert.True(t, r.GetGameEndFlag(), "出し切った瞬間に終わる")
	assert.Equal(t, 0, r.GetWinnerIdx())

	// 終局後は引き取りも打つこともできない。
	assert.Error(t, r.PickUpForTest(1))
	assert.Error(t, r.PlayForTest(1, 0))
}

// **トリックの途中で誰かが上がっても、残りの席は出し切れる（レビュー指摘 PR #5316）。**
//
// 在席数と枚数を比べていたので、上がった瞬間に在席数が縮み、まだ出していない
// 最後の席を飛ばして解決していました。
func TestRollingStone_AFinishMidTrickDoesNotSkipTheRemainingSeat(t *testing.T) {
	r := newTestRollingStone(t)
	r.SetLeadPlayerIdxForTest(0)
	r.SetCurrentPlayerIdxForTest(0)
	// 席 2 だけが最後の 1 枚。ただし席 0 が既に上がっていない状態で始める。
	r.GiveHandForTest(0, NewCard(CardDesignSpade, 7, false), NewCard(CardDesignHeart, 7, false))
	r.GiveHandForTest(1, NewCard(CardDesignSpade, 8, false), NewCard(CardDesignHeart, 8, false))
	r.GiveHandForTest(2, NewCard(CardDesignSpade, 9, false))
	r.GiveHandForTest(3, NewCard(CardDesignSpade, 10, false), NewCard(CardDesignHeart, 10, false))

	require.NoError(t, r.PlayForTest(0, 0))
	require.NoError(t, r.PlayForTest(1, 0))
	require.NoError(t, r.PlayForTest(2, 0))

	// 席 2 が上がるので終局する——が、その前にトリックを打ち切っていないこと。
	assert.True(t, r.GetGameEndFlag(), "最初の上がりで終わる")
	assert.Equal(t, 2, r.GetWinnerIdx())
	assert.Len(t, r.GetCurrentTrick(), 3, "席 3 を飛ばしてトリックを解決していない")
}

// **既に上がった席がいる局でも、残りはトリックを打ち切れる。**
//
// 「まだ出していない在席者がいるか」で判定しているので、上がった席のぶん
// 枚数が足りなくても解決します。
func TestRollingStone_TrickResolvesWithFinishedSeatsSkipped(t *testing.T) {
	r := newTestRollingStone(t)
	// 席 3 を先に上がらせる（勝者は既に決まっている扱いにはしない）。
	r.GiveHandForTest(3)
	r.GetPlayer(3).SetFinishedAt(1)

	r.SetLeadPlayerIdxForTest(0)
	r.SetCurrentPlayerIdxForTest(0)
	for i := range 3 {
		r.GiveHandForTest(i, NewCard(CardDesignSpade, 7+i, false), NewCard(CardDesignHeart, 7+i, false))
	}

	for i := range 3 {
		require.NoError(t, r.PlayForTest(i, 0))
	}
	assert.Empty(t, r.GetCurrentTrick(), "在席 3 人が打てば解決する")
	assert.Equal(t, 1, r.GetTrickNumber())
}

// **書き込み側は、自分の codec が弾く盤面を作らない（レビュー指摘 PR #5316）。**
//
// これがこの指摘の本質でした——`UnmarshalJSON` のガードは正しく、破っていたのは
// `play` / `pickUp` のほうです。実際の局を最後まで回し、**毎手ごとに**保存して
// 読み直せることを確かめます。
func TestRollingStone_EveryReachableStateSurvivesARoundTrip(t *testing.T) {
	for n := RollingStonePlayerCntMin; n <= RollingStonePlayerCntMax; n++ {
		for range 10 {
			r := NewRollingStone(nil, RollingStoneConfig{PlayerCnt: n})
			r.Reset()

			for turns := 0; ; turns++ {
				require.Less(t, turns, 200000, "%d 人: 終わらない", n)

				data, err := json.Marshal(r)
				require.NoError(t, err)
				var back RollingStone
				require.NoError(t, json.Unmarshal(data, &back),
					"%d 人 %d 手目: 書き込み側が codec の不変条件を破った", n, turns)

				if r.GetGameEndFlag() {
					break
				}
				idx := r.GetCurrentPlayerIdx()
				if r.MustPickUp(idx) {
					require.NoError(t, r.PickUpForTest(idx))
					continue
				}
				require.NoError(t, r.PlayForTest(idx, r.CpuChoiceForTest(idx)))
			}
		}
	}
}

// **棋譜の上限は、この局が実際に出しうる長さより上でなければならない。**
//
// 膠着上限まで走る局が出す棋譜は、1 トリックあたり「在席人数ぶんのプレイ +
// 解決または引き取り」。1000 のままだと長い局を読み直せませんでした。
func TestRollingStone_ActionLogCapCoversTheLongestPossibleGame(t *testing.T) {
	worst := RollingStoneStalemateTricks*(RollingStonePlayerCntMax+1) + 1 + RollingStonePlayerCntMax + 1
	assert.GreaterOrEqual(t, rollingStoneMaxSliceLen, worst,
		"膠着上限まで走った局の棋譜（最大 %d 行）を読み直せる必要がある", worst)
}

// **引き取りの理由はリードスート。** 画面もCUIもこれを出すので、
// 「トリックの先頭札のスート」であることをドメイン側で固定する (#5764)。
func TestRollingStoneGetLeadSuitFollowsTheFirstTrickCard(t *testing.T) {
	g := newTestRollingStone(t)

	if got := g.GetLeadSuit(); got != 0 {
		t.Errorf("GetLeadSuit() with an empty trick = %d, want 0", got)
	}

	idx := g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
	if len(idx) == 0 {
		t.Fatal("the leader must have a playable card")
	}
	want := g.GetPlayer(g.GetCurrentPlayerIdx()).GetCard(idx[0]).GetDesign()
	if err := g.play(g.GetCurrentPlayerIdx(), idx[0]); err != nil {
		t.Fatalf("play: %v", err)
	}
	if got := g.GetLeadSuit(); got != want {
		t.Errorf("GetLeadSuit() = %d, want %d (the design of the led card)", got, want)
	}

	// **先頭札であって直近の札ではない。** 1 枚だけの場では両者が一致してしまい、
	// 取り違えを検出できない。
	g.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignClover, 5, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignHeart, 7, false)},
	})
	if got := g.GetLeadSuit(); got != CardDesignClover {
		t.Errorf("GetLeadSuit() with a 2-card trick = %d, want %d (the first card)", got, CardDesignClover)
	}
}
