//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// BisleyPhase ビズリーのゲームフェーズ
type BisleyPhase int

// Bisleyのフェーズ定数
const (
	// BisleyPhasePlaying プレイ中
	BisleyPhasePlaying BisleyPhase = iota
	// BisleyPhaseGameClear ゲームクリア
	BisleyPhaseGameClear
	// BisleyPhaseGameOver ゲームオーバー
	BisleyPhaseGameOver
)

// BisleyTableauCnt タブローの列数。48 枚を 13 列に配ると 4 列が 3 枚、9 列が 4 枚になる。
const BisleyTableauCnt = 13

// BisleyFoundationCnt スートごとの基礎札の数（昇順・降順それぞれ 4 列）
const BisleyFoundationCnt = 4

// bisleySuitOrder 基礎札のインデックスとスートの対応。同じスートが常に同じ列を使うよう
// 固定しておくと、UI と棋譜が配り直しても安定する。
var bisleySuitOrder = [BisleyFoundationCnt]int{
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
}

// BisleySuitIndex スートに対応する基礎札インデックスを返す（不明なスートは -1）。
func BisleySuitIndex(design int) int {
	for i, d := range bisleySuitOrder {
		if d == design {
			return i
		}
	}
	return -1
}

// BisleyTableauCard タブロー上のカード。ビズリーは全札が表向きだが、他のソリティアと
// 同じ形にしておくとプレゼンターを使い回せる。
type BisleyTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// BisleyHint ビズリーのヒント
type BisleyHint struct {
	// FromCol 移動元のタブロー列
	FromCol int
	// ToZone 移動先 "ace" / "king" / "tableau"
	ToZone string
	// ToIdx 移動先のインデックス（基礎札またはタブロー列）
	ToIdx int
}

// Bisley ビズリーゲームクラス。
//
// 52 枚すべてが表向きのオープンソリティア。4 枚のエースだけが最初から基礎札として
// 置かれ、残り 48 枚が 13 列のタブローに配られる。エースの基礎札は同スートで昇順
// (A→K)、キングの基礎札は同スートで降順 (K→A) に積む。キングの列は最初は空で、
// タブローからキングを上げたときに開く。
//
// タブロー上は「同スートで隣接ランク（上下どちらでも）」のみ重ねられ、動かせるのは
// 各列の最上段 1 枚だけ。issue #4390 の仕様案は「タブロー間は自由に移動できる」と
// していたが、それでは任意のカードをいつでも掘り出せてしまい、どの配牌も必ず解ける
// ため勝負が成立しない。ここは実際のビズリーの規則を採っている。
type Bisley struct {
	trumpCards      *TrumpCards
	aceFoundations  [BisleyFoundationCnt][]*Card
	kingFoundations [BisleyFoundationCnt][]*Card
	tableau         [BisleyTableauCnt][]*BisleyTableauCard
	phase           BisleyPhase
	moveCount       int
	actionLog       []*ActionLogEntry
	history         []*bisleySnapshot
	isStalemate     bool
}

// bisleySnapshot アンドゥ用スナップショット
type bisleySnapshot struct {
	aceFoundations  [BisleyFoundationCnt][]*Card
	kingFoundations [BisleyFoundationCnt][]*Card
	tableau         [BisleyTableauCnt][]*BisleyTableauCard
	phase           BisleyPhase
	moveCount       int
	isStalemate     bool
}

// NewBisley コンストラクタ
func NewBisley(trumpCards *TrumpCards) *Bisley {
	return &Bisley{trumpCards: trumpCards}
}

// NewDefaultBisley returns Bisley with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultBisley() *Bisley {
	return NewBisley(NewTrumpCards(0))
}

// Reset ゲームリセット
func (b *Bisley) Reset() {
	b.trumpCards.Shuffle()
	b.phase = BisleyPhasePlaying
	b.moveCount = 0
	b.actionLog = nil
	b.history = nil
	b.isStalemate = false

	for i := range BisleyFoundationCnt {
		b.aceFoundations[i] = nil
		b.kingFoundations[i] = nil
	}
	for i := range BisleyTableauCnt {
		b.tableau[i] = nil
	}

	// エースだけを抜いて基礎札に置き、残り 48 枚を列へ配る。キングは配らない：
	// 降順の基礎札はプレイヤーがキングを上げたときに開く。
	remaining := make([]*Card, 0, CardCnt-BisleyFoundationCnt)
	for {
		card := b.trumpCards.DrawCard()
		if card == nil {
			break
		}
		if card.GetValue() == 1 {
			if fIdx := BisleySuitIndex(card.GetDesign()); fIdx >= 0 {
				b.aceFoundations[fIdx] = []*Card{card}
				continue
			}
		}
		remaining = append(remaining, card)
	}
	for i, card := range remaining {
		col := i % BisleyTableauCnt
		b.tableau[col] = append(b.tableau[col], &BisleyTableauCard{Card: card, FaceUp: true})
	}

	b.checkStalemate()
}

