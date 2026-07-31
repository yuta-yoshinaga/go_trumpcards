//go:build !js || !wasm || extra3

// Package domain — ポッホ (Poch) のドメインモデル。
//
// 15 世紀ドイツの 3 段階構成ゲーム。32 枚、専用盤の 9 プール。
//
// # issue #4415 の仕様案との相違
//
// 3 段階という骨格は合っているが、**各段階の中身が 3 つとも違う**。
//
//   - issue は「余りは中央ボードの区画に**配置**してチップを積む」とするが、
//     **区画に札は置かない**。毎ディール全員が **9 プール全部に 1 枚ずつ**
//     チップを置き、余り札は 1 枚だけ表向きにして、そのスートが **pay suit**
//     になる
//   - issue は「手札に該当**ランク**があれば対応区画のチップを獲得」とするが、
//     ランクだけでは取れない。**pay suit の** A/K/Q/J/10 を持つ人が取る。
//     Marriage は pay suit の K と Q の両方、Sequence は pay suit の 7-8-9
//   - issue は pochen を「『同ランクを○枚持つ』と**宣言**し、宣言を通した人が
//     ポットを取る」ブラフゲームとするが、**宣言もブラフもない**。同ランクの
//     組 (ペア/スリー/フォー) の**強さ比べ**で、4 > 3 > 2。降りずに残った中の
//     最強が賭けチップ全部と Pocher プールを取る
//   - issue は「出せなくなったら手番終了」とするが、ストップスは**同じスートの
//     次に高い札**を順に出す。誰も次を持っていなければ stop で、**最後に最高札
//     を出した人**が好きな札から再開する
//   - issue が触れていない: 出し切った人は **centre pot に加えて、他家から
//     残り札 1 枚につき 1 チップ**を受け取る。そして**取られなかったプールは
//     持ち越される**
//
// したがって issue の「CPU の poch 宣言はブラフ確率と実際の手札枚数から閾値
// 判定する」は不要になる。組の強さは確定情報なので、CPU の判断は「自分の組で
// どこまで賭けるか」だけである。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
)

// PochPlayerCnt はプレイヤー数 (4 人。原典は 3〜6 人)。
const PochPlayerCnt = 4

// PochDeckSize は 32 枚デッキ (7〜A)。
const PochDeckSize = 32

// PochMinRank は 32 枚デッキの最小ランク。7 未満は使わない。
const PochMinRank = 7

// PochBetUnit は 1 回のベット額。
const PochBetUnit = 1

// PochPhase はゲームフェーズ。
type PochPhase int

// Poch のフェーズ定数
const (
	// PochPhaseStaking 第 1 段階: pay suit のプール回収 (自動で解決する)
	PochPhaseStaking PochPhase = iota
	// PochPhasePochen 第 2 段階: 同ランクの組を賭ける
	PochPhasePochen
	// PochPhaseStops 第 3 段階: 同スート昇順で出し切る
	PochPhaseStops
	// PochPhaseDealEnd 1 ディール終了
	PochPhaseDealEnd
	// PochPhaseGameEnd 決着
	PochPhaseGameEnd
)

// PochCombo は同ランクの組。
type PochCombo struct {
	// Size は枚数 (2/3/4)。0 なら組なし。
	Size int
	// Rank は組のランク。
	Rank int
}

// Beats は c が other より強いかを返す。
//
// **枚数が先で、同数ならランク。**4 > 3 > 2 は原典が明記している
// («any set of four of a kind beats any set of three of a kind»)。
func (c PochCombo) Beats(other PochCombo) bool {
	if c.Size != other.Size {
		return c.Size > other.Size
	}
	return c.Rank > other.Rank
}

// PochBestCombo は手札から最も強い同ランクの組を返す。
//
// **ペア未満は組にならない。**単札しかなければ Size 0 を返す。
func PochBestCombo(cards []*Card) PochCombo {
	counts := map[int]int{}
	for _, c := range cards {
		if c != nil {
			counts[c.GetValue()]++
		}
	}
	// **比較は必ず順位に直してから。**生ランクのままだと A(=1) が最弱に
	// なってしまう。返り値は生ランクなので、比較用と返却用を分けて持つ。
	best, bestOrdered := PochCombo{}, PochCombo{}
	for rank, n := range counts {
		if n < 2 {
			continue
		}
		ordered := PochCombo{Size: n, Rank: pochRankOrder(rank)}
		if ordered.Beats(bestOrdered) {
			best, bestOrdered = PochCombo{Size: n, Rank: rank}, ordered
		}
	}
	return best
}

