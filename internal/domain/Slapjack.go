package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// SlapjackPlayerCnt スラップジャックのプレイヤー数 (人間 + CPU)
const SlapjackPlayerCnt = 2

// SlapjackPhase ゲームフェーズ
type SlapjackPhase int

// スラップジャックのフェーズ定数
const (
	// SlapjackPhasePlay 進行中
	SlapjackPhasePlay SlapjackPhase = 0
	// SlapjackPhaseGameEnd ゲーム終了
	SlapjackPhaseGameEnd SlapjackPhase = 1
)

// SlapjackPendingKind 保留中の CPU アクション種別
type SlapjackPendingKind int

// SlapjackPendingKind 定数
const (
	// SlapjackPendingNone 保留なし
	SlapjackPendingNone SlapjackPendingKind = 0
	// SlapjackPendingStep CPU が次のステップ (カードめくり) を予約中
	SlapjackPendingStep SlapjackPendingKind = 1
	// SlapjackPendingSlap CPU がスラップを予約中 (場のトップが J のとき)
	SlapjackPendingSlap SlapjackPendingKind = 2
)

// SlapjackEventKind 直近イベント種別 (UI 演出用)
type SlapjackEventKind int

// SlapjackEventKind 定数
const (
	// SlapjackEventNone イベント無し
	SlapjackEventNone SlapjackEventKind = 0
	// SlapjackEventStep カードがめくられた
	SlapjackEventStep SlapjackEventKind = 1
	// SlapjackEventSlapCorrect 正しい J スラップ (パイル獲得)
	SlapjackEventSlapCorrect SlapjackEventKind = 2
	// SlapjackEventSlapWrong 誤スラップ (ペナルティ)
	SlapjackEventSlapWrong SlapjackEventKind = 3
)

// SlapjackLastEvent 直近イベント情報 (UI フィードバック用)
type SlapjackLastEvent struct {
	Kind      SlapjackEventKind `json:"kind"`
	PlayerIdx int               `json:"playerIdx"`
}

// SlapjackPending 保留中の CPU アクション
type SlapjackPending struct {
	Kind       SlapjackPendingKind `json:"kind"`
	DeadlineMs int64               `json:"deadlineMs"`
}

// Slapjack スラップジャックゲームクラス
type Slapjack struct {
	trumpCards     *TrumpCards
	players        [SlapjackPlayerCnt]*SlapjackPlayer
	config         SlapjackConfig
	phase          SlapjackPhase
	centerPile     []*Card
	currentTurnIdx int
	pending        SlapjackPending
	lastEvent      SlapjackLastEvent
	gameEndFlag    bool
	winnerIdx      int
	actionLogBase

	// now 現在時刻取得関数 (テストで差し替え可能)
	now func() time.Time
	// rng CPU 反応時間抽選用 (テストで差し替え可能)
	rng *rand.Rand
}

// NewSlapjack コンストラクタ
func NewSlapjack(trumpCards *TrumpCards, players []*SlapjackPlayer, config SlapjackConfig) *Slapjack {
	g := &Slapjack{
		trumpCards: trumpCards,
		config:     config,
		winnerIdx:  -1,
		now:        time.Now,
		rng:        rand.New(rand.NewSource(rand.Int63())),
	}
	for i := 0; i < SlapjackPlayerCnt && i < len(players); i++ {
		g.players[i] = players[i]
	}
	return g
}

// NewDefaultSlapjack returns Slapjack with the standard 2-player setup
// (1 human, 1 CPU) and DefaultSlapjackConfig.
func NewDefaultSlapjack() *Slapjack {
	players := []*SlapjackPlayer{
		NewSlapjackPlayer(true),
		NewSlapjackPlayer(false),
	}
	return NewSlapjack(NewTrumpCards(0), players, DefaultSlapjackConfig())
}

// SetClock テスト/シミュレーション用に時刻関数を差し替える
func (g *Slapjack) SetClock(now func() time.Time) {
	if now != nil {
		g.now = now
	}
}

// SetRand テスト用に乱数源を差し替える
func (g *Slapjack) SetRand(r *rand.Rand) {
	if r != nil {
		g.rng = r
	}
}

