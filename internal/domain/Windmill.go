//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// WindmillPhase ウィンドミルのゲームフェーズ
type WindmillPhase int

// Windmillのフェーズ定数
const (
	// WindmillPhasePlaying プレイ中
	WindmillPhasePlaying WindmillPhase = iota
	// WindmillPhaseGameClear ゲームクリア
	WindmillPhaseGameClear
	// WindmillPhaseGameOver ゲームオーバー
	WindmillPhaseGameOver
)

// WindmillSailCnt 十字（帆）の枚数。上下左右に 2 枚ずつ。
const WindmillSailCnt = 8

// WindmillCornerCnt 四隅の K 基礎札の数
const WindmillCornerCnt = 4

// WindmillCenterTarget 中央基礎札の完成枚数（A→K を 4 周）
const WindmillCenterTarget = CardValueMax * 4

// WindmillCornerTarget 四隅の基礎札 1 つあたりの完成枚数（K→A）
const WindmillCornerTarget = CardValueMax

// WindmillTotalCards 使用する総枚数（52 枚 2 組）
const WindmillTotalCards = CardCnt * 2

// WindmillHint ウィンドミルのヒント
type WindmillHint struct {
	// FromZone 移動元 "sail" / "waste" / "corner" / "stock"
	FromZone string
	// FromIdx 移動元のインデックス（帆または四隅。それ以外は -1）
	FromIdx int
	// ToZone 移動先 "center" / "corner" / "waste"
	ToZone string
	// ToIdx 移動先のインデックス（-1 は特定の山を指さない）
	ToIdx int
}

// Windmill ウィンドミル（別名プロペラ）ゲームクラス。
//
// 52 枚 2 組（104 枚）の 1 人用ソリティア。中央に A を 1 枚置き、その上下左右に
// 2 枚ずつ計 **8 枚**の「帆」を十字に並べる。四隅の 4 枠は**最初は空**で、K が
// 出てきたときにそこへ置く。
//
// 中央の基礎札はスート無視の昇順で、**K まで行ったら A に戻って合計 52 枚**（A→K
// を 4 周）積む。四隅は K から降順に A まで 13 枚ずつ。52 + 13×4 = 104 枚で、
// すべて積み切ればクリア。
//
// 帆が空くと**即座に**捨て札（無ければ山札）から補充される。山札は 1 枚ずつ捨て札
// へめくり、めくり直しは無い。
//
// 救済手は「四隅の一番上を中央へ 1 枚移す」だけ。ただし移した直後は、次に中央へ
// 置く札が帆か捨て札から来るまで、もう一度この移動をすることはできない
// （IsTransferBlocked）。これが無いと降順の山を中央へ延々と流し込めてしまう。
//
// issue #4416 の仕様案とは 4 点異なり、いずれも実際の規則に合わせた:
//   - 十字は 4 枚ではなく **8 枚**（上下左右に 2 枚ずつ）
//   - 四隅の K は最初から置くのではなく、**出てきたときに**置く
//   - 中央は A→K を 1 周ではなく **4 周**（issue の「各ランク 8 枚まで」も誤り。
//     2 組で各ランク 8 枚あるが、その内訳は中央 4 枚 + 四隅 4 枚）
//   - 引き戻しは双方向ではなく **四隅 → 中央の一方向のみ**
type Windmill struct {
	trumpCards *TrumpCards
	center     []*Card
	corners    [WindmillCornerCnt][]*Card
	sails      [WindmillSailCnt]*Card
	stock      []*Card
	waste      []*Card
	// transferBlocked 直前に四隅→中央の引き戻しを行ったか。
	transferBlocked bool
	phase           WindmillPhase
	moveCount       int
	actionLog       []*ActionLogEntry
	history         []*windmillSnapshot
	isStalemate     bool
}

// windmillSnapshot アンドゥ用スナップショット
type windmillSnapshot struct {
	center          []*Card
	corners         [WindmillCornerCnt][]*Card
	sails           [WindmillSailCnt]*Card
	stock           []*Card
	waste           []*Card
	transferBlocked bool
	phase           WindmillPhase
	moveCount       int
	isStalemate     bool
}

