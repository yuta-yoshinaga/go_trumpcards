//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
	"math/rand"
	"sort"
	"strconv"
)

// ShengJiPlayerCnt は升级の人数。
const ShengJiPlayerCnt = 4

// ShengJiTeamCnt はチーム数。
const ShengJiTeamCnt = 2

// ShengJiHandSize は 1 人あたりの手札枚数。
//
// **108 を 4 で割らない。**25 枚ずつ配って 8 枚を底牌として伏せる
// (25 × 4 + 8 = 108)。底牌は飾りではなく、守備側が最終トリックを取ると
// 倍率つきで守備側の得点になる。
const ShengJiHandSize = 25

// ShengJiKittySize は底牌 (扣底) の枚数。
const ShengJiKittySize = 8

// ShengJiDeckSize は使用する札数 (**52 × 2 + ジョーカー 4 = 108**)。
const ShengJiDeckSize = 108

// レベルの範囲。**2 から A までの 13 段階。**
const (
	// ShengJiMinLevel 開始レベル
	ShengJiMinLevel = 2
	// ShengJiMaxLevel 勝利をかけられるレベル (A)
	ShengJiMaxLevel = 14
)

// 得点の定数。**点を集めるのは守備側**で、宣言側は守備側を抑えて勝つ。
const (
	// ShengJiTotalPoints 1 局の総得点 (5 が 8 枚 + 10 と K が各 8 枚)
	ShengJiTotalPoints = 200
	// ShengJiDefenderTarget 守備側が宣言側を降ろすのに要る点 (総得点の 4 割)
	ShengJiDefenderTarget = 80
	// ShengJiAdvanceStep 守備側が 80 点を超えてから 1 段階上がるごとの点
	ShengJiAdvanceStep = 40
)

// ShengJiNoTrump は無主 (切札スート無し) を表す。**レベル札とジョーカーだけが切札。**
const ShengJiNoTrump = 0

// 切札群の段。**切札は「切札スート」だけではない。**全スートのレベル札と
// ジョーカー 4 枚も切札で、この順に強い。
const (
	shengJiTierPlain      = 0 // 非切札
	shengJiTierTrumpSuit  = 1 // 切札スートの平札
	shengJiTierOffLevel   = 2 // 他スートのレベル札 (互いに同格)
	shengJiTierTrumpLevel = 3 // 切札スートのレベル札
	shengJiTierBlackJoker = 4
	shengJiTierRedJoker   = 5
)

// ShengJiPhase はゲームフェーズ。
type ShengJiPhase int

// 升级のフェーズ定数
const (
	// ShengJiPhaseDeclare 亮牌 (切札スートの宣言)
	ShengJiPhaseDeclare ShengJiPhase = iota
	// ShengJiPhaseKitty 底牌の埋め戻し
	ShengJiPhaseKitty
	// ShengJiPhasePlay トリック
	ShengJiPhasePlay
	// ShengJiPhaseHandEnd 局終了 (精算済み)
	ShengJiPhaseHandEnd
	// ShengJiPhaseGameEnd ゲーム終了
	ShengJiPhaseGameEnd
)

// ShengJiComboKind は 1 手の形。
type ShengJiComboKind int

// 升级の手の形
const (
	// ShengJiComboNone 形として成立していない
	ShengJiComboNone ShengJiComboKind = iota
	// ShengJiComboSingle 単張
	ShengJiComboSingle
	// ShengJiComboPair 対子 (**2 デッキなので同スート同ランクが 2 枚ある**)
	ShengJiComboPair
	// ShengJiComboTractor 拖拉機 (同スートの連続する対子 2 組以上)
	ShengJiComboTractor
)

// ShengJiCombo は 1 手の形と強さ。
type ShengJiCombo struct {
	Kind ShengJiComboKind
	// Rank は比較に使う最上位の序列 (段 × 100 + ランク)。
	Rank int
	// Size は枚数。
	Size int
	// Trump は切札群の手か。
	Trump bool
	// Suit は非切札のときのスート (切札群は ShengJiNoTrump)。
	Suit int
}

// ShengJiHandResult は 1 局の結果。
type ShengJiHandResult struct {
	// DeclarerTeam はこの局の宣言側。
	DeclarerTeam int
	// DefenderPoints は守備側が集めた点。**点を集めるのは守備側。**
	DefenderPoints int
	// KittyPoints は底牌から守備側に入った点 (倍率適用後)。
	KittyPoints int
	// KittyMultiplier は底牌に掛かった倍率 (守備側が最終トリックを取ったときのみ)。
	KittyMultiplier int
	// DeclarerHeld は宣言側が守りきったか。
	DeclarerHeld bool
	// Advance は昇級した段数。
	Advance int
	// AdvancingTeam は昇級したチーム (誰も上がらなければ -1)。
	AdvancingTeam int
}

// ShengJiDeclaration は亮牌の記録。
type ShengJiDeclaration struct {
	// Seat は宣言した席。
	Seat int
	// Suit は宣言された切札スート。
	Suit int
	// Strength は宣言の強さ (1 = 単張、2 = 対子)。**強い宣言が上書きする。**
	Strength int
}

// ShengJi は升级 (拖拉机) のゲームクラス。
//
// 中国発祥のレベルアップ式ポイントトリック。**切札は切札スートだけではなく、
// 全スートのレベル札とジョーカー 4 枚を含む。**得点を集めるのは守備側で、
// 宣言側は守備側を 80 点未満に抑えることで勝つ。
type ShengJi struct {
	players []*ShengJiPlayer
	config  ShengJiConfig
	phase   ShengJiPhase
	// levels は各チームの現在レベル。
	levels [ShengJiTeamCnt]int
	// level はこの局の基準レベル。
	level int
	// declarerTeam はこの局の宣言側。
	declarerTeam int
	// trumpSuit は切札スート (ShengJiNoTrump なら無主)。
	trumpSuit int
	// declaration は成立している亮牌 (無ければ nil)。
	declaration *ShengJiDeclaration
	// declareSeat は次に亮牌の機会が回る席。
	declareSeat int
	// kitty は底牌。
	kitty []*Card
	// currentIdx は手番。
	currentIdx int
	// trickLeader はいまのトリックのリード席。
	trickLeader int
	// trick はいまのトリックに出された手 (席順ではなくリードからの順)。
	trick [][]*Card
	// leadCombo はリードされた手の形。
	leadCombo *ShengJiCombo
	// teamPoints はこの局に各チームが集めた点。
	teamPoints [ShengJiTeamCnt]int
	// trickCount は消化したトリック数。
	trickCount int
	// lastTrickWinner は直前に解決したトリックの勝者席 (-1 は未解決)。
	lastTrickWinner int
	// lastTrickCards は直前に解決したトリックで 1 人が出した枚数。
	// **底牌の倍率がこれで決まる**ので、解決時に控えておく必要がある。
	lastTrickCards int
	lastResult     *ShengJiHandResult
	handNumber     int
	gameEndFlag    bool
	winnerTeam     int
	actionLog      []*ActionLogEntry
}

