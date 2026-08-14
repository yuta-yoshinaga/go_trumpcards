//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// tsCard は色と駒から札を作る。
func tsCard(color, piece int) *Card { return NewCard(color, piece, false) }

// **メルドは 3 つの形しかない。** 数字の並びという概念が無いので、
// 標準デッキのラミーにある「同スートの連番」は存在しない。
func TestTuSacClassifyMeld(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		cards []*Card
		want  TuSacMeldKind
	}{
		{
			"同色同種 3 枚",
			[]*Card{
				tsCard(TuSacColorRed, TuSacPieceElephant),
				tsCard(TuSacColorRed, TuSacPieceElephant),
				tsCard(TuSacColorRed, TuSacPieceElephant),
			},
			TuSacMeldSameColorSet,
		},
		{
			"異色の車馬砲",
			[]*Card{
				tsCard(TuSacColorRed, TuSacPieceChariot),
				tsCard(TuSacColorGreen, TuSacPieceHorse),
				tsCard(TuSacColorWhite, TuSacPieceCannon),
			},
			TuSacMeldChariotTrio,
		},
		{
			"卒 5 枚",
			[]*Card{
				tsCard(TuSacColorRed, TuSacPieceSoldier),
				tsCard(TuSacColorGreen, TuSacPieceSoldier),
				tsCard(TuSacColorWhite, TuSacPieceSoldier),
				tsCard(TuSacColorYellow, TuSacPieceSoldier),
				tsCard(TuSacColorRed, TuSacPieceSoldier),
			},
			TuSacMeldSoldierSet,
		},
		{
			// **同色の車馬砲はメルドにならない。** 異色であることが要る。
			"同色の車馬砲",
			[]*Card{
				tsCard(TuSacColorRed, TuSacPieceChariot),
				tsCard(TuSacColorRed, TuSacPieceHorse),
				tsCard(TuSacColorRed, TuSacPieceCannon),
			},
			TuSacMeldNone,
		},
		{
			// **異色でも 3 種そろっていなければならない。**
			"異色だが車車馬",
			[]*Card{
				tsCard(TuSacColorRed, TuSacPieceChariot),
				tsCard(TuSacColorGreen, TuSacPieceChariot),
				tsCard(TuSacColorWhite, TuSacPieceHorse),
			},
			TuSacMeldNone,
		},
		{
			"異色の同種 3 枚",
			[]*Card{
				tsCard(TuSacColorRed, TuSacPieceElephant),
				tsCard(TuSacColorGreen, TuSacPieceElephant),
				tsCard(TuSacColorWhite, TuSacPieceElephant),
			},
			TuSacMeldNone,
		},
		{
			// 將士象は色をまたげない。
			"異色の將士象",
			[]*Card{
				tsCard(TuSacColorRed, TuSacPieceGeneral),
				tsCard(TuSacColorGreen, TuSacPieceAdvisor),
				tsCard(TuSacColorWhite, TuSacPieceElephant),
			},
			TuSacMeldNone,
		},
		{
			"卒 4 枚では足りない",
			[]*Card{
				tsCard(TuSacColorRed, TuSacPieceSoldier),
				tsCard(TuSacColorGreen, TuSacPieceSoldier),
				tsCard(TuSacColorWhite, TuSacPieceSoldier),
				tsCard(TuSacColorYellow, TuSacPieceSoldier),
			},
			TuSacMeldNone,
		},
		{
			"卒 5 枚に卒でない札が混ざる",
			[]*Card{
				tsCard(TuSacColorRed, TuSacPieceSoldier),
				tsCard(TuSacColorGreen, TuSacPieceSoldier),
				tsCard(TuSacColorWhite, TuSacPieceSoldier),
				tsCard(TuSacColorYellow, TuSacPieceSoldier),
				tsCard(TuSacColorRed, TuSacPieceHorse),
			},
			TuSacMeldNone,
		},
		{"空", nil, TuSacMeldNone},
		{
			"nil が混ざる",
			[]*Card{tsCard(TuSacColorRed, TuSacPieceSoldier), nil, tsCard(TuSacColorRed, TuSacPieceSoldier)},
			TuSacMeldNone,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, TuSacClassifyMeld(tc.cards),
				"%s の判定が違う", tc.name)
		})
	}
}