// Reset ゲームをリセットして新しいゲームを開始する
func (g *Slapjack) Reset() {
	g.phase = SlapjackPhasePlay
	g.centerPile = nil
	g.currentTurnIdx = 0
	g.pending = SlapjackPending{Kind: SlapjackPendingNone}
	g.lastEvent = SlapjackLastEvent{Kind: SlapjackEventNone}
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.actionLog = nil

	for _, p := range g.players {
		p.Reset()
		p.ResetStock()
		p.SetIsFinished(false)
	}

	g.trumpCards.Shuffle()

	cards := make([]*Card, 0, CardCnt)
	for range CardCnt {
		c := g.trumpCards.DrawCard()
		if c != nil {
			cards = append(cards, c)
		}
	}

	half := len(cards) / 2
	for pi := range SlapjackPlayerCnt {
		start := pi * half
		g.players[pi].AddToStockBottom(cards[start : start+half]...)
	}
}

// GetConfig 設定取得
func (g *Slapjack) GetConfig() SlapjackConfig { return g.config }

// SetConfig 設定更新
func (g *Slapjack) SetConfig(cfg SlapjackConfig) { g.config = cfg }

// ResetWithConfig 設定を更新してゲームを初期化する
func (g *Slapjack) ResetWithConfig(cfg SlapjackConfig) {
	g.config = cfg
	g.Reset()
}

// GetPhase フェーズ取得
func (g *Slapjack) GetPhase() SlapjackPhase { return g.phase }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Slapjack) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得
func (g *Slapjack) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *Slapjack) GetPlayerCnt() int { return SlapjackPlayerCnt }

// GetPlayer プレイヤー取得
func (g *Slapjack) GetPlayer(i int) *SlapjackPlayer {
	if i < 0 || i >= SlapjackPlayerCnt {
		return nil
	}
	return g.players[i]
}

// GetCenterPileSize 場の総枚数
func (g *Slapjack) GetCenterPileSize() int { return len(g.centerPile) }

// GetTopCard 場のトップカード (空なら nil)
func (g *Slapjack) GetTopCard() *Card {
	if len(g.centerPile) == 0 {
		return nil
	}
	return g.centerPile[len(g.centerPile)-1]
}

// GetCurrentTurnIdx 現在の手番プレイヤー
func (g *Slapjack) GetCurrentTurnIdx() int { return g.currentTurnIdx }

// IsHumanTurn 人間の手番か
func (g *Slapjack) IsHumanTurn() bool {
	if g.gameEndFlag {
		return false
	}
	p := g.GetPlayer(g.currentTurnIdx)
	return p != nil && p.GetIsHuman()
}

// IsTopJack 場のトップカードが J か
func (g *Slapjack) IsTopJack() bool {
	c := g.GetTopCard()
	return c != nil && c.GetValue() == SlapjackJackValue
}

// GetPending 保留中の CPU アクション
func (g *Slapjack) GetPending() SlapjackPending { return g.pending }

// GetLastEvent 直近イベント
func (g *Slapjack) GetLastEvent() SlapjackLastEvent { return g.lastEvent }

// --- アクション ---

// Step 現手番プレイヤーがストックの先頭1枚を場に出す。
//
// ゲーム終了済みなら ErrGameEnded、ストックが空なら ErrInvalidPlay を返す
// (相手プレイヤーが勝者になる)。J が出た場合は CPU の slap 予約を組む。
func (g *Slapjack) Step() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	cur := g.players[g.currentTurnIdx]
	c := cur.DrawTop()
	if c == nil {
		// 手番プレイヤーは出すカードが無い → 相手の勝ち
		g.endGame(g.opponentIdx(g.currentTurnIdx))
		return ErrInvalidPlay
	}
	g.centerPile = append(g.centerPile, c)
	g.lastEvent = SlapjackLastEvent{Kind: SlapjackEventStep, PlayerIdx: g.currentTurnIdx}
	g.appendLog(g.currentTurnIdx, "step", fmt.Sprintf("flip from stock (top=%d)", c.GetValue()), []*Card{c})

	if g.IsTopJack() {
		// 場のトップが J → CPU は slap を予約 (人間が先に slap する可能性も残す)
		g.scheduleCpuSlap()
	} else {
		// 通常: 手番が相手に移る。ただし相手のストックが空なら自分のままで続行する
		// (相手が J を slap して復帰する余地を残す)。
		nextIdx := g.opponentIdx(g.currentTurnIdx)
		if g.players[nextIdx].HasStock() {
			g.currentTurnIdx = nextIdx
		}
		g.maybeScheduleCpuStep()
	}
	g.checkStuck()
	return nil
}