// NewShengJi コンストラクタ
func NewShengJi(players []*ShengJiPlayer, config ShengJiConfig) *ShengJi {
	s := &ShengJi{players: players, config: config, winnerTeam: -1, lastTrickWinner: -1}
	for i := range ShengJiTeamCnt {
		s.levels[i] = ShengJiMinLevel
	}
	s.level = ShengJiMinLevel
	return s
}

// NewDefaultShengJi は人間 1 人 + CPU 3 体の卓を作る。
func NewDefaultShengJi() *ShengJi {
	players := make([]*ShengJiPlayer, 0, ShengJiPlayerCnt)
	for i := range ShengJiPlayerCnt {
		players = append(players, NewShengJiPlayer(i == 0))
	}
	return NewShengJi(players, DefaultShengJiConfig())
}

// ShengJiTeamOf は席のチームを返す (範囲外は -1)。
//
// **パートナーは向かい合わせ。**
func ShengJiTeamOf(seat int) int {
	if seat < 0 || seat >= ShengJiPlayerCnt {
		return -1
	}
	return seat % ShengJiTeamCnt
}

// ---- 札の序列 ----

// shengJiNaturalRank は札の自然序列を返す (A が 14、ジョーカーは別枠)。
func shengJiNaturalRank(c *Card) int {
	if c == nil {
		return 0
	}
	if c.GetDesign() == CardDesignJoker {
		return 0
	}
	if c.GetValue() == 1 {
		return 14
	}
	return c.GetValue()
}

// ShengJiIsJoker は札がジョーカーかを返す。
func ShengJiIsJoker(c *Card) bool {
	return c != nil && c.GetDesign() == CardDesignJoker
}

// ShengJiIsLevelCard は札が現在のレベル札かを返す。
func ShengJiIsLevelCard(c *Card, level int) bool {
	if c == nil || ShengJiIsJoker(c) {
		return false
	}
	return shengJiNaturalRank(c) == level
}

// ShengJiIsTrump は札が切札群に属するかを返す。
//
// **切札スートだけではない。**全スートのレベル札とジョーカー 4 枚も切札。
// これを落とすとレベル制の意味が消える。
func ShengJiIsTrump(c *Card, level, trumpSuit int) bool {
	if c == nil {
		return false
	}
	if ShengJiIsJoker(c) || ShengJiIsLevelCard(c, level) {
		return true
	}
	return trumpSuit != ShengJiNoTrump && c.GetDesign() == trumpSuit
}

// ShengJiStrength は札の強さを返す (段 × 100 + 自然序列)。
//
// 赤ジョーカー > 黒ジョーカー > 切札スートのレベル札 > 他スートのレベル札
// > 切札スートの平札 > 非切札。**他スートのレベル札は互いに同格**なので、
// 同じ値を返す (先に出したほうが勝つのは呼び出し側で扱う)。
func ShengJiStrength(c *Card, level, trumpSuit int) int {
	if c == nil {
		return 0
	}
	switch {
	case ShengJiIsJoker(c):
		// **赤 (2) が黒 (1) より上。**
		if c.GetValue() >= 2 {
			return shengJiTierRedJoker * 100
		}
		return shengJiTierBlackJoker * 100
	case ShengJiIsLevelCard(c, level):
		if trumpSuit != ShengJiNoTrump && c.GetDesign() == trumpSuit {
			return shengJiTierTrumpLevel * 100
		}
		return shengJiTierOffLevel * 100
	case trumpSuit != ShengJiNoTrump && c.GetDesign() == trumpSuit:
		return shengJiTierTrumpSuit*100 + shengJiNaturalRank(c)
	}
	return shengJiTierPlain*100 + shengJiNaturalRank(c)
}

// ShengJiCardPoints は札の得点を返す。**5 が 5 点、10 と K が 10 点。**
func ShengJiCardPoints(c *Card) int {
	if c == nil || ShengJiIsJoker(c) {
		return 0
	}
	switch c.GetValue() {
	case 5:
		return 5
	case 10, 13:
		return 10
	}
	return 0
}

// shengJiSeqPos は連続の判定に使う目盛りを返す。
//
// **レベル札はそのスートの平札から抜ける**ので、目盛りを詰めないと隣り合わない。
// レベルが 5 のとき、4 と 6 は連続した対子として成立する。
func shengJiSeqPos(c *Card, level, trumpSuit int) int {
	st := ShengJiStrength(c, level, trumpSuit)
	if ShengJiIsJoker(c) || ShengJiIsLevelCard(c, level) {
		// 段そのものが目盛り。レベル札とジョーカーは平札の列に並ばない。
		return st
	}
	if shengJiNaturalRank(c) > level {
		return st - 1
	}
	return st
}

// shengJiSuitOf は札の所属を返す。**切札群はひとつのスートとして扱う。**
func shengJiSuitOf(c *Card, level, trumpSuit int) int {
	if ShengJiIsTrump(c, level, trumpSuit) {
		return ShengJiNoTrump
	}
	return c.GetDesign()
}

// ---- 手の判定 ----

