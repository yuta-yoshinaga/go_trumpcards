//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// HasenpfefferPhase はハーゼンプフェファーのゲームフェーズ。
type HasenpfefferPhase int

// Hasenpfeffer のフェーズ定数
const (
	// HasenpfefferPhaseBid 競り
	HasenpfefferPhaseBid HasenpfefferPhase = iota
	// HasenpfefferPhaseDiscard 落札者が伏せ札を取り込んで 1 枚捨てる
	HasenpfefferPhaseDiscard
	// HasenpfefferPhasePlay プレイ中
	HasenpfefferPhasePlay
	// HasenpfefferPhaseHandEnd ハンド終了
	HasenpfefferPhaseHandEnd
	// HasenpfefferPhaseGameEnd ゲーム終了
	HasenpfefferPhaseGameEnd
)

// HasenpfefferPlayerCnt はプレイヤー数（4 人固定・2 対 2）。
const HasenpfefferPlayerCnt = 4

// HasenpfefferTeamCnt はチーム数。
const HasenpfefferTeamCnt = 2

// HasenpfefferHandSize は各プレイヤーの手札枚数。
const HasenpfefferHandSize = 6

// HasenpfefferTricksPerRound は 1 ハンドのトリック数。
//
// **ユーカーの 5 ではなく 6。** 手札が 6 枚なので 6 トリック打ちます。
const HasenpfefferTricksPerRound = HasenpfefferHandSize

// HasenpfefferDeckSize は使用するデッキの枚数。
//
// **issue の「5 枚ずつ配る」は算術が合いません。** 4 × 5 = 20 で 25 枚のうち
// 5 枚も余ってしまいます。実際のハーゼンプフェファーは **6 枚ずつ = 24 枚 +
// 伏せ札 1 枚 = 25 枚ちょうど**で、トリック数も 5 ではなく **6** です。
const HasenpfefferDeckSize = HasenpfefferPlayerCnt*HasenpfefferHandSize + 1

// HasenpfefferMinBid / HasenpfefferMaxBid は競りの範囲。
//
// **下限は 3。** 6 トリック中 1〜2 なら切り札を持つ側はまず落とさないので、
// 競りにならず宣言する意味がありません（過半の 4 ではなく 3 なのは、
// 伏せ札を取り込める落札者に踏み込む余地を残すため）。上限は全トリック。
const (
	HasenpfefferMinBid = 3
	HasenpfefferMaxBid = HasenpfefferTricksPerRound
)

// HasenpfefferDefaultTarget は既定の目標点。
const HasenpfefferDefaultTarget = 10

// hasenpfefferMaxSliceLen caps slice sizes during deserialisation.
const hasenpfefferMaxSliceLen = 1000

// HasenpfefferTeamOf は席番号からチーム番号を返す（向かい合う席が味方）。
func HasenpfefferTeamOf(playerIdx int) int { return playerIdx % HasenpfefferTeamCnt }

// Hasenpfeffer はハーゼンプフェファーのゲームクラス。
//
// アメリカのドイツ系移民に伝わるユーカー派生。**25 枚**（ユーカーの 24 枚 +
// ジョーカー 1 枚）を 4 人 2 対 2 で **6 枚ずつ**配り、余った 1 枚を伏せます。
//
// **ジョーカーが全カード中最強の切り札（Best Bower）**になるのがこのゲームの
// 顔で、Right Bower（切り札の J）と Left Bower（同色の J）の上に立ちます。
//
// **競りは全員参加が義務。** 3 人が降りたら、親は降りられず最低 3 を宣言し
// なければなりません。誰も落札しないハンドが存在しない、という形です。
type Hasenpfeffer struct {
	players     []*HasenpfefferPlayer
	config      HasenpfefferConfig
	phase       HasenpfefferPhase
	trumpCards  *TrumpCards
	trumpSuit   int
	handNumber  int
	trickNumber int
	// blind は伏せ札（落札者が取り込むまで 1 枚、以降 0 枚）。
	blind []*Card
	// declarerIdx は落札者 (-1 = 競り中)。
	declarerIdx int
	// contract は落札額。
	contract         int
	currentTrick     []*TrickCard
	currentPlayerIdx int
	leadPlayerIdx    int
	dealerIdx        int
	scores           []int
	// lastHandEuchred は直前のハンドで落札側が失敗したか。
	lastHandEuchred bool
	// lastHandTricks は直前のハンドで落札側が取ったトリック数。
	lastHandTricks int
	gameEndFlag    bool
	winnerTeam     int
	actionLogBase
}

// NewHasenpfeffer はコンストラクタ。
func NewHasenpfeffer(players []*HasenpfefferPlayer, config HasenpfefferConfig) *Hasenpfeffer {
	return &Hasenpfeffer{
		players:     players,
		config:      config,
		declarerIdx: -1,
		scores:      make([]int, HasenpfefferTeamCnt),
		winnerTeam:  -1,
	}
}

