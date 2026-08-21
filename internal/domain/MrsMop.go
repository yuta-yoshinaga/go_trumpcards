//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// MrsMopPhase ミセス・モップソリティアゲームフェーズ
type MrsMopPhase int

// MrsMopのフェーズ定数
const (
	// MrsMopPhasePlaying プレイ中
	MrsMopPhasePlaying MrsMopPhase = iota
	// MrsMopPhaseGameClear ゲームクリア
	MrsMopPhaseGameClear
	// MrsMopPhaseGameOver ゲームオーバー
	MrsMopPhaseGameOver
)

// MrsMopTableauCnt タブローの列数。
//
// **13 列。**104 枚を 13 x 8 で配り切るので、クローン元の Spider (10列 + 山札)
// とは配りが違う。
const MrsMopTableauCnt = 13

// MrsMopColSize 1 列に配る枚数 (13 x 8 = 104)。
const MrsMopColSize = 8

// MrsMopFoundationCnt ファンデーションの数（完成スート数）
const MrsMopFoundationCnt = 8

// MrsMopTotalCards 総カード数（2デッキ）
const MrsMopTotalCards = 104

// MrsMopTableauCard タブロー上のカード
type MrsMopTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// MrsMopHint ヒント
type MrsMopHint struct {
	FromCol   int
	CardIndex int
	ToCol     int
}

// MrsMop ミセス・モップソリティアゲームクラス
type MrsMop struct {
	trumpCards     *TrumpCards
	tableau        [MrsMopTableauCnt][]*MrsMopTableauCard
	completedSuits int
	phase          MrsMopPhase
	moveCount      int
	score          int
	actionLogBase
	history     []*mrsMopSnapshot
	difficulty  MrsMopDifficulty
	isStalemate bool
}

// mrsMopSnapshot アンドゥ用スナップショット
type mrsMopSnapshot struct {
	tableau        [MrsMopTableauCnt][]*MrsMopTableauCard
	completedSuits int
	phase          MrsMopPhase
	moveCount      int
	score          int
	isStalemate    bool
}

// NewMrsMop コンストラクタ
func NewMrsMop(trumpCards *TrumpCards) *MrsMop {
	return &MrsMop{
		trumpCards: trumpCards,
		difficulty: MrsMopDifficulty4Suit,
	}
}

// NewDefaultMrsMop returns MrsMop at its proper 4-suit form: two full decks,
// 104 cards. **Not the clone source's 1-suit default** — a one-suit board is a
// relaxed variant, not Mrs. Mop.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultMrsMop() *MrsMop {
	return NewMrsMop(NewTrumpCardsWithSuits(MrsMopTotalCards,
		[]int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}))
}

// ResetWithConfig 設定付きリセット
func (s *MrsMop) ResetWithConfig(cfg MrsMopConfig) {
	// **既定 (未知の値を含む) は4スート = 本来の Mrs. Mop。**
	// クローン元の Spider は 1 スートに落とすが、それを引き継ぐと未知の値が
	// 「別のゲーム」に化ける。
	switch cfg.Difficulty {
	case MrsMopDifficulty1Suit:
		s.difficulty = MrsMopDifficulty1Suit
	case MrsMopDifficulty2Suit:
		s.difficulty = MrsMopDifficulty2Suit
	default:
		s.difficulty = MrsMopDifficulty4Suit
	}
	s.rebuildDeck()
	s.Reset()
}

// rebuildDeck 難易度に応じたデッキを再構築
func (s *MrsMop) rebuildDeck() {
	var suits []int
	switch s.difficulty {
	case MrsMopDifficulty1Suit:
		suits = []int{CardDesignSpade}
	case MrsMopDifficulty2Suit:
		suits = []int{CardDesignSpade, CardDesignHeart}
	default:
		suits = []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	}
	s.trumpCards = NewTrumpCardsWithSuits(MrsMopTotalCards, suits)
}

// Reset ゲームリセット
//
// **104 枚を 13 列 x 8 枚に配り切り、全部表向きにする。**伏せ札も山札も無いので、
// 開始時点で情報が全部見えている ── 運ではなく読みだけのゲームになる。
// クローン元の Spider は伏せ札を作り 50 枚を山札に残す。
func (s *MrsMop) Reset() {
	s.trumpCards.Shuffle()
	s.phase = MrsMopPhasePlaying
	s.moveCount = 0
	s.score = 500
	s.actionLog = nil
	s.history = nil
	s.isStalemate = false
	s.completedSuits = 0

	for i := range MrsMopTableauCnt {
		s.tableau[i] = make([]*MrsMopTableauCard, 0, MrsMopColSize)
		for range MrsMopColSize {
			card := s.trumpCards.DrawCard()
			if card == nil {
				break
			}
			s.tableau[i] = append(s.tableau[i], &MrsMopTableauCard{Card: card, FaceUp: true})
		}
	}
}

