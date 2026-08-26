//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

// chemindeFerMaxSliceLen はデシリアライズ時のスライス長上限。
const chemindeFerMaxSliceLen = 1000

// エラー値。
var (
	errChemindeFerWrongPhase   = errors.New("chemindefer: action not allowed in this phase")
	errChemindeFerNotYourSeat  = errors.New("chemindefer: it is not that seat's turn")
	errChemindeFerStakeRange   = errors.New("chemindefer: stake is out of range")
	errChemindeFerBetRange     = errors.New("chemindefer: bet is out of range")
	errChemindeFerNoChoice     = errors.New("chemindefer: that side has no choice at this total")
	errChemindeFerGameFinished = errors.New("chemindefer: the game has already finished")
)

// ChemindeFerHint は人間への助言。
type ChemindeFerHint struct {
	// Draw は「引く」を薦めるか。
	Draw bool
	// Reason は理由の識別子 (i18n キーの一部)。
	Reason string
}

// ChemindeFer はシュマン・ド・フェール (Chemin de Fer) の卓。
//
// バカラの原型にあたるフランスのバンキングゲームで、**ハウスではなく席の 1 つが親**
// になり、他の席がその親に対して賭ける。親は勝っている限り親を続けられ、負けると
// バンクが隣へ渡る。プント・バンコとの決定的な違いは 3 枚目の引き方で、あちらは
// 親も子も表で固定されているが、こちらで固定されているのは**子の 0-4 と 6-7 だけ**。
// 合計 5 の子と、**あらゆる合計の親**は自分で決める。
type ChemindeFer struct {
	shoe    *TrumpCards
	players []*ChemindeFerPlayer
	config  ChemindeFerConfig
	rng     *rand.Rand

	phase     ChemindeFerPhase
	bankerIdx int
	// betOrder は親の次の席から並べた「賭ける順番」。ラウンドごとに作り直す。
	//
	// 席番号の剰余計算を毎回やると、親がどこに居るかで条件が変わって読めなくなる。
	// 順番を 1 本の配列にしてしまえば「次は betOrder[betPos+1]」で済む。
	betOrder   []int
	betPos     int // betOrder のどこまで進んだか。-1 = 賭けは終わっている
	stake      int
	represIdx  int // 子側の代表 (最高額ベッター)。-1 = 未定
	bankerHand []*Card
	punterHand []*Card
	punterDrew bool
	result     ChemindeFerResult
	// lastNet は直前の決済での席ごとの純増減。ラウンドを始めると 0 に戻る。
	//
	// **卓の結果と自分の損益は別の情報。** banker/punter/tie だけでは、自分の
	// 賭けが勝ったのか負けたのかはチップの数字を前後で見比べるしかない (#5774)。
	lastNet     []int
	roundNumber int
	gameEndFlag bool
	actionLog   []*ActionLogEntry
	turnNumber  int
}

// NewChemindeFer は指定のシュー・席・設定で卓を構築する。
func NewChemindeFer(shoe *TrumpCards, players []*ChemindeFerPlayer, config ChemindeFerConfig) *ChemindeFer {
	return &ChemindeFer{
		shoe:        shoe,
		players:     players,
		config:      config,
		rng:         rand.New(rand.NewSource(1)),
		represIdx:   -1,
		betPos:      -1,
		roundNumber: 1,
	}
}

// NewDefaultChemindeFer は既定の卓 (席 0 が人間、6 デッキ) を構築する。
func NewDefaultChemindeFer() *ChemindeFer {
	cfg := DefaultChemindeFerConfig()
	players := make([]*ChemindeFerPlayer, ChemindeFerSeatCnt)
	for i := range players {
		players[i] = NewChemindeFerPlayer(
			fmt.Sprintf("Player%d", i+1), cfg.InitialChips, i == 0)
	}
	return NewChemindeFer(newChemindeFerShoe(), players, cfg)
}

// newChemindeFerShoe は 6 デッキのシューを返す。
func newChemindeFerShoe() *TrumpCards {
	return NewTrumpCardsWithDecks(ChemindeFerDeckCnt, 0)
}

// SetRand は CPU の判断に使う乱数源を差し替える (テスト用)。
//
// **シューの切り方までは決まらない。** TrumpCards.Shuffle はグローバルな rand を
// 使うので、配りを固定したいテストは手札を直接積むこと。
func (g *ChemindeFer) SetRand(r *rand.Rand) {
	if r != nil {
		g.rng = r
	}
}

// Reset はゲームを初期状態に戻し、人間の番まで CPU を進める。
//
// **バンクは席 0 (人間) から始まる。**
func (g *ChemindeFer) Reset() {
	g.reset()
	g.advanceCpu()
}

// reset は盤面を初期化するだけで CPU を進めない。
func (g *ChemindeFer) reset() {
	g.shoe = newChemindeFerShoe()
	g.shoe.Shuffle()
	for _, p := range g.players {
		p.SetChips(g.config.InitialChips)
		p.SetBet(0)
	}
	g.bankerIdx = 0
	g.roundNumber = 1
	g.gameEndFlag = false
	g.actionLog = nil
	g.turnNumber = 0
	g.appendLog(-1, "start", "chemin de fer begins", nil)
	g.startRound()
}

