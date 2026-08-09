//go:build !js || !wasm || classic

// Package domain スコパ (Scopa) のドメインモデル。
//
// Scopa はイタリアの国民的フィッシング系カードゲーム。40 枚デッキ
// (8,9,10 を除く) を使い、本実装では 2 人対戦 (ヘッズアップ) を扱う。
// 各プレイヤーは手札 1 枚を出し、場札と同じ値、または合計値が一致する
// 場札の組合せを捕獲する。場札をすべて取り切ると「スコパ」となりボーナス。
// ラウンド終了時に 最多カード (carte) / 最多ダイヤ (denari) / 7♦ (settebello)
// / 最多 7 (簡易プリミエラ) / スコパ回数 を集計し、先に 11 点へ到達した側が勝利。
package domain

import (
	"encoding/json"
	"fmt"
)

// ScopaPlayerCnt スコパのプレイヤー数 (ヘッズアップ固定)
const ScopaPlayerCnt = 2

// ScopaHandSize 1 回の配札で各プレイヤーに配る枚数
const ScopaHandSize = 3

// ScopaInitialTableSize 最初のラウンド開始時に場へ置くカード枚数
const ScopaInitialTableSize = 4

// Scopa ゲームのフェーズ
const (
	// ScopaPhaseDealing 配札中
	ScopaPhaseDealing = "dealing"
	// ScopaPhasePlayerTurn プレイヤーターン (人間または CPU)
	ScopaPhasePlayerTurn = "playerTurn"
	// ScopaPhaseRoundEnd ラウンド終了 (捕獲の締め処理)
	ScopaPhaseRoundEnd = "roundEnd"
	// ScopaPhaseGameEnd ゲーム終了 (誰かが TargetScore に到達)
	ScopaPhaseGameEnd = "gameEnd"
)

// ScopaAction はプレイヤー 1 ターン分の行動記録。
type ScopaAction struct {
	PlayerIdx     int     // 行動したプレイヤーインデックス
	PlayedCard    *Card   // 出した手札 1 枚
	CapturedCards []*Card // 捕獲した場札 (捕獲なしの場合は空)
	IsScopa       bool    // スコパ (場全取り) 発生
}

// scopaRoundState はラウンドごとにリセットされる状態。
type scopaRoundState struct {
	phase          string         // 現在のフェーズ
	currentTurn    int            // 現在の手番
	tableCards     []*Card        // 場札
	lastCaptureIdx int            // 最後に捕獲したプレイヤー (-1 = なし)
	humanAction    *ScopaAction   // 人間の最後の行動
	cpuActions     []*ScopaAction // 人間ターン後の CPU 行動履歴
	actionLogBase
	packsDealt      int  // これまでに配ったパック数 (1 回の配布 = 3 枚/人)
	gameEndFlag     bool // ゲーム終了フラグ (TargetScore 到達)
	roundWinners    []int
	lastRoundScores map[int]int // 前ラウンドの内訳得点 (プレイヤー別)
	lastRoundDetail *ScopaScoreDetail
}

// ScopaScoreDetail は 1 ラウンドの得点内訳。
type ScopaScoreDetail struct {
	Cards         map[int]int // プレイヤー別捕獲枚数
	Diamonds      map[int]int // プレイヤー別ダイヤ数
	Sevens        map[int]int // プレイヤー別 7 の数
	HasSetteBello int         // 7♦ を取ったプレイヤー (-1 = なし)
	Scopas        map[int]int // プレイヤー別スコパ数
	Gained        map[int]int // プレイヤー別にこのラウンドで得た点数
}

// Scopa はスコパゲームの状態を保持する集約ルート。
type Scopa struct {
	trumpCards *TrumpCards
	players    []*ScopaPlayer
	config     ScopaConfig
	round      scopaRoundState
}

// NewScopa コンストラクタ。
func NewScopa(trumpCards *TrumpCards, players []*ScopaPlayer, config ScopaConfig) *Scopa {
	return &Scopa{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		round: scopaRoundState{
			phase:          ScopaPhaseDealing,
			lastCaptureIdx: -1,
		},
	}
}

