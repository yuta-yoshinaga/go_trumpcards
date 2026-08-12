//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"slices"
)

// ボティファラのフェーズ定数
const (
	// BotifarraPhaseDeclare 切り札を宣言するか相方へ委ねる場面
	BotifarraPhaseDeclare = 0
	// BotifarraPhaseDelegated 委ねられた側が切り札を宣言する場面
	BotifarraPhaseDelegated = 1
	// BotifarraPhaseDouble 相手側が倍付けを宣言できる場面
	BotifarraPhaseDouble = 2
	// BotifarraPhasePlay トリックを進めている場面
	BotifarraPhasePlay = 3
	// BotifarraPhaseRoundEnd ラウンドの精算が済んだ場面
	BotifarraPhaseRoundEnd = 4
	// BotifarraPhaseGameEnd ゲーム終了
	BotifarraPhaseGameEnd = 5
)

// BotifarraNoTrump は切り札なし (ボティファラ宣言) を表すスート値。
//
// **どのスートの値とも一致しない負の値**にしてあります。ResolveTrickWinner は
// 切り札スートを札の design と比べるだけなので、これで「切り札なし」になります。
const BotifarraNoTrump = -1

// 倍付けの段階。**この配りの価値に掛かります**（36 を超えた側の得点に乗ります）。
const (
	BotifarraMultiplierNone     = 1
	BotifarraMultiplierContrar  = 2
	BotifarraMultiplierRecontra = 4
)

// botifarraMaxSliceLen は復元時に受け付けるスライス長の上限。
const botifarraMaxSliceLen = 2000

// errBotifarraMissingBase は復元時に土台のプレイヤーが欠けていたときのエラー。
var errBotifarraMissingBase = errors.New("botifarra player is missing its base player")

// BotifarraHint は人間への助言。
type BotifarraHint struct {
	// CardIndex は出すべき手札の位置 (宣言の場面では nil)。
	CardIndex *int
	// Suit は宣言すべき切り札 (プレイ中は nil)。
	Suit *int
	// Reason は助言の理由。
	Reason string
}

// Botifarra はボティファラ (Botifarra) のゲームクラス。
//
// カタルーニャで日常的に遊ばれる **2 対 2 のパートナー制トリックテイキング**。
// スペイン式 48 枚 (10 を含まない) を 12 枚ずつ配ります。
//
// # 競りではなく「宣言」
//
// 入札はありません。**親が切り札を宣言するか、相方に委ねる (delegar)** かの
// 二択だけです。委ねられた相方は必ず宣言します——降りる選択肢はありません。
// 切り札なし (ボティファラ) も宣言できます。
//
// # 勝てるなら勝たなければならない
//
// フォロー義務に加えて、**そのトリックに勝てる札があるなら勝たなければなりません**
// (味方が勝っている場合を除く)。切り札を持っていて場が切り札で取られているなら、
// 上から被せる義務まであります。ここが普通のトリックテイキングと違うところで、
// 「安い札を温存する」という選択が効きません。
//
// # 得点
//
// 札の点は**マニラ (9) が 5 点**、アス 4、王 3、騎士 2、ソタ 1 で 1 スート 15 点。
// 4 スートで 60 点、これに**各トリック 1 点**が加わって 1 ラウンド 72 点です。
// ちょうど半分の 36 を超えたぶんだけが得点になります。
type Botifarra struct {
	trumpCards *TrumpCards
	players    []*BotifarraPlayer
	config     BotifarraConfig

	phase int
	// dealerIdx は親。宣言する権利を持ちます。
	dealerIdx int
	// declarerIdx は実際に切り札を決めた席。
	declarerIdx int
	// trumpSuit は切り札 (BotifarraNoTrump = 切り札なし)。
	trumpSuit int
	// multiplier は倍付け。
	multiplier int

	// currentTurn はいま出す番の席。
	currentTurn int
	// trick はいま進行中のトリック。
	trick []*TrickCard
	// lastTrick は直前に完成したトリック (表示用)。
	lastTrick []*TrickCard
	// lastTrickWinner は直前のトリックを取った席 (-1 = まだ無い)。
	lastTrickWinner int
	// trickCount はこのラウンドで完成したトリック数。
	trickCount int

	// roundPoints[team] はこのラウンドでチームが取った点。
	roundPoints [2]int
	// scores[team] は通算得点。
	scores [2]int

	gameEndFlag bool
	// winnerTeam は勝ったチーム (-1 = 未確定)。
	winnerTeam int
	actionLogBase

	rng *rand.Rand
}

