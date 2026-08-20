//go:build !js || !wasm || extra3

// Package domain — トリックス (Trix / Trex) のドメインモデル。
//
// レバノン・ヨルダン・エジプトなど中東で遊ばれる契約選択型。52 枚、4 人個人戦。
// **20 ディール** (4 王国 × 5 契約) で、実装は pagat.com に従う。
//
// # issue #4410 の仕様案との相違
//
// issue は**契約 5 種のうち 2 種を取り違えている**。
//
//   - 「ノーハート」→ 実際は **Dinari = ダイヤ**。♦ 1 枚につき -10
//   - 「ノーキングス」(複数形) → 実際は **Sheikh Koobbah = ♥K の 1 枚だけ**で -75
//   - 「ドミノは **7** を起点に数字順へ連結」→ 実際は **J を起点**にする。合法手は
//     「任意の J」または「既に出ている札の 1 つ上か下」
//   - issue は「違反ごとに**失点**を加算」「累積失点が最も少ない人が勝ち」とするが、
//     ドミノは**加点**で、上がり順に +200 / +150 / +100 / +50。したがって勝者は
//     **最終得点が最も高い人**であり、正の点もありうる
//
// # 失点表はゼロサムで検算できる
//
//   - 罰点契約の合計: 75 (♥K) + 130 (♦13枚×10) + 100 (Q4枚×25) + 195 (13トリック×15) = **500**
//   - ドミノの配点: 200 + 150 + 100 + 50 = **500**
//
// **1 王国 (5 ディール) がちょうどゼロサム**になり、20 ディール後の全員の合計も 0。
// どれか 1 つでも値が違えばこの一致は起きないので、これが表の裏取りになる。
//
// # issue の再利用方針も影響を受ける
//
// issue は BarbuDominoes (7 並べ) をそのまま流用できるとするが、Trix のドミノは
// **J 起点**なので起点も伸ばし方も違う。ここでは独自に実装している。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// TrexPlayerCnt はプレイヤー数 (4 人個人戦)。
const TrexPlayerCnt = 4

// TrexHandSize は 1 人あたりの配札枚数 (52 / 4)。
const TrexHandSize = 13

// TrexContractsPerKingdom は 1 人の王が消化する契約数。
const TrexContractsPerKingdom = 5

// TrexTotalDeals は全ディール数 (4 王国 × 5 契約)。
const TrexTotalDeals = TrexPlayerCnt * TrexContractsPerKingdom

// 罰点・加点の定数。合計が 500 対 500 で釣り合うことがこの表の裏取りになる。
const (
	// TrexKingOfHeartsPenalty ♥K を取った人の失点。
	TrexKingOfHeartsPenalty = -75
	// TrexDiamondPenalty ♦ 1 枚あたりの失点。
	TrexDiamondPenalty = -10
	// TrexQueenPenalty Q 1 枚あたりの失点。
	TrexQueenPenalty = -25
	// TrexTrickPenalty 1 トリックあたりの失点。
	TrexTrickPenalty = -15
)

// TrexTrixBonuses は上がり順の加点。合計 500 で罰点契約の合計と釣り合う。
var TrexTrixBonuses = []int{200, 150, 100, 50}

// TrexContract は契約種別。
type TrexContract int

// Trex の契約定数
const (
	// TrexContractKingOfHearts ♥K を取った人が -75
	TrexContractKingOfHearts TrexContract = iota
	// TrexContractDiamonds ♦ 1 枚につき -10
	TrexContractDiamonds
	// TrexContractQueens Q 1 枚につき -25
	TrexContractQueens
	// TrexContractTricks 1 トリックにつき -15
	TrexContractTricks
	// TrexContractTrix J 起点のドミノ。上がり順に加点
	TrexContractTrix
	// TrexContractNone 未選択
	TrexContractNone
)

// TrexPhase はゲームフェーズ。
type TrexPhase int

// Trex のフェーズ定数
const (
	// TrexPhaseChoose 王が契約を選ぶ
	TrexPhaseChoose TrexPhase = iota
	// TrexPhasePlay 進行中
	TrexPhasePlay
	// TrexPhaseDealEnd 1 ディール終了
	TrexPhaseDealEnd
	// TrexPhaseGameEnd 20 ディール終了
	TrexPhaseGameEnd
)

