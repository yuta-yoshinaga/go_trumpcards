//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"
)

// KarnoffelPlayerCnt はカルニッフェルの人数。
const KarnoffelPlayerCnt = 4

// KarnoffelTeamCnt はチーム数。
const KarnoffelTeamCnt = 2

// KarnoffelHandSize は 1 人あたりの手札枚数。
//
// **5 枚であって 12 枚ではない。**48 ÷ 4 = 12 だと配り切ってしまい、切札を
// 決める表向きの札が残らない。5 トリック制であることとも整合する。
const KarnoffelHandSize = 5

// KarnoffelDeckSize は使用する札数 (**A を除いた 48 枚**)。
//
// ドイツ式の König / Ober / Unter / Banner はフランス式の K / Q / J / 10 に
// 対応する。**A が無いことが平スートの序列の前提**なので、「8 を除く」では
// 序列そのものが成り立たない。
const KarnoffelDeckSize = 48

// KarnoffelTricks は 1 局のトリック数。
const KarnoffelTricks = KarnoffelHandSize

// KarnoffelTricksToWin は 1 局を取るのに要るトリック数。
const KarnoffelTricksToWin = 3

// KarnoffelDefaultTarget は既定の勝利点 (取った局数)。
const KarnoffelDefaultTarget = 3

// 選ばれたスート内の役職札。
const (
	// KarnoffelKarnoffel カルニッフェル (**J**)。全札に勝つ。
	KarnoffelKarnoffel = 11
	// KarnoffelDevil 悪魔 (7)。**リードされたときだけ**強い。
	KarnoffelDevil = 7
	// KarnoffelPope 法王 (6)
	KarnoffelPope = 6
	// KarnoffelKaiser 皇帝 (2)
	KarnoffelKaiser = 2
	// KarnoffelOberstecher オーバーシュテッヒャー (3)。**K に負ける。**
	KarnoffelOberstecher = 3
	// KarnoffelUnterstecher ウンターシュテッヒャー (4)。**K・Q に負ける。**
	KarnoffelUnterstecher = 4
	// KarnoffelFarbenstecher ファルベンシュテッヒャー (5)。**絵札に負ける。**
	KarnoffelFarbenstecher = 5
)

// karnoffelPlainOrder は平スートの序列 (強い順)。
//
// **A は入っていない。**K > Q > J > 10 > 9 > 8 > 7 > 6 > 5 > 4 > 3 > 2。
var karnoffelPlainOrder = []int{13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2}

// karnoffelPlainRank は平スートの強さを返す (大きいほど強い)。
func karnoffelPlainRank(c *Card) int {
	if c == nil {
		return 0
	}
	for i, v := range karnoffelPlainOrder {
		if c.GetValue() == v {
			return len(karnoffelPlainOrder) - i
		}
	}
	return 0
}

// KarnoffelIsFaceCard は絵札 (K・Q・J) かを返す。
//
// **ファルベンシュテッヒャー (5) はこれに負ける。**
func KarnoffelIsFaceCard(c *Card) bool {
	if c == nil {
		return false
	}
	switch c.GetValue() {
	case 13, 12, 11:
		return true
	}
	return false
}

// karnoffelChosenRank は選ばれたスート内の役職序列を返す (大きいほど強い)。
//
// **悪魔 (7) はここに含めない。**リードされたかどうかで強さが変わるので、
// 単純な数値テーブルでは表せない。
var karnoffelChosenOrder = []int{
	KarnoffelKarnoffel,     // J
	KarnoffelPope,          // 6
	KarnoffelKaiser,        // 2
	KarnoffelOberstecher,   // 3
	KarnoffelUnterstecher,  // 4
	KarnoffelFarbenstecher, // 5
	13, 12, 10, 9, 8,       // 特権の無い切札
}

// karnoffelChosenRank は選ばれたスートの札の役職序列を返す。
func karnoffelChosenRank(c *Card) int {
	if c == nil {
		return 0
	}
	return karnoffelChosenRankOf(c.GetValue())
}

