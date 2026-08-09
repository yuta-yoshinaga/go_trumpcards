//go:build !js || !wasm || casino

// Package domain ムス (Mus) のドメインモデル。
//
// Mus はスペイン・バスク地方発祥の 4 人 2 チーム戦の vying (賭け比べ) ゲーム。
// トリックテイキングではなく、4 つの賭けラウンド (Grande / Chica / Pares / Juego)
// で手の強さを賭け合う。各プレイヤーは 4 枚の手札を持ち、全員合意で手札交換 (Mus)
// を繰り返してから賭けに入る。先に目標点 (アマ, 既定 40) に到達したチームが勝利。
//
// このドメインは原典を簡素化した実装:
//   - 賭けは「チーム対チーム」で進行し、人間はチーム 0 (席 0,2)、CPU がチーム 1
//     (席 1,3) の判断を担う。
//   - パソ / エンビード (レイズ可) / オルダゴ、応答はキエロ / ノキエロ / レイズ。
//     両チーム・パソは流局 (showdown で +1)、ノキエロは賭け手 +1、キエロは
//     showdown で賭け点、オルダゴ受諾は即ゲーム決着。
//   - 手の評価: Grande=高い順、Chica=低い順、Pares=ペア類、Juego=31 点以上
//     (31>32>40..33)、無ければ Punto (30 に近い)。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// MusPlayerCnt プレイヤー数 (人間 1 + CPU 3)
const MusPlayerCnt = 4

// MusTeamCnt チーム数
const MusTeamCnt = 2

// MusHandSize 各プレイヤーの手札枚数
const MusHandSize = 4

// MusRoundCnt 賭けラウンド数 (Grande/Chica/Pares/Juego)
const MusRoundCnt = 4

// MusMaxMusCycles Mus 交換の最大繰り返し回数 (無限ループ防止)
const MusMaxMusCycles = 3

// MusJuegoThreshold Juego (31 以上で「ゲームあり」) のしきい値
const MusJuegoThreshold = 31

// MusPhase ゲームフェーズ
type MusPhase int

// Mus のフェーズ定数
const (
	// MusPhaseMus 各プレイヤーが Mus (交換希望) か Corte (賭け開始) を宣言する
	MusPhaseMus MusPhase = 0
	// MusPhaseDiscard 全員 Mus 合意時、各プレイヤーが交換札を選ぶ
	MusPhaseDiscard MusPhase = 1
	// MusPhaseGrande Grande (高い手) の賭け
	MusPhaseGrande MusPhase = 2
	// MusPhaseChica Chica (低い手) の賭け
	MusPhaseChica MusPhase = 3
	// MusPhasePares Pares (ペア) の賭け
	MusPhasePares MusPhase = 4
	// MusPhaseJuego Juego (31+) / Punto の賭け
	MusPhaseJuego MusPhase = 5
	// MusPhaseShowdown 手を公開し賭け点を精算する
	MusPhaseShowdown MusPhase = 6
	// MusPhaseRoundEnd ラウンド終了
	MusPhaseRoundEnd MusPhase = 7
	// MusPhaseGameEnd ゲーム終了
	MusPhaseGameEnd MusPhase = 8
)

// Mus の賭けアクション種別
const (
	// MusActionPaso パス
	MusActionPaso = 0
	// MusActionEnvido 賭け (レイズ)
	MusActionEnvido = 1
	// MusActionOrdago オルダゴ (全賭け)
	MusActionOrdago = 2
	// MusActionQuiero 賭けを受ける
	MusActionQuiero = 3
	// MusActionNoQuiero 賭けを降りる
	MusActionNoQuiero = 4
)

// MusRoundResultKind ラウンド結果の種別
const (
	// MusResultPending 未解決
	MusResultPending = 0
	// MusResultDeferred 両チーム・パソ (showdown で勝者 +1)
	MusResultDeferred = 1
	// MusResultAccepted キエロ (showdown で勝者が Stake 点)
	MusResultAccepted = 2
	// MusResultAwarded ノキエロ (賭け手チームが即 Stake 点)
	MusResultAwarded = 3
	// MusResultOrdago オルダゴ受諾 (showdown で勝者がゲーム勝利)
	MusResultOrdago = 4
	// MusResultSkipped 参加者なしで不成立 (Pares で誰もペアなし等)
	MusResultSkipped = 5
)

// MusRoundResult 1 賭けラウンドの結果
type MusRoundResult struct {
	Kind  int `json:"k"`
	Stake int `json:"s"`
	Team  int `json:"t"` // Awarded 種別で得点したチーム (-1=未確定)
}

// MusHint ヒント情報
type MusHint struct {
	Mus     bool   // Mus フェーズ: 交換推奨か
	Action  int    // 賭けフェーズ: 推奨アクション
	Amount  int    // Envido の推奨額
	Indices []int  // Discard フェーズ: 交換推奨札
	Reason  string // 理由キー
}