// TrexRank はドミノ用のランク。A が最上位 (J から上へ Q,K,A、下へ 10..2)。
func TrexRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// trexJackRank は J のランク。ドミノの起点。
const trexJackRank = 11

// trexRunSlots はドミノの列を持つ配列の長さ。CardDesignSpade..Diamond が 1..4 な
// ので、添字をそのまま使うには 5 要素必要。
const trexRunSlots = CardDesignDiamond + 1

// newTrexDeck は 52 枚を生成する (シャッフル前)。
func newTrexDeck() []*Card {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	deck := make([]*Card, 0, 52)
	for _, s := range suits {
		for v := 1; v <= 13; v++ {
			deck = append(deck, NewCard(s, v, true))
		}
	}
	return deck
}

// trexShuffle は Fisher-Yates。domain の shuffleCards は casino タグのファイルに
// あり extra3 ビルドから見えないため、専用名で持つ。
func trexShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// TrexTrickCard はトリックに出された 1 枚。
type TrexTrickCard struct {
	PlayerIdx int
	Card      *Card
}

// trexSuitRun は 1 スートのドミノの伸び。low..high が場に出ている範囲。
type trexSuitRun struct {
	Started bool `json:"s"`
	Low     int  `json:"l"`
	High    int  `json:"h"`
}

// Trex はトリックスのゲームクラス。
type Trex struct {
	players []*TrexPlayer
	config  TrexConfig
	phase   TrexPhase

	// kingIdx は現在の王 (契約を選ぶ人)。
	kingIdx int
	// usedContracts[player][contract] はその王が既に選んだか。
	usedContracts [][]bool
	contract      TrexContract
	dealNo        int

	currentIdx int
	leadIdx    int

	// ---- トリック契約 ----
	trick     []TrexTrickCard
	trickNo   int
	tricksWon []int

	// ---- ドミノ契約 ----
	// **スート定数は 1 始まり** (0 はジョーカー) なので、添字をそのまま使うために
	// 5 要素持つ。4 要素にすると ♦ (=4) が範囲外になり、13 枚が永久に出せなくなる。
	runs [trexRunSlots]trexSuitRun
	// finishOrder は上がった順の席番号。
	finishOrder []int

	scores     []int
	dealScores []int

	gameEndFlag bool
	actionLogBase
}

// NewTrex はコンストラクタ。
func NewTrex(players []*TrexPlayer, config TrexConfig) *Trex {
	return &Trex{
		players:  players,
		config:   config,
		contract: TrexContractNone,
		scores:   make([]int, len(players)),
	}
}

// NewDefaultTrex は標準の 4 人セットアップを返す。
func NewDefaultTrex() *Trex {
	players := make([]*TrexPlayer, 0, TrexPlayerCnt)
	players = append(players, NewTrexPlayer(true))
	for range TrexPlayerCnt - 1 {
		players = append(players, NewTrexPlayer(false))
	}
	return NewTrex(players, DefaultTrexConfig())
}

// Reset はゲーム全体を初期化する。
func (t *Trex) Reset() {
	t.scores = make([]int, len(t.players))
	t.usedContracts = make([][]bool, len(t.players))
	for i := range t.usedContracts {
		t.usedContracts[i] = make([]bool, TrexContractTrix+1)
	}
	t.dealNo = 0
	t.gameEndFlag = false
	t.actionLog = nil
	t.dealCards()

	// **♥7 を配られた人が最初の王。**任意の席から始めると、原典の
	// 「王国が右回りに移る」順序の起点が決まらない。
	t.kingIdx = t.holderOf(CardDesignHeart, 7)
	t.addLog(t.kingIdx, "kingdom", "owns the first kingdom (dealt the seven of hearts)", nil)
	t.beginChoose()
}

// holderOf は指定の札を持つ席を返す。
func (t *Trex) holderOf(design, value int) int {
	for i, p := range t.players {
		for j := range p.GetCardsSize() {
			c := p.GetCard(j)
			if c != nil && c.GetDesign() == design && c.GetValue() == value {
				return i
			}
		}
	}
	return 0
}

