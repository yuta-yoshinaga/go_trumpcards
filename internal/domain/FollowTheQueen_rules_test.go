//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Follow the Queen's divergence from Seven-Card Stud ---
//
// The clone source is plain 7-card stud: the streets, the betting and the
// showdown are all inherited unchanged. The only rule on top is a wild card
// that moves during the hand, and everything below is about that.

// **クイーンは常にワイルド。**表向きに出たかどうかに関わらない。
func TestFollowTheQueen_QueensAreAlwaysWild(t *testing.T) {
	s := newTestFollowTheQueen()

	assert.True(t, s.IsWild(NewCard(CardDesignSpade, FollowTheQueenQueenValue, true)))
	assert.True(t, s.IsWild(NewCard(CardDesignHeart, FollowTheQueenQueenValue, false)))
	// 負のコントロール: 何も出ていないうちは Q 以外はワイルドでない。
	assert.False(t, s.IsWild(NewCard(CardDesignSpade, 7, true)))
	assert.Equal(t, 0, s.GetWildRank(), "第2のワイルドはまだ無い")
}

// **表向きの Q の次の 1 枚がワイルドのランクを決める。**これがゲーム名の由来。
func TestFollowTheQueen_TheCardAfterAnUpQueenSetsTheWildRank(t *testing.T) {
	s := newTestFollowTheQueen()

	s.noteUpCard(NewCard(CardDesignSpade, FollowTheQueenQueenValue, true))
	assert.Equal(t, 0, s.GetWildRank(), "Q 自体はまだランクを決めない")

	s.noteUpCard(NewCard(CardDesignHeart, 7, true))
	assert.Equal(t, 7, s.GetWildRank())
	assert.True(t, s.IsWild(NewCard(CardDesignClover, 7, false)),
		"伏せ札でも 7 はワイルド")

	// 負のコントロール: その後の普通の札はランクを変えない。
	s.noteUpCard(NewCard(CardDesignClover, 3, true))
	assert.Equal(t, 7, s.GetWildRank())
}

// **次の Q が出たらワイルドは移る。**同時に 2 つのランクがワイルドになることはない。
func TestFollowTheQueen_ASecondQueenMovesTheWildRank(t *testing.T) {
	s := newTestFollowTheQueen()
	s.noteUpCard(NewCard(CardDesignSpade, FollowTheQueenQueenValue, true))
	s.noteUpCard(NewCard(CardDesignHeart, 7, true))
	require.Equal(t, 7, s.GetWildRank())

	s.noteUpCard(NewCard(CardDesignClover, FollowTheQueenQueenValue, true))
	assert.Equal(t, 0, s.GetWildRank(), "次の 1 枚が来るまで第2ワイルドは無い")
	assert.False(t, s.IsWild(NewCard(CardDesignSpade, 7, false)), "7 はもうワイルドでない")

	s.noteUpCard(NewCard(CardDesignDiamond, 4, true))
	assert.Equal(t, 4, s.GetWildRank())
	assert.False(t, s.IsWild(NewCard(CardDesignSpade, 7, false)))
}

// **Q が最後の表向き札なら第2ワイルドは無い。**次の札が来ないので決まらない。
func TestFollowTheQueen_AQueenAsTheLastUpCardLeavesNoSecondWild(t *testing.T) {
	s := newTestFollowTheQueen()
	s.noteUpCard(NewCard(CardDesignSpade, FollowTheQueenQueenValue, true))
	s.noteUpCard(NewCard(CardDesignHeart, 7, true))
	require.Equal(t, 7, s.GetWildRank())

	// 最後の表向きが Q。以降 noteUpCard は呼ばれない。
	s.noteUpCard(NewCard(CardDesignClover, FollowTheQueenQueenValue, true))
	assert.Equal(t, 0, s.GetWildRank())
	// Q 自体は最後までワイルドのまま。
	assert.True(t, s.IsWild(NewCard(CardDesignDiamond, FollowTheQueenQueenValue, false)))
}

