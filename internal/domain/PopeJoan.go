//go:build !js || !wasm || extra3

// Package domain — ポープ・ジョーン (Pope Joan) のドメインモデル。
//
// 18〜19 世紀イギリスのストップス系の祖型。**51 枚**(標準 52 枚から ♦8 を抜く)、
// 専用盤の 8 区画。
//
// # issue #4389 の仕様案との相違
//
// 51 枚とストップス進行は合っているが、盤・配り方・精算の 3 つに誤りがある。
//
//   - issue は区画を「**8 種の絵札** + ポープ」とするが、盤は **8 区画**で
//     Ace / King / Queen / Jack / Game / Pope (♦9) / Matrimony (トランプの K-Q) /
//     Intrigue (トランプの Q-J)。絵札の数とは関係がない
//   - issue は「事前アンティを**配分**してから配札する」とするが、分け方は固定。
//     **ディーラーが Pope に 6、Matrimony と Intrigue に各 2、残り 5 区画に各 1**
//     (合計 15) を置く «dressing the board»。プレイヤーが賭けるものではない
//   - issue が触れていない: **プレイヤー数より 1 つ多く配る**。余った「dead
//     hand」の最後の 1 枚を表向きにし、それが**トランプ**になる。しかもその札が
//     Pope / A / K / Q / J なら**ディーラーがその区画を即座に総取り**する
//   - issue は「手札の**最小連番**から順に場に出し」とするが、最初のリードは
//     «any suit, as long as it is his lowest card» -- **スートは自由**で、条件は
//     「自分の最も低い札」であることだけ
//   - issue は「区画に該当する札 (K、Q、J、ポープなど) を出した際は…」とするが、
//     **トランプの** A/K/Q/J と Pope に限る
//   - issue は「♦8 が抜かれているため ♦7 の次でストップ」とするが、**K でも
//     ストップする** (最高位なので次が無い)。そして stop を出した人が次の並びを
//     始める
//   - issue は「上がった者は残りプレイヤーの手札枚数分のチップを徴収」とするが、
//     **Pope を持っている人は免除**される
//   - issue が触れていない: **取られなかった区画は次のディールへ持ち越される**
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// PopeJoanPlayerCnt はプレイヤー数 (4 人。原典は 3〜8 人)。
const PopeJoanPlayerCnt = 4

// PopeJoanDeckSize は ♦8 を抜いた 51 枚。
const PopeJoanDeckSize = 51

// PopeJoanPopeRank は Pope (♦9) のランク。
const PopeJoanPopeRank = 9

// PopeJoanPhase はゲームフェーズ。
type PopeJoanPhase int

// Pope Joan のフェーズ定数
const (
	// PopeJoanPhasePlay 手番進行中
	PopeJoanPhasePlay PopeJoanPhase = iota
	// PopeJoanPhaseDealEnd 1 ディール終了
	PopeJoanPhaseDealEnd
	// PopeJoanPhaseGameEnd 決着
	PopeJoanPhaseGameEnd
)

// PopeJoanAward は区画が誰にいくら渡ったかの記録。
type PopeJoanAward struct {
	Compartment PopeJoanCompartment
	Player      int
	Chips       int
	// ByTurnUp はめくりトランプでディーラーが即座に取ったか。
	ByTurnUp bool
}

// popeJoanIsPope は c が ♦9 かを返す。
func popeJoanIsPope(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignDiamond && c.GetValue() == PopeJoanPopeRank
}

// popeJoanRankOrder は A を最上位に置いた順位を返す。A-K-Q-J-10-…-2。
func popeJoanRankOrder(rank int) int {
	if rank == 1 {
		return 14
	}
	return rank
}

// newPopeJoanDeck は ♦8 を抜いた 51 枚を生成する。
func newPopeJoanDeck() []*Card {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	deck := make([]*Card, 0, PopeJoanDeckSize)
	for _, s := range suits {
		for v := 1; v <= 13; v++ {
			// **♦8 を抜くのがこの game の仕掛け。**♦7 の次が存在しないので、
			// そこで必ず並びが止まる。
			if s == CardDesignDiamond && v == 8 {
				continue
			}
			deck = append(deck, NewCard(s, v, true))
		}
	}
	return deck
}

