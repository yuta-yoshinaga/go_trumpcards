//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ScorpionPhase スコーピオンゲームフェーズ
type ScorpionPhase int

// Scorpionのフェーズ定数
const (
	// ScorpionPhasePlaying プレイ中
	ScorpionPhasePlaying ScorpionPhase = iota
	// ScorpionPhaseGameClear ゲームクリア
	ScorpionPhaseGameClear
	// ScorpionPhaseGameOver ゲームオーバー
	ScorpionPhaseGameOver
)

// ScorpionTableauCnt タブローの列数
const ScorpionTableauCnt = 7

// ScorpionFaceDownCols 先頭に裏向きカードが配られる列数（列0〜3）
const ScorpionFaceDownCols = 4

// ScorpionColSize 各列の初期カード枚数
const ScorpionColSize = 7

// ScorpionFaceDownPerCol 先頭の列で裏向きに配られる枚数
const ScorpionFaceDownPerCol = 3

// ScorpionStockSize ストックに残るカード枚数
const ScorpionStockSize = 3

// ScorpionCompletedSuitsCnt ゲームクリアに必要な完成スート数
const ScorpionCompletedSuitsCnt = 4

// ScorpionHintDeal はストックから配るのが最善手であることを示すセンチネル値。
// ScorpionHint のフィールドすべてがこの値のとき、通常のカード移動ではなく Deal コマンドを意味する。
const ScorpionHintDeal = -1

// ScorpionHint ヒント。FromCol/CardIndex/ToCol が全て ScorpionHintDeal の場合は「ストックから配る」を示す。
type ScorpionHint struct {
	FromCol   int
	CardIndex int
	ToCol     int
	// ExposesFaceDown はこの手を打つと裏向きカードが1枚表になるかどうか。
	//
	// **ヒントが優先した理由そのもの。**スコーピオンの肝は12枚の裏カードを
	// どれだけ早く開けるかで、GetHint は裏カードを開ける手を先に探している。
	// 移動先だけ伝えても、プレイヤーはなぜその手なのか学べない (#5544)。
	ExposesFaceDown bool
}

// IsDeal は「ストックから配る」ヒントかどうかを返す。
func (h *ScorpionHint) IsDeal() bool {
	return h != nil && h.FromCol == ScorpionHintDeal
}

// Scorpion スコーピオンゲームクラス
type Scorpion struct {
	trumpCards     *TrumpCards
	tableau        [ScorpionTableauCnt][]*KlondikeTableauCard
	stock          []*Card
	completedSuits int
	phase          ScorpionPhase
	moveCount      int
	actionLogBase
	history     []*scorpionSnapshot
	isStalemate bool
}

// scorpionSnapshot アンドゥ用スナップショット
type scorpionSnapshot struct {
	tableau        [ScorpionTableauCnt][]*KlondikeTableauCard
	stock          []*Card
	completedSuits int
	phase          ScorpionPhase
	moveCount      int
	isStalemate    bool
}

// NewScorpion コンストラクタ
func NewScorpion(trumpCards *TrumpCards) *Scorpion {
	return &Scorpion{
		trumpCards: trumpCards,
	}
}

// NewDefaultScorpion returns Scorpion with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultScorpion() *Scorpion {
	return NewScorpion(NewTrumpCards(0))
}

// Reset ゲームリセット
func (s *Scorpion) Reset() {
	s.trumpCards.Shuffle()
	s.phase = ScorpionPhasePlaying
	s.moveCount = 0
	s.actionLog = nil
	s.history = nil
	s.isStalemate = false
	s.completedSuits = 0

	// タブローに配る: 列0-3は先頭3枚裏+残り4枚表、列4-6は7枚全て表
	for i := range ScorpionTableauCnt {
		s.tableau[i] = make([]*KlondikeTableauCard, 0, ScorpionColSize)
		for j := range ScorpionColSize {
			card := s.trumpCards.DrawCard()
			faceDown := i < ScorpionFaceDownCols && j < ScorpionFaceDownPerCol
			s.tableau[i] = append(s.tableau[i], &KlondikeTableauCard{
				Card:   card,
				FaceUp: !faceDown,
			})
		}
	}

	// 残り3枚をストックへ
	s.stock = make([]*Card, 0, ScorpionStockSize)
	for s.trumpCards.GetRemainingCount() > 0 {
		s.stock = append(s.stock, s.trumpCards.DrawCard())
	}

	// 初期局面でも手詰まり状態を評価しておく（他のエントリポイントと同じ不変条件を保つ）
	s.checkScorpionStalemate()
}