// **普通の表向き札はワイルドを動かさない。**Q の直後の 1 枚だけが引き金で、
// その後の表向き札は何も変えない。
func TestFollowTheQueen_AnOrdinaryUpCardDoesNotMoveTheWild(t *testing.T) {
	s := newTestFollowTheQueen()
	s.noteUpCard(NewCard(CardDesignSpade, FollowTheQueenQueenValue, true))
	s.noteUpCard(NewCard(CardDesignHeart, 7, true))
	require.Equal(t, 7, s.GetWildRank())

	s.noteUpCard(NewCard(CardDesignDiamond, 9, true))
	assert.Equal(t, 7, s.GetWildRank(), "普通の表向き札は何も変えない")
}

// dealtBoard は Reset 済みで、**配りに依存しない**盤を返す。
//
// Reset はシャッフルして実際に配るので、そのままでは (a) 3rd street で Q が出て
// ワイルドが既に立っている、(b) Reset 内の CPU 行動で誰かが降りていて
// dealStreetCard が 4 枚配らない、の二つが乱数で起きる。どちらもこの下の
// テストの前提を壊す（実際に 12 回に 1 回落ちた）。両方をここで潰す。
func dealtBoard(t *testing.T) *FollowTheQueen {
	t.Helper()
	s := newTestFollowTheQueen()
	require.NoError(t, s.Reset())
	for _, p := range s.players {
		p.SetFolded(false)
		p.SetAllIn(false)
	}
	s.SetWildRankForTest(0) // queenPending も落ちる
	require.Equal(t, 0, s.GetWildRank())
	return s
}

// stackUpcoming は山札を「次に引かれる位置」から cards の順に固定する。
// 配置漏れは即座に止める —— 黙って飛ばされると、狙っていない配りを検査する
// テストになる。
func stackUpcoming(t *testing.T, s *FollowTheQueen, cards ...*Card) {
	t.Helper()
	s.trumpCards.ReplenishForTest()
	require.Equal(t, len(cards), s.trumpCards.StackTopForTest(cards...),
		"山札に積めなかったカードがある")
}

// **配りの経路そのものを見る。**上の一連は noteUpCard を直接叩いているので、
// 「配りが noteUpCard を呼ぶ」という肝心の配線が一切カバーされていなかった
// (レビュー指摘: dealStreetCard の noteUpCard 呼び出しを 2 か所とも消しても
// ドメインのテストが全部緑のままだった)。ここは山札を積んで **本番の配布関数**
// から流す。
func TestFollowTheQueen_DealingAnUpQueenSetsTheWildThroughTheRealDealPath(t *testing.T) {
	s := dealtBoard(t)

	// 4th street の表向き配布を、Q → ♦7 → ... の順に固定する。
	stackUpcoming(t, s,
		NewCard(CardDesignSpade, FollowTheQueenQueenValue, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignHeart, 4, false),
	)
	s.dealStreetCard(true)

	assert.Equal(t, 7, s.GetWildRank(),
		"表向きに配った Q の次のカードのランクがワイルドになる")
	assert.True(t, s.IsWild(NewCard(CardDesignClover, 7, false)))
	// 配布経由でもプレイヤー側まで届いていること。
	for i, p := range s.players {
		assert.Equal(t, 7, p.GetWildRank(), "player %d", i)
	}
}