// karnoffelPartialBeats は部分切札 (3/4/5) が相手札に勝てるかを返す。
//
// **3 は K に、4 は K・Q に、5 は絵札すべてに負ける。**選ばれたスートを出せば
// 平札が無力化される、という単純な話ではない。
func karnoffelPartialBeats(value int, other *Card) bool {
	if other == nil {
		return true
	}
	switch value {
	case KarnoffelOberstecher:
		return other.GetValue() != 13
	case KarnoffelUnterstecher:
		return other.GetValue() != 13 && other.GetValue() != 12
	case KarnoffelFarbenstecher:
		return !KarnoffelIsFaceCard(other)
	}
	return true
}

// karnoffelIsPartial は部分切札かを返す。
func karnoffelIsPartial(value int) bool {
	switch value {
	case KarnoffelOberstecher, KarnoffelUnterstecher, KarnoffelFarbenstecher:
		return true
	}
	return false
}

// KarnoffelBeats は c が best に勝つかを返す。
//
// cLed / bestLed はそれぞれの札がトリックの第 1 打かどうか。**悪魔 (7) は
// リードされたときだけ強い**ので、位置を見ないと比較できない。
func KarnoffelBeats(c, best, lead *Card, chosenSuit int, cLed, bestLed bool) bool {
	if c == nil || best == nil {
		return false
	}
	cScore, cPartial := karnoffelPower(c, chosenSuit, cLed)
	bScore, bPartial := karnoffelPower(best, chosenSuit, bestLed)

	// 部分切札どうしでなければ、部分切札は「勝てない相手」に阻まれる。
	if cPartial && !karnoffelPartialBeats(c.GetValue(), best) {
		return false
	}
	if bPartial && !karnoffelPartialBeats(best.GetValue(), c) {
		return true
	}
	if cScore != bScore {
		return cScore > bScore
	}
	// どちらも切札でも役職でもないなら、リードスートに追随した札だけが争う。
	if lead == nil {
		return false
	}
	cFollow := c.GetDesign() == lead.GetDesign()
	bFollow := best.GetDesign() == lead.GetDesign()
	if cFollow != bFollow {
		return cFollow
	}
	if !cFollow {
		return false
	}
	return karnoffelPlainRank(c) > karnoffelPlainRank(best)
}

// karnoffelPower は札の強さ (大きいほど強い) と、部分切札かどうかを返す。
//
// **役職の序列を 2 倍に伸ばしてある。**リードされた悪魔はカルニッフェルと法王の
// 「あいだ」に入るので、詰まった目盛りでは表せない。平札は 0。
func karnoffelPower(c *Card, chosenSuit int, led bool) (int, bool) {
	if c == nil {
		return 0, false
	}
	if chosenSuit == 0 || c.GetDesign() != chosenSuit {
		return 0, false
	}
	if c.GetValue() == KarnoffelDevil {
		// **リードされた悪魔はカルニッフェル以外に勝つ。**
		// 追随して出した悪魔は逆にあらゆる札に負ける。
		if led {
			return karnoffelChosenScore(KarnoffelKarnoffel) - 1, false
		}
		return -1, false
	}
	return karnoffelChosenScore(c.GetValue()), karnoffelIsPartial(c.GetValue())
}

// karnoffelChosenScore は選ばれたスートの札の強さを返す。
//
// 平札 (0) より必ず上に来るよう 1 を足し、悪魔を割り込ませる余地のために
// 2 倍に伸ばす。
func karnoffelChosenScore(value int) int {
	return 1 + 2*karnoffelChosenRankOf(value)
}

// karnoffelChosenRankOf は値から役職序列を引く。
func karnoffelChosenRankOf(value int) int {
	for i, v := range karnoffelChosenOrder {
		if value == v {
			return len(karnoffelChosenOrder) - i
		}
	}
	return 0
}

// KarnoffelTeamOf は席のチームを返す (範囲外は -1)。
//
// **パートナーは向かい合わせ。**Go の剰余は -1 % 2 = -1 なので素通しは危険。
func KarnoffelTeamOf(seat int) int {
	if seat < 0 || seat >= KarnoffelPlayerCnt {
		return -1
	}
	return seat % KarnoffelTeamCnt
}

