//go:build !js || !wasm || extra3

// Package domain — ボストン (Boston) のドメインモデル。
//
// 18 世紀ホイストから派生した 4 人のトリックゲーム。52 枚、各自 13 枚。
//
// # issue #4394 の仕様案との相違
//
//   - issue は「『5トリック』〜『13トリック』、『リトルミゼール』『グランド
//     ミゼール』を段階的に競り上げ」とするが、**序列は交互に挟まる**。
//     Little Misère は **7 トリックより下**、Grand Misère は **9 トリックより
//     下**である。トリック宣言を並べたあとにミゼールを置くと競りの意味が変わる
//   - issue は **Piccolissimo**（**ちょうど 1 トリック**取る）に触れていない。
//     7 と 8 の間に入る第 3 の型で、0 トリックでも失敗になる
//   - issue は「ミゼール成立時は…オープンハンドの場合あり」とするが、
//     **on the Table はオプションではなく独立した上位宣言**である
//   - issue は「4 人・**各自個人戦**」とするが、**トリック数の宣言 (5〜10) では
//     パートナーを指名できる**。単独固定なのはミゼール系・ピッコリッシモ・
//     スラムだけ。1 対 3 と 2 対 2 がオークションで決まるのが Boston 系の特徴
//   - issue は配り方の刻みに触れていない。**4 枚・4 枚・5 枚**の 3 回に分ける
//
// 序列表そのものは [BostonBid.go] にある。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// BostonPlayerCnt はプレイヤー数。
const BostonPlayerCnt = 4

// BostonHandSize は各プレイヤーの手札枚数。
const BostonHandSize = 13

// bostonDealPattern は配り方の刻み。
//
// **13 枚を一度に配らない。**4 枚・4 枚・5 枚の 3 回に分ける。
var bostonDealPattern = []int{4, 4, 5}

// BostonPhase はゲームフェーズ。
type BostonPhase int

// Boston のフェーズ定数
const (
	// BostonPhaseBid 宣言
	BostonPhaseBid BostonPhase = iota
	// BostonPhaseCallPartner 落札者がパートナーを指名するか決める
	BostonPhaseCallPartner
	// BostonPhasePlay トリックプレイ
	BostonPhasePlay
	// BostonPhaseHandEnd 局終了 (精算済み)
	BostonPhaseHandEnd
	// BostonPhaseGameEnd ゲーム終了
	BostonPhaseGameEnd
)

// BostonBidRecord は 1 件の宣言。
type BostonBidRecord struct {
	Player int
	Level  BostonBidLevel
	// Suit は切札 (ミゼール系は 0)。
	Suit int
}

// Boston はボストンのゲームクラス。
type Boston struct {
	players []*BostonPlayer
	config  BostonConfig
	phase   BostonPhase

	dealerIdx  int
	currentIdx int
	bidIdx     int
	bids       []*BostonBidRecord
	passCount  int

	highBid *BostonBidRecord
	// declarerIdx は落札者 (-1 なら未確定)。
	declarerIdx int
	// partnerIdx は指名されたパートナー (-1 なら単独)。
	partnerIdx int
	trumpSuit  int
	// exposed は落札者の手札を公開しているか (on the Table の宣言)。
	exposed bool

	trick       []*Card
	trickLeader int
	trickNumber int
	// tricksWon は各席が取ったトリック数。
	tricksWon [BostonPlayerCnt]int

	// bidMade は直前の局で落札側が達成したか。
	bidMade bool
	// chips は各席の通算。
	chips [BostonPlayerCnt]int

	handNumber  int
	targetHands int
	gameEndFlag bool
	winnerIdx   int

	actionLogBase
}

// NewBoston コンストラクタ
func NewBoston(players []*BostonPlayer, config BostonConfig) *Boston {
	return &Boston{players: players, config: config, winnerIdx: -1, declarerIdx: -1, partnerIdx: -1}
}