// Deal ストックから先頭3列に1枚ずつ配る
func (s *Scorpion) Deal() error {
	if s.phase != ScorpionPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(s.stock) == 0 {
		return errors.New("stock is empty")
	}
	// 空列がある場合は配れない（標準ルール）
	for i := range ScorpionTableauCnt {
		if len(s.tableau[i]) == 0 {
			return errors.New("cannot deal: empty column exists")
		}
	}
	s.takeSnapshot()
	// ストックのカードを列0から順に1枚ずつ表向きで配る。
	// 通常は ScorpionStockSize (3) 枚で ScorpionTableauCnt 以下だが、将来的な安全のため min を取る。
	dealCount := min(len(s.stock), ScorpionTableauCnt)
	dealt := make([]*Card, 0, dealCount)
	for i := range dealCount {
		card := s.stock[i]
		s.tableau[i] = append(s.tableau[i], &KlondikeTableauCard{Card: card, FaceUp: true})
		dealt = append(dealt, card)
	}
	s.stock = s.stock[dealCount:]
	s.moveCount++
	s.appendLog("deal", "ストックから列0-2に1枚ずつ配りました", dealt)
	// 配った後に完成スートをチェック
	for i := range ScorpionTableauCnt {
		s.checkAndRemoveCompletedSuit(i)
	}
	s.checkScorpionStalemate()
	return nil
}

// MoveTableauToTableau タブロー間でカードを移動
func (s *Scorpion) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if s.phase != ScorpionPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= ScorpionTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= ScorpionTableauCnt {
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
	// Scorpion: 移動先のルールのみチェック（移動するカード群の整列は不要、Yukonと同じ）
	bottomCard := tc.Card
	if !s.canPlaceOnTableau(bottomCard, toCol) {
		return errors.New("cannot place card on tableau")
	}
	// 移動実行
	s.takeSnapshot()
	movingCards := fromCards[cardIndex:]
	movedCards := make([]*Card, len(movingCards))
	for i, mc := range movingCards {
		s.tableau[toCol] = append(s.tableau[toCol], mc)
		movedCards[i] = mc.Card
	}
	s.tableau[fromCol] = fromCards[:cardIndex]
	// 自動フリップ
	s.autoFlipTableau(fromCol)
	s.moveCount++
	s.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), movedCards)
	// 完成スートチェック（移動先の列で末尾がK-Aの同スート完成になっているか）
	s.checkAndRemoveCompletedSuit(toCol)
	s.checkScorpionStalemate()
	return nil
}

