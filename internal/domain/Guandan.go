//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"math/rand"
	"sort"
	"strconv"
)

// GuandanPlayerCnt は掼蛋の人数。
const GuandanPlayerCnt = 4

// GuandanTeamCnt はチーム数。
const GuandanTeamCnt = 2

// GuandanHandSize は 1 人あたりの手札枚数。
const GuandanHandSize = 27

// GuandanDeckSize は使用する札数 (**52 × 2 + ジョーカー 4 = 108**)。
const GuandanDeckSize = 108

// レベルの範囲。**2 から A までの 13 段階。**
const (
	// GuandanMinLevel 開始レベル
	GuandanMinLevel = 2
	// GuandanMaxLevel 勝利をかけられるレベル (A)
	GuandanMaxLevel = 14
)

// レベル上昇量。**1 / 2 / 4 であって「1〜3」ではない。**
//
// 上位独占の +4 が出せないと、上位独占を狙う動機そのものが消える。
const (
	// GuandanAdvanceFirstFourth 1着-4着 (味方が最下位) の上昇
	GuandanAdvanceFirstFourth = 1
	// GuandanAdvanceFirstThird 1着-3着 の上昇
	GuandanAdvanceFirstThird = 2
	// GuandanAdvanceFirstSecond 1着-2着 (上位独占) の上昇
	GuandanAdvanceFirstSecond = 4
)

// 序列の目盛り。**レベル札が A と黒ジョーカーのあいだに割り込む**ので、
// 自然序列とは別の段を用意しておく必要がある。
const (
	// guandanRankAce A の自然序列
	guandanRankAce = 14
	// guandanRankLevel レベル札。**A の上、黒ジョーカーの下。**
	guandanRankLevel = 15
	// guandanRankBlackJoker 黒ジョーカー
	guandanRankBlackJoker = 16
	// guandanRankRedJoker 赤ジョーカー。最強。
	guandanRankRedJoker = 17
)

// guandanNaturalRank は札の自然序列を返す (レベルを考えない素の強さ)。
//
// 赤ジョーカー > 黒ジョーカー > A > K > Q > J > 10 > … > 2。
func guandanNaturalRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetDesign() == CardDesignJoker {
		// **赤 (2) が黒 (1) より上。**
		if c.GetValue() >= 2 {
			return guandanRankRedJoker
		}
		return guandanRankBlackJoker
	}
	if c.GetValue() == 1 {
		return guandanRankAce
	}
	return c.GetValue()
}

// GuandanIsJoker は札がジョーカーかを返す。
func GuandanIsJoker(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignJoker
}

// GuandanIsLevelCard は札が現在のレベル札かを返す。
func GuandanIsLevelCard(c *Card, level int) bool {
	if c == nil || GuandanIsJoker(c) {
		return false
	}
	return guandanNaturalRank(c) == level
}

// GuandanIsWild は札がワイルドかを返す。
//
// **ワイルドは「レベル札のうち ♥ のもの」だけ。**2 デッキなのでちょうど 2 枚。
// スートを問わないわけではない。
func GuandanIsWild(c *Card, level int) bool {
	return GuandanIsLevelCard(c, level) && c.GetDesign() == CardDesignHeart
}

// GuandanRank は現在のレベルを踏まえた札の強さを返す。
//
// **レベル札は A より上、黒ジョーカーより下**に入る。本来の位置から抜けるので、
// 序列は局ごとに変わる。固定テーブルでは表せない。
func GuandanRank(c *Card, level int) int {
	if c == nil {
		return 0
	}
	if GuandanIsJoker(c) {
		return guandanNaturalRank(c)
	}
	if GuandanIsLevelCard(c, level) {
		// **A の上、黒ジョーカーの下。**本来の位置から抜けて割り込む。
		return guandanRankLevel
	}
	return guandanNaturalRank(c)
}

// GuandanComboKind は役の種類。
type GuandanComboKind int

// 掼蛋の役。**爆弾は通常役をすべて上回る。**
const (
	// GuandanComboNone 役として成立していない
	GuandanComboNone GuandanComboKind = iota
	// GuandanComboSingle シングル
	GuandanComboSingle
	// GuandanComboPair ペア
	GuandanComboPair
	// GuandanComboTriple トリプル
	GuandanComboTriple
	// GuandanComboFullHouse スリーカード + ペア
	GuandanComboFullHouse
	// GuandanComboStraight 5 枚の連続 (同ランク不可)
	GuandanComboStraight
	// GuandanComboPlate 連続する 2 組のトリプル (飛行機)
	GuandanComboPlate
	// GuandanComboTube 連続する 3 組のペア (木板)
	GuandanComboTube
	// GuandanComboBomb 爆弾 (同ランク 4 枚以上)
	GuandanComboBomb
	// GuandanComboStraightFlush ストレートフラッシュ (同スート 5 枚連続)
	GuandanComboStraightFlush
	// GuandanComboJokerBomb ジョーカー 4 枚。**最強。**
	GuandanComboJokerBomb
)

// GuandanCombo は 1 手の役。
type GuandanCombo struct {
	Kind GuandanComboKind
	// Rank は比較に使う代表ランク。
	Rank int
	// Size は枚数 (爆弾の強さ比較に要る)。
	Size int
}

// GuandanPhase はゲームフェーズ。
type GuandanPhase int

// 掼蛋のフェーズ定数
const (
	// GuandanPhaseTribute 進貢・還貢
	GuandanPhaseTribute GuandanPhase = iota
	// GuandanPhasePlay クライミング
	GuandanPhasePlay
	// GuandanPhaseHandEnd 局終了 (精算済み)
	GuandanPhaseHandEnd
	// GuandanPhaseGameEnd ゲーム終了
	GuandanPhaseGameEnd
)

