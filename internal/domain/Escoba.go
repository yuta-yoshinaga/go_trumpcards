//go:build !js || !wasm || classic

// Package domain エスコバ (Escoba) のドメインモデル。
//
// Escoba はスペイン発祥の 40 枚カード捕獲ゲーム。Scopa 系だが、捕獲は「合計ちょうど 15」で行う。
// 4 人それぞれが個人戦で戦い、各プレイヤーに 3 枚ずつ配り、場に 4 枚を公開する。手番では手から
// 1 枚出し、その値と場札の合計が 15 になる組を捕獲する (なければ場に置く)。手札が尽きたら山札
// から再度 3 枚ずつ配る。場を一掃すると「エスコバ」で +1 点 (最終手は除く)。
//
// 全カードを配り切ってプレイし終えたら、残りの場札を最後に捕獲したプレイヤーが取り、得点を集計する:
// 最多カード / 最多エスパダ (剣=♠) / 最多 7 / 最多オロ (金貨=♦) で各 +1、エスコバ数だけ +1、
// エスパダの A と 7 を取ったプレイヤーに各 +1。先に目標点 (既定 10) に達したプレイヤーが勝利する。
//
// カード値と部分集合の列挙は Scopa と共有 (ScopaCardValue / scopaSubsetsSummingTo;
// 絵札は J=8,Q=9,K=10)。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// EscobaPlayerCnt Escoba プレイヤー数
const EscobaPlayerCnt = 4

// EscobaPackSize 1 回の配りでプレイヤーが受け取る枚数
const EscobaPackSize = 3

// EscobaInitialTable ラウンド開始時に場へ公開する枚数
const EscobaInitialTable = 4

// EscobaCaptureTarget 捕獲の目標合計値
const EscobaCaptureTarget = 15

// Escoba のフェーズ定数
const (
	// EscobaPhasePlayerTurn プレイヤーターン
	EscobaPhasePlayerTurn = "playerTurn"
	// EscobaPhaseRoundEnd ラウンド終了
	EscobaPhaseRoundEnd = "roundEnd"
	// EscobaPhaseGameEnd ゲーム終了
	EscobaPhaseGameEnd = "gameEnd"
)

// EscobaScoreDetail はラウンド得点の内訳 (プレイヤー単位)。
type EscobaScoreDetail struct {
	Cards   [EscobaPlayerCnt]int `json:"cards"`
	Espadas [EscobaPlayerCnt]int `json:"espadas"`
	Sevens  [EscobaPlayerCnt]int `json:"sevens"`
	Oros    [EscobaPlayerCnt]int `json:"oros"`
	Escobas [EscobaPlayerCnt]int `json:"escobas"`
	Gained  [EscobaPlayerCnt]int `json:"gained"`
	AceEsp  int                  `json:"aceEspada"`   // A♠ を取ったプレイヤー (-1 = なし)
	SeteEsp int                  `json:"sevenEspada"` // 7♠ を取ったプレイヤー (-1 = なし)
}

// Escoba エスコバ ゲームクラス
type Escoba struct {
	trumpCards      *TrumpCards
	players         []*ScopaPlayer
	config          EscobaConfig
	phase           string
	roundNumber     int
	currentTurn     int
	dealerIdx       int
	tableCards      []*Card
	lastCaptureIdx  int
	gameEndFlag     bool
	winnerIdx       int // -1: 未確定
	lastRoundDetail *EscobaScoreDetail
	actionLogBase
}

// NewEscoba コンストラクタ
func NewEscoba(trumpCards *TrumpCards, players []*ScopaPlayer, config EscobaConfig) *Escoba {
	return &Escoba{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		phase:          EscobaPhasePlayerTurn,
		winnerIdx:      -1,
		lastCaptureIdx: -1,
	}
}

// NewDefaultEscoba 標準の 4 人 (人間 idx 0 + CPU 3) セットアップを返す。
func NewDefaultEscoba() *Escoba {
	players := make([]*ScopaPlayer, EscobaPlayerCnt)
	for i := range players {
		players[i] = NewScopaPlayer(i == 0)
	}
	return NewEscoba(NewTrumpCardsScopa(), players, DefaultEscobaConfig())
}

// Reset ゲーム初期化
func (e *Escoba) Reset() {
	e.gameEndFlag = false
	e.winnerIdx = -1
	e.roundNumber = 1
	e.dealerIdx = 0
	e.lastRoundDetail = nil
	e.actionLog = nil
	for _, p := range e.players {
		p.ResetTotalScore()
	}
	e.startRound()
}

