package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

// FiftyOnePlayerCnt プレイヤー数
const FiftyOnePlayerCnt = 4

// FiftyOneHandSize 手札枚数
const FiftyOneHandSize = 5

// FiftyOneTableSize 場札枚数
const FiftyOneTableSize = 5

// fiftyOneShuffleCount シャッフル回数
const fiftyOneShuffleCount = 10

// FiftyOnePhase ゲームフェーズ
type FiftyOnePhase int

// FiftyOnePhase定数
const (
	// FiftyOnePhasePlay プレイ中
	FiftyOnePhasePlay FiftyOnePhase = 0
	// FiftyOnePhaseGameEnd ゲーム終了
	FiftyOnePhaseGameEnd FiftyOnePhase = 1
)

// CPU ストップ閾値
const (
	fiftyOneCpuStopThresholdEasy   = 35
	fiftyOneCpuStopThresholdNormal = 40
	fiftyOneCpuStopThresholdHard   = 45
)

// エラー定義
var (
	// ErrFiftyOneAlreadyStopped 既にストップ宣言済み
	ErrFiftyOneAlreadyStopped = errors.New("stop already called")
	// ErrFiftyOneNotCpuTurn CPU のターンではない
	ErrFiftyOneNotCpuTurn = errors.New("not CPU's turn")
)

// FiftyOne フィフティワン (51) ゲーム
type FiftyOne struct {
	trumpCards    *TrumpCards
	players       []*FiftyOnePlayer
	tableCards    []*Card
	currentTurn   int
	phase         FiftyOnePhase
	stopCallerIdx int // -1 = 誰もストップしていない
	stopRemaining int // ストップ後の残りターン数
	gameEndFlag   bool
	winnerIdx     int
	config        FiftyOneConfig
	turnNumber    int
	lastAction    string // "exchange_one", "exchange_all", "stop"
	lastHandIdx   int
	lastTableIdx  int
	actionLog     []*ActionLogEntry
}

// NewFiftyOne コンストラクタ
func NewFiftyOne(trumpCards *TrumpCards, players []*FiftyOnePlayer) *FiftyOne {
	return &FiftyOne{
		trumpCards:    trumpCards,
		players:       players,
		tableCards:    make([]*Card, 0, FiftyOneTableSize),
		stopCallerIdx: -1,
		winnerIdx:     -1,
		config:        DefaultFiftyOneConfig(),
	}
}

// NewDefaultFiftyOne returns FiftyOne with the standard 4-player setup (1 human, 3 CPU).
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultFiftyOne() *FiftyOne {
	players := []*FiftyOnePlayer{
		NewFiftyOnePlayer(true),
		NewFiftyOnePlayer(false),
		NewFiftyOnePlayer(false),
		NewFiftyOnePlayer(false),
	}
	return NewFiftyOne(NewTrumpCards(0), players)
}

// Reset ゲームを初期化する
func (fo *FiftyOne) Reset() {
	// カードリセット
	for _, p := range fo.players {
		p.Reset()
	}
	fo.tableCards = make([]*Card, 0, FiftyOneTableSize)

	// シャッフル
	for range fiftyOneShuffleCount {
		fo.trumpCards.Shuffle()
	}

	// 各プレイヤーに5枚ずつ配る
	for range FiftyOneHandSize {
		for _, p := range fo.players {
			c := fo.trumpCards.DrawCard()
			if c != nil {
				p.AddCard(c)
			}
		}
	}

	// 場札5枚
	for range FiftyOneTableSize {
		c := fo.trumpCards.DrawCard()
		if c != nil {
			fo.tableCards = append(fo.tableCards, c)
		}
	}

	// 状態初期化
	fo.currentTurn = 0
	fo.phase = FiftyOnePhasePlay
	fo.stopCallerIdx = -1
	fo.stopRemaining = 0
	fo.gameEndFlag = false
	fo.winnerIdx = -1
	fo.turnNumber = 0
	fo.lastAction = ""
	fo.lastHandIdx = -1
	fo.lastTableIdx = -1
	fo.actionLog = make([]*ActionLogEntry, 0)
}

// ExchangeOne 手札1枚と場札1枚を交換する (人間ターン用)
func (fo *FiftyOne) ExchangeOne(handIdx, tableIdx int) error {
	if fo.gameEndFlag {
		return ErrGameEnded
	}
	if !fo.IsHumanTurn() {
		return ErrNotHumanTurn
	}
	return fo.exchangeOne(0, handIdx, tableIdx)
}