// KarnoffelPhase はゲームフェーズ。
type KarnoffelPhase int

// カルニッフェルのフェーズ定数
const (
	// KarnoffelPhasePlay トリックプレイ
	KarnoffelPhasePlay KarnoffelPhase = iota
	// KarnoffelPhaseHandEnd 局終了 (精算済み)
	KarnoffelPhaseHandEnd
	// KarnoffelPhaseGameEnd ゲーム終了
	KarnoffelPhaseGameEnd
)

// KarnoffelHandResult は 1 局の結果。
type KarnoffelHandResult struct {
	// WinnerTeam はこの局を取ったチーム (未決着なら -1)。
	WinnerTeam int
	// Tricks は各チームが取ったトリック数。
	Tricks [KarnoffelTeamCnt]int
	// ChosenSuit はこの局の選ばれたスート。
	ChosenSuit int
}

// Karnoffel はカルニッフェルのゲームクラス。
//
// 1426 年の文献に名が残る現存最古の名前付きカードゲーム。**A を除いた 48 枚**
// を 4 人 2 対 2 で使い、**1 人 5 枚**を配る。最初の 1 枚だけ表向きに配られ、
// **表向きの 4 枚のうち最も低い札が「選ばれたスート」を決める。**
type Karnoffel struct {
	players []*KarnoffelPlayer
	config  KarnoffelConfig
	phase   KarnoffelPhase
	// dealerIdx は親。
	dealerIdx int
	// upCards は各席に表向きで配られた 1 枚。
	upCards []*Card
	// chosenSuit は選ばれたスート。
	chosenSuit int
	// currentIdx は現在の手番。
	currentIdx int
	trick      []*Card
	// trickLeader はこのトリックのリード席。
	trickLeader int
	trickNumber int
	// tricksWon は各席が取ったトリック数。
	tricksWon [KarnoffelPlayerCnt]int
	// handsWon は各チームが取った局数。
	handsWon    [KarnoffelTeamCnt]int
	lastResult  *KarnoffelHandResult
	handNumber  int
	gameEndFlag bool
	winnerTeam  int
	actionLogBase
}

// NewKarnoffel コンストラクタ
func NewKarnoffel(players []*KarnoffelPlayer, config KarnoffelConfig) *Karnoffel {
	return &Karnoffel{players: players, config: config, winnerTeam: -1}
}

// NewDefaultKarnoffel は人間 1 人 + CPU 3 体の卓を作る。
func NewDefaultKarnoffel() *Karnoffel {
	players := make([]*KarnoffelPlayer, 0, KarnoffelPlayerCnt)
	for i := range KarnoffelPlayerCnt {
		players = append(players, NewKarnoffelPlayer(i == 0))
	}
	return NewKarnoffel(players, DefaultKarnoffelConfig())
}

// Reset はゲームを初期化する。
func (k *Karnoffel) Reset() {
	for i := range KarnoffelTeamCnt {
		k.handsWon[i] = 0
	}
	k.handNumber = 0
	k.gameEndFlag = false
	k.winnerTeam = -1
	k.lastResult = nil
	k.dealerIdx = 0
	k.actionLog = make([]*ActionLogEntry, 0)
	k.beginHand()
}

// beginHand は 1 局を配る。
func (k *Karnoffel) beginHand() {
	k.handNumber++
	k.phase = KarnoffelPhasePlay
	k.trick = make([]*Card, 0, KarnoffelPlayerCnt)
	k.trickNumber = 0
	for i := range KarnoffelPlayerCnt {
		k.tricksWon[i] = 0
		if p := k.GetPlayer(i); p != nil {
			p.ResetRound()
		}
	}
	k.dealRound()
	// **リードは親の左から。**
	k.trickLeader = (k.dealerIdx + 1) % KarnoffelPlayerCnt
	k.currentIdx = k.trickLeader
	k.addLog(-1, "deal", "hand "+strconv.Itoa(k.handNumber)+" chosen "+strconv.Itoa(k.chosenSuit), nil)
}