// NewDefaultHasenpfeffer は標準の 4 人セットアップを返す。
func NewDefaultHasenpfeffer() *Hasenpfeffer {
	players := make([]*HasenpfefferPlayer, 0, HasenpfefferPlayerCnt)
	players = append(players, NewHasenpfefferPlayer(true))
	for range HasenpfefferPlayerCnt - 1 {
		players = append(players, NewHasenpfefferPlayer(false))
	}
	return NewHasenpfeffer(players, DefaultHasenpfefferConfig())
}

// Reset はゲームを初期化する。
func (h *Hasenpfeffer) Reset() {
	h.handNumber = 0
	h.dealerIdx = 0
	h.scores = make([]int, HasenpfefferTeamCnt)
	h.gameEndFlag = false
	h.winnerTeam = -1
	h.lastHandEuchred = false
	h.lastHandTricks = 0
	h.actionLog = nil
	for _, p := range h.players {
		p.ResetGame()
	}
	h.startHand()
}

// startHand は 1 ハンドを配り直す。
func (h *Hasenpfeffer) startHand() {
	h.phase = HasenpfefferPhaseBid
	h.trumpSuit = 0
	h.trickNumber = 0
	h.currentTrick = nil
	h.declarerIdx = -1
	h.contract = 0
	for _, p := range h.players {
		p.ResetRound()
	}

	h.trumpCards = NewTrumpCardsHasenpfeffer()
	h.trumpCards.Shuffle()
	for range HasenpfefferHandSize {
		for i := range HasenpfefferPlayerCnt {
			idx := (h.dealerIdx + 1 + i) % HasenpfefferPlayerCnt
			if c := h.trumpCards.DrawCard(); c != nil {
				h.players[idx].AddCard(c)
			}
		}
	}
	// **余った 1 枚は伏せ札。** 落札者が取り込んで 1 枚捨てます。
	h.blind = nil
	if c := h.trumpCards.DrawCard(); c != nil {
		h.blind = []*Card{c}
	}
	h.sortAllHands()

	h.handNumber++
	h.currentPlayerIdx = (h.dealerIdx + 1) % HasenpfefferPlayerCnt
	h.leadPlayerIdx = h.currentPlayerIdx
	h.addLog(-1, "deal", fmt.Sprintf("ハンド %d を配りました（伏せ札 1 枚）", h.handNumber), nil)
}

// sortAllHands は手札をスート・ランク順に整える。
func (h *Hasenpfeffer) sortAllHands() {
	for _, p := range h.players {
		sortPlayerHand(p, func(ci, cj *Card) bool {
			if ci.GetDesign() != cj.GetDesign() {
				return ci.GetDesign() < cj.GetDesign()
			}
			return hasenpfefferBaseRank(ci) < hasenpfefferBaseRank(cj)
		})
	}
}