// GuandanHandResult は 1 局の結果。
type GuandanHandResult struct {
	// Order は上がった順の席。
	Order [GuandanPlayerCnt]int
	// WinnerTeam はこの局を取ったチーム (1 着の側)。
	WinnerTeam int
	// Advance は上昇したレベル数 (1 / 2 / 4)。
	Advance int
	// FirstSecond は上位独占だったか。
	FirstSecond bool
}

// Guandan は掼蛋のゲームクラス。
//
// 中国・江蘇省発祥のレベルアップ式クライミング。**108 枚を 27 枚ずつ**配り、
// **現在のレベル札が A の上に割り込む**ため、序列が局ごとに変わる。
// **♥ のレベル札 2 枚だけがワイルド**で、ジョーカーの代用にはならない。
type Guandan struct {
	players []*GuandanPlayer
	config  GuandanConfig
	phase   GuandanPhase
	// levels は各チームの現在レベル。
	levels [GuandanTeamCnt]int
	// level はこの局の基準レベル (1 着チームのレベル)。
	level int
	// declarerTeam はこの局のレベルを提供しているチーム。
	declarerTeam int
	currentIdx   int
	// lastCombo は場に出ている最後の役 (誰も出していなければ nil)。
	lastCombo *GuandanCombo
	// lastPlayer は lastCombo を出した席。
	lastPlayer int
	// passCount は lastCombo 以降に連続でパスした人数。
	passCount int
	// finished は上がった順の席。
	finished []int
	// tributes はこの局の進貢の記録。
	tributes []*GuandanTribute
	// tributeCancelled は赤ジョーカー保持で貢が取り消されたか。
	tributeCancelled bool
	lastResult       *GuandanHandResult
	handNumber       int
	gameEndFlag      bool
	winnerTeam       int
	actionLog        []*ActionLogEntry
}

// GuandanTribute は 1 件の進貢の記録。
type GuandanTribute struct {
	// From は貢を払った席、To は受け取った席。
	From int
	To   int
	// Card は払われた札 (ワイルドを除く最強の 1 枚)。
	Card *Card
	// Returned は還貢で返された札。
	Returned *Card
}

// NewGuandan コンストラクタ
func NewGuandan(players []*GuandanPlayer, config GuandanConfig) *Guandan {
	g := &Guandan{players: players, config: config, winnerTeam: -1, lastPlayer: -1}
	for i := range GuandanTeamCnt {
		g.levels[i] = GuandanMinLevel
	}
	g.level = GuandanMinLevel
	return g
}

// NewDefaultGuandan は人間 1 人 + CPU 3 体の卓を作る。
func NewDefaultGuandan() *Guandan {
	players := make([]*GuandanPlayer, 0, GuandanPlayerCnt)
	for i := range GuandanPlayerCnt {
		players = append(players, NewGuandanPlayer(i == 0))
	}
	return NewGuandan(players, DefaultGuandanConfig())
}

// GuandanTeamOf は席のチームを返す (範囲外は -1)。
//
// **パートナーは向かい合わせ。**
func GuandanTeamOf(seat int) int {
	if seat < 0 || seat >= GuandanPlayerCnt {
		return -1
	}
	return seat % GuandanTeamCnt
}

// Reset はゲームを初期化する。
func (g *Guandan) Reset() {
	for i := range GuandanTeamCnt {
		g.levels[i] = GuandanMinLevel
	}
	g.level = GuandanMinLevel
	g.declarerTeam = 0
	g.handNumber = 0
	g.gameEndFlag = false
	g.winnerTeam = -1
	g.lastResult = nil
	g.actionLog = make([]*ActionLogEntry, 0)
	g.beginHand(nil)
}

// beginHand は 1 局を配る。prev は前局の結果 (初回は nil)。
func (g *Guandan) beginHand(prev *GuandanHandResult) {
	g.handNumber++
	g.finished = make([]int, 0, GuandanPlayerCnt)
	g.tributes = make([]*GuandanTribute, 0)
	g.tributeCancelled = false
	g.lastCombo = nil
	g.lastPlayer = -1
	g.passCount = 0
	for i := range GuandanPlayerCnt {
		if p := g.GetPlayer(i); p != nil {
			p.ResetRound()
		}
	}
	g.dealRound()

	// **2 局目以降は貢から始まる。**前局の結果が次局の手札を動かす。
	if prev != nil {
		g.phase = GuandanPhaseTribute
		g.prepareTribute(prev)
		g.currentIdx = prev.Order[0]
		g.addLog(-1, "deal", "hand "+strconv.Itoa(g.handNumber)+" level "+strconv.Itoa(g.level), nil)
		return
	}
	g.phase = GuandanPhasePlay
	g.currentIdx = 0
	g.addLog(-1, "deal", "hand "+strconv.Itoa(g.handNumber)+" level "+strconv.Itoa(g.level), nil)
}

// dealRound は 108 枚を 27 枚ずつ配る。
func (g *Guandan) dealRound() {
	deck := newGuandanDeck()
	guandanShuffle(deck)
	pos := 0
	for range GuandanHandSize {
		for i := range GuandanPlayerCnt {
			p := g.GetPlayer(i)
			if pos < len(deck) && p != nil {
				p.AddCard(deck[pos])
				pos++
			}
		}
	}
	// **27 枚を添字で指定して役を組む**ので、並べ替えておかないと実用に耐えない。
	for i := range GuandanPlayerCnt {
		g.guandanSortHand(i)
	}
}

