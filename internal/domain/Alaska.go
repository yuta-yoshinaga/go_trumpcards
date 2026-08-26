//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// AlaskaPhase アラスカのゲームフェーズ
type AlaskaPhase int

// Alaskaのフェーズ定数
const (
	// AlaskaPhasePlaying プレイ中
	AlaskaPhasePlaying AlaskaPhase = iota
	// AlaskaPhaseGameClear ゲームクリア
	AlaskaPhaseGameClear
	// AlaskaPhaseGameOver ゲームオーバー
	AlaskaPhaseGameOver
)

// AlaskaTableauCnt タブローの列数
const AlaskaTableauCnt = 7

// AlaskaFoundationCnt ファンデーションの数
const AlaskaFoundationCnt = 4

// AlaskaTableauCard タブロー上のカード。Alaska は Klondike とは別バケット
// (extra4 worker) に属するため、AlaskaTableauCard を共有せず独自に定義する。
// DoubleKlondike が同じ理由で DoubleAlaskaTableauCard を持つのと同じ扱い。
// JSON のフィールド名は Klondike と同一なので、通信形式は変わらない。
type AlaskaTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// AlaskaHint ヒント
type AlaskaHint struct {
	FromCol   int
	CardIndex int
	ToZone    string // "tableau" or "foundation"
	ToCol     int
}

// Alaska アラスカ（ユーコン系ソリティア）ゲームクラス。
//
// Yukon 系の「伏せ札の上に載った列ごと動かせる」土台をそのまま持ち、
// タブローの連結条件だけが違う:
//
//	Yukon            異色・降順のみ / 空列は K のみ
//	RussianSolitaire 同スート・降順のみ / 空列は K のみ
//	Alaska           同スート・昇順**または**降順 / 空列は任意の札
//
// 差分は canPlaceOnTableau の一箇所に閉じている。移動・ヒント・手詰まり判定は
// すべてそこを通るので、規則を足すときもこの関数だけを見ればよい。
type Alaska struct {
	trumpCards *TrumpCards
	tableau    [AlaskaTableauCnt][]*AlaskaTableauCard
	foundation [AlaskaFoundationCnt][]*Card
	phase      AlaskaPhase
	moveCount  int
	actionLogBase
	history     []*alaskaSnapshot
	isStalemate bool
}

// alaskaSnapshot アンドゥ用スナップショット
type alaskaSnapshot struct {
	tableau     [AlaskaTableauCnt][]*AlaskaTableauCard
	foundation  [AlaskaFoundationCnt][]*Card
	phase       AlaskaPhase
	moveCount   int
	isStalemate bool
}

// NewAlaska コンストラクタ
func NewAlaska(trumpCards *TrumpCards) *Alaska {
	return &Alaska{
		trumpCards: trumpCards,
	}
}

// NewDefaultAlaska returns Alaska with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultAlaska() *Alaska {
	return NewAlaska(NewTrumpCards(0))
}

