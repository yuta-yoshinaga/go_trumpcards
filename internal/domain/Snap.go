//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// SnapPhase はスナップのフェーズ。
type SnapPhase int

// SnapPhase 定数
const (
	// SnapPhasePlay プレイ中
	SnapPhasePlay SnapPhase = 0
	// SnapPhaseGameEnd 終局
	SnapPhaseGameEnd SnapPhase = 1
)

// SnapPendingKind は保留中の CPU アクションの種類。
type SnapPendingKind int

// SnapPendingKind 定数
const (
	// SnapPendingNone 保留なし
	SnapPendingNone SnapPendingKind = 0
	// SnapPendingSnap CPU がスナップを宣言する予約
	SnapPendingSnap SnapPendingKind = 1
	// SnapPendingStep CPU が 1 枚めくる予約
	SnapPendingStep SnapPendingKind = 2
)

// SnapEventKind は直近イベントの種類。
type SnapEventKind int

// SnapEventKind 定数
const (
	// SnapEventNone イベント無し
	SnapEventNone SnapEventKind = 0
	// SnapEventStep カードがめくられた
	SnapEventStep SnapEventKind = 1
	// SnapEventSnapCorrect 正しい宣言（場札を総取り）
	SnapEventSnapCorrect SnapEventKind = 2
	// SnapEventSnapWrong 誤宣言（ペナルティ）
	SnapEventSnapWrong SnapEventKind = 3
	// SnapEventEliminated ストックが尽きて脱落
	SnapEventEliminated SnapEventKind = 4
)

// SnapDeckSize は 1 デッキの枚数。
const SnapDeckSize = 52

// snapMaxSliceLen は復元時に受け付けるスライスの上限。
const snapMaxSliceLen = 1000

// SnapLastEvent は直近イベント情報（UI フィードバック用）。
type SnapLastEvent struct {
	Kind      SnapEventKind `json:"kind"`
	PlayerIdx int           `json:"playerIdx"`
}

// SnapPending は保留中の CPU アクション。
type SnapPending struct {
	Kind       SnapPendingKind `json:"kind"`
	PlayerIdx  int             `json:"playerIdx"`
	DeadlineMs int64           `json:"deadlineMs"`
}

// SnapHint はスナップの助言。
type SnapHint struct {
	// Snap が true なら「いま宣言すべき」。
	Snap   bool
	Reason string
}

// Snap はスナップのゲーム。
//
// **トリガーは固定ではありません。** スラップジャックの「J かどうか」と違い、
// 「**直前に出た札と同じランクか**」なので、場札が 1 枚のあいだは成立しません。
type Snap struct {
	trumpCards     *TrumpCards
	players        []*SnapPlayer
	config         SnapConfig
	phase          SnapPhase
	centerPile     []*Card
	currentTurnIdx int
	pending        SnapPending
	lastEvent      SnapLastEvent
	gameEndFlag    bool
	winnerIdx      int
	actionLogBase

	// now 現在時刻取得関数 (テストで差し替え可能)
	now func() time.Time
	// rng CPU 反応時間抽選用 (テストで差し替え可能)
	rng *rand.Rand
}

// NewSnap はコンストラクタ。
func NewSnap(players []*SnapPlayer, config SnapConfig) *Snap {
	if config.Validate() != nil {
		config = DefaultSnapConfig()
	}
	if len(players) != config.PlayerCnt {
		players = newSnapSeats(config.PlayerCnt)
	}
	return &Snap{
		players:   players,
		config:    config,
		winnerIdx: -1,
		now:       time.Now,
		rng:       rand.New(rand.NewSource(rand.Int63())),
	}
}

// newSnapSeats は標準の席（人間 1 + CPU）を返す。
func newSnapSeats(n int) []*SnapPlayer {
	seats := make([]*SnapPlayer, 0, n)
	for i := range n {
		seats = append(seats, NewSnapPlayer(i == 0))
	}
	return seats
}

// NewDefaultSnap は標準セットアップを返す。
func NewDefaultSnap() *Snap {
	cfg := DefaultSnapConfig()
	return NewSnap(newSnapSeats(cfg.PlayerCnt), cfg)
}

// SetClock はテスト/シミュレーション用に時刻関数を差し替える。
func (g *Snap) SetClock(now func() time.Time) {
	if now != nil {
		g.now = now
	}
}

// SetRand はテスト用に乱数源を差し替える。
func (g *Snap) SetRand(r *rand.Rand) {
	if r != nil {
		g.rng = r
	}
}

