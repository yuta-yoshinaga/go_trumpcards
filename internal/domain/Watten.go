//go:build !js || !wasm || extra

// Package domain — ヴァッテン (Watten) のドメインモデル。
//
// Watten はバイエルン/オーストリアの 4 人 2 チーム制トリックテイキングに、
// ステーク (賭け点) の引き上げ (bluff-raise) を融合させたゲーム。座席 0・2 が
// チーム 0、座席 1・3 がチーム 1 (パートナーは対面に座る)。座席 0 が人間。
//
// 本実装は歴史的変種の多いトランプ体系を「クリーンで決定的な」独自ルールに
// 固定している (現実の Weli を含む 33 枚デッキではなく、標準 52 枚から 2〜6 を
// 除いた 32 枚のクリーンな部分集合を用いる)。
//
// カードの強さ (高→低):
//  1. Max   = ♥K (King of Hearts) — 常に最強 (固定)
//  2. Belli = ♦K (King of Diamonds) — 常に 2 番手 (固定)
//  3. Spitz = ♦7 (Seven of Diamonds) — 常に 3 番手 (固定)
//  4. Schlag: ディーラーが宣言したランク R を持つ全カード (Max/Belli/Spitz を除く)。
//     同士は固定スート順 ♥>♦>♠>♣ で比較。
//  5. Critical-suit: 宣言された切り札スート T の残りカード (Schlag/上記を除く)。
//     値 A>K>Q>J>10>9>8>7 で比較。
//  6. Plain (非トランプ): 自スート内で A>K>Q>J>10>9>8>7。同じリードスートのみを上回る。
//
// トランプ群 = {Max, Belli, Spitz, 全 Schlag カード, 全 Critical-suit カード}。
// マストフォロー: トランプがリードされたらトランプ保有者はトランプを出す。プレーン
// スートがリードされたらそのスート保有者はフォローする。ボイドなら任意。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// WattenPlayerCnt ヴァッテンのプレイヤー数
const WattenPlayerCnt = 4

// WattenHandSize 各プレイヤーの手札枚数
const WattenHandSize = 5

// WattenTeamCnt チーム数
const WattenTeamCnt = 2

// WattenDeckSize デッキ枚数 (32 枚)
const WattenDeckSize = 32

// WattenBaseStake ディールの基本ステーク (点)
const WattenBaseStake = 2

// WattenPhase ゲームフェーズ
type WattenPhase int

// Wattenのフェーズ定数 (ワイヤー値としても使用)
const (
	// WattenPhaseDeclare ディーラーが Schlag ランク + 切り札スートを宣言する
	WattenPhaseDeclare WattenPhase = 0
	// WattenPhasePlay トリックプレイフェーズ
	WattenPhasePlay WattenPhase = 1
	// WattenPhaseRespond レイズへの応答待ち (相手チームが hold/fold)
	WattenPhaseRespond WattenPhase = 2
	// WattenPhaseTrickEnd トリック終了 (自動解決される内部フェーズ)
	WattenPhaseTrickEnd WattenPhase = 3
	// WattenPhaseRoundEnd ディール終了 (結果表示; 次ディール待ち or ゲーム終了)
	WattenPhaseRoundEnd WattenPhase = 4
	// WattenPhaseGameEnd マッチ終了
	WattenPhaseGameEnd WattenPhase = 5
)

// WattenResult は人間 (チーム 0) から見たマッチ結果。
// GameResult は共有ファイル internal/domain/game_result.go に移動したので到達可能に
// なったが、この型名は JSON ペイロードに出るため統合していない（#4462）。値は
// GameResult と同一。
type WattenResult int

// Wattenの結果定数
const (
	// WattenResultLose 負け
	WattenResultLose WattenResult = -1
	// WattenResultNone 未確定
	WattenResultNone WattenResult = 0
	// WattenResultWin 勝ち
	WattenResultWin WattenResult = 1
)

// WattenHint ヒント情報。Action は "declare" / "play" / "raise" / "hold" / "fold"。
type WattenHint struct {
	Action    string // 推奨アクション種別
	CardIndex *int   // 推奨カードインデックス (Action=="play")
	Rank      *int   // 推奨 Schlag ランク (Action=="declare")
	Suit      *int   // 推奨切り札スート (Action=="declare")
	Reason    string // ヒント理由キー
}

// Watten ヴァッテンゲームクラス
type Watten struct {
	trumpCards       *TrumpCards
	players          []*WattenPlayer
	config           WattenConfig
	phase            WattenPhase
	roundNumber      int // ディール番号 (1 始まり)
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	dealerIdx        int
	leadPlayerIdx    int
	schlagRank       int // 宣言された Schlag ランク (0 = 未宣言)
	criticalSuit     int // 宣言された切り札スート (0 = 未宣言)
	stake            int // 現在の確定ステーク (WattenBaseStake..)
	pendingStake     int // 応答待ちで提示中のステーク (0 = 応答待ちでない)
	raiseCount       int // 確定済みレイズ回数
	raiserTeam       int // 応答待ちのレイズ実施チーム (-1 = なし)
	responderIdx     int // 応答すべきプレイヤー (-1 = なし)
	teamScores       [WattenTeamCnt]int
	teamTricks       [WattenTeamCnt]int
	dealWinnerTeam   int  // 直近ディールの勝者チーム (-1 = 未確定)
	scored           bool // ディール得点を加算済みか (フェーズ入場時ガード)
	gameEndFlag      bool
	winnerTeam       int          // マッチ勝者チーム (-1 = 未確定)
	result           WattenResult // 人間 (チーム 0) 視点の結果
	actionLogBase
}