// dealRound は 1 枚目を表向きに、残り 4 枚を伏せて配る。
//
// **表向きの 4 枚のうち最も低い札が切札スートを決める。**「最後の 1 枚をめくる」
// のではない。
func (k *Karnoffel) dealRound() {
	deck := newKarnoffelDeck()
	karnoffelShuffle(deck)
	pos := 0
	k.upCards = make([]*Card, KarnoffelPlayerCnt)

	for i := range KarnoffelPlayerCnt {
		seat := (k.dealerIdx + 1 + i) % KarnoffelPlayerCnt
		p := k.GetPlayer(seat)
		if pos < len(deck) && p != nil {
			c := deck[pos]
			pos++
			k.upCards[seat] = c
			p.AddCard(c)
		}
	}
	for range KarnoffelHandSize - 1 {
		for i := range KarnoffelPlayerCnt {
			seat := (k.dealerIdx + 1 + i) % KarnoffelPlayerCnt
			p := k.GetPlayer(seat)
			if pos < len(deck) && p != nil {
				p.AddCard(deck[pos])
				pos++
			}
		}
	}
	k.chosenSuit = karnoffelChooseSuit(k.upCards)
}

// karnoffelChooseSuit は表向きの札のうち最も低いものからスートを決める。
//
// 同順位が並んだときはスート番号の小さいほうを採る (決定的にするため)。
func karnoffelChooseSuit(up []*Card) int {
	best, bestRank := 0, 1<<30
	for _, c := range up {
		if c == nil {
			continue
		}
		r := karnoffelPlainRank(c)
		if r < bestRank || (r == bestRank && c.GetDesign() < best) {
			best, bestRank = c.GetDesign(), r
		}
	}
	return best
}

// newKarnoffelDeck は **A を除いた 48 枚**を作る。
func newKarnoffelDeck() []*Card {
	cards := make([]*Card, 0, KarnoffelDeckSize)
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		for _, v := range karnoffelPlainOrder {
			cards = append(cards, NewCard(suit, v, true))
		}
	}
	return cards
}

// karnoffelShuffle は札をシャッフルする。
func karnoffelShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1) //nolint:gosec // ゲームのシャッフルに暗号強度は要らない
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// KarnoffelValidPlays は出せる手札インデックスを返す。
//
// **追随の義務は無い。**ただし**第 1 トリックのリードに悪魔は使えない。**
func (k *Karnoffel) KarnoffelValidPlays(player int) []int {
	p := k.GetPlayer(player)
	if p == nil {
		return nil
	}
	all := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		all = append(all, i)
	}
	leadingFirst := k.trickNumber == 0 && len(k.trick) == 0
	if !leadingFirst {
		return all
	}
	out := make([]int, 0, len(all))
	for _, i := range all {
		if k.isDevil(p.GetCard(i)) {
			continue
		}
		out = append(out, i)
	}
	// 悪魔しか無いなら出さざるを得ない。
	if len(out) == 0 {
		return all
	}
	return out
}

// isDevil は選ばれたスートの 7 かを返す。
func (k *Karnoffel) isDevil(c *Card) bool {
	return c != nil && k.chosenSuit != 0 && c.GetDesign() == k.chosenSuit && c.GetValue() == KarnoffelDevil
}

// PlayCard は手札を 1 枚出す。
func (k *Karnoffel) PlayCard(player, idx int) error {
	if k.gameEndFlag {
		return errors.New("the game is over")
	}
	if k.phase != KarnoffelPhasePlay {
		return errors.New("it is not the play phase")
	}
	if player != k.currentIdx {
		return errors.New("it is not your turn")
	}
	p := k.GetPlayer(player)
	if p == nil || idx < 0 || idx >= p.GetCardsSize() {
		return errors.New("there is no such card")
	}
	if !karnoffelContains(k.KarnoffelValidPlays(player), idx) {
		return errors.New("the devil cannot lead the first trick")
	}
	c := p.GetCard(idx)
	p.RemoveCard(idx)
	k.trick = append(k.trick, c)
	k.addLog(player, "play", "", []*Card{c})

	if len(k.trick) < KarnoffelPlayerCnt {
		k.currentIdx = (k.currentIdx + 1) % KarnoffelPlayerCnt
		return nil
	}
	k.resolveTrick()
	return nil
}

