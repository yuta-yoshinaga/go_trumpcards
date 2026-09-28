//go:build test

package domain

// TrumpCardsForTest は山札を返す（テスト用）。ドローで引かれる札を仕込むのに使う。
func (cs *CaribbeanDraw) TrumpCardsForTest() *TrumpCards { return cs.trumpCards }
