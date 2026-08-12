//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLingerLonger(t *testing.T) *LingerLonger {
	t.Helper()
	l := NewDefaultLingerLonger()
	l.Reset()
	return l
}

// **配る枚数は人数と同じ。** 4 人なら 4 枚ずつ、6 人なら 6 枚ずつ。
func TestLingerLonger_DealsAsManyCardsAsThereArePlayers(t *testing.T) {
	for n := LingerLongerPlayerCntMin; n <= LingerLongerPlayerCntMax; n++ {
		l := NewLingerLonger(nil, LingerLongerConfig{PlayerCnt: n})
		l.Reset()

		total := l.GetStockSize()
		for i := range n {
			assert.Equal(t, n, l.GetPlayer(i).GetCardsSize(), "%d 人なら %d 枚ずつ", n, n)
			total += l.GetPlayer(i).GetCardsSize()
		}
		assert.Equal(t, LingerLongerDeckSize, total, "%d 人でも 52 枚すべてに行き先がある", n)
		assert.Equal(t, LingerLongerDeckSize-n*n, l.GetStockSize(), "残りは山札")
	}
}

func TestLingerLonger_ResetStartsInPlay(t *testing.T) {
	l := newTestLingerLonger(t)
	assert.Equal(t, LingerLongerPhasePlay, l.GetPhase())
	assert.Equal(t, -1, l.GetWinnerIdx())
	assert.Equal(t, -1, l.GetLastDrawIdx())
	assert.Zero(t, l.GetEliminatedCnt())
	assert.Empty(t, l.GetCurrentTrick())
	assert.NotEmpty(t, l.GetValidPlayIndices(0))
}

// **フォローできる札があるなら、それしか出せない。**
func TestLingerLonger_FollowSuitIsCompulsory(t *testing.T) {
	l := newTestLingerLonger(t)
	l.SetLeadPlayerIdxForTest(0)
	l.SetCurrentPlayerIdxForTest(0)
	l.GiveHandForTest(0, NewCard(CardDesignSpade, 8, false), NewCard(CardDesignSpade, 2, false))
	l.GiveHandForTest(1, NewCard(CardDesignSpade, 9, false), NewCard(CardDesignHeart, 8, false))

	require.NoError(t, l.PlayForTest(0, 0))
	assert.Equal(t, []int{0}, l.GetValidPlayIndices(1))
	assert.Error(t, l.PlayForTest(1, 1))
}

// **フォローできないときは捨て札になる。** 引き取りのような罰則はない。
func TestLingerLonger_CannotFollowMeansDiscardingAnything(t *testing.T) {
	l := newTestLingerLonger(t)
	l.SetLeadPlayerIdxForTest(0)
	l.SetCurrentPlayerIdxForTest(0)
	l.GiveHandForTest(0, NewCard(CardDesignSpade, 8, false), NewCard(CardDesignSpade, 2, false))
	l.GiveHandForTest(1, NewCard(CardDesignHeart, 8, false), NewCard(CardDesignDiamond, 9, false))

	require.NoError(t, l.PlayForTest(0, 0))
	assert.Len(t, l.GetValidPlayIndices(1), 2, "どれでも出せる")
	assert.NoError(t, l.PlayForTest(1, 1))
}

// **勝った人だけが 1 枚補充できる。** これがこのゲームの全部。
func TestLingerLonger_OnlyTheTrickWinnerDraws(t *testing.T) {
	l := newTestLingerLonger(t)
	l.SetLeadPlayerIdxForTest(0)
	l.SetCurrentPlayerIdxForTest(0)
	for i := range l.GetPlayerCnt() {
		l.GiveHandForTest(i,
			NewCard(CardDesignSpade, 7+i, false),
			NewCard(CardDesignHeart, 7+i, false))
	}
	stockBefore := l.GetStockSize()

	for i := range l.GetPlayerCnt() {
		require.NoError(t, l.PlayForTest(i, 0))
	}
	// 席 3 の ♠10 がいちばん強い。
	assert.Equal(t, 3, l.GetLastDrawIdx())
	assert.Equal(t, stockBefore-1, l.GetStockSize(), "1 枚だけ減る")
	assert.Equal(t, 2, l.GetPlayer(3).GetCardsSize(), "勝者は減らない")
	assert.Equal(t, 1, l.GetPlayer(3).GetTricksWon())
	for i := range 3 {
		assert.Equal(t, 1, l.GetPlayer(i).GetCardsSize(), "負けた席は 1 枚減る")
	}
	assert.Equal(t, 3, l.GetLeadPlayerIdx(), "勝者が次のリード")
}