// dealCards は 13 枚ずつ配る。
func (t *Trex) dealCards() {
	for _, p := range t.players {
		p.ResetGame()
	}
	deck := newTrexDeck()
	trexShuffle(deck)
	for i, c := range deck {
		t.players[i%len(t.players)].AddCard(c)
	}
}

// beginChoose は契約選択フェーズに入る。
func (t *Trex) beginChoose() {
	t.phase = TrexPhaseChoose
	t.contract = TrexContractNone
	t.currentIdx = t.kingIdx
	t.trick = nil
	t.trickNo = 0
	t.tricksWon = make([]int, len(t.players))
	t.runs = [trexRunSlots]trexSuitRun{}
	t.finishOrder = nil
	t.dealScores = make([]int, len(t.players))
}

// AvailableContracts は王がまだ選んでいない契約を返す。
func (t *Trex) AvailableContracts() []TrexContract {
	if t.kingIdx < 0 || t.kingIdx >= len(t.usedContracts) {
		return nil
	}
	var out []TrexContract
	for c := TrexContractKingOfHearts; c <= TrexContractTrix; c++ {
		if !t.usedContracts[t.kingIdx][c] {
			out = append(out, c)
		}
	}
	return out
}

// ChooseContract は王が契約を選ぶ。
func (t *Trex) ChooseContract(player int, contract TrexContract) error {
	if t.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if t.phase != TrexPhaseChoose {
		return fmt.Errorf("it is not the contract-choosing phase")
	}
	if player != t.kingIdx {
		return fmt.Errorf("only the king (player %d) chooses the contract", t.kingIdx)
	}
	if contract < TrexContractKingOfHearts || contract > TrexContractTrix {
		return fmt.Errorf("unknown contract: %d", contract)
	}
	if t.usedContracts[player][contract] {
		return fmt.Errorf("that contract has already been played in this kingdom")
	}

	t.usedContracts[player][contract] = true
	t.contract = contract
	t.phase = TrexPhasePlay
	// リードは王から。ドミノでも王が最初に出す (出せなければパス)。
	t.currentIdx = t.kingIdx
	t.leadIdx = t.kingIdx
	t.addLog(player, "contract", fmt.Sprintf("chooses contract %d", contract), nil)
	if t.contract == TrexContractTrix {
		t.skipStuckTrixPlayers()
	}
	return nil
}

// IsTrix は現在の契約がドミノかを返す。
func (t *Trex) IsTrix() bool { return t.contract == TrexContractTrix }

// GetValidPlayIndices は player が出せる手札の添字を返す。
func (t *Trex) GetValidPlayIndices(player int) []int {
	p := t.GetPlayer(player)
	if p == nil || t.phase != TrexPhasePlay {
		return nil
	}
	if t.contract == TrexContractTrix {
		var out []int
		for i := range p.GetCardsSize() {
			if t.trixPlayable(p.GetCard(i)) {
				out = append(out, i)
			}
		}
		return out
	}

	all := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if p.GetCard(i) != nil {
			all = append(all, i)
		}
	}
	if len(t.trick) == 0 {
		return all
	}
	lead := t.trick[0].Card
	var follow []int
	for _, i := range all {
		if p.GetCard(i).GetDesign() == lead.GetDesign() {
			follow = append(follow, i)
		}
	}
	if len(follow) > 0 {
		return follow
	}
	return all
}

// trixPlayable は c がドミノで出せるかを返す。
//
// **合法手は「任意の J」または「既に出ている札の 1 つ上か下」。**7 起点ではなく
// J 起点で、スートごとに J から上下へ伸びる。
func (t *Trex) trixPlayable(c *Card) bool {
	if c == nil {
		return false
	}
	suit := c.GetDesign()
	if suit < 0 || suit >= len(t.runs) {
		return false
	}
	rank := TrexRank(c)
	run := t.runs[suit]
	if !run.Started {
		return rank == trexJackRank
	}
	return rank == run.Low-1 || rank == run.High+1
}

