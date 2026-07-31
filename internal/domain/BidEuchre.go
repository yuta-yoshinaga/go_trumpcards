//go:build !js || !wasm || extra2

// Package domain — ビッド・ユーカー (Bid Euchre) のドメインモデル。
//
// アメリカ中西部のユーカー変種。24 枚 (A-K-Q-J-10-9)、4 人 2 対 2、各自 6 枚。
//
// # issue #4386 の仕様案との相違
//
//   - issue は「全員に等分配札する (**残りはキティ**)」とするが、**キティは無い**。
//     24 ÷ 4 = 6 でちょうど配り切りなので、そもそも余りが出ない
//   - issue は最低ビッドに触れていない。**3 トリック**から
//   - issue は**ディーラーの同額落札**に触れていない。通常は上回らないと通らない
//     が、**ディーラーだけは同額で奪える**
//   - issue は切札スートしか書いていないが、**ノートランプが 2 種類ある**。
//     *no trump high* は通常の序列、***no trump low* は序列が逆転し 9 が最強**
//     (9-10-J-Q-K-A)
//   - issue は目標点を書いていない。**32 点**
//   - issue の「未達なら減点」は額が曖昧。**宣言したトリック数**ぶん引かれる
//     (取ったトリック数ではない)。**このときも守備側は自分のトリックを得点する**
//
// issue が合っている点: ボワー序列 (切札 J が最強、同色 J が 2 番目)、左ボワーは
// 全ての目的で切札スートに属する、達成時は両チームがトリック数ぶん加点。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// BidEuchrePlayerCnt はプレイヤー数。
const BidEuchrePlayerCnt = 4

// BidEuchreTeamCnt はチーム数。
const BidEuchreTeamCnt = 2

// BidEuchreHandSize は各プレイヤーの手札枚数。
//
// **24 ÷ 4 = 6 でちょうど配り切る。**キティは残らない。
const BidEuchreHandSize = 6

// BidEuchreDeckSize はデッキの枚数 (A-K-Q-J-10-9 × 4 スート)。
const BidEuchreDeckSize = 24

// ビッドの範囲。
const (
	// BidEuchreMinBid は最低ビッド。**3 トリックから。**
	BidEuchreMinBid = 3
	// BidEuchreMaxBid は最高ビッド (全トリック)。
	BidEuchreMaxBid = BidEuchreHandSize
)

// BidEuchreGameTarget は勝利に要る点。
const BidEuchreGameTarget = 32

// BidEuchreTrump は落札者が選ぶ宣言。
type BidEuchreTrump int

// Bid Euchre の宣言種別
const (
	// BidEuchreTrumpSpade ♠ 切札
	BidEuchreTrumpSpade BidEuchreTrump = iota
	// BidEuchreTrumpClub ♣ 切札
	BidEuchreTrumpClub
	// BidEuchreTrumpDiamond ♦ 切札
	BidEuchreTrumpDiamond
	// BidEuchreTrumpHeart ♥ 切札
	BidEuchreTrumpHeart
	// BidEuchreTrumpNoHigh ノートランプ・ハイ (通常の序列)
	BidEuchreTrumpNoHigh
	// BidEuchreTrumpNoLow ノートランプ・ロー (**序列が逆転し 9 が最強**)
	BidEuchreTrumpNoLow
	// BidEuchreTrumpCount 宣言種別の総数
	BidEuchreTrumpCount
)

// BidEuchreTrumpSuit は宣言の切札スートを返す (ノートランプは 0)。
func BidEuchreTrumpSuit(t BidEuchreTrump) int {
	switch t {
	case BidEuchreTrumpSpade:
		return CardDesignSpade
	case BidEuchreTrumpClub:
		return CardDesignClover
	case BidEuchreTrumpDiamond:
		return CardDesignDiamond
	case BidEuchreTrumpHeart:
		return CardDesignHeart
	}
	return 0
}

// BidEuchreIsNoTrump は切札なしの宣言かを返す。
func BidEuchreIsNoTrump(t BidEuchreTrump) bool {
	return t == BidEuchreTrumpNoHigh || t == BidEuchreTrumpNoLow
}

// BidEuchreTeamOf は席のチームを返す (0/2 が team 0、1/3 が team 1)。
//
// **範囲外は -1。**Go の剰余は負の被除数で負を返すので、未確定の -1 を
// そのまま渡すと -1 が返る。チーム添字として使う側が弾けるよう明示する。
func BidEuchreTeamOf(seat int) int {
	if seat < 0 || seat >= BidEuchrePlayerCnt {
		return -1
	}
	return seat % BidEuchreTeamCnt
}

// bidEuchreSameColour は 2 つのスートが同色かを返す。
//
// ♠♣ が黒、♥♦ が赤。左ボワーの判定に要る。
func bidEuchreSameColour(a, b int) bool {
	black := func(s int) bool { return s == CardDesignSpade || s == CardDesignClover }
	return black(a) == black(b)
}

