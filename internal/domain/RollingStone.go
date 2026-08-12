//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// RollingStonePhase はローリングストーンのフェーズ。
type RollingStonePhase int

const (
	// RollingStonePhasePlay はプレイ中。
	RollingStonePhasePlay RollingStonePhase = iota
	// RollingStonePhaseGameEnd は終局。
	RollingStonePhaseGameEnd
)

// RollingStoneHandSize は 1 人あたりの手札枚数。**人数に依らず 8 枚固定。**
//
// **デッキのほうを人数に合わせます。** issue の「6 人なら 32 枚」は 32 が 6 で
// 割り切れないので成立しません——32 枚は 4 人のときの値です。
const RollingStoneHandSize = 8

// RollingStoneSuitCnt はスート数。
const RollingStoneSuitCnt = 4

// rollingStoneMaxSliceLen は復元時に受け付ける棋譜の上限。
//
// **他ゲームの 1000 では足りない。** このゲームは膠着上限まで走ることがあり、
// 1 トリックあたり最大「在席人数ぶんのプレイ + トリック解決または引き取り」の
// 棋譜が出ます。上限は
//
//	500 トリック × (6 人 + 1) + 開始 1 + 上がり 6 + 結果 1 = 3508
//
// なので 4000 を取ります。**1000 のままだと、長い局を保存して読み直すと
// 自分の codec が「大きすぎる」と言って正当な対局を拒否しました**
// ——実際の局を毎手ごとに往復させるテストが見つけました。
const rollingStoneMaxSliceLen = 4000

// RollingStoneStalemateTricks は引き分け（膠着）と判定するまでのトリック数。
//
// **このゲームは放っておくと終わらない。** 札が場から抜けるのは全員フォローできた
// トリックだけで、誰かがフォローできなければ全部手札に戻るので、同じ札が延々と
// 循環します。実測（各人数 400 局、1 局 30 万手で打ち切り）:
//
//	4 人  決着 190/400  トリック数 中央値 20 / p95 31  / 最大 39
//	5 人  決着 359/400  中央値 41 / p95 141 / 最大 303
//	6 人  決着 366/400  中央値 48 / p95 131 / 最大 262
//
// **決着する局の最大が 303 トリック**なので、500 で切れば「本来決着する局を
// 途中で止める」ことはまず起きません。上限に達したら**手札のいちばん少ない席**の
// 勝ちにします（同数なら席順で、決定的に）。
const RollingStoneStalemateTricks = 500

// RollingStoneDeckSize は人数に対するデッキ枚数を返す。
//
// **人数 × 8。** 4 人 32 枚 / 5 人 40 枚 / 6 人 48 枚で、いずれも 4 スートに
// 均等（8 / 10 / 12 ランク）に割れます。
func RollingStoneDeckSize(playerCnt int) int { return playerCnt * RollingStoneHandSize }

// RollingStoneLowestRank は使うデッキの最も低いランクを返す。
//
// **A を最強として上から詰めます。** 6 人なら 3 以上（12 ランク）、
// 5 人なら 5 以上（10 ランク）、4 人なら 7 以上（8 ランク）。
func RollingStoneLowestRank(playerCnt int) int {
	ranks := RollingStoneDeckSize(playerCnt) / RollingStoneSuitCnt
	// A(1) は最強として別扱いなので、2..13 のうち上位 ranks-1 枚 + A。
	return 13 - (ranks - 2)
}

