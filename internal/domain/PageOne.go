package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// PageOnePlayerCnt ページワンプレイヤー数
const PageOnePlayerCnt = 4

// PageOneHandSize 初期配布枚数
const PageOneHandSize = 5

// PageOnePenaltyDraw 宣言を忘れた場合のペナルティ枚数
const PageOnePenaltyDraw = 2

// PageOnePhase ゲームフェーズ
type PageOnePhase int

// PageOneのフェーズ定数
const (
	// PageOnePhasePlay 通常プレイフェーズ
	PageOnePhasePlay PageOnePhase = 0
	// PageOnePhaseMustDeclare 手札が1枚になった後の宣言待ちフェーズ
	PageOnePhaseMustDeclare PageOnePhase = 1
	// PageOnePhaseRoundEnd ラウンド終了フェーズ
	PageOnePhaseRoundEnd PageOnePhase = 2
	// PageOnePhaseGameEnd ゲーム終了フェーズ
	PageOnePhaseGameEnd PageOnePhase = 3
)

// PageOne ページワンゲームクラス
type PageOne struct {
	trumpCards       *TrumpCards
	players          []*PageOnePlayer
	config           PageOneConfig
	phase            PageOnePhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	actionLog        []*ActionLogEntry
}

// NewPageOne コンストラクタ
func NewPageOne(trumpCards *TrumpCards, players []*PageOnePlayer, config PageOneConfig) *PageOne {
	return &PageOne{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
	}
}

// NewDefaultPageOne returns PageOne with the standard 4-player setup (1 human, 3 CPU)
// and DefaultPageOneConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultPageOne() *PageOne {
	players := []*PageOnePlayer{
		NewPageOnePlayer(true),
		NewPageOnePlayer(false),
		NewPageOnePlayer(false),
		NewPageOnePlayer(false),
	}
	return NewPageOne(NewTrumpCards(0), players, DefaultPageOneConfig())
}

// Reset ゲーム初期化
func (g *PageOne) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
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

	g.phase = PageOnePhasePlay
}

// NextRound 次のラウンドを開始する
func (g *PageOne) NextRound() {
	if g.phase != PageOnePhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = PageOnePhasePlay
}

// dealInitialCards 初期配布: 各プレイヤーに5枚、1枚を捨て札に
func (g *PageOne) dealInitialCards() {
	g.drawPile = make([]*Card, 0, g.trumpCards.GetTotalCount())
	for {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.drawPile = append(g.drawPile, card)
	}

	for i := 0; i < PageOneHandSize; i++ {
		for j := 0; j < PageOnePlayerCnt; j++ {
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

// PlayerPlay 人間プレイヤーがカードをプレイする
func (g *PageOne) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != PageOnePhasePlay {
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

// PlayerDraw 人間プレイヤーがカードを引く
func (g *PageOne) PlayerDraw() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != PageOnePhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if g.hasPlayableCard(g.currentPlayerIdx) {
		return NewDomainError(ErrInvalidPlay, "出せるカードがあるときは引けません")
	}

	return g.drawCard(g.currentPlayerIdx)
}

// PlayerDeclare 人間プレイヤーが「ページワン！」と宣言する
func (g *PageOne) PlayerDeclare() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != PageOnePhaseMustDeclare {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.doDeclare(g.currentPlayerIdx)
	return nil
}

// PlayerSkipDeclare 人間プレイヤーが宣言をスキップする（ペナルティを受ける）
func (g *PageOne) PlayerSkipDeclare() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != PageOnePhaseMustDeclare {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.applyDeclarePenalty(g.currentPlayerIdx)
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (g *PageOne) CpuPlay() {
	if g.gameEndFlag || g.phase != PageOnePhasePlay {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	cardIdx := g.cpuSelectPlayCard(g.currentPlayerIdx)
	if cardIdx >= 0 {
		player := g.players[g.currentPlayerIdx]
		played := player.RemoveCard(cardIdx)
		g.playCard(g.currentPlayerIdx, played)
	} else {
		_ = g.drawCard(g.currentPlayerIdx)
	}
}

// CpuDeclare CPUが自動的に「ページワン！」と宣言する
func (g *PageOne) CpuDeclare() {
	if g.gameEndFlag || g.phase != PageOnePhaseMustDeclare {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	// 難易度 Easy では一定確率で宣言を忘れる
	if g.config.CpuDifficulty == PageOneCpuDifficultyEasy && rand.Intn(4) == 0 {
		g.applyDeclarePenalty(g.currentPlayerIdx)
		return
	}
	g.doDeclare(g.currentPlayerIdx)
}

// doDeclare 宣言処理の共通実装
func (g *PageOne) doDeclare(playerIdx int) {
	g.players[playerIdx].SetHasDeclared(true)
	g.appendLog(playerIdx, "declare", fmt.Sprintf("%s declares Page One!", g.playerName(playerIdx)), nil)
	g.advanceTurn()
}

// applyDeclarePenalty 宣言忘れペナルティとして2枚引かせる
func (g *PageOne) applyDeclarePenalty(playerIdx int) {
	g.appendLog(playerIdx, "penalty", fmt.Sprintf("%s forgot to declare Page One! (+%d cards)", g.playerName(playerIdx), PageOnePenaltyDraw), nil)
	for i := 0; i < PageOnePenaltyDraw; i++ {
		if len(g.drawPile) == 0 {
			g.recycleDrawPile()
		}
		if len(g.drawPile) == 0 {
			break
		}
		card := g.drawPile[len(g.drawPile)-1]
		g.drawPile = g.drawPile[:len(g.drawPile)-1]
		g.players[playerIdx].AddCard(card)
	}
	g.sortHand(playerIdx)
	g.advanceTurn()
}

// ScoreRound ラウンドのスコアを確定する
func (g *PageOne) ScoreRound() {
	if g.phase != PageOnePhaseRoundEnd {
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
			score += pageOneCardScore(p.GetCard(j))
		}
		totalScore += score
		g.appendLog(i, "hand_score", fmt.Sprintf("%s: %d points remaining", g.playerName(i), score), nil)
	}

	g.players[winnerIdx].roundScore = totalScore
	g.appendLog(winnerIdx, "round_win", fmt.Sprintf("%s wins round %d (+%d points)", g.playerName(winnerIdx), g.roundNumber, totalScore), nil)

	g.players[winnerIdx].CommitRoundScore()

	g.checkGameEnd()
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *PageOne) GetPhase() PageOnePhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *PageOne) SetPhase(phase PageOnePhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *PageOne) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *PageOne) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *PageOne) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *PageOne) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *PageOne) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *PageOne) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDiscardTop 捨て札の一番上を取得
func (g *PageOne) GetDiscardTop() *Card {
	if len(g.discardPile) == 0 {
		return nil
	}
	return g.discardPile[len(g.discardPile)-1]
}

