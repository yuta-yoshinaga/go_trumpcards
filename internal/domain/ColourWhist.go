//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"slices"
)

// カラーホイストのフェーズ定数
const (
	// ColourWhistPhaseBid 契約を競っている場面
	ColourWhistPhaseBid = 0
	// ColourWhistPhaseCall 落札者が切り札 (と相方) を決める場面
	ColourWhistPhaseCall = 1
	// ColourWhistPhasePlay トリックを進めている場面
	ColourWhistPhasePlay = 2
	// ColourWhistPhaseRoundEnd ラウンドの精算が済んだ場面
	ColourWhistPhaseRoundEnd = 3
	// ColourWhistPhaseGameEnd ゲーム終了
	ColourWhistPhaseGameEnd = 4
)

// ColourWhistNoTrump は切り札なしを表すスート値 (どのスートとも一致しません)。
const ColourWhistNoTrump = -1

// colourWhistMaxSliceLen は復元時に受け付けるスライス長の上限。
const colourWhistMaxSliceLen = 2000

// ColourWhistHint は人間への助言。
type ColourWhistHint struct {
	// Contract は宣言すべき契約 (プレイ中は nil)。
	Contract *int
	// CardIndex は出すべき手札の位置 (競り中は nil)。
	CardIndex *int
	// Reason は助言の理由。
	Reason string
}

// ColourWhist はカラーホイスト (Colour Whist / Kleurenwiezen) のゲームクラス。
//
// ベルギー・フランドル地方の**契約制ホイスト**。標準 52 枚を 13 枚ずつ配ります。
//
// # 手札が契約を決めてしまうことがある
//
// このゲームの特徴は **Troel** です。**配った時点で誰かがエースを 3 枚持っていれば、
// 競りをせずに Troel が成立します**——4 枚目のエースの持ち主が自動的に相方になり、
// 2 対 2 で 8 トリックを目指します。競りで選ぶ契約ではなく、**手札の形が引き金**です。
//
// 該当が無ければ普通に競ります: Samen (相方を呼んで 8) < Alleen (単独 8) <
// Miserie (1 トリックも取らない)。
//
// # 得点はゼロサム
//
// 卓の得点の合計は常に 0 です。
type ColourWhist struct {
	trumpCards *TrumpCards
	players    []*ColourWhistPlayer
	config     ColourWhistConfig

	phase int
	// dealerIdx は親。競りは親の左隣から始まります。
	dealerIdx int
	// contract は成立した契約。
	contract int
	// declarerIdx は契約者 (-1 = 未定)。
	declarerIdx int
	// partnerIdx は相方 (-1 = 相方なし)。
	partnerIdx int
	// calledCard は Samen で指名した札。
	calledCard *Card
	// trumpSuit は切り札。
	trumpSuit int
	// passed[i] は席 i が降りたか。
	passed []bool
	// troelForced は配りで Troel が強制成立したか。
	troelForced bool

	currentTurn     int
	trick           []*TrickCard
	lastTrick       []*TrickCard
	lastTrickWinner int
	trickCount      int
	// declarerTricks は契約側が取ったトリック数。
	declarerTricks int
	// partnerRevealed は相方が判明したか。
	partnerRevealed bool

	roundNumber int
	gameEndFlag bool
	winnerIdx   int
	actionLogBase

	rng *rand.Rand
}

// NewColourWhist はコンストラクタ。
func NewColourWhist(trumpCards *TrumpCards, players []*ColourWhistPlayer, config ColourWhistConfig) *ColourWhist {
	if config.Validate() != nil {
		config = DefaultColourWhistConfig()
	}
	if players == nil {
		players = newColourWhistSeats()
	}
	if trumpCards == nil {
		trumpCards = NewTrumpCards(0)
	}
	return &ColourWhist{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		contract:        ColourWhistContractNone,
		declarerIdx:     -1,
		partnerIdx:      -1,
		trumpSuit:       ColourWhistNoTrump,
		lastTrickWinner: -1,
		winnerIdx:       -1,
		passed:          make([]bool, ColourWhistPlayerCnt),
		rng:             rand.New(rand.NewSource(rand.Int63())),
	}
}