// startRound は 1 ラウンド分の状態を初期化して親の張り待ちにする。
func (g *ChemindeFer) startRound() {
	for _, p := range g.players {
		p.SetBet(0)
	}
	g.stake = 0
	g.represIdx = -1
	g.bankerHand = nil
	g.punterHand = nil
	g.punterDrew = false
	g.result = ChemindeFerResultNone
	g.lastNet = make([]int, len(g.players))
	g.betOrder = nil
	g.betPos = -1
	g.phase = ChemindeFerPhaseStake
	g.ensureShoe()
}

// ensureShoe は 1 ラウンド分に足りない残り枚数ならシューを組み直す。
//
// 1 ラウンドで最大 6 枚 (親子 3 枚ずつ) 使う。**足りないまま配ると nil 札が混ざる**ので、
// 引く前に必ず通す。
func (g *ChemindeFer) ensureShoe() {
	if g.shoe.GetRemainingCount() < ChemindeFerMaxHandSize*2 {
		g.shoe.Replenish()
		g.shoe.Shuffle()
	}
}

// --- 規則: 合計値 ---

// chemindeFerCardValue は 1 枚のポイント値 (A=1、2-9 は額面、10/J/Q/K=0)。
func chemindeFerCardValue(c *Card) int {
	if c == nil {
		return 0
	}
	if v := c.GetValue(); v < 10 {
		return v
	}
	return 0
}

// ChemindeFerHandTotal は手札の合計値 (mod 10) を返す。
func ChemindeFerHandTotal(cards []*Card) int {
	total := 0
	for _, c := range cards {
		total += chemindeFerCardValue(c)
	}
	return total % 10
}

// GetBankerTotal は親の現在の合計値を返す。
func (g *ChemindeFer) GetBankerTotal() int { return ChemindeFerHandTotal(g.bankerHand) }

// GetPunterTotal は子側の現在の合計値を返す。
func (g *ChemindeFer) GetPunterTotal() int { return ChemindeFerHandTotal(g.punterHand) }

// --- 親の張り ---

// StakeRangeFor は席 idx が親として張れる額の範囲を返す。張れないなら (0, 0)。
func (g *ChemindeFer) StakeRangeFor(idx int) (minStake, maxStake int) {
	if idx < 0 || idx >= len(g.players) {
		return 0, 0
	}
	chips := g.players[idx].GetChips()
	if chips < ChemindeFerStakeMin {
		return 0, 0
	}
	return ChemindeFerStakeMin, chips
}

// SetStake は親がバンク額を張り、次に人間の番が来るまで CPU を進める。
func (g *ChemindeFer) SetStake(amount int) error {
	if err := g.setStake(amount); err != nil {
		return err
	}
	g.advanceCpu()
	return nil
}

// setStake は張りだけを行う。**CPU は進めない。**
func (g *ChemindeFer) setStake(amount int) error {
	if g.gameEndFlag {
		return errChemindeFerGameFinished
	}
	if g.phase != ChemindeFerPhaseStake {
		return errChemindeFerWrongPhase
	}
	lo, hi := g.StakeRangeFor(g.bankerIdx)
	if hi == 0 || amount < lo || amount > hi {
		return errChemindeFerStakeRange
	}
	g.stake = amount
	g.appendLog(g.bankerIdx, "stake", fmt.Sprintf("banks %d", amount), nil)

	g.betOrder = g.buildBetOrder()
	if len(g.betOrder) == 0 {
		// 賭けられる子が 1 人も居ない。バンクを渡してラウンドを流す。
		g.voidRound()
		return nil
	}
	g.betPos = 0
	g.phase = ChemindeFerPhaseBet
	return nil
}

// buildBetOrder は親の次の席から順に、賭けられる子の席を並べる。
func (g *ChemindeFer) buildBetOrder() []int {
	n := len(g.players)
	order := make([]int, 0, n-1)
	for step := 1; step < n; step++ {
		idx := (g.bankerIdx + step) % n
		if g.players[idx].GetChips() > 0 {
			order = append(order, idx)
		}
	}
	return order
}

// remainingStake はまだ誰にも覆われていないバンク額を返す。
func (g *ChemindeFer) remainingStake() int {
	left := g.stake
	for _, p := range g.players {
		left -= p.GetBet()
	}
	return left
}

// GetRemainingStake は覆われていないバンク額を返す。
func (g *ChemindeFer) GetRemainingStake() int { return g.remainingStake() }

// GetBetTurn は次に賭ける子の席を返す。賭けフェーズでなければ -1。
func (g *ChemindeFer) GetBetTurn() int {
	if g.phase != ChemindeFerPhaseBet || g.betPos < 0 || g.betPos >= len(g.betOrder) {
		return -1
	}
	return g.betOrder[g.betPos]
}

// BetRangeFor は席 idx がいま賭けられる額の範囲を返す。**0 は「降りる」で常に許す。**
func (g *ChemindeFer) BetRangeFor(idx int) (minBet, maxBet int) {
	if idx < 0 || idx >= len(g.players) || idx == g.bankerIdx {
		return 0, 0
	}
	hi := g.remainingStake()
	if chips := g.players[idx].GetChips(); chips < hi {
		hi = chips
	}
	if hi < 0 {
		hi = 0
	}
	return 0, hi
}

