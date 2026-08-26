//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// BristolPhase ブリストルゲームフェーズ
type BristolPhase int

// Bristolのフェーズ定数
const (
	// BristolPhasePlaying プレイ中
	BristolPhasePlaying BristolPhase = iota
	// BristolPhaseGameClear ゲームクリア
	BristolPhaseGameClear
	// BristolPhaseGameOver ゲームオーバー
	BristolPhaseGameOver
)

// BristolTableauCnt タブローの列数（8列）
const BristolTableauCnt = 8

// BristolTableauInitial タブロー1列あたりの初期配り枚数（3枚）
const BristolTableauInitial = 3

// BristolFanCnt ファン（予備列）の数（3つ）
const BristolFanCnt = 3

// BristolFoundationCnt ファウンデーションの数（4つ）
const BristolFoundationCnt = 4

// BristolHint ヒント
type BristolHint struct {
	// FromZone は移動元ゾーン ("tableau" / "fan")。
	FromZone string
	// FromCol は移動元の列/ファンのインデックス。
	FromCol int
	// ToZone は移動先ゾーン ("tableau" / "foundation")。
	ToZone string
	// ToCol は移動先の列/ファウンデーションのインデックス。
	ToCol int
}

// Bristol ブリストルソリティアゲームクラス
//
// ルール概要:
//   - タブロー8列に表向きで3枚ずつ（計24枚）。
//   - 残り28枚はストック。Draw で3つのファンに1枚ずつ（計3枚）配られる。
//   - タブローの最上段、またはファンの最上段を、他のタブロー列（スート不問・降順）へ移動できる。
//   - 一度空になったタブロー列にはカードを置けない。
//   - 4つのファウンデーションはAからKへ昇順（スート不問）。
type Bristol struct {
	trumpCards *TrumpCards
	tableau    [BristolTableauCnt][]*Card
	fan        [BristolFanCnt][]*Card
	stock      []*Card
	foundation [BristolFoundationCnt][]*Card
	phase      BristolPhase
	moveCount  int
	actionLogBase
	history []*bristolSnapshot
}

// bristolSnapshot アンドゥ用スナップショット
type bristolSnapshot struct {
	tableau     [BristolTableauCnt][]*Card
	fan         [BristolFanCnt][]*Card
	stock       []*Card
	foundation  [BristolFoundationCnt][]*Card
	phase       BristolPhase
	moveCount   int
	isStalemate bool
}

// NewBristol コンストラクタ
func NewBristol(trumpCards *TrumpCards) *Bristol {
	return &Bristol{trumpCards: trumpCards}
}

// NewDefaultBristol returns Bristol with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultBristol() *Bristol {
	return NewBristol(NewTrumpCards(0))
}

// Reset ゲームリセット
func (b *Bristol) Reset() {
	b.trumpCards.Shuffle()
	b.phase = BristolPhasePlaying
	b.moveCount = 0
	b.actionLog = nil
	b.history = nil

	// タブローに配る: 各列3枚、すべて表向き
	for i := 0; i < BristolTableauCnt; i++ {
		b.tableau[i] = make([]*Card, 0, BristolTableauInitial)
		for j := 0; j < BristolTableauInitial; j++ {
			b.tableau[i] = append(b.tableau[i], b.trumpCards.DrawCard())
		}
	}

	// ファンとファウンデーションを初期化
	for i := 0; i < BristolFanCnt; i++ {
		b.fan[i] = nil
	}
	for i := 0; i < BristolFoundationCnt; i++ {
		b.foundation[i] = nil
	}

	// 残りをストックへ（28枚）
	b.stock = nil
	for b.trumpCards.GetRemainingCount() > 0 {
		b.stock = append(b.stock, b.trumpCards.DrawCard())
	}
}

// Draw ストックから3つのファンに1枚ずつ配る（ストックが3枚未満なら配れる分だけ）
func (b *Bristol) Draw() error {
	if b.phase != BristolPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(b.stock) == 0 {
		return errors.New("no cards in stock")
	}
	b.takeSnapshot()
	drawn := make([]*Card, 0, BristolFanCnt)
	for i := 0; i < BristolFanCnt && len(b.stock) > 0; i++ {
		card := b.stock[len(b.stock)-1]
		b.stock = b.stock[:len(b.stock)-1]
		b.fan[i] = append(b.fan[i], card)
		drawn = append(drawn, card)
	}
	b.moveCount++
	b.appendLog("draw", "ストックからファンにカードを配りました", drawn)
	return nil
}

