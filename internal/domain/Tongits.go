//go:build !js || !wasm || extra5

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// TongitsPlayerCnt Tongitsプレイヤー数
const TongitsPlayerCnt = 3

// TongitsHandSize 初期配布枚数
const TongitsHandSize = 12

// TongitsFirstPlayerHandSize is the initial hand size of the first player.
const TongitsFirstPlayerHandSize = 13

// TongitsKnockThreshold ノック可能なデッドウッド上限
const TongitsKnockThreshold = 5

// TongitsOnDealLow 配牌時のTongits(即勝利)成立点(下限)
const TongitsOnDealLow = 49

// TongitsOnDealHigh 配牌時のTongits(即勝利)成立点(上限)
const TongitsOnDealHigh = 50

// TongitsBonus 配牌時Tongits成立時のボーナス点
const TongitsBonus = 50

// TongitsUndercutRiskMax は「アンダーカットされうる」と警告する相手の残り枚数。
//
// **ノックは相手の残りが少ないほど裏目になる。**相手が先に上がれば、こちらの
// デッドウッドが低くても負ける (#1939)。Web はこの閾値でノックボタンに警告を
// 出しているが、閾値が画面側に書かれていて CUI は何も出していなかった (#5582)。
const TongitsUndercutRiskMax = 2

// TongitsUndercutPenalty アンダーカット時のペナルティ点
const TongitsUndercutPenalty = 5

// TongitsPhase ゲームフェーズ
type TongitsPhase int

// Tongitsのフェーズ定数
const (
	// TongitsPhaseDraw ドローフェーズ (山札または捨て札から引く)
	TongitsPhaseDraw TongitsPhase = 0
	// TongitsPhaseDiscard ディスカードフェーズ (手札から1枚捨てる or ノック)
	TongitsPhaseDiscard TongitsPhase = 1
	// TongitsPhaseRoundEnd ラウンド終了フェーズ
	TongitsPhaseRoundEnd TongitsPhase = 2
	// TongitsPhaseGameEnd ゲーム終了フェーズ
	TongitsPhaseGameEnd TongitsPhase = 3
)

// Tongits Tongitsゲームクラス
type Tongits struct {
	trumpCards       *TrumpCards
	players          []*TongitsPlayer
	config           TongitsConfig
	phase            TongitsPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	actionLogBase
	knockerIdx       int       // ノック/Tongitsしたプレイヤーのインデックス (-1 = 未確定)
	knockerMelds     [][]*Card // ノッカーのメルド
	knockerDeadwood  []*Card   // ノッカーのデッドウッド
	opponentMelds    [][]*Card // 相手のメルド (スコア確定時に格納)
	opponentDeadwood []*Card   // 相手のデッドウッド
	isTongits        bool      // 配牌Tongits(49/50点即勝利)かどうか
	isUndercut       bool      // アンダーカット(ノッカーが負け)かどうか
	rng              *rand.Rand
}

// NewTongits コンストラクタ
func NewTongits(trumpCards *TrumpCards, players []*TongitsPlayer, config TongitsConfig) *Tongits {
	return &Tongits{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
		knockerIdx:  -1,
		rng:         rand.New(rand.NewSource(rand.Int63())),
	}
}

// SetRand テスト用に乱数源を差し替える
func (g *Tongits) SetRand(r *rand.Rand) {
	g.rng = r
}

// NewDefaultTongits returns Tongits with the standard 3-player setup (1 human, 2 CPU)
// and DefaultTongitsConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultTongits() *Tongits {
	players := []*TongitsPlayer{
		NewTongitsPlayer(true),
		NewTongitsPlayer(false),
		NewTongitsPlayer(false),
	}
	return NewTongits(NewTrumpCards(0), players, DefaultTongitsConfig())
}