// karnoffelLeadingCard は場の札のうち現在勝っているものを返す。
//
// **位置が強さを変える**ので、単に強さで畳み込むことができない。リードされた
// 悪魔だけが特権を持つため、勝者が第 1 打かどうかを持ち回る必要がある。
// 途中経過の判定 (CPU の読み) と最終判定 (トリック確定) で同じ規則を使うので、
// 片方だけ直して食い違うことがないよう 1 箇所にまとめてある。
func karnoffelLeadingCard(trick []*Card, chosenSuit int) (int, *Card) {
	if len(trick) == 0 {
		return 0, nil
	}
	lead := trick[0]
	winOffset, best := 0, trick[0]
	for i := 1; i < len(trick); i++ {
		if KarnoffelBeats(trick[i], best, lead, chosenSuit, false, winOffset == 0) {
			winOffset, best = i, trick[i]
		}
	}
	return winOffset, best
}

// resolveTrick はトリックの勝者を決める。
func (k *Karnoffel) resolveTrick() {
	winOffset, _ := karnoffelLeadingCard(k.trick, k.chosenSuit)
	winner := (k.trickLeader + winOffset) % KarnoffelPlayerCnt
	k.tricksWon[winner]++
	k.addLog(winner, "trickWin", "", k.trick)

	k.trick = make([]*Card, 0, KarnoffelPlayerCnt)
	k.trickNumber++
	k.trickLeader = winner
	k.currentIdx = winner

	// **3 トリック取った時点で局は決まる。**5 トリック全部を打つ必要はない。
	if k.KarnoffelTeamTricks(0) >= KarnoffelTricksToWin ||
		k.KarnoffelTeamTricks(1) >= KarnoffelTricksToWin ||
		k.trickNumber >= KarnoffelTricks {
		k.finishHand()
	}
}