// Mus ムスのゲームクラス
type Mus struct {
	trumpCards  *TrumpCards
	players     []*MusPlayer
	config      MusConfig
	phase       MusPhase
	roundNumber int
	manoIdx     int             // 親 (先手) プレイヤー
	amarrakos   [MusTeamCnt]int // チーム別累積点
	results     [MusRoundCnt]MusRoundResult
	// Mus / Discard フェーズ
	musTurn     int     // 現在宣言中のプレイヤー
	musAgreed   int     // これまで Mus に同意した人数
	musCycle    int     // Mus 交換の繰り返し回数
	discardTurn int     // 現在交換中のプレイヤー
	discarded   []*Card // このラウンドの捨て札 (山札枯渇時の再シャッフル用)
	// 賭けフェーズ
	betTeam        int  // 現在アクションするチーム
	pendingStake   int  // 応答待ちの賭け額 (0=保留中の賭けなし)
	lastBettorTeam int  // 直近に賭けたチーム (-1=なし)
	firstActorPaso bool // 先手チームがパソ済みか (両パソ判定用)
	// 終了
	gameEndFlag bool
	winnerTeam  int // -1=未確定
	actionLogBase
}

// NewMus コンストラクタ
func NewMus(trumpCards *TrumpCards, players []*MusPlayer, config MusConfig) *Mus {
	return &Mus{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		winnerTeam:     -1,
		lastBettorTeam: -1,
	}
}

// NewDefaultMus 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultMus() *Mus {
	players := make([]*MusPlayer, MusPlayerCnt)
	players[0] = NewMusPlayer(true)
	for i := 1; i < MusPlayerCnt; i++ {
		players[i] = NewMusPlayer(false)
	}
	return NewMus(NewTrumpCardsBriscola(), players, DefaultMusConfig())
}

// MusTeamOf プレイヤーが属するチーム (0 = 席0&2, 1 = 席1&3)
func MusTeamOf(playerIdx int) int { return playerIdx % MusTeamCnt }

// Reset ゲーム初期化
func (g *Mus) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.manoIdx = 0
	g.amarrakos = [MusTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Mus) NextRound() {
	if g.phase != MusPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.manoIdx = (g.manoIdx + 1) % MusPlayerCnt
	g.startRound()
}

// startRound 手札を配り、Mus フェーズを開始する。
func (g *Mus) startRound() {
	g.results = [MusRoundCnt]MusRoundResult{}
	for i := range g.results {
		g.results[i].Team = -1
	}
	g.musCycle = 0
	g.discarded = nil
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()
	g.beginMusPhase()
}

// deal 各プレイヤーへ 4 枚配る。
func (g *Mus) deal() {
	for i := 0; i < MusHandSize; i++ {
		for j := 0; j < MusPlayerCnt; j++ {
			idx := (g.manoIdx + j) % MusPlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
	g.sortAllHands()
}

// beginMusPhase Mus 宣言フェーズを開始する。
func (g *Mus) beginMusPhase() {
	g.phase = MusPhaseMus
	g.musTurn = g.manoIdx
	g.musAgreed = 0
}

// --- Mus phase ---

// PlayerMus 人間が Mus (true=交換希望) / Corte (false=賭け開始) を宣言する。
func (g *Mus) PlayerMus(mus bool) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MusPhaseMus {
		return ErrWrongPhase
	}
	if !g.players[g.musTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.resolveMus(mus)
	return nil
}

// resolveMus Mus/Corte 宣言を反映する。
func (g *Mus) resolveMus(mus bool) {
	// 最大繰り返しに達したら強制的に賭け開始。
	if g.musCycle >= MusMaxMusCycles {
		mus = false
	}
	if !mus {
		g.appendLog(g.musTurn, "corte", fmt.Sprintf("%s cuts (no mus)", playerName(g.players, g.musTurn)), nil)
		g.beginBetting()
		return
	}
	g.appendLog(g.musTurn, "mus", fmt.Sprintf("%s wants mus", playerName(g.players, g.musTurn)), nil)
	g.musAgreed++
	if g.musAgreed >= MusPlayerCnt {
		// 全員合意 → 交換フェーズ。
		g.phase = MusPhaseDiscard
		g.discardTurn = g.manoIdx
		return
	}
	g.musTurn = (g.musTurn + 1) % MusPlayerCnt
}

// PlayerDiscard 人間が交換する札を選ぶ (0 枚も可)。
func (g *Mus) PlayerDiscard(indices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != MusPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.discardTurn].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if err := g.validateDiscard(indices); err != nil {
		return err
	}
	g.applyDiscard(indices)
	return nil
}

// validateDiscard 交換札インデックスの妥当性を検証する。
func (g *Mus) validateDiscard(indices []int) error {
	p := g.players[g.discardTurn]
	seen := make(map[int]bool, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= p.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidPlay, "同じ札を 2 回指定できません")
		}
		seen[idx] = true
	}
	return nil
}