// Reset ゲーム初期化
func (g *Tongits) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0
	g.actionLog = nil
	g.knockerIdx = -1
	g.knockerMelds = nil
	g.knockerDeadwood = nil
	g.opponentMelds = nil
	g.opponentDeadwood = nil
	g.isTongits = false
	g.isUndercut = false

	for _, p := range g.players {
		p.SetRoundScore(0)
		p.SetCumulativeScore(0)
		p.Reset()
		p.SetIsFinished(false)
	}

	g.dealInitialCards()
	g.sortAllHands()

	g.phase = TongitsPhaseDraw
	g.checkTongitsOnDeal()
}

// NextRound 次のラウンドを開始する
func (g *Tongits) NextRound() {
	if g.phase != TongitsPhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0
	g.knockerIdx = -1
	g.knockerMelds = nil
	g.knockerDeadwood = nil
	g.opponentMelds = nil
	g.opponentDeadwood = nil
	g.isTongits = false
	g.isUndercut = false

	for _, p := range g.players {
		p.ResetRound()
	}

	g.dealInitialCards()
	g.sortAllHands()

	g.phase = TongitsPhaseDraw
	g.checkTongitsOnDeal()
}

// dealInitialCards 初期配布: 各プレイヤーにTongitsHandSize枚、1枚を捨て札に
func (g *Tongits) dealInitialCards() {
	// Refill draw counter so a 2nd deal isn't empty; Replenish (not Shuffle) keeps g.rng determinism.
	g.trumpCards.Replenish()

	g.drawPile = make([]*Card, 0, g.trumpCards.GetTotalCount())
	for {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.drawPile = append(g.drawPile, card)
	}

	g.rng.Shuffle(len(g.drawPile), func(i, j int) {
		g.drawPile[i], g.drawPile[j] = g.drawPile[j], g.drawPile[i]
	})

	for i := 0; i < TongitsHandSize; i++ {
		for j := 0; j < TongitsPlayerCnt; j++ {
			if len(g.drawPile) > 0 {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				g.players[j].AddCard(card)
			}
		}
	}
	if len(g.drawPile) > 0 {
		card := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.players[0].AddCard(card)
	}

	if len(g.drawPile) > 0 {
		firstCard := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, firstCard)
	}
}

// checkTongitsOnDeal 配牌Tongits(49/50点即勝利)を判定する
func (g *Tongits) checkTongitsOnDeal() {
	for i, p := range g.players {
		total := 0
		for k := 0; k < p.GetCardsSize(); k++ {
			total += TongitsCardValue(p.GetCard(k))
		}
		if total == TongitsOnDealLow || total == TongitsOnDealHigh {
			g.isTongits = true
			g.knockerIdx = i
			g.appendLog(i, "tongits_on_deal", fmt.Sprintf("%s declares Tongits on deal! (hand value: %d)", playerName(g.players, i), total), nil)
			g.scoreTongits(total)
			return
		}
	}
}

// PlayerDrawFromStock 人間プレイヤーが山札からカードを引く
func (g *Tongits) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TongitsPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return nil
	}

	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, g.currentPlayerIdx)), nil)

	g.phase = TongitsPhaseDiscard
	return nil
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札からカードを引く
func (g *Tongits) PlayerDrawFromDiscard() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TongitsPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(g.discardPile) == 0 {
		return NewDomainError(ErrInvalidPlay, "捨て札がありません")
	}

	card := g.discardPile[len(g.discardPile)-1]
	g.discardPile = g.discardPile[:len(g.discardPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s draws %s from discard", playerName(g.players, g.currentPlayerIdx), cardStr(card)), []*Card{card})

	g.phase = TongitsPhaseDiscard
	return nil
}

// PlayerDiscard 人間プレイヤーがカードを捨てる
func (g *Tongits) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TongitsPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)

	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})

	g.advanceTurn()
	return nil
}