// PlaceBet は子が賭け、次に人間の番が来るまで CPU を進める。
func (g *ChemindeFer) PlaceBet(idx, amount int) error {
	if err := g.placeBet(idx, amount); err != nil {
		return err
	}
	g.advanceCpu()
	return nil
}

// placeBet は賭けだけを行う。**CPU は進めない。**
func (g *ChemindeFer) placeBet(idx, amount int) error {
	if g.gameEndFlag {
		return errChemindeFerGameFinished
	}
	if g.phase != ChemindeFerPhaseBet {
		return errChemindeFerWrongPhase
	}
	if idx != g.GetBetTurn() {
		return errChemindeFerNotYourSeat
	}
	lo, hi := g.BetRangeFor(idx)
	if amount < lo || amount > hi {
		return errChemindeFerBetRange
	}
	if amount > 0 {
		g.players[idx].SubtractChips(amount)
		g.players[idx].SetBet(amount)
		g.appendLog(idx, "bet", fmt.Sprintf("bets %d", amount), nil)
	} else {
		g.appendLog(idx, "pass", "passes", nil)
	}
	g.advanceBetting()
	return nil
}

// advanceBetting は次の子へ順番を渡し、全員が済んだら配る。
func (g *ChemindeFer) advanceBetting() {
	g.betPos++
	// バンク額が覆い尽くされたら、順番が残っていても打ち切る。
	if g.betPos < len(g.betOrder) && g.remainingStake() > 0 {
		return
	}
	g.betPos = -1
	if g.totalBet() == 0 {
		// 誰も乗らなかった。**このラウンドは流れ、バンクは隣へ渡る。**
		//
		// ここで「同じ親がもう一度張り直す」ようにすると、全員が降り続ける限り
		// Stake -> Bet -> Stake を往復してラウンドが 1 つも進まない。進行の保証を
		// 規則の善意に預けないため、乗り手が居ないラウンドも RoundEnd で終わらせる。
		g.voidRound()
		return
	}
	g.deal()
}

// totalBet は子側が賭けている総額 = 親が晒している額を返す。
func (g *ChemindeFer) totalBet() int {
	total := 0
	for i, p := range g.players {
		if i != g.bankerIdx {
			total += p.GetBet()
		}
	}
	return total
}

// GetTotalBet は子側の賭け総額を返す。
func (g *ChemindeFer) GetTotalBet() int { return g.totalBet() }

// voidRound は乗り手の居なかったラウンドを流してバンクを渡す。
//
// 「誰も賭けられるチップを持っていない」場合と「全員が降りた」場合の両方がここに来る。
func (g *ChemindeFer) voidRound() {
	g.appendLog(-1, "voidRound", "nobody covered the bank", nil)
	g.betOrder = nil
	g.betPos = -1
	g.result = ChemindeFerResultNone
	g.phase = ChemindeFerPhaseRoundEnd
	g.passBank()
}

// --- 配りと引き ---

// deal は親子に 2 枚ずつ配り、ナチュラルなら即決着させる。
func (g *ChemindeFer) deal() {
	g.represIdx = g.highestBettor()
	// 賭けは締まった。**順番はここで捨てる。**
	//
	// 残しておくと、子側が勝ってバンクが隣へ渡ったあとも「前の親を基準に並べた順番」が
	// 居座り、新しい親が自分の賭け順に入っている盤面になる。復元側の検査はそれを
	// (正しく) 壊れていると判定するので、正しい局面が保存できなくなる。
	g.betOrder = nil
	g.betPos = -1
	for range ChemindeFerHandSize {
		g.punterHand = append(g.punterHand, g.shoe.DrawCard())
		g.bankerHand = append(g.bankerHand, g.shoe.DrawCard())
	}
	g.appendLog(g.represIdx, "deal", "cards dealt", nil)
	g.afterDeal()
}

// afterDeal はナチュラル判定と子側の判断への受け渡しを行う。
func (g *ChemindeFer) afterDeal() {
	if ChemindeFerIsNatural(g.GetPunterTotal()) || ChemindeFerIsNatural(g.GetBankerTotal()) {
		g.appendLog(-1, "natural", "a natural ends the coup at once", nil)
		g.resolve()
		return
	}
	g.phase = ChemindeFerPhasePunterDraw
	g.autoAdvancePunter()
}

// highestBettor は子側の代表 (最高額ベッター) の席を返す。
//
// **同額なら親に近い側が代表。** 実際の卓でも親の右隣から順に賭けるので、
// 先に張った方が優先される。
func (g *ChemindeFer) highestBettor() int {
	best, bestBet := -1, 0
	n := len(g.players)
	for step := 1; step < n; step++ {
		idx := (g.bankerIdx + step) % n
		if b := g.players[idx].GetBet(); b > bestBet {
			best, bestBet = idx, b
		}
	}
	return best
}

// autoAdvancePunter は子側に選択の余地が無ければ規則どおり自動で進める。
//
// **人間に押させるのは本当に選べるときだけ**にしたい。0-4 は引く以外に無く、
// 6-7 は立つ以外に無いので、そこで手を止めても意味のある選択にはならない。
func (g *ChemindeFer) autoAdvancePunter() {
	total := g.GetPunterTotal()
	switch {
	case ChemindeFerPunterMustDraw(total):
		g.punterDraw()
	case ChemindeFerPunterMustStand(total):
		g.punterStand()
	case g.represIdx >= 0 && g.players[g.represIdx].GetIsHuman():
		// 合計 5 かつ代表が人間。**ここだけが子側の判断どころ。**
	default:
		g.punterDecideCpu()
	}
}