// rollingStoneRank は札の強さ。**A が最強。**
func rollingStoneRank(c *Card) int {
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// RollingStoneHint はローリングストーンの助言。
type RollingStoneHint struct {
	CardIndex *int
	Reason    string
}

// RollingStone はローリングストーンのゲーム。
//
// **勝利条件が逆さま。** トリックを取るのは得ではなく、フォローできないと
// 場札を全部引き取らされます。**先に手札を無くした人が勝ち。**
type RollingStone struct {
	trumpCards *TrumpCards
	players    []*RollingStonePlayer
	config     RollingStoneConfig
	phase      RollingStonePhase
	actionLogBase

	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	trickNumber      int

	// finishedCnt は上がった人数。
	finishedCnt int
	gameEndFlag bool
	// winnerIdx は最初に上がった席（-1 = 未確定）。
	winnerIdx int
	// lastPickupIdx は直前に引き取った席（-1 = 無し）。表示用。
	lastPickupIdx int
	// discarded は場から抜けた札の枚数。
	//
	// **全員フォローできたトリックだけが場から抜けます。** 引き取りは席の間で
	// 動かすだけなので減りません。復元時に「手札 + 場札 + 抜けた枚数 = デッキ」で
	// 突き合わせるために持っています。
	discarded int
}

// NewRollingStone はコンストラクタ。
func NewRollingStone(players []*RollingStonePlayer, config RollingStoneConfig) *RollingStone {
	if config.Validate() != nil {
		config = DefaultRollingStoneConfig()
	}
	if len(players) != config.PlayerCnt {
		players = newRollingStoneSeats(config.PlayerCnt)
	}
	return &RollingStone{players: players, config: config, winnerIdx: -1, lastPickupIdx: -1}
}

// newRollingStoneSeats は標準の席（人間 1 + CPU）を返す。
func newRollingStoneSeats(n int) []*RollingStonePlayer {
	seats := make([]*RollingStonePlayer, 0, n)
	for i := range n {
		seats = append(seats, NewRollingStonePlayer(i == 0))
	}
	return seats
}

// NewDefaultRollingStone は標準セットアップを返す。
func NewDefaultRollingStone() *RollingStone {
	cfg := DefaultRollingStoneConfig()
	return NewRollingStone(newRollingStoneSeats(cfg.PlayerCnt), cfg)
}

// Reset はゲームを初期化する。
func (r *RollingStone) Reset() {
	for _, p := range r.players {
		p.ResetGame()
	}
	r.phase = RollingStonePhasePlay
	r.currentTrick = nil
	r.trickNumber = 0
	r.finishedCnt = 0
	r.discarded = 0
	r.gameEndFlag = false
	r.winnerIdx = -1
	r.lastPickupIdx = -1
	r.actionLog = nil

	// **低いランクを抜いて人数ちょうどのデッキを作る。** 新しい Card 型は要りません。
	r.trumpCards = NewTrumpCards(0)
	r.trumpCards.Shuffle()
	lowest := RollingStoneLowestRank(r.config.PlayerCnt)
	dealt := 0
	want := RollingStoneDeckSize(r.config.PlayerCnt)
	for dealt < want {
		c := r.trumpCards.DrawCard()
		if c == nil {
			break
		}
		// A は最強として常に使う。それ以外は lowest 未満を捨てる。
		if c.GetValue() != 1 && c.GetValue() < lowest {
			continue
		}
		r.players[dealt%r.config.PlayerCnt].AddCard(c)
		dealt++
	}
	r.sortAllHands()

	r.leadPlayerIdx = 0
	r.currentPlayerIdx = 0
	r.addLog(-1, "start", fmt.Sprintf("ローリングストーンを開始しました（%d 人、%d 枚）",
		r.config.PlayerCnt, want), nil)
}

// sortAllHands は手札をスート・ランク順に整える。
func (r *RollingStone) sortAllHands() {
	for _, p := range r.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return rollingStoneRank(ci) < rollingStoneRank(cj)
		})
	}
}

// leadSuit はこのトリックのリードスートを返す（誰も出していなければ 0）。
func (r *RollingStone) leadSuit() int {
	if len(r.currentTrick) == 0 {
		return 0
	}
	return r.currentTrick[0].Card.GetDesign()
}

// GetValidPlayIndices はプレイ可能な手札インデックスを返す。
//
// **フォローできる札があるなら、それしか出せません。** 無いときは空を返します
// ——このゲームでは「出せない」が引き取りを意味するので、捨て札にはなりません。
func (r *RollingStone) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= r.config.PlayerCnt {
		return nil
	}
	p := r.players[playerIdx]
	lead := r.leadSuit()
	out := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if lead == 0 || p.GetCard(i).GetDesign() == lead {
			out = append(out, i)
		}
	}
	return out
}

// MustPickUp は playerIdx がフォローできず引き取るしかないかを返す。
func (r *RollingStone) MustPickUp(playerIdx int) bool {
	if r.gameEndFlag || r.phase != RollingStonePhasePlay {
		return false
	}
	if len(r.currentTrick) == 0 {
		return false
	}
	return len(r.GetValidPlayIndices(playerIdx)) == 0
}

