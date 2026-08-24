//go:build !js || !wasm || extra

// Package domain グリーク (Gleek) のドメインモデル。
//
// Gleek は 16〜17 世紀イングランドで遊ばれた 3 人用のゲームで、**1 ディールの中に
// 得点段階が 4 つ**あるのが特徴。既存の `piquet` (2 人・メルド + トリック) や
// `skat` (3 人・競り + スカート交換 + トリック) がそれぞれ一部だけ持っている構造を、
// 1 ディールに全部並べたものと見ると近い。
//
// デッキ: 44 枚 = 標準 52 枚から 2 と 3 を除いたもの (各スート A,K,Q,J,10..4)。
// 12 枚ずつ配って 36 枚、残り 8 枚がストック。**ストックの一番上を表にして切り札**を
// 決める (めくれたのが 4 = Tiddy ならディーラーが各相手から 4 点貰う)。
//
// 段階 1 — ストックの競り:
//
//	エルダー (ディーラーの左隣) は **12 から始めなければならず、降りられない**。
//	以降は 2 刻みで競り上げるか降りるか。最高額の席が**その半額を各相手に払い**、
//	表向きの札を除く 7 枚を手に入れて 7 枚捨てる。半額が整数になるよう刻みは 2。
//
// 段階 2 — ラフ (ruff):
//
//	同一スートに固めた札の合計 (A=11, K/Q/J=10, 数札は額面) が最も高い席が、
//	各相手から掛け金を取る。**エース 4 枚のマーニヴァルはどんなラフにも勝つ。**
//
// 段階 3 — グリークとマーニヴァル:
//
//	A/K/Q/J の同ランク 3 枚が「グリーク」、4 枚が「マーニヴァル」。各相手から
//	A は 4 / 8、K は 3 / 6、Q は 2 / 4、J は 1 / 2 を取る。
//
// 段階 4 — 12 トリック:
//
//	エルダーがリード。マストフォロー。切り札が最強。1 トリック 3 点に加えて、
//	切り札の名札が点を持つ —— Tib (A) 15、Tom (J) 9、Tumbler (6) 6、Towser (5) 5、
//	Tiddy (4) 4、切り札の K と Q が 3 ずつ。
//
// **精算の基準点はデッキから導く。** 12 × 3 = 36 点に名札 45 点で 1 ディールに
// ちょうど 81 点あり、3 人で割ると 27。各席は自分の合計と 27 の差を受け取る/払う
// ので、この段階もゼロ和になる。(Parlett は 22 と書くが、どの名札の組み合わせでも
// 3 × 22 = 66 にはならない。書き写さずに数える。)
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// GleekPlayerCnt プレイヤー数 (人間 1 + CPU 2)
const GleekPlayerCnt = 3

// GleekHandSize 各プレイヤーの配り札枚数
const GleekHandSize = 12

// GleekTrickCount 1 ディールのトリック数 (手札 12 枚を出し切る)
const GleekTrickCount = 12

// GleekDeckSize デッキ枚数 (52 - 2,3)
const GleekDeckSize = 44

// GleekStockSize 配り切らずに残るストックの枚数
const GleekStockSize = GleekDeckSize - GleekHandSize*GleekPlayerCnt

// GleekSwapSize 落札者が入れ替える枚数 (ストックから表向きの 1 枚を除いた分)
const GleekSwapSize = GleekStockSize - 1

// GleekWinRounds マッチを構成するディール数 (既定)
const GleekWinRounds = 5

// 競りの刻みと上限。
const (
	// GleekMinBid エルダーが必ず置く最初の額。降りることは許されない。
	GleekMinBid = 12
	// GleekBidStep 競り上げの刻み。**2 なのは半額を各相手に払うから。**
	// 1 刻みにすると 13 の半分が 6.5 になり、どこかで端数を丸めた分だけ
	// 卓の点が増減する。
	GleekBidStep = 2
	// GleekMaxBid 競りの上限。**上限が無いと CPU 同士が延々と competing し、
	// ディールが終わらない。**
	GleekMaxBid = 24
)

// GleekRuffStake ラフの勝者が各相手から取る点。
const GleekRuffStake = 2

// GleekTiddyTurnUpBonus めくれた切り札が 4 (Tiddy) のときディーラーが各相手から取る点。
const GleekTiddyTurnUpBonus = 4

// GleekTrickPoints 1 トリックの点。
const GleekTrickPoints = 3

// GleekPhase ゲームフェーズ
type GleekPhase int

// Gleek のフェーズ定数
const (
	// GleekPhaseBid ストックの競りフェーズ
	GleekPhaseBid GleekPhase = 0
	// GleekPhaseDiscard 落札者がストックを取り込み 7 枚捨てるフェーズ
	GleekPhaseDiscard GleekPhase = 1
	// GleekPhasePlay トリックプレイフェーズ
	GleekPhasePlay GleekPhase = 2
	// GleekPhaseTrickEnd トリック終了フェーズ
	GleekPhaseTrickEnd GleekPhase = 3
	// GleekPhaseRoundEnd ディール終了フェーズ
	GleekPhaseRoundEnd GleekPhase = 4
	// GleekPhaseGameEnd ゲーム終了フェーズ
	GleekPhaseGameEnd GleekPhase = 5
)

// GleekPhaseMin フェーズ下限 (検証用)
const GleekPhaseMin = int(GleekPhaseBid)

// GleekPhaseMax フェーズ上限 (検証用)
const GleekPhaseMax = int(GleekPhaseGameEnd)

// GleekResult 人間視点のマッチ結果
type GleekResult int

// Gleek のマッチ結果定数
const (
	// GleekResultLose 敗北
	GleekResultLose GleekResult = -1
	// GleekResultNone 未確定 / 引き分け
	GleekResultNone GleekResult = 0
	// GleekResultWin 勝利
	GleekResultWin GleekResult = 1
)

// GleekMeld 申告されたグリーク / マーニヴァル 1 件。
type GleekMeld struct {
	// PlayerIdx 申告した席
	PlayerIdx int `json:"p"`
	// Rank 同ランクの札位 (1=A, 13=K, 12=Q, 11=J)
	Rank int `json:"r"`
	// Count 枚数 (3=gleek, 4=mournival)
	Count int `json:"c"`
	// Value 各相手から取る点
	Value int `json:"v"`
}

// GleekRuff 1 席のラフ (同一スートの最高合計)。
type GleekRuff struct {
	// PlayerIdx 席
	PlayerIdx int `json:"p"`
	// Suit 最も高くなったスート
	Suit int `json:"s"`
	// Total そのスートの合計 (A=11, K/Q/J=10, 数札は額面)
	Total int `json:"t"`
}

// GleekHint ヒント情報
type GleekHint struct {
	// CardIndices 推奨カードインデックス (play フェーズ / discard フェーズ)
	CardIndices []int
	// Bid 推奨する競り額 (0=降りる)
	Bid int
	// Reason ヒント理由キー
	Reason string
}

