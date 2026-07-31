//go:build !js || !wasm || extra3

// Package domain — クラバーヤス (Klaberjass / Clobyosh / Bela) のドメインモデル。
//
// ジャス系の **2 人用**祖型。32 枚 (7〜A) から **9 枚ずつしか配らない**。
//
// # issue #4395 の仕様案との相違
//
//   - issue は「合計 163 点相当を両者で争う」とするが、**二重に誤り**。
//     32 枚の総点は **162** (カード 152 + 最終トリック 10) であり、しかも
//     2 人戦では 32 枚のうち **18〜19 枚しか場に出ない**。残りは死に札なので、
//     1 ディールで争う点数は**配りごとに変わる**。固定値ではない
//   - issue は **dix (切札の 7 の交換)** に触れていない。切札が表向きカードの
//     スートに決まったとき、切札の 7 を持つ側は**表向きカードと交換できる**
//   - issue は **schmeiss (投げ)** に触れていない。ビッド時に「この配りを流す」
//     と提案でき、相手が拒めば提案者が自動的にメイカーになる
//   - issue はシーケンスの点数を書いていない。**3 枚 = 20 点、4 枚以上 = 50 点**。
//     良い方だけが得点し、しかも**その人の役は全部**数える
//   - issue は「ベラは切札 K+Q を**連続して出す**と成立」とするが、連続して
//     出す必要はない。K と Q を出すときにそれぞれ宣言する
//   - issue は目標点を書いていない。**501 点**
//
// # 実装上の注意
//
// **シーケンスは点数順ではなく素のランク順で数える。**切札の J と 9 が
// 20 点 / 14 点で最上位になるのはトリックの強さの話であって、並びは
// 7-8-9-10-J-Q-K-A のままである。
package domain

import (
	"encoding/json"
	"fmt"
	"sort"
)

// KlaberjassPlayerCnt はプレイヤー数 (2 人)。
const KlaberjassPlayerCnt = 2

// KlaberjassHandSize は最終的な手札枚数。
const KlaberjassHandSize = 9

// KlaberjassFirstDealSize は表向きカードを出す前に配る枚数。
const KlaberjassFirstDealSize = 6

// Klaberjass のボーナス点。
const (
	// KlaberjassLastTrickBonus は最終トリックを取った側の加点。
	KlaberjassLastTrickBonus = 10
	// KlaberjassBelaBonus は切札 K+Q を持っていた側の加点。
	KlaberjassBelaBonus = 20
	// KlaberjassTerzPoints は 3 枚のシーケンスの点数。
	KlaberjassTerzPoints = 20
	// KlaberjassFiftyPoints は 4 枚以上のシーケンスの点数。
	//
	// **5 枚以上でも 50 点のまま**で、長さに応じて増えたりはしない。
	KlaberjassFiftyPoints = 50
)

// KlaberjassCardPointsTotal は 32 枚すべてのカード点の合計。
//
// **1 ディールで争う点数ではない。**2 人戦では 18〜19 枚しか配られないので、
// 実際に場に出る点数は配りごとに変わる。issue #4395 の「163 点」は、この総点
// (162 = 152 + 最終トリック 10) とも、実際に争う点とも一致しない。
const KlaberjassCardPointsTotal = 152

// KlaberjassPhase はゲームフェーズ。
type KlaberjassPhase int

// Klaberjass のフェーズ定数
const (
	// KlaberjassPhaseBidTurnUp 第 1 ラウンド (表向きカードのスートを取るか)
	KlaberjassPhaseBidTurnUp KlaberjassPhase = iota
	// KlaberjassPhaseBidFree 第 2 ラウンド (好きなスートを指名するか)
	KlaberjassPhaseBidFree
	// KlaberjassPhaseSchmeiss 投げの提案に相手が答える
	KlaberjassPhaseSchmeiss
	// KlaberjassPhasePlay トリックプレイ
	KlaberjassPhasePlay
	// KlaberjassPhaseHandEnd ディール終了 (精算済み)
	KlaberjassPhaseHandEnd
	// KlaberjassPhaseGameEnd ゲーム終了
	KlaberjassPhaseGameEnd
)

// KlaberjassSequence は申告されたシーケンス役。
type KlaberjassSequence struct {
	// Suit はスート。
	Suit int
	// TopValue は最上位札の値 (A は 1 ではなく 14 として持つ)。
	TopValue int
	// Length は長さ。
	Length int
	// Points は点数 (20 または 50)。
	Points int
}

// Klaberjass はクラバーヤスのゲームクラス。
type Klaberjass struct {
	trumpCards *TrumpCards
	players    []*KlaberjassPlayer
	config     KlaberjassConfig
	phase      KlaberjassPhase

	dealerIdx  int
	currentIdx int
	// bidIdx はビッド中の手番。
	bidIdx int
	// bidPassCount は現ラウンドのパス数。
	bidPassCount int
	// schmeissBy は投げを提案した席 (-1 なら提案なし)。
	schmeissBy int

	trumpSuit  int
	turnUpCard *Card
	// makerIdx は切札を決めた側 (「宣言側」)。
	makerIdx int

	// trick は場に出ている札 (最大 2 枚)。
	trick []*Card
	// trickLeader はこのトリックのリード席。
	trickLeader int
	trickNumber int

	// handPoints はこのディールで各席が取った点 (役・ベラ・最終トリック込み)。
	handPoints [KlaberjassPlayerCnt]int
	// sequences は各席が持っていたシーケンス。
	sequences [KlaberjassPlayerCnt][]*KlaberjassSequence
	// sequenceWinner はシーケンス勝負に勝った席 (-1 なら誰も得点しない)。
	sequenceWinner int
	// belaHolder は切札 K+Q を持っていた席 (-1 ならなし)。
	belaHolder int
	// belaKingPlayed / belaQueenPlayed は宣言の進捗。
	belaKingPlayed  bool
	belaQueenPlayed bool
	// belaScored はベラが成立したか。
	belaScored bool
	// dixUsed は切札の 7 の交換が行われたか。
	dixUsed bool
	// beteFlag は直前のディールでメイカーがベートしたか。
	beteFlag bool
	// lastTrickWinner は最終トリックを取った席。
	lastTrickWinner int

	scores      [KlaberjassPlayerCnt]int
	dealNumber  int
	gameEndFlag bool
	winnerIdx   int

	actionLog []*ActionLogEntry
}

