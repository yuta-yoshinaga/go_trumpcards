//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// PerseverancePhase パーシビアランスゲームフェーズ
type PerseverancePhase int

// Perseveranceのフェーズ定数
const (
	// PerseverancePhasePlaying プレイ中
	PerseverancePhasePlaying PerseverancePhase = iota
	// PerseverancePhaseGameClear ゲームクリア
	PerseverancePhaseGameClear
	// PerseverancePhaseGameOver ゲームオーバー
	PerseverancePhaseGameOver
)

// PerseveranceTableauCnt タブローの列数。
//
// **13 ではなく 12。**A 4 枚を配る前にファウンデーションへ抜くので、卓に出るのは
// 48 枚しかない。クローン元の Baker's Dozen は 52 枚を 13 列に配る。
const PerseveranceTableauCnt = 12

// PerseveranceColSize 1 列に配る枚数。
const PerseveranceColSize = 4

// PerseveranceMaxRedeals 再配りの上限。
//
// Baker's Dozen には再配りが無い。Perseverance は 2 回まで、集める順は**逆順**で
// **シャッフルしない**ので、配り直した並びは直前の盤から一意に決まる。
const PerseveranceMaxRedeals = 2

// PerseveranceFoundationCnt ファンデーションの数
const PerseveranceFoundationCnt = 4

// PerseveranceTableauCard タブロー上のカード
type PerseveranceTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// PerseveranceHint ヒント
type PerseveranceHint struct {
	FromCol   int    // タブロー列インデックス
	CardIndex int    // 列内のカードインデックス
	ToZone    string // "tableau" or "foundation"
	ToCol     int    // タブロー列 or ファンデーションのインデックス
}

// PerseveranceConfig パーシビアランスゲーム設定
type PerseveranceConfig struct{}

// Perseverance パーシビアランスゲームクラス
type Perseverance struct {
	trumpCards *TrumpCards
	tableau    [PerseveranceTableauCnt][]*PerseveranceTableauCard
	foundation [PerseveranceFoundationCnt][]*Card
	phase      PerseverancePhase
	moveCount  int
	actionLogBase
	history     []*perseveranceSnapshot
	isStalemate bool
	redealsLeft int
}

// perseveranceSnapshot アンドゥ用スナップショット
type perseveranceSnapshot struct {
	tableau     [PerseveranceTableauCnt][]*PerseveranceTableauCard
	foundation  [PerseveranceFoundationCnt][]*Card
	phase       PerseverancePhase
	moveCount   int
	isStalemate bool
	redealsLeft int
}

// NewPerseverance コンストラクタ
func NewPerseverance(trumpCards *TrumpCards) *Perseverance {
	return &Perseverance{
		trumpCards: trumpCards,
	}
}

// NewDefaultPerseverance returns Perseverance with a single 52-card deck.
func NewDefaultPerseverance() *Perseverance {
	return NewPerseverance(NewTrumpCardsWithDecks(1, 0))
}

// Reset ゲームリセット
func (bd *Perseverance) Reset() {
	bd.trumpCards.Shuffle()
	bd.phase = PerseverancePhasePlaying
	bd.moveCount = 0
	bd.actionLog = nil
	bd.history = nil
	bd.isStalemate = false
	bd.redealsLeft = PerseveranceMaxRedeals

	for i := range PerseveranceFoundationCnt {
		bd.foundation[i] = nil
	}

	// **A 4 枚は配る前に抜いてファウンデーションへ置く。**Baker's Dozen は 52 枚を
	// そのまま卓に配って組札を空から始めるので、ここが最初の分岐点になる。
	// **スートの定数は 1..4 で、組札の添字は 0..3。**そのまま添字に使うと
	// ♦ (=4) で範囲外になる。
	rest := make([]*Card, 0, CardCnt-PerseveranceFoundationCnt)
	for range CardCnt {
		card := bd.trumpCards.DrawCard()
		if card == nil {
			break
		}
		if card.GetValue() == 1 {
			bd.foundation[card.GetDesign()-CardDesignSpade] = []*Card{card}
			continue
		}
		rest = append(rest, card)
	}

	bd.dealColumns(rest)

	// Detect a dead deal so the UI can offer Undo-to-Escape from move 0.
	bd.checkStalemate()
}