// NewBotifarra はコンストラクタ。
func NewBotifarra(trumpCards *TrumpCards, players []*BotifarraPlayer, config BotifarraConfig) *Botifarra {
	if config.Validate() != nil {
		config = DefaultBotifarraConfig()
	}
	if players == nil {
		players = newBotifarraSeats()
	}
	if trumpCards == nil {
		trumpCards = NewTrumpCardsReversis()
	}
	return &Botifarra{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		trumpSuit:       BotifarraNoTrump,
		multiplier:      BotifarraMultiplierNone,
		declarerIdx:     -1,
		lastTrickWinner: -1,
		winnerTeam:      -1,
		rng:             rand.New(rand.NewSource(rand.Int63())),
	}
}

// newBotifarraSeats は席 0 を人間、以降を CPU とした座席を作る。
func newBotifarraSeats() []*BotifarraPlayer {
	seats := make([]*BotifarraPlayer, BotifarraPlayerCnt)
	for i := range seats {
		seats[i] = NewBotifarraPlayer(i == 0)
	}
	return seats
}

// NewDefaultBotifarra は既定設定の Botifarra を返す。
//
// **デッキは `NewTrumpCardsReversis()`。** 名前は Reversis ですが中身は
// A,2〜9,J,Q,K × 4 = 48 枚のスペイン式そのもので、4 人に 12 枚ずつ配ると
// ちょうど 0 枚残ります。無タグのファイルにあるのでどのバケットからでも使えます。
func NewDefaultBotifarra() *Botifarra {
	return NewBotifarra(NewTrumpCardsReversis(), newBotifarraSeats(), DefaultBotifarraConfig())
}

// SetRand はテスト用に乱数源を差し替える。
func (g *Botifarra) SetRand(r *rand.Rand) {
	if r != nil {
		g.rng = r
	}
}

// Reset はゲームを初期化する。
func (g *Botifarra) Reset() {
	g.scores = [2]int{}
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.dealerIdx = 0
	g.actionLog = nil
	g.addLog(-1, "start", "ボティファラを開始しました", nil)
	g.startRound()
}

// startRound は 1 ラウンドを配り直す。
func (g *Botifarra) startRound() {
	g.trumpCards = NewTrumpCardsReversis()
	g.trumpCards.Shuffle()
	for _, p := range g.players {
		p.ResetRound()
	}
	for range BotifarraHandSize {
		for i := range g.players {
			g.players[i].AddCard(g.trumpCards.DrawCard())
		}
	}

	g.phase = BotifarraPhaseDeclare
	g.trumpSuit = BotifarraNoTrump
	g.declarerIdx = -1
	g.multiplier = BotifarraMultiplierNone
	g.trick = nil
	g.lastTrick = nil
	g.lastTrickWinner = -1
	g.trickCount = 0
	g.roundPoints = [2]int{}
	g.currentTurn = g.dealerIdx

	g.addLog(g.dealerIdx, "deal",
		fmt.Sprintf("席 %d が親です。切り札を宣言するか相方に委ねます", g.dealerIdx), nil)
	g.advanceCpu()
}

// Declare は切り札を宣言する。suit に BotifarraNoTrump を渡すと切り札なし。
func (g *Botifarra) Declare(suit int) error {
	if g.phase != BotifarraPhaseDeclare && g.phase != BotifarraPhaseDelegated {
		return errors.New("いまは宣言の場面ではありません")
	}
	if !g.players[g.currentTurn].GetIsHuman() {
		return errors.New("あなたの宣言の番ではありません")
	}
	return g.declare(g.currentTurn, suit)
}

// declare は席 idx に切り札を決めさせる。
func (g *Botifarra) declare(idx, suit int) error {
	if suit != BotifarraNoTrump &&
		(suit < CardDesignSpade || suit > CardDesignDiamond) {
		return fmt.Errorf("切り札のスートが範囲外です: %d", suit)
	}
	g.trumpSuit = suit
	g.declarerIdx = idx
	g.addLog(idx, "declare", botifarraTrumpName(suit)+" を宣言しました", nil)
	g.openDoubling()
	return nil
}

// Delegate は親が宣言を相方に委ねる。
func (g *Botifarra) Delegate() error {
	if g.phase != BotifarraPhaseDeclare {
		return errors.New("いまは委ねられる場面ではありません")
	}
	if !g.players[g.currentTurn].GetIsHuman() {
		return errors.New("あなたの宣言の番ではありません")
	}
	g.delegate(g.currentTurn)
	return nil
}

