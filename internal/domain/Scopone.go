//go:build !js || !wasm || classic

// Package domain スコポーネ (Scopone) のドメインモデル。
//
// Scopone は 4 人 2 チームで遊ぶイタリアのトリック (キャプチャ) ゲーム。40 枚デッキ
// (1〜10 × 4 スート) を全員に 10 枚ずつ配り切る (場札・山札なし)。手番のプレイヤーは手から
// 1 枚出し、その値で場札を捕獲する (単一一致が可能なら単一を取る; なければ合計が一致する組)。
// 捕獲できないカードは場に置く。場を一掃すると「スコパ」で +1 点 (最終手は除く)。
//
// 全 40 枚をプレイしたら残りの場札を最後に捕獲した側が取り、チーム単位で得点を集計する:
// 最多カード (carte) / 最多ダイヤ (denari) / 最多 7 (settant'uno の代用) / 7♦ (settebello) /
// スコパ数。先に目標点 (既定 11) に達したチームが勝利する。
//
// 捕獲・評価ロジックは Scopa と共有する (EnumerateScopaCaptures / isValidScopaCapture /
// ScopaCardValue / ScopaIsSetteBello / ScopaIsDiamond / ScopaIsSeven / ScopaScore*)。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ScoponePlayerCnt Scopone プレイヤー数
const ScoponePlayerCnt = 4

// ScoponeTeamCnt チーム数
const ScoponeTeamCnt = 2

// ScoponeHandSize 各プレイヤーの配り札枚数
const ScoponeHandSize = 10

// Scopone のフェーズ定数
const (
	// ScoponePhasePlayerTurn プレイヤーターン
	ScoponePhasePlayerTurn = "playerTurn"
	// ScoponePhaseRoundEnd ラウンド終了
	ScoponePhaseRoundEnd = "roundEnd"
	// ScoponePhaseGameEnd ゲーム終了
	ScoponePhaseGameEnd = "gameEnd"
)

// ScoponeScoreDetail はラウンド得点の内訳 (チーム単位)。
type ScoponeScoreDetail struct {
	Cards        [ScoponeTeamCnt]int `json:"cards"`
	Diamonds     [ScoponeTeamCnt]int `json:"diamonds"`
	Sevens       [ScoponeTeamCnt]int `json:"sevens"`
	Scopas       [ScoponeTeamCnt]int `json:"scopas"`
	Gained       [ScoponeTeamCnt]int `json:"gained"`
	SettebelloTm int                 `json:"settebello"` // 7♦ を取ったチーム (-1 = なし)
}

// Scopone スコポーネ ゲームクラス
type Scopone struct {
	trumpCards      *TrumpCards
	players         []*ScopaPlayer
	config          ScoponeConfig
	phase           string
	roundNumber     int
	currentTurn     int
	dealerIdx       int
	tableCards      []*Card
	lastCaptureIdx  int
	teamScores      [ScoponeTeamCnt]int
	gameEndFlag     bool
	winnerTeam      int // -1: 未確定
	lastRoundDetail *ScoponeScoreDetail
	actionLogBase
}

// NewScopone コンストラクタ
func NewScopone(trumpCards *TrumpCards, players []*ScopaPlayer, config ScoponeConfig) *Scopone {
	return &Scopone{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		phase:          ScoponePhasePlayerTurn,
		winnerTeam:     -1,
		lastCaptureIdx: -1,
	}
}

// NewDefaultScopone 標準の 4 人 (人間 idx 0 + CPU 3) セットアップを返す。
func NewDefaultScopone() *Scopone {
	players := make([]*ScopaPlayer, ScoponePlayerCnt)
	for i := range players {
		players[i] = NewScopaPlayer(i == 0)
	}
	return NewScopone(NewTrumpCardsScopa(), players, DefaultScoponeConfig())
}

// ScoponeTeamOf は playerIdx の所属チーム (0+2 / 1+3) を返す。
func ScoponeTeamOf(playerIdx int) int { return playerIdx % ScoponeTeamCnt }