// **山札が尽きたら誰も補充できない。**
func TestLingerLonger_NoDrawOnceTheStockIsEmpty(t *testing.T) {
	l := newTestLingerLonger(t)
	l.DrainStockForTest()
	l.SetLeadPlayerIdxForTest(0)
	l.SetCurrentPlayerIdxForTest(0)
	for i := range l.GetPlayerCnt() {
		l.GiveHandForTest(i,
			NewCard(CardDesignSpade, 7+i, false),
			NewCard(CardDesignHeart, 7+i, false))
	}

	for i := range l.GetPlayerCnt() {
		require.NoError(t, l.PlayForTest(i, 0))
	}
	assert.Equal(t, -1, l.GetLastDrawIdx(), "補充していない")
	for i := range l.GetPlayerCnt() {
		assert.Equal(t, 1, l.GetPlayer(i).GetCardsSize(), "全員 1 枚減る")
	}
}

// **解決したトリックは場から抜ける。** 抜けた枚数まで足して 52 になる。
//
// 取ったトリックは得点にならず手札にも戻らないので、手札・場札・山札だけを
// 数えると足りません。
func TestLingerLonger_CardsAreAccountedForIncludingTheDiscards(t *testing.T) {
	l := newTestLingerLonger(t)
	count := func() int {
		total := len(l.GetCurrentTrick()) + l.GetStockSize() + l.GetDiscarded()
		for i := range l.GetPlayerCnt() {
			total += l.GetPlayer(i).GetCardsSize()
		}
		return total
	}
	require.Equal(t, LingerLongerDeckSize, count())
	require.Zero(t, l.GetDiscarded())

	for turns := 0; !l.GetGameEndFlag() && turns < 500; turns++ {
		idx := l.GetCurrentPlayerIdx()
		require.NoError(t, l.PlayForTest(idx, l.CpuChoiceForTest(idx)))
		require.Equal(t, LingerLongerDeckSize, count(), "52 枚が保たれる")
	}
	assert.Positive(t, l.GetDiscarded(), "解決したトリックのぶん抜けている")
}

// **脱落判定は補充のあと。** 先に見ると、勝って補充する人まで落としてしまう。
func TestLingerLonger_TheWinnerIsNotEliminatedBeforeDrawing(t *testing.T) {
	l := newTestLingerLonger(t)
	l.SetLeadPlayerIdxForTest(0)
	l.SetCurrentPlayerIdxForTest(0)
	// 全員 1 枚。席 3 が勝つが、補充があるので落ちない。
	for i := range l.GetPlayerCnt() {
		l.GiveHandForTest(i, NewCard(CardDesignSpade, 7+i, false))
	}

	for i := range l.GetPlayerCnt() {
		if l.GetGameEndFlag() {
			break
		}
		require.NoError(t, l.PlayForTest(i, 0))
	}
	assert.False(t, l.GetPlayer(3).IsEliminated(), "勝者は補充で生き残る")
	assert.Equal(t, 1, l.GetPlayer(3).GetCardsSize())
	for i := range 3 {
		assert.True(t, l.GetPlayer(i).IsEliminated(), "負けた席は手札が尽きて脱落")
	}
	assert.True(t, l.GetGameEndFlag(), "残り 1 人で決着")
	assert.Equal(t, 3, l.GetWinnerIdx())
}