// PlayerMeld publishes a new set or run from the current player's hand.
func (g *Tongits) PlayerMeld(indices []int) error {
	if err := g.validateHumanPlay(TongitsPhaseDiscard); err != nil {
		return err
	}
	player := g.players[g.currentPlayerIdx]
	if len(indices) < 3 {
		return NewDomainError(ErrInvalidPlay, "メルドには3枚以上必要です")
	}
	if err := validateIndexList(indices, player.GetCardsSize()); err != nil {
		return err
	}
	cards := make([]*Card, len(indices))
	for i, idx := range indices {
		cards[i] = player.GetCard(idx)
	}
	if !tongitsIsMeld(cards) {
		return NewDomainError(ErrInvalidPlay, "有効なセットまたはランではありません")
	}
	player.AppendMeld(append([]*Card(nil), cards...))
	player.RemoveCards(indices)
	g.appendLog(g.currentPlayerIdx, "meld", fmt.Sprintf("%s lays a meld", playerName(g.players, g.currentPlayerIdx)), cards)
	if player.GetCardsSize() == 0 {
		g.finishTongits(g.currentPlayerIdx)
	}
	return nil
}

// PlayerSapaw adds one card from the current player's hand to any player's public meld.
func (g *Tongits) PlayerSapaw(targetPlayerIdx, meldIdx, cardIndex int) error {
	if err := g.validateHumanPlay(TongitsPhaseDiscard); err != nil {
		return err
	}
	if targetPlayerIdx < 0 || targetPlayerIdx >= len(g.players) {
		return NewDomainError(ErrInvalidPlay, "対象プレイヤーが不正です")
	}
	target := g.players[targetPlayerIdx]
	if meldIdx < 0 || meldIdx >= len(target.GetMelds()) {
		return NewDomainError(ErrInvalidPlay, "対象メルドが不正です")
	}
	player := g.players[g.currentPlayerIdx]
	card := player.GetCard(cardIndex)
	if card == nil {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	if !tongitsCanAddToMeld(target.GetMeld(meldIdx), card) {
		return NewDomainError(ErrInvalidPlay, "そのメルドに付け足せません")
	}
	target.AddCardToMeld(meldIdx, card)
	player.RemoveCard(cardIndex)
	g.appendLog(g.currentPlayerIdx, "sapaw", fmt.Sprintf("%s adds a card to a public meld", playerName(g.players, g.currentPlayerIdx)), []*Card{card})
	if player.GetCardsSize() == 0 {
		g.finishTongits(g.currentPlayerIdx)
	}
	return nil
}

// PlayerChallenge resolves a draw/challenge when every other player agrees.
// The player with the lowest remaining hand value wins the round.
func (g *Tongits) PlayerChallenge(agreed []bool) error {
	if err := g.validateHumanPlay(TongitsPhaseDiscard); err != nil {
		return err
	}
	if len(agreed) != TongitsPlayerCnt-1 {
		return NewDomainError(ErrInvalidPlay, "応答数が不正です")
	}
	for _, ok := range agreed {
		if !ok {
			g.advanceTurn()
			return nil
		}
	}
	winner, best := -1, int(^uint(0)>>1)
	for i, p := range g.players {
		value := 0
		for j := 0; j < p.GetCardsSize(); j++ {
			value += TongitsCardValue(p.GetCard(j))
		}
		if value < best {
			winner, best = i, value
		}
	}
	g.finishTongits(winner)
	return nil
}

func (g *Tongits) validateHumanPlay(phase TongitsPhase) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != phase {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return nil
}

func (g *Tongits) finishTongits(winner int) {
	g.winnerIdx, g.gameEndFlag, g.phase = winner, true, TongitsPhaseGameEnd
	g.players[winner].SetIsFinished(true)
	g.appendLog(winner, "tongits", fmt.Sprintf("%s wins by Tongits", playerName(g.players, winner)), nil)
}

// TongitsCardValue returns Tongits hand points: ace is 1, face cards are 10.
func TongitsCardValue(card *Card) int {
	if card == nil {
		return 0
	}
	if card.GetValue() > 10 {
		return 10
	}
	return card.GetValue()
}

func tongitsHandCards(p *TongitsPlayer) []*Card {
	cards := make([]*Card, p.GetCardsSize())
	for i := range cards {
		cards[i] = p.GetCard(i)
	}
	return cards
}

func tongitsIsMeld(cards []*Card) bool {
	if len(cards) < 3 {
		return false
	}
	set := true
	for _, c := range cards[1:] {
		if c == nil || c.GetValue() != cards[0].GetValue() {
			set = false
			break
		}
	}
	if set {
		return true
	}
	if cards[0] == nil {
		return false
	}
	values := make(map[int]bool)
	for _, c := range cards {
		if c == nil || c.GetDesign() != cards[0].GetDesign() || values[c.GetValue()] {
			return false
		}
		values[c.GetValue()] = true
	}
	min, max := 14, 0
	for v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return max-min+1 == len(values)
}

func tongitsCanAddToMeld(meld []*Card, card *Card) bool {
	if len(meld) < 3 || card == nil {
		return false
	}
	set := true
	for _, c := range meld[1:] {
		if c.GetValue() != meld[0].GetValue() {
			set = false
			break
		}
	}
	if set {
		return card.GetValue() == meld[0].GetValue()
	}
	if card.GetDesign() != meld[0].GetDesign() {
		return false
	}
	min, max := 14, 0
	for _, c := range meld {
		if c.GetValue() < min {
			min = c.GetValue()
		}
		if c.GetValue() > max {
			max = c.GetValue()
		}
	}
	return card.GetValue() == min-1 || card.GetValue() == max+1
}

// PlayerKnock 人間プレイヤーがノックする (カードを1枚捨ててノック)
func (g *Tongits) PlayerKnock(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TongitsPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	testCards := make([]*Card, 0, player.GetCardsSize()-1)
	for i := 0; i < player.GetCardsSize(); i++ {
		if i != cardIndex {
			testCards = append(testCards, player.GetCard(i))
		}
	}

	melds, deadwood := FindBestMelds(testCards)
	deadwoodValue := CalcDeadwoodValue(deadwood)

	if deadwoodValue > TongitsKnockThreshold {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("デッドウッドが%d点以下でないとノックできません（現在%d点）", TongitsKnockThreshold, deadwoodValue))
	}

	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)

	g.knockerIdx = g.currentPlayerIdx
	g.knockerMelds = melds
	g.knockerDeadwood = deadwood

	g.appendLog(g.currentPlayerIdx, "knock", fmt.Sprintf("%s knocks (deadwood: %d)", playerName(g.players, g.currentPlayerIdx), deadwoodValue), []*Card{discarded})

	g.scoreRound()
	return nil
}