// pochRankOrder は A を最上位に置いた順位を返す。A-K-Q-J-10-9-8-7。
func pochRankOrder(rank int) int {
	if rank == 1 {
		return 14
	}
	return rank
}

// newPochDeck は 32 枚 (各スート 7〜K と A) を生成する。
func newPochDeck() []*Card {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	deck := make([]*Card, 0, PochDeckSize)
	for _, s := range suits {
		deck = append(deck, NewCard(s, 1, true))
		for v := PochMinRank; v <= 13; v++ {
			deck = append(deck, NewCard(s, v, true))
		}
	}
	return deck
}

// PochStakingAward は第 1 段階で誰がどのプールを取ったかの記録。
type PochStakingAward struct {
	Pool   PochPool
	Player int
	Chips  int
}

// Poch はポッホのゲームクラス。
type Poch struct {
	players []*PochPlayer
	config  PochConfig
	phase   PochPhase
	board   PochBoard

	// paySuit は表向きにした余り札のスート。この 1 枚が第 1 段階を決める。
	paySuit int
	// turnUp は表向きにした余り札そのもの。
	turnUp *Card
	// stakingAwards は直近ディールの第 1 段階の結果。
	stakingAwards []*PochStakingAward

	currentIdx int
	dealerIdx  int
	dealNo     int

	// betTarget は現在の賭け額。全員がここに揃うまでラウンドが続く。
	betTarget int
	// pochenWinner は pochen を取った席 (-1: 未決着)。
	pochenWinner int
	// pochenPot は pochen で集まったチップ (Pocher プールを含む)。
	pochenPot int

	// stopsSuit / stopsRank は今の並びの続き。stopsSuit < 0 なら好きな札で開始。
	stopsSuit int
	stopsRank int
	// playedPile はストップスで出た札。
	playedPile []*Card

	dealWinner  int
	gameEndFlag bool
	winnerIdx   int
	actionLog   []*ActionLogEntry
}

// NewPoch はコンストラクタ。
func NewPoch(players []*PochPlayer, config PochConfig) *Poch {
	return &Poch{
		players:      players,
		config:       config,
		pochenWinner: -1,
		dealWinner:   -1,
		winnerIdx:    -1,
		stopsSuit:    -1,
	}
}

// NewDefaultPoch は標準の 4 人セットアップを返す。
func NewDefaultPoch() *Poch {
	players := make([]*PochPlayer, 0, PochPlayerCnt)
	players = append(players, NewPochPlayer(true))
	for range PochPlayerCnt - 1 {
		players = append(players, NewPochPlayer(false))
	}
	return NewPoch(players, DefaultPochConfig())
}

// Reset はゲーム全体を初期化する。
func (p *Poch) Reset() {
	p.board = PochBoard{}
	p.dealerIdx = 0
	p.dealNo = 0
	p.gameEndFlag = false
	p.winnerIdx = -1
	p.actionLog = nil
	for _, pl := range p.players {
		pl.AddChips(-pl.GetChips())
	}
	p.dealRound()
}

// dealRound は 1 ディールを配り、第 1 段階を解決する。
func (p *Poch) dealRound() {
	for _, pl := range p.players {
		pl.ResetDeal()
	}
	p.stakingAwards = nil
	p.pochenWinner = -1
	p.pochenPot = 0
	p.betTarget = 0
	p.dealWinner = -1
	p.stopsSuit = -1
	p.stopsRank = 0
	p.playedPile = nil

	// **プールは持ち越したまま、全員がさらに 1 枚ずつ足す。**
	p.board.Ante(len(p.players))
	for _, pl := range p.players {
		pl.AddChips(-int(PochPoolCount))
	}

	deck := newPochDeck()
	pochShuffle(deck)
	// **1 枚だけ残るまで 1 枚ずつ配る。**残った 1 枚を表にして pay suit を決める。
	pos := 0
	for i := 0; pos < len(deck)-1; i++ {
		p.players[i%len(p.players)].AddCard(deck[pos])
		pos++
	}
	p.turnUp = deck[len(deck)-1]
	p.paySuit = p.turnUp.GetDesign()

	p.phase = PochPhaseStaking
	p.addLog(-1, "deal", fmt.Sprintf("pay suit is %d", p.paySuit), []*Card{p.turnUp})
	p.resolveStaking()
}

