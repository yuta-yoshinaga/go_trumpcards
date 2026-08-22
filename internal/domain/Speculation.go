//go:build !js || !wasm || extra

package domain

import (
	"errors"
	"fmt"
)

// SpeculationPhase はゲームの進行段階。
type SpeculationPhase int

const (
	// SpeculationPhaseFlip は次の伏せ札をめくる段階。
	SpeculationPhaseFlip SpeculationPhase = iota
	// SpeculationPhaseAuction は今めくれた切り札に値が付いている段階。
	// **人間が売るか断るかを決める**か、人間が買い手として値を付ける。
	SpeculationPhaseAuction
	// SpeculationPhaseResult はラウンドが決着した段階。
	SpeculationPhaseResult
	// SpeculationPhaseGameEnd は規定ラウンドを終えた段階。
	SpeculationPhaseGameEnd
)

// SpeculationPhaseMax は最大のフェーズ値 (復元時の範囲検査に使う)。
const SpeculationPhaseMax = SpeculationPhaseGameEnd

// speculationMaxSliceLen は復元時のスライス長の上限。
const speculationMaxSliceLen = 512

var (
	errSpeculationWrongPhase = errors.New("speculation: not allowed in this phase")
	errSpeculationNoOffer    = errors.New("speculation: there is no offer to answer")
	errSpeculationBadAmount  = errors.New("speculation: invalid amount")
)

// Speculation はスペキュレーション（18〜19 世紀イギリスの競りゲーム）。
//
// **純粋な競り。** プレイヤーは自分の伏せ札を 1 枚ずつめくり、その時点で最高の
// 切り札が出たら、他家がその札に値を付けて買い取りを申し出る。持ち主は売っても
// 断ってもよい。全員がめくり終えた時点で最高の切り札を持つ者がポットを総取りする。
//
// 賭ける相手がバンカーではなく他プレイヤー全員である点が、同じ「賭け」でも
// モンテバンクのような対バンカー戦と根本的に違う。
type Speculation struct {
	deck    *TrumpCards
	players []*SpeculationPlayer
	config  SpeculationConfig

	phase SpeculationPhase
	// trumpSuit はこのラウンドの切り札スート。
	trumpSuit int
	// trumpCard は切り札を決めるためにめくった札。
	trumpCard *Card
	// pot はこのラウンドの賭け金の総額。
	pot int
	// turnSeat は次にめくる番の座席。
	turnSeat int

	// offerFrom は買い取りを申し出ている座席。申し出が無ければ -1。
	offerFrom int
	// offerTo は申し出を受けている（札の持ち主の）座席。無ければ -1。
	offerTo int
	// offerAmount は提示額。
	offerAmount int

	// bestSeat は現在の最高切り札を持つ座席。誰も持っていなければ -1。
	bestSeat int

	roundNo     int
	winnerSeat  int
	gameEndFlag bool
	actionLogBase
}

// NewSpeculation は設定を与えて卓を作る。
func NewSpeculation(cfg SpeculationConfig) *Speculation {
	cfg.Normalize()
	g := &Speculation{config: cfg, deck: NewTrumpCards(0)}
	g.players = make([]*SpeculationPlayer, cfg.Players)
	for i := range g.players {
		name := fmt.Sprintf("CPU%d", i)
		if i == 0 {
			name = "You"
		}
		g.players[i] = NewSpeculationPlayer(name, cfg.InitialChips)
	}
	g.Reset()
	return g
}

// NewDefaultSpeculation は既定設定の卓を作る。
func NewDefaultSpeculation() *Speculation { return NewSpeculation(NewDefaultSpeculationConfig()) }

