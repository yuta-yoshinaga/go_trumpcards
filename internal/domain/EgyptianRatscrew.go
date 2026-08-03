package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// EgyptianRatscrewPlayerCnt エジプシャン・ラットスクリューのプレイヤー数 (人間 + CPU)
const EgyptianRatscrewPlayerCnt = 2

// EgyptianRatscrewPhase ゲームフェーズ
type EgyptianRatscrewPhase int

// エジプシャン・ラットスクリューのフェーズ定数
const (
	// EgyptianRatscrewPhasePlay 進行中
	EgyptianRatscrewPhasePlay EgyptianRatscrewPhase = 0
	// EgyptianRatscrewPhaseGameEnd ゲーム終了
	EgyptianRatscrewPhaseGameEnd EgyptianRatscrewPhase = 1
)

// EgyptianRatscrewPendingKind 保留中の CPU アクション種別
type EgyptianRatscrewPendingKind int

// EgyptianRatscrewPendingKind 定数
const (
	// EgyptianRatscrewPendingNone 保留なし
	EgyptianRatscrewPendingNone EgyptianRatscrewPendingKind = 0
	// EgyptianRatscrewPendingStep CPU が次のステップ (カードめくり) を予約中
	EgyptianRatscrewPendingStep EgyptianRatscrewPendingKind = 1
	// EgyptianRatscrewPendingSlap CPU がスラップを予約中
	EgyptianRatscrewPendingSlap EgyptianRatscrewPendingKind = 2
)

// EgyptianRatscrewSlapReason 直近の正スラップ理由 (UI 演出/ログ用)
type EgyptianRatscrewSlapReason int

// EgyptianRatscrewSlapReason 定数
const (
	// EgyptianRatscrewSlapReasonNone 正スラップ無し
	EgyptianRatscrewSlapReasonNone EgyptianRatscrewSlapReason = 0
	// EgyptianRatscrewSlapReasonPair ペア (上 2 枚同ランク)
	EgyptianRatscrewSlapReasonPair EgyptianRatscrewSlapReason = 1
	// EgyptianRatscrewSlapReasonSandwich サンドイッチ (上 1 枚目と上 3 枚目が同ランク)
	EgyptianRatscrewSlapReasonSandwich EgyptianRatscrewSlapReason = 2
)

// EgyptianRatscrewEventKind 直近イベント種別 (UI 演出用)
type EgyptianRatscrewEventKind int

// EgyptianRatscrewEventKind 定数
const (
	// EgyptianRatscrewEventNone イベント無し
	EgyptianRatscrewEventNone EgyptianRatscrewEventKind = 0
	// EgyptianRatscrewEventStep カードがめくられた
	EgyptianRatscrewEventStep EgyptianRatscrewEventKind = 1
	// EgyptianRatscrewEventSlapCorrect 正しいスラップ (パイル獲得)
	EgyptianRatscrewEventSlapCorrect EgyptianRatscrewEventKind = 2
	// EgyptianRatscrewEventSlapWrong 誤スラップ (ペナルティ)
	EgyptianRatscrewEventSlapWrong EgyptianRatscrewEventKind = 3
	// EgyptianRatscrewEventChanceWin チャンスバトルで絵札側が場を獲得
	EgyptianRatscrewEventChanceWin EgyptianRatscrewEventKind = 4
)

// EgyptianRatscrewLastEvent 直近イベント情報 (UI フィードバック用)
type EgyptianRatscrewLastEvent struct {
	Kind       EgyptianRatscrewEventKind  `json:"kind"`
	PlayerIdx  int                        `json:"playerIdx"`
	SlapReason EgyptianRatscrewSlapReason `json:"slapReason"`
}

// EgyptianRatscrewPending 保留中の CPU アクション
type EgyptianRatscrewPending struct {
	Kind       EgyptianRatscrewPendingKind `json:"kind"`
	DeadlineMs int64                       `json:"deadlineMs"`
}