// IsBidEuchreRightBower は切札スートの J かを返す。
func IsBidEuchreRightBower(c *Card, trumpSuit int) bool {
	return c != nil && trumpSuit != 0 && c.GetDesign() == trumpSuit && c.GetValue() == 11
}

// IsBidEuchreLeftBower は同色スートの J かを返す。
//
// **左ボワーは全ての目的で切札スートに属する。**追随の判定でも切札として扱う。
func IsBidEuchreLeftBower(c *Card, trumpSuit int) bool {
	return c != nil && trumpSuit != 0 && c.GetValue() == 11 &&
		c.GetDesign() != trumpSuit && bidEuchreSameColour(c.GetDesign(), trumpSuit)
}

// BidEuchreEffectiveSuit は札の**実効スート**を返す。
//
// **左ボワーは切札スートとして扱う。**素の GetDesign を使うと、左ボワーを
// 元のスートに追随させてしまう。
func BidEuchreEffectiveSuit(c *Card, trumpSuit int) int {
	if c == nil {
		return 0
	}
	if IsBidEuchreLeftBower(c, trumpSuit) {
		return trumpSuit
	}
	return c.GetDesign()
}

// bidEuchrePlainRank は切札でない札の強さを返す。
//
// **ノートランプ・ローでは逆転する。**9 が最強で A が最弱になる。
func bidEuchrePlainRank(c *Card, t BidEuchreTrump) int {
	if c == nil {
		return 0
	}
	normal := map[int]int{1: 6, 13: 5, 12: 4, 11: 3, 10: 2, 9: 1}
	base, ok := normal[c.GetValue()]
	if !ok {
		return 0
	}
	if t == BidEuchreTrumpNoLow {
		return 7 - base
	}
	return base
}

// BidEuchreCardRank は札の強さを返す。
//
// 切札は **右ボワー > 左ボワー > A > K > Q > 10 > 9**。ボワーが 2 枚割り込む
// ので、切札スートの序列だけ他と違う。
func BidEuchreCardRank(c *Card, t BidEuchreTrump) int {
	if c == nil {
		return 0
	}
	trumpSuit := BidEuchreTrumpSuit(t)
	if trumpSuit == 0 {
		return bidEuchrePlainRank(c, t)
	}
	if IsBidEuchreRightBower(c, trumpSuit) {
		return 100
	}
	if IsBidEuchreLeftBower(c, trumpSuit) {
		return 99
	}
	if c.GetDesign() == trumpSuit {
		// 切札の残りは A K Q 10 9 の順。J は上で処理済み。
		switch c.GetValue() {
		case 1:
			return 96
		case 13:
			return 95
		case 12:
			return 94
		case 10:
			return 93
		case 9:
			return 92
		}
		return 90
	}
	return bidEuchrePlainRank(c, t)
}

// BidEuchrePhase はゲームフェーズ。
type BidEuchrePhase int

// Bid Euchre のフェーズ定数
const (
	// BidEuchrePhaseBid ビッド
	BidEuchrePhaseBid BidEuchrePhase = iota
	// BidEuchrePhaseChooseTrump 落札者が切札を指定する
	BidEuchrePhaseChooseTrump
	// BidEuchrePhasePlay トリックプレイ
	BidEuchrePhasePlay
	// BidEuchrePhaseHandEnd 局終了 (精算済み)
	BidEuchrePhaseHandEnd
	// BidEuchrePhaseGameEnd ゲーム終了
	BidEuchrePhaseGameEnd
)

// BidEuchreBid は 1 件の宣言。
type BidEuchreBid struct {
	Player int
	// Value は宣言したトリック数 (0 ならパス)。
	Value int
}

// BidEuchreHandResult は 1 局の精算結果。
type BidEuchreHandResult struct {
	// Points は各チームがこの局で得た点 (未達側は負)。
	Points [BidEuchreTeamCnt]int
	// Tricks は各チームが取ったトリック数。
	Tricks [BidEuchreTeamCnt]int
	// Made は落札側が達成したか。
	Made bool
	// Bid は宣言したトリック数。
	Bid int
}

// BidEuchre はビッド・ユーカーのゲームクラス。
type BidEuchre struct {
	players []*BidEuchrePlayer
	config  BidEuchreConfig
	phase   BidEuchrePhase

	dealerIdx  int
	currentIdx int
	bidIdx     int
	bids       []*BidEuchreBid
	passCount  int

	highBid *BidEuchreBid
	// declarerIdx は落札者 (-1 なら未確定)。
	declarerIdx int
	trump       BidEuchreTrump
	trumpSuit   int
	// trumpChosen は落札者が宣言を選んだか。
	trumpChosen bool

	trick       []*Card
	trickLeader int
	trickNumber int
	tricksWon   [BidEuchrePlayerCnt]int

	scores     [BidEuchreTeamCnt]int
	lastResult *BidEuchreHandResult

	handNumber  int
	gameEndFlag bool
	winnerTeam  int

	actionLog []*ActionLogEntry
}

