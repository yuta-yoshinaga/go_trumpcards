//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// HoneymoonBridgePhase はハネムーンブリッジのゲームフェーズ。
type HoneymoonBridgePhase int

// HoneymoonBridge のフェーズ定数
const (
	// HoneymoonBridgePhaseDraw 引き合い（13 トリック打ちながら山札を分け合う）
	HoneymoonBridgePhaseDraw HoneymoonBridgePhase = iota
	// HoneymoonBridgePhaseBid 競り
	HoneymoonBridgePhaseBid
	// HoneymoonBridgePhasePlay 本番のプレイ
	HoneymoonBridgePhasePlay
	// HoneymoonBridgePhaseRoundEnd ディール終了
	HoneymoonBridgePhaseRoundEnd
	// HoneymoonBridgePhaseGameEnd ゲーム終了
	HoneymoonBridgePhaseGameEnd
)

// HoneymoonBridgePlayerCnt はプレイヤー数（2 人固定）。
const HoneymoonBridgePlayerCnt = 2

// HoneymoonBridgeHandSize は各プレイヤーの手札枚数。
const HoneymoonBridgeHandSize = 13

// HoneymoonBridgeTricksPerPhase は 1 フェーズのトリック数。
const HoneymoonBridgeTricksPerPhase = HoneymoonBridgeHandSize

// HoneymoonBridgeStockSize は引き合いに使う山札の枚数。
//
// **2 人 × 13 枚 = 26 枚配ると、残りもちょうど 26 枚。** 引き合いフェーズは
// 13 トリック打ち、各トリックのあと**2 人が 1 枚ずつ引く**ので 13 × 2 = 26 で
// 山札を使い切り、両者とも再び 13 枚に戻ります。
const HoneymoonBridgeStockSize = 52 - HoneymoonBridgePlayerCnt*HoneymoonBridgeHandSize

// HoneymoonBridgeMaxLevel は競りの上限レベル（7 = 全 13 トリック）。
const HoneymoonBridgeMaxLevel = 7

// HoneymoonBridgeBookTricks は契約に含まれない「ブック」のトリック数。
//
// **レベル n の契約は 6 + n トリック必要。** ブリッジと同じ数え方です。
const HoneymoonBridgeBookTricks = 6

// HoneymoonBridgeDefaultTarget は既定の目標点。
const HoneymoonBridgeDefaultTarget = 100

// honeymoonBridgeMaxSliceLen caps slice sizes during deserialisation.
const honeymoonBridgeMaxSliceLen = 1000

// HoneymoonBridge はハネムーンブリッジのゲームクラス。
//
// コントラクトブリッジを 2 人用に翻案したもの。52 枚を 13 枚ずつ配り、残りの
// **26 枚を山札**にします。
//
// **前半は「引き合い」。** 13 トリックを切り札なしで打ち、各トリックのあと
// 勝者→敗者の順に山札から 1 枚ずつ引きます。13 × 2 = 26 で山札はちょうど
// 尽き、両者とも再び 13 枚になります。**このフェーズのトリックは得点になりま
// せん**——何を引けたか、相手が何を取ったかを見るための段です。
//
// **後半がブリッジ。** 手札が確定してから競り、契約を決めて 13 トリックを
// 打ちます。
type HoneymoonBridge struct {
	players     []*HoneymoonBridgePlayer
	config      HoneymoonBridgeConfig
	phase       HoneymoonBridgePhase
	trumpCards  *TrumpCards
	stock       []*Card
	roundNumber int
	trickNumber int
	// trumpSuit は契約のスート（0 = ノートランプ、競り前も 0）。
	trumpSuit int
	// declarerIdx は落札者 (-1 = 競り中)。
	declarerIdx int
	// contractLevel は落札レベル（0 = 未確定）。
	contractLevel int
	// passCount は連続パス数。**2 回続いたら競りが締まる。**
	passCount        int
	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	dealerIdx        int
	// lastMade は直前のディールで契約が成立したか。
	lastMade bool
	// lastTricks は直前のディールで落札者が取ったトリック数。
	lastTricks int
	// lastPoints は直前のディールで動いた点数。
	//
	// **得点式は契約レベル×10 + オーバートリック×5 / 失敗は不足×10** と細かい
	// のに、画面はトリックの過不足しか出しておらず、実際に何点動いたのかは
	// 累計の差を自分で引くしかなかった (#5760)。
	lastPoints  int
	gameEndFlag bool
	winnerIdx   int
	actionLogBase
}

