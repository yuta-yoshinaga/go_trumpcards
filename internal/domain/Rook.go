//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// RookPlayerCnt ルークのプレイヤー数
const RookPlayerCnt = 4

// RookHandSize 各プレイヤーの手札枚数
const RookHandSize = 13

// RookNestSize ネスト(場札)枚数
const RookNestSize = 5

// RookTeamCnt チーム数
const RookTeamCnt = 2

// RookTrickCnt 1ラウンドのトリック数
const RookTrickCnt = 13

// RookColorCnt 色(スート)の数
const RookColorCnt = 4

// RookBirdDesign ルーク鳥(Rook bird)の仮想デザイン値。1〜4 は色、5 が鳥。
const RookBirdDesign = 5

// RookBirdValue ルーク鳥のカード値 (色札は 1〜14, 鳥は 0)
const RookBirdValue = 0

// Rook のビッド範囲 (点数目標)。最小70から最大120まで5刻み。
const (
	// RookMinBid 最小ビッド
	RookMinBid = 70
	// RookMaxBid 最大ビッド
	RookMaxBid = 120
	// RookBidStep ビッドの刻み幅
	RookBidStep = 5
)

// rookBirdRookPoints ルーク鳥の得点
const rookBirdRookPoints = 20

// RookPhase ゲームフェーズ
type RookPhase int

// Rook のフェーズ定数
const (
	// RookPhaseBid オークション(ビッド)フェーズ
	RookPhaseBid RookPhase = 0
	// RookPhaseNestExchange ネスト交換フェーズ (落札者が5枚受け取り5枚捨て、切り札色を宣言)
	RookPhaseNestExchange RookPhase = 1
	// RookPhasePlay トリックプレイフェーズ
	RookPhasePlay RookPhase = 2
	// RookPhaseTrickEnd トリック終了フェーズ
	RookPhaseTrickEnd RookPhase = 3
	// RookPhaseRoundEnd ラウンド終了フェーズ
	RookPhaseRoundEnd RookPhase = 4
	// RookPhaseGameEnd ゲーム終了フェーズ
	RookPhaseGameEnd RookPhase = 5
)

// RookHint ヒント情報
type RookHint struct {
	Bid            *int   // 推奨ビッド点数 (ビッドフェーズ)
	Pass           *bool  // パス推奨か
	DiscardIndices []int  // 推奨ディスカード5枚 (ネスト交換フェーズ)
	TrumpColor     *int   // 推奨切り札色 (ネスト交換フェーズ)
	CardIndex      *int   // 推奨カードインデックス (プレイフェーズ)
	Reason         string // ヒント理由キー
}

// Rook ルーク(Rook)ゲームクラス
type Rook struct {
	deck             []*Card
	deckDrawCnt      int
	players          []*RookPlayer
	config           RookConfig
	phase            RookPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	dealerIdx        int
	nest             []*Card
	nestPoints       int
	leadPlayerIdx    int
	// --- bidding state ---
	bidPlayerIdx  int
	passed        [RookPlayerCnt]bool
	highestBid    int // 0 = まだビッドなし
	highestBidder int
	// --- contract state ---
	contractBid int
	declarerIdx int
	trumpColor  int // 切り札色 (1..4, 未確定は -1)
	// --- scoring ---
	teamScores  [RookTeamCnt]int
	gameEndFlag bool
	winnerTeam  int // 勝利チーム (-1 = 未確定)
	actionLogBase
}

// NewRook コンストラクタ
func NewRook(players []*RookPlayer, config RookConfig) *Rook {
	return &Rook{
		players:       players,
		config:        config,
		winnerTeam:    -1,
		roundNumber:   0,
		dealerIdx:     0,
		trumpColor:    -1,
		highestBidder: -1,
		declarerIdx:   -1,
	}
}

// NewDefaultRook returns Rook with the standard 4-player partnership setup
// (human team 0, alternating CPU teams) and DefaultRookConfig.
// Single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultRook() *Rook {
	players := []*RookPlayer{
		NewRookPlayer(true, 0),
		NewRookPlayer(false, 1),
		NewRookPlayer(false, 0),
		NewRookPlayer(false, 1),
	}
	return NewRook(players, DefaultRookConfig())
}

// buildRookDeck builds the 57-card Rook deck directly: four colors (design
// 1..4) of ranks 1..14 plus one Rook bird (design RookBirdDesign, value 0).
func buildRookDeck() []*Card {
	deck := make([]*Card, 0, RookColorCnt*14+1)
	for color := 1; color <= RookColorCnt; color++ {
		for val := 1; val <= 14; val++ {
			deck = append(deck, NewCard(color, val, false))
		}
	}
	deck = append(deck, NewCard(RookBirdDesign, RookBirdValue, false))
	return deck
}