// pochShuffle は Fisher-Yates。domain の shuffleCards は別バケットのファイルに
// あり extra3 ビルドから見えないため、専用名で持つ。
func pochShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// resolveStaking は第 1 段階を解決する。手札は確定情報なので選択の余地がなく、
// 配った直後に自動で片づく。
//
// **ランクだけでは取れない。**pay suit の札でなければならない。
func (p *Poch) resolveStaking() {
	for _, rp := range pochRankPools {
		if seat := p.holderOf(rp.rank); seat >= 0 {
			p.award(rp.pool, seat)
		}
	}
	// Marriage は pay suit の K と Q を**同じ人**が持っているとき。
	if seat := p.holderOf(13); seat >= 0 && seat == p.holderOf(12) {
		p.award(PochPoolMarriage, seat)
	}
	// Sequence は pay suit の 7-8-9 を**同じ人**が三枚とも持っているとき。
	if seat := p.holderOf(7); seat >= 0 && seat == p.holderOf(8) && seat == p.holderOf(9) {
		p.award(PochPoolSequence, seat)
	}

	p.phase = PochPhasePochen
	p.currentIdx = (p.dealerIdx + 1) % len(p.players)
	p.betTarget = 0
}

// award は pool のチップを seat に渡し、記録を残す。
func (p *Poch) award(pool PochPool, seat int) {
	n := p.board.Take(pool)
	if n == 0 {
		return
	}
	p.players[seat].AddChips(n)
	p.stakingAwards = append(p.stakingAwards, &PochStakingAward{Pool: pool, Player: seat, Chips: n})
	p.addLog(seat, "staking", fmt.Sprintf("takes the %s pool (%d)", pool, n), nil)
}

// holderOf は pay suit の rank を持つ席を返す (-1: 誰も持っていない)。
func (p *Poch) holderOf(rank int) int {
	for i, pl := range p.players {
		for j := range pl.GetCardsSize() {
			c := pl.GetCard(j)
			if c != nil && c.GetDesign() == p.paySuit && c.GetValue() == rank {
				return i
			}
		}
	}
	return -1
}

// ---- 第 2 段階: pochen ----

// Bet は 1 単位賭ける (コールまたはレイズ)。
func (p *Poch) Bet(player int) error {
	if err := p.checkPochen(player); err != nil {
		return err
	}
	pl := p.GetPlayer(player)
	// 追いつくのに要る額 + 1 単位。追いつくだけなら差額のみ。
	need := p.betTarget - pl.GetBet()
	if need <= 0 {
		need = PochBetUnit
	}
	pl.PlaceBet(need)
	if pl.GetBet() > p.betTarget {
		p.betTarget = pl.GetBet()
	}
	p.addLog(player, "bet", fmt.Sprintf("bets %d", need), nil)
	p.advancePochen()
	return nil
}

// Fold は降りる。
func (p *Poch) Fold(player int) error {
	if err := p.checkPochen(player); err != nil {
		return err
	}
	p.GetPlayer(player).Fold()
	p.addLog(player, "fold", "folds", nil)
	p.advancePochen()
	return nil
}

// checkPochen は賭けられる状態かを確かめる。
func (p *Poch) checkPochen(player int) error {
	if p.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if p.phase != PochPhasePochen {
		return fmt.Errorf("it is not the betting stage")
	}
	if player != p.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	if p.GetPlayer(player).IsFolded() {
		return fmt.Errorf("player %d has folded", player)
	}
	return nil
}

// advancePochen は次の生存者へ手番を回し、決着していれば精算する。
func (p *Poch) advancePochen() {
	if p.settlePochenIfDone() {
		return
	}
	for i := 1; i <= len(p.players); i++ {
		n := (p.currentIdx + i) % len(p.players)
		if !p.players[n].IsFolded() {
			p.currentIdx = n
			return
		}
	}
}

