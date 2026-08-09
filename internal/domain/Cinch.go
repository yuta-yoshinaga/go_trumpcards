//go:build !js || !wasm || extra

// Package domain チンチ (Cinch / Double Pedro / High Five) のドメインモデル。
//
// Cinch は All-Fours / Pitch 系のオークション・トリックテイキングゲーム。4 人 (本実装
// では 1 human + 3 CPU の個人戦) ・52 枚デッキで、各ディール 9 枚ずつ配る。ビッドで
// 切り札を指定する権利 (ピッチャー) を競り、最高ビッダーが切り札スートを宣言して
// リードする。マストフォロー (ただし切り札はいつでも合法) で 9 トリックをプレイし、
// 獲得したポイントカードで得点する。ビッダー側は宣言ポイント以上を取れなければ
// セットバック (宣言分だけ減点) となる。先に目標得点へ到達したプレイヤーの勝ち。
//
// ポイントカード (1 ディール計 14 点):
//   - High  : 切り札の A = 1
//   - King  : 切り札の K = 1
//   - Ten   : 切り札の 10 ("Game") = 1
//   - Jack  : 切り札の J = 1
//   - Right Pedro : 切り札の 5 = 5
//   - Left Pedro  : 切り札と同色スートの 5 = 5
//     合計 = 1 + 1 + 1 + 1 + 5 + 5 = 14
//
// 切り札ランク (高→低): A K Q J 10 9 8 7 6 5(Right Pedro) 5-同色(Left Pedro) 4 3 2。
// Left Pedro (切り札と同色スートの 5) は切り札として扱われ、Right Pedro のすぐ下に
// 位置づけられる (トリック勝敗・ポイント計算の双方で切り札扱い)。オフ切り札スートは
// 通常どおり A..2 のランク。
//
// 本実装は extra ワーカーから到達可能なようビッド・スコア・ランク・CPU ロジックを
// すべてインラインで持つ (Classic タグの Pitch には依存しない)。extra 到達可能な
// NewTrumpCards(0) で 52 枚デッキを生成する。
package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// CinchPlayerCnt はチンチのプレイヤー数 (固定 4, 個人戦)。
const CinchPlayerCnt = 4

// CinchHandSize は各プレイヤーの手札枚数。
const CinchHandSize = 9

// CinchTotalTricks は 1 ディールのトリック数。
const CinchTotalTricks = CinchHandSize

// CinchTotalPoints は 1 ディールの総ポイント数 (14)。
const CinchTotalPoints = 14

// CinchMinBid は最小ビッド値 (パス=0 を除く有効最小値)。
const CinchMinBid = 1

// CinchMaxBid は最大ビッド値 (= 全ポイント獲得宣言 / Cinch)。
const CinchMaxBid = CinchTotalPoints

// CinchPassBid はパスを表すビッド値。
const CinchPassBid = 0

// CinchTrumpUnset は切り札未確定を表す値。
const CinchTrumpUnset = 0

// CinchPhase はゲームフェーズ。
type CinchPhase int

// Cinch のフェーズ定数
const (
	// CinchPhaseBid ビッドフェーズ
	CinchPhaseBid CinchPhase = 0
	// CinchPhaseNameTrump ビッド勝者が切り札を宣言するフェーズ
	CinchPhaseNameTrump CinchPhase = 1
	// CinchPhasePlay トリックプレイフェーズ
	CinchPhasePlay CinchPhase = 2
	// CinchPhaseTrickEnd トリック終了フェーズ
	CinchPhaseTrickEnd CinchPhase = 3
	// CinchPhaseRoundEnd ラウンド (ディール) 終了フェーズ
	CinchPhaseRoundEnd CinchPhase = 4
	// CinchPhaseGameEnd ゲーム終了フェーズ
	CinchPhaseGameEnd CinchPhase = 5
)

// CinchDealDetail は 1 ディールの得点内訳。
type CinchDealDetail struct {
	TrumpSuit int         // 切り札スート
	BidderIdx int         // ビッダー
	Bid       int         // 宣言値
	SetBack   bool        // ビッダーがセットバックしたか
	Points    map[int]int // プレイヤー別に獲得した「生」ポイント (High/Jack/Pedro/Game 等)
	Gained    map[int]int // プレイヤー別にこのディールで実際に加算した得点 (セットバック反映後)
}

// CinchHint はヒント情報。
type CinchHint struct {
	CardIndices []int  // 推奨カードインデックス (play フェーズ)
	Bid         *int   // 推奨ビッド値 (bid フェーズ; パスは 0)
	TrumpSuit   *int   // 推奨切り札スート (nameTrump フェーズ)
	Reason      string // ヒント理由キー
}