// hasenpfefferBaseRank は切り札を考えない素の強さ（表示順のため）。
func hasenpfefferBaseRank(c *Card) int {
	if c.GetDesign() == CardDesignJoker {
		return 99
	}
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// IsJokerCard はジョーカーかどうかを返す。
func IsJokerCard(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignJoker
}

// hasenpfefferSameColorSuit は同色スートを返す (♠↔♣, ♥↔♦)。
//
// **Euchre.go の sameColorSuit は使えない。** あちらは `solo` タグなので、
// extra3 worker のビルドから見えず `undefined: sameColorSuit` で落ちる
// （6 worker 全部をビルドして初めて分かる——ホストビルドは !js で両方
// 見えるため通ってしまう）。
func hasenpfefferSameColorSuit(suit int) int {
	switch suit {
	case CardDesignSpade:
		return CardDesignClover
	case CardDesignClover:
		return CardDesignSpade
	case CardDesignHeart:
		return CardDesignDiamond
	case CardDesignDiamond:
		return CardDesignHeart
	}
	return suit
}

// isRightBower は Right Bower（切り札スートの J）かどうか。
func (h *Hasenpfeffer) isRightBower(c *Card) bool {
	return c.GetValue() == 11 && c.GetDesign() == h.trumpSuit
}

// isLeftBower は Left Bower（同色スートの J）かどうか。
func (h *Hasenpfeffer) isLeftBower(c *Card) bool {
	return c.GetValue() == 11 && c.GetDesign() == hasenpfefferSameColorSuit(h.trumpSuit)
}

// EffectiveSuit はカードの実効スートを返す。
//
// **ジョーカーと Left Bower は切り札スートとして扱います。** ジョーカーには
// 生来のスートが無いので、フォロー判定でも切り札のリードとして働きます。
func (h *Hasenpfeffer) EffectiveSuit(c *Card) int {
	if IsJokerCard(c) || h.isLeftBower(c) {
		return h.trumpSuit
	}
	return c.GetDesign()
}

// CardRank はトリック比較用の強さを返す（高い = 強い）。
//
// **切り札: Joker(700) > Right Bower(600) > Left Bower(500) > A > K > Q > 10 > 9。**
// ジョーカーが Best Bower として全カード中最強になるのがこのゲームの顔です。
func (h *Hasenpfeffer) CardRank(c *Card) int {
	if IsJokerCard(c) {
		return 700
	}
	if h.isRightBower(c) {
		return 600
	}
	if h.isLeftBower(c) {
		return 500
	}
	base := c.GetValue()
	if base == 1 {
		base = 14
	}
	if h.EffectiveSuit(c) == h.trumpSuit {
		return 400 + base
	}
	return 100 + base
}

// --- 競り -------------------------------------------------------------------

// MustBid は playerIdx が降りられない（義務競り）かを返す。
//
// **親は 3 人が降りたら降りられません。** 誰も落札しないハンドが存在しない、
// というのがこのゲームの競りの形です。
func (h *Hasenpfeffer) MustBid(playerIdx int) bool {
	if h.phase != HasenpfefferPhaseBid || playerIdx != h.dealerIdx {
		return false
	}
	for i, p := range h.players {
		if i != h.dealerIdx && p.GetBid() > 0 {
			return false // 誰かが落札しているので親は降りられる
		}
	}
	for i, p := range h.players {
		if i != h.dealerIdx && !p.HasBid() {
			return false // まだ全員が答えていない
		}
	}
	return true
}

// NextBid は次に出せる最小の宣言額を返す（0 = もう宣言できない）。
//
// **上限に張り付いたら 0 を返す。** ここで上限そのものを返すと、あとの席が
// 同額を出して落札を横取りできてしまう——競りは上回らなければ通らない。
func (h *Hasenpfeffer) NextBid() int {
	if h.contract == 0 {
		return HasenpfefferMinBid
	}
	if h.contract >= HasenpfefferMaxBid {
		return 0
	}
	return h.contract + 1
}

// PlayerBid は人間が宣言する（0 = 降りる）。
func (h *Hasenpfeffer) PlayerBid(bid int) error {
	if !h.IsHumanBidTurn() {
		return errors.New("not your turn to bid")
	}
	return h.bidBy(0, bid)
}

// CpuBid は CPU が 1 人分宣言する。
func (h *Hasenpfeffer) CpuBid() {
	if h.gameEndFlag || h.phase != HasenpfefferPhaseBid || h.IsHumanBidTurn() {
		return
	}
	idx := h.currentPlayerIdx
	_ = h.bidBy(idx, h.chooseCpuBid(idx))
}

// bidBy は playerIdx に宣言させる。
func (h *Hasenpfeffer) bidBy(playerIdx, bid int) error {
	if h.gameEndFlag {
		return errors.New("game is over")
	}
	if h.phase != HasenpfefferPhaseBid {
		return errors.New("bidding is over")
	}
	if playerIdx != h.currentPlayerIdx {
		return fmt.Errorf("not player %d's turn", playerIdx)
	}
	if bid == 0 && h.MustBid(playerIdx) {
		// **親は 3 人が降りたら降りられない。**
		return errors.New("the dealer must bid when everyone else has passed")
	}
	if bid != 0 {
		next := h.NextBid()
		if next == 0 {
			// **上限が立っていたら誰も上回れない。** 降りるしかない。
			return fmt.Errorf("bid %d is already the maximum", HasenpfefferMaxBid)
		}
		if bid < next || bid > HasenpfefferMaxBid {
			return fmt.Errorf("bid must be %d..%d", next, HasenpfefferMaxBid)
		}
	}

	h.players[playerIdx].SetBid(bid)
	if bid > 0 {
		h.declarerIdx = playerIdx
		h.contract = bid
		h.addLog(playerIdx, "bid", fmt.Sprintf("%d トリック宣言", bid), nil)
	} else {
		h.addLog(playerIdx, "pass", "降りました", nil)
	}

	if next := h.nextBidder(playerIdx); next >= 0 {
		h.currentPlayerIdx = next
		return nil
	}
	h.closeBidding()
	return nil
}

// nextBidder はまだ答えていない次の席を返す（-1 = 全員答えた）。
func (h *Hasenpfeffer) nextBidder(from int) int {
	for i := 1; i <= HasenpfefferPlayerCnt; i++ {
		j := (from + i) % HasenpfefferPlayerCnt
		if !h.players[j].HasBid() {
			return j
		}
	}
	return -1
}

// closeBidding は競りを締めて捨て札フェーズへ進む。
func (h *Hasenpfeffer) closeBidding() {
	// **義務競りがあるので落札者は必ず決まります。** 保険として親に持たせる。
	if h.declarerIdx < 0 {
		h.declarerIdx = h.dealerIdx
		h.contract = HasenpfefferMinBid
		h.players[h.dealerIdx].SetBid(HasenpfefferMinBid)
	}
	h.phase = HasenpfefferPhaseDiscard
	h.currentPlayerIdx = h.declarerIdx
	// **伏せ札は落札者の手に入る。** 25 枚目が死に札にならない形。
	for _, c := range h.blind {
		h.players[h.declarerIdx].AddCard(c)
	}
	h.blind = nil
	h.sortAllHands()
	h.addLog(h.declarerIdx, "win", fmt.Sprintf("%d トリックで落札、伏せ札を取り込みました", h.contract), nil)
}

// IsHumanBidTurn は人間が宣言する番かを返す。
func (h *Hasenpfeffer) IsHumanBidTurn() bool {
	return !h.gameEndFlag && h.phase == HasenpfefferPhaseBid && h.currentPlayerIdx == 0
}

// IsHumanDiscardTurn は人間が捨て札とスート宣言をする番かを返す。
func (h *Hasenpfeffer) IsHumanDiscardTurn() bool {
	return !h.gameEndFlag && h.phase == HasenpfefferPhaseDiscard && h.declarerIdx == 0
}

// chooseCpuBid は CPU の宣言。**強い札の枚数で決めます。**
func (h *Hasenpfeffer) chooseCpuBid(playerIdx int) int {
	p := h.players[playerIdx]
	// **どのスートを切り札にするかはここでは決めない。** 決めるのは落札後の
	// chooseCpuTrump で、競りに要るのは「いちばん強い組み合わせの強さ」だけ。
	bestScore := -1
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		bestScore = max(bestScore, h.handStrength(p, suit))
	}
	// 強さをトリック数の見積もりに落とす。届かなければ降りる。
	estimate := HasenpfefferMinBid + (bestScore-6)/3
	estimate = min(max(estimate, 0), HasenpfefferMaxBid)
	next := h.NextBid()
	// **上限が立っていたら降りるしかない。**
	if next == 0 {
		return 0
	}
	if estimate < next {
		if h.MustBid(playerIdx) {
			return next
		}
		return 0
	}
	return estimate
}

