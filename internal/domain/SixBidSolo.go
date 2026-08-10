//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"
)

// SixBidSoloPlayerCnt はシックスビッド・ソロの人数。
const SixBidSoloPlayerCnt = 3

// SixBidSoloHandSize は 1 人あたりの手札枚数。
//
// **11 枚であって 12 枚ではない。**配り方は 4 → 3 → ウィドウ 3 → 4 で、
// 11 × 3 + 3 = 36 とちょうど合う。12 枚ずつだとウィドウの 3 枚が入らない。
const SixBidSoloHandSize = 11

// SixBidSoloWidowSize はウィドウ (伏せ札) の枚数。
//
// **ウィドウは飾りではない。**最後に宣言者の得点へ加算される (ミゼール系を除く)。
const SixBidSoloWidowSize = 3

// SixBidSoloDeckSize は使用する札数 (A-10-K-Q-J-9-8-7-6 の 36 枚)。
const SixBidSoloDeckSize = 36

// SixBidSoloTotalPoints は場に出るカード点の総和。
//
// 1 スートあたり A11 + 10:10 + K4 + Q3 + J2 = 30、× 4 スートで 120。
const SixBidSoloTotalPoints = 120

// SixBidSoloBaseTarget は通常ビッドの基準点。
//
// **勝利には「60 点ちょうど」ではなく「60 点を超えること」が要る。**
// 精算額もこの値との差で決まる。
const SixBidSoloBaseTarget = 60

// ギャランティー・ソロの目標点。**スートによって違う。**
const (
	// SixBidSoloGuaranteeHeart ♥ 切札のときの目標
	SixBidSoloGuaranteeHeart = 74
	// SixBidSoloGuaranteeOther 他スート切札のときの目標
	SixBidSoloGuaranteeOther = 80
)

// SixBidSoloTricks は 1 局のトリック数 (手札枚数と同じ)。
const SixBidSoloTricks = SixBidSoloHandSize

// SixBidSoloBidKind は 6 段階のビッド。
type SixBidSoloBidKind int

// シックスビッド・ソロのビッド定数 (低い順)。
const (
	// SixBidSoloBidPass パス
	SixBidSoloBidPass SixBidSoloBidKind = iota
	// SixBidSoloBidSolo シンプル・ソロ (♠♣♦ 切札、61 点以上)
	SixBidSoloBidSolo
	// SixBidSoloBidHeartSolo ハート・ソロ (♥ 切札、61 点以上)
	SixBidSoloBidHeartSolo
	// SixBidSoloBidMisere ミゼール (切札なし、**カード点 0**)
	SixBidSoloBidMisere
	// SixBidSoloBidGuarantee ギャランティー・ソロ (♥74 / 他 80 点以上)
	SixBidSoloBidGuarantee
	// SixBidSoloBidSpreadMisere スプレッド・ミゼール (手札公開のミゼール)
	SixBidSoloBidSpreadMisere
	// SixBidSoloBidCall コール・ソロ (全 120 点 + 札の指名交換)
	SixBidSoloBidCall
	// SixBidSoloBidCount ビッド種別の総数
	SixBidSoloBidCount
)

// SixBidSoloMinBid は最低ビッド (パスの 1 つ上)。
const SixBidSoloMinBid = SixBidSoloBidSolo

// SixBidSoloMaxBid は最高ビッド。
const SixBidSoloMaxBid = SixBidSoloBidCall

// 固定額のビッド (通常ビッドは差分 × 倍率なのでここには無い)。
const (
	// SixBidSoloMisereValue ミゼールの額
	//
	// **pagat の 30 を採る。**Wikipedia は 40 とするが、それだと
	// ギャランティー (40) と並んでしまい、序列が額の上で崩れる。
	SixBidSoloMisereValue = 30
	// SixBidSoloGuaranteeValue ギャランティー・ソロの額
	SixBidSoloGuaranteeValue = 40
	// SixBidSoloSpreadValue スプレッド・ミゼールの額
	SixBidSoloSpreadValue = 60
	// SixBidSoloCallValue コール・ソロの額 (♥ 以外)
	SixBidSoloCallValue = 100
	// SixBidSoloCallHeartValue コール・ソロの額 (♥ 切札)
	SixBidSoloCallHeartValue = 150
)

// 通常ビッドの倍率。**60 点との差に掛ける。**
const (
	// SixBidSoloSoloMultiplier シンプル・ソロの倍率
	SixBidSoloSoloMultiplier = 2
	// SixBidSoloHeartMultiplier ハート・ソロの倍率
	SixBidSoloHeartMultiplier = 3
)

// SixBidSoloIsMisere はビッドがミゼール系 (カード点 0 が目標) かを返す。
func SixBidSoloIsMisere(b SixBidSoloBidKind) bool {
	return b == SixBidSoloBidMisere || b == SixBidSoloBidSpreadMisere
}

// SixBidSoloUsesWidow は宣言者がウィドウを得点に加算できるかを返す。
//
// **ミゼール系だけは加算しない。**取らないことが目的なので、伏せ札の点まで
// 背負わされては成立しない。
func SixBidSoloUsesWidow(b SixBidSoloBidKind) bool {
	return b != SixBidSoloBidPass && !SixBidSoloIsMisere(b)
}

// SixBidSoloNeedsTrump は宣言者が切札を指定するビッドかを返す。
func SixBidSoloNeedsTrump(b SixBidSoloBidKind) bool {
	switch b {
	case SixBidSoloBidSolo, SixBidSoloBidGuarantee, SixBidSoloBidCall:
		return true
	}
	// **ハート・ソロは切札が ♥ で固定。**選ぶ余地が無い。
	return false
}

// SixBidSoloFixedTrump はビッドに固定された切札を返す (無ければ 0)。
func SixBidSoloFixedTrump(b SixBidSoloBidKind) int {
	if b == SixBidSoloBidHeartSolo {
		return CardDesignHeart
	}
	return 0
}

// SixBidSoloTargetPoints はビッドの目標カード点を返す。
//
// 通常ビッドは **61 点以上** (60 を超えること)、ギャランティーはスート依存、
// コールは全点、ミゼール系は 0。
func SixBidSoloTargetPoints(b SixBidSoloBidKind, trumpSuit int) int {
	switch b {
	case SixBidSoloBidSolo, SixBidSoloBidHeartSolo:
		return SixBidSoloBaseTarget + 1
	case SixBidSoloBidGuarantee:
		if trumpSuit == CardDesignHeart {
			return SixBidSoloGuaranteeHeart
		}
		return SixBidSoloGuaranteeOther
	case SixBidSoloBidCall:
		return SixBidSoloTotalPoints
	case SixBidSoloBidMisere, SixBidSoloBidSpreadMisere:
		return 0
	}
	return 0
}