// ShengJiEvaluate は出された札から手の形を作る (成立しなければ nil)。
func ShengJiEvaluate(cards []*Card, level, trumpSuit int) *ShengJiCombo {
	if len(cards) == 0 {
		return nil
	}
	// **nil はスートを見る前に弾く。**GetCard は範囲外で nil を返す。
	for _, c := range cards {
		if c == nil {
			return nil
		}
	}
	// **1 手はひとつのスートから。**混ぜた手は形として成立しない。
	suit := shengJiSuitOf(cards[0], level, trumpSuit)
	trump := ShengJiIsTrump(cards[0], level, trumpSuit)
	for _, c := range cards {
		if shengJiSuitOf(c, level, trumpSuit) != suit || ShengJiIsTrump(c, level, trumpSuit) != trump {
			return nil
		}
	}

	top := 0
	for _, c := range cards {
		if s := ShengJiStrength(c, level, trumpSuit); s > top {
			top = s
		}
	}

	if len(cards) == 1 {
		return &ShengJiCombo{Kind: ShengJiComboSingle, Rank: top, Size: 1, Trump: trump, Suit: suit}
	}
	if !shengJiAllPaired(cards) {
		return nil
	}
	if len(cards) == 2 {
		return &ShengJiCombo{Kind: ShengJiComboPair, Rank: top, Size: 2, Trump: trump, Suit: suit}
	}
	if !shengJiPairsAreConsecutive(cards, level, trumpSuit) {
		return nil
	}
	return &ShengJiCombo{Kind: ShengJiComboTractor, Rank: top, Size: len(cards), Trump: trump, Suit: suit}
}

// shengJiAllPaired は札がすべて 2 枚組になっているかを返す。
//
// **対子は同ランクではなく同一の札 2 枚。**2 デッキなので ♠K が 2 枚ある。
func shengJiAllPaired(cards []*Card) bool {
	if len(cards)%2 != 0 {
		return false
	}
	counts := map[string]int{}
	for _, c := range cards {
		counts[shengJiCardKey(c)]++
	}
	for _, n := range counts {
		if n != 2 {
			return false
		}
	}
	return true
}

// shengJiCardKey は同一札の判定キーを返す。
func shengJiCardKey(c *Card) string {
	return strconv.Itoa(c.GetDesign()) + ":" + strconv.Itoa(c.GetValue())
}

// shengJiPairsAreConsecutive は対子が序列上で連続しているかを返す。
func shengJiPairsAreConsecutive(cards []*Card, level, trumpSuit int) bool {
	seen := map[string]bool{}
	strengths := make([]int, 0, len(cards)/2)
	for _, c := range cards {
		k := shengJiCardKey(c)
		if seen[k] {
			continue
		}
		seen[k] = true
		strengths = append(strengths, shengJiSeqPos(c, level, trumpSuit))
	}
	sort.Ints(strengths)
	// **同格の札は連続にならない。**他スートのレベル札どうしは強さが等しい。
	for i := 1; i < len(strengths); i++ {
		if strengths[i] != strengths[i-1]+1 {
			return false
		}
	}
	return true
}

// ShengJiBeats は c が prev を上回るかを返す。
//
// **形と枚数が一致しなければ勝てない。**切札は非切札を上回るが、その場合も
// 形と枚数は一致している必要がある。
func ShengJiBeats(c, prev *ShengJiCombo, ledSuit int) bool {
	if c == nil || prev == nil {
		return false
	}
	if c.Kind != prev.Kind || c.Size != prev.Size {
		return false
	}
	if c.Trump != prev.Trump {
		// **切札だけが割り込める。**
		return c.Trump
	}
	if !c.Trump && c.Suit != prev.Suit {
		return false
	}
	if !c.Trump && c.Suit != ledSuit {
		return false
	}
	return c.Rank > prev.Rank
}

// ---- 局の進行 ----

// Reset はゲームを初期化する。
func (s *ShengJi) Reset() {
	for i := range ShengJiTeamCnt {
		s.levels[i] = ShengJiMinLevel
	}
	s.level = ShengJiMinLevel
	s.declarerTeam = 0
	s.handNumber = 0
	s.gameEndFlag = false
	s.winnerTeam = -1
	s.lastResult = nil
	s.actionLog = make([]*ActionLogEntry, 0)
	s.beginHand()
}

// beginHand は 1 局を配る。
func (s *ShengJi) beginHand() {
	s.handNumber++
	s.level = s.levels[s.declarerTeam]
	s.trumpSuit = ShengJiNoTrump
	s.declaration = nil
	s.declareSeat = 0
	s.trick = make([][]*Card, 0, ShengJiPlayerCnt)
	s.leadCombo = nil
	s.trickCount = 0
	s.lastTrickWinner = -1
	s.lastTrickCards = 0
	for i := range ShengJiTeamCnt {
		s.teamPoints[i] = 0
	}
	for i := range ShengJiPlayerCnt {
		if p := s.GetPlayer(i); p != nil {
			p.ResetRound()
		}
	}
	s.dealRound()
	s.phase = ShengJiPhaseDeclare
	s.currentIdx = 0
	s.addLog(-1, "deal", "hand "+strconv.Itoa(s.handNumber)+" level "+strconv.Itoa(s.level), nil)
}

// dealRound は 25 枚ずつ配り、**残り 8 枚を底牌として伏せる**。
func (s *ShengJi) dealRound() {
	deck := newShengJiDeck()
	shengJiShuffle(deck)
	pos := 0
	for range ShengJiHandSize {
		for i := range ShengJiPlayerCnt {
			p := s.GetPlayer(i)
			if pos < len(deck) && p != nil {
				p.AddCard(deck[pos])
				pos++
			}
		}
	}
	s.kitty = make([]*Card, 0, ShengJiKittySize)
	s.kitty = append(s.kitty, deck[pos:]...)
	// **25 枚を添字で指定して手を組む**ので、並べ替えておかないと実用に耐えない。
	for i := range ShengJiPlayerCnt {
		s.sortHand(i)
	}
}