// Gleek グリークのゲームクラス
type Gleek struct {
	trumpCards       *TrumpCards
	players          []*GleekPlayer
	config           GleekConfig
	phase            GleekPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	trickResolved    bool
	leadPlayerIdx    int
	dealerIdx        int
	elderIdx         int // ディーラーの左隣。競りを開き、最初のトリックをリードする
	trumpSuit        int // 表向きの札で決まる切り札 (-1=未確定)
	turnUp           *Card

	// 競り
	currentBidderIdx int
	bids             [GleekPlayerCnt]int
	passed           [GleekPlayerCnt]bool
	buyerIdx         int // 落札者 (-1=未確定)
	winningBid       int

	// ストック
	stock []*Card

	// 段階ごとの得点
	ruffs           []*GleekRuff
	ruffWinnerIdx   int
	melds           []*GleekMeld
	trickPoints     [GleekPlayerCnt]int // 3 点 × トリック数 + 取った名札の点
	playerScores    [GleekPlayerCnt]int // 累積ゲーム点
	lastTrickWinner int

	result       GleekResult
	scored       bool
	gameEndFlag  bool
	winnerPlayer int
	actionLogBase
}

// NewGleek コンストラクタ
func NewGleek(trumpCards *TrumpCards, players []*GleekPlayer, config GleekConfig) *Gleek {
	return &Gleek{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerPlayer:    -1,
		lastTrickWinner: -1,
		buyerIdx:        -1,
		ruffWinnerIdx:   -1,
		trumpSuit:       -1,
	}
}

// NewDefaultGleek 標準の 3 人構成 (人間 1, CPU 2) と既定設定で生成する。
func NewDefaultGleek() *Gleek {
	players := make([]*GleekPlayer, GleekPlayerCnt)
	players[0] = NewGleekPlayer(true)
	for i := 1; i < GleekPlayerCnt; i++ {
		players[i] = NewGleekPlayer(false)
	}
	return NewGleek(newGleekDeck(), players, DefaultGleekConfig())
}

// newGleekDeck Gleek 用 44 枚デッキを生成する。標準 52 枚から 2 と 3 を除外する。
func newGleekDeck() *TrumpCards {
	full := NewTrumpCards(0)
	t := new(TrumpCards)
	t.deck = make([]*Card, 0, GleekDeckSize)
	for _, c := range full.deck {
		if v := c.GetValue(); v == 2 || v == 3 {
			continue
		}
		t.deck = append(t.deck, NewCard(c.GetDesign(), c.GetValue(), false))
	}
	t.deckCnt = len(t.deck)
	t.deckInit()
	return t
}

// Reset ゲーム初期化
func (g *Gleek) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	// **ディーラーは人間の右隣。** エルダー (dealer+1) が競りを開き最初のトリックを
	// リードするので、ここを最終席にしておくと人間がエルダーになる。dealerIdx=0 だと
	// 人間は毎回最後に喋る番になる。
	g.dealerIdx = GleekPlayerCnt - 1
	g.playerScores = [GleekPlayerCnt]int{}
	g.result = GleekResultNone
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のディールを開始する
func (g *Gleek) NextRound() {
	if g.phase != GleekPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % GleekPlayerCnt
	g.startRound()
}

// startRound 手札を配り、切り札をめくって競りフェーズを開始する。
func (g *Gleek) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.trickResolved = false
	g.lastTrickWinner = -1
	g.buyerIdx = -1
	g.winningBid = 0
	g.trumpSuit = -1
	g.turnUp = nil
	g.stock = nil
	g.ruffs = nil
	g.ruffWinnerIdx = -1
	g.melds = nil
	g.trickPoints = [GleekPlayerCnt]int{}
	g.scored = false
	g.bids = [GleekPlayerCnt]int{}
	g.passed = [GleekPlayerCnt]bool{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()

	g.elderIdx = (g.dealerIdx + 1) % GleekPlayerCnt
	g.sortAllHands()
	g.openAuction()
}

