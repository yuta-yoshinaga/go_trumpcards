//go:build test

package domain

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (k *Kille) SetPhaseForTest(p KillePhase) { k.phase = p }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (k *Kille) SetCurrentPlayerForTest(idx int) { k.currentIdx = idx }

// SetDealerForTest はテスト用にディーラーを差し替える。
func (k *Kille) SetDealerForTest(idx int) { k.dealerIdx = idx }

// SetPotForTest はテスト用にポットを差し替える。
func (k *Kille) SetPotForTest(n int) { k.pot = n }

// SetHandForTest はテスト用に席の手札を 1 枚に差し替える。
func (k *Kille) SetHandForTest(seat int, r KilleRank) {
	p := k.GetPlayer(seat)
	p.Reset()
	p.AddCard(NewKilleCard(r))
	p.SetHarlequinSwapped(false)
}

// SetStockForTest はテスト用に山札を差し替える。
func (k *Kille) SetStockForTest(cards []*Card) { k.stock = cards }