// handStrength は suit を切り札にしたときの手札の強さを見積もる。
func (h *Hasenpfeffer) handStrength(p *HasenpfefferPlayer, suit int) int {
	score := 0
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		switch {
		case IsJokerCard(c):
			score += 5 // Best Bower
		case c.GetValue() == 11 && c.GetDesign() == suit:
			score += 4 // Right Bower
		case c.GetValue() == 11 && c.GetDesign() == hasenpfefferSameColorSuit(suit):
			score += 3 // Left Bower
		case c.GetDesign() == suit:
			score += 2
		case c.GetValue() == 1:
			score++ // 非切り札の A
		}
	}
	return score
}

// PlayerDiscard は人間が切り札を宣言し、1 枚捨てる。
func (h *Hasenpfeffer) PlayerDiscard(cardIndex, suit int) error {
	if !h.IsHumanDiscardTurn() {
		return errors.New("not your discard")
	}
	return h.discardBy(0, cardIndex, suit)
}

// CpuDiscard は CPU の落札者が切り札を宣言して 1 枚捨てる。
func (h *Hasenpfeffer) CpuDiscard() {
	if h.gameEndFlag || h.phase != HasenpfefferPhaseDiscard || h.IsHumanDiscardTurn() {
		return
	}
	idx := h.declarerIdx
	suit := h.chooseCpuTrump(idx)
	_ = h.discardBy(idx, h.chooseCpuDiscard(idx, suit), suit)
}

// discardBy は落札者に切り札を宣言させ、1 枚捨てさせる。
func (h *Hasenpfeffer) discardBy(playerIdx, cardIndex, suit int) error {
	if h.phase != HasenpfefferPhaseDiscard {
		return errors.New("not the discard phase")
	}
	if playerIdx != h.declarerIdx {
		return errors.New("only the declarer discards")
	}
	if suit < CardDesignSpade || suit > CardDesignDiamond {
		return fmt.Errorf("invalid trump suit: %d", suit)
	}
	p := h.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}

	h.trumpSuit = suit
	discarded := p.RemoveCard(cardIndex)
	h.phase = HasenpfefferPhasePlay
	h.leadPlayerIdx = (h.dealerIdx + 1) % HasenpfefferPlayerCnt
	h.currentPlayerIdx = h.leadPlayerIdx
	h.sortAllHands()
	h.addLog(playerIdx, "trump", fmt.Sprintf("切り札は %s、%s を捨てました",
		suitStr(suit), cardStr(discarded)), nil)
	return nil
}

// chooseCpuTrump は CPU の切り札選び。
func (h *Hasenpfeffer) chooseCpuTrump(playerIdx int) int {
	p := h.players[playerIdx]
	best, bestScore := CardDesignSpade, -1
	for _, suit := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if s := h.handStrength(p, suit); s > bestScore {
			best, bestScore = suit, s
		}
	}
	return best
}

// chooseCpuDiscard は CPU が捨てる札。**いちばん弱い非切り札。**
func (h *Hasenpfeffer) chooseCpuDiscard(playerIdx, suit int) int {
	p := h.players[playerIdx]
	pick, pickRank := 0, 1<<30
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if IsJokerCard(c) || c.GetDesign() == suit ||
			(c.GetValue() == 11 && c.GetDesign() == hasenpfefferSameColorSuit(suit)) {
			continue // 切り札は捨てない
		}
		if r := hasenpfefferBaseRank(c); r < pickRank {
			pick, pickRank = i, r
		}
	}
	return pick
}

// --- プレイ -----------------------------------------------------------------

