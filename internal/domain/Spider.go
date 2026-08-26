//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// SpiderPhase スパイダーソリティアゲームフェーズ
type SpiderPhase int

// Spiderのフェーズ定数
const (
	// SpiderPhasePlaying プレイ中
	SpiderPhasePlaying SpiderPhase = iota
	// SpiderPhaseGameClear ゲームクリア
	SpiderPhaseGameClear
	// SpiderPhaseGameOver ゲームオーバー
	SpiderPhaseGameOver
)

// SpiderTableauCnt タブローの列数
const SpiderTableauCnt = 10

// SpiderFoundationCnt ファンデーションの数（完成スート数）
const SpiderFoundationCnt = 8

// SpiderTotalCards 総カード数（2デッキ）
const SpiderTotalCards = 104

// SpiderTableauCard タブロー上のカード
type SpiderTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// SpiderHint ヒント
type SpiderHint struct {
	FromCol   int
	CardIndex int
	ToCol     int
}

// Spider スパイダーソリティアゲームクラス
type Spider struct {
	trumpCards     *TrumpCards
	tableau        [SpiderTableauCnt][]*SpiderTableauCard
	stock          []*Card
	completedSuits int
	phase          SpiderPhase
	moveCount      int
	score          int
	actionLogBase
	history     []*spiderSnapshot
	difficulty  SpiderDifficulty
	isStalemate bool
}

// spiderSnapshot アンドゥ用スナップショット
type spiderSnapshot struct {
	tableau        [SpiderTableauCnt][]*SpiderTableauCard
	stock          []*Card
	completedSuits int
	phase          SpiderPhase
	moveCount      int
	score          int
	isStalemate    bool
}

// NewSpider コンストラクタ
func NewSpider(trumpCards *TrumpCards) *Spider {
	return &Spider{
		trumpCards: trumpCards,
		difficulty: SpiderDifficulty1Suit,
	}
}

// NewDefaultSpider returns Spider with the 1-suit difficulty using the
// standard Spider card count (104 spade cards).
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultSpider() *Spider {
	return NewSpider(NewTrumpCardsWithSuits(SpiderTotalCards, []int{CardDesignSpade}))
}

// ResetWithConfig 設定付きリセット
func (s *Spider) ResetWithConfig(cfg SpiderConfig) {
	switch cfg.Difficulty {
	case SpiderDifficulty2Suit:
		s.difficulty = SpiderDifficulty2Suit
	case SpiderDifficulty4Suit:
		s.difficulty = SpiderDifficulty4Suit
	default:
		s.difficulty = SpiderDifficulty1Suit
	}
	s.rebuildDeck()
	s.Reset()
}

// rebuildDeck 難易度に応じたデッキを再構築
func (s *Spider) rebuildDeck() {
	var suits []int
	switch s.difficulty {
	case SpiderDifficulty2Suit:
		suits = []int{CardDesignSpade, CardDesignHeart}
	case SpiderDifficulty4Suit:
		suits = []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	default:
		suits = []int{CardDesignSpade}
	}
	s.trumpCards = NewTrumpCardsWithSuits(SpiderTotalCards, suits)
}

// Reset ゲームリセット
func (s *Spider) Reset() {
	s.trumpCards.Shuffle()
	s.phase = SpiderPhasePlaying
	s.moveCount = 0
	s.score = 500
	s.actionLog = nil
	s.history = nil
	s.isStalemate = false
	s.completedSuits = 0

	// タブローに配る: 最初の4列に6枚（5裏+1表）、残り6列に5枚（4裏+1表）
	for i := range SpiderTableauCnt {
		count := 5
		if i < 4 {
			count = 6
		}
		s.tableau[i] = make([]*SpiderTableauCard, 0, count)
		for j := range count {
			card := s.trumpCards.DrawCard()
			tc := &SpiderTableauCard{
				Card:   card,
				FaceUp: j == count-1,
			}
			s.tableau[i] = append(s.tableau[i], tc)
		}
	}

	// 残りをストックへ (50枚 = 5回 × 10枚ずつ配る)
	s.stock = nil
	for s.trumpCards.GetRemainingCount() > 0 {
		card := s.trumpCards.DrawCard()
		s.stock = append(s.stock, card)
	}
}

