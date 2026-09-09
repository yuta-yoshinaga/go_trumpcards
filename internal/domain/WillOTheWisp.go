//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// WillOTheWispPhase ウィル・オ・ザ・ウィスプゲームフェーズ
type WillOTheWispPhase int

// WillOTheWispのフェーズ定数
const (
	// WillOTheWispPhasePlaying プレイ中
	WillOTheWispPhasePlaying WillOTheWispPhase = iota
	// WillOTheWispPhaseGameClear ゲームクリア
	WillOTheWispPhaseGameClear
	// WillOTheWispPhaseGameOver ゲームオーバー
	WillOTheWispPhaseGameOver
)

// WillOTheWispTableauCnt タブローの列数 (Klondike と同じ7列)
// スコアの決まり。**説明はこの 3 つから書く。**数字を訳文に焼き込むと、
// 計算を変えたとき案内だけが古くなる (#5593)。
const (
	// WillOTheWispStartScore は開始時のスコア。
	WillOTheWispStartScore = 500
	// WillOTheWispMovePenalty は 1 手ごとに引かれる点。
	WillOTheWispMovePenalty = 1
	// WillOTheWispSuitBonus はスートを 1 組完成させるたびに入る点。
	WillOTheWispSuitBonus = 100
)

const WillOTheWispTableauCnt = 7

// WillOTheWispInitialPerColumn は初期配置で各列に配るカード数。
const WillOTheWispInitialPerColumn = 3

// WillOTheWispFoundationCnt 完成スート数 (1デッキ4スート)
const WillOTheWispFoundationCnt = 4

// WillOTheWispDealCnt Deal で各列に配るカード数 (列数と一致)
const WillOTheWispDealCnt = WillOTheWispTableauCnt

// WillOTheWispTableauCard タブロー上のカード
type WillOTheWispTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// WillOTheWispHint ヒント
type WillOTheWispHint struct {
	FromCol   int
	CardIndex int
	ToCol     int
}

// WillOTheWisp ウィル・オ・ザ・ウィスプゲームクラス
//
// 7列タブローに3枚ずつ、すべて表向きで1デッキ52枚を配り、残り31枚をストックに置く。
// タブローはスート不問の降順で積み、スパイダーソリティアと同じ「同スート降順の連続移動」と
// 「K-A同スート完成で自動除去」のルールでプレイし、4スートすべて除去で勝利する。
type WillOTheWisp struct {
	trumpCards     *TrumpCards
	tableau        [WillOTheWispTableauCnt][]*WillOTheWispTableauCard
	stock          []*Card
	completedSuits int
	phase          WillOTheWispPhase
	moveCount      int
	score          int
	actionLogBase
	history     []*willOTheWispSnapshot
	isStalemate bool
}

// willOTheWispSnapshot アンドゥ用スナップショット
type willOTheWispSnapshot struct {
	tableau        [WillOTheWispTableauCnt][]*WillOTheWispTableauCard
	stock          []*Card
	completedSuits int
	phase          WillOTheWispPhase
	moveCount      int
	score          int
	isStalemate    bool
	// actionLogLen は、アンドゥ時にログを取り消し前の長さへ切り詰めるための
	// マーカー (#1676 review)。
	actionLogLen int
}

// NewWillOTheWisp コンストラクタ
func NewWillOTheWisp(trumpCards *TrumpCards) *WillOTheWisp {
	return &WillOTheWisp{trumpCards: trumpCards}
}

// NewDefaultWillOTheWisp returns WillOTheWisp with a standard single 52-card 4-suit deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultWillOTheWisp() *WillOTheWisp {
	return NewWillOTheWisp(NewTrumpCards(0))
}