// NextRound 次のラウンドを開始する
func (e *Escoba) NextRound() {
	if e.phase != EscobaPhaseRoundEnd {
		return
	}
	e.roundNumber++
	e.dealerIdx = (e.dealerIdx + 1) % EscobaPlayerCnt
	e.startRound()
}

// startRound 1 ラウンドを開始する: シャッフル・場に 4 枚・各 3 枚配り。
func (e *Escoba) startRound() {
	e.tableCards = nil
	e.lastCaptureIdx = -1
	for _, p := range e.players {
		p.Reset()
		p.ResetCaptured()
		p.ResetScopaCount()
	}
	e.trumpCards.Replenish()
	e.trumpCards.Shuffle()
	for i := 0; i < EscobaInitialTable; i++ {
		if c := e.trumpCards.DrawCard(); c != nil {
			e.tableCards = append(e.tableCards, c)
		}
	}
	e.dealPack()
	e.currentTurn = (e.dealerIdx + 1) % EscobaPlayerCnt
	e.phase = EscobaPhasePlayerTurn
	e.appendLog(-1, "deal", fmt.Sprintf("round %d: 4 table cards, %d to each", e.roundNumber, EscobaPackSize), nil)
}

// dealPack 各プレイヤーに EscobaPackSize 枚ずつ配る (山札が尽きるまで)。
func (e *Escoba) dealPack() {
	for range EscobaPackSize {
		for i := range EscobaPlayerCnt {
			player := e.players[(e.dealerIdx+1+i)%EscobaPlayerCnt]
			if c := e.trumpCards.DrawCard(); c != nil {
				player.AddCard(c)
			}
		}
	}
}

// EscobaCaptures は playedCard を出したとき合計 15 で捕獲できる場札インデックス集合を返す。
func EscobaCaptures(playedCard *Card, tableCards []*Card) [][]int {
	if playedCard == nil {
		return nil
	}
	target := EscobaCaptureTarget - ScopaCardValue(playedCard)
	if target <= 0 {
		return nil
	}
	return scopaSubsetsSummingTo(tableCards, target)
}

// PlayerPlay 人間プレイヤーが手札 handIdx を出し、tableIdxs の場札を捕獲する (空なら場に置く)。
func (e *Escoba) PlayerPlay(handIdx int, tableIdxs []int) error {
	if e.gameEndFlag {
		return ErrGameEnded
	}
	if e.phase != EscobaPhasePlayerTurn {
		return ErrWrongPhase
	}
	if !e.players[e.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return e.applyPlay(e.currentTurn, handIdx, tableIdxs)
}

// CpuPlay 現在の手番が CPU の場合に 1 手進める。
func (e *Escoba) CpuPlay() {
	if e.gameEndFlag || e.phase != EscobaPhasePlayerTurn {
		return
	}
	idx := e.currentTurn
	if e.players[idx].GetIsHuman() {
		return
	}
	handIdx, tableIdxs := e.chooseCpuPlay(idx)
	_ = e.applyPlay(idx, handIdx, tableIdxs)
}

// applyPlay は 1 手 (place / capture) を実行する共通処理。
func (e *Escoba) applyPlay(playerIdx, handIdx int, tableIdxs []int) error {
	player := e.players[playerIdx]
	if handIdx < 0 || handIdx >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", handIdx))
	}
	handCard := player.GetCard(handIdx)

	if len(tableIdxs) > 0 {
		if !e.isValidCapture(handCard, tableIdxs) {
			return NewDomainError(ErrInvalidPlay, "selected table cards do not sum to 15 with the played card")
		}
	} else if len(EscobaCaptures(handCard, e.tableCards)) > 0 {
		// Escoba は強制捕獲。合計 15 を作れる組が場にある場合、
		// 単に場に置く (Lay) ことは許されない。
		return NewDomainError(ErrInvalidPlay, "must capture: a combination summing to 15 exists on the table")
	}

	_ = player.RemoveCard(handIdx)

	if len(tableIdxs) == 0 {
		e.tableCards = append(e.tableCards, handCard)
		e.appendLog(playerIdx, "play", "placed on table", []*Card{handCard})
		e.postActionAdvance()
		return nil
	}

	captured := make([]*Card, 0, len(tableIdxs))
	for _, idx := range tableIdxs {
		captured = append(captured, e.tableCards[idx])
	}
	e.removeTableCardsByIndex(tableIdxs)
	pile := make([]*Card, 0, len(captured)+1)
	pile = append(pile, handCard)
	pile = append(pile, captured...)
	player.AddCaptured(pile)
	e.lastCaptureIdx = playerIdx

	if len(e.tableCards) == 0 && !e.isLastPlay() {
		player.IncrementScopa()
		e.appendLog(playerIdx, "escoba", "swept the table (escoba)", pile)
	} else {
		e.appendLog(playerIdx, "capture", fmt.Sprintf("captured %d card(s)", len(captured)), pile)
	}
	e.postActionAdvance()
	return nil
}

