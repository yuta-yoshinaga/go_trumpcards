//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// CribbageSquaresPhase はクリベッジ・スクエアズのフェーズを表す。
type CribbageSquaresPhase int

// CribbageSquaresのフェーズ定数
const (
	// CribbageSquaresPhasePlaying プレイ中
	CribbageSquaresPhasePlaying CribbageSquaresPhase = iota
	// CribbageSquaresPhaseComplete 完了
	CribbageSquaresPhaseComplete
)

// CribbageSquaresGridSize はグリッドの一辺のサイズ。
const CribbageSquaresGridSize = 4

// CribbageSquaresTotalCells は総セル数 (4x4=16)。
const CribbageSquaresTotalCells = CribbageSquaresGridSize * CribbageSquaresGridSize

// CribbageSquaresLineCnt は採点する手の数（4 行 + 4 列）。
const CribbageSquaresLineCnt = CribbageSquaresGridSize * 2

// CribbageSquaresWinScore はクリア扱いになる合計得点。
//
// 『The Complete Book of Solitaire and Patience Games』が挙げる基準値で、
// 8 手あるので 1 手あたり平均 7.6 点。ペアと 15 を両立させないと届かない。
const CribbageSquaresWinScore = 61

// cribbageSquaresMaxSliceLen はデシリアライズ時のスライスサイズ上限。
const cribbageSquaresMaxSliceLen = 1000

// CribbageSquaresHint は現在のカードを置く最善のセルを表す配置ヒント。
type CribbageSquaresHint struct {
	// Row は推奨するマスの行 (0-3)。
	Row int
	// Col は推奨するマスの列 (0-3)。
	Col int
	// Score はその配置が行と列に生む増分点。
	Score int
	// Synergy はスコアが正（既存カードと噛み合う）かどうか。
	Synergy bool
}

// CribbageSquares はクリベッジ・スクエアズのゲーム状態を表す。
//
// 52 枚 1 組から 16 枚を 4×4 のグリッドへ 1 枚ずつ置き、**置き終えてから 17 枚目
// の「スターター」をめくる**。4 行 + 4 列の計 8 手を、それぞれ「4 枚 + スターター」
// のクリベッジの手として採点し、合計 61 点以上でクリア。
//
// 置いたカードは動かせず、スターターは最後まで分からない。したがって「15 とペアを
// どう重ねるか」を、5 枚目が何になるか分からないまま決めることになる。
//
// issue #5272 の仕様案とは 2 点異なり、いずれも実際の規則に合わせた:
//   - **スターター（17 枚目）がある。** issue は 16 枚だけで採点すると書いているが、
//     クリベッジの手は 5 枚で数えるものであり、既存の
//     [CribbageScoreHand] も `starter` を受け取る形になっている。スターター無しでは
//     フラッシュもノブズも成立せず、15 の組合せも大きく減る
//   - **クリア基準（61 点）を持つ。** issue は「目標スコアを提示する」とだけ書いて
//     値を決めていない。基準が無いと勝敗の無いゲームになる
//
// なお本作は「Cribbage Squares」であって、同じく1人用の「Cribbage Solitaire」
// （手札から捨ててクリブを作る対戦形式の 1 人遊び）とは別のゲームである。
// issue のタイトルは後者の名前だが、本文が説明しているのは前者なので前者を実装した。
//
// 置き場所に制限は無く、空いているマスならどこでも良い。既に置いたカードと接して
// いなければならない変種もあるが、それは別のゲームとして扱われており、本リポジトリ
// の [PokerSquares] も制限なしを採っている。
type CribbageSquares struct {
	trumpCards  *TrumpCards
	board       [CribbageSquaresGridSize][CribbageSquaresGridSize]*Card
	currentCard *Card
	// starter は 16 枚を置き終えた時点でめくられる 17 枚目。それまでは nil。
	starter     *Card
	placedCount int
	phase       CribbageSquaresPhase
	actionLog   []*ActionLogEntry
	history     []*cribbageSquaresSnapshot
}