// dealColumns は残り札を先頭から 4 枚ずつ配り、各列の K を列の底へ沈める。
//
// 初回の配り (48 枚 = 12 列ぴったり) と再配り (48 枚以下・端数あり) の両方が通る。
// K を底へ送るのは Baker's Dozen から引き継いだ規則で、Perseverance でも変わらない。
func (bd *Perseverance) dealColumns(cards []*Card) {
	for i := range PerseveranceTableauCnt {
		bd.tableau[i] = nil
	}
	for i, card := range cards {
		col := i / PerseveranceColSize
		if col >= PerseveranceTableauCnt {
			break
		}
		bd.tableau[col] = append(bd.tableau[col], &PerseveranceTableauCard{Card: card, FaceUp: true})
	}

	// Move kings to the bottom of each column. The bottom is index 0 (the
	// buried card); the top is len-1 (the playable card).
	for i := range PerseveranceTableauCnt {
		col := bd.tableau[i]
		// Sort kings to the bottom while preserving relative order of other cards.
		kings := make([]*PerseveranceTableauCard, 0)
		others := make([]*PerseveranceTableauCard, 0, len(col))
		for _, tc := range col {
			if tc.Card.GetValue() == CardValueMax {
				kings = append(kings, tc)
			} else {
				others = append(others, tc)
			}
		}
		bd.tableau[i] = append(kings, others...)
	}
}

// Redeal 手詰まりを解く救済手段。**最大 2 回**。
//
// 集め方が決まっている: 末尾の列から順に手前の列へ重ね (12 を 11 の上、それを 10 の
// 上…)、**シャッフルせずに**先頭から 4 枚ずつ配り直す。したがって配り直した並びは
// 直前の盤から一意に決まり、運で変わる要素は無い。
func (bd *Perseverance) Redeal() error {
	if bd.phase != PerseverancePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if bd.redealsLeft <= 0 {
		return errors.New("no redeals left")
	}
	bd.takeSnapshot()

	// 逆順に重ねる。列 n を列 n-1 の「上」に置くので、下から読むと列 0, 1, ... の順。
	gathered := make([]*Card, 0, 48)
	for i := range PerseveranceTableauCnt {
		for _, tc := range bd.tableau[i] {
			gathered = append(gathered, tc.Card)
		}
	}

	bd.dealColumns(gathered)
	bd.redealsLeft--
	bd.moveCount++
	bd.appendLog("redeal", fmt.Sprintf("集めて配り直しました (残り%d回)", bd.redealsLeft), nil)
	bd.checkStalemate()
	return nil
}

// GetRedealsLeft 残りの再配り回数。
func (bd *Perseverance) GetRedealsLeft() int { return bd.redealsLeft }

