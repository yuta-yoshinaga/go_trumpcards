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

// TongitsOnDealLow 配牌時のTongits(即勝利)成立点(下限)
const TongitsOnDealLow = 49

// TongitsOnDealHigh 配牌時のTongits(即勝利)成立点(上限)
const TongitsOnDealHigh = 50

// TongitsBonus 配牌時Tongits成立時のボーナス点
const TongitsBonus = 50

// TongitsPhase ゲームフェーズ
type TongitsPhase int

// Tongitsのフェーズ定数
const (
	// TongitsPhaseDraw ドローフェーズ (山札または捨て札から引く)
	TongitsPhaseDraw TongitsPhase = 0
	// TongitsPhaseDiscard ディスカードフェーズ (手札から1枚捨てる or challenge)
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
	isTongits bool // 配牌Tongits(49/50点即勝利)かどうか
	rng       *rand.Rand
}

// NewTongits コンストラクタ
func NewTongits(trumpCards *TrumpCards, players []*TongitsPlayer, config TongitsConfig) *Tongits {
	return &Tongits{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
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
	g.isTongits = false

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
	g.isTongits = false

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
			g.appendLog(i, "tongits_on_deal", fmt.Sprintf("%s declares Tongits on deal! (hand value: %d)", playerName(g.players, i), total), nil)
			g.scoreTongits(i, total)
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
	return g.resolveChallenge(agreed)
}

func (g *Tongits) resolveChallenge(agreed []bool) error {
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

// tongitsRemainingPoints は手札の残り点を返す。ドロー宣言 (challenge) の勝敗は
// これの少なさで決まる。1 枚あたりの規則は TongitsCardValue が持つ (絵札 10 / A 1 /
// 数札は数字通り) ので、ここはその総和に徹する -- 点数の規則が 2 箇所に割れると、
// CPU の判断と実際の決着が食い違う。
func tongitsRemainingPoints(cards []*Card) int {
	total := 0
	for _, c := range cards {
		total += TongitsCardValue(c)
	}
	return total
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
		g.cpuDiscardOrChallenge()
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

		remainingWith := tongitsRemainingPoints(testCards)

		currentCards := make([]*Card, player.GetCardsSize())
		for i := 0; i < player.GetCardsSize(); i++ {
			currentCards[i] = player.GetCard(i)
		}
		remainingWithout := tongitsRemainingPoints(currentCards)

		shouldPickDiscard := false
		switch g.config.CpuDifficulty {
		case TongitsCpuDifficultyHard:
			shouldPickDiscard = remainingWith < remainingWithout
		case TongitsCpuDifficultyNormal:
			shouldPickDiscard = remainingWith < remainingWithout-3
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

// cpuDiscardOrChallenge CPUがディスカードまたはchallengeする
func (g *Tongits) cpuDiscardOrChallenge() {
	player := g.players[g.currentPlayerIdx]
	if player.GetCardsSize() == 0 {
		g.finishTongits(g.currentPlayerIdx)
		return
	}

	if g.config.CpuDifficulty == TongitsCpuDifficultyHard && tongitsRemainingPoints(tongitsHandCards(player)) <= 10 {
		agreed := make([]bool, TongitsPlayerCnt-1)
		for i := range agreed {
			agreed[i] = true
		}
		if err := g.resolveChallenge(agreed); err == nil {
			return
		}
	}

	discardIdx := 0
	for i := 1; i < player.GetCardsSize(); i++ {
		if TongitsCardValue(player.GetCard(i)) > TongitsCardValue(player.GetCard(discardIdx)) {
			discardIdx = i
		}
	}
	discarded := player.RemoveCard(discardIdx)
	g.discardPile = append(g.discardPile, discarded)
	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})
	g.advanceTurn()
}

// scoreTongits 配牌Tongits成立時のスコア処理

func (g *Tongits) scoreTongits(winner int, handValue int) {
	score := TongitsBonus + handValue
	g.players[winner].SetRoundScore(score)
	g.appendLog(winner, "tongits_score", fmt.Sprintf("%s scores %d (Tongits bonus %d + hand %d)", playerName(g.players, winner), score, TongitsBonus, handValue), nil)

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

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = TongitsPhaseRoundEnd
	}
}

// ScoreRound ラウンドのスコア処理 (NextRoundから呼ぶ用。既にscoreRoundで処理済みの場合はnoop)
func (g *Tongits) ScoreRound() {
	// Challenge または Tongits の宣言時にスコアリングは完了している。
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

// GetIsTongits 配牌Tongitsかどうか取得
func (g *Tongits) GetIsTongits() bool { return g.isTongits }

// SetIsTongits 配牌Tongits設定 (テスト用)
func (g *Tongits) SetIsTongits(isTongits bool) { g.isTongits = isTongits }

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
	IsTongits        bool              `json:"it"`
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
		IsTongits:        g.isTongits,
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
		len(j.DrawPile) > tongitsMaxSliceLen || len(j.ActionLog) > tongitsMaxSliceLen {
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
	g.isTongits = j.IsTongits
	// **復元したら必ず乱数源を張り直す。**Cloudflare Worker は毎リクエスト KV から
	// 組み直すので SetRand は一度も呼ばれない。rng を nil のままにすると、
	// シャッフル以外で rng を使う経路 (CPU の乱択など) が nil デリファレンスで
	// 落ちる。呼び出し側ごとにガードするのではなく、ここで構造的に潰す。
	g.rng = rand.New(rand.NewSource(rand.Int63()))
	return nil
}
