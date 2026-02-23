package domain

import (
	"math/rand"
	"sort"
)

// ポーカーゲームのフェーズ定数
const (
	PokerPhaseInit      = 0 // 初期状態
	PokerPhaseDeal      = 1 // カード配布後 (第1ベッティングラウンド)
	PokerPhaseExchange  = 2 // カード交換フェーズ
	PokerPhaseSecondBet = 3 // 第2ベッティングラウンド
	PokerPhaseEnd       = 4 // ゲーム終了
)

// ポーカーデフォルト値
const (
	PokerDefaultChips = 1000
	PokerDefaultAnte  = 10
	PokerMinBet       = 10
)

// ディーラーAI閾値
const (
	dealerFoldPotOddsThreshold    = 0.4 // ポットオッズがこれを超えるとフォールド候補
	dealerFoldBetMultiplierWeak   = 2   // 弱いハイカードでフォールドするベット倍率
	dealerFoldBetMultiplierStrong = 3   // 高額ベットでフォールドするベット倍率
	dealerFoldRateWeak            = 70  // 弱いハンド+悪いポットオッズ時のフォールド率(%)
	dealerFoldRateStrong          = 50  // 高額ベット時のフォールド率(%)
)

// Poker ポーカークラス (5枚ドローポーカー)
type Poker struct {
	trumpCards *TrumpCards   // トランプカード
	player     *PokerPlayer // プレイヤー
	dealer     *PokerPlayer // ディーラー(CPU)
	phase      int          // ゲームフェーズ
	pot        int          // ポット
	playerBet  int          // プレイヤーの現ラウンドベット
	dealerBet  int          // ディーラーの現ラウンドベット
	ante       int          // アンティ
	folded     int          // フォールド状態 (0:なし, 1:プレイヤー, 2:ディーラー)
}

// NewPoker コンストラクタ
func NewPoker(trumpCards *TrumpCards, player *PokerPlayer, dealer *PokerPlayer) *Poker {
	return &Poker{
		trumpCards: trumpCards,
		player:     player,
		dealer:     dealer,
		phase:      PokerPhaseInit,
		ante:       PokerDefaultAnte,
	}
}

// Reset ゲーム初期化
func (p *Poker) Reset() {
	p.phase = PokerPhaseInit
	p.pot = 0
	p.playerBet = 0
	p.dealerBet = 0
	p.folded = 0
	for i := 0; i < 10; i++ {
		p.trumpCards.Shuffle()
	}
	p.player.Reset()
	p.dealer.Reset()
	// チップ初期化 (0以下の場合はデフォルト値)
	if p.player.GetChips() <= 0 {
		p.player.SetChips(PokerDefaultChips)
	}
	if p.dealer.GetChips() <= 0 {
		p.dealer.SetChips(PokerDefaultChips)
	}
	// アンティ徴収
	p.collectAnte()
	// プレイヤー・ディーラーに5枚ずつ配る
	for i := 0; i < 5; i++ {
		p.player.AddCard(p.trumpCards.DrawCard())
		p.dealer.AddCard(p.trumpCards.DrawCard())
	}
	// ディーラー第1ベット
	p.dealerFirstBet()
	p.phase = PokerPhaseDeal
}

// collectAnte アンティ徴収
func (p *Poker) collectAnte() {
	playerAnte := p.ante
	if p.player.GetChips() < playerAnte {
		playerAnte = p.player.GetChips()
	}
	dealerAnte := p.ante
	if p.dealer.GetChips() < dealerAnte {
		dealerAnte = p.dealer.GetChips()
	}
	p.player.SubtractChips(playerAnte)
	p.dealer.SubtractChips(dealerAnte)
	p.pot = playerAnte + dealerAnte
}