// newColourWhistSeats は席 0 を人間、以降を CPU とした座席を作る。
func newColourWhistSeats() []*ColourWhistPlayer {
	seats := make([]*ColourWhistPlayer, ColourWhistPlayerCnt)
	for i := range seats {
		seats[i] = NewColourWhistPlayer(i == 0)
	}
	return seats
}

// NewDefaultColourWhist は既定設定の ColourWhist を返す。
func NewDefaultColourWhist() *ColourWhist {
	return NewColourWhist(NewTrumpCards(0), newColourWhistSeats(), DefaultColourWhistConfig())
}

// SetRand はテスト用に乱数源を差し替える。
func (g *ColourWhist) SetRand(r *rand.Rand) {
	if r != nil {
		g.rng = r
	}
}

// Reset はゲームを初期化する。
func (g *ColourWhist) Reset() {
	for _, p := range g.players {
		p.ResetGame()
	}
	g.roundNumber = 0
	g.dealerIdx = 0
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.actionLog = nil
	g.addLog(-1, "start", "カラーホイストを開始しました", nil)
	g.startRound()
}

// startRound は 1 ラウンドを配り直す。
func (g *ColourWhist) startRound() {
	g.trumpCards = NewTrumpCards(0)
	g.trumpCards.Shuffle()
	for _, p := range g.players {
		p.ResetRound()
	}
	for range ColourWhistHandSize {
		for i := range g.players {
			g.players[i].AddCard(g.trumpCards.DrawCard())
		}
	}

	g.phase = ColourWhistPhaseBid
	g.contract = ColourWhistContractNone
	g.declarerIdx = -1
	g.partnerIdx = -1
	g.calledCard = nil
	g.trumpSuit = ColourWhistNoTrump
	g.passed = make([]bool, ColourWhistPlayerCnt)
	g.troelForced = false
	g.trick = nil
	g.lastTrick = nil
	g.lastTrickWinner = -1
	g.trickCount = 0
	g.declarerTricks = 0
	g.partnerRevealed = false
	g.roundNumber++
	g.currentTurn = (g.dealerIdx + 1) % ColourWhistPlayerCnt

	g.addLog(g.dealerIdx, "deal", fmt.Sprintf("ラウンド %d を配りました", g.roundNumber), nil)

	// **配りが契約を決めてしまうことがある。** 競りより先に見ます。
	if g.detectTroel() {
		return
	}
	g.advanceCpu()
}

// detectTroel はエース 3 枚の持ち主を探し、居れば Troel を成立させる。
//
// **競りをしません。** 3 枚持っている人が契約者、4 枚目のエースの持ち主が相方です。
// 4 枚とも 1 人が持っていれば相方は付かず、単独で 8 トリックを目指します。
func (g *ColourWhist) detectTroel() bool {
	holder := -1
	for i, p := range g.players {
		if p.CountAces() >= ColourWhistTroelAces {
			holder = i
			break
		}
	}
	if holder < 0 {
		return false
	}

	g.contract = ColourWhistContractTroel
	g.declarerIdx = holder
	g.troelForced = true
	g.partnerIdx = g.fourthAceHolder(holder)
	// **4 枚とも持っていれば相方は居ません。** 単独で 8 トリックです。
	g.partnerRevealed = g.partnerIdx >= 0
	g.addLog(holder, "troel",
		fmt.Sprintf("席 %d がエースを %d 枚持っているため troel が成立しました",
			holder, g.players[holder].CountAces()), nil)

	g.currentTurn = holder
	g.phase = ColourWhistPhaseCall
	g.advanceCpu()
	return true
}

// fourthAceHolder は holder 以外でエースを持っている席を返す (-1 = 居ない)。
func (g *ColourWhist) fourthAceHolder(holder int) int {
	for i, p := range g.players {
		if i == holder {
			continue
		}
		if p.CountAces() > 0 {
			return i
		}
	}
	return -1
}

// Bid は契約を宣言する。ColourWhistContractNone を渡すとパス。
func (g *ColourWhist) Bid(contract int) error {
	if g.phase != ColourWhistPhaseBid {
		return errors.New("いまは競りの場面ではありません")
	}
	if !g.players[g.currentTurn].GetIsHuman() {
		return errors.New("あなたの番ではありません")
	}
	return g.bid(g.currentTurn, contract)
}

