//go:build !js || !wasm || extra3

// Package domain — カイザー (Kaiser) のドメインモデル。
//
// カナダ・サスカチュワン州のビッド式パートナーシップ・トリックゲーム。
//
// # issue #4393 の仕様案との相違
//
//   - issue は「**32 枚**（標準 52 枚からランク 2〜6 を除外）」とするが、
//     **ランク 2〜6 を除外すると ♥5 も ♠3 も消える**。このゲームの名前の由来で
//     あり得点の要である 2 枚が、指定されたデッキに存在しないという自己矛盾。
//     正しくは各スート A-K-Q-J-10-9-8-7 の 32 枚に **♥5 と ♠3 を加えた 34 枚**。
//     4 人 × 8 枚 = 32 枚 + **キティ 2 枚** = 34 枚と算術でも合う
//   - issue は**キティ**に触れていない。落札者はキティ 2 枚を手札に加え、
//     2 枚を伏せて捨てる。**ただし ♥5 と ♠3 は捨てられない**
//   - issue は「**取得トリック数**を宣言」とするが、宣言するのは**点数**。
//     トリックは 8 つしかないのに ♥5 が 5 点ある。範囲は **7〜12** で、
//     キティを見られる利が大きいため最低ビッドが 6 ではなく 7 に設定されている
//   - issue は**ノートランプ / ロー・ノートランプ**に触れていない。Low No Trump
//     ではランクが逆転し **7 が最強**になる
//   - issue は目標点を書いていない。**52 点**（ノートランプのビッドが成功した
//     局があれば 62 点）。加えて **45 点以上に達した側は自分がビッドしない限り
//     加点できない**
//
// issue が合っている点: 1 局の合計は **10 点**（トリック 8 + ♥5 の 5 −
// ♠3 の 3）、ビッド未達なら宣言額をマイナス。
package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// KaiserPlayerCnt はプレイヤー数 (4 人 2 対 2)。
const KaiserPlayerCnt = 4

// KaiserTeamCnt はチーム数。
const KaiserTeamCnt = 2

// KaiserHandSize は各プレイヤーの手札枚数。
const KaiserHandSize = 8

// KaiserKittySize はキティの枚数。
//
// **配り切りではない。**4 × 8 = 32 にこの 2 枚を足して 34 枚デッキになる。
const KaiserKittySize = 2

// KaiserDeckSize はデッキの枚数。
//
// **32 ではない。**各スート A-K-Q-J-10-9-8-7 の 32 枚に ♥5 と ♠3 を加える。
const KaiserDeckSize = 34

// ビッドの範囲。
const (
	// KaiserMinBid は最低ビッド。
	//
	// **6 ではなく 7。**キティを見られる利が大きいぶん下限が上げてある。
	KaiserMinBid = 7
	// KaiserMaxBid は最高ビッド (全トリック + ♥5 − ♠3 の理論最大ではなく、
	// 慣習上の上限)。
	KaiserMaxBid = 12
)

// 特殊札の点数。
const (
	// KaiserHeartFiveBonus は ♥5 を取った側の加点。
	KaiserHeartFiveBonus = 5
	// KaiserSpadeThreePenalty は ♠3 を取った側の減点。
	KaiserSpadeThreePenalty = -3
)

// KaiserHandTotal は 1 局で動く点数の合計。
//
// トリック 8 + ♥5 の 5 − ♠3 の 3 = 10。issue の「合計 10 点」は正しい。
const KaiserHandTotal = 8 + KaiserHeartFiveBonus + KaiserSpadeThreePenalty

// 目標点。
const (
	// KaiserTargetScore は既定の目標点。
	KaiserTargetScore = 52
	// KaiserNoTrumpTargetScore はノートランプのビッドが成功したあとの目標点。
	KaiserNoTrumpTargetScore = 62
	// KaiserMustBidThreshold はこれ以上ではビッドしないと加点できない境目。
	KaiserMustBidThreshold = 45
)

// KaiserContract は落札した契約の種別。
type KaiserContract int

// Kaiser の契約種別
const (
	// KaiserContractTrump 切札あり
	KaiserContractTrump KaiserContract = iota
	// KaiserContractNoTrump ノートランプ
	KaiserContractNoTrump
	// KaiserContractLowNoTrump ロー・ノートランプ (ランクが逆転し 7 が最強)
	KaiserContractLowNoTrump
)

// KaiserPhase はゲームフェーズ。
type KaiserPhase int

// Kaiser のフェーズ定数
const (
	// KaiserPhaseBid ビッド
	KaiserPhaseBid KaiserPhase = iota
	// KaiserPhaseDiscard 落札者がキティを取り込んで 2 枚捨てる
	KaiserPhaseDiscard
	// KaiserPhasePlay トリックプレイ
	KaiserPhasePlay
	// KaiserPhaseHandEnd 局終了 (精算済み)
	KaiserPhaseHandEnd
	// KaiserPhaseGameEnd ゲーム終了
	KaiserPhaseGameEnd
)