// Slap 指定プレイヤーがスラップを試みる。
//
// 場のトップが J なら playerIdx がパイルを獲得し手番もそのプレイヤーになる。
// J でない場合は誤スラップとしてペナルティ (1 枚を相手に渡す) を課す。
func (g *Slapjack) Slap(playerIdx int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if playerIdx < 0 || playerIdx >= SlapjackPlayerCnt {
		return ErrInvalidPlay
	}
	if len(g.centerPile) == 0 {
		return ErrInvalidPlay
	}
	if g.IsTopJack() {
		g.applyCorrectSlap(playerIdx)
	} else {
		g.applyWrongSlap(playerIdx)
	}
	g.checkStuck()
	return nil
}

// Tick CPU の保留アクションを (deadline 到達済みなら) 実行する。
//
// 戻り値は実際に発火したアクション種別 (発火しなかった場合は SlapjackPendingNone)。
func (g *Slapjack) Tick() SlapjackPendingKind {
	if g.gameEndFlag {
		return SlapjackPendingNone
	}
	if g.pending.Kind == SlapjackPendingNone {
		return SlapjackPendingNone
	}
	if g.now().UnixMilli() < g.pending.DeadlineMs {
		return SlapjackPendingNone
	}
	kind := g.pending.Kind
	// 予約をクリアしてから実行 (実行中に新しい予約が組まれる場合もあるため)
	g.pending = SlapjackPending{Kind: SlapjackPendingNone}
	switch kind {
	case SlapjackPendingStep:
		_ = g.Step()
	case SlapjackPendingSlap:
		_ = g.Slap(g.cpuIdx())
	}
	return kind
}

// --- 内部ヘルパ ---

// applyCorrectSlap 正しい J スラップ: パイルを獲得し手番もそのプレイヤーへ。
func (g *Slapjack) applyCorrectSlap(playerIdx int) {
	got := g.centerPile
	g.centerPile = nil
	// パイル全体をシャッフルして勝者ストックの底に裏向きで戻す
	g.rng.Shuffle(len(got), func(i, j int) { got[i], got[j] = got[j], got[i] })
	g.players[playerIdx].AddToStockBottom(got...)
	g.currentTurnIdx = playerIdx
	g.lastEvent = SlapjackLastEvent{Kind: SlapjackEventSlapCorrect, PlayerIdx: playerIdx}
	g.appendLog(playerIdx, "slap", fmt.Sprintf("correct slap, +%d cards", len(got)), nil)
	g.pending = SlapjackPending{Kind: SlapjackPendingNone}
	g.maybeScheduleCpuStep()
}

// applyWrongSlap 誤スラップ: ペナルティ 1 枚を相手に渡す。
func (g *Slapjack) applyWrongSlap(playerIdx int) {
	offender := g.players[playerIdx]
	opp := g.opponentIdx(playerIdx)
	moved := 0
	for range SlapjackPenaltyCount {
		c := offender.DrawTop()
		if c == nil {
			break
		}
		g.players[opp].AddToStockBottom(c)
		moved++
	}
	g.lastEvent = SlapjackLastEvent{Kind: SlapjackEventSlapWrong, PlayerIdx: playerIdx}
	g.appendLog(playerIdx, "slap", fmt.Sprintf("wrong slap, -%d cards", moved), nil)
	if !offender.HasStock() {
		g.endGame(opp)
		return
	}
	// 場のトップが J でなくなった (= ペナルティを促した非 J カード) ので CPU 予約をクリア
	g.pending = SlapjackPending{Kind: SlapjackPendingNone}
	g.maybeScheduleCpuStep()
}