// NewHoneymoonBridge はコンストラクタ。
//
// **2 人ちょうどでなければ標準のセットアップに差し替える（レビュー指摘 PR #5312）。**
// 席数はゲーム中どこでも固定の前提で、足りないまま配ると `startRound` が
// 範囲外を引く。いまは `NewDefaultHoneymoonBridge` からしか呼ばれないが、
// 呼び手が増えたときに黙って壊れる形にはしない。
func NewHoneymoonBridge(players []*HoneymoonBridgePlayer, config HoneymoonBridgeConfig) *HoneymoonBridge {
	if len(players) != HoneymoonBridgePlayerCnt {
		players = []*HoneymoonBridgePlayer{
			NewHoneymoonBridgePlayer(true),
			NewHoneymoonBridgePlayer(false),
		}
	}
	return &HoneymoonBridge{
		players:     players,
		config:      config,
		declarerIdx: -1,
		winnerIdx:   -1,
	}
}

// NewDefaultHoneymoonBridge は標準の 2 人セットアップを返す。
func NewDefaultHoneymoonBridge() *HoneymoonBridge {
	players := []*HoneymoonBridgePlayer{
		NewHoneymoonBridgePlayer(true),
		NewHoneymoonBridgePlayer(false),
	}
	return NewHoneymoonBridge(players, DefaultHoneymoonBridgeConfig())
}

// honeymoonBridgeRank は札の強さ。**A が最強。**
func honeymoonBridgeRank(c *Card) int {
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// Reset はゲームを初期化する。
func (h *HoneymoonBridge) Reset() {
	h.roundNumber = 0
	h.dealerIdx = 0
	h.gameEndFlag = false
	h.winnerIdx = -1
	h.lastMade = false
	h.lastTricks = 0
	h.lastPoints = 0
	h.actionLog = nil
	for _, p := range h.players {
		p.ResetGame()
	}
	h.startRound()
}

// startRound は 1 ディールを配り直す。
func (h *HoneymoonBridge) startRound() {
	h.phase = HoneymoonBridgePhaseDraw
	h.trumpSuit = 0
	h.trickNumber = 0
	h.currentTrick = nil
	h.declarerIdx = -1
	h.contractLevel = 0
	h.passCount = 0
	for _, p := range h.players {
		p.ResetRound()
	}

	h.trumpCards = NewTrumpCards(0)
	h.trumpCards.Shuffle()
	for range HoneymoonBridgeHandSize {
		for i := range HoneymoonBridgePlayerCnt {
			idx := (h.dealerIdx + 1 + i) % HoneymoonBridgePlayerCnt
			if c := h.trumpCards.DrawCard(); c != nil {
				h.players[idx].AddCard(c)
			}
		}
	}
	// **残りは山札。** 引き合いで 1 枚ずつ分け合います。
	h.stock = nil
	for range HoneymoonBridgeStockSize {
		if c := h.trumpCards.DrawCard(); c != nil {
			h.stock = append(h.stock, c)
		}
	}
	h.sortAllHands()

	h.roundNumber++
	h.leadPlayerIdx = (h.dealerIdx + 1) % HoneymoonBridgePlayerCnt
	h.currentPlayerIdx = h.leadPlayerIdx
	h.addLog(-1, "deal", fmt.Sprintf("ディール %d：13 枚ずつ配り、山札 %d 枚",
		h.roundNumber, len(h.stock)), nil)
}

// sortAllHands は手札をスート・ランク順に整える。
func (h *HoneymoonBridge) sortAllHands() {
	for _, p := range h.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return honeymoonBridgeRank(ci) < honeymoonBridgeRank(cj)
		})
	}
}

// GetValidPlayIndices は playerIdx が出せる手札の添字を返す。
//
// **どちらのフェーズもフォロー義務あり。**
func (h *HoneymoonBridge) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= HoneymoonBridgePlayerCnt || h.gameEndFlag {
		return []int{}
	}
	if h.phase != HoneymoonBridgePhaseDraw && h.phase != HoneymoonBridgePhasePlay {
		return []int{}
	}
	p := h.players[playerIdx]
	leadSuit := 0
	if len(h.currentTrick) > 0 {
		leadSuit = h.currentTrick[0].Card.GetDesign()
	}
	all := make([]int, 0, p.GetCardsSize())
	follow := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		all = append(all, i)
		if leadSuit != 0 && p.GetCard(i).GetDesign() == leadSuit {
			follow = append(follow, i)
		}
	}
	if len(follow) > 0 {
		return follow
	}
	return all
}

