//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func bolCard(design, value int) *Card { return NewCard(design, value, true) }
func bolJoker() *Card                 { return NewCard(CardDesignJoker, CardValueJoker, true) }

func newBoliviaGame(t *testing.T) *Bolivia {
	t.Helper()
	players := make([]*BoliviaPlayer, 0, BoliviaPlayerCnt)
	for i := 0; i < BoliviaPlayerCnt; i++ {
		players = append(players, NewBoliviaPlayer(i == 0, i%BoliviaTeamCnt))
	}
	g := NewBolivia(newBoliviaDeck(), players, DefaultBoliviaConfig())
	g.Reset()
	return g
}

// **ボリビアはワイルドだけのメルドを認める。** ここがサンバとの唯一無二の差。
func TestBolivia_WildOnlyMeldIsLegal(t *testing.T) {
	g := newBoliviaGame(t)
	wilds := []*Card{bolCard(CardDesignSpade, 2), bolJoker(), bolCard(CardDesignHeart, 2)}
	assert.NoError(t, g.validateNewSet(wilds), "ワイルドだけの組が断られている")

	// 負のコントロール: ナチュラルが 1 枚だけ混じった組は通らない
	// (ワイルド専用でもセットでもない中途半端な形)。
	mixed := []*Card{bolCard(CardDesignSpade, 2), bolJoker(), bolCard(CardDesignHeart, 5)}
	assert.Error(t, g.validateNewSet(mixed), "ナチュラル 1 枚の混成が通ってしまっている")
}

func TestBoliviaIsWildOnly(t *testing.T) {
	assert.True(t, BoliviaIsWildOnly([]*Card{bolCard(CardDesignSpade, 2), bolJoker()}))
	assert.False(t, BoliviaIsWildOnly([]*Card{bolCard(CardDesignSpade, 2), bolCard(CardDesignSpade, 5)}))
	// **空はメルドではない。** 何も無いものを「全部ワイルド」と読まない。
	assert.False(t, BoliviaIsWildOnly(nil))
	assert.False(t, BoliviaIsWildOnly([]*Card{}))
}

// **7 枚のワイルドメルドが「ボリビア」。** ゲーム名の由来で 2500 点。
func TestBolivia_SevenWildsIsABoliviaWorth2500(t *testing.T) {
	m := &BoliviaMeld{Kind: BoliviaMeldWild, Cards: []*Card{
		bolCard(CardDesignSpade, 2), bolCard(CardDesignHeart, 2), bolCard(CardDesignClover, 2),
		bolCard(CardDesignDiamond, 2), bolJoker(), bolJoker(), bolJoker()}}
	assert.True(t, m.IsBoliviaCanasta())
	assert.True(t, m.IsCompleted())
	// 6 枚ではまだボリビアではない。
	short := &BoliviaMeld{Kind: BoliviaMeldWild, Cards: m.Cards[:6]}
	assert.False(t, short.IsBoliviaCanasta())
	assert.Equal(t, 2500, BoliviaBoliviaBonus)
}

// **エスカレラはワイルド無しの同スート 7 枚連番。** #5465 の「3 だけで作る
// メルド」ではない ── スペイン語の「梯子」で、3 とは何の関係も無い。
func TestBolivia_EscaleraIsASuitedSequenceNotAMeldOfThrees(t *testing.T) {
	esc := &BoliviaMeld{Kind: BoliviaMeldEscalera, Cards: []*Card{
		bolCard(CardDesignSpade, 4), bolCard(CardDesignSpade, 5), bolCard(CardDesignSpade, 6),
		bolCard(CardDesignSpade, 7), bolCard(CardDesignSpade, 8), bolCard(CardDesignSpade, 9),
		bolCard(CardDesignSpade, 10)}}
	assert.True(t, esc.IsEscalera())
	assert.Equal(t, 1500, BoliviaEscaleraBonus)

	// **3 はメルドの材料ではない。** 黒 3 はそもそもメルドできず、赤 3 は
	// 場に置かれてボーナスになるだけ。
	g := newBoliviaGame(t)
	threes := []*Card{bolCard(CardDesignSpade, 3), bolCard(CardDesignClover, 3), bolCard(CardDesignSpade, 3)}
	assert.Error(t, g.validateNewSet(threes), "黒 3 のセットが通ってしまっている")
	assert.True(t, BoliviaIsBlack3(bolCard(CardDesignSpade, 3)))
	assert.True(t, BoliviaIsRed3(bolCard(CardDesignHeart, 3)))
}

