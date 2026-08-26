//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// RankAndFilePhase フォーティシーブスゲームフェーズ
type RankAndFilePhase int

// RankAndFileのフェーズ定数
const (
	// RankAndFilePhasePlaying プレイ中
	RankAndFilePhasePlaying RankAndFilePhase = iota
	// RankAndFilePhaseGameClear ゲームクリア
	RankAndFilePhaseGameClear
	// RankAndFilePhaseGameOver ゲームオーバー
	RankAndFilePhaseGameOver
)

// RankAndFileTableauCnt タブローの列数
const RankAndFileTableauCnt = 10

// RankAndFileColSize 1 列に配る枚数。下 3 枚は伏せ、最上段だけ表向き。
const RankAndFileColSize = 4

// RankAndFileFoundationCnt ファンデーションの数
const RankAndFileFoundationCnt = 8

// RankAndFileTableauCard タブロー上のカード
type RankAndFileTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// RankAndFileHint ヒント
type RankAndFileHint struct {
	// FromZone は "waste" / "tableau"、どこにも手が無くストックが残っている
	// ときは "stock" (= 引くことを勧める、#5525)。
	FromZone  string
	FromCol   int
	CardIndex int
	// ToZone は "tableau" / "foundation"、FromZone が "stock" のときは "waste"。
	ToZone string
	ToCol  int
}

// RankAndFileConfig フォーティシーブスゲーム設定
type RankAndFileConfig struct{}

// RankAndFile フォーティシーブスゲームクラス
type RankAndFile struct {
	trumpCards *TrumpCards
	tableau    [RankAndFileTableauCnt][]*RankAndFileTableauCard
	stock      []*Card
	waste      []*Card
	foundation [RankAndFileFoundationCnt][]*Card
	phase      RankAndFilePhase
	moveCount  int
	actionLogBase
	history     []*rankAndFileSnapshot
	isStalemate bool
}

// rankAndFileSnapshot アンドゥ用スナップショット
type rankAndFileSnapshot struct {
	tableau     [RankAndFileTableauCnt][]*RankAndFileTableauCard
	stock       []*Card
	waste       []*Card
	foundation  [RankAndFileFoundationCnt][]*Card
	phase       RankAndFilePhase
	moveCount   int
	isStalemate bool
}

// NewRankAndFile コンストラクタ
func NewRankAndFile(trumpCards *TrumpCards) *RankAndFile {
	return &RankAndFile{
		trumpCards: trumpCards,
	}
}