// IsHumanTurn は人間の手番かを返す。
func (r *RollingStone) IsHumanTurn() bool {
	if r.gameEndFlag || r.phase != RollingStonePhasePlay {
		return false
	}
	return r.players[r.currentPlayerIdx].GetIsHuman()
}

// PlayerPlay は人間が札を出す。
func (r *RollingStone) PlayerPlay(cardIndex int) error {
	if !r.IsHumanTurn() {
		return ErrNotHumanTurn
	}
	return r.play(r.currentPlayerIdx, cardIndex)
}

// PlayerPickUp は人間が場札を引き取る。
func (r *RollingStone) PlayerPickUp() error {
	if !r.IsHumanTurn() {
		return ErrNotHumanTurn
	}
	return r.pickUp(r.currentPlayerIdx)
}

// CpuPlay は CPU が 1 手打つ。
func (r *RollingStone) CpuPlay() {
	if r.gameEndFlag || r.phase != RollingStonePhasePlay || r.IsHumanTurn() {
		return
	}
	idx := r.currentPlayerIdx
	if r.MustPickUp(idx) {
		_ = r.pickUp(idx)
		return
	}
	_ = r.play(idx, r.chooseCpuCard(idx))
}

// play は 1 枚出す共通処理。
func (r *RollingStone) play(playerIdx, cardIndex int) error {
	if r.gameEndFlag {
		return ErrGameEnded
	}
	if r.phase != RollingStonePhasePlay {
		return ErrWrongPhase
	}
	if playerIdx != r.currentPlayerIdx {
		return ErrNotHumanTurn
	}
	valid := r.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return errors.New("cannot follow: you must pick the trick up")
	}
	if !rollingStoneContains(valid, cardIndex) {
		return errors.New("must follow the led suit")
	}

	card := r.players[playerIdx].RemoveCard(cardIndex)
	if card == nil {
		return NewDomainError(ErrInvalidCard, "カードがありません")
	}
	r.currentTrick = append(r.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	r.addLog(playerIdx, "play", "カードを出しました", []*Card{card})

	// **出し切ったら上がり。** 途中でも抜けます。
	r.checkFinished(playerIdx)

	if r.gameEndFlag {
		return nil
	}
	if r.trickComplete() {
		r.resolveTrick()
		return nil
	}
	r.advanceTurn()
	return nil
}

// trickComplete はこのトリックで打つべき席が全員打ったかを返す。
//
// **在席数と枚数を比べない（レビュー指摘 PR #5316）。** トリックの途中で誰かが
// 上がると在席数がその場で縮み、まだ出していない最後の席を飛ばして解決して
// しまいます。上がった席はこのトリックに札を出しているので、「出していない
// 在席者がいるか」で見れば数の増減に影響されません。
func (r *RollingStone) trickComplete() bool {
	if len(r.currentTrick) == 0 {
		return false
	}
	played := make(map[int]bool, len(r.currentTrick))
	for _, tc := range r.currentTrick {
		played[tc.PlayerIdx] = true
	}
	for i, p := range r.players {
		if !p.HasFinished() && !played[i] {
			return false
		}
	}
	return true
}

// pickUp は playerIdx が場札を全部引き取る。
//
// **このゲームの罰則は「手札が増える」こと。** トリックを取るのが得ではないので、
// 引き取りは勝敗そのものを遠ざけます。
func (r *RollingStone) pickUp(playerIdx int) error {
	if r.gameEndFlag {
		return ErrGameEnded
	}
	if r.phase != RollingStonePhasePlay {
		return ErrWrongPhase
	}
	if playerIdx != r.currentPlayerIdx {
		return ErrNotHumanTurn
	}
	if !r.MustPickUp(playerIdx) {
		return errors.New("you can still follow: play a card")
	}

	taken := make([]*Card, 0, len(r.currentTrick))
	for _, tc := range r.currentTrick {
		taken = append(taken, tc.Card)
		r.players[playerIdx].AddCard(tc.Card)
	}
	r.currentTrick = nil
	r.players[playerIdx].AddPickup()
	r.lastPickupIdx = playerIdx
	r.sortAllHands()
	r.addLog(playerIdx, "pickup",
		fmt.Sprintf("フォローできず %d 枚を引き取りました", len(taken)), taken)

	// **引き取った人が次のリード。**
	r.trickNumber++
	r.leadPlayerIdx = playerIdx
	r.currentPlayerIdx = playerIdx
	if r.gameEndFlag {
		return nil
	}
	r.checkStalemate()
	return nil
}