// KaiserTeamOf は席のチームを返す (0/2 が team 0、1/3 が team 1)。
func KaiserTeamOf(seat int) int { return seat % KaiserTeamCnt }

// IsKaiserHeartFive は ♥5 かを返す。
func IsKaiserHeartFive(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignHeart && c.GetValue() == 5
}

// IsKaiserSpadeThree は ♠3 かを返す。
func IsKaiserSpadeThree(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignSpade && c.GetValue() == 3
}

// newKaiserDeck は 34 枚のデッキを返す。
//
// **各スート A-K-Q-J-10-9-8-7 の 32 枚 + ♥5 + ♠3。**issue の「ランク 2〜6 を
// 除外」では ♥5 と ♠3 が消えてしまう。
func newKaiserDeck() []*Card {
	values := []int{1, 13, 12, 11, 10, 9, 8, 7} // A,K,Q,J,10,9,8,7
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
	deck := make([]*Card, 0, KaiserDeckSize)
	for _, s := range suits {
		for _, v := range values {
			deck = append(deck, NewCard(s, v, true))
		}
	}
	deck = append(deck, NewCard(CardDesignHeart, 5, true), NewCard(CardDesignSpade, 3, true))
	return deck
}

// kaiserShuffle は Fisher-Yates。
func kaiserShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// KaiserBid は 1 件のビッド。
type KaiserBid struct {
	// Player は宣言した席。
	Player int
	// Value は宣言した**点数** (トリック数ではない)。パスなら 0。
	Value int
	// Contract は種別。
	Contract KaiserContract
}

// Kaiser はカイザーのゲームクラス。
type Kaiser struct {
	players []*KaiserPlayer
	config  KaiserConfig
	phase   KaiserPhase

	deck  []*Card
	kitty []*Card

	dealerIdx  int
	currentIdx int
	bidIdx     int
	// bids はこの局のビッド履歴。
	bids []*KaiserBid
	// passCount は連続パス数。
	passCount int

	// highBid は現在の最高ビッド (Value 0 なら未落札)。
	highBid *KaiserBid
	// declarerIdx は落札者 (-1 なら未確定)。
	declarerIdx int
	// trumpSuit は切札 (0 ならノートランプ系)。
	trumpSuit int
	contract  KaiserContract

	trick       []*Card
	trickLeader int
	trickNumber int

	// handPoints はこの局で各チームが取った点。
	handPoints [KaiserTeamCnt]int
	// heartFiveBy / spadeThreeBy は特殊札を取った席 (-1 なら未取得)。
	heartFiveBy  int
	spadeThreeBy int
	// bidMade は直前の局で落札側が達成したか。
	bidMade bool

	scores [KaiserTeamCnt]int
	// targetScore は現在の目標点。ノートランプ成功で上がる。
	targetScore int

	handNumber  int
	gameEndFlag bool
	winnerTeam  int

	actionLogBase
}

// NewKaiser コンストラクタ
func NewKaiser(players []*KaiserPlayer, config KaiserConfig) *Kaiser {
	return &Kaiser{players: players, config: config, winnerTeam: -1, declarerIdx: -1}
}

// NewDefaultKaiser はデフォルト構成のゲームを返す。
func NewDefaultKaiser() *Kaiser {
	players := make([]*KaiserPlayer, KaiserPlayerCnt)
	for i := range players {
		players[i] = NewKaiserPlayer(i == 0)
	}
	return NewKaiser(players, DefaultKaiserConfig())
}

// ---- ランク ----

// KaiserCardRank は札の強さを返す。
//
// **Low No Trump では順序が逆転する。**7 が最強で A が最弱になる。
// ♥5 と ♠3 はそれぞれのスートで A の 1 つ下に入る。
func KaiserCardRank(c *Card, contract KaiserContract) int {
	if c == nil {
		return 0
	}
	// 特殊札は素のランク順に馴染まないので、A の下・K の上に固定で挟む。
	special := IsKaiserHeartFive(c) || IsKaiserSpadeThree(c)
	normal := map[int]int{1: 8, 13: 6, 12: 5, 11: 4, 10: 3, 9: 2, 8: 1, 7: 0}
	base, ok := normal[c.GetValue()]
	if !ok {
		if !special {
			return 0
		}
		base = 7 // A(8) の 1 つ下、K(6) の 1 つ上
	}
	if contract == KaiserContractLowNoTrump {
		if special {
			// **特殊札は逆転しない。**5 と 3 は常に A のすぐ下に居る。
			return 7
		}
		return 8 - base
	}
	return base
}

// ---- 進行 ----

// Reset ゲーム初期化
func (k *Kaiser) Reset() {
	k.gameEndFlag = false
	k.winnerTeam = -1
	k.scores = [KaiserTeamCnt]int{}
	k.targetScore = KaiserTargetScore
	k.handNumber = 0
	k.dealerIdx = 0
	k.actionLog = nil
	k.beginHand()
}