// Reset はゲームをリセットして新しいゲームを開始する。
func (g *Snap) Reset() {
	g.phase = SnapPhasePlay
	g.centerPile = nil
	g.currentTurnIdx = 0
	g.pending = SnapPending{Kind: SnapPendingNone}
	g.lastEvent = SnapLastEvent{Kind: SnapEventNone}
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.actionLog = nil

	g.trumpCards = NewTrumpCards(0)
	g.trumpCards.Shuffle()
	for _, p := range g.players {
		p.ResetStock()
	}
	// **配り切る。端数は 1 枚多い人が出る。** 52 は 3 人では割り切れないので、
	// 「均等に配る」ことはできません——余りを伏せて捨てると札が消えるので、
	// 1 枚ずつ配り切ります。
	for i := 0; i < SnapDeckSize; i++ {
		c := g.trumpCards.DrawCard()
		if c == nil {
			break
		}
		g.players[i%g.config.PlayerCnt].AddToStockBottom(c)
	}
	g.addLog(-1, "start", fmt.Sprintf("スナップを開始しました（%d 人）", g.config.PlayerCnt), nil)
	g.scheduleNext()
}

// IsSnapAvailable はいま宣言が正しいかを返す。
//
// **場札が 2 枚以上あって、上 2 枚のランクが同じとき**だけ成立します。
// 1 枚しか出ていないあいだは「直前の札」が無いので、決して成立しません。
func (g *Snap) IsSnapAvailable() bool {
	if len(g.centerPile) < 2 {
		return false
	}
	top := g.centerPile[len(g.centerPile)-1]
	prev := g.centerPile[len(g.centerPile)-2]
	return top.GetValue() == prev.GetValue()
}

// PlayerStep は人間が 1 枚めくる。
func (g *Snap) PlayerStep() error {
	if g.gameEndFlag || g.phase != SnapPhasePlay {
		return ErrGameEnded
	}
	if g.currentTurnIdx != 0 {
		return ErrNotHumanTurn
	}
	g.step(0)
	return nil
}

// PlayerSnap は人間が宣言する。
func (g *Snap) PlayerSnap() error {
	if g.gameEndFlag || g.phase != SnapPhasePlay {
		return ErrGameEnded
	}
	g.snap(0)
	return nil
}

// step は playerIdx が 1 枚めくる。
func (g *Snap) step(playerIdx int) {
	p := g.players[playerIdx]
	card := p.DrawTop()
	if card == nil {
		// **ストックが尽きた席は脱落。** めくる番は次へ回します。
		g.eliminate(playerIdx)
		return
	}
	g.centerPile = append(g.centerPile, card)
	g.lastEvent = SnapLastEvent{Kind: SnapEventStep, PlayerIdx: playerIdx}
	g.addLog(playerIdx, "step", "1 枚めくりました", []*Card{card})
	g.advanceTurn()
	g.scheduleNext()
}

// snap は playerIdx が宣言する。
func (g *Snap) snap(playerIdx int) {
	if g.IsSnapAvailable() {
		taken := g.centerPile
		g.centerPile = nil
		g.players[playerIdx].AddToStockBottom(taken...)
		g.lastEvent = SnapLastEvent{Kind: SnapEventSnapCorrect, PlayerIdx: playerIdx}
		g.addLog(playerIdx, "snap", fmt.Sprintf("スナップ！ %d 枚を獲得しました", len(taken)), taken)
		// **取った人が次にめくる。**
		g.currentTurnIdx = playerIdx
		g.pending = SnapPending{Kind: SnapPendingNone}
		g.checkGameEnd()
		if !g.gameEndFlag {
			g.scheduleNext()
		}
		return
	}

	// **誤宣言はストックから 1 枚を場に差し出す。**
	g.lastEvent = SnapLastEvent{Kind: SnapEventSnapWrong, PlayerIdx: playerIdx}
	if c := g.players[playerIdx].DrawTop(); c != nil {
		g.centerPile = append(g.centerPile, c)
		g.addLog(playerIdx, "penalty", "誤宣言のペナルティで 1 枚差し出しました", []*Card{c})
	} else {
		g.addLog(playerIdx, "penalty", "誤宣言しましたが、差し出す札がありません", nil)
	}
	g.checkGameEnd()
	if !g.gameEndFlag {
		g.scheduleNext()
	}
}