// guandanSortHand は手札をこの局の序列で並べ替える。
//
// 同じランクが隣り合っていないと、ペアやトリプルを添字で拾えない。**レベル札は
// A の上に固まる**ので、ワイルドの ♥ もその並びの中に入る。
func (g *Guandan) guandanSortHand(seat int) {
	p := g.GetPlayer(seat)
	if p == nil {
		return
	}
	cards := make([]*Card, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		ri, rj := GuandanRank(cards[i], g.level), GuandanRank(cards[j], g.level)
		if ri != rj {
			return ri < rj
		}
		return cards[i].GetDesign() < cards[j].GetDesign()
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// newGuandanDeck は **52 × 2 + ジョーカー 4 = 108 枚**を作る。
func newGuandanDeck() []*Card {
	cards := make([]*Card, 0, GuandanDeckSize)
	for range 2 {
		for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
			for v := 1; v <= 13; v++ {
				cards = append(cards, NewCard(suit, v, true))
			}
		}
		// **黒 1 枚・赤 1 枚を 2 デッキぶんで計 4 枚。**
		cards = append(cards, NewCard(CardDesignJoker, 1, true))
		cards = append(cards, NewCard(CardDesignJoker, 2, true))
	}
	return cards
}

// guandanShuffle は札をシャッフルする。
func guandanShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1) //nolint:gosec // ゲームのシャッフルに暗号強度は要らない
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// ---- 進貢・還貢 ----

// guandanRedJokerCount は席が持つ赤ジョーカーの枚数を返す。
func (g *Guandan) guandanRedJokerCount(seat int) int {
	p := g.GetPlayer(seat)
	if p == nil {
		return 0
	}
	n := 0
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if GuandanIsJoker(c) && c.GetValue() >= 2 {
			n++
		}
	}
	return n
}

// GuandanTributeCancelled は赤ジョーカー保持で貢が取り消されるかを返す。
//
// **上位独占の次**なら「敗者 2 人が 1 枚ずつ」または「片方が 2 枚」で取り消し。
// **それ以外**なら「最下位が 2 枚」で取り消し。
func (g *Guandan) GuandanTributeCancelled(prev *GuandanHandResult) bool {
	if prev == nil {
		return false
	}
	if prev.FirstSecond {
		a, b := prev.Order[2], prev.Order[3]
		ca, cb := g.guandanRedJokerCount(a), g.guandanRedJokerCount(b)
		return (ca >= 1 && cb >= 1) || ca >= 2 || cb >= 2
	}
	return g.guandanRedJokerCount(prev.Order[3]) >= 2
}

// prepareTribute は進貢を決める。
//
// **上位独占の次は敗者 2 人が払い、それ以外は最下位だけが払う。**払うのは
// **ワイルドを除く最強の 1 枚**。
func (g *Guandan) prepareTribute(prev *GuandanHandResult) {
	if g.GuandanTributeCancelled(prev) {
		g.tributeCancelled = true
		g.phase = GuandanPhasePlay
		g.currentIdx = prev.Order[0]
		g.addLog(-1, "tributeCancelled", "", nil)
		return
	}

	payers := []int{prev.Order[3]}
	if prev.FirstSecond {
		payers = []int{prev.Order[2], prev.Order[3]}
	}
	receivers := []int{prev.Order[0], prev.Order[1]}
	for i, from := range payers {
		to := receivers[0]
		if prev.FirstSecond && i < len(receivers) {
			to = receivers[i]
		}
		idx := g.guandanHighestNonWild(from)
		if idx < 0 {
			continue
		}
		p := g.GetPlayer(from)
		c := p.GetCard(idx)
		p.RemoveCard(idx)
		if r := g.GetPlayer(to); r != nil {
			r.AddCard(c)
			g.guandanSortHand(to)
		}
		g.tributes = append(g.tributes, &GuandanTribute{From: from, To: to, Card: c})
		g.addLog(from, "tribute", strconv.Itoa(to), []*Card{c})
	}
	// 還貢は受け取った側が選ぶ。CPU は自動で返す。

	// **誰も貢を払えなかったら還貢のフェーズに入らない。**入ってしまうと、
	// 返す相手のいない席が延々と還貢を試みて局が進まなくなる。
	if len(g.tributes) == 0 {
		g.phase = GuandanPhasePlay
	}
}

// guandanHighestNonWild は席の**ワイルドを除く最強の 1 枚**の位置を返す。
func (g *Guandan) guandanHighestNonWild(seat int) int {
	p := g.GetPlayer(seat)
	if p == nil {
		return -1
	}
	best, bestRank := -1, -1
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if GuandanIsWild(c, g.level) {
			continue
		}
		if r := GuandanRank(c, g.level); r > bestRank {
			best, bestRank = i, r
		}
	}
	return best
}

// ReturnTribute は還貢する (受け取った側が不要な札を 1 枚返す)。
func (g *Guandan) ReturnTribute(player, idx int) error {
	if g.gameEndFlag {
		return errors.New("the game is over")
	}
	if g.phase != GuandanPhaseTribute {
		return errors.New("it is not the tribute phase")
	}
	var t *GuandanTribute
	for _, x := range g.tributes {
		if x.To == player && x.Returned == nil {
			t = x
			break
		}
	}
	if t == nil {
		return errors.New("you owe no return tribute")
	}
	p := g.GetPlayer(player)
	if p == nil || idx < 0 || idx >= p.GetCardsSize() {
		return errors.New("there is no such card")
	}
	c := p.GetCard(idx)
	p.RemoveCard(idx)
	if r := g.GetPlayer(t.From); r != nil {
		r.AddCard(c)
		g.guandanSortHand(t.From)
	}
	t.Returned = c
	g.addLog(player, "returnTribute", strconv.Itoa(t.From), []*Card{c})

	for _, x := range g.tributes {
		if x.Returned == nil {
			g.currentIdx = x.To
			return nil
		}
	}
	// 全部返し終えたらプレイへ。**最初に出すのは最大の貢を払った側。**
	g.phase = GuandanPhasePlay
	if len(g.tributes) > 0 {
		g.currentIdx = g.tributes[0].From
	}
	return nil
}