// PlayCard は player が手札 handIdx の札を出す。
func (t *Trex) PlayCard(player, handIdx int) error {
	if t.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if t.phase != TrexPhasePlay {
		return fmt.Errorf("it is not the play phase")
	}
	if player != t.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	p := t.GetPlayer(player)
	if p == nil {
		return fmt.Errorf("no such player: %d", player)
	}
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return fmt.Errorf("card index %d out of range", handIdx)
	}
	legal := false
	for _, i := range t.GetValidPlayIndices(player) {
		if i == handIdx {
			legal = true
		}
	}
	if !legal {
		return fmt.Errorf("card index %d is not a legal play", handIdx)
	}

	card := p.RemoveCard(handIdx)
	if t.contract == TrexContractTrix {
		t.resolveTrixPlay(player, card)
		return nil
	}
	t.trick = append(t.trick, TrexTrickCard{PlayerIdx: player, Card: card})
	t.addLog(player, "play", "plays a card", []*Card{card})
	if len(t.trick) < len(t.players) {
		t.currentIdx = (t.currentIdx + 1) % len(t.players)
		return nil
	}
	t.resolveTrick()
	return nil
}

// Pass はドミノで出せないときに手番を渡す。
func (t *Trex) Pass(player int) error {
	if t.phase != TrexPhasePlay || t.contract != TrexContractTrix {
		return fmt.Errorf("passing is only possible in the dominoes contract")
	}
	if player != t.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	if len(t.GetValidPlayIndices(player)) > 0 {
		return fmt.Errorf("you have a legal play, so you may not pass")
	}
	t.addLog(player, "pass", "cannot play and passes", nil)
	t.advanceTrix()
	return nil
}

// resolveTrixPlay はドミノの 1 枚を処理する。
func (t *Trex) resolveTrixPlay(player int, card *Card) {
	suit := card.GetDesign()
	rank := TrexRank(card)
	if !t.runs[suit].Started {
		t.runs[suit] = trexSuitRun{Started: true, Low: rank, High: rank}
	} else if rank < t.runs[suit].Low {
		t.runs[suit].Low = rank
	} else {
		t.runs[suit].High = rank
	}
	t.addLog(player, "play", "extends the layout", []*Card{card})

	if t.GetPlayer(player).GetCardsSize() == 0 {
		t.finishOrder = append(t.finishOrder, player)
		t.GetPlayer(player).SetIsFinished(true)
		t.addLog(player, "out", fmt.Sprintf("goes out in position %d", len(t.finishOrder)), nil)
		if len(t.finishOrder) >= len(t.players)-1 {
			t.finishTrix()
			return
		}
	}
	t.advanceTrix()
}

// advanceTrix は次のまだ上がっていない人へ手番を送り、出せない人を飛ばす。
func (t *Trex) advanceTrix() {
	t.currentIdx = t.nextUnfinished(t.currentIdx)
	t.skipStuckTrixPlayers()
}

// skipStuckTrixPlayers は出せない人を飛ばす。全員出せなければ進行が止まるが、
// J が場に出る限り必ず誰かは出せるので、安全弁として回数を切る。
func (t *Trex) skipStuckTrixPlayers() {
	for range len(t.players) * 2 {
		if t.phase != TrexPhasePlay {
			return
		}
		if len(t.GetValidPlayIndices(t.currentIdx)) > 0 {
			return
		}
		t.currentIdx = t.nextUnfinished(t.currentIdx)
	}
}

// nextUnfinished は idx の次の、まだ上がっていない人を返す。
func (t *Trex) nextUnfinished(idx int) int {
	for i := 1; i <= len(t.players); i++ {
		n := (idx + i) % len(t.players)
		if t.GetPlayer(n).GetCardsSize() > 0 {
			return n
		}
	}
	return idx
}

// finishTrix はドミノを精算する。上がり順に +200/+150/+100/+50。
func (t *Trex) finishTrix() {
	// 最後の 1 人は自動的に最下位。
	for i := range t.players {
		if t.GetPlayer(i).GetCardsSize() > 0 {
			t.finishOrder = append(t.finishOrder, i)
		}
	}
	for pos, idx := range t.finishOrder {
		if pos >= len(TrexTrixBonuses) {
			break
		}
		t.dealScores[idx] += TrexTrixBonuses[pos]
	}
	t.settleDeal()
}