// beginHand は 1 局を配ってビッドへ入る。
func (k *Kaiser) beginHand() {
	k.handNumber++
	k.phase = KaiserPhaseBid
	k.bids = nil
	k.passCount = 0
	k.highBid = nil
	k.declarerIdx = -1
	k.trumpSuit = 0
	k.contract = KaiserContractTrump
	k.trick = nil
	k.trickNumber = 0
	k.trickLeader = -1
	k.handPoints = [KaiserTeamCnt]int{}
	k.heartFiveBy = -1
	k.spadeThreeBy = -1
	k.bidMade = false

	for _, p := range k.players {
		p.ResetRound()
	}

	k.deck = newKaiserDeck()
	kaiserShuffle(k.deck)
	pos := 0
	for range KaiserHandSize {
		for i := range KaiserPlayerCnt {
			k.players[i].AddCard(k.deck[pos])
			pos++
		}
	}
	// **残りがキティ。**32 枚デッキではここが空になってしまう。
	k.kitty = append([]*Card(nil), k.deck[pos:]...)

	k.bidIdx = (k.dealerIdx + 1) % KaiserPlayerCnt
	k.addLog(-1, "deal", fmt.Sprintf("8 cards each with a kitty of %d", len(k.kitty)), nil)
}

// checkBidTurn はビッドできる状態かを確かめる。
func (k *Kaiser) checkBidTurn(player int) error {
	if k.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if k.phase != KaiserPhaseBid {
		return fmt.Errorf("bidding is not in progress")
	}
	if player != k.bidIdx {
		return fmt.Errorf("it is not player %d's turn to bid", player)
	}
	return nil
}

// KaiserBidRank は比較用の順位を返す。
//
// **同じ数字でも Low No Trump が最も高い。**No Trump はその 1 つ下、切札が
// いちばん下になる。
func KaiserBidRank(value int, contract KaiserContract) int {
	return value*3 + int(contract)
}

// KaiserIsNoTrump は契約がノートランプ系かを返す。
func KaiserIsNoTrump(contract KaiserContract) bool {
	return contract == KaiserContractNoTrump || contract == KaiserContractLowNoTrump
}

// Bid は点数を宣言する。
func (k *Kaiser) Bid(player, value int, contract KaiserContract) error {
	if err := k.checkBidTurn(player); err != nil {
		return err
	}
	if value < KaiserMinBid || value > KaiserMaxBid {
		return fmt.Errorf("a bid must be between %d and %d", KaiserMinBid, KaiserMaxBid)
	}
	if contract < KaiserContractTrump || contract > KaiserContractLowNoTrump {
		return fmt.Errorf("bad contract: %d", contract)
	}
	// **設定でノートランプを切れる。**契約はここで決まるので、読まないと
	// config が黙って効かなくなる。SetTrump は切札契約のスート名指しだけを
	// 扱うので、そちらでは弾けない。
	if !k.config.AllowNoTrump && KaiserIsNoTrump(contract) {
		return fmt.Errorf("no-trump bids are switched off")
	}
	if k.highBid != nil && KaiserBidRank(value, contract) <= KaiserBidRank(k.highBid.Value, k.highBid.Contract) {
		return fmt.Errorf("a bid must beat the standing %d", k.highBid.Value)
	}
	bid := &KaiserBid{Player: player, Value: value, Contract: contract}
	k.bids = append(k.bids, bid)
	k.highBid = bid
	k.passCount = 0
	k.addLog(player, "bid", fmt.Sprintf("bids %d", value), nil)
	k.advanceBid()
	return nil
}

// PassBid はビッドを見送る。
func (k *Kaiser) PassBid(player int) error {
	if err := k.checkBidTurn(player); err != nil {
		return err
	}
	k.bids = append(k.bids, &KaiserBid{Player: player, Value: 0})
	k.passCount++
	k.addLog(player, "pass", "passes", nil)
	k.advanceBid()
	return nil
}

// advanceBid は次のビッド手番へ進め、決着していれば契約へ移る。
func (k *Kaiser) advanceBid() {
	if k.highBid != nil && k.passCount >= KaiserPlayerCnt-1 {
		k.settleBid()
		return
	}
	if k.highBid == nil && k.passCount >= KaiserPlayerCnt {
		// **全員パスなら配り直し。**同じディーラーでもう一度配る。
		k.addLog(-1, "redeal", "everybody passed", nil)
		k.handNumber--
		k.beginHand()
		return
	}
	k.bidIdx = (k.bidIdx + 1) % KaiserPlayerCnt
}

// settleBid は落札を確定してキティを渡す。
func (k *Kaiser) settleBid() {
	k.declarerIdx = k.highBid.Player
	k.contract = k.highBid.Contract
	p := k.GetPlayer(k.declarerIdx)
	for _, c := range k.kitty {
		p.AddCard(c)
	}
	k.kitty = nil
	k.phase = KaiserPhaseDiscard
	k.currentIdx = k.declarerIdx
	k.addLog(k.declarerIdx, "take_kitty", "takes the kitty", nil)
}

