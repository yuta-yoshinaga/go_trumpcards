//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// KlondikePhase クロンダイクゲームフェーズ
type KlondikePhase int

// Klondikeのフェーズ定数
const (
	// KlondikePhasePlaying プレイ中
	KlondikePhasePlaying KlondikePhase = iota
	// KlondikePhaseGameClear ゲームクリア
	KlondikePhaseGameClear
	// KlondikePhaseGameOver ゲームオーバー
	KlondikePhaseGameOver
)

// KlondikeTableauCnt タブローの列数
const KlondikeTableauCnt = 7

// KlondikeFoundationCnt ファンデーションの数
const KlondikeFoundationCnt = 4

// KlondikeTableauCard タブロー上のカード
type KlondikeTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// KlondikeHint ヒント
type KlondikeHint struct {
	FromZone  string // "waste" or "tableau"
	FromCol   int
	CardIndex int
	ToZone    string // "tableau" or "foundation"
	ToCol     int
}

// KlondikeConfig クロンダイクゲーム設定
type KlondikeConfig struct {
	DrawCount   int
	ScoringMode KlondikeScoringMode
}

// KlondikeScoringMode スコアリングモード
type KlondikeScoringMode int

// Klondikeのスコアリング方式定数
const (
	// KlondikeScoringNone スコアリングなし
	KlondikeScoringNone KlondikeScoringMode = iota
	// KlondikeScoringVegas ベガススコアリング
	KlondikeScoringVegas
)

// Klondike クロンダイクゲームクラス
type Klondike struct {
	trumpCards *TrumpCards
	tableau    [KlondikeTableauCnt][]*KlondikeTableauCard
	stock      []*Card
	waste      []*Card
	foundation [KlondikeFoundationCnt][]*Card
	phase      KlondikePhase
	moveCount  int
	actionLogBase
	drawCount            int
	history              []*klondikeSnapshot
	scoringMode          KlondikeScoringMode
	isStalemate          bool
	noProgressCycles     int
	progressSinceRecycle bool
}

// klondikeSnapshot アンドゥ用スナップショット
type klondikeSnapshot struct {
	tableau              [KlondikeTableauCnt][]*KlondikeTableauCard
	stock                []*Card
	waste                []*Card
	foundation           [KlondikeFoundationCnt][]*Card
	phase                KlondikePhase
	moveCount            int
	isStalemate          bool
	noProgressCycles     int
	progressSinceRecycle bool
}

// NewKlondike コンストラクタ
func NewKlondike(trumpCards *TrumpCards) *Klondike {
	return &Klondike{
		trumpCards: trumpCards,
		drawCount:  1,
	}
}

// NewDefaultKlondike returns Klondike with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultKlondike() *Klondike {
	return NewKlondike(NewTrumpCards(0))
}

// ResetWithConfig 設定付きリセット
func (k *Klondike) ResetWithConfig(cfg KlondikeConfig) {
	if cfg.DrawCount == 3 {
		k.drawCount = 3
	} else {
		k.drawCount = 1
	}
	if cfg.ScoringMode == KlondikeScoringVegas {
		k.scoringMode = KlondikeScoringVegas
	} else {
		k.scoringMode = KlondikeScoringNone
	}
	k.Reset()
}

// Reset ゲームリセット
func (k *Klondike) Reset() {
	k.trumpCards.Shuffle()
	k.phase = KlondikePhasePlaying
	k.moveCount = 0
	k.actionLog = nil
	k.history = nil
	k.isStalemate = false
	k.noProgressCycles = 0
	k.progressSinceRecycle = false

	// タブローに配る: 列iにはi+1枚、最後だけ表
	for i := 0; i < KlondikeTableauCnt; i++ {
		k.tableau[i] = make([]*KlondikeTableauCard, 0, i+1)
		for j := 0; j <= i; j++ {
			card := k.trumpCards.DrawCard()
			tc := &KlondikeTableauCard{
				Card:   card,
				FaceUp: j == i,
			}
			k.tableau[i] = append(k.tableau[i], tc)
		}
	}

	// ファンデーション初期化
	for i := 0; i < KlondikeFoundationCnt; i++ {
		k.foundation[i] = nil
	}

	// 残りをストックへ
	k.stock = nil
	k.waste = nil
	for k.trumpCards.GetRemainingCount() > 0 {
		card := k.trumpCards.DrawCard()
		k.stock = append(k.stock, card)
	}
}