// checkStalemate は膠着上限に達したかを見て、達していれば決着させる。
//
// **引き取りも 1 トリックとして数えます。** 数えないと、延々と引き取り合う局が
// 上限に達しません——止まらない局はまさにその形です。
func (r *RollingStone) checkStalemate() bool {
	if r.gameEndFlag || r.trickNumber < RollingStoneStalemateTricks {
		return false
	}
	best := 0
	for i := 1; i < r.config.PlayerCnt; i++ {
		// 厳密な `<` なので、同数なら先の席が残る（決定的）。
		if r.players[i].GetCardsSize() < r.players[best].GetCardsSize() {
			best = i
		}
	}
	r.phase = RollingStonePhaseGameEnd
	r.gameEndFlag = true
	r.winnerIdx = best
	r.addLog(best, "stalemate",
		fmt.Sprintf("%d トリックで決着せず。手札のいちばん少ない席の勝ちとします（%d 枚）",
			r.trickNumber, r.players[best].GetCardsSize()), nil)
	return true
}

// resolveTrick はトリックを解決する。**取っても得点にはなりません。**
func (r *RollingStone) resolveTrick() {
	winner := r.trickWinner()
	cards := make([]*Card, 0, len(r.currentTrick))
	for _, tc := range r.currentTrick {
		cards = append(cards, tc.Card)
	}
	r.addLog(winner, "trick", "トリックを取りました（得点にはなりません）", cards)
	r.discarded += len(cards)
	r.currentTrick = nil
	r.trickNumber++
	r.lastPickupIdx = -1

	if r.checkGameEnd() {
		return
	}
	if r.checkStalemate() {
		return
	}
	// **取った人が次のリード。** 既に上がっていれば次の在席者へ。
	r.leadPlayerIdx = winner
	if r.players[winner].HasFinished() {
		r.leadPlayerIdx = r.nextActive(winner)
	}
	r.currentPlayerIdx = r.leadPlayerIdx
}

// trickWinner はこのトリックの勝者を返す。**切り札はありません。**
func (r *RollingStone) trickWinner() int {
	if len(r.currentTrick) == 0 {
		return r.leadPlayerIdx
	}
	lead := r.currentTrick[0].Card.GetDesign()
	best := r.currentTrick[0]
	for _, tc := range r.currentTrick[1:] {
		if tc.Card.GetDesign() != lead {
			continue
		}
		if rollingStoneRank(tc.Card) > rollingStoneRank(best.Card) {
			best = tc
		}
	}
	return best.PlayerIdx
}

// rollingStoneContains は xs が v を含むかを返す。
func rollingStoneContains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// activeSeatCnt はまだ上がっていない席の数を返す。
func (r *RollingStone) activeSeatCnt() int {
	n := 0
	for _, p := range r.players {
		if !p.HasFinished() {
			n++
		}
	}
	return n
}

// nextActive は i の次の、まだ上がっていない席を返す。
func (r *RollingStone) nextActive(i int) int {
	for step := 1; step <= r.config.PlayerCnt; step++ {
		next := (i + step) % r.config.PlayerCnt
		if !r.players[next].HasFinished() {
			return next
		}
	}
	return i
}

// advanceTurn は次の在席者へ手番を回す。
func (r *RollingStone) advanceTurn() { r.currentPlayerIdx = r.nextActive(r.currentPlayerIdx) }

// checkFinished は手札を出し切った席に順位を付ける。
func (r *RollingStone) checkFinished(playerIdx int) {
	p := r.players[playerIdx]
	if p.HasFinished() || p.GetCardsSize() > 0 {
		return
	}
	r.finishedCnt++
	p.SetFinishedAt(r.finishedCnt)
	if r.winnerIdx < 0 {
		r.winnerIdx = playerIdx
	}
	r.addLog(playerIdx, "finish", fmt.Sprintf("%d 番目に上がりました", r.finishedCnt), nil)
	// **上がった瞬間に終局させる（レビュー指摘 PR #5316）。**
	//
	// 以前は `resolveTrick` でしか決着させていませんでした。上がった直後に
	// 別の席が引き取ると、そのトリックは `pickUp` で流れて `resolveTrick` に
	// 届かず、**勝者が決まっているのに終局していない**盤面が残ります
	// ——それは `UnmarshalJSON` が「壊れている」として弾く状態そのもので、
	// 保存して読み直すと正当な対局が拒否されました。
	r.checkGameEnd()
}