// NewDefaultBoston はデフォルト構成のゲームを返す。
func NewDefaultBoston() *Boston {
	players := make([]*BostonPlayer, BostonPlayerCnt)
	for i := range players {
		players[i] = NewBostonPlayer(i == 0)
	}
	return NewBoston(players, DefaultBostonConfig())
}

// ---- 進行 ----

// Reset ゲーム初期化
func (b *Boston) Reset() {
	b.gameEndFlag = false
	b.winnerIdx = -1
	b.chips = [BostonPlayerCnt]int{}
	b.handNumber = 0
	b.dealerIdx = 0
	b.targetHands = b.config.TargetHands
	b.actionLog = nil
	b.beginHand()
}

// beginHand は 1 局を配って宣言へ入る。
func (b *Boston) beginHand() {
	b.handNumber++
	b.phase = BostonPhaseBid
	b.bids = nil
	b.passCount = 0
	b.highBid = nil
	b.declarerIdx = -1
	b.partnerIdx = -1
	b.trumpSuit = 0
	b.exposed = false
	b.trick = nil
	b.trickNumber = 0
	b.trickLeader = -1
	b.tricksWon = [BostonPlayerCnt]int{}
	b.bidMade = false

	for _, p := range b.players {
		p.ResetRound()
	}

	deck := newBostonDeck()
	bostonShuffle(deck)
	// **4 枚・4 枚・5 枚の 3 回に分けて配る。**
	pos := 0
	for _, n := range bostonDealPattern {
		for i := range BostonPlayerCnt {
			for range n {
				b.players[i].AddCard(deck[pos])
				pos++
			}
		}
	}

	b.bidIdx = (b.dealerIdx + 1) % BostonPlayerCnt
	b.addLog(-1, "deal", fmt.Sprintf("%d cards each in %v", BostonHandSize, bostonDealPattern), nil)
}

// newBostonDeck は 52 枚のデッキを返す。
func newBostonDeck() []*Card {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	deck := make([]*Card, 0, 52)
	for _, s := range suits {
		for v := 1; v <= 13; v++ {
			deck = append(deck, NewCard(s, v, true))
		}
	}
	return deck
}

// bostonShuffle は Fisher-Yates。
func bostonShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// checkBidTurn は宣言できる状態かを確かめる。
func (b *Boston) checkBidTurn(player int) error {
	if b.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if b.phase != BostonPhaseBid {
		return fmt.Errorf("bidding is not in progress")
	}
	if player != b.bidIdx {
		return fmt.Errorf("it is not player %d's turn to bid", player)
	}
	return nil
}

// Bid は宣言する。
//
// suit は切札。**ミゼール系とピッコリッシモでは無視される。**
func (b *Boston) Bid(player int, level BostonBidLevel, suit int) error {
	if err := b.checkBidTurn(player); err != nil {
		return err
	}
	if level <= BostonBidPass || level >= BostonBidLevelCount {
		return fmt.Errorf("bad bid level: %d", level)
	}
	if b.highBid != nil && level <= b.highBid.Level {
		return fmt.Errorf("a bid must beat the standing %s", BostonBidName(b.highBid.Level))
	}
	if BostonBidNeedsTrump(level) {
		if suit < CardDesignSpade || suit > CardDesignDiamond {
			return fmt.Errorf("bad trump suit: %d", suit)
		}
	} else {
		// 切札なしの宣言では捨てる。
		suit = 0
	}
	rec := &BostonBidRecord{Player: player, Level: level, Suit: suit}
	b.bids = append(b.bids, rec)
	b.highBid = rec
	b.passCount = 0
	b.addLog(player, "bid", fmt.Sprintf("bids %s", BostonBidName(level)), nil)
	b.advanceBid()
	return nil
}

// PassBid は宣言を見送る。
func (b *Boston) PassBid(player int) error {
	if err := b.checkBidTurn(player); err != nil {
		return err
	}
	b.bids = append(b.bids, &BostonBidRecord{Player: player, Level: BostonBidPass})
	b.passCount++
	b.addLog(player, "pass", "passes", nil)
	b.advanceBid()
	return nil
}

