//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"slices"
)

// リッケンのフェーズ定数
const (
	// RikkenPhaseBid 契約を競っている場面
	RikkenPhaseBid = 0
	// RikkenPhaseCall 落札者が切り札と相方の札を決める場面
	RikkenPhaseCall = 1
	// RikkenPhasePlay トリックを進めている場面
	RikkenPhasePlay = 2
	// RikkenPhaseRoundEnd ラウンドの精算が済んだ場面
	RikkenPhaseRoundEnd = 3
	// RikkenPhaseGameEnd ゲーム終了
	RikkenPhaseGameEnd = 4
)

// RikkenNoTrump は切り札なしを表すスート値 (どのスートとも一致しません)。
const RikkenNoTrump = -1

// rikkenMaxSliceLen は復元時に受け付けるスライス長の上限。
const rikkenMaxSliceLen = 2000

// RikkenHint は人間への助言。
type RikkenHint struct {
	// Contract は宣言すべき契約 (プレイ中は nil)。
	Contract *int
	// CardIndex は出すべき手札の位置 (競り中は nil)。
	CardIndex *int
	// Reason は助言の理由。
	Reason string
}

// Rikken はリッケン (Rikken) のゲームクラス。
//
// オランダの**競り + 契約トリックテイキング**。標準 52 枚を 13 枚ずつ配ります。
//
// # 種類の違う契約が同じ梯子に並ぶ
//
// Rik (相方を呼んで 8 トリック)、Misere (1 トリックも取らない)、Solo (単独で
// 6 トリック)、Open Misere (手札を公開して 0 トリック) が**ひとつの競りに並びます**。
// 「多く取る」契約と「1 枚も取らない」契約が強さ順に混ざるのがこのゲームの形です。
//
// # Rik だけが 2 対 2
//
// Rik を落とした人は切り札を決め、**自分が持っていないエースを 1 枚指名**します。
// その札を持っている人が相方です。ほかの 3 つの契約は単独で、残り 3 人が守ります。
// **組は席では決まりません**——契約ごとに変わります。
//
// # 得点はゼロサム
//
// 卓の得点の合計は常に 0 です。誰かが得た点は必ず誰かが失っています。
type Rikken struct {
	trumpCards *TrumpCards
	players    []*RikkenPlayer
	config     RikkenConfig

	phase int
	// dealerIdx は親。競りは親の左隣から始まります。
	dealerIdx int
	// contract は落札された契約。
	contract int
	// declarerIdx は落札者 (-1 = 未定)。
	declarerIdx int
	// partnerIdx は Rik の相方 (-1 = 相方なし、または未確定)。
	partnerIdx int
	// calledCard は Rik で指名した札。
	calledCard *Card
	// trumpSuit は切り札 (RikkenNoTrump = 切り札なし)。
	trumpSuit int
	// passed[i] は席 i が降りたか。
	passed []bool

	currentTurn int
	trick       []*TrickCard
	lastTrick   []*TrickCard
	// lastTrickWinner は直前のトリックを取った席 (-1 = まだ無い)。
	lastTrickWinner int
	trickCount      int
	// declarerTricks は宣言側が取ったトリック数。
	declarerTricks int
	// partnerRevealed は相方が判明したか。**指名した札が出るまで伏せます。**
	partnerRevealed bool

	roundNumber int
	gameEndFlag bool
	// winnerIdx は最終的にいちばん点の高い席 (-1 = 未確定)。
	winnerIdx int
	actionLogBase

	rng *rand.Rand
}

// NewRikken はコンストラクタ。
func NewRikken(trumpCards *TrumpCards, players []*RikkenPlayer, config RikkenConfig) *Rikken {
	if config.Validate() != nil {
		config = DefaultRikkenConfig()
	}
	if players == nil {
		players = newRikkenSeats()
	}
	if trumpCards == nil {
		trumpCards = NewTrumpCards(0)
	}
	return &Rikken{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		contract:        RikkenContractNone,
		declarerIdx:     -1,
		partnerIdx:      -1,
		trumpSuit:       RikkenNoTrump,
		lastTrickWinner: -1,
		winnerIdx:       -1,
		passed:          make([]bool, RikkenPlayerCnt),
		rng:             rand.New(rand.NewSource(rand.Int63())),
	}
}