// ---- 役の判定 ----

// GuandanEvaluate は出そうとしている札から役を作る。
//
// **♥ のレベル札はワイルド**で、ジョーカー以外の任意の札として使える。
// 探索が要るのはこのため。
func GuandanEvaluate(cards []*Card, level int) *GuandanCombo {
	if len(cards) == 0 {
		return nil
	}
	wilds := 0
	fixed := make([]*Card, 0, len(cards))
	for _, c := range cards {
		if c == nil {
			return nil
		}
		if GuandanIsWild(c, level) {
			wilds++
			continue
		}
		fixed = append(fixed, c)
	}

	// **ジョーカー 4 枚が最強。**ワイルドでは作れない。
	if len(cards) == 4 && wilds == 0 {
		allJoker := true
		for _, c := range fixed {
			if !GuandanIsJoker(c) {
				allJoker = false
			}
		}
		if allJoker {
			return &GuandanCombo{Kind: GuandanComboJokerBomb, Rank: 100, Size: 4}
		}
	}

	counts := map[int]int{}
	for _, c := range fixed {
		counts[GuandanRank(c, level)]++
	}
	// **連続の判定だけは自然位置で数える。**レベル札が A の上に移るのは単札と
	// 対子の強弱の話で、レベルが 5 のとき 4-5-6 の連続ペアは正当な役である。
	// 序列つきの counts をそのまま渡すと、5 が 15 の枠に入って窓が抜ける。
	naturalCounts := map[int]int{}
	for _, c := range fixed {
		naturalCounts[guandanNaturalRank(c)]++
	}

	// 同ランクにまとめられるか (ワイルドを足して n 枚)。
	if len(counts) <= 1 {
		rank := 0
		for r := range counts {
			rank = r
		}
		if rank == 0 && wilds > 0 {
			// ワイルドだけ。最弱として扱う。
			rank = GuandanMinLevel
		}
		switch n := len(cards); {
		case n == 1:
			return &GuandanCombo{Kind: GuandanComboSingle, Rank: rank, Size: 1}
		case n == 2:
			return &GuandanCombo{Kind: GuandanComboPair, Rank: rank, Size: 2}
		case n == 3:
			return &GuandanCombo{Kind: GuandanComboTriple, Rank: rank, Size: 3}
		case n >= 4:
			// **同ランク 4 枚以上は爆弾。**枚数が多いほど強い。
			return &GuandanCombo{Kind: GuandanComboBomb, Rank: rank, Size: n}
		}
	}

	if c := guandanFullHouse(counts, wilds, len(cards)); c != nil {
		return c
	}
	if c := guandanStraight(fixed, wilds, len(cards), level); c != nil {
		return c
	}
	if c := guandanRepeatedRun(naturalCounts, wilds, len(cards)); c != nil {
		return c
	}
	return nil
}

// guandanFullHouse はスリーカード + ペアを判定する。
func guandanFullHouse(counts map[int]int, wilds, total int) *GuandanCombo {
	if total != 5 {
		return nil
	}
	ranks := guandanSortedRanks(counts)
	for _, three := range ranks {
		need3 := 3 - counts[three]
		if need3 < 0 {
			need3 = 0
		}
		for _, two := range ranks {
			if two == three {
				continue
			}
			need2 := 2 - counts[two]
			if need2 < 0 {
				need2 = 0
			}
			if need3+need2 <= wilds && counts[three]+counts[two]+wilds >= 5 {
				return &GuandanCombo{Kind: GuandanComboFullHouse, Rank: three, Size: 5}
			}
		}
	}
	return nil
}

// guandanStraight は 5 枚の連続を判定する。
//
// **A は上にも下にも使える** (A-2-3-4-5 と 10-J-Q-K-A)。
func guandanStraight(fixed []*Card, wilds, total, level int) *GuandanCombo {
	if total != 5 {
		return nil
	}
	// 連続の判定では**レベル札も自然位置で数える**。割り込むのは単札の強弱だけ。
	base := make([]int, 0, len(fixed))
	sameSuit := true
	suit := -1
	for _, c := range fixed {
		if GuandanIsJoker(c) {
			return nil
		}
		v := c.GetValue()
		if v == 1 {
			v = 14
		}
		base = append(base, v)
		if suit == -1 {
			suit = c.GetDesign()
		} else if suit != c.GetDesign() {
			sameSuit = false
		}
	}
	for _, low := range []bool{false, true} {
		vals := make([]int, len(base))
		copy(vals, base)
		if low {
			for i, v := range vals {
				if v == 14 {
					vals[i] = 1
				}
			}
		}
		sort.Ints(vals)
		if guandanHasDuplicate(vals) {
			continue
		}
		if len(vals) == 0 {
			continue
		}
		span := vals[len(vals)-1] - vals[0] + 1
		if span > 5 {
			continue
		}
		if 5-len(vals) <= wilds {
			kind := GuandanComboStraight
			if sameSuit && wilds == 0 {
				kind = GuandanComboStraightFlush
			}
			// **余ったワイルドは上に伸ばす。**5-6-7-8 + ワイルドは 4-5-6-7-8 とも
			// 5-6-7-8-9 とも読めるので、弱いほうで確定させると出せる手が減る。
			// 内側の穴埋めに span-len 枚使うので、残りは 5-span 枚。
			top := min(vals[len(vals)-1]+5-span, 14)
			return &GuandanCombo{Kind: kind, Rank: top, Size: 5}
		}
	}
	return nil
}

// guandanHasDuplicate は昇順の並びに重複があるかを返す。
func guandanHasDuplicate(vals []int) bool {
	for i := 1; i < len(vals); i++ {
		if vals[i] == vals[i-1] {
			return true
		}
	}
	return false
}