// TrexTrickWinner はトリックの勝者の席を返す (切札なし、リードのスートの最強)。
func TrexTrickWinner(trick []TrexTrickCard) int {
	if len(trick) == 0 {
		return -1
	}
	best := 0
	lead := trick[0].Card.GetDesign()
	for i := 1; i < len(trick); i++ {
		c := trick[i].Card
		if c.GetDesign() != lead {
			continue
		}
		if TrexRank(c) > TrexRank(trick[best].Card) {
			best = i
		}
	}
	return trick[best].PlayerIdx
}

// resolveTrick はトリックを解決し、契約に応じて失点を付ける。
func (t *Trex) resolveTrick() {
	winner := TrexTrickWinner(t.trick)
	cards := make([]*Card, 0, len(t.trick))
	penalty := 0
	for _, tc := range t.trick {
		cards = append(cards, tc.Card)
		penalty += t.cardPenalty(tc.Card)
	}
	if t.contract == TrexContractTricks {
		penalty += TrexTrickPenalty
	}
	t.dealScores[winner] += penalty
	t.tricksWon[winner]++
	t.addLog(winner, "trick", fmt.Sprintf("wins the trick for %d", penalty), cards)

	t.trick = nil
	t.trickNo++
	t.leadIdx = winner
	t.currentIdx = winner

	if t.trickNo >= TrexHandSize {
		t.settleDeal()
		return
	}
	// ♥K 契約は ♥K が出た時点で終わる。残りを消化しても点は動かない。
	if t.contract == TrexContractKingOfHearts && t.kingOfHeartsGone() {
		t.settleDeal()
	}
}

// kingOfHeartsGone は ♥K が既に取られたかを返す。
func (t *Trex) kingOfHeartsGone() bool {
	for _, p := range t.players {
		for j := range p.GetCardsSize() {
			c := p.GetCard(j)
			if c != nil && c.GetDesign() == CardDesignHeart && c.GetValue() == 13 {
				return false
			}
		}
	}
	return true
}

// cardPenalty は契約に応じた 1 枚の失点を返す。
func (t *Trex) cardPenalty(c *Card) int {
	return TrexCardPenalty(t.contract, c)
}

// TrexCardPenalty は契約 contract のもとで 1 枚が背負う失点を返す。
//
// **表示側が switch を書き写さないため。**5 つの契約が 1 王国内で入れ替わるので、
// どの札が危険かは配りごとに変わる (#4911)。得点を決めているのと同じ関数を
// 画面が呼べば、印と実際の失点が食い違いようがない (#5572)。
func TrexCardPenalty(contract TrexContract, c *Card) int {
	if c == nil {
		return 0
	}
	switch contract {
	case TrexContractKingOfHearts:
		if c.GetDesign() == CardDesignHeart && c.GetValue() == 13 {
			return TrexKingOfHeartsPenalty
		}
	case TrexContractDiamonds:
		if c.GetDesign() == CardDesignDiamond {
			return TrexDiamondPenalty
		}
	case TrexContractQueens:
		if c.GetValue() == 12 {
			return TrexQueenPenalty
		}
	case TrexContractTricks, TrexContractTrix, TrexContractNone:
		return 0
	}
	return 0
}

// settleDeal は 1 ディールを締める。
func (t *Trex) settleDeal() {
	for i := range t.players {
		t.scores[i] += t.dealScores[i]
	}
	t.dealNo++
	t.phase = TrexPhaseDealEnd
	t.addLog(-1, "deal_end", fmt.Sprintf("deal %d of %d settled", t.dealNo, TrexTotalDeals), nil)

	if t.dealNo >= TrexTotalDeals {
		t.phase = TrexPhaseGameEnd
		t.gameEndFlag = true
		t.addLog(-1, "game_end", "all twenty deals played", nil)
	}
}

