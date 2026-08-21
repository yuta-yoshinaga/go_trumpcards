//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// StHelenaPhase セント・ヘレナ・ソリティアのゲームフェーズ。
type StHelenaPhase int

// StHelena のフェーズ定数。
const (
	// StHelenaPhasePlaying プレイ中。
	StHelenaPhasePlaying StHelenaPhase = iota
	// StHelenaPhaseGameClear ゲームクリア。
	StHelenaPhaseGameClear
	// StHelenaPhaseGameOver ゲームオーバー (ギブアップまたは手詰まり)。
	StHelenaPhaseGameOver
)

// StHelenaTableauCnt タブローの列数。組札の周りに円状に並ぶ 12 山。
// 上 4 / 下 4 / 左右 2 ずつで、この並びが送り先の制限を決める。
const StHelenaTableauCnt = 12

// StHelenaTableauInitialSize 初期配置時の各タブロー列の枚数。
// 104 枚から種札 8 枚を抜いた 96 枚を 12 山に配るので 8 枚ずつ。
const StHelenaTableauInitialSize = 8

// StHelenaSideColumns 上下どちらの組札へも送れる左右の列。
// 円状に並べたときに上段とも下段とも隣り合う位置。
var StHelenaSideColumns = [...]int{4, 5, 10, 11}

// StHelenaTopColumnCnt 上段 (K 段) にしか送れない列の数。列 0..3 が該当する。
const StHelenaTopColumnCnt = 4

// StHelenaFoundationCnt ファンデーション数 (4 = 昇順 A→K、4 = 降順 K→A)。
const StHelenaFoundationCnt = 8

// StHelenaAscendingFoundationCnt 昇順ファンデーションの数 (前半 0..3)。
const StHelenaAscendingFoundationCnt = 4

// StHelenaMaxRedeals 1 ゲーム中に許される再配りの回数。
const StHelenaMaxRedeals = 2

// StHelenaTableauCard タブロー上の 1 枚 (常に表向き)。
type StHelenaTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// StHelenaHint ヒント。
//
//	FromCol  ヒントの元となるタブロー列。
//	ToZone   "tableau" もしくは "foundation"。
//	ToCol    タブローなら列番号、ファンデーションならファンデーション ID。
//	Redeal   true の場合は再配りを推奨するヒント (FromCol/ToCol は -1)。
type StHelenaHint struct {
	FromCol int
	ToZone  string
	ToCol   int
	Redeal  bool
}

// StHelenaConfig セント・ヘレナ・ソリティアのゲーム設定 (現状フィールド無し)。
type StHelenaConfig struct{}

// StHelena セント・ヘレナ・ソリティアの本体。
type StHelena struct {
	trumpCards       *TrumpCards
	tableau          [StHelenaTableauCnt][]*StHelenaTableauCard
	foundation       [StHelenaFoundationCnt][]*Card
	phase            StHelenaPhase
	moveCount        int
	redealsRemaining int
	// restrictionsActive 初回の配りの間だけ立つ。上下の列がそれぞれ片方の
	// 組札にしか送れない制限で、最初の再配りで解ける。
	restrictionsActive bool
	actionLogBase
	history     []*stHelenaSnapshot
	isStalemate bool
}

// stHelenaSnapshot アンドゥ用スナップショット。
type stHelenaSnapshot struct {
	tableau          [StHelenaTableauCnt][]*StHelenaTableauCard
	foundation       [StHelenaFoundationCnt][]*Card
	phase            StHelenaPhase
	moveCount        int
	redealsRemaining int
	// **制限も巻き戻す。**再配りを undo したのに制限が解けたままだと、
	// 初回の配りでは打てないはずの手が打てる盤が残る。
	restrictionsActive bool
	isStalemate        bool
}

// NewStHelena コンストラクタ。
func NewStHelena(trumpCards *TrumpCards) *StHelena {
	return &StHelena{trumpCards: trumpCards}
}

