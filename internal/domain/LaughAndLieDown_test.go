//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func lldCard(design, value int) *Card { return NewCard(design, value, true) }

func TestLaughAndLieDown_TheDealIsEightEachAndTwelveFaceUp(t *testing.T) {
	// #4396 は「残りを場札の**山**として中央に置く」とするが、伏せた山札は
	// 存在しない。12 枚が表向きに広がっているからこそ、どのランクが何枚残って
	// いるかが全員に見え、3 枚取りの判断が成立する。
	// **Reset() は配って終わりではない。**最後に skipPlayersWhoCannotCapture が
	// 走り、配った時点で合法手を持たない席は lieDown で手札を全部場に置く。
	// 5000 回試して 11 回（約 0.22%）はその席が出るので、「Reset 直後は全員 8 枚」
	// は配りの検証ではなく**配りに依存する賭け**になっていた。
	//
	// 枚数の保存則は降ろした後も成り立つ。そちらで配りを検証する。
	l := NewDefaultLaughAndLieDown()
	l.Reset()

	total := len(l.GetLayout())
	for i := range l.GetPlayers() {
		size := l.GetPlayer(i).GetCardsSize()
		assert.LessOrEqual(t, size, LaughAndLieDownHandSize, "seat %d cannot hold more than it was dealt", i)
		total += size
	}
	assert.Equal(t, 52, total, "no card may be held back in a hidden stock")

	// 場は 12 枚で始まる。降ろした席があればその分だけ増える。
	assert.GreaterOrEqual(t, len(l.GetLayout()), LaughAndLieDownLayoutSize)
}

func TestLaughAndLieDown_ThePotIsExactlyWhatTheSettlementPaysOut(t *testing.T) {
	// これが精算表の裏取り。ポットは 3 + 2*4 = 11。52 枚すべてが誰かの取り札に
	// なるので、8 枚との過不足の合計は必ず (52 - 40)/2 = 6。「最後の 1 人に 5」を
	// 足してちょうど 11 になる。issue の「勝者がポット総取り」ではこの一致は
	// 起きないので、どちらが正しいかはこの算術が決める。
	assert.Equal(t, 11, LaughAndLieDownPot)
	overShort := (52 - LaughAndLieDownHandSize*LaughAndLieDownPlayerCnt) / 2
	assert.Equal(t, 6, overShort)
	assert.Equal(t, LaughAndLieDownPot, LaughAndLieDownLastInBonus+overShort)
}

func TestLaughAndLieDown_CapturesAgainstAnyTableCardNotJustATopCard(t *testing.T) {
	// #4396 は「場札の**一番上**と同ランク」とするが、場は山ではなく広がった
	// 12 枚で、どの札とでも合えば取れる。
	l := NewDefaultLaughAndLieDown()
	l.Reset()
	p := l.GetPlayer(1)
	p.Reset()
	p.AddCard(lldCard(CardDesignSpade, 7))
	l.SetLayoutForTest([]*Card{
		lldCard(CardDesignHeart, 2),
		lldCard(CardDesignClover, 9),
		lldCard(CardDesignDiamond, 7), // 一番上でも一番下でもない
		lldCard(CardDesignSpade, 4),
	})
	l.SetCurrentPlayerForTest(1)

	assert.Equal(t, []int{0}, l.GetValidPlayIndices(1))
	require.NoError(t, l.PlayCard(1, 0, 1))
	// 取得枚数だけを見る。PlayCard は同じ呼び出しの中で他家の手番まで進めるので、
	// 場の枚数は「取れない席が手札を落とした」ぶん動く。
	assert.Equal(t, 2, l.GetWonCount(1), "the played card and the matched one")
}