// NewWatten コンストラクタ
func NewWatten(trumpCards *TrumpCards, players []*WattenPlayer, config WattenConfig) *Watten {
	return &Watten{
		trumpCards:     trumpCards,
		players:        players,
		config:         config,
		winnerTeam:     -1,
		dealWinnerTeam: -1,
		raiserTeam:     -1,
		responderIdx:   -1,
		roundNumber:    0,
		dealerIdx:      0,
	}
}

// NewDefaultWatten 標準4人パートナーシップ構成 (人間チーム0, CPUは交互配置)。
func NewDefaultWatten() *Watten {
	players := []*WattenPlayer{
		NewWattenPlayer(true, 0),
		NewWattenPlayer(false, 1),
		NewWattenPlayer(false, 0),
		NewWattenPlayer(false, 1),
	}
	return NewWatten(newWattenDeck(), players, DefaultWattenConfig())
}

// newWattenDeck ヴァッテン用 32 枚デッキを生成する。
// NewTrumpCards(0) の標準 52 枚から 2,3,4,5,6 を除外して構築する
// (A,7,8,9,10,J,Q,K × 4 スート = 32 枚)。TrumpCards.go を汚さないよう
// extra タグ付き Watten.go 内に自己完結させる。
func newWattenDeck() *TrumpCards {
	full := NewTrumpCards(0) // 標準 52 枚
	keep := map[int]bool{1: true, 7: true, 8: true, 9: true, 10: true, 11: true, 12: true, 13: true}
	t := new(TrumpCards)
	t.deck = make([]*Card, 0, WattenDeckSize)
	for _, c := range full.deck {
		if keep[c.GetValue()] {
			t.deck = append(t.deck, NewCard(c.GetDesign(), c.GetValue(), false))
		}
	}
	t.deckCnt = len(t.deck)
	t.deckInit()
	return t
}

// Reset ゲーム初期化
func (g *Watten) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.result = WattenResultNone
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamScores = [WattenTeamCnt]int{}
	g.actionLog = nil

	for _, p := range g.players {
		p.ResetRound()
	}
	g.beginRound()
}

// NextRound 次のディールを開始する
func (g *Watten) NextRound() {
	if g.phase != WattenPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % WattenPlayerCnt

	for _, p := range g.players {
		p.ResetRound()
	}
	g.beginRound()
}

// beginRound ディールの初期処理 (配布 + 宣言フェーズ突入)
func (g *Watten) beginRound() {
	g.trickNumber = 0
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.schlagRank = 0
	g.criticalSuit = 0
	g.stake = WattenBaseStake
	g.pendingStake = 0
	g.raiseCount = 0
	g.raiserTeam = -1
	g.responderIdx = -1
	g.teamTricks = [WattenTeamCnt]int{}
	g.dealWinnerTeam = -1
	g.scored = false

	g.dealAll()
	g.currentPlayerIdx = g.dealerIdx
	g.phase = WattenPhaseDeclare
}

// dealAll 各プレイヤーに 5 枚配る (20 枚消費)
func (g *Watten) dealAll() {
	g.trumpCards.Shuffle()
	for range WattenHandSize {
		for j := range WattenPlayerCnt {
			if card := g.trumpCards.DrawCard(); card != nil {
				g.players[j].AddCard(card)
			}
		}
	}
}

// --- Declaration ---

// PlayerDeclare 人間ディーラーが Schlag ランク + 切り札スートを宣言する
func (g *Watten) PlayerDeclare(rank, suit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != WattenPhaseDeclare {
		return ErrWrongPhase
	}
	if !g.IsHumanDeclareTurn() {
		return ErrNotHumanTurn
	}
	if !isValidSchlagRank(rank) {
		return NewDomainError(ErrInvalidPlay, "無効な Schlag ランクです")
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return NewDomainError(ErrInvalidPlay, "無効な切り札スートです")
	}
	g.doDeclare(g.dealerIdx, rank, suit)
	return nil
}

// CpuDeclare CPUディーラーが宣言する (最頻スートを T, 最頻ランクを R とする)。
func (g *Watten) CpuDeclare() {
	if g.gameEndFlag || g.phase != WattenPhaseDeclare {
		return
	}
	if g.players[g.dealerIdx].GetIsHuman() {
		return
	}
	rank, suit := g.cpuBestDeclaration(g.dealerIdx)
	g.doDeclare(g.dealerIdx, rank, suit)
}

// doDeclare 宣言を確定し、プレイフェーズへ移行する
func (g *Watten) doDeclare(playerIdx, rank, suit int) {
	g.schlagRank = rank
	g.criticalSuit = suit
	g.appendLog(playerIdx, "declare",
		fmt.Sprintf("%s declares Schlag=%d critical=%s",
			playerName(g.players, playerIdx), rank, suitStr(suit)), nil)
	g.sortAllHands()
	g.startPlayPhase()
}

// startPlayPhase プレイフェーズを開始する
func (g *Watten) startPlayPhase() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.leadPlayerIdx = (g.dealerIdx + 1) % WattenPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = WattenPhasePlay
}

// --- Play ---

// PlayerPlay 人間プレイヤーがカードをプレイする
func (g *Watten) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != WattenPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	card := player.GetCard(cardIndex)
	if err := g.validatePlay(g.currentPlayerIdx, card); err != nil {
		return err
	}
	played := player.RemoveCard(cardIndex)
	g.playCard(g.currentPlayerIdx, played)
	return nil
}