// SetTrump は落札者が切札を指定する。
//
// ノートランプ系の契約では呼べない。
func (k *Kaiser) SetTrump(player, suit int) error {
	if k.phase != KaiserPhaseDiscard {
		return fmt.Errorf("the contract is not being set")
	}
	if player != k.declarerIdx {
		return fmt.Errorf("only the declarer names trump")
	}
	if k.contract != KaiserContractTrump {
		return fmt.Errorf("a no-trump contract has no trump suit")
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return fmt.Errorf("bad suit: %d", suit)
	}
	k.trumpSuit = suit
	k.addLog(player, "set_trump", fmt.Sprintf("names suit %d trump", suit), nil)
	return nil
}

// Discard は落札者がキティを取り込んだあと 2 枚を捨てる。
//
// **♥5 と ♠3 は捨てられない。**これを許すと落札者が ♠3 を無条件で処分できて
// しまう。
func (k *Kaiser) Discard(player int, idxs []int) error {
	if k.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if k.phase != KaiserPhaseDiscard {
		return fmt.Errorf("it is not time to discard")
	}
	if player != k.declarerIdx {
		return fmt.Errorf("only the declarer discards")
	}
	if k.contract == KaiserContractTrump && k.trumpSuit == 0 {
		return fmt.Errorf("name the trump suit first")
	}
	if len(idxs) != KaiserKittySize {
		return fmt.Errorf("discard exactly %d cards", KaiserKittySize)
	}
	p := k.GetPlayer(player)
	seen := map[int]bool{}
	for _, i := range idxs {
		if i < 0 || i >= p.GetCardsSize() {
			return fmt.Errorf("bad card index: %d", i)
		}
		if seen[i] {
			return fmt.Errorf("duplicate card index: %d", i)
		}
		seen[i] = true
		c := p.GetCard(i)
		if IsKaiserHeartFive(c) || IsKaiserSpadeThree(c) {
			return fmt.Errorf("the five of hearts and three of spades may not be discarded")
		}
	}
	// 大きい索引から抜く。前から抜くと後続の索引がずれる。
	sorted := append([]int(nil), idxs...)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] > sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	for _, i := range sorted {
		p.RemoveCard(i)
	}
	k.phase = KaiserPhasePlay
	// **リードは落札者。**ディーラーの左隣ではない。
	k.trickLeader = k.declarerIdx
	k.currentIdx = k.declarerIdx
	k.addLog(player, "discard", "discards two cards", nil)
	return nil
}

// KaiserValidPlays は player が出せる手札インデックスを返す。
//
// 追随のみ強制。切札の強制はない。
func (k *Kaiser) KaiserValidPlays(player int) []int {
	p := k.GetPlayer(player)
	if p == nil {
		return nil
	}
	all := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		all = append(all, i)
	}
	if len(k.trick) == 0 || k.trick[0] == nil {
		return all
	}
	leadSuit := k.trick[0].GetDesign()
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
func (k *Kaiser) PlayCard(player, idx int) error {
	if k.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if k.phase != KaiserPhasePlay {
		return fmt.Errorf("the play phase is not in progress")
	}
	if player != k.currentIdx {
		return fmt.Errorf("it is not player %d's turn", player)
	}
	p := k.GetPlayer(player)
	if p == nil || idx < 0 || idx >= p.GetCardsSize() {
		return fmt.Errorf("bad card index: %d", idx)
	}
	if !kaiserContains(k.KaiserValidPlays(player), idx) {
		return fmt.Errorf("that card may not be played")
	}

	card := p.GetCard(idx)
	p.RemoveCard(idx)
	k.trick = append(k.trick, card)
	k.addLog(player, "play", "plays a card", []*Card{card})

	if len(k.trick) < KaiserPlayerCnt {
		k.currentIdx = (player + 1) % KaiserPlayerCnt
		return nil
	}
	k.resolveTrick()
	return nil
}

// resolveTrick はトリックを解決する。
func (k *Kaiser) resolveTrick() {
	lead := k.trick[0]
	leadSuit := lead.GetDesign()
	bestOffset := 0
	bestIsTrump := k.trumpSuit != 0 && leadSuit == k.trumpSuit
	bestRank := KaiserCardRank(lead, k.contract)
	for i := 1; i < len(k.trick); i++ {
		c := k.trick[i]
		if c == nil {
			continue
		}
		isTrump := k.trumpSuit != 0 && c.GetDesign() == k.trumpSuit
		rank := KaiserCardRank(c, k.contract)
		switch {
		case isTrump && !bestIsTrump:
			bestOffset, bestIsTrump, bestRank = i, true, rank
		case isTrump == bestIsTrump && c.GetDesign() == k.trick[bestOffset].GetDesign() && rank > bestRank:
			bestOffset, bestRank = i, rank
		}
	}
	winner := (k.trickLeader + bestOffset) % KaiserPlayerCnt
	team := KaiserTeamOf(winner)

	k.handPoints[team]++
	for _, c := range k.trick {
		if IsKaiserHeartFive(c) {
			k.heartFiveBy = winner
			k.handPoints[team] += KaiserHeartFiveBonus
		}
		if IsKaiserSpadeThree(c) {
			k.spadeThreeBy = winner
			k.handPoints[team] += KaiserSpadeThreePenalty
		}
	}

	k.trickNumber++
	k.trick = nil
	k.trickLeader = winner
	k.currentIdx = winner
	k.addLog(winner, "trick", fmt.Sprintf("takes trick %d", k.trickNumber), nil)

	if k.trickNumber >= KaiserHandSize {
		k.finishHand()
	}
}