// isValidCapture は selectedIdxs が handCard で合計 15 になる合法な捕獲かを検証する。
func (e *Escoba) isValidCapture(handCard *Card, selectedIdxs []int) bool {
	if handCard == nil || len(selectedIdxs) == 0 {
		return false
	}
	seen := make(map[int]bool, len(selectedIdxs))
	sum := ScopaCardValue(handCard)
	for _, idx := range selectedIdxs {
		if idx < 0 || idx >= len(e.tableCards) || seen[idx] {
			return false
		}
		seen[idx] = true
		sum += ScopaCardValue(e.tableCards[idx])
	}
	return sum == EscobaCaptureTarget
}

// postActionAdvance はアクション後の進行処理。手札切れ + 山札ありなら再配布。
func (e *Escoba) postActionAdvance() {
	if e.isRoundOver() {
		e.finishRound()
		return
	}
	e.currentTurn = (e.currentTurn + 1) % EscobaPlayerCnt
	if e.allHandsEmpty() && e.trumpCards.GetRemainingCount() > 0 {
		e.dealPack()
		e.currentTurn = (e.dealerIdx + 1) % EscobaPlayerCnt
	}
}

// isRoundOver は手札も山札も尽きたか。
func (e *Escoba) isRoundOver() bool {
	return e.allHandsEmpty() && e.trumpCards.GetRemainingCount() == 0
}

// isLastPlay は今の捕獲がラウンド最後の 1 手か (場全取りをエスコバに数えないため)。
func (e *Escoba) isLastPlay() bool {
	return e.allHandsEmpty() && e.trumpCards.GetRemainingCount() == 0
}

// finishRound はラウンド終了処理: 残り場札を最後の捕獲者へ渡し、得点集計と終了判定。
func (e *Escoba) finishRound() {
	if e.lastCaptureIdx >= 0 && len(e.tableCards) > 0 {
		e.players[e.lastCaptureIdx].AddCaptured(e.tableCards)
		e.appendLog(e.lastCaptureIdx, "sweep_leftover", fmt.Sprintf("took %d leftover table card(s)", len(e.tableCards)), e.tableCards)
		e.tableCards = nil
	}
	det := e.scoreRound()
	e.lastRoundDetail = det
	for i := range e.players {
		e.players[i].AddScore(det.Gained[i])
	}
	e.appendLog(-1, "round_score", fmt.Sprintf("round %d scored", e.roundNumber), nil)

	for i := range e.players {
		if e.players[i].GetTotalScore() >= e.config.TargetScore {
			e.finishGame()
			return
		}
	}
	e.phase = EscobaPhaseRoundEnd
}

// scoreRound はプレイヤー単位の得点内訳を計算する。
func (e *Escoba) scoreRound() *EscobaScoreDetail {
	det := &EscobaScoreDetail{AceEsp: -1, SeteEsp: -1}
	cardsM := map[int]int{}
	espadasM := map[int]int{}
	sevensM := map[int]int{}
	orosM := map[int]int{}
	for i, p := range e.players {
		det.Cards[i] = p.CapturedCount()
		det.Escobas[i] = p.GetScopaCount()
		cardsM[i] = p.CapturedCount()
		for _, c := range p.GetCapturedCards() {
			if c.GetDesign() == CardDesignSpade {
				det.Espadas[i]++
				espadasM[i]++
				if c.GetValue() == 1 {
					det.AceEsp = i
				}
				if c.GetValue() == 7 {
					det.SeteEsp = i
				}
			}
			if ScopaIsSeven(c) {
				det.Sevens[i]++
				sevensM[i]++
			}
			if ScopaIsDiamond(c) {
				det.Oros[i]++
				orosM[i]++
			}
		}
	}
	mostCards := uniqueMaxIndex(cardsM)
	mostEspadas := uniqueMaxIndex(espadasM)
	mostSevens := uniqueMaxIndex(sevensM)
	mostOros := uniqueMaxIndex(orosM)
	for i := range e.players {
		score := det.Escobas[i]
		if i == mostCards {
			score++
		}
		if i == mostEspadas {
			score++
		}
		if i == mostSevens {
			score++
		}
		if i == mostOros {
			score++
		}
		if det.AceEsp == i {
			score++
		}
		if det.SeteEsp == i {
			score++
		}
		det.Gained[i] = score
	}
	return det
}