// cribbageSquaresSnapshot はアンドゥ用の状態スナップショット。
type cribbageSquaresSnapshot struct {
	board       [CribbageSquaresGridSize][CribbageSquaresGridSize]*Card
	currentCard *Card
	starter     *Card
	placedCount int
	phase       CribbageSquaresPhase
	deckDrawCnt int
	actionLogLn int
}

// NewCribbageSquares はコンストラクタ。
func NewCribbageSquares(tc *TrumpCards) *CribbageSquares {
	return &CribbageSquares{trumpCards: tc}
}

// NewDefaultCribbageSquares returns CribbageSquares with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultCribbageSquares() *CribbageSquares {
	return NewCribbageSquares(NewTrumpCards(0))
}

// Reset はゲームを初期化する。デッキをシャッフルし、最初のカードを引く。
func (c *CribbageSquares) Reset() {
	c.trumpCards.Shuffle()
	c.board = [CribbageSquaresGridSize][CribbageSquaresGridSize]*Card{}
	c.placedCount = 0
	c.starter = nil
	c.phase = CribbageSquaresPhasePlaying
	c.actionLog = nil
	c.history = nil
	c.currentCard = c.trumpCards.DrawCard()
}

// Place は現在のカードを指定セルに置く。
//
// 16 枚目を置いた時点でスターターをめくり、ゲームを終了する。
func (c *CribbageSquares) Place(row, col int) error {
	if c.phase != CribbageSquaresPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if !validCribbageSquaresCell(row, col) {
		return errors.New("invalid cell position")
	}
	if c.board[row][col] != nil {
		return errors.New("cell is already occupied")
	}
	if c.currentCard == nil {
		return errors.New("no current card to place")
	}
	c.takeSnapshot()
	placed := c.currentCard
	c.board[row][col] = placed
	c.placedCount++
	c.appendLog("place", fmt.Sprintf("(%d,%d) に配置", row, col), []*Card{placed})
	if c.placedCount >= CribbageSquaresTotalCells {
		c.currentCard = nil
		// スターターは最後にめくる。ここまで手札の 5 枚目は誰にも分からない。
		c.starter = c.trumpCards.DrawCard()
		c.phase = CribbageSquaresPhaseComplete
		c.appendLog("starter", "スターターをめくった", []*Card{c.starter})
	} else {
		c.currentCard = c.trumpCards.DrawCard()
	}
	return nil
}