// finishHand は局を精算する。
func (k *Kaiser) finishHand() {
	declTeam := KaiserTeamOf(k.declarerIdx)
	defTeam := 1 - declTeam
	bid := k.highBid.Value

	if k.handPoints[declTeam] >= bid {
		k.bidMade = true
		k.scores[declTeam] += k.handPoints[declTeam]
		// **ノートランプが通ったら目標点が上がる。**
		if k.contract != KaiserContractTrump {
			k.targetScore = KaiserNoTrumpTargetScore
		}
		k.addLog(k.declarerIdx, "hand_end", fmt.Sprintf("makes %d for a bid of %d", k.handPoints[declTeam], bid), nil)
	} else {
		k.bidMade = false
		// **未達なら宣言額をそのままマイナス。**取った点は入らない。
		k.scores[declTeam] -= bid
		k.addLog(k.declarerIdx, "set", fmt.Sprintf("is set for %d", bid), nil)
	}

	// **45 点以上の側は自分がビッドしない限り加点できない。**
	if k.scores[defTeam] < KaiserMustBidThreshold {
		k.scores[defTeam] += k.handPoints[defTeam]
	} else {
		k.addLog(-1, "must_bid", fmt.Sprintf("team %d must bid to score from here", defTeam), nil)
	}

	k.phase = KaiserPhaseHandEnd
	k.checkGameEnd()
}

// checkGameEnd は目標点に届いていれば決着させる。
func (k *Kaiser) checkGameEnd() {
	a, b := k.scores[0], k.scores[1]
	if a < k.targetScore && b < k.targetScore {
		return
	}
	switch {
	case a >= k.targetScore && b >= k.targetScore:
		// 両方超えたら**落札側**が優先する。
		k.winnerTeam = KaiserTeamOf(k.declarerIdx)
	case a >= k.targetScore:
		k.winnerTeam = 0
	default:
		k.winnerTeam = 1
	}
	k.gameEndFlag = true
	k.phase = KaiserPhaseGameEnd
	k.addLog(-1, "game_end", fmt.Sprintf("team %d wins", k.winnerTeam), nil)
}

// NextHand は次の局を配る。
func (k *Kaiser) NextHand() error {
	if k.gameEndFlag {
		return fmt.Errorf("the game is over")
	}
	if k.phase != KaiserPhaseHandEnd {
		return fmt.Errorf("the hand is still in progress")
	}
	k.dealerIdx = (k.dealerIdx + 1) % KaiserPlayerCnt
	k.beginHand()
	return nil
}

// kaiserContains は s に v が含まれるかを返す。
func kaiserContains(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// ---- CPU ----

// KaiserCpuBid は CPU のビッドを決める。
//
// 一番長いスートの枚数と絵札から見込み点を粗く見積もり、最低ビッドに届く
// ときだけ宣言する。
func (k *Kaiser) KaiserCpuBid(idx int) (value int, contract KaiserContract, suit int) {
	p := k.GetPlayer(idx)
	if p == nil {
		return 0, KaiserContractTrump, 0
	}
	counts := map[int]int{}
	strength := map[int]int{}
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c == nil {
			continue
		}
		counts[c.GetDesign()]++
		// A と K を強さの目安にする。
		if c.GetValue() == 1 || c.GetValue() == 13 {
			strength[c.GetDesign()]++
		}
	}
	bestSuit, best := 0, -1
	for s := CardDesignSpade; s <= CardDesignDiamond; s++ {
		if v := counts[s]*2 + strength[s]; v > best {
			bestSuit, best = s, v
		}
	}
	// **♥5 を持っていれば 5 点が見込める。**弱い手でも宣言する価値が出る。
	estimate := counts[bestSuit] + strength[bestSuit]
	for i := range p.GetCardsSize() {
		if IsKaiserHeartFive(p.GetCard(i)) {
			estimate += KaiserHeartFiveBonus
		}
	}
	if estimate < KaiserMinBid {
		return 0, KaiserContractTrump, 0
	}
	if estimate > KaiserMaxBid {
		estimate = KaiserMaxBid
	}
	return estimate, KaiserContractTrump, bestSuit
}