// MoveTableauToTableau タブロー間でカードを移動
func (s *MrsMop) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if s.phase != MrsMopPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= MrsMopTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= MrsMopTableauCnt {
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
	if !s.isValidMrsMopSequence(movingCards) {
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
	s.checkMrsMopStalemate()
	return nil
}

// GiveUp ギブアップ
func (s *MrsMop) GiveUp() {
	if s.phase == MrsMopPhasePlaying {
		s.phase = MrsMopPhaseGameOver
		s.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (s *MrsMop) GetHint() *MrsMopHint {
	if s.phase != MrsMopPhasePlaying {
		return nil
	}

	// 各列の全表向きカード位置から有効なシーケンスを試す
	// 優先度1: 裏カードを開ける移動
	// 優先度2: その他の有効な移動
	for _, exposeOnly := range []bool{true, false} {
		for fromCol := range MrsMopTableauCnt {
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
				if !s.isValidMrsMopSequence(movingCards) {
					continue
				}
				bottomCard := movingCards[0].Card
				for toCol := range MrsMopTableauCnt {
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
					return &MrsMopHint{
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
func (s *MrsMop) AutoComplete() error {
	if s.phase != MrsMopPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if !s.AllFaceUp() {
		return errors.New("not all cards are face up")
	}
	s.takeSnapshot()
	for {
		removed := false
		for col := range MrsMopTableauCnt {
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

// AllFaceUp 全カードが表向きかどうか。
//
// **Mrs. Mop では常に true。**伏せ札も山札も無い。判定を残すのは、KV から
// 壊れた state を復元したときに嘘をつかないため。
func (s *MrsMop) AllFaceUp() bool {
	for col := range MrsMopTableauCnt {
		for _, tc := range s.tableau[col] {
			if !tc.FaceUp {
				return false
			}
		}
	}
	return true
}

// Undo 直前の操作を取り消す
func (s *MrsMop) Undo() error {
	if s.phase != MrsMopPhasePlaying {
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
func (s *MrsMop) CanUndo() bool {
	return len(s.history) > 0 && s.phase == MrsMopPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (s *MrsMop) UndoToEscape() int {
	return undoToEscape(s.isStalemate, s.history, func(s *mrsMopSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (s *MrsMop) UndoN(n int) error {
	return undoN(s, n)
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (s *MrsMop) GetPhase() MrsMopPhase { return s.phase }

// SetPhase フェーズ設定 (テスト用)
func (s *MrsMop) SetPhase(phase MrsMopPhase) { s.phase = phase }

// GetMoveCount 移動回数取得
func (s *MrsMop) GetMoveCount() int { return s.moveCount }

// GetStockCount ストック枚数取得。**Mrs. Mop に山札は無いので常に 0。**
// 共有の presenter/interface が要求するので残している。
func (s *MrsMop) GetStockCount() int { return 0 }

// GetTableau タブロー取得
func (s *MrsMop) GetTableau() [MrsMopTableauCnt][]*MrsMopTableauCard { return s.tableau }

// GetCompletedSuits 完成スート数取得
func (s *MrsMop) GetCompletedSuits() int { return s.completedSuits }

// GetGameEndFlag returns true once the game has left the playing phase.
func (s *MrsMop) GetGameEndFlag() bool { return s.phase != MrsMopPhasePlaying }

// IsStalemate 手詰まり状態取得
func (s *MrsMop) IsStalemate() bool { return s.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (s *MrsMop) SetIsStalemate(v bool) { s.isStalemate = v }

// GetScore スコア取得
func (s *MrsMop) GetScore() int { return s.score }

// GetDifficulty 難易度取得
func (s *MrsMop) GetDifficulty() MrsMopDifficulty { return s.difficulty }

// SetTableau タブロー設定 (テスト用)
func (s *MrsMop) SetTableau(tableau [MrsMopTableauCnt][]*MrsMopTableauCard) {
	s.tableau = tableau
}

// SetStock ストック設定 (テスト用)
func (s *MrsMop) SetStock(_ []*Card) {}

// SetCompletedSuits 完成スート数設定 (テスト用)
func (s *MrsMop) SetCompletedSuits(n int) { s.completedSuits = n }

// SetScore スコア設定 (テスト用)
func (s *MrsMop) SetScore(score int) { s.score = score }

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
func (s *MrsMop) canPlaceOnTableau(card *Card, col int) bool {
	colCards := s.tableau[col]
	if len(colCards) == 0 {
		// 空の列にはどのカードでも置ける
		return true
	}
	topCard := colCards[len(colCards)-1].Card
	// 値が1つ大きいカードの上に置ける（スート不問）
	return card.GetValue() == topCard.GetValue()-1
}

// isValidMrsMopSequence 同スート降順の連続かどうか判定
func (s *MrsMop) isValidMrsMopSequence(cards []*MrsMopTableauCard) bool {
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
func (s *MrsMop) checkAndRemoveCompletedSuit(col int) bool {
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
func (s *MrsMop) autoFlipTableau(col int) {
	cards := s.tableau[col]
	if len(cards) > 0 && !cards[len(cards)-1].FaceUp {
		cards[len(cards)-1].FaceUp = true
	}
}

// checkGameClear ゲームクリア判定
func (s *MrsMop) checkGameClear() {
	if s.completedSuits >= MrsMopFoundationCnt {
		s.phase = MrsMopPhaseGameClear
	}
}

// checkMrsMopStalemate 手詰まり判定
func (s *MrsMop) checkMrsMopStalemate() {
	if s.phase != MrsMopPhasePlaying {
		return
	}
	hint := s.GetHint()
	if hint != nil {
		s.isStalemate = false
		return
	}
	// **山札からの救済は無い。**クローン元の Spider は「空列が無ければ配れる」
	// ので手詰まりを免れるが、Mrs. Mop は最初に配り切る。指せる手が無ければ
	// それが詰み。
	s.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (s *MrsMop) takeSnapshot() {
	snap := &mrsMopSnapshot{
		completedSuits: s.completedSuits,
		phase:          s.phase,
		moveCount:      s.moveCount,
		score:          s.score,
		isStalemate:    s.isStalemate,
	}
	// deep copy tableau
	for i := range MrsMopTableauCnt {
		snap.tableau[i] = make([]*MrsMopTableauCard, len(s.tableau[i]))
		for j, tc := range s.tableau[i] {
			snap.tableau[i][j] = &MrsMopTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	s.history = append(s.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (s *MrsMop) restoreSnapshot(snap *mrsMopSnapshot) {
	s.tableau = snap.tableau
	s.completedSuits = snap.completedSuits
	s.phase = snap.phase
	s.moveCount = snap.moveCount
	s.score = snap.score
	s.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (s *MrsMop) appendLog(actionType, detail string, cards []*Card) {
	s.appendLogAt(s.moveCount, 0, actionType, detail, cards)
}

// mrsMopJSON is the JSON wire format for MrsMop.
type mrsMopJSON struct {
	TrumpCards     *TrumpCards                            `json:"tc"`
	Tableau        [MrsMopTableauCnt][]*MrsMopTableauCard `json:"tb"`
	CompletedSuits int                                    `json:"cs"`
	Phase          MrsMopPhase                            `json:"ps"`
	MoveCount      int                                    `json:"mc"`
	Score          int                                    `json:"sc"`
	ActionLog      []*ActionLogEntry                      `json:"al"`
	Difficulty     MrsMopDifficulty                       `json:"df"`
	IsStalemate    bool                                   `json:"sm"`
	History        []*mrsMopSnapshot                      `json:"hi,omitempty"`
}

// mrsMopSnapshotJSON is the wire format for a single undo snapshot.
// mrsMopSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// mrsMopJSON's short keys to keep the KV payload compact (#1654).
type mrsMopSnapshotJSON struct {
	Tableau        [MrsMopTableauCnt][]*MrsMopTableauCard `json:"tb"`
	CompletedSuits int                                    `json:"cs"`
	Phase          MrsMopPhase                            `json:"ps"`
	MoveCount      int                                    `json:"mc"`
	Score          int                                    `json:"sc"`
	IsStalemate    bool                                   `json:"sm"`
}

// MarshalJSON implements json.Marshaler for mrsMopSnapshot, projecting
// the unexported fields onto an exported wire shape so that
// MrsMop.MarshalJSON can persist the undo history (#1654).
func (s *mrsMopSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(mrsMopSnapshotJSON{
		Tableau:        s.tableau,
		CompletedSuits: s.completedSuits,
		Phase:          s.phase,
		MoveCount:      s.moveCount,
		Score:          s.score,
		IsStalemate:    s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for mrsMopSnapshot.
func (s *mrsMopSnapshot) UnmarshalJSON(data []byte) error {
	var j mrsMopSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > mrsMopMaxSliceLen {
			return fmt.Errorf("mrsMop: snapshot tableau column exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.completedSuits = j.CompletedSuits
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.score = j.Score
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (s *MrsMop) MarshalJSON() ([]byte, error) {
	return json.Marshal(mrsMopJSON{
		TrumpCards:     s.trumpCards,
		Tableau:        s.tableau,
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

// mrsMopMaxSliceLen caps slice sizes during deserialisation.
const mrsMopMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (s *MrsMop) UnmarshalJSON(data []byte) error {
	var j mrsMopJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > mrsMopMaxSliceLen || len(j.History) > mrsMopMaxSliceLen {
		return fmt.Errorf("mrsMop: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > mrsMopMaxSliceLen {
			return fmt.Errorf("mrsMop: tableau column exceeds maximum allowed size")
		}
	}

	s.trumpCards = j.TrumpCards
	if s.trumpCards == nil {
		s.trumpCards = NewTrumpCards(0)
	}
	s.tableau = j.Tableau
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
		s.history = make([]*mrsMopSnapshot, 0)
	}
	s.difficulty = j.Difficulty
	if s.difficulty == 0 {
		s.difficulty = MrsMopDifficulty1Suit
	}
	s.isStalemate = j.IsStalemate
	return nil
}