// newShengJiDeck は **52 × 2 + ジョーカー 4 = 108 枚**を作る。
func newShengJiDeck() []*Card {
	cards := make([]*Card, 0, ShengJiDeckSize)
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

// shengJiShuffle は札をシャッフルする。
func shengJiShuffle(cards []*Card) {
	for i := len(cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1) //nolint:gosec // ゲームのシャッフルに暗号強度は要らない
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// ---- 亮牌 ----

// ShengJiDeclareStrength は席がそのスートで出せる亮牌の強さを返す (0 は不可)。
//
// **亮牌はレベル札を見せて行う。**そのスートのレベル札を 1 枚持っていれば 1、
// 2 枚持っていれば 2。強い宣言が弱い宣言を上書きする。
func (s *ShengJi) ShengJiDeclareStrength(seat, suit int) int {
	p := s.GetPlayer(seat)
	if p == nil || suit < CardDesignSpade || suit > CardDesignDiamond {
		return 0
	}
	n := 0
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if c != nil && c.GetDesign() == suit && shengJiNaturalRank(c) == s.level {
			n++
		}
	}
	if n > 2 {
		n = 2
	}
	return n
}

// Declare は亮牌する。suit に ShengJiNoTrump を渡すとパス。
func (s *ShengJi) Declare(seat, suit int) error {
	if s.gameEndFlag {
		return errors.New("the game is over")
	}
	if s.phase != ShengJiPhaseDeclare {
		return errors.New("it is not the declaring phase")
	}
	if seat != s.currentIdx {
		return errors.New("it is not your turn")
	}
	if suit != ShengJiNoTrump {
		st := s.ShengJiDeclareStrength(seat, suit)
		if st == 0 {
			return errors.New("you hold no level card of that suit")
		}
		// **強い宣言だけが上書きできる。**同じ強さでは覆せない。
		if s.declaration != nil && st <= s.declaration.Strength {
			return errors.New("that declaration is not strong enough to override")
		}
		s.declaration = &ShengJiDeclaration{Seat: seat, Suit: suit, Strength: st}
		s.trumpSuit = suit
		s.addLog(seat, "declare", strconv.Itoa(suit)+" x"+strconv.Itoa(st), nil)
	}
	s.advanceDeclare()
	return nil
}

// advanceDeclare は亮牌の手番を進め、一巡したら底牌フェーズへ移す。
func (s *ShengJi) advanceDeclare() {
	s.declareSeat++
	if s.declareSeat < ShengJiPlayerCnt {
		s.currentIdx = s.declareSeat
		return
	}
	// **誰も亮牌しなければ無主。**レベル札とジョーカーだけが切札になる。
	s.beginKitty()
}

// beginKitty は底牌を宣言側に渡す。
func (s *ShengJi) beginKitty() {
	s.phase = ShengJiPhaseKitty
	// 底牌を取るのは亮牌した席。誰も亮牌しなければ宣言側の先頭席。
	taker := s.shengJiKittyTaker()
	if p := s.GetPlayer(taker); p != nil {
		for _, c := range s.kitty {
			p.AddCard(c)
		}
	}
	s.kitty = make([]*Card, 0, ShengJiKittySize)
	s.currentIdx = taker
	s.sortHand(taker)
	s.addLog(taker, "takeKitty", "", nil)
}

// shengJiKittyTaker は底牌を取る席を返す。
func (s *ShengJi) shengJiKittyTaker() int {
	if s.declaration != nil {
		return s.declaration.Seat
	}
	for i := range ShengJiPlayerCnt {
		if ShengJiTeamOf(i) == s.declarerTeam {
			return i
		}
	}
	return 0
}

// BuryKitty は底牌に **8 枚**を埋め戻す。
func (s *ShengJi) BuryKitty(seat int, idxs []int) error {
	if s.gameEndFlag {
		return errors.New("the game is over")
	}
	if s.phase != ShengJiPhaseKitty {
		return errors.New("it is not the kitty phase")
	}
	if seat != s.currentIdx {
		return errors.New("it is not your turn")
	}
	if len(idxs) != ShengJiKittySize {
		return errors.New("bury exactly " + strconv.Itoa(ShengJiKittySize) + " cards")
	}
	cards, err := s.takeCards(seat, idxs)
	if err != nil {
		return err
	}
	s.kitty = append(s.kitty, cards...)
	s.phase = ShengJiPhasePlay
	// **リードするのは底牌を取った側。**
	s.trickLeader = seat
	s.currentIdx = seat
	s.trick = make([][]*Card, 0, ShengJiPlayerCnt)
	s.leadCombo = nil
	s.addLog(seat, "buryKitty", "", cards)
	return nil
}

// takeCards は手札から添字の札を取り出す (重複と範囲外は拒否する)。
func (s *ShengJi) takeCards(seat int, idxs []int) ([]*Card, error) {
	p := s.GetPlayer(seat)
	if p == nil {
		return nil, errors.New("there is no such seat")
	}
	seen := map[int]bool{}
	for _, i := range idxs {
		if i < 0 || i >= p.GetCardsSize() {
			return nil, errors.New("there is no such card")
		}
		// **同じ札を 2 回数えられない。**通すと 1 枚から対子が作れてしまう。
		if seen[i] {
			return nil, errors.New("the same card was given twice")
		}
		seen[i] = true
	}
	cards := make([]*Card, 0, len(idxs))
	for _, i := range idxs {
		cards = append(cards, p.GetCard(i))
	}
	sorted := append([]int(nil), idxs...)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	for _, i := range sorted {
		p.RemoveCard(i)
	}
	return cards, nil
}

// sortHand は手札をこの局の序列で並べ替える。
//
// **25 枚を添字で指定して手を組む**ので、並べ替えておかないと実用に耐えない。
func (s *ShengJi) sortHand(seat int) {
	p := s.GetPlayer(seat)
	if p == nil {
		return
	}
	cards := make([]*Card, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		si := shengJiSuitOf(cards[i], s.level, s.trumpSuit)
		sj := shengJiSuitOf(cards[j], s.level, s.trumpSuit)
		if si != sj {
			return si < sj
		}
		return ShengJiStrength(cards[i], s.level, s.trumpSuit) <
			ShengJiStrength(cards[j], s.level, s.trumpSuit)
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// ---- トリック ----

// Play は手を出す。
func (s *ShengJi) Play(seat int, idxs []int) error {
	if s.gameEndFlag {
		return errors.New("the game is over")
	}
	if s.phase != ShengJiPhasePlay {
		return errors.New("it is not the playing phase")
	}
	if seat != s.currentIdx {
		return errors.New("it is not your turn")
	}
	p := s.GetPlayer(seat)
	if p == nil {
		return errors.New("there is no such seat")
	}
	if len(idxs) == 0 {
		return errors.New("play at least one card")
	}
	if s.leadCombo != nil && len(idxs) != s.leadCombo.Size {
		return errors.New("play the same number of cards as the lead")
	}

	cards := make([]*Card, 0, len(idxs))
	for _, i := range idxs {
		if i < 0 || i >= p.GetCardsSize() {
			return errors.New("there is no such card")
		}
		cards = append(cards, p.GetCard(i))
	}
	if s.leadCombo == nil {
		if ShengJiEvaluate(cards, s.level, s.trumpSuit) == nil {
			return errors.New("that is not a single, a pair or a tractor")
		}
	} else if err := s.checkFollow(seat, cards); err != nil {
		return err
	}

	taken, err := s.takeCards(seat, idxs)
	if err != nil {
		return err
	}
	if s.leadCombo == nil {
		s.leadCombo = ShengJiEvaluate(taken, s.level, s.trumpSuit)
		s.trickLeader = seat
	}
	s.trick = append(s.trick, taken)
	s.addLog(seat, "play", "", taken)

	if len(s.trick) >= ShengJiPlayerCnt {
		s.resolveTrick()
		return nil
	}
	s.currentIdx = (s.currentIdx + 1) % ShengJiPlayerCnt
	return nil
}

// checkFollow はフォローの義務を検査する。
//
// **リードされたスートを持っているなら、そのスートから出さなければならない。**
// 切札群はひとつのスートとして扱う。
func (s *ShengJi) checkFollow(seat int, cards []*Card) error {
	led := s.leadCombo.Suit
	if s.leadCombo.Trump {
		led = ShengJiNoTrump
	}
	held := s.countInSuit(seat, led)
	if held == 0 {
		// 持っていなければ何を出してもよい (勝てるのは切札の同形だけ)。
		return nil
	}
	need := min(held, len(cards))
	got := 0
	for _, c := range cards {
		if s.inSuit(c, led) {
			got++
		}
	}
	if got < need {
		return errors.New("you must follow the led suit while you hold it")
	}

	// **対子がリードされたら、そのスートの対子を先に出さなければならない。**
	// 枚数だけ合わせて対子を温存できると、拖拉機を出す意味が無くなる。
	if s.leadCombo.Kind == ShengJiComboPair || s.leadCombo.Kind == ShengJiComboTractor {
		wantPairs := s.leadCombo.Size / 2
		if n := min(s.countPairsInSuit(seat, led), wantPairs); shengJiPairCount(cards) < n {
			return errors.New("you must play your pairs of the led suit")
		}
	}
	return nil
}

// countPairsInSuit は席がそのスートに持っている対子の数を返す。
//
// **対子は同ランクではなく同一の札 2 枚。**2 デッキなので ♠K が 2 枚ある。
func (s *ShengJi) countPairsInSuit(seat, suit int) int {
	p := s.GetPlayer(seat)
	if p == nil {
		return 0
	}
	counts := map[string]int{}
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if s.inSuit(c, suit) {
			counts[shengJiCardKey(c)]++
		}
	}
	n := 0
	for _, v := range counts {
		n += v / 2
	}
	return n
}

// shengJiPairCount は出された札に含まれる対子の数を返す。
func shengJiPairCount(cards []*Card) int {
	counts := map[string]int{}
	for _, c := range cards {
		if c != nil {
			counts[shengJiCardKey(c)]++
		}
	}
	n := 0
	for _, v := range counts {
		n += v / 2
	}
	return n
}

// inSuit は札がそのスート (切札群を含む) に属するかを返す。
func (s *ShengJi) inSuit(c *Card, suit int) bool {
	if suit == ShengJiNoTrump {
		return ShengJiIsTrump(c, s.level, s.trumpSuit)
	}
	return !ShengJiIsTrump(c, s.level, s.trumpSuit) && c.GetDesign() == suit
}

// countInSuit は席がそのスートに何枚持っているかを返す。
func (s *ShengJi) countInSuit(seat, suit int) int {
	p := s.GetPlayer(seat)
	if p == nil {
		return 0
	}
	n := 0
	for i := range p.GetCardsSize() {
		if s.inSuit(p.GetCard(i), suit) {
			n++
		}
	}
	return n
}

// resolveTrick はトリックを解決して点を配る。
func (s *ShengJi) resolveTrick() {
	winOffset := s.trickWinnerOffset()
	winner := (s.trickLeader + winOffset) % ShengJiPlayerCnt

	pts := 0
	for _, play := range s.trick {
		for _, c := range play {
			pts += ShengJiCardPoints(c)
		}
	}
	// **点を集めるのは守備側。**宣言側が取った点はどこにも積まれない。
	if team := ShengJiTeamOf(winner); team != s.declarerTeam {
		s.teamPoints[team] += pts
	}

	s.lastTrickWinner = winner
	s.lastTrickCards = s.leadCombo.Size
	s.trickCount++
	s.trick = make([][]*Card, 0, ShengJiPlayerCnt)
	s.leadCombo = nil
	s.trickLeader = winner
	s.currentIdx = winner
	s.addLog(winner, "trickWon", strconv.Itoa(pts), nil)

	if s.GetPlayer(winner) != nil && s.GetPlayer(winner).GetCardsSize() == 0 {
		s.finishHand()
	}
}

// trickWinnerOffset はリードからの何番目が勝ったかを返す。
func (s *ShengJi) trickWinnerOffset() int {
	if len(s.trick) == 0 {
		return 0
	}
	lead := ShengJiEvaluate(s.trick[0], s.level, s.trumpSuit)
	if lead == nil {
		return 0
	}
	ledSuit := lead.Suit
	best, bestIdx := lead, 0
	for i := 1; i < len(s.trick); i++ {
		c := ShengJiEvaluate(s.trick[i], s.level, s.trumpSuit)
		if c == nil {
			continue
		}
		if ShengJiBeats(c, best, ledSuit) {
			best, bestIdx = c, i
		}
	}
	return bestIdx
}

// finishHand は局を精算する。
func (s *ShengJi) finishHand() {
	defenders := 1 - s.declarerTeam
	pts := s.teamPoints[defenders]

	kittyPts, mult := 0, 0
	// **守備側が最終トリックを取ると底牌が倍率つきで入る。**
	if s.lastTrickWinner >= 0 && ShengJiTeamOf(s.lastTrickWinner) != s.declarerTeam {
		raw := 0
		for _, c := range s.kitty {
			raw += ShengJiCardPoints(c)
		}
		mult = 2 * s.lastTrickCards
		kittyPts = raw * mult
		pts += kittyPts
	}

	res := &ShengJiHandResult{
		DeclarerTeam: s.declarerTeam, DefenderPoints: pts,
		KittyPoints: kittyPts, KittyMultiplier: mult, AdvancingTeam: -1,
	}
	if pts < ShengJiDefenderTarget {
		res.DeclarerHeld = true
		res.Advance = shengJiDeclarerAdvance(pts)
		res.AdvancingTeam = s.declarerTeam
	} else {
		// **80 点で宣言側が交代する。**そこから 40 点ごとに守備側が 1 段階。
		res.Advance = (pts - ShengJiDefenderTarget) / ShengJiAdvanceStep
		if res.Advance > 0 {
			res.AdvancingTeam = defenders
		}
		s.declarerTeam = defenders
	}

	// **A は飛び越えられない。**K から 3 段階でも A で止まり、そのうえで A の局を
	// 守りきって初めて勝ちになる (打A)。
	if res.AdvancingTeam >= 0 {
		team := res.AdvancingTeam
		if s.levels[team] == ShengJiMaxLevel {
			s.gameEndFlag = true
			s.winnerTeam = team
			s.phase = ShengJiPhaseGameEnd
		} else {
			s.levels[team] = min(s.levels[team]+res.Advance, ShengJiMaxLevel)
		}
	}
	if !s.gameEndFlag {
		s.phase = ShengJiPhaseHandEnd
	}
	s.lastResult = res
	s.addLog(-1, "handEnd", strconv.Itoa(pts), nil)
}

// shengJiDeclarerAdvance は宣言側が守りきったときの昇級量を返す。
//
// **0 点なら 3 段階、40 点未満なら 2 段階、80 点未満なら 1 段階。**
func shengJiDeclarerAdvance(defenderPoints int) int {
	switch {
	case defenderPoints == 0:
		return 3
	case defenderPoints < ShengJiAdvanceStep:
		return 2
	}
	return 1
}

// NextHand は次の局へ進む。
func (s *ShengJi) NextHand() error {
	if s.gameEndFlag {
		return errors.New("the game is over")
	}
	if s.phase != ShengJiPhaseHandEnd {
		return errors.New("the hand is still in play")
	}
	s.beginHand()
	return nil
}

// ---- アクセサ ----

// GetPlayers は全プレイヤーを返す。
func (s *ShengJi) GetPlayers() []*ShengJiPlayer { return s.players }

// GetPlayer は指定インデックスのプレイヤーを返す。
func (s *ShengJi) GetPlayer(idx int) *ShengJiPlayer {
	if idx < 0 || idx >= len(s.players) {
		return nil
	}
	return s.players[idx]
}

// GetPhase は現在のフェーズを返す。
func (s *ShengJi) GetPhase() ShengJiPhase { return s.phase }

// GetCurrentPlayerIdx は現在の手番を返す。
func (s *ShengJi) GetCurrentPlayerIdx() int { return s.currentIdx }

// GetLevel はこの局の基準レベルを返す。
func (s *ShengJi) GetLevel() int { return s.level }

// GetTeamLevel はチームの現在レベルを返す。
func (s *ShengJi) GetTeamLevel(team int) int {
	if team < 0 || team >= ShengJiTeamCnt {
		return 0
	}
	return s.levels[team]
}

// GetDeclarerTeam はこの局の宣言側を返す。
func (s *ShengJi) GetDeclarerTeam() int { return s.declarerTeam }

// GetTrumpSuit は切札スートを返す (ShengJiNoTrump なら無主)。
func (s *ShengJi) GetTrumpSuit() int { return s.trumpSuit }

// GetDeclaration は成立している亮牌を返す。
func (s *ShengJi) GetDeclaration() *ShengJiDeclaration { return s.declaration }

// GetKittySize は底牌の枚数を返す。**中身は終局まで公開しない。**
func (s *ShengJi) GetKittySize() int { return len(s.kitty) }

// GetKitty は底牌を返す (終局まで空)。
func (s *ShengJi) GetKitty() []*Card {
	if s.phase == ShengJiPhaseHandEnd || s.gameEndFlag {
		return s.kitty
	}
	return nil
}

// GetTrick はいまのトリックに出された手を返す。
func (s *ShengJi) GetTrick() [][]*Card { return s.trick }

// GetTrickLeader はいまのトリックのリード席を返す。
func (s *ShengJi) GetTrickLeader() int { return s.trickLeader }

// GetLeadCombo はリードされた手の形を返す。
func (s *ShengJi) GetLeadCombo() *ShengJiCombo { return s.leadCombo }

// GetTeamPoints はチームが集めた点を返す。
func (s *ShengJi) GetTeamPoints(team int) int {
	if team < 0 || team >= ShengJiTeamCnt {
		return 0
	}
	return s.teamPoints[team]
}

// GetTrickCount は消化したトリック数を返す。
func (s *ShengJi) GetTrickCount() int { return s.trickCount }

// GetLastTrickWinner は直前のトリックの勝者席を返す。
func (s *ShengJi) GetLastTrickWinner() int { return s.lastTrickWinner }

// GetLastResult は直前の局の結果を返す。
func (s *ShengJi) GetLastResult() *ShengJiHandResult { return s.lastResult }

// GetHandNumber は現在の局番号を返す。
func (s *ShengJi) GetHandNumber() int { return s.handNumber }

// GetGameEndFlag はゲーム終了フラグを返す。
func (s *ShengJi) GetGameEndFlag() bool { return s.gameEndFlag }

// GetWinnerTeam は勝利チームを返す。
func (s *ShengJi) GetWinnerTeam() int { return s.winnerTeam }

// GetConfig はゲーム設定を返す。
func (s *ShengJi) GetConfig() ShengJiConfig { return s.config }

// SetConfig はゲーム設定をセットする。
func (s *ShengJi) SetConfig(c ShengJiConfig) { s.config = c }

// GetActionLog は棋譜を返す。
func (s *ShengJi) GetActionLog() []*ActionLogEntry { return s.actionLog }

// addLog は棋譜に 1 件追加する。
func (s *ShengJi) addLog(playerIdx int, actionType, detail string, cards []*Card) {
	s.actionLog = append(s.actionLog, &ActionLogEntry{
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// IsHumanTurn は現在の手番が人間かを返す。
func (s *ShengJi) IsHumanTurn() bool {
	if s.gameEndFlag || s.phase == ShengJiPhaseHandEnd || s.phase == ShengJiPhaseGameEnd {
		return false
	}
	p := s.GetPlayer(s.currentIdx)
	return p != nil && p.GetIsHuman()
}

// ---- CPU ----

// CpuPlay は CPU が 1 アクション実行する。
func (s *ShengJi) CpuPlay() {
	if s.gameEndFlag || s.IsHumanTurn() {
		return
	}
	seat := s.currentIdx
	switch s.phase {
	case ShengJiPhaseDeclare:
		s.cpuDeclare(seat)
	case ShengJiPhaseKitty:
		_ = s.BuryKitty(seat, s.ShengJiCpuBury(seat))
	case ShengJiPhasePlay:
		if idxs := s.ShengJiCpuPlay(seat); len(idxs) > 0 {
			if s.Play(seat, idxs) == nil {
				return
			}
		}
		// 詰み回避。**リードでもフォローでも 1 枚は必ず出せる。**
		if p := s.GetPlayer(seat); p != nil && p.GetCardsSize() > 0 {
			_ = s.Play(seat, []int{0})
		}
	}
}

// cpuDeclare は CPU の亮牌を決める。**上書きできるなら上書きする。**
func (s *ShengJi) cpuDeclare(seat int) {
	best, bestSuit := 0, ShengJiNoTrump
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if st := s.ShengJiDeclareStrength(seat, suit); st > best {
			best, bestSuit = st, suit
		}
	}
	if best > 0 && (s.declaration == nil || best > s.declaration.Strength) {
		if s.Declare(seat, bestSuit) == nil {
			return
		}
	}
	_ = s.Declare(seat, ShengJiNoTrump)
}

// ShengJiCpuBury は CPU が底牌へ埋め戻す 8 枚の添字を返す。
//
// **得点札と切札は埋めない。**埋めた点は守備側が最終トリックを取ると倍率つきで
// 相手のものになる。
func (s *ShengJi) ShengJiCpuBury(seat int) []int {
	p := s.GetPlayer(seat)
	if p == nil {
		return nil
	}
	type scored struct {
		idx, score int
	}
	all := make([]scored, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		score := ShengJiStrength(c, s.level, s.trumpSuit)
		if ShengJiIsTrump(c, s.level, s.trumpSuit) {
			score += 1000
		}
		if ShengJiCardPoints(c) > 0 {
			score += 500
		}
		all = append(all, scored{i, score})
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].score < all[j].score })
	out := make([]int, 0, ShengJiKittySize)
	for i := 0; i < ShengJiKittySize && i < len(all); i++ {
		out = append(out, all[i].idx)
	}
	return out
}

// ShengJiCpuPlay は CPU が出す手の添字を返す。
func (s *ShengJi) ShengJiCpuPlay(seat int) []int {
	p := s.GetPlayer(seat)
	if p == nil || p.GetCardsSize() == 0 {
		return nil
	}
	if s.leadCombo == nil {
		// リードは手札のいちばん弱い札から。
		return []int{s.weakestIdx(seat, -1)}
	}
	led := s.leadCombo.Suit
	if s.leadCombo.Trump {
		led = ShengJiNoTrump
	}
	need := s.leadCombo.Size
	// **リードされたスートを持っているなら、そこから出す。**
	idxs := s.pickFromSuit(seat, led, need)
	for len(idxs) < need {
		idxs = append(idxs, s.weakestIdx(seat, -1, idxs...))
	}
	return idxs[:need]
}

// pickFromSuit は席がそのスートから出せる添字を弱い順に最大 n 件返す。
func (s *ShengJi) pickFromSuit(seat, suit, n int) []int {
	p := s.GetPlayer(seat)
	if p == nil {
		return nil
	}
	type scored struct {
		idx, score int
	}
	all := make([]scored, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		if s.inSuit(p.GetCard(i), suit) {
			all = append(all, scored{i, ShengJiStrength(p.GetCard(i), s.level, s.trumpSuit)})
		}
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].score < all[j].score })
	out := make([]int, 0, n)
	for i := 0; i < n && i < len(all); i++ {
		out = append(out, all[i].idx)
	}
	return out
}