// KaiserCpuDiscard は CPU が捨てる 2 枚を選ぶ。
//
// **♥5 と ♠3 は捨てられない。**切札以外の低い札から捨てる。
func (k *Kaiser) KaiserCpuDiscard(idx int) []int {
	p := k.GetPlayer(idx)
	if p == nil {
		return nil
	}
	type scored struct {
		idx, score int
	}
	var pool []scored
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c == nil || IsKaiserHeartFive(c) || IsKaiserSpadeThree(c) {
			continue
		}
		// 切札は残したいので大きく重み付けする。
		score := KaiserCardRank(c, k.contract)
		if k.trumpSuit != 0 && c.GetDesign() == k.trumpSuit {
			score += 100
		}
		pool = append(pool, scored{i, score})
	}
	for i := range pool {
		for j := i + 1; j < len(pool); j++ {
			if pool[j].score < pool[i].score {
				pool[i], pool[j] = pool[j], pool[i]
			}
		}
	}
	out := make([]int, 0, KaiserKittySize)
	for i := 0; i < len(pool) && i < KaiserKittySize; i++ {
		out = append(out, pool[i].idx)
	}
	return out
}

// KaiserCpuPlay は CPU が出す手札インデックスを返す。
func (k *Kaiser) KaiserCpuPlay(idx int) int {
	valid := k.KaiserValidPlays(idx)
	if len(valid) == 0 {
		return -1
	}
	p := k.GetPlayer(idx)
	if p == nil {
		return valid[0]
	}
	// **♠3 は押しつけたい。**出せるなら真っ先に出す。
	for _, i := range valid {
		if IsKaiserSpadeThree(p.GetCard(i)) && len(k.trick) > 0 {
			return i
		}
	}
	best, bestRank := valid[0], -1
	worst, worstRank := valid[0], 1<<30
	for _, i := range valid {
		r := KaiserCardRank(p.GetCard(i), k.contract)
		if r > bestRank {
			best, bestRank = i, r
		}
		if r < worstRank {
			worst, worstRank = i, r
		}
	}
	// リードは強い札、フォローは安い札。
	if len(k.trick) == 0 {
		return best
	}
	return worst
}

// IsHumanTurn は今が人間の手番かを返す。
func (k *Kaiser) IsHumanTurn() bool {
	if k.gameEndFlag {
		return false
	}
	switch k.phase {
	case KaiserPhaseBid:
		p := k.GetPlayer(k.bidIdx)
		return p != nil && p.GetIsHuman()
	case KaiserPhaseDiscard, KaiserPhasePlay:
		p := k.GetPlayer(k.currentIdx)
		return p != nil && p.GetIsHuman()
	}
	return false
}

// CpuPlay は今の手番の CPU に 1 手打たせる。
func (k *Kaiser) CpuPlay() {
	if k.gameEndFlag {
		return
	}
	switch k.phase {
	case KaiserPhaseBid:
		idx := k.bidIdx
		if p := k.GetPlayer(idx); p == nil || p.GetIsHuman() {
			return
		}
		value, contract, _ := k.KaiserCpuBid(idx)
		if value < KaiserMinBid {
			_ = k.PassBid(idx)
			return
		}
		if err := k.Bid(idx, value, contract); err != nil {
			_ = k.PassBid(idx)
		}
	case KaiserPhaseDiscard:
		idx := k.declarerIdx
		if p := k.GetPlayer(idx); p == nil || p.GetIsHuman() {
			return
		}
		if k.contract == KaiserContractTrump && k.trumpSuit == 0 {
			_, _, suit := k.KaiserCpuBid(idx)
			if suit == 0 {
				suit = CardDesignSpade
			}
			_ = k.SetTrump(idx, suit)
		}
		_ = k.Discard(idx, k.KaiserCpuDiscard(idx))
	case KaiserPhasePlay:
		idx := k.currentIdx
		if p := k.GetPlayer(idx); p == nil || p.GetIsHuman() {
			return
		}
		if i := k.KaiserCpuPlay(idx); i >= 0 {
			_ = k.PlayCard(idx, i)
		}
	}
}

// ---- 公開アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (k *Kaiser) GetPlayers() []*KaiserPlayer { return k.players }

// GetPlayer は idx のプレイヤーを返す。
func (k *Kaiser) GetPlayer(idx int) *KaiserPlayer {
	return getPlayer(k.players, idx)
}

// GetPhase は現在のフェーズを返す。
func (k *Kaiser) GetPhase() KaiserPhase { return k.phase }

// GetCurrentPlayerIdx は現在の手番を返す。
func (k *Kaiser) GetCurrentPlayerIdx() int { return k.currentIdx }

// GetBidPlayerIdx はビッド中の手番を返す。
func (k *Kaiser) GetBidPlayerIdx() int { return k.bidIdx }

// GetDealerIdx はディーラーを返す。
func (k *Kaiser) GetDealerIdx() int { return k.dealerIdx }

// GetBids はこの局のビッド履歴を返す。
func (k *Kaiser) GetBids() []*KaiserBid { return k.bids }

// GetHighBid は現在の最高ビッドを返す (未落札なら nil)。
func (k *Kaiser) GetHighBid() *KaiserBid { return k.highBid }

// GetDeclarerIdx は落札者を返す (-1 なら未確定)。
func (k *Kaiser) GetDeclarerIdx() int { return k.declarerIdx }

