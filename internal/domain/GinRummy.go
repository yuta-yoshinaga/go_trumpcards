package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// GinRummyPlayerCnt ジンラミープレイヤー数
const GinRummyPlayerCnt = 2

// GinRummyHandSize 初期配布枚数
const GinRummyHandSize = 10

// GinRummyKnockThreshold ノック可能なデッドウッド上限
const GinRummyKnockThreshold = 10

// GinRummyGinBonus ジンボーナス点
const GinRummyGinBonus = 25

// GinRummyBigGinBonus ビッグジンボーナス点
const GinRummyBigGinBonus = 31

// GinRummyUndercutBonus アンダーカットボーナス点
const GinRummyUndercutBonus = 25

// GinRummyPhase ゲームフェーズ
type GinRummyPhase int

// GinRummyのフェーズ定数
const (
	// GinRummyPhaseDraw ドローフェーズ (山札または捨て札から引く)
	GinRummyPhaseDraw GinRummyPhase = 0
	// GinRummyPhaseDiscard ディスカードフェーズ (手札から1枚捨てる or ノック/ジン)
	GinRummyPhaseDiscard GinRummyPhase = 1
	// GinRummyPhaseLayoff レイオフフェーズ (相手がノッカーのメルドにカードを付ける)
	GinRummyPhaseLayoff GinRummyPhase = 2
	// GinRummyPhaseRoundEnd ラウンド終了フェーズ
	GinRummyPhaseRoundEnd GinRummyPhase = 3
	// GinRummyPhaseGameEnd ゲーム終了フェーズ
	GinRummyPhaseGameEnd GinRummyPhase = 4
)

// GinRummy ジンラミーゲームクラス
type GinRummy struct {
	trumpCards       *TrumpCards
	players          []*GinRummyPlayer
	config           GinRummyConfig
	phase            GinRummyPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	actionLogBase
	knockerIdx      int       // ノックしたプレイヤーのインデックス (-1 = ノックなし)
	knockerMelds    [][]*Card // ノッカーのメルド (レイオフ用)
	knockerDeadwood []*Card   // ノッカーのデッドウッド
	isGin           bool      // ジンかどうか
}

// NewGinRummy コンストラクタ
func NewGinRummy(trumpCards *TrumpCards, players []*GinRummyPlayer, config GinRummyConfig) *GinRummy {
	return &GinRummy{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
		knockerIdx:  -1,
	}
}

// NewDefaultGinRummy returns GinRummy with the standard 2-player setup (1 human, 1 CPU)
// and DefaultGinRummyConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultGinRummy() *GinRummy {
	players := []*GinRummyPlayer{
		NewGinRummyPlayer(true),
		NewGinRummyPlayer(false),
	}
	return NewGinRummy(NewTrumpCards(0), players, DefaultGinRummyConfig())
}

// Reset ゲーム初期化
func (g *GinRummy) Reset() {
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
	g.isGin = false

	for _, p := range g.players {
		p.roundScore = 0
		p.cumulativeScore = 0
		p.Reset()
		p.SetIsFinished(false)
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = GinRummyPhaseDraw
}

// NextRound 次のラウンドを開始する
func (g *GinRummy) NextRound() {
	if g.phase != GinRummyPhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0
	g.knockerIdx = -1
	g.knockerMelds = nil
	g.knockerDeadwood = nil
	g.isGin = false

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = GinRummyPhaseDraw
}

// dealInitialCards 初期配布: 各プレイヤーに10枚、1枚を捨て札に
func (g *GinRummy) dealInitialCards() {
	g.drawPile = make([]*Card, 0, g.trumpCards.GetTotalCount())
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

	// 各プレイヤーに10枚配布
	for i := 0; i < GinRummyHandSize; i++ {
		for j := 0; j < GinRummyPlayerCnt; j++ {
			if len(g.drawPile) > 0 {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				g.players[j].AddCard(card)
			}
		}
	}

	// 最初の1枚を捨て札に
	if len(g.drawPile) > 0 {
		firstCard := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, firstCard)
	}
}