// Draw ストックからウェイストにカードを引く
func (k *Klondike) Draw() error {
	if k.phase != KlondikePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(k.stock) == 0 {
		// ウェイストをストックにリサイクル
		if len(k.waste) == 0 {
			return errors.New("no cards in stock or waste")
		}
		k.takeSnapshot()
		// ウェイストを逆順でストックに戻す
		for i := len(k.waste) - 1; i >= 0; i-- {
			k.stock = append(k.stock, k.waste[i])
		}
		k.waste = nil
		if !k.progressSinceRecycle {
			k.noProgressCycles++
		}
		k.progressSinceRecycle = false
		k.appendLog("recycle", "ウェイストをストックに戻しました", nil)
		k.checkKlondikeStalemate()
		return nil
	}
	k.takeSnapshot()
	// ストックからウェイストへ (drawCount枚)
	count := k.drawCount
	if count > len(k.stock) {
		count = len(k.stock)
	}
	drawnCards := make([]*Card, 0, count)
	for i := 0; i < count; i++ {
		card := k.stock[len(k.stock)-1]
		k.stock = k.stock[:len(k.stock)-1]
		k.waste = append(k.waste, card)
		drawnCards = append(drawnCards, card)
	}
	k.moveCount++
	k.appendLog("draw", "ストックからカードを引きました", drawnCards)
	k.checkKlondikeStalemate()
	return nil
}

// MoveWasteToTableau ウェイストからタブローにカードを移動
func (k *Klondike) MoveWasteToTableau(col int) error {
	if k.phase != KlondikePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= KlondikeTableauCnt {
		return errors.New("invalid column")
	}
	if len(k.waste) == 0 {
		return errors.New("waste is empty")
	}
	card := k.waste[len(k.waste)-1]
	if !k.canPlaceOnTableau(card, col) {
		return errors.New("cannot place card on tableau")
	}
	k.takeSnapshot()
	k.waste = k.waste[:len(k.waste)-1]
	k.tableau[col] = append(k.tableau[col], &KlondikeTableauCard{Card: card, FaceUp: true})
	k.moveCount++
	k.progressSinceRecycle = true
	k.appendLog("move", fmt.Sprintf("ウェイスト→タブロー列%d", col), []*Card{card})
	k.checkKlondikeStalemate()
	return nil
}

// MoveWasteToFoundation ウェイストからファンデーションにカードを移動
func (k *Klondike) MoveWasteToFoundation() error {
	if k.phase != KlondikePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(k.waste) == 0 {
		return errors.New("waste is empty")
	}
	card := k.waste[len(k.waste)-1]
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= KlondikeFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !k.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	k.takeSnapshot()
	k.waste = k.waste[:len(k.waste)-1]
	k.foundation[fIdx] = append(k.foundation[fIdx], card)
	k.moveCount++
	k.progressSinceRecycle = true
	k.appendLog("move", "ウェイスト→ファンデーション", []*Card{card})
	k.checkGameClear()
	k.checkKlondikeStalemate()
	return nil
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (k *Klondike) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if k.phase != KlondikePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= KlondikeTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= KlondikeTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := k.tableau[fromCol]
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	tc := fromCards[cardIndex]
	if !tc.FaceUp {
		return errors.New("card is face down")
	}
	// 移動するカード列が有効か確認（降順交互色）
	movingCards := fromCards[cardIndex:]
	bottomCard := movingCards[0].Card
	if !k.canPlaceOnTableau(bottomCard, toCol) {
		return errors.New("cannot place card on tableau")
	}
	// 移動実行
	k.takeSnapshot()
	movedCards := make([]*Card, len(movingCards))
	for i, mc := range movingCards {
		k.tableau[toCol] = append(k.tableau[toCol], mc)
		movedCards[i] = mc.Card
	}
	k.tableau[fromCol] = fromCards[:cardIndex]
	// 自動フリップ
	k.autoFlipTableau(fromCol)
	k.moveCount++
	k.progressSinceRecycle = true
	k.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), movedCards)
	k.checkKlondikeStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (k *Klondike) MoveTableauToFoundation(col int) error {
	if k.phase != KlondikePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= KlondikeTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := k.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	card := tc.Card
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= KlondikeFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !k.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	k.takeSnapshot()
	k.tableau[col] = fromCards[:len(fromCards)-1]
	k.foundation[fIdx] = append(k.foundation[fIdx], card)
	// 自動フリップ
	k.autoFlipTableau(col)
	k.moveCount++
	k.progressSinceRecycle = true
	k.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	k.checkGameClear()
	k.checkKlondikeStalemate()
	return nil
}