// newRikkenSeats は席 0 を人間、以降を CPU とした座席を作る。
func newRikkenSeats() []*RikkenPlayer {
	seats := make([]*RikkenPlayer, RikkenPlayerCnt)
	for i := range seats {
		seats[i] = NewRikkenPlayer(i == 0)
	}
	return seats
}

// NewDefaultRikken は既定設定の Rikken を返す。
func NewDefaultRikken() *Rikken {
	return NewRikken(NewTrumpCards(0), newRikkenSeats(), DefaultRikkenConfig())
}

// SetRand はテスト用に乱数源を差し替える。
func (g *Rikken) SetRand(r *rand.Rand) {
	if r != nil {
		g.rng = r
	}
}

// Reset はゲームを初期化する。
func (g *Rikken) Reset() {
	for _, p := range g.players {
		p.ResetGame()
	}
	g.roundNumber = 0
	g.dealerIdx = 0
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.actionLog = nil
	g.addLog(-1, "start", "リッケンを開始しました", nil)
	g.startRound()
}

// startRound は 1 ラウンドを配り直す。
func (g *Rikken) startRound() {
	g.trumpCards = NewTrumpCards(0)
	g.trumpCards.Shuffle()
	for _, p := range g.players {
		p.ResetRound()
	}
	for range RikkenHandSize {
		for i := range g.players {
			g.players[i].AddCard(g.trumpCards.DrawCard())
		}
	}

	g.phase = RikkenPhaseBid
	g.contract = RikkenContractNone
	g.declarerIdx = -1
	g.partnerIdx = -1
	g.calledCard = nil
	g.trumpSuit = RikkenNoTrump
	g.passed = make([]bool, RikkenPlayerCnt)
	g.trick = nil
	g.lastTrick = nil
	g.lastTrickWinner = -1
	g.trickCount = 0
	g.declarerTricks = 0
	g.partnerRevealed = false
	g.roundNumber++
	// 競りは親の左隣から。
	g.currentTurn = (g.dealerIdx + 1) % RikkenPlayerCnt

	g.addLog(g.dealerIdx, "deal",
		fmt.Sprintf("ラウンド %d を配りました", g.roundNumber), nil)
	g.advanceCpu()
}

// Bid は契約を宣言する。RikkenContractNone を渡すとパス。
func (g *Rikken) Bid(contract int) error {
	if g.phase != RikkenPhaseBid {
		return errors.New("いまは競りの場面ではありません")
	}
	if !g.players[g.currentTurn].GetIsHuman() {
		return errors.New("あなたの番ではありません")
	}
	return g.bid(g.currentTurn, contract)
}

// bid は席 idx に競らせる。
func (g *Rikken) bid(idx, contract int) error {
	if err := rikkenValidateContract(contract); err != nil {
		return err
	}
	if contract == RikkenContractNone {
		g.passed[idx] = true
		g.addLog(idx, "pass", "パスしました", nil)
		g.advanceBid()
		return nil
	}
	// **上へしか積めません。**
	if contract <= g.contract {
		return fmt.Errorf("いまの契約より強い宣言が要ります: %s", RikkenContractName(g.contract))
	}
	g.contract = contract
	g.declarerIdx = idx
	// **競り上げた本人は降りていない扱いに戻ります。**
	g.passed[idx] = false
	g.addLog(idx, "bid", RikkenContractName(contract)+" を宣言しました", nil)
	g.advanceBid()
	return nil
}

// advanceBid は競りを次の席へ回し、決着していれば次のフェーズへ進む。
func (g *Rikken) advanceBid() {
	if g.biddingDone() {
		g.finishBidding()
		return
	}
	for step := 1; step <= RikkenPlayerCnt; step++ {
		next := (g.currentTurn + step) % RikkenPlayerCnt
		if !g.passed[next] {
			g.currentTurn = next
			return
		}
	}
	g.finishBidding()
}

// biddingDone は競りが終わったかを返す。
//
// **降りていない席が 1 つだけになったら終わり**です。誰も宣言していないうちは
// 全員が降りるまで続きます。
func (g *Rikken) biddingDone() bool {
	alive := 0
	for _, p := range g.passed {
		if !p {
			alive++
		}
	}
	if g.contract == RikkenContractNone {
		return alive == 0
	}
	return alive <= 1
}