// Reset ゲーム初期化
func (g *Rook) Reset() {
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.teamScores = [RookTeamCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Rook) NextRound() {
	if g.phase != RookPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % RookPlayerCnt
	g.startRound()
}

// startRound ラウンドの状態を初期化して配り直す
func (g *Rook) startRound() {
	g.trickNumber = 0
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.nestPoints = 0
	g.passed = [RookPlayerCnt]bool{}
	g.highestBid = 0
	g.highestBidder = -1
	g.contractBid = 0
	g.declarerIdx = -1
	g.trumpColor = -1
	for _, p := range g.players {
		p.ResetRound()
	}
	g.dealRound()
	g.phase = RookPhaseBid
	g.bidPlayerIdx = (g.dealerIdx + 1) % RookPlayerCnt
}

// dealRound 13枚ずつ配り、残り5枚をネストにする
func (g *Rook) dealRound() {
	g.deck = buildRookDeck()
	rand.Shuffle(len(g.deck), func(i, j int) {
		g.deck[i], g.deck[j] = g.deck[j], g.deck[i]
	})
	g.deckDrawCnt = 0
	g.nest = nil
	for range RookHandSize {
		for j := range RookPlayerCnt {
			if card := g.drawCard(); card != nil {
				g.players[j].AddCard(card)
			}
		}
	}
	for {
		card := g.drawCard()
		if card == nil {
			break
		}
		g.nest = append(g.nest, card)
	}
	g.sortAllHands()
}

// drawCard デッキから1枚配る (尽きたら nil)
func (g *Rook) drawCard() *Card {
	return drawFromDeck(g.deck, &g.deckDrawCnt)
}

// --- Bidding ---

// PlayerBid 人間プレイヤーがビッドする
func (g *Rook) PlayerBid(bid int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != RookPhaseBid {
		return ErrWrongPhase
	}
	humanIdx := findHumanIdx(g.players)
	if humanIdx < 0 || g.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	if !g.validBid(bid) {
		return NewDomainError(ErrInvalidPlay, "無効なビッドです")
	}
	if bid <= g.highestBid {
		return NewDomainError(ErrInvalidPlay, "現在のビッドより高い必要があります")
	}
	g.applyBid(humanIdx, bid)
	return nil
}

// validBid ビッドが文法的に妥当か (範囲と刻み)
func (g *Rook) validBid(bid int) bool {
	return bid >= RookMinBid && bid <= RookMaxBid && bid%RookBidStep == 0
}

// PlayerPass 人間プレイヤーがパスする
func (g *Rook) PlayerPass() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != RookPhaseBid {
		return ErrWrongPhase
	}
	humanIdx := findHumanIdx(g.players)
	if humanIdx < 0 || g.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	g.applyPass(humanIdx)
	return nil
}

// CpuBid CPUプレイヤーが1ビッド実行する
func (g *Rook) CpuBid() {
	if g.gameEndFlag || g.phase != RookPhaseBid {
		return
	}
	if g.bidPlayerIdx < 0 || g.bidPlayerIdx >= RookPlayerCnt {
		return
	}
	if g.players[g.bidPlayerIdx].GetIsHuman() {
		return
	}
	if bid, ok := g.cpuSelectBid(g.bidPlayerIdx); ok {
		g.applyBid(g.bidPlayerIdx, bid)
	} else {
		g.applyPass(g.bidPlayerIdx)
	}
}

// applyBid ビッドを適用する
func (g *Rook) applyBid(idx, bid int) {
	g.players[idx].SetBid(bid)
	g.highestBid = bid
	g.highestBidder = idx
	g.appendLog(idx, "bid", fmt.Sprintf("%s bids %d", playerName(g.players, idx), bid), nil)
	g.advanceBid()
}

// applyPass パスを適用する
func (g *Rook) applyPass(idx int) {
	g.passed[idx] = true
	g.players[idx].SetPassed(true)
	g.appendLog(idx, "pass", fmt.Sprintf("%s passes", playerName(g.players, idx)), nil)
	g.advanceBid()
}

// advanceBid ビッドを次へ進め、終了判定を行う
func (g *Rook) advanceBid() {
	active := 0
	for i := range RookPlayerCnt {
		if !g.passed[i] {
			active++
		}
	}
	if g.highestBid > 0 && active <= 1 {
		g.finalizeBid()
		return
	}
	if g.highestBid == 0 && active == 0 {
		g.redeal()
		return
	}
	next := (g.bidPlayerIdx + 1) % RookPlayerCnt
	for g.passed[next] {
		next = (next + 1) % RookPlayerCnt
	}
	g.bidPlayerIdx = next
}

// redeal 全員パスした場合、同じラウンドを配り直す
func (g *Rook) redeal() {
	g.appendLog(-1, "redeal", "All players passed. Redealing.", nil)
	g.passed = [RookPlayerCnt]bool{}
	g.highestBid = 0
	g.highestBidder = -1
	for _, p := range g.players {
		p.ResetRound()
	}
	g.dealRound()
	g.phase = RookPhaseBid
	g.bidPlayerIdx = (g.dealerIdx + 1) % RookPlayerCnt
}