// GetValidPlayIndices は playerIdx が出せる手札の添字を返す。
//
// **フォロー義務あり。** ジョーカーと Left Bower は切り札スートとして扱うので、
// 切り札がリードされたらそれらも「フォローできる札」になります。
func (h *Hasenpfeffer) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= HasenpfefferPlayerCnt || h.gameEndFlag {
		return []int{}
	}
	if h.phase != HasenpfefferPhasePlay {
		return []int{}
	}
	p := h.players[playerIdx]
	leadSuit := 0
	if len(h.currentTrick) > 0 {
		leadSuit = h.EffectiveSuit(h.currentTrick[0].Card)
	}
	all := make([]int, 0, p.GetCardsSize())
	follow := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		all = append(all, i)
		if leadSuit != 0 && h.EffectiveSuit(p.GetCard(i)) == leadSuit {
			follow = append(follow, i)
		}
	}
	if len(follow) > 0 {
		return follow
	}
	return all
}

// IsHumanTurn は現在の手番が人間かを返す。
func (h *Hasenpfeffer) IsHumanTurn() bool {
	if h.gameEndFlag || h.phase != HasenpfefferPhasePlay {
		return false
	}
	return h.players[h.currentPlayerIdx].GetIsHuman()
}

// PlayerPlay は人間が 1 枚出す。
func (h *Hasenpfeffer) PlayerPlay(cardIndex int) error {
	if !h.IsHumanTurn() {
		return errors.New("not your turn")
	}
	return h.play(0, cardIndex)
}

// CpuPlay は CPU が 1 枚出す。
func (h *Hasenpfeffer) CpuPlay() {
	if h.gameEndFlag || h.phase != HasenpfefferPhasePlay || h.IsHumanTurn() {
		return
	}
	_ = h.play(h.currentPlayerIdx, h.chooseCpuCard(h.currentPlayerIdx))
}

