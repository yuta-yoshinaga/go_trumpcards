//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// KingAlbertPhase King Albert game phase
type KingAlbertPhase int

// KingAlbert のフェーズ定数
const (
	// KingAlbertPhasePlaying プレイ中
	KingAlbertPhasePlaying KingAlbertPhase = iota
	// KingAlbertPhaseGameClear ゲームクリア
	KingAlbertPhaseGameClear
	// KingAlbertPhaseGameOver ゲームオーバー
	KingAlbertPhaseGameOver
)

// KingAlbertTableauCnt タブローの列数 (9列)
const KingAlbertTableauCnt = 9

// KingAlbertReserveCnt リザーブ ("Lawrence") のカード枚数。
// King Albert deals 1+2+...+9 = 45 cards into the tableau and the remaining
// 7 cards into the reserve, all face-up.
const KingAlbertReserveCnt = 7

// KingAlbertColumnLen 初期配置時の最長列のカード枚数 (9列目)。
const KingAlbertColumnLen = 9

// KingAlbertFoundationCnt ファンデーションの数
const KingAlbertFoundationCnt = 4

// KingAlbertTableauCard タブロー上のカード
type KingAlbertTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// KingAlbertHint ヒント
type KingAlbertHint struct {
	FromZone  string // "tableau" or "reserve"
	FromCol   int    // タブロー列インデックス or リザーブインデックス
	CardIndex int    // 列内のカードインデックス (リザーブの場合 -1)
	ToZone    string // "tableau" or "foundation"
	ToCol     int    // タブロー列 or ファンデーションのインデックス
}

// KingAlbertConfig King Albert ゲーム設定
type KingAlbertConfig struct{}

// KingAlbert ゲームクラス
type KingAlbert struct {
	trumpCards *TrumpCards
	tableau    [KingAlbertTableauCnt][]*KingAlbertTableauCard
	reserve    []*Card // 7 slots; nil entries mark depleted cells (one-way)
	foundation [KingAlbertFoundationCnt][]*Card
	phase      KingAlbertPhase
	moveCount  int
	actionLogBase
	history     []*kingAlbertSnapshot
	isStalemate bool
}

// kingAlbertSnapshot アンドゥ用スナップショット
type kingAlbertSnapshot struct {
	tableau     [KingAlbertTableauCnt][]*KingAlbertTableauCard
	reserve     []*Card
	foundation  [KingAlbertFoundationCnt][]*Card
	phase       KingAlbertPhase
	moveCount   int
	isStalemate bool
}

// NewKingAlbert コンストラクタ
func NewKingAlbert(trumpCards *TrumpCards) *KingAlbert {
	return &KingAlbert{
		trumpCards: trumpCards,
	}
}

// NewDefaultKingAlbert returns KingAlbert with a single 52-card deck.
func NewDefaultKingAlbert() *KingAlbert {
	return NewKingAlbert(NewTrumpCardsWithDecks(1, 0))
}

// Reset ゲームリセット
//
// Initial layout (King Albert):
//   - The four foundations start EMPTY; the player must move every Ace out of
//     the tableau / reserve themselves.
//   - 9 tableau columns are dealt 1,2,3,...,9 cards respectively, all face-up
//     (1+2+...+9 = 45 cards).
//   - The remaining 7 cards form the reserve ("Lawrence"), laid out face-up as
//     7 single cards. Cards leave the reserve one-way; nothing is ever moved in.
func (ka *KingAlbert) Reset() {
	ka.trumpCards.Shuffle()
	ka.phase = KingAlbertPhasePlaying
	ka.moveCount = 0
	ka.actionLog = nil
	ka.history = nil
	ka.isStalemate = false

	for i := range KingAlbertFoundationCnt {
		ka.foundation[i] = nil
	}

	// Deal column N (0-indexed) with N+1 cards: 1,2,3,...,9 (45 cards total).
	for col := range KingAlbertTableauCnt {
		colLen := col + 1
		ka.tableau[col] = make([]*KingAlbertTableauCard, 0, colLen)
		for range colLen {
			card := ka.trumpCards.DrawCard()
			if card == nil {
				break
			}
			ka.tableau[col] = append(ka.tableau[col], &KingAlbertTableauCard{Card: card, FaceUp: true})
		}
	}

	// Deal the remaining 7 cards into the reserve, all face-up.
	ka.reserve = make([]*Card, 0, KingAlbertReserveCnt)
	for range KingAlbertReserveCnt {
		card := ka.trumpCards.DrawCard()
		if card == nil {
			break
		}
		ka.reserve = append(ka.reserve, card)
	}

	ka.checkStalemate()
}