// PlayerBet プレイヤーベット (フェーズ1,3)
func (p *Poker) PlayerBet(amount int) error {
	if p.phase != PokerPhaseDeal && p.phase != PokerPhaseSecondBet {
		return NewDomainError(ErrWrongPhase, "Bet is not allowed now.")
	}
	if amount < PokerMinBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	if p.player.GetChips() < amount {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	p.player.SubtractChips(amount)
	p.playerBet += amount
	p.pot += amount
	// ディーラーのベットに対する応答
	p.dealerRespondToBet()
	// フォールドした場合は終了
	if p.folded != 0 {
		p.phase = PokerPhaseEnd
		return nil
	}
	p.advanceAfterBetting()
	return nil
}

// PlayerCall プレイヤーコール (ディーラーのベットに合わせる)
func (p *Poker) PlayerCall() error {
	if p.phase != PokerPhaseDeal && p.phase != PokerPhaseSecondBet {
		return NewDomainError(ErrWrongPhase, "Call is not allowed now.")
	}
	diff := p.dealerBet - p.playerBet
	if diff <= 0 {
		return NewDomainError(ErrInvalidPlay, "Nothing to call.")
	}
	if p.player.GetChips() < diff {
		// オールインの場合はあるだけ出す
		diff = p.player.GetChips()
	}
	p.player.SubtractChips(diff)
	p.playerBet += diff
	p.pot += diff
	p.advanceAfterBetting()
	return nil
}

// PlayerRaise プレイヤーレイズ (ディーラーのベット以上に上乗せ)
func (p *Poker) PlayerRaise(amount int) error {
	if p.phase != PokerPhaseDeal && p.phase != PokerPhaseSecondBet {
		return NewDomainError(ErrWrongPhase, "Raise is not allowed now.")
	}
	if amount < PokerMinBet {
		return NewDomainError(ErrInvalidAmount, "Invalid raise amount.")
	}
	diff := p.dealerBet - p.playerBet
	if diff < 0 {
		diff = 0
	}
	// オーバーフロー防止
	if amount > p.player.GetChips()-diff {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for raise.")
	}
	totalNeeded := diff + amount
	p.player.SubtractChips(totalNeeded)
	p.playerBet += totalNeeded
	p.pot += totalNeeded
	// ディーラーのレイズに対する応答
	p.dealerRespondToBet()
	if p.folded != 0 {
		p.phase = PokerPhaseEnd
		return nil
	}
	p.advanceAfterBetting()
	return nil
}

// PlayerFold プレイヤーフォールド
func (p *Poker) PlayerFold() error {
	if p.phase != PokerPhaseDeal && p.phase != PokerPhaseSecondBet {
		return NewDomainError(ErrWrongPhase, "Fold is not allowed now.")
	}
	p.folded = 1
	// ポットをディーラーに渡す
	p.dealer.AddChips(p.pot)
	p.pot = 0
	p.phase = PokerPhaseEnd
	return nil
}

// PlayerCheck プレイヤーチェック (ベットなしでパス)
func (p *Poker) PlayerCheck() error {
	if p.phase != PokerPhaseDeal && p.phase != PokerPhaseSecondBet {
		return NewDomainError(ErrWrongPhase, "Check is not allowed now.")
	}
	// 未決済のベットがある場合はチェック不可
	if p.dealerBet > p.playerBet {
		return NewDomainError(ErrInvalidPlay, "Cannot check with outstanding bet.")
	}
	p.advanceAfterBetting()
	return nil
}

// advanceAfterBetting ベッティング後のフェーズ遷移
func (p *Poker) advanceAfterBetting() {
	if p.phase == PokerPhaseDeal {
		p.phase = PokerPhaseExchange
	} else if p.phase == PokerPhaseSecondBet {
		p.resolveShowdown()
	}
}

// PlayerExchange プレイヤーカード交換
func (p *Poker) PlayerExchange(indices []int) error {
	if p.phase != PokerPhaseExchange {
		return NewDomainError(ErrWrongPhase, "Exchange is not allowed now.")
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
	// ラウンドベットリセット
	p.playerBet = 0
	p.dealerBet = 0
	// ディーラー第2ベット
	p.dealerSecondBet()
	p.phase = PokerPhaseSecondBet
	return nil
}

// PlayerStand カード交換なしでショーダウン
func (p *Poker) PlayerStand() error {
	if p.phase != PokerPhaseExchange {
		return NewDomainError(ErrWrongPhase, "Stand is not allowed now.")
	}
	// ディーラーの自動カード交換
	p.dealerExchange()
	// ラウンドベットリセット
	p.playerBet = 0
	p.dealerBet = 0
	// ディーラー第2ベット
	p.dealerSecondBet()
	p.phase = PokerPhaseSecondBet
	return nil
}

// resolveShowdown ショーダウン: ハンド評価・ポット配分
func (p *Poker) resolveShowdown() {
	p.player.EvalHand()
	p.dealer.EvalHand()
	switch p.GameJudgment() {
	case 1:
		p.player.AddChips(p.pot)
	case -1:
		p.dealer.AddChips(p.pot)
	default:
		// 引き分けの場合はポットを均等に分配
		half := p.pot / 2
		p.player.AddChips(half)
		p.dealer.AddChips(p.pot - half)
	}
	p.pot = 0
	p.phase = PokerPhaseEnd
}

// dealerFirstBet ディーラー第1ベット (カード配布直後)
func (p *Poker) dealerFirstBet() {
	// シンプルAI: ハンド評価してからベット判断
	p.dealer.EvalHand()
	rank := p.dealer.GetHandRank()
	if rank >= PokerHandOnePair {
		// ペア以上ならベット
		bet := PokerMinBet
		if p.dealer.GetChips() >= bet {
			p.dealer.SubtractChips(bet)
			p.dealerBet = bet
			p.pot += bet
		}
	}
	// ハイカードならチェック (dealerBetは0のまま)
}

// dealerRespondToBet ディーラーのプレイヤーベットへの応答
func (p *Poker) dealerRespondToBet() {
	p.dealer.EvalHand()
	rank := p.dealer.GetHandRank()
	diff := p.playerBet - p.dealerBet
	if diff <= 0 {
		return
	}

	if rank == PokerHandHighCard {
		// ハイカードの強さを評価 (A,K を持っているか)
		hasHighCard := false
		for i := 0; i < p.dealer.GetCardsSize(); i++ {
			v := p.dealer.GetCard(i).GetValue()
			if v == 1 || v >= 12 { // A, Q, K
				hasHighCard = true
				break
			}
		}
		// ポットオッズ: コールに必要な額 vs 獲得可能なポット
		potOdds := 0.0
		if p.pot+diff > 0 {
			potOdds = float64(diff) / float64(p.pot+diff)
		}
		// フォールド判定: ハイカードなし かつ ポットオッズが悪い かつ 大きなベット
		if !hasHighCard && potOdds > dealerFoldPotOddsThreshold && diff > PokerMinBet*dealerFoldBetMultiplierWeak {
			// 70%の確率でフォールドする（残りの30%でコールを試みる）
			if rand.Intn(100) < dealerFoldRateWeak {
				p.folded = 2
				p.player.AddChips(p.pot)
				p.pot = 0
				return
			}
		} else if !hasHighCard && diff > PokerMinBet*dealerFoldBetMultiplierStrong {
			// 50%の確率でフォールドする（残りの50%でコールを試みる）
			if rand.Intn(100) < dealerFoldRateStrong {
				p.folded = 2
				p.player.AddChips(p.pot)
				p.pot = 0
				return
			}
		}
	}

	// コール
	callAmount := diff
	if p.dealer.GetChips() < callAmount {
		callAmount = p.dealer.GetChips()
	}
	p.dealer.SubtractChips(callAmount)
	p.dealerBet += callAmount
	p.pot += callAmount
}

// dealerSecondBet ディーラー第2ベット (カード交換後)
func (p *Poker) dealerSecondBet() {
	p.dealer.EvalHand()
	rank := p.dealer.GetHandRank()
	if rank >= PokerHandTwoPair {
		// ツーペア以上ならベット
		bet := PokerMinBet
		if rank >= PokerHandFullHouse {
			bet = PokerMinBet * 3
		} else if rank >= PokerHandStraight {
			bet = PokerMinBet * 2
		}
		if p.dealer.GetChips() >= bet {
			p.dealer.SubtractChips(bet)
			p.dealerBet = bet
			p.pot += bet
		}
	}
}

// dealerExchange ディーラーの自動カード交換 (AI)
func (p *Poker) dealerExchange() {
	// まずハンドを評価
	p.dealer.EvalHand()
	rank := p.dealer.GetHandRank()

	if rank >= PokerHandTwoPair {
		// ツーペア以上はカード交換しない
		return
	}

	// フラッシュドロー判定
	if rank < PokerHandOnePair {
		discardIdx := p.findFlushDrawDiscard()
		if discardIdx >= 0 {
			newCard := p.trumpCards.DrawCard()
			if newCard != nil {
				p.dealer.ExchangeCard(discardIdx, newCard)
			}
			return
		}
	}

	// ストレートドロー判定
	if rank < PokerHandOnePair {
		discardIdx := p.findStraightDrawDiscard()
		if discardIdx >= 0 {
			newCard := p.trumpCards.DrawCard()
			if newCard != nil {
				p.dealer.ExchangeCard(discardIdx, newCard)
			}
			return
		}
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

// findFlushDrawDiscard 4枚フラッシュドローの外れカード位置を返す (ドローなし: -1)
func (p *Poker) findFlushDrawDiscard() int {
	suitCounts := make(map[int]int)
	for i := 0; i < p.dealer.GetCardsSize(); i++ {
		suitCounts[p.dealer.GetCard(i).GetDesign()]++
	}
	for suit, count := range suitCounts {
		if count == 4 {
			for i := 0; i < p.dealer.GetCardsSize(); i++ {
				if p.dealer.GetCard(i).GetDesign() != suit {
					return i
				}
			}
		}
	}
	return -1
}

// straightDrawCardInfo ストレートドロー探索用のカード情報
type straightDrawCardInfo struct {
	idx   int
	value int
}

// findOpenEndedDraw sorted cardsからskip位置を除外し、残り4枚が連続かつ条件を満たすか判定する
func findOpenEndedDraw(cards []straightDrawCardInfo, check func(remaining []int) bool) int {
	for skip := 0; skip < len(cards); skip++ {
		remaining := make([]int, 0, 4)
		for j, c := range cards {
			if j != skip {
				remaining = append(remaining, c.value)
			}
		}
		if len(remaining) != 4 {
			continue
		}
		isConsecutive := true
		for k := 1; k < len(remaining); k++ {
			if remaining[k] != remaining[k-1]+1 {
				isConsecutive = false
				break
			}
		}
		if isConsecutive && check(remaining) {
			return cards[skip].idx
		}
	}
	return -1
}

// findStraightDrawDiscard 4枚オープンエンドストレートドローの外れカード位置を返す (ドローなし: -1)
func (p *Poker) findStraightDrawDiscard() int {
	// Ace high (14) として評価
	cards := make([]straightDrawCardInfo, p.dealer.GetCardsSize())
	for i := 0; i < p.dealer.GetCardsSize(); i++ {
		v := p.dealer.GetCard(i).GetValue()
		if v == 1 {
			v = 14
		}
		cards[i] = straightDrawCardInfo{i, v}
	}
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].value < cards[j].value
	})

	// オープンエンド: 両端に拡張の余地がある (2以上 and 13以下)
	idx := findOpenEndedDraw(cards, func(r []int) bool {
		return r[0] > 1 && r[3] < 14
	})
	if idx >= 0 {
		return idx
	}

	// Ace low: A-2-3-4 のパターン (Aceを1として再評価、同じスライスを再利用)
	for i := range cards {
		if cards[i].value == 14 {
			cards[i].value = 1
		}
	}
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].value < cards[j].value
	})

	return findOpenEndedDraw(cards, func(r []int) bool {
		return r[0] == 1 && r[3] <= 5
	})
}

// GameJudgment ゲーム勝敗判定 (1:勝ち, 0:引き分け, -1:負け)
func (p *Poker) GameJudgment() int {
	if p.folded == 1 {
		return -1 // プレイヤーフォールド
	}
	if p.folded == 2 {
		return 1 // ディーラーフォールド
	}
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

// GetPot ポット取得
func (p *Poker) GetPot() int {
	return p.pot
}

// GetPlayerBet プレイヤーベット取得
func (p *Poker) GetPlayerBet() int {
	return p.playerBet
}

// GetDealerBet ディーラーベット取得
func (p *Poker) GetDealerBet() int {
	return p.dealerBet
}

// GetAnte アンティ取得
func (p *Poker) GetAnte() int {
	return p.ante
}

// GetFolded フォールド状態取得
func (p *Poker) GetFolded() int {
	return p.folded
}