// Cinch はチンチゲームの状態を保持する集約ルート。
type Cinch struct {
	trumpCards       *TrumpCards
	players          []*CinchPlayer
	config           CinchConfig
	phase            CinchPhase
	roundNumber      int
	trickNumber      int
	dealerIdx        int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	lastTrick        []*TrickCard
	lastTrickWinner  int
	leadPlayerIdx    int
	bidPlayerIdx     int // 現在ビッド中のプレイヤー
	currentBid       int // 最高ビッド値 (0=未ビッド/パスのみ)
	bidWinnerIdx     int // 最高ビッダーのインデックス (-1=未確定)
	trumpSuit        int // 切り札スート (CinchTrumpUnset=未確定)
	gameEndFlag      bool
	winnerIdx        int
	lastDealDetail   *CinchDealDetail
	actionLogBase
}

// NewCinch はコンストラクタ。
func NewCinch(trumpCards *TrumpCards, players []*CinchPlayer, config CinchConfig) *Cinch {
	return &Cinch{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerIdx:       -1,
		bidWinnerIdx:    -1,
		lastTrickWinner: -1,
		dealerIdx:       CinchPlayerCnt - 1,
		roundNumber:     0,
	}
}

// NewDefaultCinch は標準の 4 人構成 (1 human + 3 CPU) と DefaultCinchConfig で Cinch を
// 生成する。CUI / Web / Worker の構築の単一情報源。
func NewDefaultCinch() *Cinch {
	players := make([]*CinchPlayer, CinchPlayerCnt)
	players[0] = NewCinchPlayer(true)
	for i := 1; i < CinchPlayerCnt; i++ {
		players[i] = NewCinchPlayer(false)
	}
	return NewCinch(newCinchDeck(), players, DefaultCinchConfig())
}

// newCinchDeck はチンチ用 52 枚デッキを生成する。NewTrumpCards はビルドタグ無しの
// TrumpCards.go にあり extra ワーカーからも到達可能。
func newCinchDeck() *TrumpCards {
	return NewTrumpCards(0)
}

// --- Cinch 特有のランク / 色ヘルパー ---

// cinchSameColorSuit は指定スートと同色のもう一方のスートを返す。
func cinchSameColorSuit(suit int) int {
	switch suit {
	case CardDesignSpade:
		return CardDesignClover
	case CardDesignClover:
		return CardDesignSpade
	case CardDesignHeart:
		return CardDesignDiamond
	case CardDesignDiamond:
		return CardDesignHeart
	default:
		return suit
	}
}

// cinchRankValue はオフ切り札スートの比較用ランク (A=14 > K=13 > … > 2=2)。
func cinchRankValue(v int) int {
	if v == 1 {
		return 14
	}
	return v
}

// cinchIsTrump はカードが (Left Pedro を含む) 切り札として扱われるかを返す。
// 切り札スートのカードすべて、および切り札と同色スートの 5 (Left Pedro) が切り札。
func cinchIsTrump(c *Card, trumpSuit int) bool {
	if trumpSuit == CinchTrumpUnset || c == nil {
		return false
	}
	if c.GetDesign() == trumpSuit {
		return true
	}
	return c.GetValue() == 5 && c.GetDesign() == cinchSameColorSuit(trumpSuit)
}

// cinchTrumpRank は切り札としての強さを返す (大きいほど強い)。
// 切り札ランク: A(20) K(19) Q(18) J(17) 10(16) 9(15) 8(14) 7(13) 6(12)
//
//	5=Right Pedro(11) 同色5=Left Pedro(10) 4(9) 3(8) 2(7)。
//
// これにより Left Pedro は Right Pedro (切り札の 5) のすぐ下に位置づけられる。
func cinchTrumpRank(c *Card, trumpSuit int) int {
	if c.GetValue() == 5 && c.GetDesign() == cinchSameColorSuit(trumpSuit) {
		return 10 // Left Pedro
	}
	// 切り札スート本体。
	switch c.GetValue() {
	case 1: // A
		return 20
	case 13: // K
		return 19
	case 12: // Q
		return 18
	case 11: // J
		return 17
	case 10:
		return 16
	case 9:
		return 15
	case 8:
		return 14
	case 7:
		return 13
	case 6:
		return 12
	case 5: // Right Pedro
		return 11
	case 4:
		return 9
	case 3:
		return 8
	case 2:
		return 7
	default:
		return 0
	}
}

// cinchPointValue はカードが獲得された際のポイント値を返す (切り札のみ)。
//
//	High(A)=1 King(K)=1 Ten(10)=1 Jack(J)=1 Right Pedro(5 of trump)=5 Left Pedro(同色5)=5。
func cinchPointValue(c *Card, trumpSuit int) int {
	if !cinchIsTrump(c, trumpSuit) {
		return 0
	}
	// Left Pedro (同色スートの 5)。
	if c.GetValue() == 5 && c.GetDesign() != trumpSuit {
		return 5
	}
	switch c.GetValue() {
	case 1, 13, 11, 10:
		return 1
	case 5: // Right Pedro
		return 5
	default:
		return 0
	}
}

// CinchHandPointsBySuit returns, for each candidate trump suit (1..4, indexed by
// the CardDesign constants), how many of the 14 deal points the hand already
// holds. Index 0 is unused.
//
// Holding a point card is not the same as capturing it, so this is a bidding
// guide, not a promise. The Web GUI has shown the same table since the game
// shipped; the CUI showed only the current high bid (#4845).
func CinchHandPointsBySuit(cards []*Card) [CardDesignDiamond + 1]int {
	var points [CardDesignDiamond + 1]int
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		for _, c := range cards {
			points[suit] += cinchPointValue(c, suit)
		}
	}
	return points
}

