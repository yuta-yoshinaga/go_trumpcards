//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// IsraeliWhistPhase イスラエリホイストのゲームフェーズ
type IsraeliWhistPhase int

// IsraeliWhist のフェーズ定数
const (
	// IsraeliWhistPhaseAuction 1 段階目。切り札と最低ノルマを競り落とす。
	IsraeliWhistPhaseAuction IsraeliWhistPhase = iota
	// IsraeliWhistPhaseBid 2 段階目。全員が改めて目標トリック数を宣言する。
	IsraeliWhistPhaseBid
	// IsraeliWhistPhasePlay プレイ中
	IsraeliWhistPhasePlay
	// IsraeliWhistPhaseRoundEnd ラウンド終了
	IsraeliWhistPhaseRoundEnd
	// IsraeliWhistPhaseGameEnd ゲーム終了
	IsraeliWhistPhaseGameEnd
)

// IsraeliWhistPlayerCnt プレイヤー数（4 人固定・個人戦）
const IsraeliWhistPlayerCnt = 4

// IsraeliWhistHandSize 各プレイヤーの手札枚数
const IsraeliWhistHandSize = 13

// IsraeliWhistTricksPerRound 1 ラウンドのトリック数
const IsraeliWhistTricksPerRound = IsraeliWhistHandSize

// IsraeliWhistMinAuctionBid オークションの最低入札
const IsraeliWhistMinAuctionBid = 5

// IsraeliWhistZeroScore 0 を宣言して守り切ったときの得点
const IsraeliWhistZeroScore = 25

// IsraeliWhistExactBonus 的中の基礎点。宣言の 2 乗に乗る。
const IsraeliWhistExactBonus = 10

// IsraeliWhistMissUnit 外したとき、過不足 1 つあたりの減点
const IsraeliWhistMissUnit = 10

// israeliWhistMaxSliceLen caps slice sizes during deserialisation.
const israeliWhistMaxSliceLen = 1000

// IsraeliWhist イスラエリホイスト ゲームクラス。
//
// イスラエル発祥の競り型トリックテイキング。4 人個人戦、52 枚を 13 枚ずつ。
//
// **入札が 2 段階あるのがこのゲームの形。** 既存のどの実装にも無い:
//
//	1 段階目（オークション）: 「n トリック + 切り札スート」を競り上げる。最低は
//	  5。降りたら戻れない。勝った人の提示したスートがそのラウンドの切り札になり、
//	  その n が本人の *最低ノルマ* として残る。
//	2 段階目（宣言）: 落札者を含む **全員が改めて** 目標トリック数を宣言する。
//	  落札者はノルマ以上を宣言しなければならない。
//
// つまり落札者は「切り札を選ぶ権利」と「最低ノルマ」を同時に買う。ノルマだけ
// 決めて終わりではなく、そこから全員が別々の目標を持つのが 1 段階入札のゲーム
// （BidWhist など）との違い。
//
// **宣言の合計は 13 になってはいけない。** 13 だと全員が同時に宣言どおり取れて
// しまい賭けが成り立たないので、最後に宣言する席がそれを避ける義務を負う。
//
// **得点表は出典が割れるので、写さずに一貫した形を選んである:**
//
//	的中 n>=1 : +(n^2 + 10)  —— 大きく宣言して当てるほど跳ねる
//	的中 n==0 : +25          —— 13 枚持って 1 つも取らないのは別格の難しさ
//	外し      : -(10 x 過不足)
//	全員的中 / 全員外し: それぞれ 2 倍
//
// 最後の 2 つが issue の言う「全員的中/全員外れのボーナス倍率」にあたる。
type IsraeliWhist struct {
	trumpCards *TrumpCards
	players    []*IsraeliWhistPlayer
	config     IsraeliWhistConfig

	phase       IsraeliWhistPhase
	roundNumber int
	trickNumber int
	trumpSuit   int
	// declarerIdx はオークションの落札者 (-1: 未決定)
	declarerIdx int
	// highBid / highSuit は現在の最高入札
	highBid  int
	highSuit int

	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	dealerIdx        int
	// auctionPlayerIdx / bidPlayerIdx は 2 段階それぞれの手番
	auctionPlayerIdx int
	bidPlayerIdx     int

	gameEndFlag bool
	winnerIdx   int

	actionLogBase
}

// NewIsraeliWhist コンストラクタ
func NewIsraeliWhist(trumpCards *TrumpCards, players []*IsraeliWhistPlayer, config IsraeliWhistConfig) *IsraeliWhist {
	return &IsraeliWhist{trumpCards: trumpCards, players: players, config: config, declarerIdx: -1, winnerIdx: -1}
}

// NewDefaultIsraeliWhist 既定構成（人間 1 + CPU 3）のコンストラクタ
func NewDefaultIsraeliWhist() *IsraeliWhist {
	players := make([]*IsraeliWhistPlayer, 0, IsraeliWhistPlayerCnt)
	for i := range IsraeliWhistPlayerCnt {
		players = append(players, NewIsraeliWhistPlayer(i == 0))
	}
	return NewIsraeliWhist(NewTrumpCards(0), players, DefaultIsraeliWhistConfig())
}