// NextDeal は次のディールを配る。ディール終了時のみ。
func (t *Trex) NextDeal() error {
	if t.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if t.phase != TrexPhaseDealEnd {
		return fmt.Errorf("the deal is still in progress")
	}
	// 王が 5 契約を消化したら王国が移る。
	if t.dealNo%TrexContractsPerKingdom == 0 {
		t.kingIdx = (t.kingIdx + 1) % len(t.players)
		t.addLog(t.kingIdx, "kingdom", "takes over the kingdom", nil)
	}
	t.dealCards()
	t.beginChoose()
	return nil
}

// ---- CPU ----

// TrexCpuAction は CPU が選んだ手。
type TrexCpuAction struct {
	// Contract は選択フェーズで選ぶ契約 (それ以外では TrexContractNone)。
	Contract TrexContract
	// HandIdx は出す札の手札添字 (パス/手なしのときは -1)。
	HandIdx int
	// Pass が真ならドミノでパスする。
	Pass bool
}

// TrexCpuDecide は idx の CPU が取る手を決める。
//
// 契約は残りから最初のものを選ぶ。プレイは「トリック契約では失点を避け、
// ドミノでは出せる中で最も端の札を先に出す」。
func (t *Trex) TrexCpuDecide(idx int) TrexCpuAction {
	if t.phase == TrexPhaseChoose {
		avail := t.AvailableContracts()
		if len(avail) == 0 {
			return TrexCpuAction{Contract: TrexContractNone, HandIdx: -1}
		}
		return TrexCpuAction{Contract: avail[0], HandIdx: -1}
	}

	valid := t.GetValidPlayIndices(idx)
	if len(valid) == 0 {
		if t.contract == TrexContractTrix {
			return TrexCpuAction{Contract: TrexContractNone, HandIdx: -1, Pass: true}
		}
		return TrexCpuAction{Contract: TrexContractNone, HandIdx: -1}
	}
	if t.contract == TrexContractTrix {
		// 出せるものは出す。詰まって全員パスになるより手札を減らす方が良い。
		return TrexCpuAction{Contract: TrexContractNone, HandIdx: valid[0]}
	}

	p := t.GetPlayer(idx)
	// 罰のある札を持たされたくないので、リードでは最も弱い札、追随では
	// 「勝たない最強」を選ぶ。勝たざるを得ないなら最弱で取る。
	if len(t.trick) == 0 {
		return TrexCpuAction{Contract: TrexContractNone, HandIdx: t.weakestOf(idx, valid)}
	}
	leadSuit := t.trick[0].Card.GetDesign()
	bestRank := 0
	for _, tc := range t.trick {
		if tc.Card.GetDesign() == leadSuit && TrexRank(tc.Card) > bestRank {
			bestRank = TrexRank(tc.Card)
		}
	}
	duck, duckRank := -1, 0
	for _, i := range valid {
		c := p.GetCard(i)
		if c.GetDesign() != leadSuit || TrexRank(c) >= bestRank {
			continue
		}
		if duck == -1 || TrexRank(c) > duckRank {
			duck, duckRank = i, TrexRank(c)
		}
	}
	if duck >= 0 {
		return TrexCpuAction{Contract: TrexContractNone, HandIdx: duck}
	}
	return TrexCpuAction{Contract: TrexContractNone, HandIdx: t.weakestOf(idx, valid)}
}