// PunterMayChoose は子側がいま自分で選べるかを返す。
func (g *ChemindeFer) PunterMayChoose() bool {
	return g.phase == ChemindeFerPhasePunterDraw &&
		ChemindeFerPunterMayChoose(g.GetPunterTotal())
}

// PunterDraw は子側が 3 枚目を引く。
func (g *ChemindeFer) PunterDraw() error { return g.advanceAfter(g.punterAct(true)) }

// PunterStand は子側が立つ (引かない)。
func (g *ChemindeFer) PunterStand() error { return g.advanceAfter(g.punterAct(false)) }

// advanceAfter は成功した操作のあとだけ CPU を進める。
func (g *ChemindeFer) advanceAfter(err error) error {
	if err != nil {
		return err
	}
	g.advanceCpu()
	return nil
}

// punterAct は子側の任意判断を適用する。**選べない合計では拒否する。**
func (g *ChemindeFer) punterAct(draw bool) error {
	if g.gameEndFlag {
		return errChemindeFerGameFinished
	}
	if g.phase != ChemindeFerPhasePunterDraw {
		return errChemindeFerWrongPhase
	}
	if !ChemindeFerPunterMayChoose(g.GetPunterTotal()) {
		return errChemindeFerNoChoice
	}
	if draw {
		g.punterDraw()
	} else {
		g.punterStand()
	}
	return nil
}

// punterDraw は子側に 3 枚目を配る。
func (g *ChemindeFer) punterDraw() {
	c := g.shoe.DrawCard()
	g.punterHand = append(g.punterHand, c)
	g.punterDrew = true
	g.appendLog(g.represIdx, "punterDraw", "the punter draws", []*Card{c})
	g.toBankerDraw()
}

// punterStand は子側を立たせる。
func (g *ChemindeFer) punterStand() {
	g.punterDrew = false
	g.appendLog(g.represIdx, "punterStand", "the punter stands", nil)
	g.toBankerDraw()
}

// punterDecideCpu は CPU 代表の合計 5 の判断。**定石どおり引く。**
func (g *ChemindeFer) punterDecideCpu() {
	g.punterDraw()
}

// toBankerDraw は親の判断へ進む。親は常に選べるので、CPU のときだけ自動で決める。
func (g *ChemindeFer) toBankerDraw() {
	g.phase = ChemindeFerPhaseBankerDraw
	if !g.players[g.bankerIdx].GetIsHuman() {
		g.bankerDecideCpu()
	}
}

// BankerDraw は親が 3 枚目を引く。
func (g *ChemindeFer) BankerDraw() error { return g.advanceAfter(g.bankerAct(true)) }

// BankerStand は親が立つ (引かない)。
func (g *ChemindeFer) BankerStand() error { return g.advanceAfter(g.bankerAct(false)) }

// bankerAct は親の任意判断を適用する。**親はどの合計でも自由。**
func (g *ChemindeFer) bankerAct(draw bool) error {
	if g.gameEndFlag {
		return errChemindeFerGameFinished
	}
	if g.phase != ChemindeFerPhaseBankerDraw {
		return errChemindeFerWrongPhase
	}
	if draw {
		c := g.shoe.DrawCard()
		g.bankerHand = append(g.bankerHand, c)
		g.appendLog(g.bankerIdx, "bankerDraw", "the banker draws", []*Card{c})
	} else {
		g.appendLog(g.bankerIdx, "bankerStand", "the banker stands", nil)
	}
	g.resolve()
	return nil
}

// punterThirdCard は子側が引いた 3 枚目を返す (引いていなければ nil)。
func (g *ChemindeFer) punterThirdCard() *Card {
	if !g.punterDrew || len(g.punterHand) < ChemindeFerMaxHandSize {
		return nil
	}
	return g.punterHand[ChemindeFerMaxHandSize-1]
}

// chemindeFerCpuBankerDraws は CPU 親の任意ドロー判断。
//
// **これは規則ではなく戦略である。** 同じ表がプント・バンコでは親の強制ドロー規則に
// なっているが、それはこの引き方が期待値上いちばん強いからで、シュマン・ド・フェールの
// 親はこれに従う義務が無い。人間の親は無視してよく、無視できること自体がこのゲームの
// 面白さなので、規則 (ChemindeFerPunterMustDraw など) とは別の場所に置いてある。
func chemindeFerCpuBankerDraws(bankerTotal int, punterThird *Card) bool {
	if punterThird == nil {
		// 子が立った = 子は 6 か 7 が濃厚。親は 5 以下なら引く。
		return bankerTotal <= ChemindeFerPunterFreeTotal
	}
	v := chemindeFerCardValue(punterThird)
	switch bankerTotal {
	case 0, 1, 2:
		return true
	case 3:
		return v != 8
	case 4:
		return v >= 2 && v <= 7
	case 5:
		return v >= 4 && v <= 7
	case 6:
		return v == 6 || v == 7
	default:
		return false
	}
}