// israeliWhistSuitRank 入札でのスートの序列。♣ < ♦ < ♥ < ♠。
//
// **CardDesign の並びとは違う。** 定数は ♠1 ♣2 ♥3 ♦4 なので、そのまま比べると
// ♦ が最強になってしまう。
func israeliWhistSuitRank(suit int) int {
	switch suit {
	case CardDesignClover:
		return 1
	case CardDesignDiamond:
		return 2
	case CardDesignHeart:
		return 3
	case CardDesignSpade:
		return 4
	}
	return 0
}

// Reset ゲーム全体を初期化する
func (w *IsraeliWhist) Reset() {
	w.roundNumber = 1
	w.dealerIdx = 0
	w.gameEndFlag = false
	w.winnerIdx = -1
	w.actionLog = nil
	for _, p := range w.players {
		p.ResetGame()
	}
	w.dealRound()
}

// dealRound 13 枚ずつ配り、オークションから始める
func (w *IsraeliWhist) dealRound() {
	w.phase = IsraeliWhistPhaseAuction
	w.trickNumber = 0
	w.currentTrick = nil
	w.trumpSuit = 0
	w.declarerIdx = -1
	w.highBid = 0
	w.highSuit = 0
	for _, p := range w.players {
		p.ResetRound()
	}

	w.trumpCards = NewTrumpCards(0)
	w.trumpCards.Shuffle()
	for range IsraeliWhistHandSize {
		for i := range IsraeliWhistPlayerCnt {
			idx := (w.dealerIdx + 1 + i) % IsraeliWhistPlayerCnt
			if c := w.trumpCards.DrawCard(); c != nil {
				w.players[idx].AddCard(c)
			}
		}
	}
	w.sortAllHands()
	w.auctionPlayerIdx = (w.dealerIdx + 1) % IsraeliWhistPlayerCnt
	w.bidPlayerIdx = w.auctionPlayerIdx
	w.currentPlayerIdx = w.auctionPlayerIdx
	w.leadPlayerIdx = w.auctionPlayerIdx
	w.appendLog(-1, "deal", fmt.Sprintf("ラウンド%d を開始", w.roundNumber), nil)
}

// sortAllHands 手札をスート・ランク順に並べ替える
func (w *IsraeliWhist) sortAllHands() {
	for _, p := range w.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return israeliWhistRank(ci) < israeliWhistRank(cj)
		})
	}
}

// israeliWhistRank 札の強さ。A が最強、以下 K,Q,J,10..2。
func israeliWhistRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetValue() == 1 {
		return CardValueMax + 1
	}
	return c.GetValue()
}

// --- 1 段階目: オークション ---

// PlayerAuctionBid 人間プレイヤーがオークションで入札する
func (w *IsraeliWhist) PlayerAuctionBid(bid, suit int) error {
	if err := w.guardAuction(0); err != nil {
		return err
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return fmt.Errorf("invalid trump suit: %d", suit)
	}
	if bid < IsraeliWhistMinAuctionBid || bid > IsraeliWhistHandSize {
		return fmt.Errorf("auction bids run from %d to %d", IsraeliWhistMinAuctionBid, IsraeliWhistHandSize)
	}
	if !w.outbids(bid, suit) {
		return fmt.Errorf("bid must beat %d %d", w.highBid, w.highSuit)
	}
	w.acceptAuctionBid(0, bid, suit)
	return nil
}

// PlayerAuctionPass 人間プレイヤーがオークションを降りる
func (w *IsraeliWhist) PlayerAuctionPass() error {
	if err := w.guardAuction(0); err != nil {
		return err
	}
	// **全員が降りると切り札が決まらない。** 最後の 1 人は降りられない。
	if w.highBid == 0 && w.activeBidders() == 1 {
		return errors.New("the last bidder standing must bid")
	}
	w.acceptAuctionPass(0)
	return nil
}

// guardAuction オークションで操作できる状態かを確かめる
func (w *IsraeliWhist) guardAuction(idx int) error {
	if w.gameEndFlag {
		return errors.New("game has ended")
	}
	if w.phase != IsraeliWhistPhaseAuction {
		return errors.New("not the auction phase")
	}
	if w.auctionPlayerIdx != idx {
		return errors.New("not your turn to bid")
	}
	if w.players[idx].GetPassed() {
		return errors.New("you have already passed")
	}
	return nil
}

// outbids その入札が現在の最高入札を上回るか。同数ならスートの序列で決める。
func (w *IsraeliWhist) outbids(bid, suit int) bool {
	if bid != w.highBid {
		return bid > w.highBid
	}
	return israeliWhistSuitRank(suit) > israeliWhistSuitRank(w.highSuit)
}

// activeBidders まだ降りていない人数
func (w *IsraeliWhist) activeBidders() int {
	n := 0
	for _, p := range w.players {
		if !p.GetPassed() {
			n++
		}
	}
	return n
}