// SixBidSoloCardPoint は 1 枚のカード点を返す。
//
// **A=11 / 10=10 / K=4 / Q=3 / J=2、9 以下は 0。**合計 120 点。
func SixBidSoloCardPoint(c *Card) int {
	if c == nil {
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

// sixBidSoloRank は札の平の強さを返す (A-10-K-Q-J-9-8-7-6)。
//
// **10 が K より強い。**スカート系の序列。
func sixBidSoloRank(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 1:
		return 9
	case 10:
		return 8
	case 13:
		return 7
	case 12:
		return 6
	case 11:
		return 5
	case 9:
		return 4
	case 8:
		return 3
	case 7:
		return 2
	case 6:
		return 1
	}
	return 0
}

// SixBidSoloPhase はゲームフェーズ。
type SixBidSoloPhase int

// シックスビッド・ソロのフェーズ定数
const (
	// SixBidSoloPhaseBid ビッド
	SixBidSoloPhaseBid SixBidSoloPhase = iota
	// SixBidSoloPhaseDeclare 落札者が切札 (と、コールなら指名札) を決める
	SixBidSoloPhaseDeclare
	// SixBidSoloPhasePlay トリックプレイ
	SixBidSoloPhasePlay
	// SixBidSoloPhaseHandEnd 局終了 (精算済み)
	SixBidSoloPhaseHandEnd
	// SixBidSoloPhaseGameEnd ゲーム終了
	SixBidSoloPhaseGameEnd
)

// SixBidSoloBid は 1 件の宣言。
type SixBidSoloBid struct {
	Player int
	Kind   SixBidSoloBidKind
}

// SixBidSoloHandResult は 1 局の精算結果。
type SixBidSoloHandResult struct {
	// Kind は落札したビッド。
	Kind SixBidSoloBidKind
	// Declarer は宣言者の席。
	Declarer int
	// DeclarerPoints は宣言者が取ったカード点 (ウィドウ込み)。
	DeclarerPoints int
	// WidowPoints はウィドウのカード点 (ミゼール系では加算されない)。
	WidowPoints int
	// Target は達成に要った点。
	Target int
	// Made は宣言者が達成したか。
	Made bool
	// Value は 1 人あたりの受け払い額。
	Value int
	// Deltas は各席の増減。
	Deltas [SixBidSoloPlayerCnt]int
}

// SixBidSolo はシックスビッド・ソロのゲームクラス。
//
// ドイツ系移民由来のアメリカ産スカート系 3 人ゲーム。**36 枚を 11 枚ずつ配り、
// 3 枚をウィドウとして伏せる。**6 段階のビッドが目標点と倍率を同時に切り替える。
type SixBidSolo struct {
	players   []*SixBidSoloPlayer
	config    SixBidSoloConfig
	phase     SixBidSoloPhase
	dealerIdx int
	// bidIdx は宣言中の席。
	bidIdx int
	// currentIdx はプレイ中の手番。
	currentIdx int
	bids       []*SixBidSoloBid
	highBid    *SixBidSoloBid
	// declarerIdx は落札者 (未定なら -1)。
	declarerIdx int
	// passed は既にパスした席。
	passed [SixBidSoloPlayerCnt]bool
	// widow は伏せ札。
	widow []*Card
	// trumpSuit は切札 (ミゼール系は 0)。
	trumpSuit int
	declared  bool
	// calledCard はコール・ソロで指名された札 (無ければ nil)。
	calledCard *Card
	// spreadOpen はスプレッド・ミゼールで手札を公開済みか。
	spreadOpen  bool
	trick       []*Card
	trickLeader int
	trickNumber int
	// points は各席が取ったカード点。
	points [SixBidSoloPlayerCnt]int
	// tricksWon は各席が取ったトリック数。
	tricksWon [SixBidSoloPlayerCnt]int
	// scores は通算得点。
	scores      [SixBidSoloPlayerCnt]int
	lastResult  *SixBidSoloHandResult
	handNumber  int
	gameEndFlag bool
	winnerIdx   int
	actionLogBase
}

// NewSixBidSolo コンストラクタ
func NewSixBidSolo(players []*SixBidSoloPlayer, config SixBidSoloConfig) *SixBidSolo {
	return &SixBidSolo{players: players, config: config, declarerIdx: -1, winnerIdx: -1}
}

// NewDefaultSixBidSolo は人間 1 人 + CPU 2 体の卓を作る。
func NewDefaultSixBidSolo() *SixBidSolo {
	players := make([]*SixBidSoloPlayer, 0, SixBidSoloPlayerCnt)
	for i := range SixBidSoloPlayerCnt {
		players = append(players, NewSixBidSoloPlayer(i == 0))
	}
	return NewSixBidSolo(players, DefaultSixBidSoloConfig())
}

// Reset はゲームを初期化する。
func (s *SixBidSolo) Reset() {
	for i := range SixBidSoloPlayerCnt {
		s.scores[i] = 0
	}
	s.handNumber = 0
	s.gameEndFlag = false
	s.winnerIdx = -1
	s.lastResult = nil
	s.dealerIdx = 0
	s.actionLog = make([]*ActionLogEntry, 0)
	s.beginHand()
}

// beginHand は 1 局を配る。
func (s *SixBidSolo) beginHand() {
	s.handNumber++
	s.phase = SixBidSoloPhaseBid
	s.bids = make([]*SixBidSoloBid, 0)
	s.highBid = nil
	s.declarerIdx = -1
	s.trumpSuit = 0
	s.declared = false
	s.calledCard = nil
	s.spreadOpen = false
	s.trick = make([]*Card, 0, SixBidSoloPlayerCnt)
	s.trickNumber = 0
	for i := range SixBidSoloPlayerCnt {
		s.points[i] = 0
		s.tricksWon[i] = 0
		s.passed[i] = false
		if p := s.GetPlayer(i); p != nil {
			p.ResetRound()
		}
	}
	s.dealRound()
	// **宣言も最初のリードも親の左から。**
	s.bidIdx = (s.dealerIdx + 1) % SixBidSoloPlayerCnt
	s.currentIdx = s.bidIdx
	s.trickLeader = s.bidIdx
	s.addLog(-1, "deal", "hand "+strconv.Itoa(s.handNumber), nil)
}

// dealRound は 4 → 3 → ウィドウ 3 → 4 の順に配る。
//
// **ウィドウは途中で抜く。**まとめて配ってから 3 枚残すのと枚数は同じだが、
// 配り方そのものがこのゲームの決まりなので順序どおりに行う。
func (s *SixBidSolo) dealRound() {
	deck := newSixBidSoloDeck()
	sixBidSoloShuffle(deck)
	pos := 0
	deal := func(n int) {
		for i := range SixBidSoloPlayerCnt {
			seat := (s.dealerIdx + 1 + i) % SixBidSoloPlayerCnt
			p := s.GetPlayer(seat)
			for range n {
				if pos < len(deck) && p != nil {
					p.AddCard(deck[pos])
					pos++
				}
			}
		}
	}
	deal(4)
	deal(3)
	s.widow = make([]*Card, 0, SixBidSoloWidowSize)
	for range SixBidSoloWidowSize {
		if pos < len(deck) {
			s.widow = append(s.widow, deck[pos])
			pos++
		}
	}
	deal(4)
}

// newSixBidSoloDeck は A-10-K-Q-J-9-8-7-6 の 36 枚を作る。
func newSixBidSoloDeck() []*Card {
	values := []int{1, 10, 13, 12, 11, 9, 8, 7, 6}
	cards := make([]*Card, 0, SixBidSoloDeckSize)
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		for _, v := range values {
			cards = append(cards, NewCard(suit, v, true))
		}
	}
	return cards
}