// MoveTableauToAceFoundation タブロー最上段を昇順（エース側）の基礎札へ移す
func (b *Bisley) MoveTableauToAceFoundation(col int) error {
	card, err := b.topOf(col)
	if err != nil {
		return err
	}
	fIdx := BisleySuitIndex(card.GetDesign())
	if fIdx < 0 || !b.canPlaceOnAce(card, fIdx) {
		return errors.New("cannot place card on ascending foundation")
	}
	b.takeSnapshot()
	b.popTop(col)
	b.aceFoundations[fIdx] = append(b.aceFoundations[fIdx], card)
	b.afterMove("move", fmt.Sprintf("列%d→昇順基礎%d", col+1, fIdx+1), card)
	return nil
}

// MoveTableauToKingFoundation タブロー最上段を降順（キング側）の基礎札へ移す
func (b *Bisley) MoveTableauToKingFoundation(col int) error {
	card, err := b.topOf(col)
	if err != nil {
		return err
	}
	fIdx := BisleySuitIndex(card.GetDesign())
	if fIdx < 0 || !b.canPlaceOnKing(card, fIdx) {
		return errors.New("cannot place card on descending foundation")
	}
	b.takeSnapshot()
	b.popTop(col)
	b.kingFoundations[fIdx] = append(b.kingFoundations[fIdx], card)
	b.afterMove("move", fmt.Sprintf("列%d→降順基礎%d", col+1, fIdx+1), card)
	return nil
}

// MoveTableauToTableau タブロー最上段を別の列の最上段へ重ねる（同スートで隣接ランクのみ）
func (b *Bisley) MoveTableauToTableau(fromCol, toCol int) error {
	card, err := b.topOf(fromCol)
	if err != nil {
		return err
	}
	if toCol < 0 || toCol >= BisleyTableauCnt {
		return errors.New("invalid destination column")
	}
	if fromCol == toCol {
		return errors.New("source and destination are the same column")
	}
	dst := b.tableau[toCol]
	if len(dst) == 0 {
		// 空き列は自由置き場ではない。ここを許すと事実上どの札でも掘り出せる。
		return errors.New("cannot move onto an empty column")
	}
	top := dst[len(dst)-1].Card
	if top.GetDesign() != card.GetDesign() || abs(top.GetValue()-card.GetValue()) != 1 {
		return errors.New("tableau builds up or down by one in the same suit")
	}
	b.takeSnapshot()
	b.popTop(fromCol)
	b.tableau[toCol] = append(b.tableau[toCol], &BisleyTableauCard{Card: card, FaceUp: true})
	b.afterMove("move", fmt.Sprintf("列%d→列%d", fromCol+1, toCol+1), card)
	return nil
}