// guandanRepeatedRun は飛行機 (連続トリプル 2 組) と木板 (連続ペア 3 組) を判定する。
func guandanRepeatedRun(counts map[int]int, wilds, total int) *GuandanCombo {
	switch total {
	case 6:
		if c := guandanRun(counts, wilds, 2, 3, GuandanComboPlate); c != nil {
			return c
		}
		return guandanRun(counts, wilds, 3, 2, GuandanComboTube)
	}
	return nil
}

// guandanRun は「連続する groups 組の each 枚」を判定する。
func guandanRun(counts map[int]int, wilds, groups, each int, kind GuandanComboKind) *GuandanCombo {
	for start := GuandanMinLevel; start+groups-1 <= 14; start++ {
		need := 0
		for i := range groups {
			have := counts[start+i]
			if have > each {
				have = each
			}
			need += each - have
		}
		if need <= wilds {
			return &GuandanCombo{Kind: kind, Rank: start + groups - 1, Size: groups * each}
		}
	}
	return nil
}

// guandanSortedRanks は counts のランクを昇順で返す (決定的にするため)。
func guandanSortedRanks(counts map[int]int) []int {
	out := make([]int, 0, len(counts))
	for r := range counts {
		out = append(out, r)
	}
	sort.Ints(out)
	return out
}

// GuandanBeats は c が prev を上回るかを返す。
//
// **爆弾は通常役をすべて上回る。**爆弾どうしは種類、次に枚数、次にランク。
func GuandanBeats(c, prev *GuandanCombo) bool {
	if c == nil {
		return false
	}
	if prev == nil {
		return true
	}
	cb, pb := guandanBombTier(c.Kind), guandanBombTier(prev.Kind)
	if cb != pb {
		return cb > pb
	}
	if cb > 0 {
		if c.Size != prev.Size {
			return c.Size > prev.Size
		}
		return c.Rank > prev.Rank
	}
	// 通常役どうしは同じ種類・同じ枚数でしか比べられない。
	if c.Kind != prev.Kind || c.Size != prev.Size {
		return false
	}
	return c.Rank > prev.Rank
}

// guandanBombTier は爆弾の階層を返す (0 なら通常役)。
func guandanBombTier(k GuandanComboKind) int {
	switch k {
	case GuandanComboBomb:
		return 1
	case GuandanComboStraightFlush:
		return 2
	case GuandanComboJokerBomb:
		return 3
	}
	return 0
}

// ---- プレイ ----

// PlayCards は手札から役を出す。
func (g *Guandan) PlayCards(player int, idxs []int) error {
	if g.gameEndFlag {
		return errors.New("the game is over")
	}
	if g.phase != GuandanPhasePlay {
		return errors.New("it is not the play phase")
	}
	if player != g.currentIdx {
		return errors.New("it is not your turn")
	}
	p := g.GetPlayer(player)
	if p == nil || len(idxs) == 0 {
		return errors.New("you must play at least one card")
	}
	seen := map[int]bool{}
	cards := make([]*Card, 0, len(idxs))
	for _, i := range idxs {
		if i < 0 || i >= p.GetCardsSize() || seen[i] {
			return errors.New("there is no such card")
		}
		seen[i] = true
		cards = append(cards, p.GetCard(i))
	}
	combo := GuandanEvaluate(cards, g.level)
	if combo == nil {
		return errors.New("that is not a legal combination")
	}
	if !GuandanBeats(combo, g.lastCombo) {
		return errors.New("that does not beat the last play")
	}

	sorted := append([]int(nil), idxs...)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	for _, i := range sorted {
		p.RemoveCard(i)
	}
	g.lastCombo = combo
	g.lastPlayer = player
	g.passCount = 0
	g.addLog(player, "play", strconv.Itoa(int(combo.Kind)), cards)

	if p.GetCardsSize() == 0 {
		g.finished = append(g.finished, player)
		g.addLog(player, "out", strconv.Itoa(len(g.finished)), nil)
		if len(g.finished) >= GuandanPlayerCnt-1 {
			g.finishHand()
			return nil
		}
	}
	g.advanceTurn()
	return nil
}

// Pass はパスする。
func (g *Guandan) Pass(player int) error {
	if g.gameEndFlag {
		return errors.New("the game is over")
	}
	if g.phase != GuandanPhasePlay {
		return errors.New("it is not the play phase")
	}
	if player != g.currentIdx {
		return errors.New("it is not your turn")
	}
	if g.lastCombo == nil {
		return errors.New("you must lead")
	}
	g.passCount++
	g.addLog(player, "pass", "", nil)
	g.advanceTurn()
	return nil
}

// guandanActiveCount はまだ手札のある席の数を返す。
func (g *Guandan) guandanActiveCount() int {
	n := 0
	for i := range GuandanPlayerCnt {
		if p := g.GetPlayer(i); p != nil && p.GetCardsSize() > 0 {
			n++
		}
	}
	return n
}

// advanceTurn は次の手番へ進める。
func (g *Guandan) advanceTurn() {
	for i := 1; i <= GuandanPlayerCnt; i++ {
		seat := (g.currentIdx + i) % GuandanPlayerCnt
		p := g.GetPlayer(seat)
		if p == nil || p.GetCardsSize() == 0 {
			continue
		}
		g.currentIdx = seat
		// **一周してリードした本人に戻ったら場が流れる。**
		if seat == g.lastPlayer {
			g.lastCombo = nil
			g.passCount = 0
			return
		}
		// **出した本人が上がっている場合は戻ってこない。**残っている全員が
		// パスした時点で流す。これが無いと全員パスのまま永久に回る。
		if g.lastCombo != nil && g.passCount >= g.guandanActiveCount() {
			g.lastCombo = nil
			g.passCount = 0
		}
		return
	}
	g.finishHand()
}