// PlayerDrawFromStock 人間プレイヤーが山札からカードを引く
func (g *GinRummy) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GinRummyPhaseDraw {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(g.drawPile) == 0 {
		// 山札なし → ラウンド引き分け
		g.endRoundDraw()
		return nil
	}

	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)

	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, g.currentPlayerIdx)), nil)

	g.phase = GinRummyPhaseDiscard
	return nil
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札からカードを引く
func (g *GinRummy) PlayerDrawFromDiscard() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GinRummyPhaseDraw {
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

	g.phase = GinRummyPhaseDiscard
	return nil
}

// PlayerDiscard 人間プレイヤーがカードを捨てる
func (g *GinRummy) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GinRummyPhaseDiscard {
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

// PlayerKnock 人間プレイヤーがノックする (カードを1枚捨ててノック)
func (g *GinRummy) PlayerKnock(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GinRummyPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	// ノック後のデッドウッド計算: カードを仮に除外してデッドウッドをチェック
	testCards := make([]*Card, 0, player.GetCardsSize()-1)
	for i := 0; i < player.GetCardsSize(); i++ {
		if i != cardIndex {
			testCards = append(testCards, player.GetCard(i))
		}
	}

	melds, deadwood := FindBestMelds(testCards)
	deadwoodValue := CalcDeadwoodValue(deadwood)

	if deadwoodValue > GinRummyKnockThreshold {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("デッドウッドが%d点以下でないとノックできません（現在%d点）", GinRummyKnockThreshold, deadwoodValue))
	}

	discarded := player.RemoveCard(cardIndex)
	g.discardPile = append(g.discardPile, discarded)

	g.knockerIdx = g.currentPlayerIdx
	g.knockerMelds = melds
	g.knockerDeadwood = deadwood
	g.isGin = deadwoodValue == 0

	g.appendLog(g.currentPlayerIdx, "knock", fmt.Sprintf("%s knocks (deadwood: %d)", playerName(g.players, g.currentPlayerIdx), deadwoodValue), []*Card{discarded})

	if g.isGin {
		// ジン → レイオフなしでスコアリング
		g.appendLog(g.currentPlayerIdx, "gin", fmt.Sprintf("%s has Gin!", playerName(g.players, g.currentPlayerIdx)), nil)
		g.scoreRound()
	} else {
		// 相手のレイオフフェーズへ
		g.phase = GinRummyPhaseLayoff
		g.currentPlayerIdx = 1 - g.currentPlayerIdx
	}
	return nil
}

// PlayerLayoff 人間プレイヤーがレイオフするカードを選択する (空のindicesでレイオフ終了)
func (g *GinRummy) PlayerLayoff(cardIndices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GinRummyPhaseLayoff {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if len(cardIndices) == 0 {
		// レイオフ終了
		g.scoreRound()
		return nil
	}

	player := g.players[g.currentPlayerIdx]

	// インデックスのバリデーション
	for _, idx := range cardIndices {
		if idx < 0 || idx >= player.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
	}

	// 重複チェック
	seen := make(map[int]bool)
	for _, idx := range cardIndices {
		if seen[idx] {
			return NewDomainError(ErrInvalidCard, "カードインデックスが重複しています")
		}
		seen[idx] = true
	}

	// レイオフ可能かチェック: 各カードがノッカーのメルドのいずれかに付けられるか
	for _, idx := range cardIndices {
		card := player.GetCard(idx)
		if !g.canLayoff(card) {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("%sはレイオフできません", cardStr(card)))
		}
	}

	// レイオフ実行 (降順にremoveして安全に)
	sort.Sort(sort.Reverse(sort.IntSlice(cardIndices)))
	for _, idx := range cardIndices {
		card := player.GetCard(idx)
		g.layoffCard(card)
		player.RemoveCard(idx)
		g.appendLog(g.currentPlayerIdx, "layoff", fmt.Sprintf("%s lays off %s", playerName(g.players, g.currentPlayerIdx), cardStr(card)), []*Card{card})
	}

	g.scoreRound()
	return nil
}