// deal 各プレイヤーへ GleekHandSize 枚を配り、残りをストックにする。
// ストックの一番上を表にして切り札を決める。
func (g *Gleek) deal() {
	for i := 0; i < GleekHandSize; i++ {
		for j := 0; j < GleekPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % GleekPlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
	g.stock = make([]*Card, 0, GleekStockSize)
	for i := 0; i < GleekStockSize; i++ {
		if c := g.trumpCards.DrawCard(); c != nil {
			g.stock = append(g.stock, c)
		}
	}
	if len(g.stock) > 0 {
		g.turnUp = g.stock[0]
		g.trumpSuit = g.turnUp.GetDesign()
	}
}

// openAuction 競りを開く。
//
// **エルダーの 12 は決定ではないので自動で置く。** 「降りられない最低額」を人間に
// 押させても選択肢が 1 つしかなく、押すまで盤が進まないだけになる。競り上げるか
// 降りるかという本当の判断は、一周して戻ってきたときに残る。
func (g *Gleek) openAuction() {
	g.phase = GleekPhaseBid
	g.bids[g.elderIdx] = GleekMinBid
	g.appendLog(g.elderIdx, "bid_open",
		fmt.Sprintf("%s must open at %d", playerName(g.players, g.elderIdx), GleekMinBid), nil)
	if g.turnUp != nil {
		g.appendLog(-1, "turn_up",
			fmt.Sprintf("turn-up %s sets trump to %s", cardStr(g.turnUp), gleekSuitName(g.trumpSuit)), []*Card{g.turnUp})
		g.payTiddyTurnUp()
	}
	g.currentBidderIdx = g.nextBidder(g.elderIdx)
}

// payTiddyTurnUp めくれた切り札が 4 (Tiddy) ならディーラーが各相手から点を取る。
func (g *Gleek) payTiddyTurnUp() {
	if g.turnUp == nil || g.turnUp.GetValue() != gleekTiddyValue {
		return
	}
	for i := 0; i < GleekPlayerCnt; i++ {
		if i == g.dealerIdx {
			g.playerScores[i] += GleekTiddyTurnUpBonus * (GleekPlayerCnt - 1)
			continue
		}
		g.playerScores[i] -= GleekTiddyTurnUpBonus
	}
	g.appendLog(g.dealerIdx, "tiddy_turn_up",
		fmt.Sprintf("Tiddy turned up: %s takes %d from each opponent",
			playerName(g.players, g.dealerIdx), GleekTiddyTurnUpBonus), nil)
}

// --- Bidding ---

// PlayerBid 人間が競る。bid=0 は降りる、それ以外は現在の最高額を上回る額。
func (g *Gleek) PlayerBid(bid int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GleekPhaseBid {
		return ErrWrongPhase
	}
	if !g.players[g.currentBidderIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if bid != 0 && bid != g.NextBidAmount() {
		return NewDomainError(ErrInvalidPlay, "競り上げる額は現在の最高額に刻みを足した値です")
	}
	g.applyBid(g.currentBidderIdx, bid)
	return nil
}

// CpuBid 現在の競り手番が CPU の場合に 1 回行動する。
func (g *Gleek) CpuBid() {
	if g.gameEndFlag || g.phase != GleekPhaseBid {
		return
	}
	idx := g.currentBidderIdx
	if idx < 0 || g.players[idx].GetIsHuman() {
		return
	}
	g.applyBid(idx, g.cpuChooseBid(idx))
}

// NextBidAmount 次に置ける額を返す (0=これ以上競り上げられない)。
func (g *Gleek) NextBidAmount() int {
	next := g.HighestBid() + GleekBidStep
	if next > GleekMaxBid {
		return 0
	}
	return next
}

// HighestBid 現在の最高額を返す。
func (g *Gleek) HighestBid() int {
	best := 0
	for _, b := range g.bids {
		if b > best {
			best = b
		}
	}
	return best
}

// highestBidderIdx 現在の最高額を置いていて**まだ降りていない**席を返す。
//
// **降りた席の額は残る。** エルダーの 12 は自動で置かれるので、`bids` を見るだけだと
// 降りたエルダーが最高額のままストックを買ってしまう盤面が作れる。降りた席を
// 除いて数えるのは、その一手を不可能にするため。
func (g *Gleek) highestBidderIdx() int {
	best, idx := 0, -1
	for i := 0; i < GleekPlayerCnt; i++ {
		cand := (g.elderIdx + i) % GleekPlayerCnt
		if g.passed[cand] {
			continue
		}
		if g.bids[cand] > best {
			best = g.bids[cand]
			idx = cand
		}
	}
	return idx
}

// applyBid 競りの 1 手を適用し、残り 1 人になったら締める。
func (g *Gleek) applyBid(playerIdx, bid int) {
	if bid == 0 {
		g.passed[playerIdx] = true
		g.appendLog(playerIdx, "bid_pass",
			fmt.Sprintf("%s drops out", playerName(g.players, playerIdx)), nil)
	} else {
		g.bids[playerIdx] = bid
		g.appendLog(playerIdx, "bid",
			fmt.Sprintf("%s bids %d", playerName(g.players, playerIdx), bid), nil)
	}

	if g.activeBidders() <= 1 || g.NextBidAmount() == 0 {
		g.finalizeAuction()
		return
	}
	g.currentBidderIdx = g.nextBidder(playerIdx)
}

// activeBidders まだ降りていない席の数。
func (g *Gleek) activeBidders() int {
	n := 0
	for _, p := range g.passed {
		if !p {
			n++
		}
	}
	return n
}

// nextBidder playerIdx の次でまだ降りていない席を返す。
func (g *Gleek) nextBidder(playerIdx int) int {
	for i := 1; i <= GleekPlayerCnt; i++ {
		cand := (playerIdx + i) % GleekPlayerCnt
		if !g.passed[cand] {
			return cand
		}
	}
	return playerIdx
}

// finalizeAuction 落札者を確定し、半額を各相手に払わせてストックを渡す。
func (g *Gleek) finalizeAuction() {
	buyer := g.highestBidderIdx()
	if buyer < 0 {
		// **エルダーは降りられないので、ここには来ない。** 復元した壊れた盤で
		// だけ起きうるので、卓が止まらないようエルダーに引き受けさせる。
		buyer = g.elderIdx
		g.bids[buyer] = GleekMinBid
	}
	g.buyerIdx = buyer
	g.winningBid = g.bids[buyer]

	// **半額を「各相手に」払う。** 落札額そのものを 1 回払う実装にすると、
	// 3 人卓で動く点が半分になり、競り上げの重みが変わる。刻みが 2 なので
	// 半額は必ず整数 —— 端数を丸めた分だけ卓の点が増減することはない。
	half := g.winningBid / 2
	for i := 0; i < GleekPlayerCnt; i++ {
		if i == buyer {
			g.playerScores[i] -= half * (GleekPlayerCnt - 1)
			continue
		}
		g.playerScores[i] += half
	}
	g.appendLog(buyer, "buy_stock",
		fmt.Sprintf("%s buys the stock for %d, paying %d to each opponent",
			playerName(g.players, buyer), g.winningBid, half), nil)

	g.giveStockToBuyer()
	g.phase = GleekPhaseDiscard
	g.currentPlayerIdx = buyer
	g.sortAllHands()
}

// giveStockToBuyer 表向きの札を除くストックを落札者の手札に加える。
func (g *Gleek) giveStockToBuyer() {
	if g.buyerIdx < 0 {
		return
	}
	taken := make([]*Card, 0, GleekSwapSize)
	for i, c := range g.stock {
		if i == 0 {
			continue // 表向きの札は切り札を決めるために場に残る
		}
		taken = append(taken, c)
		g.players[g.buyerIdx].AddCard(c)
	}
	g.stock = g.stock[:1]
	g.appendLog(g.buyerIdx, "take_stock",
		fmt.Sprintf("%s takes %d cards from the stock", playerName(g.players, g.buyerIdx), len(taken)), taken)
}

// --- Discard ---

// PlayerDiscard 落札者が捨てる札をまとめて指定する。ちょうど GleekSwapSize 枚。
func (g *Gleek) PlayerDiscard(indices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GleekPhaseDiscard {
		return ErrWrongPhase
	}
	if g.buyerIdx < 0 || !g.players[g.buyerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if err := g.validateDiscard(indices); err != nil {
		return err
	}
	g.applyDiscard(g.buyerIdx, indices)
	return nil
}

// validateDiscard 捨て札の指定を検証する。
func (g *Gleek) validateDiscard(indices []int) error {
	if len(indices) != GleekSwapSize {
		return NewDomainError(ErrInvalidPlay,
			fmt.Sprintf("捨てる札をちょうど %d 枚選んでください", GleekSwapSize))
	}
	size := g.players[g.buyerIdx].GetCardsSize()
	seen := map[int]bool{}
	for _, idx := range indices {
		if idx < 0 || idx >= size {
			return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidPlay, "同じ札を二度選べません")
		}
		seen[idx] = true
	}
	return nil
}

// applyDiscard 捨て札を落札者の手から取り除き、プレイフェーズへ進む。
func (g *Gleek) applyDiscard(playerIdx int, indices []int) {
	sorted := append([]int{}, indices...)
	// **大きい索引から抜く。** 小さい方から抜くと後ろの索引がずれて、
	// 指定したのと違う札が捨てられる。
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	discarded := make([]*Card, 0, len(sorted))
	for _, idx := range sorted {
		if c := g.players[playerIdx].RemoveCard(idx); c != nil {
			discarded = append(discarded, c)
		}
	}
	g.appendLog(playerIdx, "discard",
		fmt.Sprintf("%s discards %d cards", playerName(g.players, playerIdx), len(discarded)), discarded)
	g.startPlay()
}

// CpuDiscard CPU の落札者に捨て札を選ばせる。
func (g *Gleek) CpuDiscard() {
	if g.phase != GleekPhaseDiscard {
		return
	}
	idx := g.buyerIdx
	if idx < 0 || g.players[idx].GetIsHuman() {
		return
	}
	g.applyDiscard(idx, g.cpuSelectDiscards(idx))
}

// GetDiscardHint は捨てるべき札の索引を返す (画面のヒント)。
func (g *Gleek) GetDiscardHint() []int {
	if g.phase != GleekPhaseDiscard || g.buyerIdx < 0 {
		return nil
	}
	return g.cpuSelectDiscards(g.buyerIdx)
}

// cpuSelectDiscards 残す 12 枚を決め、それ以外の索引を返す。
//
// **残す価値でソートして上から 12 枚。** 切り札と名札、そして一番長い平札スートを
// 残し、他スートの低い札から捨てる。
func (g *Gleek) cpuSelectDiscards(playerIdx int) []int {
	p := g.players[playerIdx]
	size := p.GetCardsSize()
	idx := make([]int, 0, size)
	for i := 0; i < size; i++ {
		idx = append(idx, i)
	}
	suitCount := map[int]int{}
	for i := 0; i < size; i++ {
		if c := p.GetCard(i); c != nil {
			suitCount[c.GetDesign()]++
		}
	}
	keepScore := func(i int) int {
		c := p.GetCard(i)
		if c == nil {
			return -1
		}
		score := gleekRuffValue(c.GetValue()) + suitCount[c.GetDesign()]
		if c.GetDesign() == g.trumpSuit {
			score += 20
		}
		score += gleekHonourValue(c, g.trumpSuit)
		return score
	}
	sort.SliceStable(idx, func(a, b int) bool { return keepScore(idx[a]) > keepScore(idx[b]) })
	if len(idx) <= GleekHandSize {
		return nil
	}
	drop := append([]int{}, idx[GleekHandSize:]...)
	sort.Ints(drop)
	return drop
}

// --- Ruff and melds ---

// startPlay 捨て札後、ラフとメルドを精算してからトリックプレイを開始する。
func (g *Gleek) startPlay() {
	g.scoreRuff()
	g.scoreMelds()
	g.sortAllHands()
	g.leadPlayerIdx = g.elderIdx
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber = 1
	g.currentTrick = nil
	g.trickResolved = false
	g.phase = GleekPhasePlay
}

// scoreRuff 各席のラフを求め、最高の席が各相手から掛け金を取る。
func (g *Gleek) scoreRuff() {
	g.ruffs = make([]*GleekRuff, 0, GleekPlayerCnt)
	for i := 0; i < GleekPlayerCnt; i++ {
		suit, total := gleekBestRuff(g.players[i])
		g.ruffs = append(g.ruffs, &GleekRuff{PlayerIdx: i, Suit: suit, Total: total})
	}

	winner := -1
	best := -1
	for i := 0; i < GleekPlayerCnt; i++ {
		cand := (g.elderIdx + i) % GleekPlayerCnt
		if g.ruffs[cand].Total > best {
			best = g.ruffs[cand].Total
			winner = cand
		}
	}
	// **エースのマーニヴァルはどんなラフにも勝つ。** 4 枚とも持っていれば
	// 同一スートに固まっていなくてもラフを取る。
	for i := 0; i < GleekPlayerCnt; i++ {
		cand := (g.elderIdx + i) % GleekPlayerCnt
		if gleekCountRank(g.players[cand], gleekAceValue) == 4 {
			winner = cand
			break
		}
	}
	if winner < 0 {
		return
	}
	g.ruffWinnerIdx = winner
	for i := 0; i < GleekPlayerCnt; i++ {
		if i == winner {
			g.playerScores[i] += GleekRuffStake * (GleekPlayerCnt - 1)
			continue
		}
		g.playerScores[i] -= GleekRuffStake
	}
	g.appendLog(winner, "ruff",
		fmt.Sprintf("%s wins the ruff with %d in %s",
			playerName(g.players, winner), g.ruffs[winner].Total, gleekSuitName(g.ruffs[winner].Suit)), nil)
}

// scoreMelds グリークとマーニヴァルを申告し、各相手から点を取る。
func (g *Gleek) scoreMelds() {
	g.melds = nil
	for i := 0; i < GleekPlayerCnt; i++ {
		seat := (g.elderIdx + i) % GleekPlayerCnt
		for _, rank := range gleekMeldRanks {
			n := gleekCountRank(g.players[seat], rank)
			if n < 3 {
				continue
			}
			value := gleekMeldValue(rank, n)
			g.melds = append(g.melds, &GleekMeld{PlayerIdx: seat, Rank: rank, Count: n, Value: value})
			for j := 0; j < GleekPlayerCnt; j++ {
				if j == seat {
					g.playerScores[j] += value * (GleekPlayerCnt - 1)
					continue
				}
				g.playerScores[j] -= value
			}
			g.appendLog(seat, "meld",
				fmt.Sprintf("%s declares a %s of %s worth %d from each opponent",
					playerName(g.players, seat), gleekMeldName(n), gleekRankName(rank), value), nil)
		}
	}
}

// --- Play ---

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Gleek) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GleekPhasePlay {
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
	g.playCard(g.currentPlayerIdx, player.RemoveCard(cardIndex))
	return nil
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する。
func (g *Gleek) CpuPlay() {
	if g.gameEndFlag || g.phase != GleekPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	played := g.players[idx].RemoveCard(g.cpuSelectPlayCard(idx))
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// playCard カードをプレイする共通処理。
func (g *Gleek) playCard(playerIdx int, card *Card) {
	if card == nil {
		return
	}
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == GleekPlayerCnt {
		g.phase = GleekPhaseTrickEnd
		return
	}
	g.currentPlayerIdx = (g.currentPlayerIdx + 1) % GleekPlayerCnt
}

// ResolveTrick トリックを解決して勝者を決定する。
func (g *Gleek) ResolveTrick() {
	if g.phase != GleekPhaseTrickEnd || len(g.currentTrick) != GleekPlayerCnt {
		return
	}
	// **同じトリックを二度精算しない。** トリック終了フェーズは NextTrick まで
	// 続くので、二度呼ぶと勝者に同じ札束が二度積まれる。
	if g.trickResolved {
		return
	}
	g.trickResolved = true

	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, 0, len(g.currentTrick))
	points := GleekTrickPoints
	for _, tc := range g.currentTrick {
		trickCards = append(trickCards, tc.Card)
		points += gleekHonourValue(tc.Card, g.trumpSuit)
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.trickPoints[winnerIdx] += points
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d for %d", playerName(g.players, winnerIdx), g.trickNumber, points), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= GleekTrickCount {
		g.lastTrickWinner = winnerIdx
		g.phase = GleekPhaseRoundEnd
		g.enterRoundEnd()
		return
	}
	g.phase = GleekPhaseTrickEnd
}

// NextTrick 次のトリックを開始する。
func (g *Gleek) NextTrick() {
	if g.phase != GleekPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.trickResolved = false
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = GleekPhasePlay
}

// enterRoundEnd RoundEnd 突入時に一度だけ精算する (scored フラグでガード)。
func (g *Gleek) enterRoundEnd() {
	if g.scored {
		return
	}
	g.scored = true
	g.applyTrickSettlement()
	g.checkGameEnd()
}

// DealPoints はこのディールで実際に配られた点の合計を返す。
//
// **上限 (GleekMaxDealPoints) と一致するとは限らない。** 表向きの札と落札者の
// 捨て札は卓に出ないので、そこに名札が入ればその分だけ小さくなる。
func (g *Gleek) DealPoints() int {
	total := 0
	for _, v := range g.trickPoints {
		total += v
	}
	return total
}

// Par は精算の基準点 —— このディールに実際にあった点を人数で割った値を返す。
//
// **基準点は書き写さずにそのディールから数える。** 固定値にすると、名札が場外に
// 落ちたディールでは全員が基準点に届かず、卓から点が消え続ける (Parlett の 22 も
// どの名札の組み合わせでも 3 倍にならない)。
func (g *Gleek) Par() int { return g.DealPoints() / GleekPlayerCnt }

// applyTrickSettlement 各席のトリック点と基準点の差を精算する。
func (g *Gleek) applyTrickSettlement() {
	total := g.DealPoints()
	par := g.Par()
	for i := 0; i < GleekPlayerCnt; i++ {
		delta := g.trickPoints[i] - par
		g.playerScores[i] += delta
		g.appendLog(i, "round_score",
			fmt.Sprintf("%s takes %d of %d (par %d) for %+d",
				playerName(g.players, i), g.trickPoints[i], total, par, delta), nil)
	}
}

// ScoreRound RoundEnd フェーズでの得点計算を行う (idempotent、インタフェース互換)。
func (g *Gleek) ScoreRound() {
	if g.phase != GleekPhaseRoundEnd {
		return
	}
	g.enterRoundEnd()
}

// checkGameEnd 規定ディール数を終えたらマッチ終了を判定する。
func (g *Gleek) checkGameEnd() {
	if g.roundNumber < g.config.TargetRounds {
		return
	}
	leader, best := 0, g.playerScores[0]
	tie := false
	for i := 1; i < GleekPlayerCnt; i++ {
		if g.playerScores[i] > best {
			best = g.playerScores[i]
			leader = i
			tie = false
		} else if g.playerScores[i] == best {
			tie = true
		}
	}
	g.gameEndFlag = true
	g.winnerPlayer = leader
	g.phase = GleekPhaseGameEnd
	g.result = g.humanResult(leader, tie)
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", playerName(g.players, leader)), nil)
}

// humanResult 人間 (seat 0) の視点でマッチ結果を返す。
func (g *Gleek) humanResult(leader int, tie bool) GleekResult {
	human := findHumanIdx(g.players)
	if human < 0 {
		return GleekResultNone
	}
	if g.playerScores[human] == g.playerScores[leader] {
		if tie {
			return GleekResultNone
		}
		return GleekResultWin
	}
	return GleekResultLose
}

// --- Trick helpers ---

// validatePlay マストフォローを検証する。
func (g *Gleek) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 || card == nil {
		return nil
	}
	lead := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != lead && g.playerHasSuit(playerIdx, lead) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーが指定スートを持っているか。
func (g *Gleek) playerHasSuit(playerIdx, suit int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if c := p.GetCard(i); c != nil && c.GetDesign() == suit {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する。切り札が最強、無ければリードスートの最強札。
func (g *Gleek) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	lead := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerStr := gleekCardStrength(g.currentTrick[0].Card, g.trumpSuit, lead)
	for _, tc := range g.currentTrick[1:] {
		if s := gleekCardStrength(tc.Card, g.trumpSuit, lead); s > winnerStr {
			winnerIdx = tc.PlayerIdx
			winnerStr = s
		}
	}
	return winnerIdx
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Gleek) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// --- Card ranking ---

// 札位の別名。この実装のカード値は A=1, J=11, Q=12, K=13。
const (
	gleekAceValue   = 1
	gleekJackValue  = 11
	gleekQueenValue = 12
	gleekKingValue  = 13
	gleekTiddyValue = 4
)

// gleekMeldRanks グリーク / マーニヴァルの対象となるランク (10 より上)。
var gleekMeldRanks = []int{gleekAceValue, gleekKingValue, gleekQueenValue, gleekJackValue}

// gleekTrumpBoost 切り札に足す下駄。平札の最大 (11) を必ず上回る。
const gleekTrumpBoost = 100

// gleekCardStrength trump / リードスートを踏まえたカードの強さ。
// 場外の平札 (リードでも切り札でもないスート) は勝てないので 0。
func gleekCardStrength(c *Card, trump, lead int) int {
	if c == nil {
		return 0
	}
	switch c.GetDesign() {
	case trump:
		return gleekTrumpBoost + gleekRankStrength(c.GetValue())
	case lead:
		return gleekRankStrength(c.GetValue())
	default:
		return 0
	}
}

// gleekRankStrength スート内の序列 A > K > Q > J > 10 > 9 > 8 > 7 > 6 > 5 > 4。
func gleekRankStrength(v int) int {
	if v == gleekAceValue {
		return 14
	}
	return v
}

// gleekRuffValue ラフの計算に使う札の値 (A=11, K/Q/J=10, 数札は額面)。
func gleekRuffValue(v int) int {
	switch v {
	case gleekAceValue:
		return 11
	case gleekKingValue, gleekQueenValue, gleekJackValue:
		return 10
	default:
		return v
	}
}

// gleekHonourValue 切り札の名札の点。切り札以外は 0。
//
//	Tib (A) 15 / Tom (J) 9 / Tumbler (6) 6 / Towser (5) 5 / Tiddy (4) 4 /
//	切り札の K と Q が 3 ずつ。
func gleekHonourValue(c *Card, trump int) int {
	if c == nil || c.GetDesign() != trump {
		return 0
	}
	switch c.GetValue() {
	case gleekAceValue:
		return 15
	case gleekJackValue:
		return 9
	case 6:
		return 6
	case 5:
		return 5
	case gleekTiddyValue:
		return 4
	case gleekKingValue, gleekQueenValue:
		return 3
	default:
		return 0
	}
}

// GleekHonourTotal 切り札 1 スートに存在する名札の点の合計。
const GleekHonourTotal = 15 + 9 + 6 + 5 + 4 + 3 + 3

// GleekMaxDealPoints 1 ディールで配られうる点の上限 (トリック点 + 名札を全部)。
//
// **実際の合計はこれより小さいことがある。** 表向きの札と落札者が捨てた 7 枚は
// 卓に出ないので、そこに入った名札は誰の点にもならない。基準点をこの上限から
// 決め打ちすると、名札が場外に落ちたディールだけ全員が基準点に届かなくなる。
const GleekMaxDealPoints = GleekTrickCount*GleekTrickPoints + GleekHonourTotal

// gleekBestRuff 手札で最も高くなるスートとその合計を返す。
func gleekBestRuff(p *GleekPlayer) (int, int) {
	totals := map[int]int{}
	for i := 0; i < p.GetCardsSize(); i++ {
		if c := p.GetCard(i); c != nil {
			totals[c.GetDesign()] += gleekRuffValue(c.GetValue())
		}
	}
	bestSuit, best := CardDesignSpade, -1
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if totals[suit] > best {
			best = totals[suit]
			bestSuit = suit
		}
	}
	if best < 0 {
		return CardDesignSpade, 0
	}
	return bestSuit, best
}

// gleekCountRank 手札にある指定ランクの枚数。
func gleekCountRank(p *GleekPlayer, rank int) int {
	n := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		if c := p.GetCard(i); c != nil && c.GetValue() == rank {
			n++
		}
	}
	return n
}

// gleekMeldValue グリーク (3 枚) / マーニヴァル (4 枚) が各相手から取る点。
//
// **マーニヴァルはグリークのちょうど倍。** 表を 2 つ持つと、片方だけ直したときに
// 静かにずれる。
func gleekMeldValue(rank, count int) int {
	base := 0
	switch rank {
	case gleekAceValue:
		base = 4
	case gleekKingValue:
		base = 3
	case gleekQueenValue:
		base = 2
	case gleekJackValue:
		base = 1
	}
	if count >= 4 {
		return base * 2
	}
	return base
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Gleek) sortAllHands() {
	for _, p := range g.players {
		gleekSortHand(p, g.trumpSuit)
	}
}

// gleekSortHand 手札を切り札→スート→強さ順に並べる。
func gleekSortHand(p *GleekPlayer, trump int) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	suitKey := func(c *Card) int {
		if gleekValidSuit(trump) && c.GetDesign() == trump {
			return 0
		}
		return c.GetDesign()
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if si, sj := suitKey(cards[i]), suitKey(cards[j]); si != sj {
			return si < sj
		}
		return gleekRankStrength(cards[i].GetValue()) > gleekRankStrength(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// gleekValidSuit suit が有効なスート (1..4) か。
func gleekValidSuit(suit int) bool {
	return suit >= CardDesignSpade && suit <= CardDesignDiamond
}

// gleekSuitName スートの表示名を返す。
func gleekSuitName(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "spades"
	case CardDesignClover:
		return "clubs"
	case CardDesignHeart:
		return "hearts"
	case CardDesignDiamond:
		return "diamonds"
	default:
		return "-"
	}
}

// gleekRankName メルド対象ランクの表示名を返す。
func gleekRankName(rank int) string {
	switch rank {
	case gleekAceValue:
		return "aces"
	case gleekKingValue:
		return "kings"
	case gleekQueenValue:
		return "queens"
	case gleekJackValue:
		return "jacks"
	default:
		return "-"
	}
}

// gleekMeldName 枚数に対応するメルドの名前を返す。
func gleekMeldName(count int) string {
	if count >= 4 {
		return "mournival"
	}
	return "gleek"
}

// GleekHonourValueForTest はテスト用に名札の点を返す。
func GleekHonourValueForTest(c *Card, trump int) int { return gleekHonourValue(c, trump) }

// StartPlayForTest はテスト用にラフとメルドを精算してプレイフェーズを開始する。
func (g *Gleek) StartPlayForTest() { g.startPlay() }

// SetTrickPointsForTest はテスト用にトリック点を差し込む。
func (g *Gleek) SetTrickPointsForTest(p [GleekPlayerCnt]int) { g.trickPoints = p }

// ScoreMeldsForTest はテスト用にメルドだけを精算する。
func (g *Gleek) ScoreMeldsForTest() { g.scoreMelds() }

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Gleek) indexOfPlayerInTrick(playerIdx int) int {
	return indexOfPlayerInTrick(g.currentTrick, playerIdx)
}

// --- CPU ---

// cpuChooseBid CPU が競り額を決める (0=降りる)。
//
// **手札の価値を超えたら降りる。** 上限まで機械的に競り上げると、ストックを
// 買っただけで基準点に届かない席が毎ディール出る。
func (g *Gleek) cpuChooseBid(playerIdx int) int {
	next := g.NextBidAmount()
	if next == 0 {
		return 0
	}
	if g.config.CpuDifficulty == GleekCpuDifficultyEasy {
		return 0
	}
	if next > g.cpuStockValue(playerIdx) {
		return 0
	}
	return next
}

// cpuStockValue ストックにいくらまで出せるかの見積り。
//
// 手札が弱いほどストックで化ける余地が大きいので、**弱い手ほど高く出す**。
// 切り札を既に持っている席は伸びしろが小さい。
func (g *Gleek) cpuStockValue(playerIdx int) int {
	_, ruff := gleekBestRuff(g.players[playerIdx])
	trumps := 0
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if c := p.GetCard(i); c != nil && c.GetDesign() == g.trumpSuit {
			trumps++
		}
	}
	// ラフ 40 前後・切り札 3 枚が平均。そこから離れるほど評価が動く。
	value := GleekMinBid + (50-ruff)/4 + (3-trumps)*2
	if value < GleekMinBid {
		value = GleekMinBid
	}
	if value > GleekMaxBid {
		value = GleekMaxBid
	}
	return value
}

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Gleek) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == GleekCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 点の付いた札を意識したプレイ。
//
// **Gleek は 3 人とも自分のために取る。** 味方がいないので、勝てるなら取り、
// 勝てないなら名札を捨てない —— 相手のトリックに 15 点の Tib を放り込むのが
// 一番大きな失点になる。
func (g *Gleek) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]

	if len(g.currentTrick) == 0 {
		// リード: 一番強い切り札か、平札の最強札で主導する。
		return g.maxBy(player, valid, func(c *Card) int {
			return gleekCardStrength(c, g.trumpSuit, c.GetDesign())
		})
	}

	lead := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.trickWinner()
	topStr := gleekCardStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card, g.trumpSuit, lead)

	winners := filterIndices(valid, func(idx int) bool {
		return gleekCardStrength(player.GetCard(idx), g.trumpSuit, lead) > topStr
	})
	if len(winners) > 0 {
		// 勝てるなら最小限の勝ち札で取る。
		return g.minBy(player, winners, func(c *Card) int {
			return gleekCardStrength(c, g.trumpSuit, lead)
		})
	}
	// 取れないときは**名札を最優先で温存する**。
	return g.minBy(player, valid, func(c *Card) int {
		return gleekHonourValue(c, g.trumpSuit)*10 + gleekRankStrength(c.GetValue())
	})
}