// EgyptianRatscrew エジプシャン・ラットスクリューゲームクラス
type EgyptianRatscrew struct {
	trumpCards     *TrumpCards
	players        [EgyptianRatscrewPlayerCnt]*EgyptianRatscrewPlayer
	config         EgyptianRatscrewConfig
	phase          EgyptianRatscrewPhase
	centerPile     []*Card
	currentTurnIdx int
	// chanceRemaining > 0 のとき、currentTurnIdx は絵札に対するチャンスを払う側 (相手)
	chanceRemaining int
	// chanceFromIdx チャンスを課したプレイヤー (絵札を出した側)。-1 = チャンスバトル中ではない
	chanceFromIdx int
	pending       EgyptianRatscrewPending
	lastEvent     EgyptianRatscrewLastEvent
	gameEndFlag   bool
	winnerIdx     int
	actionLog     []*ActionLogEntry

	// now 現在時刻取得関数 (テストで差し替え可能)
	now func() time.Time
	// rng CPU 反応時間抽選用 (テストで差し替え可能)
	rng *rand.Rand
}

// NewEgyptianRatscrew コンストラクタ
func NewEgyptianRatscrew(trumpCards *TrumpCards, players []*EgyptianRatscrewPlayer, config EgyptianRatscrewConfig) *EgyptianRatscrew {
	g := &EgyptianRatscrew{
		trumpCards:    trumpCards,
		config:        config,
		winnerIdx:     -1,
		chanceFromIdx: -1,
		now:           time.Now,
		rng:           rand.New(rand.NewSource(rand.Int63())),
	}
	for i := 0; i < EgyptianRatscrewPlayerCnt && i < len(players); i++ {
		g.players[i] = players[i]
	}
	return g
}

// NewDefaultEgyptianRatscrew returns EgyptianRatscrew with the standard 2-player setup
// (1 human, 1 CPU) and DefaultEgyptianRatscrewConfig.
func NewDefaultEgyptianRatscrew() *EgyptianRatscrew {
	players := []*EgyptianRatscrewPlayer{
		NewEgyptianRatscrewPlayer(true),
		NewEgyptianRatscrewPlayer(false),
	}
	return NewEgyptianRatscrew(NewTrumpCards(0), players, DefaultEgyptianRatscrewConfig())
}

// SetClock テスト/シミュレーション用に時刻関数を差し替える
func (g *EgyptianRatscrew) SetClock(now func() time.Time) {
	if now != nil {
		g.now = now
	}
}

// SetRand テスト用に乱数源を差し替える
func (g *EgyptianRatscrew) SetRand(r *rand.Rand) {
	if r != nil {
		g.rng = r
	}
}

// Reset ゲームをリセットして新しいゲームを開始する
func (g *EgyptianRatscrew) Reset() {
	g.phase = EgyptianRatscrewPhasePlay
	g.centerPile = nil
	g.currentTurnIdx = 0
	g.chanceRemaining = 0
	g.chanceFromIdx = -1
	g.pending = EgyptianRatscrewPending{Kind: EgyptianRatscrewPendingNone}
	g.lastEvent = EgyptianRatscrewLastEvent{Kind: EgyptianRatscrewEventNone}
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
	for pi := range EgyptianRatscrewPlayerCnt {
		start := pi * half
		g.players[pi].AddToStockBottom(cards[start : start+half]...)
	}
}

// GetConfig 設定取得
func (g *EgyptianRatscrew) GetConfig() EgyptianRatscrewConfig { return g.config }

// SetConfig 設定更新
func (g *EgyptianRatscrew) SetConfig(cfg EgyptianRatscrewConfig) { g.config = cfg }

// ResetWithConfig 設定を更新してゲームを初期化する
func (g *EgyptianRatscrew) ResetWithConfig(cfg EgyptianRatscrewConfig) {
	g.config = cfg
	g.Reset()
}

// GetPhase フェーズ取得
func (g *EgyptianRatscrew) GetPhase() EgyptianRatscrewPhase { return g.phase }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *EgyptianRatscrew) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得
func (g *EgyptianRatscrew) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *EgyptianRatscrew) GetPlayerCnt() int { return EgyptianRatscrewPlayerCnt }

// GetPlayer プレイヤー取得
func (g *EgyptianRatscrew) GetPlayer(i int) *EgyptianRatscrewPlayer {
	if i < 0 || i >= EgyptianRatscrewPlayerCnt {
		return nil
	}
	return g.players[i]
}