// karnoffelContains は s に v が含まれるかを返す。
func karnoffelContains(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// KarnoffelTeamTricks はチームが取ったトリック数を返す。
func (k *Karnoffel) KarnoffelTeamTricks(team int) int {
	if team < 0 || team >= KarnoffelTeamCnt {
		return 0
	}
	total := 0
	for i := range KarnoffelPlayerCnt {
		if KarnoffelTeamOf(i) == team {
			total += k.tricksWon[i]
		}
	}
	return total
}

// finishHand は 1 局を精算する。
func (k *Karnoffel) finishHand() {
	var tricks [KarnoffelTeamCnt]int
	for team := range KarnoffelTeamCnt {
		tricks[team] = k.KarnoffelTeamTricks(team)
	}
	winner := -1
	switch {
	case tricks[0] >= KarnoffelTricksToWin:
		winner = 0
	case tricks[1] >= KarnoffelTricksToWin:
		winner = 1
	}
	if winner >= 0 {
		k.handsWon[winner]++
	}
	k.lastResult = &KarnoffelHandResult{WinnerTeam: winner, Tricks: tricks, ChosenSuit: k.chosenSuit}
	k.phase = KarnoffelPhaseHandEnd
	k.addLog(-1, "handEnd", strconv.Itoa(tricks[0])+"-"+strconv.Itoa(tricks[1]), nil)
	k.checkGameEnd()
}

// checkGameEnd は規定局数を取ったチームがいるかを見る。
func (k *Karnoffel) checkGameEnd() {
	for team := range KarnoffelTeamCnt {
		if k.handsWon[team] >= k.config.TargetHands {
			k.gameEndFlag = true
			k.phase = KarnoffelPhaseGameEnd
			k.winnerTeam = team
			return
		}
	}
}

// NextHand は次の局を配る。
func (k *Karnoffel) NextHand() error {
	if k.gameEndFlag {
		return errors.New("the game is over")
	}
	if k.phase != KarnoffelPhaseHandEnd {
		return errors.New("the hand is still in progress")
	}
	k.dealerIdx = (k.dealerIdx + 1) % KarnoffelPlayerCnt
	k.beginHand()
	return nil
}

// ---- CPU ----

// KarnoffelCpuPlay は CPU が出す手札インデックスを返す (無ければ -1)。
func (k *Karnoffel) KarnoffelCpuPlay(idx int) int {
	p := k.GetPlayer(idx)
	if p == nil {
		return -1
	}
	valid := k.KarnoffelValidPlays(idx)
	if len(valid) == 0 {
		return -1
	}
	if len(k.trick) == 0 {
		// **悪魔はリードしてこそ強い。**第 1 トリック以外なら真っ先に出す。
		for _, i := range valid {
			if k.isDevil(p.GetCard(i)) {
				return i
			}
		}
		return karnoffelStrongest(p, valid, k.chosenSuit)
	}
	lead := k.trick[0]
	winOffset, best := karnoffelLeadingCard(k.trick, k.chosenSuit)
	bestIsLead := winOffset == 0
	// 勝てる中でいちばん弱い札を出す。
	win, winRank := -1, 1<<30
	for _, i := range valid {
		c := p.GetCard(i)
		if !KarnoffelBeats(c, best, lead, k.chosenSuit, false, bestIsLead) {
			continue
		}
		if r := karnoffelPlainRank(c); r < winRank {
			win, winRank = i, r
		}
	}
	if win >= 0 {
		return win
	}
	return karnoffelWeakest(p, valid)
}

// karnoffelStrongest は valid のうちいちばん強い札を返す。
func karnoffelStrongest(p *KarnoffelPlayer, valid []int, chosenSuit int) int {
	best, bestScore := valid[0], -1
	for _, i := range valid {
		c := p.GetCard(i)
		score, _ := karnoffelPower(c, chosenSuit, true)
		score = score*100 + karnoffelPlainRank(c)
		if score > bestScore {
			best, bestScore = i, score
		}
	}
	return best
}

// karnoffelWeakest は valid のうちいちばん弱い札を返す。
func karnoffelWeakest(p *KarnoffelPlayer, valid []int) int {
	best, bestRank := valid[0], 1<<30
	for _, i := range valid {
		if r := karnoffelPlainRank(p.GetCard(i)); r < bestRank {
			best, bestRank = i, r
		}
	}
	return best
}

// IsHumanTurn は現在の手番が人間かを返す。
func (k *Karnoffel) IsHumanTurn() bool {
	if k.gameEndFlag || k.phase != KarnoffelPhasePlay {
		return false
	}
	p := k.GetPlayer(k.currentIdx)
	return p != nil && p.GetIsHuman()
}

// CpuPlay は CPU が 1 アクション実行する。
func (k *Karnoffel) CpuPlay() {
	if k.gameEndFlag || k.phase != KarnoffelPhasePlay || k.IsHumanTurn() {
		return
	}
	idx := k.currentIdx
	if i := k.KarnoffelCpuPlay(idx); i >= 0 {
		_ = k.PlayCard(idx, i)
	}
}

// ---- アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (k *Karnoffel) GetPlayers() []*KarnoffelPlayer { return k.players }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (k *Karnoffel) GetPlayer(idx int) *KarnoffelPlayer {
	return getPlayer(k.players, idx)
}

// GetPhase は現在のフェーズを返す。
func (k *Karnoffel) GetPhase() KarnoffelPhase { return k.phase }

// GetCurrentPlayerIdx は現在の手番を返す。
func (k *Karnoffel) GetCurrentPlayerIdx() int { return k.currentIdx }

// GetDealerIdx は親を返す。
func (k *Karnoffel) GetDealerIdx() int { return k.dealerIdx }

// GetChosenSuit は選ばれたスートを返す。
func (k *Karnoffel) GetChosenSuit() int { return k.chosenSuit }

