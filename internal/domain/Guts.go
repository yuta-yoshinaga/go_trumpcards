//go:build !js || !wasm || extra

// Package domain ガッツ (Guts) のドメインモデル。
//
// Guts はアメリカ発祥のシンプルなポーカー系「ヴァイイング (vying)」ポットゲーム。
// 標準 52 枚デッキを使い、2〜7 人が全員アンティを入れてから 2 枚ずつ配られる。
//
// # 1 ラウンドの流れ
//
//  1. 全員がアンティをポットに払い、2 枚ずつ配られる。
//  2. 全員が同時に「イン (勝負に残る)」か「アウト (降りる)」かを宣言する
//     (人間が宣言し、CPU は手札の強さで自動判断する)。
//  3. 「イン」のプレイヤー同士が 2 枚の手を比べる: ペアはノーペアに勝ち、ペア同士は
//     ランクの高い方、ノーペア同士はハイカード→キッカーの順で比較する (A ハイ)。
//  4. 最強手がポットを総取りする。
//  5. 「イン」して負けたプレイヤーは、その時点のポット額を「マッチ」して支払い、
//     それが次ラウンドのポットの種銭になる (エスカレーション / ペナルティ)。
//
// # エスカレーションと停止条件
//
// マッチによりポットが膨らむため、勝負が続くほどリスクが高まる。本実装では以下を
// 停止条件とする (決定的・有界):
//
//   - 誰も「イン」しなかったラウンドは、ポットが丸ごと次ラウンドへ持ち越される。
//   - ちょうど 1 人だけ「イン」したラウンドはクリーンウィン (マッチ発生なし)。ポットは
//     勝者が総取りし、次ラウンドは新規アンティから始まる。
//   - 2 人以上が「イン」したラウンドは、最強手が総取りし、他の「イン」プレイヤーは
//     ポット額をマッチして次ラウンドの種銭に積む。
//
// ゲームは「規定ラウンド数 (TargetRounds) を消化」または「アンティを払えるプレイヤーが
// 2 人未満 (人間の脱落を含む)」で終了し、最終的にチップ最多のプレイヤーが勝者となる。
//
// # デッキ
//
// 標準 52 枚 (ジョーカーなし)。NewTrumpCards(0) は extra ワーカーから到達可能。
//
// 本実装は extra ワーカーから到達可能なよう、手役評価・ポット精算ロジックをすべて
// インラインで持つ (RedDog / ThreeCardBrag は casino ビルドタグで到達不可のため)。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// GutsPhase はゲームフェーズ。
type GutsPhase int

// Guts のフェーズ定数。ワイヤー値はフロントエンドの enum と一致させる。
const (
	// GutsPhaseDeclare 宣言受付中 (配札済み; 人間の in/out 待ち)。ワイヤー値 0。
	GutsPhaseDeclare GutsPhase = 0
	// GutsPhaseResult ラウンド解決済み (結果表示; 次ラウンド待ち or ゲーム終了)。ワイヤー値 1。
	GutsPhaseResult GutsPhase = 1
)

// 手役カテゴリ (高いほど強い)。
const (
	// GutsHandHighCard ハイカード (ノーペア)
	GutsHandHighCard = 0
	// GutsHandPair ペア
	GutsHandPair = 1
)

// GutsResult は人間プレイヤーから見たラウンド結果。
// GameResult は共有ファイル internal/domain/game_result.go に移動したので到達可能に
// なったが、この型名は JSON ペイロードに出るため統合していない（#4462）。値は
// GameResult と同一。
type GutsResult int

const (
	// GutsResultLose 負け (イン宣言したが敗北; マッチ義務)
	GutsResultLose GutsResult = -1
	// GutsResultNone 結果なし (アウト宣言 / 未参加 / 未解決)
	GutsResultNone GutsResult = 0
	// GutsResultWin 勝ち (ポット獲得)
	GutsResultWin GutsResult = 1
)

// gutsMaxSliceLen はデシリアライズ時のスライス長の上限。
const gutsMaxSliceLen = 1000

// gutsCpuStayHighCard は CPU がノーペアでも「イン」を宣言する最低ハイカードランク (J=11 以上)。
const gutsCpuStayHighCard = 11

// デシリアライズ検証用のセンチネルエラー。
var (
	errGutsSnapshot      = errors.New("guts: invalid serialised game state")
	errGutsInvalidPlayer = errors.New("guts: invalid player state")
)

