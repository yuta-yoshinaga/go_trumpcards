//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// MacauPlayerCnt マカオプレイヤー数
const MacauPlayerCnt = 4

// MacauHandSize 初期配布枚数
const MacauHandSize = 5

// MacauWildValue ワイルドカード値 (8: スート変更)
const MacauWildValue = 8

// MacauDrawTwoValue ドローツーのカード値 (2: 次のプレイヤーに2枚引かせる/重ね可)
const MacauDrawTwoValue = 2

// MacauSkipValue スキップのカード値 (7: 次のプレイヤーを飛ばす)
const MacauSkipValue = 7

// MacauReverseValue リバースのカード値 (J=11: プレイ方向を反転)
const MacauReverseValue = 11

// MacauDrawTwoAmount 2を1枚出すごとに累積するペナルティ枚数
const MacauDrawTwoAmount = 2

// MacauForgotPenalty マカオ宣言を忘れた場合のペナルティ枚数
const MacauForgotPenalty = 2

// MacauPhase ゲームフェーズ
type MacauPhase int

// Macauのフェーズ定数
const (
	// MacauPhasePlay 通常プレイフェーズ
	MacauPhasePlay MacauPhase = 0
	// MacauPhaseChooseSuit 8を出した後のスート選択フェーズ
	MacauPhaseChooseSuit MacauPhase = 1
	// MacauPhaseMustDeclare 手札が1枚になった後の宣言待ちフェーズ
	MacauPhaseMustDeclare MacauPhase = 2
	// MacauPhaseRoundEnd ラウンド終了フェーズ
	MacauPhaseRoundEnd MacauPhase = 3
	// MacauPhaseGameEnd ゲーム終了フェーズ
	MacauPhaseGameEnd MacauPhase = 4
)

// Macau マカオゲームクラス (クレイジーエイト + マジックカード + マカオ宣言)
type Macau struct {
	trumpCards       *TrumpCards
	players          []*MacauPlayer
	config           MacauConfig
	phase            MacauPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	chosenSuit       int // -1 = 未選択
	penaltyDrawCount int // 累積ドローツー枚数 (0 = ペナルティなし)
	direction        int // +1 = 時計回り, -1 = 反時計回り
	pendingSkip      bool
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	actionLogBase
}

// NewMacau コンストラクタ
func NewMacau(trumpCards *TrumpCards, players []*MacauPlayer, config MacauConfig) *Macau {
	return &Macau{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
		chosenSuit:  -1,
		direction:   1,
	}
}

// NewDefaultMacau returns Macau with the standard 4-player setup (1 human, 3 CPU)
// and DefaultMacauConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultMacau() *Macau {
	players := []*MacauPlayer{
		NewMacauPlayer(true),
		NewMacauPlayer(false),
		NewMacauPlayer(false),
		NewMacauPlayer(false),
	}
	return NewMacau(NewTrumpCards(0), players, DefaultMacauConfig())
}

// Reset ゲーム初期化
func (g *Macau) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.chosenSuit = -1
	g.penaltyDrawCount = 0
	g.direction = 1
	g.pendingSkip = false
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0
	g.actionLog = nil

	for _, p := range g.players {
		p.roundScore = 0
		p.cumulativeScore = 0
		p.Reset()
		p.SetIsFinished(false)
		p.SetHasDeclared(false)
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = MacauPhasePlay
}

// NextRound 次のラウンドを開始する
func (g *Macau) NextRound() {
	if g.phase != MacauPhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.chosenSuit = -1
	g.penaltyDrawCount = 0
	g.direction = 1
	g.pendingSkip = false
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = MacauPhasePlay
}

// dealInitialCards 初期配布: 各プレイヤーに5枚、1枚を捨て札に
func (g *Macau) dealInitialCards() {
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

	for i := 0; i < MacauHandSize; i++ {
		for j := 0; j < MacauPlayerCnt; j++ {
			if len(g.drawPile) > 0 {
				card := g.drawPile[len(g.drawPile)-1]
				g.drawPile = g.drawPile[:len(g.drawPile)-1]
				g.players[j].AddCard(card)
			}
		}
	}

	// 最初の1枚を捨て札に (特殊カードでも初期の効果は発動しない)
	if len(g.drawPile) > 0 {
		firstCard := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.discardPile = append(g.discardPile, firstCard)
	}
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (g *Macau) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MacauPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if !g.isValidPlay(card) {
		return NewDomainError(ErrInvalidPlay, "そのカードは出せません")
	}

	played := player.RemoveCard(cardIndex)
	g.playCard(g.currentPlayerIdx, played)
	return nil
}