// bankerDecideCpu は CPU 親の判断を適用する。
func (g *ChemindeFer) bankerDecideCpu() {
	_ = g.bankerAct(chemindeFerCpuBankerDraws(g.GetBankerTotal(), g.punterThirdCard()))
}

// --- 決着 ---

// resolve は合計を比べてチップを動かし、バンクの行方を決める。
func (g *ChemindeFer) resolve() {
	pt, bt := g.GetPunterTotal(), g.GetBankerTotal()
	switch {
	case bt > pt:
		g.result = ChemindeFerResultBanker
	case pt > bt:
		g.result = ChemindeFerResultPunter
	default:
		g.result = ChemindeFerResultTie
	}
	g.settle()
	g.phase = ChemindeFerPhaseRoundEnd
}

// settle は決着に従ってチップを動かす。
//
// **卓の総チップは動かない。** 子の賭け金は PlaceBet の時点で手元から引かれている
// ので、ここでは行き先を決めるだけ。引き分けは賭け金をそのまま返す。親が払えなく
// なることは無い ── 賭けの総額はバンク額を超えず、バンク額は張った時点の親の手持ちを
// 超えず、その間に親の手持ちは動かないため。
func (g *ChemindeFer) settle() {
	banker := g.players[g.bankerIdx]
	g.lastNet = make([]int, len(g.players))
	for i, p := range g.players {
		if i == g.bankerIdx || p.GetBet() == 0 {
			continue
		}
		bet := p.GetBet()
		p.SetBet(0)
		switch g.result {
		case ChemindeFerResultBanker:
			banker.AddChips(bet)
			// 賭け金は張った時点で手元から引かれているので、負けの純増減は -bet。
			g.lastNet[i] = -bet
			g.lastNet[g.bankerIdx] += bet
		case ChemindeFerResultPunter:
			banker.SubtractChips(bet)
			p.AddChips(bet * 2) // 賭け金の返却 + 同額の配当 (1 対 1)
			g.lastNet[i] = bet
			g.lastNet[g.bankerIdx] -= bet
		default:
			p.AddChips(bet) // 引き分け: 賭け金を返す
		}
	}
	g.appendLog(-1, "result", ChemindeFerResultName(g.result), nil)
	if g.result == ChemindeFerResultPunter {
		g.passBank()
	}
}

// GetLastNet は直前の決済での席の純増減を返す。ラウンド中は 0。
//
// **卓の結果 (banker/punter/tie) と自分の損益は別の情報。** 賭けていない席や
// 引き分けは 0 で、勝てば正、負ければ負になります。
func (g *ChemindeFer) GetLastNet(i int) int {
	if i < 0 || i >= len(g.lastNet) {
		return 0
	}
	return g.lastNet[i]
}

// passBank はバンクを次の (張れるだけのチップを持つ) 席へ渡す。
func (g *ChemindeFer) passBank() {
	n := len(g.players)
	for step := 1; step <= n; step++ {
		idx := (g.bankerIdx + step) % n
		if g.players[idx].GetChips() >= ChemindeFerStakeMin {
			g.bankerIdx = idx
			g.appendLog(idx, "bankPassed", "the bank passes", nil)
			return
		}
	}
	// 誰も張れない。ゲームはここで終わる。
	g.gameEndFlag = true
	g.appendLog(-1, "gameEnd", "nobody can bank any more", nil)
}

// PassBank は親が自分から親を降りる。ラウンド終了後のみ。
func (g *ChemindeFer) PassBank() error {
	if g.phase != ChemindeFerPhaseRoundEnd {
		return errChemindeFerWrongPhase
	}
	if g.gameEndFlag {
		return errChemindeFerGameFinished
	}
	g.passBank()
	return nil
}

// GiveUp は投了する。**卓はそこで終わる。**
//
// 終局フラグはラウンド終了フェーズとしか両立しないので、フェーズも合わせて畳む。
func (g *ChemindeFer) GiveUp() {
	if g.gameEndFlag {
		return
	}
	g.gameEndFlag = true
	g.phase = ChemindeFerPhaseRoundEnd
	g.appendLog(-1, "giveUp", "the player gives up", nil)
}

// NextRound は次のラウンドを始める。
func (g *ChemindeFer) NextRound() error {
	if g.gameEndFlag {
		return errChemindeFerGameFinished
	}
	if g.phase != ChemindeFerPhaseRoundEnd {
		return errChemindeFerWrongPhase
	}
	if g.roundNumber >= g.config.Rounds {
		g.gameEndFlag = true
		g.appendLog(-1, "gameEnd", "the session is over", nil)
		return nil
	}
	// 親が張れなくなったらバンクは自動的に隣へ。
	if g.players[g.bankerIdx].GetChips() < ChemindeFerStakeMin {
		g.passBank()
		if g.gameEndFlag {
			return nil
		}
	}
	g.roundNumber++
	g.startRound()
	g.advanceCpu()
	return nil
}

// --- CPU ---

// IsHumanTurn は人間の入力待ちかを返す。
func (g *ChemindeFer) IsHumanTurn() bool {
	if g.gameEndFlag {
		return false
	}
	switch g.phase {
	case ChemindeFerPhaseStake:
		return g.players[g.bankerIdx].GetIsHuman()
	case ChemindeFerPhaseBet:
		turn := g.GetBetTurn()
		return turn >= 0 && g.players[turn].GetIsHuman()
	case ChemindeFerPhasePunterDraw:
		return g.represIdx >= 0 && g.players[g.represIdx].GetIsHuman() && g.PunterMayChoose()
	case ChemindeFerPhaseBankerDraw:
		return g.players[g.bankerIdx].GetIsHuman()
	default:
		return false
	}
}