// ExchangeAll 手札5枚と場札5枚を全交換する (人間ターン用)
func (fo *FiftyOne) ExchangeAll() error {
	if fo.gameEndFlag {
		return ErrGameEnded
	}
	if !fo.IsHumanTurn() {
		return ErrNotHumanTurn
	}
	return fo.exchangeAll(0)
}

// Stop ストップ宣言 (人間ターン用)
func (fo *FiftyOne) Stop() error {
	if fo.gameEndFlag {
		return ErrGameEnded
	}
	if !fo.IsHumanTurn() {
		return ErrNotHumanTurn
	}
	return fo.callStop(0)
}

// CpuPlay CPUのターンを実行する
func (fo *FiftyOne) CpuPlay() error {
	if fo.gameEndFlag {
		return ErrGameEnded
	}
	if fo.IsHumanTurn() {
		return ErrFiftyOneNotCpuTurn
	}
	turn := fo.currentTurn
	p := fo.players[turn]

	switch fo.config.CpuDifficulty {
	case FiftyOneDifficultyEasy:
		fo.cpuPlayEasy(turn, p)
	case FiftyOneDifficultyHard:
		fo.cpuPlayHard(turn, p)
	default:
		fo.cpuPlayNormal(turn, p)
	}
	return nil
}

// exchangeOne 指定プレイヤーの手札1枚と場札1枚を交換
func (fo *FiftyOne) exchangeOne(playerIdx, handIdx, tableIdx int) error {
	p := fo.players[playerIdx]
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return fmt.Errorf("invalid hand index: %d", handIdx)
	}
	if tableIdx < 0 || tableIdx >= len(fo.tableCards) {
		return fmt.Errorf("invalid table index: %d", tableIdx)
	}

	handCard := p.GetCard(handIdx)
	tableCard := fo.tableCards[tableIdx]

	// 交換実行: 手札のカードを除去して場札を入れ、場札を差し替え
	p.RemoveCard(handIdx)
	p.InsertCard(tableCard, handIdx)
	fo.tableCards[tableIdx] = handCard

	fo.lastAction = "exchange_one"
	fo.lastHandIdx = handIdx
	fo.lastTableIdx = tableIdx
	fo.appendLog(playerIdx, "exchange_one",
		fmt.Sprintf("exchanged hand[%d] with table[%d]", handIdx, tableIdx),
		[]*Card{tableCard, handCard})
	fo.advanceTurn()
	return nil
}

// exchangeAll 指定プレイヤーの手札5枚と場札5枚を全交換
func (fo *FiftyOne) exchangeAll(playerIdx int) error {
	p := fo.players[playerIdx]
	oldHand := make([]*Card, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		oldHand[i] = p.GetCard(i)
	}
	oldTable := make([]*Card, len(fo.tableCards))
	copy(oldTable, fo.tableCards)

	// 手札をリセットして場札を入れる
	p.Reset()
	for _, c := range oldTable {
		p.AddCard(c)
	}

	// 場札を旧手札に差し替え
	fo.tableCards = oldHand

	fo.lastAction = "exchange_all"
	fo.lastHandIdx = -1
	fo.lastTableIdx = -1
	fo.appendLog(playerIdx, "exchange_all", "exchanged all 5 cards", nil)
	fo.advanceTurn()
	return nil
}

// callStop ストップ宣言
func (fo *FiftyOne) callStop(playerIdx int) error {
	if fo.stopCallerIdx >= 0 {
		return ErrFiftyOneAlreadyStopped
	}
	fo.stopCallerIdx = playerIdx
	// advanceTurn() が1回分デクリメントするため、+1 して補正
	fo.stopRemaining = FiftyOnePlayerCnt

	fo.lastAction = "stop"
	fo.lastHandIdx = -1
	fo.lastTableIdx = -1
	fo.appendLog(playerIdx, "stop",
		fmt.Sprintf("player %d called stop (score: %d)", playerIdx, fo.players[playerIdx].BestSuitScore()),
		nil)
	fo.advanceTurn()
	return nil
}

// advanceTurn ターンを進める
func (fo *FiftyOne) advanceTurn() {
	fo.turnNumber++
	fo.currentTurn = (fo.currentTurn + 1) % FiftyOnePlayerCnt
	if fo.stopCallerIdx >= 0 {
		fo.stopRemaining--
		if fo.stopRemaining <= 0 {
			fo.endGame()
		}
	}
}