// GetCenterPileSize 場の総枚数
func (g *EgyptianRatscrew) GetCenterPileSize() int { return len(g.centerPile) }

// GetTopCard 場のトップカード (空なら nil)
func (g *EgyptianRatscrew) GetTopCard() *Card {
	if len(g.centerPile) == 0 {
		return nil
	}
	return g.centerPile[len(g.centerPile)-1]
}

// GetCurrentTurnIdx 現在の手番プレイヤー
func (g *EgyptianRatscrew) GetCurrentTurnIdx() int { return g.currentTurnIdx }

// IsHumanTurn 人間の手番か
func (g *EgyptianRatscrew) IsHumanTurn() bool {
	if g.gameEndFlag {
		return false
	}
	p := g.GetPlayer(g.currentTurnIdx)
	return p != nil && p.GetIsHuman()
}

// IsTopFaceCard 場のトップカードが絵札 (J/Q/K/A) か
func (g *EgyptianRatscrew) IsTopFaceCard() bool {
	c := g.GetTopCard()
	return c != nil && IsFaceCard(c.GetValue())
}

// IsSlappable 場の上 2 枚がペアまたは上 3 枚がサンドイッチ。
func (g *EgyptianRatscrew) IsSlappable() bool {
	return g.slapReason() != EgyptianRatscrewSlapReasonNone
}

// slapReason 場のトップ列が成立させているスラップ理由を返す。
// ペアとサンドイッチが同時に成立する場合 (例 5-5-5) はペアを優先。
func (g *EgyptianRatscrew) slapReason() EgyptianRatscrewSlapReason {
	n := len(g.centerPile)
	if n >= 2 {
		if g.centerPile[n-1].GetValue() == g.centerPile[n-2].GetValue() {
			return EgyptianRatscrewSlapReasonPair
		}
	}
	if n >= 3 {
		if g.centerPile[n-1].GetValue() == g.centerPile[n-3].GetValue() {
			return EgyptianRatscrewSlapReasonSandwich
		}
	}
	return EgyptianRatscrewSlapReasonNone
}

// GetChanceRemaining チャンスバトル中に相手に残された flip 回数。0 ならチャンスバトル外。
func (g *EgyptianRatscrew) GetChanceRemaining() int { return g.chanceRemaining }

// GetChanceFromIdx チャンスを課したプレイヤーインデックス (絵札を出した側)。チャンスバトル外では -1。
func (g *EgyptianRatscrew) GetChanceFromIdx() int { return g.chanceFromIdx }

// GetPending 保留中の CPU アクション
func (g *EgyptianRatscrew) GetPending() EgyptianRatscrewPending { return g.pending }

// GetLastEvent 直近イベント
func (g *EgyptianRatscrew) GetLastEvent() EgyptianRatscrewLastEvent { return g.lastEvent }

// GetActionLog 棋譜取得
func (g *EgyptianRatscrew) GetActionLog() []*ActionLogEntry { return g.actionLog }

// --- アクション ---

// Step 現手番プレイヤーがストックの先頭1枚を場に出す。
//
// チャンスバトル中なら相手 (非絵札出し側) がカードを払う。
// 絵札が出た場合は相手にチャンスバトルを課す。
// 絵札以外でチャンスバトル中なら chanceRemaining を 1 減らし、0 になったら絵札側が場を獲得。
func (g *EgyptianRatscrew) Step() error {
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
	g.lastEvent = EgyptianRatscrewLastEvent{Kind: EgyptianRatscrewEventStep, PlayerIdx: g.currentTurnIdx}
	g.appendLog(g.currentTurnIdx, "step", fmt.Sprintf("flip from stock (top=%d)", c.GetValue()), []*Card{c})

	// 場の上 2 枚 / 上 3 枚で slap が成立した場合は、絵札判定より先にスラップレースを開始する。
	// (例: 絵札ペアでもまずペアスラップで処理する)
	if g.IsSlappable() {
		g.scheduleCpuSlap()
		g.checkStuck()
		return nil
	}

	if IsFaceCard(c.GetValue()) {
		// 絵札を出した → 相手にチャンスバトルを課す
		flipper := g.currentTurnIdx
		opp := g.opponentIdx(flipper)
		g.chanceFromIdx = flipper
		g.chanceRemaining = FaceCardChances(c.GetValue())
		// 手番は相手 (チャンスを払う側) に移る
		if g.players[opp].HasStock() {
			g.currentTurnIdx = opp
			g.maybeScheduleCpuStep()
		} else {
			// 相手のストックが既に空なら絵札側が即勝利 (空 stock 側は flip も slap もできない)
			g.applyChanceWin(flipper)
		}
		g.checkStuck()
		return nil
	}

	// 非絵札カード
	if g.chanceRemaining > 0 {
		g.chanceRemaining--
		if g.chanceRemaining == 0 {
			// 絵札側が場を獲得
			winner := g.chanceFromIdx
			g.applyChanceWin(winner)
			g.checkStuck()
			return nil
		}
		// 残チャンスがあるので、引き続き同じプレイヤー (相手) が flip
		g.maybeScheduleCpuStep()
		g.checkStuck()
		return nil
	}

	// 通常の手番交代。相手のストックが空なら手番を維持して相手復帰の余地を残す。
	nextIdx := g.opponentIdx(g.currentTurnIdx)
	if g.players[nextIdx].HasStock() {
		g.currentTurnIdx = nextIdx
	}
	g.maybeScheduleCpuStep()
	g.checkStuck()
	return nil
}