// popeJoanShuffle は Fisher-Yates。
func popeJoanShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// PopeJoan はポープ・ジョーンのゲームクラス。
type PopeJoan struct {
	players []*PopeJoanPlayer
	config  PopeJoanConfig
	phase   PopeJoanPhase
	board   PopeJoanBoard

	// trumpSuit は dead hand の最後の 1 枚で決まるトランプ。
	trumpSuit int
	// turnUp はそのめくり札そのもの。
	turnUp *Card
	// deadHand は配り切れなかった 1 人分。誰も使わない。
	deadHand []*Card

	// awards はこのディールで区画が動いた記録。
	awards []*PopeJoanAward

	currentIdx int
	dealerIdx  int
	dealNo     int

	// runSuit / runRank は今の並びの続き。runSuit < 0 なら好きな札で開始。
	runSuit int
	runRank int
	// playedPile は場に出た札。
	playedPile []*Card
	// matrimonyKing / intrigueQueen は「同じ人が 2 枚とも出したか」を見るための
	// 途中経過。トランプの K を出した席、Q を出した席を覚えておく。
	trumpKingBy  int
	trumpQueenBy int
	trumpJackBy  int

	dealWinner  int
	gameEndFlag bool
	winnerIdx   int
	actionLog   []*ActionLogEntry
}

// NewPopeJoan はコンストラクタ。
func NewPopeJoan(players []*PopeJoanPlayer, config PopeJoanConfig) *PopeJoan {
	return &PopeJoan{
		players:      players,
		config:       config,
		trumpSuit:    -1,
		runSuit:      -1,
		trumpKingBy:  -1,
		trumpQueenBy: -1,
		trumpJackBy:  -1,
		dealWinner:   -1,
		winnerIdx:    -1,
	}
}

// NewDefaultPopeJoan は標準の 4 人セットアップを返す。
func NewDefaultPopeJoan() *PopeJoan {
	players := make([]*PopeJoanPlayer, 0, PopeJoanPlayerCnt)
	players = append(players, NewPopeJoanPlayer(true))
	for range PopeJoanPlayerCnt - 1 {
		players = append(players, NewPopeJoanPlayer(false))
	}
	return NewPopeJoan(players, DefaultPopeJoanConfig())
}

