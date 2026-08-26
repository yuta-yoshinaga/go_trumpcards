//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// FlowerGardenPhase Flower Garden game phase
type FlowerGardenPhase int

// FlowerGarden のフェーズ定数
const (
	// FlowerGardenPhasePlaying プレイ中
	FlowerGardenPhasePlaying FlowerGardenPhase = iota
	// FlowerGardenPhaseGameClear ゲームクリア
	FlowerGardenPhaseGameClear
	// FlowerGardenPhaseGameOver ゲームオーバー
	FlowerGardenPhaseGameOver
)

// FlowerGardenTableauCnt タブロー (flower-bed fan) の列数 (6列)
const FlowerGardenTableauCnt = 6

// FlowerGardenColumnLen 各 flower-bed fan の初期カード枚数 (6枚)。
const FlowerGardenColumnLen = 6

// FlowerGardenReserveCnt リザーブ ("the bouquet") のカード枚数。
// Flower Garden deals 6 fans of 6 cards = 36 cards into the tableau and the
// remaining 16 cards into the reserve, all face-up and all playable.
const FlowerGardenReserveCnt = 16

// FlowerGardenFoundationCnt ファンデーションの数
const FlowerGardenFoundationCnt = 4

// FlowerGardenTableauCard タブロー上のカード
type FlowerGardenTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// FlowerGardenHint ヒント
type FlowerGardenHint struct {
	FromZone  string // "tableau" or "reserve"
	FromCol   int    // タブロー列インデックス or リザーブインデックス
	CardIndex int    // 列内のカードインデックス (リザーブの場合 -1)
	ToZone    string // "tableau" or "foundation"
	ToCol     int    // タブロー列 or ファンデーションのインデックス
}

// FlowerGardenConfig Flower Garden ゲーム設定
type FlowerGardenConfig struct{}

// FlowerGarden ゲームクラス
type FlowerGarden struct {
	trumpCards *TrumpCards
	tableau    [FlowerGardenTableauCnt][]*FlowerGardenTableauCard
	reserve    []*Card // 16 slots; nil entries mark depleted cells (one-way)
	foundation [FlowerGardenFoundationCnt][]*Card
	phase      FlowerGardenPhase
	moveCount  int
	actionLogBase
	history     []*flowerGardenSnapshot
	isStalemate bool
}

// flowerGardenSnapshot アンドゥ用スナップショット
type flowerGardenSnapshot struct {
	tableau     [FlowerGardenTableauCnt][]*FlowerGardenTableauCard
	reserve     []*Card
	foundation  [FlowerGardenFoundationCnt][]*Card
	phase       FlowerGardenPhase
	moveCount   int
	isStalemate bool
}

// NewFlowerGarden コンストラクタ
func NewFlowerGarden(trumpCards *TrumpCards) *FlowerGarden {
	return &FlowerGarden{
		trumpCards: trumpCards,
	}
}

// NewDefaultFlowerGarden returns FlowerGarden with a single 52-card deck.
func NewDefaultFlowerGarden() *FlowerGarden {
	return NewFlowerGarden(NewTrumpCardsWithDecks(1, 0))
}