// Deal ストックからタブローに1枚ずつ配る
func (s *Spider) Deal() error {
	if s.phase != SpiderPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(s.stock) < SpiderTableauCnt {
		return errors.New("not enough cards in stock")
	}
	// 空列がある場合は配れない
	for i := range SpiderTableauCnt {
		if len(s.tableau[i]) == 0 {
			return errors.New("cannot deal: empty column exists")
		}
	}
	s.takeSnapshot()
	// 10列に1枚ずつ配る
	for i := range SpiderTableauCnt {
		card := s.stock[len(s.stock)-1]
		s.stock = s.stock[:len(s.stock)-1]
		s.tableau[i] = append(s.tableau[i], &SpiderTableauCard{Card: card, FaceUp: true})
	}
	s.moveCount++
	s.score--
	s.appendLog("deal", "ストックから各列にカードを配りました", nil)
	// 配った後に完成スートをチェック
	for i := range SpiderTableauCnt {
		s.checkAndRemoveCompletedSuit(i)
	}
	s.checkSpiderStalemate()
	return nil
}

// MoveTableauToTableau タブロー間でカードを移動
func (s *Spider) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if s.phase != SpiderPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= SpiderTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= SpiderTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := s.tableau[fromCol]
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	tc := fromCards[cardIndex]
	if !tc.FaceUp {
		return errors.New("card is face down")
	}

	// 移動するカード列が同スート降順の連続であること
	movingCards := fromCards[cardIndex:]
	if !s.isValidSpiderSequence(movingCards) {
		return errors.New("cards are not a valid same-suit descending sequence")
	}

	// 移動先に置けるか確認
	bottomCard := movingCards[0].Card
	if !s.canPlaceOnTableau(bottomCard, toCol) {
		return errors.New("cannot place card on tableau")
	}

	// 移動実行
	s.takeSnapshot()
	movedCards := make([]*Card, len(movingCards))
	for i, mc := range movingCards {
		s.tableau[toCol] = append(s.tableau[toCol], mc)
		movedCards[i] = mc.Card
	}
	s.tableau[fromCol] = fromCards[:cardIndex]
	// 自動フリップ
	s.autoFlipTableau(fromCol)
	s.moveCount++
	s.score--
	s.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), movedCards)
	// 完成スートチェック
	s.checkAndRemoveCompletedSuit(toCol)
	s.checkSpiderStalemate()
	return nil
}