// applyDiscard 交換札を捨てて引き直し、全員終わったら Mus フェーズへ戻る。
func (g *Mus) applyDiscard(indices []int) {
	p := g.players[g.discardTurn]
	if len(indices) > 0 {
		removed := p.RemoveCards(indices)
		// 引き直し: 山札優先、尽きたら捨て札パイル (このターンの捨て札は除外)
		// を再シャッフルして補充する。これにより手札が常に 4 枚に保たれる。
		for i := 0; i < len(indices); i++ {
			if c := g.drawForExchange(); c != nil {
				p.AddCard(c)
			}
		}
		// 自分の捨て札は引き直し後にパイルへ加える (同一ターンで引き戻さないため)。
		g.discarded = append(g.discarded, removed...)
		musSortHand(p)
	}
	g.appendLog(g.discardTurn, "discard",
		fmt.Sprintf("%s exchanges %d cards", playerName(g.players, g.discardTurn), len(indices)), nil)

	if g.discardTurn == (g.manoIdx+MusPlayerCnt-1)%MusPlayerCnt {
		// 全員交換完了 → 再び Mus 宣言へ。
		g.musCycle++
		g.beginMusPhase()
		return
	}
	g.discardTurn = (g.discardTurn + 1) % MusPlayerCnt
}

// drawForExchange 引き直し用に 1 枚引く。山札が尽きたら捨て札パイルを
// 再シャッフルして補充する (40 枚デッキでも手札が 4 枚未満にならない)。
func (g *Mus) drawForExchange() *Card {
	if c := g.trumpCards.DrawCard(); c != nil {
		return c
	}
	if len(g.discarded) == 0 {
		return nil
	}
	// 捨て札パイルをシャッフルして 1 枚取り出す。
	rand.Shuffle(len(g.discarded), func(i, j int) {
		g.discarded[i], g.discarded[j] = g.discarded[j], g.discarded[i]
	})
	c := g.discarded[len(g.discarded)-1]
	g.discarded = g.discarded[:len(g.discarded)-1]
	return c
}

// --- Betting ---

// beginBetting 最初の賭けラウンド (Grande) を開始する。
func (g *Mus) beginBetting() {
	g.phase = MusPhaseGrande
	g.startBetRound()
}

// roundIndex 現在の賭けラウンドのインデックス (0=Grande..3=Juego, -1=賭け外)。
func (g *Mus) roundIndex() int {
	switch g.phase {
	case MusPhaseGrande:
		return 0
	case MusPhaseChica:
		return 1
	case MusPhasePares:
		return 2
	case MusPhaseJuego:
		return 3
	default:
		return -1
	}
}

// startBetRound 現ラウンドの賭けを初期化する。参加チームが不足する Pares/Juego は
// 自動解決して次へ進む。
func (g *Mus) startBetRound() {
	g.pendingStake = 0
	g.lastBettorTeam = -1
	g.firstActorPaso = false
	g.betTeam = MusTeamOf(g.manoIdx)

	ri := g.roundIndex()
	if ri < 0 {
		return
	}
	// Pares / Juego の参加判定。
	if ri == 2 { // Pares
		t0, t1 := g.teamHasPares(0), g.teamHasPares(1)
		if !t0 && !t1 {
			g.results[ri] = MusRoundResult{Kind: MusResultSkipped, Team: -1}
			g.advanceRound()
			return
		}
		if t0 != t1 { // 片方のみ → 自動で +1
			win := 0
			if t1 {
				win = 1
			}
			g.results[ri] = MusRoundResult{Kind: MusResultAwarded, Stake: 1, Team: win}
			g.amarrakos[win]++
			g.checkGameEnd()
			if g.gameEndFlag {
				return
			}
			g.advanceRound()
			return
		}
	}
	if ri == 3 { // Juego
		t0, t1 := g.teamHasJuego(0), g.teamHasJuego(1)
		if t0 != t1 && (t0 || t1) { // 片方のみ Juego → 自動で +1
			win := 0
			if t1 {
				win = 1
			}
			g.results[ri] = MusRoundResult{Kind: MusResultAwarded, Stake: 1, Team: win}
			g.amarrakos[win]++
			g.checkGameEnd()
			if g.gameEndFlag {
				return
			}
			g.advanceRound()
			return
		}
		// 両方 Juego あり、または両方なし (Punto) → 通常の賭けへ。
	}
}

// PlayerBet 人間チームの賭けアクション。
func (g *Mus) PlayerBet(action, amount int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.roundIndex() < 0 {
		return ErrWrongPhase
	}
	if g.betTeam != g.humanTeam() {
		return ErrNotHumanTurn
	}
	return g.resolveBet(action, amount)
}

