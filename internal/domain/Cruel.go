//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// CruelPhase クルーエルゲームフェーズ
type CruelPhase int

// Cruelのフェーズ定数
const (
	// CruelPhasePlaying プレイ中
	CruelPhasePlaying CruelPhase = iota
	// CruelPhaseGameClear ゲームクリア
	CruelPhaseGameClear
	// CruelPhaseGameOver ゲームオーバー
	CruelPhaseGameOver
)

// CruelTableauCnt タブローの列数（クルーエルは12列）
const CruelTableauCnt = 12

// CruelFoundationCnt ファンデーションの数
const CruelFoundationCnt = 4

// CruelInitialColSize 各タブロー列の初期カード枚数
const CruelInitialColSize = 4

// CruelHint ヒント
type CruelHint struct {
	FromCol   int
	CardIndex int
	ToZone    string // "tableau" or "foundation"
	ToCol     int
}

// Cruel クルーエルゲームクラス
//
// クラシックなパズル系ソリティア。Aは初期段階でファウンデーションに自動配置され、
// 残り48枚を12列×4枚のタブローに表向きで配る。カード移動は同スートの降順のみ、
// 空の列にはカードを置けない。手詰まりのときは Shift で盤面を再構築できる。
type Cruel struct {
	trumpCards *TrumpCards
	tableau    [CruelTableauCnt][]*KlondikeTableauCard
	foundation [CruelFoundationCnt][]*Card
	phase      CruelPhase
	moveCount  int
	actionLogBase
	history     []*cruelSnapshot
	isStalemate bool
}

// cruelSnapshot アンドゥ用スナップショット
type cruelSnapshot struct {
	tableau     [CruelTableauCnt][]*KlondikeTableauCard
	foundation  [CruelFoundationCnt][]*Card
	phase       CruelPhase
	moveCount   int
	isStalemate bool
}

// NewCruel コンストラクタ
func NewCruel(trumpCards *TrumpCards) *Cruel {
	return &Cruel{
		trumpCards: trumpCards,
	}
}

// NewDefaultCruel returns Cruel with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultCruel() *Cruel {
	return NewCruel(NewTrumpCards(0))
}

// Reset ゲームリセット
func (c *Cruel) Reset() {
	c.trumpCards.Shuffle()
	c.phase = CruelPhasePlaying
	c.moveCount = 0
	c.actionLog = nil
	c.history = nil
	c.isStalemate = false

	// 配り直しに使うために52枚すべてをドロー。
	// AはファウンデーションへAutoplace、残り48枚はタブローへ。
	aces := make([]*Card, 0, CruelFoundationCnt)
	rest := make([]*Card, 0, CardCnt-CruelFoundationCnt)
	for range CardCnt {
		card := c.trumpCards.DrawCard()
		if card == nil {
			continue
		}
		if card.GetValue() == 1 {
			aces = append(aces, card)
		} else {
			rest = append(rest, card)
		}
	}

	// ファウンデーション初期化（Aを各スートに配置）
	for i := range CruelFoundationCnt {
		c.foundation[i] = nil
	}
	for _, a := range aces {
		fIdx := a.GetDesign() - 1
		if fIdx >= 0 && fIdx < CruelFoundationCnt {
			c.foundation[fIdx] = []*Card{a}
		}
	}

	// タブロー初期化（12列×4枚、左から順に表向きで配る）
	for i := range CruelTableauCnt {
		c.tableau[i] = make([]*KlondikeTableauCard, 0, CruelInitialColSize)
	}
	idx := 0
	for col := range CruelTableauCnt {
		for range CruelInitialColSize {
			if idx >= len(rest) {
				break
			}
			c.tableau[col] = append(c.tableau[col], &KlondikeTableauCard{
				Card:   rest[idx],
				FaceUp: true,
			})
			idx++
		}
	}
}

// MoveTableauToTableau タブローからタブローにカードを移動
//
// 移動できるのは各列の最上段の1枚のみ。移動先の列が空であれば不可、
// 移動先の最上段カードと同スートで値が1つ小さい場合のみ移動可能。
func (c *Cruel) MoveTableauToTableau(fromCol, toCol int) error {
	if c.phase != CruelPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= CruelTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= CruelTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := c.tableau[fromCol]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	if !c.canPlaceOnTableau(tc.Card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	c.takeSnapshot()
	c.tableau[fromCol] = fromCards[:len(fromCards)-1]
	c.tableau[toCol] = append(c.tableau[toCol], tc)
	c.moveCount++
	c.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), []*Card{tc.Card})
	c.checkStalemate()
	return nil
}