// Reset はゲーム全体を初期化する。
func (p *PopeJoan) Reset() {
	p.board = PopeJoanBoard{}
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

// dealRound は 1 ディールを配る。
func (p *PopeJoan) dealRound() {
	for _, pl := range p.players {
		pl.ResetDeal()
	}
	p.awards = nil
	p.runSuit = -1
	p.runRank = 0
	p.playedPile = nil
	p.trumpKingBy = -1
	p.trumpQueenBy = -1
	p.trumpJackBy = -1
	p.dealWinner = -1

	// **ディーラーが固定の内訳で置く。**プレイヤーが配分するのではない。
	// 取られなかった区画のぶんはそのまま残っているので、上に積み増される。
	spent := p.board.Dress()
	p.players[p.dealerIdx].AddChips(-spent)

	deck := newPopeJoanDeck()
	popeJoanShuffle(deck)

	// **プレイヤー数より 1 つ多く配る。**余った 1 人分が dead hand で、その
	// 最後の 1 枚を表にしてトランプを決める。
	//
	// **配り始めはディーラーの左隣から。**51 枚を 5 つの手 (4 人 + dead hand)
	// に配ると 1 枚余り、先頭の席だけ 1 枚多くなる。i%hands で始めると余りが
	// 毎回席 0 = 人間に落ち、**ディールを重ねるほど人間だけが不利**になる
	// (先に出し切ると他家から取り立て、抱えた札は 1 枚 1 チップの負債になる
	// ため、枚数が多いことは純粋な不利)。他のゲームと同様に回す。
	hands := len(p.players) + 1
	p.deadHand = nil
	for i, c := range deck {
		seat := (p.dealerIdx + 1 + i) % hands
		if seat == len(p.players) {
			p.deadHand = append(p.deadHand, c)
			continue
		}
		p.players[seat].AddCard(c)
	}
	p.turnUp = p.deadHand[len(p.deadHand)-1]
	p.trumpSuit = p.turnUp.GetDesign()

	p.phase = PopeJoanPhasePlay
	p.addLog(-1, "deal", fmt.Sprintf("trump is %d", p.trumpSuit), []*Card{p.turnUp})
	p.resolveTurnUp()

	p.currentIdx = (p.dealerIdx + 1) % len(p.players)
}

// resolveTurnUp はめくり札がその場で区画を払う場合を処理する。
//
// **めくり札が Pope / A / K / Q / J なら、ディーラーがその区画を即座に取る。**
// issue #4389 はこの規則に触れていないが、Pope 区画は 6 枚積まれるので効きが
// 大きい。
func (p *PopeJoan) resolveTurnUp() {
	if popeJoanIsPope(p.turnUp) {
		p.award(PopeJoanPope, p.dealerIdx, true)
		return
	}
	if p.turnUp.GetDesign() != p.trumpSuit {
		return
	}
	if comp, ok := PopeJoanCompartmentForRank(p.turnUp.GetValue()); ok {
		p.award(comp, p.dealerIdx, true)
	}
}

// award は区画のチップを seat に渡し、記録を残す。
func (p *PopeJoan) award(comp PopeJoanCompartment, seat int, byTurnUp bool) {
	n := p.board.Take(comp)
	if n == 0 {
		return
	}
	p.players[seat].AddChips(n)
	p.awards = append(p.awards, &PopeJoanAward{
		Compartment: comp, Player: seat, Chips: n, ByTurnUp: byTurnUp,
	})
	detail := fmt.Sprintf("takes the %s compartment (%d)", comp, n)
	if byTurnUp {
		detail += " from the turn-up"
	}
	p.addLog(seat, "award", detail, nil)
}

// Play は手札 1 枚を出す。
//
// 並びが止まっていれば**自分の最も低い札**なら何でも出せる。並びの途中なら
// **同じスートの次に高い札**でなければならない。
func (p *PopeJoan) Play(player, handIdx int) error {
	if p.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if p.phase != PopeJoanPhasePlay {
		return fmt.Errorf("the deal is not in progress")
	}
	if player != p.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	pl := p.GetPlayer(player)
	if pl == nil || handIdx < 0 || handIdx >= pl.GetCardsSize() {
		return fmt.Errorf("card index %d out of range", handIdx)
	}
	card := pl.GetCard(handIdx)
	if err := p.checkPlayable(player, card); err != nil {
		return err
	}

	pl.RemoveCard(handIdx)
	p.playedPile = append(p.playedPile, card)
	p.runSuit = card.GetDesign()
	p.runRank = popeJoanRankOrder(card.GetValue())
	p.addLog(player, "play", "plays a card", []*Card{card})
	p.payForCard(player, card)

	if pl.GetCardsSize() == 0 {
		p.finishDeal(player)
		return nil
	}
	p.advance(player)
	return nil
}

// PopeJoanValidPlays は player が今出せる手札インデックスを返す。
//
// **判定は checkPlayable をそのまま呼ぶ。**規則を書き写すと、示した手が拒否
// されるようになる (#4934)。**自由リードでも全部が出せるわけではない** —
// 新しい並びは自分の最も低い札で始めなければならない。手番でない場合は nil。
func (p *PopeJoan) PopeJoanValidPlays(player int) []int {
	if player != p.currentIdx {
		return nil
	}
	pl := p.GetPlayer(player)
	if pl == nil {
		return nil
	}
	out := make([]int, 0, pl.GetCardsSize())
	for i := range pl.GetCardsSize() {
		if p.checkPlayable(player, pl.GetCard(i)) == nil {
			out = append(out, i)
		}
	}
	return out
}

// checkPlayable は card が今出せるかを確かめる。
func (p *PopeJoan) checkPlayable(player int, card *Card) error {
	if card == nil {
		return fmt.Errorf("no such card")
	}
	if p.runSuit < 0 {
		// **スートは自由。条件は「自分の最も低い札」であること。**
		if !p.isLowestInHand(player, card) {
			return fmt.Errorf("a new run must be led with your lowest card")
		}
		return nil
	}
	if card.GetDesign() != p.runSuit || popeJoanRankOrder(card.GetValue()) != p.runRank+1 {
		return fmt.Errorf("you must play the next higher card of the same suit")
	}
	return nil
}

// isLowestInHand は card が player の手札で最も低いかを返す。
func (p *PopeJoan) isLowestInHand(player int, card *Card) bool {
	pl := p.GetPlayer(player)
	want := popeJoanRankOrder(card.GetValue())
	for i := range pl.GetCardsSize() {
		c := pl.GetCard(i)
		if c != nil && popeJoanRankOrder(c.GetValue()) < want {
			return false
		}
	}
	return true
}

// payForCard は出した札が区画を取るかを判定する。
//
// **トランプの札に限る。**Matrimony と Intrigue は 2 枚組で、同じ人が両方を
// 出したときにだけ払う。
func (p *PopeJoan) payForCard(player int, card *Card) {
	if popeJoanIsPope(card) {
		p.award(PopeJoanPope, player, false)
		return
	}
	if card.GetDesign() != p.trumpSuit {
		return
	}
	if comp, ok := PopeJoanCompartmentForRank(card.GetValue()); ok {
		p.award(comp, player, false)
	}
	// **両方向を見る。**並びは必ず上がっていく (J=11 < Q=12 < K=13) ので、
	// 片方向しか見ないと自然な順序のほうが落ちる。Intrigue を「Q を出した
	// あとに J」だけで判定すると、J→Q という普通の順序で永久に払われない。
	switch card.GetValue() {
	case 13:
		p.trumpKingBy = player
		if p.trumpQueenBy == player {
			p.award(PopeJoanMatrimony, player, false)
		}
	case 12:
		p.trumpQueenBy = player
		if p.trumpKingBy == player {
			p.award(PopeJoanMatrimony, player, false)
		}
		if p.trumpJackBy == player {
			p.award(PopeJoanIntrigue, player, false)
		}
	case 11:
		p.trumpJackBy = player
		if p.trumpQueenBy == player {
			p.award(PopeJoanIntrigue, player, false)
		}
	}
}

// advance は次に出せる人へ手番を回す。誰も次を持っていなければ stop で、
// **最後に札を出した人**が好きな札から再開する。
//
// K を出した時点でも並びは終わる (最高位なので次が無い)。この関数は
// 「誰も次を持っていない」で一様に扱うので、K も ♦7 も同じ経路に落ちる。
func (p *PopeJoan) advance(lastPlayer int) {
	for i := 1; i <= len(p.players); i++ {
		n := (lastPlayer + i) % len(p.players)
		if p.hasNextCard(n) {
			p.currentIdx = n
			return
		}
	}
	p.runSuit = -1
	p.runRank = 0
	p.currentIdx = lastPlayer
	p.addLog(lastPlayer, "stop", "the run is stopped and restarts here", nil)
}

// hasNextCard は seat が並びの続きを持っているかを返す。
func (p *PopeJoan) hasNextCard(seat int) bool {
	if p.runSuit < 0 {
		return false
	}
	pl := p.GetPlayer(seat)
	if pl == nil {
		return false
	}
	for i := range pl.GetCardsSize() {
		c := pl.GetCard(i)
		if c != nil && c.GetDesign() == p.runSuit && popeJoanRankOrder(c.GetValue()) == p.runRank+1 {
			return true
		}
	}
	return false
}

// finishDeal はディールを精算する。
//
// 出し切った人が Game 区画を取り、**さらに他家から残り札 1 枚につき 1 チップ**
// を受け取る。ただし **Pope を手札に持っている人は免除**される。
func (p *PopeJoan) finishDeal(winner int) {
	p.award(PopeJoanGame, winner, false)

	paid := 0
	for i, pl := range p.players {
		if i == winner {
			continue
		}
		n := pl.GetCardsSize()
		if n == 0 {
			continue
		}
		// **Pope を抱えている人は払わない。**6 枚積まれた区画を取り逃した
		// 代償が、この免除で釣り合っている。
		if p.holdsPope(i) {
			p.addLog(i, "excused", "holds the Pope and is excused payment", nil)
			continue
		}
		pl.AddChips(-n)
		paid += n
	}
	p.players[winner].AddChips(paid)
	p.dealWinner = winner
	p.addLog(winner, "deal_end", fmt.Sprintf("goes out and collects %d for the cards left", paid), nil)

	p.dealNo++
	p.phase = PopeJoanPhaseDealEnd
	if p.dealNo >= p.config.TargetDeals {
		p.finishGame()
	}
}

// holdsPope は seat が Pope を手札に持っているかを返す。
func (p *PopeJoan) holdsPope(seat int) bool {
	pl := p.GetPlayer(seat)
	if pl == nil {
		return false
	}
	for i := range pl.GetCardsSize() {
		if popeJoanIsPope(pl.GetCard(i)) {
			return true
		}
	}
	return false
}

// finishGame は最終集計する。チップが最も多い人の勝ち。
func (p *PopeJoan) finishGame() {
	best := 0
	for i := 1; i < len(p.players); i++ {
		if p.players[i].GetChips() > p.players[best].GetChips() {
			best = i
		}
	}
	p.winnerIdx = best
	p.gameEndFlag = true
	p.phase = PopeJoanPhaseGameEnd
	p.addLog(best, "game_end", "finishes with the most chips", nil)
}

// NextDeal は次のディールを配る。
func (p *PopeJoan) NextDeal() error {
	if p.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if p.phase != PopeJoanPhaseDealEnd {
		return fmt.Errorf("the deal is still in progress")
	}
	p.dealerIdx = (p.dealerIdx + 1) % len(p.players)
	p.dealRound()
	return nil
}

// PopeJoanCpuDecide は idx の CPU が出す手札の添字を返す (-1: 出せない)。
//
// 並びの途中なら唯一の続きを出す。止まっていれば最も低い札から始めるしかない
// ので、選択の余地は実質ない。
func (p *PopeJoan) PopeJoanCpuDecide(idx int) int {
	pl := p.GetPlayer(idx)
	if pl == nil {
		return -1
	}
	best, bestRank := -1, 1<<30
	for i := range pl.GetCardsSize() {
		c := pl.GetCard(i)
		if p.checkPlayable(idx, c) != nil {
			continue
		}
		if r := popeJoanRankOrder(c.GetValue()); r < bestRank {
			best, bestRank = i, r
		}
	}
	return best
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (p *PopeJoan) GetPlayers() []*PopeJoanPlayer { return p.players }

// GetPlayer は idx のプレイヤーを返す。
func (p *PopeJoan) GetPlayer(idx int) *PopeJoanPlayer {
	if idx < 0 || idx >= len(p.players) {
		return nil
	}
	return p.players[idx]
}

// GetPhase は現在のフェーズを返す。
func (p *PopeJoan) GetPhase() PopeJoanPhase { return p.phase }

// GetCurrentPlayerIdx は手番のプレイヤー添字を返す。
func (p *PopeJoan) GetCurrentPlayerIdx() int { return p.currentIdx }

// GetBoard は 8 区画の残高を返す。
func (p *PopeJoan) GetBoard() PopeJoanBoard { return p.board }

// GetTrumpSuit はトランプを返す。
func (p *PopeJoan) GetTrumpSuit() int { return p.trumpSuit }

// GetTurnUp は dead hand の最後の 1 枚を返す。
func (p *PopeJoan) GetTurnUp() *Card { return p.turnUp }

// GetAwards はこのディールで区画が動いた記録を返す。
func (p *PopeJoan) GetAwards() []*PopeJoanAward { return p.awards }

// GetPlayedPile は場に出た札を返す。
func (p *PopeJoan) GetPlayedPile() []*Card { return p.playedPile }

// GetRunSuit は今の並びのスートを返す (-1: 好きな札で開始できる)。
func (p *PopeJoan) GetRunSuit() int { return p.runSuit }

// GetRunRank は今の並びの最高ランクを返す。
func (p *PopeJoan) GetRunRank() int { return p.runRank }

// GetDealNumber は完了したディール数を返す。
func (p *PopeJoan) GetDealNumber() int { return p.dealNo }

// GetDealWinner は直近ディールで出し切った席を返す (-1: なし)。
func (p *PopeJoan) GetDealWinner() int { return p.dealWinner }

// GetGameEndFlag は決着しているかを返す。
func (p *PopeJoan) GetGameEndFlag() bool { return p.gameEndFlag }

// GetWinnerIdx は勝者の添字を返す (-1: 未決着)。
func (p *PopeJoan) GetWinnerIdx() int { return p.winnerIdx }

// GetConfig はゲーム設定を返す。
func (p *PopeJoan) GetConfig() PopeJoanConfig { return p.config }

// SetConfig はゲーム設定をセットする。
func (p *PopeJoan) SetConfig(c PopeJoanConfig) { p.config = c }

// GetActionLog は棋譜を返す。
func (p *PopeJoan) GetActionLog() []*ActionLogEntry { return p.actionLog }

// SetPhaseForTest はテスト用にフェーズを差し替える。
func (p *PopeJoan) SetPhaseForTest(ph PopeJoanPhase) { p.phase = ph }

// SetCurrentPlayerForTest はテスト用に手番を差し替える。
func (p *PopeJoan) SetCurrentPlayerForTest(idx int) { p.currentIdx = idx }

// SetTrumpSuitForTest はテスト用にトランプを差し替える。
func (p *PopeJoan) SetTrumpSuitForTest(suit int) { p.trumpSuit = suit }

// SetBoardForTest はテスト用に盤を差し替える。
func (p *PopeJoan) SetBoardForTest(b PopeJoanBoard) { p.board = b }

// SetRunForTest はテスト用に並びの状態を差し替える。
func (p *PopeJoan) SetRunForTest(suit, rank int) { p.runSuit, p.runRank = suit, rank }

// SetDealNumberForTest はテスト用にディール数を差し替える。
func (p *PopeJoan) SetDealNumberForTest(n int) { p.dealNo = n }

// SetTurnUpForTest はテスト用にめくり札を差し替える。
func (p *PopeJoan) SetTurnUpForTest(c *Card) { p.turnUp = c }

// ResolveTurnUpForTest はテスト用にめくり札の精算だけを走らせる。
func (p *PopeJoan) ResolveTurnUpForTest() { p.resolveTurnUp() }

// addLog は棋譜に 1 件追加する。
func (p *PopeJoan) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	p.actionLog = append(p.actionLog, &ActionLogEntry{
		TurnNumber: len(p.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// popeJoanJSON is the JSON wire format for PopeJoan.
type popeJoanJSON struct {
	Players    []*PopeJoanPlayer `json:"pl"`
	Config     PopeJoanConfig    `json:"cfg"`
	Phase      PopeJoanPhase     `json:"ph"`
	Board      PopeJoanBoard     `json:"bd"`
	TrumpSuit  int               `json:"ts"`
	TurnUp     *Card             `json:"tu"`
	DeadHand   []*Card           `json:"dh"`
	Awards     []*PopeJoanAward  `json:"aw"`
	Current    int               `json:"cur"`
	Dealer     int               `json:"dl"`
	DealNo     int               `json:"dn"`
	RunSuit    int               `json:"rs"`
	RunRank    int               `json:"rr"`
	PlayedPile []*Card           `json:"pi"`
	KingBy     int               `json:"kb"`
	QueenBy    int               `json:"qb"`
	JackBy     int               `json:"jb"`
	DealWinner int               `json:"dw"`
	GameEnd    bool              `json:"ge"`
	WinnerIdx  int               `json:"wi"`
	ActionLog  []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (p *PopeJoan) MarshalJSON() ([]byte, error) {
	return json.Marshal(popeJoanJSON{
		Players: p.players, Config: p.config, Phase: p.phase, Board: p.board,
		TrumpSuit: p.trumpSuit, TurnUp: p.turnUp, DeadHand: p.deadHand, Awards: p.awards,
		Current: p.currentIdx, Dealer: p.dealerIdx, DealNo: p.dealNo,
		RunSuit: p.runSuit, RunRank: p.runRank, PlayedPile: p.playedPile,
		KingBy: p.trumpKingBy, QueenBy: p.trumpQueenBy, JackBy: p.trumpJackBy,
		DealWinner: p.dealWinner, GameEnd: p.gameEndFlag, WinnerIdx: p.winnerIdx,
		ActionLog: p.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// KV から戻る生バイト列は信用できないので、席数に合わせて詰め直し、設定を
// 検証する。**盤のチップは持ち越しそのもの**なので、そのまま復元する。
func (p *PopeJoan) UnmarshalJSON(data []byte) error {
	var raw popeJoanJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Players) != PopeJoanPlayerCnt {
		return fmt.Errorf("expected %d players, got %d", PopeJoanPlayerCnt, len(raw.Players))
	}
	if err := raw.Config.Validate(); err != nil {
		return err
	}
	if raw.Phase < PopeJoanPhasePlay || raw.Phase > PopeJoanPhaseGameEnd {
		return fmt.Errorf("unknown phase: %d", raw.Phase)
	}

	p.players = raw.Players
	p.config = raw.Config
	p.phase = raw.Phase
	p.board = raw.Board
	// 他のフィールドと同じく境界を見る。比較にしか使わないので実害は無いが、
	// 「KV から戻る生バイト列は信用しない」という方針に合わせる。
	p.trumpSuit = raw.TrumpSuit
	if p.trumpSuit < CardDesignJoker || p.trumpSuit > CardDesignDiamond {
		p.trumpSuit = -1
	}
	p.turnUp = raw.TurnUp
	p.deadHand = raw.DeadHand
	p.dealNo = raw.DealNo
	p.runRank = raw.RunRank
	p.playedPile = raw.PlayedPile
	p.gameEndFlag = raw.GameEnd
	p.actionLog = raw.ActionLog

	p.currentIdx = clampPopeJoanIdx(raw.Current, len(p.players))
	p.dealerIdx = clampPopeJoanIdx(raw.Dealer, len(p.players))
	p.trumpKingBy = clampPopeJoanSeatOrNone(raw.KingBy, len(p.players))
	p.trumpQueenBy = clampPopeJoanSeatOrNone(raw.QueenBy, len(p.players))
	p.trumpJackBy = clampPopeJoanSeatOrNone(raw.JackBy, len(p.players))
	p.dealWinner = clampPopeJoanSeatOrNone(raw.DealWinner, len(p.players))
	p.winnerIdx = clampPopeJoanSeatOrNone(raw.WinnerIdx, len(p.players))

	// **-1 は「好きな札で始められる」という意味を持つ**ので潰さない。
	p.runSuit = raw.RunSuit
	if p.runSuit < -1 || p.runSuit > CardDesignDiamond {
		p.runSuit = -1
	}

	p.awards = make([]*PopeJoanAward, 0, len(raw.Awards))
	for _, a := range raw.Awards {
		if a == nil || a.Compartment < 0 || a.Compartment >= PopeJoanCompartmentCount {
			continue
		}
		if a.Player < 0 || a.Player >= len(p.players) {
			continue
		}
		p.awards = append(p.awards, a)
	}
	return nil
}

// clampPopeJoanIdx は席番号を 0..n-1 に収める。
func clampPopeJoanIdx(idx, n int) int {
	if idx < 0 || idx >= n {
		return 0
	}
	return idx
}

// clampPopeJoanSeatOrNone は席番号を -1..n-1 に収める。
func clampPopeJoanSeatOrNone(idx, n int) int {
	if idx < -1 || idx >= n {
		return -1
	}
	return idx
}
