//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// CribbagePlayerCnt クリベッジプレイヤー数
const CribbagePlayerCnt = 2

// CribbageDealSize 初期配布枚数
const CribbageDealSize = 6

// CribbageHandSize ディスカード後の手札枚数
const CribbageHandSize = 4

// CribbageDiscardSize クリブに捨てる枚数
const CribbageDiscardSize = 2

// CribbagePegLimit ペギングの上限値
const CribbagePegLimit = 31

// CribbageDefaultPointLimit デフォルトのゲーム終了スコア
const CribbageDefaultPointLimit = 121

// CribbagePhase ゲームフェーズ
type CribbagePhase int

// Cribbageのフェーズ定数
const (
	// CribbagePhaseDiscard 捨て札フェーズ (各プレイヤーが2枚をクリブに捨てる)
	CribbagePhaseDiscard CribbagePhase = 0
	// CribbagePhaseCut カットフェーズ (スターターカードを公開)
	CribbagePhaseCut CribbagePhase = 1
	// CribbagePhasePegging ペギングフェーズ (交互にカードを出して31を目指す)
	CribbagePhasePegging CribbagePhase = 2
	// CribbagePhaseShow ショーフェーズ (ハンド・クリブのスコアを計算)
	CribbagePhaseShow CribbagePhase = 3
	// CribbagePhaseRoundEnd ラウンド終了フェーズ
	CribbagePhaseRoundEnd CribbagePhase = 4
	// CribbagePhaseGameEnd ゲーム終了フェーズ
	CribbagePhaseGameEnd CribbagePhase = 5
)

// Cribbage クリベッジゲームクラス
type Cribbage struct {
	trumpCards       *TrumpCards
	players          []*CribbagePlayer
	config           CribbageConfig
	phase            CribbagePhase
	currentPlayerIdx int
	dealerIdx        int // ディーラーのインデックス (交互に入れ替わる)
	crib             []*Card
	starter          *Card
	drawPile         []*Card
	// ペギング状態
	pegCount       int     // 現在の合計 (0-31)
	pegPlayedCards []*Card // 現在のペギングシーケンスで出されたカード
	pegGoState     int     // 0=通常, 1=一方がGoを宣言, 2=両方がGoを宣言
	lastPegPlayer  int     // 最後にカードを出したプレイヤー
	// ペギング中の各プレイヤーの出したカード (手札から除外済み)
	playerPeggedCards [CribbagePlayerCnt][]*Card
	// ショーフェーズ状態
	showPhaseStep    int // 0=非ディーラー手札, 1=ディーラー手札, 2=クリブ
	handScoreDetails [3]*CribbageScoreDetail
	// ゲーム全体の状態
	gameEndFlag bool
	winnerIdx   int
	roundNumber int
	actionLog   []*ActionLogEntry
	// ディスカード状態追跡
	discardDone [CribbagePlayerCnt]bool
	// 各プレイヤーの元の手札 (ショーフェーズ用に保持)
	originalHands [CribbagePlayerCnt][]*Card
}

// NewCribbage コンストラクタ
func NewCribbage(trumpCards *TrumpCards, players []*CribbagePlayer, config CribbageConfig) *Cribbage {
	return &Cribbage{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		winnerIdx:  -1,
		dealerIdx:  1, // CPU starts as dealer, human goes first
	}
}

// NewDefaultCribbage returns Cribbage with the standard 2-player setup (1 human, 1 CPU)
// and DefaultCribbageConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultCribbage() *Cribbage {
	players := []*CribbagePlayer{
		NewCribbagePlayer(true),
		NewCribbagePlayer(false),
	}
	return NewCribbage(NewTrumpCards(0), players, DefaultCribbageConfig())
}

// Reset ゲーム初期化
func (g *Cribbage) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.dealerIdx = 1 // CPU starts as dealer
	g.actionLog = nil

	for _, p := range g.players {
		p.SetCumulativeScore(0)
		p.SetRoundScore(0)
		p.Reset()
		p.SetIsFinished(false)
	}

	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Cribbage) NextRound() {
	if g.phase != CribbagePhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = 1 - g.dealerIdx // ディーラーを交代
	g.startRound()
}