// play は playerIdx に手札の cardIndex を出させる。
func (h *Hasenpfeffer) play(playerIdx, cardIndex int) error {
	if h.gameEndFlag {
		return errors.New("game is over")
	}
	if h.phase != HasenpfefferPhasePlay {
		return errors.New("not the play phase")
	}
	if playerIdx != h.currentPlayerIdx {
		return fmt.Errorf("not player %d's turn", playerIdx)
	}
	p := h.players[playerIdx]
	if cardIndex < 0 || cardIndex >= p.GetCardsSize() {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	if !hasenpfefferContains(h.GetValidPlayIndices(playerIdx), cardIndex) {
		return errors.New("must follow the led suit")
	}

	card := p.RemoveCard(cardIndex)
	h.currentTrick = append(h.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	h.addLog(playerIdx, "play", cardStr(card), []*Card{card})

	if len(h.currentTrick) < HasenpfefferPlayerCnt {
		h.currentPlayerIdx = (h.currentPlayerIdx + 1) % HasenpfefferPlayerCnt
		return nil
	}
	h.resolveTrick()
	return nil
}

// resolveTrick はトリックを解決する。
func (h *Hasenpfeffer) resolveTrick() {
	winner := h.trickWinner()
	cards := make([]*Card, 0, HasenpfefferPlayerCnt)
	for _, tc := range h.currentTrick {
		cards = append(cards, tc.Card)
	}
	h.players[winner].AddTrick(cards)
	h.currentTrick = nil
	h.trickNumber++
	h.leadPlayerIdx = winner
	h.currentPlayerIdx = winner
	h.addLog(winner, "trick", fmt.Sprintf("トリック %d を取りました", h.trickNumber), nil)

	if h.trickNumber >= HasenpfefferTricksPerRound {
		h.finishHand()
	}
}

// trickWinner は最強札を出した人を返す。
func (h *Hasenpfeffer) trickWinner() int {
	if len(h.currentTrick) == 0 {
		return h.leadPlayerIdx
	}
	leadSuit := h.EffectiveSuit(h.currentTrick[0].Card)
	best, bestRank := h.currentTrick[0].PlayerIdx, -1
	for _, tc := range h.currentTrick {
		suit := h.EffectiveSuit(tc.Card)
		// 切り札でもリードスートでもない札は取れない。
		if suit != h.trumpSuit && suit != leadSuit {
			continue
		}
		if r := h.CardRank(tc.Card); r > bestRank {
			best, bestRank = tc.PlayerIdx, r
		}
	}
	return best
}

// TeamTricks はチームが取ったトリック数を返す。
func (h *Hasenpfeffer) TeamTricks(team int) int {
	n := 0
	for i, p := range h.players {
		if HasenpfefferTeamOf(i) == team {
			n += p.GetTrickCount()
		}
	}
	return n
}

// finishHand はハンドを精算する。
//
// **達成したら取ったトリック数がそのまま点。** 落とした（euchred）ら相手に
// 落札額が入ります。多く取っても宣言額を超えたぶんが無駄にならない形です。
func (h *Hasenpfeffer) finishHand() {
	h.phase = HasenpfefferPhaseHandEnd
	declTeam := HasenpfefferTeamOf(h.declarerIdx)
	took := h.TeamTricks(declTeam)
	h.lastHandTricks = took
	if took >= h.contract {
		h.scores[declTeam] += took
		h.lastHandEuchred = false
		h.addLog(h.declarerIdx, "score",
			fmt.Sprintf("契約 %d に対し %d トリック：達成 (+%d)", h.contract, took, took), nil)
	} else {
		other := 1 - declTeam
		h.scores[other] += h.contract
		h.lastHandEuchred = true
		h.addLog(h.declarerIdx, "score",
			fmt.Sprintf("契約 %d に対し %d トリック：失敗、相手に +%d", h.contract, took, h.contract), nil)
	}
	for team := range HasenpfefferTeamCnt {
		if h.scores[team] >= h.config.Target {
			h.finishGame()
			return
		}
	}
}

// NextHand は次のハンドを開始する。
func (h *Hasenpfeffer) NextHand() {
	if h.gameEndFlag || h.phase != HasenpfefferPhaseHandEnd {
		return
	}
	h.dealerIdx = (h.dealerIdx + 1) % HasenpfefferPlayerCnt
	h.startHand()
}

// finishGame は終局処理。
func (h *Hasenpfeffer) finishGame() {
	h.phase = HasenpfefferPhaseGameEnd
	h.gameEndFlag = true
	switch {
	case h.scores[0] > h.scores[1]:
		h.winnerTeam = 0
	case h.scores[1] > h.scores[0]:
		h.winnerTeam = 1
	default:
		h.winnerTeam = -1
	}
	h.addLog(-1, "result", fmt.Sprintf("最終得点 %d - %d", h.scores[0], h.scores[1]), nil)
}

// GiveUp は投了する。
func (h *Hasenpfeffer) GiveUp() {
	if h.gameEndFlag {
		return
	}
	h.phase = HasenpfefferPhaseGameEnd
	h.gameEndFlag = true
	h.winnerTeam = 1
	h.addLog(0, "giveup", "投了しました", nil)
}

// chooseCpuCard は CPU の手。
func (h *Hasenpfeffer) chooseCpuCard(playerIdx int) int {
	valid := h.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := h.players[playerIdx]
	// **味方が勝っているトリックには強い札を出さない。**
	partnerWinning := false
	if len(h.currentTrick) > 0 {
		w := h.trickWinner()
		partnerWinning = w != playerIdx && HasenpfefferTeamOf(w) == HasenpfefferTeamOf(playerIdx)
	}

	pick, pickRank := valid[0], h.CardRank(p.GetCard(valid[0]))
	for _, i := range valid[1:] {
		r := h.CardRank(p.GetCard(i))
		if partnerWinning && r < pickRank {
			pick, pickRank = i, r
		} else if !partnerWinning && r > pickRank {
			pick, pickRank = i, r
		}
	}
	return pick
}

// HasenpfefferHint はハーゼンプフェファーの助言。
type HasenpfefferHint struct {
	CardIndex *int
	Reason    string
	// Value は宣言すべきトリック数（プレイ中は 0）。
	Value int
	// Suit は切り札に勧めるスート（プレイ中は 0）。
	Suit int
}

// GetHint は人間への助言を返す。
func (h *Hasenpfeffer) GetHint() *HasenpfefferHint {
	if h.gameEndFlag {
		return nil
	}
	switch {
	case h.IsHumanBidTurn():
		bid := h.chooseCpuBid(0)
		reason := "hasenpfefferBid"
		if bid == 0 {
			reason = "hasenpfefferPass"
		} else if h.MustBid(0) {
			// **親は降りられない。** 降りる選択肢が無いことを言う。
			reason = "hasenpfefferMustBid"
		}
		return &HasenpfefferHint{Reason: reason, Value: bid}
	case h.IsHumanDiscardTurn():
		suit := h.chooseCpuTrump(0)
		idx := h.chooseCpuDiscard(0, suit)
		return &HasenpfefferHint{CardIndex: &idx, Reason: "hasenpfefferDiscard", Suit: suit}
	case h.IsHumanTurn():
		valid := h.GetValidPlayIndices(0)
		if len(valid) == 0 {
			return nil
		}
		idx := h.chooseCpuCard(0)
		reason := "hasenpfefferWinTrick"
		if len(h.currentTrick) > 0 {
			w := h.trickWinner()
			if w != 0 && HasenpfefferTeamOf(w) == HasenpfefferTeamOf(0) {
				reason = "hasenpfefferFeedPartner"
			}
		}
		return &HasenpfefferHint{CardIndex: &idx, Reason: reason}
	default:
		return nil
	}
}

// hasenpfefferContains は xs が v を含むかを返す。
func hasenpfefferContains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// addLog は棋譜に 1 行足す。
func (h *Hasenpfeffer) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	h.appendLog(playerIdx, actionType, detail, cards)
}

// --- アクセサ ---------------------------------------------------------------

// GetConfig はゲーム設定を返す。
func (h *Hasenpfeffer) GetConfig() HasenpfefferConfig { return h.config }

// SetConfig はゲーム設定を設定する。
func (h *Hasenpfeffer) SetConfig(cfg HasenpfefferConfig) { h.config = cfg }

// GetPhase は現在のフェーズを返す。
func (h *Hasenpfeffer) GetPhase() HasenpfefferPhase { return h.phase }

// GetGameEndFlag はゲーム終了フラグを返す。
func (h *Hasenpfeffer) GetGameEndFlag() bool { return h.gameEndFlag }

// GetHandNumber は現在のハンド番号を返す。
func (h *Hasenpfeffer) GetHandNumber() int { return h.handNumber }

// GetTrickNumber は現在のトリック番号を返す。
func (h *Hasenpfeffer) GetTrickNumber() int { return h.trickNumber }

// GetTrumpSuit は切り札のスートを返す（0 = 未宣言）。
func (h *Hasenpfeffer) GetTrumpSuit() int { return h.trumpSuit }

// GetDeclarerIdx は落札者を返す（-1 = 競り中）。
func (h *Hasenpfeffer) GetDeclarerIdx() int { return h.declarerIdx }

// GetContract は落札額を返す。
func (h *Hasenpfeffer) GetContract() int { return h.contract }

// GetBlindSize は伏せ札の枚数を返す（落札後は 0）。
func (h *Hasenpfeffer) GetBlindSize() int { return len(h.blind) }

// GetCurrentPlayerIdx は現在の手番を返す。
func (h *Hasenpfeffer) GetCurrentPlayerIdx() int { return h.currentPlayerIdx }

// GetLeadPlayerIdx はリードプレイヤーを返す。
func (h *Hasenpfeffer) GetLeadPlayerIdx() int { return h.leadPlayerIdx }

// GetDealerIdx はディーラーを返す。
func (h *Hasenpfeffer) GetDealerIdx() int { return h.dealerIdx }

// GetCurrentTrick は現在のトリックを返す。
func (h *Hasenpfeffer) GetCurrentTrick() []*TrickCard { return h.currentTrick }

// GetScore はチームの得点を返す。
func (h *Hasenpfeffer) GetScore(team int) int {
	if team < 0 || team >= HasenpfefferTeamCnt {
		return 0
	}
	return h.scores[team]
}

// SetScoreForTestUse はチームの得点を設定する（復元・テスト用）。
func (h *Hasenpfeffer) SetScoreForTestUse(team, n int) {
	if team >= 0 && team < HasenpfefferTeamCnt {
		h.scores[team] = n
	}
}

// GetLastHandEuchred は直前のハンドで落札側が失敗したかを返す。
func (h *Hasenpfeffer) GetLastHandEuchred() bool { return h.lastHandEuchred }

// GetLastHandTricks は直前のハンドで落札側が取ったトリック数を返す。
func (h *Hasenpfeffer) GetLastHandTricks() int { return h.lastHandTricks }

// GetPlayerCnt はプレイヤー数を返す。
func (h *Hasenpfeffer) GetPlayerCnt() int { return HasenpfefferPlayerCnt }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (h *Hasenpfeffer) GetPlayer(i int) *HasenpfefferPlayer {
	if i < 0 || i >= len(h.players) {
		return nil
	}
	return h.players[i]
}

// GetWinnerTeam は勝利チームを返す（-1 = 未確定/同点）。
func (h *Hasenpfeffer) GetWinnerTeam() int { return h.winnerTeam }

// GetActionLog は棋譜を返す。
func (h *Hasenpfeffer) GetActionLog() []*ActionLogEntry { return h.actionLog }

// hasenpfefferJSON は KV スナップショットの表現。
type hasenpfefferJSON struct {
	TrumpCards       *TrumpCards           `json:"tc"`
	Players          []*HasenpfefferPlayer `json:"pl"`
	Config           HasenpfefferConfig    `json:"cf"`
	Phase            HasenpfefferPhase     `json:"ph"`
	TrumpSuit        int                   `json:"ts"`
	HandNumber       int                   `json:"hn"`
	TrickNumber      int                   `json:"tn"`
	Blind            []*Card               `json:"bl"`
	DeclarerIdx      int                   `json:"di"`
	Contract         int                   `json:"co"`
	CurrentTrick     []*TrickCard          `json:"ct"`
	CurrentPlayerIdx int                   `json:"ci"`
	LeadPlayerIdx    int                   `json:"li"`
	DealerIdx        int                   `json:"dl"`
	Scores           []int                 `json:"sc"`
	LastHandEuchred  bool                  `json:"le"`
	LastHandTricks   int                   `json:"lt"`
	GameEndFlag      bool                  `json:"ge"`
	WinnerTeam       int                   `json:"wt"`
	ActionLog        []*ActionLogEntry     `json:"al"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (h *Hasenpfeffer) MarshalJSON() ([]byte, error) {
	return json.Marshal(&hasenpfefferJSON{
		TrumpCards: h.trumpCards, Players: h.players, Config: h.config, Phase: h.phase,
		TrumpSuit: h.trumpSuit, HandNumber: h.handNumber, TrickNumber: h.trickNumber,
		Blind: h.blind, DeclarerIdx: h.declarerIdx, Contract: h.contract,
		CurrentTrick: h.currentTrick, CurrentPlayerIdx: h.currentPlayerIdx,
		LeadPlayerIdx: h.leadPlayerIdx, DealerIdx: h.dealerIdx, Scores: h.scores,
		LastHandEuchred: h.lastHandEuchred, LastHandTricks: h.lastHandTricks,
		GameEndFlag: h.gameEndFlag, WinnerTeam: h.winnerTeam, ActionLog: h.actionLog,
	})
}

// UnmarshalJSON KV スナップショットからの復元
func (h *Hasenpfeffer) UnmarshalJSON(data []byte) error {
	var j hasenpfefferJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if j.Phase < HasenpfefferPhaseBid || j.Phase > HasenpfefferPhaseGameEnd {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	// **切り札は捨て札を終えてから立つ。** 素通しすると CardRank が
	// どの札も切り札と見なさなくなり、勝敗が黙って変わる (#5302〜#5305)。
	if j.Phase == HasenpfefferPhaseBid || j.Phase == HasenpfefferPhaseDiscard {
		if j.TrumpSuit != 0 {
			return fmt.Errorf("trump suit %d before it was declared", j.TrumpSuit)
		}
	} else if j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond {
		return fmt.Errorf("invalid trump suit: %d", j.TrumpSuit)
	}
	// **落札者と落札額は対。** 競り中は両方空、締めたら両方埋まっている。
	if j.Phase == HasenpfefferPhaseBid {
		if j.DeclarerIdx != -1 || j.Contract != 0 {
			return fmt.Errorf("declarer %d / contract %d during bidding", j.DeclarerIdx, j.Contract)
		}
	} else {
		if j.DeclarerIdx < 0 || j.DeclarerIdx >= HasenpfefferPlayerCnt {
			return fmt.Errorf("invalid declarer: %d", j.DeclarerIdx)
		}
		if j.Contract < HasenpfefferMinBid || j.Contract > HasenpfefferMaxBid {
			return fmt.Errorf("invalid contract: %d", j.Contract)
		}
	}
	// **伏せ札は競り中だけ 1 枚。** 落札者が取り込んだら消える。
	wantBlind := 0
	if j.Phase == HasenpfefferPhaseBid {
		wantBlind = 1
	}
	if len(j.Blind) != wantBlind {
		return fmt.Errorf("blind holds %d cards in phase %d", len(j.Blind), j.Phase)
	}
	if j.HandNumber < 1 {
		return fmt.Errorf("invalid hand number: %d", j.HandNumber)
	}
	if j.TrickNumber < 0 || j.TrickNumber > HasenpfefferTricksPerRound {
		return fmt.Errorf("invalid trick number: %d", j.TrickNumber)
	}
	if len(j.CurrentTrick) > HasenpfefferPlayerCnt {
		return fmt.Errorf("current trick holds %d cards", len(j.CurrentTrick))
	}
	// **枚数だけでなく中身も見る（レビュー指摘 PR #5310）。** 壊れた KV から
	// nil の Card が入ると EffectiveSuit → isLeftBower が nil を参照して panic し、
	// 範囲外の PlayerIdx は resolveTrick の h.players[winner] で panic する。
	// 悪意ある入力でなくても古い/壊れたエントリで到達する。Nap.go:872 と同じ形。
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil || tc.PlayerIdx < 0 || tc.PlayerIdx >= HasenpfefferPlayerCnt {
			return errors.New("invalid current trick entry")
		}
	}
	if len(j.ActionLog) > hasenpfefferMaxSliceLen {
		return errors.New("hasenpfeffer: input array exceeds maximum allowed size")
	}
	for name, idx := range map[string]int{
		"current player": j.CurrentPlayerIdx,
		"lead player":    j.LeadPlayerIdx,
		"dealer":         j.DealerIdx,
	} {
		if idx < 0 || idx >= HasenpfefferPlayerCnt {
			return fmt.Errorf("invalid %s: %d", name, idx)
		}
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= HasenpfefferTeamCnt {
		return fmt.Errorf("invalid winner team: %d", j.WinnerTeam)
	}
	if !j.GameEndFlag && j.WinnerTeam != -1 {
		return fmt.Errorf("winner %d before the game ended", j.WinnerTeam)
	}
	if len(j.Scores) != HasenpfefferTeamCnt {
		return fmt.Errorf("scores holds %d entries", len(j.Scores))
	}
	if j.LastHandTricks < 0 || j.LastHandTricks > HasenpfefferTricksPerRound {
		return fmt.Errorf("invalid last hand tricks: %d", j.LastHandTricks)
	}

	if j.TrumpCards != nil {
		h.trumpCards = j.TrumpCards
	}
	if len(j.Players) == HasenpfefferPlayerCnt {
		h.players = j.Players
	}
	h.config, h.phase, h.trumpSuit = j.Config, j.Phase, j.TrumpSuit
	h.handNumber, h.trickNumber = j.HandNumber, j.TrickNumber
	h.blind, h.declarerIdx, h.contract = j.Blind, j.DeclarerIdx, j.Contract
	h.currentTrick, h.currentPlayerIdx = j.CurrentTrick, j.CurrentPlayerIdx
	h.leadPlayerIdx, h.dealerIdx, h.scores = j.LeadPlayerIdx, j.DealerIdx, j.Scores
	h.lastHandEuchred, h.lastHandTricks = j.LastHandEuchred, j.LastHandTricks
	h.gameEndFlag, h.winnerTeam, h.actionLog = j.GameEndFlag, j.WinnerTeam, j.ActionLog
	return nil
}