// CpuAuction 手番の CPU が入札するか降りるかを決める
func (w *IsraeliWhist) CpuAuction() {
	if w.gameEndFlag || w.phase != IsraeliWhistPhaseAuction || w.auctionPlayerIdx == 0 {
		return
	}
	idx := w.auctionPlayerIdx
	bid, suit := w.cpuAuctionChoice(idx)
	// **降りられるのは、他に生きた入札者が残っているときだけ。**
	if bid > 0 && w.outbids(bid, suit) {
		w.acceptAuctionBid(idx, bid, suit)
		return
	}
	if w.highBid == 0 && w.activeBidders() == 1 {
		// 誰も入札していないのに最後の 1 人まで来た。引き受けるしかない。
		forced, forcedSuit := IsraeliWhistMinAuctionBid, w.longestSuit(idx)
		w.acceptAuctionBid(idx, forced, forcedSuit)
		return
	}
	w.acceptAuctionPass(idx)
}

// acceptAuctionBid 入札を記録して次の席へ回す
func (w *IsraeliWhist) acceptAuctionBid(idx, bid, suit int) {
	w.players[idx].SetAuction(bid, suit)
	w.highBid, w.highSuit, w.declarerIdx = bid, suit, idx
	w.appendLog(idx, "auction", fmt.Sprintf("%d トリック・切り札 %d で入札", bid, suit), nil)
	w.advanceAuction()
}

// acceptAuctionPass 降りたことを記録して次の席へ回す
func (w *IsraeliWhist) acceptAuctionPass(idx int) {
	w.players[idx].SetPassed(true)
	w.appendLog(idx, "pass", "オークションを降りた", nil)
	w.advanceAuction()
}

// advanceAuction 次の入札者へ回し、決着していればオークションを閉じる
func (w *IsraeliWhist) advanceAuction() {
	// **入札のある状態で生存者が 1 人になったら決着。** 生存者 = 落札者。
	if w.highBid > 0 && w.activeBidders() <= 1 {
		w.closeAuction()
		return
	}
	for i := 1; i <= IsraeliWhistPlayerCnt; i++ {
		next := (w.auctionPlayerIdx + i) % IsraeliWhistPlayerCnt
		if !w.players[next].GetPassed() {
			w.auctionPlayerIdx = next
			return
		}
	}
	w.closeAuction()
}

// closeAuction 切り札を確定させ、2 段階目の宣言に入る
func (w *IsraeliWhist) closeAuction() {
	w.trumpSuit = w.highSuit
	w.phase = IsraeliWhistPhaseBid
	// 宣言は落札者から始まる。落札者は自分のノルマ以上を宣言する義務がある。
	w.bidPlayerIdx = w.declarerIdx
	w.appendLog(w.declarerIdx, "trump",
		fmt.Sprintf("切り札 %d・最低ノルマ %d で落札", w.trumpSuit, w.highBid), nil)
}

// --- 2 段階目: 宣言 ---

// PlayerBid 人間プレイヤーが目標トリック数を宣言する
func (w *IsraeliWhist) PlayerBid(bid int) error {
	if w.gameEndFlag {
		return errors.New("game has ended")
	}
	if w.phase != IsraeliWhistPhaseBid {
		return errors.New("not the bidding phase")
	}
	if w.bidPlayerIdx != 0 {
		return errors.New("not your turn to bid")
	}
	if bid < 0 || bid > IsraeliWhistHandSize {
		return fmt.Errorf("invalid bid: %d", bid)
	}
	if m := w.MinimumBidFor(0); bid < m {
		return fmt.Errorf("as the declarer you must bid at least %d", m)
	}
	if r := w.GetRestrictedBid(); r >= 0 && bid == r {
		return fmt.Errorf("the last bidder cannot make the total %d", IsraeliWhistHandSize)
	}
	w.acceptBid(0, bid)
	return nil
}

// MinimumBidFor そのプレイヤーが宣言できる下限。落札者だけノルマが効く。
func (w *IsraeliWhist) MinimumBidFor(idx int) int {
	if idx == w.declarerIdx {
		return w.highBid
	}
	return 0
}

// GetRestrictedBid 最後の宣言者が選べない宣言値を返す (-1 = 制限なし)
func (w *IsraeliWhist) GetRestrictedBid() int {
	if w.phase != IsraeliWhistPhaseBid || w.bidsRemaining() != 1 {
		return -1
	}
	total := 0
	for _, p := range w.players {
		if p.GetBid() >= 0 {
			total += p.GetBid()
		}
	}
	restricted := IsraeliWhistHandSize - total
	if restricted < 0 || restricted > IsraeliWhistHandSize {
		return -1
	}
	return restricted
}

// bidsRemaining まだ宣言していない人数
func (w *IsraeliWhist) bidsRemaining() int {
	n := 0
	for _, p := range w.players {
		if p.GetBid() < 0 {
			n++
		}
	}
	return n
}