// CpuPlay 現在の手番がCPUの場合にターンを実行
func (g *Tongits) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.phase != TongitsPhaseDraw && g.phase != TongitsPhaseDiscard {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	switch g.phase {
	case TongitsPhaseDraw:
		g.cpuDraw()
	case TongitsPhaseDiscard:
		g.cpuDiscardOrKnock()
	}
}

// cpuDraw CPUがドローする
func (g *Tongits) cpuDraw() {
	if len(g.discardPile) > 0 {
		topDiscard := g.discardPile[len(g.discardPile)-1]

		player := g.players[g.currentPlayerIdx]
		testCards := make([]*Card, player.GetCardsSize()+1)
		for i := 0; i < player.GetCardsSize(); i++ {
			testCards[i] = player.GetCard(i)
		}
		testCards[player.GetCardsSize()] = topDiscard

		_, deadwoodWith := FindBestMelds(testCards)
		dwWith := CalcDeadwoodValue(deadwoodWith)

		currentCards := make([]*Card, player.GetCardsSize())
		for i := 0; i < player.GetCardsSize(); i++ {
			currentCards[i] = player.GetCard(i)
		}
		_, deadwoodWithout := FindBestMelds(currentCards)
		dwWithout := CalcDeadwoodValue(deadwoodWithout)

		shouldPickDiscard := false
		switch g.config.CpuDifficulty {
		case TongitsCpuDifficultyHard:
			shouldPickDiscard = dwWith < dwWithout
		case TongitsCpuDifficultyNormal:
			shouldPickDiscard = dwWith < dwWithout-3
		default:
			shouldPickDiscard = g.rng.Intn(3) == 0
		}

		if shouldPickDiscard {
			card := g.discardPile[len(g.discardPile)-1]
			g.discardPile = g.discardPile[:len(g.discardPile)-1]
			g.players[g.currentPlayerIdx].AddCard(card)
			g.sortHand(g.currentPlayerIdx)
			g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s draws %s from discard", playerName(g.players, g.currentPlayerIdx), cardStr(card)), []*Card{card})
			g.phase = TongitsPhaseDiscard
			return
		}
	}

	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return
	}

	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)
	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, g.currentPlayerIdx)), nil)
	g.phase = TongitsPhaseDiscard
}