// resolveBet 賭けアクションを反映する。
func (g *Mus) resolveBet(action, amount int) error {
	ri := g.roundIndex()
	switch action {
	case MusActionPaso:
		if g.pendingStake > 0 {
			return NewDomainError(ErrInvalidPlay, "保留中の賭けにはパスできません")
		}
		g.appendLog(-1, "paso", fmt.Sprintf("Team %s passes", teamName(g.betTeam)), nil)
		if g.firstActorPaso {
			// 両チーム・パソ → 流局 (showdown で +1)。
			g.results[ri] = MusRoundResult{Kind: MusResultDeferred, Stake: 1, Team: -1}
			g.advanceRound()
			return nil
		}
		g.firstActorPaso = true
		g.betTeam = 1 - g.betTeam
		return nil
	case MusActionEnvido:
		if g.pendingStake < 0 {
			// オルダゴには Quiero / NoQuiero でしか応答できない。
			return NewDomainError(ErrInvalidPlay, "オルダゴにはエンビードできません")
		}
		if amount < 1 {
			amount = 2
		}
		if amount <= g.pendingStake {
			return NewDomainError(ErrInvalidPlay, "レイズ額が不足しています")
		}
		g.pendingStake = amount
		g.lastBettorTeam = g.betTeam
		g.appendLog(-1, "envido", fmt.Sprintf("Team %s bets %d", teamName(g.betTeam), amount), nil)
		g.betTeam = 1 - g.betTeam
		return nil
	case MusActionOrdago:
		g.pendingStake = -1 // sentinel: ordago
		g.lastBettorTeam = g.betTeam
		g.appendLog(-1, "ordago", fmt.Sprintf("Team %s declares Ordago!", teamName(g.betTeam)), nil)
		g.betTeam = 1 - g.betTeam
		return nil
	case MusActionQuiero:
		if g.pendingStake == 0 {
			return NewDomainError(ErrInvalidPlay, "受ける賭けがありません")
		}
		if g.pendingStake < 0 { // ordago accepted
			g.results[ri] = MusRoundResult{Kind: MusResultOrdago, Stake: 0, Team: -1}
			g.resolveOrdago(ri)
			return nil
		}
		g.results[ri] = MusRoundResult{Kind: MusResultAccepted, Stake: g.pendingStake, Team: -1}
		g.appendLog(-1, "quiero", fmt.Sprintf("Team %s accepts (%d)", teamName(g.betTeam), g.pendingStake), nil)
		g.advanceRound()
		return nil
	case MusActionNoQuiero:
		if g.pendingStake == 0 {
			return NewDomainError(ErrInvalidPlay, "降りる賭けがありません")
		}
		// 賭け手チームが +1 (ノキエロは 1 アマ)。
		win := g.lastBettorTeam
		g.results[ri] = MusRoundResult{Kind: MusResultAwarded, Stake: 1, Team: win}
		g.amarrakos[win]++
		g.appendLog(-1, "no_quiero", fmt.Sprintf("Team %s declines; Team %s +1", teamName(g.betTeam), teamName(win)), nil)
		g.checkGameEnd()
		if g.gameEndFlag {
			return nil
		}
		g.advanceRound()
		return nil
	default:
		return NewDomainError(ErrInvalidPlay, "不正なアクションです")
	}
}

// resolveOrdago オルダゴを即時解決する。ラウンドの勝者チームがゲーム勝利。
func (g *Mus) resolveOrdago(ri int) {
	win := g.roundWinner(ri)
	g.amarrakos[win] = g.config.TargetAmarrakos
	g.gameEndFlag = true
	g.winnerTeam = win
	g.phase = MusPhaseGameEnd
	g.appendLog(-1, "ordago_result", fmt.Sprintf("Ordago resolved: Team %s wins the game!", teamName(win)), nil)
}

// advanceRound 次の賭けラウンドへ進む。Juego の後は showdown。
func (g *Mus) advanceRound() {
	switch g.phase {
	case MusPhaseGrande:
		g.phase = MusPhaseChica
		g.startBetRound()
	case MusPhaseChica:
		g.phase = MusPhasePares
		g.startBetRound()
	case MusPhasePares:
		g.phase = MusPhaseJuego
		g.startBetRound()
	case MusPhaseJuego:
		g.phase = MusPhaseShowdown
	default:
	}
}

// Showdown 受理・流局ラウンドを精算する。
func (g *Mus) Showdown() {
	if g.phase != MusPhaseShowdown {
		return
	}
	for ri := 0; ri < MusRoundCnt; ri++ {
		r := &g.results[ri]
		if r.Kind != MusResultAccepted && r.Kind != MusResultDeferred {
			continue
		}
		win := g.roundWinner(ri)
		r.Team = win
		g.amarrakos[win] += r.Stake
		g.appendLog(-1, "showdown", fmt.Sprintf("%s: Team %s wins +%d", musRoundName(ri), teamName(win), r.Stake), nil)
		if g.checkGameEnd() {
			return
		}
	}
	g.phase = MusPhaseRoundEnd
}

