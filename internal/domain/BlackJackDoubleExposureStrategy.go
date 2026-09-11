//go:build !js || !wasm || casino

package domain

// Double Exposure basic strategy.
//
// **この表は出典から写したものではなく、このゲームの規則から解いたもの。**
// ディーラーの2枚が表向きで、通常のブラックジャックと異なり同点はディーラー勝ち、
// ナチュラルBJの配当は1:1になる規則を、既存のEVソルバで解いて生成する。
//
// 生成:
//
//	go test -tags test ./internal/domain -run TestGenerateDoubleExposureTable -v
//
// 生成結果で標準表と異なる代表例は、ハード17対ディーラーのハード17がH、
// ハード20対ハード20がRh、ハード16対ハード8がRhになる点である。前二者は
// 同点負け、後者は2枚の合計を使うことの帰結で、アップカード1枚の標準表とは異なる。
// この表は生成物であり、ソルバと一致していることを TestDoubleExposureTable_MatchesSolver が厳密比較で担保する。

var doubleExposureHardTable = [16][26]BJSuggestedAction{
	{BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                                                        // hard 5
	{BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                                                        // hard 6
	{BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                                                        // hard 7
	{BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                                                  // hard 8
	{BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                                            // hard 9
	{BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                          // hard 10
	{BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit},                                                // hard 11
	{BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                                         // hard 12
	{BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender},                                   // hard 13
	{BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender},                           // hard 14
	{BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender},                     // hard 15
	{BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender},               // hard 16
	{BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender}, // hard 17
	{BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender},                 // hard 18
	{BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestSurrender, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestSurrender, BJSuggestSurrender},                         // hard 19
	{BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestSurrender},                                 // hard 20
}

var doubleExposureSoftTable = [8][26]BJSuggestedAction{
	{BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                                            // soft 13
	{BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                                            // soft 14
	{BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                                         // soft 15
	{BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                                      // soft 16
	{BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                                      // soft 17
	{BJSuggestDoubleStand, BJSuggestDoubleStand, BJSuggestDoubleStand, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestStand, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                              // soft 18
	{BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestDoubleStand, BJSuggestDoubleStand, BJSuggestDoubleStand, BJSuggestDoubleStand, BJSuggestDoubleStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestHit, BJSuggestSurrender}, // soft 19
	{BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestDoubleStand, BJSuggestDoubleStand, BJSuggestDoubleStand, BJSuggestDoubleStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestSurrender},   // soft 20
}

var doubleExposurePairTable = [10][26]BJSuggestedAction{
	{BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit},                             // pair 1
	{BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSplit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                                 // pair 2
	{BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSplit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                                 // pair 3
	{BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                                           // pair 4
	{BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestSplit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestDouble, BJSuggestDouble, BJSuggestDouble, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                               // pair 5
	{BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestStand, BJSuggestSplit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender},                                         // pair 6
	{BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestHit, BJSuggestHit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestSplit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender},           // pair 7
	{BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender, BJSuggestSurrender, BJSuggestStand, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestHit, BJSuggestSurrender, BJSuggestSurrender}, // pair 8
	{BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestStand, BJSuggestSplit, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestStand, BJSuggestSplit, BJSuggestSurrender, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestSplit, BJSuggestStand, BJSuggestSplit, BJSuggestSurrender, BJSuggestSurrender},             // pair 9
	{BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestSplit, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestSurrender, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestStand, BJSuggestSurrender},                     // pair 10
}

// GetDoubleExposureStrategyAction は、ディーラーの公開された2枚の状態に対する
// Double Exposure の推奨アクションを返す。
func GetDoubleExposureStrategyAction(hand *BlackJackHand, dealerTotal int, dealerSoft bool) BJSuggestedAction {
	if hand.GetScore() == 21 {
		return BJSuggestStand
	}
	var dealerIndex int
	if dealerSoft {
		// 範囲外は表の端へクランプする。これは不正な合計でも隣接する
		// 最も近い既知のディーラー状態として助言するためのもの。
		if dealerTotal < 12 {
			dealerTotal = 12
		}
		if dealerTotal > 20 {
			dealerTotal = 20
		}
		dealerIndex = 17 + dealerTotal - 12
	} else {
		// 範囲外は表の端へクランプする。ハード4..20の列に収める。
		if dealerTotal < 4 {
			dealerTotal = 4
		}
		if dealerTotal > 20 {
			dealerTotal = 20
		}
		dealerIndex = dealerTotal - 4
	}
	if pv := pairValue(hand); pv >= 0 {
		return doubleExposurePairTable[pv-1][dealerIndex]
	}
	if hand.IsSoft() {
		total := hand.GetScore()
		if total < 13 {
			total = 13
		}
		if total > 20 {
			total = 20
		}
		return doubleExposureSoftTable[total-13][dealerIndex]
	}
	total := hand.GetScore()
	if total < 5 {
		total = 5
	}
	if total > 20 {
		total = 20
	}
	return doubleExposureHardTable[total-5][dealerIndex]
}