// delegate は宣言を相方に渡す。
//
// **委ねられた側は必ず宣言します。** 降りる選択肢はありません。
func (g *Botifarra) delegate(idx int) {
	partner := BotifarraPartnerOf(idx)
	g.currentTurn = partner
	g.phase = BotifarraPhaseDelegated
	g.addLog(idx, "delegate", fmt.Sprintf("席 %d に宣言を委ねました", partner), nil)
	g.advanceCpu()
}

// openDoubling は倍付けの場面へ進む (許していなければ飛ばす)。
func (g *Botifarra) openDoubling() {
	if !g.config.AllowDoubling {
		g.startPlay()
		return
	}
	g.phase = BotifarraPhaseDouble
	// **倍付けを言えるのは契約していない側。** 親の左隣から見ます。
	g.currentTurn = g.firstOpponentOf(g.declarerIdx)
	g.advanceCpu()
}

// firstOpponentOf は idx の相手チームで、親にいちばん近い席を返す。
func (g *Botifarra) firstOpponentOf(idx int) int {
	for step := 1; step < BotifarraPlayerCnt; step++ {
		seat := (idx + step) % BotifarraPlayerCnt
		if BotifarraTeamOf(seat) != BotifarraTeamOf(idx) {
			return seat
		}
	}
	return (idx + 1) % BotifarraPlayerCnt
}

// Double は倍付けを宣言する。
func (g *Botifarra) Double() error {
	if g.phase != BotifarraPhaseDouble {
		return errors.New("いまは倍付けの場面ではありません")
	}
	if !g.players[g.currentTurn].GetIsHuman() {
		return errors.New("あなたの番ではありません")
	}
	return g.double(g.currentTurn)
}

// double は席 idx に倍付けさせる。
func (g *Botifarra) double(idx int) error {
	switch g.multiplier {
	case BotifarraMultiplierNone:
		g.multiplier = BotifarraMultiplierContrar
		g.addLog(idx, "contrar", "contrar（2 倍）を宣言しました", nil)
		// **相手が倍にしたら、契約側は再倍にできます。**
		g.currentTurn = g.firstOpponentOf(idx)
		g.advanceCpu()
		return nil
	case BotifarraMultiplierContrar:
		g.multiplier = BotifarraMultiplierRecontra
		g.addLog(idx, "recontrar", "recontrar（4 倍）を宣言しました", nil)
		g.startPlay()
		return nil
	default:
		return errors.New("これ以上は倍にできません")
	}
}

// PassDouble は倍付けせずに進む。
func (g *Botifarra) PassDouble() error {
	if g.phase != BotifarraPhaseDouble {
		return errors.New("いまは倍付けの場面ではありません")
	}
	if !g.players[g.currentTurn].GetIsHuman() {
		return errors.New("あなたの番ではありません")
	}
	g.passDouble(g.currentTurn)
	return nil
}

// passDouble は倍付けを見送る。
func (g *Botifarra) passDouble(idx int) {
	g.addLog(idx, "pass", "倍付けを見送りました", nil)
	g.startPlay()
}

// startPlay はトリックの場面へ進む。親のリードで始まります。
func (g *Botifarra) startPlay() {
	g.phase = BotifarraPhasePlay
	g.currentTurn = g.dealerIdx
	g.addLog(-1, "play", "プレイを開始します（切り札: "+botifarraTrumpName(g.trumpSuit)+"）", nil)
	g.advanceCpu()
}