// Slap 指定プレイヤーがスラップを試みる。
//
// 場がペア/サンドイッチなら playerIdx がパイルを獲得し手番もそのプレイヤーになる
// (チャンスバトルも中断される)。成立していない場合はペナルティ 1 枚。
func (g *EgyptianRatscrew) Slap(playerIdx int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if playerIdx < 0 || playerIdx >= EgyptianRatscrewPlayerCnt {
		return ErrInvalidPlay
	}
	if len(g.centerPile) == 0 {
		return ErrInvalidPlay
	}
	if reason := g.slapReason(); reason != EgyptianRatscrewSlapReasonNone {
		g.applyCorrectSlap(playerIdx, reason)
	} else {
		g.applyWrongSlap(playerIdx)
	}
	g.checkStuck()
	return nil
}

// Tick CPU の保留アクションを (deadline 到達済みなら) 実行する。
//
// 戻り値は実際に発火したアクション種別 (発火しなかった場合は EgyptianRatscrewPendingNone)。
func (g *EgyptianRatscrew) Tick() EgyptianRatscrewPendingKind {
	if g.gameEndFlag {
		return EgyptianRatscrewPendingNone
	}
	if g.pending.Kind == EgyptianRatscrewPendingNone {
		return EgyptianRatscrewPendingNone
	}
	if g.now().UnixMilli() < g.pending.DeadlineMs {
		return EgyptianRatscrewPendingNone
	}
	kind := g.pending.Kind
	// 予約をクリアしてから実行 (実行中に新しい予約が組まれる場合もあるため)
	g.pending = EgyptianRatscrewPending{Kind: EgyptianRatscrewPendingNone}
	switch kind {
	case EgyptianRatscrewPendingStep:
		_ = g.Step()
	case EgyptianRatscrewPendingSlap:
		_ = g.Slap(g.cpuIdx())
	}
	return kind
}

// --- 内部ヘルパ ---

// applyCorrectSlap 正しいスラップ: パイルを獲得し手番もそのプレイヤーへ。
// チャンスバトル中でも中断して獲得できる。
func (g *EgyptianRatscrew) applyCorrectSlap(playerIdx int, reason EgyptianRatscrewSlapReason) {
	got := g.centerPile
	g.centerPile = nil
	// パイル全体をシャッフルして勝者ストックの底に裏向きで戻す
	g.rng.Shuffle(len(got), func(i, j int) { got[i], got[j] = got[j], got[i] })
	g.players[playerIdx].AddToStockBottom(got...)
	g.currentTurnIdx = playerIdx
	g.chanceRemaining = 0
	g.chanceFromIdx = -1
	g.lastEvent = EgyptianRatscrewLastEvent{
		Kind:       EgyptianRatscrewEventSlapCorrect,
		PlayerIdx:  playerIdx,
		SlapReason: reason,
	}
	g.appendLog(playerIdx, "slap", fmt.Sprintf("correct slap (%s), +%d cards", slapReasonLabel(reason), len(got)), nil)
	g.pending = EgyptianRatscrewPending{Kind: EgyptianRatscrewPendingNone}
	g.maybeScheduleCpuStep()
}