// endGame ゲームを終了し、勝者を判定する
func (fo *FiftyOne) endGame() {
	fo.gameEndFlag = true
	fo.phase = FiftyOnePhaseGameEnd

	bestScore := -1
	bestIdx := 0
	for i, p := range fo.players {
		s := p.BestSuitScore()
		if s > bestScore {
			bestScore = s
			bestIdx = i
		}
	}
	fo.winnerIdx = bestIdx

	// 結果ログ
	for i, p := range fo.players {
		fo.appendLog(i, "result",
			fmt.Sprintf("score: %d (suit: %s)", p.BestSuitScore(), fiftyOneSuitName(p.BestSuit())),
			nil)
	}
}

// fiftyOneSuitName スート名を返す
func fiftyOneSuitName(design int) string {
	switch design {
	case CardDesignSpade:
		return "Spade"
	case CardDesignClover:
		return "Clover"
	case CardDesignHeart:
		return "Heart"
	case CardDesignDiamond:
		return "Diamond"
	default:
		return "Unknown"
	}
}

// --- CPU AI ---

// cpuPlayEasy ランダムに行動する
func (fo *FiftyOne) cpuPlayEasy(playerIdx int, _ *FiftyOnePlayer) {
	// ストップ判定: 高スコアなら
	if fo.stopCallerIdx < 0 && fo.players[playerIdx].BestSuitScore() >= fiftyOneCpuStopThresholdEasy {
		_ = fo.callStop(playerIdx)
		return
	}

	// ランダム: 80%の確率で1枚交換、20%で全交換
	if rand.Intn(5) < 4 {
		handIdx := rand.Intn(FiftyOneHandSize)
		tableIdx := rand.Intn(FiftyOneTableSize)
		_ = fo.exchangeOne(playerIdx, handIdx, tableIdx)
	} else {
		_ = fo.exchangeAll(playerIdx)
	}
}

// cpuPlayNormal ベストスートを狙って交換する
func (fo *FiftyOne) cpuPlayNormal(playerIdx int, p *FiftyOnePlayer) {
	currentScore := p.BestSuitScore()

	// ストップ判定
	if fo.stopCallerIdx < 0 && currentScore >= fiftyOneCpuStopThresholdNormal {
		_ = fo.callStop(playerIdx)
		return
	}

	bestSuit := p.BestSuit()

	// ベストスートのカードが場札にあれば取得し、非ベストスートの手札と交換
	bestExchange := fo.findBestSingleExchange(p, bestSuit)
	if bestExchange != nil {
		_ = fo.exchangeOne(playerIdx, bestExchange.handIdx, bestExchange.tableIdx)
		return
	}

	// 全交換が有利なら実行
	if fo.evaluateExchangeAll(p) > currentScore {
		_ = fo.exchangeAll(playerIdx)
		return
	}

	// 改善なし: ランダムに1枚交換
	handIdx := rand.Intn(FiftyOneHandSize)
	tableIdx := rand.Intn(FiftyOneTableSize)
	_ = fo.exchangeOne(playerIdx, handIdx, tableIdx)
}

// cpuPlayHard 最善手を計算する
func (fo *FiftyOne) cpuPlayHard(playerIdx int, p *FiftyOnePlayer) {
	currentScore := p.BestSuitScore()

	// ストップ判定: 高スコアなら (かつまだストップされていない)
	if fo.stopCallerIdx < 0 && currentScore >= fiftyOneCpuStopThresholdHard {
		_ = fo.callStop(playerIdx)
		return
	}

	// 全ての1枚交換を評価し、最善手を見つける
	bestScore := currentScore
	bestHandIdx := -1
	bestTableIdx := -1

	for hi := range FiftyOneHandSize {
		for ti := range FiftyOneTableSize {
			score := fo.evaluateSingleExchange(p, hi, ti)
			if score > bestScore {
				bestScore = score
				bestHandIdx = hi
				bestTableIdx = ti
			}
		}
	}

	// 全交換も評価
	allScore := fo.evaluateExchangeAll(p)

	if allScore > bestScore && allScore > currentScore {
		_ = fo.exchangeAll(playerIdx)
		return
	}
	if bestHandIdx >= 0 && bestScore > currentScore {
		_ = fo.exchangeOne(playerIdx, bestHandIdx, bestTableIdx)
		return
	}

	// 改善なし: 最も弱いカードを交換
	worstHand := fo.findWorstCard(p)
	bestTable := fo.findBestTableCard(p)
	_ = fo.exchangeOne(playerIdx, worstHand, bestTable)
}

type exchangeCandidate struct {
	handIdx  int
	tableIdx int
	score    int
}