// sixBidSoloShuffle は札をシャッフルする。
func sixBidSoloShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1) //nolint:gosec // ゲームのシャッフルに暗号強度は要らない
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// ---- ビッド ----

// checkBidTurn は宣言できる状態かを検査する。
func (s *SixBidSolo) checkBidTurn(player int) error {
	if s.gameEndFlag {
		return errors.New("the game is over")
	}
	if s.phase != SixBidSoloPhaseBid {
		return errors.New("it is not the bidding phase")
	}
	if player != s.bidIdx {
		return errors.New("it is not your turn to bid")
	}
	return nil
}

// SixBidSoloCanBid は指定のビッドを出せるかを返す。
//
// **上回る宣言だけが通る。**同額で奪える席は無い。
func (s *SixBidSolo) SixBidSoloCanBid(player int, kind SixBidSoloBidKind) bool {
	if kind < SixBidSoloMinBid || kind > SixBidSoloMaxBid {
		return false
	}
	if player < 0 || player >= SixBidSoloPlayerCnt || s.passed[player] {
		return false
	}
	if s.highBid == nil {
		return true
	}
	return kind > s.highBid.Kind
}

// Bid は宣言する。
func (s *SixBidSolo) Bid(player int, kind SixBidSoloBidKind) error {
	if err := s.checkBidTurn(player); err != nil {
		return err
	}
	if !s.SixBidSoloCanBid(player, kind) {
		return errors.New("a bid must beat the standing one")
	}
	b := &SixBidSoloBid{Player: player, Kind: kind}
	s.bids = append(s.bids, b)
	s.highBid = b
	s.addLog(player, "bid", sixBidSoloBidName(kind), nil)
	s.advanceBid()
	return nil
}

// PassBid は宣言を見送る。
func (s *SixBidSolo) PassBid(player int) error {
	if err := s.checkBidTurn(player); err != nil {
		return err
	}
	s.passed[player] = true
	s.bids = append(s.bids, &SixBidSoloBid{Player: player, Kind: SixBidSoloBidPass})
	s.addLog(player, "pass", "", nil)
	s.advanceBid()
	return nil
}

// advanceBid は次の宣言者へ進める。
func (s *SixBidSolo) advanceBid() {
	remaining := 0
	for i := range SixBidSoloPlayerCnt {
		if !s.passed[i] {
			remaining++
		}
	}
	// **全員パスなら配り直す。**局番号は進めない。
	if remaining == 0 {
		s.handNumber--
		s.rotateDealer()
		s.beginHand()
		return
	}
	if s.highBid != nil && remaining == 1 {
		s.settleBid()
		return
	}
	for range SixBidSoloPlayerCnt {
		s.bidIdx = (s.bidIdx + 1) % SixBidSoloPlayerCnt
		if !s.passed[s.bidIdx] {
			return
		}
	}
}

// settleBid は落札を確定する。
func (s *SixBidSolo) settleBid() {
	s.declarerIdx = s.highBid.Player
	s.phase = SixBidSoloPhaseDeclare
	// **ハート・ソロとミゼール系は選ぶ余地が無い。**そのままプレイへ。
	if !SixBidSoloNeedsTrump(s.highBid.Kind) {
		s.trumpSuit = SixBidSoloFixedTrump(s.highBid.Kind)
		s.startPlay()
	}
}

// Declare は切札 (とコール・ソロの指名札) を決める。
//
// suit は切札スート、called はコール・ソロで指名する札 (他のビッドでは nil)。
func (s *SixBidSolo) Declare(player, suit int, called *Card) error {
	if s.gameEndFlag {
		return errors.New("the game is over")
	}
	if s.phase != SixBidSoloPhaseDeclare {
		return errors.New("it is not the declaration phase")
	}
	if player != s.declarerIdx {
		return errors.New("only the declarer names the trump")
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return errors.New("that is not a suit")
	}
	s.trumpSuit = suit
	if s.highBid.Kind == SixBidSoloBidCall {
		if called == nil {
			return errors.New("a call solo must name a card")
		}
		// **この卓に無い札は指名できない。**A-10-K-Q-J-9-8-7-6 の 36 枚しか
		// 存在しないので、範囲外を通すと「ウィドウにある」と黙って扱われる。
		if !sixBidSoloInPack(called) {
			return errors.New("that card is not in this pack")
		}
		if err := s.applyCalledCard(called); err != nil {
			return err
		}
	}
	s.addLog(player, "declare", sixBidSoloBidName(s.highBid.Kind)+" "+sixBidSoloSuitName(suit), nil)
	s.startPlay()
	return nil
}