// cpuDiscardOrKnock CPUがディスカードまたはノックする
func (g *Tongits) cpuDiscardOrKnock() {
	player := g.players[g.currentPlayerIdx]

	bestDeadwood, bestDiscardIdx := g.GetBestDeadwood(g.currentPlayerIdx)
	if bestDiscardIdx < 0 {
		bestDiscardIdx = 0
	}

	if bestDeadwood <= TongitsKnockThreshold {
		shouldKnock := false
		switch g.config.CpuDifficulty {
		case TongitsCpuDifficultyHard:
			shouldKnock = bestDeadwood <= 3
		case TongitsCpuDifficultyNormal:
			shouldKnock = bestDeadwood <= 4
		default:
			shouldKnock = true
		}

		if shouldKnock {
			testCards := make([]*Card, 0, player.GetCardsSize()-1)
			for j := 0; j < player.GetCardsSize(); j++ {
				if j != bestDiscardIdx {
					testCards = append(testCards, player.GetCard(j))
				}
			}
			melds, deadwood := FindBestMelds(testCards)
			deadwoodValue := CalcDeadwoodValue(deadwood)

			discarded := player.RemoveCard(bestDiscardIdx)
			g.discardPile = append(g.discardPile, discarded)

			g.knockerIdx = g.currentPlayerIdx
			g.knockerMelds = melds
			g.knockerDeadwood = deadwood

			g.appendLog(g.currentPlayerIdx, "knock", fmt.Sprintf("%s knocks (deadwood: %d)", playerName(g.players, g.currentPlayerIdx), deadwoodValue), []*Card{discarded})

			g.scoreRound()
			return
		}
	}

	discarded := player.RemoveCard(bestDiscardIdx)
	g.discardPile = append(g.discardPile, discarded)
	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})
	g.advanceTurn()
}