// **完成したボリビアには足せない。**
func TestBolivia_CannotAddToACompletedBolivia(t *testing.T) {
	g := newBoliviaGame(t)
	p := g.players[0]
	full := &BoliviaMeld{Kind: BoliviaMeldWild, Cards: []*Card{
		bolCard(CardDesignSpade, 2), bolCard(CardDesignHeart, 2), bolCard(CardDesignClover, 2),
		bolCard(CardDesignDiamond, 2), bolJoker(), bolJoker(), bolJoker()}}
	p.SetMelds([]*BoliviaMeld{full})

	_, err := g.resolveMeldGroup(0, []*Card{bolJoker()})
	assert.Error(t, err, "完成したボリビアに足せてしまっている")

	// 負のコントロール: 未完成なら足せる。
	p.SetMelds([]*BoliviaMeld{{Kind: BoliviaMeldWild, Cards: full.Cards[:4]}})
	res, err := g.resolveMeldGroup(0, []*Card{bolJoker()})
	require.NoError(t, err)
	assert.False(t, res.isNew)
	assert.Equal(t, BoliviaMeldWild, res.kind)
}

// **ワイルドだけの組はセットと取り違えない。**
//
// 先にセット判定に落ちると、ワイルドは読み飛ばされて「同ランク」に見えるので、
// 7 枚揃ってもボリビア (2500) ではなくミックスカナスタ (300) として数えられる。
func TestBolivia_WildGroupResolvesAsWildNotAsSet(t *testing.T) {
	g := newBoliviaGame(t)
	res, err := g.resolveMeldGroup(0, []*Card{
		bolCard(CardDesignSpade, 2), bolJoker(), bolCard(CardDesignHeart, 2)})
	require.NoError(t, err)
	assert.True(t, res.isNew)
	assert.Equal(t, BoliviaMeldWild, res.kind, "ワイルドの組がセットとして解決されている")
}

// **上がるには完成メルド 2 つでは足りない ── 最低 1 つはエスカレラ。**
// #5465 は「規定数のカナスタ」としか書いておらず、いちばん効く縛りを落としている。
func TestBolivia_GoingOutRequiresAnEscalera(t *testing.T) {
	g := newBoliviaGame(t)
	seven := func(design, from int) []*Card {
		out := make([]*Card, 0, 7)
		for i := 0; i < 7; i++ {
			out = append(out, bolCard(design, from+i))
		}
		return out
	}
	sevenOfARank := func(v int) []*Card {
		out := make([]*Card, 0, 7)
		for i := 0; i < 7; i++ {
			out = append(out, bolCard(CardDesignSpade, v))
		}
		return out
	}

	// カナスタ 2 つだけ ── 数は足りているが上がれない。
	g.players[0].SetMelds([]*BoliviaMeld{
		{Kind: BoliviaMeldSet, Cards: sevenOfARank(5), IsNatural: true},
		{Kind: BoliviaMeldSet, Cards: sevenOfARank(9), IsNatural: true},
	})
	g.players[2].SetMelds(nil)
	require.Equal(t, 2, g.teamCompletedCount(g.players[0].team), "完成メルドが 2 つになっていない")
	assert.False(t, g.canGoOut(0), "エスカレラ無しで上がれてしまっている")

	// パートナーがエスカレラを持てば上がれる (チームで数える)。
	g.players[2].SetMelds([]*BoliviaMeld{
		{Kind: BoliviaMeldEscalera, Cards: seven(CardDesignHeart, 4), IsNatural: true},
	})
	assert.True(t, g.canGoOut(0), "エスカレラがあるのに上がれない")

	// 相手チームのエスカレラでは上がれない。
	g.players[2].SetMelds(nil)
	g.players[1].SetMelds([]*BoliviaMeld{
		{Kind: BoliviaMeldEscalera, Cards: seven(CardDesignHeart, 4), IsNatural: true},
	})
	assert.False(t, g.canGoOut(0), "相手のエスカレラで上がれてしまっている")
}