// NewDefaultScopa returns a Scopa with the standard 2-player setup
// (1 human, 1 CPU) and DefaultScopaConfig.
func NewDefaultScopa() *Scopa {
	config := DefaultScopaConfig()
	players := make([]*ScopaPlayer, ScopaPlayerCnt)
	players[0] = NewScopaPlayer(true)
	for i := 1; i < ScopaPlayerCnt; i++ {
		players[i] = NewScopaPlayer(false)
	}
	return NewScopa(NewTrumpCardsScopa(), players, config)
}

// Reset は新しい「ゲーム」を開始する。累計得点もクリアする。
func (s *Scopa) Reset() {
	for _, p := range s.players {
		p.Reset()
		p.SetIsFinished(false)
		p.ResetCaptured()
		p.ResetScopaCount()
		p.ResetTotalScore()
	}

	s.trumpCards = NewTrumpCardsScopa()
	s.trumpCards.Shuffle()

	s.round = scopaRoundState{
		phase:          ScopaPhaseDealing,
		lastCaptureIdx: -1,
		actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}

	s.startRound()
}

// NextRound は次のラウンドを開始する。
// 捕獲札、スコパ数はラウンドごとにクリア (累計得点は維持)。
func (s *Scopa) NextRound() {
	if s.round.gameEndFlag {
		return
	}
	for _, p := range s.players {
		p.Reset()
		p.SetIsFinished(false)
		p.ResetCaptured()
		p.ResetScopaCount()
	}
	s.trumpCards = NewTrumpCardsScopa()
	s.trumpCards.Shuffle()
	s.round.phase = ScopaPhaseDealing
	s.round.currentTurn = 0
	s.round.tableCards = nil
	s.round.lastCaptureIdx = -1
	s.round.humanAction = nil
	s.round.cpuActions = nil
	s.round.packsDealt = 0
	s.startRound()
}

// startRound は新しい山札の配札と初期場札 (4 枚) を設定する。
// スコパでは新しいディール (新しい山札) のたびに場へ 4 枚を表向きで置く。
func (s *Scopa) startRound() {
	s.dealNextPack()
	for i := 0; i < ScopaInitialTableSize; i++ {
		card := s.trumpCards.DrawCard()
		if card == nil {
			break
		}
		s.round.tableCards = append(s.round.tableCards, card)
	}
	s.appendLog(-1, "deal", fmt.Sprintf("dealt %d table cards", len(s.round.tableCards)), s.round.tableCards)
	s.round.phase = ScopaPhasePlayerTurn
}

// dealNextPack は各プレイヤーに ScopaHandSize 枚配る。
// 山札が尽きたら部分的に配って終わる。
func (s *Scopa) dealNextPack() {
	for k := 0; k < ScopaHandSize; k++ {
		for i := 0; i < len(s.players); i++ {
			card := s.trumpCards.DrawCard()
			if card == nil {
				s.round.packsDealt++
				return
			}
			s.players[i].AddCard(card)
		}
	}
	s.round.packsDealt++
}

// allHandsEmpty は全員の手札が空か。
func (s *Scopa) allHandsEmpty() bool {
	return allHandsEmpty(s.players)
}

// PlayerPlay は人間プレイヤーが手札 handIdx を出す。
// tableIdxs が空の場合は「場に置く」。ただし捕獲可能な手なら捕獲が必須。
func (s *Scopa) PlayerPlay(handIdx int, tableIdxs []int) error {
	if err := s.guardHumanTurn(); err != nil {
		return err
	}
	s.round.cpuActions = nil
	return s.applyPlay(s.round.currentTurn, handIdx, tableIdxs, s.setHumanAction)
}

// CpuPlay は CPU のターンを 1 回進める。
func (s *Scopa) CpuPlay() {
	if s.round.gameEndFlag || s.round.phase != ScopaPhasePlayerTurn {
		return
	}
	if s.players[s.round.currentTurn].GetIsHuman() {
		return
	}
	playerIdx := s.round.currentTurn
	plan := s.chooseCpuPlay(playerIdx)
	_ = s.applyPlay(playerIdx, plan.handIdx, plan.tableIdxs, s.appendCpuAction)
}

