//go:build !js || !wasm || extra4

package domain

import "fmt"

// BarbuDominoes.go は Dominoes コントラクト (7 並べ / fan-tan) のレイアウト処理。
// Sevens の bitmask 方式を踏襲し、各スートを 7 を起点に上下へ伸ばす。手札を
// 早く出し切ったプレイヤーから上がり順位が付く。

// barbuDominoSevenValue は 7 並べの起点となる値。
const barbuDominoSevenValue = 7

// isDominoPositionPlaced は (suit, value) が場に置かれているか。
func (b *Barbu) isDominoPositionPlaced(suit, value int) bool {
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return false
	}
	return b.tablePlaced[suit]&(uint16(1)<<uint(value)) != 0
}

// isDominoPositionPlayable は (suit, value) が今プレイ可能かを返す。
// 7 は常にプレイ可能 (未配置なら)。それ以外は 7 に近い側の隣接カードが
// 既に置かれている必要がある。
func (b *Barbu) isDominoPositionPlayable(suit, value int) bool {
	if value < CardValueMin+1 || value > CardValueMax {
		return false
	}
	if b.isDominoPositionPlaced(suit, value) {
		return false
	}
	if value == barbuDominoSevenValue {
		return true
	}
	if value < barbuDominoSevenValue {
		return b.isDominoPositionPlaced(suit, value+1)
	}
	return b.isDominoPositionPlaced(suit, value-1)
}

// isDominoCardPlayable はカードがプレイ可能かを返す。
func (b *Barbu) isDominoCardPlayable(card *Card) bool {
	if card == nil {
		return false
	}
	return b.isDominoPositionPlayable(card.GetDesign(), card.GetValue())
}

// hasDominoPlayable はプレイヤーがプレイ可能なカードを持つかを返す。
func (b *Barbu) hasDominoPlayable(playerIdx int) bool {
	return len(b.GetDominoPlayableIndices(playerIdx)) > 0
}

// GetDominoPlayableIndices はプレイヤーのプレイ可能な手札インデックス一覧を返す。
func (b *Barbu) GetDominoPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(b.players) {
		return nil
	}
	p := b.players[playerIdx]
	var idxs []int
	for i := 0; i < p.GetCardsSize(); i++ {
		if b.isDominoCardPlayable(p.GetCard(i)) {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

// applyDominoPlay は Dominoes での 1 手 (配置 or パス) を実行する。
// handIdx == -1 はパス (合法手がない場合のみ許可)。
func (b *Barbu) applyDominoPlay(playerIdx, handIdx int) error {
	player := b.players[playerIdx]

	if handIdx == -1 {
		if b.hasDominoPlayable(playerIdx) {
			return NewDomainError(ErrInvalidPlay, "you have a playable card and cannot pass")
		}
		b.passCount[playerIdx]++
		b.appendLog(playerIdx, "pass", fmt.Sprintf("player %d passes", playerIdx), nil)
		b.advanceDominoTurn()
		return nil
	}

	card := player.GetCard(handIdx)
	if card == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", handIdx))
	}
	if !b.isDominoCardPlayable(card) {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("%s cannot be placed", cardStr(card)))
	}

	played := player.RemoveCard(handIdx)
	b.tablePlaced[played.GetDesign()] |= uint16(1) << uint(played.GetValue())
	b.appendLog(playerIdx, "place", fmt.Sprintf("player %d places %s", playerIdx, cardStr(played)), []*Card{played})

	if player.GetCardsSize() == 0 {
		b.dominoFinished++
		player.SetDominoRank(b.dominoFinished)
		player.SetIsFinished(true)
		b.appendLog(playerIdx, "finish", fmt.Sprintf("player %d goes out (rank %d)", playerIdx, b.dominoFinished), nil)
	}
	b.advanceDominoTurn()
	return nil
}

// advanceDominoTurn は次の手番へ進める。残り 1 人になったらディール終了。
func (b *Barbu) advanceDominoTurn() {
	if b.dominoFinished >= BarbuPlayerCnt-1 {
		b.finishDominoDeal()
		return
	}
	next := nextActivePlayer(b.players, b.currentPlayer, 1)
	if next < 0 {
		b.finishDominoDeal()
		return
	}
	b.currentPlayer = next
}

// finishDominoDeal は Dominoes ディールを締める。最後まで上がれなかった
// プレイヤーに最下位を割り当ててから得点を確定する。
func (b *Barbu) finishDominoDeal() {
	for i, p := range b.players {
		if p.GetDominoRank() == 0 {
			b.dominoFinished++
			p.SetDominoRank(BarbuPlayerCnt)
			p.SetIsFinished(true)
			b.appendLog(i, "finish", fmt.Sprintf("player %d is last (rank %d)", i, BarbuPlayerCnt), nil)
		}
	}
	b.finishDeal()
}