// finishBidding は競りを締めて次のフェーズへ進む。
func (g *Rikken) finishBidding() {
	if g.contract == RikkenContractNone {
		// **全員が降りたら親が Rik を引き受けます。** 流局にすると終わらないためです。
		g.declarerIdx = g.dealerIdx
		g.contract = RikkenContractRik
		g.addLog(g.dealerIdx, "forced", "全員パスのため親が rik を引き受けます", nil)
	}
	g.currentTurn = g.declarerIdx
	if RikkenNeedsTrump(g.contract) {
		g.phase = RikkenPhaseCall
		g.advanceCpu()
		return
	}
	g.startPlay()
}

// Call は切り札を決め、Rik なら相方の札を指名する。
func (g *Rikken) Call(trumpSuit int) error {
	if g.phase != RikkenPhaseCall {
		return errors.New("いまは指名の場面ではありません")
	}
	if !g.players[g.currentTurn].GetIsHuman() {
		return errors.New("あなたの番ではありません")
	}
	return g.call(g.currentTurn, trumpSuit)
}

// call は席 idx に切り札と相方を決めさせる。
func (g *Rikken) call(idx, trumpSuit int) error {
	if trumpSuit < CardDesignSpade || trumpSuit > CardDesignDiamond {
		return fmt.Errorf("切り札のスートが範囲外です: %d", trumpSuit)
	}
	g.trumpSuit = trumpSuit
	g.addLog(idx, "trump", rikkenSuitName(trumpSuit)+" を切り札にしました", nil)

	if RikkenHasPartner(g.contract) {
		card := g.chooseCalledCard(idx)
		g.calledCard = card
		g.partnerIdx = g.holderOf(card)
		g.addLog(idx, "call", "相方の札を指名しました", []*Card{card})
	}
	g.startPlay()
	return nil
}

// chooseCalledCard は指名する札を選ぶ。
//
// **自分が持っていないエースを呼びます。** 4 枚とも持っているならキング、
// それも無ければ持っていない札のどれか——13 枚しか持たないので必ず見つかります。
func (g *Rikken) chooseCalledCard(idx int) *Card {
	var trumpAce *Card
	for _, value := range []int{1, 13, 12, 11} {
		for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
			c := NewCard(suit, value, false)
			if h := g.holderOf(c); h == idx || h < 0 {
				continue
			}
			if value == 1 && suit == g.trumpSuit {
				// **切り札のエースは後回しにするだけで、捨てません。** ここで
				// `continue` したまま二度と見ないと、非切り札のエース 3 枚を自分で
				// 持っている手（よくある形）で、他家が持つ切り札のエースを飛ばして
				// キングを呼んでしまいます。
				trumpAce = c
				continue
			}
			return c
		}
		if value == 1 && trumpAce != nil {
			return trumpAce
		}
	}
	// ここには来ませんが、来たら自分以外の誰かの 1 枚目を指名します。
	for i := range g.players {
		if i != idx && g.players[i].GetCardsSize() > 0 {
			return g.players[i].GetCard(0)
		}
	}
	return nil
}

// holderOf は札を持っている席を返す (-1 = 誰も持っていない)。
func (g *Rikken) holderOf(c *Card) int {
	if c == nil {
		return -1
	}
	for i, p := range g.players {
		for k := range p.GetCardsSize() {
			h := p.GetCard(k)
			if h.GetDesign() == c.GetDesign() && h.GetValue() == c.GetValue() {
				return i
			}
		}
	}
	return -1
}

// startPlay はトリックの場面へ進む。
func (g *Rikken) startPlay() {
	g.phase = RikkenPhasePlay
	// 落札者の左隣からリード。
	g.currentTurn = (g.declarerIdx + 1) % RikkenPlayerCnt
	g.addLog(-1, "play",
		fmt.Sprintf("%s で開始します", RikkenContractName(g.contract)), nil)
	g.advanceCpu()
}

// IsDeclarerSide は席 idx が宣言側かを返す。
//
// **組は席では決まりません。** Rik のときだけ相方が加わります。
func (g *Rikken) IsDeclarerSide(idx int) bool {
	if idx == g.declarerIdx {
		return true
	}
	return RikkenHasPartner(g.contract) && idx == g.partnerIdx
}

// GetValidPlayIndices は席 idx が出せる手札の位置を返す (フォロー義務のみ)。
func (g *Rikken) GetValidPlayIndices(idx int) []int {
	if g.phase != RikkenPhasePlay || idx < 0 || idx >= len(g.players) {
		return nil
	}
	p := g.players[idx]
	all := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		all = append(all, i)
	}
	if len(g.trick) == 0 || len(all) == 0 {
		return all
	}
	if follow := filterByDesign(p, all, g.trick[0].Card.GetDesign()); len(follow) > 0 {
		return follow
	}
	return all
}