// applyCalledCard はコール・ソロの札交換を行う。
//
// **持っている者は交換に応じる義務がある。**ただしウィドウにあったときは
// 交換が起こらない。
func (s *SixBidSolo) applyCalledCard(called *Card) error {
	dec := s.GetPlayer(s.declarerIdx)
	if dec == nil {
		return errors.New("there is no declarer")
	}
	if sixBidSoloIndexOf(dec, called) >= 0 {
		return errors.New("you already hold that card")
	}
	s.calledCard = NewCard(called.GetDesign(), called.GetValue(), true)
	for i := range SixBidSoloPlayerCnt {
		if i == s.declarerIdx {
			continue
		}
		p := s.GetPlayer(i)
		idx := sixBidSoloIndexOf(p, called)
		if idx < 0 {
			continue
		}
		// **交換なので宣言者も 1 枚渡す。**どちらの手札も枚数は動かない。
		givenIdx := s.pickDiscard(dec)
		got := p.GetCard(idx)
		given := dec.GetCard(givenIdx)
		if got == nil || given == nil {
			return nil
		}
		p.RemoveCard(idx)
		dec.RemoveCard(givenIdx)
		dec.AddCard(got)
		p.AddCard(given)
		return nil
	}
	// ウィドウにある場合は交換なし。
	return nil
}

// pickDiscard は交換で渡す 1 枚 (いちばん点の低い札) を選ぶ。
func (s *SixBidSolo) pickDiscard(p *SixBidSoloPlayer) int {
	best, bestScore := 0, 1<<30
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		score := SixBidSoloCardPoint(c)*10 + sixBidSoloRank(c)
		if score < bestScore {
			best, bestScore = i, score
		}
	}
	return best
}

// sixBidSoloInPack は札がこの 36 枚に含まれるかを返す。
//
// **A-10-K-Q-J-9-8-7-6 だけ。**2 から 5 はそもそも配られていない。
func sixBidSoloInPack(c *Card) bool {
	if c == nil || c.GetDesign() < CardDesignSpade || c.GetDesign() > CardDesignDiamond {
		return false
	}
	switch c.GetValue() {
	case 1, 10, 13, 12, 11, 9, 8, 7, 6:
		return true
	}
	return false
}

// sixBidSoloIndexOf は手札の中の札の位置を返す (無ければ -1)。
func sixBidSoloIndexOf(p *SixBidSoloPlayer, c *Card) int {
	if p == nil || c == nil {
		return -1
	}
	for i := range p.GetCardsSize() {
		h := p.GetCard(i)
		if h != nil && h.GetDesign() == c.GetDesign() && h.GetValue() == c.GetValue() {
			return i
		}
	}
	return -1
}

// startPlay はプレイフェーズへ移る。
func (s *SixBidSolo) startPlay() {
	s.declared = true
	s.phase = SixBidSoloPhasePlay
	// **リードは親の左。**落札者ではない。
	s.trickLeader = (s.dealerIdx + 1) % SixBidSoloPlayerCnt
	s.currentIdx = s.trickLeader
	s.trick = make([]*Card, 0, SixBidSoloPlayerCnt)
}

// ---- プレイ ----

// SixBidSoloValidPlays は出せる手札インデックスを返す。
//
// **追随は強制。**持っていなければ何を出してもよい。
func (s *SixBidSolo) SixBidSoloValidPlays(player int) []int {
	p := s.GetPlayer(player)
	if p == nil {
		return nil
	}
	all := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		all = append(all, i)
	}
	if len(s.trick) == 0 || len(all) == 0 {
		return all
	}
	lead := s.trick[0]
	if lead == nil {
		return all
	}
	follow := make([]int, 0, len(all))
	for _, i := range all {
		if c := p.GetCard(i); c != nil && c.GetDesign() == lead.GetDesign() {
			follow = append(follow, i)
		}
	}
	if len(follow) > 0 {
		return follow
	}
	return all
}

// PlayCard は手札を 1 枚出す。
func (s *SixBidSolo) PlayCard(player, idx int) error {
	if s.gameEndFlag {
		return errors.New("the game is over")
	}
	if s.phase != SixBidSoloPhasePlay {
		return errors.New("it is not the play phase")
	}
	if player != s.currentIdx {
		return errors.New("it is not your turn")
	}
	p := s.GetPlayer(player)
	if p == nil || idx < 0 || idx >= p.GetCardsSize() {
		return errors.New("there is no such card")
	}
	if !sixBidSoloContains(s.SixBidSoloValidPlays(player), idx) {
		return errors.New("you must follow suit")
	}
	c := p.GetCard(idx)
	p.RemoveCard(idx)
	s.trick = append(s.trick, c)
	s.addLog(player, "play", "", []*Card{c})

	// **公開の条件は「他の 2 人が 1 枚ずつ出したら」。**単に 2 枚出たら、
	// ではない。落札者がリードする配席もあるので、宣言者以外が何人打ったかを
	// 数える必要がある。
	if s.highBid != nil && s.highBid.Kind == SixBidSoloBidSpreadMisere && !s.spreadOpen &&
		s.opponentsPlayedInTrick() >= SixBidSoloPlayerCnt-1 {
		s.spreadOpen = true
	}

	if len(s.trick) < SixBidSoloPlayerCnt {
		s.currentIdx = (s.currentIdx + 1) % SixBidSoloPlayerCnt
		return nil
	}
	s.resolveTrick()
	return nil
}

// opponentsPlayedInTrick は今のトリックで宣言者以外が何枚出したかを返す。
func (s *SixBidSolo) opponentsPlayedInTrick() int {
	n := 0
	for j := range len(s.trick) {
		seat := (s.trickLeader + j) % SixBidSoloPlayerCnt
		if seat != s.declarerIdx {
			n++
		}
	}
	return n
}

// resolveTrick はトリックの勝者を決める。
func (s *SixBidSolo) resolveTrick() {
	lead := s.trick[0]
	winOffset, best := 0, s.trick[0]
	for i := 1; i < len(s.trick); i++ {
		if sixBidSoloBeats(s.trick[i], best, lead, s.trumpSuit) {
			winOffset, best = i, s.trick[i]
		}
	}
	winner := (s.trickLeader + winOffset) % SixBidSoloPlayerCnt
	pts := 0
	for _, c := range s.trick {
		pts += SixBidSoloCardPoint(c)
	}
	s.points[winner] += pts
	s.tricksWon[winner]++
	s.addLog(winner, "trickWin", strconv.Itoa(pts)+"pt", s.trick)

	s.trick = make([]*Card, 0, SixBidSoloPlayerCnt)
	s.trickNumber++
	s.trickLeader = winner
	s.currentIdx = winner
	if s.trickNumber >= SixBidSoloTricks {
		s.finishHand()
	}
}