// LayoffTargets はその札を足せるノッカーのメルド番号をすべて返す。
//
// **レイオフフェーズの主題は「どのメルドに付け足せるか」。**画面はその補助を
// 一切持たず、押してサーバーの応答で初めて成否が分かる状態だった (#4823)。
func (g *GinRummy) LayoffTargets(card *Card) []int {
	if card == nil {
		return nil
	}
	var out []int
	for i, meld := range g.knockerMelds {
		if canAddToMeld(meld, card) {
			out = append(out, i)
		}
	}
	return out
}

// canLayoff カードがノッカーのメルドにレイオフ可能か
func (g *GinRummy) canLayoff(card *Card) bool {
	// 表示側と同じ判定を使う。2 本置くと「置ける」と出しながら拒否される状態が作れる。
	return len(g.LayoffTargets(card)) > 0
}

// layoffCard カードをノッカーのメルドに追加する
func (g *GinRummy) layoffCard(card *Card) {
	for i, meld := range g.knockerMelds {
		if canAddToMeld(meld, card) {
			g.knockerMelds[i] = append(meld, card)
			return
		}
	}
}

// canAddToMeld メルドにカードを追加可能か
func canAddToMeld(meld []*Card, card *Card) bool {
	if len(meld) == 0 {
		return false
	}

	// セット (同ランク) かランかを判定
	if isSet(meld) {
		// セットには同ランクで別スートのカードを追加可能 (最大4枚)
		if len(meld) >= 4 {
			return false
		}
		if card.GetValue() != meld[0].GetValue() {
			return false
		}
		// 同じスートが既にあるか
		for _, m := range meld {
			if m.GetDesign() == card.GetDesign() {
				return false
			}
		}
		return true
	}

	// ラン: 同スートで連続したカードを追加可能
	if card.GetDesign() != meld[0].GetDesign() {
		return false
	}

	// メルドの最小・最大値を求める
	minVal, maxVal := meld[0].GetValue(), meld[0].GetValue()
	for _, m := range meld[1:] {
		if m.GetValue() < minVal {
			minVal = m.GetValue()
		}
		if m.GetValue() > maxVal {
			maxVal = m.GetValue()
		}
	}

	// カードがランの前後に追加可能か
	return card.GetValue() == minVal-1 || card.GetValue() == maxVal+1
}

// isSet メルドがセット (同ランク) かどうか
func isSet(meld []*Card) bool {
	if len(meld) < 2 {
		return false
	}
	for i := 1; i < len(meld); i++ {
		if meld[i].GetValue() != meld[0].GetValue() {
			return false
		}
	}
	return true
}

// CpuPlay 現在の手番がCPUの場合にターンを実行
func (g *GinRummy) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	switch g.phase {
	case GinRummyPhaseDraw:
		g.cpuDraw()
	case GinRummyPhaseDiscard:
		g.cpuDiscardOrKnock()
	case GinRummyPhaseLayoff:
		g.cpuLayoff()
	}
}

