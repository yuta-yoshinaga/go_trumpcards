// Package domain ブリスコラ (Briscola) のドメインモデル。
//
// Briscola はイタリアの古典的なトリックテイキングゲーム。40枚デッキ
// (8,9,10 を除く) を使い、本実装では 2 人対戦のみを扱う。最大の特徴は
// 「リードスートに従う義務 (must-follow) がない」ことで、validatePlay は
// 常に成功する。トリックの勝者はトランプ (briscola) > リードスートの
// ブリスコラ順位 で決まり、ブリスコラ順位は A>3>K>Q>J>7>6>5>4>2 となる。
// 各カードには独自の点数 (A=11, 3=10, K=4, Q=3, J=2, それ以外=0) が
// あり、合計 120 点を 2 人で取り合う。60 点を超えた側が勝者で、
// 60-60 は引き分け。
package domain

import (
	"encoding/json"
	"fmt"
)

// BriscolaPlayerCnt ブリスコラのプレイヤー数 (v1は2人固定)
const BriscolaPlayerCnt = 2

// BriscolaHandSize 各プレイヤーの手札最大枚数 (山札がある間は補充される)
const BriscolaHandSize = 3

// BriscolaWinThreshold 勝利点 (これを超える点数で勝ち、超えない側は負け)
const BriscolaWinThreshold = 60

// BriscolaTotalPoints デッキ全体の合計点
const BriscolaTotalPoints = 120

// BriscolaPhase ゲームフェーズ
type BriscolaPhase int

// Briscolaのフェーズ定数
const (
	// BriscolaPhasePlay トリックプレイフェーズ
	BriscolaPhasePlay BriscolaPhase = iota
	// BriscolaPhaseTrickEnd トリック終了フェーズ
	BriscolaPhaseTrickEnd
	// BriscolaPhaseGameEnd ゲーム終了フェーズ
	BriscolaPhaseGameEnd
)

// BriscolaHint ヒント情報
type BriscolaHint struct {
	CardIndex *int   // 推奨カードインデックス
	Reason    string // ヒント理由キー
}

// briscolaCardPoints ブリスコラのカード点数 (Ace=1, 3, K=13, Q=12, J=11, それ以外=0)
var briscolaCardPoints = map[int]int{
	1:  11, // Asso
	3:  10, // Tre
	13: 4,  // Re
	12: 3,  // Cavallo / Donna
	11: 2,  // Fante
}

// briscolaRankOrder スート内のカード強さ。値が大きいほど強い。
// A>3>K>Q>J>7>6>5>4>2 を 1-base で表現する。
var briscolaRankOrder = map[int]int{
	2:  1,
	4:  2,
	5:  3,
	6:  4,
	7:  5,
	11: 6,  // J
	12: 7,  // Q
	13: 8,  // K
	3:  9,  // 3
	1:  10, // A
}

// BriscolaCardPoints カードの得点を返す (公開ヘルパー)。
func BriscolaCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	return briscolaCardPoints[c.GetValue()]
}

// BriscolaRankOrder カードのスート内順位を返す (大きいほど強い)。
func BriscolaRankOrder(c *Card) int {
	if c == nil {
		return 0
	}
	return briscolaRankOrder[c.GetValue()]
}

// Briscola ブリスコラゲームクラス
type Briscola struct {
	trumpCards       *TrumpCards
	players          []*BriscolaPlayer
	config           BriscolaConfig
	phase            BriscolaPhase
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	trumpCard        *Card // 場に表向きで置かれるトランプ (山札の最後)
	trumpSuit        int
	leadPlayerIdx    int
	dealerIdx        int
	playerPoints     []int
	gameEndFlag      bool
	winnerIdx        int // -1: 未確定または引き分け
	actionLog        []*ActionLogEntry
}

// NewBriscola コンストラクタ
func NewBriscola(trumpCards *TrumpCards, players []*BriscolaPlayer, config BriscolaConfig) *Briscola {
	return &Briscola{
		trumpCards:   trumpCards,
		players:      players,
		config:       config,
		winnerIdx:    -1,
		playerPoints: make([]int, len(players)),
	}
}