// sixBidSoloBeats は c が best を上回るかを返す。
func sixBidSoloBeats(c, best, lead *Card, trumpSuit int) bool {
	if c == nil || best == nil || lead == nil {
		return false
	}
	cTrump := trumpSuit != 0 && c.GetDesign() == trumpSuit
	bTrump := trumpSuit != 0 && best.GetDesign() == trumpSuit
	if cTrump != bTrump {
		return cTrump
	}
	if !cTrump && c.GetDesign() != lead.GetDesign() {
		return false
	}
	if !cTrump && best.GetDesign() != lead.GetDesign() {
		return true
	}
	return sixBidSoloRank(c) > sixBidSoloRank(best)
}

// sixBidSoloContains は s に v が含まれるかを返す。
func sixBidSoloContains(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// SixBidSoloWidowPoints はウィドウのカード点を返す。
func (s *SixBidSolo) SixBidSoloWidowPoints() int {
	pts := 0
	for _, c := range s.widow {
		pts += SixBidSoloCardPoint(c)
	}
	return pts
}

// finishHand は 1 局を精算する。
func (s *SixBidSolo) finishHand() {
	kind := SixBidSoloBidPass
	if s.highBid != nil {
		kind = s.highBid.Kind
	}
	dec := s.declarerIdx
	if dec < 0 || dec >= SixBidSoloPlayerCnt {
		s.phase = SixBidSoloPhaseHandEnd
		return
	}

	widowPts := 0
	// **ウィドウは宣言者に加算される。ただしミゼール系は除く。**
	if SixBidSoloUsesWidow(kind) {
		widowPts = s.SixBidSoloWidowPoints()
	}
	declarerPts := s.points[dec] + widowPts
	target := SixBidSoloTargetPoints(kind, s.trumpSuit)

	made := declarerPts >= target
	if SixBidSoloIsMisere(kind) {
		// **ミゼールは「0 トリック」ではなく「0 点」。**
		// 9・8・7・6 だけのトリックは取っても構わない。
		made = s.points[dec] == 0
	}

	value := s.sixBidSoloBidValue(kind, declarerPts)
	var deltas [SixBidSoloPlayerCnt]int
	sign := 1
	if !made {
		sign = -1
	}
	for i := range SixBidSoloPlayerCnt {
		if i == dec {
			// 対戦者 2 人からの受け払い。
			deltas[i] = sign * value * (SixBidSoloPlayerCnt - 1)
		} else {
			deltas[i] = -sign * value
		}
		s.scores[i] += deltas[i]
	}

	s.lastResult = &SixBidSoloHandResult{
		Kind:           kind,
		Declarer:       dec,
		DeclarerPoints: declarerPts,
		WidowPoints:    widowPts,
		Target:         target,
		Made:           made,
		Value:          value,
		Deltas:         deltas,
	}
	s.phase = SixBidSoloPhaseHandEnd
	s.addLog(dec, "settle", strconv.Itoa(declarerPts)+"/"+strconv.Itoa(target), nil)
	s.checkGameEnd()
}

// sixBidSoloBidValue は 1 人あたりの受け払い額を返す。
//
// **通常ビッドだけは固定額ではない。**60 点との差に倍率を掛ける。
func (s *SixBidSolo) sixBidSoloBidValue(kind SixBidSoloBidKind, declarerPts int) int {
	diff := declarerPts - SixBidSoloBaseTarget
	if diff < 0 {
		diff = -diff
	}
	switch kind {
	case SixBidSoloBidSolo:
		return diff * SixBidSoloSoloMultiplier
	case SixBidSoloBidHeartSolo:
		return diff * SixBidSoloHeartMultiplier
	case SixBidSoloBidMisere:
		return SixBidSoloMisereValue
	case SixBidSoloBidGuarantee:
		return SixBidSoloGuaranteeValue
	case SixBidSoloBidSpreadMisere:
		return SixBidSoloSpreadValue
	case SixBidSoloBidCall:
		if s.trumpSuit == CardDesignHeart {
			return SixBidSoloCallHeartValue
		}
		return SixBidSoloCallValue
	}
	return 0
}

// checkGameEnd は規定局数に達したかを見る。
func (s *SixBidSolo) checkGameEnd() {
	if s.handNumber < s.config.TargetHands {
		return
	}
	s.gameEndFlag = true
	s.phase = SixBidSoloPhaseGameEnd
	best, bestScore := 0, s.scores[0]
	for i := 1; i < SixBidSoloPlayerCnt; i++ {
		if s.scores[i] > bestScore {
			best, bestScore = i, s.scores[i]
		}
	}
	s.winnerIdx = best
}

// rotateDealer は親を 1 つ進める。
func (s *SixBidSolo) rotateDealer() {
	s.dealerIdx = (s.dealerIdx + 1) % SixBidSoloPlayerCnt
}

// NextHand は次の局を配る。
func (s *SixBidSolo) NextHand() error {
	if s.gameEndFlag {
		return errors.New("the game is over")
	}
	if s.phase != SixBidSoloPhaseHandEnd {
		return errors.New("the hand is still in progress")
	}
	s.rotateDealer()
	s.beginHand()
	return nil
}

// ---- CPU ----

// SixBidSoloCpuBid は CPU の宣言を決める (パスなら SixBidSoloBidPass)。
//
// **手札のカード点と切札の長さだけで見る。**ミゼール系は点札が無いときだけ。
func (s *SixBidSolo) SixBidSoloCpuBid(idx int) SixBidSoloBidKind {
	p := s.GetPlayer(idx)
	if p == nil {
		return SixBidSoloBidPass
	}
	pts, best, bestLen := 0, CardDesignSpade, 0
	counts := map[int]int{}
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		pts += SixBidSoloCardPoint(c)
		if c != nil {
			counts[c.GetDesign()]++
		}
	}
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if counts[suit] > bestLen {
			best, bestLen = suit, counts[suit]
		}
	}

	// 点札がまったく無ければミゼールを狙う。
	if pts == 0 {
		return s.highestAllowed(idx, SixBidSoloBidMisere)
	}
	switch {
	case pts >= 60 && bestLen >= 6:
		return s.highestAllowed(idx, SixBidSoloBidGuarantee)
	case pts >= 40 && bestLen >= 5:
		if best == CardDesignHeart {
			return s.highestAllowed(idx, SixBidSoloBidHeartSolo)
		}
		return s.highestAllowed(idx, SixBidSoloBidSolo)
	}
	return SixBidSoloBidPass
}