// GiveUp ギブアップ
func (s *Scorpion) GiveUp() {
	if s.phase == ScorpionPhasePlaying {
		s.phase = ScorpionPhaseGameOver
		s.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (s *Scorpion) GetHint() *ScorpionHint {
	if s.phase != ScorpionPhasePlaying {
		return nil
	}
	// 優先度1: 裏カードを開ける移動
	// 優先度2: その他の有効な移動
	for _, exposeOnly := range []bool{true, false} {
		for fromCol := range ScorpionTableauCnt {
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
			for startIdx := firstFaceUp; startIdx < len(fromCards); startIdx++ {
				tc := fromCards[startIdx]
				if !tc.FaceUp {
					continue
				}
				// 裏カード開けパスでは、開けるカードの直上からの移動のみ
				if exposeOnly && startIdx != firstFaceUp {
					continue
				}
				card := tc.Card
				for toCol := range ScorpionTableauCnt {
					if toCol == fromCol {
						continue
					}
					if !s.canPlaceOnTableau(card, toCol) {
						continue
					}
					// 空列への全体移動は意味がないのでスキップ
					if len(s.tableau[toCol]) == 0 && startIdx == 0 {
						continue
					}
					return &ScorpionHint{
						FromCol:   fromCol,
						CardIndex: startIdx,
						ToCol:     toCol,
						// 動かす札のすぐ下に裏カードが残るなら、この手で1枚開く。
						ExposesFaceDown: startIdx > 0 && !fromCards[startIdx-1].FaceUp,
					}
				}
			}
		}
	}
	// ストックが残っていて空列がなければ、deal自体もヒントとして有効
	if len(s.stock) > 0 {
		hasEmpty := false
		for i := range ScorpionTableauCnt {
			if len(s.tableau[i]) == 0 {
				hasEmpty = true
				break
			}
		}
		if !hasEmpty {
			return &ScorpionHint{FromCol: ScorpionHintDeal, CardIndex: ScorpionHintDeal, ToCol: ScorpionHintDeal}
		}
	}
	return nil
}

// AutoComplete オートコンプリート（全表向きのとき、完成スートを自動除去）
func (s *Scorpion) AutoComplete() error {
	if s.phase != ScorpionPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if !s.AllFaceUp() {
		return errors.New("not all cards are face up")
	}
	s.takeSnapshot()
	for {
		removed := false
		for col := range ScorpionTableauCnt {
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
	// 除去によって新しい手が現れた可能性があるため、手詰まり状態を再評価する
	s.checkScorpionStalemate()
	return nil
}

// AllFaceUp 全カードが表向きかつストックが空かどうか
func (s *Scorpion) AllFaceUp() bool {
	if len(s.stock) > 0 {
		return false
	}
	for col := range ScorpionTableauCnt {
		for _, tc := range s.tableau[col] {
			if !tc.FaceUp {
				return false
			}
		}
	}
	return true
}

// Undo 直前の操作を取り消す
func (s *Scorpion) Undo() error {
	if s.phase != ScorpionPhasePlaying {
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
func (s *Scorpion) CanUndo() bool {
	return len(s.history) > 0 && s.phase == ScorpionPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (s *Scorpion) UndoToEscape() int {
	return undoToEscape(s.isStalemate, s.history, func(s *scorpionSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (s *Scorpion) UndoN(n int) error {
	return undoN(s, n)
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (s *Scorpion) GetPhase() ScorpionPhase { return s.phase }

// SetPhase フェーズ設定 (テスト用)
func (s *Scorpion) SetPhase(phase ScorpionPhase) { s.phase = phase }

// GetMoveCount 移動回数取得
func (s *Scorpion) GetMoveCount() int { return s.moveCount }

// GetStockCount ストック残り枚数取得
func (s *Scorpion) GetStockCount() int { return len(s.stock) }

// GetTableau タブロー取得
func (s *Scorpion) GetTableau() [ScorpionTableauCnt][]*KlondikeTableauCard { return s.tableau }

// GetCompletedSuits 完成スート数取得
func (s *Scorpion) GetCompletedSuits() int { return s.completedSuits }

// GetGameEndFlag returns true once the game has left the playing phase.
func (s *Scorpion) GetGameEndFlag() bool { return s.phase != ScorpionPhasePlaying }

// IsStalemate 手詰まり状態取得
func (s *Scorpion) IsStalemate() bool { return s.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (s *Scorpion) SetIsStalemate(v bool) { s.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (s *Scorpion) SetTableau(tableau [ScorpionTableauCnt][]*KlondikeTableauCard) {
	s.tableau = tableau
}

// SetStock ストック設定 (テスト用)
func (s *Scorpion) SetStock(stock []*Card) { s.stock = stock }

// SetCompletedSuits 完成スート数設定 (テスト用)
func (s *Scorpion) SetCompletedSuits(n int) { s.completedSuits = n }

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
func (s *Scorpion) canPlaceOnTableau(card *Card, col int) bool {
	colCards := s.tableau[col]
	if len(colCards) == 0 {
		// 空の列にはKのみ置ける
		return card.GetValue() == CardValueMax
	}
	topCard := colCards[len(colCards)-1].Card
	// 同じスートで降順
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()-1
}

// checkAndRemoveCompletedSuit K-Aの同スート完成シーケンスをチェックし除去する
func (s *Scorpion) checkAndRemoveCompletedSuit(col int) bool {
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
	s.appendLog("complete", fmt.Sprintf("タブロー列%dでスートが完成しました", col), nil)
	s.autoFlipTableau(col)
	s.checkGameClear()
	return true
}

// autoFlipTableau タブローの最上部の裏カードを自動フリップ
func (s *Scorpion) autoFlipTableau(col int) {
	autoFlipTopCard(s.tableau[col])
}

// checkGameClear ゲームクリア判定
func (s *Scorpion) checkGameClear() {
	if s.completedSuits >= ScorpionCompletedSuitsCnt {
		s.phase = ScorpionPhaseGameClear
	}
}

// checkScorpionStalemate 手詰まり判定
func (s *Scorpion) checkScorpionStalemate() {
	if s.phase != ScorpionPhasePlaying {
		return
	}
	if s.GetHint() != nil {
		s.isStalemate = false
		return
	}
	s.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (s *Scorpion) takeSnapshot() {
	snap := &scorpionSnapshot{
		completedSuits: s.completedSuits,
		phase:          s.phase,
		moveCount:      s.moveCount,
		isStalemate:    s.isStalemate,
	}
	for i := range ScorpionTableauCnt {
		snap.tableau[i] = make([]*KlondikeTableauCard, len(s.tableau[i]))
		for j, tc := range s.tableau[i] {
			snap.tableau[i][j] = &KlondikeTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	snap.stock = make([]*Card, len(s.stock))
	copy(snap.stock, s.stock)
	s.history = append(s.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (s *Scorpion) restoreSnapshot(snap *scorpionSnapshot) {
	s.tableau = snap.tableau
	s.stock = snap.stock
	s.completedSuits = snap.completedSuits
	s.phase = snap.phase
	s.moveCount = snap.moveCount
	s.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (s *Scorpion) appendLog(actionType, detail string, cards []*Card) {
	s.appendLogAt(s.moveCount, 0, actionType, detail, cards)
}

// scorpionJSON is the JSON wire format for Scorpion.
type scorpionJSON struct {
	TrumpCards     *TrumpCards                                `json:"tc"`
	Tableau        [ScorpionTableauCnt][]*KlondikeTableauCard `json:"tb"`
	Stock          []*Card                                    `json:"st"`
	CompletedSuits int                                        `json:"cs"`
	Phase          ScorpionPhase                              `json:"ps"`
	MoveCount      int                                        `json:"mc"`
	ActionLog      []*ActionLogEntry                          `json:"al"`
	IsStalemate    bool                                       `json:"sl"`
	History        []*scorpionSnapshot                        `json:"hi,omitempty"`
}

// scorpionSnapshotJSON is the wire format for a single undo snapshot.
// scorpionSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// scorpionJSON's short keys to keep the KV payload compact (#1654).
type scorpionSnapshotJSON struct {
	Tableau        [ScorpionTableauCnt][]*KlondikeTableauCard `json:"tb"`
	Stock          []*Card                                    `json:"st"`
	CompletedSuits int                                        `json:"cs"`
	Phase          ScorpionPhase                              `json:"ps"`
	MoveCount      int                                        `json:"mc"`
	IsStalemate    bool                                       `json:"sl"`
}

// MarshalJSON implements json.Marshaler for scorpionSnapshot.
func (s *scorpionSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(scorpionSnapshotJSON{
		Tableau:        s.tableau,
		Stock:          s.stock,
		CompletedSuits: s.completedSuits,
		Phase:          s.phase,
		MoveCount:      s.moveCount,
		IsStalemate:    s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for scorpionSnapshot.
func (s *scorpionSnapshot) UnmarshalJSON(data []byte) error {
	var j scorpionSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > scorpionMaxSliceLen {
		return fmt.Errorf("scorpion: snapshot stock exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > scorpionMaxSliceLen {
			return fmt.Errorf("scorpion: snapshot tableau column exceeds maximum allowed size")
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
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (s *Scorpion) MarshalJSON() ([]byte, error) {
	return json.Marshal(scorpionJSON{
		TrumpCards:     s.trumpCards,
		Tableau:        s.tableau,
		Stock:          s.stock,
		CompletedSuits: s.completedSuits,
		Phase:          s.phase,
		MoveCount:      s.moveCount,
		ActionLog:      s.actionLog,
		IsStalemate:    s.isStalemate,
		History:        s.history,
	})
}

// scorpionMaxSliceLen caps slice sizes during deserialisation.
const scorpionMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (s *Scorpion) UnmarshalJSON(data []byte) error {
	var j scorpionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > scorpionMaxSliceLen || len(j.ActionLog) > scorpionMaxSliceLen ||
		len(j.History) > scorpionMaxSliceLen {
		return fmt.Errorf("scorpion: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > scorpionMaxSliceLen {
			return fmt.Errorf("scorpion: tableau column exceeds maximum allowed size")
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
	s.actionLog = j.ActionLog
	if s.actionLog == nil {
		s.actionLog = make([]*ActionLogEntry, 0)
	}
	s.history = j.History
	if s.history == nil {
		s.history = make([]*scorpionSnapshot, 0)
	}
	s.isStalemate = j.IsStalemate
	return nil
}