// finishHand は 1 局を精算する。
//
// **上昇量は 1 / 2 / 4。**上位独占が +4 であることがこのゲームの動機。
func (g *Guandan) finishHand() {
	for i := range GuandanPlayerCnt {
		if !guandanContains(g.finished, i) {
			g.finished = append(g.finished, i)
		}
	}
	var order [GuandanPlayerCnt]int
	copy(order[:], g.finished)

	first := order[0]
	team := GuandanTeamOf(first)
	// 味方が何着だったか。
	partnerPos := 0
	for pos, seat := range order {
		if seat != first && GuandanTeamOf(seat) == team {
			partnerPos = pos
		}
	}
	advance := GuandanAdvanceFirstFourth
	switch partnerPos {
	case 1:
		advance = GuandanAdvanceFirstSecond
	case 2:
		advance = GuandanAdvanceFirstThird
	}

	res := &GuandanHandResult{
		Order:       order,
		WinnerTeam:  team,
		Advance:     advance,
		FirstSecond: partnerPos == 1,
	}
	g.lastResult = res
	g.phase = GuandanPhaseHandEnd
	g.addLog(-1, "handEnd", strconv.Itoa(team)+" +"+strconv.Itoa(advance), nil)

	// **A レベルで 1-2 または 1-3 を取れば勝ち。**それ以外に勝ち方は無い。
	if g.levels[team] >= GuandanMaxLevel && partnerPos <= 2 {
		g.gameEndFlag = true
		g.phase = GuandanPhaseGameEnd
		g.winnerTeam = team
		return
	}
	next := g.levels[team] + advance
	if next > GuandanMaxLevel {
		next = GuandanMaxLevel
	}
	g.levels[team] = next
}

// guandanContains は s に v が含まれるかを返す。
func guandanContains(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// NextHand は次の局を配る。
func (g *Guandan) NextHand() error {
	if g.gameEndFlag {
		return errors.New("the game is over")
	}
	if g.phase != GuandanPhaseHandEnd {
		return errors.New("the hand is still in progress")
	}
	prev := g.lastResult
	g.declarerTeam = prev.WinnerTeam
	g.level = g.levels[prev.WinnerTeam]
	g.beginHand(prev)
	return nil
}

// ---- CPU ----

// GuandanCpuPlay は CPU が出す手札インデックスを返す (出せなければ nil)。
//
// **場が空ならいちばん弱い単札でリードし、そうでなければ越えられる最小の手を
// 探す。**ワイルドは温存する (任意の札になるので終盤の価値が高い)。
func (g *Guandan) GuandanCpuPlay(seat int) []int {
	p := g.GetPlayer(seat)
	if p == nil || p.GetCardsSize() == 0 {
		return nil
	}
	// 単札を弱い順に並べる。
	order := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		order = append(order, i)
	}
	sort.SliceStable(order, func(a, b int) bool {
		return GuandanRank(p.GetCard(order[a]), g.level) < GuandanRank(p.GetCard(order[b]), g.level)
	})

	if g.lastCombo == nil {
		// **ワイルドは最後まで取っておく。**
		for _, i := range order {
			if !GuandanIsWild(p.GetCard(i), g.level) {
				return []int{i}
			}
		}
		return []int{order[0]}
	}

	// 同じ形・同じ枚数で越えられる最小の手を探す。
	if idxs := g.guandanFindBeating(seat, order); idxs != nil {
		return idxs
	}
	return nil
}

// guandanFindBeating は場を越えられる手を探す (無ければ nil)。
func (g *Guandan) guandanFindBeating(seat int, order []int) []int {
	p := g.GetPlayer(seat)
	size := g.lastCombo.Size
	if size <= 0 || size > p.GetCardsSize() {
		return nil
	}
	// 枚数が同じ組み合わせを、弱い順に総当たりする。**27 枚なので
	// 無制限の全探索はできない。**同ランクの塊と連番だけを候補にする。
	byRank := map[int][]int{}
	for _, i := range order {
		c := p.GetCard(i)
		if GuandanIsWild(c, g.level) {
			continue
		}
		r := GuandanRank(c, g.level)
		byRank[r] = append(byRank[r], i)
	}
	ranks := make([]int, 0, len(byRank))
	for r := range byRank {
		ranks = append(ranks, r)
	}
	sort.Ints(ranks)

	for _, r := range ranks {
		group := byRank[r]
		if len(group) < size {
			continue
		}
		idxs := group[:size]
		cards := make([]*Card, 0, size)
		for _, i := range idxs {
			cards = append(cards, p.GetCard(i))
		}
		combo := GuandanEvaluate(cards, g.level)
		if GuandanBeats(combo, g.lastCombo) {
			return append([]int(nil), idxs...)
		}
	}

	// **連続役には連続役で応じる。**同ランクの塊しか見ないと、チューブや
	// プレートには爆弾を切るしかなくなり、爆弾を無駄に失う。
	if k := g.lastCombo.Kind; k == GuandanComboTube || k == GuandanComboPlate {
		if idxs := g.guandanFindRunBeating(seat, order, k); idxs != nil {
			return idxs
		}
	}

	// 爆弾で流す手があるなら使う。
	for _, r := range ranks {
		group := byRank[r]
		if len(group) < 4 {
			continue
		}
		cards := make([]*Card, 0, len(group))
		for _, i := range group {
			cards = append(cards, p.GetCard(i))
		}
		if combo := GuandanEvaluate(cards, g.level); GuandanBeats(combo, g.lastCombo) {
			return append([]int(nil), group...)
		}
	}
	return nil
}