// checkGameEnd 目標到達でゲーム終了を確定する。終了したら true。
func (g *Mus) checkGameEnd() bool {
	for t := 0; t < MusTeamCnt; t++ {
		if g.amarrakos[t] >= g.config.TargetAmarrakos {
			g.gameEndFlag = true
			g.winnerTeam = t
			g.phase = MusPhaseGameEnd
			g.appendLog(-1, "game_end", fmt.Sprintf("Team %s wins the game!", teamName(t)), nil)
			return true
		}
	}
	return false
}

// --- Hand evaluation ---

// roundWinner ラウンド ri の勝者チームを返す。同点は親 (mano) のチームが勝つ。
func (g *Mus) roundWinner(ri int) int {
	k0 := g.teamKey(0, ri)
	k1 := g.teamKey(1, ri)
	if k0 > k1 {
		return 0
	}
	if k1 > k0 {
		return 1
	}
	return MusTeamOf(g.manoIdx)
}

// teamKey チーム team のラウンド ri における最強の手キー (高いほど強い)。
func (g *Mus) teamKey(team, ri int) int {
	best := -1 << 30
	for i := 0; i < MusPlayerCnt; i++ {
		if MusTeamOf(i) != team {
			continue
		}
		if k := g.handKey(i, ri); k > best {
			best = k
		}
	}
	return best
}

// handKey プレイヤー i のラウンド ri における手キー。
func (g *Mus) handKey(i, ri int) int {
	cards := g.handRanks(i)
	switch ri {
	case 0: // Grande: 高い順
		return musEncodeDesc(cards)
	case 1: // Chica: 低い順 (小さいほど良い → 符号反転)
		return -musEncodeAsc(cards)
	case 2: // Pares
		return musParesKey(cards)
	case 3: // Juego / Punto
		return musJuegoKey(g.handPoints(i))
	default:
		return 0
	}
}

// handRanks プレイヤー i の 4 枚の mus ランク (A=1..7=7,J=8,Q=9,K=10) を返す。
func (g *Mus) handRanks(i int) []int {
	p := g.players[i]
	ranks := make([]int, 0, p.GetCardsSize())
	for j := 0; j < p.GetCardsSize(); j++ {
		ranks = append(ranks, musCardRank(p.GetCard(j).GetValue()))
	}
	return ranks
}

// handPoints プレイヤー i の Juego 用合計点 (A=1,2..7,J/Q/K=10)。
func (g *Mus) handPoints(i int) int {
	p := g.players[i]
	sum := 0
	for j := 0; j < p.GetCardsSize(); j++ {
		sum += musCardPoints(p.GetCard(j).GetValue())
	}
	return sum
}

// teamHasPares チーム team にペアを持つプレイヤーがいるか。
func (g *Mus) teamHasPares(team int) bool {
	for i := 0; i < MusPlayerCnt; i++ {
		if MusTeamOf(i) == team && musParesCategory(g.handRanks(i)) > 0 {
			return true
		}
	}
	return false
}

// teamHasJuego チーム team に Juego (31+) を持つプレイヤーがいるか。
func (g *Mus) teamHasJuego(team int) bool {
	for i := 0; i < MusPlayerCnt; i++ {
		if MusTeamOf(i) == team && g.handPoints(i) >= MusJuegoThreshold {
			return true
		}
	}
	return false
}

// --- Card helpers ---

// musCardRank A=1,2..7,J=8,Q=9,K=10 (高いほど Grande で強い)。
func musCardRank(value int) int {
	switch value {
	case 11: // Sota (J)
		return 8
	case 12: // Caballo (Q)
		return 9
	case 13: // Rey (K)
		return 10
	default: // A(1)..7
		return value
	}
}

// musCardPoints Juego 用点。A=1,2..7,J/Q/K=10。
func musCardPoints(value int) int {
	switch value {
	case 11, 12, 13:
		return 10
	default:
		return value
	}
}

// musEncodeDesc ランクを降順に並べた base-11 エンコード (大きいほど高い手)。
func musEncodeDesc(ranks []int) int {
	s := append([]int(nil), ranks...)
	sort.Sort(sort.Reverse(sort.IntSlice(s)))
	return musEncode(s)
}

// musEncodeAsc ランクを昇順に並べた base-11 エンコード (小さいほど低い手)。
func musEncodeAsc(ranks []int) int {
	s := append([]int(nil), ranks...)
	sort.Ints(s)
	return musEncode(s)
}

// musEncode ランク列を base-11 で 1 つの整数に符号化する。
func musEncode(s []int) int {
	v := 0
	for _, r := range s {
		v = v*11 + r
	}
	return v
}