// GetValidPlayIndices は席 idx が出せる手札の位置を返す。
//
// **勝てるなら勝たなければなりません。** フォローできるならフォローし、その中に
// 場を上回れる札があるならそれしか出せません。味方が勝っているときだけ自由です。
func (g *Botifarra) GetValidPlayIndices(idx int) []int {
	if g.phase != BotifarraPhasePlay || idx < 0 || idx >= len(g.players) {
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

	leadSuit := g.trick[0].Card.GetDesign()
	winner := ResolveTrickWinner(g.trick, g.trumpSuit, BotifarraRank)
	partnerWinning := BotifarraTeamOf(winner) == BotifarraTeamOf(idx)
	best := g.bestRankInTrick()

	if follow := filterByDesign(p, all, leadSuit); len(follow) > 0 {
		// **リード自体が切り札のときも義務は効きます。** 除外していいのは
		// 「リードとは *別の* スートの切り札で取られた」場合だけ——そのときは
		// フォローしても勝てないからです。切り札がリードされたなら、同じスートの
		// 上位札で普通に勝てるので、上回れるなら出さなければなりません。
		ruffed := g.trickWonByTrump() && g.trumpSuit != leadSuit
		if !partnerWinning && !ruffed {
			if above := filterAbove(p, follow, best, BotifarraRank); len(above) > 0 {
				return above
			}
		}
		return follow
	}

	// フォローできないとき。**味方が勝っていなければ切り札で取る義務があります。**
	trumps := g.trumpIndices(idx, all)
	if partnerWinning || len(trumps) == 0 {
		if partnerWinning {
			return all
		}
		return all
	}
	if g.trickWonByTrump() {
		if above := filterAbove(p, trumps, best, BotifarraRank); len(above) > 0 {
			return above
		}
		// 上から被せられないなら切り札を捨てる義務は無く、自由に出せます。
		return all
	}
	return trumps
}

// trumpIndices は手札のうち切り札の位置を返す。
func (g *Botifarra) trumpIndices(idx int, all []int) []int {
	if g.trumpSuit == BotifarraNoTrump {
		return nil
	}
	return filterByDesign(g.players[idx], all, g.trumpSuit)
}

// trickWonByTrump は場が切り札で取られているかを返す。
func (g *Botifarra) trickWonByTrump() bool {
	if g.trumpSuit == BotifarraNoTrump {
		return false
	}
	for _, tc := range g.trick {
		if tc.Card.GetDesign() == g.trumpSuit {
			return true
		}
	}
	return false
}

// bestRankInTrick は場でいちばん強い札の強さを返す。
//
// 切り札が出ていれば切り札の中の最強、出ていなければリードスートの中の最強です。
func (g *Botifarra) bestRankInTrick() int {
	suit := g.trumpSuit
	if !g.trickWonByTrump() {
		if len(g.trick) == 0 {
			return 0
		}
		suit = g.trick[0].Card.GetDesign()
	}
	best := 0
	for _, tc := range g.trick {
		if tc.Card.GetDesign() != suit {
			continue
		}
		if r := BotifarraRank(tc.Card); r > best {
			best = r
		}
	}
	return best
}

// PlayCard は人間が札を出す。
func (g *Botifarra) PlayCard(cardIndex int) error {
	if g.phase != BotifarraPhasePlay {
		return errors.New("いまはプレイの場面ではありません")
	}
	if !g.players[g.currentTurn].GetIsHuman() {
		return errors.New("あなたの番ではありません")
	}
	valid := g.GetValidPlayIndices(g.currentTurn)
	if !slices.Contains(valid, cardIndex) {
		return errors.New("その札は出せません（勝てるなら勝つ義務があります）")
	}
	g.playAt(g.currentTurn, cardIndex)
	g.advanceCpu()
	return nil
}

// playAt は席 idx に cardIndex を出させる。
func (g *Botifarra) playAt(idx, cardIndex int) {
	card := g.players[idx].RemoveCard(cardIndex)
	if card == nil {
		// **手札が尽きた席に手番は回りません。** 万一回ったらループを止めます
		// (#4606 と同じ、進まない CpuPlay で回り続けるのを防ぐため)。
		return
	}
	g.trick = append(g.trick, &TrickCard{PlayerIdx: idx, Card: card})
	g.addLog(idx, "card", "札を出しました", []*Card{card})

	if len(g.trick) < BotifarraPlayerCnt {
		g.currentTurn = (idx + 1) % BotifarraPlayerCnt
		return
	}
	g.finishTrick()
}

// finishTrick はトリックを決着させる。
func (g *Botifarra) finishTrick() {
	winner := ResolveTrickWinner(g.trick, g.trumpSuit, BotifarraRank)
	team := BotifarraTeamOf(winner)

	pts := 1 // **各トリックに 1 点。**
	won := make([]*Card, 0, len(g.trick))
	for _, tc := range g.trick {
		pts += BotifarraCardPoint(tc.Card)
		won = append(won, tc.Card)
	}
	g.players[winner].AddTrick(won)
	g.roundPoints[team] += pts

	g.lastTrick = g.trick
	g.lastTrickWinner = winner
	g.trick = nil
	g.trickCount++
	g.currentTurn = winner
	g.addLog(winner, "trick", fmt.Sprintf("トリックを取りました（%d 点）", pts), nil)

	if g.trickCount >= BotifarraTrickCnt {
		g.finishRound()
	}
}

// finishRound はラウンドを精算する。
//
// **72 点のうち 36 を超えたぶんだけが得点。** 倍付けは**この配りそのものの価値**に
// 掛かるので、契約した側か守った側かに関係なく、36 を超えた側の得点に乗ります。
func (g *Botifarra) finishRound() {
	for team := range 2 {
		diff := g.roundPoints[team] - BotifarraHalfPoints
		if diff <= 0 {
			continue
		}
		diff *= g.multiplier
		g.scores[team] += diff
		g.addLog(-1, "score", fmt.Sprintf("チーム %d が %d 点を獲得しました", team, diff), nil)
	}

	g.phase = BotifarraPhaseRoundEnd
	for team := range 2 {
		if g.scores[team] >= g.config.TargetScore {
			g.finishGame()
			return
		}
	}
}

// finishGame は終局する。
func (g *Botifarra) finishGame() {
	g.phase = BotifarraPhaseGameEnd
	g.gameEndFlag = true
	g.winnerTeam = 0
	if g.scores[1] > g.scores[0] {
		g.winnerTeam = 1
	}
	g.addLog(-1, "result", fmt.Sprintf("チーム %d の勝ちです", g.winnerTeam), nil)
}

// NextRound は次のラウンドを配る。
func (g *Botifarra) NextRound() error {
	if g.gameEndFlag {
		return errors.New("ゲームは終了しています")
	}
	if g.phase != BotifarraPhaseRoundEnd {
		return errors.New("いまはラウンドの区切りではありません")
	}
	g.dealerIdx = (g.dealerIdx + 1) % BotifarraPlayerCnt
	g.startRound()
	return nil
}

// GiveUp は投了する。
func (g *Botifarra) GiveUp() {
	if g.gameEndFlag {
		return
	}
	g.phase = BotifarraPhaseGameEnd
	g.gameEndFlag = true
	// 人間は席 0 なので、相手チームの勝ちにします。
	g.winnerTeam = 1
	g.addLog(0, "giveup", "投了しました", nil)
}

// CpuPlay は CPU の手番を 1 つ進める。
func (g *Botifarra) CpuPlay() {
	g.advanceCpu()
}

// advanceCpu は人間の番になるまで CPU を進める。
func (g *Botifarra) advanceCpu() {
	for range BotifarraPlayerCnt * BotifarraTrickCnt * 4 {
		if g.gameEndFlag || g.phase == BotifarraPhaseRoundEnd {
			return
		}
		if g.currentTurn < 0 || g.currentTurn >= len(g.players) {
			return
		}
		if g.players[g.currentTurn].GetIsHuman() {
			return
		}
		switch g.phase {
		case BotifarraPhaseDeclare, BotifarraPhaseDelegated:
			g.cpuDeclare()
		case BotifarraPhaseDouble:
			g.cpuDouble()
		case BotifarraPhasePlay:
			g.playAt(g.currentTurn, g.cpuChooseCard(g.currentTurn))
		default:
			return
		}
	}
}

// cpuDeclare は CPU の宣言。
//
// **いちばん長いスートを切り札にします。** 委ねるのは、どのスートも短くて
// 引き受けるだけ損なときだけです。
func (g *Botifarra) cpuDeclare() {
	idx := g.currentTurn
	best, bestLen := BotifarraNoTrump, 0
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		n := 0
		p := g.players[idx]
		for i := range p.GetCardsSize() {
			if p.GetCard(i).GetDesign() == suit {
				n++
			}
		}
		if n > bestLen {
			best, bestLen = suit, n
		}
	}
	// 親の番で、どのスートも 4 枚に満たないなら相方に委ねます。
	if g.phase == BotifarraPhaseDeclare && bestLen < 4 {
		g.delegate(idx)
		return
	}
	if bestLen < 3 {
		best = BotifarraNoTrump
	}
	_ = g.declare(idx, best)
}