// guandanFindRunBeating は連続役 (チューブ・プレート) で場を越える手を探す。
//
// **連続の判定は自然位置。**レベル札も本来のランクで数える。
func (g *Guandan) guandanFindRunBeating(seat int, order []int, kind GuandanComboKind) []int {
	p := g.GetPlayer(seat)
	each, groups := 2, 3
	if kind == GuandanComboPlate {
		each, groups = 3, 2
	}
	byNatural := map[int][]int{}
	for _, i := range order {
		c := p.GetCard(i)
		if GuandanIsWild(c, g.level) {
			continue
		}
		r := guandanNaturalRank(c)
		byNatural[r] = append(byNatural[r], i)
	}
	for start := GuandanMinLevel; start+groups-1 <= 14; start++ {
		idxs := make([]int, 0, each*groups)
		for k := range groups {
			group := byNatural[start+k]
			if len(group) < each {
				idxs = nil
				break
			}
			idxs = append(idxs, group[:each]...)
		}
		if idxs == nil {
			continue
		}
		cards := make([]*Card, 0, len(idxs))
		for _, i := range idxs {
			cards = append(cards, p.GetCard(i))
		}
		if combo := GuandanEvaluate(cards, g.level); GuandanBeats(combo, g.lastCombo) {
			return idxs
		}
	}
	return nil
}

// IsHumanTurn は現在の手番が人間かを返す。
func (g *Guandan) IsHumanTurn() bool {
	if g.gameEndFlag || g.phase == GuandanPhaseHandEnd || g.phase == GuandanPhaseGameEnd {
		return false
	}
	p := g.GetPlayer(g.currentIdx)
	return p != nil && p.GetIsHuman()
}

// CpuPlay は CPU が 1 アクション実行する。
func (g *Guandan) CpuPlay() {
	if g.gameEndFlag || g.IsHumanTurn() {
		return
	}
	seat := g.currentIdx
	if g.phase == GuandanPhaseTribute {
		// **還貢はいちばん弱い札を返す。**
		p := g.GetPlayer(seat)
		if p == nil || p.GetCardsSize() == 0 {
			return
		}
		low, lowRank := 0, 1<<30
		for i := range p.GetCardsSize() {
			c := p.GetCard(i)
			if GuandanIsWild(c, g.level) {
				continue
			}
			if r := GuandanRank(c, g.level); r < lowRank {
				low, lowRank = i, r
			}
		}
		_ = g.ReturnTribute(seat, low)
		return
	}
	if g.phase != GuandanPhasePlay {
		return
	}
	if idxs := g.GuandanCpuPlay(seat); len(idxs) > 0 {
		if g.PlayCards(seat, idxs) == nil {
			return
		}
	}
	if g.lastCombo != nil {
		_ = g.Pass(seat)
		return
	}
	// リードなのに何も出せないことは無いはずだが、詰み回避に 1 枚出す。
	if p := g.GetPlayer(seat); p != nil && p.GetCardsSize() > 0 {
		_ = g.PlayCards(seat, []int{0})
	}
}

// ---- アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (g *Guandan) GetPlayers() []*GuandanPlayer { return g.players }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (g *Guandan) GetPlayer(idx int) *GuandanPlayer {
	return getPlayer(g.players, idx)
}

// GetPhase は現在のフェーズを返す。
func (g *Guandan) GetPhase() GuandanPhase { return g.phase }

// GetCurrentPlayerIdx は現在の手番を返す。
func (g *Guandan) GetCurrentPlayerIdx() int { return g.currentIdx }

// GetLevel はこの局の基準レベルを返す。
func (g *Guandan) GetLevel() int { return g.level }

// GetTeamLevel はチームの現在レベルを返す。
func (g *Guandan) GetTeamLevel(team int) int {
	if team < 0 || team >= GuandanTeamCnt {
		return GuandanMinLevel
	}
	return g.levels[team]
}

// GetDeclarerTeam はこの局のレベルを提供しているチームを返す。
func (g *Guandan) GetDeclarerTeam() int { return g.declarerTeam }

// GetLastCombo は場に出ている最後の役を返す。
func (g *Guandan) GetLastCombo() *GuandanCombo { return g.lastCombo }

// GetLastPlayerIdx は最後に出した席を返す (誰も出していなければ -1)。
func (g *Guandan) GetLastPlayerIdx() int { return g.lastPlayer }

// GetFinished は上がった順の席を返す。
func (g *Guandan) GetFinished() []int { return g.finished }

// GetTributes はこの局の進貢を返す。
func (g *Guandan) GetTributes() []*GuandanTribute { return g.tributes }

// IsTributeCancelled は赤ジョーカー保持で貢が取り消されたかを返す。
func (g *Guandan) IsTributeCancelled() bool { return g.tributeCancelled }

// GetLastResult は直前の局の結果を返す (まだ無ければ nil)。
func (g *Guandan) GetLastResult() *GuandanHandResult { return g.lastResult }

// GetHandNumber は現在の局番号を返す。
func (g *Guandan) GetHandNumber() int { return g.handNumber }

// GetGameEndFlag はゲーム終了フラグを返す。
func (g *Guandan) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerTeam は勝利チームを返す (未確定なら -1)。
func (g *Guandan) GetWinnerTeam() int { return g.winnerTeam }

// GetConfig はゲーム設定を返す。
func (g *Guandan) GetConfig() GuandanConfig { return g.config }

// SetConfig はゲーム設定をセットする。
func (g *Guandan) SetConfig(c GuandanConfig) { g.config = c }

// GetActionLog は棋譜を返す。
func (g *Guandan) GetActionLog() []*ActionLogEntry { return g.actionLog }

// addLog は棋譜を 1 件追加する。
func (g *Guandan) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// ---- テスト用 ----

