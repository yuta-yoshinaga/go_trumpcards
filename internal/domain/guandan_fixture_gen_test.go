//go:build test

package domain

import (
	"encoding/json"
	"math/rand"
	"os"
	"testing"
)

type gdFixCard struct {
	Design string `json:"design"`
	Value  int    `json:"value"`
}
type gdFixCase struct {
	Cards []gdFixCard `json:"cards"`
	Level int         `json:"level"`
	Kind  int         `json:"kind"`
	Rank  int         `json:"rank"`
	Size  int         `json:"size"`
}

func TestGenGuandanFixture(t *testing.T) {
	if os.Getenv("GD_FIXTURE") == "" {
		t.Skip("generator")
	}
	names := map[int]string{CardDesignSpade: "SPADE", CardDesignHeart: "HEART", CardDesignDiamond: "DIAMOND", CardDesignClover: "CLOVER", CardDesignJoker: "JOKER"}
	deck := newGuandanDeck()
	r := rand.New(rand.NewSource(20260805))
	out := []gdFixCase{}
	for range 4000 {
		level := 2 + r.Intn(13)
		n := 1 + r.Intn(6)
		picked := make([]*Card, 0, n)
		// **狭いランク窓から引く。**一様に引くと役がほとんど出ず、
		// 階段・飛行機・木板を一件も踏まないまま「全一致」になる。
		base := 1 + r.Intn(13)
		// **同スートに寄せる回も要る。**でないとストレートフラッシュが一件も出ない。
		fixedSuit := -1
		if r.Intn(5) == 0 {
			fixedSuit = []int{CardDesignSpade, CardDesignHeart, CardDesignDiamond, CardDesignClover}[r.Intn(4)]
		}
		if fixedSuit >= 0 {
			n = 5
		}
		for i := range n {
			if fixedSuit < 0 && r.Intn(12) == 0 {
				picked = append(picked, deck[r.Intn(len(deck))])
				continue
			}
			v := base + r.Intn(4)
			if fixedSuit >= 0 {
				// 同スートのときは連番にして、ストレートフラッシュを実際に作る。
				v = base + i
			}
			if v > 13 {
				v -= 13
			}
			suits := []int{CardDesignSpade, CardDesignHeart, CardDesignDiamond, CardDesignClover}
			suit := suits[r.Intn(4)]
			if fixedSuit >= 0 {
				suit = fixedSuit
			}
			var found *Card
			for _, c := range deck {
				if c.GetDesign() == suit && c.GetValue() == v {
					found = c
					break
				}
			}
			if found == nil {
				found = deck[r.Intn(len(deck))]
			}
			picked = append(picked, found)
		}
		if r.Intn(60) == 0 {
			// ジョーカーボムも踏ませる。
			picked = picked[:0]
			for _, c := range deck {
				if c.GetDesign() == CardDesignJoker {
					picked = append(picked, c)
				}
				if len(picked) == 4 {
					break
				}
			}
		}
		combo := GuandanEvaluate(picked, level)
		cs := make([]gdFixCard, 0, n)
		for _, c := range picked {
			cs = append(cs, gdFixCard{names[c.GetDesign()], c.GetValue()})
		}
		cse := gdFixCase{Cards: cs, Level: level}
		if combo != nil {
			cse.Kind, cse.Rank, cse.Size = int(combo.Kind), combo.Rank, combo.Size
		}
		out = append(out, cse)
	}
	b, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv("GD_FIXTURE"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}