// GetTrumpSuit は切札を返す (0 ならノートランプ系)。
func (k *Kaiser) GetTrumpSuit() int { return k.trumpSuit }

// GetContract は契約種別を返す。
func (k *Kaiser) GetContract() KaiserContract { return k.contract }

// GetKittySize はキティの残り枚数を返す。
func (k *Kaiser) GetKittySize() int { return len(k.kitty) }

// GetTrick は場に出ている札を返す。
func (k *Kaiser) GetTrick() []*Card { return k.trick }

// GetTrickLeaderIdx はこのトリックのリード席を返す。
func (k *Kaiser) GetTrickLeaderIdx() int { return k.trickLeader }

// GetTrickNumber は済んだトリック数を返す。
func (k *Kaiser) GetTrickNumber() int { return k.trickNumber }

// GetHandPoints はチームがこの局で取った点を返す。
func (k *Kaiser) GetHandPoints(team int) int {
	if team < 0 || team >= KaiserTeamCnt {
		return 0
	}
	return k.handPoints[team]
}

// GetHeartFiveBy は ♥5 を取った席を返す (-1 なら未取得)。
func (k *Kaiser) GetHeartFiveBy() int { return k.heartFiveBy }

// GetSpadeThreeBy は ♠3 を取った席を返す (-1 なら未取得)。
func (k *Kaiser) GetSpadeThreeBy() int { return k.spadeThreeBy }

// IsBidMade は直前の局で落札側が達成したかを返す。
func (k *Kaiser) IsBidMade() bool { return k.bidMade }

// GetScore はチームの通算点を返す。
func (k *Kaiser) GetScore(team int) int {
	if team < 0 || team >= KaiserTeamCnt {
		return 0
	}
	return k.scores[team]
}

// GetTargetScore は現在の目標点を返す。
func (k *Kaiser) GetTargetScore() int { return k.targetScore }

// GetHandNumber は現在の局番号を返す。
func (k *Kaiser) GetHandNumber() int { return k.handNumber }

// GetGameEndFlag はゲーム終了フラグを返す。
func (k *Kaiser) GetGameEndFlag() bool { return k.gameEndFlag }

// GetWinnerTeam は勝利チームを返す (-1 なら未決)。
func (k *Kaiser) GetWinnerTeam() int { return k.winnerTeam }

// GetConfig は設定を返す。
func (k *Kaiser) GetConfig() KaiserConfig { return k.config }

// SetConfig は設定をセットする。
func (k *Kaiser) SetConfig(c KaiserConfig) { k.config = c }

// addLog は棋譜を 1 行足す。
func (k *Kaiser) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	k.appendLog(playerIdx, actionType, detail, cards)
}

// SetPhaseForTest はテスト用にフェーズを設定する。
func (k *Kaiser) SetPhaseForTest(p KaiserPhase) { k.phase = p }