// NewDefaultRankAndFile returns RankAndFile with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultRankAndFile() *RankAndFile {
	return NewRankAndFile(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (ft *RankAndFile) Reset() {
	ft.trumpCards.Shuffle()
	ft.phase = RankAndFilePhasePlaying
	ft.moveCount = 0
	ft.actionLog = nil
	ft.history = nil
	ft.isStalemate = false

	// **各列4枚のうち、下3枚は伏せて最上段だけ表向き。**クローン元の
	// Forty Thieves は40枚すべて表向きに配るので、そこが最初の分岐点になる。
	for i := range RankAndFileTableauCnt {
		ft.tableau[i] = make([]*RankAndFileTableauCard, 0, RankAndFileColSize)
		for j := range RankAndFileColSize {
			card := ft.trumpCards.DrawCard()
			ft.tableau[i] = append(ft.tableau[i], &RankAndFileTableauCard{
				Card:   card,
				FaceUp: j == RankAndFileColSize-1,
			})
		}
	}

	// ファンデーション初期化
	for i := range RankAndFileFoundationCnt {
		ft.foundation[i] = nil
	}

	// 残りをストックへ（64枚）
	ft.stock = nil
	ft.waste = nil
	for ft.trumpCards.GetRemainingCount() > 0 {
		card := ft.trumpCards.DrawCard()
		ft.stock = append(ft.stock, card)
	}
}

// Draw ストックからウェイストにカードを1枚引く（リサイクルなし）
func (ft *RankAndFile) Draw() error {
	if ft.phase != RankAndFilePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(ft.stock) == 0 {
		return errors.New("no cards in stock")
	}
	ft.takeSnapshot()
	card := ft.stock[len(ft.stock)-1]
	ft.stock = ft.stock[:len(ft.stock)-1]
	ft.waste = append(ft.waste, card)
	ft.moveCount++
	ft.appendLog("draw", "ストックからカードを引きました", []*Card{card})
	ft.checkRankAndFileStalemate()
	return nil
}

// MoveWasteToTableau ウェイストからタブローにカードを移動
func (ft *RankAndFile) MoveWasteToTableau(col int) error {
	if ft.phase != RankAndFilePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= RankAndFileTableauCnt {
		return errors.New("invalid column")
	}
	if len(ft.waste) == 0 {
		return errors.New("waste is empty")
	}
	card := ft.waste[len(ft.waste)-1]
	if !ft.canPlaceOnTableau(card, col) {
		return errors.New("cannot place card on tableau")
	}
	ft.takeSnapshot()
	ft.waste = ft.waste[:len(ft.waste)-1]
	ft.tableau[col] = append(ft.tableau[col], &RankAndFileTableauCard{Card: card, FaceUp: true})
	ft.moveCount++
	ft.appendLog("move", fmt.Sprintf("ウェイスト→タブロー列%d", col), []*Card{card})
	ft.checkRankAndFileStalemate()
	return nil
}

// MoveWasteToFoundation ウェイストからファンデーションにカードを移動
func (ft *RankAndFile) MoveWasteToFoundation() error {
	if ft.phase != RankAndFilePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(ft.waste) == 0 {
		return errors.New("waste is empty")
	}
	card := ft.waste[len(ft.waste)-1]
	fIdx := ft.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	ft.takeSnapshot()
	ft.waste = ft.waste[:len(ft.waste)-1]
	ft.foundation[fIdx] = append(ft.foundation[fIdx], card)
	ft.moveCount++
	ft.appendLog("move", "ウェイスト→ファンデーション", []*Card{card})
	ft.checkGameClear()
	ft.checkRankAndFileStalemate()
	return nil
}

// MoveTableauToTableau タブローからタブローにカードを移動（1枚のみ）
func (ft *RankAndFile) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if ft.phase != RankAndFilePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= RankAndFileTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= RankAndFileTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := ft.tableau[fromCol]
	// **-1 は「その列の上札」。**CUI の短縮形 `m <from> <to>` は枚数を指定しない
	// ので -1 を渡してくる。素直に負値を弾くと、ヘルプに載っている短縮形が必ず
	// エラーになる。-1 以外の負値は従来どおり無効。
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	// **並びの一部でも一括で動かせる。**クローン元の Forty Thieves は先頭以外を
	// 拒むので、ここが3つ目の分岐点。同じ一族の Deauville は上札のみ、Emperor は
	// 列まるごとで、Rank and File だけが「任意の並び」を許す。
	if !ft.isRankAndFileSequence(fromCol, cardIndex) {
		return errors.New("cards below the top do not form an alternating-colour descending sequence")
	}
	moving := fromCards[cardIndex:]
	if !ft.canPlaceOnTableau(moving[0].Card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	ft.takeSnapshot()
	ft.tableau[toCol] = append(ft.tableau[toCol], moving...)
	ft.tableau[fromCol] = fromCards[:cardIndex]
	ft.autoFlipTop(fromCol)
	ft.moveCount++
	movedCards := make([]*Card, len(moving))
	for i, m := range moving {
		movedCards[i] = m.Card
	}
	ft.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), movedCards)
	ft.checkRankAndFileStalemate()
	return nil
}

