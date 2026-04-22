package domain

import (
	"fmt"
)

// triggerRevolutionIfNeeded 4枚出しで革命が起きるか判定し、起きた場合は革命フラグを切り替えて全プレイヤーの手札を再ソートする
func (p *President) triggerRevolutionIfNeeded(cards []*Card) {
	if !p.config.RevolutionEnabled {
		return
	}
	if len(cards) < 4 {
		return
	}
	p.round.revolutionActive = !p.round.revolutionActive
	p.appendLog(-1, "revolution", "revolution!", nil)
	p.sortAllActiveHands()
}

// performCardExchange 前ラウンドのランクに基づいてカード交換を行う
// Scum (4位) → President (1位): 最強カード 2枚
// President (1位) → Scum (4位): 最弱カード 2枚
// Vice Scum (3位) → Vice President (2位): 最強カード 1枚
// Vice President (2位) → Vice Scum (3位): 最弱カード 1枚
func (p *President) performCardExchange() {
	p.round.exchangeActions = make([]*PresidentExchangeAction, 0)

	presIdx := p.findPlayerByPrevRank(PresidentRankPresident)
	scumIdx := p.findPlayerByPrevRank(PresidentRankScum)
	if presIdx >= 0 && scumIdx >= 0 {
		p.exchangeCardsBetween(presIdx, scumIdx, PresidentExchangeCountPresident)
	}

	vpIdx := p.findPlayerByPrevRank(PresidentRankVicePresident)
	vsIdx := p.findPlayerByPrevRank(PresidentRankViceScum)
	if vpIdx >= 0 && vsIdx >= 0 {
		p.exchangeCardsBetween(vpIdx, vsIdx, PresidentExchangeCountViceTier)
	}

	// 交換記録を棋譜に追加
	for _, ex := range p.round.exchangeActions {
		p.appendLog(
			ex.FromPlayerIdx,
			"exchange",
			fmt.Sprintf("exchanged %d card(s) with player %d", len(ex.Cards), ex.ToPlayerIdx),
			ex.Cards,
		)
	}

	// 交換後に再ソート
	p.sortAllActiveHands()
}

// exchangeCardsBetween 上位プレイヤーと下位プレイヤー間でカード交換
// 下位→上位: 最強カードをcount枚渡す
// 上位→下位: 最弱カードをcount枚渡す
// 手札はすでに sortAllActiveHands で弱い順にソートされていることを前提とする
func (p *President) exchangeCardsBetween(upperIdx, lowerIdx, count int) {
	upper := p.players[upperIdx]
	lower := p.players[lowerIdx]

	if upper.GetCardsSize() < count || lower.GetCardsSize() < count {
		return
	}

	// 下位の最強カード(末尾)をcount枚取得
	lowerBestIndices := make([]int, count)
	for i := 0; i < count; i++ {
		lowerBestIndices[i] = lower.GetCardsSize() - count + i
	}
	lowerBestCards := lower.RemoveCards(lowerBestIndices)

	// 上位の最弱カード(先頭)をcount枚取得
	upperGiveIndices := make([]int, count)
	for i := 0; i < count; i++ {
		upperGiveIndices[i] = i
	}
	upperWorstCards := upper.RemoveCards(upperGiveIndices)

	// カードを交換
	for _, c := range lowerBestCards {
		upper.AddCard(c)
	}
	for _, c := range upperWorstCards {
		lower.AddCard(c)
	}

	p.round.exchangeActions = append(p.round.exchangeActions, &PresidentExchangeAction{
		FromPlayerIdx: lowerIdx,
		ToPlayerIdx:   upperIdx,
		Cards:         lowerBestCards,
	})
	p.round.exchangeActions = append(p.round.exchangeActions, &PresidentExchangeAction{
		FromPlayerIdx: upperIdx,
		ToPlayerIdx:   lowerIdx,
		Cards:         upperWorstCards,
	})
}