// CpuBid 手番の CPU が宣言する
func (w *IsraeliWhist) CpuBid() {
	if w.gameEndFlag || w.phase != IsraeliWhistPhaseBid || w.bidPlayerIdx == 0 {
		return
	}
	idx := w.bidPlayerIdx
	bid := w.estimateTricks(idx)
	if m := w.MinimumBidFor(idx); bid < m {
		bid = m
	}
	// **最後の宣言者は合計 13 を避ける義務がある。** 1 つずらして回避する。
	if r := w.GetRestrictedBid(); r >= 0 && bid == r {
		if bid > w.MinimumBidFor(idx) {
			bid--
		} else {
			bid++
		}
	}
	w.acceptBid(idx, bid)
}

// acceptBid 宣言を記録し、次の席へ回す
func (w *IsraeliWhist) acceptBid(idx, bid int) {
	w.players[idx].SetBid(bid)
	w.appendLog(idx, "bid", fmt.Sprintf("%d トリックを宣言", bid), nil)

	// **残りを数えるのは SetBid の後。** `> 1` にすると最後の 1 人が宣言せず、
	// 合計 13 禁止の制約が掛かる席がそもそも来なくなる。
	if w.bidsRemaining() > 0 {
		for i := 1; i <= IsraeliWhistPlayerCnt; i++ {
			next := (idx + i) % IsraeliWhistPlayerCnt
			if w.players[next].GetBid() < 0 {
				w.bidPlayerIdx = next
				break
			}
		}
		return
	}
	w.phase = IsraeliWhistPhasePlay
	// リードは落札者から。
	w.leadPlayerIdx = w.declarerIdx
	w.currentPlayerIdx = w.declarerIdx
}

// longestSuit いちばん枚数の多いスート
func (w *IsraeliWhist) longestSuit(idx int) int {
	p := w.players[idx]
	counts := map[int]int{}
	for i := range p.GetCardsSize() {
		counts[p.GetCard(i).GetDesign()]++
	}
	best, bestN := CardDesignSpade, -1
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if counts[suit] > bestN {
			best, bestN = suit, counts[suit]
		}
	}
	return best
}

// cpuAuctionChoice CPU のオークション判断。長いスートの強さで測る。
func (w *IsraeliWhist) cpuAuctionChoice(idx int) (int, int) {
	suit := w.longestSuit(idx)
	p := w.players[idx]
	score := 0.0
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		isTrump := c.GetDesign() == suit
		switch {
		case c.GetValue() == 1:
			score += 1.0
		case c.GetValue() == 13:
			score += 0.7
		case c.GetValue() == 12:
			score += 0.4
		case isTrump:
			score += 0.35
		}
	}
	bid := int(score)
	if bid < IsraeliWhistMinAuctionBid {
		return 0, 0
	}
	if bid > IsraeliWhistHandSize {
		bid = IsraeliWhistHandSize
	}
	return bid, suit
}

// estimateTricks CPU の宣言。A/K/切り札の枚数から取れそうな数を見積もる。
func (w *IsraeliWhist) estimateTricks(idx int) int {
	p := w.players[idx]
	score := 0.0
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		isTrump := c.GetDesign() == w.trumpSuit
		switch {
		case c.GetValue() == 1:
			score += 1.0
		case c.GetValue() == 13:
			score += 0.7
		case c.GetValue() == 12:
			score += 0.4
		case isTrump:
			score += 0.3
		}
	}
	bid := int(score)
	if bid > IsraeliWhistHandSize {
		bid = IsraeliWhistHandSize
	}
	if bid < 0 {
		bid = 0
	}
	return bid
}

// --- プレイ ---

// PlayerPlay 人間プレイヤーが手札の cardIndex を出す
func (w *IsraeliWhist) PlayerPlay(cardIndex int) error {
	if w.gameEndFlag {
		return errors.New("game has ended")
	}
	if w.phase != IsraeliWhistPhasePlay {
		return errors.New("not the play phase")
	}
	if w.currentPlayerIdx != 0 {
		return errors.New("not your turn")
	}
	return w.play(0, cardIndex)
}

// CpuPlay CPU が 1 枚出す
func (w *IsraeliWhist) CpuPlay() {
	if w.gameEndFlag || w.phase != IsraeliWhistPhasePlay || w.currentPlayerIdx == 0 {
		return
	}
	_ = w.play(w.currentPlayerIdx, w.chooseCpuCard(w.currentPlayerIdx))
}