// weakestIdx は席のいちばん弱い札の添字を返す (except は除く)。
func (s *ShengJi) weakestIdx(seat, suit int, except ...int) int {
	p := s.GetPlayer(seat)
	if p == nil {
		return 0
	}
	skip := map[int]bool{}
	for _, i := range except {
		skip[i] = true
	}
	best, bestScore := -1, 1<<30
	for i := range p.GetCardsSize() {
		if skip[i] {
			continue
		}
		c := p.GetCard(i)
		if suit >= 0 && !s.inSuit(c, suit) {
			continue
		}
		if sc := ShengJiStrength(c, s.level, s.trumpSuit); sc < bestScore {
			best, bestScore = i, sc
		}
	}
	if best < 0 {
		return 0
	}
	return best
}

// ---- KV 永続化 ----

// shengJiJSON is the KV wire format for ShengJi.
type shengJiJSON struct {
	Players         []*ShengJiPlayer    `json:"pl"`
	Config          ShengJiConfig       `json:"cf"`
	Phase           ShengJiPhase        `json:"ph"`
	Levels          [ShengJiTeamCnt]int `json:"lv"`
	Level           int                 `json:"le"`
	DeclarerTeam    int                 `json:"dt"`
	TrumpSuit       int                 `json:"ts"`
	Declaration     *ShengJiDeclaration `json:"dc"`
	DeclareSeat     int                 `json:"ds"`
	Kitty           []*Card             `json:"kt"`
	CurrentIdx      int                 `json:"ci"`
	TrickLeader     int                 `json:"tl"`
	Trick           [][]*Card           `json:"tk"`
	TeamPoints      [ShengJiTeamCnt]int `json:"tp"`
	TrickCount      int                 `json:"tc"`
	LastTrickWinner int                 `json:"lw"`
	LastTrickCards  int                 `json:"lc"`
	LastResult      *ShengJiHandResult  `json:"lr"`
	HandNumber      int                 `json:"hn"`
	GameEndFlag     bool                `json:"ge"`
	WinnerTeam      int                 `json:"wt"`
	ActionLog       []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (s *ShengJi) MarshalJSON() ([]byte, error) {
	return json.Marshal(shengJiJSON{
		Players: s.players, Config: s.config, Phase: s.phase,
		Levels: s.levels, Level: s.level, DeclarerTeam: s.declarerTeam,
		TrumpSuit: s.trumpSuit, Declaration: s.declaration, DeclareSeat: s.declareSeat,
		Kitty: s.kitty, CurrentIdx: s.currentIdx, TrickLeader: s.trickLeader, Trick: s.trick,
		TeamPoints: s.teamPoints, TrickCount: s.trickCount,
		LastTrickWinner: s.lastTrickWinner, LastTrickCards: s.lastTrickCards,
		LastResult: s.lastResult, HandNumber: s.handNumber,
		GameEndFlag: s.gameEndFlag, WinnerTeam: s.winnerTeam, ActionLog: s.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **KV から戻る値なので範囲を検査する。**壊れた状態をそのまま受け入れると
// 添字で落ちる。
func (s *ShengJi) UnmarshalJSON(data []byte) error {
	var j shengJiJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != ShengJiPlayerCnt {
		return errors.New("sheng ji needs exactly four seats")
	}
	if j.Phase < ShengJiPhaseDeclare || j.Phase > ShengJiPhaseGameEnd {
		return errors.New("unknown phase")
	}
	if j.CurrentIdx < 0 || j.CurrentIdx >= ShengJiPlayerCnt {
		return errors.New("bad current seat")
	}
	if j.TrickLeader < 0 || j.TrickLeader >= ShengJiPlayerCnt {
		return errors.New("bad trick leader")
	}
	if j.DeclarerTeam < 0 || j.DeclarerTeam >= ShengJiTeamCnt {
		return errors.New("bad declarer team")
	}
	if j.WinnerTeam < -1 || j.WinnerTeam >= ShengJiTeamCnt {
		return errors.New("bad winner team")
	}
	// **無主も正当な状態。**切札スート無しで進む局がある。
	if j.TrumpSuit != ShengJiNoTrump && (j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return errors.New("bad trump suit")
	}
	if j.Level < ShengJiMinLevel || j.Level > ShengJiMaxLevel {
		return errors.New("bad level")
	}
	for _, lv := range j.Levels {
		if lv < ShengJiMinLevel || lv > ShengJiMaxLevel+1 {
			return errors.New("bad team level")
		}
	}
	if j.LastTrickWinner < -1 || j.LastTrickWinner >= ShengJiPlayerCnt {
		return errors.New("bad last trick winner")
	}
	if d := j.Declaration; d != nil {
		if d.Seat < 0 || d.Seat >= ShengJiPlayerCnt {
			return errors.New("bad declaring seat")
		}
		if d.Suit < CardDesignSpade || d.Suit > CardDesignDiamond {
			return errors.New("bad declared suit")
		}
	}
	if len(j.Kitty) > ShengJiKittySize {
		return errors.New("the kitty holds at most " + strconv.Itoa(ShengJiKittySize) + " cards")
	}

	s.players = j.Players
	s.config = j.Config
	s.phase = j.Phase
	s.levels = j.Levels
	s.level = j.Level
	s.declarerTeam = j.DeclarerTeam
	s.trumpSuit = j.TrumpSuit
	s.declaration = j.Declaration
	s.declareSeat = j.DeclareSeat
	s.kitty = j.Kitty
	s.currentIdx = j.CurrentIdx
	s.trickLeader = j.TrickLeader
	s.trick = j.Trick
	s.teamPoints = j.TeamPoints
	s.trickCount = j.TrickCount
	s.lastTrickWinner = j.LastTrickWinner
	s.lastTrickCards = j.LastTrickCards
	s.lastResult = j.LastResult
	s.handNumber = j.HandNumber
	s.gameEndFlag = j.GameEndFlag
	s.winnerTeam = j.WinnerTeam
	s.actionLog = j.ActionLog
	// **リードの形は札から復元する。**保存すると形と札がずれ得る。
	if len(s.trick) > 0 {
		s.leadCombo = ShengJiEvaluate(s.trick[0], s.level, s.trumpSuit)
	}
	return nil
}

// ---- テスト用 ----

// SetPhaseForTest はフェーズを差し替える (テスト専用)。
func (s *ShengJi) SetPhaseForTest(p ShengJiPhase) { s.phase = p }

// SetHandForTest は手札を差し替える (テスト専用)。
func (s *ShengJi) SetHandForTest(idx int, cards []*Card) {
	p := s.GetPlayer(idx)
	if p == nil {
		return
	}
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// SetLevelForTest はこの局のレベルを差し替える (テスト専用)。
func (s *ShengJi) SetLevelForTest(level int) { s.level = level }

// SetTeamLevelForTest はチームのレベルを差し替える (テスト専用)。
func (s *ShengJi) SetTeamLevelForTest(team, level int) {
	if team < 0 || team >= ShengJiTeamCnt {
		return
	}
	s.levels[team] = level
}

// SetTrumpForTest は切札スートを差し替える (テスト専用)。
func (s *ShengJi) SetTrumpForTest(suit int) { s.trumpSuit = suit }

// SetCurrentPlayerForTest は手番を差し替える (テスト専用)。
func (s *ShengJi) SetCurrentPlayerForTest(idx int) {
	s.currentIdx = idx
	s.trickLeader = idx
}

// SetKittyForTest は底牌を差し替える (テスト専用)。
func (s *ShengJi) SetKittyForTest(cards []*Card) { s.kitty = cards }

// FinishHandForTest は局の精算を走らせる (テスト専用)。
func (s *ShengJi) FinishHandForTest() { s.finishHand() }
