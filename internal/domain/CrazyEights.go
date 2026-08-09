package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// CrazyEightsPlayerCnt クレイジーエイトプレイヤー数
const CrazyEightsPlayerCnt = 4

// CrazyEightsHandSize 初期配布枚数
const CrazyEightsHandSize = 5

// CrazyEightsWildValue ワイルドカード値 (8)
const CrazyEightsWildValue = 8

// CrazyEightsPhase ゲームフェーズ
type CrazyEightsPhase int

// CrazyEightsのフェーズ定数
const (
	// CrazyEightsPhasePlay 通常プレイフェーズ
	CrazyEightsPhasePlay CrazyEightsPhase = 0
	// CrazyEightsPhaseChooseSuit 8を出した後のスート選択フェーズ
	CrazyEightsPhaseChooseSuit CrazyEightsPhase = 1
	// CrazyEightsPhaseRoundEnd ラウンド終了フェーズ
	CrazyEightsPhaseRoundEnd CrazyEightsPhase = 2
	// CrazyEightsPhaseGameEnd ゲーム終了フェーズ
	CrazyEightsPhaseGameEnd CrazyEightsPhase = 3
)

// CrazyEights クレイジーエイトゲームクラス
type CrazyEights struct {
	trumpCards       *TrumpCards
	players          []*CrazyEightsPlayer
	config           CrazyEightsConfig
	phase            CrazyEightsPhase
	currentPlayerIdx int
	discardPile      []*Card
	drawPile         []*Card
	chosenSuit       int // -1 = 未選択
	gameEndFlag      bool
	winnerIdx        int
	roundNumber      int
	actionLogBase
}

// NewCrazyEights コンストラクタ
func NewCrazyEights(trumpCards *TrumpCards, players []*CrazyEightsPlayer, config CrazyEightsConfig) *CrazyEights {
	return &CrazyEights{
		trumpCards:  trumpCards,
		players:     players,
		config:      config,
		winnerIdx:   -1,
		roundNumber: 0,
		chosenSuit:  -1,
	}
}

// NewDefaultCrazyEights returns CrazyEights with the standard 4-player setup (1 human, 3 CPU)
// and DefaultCrazyEightsConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultCrazyEights() *CrazyEights {
	players := []*CrazyEightsPlayer{
		NewCrazyEightsPlayer(true),
		NewCrazyEightsPlayer(false),
		NewCrazyEightsPlayer(false),
		NewCrazyEightsPlayer(false),
	}
	return NewCrazyEights(NewTrumpCards(0), players, DefaultCrazyEightsConfig())
}

// Reset ゲーム初期化
func (g *CrazyEights) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.chosenSuit = -1
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0
	g.actionLog = nil

	for _, p := range g.players {
		p.roundScore = 0
		p.cumulativeScore = 0
		p.Reset()
		p.SetIsFinished(false)
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = CrazyEightsPhasePlay
}

// NextRound 次のラウンドを開始する
func (g *CrazyEights) NextRound() {
	if g.phase != CrazyEightsPhaseRoundEnd {
		return
	}

	g.roundNumber++
	g.chosenSuit = -1
	g.discardPile = nil
	g.drawPile = nil
	g.currentPlayerIdx = 0

	for _, p := range g.players {
		p.ResetRound()
	}

	g.trumpCards.Shuffle()
	g.dealInitialCards()
	g.sortAllHands()

	g.phase = CrazyEightsPhasePlay
}