// IsHumanTurn は現在の手番が人間かを返す。
func (h *HoneymoonBridge) IsHumanTurn() bool {
	if h.gameEndFlag {
		return false
	}
	if h.phase != HoneymoonBridgePhaseDraw && h.phase != HoneymoonBridgePhasePlay {
		return false
	}
	return h.players[h.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn は人間が宣言する番かを返す。
func (h *HoneymoonBridge) IsHumanBidTurn() bool {
	return !h.gameEndFlag && h.phase == HoneymoonBridgePhaseBid && h.currentPlayerIdx == 0
}

// PlayerPlay は人間が 1 枚出す。
func (h *HoneymoonBridge) PlayerPlay(cardIndex int) error {
	if !h.IsHumanTurn() {
		return errors.New("not your turn")
	}
	return h.play(0, cardIndex)
}

// CpuPlay は CPU が 1 枚出す。
func (h *HoneymoonBridge) CpuPlay() {
	if h.gameEndFlag || h.IsHumanTurn() {
		return
	}
	if h.phase != HoneymoonBridgePhaseDraw && h.phase != HoneymoonBridgePhasePlay {
		return
	}
	_ = h.play(h.currentPlayerIdx, h.chooseCpuCard(h.currentPlayerIdx))
}

// play は playerIdx に手札の cardIndex を出させる。
func (h *HoneymoonBridge) play(playerIdx, cardIndex int) error {
	if h.gameEndFlag {
		return errors.New("game is over")
	}
	if h.phase != HoneymoonBridgePhaseDraw && h.phase != HoneymoonBridgePhasePlay {
		return errors.New("not a play phase")
	}
	if playerIdx != h.currentPlayerIdx {
		return fmt.Errorf("not player %d's turn", playerIdx)
	}
	p := h.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	if !honeymoonBridgeContains(h.GetValidPlayIndices(playerIdx), cardIndex) {
		return errors.New("must follow the led suit")
	}

	card := p.RemoveCard(cardIndex)
	h.currentTrick = append(h.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	h.addLog(playerIdx, "play", cardStr(card), []*Card{card})

	if len(h.currentTrick) < HoneymoonBridgePlayerCnt {
		h.currentPlayerIdx = (h.currentPlayerIdx + 1) % HoneymoonBridgePlayerCnt
		return nil
	}
	h.resolveTrick()
	return nil
}

// resolveTrick はトリックを解決する。
func (h *HoneymoonBridge) resolveTrick() {
	winner := h.trickWinner()
	cards := make([]*Card, 0, HoneymoonBridgePlayerCnt)
	for _, tc := range h.currentTrick {
		cards = append(cards, tc.Card)
	}
	// **引き合いのトリックは得点にならない。** 数えるのは本番だけ。
	if h.phase == HoneymoonBridgePhasePlay {
		h.players[winner].AddTrick(cards)
	}
	h.currentTrick = nil
	h.trickNumber++
	h.leadPlayerIdx = winner
	h.currentPlayerIdx = winner
	h.addLog(winner, "trick", fmt.Sprintf("トリック %d を取りました", h.trickNumber), nil)

	if h.phase == HoneymoonBridgePhaseDraw {
		// **勝った人が先に引く。** 山札は必ず 2 枚ずつ減る。
		h.drawAfterTrick(winner)
		if h.trickNumber >= HoneymoonBridgeTricksPerPhase {
			h.startBidding()
		}
		return
	}
	if h.trickNumber >= HoneymoonBridgeTricksPerPhase {
		h.finishRound()
	}
}

// drawAfterTrick は勝者→敗者の順に山札から 1 枚ずつ引かせる。
func (h *HoneymoonBridge) drawAfterTrick(winner int) {
	for i := range HoneymoonBridgePlayerCnt {
		idx := (winner + i) % HoneymoonBridgePlayerCnt
		if len(h.stock) == 0 {
			break
		}
		h.players[idx].AddCard(h.stock[0])
		h.stock = h.stock[1:]
	}
	h.sortAllHands()
}

// trickWinner は切り札 > リードスートの順で最強札を出した人を返す。
//
// **引き合いフェーズは切り札なし。** trumpSuit が 0 なのでリードスートだけで決まる。
func (h *HoneymoonBridge) trickWinner() int {
	if len(h.currentTrick) == 0 {
		return h.leadPlayerIdx
	}
	leadSuit := h.currentTrick[0].Card.GetDesign()
	best, bestRank, bestTrump := h.currentTrick[0].PlayerIdx, -1, false
	for _, tc := range h.currentTrick {
		suit, rank := tc.Card.GetDesign(), honeymoonBridgeRank(tc.Card)
		isTrump := h.trumpSuit != 0 && suit == h.trumpSuit
		switch {
		case isTrump && !bestTrump:
			best, bestRank, bestTrump = tc.PlayerIdx, rank, true
		case isTrump == bestTrump && suit == leadSuit && !bestTrump && rank > bestRank:
			best, bestRank = tc.PlayerIdx, rank
		case isTrump && bestTrump && rank > bestRank:
			best, bestRank = tc.PlayerIdx, rank
		}
	}
	return best
}

// --- 競り -------------------------------------------------------------------

// startBidding は引き合いを終えて競りに入る。
func (h *HoneymoonBridge) startBidding() {
	h.phase = HoneymoonBridgePhaseBid
	h.trickNumber = 0
	h.passCount = 0
	// **競りは親の左隣から。**
	h.currentPlayerIdx = (h.dealerIdx + 1) % HoneymoonBridgePlayerCnt
	h.addLog(-1, "draw", "引き合い終了。両者 13 枚で競りに入ります", nil)
}

// NextBid は次に出せる最小の宣言（レベル, スート）を返す。
//
// **同じレベルならスートが上でなければ通らない。** ノートランプが最強。
func (h *HoneymoonBridge) NextBid() (int, int) {
	if h.contractLevel == 0 {
		return 1, CardDesignSpade
	}
	if h.trumpSuit == 0 {
		// ノートランプが立っているので、次はレベルを上げるしかない。
		if h.contractLevel >= HoneymoonBridgeMaxLevel {
			return 0, 0
		}
		return h.contractLevel + 1, CardDesignSpade
	}
	if h.trumpSuit < CardDesignDiamond {
		return h.contractLevel, h.trumpSuit + 1
	}
	// ♦ の次はノートランプ。
	return h.contractLevel, 0
}

// honeymoonBridgeSuitRank は競りのスート序列を返す。
//
// **CardDesign の並びとは違う。** 定数は ♠1 ♣2 ♥3 ♦4 だが、競りでは
// ♠ < ♣ < ♥ < ♦ < NT の順に強い（NT を最強に置くための独自序列）。
func honeymoonBridgeSuitRank(suit int) int {
	if suit == 0 {
		return 5 // ノートランプが最強
	}
	return suit
}

// outbids は (level, suit) が現在の契約を上回るかを返す。
func (h *HoneymoonBridge) outbids(level, suit int) bool {
	if h.contractLevel == 0 {
		return true
	}
	if level != h.contractLevel {
		return level > h.contractLevel
	}
	return honeymoonBridgeSuitRank(suit) > honeymoonBridgeSuitRank(h.trumpSuit)
}

// PlayerBid は人間が宣言する。
func (h *HoneymoonBridge) PlayerBid(level, suit int) error {
	if !h.IsHumanBidTurn() {
		return errors.New("not your turn to bid")
	}
	return h.bidBy(0, level, suit)
}

// PlayerPass は人間が降りる。
func (h *HoneymoonBridge) PlayerPass() error {
	if !h.IsHumanBidTurn() {
		return errors.New("not your turn to bid")
	}
	return h.bidBy(0, 0, 0)
}

// CpuBid は CPU が 1 回分宣言する。
func (h *HoneymoonBridge) CpuBid() {
	if h.gameEndFlag || h.phase != HoneymoonBridgePhaseBid || h.IsHumanBidTurn() {
		return
	}
	level, suit := h.chooseCpuBid(h.currentPlayerIdx)
	_ = h.bidBy(h.currentPlayerIdx, level, suit)
}

// bidBy は playerIdx に宣言させる（level 0 = 降りる）。
func (h *HoneymoonBridge) bidBy(playerIdx, level, suit int) error {
	if h.gameEndFlag {
		return errors.New("game is over")
	}
	if h.phase != HoneymoonBridgePhaseBid {
		return errors.New("not the bidding phase")
	}
	if playerIdx != h.currentPlayerIdx {
		return fmt.Errorf("not player %d's turn", playerIdx)
	}

	if level == 0 {
		h.passCount++
		h.players[playerIdx].SetBid(0, 0)
		h.addLog(playerIdx, "pass", "パスしました", nil)
		// **2 回続けてパスしたら競りは締まる。**
		if h.passCount >= HoneymoonBridgePlayerCnt {
			h.closeBidding()
			return nil
		}
		h.currentPlayerIdx = (h.currentPlayerIdx + 1) % HoneymoonBridgePlayerCnt
		return nil
	}

	if level < 1 || level > HoneymoonBridgeMaxLevel {
		return fmt.Errorf("bid level must be 1..%d", HoneymoonBridgeMaxLevel)
	}
	if suit < 0 || suit > CardDesignDiamond {
		return fmt.Errorf("invalid bid suit: %d", suit)
	}
	if !h.outbids(level, suit) {
		return fmt.Errorf("bid %d-%d does not outbid the standing contract", level, suit)
	}

	h.passCount = 0
	h.contractLevel, h.trumpSuit, h.declarerIdx = level, suit, playerIdx
	h.players[playerIdx].SetBid(level, suit)
	h.addLog(playerIdx, "bid", fmt.Sprintf("%d %s を宣言", level, honeymoonBridgeContractSuitStr(suit)), nil)
	h.currentPlayerIdx = (h.currentPlayerIdx + 1) % HoneymoonBridgePlayerCnt
	return nil
}

// honeymoonBridgeContractSuitStr は契約スートの表示文字列を返す。
func honeymoonBridgeContractSuitStr(suit int) string {
	if suit == 0 {
		return "NT"
	}
	return suitStr(suit)
}

// closeBidding は競りを締めて本番のプレイへ進む。
func (h *HoneymoonBridge) closeBidding() {
	// **誰も宣言しなければディールは流れる。** 次のディールへ。
	if h.declarerIdx < 0 {
		h.addLog(-1, "passout", "両者パス。ディールをやり直します", nil)
		h.phase = HoneymoonBridgePhaseRoundEnd
		h.lastMade = false
		h.lastTricks = 0
		return
	}
	h.phase = HoneymoonBridgePhasePlay
	h.trickNumber = 0
	// **リードは落札者の相手から。**
	h.leadPlayerIdx = (h.declarerIdx + 1) % HoneymoonBridgePlayerCnt
	h.currentPlayerIdx = h.leadPlayerIdx
	h.addLog(h.declarerIdx, "contract",
		fmt.Sprintf("契約 %d %s（%d トリック必要）", h.contractLevel,
			honeymoonBridgeContractSuitStr(h.trumpSuit), h.RequiredTricks()), nil)
}

// RequiredTricks は契約に必要なトリック数を返す（ブック 6 + レベル）。
func (h *HoneymoonBridge) RequiredTricks() int {
	if h.contractLevel == 0 {
		return 0
	}
	return HoneymoonBridgeBookTricks + h.contractLevel
}

// chooseCpuBid は CPU の宣言。**強い札の枚数で決めます。**
func (h *HoneymoonBridge) chooseCpuBid(playerIdx int) (int, int) {
	p := h.players[playerIdx]
	bestSuit, bestScore := 0, -1
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		score := 0
		for i := range p.GetCardsSize() {
			c := p.GetCard(i)
			if c.GetDesign() == suit {
				score += 2
			}
			if honeymoonBridgeRank(c) >= 13 {
				score++
			}
		}
		if score > bestScore {
			bestSuit, bestScore = suit, score
		}
	}
	// 強さをレベルの見積もりに落とす。届かなければ降りる。
	level := (bestScore - 12) / 4
	level = min(max(level, 0), HoneymoonBridgeMaxLevel)
	if level < 1 {
		return 0, 0
	}
	nextLevel, nextSuit := h.NextBid()
	if nextLevel == 0 {
		return 0, 0 // 上限が立っている
	}
	if !h.outbids(level, bestSuit) {
		// 上回れないなら最小の上回り手を試し、それでも重ければ降りる。
		if nextLevel > level {
			return 0, 0
		}
		return nextLevel, nextSuit
	}
	return level, bestSuit
}

// finishRound はディールを精算する。
func (h *HoneymoonBridge) finishRound() {
	h.phase = HoneymoonBridgePhaseRoundEnd
	decl := h.players[h.declarerIdx]
	took := decl.GetTrickCount()
	need := h.RequiredTricks()
	h.lastTricks = took
	if took >= need {
		// **オーバートリックも点になる。**
		points := h.contractLevel * 10
		points += (took - need) * 5
		decl.AddScore(points)
		h.lastMade = true
		h.lastPoints = points
		h.addLog(h.declarerIdx, "score",
			fmt.Sprintf("契約 %d に対し %d トリック：成立 (+%d)", need, took, points), nil)
	} else {
		other := (h.declarerIdx + 1) % HoneymoonBridgePlayerCnt
		points := (need - took) * 10
		h.players[other].AddScore(points)
		h.lastMade = false
		h.lastPoints = points
		h.addLog(h.declarerIdx, "score",
			fmt.Sprintf("契約 %d に対し %d トリック：失敗、相手に +%d", need, took, points), nil)
	}
	for _, p := range h.players {
		if p.GetScore() >= h.config.Target {
			h.finishGame()
			return
		}
	}
}

// NextRound は次のディールを開始する。
func (h *HoneymoonBridge) NextRound() {
	if h.gameEndFlag || h.phase != HoneymoonBridgePhaseRoundEnd {
		return
	}
	h.dealerIdx = (h.dealerIdx + 1) % HoneymoonBridgePlayerCnt
	h.startRound()
}

// finishGame は終局処理。
func (h *HoneymoonBridge) finishGame() {
	h.phase = HoneymoonBridgePhaseGameEnd
	h.gameEndFlag = true
	switch {
	case h.players[0].GetScore() > h.players[1].GetScore():
		h.winnerIdx = 0
	case h.players[1].GetScore() > h.players[0].GetScore():
		h.winnerIdx = 1
	default:
		h.winnerIdx = -1
	}
	h.addLog(-1, "result", fmt.Sprintf("最終得点 %d - %d",
		h.players[0].GetScore(), h.players[1].GetScore()), nil)
}

// GiveUp は投了する。
func (h *HoneymoonBridge) GiveUp() {
	if h.gameEndFlag {
		return
	}
	h.phase = HoneymoonBridgePhaseGameEnd
	h.gameEndFlag = true
	h.winnerIdx = 1
	h.addLog(0, "giveup", "投了しました", nil)
}

// chooseCpuCard は CPU の手。
func (h *HoneymoonBridge) chooseCpuCard(playerIdx int) int {
	valid := h.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := h.players[playerIdx]
	// **引き合いは取っても得点にならない**ので安く出す。本番は取りにいく。
	wantTrick := h.phase == HoneymoonBridgePhasePlay
	pick, pickRank := valid[0], honeymoonBridgeRank(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		r := honeymoonBridgeRank(p.GetCard(i))
		if wantTrick && r > pickRank {
			pick, pickRank = i, r
		} else if !wantTrick && r < pickRank {
			pick, pickRank = i, r
		}
	}
	return pick
}

// HoneymoonBridgeHint はハネムーンブリッジの助言。
type HoneymoonBridgeHint struct {
	CardIndex *int
	Reason    string
	// Level / Suit は宣言すべき契約（プレイ中は 0）。
	Level int
	Suit  int
}

// GetHint は人間への助言を返す。
func (h *HoneymoonBridge) GetHint() *HoneymoonBridgeHint {
	if h.gameEndFlag {
		return nil
	}
	if h.IsHumanBidTurn() {
		level, suit := h.chooseCpuBid(0)
		reason := "honeymoonbridgeBid"
		if level == 0 {
			reason = "honeymoonbridgePass"
		}
		return &HoneymoonBridgeHint{Reason: reason, Level: level, Suit: suit}
	}
	if !h.IsHumanTurn() {
		return nil
	}
	valid := h.GetValidPlayIndices(0)
	if len(valid) == 0 {
		return nil
	}
	idx := h.chooseCpuCard(0)
	reason := "honeymoonbridgeWinTrick"
	if h.phase == HoneymoonBridgePhaseDraw {
		// **引き合いのトリックは得点にならない。** 何を引けるかだけが問題。
		reason = "honeymoonbridgeDraw"
	}
	return &HoneymoonBridgeHint{CardIndex: &idx, Reason: reason}
}

// honeymoonBridgeContains は xs が v を含むかを返す。
func honeymoonBridgeContains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// addLog は棋譜に 1 行足す。
func (h *HoneymoonBridge) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	h.appendLog(playerIdx, actionType, detail, cards)
}

// --- アクセサ ---------------------------------------------------------------

// GetConfig はゲーム設定を返す。
func (h *HoneymoonBridge) GetConfig() HoneymoonBridgeConfig { return h.config }

// SetConfig はゲーム設定を設定する。
func (h *HoneymoonBridge) SetConfig(cfg HoneymoonBridgeConfig) { h.config = cfg }

// GetPhase は現在のフェーズを返す。
func (h *HoneymoonBridge) GetPhase() HoneymoonBridgePhase { return h.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (h *HoneymoonBridge) GetGameEndFlag() bool { return h.gameEndFlag }

// GetRoundNumber は現在のディール番号を返す。
func (h *HoneymoonBridge) GetRoundNumber() int { return h.roundNumber }

// GetTrickNumber は現在のトリック番号を返す。
func (h *HoneymoonBridge) GetTrickNumber() int { return h.trickNumber }

// GetStockSize は山札の残り枚数を返す。
func (h *HoneymoonBridge) GetStockSize() int { return len(h.stock) }

// GetTrumpSuit は契約のスートを返す（0 = ノートランプ、競り前も 0）。
func (h *HoneymoonBridge) GetTrumpSuit() int { return h.trumpSuit }

// GetDeclarerIdx は落札者を返す（-1 = 競り中）。
func (h *HoneymoonBridge) GetDeclarerIdx() int { return h.declarerIdx }

// GetContractLevel は落札レベルを返す（0 = 未確定）。
func (h *HoneymoonBridge) GetContractLevel() int { return h.contractLevel }

// GetCurrentPlayerIdx は現在の手番を返す。
func (h *HoneymoonBridge) GetCurrentPlayerIdx() int { return h.currentPlayerIdx }

// GetLeadPlayerIdx はリードプレイヤーを返す。
func (h *HoneymoonBridge) GetLeadPlayerIdx() int { return h.leadPlayerIdx }

// GetDealerIdx はディーラーを返す。
func (h *HoneymoonBridge) GetDealerIdx() int { return h.dealerIdx }

// GetCurrentTrick は現在のトリックを返す。
func (h *HoneymoonBridge) GetCurrentTrick() []*TrickCard { return h.currentTrick }

// GetLastMade は直前のディールで契約が成立したかを返す。
func (h *HoneymoonBridge) GetLastMade() bool { return h.lastMade }

// GetLastTricks は直前のディールで落札者が取ったトリック数を返す。
func (h *HoneymoonBridge) GetLastTricks() int { return h.lastTricks }

// GetLastPoints は直前のディールで動いた点数を返す。
func (h *HoneymoonBridge) GetLastPoints() int { return h.lastPoints }

// GetPlayerCnt はプレイヤー数を返す。
func (h *HoneymoonBridge) GetPlayerCnt() int { return HoneymoonBridgePlayerCnt }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (h *HoneymoonBridge) GetPlayer(i int) *HoneymoonBridgePlayer {
	if i < 0 || i >= len(h.players) {
		return nil
	}
	return h.players[i]
}

// GetWinnerIdx は勝者を返す（-1 = 未確定/同点）。
func (h *HoneymoonBridge) GetWinnerIdx() int { return h.winnerIdx }

// GetActionLog は棋譜を返す。
func (h *HoneymoonBridge) GetActionLog() []*ActionLogEntry { return h.actionLog }

// honeymoonBridgeJSON は KV スナップショットの表現。
type honeymoonBridgeJSON struct {
	TrumpCards       *TrumpCards              `json:"tc"`
	Players          []*HoneymoonBridgePlayer `json:"pl"`
	Config           HoneymoonBridgeConfig    `json:"cf"`
	Phase            HoneymoonBridgePhase     `json:"ph"`
	Stock            []*Card                  `json:"st"`
	RoundNumber      int                      `json:"rn"`
	TrickNumber      int                      `json:"tn"`
	TrumpSuit        int                      `json:"ts"`
	DeclarerIdx      int                      `json:"di"`
	ContractLevel    int                      `json:"cl"`
	PassCount        int                      `json:"pc"`
	CurrentTrick     []*TrickCard             `json:"ct"`
	CurrentPlayerIdx int                      `json:"ci"`
	LeadPlayerIdx    int                      `json:"li"`
	DealerIdx        int                      `json:"dl"`
	LastMade         bool                     `json:"lm"`
	LastTricks       int                      `json:"lt"`
	GameEndFlag      bool                     `json:"ge"`
	WinnerIdx        int                      `json:"wi"`
	ActionLog        []*ActionLogEntry        `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (h *HoneymoonBridge) MarshalJSON() ([]byte, error) {
	return json.Marshal(&honeymoonBridgeJSON{
		TrumpCards: h.trumpCards, Players: h.players, Config: h.config, Phase: h.phase,
		Stock: h.stock, RoundNumber: h.roundNumber, TrickNumber: h.trickNumber,
		TrumpSuit: h.trumpSuit, DeclarerIdx: h.declarerIdx, ContractLevel: h.contractLevel,
		PassCount: h.passCount, CurrentTrick: h.currentTrick, CurrentPlayerIdx: h.currentPlayerIdx,
		LeadPlayerIdx: h.leadPlayerIdx, DealerIdx: h.dealerIdx,
		LastMade: h.lastMade, LastTricks: h.lastTricks,
		GameEndFlag: h.gameEndFlag, WinnerIdx: h.winnerIdx, ActionLog: h.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元
func (h *HoneymoonBridge) UnmarshalJSON(data []byte) error {
	var j honeymoonBridgeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < HoneymoonBridgePhaseDraw || j.Phase > HoneymoonBridgePhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	// **落札者と契約レベルは対。** 引き合い中は両方空、本番は両方埋まっている。
	//
	// **後者を落とすと復元後に panic する（レビュー指摘 PR #5312）。** 対として
	// 一致してさえいれば下の突き合わせは通ってしまうので、「本番なのに落札者が
	// いない」は素通りし、13 トリック目の `finishRound` が `h.players[-1]` を
	// 引いて落ちる。あり得ない**組み合わせ**を弾くのが目的で、範囲検査ではない。
	if j.Phase == HoneymoonBridgePhaseDraw && j.DeclarerIdx != -1 {
		return fmt.Errorf("declarer %d during the draw phase", j.DeclarerIdx)
	}
	if j.Phase == HoneymoonBridgePhasePlay && j.DeclarerIdx < 0 {
		return errors.New("play phase without a declarer")
	}
	if j.DeclarerIdx < -1 || j.DeclarerIdx >= HoneymoonBridgePlayerCnt {
		return fmt.Errorf("invalid declarer: %d", j.DeclarerIdx)
	}
	if j.ContractLevel < 0 || j.ContractLevel > HoneymoonBridgeMaxLevel {
		return fmt.Errorf("invalid contract level: %d", j.ContractLevel)
	}
	if (j.DeclarerIdx < 0) != (j.ContractLevel == 0) {
		return fmt.Errorf("declarer %d and contract level %d disagree", j.DeclarerIdx, j.ContractLevel)
	}
	// **切り札は 0（ノートランプ）か実在するスート。** レベル 0 なら 0 のみ。
	if j.TrumpSuit < 0 || j.TrumpSuit > CardDesignDiamond {
		return fmt.Errorf("invalid trump suit: %d", j.TrumpSuit)
	}
	if j.ContractLevel == 0 && j.TrumpSuit != 0 {
		return fmt.Errorf("trump suit %d without a contract", j.TrumpSuit)
	}
	// **引き合い中は山札が減っていく。** 本番に入ったら空。
	if len(j.Stock) > HoneymoonBridgeStockSize {
		return fmt.Errorf("stock holds %d cards", len(j.Stock))
	}
	if j.Phase >= HoneymoonBridgePhaseBid && len(j.Stock) != 0 {
		return fmt.Errorf("stock still holds %d cards after the draw phase", len(j.Stock))
	}
	if j.RoundNumber < 1 {
		return fmt.Errorf("invalid round number: %d", j.RoundNumber)
	}
	if j.TrickNumber < 0 || j.TrickNumber > HoneymoonBridgeTricksPerPhase {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.PassCount < 0 || j.PassCount > HoneymoonBridgePlayerCnt {
		return fmt.Errorf("invalid pass count: %d", j.PassCount)
	}
	if len(j.CurrentTrick) > HoneymoonBridgePlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	// **枚数だけでなく中身も見る（#5310 で踏んだ panic の再発防止）。**
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= HoneymoonBridgePlayerCnt {
			return errors.New("invalid current trick entry")
		}
	}
	if len(j.ActionLog) > honeymoonBridgeMaxSliceLen {
		return errors.New("honeymoonbridge: input array exceeds maximum allowed size")
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"dealer":         j.DealerIdx,
	} {
		if idx < 0 || idx >= HoneymoonBridgePlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= HoneymoonBridgePlayerCnt {
		return fmt.Errorf("invalid winner: %d", j.WinnerIdx)
	}
	if !j.GameEndFlag && j.WinnerIdx != -1 {
		return fmt.Errorf("winner %d before the game ended", j.WinnerIdx)
	}
	if j.LastTricks < 0 || j.LastTricks > HoneymoonBridgeTricksPerPhase {
		return fmt.Errorf("invalid last tricks: %d", j.LastTricks)
	}

	if j.TrumpCards != nil {
		h.trumpCards = j.TrumpCards
	}
	if len(j.Players) == HoneymoonBridgePlayerCnt {
		h.players = j.Players
	}
	h.config, h.phase, h.stock = j.Config, j.Phase, j.Stock
	h.roundNumber, h.trickNumber, h.trumpSuit = j.RoundNumber, j.TrickNumber, j.TrumpSuit
	h.declarerIdx, h.contractLevel, h.passCount = j.DeclarerIdx, j.ContractLevel, j.PassCount
	h.currentTrick, h.currentPlayerIdx = j.CurrentTrick, j.CurrentPlayerIdx
	h.leadPlayerIdx, h.dealerIdx = j.LeadPlayerIdx, j.DealerIdx
	h.lastMade, h.lastTricks = j.LastMade, j.LastTricks
	h.gameEndFlag, h.winnerIdx, h.actionLog = j.GameEndFlag, j.WinnerIdx, j.ActionLog
	return nil
}