// startRound ラウンドの初期化と配布
func (g *Cribbage) startRound() {
	g.crib = nil
	g.starter = nil
	g.pegCount = 0
	g.pegPlayedCards = nil
	g.pegGoState = 0
	g.lastPegPlayer = -1
	g.playerPeggedCards = [CribbagePlayerCnt][]*Card{}
	g.showPhaseStep = 0
	g.handScoreDetails = [3]*CribbageScoreDetail{}
	g.discardDone = [CribbagePlayerCnt]bool{}
	g.originalHands = [CribbagePlayerCnt][]*Card{}

	for _, p := range g.players {
		p.ResetRound()
	}

	// デッキを準備してシャッフル
	g.drawPile = nil
	g.trumpCards.Shuffle()
	for {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.drawPile = append(g.drawPile, card)
	}
	rand.Shuffle(len(g.drawPile), func(i, j int) {
		g.drawPile[i], g.drawPile[j] = g.drawPile[j], g.drawPile[i]
	})

	// 各プレイヤーに6枚配布
	for range CribbageDealSize {
		for j := range CribbagePlayerCnt {
			if len(g.drawPile) > 0 {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				g.players[j].AddCard(card)
			}
		}
	}

	// 非ディーラーから先にディスカード
	g.currentPlayerIdx = 1 - g.dealerIdx
	g.phase = CribbagePhaseDiscard
}

// PlayerDiscard 人間プレイヤーがクリブに2枚捨てる
func (g *Cribbage) PlayerDiscard(indices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CribbagePhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.doDiscard(g.currentPlayerIdx, indices)
}

