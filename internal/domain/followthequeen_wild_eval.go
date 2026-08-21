//go:build !js || !wasm || casino

package domain

// evalFiveCardHandWithWilds は「どの札がワイルドか」を外から渡して 5 枚を評価する。
//
// リポジトリには既に `evalFiveCardHandWithJokers` があるが、あれはジョーカーという
// **特定の札**を見るうえ 2 枚までしか扱わない。Follow the Queen のワイルドは
// 普通の札（Q と、その時々のランク）で、1 組に最大 8 枚あるため 3 枚以上が
// 同じ手に入りうる ── そのままでは使えない。
//
// 計算量は代入の総当たりで 52^w。w が 4 以上なら必ずファイブカードになるので
// 総当たりに入る前に返し、ロイヤルフラッシュに届いた時点でも打ち切る。
func evalFiveCardHandWithWilds(cards []*Card, isWild func(*Card) bool) int {
	if len(cards) != 5 {
		return PokerHandHighCard
	}

	wildIdx := make([]int, 0, 5)
	for i, c := range cards {
		if isWild(c) {
			wildIdx = append(wildIdx, i)
		}
	}
	if len(wildIdx) == 0 {
		return evalFiveCardHand(cards)
	}

	// **ワイルドが 4 枚以上なら必ずファイブカード。**残り 1 枚に合わせれば
	// よく、5 枚なら何にでもなる。
	//
	// これは**速さのための枝で、答えを変えない** ── 消しても総当たりが同じ
	// 結論に達する。ただし 52^4 = 730 万通りを数え上げてからになるので、
	// ショーダウンが目に見えて止まる。テストで挙動の差として観測できない類の
	// 分岐なので、「テスト済み」ではなく最適化としてここに書いておく。
	//
	// 3 枚以下は特別扱いしない。`evalFiveCardHand` は同ランク 5 枚をきちんと
	// ファイブカードと判定するので、代入の総当たりがそのまま拾う。
	if len(wildIdx) >= 4 {
		return PokerHandFiveOfAKind
	}

	// 元の並びを壊さないよう写しの上で差し替える。呼び出し側の手札を書き換えると、
	// 表示とショーダウンで別の手が出る。
	work := make([]*Card, 5)
	copy(work, cards)

	best := PokerHandHighCard
	var recurse func(depth int)
	recurse = func(depth int) {
		if best >= PokerHandFiveOfAKind {
			return // これ以上は無い（ファイブカードが最高位）
		}
		if depth == len(wildIdx) {
			if r := evalFiveCardHand(work); r > best {
				best = r
			}
			return
		}
		for d := CardDesignSpade; d <= CardDesignDiamond; d++ {
			for v := 1; v <= CardValueMax; v++ {
				work[wildIdx[depth]] = NewCard(d, v, false)
				recurse(depth + 1)
				if best >= PokerHandFiveOfAKind {
					return
				}
			}
		}
	}
	recurse(0)
	return best
}
