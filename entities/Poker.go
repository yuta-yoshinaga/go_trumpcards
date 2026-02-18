package entities

import "sort"

// ポーカーゲームのフェーズ定数
const (
	PokerPhaseInit = 0 // 初期状態
	PokerPhaseDeal = 1 // カード配布後
	PokerPhaseEnd  = 2 // ゲーム終了
)

// Poker ポーカークラス (5枚ドローポーカー)
type Poker struct {
	trumpCards *TrumpCards  // トランプカード
	player     *PokerPlayer // プレイヤー
	dealer     *PokerPlayer // ディーラー(CPU)
	phase      int          // ゲームフェーズ
}

// NewPoker コンストラクタ
func NewPoker(trumpCards *TrumpCards, player *PokerPlayer, dealer *PokerPlayer) *Poker {
	return &Poker{
		trumpCards: trumpCards,
		player:     player,
		dealer:     dealer,
		phase:      PokerPhaseInit,
	}
}

// Reset ゲーム初期化
func (p *Poker) Reset() {
	p.phase = PokerPhaseInit
	for i := 0; i < 10; i++ {
		p.trumpCards.Shuffle()
	}
	p.player.Reset()
	p.dealer.Reset()
	// プレイヤー・ディーラーに5枚ずつ配る
	for i := 0; i < 5; i++ {
		p.player.AddCard(p.trumpCards.DrawCard())
		p.dealer.AddCard(p.trumpCards.DrawCard())
	}
	p.phase = PokerPhaseDeal
}

// PlayerExchange プレイヤーカード交換
func (p *Poker) PlayerExchange(indices []int) {
	if p.phase != PokerPhaseDeal {
		return
	}
	// 指定カードを新しいカードと交換
	for _, idx := range indices {
		newCard := p.trumpCards.DrawCard()
		if newCard != nil {
			p.player.ExchangeCard(idx, newCard)
		}
	}
	// ディーラーの自動カード交換
	p.dealerExchange()
	// 手札評価
	p.player.EvalHand()
	p.dealer.EvalHand()
	p.phase = PokerPhaseEnd
}

// PlayerStand カード交換なしでショーダウン
func (p *Poker) PlayerStand() {
	if p.phase != PokerPhaseDeal {
		return
	}
	// ディーラーの自動カード交換
	p.dealerExchange()
	// 手札評価
	p.player.EvalHand()
	p.dealer.EvalHand()
	p.phase = PokerPhaseEnd
}

// dealerExchange ディーラーの自動カード交換 (シンプルなAI)
func (p *Poker) dealerExchange() {
	// まずハンドを評価
	p.dealer.EvalHand()
	rank := p.dealer.GetHandRank()

	if rank >= PokerHandTwoPair {
		// ツーペア以上はカード交換しない
		return
	}

	if rank == PokerHandOnePair {
		// ワンペアならペア以外の3枚を交換
		valueCounts := make(map[int][]int)
		for i := 0; i < p.dealer.GetCardsSize(); i++ {
			v := p.dealer.GetCard(i).GetValue()
			valueCounts[v] = append(valueCounts[v], i)
		}
		indices := []int{}
		for _, idxList := range valueCounts {
			if len(idxList) == 1 {
				indices = append(indices, idxList[0])
			}
		}
		for _, idx := range indices {
			newCard := p.trumpCards.DrawCard()
			if newCard != nil {
				p.dealer.ExchangeCard(idx, newCard)
			}
		}
	} else {
		// ハイカードなら最も低い3枚を交換
		type cardIdx struct {
			idx   int
			value int
		}
		cards := make([]cardIdx, p.dealer.GetCardsSize())
		for i := 0; i < p.dealer.GetCardsSize(); i++ {
			v := p.dealer.GetCard(i).GetValue()
			if v == 1 {
				v = 14 // エースはハイカードとして扱う
			}
			cards[i] = cardIdx{i, v}
		}
		sort.Slice(cards, func(i, j int) bool {
			return cards[i].value < cards[j].value
		})
		// 最も低い3枚を交換
		for i := 0; i < 3 && i < len(cards); i++ {
			newCard := p.trumpCards.DrawCard()
			if newCard != nil {
				p.dealer.ExchangeCard(cards[i].idx, newCard)
			}
		}
	}
}

// GameJudgment ゲーム勝敗判定 (1:勝ち, 0:引き分け, -1:負け)
func (p *Poker) GameJudgment() int {
	playerRank := p.player.GetHandRank()
	dealerRank := p.dealer.GetHandRank()

	if playerRank > dealerRank {
		return 1
	} else if playerRank < dealerRank {
		return -1
	}
	return p.compareHighCards()
}

// compareHighCards 同ランク時のハイカード比較
func (p *Poker) compareHighCards() int {
	playerValues := make([]int, p.player.GetCardsSize())
	dealerValues := make([]int, p.dealer.GetCardsSize())

	for i := 0; i < p.player.GetCardsSize(); i++ {
		v := p.player.GetCard(i).GetValue()
		if v == 1 {
			v = 14
		}
		playerValues[i] = v
	}
	for i := 0; i < p.dealer.GetCardsSize(); i++ {
		v := p.dealer.GetCard(i).GetValue()
		if v == 1 {
			v = 14
		}
		dealerValues[i] = v
	}

	sort.Sort(sort.Reverse(sort.IntSlice(playerValues)))
	sort.Sort(sort.Reverse(sort.IntSlice(dealerValues)))

	for i := 0; i < len(playerValues) && i < len(dealerValues); i++ {
		if playerValues[i] > dealerValues[i] {
			return 1
		} else if playerValues[i] < dealerValues[i] {
			return -1
		}
	}
	return 0
}

// GetPhase ゲームフェーズ取得
func (p *Poker) GetPhase() int {
	return p.phase
}

// GetPlayer プレイヤー取得
func (p *Poker) GetPlayer() *PokerPlayer {
	return p.player
}

// GetDealer ディーラー取得
func (p *Poker) GetDealer() *PokerPlayer {
	return p.dealer
}