// SetPhaseForTest はフェーズを差し替える (テスト専用)。
func (g *Guandan) SetPhaseForTest(p GuandanPhase) { g.phase = p }

// SetHandForTest は手札を差し替える (テスト専用)。
func (g *Guandan) SetHandForTest(idx int, cards []*Card) {
	setHandForTest(g.GetPlayer(idx), cards)
}

// SetLevelForTest は基準レベルを差し替える (テスト専用)。
func (g *Guandan) SetLevelForTest(level int) { g.level = level }

// SetTeamLevelForTest はチームのレベルを差し替える (テスト専用)。
func (g *Guandan) SetTeamLevelForTest(team, level int) {
	if team >= 0 && team < GuandanTeamCnt {
		g.levels[team] = level
	}
}

// SetCurrentPlayerForTest は手番を差し替える (テスト専用)。
func (g *Guandan) SetCurrentPlayerForTest(idx int) { g.currentIdx = idx }

// SetFinishedForTest は上がり順を差し替える (テスト専用)。
func (g *Guandan) SetFinishedForTest(order []int) { g.finished = order }

// FinishHandForTest は精算を走らせる (テスト専用)。
func (g *Guandan) FinishHandForTest() { g.finishHand() }

// PrepareTributeForTest は進貢を走らせる (テスト専用)。
func (g *Guandan) PrepareTributeForTest(prev *GuandanHandResult) { g.prepareTribute(prev) }

// ---- JSON ----

// guandanJSON is the KV wire format for Guandan.
type guandanJSON struct {
	Players          []*GuandanPlayer    `json:"pl"`
	Config           GuandanConfig       `json:"cf"`
	Phase            GuandanPhase        `json:"ph"`
	Levels           [GuandanTeamCnt]int `json:"lv"`
	Level            int                 `json:"lc"`
	DeclarerTeam     int                 `json:"dt"`
	CurrentIdx       int                 `json:"ci"`
	LastCombo        *GuandanCombo       `json:"lb"`
	LastPlayer       int                 `json:"lp"`
	PassCount        int                 `json:"pc"`
	Finished         []int               `json:"fi"`
	Tributes         []*GuandanTribute   `json:"tb"`
	TributeCancelled bool                `json:"tc"`
	LastResult       *GuandanHandResult  `json:"lr"`
	HandNumber       int                 `json:"hn"`
	GameEndFlag      bool                `json:"ge"`
	WinnerTeam       int                 `json:"wt"`
	ActionLog        []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Guandan) MarshalJSON() ([]byte, error) {
	return json.Marshal(guandanJSON{
		Players: g.players, Config: g.config, Phase: g.phase,
		Levels: g.levels, Level: g.level, DeclarerTeam: g.declarerTeam,
		CurrentIdx: g.currentIdx, LastCombo: g.lastCombo, LastPlayer: g.lastPlayer,
		PassCount: g.passCount, Finished: g.finished,
		Tributes: g.tributes, TributeCancelled: g.tributeCancelled,
		LastResult: g.lastResult, HandNumber: g.handNumber,
		GameEndFlag: g.gameEndFlag, WinnerTeam: g.winnerTeam, ActionLog: g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **KV から戻る値なので範囲を検査する。**壊れた状態をそのまま受け入れると
// 添字で落ちる。
func (g *Guandan) UnmarshalJSON(data []byte) error {
	var j guandanJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != GuandanPlayerCnt {
		return errors.New("guandan needs exactly four seats")
	}
	if j.Phase < GuandanPhaseTribute || j.Phase > GuandanPhaseGameEnd {
		return errors.New("unknown phase")
	}
	if j.CurrentIdx < 0 || j.CurrentIdx >= GuandanPlayerCnt {
		return errors.New("bad current seat")
	}
	if j.LastPlayer < -1 || j.LastPlayer >= GuandanPlayerCnt {
		return errors.New("bad last player")
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= GuandanTeamCnt {
		return errors.New("bad winner team")
	}
	if j.DeclarerTeam < 0 || j.DeclarerTeam >= GuandanTeamCnt {
		return errors.New("bad declarer team")
	}
	if j.Level < GuandanMinLevel || j.Level > GuandanMaxLevel {
		return errors.New("bad level")
	}
	for _, lv := range j.Levels {
		if lv < GuandanMinLevel || lv > GuandanMaxLevel {
			return errors.New("bad team level")
		}
	}
	if len(j.Finished) > GuandanPlayerCnt {
		return errors.New("more finishers than seats")
	}
	for _, seat := range j.Finished {
		if seat < 0 || seat >= GuandanPlayerCnt {
			return errors.New("bad finisher seat")
		}
	}
	for _, t := range j.Tributes {
		if t == nil {
			continue
		}
		if t.From < 0 || t.From >= GuandanPlayerCnt || t.To < 0 || t.To >= GuandanPlayerCnt {
			return errors.New("bad tribute seat")
		}
	}
	if j.LastCombo != nil && (j.LastCombo.Kind < GuandanComboNone || j.LastCombo.Kind > GuandanComboJokerBomb) {
		return errors.New("unknown combination")
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}

	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.levels = j.Levels
	g.level = j.Level
	g.declarerTeam = j.DeclarerTeam
	g.currentIdx = j.CurrentIdx
	g.lastCombo = j.LastCombo
	g.lastPlayer = j.LastPlayer
	g.passCount = j.PassCount
	g.finished = j.Finished
	g.tributes = j.Tributes
	g.tributeCancelled = j.TributeCancelled
	g.lastResult = j.LastResult
	g.handNumber = j.HandNumber
	g.gameEndFlag = j.GameEndFlag
	g.winnerTeam = j.WinnerTeam
	g.actionLog = j.ActionLog
	return nil
}