// play 指定プレイヤーが 1 枚出す
func (w *IsraeliWhist) play(playerIdx, cardIndex int) error {
	p := w.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	card := p.GetCard(cardIndex)
	if !w.canPlay(playerIdx, card) {
		return errors.New("must follow suit")
	}
	p.RemoveCard(cardIndex)
	w.currentTrick = append(w.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	w.appendLog(playerIdx, "play", cardStr(card), []*Card{card})

	if len(w.currentTrick) < IsraeliWhistPlayerCnt {
		w.currentPlayerIdx = (playerIdx + 1) % IsraeliWhistPlayerCnt
		return nil
	}
	w.resolveTrick()
	return nil
}

// canPlay フォロー義務を満たすか
func (w *IsraeliWhist) canPlay(playerIdx int, card *Card) bool {
	if len(w.currentTrick) == 0 {
		return true
	}
	leadSuit := w.currentTrick[0].Card.GetDesign()
	if card.GetDesign() == leadSuit {
		return true
	}
	p := w.players[playerIdx]
	for i := range p.GetCardsSize() {
		if p.GetCard(i).GetDesign() == leadSuit {
			return false
		}
	}
	return true
}

// GetValidPlayIndices 出せる手札のインデックスを返す
func (w *IsraeliWhist) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(w.players) {
		return nil
	}
	p := w.players[playerIdx]
	valid := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if w.canPlay(playerIdx, p.GetCard(i)) {
			valid = append(valid, i)
		}
	}
	return valid
}

// resolveTrick トリックを解決する
func (w *IsraeliWhist) resolveTrick() {
	winner := w.trickWinner()
	cards := make([]*Card, 0, len(w.currentTrick))
	for _, tc := range w.currentTrick {
		cards = append(cards, tc.Card)
	}
	w.players[winner].AddTrick(cards)

	w.trickNumber++
	w.currentTrick = nil
	w.leadPlayerIdx = winner
	w.currentPlayerIdx = winner

	if w.trickNumber >= IsraeliWhistTricksPerRound {
		w.finishRound()
	}
}

// trickWinner 現在のトリックの勝者
func (w *IsraeliWhist) trickWinner() int {
	if len(w.currentTrick) == 0 {
		return w.leadPlayerIdx
	}
	leadSuit := w.currentTrick[0].Card.GetDesign()
	bestIdx, best := w.currentTrick[0].PlayerIdx, w.currentTrick[0].Card
	for _, tc := range w.currentTrick[1:] {
		if w.beats(tc.Card, best, leadSuit) {
			best, bestIdx = tc.Card, tc.PlayerIdx
		}
	}
	return bestIdx
}

// beats challenger が currentBest に勝つか
func (w *IsraeliWhist) beats(challenger, currentBest *Card, leadSuit int) bool {
	cTrump := challenger.GetDesign() == w.trumpSuit
	bTrump := currentBest.GetDesign() == w.trumpSuit
	if cTrump != bTrump {
		return cTrump
	}
	if challenger.GetDesign() != currentBest.GetDesign() {
		return challenger.GetDesign() == leadSuit
	}
	return israeliWhistRank(challenger) > israeliWhistRank(currentBest)
}

// finishRound 宣言の当否で得点を確定させる
func (w *IsraeliWhist) finishRound() {
	exact := 0
	for _, p := range w.players {
		if p.GetTrickCount() == p.GetBid() {
			exact++
		}
	}
	// **全員的中と全員外しはどちらも 2 倍。** 4 人が同じ結末になるのは珍しく、
	// そこだけ跳ねるのがこのゲームの起伏。
	doubled := exact == IsraeliWhistPlayerCnt || exact == 0

	for i, p := range w.players {
		score := IsraeliWhistScoreFor(p.GetBid(), p.GetTrickCount(), doubled)
		p.SetRoundScore(score)
		p.AddTotalScore(score)
		w.appendLog(i, "score", fmt.Sprintf("宣言%d 獲得%d: %+d", p.GetBid(), p.GetTrickCount(), score), nil)
	}
	if doubled {
		if exact == IsraeliWhistPlayerCnt {
			w.appendLog(-1, "bonus", "全員が的中。得点は 2 倍", nil)
		} else {
			w.appendLog(-1, "bonus", "全員が外した。減点は 2 倍", nil)
		}
	}

	if w.roundNumber >= w.config.Rounds {
		w.finishGame()
		return
	}
	w.phase = IsraeliWhistPhaseRoundEnd
}

// IsraeliWhistScoreFor 宣言 bid・獲得 got の増減を返す。doubled なら 2 倍。
//
//	的中 n>=1 : +(n^2 + 10)
//	的中 n==0 : +25
//	外し      : -(10 x 過不足)
func IsraeliWhistScoreFor(bid, got int, doubled bool) int {
	var score int
	switch {
	case got != bid:
		diff := got - bid
		if diff < 0 {
			diff = -diff
		}
		score = -IsraeliWhistMissUnit * diff
	case bid == 0:
		score = IsraeliWhistZeroScore
	default:
		score = bid*bid + IsraeliWhistExactBonus
	}
	if doubled {
		score *= 2
	}
	return score
}

// NextRound 次のラウンドを開始する
func (w *IsraeliWhist) NextRound() {
	if w.gameEndFlag || w.phase != IsraeliWhistPhaseRoundEnd {
		return
	}
	w.roundNumber++
	w.dealerIdx = (w.dealerIdx + 1) % IsraeliWhistPlayerCnt
	w.dealRound()
}