// NewWindmill コンストラクタ
func NewWindmill(trumpCards *TrumpCards) *Windmill {
	return &Windmill{trumpCards: trumpCards}
}

// NewDefaultWindmill returns Windmill with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultWindmill() *Windmill {
	return NewWindmill(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (w *Windmill) Reset() {
	w.trumpCards.Shuffle()
	w.phase = WindmillPhasePlaying
	w.moveCount = 0
	w.actionLog = nil
	w.history = nil
	w.isStalemate = false
	w.transferBlocked = false
	w.center = nil
	w.stock = nil
	w.waste = nil

	for i := range WindmillCornerCnt {
		w.corners[i] = nil
	}
	for i := range WindmillSailCnt {
		w.sails[i] = nil
	}

	// 中央には A を 1 枚置いて始める。山を引きながら最初に出た A を中央に据え、
	// それ以外はそのまま配り札の列に戻す（引いた順を保つので偏りは出ない）。
	remaining := make([]*Card, 0, WindmillTotalCards)
	for {
		card := w.trumpCards.DrawCard()
		if card == nil {
			break
		}
		if len(w.center) == 0 && card.GetValue() == 1 {
			w.center = append(w.center, card)
			continue
		}
		remaining = append(remaining, card)
	}

	for i := range WindmillSailCnt {
		if i < len(remaining) {
			w.sails[i] = remaining[i]
		}
	}
	if len(remaining) > WindmillSailCnt {
		w.stock = append(w.stock, remaining[WindmillSailCnt:]...)
	}

	w.checkStalemate()
}

// Draw 山札から 1 枚めくって捨て札へ送る
func (w *Windmill) Draw() error {
	if err := w.requirePlaying(); err != nil {
		return err
	}
	if len(w.stock) == 0 {
		return errors.New("stock is empty")
	}
	w.takeSnapshot()
	card := w.stock[0]
	w.stock = w.stock[1:]
	w.waste = append(w.waste, card)
	w.afterMove("draw", "山札から1枚めくった", card)
	return nil
}

// MoveSailToCenter 帆の札を中央基礎札へ送る
func (w *Windmill) MoveSailToCenter(sailIdx int) error {
	if err := w.requirePlaying(); err != nil {
		return err
	}
	if err := validWindmillSail(sailIdx); err != nil {
		return err
	}
	card := w.sails[sailIdx]
	if card == nil {
		return fmt.Errorf("sail %d is empty", sailIdx)
	}
	if !w.canPlaceOnCenter(card) {
		return errors.New("card cannot be placed on the center foundation")
	}
	w.takeSnapshot()
	w.sails[sailIdx] = nil
	w.pushCenter(card)
	w.refillSails()
	w.afterMove("move", fmt.Sprintf("帆%d→中央基礎", sailIdx), card)
	return nil
}

// MoveSailToCorner 帆の札を四隅の基礎札へ送る
func (w *Windmill) MoveSailToCorner(sailIdx, cornerIdx int) error {
	if err := w.requirePlaying(); err != nil {
		return err
	}
	if err := validWindmillSail(sailIdx); err != nil {
		return err
	}
	if err := validWindmillCorner(cornerIdx); err != nil {
		return err
	}
	card := w.sails[sailIdx]
	if card == nil {
		return fmt.Errorf("sail %d is empty", sailIdx)
	}
	if !w.canPlaceOnCorner(card, cornerIdx) {
		return errors.New("card cannot be placed on that corner foundation")
	}
	w.takeSnapshot()
	w.sails[sailIdx] = nil
	w.corners[cornerIdx] = append(w.corners[cornerIdx], card)
	w.refillSails()
	w.afterMove("move", fmt.Sprintf("帆%d→四隅基礎%d", sailIdx, cornerIdx), card)
	return nil
}

// MoveWasteToCenter 捨て札の一番上を中央基礎札へ送る
func (w *Windmill) MoveWasteToCenter() error {
	if err := w.requirePlaying(); err != nil {
		return err
	}
	card := w.wasteTop()
	if card == nil {
		return errors.New("waste is empty")
	}
	if !w.canPlaceOnCenter(card) {
		return errors.New("card cannot be placed on the center foundation")
	}
	w.takeSnapshot()
	w.popWaste()
	w.pushCenter(card)
	w.afterMove("move", "捨て札→中央基礎", card)
	return nil
}

// MoveWasteToCorner 捨て札の一番上を四隅の基礎札へ送る
func (w *Windmill) MoveWasteToCorner(cornerIdx int) error {
	if err := w.requirePlaying(); err != nil {
		return err
	}
	if err := validWindmillCorner(cornerIdx); err != nil {
		return err
	}
	card := w.wasteTop()
	if card == nil {
		return errors.New("waste is empty")
	}
	if !w.canPlaceOnCorner(card, cornerIdx) {
		return errors.New("card cannot be placed on that corner foundation")
	}
	w.takeSnapshot()
	w.popWaste()
	w.corners[cornerIdx] = append(w.corners[cornerIdx], card)
	w.afterMove("move", fmt.Sprintf("捨て札→四隅基礎%d", cornerIdx), card)
	return nil
}

// MoveCornerToCenter 四隅の一番上を中央基礎札へ引き戻す（救済手）。
//
// 直前にこの手を打っていると、次に中央へ置く札が帆か捨て札から来るまで拒否する。
func (w *Windmill) MoveCornerToCenter(cornerIdx int) error {
	if err := w.requirePlaying(); err != nil {
		return err
	}
	if err := validWindmillCorner(cornerIdx); err != nil {
		return err
	}
	if w.transferBlocked {
		return errors.New("the next center card must come from a sail or the waste")
	}
	card := w.cornerTop(cornerIdx)
	if card == nil {
		return fmt.Errorf("corner %d is empty", cornerIdx)
	}
	if !w.canPlaceOnCenter(card) {
		return errors.New("card cannot be placed on the center foundation")
	}
	w.takeSnapshot()
	w.corners[cornerIdx] = w.corners[cornerIdx][:len(w.corners[cornerIdx])-1]
	w.center = append(w.center, card)
	// pushCenter ではなくここで直接立てる。帆・捨て札から置いたときだけ解除される。
	w.transferBlocked = true
	w.afterMove("move", fmt.Sprintf("四隅基礎%d→中央基礎", cornerIdx), card)
	return nil
}

// GiveUp ギブアップ
func (w *Windmill) GiveUp() {
	if w.phase == WindmillPhasePlaying {
		w.phase = WindmillPhaseGameOver
		w.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。基礎札 → 引き戻し → 山札の順。手詰まり判定も兼ねる。
func (w *Windmill) GetHint() *WindmillHint {
	if w.phase != WindmillPhasePlaying {
		return nil
	}
	if h := w.foundationHint(); h != nil {
		return h
	}
	// 引き戻しは盤面をかき混ぜる救済手なので、通常の手が尽きてから提示する。
	if !w.transferBlocked {
		for i := range WindmillCornerCnt {
			if card := w.cornerTop(i); card != nil && w.canPlaceOnCenter(card) {
				return &WindmillHint{FromZone: "corner", FromIdx: i, ToZone: "center", ToIdx: -1}
			}
		}
	}
	if len(w.stock) > 0 {
		return &WindmillHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1}
	}
	return nil
}

// foundationHint 帆・捨て札から基礎札へ送れる手を 1 つ返す（オートコンプリート用）。
//
// 引き戻しを混ぜてはいけない — オートコンプリートは基礎札に上げるだけの操作で、
// 四隅を崩すという戦略的な判断を勝手に行ってはならない。
func (w *Windmill) foundationHint() *WindmillHint {
	if w.phase != WindmillPhasePlaying {
		return nil
	}
	for i := range WindmillSailCnt {
		card := w.sails[i]
		if card == nil {
			continue
		}
		if w.canPlaceOnCenter(card) {
			return &WindmillHint{FromZone: "sail", FromIdx: i, ToZone: "center", ToIdx: -1}
		}
		if c := w.findCorner(card); c >= 0 {
			return &WindmillHint{FromZone: "sail", FromIdx: i, ToZone: "corner", ToIdx: c}
		}
	}
	if card := w.wasteTop(); card != nil {
		if w.canPlaceOnCenter(card) {
			return &WindmillHint{FromZone: "waste", FromIdx: -1, ToZone: "center", ToIdx: -1}
		}
		if c := w.findCorner(card); c >= 0 {
			return &WindmillHint{FromZone: "waste", FromIdx: -1, ToZone: "corner", ToIdx: c}
		}
	}
	return nil
}

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (w *Windmill) AutoComplete() error {
	if w.phase != WindmillPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	moved := false
	for {
		h := w.foundationHint()
		if h == nil {
			break
		}
		var err error
		switch {
		case h.FromZone == "sail" && h.ToZone == "center":
			err = w.MoveSailToCenter(h.FromIdx)
		case h.FromZone == "sail":
			err = w.MoveSailToCorner(h.FromIdx, h.ToIdx)
		case h.ToZone == "center":
			err = w.MoveWasteToCenter()
		default:
			err = w.MoveWasteToCorner(h.ToIdx)
		}
		if err != nil {
			return err
		}
		moved = true
	}
	if !moved {
		return errors.New("no card can be auto-completed")
	}
	return nil
}

// Undo 直前の 1 手を取り消す
func (w *Windmill) Undo() error {
	if len(w.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := w.history[len(w.history)-1]
	w.history = w.history[:len(w.history)-1]
	w.center = snap.center
	w.corners = snap.corners
	w.sails = snap.sails
	w.stock = snap.stock
	w.waste = snap.waste
	w.transferBlocked = snap.transferBlocked
	w.phase = snap.phase
	w.moveCount = snap.moveCount
	w.isStalemate = snap.isStalemate
	return nil
}

// CanUndo アンドゥ可能か
func (w *Windmill) CanUndo() bool { return len(w.history) > 0 }

// UndoN n 手戻す
func (w *Windmill) UndoN(n int) error {
	if n <= 0 {
		return errors.New("n must be positive")
	}
	if n > len(w.history) {
		return errors.New("not enough history")
	}
	for range n {
		if err := w.Undo(); err != nil {
			return err
		}
	}
	return nil
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (w *Windmill) UndoToEscape() int {
	if !w.isStalemate {
		return 0
	}
	for i := len(w.history) - 1; i >= 0; i-- {
		if !w.history[i].isStalemate {
			return len(w.history) - i
		}
	}
	return -1
}

// AllFaceUp 常に全札が表向き
func (w *Windmill) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (w *Windmill) GetPhase() WindmillPhase { return w.phase }

// GetMoveCount 手数取得
func (w *Windmill) GetMoveCount() int { return w.moveCount }

// GetStockCount 山札の残り枚数
func (w *Windmill) GetStockCount() int { return len(w.stock) }

// GetWaste 捨て札を取得
func (w *Windmill) GetWaste() []*Card { return w.waste }

// GetSails 十字（帆）の 8 枠を取得。空き枠は nil。
func (w *Windmill) GetSails() [WindmillSailCnt]*Card { return w.sails }

// GetCenter 中央基礎札を取得
func (w *Windmill) GetCenter() []*Card { return w.center }

// GetCorners 四隅の基礎札を取得
func (w *Windmill) GetCorners() [WindmillCornerCnt][]*Card { return w.corners }

// IsTransferBlocked 四隅→中央の引き戻しが今は禁じられているか。
// 直前にその手を打った直後だけ真になる。
func (w *Windmill) IsTransferBlocked() bool { return w.transferBlocked }

// GetActionLog 棋譜取得
func (w *Windmill) GetActionLog() []*ActionLogEntry { return w.actionLog }

// GetGameEndFlag ゲーム終了フラグ
func (w *Windmill) GetGameEndFlag() bool { return w.phase != WindmillPhasePlaying }

// IsStalemate 手詰まりか
func (w *Windmill) IsStalemate() bool { return w.isStalemate }

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (w *Windmill) requirePlaying() error {
	if w.phase != WindmillPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

// validWindmillSail 帆のインデックスを検証する
func validWindmillSail(idx int) error {
	if idx < 0 || idx >= WindmillSailCnt {
		return fmt.Errorf("invalid sail index: %d", idx)
	}
	return nil
}

// validWindmillCorner 四隅のインデックスを検証する
func validWindmillCorner(idx int) error {
	if idx < 0 || idx >= WindmillCornerCnt {
		return fmt.Errorf("invalid corner index: %d", idx)
	}
	return nil
}

// windmillNextCenterRank 中央基礎札が次に受け付けるランク。K の次は A に戻る。
func windmillNextCenterRank(n int) int {
	return n%CardValueMax + 1
}

// canPlaceOnCenter 中央基礎札に置けるか（スート無視の昇順、K の次は A）
func (w *Windmill) canPlaceOnCenter(card *Card) bool {
	if card == nil || len(w.center) >= WindmillCenterTarget {
		return false
	}
	return card.GetValue() == windmillNextCenterRank(len(w.center))
}

// canPlaceOnCorner 四隅の基礎札に置けるか（空なら K、以降はスート無視の降順で A まで）
func (w *Windmill) canPlaceOnCorner(card *Card, cornerIdx int) bool {
	if card == nil {
		return false
	}
	pile := w.corners[cornerIdx]
	if len(pile) == 0 {
		return card.GetValue() == CardValueMax
	}
	if len(pile) >= WindmillCornerTarget {
		return false
	}
	return card.GetValue() == pile[len(pile)-1].GetValue()-1
}

// findCorner 置ける四隅を探す（見つからなければ -1）
func (w *Windmill) findCorner(card *Card) int {
	for i := range WindmillCornerCnt {
		if w.canPlaceOnCorner(card, i) {
			return i
		}
	}
	return -1
}

// cornerTop 四隅の一番上（空なら nil）
func (w *Windmill) cornerTop(cornerIdx int) *Card {
	pile := w.corners[cornerIdx]
	if len(pile) == 0 {
		return nil
	}
	return pile[len(pile)-1]
}

// wasteTop 捨て札の一番上（空なら nil）
func (w *Windmill) wasteTop() *Card {
	if len(w.waste) == 0 {
		return nil
	}
	return w.waste[len(w.waste)-1]
}

// popWaste 捨て札の一番上を取り除く
func (w *Windmill) popWaste() {
	if len(w.waste) > 0 {
		w.waste = w.waste[:len(w.waste)-1]
	}
}

// pushCenter 帆・捨て札から中央へ置く。引き戻しの禁止を解除するのはこの経路だけ。
func (w *Windmill) pushCenter(card *Card) {
	w.center = append(w.center, card)
	w.transferBlocked = false
}

// refillSails 空いた帆を捨て札（無ければ山札）から即座に補充する
func (w *Windmill) refillSails() {
	for i := range WindmillSailCnt {
		if w.sails[i] != nil {
			continue
		}
		switch {
		case len(w.waste) > 0:
			w.sails[i] = w.waste[len(w.waste)-1]
			w.waste = w.waste[:len(w.waste)-1]
		case len(w.stock) > 0:
			w.sails[i] = w.stock[0]
			w.stock = w.stock[1:]
		default:
			return
		}
	}
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (w *Windmill) afterMove(actionType, detail string, card *Card) {
	w.moveCount++
	var cards []*Card
	if card != nil {
		cards = []*Card{card}
	}
	w.appendLog(actionType, detail, cards)
	w.checkGameClear()
	w.checkStalemate()
}

// checkGameClear 中央 52 枚と四隅 13 枚×4 がすべて揃ったか
func (w *Windmill) checkGameClear() {
	if len(w.center) != WindmillCenterTarget {
		return
	}
	for i := range WindmillCornerCnt {
		if len(w.corners[i]) != WindmillCornerTarget {
			return
		}
	}
	w.phase = WindmillPhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
// GetHint は基礎札・引き戻し・山札のすべてを見るので、「ヒントが無い」と
// 「手詰まり」は同じ条件になる。
func (w *Windmill) checkStalemate() {
	if w.phase != WindmillPhasePlaying {
		return
	}
	w.isStalemate = w.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (w *Windmill) takeSnapshot() {
	snap := &windmillSnapshot{
		center:          append([]*Card(nil), w.center...),
		sails:           w.sails,
		stock:           append([]*Card(nil), w.stock...),
		waste:           append([]*Card(nil), w.waste...),
		transferBlocked: w.transferBlocked,
		phase:           w.phase,
		moveCount:       w.moveCount,
		isStalemate:     w.isStalemate,
	}
	for i := range WindmillCornerCnt {
		snap.corners[i] = append([]*Card(nil), w.corners[i]...)
	}
	w.history = append(w.history, snap)
}

// appendLog 棋譜エントリを追加
func (w *Windmill) appendLog(actionType, detail string, cards []*Card) {
	w.actionLog = append(w.actionLog, &ActionLogEntry{
		TurnNumber: w.moveCount,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// windmillJSON is the JSON wire format for Windmill.
type windmillJSON struct {
	TrumpCards      *TrumpCards                `json:"tc"`
	Center          []*Card                    `json:"ce"`
	Corners         [WindmillCornerCnt][]*Card `json:"co"`
	Sails           [WindmillSailCnt]*Card     `json:"sa"`
	Stock           []*Card                    `json:"st"`
	Waste           []*Card                    `json:"ws"`
	TransferBlocked bool                       `json:"tb"`
	Phase           WindmillPhase              `json:"ps"`
	MoveCount       int                        `json:"mc"`
	ActionLog       []*ActionLogEntry          `json:"al"`
	IsStalemate     bool                       `json:"sl"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (w *Windmill) MarshalJSON() ([]byte, error) {
	return json.Marshal(&windmillJSON{
		TrumpCards:      w.trumpCards,
		Center:          w.center,
		Corners:         w.corners,
		Sails:           w.sails,
		Stock:           w.stock,
		Waste:           w.waste,
		TransferBlocked: w.transferBlocked,
		Phase:           w.phase,
		MoveCount:       w.moveCount,
		ActionLog:       w.actionLog,
		IsStalemate:     w.isStalemate,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (w *Windmill) UnmarshalJSON(data []byte) error {
	var j windmillJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < WindmillPhasePlaying || j.Phase > WindmillPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	if len(j.Center) > WindmillCenterTarget {
		return fmt.Errorf("center holds %d cards", len(j.Center))
	}
	if len(j.Stock) > WindmillTotalCards || len(j.Waste) > WindmillTotalCards {
		return fmt.Errorf("stock/waste too large: %d/%d", len(j.Stock), len(j.Waste))
	}
	for i := range WindmillCornerCnt {
		if len(j.Corners[i]) > WindmillCornerTarget {
			return fmt.Errorf("corner %d holds %d cards", i, len(j.Corners[i]))
		}
	}
	if j.TrumpCards != nil {
		w.trumpCards = j.TrumpCards
	}
	w.center = j.Center
	w.corners = j.Corners
	w.sails = j.Sails
	w.stock = j.Stock
	w.waste = j.Waste
	w.transferBlocked = j.TransferBlocked
	w.phase = j.Phase
	w.moveCount = j.MoveCount
	w.actionLog = j.ActionLog
	w.isStalemate = j.IsStalemate
	return nil
}
