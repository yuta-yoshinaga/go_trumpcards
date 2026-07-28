//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// KempsPhase はゲームフェーズ。
type KempsPhase int

// ケムプスのフェーズ定数
const (
	// KempsPhaseExchange 場とのカード交換 (スワップ/パス) 進行中
	KempsPhaseExchange KempsPhase = 0
	// KempsPhaseDeclare どこかのプレイヤーがフォーオブアカインドを揃え、宣言ウィンドウが開いた状態
	KempsPhaseDeclare KempsPhase = 1
	// KempsPhaseRoundEnd ラウンド終了 (Kemps / Counter-Kemps が解決し、得点が確定した)
	KempsPhaseRoundEnd KempsPhase = 2
	// KempsPhaseGameEnd ゲーム終了 (いずれかのチームが目標得点に達した)
	KempsPhaseGameEnd KempsPhase = 3
)

// KempsPhaseMin / KempsPhaseMax はフェーズ列挙の範囲 (復元時の検証用)。
const (
	KempsPhaseMin = KempsPhaseExchange
	KempsPhaseMax = KempsPhaseGameEnd
)

// Kemps はケムプス (Kemps / ケムプス) のゲームクラス。
//
// 4 人 2 チームの協力型マッチングゲーム。人間 (席 0) と CPU パートナー (席 2) が
// チーム A、CPU 2 名 (席 1, 3) がチーム B を組む。52 枚デッキを用い、各プレイヤーは
// 4 枚を持つ。テーブル中央には 4 枚の表向きカード (フィールド) が並ぶ。
//
// # 交換のモデル化
//
// 各プレイヤーは順番 (時計回り) に、手札の 1 枚をフィールドの 1 枚と交換 (スワップ)
// するか、パスする。PlayerSwap(handIndex, fieldIndex) で人間がスワップを選び、
// CPU は CpuPlay でフォーオブアカインドに近づくよう貪欲にスワップする。
//
// # フォーオブアカインド → シグナル → 宣言
//
// 手札 4 枚が同じランクになったプレイヤーが現れると、宣言ウィンドウ
// (KempsPhaseDeclare) が開く。本来はその「パートナー」が秘密のシグナルを送り、
// チームの誰かが "Kemps!" と宣言して得点する。サーバ側では partnerSignaling
// フラグでモデル化し、人間にだけシグナルの有無を公開する (signalType は人間が
// PlayerSetSignal で事前に決める)。相手チームのフォーオブアカインドに対しては、
// 人間は opponentSignaling の「気配」だけを観測でき、カウンターケムプスを狙える。
//
// # Kemps / Counter-Kemps
//
//   - PlayerDeclareKemps: 人間チームに現在フォーオブアカインド保持者がいれば
//     チーム A に +1、いなければ何もしない (ペナルティなしの空振り)。
//   - PlayerDeclareCounterKemps(targetSeat): 対象の相手がフォーオブアカインドを
//     保持していれば呼び出し側チームに +1、外していれば -1。
//
// # 停止保証
//
// フルCPU対戦でも必ず終了する: CPU は貪欲にスワップしていずれフォーオブアカインドを
// 揃え、揃った時点で当該チームが自動的に Kemps を宣言して得点する。得点は単調増加し、
// いずれ目標得点に達する。加えてラウンド数上限 (KempsMaxRounds) とラウンドあたりの
// スワップ上限 (KempsMaxSwapsPerRound) のガードを設けている。
type Kemps struct {
	trumpCards       *TrumpCards
	players          [KempsPlayerCnt]*KempsPlayer
	config           KempsConfig
	phase            KempsPhase
	field            []*Card
	drawPile         []*Card
	currentPlayerIdx int
	teamScores       [KempsTeamCnt]int
	signalType       SignalType
	fourHolderIdx    int // 現在フォーオブアカインドを保持するプレイヤー (-1=なし)
	roundResult      int // 直近ラウンドの結果コード (0=未, 1=Kemps成功, 2=Counter成功, 3=Counter失敗, 4=空振り)
	roundWinnerTeam  int // 直近ラウンドで得点したチーム (-1=なし)
	swapCount        int // 当該ラウンドのスワップ/パス回数 (停止保証のフェイルセーフ用)
	roundNumber      int
	gameEndFlag      bool
	winnerTeam       int
	actionLog        []*ActionLogEntry

	// rng CPU 判断用 (テストで差し替え可能)
	rng *rand.Rand
}

