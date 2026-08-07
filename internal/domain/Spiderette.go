//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// SpiderettePhase スパイダレットゲームフェーズ
type SpiderettePhase int

// Spideretteのフェーズ定数
const (
	// SpiderettePhasePlaying プレイ中
	SpiderettePhasePlaying SpiderettePhase = iota
	// SpiderettePhaseGameClear ゲームクリア
	SpiderettePhaseGameClear
	// SpiderettePhaseGameOver ゲームオーバー
	SpiderettePhaseGameOver
)

// SpideretteTableauCnt タブローの列数 (Klondike と同じ7列)
const SpideretteTableauCnt = 7

// SpideretteFoundationCnt 完成スート数 (1デッキ4スート)
const SpideretteFoundationCnt = 4

// SpideretteDealCnt Deal で各列に配るカード数 (列数と一致)
const SpideretteDealCnt = SpideretteTableauCnt

// SpideretteTableauCard タブロー上のカード
type SpideretteTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// SpideretteHint ヒント
type SpideretteHint struct {
	FromCol   int
	CardIndex int
	ToCol     int
}

// Spiderette スパイダレットゲームクラス
//
// クロンダイクと同じ7列タブロー（列iにi+1枚、最後の1枚だけ表向き）に1デッキ52枚を配り、
// 残り24枚をストックに置く。スパイダーソリティアと同じ「同スート降順の連続移動」と
// 「K-A同スート完成で自動除去」のルールでプレイし、4スートすべて除去で勝利する。
type Spiderette struct {
	trumpCards     *TrumpCards
	tableau        [SpideretteTableauCnt][]*SpideretteTableauCard
	stock          []*Card
	completedSuits int
	phase          SpiderettePhase
	moveCount      int
	score          int
	actionLog      []*ActionLogEntry
	history        []*spideretteSnapshot
	isStalemate    bool
}

// spideretteSnapshot アンドゥ用スナップショット
type spideretteSnapshot struct {
	tableau        [SpideretteTableauCnt][]*SpideretteTableauCard
	stock          []*Card
	completedSuits int
	phase          SpiderettePhase
	moveCount      int
	score          int
	isStalemate    bool
	// actionLogLen は、アンドゥ時にログを取り消し前の長さへ切り詰めるための
	// マーカー (#1676 review)。
	actionLogLen int
}

// NewSpiderette コンストラクタ
func NewSpiderette(trumpCards *TrumpCards) *Spiderette {
	return &Spiderette{trumpCards: trumpCards}
}

// NewDefaultSpiderette returns Spiderette with a standard single 52-card 4-suit deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultSpiderette() *Spiderette {
	return NewSpiderette(NewTrumpCards(0))
}

// Reset ゲームリセット
func (s *Spiderette) Reset() {
	s.trumpCards.Shuffle()
	s.phase = SpiderettePhasePlaying
	s.moveCount = 0
	s.score = 500
	s.actionLog = nil
	s.history = nil
	s.isStalemate = false
	s.completedSuits = 0

	// Klondike と同じ階段配置: 列iにi+1枚、最後だけ表向き (合計 28 枚)
	for i := range SpideretteTableauCnt {
		s.tableau[i] = make([]*SpideretteTableauCard, 0, i+1)
		for j := 0; j <= i; j++ {
			card := s.trumpCards.DrawCard()
			tc := &SpideretteTableauCard{
				Card:   card,
				FaceUp: j == i,
			}
			s.tableau[i] = append(s.tableau[i], tc)
		}
	}

	// 残り24枚をストックへ。3回フル Deal (7枚×3=21) + 最後の Deal は
	// 残り3枚を左の3列へ配って終わり (標準 Spiderette のルール: 最後だけ
	// 部分的に配る)。
	s.stock = nil
	for s.trumpCards.GetRemainingCount() > 0 {
		card := s.trumpCards.DrawCard()
		s.stock = append(s.stock, card)
	}
}