// chemindeFerMaxCpuSteps は 1 回の自動進行で打てる手数の上限。
//
// 1 ラウンドで要るのは張り 1 + 賭け 5 + 子 1 + 親 1 = 8 手なので十分に余裕がある。
// 上限そのものは**規則の停止性を当てにしない**ための保険で、想定外の状態で
// 無限に回るくらいなら止まったほうがよい。
const chemindeFerMaxCpuSteps = 64

// CpuPlay は人間の入力待ちになるまで CPU に打たせる。
//
// **人間の操作のあとは必ずここを通る。** 通さないと、親が張った直後に卓が止まり、
// 子が誰も賭けないまま画面が固まる。
func (g *ChemindeFer) CpuPlay() { g.advanceCpu() }

// advanceCpu は人間の番・ラウンド終了・終局のいずれかまで CPU を進める。
func (g *ChemindeFer) advanceCpu() {
	for range chemindeFerMaxCpuSteps {
		if g.gameEndFlag || g.phase == ChemindeFerPhaseRoundEnd || g.IsHumanTurn() {
			return
		}
		if !g.stepCpu() {
			return
		}
	}
}

// stepCpu は CPU に 1 手打たせる。打てなければ false。
func (g *ChemindeFer) stepCpu() bool {
	switch g.phase {
	case ChemindeFerPhaseStake:
		return g.setStake(g.cpuStake()) == nil
	case ChemindeFerPhaseBet:
		turn := g.GetBetTurn()
		if turn < 0 {
			return false
		}
		return g.placeBet(turn, g.cpuBet(turn)) == nil
	case ChemindeFerPhasePunterDraw:
		g.punterDecideCpu()
		return true
	case ChemindeFerPhaseBankerDraw:
		g.bankerDecideCpu()
		return true
	default:
		return false
	}
}

// cpuStake は CPU 親が張る額。手持ちの 1 割前後を張る。
func (g *ChemindeFer) cpuStake() int {
	lo, hi := g.StakeRangeFor(g.bankerIdx)
	if hi == 0 {
		return 0
	}
	want := g.players[g.bankerIdx].GetChips()/10 + g.rng.Intn(ChemindeFerStakeMin)
	return clampChemindeFer(want, lo, hi)
}

// cpuBet は CPU 子が賭ける額。残りバンク額を超えない範囲で手持ちの 1 割前後。
func (g *ChemindeFer) cpuBet(idx int) int {
	_, hi := g.BetRangeFor(idx)
	if hi <= 0 {
		return 0
	}
	want := g.players[idx].GetChips()/10 + g.rng.Intn(ChemindeFerStakeMin)
	if want < 1 {
		want = 1 // 張れるのに 0 を返さない。全員 0 だとラウンドが空転する。
	}
	return clampChemindeFer(want, 1, hi)
}

// clampChemindeFer は v を [lo, hi] に丸める。
func clampChemindeFer(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// --- ヒント ---

// GetHint は人間への助言を返す。判断どころでなければ nil。
func (g *ChemindeFer) GetHint() *ChemindeFerHint {
	if g.gameEndFlag || !g.IsHumanTurn() {
		return nil
	}
	switch g.phase {
	case ChemindeFerPhasePunterDraw:
		return &ChemindeFerHint{Draw: true, Reason: "punterFive"} // 合計 5 の定石は引く
	case ChemindeFerPhaseBankerDraw:
		draw := chemindeFerCpuBankerDraws(g.GetBankerTotal(), g.punterThirdCard())
		reason := "bankerStand"
		if draw {
			reason = "bankerDraw"
		}
		return &ChemindeFerHint{Draw: draw, Reason: reason}
	default:
		return nil
	}
}

// --- ログ ---

// appendLog は行動ログを 1 行足す。
func (g *ChemindeFer) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.turnNumber++
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.turnNumber,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
	if len(g.actionLog) > chemindeFerMaxSliceLen {
		g.actionLog = g.actionLog[len(g.actionLog)-chemindeFerMaxSliceLen:]
	}
}

// --- アクセサ ---

// GetPhase は現在のフェーズを返す。
func (g *ChemindeFer) GetPhase() ChemindeFerPhase { return g.phase }

// GetBankerIdx は親の席を返す。
func (g *ChemindeFer) GetBankerIdx() int { return g.bankerIdx }

// GetStake は親が張ったバンク額を返す。
func (g *ChemindeFer) GetStake() int { return g.stake }

// GetRepresentativeIdx は子側の代表の席を返す (未定なら -1)。
func (g *ChemindeFer) GetRepresentativeIdx() int { return g.represIdx }

// GetBankerHand は親の手札を返す。
func (g *ChemindeFer) GetBankerHand() []*Card { return g.bankerHand }

// GetPunterHand は子側の手札を返す。
func (g *ChemindeFer) GetPunterHand() []*Card { return g.punterHand }

// GetPunterDrew は子側が 3 枚目を引いたかを返す。
func (g *ChemindeFer) GetPunterDrew() bool { return g.punterDrew }