// finishGame 累計得点の最も高いプレイヤーの勝ち
func (w *IsraeliWhist) finishGame() {
	w.phase = IsraeliWhistPhaseGameEnd
	w.gameEndFlag = true
	bestIdx, best, tied := -1, 0, false
	for i, p := range w.players {
		switch {
		case bestIdx < 0 || p.GetTotalScore() > best:
			bestIdx, best, tied = i, p.GetTotalScore(), false
		case p.GetTotalScore() == best:
			tied = true
		}
	}
	if tied {
		w.winnerIdx = -1
	} else {
		w.winnerIdx = bestIdx
	}
	w.appendLog(-1, "result", fmt.Sprintf("最終得点 %d/%d/%d/%d",
		w.players[0].GetTotalScore(), w.players[1].GetTotalScore(),
		w.players[2].GetTotalScore(), w.players[3].GetTotalScore()), nil)
}

// chooseCpuCard CPU の手。宣言に足りなければ取りに行き、足りていれば逃げる。
func (w *IsraeliWhist) chooseCpuCard(playerIdx int) int {
	valid := w.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := w.players[playerIdx]
	wantsMore := p.GetTrickCount() < p.GetBid()

	if len(w.currentTrick) == 0 {
		return w.pickExtreme(p, valid, wantsMore)
	}
	if wantsMore {
		if idx, ok := w.pickCheapestWinner(p, valid); ok {
			return idx
		}
		return w.pickExtreme(p, valid, false)
	}
	bestIdx, bestRank := -1, -1
	for _, i := range valid {
		c := p.GetCard(i)
		if w.wouldWin(c) {
			continue
		}
		if r := israeliWhistRank(c); r > bestRank {
			bestIdx, bestRank = i, r
		}
	}
	if bestIdx >= 0 {
		return bestIdx
	}
	return w.pickExtreme(p, valid, false)
}

// pickExtreme valid のうち最強 (high) または最弱の札を選ぶ
func (w *IsraeliWhist) pickExtreme(p *IsraeliWhistPlayer, valid []int, high bool) int {
	bestIdx, bestRank := valid[0], israeliWhistRank(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		r := israeliWhistRank(p.GetCard(i))
		if (high && r > bestRank) || (!high && r < bestRank) {
			bestIdx, bestRank = i, r
		}
	}
	return bestIdx
}

// pickCheapestWinner トリックを取れる札のうち一番弱いもの
func (w *IsraeliWhist) pickCheapestWinner(p *IsraeliWhistPlayer, valid []int) (int, bool) {
	bestIdx, bestRank := -1, 0
	for _, i := range valid {
		c := p.GetCard(i)
		if !w.wouldWin(c) {
			continue
		}
		if r := israeliWhistRank(c); bestIdx < 0 || r < bestRank {
			bestIdx, bestRank = i, r
		}
	}
	return bestIdx, bestIdx >= 0
}

// wouldWin その札を今出したらトリックを取ってしまうか
func (w *IsraeliWhist) wouldWin(c *Card) bool {
	if c == nil || len(w.currentTrick) == 0 {
		return true
	}
	leadSuit := w.currentTrick[0].Card.GetDesign()
	best := w.currentTrick[0].Card
	for _, tc := range w.currentTrick[1:] {
		if w.beats(tc.Card, best, leadSuit) {
			best = tc.Card
		}
	}
	return w.beats(c, best, leadSuit)
}

// IsraeliWhistHint ヒント情報
type IsraeliWhistHint struct {
	// CardIndex 推奨する手札のインデックス（オークション・宣言中は nil）
	CardIndex *int
	// Reason ヒント理由キー
	Reason string
	// Value 推奨する入札数、または宣言数
	Value int
	// Suit オークションで勧める切り札スート（それ以外は 0）
	Suit int
}

// GetHint 人間プレイヤーへの推奨手を返す
func (w *IsraeliWhist) GetHint() *IsraeliWhistHint {
	if w.gameEndFlag {
		return nil
	}
	if w.phase == IsraeliWhistPhaseAuction && w.auctionPlayerIdx == 0 && !w.players[0].GetPassed() {
		bid, suit := w.cpuAuctionChoice(0)
		if bid > 0 && w.outbids(bid, suit) {
			return &IsraeliWhistHint{Reason: "israeliwhistAuctionBid", Value: bid, Suit: suit}
		}
		return &IsraeliWhistHint{Reason: "israeliwhistAuctionPass"}
	}
	if w.phase == IsraeliWhistPhaseBid && w.bidPlayerIdx == 0 {
		bid := w.estimateTricks(0)
		if m := w.MinimumBidFor(0); bid < m {
			return &IsraeliWhistHint{Reason: "israeliwhistMeetQuota", Value: m}
		}
		if r := w.GetRestrictedBid(); r >= 0 && bid == r {
			if bid > w.MinimumBidFor(0) {
				bid--
			} else {
				bid++
			}
			return &IsraeliWhistHint{Reason: "israeliwhistAvoidRestricted", Value: bid}
		}
		return &IsraeliWhistHint{Reason: "israeliwhistBid", Value: bid}
	}
	if !w.IsHumanTurn() || w.players[0].GetCardsSize() == 0 {
		return nil
	}
	idx := w.chooseCpuCard(0)
	reason := "israeliwhistDuck"
	if w.players[0].GetTrickCount() < w.players[0].GetBid() {
		reason = "israeliwhistWinTrick"
	}
	return &IsraeliWhistHint{CardIndex: &idx, Reason: reason}
}