// finalizeBid 落札を確定し、落札者にネストを渡す
func (g *Rook) finalizeBid() {
	g.contractBid = g.highestBid
	g.declarerIdx = g.highestBidder
	g.players[g.declarerIdx].SetIsDeclarer(true)
	for _, c := range g.nest {
		g.players[g.declarerIdx].AddCard(c)
	}
	g.nest = nil
	g.appendLog(g.declarerIdx, "win_bid",
		fmt.Sprintf("%s wins the auction for %d", playerName(g.players, g.declarerIdx), g.contractBid), nil)
	g.sortAllHands()
	g.phase = RookPhaseNestExchange
	g.currentPlayerIdx = g.declarerIdx
}

// --- Nest exchange + trump declaration ---

// PlayerExchangeNest 人間(落札者)が5枚捨て、切り札色を宣言する
func (g *Rook) PlayerExchangeNest(discardIndices []int, trumpColor int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != RookPhaseNestExchange {
		return ErrWrongPhase
	}
	if g.declarerIdx < 0 || !g.players[g.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.doExchange(discardIndices, trumpColor)
}

// CpuExchange CPU(落札者)が5枚捨て、切り札色を宣言する
func (g *Rook) CpuExchange() {
	if g.gameEndFlag || g.phase != RookPhaseNestExchange {
		return
	}
	if g.declarerIdx < 0 || g.players[g.declarerIdx].GetIsHuman() {
		return
	}
	trump := g.cpuSelectTrump(g.declarerIdx)
	g.trumpColor = trump
	_ = g.doExchange(g.cpuSelectDiscards(g.declarerIdx, trump), trump)
}

// doExchange ネスト交換の共通処理。捨てた5枚をネストとして脇に置き、その得点を
// 最終トリックの勝者に加算するために記録する。
func (g *Rook) doExchange(discardIndices []int, trumpColor int) error {
	if trumpColor < 1 || trumpColor > RookColorCnt {
		return NewDomainError(ErrInvalidPlay, "切り札色は1〜4で指定してください")
	}
	player := g.players[g.declarerIdx]
	if len(discardIndices) != RookNestSize {
		return NewDomainError(ErrInvalidCard, "5枚捨ててください")
	}
	seen := make(map[int]bool, RookNestSize)
	for _, idx := range discardIndices {
		if idx < 0 || idx >= player.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidCard, "同じカードは選べません")
		}
		seen[idx] = true
	}
	discarded := player.RemoveCards(discardIndices)
	g.nest = discarded
	g.nestPoints = 0
	for _, c := range discarded {
		g.nestPoints += rookCardPoints(c)
	}
	g.trumpColor = trumpColor
	g.appendLog(g.declarerIdx, "exchange",
		fmt.Sprintf("%s discards %d cards, trump=%s", playerName(g.players, g.declarerIdx), len(discarded), rookColorName(trumpColor)), discarded)
	g.sortAllHands()
	g.startPlayPhase()
	return nil
}

// --- Play ---