// **目標はサンバの 10000 ではなく 15000。**
func TestBolivia_TargetScoreIs15000(t *testing.T) {
	assert.Equal(t, 15000, BoliviaDefaultPointLimit)
	assert.Equal(t, 15000, DefaultBoliviaConfig().PointLimit)
}

// **デッキは 3 組 + ジョーカー 6 枚 = 162 枚。** #5465 の「52×2 + ジョーカー」ではない。
func TestBolivia_UsesThreeDecksAndSixJokers(t *testing.T) {
	deck := newBoliviaDeck()
	assert.Equal(t, CardCnt*3+6, deck.GetTotalCount(), "162 枚になっていない")
}

// boliviaPlayOut は 4 席とも CPU の方針で 1 ラウンド打ち切る。
func boliviaPlayOut(t *testing.T) *Bolivia {
	t.Helper()
	players := make([]*BoliviaPlayer, 0, BoliviaPlayerCnt)
	for k := 0; k < BoliviaPlayerCnt; k++ {
		players = append(players, NewBoliviaPlayer(false, k%BoliviaTeamCnt))
	}
	g := NewBolivia(newBoliviaDeck(), players, DefaultBoliviaConfig())
	g.Reset()
	for guard := 0; guard < 4000; guard++ {
		if g.phase == BoliviaPhaseRoundEnd || g.phase == BoliviaPhaseGameEnd || g.gameEndFlag {
			break
		}
		g.CpuPlay()
	}
	return g
}

// **場に 3 枚未満のメルドが残らないこと。**
//
// クローン元の Samba は、同ランク 4 枚以上を「3 枚 + あまり」に割って提案する
// ため、あまりが 1〜2 枚の新規メルドとして場に残る ── 実測で 1405 個中 138 個。
// 同じ前提を引き継がないよう、同ランクは 1 つの組にまとめて出す。
func TestBolivia_NoMeldIsEverShorterThanThree(t *testing.T) {
	tiny, total := 0, 0
	for i := 0; i < 20; i++ {
		g := boliviaPlayOut(t)
		for _, p := range g.players {
			for _, m := range p.GetMelds() {
				total++
				if len(m.Cards) < 3 {
					tiny++
				}
			}
		}
	}
	require.Greater(t, total, 100, "メルドがほとんど作られていない -- 測っていない")
	assert.Equal(t, 0, tiny, "3 枚未満のメルドが %d 個ある (全 %d 個)", tiny, total)
}

// **エスカレラとボリビアが実際に完成すること。**
//
// 上がるにはエスカレラが要るので、CPU が連番になる札をセットに食わせていると
// チームは永久に上がれない ── 最初の実装では 60 局すべてがそうなっていた。
func TestBolivia_SignatureMeldsActuallyGetCompleted(t *testing.T) {
	escaleras, bolivias := 0, 0
	for i := 0; i < 30; i++ {
		g := boliviaPlayOut(t)
		for _, p := range g.players {
			for _, m := range p.GetMelds() {
				if m.IsEscalera() {
					escaleras++
				}
				if m.IsBoliviaCanasta() {
					bolivias++
				}
			}
		}
	}
	assert.Positive(t, escaleras, "30 局打ってエスカレラが 1 本も完成していない")
	assert.Positive(t, bolivias, "30 局打ってボリビアが 1 つも完成していない")
}