// GiveUp ギブアップ
func (b *Bisley) GiveUp() {
	if b.phase == BisleyPhasePlaying {
		b.phase = BisleyPhaseGameOver
		b.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 基礎札へ送れる手を優先して 1 つ提示する
func (b *Bisley) GetHint() *BisleyHint {
	if b.phase != BisleyPhasePlaying {
		return nil
	}
	for col := range BisleyTableauCnt {
		pile := b.tableau[col]
		if len(pile) == 0 {
			continue
		}
		card := pile[len(pile)-1].Card
		fIdx := BisleySuitIndex(card.GetDesign())
		if fIdx < 0 {
			continue
		}
		if b.canPlaceOnAce(card, fIdx) {
			return &BisleyHint{FromCol: col, ToZone: "ace", ToIdx: fIdx}
		}
		if b.canPlaceOnKing(card, fIdx) {
			return &BisleyHint{FromCol: col, ToZone: "king", ToIdx: fIdx}
		}
	}
	return nil
}

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (b *Bisley) AutoComplete() error {
	if b.phase != BisleyPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	moved := false
	for {
		h := b.GetHint()
		if h == nil {
			break
		}
		var err error
		if h.ToZone == "ace" {
			err = b.MoveTableauToAceFoundation(h.FromCol)
		} else {
			err = b.MoveTableauToKingFoundation(h.FromCol)
		}
		if err != nil {
			return err
		}
		moved = true
	}
	if !moved {
		return errors.New("no card can be auto-completed")
	}
	return nil
}

// Undo 直前の 1 手を取り消す
func (b *Bisley) Undo() error {
	if len(b.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := b.history[len(b.history)-1]
	b.history = b.history[:len(b.history)-1]
	b.aceFoundations = snap.aceFoundations
	b.kingFoundations = snap.kingFoundations
	b.tableau = snap.tableau
	b.phase = snap.phase
	b.moveCount = snap.moveCount
	b.isStalemate = snap.isStalemate
	return nil
}

// CanUndo アンドゥ可能か
func (b *Bisley) CanUndo() bool { return len(b.history) > 0 }

// UndoN n 手戻す
func (b *Bisley) UndoN(n int) error {
	if n <= 0 {
		return errors.New("n must be positive")
	}
	if n > len(b.history) {
		return errors.New("not enough history")
	}
	for range n {
		if err := b.Undo(); err != nil {
			return err
		}
	}
	return nil
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (b *Bisley) UndoToEscape() int {
	if !b.isStalemate {
		return 0
	}
	for i := len(b.history) - 1; i >= 0; i-- {
		if !b.history[i].isStalemate {
			return len(b.history) - i
		}
	}
	return -1
}

// AllFaceUp ビズリーは常に全札が表向き
func (b *Bisley) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (b *Bisley) GetPhase() BisleyPhase { return b.phase }

// GetMoveCount 手数取得
func (b *Bisley) GetMoveCount() int { return b.moveCount }

// GetAceFoundations 昇順基礎札を取得
func (b *Bisley) GetAceFoundations() [BisleyFoundationCnt][]*Card { return b.aceFoundations }

// GetKingFoundations 降順基礎札を取得
func (b *Bisley) GetKingFoundations() [BisleyFoundationCnt][]*Card { return b.kingFoundations }

// GetTableau タブローを取得
func (b *Bisley) GetTableau() [BisleyTableauCnt][]*BisleyTableauCard { return b.tableau }

// GetActionLog 棋譜取得
func (b *Bisley) GetActionLog() []*ActionLogEntry { return b.actionLog }

// GetGameEndFlag ゲーム終了フラグ
func (b *Bisley) GetGameEndFlag() bool { return b.phase != BisleyPhasePlaying }

// IsStalemate 手詰まりか
func (b *Bisley) IsStalemate() bool { return b.isStalemate }

// --- Private helpers ---

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// topOf 指定列の最上段カードを返す（列が空・範囲外ならエラー）
func (b *Bisley) topOf(col int) (*Card, error) {
	if b.phase != BisleyPhasePlaying {
		return nil, errors.New("game is not in playing phase")
	}
	if col < 0 || col >= BisleyTableauCnt {
		return nil, errors.New("invalid column index")
	}
	pile := b.tableau[col]
	if len(pile) == 0 {
		return nil, errors.New("column is empty")
	}
	return pile[len(pile)-1].Card, nil
}

// popTop 指定列の最上段を取り除く
func (b *Bisley) popTop(col int) {
	b.tableau[col] = b.tableau[col][:len(b.tableau[col])-1]
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (b *Bisley) afterMove(actionType, detail string, card *Card) {
	b.moveCount++
	b.appendLog(actionType, detail, []*Card{card})
	b.checkGameClear()
	b.checkStalemate()
}

// canPlaceOnAce 昇順基礎札に置けるか（同スートで 1 つ上、K で打ち止め）
func (b *Bisley) canPlaceOnAce(card *Card, fIdx int) bool {
	pile := b.aceFoundations[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	top := pile[len(pile)-1]
	return top.GetDesign() == card.GetDesign() && card.GetValue() == top.GetValue()+1
}

// canPlaceOnKing 降順基礎札に置けるか（空列はキングのみ、以降は同スートで 1 つ下）
func (b *Bisley) canPlaceOnKing(card *Card, fIdx int) bool {
	pile := b.kingFoundations[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == CardValueMax
	}
	top := pile[len(pile)-1]
	return top.GetDesign() == card.GetDesign() && card.GetValue() == top.GetValue()-1
}

// checkGameClear 52 枚すべてが基礎札に乗ったか
func (b *Bisley) checkGameClear() {
	total := 0
	for i := range BisleyFoundationCnt {
		total += len(b.aceFoundations[i]) + len(b.kingFoundations[i])
	}
	if total == CardCnt {
		b.phase = BisleyPhaseGameClear
	}
}

// checkStalemate 打つ手が一つも無い状態か
func (b *Bisley) checkStalemate() {
	if b.phase != BisleyPhasePlaying {
		return
	}
	if b.GetHint() != nil {
		b.isStalemate = false
		return
	}
	// 基礎札へ送れなくても、タブロー上で動かせる札が残っていれば手詰まりではない。
	for from := range BisleyTableauCnt {
		pile := b.tableau[from]
		if len(pile) == 0 {
			continue
		}
		card := pile[len(pile)-1].Card
		for to := range BisleyTableauCnt {
			if to == from || len(b.tableau[to]) == 0 {
				continue
			}
			top := b.tableau[to][len(b.tableau[to])-1].Card
			if top.GetDesign() == card.GetDesign() && abs(top.GetValue()-card.GetValue()) == 1 {
				b.isStalemate = false
				return
			}
		}
	}
	b.isStalemate = true
}

// takeSnapshot 現在の状態を保存する
func (b *Bisley) takeSnapshot() {
	snap := &bisleySnapshot{phase: b.phase, moveCount: b.moveCount, isStalemate: b.isStalemate}
	for i := range BisleyFoundationCnt {
		snap.aceFoundations[i] = append([]*Card(nil), b.aceFoundations[i]...)
		snap.kingFoundations[i] = append([]*Card(nil), b.kingFoundations[i]...)
	}
	for i := range BisleyTableauCnt {
		snap.tableau[i] = append([]*BisleyTableauCard(nil), b.tableau[i]...)
	}
	b.history = append(b.history, snap)
}

// appendLog 棋譜エントリを追加
func (b *Bisley) appendLog(actionType, detail string, cards []*Card) {
	b.actionLog = append(b.actionLog, &ActionLogEntry{
		TurnNumber: b.moveCount,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// bisleyJSON is the JSON wire format for Bisley.
type bisleyJSON struct {
	TrumpCards      *TrumpCards                            `json:"tc"`
	AceFoundations  [BisleyFoundationCnt][]*Card           `json:"af"`
	KingFoundations [BisleyFoundationCnt][]*Card           `json:"kf"`
	Tableau         [BisleyTableauCnt][]*BisleyTableauCard `json:"tb"`
	Phase           BisleyPhase                            `json:"ps"`
	MoveCount       int                                    `json:"mc"`
	ActionLog       []*ActionLogEntry                      `json:"al"`
	IsStalemate     bool                                   `json:"sl"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (b *Bisley) MarshalJSON() ([]byte, error) {
	return json.Marshal(&bisleyJSON{
		TrumpCards:      b.trumpCards,
		AceFoundations:  b.aceFoundations,
		KingFoundations: b.kingFoundations,
		Tableau:         b.tableau,
		Phase:           b.phase,
		MoveCount:       b.moveCount,
		ActionLog:       b.actionLog,
		IsStalemate:     b.isStalemate,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (b *Bisley) UnmarshalJSON(data []byte) error {
	var j bisleyJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < BisleyPhasePlaying || j.Phase > BisleyPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	for i := range BisleyFoundationCnt {
		if len(j.AceFoundations[i])+len(j.KingFoundations[i]) > CardValueMax {
			return fmt.Errorf("suit %d holds %d cards", i, len(j.AceFoundations[i])+len(j.KingFoundations[i]))
		}
	}
	if j.TrumpCards != nil {
		b.trumpCards = j.TrumpCards
	}
	b.aceFoundations = j.AceFoundations
	b.kingFoundations = j.KingFoundations
	b.tableau = j.Tableau
	b.phase = j.Phase
	b.moveCount = j.MoveCount
	b.actionLog = j.ActionLog
	b.isStalemate = j.IsStalemate
	return nil
}