// GetResult はラウンドの決着を返す。
func (g *ChemindeFer) GetResult() ChemindeFerResult { return g.result }

// GetRoundNumber は現在のラウンド番号を返す。
func (g *ChemindeFer) GetRoundNumber() int { return g.roundNumber }

// GetGameEndFlag はゲームが終了したかを返す。
func (g *ChemindeFer) GetGameEndFlag() bool { return g.gameEndFlag }

// GetPlayers は席の一覧を返す。
func (g *ChemindeFer) GetPlayers() []*ChemindeFerPlayer { return g.players }

// GetPlayer は席 idx を返す (範囲外なら nil)。
func (g *ChemindeFer) GetPlayer(idx int) *ChemindeFerPlayer {
	if idx < 0 || idx >= len(g.players) {
		return nil
	}
	return g.players[idx]
}

// GetConfig は設定を返す。
func (g *ChemindeFer) GetConfig() ChemindeFerConfig { return g.config }

// SetConfig は設定を差し替える。
func (g *ChemindeFer) SetConfig(c ChemindeFerConfig) { g.config = c }

// GetActionLog は行動ログを返す。
func (g *ChemindeFer) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetRemainingCards はシューの残り枚数を返す。
func (g *ChemindeFer) GetRemainingCards() int { return g.shoe.GetRemainingCount() }

// GetTotalChips は卓のチップ総額を返す。
//
// **ゼロサムなので常に一定。** 賭け中の額は手元から引かれているので、ここで足し戻す。
func (g *ChemindeFer) GetTotalChips() int {
	total := 0
	for _, p := range g.players {
		total += p.GetChips() + p.GetBet()
	}
	return total
}

// --- 永続化 ---

// chemindeFerJSON is the JSON wire format for ChemindeFer.
type chemindeFerJSON struct {
	Shoe        *TrumpCards          `json:"sh"`
	Players     []*ChemindeFerPlayer `json:"pl"`
	Config      ChemindeFerConfig    `json:"cf"`
	Phase       int                  `json:"ph"`
	BankerIdx   int                  `json:"bi"`
	BetOrder    []int                `json:"bo"`
	BetPos      int                  `json:"bp"`
	Stake       int                  `json:"st"`
	RepresIdx   int                  `json:"ri"`
	BankerHand  []*Card              `json:"bh"`
	PunterHand  []*Card              `json:"pn"`
	PunterDrew  bool                 `json:"pd"`
	Result      int                  `json:"rs"`
	LastNet     []int                `json:"ln"`
	RoundNumber int                  `json:"rn"`
	GameEndFlag bool                 `json:"ge"`
	ActionLog   []*ActionLogEntry    `json:"al"`
	TurnNumber  int                  `json:"tn"`
}

