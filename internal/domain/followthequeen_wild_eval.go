//go:build !js || !wasm || casino

package domain

// evalFiveCardHandWithWilds は「どの札がワイルドか」を外から渡して 5 枚を評価し、
// **役位と、その役を作る実際の 5 枚**を返す。
//
// リポジトリには既に `evalFiveCardHandWithJokers` があるが、あれはジョーカーという
// **特定の札**を見るうえ 2 枚までしか扱わない。Follow the Queen のワイルドは
// 普通の札（Q と、その時々のランク）で、1 組に最大 8 枚あるため 3 枚以上が
// 同じ手に入りうる ── そのままでは使えない。
//
// **置換後の 5 枚を返すのが要点。** 役位だけを返していたとき、同位の比較
// (`compareHighCardsSlice`) がワイルドの**印刷された額面**を読んでいた。
// Q♠Q♦3♣3♦9♥（Q は常時ワイルド ⇒ 3 のフォーカード）が、5♠5♦5♣5♥8♠（本物の
// 5 のフォーカード）に勝ってしまう。ワイルドランクすら要らない、普通に起きる形。
//
// # 探索範囲
//
// 素朴には 52^w 通り。実測で w=3 が 278 ms かかり、Worker の CPU 予算
// (無料枠 10 ms) を軽く超えていた。しかも CPU は毎ストリート全員分、CUI は
// 描画のたびに評価する。次の 3 段で、**答えを変えずに** 60 倍以上削っている。
//
//  1. 同ランクの実札 + ワイルドで 5 枚に届くなら、その場でファイブカード。
//  2. フラッシュ系（フラッシュ／ストレートフラッシュ／ロイヤル）は 5 枚が同スート
//     でなければ成立しないので、**ワイルドは全部そのスート**しかありえない。
//     4 スートそれぞれで、ワイルドのランクだけを数え上げる。
//  3. フラッシュ以外の役位はランクだけで決まる。ワイルドのスートは、いま最も
//     少ないスートを貪欲に選ぶ ── こうすると避けられる限りフラッシュにならず、
//     フラッシュ (5) がフルハウス (6) を隠す事故が起きない。
//
// ランクの数え上げは順序を持たない組み合わせ（非減少列）に畳んでいる。
// w=3 で 13^3=2197 → C(15,3)=455。`followthequeen_wild_eval_test.go` の
// オラクルテストが、乱数手 2000 通りで総当たりと一致することを毎回確認する。
func evalFiveCardHandWithWilds(cards []*Card, isWild func(*Card) bool) (int, []*Card) {
	if len(cards) != 5 {
		return PokerHandHighCard, nil
	}

	wildIdx := make([]int, 0, 5)
	for i, c := range cards {
		if isWild(c) {
			wildIdx = append(wildIdx, i)
		}
	}
	if len(wildIdx) == 0 {
		return evalFiveCardHand(cards), copyOf(cards)
	}

	work := make([]*Card, 5)
	copy(work, cards)

	// (1) 実札の最頻ランク + ワイルドで 5 枚に届けばファイブカード。
	// ワイルドが 4 枚以上なら実札 1 枚に合わせればよく、この条件に必ず入る。
	if rank, ok := followTheQueenFiveOfAKindRank(cards, wildIdx); ok {
		for _, i := range wildIdx {
			work[i] = NewCard(followTheQueenSpareSuit(work, i), rank, false)
		}
		return PokerHandFiveOfAKind, copyOf(work)
	}

	best := PokerHandHighCard
	var bestCards []*Card
	consider := func(hand []*Card) {
		r := evalFiveCardHand(hand)
		if r > best || (r == best && compareHighCardsSlice(hand, bestCards) > 0) {
			best = r
			bestCards = copyOf(hand)
		}
	}

	// (2) フラッシュ系。
	//
	// **候補のスートは高々 1 つ。** 手は 5 枚しかないので、フラッシュには
	// 実札が**全部**同じスートである必要がある（違うスートの実札は、その
	// フラッシュに参加しようがない）。実札のスートが割れていれば、この枝は
	// そもそも成立しないので丸ごと飛ばせる ── 4 スート回していたのが最大 1、
	// 普通の手では 0 になり、ここが探索時間の大半だった。
	if d, ok := followTheQueenSoleNaturalSuit(cards, wildIdx); ok {
		followTheQueenEachRankCombo(len(wildIdx), func(ranks []int) {
			for k, i := range wildIdx {
				work[i] = NewCard(d, ranks[k], false)
			}
			consider(work)
		})
		copy(work, cards)
	}

	// (3) フラッシュ以外: ランクだけが効く。スートは避けられる限りフラッシュを避ける。
	followTheQueenEachRankCombo(len(wildIdx), func(ranks []int) {
		for k, i := range wildIdx {
			work[i] = NewCard(followTheQueenSpareSuit(work, i), ranks[k], false)
		}
		consider(work)
	})

	return best, bestCards
}