// cpuDraw CPUがドローする
func (g *GinRummy) cpuDraw() {
	// 捨て札の一番上がメルドに使えるか判断
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
		case GinRummyCpuDifficultyHard:
			shouldPickDiscard = dwWith < dwWithout
		case GinRummyCpuDifficultyNormal:
			shouldPickDiscard = dwWith < dwWithout-5
		default:
			shouldPickDiscard = rand.Intn(3) == 0
		}

		if shouldPickDiscard {
			card := g.discardPile[len(g.discardPile)-1]
			g.discardPile = g.discardPile[:len(g.discardPile)-1]
			g.players[g.currentPlayerIdx].AddCard(card)
			g.sortHand(g.currentPlayerIdx)
			g.appendLog(g.currentPlayerIdx, "draw_discard", fmt.Sprintf("%s draws %s from discard", playerName(g.players, g.currentPlayerIdx), cardStr(card)), []*Card{card})
			g.phase = GinRummyPhaseDiscard
			return
		}
	}

	// 山札から引く
	if len(g.drawPile) == 0 {
		g.endRoundDraw()
		return
	}

	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[g.currentPlayerIdx].AddCard(card)
	g.sortHand(g.currentPlayerIdx)
	g.appendLog(g.currentPlayerIdx, "draw_stock", fmt.Sprintf("%s draws from stock", playerName(g.players, g.currentPlayerIdx)), nil)
	g.phase = GinRummyPhaseDiscard
}

// cpuDiscardOrKnock CPUがディスカードまたはノックする
func (g *GinRummy) cpuDiscardOrKnock() {
	player := g.players[g.currentPlayerIdx]

	// 全カードを取得
	cards := make([]*Card, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		cards[i] = player.GetCard(i)
	}

	// 最適なディスカードを見つける: 各カードを除外してデッドウッドを最小化
	bestDiscardIdx := 0
	bestDeadwood := -1

	for i := 0; i < player.GetCardsSize(); i++ {
		testCards := make([]*Card, 0, player.GetCardsSize()-1)
		for j := 0; j < player.GetCardsSize(); j++ {
			if j != i {
				testCards = append(testCards, player.GetCard(j))
			}
		}
		_, dw := FindBestMelds(testCards)
		dwVal := CalcDeadwoodValue(dw)

		if bestDeadwood < 0 || dwVal < bestDeadwood {
			bestDeadwood = dwVal
			bestDiscardIdx = i
		}
	}

	// ノック判定
	if bestDeadwood <= GinRummyKnockThreshold {
		shouldKnock := false
		switch g.config.CpuDifficulty {
		case GinRummyCpuDifficultyHard:
			shouldKnock = bestDeadwood <= 5 || bestDeadwood == 0
		case GinRummyCpuDifficultyNormal:
			shouldKnock = bestDeadwood <= 7
		default:
			shouldKnock = true
		}

		if shouldKnock {
			// ノック実行
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
			g.isGin = deadwoodValue == 0

			g.appendLog(g.currentPlayerIdx, "knock", fmt.Sprintf("%s knocks (deadwood: %d)", playerName(g.players, g.currentPlayerIdx), deadwoodValue), []*Card{discarded})

			if g.isGin {
				g.appendLog(g.currentPlayerIdx, "gin", fmt.Sprintf("%s has Gin!", playerName(g.players, g.currentPlayerIdx)), nil)
				g.scoreRound()
			} else {
				g.phase = GinRummyPhaseLayoff
				g.currentPlayerIdx = 1 - g.currentPlayerIdx
			}
			return
		}
	}

	// 通常のディスカード
	discarded := player.RemoveCard(bestDiscardIdx)
	g.discardPile = append(g.discardPile, discarded)
	g.appendLog(g.currentPlayerIdx, "discard", fmt.Sprintf("%s discards %s", playerName(g.players, g.currentPlayerIdx), cardStr(discarded)), []*Card{discarded})
	g.advanceTurn()
}

// cpuLayoff CPUがレイオフする
func (g *GinRummy) cpuLayoff() {
	player := g.players[g.currentPlayerIdx]

	// レイオフ可能なカードをすべてレイオフ
	for {
		found := false
		for i := 0; i < player.GetCardsSize(); i++ {
			card := player.GetCard(i)
			if g.canLayoff(card) {
				g.layoffCard(card)
				player.RemoveCard(i)
				g.appendLog(g.currentPlayerIdx, "layoff", fmt.Sprintf("%s lays off %s", playerName(g.players, g.currentPlayerIdx), cardStr(card)), []*Card{card})
				found = true
				break // re-iterate from start since indices shifted
			}
		}
		if !found {
			break
		}
	}

	g.scoreRound()
}