// Reset は次のラウンドを始める。チップとラウンド数は持ち越す。
func (g *Speculation) Reset() {
	// **未精算のポットは先に返す。順序が肝。** Reset は初期化として何度でも
	// 呼ばれうる (コンストラクタで一度、ユースケースの起動時にもう一度) が、
	// そのたびに参加料だけ取って古いポットを捨てると**チップが消える** ——
	// 実測で 200 → 190 → 180 と減り、ポットは 40 のままだった。
	// この返却を下の `g.pot = 0` より後ろに置くと、返す前に額が消えるので
	// 何も起きない。
	g.refundPot()

	g.phase = SpeculationPhaseFlip
	g.pot = 0
	// **Reset は「新しいゲーム」。** ラウンド数を持ち越すと、終局画面から
	// リセットしたとき roundNo が上限のままで、最初の決着で即座にまた
	// 終局する。ラウンドを 1 つ進めたいだけの NextRound は、この値を自分で
	// 退避して戻している。
	g.roundNo = 0
	g.turnSeat = 0
	g.offerFrom, g.offerTo, g.offerAmount = -1, -1, 0
	g.bestSeat = -1
	g.winnerSeat = -1
	g.gameEndFlag = false
	g.actionLog = nil
	g.trumpCard = nil
	g.trumpSuit = -1

	g.deck = NewTrumpCards(0)
	for range 10 {
		g.deck.Shuffle()
	}

	// 参加料を集める。**払えない席は 0 まで出す** —— 途中で払えなくなった
	// プレイヤーを弾くと座席番号がずれ、ラウンドを跨いだ集計が崩れる。
	for _, p := range g.players {
		stake := g.config.Stake
		if p.GetChips() < stake {
			stake = p.GetChips()
		}
		p.SubtractChips(stake)
		g.pot += stake
		p.SetBest(nil)
	}

	// 伏せ札を配る。
	for _, p := range g.players {
		hand := make([]*Card, 0, SpeculationCardsPerPlayer)
		for range SpeculationCardsPerPlayer {
			if c := g.deck.DrawCard(); c != nil {
				hand = append(hand, c)
			}
		}
		p.SetHidden(hand)
	}

	// 切り札を決める。**山から 1 枚めくる** —— この札は誰の手にもならない。
	if c := g.deck.DrawCard(); c != nil {
		g.trumpCard = c
		g.trumpSuit = c.GetDesign()
	}
	g.appendLog(-1, "deal", fmt.Sprintf("trump=%d pot=%d", g.trumpSuit, g.pot), nil)
}

// Flip は手番の席の伏せ札を 1 枚めくる。
//
// めくれた札が切り札で、かつそれまでの最高を上回れば持ち主が変わり、競りが
// 開く。**人間はどちらの向きでも必ず訊かれる** —— 自分が最高札を得たときは
// 売るかどうかを、CPU が得たときは買うかどうかを。CPU 同士だけで札が動くこと
// は無い。競りに人間が関わらない局面を作ると、盤面が勝手に進んで見える。
func (g *Speculation) Flip() error {
	if g.phase != SpeculationPhaseFlip {
		return errSpeculationWrongPhase
	}
	seat := g.turnSeat
	p := g.players[seat]
	c := p.FlipTop()
	if c == nil {
		// この席はもう出す札が無い。次へ。
		g.advanceTurn()
		return nil
	}
	g.appendLog(seat, "flip", speculationCardStr(c), nil)

	if g.isNewBest(c) {
		g.setBest(seat, c)
		g.appendLog(seat, "lead", speculationCardStr(c), nil)
		if g.openAuction(seat) {
			return nil // 人間の判断待ち
		}
	}
	g.advanceTurn()
	return nil
}

// isNewBest はその札が現在の最高切り札を上回るかを返す。
//
// **切り札でなければ何枚出ても関係ない。** スートが違う札はこのゲームでは
// ただの紙で、競りの対象にもならない。
func (g *Speculation) isNewBest(c *Card) bool {
	if c == nil || c.GetDesign() != g.trumpSuit {
		return false
	}
	if g.bestSeat < 0 || g.players[g.bestSeat].GetBest() == nil {
		return true
	}
	return speculationRank(c) > speculationRank(g.players[g.bestSeat].GetBest())
}

// setBest は最高切り札の持ち主を差し替える。前の持ち主からは取り上げる。
func (g *Speculation) setBest(seat int, c *Card) {
	if g.bestSeat >= 0 && g.bestSeat != seat {
		g.players[g.bestSeat].SetBest(nil)
	}
	g.bestSeat = seat
	g.players[seat].SetBest(c)
}