// followTheQueenFiveOfAKindRank は、ワイルドを足してファイブカードにできる
// 実札のランクを返す。同率なら高いほうを採る（比較で有利）。
func followTheQueenFiveOfAKindRank(cards []*Card, wildIdx []int) (int, bool) {
	wild := make(map[int]bool, len(wildIdx))
	for _, i := range wildIdx {
		wild[i] = true
	}
	counts := map[int]int{}
	for i, c := range cards {
		if !wild[i] {
			counts[c.GetValue()]++
		}
	}
	if len(counts) == 0 {
		// 5 枚すべてワイルド。何にでもなるので最強の A を選ぶ。
		return 1, true
	}
	bestRank, ok := 0, false
	for v, n := range counts {
		if n+len(wildIdx) < 5 {
			continue
		}
		// A(=1) は評価器で最強に扱われるので、同点なら A を優先する。
		if !ok || v == 1 || (bestRank != 1 && v > bestRank) {
			bestRank, ok = v, true
		}
	}
	return bestRank, ok
}

// followTheQueenSoleNaturalSuit は、実札が全部同じスートならそのスートを返す。
// 割れていれば ok=false ── その手はどうやってもフラッシュにならない。
// 実札が 1 枚も無い（全部ワイルド）場合はスペードを返す: どのスートでも同じ
// 結果になるので 1 つ選べば足りる。
func followTheQueenSoleNaturalSuit(cards []*Card, wildIdx []int) (int, bool) {
	wild := make(map[int]bool, len(wildIdx))
	for _, i := range wildIdx {
		wild[i] = true
	}
	suit, seen := CardDesignSpade, false
	for i, c := range cards {
		if wild[i] {
			continue
		}
		d := c.GetDesign()
		if !seen {
			suit, seen = d, true
			continue
		}
		if d != suit {
			return 0, false
		}
	}
	return suit, true
}

// followTheQueenSpareSuit は work の中でいちばん枚数の少ないスートを返す
// （idx 番目は数えない ── これから置き換える枠なので）。
//
// フラッシュを「避けられる限り避ける」ためのもの。全部同じスートを割り当てると、
// 実札 3 枚が同スートの手でワイルド 2 枚を足した瞬間に必ずフラッシュになり、
// **フルハウス (6) がフラッシュ (5) に隠れて取りこぼす**。
func followTheQueenSpareSuit(work []*Card, idx int) int {
	counts := [5]int{}
	for i, c := range work {
		if i == idx || c == nil {
			continue
		}
		if d := c.GetDesign(); d >= CardDesignSpade && d <= CardDesignDiamond {
			counts[d]++
		}
	}
	bestSuit, bestCount := CardDesignSpade, counts[CardDesignSpade]
	for d := CardDesignSpade + 1; d <= CardDesignDiamond; d++ {
		if counts[d] < bestCount {
			bestSuit, bestCount = d, counts[d]
		}
	}
	return bestSuit
}

// followTheQueenEachRankCombo は w 枚のワイルドに割り当てるランクの
// **非減少列**を全部 fn に渡す。順列ではなく組み合わせなので、w=3 なら
// 13^3=2197 ではなく C(15,3)=455 通り。どの枠にどのランクを置くかは
// 役位に影響しない（同じ多重集合になる）。
func followTheQueenEachRankCombo(w int, fn func(ranks []int)) {
	ranks := make([]int, w)
	var rec func(depth, start int)
	rec = func(depth, start int) {
		if depth == w {
			fn(ranks)
			return
		}
		for v := start; v <= CardValueMax; v++ {
			ranks[depth] = v
			rec(depth+1, v)
		}
	}
	rec(0, 1)
}
