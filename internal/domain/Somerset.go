//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// SomersetPhase Somerset game phase
type SomersetPhase int

// Somerset のフェーズ定数
const (
	// SomersetPhasePlaying プレイ中
	SomersetPhasePlaying SomersetPhase = iota
	// SomersetPhaseGameClear ゲームクリア
	SomersetPhaseGameClear
	// SomersetPhaseGameOver ゲームオーバー
	SomersetPhaseGameOver
)

// SomersetTableauCnt タブローの列数 (左右4列ずつ計8列)
const SomersetTableauCnt = 10

// SomersetColumnLen 初期配置時の各列のカード枚数
// SomersetColumnLen is the longest column after the deal: 52 = 5*10 + 2, so
// two columns hold 6 and the other eight hold 5. Used only as a capacity hint.
const SomersetColumnLen = 6

// SomersetFoundationCnt ファンデーションの数
const SomersetFoundationCnt = 4

// SomersetTableauCard タブロー上のカード
type SomersetTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// SomersetHint ヒント
type SomersetHint struct {
	FromCol   int    // タブロー列インデックス
	CardIndex int    // 列内のカードインデックス
	ToZone    string // "tableau" or "foundation"
	ToCol     int    // タブロー列 or ファンデーションのインデックス
}

// SomersetConfig サマセットのゲーム設定
type SomersetConfig struct{}

// Somerset ゲームクラス
type Somerset struct {
	trumpCards *TrumpCards
	tableau    [SomersetTableauCnt][]*SomersetTableauCard
	foundation [SomersetFoundationCnt][]*Card
	phase      SomersetPhase
	moveCount  int
	actionLogBase
	history     []*somersetSnapshot
	isStalemate bool
}

// somersetSnapshot アンドゥ用スナップショット
type somersetSnapshot struct {
	tableau     [SomersetTableauCnt][]*SomersetTableauCard
	foundation  [SomersetFoundationCnt][]*Card
	phase       SomersetPhase
	moveCount   int
	isStalemate bool
}

// NewSomerset コンストラクタ
func NewSomerset(trumpCards *TrumpCards) *Somerset {
	return &Somerset{
		trumpCards: trumpCards,
	}
}

// NewDefaultSomerset returns Somerset with a single 52-card deck.
func NewDefaultSomerset() *Somerset {
	return NewSomerset(NewTrumpCardsWithDecks(1, 0))
}

// Reset ゲームリセット
//
// Initial layout:
//   - 4 Aces are pre-seeded onto the four foundations (one per suit).
//   - The remaining 48 cards are dealt face-up across 8 tableau columns
//     (6 cards each).
func (bc *Somerset) Reset() {
	bc.trumpCards.Shuffle()
	bc.phase = SomersetPhasePlaying
	bc.moveCount = 0
	bc.actionLog = nil
	bc.history = nil
	bc.isStalemate = false

	for i := range SomersetFoundationCnt {
		bc.foundation[i] = nil
	}
	for i := range SomersetTableauCnt {
		bc.tableau[i] = make([]*SomersetTableauCard, 0, SomersetColumnLen)
	}

	// Somerset deals the ENTIRE deck to the tableau. Beleaguered Castle, which
	// this domain was cloned from, first pulls the four Aces onto the
	// foundations; Somerset starts with empty foundations and the player must
	// dig each Ace out. Round-robin over 10 columns: 52 = 5*10 + 2, so the
	// first two columns end up with 6 cards and the rest with 5.
	col := 0
	for {
		card := bc.trumpCards.DrawCard()
		if card == nil {
			break
		}
		bc.tableau[col] = append(bc.tableau[col], &SomersetTableauCard{Card: card, FaceUp: true})
		col = (col + 1) % SomersetTableauCnt
	}

	bc.checkStalemate()
}