// NewKlaberjass コンストラクタ
func NewKlaberjass(trumpCards *TrumpCards, players []*KlaberjassPlayer, config KlaberjassConfig) *Klaberjass {
	return &Klaberjass{trumpCards: trumpCards, players: players, config: config, winnerIdx: -1}
}

// NewDefaultKlaberjass はデフォルト構成のゲームを返す。
func NewDefaultKlaberjass() *Klaberjass {
	players := make([]*KlaberjassPlayer, KlaberjassPlayerCnt)
	for i := range players {
		players[i] = NewKlaberjassPlayer(i == 0)
	}
	return NewKlaberjass(NewTrumpCardsBelote(), players, DefaultKlaberjassConfig())
}

// ---- ランクと点数 ----

// klaberjassSeqRank は**並びを数えるための**ランクを返す。
//
// **点数順ではない。**切札の J と 9 がトリックで最強になるのは強さの話で、
// 並びは 7-8-9-10-J-Q-K-A のまま。A は値 1 なので 14 に読み替える。
func klaberjassSeqRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// KlaberjassCardPoints は札の点数を返す。
func KlaberjassCardPoints(c *Card, trumpSuit int) int {
	if c == nil {
		return 0
	}
	if c.GetDesign() == trumpSuit {
		switch c.GetValue() {
		case 11: // Jass
			return 20
		case 9: // Menel
			return 14
		case 1:
			return 11
		case 10:
			return 10
		case 13:
			return 4
		case 12:
			return 3
		}
		return 0
	}
	switch c.GetValue() {
	case 1:
		return 11
	case 10:
		return 10
	case 13:
		return 4
	case 12:
		return 3
	case 11:
		return 2
	}
	return 0
}

// klaberjassTrickRank はトリックの強さを返す。
//
// **切札だけ順序が違う。**J (20) と 9 (14) が A の上に割り込む。
func klaberjassTrickRank(c *Card, trumpSuit int) int {
	if c == nil {
		return 0
	}
	if c.GetDesign() == trumpSuit {
		switch c.GetValue() {
		case 11:
			return 8
		case 9:
			return 7
		case 1:
			return 6
		case 10:
			return 5
		case 13:
			return 4
		case 12:
			return 3
		case 8:
			return 2
		}
		return 1
	}
	switch c.GetValue() {
	case 1:
		return 8
	case 10:
		return 7
	case 13:
		return 6
	case 12:
		return 5
	case 11:
		return 4
	case 9:
		return 3
	case 8:
		return 2
	}
	return 1
}

// ---- 進行 ----

// Reset ゲーム初期化
func (k *Klaberjass) Reset() {
	k.gameEndFlag = false
	k.winnerIdx = -1
	k.scores = [KlaberjassPlayerCnt]int{}
	k.dealNumber = 0
	k.dealerIdx = 0
	k.actionLog = nil
	k.beginDeal()
}

// beginDeal は 1 ディールを配ってビッドへ入る。
func (k *Klaberjass) beginDeal() {
	k.dealNumber++
	k.phase = KlaberjassPhaseBidTurnUp
	k.trumpSuit = 0
	k.makerIdx = -1
	k.trick = nil
	k.trickNumber = 0
	k.trickLeader = -1
	k.handPoints = [KlaberjassPlayerCnt]int{}
	k.sequences = [KlaberjassPlayerCnt][]*KlaberjassSequence{}
	k.sequenceWinner = -1
	k.belaHolder = -1
	k.belaKingPlayed = false
	k.belaQueenPlayed = false
	k.belaScored = false
	k.dixUsed = false
	k.beteFlag = false
	k.lastTrickWinner = -1
	k.bidPassCount = 0
	k.schmeissBy = -1

	for _, p := range k.players {
		p.ResetRound()
	}

	// **毎ディール山札を丸ごと戻す。**9 枚ずつしか配らないので、戻さないと
	// 2 ディール目で札が尽きる。
	k.trumpCards.Replenish()
	k.trumpCards.Shuffle()
	for range KlaberjassFirstDealSize {
		for i := range KlaberjassPlayerCnt {
			if c := k.trumpCards.DrawCard(); c != nil {
				k.players[i].AddCard(c)
			}
		}
	}
	k.turnUpCard = k.trumpCards.DrawCard()
	// **非ディーラーから。**2 人戦なのでディーラーの相手が先に判断する。
	k.bidIdx = k.opponentOf(k.dealerIdx)
	k.addLog(-1, "deal", "6 cards each with a card turned up", nil)
}

// opponentOf は相手の席を返す。
func (k *Klaberjass) opponentOf(idx int) int { return (idx + 1) % KlaberjassPlayerCnt }

// checkBidTurn はビッドできる状態かを確かめる。
func (k *Klaberjass) checkBidTurn(player int) error {
	if k.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if k.phase != KlaberjassPhaseBidTurnUp && k.phase != KlaberjassPhaseBidFree {
		return fmt.Errorf("bidding is not in progress")
	}
	if player != k.bidIdx {
		return fmt.Errorf("it is not player %d's turn to bid", player)
	}
	return nil
}

// AcceptTrump は表向きカードのスートを切札にする (第 1 ラウンド)。
func (k *Klaberjass) AcceptTrump(player int) error {
	if err := k.checkBidTurn(player); err != nil {
		return err
	}
	if k.phase != KlaberjassPhaseBidTurnUp {
		return fmt.Errorf("the turn-up suit is no longer on offer")
	}
	if k.turnUpCard == nil {
		return fmt.Errorf("there is no turn-up card")
	}
	k.settleTrump(player, k.turnUpCard.GetDesign())
	return nil
}