// settlePochenIfDone は「1 人だけ残った」か「全員の賭け額が揃った」ときに
// pochen を締める。
func (p *Poch) settlePochenIfDone() bool {
	alive := make([]int, 0, len(p.players))
	for i, pl := range p.players {
		if !pl.IsFolded() {
			alive = append(alive, i)
		}
	}
	if len(alive) == 1 {
		p.finishPochen(alive[0])
		return true
	}
	// 誰もまだ賭けていなければ揃っていないのと同じ。1 周させる。
	if p.betTarget == 0 {
		return false
	}
	for _, i := range alive {
		if p.players[i].GetBet() != p.betTarget {
			return false
		}
	}
	p.finishPochen(p.bestComboSeat(alive))
	return true
}

// bestComboSeat は生存者のうち最も強い組を持つ席を返す。
//
// **宣言でもブラフでもなく、手札の組の比べ合い。**組が無い者しか残らなければ
// 最初の生存者が取る。
func (p *Poch) bestComboSeat(alive []int) int {
	best, bestSeat := PochCombo{}, alive[0]
	for _, i := range alive {
		hand := make([]*Card, 0, p.players[i].GetCardsSize())
		for j := range p.players[i].GetCardsSize() {
			hand = append(hand, p.players[i].GetCard(j))
		}
		c := PochBestCombo(hand)
		if c.Size > 0 && (best.Size == 0 || PochCombo{Size: c.Size, Rank: pochRankOrder(c.Rank)}.
			Beats(PochCombo{Size: best.Size, Rank: pochRankOrder(best.Rank)})) {
			best, bestSeat = c, i
		}
	}
	return bestSeat
}

// finishPochen は pochen を精算し、ストップスへ移る。
func (p *Poch) finishPochen(winner int) {
	pot := p.board.Take(PochPoolPocher)
	for _, pl := range p.players {
		pot += pl.GetBet()
		pl.ResetBetting()
	}
	p.players[winner].AddChips(pot)
	p.pochenWinner = winner
	p.pochenPot = pot
	p.betTarget = 0
	p.addLog(winner, "pochen", fmt.Sprintf("wins the pochen pot (%d)", pot), nil)

	// **pochen を取った人がストップスを始める。**
	p.currentIdx = winner
	p.stopsSuit = -1
	p.stopsRank = 0
	p.phase = PochPhaseStops
}

// ---- 第 3 段階: ストップス ----

// Play は手札 1 枚を出す。
//
// 並びの途中なら**同じスートの次に高い札**でなければならない。並びが止まって
// いれば好きな札から始められる。
func (p *Poch) Play(player, handIdx int) error {
	if p.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if p.phase != PochPhaseStops {
		return fmt.Errorf("it is not the play stage")
	}
	if player != p.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	pl := p.GetPlayer(player)
	if pl == nil || handIdx < 0 || handIdx >= pl.GetCardsSize() {
		return fmt.Errorf("card index %d out of range", handIdx)
	}
	card := pl.GetCard(handIdx)
	if !p.playable(card) {
		return fmt.Errorf("you must play the next higher card of the same suit")
	}

	pl.RemoveCard(handIdx)
	p.playedPile = append(p.playedPile, card)
	p.stopsSuit = card.GetDesign()
	p.stopsRank = pochRankOrder(card.GetValue())
	p.addLog(player, "play", "plays a card", []*Card{card})

	if pl.GetCardsSize() == 0 {
		p.finishDeal(player)
		return nil
	}
	p.advanceStops(player)
	return nil
}

// playable は card が今出せるかを返す。
func (p *Poch) playable(card *Card) bool {
	if card == nil {
		return false
	}
	if p.stopsSuit < 0 {
		return true
	}
	return card.GetDesign() == p.stopsSuit && pochRankOrder(card.GetValue()) == p.stopsRank+1
}

// advanceStops は次に出せる人へ手番を回す。誰も次を持っていなければ stop で、
// **最後に最高札を出した人**が好きな札から再開する。
func (p *Poch) advanceStops(lastPlayer int) {
	for i := 1; i <= len(p.players); i++ {
		n := (lastPlayer + i) % len(p.players)
		if p.hasPlayable(n) {
			p.currentIdx = n
			return
		}
	}
	p.stopsSuit = -1
	p.stopsRank = 0
	p.currentIdx = lastPlayer
	p.addLog(lastPlayer, "stop", "the run is stopped and restarts here", nil)
}

// hasPlayable は seat が今出せる札を持っているかを返す。
func (p *Poch) hasPlayable(seat int) bool {
	pl := p.GetPlayer(seat)
	if pl == nil {
		return false
	}
	for i := range pl.GetCardsSize() {
		if p.playable(pl.GetCard(i)) {
			return true
		}
	}
	return false
}