// PlayCard は人間が札を出す。
func (g *Rikken) PlayCard(cardIndex int) error {
	if g.phase != RikkenPhasePlay {
		return errors.New("いまはプレイの場面ではありません")
	}
	if !g.players[g.currentTurn].GetIsHuman() {
		return errors.New("あなたの番ではありません")
	}
	if !slices.Contains(g.GetValidPlayIndices(g.currentTurn), cardIndex) {
		return errors.New("その札は出せません（フォローの義務があります）")
	}
	g.playAt(g.currentTurn, cardIndex)
	g.advanceCpu()
	return nil
}

// playAt は席 idx に cardIndex を出させる。
func (g *Rikken) playAt(idx, cardIndex int) {
	card := g.players[idx].RemoveCard(cardIndex)
	if card == nil {
		return
	}
	g.trick = append(g.trick, &TrickCard{PlayerIdx: idx, Card: card})
	g.addLog(idx, "card", "札を出しました", []*Card{card})

	// **指名された札が出たら相方が判明します。**
	if g.calledCard != nil && !g.partnerRevealed &&
		card.GetDesign() == g.calledCard.GetDesign() && card.GetValue() == g.calledCard.GetValue() {
		g.partnerRevealed = true
		g.addLog(idx, "partner", "相方が判明しました", nil)
	}

	if len(g.trick) < RikkenPlayerCnt {
		g.currentTurn = (idx + 1) % RikkenPlayerCnt
		return
	}
	g.finishTrick()
}

// finishTrick はトリックを決着させる。
func (g *Rikken) finishTrick() {
	winner := ResolveTrickWinner(g.trick, g.trumpSuit, rikkenRank)
	won := make([]*Card, 0, len(g.trick))
	for _, tc := range g.trick {
		won = append(won, tc.Card)
	}
	g.players[winner].AddTrick(won)
	if g.IsDeclarerSide(winner) {
		g.declarerTricks++
	}

	g.lastTrick = g.trick
	g.lastTrickWinner = winner
	g.trick = nil
	g.trickCount++
	g.currentTurn = winner
	g.addLog(winner, "trick", "トリックを取りました", nil)

	if g.trickCount >= RikkenTrickCnt {
		g.finishRound()
	}
}

// rikkenRank は札の強さを返す。**エースが最強**です。
func rikkenRank(c *Card) int {
	if c == nil {
		return 0
	}
	if v := c.GetValue(); v == 1 {
		return 14
	}
	return c.GetValue()
}

// finishRound はラウンドを精算する。
//
// **合計は必ず 0。** 誰かが得た点は必ず誰かが失っています。
func (g *Rikken) finishRound() {
	made := g.contractMade()
	for i := range g.players {
		g.players[i].AddScore(g.roundScoreFor(i, made))
	}
	g.addLog(g.declarerIdx, "result", g.resultDetail(made), nil)

	g.phase = RikkenPhaseRoundEnd
	if g.roundNumber >= g.config.Rounds {
		g.finishGame()
	}
}

// contractMade は契約が達成されたかを返す。
func (g *Rikken) contractMade() bool {
	if RikkenIsMisere(g.contract) {
		// **1 トリックも取らないのが目標。**
		return g.declarerTricks == 0
	}
	return g.declarerTricks >= RikkenContractTarget(g.contract)
}

// roundScoreFor は席 i のこのラウンドの増減を返す。
//
// **ゼロサムになるように組み立てます。** 宣言側が受け取る合計と守備側が失う合計は
// 必ず一致します。
func (g *Rikken) roundScoreFor(i int, made bool) int {
	sign := -1
	if made {
		sign = 1
	}
	declarers, defenders := g.sideSizes()
	if declarers == 0 || defenders == 0 {
		return 0
	}

	// 守備側 1 人あたりの授受額を決め、宣言側はそれを人数比で受け取ります。
	perDefender := g.perDefenderStake()
	if g.IsDeclarerSide(i) {
		return sign * perDefender * defenders / declarers
	}
	return -sign * perDefender
}