// doDiscard 指定プレイヤーのディスカード処理
func (g *Cribbage) doDiscard(playerIdx int, indices []int) error {
	if len(indices) != CribbageDiscardSize {
		return NewDomainError(ErrInvalidIndices, fmt.Sprintf("%d枚選択してください", CribbageDiscardSize))
	}

	p := g.players[playerIdx]
	handSize := p.GetCardsSize()

	// インデックスの検証
	for _, idx := range indices {
		if idx < 0 || idx >= handSize {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
	}
	if indices[0] == indices[1] {
		return NewDomainError(ErrInvalidIndices, "同じカードを2枚選択できません")
	}

	// RemoveCards で安全に削除 (内部で降順削除される)
	removed := p.RemoveCards(indices)
	g.crib = append(g.crib, removed...)

	g.discardDone[playerIdx] = true

	g.addLog(playerIdx, "discard", "クリブに2枚捨てた", removed)

	// 両方がディスカード完了したか確認
	if g.discardDone[0] && g.discardDone[1] {
		// 元の手札を保存 (ショーフェーズ用)
		for i := range CribbagePlayerCnt {
			g.originalHands[i] = g.getPlayerCards(i)
		}
		g.enterCutPhase()
	} else {
		// もう一方のプレイヤーに切り替え
		g.currentPlayerIdx = 1 - playerIdx
	}
	return nil
}

// enterCutPhase カットフェーズに入る。スターターはまだ公開せず、非ディーラー
// (切り手) の明示的なカット操作を待つ。切り手が人間なら PlayerCut を、CPU なら
// CpuPlay 経由で doCut が呼ばれてスターターが公開される。
func (g *Cribbage) enterCutPhase() {
	g.phase = CribbagePhaseCut
	g.currentPlayerIdx = 1 - g.dealerIdx // 非ディーラーがデッキをカットする
}

// PlayerCut 人間の非ディーラーがデッキをカットしてスターターを公開する
func (g *Cribbage) PlayerCut() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CribbagePhaseCut {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.doCut()
	return nil
}

// doCut スターターカードを公開し、His Heels を適用してペギングへ移行する
func (g *Cribbage) doCut() {
	if len(g.drawPile) > 0 {
		g.starter = g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
	}

	g.addLog(-1, "cut", "スターターカード公開", []*Card{g.starter})

	// His Heels: スターターがJなら、ディーラーに2点
	if g.starter != nil && g.starter.GetValue() == CribbageJackValue {
		g.addScore(g.dealerIdx, 2, "His Heels (スターターがJ)")
		if g.checkWin() {
			return
		}
	}

	// ペギングフェーズへ移行
	g.phase = CribbagePhasePegging
	g.currentPlayerIdx = 1 - g.dealerIdx // 非ディーラーが先攻
	g.pegCount = 0
	g.pegPlayedCards = nil
	g.pegGoState = 0
	g.lastPegPlayer = -1
}

// PlayerPeg 人間プレイヤーがペギングでカードを出す
func (g *Cribbage) PlayerPeg(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CribbagePhasePegging {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.doPeg(g.currentPlayerIdx, cardIndex)
}

// doPeg 指定プレイヤーのペギング処理
func (g *Cribbage) doPeg(playerIdx int, cardIndex int) error {
	p := g.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := p.GetCard(cardIndex)
	cardVal := cribbageCardValue(card)

	if g.pegCount+cardVal > CribbagePegLimit {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("合計が%dを超えます", CribbagePegLimit))
	}

	// カードを出す
	p.RemoveCard(cardIndex)
	g.pegCount += cardVal
	g.pegPlayedCards = append(g.pegPlayedCards, card)
	g.playerPeggedCards[playerIdx] = append(g.playerPeggedCards[playerIdx], card)
	g.lastPegPlayer = playerIdx
	g.pegGoState = 0

	g.addLog(playerIdx, "peg", fmt.Sprintf("カードを出した (合計: %d)", g.pegCount), []*Card{card})

	// ペギングスコア
	pegScore := CribbageScorePegging(g.pegPlayedCards, g.pegCount)
	if pegScore > 0 {
		g.addScore(playerIdx, pegScore, fmt.Sprintf("ペギング %d点", pegScore))
		if g.checkWin() {
			return nil
		}
	}

	// 31に到達したらリセット
	if g.pegCount == CribbagePegLimit {
		g.resetPegSequence()
	}

	g.advancePegging()
	return nil
}

// PlayerGo 人間プレイヤーがGoを宣言
func (g *Cribbage) PlayerGo() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CribbagePhasePegging {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	// プレイ可能なカードがない場合のみGoを許可
	if g.canPeg(g.currentPlayerIdx) {
		return NewDomainError(ErrInvalidPlay, "まだ出せるカードがあります")
	}
	return g.doGo(g.currentPlayerIdx)
}

// doGo Goの処理
func (g *Cribbage) doGo(playerIdx int) error {
	g.addLog(playerIdx, "go", "Go", nil)
	g.pegGoState++

	if g.pegGoState >= 2 || !g.canAnyPlayerPeg() {
		// 両方Go → ラストカードの1点を最後にカードを出したプレイヤーに付与
		if g.lastPegPlayer >= 0 && g.pegCount < CribbagePegLimit {
			g.addScore(g.lastPegPlayer, 1, "ラストカード (Go)")
			if g.checkWin() {
				return nil
			}
		}
		g.resetPegSequence()
		g.advancePegging()
	} else {
		// 相手に順番を渡す
		g.currentPlayerIdx = 1 - playerIdx
	}
	return nil
}

// canPeg プレイヤーがペギングで出せるカードがあるか
func (g *Cribbage) canPeg(playerIdx int) bool {
	p := g.players[playerIdx]
	for i := range p.GetCardsSize() {
		if g.pegCount+cribbageCardValue(p.GetCard(i)) <= CribbagePegLimit {
			return true
		}
	}
	return false
}

// canAnyPlayerPeg どちらかのプレイヤーがペギング可能か
func (g *Cribbage) canAnyPlayerPeg() bool {
	for i := range CribbagePlayerCnt {
		if g.canPeg(i) {
			return true
		}
	}
	return false
}

// resetPegSequence ペギングシーケンスをリセット
func (g *Cribbage) resetPegSequence() {
	g.pegCount = 0
	g.pegPlayedCards = nil
	g.pegGoState = 0
	g.lastPegPlayer = -1
}

// advancePegging ペギングの次のステップを決定
func (g *Cribbage) advancePegging() {
	// 両プレイヤーの手札が空ならショーフェーズへ
	if g.players[0].GetCardsSize() == 0 && g.players[1].GetCardsSize() == 0 {
		// 最後のカードを出したプレイヤーに1点 (31でなかった場合)
		if g.lastPegPlayer >= 0 && g.pegCount > 0 && g.pegCount < CribbagePegLimit {
			g.addScore(g.lastPegPlayer, 1, "ラストカード")
			if g.checkWin() {
				return
			}
		}
		g.startShowPhase()
		return
	}

	// 次のプレイヤーへ
	nextPlayer := 1 - g.currentPlayerIdx
	if g.canPeg(nextPlayer) {
		g.currentPlayerIdx = nextPlayer
	} else if g.canPeg(g.currentPlayerIdx) {
		// 相手が出せないなら自分が続ける
	} else {
		// どちらも出せない
		if g.lastPegPlayer >= 0 && g.pegCount > 0 && g.pegCount < CribbagePegLimit {
			g.addScore(g.lastPegPlayer, 1, "ラストカード (Go)")
			if g.checkWin() {
				return
			}
		}
		g.resetPegSequence()
		// リセット後に出せるか再チェック
		if !g.canAnyPlayerPeg() {
			g.startShowPhase()
			return
		}
		// 非ディーラーから
		if g.canPeg(1 - g.dealerIdx) {
			g.currentPlayerIdx = 1 - g.dealerIdx
		} else {
			g.currentPlayerIdx = g.dealerIdx
		}
	}
}

// startShowPhase ショーフェーズを開始
func (g *Cribbage) startShowPhase() {
	g.phase = CribbagePhaseShow
	g.showPhaseStep = 0
}

// ShowNext ショーフェーズの次のスコア計算を実行
func (g *Cribbage) ShowNext() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CribbagePhaseShow {
		return ErrWrongPhase
	}

	switch g.showPhaseStep {
	case 0:
		// 非ディーラーの手札をスコア
		nonDealerIdx := 1 - g.dealerIdx
		detail := CribbageScoreHand(g.originalHands[nonDealerIdx], g.starter, false)
		g.handScoreDetails[0] = &detail
		if detail.Total > 0 {
			g.addScore(nonDealerIdx, detail.Total, fmt.Sprintf("ハンドスコア %d点", detail.Total))
			if g.checkWin() {
				return nil
			}
		}
		g.addLog(nonDealerIdx, "show", fmt.Sprintf("ハンドスコア: %d点", detail.Total), g.originalHands[nonDealerIdx])
	case 1:
		// ディーラーの手札をスコア
		detail := CribbageScoreHand(g.originalHands[g.dealerIdx], g.starter, false)
		g.handScoreDetails[1] = &detail
		if detail.Total > 0 {
			g.addScore(g.dealerIdx, detail.Total, fmt.Sprintf("ハンドスコア %d点", detail.Total))
			if g.checkWin() {
				return nil
			}
		}
		g.addLog(g.dealerIdx, "show", fmt.Sprintf("ハンドスコア: %d点", detail.Total), g.originalHands[g.dealerIdx])
	case 2:
		// クリブをスコア (ディーラーのもの)
		detail := CribbageScoreHand(g.crib, g.starter, true)
		g.handScoreDetails[2] = &detail
		if detail.Total > 0 {
			g.addScore(g.dealerIdx, detail.Total, fmt.Sprintf("クリブスコア %d点", detail.Total))
			if g.checkWin() {
				return nil
			}
		}
		g.addLog(g.dealerIdx, "show_crib", fmt.Sprintf("クリブスコア: %d点", detail.Total), g.crib)
		// ショーフェーズ完了
		g.phase = CribbagePhaseRoundEnd
	}

	g.showPhaseStep++
	return nil
}