// GetDrawPileCount 山札の残り枚数取得
func (g *PageOne) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定 (テスト用)
func (g *PageOne) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *PageOne) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (g *PageOne) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *PageOne) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *PageOne) GetPlayer(i int) *PageOnePlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *PageOne) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *PageOne) GetConfig() PageOneConfig { return g.config }

// SetConfig 設定変更
func (g *PageOne) SetConfig(cfg PageOneConfig) { g.config = cfg }

// GetActionLog 棋譜取得
func (g *PageOne) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (g *PageOne) GetValidPlayIndices(playerIdx int) []int {
	return g.getValidPlayIndices(playerIdx)
}

// --- Private methods ---

// isValidPlay カードがプレイ可能か判定 (スートまたはランクが一致)
// IsValidPlay は現在の場状態を踏まえカードがプレイ可能かを返す (公開版; CUI ヒントが利用)。
func (g *PageOne) IsValidPlay(card *Card) bool { return g.isValidPlay(card) }

func (g *PageOne) isValidPlay(card *Card) bool {
	top := g.GetDiscardTop()
	if top == nil {
		return true
	}
	return card.GetDesign() == top.GetDesign() || card.GetValue() == top.GetValue()
}

// playCard カードをプレイする共通処理
func (g *PageOne) playCard(playerIdx int, card *Card) {
	g.discardPile = append(g.discardPile, card)

	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})

	remaining := g.players[playerIdx].GetCardsSize()

	if remaining == 0 {
		g.players[playerIdx].SetIsFinished(true)
		g.phase = PageOnePhaseRoundEnd
		return
	}

	if remaining == 1 && !g.players[playerIdx].GetHasDeclared() {
		g.phase = PageOnePhaseMustDeclare
		return
	}

	g.advanceTurn()
}

// advanceTurn 次のプレイヤーへ
func (g *PageOne) advanceTurn() {
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % PageOnePlayerCnt
	g.phase = PageOnePhasePlay
	// 手札が2枚以上に戻った場合は宣言フラグをリセット
	for _, p := range g.players {
		if p.GetCardsSize() >= 2 {
			p.SetHasDeclared(false)
		}
	}
}

// drawCard カードを引く
func (g *PageOne) drawCard(playerIdx int) error {
	if len(g.drawPile) == 0 {
		g.recycleDrawPile()
	}

	if len(g.drawPile) == 0 {
		g.appendLog(playerIdx, "pass", fmt.Sprintf("%s passes (no cards to draw)", g.playerName(playerIdx)), nil)
		g.advanceTurn()
		return nil
	}

	card := g.drawPile[len(g.drawPile)-1]
	g.drawPile = g.drawPile[:len(g.drawPile)-1]
	g.players[playerIdx].AddCard(card)
	g.sortHand(playerIdx)

	g.appendLog(playerIdx, "draw", fmt.Sprintf("%s draws a card", g.playerName(playerIdx)), nil)

	if !g.hasPlayableCard(playerIdx) {
		g.advanceTurn()
	}

	return nil
}