// perDefenderStake は守備側 1 人あたりの授受額を返す。
func (g *Rikken) perDefenderStake() int {
	switch g.contract {
	case RikkenContractMisere:
		return RikkenMiserePoints
	case RikkenContractOpenMisere:
		return RikkenOpenMiserePoints
	case RikkenContractSolo:
		return RikkenSoloBase
	default:
		// Rik は 2 対 2 なので、超過トリックぶんを上乗せします。
		over := g.declarerTricks - RikkenRikTarget
		if over < 0 {
			over = -over
		}
		return RikkenRikBase + over
	}
}

// sideSizes は宣言側と守備側の人数を返す。
func (g *Rikken) sideSizes() (int, int) {
	declarers := 0
	for i := range g.players {
		if g.IsDeclarerSide(i) {
			declarers++
		}
	}
	return declarers, RikkenPlayerCnt - declarers
}

// resultDetail は精算の説明を返す。
func (g *Rikken) resultDetail(made bool) string {
	if made {
		return fmt.Sprintf("%s 成立（%d トリック）", RikkenContractName(g.contract), g.declarerTricks)
	}
	return fmt.Sprintf("%s 不成立（%d トリック）", RikkenContractName(g.contract), g.declarerTricks)
}

// finishGame は終局する。
func (g *Rikken) finishGame() {
	g.phase = RikkenPhaseGameEnd
	g.gameEndFlag = true
	best := 0
	for i := range g.players {
		if g.players[i].GetScore() > g.players[best].GetScore() {
			best = i
		}
	}
	g.winnerIdx = best
	g.addLog(best, "gameEnd", "いちばん点の高い席が勝ちです", nil)
}

// NextRound は次のラウンドを配る。
func (g *Rikken) NextRound() error {
	if g.gameEndFlag {
		return errors.New("ゲームは終了しています")
	}
	if g.phase != RikkenPhaseRoundEnd {
		return errors.New("いまはラウンドの区切りではありません")
	}
	g.dealerIdx = (g.dealerIdx + 1) % RikkenPlayerCnt
	g.startRound()
	return nil
}

// GiveUp は投了する。
func (g *Rikken) GiveUp() {
	if g.gameEndFlag {
		return
	}
	g.phase = RikkenPhaseGameEnd
	g.gameEndFlag = true
	best := 1
	for i := 1; i < len(g.players); i++ {
		if g.players[i].GetScore() > g.players[best].GetScore() {
			best = i
		}
	}
	g.winnerIdx = best
	g.addLog(0, "giveup", "投了しました", nil)
}

// CpuPlay は CPU の手番を進める。
func (g *Rikken) CpuPlay() { g.advanceCpu() }

// advanceCpu は人間の番になるまで CPU を進める。
func (g *Rikken) advanceCpu() {
	for range RikkenPlayerCnt * RikkenTrickCnt * 4 {
		if g.gameEndFlag || g.phase == RikkenPhaseRoundEnd {
			return
		}
		if g.currentTurn < 0 || g.currentTurn >= len(g.players) {
			return
		}
		if g.players[g.currentTurn].GetIsHuman() {
			return
		}
		switch g.phase {
		case RikkenPhaseBid:
			_ = g.bid(g.currentTurn, g.cpuChooseBid(g.currentTurn))
		case RikkenPhaseCall:
			_ = g.call(g.currentTurn, g.cpuChooseTrump(g.currentTurn))
		case RikkenPhasePlay:
			g.playAt(g.currentTurn, g.cpuChooseCard(g.currentTurn))
		default:
			return
		}
	}
}

// cpuChooseBid は CPU の競り。**手札が強いときだけ積みます。**
func (g *Rikken) cpuChooseBid(idx int) int {
	p := g.players[idx]
	aces, longest := 0, 0
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		n := 0
		for k := range p.GetCardsSize() {
			c := p.GetCard(k)
			if c.GetDesign() != suit {
				continue
			}
			n++
			if c.GetValue() == 1 {
				aces++
			}
		}
		if n > longest {
			longest = n
		}
	}
	want := RikkenContractNone
	switch {
	case aces >= 3 && longest >= 6:
		want = RikkenContractSolo
	case aces >= 2 && longest >= 5:
		want = RikkenContractRik
	}
	if want <= g.contract {
		return RikkenContractNone
	}
	return want
}

// cpuChooseTrump は CPU の切り札選び。いちばん長いスートにします。
func (g *Rikken) cpuChooseTrump(idx int) int {
	p := g.players[idx]
	best, bestLen := CardDesignSpade, -1
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		n := 0
		for k := range p.GetCardsSize() {
			if p.GetCard(k).GetDesign() == suit {
				n++
			}
		}
		if n > bestLen {
			best, bestLen = suit, n
		}
	}
	return best
}

