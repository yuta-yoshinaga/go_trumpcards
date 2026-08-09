package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// TonkPlayerCnt Tonkプレイヤー数
const TonkPlayerCnt = 2

// TonkHandSize 初期配布枚数
const TonkHandSize = 5

// TonkKnockThreshold ノック可能なデッドウッド上限
const TonkKnockThreshold = 5

// TonkOnDealLow 配牌時のTonk(即勝利)成立点(下限)
const TonkOnDealLow = 49

// TonkOnDealHigh 配牌時のTonk(即勝利)成立点(上限)
const TonkOnDealHigh = 50

// TonkBonus 配牌時Tonk成立時のボーナス点
const TonkBonus = 50

// TonkUndercutPenalty アンダーカット時のペナルティ点
const TonkUndercutPenalty = 5

// TonkPhase ゲームフェーズ
type TonkPhase int

// Tonkのフェーズ定数
const (
	// TonkPhaseDraw ドローフェーズ (山札または捨て札から引く)
	TonkPhaseDraw TonkPhase = 0
	// TonkPhaseDiscard ディスカードフェーズ (手札から1枚捨てる or ノック)
	TonkPhaseDiscard TonkPhase = 1
	// TonkPhaseRoundEnd ラウンド終了フェーズ
	TonkPhaseRoundEnd TonkPhase = 2
	// TonkPhaseGameEnd ゲーム終了フェーズ
	TonkPhaseGameEnd TonkPhase = 3
)

// Tonk Tonkゲームクラス
type Tonk struct {
	trumpCards       *TrumpCards
	players          []*TonkPlayer
	config           TonkConfig
	phase            TonkPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	actionLogBase
	knockerIdx       int       // ノック/Tonkしたプレイヤーのインデックス (-1 = 未確定)
	knockerMelds     [][]*Card // ノッカーのメルド
	knockerDeadwood  []*Card   // ノッカーのデッドウッド
	opponentMelds    [][]*Card // 相手のメルド (スコア確定時に格納)
	opponentDeadwood []*Card   // 相手のデッドウッド
	isTonk           bool      // 配牌Tonk(49/50点即勝利)かどうか
	isUndercut       bool      // アンダーカット(ノッカーが負け)かどうか
	rng              *rand.Rand
}

