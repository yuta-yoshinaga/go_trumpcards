//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func aluetteCard(design, value int) *Card { return NewCard(design, value, false) }

func aluetteTrick(entries ...[2]int) []*TrickCard {
	trick := make([]*TrickCard, 0, len(entries))
	for seat, e := range entries {
		trick = append(trick, &TrickCard{PlayerIdx: seat, Card: aluetteCard(e[0], e[1])})
	}
	return trick
}

func TestAluette_DeckIsFortyEightDistinctCards(t *testing.T) {
	deck := buildAluetteDeck()
	assert.Len(t, deck, AluetteDeckSize)
	assert.Equal(t, 48, AluetteDeckSize)

	seen := map[[2]int]bool{}
	for _, c := range deck {
		key := [2]int{c.GetDesign(), c.GetValue()}
		assert.False(t, seen[key], "duplicate card %v", key)
		seen[key] = true
		// **8 と 9 を含む。**40 枚デッキ (Tute/Briscola) の感覚で作ると、
		// 9 のうち 2 枚がリュエットなので序列表が壊れる。
		assert.NotEqual(t, 10, c.GetValue(), "the Latin deck has no 10")
	}
	for _, v := range []int{8, 9} {
		assert.True(t, seen[[2]int{1, v}], "value %d must exist in every suit", v)
	}
}

// **これがこのゲームの核心。**強さはランクではなく「特定の 1 枚」で決まる。
// 同じ value 3 でも、金貨の3 (Monsieur) は最強で、剣の3 は普通の 3 でしかない。
func TestAluette_RankIsPerCardNotPerValue(t *testing.T) {
	monsieur := aluetteCard(4, 3)   // 金貨の3
	plainThree := aluetteCard(1, 3) // 剣の3
	assert.Equal(t, "Monsieur", AluetteLuetteName(monsieur))
	assert.Empty(t, AluetteLuetteName(plainThree))
	assert.Greater(t, AluetteRank(monsieur), AluetteRank(plainThree),
		"the same value must not imply the same strength")
}

// リュエット 6 枚は名前の順に最上位を占める。
func TestAluette_LuettesOutrankEverythingInOrder(t *testing.T) {
	order := []*Card{
		aluetteCard(4, 3), // Monsieur
		aluetteCard(3, 3), // Madame
		aluetteCard(3, 2), // Borgne
		aluetteCard(4, 2), // Vache
		aluetteCard(3, 9), // GrandNeuf
		aluetteCard(4, 9), // PetitNeuf
	}
	names := []string{"Monsieur", "Madame", "Borgne", "Vache", "GrandNeuf", "PetitNeuf"}
	for i, c := range order {
		assert.Equal(t, names[i], AluetteLuetteName(c))
		if i > 0 {
			assert.Less(t, AluetteRank(c), AluetteRank(order[i-1]),
				"%s must be weaker than %s", names[i], names[i-1])
		}
	}
	// 最弱のリュエットでも、最強の通常札 (剣の3) より強い。
	assert.Greater(t, AluetteRank(order[len(order)-1]), AluetteRank(aluetteCard(1, 3)))
}

// 通常札はスートを一切見ず、値だけで 3 > 2 > A > 王 > 騎 > 従 > 9 > 8 … と並ぶ。
func TestAluette_OrdinaryOrderIgnoresSuit(t *testing.T) {
	desc := []int{3, 2, 1, 13, 12, 11, 9, 8, 7, 6, 5, 4}
	for i := 1; i < len(desc); i++ {
		// 剣 (design 1) はリュエットを 1 枚も含まないので、通常序列の確認に使える。
		hi, lo := aluetteCard(1, desc[i-1]), aluetteCard(1, desc[i])
		assert.Greater(t, AluetteRank(hi), AluetteRank(lo),
			"value %d must outrank %d", desc[i-1], desc[i])
	}
	// スートが違っても同値なら同ランク。
	assert.Equal(t, AluetteRank(aluetteCard(1, 7)), AluetteRank(aluetteCard(2, 7)))
}

func TestAluette_TrickWinner(t *testing.T) {
	cases := []struct {
		name  string
		trick []*TrickCard
		want  int
	}{
		{"the strongest card wins regardless of what was led", aluetteTrick(
			[2]int{1, 4}, [2]int{2, 13}, [2]int{1, 5}, [2]int{2, 6},
		), 1},
		{"a luette beats every ordinary card", aluetteTrick(
			[2]int{1, 3}, [2]int{2, 3}, [2]int{4, 9}, [2]int{1, 2},
		), 2},
		{"Monsieur beats Madame", aluetteTrick(
			[2]int{3, 3}, [2]int{4, 3}, [2]int{1, 1}, [2]int{2, 1},
		), 1},
		// **同ランクは先に出した方が勝つ。**後から同じ強さを重ねても奪えない。
		{"an equal rank played later does not take it", aluetteTrick(
			[2]int{1, 7}, [2]int{2, 7}, [2]int{3, 7}, [2]int{4, 7},
		), 0},
		{"an empty trick does not panic", nil, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, aluetteTrickWinnerOf(tc.trick))
		})
	}
}

// 切り札は存在しない。スートは強さに一切関与しない。
func TestAluette_NoTrumpSuitExists(t *testing.T) {
	for suit := 1; suit <= AluetteSuitCnt; suit++ {
		// 同じ 4 (最弱の通常札) はどのスートでも同じ強さ。
		assert.Equal(t, AluetteRank(aluetteCard(1, 4)), AluetteRank(aluetteCard(suit, 4)),
			"suit %d must not confer any strength", suit)
	}
}

func TestAluette_TeamsAreOpposite(t *testing.T) {
	assert.Equal(t, AluetteTeamOf(0), AluetteTeamOf(2))
	assert.Equal(t, AluetteTeamOf(1), AluetteTeamOf(3))
	assert.NotEqual(t, AluetteTeamOf(0), AluetteTeamOf(1))
}

func TestAluette_DealArithmetic(t *testing.T) {
	assert.Equal(t, 5, AluetteHandSize)
	assert.Equal(t, AluetteHandSize, AluetteTrickCount)
	// 5 戦 3 勝。過半数であることを式で固定しておく。
	assert.Equal(t, 3, AluetteTricksToWin)
	assert.Greater(t, AluetteTricksToWin*2, AluetteTrickCount, "winning must require a majority")
	// 配り切らない。残りはそのメーヌでは使わない。
	assert.Less(t, AluettePlayerCnt*AluetteHandSize, AluetteDeckSize)
}

func TestAluette_RankHandlesNil(t *testing.T) {
	assert.Equal(t, -1, AluetteRank(nil))
	assert.Empty(t, AluetteLuetteName(nil))
}