// **伏せ配布は Q でもワイルドを動かさない。**表向きに出た Q だけが引き金になる。
// 本番の伏せ配布経路 (dealStreetCard(false)) から流して確かめる —— 直接
// noteUpCard を叩いてしまうと「伏せ札は noteUpCard を通らない」という肝心の
// 取り決めを一度も検査しないことになる。
func TestFollowTheQueen_DealingADownQueenDoesNotMoveTheWild(t *testing.T) {
	s := dealtBoard(t)

	// 先に表向きでワイルドを 7 に固定しておく。
	stackUpcoming(t, s,
		NewCard(CardDesignSpade, FollowTheQueenQueenValue, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignHeart, 4, false),
	)
	s.dealStreetCard(true)
	require.Equal(t, 7, s.GetWildRank())

	// 7th street は伏せ配布。Q を配っても何も起きてはいけない。
	stackUpcoming(t, s,
		NewCard(CardDesignHeart, FollowTheQueenQueenValue, false),
		NewCard(CardDesignClover, FollowTheQueenQueenValue, false),
		NewCard(CardDesignDiamond, FollowTheQueenQueenValue, false),
		NewCard(CardDesignSpade, 5, false),
	)
	s.dealStreetCard(false)
	assert.Equal(t, 7, s.GetWildRank(), "伏せ札の Q はワイルドを動かさない")

	// **負のコントロール**: 伏せた Q が queenPending を立てていないこと。
	// 立っていれば、次の表向き札がワイルドを 5 に書き換えてしまう。
	stackUpcoming(t, s,
		NewCard(CardDesignSpade, 5, false),
		NewCard(CardDesignHeart, 6, false),
		NewCard(CardDesignClover, 8, false),
		NewCard(CardDesignDiamond, 9, false),
	)
	s.dealStreetCard(true)
	assert.Equal(t, 7, s.GetWildRank(),
		"伏せた Q が保留されていれば、この表向き札がワイルドを奪ってしまう")
}

// **サードストリートの配布経路も同じ配線を通っているか。**Reset は内部で
// シャッフルするので山札を積めない。代わりに、実際に配られた表向き札から
// ルール通りの答えを組み立てて突き合わせる —— どの配りでも成立する検査で、
// Reset の noteUpCard 呼び出しを消せば落ちる。
func TestFollowTheQueen_ResetSetsTheWildFromTheDoorCardsItDealt(t *testing.T) {
	s := newTestFollowTheQueen()

	sawQueen := false
	for round := 0; round < 200; round++ {
		require.NoError(t, s.Reset())

		// 配布順 = ストリートごとに、ディーラーの次の席から。Reset の中で
		// CPU が動いて先のストリートまで進むことがあるので、1 人 1 枚とは
		// 限らない。
		maxDoor := 0
		for _, p := range s.players {
			if n := len(p.GetDoorCards()); n > maxDoor {
				maxDoor = n
			}
		}
		want, pending := 0, false
		for street := 0; street < maxDoor; street++ {
			for j := 0; j < len(s.players); j++ {
				idx := (s.GetDealerIdx() + 1 + j) % len(s.players)
				door := s.players[idx].GetDoorCards()
				if street >= len(door) {
					continue
				}
				v := door[street].GetValue()
				switch {
				case v == FollowTheQueenQueenValue:
					want, pending, sawQueen = 0, true, true
				case pending:
					want, pending = v, false
				}
			}
		}
		require.Equal(t, want, s.GetWildRank(),
			"round %d: 配られたドアカードから決まるワイルドと一致しない", round)
	}
	require.True(t, sawQueen,
		"200 回配って表向きの Q が一度も出ていない —— 検査が 0 == 0 に退化している")
}

// **ワイルドはリセットで消える。**前のハンドのワイルドが残ると、配る前から
// 特定ランクが強い盤になる。
//
// **`NotEqual(7)` では検査にならない。** Reset は実際に配るので、新しい盤で
// Q → 7 が出れば 7 は正当な答えになる（40 回に 1 回落ちた）。「持ち越した 7」と
// 「配り直して出た 7」は値では区別できない。
// そこで **表向きの Q が 1 枚も出なかった盤**だけを見る。その盤の正解は必ず 0 で、
// クリアを外せば 7 のまま残るので、曖昧さなく落ちる。
func TestFollowTheQueen_ResetClearsTheWildRank(t *testing.T) {
	s := newTestFollowTheQueen()

	checked := 0
	for round := 0; round < 300 && checked == 0; round++ {
		s.SetWildRankForTest(7)
		require.Equal(t, 7, s.GetWildRank())
		require.NoError(t, s.Reset())

		queenShown := false
		for _, p := range s.players {
			for _, c := range p.GetDoorCards() {
				if c.GetValue() == FollowTheQueenQueenValue {
					queenShown = true
				}
			}
		}
		if queenShown {
			continue // この盤の答えは 0 とは限らない
		}
		assert.Equal(t, 0, s.GetWildRank(),
			"表向きの Q が 1 枚も出ていない盤なのでワイルドは無いはず")
		checked++
	}
	require.Equal(t, 1, checked,
		"300 回配って Q 無しの盤が一度も出ていない —— 検査が一度も走っていない")
}