// musParesCategory ペアの分類。0=なし, 1=par(1ペア), 2=medias(3カード), 3=duples(2ペア/4カード)。
func musParesCategory(ranks []int) int {
	counts := map[int]int{}
	for _, r := range ranks {
		counts[r]++
	}
	pairs, triples, quads := 0, 0, 0
	for _, c := range counts {
		switch c {
		case 2:
			pairs++
		case 3:
			triples++
		case 4:
			quads++
		}
	}
	switch {
	case quads > 0 || pairs >= 2:
		return 3 // duples
	case triples > 0:
		return 2 // medias
	case pairs == 1:
		return 1 // par
	default:
		return 0
	}
}

// musParesKey Pares の手キー (分類 + 内訳ランクで比較)。ペアなしは負値。
func musParesKey(ranks []int) int {
	cat := musParesCategory(ranks)
	if cat == 0 {
		return -1 << 20
	}
	counts := map[int]int{}
	for _, r := range ranks {
		counts[r]++
	}
	// 最も枚数の多いランク、同枚数ならランクの高い方をタイブレークに使う。
	bestRank, bestCnt := 0, 0
	for r, c := range counts {
		if c > bestCnt || (c == bestCnt && r > bestRank) {
			bestCnt, bestRank = c, r
		}
	}
	return cat*100 + bestRank
}

// musJuegoKey Juego/Punto の手キー。31 が最強、次いで 32、その後 40..33。
// 31 未満は Punto: 30 に近いほど強い (= 点が高いほど強い)。
func musJuegoKey(points int) int {
	if points >= MusJuegoThreshold {
		switch points {
		case 31:
			return 1000
		case 32:
			return 999
		default: // 33..40 → 40 が 990, 33 が 983 の順 (40>39>..>33)
			return 950 + points // 高い点ほど強い
		}
	}
	return points // Punto: 30 が最大 30
}

// --- Misc helpers ---

// humanTeam 人間プレイヤーのチームを返す (人間不在は -1)。
func (g *Mus) humanTeam() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return MusTeamOf(i)
		}
	}
	return -1
}

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Mus) sortAllHands() {
	for _, p := range g.players {
		musSortHand(p)
	}
}