// CallTrump は好きなスートを切札に指名する (第 2 ラウンド)。
//
// **表向きカードのスートは選べない。**それは第 1 ラウンドで流れている。
func (k *Klaberjass) CallTrump(player, suit int) error {
	if err := k.checkBidTurn(player); err != nil {
		return err
	}
	if k.phase != KlaberjassPhaseBidFree {
		return fmt.Errorf("the free choice round has not started")
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return fmt.Errorf("bad suit: %d", suit)
	}
	if k.turnUpCard != nil && suit == k.turnUpCard.GetDesign() {
		return fmt.Errorf("the turn-up suit was already refused")
	}
	k.settleTrump(player, suit)
	return nil
}

// Pass はビッドを見送る。
func (k *Klaberjass) Pass(player int) error {
	if err := k.checkBidTurn(player); err != nil {
		return err
	}
	k.bidPassCount++
	k.addLog(player, "pass", "passes", nil)
	if k.bidPassCount >= 2 {
		if k.phase == KlaberjassPhaseBidTurnUp {
			// 両者が表向きのスートを断ったら、好きなスートを選べる第 2 ラウンドへ。
			k.phase = KlaberjassPhaseBidFree
			k.bidPassCount = 0
			k.bidIdx = k.opponentOf(k.dealerIdx)
			return nil
		}
		// **両ラウンドとも流れたら配り直し。**手札を全部戻して同じディーラーで
		// もう一度配る。
		k.redeal()
		return nil
	}
	k.bidIdx = k.opponentOf(k.bidIdx)
	return nil
}

// Schmeiss は「この配りを流そう」と提案する。
//
// **相手が拒めば提案者がメイカーになる。**投げ得にはならない。
func (k *Klaberjass) Schmeiss(player int) error {
	if err := k.checkBidTurn(player); err != nil {
		return err
	}
	if !k.config.AllowSchmeiss {
		return fmt.Errorf("schmeiss is disabled")
	}
	if k.schmeissBy >= 0 {
		return fmt.Errorf("a schmeiss is already pending")
	}
	k.schmeissBy = player
	k.phase = KlaberjassPhaseSchmeiss
	k.bidIdx = k.opponentOf(player)
	k.addLog(player, "schmeiss", "offers to throw the deal in", nil)
	return nil
}

// AnswerSchmeiss は投げの提案に答える。accept なら配り直し、拒否なら提案者が
// メイカーになる。
func (k *Klaberjass) AnswerSchmeiss(player int, accept bool) error {
	if k.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if k.phase != KlaberjassPhaseSchmeiss {
		return fmt.Errorf("no schmeiss is pending")
	}
	if player != k.bidIdx {
		return fmt.Errorf("it is not player %d's decision", player)
	}
	thrower := k.schmeissBy
	if accept {
		k.addLog(player, "schmeiss_accept", "agrees to throw the deal in", nil)
		k.redeal()
		return nil
	}
	k.addLog(player, "schmeiss_refuse", "refuses the throw", nil)
	if k.turnUpCard == nil {
		return fmt.Errorf("there is no turn-up card")
	}
	k.settleTrump(thrower, k.turnUpCard.GetDesign())
	return nil
}

// redeal は同じディーラーで配り直す。
//
// **手札は捨てて山札を丸ごと戻す。**両ラウンドとも流れた配りは無かったことに
// なるので、ディール番号も進めない。
func (k *Klaberjass) redeal() {
	k.addLog(-1, "redeal", "nobody took the deal", nil)
	k.discardHands()
	k.turnUpCard = nil
	k.dealNumber--
	k.beginDeal()
}

// discardHands は全員の手札を捨てる。
func (k *Klaberjass) discardHands() {
	for _, p := range k.players {
		for p.GetCardsSize() > 0 {
			p.RemoveCard(0)
		}
	}
}

// settleTrump は切札を確定し、残りを配ってプレイへ入る。
func (k *Klaberjass) settleTrump(maker, suit int) {
	k.trumpSuit = suit
	k.makerIdx = maker
	k.schmeissBy = -1
	k.addLog(maker, "make_trump", fmt.Sprintf("makes suit %d trump", suit), nil)

	k.applyDix()
	k.dealRemainder()
	k.collectSequences()
	k.findBela()

	k.phase = KlaberjassPhasePlay
	// **リードは非ディーラー。**切札を決めた側ではない。
	k.trickLeader = k.opponentOf(k.dealerIdx)
	k.currentIdx = k.trickLeader
}

// applyDix は切札の 7 と表向きカードの交換を行う。
//
// **切札が表向きカードのスートに決まったときだけ。**第 2 ラウンドで別のスートに
// なった場合は交換できない。
func (k *Klaberjass) applyDix() {
	if k.turnUpCard == nil || k.turnUpCard.GetDesign() != k.trumpSuit {
		return
	}
	for i, p := range k.players {
		for j := range p.GetCardsSize() {
			c := p.GetCard(j)
			if c == nil || c.GetDesign() != k.trumpSuit || c.GetValue() != 7 {
				continue
			}
			p.RemoveCard(j)
			p.AddCard(k.turnUpCard)
			k.turnUpCard = c
			k.dixUsed = true
			k.addLog(i, "dix", "exchanges the trump seven for the turn-up", nil)
			return
		}
	}
}

// dealRemainder は 3 枚ずつ配って 9 枚にする。
//
// **表向きカードは配られない。**dix で交換された場合を除き、山札の残りとともに
// このディールでは死に札になる。
func (k *Klaberjass) dealRemainder() {
	for range KlaberjassHandSize - KlaberjassFirstDealSize {
		for i := range KlaberjassPlayerCnt {
			if c := k.trumpCards.DrawCard(); c != nil {
				k.players[i].AddCard(c)
			}
		}
	}
}

// ---- シーケンスとベラ ----