// **ワイルドが評価に届くこと。**ゲームだけが知っていて評価器が知らないと、
// 規則は飾りになる。ショーダウンも CPU の判断も表示も、同じ 1 本を通る。
func TestFollowTheQueen_WildsReachTheHandEvaluator(t *testing.T) {
	p := NewFollowTheQueenPlayer(false, 0)
	// A♠ A♥ K♦ 9♠ + Q♣。Q はワイルドなので 3 枚の A になる。
	for _, c := range []*Card{
		NewCard(CardDesignSpade, 1, true),
		NewCard(CardDesignHeart, 1, true),
	} {
		p.AddHoleCard(c)
	}
	for _, c := range []*Card{
		NewCard(CardDesignDiamond, 13, true),
		NewCard(CardDesignSpade, 9, true),
		NewCard(CardDesignClover, FollowTheQueenQueenValue, true),
	} {
		p.AddDoorCard(c)
	}

	rank, best := p.PeekBestHand()
	assert.Equal(t, PokerHandThreeOfAKind, rank, "Q がワイルドなので A が 3 枚")
	assert.Len(t, best, 5)

	// **第2ワイルドも効くこと。**9 をワイルドにすると A が 4 枚になる。
	p.SetWildRank(9)
	rank, _ = p.PeekBestHand()
	assert.Equal(t, PokerHandFourOfAKind, rank)

	// 負のコントロール: ワイルドを一切見ない実装ならワンペアのまま。
	p.SetWildRank(0)
	plain := NewFollowTheQueenPlayer(false, 0)
	for _, c := range []*Card{
		NewCard(CardDesignSpade, 1, true),
		NewCard(CardDesignHeart, 1, true),
	} {
		plain.AddHoleCard(c)
	}
	for _, c := range []*Card{
		NewCard(CardDesignDiamond, 13, true),
		NewCard(CardDesignSpade, 9, true),
		NewCard(CardDesignClover, 11, true),
	} {
		plain.AddDoorCard(c)
	}
	rank, _ = plain.PeekBestHand()
	assert.Equal(t, PokerHandOnePair, rank, "Q を含まない同型の手はワンペアのまま")
}

// **ワイルド 4 枚以上はファイブカード。**総当たりでは出せない役なので、
// 専用の枝で拾う。
func TestFollowTheQueen_FourWildsMakeFiveOfAKind(t *testing.T) {
	wild := func(c *Card) bool { return c.GetValue() == FollowTheQueenQueenValue }
	hand := []*Card{
		NewCard(CardDesignSpade, FollowTheQueenQueenValue, true),
		NewCard(CardDesignHeart, FollowTheQueenQueenValue, true),
		NewCard(CardDesignClover, FollowTheQueenQueenValue, true),
		NewCard(CardDesignDiamond, FollowTheQueenQueenValue, true),
		NewCard(CardDesignSpade, 3, true),
	}
	assert.Equal(t, PokerHandFiveOfAKind, mustWildRank(evalFiveCardHandWithWilds(hand, wild)))

	// 3 枚 + 同ランク 2 枚でもファイブカード。
	hand2 := []*Card{
		NewCard(CardDesignSpade, FollowTheQueenQueenValue, true),
		NewCard(CardDesignHeart, FollowTheQueenQueenValue, true),
		NewCard(CardDesignClover, FollowTheQueenQueenValue, true),
		NewCard(CardDesignSpade, 3, true),
		NewCard(CardDesignHeart, 3, true),
	}
	assert.Equal(t, PokerHandFiveOfAKind, mustWildRank(evalFiveCardHandWithWilds(hand2, wild)))

	// 負のコントロール: ワイルド 3 枚 + バラバラ 2 枚はファイブカードにならない。
	hand3 := []*Card{
		NewCard(CardDesignSpade, FollowTheQueenQueenValue, true),
		NewCard(CardDesignHeart, FollowTheQueenQueenValue, true),
		NewCard(CardDesignClover, FollowTheQueenQueenValue, true),
		NewCard(CardDesignSpade, 3, true),
		NewCard(CardDesignHeart, 8, true),
	}
	assert.Less(t, mustWildRank(evalFiveCardHandWithWilds(hand3, wild)), PokerHandFiveOfAKind)
}