// PlayerChooseSuit 人間プレイヤーがスートを選択する (8を出した後)
func (g *Macau) PlayerChooseSuit(suit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MacauPhaseChooseSuit {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return NewDomainError(ErrInvalidPlay, "スートは1〜4で指定してください")
	}

	g.chosenSuit = suit
	g.appendLog(g.currentPlayerIdx, "choose_suit", fmt.Sprintf("%s chooses %s", playerName(g.players, g.currentPlayerIdx), suitName(suit)), nil)

	g.finishTurn(g.currentPlayerIdx)
	return nil
}

// PlayerDraw 人間プレイヤーがカードを引く (ペナルティ中はスタックを引き受ける)
func (g *Macau) PlayerDraw() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MacauPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	return g.drawCard(g.currentPlayerIdx)
}

// PlayerDeclare 人間プレイヤーが「マカオ！」と宣言する
func (g *Macau) PlayerDeclare() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MacauPhaseMustDeclare {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.doDeclare(g.currentPlayerIdx)
	return nil
}

// PlayerSkipDeclare 人間プレイヤーが宣言をスキップする（ペナルティを受ける）
func (g *Macau) PlayerSkipDeclare() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MacauPhaseMustDeclare {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.applyDeclarePenalty(g.currentPlayerIdx)
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (g *Macau) CpuPlay() {
	if g.gameEndFlag || g.phase != MacauPhasePlay {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	cardIdx := g.cpuSelectPlayCard(g.currentPlayerIdx)
	if cardIdx >= 0 {
		player := g.players[g.currentPlayerIdx]
		played := player.RemoveCard(cardIdx)
		// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
		// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
		// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
		if played == nil {
			return
		}
		g.playCard(g.currentPlayerIdx, played)
	} else {
		// drawCard always returns nil today; the error is ignored intentionally.
		_ = g.drawCard(g.currentPlayerIdx)
	}
}

// CpuChooseSuit CPUがスートを選択する
func (g *Macau) CpuChooseSuit() {
	if g.gameEndFlag || g.phase != MacauPhaseChooseSuit {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	suit := g.cpuSelectSuit(g.currentPlayerIdx)
	g.chosenSuit = suit
	g.appendLog(g.currentPlayerIdx, "choose_suit", fmt.Sprintf("%s chooses %s", playerName(g.players, g.currentPlayerIdx), suitName(suit)), nil)
	g.finishTurn(g.currentPlayerIdx)
}

// CpuDeclare CPUが自動的に「マカオ！」と宣言する
func (g *Macau) CpuDeclare() {
	if g.gameEndFlag || g.phase != MacauPhaseMustDeclare {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	// 難易度 Easy では一定確率で宣言を忘れる
	if g.config.CpuDifficulty == MacauCpuDifficultyEasy && rand.Intn(4) == 0 {
		g.applyDeclarePenalty(g.currentPlayerIdx)
		return
	}
	g.doDeclare(g.currentPlayerIdx)
}

// doDeclare 宣言処理の共通実装
func (g *Macau) doDeclare(playerIdx int) {
	g.players[playerIdx].SetHasDeclared(true)
	g.appendLog(playerIdx, "declare", fmt.Sprintf("%s declares Macau!", playerName(g.players, playerIdx)), nil)
	g.advanceTurn()
}

// applyDeclarePenalty 宣言忘れペナルティとして規定枚数を引かせる
func (g *Macau) applyDeclarePenalty(playerIdx int) {
	drawn := g.drawCards(playerIdx, MacauForgotPenalty)
	g.appendLog(playerIdx, "penalty", fmt.Sprintf("%s forgot to declare Macau! (+%d cards)", playerName(g.players, playerIdx), drawn), nil)
	g.sortHand(playerIdx)
	g.advanceTurn()
}

// ScoreRound ラウンドのスコアを確定する
func (g *Macau) ScoreRound() {
	if g.phase != MacauPhaseRoundEnd {
		return
	}

	winnerIdx := -1
	for i, p := range g.players {
		if p.GetCardsSize() == 0 {
			winnerIdx = i
			break
		}
	}

	if winnerIdx < 0 {
		return
	}

	totalScore := 0
	for i, p := range g.players {
		if i == winnerIdx {
			continue
		}
		score := 0
		for j := 0; j < p.GetCardsSize(); j++ {
			score += crazyEightsCardScore(p.GetCard(j))
		}
		totalScore += score
		g.appendLog(i, "hand_score", fmt.Sprintf("%s: %d points remaining", playerName(g.players, i), score), nil)
	}

	g.players[winnerIdx].roundScore = totalScore
	g.appendLog(winnerIdx, "round_win", fmt.Sprintf("%s wins round %d (+%d points)", playerName(g.players, winnerIdx), g.roundNumber, totalScore), nil)

	g.players[winnerIdx].CommitRoundScore()

	g.checkGameEnd()
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Macau) GetPhase() MacauPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Macau) SetPhase(phase MacauPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Macau) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Macau) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Macau) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Macau) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *Macau) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *Macau) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDiscardTop 捨て札の一番上を取得
func (g *Macau) GetDiscardTop() *Card {
	if len(g.discardPile) == 0 {
		return nil
	}
	return g.discardPile[len(g.discardPile)-1]
}