// CinchBestTrumpSuit returns the suit holding the most points, the lowest suit
// index winning ties (same rule as the Web GUI's estimateCinchBidStrength).
func CinchBestTrumpSuit(points [CardDesignDiamond + 1]int) int {
	best := CardDesignSpade
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if points[suit] > points[best] {
			best = suit
		}
	}
	return best
}

// --- ゲーム進行 ---

// Reset は新しいゲームを開始する。累計得点もクリアする。
func (g *Cinch) Reset() {
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundNumber = 1
	g.trickNumber = 0
	g.currentTrick = nil
	g.lastTrick = nil
	g.lastTrickWinner = -1
	g.leadPlayerIdx = -1
	g.currentPlayerIdx = -1
	g.dealerIdx = CinchPlayerCnt - 1
	g.bidPlayerIdx = (g.dealerIdx + 1) % CinchPlayerCnt
	g.currentBid = 0
	g.bidWinnerIdx = -1
	g.trumpSuit = CinchTrumpUnset
	g.lastDealDetail = nil
	g.actionLog = make([]*ActionLogEntry, 0)

	for _, p := range g.players {
		p.ResetDeal()
		p.ResetTotalScore()
	}

	g.trumpCards = newCinchDeck()
	g.trumpCards.Shuffle()
	g.dealRound()
	g.sortAllHands()

	g.phase = CinchPhaseBid
}

// dealRound は各プレイヤーへ CinchHandSize 枚配る。
func (g *Cinch) dealRound() {
	for i := 0; i < CinchHandSize; i++ {
		for j := 0; j < CinchPlayerCnt; j++ {
			card := g.trumpCards.DrawCard()
			if card == nil {
				return
			}
			g.players[j].AddCard(card)
		}
	}
}

// NextRound は次のディールを開始する。
func (g *Cinch) NextRound() {
	if g.phase != CinchPhaseRoundEnd || g.gameEndFlag {
		return
	}
	g.roundNumber++
	g.trickNumber = 0
	g.currentTrick = nil
	g.lastTrick = nil
	g.lastTrickWinner = -1
	g.leadPlayerIdx = -1
	g.currentPlayerIdx = -1
	g.dealerIdx = (g.dealerIdx + 1) % CinchPlayerCnt
	g.bidPlayerIdx = (g.dealerIdx + 1) % CinchPlayerCnt
	g.currentBid = 0
	g.bidWinnerIdx = -1
	g.trumpSuit = CinchTrumpUnset

	for _, p := range g.players {
		p.ResetDeal()
	}

	g.trumpCards = newCinchDeck()
	g.trumpCards.Shuffle()
	g.dealRound()
	g.sortAllHands()

	g.phase = CinchPhaseBid
}

// PlayerBid は人間プレイヤーがビッドする (bid: 0=pass, 1..CinchMaxBid)。
func (g *Cinch) PlayerBid(bid int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CinchPhaseBid {
		return ErrWrongPhase
	}
	humanIdx := findHumanIdx(g.players)
	if humanIdx < 0 || g.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	if err := g.validateBidValue(humanIdx, bid); err != nil {
		return err
	}
	g.applyBid(humanIdx, bid)
	g.advanceBid()
	return nil
}

// CpuBid は現在のビッド手番が CPU の場合にビッドする。
func (g *Cinch) CpuBid() {
	if g.gameEndFlag || g.phase != CinchPhaseBid {
		return
	}
	if g.bidPlayerIdx < 0 || g.bidPlayerIdx >= CinchPlayerCnt {
		return
	}
	if g.players[g.bidPlayerIdx].GetIsHuman() {
		return
	}
	bid := g.cpuSelectBid(g.bidPlayerIdx)
	g.applyBid(g.bidPlayerIdx, bid)
	g.advanceBid()
}

// validateBidValue はビッド値が有効か検証する。
func (g *Cinch) validateBidValue(playerIdx, bid int) error {
	if bid == CinchPassBid {
		// 親 (dealer) は他全員パスの場合、必ず stuck されるためパス不可。
		if playerIdx == g.dealerIdx && g.currentBid == 0 && g.bidsCompleted() == CinchPlayerCnt-1 {
			return NewDomainError(ErrInvalidPlay, "親 (dealer) は全員パスの場合パスできません")
		}
		return nil
	}
	if bid < CinchMinBid || bid > CinchMaxBid {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ビッドは pass(0) または %d〜%d で指定してください", CinchMinBid, CinchMaxBid))
	}
	if bid <= g.currentBid {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ビッドは現在の最高 %d を超える必要があります", g.currentBid))
	}
	return nil
}

