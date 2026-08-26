//go:build !js || !wasm || extra2

package domain

import "math/rand"

// BaccaratBanqueHint はバンカー (人間) への推奨手。
type BaccaratBanqueHint struct {
	// Draw は 3 枚目を引くべきか。
	Draw bool
	// Reason は理由の識別子。
	Reason string
}

// punterTakesOnFive は合計 5 の子が引くかを返す。
//
// **裁量があるのはここだけ。** 慣習では引くほうが有利とされるので、Normal と
// Hard は引く。Easy は五分五分で決める ── 難易度は CPU の腕の話であって、
// 規則の話ではないので、必ず引く / 必ず止まるほうは difficulty に関わらず固定。
func (b *BaccaratBanque) punterTakesOnFive(seat int) bool {
	if b.config.CpuDifficulty == BaccaratBanqueCpuDifficultyEasy {
		return rand.Intn(2) == 0 //nolint:gosec // ゲームの手選びに暗号強度は要らない
	}
	return true
}

// bankerShouldDraw はバンカーの合計と、見えている子の状況から引くべきかを返す。
//
// **バンカーは両方の子を見てから決められる。** 決まった表は無いので、
// 慣習的な目安 ── 5 以下なら引き、6 以上なら止まる ── を使う。ただし
// **両方の子が既に自分より上なら、止まっても負けるので引く**。
func (b *BaccaratBanque) bankerShouldDraw() bool {
	banker := b.players[BaccaratBanqueBankerIdx]
	total := banker.GetTotal()
	if BaccaratBanqueIsNatural(banker.GetHand()) {
		return false
	}
	beaten := 0
	for _, idx := range []int{BaccaratBanqueRightIdx, BaccaratBanqueLeftIdx} {
		if b.players[idx].GetTotal() > total {
			beaten++
		}
	}
	if beaten == 2 {
		// 止まれば両方に負ける ── 引くしかない。
		return true
	}
	return total <= BaccaratBanqueDiscretionTotal
}

// GetHint はバンカーへの推奨手を返す。
//
// **ヒントは難易度で鈍らせない。** CPU の難易度は子の裁量にだけ効く。
func (b *BaccaratBanque) GetHint() *BaccaratBanqueHint {
	if b.gameEndFlag || b.phase != BaccaratBanquePhaseBanker {
		return &BaccaratBanqueHint{Reason: "none"}
	}
	banker := b.players[BaccaratBanqueBankerIdx]
	if BaccaratBanqueIsNatural(banker.GetHand()) {
		return &BaccaratBanqueHint{Reason: "natural"}
	}
	if b.bankerShouldDraw() {
		total := banker.GetTotal()
		beaten := 0
		for _, idx := range []int{BaccaratBanqueRightIdx, BaccaratBanqueLeftIdx} {
			if b.players[idx].GetTotal() > total {
				beaten++
			}
		}
		if beaten == 2 {
			return &BaccaratBanqueHint{Draw: true, Reason: "behind_both"}
		}
		return &BaccaratBanqueHint{Draw: true, Reason: "low_total"}
	}
	return &BaccaratBanqueHint{Reason: "stand"}
}