// Reset ゲームリセット
//
// Initial layout (Flower Garden / The Bouquet):
//   - The four foundations start EMPTY; the player must move every Ace out of
//     the tableau / reserve themselves.
//   - 6 tableau "flower-bed" fans are dealt 6 cards each, all face-up
//     (6×6 = 36 cards).
//   - The remaining 16 cards form the reserve ("the bouquet"), laid out
//     face-up as 16 single cards, all playable. Cards leave the reserve
//     one-way; nothing is ever moved in.
func (fg *FlowerGarden) Reset() {
	fg.trumpCards.Shuffle()
	fg.phase = FlowerGardenPhasePlaying
	fg.moveCount = 0
	fg.actionLog = nil
	fg.history = nil
	fg.isStalemate = false

	for i := range FlowerGardenFoundationCnt {
		fg.foundation[i] = nil
	}

	// Deal 6 fans of 6 cards each (36 cards total), all face-up.
	for col := range FlowerGardenTableauCnt {
		fg.tableau[col] = make([]*FlowerGardenTableauCard, 0, FlowerGardenColumnLen)
		for range FlowerGardenColumnLen {
			card := fg.trumpCards.DrawCard()
			if card == nil {
				break
			}
			fg.tableau[col] = append(fg.tableau[col], &FlowerGardenTableauCard{Card: card, FaceUp: true})
		}
	}

	// Deal the remaining 16 cards into the reserve, all face-up.
	fg.reserve = make([]*Card, 0, FlowerGardenReserveCnt)
	for range FlowerGardenReserveCnt {
		card := fg.trumpCards.DrawCard()
		if card == nil {
			break
		}
		fg.reserve = append(fg.reserve, card)
	}

	fg.checkStalemate()
}

// MoveTableauToTableau タブローからタブローにカードを移動（末尾の1枚のみ）
func (fg *FlowerGarden) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if fg.phase != FlowerGardenPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= FlowerGardenTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= FlowerGardenTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := fg.tableau[fromCol]
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	if cardIndex != len(fromCards)-1 {
		return errors.New("only the bottom card can be moved")
	}
	tc := fromCards[cardIndex]
	if !fg.canPlaceOnTableau(tc.Card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	fg.takeSnapshot()
	fg.tableau[toCol] = append(fg.tableau[toCol], tc)
	fg.tableau[fromCol] = fromCards[:cardIndex]
	fg.moveCount++
	fg.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), []*Card{tc.Card})
	fg.checkStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (fg *FlowerGarden) MoveTableauToFoundation(col int) error {
	if fg.phase != FlowerGardenPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= FlowerGardenTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := fg.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	card := tc.Card
	fIdx := fg.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	fg.takeSnapshot()
	fg.tableau[col] = fromCards[:len(fromCards)-1]
	fg.foundation[fIdx] = append(fg.foundation[fIdx], card)
	fg.moveCount++
	fg.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	fg.checkGameClear()
	fg.checkStalemate()
	return nil
}