// CpuPlay CPUプレイヤーが1ターン実行する (リード時にレイズ判断を含む)。
func (g *Watten) CpuPlay() {
	if g.gameEndFlag || g.phase != WattenPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	// リード (トリック先頭) かつレイズ可能なら強い手でレイズする。
	if len(g.currentTrick) == 0 && g.canRaise(idx) && g.cpuWantsToRaise(idx) {
		g.callRaise(g.players[idx].GetTeam())
		return
	}
	cardIdx := g.cpuSelectPlayCard(idx)
	played := g.players[idx].RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// playCard カードをプレイする共通処理
func (g *Watten) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == WattenPlayerCnt {
		g.phase = WattenPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % WattenPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する
func (g *Watten) ResolveTrick() {
	if g.phase != WattenPhaseTrickEnd || len(g.currentTrick) != WattenPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
	}
	winnerTeam := g.players[winnerIdx].GetTeam()
	g.players[winnerIdx].AddTrick(trickCards)
	g.teamTricks[winnerTeam]++
	g.leadPlayerIdx = winnerIdx
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", playerName(g.players, winnerIdx), g.trickNumber), trickCards)

	if g.trickNumber >= WattenHandSize {
		g.enterRoundEnd(g.dealWinnerByTricks())
		return
	}
	g.phase = WattenPhaseTrickEnd
}

// NextTrick 次のトリックを開始する
func (g *Watten) NextTrick() {
	if g.phase != WattenPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = WattenPhasePlay
}

// dealWinnerByTricks 5トリック終了時に3トリック以上取ったチームを返す
func (g *Watten) dealWinnerByTricks() int {
	if g.teamTricks[0] >= g.teamTricks[1] {
		return 0
	}
	return 1
}

// --- Raise / respond (stake mechanic) ---

// canRaise playerIdx が今レイズ可能か (リード時 & 上限未満 & 応答待ちでない)。
func (g *Watten) canRaise(playerIdx int) bool {
	if g.phase != WattenPhasePlay || g.pendingStake != 0 {
		return false
	}
	if len(g.currentTrick) != 0 || g.currentPlayerIdx != playerIdx {
		return false
	}
	return g.raiseCount < g.config.MaxRaises
}

// CanHumanRaise 現在の手番の人間 (リード) がレイズ可能かを返す。
func (g *Watten) CanHumanRaise() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman() && g.canRaise(g.currentPlayerIdx)
}

// PlayerRaise 人間プレイヤー (リード) がステークを引き上げる。
func (g *Watten) PlayerRaise() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != WattenPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !g.canRaise(g.currentPlayerIdx) {
		return NewDomainError(ErrWrongPhase, "今はレイズできません")
	}
	g.callRaise(g.players[g.currentPlayerIdx].GetTeam())
	return nil
}

// callRaise team がステークの引き上げを宣言し、相手チームの応答待ちにする。
func (g *Watten) callRaise(team int) {
	g.pendingStake = g.stake + 1
	g.raiserTeam = team
	oppTeam := 1 - team
	g.responderIdx = g.teamRepresentative(oppTeam)
	g.phase = WattenPhaseRespond
	g.appendLog(g.teamRepresentative(team), "raise",
		fmt.Sprintf("Team %d raises stake to %d", team, g.pendingStake), nil)
}

// teamRepresentative チームの代表応答プレイヤーを返す (チーム 0 → 人間 0, チーム 1 → 1)。
func (g *Watten) teamRepresentative(team int) int {
	for i, p := range g.players {
		if p.GetTeam() == team && p.GetIsHuman() {
			return i
		}
	}
	for i, p := range g.players {
		if p.GetTeam() == team {
			return i
		}
	}
	return 0
}

// PlayerRespond 人間プレイヤーがレイズに応答する (true=hold / false=fold)。
func (g *Watten) PlayerRespond(hold bool) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != WattenPhaseRespond {
		return ErrWrongPhase
	}
	if g.responderIdx < 0 || !g.players[g.responderIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.respond(g.responderIdx, hold)
	return nil
}

// CpuRespond CPU が応答フェーズで hold/fold を判断する。
func (g *Watten) CpuRespond() {
	if g.gameEndFlag || g.phase != WattenPhaseRespond {
		return
	}
	if g.responderIdx < 0 || g.players[g.responderIdx].GetIsHuman() {
		return
	}
	g.respond(g.responderIdx, g.cpuWantsToHold(g.responderIdx))
}

// respond responder が hold/fold を確定する。
func (g *Watten) respond(responder int, hold bool) {
	if hold {
		g.stake = g.pendingStake
		g.raiseCount++
		g.pendingStake = 0
		raiser := g.raiserTeam
		g.raiserTeam = -1
		g.responderIdx = -1
		g.phase = WattenPhasePlay
		// レイズしたチームのリードプレイヤーに手番を戻す。
		g.currentPlayerIdx = g.leadPlayerIdx
		g.appendLog(responder, "hold",
			fmt.Sprintf("%s holds (stake %d, team %d leads)",
				playerName(g.players, responder), g.stake, raiser), nil)
		return
	}
	// fold: レイズしたチームが直前の確定ステークでディールを取る。
	raiser := g.raiserTeam
	g.pendingStake = 0
	g.raiserTeam = -1
	g.responderIdx = -1
	g.appendLog(responder, "fold",
		fmt.Sprintf("%s folds; team %d wins deal (%d pt)",
			playerName(g.players, responder), raiser, g.stake), nil)
	g.enterRoundEnd(raiser)
}

// --- Deal-end scoring (fires on phase entry, guarded by `scored`) ---