// minBy score が最小となるインデックスを返す。
func (g *Gleek) minBy(player *GleekPlayer, indices []int, score func(*Card) int) int {
	best := indices[0]
	bestScore := score(player.GetCard(best))
	for _, idx := range indices[1:] {
		if s := score(player.GetCard(idx)); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// maxBy score が最大となるインデックスを返す。
func (g *Gleek) maxBy(player *GleekPlayer, indices []int, score func(*Card) int) int {
	best := indices[0]
	bestScore := score(player.GetCard(best))
	for _, idx := range indices[1:] {
		if s := score(player.GetCard(idx)); s > bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨アクションを返す。
func (g *Gleek) GetHint() *GleekHint {
	human := findHumanIdx(g.players)
	if human < 0 {
		return nil
	}
	switch g.phase {
	case GleekPhaseBid:
		if g.currentBidderIdx != human {
			return nil
		}
		// **CPU と同じ関数を読む。** ヒントだけ別の条件を書くと、助言に従った
		// 人間だけが手札に見合わない額でストックを買う。
		bid := g.cpuChooseBid(human)
		return &GleekHint{Bid: bid, Reason: gleekBidHintReason(bid)}
	case GleekPhaseDiscard:
		if g.buyerIdx != human {
			return nil
		}
		drop := g.cpuSelectDiscards(human)
		if len(drop) == 0 {
			return nil
		}
		return &GleekHint{CardIndices: drop, Reason: "discard_stock"}
	case GleekPhasePlay:
		if g.currentPlayerIdx != human {
			return nil
		}
		valid := g.getValidPlayIndices(human)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlaySmart(human, valid)
		return &GleekHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
	}
	return nil
}

// gleekBidHintReason 競り推奨に対応するヒント理由キーを返す。
func gleekBidHintReason(bid int) string {
	if bid == 0 {
		return "bid_pass"
	}
	return "bid_raise"
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Gleek) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_high"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	lead := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.trickWinner()
	topStr := gleekCardStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card, g.trumpSuit, lead)
	if gleekCardStrength(card, g.trumpSuit, lead) > topStr {
		return "follow_win"
	}
	if gleekHonourValue(card, g.trumpSuit) > 0 {
		return "discard_honour"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Gleek) GetPhase() GleekPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Gleek) SetPhase(phase GleekPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Gleek) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Gleek) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Gleek) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Gleek) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Gleek) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Gleek) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Gleek) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Gleek) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Gleek) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Gleek) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Gleek) GetDealerIdx() int { return g.dealerIdx }