// Reset ゲーム初期化
func (s *Scopone) Reset() {
	s.gameEndFlag = false
	s.winnerTeam = -1
	s.roundNumber = 1
	s.dealerIdx = 0
	s.teamScores = [ScoponeTeamCnt]int{}
	s.lastRoundDetail = nil
	s.actionLog = nil
	for _, p := range s.players {
		p.ResetTotalScore()
	}
	s.startRound()
}

// NextRound 次のラウンドを開始する
func (s *Scopone) NextRound() {
	if s.phase != ScoponePhaseRoundEnd {
		return
	}
	s.roundNumber++
	s.dealerIdx = (s.dealerIdx + 1) % ScoponePlayerCnt
	s.startRound()
}

// startRound 1 ラウンドを開始する: 全 40 枚を 10 枚ずつ配り切る (場札なし)。
func (s *Scopone) startRound() {
	s.tableCards = nil
	s.lastCaptureIdx = -1
	for _, p := range s.players {
		p.Reset()
		p.ResetCaptured()
		p.ResetScopaCount()
	}
	s.trumpCards.Replenish()
	s.trumpCards.Shuffle()
	for range ScoponeHandSize {
		for i := range ScoponePlayerCnt {
			player := s.players[(s.dealerIdx+1+i)%ScoponePlayerCnt]
			if c := s.trumpCards.DrawCard(); c != nil {
				player.AddCard(c)
			}
		}
	}
	s.currentTurn = (s.dealerIdx + 1) % ScoponePlayerCnt
	s.phase = ScoponePhasePlayerTurn
	s.appendLog(-1, "deal", fmt.Sprintf("round %d: dealt %d cards each", s.roundNumber, ScoponeHandSize), nil)
}

// PlayerPlay 人間プレイヤーが手札 handIdx を出し、tableIdxs の場札を捕獲する (空なら場に置く)。
func (s *Scopone) PlayerPlay(handIdx int, tableIdxs []int) error {
	if s.gameEndFlag {
		return ErrGameEnded
	}
	if s.phase != ScoponePhasePlayerTurn {
		return ErrWrongPhase
	}
	if !s.players[s.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return s.applyPlay(s.currentTurn, handIdx, tableIdxs)
}

// CpuPlay 現在の手番が CPU の場合に 1 手進める。
func (s *Scopone) CpuPlay() {
	if s.gameEndFlag || s.phase != ScoponePhasePlayerTurn {
		return
	}
	idx := s.currentTurn
	if s.players[idx].GetIsHuman() {
		return
	}
	handIdx, tableIdxs := s.chooseCpuPlay(idx)
	_ = s.applyPlay(idx, handIdx, tableIdxs)
}

// applyPlay は 1 手 (place / capture) を実行する共通処理。
func (s *Scopone) applyPlay(playerIdx, handIdx int, tableIdxs []int) error {
	player := s.players[playerIdx]
	if handIdx < 0 || handIdx >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", handIdx))
	}
	handCard := player.GetCard(handIdx)

	if len(tableIdxs) > 0 {
		if !isValidScopaCapture(handCard, s.tableCards, tableIdxs) {
			return NewDomainError(ErrInvalidPlay, "selected table cards do not form a valid capture")
		}
	} else if len(EnumerateScopaCaptures(handCard, s.tableCards)) > 0 {
		return NewDomainError(ErrInvalidPlay, "a capture is available and must be taken")
	}

	_ = player.RemoveCard(handIdx)

	if len(tableIdxs) == 0 {
		s.tableCards = append(s.tableCards, handCard)
		s.appendLog(playerIdx, "play", "placed on table", []*Card{handCard})
		s.postActionAdvance()
		return nil
	}

	captured := make([]*Card, 0, len(tableIdxs))
	for _, idx := range tableIdxs {
		captured = append(captured, s.tableCards[idx])
	}
	s.removeTableCardsByIndex(tableIdxs)
	pile := make([]*Card, 0, len(captured)+1)
	pile = append(pile, handCard)
	pile = append(pile, captured...)
	player.AddCaptured(pile)
	s.lastCaptureIdx = playerIdx

	if len(s.tableCards) == 0 && !s.allHandsEmpty() {
		player.IncrementScopa()
		s.appendLog(playerIdx, "scopa", "swept the table (scopa)", pile)
	} else {
		s.appendLog(playerIdx, "capture", fmt.Sprintf("captured %d card(s)", len(captured)), pile)
	}
	s.postActionAdvance()
	return nil
}