// startPlayPhase プレイフェーズを開始する (落札者がリード)
func (g *Rook) startPlayPhase() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.leadPlayerIdx = g.declarerIdx
	g.currentPlayerIdx = g.declarerIdx
	g.phase = RookPhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Rook) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != RookPhasePlay {
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

// CpuPlay CPUプレイヤーが1ターン実行する
func (g *Rook) CpuPlay() {
	if g.gameEndFlag || g.phase != RookPhasePlay {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	idx := g.cpuSelectPlayCard(g.currentPlayerIdx)
	played := g.players[g.currentPlayerIdx].RemoveCard(idx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(g.currentPlayerIdx, played)
}

// playCard カードをプレイする共通処理
func (g *Rook) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), rookCardLabel(card)), []*Card{card})
	if len(g.currentTrick) == RookPlayerCnt {
		g.phase = RookPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % RookPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する。獲得した得点札を勝者チームへ、
// 最終トリックの勝者にはネストの得点も加算する。
func (g *Rook) ResolveTrick() {
	if g.phase != RookPhaseTrickEnd || len(g.currentTrick) != RookPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	cards := make([]*Card, len(g.currentTrick))
	trickPts := 0
	for i, tc := range g.currentTrick {
		cards[i] = tc.Card
		trickPts += rookCardPoints(tc.Card)
	}
	g.players[winnerIdx].AddTrick(cards)
	g.players[winnerIdx].AddPoints(trickPts)
	if g.trickNumber >= RookTrickCnt {
		g.players[winnerIdx].AddPoints(g.nestPoints)
		g.appendLog(winnerIdx, "trick_win",
			fmt.Sprintf("%s wins the last trick %d (+%d nest)", playerName(g.players, winnerIdx), g.trickNumber, g.nestPoints), cards)
		g.leadPlayerIdx = winnerIdx
		g.phase = RookPhaseRoundEnd
		return
	}
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (+%d)", playerName(g.players, winnerIdx), g.trickNumber, trickPts), cards)
	g.leadPlayerIdx = winnerIdx
	g.phase = RookPhaseTrickEnd
}

// NextTrick 次のトリックを開始する
func (g *Rook) NextTrick() {
	if g.phase != RookPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = RookPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (g *Rook) ScoreRound() {
	if g.phase != RookPhaseRoundEnd {
		return
	}
	declTeam := g.players[g.declarerIdx].GetTeam()
	defTeam := 1 - declTeam
	declPoints := g.teamPoints(declTeam)
	defPoints := g.teamPoints(defTeam)

	if declPoints >= g.contractBid {
		g.teamScores[declTeam] += declPoints
		g.appendLog(-1, "contract_made",
			fmt.Sprintf("Team %d makes the bid (%d/%d). +%d", declTeam, declPoints, g.contractBid, declPoints), nil)
	} else {
		g.teamScores[declTeam] -= g.contractBid
		g.appendLog(-1, "contract_failed",
			fmt.Sprintf("Team %d is set (%d/%d). -%d", declTeam, declPoints, g.contractBid, g.contractBid), nil)
	}
	g.teamScores[defTeam] += defPoints

	for ti := range RookTeamCnt {
		g.appendLog(-1, "team_score",
			fmt.Sprintf("Team %d: %d points", ti, g.teamScores[ti]), nil)
	}
	g.checkGameEnd(declTeam)
}

// checkGameEnd ゲーム終了判定。先に TargetScore に到達したチームが勝利。両者到達時は
// 高得点側 (同点は落札チーム) の勝ち。
func (g *Rook) checkGameEnd(declTeam int) {
	target := g.config.TargetScore
	other := 1 - declTeam
	declReached := g.teamScores[declTeam] >= target
	otherReached := g.teamScores[other] >= target
	if !declReached && !otherReached {
		return
	}
	switch {
	case declReached && (!otherReached || g.teamScores[declTeam] >= g.teamScores[other]):
		g.endGame(declTeam)
	default:
		g.endGame(other)
	}
}

// endGame ゲームを終了させる
func (g *Rook) endGame(team int) {
	g.gameEndFlag = true
	g.phase = RookPhaseGameEnd
	g.winnerTeam = team
	g.appendLog(-1, "game_end", fmt.Sprintf("Team %d wins the game!", team), nil)
}

// --- Card ranking / points ---

// isRookBird カードがルーク鳥かどうか
func (g *Rook) isRookBird(c *Card) bool {
	return c != nil && c.GetDesign() == RookBirdDesign
}

// rookRankStrength 色内でのカードの強さ (1が最高, 次いで14,13,...,2)
func rookRankStrength(value int) int {
	if value == 1 {
		return 15
	}
	return value
}

// rookCardPoints カードの得点を返す (5=5, 10=10, 14=10, 1=15, ルーク鳥=20, その他=0)
func rookCardPoints(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetDesign() == RookBirdDesign {
		return rookBirdRookPoints
	}
	switch c.GetValue() {
	case 5:
		return 5
	case 10, 14:
		return 10
	case 1:
		return 15
	}
	return 0
}

// effectiveSuit カードの実効スート(色)を返す。ルーク鳥は最高の切り札のため切り札色。
func (g *Rook) effectiveSuit(c *Card) int {
	if c == nil {
		return -1
	}
	if g.isRookBird(c) {
		return g.trumpColor
	}
	return c.GetDesign()
}

// cardRank トリック比較用のカードランクを返す (高い=強い)。
// ルーク鳥(1000) > 切り札(500+強さ) > 平札(100+強さ)。
func (g *Rook) cardRank(c *Card) int {
	if c == nil {
		return 0
	}
	if g.isRookBird(c) {
		return 1000
	}
	strength := rookRankStrength(c.GetValue())
	if g.trumpColor >= 1 && c.GetDesign() == g.trumpColor {
		return 500 + strength
	}
	return 100 + strength
}

// leadSuit 現在のトリックのリードスートを返す
func (g *Rook) leadSuit() int {
	if len(g.currentTrick) == 0 {
		return -1
	}
	return g.effectiveSuit(g.currentTrick[0].Card)
}

// trickScore リードスートを踏まえたトリック比較値を返す (0=勝てない)
func (g *Rook) trickScore(c *Card, leadSuit int) int {
	if g.isRookBird(c) {
		return 1000
	}
	es := c.GetDesign()
	if g.trumpColor >= 1 && es == g.trumpColor {
		return g.cardRank(c)
	}
	if es == leadSuit {
		return g.cardRank(c)
	}
	return 0
}

// trickWinner トリックの勝者を決定する
func (g *Rook) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	ls := g.leadSuit()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerScore := g.trickScore(g.currentTrick[0].Card, ls)
	for _, tc := range g.currentTrick[1:] {
		if s := g.trickScore(tc.Card, ls); s > winnerScore {
			winnerScore = s
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// validatePlay カードのプレイが有効か検証する。ルーク鳥はいつでもプレイ可能。
func (g *Rook) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil // リードは自由
	}
	if g.isRookBird(card) {
		return nil // ルーク鳥はいつでも出せる (ワイルド扱い)
	}
	ls := g.leadSuit()
	if g.effectiveSuit(card) != ls && g.playerHasSuit(playerIdx, ls) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// GetPlayableIndices はプレイフェーズで出せる手札のインデックス一覧を返す。
//
// **判定は validatePlay と同じ経路を通す。**別に書き写すと、片方だけ直したときに
// 「出せる」と示した札が拒否される (#4928)。プレイフェーズ以外では nil。
func (g *Rook) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	if g.phase != RookPhasePlay {
		return nil
	}
	p := g.players[playerIdx]
	out := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if g.validatePlay(playerIdx, p.GetCard(i)) == nil {
			out = append(out, i)
		}
	}
	return out
}

// playerHasSuit プレイヤーが実効スートのカードを持っているか (ルーク鳥は除く)
func (g *Rook) playerHasSuit(playerIdx, suit int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if g.isRookBird(c) {
			continue
		}
		if g.effectiveSuit(c) == suit {
			return true
		}
	}
	return false
}

