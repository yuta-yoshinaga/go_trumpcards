//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// RussianSolitairePhase ロシアンソリティアゲームフェーズ
type RussianSolitairePhase int

// RussianSolitaireのフェーズ定数
const (
	// RussianSolitairePhasePlaying プレイ中
	RussianSolitairePhasePlaying RussianSolitairePhase = iota
	// RussianSolitairePhaseGameClear ゲームクリア
	RussianSolitairePhaseGameClear
	// RussianSolitairePhaseGameOver ゲームオーバー
	RussianSolitairePhaseGameOver
)

// RussianSolitaireTableauCnt タブローの列数
const RussianSolitaireTableauCnt = 7

// RussianSolitaireFoundationCnt ファンデーションの数
const RussianSolitaireFoundationCnt = 4

// RussianSolitaireHint ヒント
type RussianSolitaireHint struct {
	FromCol   int
	CardIndex int
	ToZone    string // "tableau" or "foundation"
	ToCol     int
}

// RussianSolitaire ロシアンソリティアゲームクラス
type RussianSolitaire struct {
	trumpCards *TrumpCards
	tableau    [RussianSolitaireTableauCnt][]*KlondikeTableauCard
	foundation [RussianSolitaireFoundationCnt][]*Card
	phase      RussianSolitairePhase
	moveCount  int
	actionLogBase
	history     []*russianSolitaireSnapshot
	isStalemate bool
}

// russianSolitaireSnapshot アンドゥ用スナップショット
type russianSolitaireSnapshot struct {
	tableau     [RussianSolitaireTableauCnt][]*KlondikeTableauCard
	foundation  [RussianSolitaireFoundationCnt][]*Card
	phase       RussianSolitairePhase
	moveCount   int
	isStalemate bool
}

// NewRussianSolitaire コンストラクタ
func NewRussianSolitaire(trumpCards *TrumpCards) *RussianSolitaire {
	return &RussianSolitaire{
		trumpCards: trumpCards,
	}
}

// NewDefaultRussianSolitaire returns RussianSolitaire with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultRussianSolitaire() *RussianSolitaire {
	return NewRussianSolitaire(NewTrumpCards(0))
}

// Reset ゲームリセット
func (y *RussianSolitaire) Reset() {
	y.trumpCards.Shuffle()
	y.phase = RussianSolitairePhasePlaying
	y.moveCount = 0
	y.actionLog = nil
	y.history = nil
	y.isStalemate = false

	// Phase 1: Klondike配りと同様に列iにi+1枚 (最後だけ表)
	for i := range RussianSolitaireTableauCnt {
		y.tableau[i] = make([]*KlondikeTableauCard, 0, i+1+4)
		for j := 0; j <= i; j++ {
			card := y.trumpCards.DrawCard()
			tc := &KlondikeTableauCard{
				Card:   card,
				FaceUp: j == i,
			}
			y.tableau[i] = append(y.tableau[i], tc)
		}
	}

	// Phase 2: 残り24枚を列1-6にそれぞれ4枚ずつ表向きで配る
	for i := 1; i < RussianSolitaireTableauCnt; i++ {
		for range 4 {
			card := y.trumpCards.DrawCard()
			y.tableau[i] = append(y.tableau[i], &KlondikeTableauCard{
				Card:   card,
				FaceUp: true,
			})
		}
	}

	// ファンデーション初期化
	for i := range RussianSolitaireFoundationCnt {
		y.foundation[i] = nil
	}
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (y *RussianSolitaire) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if y.phase != RussianSolitairePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= RussianSolitaireTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= RussianSolitaireTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := y.tableau[fromCol]
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	tc := fromCards[cardIndex]
	if !tc.FaceUp {
		return errors.New("card is face down")
	}
	// Russian Solitaire: 移動先のルールのみチェック（移動するカード群の整列は不要）
	bottomCard := tc.Card
	if !y.canPlaceOnTableau(bottomCard, toCol) {
		return errors.New("cannot place card on tableau")
	}
	// 移動実行
	y.takeSnapshot()
	movingCards := fromCards[cardIndex:]
	movedCards := make([]*Card, len(movingCards))
	for i, mc := range movingCards {
		y.tableau[toCol] = append(y.tableau[toCol], mc)
		movedCards[i] = mc.Card
	}
	y.tableau[fromCol] = fromCards[:cardIndex]
	// 自動フリップ
	y.autoFlipTableau(fromCol)
	y.moveCount++
	y.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), movedCards)
	y.checkRussianSolitaireStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (y *RussianSolitaire) MoveTableauToFoundation(col int) error {
	if y.phase != RussianSolitairePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= RussianSolitaireTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := y.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	card := tc.Card
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= RussianSolitaireFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !y.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	y.takeSnapshot()
	y.tableau[col] = fromCards[:len(fromCards)-1]
	y.foundation[fIdx] = append(y.foundation[fIdx], card)
	// 自動フリップ
	y.autoFlipTableau(col)
	y.moveCount++
	y.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	y.checkGameClear()
	y.checkRussianSolitaireStalemate()
	return nil
}