// weakestOf は候補のうち最も弱い札の添字を返す。
func (t *Trex) weakestOf(idx int, candidates []int) int {
	p := t.GetPlayer(idx)
	best, bestRank := candidates[0], 99
	for _, i := range candidates {
		if r := TrexRank(p.GetCard(i)); r < bestRank {
			best, bestRank = i, r
		}
	}
	return best
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (t *Trex) GetPlayers() []*TrexPlayer { return t.players }

// GetPlayer は idx のプレイヤーを返す。
func (t *Trex) GetPlayer(idx int) *TrexPlayer {
	return getPlayer(t.players, idx)
}

// GetPhase は現在のフェーズを返す。
func (t *Trex) GetPhase() TrexPhase { return t.phase }

// GetCurrentPlayerIdx は手番のプレイヤー添字を返す。
func (t *Trex) GetCurrentPlayerIdx() int { return t.currentIdx }

// GetKingIdx は王の添字を返す。
func (t *Trex) GetKingIdx() int { return t.kingIdx }

// GetContract は現在の契約を返す。
func (t *Trex) GetContract() TrexContract { return t.contract }

// GetDealNumber は完了したディール数を返す。
func (t *Trex) GetDealNumber() int { return t.dealNo }

// GetTrick は現在のトリックを返す。
func (t *Trex) GetTrick() []TrexTrickCard { return t.trick }

// GetTrickNumber は完了したトリック数を返す。
func (t *Trex) GetTrickNumber() int { return t.trickNo }

// GetTricksWon は idx のトリック数を返す。
func (t *Trex) GetTricksWon(idx int) int {
	if idx < 0 || idx >= len(t.tricksWon) {
		return 0
	}
	return t.tricksWon[idx]
}

// GetScore は idx の累計得点を返す。
func (t *Trex) GetScore(idx int) int {
	return elemAt(t.scores, idx)
}

// GetDealScore は idx の今ディールの得点を返す。
func (t *Trex) GetDealScore(idx int) int {
	if idx < 0 || idx >= len(t.dealScores) {
		return 0
	}
	return t.dealScores[idx]
}

// GetFinishOrder はドミノの上がり順を返す。
func (t *Trex) GetFinishOrder() []int { return t.finishOrder }

// GetSuitRun は suit のドミノの伸び (開始済み, 下限, 上限) を返す。
func (t *Trex) GetSuitRun(suit int) (bool, int, int) {
	if suit < 0 || suit >= len(t.runs) {
		return false, 0, 0
	}
	r := t.runs[suit]
	return r.Started, r.Low, r.High
}

// IsContractUsed は king がその契約を消化済みかを返す。
func (t *Trex) IsContractUsed(king int, contract TrexContract) bool {
	if king < 0 || king >= len(t.usedContracts) {
		return false
	}
	if contract < 0 || int(contract) >= len(t.usedContracts[king]) {
		return false
	}
	return t.usedContracts[king][contract]
}

// GetGameEndFlag は 20 ディールを終えたかを返す。
func (t *Trex) GetGameEndFlag() bool { return t.gameEndFlag }

// GetWinnerIdx は最終得点が最も高い席を返す (未終局なら -1)。
func (t *Trex) GetWinnerIdx() int {
	if !t.gameEndFlag {
		return -1
	}
	best := 0
	for i := 1; i < len(t.scores); i++ {
		if t.scores[i] > t.scores[best] {
			best = i
		}
	}
	return best
}

// GetConfig はゲーム設定を返す。
func (t *Trex) GetConfig() TrexConfig { return t.config }

// SetConfig はゲーム設定をセットする。
func (t *Trex) SetConfig(c TrexConfig) { t.config = c }

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (t *Trex) SetPhaseForTest(p TrexPhase) { t.phase = p }

// SetContractForTest はテスト用に契約を差し替える。
func (t *Trex) SetContractForTest(c TrexContract) { t.contract = c }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (t *Trex) SetCurrentPlayerForTest(idx int) { t.currentIdx = idx }

// SetKingForTest はテスト用に王を差し替える。
func (t *Trex) SetKingForTest(idx int) { t.kingIdx = idx }

// SetDealNumberForTest はテスト用にディール数を差し替える。
func (t *Trex) SetDealNumberForTest(n int) { t.dealNo = n }

// addLog は棋譜に 1 件追加する。
func (t *Trex) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	t.appendLog(playerIdx, actionType, detail, cards)
}

// trexTrickCardJSON is the JSON wire format for TrexTrickCard.
type trexTrickCardJSON struct {
	PlayerIdx int   `json:"p"`
	Card      *Card `json:"c"`
}