// cpuChooseCard は CPU が出す札を選ぶ。
//
// **Misere 系を守るときは取らせにいきます。** 逆に自分が Misere を宣言している
// なら、絶対に取らないよういちばん安い札を出します。
func (g *Rikken) cpuChooseCard(idx int) int {
	valid := g.GetValidPlayIndices(idx)
	if len(valid) == 0 {
		return 0
	}
	p := g.players[idx]
	if RikkenIsMisere(g.contract) && idx == g.declarerIdx {
		return pickLowest(p, valid, rikkenRank)
	}
	if len(g.trick) == 0 {
		return pickHighest(p, valid, rikkenRank)
	}
	winner := ResolveTrickWinner(g.trick, g.trumpSuit, rikkenRank)
	if g.IsDeclarerSide(winner) == g.IsDeclarerSide(idx) {
		return pickLowest(p, valid, rikkenRank)
	}
	return pickHighest(p, valid, rikkenRank)
}

// rikkenSuitName はスート名を返す。
func rikkenSuitName(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "spade"
	case CardDesignClover:
		return "clover"
	case CardDesignHeart:
		return "heart"
	case CardDesignDiamond:
		return "diamond"
	default:
		return "notrump"
	}
}

// GetHint は人間への助言を返す。
func (g *Rikken) GetHint() *RikkenHint {
	if g.gameEndFlag || g.currentTurn != 0 || !g.players[0].GetIsHuman() {
		return nil
	}
	switch g.phase {
	case RikkenPhaseBid:
		c := g.cpuChooseBid(0)
		return &RikkenHint{Contract: &c, Reason: "rikkenBidStrength"}
	case RikkenPhasePlay:
		idx := g.cpuChooseCard(0)
		return &RikkenHint{CardIndex: &idx, Reason: "rikkenFollowSuit"}
	default:
		return nil
	}
}

// addLog は棋譜に 1 行足す。
func (g *Rikken) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.appendLog(playerIdx, actionType, detail, cards)
}

// --- Getters ---

// GetConfig は設定を返す。
func (g *Rikken) GetConfig() RikkenConfig { return g.config }

// SetConfig は設定を更新する。
func (g *Rikken) SetConfig(cfg RikkenConfig) {
	if cfg.Validate() != nil {
		return
	}
	g.config = cfg
}

// GetPhase は現在のフェーズを返す。
func (g *Rikken) GetPhase() int { return g.phase }

// GetGameEndFlag は終局フラグを返す。
func (g *Rikken) GetGameEndFlag() bool { return g.gameEndFlag }

// IsHumanTurn は人間の入力を待っているかを返す。
func (g *Rikken) IsHumanTurn() bool {
	if g.gameEndFlag || g.phase == RikkenPhaseRoundEnd {
		return false
	}
	return g.currentTurn == 0 && g.players[0].GetIsHuman()
}

// GetDealerIdx は親の席を返す。
func (g *Rikken) GetDealerIdx() int { return g.dealerIdx }

// GetContract は落札された契約を返す。
func (g *Rikken) GetContract() int { return g.contract }

// GetDeclarerIdx は落札者の席を返す (-1 = 未定)。
func (g *Rikken) GetDeclarerIdx() int { return g.declarerIdx }

// GetPartnerIdx は相方の席を返す (-1 = 相方なし)。
//
// **判明するまでは -1 を返します。** 指名した札が出るまで伏せます。
func (g *Rikken) GetPartnerIdx() int {
	if !g.partnerRevealed {
		return -1
	}
	return g.partnerIdx
}

// GetCalledCard は指名した札を返す (nil = 指名なし)。
func (g *Rikken) GetCalledCard() *Card { return g.calledCard }

// GetTrumpSuit は切り札を返す。
func (g *Rikken) GetTrumpSuit() int { return g.trumpSuit }

// HasPassed は席 i が降りたかを返す。
func (g *Rikken) HasPassed(i int) bool {
	if i < 0 || i >= len(g.passed) {
		return false
	}
	return g.passed[i]
}

// GetCurrentTurn はいまの手番の席を返す。
func (g *Rikken) GetCurrentTurn() int { return g.currentTurn }

// GetTrick はいま進行中のトリックを返す。
func (g *Rikken) GetTrick() []*TrickCard { return g.trick }