// musSortHand 手札をランク降順にソートする。
func musSortHand(p *MusPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		return musCardRank(cards[i].GetValue()) > musCardRank(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// musRoundName ラウンド名。
func musRoundName(ri int) string {
	switch ri {
	case 0:
		return "Grande"
	case 1:
		return "Chica"
	case 2:
		return "Pares"
	case 3:
		return "Juego"
	default:
		return "?"
	}
}

// --- CPU ---

// CpuPlay 現在の手番が CPU の場合に 1 アクション実行する。
func (g *Mus) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	switch g.phase {
	case MusPhaseMus:
		if g.players[g.musTurn].GetIsHuman() {
			return
		}
		g.resolveMus(g.cpuWantsMus(g.musTurn))
	case MusPhaseDiscard:
		if g.players[g.discardTurn].GetIsHuman() {
			return
		}
		g.applyDiscard(g.cpuDiscardIndices(g.discardTurn))
	case MusPhaseGrande, MusPhaseChica, MusPhasePares, MusPhaseJuego:
		// Act for any team that is not the human's (no human → CPU drives both).
		if g.humanTeam() >= 0 && g.betTeam == g.humanTeam() {
			return
		}
		action, amount := g.cpuBet()
		_ = g.resolveBet(action, amount)
	default:
	}
}

// cpuWantsMus CPU が交換を希望するか (弱い手なら希望)。
func (g *Mus) cpuWantsMus(i int) bool {
	if g.musCycle >= MusMaxMusCycles {
		return false
	}
	// 高い手 (Grande キー) が低く、ペアも Juego もなければ交換したい。
	hasPares := musParesCategory(g.handRanks(i)) > 0
	hasJuego := g.handPoints(i) >= MusJuegoThreshold
	if hasPares || hasJuego {
		return false
	}
	if g.config.CpuDifficulty == MusCpuDifficultyEasy {
		return rand.Intn(2) == 0
	}
	return true
}

// cpuDiscardIndices CPU が交換する札 (ペアを残し、それ以外を捨てる簡易戦略)。
func (g *Mus) cpuDiscardIndices(i int) []int {
	p := g.players[i]
	ranks := g.handRanks(i)
	counts := map[int]int{}
	for _, r := range ranks {
		counts[r]++
	}
	var out []int
	for j := 0; j < p.GetCardsSize(); j++ {
		// ペアを構成しない札、かつ figure(高位)でない札を交換する。
		r := musCardRank(p.GetCard(j).GetValue())
		if counts[r] < 2 && r < 8 {
			out = append(out, j)
		}
	}
	return out
}

// cpuBet CPU チームの賭けアクションを決める。
func (g *Mus) cpuBet() (int, int) {
	ri := g.roundIndex()
	strength := g.teamStrength(g.betTeam, ri)

	if g.pendingStake != 0 {
		// 応答: 強ければ受ける/レイズ、弱ければ降りる。
		if g.pendingStake < 0 { // ordago
			if strength >= 80 {
				return MusActionQuiero, 0
			}
			return MusActionNoQuiero, 0
		}
		// Cap raises so two strong CPU teams cannot raise-war forever.
		if strength >= 75 && g.pendingStake < 10 && g.config.CpuDifficulty != MusCpuDifficultyEasy {
			return MusActionEnvido, g.pendingStake + 2
		}
		if strength >= 50 {
			return MusActionQuiero, 0
		}
		return MusActionNoQuiero, 0
	}
	// 主導: 強ければ賭ける。
	if strength >= 70 {
		return MusActionEnvido, 2
	}
	return MusActionPaso, 0
}

// teamStrength チーム team のラウンド ri における手の強さを 0-100 で粗く評価する。
func (g *Mus) teamStrength(team, ri int) int {
	myKey := g.teamKey(team, ri)
	oppKey := g.teamKey(1-team, ri)
	if myKey > oppKey {
		return 80
	}
	if myKey == oppKey {
		return 50
	}
	return 25
}

// --- Hint ---

// GetHint 人間の手番における推奨アクションを返す。
func (g *Mus) GetHint() *MusHint {
	human := findHumanIdx(g.players)
	if human < 0 {
		return nil
	}
	switch g.phase {
	case MusPhaseMus:
		if g.musTurn != human {
			return nil
		}
		want := g.cpuWantsMus(human)
		reason := "mus_cut"
		if want {
			reason = "mus_exchange"
		}
		return &MusHint{Mus: want, Reason: reason}
	case MusPhaseDiscard:
		if g.discardTurn != human {
			return nil
		}
		return &MusHint{Indices: g.cpuDiscardIndices(human), Reason: "discard_low"}
	case MusPhaseGrande, MusPhaseChica, MusPhasePares, MusPhaseJuego:
		if g.betTeam != g.humanTeam() {
			return nil
		}
		action, amount := g.cpuBet()
		return &MusHint{Action: action, Amount: amount, Reason: "bet_" + musActionName(action)}
	default:
		return nil
	}
}

// musActionName アクション名キー。
func musActionName(action int) string {
	switch action {
	case MusActionPaso:
		return "paso"
	case MusActionEnvido:
		return "envido"
	case MusActionOrdago:
		return "ordago"
	case MusActionQuiero:
		return "quiero"
	case MusActionNoQuiero:
		return "no_quiero"
	default:
		return "?"
	}
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Mus) GetPhase() MusPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Mus) SetPhase(phase MusPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Mus) GetRoundNumber() int { return g.roundNumber }

// GetManoIdx 親インデックス取得
func (g *Mus) GetManoIdx() int { return g.manoIdx }

// SetManoIdx 親設定 (テスト用)
func (g *Mus) SetManoIdx(idx int) { g.manoIdx = idx }

// GetMusTurn Mus 宣言中のプレイヤー取得
func (g *Mus) GetMusTurn() int { return g.musTurn }

// SetMusTurn Mus 宣言プレイヤー設定 (テスト用)
func (g *Mus) SetMusTurn(idx int) { g.musTurn = idx }

// GetDiscardTurn 交換中のプレイヤー取得
func (g *Mus) GetDiscardTurn() int { return g.discardTurn }

// SetDiscardTurn 交換中プレイヤー設定 (テスト用)
func (g *Mus) SetDiscardTurn(idx int) { g.discardTurn = idx }

// GetBetTeam 賭けアクションするチーム取得
func (g *Mus) GetBetTeam() int { return g.betTeam }

// SetBetTeam 賭けチーム設定 (テスト用)
func (g *Mus) SetBetTeam(t int) { g.betTeam = t }

// GetPendingStake 応答待ちの賭け額取得 (-1=ordago, 0=なし)
func (g *Mus) GetPendingStake() int { return g.pendingStake }

// SetPendingStake 賭け額設定 (テスト用)
func (g *Mus) SetPendingStake(s int) { g.pendingStake = s }

// GetLastBettorTeam 直近の賭けチーム取得
func (g *Mus) GetLastBettorTeam() int { return g.lastBettorTeam }

// SetLastBettorTeam 直近賭けチーム設定 (テスト用)
func (g *Mus) SetLastBettorTeam(t int) { g.lastBettorTeam = t }

// GetAmarrakos チーム別累積点取得
func (g *Mus) GetAmarrakos() [MusTeamCnt]int { return g.amarrakos }

// SetAmarrakos チーム別累積点設定 (テスト用)
func (g *Mus) SetAmarrakos(a [MusTeamCnt]int) { g.amarrakos = a }