// dealInitialCards 初期配布: 各プレイヤーに5枚、1枚を捨て札に
func (g *CrazyEights) dealInitialCards() {
	// 全カードを drawPile に
	g.drawPile = make([]*Card, 0, g.trumpCards.GetTotalCount())
	for {
		card := g.trumpCards.DrawCard()
		if card == nil {
			break
		}
		g.drawPile = append(g.drawPile, card)
	}

	// シャッフル
	rand.Shuffle(len(g.drawPile), func(i, j int) {
		g.drawPile[i], g.drawPile[j] = g.drawPile[j], g.drawPile[i]
	})

	// 各プレイヤーに配布
	for i := 0; i < CrazyEightsHandSize; i++ {
		for j := 0; j < CrazyEightsPlayerCnt; j++ {
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

// PlayerPlay 人間プレイヤーがカードをプレイする
func (g *CrazyEights) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CrazyEightsPhasePlay {
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
func (g *CrazyEights) PlayerChooseSuit(suit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CrazyEightsPhaseChooseSuit {
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

	g.advanceTurn()
	return nil
}

// PlayerDraw 人間プレイヤーがカードを引く
func (g *CrazyEights) PlayerDraw() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CrazyEightsPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	return g.drawCard(g.currentPlayerIdx)
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (g *CrazyEights) CpuPlay() {
	if g.gameEndFlag || g.phase != CrazyEightsPhasePlay {
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
		// ドロー
		_ = g.drawCard(g.currentPlayerIdx)
	}
}

// CpuChooseSuit CPUがスートを選択する
func (g *CrazyEights) CpuChooseSuit() {
	if g.gameEndFlag || g.phase != CrazyEightsPhaseChooseSuit {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}

	suit := g.cpuSelectSuit(g.currentPlayerIdx)
	g.chosenSuit = suit
	g.appendLog(g.currentPlayerIdx, "choose_suit", fmt.Sprintf("%s chooses %s", playerName(g.players, g.currentPlayerIdx), suitName(suit)), nil)
	g.advanceTurn()
}

// ScoreRound ラウンドのスコアを確定する
func (g *CrazyEights) ScoreRound() {
	if g.phase != CrazyEightsPhaseRoundEnd {
		return
	}

	// ラウンド勝者を見つける (手札が空のプレイヤー)
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

	// 他のプレイヤーの残りカードをスコアリング
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

	// 累積スコアに加算
	g.players[winnerIdx].CommitRoundScore()

	// ゲーム終了判定
	g.checkGameEnd()
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
// CrazyEightsHint はサーバーが計算した推奨手。
type CrazyEightsHint struct {
	// CardIndex は推奨する手札の位置。スート選択フェーズでは nil。
	CardIndex *int
	// Suit は 8 を出した後に指名すべきスート。プレイフェーズでは nil。
	Suit *int
	// Reason は理由キー (i18n で引く)。
	Reason string
}

// GetHint は人間の手番での推奨手を返す。手番でない・出せる札が無いときは nil。
//
// **Hearts / Spades は HintOutput でサーバー計算の理由付きヒントを返すのに、
// CrazyEights にはドメインの GetHint すら無く、全ゲーム共通の簡易ヒューリスティック
// (FrontendHintTooltip) しか支援が無かった。**推奨手は CPU の最善手選択
// (cpuPlayHard / cpuChooseSuit) をそのまま使う。別ロジックを書くと「CPU は選ばない
// 手を人間に勧める」ことになる。
func (g *CrazyEights) GetHint() *CrazyEightsHint {
	if g.gameEndFlag || g.currentPlayerIdx != 0 || !g.players[0].GetIsHuman() {
		return nil
	}

	if g.phase == CrazyEightsPhaseChooseSuit {
		suit := g.cpuSelectSuit(0)
		return &CrazyEightsHint{Suit: &suit, Reason: "choose_longest_suit"}
	}

	if g.phase != CrazyEightsPhasePlay {
		return nil
	}
	valid := g.getValidPlayIndices(0)
	if len(valid) == 0 {
		// 出せる札が無い = 引くしかない。推奨する札が無いのでヒントを出さない。
		return nil
	}
	idx := g.cpuPlayHard(0, valid)
	return &CrazyEightsHint{CardIndex: &idx, Reason: g.playHintReason(idx)}
}

// playHintReason は推奨札を選んだ理由キーを返す。
func (g *CrazyEights) playHintReason(idx int) string {
	card := g.players[0].GetCard(idx)
	if card == nil {
		return "play_valid"
	}
	if card.GetValue() == CrazyEightsWildValue {
		return "play_wild"
	}
	top := g.GetDiscardTop()
	if top != nil && card.GetValue() == top.GetValue() {
		return "match_rank"
	}
	return "match_suit"
}

func (g *CrazyEights) GetPhase() CrazyEightsPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *CrazyEights) SetPhase(phase CrazyEightsPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *CrazyEights) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *CrazyEights) SetRoundNumber(n int) { g.roundNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *CrazyEights) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *CrazyEights) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetDiscardPile 捨て札の山を取得
func (g *CrazyEights) GetDiscardPile() []*Card { return g.discardPile }

// SetDiscardPile 捨て札の山を設定 (テスト用)
func (g *CrazyEights) SetDiscardPile(pile []*Card) { g.discardPile = pile }

// GetDiscardTop 捨て札の一番上を取得
func (g *CrazyEights) GetDiscardTop() *Card {
	return discardTop(g.discardPile)
}

// GetDrawPileCount 山札の残り枚数取得
func (g *CrazyEights) GetDrawPileCount() int { return len(g.drawPile) }

// SetDrawPile 山札を設定 (テスト用)
func (g *CrazyEights) SetDrawPile(pile []*Card) { g.drawPile = pile }

// GetChosenSuit 選択されたスート取得 (-1 = 未選択)
func (g *CrazyEights) GetChosenSuit() int { return g.chosenSuit }

// SetChosenSuit スート設定 (テスト用)
func (g *CrazyEights) SetChosenSuit(suit int) { g.chosenSuit = suit }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *CrazyEights) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (g *CrazyEights) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *CrazyEights) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *CrazyEights) GetPlayer(i int) *CrazyEightsPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *CrazyEights) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// GetConfig 設定取得
func (g *CrazyEights) GetConfig() CrazyEightsConfig { return g.config }