// **全員が同時に尽きることがある。** そのときは最後のトリックを取った人の勝ち。
func TestLingerLonger_EveryoneRunningOutAtOnceGoesToTheLastWinner(t *testing.T) {
	l := newTestLingerLonger(t)
	l.DrainStockForTest()
	l.SetLeadPlayerIdxForTest(0)
	l.SetCurrentPlayerIdxForTest(0)
	for i := range l.GetPlayerCnt() {
		l.GiveHandForTest(i, NewCard(CardDesignSpade, 7+i, false))
	}

	for i := range l.GetPlayerCnt() {
		if l.GetGameEndFlag() {
			break
		}
		require.NoError(t, l.PlayForTest(i, 0))
	}
	assert.True(t, l.GetGameEndFlag())
	assert.Equal(t, 3, l.GetWinnerIdx(), "♠10 で取った席が勝ち")
	for i := range l.GetPlayerCnt() {
		assert.True(t, l.GetPlayer(i).IsEliminated(), "全員手札が尽きている")
	}
}

// **脱落の順番は呼び出し順で変わらない。** 勝者から時計回りに見る。
func TestLingerLonger_EliminationOrderIsDeterministic(t *testing.T) {
	l := newTestLingerLonger(t)
	l.DrainStockForTest()
	l.SetLeadPlayerIdxForTest(0)
	l.SetCurrentPlayerIdxForTest(0)
	// 席 1 が勝つ配り。席 2 → 3 → 0 の順に脱落するはず。
	l.GiveHandForTest(0, NewCard(CardDesignSpade, 7, false))
	l.GiveHandForTest(1, NewCard(CardDesignSpade, 13, false))
	l.GiveHandForTest(2, NewCard(CardDesignSpade, 8, false))
	l.GiveHandForTest(3, NewCard(CardDesignSpade, 9, false))

	for i := range l.GetPlayerCnt() {
		if l.GetGameEndFlag() {
			break
		}
		require.NoError(t, l.PlayForTest(i, 0))
	}
	assert.Equal(t, 1, l.GetWinnerIdx(), "♠K で取った席 1")
	// 勝者(1) から時計回り: 1 → 2 → 3 → 0 の順に脱落する。
	assert.Equal(t, 1, l.GetPlayer(1).GetEliminatedAt())
	assert.Equal(t, 2, l.GetPlayer(2).GetEliminatedAt())
	assert.Equal(t, 3, l.GetPlayer(3).GetEliminatedAt())
	assert.Equal(t, 4, l.GetPlayer(0).GetEliminatedAt())
}

// **切り札は無い。** リードのスートで最も強い札が取る。
func TestLingerLonger_TrickWinnerIsTheHighestOfTheLedSuit(t *testing.T) {
	l := newTestLingerLonger(t)
	l.SetLeadPlayerIdxForTest(0)
	l.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 9, false)},
		{PlayerIdx: 1, Card: NewCard(CardDesignHeart, 1, false)},
		{PlayerIdx: 2, Card: NewCard(CardDesignSpade, 1, false)},
		{PlayerIdx: 3, Card: NewCard(CardDesignSpade, 13, false)},
	})
	assert.Equal(t, 2, l.TrickWinnerForTest(), "♠A が最強、別スートの A は勝てない")
}

// **トリックの途中で脱落しても、残りの席は出し切れる。**
func TestLingerLonger_AnEliminationMidTrickDoesNotSkipTheRemainingSeat(t *testing.T) {
	l := newTestLingerLonger(t)
	l.SetLeadPlayerIdxForTest(0)
	l.SetCurrentPlayerIdxForTest(0)
	for i := range l.GetPlayerCnt() {
		l.GiveHandForTest(i, NewCard(CardDesignSpade, 7+i, false), NewCard(CardDesignHeart, 7+i, false))
	}
	l.GiveHandForTest(2, NewCard(CardDesignSpade, 9, false))

	require.NoError(t, l.PlayForTest(0, 0))
	require.NoError(t, l.PlayForTest(1, 0))
	require.NoError(t, l.PlayForTest(2, 0))
	assert.Len(t, l.GetCurrentTrick(), 3, "席 3 を飛ばしてトリックを解決していない")
}

func TestLingerLonger_RejectsOutOfTurnAndBadIndices(t *testing.T) {
	l := newTestLingerLonger(t)
	idx := l.GetCurrentPlayerIdx()
	assert.Error(t, l.PlayForTest((idx+1)%l.GetPlayerCnt(), 0), "手番でない席は打てない")
	assert.Error(t, l.PlayForTest(idx, -1))
	assert.Error(t, l.PlayForTest(idx, 999))

	l.GiveUp()
	assert.Error(t, l.PlayForTest(idx, 0), "終局後は打てない")
}

