//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// CrescentPhase クレセント・ソリティアのゲームフェーズ。
type CrescentPhase int

// Crescent のフェーズ定数。
const (
	// CrescentPhasePlaying プレイ中。
	CrescentPhasePlaying CrescentPhase = iota
	// CrescentPhaseGameClear ゲームクリア。
	CrescentPhaseGameClear
	// CrescentPhaseGameOver ゲームオーバー (ギブアップまたは手詰まり)。
	CrescentPhaseGameOver
)

// CrescentTableauCnt タブローの列数 (16)。
const CrescentTableauCnt = 16

// CrescentTableauInitialSize 初期配置時の各タブロー列の枚数 (6)。
const CrescentTableauInitialSize = 6

// CrescentFoundationCnt ファンデーション数 (4 = 昇順 A→K、4 = 降順 K→A)。
const CrescentFoundationCnt = 8

// CrescentAscendingFoundationCnt 昇順ファンデーションの数 (前半 0..3)。
const CrescentAscendingFoundationCnt = 4

// CrescentMaxRedeals 1 ゲーム中に許される再配り (シャッフル) 回数。
const CrescentMaxRedeals = 3

// CrescentTableauCard タブロー上の 1 枚 (常に表向き)。
type CrescentTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// CrescentHint ヒント。
//
//	FromCol  ヒントの元となるタブロー列。
//	ToZone   "tableau" もしくは "foundation"。
//	ToCol    タブローなら列番号、ファンデーションならファンデーション ID。
//	Redeal   true の場合は再配りを推奨するヒント (FromCol/ToCol は -1)。
type CrescentHint struct {
	FromCol int
	ToZone  string
	ToCol   int
	Redeal  bool
}

// CrescentConfig クレセント・ソリティアのゲーム設定 (現状フィールド無し)。
type CrescentConfig struct{}

// Crescent クレセント・ソリティアの本体。
type Crescent struct {
	trumpCards       *TrumpCards
	tableau          [CrescentTableauCnt][]*CrescentTableauCard
	foundation       [CrescentFoundationCnt][]*Card
	phase            CrescentPhase
	moveCount        int
	redealsRemaining int
	actionLogBase
	history     []*crescentSnapshot
	isStalemate bool
}

// crescentSnapshot アンドゥ用スナップショット。
type crescentSnapshot struct {
	tableau          [CrescentTableauCnt][]*CrescentTableauCard
	foundation       [CrescentFoundationCnt][]*Card
	phase            CrescentPhase
	moveCount        int
	redealsRemaining int
	isStalemate      bool
}

// NewCrescent コンストラクタ。
func NewCrescent(trumpCards *TrumpCards) *Crescent {
	return &Crescent{trumpCards: trumpCards}
}

