//go:build test

package domain

import (
	"math/rand"
	"testing"
)

// TestVPStrategyEveryRowIsReachable は戦略表のどの行にも、その行で推奨が決まる手が
// 存在することを確かめる。行の順序を入れ替えた結果、上の行に完全に覆われて
// 一度も選ばれなくなった行 (死んだ行) があれば落ちる。手は役が出やすいように
// 偏らせた乱数で作る (同スート・狭いランク幅・少ないランク・ロイヤルのランク)。
func TestVPStrategyEveryRowIsReachable(t *testing.T) {
	tables := []struct {
		variant string
		wilds   int
	}{
		{"jacksorbetter", 0},
		{"deuceswild", 0}, {"deuceswild", 1}, {"deuceswild", 2}, {"deuceswild", 3}, {"deuceswild", 4},
		{"jokerpoker", 0}, {"jokerpoker", 1},
	}
	rng := rand.New(rand.NewSource(10969))
	for _, tb := range tables {
		rows := vpStrategy(tb.variant, tb.wilds)
		hit := make([]bool, len(rows))
		remaining := len(rows)
		for n := 0; n < 200000 && remaining > 0; n++ {
			hand := vpBiasedHand(rng, tb.variant, tb.wilds)
			advice, row := vpRecommend(tb.variant, hand)
			if row < 0 {
				continue
			}
			if advice.RuleKey != rows[row].key {
				t.Fatalf("%s W=%d: row %d key %q but advice %q", tb.variant, tb.wilds, row, rows[row].key, advice.RuleKey)
			}
			if !hit[row] {
				hit[row] = true
				remaining--
			}
		}
		for row, ok := range hit {
			if !ok {
				t.Errorf("%s W=%d: row %d (%s) is never chosen", tb.variant, tb.wilds, row, rows[row].key)
			}
		}
	}
}

// vpBiasedHand は wilds 枚のワイルドを含む 5 枚を作る。自然札は重複しない。
func vpBiasedHand(rng *rand.Rand, variant string, wilds int) []*Card {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	hand := make([]*Card, 0, VideoPokerHandSize)
	for i := 0; i < wilds; i++ {
		if variant == "jokerpoker" {
			hand = append(hand, NewCard(CardDesignJoker, 0, false))
		} else {
			hand = append(hand, NewCard(suits[i], 2, false))
		}
	}
	natural := func(value int) bool { return variant != "deuceswild" || value != 2 }
	used := map[[2]int]bool{}
	add := func(suit, value int) {
		key := [2]int{suit, value}
		if !natural(value) || used[key] {
			return
		}
		used[key] = true
		hand = append(hand, NewCard(suit, value, false))
	}
	mode := rng.Intn(7)
	suit := suits[rng.Intn(4)]
	lo := 1 + rng.Intn(10) // ランク幅 lo..lo+4 (14 は A)
	ranks := []int{1 + rng.Intn(13), 1 + rng.Intn(13)}
	for tries := 0; len(hand) < VideoPokerHandSize && tries < 1000; tries++ {
		s := suits[rng.Intn(4)]
		v := 1 + rng.Intn(13)
		switch mode {
		case 1: // 同スート
			s = suit
		case 2: // 狭いランク幅
			v = vpWindowValue(lo, rng)
		case 3: // 少ないランク (ペア・スリーカード・フォーカード)
			v = ranks[rng.Intn(2)]
		case 4: // ロイヤルのランク
			v = vpWindowValue(10, rng)
		case 5: // 狭いランク幅かつ同スート
			s, v = suit, vpWindowValue(lo, rng)
		case 6: // ロイヤルのランクかつ同スート
			s, v = suit, vpWindowValue(10, rng)
		}
		add(s, v)
	}
	for len(hand) < VideoPokerHandSize { // 偏りで埋まらなかった分は一様に
		add(suits[rng.Intn(4)], 1+rng.Intn(13))
	}
	rng.Shuffle(len(hand), func(i, j int) { hand[i], hand[j] = hand[j], hand[i] })
	return hand
}

func vpWindowValue(lo int, rng *rand.Rand) int {
	v := lo + rng.Intn(5)
	if v == 14 {
		return 1
	}
	return v
}