func TestLingerLonger_PublicEntryPointsGuardTheTurn(t *testing.T) {
	l := newTestLingerLonger(t)
	l.SetCurrentPlayerIdxForTest(1)
	assert.False(t, l.IsHumanTurn())
	assert.Error(t, l.PlayerPlay(0))

	before := l.GetPlayer(1).GetCardsSize()
	l.CpuPlay()
	assert.Equal(t, before-1, l.GetPlayer(1).GetCardsSize())

	l.SetCurrentPlayerIdxForTest(0)
	assert.True(t, l.IsHumanTurn())
	humanBefore := l.GetPlayer(0).GetCardsSize()
	l.CpuPlay()
	assert.Equal(t, humanBefore, l.GetPlayer(0).GetCardsSize(), "人間の番では CPU は動かない")
	require.NoError(t, l.PlayerPlay(l.GetValidPlayIndices(0)[0]))
}

// **取れば補充できるので、取りにいくのが基本。**
func TestLingerLonger_CpuTakesTheTrickWhenItCan(t *testing.T) {
	l := newTestLingerLonger(t)
	l.SetLeadPlayerIdxForTest(0)
	l.SetCurrentPlayerIdxForTest(1)
	l.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 0, Card: NewCard(CardDesignSpade, 9, false)},
	})
	l.GiveHandForTest(1, NewCard(CardDesignSpade, 8, false), NewCard(CardDesignSpade, 13, false))
	assert.Equal(t, 1, l.CpuChoiceForTest(1), "勝てる札を選ぶ")

	// **取れないなら安い札を捨てる。**
	l.GiveHandForTest(1, NewCard(CardDesignSpade, 2, false), NewCard(CardDesignSpade, 8, false))
	assert.Equal(t, 0, l.CpuChoiceForTest(1))
}

// **CPU は必ず合法手を返す。**
func TestLingerLonger_CpuAlwaysChoosesLegally(t *testing.T) {
	for n := LingerLongerPlayerCntMin; n <= LingerLongerPlayerCntMax; n++ {
		for range 15 {
			l := NewLingerLonger(nil, LingerLongerConfig{PlayerCnt: n})
			l.Reset()
			for turns := 0; !l.GetGameEndFlag() && turns < 500; turns++ {
				idx := l.GetCurrentPlayerIdx()
				choice := l.CpuChoiceForTest(idx)
				require.Contains(t, l.GetValidPlayIndices(idx), choice)
				require.NoError(t, l.PlayForTest(idx, choice))
			}
		}
	}
}

// **どの局も必ず終わる。**
//
// 毎トリック場に人数ぶん出て補充は最大 1 枚なので、手札の総数は単調に減ります
// ——膠着の余地がありません。
func TestLingerLonger_GamesTerminate(t *testing.T) {
	for n := LingerLongerPlayerCntMin; n <= LingerLongerPlayerCntMax; n++ {
		for range 15 {
			l := NewLingerLonger(nil, LingerLongerConfig{PlayerCnt: n})
			l.Reset()
			for turns := 0; !l.GetGameEndFlag(); turns++ {
				require.Less(t, turns, 500, "%d 人: 終わらない", n)
				idx := l.GetCurrentPlayerIdx()
				require.NoError(t, l.PlayForTest(idx, l.CpuChoiceForTest(idx)))
			}
			assert.GreaterOrEqual(t, l.GetWinnerIdx(), 0)
			assert.Equal(t, l.GetPlayerCnt()-1, l.GetEliminatedCnt(),
				"%d 人: 勝者以外は脱落しているか、全員尽きている", n)
		}
	}
}

func TestLingerLonger_GiveUp(t *testing.T) {
	l := newTestLingerLonger(t)
	l.GiveUp()
	assert.True(t, l.GetGameEndFlag())
	assert.Positive(t, l.GetWinnerIdx(), "投了した席は勝者にならない")
	before := l.GetWinnerIdx()
	l.GiveUp()
	assert.Equal(t, before, l.GetWinnerIdx())
}