// enterRoundEnd ディール終了フェーズに入る。勝者チームにステークを加算する。
func (g *Watten) enterRoundEnd(winnerTeam int) {
	g.dealWinnerTeam = winnerTeam
	g.scoreDeal()
	if !g.gameEndFlag {
		g.phase = WattenPhaseRoundEnd
	}
}

// scoreDeal ディール得点を加算する (scored フラグで冪等)。
func (g *Watten) scoreDeal() {
	if g.scored || g.dealWinnerTeam < 0 {
		return
	}
	g.scored = true
	g.teamScores[g.dealWinnerTeam] += g.stake
	g.appendLog(-1, "deal_score",
		fmt.Sprintf("Team %d wins deal +%d (match %d-%d)",
			g.dealWinnerTeam, g.stake, g.teamScores[0], g.teamScores[1]), nil)
	g.checkGameEnd()
}

// ScoreRound インタラクター NextRound から呼ばれる。得点は入場時に加算済みのため冪等。
func (g *Watten) ScoreRound() {
	if g.phase != WattenPhaseRoundEnd && g.phase != WattenPhaseGameEnd {
		return
	}
	g.scoreDeal()
}

// checkGameEnd マッチ終了判定を行う。
func (g *Watten) checkGameEnd() {
	for ti := range WattenTeamCnt {
		if g.teamScores[ti] >= g.config.TargetScore {
			g.gameEndFlag = true
			g.phase = WattenPhaseGameEnd
			if g.teamScores[0] >= g.teamScores[1] {
				g.winnerTeam = 0
			} else {
				g.winnerTeam = 1
			}
			if g.winnerTeam == 0 {
				g.result = WattenResultWin
			} else {
				g.result = WattenResultLose
			}
			g.appendLog(-1, "game_end",
				fmt.Sprintf("Team %d wins the match!", g.winnerTeam), nil)
			return
		}
	}
}

// --- Ranking ---

// isValidSchlagRank R として有効なランク (A=1, 7..13) か。
func isValidSchlagRank(rank int) bool {
	switch rank {
	case 1, 7, 8, 9, 10, 11, 12, 13:
		return true
	}
	return false
}

// isMax ♥K か
func isMax(c *Card) bool { return c.GetDesign() == CardDesignHeart && c.GetValue() == 13 }

// isBelli ♦K か
func isBelli(c *Card) bool { return c.GetDesign() == CardDesignDiamond && c.GetValue() == 13 }

// isSpitz ♦7 か
func isSpitz(c *Card) bool { return c.GetDesign() == CardDesignDiamond && c.GetValue() == 7 }

// isTrump カードがトランプ群に属するか (Max/Belli/Spitz/Schlag/Critical)。
func (g *Watten) isTrump(c *Card) bool {
	if c == nil {
		return false
	}
	if isMax(c) || isBelli(c) || isSpitz(c) {
		return true
	}
	if g.schlagRank != 0 && c.GetValue() == g.schlagRank {
		return true
	}
	if g.criticalSuit != 0 && c.GetDesign() == g.criticalSuit {
		return true
	}
	return false
}

// WattenTrumpPreview counts, for a hand, how many cards would become trumps for
// each candidate declaration. `Permanent` is the number of Max/Belli/Spitz held
// (trumps whatever is declared), `BySuit` counts the extra cards a critical
// suit would add, and `ByRank` the extra cards a Schlag rank would add — extra
// meaning "not already counted as permanent", so the three never double-count.
//
// The Web GUI previews the same thing live while the dealer picks (#4848); the
// CUI prompt showed only the command syntax, and the declaration decides who
// holds the initiative for the whole deal.
type WattenTrumpPreview struct {
	Permanent int
	BySuit    [CardDesignDiamond + 1]int
	ByRank    map[int]int
}

// WattenPreviewTrumps builds a WattenTrumpPreview for the given hand.
func WattenPreviewTrumps(cards []*Card) WattenTrumpPreview {
	pv := WattenTrumpPreview{ByRank: map[int]int{}}
	for _, c := range cards {
		if c == nil {
			continue
		}
		if isMax(c) || isBelli(c) || isSpitz(c) {
			pv.Permanent++
			continue
		}
		if d := c.GetDesign(); d >= CardDesignSpade && d <= CardDesignDiamond {
			pv.BySuit[d]++
		}
		pv.ByRank[c.GetValue()]++
	}
	return pv
}

// IsTrumpPublic テスト用公開ラッパー。
func (g *Watten) IsTrumpPublic(c *Card) bool { return g.isTrump(c) }

// schlagSuitOrder Schlag 同士の固定スート順 ♥>♦>♠>♣。
func schlagSuitOrder(design int) int {
	switch design {
	case CardDesignHeart:
		return 3
	case CardDesignDiamond:
		return 2
	case CardDesignSpade:
		return 1
	case CardDesignClover:
		return 0
	}
	return 0
}

// wattenValueRank 値の強さ A>K>Q>J>10>9>8>7 (Critical-suit / Plain 用)。
func wattenValueRank(value int) int {
	switch value {
	case 1: // A
		return 8
	case 13: // K
		return 7
	case 12: // Q
		return 6
	case 11: // J
		return 5
	case 10:
		return 4
	case 9:
		return 3
	case 8:
		return 2
	case 7:
		return 1
	}
	return 0
}