// scoreRound ノック後のスコアを確定する
func (g *Tongits) scoreRound() {
	knockerIdx := g.knockerIdx
	knockerDeadwoodValue := CalcDeadwoodValue(g.knockerDeadwood)
	bestIdx, bestValue := knockerIdx, knockerDeadwoodValue
	for i, opponent := range g.players {
		if i == knockerIdx {
			continue
		}
		cards := tongitsHandCards(opponent)
		melds, deadwood := FindBestMelds(cards)
		value := CalcDeadwoodValue(deadwood)
		if i == (knockerIdx+1)%TongitsPlayerCnt {
			g.opponentMelds, g.opponentDeadwood = melds, deadwood
		}
		if value < bestValue {
			bestIdx, bestValue = i, value
		}
	}
	if bestIdx != knockerIdx {
		g.isUndercut = true
		score := knockerDeadwoodValue - bestValue + TongitsUndercutPenalty
		g.players[bestIdx].SetRoundScore(score)
		g.appendLog(bestIdx, "undercut", fmt.Sprintf("%s undercuts! Scores %d", playerName(g.players, bestIdx), score), nil)
	} else {
		score := 0
		for i, opponent := range g.players {
			if i == knockerIdx {
				continue
			}
			_, deadwood := FindBestMelds(tongitsHandCards(opponent))
			score += CalcDeadwoodValue(deadwood) - knockerDeadwoodValue
		}
		if score > 0 {
			g.players[knockerIdx].SetRoundScore(score)
		}
		g.appendLog(knockerIdx, "score", fmt.Sprintf("%s scores %d", playerName(g.players, knockerIdx), score), nil)
	}

	for i := range g.players {
		g.players[i].CommitRoundScore()
	}

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = TongitsPhaseRoundEnd
	}
}

// scoreTongits 配牌Tongits成立時のスコア処理
func (g *Tongits) scoreTongits(handValue int) {
	knockerIdx := g.knockerIdx
	// ノッカー(Tongits宣言者)の手札はメルド扱いせず、参考のためそのまま記録する
	g.knockerMelds = nil
	knockerCards := make([]*Card, g.players[knockerIdx].GetCardsSize())
	for i := 0; i < g.players[knockerIdx].GetCardsSize(); i++ {
		knockerCards[i] = g.players[knockerIdx].GetCard(i)
	}
	g.knockerDeadwood = knockerCards
	for i, opponent := range g.players {
		if i == knockerIdx {
			continue
		}
		melds, deadwood := FindBestMelds(tongitsHandCards(opponent))
		if g.opponentMelds == nil {
			g.opponentMelds, g.opponentDeadwood = melds, deadwood
		}
	}

	score := TongitsBonus + handValue
	g.players[knockerIdx].SetRoundScore(score)
	g.appendLog(knockerIdx, "tongits_score", fmt.Sprintf("%s scores %d (Tongits bonus %d + hand %d)", playerName(g.players, knockerIdx), score, TongitsBonus, handValue), nil)

	for i := range g.players {
		g.players[i].CommitRoundScore()
	}

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = TongitsPhaseRoundEnd
	}
}

// endRoundDraw 山札切れによる引き分け (スコアなし)
func (g *Tongits) endRoundDraw() {
	g.appendLog(-1, "draw", "Round ends in a draw (stock empty)", nil)
	g.knockerIdx = -1

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = TongitsPhaseRoundEnd
	}
}

// ScoreRound ラウンドのスコア処理 (NextRoundから呼ぶ用。既にscoreRoundで処理済みの場合はnoop)
func (g *Tongits) ScoreRound() {
	// スコアリングはknock/tongits時に既に完了している
	// このメソッドは互換性のために存在
}

// advanceTurn 次のプレイヤーへ
func (g *Tongits) advanceTurn() {
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % TongitsPlayerCnt
	g.phase = TongitsPhaseDraw
}