// speculationRank は切り札の強さ。**A が最強** (14)。
func speculationRank(c *Card) int {
	if c == nil {
		return 0
	}
	if v := c.GetValue(); v == 1 {
		return 14
	}
	return c.GetValue()
}

// openAuction は今の最高札に買い手が付くかを見る。人間の判断が要るときだけ
// true を返し、競りフェーズへ移す。
func (g *Speculation) openAuction(ownerSeat int) bool {
	buyer, amount := g.bestCPUOffer(ownerSeat)

	if ownerSeat == 0 {
		// 人間が持ち主。CPU から値が付いていれば売るか断るかを訊く。
		if buyer < 0 {
			return false
		}
		g.offerFrom, g.offerTo, g.offerAmount = buyer, ownerSeat, amount
		g.phase = SpeculationPhaseAuction
		g.appendLog(buyer, "offer", fmt.Sprintf("%d to seat %d", amount, ownerSeat), nil)
		return true
	}

	// CPU が持ち主。人間に買う気があるかを訊く。**手が届かない額は提示しない**
	// —— 払えない申し出を受けると Accept が黙って競りを閉じ、プレイヤーには
	// 理由も出ないまま申し出だけが消える。
	ask := g.askingPrice(ownerSeat)
	if chips := g.players[0].GetChips(); ask > chips {
		ask = chips
	}
	if ask <= 0 {
		return false // 1 チップも出せないなら競りにならない
	}
	g.offerFrom, g.offerTo, g.offerAmount = 0, ownerSeat, ask
	g.phase = SpeculationPhaseAuction
	g.appendLog(ownerSeat, "asking", fmt.Sprintf("%d", g.offerAmount), nil)
	return true
}

// bestCPUOffer は人間以外で最も高い買値を付ける席とその額を返す。
// 誰も買わなければ (-1, 0)。
func (g *Speculation) bestCPUOffer(ownerSeat int) (int, int) {
	bestSeat, bestAmt := -1, 0
	for seat, p := range g.players {
		if seat == ownerSeat || seat == 0 {
			continue
		}
		amt := g.cpuValuation(seat)
		if amt > bestAmt && p.GetChips() >= amt {
			bestSeat, bestAmt = seat, amt
		}
	}
	return bestSeat, bestAmt
}

// cpuValuation は CPU がその席から見て最高札に付ける値。
//
// **まだ伏せ札が多く残っているほど安く見る。** 買ってもすぐ上を出されるなら
// 価値が低い。逆に残りが少なければ、その札がそのままポットを取る公算が高い。
func (g *Speculation) cpuValuation(seat int) int {
	best := g.currentBestCard()
	if best == nil {
		return 0
	}
	remaining := g.remainingHidden()
	// 札の強さ (2..14) をポットに対する割合に均す。
	value := g.pot * speculationRank(best) / 14
	// 残り札が多いほど割り引く。全部で players*cards 枚。
	total := len(g.players) * SpeculationCardsPerPlayer
	if total > 0 {
		value = value * (total - remaining) / total
	}
	if value > g.players[seat].GetChips() {
		value = g.players[seat].GetChips()
	}
	if value < 0 {
		value = 0
	}
	return value
}

// askingPrice は CPU が持ち主のときに人間へ提示する売値。
// 買値より少し高い —— 持ち主は手放す理由が無ければ売らない。
func (g *Speculation) askingPrice(ownerSeat int) int {
	v := g.cpuValuation(ownerSeat)
	price := v + v/4 + 1
	if price < 1 {
		price = 1
	}
	return price
}

// currentBestCard は現在の最高切り札を返す。
func (g *Speculation) currentBestCard() *Card {
	if g.bestSeat < 0 {
		return nil
	}
	return g.players[g.bestSeat].GetBest()
}

// remainingHidden はまだめくられていない伏せ札の総数。
func (g *Speculation) remainingHidden() int {
	n := 0
	for _, p := range g.players {
		n += p.GetHiddenCount()
	}
	return n
}