// --- CPU AI ---

// cpuSelectBid CPUのビッド選択 (ok=false でパス)。
func (g *Rook) cpuSelectBid(playerIdx int) (int, bool) {
	est := g.evalHand(playerIdx)
	threshold := RookMinBid
	switch g.config.CpuDifficulty {
	case RookCpuDifficultyEasy:
		threshold = RookMinBid + RookBidStep*2
	case RookCpuDifficultyHard:
		threshold = RookMinBid - RookBidStep
	}
	if est < threshold {
		return 0, false
	}
	// 現在の最高ビッドを1刻み上回る。見積もりを超えないよう上限を掛ける。
	target := est
	if target > RookMaxBid {
		target = RookMaxBid
	}
	bid := g.highestBid + RookBidStep
	if g.highestBid == 0 {
		bid = RookMinBid
	}
	if bid > target || bid > RookMaxBid {
		return 0, false
	}
	return bid, true
}

// evalHand おおよその獲得可能点数を見積もる。得点札・高札・切り札候補の枚数から算出。
func (g *Rook) evalHand(playerIdx int) int {
	p := g.players[playerIdx]
	colorCount := map[int]int{}
	pointSum := 0
	highCards := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		pointSum += rookCardPoints(c)
		if g.isRookBird(c) {
			highCards += 2
			continue
		}
		colorCount[c.GetDesign()]++
		if c.GetValue() == 1 || c.GetValue() == 14 {
			highCards++
		}
	}
	longest := 0
	for _, n := range colorCount {
		if n > longest {
			longest = n
		}
	}
	// 基準70に、得点札・高札・長い色のボーナスを加える。
	return RookMinBid + pointSum/4 + highCards*RookBidStep + max(0, longest-4)*RookBidStep
}

// cpuSelectTrump CPU(落札者)が宣言する切り札色を選ぶ (最も枚数が多く得点の高い色)
func (g *Rook) cpuSelectTrump(playerIdx int) int {
	p := g.players[playerIdx]
	score := map[int]int{}
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if g.isRookBird(c) {
			continue
		}
		score[c.GetDesign()] += 2 + rookRankStrength(c.GetValue())/5
	}
	best := 1
	bestScore := -1
	for color := 1; color <= RookColorCnt; color++ {
		if score[color] > bestScore {
			bestScore = score[color]
			best = color
		}
	}
	return best
}

// cpuSelectDiscards CPU(落札者)が捨てる5枚のインデックスを選ぶ。
// 得点札・切り札・ルーク鳥を温存し、価値の低い札から捨てる。
func (g *Rook) cpuSelectDiscards(playerIdx int, trumpColor int) []int {
	p := g.players[playerIdx]
	n := p.GetCardsSize()
	idxs := make([]int, n)
	for i := range idxs {
		idxs[i] = i
	}
	// trumpColor は宣言予定/推奨の切り札色を渡す。g.trumpColor はネスト交換時点で
	// まだ -1（未宣言）のことがあり、それを使うと切り札を安く評価して捨ててしまう。
	keepValue := func(c *Card) int {
		v := rookCardPoints(c) * 10
		if g.isRookBird(c) {
			v += 1000
		}
		if trumpColor >= 1 && c.GetDesign() == trumpColor {
			v += 200
		}
		return v + rookRankStrength(c.GetValue())
	}
	sort.SliceStable(idxs, func(a, b int) bool {
		return keepValue(p.GetCard(idxs[a])) < keepValue(p.GetCard(idxs[b]))
	})
	count := RookNestSize
	if count > n {
		count = n
	}
	return append([]int(nil), idxs[:count]...)
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選ぶ
func (g *Rook) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	p := g.players[playerIdx]
	// リード: 最も強いカード
	if len(g.currentTrick) == 0 {
		bestIdx := valid[0]
		for _, idx := range valid[1:] {
			if g.cardRank(p.GetCard(idx)) > g.cardRank(p.GetCard(bestIdx)) {
				bestIdx = idx
			}
		}
		return bestIdx
	}
	ls := g.leadSuit()
	winScore := g.currentWinnerScore(ls)
	winnerTeam := g.players[g.trickWinner()].GetTeam()
	myTeam := p.GetTeam()
	// パートナーが勝っていれば得点札を渡す (最も得点の高い有効札)
	if winnerTeam == myTeam {
		return g.highestPointValid(playerIdx, valid)
	}
	// 相手が勝っている: 勝てる最弱札で取りに行く、無理なら最も安い札
	over := []int{}
	for _, idx := range valid {
		if g.trickScore(p.GetCard(idx), ls) > winScore {
			over = append(over, idx)
		}
	}
	if len(over) > 0 {
		bestIdx := over[0]
		for _, idx := range over[1:] {
			if g.cardRank(p.GetCard(idx)) < g.cardRank(p.GetCard(bestIdx)) {
				bestIdx = idx
			}
		}
		return bestIdx
	}
	return g.weakestValid(playerIdx, valid)
}