// MoveReserveToTableau リザーブからタブローにカードを移動（リザーブは一方通行で減るのみ）
func (fg *FlowerGarden) MoveReserveToTableau(reserveIdx, toCol int) error {
	if fg.phase != FlowerGardenPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if reserveIdx < 0 || reserveIdx >= len(fg.reserve) {
		return errors.New("invalid reserve index")
	}
	if toCol < 0 || toCol >= FlowerGardenTableauCnt {
		return errors.New("invalid to column")
	}
	card := fg.reserve[reserveIdx]
	if card == nil {
		return errors.New("reserve cell is empty")
	}
	if !fg.canPlaceOnTableau(card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	fg.takeSnapshot()
	fg.reserve[reserveIdx] = nil
	fg.tableau[toCol] = append(fg.tableau[toCol], &FlowerGardenTableauCard{Card: card, FaceUp: true})
	fg.moveCount++
	fg.appendLog("move", fmt.Sprintf("リザーブ%d→タブロー列%d", reserveIdx, toCol), []*Card{card})
	fg.checkStalemate()
	return nil
}

// MoveReserveToFoundation リザーブからファンデーションにカードを移動
func (fg *FlowerGarden) MoveReserveToFoundation(reserveIdx int) error {
	if fg.phase != FlowerGardenPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if reserveIdx < 0 || reserveIdx >= len(fg.reserve) {
		return errors.New("invalid reserve index")
	}
	card := fg.reserve[reserveIdx]
	if card == nil {
		return errors.New("reserve cell is empty")
	}
	fIdx := fg.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	fg.takeSnapshot()
	fg.reserve[reserveIdx] = nil
	fg.foundation[fIdx] = append(fg.foundation[fIdx], card)
	fg.moveCount++
	fg.appendLog("move", fmt.Sprintf("リザーブ%d→ファンデーション", reserveIdx), []*Card{card})
	fg.checkGameClear()
	fg.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (fg *FlowerGarden) GiveUp() {
	if fg.phase == FlowerGardenPhasePlaying {
		fg.phase = FlowerGardenPhaseGameOver
		fg.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (fg *FlowerGarden) GetHint() *FlowerGardenHint {
	if fg.phase != FlowerGardenPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := range FlowerGardenTableauCnt {
		if len(fg.tableau[col]) == 0 {
			continue
		}
		tc := fg.tableau[col][len(fg.tableau[col])-1]
		if fg.findFoundation(tc.Card) >= 0 {
			return &FlowerGardenHint{
				FromZone:  "tableau",
				FromCol:   col,
				CardIndex: len(fg.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fg.findFoundation(tc.Card),
			}
		}
	}
	// 優先度2: リザーブからファンデーションへ
	for i := range fg.reserve {
		card := fg.reserve[i]
		if card == nil {
			continue
		}
		if fIdx := fg.findFoundation(card); fIdx >= 0 {
			return &FlowerGardenHint{
				FromZone:  "reserve",
				FromCol:   i,
				CardIndex: -1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度3: タブローからタブローへ (空列への移動は除外しフォールバックに委ねる)
	for fromCol := range FlowerGardenTableauCnt {
		fromCards := fg.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		card := fromCards[len(fromCards)-1].Card
		for toCol := range FlowerGardenTableauCnt {
			if toCol == fromCol || len(fg.tableau[toCol]) == 0 {
				continue
			}
			if fg.canPlaceOnTableau(card, toCol) {
				return &FlowerGardenHint{
					FromZone:  "tableau",
					FromCol:   fromCol,
					CardIndex: len(fromCards) - 1,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	// 優先度4: リザーブからタブローへ (空列以外)
	for i := range fg.reserve {
		card := fg.reserve[i]
		if card == nil {
			continue
		}
		for toCol := range FlowerGardenTableauCnt {
			if len(fg.tableau[toCol]) == 0 {
				continue
			}
			if fg.canPlaceOnTableau(card, toCol) {
				return &FlowerGardenHint{
					FromZone:  "reserve",
					FromCol:   i,
					CardIndex: -1,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	// 優先度5: 空列への移動 (フォールバック)
	if hint := fg.getHintToEmptyColumn(); hint != nil {
		return hint
	}
	return nil
}

// getHintToEmptyColumn は空のフラワーベッドを埋めるフォールバックヒントを返す。
// リザーブからの移動を優先し、次に他のタブロー列の末尾カードを移す。
// 1枚だけの列を別の空列に移すだけの無意味な手は除外する。
func (fg *FlowerGarden) getHintToEmptyColumn() *FlowerGardenHint {
	emptyCol := -1
	for col := range FlowerGardenTableauCnt {
		if len(fg.tableau[col]) == 0 {
			emptyCol = col
			break
		}
	}
	if emptyCol < 0 {
		return nil
	}
	// リザーブ→空列
	for i := range fg.reserve {
		if fg.reserve[i] != nil {
			return &FlowerGardenHint{
				FromZone:  "reserve",
				FromCol:   i,
				CardIndex: -1,
				ToZone:    "tableau",
				ToCol:     emptyCol,
			}
		}
	}
	// タブロー→空列 (2枚以上の列のみ。1枚の列の移動は無意味な空列交換)
	for fromCol := range FlowerGardenTableauCnt {
		if fromCol == emptyCol || len(fg.tableau[fromCol]) < 2 {
			continue
		}
		return &FlowerGardenHint{
			FromZone:  "tableau",
			FromCol:   fromCol,
			CardIndex: len(fg.tableau[fromCol]) - 1,
			ToZone:    "tableau",
			ToCol:     emptyCol,
		}
	}
	return nil
}

// AutoComplete オートコンプリート（全ての場所から可能な限りファンデーションへ）
func (fg *FlowerGarden) AutoComplete() error {
	if fg.phase != FlowerGardenPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	fg.takeSnapshot()
	for {
		moved := false
		// リザーブからファンデーションへ
		for i := range fg.reserve {
			card := fg.reserve[i]
			if card == nil {
				continue
			}
			fIdx := fg.findFoundation(card)
			if fIdx < 0 {
				continue
			}
			fg.reserve[i] = nil
			fg.foundation[fIdx] = append(fg.foundation[fIdx], card)
			fg.moveCount++
			moved = true
		}
		// タブローからファンデーションへ
		for col := range FlowerGardenTableauCnt {
			if len(fg.tableau[col]) == 0 {
				continue
			}
			tc := fg.tableau[col][len(fg.tableau[col])-1]
			fIdx := fg.findFoundation(tc.Card)
			if fIdx < 0 {
				continue
			}
			fg.tableau[col] = fg.tableau[col][:len(fg.tableau[col])-1]
			fg.foundation[fIdx] = append(fg.foundation[fIdx], tc.Card)
			fg.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	fg.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	fg.checkGameClear()
	fg.checkStalemate()
	return nil
}

// AllFaceUp 全カードが表向きかどうか（Flower Garden では常にtrue）
func (fg *FlowerGarden) AllFaceUp() bool {
	return true
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (fg *FlowerGarden) GetPhase() FlowerGardenPhase { return fg.phase }

// SetPhase フェーズ設定 (テスト用)
func (fg *FlowerGarden) SetPhase(phase FlowerGardenPhase) { fg.phase = phase }

// GetMoveCount 移動回数取得
func (fg *FlowerGarden) GetMoveCount() int { return fg.moveCount }

// GetTableau タブロー取得
func (fg *FlowerGarden) GetTableau() [FlowerGardenTableauCnt][]*FlowerGardenTableauCard {
	return fg.tableau
}

// GetReserve リザーブ取得
func (fg *FlowerGarden) GetReserve() []*Card { return fg.reserve }

// GetFoundation ファンデーション取得
func (fg *FlowerGarden) GetFoundation() [FlowerGardenFoundationCnt][]*Card {
	return fg.foundation
}

// GetGameEndFlag returns true once the game has left the playing phase.
func (fg *FlowerGarden) GetGameEndFlag() bool { return fg.phase != FlowerGardenPhasePlaying }

// IsStalemate 手詰まり状態取得
func (fg *FlowerGarden) IsStalemate() bool { return fg.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (fg *FlowerGarden) SetIsStalemate(v bool) { fg.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (fg *FlowerGarden) SetTableau(tableau [FlowerGardenTableauCnt][]*FlowerGardenTableauCard) {
	fg.tableau = tableau
}

// SetReserve リザーブ設定 (テスト用)
func (fg *FlowerGarden) SetReserve(reserve []*Card) { fg.reserve = reserve }

// SetFoundation ファンデーション設定 (テスト用)
func (fg *FlowerGarden) SetFoundation(foundation [FlowerGardenFoundationCnt][]*Card) {
	fg.foundation = foundation
}

// Undo 直前の操作を取り消す
func (fg *FlowerGarden) Undo() error {
	if fg.phase != FlowerGardenPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(fg.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := fg.history[len(fg.history)-1]
	fg.history = fg.history[:len(fg.history)-1]
	fg.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (fg *FlowerGarden) CanUndo() bool {
	return len(fg.history) > 0 && fg.phase == FlowerGardenPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。
// 膠着状態でなければ0、脱出不可なら-1。
func (fg *FlowerGarden) UndoToEscape() int {
	return undoToEscape(fg.isStalemate, fg.history, func(s *flowerGardenSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (fg *FlowerGarden) UndoN(n int) error {
	return undoN(fg, n)
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定。
//
// Flower Garden: タブロー間の積み上げは「ランクが1つ下」のみ。スートは無視する
// (赤黒交互ではない)。空のフラワーベッドには任意のカードを置ける。
func (fg *FlowerGarden) canPlaceOnTableau(card *Card, col int) bool {
	colCards := fg.tableau[col]
	if len(colCards) == 0 {
		return true
	}
	topCard := colCards[len(colCards)-1].Card
	return card.GetValue() == topCard.GetValue()-1
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定（Aceから同スートで昇順）
func (fg *FlowerGarden) canPlaceOnFoundation(card *Card, fIdx int) bool {
	return canPlaceOnFoundationPile(fg.foundation[fIdx], card)
}

// findFoundation カードを置けるファンデーションのインデックスを探す（見つからない場合-1）
func (fg *FlowerGarden) findFoundation(card *Card) int {
	for i := range FlowerGardenFoundationCnt {
		if fg.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定
func (fg *FlowerGarden) checkGameClear() {
	for i := range FlowerGardenFoundationCnt {
		if len(fg.foundation[i]) != CardValueMax {
			return
		}
	}
	fg.phase = FlowerGardenPhaseGameClear
}

// checkStalemate 手詰まり判定
func (fg *FlowerGarden) checkStalemate() {
	if fg.phase != FlowerGardenPhasePlaying {
		return
	}
	if fg.GetHint() != nil {
		fg.isStalemate = false
		return
	}
	fg.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (fg *FlowerGarden) takeSnapshot() {
	snap := &flowerGardenSnapshot{
		phase:       fg.phase,
		moveCount:   fg.moveCount,
		isStalemate: fg.isStalemate,
	}
	for i := range FlowerGardenTableauCnt {
		snap.tableau[i] = make([]*FlowerGardenTableauCard, len(fg.tableau[i]))
		for j, tc := range fg.tableau[i] {
			snap.tableau[i][j] = &FlowerGardenTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	snap.reserve = make([]*Card, len(fg.reserve))
	copy(snap.reserve, fg.reserve)
	for i := range FlowerGardenFoundationCnt {
		snap.foundation[i] = make([]*Card, len(fg.foundation[i]))
		copy(snap.foundation[i], fg.foundation[i])
	}
	fg.history = appendSnapshot(fg.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (fg *FlowerGarden) restoreSnapshot(snap *flowerGardenSnapshot) {
	fg.tableau = snap.tableau
	fg.reserve = snap.reserve
	fg.foundation = snap.foundation
	fg.phase = snap.phase
	fg.moveCount = snap.moveCount
	fg.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (fg *FlowerGarden) appendLog(actionType, detail string, cards []*Card) {
	fg.appendLogAt(fg.moveCount, 0, actionType, detail, cards)
}

// flowerGardenJSON is the JSON wire format for FlowerGarden.
type flowerGardenJSON struct {
	TrumpCards  *TrumpCards                                        `json:"tc"`
	Tableau     [FlowerGardenTableauCnt][]*FlowerGardenTableauCard `json:"tb"`
	Reserve     []*Card                                            `json:"rs"`
	Foundation  [FlowerGardenFoundationCnt][]*Card                 `json:"fd"`
	Phase       FlowerGardenPhase                                  `json:"ps"`
	MoveCount   int                                                `json:"mc"`
	ActionLog   []*ActionLogEntry                                  `json:"al"`
	IsStalemate bool                                               `json:"sl"`
	History     []*flowerGardenSnapshot                            `json:"hi,omitempty"`
}

// flowerGardenSnapshotJSON keeps short keys so the KV payload stays compact
// (issue #1654).
type flowerGardenSnapshotJSON struct {
	Tableau     [FlowerGardenTableauCnt][]*FlowerGardenTableauCard `json:"tb"`
	Reserve     []*Card                                            `json:"rs"`
	Foundation  [FlowerGardenFoundationCnt][]*Card                 `json:"fd"`
	Phase       FlowerGardenPhase                                  `json:"ps"`
	MoveCount   int                                                `json:"mc"`
	IsStalemate bool                                               `json:"sl"`
}

// flowerGardenMaxSliceLen caps slice sizes during deserialisation.
const flowerGardenMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler for flowerGardenSnapshot.
func (s *flowerGardenSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(flowerGardenSnapshotJSON{
		Tableau:     s.tableau,
		Reserve:     s.reserve,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for flowerGardenSnapshot.
func (s *flowerGardenSnapshot) UnmarshalJSON(data []byte) error {
	var j flowerGardenSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Reserve) > flowerGardenMaxSliceLen {
		return fmt.Errorf("flowergarden: snapshot reserve exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > flowerGardenMaxSliceLen {
			return fmt.Errorf("flowergarden: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > flowerGardenMaxSliceLen {
			return fmt.Errorf("flowergarden: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	for i := range FlowerGardenTableauCnt {
		if s.tableau[i] == nil {
			s.tableau[i] = make([]*FlowerGardenTableauCard, 0)
			continue
		}
		for _, tc := range s.tableau[i] {
			if tc == nil || tc.Card == nil {
				return fmt.Errorf("flowergarden: snapshot tableau contains a nil card")
			}
		}
	}
	s.reserve = j.Reserve
	if s.reserve == nil {
		s.reserve = make([]*Card, 0)
	}
	s.foundation = j.Foundation
	for i := range FlowerGardenFoundationCnt {
		if s.foundation[i] == nil {
			s.foundation[i] = make([]*Card, 0)
			continue
		}
		for _, c := range s.foundation[i] {
			if c == nil {
				return fmt.Errorf("flowergarden: snapshot foundation contains a nil card")
			}
		}
	}
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (fg *FlowerGarden) MarshalJSON() ([]byte, error) {
	return json.Marshal(flowerGardenJSON{
		TrumpCards:  fg.trumpCards,
		Tableau:     fg.tableau,
		Reserve:     fg.reserve,
		Foundation:  fg.foundation,
		Phase:       fg.phase,
		MoveCount:   fg.moveCount,
		ActionLog:   fg.actionLog,
		IsStalemate: fg.isStalemate,
		History:     fg.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (fg *FlowerGarden) UnmarshalJSON(data []byte) error {
	var j flowerGardenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > flowerGardenMaxSliceLen ||
		len(j.History) > flowerGardenMaxSliceLen ||
		len(j.Reserve) > flowerGardenMaxSliceLen {
		return fmt.Errorf("flowergarden: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > flowerGardenMaxSliceLen {
			return fmt.Errorf("flowergarden: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > flowerGardenMaxSliceLen {
			return fmt.Errorf("flowergarden: foundation pile exceeds maximum allowed size")
		}
	}

	fg.trumpCards = j.TrumpCards
	if fg.trumpCards == nil {
		fg.trumpCards = NewTrumpCardsWithDecks(1, 0)
	}
	fg.tableau = j.Tableau
	for i := range FlowerGardenTableauCnt {
		if fg.tableau[i] == nil {
			fg.tableau[i] = make([]*FlowerGardenTableauCard, 0)
			continue
		}
		for _, tc := range fg.tableau[i] {
			if tc == nil || tc.Card == nil {
				return fmt.Errorf("flowergarden: tableau contains a nil card")
			}
		}
	}
	fg.reserve = j.Reserve
	if fg.reserve == nil {
		fg.reserve = make([]*Card, 0)
	}
	fg.foundation = j.Foundation
	for i := range FlowerGardenFoundationCnt {
		if fg.foundation[i] == nil {
			fg.foundation[i] = make([]*Card, 0)
			continue
		}
		for _, c := range fg.foundation[i] {
			if c == nil {
				return fmt.Errorf("flowergarden: foundation contains a nil card")
			}
		}
	}
	fg.phase = j.Phase
	fg.moveCount = j.MoveCount
	fg.actionLog = j.ActionLog
	if fg.actionLog == nil {
		fg.actionLog = make([]*ActionLogEntry, 0)
	}
	fg.history = j.History
	if fg.history == nil {
		fg.history = make([]*flowerGardenSnapshot, 0)
	}
	fg.isStalemate = j.IsStalemate
	return nil
}