// Accept は競りの申し出を受ける。
//
// 人間が持ち主なら売り、人間が買い手なら買う。どちらの向きでも、札とチップが
// 逆向きに動く。
func (g *Speculation) Accept() error {
	if g.phase != SpeculationPhaseAuction {
		return errSpeculationWrongPhase
	}
	if g.offerFrom < 0 || g.offerTo < 0 {
		return errSpeculationNoOffer
	}
	buyer, owner, amount := g.players[g.offerFrom], g.players[g.offerTo], g.offerAmount
	if !buyer.SubtractChips(amount) {
		// **黙って閉じない。** 以前は競りを閉じて手番を進めていたので、
		// プレイヤーには理由も出ないまま申し出だけが消えていた。競りは開いた
		// ままにして、断る道を残す。
		return errSpeculationBadAmount
	}
	owner.AddChips(amount)
	card := owner.GetBest()
	owner.SetBest(nil)
	g.bestSeat = g.offerFrom
	buyer.SetBest(card)
	g.appendLog(g.offerFrom, "buy", fmt.Sprintf("%s for %d", speculationCardStr(card), amount), nil)

	g.closeAuction()
	g.advanceTurn()
	return nil
}

// Decline は競りの申し出を断る。札もチップも動かない。
func (g *Speculation) Decline() error {
	if g.phase != SpeculationPhaseAuction {
		return errSpeculationWrongPhase
	}
	if g.offerFrom < 0 || g.offerTo < 0 {
		return errSpeculationNoOffer
	}
	g.appendLog(g.offerTo, "decline", fmt.Sprintf("%d", g.offerAmount), nil)
	g.closeAuction()
	g.advanceTurn()
	return nil
}

// Bid は人間が自分から値を付け直す（提示額に上乗せする）。
//
// **上乗せしか認めない。** 提示額より低い値を「申し出」と呼ぶと、断るのと
// 区別が付かなくなる。
func (g *Speculation) Bid(amount int) error {
	if g.phase != SpeculationPhaseAuction {
		return errSpeculationWrongPhase
	}
	if g.offerFrom != 0 {
		return errSpeculationNoOffer // 人間が買い手のときだけ値を付け直せる
	}
	if amount <= g.offerAmount {
		return errSpeculationBadAmount
	}
	if amount > g.players[0].GetChips() {
		return errSpeculationBadAmount
	}
	g.offerAmount = amount
	g.appendLog(0, "bid", fmt.Sprintf("%d", amount), nil)
	return g.Accept()
}

// closeAuction は競りを閉じる。
func (g *Speculation) closeAuction() {
	g.offerFrom, g.offerTo, g.offerAmount = -1, -1, 0
	g.phase = SpeculationPhaseFlip
}

// advanceTurn は手番を次の席へ送り、全員がめくり終えていれば決着させる。
func (g *Speculation) advanceTurn() {
	if g.remainingHidden() == 0 {
		g.resolve()
		return
	}
	for range len(g.players) {
		g.turnSeat = (g.turnSeat + 1) % len(g.players)
		if g.players[g.turnSeat].GetHiddenCount() > 0 {
			return
		}
	}
	g.resolve()
}

// resolve はラウンドを決着させ、ポットを最高切り札の持ち主に渡す。
//
// **誰も切り札を出さないことがある。** その場合ポットは次のラウンドへ持ち越す
// のではなく、参加料を出した全員に返す —— 取り分が決まらない賭けを繰り越すと、
// 途中で降りた席の負担だけが積み上がる。
func (g *Speculation) resolve() {
	g.winnerSeat = g.bestSeat
	if g.winnerSeat >= 0 {
		g.players[g.winnerSeat].AddChips(g.pot)
		g.appendLog(g.winnerSeat, "win", fmt.Sprintf("pot=%d", g.pot), nil)
	} else {
		share := g.pot / len(g.players)
		g.refundPot()
		g.appendLog(-1, "void", fmt.Sprintf("no trump; returned %d each", share), nil)
	}
	g.pot = 0
	g.roundNo++
	if g.roundNo >= g.config.Rounds {
		g.phase = SpeculationPhaseGameEnd
		g.gameEndFlag = true
		g.appendLog(-1, "gameend", fmt.Sprintf("rounds=%d", g.roundNo), nil)
		return
	}
	g.phase = SpeculationPhaseResult
}