// cpuDouble は CPU の倍付け判断。**点札を多く持っているときだけ倍にします。**
func (g *Botifarra) cpuDouble() {
	idx := g.currentTurn
	p := g.players[idx]
	pts := 0
	for i := range p.GetCardsSize() {
		pts += BotifarraCardPoint(p.GetCard(i))
	}
	if pts >= 12 && g.multiplier < BotifarraMultiplierRecontra {
		_ = g.double(idx)
		return
	}
	g.passDouble(idx)
}

// cpuChooseCard は CPU が出す札を選ぶ。
func (g *Botifarra) cpuChooseCard(idx int) int {
	valid := g.GetValidPlayIndices(idx)
	if len(valid) == 0 {
		return 0
	}
	p := g.players[idx]
	if len(g.trick) == 0 {
		// リードは手札のいちばん強い札から。
		return pickHighest(p, valid, BotifarraRank)
	}
	winner := ResolveTrickWinner(g.trick, g.trumpSuit, BotifarraRank)
	if BotifarraTeamOf(winner) == BotifarraTeamOf(idx) {
		// 味方が勝っているなら点札を積んで、いちばん安い札は残します。
		return pickHighest(p, valid, BotifarraCardPoint)
	}
	return pickLowest(p, valid, BotifarraRank)
}