// scoreRound ラウンドのスコアを確定する
func (g *GinRummy) scoreRound() {
	knockerIdx := g.knockerIdx
	opponentIdx := 1 - knockerIdx

	opponent := g.players[opponentIdx]

	// 相手のデッドウッドを計算
	opponentCards := make([]*Card, opponent.GetCardsSize())
	for i := 0; i < opponent.GetCardsSize(); i++ {
		opponentCards[i] = opponent.GetCard(i)
	}
	_, opponentDeadwood := FindBestMelds(opponentCards)
	opponentDeadwoodValue := CalcDeadwoodValue(opponentDeadwood)

	knockerDeadwoodValue := CalcDeadwoodValue(g.knockerDeadwood)

	if g.isGin {
		// ジン: ノッカーが相手のデッドウッド + ボーナスを獲得
		score := opponentDeadwoodValue + GinRummyGinBonus
		g.players[knockerIdx].roundScore = score
		g.appendLog(knockerIdx, "score", fmt.Sprintf("%s scores %d (Gin bonus %d + deadwood %d)", playerName(g.players, knockerIdx), score, GinRummyGinBonus, opponentDeadwoodValue), nil)
	} else if opponentDeadwoodValue <= knockerDeadwoodValue {
		// アンダーカット: 相手がデッドウッド差 + ボーナスを獲得
		score := knockerDeadwoodValue - opponentDeadwoodValue + GinRummyUndercutBonus
		g.players[opponentIdx].roundScore = score
		g.appendLog(opponentIdx, "undercut", fmt.Sprintf("%s undercuts! Scores %d (bonus %d + difference %d)", playerName(g.players, opponentIdx), score, GinRummyUndercutBonus, knockerDeadwoodValue-opponentDeadwoodValue), nil)
	} else {
		// 通常ノック: ノッカーがデッドウッド差を獲得
		score := opponentDeadwoodValue - knockerDeadwoodValue
		g.players[knockerIdx].roundScore = score
		g.appendLog(knockerIdx, "score", fmt.Sprintf("%s scores %d (deadwood difference)", playerName(g.players, knockerIdx), score), nil)
	}

	// 累積スコアに加算
	for i := range g.players {
		g.players[i].CommitRoundScore()
	}

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = GinRummyPhaseRoundEnd
	}
}

// endRoundDraw 山札切れによる引き分け (スコアなし)
func (g *GinRummy) endRoundDraw() {
	g.appendLog(-1, "draw", "Round ends in a draw (stock empty)", nil)
	g.knockerIdx = -1

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = GinRummyPhaseRoundEnd
	}
}

// ScoreRound ラウンドのスコア処理 (NextRoundから呼ぶ用。既にscoreRoundで処理済みの場合はnoop)
func (g *GinRummy) ScoreRound() {
	// スコアリングはknock/gin/layoff時に既に完了している
	// このメソッドは互換性のために存在
}

// advanceTurn 次のプレイヤーへ
func (g *GinRummy) advanceTurn() {
	g.currentPlayerIdx = 1 - g.currentPlayerIdx
	g.phase = GinRummyPhaseDraw
}

// checkGameEnd ゲーム終了判定
func (g *GinRummy) checkGameEnd() {
	hasWinner := false
	for i := 0; i < GinRummyPlayerCnt; i++ {
		if g.players[i].cumulativeScore >= g.config.PointLimit {
			hasWinner = true
			break
		}
	}

	if !hasWinner {
		return
	}

	g.gameEndFlag = true
	g.phase = GinRummyPhaseGameEnd

	// 最高スコアのプレイヤーが勝者
	maxScore := g.players[0].cumulativeScore
	g.winnerIdx = 0
	for i := 1; i < GinRummyPlayerCnt; i++ {
		if g.players[i].cumulativeScore > maxScore {
			maxScore = g.players[i].cumulativeScore
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(g.players, g.winnerIdx)), nil)
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *GinRummy) GetPhase() GinRummyPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *GinRummy) SetPhase(phase GinRummyPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *GinRummy) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *GinRummy) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *GinRummy) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *GinRummy) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *GinRummy) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *GinRummy) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDiscardTop 捨て札の一番上を取得
func (g *GinRummy) GetDiscardTop() *Card {
	return discardTop(g.discardPile)
}