// weakestValid 有効札のうち最弱 (得点も低い) のインデックスを返す
func (g *Rook) weakestValid(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	bestIdx := valid[0]
	bestKey := rookCardPoints(p.GetCard(bestIdx))*100 + g.cardRank(p.GetCard(bestIdx))
	for _, idx := range valid[1:] {
		key := rookCardPoints(p.GetCard(idx))*100 + g.cardRank(p.GetCard(idx))
		if key < bestKey {
			bestKey = key
			bestIdx = idx
		}
	}
	return bestIdx
}

// highestPointValid 有効札のうち最も得点の高い (同点は最弱ランク) インデックスを返す
func (g *Rook) highestPointValid(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	bestIdx := valid[0]
	bestPts := rookCardPoints(p.GetCard(bestIdx))
	bestRank := g.cardRank(p.GetCard(bestIdx))
	for _, idx := range valid[1:] {
		pts := rookCardPoints(p.GetCard(idx))
		rank := g.cardRank(p.GetCard(idx))
		if pts > bestPts || (pts == bestPts && rank < bestRank) {
			bestPts = pts
			bestRank = rank
			bestIdx = idx
		}
	}
	return bestIdx
}

// currentWinnerScore 現在のトリックでの暫定最強トリック値を返す
func (g *Rook) currentWinnerScore(ls int) int {
	best := 0
	for _, tc := range g.currentTrick {
		if s := g.trickScore(tc.Card, ls); s > best {
			best = s
		}
	}
	return best
}

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す (Web用)
func (g *Rook) GetValidPlayIndices(playerIdx int) []int {
	return g.getValidPlayIndices(playerIdx)
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (g *Rook) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// --- Hint ---

// GetHint 現在の人間の手番に対するヒントを返す
func (g *Rook) GetHint() *RookHint {
	humanIdx := findHumanIdx(g.players)
	if humanIdx < 0 || g.gameEndFlag {
		return nil
	}
	switch g.phase {
	case RookPhaseBid:
		if g.bidPlayerIdx != humanIdx {
			return nil
		}
		if bid, ok := g.cpuSelectBid(humanIdx); ok {
			b := bid
			return &RookHint{Bid: &b, Reason: "strategic_bid"}
		}
		pass := true
		return &RookHint{Pass: &pass, Reason: "pass_recommended"}
	case RookPhaseNestExchange:
		if g.declarerIdx != humanIdx {
			return nil
		}
		trump := g.cpuSelectTrump(humanIdx)
		return &RookHint{DiscardIndices: g.cpuSelectDiscards(humanIdx, trump), TrumpColor: &trump, Reason: "discard_weakest"}
	case RookPhasePlay:
		if g.currentPlayerIdx != humanIdx {
			return nil
		}
		valid := g.getValidPlayIndices(humanIdx)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuSelectPlayCard(humanIdx)
		return &RookHint{CardIndex: &idx, Reason: g.playHintReason(humanIdx, idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由キーを判定する
func (g *Rook) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	if len(g.currentTrick) == 0 {
		if g.trumpColor >= 1 && g.effectiveSuit(card) == g.trumpColor {
			return "lead_trump"
		}
		return "lead_strong"
	}
	if g.effectiveSuit(card) == g.leadSuit() {
		return "follow_suit"
	}
	if g.trumpColor >= 1 && g.effectiveSuit(card) == g.trumpColor {
		return "trump_cut"
	}
	return "discard_weak"
}

// --- State getters / setters ---

// GetPhase 現在のフェーズ取得
func (g *Rook) GetPhase() RookPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Rook) SetPhase(phase RookPhase) { g.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (g *Rook) GetRoundNumber() int { return g.roundNumber }

// GetTrickNumber 現在のトリック番号取得
func (g *Rook) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Rook) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Rook) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Rook) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Rook) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Rook) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Rook) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam 勝利チーム取得 (-1 = 未確定)
func (g *Rook) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt プレイヤー数取得
func (g *Rook) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Rook) GetPlayer(i int) *RookPlayer {
	return getPlayer(g.players, i)
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Rook) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Rook) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッド手番インデックス取得
func (g *Rook) GetBidPlayerIdx() int { return g.bidPlayerIdx }