// eliminate はストックが尽きた席を飛ばす。
func (g *Snap) eliminate(playerIdx int) {
	g.lastEvent = SnapLastEvent{Kind: SnapEventEliminated, PlayerIdx: playerIdx}
	g.addLog(playerIdx, "eliminate", "ストックが尽きました", nil)
	g.advanceTurn()
	g.checkGameEnd()
	if !g.gameEndFlag {
		g.scheduleNext()
	}
}

// advanceTurn は次にめくる席へ回す。**ストックのある席だけを回ります。**
func (g *Snap) advanceTurn() {
	for i := 1; i <= g.config.PlayerCnt; i++ {
		next := (g.currentTurnIdx + i) % g.config.PlayerCnt
		if g.players[next].HasStock() {
			g.currentTurnIdx = next
			return
		}
	}
	// 誰も持っていない（場に全部出ている）。手番はそのまま。
}

// checkGameEnd は 1 人が全札を集めたか、続けられなくなったかを判定する。
func (g *Snap) checkGameEnd() {
	alive := make([]int, 0, g.config.PlayerCnt)
	for i, p := range g.players {
		if p.HasStock() {
			alive = append(alive, i)
		}
	}
	switch {
	case len(alive) == 1 && len(g.centerPile) == 0:
		// **全札を 1 人が持っている。**
		g.finish(alive[0])
	case len(alive) == 0:
		// **全札が場に出たまま誰も宣言できない。** 打ち切って最多保持者を勝ちに。
		g.finish(-1)
	}
}

// finish は終局処理。
func (g *Snap) finish(winner int) {
	g.phase = SnapPhaseGameEnd
	g.gameEndFlag = true
	g.pending = SnapPending{Kind: SnapPendingNone}
	g.winnerIdx = winner
	if winner >= 0 {
		g.addLog(winner, "result", "全札を集めました", nil)
		return
	}
	g.addLog(-1, "result", "続けられなくなりました（場に全札）", nil)
}

// GiveUp は投了する。
func (g *Snap) GiveUp() {
	if g.gameEndFlag {
		return
	}
	g.phase = SnapPhaseGameEnd
	g.gameEndFlag = true
	g.pending = SnapPending{Kind: SnapPendingNone}
	// 人間（席 0）以外でいちばんストックの多い席を勝ちにする。
	best := -1
	for i := 1; i < g.config.PlayerCnt; i++ {
		if best < 0 || g.players[i].GetStockSize() > g.players[best].GetStockSize() {
			best = i
		}
	}
	g.winnerIdx = best
	g.addLog(0, "giveup", "投了しました", nil)
}

// Tick は CPU の保留アクションを（期限に達していれば）実行する。
//
// **同時反射をターン制に落とすための仕組み。** CPU は「いつ宣言するか」を
// 期限つきで予約し、人間はその期限より前に宣言できれば勝てます。
func (g *Snap) Tick() SnapPendingKind {
	if g.gameEndFlag || g.phase != SnapPhasePlay {
		return SnapPendingNone
	}
	if g.pending.Kind == SnapPendingNone {
		return SnapPendingNone
	}
	if g.now().UnixMilli() < g.pending.DeadlineMs {
		return SnapPendingNone
	}
	kind, who := g.pending.Kind, g.pending.PlayerIdx
	g.pending = SnapPending{Kind: SnapPendingNone}
	switch kind {
	case SnapPendingSnap:
		g.snap(who)
	case SnapPendingStep:
		g.step(who)
	default:
	}
	return kind
}

// scheduleNext は次に起きる CPU の行動を予約する。
//
// **宣言が成立しているなら、宣言のほうを予約する。** どの CPU も宣言できるので、
// いちばん反応の速い 1 体だけを予約します（同着は席順で決まり、決定的です）。
func (g *Snap) scheduleNext() {
	if g.gameEndFlag || g.phase != SnapPhasePlay {
		return
	}
	if g.IsSnapAvailable() {
		best, bestMs := -1, int64(0)
		for i := 1; i < g.config.PlayerCnt; i++ {
			ms := int64(g.drawReactionMs())
			if best < 0 || ms < bestMs {
				best, bestMs = i, ms
			}
		}
		if best >= 0 {
			g.pending = SnapPending{
				Kind:       SnapPendingSnap,
				PlayerIdx:  best,
				DeadlineMs: g.now().UnixMilli() + bestMs,
			}
			return
		}
	}
	// 宣言が成立していないなら、手番の CPU がめくるのを予約する。
	if g.currentTurnIdx != 0 {
		g.pending = SnapPending{
			Kind:       SnapPendingStep,
			PlayerIdx:  g.currentTurnIdx,
			DeadlineMs: g.now().UnixMilli() + int64(g.drawReactionMs()),
		}
		return
	}
	g.pending = SnapPending{Kind: SnapPendingNone}
}