// GetDrawPileCount 山札の残り枚数取得
func (g *Macau) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定 (テスト用)
func (g *Macau) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetChosenSuit 選択されたスート取得 (-1 = 未選択)
func (g *Macau) GetChosenSuit() int { return g.chosenSuit }

// SetChosenSuit スート設定 (テスト用)
func (g *Macau) SetChosenSuit(suit int) { g.chosenSuit = suit }

// GetPenaltyDrawCount 累積ドローツー枚数取得 (0 = ペナルティなし)
func (g *Macau) GetPenaltyDrawCount() int { return g.penaltyDrawCount }

// SetPenaltyDrawCount 累積ドローツー枚数設定 (テスト用)
func (g *Macau) SetPenaltyDrawCount(n int) { g.penaltyDrawCount = n }

// GetDirection プレイ方向取得 (+1 = 時計回り, -1 = 反時計回り)
func (g *Macau) GetDirection() int { return g.direction }

// SetDirection プレイ方向設定 (テスト用)
func (g *Macau) SetDirection(d int) { g.direction = d }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Macau) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (g *Macau) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Macau) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Macau) GetPlayer(i int) *MacauPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Macau) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Macau) GetConfig() MacauConfig { return g.config }

// SetConfig 設定変更
func (g *Macau) SetConfig(cfg MacauConfig) { g.config = cfg }

// --- Private methods ---

// IsValidPlay は現在の場状態 (捨て札トップ・宣言スート・ペナルティ連鎖) を踏まえ
// カードがプレイ可能かを返す (公開版; CUI ヒントが利用する)。
func (g *Macau) IsValidPlay(card *Card) bool { return g.isValidPlay(card) }

// isValidPlay カードがプレイ可能か判定
func (g *Macau) isValidPlay(card *Card) bool {
	// ペナルティ中は2のみ重ねられる
	if g.penaltyDrawCount > 0 {
		return card.GetValue() == MacauDrawTwoValue
	}

	// 8はいつでも出せる
	if card.GetValue() == MacauWildValue {
		return true
	}

	top := g.GetDiscardTop()
	if top == nil {
		return true
	}

	// chosenSuit が設定されている場合 (前の人が8を出した)
	if g.chosenSuit > 0 {
		return card.GetDesign() == g.chosenSuit
	}

	// スートまたはランクが一致
	return card.GetDesign() == top.GetDesign() || card.GetValue() == top.GetValue()
}