// findBestSingleExchange ベストスートを改善する最良の1枚交換を見つける
func (fo *FiftyOne) findBestSingleExchange(p *FiftyOnePlayer, targetSuit int) *exchangeCandidate {
	currentScore := p.BestSuitScore()
	var best *exchangeCandidate

	for ti, tc := range fo.tableCards {
		if tc.GetDesign() != targetSuit {
			continue
		}
		// 全手札との交換を評価（同スート内のアップグレードも含む）
		for hi := range p.GetCardsSize() {
			score := fo.evaluateSingleExchange(p, hi, ti)
			if score > currentScore && (best == nil || score > best.score) {
				best = &exchangeCandidate{handIdx: hi, tableIdx: ti, score: score}
			}
		}
	}
	return best
}

// evaluateSingleExchange 1枚交換後のスコアをシミュレーション
func (fo *FiftyOne) evaluateSingleExchange(p *FiftyOnePlayer, handIdx, tableIdx int) int {
	handCard := p.GetCard(handIdx)
	tableCard := fo.tableCards[tableIdx]

	// 一時的に交換してスコア計算
	p.RemoveCard(handIdx)
	p.InsertCard(tableCard, handIdx)
	fo.tableCards[tableIdx] = handCard

	score := p.BestSuitScore()

	// 元に戻す
	p.RemoveCard(handIdx)
	p.InsertCard(handCard, handIdx)
	fo.tableCards[tableIdx] = tableCard

	return score
}

// evaluateExchangeAll 全交換後のスコアを計算
func (fo *FiftyOne) evaluateExchangeAll(_ *FiftyOnePlayer) int {
	// 場札のみでベストスートスコアを計算
	scores := map[int]int{}
	for _, c := range fo.tableCards {
		scores[c.GetDesign()] += fiftyOneCardScore(c)
	}
	best := 0
	for _, s := range scores {
		if s > best {
			best = s
		}
	}
	return best
}

// findWorstCard 手札の中でベストスートに最も貢献しないカードのインデックスを返す
func (fo *FiftyOne) findWorstCard(p *FiftyOnePlayer) int {
	bestSuit := p.BestSuit()
	worstIdx := 0
	worstScore := fiftyOneCardScore(p.GetCard(0))
	worstIsBestSuit := p.GetCard(0).GetDesign() == bestSuit

	for i := 1; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		s := fiftyOneCardScore(c)
		isBestSuit := c.GetDesign() == bestSuit

		// 非ベストスートのカードを優先的に選ぶ
		if worstIsBestSuit && !isBestSuit {
			worstIdx = i
			worstScore = s
			worstIsBestSuit = false
		} else if worstIsBestSuit == isBestSuit && s < worstScore {
			worstIdx = i
			worstScore = s
			worstIsBestSuit = isBestSuit
		}
	}
	return worstIdx
}

// findBestTableCard ベストスートに合う場札、なければ最高スコアの場札を返す
func (fo *FiftyOne) findBestTableCard(p *FiftyOnePlayer) int {
	bestSuit := p.BestSuit()
	bestIdx := 0
	bestScore := 0

	for i, c := range fo.tableCards {
		s := fiftyOneCardScore(c)
		isBestSuit := c.GetDesign() == bestSuit
		if isBestSuit && s > bestScore {
			bestIdx = i
			bestScore = s
		}
	}
	if bestScore > 0 {
		return bestIdx
	}
	// ベストスートの場札がない場合、最高スコアのカードを選ぶ
	for i, c := range fo.tableCards {
		s := fiftyOneCardScore(c)
		if s > bestScore {
			bestIdx = i
			bestScore = s
		}
	}
	return bestIdx
}

// --- Getters ---

// GetPhase 現在のフェーズを取得
func (fo *FiftyOne) GetPhase() FiftyOnePhase { return fo.phase }

// GetGameEndFlag ゲーム終了フラグ
func (fo *FiftyOne) GetGameEndFlag() bool { return fo.gameEndFlag }

// GetCurrentTurn 現在のターンプレイヤーインデックス
func (fo *FiftyOne) GetCurrentTurn() int { return fo.currentTurn }

// IsHumanTurn 人間のターンかどうか
func (fo *FiftyOne) IsHumanTurn() bool {
	return fo.currentTurn < len(fo.players) && fo.players[fo.currentTurn].GetIsHuman()
}

// GetPlayerCnt プレイヤー数
func (fo *FiftyOne) GetPlayerCnt() int { return len(fo.players) }

