package domain

import (
	"errors"
	"fmt"
)

// KlondikePhase クロンダイクゲームフェーズ
type KlondikePhase int

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
	Card   *Card
	FaceUp bool
}

// KlondikeHint ヒント
type KlondikeHint struct {
	FromZone  string // "waste" or "tableau"
	FromCol   int
	CardIndex int
	ToZone    string // "tableau" or "foundation"
	ToCol     int
}

// Klondike クロンダイクゲームクラス
type Klondike struct {
	trumpCards *TrumpCards
	tableau    [KlondikeTableauCnt][]*KlondikeTableauCard
	stock      []*Card
	waste      []*Card
	foundation [KlondikeFoundationCnt][]*Card
	phase      KlondikePhase
	moveCount  int
	actionLog  []*ActionLogEntry
}

// NewKlondike コンストラクタ
func NewKlondike(trumpCards *TrumpCards) *Klondike {
	return &Klondike{
		trumpCards: trumpCards,
	}
}

// Reset ゲームリセット
func (k *Klondike) Reset() {
	k.trumpCards.Shuffle()
	k.phase = KlondikePhasePlaying
	k.moveCount = 0
	k.actionLog = nil

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
		// ウェイストを逆順でストックに戻す
		for i := len(k.waste) - 1; i >= 0; i-- {
			k.stock = append(k.stock, k.waste[i])
		}
		k.waste = nil
		k.appendLog("recycle", "ウェイストをストックに戻しました", nil)
		return nil
	}
	// ストックからウェイストへ
	card := k.stock[len(k.stock)-1]
	k.stock = k.stock[:len(k.stock)-1]
	k.waste = append(k.waste, card)
	k.moveCount++
	k.appendLog("draw", "ストックからカードを引きました", []*Card{card})
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
	k.waste = k.waste[:len(k.waste)-1]
	k.tableau[col] = append(k.tableau[col], &KlondikeTableauCard{Card: card, FaceUp: true})
	k.moveCount++
	k.appendLog("move", fmt.Sprintf("ウェイスト→タブロー列%d", col), []*Card{card})
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
	k.waste = k.waste[:len(k.waste)-1]
	k.foundation[fIdx] = append(k.foundation[fIdx], card)
	k.moveCount++
	k.appendLog("move", "ウェイスト→ファンデーション", []*Card{card})
	k.checkGameClear()
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
	movedCards := make([]*Card, len(movingCards))
	for i, mc := range movingCards {
		k.tableau[toCol] = append(k.tableau[toCol], mc)
		movedCards[i] = mc.Card
	}
	k.tableau[fromCol] = fromCards[:cardIndex]
	// 自動フリップ
	k.autoFlipTableau(fromCol)
	k.moveCount++
	k.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), movedCards)
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
	k.tableau[col] = fromCards[:len(fromCards)-1]
	k.foundation[fIdx] = append(k.foundation[fIdx], card)
	// 自動フリップ
	k.autoFlipTableau(col)
	k.moveCount++
	k.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	k.checkGameClear()
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
	for {
		moved := false
		// ウェイストからファンデーションへ
		for len(k.waste) > 0 {
			card := k.waste[len(k.waste)-1]
			fIdx := card.GetDesign() - 1
			if fIdx < 0 || fIdx >= KlondikeFoundationCnt || !k.canPlaceOnFoundation(card, fIdx) {
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
			if fIdx < 0 || fIdx >= KlondikeFoundationCnt || !k.canPlaceOnFoundation(card, fIdx) {
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

// GetActionLog 棋譜取得
func (k *Klondike) GetActionLog() []*ActionLogEntry { return k.actionLog }

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
	pile := k.foundation[fIdx]
	if len(pile) == 0 {
		// 空のファンデーションにはAのみ置ける
		return card.GetValue() == 1
	}
	topCard := pile[len(pile)-1]
	// 同じスートで昇順
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()+1
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
	cards := k.tableau[col]
	if len(cards) > 0 && !cards[len(cards)-1].FaceUp {
		cards[len(cards)-1].FaceUp = true
	}
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

// appendLog 棋譜エントリを追加
func (k *Klondike) appendLog(actionType, detail string, cards []*Card) {
	k.actionLog = append(k.actionLog, &ActionLogEntry{
		TurnNumber: k.moveCount,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}