// MoveTableauToFoundation タブロー最上段のカードをファウンデーションへ移動
func (c *Cruel) MoveTableauToFoundation(col int) error {
	if c.phase != CruelPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= CruelTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := c.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	card := tc.Card
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= CruelFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !c.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	c.takeSnapshot()
	c.tableau[col] = fromCards[:len(fromCards)-1]
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.moveCount++
	c.appendLog("move", fmt.Sprintf("タブロー列%d→ファウンデーション", col), []*Card{card})
	c.checkGameClear()
	c.checkStalemate()
	return nil
}

// Shift タブローに残っているカードを左から順に集め直し、再度12列×4枚に配り直す。
// 列の順序とカードの相対順序は保持される。スナップショットを取るため Undo で
// 元に戻せる。手詰まりの脱出手段として無制限に使用できる。
func (c *Cruel) Shift() error {
	if c.phase != CruelPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	// 1枚もタブローに残っていない場合はシフトする意味がない。
	total := 0
	for i := range CruelTableauCnt {
		total += len(c.tableau[i])
	}
	if total == 0 {
		return errors.New("no tableau cards to shift")
	}

	c.takeSnapshot()

	// 1) 平坦化：列0から列11まで、各列内の順序を維持して連結。
	pile := make([]*KlondikeTableauCard, 0, total)
	for col := range CruelTableauCnt {
		pile = append(pile, c.tableau[col]...)
		c.tableau[col] = c.tableau[col][:0]
	}

	// 2) 再配布：先頭から4枚ずつ列0..11へ。残り枚数が4未満の列は左詰めで配る。
	idx := 0
	for col := range CruelTableauCnt {
		for range CruelInitialColSize {
			if idx >= len(pile) {
				break
			}
			c.tableau[col] = append(c.tableau[col], pile[idx])
			idx++
		}
		if idx >= len(pile) {
			break
		}
	}

	c.moveCount++
	c.appendLog("shift", "盤面を再構築しました", nil)
	c.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (c *Cruel) GiveUp() {
	if c.phase == CruelPhasePlaying {
		c.phase = CruelPhaseGameOver
		c.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
//
// 優先度1: タブロー→ファウンデーション
// 優先度2: タブロー→タブロー（同スート降順で置ける移動）
func (c *Cruel) GetHint() *CruelHint {
	if c.phase != CruelPhasePlaying {
		return nil
	}
	for col := range CruelTableauCnt {
		if len(c.tableau[col]) == 0 {
			continue
		}
		card := c.tableau[col][len(c.tableau[col])-1].Card
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < CruelFoundationCnt && c.canPlaceOnFoundation(card, fIdx) {
			return &CruelHint{
				FromCol:   col,
				CardIndex: len(c.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	for fromCol := range CruelTableauCnt {
		if len(c.tableau[fromCol]) == 0 {
			continue
		}
		card := c.tableau[fromCol][len(c.tableau[fromCol])-1].Card
		for toCol := range CruelTableauCnt {
			if toCol == fromCol {
				continue
			}
			if c.canPlaceOnTableau(card, toCol) {
				return &CruelHint{
					FromCol:   fromCol,
					CardIndex: len(c.tableau[fromCol]) - 1,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	return nil
}

// AutoComplete 全カードがファウンデーションに昇順で送れる状態なら自動で送る。
func (c *Cruel) AutoComplete() error {
	if c.phase != CruelPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	c.takeSnapshot()
	for {
		moved := false
		for col := range CruelTableauCnt {
			if len(c.tableau[col]) == 0 {
				continue
			}
			card := c.tableau[col][len(c.tableau[col])-1].Card
			fIdx := card.GetDesign() - 1
			if fIdx < 0 || fIdx >= CruelFoundationCnt {
				continue
			}
			if !c.canPlaceOnFoundation(card, fIdx) {
				continue
			}
			c.tableau[col] = c.tableau[col][:len(c.tableau[col])-1]
			c.foundation[fIdx] = append(c.foundation[fIdx], card)
			c.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	c.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	c.checkGameClear()
	// AutoComplete can partially clear the board (e.g. some suits reach the
	// foundation but others remain blocked). Refresh the stalemate flag so
	// the UI doesn't show stale "stuck" / "escape" hints after a partial run.
	if c.phase == CruelPhasePlaying {
		c.checkStalemate()
	}
	return nil
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (c *Cruel) GetPhase() CruelPhase { return c.phase }

// SetPhase フェーズ設定（テスト用）
func (c *Cruel) SetPhase(phase CruelPhase) { c.phase = phase }

// GetMoveCount 移動回数取得
func (c *Cruel) GetMoveCount() int { return c.moveCount }

// GetTableau タブロー取得
func (c *Cruel) GetTableau() [CruelTableauCnt][]*KlondikeTableauCard {
	return c.tableau
}

// GetFoundation ファウンデーション取得
func (c *Cruel) GetFoundation() [CruelFoundationCnt][]*Card {
	return c.foundation
}

// GetGameEndFlag returns true once the game has left the playing phase.
func (c *Cruel) GetGameEndFlag() bool { return c.phase != CruelPhasePlaying }

// IsStalemate 手詰まり状態取得
func (c *Cruel) IsStalemate() bool { return c.isStalemate }

// SetIsStalemate 手詰まり状態設定（テスト用）
func (c *Cruel) SetIsStalemate(v bool) { c.isStalemate = v }

// SetTableau タブロー設定（テスト用）
func (c *Cruel) SetTableau(tableau [CruelTableauCnt][]*KlondikeTableauCard) {
	c.tableau = tableau
}

// SetFoundation ファウンデーション設定（テスト用）
func (c *Cruel) SetFoundation(foundation [CruelFoundationCnt][]*Card) {
	c.foundation = foundation
}

// Undo 直前の操作を取り消す
func (c *Cruel) Undo() error {
	if c.phase != CruelPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(c.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := c.history[len(c.history)-1]
	c.history = c.history[:len(c.history)-1]
	c.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (c *Cruel) CanUndo() bool {
	return len(c.history) > 0 && c.phase == CruelPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。
// 膠着状態でなければ 0、脱出不可なら -1。
func (c *Cruel) UndoToEscape() int {
	return undoToEscape(c.isStalemate, c.history, func(s *cruelSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (c *Cruel) UndoN(n int) error {
	return undoN(c, n)
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定。
// 空の列には置けず、同スートで値が1つ小さい場合のみ可。
func (c *Cruel) canPlaceOnTableau(card *Card, col int) bool {
	colCards := c.tableau[col]
	if len(colCards) == 0 {
		return false
	}
	topCard := colCards[len(colCards)-1].Card
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()-1
}

// canPlaceOnFoundation ファウンデーションにカードを置けるか判定。
// Reset() で各ファウンデーションに A を1枚配置済みなので、空のケースは通常発生しない。
func (c *Cruel) canPlaceOnFoundation(card *Card, fIdx int) bool {
	return canPlaceOnFoundationPile(c.foundation[fIdx], card)
}

// checkGameClear ゲームクリア判定
func (c *Cruel) checkGameClear() {
	for i := range CruelFoundationCnt {
		if len(c.foundation[i]) != CardValueMax {
			return
		}
	}
	c.phase = CruelPhaseGameClear
}

// checkStalemate 手詰まり判定（タブロー間移動・ファウンデーション送りが
// すべて不可能なら手詰まり）。Shift は手詰まり状態でも可能なので、
// 「真の終局」ではなく「ユーザーに Shift を促す」ためのフラグとして使う。
func (c *Cruel) checkStalemate() {
	if c.phase != CruelPhasePlaying {
		return
	}
	if c.GetHint() != nil {
		c.isStalemate = false
		return
	}
	c.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (c *Cruel) takeSnapshot() {
	snap := &cruelSnapshot{
		phase:       c.phase,
		moveCount:   c.moveCount,
		isStalemate: c.isStalemate,
	}
	for i := range CruelTableauCnt {
		snap.tableau[i] = make([]*KlondikeTableauCard, len(c.tableau[i]))
		for j, tc := range c.tableau[i] {
			snap.tableau[i][j] = &KlondikeTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	for i := range CruelFoundationCnt {
		snap.foundation[i] = make([]*Card, len(c.foundation[i]))
		copy(snap.foundation[i], c.foundation[i])
	}
	c.history = append(c.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (c *Cruel) restoreSnapshot(snap *cruelSnapshot) {
	c.tableau = snap.tableau
	c.foundation = snap.foundation
	c.phase = snap.phase
	c.moveCount = snap.moveCount
	c.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (c *Cruel) appendLog(actionType, detail string, cards []*Card) {
	c.appendLogAt(c.moveCount, 0, actionType, detail, cards)
}

// cruelJSON is the JSON wire format for Cruel.
type cruelJSON struct {
	TrumpCards  *TrumpCards                             `json:"tc"`
	Tableau     [CruelTableauCnt][]*KlondikeTableauCard `json:"tb"`
	Foundation  [CruelFoundationCnt][]*Card             `json:"fd"`
	Phase       CruelPhase                              `json:"ps"`
	MoveCount   int                                     `json:"mc"`
	ActionLog   []*ActionLogEntry                       `json:"al"`
	IsStalemate bool                                    `json:"sl"`
	History     []*cruelSnapshot                        `json:"hi,omitempty"`
}

// cruelSnapshotJSON is the wire format for a single undo snapshot.
// cruelSnapshot uses unexported fields, so we project to/from this shape with
// explicit Marshal/Unmarshal methods.
type cruelSnapshotJSON struct {
	Tableau     [CruelTableauCnt][]*KlondikeTableauCard `json:"tb"`
	Foundation  [CruelFoundationCnt][]*Card             `json:"fd"`
	Phase       CruelPhase                              `json:"ps"`
	MoveCount   int                                     `json:"mc"`
	IsStalemate bool                                    `json:"sl"`
}

// MarshalJSON implements json.Marshaler for cruelSnapshot.
func (s *cruelSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(cruelSnapshotJSON{
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for cruelSnapshot.
func (s *cruelSnapshot) UnmarshalJSON(data []byte) error {
	var j cruelSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > cruelMaxSliceLen {
			return fmt.Errorf("cruel: snapshot tableau column exceeds maximum allowed size")
		}
		// A JSON `null` array element unmarshals to a nil *KlondikeTableauCard,
		// which would panic on dereference (e.g. in checkStalemate or the Web
		// presenter). Reject the payload up front.
		for _, tc := range col {
			if tc == nil {
				return fmt.Errorf("cruel: snapshot tableau contains nil card")
			}
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > cruelMaxSliceLen {
			return fmt.Errorf("cruel: snapshot foundation pile exceeds maximum allowed size")
		}
		for _, card := range pile {
			if card == nil {
				return fmt.Errorf("cruel: snapshot foundation contains nil card")
			}
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
func (c *Cruel) MarshalJSON() ([]byte, error) {
	return json.Marshal(cruelJSON{
		TrumpCards:  c.trumpCards,
		Tableau:     c.tableau,
		Foundation:  c.foundation,
		Phase:       c.phase,
		MoveCount:   c.moveCount,
		ActionLog:   c.actionLog,
		IsStalemate: c.isStalemate,
		History:     c.history,
	})
}

// cruelMaxSliceLen caps slice sizes during deserialisation.
const cruelMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (c *Cruel) UnmarshalJSON(data []byte) error {
	var j cruelJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > cruelMaxSliceLen ||
		len(j.History) > cruelMaxSliceLen {
		return fmt.Errorf("cruel: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > cruelMaxSliceLen {
			return fmt.Errorf("cruel: tableau column exceeds maximum allowed size")
		}
		for _, tc := range col {
			if tc == nil {
				return fmt.Errorf("cruel: tableau contains nil card")
			}
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > cruelMaxSliceLen {
			return fmt.Errorf("cruel: foundation pile exceeds maximum allowed size")
		}
		for _, card := range pile {
			if card == nil {
				return fmt.Errorf("cruel: foundation contains nil card")
			}
		}
	}

	c.trumpCards = j.TrumpCards
	if c.trumpCards == nil {
		c.trumpCards = NewTrumpCards(0)
	}
	c.tableau = j.Tableau
	c.foundation = j.Foundation
	c.phase = j.Phase
	c.moveCount = j.MoveCount
	c.actionLog = j.ActionLog
	if c.actionLog == nil {
		c.actionLog = make([]*ActionLogEntry, 0)
	}
	c.history = j.History
	if c.history == nil {
		c.history = make([]*cruelSnapshot, 0)
	}
	c.isStalemate = j.IsStalemate
	return nil
}