// GutsHint はヒント情報 (人間への in/out 助言)。
type GutsHint struct {
	Declaration GutsDeclaration // 推奨宣言 (In / Out)
	Reason      string          // ヒント理由キー
}

// gutsState はゲーム進行状態。
type gutsState struct {
	phase          GutsPhase
	roundNumber    int
	pot            int   // 現在のラウンドのポット
	carryPot       int   // 次ラウンドへ持ち越す種銭 (マッチ額 or 全員アウト時の持ち越し)
	carryCount     int   // 全員アウトでポットが連続繰り越しになった回数
	winnerIdx      int   // 直近ラウンドの勝者 (-1 = なし)
	matchWinnerIdx int   // ゲーム全体の勝者 (-1 = 未確定)
	matchers       []int // このラウンドでマッチしたプレイヤー (負けたイン宣言者)
	result         GutsResult
	gameEndFlag    bool
	scored         bool // ラウンド結果を確定済みか (二重確定防止)
	actionLogBase
}

// Guts はガッツの状態を保持する集約ルート。
type Guts struct {
	trumpCards *TrumpCards
	players    []*GutsPlayer
	config     GutsConfig
	state      gutsState
}

// NewGuts はコンストラクタ。
func NewGuts(trumpCards *TrumpCards, players []*GutsPlayer, config GutsConfig) *Guts {
	return &Guts{
		trumpCards: trumpCards,
		players:    players,
		config:     config,
		state: gutsState{
			phase:          GutsPhaseDeclare,
			winnerIdx:      -1,
			matchWinnerIdx: -1,
			matchers:       make([]int, 0),
			actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
		},
	}
}

// NewDefaultGuts は標準構成 (人間 seat 0 + CPU) を生成する。CUI / Web / Worker 構築の単一情報源。
func NewDefaultGuts() *Guts {
	cfg := DefaultGutsConfig()
	g := NewGuts(NewTrumpCards(0), gutsNewPlayers(cfg), cfg)
	g.Reset()
	return g
}

// gutsNewPlayers は設定に基づいてプレイヤー列を生成する (seat 0 = 人間)。
func gutsNewPlayers(cfg GutsConfig) []*GutsPlayer {
	players := make([]*GutsPlayer, cfg.PlayerCount)
	for i := range players {
		players[i] = NewGutsPlayer(i == 0, cfg.StartingChips)
	}
	return players
}

// --- ゲーム進行 ---

// Reset は新しいゲームを開始する。チップ・プレイヤー数を設定から作り直し、第 1 ラウンドを配る。
func (g *Guts) Reset() {
	g.players = gutsNewPlayers(g.config)
	g.trumpCards = NewTrumpCards(0)
	g.trumpCards.Shuffle()
	g.state = gutsState{
		phase:          GutsPhaseDeclare,
		roundNumber:    1,
		winnerIdx:      -1,
		matchWinnerIdx: -1,
		matchers:       make([]int, 0),
		actionLogBase:  actionLogBase{actionLog: make([]*ActionLogEntry, 0)},
	}
	g.startRound()
}

// NextRound は同じチップを保持したまま次のラウンドを配る。Result フェーズかつゲーム
// 継続中のときのみ有効。
func (g *Guts) NextRound() {
	if g.state.phase != GutsPhaseResult || g.state.gameEndFlag {
		return
	}
	g.state.roundNumber++
	g.startRound()
}