// applyWrongSlap 誤スラップ: ペナルティ 1 枚を相手に渡す。
func (g *EgyptianRatscrew) applyWrongSlap(playerIdx int) {
	offender := g.players[playerIdx]
	opp := g.opponentIdx(playerIdx)
	moved := 0
	for range EgyptianRatscrewPenaltyCount {
		c := offender.DrawTop()
		if c == nil {
			break
		}
		g.players[opp].AddToStockBottom(c)
		moved++
	}
	g.lastEvent = EgyptianRatscrewLastEvent{Kind: EgyptianRatscrewEventSlapWrong, PlayerIdx: playerIdx}
	g.appendLog(playerIdx, "slap", fmt.Sprintf("wrong slap, -%d cards", moved), nil)
	if !offender.HasStock() {
		g.endGame(opp)
		return
	}
	// スラップ条件は崩れていない (上 2/3 が変わらない) ので、CPU の slap 予約は維持して良いが、
	// 相手も連続スラップのチャンスがあるため、現状の予約は破棄して再予約に任せる。
	g.pending = EgyptianRatscrewPending{Kind: EgyptianRatscrewPendingNone}
	if g.IsSlappable() {
		g.scheduleCpuSlap()
	} else {
		g.maybeScheduleCpuStep()
	}
}

// applyChanceWin チャンスバトルで絵札側が場を獲得する。
func (g *EgyptianRatscrew) applyChanceWin(winnerIdx int) {
	got := g.centerPile
	g.centerPile = nil
	g.rng.Shuffle(len(got), func(i, j int) { got[i], got[j] = got[j], got[i] })
	g.players[winnerIdx].AddToStockBottom(got...)
	g.currentTurnIdx = winnerIdx
	g.chanceRemaining = 0
	g.chanceFromIdx = -1
	g.lastEvent = EgyptianRatscrewLastEvent{Kind: EgyptianRatscrewEventChanceWin, PlayerIdx: winnerIdx}
	g.appendLog(winnerIdx, "chance", fmt.Sprintf("chance win, +%d cards", len(got)), nil)
	g.pending = EgyptianRatscrewPending{Kind: EgyptianRatscrewPendingNone}
	g.maybeScheduleCpuStep()
}

// maybeScheduleCpuStep 手番が CPU のとき、Step を一定遅延で予約する。
// すでに予約があれば上書きしない。
func (g *EgyptianRatscrew) maybeScheduleCpuStep() {
	if g.gameEndFlag {
		return
	}
	if g.pending.Kind != EgyptianRatscrewPendingNone {
		return
	}
	if g.players[g.currentTurnIdx].GetIsHuman() {
		return
	}
	g.pending = EgyptianRatscrewPending{
		Kind:       EgyptianRatscrewPendingStep,
		DeadlineMs: g.now().UnixMilli() + int64(g.drawReactionMs()),
	}
}

// scheduleCpuSlap 場がスラップ可能な状態で CPU の slap を予約する。
func (g *EgyptianRatscrew) scheduleCpuSlap() {
	if g.gameEndFlag {
		return
	}
	g.pending = EgyptianRatscrewPending{
		Kind:       EgyptianRatscrewPendingSlap,
		DeadlineMs: g.now().UnixMilli() + int64(g.drawReactionMs()),
	}
}

// drawReactionMs 設定難易度に基づき正規分布から反応時間 (ms) を抽出する。
// 下限 EgyptianRatscrewMinReactionMs でクランプ。
func (g *EgyptianRatscrew) drawReactionMs() int {
	mean := float64(g.config.ReactionMeanMs())
	std := float64(g.config.ReactionStdDevMs())
	v := mean + g.rng.NormFloat64()*std
	if v < float64(EgyptianRatscrewMinReactionMs) {
		v = float64(EgyptianRatscrewMinReactionMs)
	}
	return int(v)
}

// opponentIdx 相手プレイヤーのインデックスを返す。
func (g *EgyptianRatscrew) opponentIdx(idx int) int {
	if idx == 0 {
		return 1
	}
	return 0
}

