//go:build !js || !wasm || casino

package domain

// BJSuggestedAction ベーシックストラテジー推奨アクション
type BJSuggestedAction int

// BJSuggestedAction定数
const (
	// BJSuggestNone 推奨なし
	BJSuggestNone BJSuggestedAction = 0
	// BJSuggestHit ヒット推奨
	BJSuggestHit BJSuggestedAction = 1
	// BJSuggestStand スタンド推奨
	BJSuggestStand BJSuggestedAction = 2
	// BJSuggestDouble ダブルダウン推奨
	BJSuggestDouble BJSuggestedAction = 3
	// BJSuggestSplit スプリット推奨
	BJSuggestSplit BJSuggestedAction = 4
	// BJSuggestSurrender サレンダー推奨
	BJSuggestSurrender BJSuggestedAction = 5
	// BJSuggestDeclineInsurance インシュランス辞退推奨
	BJSuggestDeclineInsurance BJSuggestedAction = 6
	// BJSuggestDoubleStand ダブルダウン推奨（不可ならスタンド）
	BJSuggestDoubleStand BJSuggestedAction = 7
)

// dealerIdx ディーラーのアップカードをインデックスに変換
// 2→0, 3→1, …, 9→7, 10/J/Q/K→8, A→9
func dealerIdx(card *Card) int {
	v := card.GetValue()
	switch {
	case v == 1:
		return 9
	case v >= 10:
		return 8
	default:
		return v - 2
	}
}

// pairValue ペアのBJ値を返す（ペアでなければ -1）
func pairValue(hand *BlackJackHand) int {
	if hand.GetCardsSize() != 2 {
		return -1
	}
	v0 := bjValue(hand.GetCard(0))
	v1 := bjValue(hand.GetCard(1))
	if v0 == v1 {
		return v0
	}
	return -1
}

// GetBasicStrategyAction ベーシックストラテジーによる推奨アクションを返す
// (standard multi-deck, S17 or H17)
// H=Hit, S=Stand, D=Double(else Hit), Ds=Double(else Stand), Sp=Split, Rh=Surrender(else Hit)
func GetBasicStrategyAction(hand *BlackJackHand, dealerUpcard *Card, dealerHitsSoft17 bool) BJSuggestedAction {
	di := dealerIdx(dealerUpcard)

	// ① ペアチェック
	if pv := pairValue(hand); pv >= 0 {
		return pairStrategy(pv, di)
	}

	// ② ソフトハンド
	if hand.IsSoft() {
		action := softStrategy(hand.GetScore(), di)
		// H17 overrides for soft hands
		if dealerHitsSoft17 {
			action = softH17Override(hand.GetScore(), di, action)
		}
		return action
	}

	// ③ ハードハンド
	hardTotal := hand.GetScore()
	action := hardStrategy(hardTotal, di)
	// H17 overrides for hard hands
	if dealerHitsSoft17 {
		action = hardH17Override(hardTotal, di, action)
	}
	return action
}

// Dealer upcard index: 2→0, 3→1, 4→2, 5→3, 6→4, 7→5, 8→6, 9→7, 10→8, A→9

// pairStrategy ペア戦略 (pv=BJvalue 1-10, di=dealerIdx 0-9)
func pairStrategy(pv, di int) BJSuggestedAction {
	// Sp=Split, H=Hit, D=Double, S=Stand
	// Rows: pair value 1,2,3,4,5,6,7,8,9,10 (index pv-1; pv==10 → index 9)
	// pair value 1 = Ace pair (bjValue returns 1 for ace)
	type row [10]BJSuggestedAction
	Sp := BJSuggestSplit
	H := BJSuggestHit
	D := BJSuggestDouble
	S := BJSuggestStand
	table := [10]row{
		// dealer: 2  3  4  5  6  7  8  9  10  A
		{Sp, Sp, Sp, Sp, Sp, Sp, Sp, Sp, Sp, Sp}, // A,A
		{Sp, Sp, Sp, Sp, Sp, Sp, H, H, H, H},     // 2,2
		{Sp, Sp, Sp, Sp, Sp, Sp, H, H, H, H},     // 3,3
		{H, H, H, Sp, Sp, H, H, H, H, H},         // 4,4
		{D, D, D, D, D, D, D, D, H, H},           // 5,5 (treat as hard 10, never split)
		{Sp, Sp, Sp, Sp, Sp, H, H, H, H, H},      // 6,6
		{Sp, Sp, Sp, Sp, Sp, Sp, H, H, H, H},     // 7,7
		{Sp, Sp, Sp, Sp, Sp, Sp, Sp, Sp, Sp, Sp}, // 8,8
		{Sp, Sp, Sp, Sp, Sp, S, Sp, Sp, S, S},    // 9,9
		{S, S, S, S, S, S, S, S, S, S},           // 10,10
	}
	// pv is always 1–10 from bjValue, so idx is always 0–9
	return table[pv-1][di]
}