// MoveTableauToTableau タブローからタブローにカードを移動（1枚のみ）
func (bd *Perseverance) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if bd.phase != PerseverancePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= PerseveranceTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= PerseveranceTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := bd.tableau[fromCol]
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	// **同スート降順の並びは一括で動かせる。**Baker's Dozen は先頭以外を拒むので、
	// ここが 4 つ目の分岐点。並びが崩れていれば 1 枚も動かさない。
	if !bd.isRun(fromCol, cardIndex) {
		return errors.New("cards below the top do not form a same-suit descending run")
	}
	moving := fromCards[cardIndex:]
	if !bd.canPlaceOnTableau(moving[0].Card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	bd.takeSnapshot()
	bd.tableau[toCol] = append(bd.tableau[toCol], moving...)
	bd.tableau[fromCol] = fromCards[:cardIndex]
	bd.moveCount++
	movedCards := make([]*Card, len(moving))
	for i, m := range moving {
		movedCards[i] = m.Card
	}
	bd.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), movedCards)
	bd.checkStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (bd *Perseverance) MoveTableauToFoundation(col int) error {
	if bd.phase != PerseverancePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= PerseveranceTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := bd.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	card := tc.Card
	fIdx := bd.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	bd.takeSnapshot()
	bd.tableau[col] = fromCards[:len(fromCards)-1]
	bd.foundation[fIdx] = append(bd.foundation[fIdx], card)
	bd.moveCount++
	bd.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	bd.checkGameClear()
	bd.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (bd *Perseverance) GiveUp() {
	if bd.phase == PerseverancePhasePlaying {
		bd.phase = PerseverancePhaseGameOver
		bd.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (bd *Perseverance) GetHint() *PerseveranceHint {
	if bd.phase != PerseverancePhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := range PerseveranceTableauCnt {
		if len(bd.tableau[col]) == 0 {
			continue
		}
		tc := bd.tableau[col][len(bd.tableau[col])-1]
		fIdx := bd.findFoundation(tc.Card)
		if fIdx >= 0 {
			return &PerseveranceHint{
				FromCol:   col,
				CardIndex: len(bd.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: タブローからタブローへ。
	//
	// **上札だけを見てはいけない。**Perseverance は同スート降順の並びを一括で
	// 動かせるので、上札が行き詰まっていても、その下から始まる並びは動けることが
	// ある。上札しか見ないと checkGameOver がまだ手のある盤を「手詰まり」と
	// 宣言する ── クローン元の BakersDozen は 1 枚ずつしか動かせないので、
	// あちらの走査をそのまま流用すると必ずこの穴が開く。
	for fromCol := range PerseveranceTableauCnt {
		for _, cardIndex := range bd.runStarts(fromCol) {
			card := bd.tableau[fromCol][cardIndex].Card
			for toCol := range PerseveranceTableauCnt {
				if toCol == fromCol {
					continue
				}
				if bd.canPlaceOnTableau(card, toCol) {
					return &PerseveranceHint{
						FromCol:   fromCol,
						CardIndex: cardIndex,
						ToZone:    "tableau",
						ToCol:     toCol,
					}
				}
			}
		}
	}
	return nil
}

// runStarts は列 col で「そこから上が同スート降順に並んでいる」開始位置を、
// 上札から順に返す。
//
// 動かせる単位はこの位置から上の塊だけなので、合法手の探索も UI の選択可否も
// この一覧で決まる。並びは必ず上札から連続するため、上から下へ 1 つずつ伸ばして
// 崩れた時点で止めればよい。
func (bd *Perseverance) runStarts(col int) []int {
	if col < 0 || col >= PerseveranceTableauCnt {
		return nil
	}
	cards := bd.tableau[col]
	starts := make([]int, 0, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		if !bd.isRun(col, i) {
			break
		}
		starts = append(starts, i)
	}
	return starts
}

// RunStarts は列 col の移動開始位置一覧 (上札から順)。UI が「掘り下げた札を
// 掴めるか」を判断するために使う。
func (bd *Perseverance) RunStarts(col int) []int { return bd.runStarts(col) }

// AutoComplete オートコンプリート（全ての山から可能な限りファンデーションへ）
func (bd *Perseverance) AutoComplete() error {
	if bd.phase != PerseverancePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	bd.takeSnapshot()
	for {
		moved := false
		for col := range PerseveranceTableauCnt {
			if len(bd.tableau[col]) == 0 {
				continue
			}
			tc := bd.tableau[col][len(bd.tableau[col])-1]
			card := tc.Card
			fIdx := bd.findFoundation(card)
			if fIdx < 0 {
				continue
			}
			bd.tableau[col] = bd.tableau[col][:len(bd.tableau[col])-1]
			bd.foundation[fIdx] = append(bd.foundation[fIdx], card)
			bd.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	bd.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	bd.checkGameClear()
	return nil
}

// AllFaceUp 全カードが表向きかどうか（Perseveranceでは常にtrue）
func (bd *Perseverance) AllFaceUp() bool {
	return true
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (bd *Perseverance) GetPhase() PerseverancePhase { return bd.phase }

// SetPhase フェーズ設定 (テスト用)
func (bd *Perseverance) SetPhase(phase PerseverancePhase) { bd.phase = phase }

// GetMoveCount 移動回数取得
func (bd *Perseverance) GetMoveCount() int { return bd.moveCount }

// GetTableau タブロー取得
func (bd *Perseverance) GetTableau() [PerseveranceTableauCnt][]*PerseveranceTableauCard {
	return bd.tableau
}

// GetFoundation ファンデーション取得
func (bd *Perseverance) GetFoundation() [PerseveranceFoundationCnt][]*Card { return bd.foundation }

// GetGameEndFlag returns true once the game has left the playing phase.
func (bd *Perseverance) GetGameEndFlag() bool { return bd.phase != PerseverancePhasePlaying }

// IsStalemate 手詰まり状態取得
func (bd *Perseverance) IsStalemate() bool { return bd.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (bd *Perseverance) SetIsStalemate(v bool) { bd.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (bd *Perseverance) SetTableau(tableau [PerseveranceTableauCnt][]*PerseveranceTableauCard) {
	bd.tableau = tableau
}

// SetFoundation ファンデーション設定 (テスト用)
func (bd *Perseverance) SetFoundation(foundation [PerseveranceFoundationCnt][]*Card) {
	bd.foundation = foundation
}

// Undo 直前の操作を取り消す
func (bd *Perseverance) Undo() error {
	if bd.phase != PerseverancePhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(bd.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := bd.history[len(bd.history)-1]
	bd.history = bd.history[:len(bd.history)-1]
	bd.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (bd *Perseverance) CanUndo() bool {
	return len(bd.history) > 0 && bd.phase == PerseverancePhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。
// 膠着状態でなければ0、脱出不可なら-1。
func (bd *Perseverance) UndoToEscape() int {
	return undoToEscape(bd.isStalemate, bd.history, func(s *perseveranceSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (bd *Perseverance) UndoN(n int) error {
	return undoN(bd, n)
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
// Perseverance: descending IN SUIT; empty columns cannot be filled.
// (クローン元の Baker's Dozen はランクだけを見る。)
func (bd *Perseverance) canPlaceOnTableau(card *Card, col int) bool {
	colCards := bd.tableau[col]
	if len(colCards) == 0 {
		// 空列は埋めない。Baker's Dozen と同じで、ここは変えない。
		return false
	}
	topCard := colCards[len(colCards)-1].Card
	// **スートも見る。**Baker's Dozen は `GetValue()-1` だけで判定するので、
	// ♠8 を ♥9 に載せられてしまう。Perseverance は同スート降順のみ。
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()-1
}

// isRun は列 col の cardIndex 以降が同スート降順に並んでいるかを返す。
//
// 一括移動できるのはこの並びだけで、途中でスートかランクが切れたら 1 枚も動かない。
func (bd *Perseverance) isRun(col, cardIndex int) bool {
	cards := bd.tableau[col]
	for i := cardIndex; i+1 < len(cards); i++ {
		upper, lower := cards[i].Card, cards[i+1].Card
		if upper.GetDesign() != lower.GetDesign() || lower.GetValue() != upper.GetValue()-1 {
			return false
		}
	}
	return true
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (bd *Perseverance) canPlaceOnFoundation(card *Card, fIdx int) bool {
	return canPlaceOnFoundationPile(bd.foundation[fIdx], card)
}

// LegalTargets は列 fromCol の一番下の札を置ける先を返す。
//
// **13 列 + 4 組札は押して試すには広すぎる。**Web は選択・hover の瞬間に
// 置ける先をリングで示している (#4795, #4454) のに、CUI は `m <from> <to>` を
// 打ってサーバに弾かれるまで分からなかった (#5581)。判定は canPlaceOnTableau /
// canPlaceOnFoundation をそのまま使う ── 規則を二重に持たない。
func (bd *Perseverance) LegalTargets(fromCol int) (tableau []int, foundation []int) {
	if fromCol < 0 || fromCol >= PerseveranceTableauCnt || len(bd.tableau[fromCol]) == 0 {
		return nil, nil
	}
	card := bd.tableau[fromCol][len(bd.tableau[fromCol])-1].Card
	// 自分の列を明示的に飛ばす分岐は置かない。一番下の札は自分自身なので
	// 「1 つ下のランク」にはならず、canPlaceOnTableau が必ず false を返す
	// (テストで固定してある)。置いてもどのテストでも区別できない分岐になる。
	for col := range PerseveranceTableauCnt {
		if bd.canPlaceOnTableau(card, col) {
			tableau = append(tableau, col)
		}
	}
	for f := range PerseveranceFoundationCnt {
		if bd.canPlaceOnFoundation(card, f) {
			foundation = append(foundation, f)
		}
	}
	return tableau, foundation
}

// findFoundation カードを置けるファンデーションのインデックスを探す（見つからない場合-1）
func (bd *Perseverance) findFoundation(card *Card) int {
	for i := range PerseveranceFoundationCnt {
		if bd.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定
func (bd *Perseverance) checkGameClear() {
	for i := range PerseveranceFoundationCnt {
		if len(bd.foundation[i]) != CardValueMax {
			return
		}
	}
	bd.phase = PerseverancePhaseGameClear
}

// checkStalemate 手詰まり判定
func (bd *Perseverance) checkStalemate() {
	if bd.phase != PerseverancePhasePlaying {
		return
	}
	if bd.GetHint() != nil {
		bd.isStalemate = false
		return
	}
	bd.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (bd *Perseverance) takeSnapshot() {
	snap := &perseveranceSnapshot{
		phase:       bd.phase,
		moveCount:   bd.moveCount,
		isStalemate: bd.isStalemate,
		redealsLeft: bd.redealsLeft,
	}
	for i := range PerseveranceTableauCnt {
		snap.tableau[i] = make([]*PerseveranceTableauCard, len(bd.tableau[i]))
		for j, tc := range bd.tableau[i] {
			snap.tableau[i][j] = &PerseveranceTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	for i := range PerseveranceFoundationCnt {
		snap.foundation[i] = make([]*Card, len(bd.foundation[i]))
		copy(snap.foundation[i], bd.foundation[i])
	}
	bd.history = append(bd.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (bd *Perseverance) restoreSnapshot(snap *perseveranceSnapshot) {
	bd.tableau = snap.tableau
	bd.foundation = snap.foundation
	bd.phase = snap.phase
	bd.moveCount = snap.moveCount
	bd.isStalemate = snap.isStalemate
	// **再配りは巻き戻す。**戻さないと Undo するたびに残り回数だけが減り続け、
	// 盤は初回配りのままなのに救済手段だけ消える。
	bd.redealsLeft = snap.redealsLeft
}

// appendLog 棋譜エントリを追加
func (bd *Perseverance) appendLog(actionType, detail string, cards []*Card) {
	bd.appendLogAt(bd.moveCount, 0, actionType, detail, cards)
}

// perseveranceJSON is the JSON wire format for Perseverance.
type perseveranceJSON struct {
	TrumpCards  *TrumpCards                                        `json:"tc"`
	Tableau     [PerseveranceTableauCnt][]*PerseveranceTableauCard `json:"tb"`
	Foundation  [PerseveranceFoundationCnt][]*Card                 `json:"fd"`
	Phase       PerseverancePhase                                  `json:"ps"`
	MoveCount   int                                                `json:"mc"`
	ActionLog   []*ActionLogEntry                                  `json:"al"`
	IsStalemate bool                                               `json:"sl"`
	// **Worker はリクエストごとに KV から作り直す。**残り再配り回数を載せないと
	// 毎リクエストで 2 に戻り、再配りが無制限になる。
	RedealsLeft int                     `json:"rd"`
	History     []*perseveranceSnapshot `json:"hi,omitempty"`
}

// perseveranceSnapshotJSON is the wire format for a single undo snapshot.
// perseveranceSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// perseveranceJSON's short keys to keep the KV payload compact (#1654).
type perseveranceSnapshotJSON struct {
	Tableau     [PerseveranceTableauCnt][]*PerseveranceTableauCard `json:"tb"`
	Foundation  [PerseveranceFoundationCnt][]*Card                 `json:"fd"`
	Phase       PerseverancePhase                                  `json:"ps"`
	MoveCount   int                                                `json:"mc"`
	IsStalemate bool                                               `json:"sl"`
	RedealsLeft int                                                `json:"rd"`
}

// MarshalJSON implements json.Marshaler for perseveranceSnapshot.
func (s *perseveranceSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(perseveranceSnapshotJSON{
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
		RedealsLeft: s.redealsLeft,
	})
}

// UnmarshalJSON implements json.Unmarshaler for perseveranceSnapshot.
func (s *perseveranceSnapshot) UnmarshalJSON(data []byte) error {
	var j perseveranceSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > perseveranceMaxSliceLen {
			return fmt.Errorf("perseverance: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > perseveranceMaxSliceLen {
			return fmt.Errorf("perseverance: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	for i := range PerseveranceTableauCnt {
		if s.tableau[i] == nil {
			s.tableau[i] = make([]*PerseveranceTableauCard, 0)
		}
	}
	s.foundation = j.Foundation
	for i := range PerseveranceFoundationCnt {
		if s.foundation[i] == nil {
			s.foundation[i] = make([]*Card, 0)
		}
	}
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	s.redealsLeft = j.RedealsLeft
	return nil
}

// MarshalJSON implements json.Marshaler.
func (bd *Perseverance) MarshalJSON() ([]byte, error) {
	return json.Marshal(perseveranceJSON{
		TrumpCards:  bd.trumpCards,
		Tableau:     bd.tableau,
		Foundation:  bd.foundation,
		Phase:       bd.phase,
		MoveCount:   bd.moveCount,
		ActionLog:   bd.actionLog,
		IsStalemate: bd.isStalemate,
		RedealsLeft: bd.redealsLeft,
		History:     bd.history,
	})
}

// perseveranceMaxSliceLen caps slice sizes during deserialisation.
const perseveranceMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (bd *Perseverance) UnmarshalJSON(data []byte) error {
	var j perseveranceJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > perseveranceMaxSliceLen ||
		len(j.History) > perseveranceMaxSliceLen {
		return fmt.Errorf("perseverance: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > perseveranceMaxSliceLen {
			return fmt.Errorf("perseverance: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > perseveranceMaxSliceLen {
			return fmt.Errorf("perseverance: foundation pile exceeds maximum allowed size")
		}
	}

	bd.trumpCards = j.TrumpCards
	if bd.trumpCards == nil {
		bd.trumpCards = NewTrumpCardsWithDecks(1, 0)
	}
	bd.tableau = j.Tableau
	for i := range PerseveranceTableauCnt {
		if bd.tableau[i] == nil {
			bd.tableau[i] = make([]*PerseveranceTableauCard, 0)
		}
	}
	bd.foundation = j.Foundation
	for i := range PerseveranceFoundationCnt {
		if bd.foundation[i] == nil {
			bd.foundation[i] = make([]*Card, 0)
		}
	}
	bd.phase = j.Phase
	bd.moveCount = j.MoveCount
	bd.actionLog = j.ActionLog
	if bd.actionLog == nil {
		bd.actionLog = make([]*ActionLogEntry, 0)
	}
	bd.history = j.History
	if bd.history == nil {
		bd.history = make([]*perseveranceSnapshot, 0)
	}
	bd.isStalemate = j.IsStalemate
	// **範囲外を黙って受けない。**KV の値が壊れていたら、上限を超える再配りを
	// 与えるより拒む。
	if j.RedealsLeft < 0 || j.RedealsLeft > PerseveranceMaxRedeals {
		return fmt.Errorf("perseverance: redeals left out of range: %d", j.RedealsLeft)
	}
	bd.redealsLeft = j.RedealsLeft
	return nil
}