// Deal ストックからタブローに1枚ずつ配る。空列がある場合は配れない。
// 山札が SpideretteDealCnt 未満の場合は残り全カードを左の列から配る
// (標準 Spiderette ルール) ので、最後の3枚も到達可能。
// GetDealsRemaining は「配る」をあと何回押せるかを返す (#4798)。
//
// **生の残り枚数だけでは分からない。**1回の配布は最大 SpideretteDealCnt 枚で、
// 端数 (1〜6枚) の最終配布も1回として数える。Web は同じ切り上げをバッジに
// 出しているのに、CUI は7で割って切り上げる暗算を強いていた。
//
// 空き列があると Deal は弾かれるが、それは一時的な状態なのでここでは見ない
// (回数そのものは変わらない)。
func (s *Spiderette) GetDealsRemaining() int {
	n := len(s.stock)
	if n <= 0 {
		return 0
	}
	return (n + SpideretteDealCnt - 1) / SpideretteDealCnt
}

func (s *Spiderette) Deal() error {
	if s.phase != SpiderettePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(s.stock) == 0 {
		return errors.New("not enough cards in stock")
	}
	for i := range SpideretteTableauCnt {
		if len(s.tableau[i]) == 0 {
			return errors.New("cannot deal: empty column exists")
		}
	}
	s.takeSnapshot()
	numToDeal := len(s.stock)
	if numToDeal > SpideretteDealCnt {
		numToDeal = SpideretteDealCnt
	}
	for i := 0; i < numToDeal; i++ {
		card := s.stock[len(s.stock)-1]
		s.stock = s.stock[:len(s.stock)-1]
		s.tableau[i] = append(s.tableau[i], &SpideretteTableauCard{Card: card, FaceUp: true})
	}
	s.moveCount++
	s.score--
	s.appendLog("deal", "ストックから各列にカードを配りました", nil)
	for i := range SpideretteTableauCnt {
		s.checkAndRemoveCompletedSuit(i)
	}
	s.checkSpideretteStalemate()
	return nil
}

// MoveTableauToTableau タブロー間で同スート降順の連続をまとめて移動する。
func (s *Spiderette) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if s.phase != SpiderettePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= SpideretteTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= SpideretteTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := s.tableau[fromCol]
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

	movingCards := fromCards[cardIndex:]
	if !s.isValidSequence(movingCards) {
		return errors.New("cards are not a valid same-suit descending sequence")
	}

	bottomCard := movingCards[0].Card
	if !s.canPlaceOnTableau(bottomCard, toCol) {
		return errors.New("cannot place card on tableau")
	}

	s.takeSnapshot()
	movedCards := make([]*Card, len(movingCards))
	for i, mc := range movingCards {
		s.tableau[toCol] = append(s.tableau[toCol], mc)
		movedCards[i] = mc.Card
	}
	s.tableau[fromCol] = fromCards[:cardIndex]
	s.autoFlipTableau(fromCol)
	s.moveCount++
	s.score--
	s.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), movedCards)
	s.checkAndRemoveCompletedSuit(toCol)
	s.checkSpideretteStalemate()
	return nil
}