func TestLaughAndLieDown_OneCardTakesOneOrThree(t *testing.T) {
	// 原典は「1 枚で 1 枚または 3 枚を取る」。2 枚取りも 4 枚取りも無い。
	newGame := func() *LaughAndLieDown {
		l := NewDefaultLaughAndLieDown()
		l.Reset()
		p := l.GetPlayer(1)
		p.Reset()
		p.AddCard(lldCard(CardDesignSpade, 7))
		l.SetLayoutForTest([]*Card{
			lldCard(CardDesignHeart, 7), lldCard(CardDesignClover, 7), lldCard(CardDesignDiamond, 7),
			lldCard(CardDesignSpade, 4),
		})
		l.SetCurrentPlayerForTest(1)
		return l
	}

	// 取得枚数で 1 枚取りと 3 枚取りは完全に区別できる。場の枚数は、同じ
	// 呼び出しの中で他家が降りて手札を落とすぶん動くので見ない。
	three := newGame()
	assert.True(t, three.CanTakeThree(1, 0))
	require.NoError(t, three.PlayCard(1, 0, 3))
	assert.Equal(t, 4, three.GetWonCount(1), "the played card and all three")

	one := newGame()
	require.NoError(t, one.PlayCard(1, 0, 1))
	assert.Equal(t, 2, one.GetWonCount(1), "taking one is still allowed with three on the table")

	bad := newGame()
	assert.Error(t, bad.PlayCard(1, 0, 2), "two is not a legal take")
	assert.Error(t, bad.PlayCard(1, 0, 4), "nor is four")
}

func TestLaughAndLieDown_ThreeIsRefusedWhenTheTableCannotSupplyIt(t *testing.T) {
	l := NewDefaultLaughAndLieDown()
	l.Reset()
	p := l.GetPlayer(1)
	p.Reset()
	p.AddCard(lldCard(CardDesignSpade, 7))
	l.SetLayoutForTest([]*Card{lldCard(CardDesignHeart, 7), lldCard(CardDesignClover, 7)})
	l.SetCurrentPlayerForTest(1)

	assert.False(t, l.CanTakeThree(1, 0))
	assert.ErrorContains(t, l.PlayCard(1, 0, 3), "only 2 card(s)")
}

func TestLaughAndLieDown_LyingDownFeedsTheTableRatherThanJustEliminating(t *testing.T) {
	// #4396 は「山から 1 枚引いて…揃わなければ脱落」とするが、引く手順は無く、
	// 降りた人の**手札は全部場に置かれて他家の獲物になる**。ここを落とすと
	// ゲームの緊張がそのまま消える。
	l := NewDefaultLaughAndLieDown()
	l.Reset()
	for i := range l.GetPlayers() {
		l.GetPlayer(i).Reset()
	}
	// 席 1 は場のどれとも合わない 3 枚を持つ。席 2 と席 3 は札を持つので、
	// 席 1 が降りても「手札を持つのが 1 人だけ」にはならず、進行は続く。
	l.GetPlayer(1).AddCard(lldCard(CardDesignSpade, 11))
	l.GetPlayer(1).AddCard(lldCard(CardDesignHeart, 11))
	l.GetPlayer(1).AddCard(lldCard(CardDesignClover, 11))
	l.GetPlayer(2).AddCard(lldCard(CardDesignSpade, 5))
	l.GetPlayer(2).AddCard(lldCard(CardDesignDiamond, 11))
	l.GetPlayer(3).AddCard(lldCard(CardDesignClover, 5))
	l.SetLayoutForTest([]*Card{lldCard(CardDesignHeart, 5)})

	// 手番が回った時点で取れないので自動的に降りる。強制なので選択肢にしない。
	l.SetCurrentPlayerForTest(1)
	l.skipPlayersWhoCannotCapture()

	require.False(t, l.GetGameEndFlag(), "two seats still hold cards, so play continues")

	assert.True(t, l.IsLaidDown(1))
	assert.Zero(t, l.GetPlayer(1).GetCardsSize())
	assert.Len(t, l.GetLayout(), 4, "the whole hand joined the table")
	assert.Equal(t, 2, l.GetCurrentPlayerIdx())

	// 席 2 は、席 1 が置いていった J を取れるようになっている。
	assert.Contains(t, l.GetValidPlayIndices(2), 1)
}

func TestLaughAndLieDown_PlayCeasesWhenOnlyOnePlayerStillHoldsCards(t *testing.T) {
	l := NewDefaultLaughAndLieDown()
	l.Reset()
	for i := range l.GetPlayers() {
		l.GetPlayer(i).Reset()
	}
	l.GetPlayer(3).AddCard(lldCard(CardDesignSpade, 5))
	l.SetLayoutForTest([]*Card{lldCard(CardDesignHeart, 9)})
	l.SetCurrentPlayerForTest(0)
	l.skipPlayersWhoCannotCapture()

	require.True(t, l.GetGameEndFlag())
	assert.Equal(t, LaughAndLieDownPhaseGameEnd, l.GetPhase())
	assert.Equal(t, 3, l.GetLastInIdx(), "the last one still holding cards is last in")
}