// applyBid はビッド値をプレイヤーに反映し、最高ビッド/勝者を更新する。
func (g *Cinch) applyBid(playerIdx, bid int) {
	g.players[playerIdx].SetBid(bid)
	if bid > g.currentBid {
		g.currentBid = bid
		g.bidWinnerIdx = playerIdx
	}
	logBid := fmt.Sprintf("%d", bid)
	if bid == CinchPassBid {
		logBid = "pass"
	}
	g.appendLog(playerIdx, "bid", fmt.Sprintf("%s bids %s", playerName(g.players, playerIdx), logBid), nil)
}

// advanceBid は次のビッド手番へ進める。全員終わればトランプ宣言へ移る (stuck dealer も処理)。
func (g *Cinch) advanceBid() {
	bidsDone := g.bidsCompleted()
	if bidsDone < CinchPlayerCnt {
		g.bidPlayerIdx = (g.bidPlayerIdx + 1) % CinchPlayerCnt
		// 親に到達し、かつ全員パス済みの場合 stuck 強制。
		if g.bidPlayerIdx == g.dealerIdx && g.currentBid == 0 && bidsDone == CinchPlayerCnt-1 {
			g.applyBid(g.dealerIdx, CinchMinBid)
			g.appendLog(g.dealerIdx, "stuck", fmt.Sprintf("%s is stuck with %d", playerName(g.players, g.dealerIdx), CinchMinBid), nil)
			g.startNameTrump()
		}
		return
	}
	g.startNameTrump()
}

// bidsCompleted は何人のプレイヤーがビッド済みかを返す。
func (g *Cinch) bidsCompleted() int {
	cnt := 0
	for _, p := range g.players {
		if p.GetBid() != -1 {
			cnt++
		}
	}
	return cnt
}

// startNameTrump はトランプ宣言フェーズを開始する。
func (g *Cinch) startNameTrump() {
	if g.bidWinnerIdx < 0 {
		// 安全策: 親 stuck。
		g.bidWinnerIdx = g.dealerIdx
		g.currentBid = CinchMinBid
		g.players[g.dealerIdx].SetBid(CinchMinBid)
	}
	g.phase = CinchPhaseNameTrump
	g.appendLog(g.bidWinnerIdx, "bid_won",
		fmt.Sprintf("%s wins the bid at %d and will name trump", playerName(g.players, g.bidWinnerIdx), g.currentBid), nil)
}

// NameTrump は人間のビッド勝者が切り札スートを宣言する。
func (g *Cinch) NameTrump(suit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CinchPhaseNameTrump {
		return ErrWrongPhase
	}
	if !g.players[g.bidWinnerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	return g.applyNameTrump(suit)
}

// applyNameTrump は切り札宣言の共通処理 (human / CPU)。
func (g *Cinch) applyNameTrump(suit int) error {
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return NewDomainError(ErrInvalidPlay, "切り札スートは 1〜4 で指定してください")
	}
	g.trumpSuit = suit
	g.appendLog(g.bidWinnerIdx, "trump_set", fmt.Sprintf("Trump is %s", suitName(suit)), nil)
	g.startPlayPhase()
	return nil
}

// startPlayPhase はプレイフェーズを開始する: ビッド勝者がリードする。
func (g *Cinch) startPlayPhase() {
	g.leadPlayerIdx = g.bidWinnerIdx
	g.currentPlayerIdx = g.bidWinnerIdx
	g.trickNumber = 1
	g.currentTrick = nil
	g.phase = CinchPhasePlay
}

// PlayerPlay は人間プレイヤーがカードを出す。
func (g *Cinch) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CinchPhasePlay {
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

// CpuPlay は現在の手番が CPU の場合に 1 ターン実行する。ビッド・トランプ宣言・
// プレイの各フェーズを進める。
func (g *Cinch) CpuPlay() {
	if g.gameEndFlag {
		return
	}
	switch g.phase {
	case CinchPhaseBid:
		g.CpuBid()
	case CinchPhaseNameTrump:
		if g.players[g.bidWinnerIdx].GetIsHuman() {
			return
		}
		_ = g.applyNameTrump(g.cpuSelectTrump(g.bidWinnerIdx))
	case CinchPhasePlay:
		if g.players[g.currentPlayerIdx].GetIsHuman() {
			return
		}
		player := g.players[g.currentPlayerIdx]
		cardIdx := g.cpuSelectPlayCard(g.currentPlayerIdx)
		played := player.RemoveCard(cardIdx)
		// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
		// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
		// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
		if played == nil {
			return
		}
		g.playCard(g.currentPlayerIdx, played)
	}
}

// playCard はカードをプレイする共通処理。
func (g *Cinch) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == CinchPlayerCnt {
		g.phase = CinchPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % CinchPlayerCnt
	}
}