// SetBidPlayerIdx ビッド手番インデックス設定 (テスト用)
func (g *Rook) SetBidPlayerIdx(idx int) { g.bidPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Rook) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (g *Rook) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetTrumpColor 切り札色取得 (-1 = なし)
func (g *Rook) GetTrumpColor() int { return g.trumpColor }

// SetTrumpColor 切り札色設定 (テスト用)
func (g *Rook) SetTrumpColor(color int) { g.trumpColor = color }

// GetContractBid 落札ビッド取得
func (g *Rook) GetContractBid() int { return g.contractBid }

// SetContractBid 落札ビッド設定 (テスト用)
func (g *Rook) SetContractBid(bid int) { g.contractBid = bid }

// GetDeclarerIdx 落札者インデックス取得 (-1 = 未確定)
func (g *Rook) GetDeclarerIdx() int { return g.declarerIdx }

// SetDeclarerIdx 落札者インデックス設定 (テスト用)
func (g *Rook) SetDeclarerIdx(idx int) { g.declarerIdx = idx }

// GetHighestBid 現在の最高ビッド取得 (0 = なし)
func (g *Rook) GetHighestBid() int { return g.highestBid }

// SetHighestBid 最高ビッド設定 (テスト用)
func (g *Rook) SetHighestBid(bid int) { g.highestBid = bid }

// GetHighestBidder 最高ビッダーのインデックス取得 (-1 = なし)
func (g *Rook) GetHighestBidder() int { return g.highestBidder }

// GetNest ネスト取得
func (g *Rook) GetNest() []*Card { return g.nest }

// SetNest ネスト設定 (テスト用)
func (g *Rook) SetNest(nest []*Card) { g.nest = nest }

// GetNestPoints ネストの得点取得
func (g *Rook) GetNestPoints() int { return g.nestPoints }

// GetTeamScore チームスコア取得
func (g *Rook) GetTeamScore(team int) int {
	if team < 0 || team >= RookTeamCnt {
		return 0
	}
	return g.teamScores[team]
}

// SetTeamScore チームスコア設定 (テスト用)
func (g *Rook) SetTeamScore(team, score int) {
	if team >= 0 && team < RookTeamCnt {
		g.teamScores[team] = score
	}
}

// teamPoints チームがこのラウンドで獲得した得点札の合計
func (g *Rook) teamPoints(team int) int {
	sum := 0
	for _, p := range g.players {
		if p.GetTeam() == team {
			sum += p.GetPoints()
		}
	}
	return sum
}

// GetTeamPoints チームがこのラウンドで獲得した得点札の合計を取得
func (g *Rook) GetTeamPoints(team int) int {
	if team < 0 || team >= RookTeamCnt {
		return 0
	}
	return g.teamPoints(team)
}

// IsHumanTurn 現在の手番が人間かどうか
func (g *Rook) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (g *Rook) IsHumanBidTurn() bool {
	return isHumanTurn(g.players, g.bidPlayerIdx)
}

// GetConfig 設定取得
func (g *Rook) GetConfig() RookConfig { return g.config }

// SetConfig 設定変更
func (g *Rook) SetConfig(cfg RookConfig) { g.config = cfg }

// CardRankPublic カードランク取得 (テスト用)
func (g *Rook) CardRankPublic(card *Card) int { return g.cardRank(card) }

// EffectiveSuitPublic 実効スート取得 (テスト用)
func (g *Rook) EffectiveSuitPublic(card *Card) int { return g.effectiveSuit(card) }

// CardPointsPublic カード得点取得 (テスト用)
func (g *Rook) CardPointsPublic(card *Card) int { return rookCardPoints(card) }

// --- Private helpers ---

// sortAllHands 全プレイヤーの手札をソートする
func (g *Rook) sortAllHands() {
	for _, p := range g.players {
		g.sortHand(p)
	}
}