// **ワイルドはゲームからプレイヤーへ配られること。**ゲームだけが知っていると、
// 評価器はワイルドを見ないまま動く。
func TestFollowTheQueen_TheWildIsPublishedToEveryPlayer(t *testing.T) {
	s := newTestFollowTheQueen()
	require.NotEmpty(t, s.GetPlayers())

	s.noteUpCard(NewCard(CardDesignSpade, FollowTheQueenQueenValue, true))
	s.noteUpCard(NewCard(CardDesignHeart, 7, true))

	for i, p := range s.GetPlayers() {
		assert.Equal(t, 7, p.GetWildRank(), "player %d", i)
		assert.True(t, p.IsWild(NewCard(CardDesignClover, 7, false)), "player %d", i)
	}

	// 負のコントロール: 移ったら全員に移る。1 人だけ古いままにならない。
	s.noteUpCard(NewCard(CardDesignClover, FollowTheQueenQueenValue, true))
	s.noteUpCard(NewCard(CardDesignDiamond, 4, true))
	for i, p := range s.GetPlayers() {
		assert.Equal(t, 4, p.GetWildRank(), "player %d", i)
	}
}

// **ワイルドは KV の往復も生き延びること。**Worker はリクエストごとに盤を
// 組み直すので、載っていない項は毎回消える ── 復元した盤だけワイルドが効かない。
func TestFollowTheQueen_TheWildSurvivesTheKVRoundTrip(t *testing.T) {
	s := newTestFollowTheQueen()
	s.noteUpCard(NewCard(CardDesignSpade, FollowTheQueenQueenValue, true))
	s.noteUpCard(NewCard(CardDesignHeart, 7, true))
	require.Equal(t, 7, s.GetPlayers()[0].GetWildRank())

	data, err := json.Marshal(s.GetPlayers()[0])
	require.NoError(t, err)
	restored := NewFollowTheQueenPlayer(false, 0)
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, 7, restored.GetWildRank())
	assert.True(t, restored.IsWild(NewCard(CardDesignClover, 7, false)))
}

// **盤そのものの往復も見る。** 上はプレイヤー 1 人分で、ゲーム側の wildRank /
// queenPending が JSON に載っているかは一切見ていなかった。Worker は
// リクエストごとに状態を持たないので、ここから落ちると **本番だけ**で壊れる:
// 盤の途中で「ワイルドはまだ無し」に戻る。手元の CLI は 1 プロセスなので出ない。
func TestFollowTheQueen_TheBoardsWildSurvivesTheKVRoundTrip(t *testing.T) {
	s := dealtBoard(t)
	stackUpcoming(t, s,
		NewCard(CardDesignSpade, FollowTheQueenQueenValue, false),
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignHeart, 4, false),
	)
	s.dealStreetCard(true)
	require.Equal(t, 7, s.GetWildRank())

	data, err := json.Marshal(s)
	require.NoError(t, err)
	restored := NewDefaultFollowTheQueen()
	require.NoError(t, json.Unmarshal(data, restored))

	assert.Equal(t, 7, restored.GetWildRank(), "盤のワイルドが往復で消えている")
	assert.True(t, restored.IsWild(NewCard(CardDesignClover, 7, false)))
}