// GetUpCard は席に表向きで配られた札を返す。
//
// **切札はこの 4 枚のうち最も低い札が決める。**どれが決めたのかが見えないと
// 盤面が読めない。
func (k *Karnoffel) GetUpCard(idx int) *Card {
	if idx < 0 || idx >= len(k.upCards) {
		return nil
	}
	return k.upCards[idx]
}

// GetTrick は場に出ている札を返す。
func (k *Karnoffel) GetTrick() []*Card { return k.trick }

// GetTrickLeaderIdx はこのトリックのリード席を返す。
func (k *Karnoffel) GetTrickLeaderIdx() int { return k.trickLeader }

// GetTrickNumber は済んだトリック数を返す。
func (k *Karnoffel) GetTrickNumber() int { return k.trickNumber }

// GetTricksWon は席が取ったトリック数を返す。
func (k *Karnoffel) GetTricksWon(idx int) int {
	if idx < 0 || idx >= KarnoffelPlayerCnt {
		return 0
	}
	return k.tricksWon[idx]
}

// GetHandsWon はチームが取った局数を返す。
func (k *Karnoffel) GetHandsWon(team int) int {
	if team < 0 || team >= KarnoffelTeamCnt {
		return 0
	}
	return k.handsWon[team]
}

// GetLastResult は直前の局の結果を返す (まだ無ければ nil)。
func (k *Karnoffel) GetLastResult() *KarnoffelHandResult { return k.lastResult }

// GetHandNumber は現在の局番号を返す。
func (k *Karnoffel) GetHandNumber() int { return k.handNumber }

// GetGameEndFlag はゲーム終了フラグを返す。
func (k *Karnoffel) GetGameEndFlag() bool { return k.gameEndFlag }

// GetWinnerTeam は勝利チームを返す (未確定なら -1)。
func (k *Karnoffel) GetWinnerTeam() int { return k.winnerTeam }

// GetConfig はゲーム設定を返す。
func (k *Karnoffel) GetConfig() KarnoffelConfig { return k.config }

// SetConfig はゲーム設定をセットする。
func (k *Karnoffel) SetConfig(c KarnoffelConfig) { k.config = c }

// GetActionLog は棋譜を返す。
func (k *Karnoffel) GetActionLog() []*ActionLogEntry { return k.actionLog }

// addLog は棋譜を 1 件追加する。
func (k *Karnoffel) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	k.appendLogAt(0, playerIdx, actionType, detail, cards)
}

// ---- テスト用 ----

// SetPhaseForTest はフェーズを差し替える (テスト専用)。
func (k *Karnoffel) SetPhaseForTest(p KarnoffelPhase) { k.phase = p }

// SetHandForTest は手札を差し替える (テスト専用)。
func (k *Karnoffel) SetHandForTest(idx int, cards []*Card) {
	setHandForTest(k.GetPlayer(idx), cards)
}

// SetChosenSuitForTest は選ばれたスートを差し替える (テスト専用)。
func (k *Karnoffel) SetChosenSuitForTest(suit int) { k.chosenSuit = suit }

// SetCurrentPlayerForTest は手番を差し替える (テスト専用)。
func (k *Karnoffel) SetCurrentPlayerForTest(idx int) { k.currentIdx = idx }

// SetDealerForTest は親を差し替える (テスト専用)。
func (k *Karnoffel) SetDealerForTest(idx int) { k.dealerIdx = idx }

// SetTrickLeaderForTest はリード席を差し替える (テスト専用)。
func (k *Karnoffel) SetTrickLeaderForTest(idx int) { k.trickLeader = idx }

// SetTrickNumberForTest は済んだトリック数を差し替える (テスト専用)。
func (k *Karnoffel) SetTrickNumberForTest(n int) { k.trickNumber = n }

// SetTricksWonForTest は取得トリック数を差し替える (テスト専用)。
func (k *Karnoffel) SetTricksWonForTest(idx, n int) {
	if idx >= 0 && idx < KarnoffelPlayerCnt {
		k.tricksWon[idx] = n
	}
}

// SetHandsWonForTest は取得局数を差し替える (テスト専用)。
func (k *Karnoffel) SetHandsWonForTest(team, n int) {
	if team >= 0 && team < KarnoffelTeamCnt {
		k.handsWon[team] = n
	}
}