// highestAllowed は want 以下で実際に出せる最高のビッドを返す。
func (s *SixBidSolo) highestAllowed(idx int, want SixBidSoloBidKind) SixBidSoloBidKind {
	for k := want; k >= SixBidSoloMinBid; k-- {
		if s.SixBidSoloCanBid(idx, k) {
			return k
		}
	}
	// 立っている宣言を上回れないならパス。
	return SixBidSoloBidPass
}

// SixBidSoloCpuTrump は CPU が指定する切札を返す (いちばん長いスート)。
func (s *SixBidSolo) SixBidSoloCpuTrump(idx int) int {
	p := s.GetPlayer(idx)
	if p == nil {
		return CardDesignSpade
	}
	counts := map[int]int{}
	for i := range p.GetCardsSize() {
		if c := p.GetCard(i); c != nil {
			counts[c.GetDesign()]++
		}
	}
	best, bestLen := CardDesignSpade, -1
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if counts[suit] > bestLen {
			best, bestLen = suit, counts[suit]
		}
	}
	return best
}

// SixBidSoloCpuCall は CPU がコール・ソロで指名する札を返す。
//
// **持っていない切札のうち最強のもの。**手札にある札は指名できない。
func (s *SixBidSolo) SixBidSoloCpuCall(idx, suit int) *Card {
	p := s.GetPlayer(idx)
	if p == nil {
		return nil
	}
	for _, v := range []int{1, 10, 13, 12, 11} {
		c := NewCard(suit, v, true)
		if sixBidSoloIndexOf(p, c) < 0 {
			return c
		}
	}
	return NewCard(suit, 6, true)
}

// SixBidSoloCpuPlay は CPU が出す手札インデックスを返す (無ければ -1)。
func (s *SixBidSolo) SixBidSoloCpuPlay(idx int) int {
	p := s.GetPlayer(idx)
	if p == nil {
		return -1
	}
	valid := s.SixBidSoloValidPlays(idx)
	if len(valid) == 0 {
		return -1
	}
	misere := s.highBid != nil && SixBidSoloIsMisere(s.highBid.Kind)
	// **ミゼールの宣言者は点を取らないのが目的。**いちばん弱い札を捨てる。
	if misere && idx == s.declarerIdx {
		return sixBidSoloLowest(p, valid)
	}
	if len(s.trick) == 0 {
		return sixBidSoloHighest(p, valid)
	}
	lead := s.trick[0]
	best := s.trick[0]
	for _, c := range s.trick[1:] {
		if sixBidSoloBeats(c, best, lead, s.trumpSuit) {
			best = c
		}
	}
	for _, i := range valid {
		if sixBidSoloBeats(p.GetCard(i), best, lead, s.trumpSuit) {
			return i
		}
	}
	return sixBidSoloLowest(p, valid)
}

// sixBidSoloLowest は valid のうちいちばん弱い札を返す。
func sixBidSoloLowest(p *SixBidSoloPlayer, valid []int) int {
	best, bestRank := valid[0], 1<<30
	for _, i := range valid {
		if r := sixBidSoloRank(p.GetCard(i)); r < bestRank {
			best, bestRank = i, r
		}
	}
	return best
}

// sixBidSoloHighest は valid のうちいちばん強い札を返す。
func sixBidSoloHighest(p *SixBidSoloPlayer, valid []int) int {
	best, bestRank := valid[0], -1
	for _, i := range valid {
		if r := sixBidSoloRank(p.GetCard(i)); r > bestRank {
			best, bestRank = i, r
		}
	}
	return best
}

// IsHumanTurn は現在の手番が人間かを返す。
func (s *SixBidSolo) IsHumanTurn() bool {
	if s.gameEndFlag {
		return false
	}
	switch s.phase {
	case SixBidSoloPhaseBid:
		p := s.GetPlayer(s.bidIdx)
		return p != nil && p.GetIsHuman()
	case SixBidSoloPhaseDeclare:
		p := s.GetPlayer(s.declarerIdx)
		return p != nil && p.GetIsHuman()
	case SixBidSoloPhasePlay:
		p := s.GetPlayer(s.currentIdx)
		return p != nil && p.GetIsHuman()
	}
	return false
}

// CpuPlay は CPU が 1 アクション実行する。
func (s *SixBidSolo) CpuPlay() {
	if s.gameEndFlag || s.IsHumanTurn() {
		return
	}
	switch s.phase {
	case SixBidSoloPhaseBid:
		idx := s.bidIdx
		kind := s.SixBidSoloCpuBid(idx)
		if kind == SixBidSoloBidPass || s.Bid(idx, kind) != nil {
			_ = s.PassBid(idx)
		}
	case SixBidSoloPhaseDeclare:
		idx := s.declarerIdx
		suit := s.SixBidSoloCpuTrump(idx)
		var called *Card
		if s.highBid != nil && s.highBid.Kind == SixBidSoloBidCall {
			called = s.SixBidSoloCpuCall(idx, suit)
		}
		_ = s.Declare(idx, suit, called)
	case SixBidSoloPhasePlay:
		idx := s.currentIdx
		if i := s.SixBidSoloCpuPlay(idx); i >= 0 {
			_ = s.PlayCard(idx, i)
		}
	}
}

// ---- アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (s *SixBidSolo) GetPlayers() []*SixBidSoloPlayer { return s.players }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (s *SixBidSolo) GetPlayer(idx int) *SixBidSoloPlayer {
	return getPlayer(s.players, idx)
}

// GetPhase は現在のフェーズを返す。
func (s *SixBidSolo) GetPhase() SixBidSoloPhase { return s.phase }

// GetCurrentPlayerIdx は現在の手番を返す。
func (s *SixBidSolo) GetCurrentPlayerIdx() int { return s.currentIdx }

// GetBidPlayerIdx は宣言中の手番を返す。
func (s *SixBidSolo) GetBidPlayerIdx() int { return s.bidIdx }