func TestLingerLonger_Hint(t *testing.T) {
	l := newTestLingerLonger(t)
	l.SetLeadPlayerIdxForTest(0)
	l.SetCurrentPlayerIdxForTest(0)
	l.SetCurrentTrickForTest([]*TrickCard{
		{PlayerIdx: 1, Card: NewCard(CardDesignSpade, 9, false)},
	})
	l.GiveHandForTest(0, NewCard(CardDesignSpade, 13, false))

	h := l.GetHint()
	require.NotNil(t, h)
	require.NotNil(t, h.CardIndex)
	assert.Equal(t, "lingerlongerWinTrick", h.Reason, "取れば補充できる")

	// **山札が空なら取る価値が変わる。**
	l.DrainStockForTest()
	assert.Equal(t, "lingerlongerNoStock", l.GetHint().Reason)

	// 取れないなら安く出す。
	l.GiveHandForTest(0, NewCard(CardDesignSpade, 2, false))
	assert.Equal(t, "lingerlongerDuck", l.GetHint().Reason)

	l.SetCurrentPlayerIdxForTest(1)
	assert.Nil(t, l.GetHint(), "相手の手番では助言しない")

	l.SetCurrentPlayerIdxForTest(0)
	l.GiveUp()
	assert.Nil(t, l.GetHint(), "終局後は助言しない")
}

func TestLingerLonger_AccessorsAndBounds(t *testing.T) {
	l := newTestLingerLonger(t)
	assert.Nil(t, l.GetPlayer(-1))
	assert.Nil(t, l.GetPlayer(99))
	assert.Empty(t, l.GetValidPlayIndices(-1))
	assert.Empty(t, l.GetValidPlayIndices(99))
	assert.Equal(t, LingerLongerDefaultPlayerCnt, l.GetPlayerCnt())
	assert.Zero(t, l.GetTrickNumber())
	assert.NotEmpty(t, l.GetActionLog())
	assert.Equal(t, DefaultLingerLongerConfig(), l.GetConfig())

	l.SetConfig(LingerLongerConfig{PlayerCnt: 6})
	assert.Equal(t, 6, l.GetPlayerCnt())
	assert.NotPanics(t, l.Reset)
	assert.Equal(t, 6, l.GetPlayer(5).GetCardsSize())
}

func TestLingerLongerConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultLingerLongerConfig().Validate())
	for n := LingerLongerPlayerCntMin; n <= LingerLongerPlayerCntMax; n++ {
		assert.NoError(t, LingerLongerConfig{PlayerCnt: n}.Validate())
	}
	assert.Error(t, LingerLongerConfig{PlayerCnt: LingerLongerPlayerCntMin - 1}.Validate())
	assert.Error(t, LingerLongerConfig{PlayerCnt: LingerLongerPlayerCntMax + 1}.Validate())

	l := NewLingerLonger(nil, LingerLongerConfig{PlayerCnt: 99})
	assert.Equal(t, LingerLongerDefaultPlayerCnt, l.GetPlayerCnt())
}

// **書き込み側は、自分の codec が弾く盤面を作らない（#5316 で学んだ形）。**
//
// 実際の局を最後まで回し、**毎手ごとに**保存して読み直せることを確かめます。
// 手で書いた変異テストでは出ない違反がここで出ます。
func TestLingerLonger_EveryReachableStateSurvivesARoundTrip(t *testing.T) {
	for n := LingerLongerPlayerCntMin; n <= LingerLongerPlayerCntMax; n++ {
		for range 10 {
			l := NewLingerLonger(nil, LingerLongerConfig{PlayerCnt: n})
			l.Reset()

			for turns := 0; ; turns++ {
				require.Less(t, turns, 500, "%d 人: 終わらない", n)

				data, err := json.Marshal(l)
				require.NoError(t, err)
				var back LingerLonger
				require.NoError(t, json.Unmarshal(data, &back),
					"%d 人 %d 手目: 書き込み側が codec の不変条件を破った", n, turns)

				if l.GetGameEndFlag() {
					break
				}
				idx := l.GetCurrentPlayerIdx()
				require.NoError(t, l.PlayForTest(idx, l.CpuChoiceForTest(idx)))
			}
		}
	}
}