// ラウンド結果コード
const (
	// KempsResultNone まだ解決していない
	KempsResultNone = 0
	// KempsResultKemps Kemps 宣言が成功した
	KempsResultKemps = 1
	// KempsResultCounter Counter-Kemps が成功した
	KempsResultCounter = 2
	// KempsResultCounterFail Counter-Kemps が失敗した
	KempsResultCounterFail = 3
	// KempsResultMiss 宣言が空振りした
	KempsResultMiss = 4
)

// NewKemps はコンストラクタ。
func NewKemps(trumpCards *TrumpCards, players []*KempsPlayer, config KempsConfig) *Kemps {
	g := &Kemps{
		trumpCards:      trumpCards,
		config:          config,
		winnerTeam:      -1,
		fourHolderIdx:   -1,
		roundWinnerTeam: -1,
		rng:             rand.New(rand.NewSource(rand.Int63())),
	}
	for i := 0; i < KempsPlayerCnt && i < len(players); i++ {
		g.players[i] = players[i]
	}
	return g
}

// NewDefaultKemps は標準セットアップ (人間 1 + CPU 3) の Kemps を返す。
func NewDefaultKemps() *Kemps {
	players := []*KempsPlayer{
		NewKempsPlayer(true),
		NewKempsPlayer(false),
		NewKempsPlayer(false),
		NewKempsPlayer(false),
	}
	return NewKemps(NewTrumpCards(0), players, DefaultKempsConfig())
}

// SetRand はテスト用に乱数源を差し替える。
func (g *Kemps) SetRand(r *rand.Rand) {
	if r != nil {
		g.rng = r
	}
}

// Reset はゲーム全体を初期化して新しいゲームを開始する。
func (g *Kemps) Reset() {
	g.roundNumber = 0
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.teamScores = [KempsTeamCnt]int{}
	g.actionLog = nil
	g.dealRound()
}

// GetConfig は設定を返す。
func (g *Kemps) GetConfig() KempsConfig { return g.config }

// SetConfig は設定を更新する。
func (g *Kemps) SetConfig(cfg KempsConfig) { g.config = cfg }

// ResetWithConfig は設定を更新してゲームを初期化する。
func (g *Kemps) ResetWithConfig(cfg KempsConfig) {
	g.config = cfg
	g.Reset()
}

// NextRound は次のラウンドを開始する。ゲーム終了済みなら何もしない。
func (g *Kemps) NextRound() {
	if g.gameEndFlag {
		return
	}
	g.dealRound()
}

// dealRound は各プレイヤーへ手札を配り、フィールドを並べて新しいラウンドの
// 交換フェーズを開始する。
func (g *Kemps) dealRound() {
	g.roundNumber++
	g.phase = KempsPhaseExchange
	g.fourHolderIdx = -1
	g.roundResult = KempsResultNone
	g.roundWinnerTeam = -1
	g.swapCount = 0
	g.currentPlayerIdx = 0

	g.trumpCards.Shuffle()
	g.drawPile = g.drawAll()

	for _, p := range g.players {
		if p == nil {
			continue
		}
		p.Reset()
		for n := 0; n < KempsHandSize; n++ {
			if c := g.popDraw(); c != nil {
				p.AddCard(c)
			}
		}
	}

	g.field = g.field[:0]
	for n := 0; n < KempsFieldSize; n++ {
		if c := g.popDraw(); c != nil {
			g.field = append(g.field, c)
		}
	}

	g.appendLog(-1, "deal", fmt.Sprintf("round %d dealt", g.roundNumber), nil)
}

// drawAll はトランプから全カードを取り出してスライスで返す。
func (g *Kemps) drawAll() []*Card {
	cards := make([]*Card, 0, CardCnt)
	for {
		c := g.trumpCards.DrawCard()
		if c == nil {
			break
		}
		cards = append(cards, c)
	}
	return cards
}