// GetDrawPileCount 山札の残り枚数取得
func (g *GinRummy) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定 (テスト用)
func (g *GinRummy) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *GinRummy) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (g *GinRummy) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *GinRummy) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *GinRummy) GetPlayer(i int) *GinRummyPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *GinRummy) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *GinRummy) GetConfig() GinRummyConfig { return g.config }

// SetConfig 設定変更
func (g *GinRummy) SetConfig(cfg GinRummyConfig) { g.config = cfg }

// GetKnockerIdx ノッカーのインデックス取得
func (g *GinRummy) GetKnockerIdx() int { return g.knockerIdx }

// SetKnockerIdx ノッカーのインデックス設定 (テスト用)
func (g *GinRummy) SetKnockerIdx(idx int) { g.knockerIdx = idx }

// GetKnockerMelds ノッカーのメルド取得
func (g *GinRummy) GetKnockerMelds() [][]*Card { return g.knockerMelds }

// SetKnockerMelds ノッカーのメルドを設定 (テスト用)
func (g *GinRummy) SetKnockerMelds(melds [][]*Card) { g.knockerMelds = melds }

// GetKnockerDeadwood ノッカーのデッドウッド取得
func (g *GinRummy) GetKnockerDeadwood() []*Card { return g.knockerDeadwood }

// SetKnockerDeadwood ノッカーのデッドウッドを設定 (テスト用)
func (g *GinRummy) SetKnockerDeadwood(deadwood []*Card) { g.knockerDeadwood = deadwood }

// GetIsGin ジンかどうか取得
func (g *GinRummy) GetIsGin() bool { return g.isGin }

// SetIsGin ジンを設定 (テスト用)
func (g *GinRummy) SetIsGin(isGin bool) { g.isGin = isGin }

// --- Private methods ---