// finishDeal はディールを精算する。
//
// 出し切った人が centre pot を取り、**さらに他家から残り札 1 枚につき 1 チップ**
// を受け取る。issue #4415 はこの支払いに触れていないが、ここが最大の変動要因。
func (p *Poch) finishDeal(winner int) {
	centre := p.board.Take(PochPoolCentre)
	p.players[winner].AddChips(centre)

	paid := 0
	for i, pl := range p.players {
		if i == winner {
			continue
		}
		n := pl.GetCardsSize()
		if n == 0 {
			continue
		}
		pl.AddChips(-n)
		paid += n
	}
	p.players[winner].AddChips(paid)
	p.dealWinner = winner
	p.addLog(winner, "deal_end",
		fmt.Sprintf("goes out, takes the centre pot (%d) and %d for the cards left", centre, paid), nil)

	p.dealNo++
	p.phase = PochPhaseDealEnd
	if p.dealNo >= p.config.TargetDeals {
		p.finishGame()
	}
}

// finishGame は最終集計する。チップが最も多い人の勝ち。
func (p *Poch) finishGame() {
	best := 0
	for i := 1; i < len(p.players); i++ {
		if p.players[i].GetChips() > p.players[best].GetChips() {
			best = i
		}
	}
	p.winnerIdx = best
	p.gameEndFlag = true
	p.phase = PochPhaseGameEnd
	p.addLog(best, "game_end", "finishes with the most chips", nil)
}

// NextDeal は次のディールを配る。
func (p *Poch) NextDeal() error {
	if p.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if p.phase != PochPhaseDealEnd {
		return fmt.Errorf("the deal is still in progress")
	}
	p.dealerIdx = (p.dealerIdx + 1) % len(p.players)
	p.dealRound()
	return nil
}

// ---- CPU ----

// PochCpuAction は CPU が選んだ手。
type PochCpuAction struct {
	// Type は "bet" / "fold" / "play"。
	Type string
	// HandIdx は play のときの手札添字 (-1: なし)。
	HandIdx int
}

// PochCpuDecide は idx の CPU が取る手を決める。
func (p *Poch) PochCpuDecide(idx int) PochCpuAction {
	switch p.phase {
	case PochPhasePochen:
		return p.cpuPochen(idx)
	case PochPhaseStops:
		return PochCpuAction{Type: "play", HandIdx: p.cpuPlayIdx(idx)}
	default:
		return PochCpuAction{Type: "fold", HandIdx: -1}
	}
}

// cpuPochen は組の強さだけで判断する。**ブラフはしない** -- 宣言が無い以上、
// 相手に伝える手段が賭け額しかなく、組は最後に必ず開かれる。
func (p *Poch) cpuPochen(idx int) PochCpuAction {
	pl := p.GetPlayer(idx)
	if pl == nil {
		return PochCpuAction{Type: "fold", HandIdx: -1}
	}
	hand := make([]*Card, 0, pl.GetCardsSize())
	for i := range pl.GetCardsSize() {
		hand = append(hand, pl.GetCard(i))
	}
	combo := PochBestCombo(hand)
	behind := p.betTarget - pl.GetBet()
	switch {
	case combo.Size >= 3:
		return PochCpuAction{Type: "bet", HandIdx: -1}
	case combo.Size == 2 && behind <= PochBetUnit:
		return PochCpuAction{Type: "bet", HandIdx: -1}
	case behind <= 0:
		// 追加負担なしで残れるなら残る。
		return PochCpuAction{Type: "bet", HandIdx: -1}
	default:
		return PochCpuAction{Type: "fold", HandIdx: -1}
	}
}