// advanceBid は次の宣言手番へ進め、決着していれば契約へ移る。
func (b *Boston) advanceBid() {
	if b.highBid != nil && b.passCount >= BostonPlayerCnt-1 {
		b.settleBid()
		return
	}
	if b.highBid == nil && b.passCount >= BostonPlayerCnt {
		// **全員パスなら配り直し。**
		b.addLog(-1, "redeal", "everybody passed", nil)
		b.handNumber--
		b.beginHand()
		return
	}
	b.bidIdx = (b.bidIdx + 1) % BostonPlayerCnt
}

// settleBid は落札を確定する。
func (b *Boston) settleBid() {
	b.declarerIdx = b.highBid.Player
	b.trumpSuit = b.highBid.Suit
	b.exposed = BostonBidIsExposed(b.highBid.Level)
	b.addLog(b.declarerIdx, "contract", fmt.Sprintf("takes %s", BostonBidName(b.highBid.Level)), nil)

	// **パートナーを指名できるのはトリック数の宣言だけ。**
	if BostonBidCanCallPartner(b.highBid.Level) {
		b.phase = BostonPhaseCallPartner
		b.currentIdx = b.declarerIdx
		return
	}
	b.beginPlay()
}

// CallPartner は落札者がパートナーを指名する。partner が -1 なら単独で戦う。
func (b *Boston) CallPartner(player, partner int) error {
	if b.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if b.phase != BostonPhaseCallPartner {
		return fmt.Errorf("it is not time to call a partner")
	}
	if player != b.declarerIdx {
		return fmt.Errorf("only the declarer calls a partner")
	}
	if partner == player {
		return fmt.Errorf("the declarer cannot partner itself")
	}
	if partner != -1 && (partner < 0 || partner >= BostonPlayerCnt) {
		return fmt.Errorf("bad partner index: %d", partner)
	}
	b.partnerIdx = partner
	if partner >= 0 {
		b.addLog(player, "call_partner", fmt.Sprintf("calls player %d", partner), nil)
	} else {
		b.addLog(player, "go_alone", "plays alone against three", nil)
	}
	b.beginPlay()
	return nil
}

// beginPlay はプレイフェーズに入る。
func (b *Boston) beginPlay() {
	b.phase = BostonPhasePlay
	// **リードはディーラーの左隣。**落札者ではない。
	b.trickLeader = (b.dealerIdx + 1) % BostonPlayerCnt
	b.currentIdx = b.trickLeader
}

// BostonIsDeclarerSide は席が落札側かを返す。
func (b *Boston) BostonIsDeclarerSide(seat int) bool {
	return seat == b.declarerIdx || (b.partnerIdx >= 0 && seat == b.partnerIdx)
}

// BostonValidPlays は player が出せる手札インデックスを返す。
//
// 追随のみ強制。切札の強制はない。
func (b *Boston) BostonValidPlays(player int) []int {
	p := b.GetPlayer(player)
	if p == nil {
		return nil
	}
	all := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		all = append(all, i)
	}
	if len(b.trick) == 0 || b.trick[0] == nil {
		return all
	}
	leadSuit := b.trick[0].GetDesign()
	same := make([]int, 0, len(all))
	for _, i := range all {
		if c := p.GetCard(i); c != nil && c.GetDesign() == leadSuit {
			same = append(same, i)
		}
	}
	if len(same) > 0 {
		return same
	}
	return all
}

// PlayCard は 1 枚出す。
func (b *Boston) PlayCard(player, idx int) error {
	if b.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if b.phase != BostonPhasePlay {
		return fmt.Errorf("the play phase is not in progress")
	}
	if player != b.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	p := b.GetPlayer(player)
	if p == nil || idx < 0 || idx >= p.GetCardsSize() {
		return fmt.Errorf("bad card index: %d", idx)
	}
	if !bostonContains(b.BostonValidPlays(player), idx) {
		return fmt.Errorf("that card may not be played")
	}

	card := p.GetCard(idx)
	p.RemoveCard(idx)
	b.trick = append(b.trick, card)
	b.addLog(player, "play", "plays a card", []*Card{card})

	if len(b.trick) < BostonPlayerCnt {
		b.currentIdx = (player + 1) % BostonPlayerCnt
		return nil
	}
	b.resolveTrick()
	return nil
}