// NewTonk コンストラクタ
func NewTonk(trumpCards *TrumpCards, players []*TonkPlayer, config TonkConfig) *Tonk {
	return &Tonk{
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
func (g *Tonk) SetRand(r *rand.Rand) {
	g.rng = r
}

// NewDefaultTonk returns Tonk with the standard 2-player setup (1 human, 1 CPU)
// and DefaultTonkConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultTonk() *Tonk {
	players := []*TonkPlayer{
		NewTonkPlayer(true),
		NewTonkPlayer(false),
	}
	return NewTonk(NewTrumpCards(0), players, DefaultTonkConfig())
}

// Reset ゲーム初期化
func (g *Tonk) Reset() {
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
	g.isTonk = false
	g.isUndercut = false

	for _, p := range g.players {
		p.SetRoundScore(0)
		p.SetCumulativeScore(0)
		p.Reset()
		p.SetIsFinished(false)
	}

	g.dealInitialCards()
	g.sortAllHands()

	g.phase = TonkPhaseDraw
	g.checkTonkOnDeal()
}

// NextRound 次のラウンドを開始する
func (g *Tonk) NextRound() {
	if g.phase != TonkPhaseRoundEnd {
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
	g.isTonk = false
	g.isUndercut = false

	for _, p := range g.players {
		p.ResetRound()
	}

	g.dealInitialCards()
	g.sortAllHands()

	g.phase = TonkPhaseDraw
	g.checkTonkOnDeal()
}

// dealInitialCards 初期配布: 各プレイヤーにTonkHandSize枚、1枚を捨て札に
func (g *Tonk) dealInitialCards() {
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

	for i := 0; i < TonkHandSize; i++ {
		for j := 0; j < TonkPlayerCnt; j++ {
			if len(g.drawPile) > 0 {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				g.players[j].AddCard(card)
			}
		}
	}

	if len(g.drawPile) > 0 {
		firstCard := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, firstCard)
	}
}

// checkTonkOnDeal 配牌Tonk(49/50点即勝利)を判定する
func (g *Tonk) checkTonkOnDeal() {
	for i, p := range g.players {
		total := 0
		for k := 0; k < p.GetCardsSize(); k++ {
			total += GinRummyCardValue(p.GetCard(k))
		}
		if total == TonkOnDealLow || total == TonkOnDealHigh {
			g.isTonk = true
			g.knockerIdx = i
			g.appendLog(i, "tonk_on_deal", fmt.Sprintf("%s declares Tonk on deal! (hand value: %d)", playerName(g.players, i), total), nil)
			g.scoreTonk(total)
			return
		}
	}
}

// PlayerDrawFromStock 人間プレイヤーが山札からカードを引く
func (g *Tonk) PlayerDrawFromStock() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TonkPhaseDraw {
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

	g.phase = TonkPhaseDiscard
	return nil
}

// PlayerDrawFromDiscard 人間プレイヤーが捨て札からカードを引く
func (g *Tonk) PlayerDrawFromDiscard() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TonkPhaseDraw {
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

	g.phase = TonkPhaseDiscard
	return nil
}

// PlayerDiscard 人間プレイヤーがカードを捨てる
func (g *Tonk) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TonkPhaseDiscard {
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
func (g *Tonk) PlayerKnock(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TonkPhaseDiscard {
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

	if deadwoodValue > TonkKnockThreshold {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("デッドウッドが%d点以下でないとノックできません（現在%d点）", TonkKnockThreshold, deadwoodValue))
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
func (g *Tonk) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	if g.phase != TonkPhaseDraw && g.phase != TonkPhaseDiscard {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	switch g.phase {
	case TonkPhaseDraw:
		g.cpuDraw()
	case TonkPhaseDiscard:
		g.cpuDiscardOrKnock()
	}
}

// cpuDraw CPUがドローする
func (g *Tonk) cpuDraw() {
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
		case TonkCpuDifficultyHard:
			shouldPickDiscard = dwWith < dwWithout
		case TonkCpuDifficultyNormal:
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
			g.phase = TonkPhaseDiscard
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
	g.phase = TonkPhaseDiscard
}

// cpuDiscardOrKnock CPUがディスカードまたはノックする
func (g *Tonk) cpuDiscardOrKnock() {
	player := g.players[g.currentPlayerIdx]

	bestDeadwood, bestDiscardIdx := g.GetBestDeadwood(g.currentPlayerIdx)
	if bestDiscardIdx < 0 {
		bestDiscardIdx = 0
	}

	if bestDeadwood <= TonkKnockThreshold {
		shouldKnock := false
		switch g.config.CpuDifficulty {
		case TonkCpuDifficultyHard:
			shouldKnock = bestDeadwood <= 3
		case TonkCpuDifficultyNormal:
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
func (g *Tonk) scoreRound() {
	knockerIdx := g.knockerIdx
	opponentIdx := 1 - knockerIdx

	opponent := g.players[opponentIdx]

	opponentCards := make([]*Card, opponent.GetCardsSize())
	for i := 0; i < opponent.GetCardsSize(); i++ {
		opponentCards[i] = opponent.GetCard(i)
	}
	oppMelds, oppDeadwood := FindBestMelds(opponentCards)
	g.opponentMelds = oppMelds
	g.opponentDeadwood = oppDeadwood
	opponentDeadwoodValue := CalcDeadwoodValue(oppDeadwood)

	knockerDeadwoodValue := CalcDeadwoodValue(g.knockerDeadwood)

	if opponentDeadwoodValue < knockerDeadwoodValue {
		// アンダーカット: 相手がデッドウッド差 + ペナルティを獲得
		g.isUndercut = true
		score := knockerDeadwoodValue - opponentDeadwoodValue + TonkUndercutPenalty
		g.players[opponentIdx].SetRoundScore(score)
		g.appendLog(opponentIdx, "undercut", fmt.Sprintf("%s undercuts! Scores %d (penalty %d + difference %d)", playerName(g.players, opponentIdx), score, TonkUndercutPenalty, knockerDeadwoodValue-opponentDeadwoodValue), nil)
	} else {
		score := opponentDeadwoodValue - knockerDeadwoodValue
		g.players[knockerIdx].SetRoundScore(score)
		g.appendLog(knockerIdx, "score", fmt.Sprintf("%s scores %d (deadwood difference)", playerName(g.players, knockerIdx), score), nil)
	}

	for i := range g.players {
		g.players[i].CommitRoundScore()
	}

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = TonkPhaseRoundEnd
	}
}

// scoreTonk 配牌Tonk成立時のスコア処理
func (g *Tonk) scoreTonk(handValue int) {
	knockerIdx := g.knockerIdx
	opponentIdx := 1 - knockerIdx

	opponent := g.players[opponentIdx]
	opponentCards := make([]*Card, opponent.GetCardsSize())
	for i := 0; i < opponent.GetCardsSize(); i++ {
		opponentCards[i] = opponent.GetCard(i)
	}
	oppMelds, oppDeadwood := FindBestMelds(opponentCards)
	g.opponentMelds = oppMelds
	g.opponentDeadwood = oppDeadwood

	// ノッカー(Tonk宣言者)の手札はメルド扱いせず、参考のためそのまま記録する
	g.knockerMelds = nil
	knockerCards := make([]*Card, g.players[knockerIdx].GetCardsSize())
	for i := 0; i < g.players[knockerIdx].GetCardsSize(); i++ {
		knockerCards[i] = g.players[knockerIdx].GetCard(i)
	}
	g.knockerDeadwood = knockerCards

	score := TonkBonus + handValue
	g.players[knockerIdx].SetRoundScore(score)
	g.appendLog(knockerIdx, "tonk_score", fmt.Sprintf("%s scores %d (Tonk bonus %d + hand %d)", playerName(g.players, knockerIdx), score, TonkBonus, handValue), nil)

	for i := range g.players {
		g.players[i].CommitRoundScore()
	}

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = TonkPhaseRoundEnd
	}
}

// endRoundDraw 山札切れによる引き分け (スコアなし)
func (g *Tonk) endRoundDraw() {
	g.appendLog(-1, "draw", "Round ends in a draw (stock empty)", nil)
	g.knockerIdx = -1

	g.checkGameEnd()
	if !g.gameEndFlag {
		g.phase = TonkPhaseRoundEnd
	}
}

// ScoreRound ラウンドのスコア処理 (NextRoundから呼ぶ用。既にscoreRoundで処理済みの場合はnoop)
func (g *Tonk) ScoreRound() {
	// スコアリングはknock/tonk時に既に完了している
	// このメソッドは互換性のために存在
}

// advanceTurn 次のプレイヤーへ
func (g *Tonk) advanceTurn() {
	g.currentPlayerIdx = 1 - g.currentPlayerIdx
	g.phase = TonkPhaseDraw
}

// checkGameEnd ゲーム終了判定
func (g *Tonk) checkGameEnd() {
	hasWinner := false
	for i := 0; i < TonkPlayerCnt; i++ {
		if g.players[i].GetCumulativeScore() >= g.config.PointLimit {
			hasWinner = true
			break
		}
	}

	if !hasWinner {
		return
	}

	g.gameEndFlag = true
	g.phase = TonkPhaseGameEnd

	maxScore := g.players[0].GetCumulativeScore()
	g.winnerIdx = 0
	for i := 1; i < TonkPlayerCnt; i++ {
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
// CUI の表示 (tonkBestDeadwood) が別々に同じループを持っており、Web にも
// 3つ目を書くところだった。TonkKnockThreshold と比べる値なので、実装が割れると
// 「ノック可能と表示したのに弾かれる」ずれになる。
func (g *Tonk) GetBestDeadwood(playerIdx int) (best int, discardIdx int) {
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

func (g *Tonk) GetPhase() TonkPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Tonk) SetPhase(phase TonkPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Tonk) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Tonk) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Tonk) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Tonk) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *Tonk) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *Tonk) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDiscardTop 捨て札の一番上を取得
func (g *Tonk) GetDiscardTop() *Card {
	if len(g.discardPile) == 0 {
		return nil
	}
	return g.discardPile[len(g.discardPile)-1]
}

// GetDrawPileCount 山札の残り枚数取得
func (g *Tonk) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定 (テスト用)
func (g *Tonk) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Tonk) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (g *Tonk) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Tonk) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Tonk) GetPlayer(i int) *TonkPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Tonk) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *Tonk) GetConfig() TonkConfig { return g.config }