// GiveUp ギブアップ
func (s *Spider) GiveUp() {
	if s.phase == SpiderPhasePlaying {
		s.phase = SpiderPhaseGameOver
		s.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (s *Spider) GetHint() *SpiderHint {
	if s.phase != SpiderPhasePlaying {
		return nil
	}

	// 各列の全表向きカード位置から有効なシーケンスを試す
	// 優先度1: 裏カードを開ける移動
	// 優先度2: その他の有効な移動
	for _, exposeOnly := range []bool{true, false} {
		for fromCol := range SpiderTableauCnt {
			fromCards := s.tableau[fromCol]
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

			// 裏カード開け優先のパス: 裏カードがない列はスキップ
			if exposeOnly && firstFaceUp == 0 {
				continue
			}

			// 各表向きカード位置から有効シーケンスを試す
			for startIdx := firstFaceUp; startIdx < len(fromCards); startIdx++ {
				movingCards := fromCards[startIdx:]
				if !s.isValidSpiderSequence(movingCards) {
					continue
				}
				bottomCard := movingCards[0].Card
				for toCol := range SpiderTableauCnt {
					if toCol == fromCol {
						continue
					}
					if !s.canPlaceOnTableau(bottomCard, toCol) {
						continue
					}
					// 空列への移動で列全体を移すのは無意味
					if len(s.tableau[toCol]) == 0 && startIdx == 0 {
						continue
					}
					// 裏カード開けパスでは裏カードを開ける移動のみ
					if exposeOnly && startIdx != firstFaceUp {
						continue
					}
					return &SpiderHint{
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
func (s *Spider) AutoComplete() error {
	if s.phase != SpiderPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if !s.AllFaceUp() {
		return errors.New("not all cards are face up")
	}
	s.takeSnapshot()
	for {
		removed := false
		for col := range SpiderTableauCnt {
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
func (s *Spider) AllFaceUp() bool {
	if len(s.stock) > 0 {
		return false
	}
	for col := range SpiderTableauCnt {
		for _, tc := range s.tableau[col] {
			if !tc.FaceUp {
				return false
			}
		}
	}
	return true
}

// Undo 直前の操作を取り消す
func (s *Spider) Undo() error {
	if s.phase != SpiderPhasePlaying {
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
func (s *Spider) CanUndo() bool {
	return len(s.history) > 0 && s.phase == SpiderPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (s *Spider) UndoToEscape() int {
	return undoToEscape(s.isStalemate, s.history, func(s *spiderSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (s *Spider) UndoN(n int) error {
	return undoN(s, n)
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (s *Spider) GetPhase() SpiderPhase { return s.phase }

// SetPhase フェーズ設定 (テスト用)
func (s *Spider) SetPhase(phase SpiderPhase) { s.phase = phase }

// GetMoveCount 移動回数取得
func (s *Spider) GetMoveCount() int { return s.moveCount }

// GetStockCount ストック枚数取得
func (s *Spider) GetStockCount() int { return len(s.stock) }

// GetTableau タブロー取得
func (s *Spider) GetTableau() [SpiderTableauCnt][]*SpiderTableauCard { return s.tableau }

// GetCompletedSuits 完成スート数取得
func (s *Spider) GetCompletedSuits() int { return s.completedSuits }

// GetGameEndFlag returns true once the game has left the playing phase.
func (s *Spider) GetGameEndFlag() bool { return s.phase != SpiderPhasePlaying }

// IsStalemate 手詰まり状態取得
func (s *Spider) IsStalemate() bool { return s.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (s *Spider) SetIsStalemate(v bool) { s.isStalemate = v }

// GetScore スコア取得
func (s *Spider) GetScore() int { return s.score }

// GetDifficulty 難易度取得
func (s *Spider) GetDifficulty() SpiderDifficulty { return s.difficulty }

// SetTableau タブロー設定 (テスト用)
func (s *Spider) SetTableau(tableau [SpiderTableauCnt][]*SpiderTableauCard) {
	s.tableau = tableau
}

// SetStock ストック設定 (テスト用)
func (s *Spider) SetStock(stock []*Card) { s.stock = stock }

// SetCompletedSuits 完成スート数設定 (テスト用)
func (s *Spider) SetCompletedSuits(n int) { s.completedSuits = n }

// SetScore スコア設定 (テスト用)
func (s *Spider) SetScore(score int) { s.score = score }

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
func (s *Spider) canPlaceOnTableau(card *Card, col int) bool {
	colCards := s.tableau[col]
	if len(colCards) == 0 {
		// 空の列にはどのカードでも置ける
		return true
	}
	topCard := colCards[len(colCards)-1].Card
	// 値が1つ大きいカードの上に置ける（スート不問）
	return card.GetValue() == topCard.GetValue()-1
}

// isValidSpiderSequence 同スート降順の連続かどうか判定
func (s *Spider) isValidSpiderSequence(cards []*SpiderTableauCard) bool {
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

// checkAndRemoveCompletedSuit K-Aの同スート完成シーケンスをチェックし除去する
func (s *Spider) checkAndRemoveCompletedSuit(col int) bool {
	cards := s.tableau[col]
	if len(cards) < CardValueMax {
		return false
	}
	// 末尾13枚が同スートK-Aか確認
	startIdx := len(cards) - CardValueMax
	seq := cards[startIdx:]

	// 最初のカードはK (13) でなければならない
	if seq[0].Card.GetValue() != CardValueMax {
		return false
	}
	// 最後のカードはA (1) でなければならない
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

	// 完成スートを除去
	s.tableau[col] = cards[:startIdx]
	s.completedSuits++
	s.score += 100
	s.appendLog("complete", fmt.Sprintf("タブロー列%dでスートが完成しました", col), nil)

	// 自動フリップ
	s.autoFlipTableau(col)

	// ゲームクリア判定
	s.checkGameClear()

	return true
}

// autoFlipTableau タブローの最上部の裏カードを自動フリップ
func (s *Spider) autoFlipTableau(col int) {
	cards := s.tableau[col]
	if len(cards) > 0 && !cards[len(cards)-1].FaceUp {
		cards[len(cards)-1].FaceUp = true
	}
}

// checkGameClear ゲームクリア判定
func (s *Spider) checkGameClear() {
	if s.completedSuits >= SpiderFoundationCnt {
		s.phase = SpiderPhaseGameClear
	}
}

// checkSpiderStalemate 手詰まり判定
func (s *Spider) checkSpiderStalemate() {
	if s.phase != SpiderPhasePlaying {
		return
	}
	hint := s.GetHint()
	if hint != nil {
		s.isStalemate = false
		return
	}
	// ストックから配れるか（空列がなく、ストックがある）
	if len(s.stock) > 0 {
		hasEmpty := false
		for i := range SpiderTableauCnt {
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
func (s *Spider) takeSnapshot() {
	snap := &spiderSnapshot{
		completedSuits: s.completedSuits,
		phase:          s.phase,
		moveCount:      s.moveCount,
		score:          s.score,
		isStalemate:    s.isStalemate,
	}
	// deep copy tableau
	for i := range SpiderTableauCnt {
		snap.tableau[i] = make([]*SpiderTableauCard, len(s.tableau[i]))
		for j, tc := range s.tableau[i] {
			snap.tableau[i][j] = &SpiderTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	// deep copy stock
	snap.stock = make([]*Card, len(s.stock))
	copy(snap.stock, s.stock)
	s.history = appendSnapshot(s.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (s *Spider) restoreSnapshot(snap *spiderSnapshot) {
	s.tableau = snap.tableau
	s.stock = snap.stock
	s.completedSuits = snap.completedSuits
	s.phase = snap.phase
	s.moveCount = snap.moveCount
	s.score = snap.score
	s.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (s *Spider) appendLog(actionType, detail string, cards []*Card) {
	s.appendLogAt(s.moveCount, 0, actionType, detail, cards)
}

// spiderJSON is the JSON wire format for Spider.
type spiderJSON struct {
	TrumpCards     *TrumpCards                            `json:"tc"`
	Tableau        [SpiderTableauCnt][]*SpiderTableauCard `json:"tb"`
	Stock          []*Card                                `json:"st"`
	CompletedSuits int                                    `json:"cs"`
	Phase          SpiderPhase                            `json:"ps"`
	MoveCount      int                                    `json:"mc"`
	Score          int                                    `json:"sc"`
	ActionLog      []*ActionLogEntry                      `json:"al"`
	Difficulty     SpiderDifficulty                       `json:"df"`
	IsStalemate    bool                                   `json:"sm"`
	History        []*spiderSnapshot                      `json:"hi,omitempty"`
}

// spiderSnapshotJSON is the wire format for a single undo snapshot.
// spiderSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// spiderJSON's short keys to keep the KV payload compact (#1654).
type spiderSnapshotJSON struct {
	Tableau        [SpiderTableauCnt][]*SpiderTableauCard `json:"tb"`
	Stock          []*Card                                `json:"st"`
	CompletedSuits int                                    `json:"cs"`
	Phase          SpiderPhase                            `json:"ps"`
	MoveCount      int                                    `json:"mc"`
	Score          int                                    `json:"sc"`
	IsStalemate    bool                                   `json:"sm"`
}

// MarshalJSON implements json.Marshaler for spiderSnapshot, projecting
// the unexported fields onto an exported wire shape so that
// Spider.MarshalJSON can persist the undo history (#1654).
func (s *spiderSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(spiderSnapshotJSON{
		Tableau:        s.tableau,
		Stock:          s.stock,
		CompletedSuits: s.completedSuits,
		Phase:          s.phase,
		MoveCount:      s.moveCount,
		Score:          s.score,
		IsStalemate:    s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for spiderSnapshot.
func (s *spiderSnapshot) UnmarshalJSON(data []byte) error {
	var j spiderSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > spiderMaxSliceLen {
		return fmt.Errorf("spider: snapshot array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > spiderMaxSliceLen {
			return fmt.Errorf("spider: snapshot tableau column exceeds maximum allowed size")
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
	return nil
}

// MarshalJSON implements json.Marshaler.
func (s *Spider) MarshalJSON() ([]byte, error) {
	return json.Marshal(spiderJSON{
		TrumpCards:     s.trumpCards,
		Tableau:        s.tableau,
		Stock:          s.stock,
		CompletedSuits: s.completedSuits,
		Phase:          s.phase,
		MoveCount:      s.moveCount,
		Score:          s.score,
		ActionLog:      s.actionLog,
		Difficulty:     s.difficulty,
		IsStalemate:    s.isStalemate,
		History:        s.history,
	})
}

// spiderMaxSliceLen caps slice sizes during deserialisation.
const spiderMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (s *Spider) UnmarshalJSON(data []byte) error {
	var j spiderJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > spiderMaxSliceLen || len(j.ActionLog) > spiderMaxSliceLen ||
		len(j.History) > spiderMaxSliceLen {
		return fmt.Errorf("spider: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > spiderMaxSliceLen {
			return fmt.Errorf("spider: tableau column exceeds maximum allowed size")
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
		s.history = make([]*spiderSnapshot, 0)
	}
	s.difficulty = j.Difficulty
	if s.difficulty == 0 {
		s.difficulty = SpiderDifficulty1Suit
	}
	s.isStalemate = j.IsStalemate
	return nil
}