// Undo は直前の配置を取り消す。
func (c *CribbageSquares) Undo() error {
	if len(c.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := c.history[len(c.history)-1]
	c.history = c.history[:len(c.history)-1]
	for i := snap.deckDrawCnt; i < c.trumpCards.deckDrawCnt; i++ {
		c.trumpCards.deck[i].SetDraw(false)
	}
	c.trumpCards.deckDrawCnt = snap.deckDrawCnt
	c.board = snap.board
	c.currentCard = snap.currentCard
	// 最後の 1 手を戻すとスターターも伏せ札に戻る。
	c.starter = snap.starter
	c.placedCount = snap.placedCount
	c.phase = snap.phase
	if len(c.actionLog) > snap.actionLogLn {
		c.actionLog = c.actionLog[:snap.actionLogLn]
	}
	return nil
}

// CanUndo はアンドゥ可能かを返す。
func (c *CribbageSquares) CanUndo() bool { return len(c.history) > 0 }

// GiveUp はゲームを放棄する。
//
// 途中で投げた盤面はスターターをめくらない。埋まっていない手は 0 点なので、
// 合計点も伸びないまま終わる。
func (c *CribbageSquares) GiveUp() {
	if c.phase == CribbageSquaresPhasePlaying {
		c.phase = CribbageSquaresPhaseComplete
		c.currentCard = nil
		c.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetPhase はフェーズを返す。
func (c *CribbageSquares) GetPhase() CribbageSquaresPhase { return c.phase }

// GetBoard はボードを返す。
func (c *CribbageSquares) GetBoard() [CribbageSquaresGridSize][CribbageSquaresGridSize]*Card {
	return c.board
}

// GetCurrentCard は次に配置するカードを返す。
func (c *CribbageSquares) GetCurrentCard() *Card { return c.currentCard }

// GetStarter はスターター（17 枚目）を返す。めくる前は nil。
func (c *CribbageSquares) GetStarter() *Card { return c.starter }

// GetPlacedCount は配置済みカード枚数を返す。
func (c *CribbageSquares) GetPlacedCount() int { return c.placedCount }

// GetActionLog は棋譜を返す。
func (c *CribbageSquares) GetActionLog() []*ActionLogEntry { return c.actionLog }

// GetGameEndFlag returns true once the game has left the playing phase.
func (c *CribbageSquares) GetGameEndFlag() bool { return c.phase != CribbageSquaresPhasePlaying }

// IsComplete は 16 マスすべてが埋まったかを返す。
func (c *CribbageSquares) IsComplete() bool {
	return c.placedCount >= CribbageSquaresTotalCells
}

// IsWin は合計得点がクリア基準に達したかを返す。
func (c *CribbageSquares) IsWin() bool { return c.TotalScore() >= CribbageSquaresWinScore }

// RowCards は指定行のカードを返す。埋まっていない場合は nil。
func (c *CribbageSquares) RowCards(r int) []*Card {
	if r < 0 || r >= CribbageSquaresGridSize {
		return nil
	}
	cards := make([]*Card, 0, CribbageSquaresGridSize)
	for i := range CribbageSquaresGridSize {
		if c.board[r][i] == nil {
			return nil
		}
		cards = append(cards, c.board[r][i])
	}
	return cards
}

// ColCards は指定列のカードを返す。埋まっていない場合は nil。
func (c *CribbageSquares) ColCards(col int) []*Card {
	if col < 0 || col >= CribbageSquaresGridSize {
		return nil
	}
	cards := make([]*Card, 0, CribbageSquaresGridSize)
	for i := range CribbageSquaresGridSize {
		if c.board[i][col] == nil {
			return nil
		}
		cards = append(cards, c.board[i][col])
	}
	return cards
}

// RowDetail は指定行のクリベッジ得点内訳を返す。
//
// 行が埋まっていないか、スターターがまだめくられていない間は空の内訳を返す。
// 4 枚だけで採点すると本来より低い点が出てしまい、途中経過として誤解を招く。
func (c *CribbageSquares) RowDetail(r int) CribbageScoreDetail {
	return c.lineDetail(c.RowCards(r))
}

// ColDetail は指定列のクリベッジ得点内訳を返す。
func (c *CribbageSquares) ColDetail(col int) CribbageScoreDetail {
	return c.lineDetail(c.ColCards(col))
}

// lineDetail は 1 手ぶんの内訳を返す。クリブではないので isCrib は false。
func (c *CribbageSquares) lineDetail(cards []*Card) CribbageScoreDetail {
	if cards == nil || c.starter == nil {
		return CribbageScoreDetail{}
	}
	return CribbageScoreHand(cards, c.starter, false)
}

// RowScore は指定行の得点を返す。
func (c *CribbageSquares) RowScore(r int) int { return c.RowDetail(r).Total }

// ColScore は指定列の得点を返す。
func (c *CribbageSquares) ColScore(col int) int { return c.ColDetail(col).Total }

// TotalScore は 8 手の合計得点を返す。
func (c *CribbageSquares) TotalScore() int {
	total := 0
	for i := range CribbageSquaresGridSize {
		total += c.RowScore(i)
		total += c.ColScore(i)
	}
	return total
}

// GetHint は現在のカードを置く最善のセルを返す。
//
// 各空きセルについて、そのカードを置いたときに行と列で新たに成立する
// 15・ペア・ランの点を数え、合計が最大のセルを推奨する。
func (c *CribbageSquares) GetHint() *CribbageSquaresHint {
	if c.phase != CribbageSquaresPhasePlaying || c.currentCard == nil {
		return nil
	}
	var best *CribbageSquaresHint
	for r := range CribbageSquaresGridSize {
		for col := range CribbageSquaresGridSize {
			if c.board[r][col] != nil {
				continue
			}
			score := c.placementGain(c.currentCard, r, col)
			if best == nil || score > best.Score {
				best = &CribbageSquaresHint{Row: r, Col: col, Score: score, Synergy: score > 0}
			}
		}
	}
	return best
}

// placementGain は card を (row, col) に置いたときの行と列の増分点の合計。
func (c *CribbageSquares) placementGain(card *Card, row, col int) int {
	return cribbageSquaresLineGain(c.partialRow(row), card) +
		cribbageSquaresLineGain(c.partialCol(col), card)
}

// partialRow は指定行の埋まっているカードだけを返す。
func (c *CribbageSquares) partialRow(r int) []*Card {
	cards := make([]*Card, 0, CribbageSquaresGridSize)
	for i := range CribbageSquaresGridSize {
		if c.board[r][i] != nil {
			cards = append(cards, c.board[r][i])
		}
	}
	return cards
}

// partialCol は指定列の埋まっているカードだけを返す。
func (c *CribbageSquares) partialCol(col int) []*Card {
	cards := make([]*Card, 0, CribbageSquaresGridSize)
	for i := range CribbageSquaresGridSize {
		if c.board[i][col] != nil {
			cards = append(cards, c.board[i][col])
		}
	}
	return cards
}

// cribbageSquaresLineGain は既存のカード列へ card を足したときの増分点。
func cribbageSquaresLineGain(existing []*Card, card *Card) int {
	after := make([]*Card, 0, len(existing)+1)
	after = append(after, existing...)
	after = append(after, card)
	return cribbageSquaresPartialScore(after) - cribbageSquaresPartialScore(existing)
}

// cribbageSquaresPartialScore は途中の列に対して数えられる点だけを合計する。
//
// **フラッシュとノブズは意図的に除いている。** どちらも手が 4 枚揃い、かつ
// スターターがめくられて初めて確定するので、途中の盤面に対して足すと
// 実現しない点でヒントが歪む。同スートを揃える価値は、揃った時点で
// フラッシュとして一気に入る。
func cribbageSquaresPartialScore(cards []*Card) int {
	if len(cards) == 0 {
		return 0
	}
	return CribbageScoreFifteens(cards) + CribbageScorePairs(cards) + CribbageScoreRuns(cards)
}

// validCribbageSquaresCell はセル位置が盤内かを返す。
func validCribbageSquaresCell(row, col int) bool {
	return row >= 0 && row < CribbageSquaresGridSize && col >= 0 && col < CribbageSquaresGridSize
}

// takeSnapshot は現在の状態をスナップショットとして保存する。
func (c *CribbageSquares) takeSnapshot() {
	c.history = append(c.history, &cribbageSquaresSnapshot{
		board:       c.board,
		currentCard: c.currentCard,
		starter:     c.starter,
		placedCount: c.placedCount,
		phase:       c.phase,
		deckDrawCnt: c.trumpCards.deckDrawCnt,
		actionLogLn: len(c.actionLog),
	})
}

// appendLog は棋譜エントリを追加する。
func (c *CribbageSquares) appendLog(actionType, detail string, cards []*Card) {
	c.actionLog = append(c.actionLog, &ActionLogEntry{
		TurnNumber: c.placedCount,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// cribbageSquaresJSON はシリアライズ用のワイヤーフォーマット。
type cribbageSquaresJSON struct {
	TrumpCards  *TrumpCards                                             `json:"tc"`
	Board       [CribbageSquaresGridSize][CribbageSquaresGridSize]*Card `json:"bd"`
	CurrentCard *Card                                                   `json:"cc"`
	Starter     *Card                                                   `json:"st"`
	PlacedCount int                                                     `json:"pc"`
	Phase       CribbageSquaresPhase                                    `json:"ps"`
	ActionLog   []*ActionLogEntry                                       `json:"al"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo silently never works in production (#4478).
	History []*cribbageSquaresSnapshot `json:"hi,omitempty"`
}

// cribbageSquaresSnapshotJSON is the wire format for a single undo snapshot.
// cribbageSquaresSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods.
type cribbageSquaresSnapshotJSON struct {
	Board       [CribbageSquaresGridSize][CribbageSquaresGridSize]*Card `json:"bd"`
	CurrentCard *Card                                                   `json:"cc"`
	Starter     *Card                                                   `json:"st"`
	PlacedCount int                                                     `json:"pc"`
	Phase       CribbageSquaresPhase                                    `json:"ps"`
	DeckDrawCnt int                                                     `json:"dd"`
	ActionLogLn int                                                     `json:"ll"`
}

// MarshalJSON implements json.Marshaler for cribbageSquaresSnapshot.
func (s *cribbageSquaresSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(cribbageSquaresSnapshotJSON{
		Board:       s.board,
		CurrentCard: s.currentCard,
		Starter:     s.starter,
		PlacedCount: s.placedCount,
		Phase:       s.phase,
		DeckDrawCnt: s.deckDrawCnt,
		ActionLogLn: s.actionLogLn,
	})
}

// UnmarshalJSON implements json.Unmarshaler for cribbageSquaresSnapshot.
//
// DeckDrawCnt must be in [0, CardCnt] and ActionLogLn in
// [0, cribbageSquaresMaxSliceLen]: Undo() slices with both, so a malformed KV
// payload would otherwise panic the worker.
func (s *cribbageSquaresSnapshot) UnmarshalJSON(data []byte) error {
	var j cribbageSquaresSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.DeckDrawCnt < 0 || j.DeckDrawCnt > CardCnt {
		return errors.New("cribbagesquares: snapshot deckDrawCnt out of range")
	}
	if j.ActionLogLn < 0 || j.ActionLogLn > cribbageSquaresMaxSliceLen {
		return errors.New("cribbagesquares: snapshot actionLogLn out of range")
	}
	if j.PlacedCount < 0 || j.PlacedCount > CribbageSquaresTotalCells {
		return fmt.Errorf("cribbagesquares: snapshot placedCount out of range: %d", j.PlacedCount)
	}
	s.board = j.Board
	s.currentCard = j.CurrentCard
	s.starter = j.Starter
	s.placedCount = j.PlacedCount
	s.phase = j.Phase
	s.deckDrawCnt = j.DeckDrawCnt
	s.actionLogLn = j.ActionLogLn
	return nil
}

// MarshalJSON implements json.Marshaler.
func (c *CribbageSquares) MarshalJSON() ([]byte, error) {
	return json.Marshal(cribbageSquaresJSON{
		TrumpCards:  c.trumpCards,
		Board:       c.board,
		CurrentCard: c.currentCard,
		Starter:     c.starter,
		PlacedCount: c.placedCount,
		Phase:       c.phase,
		ActionLog:   c.actionLog,
		History:     c.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler. KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (c *CribbageSquares) UnmarshalJSON(data []byte) error {
	var j cribbageSquaresJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < CribbageSquaresPhasePlaying || j.Phase > CribbageSquaresPhaseComplete {
		return fmt.Errorf("cribbagesquares: invalid phase: %d", j.Phase)
	}
	if j.PlacedCount < 0 || j.PlacedCount > CribbageSquaresTotalCells {
		return fmt.Errorf("cribbagesquares: invalid placed count: %d", j.PlacedCount)
	}
	if len(j.ActionLog) > cribbageSquaresMaxSliceLen || len(j.History) > cribbageSquaresMaxSliceLen {
		return errors.New("cribbagesquares: input array exceeds maximum allowed size")
	}
	c.trumpCards = j.TrumpCards
	if c.trumpCards == nil {
		c.trumpCards = NewTrumpCards(0)
	}
	c.board = j.Board
	c.currentCard = j.CurrentCard
	c.starter = j.Starter
	c.placedCount = j.PlacedCount
	c.phase = j.Phase
	c.actionLog = j.ActionLog
	if c.actionLog == nil {
		c.actionLog = make([]*ActionLogEntry, 0)
	}
	c.history = j.History
	if c.history == nil {
		c.history = make([]*cribbageSquaresSnapshot, 0)
	}
	return nil
}