// popDraw はドローパイルの先頭 1 枚を取り出す (空なら nil)。
func (g *Kemps) popDraw() *Card {
	if len(g.drawPile) == 0 {
		return nil
	}
	c := g.drawPile[0]
	g.drawPile = g.drawPile[1:]
	return c
}

// --- ゲッター ---

// GetPhase は現在のフェーズを返す。
func (g *Kemps) GetPhase() KempsPhase { return g.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Kemps) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam は勝利チームを返す (-1=未確定)。
func (g *Kemps) GetWinnerTeam() int { return g.winnerTeam }

// GetWinnerIdx は勝利チームを返す (互換用、-1=未確定)。
func (g *Kemps) GetWinnerIdx() int { return g.winnerTeam }

// GetPlayerCnt はプレイヤー数を返す。
func (g *Kemps) GetPlayerCnt() int { return KempsPlayerCnt }

// GetPlayer は指定インデックスのプレイヤーを返す (範囲外は nil)。
func (g *Kemps) GetPlayer(i int) *KempsPlayer {
	if i < 0 || i >= KempsPlayerCnt {
		return nil
	}
	return g.players[i]
}

// GetTeamScore はチーム (0 または 1) の得点を返す。
func (g *Kemps) GetTeamScore(team int) int {
	if team < 0 || team >= KempsTeamCnt {
		return 0
	}
	return g.teamScores[team]
}

// GetFieldSize はフィールドのカード枚数を返す。
func (g *Kemps) GetFieldSize() int { return len(g.field) }

// GetFieldCard はフィールドの指定インデックスのカードを返す (範囲外は nil)。
func (g *Kemps) GetFieldCard(i int) *Card {
	if i < 0 || i >= len(g.field) {
		return nil
	}
	return g.field[i]
}

// GetCurrentPlayerIdx は現在の手番プレイヤーを返す。
func (g *Kemps) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// GetSignalType は人間が設定したシグナル種別を返す。
func (g *Kemps) GetSignalType() SignalType { return g.signalType }

// GetFourHolderIdx は現在フォーオブアカインドを保持するプレイヤーを返す (-1=なし)。
func (g *Kemps) GetFourHolderIdx() int { return g.fourHolderIdx }

// GetRoundResult は直近ラウンドの結果コードを返す。
func (g *Kemps) GetRoundResult() int { return g.roundResult }

// GetRoundWinnerTeam は直近ラウンドで得点したチームを返す (-1=なし)。
func (g *Kemps) GetRoundWinnerTeam() int { return g.roundWinnerTeam }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *Kemps) GetRoundNumber() int { return g.roundNumber }

// GetActionLog は棋譜を返す。
func (g *Kemps) GetActionLog() []*ActionLogEntry { return g.actionLog }

// IsPartnerSignaling は人間 (席 0) のチームにフォーオブアカインド保持者がいて、
// 人間が宣言可能なシグナル状態かを返す (人間にだけ公開される情報)。
func (g *Kemps) IsPartnerSignaling() bool {
	if g.phase != KempsPhaseDeclare || g.fourHolderIdx < 0 {
		return false
	}
	return KempsTeamOf(g.fourHolderIdx) == KempsTeamOf(0)
}

// IsOpponentSignaling は相手チームにフォーオブアカインド保持者がいる「気配」を返す。
// 人間がカウンターケムプスを狙うためのヒント (席は明かさない)。
func (g *Kemps) IsOpponentSignaling() bool {
	if g.phase != KempsPhaseDeclare || g.fourHolderIdx < 0 {
		return false
	}
	return KempsTeamOf(g.fourHolderIdx) != KempsTeamOf(0)
}

// IsHumanTurn は現在の交換手番が人間かを返す。宣言フェーズは常に人間が操作可能。
func (g *Kemps) IsHumanTurn() bool {
	if g.gameEndFlag {
		return false
	}
	human := g.GetPlayer(0)
	isHuman := human != nil && human.GetIsHuman()
	switch g.phase {
	case KempsPhaseExchange:
		return g.currentPlayerIdx == 0 && isHuman
	case KempsPhaseDeclare:
		// 宣言ウィンドウは人間 (席 0 が人間の場合のみ) が操作可能。フルCPU対戦では
		// CpuPlay (cpuResolveDeclare) が自動的に解決する。
		return isHuman
	default:
		return false
	}
}