// CpuPlay CPUのターンを自動実行
func (g *Cribbage) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	playerIdx := g.currentPlayerIdx
	if g.players[playerIdx].GetIsHuman() {
		return
	}

	switch g.phase {
	case CribbagePhaseDiscard:
		g.cpuDiscard(playerIdx)
	case CribbagePhaseCut:
		// CPU が切り手 (非ディーラー) の場合は従来通り自動でカットする
		g.doCut()
	case CribbagePhasePegging:
		g.cpuPeg(playerIdx)
	}
}

// cpuDiscard CPUのディスカード処理
func (g *Cribbage) cpuDiscard(playerIdx int) {
	hand := g.getPlayerCards(playerIdx)
	if len(hand) < CribbageDealSize {
		return
	}

	var i1, i2 int

	switch g.config.CpuDifficulty {
	case CribbageCpuDifficultyHard:
		i1, i2 = g.cpuDiscardHard(playerIdx)
	case CribbageCpuDifficultyNormal:
		i1, i2 = g.cpuDiscardNormal(playerIdx)
	default:
		i1, i2 = g.cpuDiscardEasy(hand)
	}

	_ = g.doDiscard(playerIdx, []int{i1, i2})
}

// cpuDiscardEasy ランダムに2枚捨てる
func (g *Cribbage) cpuDiscardEasy(hand []*Card) (int, int) {
	i1 := rand.Intn(len(hand))
	i2 := rand.Intn(len(hand) - 1)
	if i2 >= i1 {
		i2++
	}
	return i1, i2
}