// GetLastTrick は直前に完成したトリックを返す。
func (g *Rikken) GetLastTrick() []*TrickCard { return g.lastTrick }

// GetLastTrickWinner は直前のトリックを取った席を返す。
func (g *Rikken) GetLastTrickWinner() int { return g.lastTrickWinner }

// GetTrickCount はこのラウンドで完成したトリック数を返す。
func (g *Rikken) GetTrickCount() int { return g.trickCount }

// GetDeclarerTricks は宣言側が取ったトリック数を返す。
func (g *Rikken) GetDeclarerTricks() int { return g.declarerTricks }

// GetRoundNumber はラウンド数を返す。
func (g *Rikken) GetRoundNumber() int { return g.roundNumber }

// GetPlayerCnt は人数を返す。
func (g *Rikken) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は席 i のプレイヤーを返す。
func (g *Rikken) GetPlayer(i int) *RikkenPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetWinnerIdx は勝者の席を返す (-1 = 未確定)。
func (g *Rikken) GetWinnerIdx() int { return g.winnerIdx }

// rikkenJSON is the JSON wire format for Rikken.
type rikkenJSON struct {
	TrumpCards      *TrumpCards       `json:"tc"`
	Players         []*RikkenPlayer   `json:"pl"`
	Config          RikkenConfig      `json:"cf"`
	Phase           int               `json:"ph"`
	DealerIdx       int               `json:"di"`
	Contract        int               `json:"ct"`
	DeclarerIdx     int               `json:"dc"`
	PartnerIdx      int               `json:"pi"`
	CalledCard      *Card             `json:"cc"`
	TrumpSuit       int               `json:"ts"`
	Passed          []bool            `json:"ps"`
	CurrentTurn     int               `json:"cu"`
	Trick           []*TrickCard      `json:"tk"`
	LastTrick       []*TrickCard      `json:"lt"`
	LastTrickWinner int               `json:"lw"`
	TrickCount      int               `json:"tn"`
	DeclarerTricks  int               `json:"dt"`
	PartnerRevealed bool              `json:"pr"`
	RoundNumber     int               `json:"rn"`
	GameEndFlag     bool              `json:"ge"`
	WinnerIdx       int               `json:"wi"`
	ActionLog       []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Rikken) MarshalJSON() ([]byte, error) {
	return json.Marshal(rikkenJSON{
		TrumpCards: g.trumpCards, Players: g.players, Config: g.config,
		Phase: g.phase, DealerIdx: g.dealerIdx, Contract: g.contract,
		DeclarerIdx: g.declarerIdx, PartnerIdx: g.partnerIdx, CalledCard: g.calledCard,
		TrumpSuit: g.trumpSuit, Passed: g.passed, CurrentTurn: g.currentTurn,
		Trick: g.trick, LastTrick: g.lastTrick, LastTrickWinner: g.lastTrickWinner,
		TrickCount: g.trickCount, DeclarerTricks: g.declarerTricks,
		PartnerRevealed: g.partnerRevealed, RoundNumber: g.roundNumber,
		GameEndFlag: g.gameEndFlag, WinnerIdx: g.winnerIdx, ActionLog: g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **範囲チェックだけでは足りません。** 「相方が居るのに Rik ではない」「Misere
// なのに切り札がある」は、どの値も単独では範囲内なのに盤面としてはあり得ません。
func (g *Rikken) UnmarshalJSON(data []byte) error {
	var j rikkenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := rikkenValidateWire(&j); err != nil {
		return err
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.dealerIdx = j.DealerIdx
	g.contract = j.Contract
	g.declarerIdx = j.DeclarerIdx
	g.partnerIdx = j.PartnerIdx
	g.calledCard = j.CalledCard
	g.trumpSuit = j.TrumpSuit
	g.passed = j.Passed
	if len(g.passed) != RikkenPlayerCnt {
		g.passed = make([]bool, RikkenPlayerCnt)
	}
	g.currentTurn = j.CurrentTurn
	g.trick = j.Trick
	g.lastTrick = j.LastTrick
	g.lastTrickWinner = j.LastTrickWinner
	g.trickCount = j.TrickCount
	g.declarerTricks = j.DeclarerTricks
	g.partnerRevealed = j.PartnerRevealed
	g.roundNumber = j.RoundNumber
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(rand.Int63()))
	}
	return nil
}

// rikkenValidateWire は復元しようとしている盤面の不変条件を検証する。
func rikkenValidateWire(j *rikkenJSON) error {
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if len(j.Players) != RikkenPlayerCnt {
		return fmt.Errorf("rikken: seat count %d does not match the fixed %d",
			len(j.Players), RikkenPlayerCnt)
	}
	for i, p := range j.Players {
		if p == nil {
			return fmt.Errorf("rikken: seat %d is missing", i)
		}
	}
	if j.Phase < RikkenPhaseBid || j.Phase > RikkenPhaseGameEnd {
		return fmt.Errorf("rikken: phase out of range: %d", j.Phase)
	}
	if j.GameEndFlag != (j.Phase == RikkenPhaseGameEnd) {
		return fmt.Errorf("rikken: the game-end flag and the phase disagree (flag=%v, phase=%d)",
			j.GameEndFlag, j.Phase)
	}
	if err := rikkenValidateContract(j.Contract); err != nil {
		return err
	}
	if len(j.ActionLog) > rikkenMaxSliceLen {
		return fmt.Errorf("rikken: action log too long: %d", len(j.ActionLog))
	}
	if len(j.Passed) != 0 && len(j.Passed) != RikkenPlayerCnt {
		return fmt.Errorf("rikken: pass flags hold %d slots for %d seats", len(j.Passed), RikkenPlayerCnt)
	}
	for name, seat := range map[string]int{"dealer": j.DealerIdx, "current turn": j.CurrentTurn} {
		if seat < 0 || seat >= RikkenPlayerCnt {
			return fmt.Errorf("rikken: %s index out of range: %d", name, seat)
		}
	}
	for name, seat := range map[string]int{
		"declarer": j.DeclarerIdx, "partner": j.PartnerIdx, "last trick winner": j.LastTrickWinner,
		"winner": j.WinnerIdx,
	} {
		if seat < -1 || seat >= RikkenPlayerCnt {
			return fmt.Errorf("rikken: %s index out of range: %d", name, seat)
		}
	}
	return rikkenValidateContractShape(j)
}

// rikkenValidateContractShape は契約と盤面の整合を検証する。
func rikkenValidateContractShape(j *rikkenJSON) error {
	// **相方が付くのは Rik だけ。**
	if j.PartnerIdx >= 0 && !RikkenHasPartner(j.Contract) {
		return fmt.Errorf("rikken: %s has no partner but seat %d is named as one",
			RikkenContractName(j.Contract), j.PartnerIdx)
	}
	if j.CalledCard != nil && !RikkenHasPartner(j.Contract) {
		return fmt.Errorf("rikken: %s does not call a card", RikkenContractName(j.Contract))
	}
	// **Misere 系に切り札はありません。** 範囲チェックは通ってしまう食い違いです。
	if RikkenIsMisere(j.Contract) && j.TrumpSuit != RikkenNoTrump {
		return fmt.Errorf("rikken: %s is played without trump but the suit is %d",
			RikkenContractName(j.Contract), j.TrumpSuit)
	}
	if j.TrumpSuit != RikkenNoTrump &&
		(j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return fmt.Errorf("rikken: trump suit out of range: %d", j.TrumpSuit)
	}
	// **相方が判明していないのに席が載っていることはあります**（内部では保持する）が、
	// 判明しているのに席が無いのは矛盾です。
	if j.PartnerRevealed && j.PartnerIdx < 0 {
		return errors.New("rikken: the partner is revealed but no seat is named")
	}
	if j.TrickCount < 0 || j.TrickCount > RikkenTrickCnt {
		return fmt.Errorf("rikken: trick count out of range: %d", j.TrickCount)
	}
	if j.DeclarerTricks < 0 || j.DeclarerTricks > j.TrickCount {
		return fmt.Errorf("rikken: the declaring side took %d of %d tricks",
			j.DeclarerTricks, j.TrickCount)
	}
	if len(j.Trick) >= RikkenPlayerCnt {
		return fmt.Errorf("rikken: the current trick holds %d cards, it resolves at %d",
			len(j.Trick), RikkenPlayerCnt)
	}
	if len(j.LastTrick) != 0 && len(j.LastTrick) != RikkenPlayerCnt {
		return fmt.Errorf("rikken: the last trick holds %d cards, want %d",
			len(j.LastTrick), RikkenPlayerCnt)
	}
	if j.RoundNumber < 0 || j.RoundNumber > j.Config.Rounds {
		return fmt.Errorf("rikken: round number out of range: %d", j.RoundNumber)
	}
	return nil
}