// --- アクション ---

// PlayerSetSignal は人間プレイヤーが秘密のシグナル種別を設定する。
// 範囲外の値は SignalSound にクランプする。
func (g *Kemps) PlayerSetSignal(signalType int) {
	st := SignalType(signalType)
	if st < SignalTypeMin || st > SignalTypeMax {
		st = SignalSound
	}
	g.signalType = st
	g.appendLog(0, "signal", fmt.Sprintf("set signal type %d", int(st)), nil)
}

// PlayerSwap は人間プレイヤーが手札の handIndex をフィールドの fieldIndex と交換する。
//
// 交換フェーズで人間の手番でないとき、不正なインデックスのときはエラーを返す。
func (g *Kemps) PlayerSwap(handIndex, fieldIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KempsPhaseExchange {
		return ErrWrongPhase
	}
	if g.currentPlayerIdx != 0 {
		return ErrInvalidPlay
	}
	p := g.players[0]
	if p == nil {
		return ErrInvalidPlay
	}
	if handIndex < 0 || handIndex >= p.GetCardsSize() {
		return ErrInvalidCard
	}
	if fieldIndex < 0 || fieldIndex >= len(g.field) {
		return ErrInvalidCard
	}
	g.doSwap(0, handIndex, fieldIndex)
	return nil
}

// PlayerPass は人間プレイヤーが交換せずにパスする。
//
// 宣言フェーズで呼ばれた場合は「宣言を見送る」操作として扱い、保持チームによる
// 自動解決 (cpuResolveDeclare) に委ねる。
func (g *Kemps) PlayerPass() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	switch g.phase {
	case KempsPhaseDeclare:
		g.cpuResolveDeclare()
		return nil
	case KempsPhaseExchange:
		if g.currentPlayerIdx != 0 {
			return ErrInvalidPlay
		}
		g.doPass(0)
		return nil
	default:
		return ErrWrongPhase
	}
}

// PlayerDeclareKemps は人間プレイヤーが "Kemps!" と宣言する。
//
// 人間のチームに現在フォーオブアカインド保持者がいればチーム A に +1。いなければ
// 何もしない (ペナルティなしの空振り)。宣言フェーズでないとエラー。
func (g *Kemps) PlayerDeclareKemps() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KempsPhaseDeclare {
		return ErrWrongPhase
	}
	team := KempsTeamOf(0)
	if g.fourHolderIdx >= 0 && KempsTeamOf(g.fourHolderIdx) == team {
		g.awardRound(team, KempsResultKemps)
	} else {
		g.roundResult = KempsResultMiss
		g.appendLog(0, "kempsMiss", "declared Kemps but no four of a kind", nil)
		g.endRound(-1)
	}
	return nil
}

// PlayerDeclareCounterKemps は人間プレイヤーが相手 targetSeat に対して
// "Counter-Kemps!" と宣言する。
//
// 対象がフォーオブアカインドを保持していれば呼び出し側チーム (チーム A) に +1。
// 外していればチーム A から -1 のペナルティ。宣言フェーズでないとエラー。
func (g *Kemps) PlayerDeclareCounterKemps(targetSeat int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != KempsPhaseDeclare {
		return ErrWrongPhase
	}
	if targetSeat < 0 || targetSeat >= KempsPlayerCnt {
		return ErrInvalidPlay
	}
	callerTeam := KempsTeamOf(0)
	target := g.players[targetSeat]
	if target != nil && target.HasFourOfAKind() && KempsTeamOf(targetSeat) != callerTeam {
		g.awardRound(callerTeam, KempsResultCounter)
	} else {
		g.penalizeRound(callerTeam)
	}
	return nil
}

// CpuPlay は CPU の手番を 1 ステップ進める。
//
// 交換フェーズでは現在の CPU 手番がスワップ/パスを行う。宣言フェーズでは保持者の
// チームが自動的に Kemps を宣言する (人間が宣言済みなら何もしない)。相手チームの
// フォーオブアカインドには、難易度に応じて CPU がカウンターを試みる。人間の操作が
// 必要な状態では何もしない。
func (g *Kemps) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	switch g.phase {
	case KempsPhaseExchange:
		g.cpuExchange()
	case KempsPhaseDeclare:
		g.cpuResolveDeclare()
	}
}