// Reset ゲームリセット
func (y *Alaska) Reset() {
	y.trumpCards.Shuffle()
	y.phase = AlaskaPhasePlaying
	y.moveCount = 0
	y.actionLog = nil
	y.history = nil
	y.isStalemate = false

	// Phase 1: Klondike配りと同様に列iにi+1枚 (最後だけ表)
	for i := range AlaskaTableauCnt {
		y.tableau[i] = make([]*AlaskaTableauCard, 0, i+1+4)
		for j := 0; j <= i; j++ {
			card := y.trumpCards.DrawCard()
			tc := &AlaskaTableauCard{
				Card:   card,
				FaceUp: j == i,
			}
			y.tableau[i] = append(y.tableau[i], tc)
		}
	}

	// Phase 2: 残り24枚を列1-6にそれぞれ4枚ずつ表向きで配る
	for i := 1; i < AlaskaTableauCnt; i++ {
		for range 4 {
			card := y.trumpCards.DrawCard()
			y.tableau[i] = append(y.tableau[i], &AlaskaTableauCard{
				Card:   card,
				FaceUp: true,
			})
		}
	}

	// ファンデーション初期化
	for i := range AlaskaFoundationCnt {
		y.foundation[i] = nil
	}
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (y *Alaska) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if y.phase != AlaskaPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= AlaskaTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= AlaskaTableauCnt {
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
	// Alaska: 移動先のルールのみチェック（移動するカード群の整列は不要 -- Yukon 譲り）
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
	y.checkAlaskaStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (y *Alaska) MoveTableauToFoundation(col int) error {
	if y.phase != AlaskaPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= AlaskaTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := y.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	card := tc.Card
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= AlaskaFoundationCnt {
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
	y.checkAlaskaStalemate()
	return nil
}

// GiveUp ギブアップ
func (y *Alaska) GiveUp() {
	if y.phase == AlaskaPhasePlaying {
		y.phase = AlaskaPhaseGameOver
		y.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (y *Alaska) GetHint() *AlaskaHint {
	if y.phase != AlaskaPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := range AlaskaTableauCnt {
		if len(y.tableau[col]) == 0 {
			continue
		}
		tc := y.tableau[col][len(y.tableau[col])-1]
		card := tc.Card
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < AlaskaFoundationCnt && y.canPlaceOnFoundation(card, fIdx) {
			return &AlaskaHint{
				FromCol:   col,
				CardIndex: len(y.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: タブローからタブローへ（裏カードを開けるための移動を優先）
	for fromCol := range AlaskaTableauCnt {
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
		for toCol := range AlaskaTableauCnt {
			if toCol == fromCol {
				continue
			}
			if y.canPlaceOnTableau(card, toCol) {
				return &AlaskaHint{
					FromCol:   fromCol,
					CardIndex: firstFaceUp,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	// 優先度3: タブローからタブローへ（裏カードがなくても移動）
	for fromCol := range AlaskaTableauCnt {
		fromCards := y.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		for i, tc := range fromCards {
			if !tc.FaceUp {
				continue
			}
			for toCol := range AlaskaTableauCnt {
				if toCol == fromCol {
					continue
				}
				if y.canPlaceOnTableau(tc.Card, toCol) {
					return &AlaskaHint{
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
func (y *Alaska) AutoComplete() error {
	if y.phase != AlaskaPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if !y.AllFaceUp() {
		return errors.New("not all cards are face up")
	}
	y.takeSnapshot()
	for {
		moved := false
		for col := range AlaskaTableauCnt {
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
func (y *Alaska) AllFaceUp() bool {
	for col := range AlaskaTableauCnt {
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
func (y *Alaska) GetPhase() AlaskaPhase { return y.phase }

// SetPhase フェーズ設定 (テスト用)
func (y *Alaska) SetPhase(phase AlaskaPhase) { y.phase = phase }

// GetMoveCount 移動回数取得
func (y *Alaska) GetMoveCount() int { return y.moveCount }

// GetTableau タブロー取得
func (y *Alaska) GetTableau() [AlaskaTableauCnt][]*AlaskaTableauCard {
	return y.tableau
}

// GetFoundation ファンデーション取得
func (y *Alaska) GetFoundation() [AlaskaFoundationCnt][]*Card {
	return y.foundation
}

// GetGameEndFlag returns true once the game has left the playing phase.
func (y *Alaska) GetGameEndFlag() bool { return y.phase != AlaskaPhasePlaying }

// IsStalemate 手詰まり状態取得
func (y *Alaska) IsStalemate() bool { return y.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (y *Alaska) SetIsStalemate(v bool) { y.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (y *Alaska) SetTableau(tableau [AlaskaTableauCnt][]*AlaskaTableauCard) {
	y.tableau = tableau
}

// SetFoundation ファンデーション設定 (テスト用)
func (y *Alaska) SetFoundation(foundation [AlaskaFoundationCnt][]*Card) {
	y.foundation = foundation
}

// Undo 直前の操作を取り消す
func (y *Alaska) Undo() error {
	if y.phase != AlaskaPhasePlaying {
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
func (y *Alaska) CanUndo() bool {
	return len(y.history) > 0 && y.phase == AlaskaPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (y *Alaska) UndoToEscape() int {
	return undoToEscape(y.isStalemate, y.history, func(s *alaskaSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (y *Alaska) UndoN(n int) error {
	return undoN(y, n)
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
func (y *Alaska) canPlaceOnTableau(card *Card, col int) bool {
	colCards := y.tableau[col]
	if len(colCards) == 0 {
		// 空の列には任意の札を置ける。Yukon / Russian Solitaire は K のみ。
		return true
	}
	topCard := colCards[len(colCards)-1].Card
	if card.GetDesign() != topCard.GetDesign() {
		return false
	}
	// 同スートなら昇順・降順どちらの接続でもよい。隣接ランクであることは必要で、
	// 「同スートなら何でも」ではない。
	diff := card.GetValue() - topCard.GetValue()
	return diff == 1 || diff == -1
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (y *Alaska) canPlaceOnFoundation(card *Card, fIdx int) bool {
	return canPlaceOnFoundationPile(y.foundation[fIdx], card)
}

// autoFlipTableau 列 col の最上位が裏向きなら表に返す。
// Klondike の autoFlipTopCard は KlondikeTableauCard を取るので、独自の
// AlaskaTableauCard を持つ以上そのままでは使えない。DoubleKlondike と同じく
// 3 行を自前で持つ。
func (y *Alaska) autoFlipTableau(col int) {
	cards := y.tableau[col]
	if len(cards) > 0 && !cards[len(cards)-1].FaceUp {
		cards[len(cards)-1].FaceUp = true
	}
}

// checkGameClear ゲームクリア判定
func (y *Alaska) checkGameClear() {
	for i := range AlaskaFoundationCnt {
		if len(y.foundation[i]) != CardValueMax {
			return
		}
	}
	y.phase = AlaskaPhaseGameClear
}

// checkAlaskaStalemate 手詰まり判定
func (y *Alaska) checkAlaskaStalemate() {
	if y.phase != AlaskaPhasePlaying {
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
func (y *Alaska) takeSnapshot() {
	snap := &alaskaSnapshot{
		phase:       y.phase,
		moveCount:   y.moveCount,
		isStalemate: y.isStalemate,
	}
	// deep copy tableau
	for i := range AlaskaTableauCnt {
		snap.tableau[i] = make([]*AlaskaTableauCard, len(y.tableau[i]))
		for j, tc := range y.tableau[i] {
			snap.tableau[i][j] = &AlaskaTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	// deep copy foundation
	for i := range AlaskaFoundationCnt {
		snap.foundation[i] = make([]*Card, len(y.foundation[i]))
		copy(snap.foundation[i], y.foundation[i])
	}
	y.history = appendSnapshot(y.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (y *Alaska) restoreSnapshot(snap *alaskaSnapshot) {
	y.tableau = snap.tableau
	y.foundation = snap.foundation
	y.phase = snap.phase
	y.moveCount = snap.moveCount
	y.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (y *Alaska) appendLog(actionType, detail string, cards []*Card) {
	y.appendLogAt(y.moveCount, 0, actionType, detail, cards)
}

// alaskaJSON is the JSON wire format for Alaska.
type alaskaJSON struct {
	TrumpCards  *TrumpCards                            `json:"tc"`
	Tableau     [AlaskaTableauCnt][]*AlaskaTableauCard `json:"tb"`
	Foundation  [AlaskaFoundationCnt][]*Card           `json:"fd"`
	Phase       AlaskaPhase                            `json:"ps"`
	MoveCount   int                                    `json:"mc"`
	ActionLog   []*ActionLogEntry                      `json:"al"`
	IsStalemate bool                                   `json:"sl"`
	History     []*alaskaSnapshot                      `json:"hi,omitempty"`
}

// alaskaSnapshotJSON is the wire format for a single undo
// snapshot. alaskaSnapshot uses unexported fields, so we project
// to/from this shape with explicit Marshal/Unmarshal methods. Field names
// match alaskaJSON's short keys to keep the KV payload compact (#1654).
type alaskaSnapshotJSON struct {
	Tableau     [AlaskaTableauCnt][]*AlaskaTableauCard `json:"tb"`
	Foundation  [AlaskaFoundationCnt][]*Card           `json:"fd"`
	Phase       AlaskaPhase                            `json:"ps"`
	MoveCount   int                                    `json:"mc"`
	IsStalemate bool                                   `json:"sl"`
}

// MarshalJSON implements json.Marshaler for alaskaSnapshot.
func (s *alaskaSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(alaskaSnapshotJSON{
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for alaskaSnapshot.
func (s *alaskaSnapshot) UnmarshalJSON(data []byte) error {
	var j alaskaSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > alaskaMaxSliceLen {
			return fmt.Errorf("alaska: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > alaskaMaxSliceLen {
			return fmt.Errorf("alaska: snapshot foundation pile exceeds maximum allowed size")
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
func (y *Alaska) MarshalJSON() ([]byte, error) {
	return json.Marshal(alaskaJSON{
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

// alaskaMaxSliceLen caps slice sizes during deserialisation.
const alaskaMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (y *Alaska) UnmarshalJSON(data []byte) error {
	var j alaskaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > alaskaMaxSliceLen ||
		len(j.History) > alaskaMaxSliceLen {
		return fmt.Errorf("alaska: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > alaskaMaxSliceLen {
			return fmt.Errorf("alaska: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > alaskaMaxSliceLen {
			return fmt.Errorf("alaska: foundation pile exceeds maximum allowed size")
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
		y.history = make([]*alaskaSnapshot, 0)
	}
	y.isStalemate = j.IsStalemate
	return nil
}