// cpuDiscardNormal ヒューリスティックで2枚捨てる
func (g *Cribbage) cpuDiscardNormal(playerIdx int) (int, int) {
	hand := g.getPlayerCards(playerIdx)
	bestScore := -1
	bestI, bestJ := 0, 1

	// C(6,2)=15通りを評価
	for i := range len(hand) {
		for j := i + 1; j < len(hand); j++ {
			// i,jを除いた4枚の手札スコアを評価
			remaining := make([]*Card, 0, 4)
			for k, c := range hand {
				if k != i && k != j {
					remaining = append(remaining, c)
				}
			}
			score := CribbageScoreHand(remaining, nil, false).Total
			if score > bestScore {
				bestScore = score
				bestI, bestJ = i, j
			}
		}
	}
	return bestI, bestJ
}

// cpuDiscardHard 最適な2枚を捨てる (期待値を考慮)
func (g *Cribbage) cpuDiscardHard(playerIdx int) (int, int) {
	hand := g.getPlayerCards(playerIdx)
	isDealer := playerIdx == g.dealerIdx
	bestScore := -1000
	bestI, bestJ := 0, 1

	for i := range len(hand) {
		for j := i + 1; j < len(hand); j++ {
			remaining := make([]*Card, 0, 4)
			discarded := make([]*Card, 0, 2)
			for k, c := range hand {
				if k != i && k != j {
					remaining = append(remaining, c)
				} else {
					discarded = append(discarded, c)
				}
			}
			handScore := CribbageScoreHand(remaining, nil, false).Total
			// クリブへのディスカードの価値を簡易評価
			cribValue := g.estimateCribValue(discarded)
			score := handScore
			if isDealer {
				score += cribValue // ディーラーはクリブも自分のもの
			} else {
				score -= cribValue // 非ディーラーはクリブを相手に渡す
			}
			if score > bestScore {
				bestScore = score
				bestI, bestJ = i, j
			}
		}
	}
	return bestI, bestJ
}

// estimateCribValue クリブに送る2枚の期待値を簡易評価
func (g *Cribbage) estimateCribValue(cards []*Card) int {
	if len(cards) < 2 {
		return 0
	}
	score := 0
	// 15になる組み合わせ
	if cribbageCardValue(cards[0])+cribbageCardValue(cards[1]) == 15 {
		score += 2
	}
	// ペア
	if cards[0].GetValue() == cards[1].GetValue() {
		score += 2
	}
	// 5を含む (15を作りやすい)
	for _, c := range cards {
		if cribbageCardValue(c) == 5 {
			score += 1
		}
	}
	return score
}

// cpuPeg CPUのペギング処理
func (g *Cribbage) cpuPeg(playerIdx int) {
	if !g.canPeg(playerIdx) {
		_ = g.doGo(playerIdx)
		return
	}

	hand := g.getPlayerCards(playerIdx)
	switch g.config.CpuDifficulty {
	case CribbageCpuDifficultyHard:
		g.cpuPegHard(playerIdx, hand)
	case CribbageCpuDifficultyNormal:
		g.cpuPegNormal(playerIdx, hand)
	default:
		g.cpuPegEasy(playerIdx, hand)
	}
}

// cpuPegEasy ランダムに出せるカードを出す
func (g *Cribbage) cpuPegEasy(playerIdx int, hand []*Card) {
	playable := make([]int, 0)
	for i, c := range hand {
		if g.pegCount+cribbageCardValue(c) <= CribbagePegLimit {
			playable = append(playable, i)
		}
	}
	if len(playable) > 0 {
		_ = g.doPeg(playerIdx, playable[rand.Intn(len(playable))])
	}
}

// cpuPegNormal ヒューリスティックでカードを出す
func (g *Cribbage) cpuPegNormal(playerIdx int, hand []*Card) {
	bestIdx := -1
	bestScore := -100

	for i, c := range hand {
		cv := cribbageCardValue(c)
		if g.pegCount+cv > CribbagePegLimit {
			continue
		}
		newCount := g.pegCount + cv
		played := append(append([]*Card{}, g.pegPlayedCards...), c)
		score := CribbageScorePegging(played, newCount)

		// 5や21に合計を残さない (相手が15/31を作りやすい)
		if newCount == 5 || newCount == 21 {
			score -= 2
		}
		if score > bestScore || bestIdx == -1 {
			bestScore = score
			bestIdx = i
		}
	}
	if bestIdx >= 0 {
		_ = g.doPeg(playerIdx, bestIdx)
	}
}