// FinishHandForTest は精算を走らせる (テスト専用)。
func (k *Karnoffel) FinishHandForTest() { k.finishHand() }

// ---- JSON ----

// karnoffelJSON is the KV wire format for Karnoffel.
type karnoffelJSON struct {
	Players     []*KarnoffelPlayer      `json:"pl"`
	Config      KarnoffelConfig         `json:"cf"`
	Phase       KarnoffelPhase          `json:"ph"`
	DealerIdx   int                     `json:"di"`
	UpCards     []*Card                 `json:"uc"`
	ChosenSuit  int                     `json:"cs"`
	CurrentIdx  int                     `json:"ci"`
	Trick       []*Card                 `json:"tk"`
	TrickLeader int                     `json:"tl"`
	TrickNumber int                     `json:"tn"`
	TricksWon   [KarnoffelPlayerCnt]int `json:"tw"`
	HandsWon    [KarnoffelTeamCnt]int   `json:"hw"`
	LastResult  *KarnoffelHandResult    `json:"lr"`
	HandNumber  int                     `json:"hn"`
	GameEndFlag bool                    `json:"ge"`
	WinnerTeam  int                     `json:"wt"`
	ActionLog   []*ActionLogEntry       `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (k *Karnoffel) MarshalJSON() ([]byte, error) {
	return json.Marshal(karnoffelJSON{
		Players: k.players, Config: k.config, Phase: k.phase,
		DealerIdx: k.dealerIdx, UpCards: k.upCards, ChosenSuit: k.chosenSuit,
		CurrentIdx: k.currentIdx, Trick: k.trick,
		TrickLeader: k.trickLeader, TrickNumber: k.trickNumber,
		TricksWon: k.tricksWon, HandsWon: k.handsWon,
		LastResult: k.lastResult, HandNumber: k.handNumber,
		GameEndFlag: k.gameEndFlag, WinnerTeam: k.winnerTeam, ActionLog: k.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **KV から戻る値なので範囲を検査する。**壊れた状態をそのまま受け入れると
// 添字で落ちる。
func (k *Karnoffel) UnmarshalJSON(data []byte) error {
	var j karnoffelJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != KarnoffelPlayerCnt {
		return errors.New("karnoffel needs exactly four seats")
	}
	if j.Phase < KarnoffelPhasePlay || j.Phase > KarnoffelPhaseGameEnd {
		return errors.New("unknown phase")
	}
	for name, v := range map[string]int{"dealer": j.DealerIdx, "current seat": j.CurrentIdx, "trick leader": j.TrickLeader} {
		if v < 0 || v >= KarnoffelPlayerCnt {
			return errors.New("bad " + name)
		}
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= KarnoffelTeamCnt {
		return errors.New("bad winner team")
	}
	if j.ChosenSuit < 0 || j.ChosenSuit > CardDesignDiamond {
		return errors.New("bad chosen suit")
	}
	if len(j.Trick) > KarnoffelPlayerCnt {
		return errors.New("a trick cannot hold more cards than there are seats")
	}
	if len(j.UpCards) > KarnoffelPlayerCnt {
		return errors.New("there is one face-up card per seat at most")
	}
	if j.TrickNumber < 0 || j.TrickNumber > KarnoffelTricks {
		return errors.New("bad trick number")
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}

	k.players = j.Players
	k.config = j.Config
	k.phase = j.Phase
	k.dealerIdx = j.DealerIdx
	k.upCards = j.UpCards
	k.chosenSuit = j.ChosenSuit
	k.currentIdx = j.CurrentIdx
	k.trick = j.Trick
	k.trickLeader = j.TrickLeader
	k.trickNumber = j.TrickNumber
	k.tricksWon = j.TricksWon
	k.handsWon = j.HandsWon
	k.lastResult = j.LastResult
	k.handNumber = j.HandNumber
	k.gameEndFlag = j.GameEndFlag
	k.winnerTeam = j.WinnerTeam
	k.actionLog = j.ActionLog
	return nil
}