// MoveTableauToTableau タブローからタブローにカードを移動（1枚のみ）
func (bc *Somerset) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if bc.phase != SomersetPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= SomersetTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= SomersetTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := bc.tableau[fromCol]
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	if cardIndex != len(fromCards)-1 {
		return errors.New("only the top card can be moved")
	}
	tc := fromCards[cardIndex]
	if !bc.canPlaceOnTableau(tc.Card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	bc.takeSnapshot()
	bc.tableau[toCol] = append(bc.tableau[toCol], tc)
	bc.tableau[fromCol] = fromCards[:cardIndex]
	bc.moveCount++
	bc.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), []*Card{tc.Card})
	bc.checkStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (bc *Somerset) MoveTableauToFoundation(col int) error {
	if bc.phase != SomersetPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= SomersetTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := bc.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	card := tc.Card
	fIdx := bc.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	bc.takeSnapshot()
	bc.tableau[col] = fromCards[:len(fromCards)-1]
	bc.foundation[fIdx] = append(bc.foundation[fIdx], card)
	bc.moveCount++
	bc.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	bc.checkGameClear()
	bc.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (bc *Somerset) GiveUp() {
	if bc.phase == SomersetPhasePlaying {
		bc.phase = SomersetPhaseGameOver
		bc.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (bc *Somerset) GetHint() *SomersetHint {
	if bc.phase != SomersetPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := range SomersetTableauCnt {
		if len(bc.tableau[col]) == 0 {
			continue
		}
		tc := bc.tableau[col][len(bc.tableau[col])-1]
		fIdx := bc.findFoundation(tc.Card)
		if fIdx >= 0 {
			return &SomersetHint{
				FromCol:   col,
				CardIndex: len(bc.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: タブローからタブローへ
	for fromCol := range SomersetTableauCnt {
		fromCards := bc.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		card := fromCards[len(fromCards)-1].Card
		for toCol := range SomersetTableauCnt {
			if toCol == fromCol {
				continue
			}
			if bc.canPlaceOnTableau(card, toCol) {
				return &SomersetHint{
					FromCol:   fromCol,
					CardIndex: len(fromCards) - 1,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	return nil
}

// AutoComplete オートコンプリート（全ての山から可能な限りファンデーションへ）
func (bc *Somerset) AutoComplete() error {
	if bc.phase != SomersetPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	bc.takeSnapshot()
	for {
		moved := false
		for col := range SomersetTableauCnt {
			if len(bc.tableau[col]) == 0 {
				continue
			}
			tc := bc.tableau[col][len(bc.tableau[col])-1]
			card := tc.Card
			fIdx := bc.findFoundation(card)
			if fIdx < 0 {
				continue
			}
			bc.tableau[col] = bc.tableau[col][:len(bc.tableau[col])-1]
			bc.foundation[fIdx] = append(bc.foundation[fIdx], card)
			bc.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	bc.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	bc.checkGameClear()
	bc.checkStalemate()
	return nil
}

// AllFaceUp 全カードが表向きかどうか（Somerset は全札を表向きに配るので常にtrue）
func (bc *Somerset) AllFaceUp() bool {
	return true
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (bc *Somerset) GetPhase() SomersetPhase { return bc.phase }

// SetPhase フェーズ設定 (テスト用)
func (bc *Somerset) SetPhase(phase SomersetPhase) { bc.phase = phase }

// GetMoveCount 移動回数取得
func (bc *Somerset) GetMoveCount() int { return bc.moveCount }

// GetTableau タブロー取得
func (bc *Somerset) GetTableau() [SomersetTableauCnt][]*SomersetTableauCard {
	return bc.tableau
}

// GetFoundation ファンデーション取得
func (bc *Somerset) GetFoundation() [SomersetFoundationCnt][]*Card {
	return bc.foundation
}

// GetGameEndFlag returns true once the game has left the playing phase.
func (bc *Somerset) GetGameEndFlag() bool { return bc.phase != SomersetPhasePlaying }

// IsStalemate 手詰まり状態取得
func (bc *Somerset) IsStalemate() bool { return bc.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (bc *Somerset) SetIsStalemate(v bool) { bc.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (bc *Somerset) SetTableau(tableau [SomersetTableauCnt][]*SomersetTableauCard) {
	bc.tableau = tableau
}

// SetFoundation ファンデーション設定 (テスト用)
func (bc *Somerset) SetFoundation(foundation [SomersetFoundationCnt][]*Card) {
	bc.foundation = foundation
}

// Undo 直前の操作を取り消す
func (bc *Somerset) Undo() error {
	if bc.phase != SomersetPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(bc.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := bc.history[len(bc.history)-1]
	bc.history = bc.history[:len(bc.history)-1]
	bc.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (bc *Somerset) CanUndo() bool {
	return len(bc.history) > 0 && bc.phase == SomersetPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。
// 膠着状態でなければ0、脱出不可なら-1。
func (bc *Somerset) UndoToEscape() int {
	return undoToEscape(bc.isStalemate, bc.history, func(s *somersetSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (bc *Somerset) UndoN(n int) error {
	return undoN(bc, n)
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
//
// Somerset: 「同スートで隣接ランク（昇順・降順どちらでも）」でタブロー間の移動可能。
// Beleaguered Castle はスート無関係の降順のみ。
// 空列には「任意のカード」を置ける (Baker's Dozen との違い)。
func (bc *Somerset) canPlaceOnTableau(card *Card, col int) bool {
	colCards := bc.tableau[col]
	if len(colCards) == 0 {
		return true
	}
	topCard := colCards[len(colCards)-1].Card
	// Alternating colour, descending only. Fortress -- which this domain was
	// cloned from -- builds by SUIT in either direction, so both halves differ.
	return somersetIsBlack(card) != somersetIsBlack(topCard) &&
		card.GetValue() == topCard.GetValue()-1
}

// somersetIsBlack reports whether a card is a spade or a club. Yukon has the
// same predicate but as a method on its own type, so it cannot be reused.
func somersetIsBlack(card *Card) bool {
	return card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (bc *Somerset) canPlaceOnFoundation(card *Card, fIdx int) bool {
	return canPlaceOnFoundationPile(bc.foundation[fIdx], card)
}

// findFoundation カードを置けるファンデーションのインデックスを探す（見つからない場合-1）
func (bc *Somerset) findFoundation(card *Card) int {
	for i := range SomersetFoundationCnt {
		if bc.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定
func (bc *Somerset) checkGameClear() {
	for i := range SomersetFoundationCnt {
		if len(bc.foundation[i]) != CardValueMax {
			return
		}
	}
	bc.phase = SomersetPhaseGameClear
}

// checkStalemate 手詰まり判定
func (bc *Somerset) checkStalemate() {
	if bc.phase != SomersetPhasePlaying {
		return
	}
	if bc.GetHint() != nil {
		bc.isStalemate = false
		return
	}
	bc.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (bc *Somerset) takeSnapshot() {
	snap := &somersetSnapshot{
		phase:       bc.phase,
		moveCount:   bc.moveCount,
		isStalemate: bc.isStalemate,
	}
	for i := range SomersetTableauCnt {
		snap.tableau[i] = make([]*SomersetTableauCard, len(bc.tableau[i]))
		for j, tc := range bc.tableau[i] {
			snap.tableau[i][j] = &SomersetTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	for i := range SomersetFoundationCnt {
		snap.foundation[i] = make([]*Card, len(bc.foundation[i]))
		copy(snap.foundation[i], bc.foundation[i])
	}
	bc.history = appendSnapshot(bc.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (bc *Somerset) restoreSnapshot(snap *somersetSnapshot) {
	bc.tableau = snap.tableau
	bc.foundation = snap.foundation
	bc.phase = snap.phase
	bc.moveCount = snap.moveCount
	bc.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (bc *Somerset) appendLog(actionType, detail string, cards []*Card) {
	bc.appendLogAt(bc.moveCount, 0, actionType, detail, cards)
}

// somersetJSON is the JSON wire format for Somerset.
type somersetJSON struct {
	TrumpCards  *TrumpCards                                `json:"tc"`
	Tableau     [SomersetTableauCnt][]*SomersetTableauCard `json:"tb"`
	Foundation  [SomersetFoundationCnt][]*Card             `json:"fd"`
	Phase       SomersetPhase                              `json:"ps"`
	MoveCount   int                                        `json:"mc"`
	ActionLog   []*ActionLogEntry                          `json:"al"`
	IsStalemate bool                                       `json:"sl"`
	History     []*somersetSnapshot                        `json:"hi,omitempty"`
}

// somersetSnapshotJSON mirrors bakersDozenSnapshotJSON; see the
// matching wire-format note on bakersDozenSnapshot for the rationale behind
// the short keys (KV payload size, issue #1654).
type somersetSnapshotJSON struct {
	Tableau     [SomersetTableauCnt][]*SomersetTableauCard `json:"tb"`
	Foundation  [SomersetFoundationCnt][]*Card             `json:"fd"`
	Phase       SomersetPhase                              `json:"ps"`
	MoveCount   int                                        `json:"mc"`
	IsStalemate bool                                       `json:"sl"`
}

// MarshalJSON implements json.Marshaler for somersetSnapshot.
func (s *somersetSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(somersetSnapshotJSON{
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for somersetSnapshot.
func (s *somersetSnapshot) UnmarshalJSON(data []byte) error {
	var j somersetSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > somersetMaxSliceLen {
			return fmt.Errorf("somerset: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > somersetMaxSliceLen {
			return fmt.Errorf("somerset: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	for i := range SomersetTableauCnt {
		if s.tableau[i] == nil {
			s.tableau[i] = make([]*SomersetTableauCard, 0)
		}
	}
	s.foundation = j.Foundation
	for i := range SomersetFoundationCnt {
		if s.foundation[i] == nil {
			s.foundation[i] = make([]*Card, 0)
		}
	}
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (bc *Somerset) MarshalJSON() ([]byte, error) {
	return json.Marshal(somersetJSON{
		TrumpCards:  bc.trumpCards,
		Tableau:     bc.tableau,
		Foundation:  bc.foundation,
		Phase:       bc.phase,
		MoveCount:   bc.moveCount,
		ActionLog:   bc.actionLog,
		IsStalemate: bc.isStalemate,
		History:     bc.history,
	})
}

// somersetMaxSliceLen caps slice sizes during deserialisation.
const somersetMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (bc *Somerset) UnmarshalJSON(data []byte) error {
	var j somersetJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > somersetMaxSliceLen ||
		len(j.History) > somersetMaxSliceLen {
		return fmt.Errorf("somerset: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > somersetMaxSliceLen {
			return fmt.Errorf("somerset: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > somersetMaxSliceLen {
			return fmt.Errorf("somerset: foundation pile exceeds maximum allowed size")
		}
	}

	bc.trumpCards = j.TrumpCards
	if bc.trumpCards == nil {
		bc.trumpCards = NewTrumpCardsWithDecks(1, 0)
	}
	bc.tableau = j.Tableau
	for i := range SomersetTableauCnt {
		if bc.tableau[i] == nil {
			bc.tableau[i] = make([]*SomersetTableauCard, 0)
		}
	}
	bc.foundation = j.Foundation
	for i := range SomersetFoundationCnt {
		if bc.foundation[i] == nil {
			bc.foundation[i] = make([]*Card, 0)
		}
	}
	bc.phase = j.Phase
	bc.moveCount = j.MoveCount
	bc.actionLog = j.ActionLog
	if bc.actionLog == nil {
		bc.actionLog = make([]*ActionLogEntry, 0)
	}
	bc.history = j.History
	if bc.history == nil {
		bc.history = make([]*somersetSnapshot, 0)
	}
	bc.isStalemate = j.IsStalemate
	return nil
}
