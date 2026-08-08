//go:build !js || !wasm || casino

package domain

// Spanish 21 basic strategy (#4705).
//
// **この表は出典から写したものではなく、このゲームの規則から解いたもの。**
// スパニッシュ21の公表表は出典によってセルが食い違い (ハード12 vs 6 は「ヒット」と
// 「スタンド」で割れる)、Wizard of Odds の表は画像で機械可読でない。さらに市販の表は
// ほぼ H17 用だが、`DefaultBlackJackConfig()` は S17 で、しかも本実装は
// `PlayerDoubleDown` が2枚のときしかダブルを許さない (本来のスパニッシュ21は何枚でも
// ダブルできる)。他所の表をそのまま載せると、標準表のままより悪い助言になりうる。
//
// 生成と検証:
//
//	go test -tags test ./internal/domain -run TestGenerateSpanish21Table -v
//
// ソルバ (`blackjack_strategy_solver_test.go`) は 48 枚デッキ・S17・
// `Player21AlwaysWins`・ボーナス配当・レイトサレンダーという**このゲームの設定**で
// EV を解く。ソルバ自体の正しさは `TestSolver_ReproducesStandardTable` が担保する
// (同じソルバを標準デッキで走らせ、`hardStrategy`/`softStrategy`/`pairStrategy` を
// 再現できることを確認している)。
//
// 標準表との主な違い — いずれも10抜きデッキの帰結:
//   - ハード 12・13 は**常にヒット**。10 が無いぶんバストしにくい
//   - ハード 14 は 4・5・6 に対してだけスタンド
//   - ハード 11 は 9・10・A に対してダブルしない (標準は 10 までダブル)
//   - ソフト 13・14 はダブルしない
//   - ハード 17 vs A はサレンダー
//   - 4,4 はスプリットしない / 7,7 は 8 以上に対してスプリットしない

// spanish21HardStrategy はスパニッシュ21のハードハンド戦略 (hardTotal, di=0..9)。
func spanish21HardStrategy(hardTotal, di int) BJSuggestedAction {
	H := BJSuggestHit
	S := BJSuggestStand
	D := BJSuggestDouble
	Rh := BJSuggestSurrender
	type row [10]BJSuggestedAction
	// Rows: hard total clamped to [5, 17] (index = clamp-5)
	table := [13]row{
		// dealer: 2  3  4  5  6  7  8  9  10  A
		{H, H, H, H, H, H, H, H, H, H},  // hard 5
		{H, H, H, H, H, H, H, H, H, H},  // hard 6
		{H, H, H, H, H, H, H, H, H, H},  // hard 7
		{H, H, H, H, H, H, H, H, H, H},  // hard 8
		{H, H, H, H, D, H, H, H, H, H},  // hard 9
		{D, D, D, D, D, D, D, H, H, H},  // hard 10
		{D, D, D, D, D, D, D, H, H, H},  // hard 11
		{H, H, H, H, H, H, H, H, H, H},  // hard 12 — 10 抜きなのでバストしにくい
		{H, H, H, H, H, H, H, H, H, H},  // hard 13
		{H, H, S, S, S, H, H, H, H, H},  // hard 14
		{S, S, S, S, S, H, H, H, H, H},  // hard 15
		{S, S, S, S, S, H, H, H, H, H},  // hard 16
		{S, S, S, S, S, S, S, S, S, Rh}, // hard 17+
	}
	clamped := hardTotal
	if clamped < 5 {
		clamped = 5
	}
	if clamped > 17 {
		clamped = 17
	}
	return table[clamped-5][di]
}