// postActionAdvance はアクション後の進行処理。
func (s *Scopone) postActionAdvance() {
	if s.allHandsEmpty() {
		s.finishRound()
		return
	}
	s.currentTurn = (s.currentTurn + 1) % ScoponePlayerCnt
}

// finishRound はラウンド終了処理: 残りの場札を最後の捕獲者へ渡し、スコア集計と終了判定。
func (s *Scopone) finishRound() {
	if s.lastCaptureIdx >= 0 && len(s.tableCards) > 0 {
		s.players[s.lastCaptureIdx].AddCaptured(s.tableCards)
		s.appendLog(s.lastCaptureIdx, "sweep_leftover", fmt.Sprintf("took %d leftover table card(s)", len(s.tableCards)), s.tableCards)
		s.tableCards = nil
	}
	det := s.scoreRound()
	s.lastRoundDetail = det
	for t := 0; t < ScoponeTeamCnt; t++ {
		s.teamScores[t] += det.Gained[t]
	}
	s.appendLog(-1, "round_score", fmt.Sprintf("round %d: team0 +%d (%d), team1 +%d (%d)",
		s.roundNumber, det.Gained[0], s.teamScores[0], det.Gained[1], s.teamScores[1]), nil)

	if s.teamScores[0] >= s.config.TargetScore || s.teamScores[1] >= s.config.TargetScore {
		s.finishGame()
		return
	}
	s.phase = ScoponePhaseRoundEnd
}

// scoreRound はチーム単位の得点内訳を計算する。
func (s *Scopone) scoreRound() *ScoponeScoreDetail {
	det := &ScoponeScoreDetail{SettebelloTm: -1}
	for i, p := range s.players {
		t := ScoponeTeamOf(i)
		det.Cards[t] += p.CapturedCount()
		det.Scopas[t] += p.GetScopaCount()
		for _, card := range p.GetCapturedCards() {
			if ScopaIsDiamond(card) {
				det.Diamonds[t]++
			}
			if ScopaIsSeven(card) {
				det.Sevens[t]++
			}
			if ScopaIsSetteBello(card) {
				det.SettebelloTm = t
			}
		}
	}
	for t := 0; t < ScoponeTeamCnt; t++ {
		score := 0
		if scoponeUniqueMax(det.Cards, t) {
			score += ScopaScoreMostCards
		}
		if scoponeUniqueMax(det.Diamonds, t) {
			score += ScopaScoreMostDiamonds
		}
		if scoponeUniqueMax(det.Sevens, t) {
			score += ScopaScoreMostSevens
		}
		if det.SettebelloTm == t {
			score += ScopaScoreSetteBello
		}
		score += det.Scopas[t] * ScopaScoreScopa
		det.Gained[t] = score
	}
	return det
}

// scoponeUniqueMax は team t が vals の中で単独最大か (同点は false)。
func scoponeUniqueMax(vals [ScoponeTeamCnt]int, t int) bool {
	other := (t + 1) % ScoponeTeamCnt
	return vals[t] > vals[other]
}

// finishGame はゲームを終了させ勝利チームを決定する。
func (s *Scopone) finishGame() {
	s.gameEndFlag = true
	s.phase = ScoponePhaseGameEnd
	switch {
	case s.teamScores[0] > s.teamScores[1]:
		s.winnerTeam = 0
	case s.teamScores[1] > s.teamScores[0]:
		s.winnerTeam = 1
	default:
		s.winnerTeam = ScoponeTeamOf(s.lastCaptureIdx) // タイブレーク: 最後に捕獲したチーム
	}
	s.appendLog(-1, "game_end", fmt.Sprintf("team %d wins (%d-%d)", s.winnerTeam, s.teamScores[0], s.teamScores[1]), nil)
}

// removeTableCardsByIndex は降順に並び替えてから tableCards を削除する。
func (s *Scopone) removeTableCardsByIndex(idxs []int) {
	for _, idx := range sortIndicesDescending(idxs) {
		if idx >= 0 && idx < len(s.tableCards) {
			s.tableCards = append(s.tableCards[:idx], s.tableCards[idx+1:]...)
		}
	}
}