// Reset ゲームリセット
func (s *WillOTheWisp) Reset() {
	s.trumpCards.Shuffle()
	s.phase = WillOTheWispPhasePlaying
	s.moveCount = 0
	s.score = WillOTheWispStartScore
	s.actionLog = nil
	s.history = nil
	s.isStalemate = false
	s.completedSuits = 0

	// 7列に3枚ずつ、すべて表向きで初期配置する (合計21枚)。
	for i := range WillOTheWispTableauCnt {
		s.tableau[i] = make([]*WillOTheWispTableauCard, 0, WillOTheWispInitialPerColumn)
		for range WillOTheWispInitialPerColumn {
			card := s.trumpCards.DrawCard()
			tc := &WillOTheWispTableauCard{
				Card:   card,
				FaceUp: true,
			}
			s.tableau[i] = append(s.tableau[i], tc)
		}
	}

	// 残り31枚をストックへ。4回フル Deal (7枚×4=28) + 最後の Deal は
	// 残り3枚を左の3列へ配って終わる。
	s.stock = nil
	for s.trumpCards.GetRemainingCount() > 0 {
		card := s.trumpCards.DrawCard()
		s.stock = append(s.stock, card)
	}
}

// Deal ストックからタブローに1枚ずつ配る。空列がある場合は配れない。
// 山札が WillOTheWispDealCnt 未満の場合は残り全カードを左の列から配る
// (標準 WillOTheWisp ルール) ので、最後の3枚も到達可能。
// GetDealsRemaining は「配る」をあと何回押せるかを返す (#4798)。
//
// **生の残り枚数だけでは分からない。**1回の配布は最大 WillOTheWispDealCnt 枚で、
// 端数 (1〜6枚) の最終配布も1回として数える。Web は同じ切り上げをバッジに
// 出しているのに、CUI は7で割って切り上げる暗算を強いていた。
//
// 空き列があると Deal は弾かれるが、それは一時的な状態なのでここでは見ない
// (回数そのものは変わらない)。
func (s *WillOTheWisp) GetDealsRemaining() int {
	n := len(s.stock)
	if n <= 0 {
		return 0
	}
	return (n + WillOTheWispDealCnt - 1) / WillOTheWispDealCnt
}

func (s *WillOTheWisp) Deal() error {
	if s.phase != WillOTheWispPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(s.stock) == 0 {
		return errors.New("not enough cards in stock")
	}
	for i := range WillOTheWispTableauCnt {
		if len(s.tableau[i]) == 0 {
			return errors.New("cannot deal: empty column exists")
		}
	}
	s.takeSnapshot()
	numToDeal := len(s.stock)
	if numToDeal > WillOTheWispDealCnt {
		numToDeal = WillOTheWispDealCnt
	}
	for i := 0; i < numToDeal; i++ {
		card := s.stock[len(s.stock)-1]
		s.stock = s.stock[:len(s.stock)-1]
		s.tableau[i] = append(s.tableau[i], &WillOTheWispTableauCard{Card: card, FaceUp: true})
	}
	s.moveCount++
	s.score -= WillOTheWispMovePenalty
	s.appendLog("deal", "ストックから各列にカードを配りました", nil)
	for i := range WillOTheWispTableauCnt {
		s.checkAndRemoveCompletedSuit(i)
	}
	s.checkWillOTheWispStalemate()
	return nil
}

// MoveTableauToTableau タブロー間で同スート降順の連続をまとめて移動する。
func (s *WillOTheWisp) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if s.phase != WillOTheWispPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= WillOTheWispTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= WillOTheWispTableauCnt {
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
	s.moveCount++
	s.score -= WillOTheWispMovePenalty
	s.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), movedCards)
	s.checkAndRemoveCompletedSuit(toCol)
	s.checkWillOTheWispStalemate()
	return nil
}