// --- Getters ---

// GetPhase 現在のフェーズ
func (w *IsraeliWhist) GetPhase() IsraeliWhistPhase { return w.phase }

// GetConfig 現在の設定
func (w *IsraeliWhist) GetConfig() IsraeliWhistConfig { return w.config }

// SetConfig 設定を差し替える
func (w *IsraeliWhist) SetConfig(c IsraeliWhistConfig) { w.config = c }

// GetRoundNumber 現在のラウンド番号（1 起点）
func (w *IsraeliWhist) GetRoundNumber() int { return w.roundNumber }

// GetTrickNumber 現在のトリック番号（0 起点）
func (w *IsraeliWhist) GetTrickNumber() int { return w.trickNumber }

// GetTrumpSuit 切り札のスート（オークション中は 0）
func (w *IsraeliWhist) GetTrumpSuit() int { return w.trumpSuit }

// GetDeclarerIdx オークションの落札者 (-1: 未決定)
func (w *IsraeliWhist) GetDeclarerIdx() int { return w.declarerIdx }

// GetHighBid 現在の最高入札のトリック数（未入札は 0）
func (w *IsraeliWhist) GetHighBid() int { return w.highBid }

// GetHighSuit 現在の最高入札のスート（未入札は 0）
func (w *IsraeliWhist) GetHighSuit() int { return w.highSuit }

// GetCurrentTrick 現在のトリック
func (w *IsraeliWhist) GetCurrentTrick() []*TrickCard { return w.currentTrick }

// GetCurrentPlayerIdx 現在の手番
func (w *IsraeliWhist) GetCurrentPlayerIdx() int { return w.currentPlayerIdx }

// GetAuctionPlayerIdx オークションの手番
func (w *IsraeliWhist) GetAuctionPlayerIdx() int { return w.auctionPlayerIdx }

// GetBidPlayerIdx 宣言の手番
func (w *IsraeliWhist) GetBidPlayerIdx() int { return w.bidPlayerIdx }

// GetLeadPlayerIdx リードプレイヤー
func (w *IsraeliWhist) GetLeadPlayerIdx() int { return w.leadPlayerIdx }

// GetDealerIdx ディーラー
func (w *IsraeliWhist) GetDealerIdx() int { return w.dealerIdx }

// GetPlayerCnt プレイヤー数
func (w *IsraeliWhist) GetPlayerCnt() int { return len(w.players) }

// GetPlayer 指定インデックスのプレイヤー
func (w *IsraeliWhist) GetPlayer(i int) *IsraeliWhistPlayer {
	if i < 0 || i >= len(w.players) {
		return nil
	}
	return w.players[i]
}

// GetGameEndFlag ゲーム終了フラグ
func (w *IsraeliWhist) GetGameEndFlag() bool { return w.gameEndFlag }

// GetWinnerIdx 勝利プレイヤー (-1: 未確定/同点)
func (w *IsraeliWhist) GetWinnerIdx() int { return w.winnerIdx }

// IsHumanTurn 人間の手番か
func (w *IsraeliWhist) IsHumanTurn() bool {
	return !w.gameEndFlag && w.phase == IsraeliWhistPhasePlay && w.currentPlayerIdx == 0
}

// IsHumanAuctionTurn 人間がオークションで判断する番か
func (w *IsraeliWhist) IsHumanAuctionTurn() bool {
	return !w.gameEndFlag && w.phase == IsraeliWhistPhaseAuction &&
		w.auctionPlayerIdx == 0 && !w.players[0].GetPassed()
}

// IsHumanBidTurn 人間が宣言する番か
func (w *IsraeliWhist) IsHumanBidTurn() bool {
	return !w.gameEndFlag && w.phase == IsraeliWhistPhaseBid && w.bidPlayerIdx == 0
}

// GiveUp 投了する
func (w *IsraeliWhist) GiveUp() {
	if w.gameEndFlag {
		return
	}
	w.phase = IsraeliWhistPhaseGameEnd
	w.gameEndFlag = true
	w.winnerIdx = 1
	w.appendLog(0, "giveup", "ギブアップしました", nil)
}

// appendLog 棋譜エントリを追加
func (w *IsraeliWhist) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	w.appendLogAt(w.trickNumber, playerIdx, actionType, detail, cards)
}