// guardHumanTurn は人間ターンかつゲーム進行中か確認。
func (s *Scopa) guardHumanTurn() error {
	if s.round.gameEndFlag {
		return ErrGameEnded
	}
	if s.round.phase != ScopaPhasePlayerTurn {
		return NewDomainError(ErrWrongPhase, "not in player turn phase")
	}
	if !s.players[s.round.currentTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return nil
}

// setHumanAction は human の action を記録。
func (s *Scopa) setHumanAction(a *ScopaAction) { s.round.humanAction = a }

// appendCpuAction は cpu の action を記録。
func (s *Scopa) appendCpuAction(a *ScopaAction) {
	s.round.cpuActions = append(s.round.cpuActions, a)
}

// applyPlay は 1 手 (play / capture) を実行する共通処理。
func (s *Scopa) applyPlay(playerIdx, handIdx int, tableIdxs []int, record func(*ScopaAction)) error {
	player := s.players[playerIdx]
	handCard := player.GetCard(handIdx)
	if handCard == nil {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", handIdx))
	}

	if len(tableIdxs) > 0 {
		if !isValidScopaCapture(handCard, s.round.tableCards, tableIdxs) {
			return NewDomainError(ErrInvalidPlay, "selected table cards do not form a valid capture")
		}
	} else if len(EnumerateScopaCaptures(handCard, s.round.tableCards)) > 0 {
		// 捕獲可能なときは捕獲が必須 (場に置けない)。
		return NewDomainError(ErrInvalidPlay, "a capture is available and must be taken")
	}

	_ = player.RemoveCard(handIdx)

	if len(tableIdxs) == 0 {
		// 捕獲なし: 場に置く。
		s.round.tableCards = append(s.round.tableCards, handCard)
		action := &ScopaAction{PlayerIdx: playerIdx, PlayedCard: handCard}
		record(action)
		s.appendLog(playerIdx, "play", "placed on table", []*Card{handCard})
		s.postActionAdvance()
		return nil
	}

	// 捕獲。
	captured := make([]*Card, 0, len(tableIdxs))
	for _, idx := range tableIdxs {
		captured = append(captured, s.round.tableCards[idx])
	}
	s.removeTableCardsByIndex(tableIdxs)

	pile := make([]*Card, 0, len(captured)+1)
	pile = append(pile, handCard)
	pile = append(pile, captured...)
	player.AddCaptured(pile)
	s.round.lastCaptureIdx = playerIdx

	isScopa := len(s.round.tableCards) == 0 && !s.isLastPlayOfRound()
	if isScopa {
		player.IncrementScopa()
	}

	action := &ScopaAction{
		PlayerIdx:     playerIdx,
		PlayedCard:    handCard,
		CapturedCards: captured,
		IsScopa:       isScopa,
	}
	record(action)
	s.appendLog(playerIdx, "capture", fmt.Sprintf("captured %d card(s)", len(captured)), pile)
	s.postActionAdvance()
	return nil
}

// postActionAdvance はアクション後の共通進行処理。
func (s *Scopa) postActionAdvance() {
	if s.isRoundOver() {
		s.finishRound()
		return
	}
	s.round.currentTurn = (s.round.currentTurn + 1) % len(s.players)
	// 手札切れ + 山札あり → 全員へ次のパックを配る。
	if s.allHandsEmpty() && s.trumpCards.GetRemainingCount() > 0 {
		s.dealNextPack()
	}
}

// isRoundOver は現在のラウンドが終了しているか (手札 0 + 山札 0)。
func (s *Scopa) isRoundOver() bool {
	return s.allHandsEmpty() && s.trumpCards.GetRemainingCount() == 0
}

// isLastPlayOfRound は今の捕獲がラウンド最後の 1 手か。
// ラウンド末の場全取りはスコパに数えないため、この判定で弾く。
func (s *Scopa) isLastPlayOfRound() bool {
	return s.allHandsEmpty() && s.trumpCards.GetRemainingCount() == 0
}

// finishRound はラウンド終了処理: 残りの場札を最後に捕獲したプレイヤーに渡し、スコア計算。
func (s *Scopa) finishRound() {
	s.round.phase = ScopaPhaseRoundEnd
	leftover := make([]*Card, 0, len(s.round.tableCards))
	leftover = append(leftover, s.round.tableCards...)
	s.round.tableCards = nil
	if s.round.lastCaptureIdx >= 0 && len(leftover) > 0 {
		s.players[s.round.lastCaptureIdx].AddCaptured(leftover)
		s.appendLog(s.round.lastCaptureIdx, "lastTake", fmt.Sprintf("last-take: %d card(s)", len(leftover)), leftover)
	}

	detail := s.scoreRound()
	s.round.lastRoundDetail = detail
	s.round.lastRoundScores = detail.Gained
	for i, p := range s.players {
		p.AddScore(detail.Gained[i])
	}

	maxScore := 0
	for _, p := range s.players {
		if p.GetTotalScore() > maxScore {
			maxScore = p.GetTotalScore()
		}
	}
	if maxScore >= s.config.TargetScore {
		s.round.gameEndFlag = true
		s.round.phase = ScopaPhaseGameEnd
		winners := make([]int, 0)
		for i, p := range s.players {
			if p.GetTotalScore() == maxScore {
				winners = append(winners, i)
			}
		}
		s.round.roundWinners = winners
		s.appendLog(-1, "gameEnd", fmt.Sprintf("game ended at %d points", maxScore), nil)
	} else {
		s.appendLog(-1, "roundEnd", "round ended", nil)
	}
}

// scoreRound はラウンドの得点内訳を計算する。
func (s *Scopa) scoreRound() *ScopaScoreDetail {
	det := &ScopaScoreDetail{
		Cards:         make(map[int]int),
		Diamonds:      make(map[int]int),
		Sevens:        make(map[int]int),
		Scopas:        make(map[int]int),
		Gained:        make(map[int]int),
		HasSetteBello: -1,
	}
	for i, p := range s.players {
		det.Cards[i] = p.CapturedCount()
		det.Scopas[i] = p.GetScopaCount()
		for _, card := range p.GetCapturedCards() {
			if ScopaIsDiamond(card) {
				det.Diamonds[i]++
			}
			if ScopaIsSeven(card) {
				det.Sevens[i]++
			}
			if ScopaIsSetteBello(card) {
				det.HasSetteBello = i
			}
		}
	}
	mostCardsIdx := uniqueMaxIndex(det.Cards)
	mostDiamondsIdx := uniqueMaxIndex(det.Diamonds)
	mostSevensIdx := uniqueMaxIndex(det.Sevens)
	for i := range s.players {
		score := 0
		if i == mostCardsIdx {
			score += ScopaScoreMostCards
		}
		if i == mostDiamondsIdx {
			score += ScopaScoreMostDiamonds
		}
		if i == mostSevensIdx {
			score += ScopaScoreMostSevens
		}
		if det.HasSetteBello == i {
			score += ScopaScoreSetteBello
		}
		score += det.Scopas[i] * ScopaScoreScopa
		det.Gained[i] = score
	}
	return det
}

// ScopaCategoryWinner は1カテゴリの勝者と得点。
type ScopaCategoryWinner struct {
	// Key はカテゴリ ("cards" / "denari" / "primiera" / "settebello")。
	Key string
	// Winner は勝者のインデックス。同点・該当なしは -1。
	Winner int
	// Points は与えられる点数 (該当なしは 0)。
	Points int
}

// ScopaCategoryWinners は得点内訳のうち「単独勝者が決まる4カテゴリ」を返す。
//
// **CUI はなぜその点数になったのかを一切出していなかった (#4756)。**Web は
// カルテ/デナリ/プリミエラ/セッテベッロごとに誰が取ったかを出している。
//
// 判定は得点計算そのものと同じ uniqueMaxIndex を通す。**別実装にすると、
// 内訳の合計が実際の得点と合わなくなる。**
func ScopaCategoryWinners(det *ScopaScoreDetail) []ScopaCategoryWinner {
	if det == nil {
		return nil
	}
	award := func(key string, winner, points int) ScopaCategoryWinner {
		if winner < 0 {
			return ScopaCategoryWinner{Key: key, Winner: -1}
		}
		return ScopaCategoryWinner{Key: key, Winner: winner, Points: points}
	}
	return []ScopaCategoryWinner{
		award("cards", uniqueMaxIndex(det.Cards), ScopaScoreMostCards),
		award("denari", uniqueMaxIndex(det.Diamonds), ScopaScoreMostDiamonds),
		award("primiera", uniqueMaxIndex(det.Sevens), ScopaScoreMostSevens),
		award("settebello", det.HasSetteBello, ScopaScoreSetteBello),
	}
}

// removeTableCardsByIndex は降順に並び替えてから tableCards を削除する。
func (s *Scopa) removeTableCardsByIndex(idxs []int) {
	s.round.tableCards = removeIndices(s.round.tableCards, idxs)
}

// appendLog 棋譜にエントリを追加する。
func (s *Scopa) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	s.round.appendLog(playerIdx, actionType, detail, cards)
}