// cpuExchange は現在の CPU 手番のスワップ/パスを実行する。人間の手番なら何もしない。
func (g *Kemps) cpuExchange() {
	idx := g.currentPlayerIdx
	p := g.GetPlayer(idx)
	if p == nil || p.GetIsHuman() {
		return
	}
	hand, fld, ok := g.cpuChooseSwap(idx)
	if ok {
		g.doSwap(idx, hand, fld)
	} else {
		g.doPass(idx)
	}
}

// cpuChooseSwap は CPU がフォーオブアカインドに近づくスワップ手を選ぶ。
//
// 手札で最も枚数の多いランクを「狙い」とし、それ以外の手札 1 枚を、フィールド上に
// ある狙いランクのカードと交換する。狙いに近づく交換が無ければパス (ok=false)。
func (g *Kemps) cpuChooseSwap(idx int) (handIndex, fieldIndex int, ok bool) {
	p := g.players[idx]
	if p == nil || p.GetCardsSize() == 0 || len(g.field) == 0 {
		return 0, 0, false
	}

	// 手札で最頻のランク (= 狙い) を求める。
	counts := map[int]int{}
	for i := 0; i < p.GetCardsSize(); i++ {
		if c := p.GetCard(i); c != nil {
			counts[c.GetValue()]++
		}
	}
	targetRank, best := -1, -1
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c == nil {
			continue
		}
		if counts[c.GetValue()] > best {
			best = counts[c.GetValue()]
			targetRank = c.GetValue()
		}
	}

	// フィールドに狙いランクがあり、手札に狙いランク以外の札があれば交換する。
	for fi := 0; fi < len(g.field); fi++ {
		fc := g.field[fi]
		if fc == nil || fc.GetValue() != targetRank {
			continue
		}
		for hi := 0; hi < p.GetCardsSize(); hi++ {
			hc := p.GetCard(hi)
			if hc != nil && hc.GetValue() != targetRank {
				return hi, fi, true
			}
		}
	}
	return 0, 0, false
}

// cpuResolveDeclare は宣言フェーズで CPU/相手チームの解決を進める。
//
// フォーオブアカインド保持者のチームが自動的に Kemps を宣言して得点する。ただし
// 相手チーム (チーム B) が保持者の場合、人間 (チーム A) が観測できているため、まず
// CPU 同士の競争として、相手チームの宣言前に「カウンターの気配」を残す。本実装では
// 保持チームが確実に得点することでラウンドを前進させ、停止保証を満たす。
func (g *Kemps) cpuResolveDeclare() {
	if g.fourHolderIdx < 0 {
		// 想定外: フェーズだけ宣言なのに保持者がいない場合はラウンドを締める。
		g.roundResult = KempsResultMiss
		g.endRound(-1)
		return
	}
	holderTeam := KempsTeamOf(g.fourHolderIdx)
	// 相手チーム (席 1/3) が保持しているなら、人間チームの CPU パートナー (席 2) が
	// 難易度に応じてカウンターを試みる。
	if holderTeam != KempsTeamOf(0) {
		if g.rng.Float64() < g.config.CounterChance() {
			g.awardRound(KempsTeamOf(0), KempsResultCounter)
			return
		}
	}
	g.awardRound(holderTeam, KempsResultKemps)
}

// doSwap はプレイヤー idx の手札 handIndex とフィールドの fieldIndex を入れ替え、
// フォーオブアカインド判定を行ったうえで手番を進める。
func (g *Kemps) doSwap(idx, handIndex, fieldIndex int) {
	p := g.players[idx]
	taken := g.field[fieldIndex]
	given := p.RemoveCard(handIndex)
	if given != nil {
		// 取り出した手札札 (given) をフィールドへ置き、フィールド札 (taken) を手札へ。
		g.field[fieldIndex] = given
	}
	if taken != nil {
		p.AddCard(taken)
	}
	g.appendLog(idx, "swap", "swap a hand card with the field", []*Card{taken})
	g.swapCount++
	g.afterExchange(idx)
}