// trexJSON is the JSON wire format for Trex.
type trexJSON struct {
	Players       []*TrexPlayer             `json:"pl"`
	Config        TrexConfig                `json:"cfg"`
	Phase         TrexPhase                 `json:"ph"`
	KingIdx       int                       `json:"ki"`
	UsedContracts [][]bool                  `json:"uc"`
	Contract      TrexContract              `json:"ct"`
	DealNo        int                       `json:"dn"`
	Current       int                       `json:"cur"`
	LeadIdx       int                       `json:"li"`
	Trick         []trexTrickCardJSON       `json:"tk"`
	TrickNo       int                       `json:"tn"`
	TricksWon     []int                     `json:"tw"`
	Runs          [trexRunSlots]trexSuitRun `json:"rn"`
	FinishOrder   []int                     `json:"fo"`
	Scores        []int                     `json:"sc"`
	DealScores    []int                     `json:"ds"`
	GameEnd       bool                      `json:"ge"`
	ActionLog     []*ActionLogEntry         `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (t *Trex) MarshalJSON() ([]byte, error) {
	trick := make([]trexTrickCardJSON, 0, len(t.trick))
	for _, tc := range t.trick {
		trick = append(trick, trexTrickCardJSON(tc))
	}
	return json.Marshal(trexJSON{
		Players: t.players, Config: t.config, Phase: t.phase, KingIdx: t.kingIdx,
		UsedContracts: t.usedContracts, Contract: t.contract, DealNo: t.dealNo,
		Current: t.currentIdx, LeadIdx: t.leadIdx, Trick: trick, TrickNo: t.trickNo,
		TricksWon: t.tricksWon, Runs: t.runs, FinishOrder: t.finishOrder,
		Scores: t.scores, DealScores: t.dealScores, GameEnd: t.gameEndFlag,
		ActionLog: t.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// KV から戻る生バイト列は信用できないので、席数に合わせて詰め直し、設定を検証する。
// **usedContracts が欠けると同じ契約を二度選べてしまう**ので、長さも固定する。
func (t *Trex) UnmarshalJSON(data []byte) error {
	var raw trexJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Players) != TrexPlayerCnt {
		return fmt.Errorf("expected %d players, got %d", TrexPlayerCnt, len(raw.Players))
	}
	if err := raw.Config.Validate(); err != nil {
		return err
	}
	if raw.Phase < TrexPhaseChoose || raw.Phase > TrexPhaseGameEnd {
		return fmt.Errorf("unknown phase: %d", raw.Phase)
	}
	if raw.Contract < TrexContractKingOfHearts || raw.Contract > TrexContractNone {
		return fmt.Errorf("unknown contract: %d", raw.Contract)
	}

	t.players = raw.Players
	t.config = raw.Config
	t.phase = raw.Phase
	t.contract = raw.Contract
	t.dealNo = raw.DealNo
	t.trickNo = raw.TrickNo
	t.runs = raw.Runs
	t.gameEndFlag = raw.GameEnd
	t.actionLog = raw.ActionLog

	t.kingIdx = clampTrexIdx(raw.KingIdx, len(t.players))
	t.currentIdx = clampTrexIdx(raw.Current, len(t.players))
	t.leadIdx = clampTrexIdx(raw.LeadIdx, len(t.players))

	t.usedContracts = make([][]bool, len(t.players))
	for i := range t.usedContracts {
		t.usedContracts[i] = make([]bool, TrexContractTrix+1)
		if i < len(raw.UsedContracts) {
			copy(t.usedContracts[i], raw.UsedContracts[i])
		}
	}
	t.tricksWon = padTrexInts(raw.TricksWon, len(t.players))
	t.scores = padTrexInts(raw.Scores, len(t.players))
	t.dealScores = padTrexInts(raw.DealScores, len(t.players))

	t.finishOrder = make([]int, 0, len(raw.FinishOrder))
	for _, idx := range raw.FinishOrder {
		if idx >= 0 && idx < len(t.players) {
			t.finishOrder = append(t.finishOrder, idx)
		}
	}

	t.trick = make([]TrexTrickCard, 0, len(raw.Trick))
	for _, tc := range raw.Trick {
		if tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= len(t.players) {
			continue
		}
		t.trick = append(t.trick, TrexTrickCard(tc))
	}
	return nil
}

// clampTrexIdx は席番号を 0..n-1 に収める。
func clampTrexIdx(idx, n int) int {
	if idx < 0 || idx >= n {
		return 0
	}
	return idx
}

// padTrexInts は長さ n に詰め直す。
func padTrexInts(src []int, n int) []int {
	out := make([]int, n)
	copy(out, src)
	return out
}