// NextRound は決着後に次のラウンドを始める。
func (g *Speculation) NextRound() error {
	if g.phase != SpeculationPhaseResult {
		return errSpeculationWrongPhase
	}
	round := g.roundNo
	g.Reset()
	g.roundNo = round
	return nil
}

// refundPot は未精算のポットを参加者に等分して返す。
//
// 端数は切り捨てて卓に残さず、最初の席に寄せる —— 「返した」と言いながら
// 総額が減るのは、消えるのと同じくらい説明が付かない。
func (g *Speculation) refundPot() {
	if g.pot <= 0 || len(g.players) == 0 {
		return
	}
	share := g.pot / len(g.players)
	for _, p := range g.players {
		p.AddChips(share)
	}
	g.players[0].AddChips(g.pot - share*len(g.players))
	g.pot = 0
}

// speculationCardStr はログ用のカード表記。
func speculationCardStr(c *Card) string {
	if c == nil {
		return "-"
	}
	return fmt.Sprintf("%d-%d", c.GetDesign(), c.GetValue())
}

// --- Getters ---

// GetPhase は現在のフェーズを返す。
func (g *Speculation) GetPhase() SpeculationPhase { return g.phase }

// GetPlayers は全プレイヤーを返す。
func (g *Speculation) GetPlayers() []*SpeculationPlayer { return g.players }

// GetConfig は卓設定を返す。
func (g *Speculation) GetConfig() SpeculationConfig { return g.config }

// GetTrumpSuit はこのラウンドの切り札スートを返す。
func (g *Speculation) GetTrumpSuit() int { return g.trumpSuit }

// GetTrumpCard は切り札を決めた札を返す。
func (g *Speculation) GetTrumpCard() *Card { return g.trumpCard }

// GetPot は現在のポットを返す。
func (g *Speculation) GetPot() int { return g.pot }

// GetTurnSeat は次にめくる席を返す。
func (g *Speculation) GetTurnSeat() int { return g.turnSeat }

// GetBestSeat は最高切り札を持つ席を返す。誰も持っていなければ -1。
func (g *Speculation) GetBestSeat() int { return g.bestSeat }

// GetOfferFrom は買い取りを申し出ている席を返す。無ければ -1。
func (g *Speculation) GetOfferFrom() int { return g.offerFrom }

// GetOfferTo は申し出を受けている席を返す。無ければ -1。
func (g *Speculation) GetOfferTo() int { return g.offerTo }

// GetOfferAmount は提示額を返す。
func (g *Speculation) GetOfferAmount() int { return g.offerAmount }

// GetRoundNo は消化済みのラウンド数を返す。
func (g *Speculation) GetRoundNo() int { return g.roundNo }

// GetWinnerSeat は直前のラウンドの勝者席を返す。決着前・流局なら -1。
func (g *Speculation) GetWinnerSeat() int { return g.winnerSeat }

// GetGameEndFlag はゲームが終わったかを返す。
func (g *Speculation) GetGameEndFlag() bool { return g.gameEndFlag }

// --- Test helpers ---

// SetPhase はフェーズを設定する（テスト用）。
func (g *Speculation) SetPhase(p SpeculationPhase) { g.phase = p }

// SetTrumpSuit は切り札スートを設定する（テスト用）。
func (g *Speculation) SetTrumpSuit(s int) { g.trumpSuit = s }

// SetPot はポットを設定する（テスト用）。
func (g *Speculation) SetPot(n int) { g.pot = n }

// SetTurnSeat は手番の席を設定する（テスト用）。
func (g *Speculation) SetTurnSeat(s int) { g.turnSeat = s }

// SetBestSeat は最高札の席を設定する（テスト用）。
func (g *Speculation) SetBestSeat(s int) { g.bestSeat = s }

// SetOffer は競りの申し出を設定する（テスト用）。
func (g *Speculation) SetOffer(from, to, amount int) {
	g.offerFrom, g.offerTo, g.offerAmount = from, to, amount
}

// SetRoundNo はラウンド数を設定する（テスト用）。
func (g *Speculation) SetRoundNo(n int) { g.roundNo = n }

// DeckForTest は山札を返す（テスト用）。
func (g *Speculation) DeckForTest() *TrumpCards { return g.deck }