// bostonCardRank は札の強さを返す (A が最強)。
func bostonCardRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// resolveTrick はトリックを解決する。
func (b *Boston) resolveTrick() {
	lead := b.trick[0]
	leadSuit := lead.GetDesign()
	bestOffset := 0
	bestIsTrump := b.trumpSuit != 0 && leadSuit == b.trumpSuit
	bestRank := bostonCardRank(lead)
	for i := 1; i < len(b.trick); i++ {
		c := b.trick[i]
		if c == nil {
			continue
		}
		isTrump := b.trumpSuit != 0 && c.GetDesign() == b.trumpSuit
		rank := bostonCardRank(c)
		switch {
		case isTrump && !bestIsTrump:
			bestOffset, bestIsTrump, bestRank = i, true, rank
		case isTrump == bestIsTrump && c.GetDesign() == b.trick[bestOffset].GetDesign() && rank > bestRank:
			bestOffset, bestRank = i, rank
		}
	}
	winner := (b.trickLeader + bestOffset) % BostonPlayerCnt

	b.tricksWon[winner]++
	b.trickNumber++
	b.trick = nil
	b.trickLeader = winner
	b.currentIdx = winner
	b.addLog(winner, "trick", fmt.Sprintf("takes trick %d", b.trickNumber), nil)

	if b.trickNumber >= BostonHandSize {
		b.finishHand()
	}
}

// BostonDeclarerTricks は落札側が取ったトリック数を返す。
//
// **落札前は 0。**declarerIdx は競りが決まるまで -1 なので、素で添字にすると
// 落ちる。プレゼンターは宣言フェーズでもこれを呼ぶ。
func (b *Boston) BostonDeclarerTricks() int {
	if b.declarerIdx < 0 || b.declarerIdx >= BostonPlayerCnt {
		return 0
	}
	total := b.tricksWon[b.declarerIdx]
	if b.partnerIdx >= 0 && b.partnerIdx < BostonPlayerCnt {
		total += b.tricksWon[b.partnerIdx]
	}
	return total
}

// finishHand は局を精算する。
//
// **達成なら各相手から受け取り、失敗なら各相手に払う。**プールではなく席間の
// やり取りで表す。
func (b *Boston) finishHand() {
	level := b.highBid.Level
	won := b.BostonDeclarerTricks()
	b.bidMade = BostonBidSucceeded(level, won)
	pay := BostonBidPayout(level)

	for i := range BostonPlayerCnt {
		if b.BostonIsDeclarerSide(i) {
			continue
		}
		if b.bidMade {
			b.chips[i] -= pay
			b.chips[b.declarerIdx] += pay
		} else {
			b.chips[i] += pay
			b.chips[b.declarerIdx] -= pay
		}
	}

	if b.bidMade {
		b.addLog(b.declarerIdx, "hand_end", fmt.Sprintf("makes %s with %d tricks", BostonBidName(level), won), nil)
	} else {
		b.addLog(b.declarerIdx, "hand_end", fmt.Sprintf("fails %s with %d tricks", BostonBidName(level), won), nil)
	}

	b.phase = BostonPhaseHandEnd
	b.checkGameEnd()
}

// checkGameEnd は規定局数に達していれば決着させる。
func (b *Boston) checkGameEnd() {
	if b.handNumber < b.targetHands {
		return
	}
	best, bestIdx := -1<<30, -1
	for i := range BostonPlayerCnt {
		if b.chips[i] > best {
			best, bestIdx = b.chips[i], i
		}
	}
	b.winnerIdx = bestIdx
	b.gameEndFlag = true
	b.phase = BostonPhaseGameEnd
	b.addLog(bestIdx, "game_end", "finishes with the most chips", nil)
}

