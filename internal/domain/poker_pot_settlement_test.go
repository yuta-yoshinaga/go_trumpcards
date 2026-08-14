//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// **ハンドが終わったらポットは 0。**
//
// 配り終えたのに `pot` が残っていると、残高を読む側 (卓をまたいでチップを持ち回る
// ミックスゲーム、結果画面) が「まだ配っていない額がある」と読む。全員降りて終わる
// 経路 (`resolveLastPlayer`) は昔から 0 にしており、ショーダウンで終わる経路だけが
// 残していた —— 同じ「ハンドが終わった」状態なのに片方だけ違うのが誤りだった。
//
// 実測 (リバイ/アドオンを切った 200 ハンド): 176 ハンドで pot が残っていた。
// チップ自体は配られているので**総量は保存している**。
// pokerPotGame は「ハンドを閉じて残高とポットを読む」だけの共通操作。
//
// **6 実装は共有コードではない。** 同じ 1 行の修正を 6 か所に入れたので、
// 1 か所だけ戻されても気づけるように、6 つとも同じ検査を通す (レビュー指摘)。
type pokerPotGame struct {
	name       string
	reset      func() error
	fold       func() error
	phase      func() int
	endPhase   int
	stopPhases []int
	pot        func() int
	chips      func() int
}

// holdemLikePot は Holdem 系 (Holdem/Omaha/ShortDeck/Pineapple) の共通フェーズ。
func pokerPotGames() []pokerPotGame {
	games := make([]pokerPotGame, 0, 6)

	h := NewDefaultHoldem()
	cfgH := h.GetConfig()
	cfgH.RebuyEnabled, cfgH.AddonEnabled = false, false
	h.SetConfig(cfgH)
	games = append(games, pokerPotGame{
		name:  "Holdem",
		reset: h.Reset,
		fold:  func() error { return h.PlayerAction(HoldemActionFold, 0, 0) },
		phase: h.GetPhase, endPhase: HoldemPhaseEnd,
		stopPhases: []int{HoldemPhaseShowdown, HoldemPhaseRebuy},
		pot:        h.GetPot,
		chips: func() int {
			t := 0
			for i := 0; i < h.GetPlayerCnt(); i++ {
				t += h.GetPlayer(i).GetChips()
			}
			return t
		},
	})

	o := NewDefaultOmahaHiLo()
	cfgO := o.GetConfig()
	cfgO.RebuyEnabled, cfgO.AddonEnabled = false, false
	o.SetConfig(cfgO)
	games = append(games, pokerPotGame{
		name:  "OmahaHiLo",
		reset: o.Reset,
		fold:  func() error { return o.PlayerAction(OmahaActionFold, 0, 0) },
		phase: o.GetPhase, endPhase: OmahaPhaseEnd,
		stopPhases: []int{OmahaPhaseShowdown, OmahaPhaseRebuy},
		pot:        o.GetPot,
		chips: func() int {
			t := 0
			for i := 0; i < o.GetPlayerCnt(); i++ {
				t += o.GetPlayer(i).GetChips()
			}
			return t
		},
	})

	sd := NewDefaultShortDeck()
	cfgS := sd.GetConfig()
	cfgS.RebuyEnabled, cfgS.AddonEnabled = false, false
	sd.SetConfig(cfgS)
	games = append(games, pokerPotGame{
		name:  "ShortDeck",
		reset: sd.Reset,
		fold:  func() error { return sd.PlayerAction(ShortDeckActionFold, 0, 0) },
		phase: sd.GetPhase, endPhase: ShortDeckPhaseEnd,
		stopPhases: []int{ShortDeckPhaseShowdown, ShortDeckPhaseRebuy},
		pot:        sd.GetPot,
		chips: func() int {
			t := 0
			for i := 0; i < sd.GetPlayerCnt(); i++ {
				t += sd.GetPlayer(i).GetChips()
			}
			return t
		},
	})

	pa := NewDefaultPineapple()
	cfgP := pa.GetConfig()
	cfgP.RebuyEnabled, cfgP.AddonEnabled = false, false
	pa.SetConfig(cfgP)
	games = append(games, pokerPotGame{
		name:  "Pineapple",
		reset: pa.Reset,
		fold:  func() error { return pa.PlayerAction(PineappleActionFold, 0, 0) },
		phase: pa.GetPhase, endPhase: PineapplePhaseEnd,
		stopPhases: []int{PineapplePhaseShowdown, PineapplePhaseRebuy},
		pot:        pa.GetPot,
		chips: func() int {
			t := 0
			for i := 0; i < pa.GetPlayerCnt(); i++ {
				t += pa.GetPlayer(i).GetChips()
			}
			return t
		},
	})

	st := NewDefaultSevenCardStud()
	cfgSt := st.GetConfig()
	cfgSt.RebuyEnabled, cfgSt.AddonEnabled = false, false
	st.SetConfig(cfgSt)
	games = append(games, pokerPotGame{
		name:  "SevenCardStud",
		reset: st.Reset,
		fold:  func() error { return st.PlayerAction(SevenCardStudActionFold, 0, 0) },
		phase: st.GetPhase, endPhase: SevenCardStudPhaseEnd,
		stopPhases: []int{SevenCardStudPhaseShowdown, SevenCardStudPhaseRebuy},
		pot:        st.GetPot,
		chips: func() int {
			t := 0
			for i := 0; i < st.GetPlayerCnt(); i++ {
				t += st.GetPlayer(i).GetChips()
			}
			return t
		},
	})

	fs := NewDefaultFiveCardStud()
	cfgF := fs.GetConfig()
	cfgF.RebuyEnabled, cfgF.AddonEnabled = false, false
	fs.SetConfig(cfgF)
	games = append(games, pokerPotGame{
		name:  "FiveCardStud",
		reset: fs.Reset,
		fold:  func() error { return fs.PlayerAction(FiveCardStudActionFold, 0, 0) },
		phase: fs.GetPhase, endPhase: FiveCardStudPhaseEnd,
		stopPhases: []int{FiveCardStudPhaseShowdown, FiveCardStudPhaseRebuy},
		pot:        fs.GetPot,
		chips: func() int {
			t := 0
			for i := 0; i < fs.GetPlayerCnt(); i++ {
				t += fs.GetPlayer(i).GetChips()
			}
			return t
		},
	})
	return games
}

func TestPoker_PotIsZeroOnceTheHandIsOver(t *testing.T) {
	for _, g := range pokerPotGames() {
		t.Run(g.name, func(t *testing.T) {
			ended := 0
			for range 60 {
				require.NoError(t, g.reset())
				total := g.chips() + g.pot()

				for range 40 {
					ph := g.phase()
					if ph == g.endPhase || containsPhase(g.stopPhases, ph) {
						break
					}
					if err := g.fold(); err != nil {
						break
					}
				}
				if g.phase() != g.endPhase {
					continue // 人間のマック待ちなどはこの検査の対象外。
				}
				ended++
				assert.Equal(t, 0, g.pot(), "ハンドが終わったのにポットが残っている")
				// **チップの総量は変わらない。** 配られただけで、消えても湧いてもいない。
				assert.Equal(t, total, g.chips()+g.pot(), "チップの総量が変わっている")
			}
			require.Positive(t, ended, "1 ハンドも終局まで進まなかった")
		})
	}
}

// containsPhase はフェーズ番号が並びに含まれるか。
func containsPhase(phases []int, p int) bool {
	for _, v := range phases {
		if v == p {
			return true
		}
	}
	return false
}