// cardRank トリック比較用ランクを返す (高い = 強い)。
// トランプ群は 800..1000 の重ならないランクを持ち、プレーンは 1..8 (値のみ)。
func (g *Watten) cardRank(c *Card) int {
	switch {
	case isMax(c):
		return 1000
	case isBelli(c):
		return 999
	case isSpitz(c):
		return 998
	case g.schlagRank != 0 && c.GetValue() == g.schlagRank:
		return 900 + schlagSuitOrder(c.GetDesign())
	case g.criticalSuit != 0 && c.GetDesign() == g.criticalSuit:
		return 800 + wattenValueRank(c.GetValue())
	default:
		return wattenValueRank(c.GetValue())
	}
}

// CardRankPublic テスト用公開メソッド。
func (g *Watten) CardRankPublic(c *Card) int { return g.cardRank(c) }

// --- Trick play helpers ---

// validatePlay マストフォロー検証。
func (g *Watten) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	player := g.players[playerIdx]
	lead := g.currentTrick[0].Card

	if g.isTrump(lead) {
		// トランプがリード: トランプ保有者はトランプを出す。
		if g.playerHasTrump(player) && !g.isTrump(card) {
			return NewDomainError(ErrInvalidPlay, "トランプに従ってください")
		}
		return nil
	}
	// プレーンスートがリード: そのスートの非トランプ札を持つならフォロー必須。
	leadSuit := lead.GetDesign()
	if g.playerHasPlainSuit(player, leadSuit) {
		if g.isTrump(card) || card.GetDesign() != leadSuit {
			return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
		}
	}
	return nil
}

// playerHasTrump プレイヤーがトランプ札を持っているか
func (g *Watten) playerHasTrump(p *WattenPlayer) bool {
	for i := range p.GetCardsSize() {
		if g.isTrump(p.GetCard(i)) {
			return true
		}
	}
	return false
}

// playerHasPlainSuit プレイヤーが指定プレーンスートの (非トランプ) 札を持っているか
func (g *Watten) playerHasPlainSuit(p *WattenPlayer, suit int) bool {
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if !g.isTrump(c) && c.GetDesign() == suit {
			return true
		}
	}
	return false
}

// currentLeader 現在のトリックの仮勝者プレイヤーインデックスを返す。
func (g *Watten) currentLeader() int {
	if len(g.currentTrick) == 0 {
		return -1
	}
	first := g.currentTrick[0]
	winner := first.PlayerIdx
	winTrump := g.isTrump(first.Card)
	winRank := g.cardRank(first.Card)
	winSuit := first.Card.GetDesign()
	for _, tc := range g.currentTrick[1:] {
		t := g.isTrump(tc.Card)
		rank := g.cardRank(tc.Card)
		switch {
		case t && !winTrump:
			winner, winTrump, winRank, winSuit = tc.PlayerIdx, true, rank, tc.Card.GetDesign()
		case t && winTrump:
			if rank > winRank {
				winner, winRank, winSuit = tc.PlayerIdx, rank, tc.Card.GetDesign()
			}
		case !t && winTrump:
			// プレーンはトランプに勝てない。
		default:
			// 両者プレーン: 同じリードスートのみ比較。
			if tc.Card.GetDesign() == winSuit && rank > winRank {
				winner, winRank = tc.PlayerIdx, rank
			}
		}
	}
	return winner
}

// trickWinner トリックの勝者を決定する。
func (g *Watten) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	return g.currentLeader()
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Watten) GetValidPlayIndices(playerIdx int) []int {
	return g.getValidPlayIndices(playerIdx)
}

func (g *Watten) getValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Sorting / bookkeeping ---

func (g *Watten) sortAllHands() {
	for _, p := range g.players {
		cards := make([]*Card, p.GetCardsSize())
		for i := range p.GetCardsSize() {
			cards[i] = p.GetCard(i)
		}
		sort.Slice(cards, func(i, j int) bool {
			ti := g.isTrump(cards[i])
			tj := g.isTrump(cards[j])
			if ti != tj {
				return ti // トランプを先頭に
			}
			if !ti {
				si, sj := cards[i].GetDesign(), cards[j].GetDesign()
				if si != sj {
					return si < sj
				}
			}
			return g.cardRank(cards[i]) > g.cardRank(cards[j])
		})
		p.Reset()
		for _, c := range cards {
			p.AddCard(c)
		}
	}
}

// --- CPU AI ---

// cpuBestDeclaration CPU の宣言: 最頻スートを T, 最頻ランクを R とする。
func (g *Watten) cpuBestDeclaration(playerIdx int) (rank, suit int) {
	p := g.players[playerIdx]
	suitCount := map[int]int{}
	rankCount := map[int]int{}
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		suitCount[c.GetDesign()]++
		if isValidSchlagRank(c.GetValue()) {
			rankCount[c.GetValue()]++
		}
	}
	suit = CardDesignSpade
	bestS := -1
	for _, s := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if suitCount[s] > bestS {
			bestS = suitCount[s]
			suit = s
		}
	}
	rank = 1
	bestR := -1
	for _, r := range []int{1, 7, 8, 9, 10, 11, 12, 13} {
		if rankCount[r] > bestR {
			bestR = rankCount[r]
			rank = r
		}
	}
	return rank, suit
}

// cpuSelectPlayCard CPU が出すべきカードのインデックスを選択する。
func (g *Watten) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if g.config.CpuDifficulty == WattenCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlayChoose(playerIdx, valid)
}

