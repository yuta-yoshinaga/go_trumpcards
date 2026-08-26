//go:build test

package domain

// StackTopForTest は山札の「次に引かれる位置」から順に指定カードが出るように
// 並べ替え、実際に配置できた枚数を返す（テスト用）。
//
// 差し込みではなく **入れ替え** で実装している。デッキに同じカードを増やすと、
// 「そのランクは残り何枚か」を見るロジックが本番と違う山札を相手にしてしまう。
// 入れ替えなら 52 枚の内訳は変わらないまま、配られる順序だけが決まる。
//
// **戻り値を必ず検査すること。** 既に配り終えたカードは山札に無いので取り戻せず、
// その位置は元のまま残る。黙って飛ばされたことに気付かないと、狙った並びだと
// 思い込んだまま別の配りを検査するテストになる（実際に一度そうなった）。
// 呼び出し側は `require.Equal(t, len(cards), n)` で止めるのが正しい。
func (t *TrumpCards) StackTopForTest(cards ...*Card) int {
	placed := 0
	for i, want := range cards {
		pos := t.deckDrawCnt + i
		if want == nil || pos >= t.deckCnt {
			return placed
		}
		for j := pos; j < t.deckCnt; j++ {
			c := t.deck[j]
			if c.GetDesign() == want.GetDesign() && c.GetValue() == want.GetValue() {
				t.deck[pos], t.deck[j] = t.deck[j], t.deck[pos]
				placed++
				break
			}
		}
	}
	return placed
}

// ReplenishForTest は配布済みフラグだけを戻し、山札の全カードを再び引ける
// 状態にする（テスト用）。StackTopForTest の直前に呼べば、Reset が配り終えた
// カードも含めて狙った並びを作れる。
func (t *TrumpCards) ReplenishForTest() { t.deckInit() }