// **保存して読み戻してもボリビアがボリビアのままであること (レビュー指摘)。**
//
// 復元時の Kind 許可リストはクローン元の 2 種類のままだったので、ワイルドの
// メルドは黙ってセットに書き換えられていた。Worker は毎リクエストで復元する
// ので、2500 点のメルドが次の 1 手で 300 点になる。
func TestBolivia_WildMeldSurvivesAJSONRoundTrip(t *testing.T) {
	p := NewBoliviaPlayer(true, 0)
	wild := &BoliviaMeld{Kind: BoliviaMeldWild, Cards: []*Card{
		bolCard(CardDesignSpade, 2), bolCard(CardDesignHeart, 2), bolCard(CardDesignClover, 2),
		bolCard(CardDesignDiamond, 2), bolJoker(), bolJoker(), bolJoker()}}
	esc := &BoliviaMeld{Kind: BoliviaMeldEscalera, IsNatural: true, Cards: []*Card{
		bolCard(CardDesignHeart, 4), bolCard(CardDesignHeart, 5), bolCard(CardDesignHeart, 6),
		bolCard(CardDesignHeart, 7), bolCard(CardDesignHeart, 8), bolCard(CardDesignHeart, 9),
		bolCard(CardDesignHeart, 10)}}
	p.SetMelds([]*BoliviaMeld{wild, esc})
	require.True(t, p.HasBolivia())
	require.True(t, p.HasEscalera())

	data, err := json.Marshal(p)
	require.NoError(t, err)
	restored := NewBoliviaPlayer(true, 0)
	require.NoError(t, json.Unmarshal(data, restored))

	melds := restored.GetMelds()
	require.Len(t, melds, 2)
	assert.Equal(t, BoliviaMeldWild, melds[0].Kind, "ボリビアがセットに化けている")
	assert.True(t, melds[0].IsBoliviaCanasta(), "復元後にボリビアと認識されない")
	assert.Equal(t, BoliviaMeldEscalera, melds[1].Kind)
	assert.True(t, restored.HasBolivia(), "復元後に HasBolivia が落ちている")
	assert.True(t, restored.HasEscalera())

	// 負のコントロール: 知らない Kind は今までどおりセットに丸める。
	bogus := NewBoliviaPlayer(true, 0)
	require.NoError(t, json.Unmarshal([]byte(`{"me":[{"kd":99,"ca":[]}]}`), bogus))
	if got := bogus.GetMelds(); len(got) > 0 {
		assert.Equal(t, BoliviaMeldSet, got[0].Kind)
	}
}

// boliviaHumanTurn は人間の手番 (ドローフェーズ) の盤を返す。
func boliviaHumanTurn(t *testing.T) *Bolivia {
	t.Helper()
	g := newBoliviaGame(t)
	g.currentPlayerIdx = 0
	g.phase = BoliviaPhaseDraw
	return g
}

// boliviaSeven は同スート連番 7 枚を返す。
func boliviaSeven(design, from int) []*Card {
	out := make([]*Card, 0, 7)
	for i := 0; i < 7; i++ {
		out = append(out, bolCard(design, from+i))
	}
	return out
}

// **人間の入口を実際に通す。**
//
// 規則を直接叩くテストだけでは、公開 API がその規則に繋がっているかを
// 誰も見ていない ── ドロー・メルド・ディスカードの入口はどれも 0% だった。
func TestBolivia_HumanTurnRunsThroughEveryPhase(t *testing.T) {
	g := boliviaHumanTurn(t)
	before := g.players[0].GetCardsSize()

	assert.NoError(t, g.PlayerDrawFromStock())
	assert.Equal(t, before+1, g.players[0].GetCardsSize())
	assert.Equal(t, BoliviaPhaseMeld, g.GetPhase())

	// 引く番でないのに二度引けない。
	assert.Error(t, g.PlayerDrawFromStock())

	assert.NoError(t, g.PlayerSkipMeld())
	assert.Equal(t, BoliviaPhaseDiscard, g.GetPhase())

	assert.NoError(t, g.PlayerDiscard(0))
	assert.Equal(t, before, g.players[0].GetCardsSize())
	// 手番が次の席へ渡っている。
	assert.NotEqual(t, 0, g.GetCurrentPlayerIdx())
}