// validatePlay はカードのプレイが有効か検証する。
// ルール: リードスートを持っていれば、リードスート OR 切り札のみプレイ可。
// リードスートを持たなければ任意のカードをプレイ可。切り札はいつでも合法。
// Left Pedro は切り札として扱われるため、切り札リードに対しては切り札扱いとなる。
func (g *Cinch) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil // リードは任意
	}
	leadIsTrump := cinchIsTrump(g.currentTrick[0].Card, g.trumpSuit)
	if leadIsTrump {
		// 切り札がリードされた: 切り札を持っていれば切り札を出さなければならない。
		if cinchIsTrump(card, g.trumpSuit) {
			return nil
		}
		if g.playerHasTrump(playerIdx) {
			return NewDomainError(ErrInvalidPlay, "切り札のリードには切り札で従ってください")
		}
		return nil
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	// Left Pedro は切り札扱いなので、オフ切り札リードに対しては常に合法 (切ることが可能)。
	if cinchIsTrump(card, g.trumpSuit) {
		return nil
	}
	if card.GetDesign() == leadSuit {
		return nil
	}
	if g.playerHasOffSuit(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従うか切り札を切ってください")
	}
	return nil
}

// playerHasTrump はプレイヤーが (Left Pedro を含む) 切り札を持っているか。
func (g *Cinch) playerHasTrump(playerIdx int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if cinchIsTrump(p.GetCard(i), g.trumpSuit) {
			return true
		}
	}
	return false
}

// playerHasOffSuit はプレイヤーが「切り札扱いでない」指定スートのカードを持っているか。
// Left Pedro は切り札として扱うため、同色スートの 5 はカウントしない。
func (g *Cinch) playerHasOffSuit(playerIdx, design int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == design && !cinchIsTrump(c, g.trumpSuit) {
			return true
		}
	}
	return false
}

// ResolveTrick はトリックを解決して勝者を決定する。
func (g *Cinch) ResolveTrick() {
	if g.phase != CinchPhaseTrickEnd || len(g.currentTrick) != CinchPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.lastTrick = g.currentTrick
	g.lastTrickWinner = winnerIdx
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", playerName(g.players, winnerIdx), g.trickNumber), trickCards)
	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= CinchTotalTricks {
		g.phase = CinchPhaseRoundEnd
	} else {
		g.phase = CinchPhaseTrickEnd
	}
}

// NextTrick は次のトリックを開始する。
func (g *Cinch) NextTrick() {
	if g.phase != CinchPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = CinchPhasePlay
}

// trickWinner はトリックの勝者を決定する (切り札最高 > リード最高)。
// Left Pedro は切り札として cinchTrumpRank で比較される。
func (g *Cinch) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return g.leadPlayerIdx
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerCard := g.currentTrick[0].Card
	for _, tc := range g.currentTrick[1:] {
		if g.cardBeats(tc.Card, winnerCard, leadSuit) {
			winnerIdx = tc.PlayerIdx
			winnerCard = tc.Card
		}
	}
	return winnerIdx
}

// cardBeats は candidate が current より強いかを判定する。
func (g *Cinch) cardBeats(candidate, current *Card, leadSuit int) bool {
	candTrump := cinchIsTrump(candidate, g.trumpSuit)
	curTrump := cinchIsTrump(current, g.trumpSuit)
	switch {
	case candTrump && !curTrump:
		return true
	case !candTrump && curTrump:
		return false
	case candTrump && curTrump:
		return cinchTrumpRank(candidate, g.trumpSuit) > cinchTrumpRank(current, g.trumpSuit)
	default:
		// どちらも切り札でない: リードスートのみ勝負に絡む。
		if candidate.GetDesign() != leadSuit {
			return false
		}
		if current.GetDesign() != leadSuit {
			return true
		}
		return cinchRankValue(candidate.GetValue()) > cinchRankValue(current.GetValue())
	}
}

// ScoreRound はラウンドのスコアを確定し、ゲーム終了判定を行う。
func (g *Cinch) ScoreRound() {
	if g.phase != CinchPhaseRoundEnd {
		return
	}
	rawPoints := g.computeRoundPoints()
	gained := make(map[int]int, CinchPlayerCnt)
	setBack := false

	for i := 0; i < CinchPlayerCnt; i++ {
		pts := rawPoints[i]
		if i == g.bidWinnerIdx {
			if pts < g.currentBid {
				gained[i] = -g.currentBid
				setBack = true
				g.appendLog(i, "set_back",
					fmt.Sprintf("%s set back: bid=%d earned=%d -> %d",
						playerName(g.players, i), g.currentBid, pts, -g.currentBid), nil)
			} else {
				gained[i] = pts
				g.appendLog(i, "bid_made",
					fmt.Sprintf("%s makes bid: bid=%d earned=%d -> +%d",
						playerName(g.players, i), g.currentBid, pts, pts), nil)
			}
		} else {
			gained[i] = pts
			if pts > 0 {
				g.appendLog(i, "non_bidder_score", fmt.Sprintf("%s scores %d", playerName(g.players, i), pts), nil)
			}
		}
	}
	for i := 0; i < CinchPlayerCnt; i++ {
		g.players[i].AddScore(gained[i])
	}
	g.lastDealDetail = &CinchDealDetail{
		TrumpSuit: g.trumpSuit,
		BidderIdx: g.bidWinnerIdx,
		Bid:       g.currentBid,
		SetBack:   setBack,
		Points:    rawPoints,
		Gained:    gained,
	}
	for i := 0; i < CinchPlayerCnt; i++ {
		g.appendLog(i, "cumulative_score",
			fmt.Sprintf("%s: total=%d", playerName(g.players, i), g.players[i].GetTotalScore()), nil)
	}
	g.checkGameEnd()
}