// SetConfig 設定変更
func (g *CrazyEights) SetConfig(cfg CrazyEightsConfig) { g.config = cfg }

// --- Private methods ---

// isValidPlay カードがプレイ可能か判定
func (g *CrazyEights) isValidPlay(card *Card) bool {
	// 8はいつでも出せる
	if card.GetValue() == CrazyEightsWildValue {
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
func (g *CrazyEights) playCard(playerIdx int, card *Card) {
	g.discardPile = append(g.discardPile, card)
	g.chosenSuit = -1

	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	// 手札が空になったらラウンド終了
	if g.players[playerIdx].GetCardsSize() == 0 {
		g.players[playerIdx].SetIsFinished(true)
		g.phase = CrazyEightsPhaseRoundEnd
		return
	}

	// 8を出した場合はスート選択フェーズへ
	if card.GetValue() == CrazyEightsWildValue {
		g.phase = CrazyEightsPhaseChooseSuit
		return
	}

	g.advanceTurn()
}

// advanceTurn 次のプレイヤーへ
func (g *CrazyEights) advanceTurn() {
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % CrazyEightsPlayerCnt
	g.phase = CrazyEightsPhasePlay
}

// drawCard カードを引く
func (g *CrazyEights) drawCard(playerIdx int) error {
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

	// 引いたカードが出せるなら手番を保持 (プレイヤーが次に出す)
	// 出せないなら次へ
	if !g.hasPlayableCard(playerIdx) {
		g.advanceTurn()
	}

	return nil
}

// recycleDrawPile 捨て札から山札を再構築する
func (g *CrazyEights) recycleDrawPile() {
	if len(g.discardPile) <= 1 {
		return
	}

	// 一番上のカードを残して、残りを山札に
	top := g.discardPile[len(g.discardPile)-1]
	recycled := g.discardPile[:len(g.discardPile)-1]
	g.discardPile = []*Card{top}

	// シャッフル
	rand.Shuffle(len(recycled), func(i, j int) {
		recycled[i], recycled[j] = recycled[j], recycled[i]
	})

	g.drawPile = recycled
}

// hasPlayableCard プレイヤーが出せるカードを持っているか
func (g *CrazyEights) hasPlayableCard(playerIdx int) bool {
	return handHasAny(g.players[playerIdx], g.isValidPlay)
}

// checkGameEnd ゲーム終了判定
func (g *CrazyEights) checkGameEnd() {
	hasWinner := false
	for i := 0; i < CrazyEightsPlayerCnt; i++ {
		if g.players[i].cumulativeScore >= g.config.PointLimit {
			hasWinner = true
			break
		}
	}

	if !hasWinner {
		return
	}

	g.gameEndFlag = true
	g.phase = CrazyEightsPhaseGameEnd

	// 最高スコアのプレイヤーが勝者
	maxScore := g.players[0].cumulativeScore
	g.winnerIdx = 0
	for i := 1; i < CrazyEightsPlayerCnt; i++ {
		if g.players[i].cumulativeScore > maxScore {
			maxScore = g.players[i].cumulativeScore
			g.winnerIdx = i
		}
	}
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(g.players, g.winnerIdx)), nil)
}

// sortAllHands 全プレイヤーの手札をソートする
func (g *CrazyEights) sortAllHands() {
	sortHands(len(g.players), g)
}

// sortHand プレイヤーの手札をスート→値の順にソートする
func (g *CrazyEights) sortHand(playerIdx int) {
	sortPlayerHand(g.players[playerIdx], bySuitThenValue)
}