// maybeScheduleCpuStep 手番が CPU のとき、Step を一定遅延で予約する。
// すでに予約があれば上書きしない。
func (g *Slapjack) maybeScheduleCpuStep() {
	if g.gameEndFlag {
		return
	}
	if g.pending.Kind != SlapjackPendingNone {
		return
	}
	if g.players[g.currentTurnIdx].GetIsHuman() {
		return
	}
	g.pending = SlapjackPending{
		Kind:       SlapjackPendingStep,
		DeadlineMs: g.now().UnixMilli() + int64(g.drawReactionMs()),
	}
}

// scheduleCpuSlap 場のトップが J の状態で CPU の slap を予約する。
func (g *Slapjack) scheduleCpuSlap() {
	if g.gameEndFlag {
		return
	}
	g.pending = SlapjackPending{
		Kind:       SlapjackPendingSlap,
		DeadlineMs: g.now().UnixMilli() + int64(g.drawReactionMs()),
	}
}

// drawReactionMs 設定難易度に基づき正規分布から反応時間 (ms) を抽出する。
// 下限 SlapjackMinReactionMs でクランプ。
func (g *Slapjack) drawReactionMs() int {
	mean := float64(g.config.ReactionMeanMs())
	std := float64(g.config.ReactionStdDevMs())
	v := mean + g.rng.NormFloat64()*std
	if v < float64(SlapjackMinReactionMs) {
		v = float64(SlapjackMinReactionMs)
	}
	return int(v)
}

// opponentIdx 相手プレイヤーのインデックスを返す。
func (g *Slapjack) opponentIdx(idx int) int {
	if idx == 0 {
		return 1
	}
	return 0
}

// cpuIdx CPU プレイヤーのインデックスを返す。
// NewDefaultSlapjack の規約により player[0] = 人間, player[1] = CPU で固定。
func (g *Slapjack) cpuIdx() int { return 1 }

// endGame ゲームを終了し勝者を確定する。
func (g *Slapjack) endGame(winner int) {
	g.gameEndFlag = true
	g.phase = SlapjackPhaseGameEnd
	g.winnerIdx = winner
	g.pending = SlapjackPending{Kind: SlapjackPendingNone}
	if winner >= 0 && winner < SlapjackPlayerCnt {
		g.players[winner].SetIsFinished(true)
	}
}

// checkStuck 停滞検出 — 両者ストックが空でかつ場のトップが J でもないケースで
// 終了させる (誰も flip できず slap も起き得ないので永久に進まないため)。
func (g *Slapjack) checkStuck() {
	if g.gameEndFlag {
		return
	}
	if !g.players[0].HasStock() && !g.players[1].HasStock() && !g.IsTopJack() {
		g.endGame(0)
	}
}

// --- JSON ---

// slapjackJSON is the JSON wire format for Slapjack.
type slapjackJSON struct {
	TrumpCards     *TrumpCards                        `json:"tc"`
	Players        [SlapjackPlayerCnt]*SlapjackPlayer `json:"ps"`
	Config         SlapjackConfig                     `json:"cf"`
	Phase          SlapjackPhase                      `json:"ph"`
	CenterPile     []*Card                            `json:"cp"`
	CurrentTurnIdx int                                `json:"ct"`
	Pending        SlapjackPending                    `json:"pd"`
	LastEvent      SlapjackLastEvent                  `json:"le"`
	GameEndFlag    bool                               `json:"ge"`
	WinnerIdx      int                                `json:"wi"`
	ActionLog      []*ActionLogEntry                  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Slapjack) MarshalJSON() ([]byte, error) {
	return json.Marshal(slapjackJSON{
		TrumpCards:     g.trumpCards,
		Players:        g.players,
		Config:         g.config,
		Phase:          g.phase,
		CenterPile:     g.centerPile,
		CurrentTurnIdx: g.currentTurnIdx,
		Pending:        g.pending,
		LastEvent:      g.lastEvent,
		GameEndFlag:    g.gameEndFlag,
		WinnerIdx:      g.winnerIdx,
		ActionLog:      g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Slapjack) UnmarshalJSON(data []byte) error {
	var j slapjackJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.centerPile = j.CenterPile
	g.currentTurnIdx = j.CurrentTurnIdx
	g.pending = j.Pending
	g.lastEvent = j.LastEvent
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.actionLog = j.ActionLog
	if g.now == nil {
		g.now = time.Now
	}
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(rand.Int63()))
	}
	return nil
}