// SetConfig 設定変更
func (g *Tonk) SetConfig(cfg TonkConfig) { g.config = cfg }

// GetKnockerIdx ノッカーのインデックス取得
func (g *Tonk) GetKnockerIdx() int { return g.knockerIdx }

// SetKnockerIdx ノッカーのインデックス設定 (テスト用)
func (g *Tonk) SetKnockerIdx(idx int) { g.knockerIdx = idx }

// GetKnockerMelds ノッカーのメルド取得
func (g *Tonk) GetKnockerMelds() [][]*Card { return g.knockerMelds }

// SetKnockerMelds ノッカーのメルドを設定 (テスト用)
func (g *Tonk) SetKnockerMelds(melds [][]*Card) { g.knockerMelds = melds }

// GetKnockerDeadwood ノッカーのデッドウッド取得
func (g *Tonk) GetKnockerDeadwood() []*Card { return g.knockerDeadwood }

// SetKnockerDeadwood ノッカーのデッドウッドを設定 (テスト用)
func (g *Tonk) SetKnockerDeadwood(deadwood []*Card) { g.knockerDeadwood = deadwood }

// GetOpponentMelds 相手側のメルド取得
func (g *Tonk) GetOpponentMelds() [][]*Card { return g.opponentMelds }

// GetOpponentDeadwood 相手側のデッドウッド取得
func (g *Tonk) GetOpponentDeadwood() []*Card { return g.opponentDeadwood }