// MoveTableauToTableau タブローからタブローにカードを移動（最上段1枚のみ）
func (b *Bristol) MoveTableauToTableau(fromCol, toCol int) error {
	if b.phase != BristolPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= BristolTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= BristolTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := b.tableau[fromCol]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	card := fromCards[len(fromCards)-1]
	if !b.canPlaceOnTableau(card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	b.takeSnapshot()
	b.tableau[toCol] = append(b.tableau[toCol], card)
	b.tableau[fromCol] = fromCards[:len(fromCards)-1]
	b.moveCount++
	b.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), []*Card{card})
	return nil
}

// MoveTableauToFoundation タブローからファウンデーションにカードを移動（最上段1枚）
func (b *Bristol) MoveTableauToFoundation(col int) error {
	if b.phase != BristolPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= BristolTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := b.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	card := fromCards[len(fromCards)-1]
	fIdx := b.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	b.takeSnapshot()
	b.tableau[col] = fromCards[:len(fromCards)-1]
	b.foundation[fIdx] = append(b.foundation[fIdx], card)
	b.moveCount++
	b.appendLog("move", fmt.Sprintf("タブロー列%d→ファウンデーション", col), []*Card{card})
	b.checkGameClear()
	return nil
}

// MoveFanToTableau ファンからタブローにカードを移動（最上段1枚）
func (b *Bristol) MoveFanToTableau(fanIdx, toCol int) error {
	if b.phase != BristolPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fanIdx < 0 || fanIdx >= BristolFanCnt {
		return errors.New("invalid fan index")
	}
	if toCol < 0 || toCol >= BristolTableauCnt {
		return errors.New("invalid to column")
	}
	pile := b.fan[fanIdx]
	if len(pile) == 0 {
		return errors.New("fan is empty")
	}
	card := pile[len(pile)-1]
	if !b.canPlaceOnTableau(card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	b.takeSnapshot()
	b.fan[fanIdx] = pile[:len(pile)-1]
	b.tableau[toCol] = append(b.tableau[toCol], card)
	b.moveCount++
	b.appendLog("move", fmt.Sprintf("ファン%d→タブロー列%d", fanIdx, toCol), []*Card{card})
	return nil
}

// MoveFanToFoundation ファンからファウンデーションにカードを移動（最上段1枚）
func (b *Bristol) MoveFanToFoundation(fanIdx int) error {
	if b.phase != BristolPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fanIdx < 0 || fanIdx >= BristolFanCnt {
		return errors.New("invalid fan index")
	}
	pile := b.fan[fanIdx]
	if len(pile) == 0 {
		return errors.New("fan is empty")
	}
	card := pile[len(pile)-1]
	fIdx := b.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	b.takeSnapshot()
	b.fan[fanIdx] = pile[:len(pile)-1]
	b.foundation[fIdx] = append(b.foundation[fIdx], card)
	b.moveCount++
	b.appendLog("move", fmt.Sprintf("ファン%d→ファウンデーション", fanIdx), []*Card{card})
	b.checkGameClear()
	return nil
}

// GiveUp ギブアップ
func (b *Bristol) GiveUp() {
	if b.phase == BristolPhasePlaying {
		b.phase = BristolPhaseGameOver
		b.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (b *Bristol) GetHint() *BristolHint {
	if b.phase != BristolPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファウンデーションへ
	for col := 0; col < BristolTableauCnt; col++ {
		pile := b.tableau[col]
		if len(pile) == 0 {
			continue
		}
		if fIdx := b.findFoundation(pile[len(pile)-1]); fIdx >= 0 {
			return &BristolHint{FromZone: "tableau", FromCol: col, ToZone: "foundation", ToCol: fIdx}
		}
	}
	// 優先度2: ファンからファウンデーションへ
	for fanIdx := 0; fanIdx < BristolFanCnt; fanIdx++ {
		pile := b.fan[fanIdx]
		if len(pile) == 0 {
			continue
		}
		if fIdx := b.findFoundation(pile[len(pile)-1]); fIdx >= 0 {
			return &BristolHint{FromZone: "fan", FromCol: fanIdx, ToZone: "foundation", ToCol: fIdx}
		}
	}
	// 優先度3: タブローからタブローへ（空列への移動は無意味なので除外）
	for fromCol := 0; fromCol < BristolTableauCnt; fromCol++ {
		pile := b.tableau[fromCol]
		if len(pile) == 0 {
			continue
		}
		card := pile[len(pile)-1]
		for toCol := 0; toCol < BristolTableauCnt; toCol++ {
			if toCol == fromCol {
				continue
			}
			if b.canPlaceOnTableau(card, toCol) {
				return &BristolHint{FromZone: "tableau", FromCol: fromCol, ToZone: "tableau", ToCol: toCol}
			}
		}
	}
	// 優先度4: ファンからタブローへ
	for fanIdx := 0; fanIdx < BristolFanCnt; fanIdx++ {
		pile := b.fan[fanIdx]
		if len(pile) == 0 {
			continue
		}
		card := pile[len(pile)-1]
		for toCol := 0; toCol < BristolTableauCnt; toCol++ {
			if b.canPlaceOnTableau(card, toCol) {
				return &BristolHint{FromZone: "fan", FromCol: fanIdx, ToZone: "tableau", ToCol: toCol}
			}
		}
	}
	return nil
}

// AutoComplete オートコンプリート（移動可能なカードをファウンデーションへ繰り返し移動）
func (b *Bristol) AutoComplete() error {
	if b.phase != BristolPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	b.takeSnapshot()
	moved := false
	for {
		stepMoved := false
		// タブローの各列トップ
		for col := 0; col < BristolTableauCnt; col++ {
			pile := b.tableau[col]
			if len(pile) == 0 {
				continue
			}
			card := pile[len(pile)-1]
			if fIdx := b.findFoundation(card); fIdx >= 0 {
				b.tableau[col] = pile[:len(pile)-1]
				b.foundation[fIdx] = append(b.foundation[fIdx], card)
				b.moveCount++
				stepMoved = true
			}
		}
		// ファンの各トップ
		for fanIdx := 0; fanIdx < BristolFanCnt; fanIdx++ {
			pile := b.fan[fanIdx]
			if len(pile) == 0 {
				continue
			}
			card := pile[len(pile)-1]
			if fIdx := b.findFoundation(card); fIdx >= 0 {
				b.fan[fanIdx] = pile[:len(pile)-1]
				b.foundation[fIdx] = append(b.foundation[fIdx], card)
				b.moveCount++
				stepMoved = true
			}
		}
		if !stepMoved {
			break
		}
		moved = true
	}
	if !moved {
		// 何も動かなければスナップショットを取り消す
		b.history = b.history[:len(b.history)-1]
		return errors.New("no card can be auto-completed")
	}
	b.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	b.checkGameClear()
	return nil
}

// --- Getters / Setters ---

// GetPhase フェーズ取得
func (b *Bristol) GetPhase() BristolPhase { return b.phase }

// SetPhase フェーズ設定 (テスト用)
func (b *Bristol) SetPhase(p BristolPhase) { b.phase = p }

// GetMoveCount 移動回数取得
func (b *Bristol) GetMoveCount() int { return b.moveCount }

// GetStockCount ストック残枚数
func (b *Bristol) GetStockCount() int { return len(b.stock) }

// GetTableau タブロー取得
func (b *Bristol) GetTableau() [BristolTableauCnt][]*Card { return b.tableau }

// GetFan ファン取得（3つ）
func (b *Bristol) GetFan() [BristolFanCnt][]*Card { return b.fan }

// GetFoundation ファウンデーション取得
func (b *Bristol) GetFoundation() [BristolFoundationCnt][]*Card { return b.foundation }

// GetGameEndFlag returns true once the game has left the playing phase.
func (b *Bristol) GetGameEndFlag() bool { return b.phase != BristolPhasePlaying }

// SetStock ストック設定 (テスト用)
func (b *Bristol) SetStock(s []*Card) { b.stock = s }

// SetTableau タブロー設定 (テスト用)
func (b *Bristol) SetTableau(t [BristolTableauCnt][]*Card) { b.tableau = t }

// SetFan ファン設定 (テスト用)
func (b *Bristol) SetFan(f [BristolFanCnt][]*Card) { b.fan = f }

// SetFoundation ファウンデーション設定 (テスト用)
func (b *Bristol) SetFoundation(f [BristolFoundationCnt][]*Card) { b.foundation = f }

// Undo 直前の操作を取り消す
// IsStalemate は合法手が 1 つも無い状態かを返す。
//
// **ストックは作り直せない。**Draw() で 3 つのファンへ配り切ると空になるので、
// どこにも置けない盤面に普通に到達する。他のソリティアと違って検知が無く、
// プレイヤーは動かせる札を探し続けるしかなかった (#5631)。
//
// **フィールドに覚えない。**覚えると更新する場所を数え上げることになり、
// 1 つ漏れただけで永遠に false を返す (この機能の初版が Draw() を漏らした)。
// 判定は GetHint に委ねる ── 「動かせる手があるか」を 2 か所で数えると、
// ヒントは出るのに手詰まりと言う状態が作れる。
func (b *Bristol) IsStalemate() bool {
	if b.phase != BristolPhasePlaying {
		return false
	}
	if len(b.stock) > 0 {
		// 配れる札が残っていれば、まだ手はある。
		return false
	}
	return b.GetHint() == nil
}

// UndoToEscape は膠着状態から抜けるために必要なアンドゥ回数を返す。
// 膠着状態でなければ 0、脱出不可なら -1。
func (b *Bristol) UndoToEscape() int {
	return undoToEscape(b.IsStalemate(), b.history, func(s *bristolSnapshot) bool { return s.isStalemate })
}

func (b *Bristol) Undo() error {
	if b.phase != BristolPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(b.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := b.history[len(b.history)-1]
	b.history = b.history[:len(b.history)-1]
	b.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能か
func (b *Bristol) CanUndo() bool {
	return len(b.history) > 0 && b.phase == BristolPhasePlaying
}

// UndoN n回連続アンドゥ
func (b *Bristol) UndoN(n int) error {
	return undoN(b, n)
}

// --- Private helpers ---

// canPlaceOnTableau はカードを toCol のタブローに置けるか判定する。
//
// ブリストルのルール:
//   - 空のタブロー列にはカードを置けない（一度空になった列は使えない）。
//   - それ以外はスート不問で、最上段より1つ小さいランクのカードを置ける（降順）。
func (b *Bristol) canPlaceOnTableau(card *Card, col int) bool {
	colCards := b.tableau[col]
	if len(colCards) == 0 {
		return false
	}
	top := colCards[len(colCards)-1]
	return card.GetValue() == top.GetValue()-1
}

// canPlaceOnFoundation はカードを fIdx のファウンデーションに置けるか判定する（スート不問でAからK昇順）。
func (b *Bristol) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := b.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	top := pile[len(pile)-1]
	return card.GetValue() == top.GetValue()+1
}

// LegalTargets は指定した移動元の一番上の札を置ける先を返す。
// tableau は列番号、foundation はファウンデーション番号。
//
// **選んだ札で実際に動かせる先だけを示すため。**画面は選択中に全ての移動先を
// 同じ見た目で強調していて、押すまで合法か分からなかった (#4813)。判定は
// canPlaceOnTableau / canPlaceOnFoundation をそのまま使う。
func (b *Bristol) LegalTargets(fromZone string, fromCol int) (tableau []int, foundation []int) {
	var card *Card
	switch fromZone {
	case "tableau":
		if fromCol < 0 || fromCol >= BristolTableauCnt || len(b.tableau[fromCol]) == 0 {
			return nil, nil
		}
		card = b.tableau[fromCol][len(b.tableau[fromCol])-1]
	case "fan":
		if fromCol < 0 || fromCol >= BristolFanCnt || len(b.fan[fromCol]) == 0 {
			return nil, nil
		}
		card = b.fan[fromCol][len(b.fan[fromCol])-1]
	default:
		return nil, nil
	}
	for col := 0; col < BristolTableauCnt; col++ {
		// 自分の列は canPlaceOnTableau が既に弾く (自分自身の 1 つ下にはならない)。
		// それでも明示するのは、読み手が「同じ列が候補に出るのでは」と疑わずに済むため。
		if col == fromCol && fromZone == "tableau" {
			continue
		}
		if b.canPlaceOnTableau(card, col) {
			tableau = append(tableau, col)
		}
	}
	for f := 0; f < BristolFoundationCnt; f++ {
		if b.canPlaceOnFoundation(card, f) {
			foundation = append(foundation, f)
		}
	}
	return tableau, foundation
}

// findFoundation はカードを置けるファウンデーションのインデックスを返す（無ければ -1）。
func (b *Bristol) findFoundation(card *Card) int {
	for i := 0; i < BristolFoundationCnt; i++ {
		if b.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定（全ファウンデーションがKまで揃ったら）
func (b *Bristol) checkGameClear() {
	for i := 0; i < BristolFoundationCnt; i++ {
		if len(b.foundation[i]) != CardValueMax {
			return
		}
	}
	b.phase = BristolPhaseGameClear
}

func (b *Bristol) takeSnapshot() {
	snap := &bristolSnapshot{
		phase:     b.phase,
		moveCount: b.moveCount,
		// **記録は撮った時点の生の判定。**キャッシュを持つと更新漏れで古くなる
		// (この PR の初版がまさにそれで、Draw() 後に更新されなかった)。
		isStalemate: b.IsStalemate(),
	}
	for i := 0; i < BristolTableauCnt; i++ {
		snap.tableau[i] = make([]*Card, len(b.tableau[i]))
		copy(snap.tableau[i], b.tableau[i])
	}
	for i := 0; i < BristolFanCnt; i++ {
		snap.fan[i] = make([]*Card, len(b.fan[i]))
		copy(snap.fan[i], b.fan[i])
	}
	snap.stock = make([]*Card, len(b.stock))
	copy(snap.stock, b.stock)
	for i := 0; i < BristolFoundationCnt; i++ {
		snap.foundation[i] = make([]*Card, len(b.foundation[i]))
		copy(snap.foundation[i], b.foundation[i])
	}
	b.history = appendSnapshot(b.history, snap)
}

func (b *Bristol) restoreSnapshot(snap *bristolSnapshot) {
	b.tableau = snap.tableau
	b.fan = snap.fan
	b.stock = snap.stock
	b.foundation = snap.foundation
	b.phase = snap.phase
	b.moveCount = snap.moveCount
}

func (b *Bristol) appendLog(actionType, detail string, cards []*Card) {
	b.appendLogAt(b.moveCount, 0, actionType, detail, cards)
}

// bristolJSON is the JSON wire format for Bristol.
type bristolJSON struct {
	TrumpCards *TrumpCards                   `json:"tc"`
	Tableau    [BristolTableauCnt][]*Card    `json:"tb"`
	Fan        [BristolFanCnt][]*Card        `json:"fn"`
	Stock      []*Card                       `json:"st"`
	Foundation [BristolFoundationCnt][]*Card `json:"fd"`
	Phase      BristolPhase                  `json:"ps"`
	MoveCount  int                           `json:"mc"`
	ActionLog  []*ActionLogEntry             `json:"al"`
	History    []*bristolSnapshot            `json:"hi,omitempty"`
}

// bristolSnapshotJSON is the wire format for a single undo snapshot.
type bristolSnapshotJSON struct {
	Tableau    [BristolTableauCnt][]*Card    `json:"tb"`
	Fan        [BristolFanCnt][]*Card        `json:"fn"`
	Stock      []*Card                       `json:"st"`
	Foundation [BristolFoundationCnt][]*Card `json:"fd"`
	Phase      BristolPhase                  `json:"ps"`
	MoveCount  int                           `json:"mc"`
}

// bristolMaxSliceLen caps slice sizes during deserialisation.
const bristolMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler for bristolSnapshot.
func (s *bristolSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(bristolSnapshotJSON{
		Tableau:    s.tableau,
		Fan:        s.fan,
		Stock:      s.stock,
		Foundation: s.foundation,
		Phase:      s.phase,
		MoveCount:  s.moveCount,
	})
}

// UnmarshalJSON implements json.Unmarshaler for bristolSnapshot.
func (s *bristolSnapshot) UnmarshalJSON(data []byte) error {
	var j bristolSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > bristolMaxSliceLen {
		return errBristolTooLarge
	}
	for _, col := range j.Tableau {
		if len(col) > bristolMaxSliceLen {
			return errBristolTooLarge
		}
	}
	for _, pile := range j.Fan {
		if len(pile) > bristolMaxSliceLen {
			return errBristolTooLarge
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > bristolMaxSliceLen {
			return errBristolTooLarge
		}
	}
	s.tableau = bristolNormalizeCols(j.Tableau)
	s.fan = bristolNormalizeFans(j.Fan)
	s.stock = bristolNonNil(j.Stock)
	for i := 0; i < BristolFoundationCnt; i++ {
		s.foundation[i] = bristolNonNil(j.Foundation[i])
	}
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	return nil
}

// MarshalJSON implements json.Marshaler.
func (b *Bristol) MarshalJSON() ([]byte, error) {
	return json.Marshal(bristolJSON{
		TrumpCards: b.trumpCards,
		Tableau:    b.tableau,
		Fan:        b.fan,
		Stock:      b.stock,
		Foundation: b.foundation,
		Phase:      b.phase,
		MoveCount:  b.moveCount,
		ActionLog:  b.actionLog,
		History:    b.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Bristol) UnmarshalJSON(data []byte) error {
	var j bristolJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > bristolMaxSliceLen ||
		len(j.ActionLog) > bristolMaxSliceLen || len(j.History) > bristolMaxSliceLen {
		return errBristolTooLarge
	}
	for _, col := range j.Tableau {
		if len(col) > bristolMaxSliceLen {
			return errBristolTooLarge
		}
	}
	for _, pile := range j.Fan {
		if len(pile) > bristolMaxSliceLen {
			return errBristolTooLarge
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > bristolMaxSliceLen {
			return errBristolTooLarge
		}
	}
	b.trumpCards = j.TrumpCards
	if b.trumpCards == nil {
		b.trumpCards = NewTrumpCards(0)
	}
	b.tableau = bristolNormalizeCols(j.Tableau)
	b.fan = bristolNormalizeFans(j.Fan)
	b.stock = bristolNonNil(j.Stock)
	for i := 0; i < BristolFoundationCnt; i++ {
		b.foundation[i] = bristolNonNil(j.Foundation[i])
	}
	b.phase = j.Phase
	b.moveCount = j.MoveCount
	b.actionLog = j.ActionLog
	if b.actionLog == nil {
		b.actionLog = make([]*ActionLogEntry, 0)
	}
	b.history = j.History
	if b.history == nil {
		b.history = make([]*bristolSnapshot, 0)
	}
	return nil
}

// errBristolTooLarge は逆シリアライズ時のサイズ上限超過を表す共有センチネルエラー。
var errBristolTooLarge = errors.New("bristol: input array exceeds maximum allowed size")

// bristolNonNil returns a non-nil slice with nil elements removed.
func bristolNonNil(s []*Card) []*Card {
	res := make([]*Card, 0, len(s))
	for _, c := range s {
		if c != nil {
			res = append(res, c)
		}
	}
	return res
}

// bristolNormalizeCols ensures every tableau column is non-nil and free of nil elements.
func bristolNormalizeCols(t [BristolTableauCnt][]*Card) [BristolTableauCnt][]*Card {
	for i := 0; i < BristolTableauCnt; i++ {
		t[i] = bristolNonNil(t[i])
	}
	return t
}

// bristolNormalizeFans ensures every fan is non-nil and free of nil elements.
func bristolNormalizeFans(f [BristolFanCnt][]*Card) [BristolFanCnt][]*Card {
	for i := 0; i < BristolFanCnt; i++ {
		f[i] = bristolNonNil(f[i])
	}
	return f
}