func TestBolivia_HumanActionsRejectTheWrongPhase(t *testing.T) {
	g := boliviaHumanTurn(t)
	// ドローフェーズではメルドも捨て札も上がりもできない。
	assert.Error(t, g.PlayerMeld([][]int{{0, 1, 2}}))
	assert.Error(t, g.PlayerSkipMeld())
	assert.Error(t, g.PlayerDiscard(0))
	assert.Error(t, g.PlayerGoOut())

	// 範囲外の手札番号は断る。
	require.NoError(t, g.PlayerDrawFromStock())
	require.NoError(t, g.PlayerSkipMeld())
	assert.Error(t, g.PlayerDiscard(-1))
	assert.Error(t, g.PlayerDiscard(999))
}

// **断る理由を取り違えない。**
//
// 完成メルドが足りないのと、エスカレラが無いのは直し方がまったく違う。
// 「カナスタが N 個要る」とだけ言われた側は、カナスタを増やし続けて
// 永久に上がれない。
func TestBolivia_GoOutRefusalNamesTheRealReason(t *testing.T) {
	sevenOfARank := func(v int) []*Card {
		out := make([]*Card, 0, 7)
		for i := 0; i < 7; i++ {
			out = append(out, bolCard(CardDesignSpade, v))
		}
		return out
	}

	t.Run("too few melds says so", func(t *testing.T) {
		g := boliviaHumanTurn(t)
		require.NoError(t, g.PlayerDrawFromStock())
		require.NoError(t, g.PlayerSkipMeld())
		err := g.PlayerGoOut()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "完成メルド")
		assert.NotContains(t, err.Error(), "エスカレラ")
	})

	// **数は足りているがエスカレラが無い**とき、その一点を名指すこと。
	t.Run("enough melds but no escalera says escalera", func(t *testing.T) {
		g := boliviaHumanTurn(t)
		g.players[0].SetMelds([]*BoliviaMeld{
			{Kind: BoliviaMeldSet, IsNatural: true, Cards: sevenOfARank(5)},
			{Kind: BoliviaMeldSet, IsNatural: true, Cards: sevenOfARank(9)},
		})
		require.Equal(t, 2, g.teamCompletedCount(0))
		require.NoError(t, g.PlayerDrawFromStock())
		require.NoError(t, g.PlayerSkipMeld())

		err := g.PlayerGoOut()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "エスカレラ", "エスカレラが無いことを言っていない")

		// 負のコントロール: パートナーがエスカレラを持てば上がれる。
		// **上がりは手札を出し切ってから**なので、手札も 1 枚まで減らす。
		g.players[2].SetMelds([]*BoliviaMeld{
			{Kind: BoliviaMeldEscalera, IsNatural: true, Cards: boliviaSeven(CardDesignHeart, 4)},
		})
		for g.players[0].GetCardsSize() > 1 {
			g.players[0].RemoveCard(0)
		}
		assert.NoError(t, g.PlayerGoOut(), "エスカレラを足しても上がれない")
	})
}

// **エスカレラの検証はワイルドを拒む。**
func TestBolivia_EscaleraValidationRejectsWildsAndBrokenRuns(t *testing.T) {
	g := newBoliviaGame(t)
	assert.NoError(t, g.validateNewEscalera([]*Card{
		bolCard(CardDesignSpade, 4), bolCard(CardDesignSpade, 5), bolCard(CardDesignSpade, 6)}))
	assert.Error(t, g.validateNewEscalera([]*Card{
		bolCard(CardDesignSpade, 4), bolJoker(), bolCard(CardDesignSpade, 6)}),
		"エスカレラにワイルドが入ってしまっている")
	assert.Error(t, g.validateNewEscalera([]*Card{
		bolCard(CardDesignSpade, 4), bolCard(CardDesignHeart, 5), bolCard(CardDesignSpade, 6)}),
		"スートが混ざったのに通っている")
	assert.Error(t, g.validateNewEscalera([]*Card{
		bolCard(CardDesignSpade, 4), bolCard(CardDesignSpade, 6), bolCard(CardDesignSpade, 8)}),
		"連番でないのに通っている")
	assert.Error(t, g.validateNewEscalera([]*Card{
		bolCard(CardDesignSpade, 4), bolCard(CardDesignSpade, 5)}), "2 枚で通っている")
}

