//go:build test && (!js || !wasm || casino)

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// tcrHand builds a hand from (design, value) pairs.
func tcrHand(spec ...int) []*Card {
	cards := make([]*Card, 0, len(spec)/2)
	for i := 0; i < len(spec); i += 2 {
		cards = append(cards, NewCard(spec[i], spec[i+1], true))
	}
	return cards
}

func TestThreeCardRummyScore_CountsFaceCardsAsTenAndAceAsOne(t *testing.T) {
	// K♠ Q♥ A♦ -- 10 + 10 + 1. **A は 11 にならない。** 揃いでも連番でもないので
	// 素直な合計になる。
	assert.Equal(t, 21, ThreeCardRummyScore(tcrHand(0, 13, 1, 12, 2, 1)))
	// J は 10、10 も 10。
	assert.Equal(t, 25, ThreeCardRummyScore(tcrHand(0, 11, 1, 10, 2, 5)))
	// 数札はそのまま。
	assert.Equal(t, 9, ThreeCardRummyScore(tcrHand(0, 2, 1, 3, 2, 4)))
}

func TestThreeCardRummyScore_MeldsScoreZero(t *testing.T) {
	// 同ランク3枚: 素直に足すと 30 点で最悪の手になってしまう。
	assert.Equal(t, ThreeCardRummyPerfectScore, ThreeCardRummyScore(tcrHand(0, 13, 1, 13, 2, 13)))
	// 同スート連番: 2-3-4 は合計 9 だが役なので 0。
	assert.Equal(t, ThreeCardRummyPerfectScore, ThreeCardRummyScore(tcrHand(0, 2, 0, 3, 0, 4)))
	// A-2-3 (A を下端に付ける)。
	assert.Equal(t, ThreeCardRummyPerfectScore, ThreeCardRummyScore(tcrHand(1, 1, 1, 2, 1, 3)))
	// Q-K-A (A を上端に付ける) -- 素直に足すと 21 点。
	assert.Equal(t, ThreeCardRummyPerfectScore, ThreeCardRummyScore(tcrHand(2, 12, 2, 13, 2, 1)))
}

func TestThreeCardRummyScore_NearMissesAreNotMelds(t *testing.T) {
	// 連番だがスートがばらけている -- 役ではない。
	assert.Equal(t, 9, ThreeCardRummyScore(tcrHand(0, 2, 1, 3, 2, 4)))
	// 同スートだが連番ではない (2-3-5)。
	assert.Equal(t, 10, ThreeCardRummyScore(tcrHand(0, 2, 0, 3, 0, 5)))
	// K-A-2 は「輪」にはならない。A は両端に付くが、跨いだ並びは認めない。
	assert.Equal(t, 13, ThreeCardRummyScore(tcrHand(0, 13, 0, 1, 0, 2)))
	// 同ランク2枚だけ -- 揃いではない。
	assert.Equal(t, 20, ThreeCardRummyScore(tcrHand(0, 5, 1, 5, 2, 13)))
	// J-Q-K は同スートなら役 (0 点)、スート違いなら 30 点。
	assert.Equal(t, ThreeCardRummyPerfectScore, ThreeCardRummyScore(tcrHand(0, 11, 0, 12, 0, 13)))
	assert.Equal(t, 30, ThreeCardRummyScore(tcrHand(0, 11, 1, 12, 0, 13)))
}

func TestThreeCardRummyScore_PartialHandIsNotAMeld(t *testing.T) {
	// **3 枚未満で 0 点を返すと「最強の手」として扱われる。** 揃っていない手は
	// 役ではなく、見えている分の合計を返す。
	assert.Equal(t, 20, ThreeCardRummyScore(tcrHand(0, 13, 0, 13)))
	assert.Equal(t, 0, ThreeCardRummyScore(nil))
	// nil 混じり (配り途中) も役にしない。
	assert.Equal(t, 20, ThreeCardRummyScore([]*Card{tcrHand(0, 13)[0], nil, tcrHand(1, 13)[0]}))
}

func TestThreeCardRummyScore_ShuffledMeldOrderStillCounts(t *testing.T) {
	// 手札の並び順に依存しない -- 連番判定はソートしてから見る。
	assert.Equal(t, ThreeCardRummyPerfectScore, ThreeCardRummyScore(tcrHand(0, 4, 0, 2, 0, 3)))
	assert.Equal(t, ThreeCardRummyPerfectScore, ThreeCardRummyScore(tcrHand(2, 1, 2, 13, 2, 12)))
}