// startRound は 1 ラウンドを準備する: 脱落判定・アンティ徴収・配札・宣言フェーズへ遷移。
func (g *Guts) startRound() {
	// ラウンド単位の状態をクリア。
	g.state.winnerIdx = -1
	g.state.result = GutsResultNone
	g.state.matchers = g.state.matchers[:0]
	g.state.scored = false
	for _, p := range g.players {
		p.ResetForRound()
	}
	// アンティを払えないプレイヤーは脱落。
	for _, p := range g.players {
		if !p.GetOut() && p.GetChips() < g.config.Ante {
			p.SetOut(true)
		}
	}
	// 参加可能なプレイヤーが 2 人未満、または人間が脱落 → ゲーム終了。
	if g.solventCount() < GutsMinPlayerCount || g.players[0].GetOut() {
		g.endGame()
		return
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	// 持ち越しを種銭にし、アンティ徴収 + 配札。
	g.state.pot = g.state.carryPot
	g.state.carryPot = 0
	for _, p := range g.players {
		if p.GetOut() {
			continue
		}
		p.SubtractChips(g.config.Ante)
		p.AddRoundBet(g.config.Ante)
		g.state.pot += g.config.Ante
		for i := 0; i < GutsHandSize; i++ {
			if c := g.trumpCards.DrawCard(); c != nil {
				p.AddCard(c)
			}
		}
	}
	g.state.phase = GutsPhaseDeclare
	g.appendLog(-1, "deal",
		fmt.Sprintf("Round %d: ante %d, pot %d", g.state.roundNumber, g.config.Ante, g.state.pot), nil)
}

// Declare は人間 (seat 0) の in/out 宣言を受け付け、ラウンドを解決する。
func (g *Guts) Declare(stay bool) error {
	if g.state.gameEndFlag {
		return NewDomainError(ErrGameEnded, "the game has already ended")
	}
	if g.state.phase != GutsPhaseDeclare {
		return NewDomainError(ErrWrongPhase, "declaration is only allowed during the declare phase")
	}
	if g.players[0].GetOut() {
		return NewDomainError(ErrInvalidPlay, "you are out of the game")
	}
	g.players[0].SetIn(stay)
	g.appendLog(0, "declare", gutsDeclareText(0, stay), nil)
	g.resolve()
	return nil
}

// resolve は CPU の宣言を確定させ、勝敗とポット精算を行う。
func (g *Guts) resolve() {
	g.cpuDeclare()
	g.settle()
}

// cpuDeclare は各 CPU の in/out を手札の強さで決める。
func (g *Guts) cpuDeclare() {
	for i, p := range g.players {
		if p.GetOut() || p.GetIsHuman() {
			continue
		}
		stay := g.cpuStays(p)
		p.SetIn(stay)
		g.appendLog(i, "declare", gutsDeclareText(i, stay), nil)
	}
}

// cpuStays は CPU がインすべきか (ペア、またはハイカードが J 以上) を返す。
func (g *Guts) cpuStays(p *GutsPlayer) bool {
	cat, tb := gutsEvalPlayer(p)
	if cat == GutsHandPair {
		return true
	}
	return len(tb) > 0 && tb[0] >= gutsCpuStayHighCard
}

// settle は「イン」プレイヤーの手を比べ、ポットを精算し、フェーズを Result へ遷移する。
// 呼び出し前に各プレイヤーの in フラグが確定していること (Declare / テスト経由)。
func (g *Guts) settle() {
	inPlayers := g.inPlayers()
	g.state.winnerIdx = -1
	g.state.result = GutsResultNone
	g.state.matchers = g.state.matchers[:0]

	switch len(inPlayers) {
	case 0:
		// 誰も残らなかった: ポットを丸ごと次ラウンドへ持ち越す。
		g.state.carryPot = g.state.pot
		g.state.carryCount++
		g.appendLog(-1, "result", fmt.Sprintf("nobody stayed; pot %d carries over", g.state.pot), nil)
	default:
		winner := g.bestHand(inPlayers)
		g.state.winnerIdx = winner
		g.state.carryCount = 0
		g.players[winner].AddChips(g.state.pot)
		g.appendLog(winner, "win",
			fmt.Sprintf("%s wins the pot (%d)", playerName(g.players, winner), g.state.pot), nil)
		// 勝者以外の「イン」プレイヤーはポット額をマッチして次ラウンドの種銭に積む。
		for _, idx := range inPlayers {
			if idx == winner {
				continue
			}
			pay := min(g.state.pot, g.players[idx].GetChips())
			g.players[idx].SubtractChips(pay)
			g.players[idx].AddRoundBet(pay)
			g.state.carryPot += pay
			g.state.matchers = append(g.state.matchers, idx)
			g.appendLog(idx, "match",
				fmt.Sprintf("%s matches the pot (pays %d)", playerName(g.players, idx), pay), nil)
		}
		g.setHumanResult(winner)
	}

	g.state.pot = 0
	g.state.scored = true
	g.state.phase = GutsPhaseResult
	g.checkGameEnd()
}

// setHumanResult は人間 (seat 0) の勝敗結果を設定する。
func (g *Guts) setHumanResult(winnerIdx int) {
	human := g.players[0]
	if human.GetOut() || !human.GetIn() {
		g.state.result = GutsResultNone
		return
	}
	if winnerIdx == 0 {
		g.state.result = GutsResultWin
	} else {
		g.state.result = GutsResultLose
	}
}

// checkGameEnd は停止条件 (規定ラウンド到達 or 参加可能者 2 人未満) を判定し、
// 満たせばゲームを終了させる。
func (g *Guts) checkGameEnd() {
	if g.state.roundNumber >= g.config.TargetRounds || g.solventCount() < GutsMinPlayerCount {
		g.endGame()
	}
}

// endGame はゲームを終了し、チップ最多のプレイヤーを勝者に設定する。
func (g *Guts) endGame() {
	g.state.gameEndFlag = true
	g.state.phase = GutsPhaseResult
	g.state.matchWinnerIdx = g.richestIdx()
	g.appendLog(g.state.matchWinnerIdx, "game_end",
		fmt.Sprintf("%s wins the game", playerName(g.players, g.state.matchWinnerIdx)), nil)
}

// --- 手役評価 (インライン) ---

// gutsRank はカードのランクを返す (A=14, K=13 ... 2=2)。
func gutsRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// GutsEval は 2 枚の手の役カテゴリと比較用タイブレーク列 (降順) を返す。
// カテゴリが同じならこの列を辞書順比較すれば勝敗が決まる。
func GutsEval(cards []*Card) (int, []int) {
	if len(cards) != GutsHandSize {
		return -1, nil
	}
	r := []int{gutsRank(cards[0]), gutsRank(cards[1])}
	if r[0] < r[1] {
		r[0], r[1] = r[1], r[0]
	}
	if r[0] == r[1] {
		return GutsHandPair, []int{r[0]}
	}
	return GutsHandHighCard, r
}

// gutsEvalPlayer は playerIdx の手役を返す。
func gutsEvalPlayer(p *GutsPlayer) (int, []int) {
	cards := make([]*Card, 0, GutsHandSize)
	for i := 0; i < p.GetCardsSize(); i++ {
		cards = append(cards, p.GetCard(i))
	}
	return GutsEval(cards)
}

// Guts の宣言ガイドの強さ区分。Web (frontend/src/utils/gutsGuideUtils.ts) と
// 同じ値を返す必要があるため、両者は
// frontend/src/utils/__fixtures__/gutsGuide.golden.json を共有して検証している。
const (
	// GutsGuideTierHigh ペアあり (どのペアでも強い)。
	GutsGuideTierHigh = "high"
	// GutsGuideTierMedium ノーペアだが最高札が K か A。
	GutsGuideTierMedium = "medium"
	// GutsGuideTierLow それ以外のノーペア。
	GutsGuideTierLow = "low"
)

// gutsGuideMediumRank はノーペアが "medium" になる最低ランク (K=13)。
// CPU の in 判断 (gutsCpuStayHighCard = 11) とは別の基準であることに注意。
const gutsGuideMediumRank = 13

// GutsGuide は宣言前の手役診断 (役の有無と勝ち目の目安)。
type GutsGuide struct {
	// Pair は手がペアかどうか。false はハイカード。
	Pair bool
	// Tier は勝ち目の目安 (GutsGuideTier*)。
	Tier string
}

// GutsEvaluateGuide は手札から宣言ガイドを返す。手札が揃っていなければ nil。
// 判定は Web の evaluateGutsGuide と一致していなければならない。
func GutsEvaluateGuide(cards []*Card) *GutsGuide {
	cat, tb := GutsEval(cards)
	if cat < 0 {
		return nil
	}
	if cat == GutsHandPair {
		return &GutsGuide{Pair: true, Tier: GutsGuideTierHigh}
	}
	tier := GutsGuideTierLow
	if len(tb) > 0 && tb[0] >= gutsGuideMediumRank {
		tier = GutsGuideTierMedium
	}
	return &GutsGuide{Pair: false, Tier: tier}
}

// GutsCompare は手 a が手 b に勝てば 1、負ければ -1、引き分けは 0 を返す。
func GutsCompare(catA int, tbA []int, catB int, tbB []int) int {
	if catA != catB {
		if catA > catB {
			return 1
		}
		return -1
	}
	for i := 0; i < len(tbA) && i < len(tbB); i++ {
		if tbA[i] != tbB[i] {
			if tbA[i] > tbB[i] {
				return 1
			}
			return -1
		}
	}
	return 0
}

// bestHand は「イン」プレイヤーのうち最強手のインデックスを返す。引き分けは
// 最も座席番号の小さいプレイヤーが総取りする (決定的)。
func (g *Guts) bestHand(inPlayers []int) int {
	best := inPlayers[0]
	bestCat, bestTb := gutsEvalPlayer(g.players[best])
	for _, idx := range inPlayers[1:] {
		cat, tb := gutsEvalPlayer(g.players[idx])
		if GutsCompare(cat, tb, bestCat, bestTb) > 0 {
			best, bestCat, bestTb = idx, cat, tb
		}
	}
	return best
}

// --- ヘルパー ---

// inPlayers は「イン」を宣言した非脱落プレイヤーのインデックス列を返す。
func (g *Guts) inPlayers() []int {
	out := make([]int, 0, len(g.players))
	for i, p := range g.players {
		if !p.GetOut() && p.GetIn() {
			out = append(out, i)
		}
	}
	return out
}

// solventCount はアンティを払える (非脱落かつチップ >= アンティ) プレイヤー数を返す。
func (g *Guts) solventCount() int {
	return countPlayers(g.players, func(p *GutsPlayer) bool { return !p.GetOut() && p.GetChips() >= g.config.Ante })
}

// richestIdx はチップが最多のプレイヤーのインデックスを返す (同数は座席番号の小さい方)。
func (g *Guts) richestIdx() int {
	return maxIndexBy(g.players, func(p *GutsPlayer) int { return p.GetChips() })
}

// gutsDeclareText は宣言の棋譜テキストを返す。
func gutsDeclareText(idx int, stay bool) string {
	verb := "OUT"
	if stay {
		verb = "IN"
	}
	return fmt.Sprintf("player %d declares %s", idx, verb)
}

func (g *Guts) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.state.appendLog(playerIdx, actionType, detail, cards)
}