// MoveTableauToTableau タブローからタブローにカードを移動（末尾の1枚のみ）
func (ka *KingAlbert) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if ka.phase != KingAlbertPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= KingAlbertTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= KingAlbertTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := ka.tableau[fromCol]
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
	if !ka.canPlaceOnTableau(tc.Card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	ka.takeSnapshot()
	ka.tableau[toCol] = append(ka.tableau[toCol], tc)
	ka.tableau[fromCol] = fromCards[:cardIndex]
	ka.moveCount++
	ka.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), []*Card{tc.Card})
	ka.checkStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (ka *KingAlbert) MoveTableauToFoundation(col int) error {
	if ka.phase != KingAlbertPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= KingAlbertTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := ka.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	card := tc.Card
	fIdx := ka.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	ka.takeSnapshot()
	ka.tableau[col] = fromCards[:len(fromCards)-1]
	ka.foundation[fIdx] = append(ka.foundation[fIdx], card)
	ka.moveCount++
	ka.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	ka.checkGameClear()
	ka.checkStalemate()
	return nil
}

// MoveReserveToTableau リザーブからタブローにカードを移動（リザーブは一方通行で減るのみ）
func (ka *KingAlbert) MoveReserveToTableau(reserveIdx, toCol int) error {
	if ka.phase != KingAlbertPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if reserveIdx < 0 || reserveIdx >= len(ka.reserve) {
		return errors.New("invalid reserve index")
	}
	if toCol < 0 || toCol >= KingAlbertTableauCnt {
		return errors.New("invalid to column")
	}
	card := ka.reserve[reserveIdx]
	if card == nil {
		return errors.New("reserve cell is empty")
	}
	if !ka.canPlaceOnTableau(card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	ka.takeSnapshot()
	ka.reserve[reserveIdx] = nil
	ka.tableau[toCol] = append(ka.tableau[toCol], &KingAlbertTableauCard{Card: card, FaceUp: true})
	ka.moveCount++
	ka.appendLog("move", fmt.Sprintf("リザーブ%d→タブロー列%d", reserveIdx, toCol), []*Card{card})
	ka.checkStalemate()
	return nil
}

// MoveReserveToFoundation リザーブからファンデーションにカードを移動
func (ka *KingAlbert) MoveReserveToFoundation(reserveIdx int) error {
	if ka.phase != KingAlbertPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if reserveIdx < 0 || reserveIdx >= len(ka.reserve) {
		return errors.New("invalid reserve index")
	}
	card := ka.reserve[reserveIdx]
	if card == nil {
		return errors.New("reserve cell is empty")
	}
	fIdx := ka.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	ka.takeSnapshot()
	ka.reserve[reserveIdx] = nil
	ka.foundation[fIdx] = append(ka.foundation[fIdx], card)
	ka.moveCount++
	ka.appendLog("move", fmt.Sprintf("リザーブ%d→ファンデーション", reserveIdx), []*Card{card})
	ka.checkGameClear()
	ka.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (ka *KingAlbert) GiveUp() {
	if ka.phase == KingAlbertPhasePlaying {
		ka.phase = KingAlbertPhaseGameOver
		ka.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (ka *KingAlbert) GetHint() *KingAlbertHint {
	if ka.phase != KingAlbertPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := range KingAlbertTableauCnt {
		if len(ka.tableau[col]) == 0 {
			continue
		}
		tc := ka.tableau[col][len(ka.tableau[col])-1]
		if ka.findFoundation(tc.Card) >= 0 {
			return &KingAlbertHint{
				FromZone:  "tableau",
				FromCol:   col,
				CardIndex: len(ka.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     ka.findFoundation(tc.Card),
			}
		}
	}
	// 優先度2: リザーブからファンデーションへ
	for i := range ka.reserve {
		card := ka.reserve[i]
		if card == nil {
			continue
		}
		if fIdx := ka.findFoundation(card); fIdx >= 0 {
			return &KingAlbertHint{
				FromZone:  "reserve",
				FromCol:   i,
				CardIndex: -1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度3: タブローからタブローへ (空列への移動は除外しフォールバックに委ねる)
	for fromCol := range KingAlbertTableauCnt {
		fromCards := ka.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		card := fromCards[len(fromCards)-1].Card
		for toCol := range KingAlbertTableauCnt {
			if toCol == fromCol || len(ka.tableau[toCol]) == 0 {
				continue
			}
			if ka.canPlaceOnTableau(card, toCol) {
				return &KingAlbertHint{
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
	for i := range ka.reserve {
		card := ka.reserve[i]
		if card == nil {
			continue
		}
		for toCol := range KingAlbertTableauCnt {
			if len(ka.tableau[toCol]) == 0 {
				continue
			}
			if ka.canPlaceOnTableau(card, toCol) {
				return &KingAlbertHint{
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
	if hint := ka.getHintToEmptyColumn(); hint != nil {
		return hint
	}
	return nil
}

// getHintToEmptyColumn は空列を埋めるフォールバックヒントを返す。
// リザーブからの移動を優先し、次に他のタブロー列の末尾カードを移す。
// 1枚だけの列を別の空列に移すだけの無意味な手は除外する。
func (ka *KingAlbert) getHintToEmptyColumn() *KingAlbertHint {
	emptyCol := -1
	for col := range KingAlbertTableauCnt {
		if len(ka.tableau[col]) == 0 {
			emptyCol = col
			break
		}
	}
	if emptyCol < 0 {
		return nil
	}
	// リザーブ→空列
	for i := range ka.reserve {
		if ka.reserve[i] != nil {
			return &KingAlbertHint{
				FromZone:  "reserve",
				FromCol:   i,
				CardIndex: -1,
				ToZone:    "tableau",
				ToCol:     emptyCol,
			}
		}
	}
	// タブロー→空列 (2枚以上の列のみ。1枚の列の移動は無意味な空列交換)
	for fromCol := range KingAlbertTableauCnt {
		if fromCol == emptyCol || len(ka.tableau[fromCol]) < 2 {
			continue
		}
		return &KingAlbertHint{
			FromZone:  "tableau",
			FromCol:   fromCol,
			CardIndex: len(ka.tableau[fromCol]) - 1,
			ToZone:    "tableau",
			ToCol:     emptyCol,
		}
	}
	return nil
}

// AutoComplete オートコンプリート（全ての場所から可能な限りファンデーションへ）
func (ka *KingAlbert) AutoComplete() error {
	if ka.phase != KingAlbertPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	ka.takeSnapshot()
	for {
		moved := false
		// リザーブからファンデーションへ
		for i := range ka.reserve {
			card := ka.reserve[i]
			if card == nil {
				continue
			}
			fIdx := ka.findFoundation(card)
			if fIdx < 0 {
				continue
			}
			ka.reserve[i] = nil
			ka.foundation[fIdx] = append(ka.foundation[fIdx], card)
			ka.moveCount++
			moved = true
		}
		// タブローからファンデーションへ
		for col := range KingAlbertTableauCnt {
			if len(ka.tableau[col]) == 0 {
				continue
			}
			tc := ka.tableau[col][len(ka.tableau[col])-1]
			fIdx := ka.findFoundation(tc.Card)
			if fIdx < 0 {
				continue
			}
			ka.tableau[col] = ka.tableau[col][:len(ka.tableau[col])-1]
			ka.foundation[fIdx] = append(ka.foundation[fIdx], tc.Card)
			ka.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	ka.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	ka.checkGameClear()
	ka.checkStalemate()
	return nil
}

// AllFaceUp 全カードが表向きかどうか（King Albert では常にtrue）
func (ka *KingAlbert) AllFaceUp() bool {
	return true
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (ka *KingAlbert) GetPhase() KingAlbertPhase { return ka.phase }

// SetPhase フェーズ設定 (テスト用)
func (ka *KingAlbert) SetPhase(phase KingAlbertPhase) { ka.phase = phase }

// GetMoveCount 移動回数取得
func (ka *KingAlbert) GetMoveCount() int { return ka.moveCount }

// GetTableau タブロー取得
func (ka *KingAlbert) GetTableau() [KingAlbertTableauCnt][]*KingAlbertTableauCard {
	return ka.tableau
}

// GetReserve リザーブ取得
func (ka *KingAlbert) GetReserve() []*Card { return ka.reserve }

// GetFoundation ファンデーション取得
func (ka *KingAlbert) GetFoundation() [KingAlbertFoundationCnt][]*Card {
	return ka.foundation
}

// GetGameEndFlag returns true once the game has left the playing phase.
func (ka *KingAlbert) GetGameEndFlag() bool { return ka.phase != KingAlbertPhasePlaying }

// IsStalemate 手詰まり状態取得
func (ka *KingAlbert) IsStalemate() bool { return ka.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (ka *KingAlbert) SetIsStalemate(v bool) { ka.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (ka *KingAlbert) SetTableau(tableau [KingAlbertTableauCnt][]*KingAlbertTableauCard) {
	ka.tableau = tableau
}

// SetReserve リザーブ設定 (テスト用)
func (ka *KingAlbert) SetReserve(reserve []*Card) { ka.reserve = reserve }

// SetFoundation ファンデーション設定 (テスト用)
func (ka *KingAlbert) SetFoundation(foundation [KingAlbertFoundationCnt][]*Card) {
	ka.foundation = foundation
}

// Undo 直前の操作を取り消す
func (ka *KingAlbert) Undo() error {
	if ka.phase != KingAlbertPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(ka.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := ka.history[len(ka.history)-1]
	ka.history = ka.history[:len(ka.history)-1]
	ka.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (ka *KingAlbert) CanUndo() bool {
	return len(ka.history) > 0 && ka.phase == KingAlbertPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。
// 膠着状態でなければ0、脱出不可なら-1。
func (ka *KingAlbert) UndoToEscape() int {
	return undoToEscape(ka.isStalemate, ka.history, func(s *kingAlbertSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (ka *KingAlbert) UndoN(n int) error {
	return undoN(ka, n)
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定。
//
// King Albert: タブロー間の積み上げは「赤黒交互の降順」(red on black / black on
// red、ランクは1つ下)。空列には任意のカードを置ける。
func (ka *KingAlbert) canPlaceOnTableau(card *Card, col int) bool {
	colCards := ka.tableau[col]
	if len(colCards) == 0 {
		return true
	}
	topCard := colCards[len(colCards)-1].Card
	return card.GetValue() == topCard.GetValue()-1 && ka.isAlternateColor(card, topCard)
}

// isAlternateColor 交互の色かどうか判定
func (ka *KingAlbert) isAlternateColor(card1, card2 *Card) bool {
	return ka.isBlack(card1) != ka.isBlack(card2)
}

// isBlack 黒いカードかどうか
func (ka *KingAlbert) isBlack(card *Card) bool {
	return card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定（Aceから同スートで昇順）
func (ka *KingAlbert) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := ka.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	topCard := pile[len(pile)-1]
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()+1
}

// findFoundation カードを置けるファンデーションのインデックスを探す（見つからない場合-1）
func (ka *KingAlbert) findFoundation(card *Card) int {
	for i := range KingAlbertFoundationCnt {
		if ka.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定
func (ka *KingAlbert) checkGameClear() {
	for i := range KingAlbertFoundationCnt {
		if len(ka.foundation[i]) != CardValueMax {
			return
		}
	}
	ka.phase = KingAlbertPhaseGameClear
}

// checkStalemate 手詰まり判定
func (ka *KingAlbert) checkStalemate() {
	if ka.phase != KingAlbertPhasePlaying {
		return
	}
	if ka.GetHint() != nil {
		ka.isStalemate = false
		return
	}
	ka.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (ka *KingAlbert) takeSnapshot() {
	snap := &kingAlbertSnapshot{
		phase:       ka.phase,
		moveCount:   ka.moveCount,
		isStalemate: ka.isStalemate,
	}
	for i := range KingAlbertTableauCnt {
		snap.tableau[i] = make([]*KingAlbertTableauCard, len(ka.tableau[i]))
		for j, tc := range ka.tableau[i] {
			snap.tableau[i][j] = &KingAlbertTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	snap.reserve = make([]*Card, len(ka.reserve))
	copy(snap.reserve, ka.reserve)
	for i := range KingAlbertFoundationCnt {
		snap.foundation[i] = make([]*Card, len(ka.foundation[i]))
		copy(snap.foundation[i], ka.foundation[i])
	}
	ka.history = append(ka.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (ka *KingAlbert) restoreSnapshot(snap *kingAlbertSnapshot) {
	ka.tableau = snap.tableau
	ka.reserve = snap.reserve
	ka.foundation = snap.foundation
	ka.phase = snap.phase
	ka.moveCount = snap.moveCount
	ka.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (ka *KingAlbert) appendLog(actionType, detail string, cards []*Card) {
	ka.appendLogAt(ka.moveCount, 0, actionType, detail, cards)
}

// kingAlbertJSON is the JSON wire format for KingAlbert.
type kingAlbertJSON struct {
	TrumpCards  *TrumpCards                                    `json:"tc"`
	Tableau     [KingAlbertTableauCnt][]*KingAlbertTableauCard `json:"tb"`
	Reserve     []*Card                                        `json:"rs"`
	Foundation  [KingAlbertFoundationCnt][]*Card               `json:"fd"`
	Phase       KingAlbertPhase                                `json:"ps"`
	MoveCount   int                                            `json:"mc"`
	ActionLog   []*ActionLogEntry                              `json:"al"`
	IsStalemate bool                                           `json:"sl"`
	History     []*kingAlbertSnapshot                          `json:"hi,omitempty"`
}

// kingAlbertSnapshotJSON mirrors the streetsAndAlleys wire format; the short
// keys keep the KV payload compact (issue #1654).
type kingAlbertSnapshotJSON struct {
	Tableau     [KingAlbertTableauCnt][]*KingAlbertTableauCard `json:"tb"`
	Reserve     []*Card                                        `json:"rs"`
	Foundation  [KingAlbertFoundationCnt][]*Card               `json:"fd"`
	Phase       KingAlbertPhase                                `json:"ps"`
	MoveCount   int                                            `json:"mc"`
	IsStalemate bool                                           `json:"sl"`
}

// kingAlbertMaxSliceLen caps slice sizes during deserialisation.
const kingAlbertMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler for kingAlbertSnapshot.
func (s *kingAlbertSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(kingAlbertSnapshotJSON{
		Tableau:     s.tableau,
		Reserve:     s.reserve,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for kingAlbertSnapshot.
func (s *kingAlbertSnapshot) UnmarshalJSON(data []byte) error {
	var j kingAlbertSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Reserve) > kingAlbertMaxSliceLen {
		return fmt.Errorf("kingalbert: snapshot reserve exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > kingAlbertMaxSliceLen {
			return fmt.Errorf("kingalbert: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > kingAlbertMaxSliceLen {
			return fmt.Errorf("kingalbert: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	for i := range KingAlbertTableauCnt {
		if s.tableau[i] == nil {
			s.tableau[i] = make([]*KingAlbertTableauCard, 0)
			continue
		}
		for _, tc := range s.tableau[i] {
			if tc == nil || tc.Card == nil {
				return fmt.Errorf("kingalbert: snapshot tableau contains a nil card")
			}
		}
	}
	s.reserve = j.Reserve
	if s.reserve == nil {
		s.reserve = make([]*Card, 0)
	}
	s.foundation = j.Foundation
	for i := range KingAlbertFoundationCnt {
		if s.foundation[i] == nil {
			s.foundation[i] = make([]*Card, 0)
			continue
		}
		for _, c := range s.foundation[i] {
			if c == nil {
				return fmt.Errorf("kingalbert: snapshot foundation contains a nil card")
			}
		}
	}
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (ka *KingAlbert) MarshalJSON() ([]byte, error) {
	return json.Marshal(kingAlbertJSON{
		TrumpCards:  ka.trumpCards,
		Tableau:     ka.tableau,
		Reserve:     ka.reserve,
		Foundation:  ka.foundation,
		Phase:       ka.phase,
		MoveCount:   ka.moveCount,
		ActionLog:   ka.actionLog,
		IsStalemate: ka.isStalemate,
		History:     ka.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (ka *KingAlbert) UnmarshalJSON(data []byte) error {
	var j kingAlbertJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > kingAlbertMaxSliceLen ||
		len(j.History) > kingAlbertMaxSliceLen ||
		len(j.Reserve) > kingAlbertMaxSliceLen {
		return fmt.Errorf("kingalbert: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > kingAlbertMaxSliceLen {
			return fmt.Errorf("kingalbert: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > kingAlbertMaxSliceLen {
			return fmt.Errorf("kingalbert: foundation pile exceeds maximum allowed size")
		}
	}

	ka.trumpCards = j.TrumpCards
	if ka.trumpCards == nil {
		ka.trumpCards = NewTrumpCardsWithDecks(1, 0)
	}
	ka.tableau = j.Tableau
	for i := range KingAlbertTableauCnt {
		if ka.tableau[i] == nil {
			ka.tableau[i] = make([]*KingAlbertTableauCard, 0)
			continue
		}
		for _, tc := range ka.tableau[i] {
			if tc == nil || tc.Card == nil {
				return fmt.Errorf("kingalbert: tableau contains a nil card")
			}
		}
	}
	ka.reserve = j.Reserve
	if ka.reserve == nil {
		ka.reserve = make([]*Card, 0)
	}
	ka.foundation = j.Foundation
	for i := range KingAlbertFoundationCnt {
		if ka.foundation[i] == nil {
			ka.foundation[i] = make([]*Card, 0)
			continue
		}
		for _, c := range ka.foundation[i] {
			if c == nil {
				return fmt.Errorf("kingalbert: foundation contains a nil card")
			}
		}
	}
	ka.phase = j.Phase
	ka.moveCount = j.MoveCount
	ka.actionLog = j.ActionLog
	if ka.actionLog == nil {
		ka.actionLog = make([]*ActionLogEntry, 0)
	}
	ka.history = j.History
	if ka.history == nil {
		ka.history = make([]*kingAlbertSnapshot, 0)
	}
	ka.isStalemate = j.IsStalemate
	return nil
}