// cpuPlayChoose 標準ヒューリスティック:
//   - リード時: 最強札を出す。
//   - フォロー時: パートナーが勝っていれば最弱札を温存、そうでなければ勝てる最弱札、
//     勝てなければ最弱札を捨てる。
func (g *Watten) cpuPlayChoose(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		best, bestRank := valid[0], -1
		for _, idx := range valid {
			if r := g.cardRank(player.GetCard(idx)); r > bestRank {
				bestRank, best = r, idx
			}
		}
		return best
	}
	winnerIdx := g.currentLeader()
	partnerIdx := (playerIdx + 2) % WattenPlayerCnt
	if winnerIdx == partnerIdx {
		return g.weakestIdx(playerIdx, valid)
	}
	// 勝てる最弱札を探す。
	winnable, winnableRank := -1, 100000
	for _, idx := range valid {
		if g.cardWouldWinTrick(player.GetCard(idx)) {
			if r := g.cardRank(player.GetCard(idx)); r < winnableRank {
				winnableRank, winnable = r, idx
			}
		}
	}
	if winnable >= 0 {
		return winnable
	}
	return g.weakestIdx(playerIdx, valid)
}

// weakestIdx valid の中で最弱ランクのインデックスを返す。
func (g *Watten) weakestIdx(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	worst, worstRank := valid[0], 100000
	for _, idx := range valid {
		if r := g.cardRank(player.GetCard(idx)); r < worstRank {
			worstRank, worst = r, idx
		}
	}
	return worst
}

// cardWouldWinTrick 指定カードを今出した場合に現状の仮勝者を上回るか。
func (g *Watten) cardWouldWinTrick(c *Card) bool {
	if len(g.currentTrick) == 0 {
		return true
	}
	winIdx := g.currentLeader()
	var winCard *Card
	for _, tc := range g.currentTrick {
		if tc.PlayerIdx == winIdx {
			winCard = tc.Card
			break
		}
	}
	if winCard == nil {
		return true
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	cTrump := g.isTrump(c)
	wTrump := g.isTrump(winCard)
	switch {
	case cTrump && !wTrump:
		return true
	case cTrump && wTrump:
		return g.cardRank(c) > g.cardRank(winCard)
	case !cTrump && wTrump:
		return false
	default:
		if c.GetDesign() == winCard.GetDesign() {
			return g.cardRank(c) > g.cardRank(winCard)
		}
		// リードスートを上回る必要がある。
		return c.GetDesign() == leadSuit && g.cardRank(c) > g.cardRank(winCard)
	}
}

// handStrength 手札のトランプ強度合計を返す (レイズ/応答判断用)。
func (g *Watten) handStrength(playerIdx int) int {
	p := g.players[playerIdx]
	score := 0
	for i := range p.GetCardsSize() {
		if r := g.cardRank(p.GetCard(i)); r >= 800 {
			score += r - 790 // トランプに応じた重み
		}
	}
	return score
}

// cpuWantsToRaise CPU がレイズしたいか (強い手ほど積極的)。
func (g *Watten) cpuWantsToRaise(playerIdx int) bool {
	s := g.handStrength(playerIdx)
	r := rand.Float64()
	threshold := 60
	switch g.config.CpuDifficulty {
	case WattenCpuDifficultyEasy:
		threshold = 90
	case WattenCpuDifficultyHard:
		threshold = 45
	}
	if s >= threshold {
		return r < 0.6
	}
	return r < 0.05 // 稀にブラフ
}

// cpuWantsToHold CPU が応答フェーズで hold するか (非常に弱い手のみ fold)。
func (g *Watten) cpuWantsToHold(playerIdx int) bool {
	s := g.handStrength(playerIdx)
	r := rand.Float64()
	if s >= 40 {
		return true
	}
	if s >= 15 {
		return r < 0.8
	}
	return r < 0.4 // 弱い手でも稀に hold
}

// --- Hints ---

// GetHint 現フェーズのヒントを返す (人間プレイヤー視点)。
func (g *Watten) GetHint() *WattenHint {
	humanIdx := findHumanIdx(g.players)
	if humanIdx < 0 {
		return nil
	}
	switch g.phase {
	case WattenPhaseDeclare:
		if g.dealerIdx != humanIdx {
			return nil
		}
		rank, suit := g.cpuBestDeclaration(humanIdx)
		return &WattenHint{Action: "declare", Rank: &rank, Suit: &suit, Reason: "declare_strong"}
	case WattenPhasePlay:
		if g.currentPlayerIdx != humanIdx {
			return nil
		}
		if g.canRaise(humanIdx) && g.cpuWantsToRaise(humanIdx) {
			return &WattenHint{Action: "raise", Reason: "raise_strong"}
		}
		valid := g.getValidPlayIndices(humanIdx)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlayChoose(humanIdx, valid)
		return &WattenHint{Action: "play", CardIndex: &idx, Reason: g.playHintReason(humanIdx, idx)}
	case WattenPhaseRespond:
		if g.responderIdx != humanIdx {
			return nil
		}
		if g.cpuWantsToHold(humanIdx) {
			return &WattenHint{Action: "hold", Reason: "hold_ok"}
		}
		return &WattenHint{Action: "fold", Reason: "fold_weak"}
	}
	return nil
}

func (g *Watten) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	if len(g.currentTrick) == 0 {
		if g.isTrump(card) {
			return "lead_trump"
		}
		return "lead_plain"
	}
	if g.cardWouldWinTrick(card) {
		return "follow_win"
	}
	return "follow_dump"
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ取得
func (g *Watten) GetPhase() WattenPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Watten) SetPhase(p WattenPhase) { g.phase = p }

// GetRoundNumber 現在のディール番号取得
func (g *Watten) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ディール番号設定 (テスト用)
func (g *Watten) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber 現在のトリック番号取得
func (g *Watten) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Watten) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Watten) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Watten) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Watten) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Watten) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetDealerIdx ディーラーインデックス取得
func (g *Watten) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (g *Watten) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Watten) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Watten) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetSchlagRank Schlag ランク取得 (0 = 未宣言)
func (g *Watten) GetSchlagRank() int { return g.schlagRank }