// --- Hint ---

// GetHint は宣言フェーズで人間に in/out の助言を返す。ペア or ハイカード J 以上なら
// イン、それ以外はアウトを推奨する。
func (g *Guts) GetHint() *GutsHint {
	if g.state.phase != GutsPhaseDeclare || g.state.gameEndFlag {
		return nil
	}
	human := g.players[0]
	if human.GetOut() || human.GetCardsSize() != GutsHandSize {
		return nil
	}
	if g.cpuStays(human) {
		return &GutsHint{Declaration: GutsDeclarationIn, Reason: "strong_hand"}
	}
	return &GutsHint{Declaration: GutsDeclarationOut, Reason: "weak_hand"}
}

// --- 状態アクセサ ---

// GetPhase は現在のフェーズを返す。
func (g *Guts) GetPhase() GutsPhase { return g.state.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (g *Guts) SetPhase(p GutsPhase) { g.state.phase = p }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Guts) GetGameEndFlag() bool { return g.state.gameEndFlag }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *Guts) GetRoundNumber() int { return g.state.roundNumber }

// GetPot は現在のポットを返す。
func (g *Guts) GetPot() int { return g.state.pot }

// SetPot はポットを設定する (テスト用)。
func (g *Guts) SetPot(v int) { g.state.pot = v }