// NewBidEuchre コンストラクタ
func NewBidEuchre(players []*BidEuchrePlayer, config BidEuchreConfig) *BidEuchre {
	return &BidEuchre{players: players, config: config, winnerTeam: -1, declarerIdx: -1}
}

// NewDefaultBidEuchre はデフォルト構成のゲームを返す。
func NewDefaultBidEuchre() *BidEuchre {
	players := make([]*BidEuchrePlayer, BidEuchrePlayerCnt)
	for i := range players {
		players[i] = NewBidEuchrePlayer(i == 0)
	}
	return NewBidEuchre(players, DefaultBidEuchreConfig())
}

// ---- 進行 ----

// Reset ゲーム初期化
func (b *BidEuchre) Reset() {
	b.gameEndFlag = false
	b.winnerTeam = -1
	b.scores = [BidEuchreTeamCnt]int{}
	b.lastResult = nil
	b.handNumber = 0
	b.dealerIdx = 0
	b.actionLog = nil
	b.beginHand()
}

// beginHand は 1 局を配ってビッドへ入る。
func (b *BidEuchre) beginHand() {
	b.handNumber++
	b.phase = BidEuchrePhaseBid
	b.bids = nil
	b.passCount = 0
	b.highBid = nil
	b.declarerIdx = -1
	b.trump = BidEuchreTrumpSpade
	b.trumpSuit = 0
	b.trumpChosen = false
	b.trick = nil
	b.trickNumber = 0
	b.trickLeader = -1
	b.tricksWon = [BidEuchrePlayerCnt]int{}

	for _, p := range b.players {
		p.ResetRound()
	}

	deck := newBidEuchreDeck()
	bidEuchreShuffle(deck)
	// **配り切る。**24 ÷ 4 = 6 なのでキティは残らない。
	pos := 0
	for range BidEuchreHandSize {
		for i := range BidEuchrePlayerCnt {
			b.players[i].AddCard(deck[pos])
			pos++
		}
	}

	b.bidIdx = (b.dealerIdx + 1) % BidEuchrePlayerCnt
	b.addLog(-1, "deal", fmt.Sprintf("%d cards each, no kitty", BidEuchreHandSize), nil)
}

// newBidEuchreDeck は 24 枚 (A-K-Q-J-10-9) のデッキを返す。
func newBidEuchreDeck() []*Card {
	values := []int{1, 13, 12, 11, 10, 9}
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	deck := make([]*Card, 0, BidEuchreDeckSize)
	for _, s := range suits {
		for _, v := range values {
			deck = append(deck, NewCard(s, v, true))
		}
	}
	return deck
}

// bidEuchreShuffle は Fisher-Yates。
func bidEuchreShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// checkBidTurn は宣言できる状態かを確かめる。
func (b *BidEuchre) checkBidTurn(player int) error {
	if b.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if b.phase != BidEuchrePhaseBid {
		return fmt.Errorf("bidding is not in progress")
	}
	if player != b.bidIdx {
		return fmt.Errorf("it is not player %d's turn to bid", player)
	}
	return nil
}

// BidEuchreCanBid は player がその額で宣言できるかを返す。
//
// **ディーラーだけは同額で奪える。**他の席は上回らないと通らない。
func (b *BidEuchre) BidEuchreCanBid(player, value int) bool {
	if value < BidEuchreMinBid || value > BidEuchreMaxBid {
		return false
	}
	if b.highBid == nil {
		return true
	}
	if player == b.dealerIdx {
		return value >= b.highBid.Value
	}
	return value > b.highBid.Value
}

// Bid はトリック数を宣言する。
func (b *BidEuchre) Bid(player, value int) error {
	if err := b.checkBidTurn(player); err != nil {
		return err
	}
	if value < BidEuchreMinBid || value > BidEuchreMaxBid {
		return fmt.Errorf("a bid must be between %d and %d", BidEuchreMinBid, BidEuchreMaxBid)
	}
	if !b.BidEuchreCanBid(player, value) {
		return fmt.Errorf("a bid must beat the standing %d", b.highBid.Value)
	}
	rec := &BidEuchreBid{Player: player, Value: value}
	b.bids = append(b.bids, rec)
	b.highBid = rec
	b.passCount = 0
	b.addLog(player, "bid", fmt.Sprintf("bids %d", value), nil)
	b.advanceBid()
	return nil
}

// PassBid は宣言を見送る。
func (b *BidEuchre) PassBid(player int) error {
	if err := b.checkBidTurn(player); err != nil {
		return err
	}
	b.bids = append(b.bids, &BidEuchreBid{Player: player, Value: 0})
	b.passCount++
	b.addLog(player, "pass", "passes", nil)
	b.advanceBid()
	return nil
}