// GetDealerIdx は親を返す。
func (s *SixBidSolo) GetDealerIdx() int { return s.dealerIdx }

// GetBids はこの局の宣言履歴を返す。
func (s *SixBidSolo) GetBids() []*SixBidSoloBid { return s.bids }

// GetHighBid は現在の最高宣言を返す。
func (s *SixBidSolo) GetHighBid() *SixBidSoloBid { return s.highBid }

// GetDeclarerIdx は落札者を返す (未定なら -1)。
func (s *SixBidSolo) GetDeclarerIdx() int { return s.declarerIdx }

// GetTrumpSuit は切札を返す (ミゼール系は 0)。
func (s *SixBidSolo) GetTrumpSuit() int { return s.trumpSuit }

// IsDeclared は切札が確定済みかを返す。
func (s *SixBidSolo) IsDeclared() bool { return s.declared }

// GetCalledCard はコール・ソロで指名された札を返す (無ければ nil)。
func (s *SixBidSolo) GetCalledCard() *Card { return s.calledCard }

// IsSpreadOpen はスプレッド・ミゼールで手札が公開済みかを返す。
func (s *SixBidSolo) IsSpreadOpen() bool { return s.spreadOpen }

// GetWidow はウィドウを返す。
//
// **中身が見えてよいのは精算後だけ。**公開判断はプレゼンター側で行う。
func (s *SixBidSolo) GetWidow() []*Card { return s.widow }

// GetTrick は場に出ている札を返す。
func (s *SixBidSolo) GetTrick() []*Card { return s.trick }

// GetTrickLeaderIdx はこのトリックのリード席を返す。
func (s *SixBidSolo) GetTrickLeaderIdx() int { return s.trickLeader }

// GetTrickNumber は済んだトリック数を返す。
func (s *SixBidSolo) GetTrickNumber() int { return s.trickNumber }

// GetPoints は席が取ったカード点を返す。
func (s *SixBidSolo) GetPoints(idx int) int {
	if idx < 0 || idx >= SixBidSoloPlayerCnt {
		return 0
	}
	return s.points[idx]
}

// GetTricksWon は席が取ったトリック数を返す。
func (s *SixBidSolo) GetTricksWon(idx int) int {
	if idx < 0 || idx >= SixBidSoloPlayerCnt {
		return 0
	}
	return s.tricksWon[idx]
}

// GetScore は席の通算得点を返す。
func (s *SixBidSolo) GetScore(idx int) int {
	if idx < 0 || idx >= SixBidSoloPlayerCnt {
		return 0
	}
	return s.scores[idx]
}

// GetLastResult は直前の局の精算を返す (まだ無ければ nil)。
func (s *SixBidSolo) GetLastResult() *SixBidSoloHandResult { return s.lastResult }

// GetHandNumber は現在の局番号を返す。
func (s *SixBidSolo) GetHandNumber() int { return s.handNumber }

// GetGameEndFlag はゲーム終了フラグを返す。
func (s *SixBidSolo) GetGameEndFlag() bool { return s.gameEndFlag }

// GetWinnerIdx は勝者の席を返す (未確定なら -1)。
func (s *SixBidSolo) GetWinnerIdx() int { return s.winnerIdx }

// GetConfig はゲーム設定を返す。
func (s *SixBidSolo) GetConfig() SixBidSoloConfig { return s.config }

// SetConfig はゲーム設定をセットする。
func (s *SixBidSolo) SetConfig(c SixBidSoloConfig) { s.config = c }

// GetActionLog は棋譜を返す。
func (s *SixBidSolo) GetActionLog() []*ActionLogEntry { return s.actionLog }

// addLog は棋譜を 1 件追加する。
func (s *SixBidSolo) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	s.appendLogAt(0, playerIdx, actionType, detail, cards)
}

// sixBidSoloBidName はビッドの内部名を返す (棋譜用)。
func sixBidSoloBidName(k SixBidSoloBidKind) string {
	switch k {
	case SixBidSoloBidSolo:
		return "solo"
	case SixBidSoloBidHeartSolo:
		return "heartSolo"
	case SixBidSoloBidMisere:
		return "misere"
	case SixBidSoloBidGuarantee:
		return "guarantee"
	case SixBidSoloBidSpreadMisere:
		return "spreadMisere"
	case SixBidSoloBidCall:
		return "callSolo"
	}
	return "pass"
}

// sixBidSoloSuitName はスートの内部名を返す (棋譜用)。
func sixBidSoloSuitName(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "S"
	case CardDesignClover:
		return "C"
	case CardDesignHeart:
		return "H"
	case CardDesignDiamond:
		return "D"
	}
	return "-"
}

// ---- テスト用 ----

// SetPhaseForTest はフェーズを差し替える (テスト専用)。
func (s *SixBidSolo) SetPhaseForTest(p SixBidSoloPhase) { s.phase = p }

// SetHandForTest は手札を差し替える (テスト専用)。
func (s *SixBidSolo) SetHandForTest(idx int, cards []*Card) {
	setHandForTest(s.GetPlayer(idx), cards)
}

// SetWidowForTest はウィドウを差し替える (テスト専用)。
func (s *SixBidSolo) SetWidowForTest(cards []*Card) { s.widow = cards }

// SetContractForTest は契約を差し替える (テスト専用)。
func (s *SixBidSolo) SetContractForTest(declarer int, kind SixBidSoloBidKind, trumpSuit int) {
	s.declarerIdx = declarer
	s.highBid = &SixBidSoloBid{Player: declarer, Kind: kind}
	s.trumpSuit = trumpSuit
	s.declared = true
}

// SetCurrentPlayerForTest は手番を差し替える (テスト専用)。
func (s *SixBidSolo) SetCurrentPlayerForTest(idx int) { s.currentIdx = idx }

// SetDealerForTest は親を差し替える (テスト専用)。
func (s *SixBidSolo) SetDealerForTest(idx int) { s.dealerIdx = idx }

// SetBidPlayerForTest は宣言中の手番を差し替える (テスト専用)。
func (s *SixBidSolo) SetBidPlayerForTest(idx int) { s.bidIdx = idx }

// SetTrickLeaderForTest はリード席を差し替える (テスト専用)。
func (s *SixBidSolo) SetTrickLeaderForTest(idx int) { s.trickLeader = idx }