// allHandsEmpty 全プレイヤーの手札が空か。
func (s *Scopone) allHandsEmpty() bool {
	return allHandsEmpty(s.players)
}

// --- CPU ---

// chooseCpuPlay は CPU の手 (handIdx, tableIdxs) を選ぶ。捕獲できるなら最大枚数 (スコパ優先)、
// 無ければ最小値カードを場に出す。
func (s *Scopone) chooseCpuPlay(playerIdx int) (int, []int) {
	p := s.players[playerIdx]
	bestHand, bestCapture := -1, []int(nil)
	bestScore := -1
	for h := 0; h < p.GetCardsSize(); h++ {
		card := p.GetCard(h)
		caps := EnumerateScopaCaptures(card, s.tableCards)
		if len(caps) == 0 {
			continue
		}
		for _, cap := range caps {
			score := len(cap)
			if len(s.tableCards) == len(cap) { // スコパ
				score += 100
			}
			if score > bestScore {
				bestScore, bestHand, bestCapture = score, h, cap
			}
		}
	}
	if bestHand >= 0 {
		return bestHand, bestCapture
	}
	// 捕獲不可: 最小値の手札を場に置く。
	dump, dumpVal := 0, 999
	for h := 0; h < p.GetCardsSize(); h++ {
		if v := ScopaCardValue(p.GetCard(h)); v < dumpVal {
			dumpVal, dump = v, h
		}
	}
	return dump, nil
}

// --- 状態アクセサ ---

// GetPhase 現在のフェーズ取得
func (s *Scopone) GetPhase() string { return s.phase }