// advanceBid は次の宣言手番へ進め、決着していれば切札指定へ移る。
func (b *BidEuchre) advanceBid() {
	// **ディーラーまで一周したら終わり。**同額落札の特権があるので、
	// ディーラーが動いた時点で競りは閉じる。
	last := b.bidIdx == b.dealerIdx
	if b.highBid != nil && (last || b.passCount >= BidEuchrePlayerCnt-1) {
		b.settleBid()
		return
	}
	if b.highBid == nil && b.passCount >= BidEuchrePlayerCnt {
		// **全員パスなら配り直し。**
		b.addLog(-1, "redeal", "everybody passed", nil)
		b.handNumber--
		b.beginHand()
		return
	}
	b.bidIdx = (b.bidIdx + 1) % BidEuchrePlayerCnt
}

// settleBid は落札を確定して切札指定へ移る。
func (b *BidEuchre) settleBid() {
	b.declarerIdx = b.highBid.Player
	b.phase = BidEuchrePhaseChooseTrump
	b.currentIdx = b.declarerIdx
	b.addLog(b.declarerIdx, "won_bid", fmt.Sprintf("wins the auction at %d", b.highBid.Value), nil)
}

// ChooseTrump は落札者が切札またはノートランプを指定する。
func (b *BidEuchre) ChooseTrump(player int, t BidEuchreTrump) error {
	if b.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if b.phase != BidEuchrePhaseChooseTrump {
		return fmt.Errorf("it is not time to name trump")
	}
	if player != b.declarerIdx {
		return fmt.Errorf("only the declarer names trump")
	}
	if t < BidEuchreTrumpSpade || t >= BidEuchreTrumpCount {
		return fmt.Errorf("bad trump declaration: %d", t)
	}
	b.trump = t
	b.trumpSuit = BidEuchreTrumpSuit(t)
	b.trumpChosen = true
	b.phase = BidEuchrePhasePlay
	// **リードは落札者。**
	b.trickLeader = b.declarerIdx
	b.currentIdx = b.declarerIdx
	b.addLog(player, "trump", fmt.Sprintf("names declaration %d", t), nil)
	return nil
}

// BidEuchreValidPlays は player が出せる手札インデックスを返す。
//
// **左ボワーは切札スートとして追随する。**素のスートで判定すると、切札が
// リードされたときに左ボワーを出さずに済んでしまう。
func (b *BidEuchre) BidEuchreValidPlays(player int) []int {
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
	leadSuit := BidEuchreEffectiveSuit(b.trick[0], b.trumpSuit)
	same := make([]int, 0, len(all))
	for _, i := range all {
		if BidEuchreEffectiveSuit(p.GetCard(i), b.trumpSuit) == leadSuit {
			same = append(same, i)
		}
	}
	if len(same) > 0 {
		return same
	}
	return all
}

// PlayCard は 1 枚出す。
func (b *BidEuchre) PlayCard(player, idx int) error {
	if b.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if b.phase != BidEuchrePhasePlay {
		return fmt.Errorf("the play phase is not in progress")
	}
	if player != b.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	p := b.GetPlayer(player)
	if p == nil || idx < 0 || idx >= p.GetCardsSize() {
		return fmt.Errorf("bad card index: %d", idx)
	}
	if !bidEuchreContains(b.BidEuchreValidPlays(player), idx) {
		return fmt.Errorf("that card may not be played")
	}

	card := p.GetCard(idx)
	p.RemoveCard(idx)
	b.trick = append(b.trick, card)
	b.addLog(player, "play", "plays a card", []*Card{card})

	if len(b.trick) < BidEuchrePlayerCnt {
		b.currentIdx = (player + 1) % BidEuchrePlayerCnt
		return nil
	}
	b.resolveTrick()
	return nil
}

// resolveTrick はトリックを解決する。
func (b *BidEuchre) resolveTrick() {
	lead := b.trick[0]
	leadSuit := BidEuchreEffectiveSuit(lead, b.trumpSuit)
	bestOffset := 0
	bestIsTrump := b.trumpSuit != 0 && leadSuit == b.trumpSuit
	bestRank := BidEuchreCardRank(lead, b.trump)
	for i := 1; i < len(b.trick); i++ {
		c := b.trick[i]
		if c == nil {
			continue
		}
		suit := BidEuchreEffectiveSuit(c, b.trumpSuit)
		isTrump := b.trumpSuit != 0 && suit == b.trumpSuit
		rank := BidEuchreCardRank(c, b.trump)
		switch {
		case isTrump && !bestIsTrump:
			bestOffset, bestIsTrump, bestRank = i, true, rank
		case isTrump == bestIsTrump && suit == BidEuchreEffectiveSuit(b.trick[bestOffset], b.trumpSuit) && rank > bestRank:
			bestOffset, bestRank = i, rank
		}
	}
	winner := (b.trickLeader + bestOffset) % BidEuchrePlayerCnt

	b.tricksWon[winner]++
	b.trickNumber++
	b.trick = nil
	b.trickLeader = winner
	b.currentIdx = winner
	b.addLog(winner, "trick", fmt.Sprintf("takes trick %d", b.trickNumber), nil)

	if b.trickNumber >= BidEuchreHandSize {
		b.finishHand()
	}
}