// GiveUp ギブアップ
func (y *RussianSolitaire) GiveUp() {
	if y.phase == RussianSolitairePhasePlaying {
		y.phase = RussianSolitairePhaseGameOver
		y.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (y *RussianSolitaire) GetHint() *RussianSolitaireHint {
	if y.phase != RussianSolitairePhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := range RussianSolitaireTableauCnt {
		if len(y.tableau[col]) == 0 {
			continue
		}
		tc := y.tableau[col][len(y.tableau[col])-1]
		card := tc.Card
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < RussianSolitaireFoundationCnt && y.canPlaceOnFoundation(card, fIdx) {
			return &RussianSolitaireHint{
				FromCol:   col,
				CardIndex: len(y.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: タブローからタブローへ（裏カードを開けるための移動を優先）
	for fromCol := range RussianSolitaireTableauCnt {
		fromCards := y.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		// 表向きの最初のカードを探す
		firstFaceUp := -1
		for i, tc := range fromCards {
			if tc.FaceUp {
				firstFaceUp = i
				break
			}
		}
		if firstFaceUp < 0 {
			continue
		}
		// 裏カードがない列からの移動はスキップ（既に全部表）
		if firstFaceUp == 0 {
			continue
		}
		card := fromCards[firstFaceUp].Card
		for toCol := range RussianSolitaireTableauCnt {
			if toCol == fromCol {
				continue
			}
			if y.canPlaceOnTableau(card, toCol) {
				return &RussianSolitaireHint{
					FromCol:   fromCol,
					CardIndex: firstFaceUp,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	// 優先度3: タブローからタブローへ（裏カードがなくても移動）
	for fromCol := range RussianSolitaireTableauCnt {
		fromCards := y.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		for i, tc := range fromCards {
			if !tc.FaceUp {
				continue
			}
			for toCol := range RussianSolitaireTableauCnt {
				if toCol == fromCol {
					continue
				}
				if y.canPlaceOnTableau(tc.Card, toCol) {
					return &RussianSolitaireHint{
						FromCol:   fromCol,
						CardIndex: i,
						ToZone:    "tableau",
						ToCol:     toCol,
					}
				}
			}
		}
	}
	return nil
}

// AutoComplete オートコンプリート（全カード表向きの場合に自動でファンデーションへ移動）
func (y *RussianSolitaire) AutoComplete() error {
	if y.phase != RussianSolitairePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if !y.AllFaceUp() {
		return errors.New("not all cards are face up")
	}
	y.takeSnapshot()
	for {
		moved := false
		for col := range RussianSolitaireTableauCnt {
			if len(y.tableau[col]) == 0 {
				continue
			}
			tc := y.tableau[col][len(y.tableau[col])-1]
			card := tc.Card
			fIdx := card.GetDesign() - 1
			if !y.canPlaceOnFoundation(card, fIdx) {
				continue
			}
			y.tableau[col] = y.tableau[col][:len(y.tableau[col])-1]
			y.foundation[fIdx] = append(y.foundation[fIdx], card)
			y.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	y.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	y.checkGameClear()
	return nil
}

// AllFaceUp 全カードが表向きかどうか
func (y *RussianSolitaire) AllFaceUp() bool {
	for col := range RussianSolitaireTableauCnt {
		for _, tc := range y.tableau[col] {
			if !tc.FaceUp {
				return false
			}
		}
	}
	return true
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (y *RussianSolitaire) GetPhase() RussianSolitairePhase { return y.phase }

// SetPhase フェーズ設定 (テスト用)
func (y *RussianSolitaire) SetPhase(phase RussianSolitairePhase) { y.phase = phase }

// GetMoveCount 移動回数取得
func (y *RussianSolitaire) GetMoveCount() int { return y.moveCount }

// GetTableau タブロー取得
func (y *RussianSolitaire) GetTableau() [RussianSolitaireTableauCnt][]*KlondikeTableauCard {
	return y.tableau
}

// GetFoundation ファンデーション取得
func (y *RussianSolitaire) GetFoundation() [RussianSolitaireFoundationCnt][]*Card {
	return y.foundation
}

// GetGameEndFlag returns true once the game has left the playing phase.
func (y *RussianSolitaire) GetGameEndFlag() bool { return y.phase != RussianSolitairePhasePlaying }

// IsStalemate 手詰まり状態取得
func (y *RussianSolitaire) IsStalemate() bool { return y.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (y *RussianSolitaire) SetIsStalemate(v bool) { y.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (y *RussianSolitaire) SetTableau(tableau [RussianSolitaireTableauCnt][]*KlondikeTableauCard) {
	y.tableau = tableau
}

// SetFoundation ファンデーション設定 (テスト用)
func (y *RussianSolitaire) SetFoundation(foundation [RussianSolitaireFoundationCnt][]*Card) {
	y.foundation = foundation
}

// Undo 直前の操作を取り消す
func (y *RussianSolitaire) Undo() error {
	if y.phase != RussianSolitairePhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(y.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := y.history[len(y.history)-1]
	y.history = y.history[:len(y.history)-1]
	y.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (y *RussianSolitaire) CanUndo() bool {
	return len(y.history) > 0 && y.phase == RussianSolitairePhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (y *RussianSolitaire) UndoToEscape() int {
	return undoToEscape(y.isStalemate, y.history, func(s *russianSolitaireSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (y *RussianSolitaire) UndoN(n int) error {
	for i := range n {
		if err := y.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
func (y *RussianSolitaire) canPlaceOnTableau(card *Card, col int) bool {
	colCards := y.tableau[col]
	if len(colCards) == 0 {
		// 空の列にはKのみ置ける
		return card.GetValue() == CardValueMax
	}
	topCard := colCards[len(colCards)-1].Card
	// 同スートで降順 (Russian Solitaire の Yukon との唯一の違い)
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()-1
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (y *RussianSolitaire) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := y.foundation[fIdx]
	if len(pile) == 0 {
		// 空のファンデーションにはAのみ置ける
		return card.GetValue() == 1
	}
	topCard := pile[len(pile)-1]
	// 同じスートで昇順
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()+1
}

// autoFlipTableau タブローの最上部の裏カードを自動フリップ
func (y *RussianSolitaire) autoFlipTableau(col int) {
	cards := y.tableau[col]
	if len(cards) > 0 && !cards[len(cards)-1].FaceUp {
		cards[len(cards)-1].FaceUp = true
	}
}

// checkGameClear ゲームクリア判定
func (y *RussianSolitaire) checkGameClear() {
	for i := range RussianSolitaireFoundationCnt {
		if len(y.foundation[i]) != CardValueMax {
			return
		}
	}
	y.phase = RussianSolitairePhaseGameClear
}

// checkRussianSolitaireStalemate 手詰まり判定
func (y *RussianSolitaire) checkRussianSolitaireStalemate() {
	if y.phase != RussianSolitairePhasePlaying {
		return
	}
	hint := y.GetHint()
	if hint != nil {
		y.isStalemate = false
		return
	}
	// ヒントがない = 有効な手がない → 手詰まり
	y.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (y *RussianSolitaire) takeSnapshot() {
	snap := &russianSolitaireSnapshot{
		phase:       y.phase,
		moveCount:   y.moveCount,
		isStalemate: y.isStalemate,
	}
	// deep copy tableau
	for i := range RussianSolitaireTableauCnt {
		snap.tableau[i] = make([]*KlondikeTableauCard, len(y.tableau[i]))
		for j, tc := range y.tableau[i] {
			snap.tableau[i][j] = &KlondikeTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	// deep copy foundation
	for i := range RussianSolitaireFoundationCnt {
		snap.foundation[i] = make([]*Card, len(y.foundation[i]))
		copy(snap.foundation[i], y.foundation[i])
	}
	y.history = append(y.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (y *RussianSolitaire) restoreSnapshot(snap *russianSolitaireSnapshot) {
	y.tableau = snap.tableau
	y.foundation = snap.foundation
	y.phase = snap.phase
	y.moveCount = snap.moveCount
	y.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (y *RussianSolitaire) appendLog(actionType, detail string, cards []*Card) {
	y.appendLogAt(y.moveCount, 0, actionType, detail, cards)
}

// russianSolitaireJSON is the JSON wire format for RussianSolitaire.
type russianSolitaireJSON struct {
	TrumpCards  *TrumpCards                                        `json:"tc"`
	Tableau     [RussianSolitaireTableauCnt][]*KlondikeTableauCard `json:"tb"`
	Foundation  [RussianSolitaireFoundationCnt][]*Card             `json:"fd"`
	Phase       RussianSolitairePhase                              `json:"ps"`
	MoveCount   int                                                `json:"mc"`
	ActionLog   []*ActionLogEntry                                  `json:"al"`
	IsStalemate bool                                               `json:"sl"`
	History     []*russianSolitaireSnapshot                        `json:"hi,omitempty"`
}

// russianSolitaireSnapshotJSON is the wire format for a single undo
// snapshot. russianSolitaireSnapshot uses unexported fields, so we project
// to/from this shape with explicit Marshal/Unmarshal methods. Field names
// match russianSolitaireJSON's short keys to keep the KV payload compact (#1654).
type russianSolitaireSnapshotJSON struct {
	Tableau     [RussianSolitaireTableauCnt][]*KlondikeTableauCard `json:"tb"`
	Foundation  [RussianSolitaireFoundationCnt][]*Card             `json:"fd"`
	Phase       RussianSolitairePhase                              `json:"ps"`
	MoveCount   int                                                `json:"mc"`
	IsStalemate bool                                               `json:"sl"`
}

// MarshalJSON implements json.Marshaler for russianSolitaireSnapshot.
func (s *russianSolitaireSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(russianSolitaireSnapshotJSON{
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for russianSolitaireSnapshot.
func (s *russianSolitaireSnapshot) UnmarshalJSON(data []byte) error {
	var j russianSolitaireSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > russianSolitaireMaxSliceLen {
			return fmt.Errorf("russiansolitaire: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > russianSolitaireMaxSliceLen {
			return fmt.Errorf("russiansolitaire: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.foundation = j.Foundation
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (y *RussianSolitaire) MarshalJSON() ([]byte, error) {
	return json.Marshal(russianSolitaireJSON{
		TrumpCards:  y.trumpCards,
		Tableau:     y.tableau,
		Foundation:  y.foundation,
		Phase:       y.phase,
		MoveCount:   y.moveCount,
		ActionLog:   y.actionLog,
		IsStalemate: y.isStalemate,
		History:     y.history,
	})
}

// russianSolitaireMaxSliceLen caps slice sizes during deserialisation.
const russianSolitaireMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (y *RussianSolitaire) UnmarshalJSON(data []byte) error {
	var j russianSolitaireJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > russianSolitaireMaxSliceLen ||
		len(j.History) > russianSolitaireMaxSliceLen {
		return fmt.Errorf("russiansolitaire: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > russianSolitaireMaxSliceLen {
			return fmt.Errorf("russiansolitaire: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > russianSolitaireMaxSliceLen {
			return fmt.Errorf("russiansolitaire: foundation pile exceeds maximum allowed size")
		}
	}

	y.trumpCards = j.TrumpCards
	if y.trumpCards == nil {
		y.trumpCards = NewTrumpCards(0)
	}
	y.tableau = j.Tableau
	y.foundation = j.Foundation
	y.phase = j.Phase
	y.moveCount = j.MoveCount
	y.actionLog = j.ActionLog
	if y.actionLog == nil {
		y.actionLog = make([]*ActionLogEntry, 0)
	}
	y.history = j.History
	if y.history == nil {
		y.history = make([]*russianSolitaireSnapshot, 0)
	}
	y.isStalemate = j.IsStalemate
	return nil
}