// **保留中の Q も往復させる。** ストリートの最後の表向き札が Q だった場合、
// 次のストリートの 1 枚目がワイルドを決める。その「保留」が JSON に載っていないと、
// Worker では **ストリートをまたいだ瞬間にワイルドが動かなくなる** —— しかも
// wildRank だけを見ている検査は全部通ってしまう (往復の前後で 0 のままなので)。
func TestFollowTheQueen_APendingQueenSurvivesTheKVRoundTrip(t *testing.T) {
	s := dealtBoard(t)

	// このストリートの表向きを全員 Q で終える = 最後の 1 枚が Q。
	stackUpcoming(t, s,
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignHeart, 4, false),
		NewCard(CardDesignDiamond, 5, false),
		NewCard(CardDesignClover, FollowTheQueenQueenValue, false),
	)
	s.dealStreetCard(true)
	require.Equal(t, 0, s.GetWildRank(), "Q を出した直後はまだワイルド未確定")

	data, err := json.Marshal(s)
	require.NoError(t, err)
	restored := NewDefaultFollowTheQueen()
	require.NoError(t, json.Unmarshal(data, restored))

	// **復元した盤で配りを続ける。** フィールドを覗くのではなく、次の 1 枚が
	// ワイルドを取ることを実際に起こさせる。
	stackUpcoming(t, restored,
		NewCard(CardDesignSpade, 8, false),
		NewCard(CardDesignHeart, 9, false),
		NewCard(CardDesignDiamond, 10, false),
		NewCard(CardDesignClover, 6, false),
	)
	restored.dealStreetCard(true)

	assert.Equal(t, 8, restored.GetWildRank(),
		"保留中の Q が往復で失われ、次の札がワイルドを取っていない")
}

// **ワイルド 5 枚**（Q 4 枚 + ワイルドランク 1 枚）でもファイブカード。
// 固定札が 1 枚も無いので、「残り 1 枚に合わせる」理屈が使えない形。
func TestFollowTheQueen_FiveWildsAreStillFiveOfAKind(t *testing.T) {
	wild := func(c *Card) bool { return true }
	hand := []*Card{
		NewCard(CardDesignSpade, FollowTheQueenQueenValue, true),
		NewCard(CardDesignHeart, FollowTheQueenQueenValue, true),
		NewCard(CardDesignClover, FollowTheQueenQueenValue, true),
		NewCard(CardDesignDiamond, FollowTheQueenQueenValue, true),
		NewCard(CardDesignSpade, 7, true),
	}
	assert.Equal(t, PokerHandFiveOfAKind, mustWildRank(evalFiveCardHandWithWilds(hand, wild)))
}

// **全員オールインで畳むときに 8 枚配っていた。** `advancePhase` は先に
// `s.phase` を次のストリートへ進めてその 1 枚を配り、そのあと
// `dealRemainingStreets` を呼ぶ。ところが同関数のループが `s.phase` から
// 始まっていたので、**いま配ったストリートをもう一度配る**。
//
// このゲームでは素のスタッドより悪い: 余分な表向き札が `noteUpCard` を通るので、
// **ベッティングが全部終わったあとにワイルドが動く**。
func TestFollowTheQueen_AllInShowdownDealsSevenCardsNotEight(t *testing.T) {
	s := dealtBoard(t)
	s.SetPhase(FollowTheQueenPhaseThirdStreet)

	// 1 人を残して全員オールイン → advancePhase がショーダウンまで畳む。
	for i, p := range s.players {
		if i > 0 {
			p.SetAllIn(true)
		}
	}
	s.advancePhase()

	for i, p := range s.players {
		hole, door := len(p.GetHoleCards()), len(p.GetDoorCards())
		assert.Equal(t, 7, hole+door,
			"player %d は 7 枚のはず (hole=%d door=%d)", i, hole, door)
		assert.Equal(t, 4, door, "player %d の表向きは 4 枚のはず", i)
	}
}