// GetCarryPot は次ラウンドへの持ち越し種銭を返す。
func (g *Guts) GetCarryPot() int { return g.state.carryPot }

// GetCarryCount 全員アウトでポットが連続して繰り越された回数を返す。
func (g *Guts) GetCarryCount() int { return g.state.carryCount }

// GetAnte はアンティ額を返す。
func (g *Guts) GetAnte() int { return g.config.Ante }

// GetWinnerIdx は直近ラウンドの勝者を返す (-1 = なし)。
func (g *Guts) GetWinnerIdx() int { return g.state.winnerIdx }

// GetMatchWinnerIdx はゲーム全体の勝者を返す (-1 = 未確定)。
func (g *Guts) GetMatchWinnerIdx() int { return g.state.matchWinnerIdx }

// GetResult は人間から見たラウンド結果を返す。
func (g *Guts) GetResult() GutsResult { return g.state.result }

// GetMatchers はこのラウンドでマッチしたプレイヤーのインデックス列を返す。
func (g *Guts) GetMatchers() []int { return g.state.matchers }

// IsMatcher は idx がこのラウンドでマッチしたか (負けたイン宣言者) を返す。
func (g *Guts) IsMatcher(idx int) bool {
	for _, m := range g.state.matchers {
		if m == idx {
			return true
		}
	}
	return false
}

// GetPlayerCnt はプレイヤー数を返す。
func (g *Guts) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *Guts) GetPlayer(i int) *GutsPlayer {
	return getPlayer(g.players, i)
}

// GetChips は人間 (seat 0) の保有チップを返す。
func (g *Guts) GetChips() int {
	return chipsOfFirst(g.players)
}