// BidEuchreTeamTricks はチームが取ったトリック数を返す。
func (b *BidEuchre) BidEuchreTeamTricks(team int) int {
	if team < 0 || team >= BidEuchreTeamCnt {
		return 0
	}
	total := 0
	for i := range BidEuchrePlayerCnt {
		if BidEuchreTeamOf(i) == team {
			total += b.tricksWon[i]
		}
	}
	return total
}

// finishHand は局を精算する。
//
// **達成なら両チームが取ったトリック数ぶん得点する。**未達なら落札側は
// **宣言したトリック数**ぶんマイナスで、守備側は自分のトリックを得点する。
func (b *BidEuchre) finishHand() {
	declTeam := BidEuchreTeamOf(b.declarerIdx)
	defTeam := 1 - declTeam
	bid := b.highBid.Value

	res := &BidEuchreHandResult{Bid: bid}
	for team := range BidEuchreTeamCnt {
		res.Tricks[team] = b.BidEuchreTeamTricks(team)
	}
	res.Made = res.Tricks[declTeam] >= bid

	if res.Made {
		res.Points[declTeam] = res.Tricks[declTeam]
		b.addLog(b.declarerIdx, "made", fmt.Sprintf("makes %d for a bid of %d", res.Tricks[declTeam], bid), nil)
	} else {
		// **引かれるのは宣言額。**取ったトリック数ではない。
		res.Points[declTeam] = -bid
		b.addLog(b.declarerIdx, "set", fmt.Sprintf("is set and loses %d", bid), nil)
	}
	// **守備側は達成/未達に関係なく自分のトリックを得点する。**
	res.Points[defTeam] = res.Tricks[defTeam]

	for team := range BidEuchreTeamCnt {
		b.scores[team] += res.Points[team]
	}
	b.lastResult = res

	b.phase = BidEuchrePhaseHandEnd
	b.checkGameEnd()
}

// checkGameEnd は目標点に届いていれば決着させる。
func (b *BidEuchre) checkGameEnd() {
	a, c := b.scores[0], b.scores[1]
	if a < BidEuchreGameTarget && c < BidEuchreGameTarget {
		return
	}
	switch {
	case a >= BidEuchreGameTarget && c >= BidEuchreGameTarget:
		// 両方超えたら**落札側**が優先する。
		b.winnerTeam = BidEuchreTeamOf(b.declarerIdx)
	case a >= BidEuchreGameTarget:
		b.winnerTeam = 0
	default:
		b.winnerTeam = 1
	}
	b.gameEndFlag = true
	b.phase = BidEuchrePhaseGameEnd
	b.addLog(-1, "game_end", fmt.Sprintf("team %d wins", b.winnerTeam), nil)
}

// NextHand は次の局を配る。
func (b *BidEuchre) NextHand() error {
	if b.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if b.phase != BidEuchrePhaseHandEnd {
		return fmt.Errorf("the hand is still in progress")
	}
	b.dealerIdx = (b.dealerIdx + 1) % BidEuchrePlayerCnt
	b.beginHand()
	return nil
}