// klaberjassHandOf は手札をスライスで返す。
func klaberjassHandOf(p *KlaberjassPlayer) []*Card {
	out := make([]*Card, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		out = append(out, p.GetCard(i))
	}
	return out
}

// klaberjassFindSequences は手札から最長のシーケンス群を返す。
//
// 同じスートの連続 3 枚以上。**並びは素のランク順**で数える。
func klaberjassFindSequences(cards []*Card) []*KlaberjassSequence {
	bySuit := map[int][]int{}
	for _, c := range cards {
		if c == nil {
			continue
		}
		bySuit[c.GetDesign()] = append(bySuit[c.GetDesign()], klaberjassSeqRank(c))
	}
	var out []*KlaberjassSequence
	suits := make([]int, 0, len(bySuit))
	for s := range bySuit {
		suits = append(suits, s)
	}
	sort.Ints(suits)
	for _, s := range suits {
		ranks := bySuit[s]
		sort.Ints(ranks)
		run := 1
		for i := 1; i <= len(ranks); i++ {
			if i < len(ranks) && ranks[i] == ranks[i-1]+1 {
				run++
				continue
			}
			if run >= 3 {
				out = append(out, &KlaberjassSequence{
					Suit:     s,
					TopValue: ranks[i-1],
					Length:   run,
					Points:   klaberjassSequencePoints(run),
				})
			}
			run = 1
		}
	}
	return out
}

// klaberjassSequencePoints は長さから点数を返す。
func klaberjassSequencePoints(length int) int {
	if length >= 4 {
		return KlaberjassFiftyPoints
	}
	if length >= 3 {
		return KlaberjassTerzPoints
	}
	return 0
}

// klaberjassBestSequence は一番強いシーケンスを返す。
//
// 長い方が強く、同じ長さなら最上位札が高い方が強い。
func klaberjassBestSequence(seqs []*KlaberjassSequence) *KlaberjassSequence {
	var best *KlaberjassSequence
	for _, s := range seqs {
		if best == nil || s.Length > best.Length ||
			(s.Length == best.Length && s.TopValue > best.TopValue) {
			best = s
		}
	}
	return best
}

// collectSequences は両者の役を比べ、**良い方だけ**に得点させる。
//
// **勝った側は自分の役を全部数える。**1 つだけではない。引き分けたときは
// **どちらも得点しない**。
func (k *Klaberjass) collectSequences() {
	for i, p := range k.players {
		k.sequences[i] = klaberjassFindSequences(klaberjassHandOf(p))
	}
	a := klaberjassBestSequence(k.sequences[0])
	b := klaberjassBestSequence(k.sequences[1])
	winner := -1
	switch {
	case a != nil && b == nil:
		winner = 0
	case b != nil && a == nil:
		winner = 1
	case a != nil && b != nil:
		if a.Length > b.Length || (a.Length == b.Length && a.TopValue > b.TopValue) {
			winner = 0
		} else if b.Length > a.Length || (b.Length == a.Length && b.TopValue > a.TopValue) {
			winner = 1
		}
	}
	k.sequenceWinner = winner
	if winner < 0 {
		return
	}
	total := 0
	for _, s := range k.sequences[winner] {
		total += s.Points
	}
	k.handPoints[winner] += total
	k.addLog(winner, "sequence", fmt.Sprintf("scores %d for sequences", total), nil)
}

// findBela は切札の K と Q を両方持っている席を記録する。
func (k *Klaberjass) findBela() {
	for i, p := range k.players {
		king, queen := false, false
		for j := range p.GetCardsSize() {
			c := p.GetCard(j)
			if c == nil || c.GetDesign() != k.trumpSuit {
				continue
			}
			if c.GetValue() == 13 {
				king = true
			}
			if c.GetValue() == 12 {
				queen = true
			}
		}
		if king && queen {
			k.belaHolder = i
			return
		}
	}
}

// ---- プレイ ----

// KlaberjassValidPlays は player が出せる手札インデックスを返す。
//
// **追随も切札も上乗せも強制。**フォローできなければ切札を出さねばならず、
// 切札がリードされたら勝てる切札があるかぎり上乗せしなければならない。
func (k *Klaberjass) KlaberjassValidPlays(player int) []int {
	p := k.GetPlayer(player)
	if p == nil {
		return nil
	}
	all := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		all = append(all, i)
	}
	if len(k.trick) == 0 {
		return all
	}
	lead := k.trick[0]
	if lead == nil {
		return all
	}
	leadSuit := lead.GetDesign()

	same := k.indexesOfSuit(p, leadSuit)
	if len(same) > 0 {
		if leadSuit != k.trumpSuit {
			return same
		}
		// 切札リードには勝てる切札があれば上乗せ。
		higher := make([]int, 0, len(same))
		for _, i := range same {
			if klaberjassTrickRank(p.GetCard(i), k.trumpSuit) > klaberjassTrickRank(lead, k.trumpSuit) {
				higher = append(higher, i)
			}
		}
		if len(higher) > 0 {
			return higher
		}
		return same
	}

	trumps := k.indexesOfSuit(p, k.trumpSuit)
	if len(trumps) == 0 {
		return all
	}
	// 相手が既に切っているなら、勝てる切札があればそれで上乗せ。
	best := -1
	for _, c := range k.trick {
		if c != nil && c.GetDesign() == k.trumpSuit {
			if r := klaberjassTrickRank(c, k.trumpSuit); r > best {
				best = r
			}
		}
	}
	if best < 0 {
		return trumps
	}
	higher := make([]int, 0, len(trumps))
	for _, i := range trumps {
		if klaberjassTrickRank(p.GetCard(i), k.trumpSuit) > best {
			higher = append(higher, i)
		}
	}
	if len(higher) > 0 {
		return higher
	}
	return trumps
}

// indexesOfSuit は指定スートの手札インデックスを返す。
func (k *Klaberjass) indexesOfSuit(p *KlaberjassPlayer, suit int) []int {
	var out []int
	for i := range p.GetCardsSize() {
		if c := p.GetCard(i); c != nil && c.GetDesign() == suit {
			out = append(out, i)
		}
	}
	return out
}