// computeRoundPoints は各プレイヤーが獲得したポイント数を返す。
// High(切り札 A)=1 King(切り札 K)=1 Ten(切り札 10, Game)=1 Jack(切り札 J)=1
// Right Pedro(切り札 5)=5 Left Pedro(同色 5)=5, 計 14 点。
func (g *Cinch) computeRoundPoints() map[int]int {
	points := make(map[int]int, CinchPlayerCnt)
	for i := 0; i < CinchPlayerCnt; i++ {
		points[i] = 0
	}
	if g.trumpSuit == CinchTrumpUnset {
		return points
	}
	for playerIdx, p := range g.players {
		for _, trick := range p.GetTricksTaken() {
			for _, card := range trick {
				if pv := cinchPointValue(card, g.trumpSuit); pv > 0 {
					points[playerIdx] += pv
					g.appendLog(playerIdx, "score_point",
						fmt.Sprintf("%s captures %s (%d pt)", playerName(g.players, playerIdx), cardStr(card), pv), nil)
				}
			}
		}
	}
	return points
}

// checkGameEnd はゲーム終了判定を行う: PointLimit 到達者がいれば終了。
// ビッダーは同点時に到達優先とする (自分のビッドで攻めた側を勝者に)。
func (g *Cinch) checkGameEnd() {
	bidder := g.bidWinnerIdx
	if bidder >= 0 && g.players[bidder].GetTotalScore() >= g.config.PointLimit {
		g.finishGame(bidder)
		return
	}
	maxScore := -1 << 30
	winner := -1
	hasWinner := false
	for i := 0; i < CinchPlayerCnt; i++ {
		score := g.players[i].GetTotalScore()
		if score >= g.config.PointLimit {
			hasWinner = true
		}
		if score > maxScore {
			maxScore = score
			winner = i
		}
	}
	if !hasWinner {
		return
	}
	g.finishGame(winner)
}

// finishGame はゲーム終了状態へ遷移する。
func (g *Cinch) finishGame(winner int) {
	g.gameEndFlag = true
	g.phase = CinchPhaseGameEnd
	g.winnerIdx = winner
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", playerName(g.players, winner)), nil)
}

// --- ヘルパー ---

func (g *Cinch) sortAllHands() {
	for _, p := range g.players {
		cinchSortHand(p)
	}
}