// NextHand は次の局を配る。
func (b *Boston) NextHand() error {
	if b.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if b.phase != BostonPhaseHandEnd {
		return fmt.Errorf("the hand is still in progress")
	}
	b.dealerIdx = (b.dealerIdx + 1) % BostonPlayerCnt
	b.beginHand()
	return nil
}

// bostonContains は s に v が含まれるかを返す。
func bostonContains(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// ---- CPU ----

// BostonCpuBid は CPU の宣言を決める。
//
// 一番長いスートの枚数と絵札から見込みトリック数を粗く見積もる。
func (b *Boston) BostonCpuBid(idx int) (BostonBidLevel, int) {
	p := b.GetPlayer(idx)
	if p == nil {
		return BostonBidPass, 0
	}
	counts := map[int]int{}
	honours := map[int]int{}
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c == nil {
			continue
		}
		counts[c.GetDesign()]++
		if v := c.GetValue(); v == 1 || v >= 12 {
			honours[c.GetDesign()]++
		}
	}
	bestSuit, best := CardDesignSpade, -1
	for s := CardDesignSpade; s <= CardDesignDiamond; s++ {
		if v := counts[s] + honours[s]*2; v > best {
			bestSuit, best = s, v
		}
	}
	estimate := counts[bestSuit] + honours[bestSuit]

	// 見込みトリック数に対応する一番低い宣言を選ぶ。
	level := BostonBidPass
	for l := BostonBidFive; l < BostonBidLevelCount; l++ {
		if BostonBidKindOf(l) != BostonKindTricks {
			continue
		}
		if BostonBidTricks(l) <= estimate {
			level = l
		}
	}
	if level == BostonBidPass {
		return BostonBidPass, 0
	}
	// 立っている宣言を上回れないならパス。
	if b.highBid != nil && level <= b.highBid.Level {
		return BostonBidPass, 0
	}
	return level, bestSuit
}

// BostonCpuPlay は CPU が出す手札インデックスを返す。
func (b *Boston) BostonCpuPlay(idx int) int {
	valid := b.BostonValidPlays(idx)
	if len(valid) == 0 {
		return -1
	}
	p := b.GetPlayer(idx)
	if p == nil {
		return valid[0]
	}
	// **ミゼール側は取らないことが目的。**常に一番安い札を出す。
	misere := b.highBid != nil && BostonBidKindOf(b.highBid.Level) == BostonKindMisere
	if misere && b.BostonIsDeclarerSide(idx) {
		return bostonLowest(p, valid)
	}
	if len(b.trick) == 0 {
		return bostonHighest(p, valid)
	}
	lead := b.trick[0]
	winning, winRank := -1, -1
	for _, i := range valid {
		c := p.GetCard(i)
		if bostonBeats(c, lead, b.trumpSuit) && bostonCardRank(c) > winRank {
			winning, winRank = i, bostonCardRank(c)
		}
	}
	if winning >= 0 {
		return winning
	}
	return bostonLowest(p, valid)
}

// bostonLowest は valid のうち一番弱い札の索引を返す。
func bostonLowest(p *BostonPlayer, valid []int) int {
	best, bestRank := valid[0], 1<<30
	for _, i := range valid {
		if r := bostonCardRank(p.GetCard(i)); r < bestRank {
			best, bestRank = i, r
		}
	}
	return best
}

// bostonHighest は valid のうち一番強い札の索引を返す。
func bostonHighest(p *BostonPlayer, valid []int) int {
	best, bestRank := valid[0], -1
	for _, i := range valid {
		if r := bostonCardRank(p.GetCard(i)); r > bestRank {
			best, bestRank = i, r
		}
	}
	return best
}

// bostonBeats は c が lead に勝つかを返す。
func bostonBeats(c, lead *Card, trumpSuit int) bool {
	if c == nil || lead == nil {
		return false
	}
	if trumpSuit != 0 && c.GetDesign() == trumpSuit && lead.GetDesign() != trumpSuit {
		return true
	}
	if c.GetDesign() != lead.GetDesign() {
		return false
	}
	return bostonCardRank(c) > bostonCardRank(lead)
}