// sortHand プレイヤーの手札を色→強さ順にソートする
func (g *Rook) sortHand(p *RookPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		di := cards[i].GetDesign()
		dj := cards[j].GetDesign()
		if di != dj {
			return di < dj
		}
		return rookRankStrength(cards[i].GetValue()) < rookRankStrength(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// rookColorName 色番号を英字ラベルにする (ログ用)
func rookColorName(color int) string {
	switch color {
	case 1:
		return "Red"
	case 2:
		return "Yellow"
	case 3:
		return "Green"
	case 4:
		return "Black"
	}
	return "?"
}

// rookCardLabel カードのログ表示文字列 (ルーク鳥対応)
func rookCardLabel(c *Card) string {
	if c == nil {
		return "??"
	}
	if c.GetDesign() == RookBirdDesign {
		return "Rook"
	}
	return fmt.Sprintf("%s%d", rookColorName(c.GetDesign()), c.GetValue())
}

// --- JSON ---

// rookJSON is the JSON wire format for Rook.
type rookJSON struct {
	Deck             []*Card             `json:"dk"`
	DeckDrawCnt      int                 `json:"dw"`
	Players          []*RookPlayer       `json:"ps"`
	Config           RookConfig          `json:"cf"`
	Phase            RookPhase           `json:"ph"`
	RoundNumber      int                 `json:"rn"`
	TrickNumber      int                 `json:"tn"`
	CurrentPlayerIdx int                 `json:"ci"`
	CurrentTrick     []*TrickCard        `json:"ct"`
	DealerIdx        int                 `json:"di"`
	Nest             []*Card             `json:"nt"`
	NestPoints       int                 `json:"np"`
	LeadPlayerIdx    int                 `json:"li"`
	BidPlayerIdx     int                 `json:"bi"`
	Passed           [RookPlayerCnt]bool `json:"pd"`
	HighestBid       int                 `json:"hb"`
	HighestBidder    int                 `json:"hr"`
	ContractBid      int                 `json:"cn"`
	DeclarerIdx      int                 `json:"dc"`
	TrumpColor       int                 `json:"tp"`
	TeamScores       [RookTeamCnt]int    `json:"sc"`
	GameEndFlag      bool                `json:"ge"`
	WinnerTeam       int                 `json:"wt"`
	ActionLog        []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Rook) MarshalJSON() ([]byte, error) {
	return json.Marshal(rookJSON{
		Deck:             g.deck,
		DeckDrawCnt:      g.deckDrawCnt,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		DealerIdx:        g.dealerIdx,
		Nest:             g.nest,
		NestPoints:       g.nestPoints,
		LeadPlayerIdx:    g.leadPlayerIdx,
		BidPlayerIdx:     g.bidPlayerIdx,
		Passed:           g.passed,
		HighestBid:       g.highestBid,
		HighestBidder:    g.highestBidder,
		ContractBid:      g.contractBid,
		DeclarerIdx:      g.declarerIdx,
		TrumpColor:       g.trumpColor,
		TeamScores:       g.teamScores,
		GameEndFlag:      g.gameEndFlag,
		WinnerTeam:       g.winnerTeam,
		ActionLog:        g.actionLog,
	})
}

// rookMaxSliceLen caps slice sizes during deserialisation.
const rookMaxSliceLen = 1000

// rookValidCard はデシリアライズ時のカード妥当性を検証する (nil拒否, 値域チェック)。
func rookValidCard(c *Card) bool {
	if c == nil {
		return false
	}
	d := c.GetDesign()
	v := c.GetValue()
	if d == RookBirdDesign {
		return v == RookBirdValue
	}
	return d >= 1 && d <= RookColorCnt && v >= 1 && v <= 14
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Rook) UnmarshalJSON(data []byte) error {
	var j rookJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > rookMaxSliceLen || len(j.CurrentTrick) > rookMaxSliceLen ||
		len(j.Nest) > rookMaxSliceLen || len(j.Deck) > rookMaxSliceLen ||
		len(j.ActionLog) > rookMaxSliceLen {
		return fmt.Errorf("rook: input array exceeds maximum allowed size")
	}
	g.players = j.Players
	if len(g.players) != RookPlayerCnt {
		return fmt.Errorf("rook: invalid player count: %d", len(g.players))
	}
	for _, p := range g.players {
		if p == nil {
			return fmt.Errorf("rook: player is nil")
		}
		if p.GetTeam() < 0 || p.GetTeam() >= RookTeamCnt {
			return fmt.Errorf("rook: invalid team index: %d", p.GetTeam())
		}
	}
	for _, c := range j.Deck {
		if !rookValidCard(c) {
			return fmt.Errorf("rook: invalid deck card")
		}
	}
	for _, c := range j.Nest {
		if !rookValidCard(c) {
			return fmt.Errorf("rook: invalid nest card")
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || !rookValidCard(tc.Card) {
			return fmt.Errorf("rook: invalid trick card")
		}
		if tc.PlayerIdx < 0 || tc.PlayerIdx >= RookPlayerCnt {
			return fmt.Errorf("rook: invalid trick player index: %d", tc.PlayerIdx)
		}
	}
	g.config = j.Config
	if err := g.config.Validate(); err != nil {
		return fmt.Errorf("rook: invalid config: %w", err)
	}
	g.deck = j.Deck
	if g.deck == nil {
		g.deck = make([]*Card, 0)
	}
	g.deckDrawCnt = j.DeckDrawCnt
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.dealerIdx = j.DealerIdx
	g.nest = j.Nest
	if g.nest == nil {
		g.nest = make([]*Card, 0)
	}
	g.nestPoints = j.NestPoints
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.bidPlayerIdx = j.BidPlayerIdx
	g.passed = j.Passed
	g.highestBid = j.HighestBid
	g.highestBidder = j.HighestBidder
	g.contractBid = j.ContractBid
	g.declarerIdx = j.DeclarerIdx
	g.trumpColor = j.TrumpColor
	g.teamScores = j.TeamScores
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