// cpuPegHard 最適なカードを出す
func (g *Cribbage) cpuPegHard(playerIdx int, hand []*Card) {
	bestIdx := -1
	bestScore := -100

	for i, c := range hand {
		cv := cribbageCardValue(c)
		if g.pegCount+cv > CribbagePegLimit {
			continue
		}
		newCount := g.pegCount + cv
		played := append(append([]*Card{}, g.pegPlayedCards...), c)
		score := CribbageScorePegging(played, newCount)

		// 31 or 15 に到達するのは良い
		if newCount == CribbagePegLimit {
			score += 3
		}
		// 相手が15/31を作りやすい数値を避ける
		if newCount == 5 || newCount == 21 {
			score -= 3
		}
		if score > bestScore || bestIdx == -1 {
			bestScore = score
			bestIdx = i
		}
	}
	if bestIdx >= 0 {
		_ = g.doPeg(playerIdx, bestIdx)
	}
}

// addScore プレイヤーにスコアを加算
func (g *Cribbage) addScore(playerIdx int, points int, reason string) {
	p := g.players[playerIdx]
	p.SetRoundScore(p.GetRoundScore() + points)
	p.SetCumulativeScore(p.GetCumulativeScore() + points)
	if reason != "" {
		g.addLog(playerIdx, "score", reason, nil)
	}
}

// checkWin 勝利条件を確認
func (g *Cribbage) checkWin() bool {
	for i, p := range g.players {
		if p.GetCumulativeScore() >= g.config.PointLimit {
			g.gameEndFlag = true
			g.winnerIdx = i
			g.phase = CribbagePhaseGameEnd
			return true
		}
	}
	return false
}

// getPlayerCards プレイヤーの手札をスライスとして返すヘルパー
func (g *Cribbage) getPlayerCards(playerIdx int) []*Card {
	p := g.players[playerIdx]
	cards := make([]*Card, p.GetCardsSize())
	for i := range cards {
		cards[i] = p.GetCard(i)
	}
	return cards
}

// addLog アクションログを追加
func (g *Cribbage) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.roundNumber,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// ---- Getter/Setter ----

// GetPhase フェーズ取得
func (g *Cribbage) GetPhase() CribbagePhase { return g.phase }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Cribbage) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// IsHumanTurn 人間のターンかどうか
func (g *Cribbage) IsHumanTurn() bool {
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetDealerIdx ディーラーインデックス取得
func (g *Cribbage) GetDealerIdx() int { return g.dealerIdx }

// GetCrib クリブ取得
func (g *Cribbage) GetCrib() []*Card { return g.crib }

// GetStarter スターターカード取得
func (g *Cribbage) GetStarter() *Card { return g.starter }

// GetPegCount ペギングカウント取得
func (g *Cribbage) GetPegCount() int { return g.pegCount }

// GetPegPlayedCards ペギングで出されたカード取得
func (g *Cribbage) GetPegPlayedCards() []*Card { return g.pegPlayedCards }

// GetShowPhaseStep ショーフェーズのステップ取得
func (g *Cribbage) GetShowPhaseStep() int { return g.showPhaseStep }

// GetHandScoreDetails ハンドスコア詳細取得
func (g *Cribbage) GetHandScoreDetails() [3]*CribbageScoreDetail { return g.handScoreDetails }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Cribbage) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得
func (g *Cribbage) GetWinnerIdx() int { return g.winnerIdx }

// GetRoundNumber ラウンド番号取得
func (g *Cribbage) GetRoundNumber() int { return g.roundNumber }

// GetActionLog アクションログ取得
func (g *Cribbage) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetPlayer プレイヤー取得
func (g *Cribbage) GetPlayer(idx int) *CribbagePlayer {
	if idx < 0 || idx >= len(g.players) {
		return nil
	}
	return g.players[idx]
}

// GetConfig 設定取得
func (g *Cribbage) GetConfig() CribbageConfig { return g.config }

// SetConfig 設定変更
func (g *Cribbage) SetConfig(config CribbageConfig) { g.config = config }

// GetOriginalHand ショーフェーズ用の元の手札を取得
func (g *Cribbage) GetOriginalHand(playerIdx int) []*Card {
	if playerIdx < 0 || playerIdx >= CribbagePlayerCnt {
		return nil
	}
	return g.originalHands[playerIdx]
}

// GetPlayerPeggedCards プレイヤーがペギングで出したカードを取得
func (g *Cribbage) GetPlayerPeggedCards(playerIdx int) []*Card {
	if playerIdx < 0 || playerIdx >= CribbagePlayerCnt {
		return nil
	}
	return g.playerPeggedCards[playerIdx]
}