// finishGame はゲームを終了させ勝者を決定する。
func (e *Escoba) finishGame() {
	e.gameEndFlag = true
	e.phase = EscobaPhaseGameEnd
	best, bestScore := 0, -1
	for i, p := range e.players {
		if p.GetTotalScore() > bestScore {
			bestScore = p.GetTotalScore()
			best = i
		}
	}
	e.winnerIdx = best
	e.appendLog(-1, "game_end", fmt.Sprintf("player %d wins with %d", best, bestScore), nil)
}

// removeTableCardsByIndex は降順に並び替えてから tableCards を削除する。
func (e *Escoba) removeTableCardsByIndex(idxs []int) {
	for _, idx := range sortIndicesDescending(idxs) {
		if idx >= 0 && idx < len(e.tableCards) {
			e.tableCards = append(e.tableCards[:idx], e.tableCards[idx+1:]...)
		}
	}
}

// allHandsEmpty 全プレイヤーの手札が空か。
func (e *Escoba) allHandsEmpty() bool {
	for _, p := range e.players {
		if p.GetCardsSize() > 0 {
			return false
		}
	}
	return true
}

// chooseCpuPlay は CPU の手を選ぶ。15 で捕獲できるなら最大枚数 (エスコバ優先)、無ければ最大値を置く。
func (e *Escoba) chooseCpuPlay(playerIdx int) (int, []int) {
	p := e.players[playerIdx]
	bestHand, bestCapture, bestScore := -1, []int(nil), -1
	for h := 0; h < p.GetCardsSize(); h++ {
		caps := EscobaCaptures(p.GetCard(h), e.tableCards)
		for _, cap := range caps {
			score := len(cap)
			if len(e.tableCards) == len(cap) {
				score += 100 // エスコバ
			}
			if score > bestScore {
				bestScore, bestHand, bestCapture = score, h, cap
			}
		}
	}
	if bestHand >= 0 {
		return bestHand, bestCapture
	}
	// 捕獲不可: 最大値の札を場に置く (相手に 15 を作らせにくい簡易戦略)。
	dump, dumpVal := 0, -1
	for h := 0; h < p.GetCardsSize(); h++ {
		if v := ScopaCardValue(p.GetCard(h)); v > dumpVal {
			dumpVal, dump = v, h
		}
	}
	return dump, nil
}

// --- 状態アクセサ ---

// GetPhase 現在のフェーズ取得
func (e *Escoba) GetPhase() string { return e.phase }

// SetPhase フェーズ設定 (テスト用)
func (e *Escoba) SetPhase(phase string) { e.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (e *Escoba) GetRoundNumber() int { return e.roundNumber }

// GetCurrentTurn 現在の手番プレイヤーインデックス取得
func (e *Escoba) GetCurrentTurn() int { return e.currentTurn }

// SetCurrentTurn 手番設定 (テスト用)
func (e *Escoba) SetCurrentTurn(idx int) { e.currentTurn = idx }

// GetDealerIdx ディーラーインデックス取得
func (e *Escoba) GetDealerIdx() int { return e.dealerIdx }

// GetTableCards 場札取得
func (e *Escoba) GetTableCards() []*Card { return e.tableCards }

// SetTableCards 場札設定 (テスト用)
func (e *Escoba) SetTableCards(cards []*Card) { e.tableCards = cards }

// GetStockRemaining 山札の残り枚数
func (e *Escoba) GetStockRemaining() int { return e.trumpCards.GetRemainingCount() }

// GetLastCaptureIdx 最後に捕獲したプレイヤー取得 (-1: なし)
func (e *Escoba) GetLastCaptureIdx() int { return e.lastCaptureIdx }

// GetGameEndFlag ゲーム終了フラグ取得
func (e *Escoba) GetGameEndFlag() bool { return e.gameEndFlag }

// GetWinnerIdx 勝者プレイヤー取得 (-1: 未確定)
func (e *Escoba) GetWinnerIdx() int { return e.winnerIdx }

// GetLastRoundDetail 直近ラウンドの得点内訳取得
func (e *Escoba) GetLastRoundDetail() *EscobaScoreDetail { return e.lastRoundDetail }

// GetPlayerCnt プレイヤー数取得
func (e *Escoba) GetPlayerCnt() int { return len(e.players) }

// GetPlayer プレイヤー取得
func (e *Escoba) GetPlayer(i int) *ScopaPlayer {
	return getPlayer(e.players, i)
}

// GetConfig 設定取得
func (e *Escoba) GetConfig() EscobaConfig { return e.config }

// SetConfig 設定変更
func (e *Escoba) SetConfig(cfg EscobaConfig) { e.config = cfg }

// IsHumanTurn 現在の手番が人間かどうか
func (e *Escoba) IsHumanTurn() bool {
	if e.gameEndFlag || e.currentTurn < 0 || e.currentTurn >= len(e.players) {
		return false
	}
	return e.players[e.currentTurn].GetIsHuman()
}

// GetValidCaptures は handIdx のカードで合計 15 になる捕獲パターンを返す。
func (e *Escoba) GetValidCaptures(handIdx int) [][]int {
	if e.currentTurn < 0 || e.currentTurn >= len(e.players) {
		return nil
	}
	p := e.players[e.currentTurn]
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return nil
	}
	return EscobaCaptures(p.GetCard(handIdx), e.tableCards)
}