// PlayCard は 1 枚出す。
func (k *Klaberjass) PlayCard(player, idx int) error {
	if k.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if k.phase != KlaberjassPhasePlay {
		return fmt.Errorf("the play phase is not in progress")
	}
	if player != k.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	p := k.GetPlayer(player)
	if p == nil || idx < 0 || idx >= p.GetCardsSize() {
		return fmt.Errorf("bad card index: %d", idx)
	}
	valid := k.KlaberjassValidPlays(player)
	if !klaberjassContains(valid, idx) {
		return fmt.Errorf("that card may not be played")
	}

	card := p.GetCard(idx)
	p.RemoveCard(idx)
	k.trick = append(k.trick, card)
	k.noteBela(player, card)
	k.addLog(player, "play", "plays a card", []*Card{card})

	if len(k.trick) < KlaberjassPlayerCnt {
		k.currentIdx = k.opponentOf(player)
		return nil
	}
	k.resolveTrick()
	return nil
}

// noteBela は切札 K / Q が出たことを記録し、両方出たらベラを成立させる。
//
// **連続して出す必要はない。**issue の «連続して出すと» は誤り。
func (k *Klaberjass) noteBela(player int, card *Card) {
	if k.belaHolder != player || card == nil || card.GetDesign() != k.trumpSuit {
		return
	}
	switch card.GetValue() {
	case 13:
		k.belaKingPlayed = true
	case 12:
		k.belaQueenPlayed = true
	default:
		return
	}
	if k.belaKingPlayed && k.belaQueenPlayed && !k.belaScored {
		k.belaScored = true
		k.handPoints[player] += KlaberjassBelaBonus
		k.addLog(player, "bela", "declares bela", nil)
	}
}

// resolveTrick はトリックを解決する。
func (k *Klaberjass) resolveTrick() {
	lead := k.trick[0]
	winnerOffset := 0
	bestRank := klaberjassTrickRank(lead, k.trumpSuit)
	leadSuit := lead.GetDesign()
	for i := 1; i < len(k.trick); i++ {
		c := k.trick[i]
		if c == nil {
			continue
		}
		beats := false
		switch {
		case c.GetDesign() == k.trumpSuit && leadSuit != k.trumpSuit:
			beats = true
		case c.GetDesign() == leadSuit || (c.GetDesign() == k.trumpSuit && leadSuit == k.trumpSuit):
			beats = klaberjassTrickRank(c, k.trumpSuit) > bestRank
		}
		if beats {
			winnerOffset = i
			bestRank = klaberjassTrickRank(c, k.trumpSuit)
			leadSuit = c.GetDesign()
		}
	}
	winner := (k.trickLeader + winnerOffset) % KlaberjassPlayerCnt

	points := 0
	for _, c := range k.trick {
		points += KlaberjassCardPoints(c, k.trumpSuit)
	}
	k.handPoints[winner] += points
	k.lastTrickWinner = winner
	k.trickNumber++
	k.trick = nil
	k.trickLeader = winner
	k.currentIdx = winner
	k.addLog(winner, "trick", fmt.Sprintf("takes the trick for %d", points), nil)

	if k.players[0].GetCardsSize() == 0 && k.players[1].GetCardsSize() == 0 {
		k.finishHand()
	}
}

// finishHand は最終トリックを加点してディールを精算する。
func (k *Klaberjass) finishHand() {
	if k.lastTrickWinner >= 0 {
		k.handPoints[k.lastTrickWinner] += KlaberjassLastTrickBonus
	}
	maker := k.makerIdx
	opp := k.opponentOf(maker)

	// **メイカーは相手より「多く」取らねばならない。**同点はベート。
	if k.handPoints[maker] > k.handPoints[opp] {
		k.scores[maker] += k.handPoints[maker]
		k.scores[opp] += k.handPoints[opp]
		k.addLog(maker, "hand_end", fmt.Sprintf("makes it with %d to %d", k.handPoints[maker], k.handPoints[opp]), nil)
	} else {
		k.beteFlag = true
		k.scores[opp] += k.handPoints[maker] + k.handPoints[opp]
		k.addLog(maker, "bete", fmt.Sprintf("goes bete; %d points go to the opponent", k.handPoints[maker]), nil)
	}

	k.phase = KlaberjassPhaseHandEnd
	k.checkGameEnd()
}

// checkGameEnd は目標点に届いていれば決着させる。
func (k *Klaberjass) checkGameEnd() {
	target := k.config.TargetScore
	if k.scores[0] < target && k.scores[1] < target {
		return
	}
	// 両者同時に超えたら**メイカー側が優先**する。
	switch {
	case k.scores[0] >= target && k.scores[1] >= target:
		if k.scores[0] > k.scores[1] {
			k.winnerIdx = 0
		} else if k.scores[1] > k.scores[0] {
			k.winnerIdx = 1
		} else {
			k.winnerIdx = k.makerIdx
		}
	case k.scores[0] >= target:
		k.winnerIdx = 0
	default:
		k.winnerIdx = 1
	}
	k.gameEndFlag = true
	k.phase = KlaberjassPhaseGameEnd
	k.addLog(k.winnerIdx, "game_end", "wins the game", nil)
}

// NextDeal は次のディールを配る。
func (k *Klaberjass) NextDeal() error {
	if k.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if k.phase != KlaberjassPhaseHandEnd {
		return fmt.Errorf("the deal is still in progress")
	}
	k.discardHands()
	k.dealerIdx = k.opponentOf(k.dealerIdx)
	k.beginDeal()
	return nil
}

// ---- CPU ----