// bid は席 idx に競らせる。
func (g *ColourWhist) bid(idx, contract int) error {
	if contract != ColourWhistContractNone &&
		(contract < ColourWhistContractSamen || contract > ColourWhistBidMax) {
		// **Troel は競りに出てきません。** 配りでしか成立しません。
		return fmt.Errorf("競りで宣言できない契約です: %d", contract)
	}
	if contract == ColourWhistContractNone {
		g.passed[idx] = true
		g.addLog(idx, "pass", "パスしました", nil)
		g.advanceBid()
		return nil
	}
	if contract <= g.contract {
		return fmt.Errorf("いまの契約より強い宣言が要ります: %s", ColourWhistContractName(g.contract))
	}
	g.contract = contract
	g.declarerIdx = idx
	g.passed[idx] = false
	g.addLog(idx, "bid", ColourWhistContractName(contract)+" を宣言しました", nil)
	g.advanceBid()
	return nil
}

// advanceBid は競りを次の席へ回し、決着していれば次のフェーズへ進む。
func (g *ColourWhist) advanceBid() {
	if g.biddingDone() {
		g.finishBidding()
		return
	}
	for step := 1; step <= ColourWhistPlayerCnt; step++ {
		next := (g.currentTurn + step) % ColourWhistPlayerCnt
		if !g.passed[next] {
			g.currentTurn = next
			return
		}
	}
	g.finishBidding()
}

// biddingDone は競りが終わったかを返す。
func (g *ColourWhist) biddingDone() bool {
	alive := 0
	for _, p := range g.passed {
		if !p {
			alive++
		}
	}
	if g.contract == ColourWhistContractNone {
		return alive == 0
	}
	return alive <= 1
}

// finishBidding は競りを締めて次のフェーズへ進む。
func (g *ColourWhist) finishBidding() {
	if g.contract == ColourWhistContractNone {
		// **全員が降りたら親が Samen を引き受けます。** 流局にすると終わりません。
		g.declarerIdx = g.dealerIdx
		g.contract = ColourWhistContractSamen
		g.addLog(g.dealerIdx, "forced", "全員パスのため親が samen を引き受けます", nil)
	}
	g.currentTurn = g.declarerIdx
	if ColourWhistNeedsTrump(g.contract) {
		g.phase = ColourWhistPhaseCall
		g.advanceCpu()
		return
	}
	g.startPlay()
}

// Call は切り札を決め、Samen なら相方の札を指名する。
func (g *ColourWhist) Call(trumpSuit int) error {
	if g.phase != ColourWhistPhaseCall {
		return errors.New("いまは指名の場面ではありません")
	}
	if !g.players[g.currentTurn].GetIsHuman() {
		return errors.New("あなたの番ではありません")
	}
	return g.call(g.currentTurn, trumpSuit)
}

// call は席 idx に切り札 (と相方) を決めさせる。
func (g *ColourWhist) call(idx, trumpSuit int) error {
	if trumpSuit < CardDesignSpade || trumpSuit > CardDesignDiamond {
		return fmt.Errorf("切り札のスートが範囲外です: %d", trumpSuit)
	}
	g.trumpSuit = trumpSuit
	g.addLog(idx, "trump", colourWhistSuitName(trumpSuit)+" を切り札にしました", nil)

	// **Troel の相方は配りで決まっているので指名しません。**
	if g.contract == ColourWhistContractSamen {
		card := g.chooseCalledCard(idx)
		g.calledCard = card
		g.partnerIdx = g.holderOf(card)
	}
	g.startPlay()
	return nil
}

// chooseCalledCard は Samen で指名する札を選ぶ。
//
// **自分が持っていない札のうち、いちばん強いもの**を呼びます。
func (g *ColourWhist) chooseCalledCard(idx int) *Card {
	for _, value := range []int{1, 13, 12, 11} {
		for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
			c := NewCard(suit, value, false)
			if h := g.holderOf(c); h >= 0 && h != idx {
				return c
			}
		}
	}
	for i := range g.players {
		if i != idx && g.players[i].GetCardsSize() > 0 {
			return g.players[i].GetCard(0)
		}
	}
	return nil
}