// autoFlipTop は列の最上段が伏せ札ならめくる。
//
// **伏せ札があるので、札が減ったら必ず呼ぶ。**呼び忘れると、下に札があるのに
// 何も置けない列ができて盤が静かに詰む。
func (ft *RankAndFile) autoFlipTop(col int) {
	cards := ft.tableau[col]
	if len(cards) == 0 {
		return
	}
	if top := cards[len(cards)-1]; !top.FaceUp {
		top.FaceUp = true
	}
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (ft *RankAndFile) MoveTableauToFoundation(col int) error {
	if ft.phase != RankAndFilePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= RankAndFileTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := ft.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	// **伏せ札は組札にも送れない。**めくれていない札は見えていない。
	if !tc.FaceUp {
		return errors.New("the top card is face down")
	}
	card := tc.Card
	fIdx := ft.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	ft.takeSnapshot()
	ft.tableau[col] = fromCards[:len(fromCards)-1]
	ft.foundation[fIdx] = append(ft.foundation[fIdx], card)
	// **札が減ったら必ずめくる。**Forty Thieves は全部表向きなので不要だった。
	ft.autoFlipTop(col)
	ft.moveCount++
	ft.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	ft.checkGameClear()
	ft.checkRankAndFileStalemate()
	return nil
}

// GiveUp ギブアップ
func (ft *RankAndFile) GiveUp() {
	if ft.phase == RankAndFilePhasePlaying {
		ft.phase = RankAndFilePhaseGameOver
		ft.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (ft *RankAndFile) GetHint() *RankAndFileHint {
	if ft.phase != RankAndFilePhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := range RankAndFileTableauCnt {
		if len(ft.tableau[col]) == 0 {
			continue
		}
		tc := ft.tableau[col][len(ft.tableau[col])-1]
		// **伏せ札は勧めない。**送れないので、勧めても弾かれるだけ。
		if !tc.FaceUp {
			continue
		}
		fIdx := ft.findFoundation(tc.Card)
		if fIdx >= 0 {
			return &RankAndFileHint{
				FromZone:  "tableau",
				FromCol:   col,
				CardIndex: len(ft.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: ウェイストからファンデーションへ
	if len(ft.waste) > 0 {
		card := ft.waste[len(ft.waste)-1]
		fIdx := ft.findFoundation(card)
		if fIdx >= 0 {
			return &RankAndFileHint{
				FromZone:  "waste",
				FromCol:   -1,
				CardIndex: -1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度3: タブローからタブローへ。
	//
	// **上札だけを見てはいけない。**Rank and File は並びの一部を一括で動かせるので、
	// 上札が行き詰まっていても、その下から始まる並びは動けることがある。上札しか
	// 見ないと「手があるのに手詰まり」と宣言する ── クローン元の Forty Thieves は
	// 1枚ずつしか動かせないので、あちらの走査を流用すると必ずこの穴が開く。
	for fromCol := range RankAndFileTableauCnt {
		for _, cardIndex := range ft.sequenceStarts(fromCol) {
			card := ft.tableau[fromCol][cardIndex].Card
			for toCol := range RankAndFileTableauCnt {
				if toCol == fromCol {
					continue
				}
				// 空列への移動はヒントとして提示しない（意味のない移動）
				if len(ft.tableau[toCol]) == 0 {
					continue
				}
				if ft.canPlaceOnTableau(card, toCol) {
					return &RankAndFileHint{
						FromZone:  "tableau",
						FromCol:   fromCol,
						CardIndex: cardIndex,
						ToZone:    "tableau",
						ToCol:     toCol,
					}
				}
			}
		}
	}
	// 優先度4: ウェイストからタブローへ
	if len(ft.waste) > 0 {
		card := ft.waste[len(ft.waste)-1]
		for toCol := range RankAndFileTableauCnt {
			if ft.canPlaceOnTableau(card, toCol) {
				return &RankAndFileHint{
					FromZone:  "waste",
					FromCol:   -1,
					CardIndex: -1,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	// 優先度5: 盤上に手が無く、ストックが残っているなら「引く」。
	//
	// **ここで nil を返すと「ヒントはありません」になる。**行き詰まり判定
	// (checkRankAndFileStalemate) はストックが残っていれば手詰まりではないと
	// 正しく扱うのに、ヒントだけが同じ局面で黙っていた。プレイヤーからは
	// 「詰んだ」のか「引けば良いだけ」なのか区別が付かない (#5525)。
	if len(ft.stock) > 0 {
		return &RankAndFileHint{
			FromZone:  "stock",
			FromCol:   -1,
			CardIndex: -1,
			ToZone:    "waste",
			ToCol:     -1,
		}
	}
	return nil
}

// AutoComplete オートコンプリート（ストックが空の場合に自動でファンデーションへ移動）
func (ft *RankAndFile) AutoComplete() error {
	if ft.phase != RankAndFilePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if !ft.AllFaceUp() {
		return errors.New("not all cards are face up")
	}
	ft.takeSnapshot()
	for {
		moved := false
		// ウェイストからファンデーションへ
		for len(ft.waste) > 0 {
			card := ft.waste[len(ft.waste)-1]
			fIdx := ft.findFoundation(card)
			if fIdx < 0 {
				break
			}
			ft.waste = ft.waste[:len(ft.waste)-1]
			ft.foundation[fIdx] = append(ft.foundation[fIdx], card)
			ft.moveCount++
			moved = true
		}
		// タブローからファンデーションへ
		for col := range RankAndFileTableauCnt {
			if len(ft.tableau[col]) == 0 {
				continue
			}
			tc := ft.tableau[col][len(ft.tableau[col])-1]
			card := tc.Card
			fIdx := ft.findFoundation(card)
			if fIdx < 0 {
				continue
			}
			ft.tableau[col] = ft.tableau[col][:len(ft.tableau[col])-1]
			ft.foundation[fIdx] = append(ft.foundation[fIdx], card)
			ft.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	ft.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	ft.checkGameClear()
	return nil
}

// AllFaceUp 全カードが表向きかどうか。
//
// **タブローの伏せ札も見る。**クローン元の Forty Thieves は40枚すべて表向きに
// 配るので「ストックが空 = 全部見えている」で足りたが、Rank and File は各列
// 3枚を伏せる。その判定のままだと、伏せ札が残っているのに AutoComplete が
// 走ってしまう。
func (ft *RankAndFile) AllFaceUp() bool {
	if len(ft.stock) > 0 || len(ft.waste) > 0 {
		return false
	}
	for col := range RankAndFileTableauCnt {
		for _, tc := range ft.tableau[col] {
			if !tc.FaceUp {
				return false
			}
		}
	}
	return true
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (ft *RankAndFile) GetPhase() RankAndFilePhase { return ft.phase }

// SetPhase フェーズ設定 (テスト用)
func (ft *RankAndFile) SetPhase(phase RankAndFilePhase) { ft.phase = phase }

// GetMoveCount 移動回数取得
func (ft *RankAndFile) GetMoveCount() int { return ft.moveCount }

// GetStockCount ストック枚数取得
func (ft *RankAndFile) GetStockCount() int { return len(ft.stock) }

// GetWaste ウェイスト取得
func (ft *RankAndFile) GetWaste() []*Card { return ft.waste }

// GetTableau タブロー取得
func (ft *RankAndFile) GetTableau() [RankAndFileTableauCnt][]*RankAndFileTableauCard {
	return ft.tableau
}

// GetFoundation ファンデーション取得
func (ft *RankAndFile) GetFoundation() [RankAndFileFoundationCnt][]*Card { return ft.foundation }

// GetGameEndFlag returns true once the game has left the playing phase.
func (ft *RankAndFile) GetGameEndFlag() bool { return ft.phase != RankAndFilePhasePlaying }

// IsStalemate 手詰まり状態取得
func (ft *RankAndFile) IsStalemate() bool { return ft.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (ft *RankAndFile) SetIsStalemate(v bool) { ft.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (ft *RankAndFile) SetTableau(tableau [RankAndFileTableauCnt][]*RankAndFileTableauCard) {
	ft.tableau = tableau
}

// SetStock ストック設定 (テスト用)
func (ft *RankAndFile) SetStock(stock []*Card) { ft.stock = stock }

// SetWaste ウェイスト設定 (テスト用)
func (ft *RankAndFile) SetWaste(waste []*Card) { ft.waste = waste }

// SetFoundation ファンデーション設定 (テスト用)
func (ft *RankAndFile) SetFoundation(foundation [RankAndFileFoundationCnt][]*Card) {
	ft.foundation = foundation
}

// Undo 直前の操作を取り消す
func (ft *RankAndFile) Undo() error {
	if ft.phase != RankAndFilePhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(ft.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := ft.history[len(ft.history)-1]
	ft.history = ft.history[:len(ft.history)-1]
	ft.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (ft *RankAndFile) CanUndo() bool {
	return len(ft.history) > 0 && ft.phase == RankAndFilePhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (ft *RankAndFile) UndoToEscape() int {
	return undoToEscape(ft.isStalemate, ft.history, func(s *rankAndFileSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (ft *RankAndFile) UndoN(n int) error {
	return undoN(ft, n)
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定（同スート降順）
func (ft *RankAndFile) canPlaceOnTableau(card *Card, col int) bool {
	colCards := ft.tableau[col]
	if len(colCards) == 0 {
		// 空の列にはどのカードでも置ける
		return true
	}
	top := colCards[len(colCards)-1]
	// **伏せ札の上には置けない。**めくれていない札は積み先にならない。
	if !top.FaceUp {
		return false
	}
	// **異色で降順。**クローン元の Forty Thieves は「同スートで降順」なので、
	// その述語を残すと赤の上に赤を積めてしまい、置けるはずの手が置けなくなる。
	return rankAndFileIsRed(card.GetDesign()) != rankAndFileIsRed(top.Card.GetDesign()) &&
		card.GetValue() == top.Card.GetValue()-1
}

// rankAndFileIsRed 赤いスートか。
//
// MissMilligan にも同じ判定があるが、あちらは `extra2` タグなので extra4 から
// 参照すると stranded symbol になる。共有ヘルパではないので自前に持つ。
func rankAndFileIsRed(design int) bool {
	return design == CardDesignHeart || design == CardDesignDiamond
}

// sequenceStarts は列 col で「そこから上が異色降順に並んでいる」開始位置を、
// 上札から順に返す。
//
// 動かせる単位はこの位置から上の塊だけなので、合法手の探索も UI の選択可否も
// この一覧で決まる。並びは必ず上札から連続するため、上から下へ 1 つずつ伸ばして
// 崩れた時点で止めればよい。伏せ札に当たったところで必ず止まる。
func (ft *RankAndFile) sequenceStarts(col int) []int {
	if col < 0 || col >= RankAndFileTableauCnt {
		return nil
	}
	cards := ft.tableau[col]
	starts := make([]int, 0, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		if !ft.isRankAndFileSequence(col, i) {
			break
		}
		starts = append(starts, i)
	}
	return starts
}

// SequenceStarts は列 col の移動開始位置一覧 (上札から順)。UI が「掘り下げた札を
// 掴めるか」を判断するために使う。
func (ft *RankAndFile) SequenceStarts(col int) []int { return ft.sequenceStarts(col) }

// isRankAndFileSequence は列 col の cardIndex 以降が異色降順に並んでいるかを返す。
//
// 一括で動かせるのはこの並びだけで、途中で色かランクが切れたら1枚も動かない。
// 伏せ札が混ざっていたら並びとして扱わない。
func (ft *RankAndFile) isRankAndFileSequence(col, cardIndex int) bool {
	cards := ft.tableau[col]
	for i := cardIndex; i < len(cards); i++ {
		if !cards[i].FaceUp {
			return false
		}
		if i+1 < len(cards) {
			upper, lower := cards[i].Card, cards[i+1].Card
			if rankAndFileIsRed(upper.GetDesign()) == rankAndFileIsRed(lower.GetDesign()) ||
				lower.GetValue() != upper.GetValue()-1 {
				return false
			}
		}
	}
	return true
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (ft *RankAndFile) canPlaceOnFoundation(card *Card, fIdx int) bool {
	return canPlaceOnFoundationPile(ft.foundation[fIdx], card)
}

// findFoundation カードを置けるファンデーションのインデックスを探す（見つからない場合-1）
func (ft *RankAndFile) findFoundation(card *Card) int {
	for i := range RankAndFileFoundationCnt {
		if ft.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定
func (ft *RankAndFile) checkGameClear() {
	for i := range RankAndFileFoundationCnt {
		if len(ft.foundation[i]) != CardValueMax {
			return
		}
	}
	ft.phase = RankAndFilePhaseGameClear
}

// checkRankAndFileStalemate 手詰まり判定
func (ft *RankAndFile) checkRankAndFileStalemate() {
	if ft.phase != RankAndFilePhasePlaying {
		return
	}
	hint := ft.GetHint()
	if hint != nil {
		ft.isStalemate = false
		return
	}
	// ヒントがない場合
	if len(ft.stock) == 0 && len(ft.waste) == 0 {
		// ストックもウェイストも空で移動先なし → 手詰まり
		ft.isStalemate = true
		return
	}
	if len(ft.stock) == 0 {
		// ストック空でリサイクルなし、ウェイストのカードも移動不可 → 手詰まり
		ft.isStalemate = true
		return
	}
	// ストックにカードが残っている場合はまだ引ける
	ft.isStalemate = false
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (ft *RankAndFile) takeSnapshot() {
	snap := &rankAndFileSnapshot{
		phase:       ft.phase,
		moveCount:   ft.moveCount,
		isStalemate: ft.isStalemate,
	}
	// deep copy tableau
	for i := range RankAndFileTableauCnt {
		snap.tableau[i] = make([]*RankAndFileTableauCard, len(ft.tableau[i]))
		for j, tc := range ft.tableau[i] {
			snap.tableau[i][j] = &RankAndFileTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	// deep copy stock
	snap.stock = make([]*Card, len(ft.stock))
	copy(snap.stock, ft.stock)
	// deep copy waste
	snap.waste = make([]*Card, len(ft.waste))
	copy(snap.waste, ft.waste)
	// deep copy foundation
	for i := range RankAndFileFoundationCnt {
		snap.foundation[i] = make([]*Card, len(ft.foundation[i]))
		copy(snap.foundation[i], ft.foundation[i])
	}
	ft.history = appendSnapshot(ft.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (ft *RankAndFile) restoreSnapshot(snap *rankAndFileSnapshot) {
	ft.tableau = snap.tableau
	ft.stock = snap.stock
	ft.waste = snap.waste
	ft.foundation = snap.foundation
	ft.phase = snap.phase
	ft.moveCount = snap.moveCount
	ft.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (ft *RankAndFile) appendLog(actionType, detail string, cards []*Card) {
	ft.appendLogAt(ft.moveCount, 0, actionType, detail, cards)
}

// rankAndFileJSON is the JSON wire format for RankAndFile.
type rankAndFileJSON struct {
	TrumpCards  *TrumpCards                                      `json:"tc"`
	Tableau     [RankAndFileTableauCnt][]*RankAndFileTableauCard `json:"tb"`
	Stock       []*Card                                          `json:"st"`
	Waste       []*Card                                          `json:"wa"`
	Foundation  [RankAndFileFoundationCnt][]*Card                `json:"fd"`
	Phase       RankAndFilePhase                                 `json:"ps"`
	MoveCount   int                                              `json:"mc"`
	ActionLog   []*ActionLogEntry                                `json:"al"`
	IsStalemate bool                                             `json:"sl"`
	History     []*rankAndFileSnapshot                           `json:"hi,omitempty"`
}

// rankAndFileSnapshotJSON is the wire format for a single undo snapshot.
// rankAndFileSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// rankAndFileJSON's short keys to keep the KV payload compact (#1654).
type rankAndFileSnapshotJSON struct {
	Tableau     [RankAndFileTableauCnt][]*RankAndFileTableauCard `json:"tb"`
	Stock       []*Card                                          `json:"st"`
	Waste       []*Card                                          `json:"wa"`
	Foundation  [RankAndFileFoundationCnt][]*Card                `json:"fd"`
	Phase       RankAndFilePhase                                 `json:"ps"`
	MoveCount   int                                              `json:"mc"`
	IsStalemate bool                                             `json:"sl"`
}

// MarshalJSON implements json.Marshaler for rankAndFileSnapshot.
func (s *rankAndFileSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(rankAndFileSnapshotJSON{
		Tableau:     s.tableau,
		Stock:       s.stock,
		Waste:       s.waste,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for rankAndFileSnapshot.
func (s *rankAndFileSnapshot) UnmarshalJSON(data []byte) error {
	var j rankAndFileSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > rankAndFileMaxSliceLen || len(j.Waste) > rankAndFileMaxSliceLen {
		return fmt.Errorf("rankandfile: snapshot array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > rankAndFileMaxSliceLen {
			return fmt.Errorf("rankandfile: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > rankAndFileMaxSliceLen {
			return fmt.Errorf("rankandfile: snapshot foundation pile exceeds maximum allowed size")
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
	return nil
}

// MarshalJSON implements json.Marshaler.
func (ft *RankAndFile) MarshalJSON() ([]byte, error) {
	return json.Marshal(rankAndFileJSON{
		TrumpCards:  ft.trumpCards,
		Tableau:     ft.tableau,
		Stock:       ft.stock,
		Waste:       ft.waste,
		Foundation:  ft.foundation,
		Phase:       ft.phase,
		MoveCount:   ft.moveCount,
		ActionLog:   ft.actionLog,
		IsStalemate: ft.isStalemate,
		History:     ft.history,
	})
}

// rankAndFileMaxSliceLen caps slice sizes during deserialisation.
const rankAndFileMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (ft *RankAndFile) UnmarshalJSON(data []byte) error {
	var j rankAndFileJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > rankAndFileMaxSliceLen || len(j.Waste) > rankAndFileMaxSliceLen ||
		len(j.ActionLog) > rankAndFileMaxSliceLen || len(j.History) > rankAndFileMaxSliceLen {
		return fmt.Errorf("rankandfile: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > rankAndFileMaxSliceLen {
			return fmt.Errorf("rankandfile: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > rankAndFileMaxSliceLen {
			return fmt.Errorf("rankandfile: foundation pile exceeds maximum allowed size")
		}
	}

	ft.trumpCards = j.TrumpCards
	if ft.trumpCards == nil {
		ft.trumpCards = NewTrumpCardsWithDecks(2, 0)
	}
	ft.tableau = j.Tableau
	ft.stock = j.Stock
	if ft.stock == nil {
		ft.stock = make([]*Card, 0)
	}
	ft.waste = j.Waste
	if ft.waste == nil {
		ft.waste = make([]*Card, 0)
	}
	ft.foundation = j.Foundation
	ft.phase = j.Phase
	ft.moveCount = j.MoveCount
	ft.actionLog = j.ActionLog
	if ft.actionLog == nil {
		ft.actionLog = make([]*ActionLogEntry, 0)
	}
	ft.history = j.History
	if ft.history == nil {
		ft.history = make([]*rankAndFileSnapshot, 0)
	}
	ft.isStalemate = j.IsStalemate
	return nil
}