// sortAllHands 全プレイヤーの手札をソートする
func (g *GinRummy) sortAllHands() {
	for i := range g.players {
		g.sortHand(i)
	}
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func (g *GinRummy) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// --- Meld Detection ---

// GinRummyCardValue カードの点数を返す (A=1, 2-9=face value, 10/J/Q/K=10)
func GinRummyCardValue(card *Card) int {
	v := card.GetValue()
	if v == 1 { // Ace
		return 1
	}
	if v >= 10 { // 10, J, Q, K
		return 10
	}
	return v
}

// CalcDeadwoodValue デッドウッドの合計点数を計算する
func CalcDeadwoodValue(deadwood []*Card) int {
	total := 0
	for _, card := range deadwood {
		total += GinRummyCardValue(card)
	}
	return total
}

// FindBestMelds カードの最適なメルド分割を見つける (デッドウッドを最小化)
func FindBestMelds(cards []*Card) (melds [][]*Card, deadwood []*Card) {
	if len(cards) == 0 {
		return nil, nil
	}

	bestMelds, bestDeadwood := findBestMeldsRecursive(cards, nil)
	return bestMelds, bestDeadwood
}

// findBestMeldsRecursive 再帰的にメルドを見つける
func findBestMeldsRecursive(remaining []*Card, currentMelds [][]*Card) ([][]*Card, []*Card) {
	// 候補となるメルドをすべて列挙
	candidates := findAllPossibleMelds(remaining)

	if len(candidates) == 0 {
		// メルドが見つからない → 残りはすべてデッドウッド
		return currentMelds, remaining
	}

	// 各候補を試す
	bestDeadwoodValue := CalcDeadwoodValue(remaining)
	bestMelds := currentMelds
	bestDeadwood := remaining

	for _, meld := range candidates {
		// メルドに使ったカードを除外
		rest := excludeCards(remaining, meld)
		newMelds := append(copyMelds(currentMelds), meld)

		// 再帰
		resultMelds, resultDeadwood := findBestMeldsRecursive(rest, newMelds)
		dv := CalcDeadwoodValue(resultDeadwood)

		if dv < bestDeadwoodValue {
			bestDeadwoodValue = dv
			bestMelds = resultMelds
			bestDeadwood = resultDeadwood
		}

		if bestDeadwoodValue == 0 {
			break // 完全メルド → 最適解
		}
	}

	return bestMelds, bestDeadwood
}

// findAllPossibleMelds 利用可能なメルドをすべて列挙する
func findAllPossibleMelds(cards []*Card) [][]*Card {
	var melds [][]*Card

	// セットを探す (同ランク3枚以上)
	byRank := make(map[int][]*Card)
	for _, c := range cards {
		byRank[c.GetValue()] = append(byRank[c.GetValue()], c)
	}
	for _, group := range byRank {
		if len(group) >= 3 {
			melds = append(melds, group[:3])
			if len(group) >= 4 {
				melds = append(melds, group[:4])
			}
		}
	}

	// ランを探す (同スート連続3枚以上)
	bySuit := make(map[int][]*Card)
	for _, c := range cards {
		bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], c)
	}
	for _, group := range bySuit {
		if len(group) < 3 {
			continue
		}
		// 値でソート
		sorted := make([]*Card, len(group))
		copy(sorted, group)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].GetValue() < sorted[j].GetValue()
		})

		// 連続するカードを見つける
		for i := 0; i < len(sorted); i++ {
			run := []*Card{sorted[i]}
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j].GetValue() == run[len(run)-1].GetValue()+1 {
					run = append(run, sorted[j])
					if len(run) >= 3 {
						runCopy := make([]*Card, len(run))
						copy(runCopy, run)
						melds = append(melds, runCopy)
					}
				} else {
					break
				}
			}
		}
	}

	return melds
}

// excludeCards 元の配列からメルドに使ったカードを除外する
func excludeCards(cards []*Card, meld []*Card) []*Card {
	used := make(map[*Card]bool)
	for _, c := range meld {
		used[c] = true
	}
	var rest []*Card
	for _, c := range cards {
		if !used[c] {
			rest = append(rest, c)
		}
	}
	return rest
}

// copyMelds メルドのスライスをコピーする
func copyMelds(melds [][]*Card) [][]*Card {
	if melds == nil {
		return nil
	}
	result := make([][]*Card, len(melds))
	copy(result, melds)
	return result
}

// ginRummyJSON is the JSON wire format for GinRummy.
type ginRummyJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*GinRummyPlayer `json:"pl"`
	Config           GinRummyConfig    `json:"cf"`
	Phase            GinRummyPhase     `json:"ps"`
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
	IsGin            bool              `json:"ig"`
}

// MarshalJSON implements json.Marshaler.
func (g *GinRummy) MarshalJSON() ([]byte, error) {
	return json.Marshal(ginRummyJSON{
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
		IsGin:            g.isGin,
	})
}

// ginRummyMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const ginRummyMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (g *GinRummy) UnmarshalJSON(data []byte) error {
	var j ginRummyJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > ginRummyMaxSliceLen || len(j.DiscardPile) > ginRummyMaxSliceLen ||
		len(j.DrawPile) > ginRummyMaxSliceLen || len(j.ActionLog) > ginRummyMaxSliceLen ||
		len(j.KnockerMelds) > ginRummyMaxSliceLen || len(j.KnockerDeadwood) > ginRummyMaxSliceLen {
		return fmt.Errorf("ginrummy: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*GinRummyPlayer, 0)
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
	g.isGin = j.IsGin
	return nil
}