// NewDefaultCrescent は 2 デッキ (104 枚) を持つ Crescent を返す。
// CUI / Web / Worker の生成サイトはすべてこの関数を経由する。
func NewDefaultCrescent() *Crescent {
	return NewCrescent(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲーム初期化。各スートの A 1 枚と K 1 枚を取り出してファンデーションに種を置き、
// 残り 96 枚を 16 列 × 6 枚のタブローへ配る。再配り回数は 3 にリセットされる。
func (cr *Crescent) Reset() {
	cr.trumpCards.Shuffle()
	cr.phase = CrescentPhasePlaying
	cr.moveCount = 0
	cr.redealsRemaining = CrescentMaxRedeals
	cr.actionLog = nil
	cr.history = nil
	cr.isStalemate = false

	deck := make([]*Card, 0, CardCnt*2)
	for cr.trumpCards.GetRemainingCount() > 0 {
		deck = append(deck, cr.trumpCards.DrawCard())
	}

	for i := range CrescentFoundationCnt {
		cr.foundation[i] = nil
	}

	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	for idx, suit := range suits {
		aceIdx := findFirstCard(deck, suit, 1)
		if aceIdx < 0 {
			panic(fmt.Sprintf("crescent Reset: ace of suit %d not found in deck", suit))
		}
		deck, cr.foundation[idx] = takeCardAt(deck, aceIdx), []*Card{deck[aceIdx]}
		kingIdx := findFirstCard(deck, suit, CardValueMax)
		if kingIdx < 0 {
			panic(fmt.Sprintf("crescent Reset: king of suit %d not found in deck", suit))
		}
		deck, cr.foundation[idx+CrescentAscendingFoundationCnt] = takeCardAt(deck, kingIdx), []*Card{deck[kingIdx]}
	}

	for i := range CrescentTableauCnt {
		cr.tableau[i] = make([]*CrescentTableauCard, 0, CrescentTableauInitialSize)
		for j := range CrescentTableauInitialSize {
			card := deck[i*CrescentTableauInitialSize+j]
			cr.tableau[i] = append(cr.tableau[i], &CrescentTableauCard{Card: card, FaceUp: true})
		}
	}
}

// findFirstCard は deck から (design, value) に最初にマッチするカードのインデックスを返す。
// 見つからない場合は -1 を返す。
func findFirstCard(deck []*Card, design, value int) int {
	for i, c := range deck {
		if c == nil {
			continue
		}
		if c.GetDesign() == design && c.GetValue() == value {
			return i
		}
	}
	return -1
}

// takeCardAt は deck から idx 番目を取り除いた新しいスライスを返す。idx が範囲外なら元のスライスを返す。
func takeCardAt(deck []*Card, idx int) []*Card {
	if idx < 0 || idx >= len(deck) {
		return deck
	}
	out := make([]*Card, 0, len(deck)-1)
	out = append(out, deck[:idx]...)
	out = append(out, deck[idx+1:]...)
	return out
}

// MoveTableauToTableau は fromCol の最上段カードを toCol へ移す。
// 同スートで値 ±1、A↔K の wrap-around を許可。空タブローは移動先にできない。
func (cr *Crescent) MoveTableauToTableau(fromCol, toCol int) error {
	if cr.phase != CrescentPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= CrescentTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= CrescentTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := cr.tableau[fromCol]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	if !cr.canPlaceOnTableau(tc.Card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	cr.takeSnapshot()
	cr.tableau[toCol] = append(cr.tableau[toCol], tc)
	cr.tableau[fromCol] = fromCards[:len(fromCards)-1]
	cr.moveCount++
	cr.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), []*Card{tc.Card})
	cr.checkCrescentStalemate()
	return nil
}

// MoveTableauToFoundation は fromCol の最上段カードを foundationIdx へ移す。
// 昇順 (0..3) / 降順 (4..7) の方向はインデックスから決まる。
func (cr *Crescent) MoveTableauToFoundation(fromCol, foundationIdx int) error {
	if cr.phase != CrescentPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= CrescentTableauCnt {
		return errors.New("invalid from column")
	}
	if foundationIdx < 0 || foundationIdx >= CrescentFoundationCnt {
		return errors.New("invalid foundation index")
	}
	fromCards := cr.tableau[fromCol]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	if !cr.canPlaceOnFoundation(tc.Card, foundationIdx) {
		return errors.New("cannot place card on foundation")
	}
	cr.takeSnapshot()
	cr.tableau[fromCol] = fromCards[:len(fromCards)-1]
	cr.foundation[foundationIdx] = append(cr.foundation[foundationIdx], tc.Card)
	cr.moveCount++
	cr.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション%d", fromCol, foundationIdx), []*Card{tc.Card})
	cr.checkGameClear()
	cr.checkCrescentStalemate()
	return nil
}

// Redeal は再配りを実行する。各タブロー列を逆順に並べ替える。残り回数 0 か非プレイ中ならエラー。
func (cr *Crescent) Redeal() error {
	if cr.phase != CrescentPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if cr.redealsRemaining <= 0 {
		return errors.New("no redeals remaining")
	}
	cr.takeSnapshot()
	for i := range CrescentTableauCnt {
		col := cr.tableau[i]
		for l, r := 0, len(col)-1; l < r; l, r = l+1, r-1 {
			col[l], col[r] = col[r], col[l]
		}
		cr.tableau[i] = col
	}
	cr.redealsRemaining--
	cr.moveCount++
	cr.appendLog("redeal", fmt.Sprintf("再配り (残り%d回)", cr.redealsRemaining), nil)
	cr.checkCrescentStalemate()
	return nil
}

// GiveUp ギブアップ。
func (cr *Crescent) GiveUp() {
	if cr.phase == CrescentPhasePlaying {
		cr.phase = CrescentPhaseGameOver
		cr.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得する。優先度: タブロー→ファンデーション > タブロー→タブロー > 再配り。
func (cr *Crescent) GetHint() *CrescentHint {
	if cr.phase != CrescentPhasePlaying {
		return nil
	}
	for col := range CrescentTableauCnt {
		if len(cr.tableau[col]) == 0 {
			continue
		}
		tc := cr.tableau[col][len(cr.tableau[col])-1]
		for fIdx := range CrescentFoundationCnt {
			if cr.canPlaceOnFoundation(tc.Card, fIdx) {
				return &CrescentHint{FromCol: col, ToZone: "foundation", ToCol: fIdx}
			}
		}
	}
	for fromCol := range CrescentTableauCnt {
		fromCards := cr.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		card := fromCards[len(fromCards)-1].Card
		for toCol := range CrescentTableauCnt {
			if toCol == fromCol {
				continue
			}
			if cr.canPlaceOnTableau(card, toCol) {
				return &CrescentHint{FromCol: fromCol, ToZone: "tableau", ToCol: toCol}
			}
		}
	}
	if cr.redealsRemaining > 0 {
		return &CrescentHint{FromCol: -1, ToZone: "", ToCol: -1, Redeal: true}
	}
	return nil
}

// AutoComplete タブロー上のカードを可能な限りファンデーションへ移動させる。
func (cr *Crescent) AutoComplete() error {
	if cr.phase != CrescentPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	cr.takeSnapshot()
	for {
		moved := false
		for col := range CrescentTableauCnt {
			if len(cr.tableau[col]) == 0 {
				continue
			}
			tc := cr.tableau[col][len(cr.tableau[col])-1]
			for fIdx := range CrescentFoundationCnt {
				if cr.canPlaceOnFoundation(tc.Card, fIdx) {
					cr.tableau[col] = cr.tableau[col][:len(cr.tableau[col])-1]
					cr.foundation[fIdx] = append(cr.foundation[fIdx], tc.Card)
					cr.moveCount++
					moved = true
					break
				}
			}
		}
		if !moved {
			break
		}
	}
	cr.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	cr.checkGameClear()
	cr.checkCrescentStalemate()
	return nil
}

// AllFaceUp Crescent では常にすべて表向き。インターフェース整合のため定義する。
func (cr *Crescent) AllFaceUp() bool { return true }

// --- State getters/setters ---

// GetPhase 現在のフェーズを返す。
func (cr *Crescent) GetPhase() CrescentPhase { return cr.phase }

// SetPhase テスト用フェーズ設定。
func (cr *Crescent) SetPhase(phase CrescentPhase) { cr.phase = phase }

// GetMoveCount 移動回数を返す。
func (cr *Crescent) GetMoveCount() int { return cr.moveCount }

// GetRedealsRemaining 残り再配り回数を返す。
func (cr *Crescent) GetRedealsRemaining() int { return cr.redealsRemaining }

// SetRedealsRemaining テスト用再配り回数設定。
func (cr *Crescent) SetRedealsRemaining(n int) { cr.redealsRemaining = n }

// GetTableau タブローを返す。
func (cr *Crescent) GetTableau() [CrescentTableauCnt][]*CrescentTableauCard { return cr.tableau }

// GetFoundation ファンデーションを返す。
func (cr *Crescent) GetFoundation() [CrescentFoundationCnt][]*Card { return cr.foundation }

// GetGameEndFlag プレイ中でなくなったかを返す。
func (cr *Crescent) GetGameEndFlag() bool { return cr.phase != CrescentPhasePlaying }

// IsStalemate 手詰まり状態を返す。
func (cr *Crescent) IsStalemate() bool { return cr.isStalemate }

// SetIsStalemate テスト用手詰まり設定。
func (cr *Crescent) SetIsStalemate(v bool) { cr.isStalemate = v }

// SetTableau テスト用タブロー設定。
func (cr *Crescent) SetTableau(tableau [CrescentTableauCnt][]*CrescentTableauCard) {
	cr.tableau = tableau
}

// SetFoundation テスト用ファンデーション設定。
func (cr *Crescent) SetFoundation(foundation [CrescentFoundationCnt][]*Card) {
	cr.foundation = foundation
}

// Undo 直前の操作を取り消す。
func (cr *Crescent) Undo() error {
	if cr.phase != CrescentPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(cr.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := cr.history[len(cr.history)-1]
	cr.history = cr.history[:len(cr.history)-1]
	cr.restoreSnapshot(snap)
	// Truncate the matching log entry so the action log stays in sync with
	// state. Each undoable action (Move*/Redeal/AutoComplete) appends exactly
	// one log entry after its takeSnapshot+state change, so the last entry is
	// always the one being reverted.
	if len(cr.actionLog) > 0 {
		cr.actionLog = cr.actionLog[:len(cr.actionLog)-1]
	}
	return nil
}

// CanUndo アンドゥ可能かを返す。
func (cr *Crescent) CanUndo() bool {
	return len(cr.history) > 0 && cr.phase == CrescentPhasePlaying
}

// UndoToEscape 手詰まりから脱出するためのアンドゥ回数。手詰まりでなければ 0、脱出不可なら -1。
func (cr *Crescent) UndoToEscape() int {
	return undoToEscape(cr.isStalemate, cr.history, func(s *crescentSnapshot) bool { return s.isStalemate })
}

// UndoN n 回アンドゥする。
func (cr *Crescent) UndoN(n int) error {
	for i := range n {
		if err := cr.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
}

// --- Private helpers ---

// canPlaceOnTableau タブロー toCol の最上段に card を置けるか判定。
// 同スートで値差 ±1、A↔K の wrap を許可。空タブローは許可しない。
func (cr *Crescent) canPlaceOnTableau(card *Card, toCol int) bool {
	col := cr.tableau[toCol]
	if len(col) == 0 {
		return false
	}
	top := col[len(col)-1].Card
	if card.GetDesign() != top.GetDesign() {
		return false
	}
	cv, tv := card.GetValue(), top.GetValue()
	switch {
	case cv == tv+1:
		return true
	case cv == tv-1:
		return true
	case cv == 1 && tv == CardValueMax:
		return true
	case cv == CardValueMax && tv == 1:
		return true
	}
	return false
}

// canPlaceOnFoundation ファンデーション fIdx に card を置けるか判定。
// fIdx 0..3 は昇順 A→K、fIdx 4..7 は降順 K→A。スートはインデックスから一意に決まる。
func (cr *Crescent) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if fIdx < 0 || fIdx >= CrescentFoundationCnt {
		return false
	}
	suit := crescentFoundationSuit(fIdx)
	if card.GetDesign() != suit {
		return false
	}
	pile := cr.foundation[fIdx]
	if len(pile) == 0 {
		if fIdx < CrescentAscendingFoundationCnt {
			return card.GetValue() == 1
		}
		return card.GetValue() == CardValueMax
	}
	top := pile[len(pile)-1]
	if fIdx < CrescentAscendingFoundationCnt {
		return card.GetValue() == top.GetValue()+1
	}
	return card.GetValue() == top.GetValue()-1
}

// crescentFoundationSuit ファンデーション ID から対応スートを返す。
func crescentFoundationSuit(fIdx int) int {
	suits := [...]int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	return suits[fIdx%CrescentAscendingFoundationCnt]
}

// CrescentFoundationSuit ファンデーション ID から対応スートを返す (公開ヘルパ)。
func CrescentFoundationSuit(fIdx int) int { return crescentFoundationSuit(fIdx) }

// CrescentIsAscendingFoundation 昇順ファンデーションかを返す。
func CrescentIsAscendingFoundation(fIdx int) bool { return fIdx < CrescentAscendingFoundationCnt }

// checkGameClear ゲームクリア判定。
func (cr *Crescent) checkGameClear() {
	total := 0
	for i := range CrescentFoundationCnt {
		total += len(cr.foundation[i])
	}
	if total == CardCnt*2 {
		cr.phase = CrescentPhaseGameClear
	}
}

// checkCrescentStalemate 手詰まり判定。残り再配り 0 で合法手も無ければ stalemate。
func (cr *Crescent) checkCrescentStalemate() {
	if cr.phase != CrescentPhasePlaying {
		return
	}
	hasMove := cr.hasAnyLegalMove()
	if hasMove {
		cr.isStalemate = false
		return
	}
	if cr.redealsRemaining > 0 {
		cr.isStalemate = false
		return
	}
	cr.isStalemate = true
}

// hasAnyLegalMove タブロー間/ファンデーションへの合法手が存在するか判定。
func (cr *Crescent) hasAnyLegalMove() bool {
	for col := range CrescentTableauCnt {
		if len(cr.tableau[col]) == 0 {
			continue
		}
		tc := cr.tableau[col][len(cr.tableau[col])-1]
		for fIdx := range CrescentFoundationCnt {
			if cr.canPlaceOnFoundation(tc.Card, fIdx) {
				return true
			}
		}
	}
	for fromCol := range CrescentTableauCnt {
		if len(cr.tableau[fromCol]) == 0 {
			continue
		}
		card := cr.tableau[fromCol][len(cr.tableau[fromCol])-1].Card
		for toCol := range CrescentTableauCnt {
			if toCol == fromCol {
				continue
			}
			if cr.canPlaceOnTableau(card, toCol) {
				return true
			}
		}
	}
	return false
}

// takeSnapshot 現状を履歴に保存する。
func (cr *Crescent) takeSnapshot() {
	snap := &crescentSnapshot{
		phase:            cr.phase,
		moveCount:        cr.moveCount,
		redealsRemaining: cr.redealsRemaining,
		isStalemate:      cr.isStalemate,
	}
	for i := range CrescentTableauCnt {
		snap.tableau[i] = make([]*CrescentTableauCard, len(cr.tableau[i]))
		for j, tc := range cr.tableau[i] {
			snap.tableau[i][j] = &CrescentTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	for i := range CrescentFoundationCnt {
		snap.foundation[i] = make([]*Card, len(cr.foundation[i]))
		copy(snap.foundation[i], cr.foundation[i])
	}
	cr.history = append(cr.history, snap)
}

// restoreSnapshot 状態を復元する。
func (cr *Crescent) restoreSnapshot(snap *crescentSnapshot) {
	cr.tableau = snap.tableau
	cr.foundation = snap.foundation
	cr.phase = snap.phase
	cr.moveCount = snap.moveCount
	cr.redealsRemaining = snap.redealsRemaining
	cr.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加。
func (cr *Crescent) appendLog(actionType, detail string, cards []*Card) {
	cr.appendLogAt(cr.moveCount, 0, actionType, detail, cards)
}

// crescentJSON Crescent の永続化用ワイヤーフォーマット。
type crescentJSON struct {
	TrumpCards       *TrumpCards                                `json:"tc"`
	Tableau          [CrescentTableauCnt][]*CrescentTableauCard `json:"tb"`
	Foundation       [CrescentFoundationCnt][]*Card             `json:"fd"`
	Phase            CrescentPhase                              `json:"ps"`
	MoveCount        int                                        `json:"mc"`
	RedealsRemaining int                                        `json:"rd"`
	ActionLog        []*ActionLogEntry                          `json:"al"`
	IsStalemate      bool                                       `json:"sl"`
	History          []*crescentSnapshot                        `json:"hi,omitempty"`
}

// crescentSnapshotJSON 1 件のスナップショットのワイヤーフォーマット。
type crescentSnapshotJSON struct {
	Tableau          [CrescentTableauCnt][]*CrescentTableauCard `json:"tb"`
	Foundation       [CrescentFoundationCnt][]*Card             `json:"fd"`
	Phase            CrescentPhase                              `json:"ps"`
	MoveCount        int                                        `json:"mc"`
	RedealsRemaining int                                        `json:"rd"`
	IsStalemate      bool                                       `json:"sl"`
}

// MarshalJSON crescentSnapshot 用シリアライザ。
func (s *crescentSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(crescentSnapshotJSON{
		Tableau:          s.tableau,
		Foundation:       s.foundation,
		Phase:            s.phase,
		MoveCount:        s.moveCount,
		RedealsRemaining: s.redealsRemaining,
		IsStalemate:      s.isStalemate,
	})
}

// UnmarshalJSON crescentSnapshot 用デシリアライザ。
func (s *crescentSnapshot) UnmarshalJSON(data []byte) error {
	var j crescentSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > crescentMaxSliceLen {
			return fmt.Errorf("crescent: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > crescentMaxSliceLen {
			return fmt.Errorf("crescent: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.foundation = j.Foundation
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.redealsRemaining = j.RedealsRemaining
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON Crescent 用シリアライザ。
func (cr *Crescent) MarshalJSON() ([]byte, error) {
	return json.Marshal(crescentJSON{
		TrumpCards:       cr.trumpCards,
		Tableau:          cr.tableau,
		Foundation:       cr.foundation,
		Phase:            cr.phase,
		MoveCount:        cr.moveCount,
		RedealsRemaining: cr.redealsRemaining,
		ActionLog:        cr.actionLog,
		IsStalemate:      cr.isStalemate,
		History:          cr.history,
	})
}

// crescentMaxSliceLen 永続データの最大スライス長 (DoS 対策)。
const crescentMaxSliceLen = 1000

// UnmarshalJSON Crescent 用デシリアライザ。
func (cr *Crescent) UnmarshalJSON(data []byte) error {
	var j crescentJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > crescentMaxSliceLen || len(j.History) > crescentMaxSliceLen {
		return fmt.Errorf("crescent: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > crescentMaxSliceLen {
			return fmt.Errorf("crescent: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > crescentMaxSliceLen {
			return fmt.Errorf("crescent: foundation pile exceeds maximum allowed size")
		}
	}

	cr.trumpCards = j.TrumpCards
	if cr.trumpCards == nil {
		cr.trumpCards = NewTrumpCardsWithDecks(2, 0)
	}
	cr.tableau = j.Tableau
	cr.foundation = j.Foundation
	cr.phase = j.Phase
	cr.moveCount = j.MoveCount
	cr.redealsRemaining = j.RedealsRemaining
	cr.actionLog = j.ActionLog
	if cr.actionLog == nil {
		cr.actionLog = make([]*ActionLogEntry, 0)
	}
	cr.history = j.History
	if cr.history == nil {
		cr.history = make([]*crescentSnapshot, 0)
	}
	cr.isStalemate = j.IsStalemate
	return nil
}