// SetSchlagRank Schlag ランク設定 (テスト用)
func (g *Watten) SetSchlagRank(r int) { g.schlagRank = r }

// GetCriticalSuit 切り札スート取得 (0 = 未宣言)
func (g *Watten) GetCriticalSuit() int { return g.criticalSuit }

// SetCriticalSuit 切り札スート設定 (テスト用)
func (g *Watten) SetCriticalSuit(s int) { g.criticalSuit = s }

// GetStake 現在の確定ステーク取得
func (g *Watten) GetStake() int { return g.stake }

// SetStake ステーク設定 (テスト用)
func (g *Watten) SetStake(v int) { g.stake = v }

// GetPendingStake 応答待ちで提示中のステーク取得 (0 = 応答待ちでない)
func (g *Watten) GetPendingStake() int { return g.pendingStake }

// GetRaiseCount 確定済みレイズ回数取得
func (g *Watten) GetRaiseCount() int { return g.raiseCount }

// GetRaiserTeam 応答待ちのレイズ実施チーム取得 (-1 = なし)
func (g *Watten) GetRaiserTeam() int { return g.raiserTeam }

// GetResponderIdx 応答すべきプレイヤーインデックス取得 (-1 = なし)
func (g *Watten) GetResponderIdx() int { return g.responderIdx }

// SetResponderIdx 応答プレイヤー設定 (テスト用)
func (g *Watten) SetResponderIdx(idx int) { g.responderIdx = idx }

// GetTeamScore チームスコア取得
func (g *Watten) GetTeamScore(team int) int {
	if team < 0 || team >= WattenTeamCnt {
		return 0
	}
	return g.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (g *Watten) SetTeamScore(team, score int) {
	if team >= 0 && team < WattenTeamCnt {
		g.teamScores[team] = score
	}
}

// GetTeamTricks 当ディールのチーム別トリック数取得
func (g *Watten) GetTeamTricks(team int) int {
	if team < 0 || team >= WattenTeamCnt {
		return 0
	}
	return g.teamTricks[team]
}

// GetDealWinnerTeam 直近ディールの勝者チーム取得 (-1 = 未確定)
func (g *Watten) GetDealWinnerTeam() int { return g.dealWinnerTeam }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Watten) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (g *Watten) GetWinnerTeam() int { return g.winnerTeam }

// GetResult 人間 (チーム 0) 視点のマッチ結果取得
func (g *Watten) GetResult() WattenResult { return g.result }

// GetPlayerCnt プレイヤー数取得
func (g *Watten) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Watten) GetPlayer(i int) *WattenPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在のプレイ手番が人間かどうか
func (g *Watten) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// IsHumanDeclareTurn 現在の宣言手番 (ディーラー) が人間かどうか
func (g *Watten) IsHumanDeclareTurn() bool {
	if g.dealerIdx < 0 || g.dealerIdx >= len(g.players) {
		return false
	}
	return g.players[g.dealerIdx].GetIsHuman()
}

// IsHumanRespondTurn 現在の応答手番が人間かどうか
func (g *Watten) IsHumanRespondTurn() bool {
	if g.responderIdx < 0 || g.responderIdx >= len(g.players) {
		return false
	}
	return g.players[g.responderIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Watten) GetConfig() WattenConfig { return g.config }

// SetConfig 設定変更
func (g *Watten) SetConfig(cfg WattenConfig) { g.config = cfg }

// --- Test-only helpers ---

// GetConfigDeckHelper returns a fresh 32-card Watten deck (テスト用コンストラクタ補助)。
func (g *Watten) GetConfigDeckHelper() *TrumpCards { return newWattenDeck() }

// SetupRaiseForTest configures a pending-raise/respond state (テスト用)。
func (g *Watten) SetupRaiseForTest(pending, raiserTeam, responderIdx int) {
	g.phase = WattenPhaseRespond
	g.pendingStake = pending
	g.raiserTeam = raiserTeam
	g.responderIdx = responderIdx
}

// SetTeamTricksForTest sets a team's trick count for the current deal (テスト用)。
func (g *Watten) SetTeamTricksForTest(team, n int) {
	if team >= 0 && team < WattenTeamCnt {
		g.teamTricks[team] = n
	}
}

// SetRaiseCountForTest sets the accepted raise count (テスト用)。
func (g *Watten) SetRaiseCountForTest(n int) { g.raiseCount = n }

// --- JSON ---

// wattenJSON Watten の JSON 表現
type wattenJSON struct {
	TrumpCards       *TrumpCards        `json:"tc"`
	Players          []*WattenPlayer    `json:"pl"`
	Config           WattenConfig       `json:"cfg"`
	Phase            WattenPhase        `json:"ph"`
	RoundNumber      int                `json:"rn"`
	TrickNumber      int                `json:"tn"`
	CurrentPlayerIdx int                `json:"cp"`
	CurrentTrick     []*TrickCard       `json:"ct"`
	DealerIdx        int                `json:"di"`
	LeadPlayerIdx    int                `json:"li"`
	SchlagRank       int                `json:"sr"`
	CriticalSuit     int                `json:"cs"`
	Stake            int                `json:"st"`
	PendingStake     int                `json:"ps"`
	RaiseCount       int                `json:"rc"`
	RaiserTeam       int                `json:"rt"`
	ResponderIdx     int                `json:"ri"`
	TeamScores       [WattenTeamCnt]int `json:"sc"`
	TeamTricks       [WattenTeamCnt]int `json:"tt"`
	DealWinnerTeam   int                `json:"dw"`
	Scored           bool               `json:"sd"`
	GameEndFlag      bool               `json:"ge"`
	WinnerTeam       int                `json:"wt"`
	Result           WattenResult       `json:"re"`
	ActionLog        []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Watten) MarshalJSON() ([]byte, error) {
	return json.Marshal(wattenJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		DealerIdx:        g.dealerIdx,
		LeadPlayerIdx:    g.leadPlayerIdx,
		SchlagRank:       g.schlagRank,
		CriticalSuit:     g.criticalSuit,
		Stake:            g.stake,
		PendingStake:     g.pendingStake,
		RaiseCount:       g.raiseCount,
		RaiserTeam:       g.raiserTeam,
		ResponderIdx:     g.responderIdx,
		TeamScores:       g.teamScores,
		TeamTricks:       g.teamTricks,
		DealWinnerTeam:   g.dealWinnerTeam,
		Scored:           g.scored,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		Result:           g.result,
		ActionLog:        g.actionLog,
	})
}