// israeliWhistJSON is the KV snapshot format for IsraeliWhist.
type israeliWhistJSON struct {
	TrumpCards       *TrumpCards           `json:"tc"`
	Players          []*IsraeliWhistPlayer `json:"pl"`
	Config           IsraeliWhistConfig    `json:"cf"`
	Phase            IsraeliWhistPhase     `json:"ph"`
	RoundNumber      int                   `json:"rn"`
	TrickNumber      int                   `json:"tn"`
	TrumpSuit        int                   `json:"ts"`
	DeclarerIdx      int                   `json:"dc"`
	HighBid          int                   `json:"hb"`
	HighSuit         int                   `json:"hs"`
	CurrentTrick     []*TrickCard          `json:"ct"`
	CurrentPlayerIdx int                   `json:"cp"`
	AuctionPlayerIdx int                   `json:"ap"`
	BidPlayerIdx     int                   `json:"bp"`
	LeadPlayerIdx    int                   `json:"lp"`
	DealerIdx        int                   `json:"di"`
	GameEndFlag      bool                  `json:"ge"`
	WinnerIdx        int                   `json:"wi"`
	ActionLog        []*ActionLogEntry     `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (w *IsraeliWhist) MarshalJSON() ([]byte, error) {
	return json.Marshal(&israeliWhistJSON{
		TrumpCards:       w.trumpCards,
		Players:          w.players,
		Config:           w.config,
		Phase:            w.phase,
		RoundNumber:      w.roundNumber,
		TrickNumber:      w.trickNumber,
		TrumpSuit:        w.trumpSuit,
		DeclarerIdx:      w.declarerIdx,
		HighBid:          w.highBid,
		HighSuit:         w.highSuit,
		CurrentTrick:     w.currentTrick,
		CurrentPlayerIdx: w.currentPlayerIdx,
		AuctionPlayerIdx: w.auctionPlayerIdx,
		BidPlayerIdx:     w.bidPlayerIdx,
		LeadPlayerIdx:    w.leadPlayerIdx,
		DealerIdx:        w.dealerIdx,
		GameEndFlag:      w.gameEndFlag,
		WinnerIdx:        w.winnerIdx,
		ActionLog:        w.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元。値域を検証する。
func (w *IsraeliWhist) UnmarshalJSON(data []byte) error {
	var j israeliWhistJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < IsraeliWhistPhaseAuction || j.Phase > IsraeliWhistPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	// **切り札はフェーズと整合していなければならない。** オークション中は
	// まだ 0、決まったあとは実在するスート。壊れた値を通すと、切り札の無い
	// ラウンドや、誰も選んでいないスートが切り札のラウンドが復元される。
	if j.Phase == IsraeliWhistPhaseAuction {
		if j.TrumpSuit != 0 {
			return fmt.Errorf("trump suit %d before the auction closed", j.TrumpSuit)
		}
	} else if j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond {
		return fmt.Errorf("invalid trump suit: %d", j.TrumpSuit)
	}
	if j.TrickNumber < 0 || j.TrickNumber > IsraeliWhistTricksPerRound {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if j.RoundNumber < 1 {
		return fmt.Errorf("invalid round number: %d", j.RoundNumber)
	}
	if j.HighBid < 0 || j.HighBid > IsraeliWhistHandSize {
		return fmt.Errorf("invalid high bid: %d", j.HighBid)
	}
	if len(j.ActionLog) > israeliWhistMaxSliceLen {
		return errors.New("israeliwhist: input array exceeds maximum allowed size")
	}
	if len(j.CurrentTrick) > IsraeliWhistPlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"auction player": j.AuctionPlayerIdx,
		"bid player":     j.BidPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"dealer":         j.DealerIdx,
	} {
		if idx < 0 || idx >= IsraeliWhistPlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	if j.DeclarerIdx < -1 || j.DeclarerIdx >= IsraeliWhistPlayerCnt {
		return fmt.Errorf("invalid declarer: %d", j.DeclarerIdx)
	}
	if j.WinnerIdx < -1 || j.WinnerIdx >= IsraeliWhistPlayerCnt {
		return fmt.Errorf("invalid winner: %d", j.WinnerIdx)
	}
	if j.TrumpCards != nil {
		w.trumpCards = j.TrumpCards
	}
	if len(j.Players) == IsraeliWhistPlayerCnt {
		w.players = j.Players
	}
	w.config = j.Config
	w.phase = j.Phase
	w.roundNumber = j.RoundNumber
	w.trickNumber = j.TrickNumber
	w.trumpSuit = j.TrumpSuit
	w.declarerIdx = j.DeclarerIdx
	w.highBid = j.HighBid
	w.highSuit = j.HighSuit
	w.currentTrick = j.CurrentTrick
	w.currentPlayerIdx = j.CurrentPlayerIdx
	w.auctionPlayerIdx = j.AuctionPlayerIdx
	w.bidPlayerIdx = j.BidPlayerIdx
	w.leadPlayerIdx = j.LeadPlayerIdx
	w.dealerIdx = j.DealerIdx
	w.gameEndFlag = j.GameEndFlag
	w.winnerIdx = j.WinnerIdx
	w.actionLog = j.ActionLog
	return nil
}