// softStrategy ソフトハンド戦略 (softTotal=13..20, di=0..9)
func softStrategy(softTotal, di int) BJSuggestedAction {
	H := BJSuggestHit
	S := BJSuggestStand
	D := BJSuggestDouble       // double else hit
	Ds := BJSuggestDoubleStand // double else stand
	type row [10]BJSuggestedAction
	// Rows: softTotal 13..20 (index = softTotal-13)
	table := [8]row{
		// dealer: 2   3   4   5   6   7   8   9  10   A
		{H, H, H, D, D, H, H, H, H, H},      // soft 13 (A+2)
		{H, H, H, D, D, H, H, H, H, H},      // soft 14 (A+3)
		{H, H, D, D, D, H, H, H, H, H},      // soft 15 (A+4)
		{H, H, D, D, D, H, H, H, H, H},      // soft 16 (A+5)
		{H, D, D, D, D, H, H, H, H, H},      // soft 17 (A+6)
		{Ds, Ds, Ds, Ds, Ds, S, S, H, H, H}, // soft 18 (A+7)
		// soft 19 vs 6 は **S17 ならスタンド**。ここを Ds にすると softH17Override が
		// 何も変えられず、S17 のプレイヤーに H17 用の助言が出る (#4705 のソルバが検出。
		// 実測 EV: S17 stand 0.4960 > double 0.4796 / H17 double 0.4611 > stand 0.4531)。
		{S, S, S, S, S, S, S, S, S, S}, // soft 19 (A+8)
		{S, S, S, S, S, S, S, S, S, S}, // soft 20 (A+9)
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

// hardStrategy ハードハンド戦略 (hardTotal, di=0..9)
func hardStrategy(hardTotal, di int) BJSuggestedAction {
	H := BJSuggestHit
	S := BJSuggestStand
	D := BJSuggestDouble
	Rh := BJSuggestSurrender // surrender else hit
	type row [10]BJSuggestedAction
	// Rows: hard total clamped to [5, 17] (index = clamp-5, so index 0..12)
	table := [13]row{
		// dealer: 2  3  4  5  6  7  8  9  10  A
		{H, H, H, H, H, H, H, H, H, H},    // hard 5 (≤8 all hit)
		{H, H, H, H, H, H, H, H, H, H},    // hard 6
		{H, H, H, H, H, H, H, H, H, H},    // hard 7
		{H, H, H, H, H, H, H, H, H, H},    // hard 8
		{H, D, D, D, D, H, H, H, H, H},    // hard 9
		{D, D, D, D, D, D, D, D, H, H},    // hard 10
		{D, D, D, D, D, D, D, D, D, H},    // hard 11
		{H, H, S, S, S, H, H, H, H, H},    // hard 12
		{S, S, S, S, S, H, H, H, H, H},    // hard 13
		{S, S, S, S, S, H, H, H, H, H},    // hard 14
		{S, S, S, S, S, H, H, H, Rh, H},   // hard 15
		{S, S, S, S, S, H, H, Rh, Rh, Rh}, // hard 16
		{S, S, S, S, S, S, S, S, S, S},    // hard 17+
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

// hardH17Override ハードハンドのH17ルールでの変更
// S17→H17で変わるセル: Hard 11 vs A: H→D, Hard 15 vs A: H→Rh, Hard 17 vs A: S→Rh
func hardH17Override(hardTotal, di int, s17Action BJSuggestedAction) BJSuggestedAction {
	// di=9 はエース
	if di != 9 {
		return s17Action
	}
	switch hardTotal {
	case 11:
		return BJSuggestDouble // S17でもDだが明示的にDを返す
	case 15:
		return BJSuggestSurrender
	case 17:
		return BJSuggestSurrender
	default:
		return s17Action
	}
}

// softH17Override ソフトハンドのH17ルールでの変更
// Soft 19 vs 6: S→Ds(DoubleStand)
func softH17Override(softTotal, di int, s17Action BJSuggestedAction) BJSuggestedAction {
	// di=4 は6
	if softTotal == 19 && di == 4 {
		return BJSuggestDoubleStand
	}
	return s17Action
}