// GiveUp ギブアップ
func (s *WillOTheWisp) GiveUp() {
	if s.phase == WillOTheWispPhasePlaying {
		s.phase = WillOTheWispPhaseGameOver
		s.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (s *WillOTheWisp) GetHint() *WillOTheWispHint {
	if s.phase != WillOTheWispPhasePlaying {
		return nil
	}
	for _, exposeOnly := range []bool{true, false} {
		for fromCol := range WillOTheWispTableauCnt {
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
				for toCol := range WillOTheWispTableauCnt {
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
					return &WillOTheWispHint{
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
func (s *WillOTheWisp) AutoComplete() error {
	if s.phase != WillOTheWispPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if !s.AllFaceUp() {
		return errors.New("not all cards are face up")
	}
	s.takeSnapshot()
	for {
		removed := false
		for col := range WillOTheWispTableauCnt {
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
func (s *WillOTheWisp) AllFaceUp() bool {
	return true
}

// Undo 直前の操作を取り消す
func (s *WillOTheWisp) Undo() error {
	if s.phase != WillOTheWispPhasePlaying {
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
func (s *WillOTheWisp) CanUndo() bool {
	return len(s.history) > 0 && s.phase == WillOTheWispPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数。膠着でなければ0、脱出不可なら-1。
func (s *WillOTheWisp) UndoToEscape() int {
	return undoToEscape(s.isStalemate, s.history, func(s *willOTheWispSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (s *WillOTheWisp) UndoN(n int) error {
	return undoN(s, n)
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (s *WillOTheWisp) GetPhase() WillOTheWispPhase { return s.phase }

// SetPhase フェーズ設定 (テスト用)
func (s *WillOTheWisp) SetPhase(phase WillOTheWispPhase) { s.phase = phase }

// GetMoveCount 移動回数取得
func (s *WillOTheWisp) GetMoveCount() int { return s.moveCount }

// GetStockCount ストック枚数取得
func (s *WillOTheWisp) GetStockCount() int { return len(s.stock) }

// GetTableau タブロー取得
func (s *WillOTheWisp) GetTableau() [WillOTheWispTableauCnt][]*WillOTheWispTableauCard {
	return s.tableau
}

// GetCompletedSuits 完成スート数取得
func (s *WillOTheWisp) GetCompletedSuits() int { return s.completedSuits }

// GetGameEndFlag returns true once the game has left the playing phase.
func (s *WillOTheWisp) GetGameEndFlag() bool { return s.phase != WillOTheWispPhasePlaying }

// IsStalemate 手詰まり状態取得
func (s *WillOTheWisp) IsStalemate() bool { return s.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (s *WillOTheWisp) SetIsStalemate(v bool) { s.isStalemate = v }

// GetScore スコア取得
func (s *WillOTheWisp) GetScore() int { return s.score }

// SetTableau タブロー設定 (テスト用)
func (s *WillOTheWisp) SetTableau(tableau [WillOTheWispTableauCnt][]*WillOTheWispTableauCard) {
	s.tableau = tableau
}

// SetStock ストック設定 (テスト用)
func (s *WillOTheWisp) SetStock(stock []*Card) { s.stock = stock }

// SetCompletedSuits 完成スート数設定 (テスト用)
func (s *WillOTheWisp) SetCompletedSuits(n int) { s.completedSuits = n }

// SetScore スコア設定 (テスト用)
func (s *WillOTheWisp) SetScore(score int) { s.score = score }

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定 (スート不問・値が1つ大きい上のみ可)
func (s *WillOTheWisp) canPlaceOnTableau(card *Card, col int) bool {
	colCards := s.tableau[col]
	if len(colCards) == 0 {
		return true
	}
	topCard := colCards[len(colCards)-1].Card
	return card.GetValue() == topCard.GetValue()-1
}

// isValidSequence 同スート降順の連続かどうか判定
func (s *WillOTheWisp) isValidSequence(cards []*WillOTheWispTableauCard) bool {
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
func (s *WillOTheWisp) checkAndRemoveCompletedSuit(col int) bool {
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
	s.score += WillOTheWispSuitBonus
	s.appendLog("complete", fmt.Sprintf("タブロー列%dでスートが完成しました", col), nil)

	s.checkGameClear()
	return true
}

// checkGameClear ゲームクリア判定
func (s *WillOTheWisp) checkGameClear() {
	if s.completedSuits >= WillOTheWispFoundationCnt {
		s.phase = WillOTheWispPhaseGameClear
	}
}

// checkWillOTheWispStalemate 手詰まり判定
func (s *WillOTheWisp) checkWillOTheWispStalemate() {
	if s.phase != WillOTheWispPhasePlaying {
		return
	}
	if hint := s.GetHint(); hint != nil {
		s.isStalemate = false
		return
	}
	if len(s.stock) > 0 {
		hasEmpty := false
		for i := range WillOTheWispTableauCnt {
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
func (s *WillOTheWisp) takeSnapshot() {
	snap := &willOTheWispSnapshot{
		completedSuits: s.completedSuits,
		phase:          s.phase,
		moveCount:      s.moveCount,
		score:          s.score,
		isStalemate:    s.isStalemate,
		actionLogLen:   len(s.actionLog),
	}
	for i := range WillOTheWispTableauCnt {
		snap.tableau[i] = make([]*WillOTheWispTableauCard, len(s.tableau[i]))
		for j, tc := range s.tableau[i] {
			snap.tableau[i][j] = &WillOTheWispTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	snap.stock = make([]*Card, len(s.stock))
	copy(snap.stock, s.stock)
	s.history = appendSnapshot(s.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (s *WillOTheWisp) restoreSnapshot(snap *willOTheWispSnapshot) {
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
func (s *WillOTheWisp) appendLog(actionType, detail string, cards []*Card) {
	s.appendLogAt(s.moveCount, 0, actionType, detail, cards)
}

// willOTheWispJSON is the JSON wire format for WillOTheWisp.
type willOTheWispJSON struct {
	TrumpCards     *TrumpCards                                        `json:"tc"`
	Tableau        [WillOTheWispTableauCnt][]*WillOTheWispTableauCard `json:"tb"`
	Stock          []*Card                                            `json:"st"`
	CompletedSuits int                                                `json:"cs"`
	Phase          WillOTheWispPhase                                  `json:"ps"`
	MoveCount      int                                                `json:"mc"`
	Score          int                                                `json:"sc"`
	ActionLog      []*ActionLogEntry                                  `json:"al"`
	IsStalemate    bool                                               `json:"sm"`
	History        []*willOTheWispSnapshot                            `json:"hi,omitempty"`
}

// willOTheWispSnapshotJSON is the wire format for a single undo snapshot.
type willOTheWispSnapshotJSON struct {
	Tableau        [WillOTheWispTableauCnt][]*WillOTheWispTableauCard `json:"tb"`
	Stock          []*Card                                            `json:"st"`
	CompletedSuits int                                                `json:"cs"`
	Phase          WillOTheWispPhase                                  `json:"ps"`
	MoveCount      int                                                `json:"mc"`
	Score          int                                                `json:"sc"`
	IsStalemate    bool                                               `json:"sm"`
	ActionLogLen   int                                                `json:"al"`
}

// MarshalJSON implements json.Marshaler for willOTheWispSnapshot.
func (s *willOTheWispSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(willOTheWispSnapshotJSON{
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

// UnmarshalJSON implements json.Unmarshaler for willOTheWispSnapshot.
func (s *willOTheWispSnapshot) UnmarshalJSON(data []byte) error {
	var j willOTheWispSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > willOTheWispMaxSliceLen {
		return fmt.Errorf("willOTheWisp: snapshot array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > willOTheWispMaxSliceLen {
			return fmt.Errorf("willOTheWisp: snapshot tableau column exceeds maximum allowed size")
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
func (s *WillOTheWisp) MarshalJSON() ([]byte, error) {
	return json.Marshal(willOTheWispJSON{
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

// willOTheWispMaxSliceLen caps slice sizes during deserialisation.
const willOTheWispMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (s *WillOTheWisp) UnmarshalJSON(data []byte) error {
	var j willOTheWispJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > willOTheWispMaxSliceLen || len(j.ActionLog) > willOTheWispMaxSliceLen ||
		len(j.History) > willOTheWispMaxSliceLen {
		return fmt.Errorf("willOTheWisp: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > willOTheWispMaxSliceLen {
			return fmt.Errorf("willOTheWisp: tableau column exceeds maximum allowed size")
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
		s.history = make([]*willOTheWispSnapshot, 0)
	}
	s.isStalemate = j.IsStalemate
	return nil
}