// IsHumanTurn は今が人間の手番かを返す。
func (b *Boston) IsHumanTurn() bool {
	if b.gameEndFlag {
		return false
	}
	switch b.phase {
	case BostonPhaseBid:
		p := b.GetPlayer(b.bidIdx)
		return p != nil && p.GetIsHuman()
	case BostonPhaseCallPartner, BostonPhasePlay:
		p := b.GetPlayer(b.currentIdx)
		return p != nil && p.GetIsHuman()
	}
	return false
}

// CpuPlay は今の手番の CPU に 1 手打たせる。
func (b *Boston) CpuPlay() {
	if b.gameEndFlag {
		return
	}
	switch b.phase {
	case BostonPhaseBid:
		idx := b.bidIdx
		if p := b.GetPlayer(idx); p == nil || p.GetIsHuman() {
			return
		}
		level, suit := b.BostonCpuBid(idx)
		if level == BostonBidPass || b.Bid(idx, level, suit) != nil {
			_ = b.PassBid(idx)
		}
	case BostonPhaseCallPartner:
		idx := b.declarerIdx
		if p := b.GetPlayer(idx); p == nil || p.GetIsHuman() {
			return
		}
		// CPU は単独で戦う。指名の駆け引きは v1 では扱わない。
		_ = b.CallPartner(idx, -1)
	case BostonPhasePlay:
		idx := b.currentIdx
		if p := b.GetPlayer(idx); p == nil || p.GetIsHuman() {
			return
		}
		if i := b.BostonCpuPlay(idx); i >= 0 {
			_ = b.PlayCard(idx, i)
		}
	}
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (b *Boston) GetPlayers() []*BostonPlayer { return b.players }

// GetPlayer は idx のプレイヤーを返す。
func (b *Boston) GetPlayer(idx int) *BostonPlayer {
	return getPlayer(b.players, idx)
}

// GetPhase は現在のフェーズを返す。
func (b *Boston) GetPhase() BostonPhase { return b.phase }

// GetCurrentPlayerIdx は現在の手番を返す。
func (b *Boston) GetCurrentPlayerIdx() int { return b.currentIdx }

// GetBidPlayerIdx は宣言中の手番を返す。
func (b *Boston) GetBidPlayerIdx() int { return b.bidIdx }

// GetDealerIdx はディーラーを返す。
func (b *Boston) GetDealerIdx() int { return b.dealerIdx }

// GetBids はこの局の宣言履歴を返す。
func (b *Boston) GetBids() []*BostonBidRecord { return b.bids }

// GetHighBid は現在の最高宣言を返す (未落札なら nil)。
func (b *Boston) GetHighBid() *BostonBidRecord { return b.highBid }

// GetDeclarerIdx は落札者を返す (-1 なら未確定)。
func (b *Boston) GetDeclarerIdx() int { return b.declarerIdx }

// GetPartnerIdx は指名されたパートナーを返す (-1 なら単独)。
func (b *Boston) GetPartnerIdx() int { return b.partnerIdx }

// GetTrumpSuit は切札を返す (0 なら切札なし)。
func (b *Boston) GetTrumpSuit() int { return b.trumpSuit }

// IsExposed は落札者の手札を公開しているかを返す。
func (b *Boston) IsExposed() bool { return b.exposed }

// GetTrick は場に出ている札を返す。
func (b *Boston) GetTrick() []*Card { return b.trick }

// GetTrickLeaderIdx はこのトリックのリード席を返す。
func (b *Boston) GetTrickLeaderIdx() int { return b.trickLeader }

// GetTrickNumber は済んだトリック数を返す。
func (b *Boston) GetTrickNumber() int { return b.trickNumber }

// GetTricksWon は席が取ったトリック数を返す。
func (b *Boston) GetTricksWon(idx int) int {
	if idx < 0 || idx >= BostonPlayerCnt {
		return 0
	}
	return b.tricksWon[idx]
}

// IsBidMade は直前の局で落札側が達成したかを返す。
func (b *Boston) IsBidMade() bool { return b.bidMade }

// GetChips は席の通算を返す。
func (b *Boston) GetChips(idx int) int {
	if idx < 0 || idx >= BostonPlayerCnt {
		return 0
	}
	return b.chips[idx]
}

// GetHandNumber は現在の局番号を返す。
func (b *Boston) GetHandNumber() int { return b.handNumber }

// GetTargetHands は規定局数を返す。
func (b *Boston) GetTargetHands() int { return b.targetHands }

// GetGameEndFlag はゲーム終了フラグを返す。
func (b *Boston) GetGameEndFlag() bool { return b.gameEndFlag }

// GetWinnerIdx は勝者を返す (-1 なら未決)。
func (b *Boston) GetWinnerIdx() int { return b.winnerIdx }

// GetConfig は設定を返す。
func (b *Boston) GetConfig() BostonConfig { return b.config }

// SetConfig は設定をセットする。
func (b *Boston) SetConfig(c BostonConfig) { b.config = c }

// addLog は棋譜を 1 行足す。
func (b *Boston) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	b.appendLog(playerIdx, actionType, detail, cards)
}

