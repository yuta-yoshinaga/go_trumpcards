//go:build !js || !wasm || solo

package domain

import "fmt"

// QuodlibetShedding.go は第 3 の輪の 2 種目 ── 四分 (Quadrature) と
// 小食い (Snack) ── を担う。**どちらもトリックを取らない。** 手札を早く
// 出し切った順に、残した札 1 枚あたりの罰点が軽くなる。

// quodlibetRankOrder は 32 枚デッキの弱い順の並び (7 → A)。
//
// **エースの値は 1 なので、数値の大小では並ばない。** 「ちょうど 3 つ上」も
// 「隣」も、この並びの上での話。
var quodlibetRankOrder = []int{7, 8, 9, 10, 11, 12, 13, 1}

// QuodlibetSnackAnchorIndex は小食いの起点 (J) の位置。
//
// 32 枚デッキは 7 から A までなので、7 並べの「7」にあたるのは J。
const QuodlibetSnackAnchorIndex = 4

// QuodlibetQuadratureStep は四分で重ねられる段差 (ちょうど 3 つ上)。
const QuodlibetQuadratureStep = 3

// QuodlibetRankIndex は札の値を弱い順の位置 (0-7) に直す。範囲外は -1。
func QuodlibetRankIndex(value int) int {
	for i, v := range quodlibetRankOrder {
		if v == value {
			return i
		}
	}
	return -1
}

// isSnackPlaced は (suit, rankIdx) が場に置かれているかを返す。
func (q *Quodlibet) isSnackPlaced(suit, rankIdx int) bool {
	if suit < CardDesignSpade || suit > CardDesignDiamond || rankIdx < 0 || rankIdx >= len(quodlibetRankOrder) {
		return false
	}
	return q.tablePlaced[suit]&(uint16(1)<<uint(rankIdx)) != 0
}

// isSnackCardPlayable は小食いでその札が置けるかを返す。
func (q *Quodlibet) isSnackCardPlayable(card *Card) bool {
	if card == nil {
		return false
	}
	idx := QuodlibetRankIndex(card.GetValue())
	if idx < 0 || q.isSnackPlaced(card.GetDesign(), idx) {
		return false
	}
	if idx == QuodlibetSnackAnchorIndex {
		return true
	}
	if idx < QuodlibetSnackAnchorIndex {
		return q.isSnackPlaced(card.GetDesign(), idx+1)
	}
	return q.isSnackPlaced(card.GetDesign(), idx-1)
}

// isQuadratureCardPlayable は四分でその札が重ねられるかを返す。
//
// **重ねは同じスートのちょうど 3 つ上でしか続かない。** 続かなくなったら
// 全員がパスし、次の人が好きな札で新しい重ねを始める。
func (q *Quodlibet) isQuadratureCardPlayable(card *Card) bool {
	if card == nil {
		return false
	}
	if len(q.stack) == 0 {
		return true
	}
	top := q.stack[len(q.stack)-1]
	if top == nil || card.GetDesign() != top.GetDesign() {
		return false
	}
	return QuodlibetRankIndex(card.GetValue()) == QuodlibetRankIndex(top.GetValue())+QuodlibetQuadratureStep
}

// GetSheddingPlayableIndices はシェディング系で出せる手札インデックスを返す。
func (q *Quodlibet) GetSheddingPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(q.players) {
		return nil
	}
	p := q.players[playerIdx]
	idxs := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		card := p.GetCard(i)
		playable := false
		switch q.currentContract {
		case QuodlibetSnack:
			playable = q.isSnackCardPlayable(card)
		case QuodlibetQuadrature:
			playable = q.isQuadratureCardPlayable(card)
		}
		if playable {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

// applySheddingPlay はシェディング系の 1 手を実行する。handIdx == -1 はパス。
func (q *Quodlibet) applySheddingPlay(playerIdx, handIdx int) error {
	player := q.players[playerIdx]
	legal := q.GetSheddingPlayableIndices(playerIdx)

	if handIdx == -1 {
		if len(legal) > 0 {
			return NewDomainErrorCode(ErrInvalidPlay, "quodlibet.errCannotPass", nil)
		}
		q.passCount[playerIdx]++
		q.appendLog(playerIdx, "pass", fmt.Sprintf("player %d passes", playerIdx), nil)
		q.advanceSheddingTurn(true)
		return nil
	}

	if handIdx < 0 || handIdx >= player.GetCardsSize() {
		return NewDomainErrorCode(ErrInvalidCard, "quodlibet.errCardRange", nil)
	}
	if err := validateCardIsPlayable(legal, player, player.GetCard(handIdx)); err != nil {
		return err
	}

	played := player.RemoveCard(handIdx)
	switch q.currentContract {
	case QuodlibetSnack:
		q.tablePlaced[played.GetDesign()] |= uint16(1) << uint(QuodlibetRankIndex(played.GetValue()))
	case QuodlibetQuadrature:
		q.stack = append(q.stack, played)
	}
	q.appendLog(playerIdx, "place",
		fmt.Sprintf("player %d places %s", playerIdx, cardStr(played)), []*Card{played})

	if player.GetCardsSize() == 0 {
		q.outCount++
		player.SetOutRank(q.outCount)
		player.SetIsFinished(true)
		q.appendLog(playerIdx, "finish",
			fmt.Sprintf("player %d goes out (rank %d)", playerIdx, q.outCount), nil)
	}
	q.advanceSheddingTurn(false)
	return nil
}

// advanceSheddingTurn は次の手番へ進める。
//
// **全員が続けてパスしたら重ねを畳む。** 畳まないと四分は誰も出せないまま
// 手番だけが回り続ける。
func (q *Quodlibet) advanceSheddingTurn(passed bool) {
	if q.outCount >= QuodlibetPlayerCnt-1 {
		q.finishSheddingDeal()
		return
	}
	active := q.activeSheddingCount()
	if !passed {
		q.passCount = [QuodlibetPlayerCnt]int{}
	} else if q.currentContract == QuodlibetQuadrature && q.consecutivePasses() >= active {
		q.stack = nil
		q.passCount = [QuodlibetPlayerCnt]int{}
		q.appendLog(-1, "clear_stack", "nobody could follow; the stack is cleared", nil)
	}
	next := nextActivePlayer(q.players, q.currentPlayer, 1)
	if next < 0 {
		q.finishSheddingDeal()
		return
	}
	q.currentPlayer = next
}

// activeSheddingCount はまだ上がっていない人数を返す。
func (q *Quodlibet) activeSheddingCount() int {
	cnt := 0
	for _, p := range q.players {
		if !p.GetIsFinished() {
			cnt++
		}
	}
	return cnt
}

// consecutivePasses は畳み判定に使うパスの総数を返す。
func (q *Quodlibet) consecutivePasses() int {
	total := 0
	for _, n := range q.passCount {
		total += n
	}
	return total
}

// finishSheddingDeal はシェディング系のディールを締める。
//
// **上がれなかった人にも順位を付ける。** 順位がそのまま 1 枚あたりの罰点の
// 刻みになるので、0 のまま残すと罰点が消える。
func (q *Quodlibet) finishSheddingDeal() {
	for i, p := range q.players {
		if p.GetOutRank() != 0 {
			continue
		}
		q.outCount++
		p.SetOutRank(q.outCount)
		p.SetIsFinished(true)
		q.appendLog(i, "finish", fmt.Sprintf("player %d ends with %d cards (rank %d)",
			i, p.GetCardsSize(), q.outCount), nil)
	}
	q.finishDeal()
}