// GetElderIdx エルダー (競りを開き最初にリードする席) を取得
func (g *Gleek) GetElderIdx() int { return g.elderIdx }

// GetTrumpSuit 切り札スート取得 (-1=未確定, 1..4)
func (g *Gleek) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Gleek) SetTrumpSuit(s int) { g.trumpSuit = s }

// GetTurnUp 表向きになった切り札の札を返す (nil=未確定)
func (g *Gleek) GetTurnUp() *Card { return g.turnUp }

// GetCurrentBidderIdx 現在の競り手番インデックス取得
func (g *Gleek) GetCurrentBidderIdx() int { return g.currentBidderIdx }

// GetBids 各席の競り額を取得
func (g *Gleek) GetBids() [GleekPlayerCnt]int { return g.bids }

// GetPassed 各席が降りたかを取得
func (g *Gleek) GetPassed() [GleekPlayerCnt]bool { return g.passed }

// GetBuyerIdx ストックを買った席を取得 (-1=未確定)
func (g *Gleek) GetBuyerIdx() int { return g.buyerIdx }

// SetBuyerIdx ストックを買った席を設定 (テスト用)
func (g *Gleek) SetBuyerIdx(idx int) { g.buyerIdx = idx }

// GetWinningBid 落札額を取得
func (g *Gleek) GetWinningBid() int { return g.winningBid }