// KlaberjassCpuBid は CPU のビッドを決める。
//
// 切札候補のスートを 3 枚以上持っていれば取り、第 2 ラウンドでは一番長い
// スートを指名する。
func (k *Klaberjass) KlaberjassCpuBid(idx int) (action string, suit int) {
	p := k.GetPlayer(idx)
	if p == nil {
		return "pass", 0
	}
	counts := map[int]int{}
	for i := range p.GetCardsSize() {
		if c := p.GetCard(i); c != nil {
			counts[c.GetDesign()]++
		}
	}
	if k.phase == KlaberjassPhaseBidTurnUp {
		if k.turnUpCard != nil && counts[k.turnUpCard.GetDesign()] >= 3 {
			return "accept", k.turnUpCard.GetDesign()
		}
		return "pass", 0
	}
	bestSuit, bestCount := 0, 0
	for s := CardDesignSpade; s <= CardDesignDiamond; s++ {
		if k.turnUpCard != nil && s == k.turnUpCard.GetDesign() {
			continue
		}
		if counts[s] > bestCount {
			bestSuit, bestCount = s, counts[s]
		}
	}
	// **ディーラーは最後の砦。**ここで降りると配り直しになるので、多少弱くても取る。
	if bestCount >= 3 || (idx == k.dealerIdx && bestSuit != 0) {
		return "call", bestSuit
	}
	return "pass", 0
}

// KlaberjassCpuPlay は CPU が出す手札インデックスを返す。
//
// 出せる中で、勝てるなら一番高い札、勝てないなら一番安い札。
func (k *Klaberjass) KlaberjassCpuPlay(idx int) int {
	valid := k.KlaberjassValidPlays(idx)
	if len(valid) == 0 {
		return -1
	}
	p := k.GetPlayer(idx)
	if p == nil {
		return valid[0]
	}
	if len(k.trick) == 0 {
		// リードは一番点の低い札から。
		best, bestPts := valid[0], 1<<30
		for _, i := range valid {
			if pts := KlaberjassCardPoints(p.GetCard(i), k.trumpSuit); pts < bestPts {
				best, bestPts = i, pts
			}
		}
		return best
	}
	lead := k.trick[0]
	winning, winPts := -1, -1
	cheap, cheapPts := valid[0], 1<<30
	for _, i := range valid {
		c := p.GetCard(i)
		pts := KlaberjassCardPoints(c, k.trumpSuit)
		if klaberjassBeats(c, lead, k.trumpSuit) && pts > winPts {
			winning, winPts = i, pts
		}
		if pts < cheapPts {
			cheap, cheapPts = i, pts
		}
	}
	if winning >= 0 {
		return winning
	}
	return cheap
}

// klaberjassBeats は c が lead に勝つかを返す。
func klaberjassBeats(c, lead *Card, trumpSuit int) bool {
	if c == nil || lead == nil {
		return false
	}
	if c.GetDesign() == trumpSuit && lead.GetDesign() != trumpSuit {
		return true
	}
	if c.GetDesign() != lead.GetDesign() {
		return false
	}
	return klaberjassTrickRank(c, trumpSuit) > klaberjassTrickRank(lead, trumpSuit)
}

// IsHumanTurn は今が人間の手番かを返す。
func (k *Klaberjass) IsHumanTurn() bool {
	if k.gameEndFlag {
		return false
	}
	switch k.phase {
	case KlaberjassPhaseBidTurnUp, KlaberjassPhaseBidFree, KlaberjassPhaseSchmeiss:
		p := k.GetPlayer(k.bidIdx)
		return p != nil && p.GetIsHuman()
	case KlaberjassPhasePlay:
		p := k.GetPlayer(k.currentIdx)
		return p != nil && p.GetIsHuman()
	}
	return false
}

// CpuPlay は今の手番の CPU に 1 手打たせる。
func (k *Klaberjass) CpuPlay() {
	if k.gameEndFlag {
		return
	}
	switch k.phase {
	case KlaberjassPhaseBidTurnUp, KlaberjassPhaseBidFree:
		idx := k.bidIdx
		if p := k.GetPlayer(idx); p == nil || p.GetIsHuman() {
			return
		}
		action, suit := k.KlaberjassCpuBid(idx)
		switch action {
		case "accept":
			_ = k.AcceptTrump(idx)
		case "call":
			_ = k.CallTrump(idx, suit)
		default:
			_ = k.Pass(idx)
		}
	case KlaberjassPhaseSchmeiss:
		idx := k.bidIdx
		if p := k.GetPlayer(idx); p == nil || p.GetIsHuman() {
			return
		}
		// CPU は投げに乗る。拒めばメイカーを押し付けられる側になる。
		_ = k.AnswerSchmeiss(idx, true)
	case KlaberjassPhasePlay:
		idx := k.currentIdx
		if p := k.GetPlayer(idx); p == nil || p.GetIsHuman() {
			return
		}
		if i := k.KlaberjassCpuPlay(idx); i >= 0 {
			_ = k.PlayCard(idx, i)
		}
	}
}