// GetConfig はローカルルール設定を返す。
func (g *Guts) GetConfig() GutsConfig { return g.config }

// SetConfig はローカルルール設定を変更する。
func (g *Guts) SetConfig(cfg GutsConfig) { g.config = cfg }

// GetActionLog は棋譜を返す。
func (g *Guts) GetActionLog() []*ActionLogEntry { return g.state.actionLog }

// SettleForTest は in フラグを設定済みの状態でラウンドを解決する (テスト用)。乱数配札を
// 迂回して勝敗解決・マッチ精算を決定的に検証するためのショートカット。
func (g *Guts) SettleForTest() { g.settle() }

// --- JSON Serialization ---

// gutsJSON is the JSON wire format for Guts.
type gutsJSON struct {
	TrumpCards     *TrumpCards       `json:"tc"`
	Players        []*GutsPlayer     `json:"ps"`
	Config         GutsConfig        `json:"cf"`
	Phase          GutsPhase         `json:"ph"`
	RoundNumber    int               `json:"rn"`
	Pot            int               `json:"pt"`
	CarryPot       int               `json:"cp"`
	WinnerIdx      int               `json:"wi"`
	MatchWinnerIdx int               `json:"mw"`
	Matchers       []int             `json:"ma"`
	Result         GutsResult        `json:"re"`
	GameEndFlag    bool              `json:"ge"`
	Scored         bool              `json:"sc"`
	ActionLog      []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Guts) MarshalJSON() ([]byte, error) {
	return json.Marshal(gutsJSON{
		TrumpCards:     g.trumpCards,
		Players:        g.players,
		Config:         g.config,
		Phase:          g.state.phase,
		RoundNumber:    g.state.roundNumber,
		Pot:            g.state.pot,
		CarryPot:       g.state.carryPot,
		WinnerIdx:      g.state.winnerIdx,
		MatchWinnerIdx: g.state.matchWinnerIdx,
		Matchers:       g.state.matchers,
		Result:         g.state.result,
		GameEndFlag:    g.state.gameEndFlag,
		Scored:         g.state.scored,
		ActionLog:      g.state.actionLog,
	})
}

// gutsValidPhase は有効なフェーズかどうか。
func gutsValidPhase(p GutsPhase) bool {
	return p == GutsPhaseDeclare || p == GutsPhaseResult
}

// UnmarshalJSON implements json.Unmarshaler. 不正な永続化データを拒否するための
// バリデーションを行う。
func (g *Guts) UnmarshalJSON(data []byte) error {
	var j gutsJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("guts: invalid config: %w", err)
	}
	n := len(j.Players)
	if n < GutsMinPlayerCount || n > GutsMaxPlayerCount || n != j.Config.PlayerCount {
		return errGutsSnapshot
	}
	if len(j.Matchers) > gutsMaxSliceLen || len(j.ActionLog) > gutsMaxSliceLen {
		return errGutsSnapshot
	}
	if !gutsValidPhase(j.Phase) {
		return errGutsSnapshot
	}
	if j.RoundNumber < 1 || j.Pot < 0 || j.CarryPot < 0 {
		return errGutsSnapshot
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= n || j.MatchWinnerIdx < -1 || j.MatchWinnerIdx >= n {
		return errGutsSnapshot
	}
	if j.Result < GutsResultLose || j.Result > GutsResultWin {
		return errGutsSnapshot
	}
	for _, m := range j.Matchers {
		if m < 0 || m >= n {
			return errGutsSnapshot
		}
	}
	for _, p := range j.Players {
		if p == nil {
			return errGutsSnapshot
		}
	}
	for _, e := range j.ActionLog {
		if e == nil {
			return errGutsSnapshot
		}
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	g.config = j.Config
	g.state = gutsState{
		phase:          j.Phase,
		roundNumber:    j.RoundNumber,
		pot:            j.Pot,
		carryPot:       j.CarryPot,
		winnerIdx:      j.WinnerIdx,
		matchWinnerIdx: j.MatchWinnerIdx,
		matchers:       j.Matchers,
		result:         j.Result,
		gameEndFlag:    j.GameEndFlag,
		scored:         j.Scored,
		actionLogBase:  actionLogBase{actionLog: j.ActionLog},
	}
	if g.state.matchers == nil {
		g.state.matchers = make([]int, 0)
	}
	if g.state.actionLog == nil {
		g.state.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