// SetHandForTest はテスト用に手札を差し替える。
func (k *Kaiser) SetHandForTest(idx int, cards []*Card) {
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

// SetContractForTest はテスト用に契約を設定する。
func (k *Kaiser) SetContractForTest(declarer, value, suit int, contract KaiserContract) {
	k.declarerIdx = declarer
	k.highBid = &KaiserBid{Player: declarer, Value: value, Contract: contract}
	k.trumpSuit = suit
	k.contract = contract
}

// SetCurrentPlayerForTest はテスト用に手番を設定する。
func (k *Kaiser) SetCurrentPlayerForTest(idx int) { k.currentIdx = idx }

// SetTrickLeaderForTest はテスト用にリード席を設定する。
func (k *Kaiser) SetTrickLeaderForTest(idx int) { k.trickLeader = idx }

// SetHandPointsForTest はテスト用に局中の得点を設定する。
func (k *Kaiser) SetHandPointsForTest(team, pts int) {
	if team >= 0 && team < KaiserTeamCnt {
		k.handPoints[team] = pts
	}
}

// SetScoreForTest はテスト用に通算点を設定する。
func (k *Kaiser) SetScoreForTest(team, score int) {
	if team >= 0 && team < KaiserTeamCnt {
		k.scores[team] = score
	}
}

// SetTrickNumberForTest はテスト用に済んだトリック数を設定する。
func (k *Kaiser) SetTrickNumberForTest(n int) { k.trickNumber = n }

// FinishHandForTest はテスト用に精算を走らせる。
func (k *Kaiser) FinishHandForTest() { k.finishHand() }

// kaiserJSON is the JSON wire format for Kaiser.
type kaiserJSON struct {
	Players      []*KaiserPlayer    `json:"pl"`
	Config       KaiserConfig       `json:"cf"`
	Phase        KaiserPhase        `json:"ph"`
	Kitty        []*Card            `json:"ki"`
	DealerIdx    int                `json:"di"`
	CurrentIdx   int                `json:"ci"`
	BidIdx       int                `json:"bi"`
	Bids         []*KaiserBid       `json:"bd"`
	PassCount    int                `json:"pc"`
	HighBid      *KaiserBid         `json:"hb"`
	DeclarerIdx  int                `json:"de"`
	TrumpSuit    int                `json:"ts"`
	Contract     KaiserContract     `json:"co"`
	Trick        []*Card            `json:"tk"`
	TrickLeader  int                `json:"tl"`
	TrickNumber  int                `json:"tn"`
	HandPoints   [KaiserTeamCnt]int `json:"hp"`
	HeartFiveBy  int                `json:"h5"`
	SpadeThreeBy int                `json:"s3"`
	BidMade      bool               `json:"bm"`
	Scores       [KaiserTeamCnt]int `json:"sc"`
	TargetScore  int                `json:"tg"`
	HandNumber   int                `json:"hn"`
	GameEndFlag  bool               `json:"ge"`
	WinnerTeam   int                `json:"wt"`
	ActionLog    []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (k *Kaiser) MarshalJSON() ([]byte, error) {
	return json.Marshal(kaiserJSON{
		Players: k.players, Config: k.config, Phase: k.phase, Kitty: k.kitty,
		DealerIdx: k.dealerIdx, CurrentIdx: k.currentIdx, BidIdx: k.bidIdx,
		Bids: k.bids, PassCount: k.passCount, HighBid: k.highBid,
		DeclarerIdx: k.declarerIdx, TrumpSuit: k.trumpSuit, Contract: k.contract,
		Trick: k.trick, TrickLeader: k.trickLeader, TrickNumber: k.trickNumber,
		HandPoints: k.handPoints, HeartFiveBy: k.heartFiveBy, SpadeThreeBy: k.spadeThreeBy,
		BidMade: k.bidMade, Scores: k.scores, TargetScore: k.targetScore,
		HandNumber: k.handNumber, GameEndFlag: k.gameEndFlag, WinnerTeam: k.winnerTeam,
		ActionLog: k.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **復元でしか入らない値を弾く。**KV から戻ってきた壊れた状態でプレイが
// 詰まないよう、席番号・契約・スートを検証する。
func (k *Kaiser) UnmarshalJSON(data []byte) error {
	var j kaiserJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != KaiserPlayerCnt {
		return fmt.Errorf("bad player count: %d", len(j.Players))
	}
	if j.Phase < KaiserPhaseBid || j.Phase > KaiserPhaseGameEnd {
		return fmt.Errorf("bad phase: %d", j.Phase)
	}
	for name, v := range map[string]int{"dealer": j.DealerIdx, "current": j.CurrentIdx, "bid": j.BidIdx} {
		if v < 0 || v >= KaiserPlayerCnt {
			return fmt.Errorf("bad %s index: %d", name, v)
		}
	}
	for name, v := range map[string]int{
		"declarer": j.DeclarerIdx, "trick leader": j.TrickLeader,
		"heart five": j.HeartFiveBy, "spade three": j.SpadeThreeBy,
	} {
		if v < -1 || v >= KaiserPlayerCnt {
			return fmt.Errorf("bad %s index: %d", name, v)
		}
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= KaiserTeamCnt {
		return fmt.Errorf("bad winner team: %d", j.WinnerTeam)
	}
	if j.Contract < KaiserContractTrump || j.Contract > KaiserContractLowNoTrump {
		return fmt.Errorf("bad contract: %d", j.Contract)
	}
	// 0 は「ノートランプ」。それ以外はスートの範囲でなければならない。
	if j.TrumpSuit != 0 && (j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return fmt.Errorf("bad trump suit: %d", j.TrumpSuit)
	}
	if len(j.Trick) > KaiserPlayerCnt {
		return fmt.Errorf("bad trick size: %d", len(j.Trick))
	}
	if j.HighBid != nil && j.HighBid.Value != 0 &&
		(j.HighBid.Value < KaiserMinBid || j.HighBid.Value > KaiserMaxBid) {
		return fmt.Errorf("bad high bid: %d", j.HighBid.Value)
	}

	k.players = j.Players
	k.config = j.Config
	k.phase = j.Phase
	k.kitty = j.Kitty
	k.dealerIdx = j.DealerIdx
	k.currentIdx = j.CurrentIdx
	k.bidIdx = j.BidIdx
	k.bids = j.Bids
	k.passCount = j.PassCount
	k.highBid = j.HighBid
	k.declarerIdx = j.DeclarerIdx
	k.trumpSuit = j.TrumpSuit
	k.contract = j.Contract
	k.trick = j.Trick
	k.trickLeader = j.TrickLeader
	k.trickNumber = j.TrickNumber
	k.handPoints = j.HandPoints
	k.heartFiveBy = j.HeartFiveBy
	k.spadeThreeBy = j.SpadeThreeBy
	k.bidMade = j.BidMade
	k.scores = j.Scores
	k.targetScore = j.TargetScore
	if k.targetScore == 0 {
		k.targetScore = KaiserTargetScore
	}
	k.handNumber = j.HandNumber
	k.gameEndFlag = j.GameEndFlag
	k.winnerTeam = j.WinnerTeam
	k.actionLog = j.ActionLog
	return nil
}