// checkGameEnd は決着したかを返す。**最初に上がった人が勝ちなので 1 人で終わり。**
func (r *RollingStone) checkGameEnd() bool {
	if r.gameEndFlag {
		return true
	}
	if r.winnerIdx < 0 {
		return false
	}
	r.phase = RollingStonePhaseGameEnd
	r.gameEndFlag = true
	r.addLog(r.winnerIdx, "result", "手札を出し切りました", nil)
	return true
}

// NextTrickLead は次のリード席を返す（表示用）。
func (r *RollingStone) NextTrickLead() int { return r.leadPlayerIdx }

// GiveUp は投了する。
func (r *RollingStone) GiveUp() {
	if r.gameEndFlag {
		return
	}
	r.phase = RollingStonePhaseGameEnd
	r.gameEndFlag = true
	// 人間（席 0）以外でいちばん手札の少ない席を勝ちにする。
	best := -1
	for i := 1; i < r.config.PlayerCnt; i++ {
		if best < 0 || r.players[i].GetCardsSize() < r.players[best].GetCardsSize() {
			best = i
		}
	}
	r.winnerIdx = best
	r.addLog(0, "giveup", "投了しました", nil)
}

// chooseCpuCard は CPU の手。**手札を減らしたいので、安全に高い札から出す。**
func (r *RollingStone) chooseCpuCard(playerIdx int) int {
	valid := r.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := r.players[playerIdx]
	// **リードなら、いちばん長いスートから出す。** 短いスートを残すと引き取りやすい。
	if len(r.currentTrick) == 0 {
		counts := map[int]int{}
		for i := range p.GetCardsSize() {
			counts[p.GetCard(i).GetDesign()]++
		}
		pick, bestLen, bestRank := valid[0], -1, -1
		for _, i := range valid {
			c := p.GetCard(i)
			n, rank := counts[c.GetDesign()], rollingStoneRank(c)
			if n > bestLen || (n == bestLen && rank > bestRank) {
				pick, bestLen, bestRank = i, n, rank
			}
		}
		return pick
	}
	// フォローするだけなら安い札から。取っても得はない。
	pick, pickRank := valid[0], rollingStoneRank(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		if rk := rollingStoneRank(p.GetCard(i)); rk < pickRank {
			pick, pickRank = i, rk
		}
	}
	return pick
}

// GetHint は人間への助言を返す。
func (r *RollingStone) GetHint() *RollingStoneHint {
	if r.gameEndFlag || !r.IsHumanTurn() {
		return nil
	}
	if r.MustPickUp(r.currentPlayerIdx) {
		return &RollingStoneHint{Reason: "rollingstonePickUp"}
	}
	valid := r.GetValidPlayIndices(r.currentPlayerIdx)
	if len(valid) == 0 {
		return nil
	}
	idx := r.chooseCpuCard(r.currentPlayerIdx)
	reason := "rollingstoneFollow"
	if len(r.currentTrick) == 0 {
		reason = "rollingstoneLead"
	}
	return &RollingStoneHint{CardIndex: &idx, Reason: reason}
}

// addLog は棋譜に 1 行足す。
func (r *RollingStone) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	r.appendLog(playerIdx, actionType, detail, cards)
}

// --- アクセサ ---------------------------------------------------------------

// GetConfig はゲーム設定を返す。
func (r *RollingStone) GetConfig() RollingStoneConfig { return r.config }

// SetConfig はゲーム設定を設定する。**人数が変わると席も作り直す。**
func (r *RollingStone) SetConfig(cfg RollingStoneConfig) {
	r.config = cfg
	if len(r.players) != cfg.PlayerCnt {
		r.players = newRollingStoneSeats(cfg.PlayerCnt)
	}
}