// SetPhaseForTest はテスト用にフェーズを設定する。
func (b *Boston) SetPhaseForTest(p BostonPhase) { b.phase = p }

// SetHandForTest はテスト用に手札を差し替える。
func (b *Boston) SetHandForTest(idx int, cards []*Card) {
	setHandForTest(b.GetPlayer(idx), cards)
}

// SetContractForTest はテスト用に契約を設定する。
func (b *Boston) SetContractForTest(declarer, partner int, level BostonBidLevel, suit int) {
	b.declarerIdx = declarer
	b.partnerIdx = partner
	b.highBid = &BostonBidRecord{Player: declarer, Level: level, Suit: suit}
	b.trumpSuit = suit
	b.exposed = BostonBidIsExposed(level)
}

// SetCurrentPlayerForTest はテスト用に手番を設定する。
func (b *Boston) SetCurrentPlayerForTest(idx int) { b.currentIdx = idx }

// SetTrickLeaderForTest はテスト用にリード席を設定する。
func (b *Boston) SetTrickLeaderForTest(idx int) { b.trickLeader = idx }

// SetTricksWonForTest はテスト用に取得トリック数を設定する。
func (b *Boston) SetTricksWonForTest(idx, n int) {
	if idx >= 0 && idx < BostonPlayerCnt {
		b.tricksWon[idx] = n
	}
}

// SetHandNumberForTest はテスト用に局番号を設定する。
func (b *Boston) SetHandNumberForTest(n int) { b.handNumber = n }

// FinishHandForTest はテスト用に精算を走らせる。
func (b *Boston) FinishHandForTest() { b.finishHand() }