// GetRuffs 各席のラフを取得
func (g *Gleek) GetRuffs() []*GleekRuff { return g.ruffs }

// GetRuffWinnerIdx ラフを取った席を取得 (-1=未確定)
func (g *Gleek) GetRuffWinnerIdx() int { return g.ruffWinnerIdx }

// GetMelds 申告されたグリーク / マーニヴァルを取得
func (g *Gleek) GetMelds() []*GleekMeld { return g.melds }

// GetTrickPoints 各席のトリック点を取得
func (g *Gleek) GetTrickPoints() [GleekPlayerCnt]int { return g.trickPoints }

// GetPlayerScores プレイヤー別累積点取得
func (g *Gleek) GetPlayerScores() [GleekPlayerCnt]int { return g.playerScores }

// SetPlayerScores プレイヤー別累積点設定 (テスト用)
func (g *Gleek) SetPlayerScores(s [GleekPlayerCnt]int) { g.playerScores = s }

// GetResult 人間視点のマッチ結果取得
func (g *Gleek) GetResult() GleekResult { return g.result }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Gleek) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *Gleek) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *Gleek) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Gleek) GetPlayer(i int) *GleekPlayer { return getPlayer(g.players, i) }

// IsHumanTurn 現在の手番 (プレイ) が人間か。
func (g *Gleek) IsHumanTurn() bool { return isHumanTurn(g.players, g.currentPlayerIdx) }