// wattenMaxSliceLen はデシリアライズ時のスライス長上限 (DoS 対策)。
const wattenMaxSliceLen = 2000

// UnmarshalJSON implements json.Unmarshaler. 列挙値・インデックス範囲・スライス要素を
// 検証し、不正な場合はエラーを返す (パニックさせない)。
func (g *Watten) UnmarshalJSON(data []byte) error {
	var j wattenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < WattenPhaseDeclare || j.Phase > WattenPhaseGameEnd {
		return NewDomainError(ErrInvalidPlay, "無効なフェーズです")
	}
	if len(j.Players) != WattenPlayerCnt {
		return NewDomainError(ErrInvalidPlay, "プレイヤー数が不正です")
	}
	for i, p := range j.Players {
		if p == nil {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("プレイヤー %d が nil です", i))
		}
		if t := p.GetTeam(); t < 0 || t >= WattenTeamCnt {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("プレイヤー %d のチームが範囲外です", i))
		}
	}
	if len(j.CurrentTrick) > WattenPlayerCnt {
		return NewDomainError(ErrInvalidPlay, "トリックカードが多すぎます")
	}
	for i, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("トリックカード %d が nil です", i))
		}
		if tc.PlayerIdx < 0 || tc.PlayerIdx >= WattenPlayerCnt {
			return NewDomainError(ErrInvalidPlay, "トリックカードのプレイヤーインデックスが範囲外です")
		}
	}
	if len(j.ActionLog) > wattenMaxSliceLen {
		return NewDomainError(ErrInvalidPlay, "アクションログが大きすぎます")
	}
	// 未宣言 (0) は許可。宣言済みの場合のみ範囲チェック。
	if j.SchlagRank != 0 && !isValidSchlagRank(j.SchlagRank) {
		return NewDomainError(ErrInvalidPlay, "無効な Schlag ランクです")
	}
	if j.CriticalSuit != 0 && (j.CriticalSuit < CardDesignSpade || j.CriticalSuit > CardDesignDiamond) {
		return NewDomainError(ErrInvalidPlay, "無効な切り札スートです")
	}
	// 宣言後は Schlag/critical が必須。
	if j.Phase != WattenPhaseDeclare && (j.SchlagRank == 0 || j.CriticalSuit == 0) {
		return NewDomainError(ErrInvalidPlay, "宣言後は Schlag と切り札が必要です")
	}
	// 0 起点で必須のインデックス。
	if j.CurrentPlayerIdx < 0 || j.CurrentPlayerIdx >= WattenPlayerCnt {
		return NewDomainError(ErrInvalidPlay, "currentPlayerIdx が範囲外です")
	}
	if j.DealerIdx < 0 || j.DealerIdx >= WattenPlayerCnt {
		return NewDomainError(ErrInvalidPlay, "dealerIdx が範囲外です")
	}
	// -1 センチネル許可のプレイヤーインデックス (0..WattenPlayerCnt-1)。
	for _, v := range []int{j.LeadPlayerIdx, j.ResponderIdx} {
		if v < -1 || v >= WattenPlayerCnt {
			return NewDomainError(ErrInvalidPlay, "プレイヤーインデックスが範囲外です")
		}
	}
	// -1 センチネル許可のチームインデックス (0..WattenTeamCnt-1)。teamScores/teamTricks
	// を直接インデックスするため WattenTeamCnt で検証する (WattenPlayerCnt ではない)。
	for _, v := range []int{j.RaiserTeam, j.DealWinnerTeam, j.WinnerTeam} {
		if v < -1 || v >= WattenTeamCnt {
			return NewDomainError(ErrInvalidPlay, "チームインデックスが範囲外です")
		}
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newWattenDeck()
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	g.dealerIdx = j.DealerIdx
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.schlagRank = j.SchlagRank
	g.criticalSuit = j.CriticalSuit
	g.stake = j.Stake
	if g.stake < WattenBaseStake {
		g.stake = WattenBaseStake
	}
	g.pendingStake = j.PendingStake
	g.raiseCount = j.RaiseCount
	if g.raiseCount < 0 {
		g.raiseCount = 0
	}
	g.raiserTeam = j.RaiserTeam
	g.responderIdx = j.ResponderIdx
	g.teamScores = j.TeamScores
	g.teamTricks = j.TeamTricks
	g.dealWinnerTeam = j.DealWinnerTeam
	g.scored = j.Scored
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.result = j.Result
	g.actionLog = j.ActionLog
	return nil
}