// bidEuchreContains は s に v が含まれるかを返す。
func bidEuchreContains(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// ---- CPU ----

// BidEuchreCpuBid は CPU の宣言額を決める (0 ならパス)。
func (b *BidEuchre) BidEuchreCpuBid(idx int) int {
	p := b.GetPlayer(idx)
	if p == nil {
		return 0
	}
	// 一番強くなるスートで見込みトリック数を数える。
	best := 0
	for t := BidEuchreTrumpSpade; t <= BidEuchreTrumpHeart; t++ {
		suit := BidEuchreTrumpSuit(t)
		count := 0
		for i := range p.GetCardsSize() {
			c := p.GetCard(i)
			if IsBidEuchreRightBower(c, suit) || IsBidEuchreLeftBower(c, suit) {
				count++
				continue
			}
			if c != nil && c.GetDesign() == suit && (c.GetValue() == 1 || c.GetValue() == 13) {
				count++
			}
		}
		if count > best {
			best = count
		}
	}
	if best < BidEuchreMinBid {
		return 0
	}
	if best > BidEuchreMaxBid {
		best = BidEuchreMaxBid
	}
	if !b.BidEuchreCanBid(idx, best) {
		return 0
	}
	return best
}

// BidEuchreCpuTrump は CPU が選ぶ宣言を返す。
func (b *BidEuchre) BidEuchreCpuTrump(idx int) BidEuchreTrump {
	p := b.GetPlayer(idx)
	if p == nil {
		return BidEuchreTrumpSpade
	}
	best, bestCount := BidEuchreTrumpSpade, -1
	for t := BidEuchreTrumpSpade; t <= BidEuchreTrumpHeart; t++ {
		suit := BidEuchreTrumpSuit(t)
		count := 0
		for i := range p.GetCardsSize() {
			c := p.GetCard(i)
			if BidEuchreEffectiveSuit(c, suit) == suit {
				count++
			}
		}
		if count > bestCount {
			best, bestCount = t, count
		}
	}
	return best
}

// BidEuchreCpuPlay は CPU が出す手札インデックスを返す。
func (b *BidEuchre) BidEuchreCpuPlay(idx int) int {
	valid := b.BidEuchreValidPlays(idx)
	if len(valid) == 0 {
		return -1
	}
	p := b.GetPlayer(idx)
	if p == nil {
		return valid[0]
	}
	if len(b.trick) == 0 {
		return bidEuchreHighest(p, valid, b.trump)
	}
	lead := b.trick[0]
	winning, winRank := -1, -1
	for _, i := range valid {
		c := p.GetCard(i)
		if bidEuchreBeats(c, lead, b.trump, b.trumpSuit) && BidEuchreCardRank(c, b.trump) > winRank {
			winning, winRank = i, BidEuchreCardRank(c, b.trump)
		}
	}
	if winning >= 0 {
		return winning
	}
	return bidEuchreLowest(p, valid, b.trump)
}

// bidEuchreLowest は valid のうち一番弱い札の索引を返す。
func bidEuchreLowest(p *BidEuchrePlayer, valid []int, t BidEuchreTrump) int {
	best, bestRank := valid[0], 1<<30
	for _, i := range valid {
		if r := BidEuchreCardRank(p.GetCard(i), t); r < bestRank {
			best, bestRank = i, r
		}
	}
	return best
}

// bidEuchreHighest は valid のうち一番強い札の索引を返す。
func bidEuchreHighest(p *BidEuchrePlayer, valid []int, t BidEuchreTrump) int {
	best, bestRank := valid[0], -1
	for _, i := range valid {
		if r := BidEuchreCardRank(p.GetCard(i), t); r > bestRank {
			best, bestRank = i, r
		}
	}
	return best
}

// bidEuchreBeats は c が lead に勝つかを返す。
func bidEuchreBeats(c, lead *Card, t BidEuchreTrump, trumpSuit int) bool {
	if c == nil || lead == nil {
		return false
	}
	cs := BidEuchreEffectiveSuit(c, trumpSuit)
	ls := BidEuchreEffectiveSuit(lead, trumpSuit)
	if trumpSuit != 0 && cs == trumpSuit && ls != trumpSuit {
		return true
	}
	if cs != ls {
		return false
	}
	return BidEuchreCardRank(c, t) > BidEuchreCardRank(lead, t)
}

// IsHumanTurn は今が人間の手番かを返す。
func (b *BidEuchre) IsHumanTurn() bool {
	if b.gameEndFlag {
		return false
	}
	switch b.phase {
	case BidEuchrePhaseBid:
		p := b.GetPlayer(b.bidIdx)
		return p != nil && p.GetIsHuman()
	case BidEuchrePhaseChooseTrump, BidEuchrePhasePlay:
		p := b.GetPlayer(b.currentIdx)
		return p != nil && p.GetIsHuman()
	}
	return false
}

// CpuPlay は今の手番の CPU に 1 手打たせる。
func (b *BidEuchre) CpuPlay() {
	if b.gameEndFlag {
		return
	}
	switch b.phase {
	case BidEuchrePhaseBid:
		idx := b.bidIdx
		if p := b.GetPlayer(idx); p == nil || p.GetIsHuman() {
			return
		}
		value := b.BidEuchreCpuBid(idx)
		if value < BidEuchreMinBid || b.Bid(idx, value) != nil {
			_ = b.PassBid(idx)
		}
	case BidEuchrePhaseChooseTrump:
		idx := b.declarerIdx
		if p := b.GetPlayer(idx); p == nil || p.GetIsHuman() {
			return
		}
		_ = b.ChooseTrump(idx, b.BidEuchreCpuTrump(idx))
	case BidEuchrePhasePlay:
		idx := b.currentIdx
		if p := b.GetPlayer(idx); p == nil || p.GetIsHuman() {
			return
		}
		if i := b.BidEuchreCpuPlay(idx); i >= 0 {
			_ = b.PlayCard(idx, i)
		}
	}
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (b *BidEuchre) GetPlayers() []*BidEuchrePlayer { return b.players }

// GetPlayer は idx のプレイヤーを返す。
func (b *BidEuchre) GetPlayer(idx int) *BidEuchrePlayer {
	if idx < 0 || idx >= len(b.players) {
		return nil
	}
	return b.players[idx]
}

// GetPhase は現在のフェーズを返す。
func (b *BidEuchre) GetPhase() BidEuchrePhase { return b.phase }

// GetCurrentPlayerIdx は現在の手番を返す。
func (b *BidEuchre) GetCurrentPlayerIdx() int { return b.currentIdx }

// GetBidPlayerIdx は宣言中の手番を返す。
func (b *BidEuchre) GetBidPlayerIdx() int { return b.bidIdx }

// GetDealerIdx はディーラーを返す。
func (b *BidEuchre) GetDealerIdx() int { return b.dealerIdx }

// GetBids はこの局の宣言履歴を返す。
func (b *BidEuchre) GetBids() []*BidEuchreBid { return b.bids }

// GetHighBid は現在の最高宣言を返す (未落札なら nil)。
func (b *BidEuchre) GetHighBid() *BidEuchreBid { return b.highBid }

// GetDeclarerIdx は落札者を返す (-1 なら未確定)。
func (b *BidEuchre) GetDeclarerIdx() int { return b.declarerIdx }

// GetTrump は宣言種別を返す。
func (b *BidEuchre) GetTrump() BidEuchreTrump { return b.trump }

// GetTrumpSuit は切札スートを返す (0 ならノートランプ)。
func (b *BidEuchre) GetTrumpSuit() int { return b.trumpSuit }

// IsTrumpChosen は落札者が宣言を選んだかを返す。
func (b *BidEuchre) IsTrumpChosen() bool { return b.trumpChosen }

// GetTrick は場に出ている札を返す。
func (b *BidEuchre) GetTrick() []*Card { return b.trick }

// GetTrickLeaderIdx はこのトリックのリード席を返す。
func (b *BidEuchre) GetTrickLeaderIdx() int { return b.trickLeader }

// GetTrickNumber は済んだトリック数を返す。
func (b *BidEuchre) GetTrickNumber() int { return b.trickNumber }

// GetTricksWon は席が取ったトリック数を返す。
func (b *BidEuchre) GetTricksWon(idx int) int {
	if idx < 0 || idx >= BidEuchrePlayerCnt {
		return 0
	}
	return b.tricksWon[idx]
}

// GetScore はチームの通算点を返す。
func (b *BidEuchre) GetScore(team int) int {
	if team < 0 || team >= BidEuchreTeamCnt {
		return 0
	}
	return b.scores[team]
}

// GetLastResult は直前の局の精算を返す (まだ無ければ nil)。
func (b *BidEuchre) GetLastResult() *BidEuchreHandResult { return b.lastResult }

// GetHandNumber は現在の局番号を返す。
func (b *BidEuchre) GetHandNumber() int { return b.handNumber }

// GetGameEndFlag はゲーム終了フラグを返す。
func (b *BidEuchre) GetGameEndFlag() bool { return b.gameEndFlag }

// GetWinnerTeam は勝利チームを返す (-1 なら未決)。
func (b *BidEuchre) GetWinnerTeam() int { return b.winnerTeam }

// GetConfig は設定を返す。
func (b *BidEuchre) GetConfig() BidEuchreConfig { return b.config }

// SetConfig は設定をセットする。
func (b *BidEuchre) SetConfig(c BidEuchreConfig) { b.config = c }

// GetActionLog は棋譜を返す。
func (b *BidEuchre) GetActionLog() []*ActionLogEntry { return b.actionLog }

// addLog は棋譜を 1 行足す。
func (b *BidEuchre) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	b.actionLog = append(b.actionLog, &ActionLogEntry{
		TurnNumber: len(b.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// SetPhaseForTest はテスト用にフェーズを設定する。
func (b *BidEuchre) SetPhaseForTest(p BidEuchrePhase) { b.phase = p }

// SetHandForTest はテスト用に手札を差し替える。
func (b *BidEuchre) SetHandForTest(idx int, cards []*Card) {
	p := b.GetPlayer(idx)
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

// SetContractForTest はテスト用に契約を設定する。
func (b *BidEuchre) SetContractForTest(declarer, value int, t BidEuchreTrump) {
	b.declarerIdx = declarer
	b.highBid = &BidEuchreBid{Player: declarer, Value: value}
	b.trump = t
	b.trumpSuit = BidEuchreTrumpSuit(t)
	b.trumpChosen = true
}

// SetCurrentPlayerForTest はテスト用に手番を設定する。
func (b *BidEuchre) SetCurrentPlayerForTest(idx int) { b.currentIdx = idx }

// SetDealerForTest はテスト用にディーラーを設定する。
func (b *BidEuchre) SetDealerForTest(idx int) { b.dealerIdx = idx }

// SetBidPlayerForTest はテスト用に宣言手番を設定する。
func (b *BidEuchre) SetBidPlayerForTest(idx int) { b.bidIdx = idx }

// SetTrickLeaderForTest はテスト用にリード席を設定する。
func (b *BidEuchre) SetTrickLeaderForTest(idx int) { b.trickLeader = idx }

// SetTricksWonForTest はテスト用に取得トリック数を設定する。
func (b *BidEuchre) SetTricksWonForTest(idx, n int) {
	if idx >= 0 && idx < BidEuchrePlayerCnt {
		b.tricksWon[idx] = n
	}
}

// SetScoreForTest はテスト用に通算点を設定する。
func (b *BidEuchre) SetScoreForTest(team, n int) {
	if team >= 0 && team < BidEuchreTeamCnt {
		b.scores[team] = n
	}
}

// FinishHandForTest はテスト用に精算を走らせる。
func (b *BidEuchre) FinishHandForTest() { b.finishHand() }

// bidEuchreJSON is the JSON wire format for BidEuchre.
type bidEuchreJSON struct {
	Players     []*BidEuchrePlayer      `json:"pl"`
	Config      BidEuchreConfig         `json:"cf"`
	Phase       BidEuchrePhase          `json:"ph"`
	DealerIdx   int                     `json:"di"`
	CurrentIdx  int                     `json:"ci"`
	BidIdx      int                     `json:"bi"`
	Bids        []*BidEuchreBid         `json:"bd"`
	PassCount   int                     `json:"pc"`
	HighBid     *BidEuchreBid           `json:"hb"`
	DeclarerIdx int                     `json:"de"`
	Trump       BidEuchreTrump          `json:"tp"`
	TrumpSuit   int                     `json:"ts"`
	TrumpChosen bool                    `json:"tc"`
	Trick       []*Card                 `json:"tk"`
	TrickLeader int                     `json:"tl"`
	TrickNumber int                     `json:"tn"`
	TricksWon   [BidEuchrePlayerCnt]int `json:"tw"`
	Scores      [BidEuchreTeamCnt]int   `json:"sc"`
	LastResult  *BidEuchreHandResult    `json:"lr"`
	HandNumber  int                     `json:"hn"`
	GameEndFlag bool                    `json:"ge"`
	WinnerTeam  int                     `json:"wt"`
	ActionLog   []*ActionLogEntry       `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (b *BidEuchre) MarshalJSON() ([]byte, error) {
	return json.Marshal(bidEuchreJSON{
		Players: b.players, Config: b.config, Phase: b.phase,
		DealerIdx: b.dealerIdx, CurrentIdx: b.currentIdx, BidIdx: b.bidIdx,
		Bids: b.bids, PassCount: b.passCount, HighBid: b.highBid,
		DeclarerIdx: b.declarerIdx, Trump: b.trump, TrumpSuit: b.trumpSuit,
		TrumpChosen: b.trumpChosen,
		Trick:       b.trick, TrickLeader: b.trickLeader, TrickNumber: b.trickNumber,
		TricksWon: b.tricksWon, Scores: b.scores, LastResult: b.lastResult,
		HandNumber: b.handNumber, GameEndFlag: b.gameEndFlag, WinnerTeam: b.winnerTeam,
		ActionLog: b.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **復元でしか入らない値を弾く。**KV から戻ってきた壊れた状態でプレイが
// 詰まないよう、席番号・宣言・スートを検証する。
func (b *BidEuchre) UnmarshalJSON(data []byte) error {
	var j bidEuchreJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != BidEuchrePlayerCnt {
		return fmt.Errorf("bad player count: %d", len(j.Players))
	}
	if j.Phase < BidEuchrePhaseBid || j.Phase > BidEuchrePhaseGameEnd {
		return fmt.Errorf("bad phase: %d", j.Phase)
	}
	for name, v := range map[string]int{"dealer": j.DealerIdx, "current": j.CurrentIdx, "bid": j.BidIdx} {
		if v < 0 || v >= BidEuchrePlayerCnt {
			return fmt.Errorf("bad %s index: %d", name, v)
		}
	}
	for name, v := range map[string]int{"declarer": j.DeclarerIdx, "trick leader": j.TrickLeader} {
		if v < -1 || v >= BidEuchrePlayerCnt {
			return fmt.Errorf("bad %s index: %d", name, v)
		}
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= BidEuchreTeamCnt {
		return fmt.Errorf("bad winner team: %d", j.WinnerTeam)
	}
	if j.Trump < BidEuchreTrumpSpade || j.Trump >= BidEuchreTrumpCount {
		return fmt.Errorf("bad trump declaration: %d", j.Trump)
	}
	if j.TrumpSuit != 0 && (j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return fmt.Errorf("bad trump suit: %d", j.TrumpSuit)
	}
	if len(j.Trick) > BidEuchrePlayerCnt {
		return fmt.Errorf("bad trick size: %d", len(j.Trick))
	}
	if j.HighBid != nil && j.HighBid.Value != 0 &&
		(j.HighBid.Value < BidEuchreMinBid || j.HighBid.Value > BidEuchreMaxBid) {
		return fmt.Errorf("bad high bid: %d", j.HighBid.Value)
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
	b.trump = j.Trump
	b.trumpSuit = j.TrumpSuit
	b.trumpChosen = j.TrumpChosen
	b.trick = j.Trick
	b.trickLeader = j.TrickLeader
	b.trickNumber = j.TrickNumber
	b.tricksWon = j.TricksWon
	b.scores = j.Scores
	b.lastResult = j.LastResult
	b.handNumber = j.HandNumber
	b.gameEndFlag = j.GameEndFlag
	b.winnerTeam = j.WinnerTeam
	b.actionLog = j.ActionLog
	return nil
}