// GiveUp ギブアップ
func (s *Spiderette) GiveUp() {
	if s.phase == SpiderettePhasePlaying {
		s.phase = SpiderettePhaseGameOver
		s.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (s *Spiderette) GetHint() *SpideretteHint {
	if s.phase != SpiderettePhasePlaying {
		return nil
	}
	for _, exposeOnly := range []bool{true, false} {
		for fromCol := range SpideretteTableauCnt {
			fromCards := s.tableau[fromCol]
			if len(fromCards) == 0 {
				continue
			}
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
			if exposeOnly && firstFaceUp == 0 {
				continue
			}
			for startIdx := firstFaceUp; startIdx < len(fromCards); startIdx++ {
				movingCards := fromCards[startIdx:]
				if !s.isValidSequence(movingCards) {
					continue
				}
				bottomCard := movingCards[0].Card
				for toCol := range SpideretteTableauCnt {
					if toCol == fromCol {
						continue
					}
					if !s.canPlaceOnTableau(bottomCard, toCol) {
						continue
					}
					if len(s.tableau[toCol]) == 0 && startIdx == 0 {
						continue
					}
					if exposeOnly && startIdx != firstFaceUp {
						continue
					}
					return &SpideretteHint{
						FromCol:   fromCol,
						CardIndex: startIdx,
						ToCol:     toCol,
					}
				}
			}
		}
	}
	return nil
}

// AutoComplete オートコンプリート（全カード表向きの場合に完成スートを自動除去）
func (s *Spiderette) AutoComplete() error {
	if s.phase != SpiderettePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if !s.AllFaceUp() {
		return errors.New("not all cards are face up")
	}
	s.takeSnapshot()
	for {
		removed := false
		for col := range SpideretteTableauCnt {
			if s.checkAndRemoveCompletedSuit(col) {
				removed = true
			}
		}
		if !removed {
			break
		}
	}
	s.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	s.checkGameClear()
	return nil
}

// AllFaceUp 全カードが表向きかどうか
func (s *Spiderette) AllFaceUp() bool {
	if len(s.stock) > 0 {
		return false
	}
	for col := range SpideretteTableauCnt {
		for _, tc := range s.tableau[col] {
			if !tc.FaceUp {
				return false
			}
		}
	}
	return true
}

// Undo 直前の操作を取り消す
func (s *Spiderette) Undo() error {
	if s.phase != SpiderettePhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(s.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := s.history[len(s.history)-1]
	s.history = s.history[:len(s.history)-1]
	s.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (s *Spiderette) CanUndo() bool {
	return len(s.history) > 0 && s.phase == SpiderettePhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数。膠着でなければ0、脱出不可なら-1。
func (s *Spiderette) UndoToEscape() int {
	if !s.isStalemate {
		return 0
	}
	for i := len(s.history) - 1; i >= 0; i-- {
		if !s.history[i].isStalemate {
			return len(s.history) - i
		}
	}
	return -1
}

// UndoN n回連続でアンドゥを実行する。
func (s *Spiderette) UndoN(n int) error {
	for i := 0; i < n; i++ {
		if err := s.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (s *Spiderette) GetPhase() SpiderettePhase { return s.phase }

// SetPhase フェーズ設定 (テスト用)
func (s *Spiderette) SetPhase(phase SpiderettePhase) { s.phase = phase }

// GetMoveCount 移動回数取得
func (s *Spiderette) GetMoveCount() int { return s.moveCount }

// GetStockCount ストック枚数取得
func (s *Spiderette) GetStockCount() int { return len(s.stock) }

// GetTableau タブロー取得
func (s *Spiderette) GetTableau() [SpideretteTableauCnt][]*SpideretteTableauCard { return s.tableau }

// GetCompletedSuits 完成スート数取得
func (s *Spiderette) GetCompletedSuits() int { return s.completedSuits }

// GetActionLog 棋譜取得
func (s *Spiderette) GetActionLog() []*ActionLogEntry { return s.actionLog }

// GetGameEndFlag returns true once the game has left the playing phase.
func (s *Spiderette) GetGameEndFlag() bool { return s.phase != SpiderettePhasePlaying }

// IsStalemate 手詰まり状態取得
func (s *Spiderette) IsStalemate() bool { return s.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (s *Spiderette) SetIsStalemate(v bool) { s.isStalemate = v }

// GetScore スコア取得
func (s *Spiderette) GetScore() int { return s.score }

// SetTableau タブロー設定 (テスト用)
func (s *Spiderette) SetTableau(tableau [SpideretteTableauCnt][]*SpideretteTableauCard) {
	s.tableau = tableau
}

// SetStock ストック設定 (テスト用)
func (s *Spiderette) SetStock(stock []*Card) { s.stock = stock }

// SetCompletedSuits 完成スート数設定 (テスト用)
func (s *Spiderette) SetCompletedSuits(n int) { s.completedSuits = n }

// SetScore スコア設定 (テスト用)
func (s *Spiderette) SetScore(score int) { s.score = score }

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定 (スート不問・値が1つ大きい上のみ可)
func (s *Spiderette) canPlaceOnTableau(card *Card, col int) bool {
	colCards := s.tableau[col]
	if len(colCards) == 0 {
		return true
	}
	topCard := colCards[len(colCards)-1].Card
	return card.GetValue() == topCard.GetValue()-1
}

// isValidSequence 同スート降順の連続かどうか判定
func (s *Spiderette) isValidSequence(cards []*SpideretteTableauCard) bool {
	if len(cards) <= 1 {
		return true
	}
	for i := 1; i < len(cards); i++ {
		prev := cards[i-1].Card
		curr := cards[i].Card
		if curr.GetDesign() != prev.GetDesign() {
			return false
		}
		if curr.GetValue() != prev.GetValue()-1 {
			return false
		}
		if !cards[i].FaceUp {
			return false
		}
	}
	return true
}

// checkAndRemoveCompletedSuit 列末尾13枚がK-A同スートか判定し、該当なら除去する。
func (s *Spiderette) checkAndRemoveCompletedSuit(col int) bool {
	cards := s.tableau[col]
	if len(cards) < CardValueMax {
		return false
	}
	startIdx := len(cards) - CardValueMax
	seq := cards[startIdx:]

	if seq[0].Card.GetValue() != CardValueMax {
		return false
	}
	if seq[len(seq)-1].Card.GetValue() != 1 {
		return false
	}

	suit := seq[0].Card.GetDesign()
	for i, tc := range seq {
		if !tc.FaceUp {
			return false
		}
		if tc.Card.GetDesign() != suit {
			return false
		}
		if tc.Card.GetValue() != CardValueMax-i {
			return false
		}
	}

	s.tableau[col] = cards[:startIdx]
	s.completedSuits++
	s.score += 100
	s.appendLog("complete", fmt.Sprintf("タブロー列%dでスートが完成しました", col), nil)

	s.autoFlipTableau(col)
	s.checkGameClear()
	return true
}

// autoFlipTableau タブローの最上部の裏カードを自動フリップ
func (s *Spiderette) autoFlipTableau(col int) {
	cards := s.tableau[col]
	if len(cards) > 0 && !cards[len(cards)-1].FaceUp {
		cards[len(cards)-1].FaceUp = true
	}
}

// checkGameClear ゲームクリア判定
func (s *Spiderette) checkGameClear() {
	if s.completedSuits >= SpideretteFoundationCnt {
		s.phase = SpiderettePhaseGameClear
	}
}

// checkSpideretteStalemate 手詰まり判定
func (s *Spiderette) checkSpideretteStalemate() {
	if s.phase != SpiderettePhasePlaying {
		return
	}
	if hint := s.GetHint(); hint != nil {
		s.isStalemate = false
		return
	}
	if len(s.stock) > 0 {
		hasEmpty := false
		for i := range SpideretteTableauCnt {
			if len(s.tableau[i]) == 0 {
				hasEmpty = true
				break
			}
		}
		if !hasEmpty {
			s.isStalemate = false
			return
		}
	}
	s.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (s *Spiderette) takeSnapshot() {
	snap := &spideretteSnapshot{
		completedSuits: s.completedSuits,
		phase:          s.phase,
		moveCount:      s.moveCount,
		score:          s.score,
		isStalemate:    s.isStalemate,
		actionLogLen:   len(s.actionLog),
	}
	for i := range SpideretteTableauCnt {
		snap.tableau[i] = make([]*SpideretteTableauCard, len(s.tableau[i]))
		for j, tc := range s.tableau[i] {
			snap.tableau[i][j] = &SpideretteTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	snap.stock = make([]*Card, len(s.stock))
	copy(snap.stock, s.stock)
	s.history = append(s.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (s *Spiderette) restoreSnapshot(snap *spideretteSnapshot) {
	s.tableau = snap.tableau
	s.stock = snap.stock
	s.completedSuits = snap.completedSuits
	s.phase = snap.phase
	s.moveCount = snap.moveCount
	s.score = snap.score
	s.isStalemate = snap.isStalemate
	if snap.actionLogLen >= 0 && snap.actionLogLen <= len(s.actionLog) {
		s.actionLog = s.actionLog[:snap.actionLogLen]
	}
}

// appendLog 棋譜エントリを追加
func (s *Spiderette) appendLog(actionType, detail string, cards []*Card) {
	s.actionLog = append(s.actionLog, &ActionLogEntry{
		TurnNumber: s.moveCount,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// spideretteJSON is the JSON wire format for Spiderette.
type spideretteJSON struct {
	TrumpCards     *TrumpCards                                    `json:"tc"`
	Tableau        [SpideretteTableauCnt][]*SpideretteTableauCard `json:"tb"`
	Stock          []*Card                                        `json:"st"`
	CompletedSuits int                                            `json:"cs"`
	Phase          SpiderettePhase                                `json:"ps"`
	MoveCount      int                                            `json:"mc"`
	Score          int                                            `json:"sc"`
	ActionLog      []*ActionLogEntry                              `json:"al"`
	IsStalemate    bool                                           `json:"sm"`
	History        []*spideretteSnapshot                          `json:"hi,omitempty"`
}

// spideretteSnapshotJSON is the wire format for a single undo snapshot.
type spideretteSnapshotJSON struct {
	Tableau        [SpideretteTableauCnt][]*SpideretteTableauCard `json:"tb"`
	Stock          []*Card                                        `json:"st"`
	CompletedSuits int                                            `json:"cs"`
	Phase          SpiderettePhase                                `json:"ps"`
	MoveCount      int                                            `json:"mc"`
	Score          int                                            `json:"sc"`
	IsStalemate    bool                                           `json:"sm"`
	ActionLogLen   int                                            `json:"al"`
}

// MarshalJSON implements json.Marshaler for spideretteSnapshot.
func (s *spideretteSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(spideretteSnapshotJSON{
		Tableau:        s.tableau,
		Stock:          s.stock,
		CompletedSuits: s.completedSuits,
		Phase:          s.phase,
		MoveCount:      s.moveCount,
		Score:          s.score,
		IsStalemate:    s.isStalemate,
		ActionLogLen:   s.actionLogLen,
	})
}

// UnmarshalJSON implements json.Unmarshaler for spideretteSnapshot.
func (s *spideretteSnapshot) UnmarshalJSON(data []byte) error {
	var j spideretteSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > spideretteMaxSliceLen {
		return fmt.Errorf("spiderette: snapshot array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > spideretteMaxSliceLen {
			return fmt.Errorf("spiderette: snapshot tableau column exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.stock = j.Stock
	if s.stock == nil {
		s.stock = make([]*Card, 0)
	}
	s.completedSuits = j.CompletedSuits
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.score = j.Score
	s.isStalemate = j.IsStalemate
	s.actionLogLen = j.ActionLogLen
	return nil
}

// MarshalJSON implements json.Marshaler.
func (s *Spiderette) MarshalJSON() ([]byte, error) {
	return json.Marshal(spideretteJSON{
		TrumpCards:     s.trumpCards,
		Tableau:        s.tableau,
		Stock:          s.stock,
		CompletedSuits: s.completedSuits,
		Phase:          s.phase,
		MoveCount:      s.moveCount,
		Score:          s.score,
		ActionLog:      s.actionLog,
		IsStalemate:    s.isStalemate,
		History:        s.history,
	})
}

// spideretteMaxSliceLen caps slice sizes during deserialisation.
const spideretteMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (s *Spiderette) UnmarshalJSON(data []byte) error {
	var j spideretteJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > spideretteMaxSliceLen || len(j.ActionLog) > spideretteMaxSliceLen ||
		len(j.History) > spideretteMaxSliceLen {
		return fmt.Errorf("spiderette: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > spideretteMaxSliceLen {
			return fmt.Errorf("spiderette: tableau column exceeds maximum allowed size")
		}
	}

	s.trumpCards = j.TrumpCards
	if s.trumpCards == nil {
		s.trumpCards = NewTrumpCards(0)
	}
	s.tableau = j.Tableau
	s.stock = j.Stock
	if s.stock == nil {
		s.stock = make([]*Card, 0)
	}
	s.completedSuits = j.CompletedSuits
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.score = j.Score
	s.actionLog = j.ActionLog
	if s.actionLog == nil {
		s.actionLog = make([]*ActionLogEntry, 0)
	}
	s.history = j.History
	if s.history == nil {
		s.history = make([]*spideretteSnapshot, 0)
	}
	s.isStalemate = j.IsStalemate
	return nil
}