// SetPointsForTest は取得カード点を差し替える (テスト専用)。
func (s *SixBidSolo) SetPointsForTest(idx, n int) {
	if idx >= 0 && idx < SixBidSoloPlayerCnt {
		s.points[idx] = n
	}
}

// SetTricksWonForTest は取得トリック数を差し替える (テスト専用)。
func (s *SixBidSolo) SetTricksWonForTest(idx, n int) {
	if idx >= 0 && idx < SixBidSoloPlayerCnt {
		s.tricksWon[idx] = n
	}
}

// SetScoreForTest は通算得点を差し替える (テスト専用)。
func (s *SixBidSolo) SetScoreForTest(idx, n int) {
	if idx >= 0 && idx < SixBidSoloPlayerCnt {
		s.scores[idx] = n
	}
}

// SetHandNumberForTest は局番号を差し替える (テスト専用)。
func (s *SixBidSolo) SetHandNumberForTest(n int) { s.handNumber = n }

// FinishHandForTest は精算を走らせる (テスト専用)。
func (s *SixBidSolo) FinishHandForTest() { s.finishHand() }

// ---- JSON ----

// sixBidSoloJSON is the KV wire format for SixBidSolo.
type sixBidSoloJSON struct {
	Players     []*SixBidSoloPlayer       `json:"pl"`
	Config      SixBidSoloConfig          `json:"cf"`
	Phase       SixBidSoloPhase           `json:"ph"`
	DealerIdx   int                       `json:"di"`
	BidIdx      int                       `json:"bi"`
	CurrentIdx  int                       `json:"ci"`
	Bids        []*SixBidSoloBid          `json:"bd"`
	HighBid     *SixBidSoloBid            `json:"hb"`
	DeclarerIdx int                       `json:"de"`
	Passed      [SixBidSoloPlayerCnt]bool `json:"ps"`
	Widow       []*Card                   `json:"wd"`
	TrumpSuit   int                       `json:"ts"`
	Declared    bool                      `json:"dc"`
	CalledCard  *Card                     `json:"cc"`
	SpreadOpen  bool                      `json:"so"`
	Trick       []*Card                   `json:"tk"`
	TrickLeader int                       `json:"tl"`
	TrickNumber int                       `json:"tn"`
	Points      [SixBidSoloPlayerCnt]int  `json:"pt"`
	TricksWon   [SixBidSoloPlayerCnt]int  `json:"tw"`
	Scores      [SixBidSoloPlayerCnt]int  `json:"sc"`
	LastResult  *SixBidSoloHandResult     `json:"lr"`
	HandNumber  int                       `json:"hn"`
	GameEndFlag bool                      `json:"ge"`
	WinnerIdx   int                       `json:"wi"`
	ActionLog   []*ActionLogEntry         `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (s *SixBidSolo) MarshalJSON() ([]byte, error) {
	return json.Marshal(sixBidSoloJSON{
		Players: s.players, Config: s.config, Phase: s.phase,
		DealerIdx: s.dealerIdx, BidIdx: s.bidIdx, CurrentIdx: s.currentIdx,
		Bids: s.bids, HighBid: s.highBid, DeclarerIdx: s.declarerIdx, Passed: s.passed,
		Widow: s.widow, TrumpSuit: s.trumpSuit, Declared: s.declared,
		CalledCard: s.calledCard, SpreadOpen: s.spreadOpen,
		Trick: s.trick, TrickLeader: s.trickLeader, TrickNumber: s.trickNumber,
		Points: s.points, TricksWon: s.tricksWon, Scores: s.scores,
		LastResult: s.lastResult, HandNumber: s.handNumber,
		GameEndFlag: s.gameEndFlag, WinnerIdx: s.winnerIdx, ActionLog: s.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **KV から戻る値なので範囲を検査する。**壊れた状態をそのまま受け入れると
// 添字で落ちる。
func (s *SixBidSolo) UnmarshalJSON(data []byte) error {
	var j sixBidSoloJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != SixBidSoloPlayerCnt {
		return errors.New("six-bid solo needs exactly three seats")
	}
	if j.Phase < SixBidSoloPhaseBid || j.Phase > SixBidSoloPhaseGameEnd {
		return errors.New("unknown phase")
	}
	for name, v := range map[string]int{"dealer": j.DealerIdx, "bid seat": j.BidIdx, "current seat": j.CurrentIdx, "trick leader": j.TrickLeader} {
		if v < 0 || v >= SixBidSoloPlayerCnt {
			return errors.New("bad " + name)
		}
	}
	for name, v := range map[string]int{"declarer": j.DeclarerIdx, "winner": j.WinnerIdx} {
		if v < -1 || v >= SixBidSoloPlayerCnt {
			return errors.New("bad " + name)
		}
	}
	if j.TrumpSuit < 0 || j.TrumpSuit > CardDesignDiamond {
		return errors.New("bad trump suit")
	}
	if len(j.Trick) > SixBidSoloPlayerCnt {
		return errors.New("a trick cannot hold more cards than there are seats")
	}
	if len(j.Widow) > SixBidSoloWidowSize {
		return errors.New("the widow cannot hold more than three cards")
	}
	if j.HighBid != nil && (j.HighBid.Kind < SixBidSoloBidPass || j.HighBid.Kind >= SixBidSoloBidCount) {
		return errors.New("unknown bid")
	}
	if j.TrickNumber < 0 || j.TrickNumber > SixBidSoloTricks {
		return errors.New("bad trick number")
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}

	s.players = j.Players
	s.config = j.Config
	s.phase = j.Phase
	s.dealerIdx = j.DealerIdx
	s.bidIdx = j.BidIdx
	s.currentIdx = j.CurrentIdx
	s.bids = j.Bids
	s.highBid = j.HighBid
	s.declarerIdx = j.DeclarerIdx
	s.passed = j.Passed
	s.widow = j.Widow
	s.trumpSuit = j.TrumpSuit
	s.declared = j.Declared
	s.calledCard = j.CalledCard
	s.spreadOpen = j.SpreadOpen
	s.trick = j.Trick
	s.trickLeader = j.TrickLeader
	s.trickNumber = j.TrickNumber
	s.points = j.Points
	s.tricksWon = j.TricksWon
	s.scores = j.Scores
	s.lastResult = j.LastResult
	s.handNumber = j.HandNumber
	s.gameEndFlag = j.GameEndFlag
	s.winnerIdx = j.WinnerIdx
	s.actionLog = j.ActionLog
	return nil
}