// cpuPlayIdx は出せる札のうち最も低いものを選ぶ。高い札を抱えると止められる。
func (p *Poch) cpuPlayIdx(idx int) int {
	pl := p.GetPlayer(idx)
	if pl == nil {
		return -1
	}
	best, bestRank := -1, 1<<30
	for i := range pl.GetCardsSize() {
		c := pl.GetCard(i)
		if !p.playable(c) {
			continue
		}
		if r := pochRankOrder(c.GetValue()); r < bestRank {
			best, bestRank = i, r
		}
	}
	return best
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (p *Poch) GetPlayers() []*PochPlayer { return p.players }

// GetPlayer は idx のプレイヤーを返す。
func (p *Poch) GetPlayer(idx int) *PochPlayer {
	if idx < 0 || idx >= len(p.players) {
		return nil
	}
	return p.players[idx]
}

// GetPhase は現在のフェーズを返す。
func (p *Poch) GetPhase() PochPhase { return p.phase }

// GetCurrentPlayerIdx は手番のプレイヤー添字を返す。
func (p *Poch) GetCurrentPlayerIdx() int { return p.currentIdx }

// GetBoard は 9 プールの残高を返す。
func (p *Poch) GetBoard() PochBoard { return p.board }

// GetPaySuit は pay suit を返す。
func (p *Poch) GetPaySuit() int { return p.paySuit }

// GetTurnUp は表向きにした余り札を返す。
func (p *Poch) GetTurnUp() *Card { return p.turnUp }

// GetStakingAwards は直近ディールの第 1 段階の結果を返す。
func (p *Poch) GetStakingAwards() []*PochStakingAward { return p.stakingAwards }

// GetBetTarget は現在の賭け額を返す。
func (p *Poch) GetBetTarget() int { return p.betTarget }

// GetPochenWinner は pochen を取った席を返す (-1: 未決着)。
func (p *Poch) GetPochenWinner() int { return p.pochenWinner }

// GetPochenPot は pochen で動いたチップを返す。
func (p *Poch) GetPochenPot() int { return p.pochenPot }

// GetPlayedPile はストップスで出た札を返す。
func (p *Poch) GetPlayedPile() []*Card { return p.playedPile }

// GetStopsSuit は今の並びのスートを返す (-1: 好きな札で開始できる)。
func (p *Poch) GetStopsSuit() int { return p.stopsSuit }

// GetStopsRank は今の並びの最高ランクを返す。
func (p *Poch) GetStopsRank() int { return p.stopsRank }

// GetDealNumber は完了したディール数を返す。
func (p *Poch) GetDealNumber() int { return p.dealNo }

// GetDealWinner は直近ディールで出し切った席を返す (-1: なし)。
func (p *Poch) GetDealWinner() int { return p.dealWinner }

// GetGameEndFlag は決着しているかを返す。
func (p *Poch) GetGameEndFlag() bool { return p.gameEndFlag }

// GetWinnerIdx は勝者の添字を返す (-1: 未決着)。
func (p *Poch) GetWinnerIdx() int { return p.winnerIdx }

// GetConfig はゲーム設定を返す。
func (p *Poch) GetConfig() PochConfig { return p.config }

// SetConfig はゲーム設定をセットする。
func (p *Poch) SetConfig(c PochConfig) { p.config = c }

// GetActionLog は棋譜を返す。
func (p *Poch) GetActionLog() []*ActionLogEntry { return p.actionLog }

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (p *Poch) SetPhaseForTest(ph PochPhase) { p.phase = ph }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (p *Poch) SetCurrentPlayerForTest(idx int) { p.currentIdx = idx }

// SetPaySuitForTest はテスト用に pay suit を差し替える。
func (p *Poch) SetPaySuitForTest(suit int) { p.paySuit = suit }

// SetBoardForTest はテスト用に盤を差し替える。
func (p *Poch) SetBoardForTest(b PochBoard) { p.board = b }

// SetDealNumberForTest はテスト用にディール数を差し替える。
func (p *Poch) SetDealNumberForTest(n int) { p.dealNo = n }

// SetStopsForTest はテスト用に並びの状態を差し替える。
func (p *Poch) SetStopsForTest(suit, rank int) { p.stopsSuit, p.stopsRank = suit, rank }

// ResolveStakingForTest はテスト用に第 1 段階だけを走らせる。
func (p *Poch) ResolveStakingForTest() { p.resolveStaking() }

// addLog は棋譜に 1 件追加する。
func (p *Poch) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	p.actionLog = append(p.actionLog, &ActionLogEntry{
		TurnNumber: len(p.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// pochJSON is the JSON wire format for Poch.
type pochJSON struct {
	Players       []*PochPlayer       `json:"pl"`
	Config        PochConfig          `json:"cfg"`
	Phase         PochPhase           `json:"ph"`
	Board         PochBoard           `json:"bd"`
	PaySuit       int                 `json:"ps"`
	TurnUp        *Card               `json:"tu"`
	StakingAwards []*PochStakingAward `json:"sa"`
	Current       int                 `json:"cur"`
	Dealer        int                 `json:"dl"`
	DealNo        int                 `json:"dn"`
	BetTarget     int                 `json:"bt"`
	PochenWinner  int                 `json:"pw"`
	PochenPot     int                 `json:"pp"`
	StopsSuit     int                 `json:"ss"`
	StopsRank     int                 `json:"sr"`
	PlayedPile    []*Card             `json:"pi"`
	DealWinner    int                 `json:"dw"`
	GameEnd       bool                `json:"ge"`
	WinnerIdx     int                 `json:"wi"`
	ActionLog     []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (p *Poch) MarshalJSON() ([]byte, error) {
	return json.Marshal(pochJSON{
		Players: p.players, Config: p.config, Phase: p.phase, Board: p.board,
		PaySuit: p.paySuit, TurnUp: p.turnUp, StakingAwards: p.stakingAwards,
		Current: p.currentIdx, Dealer: p.dealerIdx, DealNo: p.dealNo,
		BetTarget: p.betTarget, PochenWinner: p.pochenWinner, PochenPot: p.pochenPot,
		StopsSuit: p.stopsSuit, StopsRank: p.stopsRank, PlayedPile: p.playedPile,
		DealWinner: p.dealWinner, GameEnd: p.gameEndFlag, WinnerIdx: p.winnerIdx,
		ActionLog: p.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// KV から戻る生バイト列は信用できないので、席数に合わせて詰め直し、設定を
// 検証する。**盤のチップは持ち越しそのもの**なので、そのまま復元する。
func (p *Poch) UnmarshalJSON(data []byte) error {
	var raw pochJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Players) != PochPlayerCnt {
		return fmt.Errorf("expected %d players, got %d", PochPlayerCnt, len(raw.Players))
	}
	if err := raw.Config.Validate(); err != nil {
		return err
	}
	if raw.Phase < PochPhaseStaking || raw.Phase > PochPhaseGameEnd {
		return fmt.Errorf("unknown phase: %d", raw.Phase)
	}

	p.players = raw.Players
	p.config = raw.Config
	p.phase = raw.Phase
	p.board = raw.Board
	p.paySuit = raw.PaySuit
	p.turnUp = raw.TurnUp
	p.dealNo = raw.DealNo
	p.betTarget = raw.BetTarget
	p.pochenPot = raw.PochenPot
	p.stopsRank = raw.StopsRank
	p.playedPile = raw.PlayedPile
	p.gameEndFlag = raw.GameEnd
	p.actionLog = raw.ActionLog

	p.currentIdx = clampPochIdx(raw.Current, len(p.players))
	p.dealerIdx = clampPochIdx(raw.Dealer, len(p.players))
	p.pochenWinner = clampPochSeatOrNone(raw.PochenWinner, len(p.players))
	p.dealWinner = clampPochSeatOrNone(raw.DealWinner, len(p.players))
	p.winnerIdx = clampPochSeatOrNone(raw.WinnerIdx, len(p.players))

	// **-1 は「好きな札で開始できる」という意味を持つ**ので潰さない。
	p.stopsSuit = raw.StopsSuit
	if p.stopsSuit < -1 || p.stopsSuit > CardDesignDiamond {
		p.stopsSuit = -1
	}

	p.stakingAwards = make([]*PochStakingAward, 0, len(raw.StakingAwards))
	for _, a := range raw.StakingAwards {
		if a == nil || a.Pool < 0 || a.Pool >= PochPoolCount {
			continue
		}
		if a.Player < 0 || a.Player >= len(p.players) {
			continue
		}
		p.stakingAwards = append(p.stakingAwards, a)
	}
	sort.SliceStable(p.stakingAwards, func(i, j int) bool {
		return p.stakingAwards[i].Pool < p.stakingAwards[j].Pool
	})
	return nil
}

// clampPochIdx は席番号を 0..n-1 に収める。
func clampPochIdx(idx, n int) int {
	if idx < 0 || idx >= n {
		return 0
	}
	return idx
}

// clampPochSeatOrNone は席番号を -1..n-1 に収める。
func clampPochSeatOrNone(idx, n int) int {
	if idx < -1 || idx >= n {
		return -1
	}
	return idx
}