// MarshalJSON implements json.Marshaler.
func (g *ChemindeFer) MarshalJSON() ([]byte, error) {
	return json.Marshal(chemindeFerJSON{
		Shoe:        g.shoe,
		Players:     g.players,
		Config:      g.config,
		Phase:       int(g.phase),
		BankerIdx:   g.bankerIdx,
		BetOrder:    g.betOrder,
		BetPos:      g.betPos,
		Stake:       g.stake,
		RepresIdx:   g.represIdx,
		BankerHand:  g.bankerHand,
		PunterHand:  g.punterHand,
		PunterDrew:  g.punterDrew,
		Result:      int(g.result),
		LastNet:     g.lastNet,
		RoundNumber: g.roundNumber,
		GameEndFlag: g.gameEndFlag,
		ActionLog:   g.actionLog,
		TurnNumber:  g.turnNumber,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **範囲チェックだけでは足りない。** 席番号やフェーズが個別に「範囲内」でも、
// 組み合わせとして到達できない盤面 (賭けの総額がバンク額を超えている、決着済みなのに
// フェーズが進行中、代表が親自身、3 枚目を引いたことになっているのに手札が 2 枚) は
// いくらでも作れる。壊れた保存データが静かに勝敗を変えるのを防ぐため、
// **フェーズと値の整合**まで見る。
func (g *ChemindeFer) UnmarshalJSON(data []byte) error {
	var j chemindeFerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := chemindeFerValidateSeats(j.Players); err != nil {
		return err
	}
	if err := chemindeFerValidateScalars(&j); err != nil {
		return err
	}
	if err := chemindeFerValidateConsistency(&j); err != nil {
		return err
	}

	g.shoe = j.Shoe
	if g.shoe == nil {
		g.shoe = newChemindeFerShoe()
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = ChemindeFerPhase(j.Phase)
	g.bankerIdx = j.BankerIdx
	g.betOrder = j.BetOrder
	g.betPos = j.BetPos
	g.stake = j.Stake
	g.represIdx = j.RepresIdx
	g.bankerHand = j.BankerHand
	g.punterHand = j.PunterHand
	g.punterDrew = j.PunterDrew
	g.result = ChemindeFerResult(j.Result)
	g.lastNet = j.LastNet
	if len(g.lastNet) != len(g.players) {
		// **長さが席数と合わない保存は損益を席に貼り違える。** 0 で埋め直す。
		g.lastNet = make([]int, len(g.players))
	}
	g.roundNumber = j.RoundNumber
	g.gameEndFlag = j.GameEndFlag
	g.actionLog = j.ActionLog
	g.turnNumber = j.TurnNumber
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(1))
	}
	return nil
}

// chemindeFerValidateSeats は席の枚数と欠けを検証する。
func chemindeFerValidateSeats(players []*ChemindeFerPlayer) error {
	if len(players) != ChemindeFerSeatCnt {
		return fmt.Errorf("chemindefer: seat count %d does not match %d",
			len(players), ChemindeFerSeatCnt)
	}
	for i, p := range players {
		if p == nil {
			return fmt.Errorf("chemindefer: seat %d is missing", i)
		}
	}
	return nil
}

// chemindeFerValidateScalars は単独で範囲を判定できる値を検証する。
func chemindeFerValidateScalars(j *chemindeFerJSON) error {
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < int(ChemindeFerPhaseStake) || j.Phase > int(ChemindeFerPhaseMax) {
		return fmt.Errorf("chemindefer: phase out of range: %d", j.Phase)
	}
	if j.Result < int(ChemindeFerResultNone) || j.Result > int(ChemindeFerResultMax) {
		return fmt.Errorf("chemindefer: result out of range: %d", j.Result)
	}
	if j.BankerIdx < 0 || j.BankerIdx >= ChemindeFerSeatCnt {
		return fmt.Errorf("chemindefer: banker index out of range: %d", j.BankerIdx)
	}
	if j.RepresIdx < -1 || j.RepresIdx >= ChemindeFerSeatCnt {
		return fmt.Errorf("chemindefer: representative index out of range: %d", j.RepresIdx)
	}
	if j.Stake < 0 {
		return fmt.Errorf("chemindefer: stake must not be negative: %d", j.Stake)
	}
	if j.RoundNumber < 1 || j.RoundNumber > j.Config.Rounds {
		return fmt.Errorf("chemindefer: round number out of range: %d", j.RoundNumber)
	}
	if len(j.BankerHand) > ChemindeFerMaxHandSize || len(j.PunterHand) > ChemindeFerMaxHandSize {
		return fmt.Errorf("chemindefer: a hand holds more than %d cards", ChemindeFerMaxHandSize)
	}
	if len(j.ActionLog) > chemindeFerMaxSliceLen {
		return fmt.Errorf("chemindefer: action log too long: %d", len(j.ActionLog))
	}
	return nil
}

// chemindeFerValidateConsistency は値どうしの整合を検証する。
func chemindeFerValidateConsistency(j *chemindeFerJSON) error {
	if err := chemindeFerValidateBetOrder(j); err != nil {
		return err
	}
	// **決着後は同じ席で構わない。** 子側が勝つとバンクは隣へ渡り、その隣が
	// たまたま代表だったということが普通に起きる。取り違えると、正しい盤面を
	// 壊れていると判定してラウンド終了のたびに復元が落ちる。
	if j.Phase != int(ChemindeFerPhaseRoundEnd) && j.RepresIdx == j.BankerIdx {
		return fmt.Errorf("chemindefer: seat %d is both the banker and the punter representative",
			j.BankerIdx)
	}
	total := 0
	for i, p := range j.Players {
		if i != j.BankerIdx {
			total += p.GetBet()
		}
	}
	if total > j.Stake {
		return fmt.Errorf("chemindefer: punters staked %d against a bank of %d", total, j.Stake)
	}
	if j.Phase == int(ChemindeFerPhaseBet) && j.Stake == 0 {
		return fmt.Errorf("chemindefer: betting is open but the bank is 0")
	}
	if j.Phase == int(ChemindeFerPhaseStake) &&
		(len(j.BankerHand) > 0 || len(j.PunterHand) > 0) {
		return fmt.Errorf("chemindefer: cards are dealt before the bank is staked")
	}
	if j.PunterDrew && len(j.PunterHand) < ChemindeFerMaxHandSize {
		return fmt.Errorf("chemindefer: the punter drew but holds only %d cards", len(j.PunterHand))
	}
	if j.Result != int(ChemindeFerResultNone) && j.Phase != int(ChemindeFerPhaseRoundEnd) {
		return fmt.Errorf("chemindefer: the coup is decided but the phase is %d", j.Phase)
	}
	if j.GameEndFlag && j.Phase != int(ChemindeFerPhaseRoundEnd) {
		return fmt.Errorf("chemindefer: the game-end flag and the phase disagree")
	}
	return nil
}

// chemindeFerValidateBetOrder は賭け順と現在位置を検証する。
func chemindeFerValidateBetOrder(j *chemindeFerJSON) error {
	seen := make(map[int]bool, len(j.BetOrder))
	for _, idx := range j.BetOrder {
		if idx < 0 || idx >= ChemindeFerSeatCnt {
			return fmt.Errorf("chemindefer: bet order holds seat %d", idx)
		}
		if idx == j.BankerIdx {
			return fmt.Errorf("chemindefer: the banker (seat %d) is in the bet order", idx)
		}
		if seen[idx] {
			return fmt.Errorf("chemindefer: seat %d appears twice in the bet order", idx)
		}
		seen[idx] = true
	}
	if j.BetPos < -1 || j.BetPos >= len(j.BetOrder) {
		return fmt.Errorf("chemindefer: bet position out of range: %d", j.BetPos)
	}
	return nil
}