// ---- テスト用セッター ----

// SetPhase テスト用: フェーズを設定
func (g *Cribbage) SetPhase(phase CribbagePhase) { g.phase = phase }

// SetCurrentPlayerIdx テスト用: 現在のプレイヤーインデックスを設定
func (g *Cribbage) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// SetDealerIdx テスト用: ディーラーインデックスを設定
func (g *Cribbage) SetDealerIdx(idx int) { g.dealerIdx = idx }

// SetPegCount テスト用: ペギングカウントを設定
func (g *Cribbage) SetPegCount(count int) { g.pegCount = count }

// SetPegPlayedCards テスト用: ペギングで出されたカードを設定
func (g *Cribbage) SetPegPlayedCards(cards []*Card) { g.pegPlayedCards = cards }

// SetStarter テスト用: スターターカードを設定
func (g *Cribbage) SetStarter(card *Card) { g.starter = card }

// SetCrib テスト用: クリブを設定
func (g *Cribbage) SetCrib(cards []*Card) { g.crib = cards }

// SetGameEndFlag テスト用: ゲーム終了フラグを設定
func (g *Cribbage) SetGameEndFlag(flag bool) { g.gameEndFlag = flag }

// SetWinnerIdx テスト用: 勝者インデックスを設定
func (g *Cribbage) SetWinnerIdx(idx int) { g.winnerIdx = idx }

// SetOriginalHand テスト用: 元の手札を設定
func (g *Cribbage) SetOriginalHand(playerIdx int, cards []*Card) {
	if playerIdx >= 0 && playerIdx < CribbagePlayerCnt {
		g.originalHands[playerIdx] = cards
	}
}

// SetDiscardDone テスト用: ディスカード完了フラグを設定
func (g *Cribbage) SetDiscardDone(playerIdx int, done bool) {
	if playerIdx >= 0 && playerIdx < CribbagePlayerCnt {
		g.discardDone[playerIdx] = done
	}
}

// --- JSON Serialization ---