// checkGameEnd ゲーム終了判定
func (g *Tongits) checkGameEnd() {
	hasWinner := false
	for i := 0; i < TongitsPlayerCnt; i++ {
		if g.players[i].GetCumulativeScore() >= g.config.PointLimit {
			hasWinner = true
			break
		}
	}

	if !hasWinner {
		return
	}

	g.gameEndFlag = true
	g.phase = TongitsPhaseGameEnd

	maxScore := g.players[0].GetCumulativeScore()
	g.winnerIdx = 0
	for i := 1; i < TongitsPlayerCnt; i++ {
		if g.players[i].GetCumulativeScore() > maxScore {
			maxScore = g.players[i].GetCumulativeScore()
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(g.players, g.winnerIdx)), nil)
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
// GetBestDeadwood は1枚捨てたときに到達できる最小デッドウッド値と、その捨て札の
// 位置を返す。手札が空なら (0, -1)。
//
// **この計算は元々3箇所に散る寸前だった。**CPU の判断 (cpuDiscardOrKnock) と
// CUI の表示 (tongitsBestDeadwood) が別々に同じループを持っており、Web にも
// 3つ目を書くところだった。TongitsKnockThreshold と比べる値なので、実装が割れると
// 「ノック可能と表示したのに弾かれる」ずれになる。
func (g *Tongits) GetBestDeadwood(playerIdx int) (best int, discardIdx int) {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return 0, -1
	}
	player := g.players[playerIdx]
	n := player.GetCardsSize()
	best, discardIdx = -1, -1
	for i := 0; i < n; i++ {
		sub := make([]*Card, 0, n-1)
		for j := 0; j < n; j++ {
			if j != i {
				sub = append(sub, player.GetCard(j))
			}
		}
		_, dw := FindBestMelds(sub)
		if v := CalcDeadwoodValue(dw); best < 0 || v < best {
			best, discardIdx = v, i
		}
	}
	if best < 0 {
		return 0, -1
	}
	return best, discardIdx
}

func (g *Tongits) GetPhase() TongitsPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Tongits) SetPhase(phase TongitsPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Tongits) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Tongits) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Tongits) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Tongits) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *Tongits) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *Tongits) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDiscardTop 捨て札の一番上を取得
func (g *Tongits) GetDiscardTop() *Card {
	return discardTop(g.discardPile)
}

// GetDrawPileCount 山札の残り枚数取得
func (g *Tongits) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定 (テスト用)
func (g *Tongits) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Tongits) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (g *Tongits) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Tongits) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Tongits) GetPlayer(i int) *TongitsPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Tongits) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Tongits) GetConfig() TongitsConfig { return g.config }

// SetConfig 設定変更
func (g *Tongits) SetConfig(cfg TongitsConfig) { g.config = cfg }

// GetKnockerIdx ノッカーのインデックス取得
func (g *Tongits) GetKnockerIdx() int { return g.knockerIdx }

// SetKnockerIdx ノッカーのインデックス設定 (テスト用)
func (g *Tongits) SetKnockerIdx(idx int) { g.knockerIdx = idx }

// GetKnockerMelds ノッカーのメルド取得
func (g *Tongits) GetKnockerMelds() [][]*Card { return g.knockerMelds }

// SetKnockerMelds ノッカーのメルドを設定 (テスト用)
func (g *Tongits) SetKnockerMelds(melds [][]*Card) { g.knockerMelds = melds }

// GetKnockerDeadwood ノッカーのデッドウッド取得
func (g *Tongits) GetKnockerDeadwood() []*Card { return g.knockerDeadwood }

// SetKnockerDeadwood ノッカーのデッドウッドを設定 (テスト用)
func (g *Tongits) SetKnockerDeadwood(deadwood []*Card) { g.knockerDeadwood = deadwood }

// GetOpponentMelds 相手側のメルド取得
func (g *Tongits) GetOpponentMelds() [][]*Card { return g.opponentMelds }

// GetOpponentDeadwood 相手側のデッドウッド取得
func (g *Tongits) GetOpponentDeadwood() []*Card { return g.opponentDeadwood }

// GetIsTongits 配牌Tongitsかどうか取得
func (g *Tongits) GetIsTongits() bool { return g.isTongits }

// SetIsTongits 配牌Tongits設定 (テスト用)
func (g *Tongits) SetIsTongits(isTongits bool) { g.isTongits = isTongits }

// GetIsUndercut アンダーカットかどうか取得
func (g *Tongits) GetIsUndercut() bool { return g.isUndercut }

// --- Private methods ---