// drawReactionMs は設定難易度に基づき正規分布から反応時間 (ms) を抽出する。
// 下限 SnapMinReactionMs でクランプ。
func (g *Snap) drawReactionMs() int {
	mean, sd := 900.0, 250.0
	switch g.config.CpuDifficulty {
	case SnapCpuEasy:
		mean, sd = 1400.0, 350.0
	case SnapCpuHard:
		mean, sd = 500.0, 150.0
	default:
	}
	ms := int(g.rng.NormFloat64()*sd + mean)
	if ms < SnapMinReactionMs {
		ms = SnapMinReactionMs
	}
	return ms
}

// GetHint は人間への助言を返す。
func (g *Snap) GetHint() *SnapHint {
	if g.gameEndFlag {
		return nil
	}
	if g.IsSnapAvailable() {
		return &SnapHint{Snap: true, Reason: "snapDeclare"}
	}
	if g.currentTurnIdx == 0 {
		return &SnapHint{Reason: "snapStep"}
	}
	return &SnapHint{Reason: "snapWait"}
}

// addLog は棋譜に 1 行足す。
func (g *Snap) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.appendLog(playerIdx, actionType, detail, cards)
}

// --- アクセサ ---------------------------------------------------------------

// GetConfig はゲーム設定を返す。
func (g *Snap) GetConfig() SnapConfig { return g.config }

// SetConfig はゲーム設定を設定する。**人数が変わると席も作り直す。**
func (g *Snap) SetConfig(cfg SnapConfig) {
	g.config = cfg
	if len(g.players) != cfg.PlayerCnt {
		g.players = newSnapSeats(cfg.PlayerCnt)
	}
}

// GetPhase は現在のフェーズを返す。
func (g *Snap) GetPhase() SnapPhase { return g.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Snap) GetGameEndFlag() bool { return g.gameEndFlag }

// GetCenterPile は場札を返す。
func (g *Snap) GetCenterPile() []*Card { return g.centerPile }

// GetCenterPileSize は場札の枚数を返す。
func (g *Snap) GetCenterPileSize() int { return len(g.centerPile) }

// GetTopCard は場札のいちばん上を返す（無ければ nil）。
func (g *Snap) GetTopCard() *Card {
	if len(g.centerPile) == 0 {
		return nil
	}
	return g.centerPile[len(g.centerPile)-1]
}

// GetCurrentTurnIdx は次にめくる席を返す。
func (g *Snap) GetCurrentTurnIdx() int { return g.currentTurnIdx }

// GetPending は保留中の CPU アクションを返す。
func (g *Snap) GetPending() SnapPending { return g.pending }

// GetLastEvent は直近イベントを返す。
func (g *Snap) GetLastEvent() SnapLastEvent { return g.lastEvent }

// GetPlayerCnt はプレイヤー数を返す。
func (g *Snap) GetPlayerCnt() int { return g.config.PlayerCnt }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *Snap) GetPlayer(i int) *SnapPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetWinnerIdx は勝った席を返す（-1: 未確定/決着なし）。
func (g *Snap) GetWinnerIdx() int { return g.winnerIdx }