// cribbageJSON is the JSON wire format for Cribbage.
type cribbageJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*CribbagePlayer    `json:"pl"`
	Config           CribbageConfig       `json:"cf"`
	Phase            CribbagePhase        `json:"ph"`
	CurrentPlayerIdx int                  `json:"ci"`
	DealerIdx        int                  `json:"di"`
	Crib             []*Card              `json:"cb"`
	Starter          *Card                `json:"st"`
	DrawPile         []*Card              `json:"dp"`
	PegCount         int                  `json:"pc"`
	PegPlayedCards   []*Card              `json:"pp"`
	PegGoState       int                  `json:"pg"`
	LastPegPlayer    int                  `json:"lp"`
	PlayerPegCards0  []*Card              `json:"p0"`
	PlayerPegCards1  []*Card              `json:"p1"`
	ShowPhaseStep    int                  `json:"ss"`
	HandScoreDetail0 *CribbageScoreDetail `json:"h0,omitempty"`
	HandScoreDetail1 *CribbageScoreDetail `json:"h1,omitempty"`
	HandScoreDetail2 *CribbageScoreDetail `json:"h2,omitempty"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerIdx        int                  `json:"wi"`
	RoundNumber      int                  `json:"rn"`
	ActionLog        []*ActionLogEntry    `json:"al"`
	DiscardDone0     bool                 `json:"d0"`
	DiscardDone1     bool                 `json:"d1"`
	OriginalHand0    []*Card              `json:"o0"`
	OriginalHand1    []*Card              `json:"o1"`
}

// cribbageMaxSliceLen caps slice sizes during deserialisation.
const cribbageMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (g *Cribbage) MarshalJSON() ([]byte, error) {
	return json.Marshal(cribbageJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DealerIdx:        g.dealerIdx,
		Crib:             g.crib,
		Starter:          g.starter,
		DrawPile:         g.drawPile,
		PegCount:         g.pegCount,
		PegPlayedCards:   g.pegPlayedCards,
		PegGoState:       g.pegGoState,
		LastPegPlayer:    g.lastPegPlayer,
		PlayerPegCards0:  g.playerPeggedCards[0],
		PlayerPegCards1:  g.playerPeggedCards[1],
		ShowPhaseStep:    g.showPhaseStep,
		HandScoreDetail0: g.handScoreDetails[0],
		HandScoreDetail1: g.handScoreDetails[1],
		HandScoreDetail2: g.handScoreDetails[2],
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		ActionLog:        g.actionLog,
		DiscardDone0:     g.discardDone[0],
		DiscardDone1:     g.discardDone[1],
		OriginalHand0:    g.originalHands[0],
		OriginalHand1:    g.originalHands[1],
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Cribbage) UnmarshalJSON(data []byte) error {
	var j cribbageJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > cribbageMaxSliceLen || len(j.Crib) > cribbageMaxSliceLen ||
		len(j.DrawPile) > cribbageMaxSliceLen || len(j.PegPlayedCards) > cribbageMaxSliceLen ||
		len(j.ActionLog) > cribbageMaxSliceLen {
		return fmt.Errorf("cribbage: input array exceeds maximum allowed size")
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*CribbagePlayer, 0)
	}
	g.config = j.Config
	g.phase = j.Phase
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.crib = j.Crib
	if g.crib == nil {
		g.crib = make([]*Card, 0)
	}
	g.starter = j.Starter
	g.drawPile = j.DrawPile
	if g.drawPile == nil {
		g.drawPile = make([]*Card, 0)
	}
	g.pegCount = j.PegCount
	g.pegPlayedCards = j.PegPlayedCards
	if g.pegPlayedCards == nil {
		g.pegPlayedCards = make([]*Card, 0)
	}
	g.pegGoState = j.PegGoState
	g.lastPegPlayer = j.LastPegPlayer
	g.playerPeggedCards = [CribbagePlayerCnt][]*Card{j.PlayerPegCards0, j.PlayerPegCards1}
	g.showPhaseStep = j.ShowPhaseStep
	g.handScoreDetails = [3]*CribbageScoreDetail{j.HandScoreDetail0, j.HandScoreDetail1, j.HandScoreDetail2}
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.roundNumber = j.RoundNumber
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	g.discardDone = [CribbagePlayerCnt]bool{j.DiscardDone0, j.DiscardDone1}
	g.originalHands = [CribbagePlayerCnt][]*Card{j.OriginalHand0, j.OriginalHand1}
	return nil
}

// CribbageHint 人間プレイヤー向けの推奨アクション
type CribbageHint struct {
	Type    string // "discard"（クリブへ捨てる2枚） or "play"（ペギングで出す1枚）
	Indices []int  // 手札インデックス
}

// GetHint 人間プレイヤーの現フェーズにおける推奨アクションを返す（対象外フェーズや手番外は nil）
func (g *Cribbage) GetHint() *CribbageHint {
	if g.gameEndFlag || !g.IsHumanTurn() {
		return nil
	}
	switch g.phase {
	case CribbagePhaseDiscard:
		return g.discardHint()
	case CribbagePhasePegging:
		return g.peggingHint()
	default:
		return nil
	}
}

// discardHint 残す4枚の素点（スターターなし）が最大になる捨て札2枚を返す
func (g *Cribbage) discardHint() *CribbageHint {
	p := g.players[g.currentPlayerIdx]
	handSize := p.GetCardsSize()
	if handSize != CribbageHandSize+CribbageDiscardSize {
		return nil
	}
	bestScore := -1
	var bestI, bestJ int
	keep := make([]*Card, 0, CribbageHandSize)
	for i := 0; i < handSize; i++ {
		for j := i + 1; j < handSize; j++ {
			keep = keep[:0]
			for k := 0; k < handSize; k++ {
				if k != i && k != j {
					keep = append(keep, p.GetCard(k))
				}
			}
			score := CribbageScoreHand(keep, nil, false).Total
			if score > bestScore {
				bestScore = score
				bestI, bestJ = i, j
			}
		}
	}
	return &CribbageHint{Type: "discard", Indices: []int{bestI, bestJ}}
}

// peggingHint 即時ペギング得点が最大になる合法カード1枚を返す（合法手なし＝Goのみなら nil）
func (g *Cribbage) peggingHint() *CribbageHint {
	p := g.players[g.currentPlayerIdx]
	bestScore := -1
	bestIdx := -1
	seq := make([]*Card, len(g.pegPlayedCards)+1)
	copy(seq, g.pegPlayedCards)
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		v := cribbageCardValue(c)
		if g.pegCount+v > CribbagePegLimit {
			continue
		}
		seq[len(seq)-1] = c
		score := CribbageScorePegging(seq, g.pegCount+v)
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		return nil
	}
	return &CribbageHint{Type: "play", Indices: []int{bestIdx}}
}