// GetIsTonk 配牌Tonkかどうか取得
func (g *Tonk) GetIsTonk() bool { return g.isTonk }

// SetIsTonk 配牌Tonk設定 (テスト用)
func (g *Tonk) SetIsTonk(isTonk bool) { g.isTonk = isTonk }

// GetIsUndercut アンダーカットかどうか取得
func (g *Tonk) GetIsUndercut() bool { return g.isUndercut }

// --- Private methods ---

// sortAllHands 全プレイヤーの手札をソートする
func (g *Tonk) sortAllHands() {
	for i := range g.players {
		g.sortHand(i)
	}
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func (g *Tonk) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// tonkJSON is the JSON wire format for Tonk.
type tonkJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*TonkPlayer     `json:"pl"`
	Config           TonkConfig        `json:"cf"`
	Phase            TonkPhase         `json:"ps"`
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
	IsTonk           bool              `json:"it"`
	IsUndercut       bool              `json:"iu"`
}

// MarshalJSON implements json.Marshaler.
func (g *Tonk) MarshalJSON() ([]byte, error) {
	return json.Marshal(tonkJSON{
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
		IsTonk:           g.isTonk,
		IsUndercut:       g.isUndercut,
	})
}

// tonkMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const tonkMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (g *Tonk) UnmarshalJSON(data []byte) error {
	var j tonkJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > tonkMaxSliceLen || len(j.DiscardPile) > tonkMaxSliceLen ||
		len(j.DrawPile) > tonkMaxSliceLen || len(j.ActionLog) > tonkMaxSliceLen ||
		len(j.KnockerMelds) > tonkMaxSliceLen || len(j.KnockerDeadwood) > tonkMaxSliceLen ||
		len(j.OpponentMelds) > tonkMaxSliceLen || len(j.OpponentDeadwood) > tonkMaxSliceLen {
		return fmt.Errorf("tonk: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*TonkPlayer, 0)
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
	g.isTonk = j.IsTonk
	g.isUndercut = j.IsUndercut
	// **復元したら必ず乱数源を張り直す。**Cloudflare Worker は毎リクエスト KV から
	// 組み直すので SetRand は一度も呼ばれない。rng を nil のままにすると、
	// シャッフル以外で rng を使う経路 (CPU の乱択など) が nil デリファレンスで
	// 落ちる。呼び出し側ごとにガードするのではなく、ここで構造的に潰す。
	g.rng = rand.New(rand.NewSource(rand.Int63()))
	return nil
}