// bostonJSON is the JSON wire format for Boston.
type bostonJSON struct {
	Players     []*BostonPlayer      `json:"pl"`
	Config      BostonConfig         `json:"cf"`
	Phase       BostonPhase          `json:"ph"`
	DealerIdx   int                  `json:"di"`
	CurrentIdx  int                  `json:"ci"`
	BidIdx      int                  `json:"bi"`
	Bids        []*BostonBidRecord   `json:"bd"`
	PassCount   int                  `json:"pc"`
	HighBid     *BostonBidRecord     `json:"hb"`
	DeclarerIdx int                  `json:"de"`
	PartnerIdx  int                  `json:"pa"`
	TrumpSuit   int                  `json:"ts"`
	Exposed     bool                 `json:"ex"`
	Trick       []*Card              `json:"tk"`
	TrickLeader int                  `json:"tl"`
	TrickNumber int                  `json:"tn"`
	TricksWon   [BostonPlayerCnt]int `json:"tw"`
	BidMade     bool                 `json:"bm"`
	Chips       [BostonPlayerCnt]int `json:"ch"`
	HandNumber  int                  `json:"hn"`
	TargetHands int                  `json:"th"`
	GameEndFlag bool                 `json:"ge"`
	WinnerIdx   int                  `json:"wi"`
	ActionLog   []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (b *Boston) MarshalJSON() ([]byte, error) {
	return json.Marshal(bostonJSON{
		Players: b.players, Config: b.config, Phase: b.phase,
		DealerIdx: b.dealerIdx, CurrentIdx: b.currentIdx, BidIdx: b.bidIdx,
		Bids: b.bids, PassCount: b.passCount, HighBid: b.highBid,
		DeclarerIdx: b.declarerIdx, PartnerIdx: b.partnerIdx,
		TrumpSuit: b.trumpSuit, Exposed: b.exposed,
		Trick: b.trick, TrickLeader: b.trickLeader, TrickNumber: b.trickNumber,
		TricksWon: b.tricksWon, BidMade: b.bidMade, Chips: b.chips,
		HandNumber: b.handNumber, TargetHands: b.targetHands,
		GameEndFlag: b.gameEndFlag, WinnerIdx: b.winnerIdx, ActionLog: b.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **復元でしか入らない値を弾く。**KV から戻ってきた壊れた状態でプレイが
// 詰まないよう、席番号・宣言序列・スートを検証する。
func (b *Boston) UnmarshalJSON(data []byte) error {
	var j bostonJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != BostonPlayerCnt {
		return fmt.Errorf("bad player count: %d", len(j.Players))
	}
	if j.Phase < BostonPhaseBid || j.Phase > BostonPhaseGameEnd {
		return fmt.Errorf("bad phase: %d", j.Phase)
	}
	for name, v := range map[string]int{"dealer": j.DealerIdx, "current": j.CurrentIdx, "bid": j.BidIdx} {
		if v < 0 || v >= BostonPlayerCnt {
			return fmt.Errorf("bad %s index: %d", name, v)
		}
	}
	for name, v := range map[string]int{
		"declarer": j.DeclarerIdx, "partner": j.PartnerIdx,
		"trick leader": j.TrickLeader, "winner": j.WinnerIdx,
	} {
		if v < -1 || v >= BostonPlayerCnt {
			return fmt.Errorf("bad %s index: %d", name, v)
		}
	}
	if j.TrumpSuit != 0 && (j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return fmt.Errorf("bad trump suit: %d", j.TrumpSuit)
	}
	if len(j.Trick) > BostonPlayerCnt {
		return fmt.Errorf("bad trick size: %d", len(j.Trick))
	}
	if j.HighBid != nil && (j.HighBid.Level < BostonBidPass || j.HighBid.Level >= BostonBidLevelCount) {
		return fmt.Errorf("bad high bid level: %d", j.HighBid.Level)
	}
	// **落札者とパートナーが同じ席では 1 対 3 も 2 対 2 も成り立たない。**
	if j.DeclarerIdx >= 0 && j.PartnerIdx == j.DeclarerIdx {
		return fmt.Errorf("the declarer cannot partner itself")
	}

	b.players = j.Players
	b.config = j.Config
	b.phase = j.Phase
	b.dealerIdx = j.DealerIdx
	b.currentIdx = j.CurrentIdx
	b.bidIdx = j.BidIdx
	b.bids = j.Bids
	b.passCount = j.PassCount
	b.highBid = j.HighBid
	b.declarerIdx = j.DeclarerIdx
	b.partnerIdx = j.PartnerIdx
	b.trumpSuit = j.TrumpSuit
	b.exposed = j.Exposed
	b.trick = j.Trick
	b.trickLeader = j.TrickLeader
	b.trickNumber = j.TrickNumber
	b.tricksWon = j.TricksWon
	b.bidMade = j.BidMade
	b.chips = j.Chips
	b.handNumber = j.HandNumber
	b.targetHands = j.TargetHands
	if b.targetHands <= 0 {
		b.targetHands = DefaultBostonConfig().TargetHands
	}
	b.gameEndFlag = j.GameEndFlag
	b.winnerIdx = j.WinnerIdx
	b.actionLog = j.ActionLog
	return nil
}