// NewDefaultBriscola 標準の 2 人対戦セットアップを返す。
// 人間プレイヤー (idx 0) と CPU (idx 1) の組み合わせ。
func NewDefaultBriscola() *Briscola {
	players := []*BriscolaPlayer{
		NewBriscolaPlayer(true),
		NewBriscolaPlayer(false),
	}
	return NewBriscola(NewTrumpCardsBriscola(), players, DefaultBriscolaConfig())
}

// Reset ゲーム初期化
func (b *Briscola) Reset() {
	b.gameEndFlag = false
	b.winnerIdx = -1
	b.trickNumber = 0
	b.currentTrick = nil
	b.leadPlayerIdx = -1
	b.currentPlayerIdx = -1
	b.dealerIdx = 0
	b.playerPoints = make([]int, len(b.players))
	b.actionLog = nil
	b.trumpCard = nil
	b.trumpSuit = 0

	for _, p := range b.players {
		p.ResetGame()
	}

	b.trumpCards.Shuffle()
	b.dealInitial()
	b.sortAllHands()

	b.startPlayPhase()
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (b *Briscola) PlayerPlay(cardIndex int) error {
	if b.gameEndFlag {
		return ErrGameEnded
	}
	if b.phase != BriscolaPhasePlay {
		return ErrWrongPhase
	}
	if !b.players[b.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := b.players[b.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := b.validatePlay(b.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	b.playCard(b.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する
func (b *Briscola) CpuPlay() {
	if b.gameEndFlag || b.phase != BriscolaPhasePlay {
		return
	}
	if b.players[b.currentPlayerIdx].GetIsHuman() {
		return
	}
	player := b.players[b.currentPlayerIdx]
	cardIdx := b.cpuSelectPlayCard(b.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	b.playCard(b.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (b *Briscola) ResolveTrick() {
	if b.phase != BriscolaPhaseTrickEnd || len(b.currentTrick) != BriscolaPlayerCnt {
		return
	}

	winnerIdx := b.trickWinner()
	trickCards := make([]*Card, len(b.currentTrick))
	trickPoints := 0
	for i, tc := range b.currentTrick {
		trickCards[i] = tc.Card
		trickPoints += BriscolaCardPoints(tc.Card)
	}

	b.players[winnerIdx].AddTrick(trickCards)
	b.playerPoints[winnerIdx] += trickPoints

	b.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (%d pt)", b.playerName(winnerIdx), b.trickNumber, trickPoints),
		trickCards)

	b.leadPlayerIdx = winnerIdx
	// Phase is already BriscolaPhaseTrickEnd (guarded at function entry); leave it.
}

// NextTrick 次のトリックを開始する。山札が残っていれば補充も行う。
// 全カードが尽きたらゲーム終了処理を実行する。
func (b *Briscola) NextTrick() {
	if b.phase != BriscolaPhaseTrickEnd {
		return
	}

	b.drawReplenish()

	if b.allHandsEmpty() {
		b.finishGame()
		return
	}

	b.currentTrick = nil
	b.currentPlayerIdx = b.leadPlayerIdx
	b.trickNumber++
	b.phase = BriscolaPhasePlay
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (b *Briscola) GetPhase() BriscolaPhase { return b.phase }

// SetPhase フェーズ設定 (テスト用)
func (b *Briscola) SetPhase(phase BriscolaPhase) { b.phase = phase }

// GetTrickNumber 現在のトリック番号取得
func (b *Briscola) GetTrickNumber() int { return b.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (b *Briscola) SetTrickNumber(n int) { b.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (b *Briscola) GetCurrentPlayerIdx() int { return b.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (b *Briscola) SetCurrentPlayerIdx(idx int) { b.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (b *Briscola) GetCurrentTrick() []*TrickCard { return b.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (b *Briscola) SetCurrentTrick(trick []*TrickCard) { b.currentTrick = trick }

// GetTrumpSuit トランプスート取得
func (b *Briscola) GetTrumpSuit() int { return b.trumpSuit }

// SetTrumpSuit トランプスート設定 (テスト用)
func (b *Briscola) SetTrumpSuit(suit int) { b.trumpSuit = suit }

// GetTrumpCard 場に表向きで置かれているトランプカードを取得 (山札に残っていなければ nil)
func (b *Briscola) GetTrumpCard() *Card { return b.trumpCard }

// SetTrumpCard トランプカード設定 (テスト用)
func (b *Briscola) SetTrumpCard(c *Card) { b.trumpCard = c }

// GetGameEndFlag ゲーム終了フラグ取得
func (b *Briscola) GetGameEndFlag() bool { return b.gameEndFlag }

// SetGameEndFlag ゲーム終了フラグ設定 (テスト用)
func (b *Briscola) SetGameEndFlag(flag bool) { b.gameEndFlag = flag }

// GetWinnerIdx 勝者プレイヤーインデックス (-1: 未確定または引き分け)
func (b *Briscola) GetWinnerIdx() int { return b.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (b *Briscola) GetPlayerCnt() int { return len(b.players) }

// GetPlayer プレイヤー取得
func (b *Briscola) GetPlayer(i int) *BriscolaPlayer {
	if i < 0 || i >= len(b.players) {
		return nil
	}
	return b.players[i]
}

// GetPlayerPoints プレイヤーの累積得点取得
func (b *Briscola) GetPlayerPoints(i int) int {
	if i < 0 || i >= len(b.playerPoints) {
		return 0
	}
	return b.playerPoints[i]
}

// SetPlayerPoints プレイヤー得点設定 (テスト用)
func (b *Briscola) SetPlayerPoints(i, points int) {
	if i >= 0 && i < len(b.playerPoints) {
		b.playerPoints[i] = points
	}
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (b *Briscola) GetLeadPlayerIdx() int { return b.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (b *Briscola) SetLeadPlayerIdx(idx int) { b.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (b *Briscola) GetDealerIdx() int { return b.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (b *Briscola) SetDealerIdx(idx int) { b.dealerIdx = idx }

// GetStockRemaining 山札の残り枚数 (場に出ている表向きトランプは含まない;
// それは GetTrumpCard() != nil の間 別カウントとして残る最後の 1 枚)。
func (b *Briscola) GetStockRemaining() int {
	return b.trumpCards.GetRemainingCount()
}

// IsHumanTurn 現在の手番が人間かどうか
func (b *Briscola) IsHumanTurn() bool {
	if b.currentPlayerIdx < 0 || b.currentPlayerIdx >= len(b.players) {
		return false
	}
	return b.players[b.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (b *Briscola) GetConfig() BriscolaConfig { return b.config }

// SetConfig 設定変更
func (b *Briscola) SetConfig(cfg BriscolaConfig) { b.config = cfg }

// GetActionLog 棋譜取得
func (b *Briscola) GetActionLog() []*ActionLogEntry { return b.actionLog }

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す。
// Briscola には must-follow 制約がないため、現在の手札全てが対象。
func (b *Briscola) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(b.players) {
		return nil
	}
	p := b.players[playerIdx]
	out := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		out = append(out, i)
	}
	return out
}

// GetHint 人間プレイヤーへのヒントを取得する
func (b *Briscola) GetHint() *BriscolaHint {
	if b.phase != BriscolaPhasePlay || b.currentPlayerIdx != 0 {
		return nil
	}
	humanIdx := 0
	if b.players[humanIdx].GetCardsSize() == 0 {
		return nil
	}
	idx := b.cpuSelectPlayCard(humanIdx)
	return &BriscolaHint{CardIndex: &idx, Reason: b.playHintReason(humanIdx, idx)}
}

// --- Private methods ---

// dealInitial 各プレイヤーに 3 枚配り、その次の 1 枚を表向きトランプとして山札の底に置く。
func (b *Briscola) dealInitial() {
	for range BriscolaHandSize {
		for i := range BriscolaPlayerCnt {
			player := b.players[(b.dealerIdx+1+i)%BriscolaPlayerCnt]
			if c := b.trumpCards.DrawCard(); c != nil {
				player.AddCard(c)
			}
		}
	}
	// 次の 1 枚をトランプとして表向きに置く (デッキの底相当: 最後に引かれる)
	b.trumpCard = b.trumpCards.DrawCard()
	if b.trumpCard != nil {
		b.trumpSuit = b.trumpCard.GetDesign()
		b.appendLog(-1, "trump", fmt.Sprintf("Trump: %s", cardStr(b.trumpCard)), []*Card{b.trumpCard})
	}
}

// startPlayPhase プレイフェーズ開始: ディーラーの左隣がリード
func (b *Briscola) startPlayPhase() {
	b.trickNumber = 1
	b.currentTrick = nil
	b.leadPlayerIdx = (b.dealerIdx + 1) % BriscolaPlayerCnt
	b.currentPlayerIdx = b.leadPlayerIdx
	b.phase = BriscolaPhasePlay
}

// playCard カードをプレイする共通処理
func (b *Briscola) playCard(playerIdx int, card *Card) {
	b.currentTrick = append(b.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})
	b.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", b.playerName(playerIdx), cardStr(card)),
		[]*Card{card})

	if len(b.currentTrick) == BriscolaPlayerCnt {
		b.phase = BriscolaPhaseTrickEnd
	} else {
		b.currentPlayerIdx = (b.currentPlayerIdx + 1) % BriscolaPlayerCnt
	}
}

// validatePlay カードのプレイがルール上有効かを検証する。
// Briscola には must-follow がないため、プレイヤーが手札に持つカードであれば常に有効。
func (b *Briscola) validatePlay(_ int, card *Card) error {
	if card == nil {
		return NewDomainError(ErrInvalidCard, "カードが nil です")
	}
	return nil
}

// trickWinner 現在のトリックの勝者インデックスを決定する
func (b *Briscola) trickWinner() int {
	if len(b.currentTrick) == 0 {
		return 0
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	winnerIdx := b.currentTrick[0].PlayerIdx
	winnerCard := b.currentTrick[0].Card

	for _, tc := range b.currentTrick[1:] {
		if briscolaBeats(tc.Card, winnerCard, leadSuit, b.trumpSuit) {
			winnerIdx = tc.PlayerIdx
			winnerCard = tc.Card
		}
	}
	return winnerIdx
}

// briscolaBeats challenger が currentBest に勝つかを判定する。
// ・両者がトランプ: ブリスコラ順位の高い方が勝つ
// ・challenger のみトランプ: challenger が勝つ
// ・両者とも非トランプかつ同じリードスート: ブリスコラ順位の高い方が勝つ
// ・両者とも非トランプで challenger がリードスート以外: challenger は勝てない
func briscolaBeats(challenger, currentBest *Card, leadSuit, trumpSuit int) bool {
	cIsTrump := challenger.GetDesign() == trumpSuit
	bIsTrump := currentBest.GetDesign() == trumpSuit

	switch {
	case cIsTrump && bIsTrump:
		return BriscolaRankOrder(challenger) > BriscolaRankOrder(currentBest)
	case cIsTrump:
		return true
	case bIsTrump:
		return false
	}
	// ともに非トランプ
	if challenger.GetDesign() != leadSuit {
		return false
	}
	if currentBest.GetDesign() != leadSuit {
		return true
	}
	return BriscolaRankOrder(challenger) > BriscolaRankOrder(currentBest)
}

// drawReplenish トリック勝者が先に 1 枚、次に敗者が 1 枚を山札から引く。
// 山札が空になっていく過程では、最後の 1 枚は表向きトランプ (trumpCard) を引いた扱いになる。
func (b *Briscola) drawReplenish() {
	if b.trumpCards.GetRemainingCount() == 0 && b.trumpCard == nil {
		return
	}
	winnerIdx := b.leadPlayerIdx
	loserIdx := (winnerIdx + 1) % BriscolaPlayerCnt
	for _, idx := range []int{winnerIdx, loserIdx} {
		if c := b.drawOne(); c != nil {
			b.players[idx].AddCard(c)
			b.sortHand(b.players[idx])
		}
	}
}

// drawOne 山札またはトランプカードから 1 枚引く。優先順位は山札 → トランプカード。
func (b *Briscola) drawOne() *Card {
	if c := b.trumpCards.DrawCard(); c != nil {
		return c
	}
	if b.trumpCard != nil {
		c := b.trumpCard
		b.trumpCard = nil
		return c
	}
	return nil
}

// allHandsEmpty 全プレイヤーの手札が空かを返す
func (b *Briscola) allHandsEmpty() bool {
	for _, p := range b.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// finishGame ゲームを終了させ、勝者を決定する
func (b *Briscola) finishGame() {
	b.gameEndFlag = true
	b.phase = BriscolaPhaseGameEnd
	b.winnerIdx = BriscolaDetermineWinner(b.playerPoints[0], b.playerPoints[1])
	detail := fmt.Sprintf("Game end: %d-%d", b.playerPoints[0], b.playerPoints[1])
	b.appendLog(-1, "game_end", detail, nil)
}

// BriscolaDetermineWinner 二人ブリスコラの勝者を決定する。
// 60 を超えた側が勝ち。両者 <=60 (典型的には 60-60) は -1 (引き分け) を返す。
func BriscolaDetermineWinner(p0, p1 int) int {
	switch {
	case p0 > BriscolaWinThreshold:
		return 0
	case p1 > BriscolaWinThreshold:
		return 1
	default:
		return -1
	}
}

// sortAllHands 全プレイヤーの手札をソートする
func (b *Briscola) sortAllHands() {
	for _, p := range b.players {
		b.sortHand(p)
	}
}

// sortHand プレイヤーの手札をスート (トランプ最後) → ブリスコラ順位 でソートする
func (b *Briscola) sortHand(p *BriscolaPlayer) {
	trumpSuit := b.trumpSuit
	sortPlayerHand(p, func(ci, cj *Card) bool {
		iTrump := ci.GetDesign() == trumpSuit
		jTrump := cj.GetDesign() == trumpSuit
		if iTrump != jTrump {
			return !iTrump // 非トランプを先に
		}
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return BriscolaRankOrder(ci) < BriscolaRankOrder(cj)
	})
}

// playerName プレイヤー名を返す (ログ用)
func (b *Briscola) playerName(idx int) string {
	if idx < 0 || idx >= len(b.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if b.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// appendLog 棋譜エントリを追加する
func (b *Briscola) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	b.actionLog = append(b.actionLog, &ActionLogEntry{
		TurnNumber: len(b.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// playHintReason ヒント理由キーを判定する
func (b *Briscola) playHintReason(playerIdx, chosenIdx int) string {
	card := b.players[playerIdx].GetCard(chosenIdx)
	pts := BriscolaCardPoints(card)
	if len(b.currentTrick) == 0 {
		if card.GetDesign() == b.trumpSuit {
			return "lead_trump"
		}
		if pts == 0 {
			return "lead_low"
		}
		return "lead_value"
	}
	leadCard := b.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()
	if briscolaBeats(card, leadCard, leadSuit, b.trumpSuit) {
		if card.GetDesign() == b.trumpSuit && leadSuit != b.trumpSuit {
			return "follow_cut"
		}
		return "follow_win"
	}
	return "follow_dump"
}

// --- CPU AI (single-difficulty heuristic) ---

// cpuSelectPlayCard CPU が出すべきカードのインデックスを選択する
func (b *Briscola) cpuSelectPlayCard(playerIdx int) int {
	player := b.players[playerIdx]
	if player.GetCardsSize() == 1 {
		return 0
	}

	if len(b.currentTrick) == 0 {
		return b.cpuLead(playerIdx)
	}
	return b.cpuFollow(playerIdx)
}

// cpuLead リード時の選択: 最も低い点数の非トランプを優先する。
// 全カードがトランプ・点数札しか無い場合は最も弱い順位のカードを選ぶ。
func (b *Briscola) cpuLead(playerIdx int) int {
	player := b.players[playerIdx]
	bestIdx := 0
	bestScore := cpuLeadScore(player.GetCard(0), b.trumpSuit)
	for i := 1; i < player.GetCardsSize(); i++ {
		s := cpuLeadScore(player.GetCard(i), b.trumpSuit)
		if s < bestScore {
			bestScore = s
			bestIdx = i
		}
	}
	return bestIdx
}

// cpuLeadScore 値が小さいほど「リードに適している」(トランプを温存し、点数の高い札を温存する)
func cpuLeadScore(c *Card, trumpSuit int) int {
	score := BriscolaCardPoints(c)*10 + BriscolaRankOrder(c)
	if c.GetDesign() == trumpSuit {
		score += 1000
	}
	return score
}

// cpuFollow 追随時の選択。
// 1) リードがトランプ: 勝てる最小トランプ、無ければ最小点数札を捨てる
// 2) リードが点数札 (A/3): 勝てる最小トランプ、無ければ最小点数札を捨てる
// 3) リードが低点数: 同スートの最小勝ち札 (非トランプ) があればそれ、無ければ最小点数札を捨てる
//
// いずれの分岐でも、トリック既出点数が高い (>= 11) 場合は積極的にトランプで奪取する。
func (b *Briscola) cpuFollow(playerIdx int) int {
	player := b.players[playerIdx]
	leadCard := b.currentTrick[0].Card
	leadSuit := leadCard.GetDesign()
	trickPoints := BriscolaCardPoints(leadCard)

	// 既存の同スート勝ちカード (非トランプ前提) を探す
	if leadSuit != b.trumpSuit {
		if idx := smallestSameSuitWinner(player, leadCard, leadSuit); idx >= 0 && trickPoints == 0 {
			return idx
		}
	}

	// 高点数または相手がトランプ → トランプで取りに行く価値あり
	if trickPoints >= 10 || leadCard.GetDesign() == b.trumpSuit {
		if idx := smallestWinningTrump(player, leadCard, leadSuit, b.trumpSuit); idx >= 0 {
			return idx
		}
	}

	// 同スート勝ちで点数があるなら奪取
	if leadSuit != b.trumpSuit {
		if idx := smallestSameSuitWinner(player, leadCard, leadSuit); idx >= 0 {
			return idx
		}
	}

	return smallestDump(player, b.trumpSuit)
}

// smallestSameSuitWinner リードスートに従って勝てる最小ランクのカード (非トランプ) を返す
func smallestSameSuitWinner(player *BriscolaPlayer, leadCard *Card, leadSuit int) int {
	bestIdx := -1
	bestRank := -1
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetDesign() != leadSuit {
			continue
		}
		if BriscolaRankOrder(c) <= BriscolaRankOrder(leadCard) {
			continue
		}
		r := BriscolaRankOrder(c)
		if bestIdx < 0 || r < bestRank {
			bestIdx = i
			bestRank = r
		}
	}
	return bestIdx
}

// smallestWinningTrump 勝てる最小ランクのトランプを返す。
// リード自体がトランプの場合はそれより強いトランプを探す。
func smallestWinningTrump(player *BriscolaPlayer, leadCard *Card, leadSuit, trumpSuit int) int {
	bestIdx := -1
	bestRank := -1
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if c.GetDesign() != trumpSuit {
			continue
		}
		// リードがトランプならランクで上回る必要がある
		if leadSuit == trumpSuit && BriscolaRankOrder(c) <= BriscolaRankOrder(leadCard) {
			continue
		}
		r := BriscolaRankOrder(c)
		if bestIdx < 0 || r < bestRank {
			bestIdx = i
			bestRank = r
		}
	}
	return bestIdx
}

// smallestDump 取られても痛くないカードを 1 枚捨てる。
// 優先順: 非トランプの 0 点札 → 非トランプの低点数札 → 低ランクのトランプ。
func smallestDump(player *BriscolaPlayer, trumpSuit int) int {
	bestIdx := 0
	bestScore := dumpScore(player.GetCard(0), trumpSuit)
	for i := 1; i < player.GetCardsSize(); i++ {
		s := dumpScore(player.GetCard(i), trumpSuit)
		if s < bestScore {
			bestScore = s
			bestIdx = i
		}
	}
	return bestIdx
}

// dumpScore 値が小さいほど「失っても良い」カード
func dumpScore(c *Card, trumpSuit int) int {
	score := BriscolaCardPoints(c)*10 + BriscolaRankOrder(c)
	if c.GetDesign() == trumpSuit {
		score += 1000
	}
	return score
}

// --- JSON ---

// briscolaJSON is the JSON wire format for Briscola.
type briscolaJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*BriscolaPlayer `json:"ps"`
	Config           BriscolaConfig    `json:"cf"`
	Phase            BriscolaPhase     `json:"ph"`
	TrickNumber      int               `json:"tn"`
	CurrentPlayerIdx int               `json:"ci"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	TrumpCard        *Card             `json:"tu"`
	TrumpSuit        int               `json:"ts"`
	LeadPlayerIdx    int               `json:"li"`
	DealerIdx        int               `json:"di"`
	PlayerPoints     []int             `json:"pp"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (b *Briscola) MarshalJSON() ([]byte, error) {
	return json.Marshal(briscolaJSON{
		TrumpCards:       b.trumpCards,
		Players:          b.players,
		Config:           b.config,
		Phase:            b.phase,
		TrickNumber:      b.trickNumber,
		CurrentPlayerIdx: b.currentPlayerIdx,
		CurrentTrick:     b.currentTrick,
		TrumpCard:        b.trumpCard,
		TrumpSuit:        b.trumpSuit,
		LeadPlayerIdx:    b.leadPlayerIdx,
		DealerIdx:        b.dealerIdx,
		PlayerPoints:     b.playerPoints,
		GameEndFlag:      b.gameEndFlag,
		WinnerIdx:        b.winnerIdx,
		ActionLog:        b.actionLog,
	})
}

// briscolaMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const briscolaMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
//
// Validates that the deserialised game state matches Briscola's fixed shape
// (BriscolaPlayerCnt = 2 players, at most BriscolaPlayerCnt cards on the current
// trick, PlayerPoints aligned to the player count) and that the variable-length
// ActionLog does not exceed briscolaMaxSliceLen, preventing DoS via crafted
// payloads and out-of-bounds access during play.
func (b *Briscola) UnmarshalJSON(data []byte) error {
	var j briscolaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != BriscolaPlayerCnt {
		return fmt.Errorf("briscola: expected %d players, got %d", BriscolaPlayerCnt, len(j.Players))
	}
	if len(j.CurrentTrick) > BriscolaPlayerCnt {
		return fmt.Errorf("briscola: current trick has %d cards (max %d)", len(j.CurrentTrick), BriscolaPlayerCnt)
	}
	if j.PlayerPoints != nil && len(j.PlayerPoints) != BriscolaPlayerCnt {
		return fmt.Errorf("briscola: expected %d player points entries, got %d", BriscolaPlayerCnt, len(j.PlayerPoints))
	}
	if len(j.ActionLog) > briscolaMaxSliceLen {
		return fmt.Errorf("briscola: action log exceeds maximum allowed size")
	}
	b.trumpCards = j.TrumpCards
	if b.trumpCards == nil {
		b.trumpCards = NewTrumpCardsBriscola()
	}
	b.players = j.Players
	b.config = j.Config
	b.phase = j.Phase
	b.trickNumber = j.TrickNumber
	b.currentPlayerIdx = j.CurrentPlayerIdx
	b.currentTrick = j.CurrentTrick
	if b.currentTrick == nil {
		b.currentTrick = make([]*TrickCard, 0)
	}
	b.trumpCard = j.TrumpCard
	b.trumpSuit = j.TrumpSuit
	b.leadPlayerIdx = j.LeadPlayerIdx
	b.dealerIdx = j.DealerIdx
	b.playerPoints = j.PlayerPoints
	if b.playerPoints == nil {
		b.playerPoints = make([]int, BriscolaPlayerCnt)
	}
	b.gameEndFlag = j.GameEndFlag
	b.winnerIdx = j.WinnerIdx
	b.actionLog = j.ActionLog
	if b.actionLog == nil {
		b.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