// playCard カードをプレイする共通処理
func (g *Macau) playCard(playerIdx int, card *Card) {
	g.discardPile = append(g.discardPile, card)
	g.chosenSuit = -1

	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	// マジックカードの状態更新
	switch card.GetValue() {
	case MacauDrawTwoValue:
		g.penaltyDrawCount += MacauDrawTwoAmount
		g.appendLog(playerIdx, "draw_two", fmt.Sprintf("Draw stack is now %d", g.penaltyDrawCount), nil)
	case MacauSkipValue:
		g.pendingSkip = true
		g.appendLog(playerIdx, "skip", "Next player is skipped", nil)
	case MacauReverseValue:
		g.direction = -g.direction
		g.appendLog(playerIdx, "reverse", "Play direction reversed", nil)
	}

	// 手札が空になったらラウンド終了
	if g.players[playerIdx].GetCardsSize() == 0 {
		g.players[playerIdx].SetIsFinished(true)
		g.phase = MacauPhaseRoundEnd
		return
	}

	// 8を出した場合はスート選択フェーズへ
	if card.GetValue() == MacauWildValue {
		g.phase = MacauPhaseChooseSuit
		return
	}

	g.finishTurn(playerIdx)
}

// finishTurn 宣言チェックを行い、必要なら宣言フェーズへ、そうでなければ次の手番へ
func (g *Macau) finishTurn(playerIdx int) {
	if g.players[playerIdx].GetCardsSize() == 1 && !g.players[playerIdx].GetHasDeclared() {
		g.phase = MacauPhaseMustDeclare
		return
	}
	g.advanceTurn()
}

// advanceTurn 次のプレイヤーへ (方向・スキップを反映)
func (g *Macau) advanceTurn() {
	steps := 1
	if g.pendingSkip {
		steps = 2
		g.pendingSkip = false
	}
	g.currentPlayerIdx = g.wrapIdx(g.currentPlayerIdx + steps*g.direction)
	g.phase = MacauPhasePlay
	// 手札が2枚以上に戻った場合は宣言フラグをリセット
	for _, p := range g.players {
		if p.GetCardsSize() >= 2 {
			p.SetHasDeclared(false)
		}
	}
}

// wrapIdx プレイヤーインデックスを 0..len(players)-1 に正規化する (負の方向にも対応)
func (g *Macau) wrapIdx(i int) int {
	n := len(g.players)
	if n == 0 {
		return 0
	}
	return ((i % n) + n) % n
}

// drawCard カードを引く (ペナルティ中はスタックを引き受けて手番終了)
func (g *Macau) drawCard(playerIdx int) error {
	if g.penaltyDrawCount > 0 {
		drawn := g.drawCards(playerIdx, g.penaltyDrawCount)
		g.penaltyDrawCount = 0
		g.appendLog(playerIdx, "take_penalty", fmt.Sprintf("%s takes %d penalty cards", playerName(g.players, playerIdx), drawn), nil)
		g.sortHand(playerIdx)
		g.advanceTurn()
		return nil
	}

	if len(g.drawPile) == 0 {
		g.recycleDrawPile()
	}

	if len(g.drawPile) == 0 {
		// 引けるカードがない→パス
		g.appendLog(playerIdx, "pass", fmt.Sprintf("%s passes (no cards to draw)", playerName(g.players, playerIdx)), nil)
		g.advanceTurn()
		return nil
	}

	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[playerIdx].AddCard(card)
	g.sortHand(playerIdx)

	g.appendLog(playerIdx, "draw", fmt.Sprintf("%s draws a card", playerName(g.players, playerIdx)), nil)

	// 引いたカードが出せないなら次へ
	if !g.hasPlayableCard(playerIdx) {
		g.advanceTurn()
	}

	return nil
}

// drawCards 指定枚数を引く (山札が尽きたら捨て札を再利用)。実際に引けた枚数を返す。
func (g *Macau) drawCards(playerIdx, n int) int {
	drawn := 0
	for i := 0; i < n; i++ {
		if len(g.drawPile) == 0 {
			g.recycleDrawPile()
		}
		if len(g.drawPile) == 0 {
			break
		}
		card := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.players[playerIdx].AddCard(card)
		drawn++
	}
	return drawn
}