// **得点はそろえにくさに見合っている。** 同じ点だと大きいメルドを狙う理由が
// 無くなる。
func TestTuSacMeldPoints(t *testing.T) {
	t.Parallel()
	assert.Greater(t, TuSacMeldPoints(TuSacMeldSoldierSet), TuSacMeldPoints(TuSacMeldChariotTrio))
	assert.Greater(t, TuSacMeldPoints(TuSacMeldChariotTrio), TuSacMeldPoints(TuSacMeldSameColorSet))
	assert.Positive(t, TuSacMeldPoints(TuSacMeldSameColorSet))
	assert.Zero(t, TuSacMeldPoints(TuSacMeldNone))
}

// **添字で選ばせる。** 同じ色・同じ駒が 4 枚あるので、札そのものを渡させると
// 「どの 1 枚か」が決まらない。
func TestTuSacFindMeld(t *testing.T) {
	t.Parallel()
	hand := []*Card{
		tsCard(TuSacColorRed, TuSacPieceElephant),
		tsCard(TuSacColorGreen, TuSacPieceHorse),
		tsCard(TuSacColorRed, TuSacPieceElephant),
		tsCard(TuSacColorRed, TuSacPieceElephant),
	}

	picked, kind := TuSacFindMeld(hand, []int{0, 2, 3})
	assert.Equal(t, TuSacMeldSameColorSet, kind)
	assert.Len(t, picked, TuSacSetSize)

	_, kind = TuSacFindMeld(hand, []int{0, 1, 2})
	assert.Equal(t, TuSacMeldNone, kind)

	// **同じ添字を 2 回使えない。** 1 枚を 2 枚として数えられてしまう。
	_, kind = TuSacFindMeld(hand, []int{0, 0, 0})
	assert.Equal(t, TuSacMeldNone, kind, "同じ札を 3 枚として数えている")

	// 範囲外。
	_, kind = TuSacFindMeld(hand, []int{0, 2, 99})
	assert.Equal(t, TuSacMeldNone, kind)
	_, kind = TuSacFindMeld(hand, []int{-1, 0, 2})
	assert.Equal(t, TuSacMeldNone, kind)
}

func TestTuSacConfig_Validate(t *testing.T) {
	t.Parallel()
	assert.NoError(t, DefaultTuSacConfig().Validate())
	assert.ErrorIs(t, TuSacConfig{Seats: 1, Rounds: 5}.Validate(), errTuSacSeatsRange)
	assert.ErrorIs(t, TuSacConfig{Seats: 5, Rounds: 5}.Validate(), errTuSacSeatsRange)
	assert.ErrorIs(t, TuSacConfig{Seats: 4, Rounds: 0}.Validate(), errTuSacRoundsRange)
	assert.ErrorIs(t, TuSacConfig{Seats: 4, Rounds: 99}.Validate(), errTuSacRoundsRange)

	// **配り切らない。** 4 席 × 20 = 80 枚で、山に 32 枚残る。
	assert.Less(t, TuSacMaxSeats*TuSacHandSize, TuSacDeckSize, "配り切ってしまう")
}

func TestTuSacMeldKindName(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, k := range []TuSacMeldKind{
		TuSacMeldSameColorSet, TuSacMeldChariotTrio, TuSacMeldSoldierSet,
	} {
		n := TuSacMeldKindName(k)
		assert.NotEqual(t, "none", n)
		assert.False(t, seen[n], "メルド名が重複: %s", n)
		seen[n] = true
	}
	assert.Equal(t, "none", TuSacMeldKindName(TuSacMeldNone))
}