func TestLaughAndLieDown_TheDealerSweepsTheLeftoversAndEveryCardIsAccountedFor(t *testing.T) {
	// 親だけ 1 多く出しているので、残り札は親の取り札に入る (原典どおり)。
	// 副次的に、52 枚が必ず誰かの取り札になるので過不足の合計が 6 に固定される。
	l := NewDefaultLaughAndLieDown()
	l.Reset()
	for range 200 {
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

	total := 0
	for i := range l.GetPlayers() {
		total += l.GetWonCount(i)
		assert.Zero(t, l.GetPlayer(i).GetCardsSize(), "seat %d still holds cards", i)
	}
	assert.Empty(t, l.GetLayout())
	assert.Equal(t, 52, total, "every card must end up in somebody's won pile")
}

func TestLaughAndLieDown_SettlementPaysTheLastInAndThenTheOverShort(t *testing.T) {
	l := NewDefaultLaughAndLieDown()
	l.Reset()
	for range 200 {
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

	for i := range l.GetPlayers() {
		ante := LaughAndLieDownAnte
		if i == l.GetDealerIdx() {
			ante = LaughAndLieDownDealerAnte
		}
		want := -ante + (l.GetWonCount(i)-LaughAndLieDownHandSize)/2
		if i == l.GetLastInIdx() {
			want += LaughAndLieDownLastInBonus
		}
		assert.Equal(t, want, l.GetScore(i), "seat %d", i)
	}
}

func TestLaughAndLieDown_CpuPrefersTheThreeCardTake(t *testing.T) {
	// 取得枚数がそのまま精算になるので、3 枚取れるときに 1 枚で済ませる理由が無い。
	l := NewDefaultLaughAndLieDown()
	l.Reset()
	p := l.GetPlayer(1)
	p.Reset()
	p.AddCard(lldCard(CardDesignSpade, 2))
	p.AddCard(lldCard(CardDesignSpade, 7))
	l.SetLayoutForTest([]*Card{
		lldCard(CardDesignHeart, 2),
		lldCard(CardDesignHeart, 7), lldCard(CardDesignClover, 7), lldCard(CardDesignDiamond, 7),
	})
	l.SetCurrentPlayerForTest(1)

	action := l.LaughAndLieDownCpuDecide(1)
	assert.Equal(t, 1, action.HandIdx, "the seven, not the two")
	assert.Equal(t, 3, action.TakeCount)
}

func TestLaughAndLieDown_CpuReportsNoMoveWhenItCannotCapture(t *testing.T) {
	l := NewDefaultLaughAndLieDown()
	l.Reset()
	p := l.GetPlayer(1)
	p.Reset()
	p.AddCard(lldCard(CardDesignSpade, 2))
	l.SetLayoutForTest([]*Card{lldCard(CardDesignHeart, 9)})
	l.SetCurrentPlayerForTest(1)

	assert.Equal(t, -1, l.LaughAndLieDownCpuDecide(1).HandIdx)
}

func TestLaughAndLieDown_RejectsIllegalRequests(t *testing.T) {
	l := NewDefaultLaughAndLieDown()
	l.Reset()
	cur := l.GetCurrentPlayerIdx()

	assert.Error(t, l.PlayCard(cur, -1, 1))
	assert.Error(t, l.PlayCard(cur, 99, 1))
	assert.Error(t, l.PlayCard((cur+1)%LaughAndLieDownPlayerCnt, 0, 1), "not that player's turn")

	// 場に無いランクは出せない。
	p := l.GetPlayer(cur)
	p.Reset()
	p.AddCard(lldCard(CardDesignSpade, 2))
	l.SetLayoutForTest([]*Card{lldCard(CardDesignHeart, 9)})
	assert.ErrorContains(t, l.PlayCard(cur, 0, 1), "no card of that rank")
}

func TestLaughAndLieDown_SurvivesAKVRoundTrip(t *testing.T) {
	l := NewDefaultLaughAndLieDown()
	l.Reset()
	for range 3 {
		idx := l.GetCurrentPlayerIdx()
		action := l.LaughAndLieDownCpuDecide(idx)
		if action.HandIdx < 0 || l.GetGameEndFlag() {
			break
		}
		require.NoError(t, l.PlayCard(idx, action.HandIdx, action.TakeCount))
	}

	data, err := json.Marshal(l)
	require.NoError(t, err)

	restored := NewDefaultLaughAndLieDown()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, l.GetPhase(), restored.GetPhase())
	assert.Equal(t, len(l.GetLayout()), len(restored.GetLayout()))
	assert.Equal(t, l.GetCurrentPlayerIdx(), restored.GetCurrentPlayerIdx())
	assert.Equal(t, l.GetDealerIdx(), restored.GetDealerIdx())
	assert.Equal(t, l.GetLastInIdx(), restored.GetLastInIdx())
	for i := range l.GetPlayers() {
		assert.Equal(t, l.GetPlayer(i).GetCardsSize(), restored.GetPlayer(i).GetCardsSize(), "hand %d", i)
		// 取得枚数は精算そのものなので、往復で失うと点が変わる。
		assert.Equal(t, l.GetWonCount(i), restored.GetWonCount(i), "won %d", i)
		assert.Equal(t, l.IsLaidDown(i), restored.IsLaidDown(i), "laidDown %d", i)
		assert.Equal(t, l.GetScore(i), restored.GetScore(i), "score %d", i)
	}
}

func TestLaughAndLieDown_UnmarshalRejectsAndClampsHostileSnapshots(t *testing.T) {
	valid, err := json.Marshal(NewDefaultLaughAndLieDown())
	require.NoError(t, err)

	for name, payload := range map[string]string{
		"not json":      "{",
		"seat count":    `{"pl":[],"cfg":{"cd":0},"ph":0}`,
		"bad config":    `{"pl":[{},{},{},{},{}],"cfg":{"cd":99},"ph":0}`,
		"unknown phase": `{"pl":[{},{},{},{},{}],"cfg":{"cd":0},"ph":9}`,
	} {
		t.Run(name, func(t *testing.T) {
			assert.Error(t, json.Unmarshal([]byte(payload), NewDefaultLaughAndLieDown()))
		})
	}

	// 席数に足りない補助スライスは、落ちるのではなく詰め直される。
	short := `{"pl":[{},{},{},{},{}],"cfg":{"cd":0},"ph":0,"wn":[[]],"ld":[true],"sc":[1],"cur":99,"dl":-1,"li":42}`
	l := NewDefaultLaughAndLieDown()
	require.NoError(t, json.Unmarshal([]byte(short), l))
	assert.Equal(t, 0, l.GetCurrentPlayerIdx(), "an out-of-range seat is clamped, not trusted")
	assert.Equal(t, 0, l.GetDealerIdx())
	assert.Equal(t, -1, l.GetLastInIdx())
	for i := range l.GetPlayers() {
		assert.Zero(t, l.GetWonCount(i), "seat %d", i)
	}
	assert.False(t, l.IsLaidDown(4), "the padded tail must not read past the supplied slice")

	require.NoError(t, json.Unmarshal(valid, NewDefaultLaughAndLieDown()))
}

// **配りそのものは 8 枚ずつ。**上のテストが枚数の保存則に寄せたぶん、
// 「配った直後」の姿はここで固定する。合法手を持たない席が降ろす前の状態を
// 見たいので、skipPlayersWhoCannotCapture を走らせない Reset の中身を辿る。
func TestLaughAndLieDown_TheDealItselfIsEightEach(t *testing.T) {
	// 降ろしが起きるのは約 0.22%。1000 回のうち少なくとも 1 回は「全員 8 枚」の
	// 局面が出るので、そこで配りの形を確かめる。出なければ配りが壊れている。
	for range 1000 {
		l := NewDefaultLaughAndLieDown()
		l.Reset()

		allFull := true
		for i := range l.GetPlayers() {
			if l.GetPlayer(i).GetCardsSize() != LaughAndLieDownHandSize {
				allFull = false
				break
			}
		}
		if !allFull {
			continue
		}
		assert.Equal(t, LaughAndLieDownLayoutSize, len(l.GetLayout()),
			"with every seat still holding its deal, the layout is exactly the face-up spread")
		return
	}
	t.Fatal("1000 deals and never once did every seat keep its eight cards")
}