// SetPhase フェーズ設定 (テスト用)
func (s *Scopone) SetPhase(phase string) { s.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (s *Scopone) GetRoundNumber() int { return s.roundNumber }

// GetCurrentTurn 現在の手番プレイヤーインデックス取得
func (s *Scopone) GetCurrentTurn() int { return s.currentTurn }

// SetCurrentTurn 手番設定 (テスト用)
func (s *Scopone) SetCurrentTurn(idx int) { s.currentTurn = idx }

// GetDealerIdx ディーラーインデックス取得
func (s *Scopone) GetDealerIdx() int { return s.dealerIdx }

// GetTableCards 場札取得
func (s *Scopone) GetTableCards() []*Card { return s.tableCards }

// SetTableCards 場札設定 (テスト用)
func (s *Scopone) SetTableCards(cards []*Card) { s.tableCards = cards }

// GetLastCaptureIdx 最後に捕獲したプレイヤー取得 (-1: なし)
func (s *Scopone) GetLastCaptureIdx() int { return s.lastCaptureIdx }

// GetTeamScore チームスコア取得
func (s *Scopone) GetTeamScore(team int) int {
	if team < 0 || team >= ScoponeTeamCnt {
		return 0
	}
	return s.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (s *Scopone) SetTeamScore(team, score int) {
	if team >= 0 && team < ScoponeTeamCnt {
		s.teamScores[team] = score
	}
}

// GetGameEndFlag ゲーム終了フラグ取得
func (s *Scopone) GetGameEndFlag() bool { return s.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1: 未確定)
func (s *Scopone) GetWinnerTeam() int { return s.winnerTeam }

// GetLastRoundDetail 直近ラウンドの得点内訳取得
func (s *Scopone) GetLastRoundDetail() *ScoponeScoreDetail { return s.lastRoundDetail }

// GetPlayerCnt プレイヤー数取得
func (s *Scopone) GetPlayerCnt() int { return len(s.players) }

// GetPlayer プレイヤー取得
func (s *Scopone) GetPlayer(i int) *ScopaPlayer {
	return getPlayer(s.players, i)
}

// GetConfig 設定取得
func (s *Scopone) GetConfig() ScoponeConfig { return s.config }

// SetConfig 設定変更
func (s *Scopone) SetConfig(cfg ScoponeConfig) { s.config = cfg }

// IsHumanTurn 現在の手番が人間かどうか
func (s *Scopone) IsHumanTurn() bool {
	if s.gameEndFlag || s.currentTurn < 0 || s.currentTurn >= len(s.players) {
		return false
	}
	return s.players[s.currentTurn].GetIsHuman()
}

// GetValidCaptures は handIdx のカードで可能な捕獲パターン (場札インデックス集合) を返す。
func (s *Scopone) GetValidCaptures(handIdx int) [][]int {
	if s.currentTurn < 0 || s.currentTurn >= len(s.players) {
		return nil
	}
	p := s.players[s.currentTurn]
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return nil
	}
	return EnumerateScopaCaptures(p.GetCard(handIdx), s.tableCards)
}

// --- JSON ---

type scoponeJSON struct {
	TrumpCards      *TrumpCards         `json:"tc"`
	Players         []*ScopaPlayer      `json:"ps"`
	Config          ScoponeConfig       `json:"cf"`
	Phase           string              `json:"ph"`
	RoundNumber     int                 `json:"rn"`
	CurrentTurn     int                 `json:"ct"`
	DealerIdx       int                 `json:"di"`
	TableCards      []*Card             `json:"tb"`
	LastCaptureIdx  int                 `json:"lc"`
	TeamScores      [ScoponeTeamCnt]int `json:"sc"`
	GameEndFlag     bool                `json:"ge"`
	WinnerTeam      int                 `json:"wt"`
	LastRoundDetail *ScoponeScoreDetail `json:"rd"`
	ActionLog       []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (s *Scopone) MarshalJSON() ([]byte, error) {
	return json.Marshal(scoponeJSON{
		TrumpCards:      s.trumpCards,
		Players:         s.players,
		Config:          s.config,
		Phase:           s.phase,
		RoundNumber:     s.roundNumber,
		CurrentTurn:     s.currentTurn,
		DealerIdx:       s.dealerIdx,
		TableCards:      s.tableCards,
		LastCaptureIdx:  s.lastCaptureIdx,
		TeamScores:      s.teamScores,
		GameEndFlag:     s.gameEndFlag,
		WinnerTeam:      s.winnerTeam,
		LastRoundDetail: s.lastRoundDetail,
		ActionLog:       s.actionLog,
	})
}

const scoponeMaxSliceLen = 5000

var errScoponeSnapshot = errors.New("scopone: invalid serialised game state")

func scoponeIdxInRange(i int) bool { return i >= 0 && i < ScoponePlayerCnt }

// UnmarshalJSON implements json.Unmarshaler.
func (s *Scopone) UnmarshalJSON(data []byte) error {
	var j scoponeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != ScoponePlayerCnt ||
		len(j.TableCards) > scoponeMaxSliceLen || len(j.ActionLog) > scoponeMaxSliceLen ||
		!scoponeIdxInRange(j.CurrentTurn) || !scoponeIdxInRange(j.DealerIdx) ||
		j.LastCaptureIdx < -1 || j.LastCaptureIdx >= ScoponePlayerCnt ||
		j.WinnerTeam < -1 || j.WinnerTeam >= ScoponeTeamCnt ||
		j.RoundNumber < 1 ||
		(j.Phase != ScoponePhasePlayerTurn && j.Phase != ScoponePhaseRoundEnd && j.Phase != ScoponePhaseGameEnd) {
		return errScoponeSnapshot
	}
	for _, p := range j.Players {
		if p == nil {
			return errScoponeSnapshot
		}
	}
	for _, c := range j.TableCards {
		if c == nil {
			return errScoponeSnapshot
		}
	}
	for _, entry := range j.ActionLog {
		if entry == nil {
			return errScoponeSnapshot
		}
	}
	s.trumpCards = j.TrumpCards
	if s.trumpCards == nil {
		s.trumpCards = NewTrumpCardsScopa()
	}
	s.players = j.Players
	s.config = j.Config
	s.phase = j.Phase
	s.roundNumber = j.RoundNumber
	s.currentTurn = j.CurrentTurn
	s.dealerIdx = j.DealerIdx
	s.tableCards = j.TableCards
	if s.tableCards == nil {
		s.tableCards = make([]*Card, 0)
	}
	s.lastCaptureIdx = j.LastCaptureIdx
	s.teamScores = j.TeamScores
	s.gameEndFlag = j.GameEndFlag
	s.winnerTeam = j.WinnerTeam
	s.lastRoundDetail = j.LastRoundDetail
	s.actionLog = j.ActionLog
	if s.actionLog == nil {
		s.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