// --- 状態アクセサ ---

// IsHumanTurn 現在の手番が人間かどうか
func (s *Scopa) IsHumanTurn() bool {
	if s.round.gameEndFlag {
		return false
	}
	return s.players[s.round.currentTurn].GetIsHuman()
}

// GetCurrentTurn 現在の手番プレイヤーインデックス取得
func (s *Scopa) GetCurrentTurn() int { return s.round.currentTurn }

// GetGameEndFlag ゲーム終了フラグ取得
func (s *Scopa) GetGameEndFlag() bool { return s.round.gameEndFlag }

// GetTableCards 場札取得。
func (s *Scopa) GetTableCards() []*Card { return s.round.tableCards }

// GetPlayer プレイヤー取得
func (s *Scopa) GetPlayer(idx int) *ScopaPlayer {
	return getPlayer(s.players, idx)
}

// GetPlayerCnt プレイヤー数取得
func (s *Scopa) GetPlayerCnt() int { return len(s.players) }

// GetCpuActions CPU ターンの行動履歴取得
func (s *Scopa) GetCpuActions() []*ScopaAction { return s.round.cpuActions }

// GetHumanAction 人間の最後の行動取得
func (s *Scopa) GetHumanAction() *ScopaAction { return s.round.humanAction }

// GetConfig ローカルルール設定取得
func (s *Scopa) GetConfig() ScopaConfig { return s.config }