// GiveUp ギブアップ
func (k *Klondike) GiveUp() {
	if k.phase == KlondikePhasePlaying {
		k.phase = KlondikePhaseGameOver
		k.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (k *Klondike) GetHint() *KlondikeHint {
	if k.phase != KlondikePhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := 0; col < KlondikeTableauCnt; col++ {
		if len(k.tableau[col]) == 0 {
			continue
		}
		tc := k.tableau[col][len(k.tableau[col])-1]
		card := tc.Card
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < KlondikeFoundationCnt && k.canPlaceOnFoundation(card, fIdx) {
			return &KlondikeHint{
				FromZone:  "tableau",
				FromCol:   col,
				CardIndex: len(k.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: ウェイストからファンデーションへ
	if len(k.waste) > 0 {
		card := k.waste[len(k.waste)-1]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < KlondikeFoundationCnt && k.canPlaceOnFoundation(card, fIdx) {
			return &KlondikeHint{
				FromZone:  "waste",
				FromCol:   -1,
				CardIndex: -1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度3: タブローからタブローへ（裏カードを開けるための移動）
	for fromCol := 0; fromCol < KlondikeTableauCnt; fromCol++ {
		fromCards := k.tableau[fromCol]
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
		for toCol := 0; toCol < KlondikeTableauCnt; toCol++ {
			if toCol == fromCol {
				continue
			}
			if k.canPlaceOnTableau(card, toCol) {
				return &KlondikeHint{
					FromZone:  "tableau",
					FromCol:   fromCol,
					CardIndex: firstFaceUp,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	// 優先度4: ウェイストからタブローへ
	if len(k.waste) > 0 {
		card := k.waste[len(k.waste)-1]
		for toCol := 0; toCol < KlondikeTableauCnt; toCol++ {
			if k.canPlaceOnTableau(card, toCol) {
				return &KlondikeHint{
					FromZone:  "waste",
					FromCol:   -1,
					CardIndex: -1,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	return nil
}

// AutoComplete オートコンプリート（全カード表向きの場合に自動でファンデーションへ移動）
func (k *Klondike) AutoComplete() error {
	if k.phase != KlondikePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if !k.AllFaceUp() {
		return errors.New("not all cards are face up")
	}
	k.takeSnapshot()
	for {
		moved := false
		// ウェイストからファンデーションへ
		for len(k.waste) > 0 {
			card := k.waste[len(k.waste)-1]
			fIdx := card.GetDesign() - 1
			if !k.canPlaceOnFoundation(card, fIdx) {
				break
			}
			k.waste = k.waste[:len(k.waste)-1]
			k.foundation[fIdx] = append(k.foundation[fIdx], card)
			k.moveCount++
			moved = true
		}
		// タブローからファンデーションへ
		for col := 0; col < KlondikeTableauCnt; col++ {
			if len(k.tableau[col]) == 0 {
				continue
			}
			tc := k.tableau[col][len(k.tableau[col])-1]
			card := tc.Card
			fIdx := card.GetDesign() - 1
			if !k.canPlaceOnFoundation(card, fIdx) {
				continue
			}
			k.tableau[col] = k.tableau[col][:len(k.tableau[col])-1]
			k.foundation[fIdx] = append(k.foundation[fIdx], card)
			k.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	k.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	k.checkGameClear()
	return nil
}

// CanAutoComplete はいまオートコンプリートが実行できるかを返す (#4776)。
//
// **AutoComplete が通る条件と同じものを見る。**Web は条件が揃うとボタンを
// 光らせてバッジも出すのに、CUI は ac コマンドがあること自体もいま使えるかも
// 出していなかった。表示と実行が別条件だと、光っているのに動かない。
func (k *Klondike) CanAutoComplete() bool {
	return k.phase == KlondikePhasePlaying && k.AllFaceUp()
}

// AllFaceUp 全カードが表向きかどうか（ストックとウェイストも含む）
func (k *Klondike) AllFaceUp() bool {
	if len(k.stock) > 0 {
		return false
	}
	for col := 0; col < KlondikeTableauCnt; col++ {
		for _, tc := range k.tableau[col] {
			if !tc.FaceUp {
				return false
			}
		}
	}
	return true
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (k *Klondike) GetPhase() KlondikePhase { return k.phase }

// SetPhase フェーズ設定 (テスト用)
func (k *Klondike) SetPhase(phase KlondikePhase) { k.phase = phase }

// GetMoveCount 移動回数取得
func (k *Klondike) GetMoveCount() int { return k.moveCount }

// GetStockCount ストック枚数取得
func (k *Klondike) GetStockCount() int { return len(k.stock) }

// GetWaste ウェイスト取得
func (k *Klondike) GetWaste() []*Card { return k.waste }

// GetTableau タブロー取得
func (k *Klondike) GetTableau() [KlondikeTableauCnt][]*KlondikeTableauCard { return k.tableau }

// GetFoundation ファンデーション取得
func (k *Klondike) GetFoundation() [KlondikeFoundationCnt][]*Card { return k.foundation }

// GetGameEndFlag returns true once the game has left the playing phase.
func (k *Klondike) GetGameEndFlag() bool { return k.phase != KlondikePhasePlaying }

// IsStalemate 手詰まり状態取得
func (k *Klondike) IsStalemate() bool { return k.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (k *Klondike) SetIsStalemate(v bool) { k.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (k *Klondike) SetTableau(tableau [KlondikeTableauCnt][]*KlondikeTableauCard) {
	k.tableau = tableau
}

// SetStock ストック設定 (テスト用)
func (k *Klondike) SetStock(stock []*Card) { k.stock = stock }

// SetWaste ウェイスト設定 (テスト用)
func (k *Klondike) SetWaste(waste []*Card) { k.waste = waste }

// SetFoundation ファンデーション設定 (テスト用)
func (k *Klondike) SetFoundation(foundation [KlondikeFoundationCnt][]*Card) {
	k.foundation = foundation
}

// GetDrawCount ドローカウント取得
func (k *Klondike) GetDrawCount() int { return k.drawCount }

// SetDrawCount ドローカウント設定 (テスト用)
func (k *Klondike) SetDrawCount(n int) { k.drawCount = n }

// Undo 直前の操作を取り消す
func (k *Klondike) Undo() error {
	if k.phase != KlondikePhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(k.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := k.history[len(k.history)-1]
	k.history = k.history[:len(k.history)-1]
	k.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (k *Klondike) CanUndo() bool {
	return len(k.history) > 0 && k.phase == KlondikePhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (k *Klondike) UndoToEscape() int {
	return undoToEscape(k.isStalemate, k.history, func(s *klondikeSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (k *Klondike) UndoN(n int) error {
	return undoN(k, n)
}

// GetScore スコア取得 (ベガス式: -52 + 5 * ファンデーション枚数)
func (k *Klondike) GetScore() int {
	total := 0
	for i := 0; i < KlondikeFoundationCnt; i++ {
		total += len(k.foundation[i])
	}
	return -52 + 5*total
}

// GetScoringMode スコアリングモード取得
func (k *Klondike) GetScoringMode() KlondikeScoringMode { return k.scoringMode }

// SetScoringMode スコアリングモード設定 (テスト用)
func (k *Klondike) SetScoringMode(m KlondikeScoringMode) { k.scoringMode = m }

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
func (k *Klondike) canPlaceOnTableau(card *Card, col int) bool {
	colCards := k.tableau[col]
	if len(colCards) == 0 {
		// 空の列にはKのみ置ける
		return card.GetValue() == CardValueMax
	}
	topCard := colCards[len(colCards)-1].Card
	// 交互の色で降順
	return k.isAlternateColor(card, topCard) && card.GetValue() == topCard.GetValue()-1
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (k *Klondike) canPlaceOnFoundation(card *Card, fIdx int) bool {
	return canPlaceOnFoundationPile(k.foundation[fIdx], card)
}

// isAlternateColor 交互の色かどうか判定
func (k *Klondike) isAlternateColor(card1, card2 *Card) bool {
	return k.isBlack(card1) != k.isBlack(card2)
}

// isBlack 黒いカードかどうか
func (k *Klondike) isBlack(card *Card) bool {
	return card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover
}

// autoFlipTableau タブローの最上部の裏カードを自動フリップ
func (k *Klondike) autoFlipTableau(col int) {
	autoFlipTopCard(k.tableau[col])
}

// checkGameClear ゲームクリア判定
func (k *Klondike) checkGameClear() {
	for i := 0; i < KlondikeFoundationCnt; i++ {
		if len(k.foundation[i]) != CardValueMax {
			return
		}
	}
	k.phase = KlondikePhaseGameClear
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (k *Klondike) takeSnapshot() {
	snap := &klondikeSnapshot{
		phase:                k.phase,
		moveCount:            k.moveCount,
		isStalemate:          k.isStalemate,
		noProgressCycles:     k.noProgressCycles,
		progressSinceRecycle: k.progressSinceRecycle,
	}
	// deep copy tableau
	for i := 0; i < KlondikeTableauCnt; i++ {
		snap.tableau[i] = make([]*KlondikeTableauCard, len(k.tableau[i]))
		for j, tc := range k.tableau[i] {
			snap.tableau[i][j] = &KlondikeTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	// deep copy stock
	snap.stock = make([]*Card, len(k.stock))
	copy(snap.stock, k.stock)
	// deep copy waste
	snap.waste = make([]*Card, len(k.waste))
	copy(snap.waste, k.waste)
	// deep copy foundation
	for i := 0; i < KlondikeFoundationCnt; i++ {
		snap.foundation[i] = make([]*Card, len(k.foundation[i]))
		copy(snap.foundation[i], k.foundation[i])
	}
	k.history = append(k.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (k *Klondike) restoreSnapshot(snap *klondikeSnapshot) {
	k.tableau = snap.tableau
	k.stock = snap.stock
	k.waste = snap.waste
	k.foundation = snap.foundation
	k.phase = snap.phase
	k.moveCount = snap.moveCount
	k.isStalemate = snap.isStalemate
	k.noProgressCycles = snap.noProgressCycles
	k.progressSinceRecycle = snap.progressSinceRecycle
}

// appendLog 棋譜エントリを追加
func (k *Klondike) appendLog(actionType, detail string, cards []*Card) {
	k.appendLogAt(k.moveCount, 0, actionType, detail, cards)
}

// klondikeJSON is the JSON wire format for Klondike.
type klondikeJSON struct {
	TrumpCards           *TrumpCards                                `json:"tc"`
	Tableau              [KlondikeTableauCnt][]*KlondikeTableauCard `json:"tb"`
	Stock                []*Card                                    `json:"st"`
	Waste                []*Card                                    `json:"wa"`
	Foundation           [KlondikeFoundationCnt][]*Card             `json:"fd"`
	Phase                KlondikePhase                              `json:"ps"`
	MoveCount            int                                        `json:"mc"`
	ActionLog            []*ActionLogEntry                          `json:"al"`
	DrawCount            int                                        `json:"dc"`
	ScoringMode          KlondikeScoringMode                        `json:"sm"`
	IsStalemate          bool                                       `json:"sl"`
	NoProgressCycles     int                                        `json:"np"`
	ProgressSinceRecycle bool                                       `json:"pr"`
	History              []*klondikeSnapshot                        `json:"hi,omitempty"`
}

// klondikeSnapshotJSON is the wire format for a single undo snapshot.
// klondikeSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// klondikeJSON's short keys to keep the KV payload compact (#1654).
type klondikeSnapshotJSON struct {
	Tableau              [KlondikeTableauCnt][]*KlondikeTableauCard `json:"tb"`
	Stock                []*Card                                    `json:"st"`
	Waste                []*Card                                    `json:"wa"`
	Foundation           [KlondikeFoundationCnt][]*Card             `json:"fd"`
	Phase                KlondikePhase                              `json:"ps"`
	MoveCount            int                                        `json:"mc"`
	IsStalemate          bool                                       `json:"sl"`
	NoProgressCycles     int                                        `json:"np"`
	ProgressSinceRecycle bool                                       `json:"pr"`
}

// MarshalJSON implements json.Marshaler for klondikeSnapshot, projecting
// the unexported fields onto an exported wire shape. Used so that
// Klondike.MarshalJSON can persist the undo history (#1654).
func (s *klondikeSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(klondikeSnapshotJSON{
		Tableau:              s.tableau,
		Stock:                s.stock,
		Waste:                s.waste,
		Foundation:           s.foundation,
		Phase:                s.phase,
		MoveCount:            s.moveCount,
		IsStalemate:          s.isStalemate,
		NoProgressCycles:     s.noProgressCycles,
		ProgressSinceRecycle: s.progressSinceRecycle,
	})
}

// UnmarshalJSON implements json.Unmarshaler for klondikeSnapshot.
func (s *klondikeSnapshot) UnmarshalJSON(data []byte) error {
	var j klondikeSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > klondikeMaxSliceLen || len(j.Waste) > klondikeMaxSliceLen {
		return fmt.Errorf("klondike: snapshot array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > klondikeMaxSliceLen {
			return fmt.Errorf("klondike: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > klondikeMaxSliceLen {
			return fmt.Errorf("klondike: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.stock = j.Stock
	if s.stock == nil {
		s.stock = make([]*Card, 0)
	}
	s.waste = j.Waste
	if s.waste == nil {
		s.waste = make([]*Card, 0)
	}
	s.foundation = j.Foundation
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	s.noProgressCycles = j.NoProgressCycles
	s.progressSinceRecycle = j.ProgressSinceRecycle
	return nil
}

// MarshalJSON implements json.Marshaler.
func (k *Klondike) MarshalJSON() ([]byte, error) {
	return json.Marshal(klondikeJSON{
		TrumpCards:           k.trumpCards,
		Tableau:              k.tableau,
		Stock:                k.stock,
		Waste:                k.waste,
		Foundation:           k.foundation,
		Phase:                k.phase,
		MoveCount:            k.moveCount,
		ActionLog:            k.actionLog,
		DrawCount:            k.drawCount,
		ScoringMode:          k.scoringMode,
		IsStalemate:          k.isStalemate,
		NoProgressCycles:     k.noProgressCycles,
		ProgressSinceRecycle: k.progressSinceRecycle,
		History:              k.history,
	})
}

// klondikeMaxSliceLen caps slice sizes during deserialisation.
const klondikeMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (k *Klondike) UnmarshalJSON(data []byte) error {
	var j klondikeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > klondikeMaxSliceLen || len(j.Waste) > klondikeMaxSliceLen ||
		len(j.ActionLog) > klondikeMaxSliceLen || len(j.History) > klondikeMaxSliceLen {
		return fmt.Errorf("klondike: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > klondikeMaxSliceLen {
			return fmt.Errorf("klondike: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > klondikeMaxSliceLen {
			return fmt.Errorf("klondike: foundation pile exceeds maximum allowed size")
		}
	}

	k.trumpCards = j.TrumpCards
	if k.trumpCards == nil {
		k.trumpCards = NewTrumpCards(0)
	}
	k.tableau = j.Tableau
	k.stock = j.Stock
	if k.stock == nil {
		k.stock = make([]*Card, 0)
	}
	k.waste = j.Waste
	if k.waste == nil {
		k.waste = make([]*Card, 0)
	}
	k.foundation = j.Foundation
	k.phase = j.Phase
	k.moveCount = j.MoveCount
	k.actionLog = j.ActionLog
	if k.actionLog == nil {
		k.actionLog = make([]*ActionLogEntry, 0)
	}
	k.drawCount = j.DrawCount
	if k.drawCount == 0 {
		k.drawCount = 1
	}
	k.history = j.History
	if k.history == nil {
		k.history = make([]*klondikeSnapshot, 0)
	}
	k.scoringMode = j.ScoringMode
	k.isStalemate = j.IsStalemate
	k.noProgressCycles = j.NoProgressCycles
	k.progressSinceRecycle = j.ProgressSinceRecycle
	return nil
}

// autoFlipTopCard turns the bottom card of a tableau column face up when it is
// not already. 6 solitaires sharing KlondikeTableauCard had this written out.
//
// Lives here beside the type rather than in player_helpers.go: the other four
// games with this body use their own tableau card types, which Go's type
// constraints cannot reach because FaceUp is a field, not a method.
func autoFlipTopCard(cards []*KlondikeTableauCard) {
	if len(cards) > 0 && !cards[len(cards)-1].FaceUp {
		cards[len(cards)-1].FaceUp = true
	}
}