// GetPlayer 指定インデックスのプレイヤー
func (fo *FiftyOne) GetPlayer(i int) *FiftyOnePlayer {
	return getPlayer(fo.players, i)
}

// GetWinnerIdx 勝者インデックス (-1 = 未確定)
func (fo *FiftyOne) GetWinnerIdx() int { return fo.winnerIdx }

// GetTableCards 場札
func (fo *FiftyOne) GetTableCards() []*Card { return fo.tableCards }

// GetStopCallerIdx ストップ宣言者インデックス (-1 = 未宣言)
func (fo *FiftyOne) GetStopCallerIdx() int { return fo.stopCallerIdx }

// GetTurnNumber ターン番号
func (fo *FiftyOne) GetTurnNumber() int { return fo.turnNumber }

// GetLastAction 直前のアクション種別
func (fo *FiftyOne) GetLastAction() string { return fo.lastAction }

// GetLastHandIdx 直前の手札インデックス
func (fo *FiftyOne) GetLastHandIdx() int { return fo.lastHandIdx }

// GetLastTableIdx 直前の場札インデックス
func (fo *FiftyOne) GetLastTableIdx() int { return fo.lastTableIdx }

// GetConfig 設定取得
func (fo *FiftyOne) GetConfig() FiftyOneConfig { return fo.config }

// SetConfig 設定セット
func (fo *FiftyOne) SetConfig(cfg FiftyOneConfig) { fo.config = cfg }

// GetActionLog アクションログ取得
func (fo *FiftyOne) GetActionLog() []*ActionLogEntry { return fo.actionLog }

// appendLog アクションログ追加
func (fo *FiftyOne) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	fo.actionLog = append(fo.actionLog, &ActionLogEntry{
		TurnNumber: fo.turnNumber,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- JSON ---

// fiftyOneJSON is the JSON wire format for FiftyOne.
type fiftyOneJSON struct {
	TrumpCards    *TrumpCards       `json:"tc"`
	Players       []*FiftyOnePlayer `json:"ps"`
	TableCards    []*Card           `json:"tb"`
	CurrentTurn   int               `json:"ct"`
	Phase         FiftyOnePhase     `json:"ph"`
	StopCallerIdx int               `json:"sc"`
	StopRemaining int               `json:"sr"`
	GameEndFlag   bool              `json:"ge"`
	WinnerIdx     int               `json:"wi"`
	Config        FiftyOneConfig    `json:"cf"`
	TurnNumber    int               `json:"tn"`
	LastAction    string            `json:"la"`
	LastHandIdx   int               `json:"lh"`
	LastTableIdx  int               `json:"lt"`
	ActionLog     []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (fo *FiftyOne) MarshalJSON() ([]byte, error) {
	return json.Marshal(fiftyOneJSON{
		TrumpCards:    fo.trumpCards,
		Players:       fo.players,
		TableCards:    fo.tableCards,
		CurrentTurn:   fo.currentTurn,
		Phase:         fo.phase,
		StopCallerIdx: fo.stopCallerIdx,
		StopRemaining: fo.stopRemaining,
		GameEndFlag:   fo.gameEndFlag,
		WinnerIdx:     fo.winnerIdx,
		Config:        fo.config,
		TurnNumber:    fo.turnNumber,
		LastAction:    fo.lastAction,
		LastHandIdx:   fo.lastHandIdx,
		LastTableIdx:  fo.lastTableIdx,
		ActionLog:     fo.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (fo *FiftyOne) UnmarshalJSON(data []byte) error {
	var j fiftyOneJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	fo.trumpCards = j.TrumpCards
	if fo.trumpCards == nil {
		fo.trumpCards = NewTrumpCards(0)
	}
	fo.players = j.Players
	if fo.players == nil {
		fo.players = make([]*FiftyOnePlayer, 0)
	}
	fo.tableCards = j.TableCards
	if fo.tableCards == nil {
		fo.tableCards = make([]*Card, 0)
	}
	fo.currentTurn = j.CurrentTurn
	fo.phase = j.Phase
	fo.stopCallerIdx = j.StopCallerIdx
	fo.stopRemaining = j.StopRemaining
	fo.gameEndFlag = j.GameEndFlag
	fo.winnerIdx = j.WinnerIdx
	fo.config = j.Config
	fo.turnNumber = j.TurnNumber
	fo.lastAction = j.LastAction
	fo.lastHandIdx = j.LastHandIdx
	fo.lastTableIdx = j.LastTableIdx
	fo.actionLog = j.ActionLog
	if fo.actionLog == nil {
		fo.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