// sortAllHands 全プレイヤーの手札をソートする
func (g *Tongits) sortAllHands() {
	sortHands(len(g.players), g)
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func (g *Tongits) sortHand(playerIdx int) {
	sortPlayerHand(g.players[playerIdx], bySuitThenValue)
}

// tongitsJSON is the JSON wire format for Tongits.
type tongitsJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*TongitsPlayer  `json:"pl"`
	Config           TongitsConfig     `json:"cf"`
	Phase            TongitsPhase      `json:"ps"`
	CurrentPlayerIdx int               `json:"ci"`
	DiscardPile      []*Card           `json:"dp"`
	DrawPile         []*Card           `json:"wp"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	RoundNumber      int               `json:"rn"`
	ActionLog        []*ActionLogEntry `json:"al"`
	KnockerIdx       int               `json:"ki"`
	KnockerMelds     [][]*Card         `json:"km"`
	KnockerDeadwood  []*Card           `json:"kd"`
	OpponentMelds    [][]*Card         `json:"om"`
	OpponentDeadwood []*Card           `json:"od"`
	IsTongits        bool              `json:"it"`
	IsUndercut       bool              `json:"iu"`
}

// MarshalJSON implements json.Marshaler.
func (g *Tongits) MarshalJSON() ([]byte, error) {
	return json.Marshal(tongitsJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		ActionLog:        g.actionLog,
		KnockerIdx:       g.knockerIdx,
		KnockerMelds:     g.knockerMelds,
		KnockerDeadwood:  g.knockerDeadwood,
		OpponentMelds:    g.opponentMelds,
		OpponentDeadwood: g.opponentDeadwood,
		IsTongits:        g.isTongits,
		IsUndercut:       g.isUndercut,
	})
}

// tongitsMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const tongitsMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (g *Tongits) UnmarshalJSON(data []byte) error {
	var j tongitsJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > tongitsMaxSliceLen || len(j.DiscardPile) > tongitsMaxSliceLen ||
		len(j.DrawPile) > tongitsMaxSliceLen || len(j.ActionLog) > tongitsMaxSliceLen ||
		len(j.KnockerMelds) > tongitsMaxSliceLen || len(j.KnockerDeadwood) > tongitsMaxSliceLen ||
		len(j.OpponentMelds) > tongitsMaxSliceLen || len(j.OpponentDeadwood) > tongitsMaxSliceLen {
		return fmt.Errorf("tongits: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*TongitsPlayer, 0)
	}
	g.config = j.Config
	g.phase = j.Phase
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.discardPile = j.DiscardPile
	if g.discardPile == nil {
		g.discardPile = make([]*Card, 0)
	}
	g.drawPile = j.DrawPile
	if g.drawPile == nil {
		g.drawPile = make([]*Card, 0)
	}
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.roundNumber = j.RoundNumber
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	g.knockerIdx = j.KnockerIdx
	g.knockerMelds = j.KnockerMelds
	if g.knockerMelds == nil {
		g.knockerMelds = make([][]*Card, 0)
	}
	g.knockerDeadwood = j.KnockerDeadwood
	if g.knockerDeadwood == nil {
		g.knockerDeadwood = make([]*Card, 0)
	}
	g.opponentMelds = j.OpponentMelds
	if g.opponentMelds == nil {
		g.opponentMelds = make([][]*Card, 0)
	}
	g.opponentDeadwood = j.OpponentDeadwood
	if g.opponentDeadwood == nil {
		g.opponentDeadwood = make([]*Card, 0)
	}
	g.isTongits = j.IsTongits
	g.isUndercut = j.IsUndercut
	// **復元したら必ず乱数源を張り直す。**Cloudflare Worker は毎リクエスト KV から
	// 組み直すので SetRand は一度も呼ばれない。rng を nil のままにすると、
	// シャッフル以外で rng を使う経路 (CPU の乱択など) が nil デリファレンスで
	// 落ちる。呼び出し側ごとにガードするのではなく、ここで構造的に潰す。
	g.rng = rand.New(rand.NewSource(rand.Int63()))
	return nil
}