// recycleDrawPile 捨て札から山札を再構築する
func (g *PageOne) recycleDrawPile() {
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
func (g *PageOne) hasPlayableCard(playerIdx int) bool {
	player := g.players[playerIdx]
	for i := 0; i < player.GetCardsSize(); i++ {
		if g.isValidPlay(player.GetCard(i)) {
			return true
		}
	}
	return false
}

// checkGameEnd ゲーム終了判定
func (g *PageOne) checkGameEnd() {
	hasWinner := false
	for i := 0; i < PageOnePlayerCnt; i++ {
		if g.players[i].cumulativeScore >= g.config.PointLimit {
			hasWinner = true
			break
		}
	}

	if !hasWinner {
		return
	}

	g.gameEndFlag = true
	g.phase = PageOnePhaseGameEnd

	maxScore := g.players[0].cumulativeScore
	g.winnerIdx = 0
	for i := 1; i < PageOnePlayerCnt; i++ {
		if g.players[i].cumulativeScore > maxScore {
			maxScore = g.players[i].cumulativeScore
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", g.playerName(g.winnerIdx)), nil)
}

// sortAllHands 全プレイヤーの手札をソートする
func (g *PageOne) sortAllHands() {
	for i := range g.players {
		g.sortHand(i)
	}
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func (g *PageOne) sortHand(playerIdx int) {
	p := g.players[playerIdx]
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// playerName プレイヤー名を返す
func (g *PageOne) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// appendLog 棋譜にエントリを追加する
func (g *PageOne) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- CPU AI ---

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する (-1 = プレイ不可)
func (g *PageOne) cpuSelectPlayCard(playerIdx int) int {
	validIndices := g.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return -1
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch g.config.CpuDifficulty {
	case PageOneCpuDifficultyHard:
		return g.cpuPlayHard(playerIdx, validIndices)
	case PageOneCpuDifficultyNormal:
		return g.cpuPlayNormal(playerIdx, validIndices)
	default:
		return validIndices[rand.Intn(len(validIndices))]
	}
}

// cpuPlayNormal 最も多いスートを優先
func (g *PageOne) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := g.players[playerIdx]
	suitCount := g.countSuits(playerIdx)
	bestIdx := validIndices[0]
	bestCount := suitCount[player.GetCard(validIndices[0]).GetDesign()]
	for _, idx := range validIndices[1:] {
		sc := suitCount[player.GetCard(idx).GetDesign()]
		if sc > bestCount {
			bestCount = sc
			bestIdx = idx
		}
	}
	return bestIdx
}

// cpuPlayHard 戦略的プレイ: 高得点カードを優先消費しつつ、多いスートを維持
func (g *PageOne) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := g.players[playerIdx]
	suitCount := g.countSuits(playerIdx)
	bestIdx := validIndices[0]
	bestScore := g.cpuCardPriority(player.GetCard(validIndices[0]), suitCount)
	for _, idx := range validIndices[1:] {
		score := g.cpuCardPriority(player.GetCard(idx), suitCount)
		if score > bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return bestIdx
}

// cpuCardPriority カードの優先度スコアを計算 (大きいほど先に出したい)
func (g *PageOne) cpuCardPriority(card *Card, suitCount map[int]int) int {
	score := pageOneCardScore(card)
	score += suitCount[card.GetDesign()]
	return score
}

// countSuits プレイヤーの手札のスート別枚数をカウント
func (g *PageOne) countSuits(playerIdx int) map[int]int {
	player := g.players[playerIdx]
	counts := make(map[int]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		counts[player.GetCard(i).GetDesign()]++
	}
	return counts
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (g *PageOne) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.isValidPlay(player.GetCard(i))
	})
}

// pageOneCardScore カードのスコアを計算する
func pageOneCardScore(card *Card) int {
	v := card.GetValue()
	switch {
	case v == 1:
		return 1
	case v >= 11:
		return 10
	default:
		return v
	}
}

// pageOneJSON is the JSON wire format for PageOne.
type pageOneJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*PageOnePlayer  `json:"pl"`
	Config           PageOneConfig     `json:"cf"`
	Phase            PageOnePhase      `json:"ps"`
	CurrentPlayerIdx int               `json:"ci"`
	DiscardPile      []*Card           `json:"dp"`
	DrawPile         []*Card           `json:"wp"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	RoundNumber      int               `json:"rn"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *PageOne) MarshalJSON() ([]byte, error) {
	return json.Marshal(pageOneJSON{
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
	})
}

// pageOneMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const pageOneMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (g *PageOne) UnmarshalJSON(data []byte) error {
	var j pageOneJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > pageOneMaxSliceLen || len(j.DiscardPile) > pageOneMaxSliceLen ||
		len(j.DrawPile) > pageOneMaxSliceLen || len(j.ActionLog) > pageOneMaxSliceLen {
		return fmt.Errorf("pageone: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*PageOnePlayer, 0)
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
	return nil
}