// GetPhase は現在のフェーズを返す。
func (r *RollingStone) GetPhase() RollingStonePhase { return r.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (r *RollingStone) GetGameEndFlag() bool { return r.gameEndFlag }

// GetCurrentTrick は現在のトリックを返す。
func (r *RollingStone) GetCurrentTrick() []*TrickCard { return r.currentTrick }

// GetCurrentPlayerIdx は現在の手番を返す。
func (r *RollingStone) GetCurrentPlayerIdx() int { return r.currentPlayerIdx }

// GetLeadPlayerIdx はリード席を返す。
func (r *RollingStone) GetLeadPlayerIdx() int { return r.leadPlayerIdx }

// GetTrickNumber は解決済みのトリック数を返す。
func (r *RollingStone) GetTrickNumber() int { return r.trickNumber }

// GetLastPickupIdx は直前に引き取った席を返す（-1 = 無し）。
func (r *RollingStone) GetLastPickupIdx() int { return r.lastPickupIdx }

// GetDiscarded は場から抜けた札の枚数を返す。
func (r *RollingStone) GetDiscarded() int { return r.discarded }

// GetFinishedCnt は上がった人数を返す。
func (r *RollingStone) GetFinishedCnt() int { return r.finishedCnt }

// GetDeckSize はこの卓で使うデッキ枚数を返す。
func (r *RollingStone) GetDeckSize() int { return RollingStoneDeckSize(r.config.PlayerCnt) }

// GetPlayerCnt はプレイヤー数を返す。
func (r *RollingStone) GetPlayerCnt() int { return r.config.PlayerCnt }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (r *RollingStone) GetPlayer(i int) *RollingStonePlayer {
	if i < 0 || i >= len(r.players) {
		return nil
	}
	return r.players[i]
}

// GetWinnerIdx は勝った席を返す（-1: 未確定）。
func (r *RollingStone) GetWinnerIdx() int { return r.winnerIdx }

// rollingStoneJSON は KV スナップショットの表現。
type rollingStoneJSON struct {
	TrumpCards       *TrumpCards           `json:"tc"`
	Players          []*RollingStonePlayer `json:"pl"`
	Config           RollingStoneConfig    `json:"cf"`
	Phase            RollingStonePhase     `json:"ph"`
	CurrentTrick     []*TrickCard          `json:"ct"`
	CurrentPlayerIdx int                   `json:"ci"`
	LeadPlayerIdx    int                   `json:"li"`
	TrickNumber      int                   `json:"tn"`
	FinishedCnt      int                   `json:"fc"`
	Discarded        int                   `json:"dc"`
	GameEndFlag      bool                  `json:"ge"`
	WinnerIdx        int                   `json:"wi"`
	LastPickupIdx    int                   `json:"lp"`
	ActionLog        []*ActionLogEntry     `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (r *RollingStone) MarshalJSON() ([]byte, error) {
	return json.Marshal(&rollingStoneJSON{
		TrumpCards: r.trumpCards, Players: r.players, Config: r.config, Phase: r.phase,
		CurrentTrick: r.currentTrick, CurrentPlayerIdx: r.currentPlayerIdx,
		LeadPlayerIdx: r.leadPlayerIdx, TrickNumber: r.trickNumber,
		FinishedCnt: r.finishedCnt, Discarded: r.discarded, GameEndFlag: r.gameEndFlag,
		WinnerIdx: r.winnerIdx, LastPickupIdx: r.lastPickupIdx, ActionLog: r.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元
//
// **9 PR 連続でここに実バグが出た**ので、今回は書き込み側の関数が保っている関係を
// 先に写しています (#5302〜#5315)。9 回とも範囲検査では捕まらず、落ちたのは
// **フィールド同士の関係**でした:
//
//   - まとめて立つものは等値で（終了フラグ × フェーズ、勝者 × 終了フラグ）
//   - 数え上げは数え元と突き合わせる（上がった人数、札の総数）
//   - 「その値を書き込む唯一の関数」が保っている条件を写す
//     （手番とリード席は上がった席にならない、場札は在席数未満）
//   - ポインタと「範囲」を持たないものほど漏れる
func (r *RollingStone) UnmarshalJSON(data []byte) error {
	var j rollingStoneJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < RollingStonePhasePlay || j.Phase > RollingStonePhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.GameEndFlag != (j.Phase == RollingStonePhaseGameEnd) {
		return fmt.Errorf("game end flag %v disagrees with phase %d", j.GameEndFlag, j.Phase)
	}
	if j.TrumpCards == nil {
		return errors.New("missing trump cards")
	}
	if len(j.Players) != j.Config.PlayerCnt {
		return fmt.Errorf("players has %d entries for %d seats", len(j.Players), j.Config.PlayerCnt)
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
	} {
		if idx < 0 || idx >= j.Config.PlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= j.Config.PlayerCnt {
		return fmt.Errorf("invalid winner: %d", j.WinnerIdx)
	}
	// **勝者と終了フラグは対。** 勝者が決まるのは上がった瞬間で、そこで終局します。
	if j.GameEndFlag != (j.WinnerIdx >= 0) {
		return fmt.Errorf("winner %d disagrees with game end flag %v", j.WinnerIdx, j.GameEndFlag)
	}
	if j.LastPickupIdx < -1 || j.LastPickupIdx >= j.Config.PlayerCnt {
		return fmt.Errorf("invalid last pickup: %d", j.LastPickupIdx)
	}
	if j.TrickNumber < 0 {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.FinishedCnt < 0 || j.FinishedCnt > j.Config.PlayerCnt {
		return fmt.Errorf("invalid finished count: %d", j.FinishedCnt)
	}
	if len(j.CurrentTrick) > j.Config.PlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	// **枚数だけでなく中身も見る (#5310 の再発防止)。**
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= j.Config.PlayerCnt {
			return errors.New("invalid current trick entry")
		}
	}
	if len(j.ActionLog) > rollingStoneMaxSliceLen {
		return errors.New("rollingstone: input array exceeds maximum allowed size")
	}

	// **上がった順位は 1..finishedCnt の並べ替えでなければならない。**
	// 数だけ合わせても、同じ順位が 2 つあれば「その順で上がった」経路が無い。
	seen := make(map[int]bool, j.FinishedCnt)
	total := len(j.CurrentTrick)
	active := 0
	for i, p := range j.Players {
		if p == nil {
			return errors.New("nil player")
		}
		total += p.GetCardsSize()
		if !p.HasFinished() {
			active++
			continue
		}
		if p.GetFinishedAt() > j.FinishedCnt || seen[p.GetFinishedAt()] {
			return fmt.Errorf("seat %d has finishing rank %d, which is not a place in 1..%d",
				i, p.GetFinishedAt(), j.FinishedCnt)
		}
		seen[p.GetFinishedAt()] = true
	}
	if len(seen) != j.FinishedCnt {
		return fmt.Errorf("%d seats have finished but the count says %d", len(seen), j.FinishedCnt)
	}
	// **札は人数 × 8 枚しかない（#5314 で踏んだ「数え元と突き合わせる」形）。**
	//
	// **抜けた枚数を足して数える。** 全員フォローできたトリックは場から抜けるので、
	// 手札と場札だけを足すと途中の盤面では必ず足りません——最初に書いた
	// 「total == デッキ」は、トリックが 1 つ解決した時点で正しい盤面を拒否しました。
	if j.Discarded < 0 {
		return fmt.Errorf("invalid discarded count: %d", j.Discarded)
	}
	if want := RollingStoneDeckSize(j.Config.PlayerCnt); total+j.Discarded != want {
		return fmt.Errorf("hands, the trick and the discards hold %d cards, want %d",
			total+j.Discarded, want)
	}
	// **手番とリード席は上がった席にならない。** `advanceTurn`/`resolveTrick` が
	// そう保っています。通すと上がった席が打つ盤面になります。
	if !j.GameEndFlag {
		if j.Players[j.CurrentPlayerIdx].HasFinished() {
			return fmt.Errorf("seat %d is on turn but has already finished", j.CurrentPlayerIdx)
		}
		if j.Players[j.LeadPlayerIdx].HasFinished() {
			return fmt.Errorf("seat %d leads but has already finished", j.LeadPlayerIdx)
		}
		// **場札は在席数より少ない。** 揃った時点で解決されるので残りません。
		if len(j.CurrentTrick) >= active {
			return fmt.Errorf("current trick holds %d cards with %d seats still in", len(j.CurrentTrick), active)
		}
	}

	r.trumpCards = j.TrumpCards
	r.players, r.config, r.phase = j.Players, j.Config, j.Phase
	r.currentTrick, r.currentPlayerIdx = j.CurrentTrick, j.CurrentPlayerIdx
	r.leadPlayerIdx, r.trickNumber = j.LeadPlayerIdx, j.TrickNumber
	r.finishedCnt, r.discarded, r.gameEndFlag = j.FinishedCnt, j.Discarded, j.GameEndFlag
	r.winnerIdx, r.lastPickupIdx, r.actionLog = j.WinnerIdx, j.LastPickupIdx, j.ActionLog
	return nil
}