// spanish21SoftStrategy はスパニッシュ21のソフトハンド戦略 (softTotal=13..20, di=0..9)。
func spanish21SoftStrategy(softTotal, di int) BJSuggestedAction {
	H := BJSuggestHit
	S := BJSuggestStand
	D := BJSuggestDouble
	Ds := BJSuggestDoubleStand
	type row [10]BJSuggestedAction
	// Rows: softTotal 13..20 (index = softTotal-13)
	table := [8]row{
		// dealer: 2   3   4   5   6   7   8   9  10   A
		{H, H, H, H, H, H, H, H, H, H},    // soft 13
		{H, H, H, H, H, H, H, H, H, H},    // soft 14
		{H, H, H, H, D, H, H, H, H, H},    // soft 15
		{H, H, H, D, D, H, H, H, H, H},    // soft 16
		{H, H, D, D, D, H, H, H, H, H},    // soft 17
		{S, S, Ds, Ds, Ds, S, S, H, H, H}, // soft 18
		{S, S, S, S, S, S, S, S, S, S},    // soft 19
		{S, S, S, S, S, S, S, S, S, S},    // soft 20
	}
	idx := softTotal - 13
	if idx < 0 {
		idx = 0
	}
	if idx > 7 {
		idx = 7
	}
	return table[idx][di]
}

// spanish21PairStrategy はスパニッシュ21のペア戦略 (pv=BJ値 1-10, di=0..9)。
func spanish21PairStrategy(pv, di int) BJSuggestedAction {
	Sp := BJSuggestSplit
	H := BJSuggestHit
	D := BJSuggestDouble
	S := BJSuggestStand
	type row [10]BJSuggestedAction
	table := [10]row{
		// dealer: 2  3  4  5  6  7  8  9  10  A
		{Sp, Sp, Sp, Sp, Sp, Sp, Sp, Sp, Sp, Sp}, // A,A
		{Sp, Sp, Sp, Sp, Sp, Sp, Sp, H, H, H},    // 2,2
		{Sp, Sp, Sp, Sp, Sp, Sp, Sp, H, H, H},    // 3,3
		{H, H, H, H, H, H, H, H, H, H},           // 4,4 — 割らない
		{D, D, D, D, D, D, D, H, H, H},           // 5,5 (ハード10扱い)
		{H, H, Sp, Sp, Sp, H, H, H, H, H},        // 6,6
		{Sp, Sp, Sp, Sp, Sp, Sp, H, H, H, H},     // 7,7
		{Sp, Sp, Sp, Sp, Sp, Sp, Sp, Sp, Sp, Sp}, // 8,8
		{S, Sp, Sp, Sp, Sp, S, Sp, Sp, S, S},     // 9,9
		{S, S, S, S, S, S, S, S, S, S},           // 10,10
	}
	return table[pv-1][di]
}

// GetSpanish21StrategyAction はスパニッシュ21の推奨アクションを返す。
//
// ディーラーのソフト17ルールは受け取らない: 表は S17 (このゲームの設定) で解いて
// あり、H17 用の補正は解き直さないと出せないため。設定が H17 に変わったら
// `spanish21Rules()` の `dealerHitsSoft17` を立てて再生成すること。
func GetSpanish21StrategyAction(hand *BlackJackHand, dealerUpcard *Card) BJSuggestedAction {
	di := dealerIdx(dealerUpcard)
	if pv := pairValue(hand); pv >= 0 {
		return spanish21PairStrategy(pv, di)
	}
	if hand.IsSoft() {
		return spanish21SoftStrategy(hand.GetScore(), di)
	}
	return spanish21HardStrategy(hand.GetScore(), di)
}

// GetVariantStrategyAction はバリアントに応じた推奨アクションを返す。
// スパニッシュ21以外は標準の基本戦略表にフォールバックする。
func GetVariantStrategyAction(
	hand *BlackJackHand,
	dealerUpcard *Card,
	dealerHitsSoft17 bool,
	variant BlackJackVariantName,
) BJSuggestedAction {
	if variant == BJVariantSpanish21 {
		return GetSpanish21StrategyAction(hand, dealerUpcard)
	}
	return GetBasicStrategyAction(hand, dealerUpcard, dealerHitsSoft17)
}