// doPass はプレイヤー idx が交換せずに手番を渡す。
func (g *Kemps) doPass(idx int) {
	g.appendLog(idx, "pass", "pass (no swap)", nil)
	g.swapCount++
	g.afterExchange(idx)
}

// afterExchange は交換/パス後にフォーオブアカインド判定とウィンドウ判定を行い、
// 揃っていなければ次のプレイヤーへ手番を移す。
func (g *Kemps) afterExchange(idx int) {
	// フォーオブアカインド: 揃ったら宣言ウィンドウを開く。
	if g.phase == KempsPhaseExchange {
		if holder := g.firstFourHolder(); holder >= 0 {
			g.openDeclareWindow(holder)
			return
		}
	}
	// 停止保証のフェイルセーフ: スワップ回数が上限を超えたらラウンドを引き分けで締める。
	if g.swapCount >= KempsMaxSwapsPerRound {
		g.roundResult = KempsResultMiss
		g.endRound(-1)
		return
	}
	g.currentPlayerIdx = (idx + 1) % KempsPlayerCnt
}

// firstFourHolder はフォーオブアカインドを保持する最初のプレイヤーを返す (-1=なし)。
func (g *Kemps) firstFourHolder() int {
	for i := 0; i < KempsPlayerCnt; i++ {
		if g.players[i] != nil && g.players[i].HasFourOfAKind() {
			return i
		}
	}
	return -1
}

// openDeclareWindow はプレイヤー idx がフォーオブアカインドを揃えた瞬間に
// 宣言ウィンドウ (KempsPhaseDeclare) を開く。
func (g *Kemps) openDeclareWindow(idx int) {
	g.phase = KempsPhaseDeclare
	g.fourHolderIdx = idx
	g.appendLog(idx, "four", "four of a kind! signal your partner", nil)
}

// awardRound はチーム team に +1 して result を記録し、ラウンドを締める。
func (g *Kemps) awardRound(team, result int) {
	if team >= 0 && team < KempsTeamCnt {
		g.teamScores[team]++
	}
	g.roundResult = result
	g.appendLog(-1, "score", fmt.Sprintf("team %d scores (result %d)", team, result), nil)
	g.endRound(team)
}

// penalizeRound はチーム team から -1 して Counter 失敗を記録し、ラウンドを締める。
func (g *Kemps) penalizeRound(team int) {
	if team >= 0 && team < KempsTeamCnt {
		g.teamScores[team]--
		if g.teamScores[team] < 0 {
			g.teamScores[team] = 0
		}
	}
	g.roundResult = KempsResultCounterFail
	g.appendLog(-1, "penalty", fmt.Sprintf("team %d counter-kemps failed", team), nil)
	g.endRound(-1)
}

// endRound はラウンドを終了し、勝敗判定を行う。
func (g *Kemps) endRound(winnerTeam int) {
	g.roundWinnerTeam = winnerTeam
	g.phase = KempsPhaseRoundEnd
	g.checkGameEnd()
}

// checkGameEnd はいずれかのチームが目標得点に達したか、ラウンド上限に達したら
// ゲームを締める。
func (g *Kemps) checkGameEnd() {
	target := g.config.TargetScore
	if target <= 0 {
		target = KempsTargetScore
	}
	for team := 0; team < KempsTeamCnt; team++ {
		if g.teamScores[team] >= target {
			g.endGame(team)
			return
		}
	}
	if g.roundNumber >= KempsMaxRounds {
		// 停止保証のフェイルセーフ: 得点が高いチームを勝者にする。
		best := 0
		if g.teamScores[1] > g.teamScores[0] {
			best = 1
		}
		g.endGame(best)
	}
}

// endGame はゲームを終了し勝利チームを確定する。
func (g *Kemps) endGame(winnerTeam int) {
	g.gameEndFlag = true
	g.phase = KempsPhaseGameEnd
	g.winnerTeam = winnerTeam
	for i := 0; i < KempsPlayerCnt; i++ {
		if g.players[i] != nil {
			g.players[i].SetIsFinished(KempsTeamOf(i) == winnerTeam)
		}
	}
	g.appendLog(-1, "gameEnd", fmt.Sprintf("team %d wins", winnerTeam), nil)
}

