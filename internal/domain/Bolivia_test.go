//go:build test

package domain

import (
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