// cpuIdx CPU プレイヤーのインデックスを返す。
// NewDefaultEgyptianRatscrew の規約により player[0] = 人間, player[1] = CPU で固定。
func (g *EgyptianRatscrew) cpuIdx() int { return 1 }

// endGame ゲームを終了し勝者を確定する。
func (g *EgyptianRatscrew) endGame(winner int) {
	g.gameEndFlag = true
	g.phase = EgyptianRatscrewPhaseGameEnd
	g.winnerIdx = winner
	g.pending = EgyptianRatscrewPending{Kind: EgyptianRatscrewPendingNone}
	g.chanceRemaining = 0
	g.chanceFromIdx = -1
	if winner >= 0 && winner < EgyptianRatscrewPlayerCnt {
		g.players[winner].SetIsFinished(true)
	}
}

// checkStuck 停滞検出 — 両者ストックが空でかつ場のスラップ条件もチャンスバトルも
// 残っていないケースで終了させる (誰も flip できず slap も起き得ないので進まない)。
func (g *EgyptianRatscrew) checkStuck() {
	if g.gameEndFlag {
		return
	}
	if g.players[0].HasStock() || g.players[1].HasStock() {
		return
	}
	// 両者空。スラップ可能か、チャンスバトル中ならまだ進む余地がある。
	if g.IsSlappable() {
		return
	}
	if g.chanceRemaining > 0 {
		// 払う側 (currentTurnIdx) のストックが空なら払えないので絵札側勝利、それ以外は復帰可能
		if !g.players[g.currentTurnIdx].HasStock() {
			g.applyChanceWin(g.chanceFromIdx)
		}
		return
	}
	// 両者ストック空 + スラップ不可 + チャンスバトルなしで詰み。引き分けとする。
	g.endGame(-1)
}

// appendLog 棋譜にエントリを追加する
func (g *EgyptianRatscrew) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// slapReasonLabel スラップ理由のログ用ラベル
func slapReasonLabel(r EgyptianRatscrewSlapReason) string {
	switch r {
	case EgyptianRatscrewSlapReasonPair:
		return "pair"
	case EgyptianRatscrewSlapReasonSandwich:
		return "sandwich"
	default:
		return "none"
	}
}

// --- JSON ---

// egyptianRatscrewJSON is the JSON wire format for EgyptianRatscrew.
type egyptianRatscrewJSON struct {
	TrumpCards      *TrumpCards                                        `json:"tc"`
	Players         [EgyptianRatscrewPlayerCnt]*EgyptianRatscrewPlayer `json:"ps"`
	Config          EgyptianRatscrewConfig                             `json:"cf"`
	Phase           EgyptianRatscrewPhase                              `json:"ph"`
	CenterPile      []*Card                                            `json:"cp"`
	CurrentTurnIdx  int                                                `json:"ct"`
	ChanceRemaining int                                                `json:"cr"`
	ChanceFromIdx   int                                                `json:"ci"`
	Pending         EgyptianRatscrewPending                            `json:"pd"`
	LastEvent       EgyptianRatscrewLastEvent                          `json:"le"`
	GameEndFlag     bool                                               `json:"ge"`
	WinnerIdx       int                                                `json:"wi"`
	ActionLog       []*ActionLogEntry                                  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *EgyptianRatscrew) MarshalJSON() ([]byte, error) {
	return json.Marshal(egyptianRatscrewJSON{
		TrumpCards:      g.trumpCards,
		Players:         g.players,
		Config:          g.config,
		Phase:           g.phase,
		CenterPile:      g.centerPile,
		CurrentTurnIdx:  g.currentTurnIdx,
		ChanceRemaining: g.chanceRemaining,
		ChanceFromIdx:   g.chanceFromIdx,
		Pending:         g.pending,
		LastEvent:       g.lastEvent,
		GameEndFlag:     g.gameEndFlag,
		WinnerIdx:       g.winnerIdx,
		ActionLog:       g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *EgyptianRatscrew) UnmarshalJSON(data []byte) error {
	var j egyptianRatscrewJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.centerPile = j.CenterPile
	g.currentTurnIdx = j.CurrentTurnIdx
	g.chanceRemaining = j.ChanceRemaining
	g.chanceFromIdx = j.ChanceFromIdx
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