// --- JSON ---

type escobaJSON struct {
	TrumpCards      *TrumpCards        `json:"tc"`
	Players         []*ScopaPlayer     `json:"ps"`
	Config          EscobaConfig       `json:"cf"`
	Phase           string             `json:"ph"`
	RoundNumber     int                `json:"rn"`
	CurrentTurn     int                `json:"ct"`
	DealerIdx       int                `json:"di"`
	TableCards      []*Card            `json:"tb"`
	LastCaptureIdx  int                `json:"lc"`
	GameEndFlag     bool               `json:"ge"`
	WinnerIdx       int                `json:"wi"`
	LastRoundDetail *EscobaScoreDetail `json:"rd"`
	ActionLog       []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (e *Escoba) MarshalJSON() ([]byte, error) {
	return json.Marshal(escobaJSON{
		TrumpCards:      e.trumpCards,
		Players:         e.players,
		Config:          e.config,
		Phase:           e.phase,
		RoundNumber:     e.roundNumber,
		CurrentTurn:     e.currentTurn,
		DealerIdx:       e.dealerIdx,
		TableCards:      e.tableCards,
		LastCaptureIdx:  e.lastCaptureIdx,
		GameEndFlag:     e.gameEndFlag,
		WinnerIdx:       e.winnerIdx,
		LastRoundDetail: e.lastRoundDetail,
		ActionLog:       e.actionLog,
	})
}

const escobaMaxSliceLen = 5000

var errEscobaSnapshot = errors.New("escoba: invalid serialised game state")

func escobaIdxInRange(i int) bool { return i >= 0 && i < EscobaPlayerCnt }

// UnmarshalJSON implements json.Unmarshaler.
func (e *Escoba) UnmarshalJSON(data []byte) error {
	var j escobaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != EscobaPlayerCnt ||
		len(j.TableCards) > escobaMaxSliceLen || len(j.ActionLog) > escobaMaxSliceLen ||
		!escobaIdxInRange(j.CurrentTurn) || !escobaIdxInRange(j.DealerIdx) ||
		j.LastCaptureIdx < -1 || j.LastCaptureIdx >= EscobaPlayerCnt ||
		j.WinnerIdx < -1 || j.WinnerIdx >= EscobaPlayerCnt ||
		j.RoundNumber < 1 ||
		(j.Phase != EscobaPhasePlayerTurn && j.Phase != EscobaPhaseRoundEnd && j.Phase != EscobaPhaseGameEnd) {
		return errEscobaSnapshot
	}
	for _, p := range j.Players {
		if p == nil {
			return errEscobaSnapshot
		}
	}
	for _, c := range j.TableCards {
		if c == nil {
			return errEscobaSnapshot
		}
	}
	for _, entry := range j.ActionLog {
		if entry == nil {
			return errEscobaSnapshot
		}
	}
	e.trumpCards = j.TrumpCards
	if e.trumpCards == nil {
		e.trumpCards = NewTrumpCardsScopa()
	}
	e.players = j.Players
	e.config = j.Config
	e.phase = j.Phase
	e.roundNumber = j.RoundNumber
	e.currentTurn = j.CurrentTurn
	e.dealerIdx = j.DealerIdx
	e.tableCards = j.TableCards
	if e.tableCards == nil {
		e.tableCards = make([]*Card, 0)
	}
	e.lastCaptureIdx = j.LastCaptureIdx
	e.gameEndFlag = j.GameEndFlag
	e.winnerIdx = j.WinnerIdx
	e.lastRoundDetail = j.LastRoundDetail
	e.actionLog = j.ActionLog
	if e.actionLog == nil {
		e.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