// GetResult ラウンド ri の結果取得
func (g *Mus) GetResult(ri int) MusRoundResult {
	if ri < 0 || ri >= MusRoundCnt {
		return MusRoundResult{Team: -1}
	}
	return g.results[ri]
}

// GetMusCycle Mus 交換の繰り返し回数取得
func (g *Mus) GetMusCycle() int { return g.musCycle }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Mus) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1=未確定)
func (g *Mus) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *Mus) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Mus) GetPlayer(i int) *MusPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番が人間か。
func (g *Mus) IsHumanTurn() bool {
	switch g.phase {
	case MusPhaseMus:
		return g.players[g.musTurn].GetIsHuman()
	case MusPhaseDiscard:
		return g.players[g.discardTurn].GetIsHuman()
	case MusPhaseGrande, MusPhaseChica, MusPhasePares, MusPhaseJuego:
		return g.humanTeam() >= 0 && g.betTeam == g.humanTeam()
	default:
		return false
	}
}

// GetConfig 設定取得
func (g *Mus) GetConfig() MusConfig { return g.config }

// SetConfig 設定変更
func (g *Mus) SetConfig(cfg MusConfig) { g.config = cfg }

// --- JSON ---

// musJSON is the JSON wire format for Mus.
type musJSON struct {
	TrumpCards     *TrumpCards                 `json:"tc"`
	Players        []*MusPlayer                `json:"ps"`
	Config         MusConfig                   `json:"cf"`
	Phase          MusPhase                    `json:"ph"`
	RoundNumber    int                         `json:"rn"`
	ManoIdx        int                         `json:"mi"`
	Amarrakos      [MusTeamCnt]int             `json:"am"`
	Results        [MusRoundCnt]MusRoundResult `json:"rs"`
	MusTurn        int                         `json:"mt"`
	MusAgreed      int                         `json:"ma"`
	MusCycle       int                         `json:"mc"`
	DiscardTurn    int                         `json:"dt"`
	BetTeam        int                         `json:"bt"`
	PendingStake   int                         `json:"pk"`
	LastBettorTeam int                         `json:"lb"`
	FirstActorPaso bool                        `json:"fp"`
	Discarded      []*Card                     `json:"ds"`
	GameEndFlag    bool                        `json:"ge"`
	WinnerTeam     int                         `json:"wt"`
	ActionLog      []*ActionLogEntry           `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Mus) MarshalJSON() ([]byte, error) {
	return json.Marshal(musJSON{
		TrumpCards:     g.trumpCards,
		Players:        g.players,
		Config:         g.config,
		Phase:          g.phase,
		RoundNumber:    g.roundNumber,
		ManoIdx:        g.manoIdx,
		Amarrakos:      g.amarrakos,
		Results:        g.results,
		MusTurn:        g.musTurn,
		MusAgreed:      g.musAgreed,
		MusCycle:       g.musCycle,
		DiscardTurn:    g.discardTurn,
		BetTeam:        g.betTeam,
		PendingStake:   g.pendingStake,
		LastBettorTeam: g.lastBettorTeam,
		FirstActorPaso: g.firstActorPaso,
		Discarded:      g.discarded,
		GameEndFlag:    g.gameEndFlag,
		WinnerTeam:     g.winnerTeam,
		ActionLog:      g.actionLog,
	})
}

// musMaxSliceLen caps slice sizes during deserialisation. Set well above the
// largest realistic action-log length so a long game still loads.
const musMaxSliceLen = 5000

// errMusOversized is the single sentinel error for oversized input arrays.
var errMusOversized = errors.New("mus: input array exceeds maximum allowed size")

// errMusInvalidPlayers is returned when restored state lacks exactly MusPlayerCnt players.
var errMusInvalidPlayers = errors.New("mus: invalid player count")

// UnmarshalJSON implements json.Unmarshaler.
func (g *Mus) UnmarshalJSON(data []byte) error {
	var j musJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > musMaxSliceLen || len(j.ActionLog) > musMaxSliceLen {
		return errMusOversized
	}
	// A real Mus game always has exactly MusPlayerCnt players; reject otherwise
	// so the fixed-index player access below cannot panic on malformed input.
	if len(j.Players) != MusPlayerCnt {
		return errMusInvalidPlayers
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsBriscola()
	}
	g.players = j.Players
	if g.players == nil {
		g.players = make([]*MusPlayer, 0)
	}
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.manoIdx = j.ManoIdx
	g.amarrakos = j.Amarrakos
	g.results = j.Results
	g.musTurn = j.MusTurn
	g.musAgreed = j.MusAgreed
	g.musCycle = j.MusCycle
	g.discardTurn = j.DiscardTurn
	g.betTeam = j.BetTeam
	g.pendingStake = j.PendingStake
	g.lastBettorTeam = j.LastBettorTeam
	g.firstActorPaso = j.FirstActorPaso
	g.discarded = j.Discarded
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