// NewDefaultStHelena は 2 デッキ (104 枚) を持つ StHelena を返す。
// CUI / Web / Worker の生成サイトはすべてこの関数を経由する。
func NewDefaultStHelena() *StHelena {
	return NewStHelena(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲーム初期化。各スートの A 1 枚と K 1 枚を取り出してファンデーションに種を置き、
// 各スートの A と K を 1 枚ずつ抜いて組札の種札にし、残り 96 枚を 12 列 × 8 枚の
// タブローへ配る。再配り回数は 2、送り先の制限は有効にリセットされる。
func (cr *StHelena) Reset() {
	cr.trumpCards.Shuffle()
	cr.phase = StHelenaPhasePlaying
	cr.moveCount = 0
	cr.redealsRemaining = StHelenaMaxRedeals
	cr.restrictionsActive = true
	cr.actionLog = nil
	cr.history = nil
	cr.isStalemate = false

	deck := make([]*Card, 0, CardCnt*2)
	for cr.trumpCards.GetRemainingCount() > 0 {
		deck = append(deck, cr.trumpCards.DrawCard())
	}

	for i := range StHelenaFoundationCnt {
		cr.foundation[i] = nil
	}

	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	for idx, suit := range suits {
		aceIdx := stHelenaFindFirstCard(deck, suit, 1)
		if aceIdx < 0 {
			panic(fmt.Sprintf("sthelena Reset: ace of suit %d not found in deck", suit))
		}
		deck, cr.foundation[idx] = stHelenaTakeCardAt(deck, aceIdx), []*Card{deck[aceIdx]}
		kingIdx := stHelenaFindFirstCard(deck, suit, CardValueMax)
		if kingIdx < 0 {
			panic(fmt.Sprintf("sthelena Reset: king of suit %d not found in deck", suit))
		}
		deck, cr.foundation[idx+StHelenaAscendingFoundationCnt] = stHelenaTakeCardAt(deck, kingIdx), []*Card{deck[kingIdx]}
	}

	for i := range StHelenaTableauCnt {
		cr.tableau[i] = make([]*StHelenaTableauCard, 0, StHelenaTableauInitialSize)
		for j := range StHelenaTableauInitialSize {
			card := deck[i*StHelenaTableauInitialSize+j]
			cr.tableau[i] = append(cr.tableau[i], &StHelenaTableauCard{Card: card, FaceUp: true})
		}
	}
}

// stHelenaFindFirstCard は deck から (design, value) に最初にマッチするカードのインデックスを返す。
// 見つからない場合は -1 を返す。
func stHelenaFindFirstCard(deck []*Card, design, value int) int {
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

// stHelenaTakeCardAt は deck から idx 番目を取り除いた新しいスライスを返す。idx が範囲外なら元のスライスを返す。
func stHelenaTakeCardAt(deck []*Card, idx int) []*Card {
	if idx < 0 || idx >= len(deck) {
		return deck
	}
	out := make([]*Card, 0, len(deck)-1)
	out = append(out, deck[:idx]...)
	out = append(out, deck[idx+1:]...)
	return out
}

// MoveTableauToTableau は fromCol の最上段カードを toCol へ移す。
// **スート不問で値 ±1。K と A は繋がらない。**空タブローは移動先にできない
// （このゲームの空列は二度と埋まらない）。
func (cr *StHelena) MoveTableauToTableau(fromCol, toCol int) error {
	if cr.phase != StHelenaPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= StHelenaTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= StHelenaTableauCnt {
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
	cr.checkStHelenaStalemate()
	return nil
}

// stHelenaColumnCanReach は初回の配りの間、列 fromCol が組札 fIdx へ送れるかを返す。
//
// **上 4 列は K 段 (降順、4..7) だけ、下 4 列は A 段 (昇順、0..3) だけ。**
// 左右の 4 列はどちらへも送れる。円状に並べたとき、それぞれの列が隣り合って
// いる段しか使えない、という規則。最初の再配りで解ける。
func (cr *StHelena) columnCanReach(fromCol, fIdx int) bool {
	if !cr.restrictionsActive {
		return true
	}
	for _, side := range StHelenaSideColumns {
		if fromCol == side {
			return true
		}
	}
	ascending := fIdx < StHelenaAscendingFoundationCnt
	if fromCol < StHelenaTopColumnCnt {
		// 上の列は降順 (K 段) のみ。
		return !ascending
	}
	// 下の列は昇順 (A 段) のみ。
	return ascending
}

// MoveTableauToFoundation は fromCol の最上段カードを foundationIdx へ移す。
// 昇順 (0..3) / 降順 (4..7) の方向はインデックスから決まる。
func (cr *StHelena) MoveTableauToFoundation(fromCol, foundationIdx int) error {
	if cr.phase != StHelenaPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= StHelenaTableauCnt {
		return errors.New("invalid from column")
	}
	if foundationIdx < 0 || foundationIdx >= StHelenaFoundationCnt {
		return errors.New("invalid foundation index")
	}
	fromCards := cr.tableau[fromCol]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	if !cr.columnCanReach(fromCol, foundationIdx) {
		return NewDomainErrorCode(ErrInvalidPlay, "sthelena.errRestricted", nil)
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
	cr.checkStHelenaStalemate()
	return nil
}

// Redeal は再配りを実行する。各タブロー列を逆順に並べ替える。残り回数 0 か非プレイ中ならエラー。
func (cr *StHelena) Redeal() error {
	if cr.phase != StHelenaPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if cr.redealsRemaining <= 0 {
		return errors.New("no redeals remaining")
	}
	cr.takeSnapshot()
	// **最後の列から集めて、同じ順に配り直す。**シャッフルはしない。クレセントは
	// 各列をその場で逆順にするだけなので、列をまたいだ並べ替えが起きない ──
	// このゲームで再配りが手を生むのは、まさに列をまたいで並び替わるから。
	gathered := make([]*StHelenaTableauCard, 0, CardCnt*2)
	for i := StHelenaTableauCnt - 1; i >= 0; i-- {
		gathered = append(gathered, cr.tableau[i]...)
		cr.tableau[i] = nil
	}
	for i, tc := range gathered {
		col := i / StHelenaTableauInitialSize
		if col >= StHelenaTableauCnt {
			// 組札へ送った枚数だけ全体が減るので、割り切れないぶんは
			// 最後の列に寄せる（枚数が減っても列数は変えない）。
			col = StHelenaTableauCnt - 1
		}
		cr.tableau[col] = append(cr.tableau[col], tc)
	}
	// **制限はここで解ける。**規則の眼目で、解けないと後半に手が無くなる。
	cr.restrictionsActive = false
	cr.redealsRemaining--
	cr.moveCount++
	cr.appendLog("redeal", fmt.Sprintf("再配り (残り%d回)", cr.redealsRemaining), nil)
	cr.checkStHelenaStalemate()
	return nil
}

// GiveUp ギブアップ。
func (cr *StHelena) GiveUp() {
	if cr.phase == StHelenaPhasePlaying {
		cr.phase = StHelenaPhaseGameOver
		cr.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得する。優先度: タブロー→ファンデーション > タブロー→タブロー > 再配り。
func (cr *StHelena) GetHint() *StHelenaHint {
	if cr.phase != StHelenaPhasePlaying {
		return nil
	}
	for col := range StHelenaTableauCnt {
		if len(cr.tableau[col]) == 0 {
			continue
		}
		tc := cr.tableau[col][len(cr.tableau[col])-1]
		for fIdx := range StHelenaFoundationCnt {
			if cr.canPlaceOnFoundation(tc.Card, fIdx) {
				return &StHelenaHint{FromCol: col, ToZone: "foundation", ToCol: fIdx}
			}
		}
	}
	for fromCol := range StHelenaTableauCnt {
		fromCards := cr.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		card := fromCards[len(fromCards)-1].Card
		for toCol := range StHelenaTableauCnt {
			if toCol == fromCol {
				continue
			}
			if cr.canPlaceOnTableau(card, toCol) {
				return &StHelenaHint{FromCol: fromCol, ToZone: "tableau", ToCol: toCol}
			}
		}
	}
	if cr.redealsRemaining > 0 {
		return &StHelenaHint{FromCol: -1, ToZone: "", ToCol: -1, Redeal: true}
	}
	return nil
}

// AutoComplete タブロー上のカードを可能な限りファンデーションへ移動させる。
func (cr *StHelena) AutoComplete() error {
	if cr.phase != StHelenaPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	cr.takeSnapshot()
	for {
		moved := false
		for col := range StHelenaTableauCnt {
			if len(cr.tableau[col]) == 0 {
				continue
			}
			tc := cr.tableau[col][len(cr.tableau[col])-1]
			for fIdx := range StHelenaFoundationCnt {
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
	cr.checkStHelenaStalemate()
	return nil
}

// AllFaceUp StHelena では常にすべて表向き。インターフェース整合のため定義する。
func (cr *StHelena) AllFaceUp() bool { return true }

// --- State getters/setters ---

// GetPhase 現在のフェーズを返す。
func (cr *StHelena) GetPhase() StHelenaPhase { return cr.phase }

// SetPhase テスト用フェーズ設定。
func (cr *StHelena) SetPhase(phase StHelenaPhase) { cr.phase = phase }

// GetMoveCount 移動回数を返す。
func (cr *StHelena) GetMoveCount() int { return cr.moveCount }

// RestrictionsActive は初回の配りの制限がまだ効いているかを返す。
// 上下の列がそれぞれ片方の組札にしか送れない状態。
func (cr *StHelena) RestrictionsActive() bool { return cr.restrictionsActive }

// SetRestrictionsActive は制限の有無を設定する (テストと KV 復元用)。
func (cr *StHelena) SetRestrictionsActive(v bool) { cr.restrictionsActive = v }

// GetRedealsRemaining 残り再配り回数を返す。
func (cr *StHelena) GetRedealsRemaining() int { return cr.redealsRemaining }

// SetRedealsRemaining テスト用再配り回数設定。
func (cr *StHelena) SetRedealsRemaining(n int) { cr.redealsRemaining = n }

// GetTableau タブローを返す。
func (cr *StHelena) GetTableau() [StHelenaTableauCnt][]*StHelenaTableauCard { return cr.tableau }

// GetFoundation ファンデーションを返す。
func (cr *StHelena) GetFoundation() [StHelenaFoundationCnt][]*Card { return cr.foundation }

// GetGameEndFlag プレイ中でなくなったかを返す。
func (cr *StHelena) GetGameEndFlag() bool { return cr.phase != StHelenaPhasePlaying }

// IsStalemate 手詰まり状態を返す。
func (cr *StHelena) IsStalemate() bool { return cr.isStalemate }

// SetIsStalemate テスト用手詰まり設定。
func (cr *StHelena) SetIsStalemate(v bool) { cr.isStalemate = v }

// SetTableau テスト用タブロー設定。
func (cr *StHelena) SetTableau(tableau [StHelenaTableauCnt][]*StHelenaTableauCard) {
	cr.tableau = tableau
}

// SetFoundation テスト用ファンデーション設定。
func (cr *StHelena) SetFoundation(foundation [StHelenaFoundationCnt][]*Card) {
	cr.foundation = foundation
}

// Undo 直前の操作を取り消す。
func (cr *StHelena) Undo() error {
	if cr.phase != StHelenaPhasePlaying {
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
func (cr *StHelena) CanUndo() bool {
	return len(cr.history) > 0 && cr.phase == StHelenaPhasePlaying
}

// UndoToEscape 手詰まりから脱出するためのアンドゥ回数。手詰まりでなければ 0、脱出不可なら -1。
func (cr *StHelena) UndoToEscape() int {
	return undoToEscape(cr.isStalemate, cr.history, func(s *stHelenaSnapshot) bool { return s.isStalemate })
}

// UndoN n 回アンドゥする。
func (cr *StHelena) UndoN(n int) error {
	return undoN(cr, n)
}

// --- Private helpers ---

// canPlaceOnTableau タブロー toCol の最上段に card を置けるか判定。
// **スート不問で値差 ±1。K↔A の折り返しは無い。**空タブローは許可しない。
func (cr *StHelena) canPlaceOnTableau(card *Card, toCol int) bool {
	col := cr.tableau[toCol]
	if len(col) == 0 {
		return false
	}
	top := col[len(col)-1].Card
	// **スートは見ない。**クローン元のクレセントは同スートの ±1 なので、その
	// 述語を残すと異なるスートの手が全部消える。
	cv, tv := card.GetValue(), top.GetValue()
	// **K と A は繋がらない。**クレセントの A↔K 折り返しはこのゲームには無い。
	return cv == tv+1 || cv == tv-1
}

// canPlaceOnFoundation ファンデーション fIdx に card を置けるか判定。
// fIdx 0..3 は昇順 A→K、fIdx 4..7 は降順 K→A。スートはインデックスから一意に決まる。
func (cr *StHelena) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if fIdx < 0 || fIdx >= StHelenaFoundationCnt {
		return false
	}
	suit := stHelenaFoundationSuit(fIdx)
	if card.GetDesign() != suit {
		return false
	}
	pile := cr.foundation[fIdx]
	if len(pile) == 0 {
		if fIdx < StHelenaAscendingFoundationCnt {
			return card.GetValue() == 1
		}
		return card.GetValue() == CardValueMax
	}
	top := pile[len(pile)-1]
	if fIdx < StHelenaAscendingFoundationCnt {
		return card.GetValue() == top.GetValue()+1
	}
	return card.GetValue() == top.GetValue()-1
}

// stHelenaFoundationSuit ファンデーション ID から対応スートを返す。
func stHelenaFoundationSuit(fIdx int) int {
	suits := [...]int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	return suits[fIdx%StHelenaAscendingFoundationCnt]
}

// StHelenaFoundationSuit ファンデーション ID から対応スートを返す (公開ヘルパ)。
func StHelenaFoundationSuit(fIdx int) int { return stHelenaFoundationSuit(fIdx) }

// StHelenaIsAscendingFoundation 昇順ファンデーションかを返す。
func StHelenaIsAscendingFoundation(fIdx int) bool { return fIdx < StHelenaAscendingFoundationCnt }

// checkGameClear ゲームクリア判定。
func (cr *StHelena) checkGameClear() {
	total := 0
	for i := range StHelenaFoundationCnt {
		total += len(cr.foundation[i])
	}
	if total == CardCnt*2 {
		cr.phase = StHelenaPhaseGameClear
	}
}

// checkStHelenaStalemate 手詰まり判定。残り再配り 0 で合法手も無ければ stalemate。
func (cr *StHelena) checkStHelenaStalemate() {
	if cr.phase != StHelenaPhasePlaying {
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
func (cr *StHelena) hasAnyLegalMove() bool {
	for col := range StHelenaTableauCnt {
		if len(cr.tableau[col]) == 0 {
			continue
		}
		tc := cr.tableau[col][len(cr.tableau[col])-1]
		for fIdx := range StHelenaFoundationCnt {
			if cr.canPlaceOnFoundation(tc.Card, fIdx) {
				return true
			}
		}
	}
	for fromCol := range StHelenaTableauCnt {
		if len(cr.tableau[fromCol]) == 0 {
			continue
		}
		card := cr.tableau[fromCol][len(cr.tableau[fromCol])-1].Card
		for toCol := range StHelenaTableauCnt {
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
func (cr *StHelena) takeSnapshot() {
	snap := &stHelenaSnapshot{
		phase:              cr.phase,
		moveCount:          cr.moveCount,
		redealsRemaining:   cr.redealsRemaining,
		restrictionsActive: cr.restrictionsActive,
		isStalemate:        cr.isStalemate,
	}
	for i := range StHelenaTableauCnt {
		snap.tableau[i] = make([]*StHelenaTableauCard, len(cr.tableau[i]))
		for j, tc := range cr.tableau[i] {
			snap.tableau[i][j] = &StHelenaTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	for i := range StHelenaFoundationCnt {
		snap.foundation[i] = make([]*Card, len(cr.foundation[i]))
		copy(snap.foundation[i], cr.foundation[i])
	}
	cr.history = append(cr.history, snap)
}

// restoreSnapshot 状態を復元する。
func (cr *StHelena) restoreSnapshot(snap *stHelenaSnapshot) {
	cr.tableau = snap.tableau
	cr.foundation = snap.foundation
	cr.phase = snap.phase
	cr.moveCount = snap.moveCount
	cr.redealsRemaining = snap.redealsRemaining
	cr.restrictionsActive = snap.restrictionsActive
	cr.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加。
func (cr *StHelena) appendLog(actionType, detail string, cards []*Card) {
	cr.appendLogAt(cr.moveCount, 0, actionType, detail, cards)
}

// stHelenaJSON StHelena の永続化用ワイヤーフォーマット。
type stHelenaJSON struct {
	TrumpCards       *TrumpCards                                `json:"tc"`
	Tableau          [StHelenaTableauCnt][]*StHelenaTableauCard `json:"tb"`
	Foundation       [StHelenaFoundationCnt][]*Card             `json:"fd"`
	Phase            StHelenaPhase                              `json:"ps"`
	MoveCount        int                                        `json:"mc"`
	RedealsRemaining int                                        `json:"rd"`
	ActionLog        []*ActionLogEntry                          `json:"al"`
	IsStalemate      bool                                       `json:"sl"`
	History          []*stHelenaSnapshot                        `json:"hi,omitempty"`
}

// stHelenaSnapshotJSON 1 件のスナップショットのワイヤーフォーマット。
type stHelenaSnapshotJSON struct {
	Tableau          [StHelenaTableauCnt][]*StHelenaTableauCard `json:"tb"`
	Foundation       [StHelenaFoundationCnt][]*Card             `json:"fd"`
	Phase            StHelenaPhase                              `json:"ps"`
	MoveCount        int                                        `json:"mc"`
	RedealsRemaining int                                        `json:"rd"`
	IsStalemate      bool                                       `json:"sl"`
}

// MarshalJSON stHelenaSnapshot 用シリアライザ。
func (s *stHelenaSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(stHelenaSnapshotJSON{
		Tableau:          s.tableau,
		Foundation:       s.foundation,
		Phase:            s.phase,
		MoveCount:        s.moveCount,
		RedealsRemaining: s.redealsRemaining,
		IsStalemate:      s.isStalemate,
	})
}

// UnmarshalJSON stHelenaSnapshot 用デシリアライザ。
func (s *stHelenaSnapshot) UnmarshalJSON(data []byte) error {
	var j stHelenaSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > stHelenaMaxSliceLen {
			return fmt.Errorf("sthelena: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > stHelenaMaxSliceLen {
			return fmt.Errorf("sthelena: snapshot foundation pile exceeds maximum allowed size")
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

// MarshalJSON StHelena 用シリアライザ。
func (cr *StHelena) MarshalJSON() ([]byte, error) {
	return json.Marshal(stHelenaJSON{
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

// stHelenaMaxSliceLen 永続データの最大スライス長 (DoS 対策)。
const stHelenaMaxSliceLen = 1000

// UnmarshalJSON StHelena 用デシリアライザ。
func (cr *StHelena) UnmarshalJSON(data []byte) error {
	var j stHelenaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > stHelenaMaxSliceLen || len(j.History) > stHelenaMaxSliceLen {
		return fmt.Errorf("sthelena: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > stHelenaMaxSliceLen {
			return fmt.Errorf("sthelena: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > stHelenaMaxSliceLen {
			return fmt.Errorf("sthelena: foundation pile exceeds maximum allowed size")
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
		cr.history = make([]*stHelenaSnapshot, 0)
	}
	cr.isStalemate = j.IsStalemate
	return nil
}