// IsHumanBidTurn 現在の競り手番が人間か。
func (g *Gleek) IsHumanBidTurn() bool {
	if g.phase != GleekPhaseBid {
		return false
	}
	if g.currentBidderIdx < 0 || g.currentBidderIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentBidderIdx].GetIsHuman()
}

// IsHumanDiscardTurn 人間の捨て札待ちか。
func (g *Gleek) IsHumanDiscardTurn() bool {
	return g.phase == GleekPhaseDiscard && g.buyerIdx >= 0 && g.players[g.buyerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Gleek) GetConfig() GleekConfig { return g.config }

// SetConfig 設定変更
func (g *Gleek) SetConfig(cfg GleekConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Gleek) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != GleekPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// gleekJSON is the JSON wire format for Gleek.
type gleekJSON struct {
	TrumpCards       *TrumpCards          `json:"tc"`
	Players          []*GleekPlayer       `json:"ps"`
	Config           GleekConfig          `json:"cf"`
	Phase            GleekPhase           `json:"ph"`
	RoundNumber      int                  `json:"rn"`
	TrickNumber      int                  `json:"tn"`
	CurrentPlayerIdx int                  `json:"ci"`
	CurrentTrick     []*TrickCard         `json:"ct"`
	TrickResolved    bool                 `json:"tr"`
	LeadPlayerIdx    int                  `json:"li"`
	DealerIdx        int                  `json:"di"`
	ElderIdx         int                  `json:"ei"`
	TrumpSuit        int                  `json:"ts"`
	TurnUp           *Card                `json:"tu"`
	CurrentBidderIdx int                  `json:"cbi"`
	Bids             [GleekPlayerCnt]int  `json:"bd"`
	Passed           [GleekPlayerCnt]bool `json:"pa"`
	BuyerIdx         int                  `json:"bi"`
	WinningBid       int                  `json:"wb"`
	Stock            []*Card              `json:"st"`
	Ruffs            []*GleekRuff         `json:"rf"`
	RuffWinnerIdx    int                  `json:"rw"`
	Melds            []*GleekMeld         `json:"ml"`
	TrickPoints      [GleekPlayerCnt]int  `json:"tp"`
	PlayerScores     [GleekPlayerCnt]int  `json:"sc"`
	LastTrickWinner  int                  `json:"lt"`
	Result           GleekResult          `json:"rs"`
	Scored           bool                 `json:"sd"`
	GameEndFlag      bool                 `json:"ge"`
	WinnerPlayer     int                  `json:"wp"`
	ActionLog        []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Gleek) MarshalJSON() ([]byte, error) {
	return json.Marshal(gleekJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		TrickResolved:    g.trickResolved,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		ElderIdx:         g.elderIdx,
		TrumpSuit:        g.trumpSuit,
		TurnUp:           g.turnUp,
		CurrentBidderIdx: g.currentBidderIdx,
		Bids:             g.bids,
		Passed:           g.passed,
		BuyerIdx:         g.buyerIdx,
		WinningBid:       g.winningBid,
		Stock:            g.stock,
		Ruffs:            g.ruffs,
		RuffWinnerIdx:    g.ruffWinnerIdx,
		Melds:            g.melds,
		TrickPoints:      g.trickPoints,
		PlayerScores:     g.playerScores,
		LastTrickWinner:  g.lastTrickWinner,
		Result:           g.result,
		Scored:           g.scored,
		GameEndFlag:      g.gameEndFlag,
		WinnerPlayer:     g.winnerPlayer,
		ActionLog:        g.actionLog,
	})
}