// --- CPU AI ---

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する (-1 = プレイ不可)
func (g *CrazyEights) cpuSelectPlayCard(playerIdx int) int {
	validIndices := g.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return -1
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch g.config.CpuDifficulty {
	case CrazyEightsCpuDifficultyHard:
		return g.cpuPlayHard(playerIdx, validIndices)
	case CrazyEightsCpuDifficultyNormal:
		return g.cpuPlayNormal(playerIdx, validIndices)
	default:
		return g.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (g *CrazyEights) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal 最も多いスートを優先、8を温存
func (g *CrazyEights) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := g.players[playerIdx]

	// 非8カードがあれば8を温存
	nonWild := make([]int, 0)
	for _, idx := range validIndices {
		if player.GetCard(idx).GetValue() != CrazyEightsWildValue {
			nonWild = append(nonWild, idx)
		}
	}

	candidates := validIndices
	if len(nonWild) > 0 {
		candidates = nonWild
	}

	// 最も多いスートのカードを優先
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

// cpuPlayHard 戦略的プレイ: スート支配、8を温存、相手の手札が少ないときに攻める
func (g *CrazyEights) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := g.players[playerIdx]

	// 非8カードがあれば8を温存
	nonWild := make([]int, 0)
	for _, idx := range validIndices {
		if player.GetCard(idx).GetValue() != CrazyEightsWildValue {
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

	// 高得点カードを優先的に消費 (リスク軽減)
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
func (g *CrazyEights) cpuCardPriority(card *Card, suitCount map[int]int) int {
	score := crazyEightsCardScore(card)  // 高得点カードを優先消費
	score += suitCount[card.GetDesign()] // 多いスートを優先
	return score
}

// cpuSelectSuit CPUがスートを選択する
func (g *CrazyEights) cpuSelectSuit(playerIdx int) int {
	switch g.config.CpuDifficulty {
	case CrazyEightsCpuDifficultyHard, CrazyEightsCpuDifficultyNormal:
		return g.cpuSelectSuitSmart(playerIdx)
	default:
		return g.cpuSelectSuitRandom()
	}
}

// cpuSelectSuitRandom ランダムにスートを選択
func (g *CrazyEights) cpuSelectSuitRandom() int {
	return rand.Intn(4) + 1 // 1-4 (Spade, Clover, Heart, Diamond)
}

// cpuSelectSuitSmart 手札で最も多いスートを選択
func (g *CrazyEights) cpuSelectSuitSmart(playerIdx int) int {
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
func (g *CrazyEights) countSuits(playerIdx int) map[int]int {
	player := g.players[playerIdx]
	counts := make(map[int]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		if card.GetValue() != CrazyEightsWildValue {
			counts[card.GetDesign()]++
		}
	}
	return counts
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (g *CrazyEights) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.isValidPlay(c) })
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (g *CrazyEights) GetValidPlayIndices(playerIdx int) []int {
	return g.getValidPlayIndices(playerIdx)
}

// crazyEightsCardScore カードのスコアを計算する
func crazyEightsCardScore(card *Card) int {
	v := card.GetValue()
	switch {
	case v == CrazyEightsWildValue:
		return 50
	case v == 1: // Ace
		return 1
	case v >= 11: // J, Q, K
		return 10
	default:
		return v
	}
}

// suitName スート名を返す
func suitName(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "♠"
	case CardDesignClover:
		return "♣"
	case CardDesignHeart:
		return "♥"
	case CardDesignDiamond:
		return "♦"
	default:
		return "?"
	}
}

// crazyEightsJSON is the JSON wire format for CrazyEights.
type crazyEightsJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*CrazyEightsPlayer `json:"pl"`
	Config           CrazyEightsConfig    `json:"cf"`
	Phase            CrazyEightsPhase     `json:"ps"`
	CurrentPlayerIdx int                  `json:"ci"`
	DiscardPile      []*Card              `json:"dp"`
	DrawPile         []*Card              `json:"wp"`
	ChosenSuit       int                  `json:"cs"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerIdx        int                  `json:"wi"`
	RoundNumber      int                  `json:"rn"`
	ActionLog        []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *CrazyEights) MarshalJSON() ([]byte, error) {
	return json.Marshal(crazyEightsJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CurrentPlayerIdx: g.currentPlayerIdx,
		DiscardPile:      g.discardPile,
		DrawPile:         g.drawPile,
		ChosenSuit:       g.chosenSuit,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundNumber:      g.roundNumber,
		ActionLog:        g.actionLog,
	})
}

// crazyEightsMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const crazyEightsMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (g *CrazyEights) UnmarshalJSON(data []byte) error {
	var j crazyEightsJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > crazyEightsMaxSliceLen || len(j.DiscardPile) > crazyEightsMaxSliceLen ||
		len(j.DrawPile) > crazyEightsMaxSliceLen || len(j.ActionLog) > crazyEightsMaxSliceLen {
		return fmt.Errorf("crazyeights: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*CrazyEightsPlayer, 0)
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
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.roundNumber = j.RoundNumber
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