// **壊れたスナップショットは弾く。**
func TestLingerLonger_UnmarshalRejectsBrokenSnapshots(t *testing.T) {
	snapshot := func(t *testing.T) map[string]any {
		t.Helper()
		l := newTestLingerLonger(t)
		require.NoError(t, l.PlayForTest(l.GetCurrentPlayerIdx(), l.CpuChoiceForTest(l.GetCurrentPlayerIdx())))
		data, err := json.Marshal(l)
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
		{"winner without the game ending", func(m map[string]any) { m["wi"] = 1 }},
		{"current player out of range", func(m map[string]any) { m["ci"] = 9 }},
		{"lead player out of range", func(m map[string]any) { m["li"] = -1 }},
		{"last draw out of range", func(m map[string]any) { m["ld"] = 9 }},
		{"negative trick number", func(m map[string]any) { m["tn"] = -1 }},
		{"eliminated count above the table", func(m map[string]any) { m["ec"] = 9 }},
		{"eliminated count with nobody out", func(m map[string]any) { m["ec"] = 1 }},
		{"config out of range", func(m map[string]any) { m["cf"] = map[string]any{"p": 9} }},
		{"player count disagrees with the seats", func(m map[string]any) { m["cf"] = map[string]any{"p": 5} }},
		{"missing trump cards", func(m map[string]any) { m["tc"] = nil }},
		{"a trick entry with no card", func(m map[string]any) {
			m["ct"] = []any{map[string]any{"playerIdx": 0}}
		}},
		{"a trick entry with a bad seat", func(m map[string]any) {
			m["ct"] = []any{map[string]any{"playerIdx": 9, "card": map[string]any{"d": 1, "v": 9, "j": false}}}
		}},
		// **札は 52 枚しかない（#5314 の形）。**
		{"a card appears from nowhere", func(m map[string]any) {
			m["ct"] = append(m["ct"].([]any), map[string]any{
				"playerIdx": 1, "card": map[string]any{"d": 1, "v": 9, "j": false},
			})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := snapshot(t)
			tc.mutate(s)
			data, err := json.Marshal(s)
			require.NoError(t, err)
			var restored LingerLonger
			assert.Error(t, json.Unmarshal(data, &restored))
		})
	}

	// **負のコントロール: 手を加えていないスナップショットは通り、使っても落ちない。**
	data, err := json.Marshal(snapshot(t))
	require.NoError(t, err)
	var ok LingerLonger
	require.NoError(t, json.Unmarshal(data, &ok))
	assert.NotPanics(t, func() {
		_ = ok.GetValidPlayIndices(ok.GetCurrentPlayerIdx())
		_ = ok.GetHint()
		_ = ok.TrickWinnerForTest()
	})
}

// **脱落と手札は対。** 脱落しているのに手札がある盤面は存在しない。
func TestLingerLongerPlayer_UnmarshalRejectsAContradictoryElimination(t *testing.T) {
	held := NewLingerLongerPlayer(false)
	held.AddCard(NewCard(CardDesignSpade, 5, false))
	held.SetEliminatedAt(1)
	data, err := json.Marshal(held)
	require.NoError(t, err)
	var p LingerLongerPlayer
	assert.Error(t, json.Unmarshal(data, &p), string(data))

	for _, body := range []string{`{"ea":-1}`, `{"tw":-1}`, `{"ea":99}`} {
		var bad LingerLongerPlayer
		assert.Error(t, json.Unmarshal([]byte(body), &bad), body)
	}

	var ok LingerLongerPlayer
	assert.NoError(t, json.Unmarshal([]byte(`{"ea":1,"tw":3}`), &ok))
	assert.Equal(t, 1, ok.GetEliminatedAt())
	assert.Equal(t, 3, ok.GetTricksWon())

	// **逆は成り立たない。** 最後の 1 枚を出した席は、補充が済むまで脱落が決まらない
	// ので、「手札が空だが在席」はトリックの途中の正当な状態です。
	var midTrick LingerLongerPlayer
	assert.NoError(t, json.Unmarshal([]byte(`{"ea":0,"tw":1}`), &midTrick))
	assert.False(t, midTrick.IsEliminated())
}