// gleekMaxSliceLen caps slice sizes during deserialisation.
const gleekMaxSliceLen = 5000

// errGleekOversized is the single sentinel error for oversized input arrays.
var errGleekOversized = errors.New("gleek: input array exceeds maximum allowed size")

// errGleekInvalidPlayers is returned when restored state lacks exactly GleekPlayerCnt players.
var errGleekInvalidPlayers = errors.New("gleek: invalid player count")

// errGleekInvalidTrick is returned when a restored trick card or its card is nil / out of range.
var errGleekInvalidTrick = errors.New("gleek: invalid trick card")

// errGleekInvalidIndex is returned when a restored index field is out of range.
var errGleekInvalidIndex = errors.New("gleek: index field out of range")

// errGleekInvalidPhase is returned when a restored phase is out of range.
var errGleekInvalidPhase = errors.New("gleek: phase out of range")

// errGleekInvalidTrump is returned when a restored trump suit is out of range.
var errGleekInvalidTrump = errors.New("gleek: trump suit out of range")

// errGleekInvalidResult is returned when a restored result value is out of range.
var errGleekInvalidResult = errors.New("gleek: result value out of range")

// gleekInRange reports whether v is in [0, GleekPlayerCnt).
func gleekInRange(v int) bool { return v >= 0 && v < GleekPlayerCnt }

// gleekInRangeOrUnset reports whether v is -1 (unset) or in [0, GleekPlayerCnt).
func gleekInRangeOrUnset(v int) bool { return v == -1 || gleekInRange(v) }

// UnmarshalJSON implements json.Unmarshaler.
func (g *Gleek) UnmarshalJSON(data []byte) error {
	var j gleekJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > gleekMaxSliceLen || len(j.CurrentTrick) > gleekMaxSliceLen ||
		len(j.ActionLog) > gleekMaxSliceLen || len(j.Stock) > gleekMaxSliceLen ||
		len(j.Ruffs) > gleekMaxSliceLen || len(j.Melds) > gleekMaxSliceLen {
		return errGleekOversized
	}
	if len(j.Players) != GleekPlayerCnt {
		return errGleekInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errGleekInvalidPlayers
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || !gleekInRange(tc.PlayerIdx) {
			return errGleekInvalidTrick
		}
	}
	if int(j.Phase) < GleekPhaseMin || int(j.Phase) > GleekPhaseMax {
		return errGleekInvalidPhase
	}
	if !gleekInRange(j.CurrentPlayerIdx) || !gleekInRange(j.DealerIdx) ||
		!gleekInRange(j.ElderIdx) || !gleekInRange(j.CurrentBidderIdx) {
		return errGleekInvalidIndex
	}
	if !gleekInRangeOrUnset(j.LeadPlayerIdx) || !gleekInRangeOrUnset(j.BuyerIdx) ||
		!gleekInRangeOrUnset(j.RuffWinnerIdx) || !gleekInRangeOrUnset(j.LastTrickWinner) ||
		!gleekInRangeOrUnset(j.WinnerPlayer) {
		return errGleekInvalidIndex
	}
	// **切り札は配った時点で決まる。** 表向きの札で決まるので、競りの最中から
	// 既に確定している。-1 のまま復元すると、どの札も切り札にならない盤で
	// 12 トリック打つことになり、名札の点が 1 つも出ない。
	if !gleekValidSuit(j.TrumpSuit) {
		return errGleekInvalidTrump
	}
	// **捨て札以降は買い手が確定している。** -1 のまま通すと、
	// PlayerDiscard も CpuDiscard も何もしないフェーズで止まる。
	if j.Phase >= GleekPhaseDiscard && !gleekInRange(j.BuyerIdx) {
		return errGleekInvalidIndex
	}
	if j.Phase >= GleekPhasePlay && !gleekInRange(j.LeadPlayerIdx) {
		return errGleekInvalidIndex
	}
	for _, r := range j.Ruffs {
		if r == nil || !gleekInRange(r.PlayerIdx) {
			return errGleekInvalidIndex
		}
	}
	for _, m := range j.Melds {
		if m == nil || !gleekInRange(m.PlayerIdx) {
			return errGleekInvalidIndex
		}
	}
	if j.Result < GleekResultLose || j.Result > GleekResultWin {
		return errGleekInvalidResult
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newGleekDeck()
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.trickResolved = j.TrickResolved
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.elderIdx = j.ElderIdx
	g.trumpSuit = j.TrumpSuit
	g.turnUp = j.TurnUp
	g.currentBidderIdx = j.CurrentBidderIdx
	g.bids = j.Bids
	g.passed = j.Passed
	g.buyerIdx = j.BuyerIdx
	g.winningBid = j.WinningBid
	g.stock = j.Stock
	g.ruffs = j.Ruffs
	g.ruffWinnerIdx = j.RuffWinnerIdx
	g.melds = j.Melds
	g.trickPoints = j.TrickPoints
	g.playerScores = j.PlayerScores
	g.lastTrickWinner = j.LastTrickWinner
	g.result = j.Result
	g.scored = j.Scored
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	return nil
}