// SetConfig ローカルルール設定を変更
func (s *Scopa) SetConfig(config ScopaConfig) { s.config = config }

// GetActionLog 棋譜取得
func (s *Scopa) GetActionLog() []*ActionLogEntry { return s.round.actionLog }

// GetPhase 現在のフェーズ取得
func (s *Scopa) GetPhase() string { return s.round.phase }

// GetLastRoundDetail 直前ラウンドの得点詳細取得 (nil の場合もあり得る)
func (s *Scopa) GetLastRoundDetail() *ScopaScoreDetail { return s.round.lastRoundDetail }

// GetLastCaptureIdx 最後に捕獲したプレイヤー (-1 = なし)
func (s *Scopa) GetLastCaptureIdx() int { return s.round.lastCaptureIdx }

// GetRoundWinners ゲーム終了時の勝者リスト。
func (s *Scopa) GetRoundWinners() []int { return s.round.roundWinners }

// GetRemainingDeck 山札の残り枚数。
func (s *Scopa) GetRemainingDeck() int { return s.trumpCards.GetRemainingCount() }

// GetPacksDealt これまでに配布されたパック数。
func (s *Scopa) GetPacksDealt() int { return s.round.packsDealt }

// --- JSON Serialization ---

// scopaActionJSON is the JSON wire format for ScopaAction.
type scopaActionJSON struct {
	PlayerIdx     int     `json:"pi"`
	PlayedCard    *Card   `json:"pc"`
	CapturedCards []*Card `json:"cc"`
	IsScopa       bool    `json:"sp"`
}