// cinchSortHand はプレイヤーの手札をスート→ランクの順にソートする (表示用, 切り札未確定)。
func cinchSortHand(p *CinchPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return cinchRankValue(cards[i].GetValue()) < cinchRankValue(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

func (g *Cinch) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- 状態アクセサ ---

// GetPhase は現在のフェーズを返す。
func (g *Cinch) GetPhase() CinchPhase { return g.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (g *Cinch) SetPhase(phase CinchPhase) { g.phase = phase }

// GetRoundNumber はラウンド番号を返す。
func (g *Cinch) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber はラウンド番号を設定する (テスト用)。
func (g *Cinch) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber はトリック番号を返す。
func (g *Cinch) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber はトリック番号を設定する (テスト用)。
func (g *Cinch) SetTrickNumber(n int) { g.trickNumber = n }

// GetDealerIdx は親インデックスを返す。
func (g *Cinch) GetDealerIdx() int { return g.dealerIdx }

// SetDealerIdx は親インデックスを設定する (テスト用)。
func (g *Cinch) SetDealerIdx(idx int) { g.dealerIdx = idx }

// GetCurrentTurn は現在の手番プレイヤーインデックスを返す。
func (g *Cinch) GetCurrentTurn() int { return g.currentPlayerIdx }

// SetCurrentTurn は現在の手番を設定する (テスト用)。
func (g *Cinch) SetCurrentTurn(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick は進行中のトリックを返す。
func (g *Cinch) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick は進行中のトリックを設定する (テスト用)。
func (g *Cinch) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLastTrick は直前に完了したトリックを返す。
func (g *Cinch) GetLastTrick() []*TrickCard { return g.lastTrick }

// GetLastTrickWinner は直前トリックの勝者を返す (-1=なし)。
func (g *Cinch) GetLastTrickWinner() int { return g.lastTrickWinner }

// GetLeadPlayerIdx はリードプレイヤーインデックスを返す。
func (g *Cinch) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx はリードプレイヤーインデックスを設定する (テスト用)。
func (g *Cinch) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetBidPlayerIdx は現在ビッド中のプレイヤーインデックスを返す。
func (g *Cinch) GetBidPlayerIdx() int { return g.bidPlayerIdx }

// SetBidPlayerIdx はビッドプレイヤーインデックスを設定する (テスト用)。
func (g *Cinch) SetBidPlayerIdx(idx int) { g.bidPlayerIdx = idx }

// GetCurrentBid は現在の最高ビッド値を返す。
func (g *Cinch) GetCurrentBid() int { return g.currentBid }

// SetCurrentBid は現在の最高ビッド値を設定する (テスト用)。
func (g *Cinch) SetCurrentBid(bid int) { g.currentBid = bid }

// GetBidWinnerIdx は最高ビッダーのインデックスを返す (-1=未確定)。
func (g *Cinch) GetBidWinnerIdx() int { return g.bidWinnerIdx }

// SetBidWinnerIdx は最高ビッダーのインデックスを設定する (テスト用)。
func (g *Cinch) SetBidWinnerIdx(idx int) { g.bidWinnerIdx = idx }

// GetTrumpSuit は切り札スートを返す (CinchTrumpUnset=未確定)。
func (g *Cinch) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit は切り札スートを設定する (テスト用)。
func (g *Cinch) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Cinch) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx は勝者インデックスを返す (-1=未確定)。
func (g *Cinch) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt はプレイヤー数を返す。
func (g *Cinch) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *Cinch) GetPlayer(i int) *CinchPlayer {
	return getPlayer(g.players, i)
}

// GetLastDealDetail は直前ディールの得点内訳を返す (nil の場合もある)。
func (g *Cinch) GetLastDealDetail() *CinchDealDetail { return g.lastDealDetail }

// IsHumanTurn は現在の意思決定者が人間かを返す。
func (g *Cinch) IsHumanTurn() bool {
	if g.gameEndFlag {
		return false
	}
	switch g.phase {
	case CinchPhaseBid:
		return g.bidPlayerIdx >= 0 && g.bidPlayerIdx < len(g.players) && g.players[g.bidPlayerIdx].GetIsHuman()
	case CinchPhaseNameTrump:
		return g.bidWinnerIdx >= 0 && g.bidWinnerIdx < len(g.players) && g.players[g.bidWinnerIdx].GetIsHuman()
	case CinchPhasePlay:
		return g.currentPlayerIdx >= 0 && g.currentPlayerIdx < len(g.players) && g.players[g.currentPlayerIdx].GetIsHuman()
	default:
		return false
	}
}

// GetConfig はローカルルール設定を返す。
func (g *Cinch) GetConfig() CinchConfig { return g.config }

// SetConfig はローカルルール設定を変更する。
func (g *Cinch) SetConfig(cfg CinchConfig) { g.config = cfg }

// GetPlayableIndices はプレイフェーズでプレイ可能な手札インデックスを返す。
func (g *Cinch) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != CinchPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// GetRoundWinners は (ゲーム終了時) 最高得点プレイヤーのリストを返す。
func (g *Cinch) GetRoundWinners() []int {
	if !g.gameEndFlag {
		return nil
	}
	best := g.players[0].GetTotalScore()
	for _, p := range g.players[1:] {
		if p.GetTotalScore() > best {
			best = p.GetTotalScore()
		}
	}
	winners := make([]int, 0)
	for i, p := range g.players {
		if p.GetTotalScore() == best {
			winners = append(winners, i)
		}
	}
	return winners
}

// --- JSON Serialization ---

// cinchJSON is the JSON wire format for Cinch.
type cinchJSON struct {
	TrumpCards       *TrumpCards       `json:"tc"`
	Players          []*CinchPlayer    `json:"ps"`
	Config           CinchConfig       `json:"cf"`
	Phase            CinchPhase        `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TrickNumber      int               `json:"tn"`
	DealerIdx        int               `json:"di"`
	CurrentPlayerIdx int               `json:"ci"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	LastTrick        []*TrickCard      `json:"lt"`
	LastTrickWinner  int               `json:"lw"`
	LeadPlayerIdx    int               `json:"li"`
	BidPlayerIdx     int               `json:"bi"`
	CurrentBid       int               `json:"cb"`
	BidWinnerIdx     int               `json:"bw"`
	TrumpSuit        int               `json:"ts"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	LastDealDetail   *CinchDealDetail  `json:"ld"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Cinch) MarshalJSON() ([]byte, error) {
	return json.Marshal(cinchJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		DealerIdx:        g.dealerIdx,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		LastTrick:        g.lastTrick,
		LastTrickWinner:  g.lastTrickWinner,
		LeadPlayerIdx:    g.leadPlayerIdx,
		BidPlayerIdx:     g.bidPlayerIdx,
		CurrentBid:       g.currentBid,
		BidWinnerIdx:     g.bidWinnerIdx,
		TrumpSuit:        g.trumpSuit,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		LastDealDetail:   g.lastDealDetail,
		ActionLog:        g.actionLog,
	})
}

// cinchMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const cinchMaxSliceLen = 1000

// cinchInRange reports whether v is in [0, CinchPlayerCnt).
func cinchInRange(v int) bool { return v >= 0 && v < CinchPlayerCnt }

// cinchInRangeOrUnset reports whether v is -1 (unset) or in [0, CinchPlayerCnt).
func cinchInRangeOrUnset(v int) bool { return v == -1 || cinchInRange(v) }

// cinchValidateTrick は復元したトリック配列の各要素を検証する。
func cinchValidateTrick(trick []*TrickCard) error {
	for _, tc := range trick {
		if tc == nil || tc.Card == nil {
			return fmt.Errorf("cinch: invalid trick card")
		}
		if !cinchInRange(tc.PlayerIdx) {
			return fmt.Errorf("cinch: trick card player index out of range")
		}
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Cinch) UnmarshalJSON(data []byte) error {
	var j cinchJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > cinchMaxSliceLen || len(j.CurrentTrick) > cinchMaxSliceLen ||
		len(j.LastTrick) > cinchMaxSliceLen || len(j.ActionLog) > cinchMaxSliceLen {
		return fmt.Errorf("cinch: input array exceeds maximum allowed size")
	}
	if len(j.Players) != CinchPlayerCnt {
		return fmt.Errorf("cinch: invalid player count %d, expected %d", len(j.Players), CinchPlayerCnt)
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("cinch: nil player in state")
		}
	}
	if err := cinchValidateTrick(j.CurrentTrick); err != nil {
		return err
	}
	if err := cinchValidateTrick(j.LastTrick); err != nil {
		return err
	}
	// フェーズ検証。
	switch j.Phase {
	case CinchPhaseBid, CinchPhaseNameTrump, CinchPhasePlay, CinchPhaseTrickEnd,
		CinchPhaseRoundEnd, CinchPhaseGameEnd:
	default:
		return fmt.Errorf("cinch: invalid phase %d", j.Phase)
	}
	// ビッド値の範囲 (0=pass .. CinchMaxBid)。
	if j.CurrentBid < 0 || j.CurrentBid > CinchMaxBid {
		return fmt.Errorf("cinch: invalid current bid %d", j.CurrentBid)
	}
	// 切り札スート: CinchTrumpUnset(0) 許容、それ以外は [Spade, Diamond]。
	if j.TrumpSuit != CinchTrumpUnset && (j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return fmt.Errorf("cinch: invalid trump suit %d", j.TrumpSuit)
	}
	// 常に範囲内であるべきインデックス。
	if !cinchInRange(j.DealerIdx) || !cinchInRange(j.BidPlayerIdx) {
		return fmt.Errorf("cinch: index field out of range")
	}
	// -1 (未確定) 許容のインデックス。
	if !cinchInRangeOrUnset(j.CurrentPlayerIdx) || !cinchInRangeOrUnset(j.LeadPlayerIdx) ||
		!cinchInRangeOrUnset(j.BidWinnerIdx) || !cinchInRangeOrUnset(j.LastTrickWinner) ||
		!cinchInRangeOrUnset(j.WinnerIdx) {
		return fmt.Errorf("cinch: sentinel index field out of range")
	}
	// フェーズが play 以降では切り札とビッダーが確定していなければならない。
	if j.Phase == CinchPhasePlay || j.Phase == CinchPhaseTrickEnd || j.Phase == CinchPhaseRoundEnd {
		if j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond {
			return fmt.Errorf("cinch: trump must be set once play begins")
		}
		if !cinchInRange(j.BidWinnerIdx) {
			return fmt.Errorf("cinch: bid winner must be set once play begins")
		}
		if !cinchInRange(j.CurrentPlayerIdx) || !cinchInRange(j.LeadPlayerIdx) {
			return fmt.Errorf("cinch: play indices must be set once play begins")
		}
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newCinchDeck()
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.dealerIdx = j.DealerIdx
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.lastTrick = j.LastTrick
	g.lastTrickWinner = j.LastTrickWinner
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.bidPlayerIdx = j.BidPlayerIdx
	g.currentBid = j.CurrentBid
	g.bidWinnerIdx = j.BidWinnerIdx
	g.trumpSuit = j.TrumpSuit
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.lastDealDetail = j.LastDealDetail
	if j.ActionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	} else {
		g.actionLog = j.ActionLog
	}
	return nil
}

// cinchDealDetailJSON is the JSON wire format for CinchDealDetail.
type cinchDealDetailJSON struct {
	TrumpSuit int         `json:"ts"`
	BidderIdx int         `json:"bi"`
	Bid       int         `json:"bd"`
	SetBack   bool        `json:"sb"`
	Points    map[int]int `json:"pt"`
	Gained    map[int]int `json:"gn"`
}

// MarshalJSON implements json.Marshaler.
func (d *CinchDealDetail) MarshalJSON() ([]byte, error) {
	return json.Marshal(cinchDealDetailJSON{
		TrumpSuit: d.TrumpSuit,
		BidderIdx: d.BidderIdx,
		Bid:       d.Bid,
		SetBack:   d.SetBack,
		Points:    d.Points,
		Gained:    d.Gained,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *CinchDealDetail) UnmarshalJSON(data []byte) error {
	var j cinchDealDetailJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	d.TrumpSuit = j.TrumpSuit
	d.BidderIdx = j.BidderIdx
	d.Bid = j.Bid
	d.SetBack = j.SetBack
	d.Points = j.Points
	d.Gained = j.Gained
	return nil
}