// holderOf は札を持っている席を返す (-1 = 誰も持っていない)。
func (g *ColourWhist) holderOf(c *Card) int {
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
func (g *ColourWhist) startPlay() {
	g.phase = ColourWhistPhasePlay
	g.currentTurn = (g.dealerIdx + 1) % ColourWhistPlayerCnt
	g.addLog(-1, "play", ColourWhistContractName(g.contract)+" で開始します", nil)
	g.advanceCpu()
}

// IsDeclarerSide は席 idx が契約側かを返す。
func (g *ColourWhist) IsDeclarerSide(idx int) bool {
	if idx == g.declarerIdx {
		return true
	}
	return ColourWhistHasPartner(g.contract) && idx == g.partnerIdx
}

// GetValidPlayIndices は席 idx が出せる手札の位置を返す (フォロー義務のみ)。
func (g *ColourWhist) GetValidPlayIndices(idx int) []int {
	if g.phase != ColourWhistPhasePlay || idx < 0 || idx >= len(g.players) {
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
func (g *ColourWhist) PlayCard(cardIndex int) error {
	if g.phase != ColourWhistPhasePlay {
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
func (g *ColourWhist) playAt(idx, cardIndex int) {
	card := g.players[idx].RemoveCard(cardIndex)
	if card == nil {
		return
	}
	g.trick = append(g.trick, &TrickCard{PlayerIdx: idx, Card: card})
	g.addLog(idx, "card", "札を出しました", []*Card{card})

	if g.calledCard != nil && !g.partnerRevealed &&
		card.GetDesign() == g.calledCard.GetDesign() && card.GetValue() == g.calledCard.GetValue() {
		g.partnerRevealed = true
		g.addLog(idx, "partner", "相方が判明しました", nil)
	}

	if len(g.trick) < ColourWhistPlayerCnt {
		g.currentTurn = (idx + 1) % ColourWhistPlayerCnt
		return
	}
	g.finishTrick()
}

// finishTrick はトリックを決着させる。
func (g *ColourWhist) finishTrick() {
	winner := ResolveTrickWinner(g.trick, g.trumpSuit, colourWhistRank)
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

	if g.trickCount >= ColourWhistTrickCnt {
		g.finishRound()
	}
}

// colourWhistRank は札の強さを返す。**エースが最強**です。
func colourWhistRank(c *Card) int {
	if c == nil {
		return 0
	}
	if v := c.GetValue(); v == 1 {
		return 14
	}
	return c.GetValue()
}

// finishRound はラウンドを精算する。**合計は必ず 0。**
func (g *ColourWhist) finishRound() {
	made := g.contractMade()
	for i := range g.players {
		g.players[i].AddScore(g.roundScoreFor(i, made))
	}
	g.addLog(g.declarerIdx, "result", g.resultDetail(made), nil)

	g.phase = ColourWhistPhaseRoundEnd
	if g.roundNumber >= g.config.Rounds {
		g.finishGame()
	}
}

// contractMade は契約が達成されたかを返す。
func (g *ColourWhist) contractMade() bool {
	if ColourWhistIsMiserie(g.contract) {
		return g.declarerTricks == 0
	}
	return g.declarerTricks >= ColourWhistContractTarget(g.contract)
}

// roundScoreFor は席 i のこのラウンドの増減を返す (ゼロサム)。
func (g *ColourWhist) roundScoreFor(i int, made bool) int {
	sign := -1
	if made {
		sign = 1
	}
	declarers, defenders := g.sideSizes()
	if declarers == 0 || defenders == 0 {
		return 0
	}
	perDefender := g.perDefenderStake()
	if g.IsDeclarerSide(i) {
		return sign * perDefender * defenders / declarers
	}
	return -sign * perDefender
}

// perDefenderStake は守備側 1 人あたりの授受額を返す。
func (g *ColourWhist) perDefenderStake() int {
	switch g.contract {
	case ColourWhistContractMiserie:
		return ColourWhistMiseriePoints
	case ColourWhistContractAlleen:
		return ColourWhistAlleenBase
	default:
		// Samen / Troel は 2 対 2。超過トリックぶんを上乗せします。
		over := g.declarerTricks - ColourWhistTrickTarget
		if over < 0 {
			over = -over
		}
		return ColourWhistSamenBase + over
	}
}

// sideSizes は契約側と守備側の人数を返す。
func (g *ColourWhist) sideSizes() (int, int) {
	declarers := 0
	for i := range g.players {
		if g.IsDeclarerSide(i) {
			declarers++
		}
	}
	return declarers, ColourWhistPlayerCnt - declarers
}

// resultDetail は精算の説明を返す。
func (g *ColourWhist) resultDetail(made bool) string {
	state := "不成立"
	if made {
		state = "成立"
	}
	return fmt.Sprintf("%s %s（%d トリック）", ColourWhistContractName(g.contract), state, g.declarerTricks)
}

// finishGame は終局する。
func (g *ColourWhist) finishGame() {
	g.phase = ColourWhistPhaseGameEnd
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
func (g *ColourWhist) NextRound() error {
	if g.gameEndFlag {
		return errors.New("ゲームは終了しています")
	}
	if g.phase != ColourWhistPhaseRoundEnd {
		return errors.New("いまはラウンドの区切りではありません")
	}
	g.dealerIdx = (g.dealerIdx + 1) % ColourWhistPlayerCnt
	g.startRound()
	return nil
}

// GiveUp は投了する。
func (g *ColourWhist) GiveUp() {
	if g.gameEndFlag {
		return
	}
	g.phase = ColourWhistPhaseGameEnd
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
func (g *ColourWhist) CpuPlay() { g.advanceCpu() }

// advanceCpu は人間の番になるまで CPU を進める。
func (g *ColourWhist) advanceCpu() {
	for range ColourWhistPlayerCnt * ColourWhistTrickCnt * 4 {
		if g.gameEndFlag || g.phase == ColourWhistPhaseRoundEnd {
			return
		}
		if g.currentTurn < 0 || g.currentTurn >= len(g.players) {
			return
		}
		if g.players[g.currentTurn].GetIsHuman() {
			return
		}
		switch g.phase {
		case ColourWhistPhaseBid:
			_ = g.bid(g.currentTurn, g.cpuChooseBid(g.currentTurn))
		case ColourWhistPhaseCall:
			_ = g.call(g.currentTurn, g.cpuChooseTrump(g.currentTurn))
		case ColourWhistPhasePlay:
			g.playAt(g.currentTurn, g.cpuChooseCard(g.currentTurn))
		default:
			return
		}
	}
}

// cpuChooseBid は CPU の競り。**手札が強いときだけ積みます。**
func (g *ColourWhist) cpuChooseBid(idx int) int {
	p := g.players[idx]
	aces, longest := p.CountAces(), 0
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		n := 0
		for k := range p.GetCardsSize() {
			if p.GetCard(k).GetDesign() == suit {
				n++
			}
		}
		if n > longest {
			longest = n
		}
	}
	want := ColourWhistContractNone
	switch {
	case aces >= 2 && longest >= 6:
		want = ColourWhistContractAlleen
	case aces >= 1 && longest >= 5:
		want = ColourWhistContractSamen
	}
	if want <= g.contract {
		return ColourWhistContractNone
	}
	return want
}

// cpuChooseTrump は CPU の切り札選び。いちばん長いスートにします。
func (g *ColourWhist) cpuChooseTrump(idx int) int {
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
func (g *ColourWhist) cpuChooseCard(idx int) int {
	valid := g.GetValidPlayIndices(idx)
	if len(valid) == 0 {
		return 0
	}
	p := g.players[idx]
	if ColourWhistIsMiserie(g.contract) && idx == g.declarerIdx {
		return pickLowest(p, valid, colourWhistRank)
	}
	if len(g.trick) == 0 {
		return pickHighest(p, valid, colourWhistRank)
	}
	winner := ResolveTrickWinner(g.trick, g.trumpSuit, colourWhistRank)
	if g.IsDeclarerSide(winner) == g.IsDeclarerSide(idx) {
		return pickLowest(p, valid, colourWhistRank)
	}
	return pickHighest(p, valid, colourWhistRank)
}

// colourWhistSuitName はスート名を返す。
func colourWhistSuitName(suit int) string {
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
func (g *ColourWhist) GetHint() *ColourWhistHint {
	if g.gameEndFlag || g.currentTurn != 0 || !g.players[0].GetIsHuman() {
		return nil
	}
	switch g.phase {
	case ColourWhistPhaseBid:
		c := g.cpuChooseBid(0)
		return &ColourWhistHint{Contract: &c, Reason: "colourWhistBidStrength"}
	case ColourWhistPhasePlay:
		idx := g.cpuChooseCard(0)
		return &ColourWhistHint{CardIndex: &idx, Reason: "colourWhistFollowSuit"}
	default:
		return nil
	}
}

// addLog は棋譜に 1 行足す。
func (g *ColourWhist) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.appendLog(playerIdx, actionType, detail, cards)
}

// --- Getters ---

// GetConfig は設定を返す。
func (g *ColourWhist) GetConfig() ColourWhistConfig { return g.config }

// SetConfig は設定を更新する。
func (g *ColourWhist) SetConfig(cfg ColourWhistConfig) {
	if cfg.Validate() != nil {
		return
	}
	g.config = cfg
}

// GetPhase は現在のフェーズを返す。
func (g *ColourWhist) GetPhase() int { return g.phase }

// GetGameEndFlag は終局フラグを返す。
func (g *ColourWhist) GetGameEndFlag() bool { return g.gameEndFlag }

// IsHumanTurn は人間の入力を待っているかを返す。
func (g *ColourWhist) IsHumanTurn() bool {
	if g.gameEndFlag || g.phase == ColourWhistPhaseRoundEnd {
		return false
	}
	return g.currentTurn == 0 && g.players[0].GetIsHuman()
}

// GetDealerIdx は親の席を返す。
func (g *ColourWhist) GetDealerIdx() int { return g.dealerIdx }

// GetContract は成立した契約を返す。
func (g *ColourWhist) GetContract() int { return g.contract }

// GetDeclarerIdx は契約者の席を返す (-1 = 未定)。
func (g *ColourWhist) GetDeclarerIdx() int { return g.declarerIdx }

// GetPartnerIdx は相方の席を返す (-1 = 相方なし、または未公開)。
//
// **Troel の相方は最初から公開**です（配りで決まるため）。Samen の相方は
// 指名した札が出るまで伏せます。
func (g *ColourWhist) GetPartnerIdx() int {
	if !g.partnerRevealed {
		return -1
	}
	return g.partnerIdx
}

// GetCalledCard は指名した札を返す (nil = 指名なし)。
func (g *ColourWhist) GetCalledCard() *Card { return g.calledCard }

// GetTrumpSuit は切り札を返す。
func (g *ColourWhist) GetTrumpSuit() int { return g.trumpSuit }

// IsTroelForced は配りで Troel が強制成立したかを返す。
func (g *ColourWhist) IsTroelForced() bool { return g.troelForced }

// HasPassed は席 i が降りたかを返す。
func (g *ColourWhist) HasPassed(i int) bool {
	if i < 0 || i >= len(g.passed) {
		return false
	}
	return g.passed[i]
}

// GetCurrentTurn はいまの手番の席を返す。
func (g *ColourWhist) GetCurrentTurn() int { return g.currentTurn }

// GetTrick はいま進行中のトリックを返す。
func (g *ColourWhist) GetTrick() []*TrickCard { return g.trick }

// GetLastTrick は直前に完成したトリックを返す。
func (g *ColourWhist) GetLastTrick() []*TrickCard { return g.lastTrick }

// GetLastTrickWinner は直前のトリックを取った席を返す。
func (g *ColourWhist) GetLastTrickWinner() int { return g.lastTrickWinner }

// GetTrickCount はこのラウンドで完成したトリック数を返す。
func (g *ColourWhist) GetTrickCount() int { return g.trickCount }

// GetDeclarerTricks は契約側が取ったトリック数を返す。
func (g *ColourWhist) GetDeclarerTricks() int { return g.declarerTricks }

// GetRoundNumber はラウンド数を返す。
func (g *ColourWhist) GetRoundNumber() int { return g.roundNumber }

// GetPlayerCnt は人数を返す。
func (g *ColourWhist) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は席 i のプレイヤーを返す。
func (g *ColourWhist) GetPlayer(i int) *ColourWhistPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetWinnerIdx は勝者の席を返す (-1 = 未確定)。
func (g *ColourWhist) GetWinnerIdx() int { return g.winnerIdx }

// colourWhistJSON is the JSON wire format for ColourWhist.
type colourWhistJSON struct {
	TrumpCards      *TrumpCards          `json:"tc"`
	Players         []*ColourWhistPlayer `json:"pl"`
	Config          ColourWhistConfig    `json:"cf"`
	Phase           int                  `json:"ph"`
	DealerIdx       int                  `json:"di"`
	Contract        int                  `json:"ct"`
	DeclarerIdx     int                  `json:"dc"`
	PartnerIdx      int                  `json:"pi"`
	CalledCard      *Card                `json:"cc"`
	TrumpSuit       int                  `json:"ts"`
	Passed          []bool               `json:"ps"`
	TroelForced     bool                 `json:"tf"`
	CurrentTurn     int                  `json:"cu"`
	Trick           []*TrickCard         `json:"tk"`
	LastTrick       []*TrickCard         `json:"lt"`
	LastTrickWinner int                  `json:"lw"`
	TrickCount      int                  `json:"tn"`
	DeclarerTricks  int                  `json:"dt"`
	PartnerRevealed bool                 `json:"pr"`
	RoundNumber     int                  `json:"rn"`
	GameEndFlag     bool                 `json:"ge"`
	WinnerIdx       int                  `json:"wi"`
	ActionLog       []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *ColourWhist) MarshalJSON() ([]byte, error) {
	return json.Marshal(colourWhistJSON{
		TrumpCards: g.trumpCards, Players: g.players, Config: g.config,
		Phase: g.phase, DealerIdx: g.dealerIdx, Contract: g.contract,
		DeclarerIdx: g.declarerIdx, PartnerIdx: g.partnerIdx, CalledCard: g.calledCard,
		TrumpSuit: g.trumpSuit, Passed: g.passed, TroelForced: g.troelForced,
		CurrentTurn: g.currentTurn, Trick: g.trick, LastTrick: g.lastTrick,
		LastTrickWinner: g.lastTrickWinner, TrickCount: g.trickCount,
		DeclarerTricks: g.declarerTricks, PartnerRevealed: g.partnerRevealed,
		RoundNumber: g.roundNumber, GameEndFlag: g.gameEndFlag,
		WinnerIdx: g.winnerIdx, ActionLog: g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **範囲チェックだけでは足りません。** 「Troel なのに競りで降りた記録がある」
// 「Miserie なのに切り札がある」は、どの値も単独では範囲内なのに盤面としては
// あり得ない状態です。
func (g *ColourWhist) UnmarshalJSON(data []byte) error {
	var j colourWhistJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := colourWhistValidateWire(&j); err != nil {
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
	if len(g.passed) != ColourWhistPlayerCnt {
		g.passed = make([]bool, ColourWhistPlayerCnt)
	}
	g.troelForced = j.TroelForced
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

// colourWhistValidateWire は復元しようとしている盤面の不変条件を検証する。
func colourWhistValidateWire(j *colourWhistJSON) error {
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if len(j.Players) != ColourWhistPlayerCnt {
		return fmt.Errorf("colourwhist: seat count %d does not match the fixed %d",
			len(j.Players), ColourWhistPlayerCnt)
	}
	for i, p := range j.Players {
		if p == nil {
			return fmt.Errorf("colourwhist: seat %d is missing", i)
		}
	}
	if j.Phase < ColourWhistPhaseBid || j.Phase > ColourWhistPhaseGameEnd {
		return fmt.Errorf("colourwhist: phase out of range: %d", j.Phase)
	}
	if j.GameEndFlag != (j.Phase == ColourWhistPhaseGameEnd) {
		return fmt.Errorf("colourwhist: the game-end flag and the phase disagree (flag=%v, phase=%d)",
			j.GameEndFlag, j.Phase)
	}
	if err := colourWhistValidateContract(j.Contract); err != nil {
		return err
	}
	if len(j.ActionLog) > colourWhistMaxSliceLen {
		return fmt.Errorf("colourwhist: action log too long: %d", len(j.ActionLog))
	}
	if len(j.Passed) != 0 && len(j.Passed) != ColourWhistPlayerCnt {
		return fmt.Errorf("colourwhist: pass flags hold %d slots for %d seats",
			len(j.Passed), ColourWhistPlayerCnt)
	}
	for name, seat := range map[string]int{"dealer": j.DealerIdx, "current turn": j.CurrentTurn} {
		if seat < 0 || seat >= ColourWhistPlayerCnt {
			return fmt.Errorf("colourwhist: %s index out of range: %d", name, seat)
		}
	}
	for name, seat := range map[string]int{
		"declarer": j.DeclarerIdx, "partner": j.PartnerIdx,
		"last trick winner": j.LastTrickWinner, "winner": j.WinnerIdx,
	} {
		if seat < -1 || seat >= ColourWhistPlayerCnt {
			return fmt.Errorf("colourwhist: %s index out of range: %d", name, seat)
		}
	}
	return colourWhistValidateContractShape(j)
}

// colourWhistValidateContractShape は契約と盤面の整合を検証する。
func colourWhistValidateContractShape(j *colourWhistJSON) error {
	if j.PartnerIdx >= 0 && !ColourWhistHasPartner(j.Contract) {
		return fmt.Errorf("colourwhist: %s has no partner but seat %d is named as one",
			ColourWhistContractName(j.Contract), j.PartnerIdx)
	}
	// **Troel は配りで決まるので、指名した札はありません。**
	if j.CalledCard != nil && j.Contract != ColourWhistContractSamen {
		return fmt.Errorf("colourwhist: %s does not call a card", ColourWhistContractName(j.Contract))
	}
	if ColourWhistIsMiserie(j.Contract) && j.TrumpSuit != ColourWhistNoTrump {
		return fmt.Errorf("colourwhist: %s is played without trump but the suit is %d",
			ColourWhistContractName(j.Contract), j.TrumpSuit)
	}
	if j.TrumpSuit != ColourWhistNoTrump &&
		(j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return fmt.Errorf("colourwhist: trump suit out of range: %d", j.TrumpSuit)
	}
	// **Troel は競りをしません。** 降りた記録があれば矛盾です。
	if j.TroelForced {
		if j.Contract != ColourWhistContractTroel {
			return fmt.Errorf("colourwhist: the troel flag is set but the contract is %s",
				ColourWhistContractName(j.Contract))
		}
		for i, passed := range j.Passed {
			if passed {
				return fmt.Errorf("colourwhist: troel skips the auction but seat %d is marked as passed", i)
			}
		}
	}
	if j.PartnerRevealed && j.PartnerIdx < 0 {
		return errors.New("colourwhist: the partner is revealed but no seat is named")
	}
	if j.TrickCount < 0 || j.TrickCount > ColourWhistTrickCnt {
		return fmt.Errorf("colourwhist: trick count out of range: %d", j.TrickCount)
	}
	if j.DeclarerTricks < 0 || j.DeclarerTricks > j.TrickCount {
		return fmt.Errorf("colourwhist: the declaring side took %d of %d tricks",
			j.DeclarerTricks, j.TrickCount)
	}
	if len(j.Trick) >= ColourWhistPlayerCnt {
		return fmt.Errorf("colourwhist: the current trick holds %d cards, it resolves at %d",
			len(j.Trick), ColourWhistPlayerCnt)
	}
	if len(j.LastTrick) != 0 && len(j.LastTrick) != ColourWhistPlayerCnt {
		return fmt.Errorf("colourwhist: the last trick holds %d cards, want %d",
			len(j.LastTrick), ColourWhistPlayerCnt)
	}
	if j.RoundNumber < 0 || j.RoundNumber > j.Config.Rounds {
		return fmt.Errorf("colourwhist: round number out of range: %d", j.RoundNumber)
	}
	return nil
}