// MarshalJSON implements json.Marshaler.
func (a *ScopaAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(scopaActionJSON{
		PlayerIdx:     a.PlayerIdx,
		PlayedCard:    a.PlayedCard,
		CapturedCards: a.CapturedCards,
		IsScopa:       a.IsScopa,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *ScopaAction) UnmarshalJSON(data []byte) error {
	var j scopaActionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	a.PlayerIdx = j.PlayerIdx
	a.PlayedCard = j.PlayedCard
	a.CapturedCards = j.CapturedCards
	a.IsScopa = j.IsScopa
	return nil
}

// scopaScoreDetailJSON is the JSON wire format for ScopaScoreDetail.
type scopaScoreDetailJSON struct {
	Cards         map[int]int `json:"cd"`
	Diamonds      map[int]int `json:"dm"`
	Sevens        map[int]int `json:"sv"`
	HasSetteBello int         `json:"sb"`
	Scopas        map[int]int `json:"sp"`
	Gained        map[int]int `json:"gn"`
}

// MarshalJSON implements json.Marshaler.
func (d *ScopaScoreDetail) MarshalJSON() ([]byte, error) {
	return json.Marshal(scopaScoreDetailJSON{
		Cards:         d.Cards,
		Diamonds:      d.Diamonds,
		Sevens:        d.Sevens,
		HasSetteBello: d.HasSetteBello,
		Scopas:        d.Scopas,
		Gained:        d.Gained,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *ScopaScoreDetail) UnmarshalJSON(data []byte) error {
	var j scopaScoreDetailJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	d.Cards = j.Cards
	d.Diamonds = j.Diamonds
	d.Sevens = j.Sevens
	d.HasSetteBello = j.HasSetteBello
	d.Scopas = j.Scopas
	d.Gained = j.Gained
	return nil
}

// scopaJSON is the JSON wire format for Scopa.
type scopaJSON struct {
	TrumpCards      *TrumpCards       `json:"tc"`
	Players         []*ScopaPlayer    `json:"pl"`
	Config          ScopaConfig       `json:"cf"`
	Phase           string            `json:"ph"`
	CurrentTurn     int               `json:"ct"`
	TableCards      []*Card           `json:"tb"`
	LastCaptureIdx  int               `json:"lc"`
	HumanAction     *ScopaAction      `json:"ha"`
	CpuActions      []*ScopaAction    `json:"ca"`
	ActionLog       []*ActionLogEntry `json:"al"`
	PacksDealt      int               `json:"pd"`
	GameEndFlag     bool              `json:"ge"`
	RoundWinners    []int             `json:"rw"`
	LastRoundScores map[int]int       `json:"ls"`
	LastRoundDetail *ScopaScoreDetail `json:"ld"`
}

// scopaMaxSliceLen caps slice sizes during deserialisation.
const scopaMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (s *Scopa) MarshalJSON() ([]byte, error) {
	return json.Marshal(scopaJSON{
		TrumpCards:      s.trumpCards,
		Players:         s.players,
		Config:          s.config,
		Phase:           s.round.phase,
		CurrentTurn:     s.round.currentTurn,
		TableCards:      s.round.tableCards,
		LastCaptureIdx:  s.round.lastCaptureIdx,
		HumanAction:     s.round.humanAction,
		CpuActions:      s.round.cpuActions,
		ActionLog:       s.round.actionLog,
		PacksDealt:      s.round.packsDealt,
		GameEndFlag:     s.round.gameEndFlag,
		RoundWinners:    s.round.roundWinners,
		LastRoundScores: s.round.lastRoundScores,
		LastRoundDetail: s.round.lastRoundDetail,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *Scopa) UnmarshalJSON(data []byte) error {
	var j scopaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > scopaMaxSliceLen || len(j.TableCards) > scopaMaxSliceLen ||
		len(j.CpuActions) > scopaMaxSliceLen || len(j.ActionLog) > scopaMaxSliceLen {
		return fmt.Errorf("scopa: input array exceeds maximum allowed size")
	}
	if j.TrumpCards == nil {
		return fmt.Errorf("scopa: missing trump cards in state")
	}
	s.trumpCards = j.TrumpCards
	s.players = j.Players
	if s.players == nil {
		s.players = make([]*ScopaPlayer, 0)
	}
	s.config = j.Config
	s.round = scopaRoundState{
		phase:           j.Phase,
		currentTurn:     j.CurrentTurn,
		tableCards:      j.TableCards,
		lastCaptureIdx:  j.LastCaptureIdx,
		humanAction:     j.HumanAction,
		cpuActions:      j.CpuActions,
		actionLogBase:   actionLogBase{actionLog: j.ActionLog},
		packsDealt:      j.PacksDealt,
		gameEndFlag:     j.GameEndFlag,
		roundWinners:    j.RoundWinners,
		lastRoundScores: j.LastRoundScores,
		lastRoundDetail: j.LastRoundDetail,
	}
	if s.round.actionLog == nil {
		s.round.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