// recycleDrawPile 捨て札から山札を再構築する
func (g *Macau) recycleDrawPile() {
	if len(g.discardPile) <= 1 {
		return
	}

	top := g.discardPile[len(g.discardPile)-1]
	recycled := g.discardPile[:len(g.discardPile)-1]
	g.discardPile = []*Card{top}

	rand.Shuffle(len(recycled), func(i, j int) {
		recycled[i], recycled[j] = recycled[j], recycled[i]
	})

	g.drawPile = recycled
}

// hasPlayableCard プレイヤーが出せるカードを持っているか
func (g *Macau) hasPlayableCard(playerIdx int) bool {
	player := g.players[playerIdx]
	for i := 0; i < player.GetCardsSize(); i++ {
		if g.isValidPlay(player.GetCard(i)) {
			return true
		}
	}
	return false
}

// checkGameEnd ゲーム終了判定
func (g *Macau) checkGameEnd() {
	hasWinner := false
	for i := 0; i < MacauPlayerCnt; i++ {
		if g.players[i].cumulativeScore >= g.config.PointLimit {
			hasWinner = true
			break
		}
	}

	if !hasWinner {
		return
	}

	g.gameEndFlag = true
	g.phase = MacauPhaseGameEnd

	// 最高スコアのプレイヤーが勝者
	maxScore := g.players[0].cumulativeScore
	g.winnerIdx = 0
	for i := 1; i < MacauPlayerCnt; i++ {
		if g.players[i].cumulativeScore > maxScore {
			maxScore = g.players[i].cumulativeScore
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(g.players, g.winnerIdx)), nil)
}

// sortAllHands 全プレイヤーの手札をソートする
func (g *Macau) sortAllHands() {
	for i := range g.players {
		g.sortHand(i)
	}
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func (g *Macau) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// --- CPU AI ---

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する (-1 = プレイ不可)
func (g *Macau) cpuSelectPlayCard(playerIdx int) int {
	validIndices := g.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return -1
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch g.config.CpuDifficulty {
	case MacauCpuDifficultyHard:
		return g.cpuPlayHard(playerIdx, validIndices)
	case MacauCpuDifficultyNormal:
		return g.cpuPlayNormal(playerIdx, validIndices)
	default:
		return g.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (g *Macau) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal 最も多いスートを優先、8を温存
func (g *Macau) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := g.players[playerIdx]

	nonWild := make([]int, 0)
	for _, idx := range validIndices {
		if player.GetCard(idx).GetValue() != MacauWildValue {
			nonWild = append(nonWild, idx)
		}
	}

	candidates := validIndices
	if len(nonWild) > 0 {
		candidates = nonWild
	}

	suitCount := g.countSuits(playerIdx)
	bestIdx := candidates[0]
	bestCount := suitCount[player.GetCard(candidates[0]).GetDesign()]
	for _, idx := range candidates[1:] {
		sc := suitCount[player.GetCard(idx).GetDesign()]
		if sc > bestCount {
			bestCount = sc
			bestIdx = idx
		}
	}
	return bestIdx
}

// cpuPlayHard 戦略的プレイ: 8を温存しつつ高得点カードを優先消費する
func (g *Macau) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := g.players[playerIdx]

	nonWild := make([]int, 0)
	for _, idx := range validIndices {
		if player.GetCard(idx).GetValue() != MacauWildValue {
			nonWild = append(nonWild, idx)
		}
	}

	// 手札が2枚以下なら8を使ってもOK
	if player.GetCardsSize() <= 2 {
		nonWild = validIndices
	}

	candidates := validIndices
	if len(nonWild) > 0 {
		candidates = nonWild
	}

	suitCount := g.countSuits(playerIdx)
	bestIdx := candidates[0]
	bestScore := g.cpuCardPriority(player.GetCard(candidates[0]), suitCount)
	for _, idx := range candidates[1:] {
		score := g.cpuCardPriority(player.GetCard(idx), suitCount)
		if score > bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// cpuCardPriority カードの優先度スコアを計算 (大きいほど先に出したい)
func (g *Macau) cpuCardPriority(card *Card, suitCount map[int]int) int {
	score := crazyEightsCardScore(card)  // 高得点カードを優先消費
	score += suitCount[card.GetDesign()] // 多いスートを優先
	return score
}

// cpuSelectSuit CPUがスートを選択する
func (g *Macau) cpuSelectSuit(playerIdx int) int {
	switch g.config.CpuDifficulty {
	case MacauCpuDifficultyHard, MacauCpuDifficultyNormal:
		return g.cpuSelectSuitSmart(playerIdx)
	default:
		return g.cpuSelectSuitRandom()
	}
}

// cpuSelectSuitRandom ランダムにスートを選択
func (g *Macau) cpuSelectSuitRandom() int {
	return rand.Intn(4) + 1 // 1-4 (Spade, Clover, Heart, Diamond)
}

// cpuSelectSuitSmart 手札で最も多いスートを選択
func (g *Macau) cpuSelectSuitSmart(playerIdx int) int {
	suitCount := g.countSuits(playerIdx)

	bestSuit := CardDesignSpade
	bestCount := 0
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if suitCount[suit] > bestCount {
			bestCount = suitCount[suit]
			bestSuit = suit
		}
	}
	return bestSuit
}

// countSuits プレイヤーの手札のスート別枚数をカウント (8は除外)
func (g *Macau) countSuits(playerIdx int) map[int]int {
	player := g.players[playerIdx]
	counts := make(map[int]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetValue() != MacauWildValue {
			counts[card.GetDesign()]++
		}
	}
	return counts
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (g *Macau) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.isValidPlay(player.GetCard(i))
	})
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (g *Macau) GetValidPlayIndices(playerIdx int) []int {
	return g.getValidPlayIndices(playerIdx)
}

// macauJSON is the JSON wire format for Macau.
type macauJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*MacauPlayer    `json:"pl"`
	Config           MacauConfig       `json:"cf"`
	Phase            MacauPhase        `json:"ps"`
	CurrentPlayerIdx int               `json:"ci"`
	DiscardPile      []*Card           `json:"dp"`
	DrawPile         []*Card           `json:"wp"`
	ChosenSuit       int               `json:"cs"`
	PenaltyDrawCount int               `json:"pd"`
	Direction        int               `json:"dr"`
	PendingSkip      bool              `json:"sk"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	RoundNumber      int               `json:"rn"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Macau) MarshalJSON() ([]byte, error) {
	return json.Marshal(macauJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		ChosenSuit:       g.chosenSuit,
		PenaltyDrawCount: g.penaltyDrawCount,
		Direction:        g.direction,
		PendingSkip:      g.pendingSkip,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		ActionLog:        g.actionLog,
	})
}

// macauMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const macauMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (g *Macau) UnmarshalJSON(data []byte) error {
	var j macauJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > macauMaxSliceLen || len(j.DiscardPile) > macauMaxSliceLen ||
		len(j.DrawPile) > macauMaxSliceLen || len(j.ActionLog) > macauMaxSliceLen {
		return fmt.Errorf("macau: input array exceeds maximum allowed size")
	}
	// Macau is strictly a 4-player game; reject malformed states that would
	// otherwise cause out-of-bounds panics during play (many methods iterate
	// with the MacauPlayerCnt bound or index g.players[g.currentPlayerIdx]).
	if len(j.Players) != MacauPlayerCnt {
		return fmt.Errorf("macau: invalid player count: expected %d, got %d", MacauPlayerCnt, len(j.Players))
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= MacauPlayerCnt {
		return fmt.Errorf("macau: currentPlayerIdx %d out of range [0, %d)", j.CurrentPlayerIdx, MacauPlayerCnt)
	}
	if j.Phase < MacauPhasePlay || j.Phase > MacauPhaseGameEnd {
		return fmt.Errorf("macau: invalid phase: %d", j.Phase)
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*MacauPlayer, 0)
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
	g.chosenSuit = j.ChosenSuit
	g.penaltyDrawCount = j.PenaltyDrawCount
	if g.penaltyDrawCount < 0 {
		g.penaltyDrawCount = 0
	} else if g.penaltyDrawCount > macauMaxSliceLen {
		g.penaltyDrawCount = macauMaxSliceLen
	}
	g.direction = j.Direction
	if g.direction != 1 && g.direction != -1 {
		g.direction = 1
	}
	g.pendingSkip = j.PendingSkip
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.roundNumber = j.RoundNumber
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