// snapJSON は KV スナップショットの表現。
type snapJSON struct {
	TrumpCards     *TrumpCards       `json:"tc"`
	Players        []*SnapPlayer     `json:"pl"`
	Config         SnapConfig        `json:"cf"`
	Phase          SnapPhase         `json:"ph"`
	CenterPile     []*Card           `json:"cp"`
	CurrentTurnIdx int               `json:"ct"`
	Pending        SnapPending       `json:"pd"`
	LastEvent      SnapLastEvent     `json:"le"`
	GameEndFlag    bool              `json:"ge"`
	WinnerIdx      int               `json:"wi"`
	ActionLog      []*ActionLogEntry `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (g *Snap) MarshalJSON() ([]byte, error) {
	return json.Marshal(&snapJSON{
		TrumpCards: g.trumpCards, Players: g.players, Config: g.config, Phase: g.phase,
		CenterPile: g.centerPile, CurrentTurnIdx: g.currentTurnIdx,
		Pending: g.pending, LastEvent: g.lastEvent,
		GameEndFlag: g.gameEndFlag, WinnerIdx: g.winnerIdx, ActionLog: g.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元
//
// **8 PR 連続で「個々のフィールドは範囲内だが組み合わせがあり得ない」を通していた**
// ので、対で立つものは等値で、数え上げは数え元と突き合わせて検証します
// (#5302〜#5314)。**「範囲」を持たないもの**——bool・ポインタ・スライス・
// フィールド間の関係——ほど漏れます。
func (g *Snap) UnmarshalJSON(data []byte) error {
	var j snapJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < SnapPhasePlay || j.Phase > SnapPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	// **終了フラグとフェーズは対（#5313 で踏んだ形）。**
	if j.GameEndFlag != (j.Phase == SnapPhaseGameEnd) {
		return fmt.Errorf("game end flag %v disagrees with phase %d", j.GameEndFlag, j.Phase)
	}
	if j.CurrentTurnIdx < 0 || j.CurrentTurnIdx >= j.Config.PlayerCnt {
		return fmt.Errorf("invalid current turn: %d", j.CurrentTurnIdx)
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= j.Config.PlayerCnt {
		return fmt.Errorf("invalid winner: %d", j.WinnerIdx)
	}
	if !j.GameEndFlag && j.WinnerIdx != -1 {
		return fmt.Errorf("winner %d before the game ended", j.WinnerIdx)
	}
	if j.Pending.Kind < SnapPendingNone || j.Pending.Kind > SnapPendingStep {
		return fmt.Errorf("invalid pending kind: %d", j.Pending.Kind)
	}
	// **保留の種類と席は対。** 保留が無いのに席が立っていたら壊れています。
	if j.Pending.Kind == SnapPendingNone {
		if j.Pending.PlayerIdx != 0 || j.Pending.DeadlineMs != 0 {
			return errors.New("pending payload without a pending action")
		}
	} else {
		// **CPU の予約なので席 0（人間）はあり得ない。**
		if j.Pending.PlayerIdx <= 0 || j.Pending.PlayerIdx >= j.Config.PlayerCnt {
			return fmt.Errorf("invalid pending player: %d", j.Pending.PlayerIdx)
		}
		if j.GameEndFlag {
			return errors.New("pending action after the game ended")
		}
	}
	if j.LastEvent.Kind < SnapEventNone || j.LastEvent.Kind > SnapEventEliminated {
		return fmt.Errorf("invalid last event kind: %d", j.LastEvent.Kind)
	}
	if j.LastEvent.Kind != SnapEventNone &&
		(j.LastEvent.PlayerIdx < 0 || j.LastEvent.PlayerIdx >= j.Config.PlayerCnt) {
		return fmt.Errorf("invalid last event player: %d", j.LastEvent.PlayerIdx)
	}
	if len(j.CenterPile) > snapMaxSliceLen || len(j.ActionLog) > snapMaxSliceLen {
		return errors.New("snap: input array exceeds maximum allowed size")
	}
	// **枚数だけでなく中身も見る (#5310 の再発防止)。**
	for _, c := range j.CenterPile {
		if c == nil {
			return errors.New("nil card in the center pile")
		}
	}
	if len(j.Players) != j.Config.PlayerCnt {
		return fmt.Errorf("players has %d entries for %d seats", len(j.Players), j.Config.PlayerCnt)
	}
	// **札は 52 枚しかない（#5314 で踏んだ「数え元と突き合わせる」形）。**
	// 全員のストックと場札を足して 52 にならない盤面は存在しません。
	total := len(j.CenterPile)
	for _, p := range j.Players {
		if p == nil {
			return errors.New("nil player")
		}
		total += p.GetStockSize()
	}
	if total != SnapDeckSize {
		return fmt.Errorf("stocks and the pile hold %d cards, want %d", total, SnapDeckSize)
	}

	if j.TrumpCards != nil {
		g.trumpCards = j.TrumpCards
	} else {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players, g.config, g.phase = j.Players, j.Config, j.Phase
	g.centerPile, g.currentTurnIdx = j.CenterPile, j.CurrentTurnIdx
	g.pending, g.lastEvent = j.Pending, j.LastEvent
	g.gameEndFlag, g.winnerIdx, g.actionLog = j.GameEndFlag, j.WinnerIdx, j.ActionLog
	if g.now == nil {
		g.now = time.Now
	}
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(rand.Int63()))
	}
	return nil
}