// botifarraTrumpName は切り札の表示名を返す。
func botifarraTrumpName(suit int) string {
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
func (g *Botifarra) GetHint() *BotifarraHint {
	if g.gameEndFlag || !g.players[0].GetIsHuman() {
		return nil
	}
	switch g.phase {
	case BotifarraPhaseDeclare, BotifarraPhaseDelegated:
		if g.currentTurn != 0 {
			return nil
		}
		suit := g.longestSuitOf(0)
		return &BotifarraHint{Suit: &suit, Reason: "botifarraDeclareLongest"}
	case BotifarraPhasePlay:
		if g.currentTurn != 0 {
			return nil
		}
		idx := g.cpuChooseCard(0)
		return &BotifarraHint{CardIndex: &idx, Reason: "botifarraMustWin"}
	default:
		return nil
	}
}

// longestSuitOf は席 idx のいちばん長いスートを返す。
func (g *Botifarra) longestSuitOf(idx int) int {
	best, bestLen := CardDesignSpade, -1
	p := g.players[idx]
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		n := 0
		for i := range p.GetCardsSize() {
			if p.GetCard(i).GetDesign() == suit {
				n++
			}
		}
		if n > bestLen {
			best, bestLen = suit, n
		}
	}
	return best
}

// addLog は棋譜に 1 行足す。
func (g *Botifarra) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.appendLog(playerIdx, actionType, detail, cards)
}

// --- Getters ---

// GetConfig は設定を返す。
func (g *Botifarra) GetConfig() BotifarraConfig { return g.config }

// SetConfig は設定を更新する。
func (g *Botifarra) SetConfig(cfg BotifarraConfig) {
	if cfg.Validate() != nil {
		return
	}
	g.config = cfg
}

// GetPhase は現在のフェーズを返す。
func (g *Botifarra) GetPhase() int { return g.phase }

// GetGameEndFlag は終局フラグを返す。
func (g *Botifarra) GetGameEndFlag() bool { return g.gameEndFlag }

// GetDealerIdx は親の席を返す。
func (g *Botifarra) GetDealerIdx() int { return g.dealerIdx }

// GetDeclarerIdx は切り札を決めた席を返す (-1 = 未決定)。
func (g *Botifarra) GetDeclarerIdx() int { return g.declarerIdx }

// GetTrumpSuit は切り札を返す (BotifarraNoTrump = 切り札なし)。
func (g *Botifarra) GetTrumpSuit() int { return g.trumpSuit }

// GetMultiplier は倍付けの倍率を返す。
func (g *Botifarra) GetMultiplier() int { return g.multiplier }

// GetCurrentTurn はいま出す番の席を返す。
func (g *Botifarra) GetCurrentTurn() int { return g.currentTurn }

// IsHumanTurn は人間の入力を待っているかを返す。
func (g *Botifarra) IsHumanTurn() bool {
	if g.gameEndFlag || g.phase == BotifarraPhaseRoundEnd {
		return false
	}
	return g.currentTurn == 0 && g.players[0].GetIsHuman()
}

// GetTrick はいま進行中のトリックを返す。
func (g *Botifarra) GetTrick() []*TrickCard { return g.trick }

// GetLastTrick は直前に完成したトリックを返す。
func (g *Botifarra) GetLastTrick() []*TrickCard { return g.lastTrick }

// GetLastTrickWinner は直前のトリックを取った席を返す (-1 = まだ無い)。
func (g *Botifarra) GetLastTrickWinner() int { return g.lastTrickWinner }

// GetTrickCount はこのラウンドで完成したトリック数を返す。
func (g *Botifarra) GetTrickCount() int { return g.trickCount }