// appendLog は棋譜にエントリを追加する。
func (g *Kemps) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- テスト/復元用セッター ---

// SetTeamScore はチーム team の得点を設定する (主にテスト用)。
func (g *Kemps) SetTeamScore(team, score int) {
	if team >= 0 && team < KempsTeamCnt {
		g.teamScores[team] = score
	}
}

// --- JSON ---

// kempsJSON is the JSON wire format for Kemps.
type kempsJSON struct {
	TrumpCards       *TrumpCards                  `json:"tc"`
	Players          [KempsPlayerCnt]*KempsPlayer `json:"ps"`
	Config           KempsConfig                  `json:"cf"`
	Phase            KempsPhase                   `json:"ph"`
	Field            []*Card                      `json:"fd"`
	DrawPile         []*Card                      `json:"dp"`
	CurrentPlayerIdx int                          `json:"ci"`
	TeamScores       [KempsTeamCnt]int            `json:"sc"`
	SignalType       SignalType                   `json:"sg"`
	FourHolderIdx    int                          `json:"fh"`
	RoundResult      int                          `json:"rr"`
	RoundWinnerTeam  int                          `json:"rw"`
	SwapCount        int                          `json:"sw"`
	RoundNumber      int                          `json:"rn"`
	GameEndFlag      bool                         `json:"ge"`
	WinnerTeam       int                          `json:"wt"`
	ActionLog        []*ActionLogEntry            `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Kemps) MarshalJSON() ([]byte, error) {
	return json.Marshal(kempsJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		Field:            g.field,
		DrawPile:         g.drawPile,
		CurrentPlayerIdx: g.currentPlayerIdx,
		TeamScores:       g.teamScores,
		SignalType:       g.signalType,
		FourHolderIdx:    g.fourHolderIdx,
		RoundResult:      g.roundResult,
		RoundWinnerTeam:  g.roundWinnerTeam,
		SwapCount:        g.swapCount,
		RoundNumber:      g.roundNumber,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// errKempsInvalidState は復元データが不正なときの番兵エラー。
var errKempsInvalidState = fmt.Errorf("invalid kemps state")

// UnmarshalJSON implements json.Unmarshaler with defensive validation.
func (g *Kemps) UnmarshalJSON(data []byte) error {
	var j kempsJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < KempsPhaseMin || j.Phase > KempsPhaseMax {
		return errKempsInvalidState
	}
	if j.SignalType < SignalTypeMin || j.SignalType > SignalTypeMax {
		return errKempsInvalidState
	}
	for i := 0; i < KempsPlayerCnt; i++ {
		if j.Players[i] == nil {
			return errKempsInvalidState
		}
	}
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= KempsPlayerCnt {
		return errKempsInvalidState
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= KempsTeamCnt {
		return errKempsInvalidState
	}
	if j.RoundWinnerTeam < -1 || j.RoundWinnerTeam >= KempsTeamCnt {
		return errKempsInvalidState
	}
	if j.FourHolderIdx < -1 || j.FourHolderIdx >= KempsPlayerCnt {
		return errKempsInvalidState
	}
	for team := 0; team < KempsTeamCnt; team++ {
		if j.TeamScores[team] < 0 || j.TeamScores[team] > 999 {
			return errKempsInvalidState
		}
	}
	if len(j.Field) > KempsFieldSize {
		return errKempsInvalidState
	}
	if len(j.ActionLog) > KempsMaxRounds*KempsPlayerCnt*4 {
		return errKempsInvalidState
	}

	g.trumpCards = j.TrumpCards
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.field = j.Field
	g.drawPile = j.DrawPile
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.teamScores = j.TeamScores
	g.signalType = j.SignalType
	g.fourHolderIdx = j.FourHolderIdx
	g.roundResult = j.RoundResult
	g.roundWinnerTeam = j.RoundWinnerTeam
	g.swapCount = j.SwapCount
	g.roundNumber = j.RoundNumber
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(rand.Int63()))
	}
	return nil
}