// klaberjassContains は s に v が含まれるかを返す。
func klaberjassContains(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (k *Klaberjass) GetPlayers() []*KlaberjassPlayer { return k.players }

// GetPlayer は idx のプレイヤーを返す。
func (k *Klaberjass) GetPlayer(idx int) *KlaberjassPlayer {
	if idx < 0 || idx >= len(k.players) {
		return nil
	}
	return k.players[idx]
}

// GetPhase は現在のフェーズを返す。
func (k *Klaberjass) GetPhase() KlaberjassPhase { return k.phase }

// GetCurrentPlayerIdx は現在の手番を返す。
func (k *Klaberjass) GetCurrentPlayerIdx() int { return k.currentIdx }

// GetBidPlayerIdx はビッド中の手番を返す。
func (k *Klaberjass) GetBidPlayerIdx() int { return k.bidIdx }

// GetDealerIdx はディーラーを返す。
func (k *Klaberjass) GetDealerIdx() int { return k.dealerIdx }

// GetTrumpSuit は切札スートを返す (0 なら未確定)。
func (k *Klaberjass) GetTrumpSuit() int { return k.trumpSuit }

// GetTurnUpCard は表向きカードを返す。
func (k *Klaberjass) GetTurnUpCard() *Card { return k.turnUpCard }

// GetMakerIdx は切札を決めた席を返す (-1 なら未確定)。
func (k *Klaberjass) GetMakerIdx() int { return k.makerIdx }

// GetTrick は場に出ている札を返す。
func (k *Klaberjass) GetTrick() []*Card { return k.trick }

// GetTrickLeaderIdx はこのトリックのリード席を返す。
func (k *Klaberjass) GetTrickLeaderIdx() int { return k.trickLeader }

// GetTrickNumber は済んだトリック数を返す。
func (k *Klaberjass) GetTrickNumber() int { return k.trickNumber }

// GetHandPoints は idx がこのディールで取った点を返す。
func (k *Klaberjass) GetHandPoints(idx int) int {
	if idx < 0 || idx >= KlaberjassPlayerCnt {
		return 0
	}
	return k.handPoints[idx]
}

// GetSequences は idx のシーケンス役を返す。
func (k *Klaberjass) GetSequences(idx int) []*KlaberjassSequence {
	if idx < 0 || idx >= KlaberjassPlayerCnt {
		return nil
	}
	return k.sequences[idx]
}

// GetSequenceWinner はシーケンス勝負に勝った席を返す (-1 なら得点なし)。
func (k *Klaberjass) GetSequenceWinner() int { return k.sequenceWinner }

// GetBelaHolder は切札 K+Q を持っていた席を返す (-1 ならなし)。
func (k *Klaberjass) GetBelaHolder() int { return k.belaHolder }

// IsBelaScored はベラが成立したかを返す。
func (k *Klaberjass) IsBelaScored() bool { return k.belaScored }

// IsDixUsed は切札の 7 の交換が行われたかを返す。
func (k *Klaberjass) IsDixUsed() bool { return k.dixUsed }

// IsBete は直前のディールでメイカーがベートしたかを返す。
func (k *Klaberjass) IsBete() bool { return k.beteFlag }

// GetSchmeissBy は投げを提案した席を返す (-1 なら提案なし)。
func (k *Klaberjass) GetSchmeissBy() int { return k.schmeissBy }

// GetScore は idx の通算点を返す。
func (k *Klaberjass) GetScore(idx int) int {
	if idx < 0 || idx >= KlaberjassPlayerCnt {
		return 0
	}
	return k.scores[idx]
}

// GetDealNumber は現在のディール番号を返す。
func (k *Klaberjass) GetDealNumber() int { return k.dealNumber }

// GetGameEndFlag はゲーム終了フラグを返す。
func (k *Klaberjass) GetGameEndFlag() bool { return k.gameEndFlag }

// GetWinnerIdx は勝者を返す (-1 なら未決)。
func (k *Klaberjass) GetWinnerIdx() int { return k.winnerIdx }

// GetConfig は設定を返す。
func (k *Klaberjass) GetConfig() KlaberjassConfig { return k.config }

// SetConfig は設定をセットする。
func (k *Klaberjass) SetConfig(c KlaberjassConfig) { k.config = c }

// GetActionLog は棋譜を返す。
func (k *Klaberjass) GetActionLog() []*ActionLogEntry { return k.actionLog }

// SetPhaseForTest はテスト用にフェーズを設定する。
func (k *Klaberjass) SetPhaseForTest(p KlaberjassPhase) { k.phase = p }

// SetTrumpForTest はテスト用に切札を設定する。
func (k *Klaberjass) SetTrumpForTest(suit int) { k.trumpSuit = suit }

// SetMakerForTest はテスト用にメイカーを設定する。
func (k *Klaberjass) SetMakerForTest(idx int) { k.makerIdx = idx }

// SetCurrentPlayerForTest はテスト用に手番を設定する。
func (k *Klaberjass) SetCurrentPlayerForTest(idx int) { k.currentIdx = idx }

// SetTrickLeaderForTest はテスト用にリード席を設定する。
func (k *Klaberjass) SetTrickLeaderForTest(idx int) { k.trickLeader = idx }

// SetHandForTest はテスト用に手札を差し替える。
func (k *Klaberjass) SetHandForTest(idx int, cards []*Card) {
	p := k.GetPlayer(idx)
	if p == nil {
		return
	}
	for p.GetCardsSize() > 0 {
		p.RemoveCard(0)
	}
	for _, c := range cards {
		p.AddCard(c)
	}
}

// SetHandPointsForTest はテスト用にディール中の得点を設定する。
func (k *Klaberjass) SetHandPointsForTest(idx, pts int) {
	if idx >= 0 && idx < KlaberjassPlayerCnt {
		k.handPoints[idx] = pts
	}
}

// SetTurnUpForTest はテスト用に表向きカードを設定する。
func (k *Klaberjass) SetTurnUpForTest(c *Card) { k.turnUpCard = c }

// SetScoreForTest はテスト用に通算点を設定する。
func (k *Klaberjass) SetScoreForTest(idx, score int) {
	if idx >= 0 && idx < KlaberjassPlayerCnt {
		k.scores[idx] = score
	}
}

// FinishHandForTest はテスト用に精算を走らせる。
func (k *Klaberjass) FinishHandForTest() { k.finishHand() }

// CollectSequencesForTest はテスト用に役の比較を走らせる。
func (k *Klaberjass) CollectSequencesForTest() { k.collectSequences() }

// FindBelaForTest はテスト用にベラ保持者を探す。
func (k *Klaberjass) FindBelaForTest() { k.findBela() }

// addLog は棋譜を 1 行足す。
func (k *Klaberjass) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	k.actionLog = append(k.actionLog, &ActionLogEntry{
		TurnNumber: len(k.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// klaberjassJSON is the JSON wire format for Klaberjass.
type klaberjassJSON struct {
	Players         []*KlaberjassPlayer      `json:"pl"`
	Config          KlaberjassConfig         `json:"cf"`
	Phase           KlaberjassPhase          `json:"ph"`
	DealerIdx       int                      `json:"di"`
	CurrentIdx      int                      `json:"ci"`
	BidIdx          int                      `json:"bi"`
	BidPassCount    int                      `json:"bp"`
	SchmeissBy      int                      `json:"sb"`
	TrumpSuit       int                      `json:"ts"`
	TurnUpCard      *Card                    `json:"tu"`
	MakerIdx        int                      `json:"mi"`
	Trick           []*Card                  `json:"tk"`
	TrickLeader     int                      `json:"tl"`
	TrickNumber     int                      `json:"tn"`
	HandPoints      [KlaberjassPlayerCnt]int `json:"hp"`
	SequenceWinner  int                      `json:"sw"`
	BelaHolder      int                      `json:"bh"`
	BelaKingPlayed  bool                     `json:"bk"`
	BelaQueenPlayed bool                     `json:"bq"`
	BelaScored      bool                     `json:"bs"`
	DixUsed         bool                     `json:"du"`
	BeteFlag        bool                     `json:"bt"`
	LastTrickWinner int                      `json:"lw"`
	Scores          [KlaberjassPlayerCnt]int `json:"sc"`
	DealNumber      int                      `json:"dn"`
	GameEndFlag     bool                     `json:"ge"`
	WinnerIdx       int                      `json:"wi"`
	ActionLog       []*ActionLogEntry        `json:"al"`
	Sequences       [][]*KlaberjassSequence  `json:"sq"`
}

// MarshalJSON implements json.Marshaler.
func (k *Klaberjass) MarshalJSON() ([]byte, error) {
	seqs := make([][]*KlaberjassSequence, KlaberjassPlayerCnt)
	for i := range KlaberjassPlayerCnt {
		seqs[i] = k.sequences[i]
	}
	return json.Marshal(klaberjassJSON{
		Players: k.players, Config: k.config, Phase: k.phase,
		DealerIdx: k.dealerIdx, CurrentIdx: k.currentIdx, BidIdx: k.bidIdx,
		BidPassCount: k.bidPassCount, SchmeissBy: k.schmeissBy,
		TrumpSuit: k.trumpSuit, TurnUpCard: k.turnUpCard, MakerIdx: k.makerIdx,
		Trick: k.trick, TrickLeader: k.trickLeader, TrickNumber: k.trickNumber,
		HandPoints: k.handPoints, SequenceWinner: k.sequenceWinner,
		BelaHolder: k.belaHolder, BelaKingPlayed: k.belaKingPlayed,
		BelaQueenPlayed: k.belaQueenPlayed, BelaScored: k.belaScored,
		DixUsed: k.dixUsed, BeteFlag: k.beteFlag, LastTrickWinner: k.lastTrickWinner,
		Scores: k.scores, DealNumber: k.dealNumber, GameEndFlag: k.gameEndFlag,
		WinnerIdx: k.winnerIdx, ActionLog: k.actionLog, Sequences: seqs,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **復元でしか入らない値を弾く。**KV から戻ってきた壊れた状態でプレイが
// 詰まないよう、席番号とスートを検証する。
func (k *Klaberjass) UnmarshalJSON(data []byte) error {
	var j klaberjassJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != KlaberjassPlayerCnt {
		return fmt.Errorf("bad player count: %d", len(j.Players))
	}
	if j.Phase < KlaberjassPhaseBidTurnUp || j.Phase > KlaberjassPhaseGameEnd {
		return fmt.Errorf("bad phase: %d", j.Phase)
	}
	for name, v := range map[string]int{
		"dealer": j.DealerIdx, "current": j.CurrentIdx, "bid": j.BidIdx,
	} {
		if v < 0 || v >= KlaberjassPlayerCnt {
			return fmt.Errorf("bad %s index: %d", name, v)
		}
	}
	for name, v := range map[string]int{
		"maker": j.MakerIdx, "trick leader": j.TrickLeader,
		"sequence winner": j.SequenceWinner, "bela holder": j.BelaHolder,
		"last trick winner": j.LastTrickWinner, "schmeiss": j.SchmeissBy,
		"winner": j.WinnerIdx,
	} {
		if v < -1 || v >= KlaberjassPlayerCnt {
			return fmt.Errorf("bad %s index: %d", name, v)
		}
	}
	// 0 は「未確定」。それ以外はスートの範囲でなければならない。
	if j.TrumpSuit != 0 && (j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return fmt.Errorf("bad trump suit: %d", j.TrumpSuit)
	}
	if len(j.Trick) > KlaberjassPlayerCnt {
		return fmt.Errorf("bad trick size: %d", len(j.Trick))
	}

	k.players = j.Players
	k.config = j.Config
	k.phase = j.Phase
	k.dealerIdx = j.DealerIdx
	k.currentIdx = j.CurrentIdx
	k.bidIdx = j.BidIdx
	k.bidPassCount = j.BidPassCount
	k.schmeissBy = j.SchmeissBy
	k.trumpSuit = j.TrumpSuit
	k.turnUpCard = j.TurnUpCard
	k.makerIdx = j.MakerIdx
	k.trick = j.Trick
	k.trickLeader = j.TrickLeader
	k.trickNumber = j.TrickNumber
	k.handPoints = j.HandPoints
	k.sequenceWinner = j.SequenceWinner
	k.belaHolder = j.BelaHolder
	k.belaKingPlayed = j.BelaKingPlayed
	k.belaQueenPlayed = j.BelaQueenPlayed
	k.belaScored = j.BelaScored
	k.dixUsed = j.DixUsed
	k.beteFlag = j.BeteFlag
	k.lastTrickWinner = j.LastTrickWinner
	k.scores = j.Scores
	k.dealNumber = j.DealNumber
	k.gameEndFlag = j.GameEndFlag
	k.winnerIdx = j.WinnerIdx
	k.actionLog = j.ActionLog
	for i := range KlaberjassPlayerCnt {
		if i < len(j.Sequences) {
			k.sequences[i] = j.Sequences[i]
		}
	}
	if k.trumpCards == nil {
		k.trumpCards = NewTrumpCardsBelote()
	}
	return nil
}