// **旗と取り分は別の面。** `IsBoliviaCanasta()` が真になることと、その 2500 点が
// 実際にチームの得点に載ることは違う ── 片方だけ直すと、画面に「ボリビア」と
// 出ているのに点が動かない。
func TestBolivia_ScoringPaysEachMeldKindItsOwnBonus(t *testing.T) {
	sevenOfARank := func(v int) []*Card {
		out := make([]*Card, 0, 7)
		for i := 0; i < 7; i++ {
			out = append(out, bolCard(CardDesignSpade, v))
		}
		return out
	}
	wildSeven := []*Card{
		bolCard(CardDesignSpade, 2), bolCard(CardDesignHeart, 2), bolCard(CardDesignClover, 2),
		bolCard(CardDesignDiamond, 2), bolJoker(), bolJoker(), bolJoker()}

	// **配らない盤で測る。** `Reset()` を通すと配りが混ざり、そのぶんの
	// 変動が加点の逆算に乗って 3 回に 1 回落ちた ── 実測。この試験が見たいのは
	// 「メルド 1 つにいくら付くか」だけなので、卓は手で組む。
	roundScoreFor := func(t *testing.T, meld *BoliviaMeld) int {
		t.Helper()
		players := make([]*BoliviaPlayer, 0, BoliviaPlayerCnt)
		for i := 0; i < BoliviaPlayerCnt; i++ {
			p := NewBoliviaPlayer(i == 0, i%BoliviaTeamCnt)
			p.SetHasInitMeld(true) // 赤3の減算を避ける (枚数は 0 なので影響しない)
			players = append(players, p)
		}
		g := NewBolivia(newBoliviaDeck(), players, DefaultBoliviaConfig())

		cards := 0
		if meld != nil {
			players[0].SetMelds([]*BoliviaMeld{meld})
			for _, c := range meld.Cards {
				cards += BoliviaCardValue(c)
			}
		}
		before := g.GetTeamScore(0)
		g.scoreRound(-1, 0)
		return g.GetTeamScore(0) - before - cards
	}

	assert.Equal(t, BoliviaNaturalCanastaBonus,
		roundScoreFor(t, &BoliviaMeld{Kind: BoliviaMeldSet, IsNatural: true, Cards: sevenOfARank(5)}),
		"ナチュラルカナスタの加点が出ていない")
	assert.Equal(t, BoliviaEscaleraBonus,
		roundScoreFor(t, &BoliviaMeld{Kind: BoliviaMeldEscalera, IsNatural: true, Cards: boliviaSeven(CardDesignHeart, 4)}),
		"エスカレラの 1500 点が出ていない")
	// **ここが本命。** ゲーム名になっている役がいちばん重い。
	assert.Equal(t, BoliviaBoliviaBonus,
		roundScoreFor(t, &BoliviaMeld{Kind: BoliviaMeldWild, Cards: wildSeven}),
		"ボリビアの 2500 点が出ていない")

	// 負のコントロール: 6 枚では未完成なので加点は無い。
	assert.Equal(t, 0,
		roundScoreFor(t, &BoliviaMeld{Kind: BoliviaMeldWild, Cards: wildSeven[:6]}),
		"未完成のワイルドメルドに加点が付いている")
	assert.Equal(t, 0, roundScoreFor(t, nil))

	// **重さの順序**: ボリビア > エスカレラ > ナチュラルカナスタ。
	assert.Greater(t, BoliviaBoliviaBonus, BoliviaEscaleraBonus)
	assert.Greater(t, BoliviaEscaleraBonus, BoliviaNaturalCanastaBonus)
}