// GetRoundPoints はチーム team がこのラウンドで取った点を返す。
func (g *Botifarra) GetRoundPoints(team int) int {
	if team < 0 || team > 1 {
		return 0
	}
	return g.roundPoints[team]
}

// GetScore はチーム team の通算得点を返す。
func (g *Botifarra) GetScore(team int) int {
	if team < 0 || team > 1 {
		return 0
	}
	return g.scores[team]
}

// GetWinnerTeam は勝ったチームを返す (-1 = 未確定)。
func (g *Botifarra) GetWinnerTeam() int { return g.winnerTeam }

// GetPlayerCnt は人数を返す。
func (g *Botifarra) GetPlayerCnt() int { return len(g.players) }

// GetPlayer は席 i のプレイヤーを返す。
func (g *Botifarra) GetPlayer(i int) *BotifarraPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// botifarraJSON is the JSON wire format for Botifarra.
type botifarraJSON struct {
	TrumpCards      *TrumpCards        `json:"tc"`
	Players         []*BotifarraPlayer `json:"pl"`
	Config          BotifarraConfig    `json:"cf"`
	Phase           int                `json:"ph"`
	DealerIdx       int                `json:"di"`
	DeclarerIdx     int                `json:"dc"`
	TrumpSuit       int                `json:"ts"`
	Multiplier      int                `json:"ml"`
	CurrentTurn     int                `json:"ct"`
	Trick           []*TrickCard       `json:"tk"`
	LastTrick       []*TrickCard       `json:"lt"`
	LastTrickWinner int                `json:"lw"`
	TrickCount      int                `json:"tn"`
	RoundPoints     [2]int             `json:"rp"`
	Scores          [2]int             `json:"sc"`
	GameEndFlag     bool               `json:"ge"`
	WinnerTeam      int                `json:"wt"`
	ActionLog       []*ActionLogEntry  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Botifarra) MarshalJSON() ([]byte, error) {
	return json.Marshal(botifarraJSON{
		TrumpCards:      g.trumpCards,
		Players:         g.players,
		Config:          g.config,
		Phase:           g.phase,
		DealerIdx:       g.dealerIdx,
		DeclarerIdx:     g.declarerIdx,
		TrumpSuit:       g.trumpSuit,
		Multiplier:      g.multiplier,
		CurrentTurn:     g.currentTurn,
		Trick:           g.trick,
		LastTrick:       g.lastTrick,
		LastTrickWinner: g.lastTrickWinner,
		TrickCount:      g.trickCount,
		RoundPoints:     g.roundPoints,
		Scores:          g.scores,
		GameEndFlag:     g.gameEndFlag,
		WinnerTeam:      g.winnerTeam,
		ActionLog:       g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **範囲チェックだけでは足りません。** 「トリックに 5 枚ある」「切り札なのに倍率が
// 付いていない段階」は、どの値も単独では範囲内なのに盤面としてはあり得ません。
// ここを通すと勝敗だけが静かに変わります。
func (g *Botifarra) UnmarshalJSON(data []byte) error {
	var j botifarraJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := botifarraValidateWire(&j); err != nil {
		return err
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCardsReversis()
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.dealerIdx = j.DealerIdx
	g.declarerIdx = j.DeclarerIdx
	g.trumpSuit = j.TrumpSuit
	g.multiplier = j.Multiplier
	g.currentTurn = j.CurrentTurn
	g.trick = j.Trick
	g.lastTrick = j.LastTrick
	g.lastTrickWinner = j.LastTrickWinner
	g.trickCount = j.TrickCount
	g.roundPoints = j.RoundPoints
	g.scores = j.Scores
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	if g.rng == nil {
		g.rng = rand.New(rand.NewSource(rand.Int63()))
	}
	return nil
}

// botifarraValidateWire は復元しようとしている盤面の不変条件を検証する。
func botifarraValidateWire(j *botifarraJSON) error {
	if err := j.Config.Validate(); err != nil {
		return err
	}
	if len(j.Players) != BotifarraPlayerCnt {
		return fmt.Errorf("botifarra: seat count %d does not match the fixed %d",
			len(j.Players), BotifarraPlayerCnt)
	}
	for i, p := range j.Players {
		if p == nil {
			return fmt.Errorf("botifarra: seat %d is missing", i)
		}
	}
	if j.Phase < BotifarraPhaseDeclare || j.Phase > BotifarraPhaseGameEnd {
		return fmt.Errorf("botifarra: phase out of range: %d", j.Phase)
	}
	if j.GameEndFlag != (j.Phase == BotifarraPhaseGameEnd) {
		return fmt.Errorf("botifarra: the game-end flag and the phase disagree (flag=%v, phase=%d)",
			j.GameEndFlag, j.Phase)
	}
	if len(j.ActionLog) > botifarraMaxSliceLen {
		return fmt.Errorf("botifarra: action log too long: %d", len(j.ActionLog))
	}
	for name, seat := range map[string]int{"dealer": j.DealerIdx, "current turn": j.CurrentTurn} {
		if seat < 0 || seat >= BotifarraPlayerCnt {
			return fmt.Errorf("botifarra: %s index out of range: %d", name, seat)
		}
	}
	if j.DeclarerIdx < -1 || j.DeclarerIdx >= BotifarraPlayerCnt {
		return fmt.Errorf("botifarra: declarer index out of range: %d", j.DeclarerIdx)
	}
	if j.LastTrickWinner < -1 || j.LastTrickWinner >= BotifarraPlayerCnt {
		return fmt.Errorf("botifarra: last trick winner out of range: %d", j.LastTrickWinner)
	}
	if j.WinnerTeam < -1 || j.WinnerTeam > 1 {
		return fmt.Errorf("botifarra: winner team out of range: %d", j.WinnerTeam)
	}
	if j.GameEndFlag != (j.WinnerTeam >= 0) {
		return fmt.Errorf("botifarra: a finished game has a winning team and an unfinished one does not (flag=%v, team=%d)",
			j.GameEndFlag, j.WinnerTeam)
	}
	if j.TrumpSuit != BotifarraNoTrump &&
		(j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return fmt.Errorf("botifarra: trump suit out of range: %d", j.TrumpSuit)
	}
	switch j.Multiplier {
	case BotifarraMultiplierNone, BotifarraMultiplierContrar, BotifarraMultiplierRecontra:
	default:
		return fmt.Errorf("botifarra: multiplier out of range: %d", j.Multiplier)
	}
	// **切り札を決める前に倍率は上がりません。**
	if j.DeclarerIdx < 0 && j.Multiplier != BotifarraMultiplierNone {
		return fmt.Errorf("botifarra: the stake is doubled before anyone declared (multiplier=%d)", j.Multiplier)
	}
	return botifarraValidateTricks(j)
}

// botifarraValidateTricks はトリックと点の整合を検証する。
func botifarraValidateTricks(j *botifarraJSON) error {
	if j.TrickCount < 0 || j.TrickCount > BotifarraTrickCnt {
		return fmt.Errorf("botifarra: trick count out of range: %d", j.TrickCount)
	}
	// **1 トリックは 4 枚まで。** 5 枚目が積まれることはありません。
	if len(j.Trick) >= BotifarraPlayerCnt {
		return fmt.Errorf("botifarra: the current trick holds %d cards, it resolves at %d",
			len(j.Trick), BotifarraPlayerCnt)
	}
	if len(j.LastTrick) != 0 && len(j.LastTrick) != BotifarraPlayerCnt {
		return fmt.Errorf("botifarra: the last trick holds %d cards, want %d",
			len(j.LastTrick), BotifarraPlayerCnt)
	}
	for _, tc := range j.Trick {
		if tc == nil || tc.Card == nil {
			return errors.New("botifarra: the current trick holds an empty slot")
		}
		if tc.PlayerIdx < 0 || tc.PlayerIdx >= BotifarraPlayerCnt {
			return fmt.Errorf("botifarra: a trick card names seat %d", tc.PlayerIdx)
		}
	}
	for team := range 2 {
		if j.RoundPoints[team] < 0 || j.RoundPoints[team] > BotifarraTotalPoints {
			return fmt.Errorf("botifarra: round points for team %d out of range: %d",
				team, j.RoundPoints[team])
		}
		if j.Scores[team] < 0 {
			return fmt.Errorf("botifarra: score for team %d cannot be negative: %d", team, j.Scores[team])
		}
	}
	// **配り切ったラウンドの点は 72 ちょうど。** 途中なら 72 を超えません。
	sum := j.RoundPoints[0] + j.RoundPoints[1]
	if j.TrickCount == BotifarraTrickCnt && sum != BotifarraTotalPoints {
		return fmt.Errorf("botifarra: a finished round holds %d points, want %d", sum, BotifarraTotalPoints)
	}
	if sum > BotifarraTotalPoints {
		return fmt.Errorf("botifarra: %d points are in play, at most %d exist", sum, BotifarraTotalPoints)
	}
	return nil
}
